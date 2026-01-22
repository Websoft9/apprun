package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid mock backend",
			config: Config{
				Backend: "mock",
				Timeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty backend",
			config: Config{
				Backend: "",
			},
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name: "badger backend implemented",
			config: Config{
				Backend: "badger",
				Settings: map[string]interface{}{
					"path": t.TempDir() + "/testdb",
				},
			},
			wantErr: false,
		},
		{
			name: "prometheus backend not implemented",
			config: Config{
				Backend: "prometheus",
			},
			wantErr:     true,
			errContains: "not yet implemented",
		},
		{
			name: "unknown backend",
			config: Config{
				Backend: "unknown",
			},
			wantErr:     true,
			errContains: "unknown backend",
		},
		{
			name: "defaults applied",
			config: Config{
				Backend: "mock",
				// No timeout specified
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewStorage(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, storage)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, storage)
				defer storage.Close()
			}
		})
	}
}

func TestMockStorage_Store(t *testing.T) {
	storage := newMockStorage()
	defer storage.Close()

	ctx := context.Background()

	metric := Metric{
		Name:      "test_metric",
		Value:     42.0,
		Tags:      map[string]string{"env": "test"},
		Timestamp: time.Now(),
	}

	err := storage.Store(ctx, metric)
	require.NoError(t, err)

	// Verify metric was stored
	assert.Equal(t, 1, storage.Count())
}

func TestMockStorage_StoreAfterClose(t *testing.T) {
	storage := newMockStorage()
	storage.Close()

	ctx := context.Background()
	metric := Metric{
		Name:      "test_metric",
		Value:     42.0,
		Timestamp: time.Now(),
	}

	err := storage.Store(ctx, metric)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unavailable")
}

func TestMockStorage_Query(t *testing.T) {
	storage := newMockStorage()
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	// Store test metrics
	metrics := []Metric{
		{Name: "test_metric", Value: 10.0, Timestamp: now.Add(-3 * time.Hour)},
		{Name: "test_metric", Value: 20.0, Timestamp: now.Add(-2 * time.Hour)},
		{Name: "test_metric", Value: 30.0, Timestamp: now.Add(-1 * time.Hour)},
		{Name: "other_metric", Value: 100.0, Timestamp: now.Add(-1 * time.Hour)},
	}

	for _, m := range metrics {
		require.NoError(t, storage.Store(ctx, m))
	}

	// Query time range
	start := now.Add(-2*time.Hour - 30*time.Minute)
	end := now.Add(-30 * time.Minute)

	results, err := storage.Query(ctx, "test_metric", start, end, 0)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Verify results are sorted by timestamp
	assert.Equal(t, 20.0, results[0].Value)
	assert.Equal(t, 30.0, results[1].Value)
}

func TestMockStorage_QueryNotFound(t *testing.T) {
	storage := newMockStorage()
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	results, err := storage.Query(ctx, "nonexistent", now.Add(-1*time.Hour), now, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, results)
}

func TestMockStorage_QueryAfterClose(t *testing.T) {
	storage := newMockStorage()
	storage.Close()

	ctx := context.Background()
	now := time.Now()

	results, err := storage.Query(ctx, "test", now.Add(-1*time.Hour), now, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unavailable")
	assert.Nil(t, results)
}

func TestMockStorage_Delete(t *testing.T) {
	storage := newMockStorage()
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	// Store test metrics
	metrics := []Metric{
		{Name: "test_metric", Value: 10.0, Timestamp: now.Add(-3 * time.Hour)},
		{Name: "test_metric", Value: 20.0, Timestamp: now.Add(-2 * time.Hour)},
		{Name: "test_metric", Value: 30.0, Timestamp: now.Add(-1 * time.Hour)},
	}

	for _, m := range metrics {
		require.NoError(t, storage.Store(ctx, m))
	}

	// Delete metrics older than 90 minutes ago
	cutoff := now.Add(-90 * time.Minute)
	err := storage.Delete(ctx, cutoff)
	require.NoError(t, err)

	// Query remaining metrics
	results, err := storage.Query(ctx, "test_metric", now.Add(-4*time.Hour), now, 0)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, 30.0, results[0].Value)
}

func TestMockStorage_Health(t *testing.T) {
	storage := newMockStorage()
	defer storage.Close()

	ctx := context.Background()

	// Healthy by default
	err := storage.Health(ctx)
	require.NoError(t, err)

	// Mark unhealthy
	storage.SetHealthy(false)
	err = storage.Health(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unavailable")

	// Mark healthy again
	storage.SetHealthy(true)
	err = storage.Health(ctx)
	require.NoError(t, err)
}

func TestMockStorage_HealthAfterClose(t *testing.T) {
	storage := newMockStorage()
	storage.Close()

	ctx := context.Background()

	err := storage.Health(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unavailable")
}

func TestMockStorage_Close(t *testing.T) {
	storage := newMockStorage()

	err := storage.Close()
	require.NoError(t, err)

	// Verify storage is closed
	ctx := context.Background()
	err = storage.Health(ctx)
	require.Error(t, err)
}

func TestMockStorage_Reset(t *testing.T) {
	storage := newMockStorage()
	defer storage.Close()

	ctx := context.Background()

	// Store metrics
	metric := Metric{
		Name:      "test_metric",
		Value:     42.0,
		Timestamp: time.Now(),
	}
	require.NoError(t, storage.Store(ctx, metric))
	assert.Equal(t, 1, storage.Count())

	// Reset storage
	storage.Reset()
	assert.Equal(t, 0, storage.Count())

	// Verify storage is healthy after reset
	err := storage.Health(ctx)
	require.NoError(t, err)
}

func TestMockStorage_ConcurrentAccess(t *testing.T) {
	storage := newMockStorage()
	defer storage.Close()

	ctx := context.Background()
	now := time.Now()

	// Launch 10 goroutines writing concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				metric := Metric{
					Name:      "concurrent_metric",
					Value:     float64(id*100 + j),
					Timestamp: now.Add(time.Duration(j) * time.Millisecond),
				}
				storage.Store(ctx, metric)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all metrics were stored
	assert.Equal(t, 1000, storage.Count())

	// Query should work without race conditions
	results, err := storage.Query(ctx, "concurrent_metric", now, now.Add(1*time.Second), 0)
	require.NoError(t, err)
	assert.Len(t, results, 1000)
}

func TestMockStorage_ContextCancellation(t *testing.T) {
	storage := newMockStorage()
	defer storage.Close()

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	metric := Metric{
		Name:      "test_metric",
		Value:     42.0,
		Timestamp: time.Now(),
	}

	// Store should respect context cancellation
	err := storage.Store(ctx, metric)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
