package service

import (
	"context"
	"strings"
	"time"

	"apprun/ent"
	"apprun/ent/schema"
	"apprun/internal/jwt"
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
	// ErrInvalidCredentials is returned for failed login attempts
	ErrInvalidCredentials = errors.New(errors.ErrCodeAuthInvalidCredentials, "Invalid email or password")
	// ErrAccountDisabled is returned when user account is disabled
	ErrAccountDisabled = errors.New(errors.ErrCodeAuthAccountDisabled, "Account has been disabled")
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo       *repository.UserRepository
	projectService *ProjectService
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo *repository.UserRepository, projectService *ProjectService) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		projectService: projectService,
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

// LoginRequest holds user login credentials.
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required" example:"user@example.com"` // Username or email
	Password   string `json:"password" binding:"required" example:"SecurePass123"`      // Password
}

// LoginResponse holds login result with JWT tokens.
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`  // JWT access token (short-lived)
	RefreshToken string      `json:"refresh_token"` // JWT refresh token (long-lived)
	ExpiresIn    int64       `json:"expires_in"`    // Seconds until access token expires
	User         UserProfile `json:"user"`          // User profile data
}

// UserProfile represents sanitized user data for API responses.
type UserProfile struct {
	UUID      string  `json:"id"`
	Email     string  `json:"email"`
	Username  *string `json:"username,omitempty"`
	Nickname  *string `json:"nickname,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Gender    int8    `json:"gender"`
	Status    int8    `json:"status"`
	Timezone  string  `json:"timezone"`
	Language  string  `json:"language"`
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

	// 7. Create personal project for the user (non-blocking)
	if s.projectService != nil {
		username := user.Email
		if user.Username != "" {
			username = user.Username
		}
		_, err := s.projectService.CreatePersonalProject(ctx, user.ID, username)
		if err != nil {
			// Log error but don't fail registration
			logger.Warn("Failed to create personal project for user",
				logger.Field{Key: "user_id", Value: user.ID},
				logger.Field{Key: "error", Value: err.Error()})
		} else {
			logger.Info("Personal project created for user", logger.Field{Key: "user_id", Value: user.ID})
		}
	}

	// 8. Build response (exclude sensitive fields)
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

// Login authenticates a user and generates a JWT token.
//
// Business Logic:
//  1. Parse identifier (email or username)
//  2. Query user by identifier (username OR email)
//  3. Verify password using bcrypt
//  4. Check user status (must be active)
//  5. Generate JWT token
//  6. Update login history (async)
//  7. Return token + user profile
//
// Returns:
//   - LoginResponse: JWT token, expiration, and user profile
//   - error: ErrInvalidCredentials (generic) or ErrAccountDisabled
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, clientIP string) (*LoginResponse, error) {
	logger.Info("User login attempt", logger.Field{Key: "identifier", Value: req.Identifier})

	// 1. Query user by identifier (username or email)
	user, err := s.userRepo.FindByIdentifier(ctx, req.Identifier)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("User not found", logger.Field{Key: "identifier", Value: req.Identifier})
			return nil, ErrInvalidCredentials // Generic error - don't reveal existence
		}
		logger.Error("Failed to query user",
			logger.Field{Key: "identifier", Value: req.Identifier},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to query user")
	}

	// 2. Verify password
	if verifyErr := password.Verify(req.Password, user.PasswordHash); verifyErr != nil {
		logger.Warn("Invalid password",
			logger.Field{Key: "user_id", Value: user.ID},
			logger.Field{Key: "identifier", Value: req.Identifier})
		return nil, ErrInvalidCredentials // Generic error - don't reveal password wrong
	}

	// 3. Check user status (active)
	if user.Status != schema.UserStatusActive {
		logger.Warn("Disabled account login attempt",
			logger.Field{Key: "user_id", Value: user.ID},
			logger.Field{Key: "status", Value: user.Status})
		return nil, ErrAccountDisabled
	}

	// 4. Generate JWT token pair
	accessToken, refreshToken, expiresAt, err := s.generateTokenPair(user)
	if err != nil {
		logger.Error("Failed to generate JWT token pair",
			logger.Field{Key: "user_id", Value: user.ID},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to generate tokens")
	}

	// 5. Update login history (non-blocking)
	go s.updateLoginHistory(context.Background(), user.ID, clientIP)

	logger.Info("User logged in successfully",
		logger.Field{Key: "user_id", Value: user.ID},
		logger.Field{Key: "email", Value: user.Email})

	// 6. Build response
	expiresIn := int64(time.Until(expiresAt).Seconds())
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User:         buildUserProfile(user),
	}, nil
}

// generateTokenPair creates both access and refresh JWT tokens for the user.
func (s *AuthService) generateTokenPair(user *ent.User) (string, string, time.Time, error) {
	userClaims := map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}
	return jwt.GenerateTokenPair(user.ID, userClaims)
}

// updateLoginHistory updates last_login_at and last_login_ip (async).
func (s *AuthService) updateLoginHistory(ctx context.Context, userID int64, clientIP string) {
	if err := s.userRepo.UpdateLoginHistory(ctx, userID, clientIP); err != nil {
		logger.Error("Failed to update login history",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "error", Value: err.Error()})
	}
}

// buildUserProfile converts ent.User to UserProfile (sanitized).
func buildUserProfile(user *ent.User) UserProfile {
	var username, nickname, phone *string
	if user.Username != "" {
		username = &user.Username
	}
	if user.Nickname != "" {
		nickname = &user.Nickname
	}
	if user.Phone != "" {
		phone = &user.Phone
	}

	return UserProfile{
		UUID:      user.UUID.String(),
		Email:     user.Email,
		Username:  username,
		Nickname:  nickname,
		Phone:     phone,
		Gender:    user.Gender,
		Status:    user.Status,
		Timezone:  user.Timezone,
		Language:  user.Language,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GetUserProfile retrieves a user's profile by ID.
func (s *AuthService) GetUserProfile(ctx context.Context, userID int64) (*UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New(errors.ErrCodeAuthUserNotFound, "User not found")
		}
		logger.Error("Failed to get user by ID",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "error", Value: err.Error()})
		return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to query user")
	}

	profile := buildUserProfile(user)
	return &profile, nil
}
