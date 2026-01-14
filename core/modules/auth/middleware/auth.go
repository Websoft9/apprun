// Package middleware provides HTTP middleware components for authentication and authorization.
// This package includes JWT authentication middleware that validates tokens and injects user context.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"apprun/internal/jwt"
	"apprun/pkg/errors"
	"apprun/pkg/response"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyUsername is the context key for username
	ContextKeyUsername contextKey = "username"
	// ContextKeyEmail is the context key for email
	ContextKeyEmail contextKey = "email"
)

// JWTMiddleware provides JWT authentication middleware
type JWTMiddleware struct{}

// NewJWTMiddleware creates a JWT middleware instance
func NewJWTMiddleware() *JWTMiddleware {
	return &JWTMiddleware{}
}

// JWTAuth is the JWT authentication middleware handler
func (m *JWTMiddleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.ErrorWithRequest(w, r, http.StatusUnauthorized, errors.ErrCodeAuthInvalidToken, "Missing authorization token")
			return
		}

		// Verify Bearer format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.ErrorWithRequest(w, r, http.StatusUnauthorized, errors.ErrCodeAuthInvalidToken, "Invalid authorization format")
			return
		}

		tokenString := parts[1]

		// Validate token using viper-based JWT package
		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			response.ErrorWithRequest(w, r, http.StatusUnauthorized, errors.ErrCodeAuthInvalidToken, "Invalid or expired token")
			return
		}

		// Inject context with user information
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeyUsername, claims.Username)
		ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
