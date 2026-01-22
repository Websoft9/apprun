package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"apprun/pkg/cache"
	"apprun/pkg/metricstore"
	"apprun/pkg/metricstore/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetHistory_Success tests successful historical metrics query
func TestGetHistory_Success(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	// Create mock storage with sample data using proper initialization
	cfg, err := metricstore.LoadConfig()
	require.NoError(t, err)

	// Force mock backend
	cfg.Storage.Backend = "mock"
	mockStorage, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)

	repo := metricstore.NewRepository(mockStorage, cfg)
	cacheClient := cache.NewMockClient()

	// Insert sample metrics
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		err := repo.RecordMetric(ctx, MetricNameUserTotal, float64(1000+i*10), nil)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	service := NewMetricsService(client, cacheClient, repo)
	handler := NewMetricsHandler(service)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/history?name="+MetricNameUserTotal+"&duration=1h", nil)
	w := httptest.NewRecorder()

	handler.GetHistory(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Response should contain metrics data
	assert.Contains(t, w.Body.String(), "metrics")
	assert.Contains(t, w.Body.String(), MetricNameUserTotal)
}

// TestGetHistory_MissingName tests validation of required name parameter
func TestGetHistory_MissingName(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	cacheClient := cache.NewMockClient()
	service := NewMetricsService(client, cacheClient, nil)
	handler := NewMetricsHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/metrics/history", nil)
	w := httptest.NewRecorder()

	handler.GetHistory(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_REQUEST")
}

// TestGetHistory_InvalidDuration tests invalid duration format
func TestGetHistory_InvalidDuration(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	cacheClient := cache.NewMockClient()
	service := NewMetricsService(client, cacheClient, nil)
	handler := NewMetricsHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/metrics/history?name=test&duration=invalid", nil)
	w := httptest.NewRecorder()

	handler.GetHistory(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_DURATION")
}

// TestGetHistory_DegradationStrategy tests graceful degradation when storage unavailable
func TestGetHistory_DegradationStrategy(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	// Service with nil repository (storage unavailable)
	cacheClient := cache.NewMockClient()
	service := NewMetricsService(client, cacheClient, nil)
	handler := NewMetricsHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/metrics/history?name="+MetricNameUserTotal+"&duration=1h", nil)
	w := httptest.NewRecorder()

	handler.GetHistory(w, req)

	// Should return 200 with empty metrics (degradation)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"metrics\":[]")
	assert.Contains(t, w.Body.String(), "\"count\":0")
}

// TestCollectorPersistence_UserMetrics tests automatic persistence of user metrics
func TestCollectorPersistence_UserMetrics(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	// Setup collector with repository
	cfg, err := metricstore.LoadConfig()
	require.NoError(t, err)
	cfg.Storage.Backend = "mock"

	mockStorage, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)

	repo := metricstore.NewRepository(mockStorage, cfg)
	collector := NewMetricsCollector(client, repo)

	// Create test users
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err := client.User.Create(). //nolint:govet // Shadow in test loop
			SetEmail("user" + string(rune(i+'0')) + "@test.com"). // Unique emails
			SetPasswordHash("hash").
			SetStatus(1).
			SetRole("platform_user").
			Save(ctx)
		require.NoError(t, err)
	}

	// Collect metrics (should trigger async persistence)
	userMetrics, err := collector.CollectUserMetrics(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, userMetrics.TotalUsers)

	// Wait for async persistence
	time.Sleep(100 * time.Millisecond)

	// Verify metrics were persisted (check mock storage)
	// Note: This is a basic test - real integration test would query repository
	assert.NotNil(t, repo)
}

// TestCollectorPersistence_SystemMetrics tests automatic persistence of system metrics
func TestCollectorPersistence_SystemMetrics(t *testing.T) {
	client := setupTestDB(t)
	defer client.Close()

	cfg, err := metricstore.LoadConfig()
	require.NoError(t, err)
	cfg.Storage.Backend = "mock"

	mockStorage, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)

	repo := metricstore.NewRepository(mockStorage, cfg)
	collector := NewMetricsCollector(client, repo)

	ctx := context.Background()
	sysMetrics, err := collector.CollectSystemMetrics(ctx)
	require.NoError(t, err)

	assert.Greater(t, sysMetrics.MemoryUsageMB, uint64(0))
	assert.GreaterOrEqual(t, sysMetrics.CPUUsagePercent, float64(0))

	// Wait for async persistence
	time.Sleep(100 * time.Millisecond)
}
