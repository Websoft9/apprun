package services
package services

import (
	"context"
	"encoding/json"
	"time"

	"gofr.dev/pkg/gofr"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Websoft9/apprun/core/internal/models"
)

// AuthService 认证服务
type AuthService struct{}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	TenantName string `json:"tenant_name"` // 可选，如果不提供则使用邮箱域名
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         models.User `json:"user"`
}

// Register 用户注册
func (s *AuthService) Register(ctx *gofr.Context, req RegisterRequest) (*AuthResponse, error) {
	// 1. 检查邮箱是否已存在
	var existingUser models.User
	result := ctx.GORM().Where("email = ?", req.Email).First(&existingUser)
	if result.Error == nil {
		return nil, &gofr.HTTP{
			StatusCode: 409,
			Message:    "email already exists",
		}
	}

	// 2. 创建租户（如果是新租户）
	var tenant models.Tenant
	if req.TenantName != "" {
		tenant.Name = req.TenantName
	} else {
		// 使用邮箱域名作为租户名
		tenant.Name = req.Email
	}
	tenant.DisplayName = tenant.Name
	tenant.Status = "active"
	tenant.Plan = "free"

	if err := ctx.GORM().Create(&tenant).Error; err != nil {
		return nil, err
	}

	// 3. 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 4. 创建用户
	user := models.User{
		TenantID:     tenant.ID,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
		Status:       "active",
	}

	if err := ctx.GORM().Create(&user).Error; err != nil {
		return nil, err
	}

	// 5. 生成 JWT Token
	accessToken, err := s.generateAccessToken(user.ID, user.TenantID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// 6. 发布用户注册事件（NATS）
	event := map[string]interface{}{
		"event_type": "user.registered",
		"timestamp":  time.Now(),
		"data": map[string]interface{}{
			"user_id":   user.ID,
			"email":     user.Email,
			"tenant_id": user.TenantID,
		},
	}
	eventData, _ := json.Marshal(event)
	ctx.PubSub.Publish(ctx, "user.registered", eventData)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
		User:         user,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx *gofr.Context, req LoginRequest) (*AuthResponse, error) {
	// 1. 查找用户
	var user models.User
	result := ctx.GORM().Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		return nil, &gofr.HTTP{
			StatusCode: 401,
			Message:    "invalid email or password",
		}
	}

	// 2. 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, &gofr.HTTP{
			StatusCode: 401,
			Message:    "invalid email or password",
		}
	}

	// 3. 检查用户状态
	if user.Status != "active" {
		return nil, &gofr.HTTP{
			StatusCode: 403,
			Message:    "user account is not active",
		}
	}

	// 4. 生成 JWT Token
	accessToken, err := s.generateAccessToken(user.ID, user.TenantID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// 5. 发布登录事件
	event := map[string]interface{}{
		"event_type": "user.logged_in",
		"timestamp":  time.Now(),
		"data": map[string]interface{}{
			"user_id":   user.ID,
			"email":     user.Email,
			"tenant_id": user.TenantID,
		},
	}
	eventData, _ := json.Marshal(event)
	ctx.PubSub.Publish(ctx, "user.logged_in", eventData)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
		User:         user,
	}, nil
}

// generateAccessToken 生成访问令牌
func (s *AuthService) generateAccessToken(userID, tenantID uint, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":       userID,
		"tenant_id": tenantID,
		"email":     email,
		"exp":       time.Now().Add(time.Hour * 1).Unix(), // 1 hour
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key")) // TODO: 从配置读取
}

// generateRefreshToken 生成刷新令牌
func (s *AuthService) generateRefreshToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-refresh-secret")) // TODO: 从配置读取
}
