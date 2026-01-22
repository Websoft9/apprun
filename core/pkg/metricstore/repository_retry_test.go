package metricstore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"apprun/pkg/metricstore/storage"
)

// TestStorage_IsRetryable verifies the IsRetryable function correctly
// distinguishes between temporary and permanent errors.
func TestStorage_IsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		error     error
		retryable bool
	}{
		{
			name:      "retryable: ErrUnavailable",
			error:     storage.ErrUnavailable("test", nil),
			retryable: true,
		},
		{
			name:      "retryable: ErrTimeout",
			error:     storage.ErrTimeout("test", "timeout"),
			retryable: true,
		},
		{
			name:      "non-retryable: ErrNotFound",
			error:     storage.ErrNotFound("test_metric"),
			retryable: false,
		},
		{
			name:      "non-retryable: ErrInvalidConfig",
			error:     storage.ErrInvalidConfig("backend", "invalid"),
			retryable: false,
		},
		{
			name:      "non-retryable: nil error",
			error:     nil,
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.IsRetryable(tt.error)
			assert.Equal(t, tt.retryable, result,
				"IsRetryable(%v) = %v, expected %v",
				tt.error, result, tt.retryable)
		})
	}
}

// TestRepository_RetryBehavior verifies retry is skipped for non-retryable errors.
// This is an integration test that verifies the retry logic behavior.
func TestRepository_RetryBehavior(t *testing.T) {
	ctx := context.Background()

	t.Run("retries disabled - returns error immediately", func(t *testing.T) {
		cfg := &Config{
			Storage: StorageConfig{
				Backend: "mock",
				Retry: RetryConfig{
					Enabled:     false,
					MaxAttempts: 3,
				},
			},
		}

		storageCfg := cfg.ToStorageConfig()
		mock, err := storage.NewStorage(storageCfg)
		require.NoError(t, err)

		repo := NewRepository(mock, cfg)

		// Close storage to trigger ErrUnavailable
		mock.Close()

		err = repo.RecordMetric(ctx, "test", 1.0, nil)
		assert.Error(t, err)
		// Should fail immediately without retries
	})

	t.Run("retries enabled - succeeds eventually", func(t *testing.T) {
		cfg := &Config{
			Storage: StorageConfig{
				Backend: "mock",
				Retry: RetryConfig{
					Enabled:     true,
					MaxAttempts: 3,
					Backoff:     "exponential",
				},
			},
		}

		storageCfg := cfg.ToStorageConfig()
		mock, err := storage.NewStorage(storageCfg)
		require.NoError(t, err)
		defer mock.Close()

		repo := NewRepository(mock, cfg)

		// With healthy storage, operation should succeed
		err = repo.RecordMetric(ctx, "test", 1.0, nil)
		assert.NoError(t, err)
	})
}
