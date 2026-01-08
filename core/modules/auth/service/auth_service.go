// Package service provides business logic for authentication operations.
package service

import (
	"context"
	"strings"

	"apprun/ent"
	"apprun/internal/password"
	"apprun/modules/auth/repository"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
)

var (
	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New(errors.ErrCodeAuthInvalidEmail, "Invalid email format")
	// ErrWeakPassword is returned when password doesn't meet strength requirements
	ErrWeakPassword = errors.New(errors.ErrCodeAuthWeakPassword, "Password does not meet strength requirements")
	// ErrEmailExists wraps repository error
	ErrEmailExists = errors.New(errors.ErrCodeAuthEmailExists, "Email already registered")
	// ErrUsernameExists wraps repository error
	ErrUsernameExists = errors.New(errors.ErrCodeAuthUsernameExists, "Username already taken")
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// RegisterRequest holds user registration data.
type RegisterRequest struct {
	Email    string  `json:"email" binding:"required" example:"user@example.com"` // 用户邮箱（必填）
	Password string  `json:"password" binding:"required" example:"SecurePass123"` // 密码（必填，8+字符，含大小写和数字）
	Username *string `json:"username,omitempty" example:"john_doe"`               // 用户名（可选，3-64字符）
	Nickname *string `json:"nickname,omitempty" example:"John"`                   // 昵称（可选）
	Phone    *string `json:"phone,omitempty" example:"+86-13800138000"`           // 手机号（可选）
	Gender   *int8   `json:"gender,omitempty" example:"1"`                        // 性别（可选：0=未知,1=男,2=女）
	Timezone *string `json:"timezone,omitempty" example:"Asia/Shanghai"`          // 时区（可选）
	Language *string `json:"language,omitempty" example:"zh-CN"`                  // 语言（可选）
}

// RegisterResponse holds registration result.
type RegisterResponse struct {
	UUID      string  `json:"id"` // Public UUID (exposed as "id" for external API)
	Email     string  `json:"email"`
	Username  *string `json:"username,omitempty"`
	Nickname  *string `json:"nickname,omitempty"`
	Status    int8    `json:"status"`
	CreatedAt string  `json:"created_at"`
}

// Register creates a new user account with validation and password hashing.
//
// Business Logic:
//  1. Validate email format
//  2. Validate password strength (8+ chars, upper/lower/digit)
//  3. Validate username format (if provided)
//  4. Check email uniqueness
//  5. Check username uniqueness (if provided)
//  6. Hash password with bcrypt (cost=12)
//  7. Create user record
//
// Returns:
//   - RegisterResponse: user data (excluding password hash)
//   - error: validation or database error
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	logger.Info("User registration attempt", logger.Field{Key: "email", Value: req.Email})

	// 1. Validate email format (basic check)
	if !isValidEmail(req.Email) {
		logger.Warn("Invalid email format", logger.Field{Key: "email", Value: req.Email})
		return nil, ErrInvalidEmail
	}

	// 2. Validate password strength
	if err := password.Validate(req.Password); err != nil {
		logger.Warn("Weak password",
			logger.Field{Key: "email", Value: req.Email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, err
	}

	// 3. Validate username format (if provided)
	if req.Username != nil && *req.Username != "" {
		if err := password.ValidateUsername(*req.Username); err != nil {
			logger.Warn("Invalid username format",
				logger.Field{Key: "username", Value: *req.Username},
				logger.Field{Key: "error", Value: err.Error()})
			return nil, err
		}

		// Check username uniqueness
		exists, err := s.userRepo.UsernameExists(ctx, *req.Username)
		if err != nil {
			logger.Error("Failed to check username existence",
				logger.Field{Key: "username", Value: *req.Username},
				logger.Field{Key: "error", Value: err.Error()})
			return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to check username")
		}
		if exists {
			logger.Warn("Username already exists", logger.Field{Key: "username", Value: *req.Username})
			return nil, ErrUsernameExists.WithContext("username", *req.Username)
		}
	}

	// 4. Check email uniqueness
	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to check email existence",
			logger.Field{Key: "email", Value: req.Email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to check email").
			WithContext(errors.ContextKeyUserID, req.Email)
	}
	if exists {
		logger.Warn("Email already registered", logger.Field{Key: "email", Value: req.Email})
		return nil, ErrEmailExists.WithContext(errors.ContextKeyUserID, req.Email)
	}

	// 5. Hash password
	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		logger.Error("Failed to hash password",
			logger.Field{Key: "email", Value: req.Email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to hash password")
	}

	// 6. Create user
	user, err := s.userRepo.CreateUser(ctx, &repository.CreateUserParams{
		Email:        req.Email,
		PasswordHash: passwordHash,
		Username:     req.Username,
		Nickname:     req.Nickname,
		Phone:        req.Phone,
		Gender:       req.Gender,
		Timezone:     req.Timezone,
		Language:     req.Language,
	})
	if err != nil {
		logger.Error("Failed to create user",
			logger.Field{Key: "email", Value: req.Email},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to create user").
			WithContext(errors.ContextKeyUserID, req.Email)
	}

	logger.Info("User registered successfully",
		logger.Field{Key: "user_uuid", Value: user.UUID.String()},
		logger.Field{Key: "email", Value: user.Email},
		logger.Field{Key: "username", Value: user.Username})

	// 7. Build response (exclude sensitive fields)
	return buildRegisterResponse(user), nil
}

// buildRegisterResponse converts ent.User to RegisterResponse.
func buildRegisterResponse(user *ent.User) *RegisterResponse {
	var username *string
	if user.Username != "" {
		username = &user.Username
	}

	var nickname *string
	if user.Nickname != "" {
		nickname = &user.Nickname
	}

	return &RegisterResponse{
		UUID:      user.UUID.String(),
		Email:     user.Email,
		Username:  username,
		Nickname:  nickname,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// isValidEmail performs basic email format validation.
// Note: Uses simple check for MVP. Consider using a library for production.
func isValidEmail(email string) bool {
	// Basic RFC 5322 check: contains @ and domain
	if len(email) < 3 || len(email) > 255 {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart, domain := parts[0], parts[1]
	if localPart == "" || domain == "" {
		return false
	}

	// Domain must have at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}
