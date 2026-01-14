package handler

import (
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"time"

	"apprun/internal/jwt"
	"apprun/pkg/errors"
	"apprun/pkg/i18n"
	"apprun/pkg/logger"
	"apprun/pkg/response"
)

// RefreshRequest holds the refresh token request.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGc..."`
}

// RefreshResponse holds the new token pair.
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`  // New JWT access token
	RefreshToken string `json:"refresh_token"` // New JWT refresh token
	ExpiresIn    int64  `json:"expires_in"`    // Seconds until access token expires
}

// Refresh handles token refresh requests (POST /api/auth/refresh).
// Validates refresh token, checks blacklist, verifies user status, and issues new token pair.
//
//	@Summary		Refresh access token
//	@Description	Exchange refresh token for new access/refresh token pair
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	response.Response{data=RefreshResponse}
//	@Failure		400		{object}	response.Response	"Missing or invalid refresh token"
//	@Failure		401		{object}	response.Response	"Token expired, invalid type, or blacklisted"
//	@Failure		403		{object}	response.Response	"Account disabled"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/api/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguage(r.Context())

	// 1. Parse request body
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Failed to parse refresh request",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "remote_addr", Value: r.RemoteAddr})
		msg := i18n.Translate(lang, "auth.error.invalid_request", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}

	// 2. Validate required fields
	if req.RefreshToken == "" {
		logger.Warn("Missing refresh token",
			logger.Field{Key: "remote_addr", Value: r.RemoteAddr})
		msg := i18n.Translate(lang, "auth.error.missing_refresh_token", nil)
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", msg)
		return
	}

	logger.Info("Refresh token request received",
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr})

	// 3. Validate refresh token and check token type
	claims, err := jwt.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		logger.Warn("Invalid refresh token",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "remote_addr", Value: r.RemoteAddr})

		switch {
		case stdErrors.Is(err, jwt.ErrTokenExpired):
			msg := i18n.Translate(lang, "auth.error.refresh_token_expired", nil)
			response.Error(w, http.StatusUnauthorized, errors.ErrCodeAuthTokenExpired, msg)
		case stdErrors.Is(err, jwt.ErrInvalidTokenType):
			msg := i18n.Translate(lang, "auth.error.invalid_token_type", nil)
			response.Error(w, http.StatusUnauthorized, errors.ErrCodeAuthInvalidTokenType, msg)
		default:
			msg := i18n.Translate(lang, "auth.error.invalid_refresh_token", nil)
			response.Error(w, http.StatusUnauthorized, errors.ErrCodeAuthInvalidToken, msg)
		}
		return
	}

	// 4. Check if token is blacklisted (if blacklist enabled)
	if jwt.IsBlacklistEnabled() {
		blacklisted, blacklistErr := jwt.IsBlacklisted(claims.ID)
		if blacklistErr != nil {
			logger.Error("Failed to check token blacklist",
				logger.Field{Key: "token_id", Value: claims.ID},
				logger.Field{Key: "error", Value: blacklistErr.Error()})
			// Continue anyway (fail open)
		}
		if blacklisted {
			logger.Warn("Attempted use of blacklisted refresh token",
				logger.Field{Key: "token_id", Value: claims.ID},
				logger.Field{Key: "user_id", Value: claims.UserID})
			msg := i18n.Translate(lang, "auth.error.token_revoked", nil)
			response.Error(w, http.StatusUnauthorized, errors.ErrCodeAuthTokenRevoked, msg)
			return
		}
	}

	// 5. Verify user still exists and is active
	userProfile, err := h.authService.GetUserProfile(r.Context(), claims.UserID)
	if err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			if appErr.Code == errors.ErrCodeAuthUserNotFound {
				logger.Warn("User not found for refresh token",
					logger.Field{Key: "user_id", Value: claims.UserID})
				msg := i18n.Translate(lang, "auth.error.user_not_found", nil)
				response.Error(w, http.StatusUnauthorized, appErr.Code, msg)
				return
			}
		}
		logger.Error("Failed to get user profile",
			logger.Field{Key: "user_id", Value: claims.UserID},
			logger.Field{Key: "error", Value: err.Error()})
		msg := i18n.Translate(lang, "auth.error.token_refresh_failed", nil)
		response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
		return
	}

	// Check if user is still active
	if userProfile.Status != 1 {
		logger.Warn("Refresh token for disabled account",
			logger.Field{Key: "user_id", Value: claims.UserID},
			logger.Field{Key: "status", Value: userProfile.Status})
		msg := i18n.Translate(lang, "auth.error.account_disabled", nil)
		response.Error(w, http.StatusForbidden, errors.ErrCodeAuthAccountDisabled, msg)
		return
	}

	// 6. Generate new token pair (token rotation)
	userClaims := map[string]interface{}{
		"username": claims.Username,
		"email":    claims.Email,
	}

	newAccessToken, newRefreshToken, expiresAt, err := jwt.GenerateTokenPair(claims.UserID, userClaims)
	if err != nil {
		logger.Error("Failed to generate new token pair",
			logger.Field{Key: "user_id", Value: claims.UserID},
			logger.Field{Key: "error", Value: err.Error()})
		msg := i18n.Translate(lang, "auth.error.token_generation_failed", nil)
		response.Error(w, http.StatusInternalServerError, errors.ErrCodeInternalError, msg)
		return
	}

	// 7. Blacklist old refresh token (one-time use policy, if blacklist enabled)
	if jwt.IsBlacklistEnabled() {
		remainingTTL := time.Until(claims.ExpiresAt.Time)
		if err := jwt.AddToBlacklist(claims.ID, remainingTTL); err != nil {
			logger.Error("Failed to blacklist old refresh token",
				logger.Field{Key: "token_id", Value: claims.ID},
				logger.Field{Key: "error", Value: err.Error()})
			// Continue anyway - new tokens already generated
		}
	}

	logger.Info("Token refresh successful",
		logger.Field{Key: "user_id", Value: claims.UserID},
		logger.Field{Key: "old_token_id", Value: claims.ID})

	// 8. Return new token pair
	expiresIn := int64(time.Until(expiresAt).Seconds())
	response.Success(w, RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
	})
}
