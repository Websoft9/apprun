package jwt

import (
	"errors"
	"time"

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

// GenerateToken generates a JWT token with user claims
func GenerateToken(cfg *RuntimeConfig, userID int64, username, email string) (string, error) {
	if cfg.Secret == "" {
		return "", ErrMissingSecret
	}

	now := time.Now()
	expiresAt := now.Add(cfg.Expiry)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidateToken validates a JWT token and returns claims
func ValidateToken(cfg *RuntimeConfig, tokenString string) (*Claims, error) {
	if cfg.Secret == "" {
		return nil, ErrMissingSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
