package handlers

import (
	"net/http"

	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
)

// DemoHandler demonstrates usage of the unified response package
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
		r.Get("/error", h.Error)
		r.Post("/validate", h.Validate)
	})
}

// Success demonstrates Success() response
func (h *DemoHandler) Success(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"message": "Operation successful",
		"status":  "ok",
	}
	response.Success(w, data)
}

// Create demonstrates Created() response with Location header
func (h *DemoHandler) Create(w http.ResponseWriter, r *http.Request) {
	newResource := map[string]string{
		"id":   "123",
		"name": "New Resource",
	}
	location := "/api/v1/demo/123"
	response.Created(w, newResource, location)
}

// Delete demonstrates NoContent() response
func (h *DemoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Simulate deletion logic
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

// Error demonstrates Error() response with standard error codes
func (h *DemoHandler) Error(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotFound, response.ErrCodeNotFound, "Resource not found")
}

// Validate demonstrates ValidationError() response
func (h *DemoHandler) Validate(w http.ResponseWriter, r *http.Request) {
	response.ValidationError(w, "email", "Email format is invalid")
}
