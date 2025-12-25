# 技术架构文档
# apprun BaaS Platform

**创建日期**: 2025-12-25  
**架构师**: Winston (Architect Agent)  
**版本**: 1.1 (精简版)  
**状态**: Draft

---

## 1. 架构概览

### 1.1 架构风格

**模块化单体架构 (Modular Monolith)**

- 单进程部署，模块清晰分离
- 便于 MVP 快速交付
- 预留微服务演进路径

```
┌─────────────────────────────────────────────────────────────┐
│                      apprun BaaS Platform                    │
├─────────────────────────────────────────────────────────────┤
│  API Gateway (Chi Router + ReverseProxy)                     │
├──────────┬──────────┬──────────┬──────────┬─────────────────┤
│ Auth     │ Data     │ Storage  │ Functions│ Config          │
│ Module   │ Module   │ Module   │ Module   │ Module          │
├──────────┼──────────┼──────────┼──────────┼─────────────────┤
│ Workflow │ Events   │ Realtime │ I18N     │ License         │
│ Module   │ Module   │ Module   │ Module   │ Module          │
├──────────┴──────────┴──────────┴──────────┴─────────────────┤
│            Middleware (RBAC, Logging, Metrics)               │
├──────────────────────────────────────────────────────────────┤
│  Data Access Layer (Ent ORM + Repository Pattern)           │
└─────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
   PostgreSQL 14+        Redis 7+            Waterflow
   (主数据库)         (可选/事件)          (独立服务)
```

> **注**: 详细部署架构请参阅 [deployment-architecture.md](./deployment-architecture.md)

---

## 2. 技术栈

### 2.1 核心技术

| 层级 | 技术选型 | 版本 | 用途 |
|------|---------|------|------|
| **语言** | Go | 1.24+ | 主要开发语言 |
| **数据库** | PostgreSQL | 14+ | 主数据库 (ACID + JSON + 全文搜索) |
| **缓存** | Redis | 7+ | 可选/事件中心 (Streams) + L2 缓存 |
| **ORM** | Ent | latest | 类型安全的 ORM + 代码生成 |
| **Schema** | Atlas | latest | 声明式 Schema 管理和迁移 |
| **路由** | Chi | v5 | HTTP 路由 + 中间件 |
| **认证** | Ory Kratos | latest | 生产级认证服务 |
| **工作流** | Waterflow | latest | 基于 Temporal 的工作流引擎 |
| **WebSocket** | coder/websocket | latest | 实时推送 |
| **VFS** | spf13/afero | latest | 虚拟文件系统 (本地 + S3) |
| **配置** | Viper | v1 | 配置管理 + Watch |
| **授权** | Casbin | v2 | RBAC 策略引擎 |
| **监控** | Prometheus | latest | 指标采集 |
| **可视化** | Grafana | latest | 监控面板 |
| **容器** | Docker | 20.10+ | 容器化部署 |

### 2.2 开发工具

- **Linter**: golangci-lint 1.64.8
- **Security**: gosec 2.22.7+
- **Testing**: Testify
- **CI/CD**: GitHub Actions
- **文档**: Swagger/OpenAPI

---

## 3. 核心模块设计

### 3.1 认证模块 (Auth)

**集成方式**: Ory Kratos + 共享数据库

```go
// 共享数据库表 (只读)
- identities       // 用户身份信息
- identity_credentials  // 登录凭证
- sessions         // 会话管理

// apprun 自有表
- users            // 用户扩展信息
- user_projects    // 用户项目关系
```

**认证流程**:
1. 用户通过 Kratos 登录 → 生成 Session
2. apprun 读取 Kratos Session 验证身份
3. apprun 基于 `identity_id` 查询 RBAC 权限

**关键接口**:
- `GET /auth/whoami` - 获取当前用户信息
- `POST /auth/logout` - 退出登录

### 3.2 授权模块 (RBAC)

**模型**: Project-based RBAC + 平台全局权限

```go
// 权限模型
User → ProjectRole (per Project) → Permissions
User → GlobalRole (platform-level) → Permissions

// Casbin Policy 示例
p, admin, project:1, *, *           // Project 1 的 Admin
p, viewer, project:1, resource:*, read  // Project 1 的 Viewer
p, platform_admin, *, *, *          // 平台管理员
```

**中间件**:
```go
// 检查权限
func RequirePermission(resource, action string) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. 从 Session 获取 user_id
            // 2. 从 Context 获取 project_id
            // 3. Casbin.Enforce(user, project, resource, action)
            // 4. 通过则 next.ServeHTTP(w, r)
        })
    }
}
```

### 3.3 数据建模模块 (Data)

**DSL → Ent Schema 流程**:

```
用户定义 YAML/JSON
     ↓
解析并验证 DSL
     ↓
生成 Ent Schema (Go)
     ↓
go generate ./ent
     ↓
生成 CRUD 代码
     ↓
Atlas 生成迁移脚本
     ↓
应用数据库变更
```

**示例 DSL**:
```yaml
model:
  name: Product
  fields:
    - name: title
      type: string
      required: true
    - name: price
      type: decimal
      required: true
    - name: stock
      type: int
      default: 0
  relations:
    - name: category
      type: many-to-one
      target: Category
```

**生成的 Ent Schema**:
```go
// ent/schema/product.go
type Product struct {
    ent.Schema
}

func (Product) Fields() []ent.Field {
    return []ent.Field{
        field.String("title"),
        field.Float("price"),
        field.Int("stock").Default(0),
    }
}

func (Product) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("category", Category.Type).Ref("products").Unique(),
    }
}
```

### 3.4 文件存储模块 (Storage)

**虚拟文件系统 (VFS)**:

```go
// 抽象接口
type FileStorage interface {
    Upload(ctx context.Context, path string, reader io.Reader) error
    Download(ctx context.Context, path string) (io.Reader, error)
    Delete(ctx context.Context, path string) error
    List(ctx context.Context, prefix string) ([]FileInfo, error)
}

// 实现
- LocalStorage (基于 afero.OsFs)
- S3Storage (基于 afero.S3Fs)

// 使用
fs := afero.NewOsFs() // 本地优先
// fs := afero.NewS3Fs(...) // 可切换到 S3
```

**文件夹模拟**:
```go
// 数据库表
type Folder struct {
    ID       int
    Path     string  // /project-1/docs/
    Metadata JSON    // {"owner": "user_1"}
}

type File struct {
    ID       int
    Path     string  // /project-1/docs/file.pdf
    FolderID int
    Storage  string  // "local" or "s3"
}
```

### 3.5 函数服务模块 (Functions)

**执行方式**: 进程隔离

```go
// 函数定义
type Function struct {
    ID      int
    Name    string
    Code    string  // Go 代码
    Runtime string  // "go1.24"
    Trigger string  // "http" or "event"
}

// 执行流程
1. 编译函数代码 → 二进制文件
2. 启动独立进程执行
3. 通过 stdin/stdout 传递数据
4. 超时自动杀死进程
```

**HTTP 触发**:
```
POST /api/v1/functions/:name/invoke
{
  "input": {"key": "value"}
}
```

**事件触发**:
```go
// 订阅 Redis Streams
eventBus.Subscribe("user.created", func(event Event) {
    // 调用函数
    executeFunction("on-user-created", event.Data)
})
```

### 3.6 插件系统模块 (Plugins)

**gRPC 插件协议**:

```protobuf
// plugin.proto
service Plugin {
    rpc Execute(Request) returns (Response);
}

message Request {
    string operation = 1;  // "auth", "storage", "api"
    bytes payload = 2;
}
```

**插件类型**:
- **认证插件**: LDAP、OAuth、SAML
- **存储插件**: 新存储后端 (OSS、COS)
- **API 插件**: 中间件扩展
- **工作流插件**: 自定义节点

**加载流程**:
```go
// 启动时加载插件
plugins := []Plugin{
    LoadPlugin("plugins/ldap-auth"),
    LoadPlugin("plugins/oss-storage"),
}

// 调用插件
for _, plugin := range plugins {
    plugin.Execute(request)
}
```

### 3.7 工作流模块 (Workflow)

**集成 Waterflow**:

```
apprun                    Waterflow (独立服务)
   │                            │
   │  POST /api/workflows       │
   ├───────────────────────────>│
   │  {definition, trigger}     │
   │                            │ 存储工作流定义
   │                            │
   │  POST /api/workflows/:id/execute
   ├───────────────────────────>│
   │                            │ 执行工作流
   │                            │
   │  Webhook 回调              │
   │<────────────────────────────┤
   │  {status, result}          │
```

**工作流定义** (YAML):
```yaml
name: user-onboarding
trigger:
  type: event
  event: user.created
steps:
  - name: send-welcome-email
    type: function
    function: send-email
    params:
      template: welcome
  - name: create-default-project
    type: api
    api: POST /api/v1/projects
    params:
      name: "{{ user.name }}'s Project"
```

### 3.8 事件中心模块 (Events)

**Redis Streams 实现**:

```go
// 发布事件
eventBus.Publish("user.created", map[string]interface{}{
    "user_id": 123,
    "email": "alice@example.com",
})

// 订阅事件
eventBus.Subscribe("user.created", func(event Event) {
    log.Printf("New user: %v", event.Data)
})

// Redis Streams 操作
XADD events:user.created * user_id 123 email alice@example.com
XREAD BLOCK 1000 STREAMS events:user.created 0
```

**特性**:
- 短期持久化 (1-7 天)
- 分区内有序
- 消费者组支持

### 3.9 实时推送模块 (Realtime)

**WebSocket + CDC**:

```go
// WebSocket 连接管理
type ConnectionManager struct {
    connections map[string]*websocket.Conn
}

// ORM Hook (Ent)
func (u *User) AfterCreate(ctx context.Context) error {
    // 发送实时通知
    wsManager.Broadcast("user.created", u)
    return nil
}

// 客户端订阅
ws://apprun/ws?subscribe=user.created,project.updated
```

---

## 4. API 设计

### 4.1 路由规则

```
/api/v1/
├── auth/                 # 认证模块
│   ├── whoami
│   └── logout
├── users/                # 资源路由
├── projects/             # 资源路由
│   └── :id/
│       ├── models/       # 嵌套资源
│       └── members/
├── storage/              # 模块路由
│   ├── upload
│   └── download/:path
├── functions/            # 模块路由
│   └── :name/invoke
└── workflows/            # 模块路由
    └── :id/execute
```

> **注**: 详细数据架构（数据模型、缓存策略）请参阅 [data-architecture.md](./data-architecture.md)

### 4.2 统一响应格式

**成功响应**:
```json
{
  "success": true,
  "code": 200,
  "message": "操作成功",
  "data": {
    "user": {"id": 1, "name": "Alice"}
  }
}
```

**分页响应**:
```json
{
  "success": true,
  "code": 200,
  "data": {
    "items": [{"id": 1}, {"id": 2}],
    "pagination": {
      "total": 100,
      "page": 1,
      "pageSize": 10,
      "totalPages": 10
    }
  }
}
```

**错误响应**:
```json
{
  "success": false,
  "code": 400,
  "message": "参数错误",
  "error": {
    "code": "INVALID_PARAM",
    "details": "email 格式不正确"
  }
}
```

---

## 5. 安全架构

### 5.1 传输安全

- **HTTPS/TLS 1.3** - 所有 API 通信
- **CORS** - 跨域资源共享控制
- **CSRF Token** - 防跨站请求伪造

### 5.2 数据安全

- **密码**: bcrypt 加密
- **敏感数据**: AES-256 加密
- **密钥管理**: 环境变量

### 5.3 应用安全

- **SQL 注入**: Ent 参数化查询
- **XSS**: 输入过滤 + 输出转义
- **RBAC**: Casbin 策略引擎

---

## 6. 性能目标

### 6.1 响应时间

- API P95 < 100ms
- API P99 < 200ms

### 6.2 吞吐量

- QPS > 10,000 (单机)

### 6.3 优化策略

- 数据库索引优化
- 多层缓存 (L1+L2)
- 连接池 (DB、Redis)
- 异步处理 (事件队列)

---

## 7. 监控与日志

### 7.1 监控指标 (Prometheus)

```go
// 系统指标
- cpu_usage, memory_usage, disk_usage

// API 指标
- http_requests_total
- http_request_duration_seconds
- http_errors_total

// 业务指标
- user_count, project_count
- function_executions_total
```

### 7.2 日志方案

- **结构化日志**: JSON 格式
- **日志级别**: DEBUG, INFO, WARN, ERROR
- **日志存储**: 文件 + 简单查询 API
- **Trace ID**: 请求链路追踪 (MVP 不实现)

---

## 8. 演进路径

### 8.1 单体 → 微服务

```
当前 (MVP):
┌────────────────┐
│  apprun (单体) │
└────────────────┘

未来 (微服务):
┌──────┐ ┌──────┐ ┌──────┐
│ Auth │ │ Data │ │ Store│
└──────┘ └──────┘ └──────┘
```

**演进条件**:
- 单机性能瓶颈
- 团队规模扩大
- 模块独立部署需求

### 8.2 水平扩展

```
当前 (单机):
┌────────────────┐
│  apprun + DB   │
└────────────────┘

未来 (多实例):
    ┌─────────────┐
    │ Load Balancer│
    └─────────────┘
         │
    ┌────┴────┐
    │         │
┌───┴──┐  ┌──┴───┐
│apprun│  │apprun│
└──────┘  └──────┘
    │         │
    └────┬────┘
         │
    ┌────┴────┐
    │ PostgreSQL Cluster │
    │ Redis Cluster      │
    └───────────────────┘
```

---

## 9. 关键技术决策记录 (ADR)

### ADR-001: 选择模块化单体架构
- **决策**: MVP 使用模块化单体，保留微服务演进路径
- **理由**: 降低部署复杂度，加速 MVP 交付
- **权衡**: 牺牲部分独立扩展性

### ADR-002: 集成 Ory Kratos 而非自研认证
- **决策**: 集成 Ory Kratos + 共享数据库
- **理由**: 生产级安全性，节省开发周期
- **权衡**: 额外依赖，但风险可控

### ADR-003: 选择 Redis Streams 作为事件总线
- **决策**: 使用 Redis Streams，而非 Kafka/NATS
- **理由**: 轻量级，复用已有 Redis
- **权衡**: 功能相对简单，但 MVP 足够

### ADR-004: 虚拟文件系统 (VFS)
- **决策**: 使用 afero 抽象本地和 S3 存储
- **理由**: 云中立，灵活切换存储后端
- **权衡**: 略微增加抽象层复杂度

---

## 附录

### A. 参考资源

- [Ent Documentation](https://entgo.io/docs/getting-started)
- [Atlas Migration](https://atlasgo.io/)
- [Ory Kratos](https://www.ory.sh/docs/kratos/quickstart)
- [Chi Router](https://github.com/go-chi/chi)
- [Casbin](https://casbin.org/)
- [Waterflow](https://github.com/websoft9/waterflow)

### B. 术语表

- **VFS**: Virtual File System (虚拟文件系统)
- **CDC**: Change Data Capture (数据变更捕获)
- **RBAC**: Role-Based Access Control (基于角色的访问控制)
- **ADR**: Architecture Decision Record (架构决策记录)
- **DSL**: Domain Specific Language (领域特定语言)

---

**文档维护**: Winston (Architect Agent)  
**审核状态**: 待技术团队评审  
**下一步**: 创建部署架构文档 (deployment-architecture.md)
