// Package metricstore provides metrics storage and retrieval functionality for the application.
// It supports multiple storage backends and includes configuration management for metrics collection.
package metricstore

import (
	"context"
	"time"

	"apprun/pkg/logger"
	"apprun/pkg/metricstore/storage"
)

// Repository provides business logic layer for metrics operations.
// It wraps the storage interface and adds retry logic, logging, and convenience methods.
type Repository struct {
	storage storage.Storage
	cfg     *Config
	logger  logger.Logger
}

// NewRepository creates a new metrics repository.
//
// Example:
//
//	cfg, _ := metrics.LoadConfig()
//	storage, _ := storage.NewStorage(cfg.ToStorageConfig())
//	repo := metrics.NewRepository(storage, cfg)
//	defer repo.Close()
func NewRepository(storage storage.Storage, cfg *Config) *Repository { //nolint:gocritic // Parameter name is intentional
	return &Repository{
		storage: storage,
		cfg:     cfg,
		logger:  logger.L(),
	}
}

// RecordMetric stores a metric with the current timestamp.
// It wraps the storage Store method and adds retry logic.
//
// Example:
//
//	err := repo.RecordMetric(ctx, "api_requests_total", 1.0, map[string]string{
//	    "method": "GET",
//	    "endpoint": "/users",
//	})
func (r *Repository) RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error {
	metric := storage.Metric{
		Name:      name,
		Value:     value,
		Tags:      tags,
		Timestamp: time.Now(),
	}

	// Execute with retry
	err := r.executeWithRetry(ctx, func() error {
		return r.storage.Store(ctx, metric)
	})

	if err != nil {
		r.logger.Error("Failed to record metric",
			logger.Field{Key: "metric", Value: name},
			logger.Field{Key: "error", Value: err},
		)
		return err
	}

	r.logger.Debug("Metric recorded",
		logger.Field{Key: "metric", Value: name},
		logger.Field{Key: "value", Value: value},
	)

	return nil
}

// GetMetrics retrieves metrics by name for a specific duration looking backwards from now.
// It's a convenience wrapper around GetMetricsByRange.
//
// Example:
//
//	metrics, err := repo.GetMetrics(ctx, "api_requests_total", 1*time.Hour)
func (r *Repository) GetMetrics(ctx context.Context, name string, duration time.Duration) ([]storage.Metric, error) {
	end := time.Now()
	start := end.Add(-duration)
	return r.GetMetricsByRange(ctx, name, start, end, 0)
}

// GetMetricsByRange retrieves metrics by name within a specific time range.
// It wraps the storage Query method and adds retry logic.
// limit: Maximum results to return (0 = default 1000, max 10000)
//
// Example:
//
//	start := time.Now().Add(-2 * time.Hour)
//	end := time.Now()
//	metrics, err := repo.GetMetricsByRange(ctx, "api_requests_total", start, end, 1000)
func (r *Repository) GetMetricsByRange(ctx context.Context, name string, start, end time.Time, limit int) ([]storage.Metric, error) {
	var metrics []storage.Metric

	// Execute with retry
	err := r.executeWithRetry(ctx, func() error {
		var queryErr error
		metrics, queryErr = r.storage.Query(ctx, name, start, end, limit)
		return queryErr
	})

	if err != nil {
		r.logger.Error("Failed to query metrics",
			logger.Field{Key: "metric", Value: name},
			logger.Field{Key: "start", Value: start},
			logger.Field{Key: "end", Value: end},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	r.logger.Debug("Metrics queried",
		logger.Field{Key: "metric", Value: name},
		logger.Field{Key: "count", Value: len(metrics)},
	)

	return metrics, nil
}

// DeleteOldMetrics removes metrics older than the specified time.
// This is typically used for implementing retention policies.
//
// Example:
//
//	cutoff := time.Now().Add(-24 * time.Hour)
//	err := repo.DeleteOldMetrics(ctx, cutoff)
func (r *Repository) DeleteOldMetrics(ctx context.Context, before time.Time) error {
	err := r.executeWithRetry(ctx, func() error {
		return r.storage.Delete(ctx, before)
	})

	if err != nil {
		r.logger.Error("Failed to delete old metrics",
			logger.Field{Key: "before", Value: before},
			logger.Field{Key: "error", Value: err},
		)
		return err
	}

	r.logger.Info("Old metrics deleted",
		logger.Field{Key: "before", Value: before},
	)

	return nil
}

// Health checks if the storage backend is healthy.
func (r *Repository) Health(ctx context.Context) error {
	return r.storage.Health(ctx)
}

// Close releases resources held by the repository.
func (r *Repository) Close() error {
	return r.storage.Close()
}

// executeWithRetry executes a function with retry logic based on repository configuration.
func (r *Repository) executeWithRetry(ctx context.Context, fn func() error) error {
	if !r.cfg.Storage.Retry.Enabled {
		// No retries - execute once
		return fn()
	}

	maxAttempts := r.cfg.Storage.Retry.MaxAttempts
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return storage.ErrTimeout("operation", "context canceled")
		default:
		}

		lastErr = fn()
		if lastErr == nil {
			// Success
			return nil
		}

		// Check if error is retryable
		if !storage.IsRetryable(lastErr) {
			r.logger.Error("Operation failed with non-retryable error",
				logger.Field{Key: "error", Value: lastErr},
			)
			return lastErr
		}

		// Log retry attempt only for retryable errors
		if attempt < maxAttempts {
			backoff := r.calculateBackoff(attempt)
			r.logger.Warn("Operation failed, retrying",
				logger.Field{Key: "attempt", Value: attempt},
				logger.Field{Key: "max_attempts", Value: maxAttempts},
				logger.Field{Key: "backoff", Value: backoff},
				logger.Field{Key: "error", Value: lastErr},
			)

			// Wait before retry
			timer := time.NewTimer(backoff)
			select {
			case <-timer.C:
				// Continue to next attempt
			case <-ctx.Done():
				timer.Stop()
				return storage.ErrTimeout("operation", "context canceled during backoff")
			}
		}
	}

	// All retries exhausted
	r.logger.Error("Operation failed after retries",
		logger.Field{Key: "attempts", Value: maxAttempts},
		logger.Field{Key: "error", Value: lastErr},
	)

	return lastErr
}

// calculateBackoff returns the backoff duration for a given attempt number.
func (r *Repository) calculateBackoff(attempt int) time.Duration {
	baseDelay := 100 * time.Millisecond

	switch r.cfg.Storage.Retry.Backoff {
	case "exponential":
		// Exponential: 100ms, 200ms, 400ms, 800ms, ...
		if attempt > 10 { // Cap max backoff
			return baseDelay * time.Duration(1<<10)
		}
		return baseDelay * time.Duration(1<<uint(attempt-1)) //nolint:gosec // Capped above
	case "linear":
		// Linear: 100ms, 200ms, 300ms, 400ms, ...
		return baseDelay * time.Duration(attempt)
	default:
		return baseDelay
	}
}

// GetBackend returns the configured storage backend type.
func (r *Repository) GetBackend() string {
	return r.cfg.Storage.Backend
}
