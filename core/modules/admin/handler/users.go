// Package handler provides HTTP handlers for user management endpoints.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"apprun/internal/jwt"
	admin "apprun/modules/admin"
	adminService "apprun/modules/admin/service"
	pkgErrors "apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
)

// UsersHandler handles user management HTTP requests (Story 5.7)
type UsersHandler struct {
	userMgmtSvc *adminService.UserMgmtService
}

// NewUsersHandler creates a new users handler
func NewUsersHandler(userMgmtSvc *adminService.UserMgmtService) *UsersHandler {
	return &UsersHandler{userMgmtSvc: userMgmtSvc}
}

// ListUsers godoc
// @Summary List all users (admin only)
// @Description Get paginated list of all platform users with optional filtering by role, status, and search term
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Param search query string false "Search term (email, nickname)"
// @Param role query string false "Filter by role (platform_admin, platform_user)"
// @Param status query int false "Filter by status (0=disabled, 1=active)"
// @Success 200 {object} response.Response{data=adminService.ListUsersResponse}
// @Failure 401 {object} response.Response "Unauthorized - JWT token required"
// @Failure 403 {object} response.Response "Forbidden - platform_admin role required"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/admin/users [get]
func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		pageSize = 20
	}
	search := r.URL.Query().Get("search")
	role := r.URL.Query().Get("role")

	var status *int8
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		statusVal, parseErr := strconv.ParseInt(statusStr, 10, 8)
		if parseErr == nil {
			statusInt8 := int8(statusVal)
			status = &statusInt8
		}
	}

	req := &adminService.ListUsersRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		Role:     role,
		Status:   status,
	}

	result, err := h.userMgmtSvc.ListUsers(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.Success(w, result)
}

// GetUser godoc
// @Summary Get user by ID (admin only)
// @Description Get detailed information about a specific user
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response "Unauthorized - JWT token required"
// @Failure 403 {object} response.Response "Forbidden - platform_admin role required"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/admin/users/{id} [get]
func (h *UsersHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	user, err := h.userMgmtSvc.GetUserByID(r.Context(), userID)
	if err != nil {
		if pkgErrors.IsNotFound(err) {
			response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		} else {
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.Success(w, admin.ToUserResponse(user, true, true))
}

// CreateUser godoc
// @Summary Create new user (admin only)
// @Description Create a new platform user or administrator
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body adminService.CreateUserRequest true "User creation data"
// @Success 201 {object} response.Response{data=object}
// @Failure 400 {object} response.Response "Bad request - invalid input"
// @Failure 401 {object} response.Response "Unauthorized - JWT token required"
// @Failure 403 {object} response.Response "Forbidden - platform_admin role required"
// @Failure 409 {object} response.Response "Conflict - email already exists"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/admin/users [post]
func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req adminService.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, generatedPassword, err := h.userMgmtSvc.CreateUser(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, pkgErrors.ErrAdminEmailExists):
			response.Error(w, http.StatusConflict, pkgErrors.ErrCodeAdminEmailExists, err.Error())
		case pkgErrors.IsValidation(err):
			response.Error(w, http.StatusBadRequest, pkgErrors.ErrCodeInvalidParam, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	operatorID := jwt.GetUserID(r.Context())
	logger.Info("User created",
		logger.Field{Key: "new_user_id", Value: user.ID},
		logger.Field{Key: "operator_id", Value: operatorID})

	// Build response with generated password if applicable
	resp := admin.CreateUserResponse{
		User: admin.ToUserResponse(user, false, true),
	}
	if generatedPassword != "" {
		resp.GeneratedPassword = &generatedPassword
	}

	w.WriteHeader(http.StatusCreated)
	response.Success(w, resp)
}

// ChangeUserRole godoc
// @Summary Change user role (admin only)
// @Description Change a user's platform role (upgrade/downgrade between platform_admin and platform_user)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body adminService.ChangeUserRoleRequest true "New role"
// @Success 200 {object} response.Response{data=object}
// @Failure 400 {object} response.Response "Bad request - invalid role"
// @Failure 401 {object} response.Response "Unauthorized - JWT token required"
// @Failure 403 {object} response.Response "Forbidden - cannot demote last admin or modify system user"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/admin/users/{id}/role [put]
//
//nolint:dupl // Similar structure but different business logic (role vs status)
func (h *UsersHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	var req adminService.ChangeUserRoleRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	operatorID := jwt.GetUserID(r.Context())
	user, err := h.userMgmtSvc.ChangeUserRole(r.Context(), userID, operatorID, req.Role)
	if err != nil {
		switch {
		case pkgErrors.IsNotFound(err):
			response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		case pkgErrors.IsValidation(err):
			response.Error(w, http.StatusBadRequest, pkgErrors.ErrCodeInvalidParam, err.Error())
		case errors.Is(err, pkgErrors.ErrAdminCannotDemoteLastAdmin):
			response.Error(w, http.StatusForbidden, pkgErrors.ErrCodeAdminCannotDemoteLastAdmin, err.Error())
		case errors.Is(err, pkgErrors.ErrAdminCannotModifySystem):
			response.Error(w, http.StatusForbidden, pkgErrors.ErrCodeAdminCannotModifySystem, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.Success(w, admin.ToUserResponse(user, false, false))
}

// ChangeUserStatus godoc
// @Summary Change user status (admin only)
// @Description Enable or disable a user account
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body adminService.ChangeUserStatusRequest true "New status (0=disabled, 1=active)"
// @Success 200 {object} response.Response{data=object}
// @Failure 400 {object} response.Response "Bad request - invalid status"
// @Failure 401 {object} response.Response "Unauthorized - JWT token required"
// @Failure 403 {object} response.Response "Forbidden - cannot disable self or modify system user"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/admin/users/{id}/status [put]
//
//nolint:dupl // Similar structure but different business logic (status vs role)
func (h *UsersHandler) ChangeUserStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	var req adminService.ChangeUserStatusRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	operatorID := jwt.GetUserID(r.Context())
	user, err := h.userMgmtSvc.ChangeUserStatus(r.Context(), userID, operatorID, req.Status)
	if err != nil {
		switch {
		case pkgErrors.IsNotFound(err):
			response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		case pkgErrors.IsValidation(err):
			response.Error(w, http.StatusBadRequest, pkgErrors.ErrCodeInvalidParam, err.Error())
		case errors.Is(err, pkgErrors.ErrAdminCannotDisableSelf):
			response.Error(w, http.StatusForbidden, pkgErrors.ErrCodeAdminCannotDisableSelf, err.Error())
		case errors.Is(err, pkgErrors.ErrAdminCannotModifySystem):
			response.Error(w, http.StatusForbidden, pkgErrors.ErrCodeAdminCannotModifySystem, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.Success(w, admin.ToUserResponse(user, false, false))
}

// DeleteUser godoc
// @Summary Delete user (admin only)
// @Description Soft delete a user account (preserves data relations)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{message=string}
// @Failure 400 {object} response.Response "Bad request - invalid user ID"
// @Failure 401 {object} response.Response "Unauthorized - JWT token required"
// @Failure 403 {object} response.Response "Forbidden - cannot delete self, last admin, or system user"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/admin/users/{id} [delete]
func (h *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	operatorID := jwt.GetUserID(r.Context())
	err = h.userMgmtSvc.DeleteUser(r.Context(), userID, operatorID)
	if err != nil {
		switch {
		case pkgErrors.IsNotFound(err):
			response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		case errors.Is(err, pkgErrors.ErrAdminCannotDeleteSelf):
			response.Error(w, http.StatusForbidden, pkgErrors.ErrCodeAdminCannotDeleteSelf, err.Error())
		case errors.Is(err, pkgErrors.ErrAdminCannotDeleteLastAdmin):
			response.Error(w, http.StatusForbidden, pkgErrors.ErrCodeAdminCannotDeleteLastAdmin, err.Error())
		case errors.Is(err, pkgErrors.ErrAdminCannotModifySystem):
			response.Error(w, http.StatusForbidden, pkgErrors.ErrCodeAdminCannotModifySystem, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	logger.Info("User deleted",
		logger.Field{Key: "deleted_user_id", Value: userID},
		logger.Field{Key: "operator_id", Value: operatorID})

	response.Success(w, map[string]string{"message": "User deleted successfully"})
}
