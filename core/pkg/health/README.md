# Health Package

Reusable health check library for monitoring system and component health.

## Features

- Component-based architecture with extensible checker interface
- Parallel execution with goroutines
- Timeout control (per-component 2s, overall 5s)
- Critical vs non-critical component classification
- Automatic status aggregation (healthy/degraded/unhealthy)

## Usage

```go
import "apprun/pkg/health"
import "apprun/pkg/health/checkers"

// Create checkers
checker := health.NewHealthChecker("apprun", "1.0.0",
    checkers.NewDatabaseChecker(dbClient),
    checkers.NewRedisChecker(cacheClient),
    checkers.NewMetricsChecker(true),
)

// Check health
report := checker.CheckAll(ctx)
```

## Status Logic

- **healthy**: All components operational
- **degraded**: Non-critical component failed (returns HTTP 207)
- **unhealthy**: Critical component failed (returns HTTP 503)

**Critical components**: database  
**Non-critical**: cache, metrics_storage

---
**Documentation**: See [Story 1-20](../../../specs/sprint-artifacts/sprint-3/1-20-health-check-api.md)
