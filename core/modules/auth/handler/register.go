// Package handler provides HTTP handlers for authentication endpoints.
package handler

import (
	"encoding/json"
	stdErrors "errors"
	"net/http"

	"apprun/modules/auth/service"
	"apprun/pkg/errors"
	"apprun/pkg/i18n"
	"apprun/pkg/logger"
	"apprun/pkg/response"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration requests.
//
// @Summary      Register new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.RegisterRequest true "Registration data"
// @Success      201 {object} response.Response{data=service.RegisterResponse}
// @Failure      400 {object} response.Response "Validation error"
// @Failure      409 {object} response.Response "Email or username already exists"
// @Failure      500 {object} response.Response "Internal server error"
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	logger.Info("Registration request received",
		logger.Field{Key: "method", Value: r.Method},
		logger.Field{Key: "path", Value: r.URL.Path})

	// Get user's preferred language from context (set by middleware)
	lang := i18n.GetLanguage(r.Context())

	// 1. Parse request body
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Failed to parse request body", logger.Field{Key: "error", Value: err.Error()})
		msg := i18n.Translate(lang, "auth.error.invalid_request_body", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}

	// 2. Validate required fields
	if req.Email == "" {
		logger.Warn("Missing required field: email")
		msg := i18n.Translate(lang, "auth.error.email_required", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}
	if req.Password == "" {
		logger.Warn("Missing required field: password")
		msg := i18n.Translate(lang, "auth.error.password_required", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}

	// 3. Call service layer
	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		// Map service errors to HTTP responses using pkg/errors
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			switch appErr.Code {
			case errors.ErrCodeAuthInvalidEmail:
				msg := i18n.Translate(lang, "auth.error.invalid_email", nil)
				response.Error(w, http.StatusBadRequest, appErr.Code, msg)
			case errors.ErrCodeAuthEmailExists:
				msg := i18n.Translate(lang, "auth.error.email_exists", nil)
				response.Error(w, http.StatusConflict, appErr.Code, msg)
			case errors.ErrCodeAuthUsernameExists:
				msg := i18n.Translate(lang, "auth.error.username_exists", nil)
				response.Error(w, http.StatusConflict, appErr.Code, msg)
			case errors.ErrCodeAuthWeakPassword:
				msg := i18n.Translate(lang, "auth.error.weak_password", nil)
				response.Error(w, http.StatusBadRequest, appErr.Code, msg)
			default:
				logger.Error("Registration failed", logger.Field{Key: "error", Value: err.Error()})
				msg := i18n.Translate(lang, "auth.error.registration_failed", nil)
				response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
			}
		} else {
			// Non-AppError (shouldn't happen, but handle gracefully)
			logger.Error("Unexpected error type", logger.Field{Key: "error", Value: err.Error()})
			msg := i18n.Translate(lang, "auth.error.registration_failed", nil)
			response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
		}
		return
	}

	logger.Info("Registration successful", logger.Field{Key: "email", Value: req.Email})
	// 4. Return success response
	successMsg := i18n.Translate(lang, "auth.success.registration", nil)
	response.CreatedWithRequest(w, r, user, successMsg)
}
