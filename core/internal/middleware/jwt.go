// Package middleware provides HTTP middleware components for authentication and authorization.
// This package includes JWT authentication middleware that validates tokens and injects user context.
package middleware

import (
	"net/http"
	"strings"

	"apprun/ent"
	"apprun/internal/jwt"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/response"
)

// JWTMiddleware provides JWT authentication middleware
type JWTMiddleware struct {
	dbClient *ent.Client // Database client for token version validation (Story 5.7)
}

// NewJWTMiddleware creates a JWT middleware instance
func NewJWTMiddleware() *JWTMiddleware {
	return &JWTMiddleware{}
}

// NewJWTMiddlewareWithDB creates a JWT middleware instance with database client for token version validation
func NewJWTMiddlewareWithDB(dbClient *ent.Client) *JWTMiddleware {
	return &JWTMiddleware{dbClient: dbClient}
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

		// Verify token_version if database client is available (Story 5.7)
		if m.dbClient != nil {
			user, err := m.dbClient.User.Get(r.Context(), claims.UserID)
			if err != nil {
				logger.Warn("Token validation: user not found",
					logger.Field{Key: "user_id", Value: claims.UserID},
					logger.Field{Key: "error", Value: err.Error()})
				response.ErrorWithRequest(w, r, http.StatusUnauthorized, errors.ErrCodeAuthInvalidToken, "User not found")
				return
			}

			// Check if token version matches (Story 5.7: token revocation)
			if claims.TokenVersion != user.TokenVersion {
				logger.Warn("Token revoked: version mismatch",
					logger.Field{Key: "user_id", Value: claims.UserID},
					logger.Field{Key: "token_version", Value: claims.TokenVersion},
					logger.Field{Key: "current_version", Value: user.TokenVersion})
				response.ErrorWithRequest(w, r, http.StatusUnauthorized, "TOKEN_REVOKED", "Token has been revoked")
				return
			}

			// Check user status (Story 5.7: account status validation)
			if user.Status != 1 {
				logger.Warn("Token validation: user account disabled",
					logger.Field{Key: "user_id", Value: claims.UserID},
					logger.Field{Key: "status", Value: user.Status})
				response.ErrorWithRequest(w, r, http.StatusUnauthorized, "USER_DISABLED", "User account is disabled")
				return
			}

			// Check soft delete (Story 5.7)
			if user.DeletedAt != nil {
				logger.Warn("Token validation: user account deleted",
					logger.Field{Key: "user_id", Value: claims.UserID})
				response.ErrorWithRequest(w, r, http.StatusUnauthorized, "USER_DELETED", "User account has been deleted")
				return
			}
		}

		// Inject context with user information using jwt package keys for compatibility
		ctx := r.Context()
		ctx = jwt.SetUserID(ctx, claims.UserID)
		ctx = jwt.SetUsername(ctx, claims.Username)
		ctx = jwt.SetEmail(ctx, claims.Email)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
