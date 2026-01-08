package jwt

import "github.com/golang-jwt/jwt/v5"

// Claims defines custom JWT claims structure
// Extends jwt.RegisteredClaims with user-specific fields
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}
