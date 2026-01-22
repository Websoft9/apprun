# Story 9.1: Metrics Exposure
# Sprint 3: Observability & Monitoring (Epic 9)

**Priority**: P1  
**Effort**: 4 天 (Updated after SM review)  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 1.16 (Redis Cache Package) - ✅ Complete
- Story 5.7 (Platform User Management) - ✅ Complete
- **Story 9.2 (Metrics Storage ACL)** - ✅ Complete (核心依赖)
- **Story 9.3 (BadgerDB Backend)** - ✅ Complete (核心依赖)
- Story 5.9 (Audit Logging) - Optional (for auth metrics)

**Status**: ✅ Ready for Dev (基于 Storage 架构)  
**Updated**: 2026-01-21  
**Reviewed By**: Bob (Scrum Master)  
**Module**: Metrics  
**Epic**: [observability-epic](../../epics/9-observability-epic.md)  
**Related PRD**: [FR-MON-001](../../prd.md#fr-mon-001-system-monitoring--metrics)  
**Issue**: #TBD

---

## User Story

作为 **平台管理员 (platform_admin)**，我希望能通过 API 查看系统运行指标（实时 + 历史数据），以便监控平台健康状态、分析趋势并做出运维决策。

### 架构变化 (基于 Story 9.2 + 9.3)

**原设计**: 实时计算 API（每次请求查询数据库）  
**新设计**: **Storage 层查询** + 实时计算混合模式

- **历史指标**（trends）：从 BadgerDB 查询存储的时间序列数据
- **实时指标**（current）：实时计算数据库统计
- **性能提升**：减少数据库查询，利用持久化指标

---

## Acceptance Criteria

### 功能验收
- [ ] 实现 `GET /api/metrics` - 获取所有指标（实时快照）
- [ ] 实现 `GET /api/metrics/users` - 获取用户指标
- [ ] 实现 `GET /api/metrics/system` - 获取系统健康指标
- [ ] 实现 `GET /api/metrics/performance` - 获取性能指标
- [ ] 实现 `GET /api/metrics/history` - 获取历史趋势（复用 Storage Query）
- [ ] 支持时间范围查询 (`?duration=1h`, `?start=...&end=...`)
- [ ] 集成 `pkg/metricstore/Repository` 查询持久化指标

### Storage 集成验收（关键新增）
- [ ] 实时指标自动写入 BadgerDB（每次采集触发）
- [ ] 指标名称标准化定义：
  - `user_count_total` - 总用户数
  - `user_count_active` - 活跃用户数
  - `api_requests_total` - API 请求总数
  - `system_memory_mb` - 内存使用
  - `system_cpu_percent` - CPU 使用率
- [ ] Storage 写入延迟 < 10ms (p95)
- [ ] 历史查询验证：返回正确时间范围数据
- [ ] 降级测试：Storage 不可用时仅返回实时数据（不报错）

### 性能验收
- [ ] 实时指标响应时间 < 200ms（Redis 缓存）
- [ ] 历史指标响应时间 < 100ms（Storage 层查询）
- [ ] Redis 缓存 TTL = 1 分钟（平衡实时性与性能）
- [ ] Storage 层查询使用时间范围索引

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

### 1. Architecture

```
┌────────────────────────────────────────────────────┐
│         GET /api/metrics/* (统一 API)              │
└───────────────────┬────────────────────────────────┘
                    │
                    v
┌────────────────────────────────────────────────────┐
│  modules/metrics/handler.go (扩展 MetricsHandler)      │
│  ┌─────────────────────────────────────────────┐  │
│  │ 已存在 (Story 9.3)       │ 新增 (Story 9.1) │  │
│  │ - Ingest()    (POST)    │ - GetAll()        │  │
│  │ - Query()     (GET)     │ - GetUsers()      │  │
│  │ - Health()    (GET)     │ - GetSystem()     │  │
│  │                         │ - GetPerformance()│  │
│  └─────────────────────────────────────────────┘  │
└───────────────────┬────────────────────────────────┘
                    │
         ┌──────────┴───────────┐
         │                      │
         v                      v
┌──────────────────┐   ┌─────────────────────┐
│ collector.go     │   │ pkg/metricstore/Repo    │
│ (实时采集 + 写入) │──>│ (Storage 查询/写入) │
│ [NEW]            │   │ [Story 9.2]         │
└──────────────────┘   └──────────┬──────────┘
                                  │
                                  v
                        ┌──────────────────┐
                        │ BadgerDB         │
                        │ (时间序列数据)    │
                        │ [Story 9.3]      │
                        └──────────────────┘
```

**关键架构决策**:
1. **统一 API 路由**: `/api/metrics/*` (不再分离 `/api/metrics/storage/`)
2. **复用 `modules/metrics/`**: 扩展 Story 9.3 已创建的文件，不创建新模块
3. **自动持久化**: 实时采集时自动写入 Storage，供历史查询
4. **降级策略**: Storage 不可用时，仅返回实时数据（不报错）

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

#### 2.5 GET /api/metrics/history (NEW - 复用 Storage Query)
**描述**: 获取历史趋势数据（时间序列），复用 Story 9.3 的 Storage Query 逻辑  
**认证**: JWT + `platform_admin`  
**查询参数**: 
- `name` (required): 指标名称 (如 `user_count_total`, `api_requests_total`)
- `duration` (optional): 时间范围 (如 `1h`, `24h`, `7d`，默认 24h)
- `start` / `end` (optional): 精确时间范围 (RFC3339)
- `limit` (optional): 最大返回数据点数（默认 1000）

**示例请求**:
```
GET /api/metrics/history?name=user_count_total&duration=24h
```

**响应 (200 OK)**:
```json
{
  "success": true,
  "data": {
    "metrics": [
      {
        "name": "user_count_total",
        "value": 1200,
        "timestamp": "2026-01-20T00:00:00Z"
      },
      {
        "name": "user_count_total",
        "value": 1205,
        "timestamp": "2026-01-20T01:00:00Z"
      }
    ],
    "count": 24,
    "start": "2026-01-20T00:00:00Z",
    "end": "2026-01-21T00:00:00Z",
    "has_more": false
  }
}

**注意**: 此端点直接调用 `Repository.GetMetrics()`，与 Story 9.3 保持一致

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

#### 4.1 实时采集 + 自动持久化 (modules/metrics/collector.go - NEW)
```go
package obs

import (
    "context"
    "time"
    "apprun/pkg/metricstore"
)

type MetricsCollector struct {
    entClient *ent.Client
    repo      *metrics.Repository // Storage 层
}

// CollectUserMetrics - 实时计算 + 自动写入 Storage
func (c *MetricsCollector) CollectUserMetrics(ctx context.Context) (*UserMetrics, error) {
    // 1. 数据库实时查询
    totalUsers, _ := c.entClient.User.Query().Count(ctx)
    activeUsers, _ := c.entClient.User.Query().Where(user.IsActive(true)).Count(ctx)
    
    // 2. 异步持久化到 Storage（不阻塞响应）
    go c.persistMetrics(context.Background(), map[string]float64{
        "user_count_total":  float64(totalUsers),
        "user_count_active": float64(activeUsers),
    })
    
    return &UserMetrics{
        TotalUsers:  totalUsers,
        ActiveUsers: activeUsers,
        Timestamp:   time.Now(),
    }, nil
}

// persistMetrics - 批量写入 Storage
func (c *MetricsCollector) persistMetrics(ctx context.Context, data map[string]float64) {
    for name, value := range data {
        // 忽略错误，降级处理
        _ = c.repo.RecordMetric(ctx, metrics.Metric{
            Name:      name,
            Value:     value,
            Timestamp: time.Now(),
        })
    }
}
```

#### 4.2 历史查询 - 复用 Repository (modules/metrics/handler.go - 扩展)
```go
// GetHistory - 复用 Story 9.3 的 Query 逻辑
func (h *MetricsHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" {
        response.ErrorWithRequest(w, r, 400, "INVALID_REQUEST", "name is required")
        return
    }
    
    // 解析时间范围（默认 24h）
    duration := r.URL.Query().Get("duration")
    end := time.Now()
    start := end.Add(-parseDuration(duration, 24*time.Hour))
    
    // 查询 Storage 层（BadgerDB）
    results, err := h.repo.GetMetrics(r.Context(), name, start, end)
    if err != nil {
        // 降级：Storage 不可用，返回空数组而不是错误
        logger.Warn("Storage unavailable, returning empty history", "error", err)
        results = []metrics.Metric{}
    }
    
    response.SuccessWithRequest(w, r, map[string]interface{}{
        "metrics":  results,
        "count":    len(results),
        "start":    start,
        "end":      end,
        "has_more": false,
    })
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

### 5. Caching Strategy (modules/metrics/service.go - NEW)

```go
package obs

import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
)

const (
    MetricsCacheTTL       = 1 * time.Minute  // 1分钟：平衡实时性与性能
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
// Metrics API - 统一命名空间 (合并 Story 9.1 + 9.3)
metricsGroup := r.Group("/api/metrics")
metricsGroup.Use(middleware.RequirePlatformAdmin())
{
    // 实时指标暴露 (Story 9.1)
    metricsGroup.GET("", metricsHandler.GetAll)              // 所有指标快照
    metricsGroup.GET("/users", metricsHandler.GetUsers)      // 用户指标
    metricsGroup.GET("/system", metricsHandler.GetSystem)    // 系统指标
    metricsGroup.GET("/performance", metricsHandler.GetPerformance) // 性能指标
    metricsGroup.GET("/history", metricsHandler.GetHistory)  // 历史查询
    
    // Storage 操作 (Story 9.3 - 可选保留)
    metricsGroup.POST("", metricsHandler.Ingest)             // 手动写入指标
    metricsGroup.GET("/health", metricsHandler.Health)       // Storage 健康
}
```

**架构决策 - 统一 API 路由**:
- ✅ **单一命名空间**: `/api/metrics/*` 统一管理所有指标相关功能
- ✅ **RESTful 设计**: GET 查询，POST 写入，符合语义
- ✅ **向后兼容**: Story 9.3 的 Ingest/Health 端点保留（可选使用）
- ✅ **权限统一**: 全部使用 `RequirePlatformAdmin()` 中间件
- ❌ **废弃**: `/api/metrics/storage/*` 不再使用

---

## Implementation Checklist

### Day 1: Storage 集成 + 实时采集
- [ ] **复用** `modules/metrics/` 目录（不创建新 modules/metrics/）
- [ ] 新增 `modules/metrics/collector.go` - 实时采集逻辑
- [ ] 新增 `modules/metrics/service.go` - 缓存服务层
- [ ] 扩展 `modules/metrics/types.go` - 新增 UserMetrics, SystemMetrics 等类型
- [ ] 实现 Collector 方法 + 自动写入 Storage
- [ ] 单元测试：验证 Storage 自动写入（异步）
- [ ] 安装依赖: `go get github.com/shirou/gopsutil/v3`

### Day 2: 实时指标端点
- [ ] 扩展 `modules/metrics/handler.go` (不创建新文件)
- [ ] 实现 `GetAll()`, `GetUsers()`, `GetSystem()`, `GetPerformance()`
- [ ] 集成 Redis 缓存（TTL = 1 分钟）
- [ ] 集成 MetricsCollector 实时采集
- [ ] 更新 `routes/router.go` - 统一到 `/api/metrics/*`
- [ ] 单元测试：各端点逻辑 + 缓存命中/未命中

### Day 3: 历史查询 + 集成测试
- [ ] 实现 `GetHistory()` 端点（复用 Repository.GetMetrics）
- [ ] 实现降级策略：Storage 不可用时返回空数组
- [ ] 集成测试：实时 + 历史混合查询
- [ ] 集成测试：权限校验（admin vs non-admin）
- [ ] 集成测试：Storage 降级场景
- [ ] 性能测试：缓存效果验证

### Day 4: 故障测试 + 文档
- [ ] 并发测试：100+ 请求同时访问
- [ ] Storage 写入压力测试：验证异步写入不阻塞
- [ ] 端到端测试：采集 → 持久化 → 历史查询
- [ ] 更新 API 文档 (docs/api.md)
- [ ] 更新 Swagger spec (core/docs/swagger.yaml)
- [ ] 代码审查 + 合并

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

## Fault Tolerance & Degradation

### Storage 降级策略
```go
// 场景 1: Storage 写入失败
// 行为：忽略错误，继续返回实时数据（日志记录）
if err := repo.RecordMetric(ctx, metric); err != nil {
    logger.Warn("Failed to persist metric, continuing", "error", err)
    // 不影响实时查询响应
}

// 场景 2: Storage 查询失败 (历史数据)
// 行为：返回空数组，不报 500 错误
results, err := repo.GetMetrics(ctx, name, start, end)
if err != nil {
    logger.Warn("Storage unavailable, returning empty history", "error", err)
    results = []Metric{} // 降级为空数据
}
```

### Redis 降级策略
```go
// 场景：Redis 不可用
// 行为：跳过缓存，直接返回实时计算结果
cached, err := redis.Get(ctx, key).Result()
if err != nil {
    // Cache miss or Redis down - fallback to real-time
    return collector.CollectUserMetrics(ctx)
}
```

---

## Out of Scope

### Not Included in This Story
- ❌ Real-time performance metrics collection (requires middleware instrumentation)
- ❌ Prometheus exporter integration
- ❌ Custom alerting rules
- ❌ Metrics dashboard UI

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

1. **为什么统一到 `/api/metrics`？**
   - 原 Story 9.3 使用 `/api/metrics/storage/*` 太冗长
   - 统一命名空间更符合 RESTful 设计
   - 避免用户困惑（两套 API 路径）
   - 权限通过中间件控制，不依赖 URL 路径

2. **为什么复用 `modules/metrics/` 而不创建新模块？**
   - Story 9.3 已创建 `modules/metrics/` (Observability 模块)
   - 避免模块碎片化（metrics 属于 observability 子集）
   - 减少重复代码（共享 types, config）
   - 符合单一职责原则（Observability 统一管理）

3. **为什么缓存 TTL 选择 1 分钟？**
   - 5 分钟过长，实时性不足（用户看到过时数据）
   - 1 分钟平衡实时性与数据库压力
   - 管理员查看指标频率不高（非高频 API）

4. **为什么 Storage 写入采用异步？**
   - 避免阻塞实时指标响应（用户体验优先）
   - Storage 失败不影响实时查询功能
   - 降级策略：持久化是增强功能，不是核心功能

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
