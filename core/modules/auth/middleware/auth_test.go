package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"apprun/internal/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTAuthMiddleware(t *testing.T) {
	cfg := &jwt.Config{
		Secret:         "test-secret-key-with-32-chars!",
		Expiry:         "24h",
		Issuer:         "apprun-test",
		WhitelistPaths: []string{"/api/v1/auth/register", "/api/v1/auth/login", "/health", "/metrics"},
	}

	middleware, err := NewJWTMiddlewareFromConfig(cfg)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := jwt.GetUserID(r.Context())
		username := jwt.GetUsername(r.Context())
		assert.Equal(t, int64(123), userID)
		assert.Equal(t, "testuser", username)
		w.WriteHeader(http.StatusOK)
	})

	runtimeCfg, _ := cfg.ToRuntimeConfig()
	token, err := jwt.GenerateToken(runtimeCfg, 123, "testuser", "test@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.JWTAuth(handler).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthMiddlewareNoToken(t *testing.T) {
	cfg := &jwt.Config{
		Secret:         "test-secret-key-with-32-chars!",
		Expiry:         "24h",
		Issuer:         "apprun-test",
		WhitelistPaths: []string{"/api/v1/auth/register", "/api/v1/auth/login", "/health", "/metrics"},
	}

	middleware, err := NewJWTMiddlewareFromConfig(cfg)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	w := httptest.NewRecorder()

	middleware.JWTAuth(handler).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddlewareInvalidToken(t *testing.T) {
	cfg := &jwt.Config{
		Secret:         "test-secret-key-with-32-chars!",
		Expiry:         "24h",
		Issuer:         "apprun-test",
		WhitelistPaths: []string{"/api/v1/auth/register", "/api/v1/auth/login", "/health", "/metrics"},
	}

	middleware, err := NewJWTMiddlewareFromConfig(cfg)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	middleware.JWTAuth(handler).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddlewareExpiredToken(t *testing.T) {
	cfg := &jwt.Config{
		Secret:         "test-secret-key-with-32-chars!",
		Expiry:         "-1h", // Negative duration for expired token
		Issuer:         "apprun-test",
		WhitelistPaths: []string{"/api/v1/auth/register", "/api/v1/auth/login", "/health", "/metrics"},
	}

	middleware, err := NewJWTMiddlewareFromConfig(cfg)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	runtimeCfg, _ := cfg.ToRuntimeConfig()
	token, err := jwt.GenerateToken(runtimeCfg, 123, "testuser", "test@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.JWTAuth(handler).ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestJWTAuthMiddlewareWhitelist(t *testing.T) {
	cfg := &jwt.Config{
		Secret:         "test-secret-key-with-32-chars!",
		Expiry:         "24h",
		Issuer:         "apprun-test",
		WhitelistPaths: []string{"/api/v1/auth/register", "/api/v1/auth/login", "/health", "/metrics"},
	}

	middleware, err := NewJWTMiddlewareFromConfig(cfg)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test all configured whitelist paths
	for _, path := range cfg.WhitelistPaths {
		req := httptest.NewRequest("POST", path, nil)
		w := httptest.NewRecorder()

		middleware.JWTAuth(handler).ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Path %s should be whitelisted", path)
	}
}

func TestJWTAuthMiddlewareCustomWhitelist(t *testing.T) {
	cfg := &jwt.Config{
		Secret:         "test-secret-key-with-32-chars!",
		Expiry:         "24h",
		Issuer:         "apprun-test",
		WhitelistPaths: []string{"/custom/path", "/another/public"},
	}

	middleware, err := NewJWTMiddlewareFromConfig(cfg)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Custom whitelist should work
	req := httptest.NewRequest("GET", "/custom/path", nil)
	w := httptest.NewRecorder()
	middleware.JWTAuth(handler).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Default paths should NOT be whitelisted if not in config
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	middleware.JWTAuth(handler).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Default /health should NOT be whitelisted when custom config is used")
}

func TestJWTAuthMiddlewareInvalidFormat(t *testing.T) {
	cfg := &jwt.Config{
		Secret:         "test-secret-key-with-32-chars!",
		Expiry:         "24h",
		Issuer:         "apprun-test",
		WhitelistPaths: []string{"/api/v1/auth/register", "/api/v1/auth/login", "/health", "/metrics"},
	}

	middleware, err := NewJWTMiddlewareFromConfig(cfg)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	w := httptest.NewRecorder()

	middleware.JWTAuth(handler).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
