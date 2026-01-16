package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"apprun/internal/jwt"
	"apprun/modules/auth/service"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/response"
)

// PermissionHandler handles permission-related HTTP endpoints
type PermissionHandler struct {
	permissionService *service.PermissionService
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(permissionService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// CheckPermissionRequest represents the request body for permission check
type CheckPermissionRequest struct {
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
}

// Validate validates the check permission request
func (r *CheckPermissionRequest) Validate() error {
	if r.Resource == "" {
		return errors.New(errors.ErrCodeInvalidParam, "resource is required")
	}
	if r.Action == "" {
		return errors.New(errors.ErrCodeInvalidParam, "action is required")
	}
	return nil
}

// PermissionsResponse represents the permission list response
type PermissionsResponse struct {
	Permissions []PermissionItem `json:"permissions"`
}

// PermissionItem represents a single permission
type PermissionItem struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

// CheckPermissionResponse represents the permission check result
type CheckPermissionResponse struct {
	Allowed bool `json:"allowed"`
}

// GetMyPermissions handles getting current user's permissions in a project
//
//	@Summary		Get my permissions
//	@Description	Get the current user's permissions in a project
//	@Tags			rbac
//	@Produce		json
//	@Param			project_id	path		int	true	"Project ID"
//	@Success		200			{object}	response.Response{data=PermissionsResponse}
//	@Failure		403			{object}	response.Response
//	@Security		BearerAuth
//	@Router			/api/projects/{project_id}/permissions/me [get]
func (h *PermissionHandler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract project_id
	projectID, err := strconv.ParseInt(chi.URLParam(r, "project_id"), 10, 64)
	if err != nil {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid project_id")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get user_id from JWT context
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		appErr := errors.New(errors.ErrCodeAuthMissingToken, "User not authenticated")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get user's role in project
	role, err := h.permissionService.GetUserRoleInProject(ctx, userID, projectID)
	if err != nil {
		logger.Error("Failed to get user role",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "project_id", Value: projectID},
		)
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Get permissions for role
	permissions, err := h.permissionService.GetPermissionsForRole(ctx, projectID, role)
	if err != nil {
		logger.Error("Failed to get permissions",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "role", Value: role},
			logger.Field{Key: "project_id", Value: projectID},
		)
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Convert to response format
	resp := PermissionsResponse{
		Permissions: buildPermissionItems(permissions),
	}

	response.SuccessWithRequest(w, r, resp)
}

// CheckPermission handles checking if user has a specific permission
//
//	@Summary		Check permission
//	@Description	Check if the current user has a specific permission in a project
//	@Tags			rbac
//	@Accept			json
//	@Produce		json
//	@Param			project_id	path		int						true	"Project ID"
//	@Param			request		body		CheckPermissionRequest	true	"Permission to check"
//	@Success		200			{object}	response.Response{data=CheckPermissionResponse}
//	@Failure		400			{object}	response.Response
//	@Failure		403			{object}	response.Response
//	@Security		BearerAuth
//	@Router			/api/projects/{project_id}/permissions/check [post]
func (h *PermissionHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract project_id
	projectID, err := strconv.ParseInt(chi.URLParam(r, "project_id"), 10, 64)
	if err != nil {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid project_id")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Parse request body
	var req CheckPermissionRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid request body")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Validate request
	if validateErr := req.Validate(); validateErr != nil {
		response.AppErrorWithRequest(w, r, validateErr)
		return
	}

	// Get user_id from JWT context
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		appErr := errors.New(errors.ErrCodeAuthMissingToken, "User not authenticated")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Check permission
	allowed, err := h.permissionService.CheckPermission(ctx, userID, projectID, req.Resource, req.Action)
	if err != nil {
		logger.Error("Failed to check permission",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "project_id", Value: projectID},
			logger.Field{Key: "resource", Value: req.Resource},
			logger.Field{Key: "action", Value: req.Action},
		)
		response.AppErrorWithRequest(w, r, err)
		return
	}

	resp := CheckPermissionResponse{
		Allowed: allowed,
	}

	response.SuccessWithRequest(w, r, resp)
}

// buildPermissionItems converts a map of permissions to a structured list
// Input format: map[resource][]action
func buildPermissionItems(permissions map[string][]string) []PermissionItem {
	items := make([]PermissionItem, 0, len(permissions))
	for resource, actions := range permissions {
		items = append(items, PermissionItem{
			Resource: resource,
			Actions:  actions,
		})
	}
	return items
}
