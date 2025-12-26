# Epic: 认证与授权
# apprun BaaS Platform

**关联 PRD**: [FR-AUTH-001](../prd.md#21-认证与权限)  
**负责人**: Architect Agent  
**状态**: Planning  
**优先级**: P0 (必需)  
**预估工作量**: 3-4 周

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

- [ ] 用户可通过 Kratos 登录（Web 端 + API）
- [ ] JWT Token 正确签发和验证
- [ ] 项目级权限隔离正常工作
- [ ] 资源级权限控制生效
- [ ] API 响应时间 P95 < 100ms
- [ ] 单元测试覆盖率 > 80%

---

## 2. 技术规范

> 📖 **通用规范参考**：[API 设计规范](../standards/api-design.md) | [编码规范](../standards/coding-standards.md)

### 2.1 架构设计

#### 集成方式
- **Ory Kratos**: 用户身份管理（共享数据库）
- **JWT**: API 客户端认证
- **Casbin**: RBAC 权限引擎
- **中间件**: Chi Router 中间件链

#### 认证流程

**Web 端**：
```
用户 → Kratos UI → 登录成功 → Session Cookie
     → apprun API (携带 Cookie) → 验证 Session → 业务逻辑
```

**API 客户端**：
```
用户 → Kratos 登录 → Session
     → POST /api/v1/auth/token (携带 Session) → JWT Token
     → API 请求 (携带 JWT) → 验证 JWT → 业务逻辑
```

### 2.2 API 端点

| 端点 | 方法 | 功能 | 认证 |
|-----|------|------|------|
| `/api/v1/auth/token` | POST | 换取 JWT Token | Kratos Session |
| `/api/v1/auth/refresh` | POST | 刷新 Access Token | Refresh Token |
| `/api/v1/auth/me` | GET | 获取当前用户信息 | JWT |
| `/api/v1/auth/logout` | POST | 登出 | JWT |

#### 示例：换取 JWT Token

**请求**：
```http
POST /api/v1/auth/token
Cookie: ory_kratos_session=<session>
```

**响应**：
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "expires_in": 3600
  }
}
```

### 2.3 数据模型

#### 用户扩展表（apprun.users）
```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    identity_id VARCHAR(36) NOT NULL UNIQUE,  -- Kratos Identity ID
    email VARCHAR(255) NOT NULL,
    name VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
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

```ini
# config/casbin_model.conf
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

### 2.5 中间件设计

#### 认证中间件（伪代码）
```go
func AuthMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. 提取 Token（Cookie 或 Header）
            token := extractToken(r)
            
            // 2. 验证 Token
            userID, err := validateToken(token)
            if err != nil {
                response.Error(w, 401, "AUTH_INVALID_TOKEN", "Invalid token")
                return
            }
            
            // 3. 存入 Context
            ctx := context.WithValue(r.Context(), "user_id", userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
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
    secret: "${JWT_SECRET}"
    access_token_expire: 3600      # 1 小时
    refresh_token_expire: 604800   # 7 天
    algorithm: "HS256"
  
  kratos:
    public_url: "http://kratos:4433"
    admin_url: "http://kratos:4434"
    session_cookie_name: "ory_kratos_session"
  
  casbin:
    model_path: "./config/casbin_model.conf"
    policy_path: "./config/casbin_policy.csv"
```

---

## 3. Stories 拆分

### Story 1: Kratos 集成与 Session 验证
**优先级**: P0  
**工作量**: 3 天
- [ ] 集成 Kratos Public API
- [ ] 实现 Session Cookie 验证
- [ ] 实现用户信息同步（Kratos → apprun）
- [ ] 编写单元测试

### Story 2: JWT Token 管理
**优先级**: P0  
**工作量**: 2 天
- [ ] 实现 JWT 签发逻辑
- [ ] 实现 JWT 验证逻辑
- [ ] 实现 Token 刷新机制
- [ ] 实现 `/api/v1/auth/token` 端点
- [ ] 编写单元测试

### Story 3: 认证中间件
**优先级**: P0  
**工作量**: 2 天
- [ ] 实现 AuthMiddleware
- [ ] 集成到路由系统
- [ ] 处理多种 Token 来源（Cookie, Header）
- [ ] 编写集成测试

### Story 4: RBAC 权限控制
**优先级**: P0  
**工作量**: 4 天
- [ ] 集成 Casbin
- [ ] 实现项目成员管理
- [ ] 实现 RequirePermission 中间件
- [ ] 定义权限策略
- [ ] 编写权限测试用例

### Story 5: 用户信息接口
**优先级**: P1  
**工作量**: 1 天
- [ ] 实现 `/api/v1/auth/me` 端点
- [ ] 实现 `/api/v1/auth/logout` 端点
- [ ] 编写 API 文档

---

## 4. 依赖关系

### 技术依赖
- Ory Kratos (外部服务)
- Casbin v2 (Go 库)
- JWT 库 (github.com/golang-jwt/jwt/v5)

### 模块依赖
- 数据库模块（Ent ORM）
- 配置模块（Viper）
- 日志模块（Logrus）

### 外部依赖
- PostgreSQL 14+
- Redis 7+ (可选，缓存权限)

---

## 5. 风险与挑战

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| Kratos Session 验证性能 | 中 | 使用 Redis 缓存 Session 数据 |
| JWT Secret 泄露 | 高 | 使用环境变量，定期轮换 |
| Casbin 策略复杂度 | 中 | 从简单策略开始，逐步扩展 |
| 多租户权限隔离 | 高 | 严格测试权限边界 |

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

## 附录

### A. 错误码定义

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| `AUTH_INVALID_TOKEN` | 401 | Token 无效或已过期 |
| `AUTH_SESSION_NOT_FOUND` | 401 | Kratos Session 不存在 |
| `PERM_FORBIDDEN` | 403 | 无权限访问 |
| `PERM_PROJECT_NOT_MEMBER` | 403 | 不是项目成员 |

### B. 相关文档

- [PRD - 认证与权限](../prd.md#21-认证与权限)
- [API 设计规范](../standards/api-design.md)
- [技术架构文档](../architecture/tech-architecture.md)

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26
