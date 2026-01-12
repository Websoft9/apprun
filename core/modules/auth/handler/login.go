// Package handler provides HTTP handlers for authentication endpoints.
package handler

import (
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"strings"

	"apprun/modules/auth/service"
	"apprun/pkg/errors"
	"apprun/pkg/i18n"
	"apprun/pkg/logger"
	"apprun/pkg/response"
)

// Login handles user login requests.
//
// @Summary      User login
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body service.LoginRequest true "Login credentials"
// @Success      200 {object} response.Response{data=service.LoginResponse}
// @Failure      400 {object} response.Response "Validation error"
// @Failure      401 {object} response.Response "Invalid credentials"
// @Failure      403 {object} response.Response "Account disabled"
// @Failure      500 {object} response.Response "Internal server error"
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger.Info("Login request received",
		logger.Field{Key: "method", Value: r.Method},
		logger.Field{Key: "path", Value: r.URL.Path},
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr})

	// Get user's preferred language from context
	lang := i18n.GetLanguage(r.Context())

	// 1. Parse request body
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Failed to parse request body", logger.Field{Key: "error", Value: err.Error()})
		msg := i18n.Translate(lang, "auth.error.invalid_request_body", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}

	// 2. Validate required fields
	if req.Identifier == "" {
		logger.Warn("Missing required field: identifier")
		msg := i18n.Translate(lang, "auth.error.identifier_required", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}
	if req.Password == "" {
		logger.Warn("Missing required field: password")
		msg := i18n.Translate(lang, "auth.error.password_required", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}

	// 3. Extract client IP
	clientIP := getClientIP(r)

	// 4. Call service layer
	loginResp, err := h.authService.Login(r.Context(), &req, clientIP)
	if err != nil {
		// Map service errors to HTTP responses
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			switch appErr.Code {
			case errors.ErrCodeAuthInvalidCredentials:
				// Generic error - don't reveal user existence
				logger.Warn("Invalid credentials",
					logger.Field{Key: "identifier", Value: req.Identifier},
					logger.Field{Key: "client_ip", Value: clientIP})
				msg := i18n.Translate(lang, "auth.error.invalid_credentials", nil)
				response.Error(w, http.StatusUnauthorized, appErr.Code, msg)
			case errors.ErrCodeAuthAccountDisabled:
				logger.Warn("Disabled account login attempt",
					logger.Field{Key: "identifier", Value: req.Identifier},
					logger.Field{Key: "client_ip", Value: clientIP})
				msg := i18n.Translate(lang, "auth.error.account_disabled", nil)
				response.Error(w, http.StatusForbidden, appErr.Code, msg)
			default:
				logger.Error("Login failed",
					logger.Field{Key: "identifier", Value: req.Identifier},
					logger.Field{Key: "error", Value: err.Error()})
				msg := i18n.Translate(lang, "auth.error.login_failed", nil)
				response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
			}
		} else {
			// Non-AppError (shouldn't happen, but handle gracefully)
			logger.Error("Unexpected error type",
				logger.Field{Key: "identifier", Value: req.Identifier},
				logger.Field{Key: "error", Value: err.Error()})
			msg := i18n.Translate(lang, "auth.error.login_failed", nil)
			response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
		}
		return
	}

	logger.Info("Login successful",
		logger.Field{Key: "user_id", Value: loginResp.User.UUID},
		logger.Field{Key: "email", Value: loginResp.User.Email})

	// 5. Return success response
	response.Success(w, loginResp)
}

// Me returns the authenticated user's profile.
//
// @Summary      Get current user profile
// @Description  Retrieve profile of the authenticated user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response{data=service.UserProfile}
// @Failure      401 {object} response.Response "Missing or invalid token"
// @Failure      404 {object} response.Response "User not found"
// @Failure      500 {object} response.Response "Internal server error"
// @Router       /api/auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	logger.Info("Profile request received",
		logger.Field{Key: "method", Value: r.Method},
		logger.Field{Key: "path", Value: r.URL.Path})

	// Get user's preferred language
	lang := i18n.GetLanguage(r.Context())

	// 1. Extract user ID from JWT context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		logger.Error("Failed to extract user_id from context")
		msg := i18n.Translate(lang, "auth.error.invalid_token", nil)
		response.Error(w, http.StatusUnauthorized, errors.ErrCodeAuthInvalidToken, msg)
		return
	}

	// 2. Get user profile
	profile, err := h.authService.GetUserProfile(r.Context(), userID)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			if appErr.Code == errors.ErrCodeAuthUserNotFound {
				logger.Warn("User not found", logger.Field{Key: "user_id", Value: userID})
				msg := i18n.Translate(lang, "auth.error.user_not_found", nil)
				response.Error(w, http.StatusNotFound, appErr.Code, msg)
			} else {
				logger.Error("Failed to get user profile",
					logger.Field{Key: "user_id", Value: userID},
					logger.Field{Key: "error", Value: err.Error()})
				msg := i18n.Translate(lang, "auth.error.profile_fetch_failed", nil)
				response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
			}
		} else {
			logger.Error("Unexpected error type",
				logger.Field{Key: "user_id", Value: userID},
				logger.Field{Key: "error", Value: err.Error()})
			msg := i18n.Translate(lang, "auth.error.profile_fetch_failed", nil)
			response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
		}
		return
	}

	logger.Info("Profile fetched successfully",
		logger.Field{Key: "user_id", Value: userID},
		logger.Field{Key: "email", Value: profile.Email})

	// 3. Return success response
	response.Success(w, profile)
}

// getClientIP extracts the client's IP address from the request.
// It checks X-Forwarded-For, X-Real-IP headers before falling back to RemoteAddr.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain comma-separated IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr (format: "IP:port")
	if colonIdx := strings.LastIndex(r.RemoteAddr, ":"); colonIdx != -1 {
		return r.RemoteAddr[:colonIdx]
	}
	return r.RemoteAddr
}
