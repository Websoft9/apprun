# Story 9.1: Metrics Exposure
# Sprint 3: Observability & Monitoring (Epic 9)

**Priority**: P1  
**Effort**: 3 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 1.16 (Redis Cache Package) - ✅ Complete
- Story 5.7 (Platform User Management) - ✅ Complete
- Story 5.9 (Audit Logging) - Optional (for auth metrics)

**Status**: 📝 Ready for Dev  
**Module**: Metrics  
**Epic**: [observability-epic](../../epics/9-observability-epic.md)  
**Related PRD**: [FR-MON-001](../../prd.md#fr-mon-001-system-monitoring--metrics)  
**Issue**: #TBD

---

## User Story

作为 **平台管理员 (platform_admin)**，我希望能通过 API 查看系统运行指标（用户统计、系统健康、API性能），以便监控平台健康状态并做出运维决策。

---

## Acceptance Criteria

### 功能验收
- [ ] 实现 `GET /api/metrics` - 获取所有指标
- [ ] 实现 `GET /api/metrics/users` - 获取用户指标
- [ ] 实现 `GET /api/metrics/system` - 获取系统健康指标
- [ ] 实现 `GET /api/metrics/performance` - 获取性能指标
- [ ] 支持4类指标：User, Auth, Performance, System Health

### 性能验收
- [ ] 启用缓存时响应时间 < 200ms
- [ ] 未命中缓存时响应时间 < 1s
- [ ] Redis 缓存 TTL = 5 分钟
- [ ] 数据库查询使用索引（users表、audit_logs表）

### 安全验收
- [ ] 所有端点仅限 `platform_admin` 角色访问
- [ ] 非管理员返回 403 + `PERM_ADMIN_REQUIRED` 错误码
- [ ] 指标不包含 PII 或敏感用户数据

### 测试验收
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖缓存命中/未命中场景
- [ ] 测试权限校验（admin vs non-admin）

---

## Technical Design

### 1. Module Structure

```
core/modules/metrics/
├── handler.go           # NEW: Metrics endpoints
├── service.go           # NEW: Metrics business logic
├── collector.go         # NEW: Data collection logic
├── types.go             # NEW: Metrics DTOs
└── config.go            # NEW: Metrics constants
```

**Design Decision**: Metrics is an independent module, not part of admin. Admin operations (user CRUD) live in `core/modules/admin/`, while metrics (read-only resource) has its own module.

### 2. API Endpoints

#### 2.1 GET /api/metrics
**描述**: 获取所有指标（聚合）  
**认证**: JWT + `platform_admin`  
**查询参数**: 无

**响应 (200 OK)**:
```json
{
  "success": true,
  "data": {
    "metrics": {
      "total_users": 1250,
      "active_users": 1180,
      "admin_users": 5,
      "banned_users": 15,
      "new_users_today": 23,
      "user_registrations_last_7_days": 142,
      "login_attempts_total": 15420,
      "login_success_rate": 94.5,
      "failed_login_attempts": 847,
      "token_issued_total": 14573,
      "api_requests_total": 324567,
      "api_response_time_p95": 245,
      "api_error_rate": 1.2,
      "uptime_seconds": 86400,
      "memory_usage_mb": 512,
      "cpu_usage_percent": 35.2,
      "disk_usage_percent": 42.8
    },
    "timestamp": "2026-01-20T12:00:00Z",
    "cache_hit": true
  }
}
```

**响应 (403 Forbidden - 非管理员)**:
```json
{
  "success": false,
  "error": {
    "code": "PERM_ADMIN_REQUIRED",
    "message": "需要平台管理员权限"
  }
}
```

#### 2.2 GET /api/metrics/users
**描述**: 获取用户指标  
**认证**: JWT + `platform_admin`

**响应 (200 OK)**:
```json
{
  "success": true,
  "data": {
    "total_users": 1250,
    "active_users": 1180,
    "admin_users": 5,
    "banned_users": 15,
    "new_users_today": 23,
    "user_registrations_last_7_days": 142,
    "timestamp": "2026-01-20T12:00:00Z",
    "cache_hit": true
  }
}
```

#### 2.3 GET /api/metrics/system
**描述**: 获取系统健康指标  
**认证**: JWT + `platform_admin`

**响应 (200 OK)**:
```json
{
  "success": true,
  "data": {
    "uptime_seconds": 86400,
    "memory_usage_mb": 512,
    "cpu_usage_percent": 35.2,
    "disk_usage_percent": 42.8,
    "goroutines": 245,
    "timestamp": "2026-01-20T12:00:00Z"
  }
}
```

#### 2.4 GET /api/metrics/performance
**描述**: 获取API性能指标  
**认证**: JWT + `platform_admin`

**响应 (200 OK)**:
```json
{
  "success": true,
  "data": {
    "api_requests_total": 324567,
    "api_response_time_p95": 245,
    "api_error_rate": 1.2,
    "database_query_duration_avg": 12.5,
    "timestamp": "2026-01-20T12:00:00Z",
    "cache_hit": true
  }
}
```

---

### 3. Data Types (types.go)

```go
// MetricsResponse - 通用指标响应
type MetricsResponse struct {
    Metrics   map[string]interface{} `json:"metrics"`
    Timestamp time.Time              `json:"timestamp"`
    CacheHit  bool                   `json:"cache_hit,omitempty"`
}

// UserMetrics - 用户指标
type UserMetrics struct {
    TotalUsers                int       `json:"total_users"`
    ActiveUsers               int       `json:"active_users"`
    AdminUsers                int       `json:"admin_users"`
    BannedUsers               int       `json:"banned_users"`
    NewUsersToday             int       `json:"new_users_today"`
    UserRegistrationsLast7Days int      `json:"user_registrations_last_7_days"`
    Timestamp                 time.Time `json:"timestamp"`
    CacheHit                  bool      `json:"cache_hit,omitempty"`
}

// SystemMetrics - 系统健康指标
type SystemMetrics struct {
    UptimeSeconds    int64     `json:"uptime_seconds"`
    MemoryUsageMB    uint64    `json:"memory_usage_mb"`
    CPUUsagePercent  float64   `json:"cpu_usage_percent"`
    DiskUsagePercent float64   `json:"disk_usage_percent"`
    Goroutines       int       `json:"goroutines"`
    Timestamp        time.Time `json:"timestamp"`
}

// PerformanceMetrics - API性能指标
type PerformanceMetrics struct {
    APIRequestsTotal          int64     `json:"api_requests_total"`
    APIResponseTimeP95        int       `json:"api_response_time_p95"` // milliseconds
    APIErrorRate              float64   `json:"api_error_rate"`        // percentage
    DatabaseQueryDurationAvg  float64   `json:"database_query_duration_avg"` // milliseconds
    Timestamp                 time.Time `json:"timestamp"`
    CacheHit                  bool      `json:"cache_hit,omitempty"`
}

// AuthMetrics - 认证指标
type AuthMetrics struct {
    LoginAttemptsTotal  int64     `json:"login_attempts_total"`
    LoginSuccessRate    float64   `json:"login_success_rate"` // percentage
    FailedLoginAttempts int64     `json:"failed_login_attempts"`
    TokenIssuedTotal    int64     `json:"token_issued_total"`
    Timestamp           time.Time `json:"timestamp"`
    CacheHit            bool      `json:"cache_hit,omitempty"`
}
```

---

### 4. Metrics Collection Logic

#### 4.1 User Metrics (metrics_collector.go)
```go
// Data Sources:
// - users table: total_users, active_users, admin_users, banned_users
// - Filter: created_at for new_users_today and last_7_days

func (c *MetricsCollector) CollectUserMetrics(ctx context.Context) (*UserMetrics, error) {
    // Query 1: Total users
    totalUsers, err := c.entClient.User.Query().Count(ctx)
    
    // Query 2: Active users (is_active = true)
    activeUsers, err := c.entClient.User.Query().
        Where(user.IsActive(true)).
        Count(ctx)
    
    // Query 3: Admin users (role = platform_admin)
    adminUsers, err := c.entClient.User.Query().
        Where(user.Role("platform_admin")).
        Count(ctx)
    
    // Query 4: Banned users (status = banned)
    bannedUsers, err := c.entClient.User.Query().
        Where(user.Status("banned")).
        Count(ctx)
    
    // Query 5: New users today (created_at >= today 00:00:00)
    today := time.Now().Truncate(24 * time.Hour)
    newUsersToday, err := c.entClient.User.Query().
        Where(user.CreatedAtGTE(today)).
        Count(ctx)
    
    // Query 6: Registrations last 7 days
    sevenDaysAgo := time.Now().AddDate(0, 0, -7)
    registrationsLast7Days, err := c.entClient.User.Query().
        Where(user.CreatedAtGTE(sevenDaysAgo)).
        Count(ctx)
    
    return &UserMetrics{
        TotalUsers:                totalUsers,
        ActiveUsers:               activeUsers,
        AdminUsers:                adminUsers,
        BannedUsers:               bannedUsers,
        NewUsersToday:             newUsersToday,
        UserRegistrationsLast7Days: registrationsLast7Days,
        Timestamp:                 time.Now(),
    }, nil
}
```

#### 4.2 System Metrics (metrics_collector.go)
```go
import (
    "runtime"
    "time"
    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/disk"
    "github.com/shirou/gopsutil/v3/mem"
)

var startTime = time.Now()

func (c *MetricsCollector) CollectSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
    // Uptime
    uptime := int64(time.Since(startTime).Seconds())
    
    // Memory (Go runtime)
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    memoryUsageMB := memStats.Alloc / 1024 / 1024
    
    // CPU Usage (gopsutil)
    cpuPercent, err := cpu.Percent(time.Second, false)
    if err != nil || len(cpuPercent) == 0 {
        cpuPercent = []float64{0}
    }
    
    // Disk Usage (gopsutil)
    diskStat, err := disk.Usage("/")
    diskUsagePercent := 0.0
    if err == nil {
        diskUsagePercent = diskStat.UsedPercent
    }
    
    // Goroutines
    goroutines := runtime.NumGoroutine()
    
    return &SystemMetrics{
        UptimeSeconds:    uptime,
        MemoryUsageMB:    memoryUsageMB,
        CPUUsagePercent:  cpuPercent[0],
        DiskUsagePercent: diskUsagePercent,
        Goroutines:       goroutines,
        Timestamp:        time.Now(),
    }, nil
}
```

#### 4.3 Auth Metrics (metrics_collector.go)
```go
// Data Source: audit_logs table (if Story 5.9 completed)
// Action types: "login_attempt", "login_success", "token_issued"

func (c *MetricsCollector) CollectAuthMetrics(ctx context.Context) (*AuthMetrics, error) {
    // If audit_logs not available, return zeros
    if c.entClient.AuditLog == nil {
        return &AuthMetrics{Timestamp: time.Now()}, nil
    }
    
    // Query login attempts
    loginAttempts, err := c.entClient.AuditLog.Query().
        Where(auditlog.Action("login_attempt")).
        Count(ctx)
    
    // Query successful logins
    loginSuccess, err := c.entClient.AuditLog.Query().
        Where(auditlog.Action("login_success")).
        Count(ctx)
    
    // Calculate success rate
    successRate := 0.0
    if loginAttempts > 0 {
        successRate = float64(loginSuccess) / float64(loginAttempts) * 100
    }
    
    // Token issued
    tokenIssued, err := c.entClient.AuditLog.Query().
        Where(auditlog.Action("token_issued")).
        Count(ctx)
    
    return &AuthMetrics{
        LoginAttemptsTotal:  int64(loginAttempts),
        LoginSuccessRate:    successRate,
        FailedLoginAttempts: int64(loginAttempts - loginSuccess),
        TokenIssuedTotal:    int64(tokenIssued),
        Timestamp:           time.Now(),
    }, nil
}
```

#### 4.4 Performance Metrics (metrics_collector.go)
```go
// Phase 1: Return stub data (future: middleware instrumentation)
func (c *MetricsCollector) CollectPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
    // TODO: Implement middleware to collect real data
    // For now, return placeholder values
    return &PerformanceMetrics{
        APIRequestsTotal:         0,
        APIResponseTimeP95:       0,
        APIErrorRate:             0,
        DatabaseQueryDurationAvg: 0,
        Timestamp:                time.Now(),
    }, nil
}
```

---

### 5. Caching Strategy (metrics_service.go)

```go
const (
    MetricsCacheTTL       = 5 * time.Minute
    MetricsCacheKeyPrefix = "metrics:"
)

// Cache keys
const (
    CacheKeyAllMetrics    = "metrics:all"
    CacheKeyUserMetrics   = "metrics:users"
    CacheKeySystemMetrics = "metrics:system"
    CacheKeyPerfMetrics   = "metrics:performance"
    CacheKeyAuthMetrics   = "metrics:auth"
)

func (s *MetricsService) GetUserMetrics(ctx context.Context) (*UserMetrics, error) {
    // Try cache first
    cached, err := s.redis.Get(ctx, CacheKeyUserMetrics).Result()
    if err == nil {
        var metrics UserMetrics
        if err := json.Unmarshal([]byte(cached), &metrics); err == nil {
            metrics.CacheHit = true
            return &metrics, nil
        }
    }
    
    // Cache miss - collect fresh data
    metrics, err := s.collector.CollectUserMetrics(ctx)
    if err != nil {
        return nil, errors.Wrap(err, errors.CodeMetricsUnavailable, "failed to collect user metrics")
    }
    
    // Store in cache
    data, _ := json.Marshal(metrics)
    s.redis.Set(ctx, CacheKeyUserMetrics, data, MetricsCacheTTL)
    
    metrics.CacheHit = false
    return metrics, nil
}
```

---

### 6. Error Codes (pkg/errors/codes.go)

```go
// Metrics error codes
const (
    CodeMetricsUnavailable  = "METRICS_UNAVAILABLE"
    CodeMetricsCacheError   = "METRICS_CACHE_ERROR"
    CodeMetricsInvalidQuery = "METRICS_INVALID_QUERY"
)

var messages = map[string]string{
    CodeMetricsUnavailable:  "Metrics temporarily unavailable",
    CodeMetricsCacheError:   "Metrics cache error",
    CodeMetricsInvalidQuery: "Invalid metrics query",
}
```

---

### 7. Routing (routes/router.go)

```go
// Admin routes (user management)
adminGroup := r.Group("/api/admin")
adminGroup.Use(middleware.RequirePlatformAdmin())
{
    adminGroup.GET("/users", adminHandlers.ListUsers)
    adminGroup.POST("/users", adminHandlers.CreateUser)
    // ... other admin CRUD operations
}

// Metrics routes (read-only resource, admin-only access)
metricsGroup := r.Group("/api/metrics")
metricsGroup.Use(middleware.RequirePlatformAdmin())
{
    metricsGroup.GET("", metricsHandlers.GetAll)
    metricsGroup.GET("/users", metricsHandlers.GetUsers)
    metricsGroup.GET("/system", metricsHandlers.GetSystem)
    metricsGroup.GET("/performance", metricsHandlers.GetPerformance)
}
```

**Key Design Principle**: 
- `/api/admin/*` for **management operations** (CRUD on users, projects, etc.)
- `/api/metrics/*` for **read-only resources** (system observability)
- Both use `RequirePlatformAdmin()` middleware for access control
- URL path represents resource type, not permission level

---

## Implementation Checklist

### Phase 1: Foundation (Day 1)
- [ ] Create `core/modules/metrics/` directory
- [ ] Create `collector.go` with data collection logic
- [ ] Create `service.go` with caching logic
- [ ] Create `types.go` with metrics DTOs
- [ ] Create `config.go` with constants
- [ ] Add error codes to `pkg/errors/codes.go`
- [ ] Install dependency: `go get github.com/shirou/gopsutil/v3`

### Phase 2: Endpoints (Day 2)
- [ ] Create `handler.go` with all metrics endpoints
- [ ] Implement `GetAll` handler
- [ ] Implement `GetUsers` handler
- [ ] Implement `GetSystem` handler
- [ ] Implement `GetPerformance` handler (stub)
- [ ] Register independent metrics routes in `router.go`

### Phase 3: Testing (Day 3)
- [ ] Unit tests for `metrics_collector.go`
- [ ] Unit tests for `metrics_service.go` (cache scenarios)
- [ ] Integration tests for endpoints
- [ ] Test permission enforcement (admin vs non-admin)
- [ ] Performance test (cache hit vs miss)

---

## Testing Scenarios

### Unit Tests

#### Test: CollectUserMetrics
```go
func TestCollectUserMetrics(t *testing.T) {
    // Setup test database with sample users
    // - 10 total users
    // - 8 active users
    // - 2 admin users
    // - 1 banned user
    // - 3 created today
    
    metrics, err := collector.CollectUserMetrics(ctx)
    
    assert.NoError(t, err)
    assert.Equal(t, 10, metrics.TotalUsers)
    assert.Equal(t, 8, metrics.ActiveUsers)
    assert.Equal(t, 2, metrics.AdminUsers)
    assert.Equal(t, 1, metrics.BannedUsers)
    assert.Equal(t, 3, metrics.NewUsersToday)
}
```

#### Test: Cache Hit/Miss
```go
func TestMetricsServiceCacheHit(t *testing.T) {
    // First call - cache miss
    metrics1, err := service.GetUserMetrics(ctx)
    assert.False(t, metrics1.CacheHit)
    
    // Second call - cache hit
    metrics2, err := service.GetUserMetrics(ctx)
    assert.True(t, metrics2.CacheHit)
    assert.Equal(t, metrics1.TotalUsers, metrics2.TotalUsers)
}
```

### Integration Tests

#### Test: Admin Access Only
```go
func TestMetricsEndpointRequiresAdmin(t *testing.T) {
    // Non-admin user
    resp := request(t, "GET", "/api/metrics", userToken)
    assert.Equal(t, 403, resp.StatusCode)
    assert.Contains(t, resp.Body, "PERM_ADMIN_REQUIRED")
    
    // Admin user
    resp = request(t, "GET", "/api/metrics", adminToken)
    assert.Equal(t, 200, resp.StatusCode)
}
```

---

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Cache Hit Response | < 50ms | P95 |
| Cache Miss Response | < 1s | P95 |
| Database Queries | ≤ 6 | Per user metrics call |
| Cache TTL | 5 min | Redis expiration |
| Concurrent Requests | 100+ | Load testing |

---

## Dependencies

### Go Packages
```bash
go get github.com/shirou/gopsutil/v3
```

### Database Indexes (Already Exists)
```sql
-- users table
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at);

-- audit_logs table (if Story 5.9 completed)
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
```

---

## Out of Scope

### Not Included in This Story
- ❌ Real-time performance metrics collection (requires middleware instrumentation)
- ❌ Historical trends and time-series data
- ❌ Prometheus exporter integration
- ❌ Custom alerting rules
- ❌ Metrics dashboard UI
- ❌ Cache invalidation on user events (accept 5-min delay)

### Future Enhancements
- Story 1.21: HTTP Middleware Instrumentation (for real-time performance metrics)
- Story 1.22: Prometheus Exporter Integration
- Epic 10: Admin Dashboard UI

---

## Migration Notes

### Database Changes
**None** - Uses existing tables (users, audit_logs)

### Configuration Changes
**None** - Uses existing Redis connection from Story 1.16

---

## Documentation Updates

### After Implementation
- [ ] Update [API Documentation](../../api.md) with metrics endpoints
- [ ] Add metrics examples to Swagger/OpenAPI spec
- [ ] Update [Observability Epic](../../epics/9-observability-epic.md) status

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Code reviewed and merged
- [ ] Unit tests pass (≥80% coverage)
- [ ] Integration tests pass
- [ ] API documentation updated
- [ ] Swagger spec updated
- [ ] Performance targets met
- [ ] Security validation complete (admin-only access)
- [ ] Story marked complete in sprint tracking

---

## Notes

**Architecture Rationale**:
1. **Why `/api/metrics` not `/api/admin/metrics`?**
   - URL path represents **resource type**, not **permission level**
   - `/api/admin/*` is for **management operations** (user CRUD, config changes)
   - `/api/metrics/*` is for **read-only resources** (observability data)
   - Permission control via middleware, not URL naming
   - Future extensibility: can add `/api/projects/:id/metrics` for project-level metrics

2. **Why independent `metrics` module?**
   - Metrics collection is distinct from user management logic
   - Avoids bloating the admin module with unrelated code
   - Follows single responsibility principle
   - Easier to test and maintain independently

**Design Decisions**:
1. **Cache-First Strategy**: Prioritizes response time over real-time accuracy (5-min delay acceptable)
2. **Stub Performance Metrics**: Phase 1 returns zeros; real implementation requires middleware (future story)
3. **Optional Auth Metrics**: Gracefully degrades if Story 5.9 not completed
4. **gopsutil Library**: Industry-standard for system metrics (CPU, disk)

**Security Considerations**:
- All endpoints protected by `RequirePlatformAdmin` middleware
- No PII exposed in metrics (aggregated counts only)
- Error responses follow unified error standard (Story 1.19)

**Monitoring Philosophy**:
- Focus on admin dashboard needs (internal use)
- Future integration with external monitoring tools (Prometheus) is a separate concern
- Keep it simple for MVP
