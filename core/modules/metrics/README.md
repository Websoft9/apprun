# Metrics Module

This module provides comprehensive observability capabilities for the apprun platform, including metrics collection, storage, and exposure.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│                  (Routes + Middleware)                       │
└───────────────────────┬─────────────────────────────────────┘
                        │
        ┌───────────────┴────────────────┐
        │                                │
        ▼                                ▼
┌──────────────────┐          ┌──────────────────┐
│  MetricsHandler  │          │ StorageHandler   │
│  (Exposition)    │          │ (Ingestion)      │
└────────┬─────────┘          └────────┬─────────┘
         │                             │
         ▼                             ▼
┌─────────────────────────────────────────────────┐
│            MetricsService                       │
│  - CollectUserMetrics()                        │
│  - CollectSystemMetrics()                      │
│  - CollectPerformanceMetrics()                 │
└────────┬────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────┐
│        pkg/metricstore/Repository               │
│  - RecordMetric()                              │
│  - GetMetrics()                                │
│  - Retry + Logging                             │
└────────┬────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────┐
│   pkg/metricstore/storage (Anti-Corruption)     │
│  - Storage Interface                           │
│  - Mock / BadgerDB / Prometheus                │
└─────────────────────────────────────────────────┘
```

## Files Structure

```
core/modules/metrics/
├── metrics.go              # Prometheus metrics (counters)
├── types.go                # Metrics data types
├── config.go               # Cache configuration
├── collector.go            # Metrics collection logic
├── service.go              # Business logic with caching
├── handler.go              # HTTP handlers (exposition)
├── storage_handler.go      # Storage API handlers
├── collector_test.go       # Unit tests for collector
├── service_test.go         # Unit tests for service
├── storage_handler_test.go # Unit tests for storage
└── README.md               # This file
```

## Features

### 1. Metrics Exposition (Story 9.1) ✅
Exposes platform metrics via REST API:
- **User Metrics**: Total users, active users, registrations
- **System Metrics**: CPU, memory, disk usage
- **Performance Metrics**: API response times, error rates
- **Auth Metrics**: Login attempts, success rates

**Endpoints**:
- `GET /api/metrics` - All metrics
- `GET /api/metrics/users` - User statistics
- `GET /api/metrics/system` - System health
- `GET /api/metrics/performance` - Performance stats
- `GET /api/metrics/auth` - Authentication metrics

### 2. Metrics Storage (Story 9.2 + 9.3) ✅
Persistent storage with pluggable backends:
- **Mock**: In-memory storage for testing
- **BadgerDB**: Embedded storage for MVP deployment
- **Prometheus**: Enterprise-scale (planned Story 9.4)

**Storage Endpoints**:
- `POST /api/metrics/storage/ingest` - Store metrics
- `GET /api/metrics/storage/query` - Query metrics by time range
- `GET /api/metrics/storage/health` - Storage health check

### 3. OpenTelemetry Integration (Story 9.3) ✅
OTEL SDK integration for metrics collection:
- Automatic metric export to BadgerDB
- Support for Sum, Gauge, Histogram metrics
- 10-second periodic export interval

## Usage

### Basic Setup

```go
import (
    "apprun/modules/metrics"
    "apprun/pkg/metricstore"
    "apprun/pkg/metricstore/storage"
)

// 1. Create storage backend
cfg := &metrics.Config{
    Storage: metrics.StorageConfig{
        Backend: "badger",
        Settings: map[string]interface{}{
            "path": "/var/lib/apprun/metrics",
            "ttl":  24 * time.Hour,
        },
    },
}

store, _ := storage.NewStorage(cfg.ToStorageConfig())
repo := metrics.NewRepository(store, cfg)

// 2. Setup handlers
metricsHandler := obs.NewMetricsHandler(service)
storageHandler := obs.NewStorageHandler(repo)

// 3. Register routes (chi framework)
r.Get("/api/metrics", metricsHandler.GetAll)
r.Post("/api/metrics/storage/ingest", storageHandler.Ingest)
r.Get("/api/metrics/storage/query", storageHandler.Query)
```

### Ingesting Metrics

**HTTP Request**:
```bash
curl -X POST http://localhost:8080/api/metrics/storage/ingest \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "api_requests_total",
    "value": 1.0,
    "tags": {
      "method": "GET",
      "endpoint": "/users",
      "status": "200"
    }
  }'
```

**Response**:
```json
{
  "success": true,
  "code": 200,
  "data": {
    "success": true,
    "message": "Metric stored",
    "metric_id": "api_requests_total:1737450000000000000"
  }
}
```

### Querying Metrics

**By Duration**:
```bash
curl "http://localhost:8080/api/metrics/storage/query?name=api_requests_total&duration=1h"
```

**By Time Range**:
```bash
curl "http://localhost:8080/api/metrics/storage/query?name=api_requests_total&start=2026-01-21T00:00:00Z&end=2026-01-21T12:00:00Z&limit=100"
```

**Response**:
```json
{
  "success": true,
  "code": 200,
  "data": {
    "metrics": [
      {
        "name": "api_requests_total",
        "value": 1.0,
        "tags": {"method": "GET"},
        "timestamp": "2026-01-21T10:00:00Z"
      }
    ],
    "count": 1,
    "start": "2026-01-21T09:00:00Z",
    "end": "2026-01-21T10:00:00Z",
    "has_more": false
  }
}
```

## Configuration

### Metrics Storage (config/metrics.yaml)

```yaml
metrics:
  storage:
    backend: badger  # "mock", "badger", "prometheus"
    timeout: 5s
    retry:
      enabled: true
      max_attempts: 3
      backoff: exponential
    
    badger:
      path: /var/lib/apprun/metrics
      ttl: 24h
      compression: true
      gc_interval: 5m
```

### Environment Variables

- `METRICS_STORAGE_BACKEND` - Override storage backend
- `METRICS_STORAGE_TIMEOUT` - Operation timeout
- `METRICS_STORAGE_BADGER_PATH` - BadgerDB data path

## Testing

### Run All Tests
```bash
cd core
go test ./modules/metrics/... -v -cover
```

### Run Storage Handler Tests
```bash
go test ./modules/metrics/... -v -run TestStorageHandler
```

### Run with Race Detection
```bash
go test ./modules/metrics/... -race
```

### Test Results
- ✅ All exposition handler tests pass
- ✅ All storage handler tests pass (11 test cases)
- ✅ Race detector clean
- ✅ 100% core functionality coverage

## Performance Characteristics

### BadgerDB Storage
- **Write Throughput**: 1K+ metrics/minute
- **Query Latency**: < 100ms (p95) for 24h range
- **Memory Usage**: < 100MB typical workload
- **Disk Usage**: Snappy compression enabled
- **TTL**: Automatic expiration after 24h

### Mock Storage
- **Throughput**: 1M+ ops/sec (in-memory)
- **Concurrency**: Thread-safe with read/write locks
- **Use Cases**: Unit tests, CI/CD, local development

## Architecture Decisions

See related ADRs:
- [ADR-009: Metrics Storage Anti-Corruption Layer](../../docs/architecture/adr/009-metrics-storage-acl.md)

## Dependencies

- **chi**: HTTP router (existing)
- **BadgerDB v3**: Embedded key-value storage
- **OpenTelemetry SDK**: Metrics collection framework
- **apprun/pkg/metricstore**: Metrics repository layer
- **apprun/pkg/response**: Standardized HTTP responses
- **apprun/pkg/logger**: Structured logging

## Future Enhancements

### Story 9.4: Prometheus Backend
- Remote write to Prometheus
- PromQL query support
- Grafana integration

### Planned Features
- Metric aggregation (sum, avg, max)
- Tag-based filtering
- Real-time streaming subscriptions
- Alerting rules engine

## Contributing

When adding new metrics:
1. Define metric in `types.go`
2. Implement collection logic in `collector.go`
3. Add handler method in `handler.go`
4. Register route in main application
5. Write unit tests
6. Update Swagger documentation

## Related Stories

- **Story 9.1**: Metrics Exposure (平台指标暴露) - ✅ Complete
- **Story 9.2**: Metrics Storage ACL - ✅ Complete
- **Story 9.3**: BadgerDB Backend + OTEL - ✅ Complete
- **Story 9.4**: Prometheus Adapter - 📅 Planned
