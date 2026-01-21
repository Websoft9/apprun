// Package storage provides an anti-corruption layer for metrics persistence.
// It defines a storage abstraction interface that allows switching between
// different backend implementations (BadgerDB, Prometheus) without code changes.
//
// This follows the Anti-Corruption Layer pattern used in pkg/logger and pkg/cache.
package storage

import (
	"context"
	"time"
)

// Metric represents a single metric data point with metadata.
// It contains the metric name, numeric value, optional tags for dimensions,
// and the timestamp when the metric was recorded.
type Metric struct {
	// Name is the metric identifier (e.g., "http_requests_total")
	Name string

	// Value is the numeric measurement
	Value float64

	// Tags are key-value pairs for metric dimensions (e.g., {"method": "GET", "status": "200"})
	Tags map[string]string

	// Timestamp is when the metric was recorded
	Timestamp time.Time
}

// Storage defines the interface for metrics persistence backends.
// Implementations must be thread-safe and handle errors gracefully.
//
// All methods accept context.Context for cancellation and timeout support.
// Errors should use typed errors from pkg/errors for consistent handling.
type Storage interface {
	// Store persists a metric to the backend.
	// Returns an error if the metric cannot be stored.
	//
	// Example:
	//   err := storage.Store(ctx, storage.Metric{
	//       Name: "api_latency_ms",
	//       Value: 42.5,
	//       Tags: map[string]string{"endpoint": "/users"},
	//       Timestamp: time.Now(),
	//   })
	Store(ctx context.Context, metric Metric) error

	// Query retrieves metrics by name within a time range.
	// Returns an empty slice if no metrics match the criteria.
	// Returns ErrNotFound if the metric name doesn't exist.
	//
	// Parameters:
	//   - name: Metric identifier to query
	//   - start: Beginning of time range (inclusive)
	//   - end: End of time range (inclusive)
	//   - limit: Maximum number of results (0 = no limit, default 1000)
	//
	// Example:
	//   metrics, err := storage.Query(ctx, "api_latency_ms",
	//       time.Now().Add(-1*time.Hour), time.Now(), 1000)
	Query(ctx context.Context, name string, start, end time.Time, limit int) ([]Metric, error)

	// Delete removes metrics older than the specified time.
	// This is typically used for implementing retention policies.
	// Returns the number of metrics deleted (if supported by backend).
	//
	// Example:
	//   err := storage.Delete(ctx, time.Now().Add(-24*time.Hour))
	Delete(ctx context.Context, before time.Time) error

	// Health checks if the storage backend is available and responsive.
	// Returns ErrBackendUnavailable if the backend cannot be reached.
	//
	// Example:
	//   if err := storage.Health(ctx); err != nil {
	//       log.Error("Storage unhealthy", logger.Field{Key: "error", Value: err})
	//   }
	Health(ctx context.Context) error

	// Close releases resources held by the storage backend.
	// Should be called when the storage is no longer needed.
	// After Close is called, other methods should not be used.
	//
	// Example:
	//   defer storage.Close()
	Close() error
}

// Config holds configuration for creating a storage backend.
type Config struct {
	// Backend specifies which implementation to use ("badger", "prometheus", "mock")
	Backend string

	// Settings contains backend-specific configuration
	// For BadgerDB: {"path": "./data/metrics.db", "ttl": "24h"}
	// For Prometheus: {"remote_write_url": "http://localhost:9090/api/v1/write"}
	Settings map[string]interface{}

	// Timeout for storage operations (default: 5s)
	Timeout time.Duration

	// Retry configuration
	Retry RetryConfig
}

// RetryConfig defines retry behavior for transient failures.
type RetryConfig struct {
	// Enabled determines if retries are active
	Enabled bool

	// MaxAttempts is the maximum number of retry attempts (default: 3)
	MaxAttempts int

	// Backoff strategy: "exponential" or "linear"
	Backoff string
}

// NewStorage creates a storage backend based on the provided configuration.
// Returns an error if the backend type is unknown or configuration is invalid.
//
// Supported backends:
//   - "mock": In-memory storage for testing
//   - "badger": BadgerDB embedded storage (implemented in Story 9.3)
//   - "prometheus": Prometheus remote write (implemented in Story 9.4)
//
// Example:
//
//	cfg := storage.Config{
//	    Backend: "mock",
//	    Timeout: 5 * time.Second,
//	}
//	storage, err := storage.NewStorage(cfg)
//	if err != nil {
//	    log.Fatal("Failed to create storage", err)
//	}
//	defer storage.Close()
func NewStorage(cfg Config) (Storage, error) {
	// Validate backend
	if cfg.Backend == "" {
		return nil, ErrInvalidConfig("backend", "cannot be empty")
	}

	// Set defaults
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if !cfg.Retry.Enabled {
		cfg.Retry.MaxAttempts = 1 // No retries
	} else if cfg.Retry.MaxAttempts == 0 {
		cfg.Retry.MaxAttempts = 3
	}

	// Factory pattern: select backend implementation
	switch cfg.Backend {
	case "mock":
		// Mock backend is implemented in this package
		return newMockStorage(), nil
	case "badger":
		// BadgerDB implementation (Story 9.3)
		badgerCfg := BadgerConfig{
			Path:        "/var/lib/apprun/metrics", // Default path
			TTL:         24 * time.Hour,
			Compression: true,
			GCInterval:  5 * time.Minute,
		}

		// Override with settings from config
		if cfg.Settings != nil {
			if path, ok := cfg.Settings["path"].(string); ok {
				badgerCfg.Path = path
			}
			if ttl, ok := cfg.Settings["ttl"].(time.Duration); ok {
				badgerCfg.TTL = ttl
			}
			if compression, ok := cfg.Settings["compression"].(bool); ok {
				badgerCfg.Compression = compression
			}
			if gcInterval, ok := cfg.Settings["gc_interval"].(time.Duration); ok {
				badgerCfg.GCInterval = gcInterval
			}
		}

		return NewBadgerStorage(badgerCfg)
	case "prometheus":
		// Prometheus implementation will be added in Story 9.4
		return nil, ErrInvalidConfig("backend", "prometheus backend not yet implemented (Story 9.4)")
	default:
		return nil, ErrInvalidConfig("backend", "unknown backend: "+cfg.Backend)
	}
}
