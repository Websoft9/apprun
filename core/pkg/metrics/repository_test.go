package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"apprun/pkg/metrics/storage"
)

func TestNewRepository(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
			Retry: RetryConfig{
				Enabled:     true,
				MaxAttempts: 3,
				Backoff:     "exponential",
			},
		},
	}

	storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)
	defer storageMock.Close()

	repo := NewRepository(storageMock, cfg)
	require.NotNil(t, repo)
	assert.NotNil(t, repo.storage)
	assert.NotNil(t, repo.cfg)
	assert.NotNil(t, repo.logger)
}

func TestRepository_RecordMetric(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
			Retry: RetryConfig{
				Enabled: false,
			},
		},
	}

	storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)
	defer storageMock.Close()

	repo := NewRepository(storageMock, cfg)
	ctx := context.Background()

	// Record a metric
	err = repo.RecordMetric(ctx, "test_metric", 42.0, map[string]string{
		"env": "test",
	})
	require.NoError(t, err)

	// Verify it was stored
	metrics, err := storageMock.Query(ctx, "test_metric", time.Now().Add(-1*time.Minute), time.Now().Add(1*time.Minute), 0)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "test_metric", metrics[0].Name)
	assert.Equal(t, 42.0, metrics[0].Value)
	assert.Equal(t, "test", metrics[0].Tags["env"])
}

func TestRepository_RecordMetric_WithRetry(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
			Retry: RetryConfig{
				Enabled:     true,
				MaxAttempts: 3,
				Backoff:     "exponential",
			},
		},
	}

	storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)
	defer storageMock.Close()

	mockStorage := storageMock.(*storage.MockStorage)

	// Simulate failure, then success
	mockStorage.SetHealthy(false)

	repo := NewRepository(storageMock, cfg)
	ctx := context.Background()

	// First attempt fails, retry should work after setting healthy
	go func() {
		time.Sleep(50 * time.Millisecond)
		mockStorage.SetHealthy(true)
	}()

	err = repo.RecordMetric(ctx, "test_metric", 42.0, nil)
	require.NoError(t, err)
}

func TestRepository_GetMetrics(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
			Retry: RetryConfig{
				Enabled: false,
			},
		},
	}

	storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)
	defer storageMock.Close()

	repo := NewRepository(storageMock, cfg)
	ctx := context.Background()

	// Record some metrics at recent times to ensure they fall within query range
	now := time.Now()
	for i := 0; i < 5; i++ {
		metric := storage.Metric{
			Name:      "test_metric",
			Value:     float64(i),
			Timestamp: now.Add(-time.Duration(5-i) * time.Second), // -5s, -4s, -3s, -2s, -1s
		}
		err := storageMock.Store(ctx, metric)
		require.NoError(t, err)
	}

	// Query last 10 seconds
	metrics, err := repo.GetMetrics(ctx, "test_metric", 10*time.Second)
	require.NoError(t, err)
	assert.Len(t, metrics, 5)
}

func TestRepository_GetMetricsByRange(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
			Retry: RetryConfig{
				Enabled: false,
			},
		},
	}

	storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)
	defer storageMock.Close()

	repo := NewRepository(storageMock, cfg)
	ctx := context.Background()

	// Record metrics at different times
	baseTime := time.Now()
	for i := 0; i < 10; i++ {
		metric := storage.Metric{
			Name:      "test_metric",
			Value:     float64(i),
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
		}
		err := storageMock.Store(ctx, metric)
		require.NoError(t, err)
	}

	// Query specific range (minutes 3-7)
	start := baseTime.Add(3 * time.Minute)
	end := baseTime.Add(7 * time.Minute)
	metrics, err := repo.GetMetricsByRange(ctx, "test_metric", start, end, 0)
	require.NoError(t, err)
	assert.Len(t, metrics, 5) // Metrics 3, 4, 5, 6, 7

	// Verify values
	for i, m := range metrics {
		assert.Equal(t, float64(i+3), m.Value)
	}
}

func TestRepository_DeleteOldMetrics(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
			Retry: RetryConfig{
				Enabled: false,
			},
		},
	}

	storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)
	defer storageMock.Close()

	mockStorage := storageMock.(*storage.MockStorage)

	repo := NewRepository(storageMock, cfg)
	ctx := context.Background()

	// Record metrics at different times
	now := time.Now()
	for i := 0; i < 10; i++ {
		metric := storage.Metric{
			Name:      "test_metric",
			Value:     float64(i),
			Timestamp: now.Add(time.Duration(i-5) * time.Hour), // -5 to +4 hours
		}
		err := storageMock.Store(ctx, metric)
		require.NoError(t, err)
	}

	// Initial count
	assert.Equal(t, 10, mockStorage.Count())

	// Delete metrics older than 2 hours ago
	cutoff := now.Add(-2 * time.Hour)
	err = repo.DeleteOldMetrics(ctx, cutoff)
	require.NoError(t, err)

	// Verify deletion (should have 7 metrics left: -1h to +4h)
	assert.Equal(t, 7, mockStorage.Count())
}

func TestRepository_Health(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
		},
	}

	storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)
	defer storageMock.Close()

	mockStorage := storageMock.(*storage.MockStorage)

	repo := NewRepository(storageMock, cfg)
	ctx := context.Background()

	// Should be healthy
	err = repo.Health(ctx)
	require.NoError(t, err)

	// Make unhealthy
	mockStorage.SetHealthy(false)
	err = repo.Health(ctx)
	require.Error(t, err)
}

func TestRepository_Close(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
		},
	}

	storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)

	repo := NewRepository(storageMock, cfg)

	// Close should succeed
	err = repo.Close()
	require.NoError(t, err)

	// Operations after close should fail
	ctx := context.Background()
	err = repo.RecordMetric(ctx, "test", 1.0, nil)
	require.Error(t, err)
}

func TestRepository_ContextCancellation(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			Backend: "mock",
			Timeout: 5 * time.Second,
			Retry: RetryConfig{
				Enabled: false,
			},
		},
	}

	storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
	require.NoError(t, err)
	defer storageMock.Close()

	repo := NewRepository(storageMock, cfg)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Operation should fail immediately
	err = repo.RecordMetric(ctx, "test", 1.0, nil)
	require.Error(t, err)
}

func TestRepository_RetryBackoff(t *testing.T) {
	tests := []struct {
		name          string
		backoff       string
		attempt       int
		expectedMinMS int64
		expectedMaxMS int64
	}{
		{
			name:          "exponential_attempt_1",
			backoff:       "exponential",
			attempt:       1,
			expectedMinMS: 100,
			expectedMaxMS: 100,
		},
		{
			name:          "exponential_attempt_2",
			backoff:       "exponential",
			attempt:       2,
			expectedMinMS: 200,
			expectedMaxMS: 200,
		},
		{
			name:          "exponential_attempt_3",
			backoff:       "exponential",
			attempt:       3,
			expectedMinMS: 400,
			expectedMaxMS: 400,
		},
		{
			name:          "linear_attempt_1",
			backoff:       "linear",
			attempt:       1,
			expectedMinMS: 100,
			expectedMaxMS: 100,
		},
		{
			name:          "linear_attempt_2",
			backoff:       "linear",
			attempt:       2,
			expectedMinMS: 200,
			expectedMaxMS: 200,
		},
		{
			name:          "linear_attempt_3",
			backoff:       "linear",
			attempt:       3,
			expectedMinMS: 300,
			expectedMaxMS: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Storage: StorageConfig{
					Backend: "mock",
					Timeout: 5 * time.Second,
					Retry: RetryConfig{
						Enabled:     true,
						MaxAttempts: 3,
						Backoff:     tt.backoff,
					},
				},
			}

			storageMock, err := storage.NewStorage(cfg.ToStorageConfig())
			require.NoError(t, err)
			defer storageMock.Close()

			repo := NewRepository(storageMock, cfg)

			backoff := repo.calculateBackoff(tt.attempt)
			backoffMS := backoff.Milliseconds()

			assert.GreaterOrEqual(t, backoffMS, tt.expectedMinMS)
			assert.LessOrEqual(t, backoffMS, tt.expectedMaxMS)
		})
	}
}
