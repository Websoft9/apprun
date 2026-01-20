package obs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"apprun/ent/enttest"
	"apprun/pkg/cache"
	"apprun/pkg/response"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsHandler_GetAll tests the GetAll handler
func TestMetricsHandler_GetAll(t *testing.T) {
	// Setup test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Setup test cache
	cacheClient := cache.NewMockClient()

	// Create service and handler
	service := NewMetricsService(client, cacheClient)
	handler := NewMetricsHandler(service)

	// Create test request
	req, err := http.NewRequest("GET", "/api/metrics", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetAll(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var resp response.Response
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
}

// TestMetricsHandler_GetUsers tests the GetUsers handler
func TestMetricsHandler_GetUsers(t *testing.T) {
	// Setup test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Create test users
	ctx := context.Background()
	_, err := client.User.Create().
		SetUsername("testuser").
		SetPasswordHash("hashedpass").
		SetEmail("test@example.com").
		SetRole("platform_admin").
		SetStatus(1).
		Save(ctx)
	require.NoError(t, err)

	// Setup test cache
	cacheClient := cache.NewMockClient()

	// Create service and handler
	service := NewMetricsService(client, cacheClient)
	handler := NewMetricsHandler(service)

	// Create test request
	req, err := http.NewRequest("GET", "/api/metrics/users", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetUsers(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var resp struct {
		Success bool         `json:"success"`
		Data    *UserMetrics `json:"data"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify response
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, 1, resp.Data.TotalUsers)
	assert.Equal(t, 1, resp.Data.ActiveUsers)
	assert.Equal(t, 1, resp.Data.AdminUsers)
}

// TestMetricsHandler_GetSystem tests the GetSystem handler
func TestMetricsHandler_GetSystem(t *testing.T) {
	// Setup test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Setup test cache
	cacheClient := cache.NewMockClient()

	// Create service and handler
	service := NewMetricsService(client, cacheClient)
	handler := NewMetricsHandler(service)

	// Create test request
	req, err := http.NewRequest("GET", "/api/metrics/system", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetSystem(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var resp struct {
		Success bool           `json:"success"`
		Data    *SystemMetrics `json:"data"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify response
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.GreaterOrEqual(t, resp.Data.CPUUsagePercent, 0.0)
	assert.GreaterOrEqual(t, resp.Data.MemoryUsageMB, uint64(0))
}

// TestMetricsHandler_GetPerformance tests the GetPerformance handler
func TestMetricsHandler_GetPerformance(t *testing.T) {
	// Setup test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Setup test cache
	cacheClient := cache.NewMockClient()

	// Create service and handler
	service := NewMetricsService(client, cacheClient)
	handler := NewMetricsHandler(service)

	// Create test request
	req, err := http.NewRequest("GET", "/api/metrics/performance", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetPerformance(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse response
	var resp struct {
		Success bool                `json:"success"`
		Data    *PerformanceMetrics `json:"data"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify response
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
}

// TestMetricsHandler_Caching tests that caching works across multiple requests
func TestMetricsHandler_Caching(t *testing.T) {
	// Setup test database
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	// Setup test cache
	cacheClient := cache.NewMockClient()

	// Create service and handler
	service := NewMetricsService(client, cacheClient)
	handler := NewMetricsHandler(service)

	// First request - should miss cache
	req1, err := http.NewRequest("GET", "/api/metrics/users", nil)
	require.NoError(t, err)
	rr1 := httptest.NewRecorder()
	handler.GetUsers(rr1, req1)

	var resp1 struct {
		Success bool         `json:"success"`
		Data    *UserMetrics `json:"data"`
	}
	err = json.Unmarshal(rr1.Body.Bytes(), &resp1)
	require.NoError(t, err)
	assert.False(t, resp1.Data.CacheHit) // First request should be cache miss

	// Second request - should hit cache
	req2, err := http.NewRequest("GET", "/api/metrics/users", nil)
	require.NoError(t, err)
	rr2 := httptest.NewRecorder()
	handler.GetUsers(rr2, req2)

	var resp2 struct {
		Success bool         `json:"success"`
		Data    *UserMetrics `json:"data"`
	}
	err = json.Unmarshal(rr2.Body.Bytes(), &resp2)
	require.NoError(t, err)
	assert.True(t, resp2.Data.CacheHit) // Second request should be cache hit
}
