# Story 9.3: BadgerDB Backend with OTEL Integration

**Story ID**: story-9.3  
**Epic**: [Epic 9: Observability & Monitoring](../../epics/9-observability-epic.md)  
**Status**: ✅ Complete  
**Priority**: P1 (High)  
**Estimate**: 8 Story Points  
**Sprint**: Sprint 3  
**Completed**: 2026-01-21

---

## 📋 User Story

**As a** platform engineer  
**I want** an embedded metrics storage using BadgerDB with OTEL integration  
**So that** we can deploy observability without external dependencies in MVP

---

## 🎯 Acceptance Criteria

### Core Requirements
- [x] BadgerDB storage adapter implements storage interface from Story 9.2
- [x] OTEL receiver/exporter for metrics collection
- [x] 24-hour TTL with automatic expiration
- [x] Key-value storage schema supports < 100ms query for 24h range
- [x] Query API supporting time-range retrieval with pagination

### Technical Requirements
- [x] BadgerDB initialization with proper configuration (Snappy compression, WAL enabled)
- [x] OTEL SDK integration for metrics collection
- [x] Custom OTEL exporter writing to BadgerDB
- [x] Time-based key encoding (timestamp prefix for efficient range queries)
- [x] Concurrent access test: 10+ goroutines write 1K metrics without race conditions
- [x] Graceful shutdown: all in-flight writes complete within 5s, zero data loss

### Performance
- [x] Support 1K+ metrics/minute write throughput
- [x] Query response time < 100ms for 24h range
- [x] Memory usage < 100MB for typical workload
- [x] Disk usage with compression enabled

### API Endpoints
- [x] `POST /api/metrics/storage/ingest` - Store metrics
- [x] `GET /api/metrics/storage/query` - Query by name and time range
- [x] `GET /api/metrics/storage/health` - Storage health check

### Security Requirements
- [x] API endpoints require authentication (admin role, reuse Story 9.1 auth)
- [x] Rate limiting: max 1000 requests/min per client
- [x] Input validation: metric name max 256 chars, tags max 10 keys
- [x] Sanitize inputs to prevent injection attacks

### Observability Requirements
- [x] Expose BadgerDB internal metrics:
  - `badger_disk_usage_bytes` - Current disk usage
  - `badger_write_latency_ms` - Write operation latency
  - `badger_query_latency_ms` - Query operation latency
  - `badger_gc_duration_ms` - Garbage collection duration
  - `badger_error_total` - Error count by type
- [x] Integrate storage metrics with existing `/api/metrics` endpoint

---

## 📐 Technical Design

### API Specification

#### POST /api/metrics/storage/ingest
**Authentication**: Required (admin role)

**Request**:
```json
{
  "name": "http_requests_total",
  "value": 42.0,
  "tags": {
    "method": "GET",
    "endpoint": "/api/users",
    "status": "200"
  },
  "timestamp": "2026-01-21T10:00:00Z"
}
```

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "Metric stored",
  "metric_id": "http_requests_total:1737456000000000000"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid metric format
- `401 Unauthorized`: Missing or invalid authentication
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Storage failure

#### GET /api/metrics/storage/query
**Authentication**: Required (admin role)

**Query Parameters**:
- `name` (required): Metric name
- `duration` (optional): Time range (e.g., "1h", "24h"), default: "24h"
- `start` (optional): Start timestamp (RFC3339)
- `end` (optional): End timestamp (RFC3339)
- `limit` (optional): Max results, default: 1000

**Example**:
```
GET /api/metrics/storage/query?name=http_requests_total&duration=1h&limit=100
```

**Response (200 OK)**:
```json
{
  "metrics": [
    {
      "name": "http_requests_total",
      "value": 42.0,
      "tags": {"method": "GET"},
      "timestamp": "2026-01-21T10:00:00Z"
    }
  ],
  "count": 100,
  "start": "2026-01-21T09:00:00Z",
  "end": "2026-01-21T10:00:00Z",
  "has_more": false
}
```

#### GET /api/metrics/storage/health
**Authentication**: Optional (public or admin)

**Response (200 OK)**:
```json
{
  "status": "healthy",
  "backend": "badger",
  "disk_usage_mb": 45.2,
  "metrics_count": 125000,
  "uptime_seconds": 86400
}
```

**Response (503 Service Unavailable)**:
```json
{
  "status": "unhealthy",
  "error": "BadgerDB connection failed"
}
```

### BadgerDB Adapter Implementation
```go
// filepath: pkg/metricstore/storage/badger/adapter.go
package badger

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/dgraph-io/badger/v3"
    "apprun/pkg/metricstore/storage"
)

type BadgerStorage struct {
    db     *badger.DB
    ttl    time.Duration
    config Config
}

type Config struct {
    Path        string
    TTL         time.Duration
    Compression bool
}

func NewBadgerStorage(cfg Config) (*BadgerStorage, error) {
    opts := badger.DefaultOptions(cfg.Path).
        WithCompression(badger.ZSTD).
        WithLoggingLevel(badger.WARNING)
    
    db, err := badger.Open(opts)
    if err != nil {
        return nil, fmt.Errorf("failed to open badger: %w", err)
    }
    
    storage := &BadgerStorage{
        db:     db,
        ttl:    cfg.TTL,
        config: cfg,
    }
    
    // Start TTL cleanup goroutine
    go storage.runGC()
    
    return storage, nil
}

func (s *BadgerStorage) Store(ctx context.Context, metric storage.Metric) error {
    // Key format: {metric_name}:{timestamp_ns}
    key := fmt.Sprintf("%s:%d", metric.Name, metric.Timestamp.UnixNano())
    
    value, err := json.Marshal(metric)
    if err != nil {
        return fmt.Errorf("failed to marshal metric: %w", err)
    }
    
    return s.db.Update(func(txn *badger.Txn) error {
        e := badger.NewEntry([]byte(key), value).WithTTL(s.ttl)
        return txn.SetEntry(e)
    })
}

func (s *BadgerStorage) Query(ctx context.Context, name string, start, end time.Time) ([]storage.Metric, error) {
    var metrics []storage.Metric
    
    err := s.db.View(func(txn *badger.Txn) error {
        opts := badger.DefaultIteratorOptions
        opts.Prefix = []byte(name + ":")
        
        it := txn.NewIterator(opts)
        defer it.Close()
        
        startKey := fmt.Sprintf("%s:%d", name, start.UnixNano())
        endKey := fmt.Sprintf("%s:%d", name, end.UnixNano())
        
        for it.Seek([]byte(startKey)); it.Valid(); it.Next() {
            item := it.Item()
            key := string(item.Key())
            
            if key > endKey {
                break
            }
            
            err := item.Value(func(val []byte) error {
                var metric storage.Metric
                if err := json.Unmarshal(val, &metric); err != nil {
                    return err
                }
                metrics = append(metrics, metric)
                return nil
            })
            
            if err != nil {
                return err
            }
        }
        
        return nil
    })
    
    return metrics, err
}

func (s *BadgerStorage) Delete(ctx context.Context, before time.Time) error {
    // TTL handles automatic deletion, but can implement manual cleanup
    return nil
}

func (s *BadgerStorage) Health(ctx context.Context) error {
    // Try a simple read operation
    return s.db.View(func(txn *badger.Txn) error {
        return nil
    })
}

func (s *BadgerStorage) Close() error {
    return s.db.Close()
}

func (s *BadgerStorage) runGC() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        s.db.RunValueLogGC(0.5)
    }
}
```

### OTEL Integration
```go
// filepath: pkg/metricstore/otel/exporter.go
package otel

import (
    "context"
    
    "go.opentelemetry.io/otel/sdk/metric"
    "apprun/pkg/metricstore/storage"
)

type BadgerExporter struct {
    storage storage.Storage
}

func NewBadgerExporter(storage storage.Storage) (*BadgerExporter, error) {
    return &BadgerExporter{storage: storage}, nil
}

func (e *BadgerExporter) Export(ctx context.Context, data metric.ResourceMetrics) error {
    // Convert OTEL metrics to storage format
    for _, scopeMetrics := range data.ScopeMetrics {
        for _, m := range scopeMetrics.Metrics {
            metric := convertOTELMetric(m)
            if err := e.storage.Store(ctx, metric); err != nil {
                return err
            }
        }
    }
    return nil
}

func (e *BadgerExporter) Shutdown(ctx context.Context) error {
    return e.storage.Close()
}

// OTEL SDK Setup
func SetupOTEL(storage storage.Storage) (*metric.MeterProvider, error) {
    exporter, err := NewBadgerExporter(storage)
    if err != nil {
        return nil, err
    }
    
    reader := metric.NewPeriodicReader(exporter,
        metric.WithInterval(10*time.Second),
    )
    
    provider := metric.NewMeterProvider(
        metric.WithReader(reader),
    )
    
    otel.SetMeterProvider(provider)
    
    return provider, nil
}
```

### API Handlers
```go
// filepath: core/handlers/observability/metrics_handler.go
package observability

import (
    "encoding/json"
    "net/http"
    "time"
    
    "apprun/pkg/metricstore/storage"
)

type MetricsHandler struct {
    storage storage.Storage
}

func (h *MetricsHandler) Query(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    durationStr := r.URL.Query().Get("duration")
    
    duration, _ := time.ParseDuration(durationStr)
    if duration == 0 {
        duration = 24 * time.Hour
    }
    
    end := time.Now()
    start := end.Add(-duration)
    
    metrics, err := h.storage.Query(r.Context(), name, start, end)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(metrics)
}

func (h *MetricsHandler) Health(w http.ResponseWriter, r *http.Request) {
    if err := h.storage.Health(r.Context()); err != nil {
        http.Error(w, "unhealthy", http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
```

---

## 🔧 Implementation Steps

1. **BadgerDB Adapter** (3 SP)
   - Implement storage interface
   - Key encoding strategy
   - TTL and GC handling
   - Unit tests

2. **OTEL Integration** (2 SP)
   - Custom OTEL exporter
   - SDK setup and configuration
   - Metric conversion logic
   - Integration tests

3. **API Endpoints** (2 SP)
   - Query handler
   - Health check handler
   - Input validation
   - Error handling

4. **Testing & Optimization** (1 SP)
   - Performance testing
   - Load testing (1K+ metrics/min)
   - Memory profiling
   - Documentation

---

## 🧪 Testing Strategy

### Unit Tests
```go
func TestBadgerStorage_Store(t *testing.T) {
    storage := setupTestStorage(t)
    defer storage.Close()
    
    metric := storage.Metric{
        Name:      "test_counter",
        Value:     42.0,
        Tags:      map[string]string{"env": "test"},
        Timestamp: time.Now(),
    }
    
    err := storage.Store(context.Background(), metric)
    assert.NoError(t, err)
}

func TestBadgerStorage_Query(t *testing.T) {
    storage := setupTestStorage(t)
    defer storage.Close()
    
    // Store test data
    storeTestMetrics(storage, 100)
    
    // Query
    start := time.Now().Add(-1 * time.Hour)
    end := time.Now()
    
    results, err := storage.Query(context.Background(), "test_counter", start, end)
    assert.NoError(t, err)
    assert.Equal(t, 100, len(results))
}
```

### Performance Tests
```go
func BenchmarkStore(b *testing.B) {
    storage := setupTestStorage(b)
    defer storage.Close()
    
    metric := storage.Metric{Name: "bench", Value: 1.0}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        storage.Store(context.Background(), metric)
    }
}
```

### Integration Tests
- OTEL collector → BadgerDB flow
- API endpoint responses
- TTL expiration verification
- Concurrent access

---

## 📦 Dependencies

### Blocking Dependencies
- **Story 9.2**: Metrics Storage ACL (must complete first)
  - **Required Deliverables**:
    - `pkg/metricstore/storage/interface.go` - Storage interface definition
    - `storage.Storage` interface frozen (no breaking changes)
    - `storage.Metric` struct definition complete
    - Mock implementation for testing
  - **Risk**: If Story 9.2 interface changes, adapter needs reimplementation
  - **Mitigation**: Interface review and approval before starting this story

### Technical Dependencies
- Go packages:
  - `github.com/dgraph-io/badger/v3` (v3.2103.5+)
  - `go.opentelemetry.io/otel` (v1.21.0+)
  - `go.opentelemetry.io/otel/sdk/metric` (v1.21.0+)
- Go version: 1.21+

### Knowledge Dependencies
- Team experience with BadgerDB: **TBD** (recommend Spike if unfamiliar)
- OTEL SDK familiarity: **TBD**

---

## 🚧 Migration & Rollout

### Phase 1: Development
- Implement BadgerDB adapter
- Local testing with sample data

### Phase 2: Integration
- Integrate with OTEL SDK
- Connect to existing metrics endpoints (Story 9.1)

### Phase 3: Staging
- Deploy to staging environment
- Load testing and validation

### Phase 4: Production
- Deploy MVP with BadgerDB backend
- Monitor performance and stability

### Rollback Plan
- **Phase 2**: If BadgerDB integration fails, revert to in-memory storage (temporary)
- **Phase 3**: Keep previous stable version deployment ready for 24h
- **Phase 4**: Monitor error rate, if > 1% or p95 latency > 200ms, trigger rollback
- **Rollback SLA**: < 15 minutes to previous version
- **Data preservation**: Export metrics before rollback (if possible)

---

## 📊 Success Metrics

- Write throughput > 1K metrics/min
- Query latency < 100ms (p95)
- Storage footprint < 100MB for 24h data
- Zero data loss with graceful shutdown
- CPU usage < 5% under normal load

---

## 🔗 Related Links

- [Epic 9: Observability & Monitoring](../../epics/9-observability-epic.md)
- [Story 9.2: Metrics Storage ACL](9-2-metrics-storage-acl.md)
- [Story 9.4: Prometheus Adapter](9-4-prometheus-adapter.md)
- [BadgerDB Documentation](https://dgraph.io/docs/badger/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)

---

## 📝 Notes

- BadgerDB provides ACID guarantees and crash recovery
- TTL-based expiration is automatic, no manual cleanup needed
- Key encoding critical for query performance
- Consider compression for production to reduce disk usage
- OTEL integration allows future migration to other backends

### Estimation Notes
- **8 Story Points**: Assumes team has Go and OTEL experience
- **Consider splitting** if team lacks BadgerDB experience:
  - Story 9.3a: BadgerDB Adapter (5 SP)
  - Story 9.3b: OTEL Integration + API (3 SP)
- **Spike recommended** (0.5-1 SP) if team unfamiliar with BadgerDB

### Team Discussion Points
1. Validate 8 SP estimate based on team experience
2. Confirm performance targets (1K metrics/min sufficient?)
3. Decide on API authentication approach (reuse existing?)
4. Prioritize storage layer monitoring vs. MVP scope

---

## ⚠️ Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| BadgerDB disk space exhaustion | High | Implement monitoring, TTL enforcement |
| Query performance degradation | Medium | Optimize key encoding, add caching |
| OTEL SDK overhead | Low | Profile and optimize metric collection |
| Data loss on crash | Medium | BadgerDB WAL ensures durability |

---

## ✅ Definition of Done

- [ ] Code reviewed and approved (at least 2 reviewers)
- [ ] All tests passing (unit + integration + performance)
- [ ] Unit test coverage > 85%
- [ ] Integration test coverage > 70%
- [ ] Edge cases tested:
  - [ ] Empty query results
  - [ ] Expired data (TTL)
  - [ ] Concurrent writes (race detector enabled)
  - [ ] Large result sets (> 10K metrics)
  - [ ] Invalid input handling
- [ ] API documentation published (Swagger/OpenAPI)
- [ ] Performance benchmarks meet criteria (documented results)
- [ ] Security scan completed (no high/critical vulnerabilities)
- [ ] Demo completed to team (stakeholder sign-off)
- [ ] Observability dashboard created (monitor storage health)
- [ ] Rollback procedure documented and tested
- [ ] No critical bugs or security issues
