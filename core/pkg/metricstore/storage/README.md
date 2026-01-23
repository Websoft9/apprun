# Metrics Storage Package

This package provides an anti-corruption layer (ACL) for metrics storage backends in the apprun platform. It defines a pluggable storage interface that allows switching between different storage implementations without affecting business logic.

## Architecture

```
┌─────────────────┐
│   Application   │
│   (Business)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Repository    │  ← Retry, Logging, Convenience Methods
│   (pkg/metricstore) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│Storage Interface│  ← Anti-Corruption Layer (ACL)
│  (This Package) │
└────────┬────────┘
         │
    ┌────┴────┬─────────────┬──────────────┐
    ▼         ▼             ▼              ▼
┌────────┐ ┌──────┐  ┌────────────┐  ┌────────┐
│  Mock  │ │BadgerDB│ │ Prometheus │  │ Future │
└────────┘ └──────┘  └────────────┘  └────────┘
```

## Core Concepts

### Storage Interface

The `Storage` interface defines five core operations:

- **Store**: Write a single metric
- **Query**: Retrieve metrics by name and time range
- **Delete**: Remove metrics older than a specified time
- **Health**: Check storage backend health
- **Close**: Release resources

### Metric Structure

```go
type Metric struct {
    Name      string            // Metric identifier (e.g., "api_requests_total")
    Value     float64           // Numeric value
    Tags      map[string]string // Key-value labels (e.g., {"method": "GET"})
    Timestamp time.Time         // When the metric was recorded
}
```

### Error Handling

This package uses the `apprun/pkg/errors` framework with four typed error codes:

- `ErrCodeMetricsStorageNotFound`: Metric/data not found
- `ErrCodeMetricsStorageUnavailable`: Backend unavailable (closed/unhealthy)
- `ErrCodeMetricsStorageInvalidConfig`: Invalid configuration
- `ErrCodeMetricsStorageTimeout`: Operation timeout or context cancellation

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "time"
    
    "apprun/pkg/metricstore/storage"
)

func main() {
    // Create storage configuration
    cfg := storage.Config{
        Backend: "mock",
        Timeout: 5 * time.Second,
    }
    
    // Create storage instance
    store, err := storage.NewStorage(cfg)
    if err != nil {
        panic(err)
    }
    defer store.Close()
    
    // Store a metric
    ctx := context.Background()
    metric := storage.Metric{
        Name:      "api_requests_total",
        Value:     1.0,
        Tags:      map[string]string{"method": "GET", "endpoint": "/users"},
        Timestamp: time.Now(),
    }
    
    err = store.Store(ctx, metric)
    if err != nil {
        panic(err)
    }
    
    // Query metrics
    start := time.Now().Add(-1 * time.Hour)
    end := time.Now()
    metrics, err := store.Query(ctx, "api_requests_total", start, end)
    if err != nil {
        panic(err)
    }
    
    // Process metrics
    for _, m := range metrics {
        println(m.Name, m.Value, m.Timestamp)
    }
}
```

### Using Mock Storage for Testing

```go
func TestMyService(t *testing.T) {
    // Create mock storage
    cfg := storage.Config{Backend: "mock"}
    store, err := storage.NewStorage(cfg)
    require.NoError(t, err)
    defer store.Close()
    
    // Cast to MockStorage for test helpers
    mockStore := store.(*storage.MockStorage)
    
    // Test your service
    service := NewMyService(store)
    service.RecordMetric(ctx, "test_metric", 42.0)
    
    // Verify using mock helper
    assert.Equal(t, 1, mockStore.Count())
    
    // Reset for next test
    mockStore.Reset()
}
```

### Configuration

Storage configuration is loaded from `config/metrics.yaml`:

```yaml
storage:
  backend: "mock"  # "mock", "badger", "prometheus"
  timeout: 5s
  retry:
    enabled: true
    maxAttempts: 3
    backoff: "exponential"  # "exponential", "linear"
  
  # Backend-specific settings
  badger:
    path: "/var/lib/apprun/metrics"
    maxSize: 1073741824  # 1GB
    
  prometheus:
    url: "http://localhost:9090"
    pushgateway: "http://localhost:9091"
```

Environment variables override config file:

- `METRICS_STORAGE_BACKEND`: Storage backend type
- `METRICS_STORAGE_TIMEOUT`: Operation timeout
- `METRICS_STORAGE_BADGER_PATH`: BadgerDB data path

## Backends

### Mock (In-Memory)

- **Purpose**: Testing and development
- **Status**: ✅ Implemented
- **Features**: Thread-safe, no persistence, test helpers
- **Use Cases**: Unit tests, CI/CD, local development

```go
cfg := storage.Config{Backend: "mock"}
store, _ := storage.NewStorage(cfg)
```

### BadgerDB (Embedded)

- **Purpose**: MVP deployment, single-node production
- **Status**: 🚧 Planned (Story 9.3)
- **Features**: Embedded, persistent, time-series optimized
- **Use Cases**: Edge deployments, standalone instances

```go
cfg := storage.Config{
    Backend: "badger",
    Settings: map[string]interface{}{
        "path": "/var/lib/apprun/metrics",
    },
}
store, _ := storage.NewStorage(cfg)
```

### Prometheus (Enterprise)

- **Purpose**: Enterprise-scale deployments
- **Status**: 📅 Planned (Story 9.4)
- **Features**: Distributed, PromQL, Grafana integration
- **Use Cases**: Multi-node clusters, observability platforms

```go
cfg := storage.Config{
    Backend: "prometheus",
    Settings: map[string]interface{}{
        "url": "http://prometheus:9090",
    },
}
store, _ := storage.NewStorage(cfg)
```

## API Reference

### Factory Function

#### `NewStorage(cfg Config) (Storage, error)`

Creates a new storage backend based on configuration.

**Parameters:**
- `cfg`: Storage configuration (backend type, timeout, settings)

**Returns:**
- `Storage`: Storage interface implementation
- `error`: Configuration or initialization error

**Supported Backends:**
- `"mock"`: In-memory storage (always available)
- `"badger"`: BadgerDB (requires Story 9.3)
- `"prometheus"`: Prometheus (requires Story 9.4)

**Example:**
```go
cfg := storage.Config{Backend: "mock"}
store, err := storage.NewStorage(cfg)
```

### Storage Interface Methods

#### `Store(ctx context.Context, metric Metric) error`

Stores a single metric.

**Thread-Safety**: Safe for concurrent calls.

**Error Codes:**
- `ErrCodeMetricsStorageUnavailable`: Backend closed or unhealthy
- `ErrCodeMetricsStorageTimeout`: Context cancelled

**Example:**
```go
metric := storage.Metric{
    Name:      "cpu_usage",
    Value:     75.5,
    Tags:      map[string]string{"host": "web-01"},
    Timestamp: time.Now(),
}
err := store.Store(ctx, metric)
```

#### `Query(ctx context.Context, name string, start, end time.Time) ([]Metric, error)`

Retrieves metrics by name within a time range.

**Returns:** Metrics sorted by timestamp (ascending).

**Error Codes:**
- `ErrCodeMetricsStorageNotFound`: No metrics found (returns empty slice)
- `ErrCodeMetricsStorageUnavailable`: Backend closed or unhealthy
- `ErrCodeMetricsStorageTimeout`: Context cancelled

**Example:**
```go
start := time.Now().Add(-1 * time.Hour)
end := time.Now()
metrics, err := store.Query(ctx, "cpu_usage", start, end)
```

#### `Delete(ctx context.Context, before time.Time) error`

Deletes metrics older than the specified time.

**Use Case**: Implement retention policies (e.g., delete metrics older than 30 days).

**Error Codes:**
- `ErrCodeMetricsStorageUnavailable`: Backend closed or unhealthy
- `ErrCodeMetricsStorageTimeout`: Context cancelled

**Example:**
```go
cutoff := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
err := store.Delete(ctx, cutoff)
```

#### `Health(ctx context.Context) error`

Checks if the storage backend is healthy.

**Returns:**
- `nil`: Backend is healthy
- `error`: Backend is unhealthy or closed

**Example:**
```go
if err := store.Health(ctx); err != nil {
    log.Error("Storage unhealthy", err)
}
```

#### `Close() error`

Releases resources held by the storage backend.

**Behavior**: All subsequent operations will fail with `ErrCodeMetricsStorageUnavailable`.

**Example:**
```go
defer store.Close()
```

## MockStorage Test Helpers

The `MockStorage` type provides additional methods for testing:

### `Reset()`

Clears all stored metrics.

```go
mockStore.Reset()
```

### `SetHealthy(healthy bool)`

Simulates backend health status.

```go
// Simulate unhealthy backend
mockStore.SetHealthy(false)

// Restore health
mockStore.SetHealthy(true)
```

### `Count() int`

Returns the total number of stored metrics.

```go
assert.Equal(t, 5, mockStore.Count())
```

## Performance Characteristics

### Mock Storage

- **Throughput**: 1M+ ops/sec (in-memory)
- **Concurrency**: Thread-safe with read/write locks
- **Memory**: O(n) where n = number of metrics

### BadgerDB Storage (Planned)

- **Throughput**: 100K+ writes/sec
- **Latency**: <5ms p99 for writes
- **Storage**: Configurable max size with compression

### Prometheus Storage (Planned)

- **Throughput**: Depends on Prometheus cluster size
- **Latency**: Network-dependent (10-50ms typical)
- **Scalability**: Horizontal scaling via Prometheus federation

## Testing

Run all storage tests:

```bash
cd core
go test ./pkg/metricstore/storage/... -v -race -cover
```

Run with coverage report:

```bash
go test ./pkg/metricstore/storage/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Benchmark storage operations:

```bash
go test ./pkg/metricstore/storage/... -bench=. -benchmem
```

## Error Handling Examples

### Check Specific Error Codes

```go
import "apprun/pkg/errors"

metrics, err := store.Query(ctx, "my_metric", start, end)
if err != nil {
    if errors.IsCode(err, storage.ErrCodeMetricsStorageNotFound) {
        // No metrics found - handle gracefully
        return []Metric{}
    }
    // Other error - propagate
    return nil, err
}
```

### Retry on Unavailability

```go
for attempt := 0; attempt < 3; attempt++ {
    err := store.Store(ctx, metric)
    if err == nil {
        break
    }
    
    if errors.IsCode(err, storage.ErrCodeMetricsStorageUnavailable) {
        time.Sleep(time.Second * time.Duration(attempt+1))
        continue
    }
    
    // Non-retryable error
    return err
}
```

## Migration Guide

### From Direct BadgerDB Usage

**Before:**
```go
db, _ := badger.Open(badger.DefaultOptions("/tmp/metrics"))
defer db.Close()

// Write metric
db.Update(func(txn *badger.Txn) error {
    return txn.Set([]byte("key"), []byte("value"))
})
```

**After:**
```go
cfg := storage.Config{
    Backend: "badger",
    Settings: map[string]interface{}{"path": "/tmp/metrics"},
}
store, _ := storage.NewStorage(cfg)
defer store.Close()

// Write metric
metric := storage.Metric{Name: "key", Value: 1.0, Timestamp: time.Now()}
store.Store(ctx, metric)
```

**Benefits:**
- Pluggable backends (switch to Prometheus without code changes)
- Typed errors with framework integration
- Built-in retry logic via repository layer
- Consistent API across all storage types

## Design Decisions (ADR Reference)

See [ADR-009: Metrics Storage Anti-Corruption Layer](../../../specs/architecture/adr/009-metrics-storage-acl.md) for architectural decisions:

- Why interface-based design
- Backend selection criteria (MVP vs Enterprise)
- Error handling strategy
- Configuration management approach

## Contributing

When adding a new storage backend:

1. Implement the `Storage` interface
2. Add backend type to `NewStorage()` factory
3. Add configuration section to `config/metrics.yaml`
4. Write unit tests with 85%+ coverage
5. Update this README with usage examples
6. Create ADR documenting design decisions

## Related Packages

- `apprun/pkg/metricstore`: Repository layer with retry logic
- `apprun/pkg/errors`: Error code framework
- `apprun/pkg/logger`: Structured logging
- `apprun/pkg/config`: Configuration management

## Support

For questions or issues:
- Epic: [9-observability-epic.md](../../../specs/epics/9-observability-epic.md)
- Story: [9-2-metrics-storage-acl.md](../../../specs/sprint-artifacts/sprint-3/9-2-metrics-storage-acl.md)
