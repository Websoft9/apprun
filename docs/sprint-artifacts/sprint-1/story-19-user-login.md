# Story 19: User Login
# Sprint 2: 认证与授权 (Auth Epic)

**Priority**: P0 (必需)  
**Effort**: 1.5 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 18 (User Registration) - 已完成
- Story 20 (JWT Middleware) - 已完成

**Status**: Todo  
**Module**: Authentication  
**Epic**: [auth-epic](../../epics/auth-epic.md)  
**Issue**: #TBD  

---

## User Story

作为 **apprun 平台用户**，我希望能够使用用户名或邮箱登录系统，获取 JWT Token，并能查询我的个人信息，以便访问受保护的 API。

---

## Acceptance Criteria

### 功能验收
- [ ] 实现 POST /api/v1/auth/login 端点
- [ ] 实现 GET /api/v1/auth/me 端点
- [ ] 支持用户名或邮箱作为登录标识
- [ ] 密码验证使用 bcrypt
- [ ] 成功登录返回 JWT Token 和用户信息
- [ ] 更新用户最后登录时间和 IP
- [ ] /auth/me 返回完整用户信息（需 JWT 认证）
- [ ] 错误处理：用户不存在、密码错误、账户状态异常

### 非功能验收
- [ ] 登录响应时间 < 500ms
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 支持并发登录

### 安全验收
- [ ] 密码验证失败不暴露具体原因
- [ ] 登录失败记录日志（不含敏感信息）
- [ ] 支持账户锁定（可选，未来扩展）

---

## Technical Design

### 架构分层

```
HTTP Request (POST /auth/login)
    ↓
Chi Router
    ↓
Auth Handler (验证凭证)
    ↓
Ent Repository (查询用户)
    ↓
JWT Tool (生成 Token)
    ↓
Response (Token + 用户信息)
```

### 核心组件

#### 1. 登录请求结构

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `identifier` | string | 是 | 用户名或邮箱 |
| `password` | string | 是 | 密码 |

#### 2. 登录响应结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `user.id` | int64 | 用户 ID |
| `user.username` | string | 用户名 |
| `user.email` | string | 邮箱 |
| `token` | string | JWT Token |
| `expires_at` | string | Token 过期时间 |

#### 3. 业务逻辑

**登录流程**:
1. 解析请求体
2. 根据 identifier 查询用户（用户名或邮箱）
3. 验证密码哈希
4. 检查用户状态（active）
5. 生成 JWT Token
6. 更新登录历史
7. 返回 Token 和用户信息

**错误场景**:
- 用户不存在 → 401 Unauthorized
- 密码错误 → 401 Unauthorized
- 账户禁用 → 403 Forbidden

---

## API Behavior

### 1. 用户登录 API

#### 请求示例

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "identifier": "johndoe",
  "password": "secure_password"
}
```

### 成功响应

```json
{
  "success": true,
  "data": {
    "user": {
      "id": 123,
      "username": "johndoe",
      "email": "john@example.com"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-01-09T10:00:00Z"
  }
}
```

### 错误响应

**401 Unauthorized - 凭证无效**
```json
{
  "success": false,
  "error": {
    "code": "AUTH_INVALID_CREDENTIALS",
    "message": "Invalid credentials"
  }
}
```

**403 Forbidden - 账户禁用**
```json
{
  "success": false,
  "error": {
    "code": "AUTH_ACCOUNT_DISABLED",
    "message": "Account is disabled"
  }
}
```

---

### 2. 获取当前用户信息 API

#### Endpoint

```http
GET /api/v1/auth/me
Authorization: Bearer <JWT_TOKEN>
```

#### Response

##### ✅ 成功 (200 OK)

```json
{
  "success": true,
  "data": {
    "id": 123,
    "username": "johndoe",
    "email": "john@example.com",
    "nickname": "John Doe",
    "avatar": "https://cdn.example.com/avatars/123.jpg",
    "phone": "+86 13800138000",
    "gender": 1,
    "signature": "Life is beautiful",
    "status": 1,
    "last_login_at": "2026-01-08T09:30:00Z",
    "timezone": "Asia/Shanghai",
    "language": "zh-CN",
    "created_at": "2026-01-08T10:00:00Z"
  }
}
```

##### ❌ 错误响应

**401 Unauthorized**
```json
{
  "success": false,
  "error": {
    "code": "AUTH_UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

---

## Implementation Steps

### Phase 1: Handler 实现 (3 小时)
- [ ] 创建 `modules/auth/handler/login.go`
- [ ] 创建 `modules/auth/handler/me.go`
- [ ] 实现请求解析和验证
- [ ] 集成 Ent 用户查询
- [ ] 实现密码验证逻辑

### Phase 2: 业务逻辑 (2.5 小时)
- [ ] 实现登录历史更新
- [ ] 集成 JWT Token 生成
- [ ] 添加用户状态检查
- [ ] /auth/me JWT 认证集成
- [ ] 编写单元测试

### Phase 3: 集成测试 (1.5 小时)
- [ ] 集成到 Chi Router
- [ ] 测试成功登录场景
- [ ] 测试 /auth/me 认证流程
- [ ] 测试错误场景（无效凭证、禁用账户）
- [ ] 性能测试

### Phase 4: 文档与配置 (1 小时)
- [ ] 更新路由配置
- [ ] 添加 API 文档
- [ ] 代码审查

**总工时**: ~8 小时 ≈ **1.5 天**

---

## Testing Strategy

### 单元测试
- 有效凭证登录成功
- 无效用户名返回 401
- 错误密码返回 401
- 禁用账户返回 403
- Token 生成正确
- 登录历史更新
- /auth/me 认证成功返回用户信息
- /auth/me 无效 Token 返回 401

### 集成测试

```bash
# 成功登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"johndoe","password":"password"}'

# 获取当前用户信息
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <JWT_TOKEN>"

# 无效凭证
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"invalid","password":"wrong"}'
```

---

## Dependencies

### Go 包依赖
```go
require (
    golang.org/x/crypto/bcrypt v0.17.0  // 密码哈希
    github.com/go-chi/chi/v5 v5.0.11    // HTTP Router
)
```

### 框架模块依赖
- **core/pkg/response** - 标准响应格式
- **core/internal/jwt** - JWT Token 生成 (Story 20)
- **ent** - 用户数据访问

### Story 依赖关系
- **Story 18** - 用户注册（提供 User 表）
- **Story 20** - JWT 中间件（提供 Token 生成工具）

### 业务模块（本 Story 交付）
- `modules/auth/handler/login.go` - 登录 Handler
- `modules/auth/handler/me.go` - 获取当前用户 Handler
- `modules/auth/service/auth.go` - 认证服务

---

## Deliverables

1. `modules/auth/handler/login.go` - 登录 API Handler
2. `modules/auth/handler/me.go` - 获取当前用户 Handler
3. `modules/auth/service/auth.go` - 认证业务逻辑
4. `modules/auth/handler/login_test.go` - Handler 测试
5. `modules/auth/service/auth_test.go` - 服务测试
6. API 文档更新（2 个端点）

---

## Risks and Mitigations

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 暴力破解 | 中 | 密码验证失败延迟，账户锁定 |
| 凭证泄露 | 高 | HTTPS 传输，密码哈希存储 |
| 并发登录 | 低 | 无状态设计，支持并发 |

---

## Definition of Done

- [ ] 所有验收标准通过
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 集成测试通过（登录 + /auth/me）
- [ ] 代码审查完成
- [ ] 性能测试达标（登录 < 500ms，/me < 100ms）
- [ ] Story 20 的 JWT 集成成功
- [ ] 部署到开发环境并验证

---

## References

- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [bcrypt 密码哈希](https://en.wikipedia.org/wiki/Bcrypt)
- [Epic: 认证与授权](../../epics/auth-epic.md)

---

**Story Owner**: Backend Dev Team  
**Created**: 2026-01-08  
**Last Updated**: 2026-01-08  
**Sprint**: Sprint 2</content>
<parameter name="filePath">/data/cdl/apprun/docs/sprint-artifacts/sprint-1/story-19-user-login.md