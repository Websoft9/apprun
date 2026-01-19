// Package service provides business logic for user management.
package service

import (
	"context"
	"strings"

	"apprun/ent"
	"apprun/internal/jwt"
	"apprun/internal/password"
	"apprun/modules/auth/repository"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/validator"
)

var (
	// ErrOldPasswordIncorrect is returned when old password doesn't match
	ErrOldPasswordIncorrect = errors.New(errors.ErrCodeAuthInvalidCredentials, "Old password is incorrect")
	// ErrSamePassword is returned when new password equals old password
	ErrSamePassword = errors.New(errors.ErrCodeInvalidParam, "New password must be different from old password")
)

// UserService handles user self-service business logic.
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service.
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// ProfileResponse represents user profile data for API responses.
type ProfileResponse struct {
	ID          string  `json:"id"`                      // Public UUID
	Email       string  `json:"email"`                   // Email address
	Name        *string `json:"name,omitempty"`          // Display name (nickname)
	Avatar      *string `json:"avatar,omitempty"`        // Avatar URL (HTTPS)
	Bio         *string `json:"bio,omitempty"`           // Biography/description
	Role        string  `json:"role"`                    // User role (from schema constants)
	IsActive    bool    `json:"is_active"`               // Account active status
	CreatedAt   string  `json:"created_at"`              // Creation timestamp
	LastLoginAt *string `json:"last_login_at,omitempty"` // Last login timestamp
}

// UpdateProfileRequest holds profile update data.
type UpdateProfileRequest struct {
	Name   *string `json:"name,omitempty"`   // Display name (2-50 chars)
	Avatar *string `json:"avatar,omitempty"` // Avatar URL (HTTPS, ≤255 chars)
	Bio    *string `json:"bio,omitempty"`    // Biography (≤500 chars)
}

// ChangePasswordRequest holds password change data.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"` // Current password
	NewPassword string `json:"new_password" binding:"required"` // New password (≥8 chars, complexity requirements)
}

// GetCurrentUser retrieves the current user's profile.
//
// Business Logic:
//  1. Extract user ID from JWT context
//  2. Query user by ID
//  3. Return sanitized profile data
//
// Returns:
//   - ProfileResponse: complete user profile
//   - error: if user not found or database error
func (s *UserService) GetCurrentUser(ctx context.Context) (*ProfileResponse, error) {
	// 1. Extract user ID from JWT context
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		logger.Warn("Failed to extract user ID from context")
		return nil, errors.New(errors.ErrCodeUnauthorized, "Authentication required")
	}

	logger.Info("Fetching current user profile", logger.Field{Key: "user_id", Value: userID})

	// 2. Query user by ID
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("User not found", logger.Field{Key: "user_id", Value: userID})
			return nil, errors.New(errors.ErrCodeNotFound, "User not found")
		}
		logger.Error("Failed to query user", logger.Field{Key: "user_id", Value: userID}, logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.New(errors.ErrCodeInternalError, "Failed to retrieve user profile")
	}

	// 3. Convert to ProfileResponse
	return s.toProfileResponse(u), nil
}

// UpdateProfile updates the current user's profile information.
//
// Business Logic:
//  1. Extract user ID from JWT context
//  2. Validate input fields (whitelist: name, avatar, bio)
//  3. Query user by ID
//  4. Apply updates (only allowed fields)
//  5. Save to database
//
// Security:
//   - Only allows updating: name, avatar, bio
//   - Prevents updating: email, role, is_active, is_system
//
// Returns:
//   - ProfileResponse: updated user profile
//   - error: validation or database error
func (s *UserService) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*ProfileResponse, error) {
	// 1. Extract user ID from JWT context
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		logger.Warn("Failed to extract user ID from context")
		return nil, errors.New(errors.ErrCodeUnauthorized, "Authentication required")
	}

	logger.Info("Updating user profile", logger.Field{Key: "user_id", Value: userID})

	// 2. Validate input fields
	if req.Name != nil {
		if err := validator.ValidateName(*req.Name, 2, 50); err != nil {
			return nil, err
		}
	}
	if req.Avatar != nil && *req.Avatar != "" {
		if err := validator.ValidateHTTPSURL(*req.Avatar, 255, "avatar"); err != nil {
			return nil, err
		}
	}
	if req.Bio != nil {
		if err := validator.ValidateText(*req.Bio, 500, "bio"); err != nil {
			return nil, err
		}
	}

	// 3. Query user by ID
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("User not found", logger.Field{Key: "user_id", Value: userID})
			return nil, errors.New(errors.ErrCodeNotFound, "User not found")
		}
		logger.Error("Failed to query user", logger.Field{Key: "user_id", Value: userID}, logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.New(errors.ErrCodeInternalError, "Failed to retrieve user")
	}

	// 4. Build update query (only allowed fields)
	updateQuery := u.Update()
	fieldsUpdated := []string{}

	if req.Name != nil {
		updateQuery = updateQuery.SetNillableNickname(req.Name)
		fieldsUpdated = append(fieldsUpdated, "name")
	}
	if req.Avatar != nil {
		updateQuery = updateQuery.SetNillableAvatar(req.Avatar)
		fieldsUpdated = append(fieldsUpdated, "avatar")
	}
	if req.Bio != nil {
		updateQuery = updateQuery.SetNillableBio(req.Bio)
		fieldsUpdated = append(fieldsUpdated, "bio")
	}

	// 5. Save updates
	updatedUser, err := updateQuery.Save(ctx)
	if err != nil {
		logger.Error("Failed to update user profile",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.New(errors.ErrCodeInternalError, "Failed to update profile")
	}

	logger.Info("User profile updated successfully",
		logger.Field{Key: "user_id", Value: userID},
		logger.Field{Key: "fields", Value: strings.Join(fieldsUpdated, ",")})

	return s.toProfileResponse(updatedUser), nil
}

// ChangePassword changes the current user's password.
//
// Business Logic:
//  1. Extract user ID from JWT context
//  2. Query user by ID
//  3. Verify old password
//  4. Validate new password strength
//  5. Check new password != old password
//  6. Hash new password
//  7. Update password in database
//
// Security:
//   - Requires old password verification (prevents unauthorized changes)
//   - New password must meet strength requirements
//   - New password must differ from old password
//   - Does NOT invalidate existing JWT tokens
//
// Returns:
//   - error: if verification fails or database error
func (s *UserService) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	// 1. Extract user ID from JWT context
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		logger.Warn("Failed to extract user ID from context")
		return errors.New(errors.ErrCodeUnauthorized, "Authentication required")
	}

	logger.Info("Password change attempt", logger.Field{Key: "user_id", Value: userID})

	// 2. Query user by ID
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("User not found", logger.Field{Key: "user_id", Value: userID})
			return errors.New(errors.ErrCodeNotFound, "User not found")
		}
		logger.Error("Failed to query user", logger.Field{Key: "user_id", Value: userID}, logger.Field{Key: "error", Value: err.Error()})
		return errors.New(errors.ErrCodeInternalError, "Failed to retrieve user")
	}

	// 3. Verify old password
	if verifyErr := password.Verify(req.OldPassword, u.PasswordHash); verifyErr != nil {
		logger.Warn("Old password verification failed", logger.Field{Key: "user_id", Value: userID})
		return ErrOldPasswordIncorrect
	}

	// 4. Validate new password strength
	if validateErr := password.Validate(req.NewPassword); validateErr != nil {
		logger.Warn("New password doesn't meet requirements",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "error", Value: validateErr.Error()})
		return validateErr
	}

	// 5. Check new password != old password
	if req.OldPassword == req.NewPassword {
		logger.Warn("New password is same as old password", logger.Field{Key: "user_id", Value: userID})
		return ErrSamePassword
	}

	// 6. Hash new password
	hashedPassword, err := password.Hash(req.NewPassword)
	if err != nil {
		logger.Error("Failed to hash new password",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "error", Value: err.Error()})
		return errors.New(errors.ErrCodeInternalError, "Failed to process password")
	}

	// 7. Update password in database
	if err := s.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		logger.Error("Failed to update password",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "error", Value: err.Error()})
		return errors.New(errors.ErrCodeInternalError, "Failed to update password")
	}

	logger.Info("Password changed successfully", logger.Field{Key: "user_id", Value: userID})
	return nil
}

// toProfileResponse converts ent.User to ProfileResponse.
func (s *UserService) toProfileResponse(u *ent.User) *ProfileResponse {
	resp := &ProfileResponse{
		ID:        u.UUID.String(),
		Email:     u.Email,
		Name:      &u.Nickname,
		Avatar:    &u.Avatar,
		Bio:       &u.Bio,
		Role:      "platform_user", // TODO: Add role field to User schema (Story 5.5) and map from u.Role
		IsActive:  u.Status == 1,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Handle optional fields
	if u.LastLoginAt != nil {
		lastLogin := u.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		resp.LastLoginAt = &lastLogin
	}

	// Clear empty optional fields
	if u.Nickname == "" {
		resp.Name = nil
	}
	if u.Avatar == "" {
		resp.Avatar = nil
	}
	if u.Bio == "" {
		resp.Bio = nil
	}

	return resp
}

// Note: Validation functions moved to pkg/validator package for reusability
