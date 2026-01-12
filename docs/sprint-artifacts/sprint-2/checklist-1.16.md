# Story 1.16 - Development Checklist

**Story**: Redis Cache Package  
**Sprint**: Sprint 2  
**Developer**: TBD  
**Status**: 📝 Ready to Start

---

## 🎯 Pre-Development Checklist

### Before You Start
- [ ] Read [Full Story Spec](../../epics/stories/story-1.16-redis-cache-package.md)
- [ ] Review [Redis Cache Quick Start Guide](../../standards/redis-cache-guide.md)
- [ ] Check current [JWT Blacklist Implementation](../../../core/internal/jwt/blacklist.go)
- [ ] Ensure Redis running: `docker-compose ps redis`
- [ ] Verify development environment: `make dev-start`

---

## 📋 Implementation Checklist

### Phase 1: Package Setup (Day 1 - 2h)
- [ ] Create `pkg/cache/` directory structure
- [ ] Define `Client` interface in `cache.go`
- [ ] Create `Config` struct in `config.go`
- [ ] Define error types in `errors.go`
- [ ] Create basic README.md template

**Exit Criteria**: Package compiles without errors

---

### Phase 2: Core Implementation (Day 1-2 - 4h)
- [ ] Implement connection pool in `client.go`
  - [ ] LoadConfig() from environment
  - [ ] NewClient() with pool initialization
  - [ ] Ping() health check
  - [ ] Close() graceful shutdown
  
- [ ] Implement Key-Value operations in `operations.go`
  - [ ] Set(ctx, key, value, ttl)
  - [ ] Get(ctx, key)
  - [ ] Delete(ctx, key)
  - [ ] Exists(ctx, key)
  - [ ] TTL(ctx, key)
  - [ ] Expire(ctx, key, ttl)

**Exit Criteria**: All KV operations work with real Redis

---

### Phase 3: Pub/Sub (Day 2 - 2h)
- [ ] Implement Pub/Sub in `pubsub.go`
  - [ ] Publish(ctx, channel, message)
  - [ ] Subscribe(ctx, channel, handler)
  - [ ] Unsubscribe(ctx, channel)
  - [ ] Goroutine lifecycle management

**Exit Criteria**: Pub/Sub message delivery verified

---

### Phase 4: Testing (Day 2-3 - 3h)
- [ ] Unit tests in `cache_test.go`
  - [ ] Mock Redis client
  - [ ] Test all operations
  - [ ] Test error scenarios
  - [ ] Test fail-open/fail-fast
  
- [ ] Integration tests in `integration_test.go`
  - [ ] Use real Redis (docker-compose)
  - [ ] Test connection pool
  - [ ] Test concurrent operations
  - [ ] Benchmark performance

- [ ] Verify coverage: `go test -cover ./pkg/cache/...`
  - [ ] Target: > 80%

**Exit Criteria**: All tests pass, coverage > 80%

---

### Phase 5: JWT Refactor (Day 3 - 2h)
- [ ] Update `internal/jwt/blacklist.go`
  - [ ] Replace `redis.Client` with `cache.Client`
  - [ ] Update `InitBlacklist()` signature
  - [ ] Maintain exact API contract
  
- [ ] Run existing JWT tests
  - [ ] `go test ./internal/jwt/... -v`
  - [ ] All tests must pass
  - [ ] Zero behavioral changes

**Exit Criteria**: JWT tests pass without modification

---

### Phase 6: Documentation (Day 3 - 1h)
- [ ] Complete `pkg/cache/README.md`
  - [ ] Installation instructions
  - [ ] Configuration examples
  - [ ] Usage examples
  - [ ] Common patterns
  
- [ ] Add Godoc comments
  - [ ] All public types
  - [ ] All public functions
  - [ ] Usage examples in comments

**Exit Criteria**: `godoc` output is clear and complete

---

## ✅ Acceptance Criteria Verification

### AC1: Core Package Structure
- [ ] Package at `pkg/cache/` ✓
- [ ] Configuration via env vars ✓
- [ ] Connection pool working ✓
- [ ] Graceful shutdown tested ✓

### AC2: Key-Value Operations
- [ ] All 6 operations implemented ✓
- [ ] String/int/JSON types tested ✓
- [ ] TTL management verified ✓

### AC3: Pub/Sub Support
- [ ] Publish working ✓
- [ ] Subscribe working ✓
- [ ] Unsubscribe working ✓

### AC4: Error Handling
- [ ] Unified error wrapping ✓
- [ ] Timeout handling ✓
- [ ] Fail-open/fail-fast strategies ✓
- [ ] Structured logging ✓

### AC5: Configuration
- [ ] All env vars defined ✓
- [ ] Defaults set ✓
- [ ] Validation on startup ✓

### AC6: JWT Refactor
- [ ] `internal/jwt/blacklist.go` updated ✓
- [ ] All JWT tests passing ✓
- [ ] No breaking changes ✓

### AC7: Testing & Docs
- [ ] Unit tests > 80% coverage ✓
- [ ] Integration tests passing ✓
- [ ] README.md complete ✓
- [ ] Godoc comments added ✓

---

## 🔍 Code Review Checklist

### Before Submitting PR
- [ ] All tests passing: `make test`
- [ ] No linter warnings: `make lint`
- [ ] Code formatted: `make fmt`
- [ ] Documentation updated
- [ ] Epic 1 progress updated
- [ ] sprint-status.yaml updated to "done"

---

## 📊 Testing Commands

```bash
# Run all tests
make test

# Run cache package tests only
go test -v ./pkg/cache/...

# Run with coverage
go test -cover ./pkg/cache/...

# Run integration tests (requires Redis)
docker-compose up -d redis
go test -v -tags=integration ./pkg/cache/...

# Run JWT tests (after refactor)
go test -v ./internal/jwt/...

# Benchmark
go test -bench=. ./pkg/cache/...
```

---

## 🚨 Common Issues & Solutions

### Issue: Redis connection refused
```bash
# Solution: Start Redis
docker-compose up -d redis
```

### Issue: Tests fail with timeout
```bash
# Solution: Increase timeout in tests
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
```

### Issue: JWT tests fail after refactor
```bash
# Solution: Check API contract
# Ensure blacklist.go maintains same function signatures
```

---

## 📝 Daily Standup Template

```markdown
**Story 1.16 - [Date]**
- Yesterday: [Completed phases/tasks]
- Today: [Planned phases/tasks]
- Blockers: [Any impediments]
- Progress: [X/7 phases complete]
```

---

## ✅ Definition of Done

- [ ] All 7 acceptance criteria met
- [ ] Code reviewed and approved
- [ ] All tests passing (unit + integration + JWT)
- [ ] Test coverage > 80%
- [ ] Documentation complete
- [ ] No new linter warnings
- [ ] Deployed to dev environment
- [ ] Epic 1 progress updated
- [ ] PR merged to develop branch

---

**Created**: 2026-01-12  
**Scrum Master**: Bob  
**Status**: ✅ Ready for Development
