package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"apprun/internal/jwt"
	"apprun/pkg/config"

	"github.com/spf13/viper"
)

func setupTestConfig() {
	viper.Set("jwt.secret", "test-secret-key-for-testing-purposes-only")
	viper.Set("jwt.access_token_expiration", "24h")
	viper.Set("jwt.issuer", "test-issuer")
	viper.Set("jwt.audience", "test-audience")
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	// Setup
	middleware := NewJWTMiddleware()
	handler := middleware.JWTAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_InvalidFormat(t *testing.T) {
	// Setup
	middleware := NewJWTMiddleware()
	handler := middleware.JWTAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidToken")
	rec := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	// Setup test config
	setupTestConfig()
	cfg := config.NewViperProvider(nil)

	// Generate valid token
	tokenSvc := jwt.NewTokenService(cfg)
	token, _, err := tokenSvc.GenerateToken(123, map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create middleware and handler
	middleware := NewJWTMiddleware()
	var capturedUserID int64
	handler := middleware.JWTAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = jwt.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
	if capturedUserID != 123 {
		t.Errorf("Expected user ID 123, got %d", capturedUserID)
	}
}

func TestJWTMiddleware_ContextInjection(t *testing.T) {
	setupTestConfig()
	// Setup test config
	cfg := config.NewViperProvider(nil)

	// Generate valid token
	tokenSvc := jwt.NewTokenService(cfg)
	token, _, err := tokenSvc.GenerateToken(456, map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create middleware and handler
	middleware := NewJWTMiddleware()
	var (
		capturedUserID   int64
		capturedUsername string
		capturedEmail    string
	)
	handler := middleware.JWTAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = jwt.GetUserID(r.Context())
		capturedUsername = jwt.GetUsername(r.Context())
		capturedEmail = jwt.GetEmail(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
	if capturedUserID != 456 {
		t.Errorf("Expected user ID 456, got %d", capturedUserID)
	}
	if capturedUsername != "alice" {
		t.Errorf("Expected username 'alice', got '%s'", capturedUsername)
	}
	if capturedEmail != "alice@example.com" {
		t.Errorf("Expected email 'alice@example.com', got '%s'", capturedEmail)
	}
}
