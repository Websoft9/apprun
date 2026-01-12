# Story 1.16: Redis Cache Package - Sprint Artifact

**Sprint**: Sprint 2  
**Story ID**: 1.16  
**Epic**: Epic 1 - Infrastructure & Foundation  
**Priority**: P1 (High)  
**Estimate**: 3 SP (13 hours)  
**Status**: ✅ Development Ready (Quality Validated)  

**Quality Score**: **A (Outstanding)** - 12 improvements applied

---

## 🎯 Quick Summary

**Goal**: Create unified Redis cache package for token blacklist, config caching, and session storage

**Why Now**: JWT blacklist has technical debt (direct redis.Client usage)

**Impact**: Foundation for 3+ Epic features (Auth, Config, Functions)

**Quality Validation**: Independent SM review completed, 12 improvements applied (see `validation-report-1.16-20260112.md`)

---

## ✅ Acceptance Criteria (7 total)

### AC1: Core Package Structure ✓
- [ ] Package at `pkg/cache/` with clean API
- [ ] Configuration via environment variables
- [ ] Connection pool + health check
- [ ] Graceful shutdown

### AC2: Key-Value Operations ✓
- [ ] Set/Get/Delete/Exists/TTL/Expire implemented
- [ ] Support string, int, JSON types
- [ ] TTL management working

### AC3: Pub/Sub Support ✓
- [ ] Publish/Subscribe/Unsubscribe implemented
- [ ] Message delivery verified

### AC4: Error Handling ✓
- [ ] Unified error wrapping (Story 1.3 pattern)
- [ ] Timeout handling (2s default)
- [ ] Fail-open/fail-fast strategies
- [ ] Structured logging (Story 1.10)

### AC5: Configuration ✓
- [ ] Environment variables defined (see below)
- [ ] Default values set
- [ ] Validation on startup

### AC6: JWT Refactor ✓
- [ ] Remove `redis.Client` from `internal/jwt/blacklist.go`
- [ ] Use `pkg/cache` instead
- [ ] All JWT tests pass

### AC7: Testing & Docs ✓
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests (real Redis)
- [ ] README.md with examples
- [ ] Godoc comments

---

## 🔧 Environment Variables

```bash
REDIS_HOST=localhost         # Redis server host
REDIS_PORT=6379             # Redis server port
REDIS_PASSWORD=             # Optional password
REDIS_DB=0                  # Database number
REDIS_POOL_SIZE=10          # Connection pool size
REDIS_TIMEOUT=2s            # Operation timeout
REDIS_FAIL_STRATEGY=open    # "open" or "fast"
```

---

## 📦 Implementation Tasks

### Task 1: Package Skeleton (1h)
```
- Create pkg/cache/ directory
- Define Client interface
- Create configuration struct
- Set up error types
```

### Task 2: Redis Client Wrapper (2h)
```
- Implement connection pool
- Add health check (Ping)
- Implement graceful shutdown
- Add config loading from env
```

### Task 3: Key-Value Operations (2h)
```
- Implement Set/Get/Delete
- Implement Exists/TTL/Expire
- Add error wrapping
- Add structured logging
```

### Task 4: Pub/Sub Support (2h)
```
- Implement Publish
- Implement Subscribe with callback
- Implement Unsubscribe
- Handle goroutine lifecycle
```

### Task 5: Testing (3h)
```
- Write unit tests (mock Redis)
- Write integration tests (real Redis)
- Achieve 80%+ coverage
- Add benchmarks
```

### Task 6: JWT Blacklist Refactor (2h)
```
- Update internal/jwt/blacklist.go
- Remove redis.Client dependency
- Run existing JWT tests
- Fix any breaking changes
```

### Task 7: Documentation (1h)
```
- Write README.md with examples
- Add Godoc comments
- Update Epic 1 progress
- Create usage guide
```

**Total**: 13 hours (~3 SP)

---

## 🧪 Testing Strategy

### Unit Tests (Mock Redis)
```go
// Use interface mocks
mockCache := mocks.NewMockClient()
mockCache.On("Set", ...).Return(nil)
```

### Integration Tests (Real Redis)
```go
// Use docker-compose Redis
client := cache.NewClient(&cache.Config{...})
// Test actual operations
```

### Regression Tests
```bash
# Run existing JWT tests
go test ./internal/jwt/... -v
# All must pass after refactor
```

---

## 📊 Definition of Done

- [ ] All 7 acceptance criteria met
- [ ] Code reviewed and approved
- [ ] All tests passing (unit + integration + JWT)
- [ ] Test coverage > 80%
- [ ] Documentation complete (README + Godoc)
- [ ] No new linter warnings
- [ ] Deployed to dev environment
- [ ] Epic 1 progress updated in sprint-status.yaml

---

## 🚨 Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| JWT tests fail after refactor | Medium | High | Maintain exact API contract |
| Redis unavailable in CI | Low | Medium | Use testcontainers or mock |
| Performance degradation | Low | Medium | Add benchmarks, compare with baseline |

---

## 🔗 Related Documentation

### Required Reading (Before Starting)
1. [Full Story Spec](../../epics/stories/story-1.16-redis-cache-package.md) - **✅ Updated with 12 improvements**
2. [Redis Cache Quick Start](../../standards/redis-cache-guide.md) - Usage patterns
3. [Current JWT Blacklist](../../../core/internal/jwt/blacklist.go) - Code to refactor
4. [Validation Report](validation-report-1.16-20260112.md) - Quality analysis

### Critical Context (NEW - Must Read)
- **Redis Version**: v9.17.2 (pinned, DO NOT upgrade)
- **JWT API Contract**: Explicit function signatures documented
- **Error Codes**: `CACHE_CONNECT_001`, `CACHE_OP_002`, etc.
- **Config Pattern**: Use `pkg/env.Get()` (follows database package)
- **Test Fixtures**: `testdata/fixtures.go` pattern provided

### Reference Patterns
- [Logger Package](../../epics/stories/story-1.10-logger-package.md) - Logging pattern
- [Database Package](../../epics/stories/story-1.14-database-package.md) - Anti-corruption layer
- [Error Framework](../../epics/stories/story-1.3-error-framework.md) - Error handling

---

## 📝 Developer Notes

### Key Design Decisions
1. **Fail-Open by default** - Business is stateless, continue if Redis down
2. **Interface-based design** - Enable mocking and testing
3. **Key naming convention** - Use `module:type:id` format (e.g., `jwt:blacklist:token123`)

### Common Pitfalls to Avoid
- ❌ Don't cache large objects (>1MB)
- ❌ Don't forget TTL for temporary data
- ❌ Don't store sensitive data without encryption

### Performance Targets
- Latency: < 5ms (p95)
- Throughput: > 1000 ops/sec per connection
- Pool size: 10 connections handles ~10k ops/sec

---

## 🏃 Sprint Tracking

**Assigned To**: TBD  
**Started**: -  
**Completed**: -  
**Actual Effort**: - SP

### Daily Progress
```markdown
[YYYY-MM-DD] - Task X completed, Y% done
[YYYY-MM-DD] - Blocker: [description]
```

---

**Created**: 2026-01-12  
**Prepared By**: Bob (Scrum Master Agent)  
**Status**: ✅ Development Ready  
**Sprint**: Sprint 2
