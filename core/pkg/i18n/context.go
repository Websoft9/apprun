// Package i18n provides internationalization support for the application,
// including context-based language management.
package i18n

import (
	"context"
)

type contextKey string

const languageKey contextKey = "language"

// WithLanguage stores the language preference in context
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, languageKey, lang)
}

// GetLanguage retrieves the language from context, returns default if not set
func GetLanguage(ctx context.Context) string {
	if lang, ok := ctx.Value(languageKey).(string); ok {
		return lang
	}
	return GetDefaultLanguage()
}

// TranslateContext is a convenience function that gets language from context and translates
func TranslateContext(ctx context.Context, messageID string, data map[string]interface{}) string {
	lang := GetLanguage(ctx)
	return Translate(lang, messageID, data)
}
