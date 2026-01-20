package obs

import (
	"context"
	"testing"

	"apprun/ent/enttest"
	"apprun/pkg/cache"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsService_GetUserMetrics_CacheMiss(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	cacheClient := cache.NewMockClient()
	service := NewMetricsService(client, cacheClient)

	ctx := context.Background()

	// Create test user
	_, err := client.User.Create().
		SetEmail("test@example.com").
		SetPasswordHash("hash").
		SetRole("platform_user").
		SetStatus(1).
		Save(ctx)
	require.NoError(t, err)

	// First call - cache miss
	metrics, err := service.GetUserMetrics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.False(t, metrics.CacheHit, "First call should be cache miss")
	assert.Equal(t, 1, metrics.TotalUsers)
}

func TestMetricsService_GetUserMetrics_CacheHit(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	cacheClient := cache.NewMockClient()
	service := NewMetricsService(client, cacheClient)

	ctx := context.Background()

	// Create test user
	_, err := client.User.Create().
		SetEmail("test@example.com").
		SetPasswordHash("hash").
		SetRole("platform_user").
		SetStatus(1).
		Save(ctx)
	require.NoError(t, err)

	// First call - cache miss
	metrics1, err := service.GetUserMetrics(ctx)
	require.NoError(t, err)
	assert.False(t, metrics1.CacheHit)

	// Second call - cache hit
	metrics2, err := service.GetUserMetrics(ctx)
	require.NoError(t, err)
	assert.True(t, metrics2.CacheHit, "Second call should be cache hit")
	assert.Equal(t, metrics1.TotalUsers, metrics2.TotalUsers)
}

func TestMetricsService_GetSystemMetrics(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	cacheClient := cache.NewMockClient()
	service := NewMetricsService(client, cacheClient)

	ctx := context.Background()

	// Get system metrics (no caching)
	metrics, err := service.GetSystemMetrics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.True(t, metrics.MemoryUsageMB > 0)
	assert.True(t, metrics.Goroutines > 0)
}

func TestMetricsService_GetAllMetrics_CacheHit(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	cacheClient := cache.NewMockClient()
	service := NewMetricsService(client, cacheClient)

	ctx := context.Background()

	// Create test data
	_, err := client.User.Create().
		SetEmail("admin@example.com").
		SetPasswordHash("hash").
		SetRole("platform_admin").
		SetStatus(1).
		Save(ctx)
	require.NoError(t, err)

	// First call - cache miss
	metrics1, err := service.GetAllMetrics(ctx)
	require.NoError(t, err)
	assert.False(t, metrics1.CacheHit)
	assert.Equal(t, 1, metrics1.TotalUsers)

	// Second call - cache hit
	metrics2, err := service.GetAllMetrics(ctx)
	require.NoError(t, err)
	assert.True(t, metrics2.CacheHit, "Second call should be cache hit")
	assert.Equal(t, metrics1.TotalUsers, metrics2.TotalUsers)
}

func TestMetricsService_CacheExpiration(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	mockCache := cache.NewMockClient()
	service := NewMetricsService(client, mockCache)

	ctx := context.Background()

	// Create test user
	_, err := client.User.Create().
		SetEmail("test@example.com").
		SetPasswordHash("hash").
		SetRole("platform_user").
		SetStatus(1).
		Save(ctx)
	require.NoError(t, err)

	// First call - cache miss
	metrics1, err := service.GetUserMetrics(ctx)
	require.NoError(t, err)
	assert.False(t, metrics1.CacheHit)

	// Simulate cache expiration
	mockCache.Reset()

	// After reset - cache miss again
	metrics2, err := service.GetUserMetrics(ctx)
	require.NoError(t, err)
	assert.False(t, metrics2.CacheHit, "After cache reset should be cache miss")
}

func TestMetricsService_PerformanceMetrics(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	cacheClient := cache.NewMockClient()
	service := NewMetricsService(client, cacheClient)

	ctx := context.Background()

	// Get performance metrics (stub)
	metrics, err := service.GetPerformanceMetrics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.APIRequestsTotal, "Stub should return 0")
}

func TestMetricsService_AuthMetrics(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	cacheClient := cache.NewMockClient()
	service := NewMetricsService(client, cacheClient)

	ctx := context.Background()

	// Get auth metrics (stub)
	metrics, err := service.GetAuthMetrics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.LoginAttemptsTotal, "Stub should return 0")
}
