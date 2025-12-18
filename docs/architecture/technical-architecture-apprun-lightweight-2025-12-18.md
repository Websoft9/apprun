# 技术架构文档：apprun BaaS 平台（轻量级版本）

**文档版本：** 2.0 (轻量级架构)  
**创建日期：** 2025-12-18  
**架构师：** Root  
**评审人：** 待定  
**状态：** ✅ 架构设计完成，POC环境已准备就绪  

---

## 文档说明

本文档是 apprun 的**轻量级架构设计**，基于以下核心约束：
- ✅ 作为公司产品底层基座，需长期演进能力
- ✅ 小团队（<5人）快速交付，6-12个月时间窗口
- ✅ **轻量级部署**是核心要求（单机 512MB 可启动）
- ✅ 排除重量级 BaaS 方案（Supabase/Appwrite 部署过重）
- ✅ 技术改造而非商业验证，需要技术掌控力

**与 v1.0 架构的区别：**

- v1.0（2025-12-13）：重量级微服务架构，基于 Kubernetes + Kong + Temporal
- v2.0（本文档）：轻量级自建架构，单二进制部署 + 精选组件集成

---

## 1. 架构概述

### 1.1 架构愿景

apprun 采用**轻量级自建架构**，以**单二进制部署**和**按需组件**为核心设计理念，为企业提供简单、高效、可控的 BaaS 平台。

### 1.2 核心架构原则
- **轻量级优先**：最小化依赖，单二进制部署，资源占用低
- **模块化设计**：核心模块可独立启用/禁用，避免强制依赖
- **渐进式增强**：从最小可行架构开始，按需扩展
- **技术掌控**：核心模块自建，关键技术完全掌握
- **运维友好**：简单部署，易于调试，故障排查快速

### 1.3 技术栈总览（轻量级）

| 层次 | 技术选型 | 说明 | 资源占用 |
|------|---------|------|---------|
| **核心语言** | Go 1.21+ | 单二进制编译，高性能 | - |
| **Web框架** | Gin / Fiber | 轻量级HTTP框架 | ~5MB |
| **数据库** | PostgreSQL 15+ | 单一数据库，简化运维 | ~100MB |
| **数据层** | PostgREST 12+ | 自动生成 REST API | ~10MB |
| **API网关** | Traefik / 自建 | 轻量级路由与认证 | ~20MB |
| **对象存储** | MinIO（可选） | 轻量级文件存储 | ~50MB |
| **缓存** | 内嵌 Redis（可选） | 可选性能优化 | ~10MB |
| **监控** | 内嵌指标采集 | Prometheus 格式暴露 | ~5MB |
| **工作流** | Temporal（SQLite） | 成熟工作流引擎（POC已验证） | ~150MB |
| **实时推送** | Centrifugo（可选） | Go实现的轻量级方案 | ~15MB |

**最小部署资源：**
- CPU: 2核（推荐）
- 内存: 512MB（apprun-core） + 256MB（PostgreSQL） + 150MB（Temporal） + 50MB（Ory Kratos） + 60MB（其他） = ~1GB
- 磁盘: 200MB（程序） + 按需（数据）

**说明：** 相比纯自建增加约 260MB 内存，但换来成熟的工作流引擎（Temporal）和认证系统（Ory），节省 2-3个月开发时间。

---

## 2. 系统架构图

### 2.1 总体架构（轻量级）

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端层                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Web控制台    │  │  HTTP API    │  │  CLI工具     │          │
│  │ (可选)       │  │  (核心)      │  │  (可选)      │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│               apprun-core (单二进制 Go 程序)                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │            轻量级 API 网关层 (Gin/Fiber)                  │  │
│  │  • 路由管理  • JWT认证  • 限流  • CORS                   │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ▼                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    核心服务层                             │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │  │
│  │  │认证权限  │ │函数执行  │ │存储代理  │ │国际化    │   │  │
│  │  │Module    │ │Module    │ │Module    │ │Module    │   │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │  │
│  │  ┌─────────────────────────────────────────────────┐   │  │
│  │  │       工作流引擎 (事件驱动核心)                  │   │  │
│  │  │  • 事件总线  • 任务队列  • 状态管理            │   │  │
│  │  └─────────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      数据层（最小化）                             │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────┐     │
│  │  PostgreSQL      │  │  PostgREST       │  │ MinIO    │     │
│  │  (主数据库)      │  │  (API生成)       │  │ (可选)   │     │
│  └──────────────────┘  └──────────────────┘  └──────────┘     │
└─────────────────────────────────────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                 部署环境（灵活支持）                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  裸机/VM     │  │  Docker      │  │  K8s (可选)  │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 核心组件架构（apprun-core 内部）

```
apprun-core (单二进制 Go 程序)
│
├── cmd/
│   └── server/          # 主入口程序
│       └── main.go      # 启动引导
│
├── internal/
│   ├── gateway/         # 轻量级API网关
│   │   ├── router.go    # 路由管理
│   │   ├── auth.go      # JWT认证中间件
│   │   ├── ratelimit.go # 限流控制
│   │   └── cors.go      # 跨域处理
│   │
│   ├── auth/            # 认证与权限模块
│   │   ├── jwt.go       # JWT令牌管理
│   │   ├── rbac.go      # 基于角色的访问控制
│   │   ├── user.go      # 用户管理
│   │   └── session.go   # 会话管理
│   │
│   ├── datamodel/       # 数据模型模块
│   │   ├── schema.go    # 模型定义
│   │   ├── migration.go # 数据库迁移
│   │   └── postgrest.go # PostgREST集成
│   │
│   ├── function/        # 函数服务模块
│   │   ├── runtime.go   # 运行时管理(WASM/Docker)
│   │   ├── executor.go  # 函数执行器
│   │   └── deploy.go    # 函数部署
│   │
│   ├── workflow/        # 工作流引擎（核心差异化）
│   │   ├── eventbus.go  # 事件总线
│   │   ├── queue.go     # 任务队列
│   │   ├── executor.go  # 工作流执行
│   │   └── state.go     # 状态管理
│   │
│   ├── storage/         # 存储服务模块
│   │   ├── s3.go        # S3兼容接口
│   │   ├── minio.go     # MinIO集成
│   │   └── local.go     # 本地文件存储
│   │
│   ├── i18n/            # 国际化模块
│   │   ├── loader.go    # 语言包加载
│   │   └── translator.go# 翻译引擎
│   │
│   ├── realtime/        # 实时推送模块（可选）
│   │   ├── websocket.go # WebSocket服务
│   │   └── centrifugo.go# Centrifugo集成
│   │
│   └── observability/   # 可观测性模块
│       ├── metrics.go   # Prometheus指标
│       ├── logging.go   # 结构化日志
│       └── tracing.go   # 分布式追踪(可选)
│
└── pkg/
    ├── config/          # 配置管理
    ├── db/              # 数据库连接池
    └── utils/           # 通用工具
```

---

## 3. 核心模块设计

### 3.1 认证与权限模块（基于 Ory Kratos + Casbin）

**设计目标：**
- 企业级认证（OAuth2/OIDC）
- 灵活的 RBAC/ABAC 权限控制
- 支持多租户隔离

**技术方案：**

#### 3.1.1 认证层（Ory Kratos）

Ory Kratos 负责：
- 用户注册/登录
- 密码管理（重置、修改）
- 会话管理
- 多因素认证（MFA）
- OAuth2/OIDC 集成

```go
// apprun-core 与 Ory Kratos 集成
import (
    oryClient "github.com/ory/kratos-client-go"
)

type AuthService struct {
    kratosClient *oryClient.APIClient
}

// 验证会话
func (s *AuthService) ValidateSession(sessionToken string) (*Session, error) {
    session, resp, err := s.kratosClient.FrontendApi.ToSession(ctx).
        XSessionToken(sessionToken).
        Execute()
    if err != nil {
        return nil, err
    }
    
    return &Session{
        UserID:   session.Identity.Id,
        Email:    session.Identity.Traits["email"].(string),
        TenantID: session.Identity.Traits["tenant_id"].(string),
    }, nil
}
```

#### 3.1.2 授权层（Casbin）

Casbin 负责：
- RBAC 权限模型
- 动态权限检查
- 策略管理

```go
// Casbin 集成（嵌入式）
import (
    "github.com/casbin/casbin/v2"
    "github.com/casbin/casbin/v2/model"
    "github.com/casbin/gorm-adapter/v3"
)

type AuthzService struct {
    enforcer *casbin.Enforcer
}

// 权限检查
func (s *AuthzService) Enforce(user, resource, action string) (bool, error) {
    return s.enforcer.Enforce(user, resource, action)
}

// RBAC 模型定义（model.conf）
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

**权限策略示例：**
```csv
# policy.csv
p, admin, *, *
p, user, /api/v1/workflows, read
p, user, /api/v1/workflows, create
g, alice@example.com, admin
g, bob@example.com, user
```

**API设计：**
- `POST /auth/register` - 用户注册
- `POST /auth/login` - 用户登录（返回JWT）
- `POST /auth/refresh` - 刷新Token
- `POST /auth/logout` - 登出
- `GET /auth/me` - 获取当前用户信息

**存储设计：**
```sql
-- PostgreSQL schema
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id),
    role_id UUID REFERENCES roles(id),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID REFERENCES roles(id),
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    scope VARCHAR(50) NOT NULL
);
```

### 3.2 数据模型模块（基于PostgREST）

**设计目标：**
- 通过SQL定义数据模型
- 自动生成RESTful API
- 支持复杂查询和关联

**技术方案：**
集成PostgREST作为数据API层：

```
PostgreSQL (数据定义)
    ↓
PostgREST (自动生成API)
    ↓
apprun-core (认证+权限包装)
    ↓
客户端
```

**示例：创建数据模型**
```sql
-- 用户通过管理界面或CLI定义模型
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT NOW()
);

-- PostgREST自动生成API：
-- GET  /api/products          # 查询列表
-- GET  /api/products?id=eq.xxx# 单条查询
-- POST /api/products          # 创建
-- PATCH /api/products?id=eq.xxx # 更新
-- DELETE /api/products?id=eq.xxx # 删除
```

**apprun-core的角色：**
- 验证JWT Token
- 注入tenant_id（多租户隔离）
- 应用RBAC权限
- 转发请求到PostgREST
- 记录审计日志

### 3.3 工作流引擎（核心差异化模块）

**设计目标：**
- 轻量级事件驱动架构
- 支持异步任务编排
- 可视化工作流定义（Phase 3）

**核心组件：**

#### 3.3.1 事件总线（Event Bus）
```go
type Event struct {
    ID        string
    Type      string // 事件类型: user.created, order.paid
    Payload   map[string]interface{}
    Timestamp time.Time
    TenantID  string
}

type EventBus interface {
    Publish(event Event) error
    Subscribe(eventType string, handler EventHandler) error
    Unsubscribe(eventType string, handler EventHandler) error
}

// 内存实现(MVP) + 持久化队列(Phase 2)
```

#### 3.3.2 工作流定义（YAML/JSON）
```yaml
# 示例：用户注册工作流
workflow:
  name: user_registration
  trigger:
    event: user.registered
  steps:
    - name: send_welcome_email
      type: function
      function: sendEmail
      params:
        template: welcome
        to: "{{event.user.email}}"
    
    - name: create_default_project
      type: http
      url: /api/projects
      method: POST
      body:
        name: "My First Project"
        owner_id: "{{event.user.id}}"
    
    - name: log_registration
      type: log
      message: "User {{event.user.email}} registered"
```

#### 3.3.3 任务队列
```go
type Task struct {
    ID          string
    WorkflowID  string
    StepName    string
    Status      string // pending, running, completed, failed
    Payload     map[string]interface{}
    RetryCount  int
    MaxRetries  int
    CreatedAt   time.Time
    CompletedAt *time.Time
}

// 使用PostgreSQL作为队列存储（简单可靠）
// Phase 2可升级到Redis或专用消息队列
```

### 3.4 函数服务模块

**设计目标：**
- 支持多语言函数执行
- 轻量级运行时（优先WASM）
- 快速启动和执行

**技术方案选型：**

| 方案 | 优势 | 劣势 | 推荐场景 |
|------|------|------|---------|
| **WASM (推荐)** | 极轻量、秒级启动、沙箱安全 | 生态较新 | MVP首选 |
| **Docker** | 生态成熟、语言支持广 | 资源占用大 | Phase 2扩展 |
| **内嵌解释器** | 简单直接 | 语言受限 | 特定场景 |

**WASM实现方案：**
```go
import "github.com/wasmerio/wasmer-go/wasmer"

type FunctionRuntime struct {
    instance *wasmer.Instance
}

func (r *FunctionRuntime) Execute(input []byte) ([]byte, error) {
    // 调用WASM函数
    result, err := r.instance.Exports.GetFunction("handler")
    if err != nil {
        return nil, err
    }
    
    // 执行并获取结果
    output, err := result(input)
    return output, err
}
```

**函数定义：**
```javascript
// 用户编写的函数（支持多种语言编译到WASM）
export function handler(input) {
    // 业务逻辑
    const data = JSON.parse(input);
    const result = processData(data);
    return JSON.stringify(result);
}
```

### 3.5 存储服务模块

**设计目标：**
- S3兼容API
- 支持本地存储和MinIO
- 可扩展到云存储

**架构设计：**
```
客户端
  ↓
apprun-core (存储代理)
  ↓
┌─────────┬──────────┬──────────┐
│本地存储  │ MinIO   │ 云存储   │
│(开发)    │(生产)   │(可选)    │
└─────────┴──────────┴──────────┘
```

**API设计（S3兼容）：**
- `PUT /storage/{bucket}/{key}` - 上传文件
- `GET /storage/{bucket}/{key}` - 下载文件
- `DELETE /storage/{bucket}/{key}` - 删除文件
- `GET /storage/{bucket}?prefix=xxx` - 列出文件

### 3.6 实时推送模块（可选）

**设计目标：**
- WebSocket支持
- 房间/频道管理
- 轻量级实现

**技术方案：**

**方案A：内嵌WebSocket（MVP）**
```go
// 使用gorilla/websocket
type RealtimeServer struct {
    clients map[string]*Client
    rooms   map[string]*Room
}

func (s *RealtimeServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    // ...处理连接
}
```

**方案B：集成Centrifugo（Phase 2）**
- Go实现，轻量级（~15MB）
- 支持百万级连接
- 提供管理API

### 3.7 国际化模块

**设计目标：**
- 支持50+语言
- 动态语言切换
- 简单集成

**技术方案：**
```go
// 使用go-i18n
type I18n struct {
    bundles map[string]*i18n.Bundle
}

// 语言文件格式（JSON）
{
    "welcome": {
        "en": "Welcome to apprun",
        "zh": "欢迎使用 apprun",
        "ja": "apprun へようこそ"
    }
}

// API
// GET /i18n/messages?lang=zh
// POST /i18n/messages (管理员更新)
```

---

## 4. 数据架构设计

### 4.1 数据库选型

**单一数据库策略：PostgreSQL 15+**

理由：
- ✅ 关系型数据 + JSONB支持（替代MongoDB）
- ✅ 全文搜索（替代Elasticsearch基础场景）
- ✅ 丰富扩展（PostGIS、pg_cron等）
- ✅ 成熟稳定，运维简单

### 4.2 核心数据表设计

```sql
-- 租户表
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    plan VARCHAR(50), -- free, pro, enterprise
    created_at TIMESTAMP DEFAULT NOW()
);

-- 用户表（参见3.1）
CREATE TABLE users (...);

-- 工作流定义表
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    definition JSONB NOT NULL, -- YAML转JSON存储
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 工作流执行历史
CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID REFERENCES workflows(id),
    status VARCHAR(50), -- pending, running, completed, failed
    input JSONB,
    output JSONB,
    error TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- 函数定义表
CREATE TABLE functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    runtime VARCHAR(50), -- wasm, docker
    code TEXT, -- 函数代码或Docker镜像
    config JSONB, -- 环境变量、资源限制等
    created_at TIMESTAMP DEFAULT NOW()
);

-- 审计日志表
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    user_id UUID,
    action VARCHAR(100), -- create, update, delete
    resource_type VARCHAR(100), -- user, workflow, function
    resource_id UUID,
    details JSONB,
    ip_address INET,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 索引优化
CREATE INDEX idx_workflows_tenant ON workflows(tenant_id);
CREATE INDEX idx_executions_workflow ON workflow_executions(workflow_id);
CREATE INDEX idx_executions_status ON workflow_executions(status);
CREATE INDEX idx_audit_logs_tenant ON audit_logs(tenant_id, created_at DESC);
```

---

## 5. API设计

### 5.1 API分层

```
┌─────────────────────────────────────────┐
│  RESTful API (apprun-core提供)          │
│  /api/v1/*                              │
│  - 认证、权限、工作流、函数              │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│  数据API (PostgREST提供)                 │
│  /data/*                                │
│  - 自动生成的CRUD API                    │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│  管理API (apprun-core提供)               │
│  /admin/*                               │
│  - 系统配置、租户管理、监控              │
└─────────────────────────────────────────┘
```

### 5.2 核心API端点

**认证API：**
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
GET    /api/v1/auth/me
```

**工作流API：**
```
GET    /api/v1/workflows
POST   /api/v1/workflows
GET    /api/v1/workflows/:id
PUT    /api/v1/workflows/:id
DELETE /api/v1/workflows/:id
POST   /api/v1/workflows/:id/execute
GET    /api/v1/workflows/:id/executions
```

**函数API：**
```
GET    /api/v1/functions
POST   /api/v1/functions
GET    /api/v1/functions/:id
PUT    /api/v1/functions/:id
DELETE /api/v1/functions/:id
POST   /api/v1/functions/:id/invoke
```

**存储API：**
```
PUT    /api/v1/storage/:bucket/:key
GET    /api/v1/storage/:bucket/:key
DELETE /api/v1/storage/:bucket/:key
GET    /api/v1/storage/:bucket
```

**数据API（PostgREST）：**
```
GET    /data/:table
POST   /data/:table
PATCH  /data/:table?id=eq.:id
DELETE /data/:table?id=eq.:id
```

---

## 6. 部署架构

### 6.1 最小部署（单机）

**适用场景：** 开发环境、小规模生产（<100用户）

```
┌────────────────────────────────────────┐
│         单台服务器 (2C4G)               │
│  ┌──────────────────────────────────┐  │
│  │  apprun-core (单二进制)           │  │
│  │  - 监听 :8080                    │  │
│  │  - 内存: ~512MB                  │  │
│  └──────────────────────────────────┘  │
│  ┌──────────────────────────────────┐  │
│  │  PostgreSQL 15                   │  │
│  │  - 监听 :5432                    │  │
│  │  - 内存: ~256MB                  │  │
│  └──────────────────────────────────┘  │
│  ┌──────────────────────────────────┐  │
│  │  PostgREST (可选独立进程)         │  │
│  │  - 监听 :3000                    │  │
│  │  - 内存: ~50MB                   │  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
```

**启动命令：**
```bash
# 1. 启动PostgreSQL
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=xxx postgres:15

# 2. 启动apprun-core
./apprun-core server \
  --db-url postgres://localhost:5432/apprun \
  --port 8080 \
  --enable-postgrest

# 单命令启动（内嵌PostgreSQL）- Phase 2特性
./apprun-core server --embedded-db --port 8080
```

### 6.2 标准部署（小规模生产）

**适用场景：** 100-1000用户，需要高可用

```
              ┌──────────────┐
              │ Load Balancer│ (Nginx/Traefik)
              └──────────────┘
                      │
        ┌─────────────┴─────────────┐
        │                           │
┌───────▼───────┐          ┌────────▼──────┐
│ apprun-core-1 │          │ apprun-core-2 │
│ (主节点)       │          │ (备节点)       │
└───────┬───────┘          └────────┬───────┘
        │                           │
        └─────────────┬─────────────┘
                      │
          ┌───────────▼────────────┐
          │  PostgreSQL (主从复制) │
          │  - Master: 读写        │
          │  - Replica: 只读       │
          └────────────────────────┘
          ┌────────────────────────┐
          │  MinIO (分布式存储)     │
          │  - 4节点集群            │
          └────────────────────────┘
```

### 6.3 Docker Compose 部署（更新版）

```yaml
version: '3.8'

services:
  # PostgreSQL 数据库
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: apprun
      POSTGRES_USER: apprun
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U apprun"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Ory Kratos - 认证服务
  kratos-migrate:
    image: oryd/kratos:latest
    environment:
      - DSN=postgres://apprun:${DB_PASSWORD}@postgres:5432/apprun?sslmode=disable
    command: migrate sql -e --yes
    depends_on:
      postgres:
        condition: service_healthy

  kratos:
    image: oryd/kratos:latest
    environment:
      - DSN=postgres://apprun:${DB_PASSWORD}@postgres:5432/apprun?sslmode=disable
      - SERVE_PUBLIC_BASE_URL=http://localhost:4433
      - SERVE_ADMIN_BASE_URL=http://localhost:4434
    ports:
      - "4433:4433"  # Public API
      - "4434:4434"  # Admin API
    volumes:
      - ./kratos:/etc/config/kratos
    command: serve -c /etc/config/kratos/kratos.yml --dev --watch-courier
    depends_on:
      - kratos-migrate
    restart: unless-stopped

  # Temporal - 工作流引擎
  temporal:
    image: temporalio/auto-setup:latest
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=apprun
      - POSTGRES_PWD=${DB_PASSWORD}
      - POSTGRES_SEEDS=postgres
      - DYNAMIC_CONFIG_FILE_PATH=config/dynamicconfig/development-sql.yaml
    ports:
      - "7233:7233"  # gRPC
      - "8233:8233"  # Web UI
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

  # apprun 核心服务
  apprun-core:
    image: apprun/core:latest
    environment:
      DATABASE_URL: postgres://apprun:${DB_PASSWORD}@postgres:5432/apprun
      KRATOS_PUBLIC_URL: http://kratos:4433
      KRATOS_ADMIN_URL: http://kratos:4434
      TEMPORAL_HOST: temporal:7233
      ENABLE_POSTGREST: "true"
      LOG_LEVEL: info
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - kratos
      - temporal
    restart: unless-stopped

  # MinIO 对象存储
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_PASSWORD}
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"
    restart: unless-stopped

volumes:
  postgres_data:
  minio_data:
```

**资源占用总览：**
- PostgreSQL: ~256MB
- Ory Kratos: ~50MB
- Temporal: ~150MB
- apprun-core: ~512MB
- MinIO: ~50MB
- **总计：~1018MB (约 1GB)**

**启动命令：**
```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 设置密码

# 2. 启动所有服务
docker-compose up -d

# 3. 初始化 Kratos 身份架构
./scripts/init-kratos.sh

# 4. 验证服务状态
docker-compose ps
curl http://localhost:8080/health
```

### 6.4 Kubernetes 部署（可选 - Phase 3）

仅在需要大规模扩展时考虑（1000+用户）。

```yaml
# 简化的K8s部署示例
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apprun-core
spec:
  replicas: 3
  selector:
    matchLabels:
      app: apprun-core
  template:
    metadata:
      labels:
        app: apprun-core
    spec:
      containers:
      - name: apprun-core
        image: apprun/core:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: apprun-secrets
              key: database-url
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "1000m"
            memory: "1Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

---

## 7. 安全架构

### 7.1 安全原则
- **零信任架构**：所有请求必须认证和授权
- **最小权限**：用户和服务仅获得必要权限
- **数据加密**：传输加密(TLS) + 存储加密(可选)
- **审计日志**：所有敏感操作记录审计

### 7.2 认证流程

```
1. 用户登录
   ↓
2. 验证凭据（bcrypt密码）
   ↓
3. 生成JWT Token
   - Access Token (15分钟有效期)
   - Refresh Token (7天有效期)
   ↓
4. 客户端携带Token访问API
   ↓
5. API网关验证Token
   - 解析JWT
   - 检查签名
   - 验证过期时间
   ↓
6. 提取用户信息和权限
   ↓
7. 执行RBAC权限检查
   ↓
8. 允许访问或拒绝(401/403)
```

### 7.3 多租户隔离

**数据隔离策略：**
```go
// 所有查询自动注入tenant_id
func (s *Service) GetRecords(ctx context.Context, userID string) ([]Record, error) {
    tenantID := ctx.Value("tenant_id").(string)
    
    // 自动添加租户过滤
    query := "SELECT * FROM records WHERE tenant_id = $1"
    return s.db.Query(query, tenantID)
}

// PostgreSQL Row Level Security (RLS)
ALTER TABLE records ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON records
    USING (tenant_id = current_setting('app.current_tenant')::uuid);
```

### 7.4 API安全

**限流策略：**
```go
// 基于Token Bucket算法
type RateLimiter struct {
    rate  int // 每秒请求数
    burst int // 突发容量
}

// 不同用户等级不同限流
// Free: 100 req/min
// Pro: 1000 req/min
// Enterprise: 10000 req/min
```

**输入验证：**
```go
// 使用go-playground/validator
type CreateWorkflowRequest struct {
    Name       string `json:"name" validate:"required,min=3,max=100"`
    Definition string `json:"definition" validate:"required,yaml"`
}
```

---

## 8. 监控与可观测性

### 8.1 指标采集（Prometheus格式）

**内嵌指标端点：**
```
GET /metrics

# 暴露指标
apprun_http_requests_total{method="GET", status="200"}
apprun_http_request_duration_seconds{handler="/api/v1/workflows"}
apprun_workflow_executions_total{status="completed"}
apprun_function_invocations_total{function="sendEmail"}
apprun_db_connections_active
apprun_memory_usage_bytes
```

### 8.2 日志设计

**结构化日志（JSON格式）：**
```json
{
  "timestamp": "2025-12-18T10:30:00Z",
  "level": "info",
  "tenant_id": "xxx",
  "user_id": "yyy",
  "request_id": "zzz",
  "method": "POST",
  "path": "/api/v1/workflows",
  "status": 201,
  "duration_ms": 45,
  "message": "Workflow created successfully"
}
```

**日志级别：**
- ERROR: 系统错误，需要立即处理
- WARN: 潜在问题，需要关注
- INFO: 正常业务日志
- DEBUG: 调试信息（开发环境）

### 8.3 健康检查

```go
// GET /health
{
  "status": "healthy",
  "version": "2.0.0",
  "checks": {
    "database": "ok",
    "postgrest": "ok",
    "storage": "ok"
  },
  "uptime_seconds": 3600
}

// GET /ready (Kubernetes readiness probe)
{
  "ready": true
}
```

---

## 9. 性能优化

### 9.1 缓存策略

**多级缓存：**
```
┌──────────────────┐
│  内存缓存 (热数据) │ <- 1ms
└──────────────────┘
        ↓ 未命中
┌──────────────────┐
│  Redis (温数据)   │ <- 10ms
└──────────────────┘
        ↓ 未命中
┌──────────────────┐
│  PostgreSQL       │ <- 50ms
└──────────────────┘
```

**缓存场景：**
- 用户权限信息（5分钟）
- 工作流定义（10分钟）
- 国际化文本（1小时）

### 9.2 数据库优化

**连接池配置：**
```go
db, err := sql.Open("postgres", dsn)
db.SetMaxOpenConns(25)      // 最大连接数
db.SetMaxIdleConns(5)       // 空闲连接数
db.SetConnMaxLifetime(5 * time.Minute)
```

**查询优化：**
- 使用索引覆盖常见查询
- 避免N+1查询（使用JOIN）
- 分页查询（limit/offset）
- 只查询需要的字段

### 9.3 并发处理

**Goroutine池：**
```go
// 限制并发数，避免资源耗尽
type WorkerPool struct {
    maxWorkers int
    tasks      chan Task
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.maxWorkers; i++ {
        go p.worker()
    }
}
```

---

## 10. 技术债务管理

### 10.1 已知限制（MVP阶段）

| 限制 | 影响 | 计划改进 |
|------|------|---------|
| 单机部署 | 无高可用 | Phase 2支持集群 |
| 内存队列 | 重启丢失 | Phase 2持久化队列 |
| 无分布式追踪 | 调试困难 | Phase 3集成Jaeger |
| 基础监控 | 功能有限 | Phase 3集成Grafana |
| 无CDN | 静态文件慢 | Phase 3集成CDN |

### 10.2 技术演进路线

**Phase 1 (MVP - 0-3个月)：**
- ✅ 单机部署
- ✅ 基础功能完整
- ✅ 内存队列
- ✅ 简单监控

**Phase 2 (优化 - 3-6个月)：**
- 🔄 主从复制（高可用）
- 🔄 Redis队列（持久化）
- 🔄 完善监控
- 🔄 性能优化

**Phase 3 (平台化 - 6-12个月)：**
- 🔄 Kubernetes支持
- 🔄 分布式追踪
- 🔄 可视化工作流
- 🔄 高级监控

---

## 11. 技术选型对比

### 11.1 编程语言决策

| 语言 | 优势 | 劣势 | 评分 | 推荐 |
|------|------|------|------|------|
| **Go** | 单二进制、高性能、并发强、生态好 | 泛型支持弱 | ★★★★★ | ✅ 首选 |
| Rust | 极致性能、内存安全 | 学习曲线陡、开发慢 | ★★★★☆ | Phase 3 |
| Node.js | 生态丰富、开发快 | 单二进制难、性能一般 | ★★★☆☆ | 不推荐 |

**最终决策：Go 1.21+**

### 11.2 认证与授权决策（核心决策）

| 方案 | 开发时间 | 资源占用 | 安全性 | 维护成本 | 评分 | 推荐 |
|------|---------|---------|--------|---------|------|------|
| **Ory Kratos + Casbin** | 1-2周 | ~50MB | ✅ 专业 | ✅ 低 | ★★★★★ | ✅ **首选** |
| 完全自建 | 6-8周 | ~20MB | ⚠️ 需验证 | ❌ 高 | ★★★☆☆ | Phase 2 可选 |
| Auth0/Keycloak | 1周 | ~200MB+ | ✅ 专业 | ✅ 低 | ★★★☆☆ | 太重 |

**最终决策：Ory Kratos（认证） + Casbin（授权）**

**理由：**
1. ✅ **时间窗口紧张**：节省 5-6周开发时间
2. ✅ **安全性保证**：Ory 是专业认证方案，久经考验
3. ✅ **OAuth2/OIDC**：开箱即用，无需自己实现复杂协议
4. ✅ **小团队友好**：维护成本低，社区支持好
5. ✅ **Casbin 灵活**：权限模型强大，嵌入式部署（0额外资源）
6. ⚠️ **50MB 成本**：相比自建增加 30MB，但极具性价比

**渐进式演进策略：**
- **MVP (0-3月)**：Ory + Casbin 快速上线
- **Phase 2 (3-6月)**：评估是否需要自建替换
- **Phase 3 (6-12月)**：根据实际需求决定（大多数情况无需替换）

**Casbin 优势（为何保留）：**
- 嵌入式 Go 库，无额外资源
- 支持 RBAC、ABAC、RESTful 等多种模型
- 性能优秀（纳秒级权限检查）
- 灵活的策略定义

### 11.3 API网关决策

| 方案 | 优势 | 劣势 | 评分 | 推荐 |
|------|------|------|------|------|
| **自建(Gin)** | 轻量、可控、无依赖 | 功能需要自己实现 | ★★★★★ | ✅ MVP首选 |
| Traefik | 轻量、配置简单 | 额外进程 | ★★★★☆ | Phase 2 |
| Caddy | 自动HTTPS、简单 | Go语言但独立进程 | ★★★★☆ | 可选 |
| Kong | 功能强大 | 太重（Nginx+Lua） | ★★☆☆☆ | 不推荐 |

**最终决策：MVP自建，Phase 2可考虑Traefik**

### 11.3 工作流引擎决策（已修订）

| 方案 | 优势 | 劣势 | 评分 | 推荐 |
|------|------|------|------|------|
| **Temporal** | POC已验证、功能强大、久经考验 | 需额外服务（~150MB） | ★★★★★ | ✅ **首选** |
| 自建事件总线 | 轻量、可控 | 需2-3个月开发、功能有限 | ★★★☆☆ | 不推荐 |
| Apache Airflow | 成熟、可视化 | Python、太重 | ★★☆☆☆ | 不推荐 |

**最终决策：Temporal（SQLite 模式轻量化部署）**

**理由：**
1. ✅ POC 已验证可行，技术风险低
2. ✅ 工作流是核心差异化，应使用成熟方案
3. ✅ 节省 2-3个月开发时间
4. ✅ 持久化、重试、补偿等特性开箱即用
5. ⚠️ 增加 ~150MB 内存，但性价比极高

---

## 12. 开发指南

### 12.1 项目结构

```
apprun/
├── cmd/
│   └── server/
│       └── main.go              # 主入口
├── internal/                    # 私有代码
│   ├── gateway/                 # API网关
│   ├── auth/                    # 认证模块
│   ├── datamodel/               # 数据模型
│   ├── function/                # 函数服务
│   ├── workflow/                # 工作流引擎
│   ├── storage/                 # 存储服务
│   ├── i18n/                    # 国际化
│   └── observability/           # 监控日志
├── pkg/                         # 公共库
│   ├── config/                  # 配置管理
│   ├── db/                      # 数据库
│   └── utils/                   # 工具函数
├── api/                         # API定义
│   └── openapi.yaml             # OpenAPI规范
├── web/                         # Web控制台
│   ├── src/
│   └── package.json
├── migrations/                  # 数据库迁移
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── docs/                        # 文档
├── scripts/                     # 脚本
│   ├── build.sh
│   └── deploy.sh
├── docker-compose.yml           # Docker部署
├── Dockerfile                   # 容器镜像
├── go.mod
├── go.sum
└── README.md
```

### 12.2 开发环境搭建

```bash
# 1. 克隆代码
git clone https://github.com/your-org/apprun.git
cd apprun

# 2. 启动依赖服务
docker-compose up -d postgres minio

# 3. 运行数据库迁移
make migrate-up

# 4. 启动开发服务
make dev

# 5. 运行测试
make test
```

### 12.3 编译部署

```bash
# 编译单二进制
make build
# 输出: ./bin/apprun-core

# 运行
./bin/apprun-core server \
  --config config.yaml \
  --port 8080

# 构建Docker镜像
make docker-build

# 部署到生产
make deploy-prod
```

---

## 13. 测试策略

### 13.1 测试金字塔

```
        ┌──────────┐
        │  E2E测试  │ <- 10%
        └──────────┘
      ┌──────────────┐
      │   集成测试    │ <- 30%
      └──────────────┘
    ┌──────────────────┐
    │     单元测试      │ <- 60%
    └──────────────────┘
```

### 13.2 测试覆盖率目标

- 单元测试：> 80%
- 集成测试：核心流程全覆盖
- E2E测试：关键用户场景

### 13.3 性能测试

**目标指标：**
- API响应时间: P95 < 200ms
- 并发支持: 1000+ RPS
- 数据库连接: < 50个
- 内存占用: < 1GB

**压测工具：**
```bash
# 使用k6进行压测
k6 run --vus 100 --duration 30s load-test.js
```

---

## 14. 风险与缓解

### 14.1 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| Go团队经验不足 | 中 | 中 | 培训 + Code Review |
| WASM生态不成熟 | 低 | 低 | 备选Docker方案 |
| 性能不达标 | 高 | 低 | 早期压测验证 |
| 数据库性能瓶颈 | 中 | 中 | 读写分离 + 缓存 |

### 14.2 运维风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 单机故障 | 高 | 中 | Phase 2集群部署 |
| 数据丢失 | 高 | 低 | 定时备份 + 灾备 |
| 监控盲区 | 中 | 中 | 完善监控指标 |
| 安全漏洞 | 高 | 低 | 安全审计 + 渗透测试 |

---

## 15. 总结

### 15.1 架构亮点

1. **轻量级设计**：单二进制部署，512MB内存可启动
2. **技术掌控**：核心模块自建，长期演进可控
3. **渐进式架构**：从简单开始，按需扩展
4. **运维友好**：部署简单，故障排查快速
5. **成本可控**：资源占用低，适合小团队

### 15.2 关键决策总结（已修订）

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 编程语言 | Go 1.21+ | 单二进制、高性能、并发强 |
| 数据库 | PostgreSQL 15+ | 单一数据库，JSONB支持 |
| 数据API | PostgREST 12+ | 自动生成API，成熟稳定 |
| **认证** | **Ory Kratos** | **企业级、OAuth2/OIDC、节省6-8周** |
| **授权** | **Casbin（嵌入）** | **灵活强大、零额外资源** |
| API网关 | 自建(Gin) | 轻量、可控、无依赖 |
| **工作流** | **Temporal（SQLite→PG）** | **POC已验证、可靠、节省2-3个月** |
| 函数运行时 | WASM优先 | 轻量、安全、快速启动 |
| 存储 | MinIO | 轻量、S3兼容 |
| 部署 | Docker Compose | 简单实用，K8s可选 |

**核心决策变更说明：**
1. ✅ **工作流**：从"自建"改为"Temporal"（POC已验证，成熟可靠）
2. ✅ **认证**：从"自建"改为"Ory Kratos"（节省开发时间，专业安全）
3. ✅ **授权**：明确使用"Casbin"（嵌入式，灵活强大）

**权衡分析：**
- 增加资源：~200MB（Kratos 50MB + Temporal 150MB）
- 节省时间：8-11周开发时间
- 提升质量：企业级认证 + 可靠工作流
- 降低风险：成熟方案，生产验证

**结论：** 在轻量级架构基础上，对核心复杂模块（认证+工作流）采用成熟方案，是小团队在紧张时间窗口下的最优选择。

### 15.3 下一步行动

1. **技术验证（1周）**
   - [ ] Go + PostgREST 集成验证
   - [ ] WASM 函数执行性能测试
   - [ ] 事件总线原型实现

2. **MVP开发（2-3个月）**
   - [ ] 核心模块开发
   - [ ] API实现
   - [ ] 基础测试

3. **Alpha测试（1个月）**
   - [ ] 内部试用
   - [ ] 性能调优
   - [ ] Bug修复

4. **Beta发布（1个月）**
   - [ ] 生产部署
   - [ ] 用户反馈收集
   - [ ] 迭代优化

---

**附录：**
- [A] 详细API文档（OpenAPI规范）
- [B] 数据库Schema完整定义
- [C] 部署运维手册
- [D] 安全审计清单

---

**文档维护：**
- 更新频率：每Sprint结束后更新
- 责任人：架构师 Root
- 评审周期：每月架构评审会

**文档历史：**
- 2025-12-18 v2.0: 轻量级架构设计（本文档）
- 2025-12-13 v1.0: 初始微服务架构（已归档）
