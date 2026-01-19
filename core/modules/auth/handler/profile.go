// Package handler provides HTTP handlers for user profile endpoints.
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

// ProfileHandler handles user profile HTTP requests.
type ProfileHandler struct {
	userService *service.UserService
}

// NewProfileHandler creates a new profile handler.
func NewProfileHandler(userService *service.UserService) *ProfileHandler {
	return &ProfileHandler{
		userService: userService,
	}
}

// GetProfile handles GET /api/users/me - retrieve current user's profile.
//
//	@Summary		Get current user profile
//	@Description	Retrieve the complete profile of the authenticated user
//	@Tags			users
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	response.Response{data=service.ProfileResponse}
//	@Failure		401	{object}	response.Response	"Unauthorized - missing or invalid token"
//	@Failure		404	{object}	response.Response	"User not found"
//	@Failure		500	{object}	response.Response	"Internal server error"
//	@Router			/api/users/me [get]
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	logger.Info("Get profile request received",
		logger.Field{Key: "method", Value: r.Method},
		logger.Field{Key: "path", Value: r.URL.Path})

	lang := i18n.GetLanguage(r.Context())

	// Call service layer
	profile, err := h.userService.GetCurrentUser(r.Context())
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
			default:
				logger.Error("Failed to get profile", logger.Field{Key: "error", Value: err.Error()})
				msg := i18n.Translate(lang, "user.error.get_profile_failed", nil)
				response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
			}
		} else {
			logger.Error("Unexpected error type", logger.Field{Key: "error", Value: err.Error()})
			msg := i18n.Translate(lang, "user.error.get_profile_failed", nil)
			response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
		}
		return
	}

	logger.Info("Profile retrieved successfully", logger.Field{Key: "user_id", Value: profile.ID})
	response.Success(w, profile)
}

// UpdateProfile handles PUT /api/users/me - update current user's profile.
//
//	@Summary		Update current user profile
//	@Description	Update the authenticated user's profile (name, avatar, bio)
//	@Tags			users
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		service.UpdateProfileRequest	true	"Profile update data"
//	@Success		200		{object}	response.Response{data=service.ProfileResponse}
//	@Failure		400		{object}	response.Response	"Validation error"
//	@Failure		401		{object}	response.Response	"Unauthorized"
//	@Failure		404		{object}	response.Response	"User not found"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/api/users/me [put]
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	logger.Info("Update profile request received",
		logger.Field{Key: "method", Value: r.Method},
		logger.Field{Key: "path", Value: r.URL.Path})

	lang := i18n.GetLanguage(r.Context())

	// 1. Parse request body
	var req service.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Failed to parse request body", logger.Field{Key: "error", Value: err.Error()})
		msg := i18n.Translate(lang, "user.error.invalid_request_body", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}

	// 2. Call service layer
	profile, err := h.userService.UpdateProfile(r.Context(), &req)
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
			case errors.ErrCodeInvalidParam:
				// Handle specific validation errors
				msg := i18n.Translate(lang, "user.error.validation_failed", nil)
				if appErr.Message != "" {
					msg = appErr.Message // Use specific error message
				}
				response.Error(w, http.StatusBadRequest, appErr.Code, msg)
			default:
				logger.Error("Failed to update profile", logger.Field{Key: "error", Value: err.Error()})
				msg := i18n.Translate(lang, "user.error.update_profile_failed", nil)
				response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
			}
		} else {
			logger.Error("Unexpected error type", logger.Field{Key: "error", Value: err.Error()})
			msg := i18n.Translate(lang, "user.error.update_profile_failed", nil)
			response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
		}
		return
	}

	logger.Info("Profile updated successfully", logger.Field{Key: "user_id", Value: profile.ID})
	response.Success(w, profile)
}
