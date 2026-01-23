# Metrics Refactoring Summary - 2026-01-23

## 修复内容

### 1. ✅ 删除废弃的 API 端点
**问题**: 三个历史 Storage API 端点已废弃但代码未清理
- `GET /api/metrics/storage/health`
- `POST /api/metrics/storage/ingest`  
- `GET /api/metrics/storage/query`

**解决方案**:
- 删除 `core/modules/metrics/storage_handler.go` (300+ lines)
- 删除 `core/modules/metrics/storage_handler_test.go`
- 重新生成 swagger 文档（已自动移除废弃端点）

**当前 API** (Story 9.5):
- `GET /api/metrics/snapshot` - 聚合视图
- `GET /api/metrics/history` - 历史数据（替代 storage/query）
- `POST /api/metrics/ingest` - 批量摄入（替代 storage/ingest）
- `GET /api/metrics/keys` - 可用指标名
- `GET /api/metrics/scopes` - 可用范围

---

### 2. ✅ 完善 config.go 配置标签

#### pkg/metricstore/config.go
**添加的标签**:
```go
type StorageConfig struct {
    Backend   string        `mapstructure:"backend" default:"mock" db:"false" validate:"oneof=mock badger prometheus"`
    Timeout   time.Duration `mapstructure:"timeout" default:"5s" db:"true" validate:"gte=0"`
    Retention time.Duration `mapstructure:"retention" default:"24h" db:"true" validate:"gte=0"`
    // ...
}
```

#### modules/metrics/config.go
**新增配置结构体**:
```go
type Config struct {
    Collection CollectionConfig
    Cache      CacheConfig
    Export     ExportConfig
    RateLimit  RateLimitConfig
}
```

**标签规范** (遵循 [docs/standards/coding-standards.md](../../docs/standards/coding-standards.md)):
- `mapstructure:"field_name"` - YAML 映射键（snake_case）
- `default:"value"` - 默认值文档
- `db:"true|false"` - 是否可动态配置
  - `db:"false"` - 基础设施配置（Backend, KeyPrefix）
  - `db:"true"` - 业务配置（Timeout, Retention, Interval）
- `validate:"rules"` - 验证规则

---

### 3. ✅ 配置结构体职责分离

**问题**: `metricstore` 和 `metrics` 配置项定义可能重复

**解决方案**:

#### pkg/metricstore (存储层)
**职责**: WHERE & HOW - 数据持久化
```yaml
metricstore:
  storage:
    backend: badger        # 存储后端类型
    timeout: 5s            # 存储操作超时
    retention: 24h         # 数据保留期
    retry:                 # 重试策略
      enabled: true
      max_attempts: 3
```

#### modules/metrics (应用层)
**职责**: WHAT & WHEN - 指标收集和暴露
```yaml
metrics:
  collection:
    enabled: true
    interval: 15s          # 收集频率
    include_runtime: true  # 包含 Go 运行时指标
  cache:
    ttl: 1m               # 缓存过期时间
  export:
    path: "/metrics"      # Prometheus 端点
  rate_limit:
    requests: 100         # 速率限制
```

**关键区别**:
- `metricstore.storage.retention` - 存储层数据保留（删除旧数据）
- `metrics.cache.ttl` - 应用层缓存过期（性能优化）
- **无重复项**，职责清晰

---

### 4. ✅ 更新文档

#### 代码文档
- ✅ `core/config/metrics.yaml.example` - 精简配置示例
- ✅ `core/modules/metrics/config.go` - 新增 Config 结构体
- ✅ `core/pkg/metricstore/config.go` - 完善标签

#### Story 文档
- ✅ `docs/sprint-artifacts/sprint-3/9-3-badgerdb-otel.md`
  - 标记废弃 API 端点
  - 添加重构说明（参考 Story 9.5）

#### Epic 文档
- ✅ `docs/epics/9-observability-epic.md`
  - 更新 API 结构说明
  - 更新配置分离说明
  - 修正包路径（metricstoretore → metricstore）

#### 状态跟踪
- ✅ `docs/sprint-artifacts/sprint-status.yaml`
  - `9-2-metrics-storage-acl`: review → **done**
  - `9-3-badgerdb-otel`: ready-for-dev → **done**
  - 添加重构说明

---

## 架构验证

### ✅ 高内聚低耦合
```
pkg/metricstore/          modules/metrics/
    ├─ storage/               ├─ handler.go      (HTTP API)
    ├─ repository.go          ├─ service.go      (业务逻辑)
    └─ config.go              ├─ collector.go    (数据收集)
                              └─ config.go       (配置管理)
          ↓                           ↓
    存储层配置                   应用层配置
    (WHERE/HOW)                (WHAT/WHEN)
```

### ✅ 配置示例
```yaml
# 精简的三层结构（≤3 层）
metricstore:
  storage:                    # 层级1
    backend: badger           # 层级2
    retry:                    # 层级2
      enabled: true           # 层级3

metrics:
  collection:                 # 层级1
    interval: 15s             # 层级2
```

---

## 编译验证

```bash
$ cd core && go build -o bin/apprun-check ./cmd/
# ✅ 编译成功，无错误
```

---

## 改进点

### 已解决
1. ✅ 删除废弃 API 代码（300+ lines）
2. ✅ 配置标签完整性（default, db, validate）
3. ✅ 配置职责分离（无重复定义）
4. ✅ 文档同步更新（story, epic, status）

### 遗留常量
以下常量已标记 DEPRECATED，建议后续迁移到 Config：
```go
// core/modules/metrics/config.go
const (
    MetricsCacheTTL              = 1 * time.Minute  // → Config.Cache.TTL
    MetricsRateLimitRequests     = 100              // → Config.RateLimit.Requests
    MetricsStorageRetentionBadger = 24h            // → metricstore.Config.Retention
)
```

---

## 配置文件位置

- 📄 `core/config/metrics.yaml.example` - 完整配置示例
- 📄 `core/config/default.yaml` - 默认配置（可添加 metrics 段）
- 📁 `core/config/conf_d/` - 配置覆盖目录

---

## 相关 Story

- ✅ Story 9.1: Metrics Exposure
- ✅ Story 9.2: Metrics Storage Anti-Corruption Layer
- ✅ Story 9.3: BadgerDB Backend with OTEL Integration
- 📝 Story 9.4: Prometheus Backend Adapter (Backlog)

---

**修复完成时间**: 2026-01-23  
**修复范围**: API 清理 + 配置规范 + 文档同步  
**影响模块**: pkg/metricstore, modules/metrics, docs/
