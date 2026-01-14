package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"apprun/ent"
	"apprun/internal/jwt"
	"apprun/modules/auth/service"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/response"
)

// ProjectMemberHandler handles project member management HTTP endpoints
type ProjectMemberHandler struct {
	memberService *service.ProjectMemberService
}

// NewProjectMemberHandler creates a new project member handler
func NewProjectMemberHandler(memberService *service.ProjectMemberService) *ProjectMemberHandler {
	return &ProjectMemberHandler{
		memberService: memberService,
	}
}

// AddMemberRequest represents the request body for adding a member
type AddMemberRequest struct {
	UserID int64  `json:"user_id" validate:"required,gt=0"`
	Role   string `json:"role" validate:"required"`
}

// Validate validates the add member request
func (r *AddMemberRequest) Validate() error {
	if r.UserID <= 0 {
		return errors.New(errors.ErrCodeInvalidParam, "user_id must be positive")
	}

	if !IsValidProjectRole(r.Role) {
		return errors.New(errors.ErrCodeAuthInvalidRole, "Invalid role")
	}

	return nil
}

// UpdateRoleRequest represents the request body for updating a member's role
type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

// Validate validates the update role request
func (r *UpdateRoleRequest) Validate() error {
	if !IsValidProjectRole(r.Role) {
		return errors.New(errors.ErrCodeAuthInvalidRole, "Invalid role")
	}

	return nil
}

// MemberListResponse represents the response for listing members
type MemberListResponse struct {
	Items      []*ent.ProjectMember     `json:"items"`
	Pagination *response.PaginationInfo `json:"pagination,omitempty"`
}

// AddMember handles adding a member to a project
// @Summary      Add a member to a project
// @Description  Add a user to a project with a specific role (requires admin permission)
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Param        project_id   path      int                      true  "Project ID"
// @Param        request      body      AddMemberRequest         true  "Member info"
// @Success      201          {object}  response.Response{data=ent.ProjectMember}
// @Failure      400          {object}  response.Response
// @Failure      403          {object}  response.Response
// @Failure      404          {object}  response.Response
// @Failure      409          {object}  response.Response
// @Security     BearerAuth
// @Router       /api/projects/{project_id}/members [post]
func (h *ProjectMemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Extract project_id from URL
	projectID, err := strconv.ParseInt(chi.URLParam(r, "project_id"), 10, 64)
	if err != nil {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid project_id")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// 2. Parse request body
	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid request body")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// 3. Validate request
	if err := req.Validate(); err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// 4. Call service layer
	member, err := h.memberService.AddMember(ctx, projectID, req.UserID, req.Role)
	if err != nil {
		logger.Error("Failed to add member",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "project_id", Value: projectID},
			logger.Field{Key: "user_id", Value: req.UserID},
			logger.Field{Key: "role", Value: req.Role},
		)
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// 5. Audit log
	operatorID := jwt.GetUserID(ctx)
	logger.Info("Member added",
		logger.Field{Key: "operator_id", Value: operatorID},
		logger.Field{Key: "project_id", Value: projectID},
		logger.Field{Key: "new_member_user_id", Value: req.UserID},
		logger.Field{Key: "role", Value: req.Role},
	)

	// 6. Return success response
	response.CreatedWithRequest(w, r, member, "")
}

// ListMembers handles listing project members
// @Summary      List project members
// @Description  Get a list of all members in a project
// @Tags         rbac
// @Produce      json
// @Param        project_id   path      int     true   "Project ID"
// @Param        role         query     string  false  "Filter by role"
// @Param        page         query     int     false  "Page number" default(1)
// @Param        page_size    query     int     false  "Page size" default(20)
// @Success      200          {object}  response.Response{data=MemberListResponse}
// @Failure      403          {object}  response.Response
// @Security     BearerAuth
// @Router       /api/projects/{project_id}/members [get]
func (h *ProjectMemberHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract project_id
	projectID, err := strconv.ParseInt(chi.URLParam(r, "project_id"), 10, 64)
	if err != nil {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid project_id")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get members from service
	members, err := h.memberService.ListMembers(ctx, projectID)
	if err != nil {
		logger.Error("Failed to list members",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "project_id", Value: projectID},
		)
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// TODO: Add pagination support (currently returns all members)
	pagination := &response.PaginationInfo{
		Total:      len(members),
		Page:       1,
		PageSize:   len(members),
		TotalPages: 1,
	}

	resp := MemberListResponse{
		Items:      members,
		Pagination: pagination,
	}

	response.SuccessWithRequest(w, r, resp)
}

// UpdateRole handles updating a member's role
// @Summary      Update member role
// @Description  Update a project member's role (requires admin permission)
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Param        project_id   path      int                      true  "Project ID"
// @Param        member_id    path      int                      true  "Member ID"
// @Param        request      body      UpdateRoleRequest        true  "New role"
// @Success      200          {object}  response.Response{data=ent.ProjectMember}
// @Failure      400          {object}  response.Response
// @Failure      403          {object}  response.Response
// @Failure      404          {object}  response.Response
// @Security     BearerAuth
// @Router       /api/projects/{project_id}/members/{member_id} [put]
func (h *ProjectMemberHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract member_id
	memberID, err := strconv.ParseInt(chi.URLParam(r, "member_id"), 10, 64)
	if err != nil {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid member_id")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Parse request body
	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid request body")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Get existing member first to check for owner
	member, err := h.memberService.GetMemberByID(ctx, memberID)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Prevent modifying owner role
	if member.Role == "owner" {
		appErr := errors.New(
			errors.ErrCodeAuthCannotModifyOwner,
			"Cannot modify project owner role",
		)
		w.WriteHeader(http.StatusForbidden)
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Update role
	updatedMember, err := h.memberService.UpdateMemberRole(ctx, memberID, req.Role)
	if err != nil {
		logger.Error("Failed to update member role",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "member_id", Value: memberID},
			logger.Field{Key: "new_role", Value: req.Role},
		)
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Audit log
	operatorID := jwt.GetUserID(ctx)
	logger.Info("Member role updated",
		logger.Field{Key: "operator_id", Value: operatorID},
		logger.Field{Key: "member_id", Value: memberID},
		logger.Field{Key: "old_role", Value: member.Role},
		logger.Field{Key: "new_role", Value: req.Role},
	)

	response.SuccessWithRequest(w, r, updatedMember)
}

// RemoveMember handles removing a member from a project
// @Summary      Remove project member
// @Description  Remove a user from a project (requires admin permission)
// @Tags         rbac
// @Produce      json
// @Param        project_id   path      int  true  "Project ID"
// @Param        member_id    path      int  true  "Member ID"
// @Success      204          "No Content"
// @Failure      403          {object}  response.Response
// @Failure      404          {object}  response.Response
// @Security     BearerAuth
// @Router       /api/projects/{project_id}/members/{member_id} [delete]
func (h *ProjectMemberHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract member_id
	memberID, err := strconv.ParseInt(chi.URLParam(r, "member_id"), 10, 64)
	if err != nil {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid member_id")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get member first to check for owner
	member, err := h.memberService.GetMemberByID(ctx, memberID)
	if err != nil {
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Prevent removing owner
	if member.Role == "owner" {
		appErr := errors.New(
			errors.ErrCodeAuthCannotModifyOwner,
			"Cannot remove project owner",
		)
		w.WriteHeader(http.StatusForbidden)
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Remove member
	if err := h.memberService.RemoveMember(ctx, memberID); err != nil {
		logger.Error("Failed to remove member",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "member_id", Value: memberID},
		)
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Audit log
	operatorID := jwt.GetUserID(ctx)
	logger.Info("Member removed",
		logger.Field{Key: "operator_id", Value: operatorID},
		logger.Field{Key: "member_id", Value: memberID},
		logger.Field{Key: "removed_user_id", Value: member.UserID},
	)

	response.NoContent(w)
}

// GetMemberByID is a helper to get member by ID (used internally)
func (h *ProjectMemberHandler) GetMemberByID(ctx interface{}, memberID int64) (*ent.ProjectMember, error) {
	return h.memberService.GetMemberByID(ctx, memberID)
}
