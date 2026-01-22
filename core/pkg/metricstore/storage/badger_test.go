package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBadgerStorage(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path:        filepath.Join(tmpDir, "testdb"),
		TTL:         1 * time.Hour,
		Compression: true,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	require.NotNil(t, storage)
	defer storage.Close()

	// Verify config applied
	assert.Equal(t, cfg.Path, storage.cfg.Path)
	assert.Equal(t, cfg.TTL, storage.ttl)
}

func TestBadgerStorage_Store(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path: filepath.Join(tmpDir, "testdb"),
		TTL:  1 * time.Hour,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()
	metric := Metric{
		Name:      "test_metric",
		Value:     42.0,
		Tags:      map[string]string{"env": "test"},
		Timestamp: time.Now(),
	}

	err = storage.Store(ctx, metric)
	require.NoError(t, err)

	// Verify metric can be queried
	start := metric.Timestamp.Add(-1 * time.Minute)
	end := metric.Timestamp.Add(1 * time.Minute)
	metrics, err := storage.Query(ctx, "test_metric", start, end, 0)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	assert.Equal(t, metric.Name, metrics[0].Name)
	assert.Equal(t, metric.Value, metrics[0].Value)
}

func TestBadgerStorage_Query(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path: filepath.Join(tmpDir, "testdb"),
		TTL:  1 * time.Hour,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store multiple metrics at different times
	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		metric := Metric{
			Name:      "test_metric",
			Value:     float64(i),
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
		}
		err := storage.Store(ctx, metric) //nolint:govet // Shadow in test //nolint:govet // Shadow in test
		require.NoError(t, err)
	}

	// Query full range
	start := baseTime.Add(-1 * time.Minute)
	end := baseTime.Add(10 * time.Minute)
	metrics, err := storage.Query(ctx, "test_metric", start, end, 0)
	require.NoError(t, err)
	assert.Len(t, metrics, 5)

	// Verify sorted by timestamp
	for i := 0; i < len(metrics)-1; i++ {
		assert.True(t, metrics[i].Timestamp.Before(metrics[i+1].Timestamp))
	}

	// Query partial range
	start = baseTime.Add(2 * time.Minute)
	end = baseTime.Add(4 * time.Minute)
	metrics, err = storage.Query(ctx, "test_metric", start, end, 0)
	require.NoError(t, err)
	assert.Len(t, metrics, 3) // Metrics 2, 3, 4
}

func TestBadgerStorage_QueryNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path: filepath.Join(tmpDir, "testdb"),
		TTL:  1 * time.Hour,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Query non-existent metric
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	metrics, err := storage.Query(ctx, "nonexistent", start, end, 0)
	require.NoError(t, err)
	assert.Empty(t, metrics)
}

func TestBadgerStorage_Delete(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path: filepath.Join(tmpDir, "testdb"),
		TTL:  1 * time.Hour,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Store metrics at different times
	now := time.Now()
	for i := 0; i < 10; i++ {
		metric := Metric{
			Name:      "test_metric",
			Value:     float64(i),
			Timestamp: now.Add(time.Duration(i-5) * time.Hour), // -5h to +4h
		}
		err := storage.Store(ctx, metric) //nolint:govet // Shadow in test //nolint:govet // Shadow in test
		require.NoError(t, err)
	}

	// Delete metrics older than 2 hours ago
	cutoff := now.Add(-2 * time.Hour)
	err = storage.Delete(ctx, cutoff)
	require.NoError(t, err)

	// Verify remaining metrics
	start := now.Add(-10 * time.Hour)
	end := now.Add(10 * time.Hour)
	metrics, err := storage.Query(ctx, "test_metric", start, end, 0)
	require.NoError(t, err)
	assert.Len(t, metrics, 7) // Metrics from -1h to +4h
}

func TestBadgerStorage_Health(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path: filepath.Join(tmpDir, "testdb"),
		TTL:  1 * time.Hour,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Should be healthy
	err = storage.Health(ctx)
	require.NoError(t, err)
}

func TestBadgerStorage_Close(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path: filepath.Join(tmpDir, "testdb"),
		TTL:  1 * time.Hour,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Close storage
	err = storage.Close()
	require.NoError(t, err)

	// Operations after close should fail
	metric := Metric{
		Name:      "test",
		Value:     1.0,
		Timestamp: time.Now(),
	}
	err = storage.Store(ctx, metric)
	require.Error(t, err)

	err = storage.Health(ctx)
	require.Error(t, err)

	// Double close should be idempotent
	err = storage.Close()
	require.NoError(t, err)
}

func TestBadgerStorage_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path: filepath.Join(tmpDir, "testdb"),
		TTL:  1 * time.Hour,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage.Close()

	ctx := context.Background()

	// Concurrent writes
	const numGoroutines = 10
	const metricsPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			for i := 0; i < metricsPerGoroutine; i++ {
				metric := Metric{
					Name:      "concurrent_test",
					Value:     float64(id*1000 + i),
					Tags:      map[string]string{"goroutine": string(rune(id))},
					Timestamp: time.Now().Add(time.Duration(i) * time.Millisecond),
				}
				err := storage.Store(ctx, metric) //nolint:govet // Shadow in test //nolint:govet // Shadow in test
				assert.NoError(t, err)
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all metrics stored
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)
	metrics, err := storage.Query(ctx, "concurrent_test", start, end, 0)
	require.NoError(t, err)
	assert.Equal(t, numGoroutines*metricsPerGoroutine, len(metrics))
}

func TestBadgerStorage_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path: filepath.Join(tmpDir, "testdb"),
		TTL:  1 * time.Hour,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage.Close()

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	metric := Metric{
		Name:      "test",
		Value:     1.0,
		Timestamp: time.Now(),
	}

	// Operations should fail with canceled context
	err = storage.Store(ctx, metric)
	require.Error(t, err)
}

func TestBadgerStorage_GetStats(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := BadgerConfig{
		Path: filepath.Join(tmpDir, "testdb"),
		TTL:  1 * time.Hour,
	}

	storage, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage.Close()

	// Get stats
	stats := storage.GetStats()
	require.NotNil(t, stats)

	// Verify expected keys
	assert.Contains(t, stats, "lsm_size_bytes")
	assert.Contains(t, stats, "vlog_size_bytes")
	assert.Contains(t, stats, "total_size_bytes")
	assert.Contains(t, stats, "path")
	assert.Contains(t, stats, "ttl_seconds")
}

func TestBadgerStorage_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "testdb")

	cfg := BadgerConfig{
		Path: dbPath,
		TTL:  1 * time.Hour,
	}

	// Create and write metrics
	storage1, err := NewBadgerStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	metric := Metric{
		Name:      "persistent_metric",
		Value:     123.45,
		Timestamp: time.Now(),
	}

	err = storage1.Store(ctx, metric)
	require.NoError(t, err)

	// Close first instance
	err = storage1.Close()
	require.NoError(t, err)

	// Reopen database
	storage2, err := NewBadgerStorage(cfg)
	require.NoError(t, err)
	defer storage2.Close()

	// Verify metric persisted
	start := metric.Timestamp.Add(-1 * time.Minute)
	end := metric.Timestamp.Add(1 * time.Minute)
	metrics, err := storage2.Query(ctx, "persistent_metric", start, end, 0)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	assert.Equal(t, metric.Name, metrics[0].Name)
	assert.Equal(t, metric.Value, metrics[0].Value)
}

func TestNewStorage_Badger(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Backend: "badger",
		Timeout: 5 * time.Second,
		Settings: map[string]interface{}{
			"path": filepath.Join(tmpDir, "factorydb"),
			"ttl":  2 * time.Hour,
		},
	}

	storage, err := NewStorage(cfg)
	require.NoError(t, err)
	require.NotNil(t, storage)
	defer storage.Close()

	// Verify it's a BadgerStorage
	badgerStorage, ok := storage.(*BadgerStorage)
	require.True(t, ok)
	assert.Equal(t, 2*time.Hour, badgerStorage.ttl)
}

// Cleanup helper
func cleanupTestDB(path string) { //nolint:unused // Helper function for manual cleanup
	os.RemoveAll(path)
}
