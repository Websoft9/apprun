# Metrics API Refactoring

## 变更概览

根据新的 API 设计规范重构了 metrics 模块，提供更清晰的 API 结构和更好的可扩展性。

## API 变更

### 1. 核心数据接口 (Core Data APIs)

| 旧接口 | 新接口 | 变更说明 |
|--------|--------|----------|
| `GET /api/metrics/` | `GET /api/metrics/snapshot?scope=all` | 统一到snapshot，默认scope=all |
| `GET /api/metrics/users` | `GET /api/metrics/snapshot?scope=users` | 统一到snapshot |
| `GET /api/metrics/system` | `GET /api/metrics/snapshot?scope=system` | 统一到snapshot |
| `GET /api/metrics/performance` | `GET /api/metrics/snapshot?scope=all` | 合并到all scope |
| `GET /api/metrics/history` | `GET /api/metrics/history` | **保持不变** |
| `POST /api/metrics/storage/ingest` | `POST /api/metrics/ingest` | **路径简化** |

### 2. 新增接口 (New APIs)

| 接口 | 说明 |
|------|------|
| `GET /api/metrics/keys` | 列出所有可用的指标名称（用于history查询） |
| `GET /api/metrics/scopes` | 列出所有可用的snapshot scope选项 |

### 3. 弃用接口 (Deprecated)

| 接口 | 说明 |
|------|------|
| `GET /api/metrics/storage/health` | 移至 `/health` (Story 1.20) |

## 新的数据类型

### SnapshotResponse
```json
{
  "scope": "all|users|system",
  "metrics": {
    "users": {...},
    "system": {...},
    "auth": {...},
    "performance": {...}
  },
  "timestamp": "2026-01-22T10:30:00Z",
  "cache_hit": true
}
```

### KeysResponse
```json
{
  "keys": [
    {
      "name": "user_count_total",
      "source": "system",
      "description": "Total user count",
      "unit": "count"
    }
  ],
  "count": 15
}
```

### ScopesResponse
```json
{
  "scopes": [
    {
      "scope": "system",
      "metrics": ["cpu", "memory", "disk", "goroutines", "uptime"],
      "description": "System resource metrics"
    },
    {
      "scope": "users",
      "metrics": ["total_users", "active_users", ...],
      "description": "User-related statistics"
    },
    {
      "scope": "all",
      "metrics": ["users", "system", "auth", "performance"],
      "description": "Complete platform metrics"
    }
  ],
  "count": 3
}
```

### IngestBatchResponse
```json
{
  "success": true,
  "ingested": 100,
  "failed": 0,
  "failed_metrics": [],
  "message": "All metrics ingested successfully"
}
```

## 迁移指南

### 前端迁移

**旧代码:**
```javascript
// 获取用户指标
fetch('/api/metrics/users')

// 获取系统指标
fetch('/api/metrics/system')

// 获取所有指标
fetch('/api/metrics/')
```

**新代码:**
```javascript
// 获取用户指标
fetch('/api/metrics/snapshot?scope=users')

// 获取系统指标
fetch('/api/metrics/snapshot?scope=system')

// 获取所有指标
fetch('/api/metrics/snapshot?scope=all') // 或不传scope参数
```

### 数据上报迁移

**旧代码:**
```javascript
// 单条上报
fetch('/api/metrics/storage/ingest', {
  method: 'POST',
  body: JSON.stringify({
    name: 'custom_metric',
    value: 100
  })
})
```

**新代码:**
```javascript
// 批量上报（推荐）
fetch('/api/metrics/ingest', {
  method: 'POST',
  body: JSON.stringify([
    {
      name: 'custom_metric_1',
      value: 100,
      tags: { source: 'user' }
    },
    {
      name: 'custom_metric_2',
      value: 200,
      tags: { source: 'user' }
    }
  ])
})
```

### 动态UI构建

**新功能 - 获取可用指标**:
```javascript
// 1. 先获取可用的scopes
const scopes = await fetch('/api/metrics/scopes').then(r => r.json())
// 返回: {scopes: [{scope: "users", description: "..."}, ...]}

// 2. 让用户选择要查看的scope
const selectedScope = 'users'

// 3. 获取该scope的快照
const snapshot = await fetch(`/api/metrics/snapshot?scope=${selectedScope}`)
```

**新功能 - 获取历史指标列表**:
```javascript
// 1. 获取所有可用的历史指标
const keys = await fetch('/api/metrics/keys').then(r => r.json())
// 返回: {keys: [{name: "user_count_total", source: "system", ...}, ...]}

// 2. 用户选择要查看的指标
const selectedMetric = 'user_count_total'

// 3. 查询该指标的历史数据
const history = await fetch(`/api/metrics/history?name=${selectedMetric}&duration=24h`)
```

## 向后兼容性

### 已移除的Handler方法
- `GetAll()` → 使用 `GetSnapshot(scope="all")`
- `GetUsers()` → 使用 `GetSnapshot(scope="users")`
- `GetSystem()` → 使用 `GetSnapshot(scope="system")`
- `GetPerformance()` → 数据包含在 `GetSnapshot(scope="all")` 中

### 内部Service方法保留
以下service方法仍然可用于内部调用：
- `GetUserMetrics()`
- `GetSystemMetrics()`
- `GetAuthMetrics()`
- `GetPerformanceMetrics()`
- `GetAllMetrics()`

## 优势

1. **更清晰的API结构**: snapshot/history/ingest 三个核心端点，职责明确
2. **可发现性**: 通过 /keys 和 /scopes 端点，前端可以动态构建UI
3. **扩展性**: 新增scope或metric时，前端无需硬编码
4. **批量操作**: ingest支持批量上报，减少HTTP请求
5. **来源标识**: keys接口明确标识预置(system)和自定义(user)指标

## 测试覆盖

所有测试已更新并通过：
- ✅ Snapshot API 测试
- ✅ Keys API 测试  
- ✅ Scopes API 测试
- ✅ Ingest batch 功能
- ✅ 缓存机制
- ✅ 权限控制
- ✅ 历史查询
- ✅ 错误处理

## 文件变更列表

### 修改的文件
- `core/modules/metrics/handler.go` - 重构handler方法
- `core/modules/metrics/service.go` - 添加新的service方法
- `core/modules/metrics/types.go` - 添加新的响应类型
- `core/modules/metrics/storage_handler.go` - 移除重复定义
- `core/modules/metrics/handler_test.go` - 更新测试
- `core/routes/router.go` - 更新路由配置

### 未修改的文件
- `core/modules/metrics/collector.go` - 保持不变
- `core/modules/metrics/metrics.go` - 保持不变（常量定义）

## 下一步

1. 更新 Swagger 文档: `make swagger`
2. 更新 API 文档: `docs/api.md`
3. 前端适配新API
4. 监控旧API的使用情况，逐步迁移

## 联系

如有问题，请联系 Platform Team。
