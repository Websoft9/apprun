# Story 5.7: 平台用户管理
# Sprint 2: 认证与授权 (Auth Epic)

**Priority**: P1  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 5.1 (用户注册) - 已完成
- Story 5.3 (JWT 中间件) - 已完成
- Story 5.5 (RBAC 权限控制) - 已完成

**Status**: 📋 Backlog  
**Module**: Authentication / Admin  
**Epic**: [auth-epic](../../epics/5-auth-epic.md)  
**Issue**: #TBD

---

## User Story

作为 **平台管理员 (platform_admin)**，我希望能够查看所有用户列表、创建新的管理员账户、修改用户的角色和状态（启用/禁用），以便我能有效管理平台的用户和权限体系，同时系统需要防止我误删自己或降级最后一个管理员。

---

## Acceptance Criteria

### 功能验收
- [ ] 实现 `GET /api/admin/users` - 列出所有用户
  - 支持分页（page, page_size）
  - 支持搜索（email, name）
  - 支持角色过滤（platform_admin, platform_user）
  - 支持状态过滤（active, inactive）
- [ ] 实现 `POST /api/admin/users` - 创建新用户/管理员
  - 可指定 email, name, password, role
  - 自动生成随机密码（如未提供）
- [ ] 实现 `GET /api/admin/users/:id` - 查看指定用户详细信息
- [ ] 实现 `PUT /api/admin/users/:id/role` - 修改用户角色
  - 支持角色：platform_admin, platform_user
- [ ] 实现 `PUT /api/admin/users/:id/status` - 修改用户状态
  - 支持状态：active, inactive, banned
- [ ] 实现 `DELETE /api/admin/users/:id` - 删除用户（软删除）

### 安全验证
- [ ] 所有端点仅限 `platform_admin` 角色访问
- [ ] 实现 `RequirePlatformAdmin` 中间件
- [ ] 防止管理员删除自己
- [ ] 防止降级最后一个 `platform_admin`
- [ ] 防止修改 system 用户（is_system=true）
- [ ] 审计日志记录所有管理操作（操作者、被操作者、操作类型、时间）

### 非功能验收
- [ ] API 响应时间 P95 < 300ms
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖关键安全场景

---

## Technical Design

### API 端点

#### 1. GET /api/admin/users
**描述**: 获取用户列表  
**认证**: JWT Required + platform_admin  
**查询参数**:
- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 20）
- `search`: 搜索关键词（匹配 email, name）
- `role`: 角色过滤（platform_admin / platform_user）
- `status`: 状态过滤（active / inactive / banned）

**响应示例**:
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "uuid",
        "email": "admin@example.com",
        "name": "Admin User",
        "role": "platform_admin",
        "status": "active",
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "total": 42,
    "page": 1,
    "page_size": 20
  }
}
```

#### 2. POST /api/admin/users
**描述**: 创建新用户  
**认证**: JWT Required + platform_admin  
**请求体**:
```json
{
  "email": "newuser@example.com",
  "name": "New User",
  "password": "SecurePass123!",
  "role": "platform_user"
}
```

**响应示例 (Success 201)**:
```json
{
  "success": true,
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440002",
    "email": "newuser@example.com",
    "name": "New User",
    "role": "platform_user",
    "is_active": true,
    "created_at": "2026-01-19T08:00:00Z"
  }
}
```

**响应示例 (Error 400)**:
```json
{
  "success": false,
  "error": {
    "code": "USER_EMAIL_EXISTS",
    "message": "该邮箱已被注册"
  }
}
```

#### 3. PUT /api/admin/users/:id/role
**描述**: 修改用户角色  
**认证**: JWT Required + platform_admin  
**请求体**:
```json
{
  "role": "platform_admin"
}
```

**响应示例 (Success 200)**:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "User Name",
    "role": "platform_admin",
    "is_active": true,
    "updated_at": "2026-01-19T08:15:00Z"
  }
}
```

**响应示例 (Error 403)**:
```json
{
  "success": false,
  "error": {
    "code": "USER_CANNOT_DEMOTE_LAST_ADMIN",
    "message": "不能降级最后一个管理员"
  }
}
```

#### 4. PUT /api/admin/users/:id/status
**描述**: 修改用户状态  
**认证**: JWT Required + platform_admin  
**请求体**:
```json
{
  "status": "inactive"
}
```

**响应示例 (Success 200)**:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "User Name",
    "role": "platform_user",
    "is_active": false,
    "updated_at": "2026-01-19T08:20:00Z"
  }
}
```

#### 5. DELETE /api/admin/users/:id
**描述**: 删除用户（软删除）  
**认证**: JWT Required + platform_admin  

**响应示例 (Success 200)**:
```json
{
  "success": true,
  "message": "用户已删除"
}
```

**响应示例 (Error 403)**:
```json
{
  "success": false,
  "error": {
    "code": "USER_CANNOT_DELETE_SELF",
    "message": "不能删除自己"
  }
}
```

**逻辑**:
1. 验证用户权限（RequirePlatformAdmin）
2. 检查是否为 system 用户（不允许删除）
3. 检查是否为自己（不允许自删除）
4. 检查是否为最后一个管理员（不允许删除）
5. 软删除用户（设置 deleted_at 字段，保留数据关联）
6. 撤销该用户的所有活跃 token
7. 记录审计日志

**级联处理策略**:
- **推荐方案**: 软删除（Soft Delete）
  - 设置 `deleted_at` 时间戳
  - 保留用户创建的数据关联（配置项、项目等）
  - 用户不能再登录，但历史数据可追溯
- **替代方案**: 转移所有权
  - 将用户创建的资源所有权转移给 system 用户
  - 用于硬删除场景

### 目录结构

```
modules/admin/
├── handlers/
│   └── users.go           # 用户管理 API
├── services/
│   └── user_mgmt.go       # 用户管理业务逻辑
└── middleware/
    └── require_admin.go   # 管理员权限中间件
```

---

## Data Model Changes

### 用户表增强

**Ent Schema** (`ent/schema/user.go`):
```go
import (
    "time"
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

func (User) Fields() []ent.Field {
    return []ent.Field{
        // ... 现有字段
        field.Time("deleted_at").Optional().Nillable(),
    }
}

func (User) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("email", "deleted_at"),
        index.Fields("role"),
        index.Fields("is_active"),
    }
}
```

**Migration**:
```sql
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;
CREATE INDEX idx_users_email_deleted ON users(email, deleted_at);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);
```

### 审计日志表

**Ent Schema** (`ent/schema/auditlog.go`):
```go
func (AuditLog) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.Time("timestamp").Default(time.Now),
        field.UUID("operator_id", uuid.UUID{}), // 操作者 ID
        field.String("action"),                  // user.create, user.update_role, user.delete
        field.UUID("target_id", uuid.UUID{}),    // 目标用户 ID
        field.JSON("changes", map[string]interface{}{}), // 修改内容
        field.String("ip_address").Optional(),
        field.String("user_agent").Optional(),
    }
}
```

**审计日志格式**:
```json
{
  "timestamp": "2026-01-19T08:15:00Z",
  "operator_id": "550e8400-e29b-41d4-a716-446655440000",
  "action": "user.update_role",
  "target_id": "660e8400-e29b-41d4-a716-446655440001",
  "changes": {
    "role": {
      "from": "platform_user",
      "to": "platform_admin"
    }
  },
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0..."
}
```

---

## Implementation Tasks

1. **中间件实现**
   - 创建 `RequirePlatformAdmin` 中间件
   - 从 JWT Claims 中提取用户角色
   - 验证是否为 `platform_admin`

2. **Handler 层实现**
   - 实现用户列表 API（带分页、搜索、过滤）
   - 实现创建用户 API
   - 实现修改角色 API
   - 实现修改状态 API
   - 实现删除用户 API

3. **Service 层实现**
   - 实现 `ListUsers()` 方法（支持过滤条件）
   - 实现 `CreateUser()` 方法
   - 实现 `ChangeUserRole()` 方法（包含安全检查）
   - 实现 `ChangeUserStatus()` 方法（防止自我禁用）
   - 实现 `DeleteUser()` 方法（软删除，防止自删）

4. **安全检查**
   - 实现"防止删除自己"逻辑
   - 实现"防止降级最后一个管理员"逻辑
   - 实现"禁止操作 system 用户"逻辑
   - 添加审计日志

5. **测试**
   - 单元测试（各个 Service 方法）
   - 集成测试（完整 API 流程）
   - 安全测试（尝试非法操作）

6. **文档**
   - 更新 Swagger API 文档
   - 添加错误码定义

---

## Testing Strategy

### 单元测试场景
- 管理员可以列出所有用户
- 管理员可以创建新用户
- 管理员可以修改其他用户的角色
- 管理员不能删除自己
- 不能降级最后一个 platform_admin
- 不能修改 system 用户

### 集成测试场景
- 非管理员用户访问 `/api/admin/users` → 403 Forbidden
- 管理员创建新管理员 → 成功
- 管理员创建用户 → 新用户可以登录
- 管理员禁用用户 → 该用户无法登录且所有 token 失效
- 管理员尝试删除自己 → 400 Bad Request
- 管理员尝试删除最后一个管理员 → 返回 403
- 管理员尝试修改 system 用户角色 → 返回 403
- 审计日志正确记录所有操作

### 性能测试场景
- 10000+ 用户场景下列表查询响应时间 < 500ms
- 并发创建用户（10 req/s）成功率 100%

---

## Security Considerations

### 权限控制
- 所有 `/api/admin/*` 端点必须通过 `RequirePlatformAdmin` 中间件
- 使用 Casbin 或自定义逻辑验证角色
- JWT Token 中必须包含 `role` 字段

### 审计日志
每次管理操作需记录：
- 操作者 ID 和 Email
- 被操作者 ID 和 Email
- 操作类型（create / update_role / update_status / delete）
- 操作前后的值
- 时间戳和 IP 地址

### 防护机制
- **防自删**: 检查 `target_user_id != current_user_id`
- **防降级最后管理员**: 查询 `platform_admin` 数量，若 ≤1 则拒绝
- **防修改 system 用户**: 检查 `is_system` 字段

---

## Error Codes

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| `PERM_ADMIN_REQUIRED` | 403 | 需要平台管理员权限 |
| `USER_CANNOT_DELETE_SELF` | 400 | 不能删除自己 |
| `USER_CANNOT_DEMOTE_LAST_ADMIN` | 400 | 不能降级最后一个管理员 |
| `USER_CANNOT_MODIFY_SYSTEM` | 403 | 不能修改系统用户 |
| `USER_NOT_FOUND` | 404 | 用户不存在 |
| `USER_EMAIL_EXISTS` | 409 | 邮箱已存在 |

---

## Documentation

- [Epic 5: 认证与授权](../../epics/5-auth-epic.md)
- [API 设计规范](../../standards/api-design.md)
- [RBAC 权限模型](../../architecture/rbac-design.md)
