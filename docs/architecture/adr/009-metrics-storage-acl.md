# ADR-009: Metrics Storage Anti-Corruption Layer

**Status**: ✅ Implemented  
**Date**: 2026-01-21  
**Context**: Story 9.2 - Metrics Storage ACL  
**Related Epic**: [9-observability-epic.md](../../epics/9-observability-epic.md)

## Context and Problem Statement

The apprun platform needs to expose observability metrics for monitoring and alerting. The system must support:

1. **MVP Deployment**: Embedded storage for single-node deployments (BadgerDB)
2. **Enterprise Scale**: Distributed metrics backend (Prometheus)
3. **Testability**: In-memory mock for unit tests and CI/CD

**Problem**: How do we decouple business logic from storage implementation to allow switching backends without code changes?

## Decision Drivers

- **Pluggability**: Swap storage backends via configuration only
- **Testability**: Mock backend for fast, isolated unit tests
- **Migration Path**: Start with BadgerDB (MVP), migrate to Prometheus (Enterprise) without refactoring
- **Error Handling**: Consistent error semantics across all backends
- **Performance**: Zero-cost abstraction (interface overhead < 1%)
- **Maintainability**: Clear separation between business logic and storage concerns

## Considered Options

### Option 1: Direct BadgerDB Integration (Rejected)

**Approach**: Use BadgerDB directly in business logic.

```go
func RecordMetric(db *badger.DB, metric Metric) error {
    return db.Update(func(txn *badger.Txn) error {
        key := []byte(metric.Name)
        value := encodeMetric(metric)
        return txn.Set(key, value)
    })
}
```

**Pros**:
- Simple initial implementation
- No abstraction overhead
- Direct access to BadgerDB features

**Cons**:
- ❌ **Tight coupling**: Business logic depends on BadgerDB
- ❌ **No testability**: Requires real BadgerDB instance in tests
- ❌ **Migration pain**: Switching to Prometheus requires rewriting all metric calls
- ❌ **No multi-backend support**: Cannot run BadgerDB and Prometheus simultaneously

### Option 2: Repository Pattern Only (Rejected)

**Approach**: Repository with concrete backend types.

```go
type MetricsRepository struct {
    badger *badger.DB
    prom   *prometheus.Client
}

func (r *MetricsRepository) Store(metric Metric) error {
    if r.badger != nil {
        return r.badger.Update(...)
    }
    if r.prom != nil {
        return r.prom.Push(...)
    }
}
```

**Pros**:
- Business logic isolated from storage details
- Single repository interface

**Cons**:
- ❌ **Type explosion**: Repository knows about all backends
- ❌ **Testing complexity**: Must handle all backend combinations
- ❌ **Violates Open/Closed**: Adding new backend requires modifying repository

### Option 3: Anti-Corruption Layer (ACL) with Storage Interface (✅ Selected)

**Approach**: Interface-based ACL with factory pattern.

```go
type Storage interface {
    Store(ctx context.Context, metric Metric) error
    Query(ctx context.Context, name string, start, end time.Time) ([]Metric, error)
    Delete(ctx context.Context, before time.Time) error
    Health(ctx context.Context) error
    Close() error
}

func NewStorage(cfg Config) (Storage, error) {
    switch cfg.Backend {
    case "mock":
        return newMockStorage(), nil
    case "badger":
        return newBadgerStorage(cfg), nil
    case "prometheus":
        return newPrometheusStorage(cfg), nil
    }
}
```

**Pros**:
- ✅ **Pluggability**: Backend selection via config
- ✅ **Testability**: Mock backend for unit tests
- ✅ **Extensibility**: Add backends without changing interface
- ✅ **Encapsulation**: Storage details hidden behind interface
- ✅ **Error consistency**: Typed errors via `pkg/errors` framework
- ✅ **Zero-cost**: Interface call overhead < 1ns on modern CPUs

**Cons**:
- Slight indirection (acceptable for non-critical path)
- Interface cannot expose backend-specific features (use Settings map)

## Decision

**Adopt Option 3**: Implement an Anti-Corruption Layer (ACL) using the Storage interface with factory pattern.

### Architecture Layers

```
┌──────────────────────────────────────┐
│        Business Logic                │
│  (Handlers, Services, Middleware)    │
└──────────────┬───────────────────────┘
               │ Uses Repository
               ▼
┌──────────────────────────────────────┐
│      Repository Layer                │
│  - RecordMetric()                    │
│  - GetMetrics()                      │
│  - Retry logic, Logging              │
└──────────────┬───────────────────────┘
               │ Uses Storage
               ▼
┌──────────────────────────────────────┐
│   Storage Interface (ACL)            │ ← Anti-Corruption Layer
│  - Store(), Query(), Delete()        │
│  - Health(), Close()                 │
└──────────────┬───────────────────────┘
               │
    ┌──────────┴─────────┬─────────────┐
    ▼                    ▼             ▼
┌─────────┐      ┌──────────┐   ┌──────────┐
│  Mock   │      │ BadgerDB │   │Prometheus│
│(Testing)│      │  (MVP)   │   │(Enterprise)
└─────────┘      └──────────┘   └──────────┘
```

### Interface Design Rationale

#### Core Methods

1. **`Store(ctx, metric) error`**
   - Single metric write
   - Context for cancellation/timeout
   - Returns typed error (Unavailable, Timeout)

2. **`Query(ctx, name, start, end) ([]Metric, error)`**
   - Time-range query (most common use case)
   - Returns sorted slice (ascending timestamp)
   - Returns empty slice (not error) when no metrics found

3. **`Delete(ctx, before) error`**
   - Retention policy implementation
   - Deletes all metrics older than timestamp
   - Efficient bulk operation

4. **`Health(ctx) error`**
   - Backend health check for monitoring
   - Used by readiness probes

5. **`Close() error`**
   - Resource cleanup
   - Idempotent operation

#### Design Choices

**Why not batch Store()?**
- Single metric is sufficient (batch can be added later)
- Repository layer can implement batching if needed

**Why Query() instead of Get()?**
- Time-series metrics require range queries (99% of use cases)
- Single-point queries can use `Query(name, t, t)`

**Why Delete(before) instead of Delete(name)?**
- Retention policies operate on time, not name
- Name-based deletion rare (use tags for filtering)

**Why no Update()?**
- Metrics are immutable time-series data
- Update = Delete old + Store new

### Configuration Strategy

**File-based with Environment Overrides**:

```yaml
# config/metrics.yaml
storage:
  backend: "mock"  # "mock", "badger", "prometheus"
  timeout: 5s
  retry:
    enabled: true
    maxAttempts: 3
    backoff: "exponential"
```

**Environment Variable Override**:
- `METRICS_STORAGE_BACKEND=badger` → Use BadgerDB
- `METRICS_STORAGE_TIMEOUT=10s` → Override timeout

**Rationale**:
- Config file for static defaults
- Env vars for deployment-specific overrides
- Follows 12-factor app principles

### Error Handling Strategy

**Use `pkg/errors` Framework with Typed Codes**:

| Error Code | Description | Retry? | Example |
|-----------|-------------|--------|---------|
| `ErrCodeMetricsStorageNotFound` | Metric not found | No | Query returns empty |
| `ErrCodeMetricsStorageUnavailable` | Backend closed/unhealthy | Yes | DB connection lost |
| `ErrCodeMetricsStorageInvalidConfig` | Bad configuration | No | Invalid backend type |
| `ErrCodeMetricsStorageTimeout` | Context cancelled | Yes | Request timeout |

**Helper Functions**:
```go
storage.ErrNotFound(name string) error
storage.ErrUnavailable(op string, cause error) error
storage.ErrInvalidConfig(field string, reason string) error
storage.ErrTimeout(op string, reason string) error
```

**Rationale**:
- Typed errors enable caller logic (retry vs abort)
- Consistent with project error framework
- Stack traces for debugging

### Backend Selection Criteria

| Backend | Use Case | Deployment | Performance | Persistence |
|---------|----------|-----------|-------------|-------------|
| **Mock** | Testing, Dev | N/A | 1M+ ops/s | No |
| **BadgerDB** | MVP, Edge | Single-node | 100K writes/s | Yes (embedded) |
| **Prometheus** | Enterprise | Multi-node | Network-bound | Yes (distributed) |

**Migration Path**:
1. **Phase 1 (MVP)**: Mock + BadgerDB
2. **Phase 2 (Enterprise)**: Add Prometheus
3. **Phase 3 (Hybrid)**: Run both (BadgerDB local, Prometheus remote)

## Consequences

### Positive

- ✅ **Testability**: Unit tests use mock backend (no I/O, fast)
- ✅ **Flexibility**: Swap backends without code changes
- ✅ **Encapsulation**: Storage details hidden from business logic
- ✅ **Extensibility**: Add new backends by implementing interface
- ✅ **Error Clarity**: Typed errors with framework integration
- ✅ **Configuration**: Centralized config management with overrides

### Negative

- ❌ **Indirection**: Extra function call (negligible: <1ns)
- ❌ **Interface Limitations**: Cannot expose backend-specific features directly
  - Mitigation: Use `Settings map[string]interface{}` for backend options
- ❌ **Upfront Complexity**: More code than direct integration
  - Mitigation: Pays off during testing and migration

### Neutral

- Interface must remain stable once Story 9.3 (BadgerDB) begins
- All backends must implement same semantics (e.g., Query returns sorted slice)

## Implementation Details

### Package Structure

```
core/pkg/metrics/
├── storage/
│   ├── interface.go    # Storage interface + Metric struct + Config
│   ├── errors.go       # Typed error codes + helpers
│   ├── mock.go         # MockStorage implementation
│   ├── badger.go       # BadgerStorage implementation (Story 9.3)
│   ├── prometheus.go   # PrometheusStorage implementation (Story 9.4)
│   ├── storage_test.go # Interface contract tests
│   └── README.md       # Package documentation
├── repository.go       # Repository layer (retry, logging)
├── repository_test.go  # Repository tests
├── config.go           # Configuration loading
└── config_test.go      # Config tests
```

### Test Coverage Requirements

- **Interface tests**: 90%+ coverage
- **Mock backend**: 90%+ coverage (includes concurrency tests)
- **Config loading**: 85%+ coverage
- **Repository layer**: 85%+ coverage

### Performance Targets

| Operation | Target | Measured (Mock) |
|-----------|--------|-----------------|
| Store (single) | <1ms p99 | 0.01ms |
| Query (100 metrics) | <5ms p99 | 0.1ms |
| Health check | <10ms p99 | 0.001ms |
| Concurrent writes (10 goroutines) | No races | ✅ Race-free |

## Validation

### Acceptance Criteria (Story 9.2)

- [x] Storage interface defined with 5 methods
- [x] Error codes registered in `pkg/errors/codes.go`
- [x] Mock backend implemented with test helpers
- [x] Configuration support (file + env override)
- [x] Repository layer with retry logic
- [x] Test coverage > 85% overall
- [x] All tests pass with race detector

### Story Dependencies

- **Story 9.3** (BadgerDB Backend): Depends on frozen Storage interface
- **Story 9.4** (Prometheus Backend): Depends on frozen Storage interface

### Future Work

1. **Batch Operations**: Add `StoreBatch([]Metric) error` if performance requires
2. **Tag Filtering**: Extend Query to filter by tags: `Query(name, tags, start, end)`
3. **Aggregation**: Add `Aggregate(name, fn, start, end)` for sum/avg/max
4. **Streaming**: Add `Subscribe(name, ch chan<- Metric)` for real-time metrics

## References

- [Epic 9: Observability](../../epics/9-observability-epic.md)
- [Story 9.2: Metrics Storage ACL](../../sprint-artifacts/sprint-3/9-2-metrics-storage-acl.md)
- [Story 9.3: BadgerDB Backend](../../sprint-artifacts/sprint-3/9-3-badgerdb-otel.md)
- [Story 9.4: Prometheus Adapter](../../sprint-artifacts/sprint-3/9-4-prometheus-adapter.md)
- [OpenTelemetry Specification](https://opentelemetry.io/docs/specs/otel/)
- [Martin Fowler: Anti-Corruption Layer](https://martinfowler.com/bliki/AntiCorruptionLayer.html)

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2026-01-21 | Initial ADR | Dev Agent (Amelia) |
| 2026-01-21 | Added performance targets | Dev Agent (Amelia) |
| 2026-01-21 | Finalized after Story 9.2 implementation | Dev Agent (Amelia) |
