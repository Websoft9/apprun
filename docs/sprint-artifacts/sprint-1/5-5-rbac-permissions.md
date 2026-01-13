# Story 5.5: RBAC 权限控制
# Sprint 2: 认证与授权 (Auth Epic)

**Priority**: P0 (必需)  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 5.1 (User Registration) - ✅ 已完成
- Story 5.3 (JWT Middleware) - ✅ 已完成 (提供 user_id 到 Context)
- Story 3.2 (Configuration Center Foundation) - ✅ 已完成
- Story 1.14 (Database Package) - ✅ 已完成

**Status**: ready-for-dev  
**Module**: Authorization  
**Epic**: [auth-epic](../../epics/5-auth-epic.md)  
**Issue**: #TBD  

---

## User Story

作为 **apprun 平台开发者和管理员**，我希望实现完整的 RBAC (Role-Based Access Control) + Domain 权限控制体系。

**核心理念：虚拟资源中心 (Virtual Resource Center)**
本项目不采用物理多租户隔离（如独立数据库/Schema），而是基于 **Project** 作为一个逻辑边界（虚拟资源中心）。
- **Project = Domain**: Casbin 中的 Domain 概念映射为 Project。
- **Contextual Roles**: 用户在不同 Project 中拥有不同角色（如 Project A 的 Admin，Project B 的 Viewer）。
- **Logical Isolation**: 所有资源API和数据库查询必须严格携带 `project_id` 上下文，确保逻辑隔离。
- **Platform Resources**: 部分资源（如系统配置、全局用户管理）属于平台全局，不归属特定 Project，使用统一的 Platform Domain 进行管理。

我希望系统支持项目级权限隔离和细粒度资源访问控制，以便用户只能访问其有权限的资源，确保多租户环境下的数据安全和权限边界。Casbin 负责表级/功能级权限，行级/数据级权限由业务代码检查 Ownership。

---

## Acceptance Criteria

### 功能验收
- [ ] 集成 Casbin v2 权限引擎到项目中
- [ ] 实现项目成员管理（添加/移除成员、分配角色）
- [ ] 实现 `RequirePermission` 中间件（权限验证）
- [ ] 定义 RBAC 角色模型（Platform + Project 双层）
- [ ] 实现 Casbin Model 和 Policy 配置加载
- [ ] 支持动态权限检查（不需重启服务）
- [ ] 实现权限缓存机制（基于 sync.Map 或 Redis）
- [ ] 提供权限管理 API（查询用户权限、角色权限）

### 非功能验收
- [ ] 权限检查延迟 P95 < 5ms（使用缓存）
- [ ] 单元测试覆盖率 ≥ 85%
- [ ] 性能测试：1000 并发请求无性能退化
- [ ] API 文档完整（Swagger annotations）

### 安全验收
- [ ] 默认拒绝策略（未配置权限的资源拒绝访问）
- [ ] 项目间权限完全隔离（user A 不能访问 project B 资源）
- [ ] 权限日志记录（记录所有权限拒绝事件）
- [ ] 敏感操作需要二次验证（owner 角色权限）

### 质量提升
- [ ] 结构化日志（记录权限检查详情）
- [ ] 国际化支持（权限错误提示多语言）
- [ ] 统一错误处理（权限错误码标准化，使用 core/pkg/errors）
- [ ] 权限策略可通过配置文件动态加载
- [ ] Prometheus 监控指标（权限检查延迟、拒绝次数）

---

## Previous Story Context

### Story 5.3 已完成功能（MUST REUSE）

**Context Access Pattern:**
Story 5.3 使用私有类型 context key（避免冲突）：
```go
// core/internal/jwt/context.go
type contextKey struct{ name string }
var UserIDKey = &contextKey{"user_id"}

// Helper functions to use
func GetUserID(ctx context.Context) int64
func GetUsername(ctx context.Context) string
func GetEmail(ctx context.Context) string
```

**⚠️ CRITICAL:** 本 Story 必须使用 Story 5.3 的 helper functions，不要直接访问 Context：
```go
// ❌ WRONG - String-based context access
userID := r.Context().Value("user_id").(int64)

// ✅ CORRECT - Use Story 5.3's helper
userID := jwt.GetUserID(r.Context())
```

**JWT Package Location:** `core/internal/jwt/`

---

## Technical Design


### 资源分类原则 (Resource Classification Principles)

在 AppRun 的权限体系中，**万物皆资源**，但需要区分以下类别：

1.  **非资源 (Non-Resources)**:
    *   **登录/注册 (Login/Register)**: 这些是认证入口 (AuthN)，不属于 RBAC 管控。它们发生在权限检查之前，用于建立 `user_id` 上下文。
    *   **原因**: 权限检查需要知道“谁”，而登录正是用来验证“谁”的。

2.  **数据资源 (Data Resources)**:
    *   **例子**: `server` 的元数据 (IP、名称、规格) - CRUD 操作。
    *   **RBAC**: `resource = "server"`, `action = "read/update/delete"`。

3.  **操作资源 (Operational Resources)**:
    *   **例子**: `server` 的具体动作 (重启、SSH 连接)。
    *   **RBAC**: `resource = "server"`, `action = "reboot/connect"`。
    *   **原则**: 操作动作是独立的权限点，不等同于数据修改。

**实现影响**: 只有数据和操作资源需要 `RequirePermission` 中间件。登录路由应放在受保护区之外。

### 架构分层

```
┌──────────────────────────────────────────┐
│  HTTP Middleware Chain                    │
├──────────────────────────────────────────┤
│  1. AuthMiddleware (JWT)                  │  ← Story 5.3 提供 user_id
│  2. ProjectContextMiddleware             │  ← 提取 project_id
│  3. RequirePermission(resource, action)  │  ← 本 Story 实现
├──────────────────────────────────────────┤
│  Casbin Enforcer                          │  ← 权限引擎
│  - Model: RBAC with domains               │
│  - Policy: CSV/Database                   │
│  - Adapter: File/Database                 │
├──────────────────────────────────────────┤
│  Permission Cache Layer                   │  ← sync.Map (MVP) / Redis
├──────────────────────────────────────────┤
│  Database (Ent ORM)                       │
│  - projects table                         │
│  - project_members table                  │
│  - casbin_rule table (optional)           │
└──────────────────────────────────────────┘
```

### 目录结构

```
core/
├── internal/
│   ├── jwt/                     # Story 5.3 已实现
│   │   ├── jwt.go              # JWT 工具函数
│   │   └── context.go          # Context helpers (GetUserID 等)
│   ├── middleware/
│   │   ├── auth.go              # JWT 认证（Story 5.3）
│   │   ├── rbac.go              # ✨ 本 Story 新增：RBAC 中间件
│   │   └── project_context.go  # ✨ 提取 project_id
│   └── rbac/
│       ├── enforcer.go          # ✨ Casbin Enforcer 初始化
│       ├── policy.go            # ✨ 策略加载和管理
│       ├── cache.go             # ✨ 权限结果缓存
│       └── roles.go             # ✨ 角色定义常量
├── ent/schema/
│   ├── project.go               # ✨ 项目模型
│   ├── project_member.go        # ✨ 项目成员模型
│   └── casbin_rule.go           # ✨ Casbin 策略存储（可选）
└── config/
    ├── casbin_model.conf        # ✨ Casbin RBAC Model
    └── casbin_policy.csv        # ✨ 初始策略（可选）

modules/auth/
├── handlers/
│   ├── permission_handler.go    # ✨ 权限查询 API
│   └── project_member_handler.go # ✨ 项目成员管理 API
├── services/
│   ├── permission_service.go    # ✨ 权限业务逻辑
│   └── project_member_service.go # ✨ 成员管理业务逻辑
└── repository/
    ├── project_repo.go          # ✨ 项目数据访问
    └── project_member_repo.go   # ✨ 成员关系数据访问
```

### 核心数据模型

#### 1. Project Schema (Ent)

```go
// core/ent/schema/project.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "time"
    "github.com/google/uuid"
)

type Project struct {
    ent.Schema
}

func (Project) Fields() []ent.Field {
    return []ent.Field{
        field.Int64("id").Unique(),
        field.String("uuid").
            MaxLen(36).
            Unique().
            Immutable().
            DefaultFunc(func() string {
                return uuid.New().String()
            }).
            Comment("外部 API 使用的 UUID"),
        field.String("name").MaxLen(100).NotEmpty(),
        field.String("description").MaxLen(500).Optional(),
        field.Int64("owner_id").Comment("项目所有者 user_id"),
        field.Int8("status").Default(1).Comment("状态：0-禁用，1-启用，2-归档"),
        field.Time("created_at").Immutable().Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Project) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("owner", User.Type).
            Ref("owned_projects").
            Field("owner_id").
            Unique().
            Required(),
        edge.To("members", ProjectMember.Type),
    }
}

func (Project) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("uuid").Unique(),
        index.Fields("owner_id"),
        index.Fields("status"),
        index.Fields("created_at"),
    }
}
```

**⚠️ REQUIRED:** User schema 需要添加反向边：
```go
// core/ent/schema/user.go - ADD THESE EDGES
func (User) Edges() []ent.Edge {
    return []ent.Edge{
        // ... existing edges ...
        edge.To("owned_projects", Project.Type),      // ✨ 新增
        edge.To("project_memberships", ProjectMember.Type), // ✨ 新增
    }
}
```

#### 2. ProjectMember Schema (Ent)

```go
// core/ent/schema/project_member.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "time"
)

type ProjectMember struct {
    ent.Schema
}

func (ProjectMember) Fields() []ent.Field {
    return []ent.Field{
        field.Int64("id").
            Unique(),
        field.Int64("project_id"),
        field.Int64("user_id"),
        field.String("role").
            MaxLen(20).
            NotEmpty().
            Comment("角色：owner, admin, member, viewer"),
        field.Time("joined_at").
            Immutable().
            Default(time.Now),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (ProjectMember) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).
            Ref("members").
            Field("project_id").
            Unique().
            Required(),
        edge.From("user", User.Type).
            Ref("project_memberships").
            Field("user_id").
            Unique().
            Required(),
    }
}

func (ProjectMember) Indexes() []ent.Index {
    return []ent.Index{
        // 确保一个用户在一个项目中只有一个角色
        index.Fields("project_id", "user_id").Unique(),
        index.Fields("user_id"),
        index.Fields("role"),
    }
}
```

#### 3. CasbinRule Schema (可选 - Database Adapter)

```go
// core/ent/schema/casbin_rule.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type CasbinRule struct {
    ent.Schema
}

func (CasbinRule) Fields() []ent.Field {
    return []ent.Field{
        field.Int64("id"),
        field.String("ptype").MaxLen(100),  // p, g
        field.String("v0").MaxLen(100).Optional(),  // sub
        field.String("v1").MaxLen(100).Optional(),  // dom (project_id)
        field.String("v2").MaxLen(100).Optional(),  // obj (resource)
        field.String("v3").MaxLen(100).Optional(),  // act (action)
        field.String("v4").MaxLen(100).Optional(),
        field.String("v5").MaxLen(100).Optional(),
    }
}

func (CasbinRule) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("ptype"),
        index.Fields("v0"),
        index.Fields("v1"),
    }
}
```

---

## RBAC Permission Model Design

### Casbin Model Configuration

此模型采用 "Global Role Definitions, Local Assignments" 模式。角色带来的权限定义是全局统一的（如 "Owner" 在任何项目中都有相同的权限集），但用户的角色分配是项目隔离的。

```ini
# config/casbin_model.conf
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _, _
g2 = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
# 1. 项目内角色匹配: g(user, role, project_id) && role 拥有权限
# 2. 平台角色匹配: g2(user, role) && role 拥有权限 (忽略 domain)
m = g(r.sub, p.sub, r.dom) && r.obj == p.obj && r.act == p.act || g2(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

**Model 说明**：
- `r = sub, dom, obj, act`: 请求格式 (user_id, project_id, resource, action)
- `p = sub, obj, act`: 策略格式 (role/user, resource, action) - **注意**：策略中不包含 domain，因为"角色定义"是通用的
- `g = _, _, _`: 项目级角色绑定 (user, role, project_id)
- `g2 = _, _`: 平台级角色绑定 (user, global_role)

### Role Definitions

```go
// core/internal/rbac/roles.go
package rbac

// Platform-level roles (全局角色)
const (
    RolePlatformAdmin = "platform_admin"  // 平台管理员
    RolePlatformUser  = "platform_user"   // 普通用户
)

// Project-level roles (项目角色)
const (
    RoleProjectOwner  = "owner"   // 项目所有者
    RoleProjectAdmin  = "admin"   // 项目管理员
    RoleProjectMember = "member"  // 项目成员
    RoleProjectViewer = "viewer"  // 查看者
)

// Resources (资源类型)
const (
    ResourceConfig    = "config"      // 配置管理
    ResourceData      = "data"        // 数据模型
    ResourceStorage   = "storage"     // 文件存储
    ResourceFunction  = "function"    // 函数服务
    ResourceWorkflow  = "workflow"    // 工作流
    ResourceMember    = "member"      // 成员管理
    ResourceProject   = "project"     // 项目设置
)

// Actions (操作类型)
const (
    ActionCreate = "create"
    ActionRead   = "read"
    ActionUpdate = "update"
    ActionDelete = "delete"
    ActionExecute = "execute"
    ActionManage = "manage"
)
```

### Default Policy Configuration

```csv
# config/casbin_policy.csv

# Platform-level policies (平台级权限)
# 平台管理员拥有无限权限，包括管理平台资源
p, platform_admin, *, *

# 也可以细化平台资源权限 (Example)
# p, platform_admin, system_setting, update
# p, platform_admin, license, read

# Project-level policies (项目级权限 - Role Definitions)

## Owner (所有权限)
p, owner, config, *
p, owner, data, *
p, owner, storage, *
p, owner, function, *
p, owner, workflow, *
p, owner, member, *
p, owner, project, *

## Admin (管理员：除删除项目外的所有权限)
p, admin, config, create
p, admin, config, read
p, admin, config, update
p, admin, config, delete
p, admin, data, *
p, admin, storage, *
p, admin, function, *
p, admin, workflow, *
p, admin, member, create
p, admin, member, read
p, admin, member, update
p, admin, member, delete
p, admin, project, read
p, admin, project, update

## Member (成员：读写数据和文件)
p, member, config, read
p, member, data, create
p, member, data, read
p, member, data, update
p, member, data, delete
p, member, storage, create
p, member, storage, read
p, member, storage, update
p, member, storage, delete
p, member, function, read
p, member, function, execute
p, member, workflow, read
p, member, workflow, execute

## Viewer (查看者：只读权限)
p, viewer, config, read
p, viewer, data, read
p, viewer, storage, read
p, viewer, function, read
p, viewer, workflow, read
```

---

## Implementation Details

### 1. Casbin Enforcer 初始化

**核心结构：**
```go
// core/internal/rbac/enforcer.go
package rbac

import (
    "github.com/casbin/casbin/v2"
    fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

type Config struct {
    ModelPath   string
    PolicyPath  string
    UseDatabase bool  // MVP: false (使用文件), Production: true (使用 DB)
}

// 全局单例
var (
    enforcer *casbin.Enforcer
    once     sync.Once
)
```

**关键函数：**
- `InitEnforcer(cfg Config) error` - 启动时初始化（加载 model + policy）
- `CheckPermission(userID, projectID int64, resource, action string) (bool, error)` - 权限检查（带缓存）
- `AddUserRole(userID, projectID int64, role string) error` - 分配角色
- `RemoveUserRole(userID, projectID int64, role string) error` - 移除角色
- `GetUserRoles(userID, projectID int64) ([]string, error)` - 查询角色

**实现要点：**
1. 使用 `sync.Once` 确保单例初始化
2. 请求格式：`(sub, dom, obj, act)` = `("user:123", "project:1", "config", "read")`
3. 权限结果缓存 5 分钟（见 cache.go）
4. 角色变更后清除缓存
5. 从 Config Center 加载 model/policy 路径（Story 3.2）

### 2. Permission Cache Layer

```go
// core/internal/rbac/cache.go
var permCache sync.Map  // MVP: 内存缓存，Production: Redis

// Cache key: "userID:projectID:resource:action"
// TTL: 5 minutes
// Invalidation: Role changes, policy updates
```

**函数：**
- `clearUserCache(userID, projectID int64)` - 清除用户缓存
- `ClearAllCache()` - 清除所有缓存（策略更新时）

### 3. RBAC Middleware

**核心模式（参考 Story 5.3 工厂模式）：**
```go
// core/internal/middleware/rbac.go
package middleware

const PlatformDomain = "platform" // ✨ 定义平台级 Domain 常量

// 1. ProjectContextMiddleware - 提取 project_id，验证成员身份
func ProjectContextMiddleware(next http.Handler) http.Handler {
    // ...existing code...
}

// 2. RequirePermission - 权限验证中间件工厂
func RequirePermission(resource, action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Get user_id using jwt.GetUserID(r.Context())
            
            // Determine Domain:
            // - If project_id is in context, use it.
            // - If not, assume Platform Domain (for /api/v1/platform/...) use "platform"
            projectID := GetProjectID(r.Context())
            domain := strconv.FormatInt(projectID, 10)
            if projectID == 0 { 
                domain = PlatformDomain 
            }
            
            // Call rbac.CheckPermission(userID, domain, resource, action)
            // ...
        })
    }
}
```

#### 4. Platform Resources Handling (New)

对于不归属于特定 Project 的平台级资源（如 `system_setting`, `license`），采用 **Platform Scope** 策略：

1.  **API 路由分离**:
    - **Project API**: `/api/v1/projects/{project_id}/...` (必须携带 `project_id`)
    - **Platform API**: `/api/v1/platform/...` (无 `project_id`，上下文默认为 `platform`)

2.  **鉴权逻辑**:
    - Casbin Request: `(user_id, "platform", resource, action)`
    - Model Logic: Matcher 中的 `g(r.sub, ...)` 即使失败，`g2(r.sub, p.sub)` 会检测用户的全局角色（如 `platform_admin`）。

3.  **Policy 配置**:
    - 需在 Policy 中明确允许全局角色访问平台资源：
      `p, platform_admin, system_setting, update`

**错误处理（使用 core/pkg/errors）：**
- `AUTH_REQUIRED` (401) - Missing user_id
- `PERM_NOT_MEMBER` (403) - Not project member  
- `PERM_FORBIDDEN` (403) - Permission denied
- `PERM_CHECK_ERROR` (500) - Enforcer error

#### 5. Private Resources Strategy (New)

针对“用户私有资源”，系统不引入新的鉴权模式，而是根据资源性质归类：

1.  **Identity Resources (身份资源)**
    *   **例子**: 用户资料(Profile)、密码修改、账号绑定。
    *   **归属**: **Platform Domain**。
    *   **鉴权**: Casbin 检查 `(user, "platform", "profile", "update")` + **代码层 ownership 检查** (`id == current_user_id`)。

2.  **Asset Resources (资产资源)**
    *   **例子**: 用户的个人 API Key、私有配置、个人沙盒环境。
    *   **归属**: **Personal Project** (统一模型)。
    *   **机制**: 用户注册时自动创建一个 "Personal Space" 项目，用户对该项目拥有 `Owner` 权限。所有“私有资产”均视为该项目下的资源。
    *   **优势**: 复用 Project RBAC 逻辑，无需特殊处理。

---

## API Endpoints

### 1. 项目成员管理 API

#### 1.1 添加项目成员
```http
POST /api/v1/projects/{project_id}/members
Authorization: Bearer <JWT>

Request Body:
{
  "user_id": 123,
  "role": "member"  // owner, admin, member, viewer
}

Response (201):
{
  "success": true,
  "data": {
    "id": 1,
    "project_id": 1,
    "user_id": 123,
    "role": "member",
    "joined_at": "2026-01-13T10:00:00Z"
  }
}
```

#### 1.2 查询项目成员列表
```http
GET /api/v1/projects/{project_id}/members
Authorization: Bearer <JWT>

Response (200):
{
  "success": true,
  "data": {
    "members": [
      {
        "id": 1,
        "user_id": 100,
        "username": "alice",
        "email": "alice@example.com",
        "role": "owner",
        "joined_at": "2026-01-01T10:00:00Z"
      },
      {
        "id": 2,
        "user_id": 123,
        "username": "bob",
        "email": "bob@example.com",
        "role": "member",
        "joined_at": "2026-01-13T10:00:00Z"
      }
    ],
    "total": 2
  }
}
```

#### 1.3 更新成员角色
```http
PUT /api/v1/projects/{project_id}/members/{member_id}
Authorization: Bearer <JWT>

Request Body:
{
  "role": "admin"
}

Response (200):
{
  "success": true,
  "data": {
    "id": 2,
    "role": "admin",
    "updated_at": "2026-01-13T11:00:00Z"
  }
}
```

#### 1.4 移除项目成员
```http
DELETE /api/v1/projects/{project_id}/members/{member_id}
Authorization: Bearer <JWT>

Response (200):
{
  "success": true,
  "message": "Member removed successfully"
}
```

### 2. 权限查询 API

#### 2.1 查询用户在项目中的权限
```http
GET /api/v1/projects/{project_id}/permissions/me
Authorization: Bearer <JWT>

Response (200):
{
  "success": true,
  "data": {
    "user_id": 123,
    "project_id": 1,
    "roles": ["member"],
    "permissions": [
      {"resource": "config", "actions": ["read"]},
      {"resource": "data", "actions": ["create", "read", "update", "delete"]},
      {"resource": "storage", "actions": ["create", "read", "update", "delete"]}
    ]
  }
}
```

#### 2.2 检查单个权限（用于前端按钮显示控制）
```http
GET /api/v1/projects/{project_id}/permissions/check?resource=data&action=delete
Authorization: Bearer <JWT>

Response (200):
{
  "success": true,
  "data": {
    "allowed": true
  }
}
```

---

## Middleware Usage Examples

### Example 1: 保护 Config API

```go
// modules/config/routes.go
package config

import (
    "github.com/go-chi/chi/v5"
    "apprun/core/internal/middleware"
)

func RegisterRoutes(r chi.Router) {
    r.Route("/api/v1/projects/{project_id}/configs", func(r chi.Router) {
        // 1. JWT 认证（提取 user_id）
        r.Use(middleware.AuthMiddleware)
        
        // 2. 项目上下文（提取 project_id，验证成员身份）
        r.Use(middleware.ProjectContextMiddleware)
        
        // 3. 权限验证
        r.With(middleware.RequirePermission("config", "read")).Get("/", ListConfigs)
        r.With(middleware.RequirePermission("config", "create")).Post("/", CreateConfig)
        r.With(middleware.RequirePermission("config", "update")).Put("/{id}", UpdateConfig)
        r.With(middleware.RequirePermission("config", "delete")).Delete("/{id}", DeleteConfig)
    })
}
```

### Example 2: 成员管理需要 admin 权限

```go
// modules/auth/routes.go
package auth

func RegisterRoutes(r chi.Router) {
    r.Route("/api/v1/projects/{project_id}/members", func(r chi.Router) {
        r.Use(middleware.AuthMiddleware)
        r.Use(middleware.ProjectContextMiddleware)
        
        // 普通成员可以查看成员列表
        r.With(middleware.RequirePermission("member", "read")).Get("/", ListMembers)
        
        // 只有 admin 和 owner 可以管理成员
        r.With(middleware.RequirePermission("member", "create")).Post("/", AddMember)
        r.With(middleware.RequirePermission("member", "update")).Put("/{id}", UpdateMember)
        r.With(middleware.RequirePermission("member", "delete")).Delete("/{id}", RemoveMember)
    })
}
```

---

## Testing Strategy

### Unit Tests (参考 Story 1.8 Testing Framework)

**测试文件：** `core/internal/rbac/enforcer_test.go`

**关键测试场景：**
```go
func TestCheckPermission_Owner(t *testing.T) {
    // Owner 应该有所有权限
    // 测试: config/data/storage 的 create/read/update/delete
}

func TestCheckPermission_Viewer(t *testing.T) {
    // Viewer 只有只读权限
    // 测试: read 成功, create/update/delete 失败
}

func TestProjectIsolation(t *testing.T) {
    // 用户不能访问其他项目资源
    // User 1 in Project 1 不能访问 Project 2
}

func TestRoleAssignment(t *testing.T) {
    // 测试角色分配和移除
    // AddUserRole, RemoveUserRole, GetUserRoles
}

func TestCacheInvalidation(t *testing.T) {
    // 角色变更后缓存应失效
}
```

### Integration Tests

**测试文件：** `modules/auth/handlers/member_handler_test.go`

**关键场景：**
- `TestAddProjectMember_Success` - 成功添加成员
- `TestAddProjectMember_Duplicate` - 重复添加应失败
- `TestUpdateMemberRole` - 更新角色
- `TestRemoveMember` - 移除成员
- `TestAccessControl_Forbidden` - 权限拒绝测试（viewer 尝试 create）
- `TestAccessControl_ProjectIsolation` - 跨项目访问拒绝

**Test Utilities (复用现有):**
- 使用 Story 1.8 的测试框架
- Database fixtures 和 cleanup 策略
- Mock Casbin enforcer for unit tests

---

## Monitoring & Observability

### Prometheus Metrics

```go
// core/internal/rbac/metrics.go
var (
    // 权限检查总数
    permCheckTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rbac_permission_check_total",
            Help: "Total number of permission checks",
        },
        []string{"allowed", "resource", "action"},
    )
    
    // 权限检查延迟
    permCheckDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "rbac_permission_check_duration_seconds",
            Help: "Permission check duration",
            Buckets: []float64{.001, .005, .01, .025, .05, .1},
        },
    )
    
    // 缓存命中率
    cacheHitTotal = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "rbac_cache_hit_total",
            Help: "Total cache hits",
        },
    )
    
    cacheMissTotal = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "rbac_cache_miss_total",
            Help: "Total cache misses",
        },
    )
)
```

### Audit Logging

```go
// 所有权限拒绝事件
logger.Warn("Permission denied",
    "user_id", userID,
    "project_id", projectID,
    "resource", resource,
    "action", action,
    "ip", r.RemoteAddr,
    "user_agent", r.UserAgent(),
)
```

---

## Performance Optimization

### 1. Permission Caching Strategy

```go
// MVP: In-memory cache with sync.Map
// - Cache permission check results for 5 minutes
// - Cache key: "user_id:project_id:resource:action"
// - Invalidate on role changes

// Production: Redis cache
// - TTL: 5 minutes
// - Pattern-based invalidation: "perm:user:{user_id}:project:{project_id}:*"
```

### 2. Batch Permission Loading

```go
// 在需要多次权限检查的场景，批量加载权限
func LoadUserPermissions(userID, projectID int64) map[string]bool {
    // 一次性加载该用户在该项目的所有权限
    // 缓存到 Context 中
}
```

### 3. Database Optimization

```sql
-- 为频繁查询的字段添加索引
CREATE INDEX idx_project_members_user_project ON project_members(user_id, project_id);
CREATE INDEX idx_project_members_role ON project_members(role);

-- 使用 covering index 优化权限查询
CREATE INDEX idx_project_members_covering ON project_members(user_id, project_id, role);
```

---

## Security Considerations

### 1. Default Deny Policy
- 所有未明确授权的资源默认拒绝访问
- 新增资源类型需要明确配置权限策略

### 2. Project Isolation
- 严格验证 project_id 参数，防止越权访问
- 每个请求必须验证用户是否是该项目成员

### 3. Owner Protection
- Project owner 不能被移除（除非转移所有权）
- Project owner 不能被降级角色

### 4. Audit Logging
```go
// 记录所有权限拒绝事件
logger.Warn("Permission denied",
    "user_id", userID,
    "project_id", projectID,
    "resource", resource,
    "action", action,
    "ip", r.RemoteAddr,
    "user_agent", r.UserAgent(),
)
```

---

## Migration Guide

### Database Migrations (参考 Story 1.5 Atlas Migration)

**迁移文件：** `core/migrations/{timestamp}_create_rbac_tables.sql`

**表结构要求：**

1. **projects** 表
   - UUID 字段用于外部 API（自动生成）
   - owner_id 外键关联 users(id)
   - 索引：uuid, owner_id, status

2. **project_members** 表
   - 唯一约束：(project_id, user_id)
   - role: owner, admin, member, viewer
   - 外键：project_id, user_id
   - 索引：(user_id, project_id) covering index

3. **casbin_rule** 表（可选 - Database Adapter）
   - 标准 Casbin 表结构
   - 索引：ptype, v0, v1

**Atlas 命令：**
```bash
# 生成迁移脚本（参考 Story 1.5）
atlas migrate diff create_rbac_tables \
  --dir "file://migrations" \
  --to "ent://ent/schema" \
  --dev-url "docker://postgres/15/dev"

# 应用迁移
atlas migrate apply \
  --dir "file://migrations" \
  --url "$DATABASE_URL"
```

### Ent Schema Generation

```bash
# 1. 创建 Schema
cd core
go run -mod=mod entgo.io/ent/cmd/ent new Project ProjectMember CasbinRule

# 2. 编辑 Schema（参考上面定义）
# 3. 生成代码
go generate ./ent

# 4. 更新 User Schema 添加反向边
# 5. 重新生成
go generate ./ent
```

---

## Dependencies

### Go Packages

```go
require (
    github.com/casbin/casbin/v2 v2.82.0
    github.com/google/uuid v1.6.0  // UUID generation (Story 5.1 已引入)
    // Optional: github.com/casbin/ent-adapter v0.5.0 (Database Adapter)
)
```

### Configuration Files

```bash
core/config/casbin_model.conf      # Casbin RBAC Model（必需）
core/config/casbin_policy.csv      # 初始权限策略（可选，可用代码加载）
```

### Configuration Center Integration (Story 3.2)

从 Config Center 加载配置路径：
```yaml
# config/rbac.yaml
rbac:
  model_path: "${CASBIN_MODEL_PATH:./config/casbin_model.conf}"
  policy_path: "${CASBIN_POLICY_PATH:./config/casbin_policy.csv}"
  use_database: false  # MVP: false, Production: true
  cache_ttl: 300       # 5 minutes
```

---

## Key Integration Points

### 1. Context Access Pattern (Story 5.3)
- ✅ MUST use `jwt.GetUserID(ctx)` NOT `ctx.Value("user_id")`
- ✅ Import from `core/internal/jwt`

### 2. Error Handling (Story 1.3)
- ✅ Use `core/pkg/errors` for error codes
- ✅ Use `core/pkg/response` for HTTP responses

### 3. Logging (Story 1.10)
- ✅ Use `core/pkg/logger` structured logging
- ✅ Log all permission denied events

### 4. Config Center (Story 3.2)
- ✅ Load Casbin paths from config
- ✅ Support dynamic policy reload

### 5. Database (Story 1.14)
- ✅ Use Ent ORM for schema definition
- ✅ Use Atlas for migrations (Story 1.5)

### 6. Testing (Story 1.8)
- ✅ Follow established testing framework
- ✅ Use test fixtures and cleanup utilities

---

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context will be added here by context workflow -->

### Agent Model Used

<!-- To be filled by dev agent -->

### Implementation Log

<!-- To be filled by dev agent during implementation -->

### Debug Log References

<!-- To be filled by dev agent -->

### Completion Notes List

**Core Implementation:**
- [ ] Casbin v2 集成（enforcer.go, cache.go, roles.go, policy.go）
- [ ] Ent Schemas 创建（Project, ProjectMember, CasbinRule）
- [ ] User Schema 反向边添加（owned_projects, project_memberships）
- [ ] RBAC 中间件（RequirePermission, ProjectContextMiddleware）
- [ ] 使用 Story 5.3 的 jwt.GetUserID() helper（不要直接访问 Context）
- [ ] 项目成员管理 API（handlers, services, repository）
- [ ] 权限查询 API（permissions/me, permissions/check）

**Integration:**
- [ ] Config Center 集成（加载 model/policy 路径）
- [ ] Error handling（使用 core/pkg/errors）
- [ ] Response formatting（使用 core/pkg/response）
- [ ] Structured logging（使用 core/pkg/logger）
- [ ] Prometheus metrics（permission check, cache hit/miss）

**Testing & Quality:**
- [ ] 单元测试通过（覆盖率 ≥ 85%）
- [ ] 集成测试通过（member management, access control）
- [ ] 项目隔离验证（跨项目访问拒绝）
- [ ] 性能测试（P95 < 5ms with cache）
- [ ] API 文档更新（Swagger annotations）

**Migration:**
- [ ] Atlas 迁移脚本生成和应用
- [ ] Casbin model/policy 配置文件创建

### File List

```
# Core RBAC Implementation
core/internal/rbac/enforcer.go
core/internal/rbac/cache.go
core/internal/rbac/policy.go
core/internal/rbac/roles.go
core/internal/rbac/metrics.go

# Middleware
core/internal/middleware/rbac.go
core/internal/middleware/project_context.go

# Ent Schemas
core/ent/schema/project.go
core/ent/schema/project_member.go
core/ent/schema/casbin_rule.go
core/ent/schema/user.go  # Update: add edges

# Business Logic
modules/auth/handlers/project_member_handler.go
modules/auth/handlers/permission_handler.go
modules/auth/services/project_member_service.go
modules/auth/services/permission_service.go
modules/auth/repository/project_repo.go
modules/auth/repository/project_member_repo.go
modules/auth/routes.go

# Configuration
core/config/casbin_model.conf
core/config/casbin_policy.csv
config/rbac.yaml  # Optional: Config Center integration

# Migration
core/migrations/{timestamp}_create_rbac_tables.sql

# Tests
core/internal/rbac/enforcer_test.go
core/internal/middleware/rbac_test.go
modules/auth/handlers/member_handler_test.go
modules/auth/handlers/permission_handler_test.go
```

---

## References

- **Epic**: [5-auth-epic.md](../../epics/5-auth-epic.md)
- **PRD**: [FR-AUTH-001 - Authentication & Authorization](../../prd.md#21-authentication--authorization)
- **Architecture**: [tech-architecture.md - 授权模块](../../architecture/tech-architecture.md#32-授权模块-rbac)
- **Story 5.3**: [5-3-jwt-middleware.md](./5-3-jwt-middleware.md) - Context access patterns
- **Story 3.2**: Configuration Center Foundation
- **Story 1.3**: Error Handling Framework
- **Story 1.5**: Database Migration with Atlas
- **Story 1.8**: Testing Framework
- **Story 1.10**: Logger Package
- **Story 1.14**: Database Package
- **Casbin Docs**: https://casbin.org/docs/overview
- **Ent Docs**: https://entgo.io/docs/getting-started

---

**Story Created By**: Bob (Scrum Master) - BMad SM Agent  
**Date**: 2026-01-13  
**Workflow**: create-story (YOLO mode)  
**Validation**: ✅ Enhanced with critical improvements  
**Optimization**: ✅ Streamlined for conciseness while keeping essential context  
**Ultimate Context Engine**: ✅ Complete
