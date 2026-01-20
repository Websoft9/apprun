// Package service provides business logic for authentication and authorization.
package service

import (
	"context"
	stdErrors "errors"
	"strings"

	"apprun/ent"
	"apprun/internal/password"
	"apprun/modules/auth/repository"
	"apprun/pkg/errors"
	"apprun/pkg/logger"

	"github.com/google/uuid"
)

var (
	// SystemUserUUID is the fixed UUID for the system user
	SystemUserUUID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// ErrSystemCannotLogin is returned when system user attempts to login
	ErrSystemCannotLogin = errors.New(errors.ErrCodeAuthAccountDisabled, "System user cannot login")
	// ErrSystemUserCorruption is returned when system user data is inconsistent
	ErrSystemUserCorruption = errors.New(errors.ErrCodeInternalError, "System user data corruption detected")
)

// EnsureSystemUser ensures the system user exists in the database.
// This method is idempotent - calling it multiple times will return the same user.
//
// The system user has the following characteristics:
//   - Fixed UUID: 00000000-0000-0000-0000-000000000000
//   - Username: "system"
//   - Email: "system@internal"
//   - No password (password_hash is empty)
//   - is_system: true
//   - role: platform_user
//   - Cannot login through any API
//
// Returns:
//   - *ent.User: the system user object
//   - error: ErrSystemUserCorruption if data is inconsistent, or database error
func (s *AuthService) EnsureSystemUser(ctx context.Context) (*ent.User, error) {
	return s.ensureSystemUserWithRetry(ctx, 0)
}

const maxRetryAttempts = 3

// ensureSystemUserWithRetry implements EnsureSystemUser with retry limit
func (s *AuthService) ensureSystemUserWithRetry(ctx context.Context, attempt int) (*ent.User, error) {
	if attempt >= maxRetryAttempts {
		return nil, errors.New(errors.ErrCodeInternalError, "max retry attempts reached for system user creation")
	}
	// Try to find existing system user by username
	existingUser, err := s.userRepo.GetByUsername(ctx, "system")
	if err == nil {
		// User exists - verify consistency
		if !existingUser.IsSystem {
			logger.Error("System user found but is_system=false",
				logger.Field{Key: "actual_uuid", Value: existingUser.UUID.String()},
				logger.Field{Key: "username", Value: existingUser.Username})
			return nil, ErrSystemUserCorruption.WithContext("reason", "is_system flag is false")
		}

		if existingUser.UUID != SystemUserUUID {
			logger.Error("System user found but UUID mismatch",
				logger.Field{Key: "expected_uuid", Value: SystemUserUUID.String()},
				logger.Field{Key: "actual_uuid", Value: existingUser.UUID.String()})
			return nil, ErrSystemUserCorruption.WithContext("reason", "UUID mismatch")
		}

		logger.Debug("System user already exists",
			logger.Field{Key: "system_uuid", Value: existingUser.UUID.String()})
		return existingUser, nil
	}

	// User doesn't exist - create it
	logger.Info("Creating system user",
		logger.Field{Key: "uuid", Value: SystemUserUUID.String()})

	username := "system"
	isSystem := true
	role := "platform_user"
	createdUser, err := s.userRepo.CreateUser(ctx, &repository.CreateUserParams{
		UUID:         &SystemUserUUID,
		Email:        "system@internal",
		PasswordHash: "", // No password for system user
		Username:     &username,
		IsSystem:     &isSystem,
		Role:         &role,
	})

	if err != nil {
		// Check if it's a uniqueness conflict - another process may have created it
		if stdErrors.Is(err, repository.ErrEmailExists) || stdErrors.Is(err, repository.ErrUsernameExists) {
			logger.Warn("System user creation conflict, retrying query",
				logger.Field{Key: "attempt", Value: attempt},
				logger.Field{Key: "error", Value: err.Error()})
			// Retry with incremented attempt counter
			return s.ensureSystemUserWithRetry(ctx, attempt+1)
		}

		logger.Error("Failed to create system user",
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to create system user")
	}

	logger.Info("System user created successfully",
		logger.Field{Key: "uuid", Value: createdUser.UUID.String()})

	return createdUser, nil
}

// EnsureAdminUser ensures an admin user exists with the given email and password.
// This method is idempotent - if a user with the given email already exists, it will
// return that user WITHOUT modifying the password or any other fields.
//
// Parameters:
//   - email: admin email address (must be valid format)
//   - password: admin password (must meet strength requirements)
//
// Returns:
//   - *ent.User: the admin user object (without password_hash field)
//   - error: validation error or database error
func (s *AuthService) EnsureAdminUser(ctx context.Context, email, pwd string) (*ent.User, error) {
	return s.ensureAdminUserWithRetry(ctx, email, pwd, 0)
}

// ensureAdminUserWithRetry implements EnsureAdminUser with retry limit
func (s *AuthService) ensureAdminUserWithRetry(ctx context.Context, email, pwd string, attempt int) (*ent.User, error) {
	if attempt >= maxRetryAttempts {
		return nil, errors.New(errors.ErrCodeInternalError, "max retry attempts reached for admin user creation")
	}
	// 1. Validate email format
	if email == "" {
		return nil, ErrInvalidEmail.WithContext("field", "email")
	}

	if !isValidEmail(email) {
		logger.Warn("Invalid admin email format",
			logger.Field{Key: "email", Value: email})
		return nil, ErrInvalidEmail.WithContext("email", email)
	}

	// 2. Validate password strength
	if pwd == "" {
		return nil, ErrWeakPassword.WithContext("reason", "password is empty")
	}

	if err := password.Validate(pwd); err != nil {
		logger.Warn("Weak admin password",
			logger.Field{Key: "email", Value: email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.New(errors.ErrCodeAuthWeakPassword, err.Error())
	}

	// 3. Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		// User exists - return without modification
		logger.Info("Admin user already exists",
			logger.Field{Key: "email", Value: email},
			logger.Field{Key: "uuid", Value: existingUser.UUID.String()})
		return existingUser, nil
	}

	// 4. User doesn't exist - create new admin user
	logger.Info("Creating admin user",
		logger.Field{Key: "email", Value: email})

	// Hash the password
	passwordHash, err := password.Hash(pwd)
	if err != nil {
		logger.Error("Failed to hash admin password",
			logger.Field{Key: "email", Value: email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to hash admin password")
	}

	// Extract username from email (part before @) or use default
	username := "admin"
	if idx := strings.IndexByte(email, '@'); idx > 0 {
		username = email[:idx]
	}

	role := "platform_admin"
	createdUser, err := s.userRepo.CreateUser(ctx, &repository.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Username:     &username,
		Role:         &role,
	})

	if err != nil {
		// Check if it's a uniqueness conflict
		if stdErrors.Is(err, repository.ErrEmailExists) || stdErrors.Is(err, repository.ErrUsernameExists) {
			logger.Warn("Admin user creation conflict, user may already exist",
				logger.Field{Key: "email", Value: email},
				logger.Field{Key: "attempt", Value: attempt})
			// Retry with incremented attempt counter
			return s.ensureAdminUserWithRetry(ctx, email, pwd, attempt+1)
		}

		logger.Error("Failed to create admin user",
			logger.Field{Key: "email", Value: email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to create admin user")
	}

	logger.Info("Admin user created successfully",
		logger.Field{Key: "email", Value: email},
		logger.Field{Key: "uuid", Value: createdUser.UUID.String()})

	return createdUser, nil
}
