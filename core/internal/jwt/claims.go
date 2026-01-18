// Package jwt provides JWT token generation and validation utilities
package jwt

import "github.com/golang-jwt/jwt/v5"

// CustomClaims represents the custom claims structure for JWT tokens
// Used for authentication and authorization throughout the apprun platform
type CustomClaims struct {
	UserID               int64  `json:"user_id"`    // User's unique identifier
	Username             string `json:"username"`   // User's username
	Email                string `json:"email"`      // User's email address
	TokenType            string `json:"token_type"` // Token type: "access" or "refresh"
	jwt.RegisteredClaims        // Standard JWT claims (exp, iat, iss, aud)
}
