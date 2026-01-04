package handlers

import (
	"net/http"

	"apprun/pkg/errors"
	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
)

// DemoHandler demonstrates usage of the unified response package with pkg/errors integration
type DemoHandler struct{}

// NewDemoHandler creates a new demo handler instance
func NewDemoHandler() *DemoHandler {
	return &DemoHandler{}
}

// RegisterRoutes registers demo routes
func (h *DemoHandler) RegisterRoutes(r chi.Router) {
	r.Route("/demo", func(r chi.Router) {
		r.Get("/success", h.Success)
		r.Post("/create", h.Create)
		r.Delete("/delete", h.Delete)
		r.Get("/list", h.List)
		r.Get("/error/not-found", h.ErrorNotFound)
		r.Get("/error/validation", h.ErrorValidation)
		r.Get("/error/auth", h.ErrorAuth)
		r.Get("/error/permission", h.ErrorPermission)
		r.Get("/error/business", h.ErrorBusiness)
		r.Get("/error/system", h.ErrorSystem)
	})
}

// Success demonstrates Success() response
func (h *DemoHandler) Success(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"message": "Operation completed successfully",
		"user_id": "12345",
	}
	response.Success(w, data)
}

// Create demonstrates Created() response with Location header
func (h *DemoHandler) Create(w http.ResponseWriter, r *http.Request) {
	newResource := map[string]interface{}{
		"id":   "new-resource-123",
		"name": "New Resource",
	}
	location := "/api/v1/demo/new-resource-123"
	response.Created(w, newResource, location)
}

// Delete demonstrates NoContent() response
func (h *DemoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	response.NoContent(w)
}

// List demonstrates List() with pagination
func (h *DemoHandler) List(w http.ResponseWriter, r *http.Request) {
	items := []map[string]string{
		{"id": "1", "name": "Item 1"},
		{"id": "2", "name": "Item 2"},
		{"id": "3", "name": "Item 3"},
	}

	pagination := &response.PaginationInfo{
		Total:      50,
		Page:       1,
		PageSize:   10,
		TotalPages: 5,
	}

	response.List(w, items, pagination)
}

// ErrorNotFound demonstrates RES (Resource) category error → 404
func (h *DemoHandler) ErrorNotFound(w http.ResponseWriter, r *http.Request) {
	err := errors.New(errors.ErrCodeNotFound, "User not found").
		WithContext(errors.ContextKeyUserID, "user123").
		WithContext(errors.ContextKeyRequestID, r.Header.Get("X-Request-ID"))

	response.AppErrorWithRequest(w, r, err)
}

// ErrorValidation demonstrates VAL (Validation) category error → 400
func (h *DemoHandler) ErrorValidation(w http.ResponseWriter, r *http.Request) {
	err := errors.New(errors.ErrCodeInvalidParam, "Email format is invalid").
		WithContext("field", "email").
		WithContext("value", "invalid-email")

	response.AppError(w, err)
}

// ErrorAuth demonstrates AUTH (Authentication) category error → 401
func (h *DemoHandler) ErrorAuth(w http.ResponseWriter, r *http.Request) {
	err := errors.New(errors.ErrCodeAuthTokenExpired, "Your session has expired").
		WithContext(errors.ContextKeyUserID, "user456")

	response.AppError(w, err)
}

// ErrorPermission demonstrates PERM (Permission) category error → 403
func (h *DemoHandler) ErrorPermission(w http.ResponseWriter, r *http.Request) {
	err := errors.New(errors.ErrCodeAuthNoPermission, "You don't have permission to access this resource").
		WithContext(errors.ContextKeyUserID, "user789").
		WithContext("resource_id", "project-123")

	response.AppError(w, err)
}

// ErrorBusiness demonstrates BIZ (Business) category error → 422
func (h *DemoHandler) ErrorBusiness(w http.ResponseWriter, r *http.Request) {
	err := errors.New(errors.ErrCodeConflict, "Cannot delete user with active subscriptions").
		WithContext(errors.ContextKeyUserID, "user999")

	response.AppError(w, err)
}

// ErrorSystem demonstrates SYS (System) category error → 500
func (h *DemoHandler) ErrorSystem(w http.ResponseWriter, r *http.Request) {
	err := errors.New(errors.ErrCodeInternalError, "Database connection failed").
		WithContext(errors.ContextKeyTraceID, "trace-xyz")

	response.AppError(w, err)
}
