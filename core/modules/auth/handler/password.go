// Package handler provides HTTP handlers for password management endpoints.
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

// PasswordHandler handles password management HTTP requests.
type PasswordHandler struct {
	userService *service.UserService
}

// NewPasswordHandler creates a new password handler.
func NewPasswordHandler(userService *service.UserService) *PasswordHandler {
	return &PasswordHandler{
		userService: userService,
	}
}

// ChangePassword handles PUT /api/users/me/password - change current user's password.
//
//	@Summary		Change user password
//	@Description	Change the authenticated user's password (requires old password verification)
//	@Tags			users
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		service.ChangePasswordRequest	true	"Password change data"
//	@Success		200		{object}	response.Response{message=string}
//	@Failure		400		{object}	response.Response	"Validation error or wrong old password"
//	@Failure		401		{object}	response.Response	"Unauthorized"
//	@Failure		404		{object}	response.Response	"User not found"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/api/users/me/password [put]
func (h *PasswordHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	logger.Info("Change password request received",
		logger.Field{Key: "method", Value: r.Method},
		logger.Field{Key: "path", Value: r.URL.Path})

	lang := i18n.GetLanguage(r.Context())

	// 1. Parse request body
	var req service.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Failed to parse request body", logger.Field{Key: "error", Value: err.Error()})
		msg := i18n.Translate(lang, "user.error.invalid_request_body", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}

	// 2. Validate required fields
	if req.OldPassword == "" {
		logger.Warn("Missing required field: old_password")
		msg := i18n.Translate(lang, "user.error.old_password_required", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}
	if req.NewPassword == "" {
		logger.Warn("Missing required field: new_password")
		msg := i18n.Translate(lang, "user.error.new_password_required", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}

	// 3. Call service layer
	err := h.userService.ChangePassword(r.Context(), &req)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			switch appErr.Code {
			case errors.ErrCodeUnauthorized:
				msg := i18n.Translate(lang, "auth.error.unauthorized", nil)
				response.Error(w, http.StatusUnauthorized, appErr.Code, msg)
			case errors.ErrCodeNotFound:
				msg := i18n.Translate(lang, "user.error.not_found", nil)
				response.Error(w, http.StatusNotFound, appErr.Code, msg)
			case errors.ErrCodeAuthInvalidCredentials:
				// Old password incorrect
				msg := i18n.Translate(lang, "user.error.old_password_incorrect", nil)
				response.Error(w, http.StatusBadRequest, appErr.Code, msg)
			case errors.ErrCodeAuthWeakPassword:
				// New password doesn't meet requirements
				msg := i18n.Translate(lang, "auth.error.weak_password", nil)
				response.Error(w, http.StatusBadRequest, appErr.Code, msg)
			case errors.ErrCodeInvalidParam:
				// Same password or other validation error
				msg := i18n.Translate(lang, "user.error.validation_failed", nil)
				if appErr.Message != "" {
					msg = appErr.Message
				}
				response.Error(w, http.StatusBadRequest, appErr.Code, msg)
			default:
				logger.Error("Failed to change password", logger.Field{Key: "error", Value: err.Error()})
				msg := i18n.Translate(lang, "user.error.change_password_failed", nil)
				response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
			}
		} else {
			logger.Error("Unexpected error type", logger.Field{Key: "error", Value: err.Error()})
			msg := i18n.Translate(lang, "user.error.change_password_failed", nil)
			response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
		}
		return
	}

	logger.Info("Password changed successfully")
	msg := i18n.Translate(lang, "user.success.password_changed", nil)
	response.Success(w, map[string]string{"message": msg})
}
