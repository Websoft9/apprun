package handlers

import (
	"net/http"

	"apprun/pkg/i18n"
	"apprun/pkg/response"
)

// I18nDemoHandler demonstrates i18n usage with different messages
//
//	@Summary		i18n Demo
//	@Description	Demonstrates internationalization with context-based translation
//	@Tags			demo
//	@Produce		json
//	@Param			lang	query		string	false	"Language code (en-US, zh-CN)"
//	@Success		200		{object}	response.Response{data=map[string]string}
//	@Router			/api/demo/i18n [get]
func I18nDemoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get current language from context
	lang := i18n.GetLanguage(ctx)

	// Translate multiple messages
	messages := map[string]string{
		"language":        lang,
		"welcome":         i18n.TranslateContext(ctx, "common.welcome", nil),
		"success_created": i18n.TranslateContext(ctx, "success.created", nil),
		"error_not_found": i18n.TranslateContext(ctx, "errors.not_found", nil),
		"error_invalid":   i18n.TranslateContext(ctx, "errors.invalid_param", nil),
	}

	response.Success(w, messages)
}
