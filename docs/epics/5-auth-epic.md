# Epic 5: 认证与授权
# apprun BaaS Platform

**Epic ID**: epic-5  
**关联 PRD**: [FR-AUTH-001](../prd.md#21-认证与权限)  
**负责人**: Architect Agent  
**状态**: Planning  
**优先级**: P0 (必需)  
**预估工作量**: 11.5 天 (MVP)

---

## 1. Epic 概述

### 1.1 业务目标

实现完整的用户认证和基于项目的权限控制体系，支持 Web 端和 API 客户端。

### 1.2 核心价值

- 用户可以安全登录和访问系统
- 项目间权限完全隔离
- 支持多种客户端类型（浏览器、移动端、API）
- 细粒度的资源访问控制

### 1.3 验收标准

- [ ] 用户可通过 API 注册和登录（Email + Password）
- [ ] JWT Token 正确签发和验证
- [ ] 密码安全存储（bcrypt hashing）
- [ ] 项目级权限隔离正常工作
- [ ] 资源级权限控制生效
- [ ] API 响应时间 P95 < 100ms
- [ ] 单元测试覆盖率 > 80%

---

## 2. 技术规范

> 📖 **通用规范参考**：[API 设计规范](../standards/api-design.md) | [编码规范](../standards/coding-standards.md)

### 2.1 架构设计

#### 核心理念：虚拟资源中心 (Virtual Resource Center)
AppRun 采用 **逻辑多租户** 架构，而非物理隔离。
- **Project** 是权限和资源的引力中心。
- **RBAC with Domains**: 采用 Casbin 的 `(sub, dom, obj, act)` 模型，其中 `dom` 为 `project_id`。
- **Context Isolation**: 身份认证解决 "你是谁"，项目上下文解决 "你在哪里"，权限引擎解决 "你能做什么"。

#### 技术栈
- **bcrypt**: 密码哈希 (golang.org/x/crypto/bcrypt)
- **JWT**: Token 认证 (github.com/golang-jwt/jwt/v5)
- **Casbin**: RBAC 权限引擎
- **中间件**: Chi Router 中间件链
- **Session Store**: 可选的 Cookie Session (github.com/gorilla/sessions)

#### 认证流程

**注册流程**：
```
用户 → POST /api/auth/register (email, password, name)
     → 验证邮箱格式
     → 密码强度检查
     → bcrypt.GenerateFromPassword
     → 存储到数据库
     → 返回 User 对象
```

**登录流程**：
```
用户 → POST /api/auth/login (email, password)
     → 查询用户
     → bcrypt.CompareHashAndPassword
     → 生成 JWT Token (access + refresh)
     → 返回 Token
```

**API 访问流程**：
```
客户端 → API 请求 (Authorization: Bearer <JWT>)
       → AuthMiddleware 验证 Token
       → 解析 Claims (user_id, project_id)
       → 存入 Context
       → RequirePermission 检查权限
       → 业务逻辑
```

### 2.2 API 端点

| 端点 | 方法 | 功能 | 认证 |
|-----|------|------|------|
| `/api/auth/register` | POST | 用户注册 | Public |
| `/api/auth/login` | POST | 用户登录 | Public |
| `/api/auth/refresh` | POST | 刷新 Access Token | Refresh Token |
| `/api/users/me` | GET | 获取当前用户信息 | JWT |
| `/api/auth/logout` | POST | 登出（可选） | JWT |
| `/api/users/me/password` | PUT | 修改密码 | JWT |

#### 示例 1：用户注册

**请求**：
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**响应**：
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "John Doe",
      "created_at": "2026-01-08T10:00:00Z"
    }
  }
}
```

#### 示例 2：用户登录

**请求**：
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**响应**：
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "John Doe"
    }
  }
}
```

### 2.3 数据模型

#### 用户表（apprun.users）
```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,  -- bcrypt hash
    name VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_login_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

#### 项目成员表（apprun.project_members）
```sql
CREATE TABLE project_members (
    id VARCHAR(36) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    role VARCHAR(20) NOT NULL,  -- owner, admin, member, viewer
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, user_id)
);
```

### 2.4 权限模型

#### RBAC 角色定义

**全局角色**：
- `platform_admin`: 平台管理员（所有权限）
- `platform_user`: 普通用户（创建项目、加入项目）

**项目角色**：
- `owner`: 项目所有者（所有权限）
- `admin`: 项目管理员（管理成员、配置）
- `member`: 项目成员（读写数据）
- `viewer`: 查看者（只读）

#### Casbin 策略配置

Model 文件内嵌在 `core/internal/rbac/model.conf`，使用 `go:embed` 指令，避免对配置中心的依赖。

```ini
# core/internal/rbac/model.conf (内嵌资源)
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

### 2.5 中间件设计

#### 认证中间件（实现参考）
```go
func AuthMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. 提取 Token（Authorization Header）
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                response.Error(w, 401, "AUTH_MISSING_TOKEN", "Missing authorization token")
                return
            }
            
            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            
            // 2. 验证 JWT Token
            token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("unexpected signing method")
                }
                return jwtSecret, nil
            })
            
            if err != nil || !token.Valid {
                response.Error(w, 401, "AUTH_INVALID_TOKEN", "Invalid token")
                return
            }
            
            // 3. 提取 Claims 并存入 Context
            claims, ok := token.Claims.(*Claims)
            if !ok {
                response.Error(w, 401, "AUTH_INVALID_CLAIMS", "Invalid token claims")
                return
            }
            
            ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
            ctx = context.WithValue(ctx, "email", claims.Email)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// JWT Claims 结构
type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}
```

#### 权限验证中间件（伪代码）
```go
func RequirePermission(resource, action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := r.Context().Value("user_id").(string)
            projectID := chi.URLParam(r, "project_id")
            
            // Casbin 权限检查
            allowed := enforcer.Enforce(userID, projectID, resource, action)
            if !allowed {
                response.Error(w, 403, "PERM_FORBIDDEN", "Permission denied")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### 2.6 配置

```yaml
# config/auth.yaml
auth:
  jwt:
    secret: "${JWT_SECRET}"              # 环境变量，生产环境必须设置
    access_token_expire: 3600            # 1 小时
    refresh_token_expire: 604800         # 7 天
    algorithm: "HS256"
  
  password:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_number: true
    require_special: false
    bcrypt_cost: 12                      # bcrypt 计算成本 (10-14)
  
  rate_limit:
    login_attempts: 5                    # 5 次失败后锁定
    lockout_duration: 900                # 锁定 15 分钟
  
  casbin:
    # model 已内嵌到 core/internal/rbac/model.conf
    policy_path: "./config/casbin_policy.csv"
```

---

## 3. Stories 拆分

### Story 5.1: 用户注册与密码安全
**优先级**: P0  
**工作量**: 1.5 天  
**文件**: [5-1-user-registration.md](../sprint-artifacts/sprint-1/5-1-user-registration.md)

- [x] 实现用户注册 API (`POST /api/auth/register`)
- [x] 邮箱格式验证
- [x] 密码强度验证（长度、复杂度）
- [x] bcrypt 密码哈希存储
- [x] 用户数据模型（Ent Schema）
- [x] 编写单元测试

### Story 5.2: 用户登录与 JWT 签发
**优先级**: P0  
**工作量**: 1.5 天  
**文件**: [5-2-user-login.md](../sprint-artifacts/sprint-1/5-2-user-login.md)

- [ ] 实现登录 API (`POST /api/auth/login`)
- [ ] 密码验证（bcrypt.CompareHashAndPassword）
- [ ] JWT Token 签发（Access + Refresh）
- [ ] Token Claims 设计（user_id, email, exp）
- [ ] 更新最后登录时间
- [ ] 编写单元测试

### Story 5.3: JWT 认证中间件
**优先级**: P0  
**工作量**: 1 天  
**文件**: [5-3-jwt-middleware.md](../sprint-artifacts/sprint-1/5-3-jwt-middleware.md)

- [ ] 实现 AuthMiddleware（Token 提取和验证）
- [ ] 集成到 Chi Router
- [ ] Context 注入（user_id, email）
- [ ] 错误处理（401 响应）
- [ ] 编写集成测试

### Story 5.4: Token 刷新机制
**优先级**: P0  
**工作量**: 1 天

- [ ] 实现 Refresh Token 逻辑
- [ ] 实现 `/api/auth/refresh` 端点
- [ ] Refresh Token 黑名单（可选，Redis）
- [ ] 编写单元测试

### Story 5.5: RBAC 权限控制
**优先级**: P0  
**工作量**: 2 天

- [ ] 集成 Casbin
- [ ] 实现项目成员管理
- [ ] 实现 RequirePermission 中间件
- [ ] 定义权限策略
- [ ] 编写权限测试用例

### Story 5.6: 用户自我管理 (User Self-Service)
**优先级**: P1  
**工作量**: 1 天

- [ ] 实现 `/api/auth/me` 端点（获取当前用户信息）
- [ ] 实现 `/api/users/me` 端点（查看自己的详细资料）
- [ ] 实现 `PUT /api/users/me` 端点（修改自己的资料：name, avatar 等）
- [ ] 实现 `/api/auth/change-password` 端点（修改自己的密码）
- [ ] 实现 `/api/auth/logout`（可选，客户端删除 Token）
- [ ] 输入验证（防止修改敏感字段如 role, is_active）
- [ ] 编写单元测试
- [ ] 编写 API 文档

### Story 5.7: 平台用户管理 (Platform User Management)
**优先级**: P1  
**工作量**: 2 天

**目标受众**: 仅限 `platform_admin` 角色访问

**API 端点**:
- [ ] `GET /api/admin/users` - 列出所有用户（支持分页、搜索、角色过滤）
- [ ] `POST /api/admin/users` - 创建新用户/管理员（指定 email, name, role）
- [ ] `PUT /api/admin/users/:id/role` - 修改用户角色（升职/降职）
- [ ] `PUT /api/admin/users/:id/status` - 修改用户状态（启用/禁用/封禁）
- [ ] `GET /api/admin/users/:id` - 查看指定用户详细信息
- [ ] `DELETE /api/admin/users/:id` - 删除用户（软删除）

**功能要求**:
- [ ] 实现 `RequirePlatformAdmin` 中间件（权限拦截）
- [ ] 防止管理员删除自己
- [ ] 防止降级最后一个 platform_admin
- [ ] 审计日志（记录所有管理操作）
- [ ] 输入验证和安全检查
- [ ] 编写单元测试和集成测试
- [ ] 编写 API 文档

### Story 5.8: 系统用户初始化支持 (System User Initialization Support)
**优先级**: P0  
**工作量**: 0.5 天  
**依赖**: Story 5.1 (User Schema), Story 1.17 (Platform Initialization) 将调用此能力

**目标**: 为 Bootstrap 模块提供创建系统内置账号的 Service 方法

**Service 方法**:
- [ ] 实现 `AuthService.EnsureSystemUser(ctx)` 方法
  - 检查是否存在 `name="system"` 用户
  - 不存在则创建（固定 UUID: `00000000-0000-0000-0000-000000000000`, 无密码, 特殊标记 `is_system=true`）
  - 禁止通过 Login API 登录
- [ ] 实现 `AuthService.EnsureAdminUser(ctx, email, password)` 方法
  - 检查是否存在该 email 的用户
  - 不存在则创建（生成 UUID, bcrypt 密码, `platform_admin` 角色）
  - 存在则跳过（幂等性）
- [ ] 幂等性保证（多次调用安全，不会重复创建）
- [ ] 编写单元测试（幂等性、边界条件、错误处理）
- [ ] 更新 Auth Service 接口文档

**验收标准**:
- [ ] 方法幂等：连续调用 2 次返回相同结果
- [ ] System 用户无法通过 `POST /api/auth/login` 登录
- [ ] Admin 用户可正常登录获取 Token
- [ ] Bootstrap 模块可成功调用这些方法

### Story 5.9: 审计日志中间件 (Audit Logging Middleware)
**优先级**: P1  
**工作量**: 1 天  
**依赖**: Story 5.3 (JWT 中间件)

**目标**: 自动记录所有关键操作，支持安全审计和合规性要求

**功能要求**:
- [ ] 实现 `AuditMiddleware` 自动记录 HTTP 请求
- [ ] 实现 `AuditService.Log()` 和 `LogAction()` 方法
- [ ] 捕获关键字段（时间戳、操作者、操作类型、目标资源、IP、User-Agent）
- [ ] 支持异步写入（避免阻塞业务）
- [ ] 创建 `AuditLog` 数据模型（Ent Schema）
- [ ] 支持数据库和文件两种存储方式（可配置）
- [ ] 自动记录认证事件、权限拒绝、资源变更、管理操作
- [ ] 敏感字段自动脱敏（password, token, secret）
- [ ] 编写单元测试和集成测试

**验收标准**:
- [ ] 所有 HTTP 请求自动记录（排除 /health, /metrics）
- [ ] 日志写入延迟 < 50ms (P95)
- [ ] 管理操作（Story 5.7）正确调用审计日志
- [ ] 审计日志包含完整追溯信息

---

## 4. 依赖关系

### 技术依赖（Go 包）
- `golang.org/x/crypto/bcrypt` - 密码哈希
- `github.com/golang-jwt/jwt/v5` - JWT Token
- `github.com/casbin/casbin/v2` - RBAC 引擎
- `github.com/gorilla/sessions` - Session 管理（可选）
- `golang.org/x/time/rate` - 速率限制（可选）

### 模块依赖
- 数据库模块（Ent ORM）
- 配置模块（Viper）
- 日志模块（Logrus）
- Response 标准化模块

### 外部依赖
- PostgreSQL 14+ (用户数据存储)
- Redis 7+ (可选：Token 黑名单、权限缓存)

---

## 5. 风险与挑战

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| JWT Secret 泄露 | 高 | 使用环境变量，定期轮换，考虑 RSA 签名 |
| 密码暴力破解 | 高 | 速率限制、账户锁定、bcrypt 高成本 |
| Token 过期处理 | 中 | 实现完善的 Refresh Token 机制 |
| Casbin 策略复杂度 | 中 | 从简单策略开始，逐步扩展 |
| 多租户权限隔离 | 高 | 严格测试权限边界，Casbin 策略审计 |
| 缺少 MFA 支持 | 中 | MVP 阶段可接受，Post-MVP 添加 TOTP |

---

## 6. 测试策略

### 单元测试
- JWT 签发和验证逻辑
- Casbin 策略匹配
- 中间件功能测试

### 集成测试
- 完整认证流程（登录 → Token → API 调用）
- 权限验证场景（正常访问、拒绝访问）
- Token 刷新流程

### 性能测试
- 认证中间件延迟 < 10ms
- 权限检查延迟 < 5ms
- 并发登录场景

---

## 7. 监控指标

- `auth_token_generated_total` - Token 签发总数
- `auth_token_validation_failed_total` - Token 验证失败次数
- `auth_permission_denied_total` - 权限拒绝次数
- `auth_session_validation_duration_seconds` - Session 验证耗时

---

## 8. 安全最佳实践

### 8.1 JWT Secret 管理
- **生产环境**：使用强随机 Secret（至少 32 字节）
- **轮换策略**：每 90 天轮换一次
- **存储方式**：环境变量或密钥管理服务（如 AWS Secrets Manager）
- **多环境隔离**：开发/测试/生产使用不同 Secret

### 8.2 HTTPS 强制
- 生产环境必须启用 HTTPS
- 使用 HSTS Header 强制浏览器使用 HTTPS
- JWT Token 仅通过 HTTPS 传输

### 8.3 CORS 配置
- 明确允许的 Origin 列表
- 不允许使用通配符 `*`
- 正确设置 Credentials 模式

### 8.4 输入验证
- 所有用户输入必须验证和消毒
- 使用参数化查询防止 SQL 注入
- 输出转义防止 XSS

### 8.5 速率限制
- 登录端点：5 次/15 分钟（每 IP）
- 注册端点：3 次/小时（每 IP）
- Token 刷新：10 次/小时（每用户）

### 8.6 审计日志
- 记录所有认证事件（成功/失败）
- 记录权限拒绝事件
- 记录资源变更操作（POST/PUT/DELETE）
- 记录管理操作（用户角色变更、删除用户）
- 日志包含：时间戳、操作者 ID、操作类型、目标资源、变更内容、IP、User-Agent
- 敏感字段自动脱敏
- 支持异步写入（不阻塞业务）

---

## 附录

### A. 错误码定义

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| `AUTH_INVALID_CREDENTIALS` | 401 | 邮箱或密码错误 |
| `AUTH_INVALID_TOKEN` | 401 | Token 无效或已过期 |
| `AUTH_MISSING_TOKEN` | 401 | 缺少认证 Token |
| `AUTH_INVALID_CLAIMS` | 401 | Token Claims 无效 |
| `AUTH_ACCOUNT_LOCKED` | 403 | 账户已锁定（登录失败次数过多） |
| `AUTH_EMAIL_EXISTS` | 409 | 邮箱已被注册 |
| `AUTH_WEAK_PASSWORD` | 400 | 密码强度不足 |
| `PERM_FORBIDDEN` | 403 | 无权限访问 |
| `PERM_PROJECT_NOT_MEMBER` | 403 | 不是项目成员 |

### B. 相关文档

- [PRD - 认证与权限](../prd.md#21-认证与权限)
- [API 设计规范](../standards/api-design.md)
- [技术架构文档](../architecture/tech-architecture.md)

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26
