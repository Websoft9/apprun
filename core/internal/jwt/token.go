package jwt

import (
	"errors"
	"time"

	pkgconfig "apprun/pkg/config"
	pkgjwt "apprun/pkg/jwt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken indicates the token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired indicates the token has expired
	ErrTokenExpired = errors.New("token has expired")
	// ErrMissingSecret indicates JWT secret is not configured
	ErrMissingSecret = errors.New("JWT secret is not configured")
)

// TokenService handles JWT token operations with configurable backend.
type TokenService struct {
	config pkgconfig.Provider
}

// NewTokenService creates a new token service with the given config provider.
func NewTokenService(cfg pkgconfig.Provider) *TokenService {
	return &TokenService{config: cfg}
}

// GenerateToken generates a JWT token with user claims.
// Returns the token string, expiration time, and any error.
func (s *TokenService) GenerateToken(userID int64, userClaims map[string]interface{}) (string, time.Time, error) {
	secret := s.config.GetString("jwt.secret")
	if secret == "" {
		return "", time.Time{}, ErrMissingSecret
	}

	expiresIn := s.config.GetDuration("jwt.access_token_expiration")
	if expiresIn == 0 {
		expiresIn = pkgjwt.DefaultExpiry // Fallback to 24h
	}
	expiresAt := time.Now().Add(expiresIn)

	claims := CustomClaims{
		UserID:   userID,
		Username: userClaims["username"].(string),
		Email:    userClaims["email"].(string),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.GetString("jwt.issuer"),
			Audience:  jwt.ClaimStrings{s.config.GetString("jwt.audience")},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	return tokenString, expiresAt, err
}

// ValidateToken validates a JWT token and returns the custom claims.
func (s *TokenService) ValidateToken(tokenString string) (*CustomClaims, error) {
	secret := s.config.GetString("jwt.secret")
	if secret == "" {
		return nil, ErrMissingSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ============================================================================
// Backward Compatibility - Package-level functions (Deprecated)
// ============================================================================

var globalTokenService *TokenService

// InitGlobalService initializes the global token service with a config provider.
// This is for backward compatibility with code that uses package-level functions.
func InitGlobalService(cfg pkgconfig.Provider) {
	globalTokenService = NewTokenService(cfg)
}

// GenerateToken is a backward-compatible package-level function.
// Deprecated: Use TokenService.GenerateToken instead.
func GenerateToken(userID int64, userClaims map[string]interface{}) (string, time.Time, error) {
	if globalTokenService == nil {
		// Fallback to Viper for backward compatibility
		globalTokenService = NewTokenService(pkgconfig.NewViperProvider(nil))
	}
	return globalTokenService.GenerateToken(userID, userClaims)
}

// ValidateToken is a backward-compatible package-level function.
// Deprecated: Use TokenService.ValidateToken instead.
func ValidateToken(tokenString string) (*CustomClaims, error) {
	if globalTokenService == nil {
		// Fallback to Viper for backward compatibility
		globalTokenService = NewTokenService(pkgconfig.NewViperProvider(nil))
	}
	return globalTokenService.ValidateToken(tokenString)
}
