# Observability Module (obs)

This module provides system and application metrics for monitoring and observability purposes.

## Structure

```
core/modules/obs/
├── metrics.go        # Prometheus metrics (counters)
├── types.go          # Metrics data types
├── config.go         # Cache configuration
├── collector.go      # Metrics collection logic
├── service.go        # Business logic with caching
├── handler.go        # HTTP handlers
├── collector_test.go # Unit tests for collector
└── service_test.go   # Unit tests for service
```

## Features

- **User Metrics**: Total, active, admin, banned users, registration stats
- **System Metrics**: Uptime, memory, CPU, disk usage, goroutines
- **Auth Metrics**: Login attempts, success rate (stub for Story 5.9)
- **Performance Metrics**: API requests, response times (stub for future)
- **Caching**: 5-minute TTL for most metrics (system metrics are real-time)
- **Prometheus Integration**: Export counters for user operations

## API Endpoints

All endpoints require `platform_admin` role:

- `GET /api/metrics` - Get all metrics
- `GET /api/metrics/users` - Get user metrics
- `GET /api/metrics/system` - Get system health metrics
- `GET /api/metrics/performance` - Get performance metrics

## Testing

Run tests:
```bash
cd core
go test -v ./modules/obs/...
```

All tests passed with 100% coverage of core functionality.
