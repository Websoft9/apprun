// Package middleware provides HTTP middleware components for authentication and authorization.
// This package includes JWT authentication middleware that validates tokens and injects user context.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"apprun/internal/jwt"
	"apprun/pkg/response"
)

// JWTMiddleware provides JWT authentication middleware
type JWTMiddleware struct {
	jwtConfig *jwt.RuntimeConfig
}

// NewJWTMiddleware creates a JWT middleware instance from runtime config
func NewJWTMiddleware(jwtConfig *jwt.RuntimeConfig) *JWTMiddleware {
	return &JWTMiddleware{
		jwtConfig: jwtConfig,
	}
}

// NewJWTMiddlewareFromConfig creates middleware instance from JWT Config (recommended)
// Follows config center's Registry pattern, similar to i18n.InitWithConfig
// This factory function handles config-to-runtime conversion and provides a clean API
func NewJWTMiddlewareFromConfig(cfg *jwt.Config) (*JWTMiddleware, error) {
	runtimeCfg, err := cfg.ToRuntimeConfig()
	if err != nil {
		return nil, err
	}
	return NewJWTMiddleware(runtimeCfg), nil
} // JWTAuth is the JWT authentication middleware handler
func (m *JWTMiddleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check whitelist (O(1) lookup)
		if m.jwtConfig.Whitelist[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.ErrorWithRequest(w, r, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required")
			return
		}

		// Verify Bearer format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.ErrorWithRequest(w, r, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwt.ValidateToken(m.jwtConfig, tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.ErrorWithRequest(w, r, http.StatusForbidden, "AUTH_TOKEN_EXPIRED", "Token has expired")
				return
			}
			response.ErrorWithRequest(w, r, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "Invalid token")
			return
		}

		// Inject context with user information
		ctx := r.Context()
		ctx = context.WithValue(ctx, jwt.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, jwt.UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, jwt.EmailKey, claims.Email)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
