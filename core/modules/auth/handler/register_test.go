package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestRegister_InvalidJSON tests malformed JSON input
func TestRegister_InvalidJSON(t *testing.T) {
	handler := &AuthHandler{
		authService: nil, // Not needed for JSON parsing test
	}

	// Create malformed JSON
	body := []byte(`{"email": "test@example.com", "password": `)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := chi.NewRouter()
	router.Post("/api/auth/register", handler.Register)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Register() status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// TestRegister_MissingFields tests required field validation
func TestRegister_MissingFields(t *testing.T) {
	handler := &AuthHandler{
		authService: nil,
	}

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{"missing email", map[string]interface{}{"password": "SecurePass123"}},
		{"missing password", map[string]interface{}{"email": "test@example.com"}},
		{"empty email", map[string]interface{}{"email": "", "password": "SecurePass123"}},
		{"empty password", map[string]interface{}{"email": "test@example.com", "password": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router := chi.NewRouter()
			router.Post("/api/auth/register", handler.Register)

			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Register() status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
		})
	}
}

// Note: Full integration tests with database are in integration_test_story18.sh
// These unit tests focus on HTTP layer validation
