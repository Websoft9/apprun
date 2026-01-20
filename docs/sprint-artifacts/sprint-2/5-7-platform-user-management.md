# Story 5.7: 平台用户管理
# Sprint 2: 认证与授权 (Auth Epic)

**Priority**: P1  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 5.1 (用户注册) - 已完成
- Story 5.3 (JWT 中间件) - 已完成
- Story 5.5 (RBAC 权限控制) - 已完成
- Story 5.9 (审计日志中间件) - 可选依赖

**Status**: 📋 Backlog  
**Module**: Authentication / Admin  
**Epic**: [auth-epic](../../epics/5-auth-epic.md)  
**Issue**: #TBD

---

## ⚠️ 重要说明

**审计日志职责分离**:
- ✅ **本 Story (5.7)**: 实现用户管理 CRUD API 和业务逻辑
- ✅ **Story 5.9**: 独立实现审计日志中间件和存储
- ✅ **集成方式**: 审计日志作为中间件自动拦截并记录所有 `/api/admin/*` 操作
- ✅ **开发顺序**: 可并行开发，也可先开发 5.7（无审计），后接入 5.9 中间件

**Token 撤销方案**:
- ✅ **采用方案B（版本号）**: 在 User 表增加 `token_version` 字段
- ✅ **立即生效**: 禁用/删除用户后，所有旧Token下次请求立即失效
- ✅ **简单高效**: 无需Redis，单表查询，性能影响最小

**本 Story 不包含**:
- ❌ 审计日志表结构设计
- ❌ 审计日志写入逻辑
- ❌ 审计日志查询 API
- ❌ 审计日志归档策略

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
- [ ] 管理操作将被审计日志中间件自动记录（依赖 Story 5.9）

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
6. 撤销该用户的所有活跃 token（通过 Token 黑名单）
7. 审计日志由中间件自动记录（Story 5.9 提供）

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

**注**: 审计日志功能由 Story 5.9 独立实现，作为中间件自动记录管理操作

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
        field.Time("deleted_at").Optional().Nillable().Comment("软删除时间戳"),
        field.Int("token_version").Default(0).Comment("Token版本号，用于撤销旧Token"),
    }
}

func (User) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("email", "deleted_at"),
        index.Fields("role"),
        index.Fields("is_active"),
        index.Fields("token_version"),  // Token版本号索引
    }
}
```

**Migration**:
```sql
-- 软删除支持
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- Token版本号支持（方案B）
ALTER TABLE users ADD COLUMN token_version INT DEFAULT 0 NOT NULL;

-- 索引优化
CREATE INDEX idx_users_email_deleted ON users(email, deleted_at);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_token_version ON users(token_version);
```

**注**: 审计日志表结构由 Story 5.9 定义，本 Story 不涉及

---

## Token 撤销策略（方案B：版本号）✅ 已选定

### 设计原理

用户表增加 `token_version` 字段，每次禁用/删除用户时递增版本号，使所有旧Token立即失效。

### 完整实现方案
### 完整实现方案

#### 1. JWT Claims 扩展
```go
// core/internal/jwt/jwt.go
type Claims struct {
    UserID       int64  `json:"user_id"`
    TokenVersion int    `json:"token_version"`  // 新增字段
    jwt.RegisteredClaims
}
```

#### 2. 登录时生成Token（带版本号）
```go
// core/modules/auth/service/auth.go
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // ... 验证密码 ...
    
    user, _ := s.userRepo.GetByEmail(ctx, req.Email)
    
    // 生成Token时包含版本号
    token, err := jwt.GenerateToken(user.ID, user.TokenVersion)
    if err != nil {
        return nil, err
    }
    
    return &LoginResponse{
        Token: token,
        User:  toUserResponse(user),
    }, nil
}

// core/internal/jwt/jwt.go
func GenerateToken(userID int64, tokenVersion int) (string, error) {
    claims := Claims{
        UserID:       userID,
        TokenVersion: tokenVersion,  // 从数据库读取
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(jwtSecret))
}
```

#### 3. JWT中间件验证版本号
```go
// core/internal/middleware/jwt.go
func (m *JWTMiddleware) JWTAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. 解析Token
        tokenString := extractToken(r)
        claims, err := jwt.ValidateToken(tokenString)
        if err != nil {
            response.Error(w, 401, "INVALID_TOKEN", "Invalid or expired token")
            return
        }
        
        // 2. 查询用户
        user, err := userRepo.GetByID(ctx, claims.UserID)
        if err != nil {
            response.Error(w, 401, "USER_NOT_FOUND", "User not found")
            return
        }
        
        // 3. 验证版本号（关键逻辑）
        if claims.TokenVersion != user.TokenVersion {
            response.Error(w, 401, "TOKEN_REVOKED", "Token has been revoked")
            return
        }
        
        // 4. 检查用户状态
        if user.Status != 1 {
            response.Error(w, 401, "USER_DISABLED", "User account is disabled")
            return
        }
        
        // 5. 注入用户ID到Context
        ctx = jwt.SetUserID(ctx, user.ID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### 4. 禁用/删除用户时递增版本号
```go
// core/modules/admin/services/user_mgmt.go

// 禁用用户
func (s *UserMgmtService) DisableUser(ctx context.Context, userID int64) error {
    return s.client.User.UpdateOneID(userID).
        SetStatus(0).                // 禁用账户
        AddTokenVersion(1).          // 版本号+1（所有旧Token立即失效）
        Exec(ctx)
}

// 删除用户（软删除）
func (s *UserMgmtService) DeleteUser(ctx context.Context, userID int64) error {
    now := time.Now()
    return s.client.User.UpdateOneID(userID).
        SetDeletedAt(now).           // 软删除标记
        SetStatus(0).                // 禁用账户
        AddTokenVersion(1).          // 版本号+1（所有旧Token立即失效）
        Exec(ctx)
}

// 修改角色（可选：是否撤销Token取决于业务需求）
func (s *UserMgmtService) ChangeUserRole(ctx context.Context, userID int64, newRole string) error {
    return s.client.User.UpdateOneID(userID).
        SetRole(newRole).
        AddTokenVersion(1).          // 角色变更后撤销旧Token（可选）
        Exec(ctx)
}
```

### 方案优势

| 优势 | 说明 |
|------|------|
| ✅ **立即生效** | 禁用后下次请求立即失效（不用等24小时） |
| ✅ **简单实现** | 只需一个整型字段，无需Redis |
| ✅ **高性能** | 单表查询，JWT中间件已有查询逻辑 |
| ✅ **批量撤销** | 一次递增使该用户所有Token失效 |
| ✅ **易维护** | 逻辑清晰，无外部依赖 |
| ✅ **可追溯** | 版本号可用于审计和调试 |

### 测试验证

```go
// 测试场景
func TestTokenRevocation(t *testing.T) {
    // 1. 用户登录获取Token（version=0）
    token1, _ := authService.Login(ctx, email, password)
    
    // 2. 使用Token访问API（成功）
    resp1 := callProtectedAPI(token1)  // ✅ 200 OK
    
    // 3. 管理员禁用该用户（version变为1）
    adminService.DisableUser(ctx, userID)
    
    // 4. 再次使用旧Token访问API（失败）
    resp2 := callProtectedAPI(token1)  // ❌ 401 TOKEN_REVOKED
    
    // 5. 用户重新登录获取新Token（version=1）
    token2, _ := authService.Login(ctx, email, password)
    
    // 6. 使用新Token访问API（成功）
    resp3 := callProtectedAPI(token2)  // ✅ 200 OK
}
```

---

### 备选方案对比（仅供参考）

<details>
<summary>方案 A：延迟生效（点击展开）</summary>

**原理**: 不做任何特殊处理，依赖Token自然过期

**优点**: 
- 最简单，无需任何额外开发

**缺点**: 
- 延迟最多24小时生效

**适用**: MVP快速验证阶段

</details>

<details>
<summary>方案 C：更新时间检查（点击展开）</summary>

**原理**: 利用 `updated_at` 字段对比Token签发时间

**优点**: 
- 零字段新增

**缺点**: 
- 任何用户信息修改都会使Token失效（包括修改昵称）

**适用**: 用户不可自己修改资料的场景

</details>

<details>
<summary>方案 D：Redis黑名单（点击展开）</summary>

**原理**: 在Redis维护被撤销的Token列表

**优点**: 
- 最灵活，可撤销单个Token

**缺点**: 
- 需要Redis依赖
- 每次请求需额外查询Redis

**适用**: 已有Redis且需要精细控制

</details>

---

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

4. **Token 版本号支持**
   - 修改 JWT 生成逻辑，Claims 中包含 `token_version`
   - 修改 JWT 中间件，验证版本号一致性
   - 禁用/删除用户时递增 `token_version`

5. **安全检查**
   - 实现"防止删除自己"逻辑
   - 实现"防止降级最后一个管理员"逻辑
   - 实现"禁止操作 system 用户"逻辑

6. **测试**
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
- 禁用用户后，旧Token立即失效（版本号验证）
- 用户重新登录后获取新Token可正常访问
- 审计日志自动记录（需 Story 5.9 审计中间件）

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
- **重要**: 审计日志由 Story 5.9 独立实现
- 本 Story 的所有管理操作将被审计中间件自动记录
- 无需在业务代码中手动记录日志
- 审计中间件将捕获：操作者 ID、操作类型、目标用户 ID、IP 地址、时间戳

### Token 撤销
- **推荐方案 A（延迟生效）**: JWT 中间件检查用户状态，禁用用户后现有 Token 在过期前仍有效（最多 24 小时）
- **可选升级到方案 B（版本号）**: 增加 `token_version` 字段，立即撤销所有旧 Token
- 详见下文"Token 撤销策略"章节的 4 种方案对比

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
