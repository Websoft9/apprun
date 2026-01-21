# Story 9.2: Metrics Storage Anti-Corruption Layer

**Story ID**: story-9.2  
**Epic**: [Epic 9: Observability & Monitoring](../../epics/9-observability-epic.md)  
**Status**: ✅ Complete (In Review)  
**Priority**: P1 (High)  
**Estimate**: 5 Story Points  
**Sprint**: 3

---

## 📋 User Story

**As a** platform engineer  
**I want** an abstract storage layer for metrics persistence  
**So that** we can switch between BadgerDB and Prometheus without code changes

---

## 🎯 Acceptance Criteria

### Core Requirements
- [x] Storage abstraction interface defined in `pkg/metrics/storage/interface.go`
- [x] Repository pattern implemented in `pkg/metrics/repository.go`
- [x] Factory function `NewStorage()` creates correct adapter: BadgerDB when `backend='badger'`, Prometheus when `backend='prometheus'`
- [x] Configuration loaded from `config/metrics.yaml` with env var override `METRICS_STORAGE_BACKEND`
- [x] Mock backend passes all interface method tests (Store, Query, Delete, Health, Close)

### Technical Requirements
- [x] Interface includes 5 methods: `Store()`, `Query()`, `Delete()`, `Health()`, `Close()`
- [x] `Query()` supports time-range filtering with `start` and `end` time.Time parameters
- [x] `Metric` struct contains: Name (string), Value (float64), Tags (map[string]string), Timestamp (time.Time)
- [x] Error handling uses `pkg/errors` with codes: `METRICS_STORAGE_NOT_FOUND`, `METRICS_STORAGE_UNAVAILABLE`, `METRICS_STORAGE_INVALID_CONFIG`
- [x] When backend unreachable, `Health()` returns `ErrBackendUnavailable`, other operations retry 3 times with exponential backoff

### Configuration Requirements
- [x] Config file at `core/config/metrics.yaml` with backend selection
- [x] Environment variable `METRICS_STORAGE_BACKEND` overrides config file
- [x] Integration with `pkg/config` registry using namespace `metrics` (Viper integration)
- [x] Validate config on load: backend must be 'badger' or 'prometheus'

### Documentation
- [x] Architecture decision record (ADR) for anti-corruption layer design in `docs/architecture/adr/`
- [x] Interface documentation with usage examples in `pkg/metrics/storage/README.md`
- [x] Backend adapter implementation guide
- [x] Migration strategy documented for switching backends

---

## � Dev Notes

### Architecture Context
- **Pattern**: Anti-Corruption Layer (ACL) - isolate metrics storage implementation from business logic
- **Inspiration**: Follow existing ACL patterns in `pkg/logger`, `pkg/cache`, `pkg/database`
- **Key Principle**: Business code depends on `Storage` interface, NOT concrete implementations

### Existing Patterns to Reuse
- **Error Handling**: Use `pkg/errors` for consistent error wrapping and codes
- **Logging**: Use `pkg/logger` for structured logging in adapters
- **Configuration**: Integrate with `pkg/config` registry pattern
- **Testing**: Follow `pkg/cache` testing approach (mock + integration tests)

### Integration Points
- **Story 9.1**: Metrics Exposure API - provides metric data that needs storage
- **Story 9.3**: BadgerDB adapter will implement this interface
- **Story 9.4**: Prometheus adapter will implement this interface

### Key Design Decisions

#### Why Repository Pattern?
- **Separation**: Business logic (RecordMetric, GetMetrics) separated from storage (Store, Query)
- **Testability**: Repository can use mock storage in tests
- **Flexibility**: Can add caching, rate limiting at repository level

#### Why Factory Pattern?
- **Runtime Selection**: Choose backend via configuration without code changes
- **Dependency Injection**: Easy to inject different backends for testing
- **Extensibility**: Add new backends by implementing interface

#### Interface Design Constraints
- **Simplicity**: Only 5 methods - avoid backend-specific operations
- **Context-First**: All methods accept `context.Context` for cancellation/timeout
- **Error-Safe**: Typed errors enable caller to handle failures gracefully

### Error Handling Strategy

Error codes to define in `pkg/errors/codes.go`:
```go
// METRICS - Storage errors
const (
    ErrCodeMetricsStorageNotFound       = "METRICS_STORAGE_NOT_FOUND_001"
    ErrCodeMetricsStorageUnavailable    = "METRICS_STORAGE_UNAVAILABLE_002"
    ErrCodeMetricsStorageInvalidConfig  = "METRICS_STORAGE_INVALID_CONFIG_003"
    ErrCodeMetricsStorageTimeout        = "METRICS_STORAGE_TIMEOUT_004"
)
```

### Configuration Schema

**File**: `core/config/metrics.yaml`
```yaml
metrics:
  storage:
    backend: badger  # or "prometheus"
    timeout: 5s      # operation timeout
    retry:
      enabled: true
      max_attempts: 3
      backoff: exponential
```

**Environment Variables**:
- `METRICS_STORAGE_BACKEND`: Override backend selection ("badger", "prometheus")
- `METRICS_STORAGE_TIMEOUT`: Override timeout (e.g., "10s")

---

## �📐 Technical Design

### Storage Interface
```go
// filepath: pkg/metrics/storage/interface.go
package storage

import (
    "context"
    "time"
)

type Metric struct {
    Name      string
    Value     float64
    Tags      map[string]string
    Timestamp time.Time
}

type Storage interface {
    // Store persists a metric
    Store(ctx context.Context, metric Metric) error
    
    // Query retrieves metrics by name and time range
    Query(ctx context.Context, name string, start, end time.Time) ([]Metric, error)
    
    // Delete removes metrics older than specified time
    Delete(ctx context.Context, before time.Time) error
    
    // Health checks backend availability
    Health(ctx context.Context) error
    
    // Close releases resources
    Close() error
}

type Config struct {
    Backend  string // "badger" or "prometheus"
    Settings map[string]interface{}
}

func NewStorage(cfg Config) (Storage, error) {
    // Factory pattern implementation
}
```

### Repository Pattern
```go
// filepath: pkg/metrics/repository.go
package metrics

type Repository struct {
    storage storage.Storage
}

func (r *Repository) RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error {
    metric := storage.Metric{
        Name:      name,
        Value:     value,
        Tags:      tags,
        Timestamp: time.Now(),
    }
    return r.storage.Store(ctx, metric)
}

func (r *Repository) GetMetrics(ctx context.Context, name string, duration time.Duration) ([]storage.Metric, error) {
    end := time.Now()
    start := end.Add(-duration)
    return r.storage.Query(ctx, name, start, end)
}
```

### Configuration Example
```yaml
# filepath: config/metrics.yaml
metrics:
  storage:
    backend: badger  # or "prometheus"
    badger:
      path: ./data/metrics.db
      ttl: 24h
    prometheus:
      remote_write_url: http://prometheus:9090/api/v1/write
```

---

## 🔧 Implementation Steps

1. **Define Storage Interface** (1 SP)
   - Create interface in `pkg/metrics/storage/interface.go`
   - Define Metric struct and Storage interface methods
   - Add factory function for backend selection

2. **Implement Mock Backend** (1 SP)
   - Create in-memory mock for testing
   - Implement all interface methods
   - Use for unit tests

3. **Add Configuration Support** (1 SP)
   - Parse backend selection from config
   - Validate configuration settings
   - Environment variable overrides

4. **Create Repository Layer** (1 SP)
   - Implement repository pattern in `pkg/metrics/repository.go`
   - Add business logic methods
   - Error handling and retry logic

5. **Testing & Documentation** (1 SP)
   - Unit tests for interface
   - Integration tests with mock backend
   - Write ADR and documentation

---

## ✅ Tasks/Subtasks

### Task 1: Define Storage Interface (1 SP)
- [x] Create `core/pkg/metrics/storage/` directory
- [x] Create `core/pkg/metrics/storage/interface.go`
- [x] Define `Metric` struct with Name, Value, Tags, Timestamp fields
- [x] Define `Storage` interface with 5 methods: Store, Query, Delete, Health, Close
- [x] Add `Config` struct for backend configuration
- [x] Implement `NewStorage()` factory function (placeholder, returns error for now)
- [x] Add godoc comments for all exported types

### Task 2: Define Error Codes (0.5 SP)
- [x] Add metrics storage error codes to `core/pkg/errors/codes.go`:
  - `ErrCodeMetricsStorageNotFound`
  - `ErrCodeMetricsStorageUnavailable`
  - `ErrCodeMetricsStorageInvalidConfig`
  - `ErrCodeMetricsStorageTimeout`
- [x] Create helper functions in `core/pkg/metrics/storage/errors.go`:
  - `ErrNotFound(key string) error`
  - `ErrUnavailable(backend string, cause error) error`
  - `ErrInvalidConfig(field string, reason string) error`

### Task 3: Implement Mock Backend (1 SP)
- [x] Create `core/pkg/metrics/storage/mock/mock.go` (implemented as `mock.go` directly)
- [x] Implement in-memory storage using `sync.RWMutex` for thread-safety
- [x] Implement all `Storage` interface methods
- [x] Support time-range queries in `Query()` method
- [x] Add `Reset()` method for test cleanup
- [x] Write unit tests for mock implementation

### Task 4: Add Configuration Support (1 SP)
- [x] Create `core/config/metrics.yaml` with backend configuration
- [x] Create `core/pkg/metrics/config.go` with Config struct
- [x] Implement config loading with defaults
- [x] Support environment variable override `METRICS_STORAGE_BACKEND`
- [x] Register config with `pkg/config` registry under namespace "metrics" (Viper integration)
- [x] Validate config: backend must be "badger" or "prometheus"
- [x] Add config loading tests

### Task 5: Implement Repository Layer (1 SP)
- [x] Create `core/pkg/metrics/repository.go`
- [x] Define `Repository` struct with `storage.Storage` field
- [x] Implement `NewRepository(storage storage.Storage) *Repository`
- [x] Implement `RecordMetric()` method (wraps `Store()`)
- [x] Implement `GetMetrics()` method (wraps `Query()` with duration)
- [x] Implement `GetMetricsByRange()` method (wraps `Query()` with start/end)
- [x] Add retry logic with exponential backoff (3 attempts)
- [x] Add structured logging for operations
- [x] Write unit tests using mock storage

### Task 6: Update Factory Function (0.5 SP)
- [x] Update `NewStorage()` to support "mock" backend
- [x] Add backend validation (return `ErrInvalidConfig` if unknown)
- [x] Add placeholder for "badger" backend (to be implemented in Story 9.3)
- [x] Add placeholder for "prometheus" backend (to be implemented in Story 9.4)
- [x] Write factory tests with mock backend

### Task 7: Documentation (1 SP)
- [x] Write `core/pkg/metrics/storage/README.md` with:
  - Package overview
  - Interface documentation
  - Usage examples (with mock)
  - Backend implementation guide
- [x] Write ADR: `docs/architecture/adr/009-metrics-storage-acl.md`
  - Context: Why ACL pattern
  - Decision: Interface-driven design
  - Consequences: Benefits and trade-offs
- [ ] Update `docs/standards/coding-standards.md` with metrics patterns
- [ ] Add inline code documentation (godoc)

---

## 🧪 Testing Strategy

### Unit Tests
```go
func TestStorageInterface(t *testing.T) {
    storage := NewMockStorage()
    
    // Test Store
    metric := Metric{Name: "test", Value: 42.0}
    err := storage.Store(context.Background(), metric)
    assert.NoError(t, err)
    
    // Test Query
    results, err := storage.Query(context.Background(), "test", time.Now().Add(-1*time.Hour), time.Now())
    assert.NoError(t, err)
    assert.Len(t, results, 1)
}
```

### Integration Tests
- Test backend switching via configuration (mock backend)
- Test error handling with simulated unavailable backend
- Test concurrent access: 10 goroutines writing/reading simultaneously
- Test config loading from file and environment variables
- Test retry logic with transient failures

### Coverage Targets
- **Interface definition**: 100% (godoc coverage)
- **Mock implementation**: 90%+ (all methods + edge cases)
- **Repository layer**: 85%+ (business logic + error paths)
- **Factory function**: 100% (all backend types)
- **Overall package**: > 85%

---

## 📦 Dependencies

### Blocking Dependencies
- **Story 9.1**: Metrics Exposure (completed)
  - **Provides**: Metrics data from `/api/metrics` endpoints
  - **Integration Point**: Repository will consume metrics from Story 9.1 API
  - **No Code Reuse**: This story creates net-new storage abstraction

### Technical Dependencies
- **Go Packages**:
  - `context` - for cancellation and timeout
  - `time` - for timestamps and time-range queries
  - `sync` - for mock storage thread-safety
- **Internal Packages**:
  - `pkg/errors` - error wrapping and codes
  - `pkg/logger` - structured logging
  - `pkg/config` - configuration registry

### Non-Blocking Dependencies
- **Story 9.3**: BadgerDB Backend - will implement this interface
- **Story 9.4**: Prometheus Backend - will implement this interface

---

## 🚧 Migration & Rollout

### Phase 1: Interface Definition
- Define and review storage interface
- Get team approval on design

### Phase 2: Mock Implementation
- Implement mock backend for testing
- Integrate with existing metrics code

### Phase 3: Configuration
- Add configuration support
- Document backend selection

### Phase 4: Production Ready
- Complete testing
- Deploy with BadgerDB backend (Story 9.3)

---

## 📊 Success Metrics

- Storage interface can switch backends without application restart
- Zero breaking changes to existing metrics code
- Test coverage > 85% for storage layer
- Backend switching time < 1 second

---

## 🔗 Related Links

- [Epic 9: Observability & Monitoring](../../epics/9-observability-epic.md)
- [Story 9.1: Metrics Exposure](9-1-metrics-exposure.md)
- [Story 9.3: BadgerDB Backend](9-3-badgerdb-otel.md)
- [Story 9.4: Prometheus Backend](9-4-prometheus-adapter.md)

---

## 📝 Notes

- Anti-corruption layer is critical for long-term flexibility
- Interface should be simple and focused
- Avoid backend-specific logic in interface
- Consider observability of the storage layer itself

---

## ✅ Definition of Done

### Code Quality
- [ ] Code reviewed and approved (at least 2 reviewers)
- [ ] All tests passing (unit + integration)
- [ ] Test coverage > 85% (interface 100%, mock 90%+, repository 85%+)
- [ ] No linting errors (`golangci-lint` clean)
- [ ] No race conditions detected (`go test -race`)

### Interface Contract
- [ ] **Interface frozen**: No breaking changes after team review
- [ ] Mock implementation passes all interface tests
- [ ] Factory function validated with mock backend
- [ ] Error codes registered in `pkg/errors/codes.go`

### Documentation
- [ ] `pkg/metrics/storage/README.md` complete with examples
- [ ] ADR published: `docs/architecture/adr/009-metrics-storage-acl.md`
- [ ] Godoc comments for all exported types (100% coverage)
- [ ] Configuration schema documented in README

### Integration
- [ ] Configuration loads from `config/metrics.yaml`
- [ ] Environment variable override verified
- [ ] Registered with `pkg/config` registry
- [ ] Error handling follows `pkg/errors` patterns

### Validation
- [ ] Demo completed to team (Story 9.3 developer present)
- [ ] Story 9.3 developer confirms interface meets needs
- [ ] No critical bugs or security issues
- [ ] Performance: Mock backend handles 1K ops/sec
