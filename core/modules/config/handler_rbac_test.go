package config_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/httptest"
	"testing"

	"apprun/ent"
	"apprun/ent/enttest"
	"apprun/internal/jwt"
	"apprun/internal/rbac"
	"apprun/modules/config"
	"apprun/pkg/response"
	"apprun/routes"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a test router with config routes and RBAC protection
func setupTestRouter(t *testing.T) (*chi.Mux, *ent.Client, *config.Service) {
	// Create test database client
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() {
		client.Close()
	})

	// Initialize RBAC enforcer
	enforcer := rbac.GetEnforcer()
	enforcer.LoadPolicy()

	// Create config service
	tmpDir := t.TempDir()
	repo := config.NewRepository(client)
	loader, err := config.NewLoader(tmpDir, repo)
	require.NoError(t, err)
	configService := config.NewService(loader, repo)

	// Setup routes (this will include RBAC protection from Story 3-2-1)
	router := routes.SetupRoutes(client, configService, nil)

	return router, client, configService
}

// createTestUser creates a test user and returns a JWT token
func createTestUser(t *testing.T, client *ent.Client, isPlatformAdmin bool) string {
	// Create test user
	user, err := client.User.Create().
		SetEmail("test@example.com").
		SetUsername("testuser").
		SetPasswordHash("hashed").
		SetStatus("active").
		Save(context.Background())
	require.NoError(t, err)

	// Assign platform_admin role if requested
	if isPlatformAdmin {
		enforcer := rbac.GetEnforcer()
		_, err := enforcer.AddRoleForUser(user.ID.String(), "platform_admin", "platform")
		require.NoError(t, err)
	}

	// Generate JWT token
	token, err := jwt.GenerateToken(jwt.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
	})
	require.NoError(t, err)

	return token
}

// TestConfigAPI_Unauthenticated tests that config APIs require authentication
func TestConfigAPI_Unauthenticated(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GET config", http.MethodGet, "/api/config?key=test.key"},
		{"PUT config", http.MethodPut, "/api/config"},
		{"DELETE config", http.MethodDelete, "/api/config?key=test.key"},
		{"LIST configs", http.MethodGet, "/api/config/list"},
		{"GET allowed keys", http.MethodGet, "/api/config/allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPut {
				body := []byte(`{"key":"test.key","value":"test"}`)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "Should return 401 without JWT token")
		})
	}
}

// TestConfigAPI_Unauthorized tests that non-admin users cannot access config APIs
func TestConfigAPI_Unauthorized(t *testing.T) {
	router, client, _ := setupTestRouter(t)

	// Create regular user (not platform admin)
	token := createTestUser(t, client, false)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GET config", http.MethodGet, "/api/config?key=test.key"},
		{"PUT config", http.MethodPut, "/api/config"},
		{"DELETE config", http.MethodDelete, "/api/config?key=test.key"},
		{"LIST configs", http.MethodGet, "/api/config/list"},
		{"GET allowed keys", http.MethodGet, "/api/config/allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPut {
				body := []byte(`{"key":"app.theme","value":"dark"}`)
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "Should return 403 for non-admin user")
		})
	}
}

// TestConfigAPI_PlatformAdmin_CanRead tests that platform admin can read configs
func TestConfigAPI_PlatformAdmin_CanRead(t *testing.T) {
	router, client, configService := setupTestRouter(t)

	// Create platform admin user
	token := createTestUser(t, client, true)

	// Set a test config value
	ctx := context.Background()
	err := configService.UpdateConfig(ctx, "app.theme", "light")
	require.NoError(t, err)

	// Test GET config
	t.Run("GET single config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/config?key=app.theme", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	// Test LIST configs
	t.Run("LIST configs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/config/list", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	// Test GET allowed keys
	t.Run("GET allowed keys", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/config/allowed", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})
}

// TestConfigAPI_PlatformAdmin_CanWrite tests that platform admin can modify configs
func TestConfigAPI_PlatformAdmin_CanWrite(t *testing.T) {
	router, client, _ := setupTestRouter(t)

	// Create platform admin user
	token := createTestUser(t, client, true)

	// Test PUT config
	t.Run("PUT config", func(t *testing.T) {
		body := []byte(`{"key":"app.theme","value":"dark"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	// Test DELETE config
	t.Run("DELETE config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/config?key=app.theme", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})
}

// TestConfigAPI_AuditLog_Created tests that audit logs are created for config changes
func TestConfigAPI_AuditLog_Created(t *testing.T) {
	router, client, _ := setupTestRouter(t)

	// Create platform admin user
	token := createTestUser(t, client, true)

	// Update a config
	body := []byte(`{"key":"app.theme","value":"dark"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Check audit log created
	ctx := context.Background()
	count, err := client.AuditLog.Query().
		Where(func(s *ent.AuditLogQuery) {
			s.Where(enttest.HasAction("config.update"))
		}).
		Count(ctx)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "Audit log should be created for config update")
}

// TestConfigAPI_AuditLog_Delete tests that audit logs are created for config deletion
func TestConfigAPI_AuditLog_Delete(t *testing.T) {
	router, client, configService := setupTestRouter(t)

	// Create platform admin user
	token := createTestUser(t, client, true)

	// Create a config first
	ctx := context.Background()
	err := configService.UpdateConfig(ctx, "test.key", "test.value")
	require.NoError(t, err)

	// Delete the config
	req := httptest.NewRequest(http.MethodDelete, "/api/config?key=test.key", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Check audit log created
	count, err := client.AuditLog.Query().
		Where(func(s *ent.AuditLogQuery) {
			s.Where(enttest.HasAction("config.delete"))
		}).
		Count(ctx)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "Audit log should be created for config deletion")
}
