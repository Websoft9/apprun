package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"apprun/pkg/i18n"
)

const (
	testLangZhCN = "zh-CN"
	testLangEnUS = "en-US"
)

func init() {
	// Initialize i18n for testing
	if err := i18n.Init("en-US", []string{"en-US", "zh-CN"}, ""); err != nil {
		panic("Failed to initialize i18n: " + err.Error())
	}
}

func TestLanguageDetector_QueryParameter(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		if _, err := w.Write([]byte(lang)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/?lang="+testLangZhCN, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != testLangZhCN {
		t.Errorf("Expected '%s', got '%s'", testLangZhCN, rec.Body.String())
	}
}

func TestLanguageDetector_Cookie(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		if _, err := w.Write([]byte(lang)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: testLangZhCN})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != testLangZhCN {
		t.Errorf("Expected '%s', got '%s'", testLangZhCN, rec.Body.String())
	}
}

func TestLanguageDetector_AcceptLanguageHeader(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		if _, err := w.Write([]byte(lang)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", testLangZhCN+",en;q=0.9")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != testLangZhCN {
		t.Errorf("Expected '%s', got '%s'", testLangZhCN, rec.Body.String())
	}
}

func TestLanguageDetector_DefaultLanguage(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		if _, err := w.Write([]byte(lang)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != testLangEnUS {
		t.Errorf("Expected '%s', got '%s'", testLangEnUS, rec.Body.String())
	}
}

func TestLanguageDetector_UnsupportedLanguageFallback(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		if _, err := w.Write([]byte(lang)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/?lang=fr-FR", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != testLangEnUS {
		t.Errorf("Expected fallback to '%s', got '%s'", testLangEnUS, rec.Body.String())
	}
}

func TestLanguageDetector_Priority(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		if _, err := w.Write([]byte(lang)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	// Query parameter should override cookie and header
	req := httptest.NewRequest("GET", "/?lang="+testLangEnUS, nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: testLangZhCN})
	req.Header.Set("Accept-Language", testLangZhCN)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != testLangEnUS {
		t.Errorf("Expected query param '%s', got '%s'", testLangEnUS, rec.Body.String())
	}
}

func TestLanguageDetector_LanguageVariant(t *testing.T) {
	handler := LanguageDetector()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.GetLanguage(r.Context())
		if _, err := w.Write([]byte(lang)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	// Test "zh" should match to "zh-CN"
	req := httptest.NewRequest("GET", "/?lang=zh", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != testLangZhCN {
		t.Errorf("Expected 'zh' to match '%s', got '%s'", testLangZhCN, rec.Body.String())
	}

	// Test "en" should match to "en-US"
	req = httptest.NewRequest("GET", "/?lang=en", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != testLangEnUS {
		t.Errorf("Expected 'en' to match '%s', got '%s'", testLangEnUS, rec.Body.String())
	}
}
