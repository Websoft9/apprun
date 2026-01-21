// Package service provides business logic for user management operations.
package service

import (
	"context"
	"time"

	"apprun/ent"
	"apprun/ent/user"
	"apprun/internal/password"
	admin "apprun/modules/admin"
	"apprun/modules/obs"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
)

// UserMgmtService handles platform user management operations (Story 5.7)
type UserMgmtService struct {
	client *ent.Client
}

// NewUserMgmtService creates a new user management service
func NewUserMgmtService(client *ent.Client) *UserMgmtService {
	return &UserMgmtService{client: client}
}

// ListUsersRequest holds filters for listing users
type ListUsersRequest struct {
	Page     int    `json:"page" form:"page"`           // Page number (default: 1)
	PageSize int    `json:"page_size" form:"page_size"` // Items per page (default: 20)
	Search   string `json:"search" form:"search"`       // Search term (email, nickname)
	Role     string `json:"role" form:"role"`           // Filter by role (platform_admin, platform_user)
	Status   *int8  `json:"status" form:"status"`       // Filter by status (0=disabled, 1=active)
}

// ListUsersResponse holds paginated user list
type ListUsersResponse struct {
	Users    []UserSummary `json:"users"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// UserSummary represents a user in the list view
type UserSummary struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname,omitempty"` // Plain string, not pointer (Ent schema)
	Role      string `json:"role"`
	Status    int8   `json:"status"`
	IsSystem  bool   `json:"is_system"`
	CreatedAt string `json:"created_at"`
}

// CreateUserRequest holds data for creating a new user
type CreateUserRequest struct {
	Email    string  `json:"email" binding:"required"`
	Name     *string `json:"name"`
	Password string  `json:"password"` // Optional, will be auto-generated if empty
	Role     string  `json:"role" binding:"required"`
}

// ChangeUserRoleRequest holds data for changing user role
type ChangeUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// ChangeUserStatusRequest holds data for changing user status
type ChangeUserStatusRequest struct {
	Status int8 `json:"status" binding:"required"`
}

// ListUsers retrieves a paginated list of users with optional filters
func (s *UserMgmtService) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Build query
	query := s.client.User.Query().Where(user.DeletedAtIsNil()) // Exclude soft-deleted users

	// Apply filters
	if req.Search != "" {
		query = query.Where(
			user.Or(
				user.EmailContains(req.Search),
				user.NicknameContains(req.Search),
			),
		)
	}

	if req.Role != "" {
		query = query.Where(user.RoleEQ(req.Role))
	}

	if req.Status != nil {
		query = query.Where(user.StatusEQ(*req.Status))
	}

	// Count total
	total, err := query.Count(ctx)
	if err != nil {
		logger.Error("Failed to count users",
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to count users")
	}

	// Fetch paginated results
	offset := (req.Page - 1) * req.PageSize
	users, err := query.
		Order(ent.Desc(user.FieldCreatedAt)).
		Limit(req.PageSize).
		Offset(offset).
		All(ctx)
	if err != nil {
		logger.Error("Failed to list users",
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to list users")
	}

	// Convert to summary
	summaries := make([]UserSummary, len(users))
	for i, u := range users {
		summaries[i] = UserSummary{
			ID:        u.ID,
			Email:     u.Email,
			Nickname:  u.Nickname, // Plain string from Ent
			Role:      u.Role,
			Status:    u.Status,
			IsSystem:  u.IsSystem,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
		}
	}

	return &ListUsersResponse{
		Users:    summaries,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetUserByID retrieves a single user by ID
func (s *UserMgmtService) GetUserByID(ctx context.Context, userID int64) (*ent.User, error) {
	u, err := s.client.User.Query().
		Where(user.ID(userID), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrCodeNotFound, "User not found")
		}
		logger.Error("Failed to get user",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to get user")
	}
	return u, nil
}

// CreateUser creates a new user (Story 5.7)
// Returns the created user, generated password (if auto-generated), and error
func (s *UserMgmtService) CreateUser(ctx context.Context, req *CreateUserRequest) (*ent.User, string, error) {
	// Validate email format
	if req.Email == "" {
		return nil, "", errors.New(errors.ErrCodeInvalidParam, "Email is required")
	}

	// Validate role
	if !admin.IsValidRole(req.Role) {
		return nil, "", errors.New(errors.ErrCodeInvalidParam, "Invalid role")
	}

	// Check if email already exists
	exists, err := s.client.User.Query().Where(user.Email(req.Email)).Exist(ctx)
	if err != nil {
		logger.Error("Failed to check email existence",
			logger.Field{Key: "email", Value: req.Email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, "", errors.Wrap(err, errors.ErrCodeInternalError, "Failed to check email")
	}
	if exists {
		return nil, "", errors.ErrAdminEmailExists
	}

	// Generate password if not provided
	var generatedPassword string
	pwd := req.Password
	if pwd == "" {
		var genErr error
		pwd, genErr = password.GenerateRandomPassword(16)
		if genErr != nil {
			logger.Error("Failed to generate password",
				logger.Field{Key: "error", Value: genErr.Error()})
			return nil, "", errors.Wrap(genErr, errors.ErrCodeInternalError, "Failed to generate password")
		}
		generatedPassword = pwd // Save for return
	}

	// Validate password strength
	if validateErr := password.Validate(pwd); validateErr != nil {
		return nil, "", errors.Wrap(validateErr, errors.ErrCodeInvalidParam, "Password does not meet requirements")
	}

	// Hash password
	passwordHash, err := password.Hash(pwd)
	if err != nil {
		logger.Error("Failed to hash password",
			logger.Field{Key: "email", Value: req.Email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, "", errors.Wrap(err, errors.ErrCodeInternalError, "Failed to hash password")
	}

	// Create user
	nickname := req.Name
	builder := s.client.User.Create().
		SetEmail(req.Email).
		SetPasswordHash(passwordHash).
		SetRole(req.Role).
		SetStatus(1). // Active by default
		SetTokenVersion(0)

	if nickname != nil && *nickname != "" {
		builder = builder.SetNickname(*nickname)
	}

	newUser, err := builder.Save(ctx)
	if err != nil {
		logger.Error("Failed to create user",
			logger.Field{Key: "email", Value: req.Email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, "", errors.Wrap(err, errors.ErrCodeInternalError, "Failed to create user")
	}

	logger.Info("User created by admin",
		logger.Field{Key: "user_id", Value: newUser.ID},
		logger.Field{Key: "email", Value: newUser.Email},
		logger.Field{Key: "role", Value: newUser.Role})

	// Metrics
	obs.UserCreationCounter.Inc()

	return newUser, generatedPassword, nil
}

// ChangeUserRole changes a user's role with safety checks (Story 5.7)
// Uses database transaction to prevent race conditions
func (s *UserMgmtService) ChangeUserRole(ctx context.Context, targetUserID, operatorUserID int64, newRole string) (*ent.User, error) {
	// Validate role
	if !admin.IsValidRole(newRole) {
		return nil, errors.New(errors.ErrCodeInvalidParam, "Invalid role")
	}

	// Start transaction for concurrent safety
	tx, err := s.client.Tx(ctx)
	if err != nil {
		logger.Error("Failed to start transaction",
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback() //nolint:errcheck // Rollback on panic is best-effort
			panic(r)
		}
	}()

	// Get target user in transaction
	targetUser, err := tx.User.Query().
		Where(user.ID(targetUserID), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		tx.Rollback() //nolint:errcheck // Error already being returned
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrCodeNotFound, "User not found")
		}
		logger.Error("Failed to get user",
			logger.Field{Key: "user_id", Value: targetUserID},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to get user")
	}

	// Prevent modifying system users
	if targetUser.IsSystem {
		tx.Rollback() //nolint:errcheck // Error already being returned
		return nil, errors.ErrAdminCannotModifySystem
	}

	// If demoting from platform_admin, check if this is the last admin
	if targetUser.Role == admin.RolePlatformAdmin && newRole != admin.RolePlatformAdmin {
		adminCount, countErr := tx.User.Query().
			Where(user.RoleEQ(admin.RolePlatformAdmin), user.DeletedAtIsNil()).
			Count(ctx)
		if countErr != nil {
			tx.Rollback() //nolint:errcheck // Error already being returned
			logger.Error("Failed to count platform admins",
				logger.Field{Key: "error", Value: countErr.Error()})
			return nil, errors.Wrap(countErr, errors.ErrCodeInternalError, "Failed to verify admin count")
		}

		if adminCount <= 1 {
			tx.Rollback() //nolint:errcheck // Error already being returned
			return nil, errors.ErrAdminCannotDemoteLastAdmin
		}
	}

	// Update role and increment token version (revoke existing tokens)
	updatedUser, err := tx.User.UpdateOneID(targetUserID).
		SetRole(newRole).
		AddTokenVersion(1).
		Save(ctx)
	if err != nil {
		tx.Rollback() //nolint:errcheck // Error already being returned
		logger.Error("Failed to update user role",
			logger.Field{Key: "user_id", Value: targetUserID},
			logger.Field{Key: "new_role", Value: newRole},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to update user role")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction",
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to commit transaction")
	}

	logger.Info("User role changed",
		logger.Field{Key: "user_id", Value: targetUserID},
		logger.Field{Key: "old_role", Value: targetUser.Role},
		logger.Field{Key: "new_role", Value: newRole},
		logger.Field{Key: "operator_id", Value: operatorUserID})

	// Metrics: token revoked due to role change
	obs.TokenRevocationCounter.WithLabelValues("role_change").Inc()

	return updatedUser, nil
}

// ChangeUserStatus changes a user's status with safety checks (Story 5.7)
func (s *UserMgmtService) ChangeUserStatus(ctx context.Context, targetUserID, operatorUserID int64, newStatus int8) (*ent.User, error) {
	// Validate status (0=disabled, 1=active)
	if newStatus != 0 && newStatus != 1 {
		return nil, errors.New(errors.ErrCodeInvalidParam, "Invalid status value")
	}

	// Prevent self-disable
	if targetUserID == operatorUserID && newStatus == 0 {
		return nil, errors.ErrAdminCannotDisableSelf
	}

	// Get target user
	targetUser, err := s.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	// Prevent modifying system users
	if targetUser.IsSystem {
		return nil, errors.ErrAdminCannotModifySystem
	}

	// Update status and increment token version (revoke existing tokens)
	updatedUser, err := s.client.User.UpdateOneID(targetUserID).
		SetStatus(newStatus).
		AddTokenVersion(1).
		Save(ctx)
	if err != nil {
		logger.Error("Failed to update user status",
			logger.Field{Key: "user_id", Value: targetUserID},
			logger.Field{Key: "new_status", Value: newStatus},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to update user status")
	}

	logger.Info("User status changed",
		logger.Field{Key: "user_id", Value: targetUserID},
		logger.Field{Key: "old_status", Value: targetUser.Status},
		logger.Field{Key: "new_status", Value: newStatus},
		logger.Field{Key: "operator_id", Value: operatorUserID})

	// Metrics: token revoked due to status change
	obs.TokenRevocationCounter.WithLabelValues("status_change").Inc()

	return updatedUser, nil
}

// DeleteUser soft-deletes a user with safety checks (Story 5.7)
// Uses database transaction to prevent race conditions
func (s *UserMgmtService) DeleteUser(ctx context.Context, targetUserID, operatorUserID int64) error {
	// Prevent self-deletion
	if targetUserID == operatorUserID {
		return errors.ErrAdminCannotDeleteSelf
	}

	// Start transaction for concurrent safety
	tx, err := s.client.Tx(ctx)
	if err != nil {
		logger.Error("Failed to start transaction",
			logger.Field{Key: "error", Value: err.Error()})
		return errors.Wrap(err, errors.ErrCodeInternalError, "Failed to start transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback() //nolint:errcheck // Rollback on panic is best-effort
			panic(r)
		}
	}()

	// Get target user in transaction
	targetUser, err := tx.User.Query().
		Where(user.ID(targetUserID), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		tx.Rollback() //nolint:errcheck // Error already being returned
		if ent.IsNotFound(err) {
			return errors.New(errors.ErrCodeNotFound, "User not found")
		}
		logger.Error("Failed to get user",
			logger.Field{Key: "user_id", Value: targetUserID},
			logger.Field{Key: "error", Value: err.Error()})
		return errors.Wrap(err, errors.ErrCodeInternalError, "Failed to get user")
	}

	// Prevent deleting system users
	if targetUser.IsSystem {
		tx.Rollback() //nolint:errcheck // Error already being returned
		return errors.ErrAdminCannotModifySystem
	}

	// If deleting a platform_admin, check if this is the last admin
	if targetUser.Role == admin.RolePlatformAdmin {
		countResult, countErr := tx.User.Query().
			Where(user.RoleEQ(admin.RolePlatformAdmin), user.DeletedAtIsNil()).
			Count(ctx)
		if countErr != nil {
			tx.Rollback() //nolint:errcheck // Error already being returned
			logger.Error("Failed to count platform admins",
				logger.Field{Key: "error", Value: countErr.Error()})
			return errors.Wrap(countErr, errors.ErrCodeInternalError, "Failed to verify admin count")
		}

		if countResult <= 1 {
			tx.Rollback() //nolint:errcheck // Error already being returned
			return errors.ErrAdminCannotDeleteLastAdmin
		}
	}

	// Soft delete: set deleted_at, disable status, and increment token version
	now := time.Now()
	_, err = tx.User.UpdateOneID(targetUserID).
		SetDeletedAt(now).
		SetStatus(0).
		AddTokenVersion(1).
		Save(ctx)
	if err != nil {
		tx.Rollback() //nolint:errcheck // Error already being returned
		logger.Error("Failed to delete user",
			logger.Field{Key: "user_id", Value: targetUserID},
			logger.Field{Key: "error", Value: err.Error()})
		return errors.Wrap(err, errors.ErrCodeInternalError, "Failed to delete user")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction",
			logger.Field{Key: "error", Value: err.Error()})
		return errors.Wrap(err, errors.ErrCodeInternalError, "Failed to commit transaction")
	}

	logger.Info("User deleted",
		logger.Field{Key: "user_id", Value: targetUserID},
		logger.Field{Key: "operator_id", Value: operatorUserID})

	// Metrics
	obs.UserDeletionCounter.WithLabelValues("soft_delete").Inc()
	obs.TokenRevocationCounter.WithLabelValues("deletion").Inc()

	return nil
}
