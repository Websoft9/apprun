# Story 20: JWT 认证中间件
# Sprint 2: 认证与授权 (Auth Epic)

**Priority**: P0 (必需)  
**Effort**: 1 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 18 (User Registration) - 已完成
- Story 19 (User Login) - 已完成

**Status**: Done  
**Module**: Authentication  
**Epic**: [auth-epic](../../epics/auth-epic.md)  
**Issue**: #TBD  

---

## User Story

作为 **apprun 平台开发者**，我希望有一个可复用的 JWT 认证中间件，能够验证请求中的 JWT Token，提取用户身份信息，并将其注入到请求上下文中，以便后续 Handler 可以获取当前登录用户信息。

---

## Acceptance Criteria

### Functional Acceptance
- [x] Implement JWT authentication middleware
- [x] Extract token from HTTP Header `Authorization: Bearer <token>`
- [x] Validate token signature and expiration
- [x] Extract user ID from token and inject into Context
- [x] Support configurable JWT Secret and expiration time
- [x] Return 401 error for invalid tokens
- [x] Return 403 error for expired tokens
- [x] **Whitelist configuration from YAML/Config Center**

### Non-Functional Acceptance
- [x] Unit test coverage ≥ 80% (achieved 91.0%)
- [x] Support concurrent requests

### Security Acceptance
- [x] JWT Secret loaded from environment variables, not hardcoded
- [x] Token validation errors do not expose detailed information
- [x] **Whitelist paths configurable, not hardcoded**

---

## Technical Design

### Chi 中间件标准

#### 中间件接口
Chi 使用标准的 Go HTTP 中间件接口，所有中间件必须符合：
```go
type Middleware func(http.Handler) http.Handler
```

#### 中间件职责
1. 接收 `next http.Handler`
2. 返回新的 `http.Handler`
3. 执行前置逻辑（验证 Token）
4. 注入 Context
5. 调用 `next.ServeHTTP()` 或返回错误

#### 中间件顺序
```
RequestID → Logger → Recoverer → JWTAuth → Timeout → Handler
```

### 核心组件

#### 1. Context Key 定义
```go
type contextKey struct{ name string }
var UserIDKey = &contextKey{"user_id"}
```

| Key | 类型 | 说明 |
|-----|------|------|
| `user_id` | int64 | 用户 ID |
| `username` | string | 用户名 |
| `email` | string | 邮箱 |

**注意**：使用私有类型 Context Key 避免冲突

#### 2. JWT Claims 结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `user_id` | int64 | 用户 ID |
| `username` | string | 用户名 |
| `email` | string | 邮箱 |
| `exp` | int64 | 过期时间戳 |
| `iat` | int64 | 签发时间戳 |
| `iss` | string | 签发者 |

#### 3. JWT Token 工具

**核心函数**：
- `GenerateToken(userID int64, username, email string, expiry time.Duration) (string, error)`
- `ValidateToken(tokenString string) (*Claims, error)`

**辅助函数**：
- `GetUserID(ctx context.Context) int64` - 从 Context 获取用户 ID
- `GetUsername(ctx context.Context) string` - 从 Context 获取用户名

**配置项**：
- `JWT_SECRET` - 签名密钥（环境变量，≥32 字符）
- `JWT_EXPIRY` - 过期时间（默认 24 小时）

#### 4. 认证中间件

**接口定义**：
```go
func JWTAuth(next http.Handler) http.Handler
```

**核心职责**：
1. 提取 Authorization Header
2. 验证 Token 格式（Bearer <token>）
3. 验证 Token 签名和有效期
4. 提取 Claims 并注入 Context
5. 调用 `next.ServeHTTP()` 或返回错误（使用 core/pkg/response 标准错误）

**跳过路径**（白名单）- **从配置加载，非硬编码**：
- `/api/v1/auth/register`
- `/api/v1/auth/login`
- `/health`
- `/metrics`

**工厂函数（推荐使用）**：
```go
// From Config struct (recommended)
func NewJWTMiddlewareFromConfig(cfg *jwt.Config) (*JWTMiddleware, error)

// From RuntimeConfig (for direct use)
func NewJWTMiddleware(runtimeCfg *jwt.RuntimeConfig) *JWTMiddleware
```

---

## API Behavior

### 1. 受保护的 API（需要认证）

#### 请求示例
```http
GET /api/v1/auth/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 成功场景
- Token 有效 → 继续执行 Handler
- Context 中包含 `user_id`

#### 失败场景

**401 Unauthorized - Token 缺失或格式错误**
```json
{
  "success": false,
  "error": {
    "code": "AUTH_UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

**401 Unauthorized - Token 签名无效**
```json
{
  "success": false,
  "error": {
    "code": "AUTH_INVALID_TOKEN",
    "message": "Invalid token"
  }
}
```

**403 Forbidden - Token 已过期**
```json
{
  "success": false,
  "error": {
    "code": "AUTH_TOKEN_EXPIRED",
    "message": "Token has expired"
  }
}
```

### 2. 公开 API（无需认证）

以下路径不经过 JWT 中间件验证：
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /health`
- `GET /metrics`

---

## Configuration

### 环境变量

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `JWT_SECRET` | string | - | JWT 签名密钥（必填，≥32 字符） |
| `JWT_EXPIRY` | duration | 24h | Token 过期时间 |
| `JWT_ISSUER` | string | apprun | Token 签发者 |

### 配置文件示例

```yaml
# config/default.yaml
jwt:
  secret: ${JWT_SECRET}           # 从环境变量读取
  expiry: 24h                     # 24 小时过期
  issuer: "apprun"
  whitelist_paths:                # 白名单路径（可配置）
    - /api/v1/auth/register
    - /api/v1/auth/login
    - /health
    - /metrics
```

---

## Usage Example

### 1. Chi 路由配置

```go
r := chi.NewRouter()

// 全局中间件
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

// 加载JWT配置并创建中间件
jwtCfg := &jwt.Config{
    Secret:         os.Getenv("JWT_SECRET"),
    Expiry:         "24h",
    Issuer:         "apprun",
    WhitelistPaths: []string{"/api/v1/auth/register", "/api/v1/auth/login", "/health", "/metrics"},
}
jwtMiddleware, _ := middleware.NewJWTMiddlewareFromConfig(jwtCfg)

r.Route("/api/v1", func(r chi.Router) {
    // 公开路由（无需认证）
    r.Post("/auth/register", authHandler.Register)
    r.Post("/auth/login", authHandler.Login)
    
    // 受保护路由（需要认证）
    r.Group(func(r chi.Router) {
        r.Use(jwtMiddleware.JWTAuth)  // 应用 JWT 中间件
        r.Get("/auth/me", authHandler.GetMe)
    })
})
```

### 2. Handler 中获取用户信息

```go
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
    userID := jwt.GetUserID(r.Context())      // 使用辅助函数
    username := jwt.GetUsername(r.Context())  // 类型安全
    // ...
}
```

### 3. 登录时生成 Token（Story 19）

```go
token, _ := jwt.GenerateToken(user.ID, user.Username, user.Email, 24*time.Hour)
```

---

## Dependencies

### Go 包依赖
```go
require (
    github.com/golang-jwt/jwt/v5 v5.2.0   // JWT 库
    github.com/go-chi/chi/v5 v5.0.11      // HTTP Router
)
```

### 框架模块依赖
- **core/pkg/response** - 标准响应格式
- **core/pkg/config** - 配置管理

### Story 依赖关系
- **Story 18** - 用户注册（提供 User 表）
- **Story 19** - 用户登录（生成 JWT Token）

### 业务模块（本 Story 交付）
- `core/internal/jwt` - JWT 工具
- `modules/auth/middleware` - 认证中间件

---

## Deliverables

### 框架核心层
1. `core/internal/jwt/claims.go` - Claims 结构定义
2. `core/internal/jwt/token.go` - Token 生成和验证
3. `core/internal/jwt/token_test.go` - 单元测试

### 业务模块层
4. `modules/auth/middleware/auth.go` - JWT 认证中间件
5. `modules/auth/middleware/auth_test.go` - 中间件测试

### 配置与文档
6. `config/default.yaml` - JWT 配置项
7. `.env.example` - 环境变量示例

---

## Risks and Mitigations

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| JWT Secret 泄露 | 高 | 环境变量管理，定期轮换 Secret |
| Token 被窃取 | 中 | 仅通过 HTTPS，短过期时间 |
| 并发安全 | 低 | JWT 验证是无状态的，天然支持并发 |

---

## Definition of Done

- [x] 所有验收标准通过
- [x] 单元测试覆盖率 = 91.0% (JWT: 86.5%, Middleware: 96.7%)
- [x] 集成测试通过（有效/无效/过期 Token）
- [x] 代码审查完成
- [x] 配置文档完整
- [x] **白名单从配置加载，非硬编码**
- [x] **工厂函数NewJWTMiddlewareFromConfig()提供优雅初始化**
- [x] **所有代码注释已英文化**
- [x] Story 20 配置重构完成（Registry Pattern）
- [x] 部署到开发环境并验证

---

## References

- [JWT 规范 RFC 7519](https://tools.ietf.org/html/rfc7519)
- [golang-jwt/jwt 文档](https://github.com/golang-jwt/jwt)
- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [Epic: 认证与授权](../../epics/auth-epic.md)

---

**Story Owner**: Backend Dev Team  
**Created**: 2026-01-08  
**Last Updated**: 2026-01-08  
**Sprint**: Sprint 2
