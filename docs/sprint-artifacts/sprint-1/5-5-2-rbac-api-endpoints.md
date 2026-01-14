# Story 5.5.2: RBAC API Endpoints
# Sprint 2: 认证与授权 (Auth Epic)

**Priority**: P0 (必需)  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 5.5 (RBAC 权限控制) - ✅ 已完成核心基础设施
- Story 5.3 (JWT Middleware) - ✅ 已完成
- Story 4.1 (Swagger Docs) - ✅ 已完成

**Status**: not-started  
**Related Stories**: 
- Story 5.5 (RBAC 核心基础设施 - 已完成)
- Story 5.5.1 (RBAC 高级特性：缓存、监控、性能优化)

**Module**: Authorization  
**Epic**: [auth-epic](../../epics/5-auth-epic.md)  
**Issue**: #TBD  

---

## User Story

作为 **apprun 平台开发者和管理员**，我希望通过 RESTful API 管理项目成员和查询权限信息，以便：
- 在项目中添加/移除/更新成员角色
- 查询当前用户在项目中的权限
- 检查特定操作的权限（用于前端 UI 控制）
- 获取项目成员列表和详细信息

**背景**：Story 5.5 已完成 RBAC 核心基础设施（Casbin 集成、中间件、Service 层），本 Story 专注于提供 HTTP API 接口层。

---

## Acceptance Criteria

### 功能验收
- [ ] 实现项目成员管理 API（添加、列出、更新、删除成员）
- [ ] 实现权限查询 API（查询当前用户权限、检查特定权限）
- [ ] 实现 RBAC 策略重载 API（管理员功能）
- [ ] 所有 API 使用 Story 5.5 的中间件保护
- [ ] 请求/响应格式符合项目规范（pkg/response）

### 非功能验收
- [ ] API 文档完整（Swagger annotations）
- [ ] 集成测试覆盖所有端点（正常流程 + 异常流程）
- [ ] 错误处理统一（使用 pkg/errors）
- [ ] 结构化日志（使用 pkg/logger）
- [ ] 性能测试：响应时间 < 100ms (p95)

### 安全验收
- [ ] 成员管理 API 需要 admin 权限
- [ ] 项目隔离验证（不能操作其他项目成员）
- [ ] Owner 保护（不能移除或降级 owner）
- [ ] 输入验证（角色名称、用户 ID 格式）
- [ ] 审计日志（记录所有成员变更操作）

### 可维护性验收
- [ ] Handler 代码简洁（委托给 Service 层）
- [ ] 错误码统一（复用 pkg/errors）
- [ ] 测试覆盖率 > 80%
- [ ] API 版本化（/api/v1）

---

## Technical Design

### 目录结构

```
core/modules/auth/
├── handler/
│   ├── project_member_handler.go    # ✨ 新增：项目成员管理 API
│   ├── permission_handler.go        # ✨ 新增：权限查询 API
│   ├── rbac_admin_handler.go        # ✨ 新增：RBAC 管理 API
│   └── handler_test.go              # ✨ 新增：API 集成测试
├── service/
│   ├── project_member_service.go    # ✅ 已存在（Story 5.5）
│   └── permission_service.go        # ✅ 已存在（Story 5.5）
├── repository/
│   ├── project_repo.go              # ✅ 已存在（Story 5.5）
│   └── project_member_repo.go       # ✅ 已存在（Story 5.5）
└── routes.go                         # ✨ 新增：路由注册
```

---

## API Endpoints Specification

### 1. 项目成员管理 API

#### 1.1 添加项目成员

**Endpoint**: `POST /api/v1/projects/{project_id}/members`

**权限要求**: `member:create` (Admin/Owner only)

**请求体**:
```json
{
  "user_id": 123,
  "role": "member"  // 可选值: owner, admin, member, viewer
}
```

**响应** (201 Created):
```json
{
  "code": 0,
  "message": "Member added successfully",
  "data": {
    "id": 456,
    "project_id": 789,
    "user_id": 123,
    "role": "member",
    "joined_at": "2026-01-14T10:00:00Z",
    "updated_at": "2026-01-14T10:00:00Z"
  }
}
```

**错误响应**:
- `400 Bad Request`: 无效的角色名称
- `403 Forbidden`: 无权限添加成员
- `404 Not Found`: 项目或用户不存在
- `409 Conflict`: 用户已是项目成员

**Swagger Annotation**:
```go
// AddProjectMember godoc
// @Summary      Add a member to a project
// @Description  Add a user to a project with a specific role (requires admin permission)
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Param        project_id   path      int                      true  "Project ID"
// @Param        request      body      AddMemberRequest         true  "Member info"
// @Success      201          {object}  response.Response{data=ent.ProjectMember}
// @Failure      400          {object}  response.Response
// @Failure      403          {object}  response.Response
// @Failure      404          {object}  response.Response
// @Failure      409          {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/projects/{project_id}/members [post]
```

---

#### 1.2 列出项目成员

**Endpoint**: `GET /api/v1/projects/{project_id}/members`

**权限要求**: `member:read` (所有成员可访问)

**查询参数**:
- `role` (可选): 按角色过滤 (owner, admin, member, viewer)
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 20

**响应** (200 OK):
```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "items": [
      {
        "id": 456,
        "project_id": 789,
        "user_id": 123,
        "role": "owner",
        "joined_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z",
        "user": {
          "id": 123,
          "username": "admin",
          "email": "admin@example.com"
        }
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

**Swagger Annotation**:
```go
// ListProjectMembers godoc
// @Summary      List project members
// @Description  Get a list of all members in a project
// @Tags         rbac
// @Produce      json
// @Param        project_id   path      int     true   "Project ID"
// @Param        role         query     string  false  "Filter by role"
// @Param        page         query     int     false  "Page number" default(1)
// @Param        page_size    query     int     false  "Page size" default(20)
// @Success      200          {object}  response.Response{data=MemberListResponse}
// @Failure      403          {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/projects/{project_id}/members [get]
```

---

#### 1.3 更新成员角色

**Endpoint**: `PUT /api/v1/projects/{project_id}/members/{member_id}`

**权限要求**: `member:update` (Admin/Owner only)

**请求体**:
```json
{
  "role": "admin"
}
```

**响应** (200 OK):
```json
{
  "code": 0,
  "message": "Member role updated successfully",
  "data": {
    "id": 456,
    "project_id": 789,
    "user_id": 123,
    "role": "admin",
    "joined_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-14T10:00:00Z"
  }
}
```

**错误响应**:
- `400 Bad Request`: 无效的角色名称
- `403 Forbidden`: 无权限或尝试修改 owner 角色
- `404 Not Found`: 成员不存在

**业务规则**:
- ❌ 不能修改 owner 角色（必须通过转移所有权功能）
- ✅ Admin 可以修改除 owner 外的任何角色
- ✅ Owner 可以修改任何角色（除了自己）

---

#### 1.4 移除项目成员

**Endpoint**: `DELETE /api/v1/projects/{project_id}/members/{member_id}`

**权限要求**: `member:delete` (Admin/Owner only)

**响应** (200 OK):
```json
{
  "code": 0,
  "message": "Member removed successfully",
  "data": null
}
```

**错误响应**:
- `403 Forbidden`: 无权限或尝试移除 owner
- `404 Not Found`: 成员不存在

**业务规则**:
- ❌ 不能移除项目 owner
- ✅ Owner 可以移除任何非 owner 成员
- ✅ Admin 可以移除 member 和 viewer

---

### 2. 权限查询 API

#### 2.1 查询当前用户权限

**Endpoint**: `GET /api/v1/projects/{project_id}/permissions/me`

**权限要求**: 项目成员

**响应** (200 OK):
```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "user_id": 123,
    "project_id": 789,
    "roles": ["admin"],
    "permissions": [
      {
        "resource": "config",
        "actions": ["create", "read", "update", "delete"]
      },
      {
        "resource": "data",
        "actions": ["create", "read", "update", "delete"]
      },
      {
        "resource": "member",
        "actions": ["create", "read", "update", "delete"]
      }
    ]
  }
}
```

**Swagger Annotation**:
```go
// GetMyPermissions godoc
// @Summary      Get current user permissions
// @Description  Get all permissions for the current user in a project
// @Tags         rbac
// @Produce      json
// @Param        project_id   path      int  true  "Project ID"
// @Success      200          {object}  response.Response{data=UserPermissionsResponse}
// @Failure      403          {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/projects/{project_id}/permissions/me [get]
```

---

#### 2.2 检查特定权限

**Endpoint**: `POST /api/v1/projects/{project_id}/permissions/check`

**权限要求**: 项目成员

**请求体**:
```json
{
  "resource": "config",
  "action": "delete"
}
```

**响应** (200 OK):
```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "allowed": true,
    "resource": "config",
    "action": "delete"
  }
}
```

**用途**: 前端根据此 API 动态显示/隐藏操作按钮

**Swagger Annotation**:
```go
// CheckPermission godoc
// @Summary      Check specific permission
// @Description  Check if the current user has a specific permission
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Param        project_id   path      int                    true  "Project ID"
// @Param        request      body      CheckPermissionRequest true  "Permission to check"
// @Success      200          {object}  response.Response{data=PermissionCheckResponse}
// @Failure      403          {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/projects/{project_id}/permissions/check [post]
```

---

### 3. RBAC 管理 API (Admin Only)

#### 3.1 重载 RBAC 策略

**Endpoint**: `POST /api/v1/admin/rbac/reload`

**权限要求**: `platform_admin` (平台管理员)

**响应** (200 OK):
```json
{
  "code": 0,
  "message": "RBAC policies reloaded successfully",
  "data": {
    "reloaded_at": "2026-01-14T10:00:00Z",
    "policy_count": 42
  }
}
```

**使用场景**: 
- 配置文件更新后手动触发重载
- Database Adapter 迁移后刷新策略

---

## Implementation Tasks

### Task 1: 创建 Handler 文件

**文件**: `core/modules/auth/handler/project_member_handler.go`

**实现要点**:
```go
package handler

import (
    "net/http"
    "strconv"
    
    "github.com/go-chi/chi/v5"
    "apprun/ent"
    "apprun/internal/jwt"
    "apprun/modules/auth/service"
    "apprun/pkg/errors"
    "apprun/pkg/logger"
    "apprun/pkg/response"
)

type ProjectMemberHandler struct {
    memberService *service.ProjectMemberService
}

func NewProjectMemberHandler(memberService *service.ProjectMemberService) *ProjectMemberHandler {
    return &ProjectMemberHandler{
        memberService: memberService,
    }
}

// AddMember - POST /api/v1/projects/{project_id}/members
func (h *ProjectMemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 从路径提取 project_id
    projectID, err := strconv.ParseInt(chi.URLParam(r, "project_id"), 10, 64)
    if err != nil {
        response.Error(w, errors.NewAppError(errors.ErrCodeBadRequest, "Invalid project_id"))
        return
    }
    
    // 2. 解析请求体
    var req AddMemberRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, errors.NewAppError(errors.ErrCodeBadRequest, "Invalid request body"))
        return
    }
    
    // 3. 验证请求
    if err := req.Validate(); err != nil {
        response.Error(w, err)
        return
    }
    
    // 4. 调用 Service 层
    member, err := h.memberService.AddMember(ctx, projectID, req.UserID, req.Role)
    if err != nil {
        logger.Error("Failed to add member", 
            "error", err,
            "project_id", projectID,
            "user_id", req.UserID,
            "role", req.Role,
        )
        response.Error(w, err)
        return
    }
    
    // 5. 审计日志
    operatorID := jwt.GetUserID(ctx)
    logger.Info("Member added",
        "operator_id", operatorID,
        "project_id", projectID,
        "new_member_user_id", req.UserID,
        "role", req.Role,
    )
    
    // 6. 返回成功响应
    response.Success(w, member, http.StatusCreated)
}

// 其他方法类似...
```

**验证规则**:
```go
type AddMemberRequest struct {
    UserID int64  `json:"user_id" validate:"required,gt=0"`
    Role   string `json:"role" validate:"required,oneof=owner admin member viewer"`
}

func (r *AddMemberRequest) Validate() error {
    if r.UserID <= 0 {
        return errors.NewAppError(errors.ErrCodeBadRequest, "user_id must be positive")
    }
    
    validRoles := map[string]bool{
        "owner": true, "admin": true, "member": true, "viewer": true,
    }
    if !validRoles[r.Role] {
        return errors.NewAppError(errors.ErrCodeBadRequest, "Invalid role")
    }
    
    return nil
}
```

---

### Task 2: 创建权限查询 Handler

**文件**: `core/modules/auth/handler/permission_handler.go`

**实现要点**:
```go
package handler

type PermissionHandler struct {
    permService *service.PermissionService
}

func NewPermissionHandler(permService *service.PermissionService) *PermissionHandler {
    return &PermissionHandler{
        permService: permService,
    }
}

// GetMyPermissions - GET /api/v1/projects/{project_id}/permissions/me
func (h *PermissionHandler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 提取 user_id 和 project_id
    userID := jwt.GetUserID(ctx)
    projectID, err := strconv.ParseInt(chi.URLParam(r, "project_id"), 10, 64)
    if err != nil {
        response.Error(w, errors.NewAppError(errors.ErrCodeBadRequest, "Invalid project_id"))
        return
    }
    
    // 2. 获取用户角色
    roles, err := h.permService.GetUserRoles(userID, projectID)
    if err != nil {
        response.Error(w, err)
        return
    }
    
    // 3. 获取用户权限
    permissions, err := h.permService.GetUserPermissions(ctx, userID, projectID)
    if err != nil {
        response.Error(w, err)
        return
    }
    
    // 4. 构造响应
    resp := UserPermissionsResponse{
        UserID:      userID,
        ProjectID:   projectID,
        Roles:       roles,
        Permissions: permissions,
    }
    
    response.Success(w, resp, http.StatusOK)
}

// CheckPermission - POST /api/v1/projects/{project_id}/permissions/check
func (h *PermissionHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
    // 类似实现...
}
```

---

### Task 3: 注册路由

**文件**: `core/modules/auth/routes.go`

```go
package auth

import (
    "github.com/go-chi/chi/v5"
    "apprun/internal/middleware"
    "apprun/modules/auth/handler"
)

func RegisterRoutes(r chi.Router, memberHandler *handler.ProjectMemberHandler, permHandler *handler.PermissionHandler) {
    // 项目成员管理路由
    r.Route("/api/v1/projects/{project_id}/members", func(r chi.Router) {
        // 中间件链
        r.Use(middleware.AuthMiddleware)           // JWT 认证
        r.Use(middleware.ProjectContextMiddleware) // 项目上下文
        
        // 列出成员（所有成员可访问）
        r.With(middleware.RequirePermission("member", "read")).Get("/", memberHandler.ListMembers)
        
        // 管理成员（需要 admin 权限）
        r.With(middleware.RequirePermission("member", "create")).Post("/", memberHandler.AddMember)
        r.With(middleware.RequirePermission("member", "update")).Put("/{member_id}", memberHandler.UpdateRole)
        r.With(middleware.RequirePermission("member", "delete")).Delete("/{member_id}", memberHandler.RemoveMember)
    })
    
    // 权限查询路由
    r.Route("/api/v1/projects/{project_id}/permissions", func(r chi.Router) {
        r.Use(middleware.AuthMiddleware)
        r.Use(middleware.ProjectContextMiddleware)
        
        r.Get("/me", permHandler.GetMyPermissions)
        r.Post("/check", permHandler.CheckPermission)
    })
    
    // RBAC 管理路由（平台管理员）
    r.Route("/api/v1/admin/rbac", func(r chi.Router) {
        r.Use(middleware.AuthMiddleware)
        r.Use(middleware.RequirePlatformAdmin) // 平台管理员中间件
        
        r.Post("/reload", rbacHandler.ReloadPolicies)
    })
}
```

---

### Task 4: 集成测试

**文件**: `core/modules/auth/handler/handler_test.go`

**测试场景**:

```go
package handler_test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAddMember_Success(t *testing.T) {
    // Setup
    handler := setupTestHandler(t)
    
    // 创建测试项目和用户
    project := createTestProject(t)
    owner := createTestUser(t, "owner")
    newMember := createTestUser(t, "member")
    
    // 构造请求
    reqBody := `{"user_id": %d, "role": "member"}`
    req := httptest.NewRequest("POST", 
        fmt.Sprintf("/api/v1/projects/%d/members", project.ID), 
        strings.NewReader(fmt.Sprintf(reqBody, newMember.ID)))
    req.Header.Set("Authorization", "Bearer "+generateTestToken(owner.ID))
    
    // 执行请求
    w := httptest.NewRecorder()
    handler.AddMember(w, req)
    
    // 验证响应
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var resp response.Response
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.Equal(t, 0, resp.Code)
    
    // 验证数据库
    member, err := memberRepo.GetMember(context.Background(), project.ID, newMember.ID)
    require.NoError(t, err)
    assert.Equal(t, "member", member.Role)
}

func TestAddMember_Forbidden(t *testing.T) {
    // 测试无权限用户尝试添加成员
    // 应返回 403 Forbidden
}

func TestAddMember_DuplicateMember(t *testing.T) {
    // 测试重复添加成员
    // 应返回 409 Conflict
}

func TestListMembers_Success(t *testing.T) {
    // 测试列出成员
}

func TestUpdateRole_Success(t *testing.T) {
    // 测试更新角色
}

func TestUpdateRole_CannotModifyOwner(t *testing.T) {
    // 测试不能修改 owner 角色
    // 应返回 403 Forbidden
}

func TestRemoveMember_Success(t *testing.T) {
    // 测试移除成员
}

func TestRemoveMember_CannotRemoveOwner(t *testing.T) {
    // 测试不能移除 owner
    // 应返回 403 Forbidden
}

func TestGetMyPermissions_Success(t *testing.T) {
    // 测试查询当前用户权限
}

func TestCheckPermission_Allowed(t *testing.T) {
    // 测试权限检查（允许）
}

func TestCheckPermission_Denied(t *testing.T) {
    // 测试权限检查（拒绝）
}
```

---

### Task 5: 更新 Swagger 文档

**步骤**:
1. 在所有 Handler 方法上添加 Swagger annotations（参考上面示例）
2. 运行 `make swagger` 生成文档
3. 验证 Swagger UI：访问 `http://localhost:8080/swagger/index.html`

---

## Error Codes

复用 Story 1.3 和 5.5 的错误码：

```go
// pkg/errors/codes.go (已存在)
const (
    ErrCodeBadRequest        = 400001  // 通用
    ErrCodeAuthRequired      = 401001  // Story 5.3
    ErrCodeAuthNotMember     = 403002  // Story 5.5
    ErrCodePermForbidden     = 403003  // Story 5.5
    ErrCodeNotFound          = 404001  // 通用
    ErrCodeConflict          = 409001  // 通用
    
    // ✨ 新增
    ErrCodeInvalidRole       = 400002  // 无效角色名称
    ErrCodeCannotModifyOwner = 403004  // 不能修改/移除 owner
    ErrCodeUserNotFound      = 404002  // 用户不存在
    ErrCodeMemberExists      = 409002  // 成员已存在
)
```

---

## Testing Strategy

### 单元测试
- Handler 逻辑测试（使用 mock Service）
- 请求验证测试
- 错误处理测试

### 集成测试
- 完整 API 流程测试（包含中间件）
- 数据库交互测试
- 权限验证测试

### 性能测试
- 并发请求测试（50+ concurrent users）
- 响应时间测试（< 100ms p95）
- 数据库查询优化验证

---

## Integration Points

### Story 5.5 (RBAC 核心)
- ✅ 使用已实现的 Service 层（ProjectMemberService, PermissionService）
- ✅ 使用已实现的 Repository 层
- ✅ 使用已实现的中间件（ProjectContextMiddleware, RequirePermission）

### Story 5.3 (JWT)
- ✅ 使用 jwt.GetUserID(ctx) 获取当前用户

### Story 4.1 (Swagger)
- ✅ 添加 Swagger annotations
- ✅ 更新 API 文档

### Story 1.3 (Error Handling)
- ✅ 使用 pkg/errors.AppError
- ✅ 使用 pkg/response.Success/Error

### Story 1.10 (Logger)
- ✅ 使用 pkg/logger 结构化日志
- ✅ 记录审计日志

---

## Performance Optimization

### 1. 数据库查询优化
```go
// 使用预加载避免 N+1 查询
members, err := repo.ListMembers(ctx, projectID)
// 内部使用 .WithUser() 预加载用户信息
```

### 2. 响应缓存（可选）
```go
// 对 GET /permissions/me 使用短期缓存
// Cache-Control: max-age=60
```

### 3. 批量权限检查
```go
// 前端可以批量检查多个权限
POST /api/v1/projects/{project_id}/permissions/batch-check
{
  "checks": [
    {"resource": "config", "action": "create"},
    {"resource": "data", "action": "delete"}
  ]
}
```

---

## Security Considerations

### 1. Owner 保护
```go
// 业务规则验证
func (h *ProjectMemberHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
    member, err := h.memberService.GetMemberByID(ctx, memberID)
    if err != nil {
        response.Error(w, err)
        return
    }
    
    if member.Role == "owner" {
        response.Error(w, errors.NewAppError(
            errors.ErrCodeCannotModifyOwner, 
            "Cannot remove project owner",
        ))
        return
    }
    
    // 继续处理...
}
```

### 2. 输入验证
```go
// 使用 validator 库
type AddMemberRequest struct {
    UserID int64  `json:"user_id" validate:"required,gt=0"`
    Role   string `json:"role" validate:"required,oneof=owner admin member viewer"`
}
```

### 3. 审计日志
```go
// 记录所有成员变更
logger.Info("Member added",
    "operator_id", operatorID,
    "project_id", projectID,
    "new_member_user_id", req.UserID,
    "role", req.Role,
    "ip", r.RemoteAddr,
)
```

---

## Migration & Deployment

### 数据库
- ✅ 无需新的迁移（Story 5.5 已完成）

### 配置
- ✅ 无需新配置（复用 Story 5.5）

### 部署步骤
1. 合并代码到主分支
2. 运行测试：`make test`
3. 生成 Swagger 文档：`make swagger`
4. 构建：`make build`
5. 部署到测试环境
6. 验证 API endpoints
7. 发布到生产环境

---

## Verification Checklist (DoD)

### 功能验证
- [ ] 所有 API endpoints 正常工作
- [ ] Swagger 文档完整且准确
- [ ] 权限控制正确（admin 可管理，viewer 只读）
- [ ] Owner 保护生效（不能移除/降级）
- [ ] 项目隔离验证（不能操作其他项目）

### 测试验证
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试全部通过
- [ ] 性能测试达标（< 100ms p95）

### 安全验证
- [ ] 输入验证完整
- [ ] 权限检查严格
- [ ] 审计日志完整

### 文档验证
- [ ] Swagger 文档准确
- [ ] API 示例可用
- [ ] 错误码文档完整

---

## References

- **Parent Story**: [5-5-rbac-permissions.md](./5-5-rbac-permissions.md) - RBAC 核心基础设施
- **Epic**: [5-auth-epic.md](../../epics/5-auth-epic.md)
- **PRD**: [FR-AUTH-001 - Authentication & Authorization](../../prd.md#21-authentication--authorization)
- **Story 5.3**: [5-3-jwt-middleware.md](./5-3-jwt-middleware.md) - JWT 中间件
- **Story 4.1**: [4-1-swagger-docs.md](./4-1-swagger-docs.md) - Swagger 文档
- **Story 1.3**: Error Handling Framework
- **Story 1.10**: Logger Package

---

**Story Created By**: Bob (Scrum Master) - BMad SM Agent  
**Date**: 2026-01-14  
**Workflow**: create-story (Manual split from Story 5-5)  
**Reason**: 关注点分离 - 核心 RBAC 基础设施 vs HTTP API 接口层  
**Status**: Ready for Development
