package middleware

import (
	"net/http"
	"strconv"

	"apprun/ent"
	"apprun/internal/jwt"
	"apprun/internal/rbac"
	"apprun/modules/auth/repository"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
)

// ProjectContextMiddleware extracts project_id from URL and validates membership
// Must be used after AuthMiddleware (requires user_id in context)
// Note: Repository dependency should be injected via factory function in production
func ProjectContextMiddleware(client *ent.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user_id from context (set by AuthMiddleware)
			userID := jwt.GetUserID(r.Context())
			if userID == 0 {
				logger.Warn("ProjectContextMiddleware: missing user_id",
					logger.Field{Key: "path", Value: r.URL.Path},
					logger.Field{Key: "method", Value: r.Method},
				)
				err := errors.New(errors.ErrCodeUnauthorized, "Authentication required")
				response.Error(w, 401, err.Code, err.Message)
				return
			}

			// Extract project_id from URL path parameter
			projectIDStr := chi.URLParam(r, "project_id")
			if projectIDStr == "" {
				// No project_id in path - this is a platform-level request
				// Continue without project context
				next.ServeHTTP(w, r)
				return
			}

			// Parse project_id
			projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
			if err != nil {
				logger.Warn("ProjectContextMiddleware: invalid project_id",
					logger.Field{Key: "project_id", Value: projectIDStr},
					logger.Field{Key: "user_id", Value: userID},
					logger.Field{Key: "error", Value: err},
				)
				appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid project_id")
				response.Error(w, 400, appErr.Code, appErr.Message)
				return
			}

			// Verify user is a member of this project
			if client != nil {
				memberRepo := repository.NewProjectMemberRepository(client)
				isMember, err := memberRepo.IsMember(r.Context(), projectID, userID)
				if err != nil {
					logger.Error("ProjectContextMiddleware: failed to check membership",
						logger.Field{Key: "error", Value: err},
						logger.Field{Key: "user_id", Value: userID},
						logger.Field{Key: "project_id", Value: projectID},
					)
					appErr := errors.New(errors.ErrCodeAuthPermCheckError, "Membership check failed")
					response.Error(w, 500, appErr.Code, appErr.Message)
					return
				}

				if !isMember {
					logger.Warn("ProjectContextMiddleware: user not a project member",
						logger.Field{Key: "user_id", Value: userID},
						logger.Field{Key: "project_id", Value: projectID},
						logger.Field{Key: "path", Value: r.URL.Path},
					)
					appErr := errors.New(errors.ErrCodeAuthNotMember, "Not a project member")
					response.Error(w, 403, appErr.Code, appErr.Message)
					return
				}
			}

			// Inject project_id into context
			ctx := jwt.SetProjectID(r.Context(), projectID)

			// Continue with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission returns a middleware that checks if the user has permission
// to perform the specified action on the specified resource
func RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user_id from context
			userID := jwt.GetUserID(r.Context())
			if userID == 0 {
				logger.Warn("RequirePermission: missing user_id",
					logger.Field{Key: "resource", Value: resource},
					logger.Field{Key: "action", Value: action},
					logger.Field{Key: "path", Value: r.URL.Path},
				)
				err := errors.New(errors.ErrCodeUnauthorized, "Authentication required")
				response.Error(w, 401, err.Code, err.Message)
				return
			}

			// Get project_id from context (0 means platform-level)
			projectID := jwt.GetProjectID(r.Context())

			// Check permission using RBAC enforcer
			enforcer := rbac.GetEnforcer()
			allowed, err := enforcer.Enforce(rbac.FormatUserKey(userID), rbac.FormatDomain(projectID), resource, action)
			if err != nil {
				logger.Error("RequirePermission: permission check failed",
					logger.Field{Key: "user_id", Value: userID},
					logger.Field{Key: "project_id", Value: projectID},
					logger.Field{Key: "resource", Value: resource},
					logger.Field{Key: "action", Value: action},
					logger.Field{Key: "error", Value: err},
				)
				appErr := errors.New(errors.ErrCodeAuthPermCheckError, "Permission check failed")
				response.Error(w, 500, appErr.Code, appErr.Message)
				return
			}

			if !allowed {
				logger.Warn("RequirePermission: permission denied",
					logger.Field{Key: "user_id", Value: userID},
					logger.Field{Key: "project_id", Value: projectID},
					logger.Field{Key: "resource", Value: resource},
					logger.Field{Key: "action", Value: action},
					logger.Field{Key: "path", Value: r.URL.Path},
					logger.Field{Key: "method", Value: r.Method},
					logger.Field{Key: "ip", Value: r.RemoteAddr},
					logger.Field{Key: "user_agent", Value: r.UserAgent()},
				)
				appErr := errors.New(errors.ErrCodeAuthNoPermission, "Permission denied")
				response.Error(w, 403, appErr.Code, appErr.Message)
				return
			}

			// Permission granted, continue
			next.ServeHTTP(w, r)
		})
	}
}
