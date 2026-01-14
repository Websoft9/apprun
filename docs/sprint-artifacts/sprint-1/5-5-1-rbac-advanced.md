# Story 5.5.1: RBAC Advanced Features (缓存、监控、性能优化)
# Sprint 2+: 认证与授权 (Auth Epic)

**Priority**: P1 (可选优化)  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 5.5 (RBAC Permissions) - 必须完成

**Status**: backlog  
**Module**: Authorization  
**Epic**: [auth-epic](../../epics/5-auth-epic.md)  
**Issue**: #TBD  

---

## User Story

作为 **apprun 平台运维人员和性能工程师**，我希望为 RBAC 系统添加高级特性，包括权限缓存、监控指标和性能优化，以支持高并发场景和生产环境运维需求。

---

## Acceptance Criteria

### 功能验收
- [ ] 实现权限结果缓存（基于 Redis + 内存两层）
- [ ] 实现缓存失效策略（TTL + 显式清除）
- [ ] 添加 policyVersion 机制（避免脏读）
- [ ] 实现自动 policy reload（文件监听 fsnotify）
- [ ] 支持定时 policy reload（可配置间隔）

### 非功能验收
- [ ] 权限检查延迟 P95 < 5ms（使用缓存）
- [ ] 单元测试覆盖率 ≥ 85%
- [ ] 性能测试：1000 并发请求无性能退化
- [ ] 缓存命中率 > 90%

### 监控验收
- [ ] Prometheus 指标：权限检查延迟（P50/P95/P99）
- [ ] Prometheus 指标：缓存命中/未命中次数
- [ ] Prometheus 指标：权限拒绝次数（按资源/动作分组）
- [ ] Prometheus 指标：Policy reload 成功/失败次数
- [ ] Grafana Dashboard 模板

### 可运维性
- [ ] 缓存统计 API（GET /api/admin/rbac/cache/stats）
- [ ] 手动清除缓存 API（POST /api/admin/rbac/cache/clear）
- [ ] Policy reload 日志（记录版本变更）
- [ ] 缓存降级机制（Redis 不可用时自动降级到无缓存）

---

## Technical Design

### 1. Permission Cache Architecture

```
┌─────────────────────────────────────┐
│  RequirePermission Middleware       │
├─────────────────────────────────────┤
│  1. 从缓存查询 (L1: sync.Map)      │
│  2. 未命中 -> 查询 L2 (Redis)      │
│  3. 未命中 -> 调用 Enforcer        │
│  4. 写入 L1 + L2 缓存             │
└─────────────────────────────────────┘
```

**Cache Key Format**:
```
perm:{policyVersion}:{userID}:{projectID}:{resource}:{action}
```

**TTL Strategy**:
- L1 (Memory): 5 minutes
- L2 (Redis): 10 minutes
- 失效触发：角色变更、Policy reload

### 2. Policy Version Management

```go
// core/internal/rbac/version.go
var (
    policyVersion     atomic.Int64
    policyUpdatedAt   atomic.Value // time.Time
)

func IncrementPolicyVersion() {
    policyVersion.Add(1)
    policyUpdatedAt.Store(time.Now())
}

func GetPolicyVersion() int64 {
    return policyVersion.Load()
}
```

### 3. Auto Reload Mechanisms

**文件监听（fsnotify）**:
```go
watcher.Add("config/casbin_policy.csv")
go func() {
    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                ReloadPolicies()
            }
        }
    }
}()
```

**定时轮询**:
```go
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        ReloadPolicies()
    }
}()
```

### 4. Prometheus Metrics

```go
var (
    permCheckDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "rbac_permission_check_duration_seconds",
            Buckets: []float64{.001, .005, .01, .025, .05},
        },
        []string{"resource", "action"},
    )
    
    permDeniedTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rbac_permission_denied_total",
        },
        []string{"resource", "action"},
    )
    
    cacheHitTotal = prometheus.NewCounter(...)
    cacheMissTotal = prometheus.NewCounter(...)
)
```

---

## Implementation Tasks

### Task 1: Cache Layer Implementation
- [ ] 实现两层缓存（sync.Map + Redis）
- [ ] 实现缓存 key 生成（带 policyVersion）
- [ ] 实现缓存失效逻辑
- [ ] 添加缓存降级机制

### Task 2: Policy Version & Reload
- [ ] 实现 policyVersion 管理
- [ ] 实现文件监听 reload
- [ ] 实现定时 reload
- [ ] 添加 reload 日志和指标

### Task 3: Monitoring & Metrics
- [ ] 添加 Prometheus 指标
- [ ] 实现缓存统计 API
- [ ] 创建 Grafana Dashboard 模板
- [ ] 添加性能测试脚本

### Task 4: Performance Testing
- [ ] 1000 并发测试
- [ ] P95 延迟测试
- [ ] 缓存命中率测试
- [ ] Redis 故障演练

---

## Performance Goals

| Metric | Target | Measurement |
|--------|--------|-------------|
| P50 延迟 | < 2ms | 使用缓存 |
| P95 延迟 | < 5ms | 使用缓存 |
| P99 延迟 | < 10ms | 使用缓存 |
| 缓存命中率 | > 90% | 稳定负载 |
| 1000 并发 QPS | > 5000 | 无性能退化 |

---

## References

- **Parent Story**: [5-5-rbac-permissions.md](./5-5-rbac-permissions.md)
- **Epic**: [auth-epic](../../epics/5-auth-epic.md)
