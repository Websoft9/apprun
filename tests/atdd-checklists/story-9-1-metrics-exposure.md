# ATDD Implementation Checklist
# Story 9.1: Metrics Exposure

**Status**: 🔴 RED PHASE - Tests written, implementation pending  
**Story File**: [specs/sprint-artifacts/sprint-3/9-1-metrics-exposure.md](../../specs/sprint-artifacts/sprint-3/9-1-metrics-exposure.md)  
**Test File**: [tests/integration/metrics/metrics_exposure_test.go](../integration/metrics/metrics_exposure_test.go)  
**Created**: 2026-01-23  
**Owner**: Backend Dev Team

---

## 📋 Overview

This document guides the implementation of Story 9.1 following **Test-Driven Development (TDD)** red-green-refactor cycle.

**Current Phase**: 🔴 **RED** - Tests are failing (expected)  
**Next Phase**: 🟢 **GREEN** - Implement code to pass tests  
**Final Phase**: 🔵 **REFACTOR** - Improve code quality

---

## 🔴 RED PHASE: Test Status

### Failing Tests Created

All tests are in **RED phase** (failing as expected):

```bash
# Run failing tests
cd tests
go test -v ./integration/metrics/

# Expected output: FAIL (no implementation yet)
```

| Test Function | Status | Covers AC |
|--------------|--------|-----------|
| `TestMetricsSnapshot_Success` | 🔴 FAIL | AC1: Basic endpoint |
| `TestMetricsSnapshot_AdminOnly` | 🔴 FAIL | AC: Security |
| `TestMetricsHistory_WithTimeRange` | 🔴 FAIL | AC: Historical query |
| `TestMetricsHistory_TagFiltering` | 🔴 FAIL | AC: Tag filtering |
| `TestMetricsHistory_StorageDegradation` | 🔴 FAIL | AC: Graceful degradation |
| `TestMetricsIngest_BatchIngestion` | 🔴 FAIL | AC: Batch ingest |
| `TestMetricsKeys_Discovery` | 🔴 FAIL | AC: Metric discovery |
| `TestMetricsCache_HitAndMiss` | 🔴 FAIL | AC: Redis caching |

---

## 🟢 GREEN PHASE: Implementation Checklist

Follow these tasks **in order** to turn tests green. Check off each item as completed.

### Day 1: Core Infrastructure (4-6 hours)

#### Task 1.1: Create Metrics Collector
- [ ] Create `core/modules/metrics/collector.go`
- [ ] Implement `MetricsCollector` struct with fields:
  - [ ] `entClient *ent.Client`
  - [ ] `repo *metricstore.Repository`
  - [ ] `hostname string`
  - [ ] `env string` (production/staging/dev)
- [ ] Implement `NewMetricsCollector()` constructor
- [ ] Add `persistMetricsWithTags()` helper method

**Verification**:
```bash
go build ./core/modules/metrics/collector.go
# Should compile without errors
```

#### Task 1.2: Implement User Metrics Collection
- [ ] Implement `CollectUserMetrics(ctx context.Context) (*UserMetrics, error)`
- [ ] Query database for:
  - [ ] `total_users` - Count all users
  - [ ] `active_users` - Count where `is_active = true`
  - [ ] `admin_users` - Count where `role = 'platform_admin'`
  - [ ] `banned_users` - Count where `is_banned = true`
  - [ ] `new_users_today` - Count where `created_at >= today`
- [ ] Add tags to metrics:
  - [ ] `env` - Environment identifier
  - [ ] `instance` - Hostname/instance ID
  - [ ] `source: "database"` - Data source
- [ ] Call `persistMetricsWithTags()` asynchronously (use goroutine)

**Verification**:
```bash
# Run unit test (create if needed)
go test -run TestCollectUserMetrics ./core/modules/metrics/
```

#### Task 1.3: Implement System Metrics Collection
- [ ] Install dependency: `go get github.com/shirou/gopsutil/v3`
- [ ] Implement `CollectSystemMetrics(ctx context.Context) (*SystemMetrics, error)`
- [ ] Collect system metrics:
  - [ ] `uptime_seconds` - Time since service start
  - [ ] `memory_usage_mb` - Go runtime memory
  - [ ] `cpu_usage_percent` - CPU usage (gopsutil)
  - [ ] `disk_usage_percent` - Disk usage (gopsutil)
  - [ ] `goroutines` - `runtime.NumGoroutine()`
- [ ] Add tags including `host` label
- [ ] Persist to storage asynchronously

**Verification**:
```bash
go test -run TestCollectSystemMetrics ./core/modules/metrics/
```

---

### Day 2: API Endpoints (6-8 hours)

#### Task 2.1: Extend Metrics Handler
- [ ] Open `core/modules/metrics/handler.go`
- [ ] Add fields to `MetricsHandler`:
  - [ ] `collector *MetricsCollector`
  - [ ] `redis *redis.Client`
- [ ] Update `NewMetricsHandler()` to accept dependencies

#### Task 2.2: Implement GET /api/metrics/snapshot
- [ ] Implement `GetSnapshot(w http.ResponseWriter, r *http.Request)` method
- [ ] Check Redis cache first (key: `metrics:snapshot`)
- [ ] On cache miss:
  - [ ] Call `collector.CollectUserMetrics()`
  - [ ] Call `collector.CollectSystemMetrics()`
  - [ ] Merge results into single map
  - [ ] Store in Redis (TTL: 1 minute)
- [ ] Return JSON response with `cache_hit` flag
- [ ] Add Swagger annotations

**Test Command**:
```bash
go test -run TestMetricsSnapshot_Success ./tests/integration/metrics/
```

**Expected**: Test should PASS ✅

#### Task 2.3: Implement Admin-Only Middleware
- [ ] Open `core/middleware/auth.go`
- [ ] Implement `RequirePlatformAdmin()` middleware
- [ ] Check JWT token role == `"platform_admin"`
- [ ] Return 403 with `PERM_ADMIN_REQUIRED` error if not admin
- [ ] Allow request to proceed if admin

**Test Command**:
```bash
go test -run TestMetricsSnapshot_AdminOnly ./tests/integration/metrics/
```

**Expected**: Test should PASS ✅

#### Task 2.4: Implement GET /api/metrics/history
- [ ] Implement `GetHistory(w http.ResponseWriter, r *http.Request)` method
- [ ] Parse query parameters:
  - [ ] `name` (required) - Metric name
  - [ ] `duration` (optional, default: 24h) - Time range
  - [ ] `start` / `end` (optional) - Explicit time bounds
  - [ ] `tags[key]` (optional) - Tag filters
- [ ] Call `repo.GetMetrics(ctx, name, start, end)`
- [ ] Implement graceful degradation:
  ```go
  metrics, err := repo.GetMetrics(...)
  if err != nil {
      logger.Warn("Storage unavailable", "error", err)
      metrics = []Metric{} // Return empty array
  }
  ```
- [ ] Implement tag filtering:
  - [ ] Parse `tags[key]=value` from query params
  - [ ] Filter results by matching tags
- [ ] Return JSON with `metrics`, `count`, `start`, `end`, `has_more`

**Test Commands**:
```bash
go test -run TestMetricsHistory_WithTimeRange ./tests/integration/metrics/
go test -run TestMetricsHistory_TagFiltering ./tests/integration/metrics/
go test -run TestMetricsHistory_StorageDegradation ./tests/integration/metrics/
```

**Expected**: All 3 tests should PASS ✅

---

### Day 3: Discovery & Ingestion (4-6 hours)

#### Task 3.1: Implement GET /api/metrics/keys
- [ ] Create metric catalog in `core/modules/metrics/catalog.go`
- [ ] Define standard metrics:
  ```go
  type MetricDefinition struct {
      Name        string
      Source      string
      Description string
  }
  
  var StandardMetrics = []MetricDefinition{
      {Name: "user_count_total", Source: "user", Description: "Total user count"},
      {Name: "system_cpu_percent", Source: "system", Description: "CPU usage percentage"},
      // ... add all metrics
  }
  ```
- [ ] Implement `GetKeys(w http.ResponseWriter, r *http.Request)` method
- [ ] Return metric catalog as JSON

**Test Command**:
```bash
go test -run TestMetricsKeys_Discovery ./tests/integration/metrics/
```

**Expected**: Test should PASS ✅

#### Task 3.2: Implement POST /api/metrics/ingest
- [ ] Implement `Ingest(w http.ResponseWriter, r *http.Request)` method
- [ ] Parse JSON array of metrics from request body
- [ ] Validate each metric:
  - [ ] `name` is required
  - [ ] `value` is required (float64)
  - [ ] `timestamp` is optional (default: now)
  - [ ] `tags` is optional (map[string]string)
- [ ] Call `repo.RecordMetric()` for each valid metric
- [ ] Count accepted/rejected metrics
- [ ] Return summary: `{accepted: N, rejected: M, errors: []}`

**Test Command**:
```bash
go test -run TestMetricsIngest_BatchIngestion ./tests/integration/metrics/
```

**Expected**: Test should PASS ✅

#### Task 3.3: Update Router Configuration
- [ ] Open `core/routes/router.go`
- [ ] Create metrics group:
  ```go
  metricsGroup := r.Group("/api/metrics")
  metricsGroup.Use(middleware.RequirePlatformAdmin())
  {
      metricsGroup.GET("/snapshot", metricsHandler.GetSnapshot)
      metricsGroup.GET("/history", metricsHandler.GetHistory)
      metricsGroup.GET("/keys", metricsHandler.GetKeys)
      metricsGroup.POST("/ingest", metricsHandler.Ingest)
  }
  ```

---

### Day 4: Caching & Polish (4 hours)

#### Task 4.1: Implement Redis Caching Service
- [ ] Create `core/modules/metrics/service.go`
- [ ] Implement `MetricsService` struct:
  ```go
  type MetricsService struct {
      collector *MetricsCollector
      redis     *redis.Client
  }
  ```
- [ ] Implement cache methods:
  - [ ] `GetUserMetrics()` - with cache
  - [ ] `GetSystemMetrics()` - with cache
- [ ] Set TTL to 1 minute
- [ ] Set `cache_hit` flag in response

**Test Command**:
```bash
go test -run TestMetricsCache_HitAndMiss ./tests/integration/metrics/
```

**Expected**: Test should PASS ✅

#### Task 4.2: Run All Tests
- [ ] Run full test suite:
  ```bash
  cd tests
  go test -v ./integration/metrics/
  ```
- [ ] **All 8 tests should PASS** ✅

#### Task 4.3: Update Documentation
- [ ] Regenerate Swagger docs:
  ```bash
  make swagger
  ```
- [ ] Update [docs/api.md](../docs/api.md) with new endpoints
- [ ] Verify examples in story file work

---

## 🔵 REFACTOR PHASE: Code Quality

Once all tests are **GREEN** ✅, improve code quality:

### Refactoring Checklist

- [ ] **Extract Duplications**:
  - [ ] Common tag generation logic
  - [ ] Redis cache wrapper functions
  - [ ] Query parameter parsing helpers

- [ ] **Add Error Handling**:
  - [ ] Wrap errors with context
  - [ ] Log errors appropriately
  - [ ] Return user-friendly error messages

- [ ] **Optimize Performance**:
  - [ ] Use connection pooling for DB queries
  - [ ] Add query timeouts
  - [ ] Profile slow queries

- [ ] **Improve Testability**:
  - [ ] Extract interfaces for collector
  - [ ] Add mock implementations
  - [ ] Reduce coupling between components

- [ ] **Code Review**:
  - [ ] Run linters: `make lint`
  - [ ] Fix all warnings
  - [ ] Get peer review approval

---

## 📊 Acceptance Criteria Mapping

| AC | Test Function | Implementation Task |
|----|--------------|---------------------|
| AC1: Basic endpoint | `TestMetricsSnapshot_Success` | Task 2.2 |
| AC2: Database health check | `TestMetricsSnapshot_Success` | Task 1.2 |
| AC3: Cache health check | `TestMetricsCache_HitAndMiss` | Task 4.1 |
| AC4: Metrics storage | `TestMetricsHistory_WithTimeRange` | Task 2.4 |
| AC5: Overall status | `TestMetricsSnapshot_Success` | Task 2.2 |
| AC6: HTTP status codes | `TestMetricsSnapshot_AdminOnly` | Task 2.3 |
| AC7: Performance & timeout | (Load testing - separate) | - |
| AC: Security | `TestMetricsSnapshot_AdminOnly` | Task 2.3 |
| AC: Tag support | `TestMetricsHistory_TagFiltering` | Task 2.4 |
| AC: Graceful degradation | `TestMetricsHistory_StorageDegradation` | Task 2.4 |

---

## 🛠️ Required Data Structures

### Metrics Types (core/modules/metrics/types.go)

Already defined in story file - ensure these exist:

- `MetricsResponse` - Generic metrics container
- `UserMetrics` - User statistics
- `SystemMetrics` - System health metrics
- `PerformanceMetrics` - API performance (stub for now)
- `AuthMetrics` - Authentication metrics (optional)

### Database Indexes

Already exist (no migration needed):
```sql
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at);
```

---

## 🎯 Definition of Done

Before marking story complete, verify:

- [x] **All 8 ATDD tests PASS** ✅
- [ ] Unit test coverage ≥ 80%
- [ ] Integration tests pass
- [ ] Linter checks pass (`make lint`)
- [ ] Manual API testing successful:
  ```bash
  # Test as admin
  curl -H "Authorization: Bearer $ADMIN_TOKEN" \
       http://localhost:8080/api/metrics/snapshot
  
  # Test as regular user (should fail)
  curl -H "Authorization: Bearer $USER_TOKEN" \
       http://localhost:8080/api/metrics/snapshot
  
  # Test history query
  curl -H "Authorization: Bearer $ADMIN_TOKEN" \
       "http://localhost:8080/api/metrics/history?name=user_count_total&duration=1h"
  ```
- [ ] Swagger docs updated
- [ ] Code review approved
- [ ] Performance targets met:
  - Cache hit response < 50ms (P95)
  - Cache miss response < 200ms (P95)
- [ ] Story marked **DONE** in sprint tracking

---

## 🚀 Running Tests

### Run Specific Test
```bash
cd tests
go test -v -run TestMetricsSnapshot_Success ./integration/metrics/
```

### Run All Metrics Tests
```bash
cd tests
go test -v ./integration/metrics/
```

### Run with Coverage
```bash
cd tests
go test -coverprofile=coverage.out ./integration/metrics/
go tool cover -html=coverage.out
```

### Watch Mode (Optional)
```bash
# Install gotestsum
go install gotest.tools/gotestsum@latest

# Watch tests
gotestsum --watch -- ./integration/metrics/
```

---

## 📚 Resources

- **Story File**: [specs/sprint-artifacts/sprint-3/9-1-metrics-exposure.md](../sprint-artifacts/sprint-3/9-1-metrics-exposure.md)
- **Test File**: [tests/integration/metrics/metrics_exposure_test.go](../../tests/integration/metrics/metrics_exposure_test.go)
- **Test Helpers**: 
  - [tests/testutils/metrics_helper.go](../../tests/testutils/metrics_helper.go)
  - [tests/testutils/auth_helper.go](../../tests/testutils/auth_helper.go)
- **Related Stories**:
  - Story 9.2: Metrics Storage ACL ✅ Complete
  - Story 9.3: BadgerDB Backend ✅ Complete

---

## 🐛 Common Issues & Solutions

### Issue: Tests can't find MetricsCollector
**Solution**: Ensure `core/modules/metrics/collector.go` exists and compiles

### Issue: Redis connection failed
**Solution**: 
```bash
# Start Redis in Docker
make deps-start

# Or use mock Redis for tests
```

### Issue: BadgerDB not initialized
**Solution**: Initialize repository in test setup:
```go
repo := metricstore.NewRepository(badgerDB)
```

### Issue: JWT token invalid
**Solution**: Verify JWT secret matches in tests and config:
```go
// Use consistent secret in tests
tokenString, _ := token.SignedString([]byte("test_jwt_secret"))
```

---

## ✅ Progress Tracking

**Created**: 2026-01-23  
**Last Updated**: 2026-01-23  
**Status**: 🔴 RED PHASE  

Track your progress by checking off tasks above. Once all tests pass, move to REFACTOR phase.

**Good luck! 🚀**
