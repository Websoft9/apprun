# Story 9.4: Prometheus Backend Adapter

**Story ID**: story-9.4  
**Epic**: [Epic 9: Observability & Monitoring](../../epics/9-observability-epic.md)  
**Status**: 📝 Ready for Dev  
**Priority**: P2 (Medium)  
**Estimate**: 7 Story Points (Updated after SM review)  
**Sprint**: TBD (Post-MVP)  
**Review Date**: 2026-01-21  
**Reviewed By**: Bob (Scrum Master)

---

## 📋 User Story

**As a** platform engineer  
**I want** a Prometheus storage adapter using the anti-corruption layer  
**So that** we can seamlessly migrate from BadgerDB to enterprise-grade monitoring

---

## 🎯 Acceptance Criteria

### Core Requirements
- [ ] Prometheus storage adapter implements storage interface from Story 9.2
- [ ] OTLP exporter configured for Prometheus remote write
- [ ] Seamless migration from BadgerDB without code changes
- [ ] Configuration-based backend switching
- [ ] Query API compatible with existing endpoints (Story 9.3)
- [ ] Dual-write mode for zero-downtime migration
- [ ] Configuration hot-reload without service restart
- [ ] Graceful degradation when Prometheus unavailable (fallback to BadgerDB)
- [ ] Query pagination consistent with Story 9.3 API
- [ ] Grafana dashboard compatibility verified

### Technical Requirements
- [ ] Prometheus remote write protocol support
- [ ] OTLP/HTTP exporter configuration
- [ ] Metric format conversion (OTEL → Prometheus)
- [ ] Error handling and retry logic with exponential backoff
- [ ] Health check for Prometheus connectivity
- [ ] TLS/SSL support for production deployment
- [ ] Bearer token authentication support
- [ ] Concurrent access test: 100 goroutines write metrics without race conditions
- [ ] Query result caching for high-frequency queries

### API Compatibility (Story 9.3 Endpoints)
- [ ] `POST /api/metrics/ingest` - Writes via OTLP exporter (API 已重构)
- [ ] `GET /api/metrics/history` - Queries Prometheus with PromQL translation (API 已重构)
- [ ] Storage health integrated into `/health` global endpoint (API 已重构)
- [ ] Response format identical to BadgerDB backend
- [ ] Error codes consistent with Story 9.3

### Performance Requirements
- [ ] Write throughput: ≥ 1K metrics/minute (same as BadgerDB)
- [ ] Query latency: < 200ms p95 (vs. BadgerDB < 100ms acceptable)
- [ ] Concurrent queries: Support 10+ simultaneous queries
- [ ] Memory usage: < 150MB typical workload
- [ ] Network latency handling: < 50ms to Prometheus

### Migration
- [ ] Data export utility from BadgerDB
- [ ] Import utility to Prometheus
- [ ] Migration documentation and runbook
- [ ] Zero downtime migration strategy
- [ ] Data consistency validation tool
- [ ] Rollback procedure documented

---

## 📐 Technical Design

### Prometheus Adapter Implementation
```go
// filepath: pkg/metricstore/storage/prometheus/adapter.go
package prometheus

import (
    "context"
    "fmt"
    "time"
    
    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
    
    "apprun/pkg/metricstore/storage"
)

type PrometheusStorage struct {
    client   api.Client
    api      v1.API
    exporter *otlpmetrichttp.Exporter
    config   Config
}

type Config struct {
    RemoteWriteURL string
    QueryURL       string
    Timeout        time.Duration
}

func NewPrometheusStorage(cfg Config) (*PrometheusStorage, error) {
    client, err := api.NewClient(api.Config{
        Address: cfg.QueryURL,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create prometheus client: %w", err)
    }
    
    // Setup OTLP exporter
    exporter, err := otlpmetrichttp.New(
        context.Background(),
        otlpmetrichttp.WithEndpoint(cfg.RemoteWriteURL),
        otlpmetrichttp.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create otlp exporter: %w", err)
    }
    
    return &PrometheusStorage{
        client:   client,
        api:      v1.NewAPI(client),
        exporter: exporter,
        config:   cfg,
    }, nil
}

func (s *PrometheusStorage) Store(ctx context.Context, metric storage.Metric) error {
    // Metrics are pushed via OTLP exporter, which is handled by OTEL SDK
    // This method can be no-op or handle direct remote write if needed
    return nil
}

func (s *PrometheusStorage) Query(ctx context.Context, name string, start, end time.Time) ([]storage.Metric, error) {
    // Query Prometheus using PromQL
    query := fmt.Sprintf("%s", name)
    
    result, warnings, err := s.api.QueryRange(ctx, query, v1.Range{
        Start: start,
        End:   end,
        Step:  time.Minute,
    })
    
    if err != nil {
        return nil, fmt.Errorf("prometheus query failed: %w", err)
    }
    
    if len(warnings) > 0 {
        // Log warnings
    }
    
    return convertPrometheusResult(result), nil
}

func (s *PrometheusStorage) Delete(ctx context.Context, before time.Time) error {
    // Prometheus handles retention via configuration
    return nil
}

func (s *PrometheusStorage) Health(ctx context.Context) error {
    _, err := s.api.LabelNames(ctx, nil, time.Now().Add(-1*time.Minute), time.Now())
    return err
}

func (s *PrometheusStorage) Close() error {
    return s.exporter.Shutdown(context.Background())
}

func convertPrometheusResult(result model.Value) []storage.Metric {
    // Convert Prometheus matrix/vector to storage.Metric slice
    var metrics []storage.Metric
    // ... conversion logic
    return metrics
}
```

### OTLP Configuration
```go
// filepath: pkg/metricstore/otel/prometheus.go
package otel

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
    "go.opentelemetry.io/otel/sdk/metric"
)

func SetupPrometheusOTEL(remoteWriteURL string) (*metric.MeterProvider, error) {
    exporter, err := otlpmetrichttp.New(
        context.Background(),
        otlpmetrichttp.WithEndpoint(remoteWriteURL),
        otlpmetrichttp.WithInsecure(), // Use WithTLSConfig for production
    )
    if err != nil {
        return nil, err
    }
    
    provider := metric.NewMeterProvider(
        metric.WithReader(metric.NewPeriodicReader(exporter,
            metric.WithInterval(10*time.Second),
        )),
    )
    
    otel.SetMeterProvider(provider)
    
    return provider, nil
}
```

### Configuration
```yaml
# filepath: config/metrics.yaml
metrics:
  storage:
    backend: prometheus  # Switch from "badger"
    dual_write: false    # Enable during migration phase
    fallback_backend: badger  # Fallback when Prometheus unavailable
    timeout: 30s
    retry:
      enabled: true
      max_attempts: 3
      backoff: exponential
    
    prometheus:
      remote_write_url: https://prometheus:9090/api/v1/write
      query_url: https://prometheus:9090
      timeout: 30s
      
      # TLS Configuration (Production)
      tls:
        enabled: true
        cert_file: /etc/apprun/certs/prometheus-client.crt
        key_file: /etc/apprun/certs/prometheus-client.key
        ca_file: /etc/apprun/certs/prometheus-ca.crt
        insecure_skip_verify: false
      
      # Authentication
      auth:
        type: bearer_token  # "bearer_token" or "basic"
        token_file: /etc/apprun/secrets/prometheus-token
        # For basic auth:
        # username: admin
        # password_file: /etc/apprun/secrets/prometheus-password
      
      # Query optimization
      query:
        max_samples: 50000000  # Limit result size
        timeout: 30s
        cache_ttl: 60s  # Cache frequently accessed queries
      
      # Health check
      health_check:
        enabled: true
        interval: 30s
        timeout: 5s

# Environment variable overrides:
# METRICS_STORAGE_BACKEND=prometheus
# METRICS_PROMETHEUS_REMOTE_WRITE_URL=https://prometheus:9090/api/v1/write
# METRICS_PROMETHEUS_TOKEN_FILE=/etc/secrets/prometheus-token
```

---

## � API Compatibility with Story 9.3

### Endpoint Behavior

All HTTP endpoints from Story 9.3 remain **unchanged** when switching to Prometheus backend:

#### POST /api/metrics/ingest (Updated API)
- **Backend Behavior**: Metrics pushed to OTLP exporter → Prometheus remote write
- **Adapter Method**: `Store()` is no-op, OTEL SDK handles export
- **Response Format**: Identical to BadgerDB (returns metric_id)
- **Dual-Write Mode**: Writes to both BadgerDB and Prometheus simultaneously

#### GET /api/metrics/history (Updated API)
- **Backend Behavior**: Query translated to PromQL, executed against Prometheus
- **PromQL Translation**: 
  - `name=http_requests_total` → `http_requests_total`
  - Time range → Prometheus `start`/`end` parameters
  - Pagination → `limit` parameter in PromQL
- **Response Format**: Converted from Prometheus matrix to Story 9.3 format
- **Fallback**: If Prometheus unavailable, query BadgerDB (if dual-write enabled)

#### Storage Health Check (Integrated into /health)
- **Backend Behavior**: Checks Prometheus connectivity via API
- **Health Criteria**:
  - ✅ Healthy: Prometheus responds < 5s
  - ⚠️ Degraded: Prometheus responds but slow (5-10s)
  - ❌ Unhealthy: Prometheus timeout or unreachable
- **Response Example**:
```json
{
  "status": "healthy",
  "backend": "prometheus",
  "details": {
    "prometheus_url": "https://prometheus:9090",
    "last_check": "2026-01-21T10:00:00Z",
    "latency_ms": 45
  }
}
```

### Migration Compatibility

**Dual-Write Mode** (config: `dual_write: true`):
- Metrics written to **both** BadgerDB and Prometheus
- Reads from **primary** backend (configurable)
- Automatic fallback if primary fails

**Fallback Strategy**:
```
Query Request
    |
    v
[Try Prometheus]
    |
    +-- Success --> Return results
    |
    +-- Failure --> [Try BadgerDB] --> Return results or error
```

---

## �🔧 Implementation Steps

1. **Prometheus Adapter Core** (2.5 SP)
   - Implement storage interface
   - Query API using PromQL translation
   - Health checks with retry logic
   - TLS and authentication support
   - Unit tests (15+ test cases)

2. **OTLP Integration + Dual-Write** (2 SP)
   - Configure OTLP/HTTP exporter
   - Metric format conversion (OTEL ↔ Prometheus)
   - Dual-write logic for migration phase
   - Fallback mechanism
   - Error handling with exponential backoff

3. **Migration Utilities** (1.5 SP)
   - Export from BadgerDB (streaming for large datasets)
   - Import to Prometheus via remote write
   - Data consistency validation tool
   - Rollback procedure implementation
   - Migration progress tracking

4. **Testing & Documentation** (1 SP)
   - Integration tests (local Prometheus setup)
   - Edge case tests (timeouts, network failures)
   - Concurrent write stress test (100 goroutines)
   - Performance benchmarking vs BadgerDB
   - Migration runbook with troubleshooting guide
   - API compatibility validation

---

## 🧪 Testing Strategy

### Unit Tests
```go
func TestPrometheusStorage_Query(t *testing.T) {
    // Mock Prometheus API
    storage := setupTestPrometheusStorage(t)
    
    start := time.Now().Add(-1 * time.Hour)
    end := time.Now()
    
    results, err := storage.Query(context.Background(), "http_requests_total", start, end)
    assert.NoError(t, err)
    assert.NotEmpty(t, results)
}

func TestPrometheusStorage_QueryEmptyResult(t *testing.T) {
    storage := setupTestPrometheusStorage(t)
    
    // Query non-existent metric
    results, err := storage.Query(context.Background(), "non_existent_metric", start, end)
    assert.NoError(t, err)
    assert.Empty(t, results)
}

func TestPrometheusStorage_QueryTimeout(t *testing.T) {
    storage := setupTestPrometheusStorageWithTimeout(t, 1*time.Millisecond)
    
    results, err := storage.Query(context.Background(), "http_requests_total", start, end)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}

func TestPrometheusStorage_ConcurrentWrites(t *testing.T) {
    storage := setupTestPrometheusStorage(t)
    
    // 100 concurrent goroutines
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            metric := storage.Metric{
                Name: "concurrent_test",
                Value: float64(id),
            }
            err := storage.Store(context.Background(), metric)
            assert.NoError(t, err)
        }(i)
    }
    wg.Wait()
}
```

### Edge Cases Testing
- [ ] Empty query results (metric doesn't exist)
- [ ] Prometheus timeout handling
- [ ] Network partition scenarios
- [ ] Invalid metric name/tags
- [ ] Query result size exceeding max_samples
- [ ] TLS certificate expiration
- [ ] Authentication token rotation
- [ ] Concurrent write stress (100+ goroutines)
- [ ] Query during Prometheus restart
- [ ] Fallback to BadgerDB when Prometheus unavailable

### Integration Tests
- Deploy local Prometheus (Docker)
- Test OTLP export end-to-end
- Verify query results match BadgerDB
- Test dual-write consistency
- Test migration workflow
- Grafana dashboard rendering

### Performance Testing
```bash
# Write throughput test
go test -bench=BenchmarkPrometheusWrite -benchtime=60s

# Query latency test
go test -bench=BenchmarkPrometheusQuery -benchtime=30s

# Concurrent query test
go test -bench=BenchmarkPrometheusQueryConcurrent -benchtime=30s
```

### Migration Testing
- Export 10K sample metrics from BadgerDB
- Import to Prometheus via remote write
- Validate data integrity (checksum comparison)
- Query performance comparison (BadgerDB vs Prometheus)
- Rollback test: revert to BadgerDB and verify data

---

## 📦 Dependencies

- **Story 9.2**: Metrics Storage ACL (must complete first)
- **Story 9.3**: BadgerDB implementation (for migration baseline)
- Go packages:
  - `github.com/prometheus/client_golang`
  - `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp`
- Infrastructure: Prometheus server deployment

---

## 🚧 Migration & Rollout

### Migration Strategy

#### Phase 1: Parallel Running
1. Deploy Prometheus alongside BadgerDB
2. Configure dual-write to both backends
3. Validate data consistency
4. Duration: 1-2 days

#### Phase 2: Switch Primary
1. Change configuration to use Prometheus as primary
2. Keep BadgerDB as fallback
3. Monitor for issues
4. Duration: 1 week

#### Phase 3: Cleanup
1. Stop BadgerDB writes
2. Archive historical data
3. Remove BadgerDB dependency
4. Duration: 1 day

### Migration Commands
```bash
# Export from BadgerDB
apprun observability export --backend badger --output /tmp/metrics.json

# Import to Prometheus (via remote write)
apprun observability import --backend prometheus --input /tmp/metrics.json

# Verify migration
apprun observability verify --compare badger,prometheus
```

---

## 📊 Success Metrics

### Performance Metrics
- **Write Throughput**: ≥ 1K metrics/minute (same as BadgerDB)
- **Query Latency**: 
  - p50: < 100ms
  - p95: < 200ms (vs. BadgerDB < 100ms acceptable)
  - p99: < 500ms
- **Concurrent Queries**: Support 10+ simultaneous queries without degradation
- **Memory Usage**: < 150MB typical workload (vs. BadgerDB < 100MB)
- **Network Latency**: < 50ms to Prometheus instance

### Migration Metrics
- **Zero Data Loss**: 100% of BadgerDB metrics migrated successfully
- **Downtime**: 0 seconds (dual-write strategy)
- **Configuration Switch**: < 5 minutes (hot reload)
- **Rollback Time**: < 2 minutes to revert to BadgerDB

### Quality Metrics
- **Test Coverage**: > 80% (unit + integration)
- **Edge Case Coverage**: All 10+ edge cases tested
- **Concurrent Safety**: Pass race detector with 100 goroutines
- **API Compatibility**: 100% endpoint behavior match with Story 9.3

### Operational Metrics
- **Prometheus Health Check**: < 5 seconds response time
- **Fallback Success Rate**: > 99% when Prometheus unavailable
- **TLS Handshake**: < 500ms
- **Authentication Success**: 100% with valid tokens

---

## 🔗 Related Links

- [Epic 9: Observability & Monitoring](../../epics/9-observability-epic.md)
- [Story 9.2: Metrics Storage ACL](9-2-metrics-storage-acl.md)
- [Story 9.3: BadgerDB Backend](9-3-badgerdb-otel.md)
- [Story 9.5: Prometheus Server Setup](9-5-prometheus-setup.md)
- [Prometheus Remote Write](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write)
- [OTLP Exporter](https://opentelemetry.io/docs/specs/otlp/)

---

## 📝 Notes

### Technical Decisions
- OTLP is the standard protocol for OpenTelemetry metrics
- Prometheus remote write protocol is production-ready
- Migration can be zero-downtime with proper planning
- Query interface remains identical thanks to ACL (Story 9.2)

### Known Technical Debt
1. **Store() Method as No-Op**: Current design relies on OTEL SDK for writing. Future may need direct remote write support for non-OTEL workflows.
   - **Impact**: Low (OTEL SDK covers 99% use cases)
   - **Resolution**: Story 9.6 (if needed)

2. **Delete() Method Limitation**: Prometheus retention is configuration-based, no API for selective deletion.
   - **Impact**: Medium (cannot delete specific metrics programmatically)
   - **Workaround**: Relabel metrics to expire them
   - **Resolution**: Document limitation in runbook

3. **No Query Result Caching**: High-frequency queries may cause Prometheus load.
   - **Impact**: Medium (performance under heavy load)
   - **Workaround**: cache_ttl configuration
   - **Resolution**: Story 9.7 - Redis cache layer

4. **Network Dependency**: Unlike BadgerDB (embedded), Prometheus requires network.
   - **Impact**: High (single point of failure)
   - **Mitigation**: Dual-write + fallback strategy
   - **Long-term**: Prometheus HA setup (Story 9.8)

### Questions for Product Owner
1. **Multi-tenancy**: Does Prometheus need tenant isolation? (Affects metric naming)
2. **Data Retention**: How long to keep historical data? (Affects storage planning)
3. **High Availability**: Need multiple Prometheus instances? (Affects architecture)

### Questions for Tech Lead
1. **Deployment**: Prometheus on K8s/VM/Docker?
2. **Operator**: Use Prometheus Operator for K8s?
3. **Certificate Management**: How to rotate TLS certs? (Cert-manager?)

---

## ⚠️ Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Prometheus unavailable during migration | High | Dual-write strategy, fallback to BadgerDB |
| Data format incompatibility | Medium | Thorough conversion testing |
| Query performance regression | Medium | Performance benchmarking, caching |
| Network latency to Prometheus | Low | Deploy Prometheus in same network |

---

## ✅ Definition of Done

- [ ] Code reviewed and approved
- [ ] All tests passing (unit + integration)
- [ ] Migration runbook documented and tested
- [ ] Performance benchmarks meet criteria
- [ ] Zero downtime migration validated
- [ ] Demo completed to team
- [ ] No critical bugs or security issues
