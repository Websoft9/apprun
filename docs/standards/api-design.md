# API Design Standards
# apprun BaaS Platform

**Version**: 2.0  
**Status**: Active  
**Last Updated**: 2026-01-09

> 本文档仅记录项目特有的 API 设计决策。通用 REST 规范参考 [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html)

---

## 1. 架构决策

### 1.1 API 版本策略
**决策**: 目前不使用版本化（快速 MVP）
```
✅ GET /api/users
❌ GET /api/v1/users
```

### 1.2 资源层级
**限制**: 最多 3 层嵌套
```
✅ /api/projects/123/files
❌ /api/projects/123/models/456/fields/789/validations
```

---

## 2. 认证方式

| 场景 | 方式 | Header |
|------|------|--------|
| Web 浏览器 | Session Cookie | `Cookie: ory_kratos_session=xxx` |
| API 客户端 | JWT | `Authorization: Bearer xxx` |
| 服务调用 | API Key | `X-API-Key: sk_live_xxx` |

**API Key 格式**: `{prefix}_{env}_{random}` (例: `sk_live_abc123`)  
**安全规则**: 仅 Header 传递，禁止 URL 参数

详见: [认证模块规范](./auth-module.md)

---

## 3. 响应格式

### 3.1 统一包装（强制）
```json
// 成功
{"success": true, "code": 200, "message": "OK", "data": {...}}

// 错误
{"success": false, "code": 400, "message": "参数错误", "error": {"code": "VAL_INVALID_PARAM_001", "message": "..."}}
```

实现: `pkg/response`

### 3.2 分页格式
```json
{
  "data": {
    "items": [...],
    "pagination": {"total": 100, "page": 1, "pageSize": 20, "totalPages": 5}
  }
}
```

---

## 4. 错误码

**格式**: `<MODULE>_<TYPE>_<NUMBER>`

示例:
- `AUTH_INVALID_TOKEN_001`
- `RES_NOT_FOUND_001`
- `VAL_MISSING_FIELD_002`
- `SYS_DATABASE_ERROR_003`

**HTTP 状态码映射**:
- 400 → `VAL_*` (参数错误)
- 401 → `AUTH_*` (未认证)
- 403 → `PERM_*` (无权限)
- 404 → `RES_NOT_FOUND_*`
- 429 → `RATE_LIMIT_*`
- 500 → `SYS_*` (系统错误)

定义位置: `internal/errors/codes.go`

---

## 5. 命名约定

| 层级 | 格式 | 示例 |
|------|------|------|
| JSON 输出 | `snake_case` | `user_id`, `created_at` |
| Go 代码 | `CamelCase` | `UserID`, `CreatedAt` |
| 枚举值 | `lowercase` | `active`, `pending` |
| 布尔字段 | `is_*` / `has_*` | `is_active` |
| URL | 复数名词 | `/users`, `/projects` |

---

## 6. 安全

### Rate Limiting
- 免费用户: 1000/小时
- 付费用户: 10000/小时
- 管理员: 无限制

### CORS
白名单管理 (`internal/middleware/cors.go`)，开发环境允许 `localhost:3000`

---

## 7. OpenAPI 文档

### Swagger 注解（强制）
所有公开 API 必须添加注解：

```go
// @Summary      Get configuration item
// @Tags         config
// @Param        key  query  string  true  "Configuration key"
// @Success      200  {object}  GetConfigResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /config [get]
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) { ... }
```

### 开发流程
1. 添加注解 → 2. `make swagger` → 3. 验证 `/api/docs/` → 4. 提交 `core/docs/`

详见: [Story 11](../sprint-artifacts/sprint-0/story-11-swagger-docs.md)

---

## 8. 查询参数

- 分页: `?page=1&pageSize=20`
- 过滤: `?status=active&owner_id=123`
- 排序: `?sort=created_at&order=desc`
- 搜索: `?q=keyword`

---

## 相关文档

- [认证模块规范](./auth-module.md)
- [错误处理](../../pkg/errors/README.md)
- [响应工具](../../pkg/response/README.md)
- [技术架构](../architecture/tech-architecture.md)

---

**维护者**: Architect Agent
