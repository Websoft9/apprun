// Package middleware provides HTTP middleware components for authentication and authorization.
package middleware

import (
	"net/http"

	"apprun/ent"
	"apprun/internal/jwt"
	admin "apprun/modules/admin"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/response"
)

// RequirePlatformAdminMiddleware provides platform administrator authorization middleware
type RequirePlatformAdminMiddleware struct {
	dbClient *ent.Client
}

// NewRequirePlatformAdminMiddleware creates a new platform admin middleware instance
func NewRequirePlatformAdminMiddleware(dbClient *ent.Client) *RequirePlatformAdminMiddleware {
	return &RequirePlatformAdminMiddleware{dbClient: dbClient}
}

// RequirePlatformAdmin middleware ensures the user has platform_admin role
func (m *RequirePlatformAdminMiddleware) RequirePlatformAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user_id from context (set by JWT middleware)
		userID := jwt.GetUserID(r.Context())
		if userID == 0 {
			logger.Warn("RequirePlatformAdmin: missing user_id in context",
				logger.Field{Key: "path", Value: r.URL.Path})
			response.ErrorWithRequest(w, r, http.StatusUnauthorized, errors.ErrCodeUnauthorized, "Authentication required")
			return
		}

		// Query user to check role
		user, err := m.dbClient.User.Get(r.Context(), userID)
		if err != nil {
			logger.Error("RequirePlatformAdmin: failed to query user",
				logger.Field{Key: "user_id", Value: userID},
				logger.Field{Key: "error", Value: err.Error()})
			response.ErrorWithRequest(w, r, http.StatusInternalServerError, errors.ErrCodeInternalError, "Failed to verify permissions")
			return
		}

		// Check if user has platform_admin role
		if user.Role != admin.RolePlatformAdmin {
			logger.Warn("RequirePlatformAdmin: access denied",
				logger.Field{Key: "user_id", Value: userID},
				logger.Field{Key: "role", Value: user.Role},
				logger.Field{Key: "path", Value: r.URL.Path})
			response.ErrorWithRequest(w, r, http.StatusForbidden, "PERM_ADMIN_REQUIRED", "Platform administrator permission required")
			return
		}

		// User is platform admin, proceed
		next.ServeHTTP(w, r)
	})
}
