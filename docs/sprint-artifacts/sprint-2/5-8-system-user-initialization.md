# Story 5.8: 系统用户初始化支持
# Sprint 2: 认证与授权 (Auth Epic)

**Priority**: P0 (必需)  
**Effort**: 0.5 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 5.1 (用户注册与密码安全) - 已完成
- Story 1.17 (平台初始化) - 将调用此能力

**Status**: 📋 Backlog  
**Module**: Authentication / Bootstrap  
**Epic**: [auth-epic](../../epics/5-auth-epic.md)  
**Related**: [1-17-platform-initialization](../sprint-2/1-17-platform-initialization.md)  
**Issue**: #TBD

---

## User Story

作为 **系统开发者**，我希望 Auth 模块提供幂等的内部方法来创建 system 用户和首个管理员账户，以便 Bootstrap 模块在平台初始化时能够安全地创建这些内置账户，而不需要直接操作数据库。

---

## Acceptance Criteria

### 功能验收
- [ ] 实现 `AuthService.EnsureSystemUser(ctx)` 方法
  - 检查是否存在 `name="system"` 的用户
  - 不存在则创建，存在则返回已有用户
  - System 用户特性：
    - 固定 UUID: `00000000-0000-0000-0000-000000000000`
    - 无密码（password_hash 为空）
    - 特殊标记：`is_system=true`
    - 角色：`platform_user`（无特殊权限）
- [ ] 实现 `AuthService.EnsureAdminUser(ctx, email, password)` 方法
  - 检查是否存在该 email 的用户
  - 不存在则创建，存在则返回已有用户（不更新密码）
  - Admin 用户特性：
    - 生成随机 UUID
    - bcrypt 密码哈希
    - 角色：`platform_admin`
- [ ] System 用户禁止通过 `POST /api/auth/login` 登录
- [ ] 方法幂等性：多次调用返回相同结果，不重复创建

### 安全验证
- [ ] System 用户无密码，无法通过常规 API 登录
- [ ] Admin 用户密码必须符合强度要求
- [ ] 方法仅供内部调用，不暴露为 HTTP 接口
- [ ] 记录创建日志（INFO 级别）

### 非功能验收
- [ ] 方法执行时间 < 100ms
- [ ] 单元测试覆盖率 ≥ 80%（与 0.5 天工作量匹配）
- [ ] 幂等性测试通过

---

## Technical Design

### Service 方法签名

#### 1. EnsureSystemUser
**描述**: 确保 system 用户存在  
**签名**:
```go
func (s *AuthService) EnsureSystemUser(ctx context.Context) (*ent.User, error)
```

**逻辑**:
1. 查询 `name="system"` 的用户
2. 如果存在：
   - 验证一致性：`is_system=true` 且 `id="00000000-0000-0000-0000-000000000000"`
   - 一致性检查失败 → 返回错误（数据损坏）
   - 一致性检查通过 → 返回该用户
3. 如果不存在 → 创建新用户：
   - `id`: `00000000-0000-0000-0000-000000000000`
   - `name`: `"system"`
   - `email`: `"system@internal"`
   - `password_hash`: `""` (空)
   - `is_system`: `true`
   - `is_active`: `true`
   - `role`: `platform_user`

**返回值**:
- `*ent.User`: 用户对象
- `error`: 错误信息

#### 2. EnsureAdminUser
**描述**: 确保管理员用户存在  
**签名**:
```go
func (s *AuthService) EnsureAdminUser(ctx context.Context, email, password string) (*ent.User, error)
```

**逻辑**:
1. 验证参数（email 格式，password 强度）
2. 查询该 email 的用户
3. 如果存在：
   - 记录 INFO 日志：Admin user already exists
   - 返回该用户（不修改密码，不更新任何字段）
4. 如果不存在 → 创建新用户：
   - `id`: 随机 UUID
   - `email`: 参数提供
   - `name`: 从 email 提取或使用 "Platform Admin"
   - `password_hash`: bcrypt(password)
   - `is_system`: `false`
   - `is_active`: `true`
   - `role`: `platform_admin`

**参数验证**:
- email 不能为空，必须符合邮箱格式
- password 不能为空，必须符合 Story 5.1 定义的密码强度要求（调用 `AuthService.ValidatePassword()` 方法）

**返回值**:
- `*ent.User`: 用户对象（不包含 password_hash 字段）
- `error`: 错误信息

### 调用示例

```go
// 在 core/internal/bootstrap/server.go 中调用
func InitPlatform(container *Container) error {
    authSvc := container.AuthService
    
    // 1. 创建 system 用户
    sysUser, err := authSvc.EnsureSystemUser(ctx)
    if err != nil {
        return fmt.Errorf("init system user: %w", err)
    }
    log.Infof("System user ensured: %s", sysUser.ID)
    
    // 2. 创建管理员用户（从环境变量读取）
    adminEmail := os.Getenv("ADMIN_EMAIL")
    adminPass := os.Getenv("ADMIN_PASSWORD")
    
    // 验证环境变量：必须同时提供或同时为空
    if (adminEmail != "" && adminPass == "") || (adminEmail == "" && adminPass != "") {
        log.Warn("ADMIN_EMAIL and ADMIN_PASSWORD must both be provided or both be empty, skipping admin creation")
    } else if adminEmail != "" && adminPass != "" {
        admin, err := authSvc.EnsureAdminUser(ctx, adminEmail, adminPass)
        if err != nil {
            return fmt.Errorf("init admin user: %w", err)
        }
        log.Infof("Admin user ensured: %s", admin.Email)
    }
    
    return nil
}
```

### 目录结构

```
modules/auth/
└── services/
    ├── auth_service.go       # 现有文件
    └── bootstrap.go          # 新增：初始化方法
```

---

## Implementation Tasks

1. **Service 方法实现**
   - 在 `modules/auth/services/` 创建 `bootstrap.go`
   - 实现 `EnsureSystemUser()` 方法
   - 实现 `EnsureAdminUser()` 方法

2. **防止 system 用户登录**
   - 在 `modules/auth/services/auth_service.go` 的 `Login()` 方法中
   - 添加检查：如果 `user.IsSystem == true`，返回错误

3. **幂等性保证**
   - 使用数据库的 UNIQUE 约束（email, name）
   - 捕获重复插入错误，转换为查询逻辑

4. **日志记录**
   - 创建成功时记录 INFO 日志
   - 已存在时记录 DEBUG 日志
   - 错误时记录 ERROR 日志

5. **测试**
   - 单元测试：首次创建
   - 单元测试：幂等性（多次调用）
   - 单元测试：参数验证（空 email, 弱密码）
   - 集成测试：system 用户无法登录

6. **文档**
   - 更新 Auth Service 接口文档
   - 在 Story 1.17 中添加调用示例

---

## Testing Strategy

### 单元测试场景
- **EnsureSystemUser**:
  - 首次调用 → 创建成功
  - 再次调用 → 返回已有用户
  - UUID 固定为 `00000000-0000-0000-0000-000000000000`
  - `is_system=true`
  
- **EnsureAdminUser**:
  - 首次调用 → 创建成功
  - 再次调用 → 返回已有用户（不修改密码）
  - 空 email → 返回错误
  - 弱密码 → 返回错误
  - 角色为 `platform_admin`

### 集成测试场景
- System 用户尝试登录 → 返回 403 Forbidden
- 使用环境变量创建的 Admin 用户可以正常登录

---

## Data Model Changes

### 用户表增强

**Ent Schema** (`ent/schema/user.go`):
```go
import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

func (User) Fields() []ent.Field {
    return []ent.Field{
        // ... 现有字段
        field.Bool("is_system").
            Default(false).
            Comment("是否为系统用户"),
        field.Enum("role").
            Values("platform_user", "platform_admin").
            Default("platform_user").
            Comment("用户角色"),
    }
}

func (User) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("name").Unique(),
        index.Fields("is_system"),
    }
}
```

**Migration**:
```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'platform_user';
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_name ON users(name);
CREATE INDEX IF NOT EXISTS idx_users_is_system ON users(is_system);
```

---

## Environment Variables

在 `.env` 或环境变量中配置：

```bash
# 首个管理员账户（可选，仅在首次部署时需要）
# 注意：ADMIN_EMAIL 和 ADMIN_PASSWORD 必须同时提供或同时为空
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=ChangeMe123!
```

**验证规则**:
- 两个变量必须同时存在或同时为空
- 如果只提供其中一个 → 记录 WARN 日志并跳过 Admin 创建
- 在 Bootstrap 启动时验证（非在 EnsureAdminUser 中验证）

---

## Error Codes

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| `USER_INVALID_EMAIL` | 400 | 邮箱格式无效 |
| `USER_WEAK_PASSWORD` | 400 | 密码强度不足（引用 Story 5.1 规则） |
| `USER_SYSTEM_CANNOT_LOGIN` | 403 | System 用户不能登录 |
| `USER_DATA_CORRUPTION` | 500 | System 用户数据不一致（UUID 或 is_system 字段异常） |
| `INTERNAL_ERROR` | 500 | 内部错误（数据库操作失败） |

---

## Security Considerations

### System 用户安全
- System 用户仅用于标记资源所有权（如系统预置的配置项）
- 禁止通过任何 API 登录（无密码）
- 不赋予特殊权限（role=platform_user）

### Admin 用户安全
- 密码必须符合强度要求
- 仅在环境变量显式提供时创建
- 生产环境部署后应立即修改密码

### 方法安全
- 方法不暴露为 HTTP 接口
- 仅在 Bootstrap 阶段调用一次
- 记录所有创建操作的审计日志

---

## Documentation

- [Epic 5: 认证与授权](../../epics/5-auth-epic.md)
- [Story 1.17: 平台初始化](../sprint-2/1-17-platform-initialization.md)
- [编码规范](../../standards/coding-standards.md)
