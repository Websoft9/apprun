// Package repository provides data access layer for authentication module.
package repository

import (
	"context"
	"errors"
	"time"

	"apprun/ent"
	"apprun/ent/user"
)

var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailExists is returned when email already exists
	ErrEmailExists = errors.New("email already registered")
	// ErrUsernameExists is returned when username already exists
	ErrUsernameExists = errors.New("username already taken")
)

// UserRepository handles user data persistence operations.
type UserRepository struct {
	client *ent.Client
}

// NewUserRepository creates a new user repository instance.
func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

// CreateUser creates a new user record in the database.
//
// Parameters:
//   - ctx: context for cancellation and timeout
//   - email: user email (required, unique)
//   - passwordHash: bcrypt hash of the password
//   - username: optional username (unique if provided)
//   - nickname: optional display name
//   - phone: optional phone number
//   - gender: gender code (0=unknown, 1=male, 2=female)
//   - timezone: user timezone (default: UTC)
//   - language: preferred language (default: zh-CN)
//
// Returns created user or error (ErrEmailExists, ErrUsernameExists)
func (r *UserRepository) CreateUser(ctx context.Context, params *CreateUserParams) (*ent.User, error) {
	builder := r.client.User.Create().
		SetEmail(params.Email).
		SetPasswordHash(params.PasswordHash).
		SetNillableUsername(params.Username).
		SetNillableNickname(params.Nickname).
		SetNillablePhone(params.Phone).
		SetNillableTimezone(params.Timezone).
		SetNillableLanguage(params.Language)

	if params.Gender != nil {
		builder.SetGender(*params.Gender)
	}

	userRecord, err := builder.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			// Check which constraint was violated
			if containsString(err.Error(), "email") {
				return nil, ErrEmailExists
			}
			if containsString(err.Error(), "username") {
				return nil, ErrUsernameExists
			}
		}
		return nil, err
	}

	return userRecord, nil
}

// GetByEmail retrieves a user by email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	userRecord, err := r.client.User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return userRecord, nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	userRecord, err := r.client.User.Query().
		Where(user.UsernameEQ(username)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return userRecord, nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*ent.User, error) {
	userRecord, err := r.client.User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return userRecord, nil
}

// EmailExists checks if an email is already registered.
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.client.User.Query().
		Where(user.EmailEQ(email)).
		Exist(ctx)
}

// UsernameExists checks if a username is already taken.
func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	return r.client.User.Query().
		Where(user.UsernameEQ(username)).
		Exist(ctx)
}

// FindByIdentifier finds a user by username OR email.
// This is used for login where the identifier can be either.
func (r *UserRepository) FindByIdentifier(ctx context.Context, identifier string) (*ent.User, error) {
	userRecord, err := r.client.User.Query().
		Where(user.Or(
			user.UsernameEQ(identifier),
			user.EmailEQ(identifier),
		)).
		Only(ctx)
	if err != nil {
		return nil, err // Caller will check ent.IsNotFound
	}
	return userRecord, nil
}

// UpdateLoginHistory updates the user's last login timestamp and IP address.
func (r *UserRepository) UpdateLoginHistory(ctx context.Context, userID int64, clientIP string) error {
	return r.client.User.UpdateOneID(userID).
		SetLastLoginAt(time.Now()).
		SetLastLoginIP(clientIP).
		Exec(ctx)
}

// UpdatePassword updates a user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	return r.client.User.UpdateOneID(userID).
		SetPasswordHash(passwordHash).
		Exec(ctx)
}

// CreateUserParams holds parameters for user creation.
type CreateUserParams struct {
	Email        string
	PasswordHash string
	Username     *string
	Nickname     *string
	Phone        *string
	Gender       *int8
	Timezone     *string
	Language     *string
}

// containsString is a helper to check if a string contains a substring (case-insensitive).
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
