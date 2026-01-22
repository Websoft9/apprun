package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"apprun/ent/enttest"
	"apprun/pkg/cache"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsHandler_GetSnapshot tests the GetSnapshot handler
func TestMetricsHandler_GetSnapshot(t *testing.T) {
	// Setup test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Setup test cache
	cacheClient := cache.NewMockClient()

	// Create service and handler
	service := NewMetricsService(client, cacheClient, nil)
	handler := NewMetricsHandler(service)

	tests := []struct {
		name       string
		scope      string
		wantStatus int
	}{
		{"all scope", "all", http.StatusOK},
		{"users scope", "users", http.StatusOK},
		{"system scope", "system", http.StatusOK},
		{"default scope", "", http.StatusOK}, // defaults to "all"
		{"invalid scope", "invalid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/metrics/snapshot"
			if tt.scope != "" {
				url += "?scope=" + tt.scope
			}

			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.GetSnapshot(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Success bool              `json:"success"`
					Data    *SnapshotResponse `json:"data"`
				}
				err = json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.True(t, resp.Success)
				assert.NotNil(t, resp.Data)
			}
		})
	}
}

// TestMetricsHandler_GetKeys tests the GetKeys handler
func TestMetricsHandler_GetKeys(t *testing.T) {
	// Setup test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Setup test cache
	cacheClient := cache.NewMockClient()

	// Create service and handler
	service := NewMetricsService(client, cacheClient, nil)
	handler := NewMetricsHandler(service)

	// Create test request
	req, err := http.NewRequest("GET", "/api/metrics/keys", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetKeys(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var resp struct {
		Success bool          `json:"success"`
		Data    *KeysResponse `json:"data"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Greater(t, resp.Data.Count, 0)
	assert.NotEmpty(t, resp.Data.Keys)
}

// TestMetricsHandler_GetScopes tests the GetScopes handler
func TestMetricsHandler_GetScopes(t *testing.T) {
	// Setup test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Setup test cache
	cacheClient := cache.NewMockClient()

	// Create service and handler
	service := NewMetricsService(client, cacheClient, nil)
	handler := NewMetricsHandler(service)

	// Create test request
	req, err := http.NewRequest("GET", "/api/metrics/scopes", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetScopes(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var resp struct {
		Success bool            `json:"success"`
		Data    *ScopesResponse `json:"data"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, 4, resp.Data.Count) // users, system, auth, performance
	assert.Len(t, resp.Data.Scopes, 4)
}

// TestMetricsHandler_Snapshot_Caching tests that caching works for snapshot endpoint
func TestMetricsHandler_Snapshot_Caching(t *testing.T) {
	// Setup test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Setup test cache
	cacheClient := cache.NewMockClient()

	// Create service and handler
	service := NewMetricsService(client, cacheClient, nil)
	handler := NewMetricsHandler(service)

	// First request - should miss cache for users scope
	req1, err := http.NewRequest("GET", "/api/metrics/snapshot?scope=users", nil)
	require.NoError(t, err)
	rr1 := httptest.NewRecorder()
	handler.GetSnapshot(rr1, req1)

	var resp1 struct {
		Success bool              `json:"success"`
		Data    *SnapshotResponse `json:"data"`
	}
	err = json.Unmarshal(rr1.Body.Bytes(), &resp1)
	require.NoError(t, err)
	assert.False(t, resp1.Data.CacheHit) // First request should be cache miss

	// Second request - should hit cache
	req2, err := http.NewRequest("GET", "/api/metrics/snapshot?scope=users", nil)
	require.NoError(t, err)
	rr2 := httptest.NewRecorder()
	handler.GetSnapshot(rr2, req2)

	var resp2 struct {
		Success bool              `json:"success"`
		Data    *SnapshotResponse `json:"data"`
	}
	err = json.Unmarshal(rr2.Body.Bytes(), &resp2)
	require.NoError(t, err)
	assert.True(t, resp2.Data.CacheHit) // Second request should be cache hit
}
