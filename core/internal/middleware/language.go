// Package middleware provides HTTP middleware components for request processing,
// including language detection and internationalization support.
package middleware

import (
	"net/http"

	"apprun/pkg/i18n"

	"golang.org/x/text/language"
)

// LanguageDetector is a middleware that detects user's language preference
// Priority: Query > Cookie > Accept-Language Header > Default
func LanguageDetector() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := detectLanguage(r)

			// Match language with support for variants (zh -> zh-CN)
			matchedLang := i18n.GetMatchedLanguage(lang)

			ctx := i18n.WithLanguage(r.Context(), matchedLang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// detectLanguage detects language from request
func detectLanguage(r *http.Request) string {
	// 1. Check query parameter (highest priority)
	if lang := r.URL.Query().Get("lang"); lang != "" {
		return lang
	}

	// 2. Check cookie
	if cookie, err := r.Cookie("lang"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 3. Parse Accept-Language header
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		lang := parseAcceptLanguage(acceptLang)
		if lang != "" {
			return lang
		}
	}

	// 4. Return default language
	return i18n.GetDefaultLanguage()
}

// parseAcceptLanguage parses Accept-Language header and returns the first supported language
// Example: "zh-CN,zh;q=0.9,en;q=0.8" -> "zh-CN"
// Uses golang.org/x/text/language for proper RFC 4647 compliance
func parseAcceptLanguage(acceptLang string) string {
	// Parse Accept-Language header
	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err != nil {
		return ""
	}

	// Find first supported language
	for _, tag := range tags {
		// Convert tag to string (e.g., "zh-CN")
		langCode := tag.String()
		if i18n.IsSupported(langCode) {
			return langCode
		}
	}

	return ""
}
