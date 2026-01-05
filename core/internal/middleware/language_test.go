package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"apprun/pkg/i18n"
)

func init() {
	// Initialize i18n for testing
	i18n.Init("en-US", []string{"en-US", "zh-CN"}, "")
}

func TestLanguageDetector_QueryParameter(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		w.Write([]byte(lang))
	}))

	req := httptest.NewRequest("GET", "/?lang=zh-CN", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "zh-CN" {
		t.Errorf("Expected 'zh-CN', got '%s'", rec.Body.String())
	}
}

func TestLanguageDetector_Cookie(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		w.Write([]byte(lang))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "zh-CN"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "zh-CN" {
		t.Errorf("Expected 'zh-CN', got '%s'", rec.Body.String())
	}
}

func TestLanguageDetector_AcceptLanguageHeader(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		w.Write([]byte(lang))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "zh-CN,en;q=0.9")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "zh-CN" {
		t.Errorf("Expected 'zh-CN', got '%s'", rec.Body.String())
	}
}

func TestLanguageDetector_DefaultLanguage(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		w.Write([]byte(lang))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "en-US" {
		t.Errorf("Expected 'en-US', got '%s'", rec.Body.String())
	}
}

func TestLanguageDetector_UnsupportedLanguageFallback(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		w.Write([]byte(lang))
	}))

	req := httptest.NewRequest("GET", "/?lang=fr-FR", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "en-US" {
		t.Errorf("Expected fallback to 'en-US', got '%s'", rec.Body.String())
	}
}

func TestLanguageDetector_Priority(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		w.Write([]byte(lang))
	}))

	// Query parameter should override cookie and header
	req := httptest.NewRequest("GET", "/?lang=en-US", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "zh-CN"})
	req.Header.Set("Accept-Language", "zh-CN")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "en-US" {
		t.Errorf("Expected query param 'en-US', got '%s'", rec.Body.String())
	}
}

func TestLanguageDetector_LanguageVariant(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		w.Write([]byte(lang))
	}))

	// Test "zh" should match to "zh-CN"
	req := httptest.NewRequest("GET", "/?lang=zh", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "zh-CN" {
		t.Errorf("Expected 'zh' to match 'zh-CN', got '%s'", rec.Body.String())
	}

	// Test "en" should match to "en-US"
	req = httptest.NewRequest("GET", "/?lang=en", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "en-US" {
		t.Errorf("Expected 'en' to match 'en-US', got '%s'", rec.Body.String())
	}
}
