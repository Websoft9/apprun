package handler

import (
	"encoding/json"
	stderrors "errors"
	"net/http"

	"apprun/internal/jwt"
	"apprun/internal/rbac"
	"apprun/modules/auth/service"
	"apprun/pkg/errors"
	"apprun/pkg/logger"
	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	projectService *service.ProjectService
	memberService  *service.ProjectMemberService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(projectService *service.ProjectService, memberService *service.ProjectMemberService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		memberService:  memberService,
	}
}

// CreateProject godoc
//
//	@Summary		Create a new project
//	@Description	Creates a new project with the authenticated user as owner
//	@Tags			Projects
//	@Accept			json
//	@Produce		json
//	@Param			request	body		service.CreateProjectRequest					true	"Project creation data"
//	@Success		201		{object}	response.Response{data=service.ProjectResponse}	"Project created successfully"
//	@Failure		400		{object}	response.Response								"Invalid request"
//	@Failure		401		{object}	response.Response								"Unauthorized"
//	@Failure		500		{object}	response.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/projects [post]
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from JWT context
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		logger.Warn("User not authenticated")
		appErr := errors.New(errors.ErrCodeAuthRequired, "User not authenticated")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Parse request
	var req service.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Invalid request body", logger.Field{Key: "error", Value: err.Error()})
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid request body")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Validate request
	if req.Name == "" {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Project name is required")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Create project
	project, err := h.projectService.CreateProject(ctx, req.Name, req.Description, userID)
	if err != nil {
		logger.Error("Failed to create project", logger.Field{Key: "error", Value: err.Error()})
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Return response
	response.CreatedWithRequest(w, r, service.ToProjectResponse(project), "")
}

// ListProjects godoc
//
//	@Summary		List user's projects
//	@Description	Lists all projects where the authenticated user is a member
//	@Tags			Projects
//	@Produce		json
//	@Success		200	{object}	response.Response{data=[]service.ProjectResponse}	"List of projects"
//	@Failure		401	{object}	response.Response									"Unauthorized"
//	@Failure		500	{object}	response.Response									"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/projects [get]
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from JWT context
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		logger.Warn("User not authenticated")
		appErr := errors.New(errors.ErrCodeAuthRequired, "User not authenticated")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// List projects
	projects, err := h.projectService.ListUserProjects(ctx, userID)
	if err != nil {
		logger.Error("Failed to list projects", logger.Field{Key: "error", Value: err.Error()})
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Convert to response format
	projectResponses := make([]service.ProjectResponse, len(projects))
	for i, p := range projects {
		projectResponses[i] = service.ToProjectResponse(p)
	}

	response.SuccessWithRequest(w, r, projectResponses)
}

// GetProject godoc
//
//	@Summary		Get project details
//	@Description	Gets detailed information about a specific project
//	@Tags			Projects
//	@Produce		json
//	@Param			id	path		string											true	"Project UUID"
//	@Success		200	{object}	response.Response{data=service.ProjectResponse}	"Project details"
//	@Failure		401	{object}	response.Response								"Unauthorized"
//	@Failure		404	{object}	response.Response								"Project not found"
//	@Failure		500	{object}	response.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/projects/{id} [get]
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectUUID := chi.URLParam(r, "id")
	if projectUUID == "" {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Project ID is required")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get current user ID
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		appErr := errors.New(errors.ErrCodeAuthRequired, "User not authenticated")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get project
	project, err := h.projectService.GetProjectByUUID(ctx, projectUUID)
	if err != nil {
		if stderrors.Is(err, service.ErrProjectNotFound) {
			response.AppErrorWithRequest(w, r, err)
			return
		}
		logger.Error("Failed to get project",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "uuid", Value: projectUUID})
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Check if user is a member of the project
	isMember, err := h.memberService.IsMember(ctx, project.ID, userID)
	if err != nil {
		logger.Error("Failed to check membership",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "project_id", Value: project.ID},
			logger.Field{Key: "user_id", Value: userID})
		appErr := errors.New(errors.ErrCodeInternalError, "Failed to check membership")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}
	if !isMember {
		appErr := errors.New(errors.ErrCodeAuthNoPermission, "You are not a member of this project")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	response.SuccessWithRequest(w, r, service.ToProjectResponse(project))
}

// UpdateProject godoc
//
//	@Summary		Update project
//	@Description	Updates project information (name, description)
//	@Tags			Projects
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string											true	"Project UUID"
//	@Param			request	body		service.CreateProjectRequest					true	"Project update data"
//	@Success		200		{object}	response.Response{data=service.ProjectResponse}	"Project updated successfully"
//	@Failure		400		{object}	response.Response								"Invalid request"
//	@Failure		401		{object}	response.Response								"Unauthorized"
//	@Failure		404		{object}	response.Response								"Project not found"
//	@Failure		500		{object}	response.Response								"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectUUID := chi.URLParam(r, "id")
	if projectUUID == "" {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Project ID is required")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get current user ID
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		appErr := errors.New(errors.ErrCodeAuthRequired, "User not authenticated")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get project first to get internal ID
	project, err := h.projectService.GetProjectByUUID(ctx, projectUUID)
	if err != nil {
		if stderrors.Is(err, service.ErrProjectNotFound) {
			response.AppErrorWithRequest(w, r, err)
			return
		}
		logger.Error("Failed to get project",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "uuid", Value: projectUUID})
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Check if user is owner or admin of the project
	role, err := h.memberService.GetMemberRole(ctx, project.ID, userID)
	if err != nil {
		logger.Error("Failed to get member role",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "project_id", Value: project.ID},
			logger.Field{Key: "user_id", Value: userID})
		appErr := errors.New(errors.ErrCodeAuthNotMember, "You are not a member of this project")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}
	if role != rbac.RoleProjectOwner && role != rbac.RoleProjectAdmin {
		appErr := errors.New(errors.ErrCodeAuthNoPermission, "Only owners and admins can update projects")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Parse request
	var req service.CreateProjectRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		logger.Warn("Invalid request body", logger.Field{Key: "error", Value: decodeErr.Error()})
		appErr := errors.New(errors.ErrCodeInvalidParam, "Invalid request body")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Update project
	var name, description *string
	if req.Name != "" {
		name = &req.Name
	}
	if req.Description != "" {
		description = &req.Description
	}

	updated, err := h.projectService.UpdateProject(ctx, project.ID, name, description)
	if err != nil {
		logger.Error("Failed to update project",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "project_id", Value: project.ID})
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, service.ToProjectResponse(updated))
}

// DeleteProject godoc
//
//	@Summary		Delete project
//	@Description	Archives a project (soft delete)
//	@Tags			Projects
//	@Produce		json
//	@Param			id	path		string				true	"Project UUID"
//	@Success		200	{object}	response.Response	"Project deleted successfully"
//	@Failure		401	{object}	response.Response	"Unauthorized"
//	@Failure		404	{object}	response.Response	"Project not found"
//	@Failure		500	{object}	response.Response	"Internal server error"
//	@Security		BearerAuth
//	@Router			/api/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectUUID := chi.URLParam(r, "id")
	if projectUUID == "" {
		appErr := errors.New(errors.ErrCodeInvalidParam, "Project ID is required")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get current user ID
	userID := jwt.GetUserID(ctx)
	if userID == 0 {
		appErr := errors.New(errors.ErrCodeAuthRequired, "User not authenticated")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Get project first to get internal ID
	project, err := h.projectService.GetProjectByUUID(ctx, projectUUID)
	if err != nil {
		if stderrors.Is(err, service.ErrProjectNotFound) {
			response.AppErrorWithRequest(w, r, err)
			return
		}
		logger.Error("Failed to get project",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "uuid", Value: projectUUID})
		response.AppErrorWithRequest(w, r, err)
		return
	}

	// Check if user is owner of the project (only owner can delete)
	role, err := h.memberService.GetMemberRole(ctx, project.ID, userID)
	if err != nil {
		logger.Error("Failed to get member role",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "project_id", Value: project.ID},
			logger.Field{Key: "user_id", Value: userID})
		appErr := errors.New(errors.ErrCodeAuthNotMember, "You are not a member of this project")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}
	if role != "owner" {
		appErr := errors.New(errors.ErrCodeAuthNoPermission, "Only project owner can delete the project")
		response.AppErrorWithRequest(w, r, appErr)
		return
	}

	// Delete project
	if err := h.projectService.DeleteProject(ctx, project.ID); err != nil {
		logger.Error("Failed to delete project",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "project_id", Value: project.ID})
		response.AppErrorWithRequest(w, r, err)
		return
	}

	response.SuccessWithRequest(w, r, map[string]string{"message": "Project deleted successfully"})
}
