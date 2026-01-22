# API Spec

## 按 “领域 / 职责” 拆分 API 路径

将 CRUD 和操作类 API 拆分为不同的路径前缀，明确归属，保持领域模型纯粹：
```
# 1. 核心领域：服务配置的CRUD（静态数据）→ 归属于“配置域”
GET    /api/v1/services/{id}          # 查询服务配置（CRUD-查）
PUT    /api/v1/services/{id}          # 更新服务配置（CRUD-改）
POST   /api/v1/services               # 创建服务配置（CRUD-增）
DELETE /api/v1/services/{id}          # 删除服务配置（CRUD-删）

# 2. 操作域：服务的行为操作 → 归属于“操作/运维域”
POST   /api/v1/services/{id}/actions/restart  # 重启服务（操作类）
POST   /api/v1/services/{id}/actions/stop     # 停止服务（操作类）
POST   /api/v1/services/{id}/actions/upgrade  # 升级服务（操作类）
```

核心设计逻辑：

- 用/actions后缀标识 “操作类 API”，与 CRUD 路径形成明确区分；
- 操作类 API 统一用POST（符合 REST 语义：POST 用于 “触发有副作用的动作”）；
- 核心领域模型（如ServiceConfig结构体）仅承载配置数据，操作逻辑封装在独立的 “操作服务” 中（如ServiceOperationService）。
---

## Metrics API (平台指标与监控)

### 概述

Metrics API 提供平台监控指标的查询、上报和发现功能，支持实时快照和历史趋势分析。

**权限要求**: 所有 Metrics API 需要 `platform:metrics:read` 权限（仅平台管理员）。

**速率限制**: 100 请求/分钟

### 1. 核心数据接口

#### 1.1 获取指标快照

**端点**: `GET /api/metrics/snapshot`

**描述**: 获取指定范围的聚合指标快照（实时或近实时数据）

**查询参数**:
- `scope` (string, 可选): 指标范围
  - `system` - 系统资源指标（CPU、内存、磁盘）
  - `users` - 用户统计指标
  - `all` - 所有指标（默认值）

**响应示例** (`scope=all`):
```json
{
  "success": true,
  "data": {
    "scope": "all",
    "metrics": {
      "users": {
        "total_users": 1250,
        "active_users": 980,
        "admin_users": 15,
        "banned_users": 5,
        "new_users_today": 23,
        "user_registrations_last_7_days": 156,
        "timestamp": "2026-01-22T10:30:00Z"
      },
      "system": {
        "uptime_seconds": 864000,
        "memory_usage_mb": 2048,
        "cpu_usage_percent": 45.5,
        "disk_usage_percent": 65.2,
        "goroutines": 128,
        "timestamp": "2026-01-22T10:30:00Z"
      },
      "auth": {
        "login_attempts_total": 5420,
        "login_success_rate": 94.5,
        "failed_login_attempts": 298,
        "token_issued_total": 5122,
        "timestamp": "2026-01-22T10:30:00Z"
      },
      "performance": {
        "api_requests_total": 125000,
        "api_response_time_p95": 85,
        "api_error_rate": 0.5,
        "database_query_duration_avg": 12.3,
        "timestamp": "2026-01-22T10:30:00Z"
      }
    },
    "timestamp": "2026-01-22T10:30:00Z",
    "cache_hit": true
  }
}
```

**响应示例** (`scope=users`):
```json
{
  "success": true,
  "data": {
    "scope": "users",
    "metrics": {
      "users": {
        "total_users": 1250,
        "active_users": 980,
        "admin_users": 15,
        "banned_users": 5,
        "new_users_today": 23,
        "user_registrations_last_7_days": 156,
        "timestamp": "2026-01-22T10:30:00Z",
        "cache_hit": true
      }
    },
    "timestamp": "2026-01-22T10:30:00Z",
    "cache_hit": true
  }
}
```

**状态码**:
- `200 OK` - 成功
- `400 Bad Request` - 无效的 scope 参数
- `401 Unauthorized` - 未认证
- `403 Forbidden` - 无权限
- `500 Internal Server Error` - 服务器错误

**缓存**: 用户和认证指标缓存 5 分钟，系统指标实时查询

---

#### 1.2 获取历史趋势

**端点**: `GET /api/metrics/history`

**描述**: 查询单个指标的历史时序数据

**查询参数**:
- `name` (string, **必填**): 指标名称（如 `user_count_total`）
- `duration` (string, 可选): 时间范围（如 `1h`, `24h`, `7d`），默认 `24h`
- `start` (string, 可选): 开始时间（RFC3339 格式）
- `end` (string, 可选): 结束时间（RFC3339 格式）
- `limit` (int, 可选): 最大返回数量，默认 `1000`

**注意**: `duration` 和 `start/end` 二选一，优先使用 `start/end`。

**响应示例**:
```json
{
  "success": true,
  "data": {
    "metrics": [
      {
        "name": "user_count_total",
        "value": 1245,
        "tags": {"source": "system"},
        "timestamp": "2026-01-22T09:00:00Z"
      },
      {
        "name": "user_count_total",
        "value": 1248,
        "tags": {"source": "system"},
        "timestamp": "2026-01-22T10:00:00Z"
      },
      {
        "name": "user_count_total",
        "value": 1250,
        "tags": {"source": "system"},
        "timestamp": "2026-01-22T11:00:00Z"
      }
    ],
    "count": 3,
    "start": "2026-01-21T11:00:00Z",
    "end": "2026-01-22T11:00:00Z",
    "has_more": false
  }
}
```

**状态码**:
- `200 OK` - 成功（即使存储不可用，返回空数组）
- `400 Bad Request` - 缺少 name 参数或时间格式错误
- `401 Unauthorized` - 未认证
- `403 Forbidden` - 无权限

**降级策略**: 如果存储不可用，返回空数组而非错误，确保前端不中断。

---

#### 1.3 上报指标数据

**端点**: `POST /api/metrics/ingest`

**描述**: 批量上报指标数据到存储层

**权限**: 需要 `platform:metrics:write` 权限

**请求体** (JSON 数组):
```json
[
  {
    "name": "custom_metric_1",
    "value": 100.5,
    "tags": {
      "source": "user",
      "app": "myapp"
    },
    "timestamp": "2026-01-22T10:30:00Z"
  },
  {
    "name": "custom_metric_2",
    "value": 200,
    "tags": {
      "source": "user"
    }
  }
]
```

**请求字段**:
- `name` (string, **必填**): 指标名称（字母开头，仅允许字母数字下划线横杠点）
- `value` (float64, **必填**): 指标值
- `tags` (object, 可选): 标签键值对（最多 10 个）
- `timestamp` (string, 可选): 时间戳（RFC3339），默认当前时间

**限制**:
- 单次最多上报 1000 个指标
- 指标名称最长 256 字符
- 标签最多 10 个

**响应示例**:
```json
{
  "success": true,
  "data": {
    "success": true,
    "ingested": 2,
    "failed": 0,
    "failed_metrics": [],
    "message": "All metrics ingested successfully"
  }
}
```

**部分失败响应**:
```json
{
  "success": true,
  "data": {
    "success": false,
    "ingested": 1,
    "failed": 1,
    "failed_metrics": ["invalid_metric_name"],
    "message": "Partial success: some metrics failed to ingest"
  }
}
```

**状态码**:
- `200 OK` - 成功（包括部分失败）
- `400 Bad Request` - 请求格式错误
- `401 Unauthorized` - 未认证
- `403 Forbidden` - 无写权限
- `429 Too Many Requests` - 超过速率限制

---

### 2. 元数据 / 发现接口

#### 2.1 获取指标名录

**端点**: `GET /api/metrics/keys`

**描述**: 列出所有可用于历史查询的指标名称及元数据

**查询参数**:
- `source` (string, 可选): 过滤来源
  - `system` - 仅系统预置指标
  - `user` - 仅用户自定义指标（未来支持）

**响应示例**:
```json
{
  "success": true,
  "data": {
    "keys": [
      {
        "name": "user_count_total",
        "source": "system",
        "description": "Total user count",
        "unit": "count"
      },
      {
        "name": "user_count_active",
        "source": "system",
        "description": "Active user count",
        "unit": "count"
      },
      {
        "name": "system_cpu_percent",
        "source": "system",
        "description": "CPU usage",
        "unit": "percent"
      },
      {
        "name": "system_memory_mb",
        "source": "system",
        "description": "Memory usage",
        "unit": "megabytes"
      },
      {
        "name": "api_requests_total",
        "source": "system",
        "description": "Total API requests",
        "unit": "count"
      }
    ],
    "count": 15
  }
}
```

**使用场景**: 前端动态生成图表选择器，用户选择指标后调用 `/history` 查询数据。

**状态码**:
- `200 OK` - 成功
- `401 Unauthorized` - 未认证
- `403 Forbidden` - 无权限

---

#### 2.2 获取快照范围目录

**端点**: `GET /api/metrics/scopes`

**描述**: 列出所有可用于快照查询的 scope 选项及其包含的指标

**响应示例**:
```json
{
  "success": true,
  "data": {
    "scopes": [
      {
        "scope": "system",
        "metrics": ["cpu", "memory", "disk", "goroutines", "uptime"],
        "description": "System resource metrics (CPU, memory, disk usage)"
      },
      {
        "scope": "users",
        "metrics": [
          "total_users",
          "active_users",
          "admin_users",
          "banned_users",
          "new_users_today",
          "registrations_last_7d"
        ],
        "description": "User-related statistics and demographics"
      },
      {
        "scope": "all",
        "metrics": ["users", "system", "auth", "performance"],
        "description": "Complete platform metrics snapshot (all categories)"
      }
    ],
    "count": 3
  }
}
```

**使用场景**: 前端动态生成仪表盘配置器，告知用户有哪些可用视图。

**状态码**:
- `200 OK` - 成功
- `401 Unauthorized` - 未认证
- `403 Forbidden` - 无权限

---

### 3. 系统预置指标列表

#### 3.1 用户指标 (User Metrics)

| 指标名称 | 描述 | 单位 |
|---------|------|-----|
| `user_count_total` | 总用户数 | count |
| `user_count_active` | 活跃用户数 | count |
| `user_count_admin` | 管理员用户数 | count |
| `user_count_banned` | 被封禁用户数 | count |

#### 3.2 系统指标 (System Metrics)

| 指标名称 | 描述 | 单位 |
|---------|------|-----|
| `system_cpu_percent` | CPU 使用率 | percent |
| `system_memory_mb` | 内存使用量 | megabytes |
| `system_disk_percent` | 磁盘使用率 | percent |
| `system_goroutines` | Goroutine 数量 | count |

#### 3.3 API 性能指标 (Performance Metrics)

| 指标名称 | 描述 | 单位 |
|---------|------|-----|
| `api_requests_total` | API 请求总数 | count |
| `api_response_time_p95` | API 响应时间 P95 | milliseconds |
| `api_error_rate` | API 错误率 | percent |

#### 3.4 认证指标 (Auth Metrics)

| 指标名称 | 描述 | 单位 |
|---------|------|-----|
| `auth_login_attempts` | 登录尝试次数 | count |
| `auth_login_success_rate` | 登录成功率 | percent |
| `auth_failed_logins` | 登录失败次数 | count |
| `auth_tokens_issued` | JWT Token 签发数 | count |

---

### 4. 前端集成示例

#### 4.1 仪表盘 - 显示实时快照

```javascript
// 获取所有指标快照
const response = await fetch('/api/metrics/snapshot?scope=all', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
const data = await response.json();

// 渲染用户指标
renderUserStats(data.data.metrics.users);

// 渲染系统指标
renderSystemMetrics(data.data.metrics.system);
```

#### 4.2 动态图表 - 查询历史趋势

```javascript
// 1. 获取可用指标列表
const keysResponse = await fetch('/api/metrics/keys');
const keys = await keysResponse.json();

// 2. 让用户选择指标
const selectedMetric = 'user_count_total';

// 3. 查询该指标的 24 小时历史数据
const historyResponse = await fetch(
  `/api/metrics/history?name=${selectedMetric}&duration=24h`
);
const history = await historyResponse.json();

// 4. 使用 Chart.js 等库绘制趋势图
renderLineChart(history.data.metrics);
```

#### 4.3 自定义指标上报

```javascript
// 批量上报自定义指标
await fetch('/api/metrics/ingest', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify([
    {
      name: 'app_user_actions',
      value: 156,
      tags: { action: 'click', page: 'dashboard' }
    },
    {
      name: 'app_load_time_ms',
      value: 234
    }
  ])
});
```

---

### 5. 最佳实践

#### 5.1 性能优化
- **使用缓存**: `/snapshot` 接口有 5 分钟缓存，高频轮询时利用 `cache_hit` 字段判断
- **限制历史查询范围**: `/history` 查询尽量使用较短的 duration（如 1h, 6h），避免查询超大数据集
- **批量上报**: `/ingest` 支持批量（最多 1000 个），合并上报减少请求

#### 5.2 错误处理
- **降级策略**: `/history` 在存储不可用时返回空数组，前端应优雅展示"暂无数据"
- **部分失败**: `/ingest` 可能部分成功，检查 `failed_metrics` 字段重试失败项

#### 5.3 监控集成
- **Kubernetes**: 使用 `/health` 端点（包含 database, cache, metrics storage 健康检查）作为 liveness 和 readiness probe
- **Prometheus**: 配置 blackbox_exporter 定期探测 `/snapshot` 和 `/history` 可用性
- **Grafana**: 基于 `/keys` 接口自动发现可用指标，构建动态仪表盘

**Health Check API**: 参见 [Story 1-20](./sprint-artifacts/sprint-3/1-20-health-check-api.md)

---

### 6. API 迁移指南

**旧 API (已废弃)**:
```
GET /api/metrics/          → 使用 /api/metrics/snapshot?scope=all
GET /api/metrics/users     → 使用 /api/metrics/snapshot?scope=users
GET /api/metrics/system    → 使用 /api/metrics/snapshot?scope=system
GET /api/metrics/performance → 数据包含在 /api/metrics/snapshot?scope=all 中
POST /api/metrics/storage/ingest → 使用 /api/metrics/ingest
```

**兼容性**: 旧 API 在内部仍可用（service 方法保留），但新开发应使用新端点。

---

### 7. 相关文档

- **Swagger 文档**: http://localhost:8080/api/docs/
- **重构说明**: [core/modules/metrics/REFACTORING.md](../core/modules/metrics/REFACTORING.md)
- **Epic 9 - Observability**: [docs/epics/9-observability-epic.md](./epics/9-observability-epic.md)
- **Story 9.1 - Metrics Exposure**: [docs/sprint-artifacts/sprint-3/9-1-metrics-exposure.md](./sprint-artifacts/sprint-3/9-1-metrics-exposure.md)