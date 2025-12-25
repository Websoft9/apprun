# API 设计规范
# apprun BaaS Platform

**创建日期**: 2025-12-25  
**维护者**: Winston (Architect Agent)  
**版本**: 1.0  
**状态**: Draft

---

## 1. RESTful API 规范

### 1.1 基本原则

- **资源导向**: URL 表示资源，HTTP 方法表示操作
- **无状态**: 每个请求包含完整的认证和上下文信息
- **统一接口**: 使用标准的 HTTP 方法和状态码
- **可缓存**: 合理使用 HTTP 缓存机制

### 1.2 URL 设计

#### 1.2.1 命名规范

```
✅ 推荐
GET  /api/v1/users              # 资源用复数名词
GET  /api/v1/projects/123       # 使用 ID 标识具体资源
POST /api/v1/projects           # 创建资源
GET  /api/v1/projects/123/members  # 嵌套资源

❌ 避免
GET  /api/v1/getUsers           # 不要在 URL 中包含动词
GET  /api/v1/user               # 不要使用单数
GET  /api/v1/projects-list      # 不要使用连字符表示操作
```

#### 1.2.2 版本控制

```
# URL 路径版本（推荐）
GET /api/v1/users
GET /api/v2/users

# 不推荐：HTTP Header 版本（增加复杂度）
GET /api/users
Header: X-API-Version: 1
```

#### 1.2.3 资源层级

```
# 浅层级（推荐）
GET /api/v1/projects/123/files

# 深层级（避免超过 3 层）
❌ GET /api/v1/projects/123/models/456/fields/789/validations
✅ GET /api/v1/model-fields/789/validations
```

---

## 2. HTTP 方法

### 2.1 标准方法

| 方法 | 用途 | 幂等性 | 安全性 | 示例 |
|------|------|--------|--------|------|
| **GET** | 查询资源 | ✅ | ✅ | `GET /api/v1/users/123` |
| **POST** | 创建资源 | ❌ | ❌ | `POST /api/v1/users` |
| **PUT** | 完整更新资源 | ✅ | ❌ | `PUT /api/v1/users/123` |
| **PATCH** | 部分更新资源 | ✅ | ❌ | `PATCH /api/v1/users/123` |
| **DELETE** | 删除资源 | ✅ | ❌ | `DELETE /api/v1/users/123` |

### 2.2 方法使用示例

```bash
# 列表查询
GET /api/v1/projects?page=1&pageSize=10&status=active

# 详情查询
GET /api/v1/projects/123

# 创建资源
POST /api/v1/projects
Content-Type: application/json
{
  "name": "My Project",
  "description": "Project description"
}

# 完整更新（必须包含所有字段）
PUT /api/v1/projects/123
Content-Type: application/json
{
  "name": "Updated Project",
  "description": "Updated description",
  "status": "active"
}

# 部分更新（只更新指定字段）
PATCH /api/v1/projects/123
Content-Type: application/json
{
  "name": "New Name"
}

# 删除资源
DELETE /api/v1/projects/123
```

---

## 3. 请求规范

### 3.1 请求头

```http
# 必需的请求头
Content-Type: application/json
Accept: application/json
Authorization: Bearer <token>

# 可选的请求头
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000  # 请求追踪 ID
X-Client-Version: 1.0.0                              # 客户端版本
Accept-Language: zh-CN,en                            # 语言偏好
```

### 3.2 请求体

#### 3.2.1 JSON 格式

```json
// 创建项目
POST /api/v1/projects
{
  "name": "My Project",              // 必需字段
  "description": "Description",      // 可选字段
  "settings": {                      // 嵌套对象
    "visibility": "private",
    "features": ["functions", "storage"]
  }
}

// 批量操作
POST /api/v1/projects/batch-delete
{
  "ids": ["123", "456", "789"]
}
```

#### 3.2.2 查询参数

```bash
# 分页
GET /api/v1/projects?page=1&pageSize=20

# 过滤
GET /api/v1/projects?status=active&owner_id=123

# 排序
GET /api/v1/projects?sort=created_at&order=desc

# 搜索
GET /api/v1/projects?q=keyword

# 字段选择（减少响应大小）
GET /api/v1/projects?fields=id,name,created_at

# 展开关联资源
GET /api/v1/projects?expand=owner,members
```

### 3.3 文件上传

```http
# 单文件上传
POST /api/v1/storage/upload
Content-Type: multipart/form-data

file: <binary>
project_id: 123
folder_path: /docs

# 多文件上传
POST /api/v1/storage/upload-batch
Content-Type: multipart/form-data

files[]: <binary>
files[]: <binary>
project_id: 123
```

---

## 4. 响应规范

### 4.1 统一响应格式

#### 4.1.1 成功响应

```json
// 单个资源
{
  "success": true,
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": "123",
    "name": "My Project",
    "created_at": "2025-12-25T10:00:00Z"
  }
}

// 列表响应
{
  "success": true,
  "code": 200,
  "message": "查询成功",
  "data": {
    "items": [
      {"id": "1", "name": "Project 1"},
      {"id": "2", "name": "Project 2"}
    ],
    "pagination": {
      "total": 100,
      "page": 1,
      "pageSize": 10,
      "totalPages": 10
    }
  }
}

// 创建成功（包含资源 URL）
{
  "success": true,
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": "123",
    "name": "New Project"
  },
  "location": "/api/v1/projects/123"
}

// 无数据响应
{
  "success": true,
  "code": 204,
  "message": "删除成功"
}
```

#### 4.1.2 错误响应

```json
// 客户端错误
{
  "success": false,
  "code": 400,
  "message": "请求参数错误",
  "error": {
    "code": "INVALID_PARAM",
    "message": "name 字段不能为空",
    "details": {
      "field": "name",
      "constraint": "required"
    }
  }
}

// 认证错误
{
  "success": false,
  "code": 401,
  "message": "未授权访问",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Token 已过期"
  }
}

// 权限错误
{
  "success": false,
  "code": 403,
  "message": "无权限访问",
  "error": {
    "code": "FORBIDDEN",
    "message": "您不是该项目的成员"
  }
}

// 资源不存在
{
  "success": false,
  "code": 404,
  "message": "资源不存在",
  "error": {
    "code": "NOT_FOUND",
    "message": "项目 ID 123 不存在"
  }
}

// 服务器错误
{
  "success": false,
  "code": 500,
  "message": "服务器内部错误",
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "数据库连接失败",
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### 4.2 HTTP 状态码

| 状态码 | 含义 | 使用场景 |
|--------|------|----------|
| **200** | OK | 成功获取资源或执行操作 |
| **201** | Created | 成功创建资源 |
| **204** | No Content | 成功删除资源或无返回数据 |
| **400** | Bad Request | 请求参数错误 |
| **401** | Unauthorized | 未认证或 Token 无效 |
| **403** | Forbidden | 已认证但无权限 |
| **404** | Not Found | 资源不存在 |
| **409** | Conflict | 资源冲突（如重复创建） |
| **422** | Unprocessable Entity | 请求格式正确但语义错误 |
| **429** | Too Many Requests | 请求频率限制 |
| **500** | Internal Server Error | 服务器内部错误 |
| **503** | Service Unavailable | 服务暂时不可用 |

### 4.3 响应头

```http
# 标准响应头
Content-Type: application/json; charset=utf-8
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640390400

# 创建资源时返回 Location
Location: /api/v1/projects/123

# 分页响应头（可选）
X-Total-Count: 100
X-Page: 1
X-Page-Size: 10
```

---

## 5. 错误码规范

### 5.1 错误码设计

```
格式: <MODULE>_<ERROR_TYPE>_<NUMBER>

示例:
- AUTH_INVALID_TOKEN_001
- PROJECT_NOT_FOUND_001
- STORAGE_QUOTA_EXCEEDED_001
```

### 5.2 常用错误码

```go
// internal/errors/codes.go

const (
    // 认证错误 (AUTH_*)
    ErrAuthInvalidToken     = "AUTH_INVALID_TOKEN_001"
    ErrAuthTokenExpired     = "AUTH_TOKEN_EXPIRED_002"
    ErrAuthUnauthorized     = "AUTH_UNAUTHORIZED_003"
    
    // 权限错误 (PERM_*)
    ErrPermForbidden        = "PERM_FORBIDDEN_001"
    ErrPermInsufficientRole = "PERM_INSUFFICIENT_ROLE_002"
    
    // 资源错误 (RES_*)
    ErrResNotFound          = "RES_NOT_FOUND_001"
    ErrResAlreadyExists     = "RES_ALREADY_EXISTS_002"
    ErrResConflict          = "RES_CONFLICT_003"
    
    // 验证错误 (VAL_*)
    ErrValInvalidParam      = "VAL_INVALID_PARAM_001"
    ErrValMissingField      = "VAL_MISSING_FIELD_002"
    ErrValFormatError       = "VAL_FORMAT_ERROR_003"
    
    // 业务错误 (BIZ_*)
    ErrBizQuotaExceeded     = "BIZ_QUOTA_EXCEEDED_001"
    ErrBizOperationFailed   = "BIZ_OPERATION_FAILED_002"
    
    // 系统错误 (SYS_*)
    ErrSysInternalError     = "SYS_INTERNAL_ERROR_001"
    ErrSysServiceUnavailable = "SYS_SERVICE_UNAVAILABLE_002"
    ErrSysDatabaseError     = "SYS_DATABASE_ERROR_003"
)
```

---

## 6. 认证与授权

### 6.1 认证方式

```http
# Bearer Token（推荐）
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# API Key（用于服务端调用）
X-API-Key: sk_live_1234567890abcdef
```

### 6.2 认证流程

```
1. 用户登录 → Kratos
2. Kratos 生成 Session Token
3. 客户端在请求中携带 Token
4. apprun 验证 Token → 查询用户信息
5. 检查用户权限 → 执行业务逻辑
```

### 6.3 权限控制

```bash
# 项目级权限检查
GET /api/v1/projects/123/files
# 需要验证：用户是否是项目成员 && 是否有读取权限

# 资源级权限检查
DELETE /api/v1/files/456
# 需要验证：用户是否有删除该文件的权限
```

---

## 7. 分页与过滤

### 7.1 分页

```bash
# Offset 分页（推荐）
GET /api/v1/projects?page=1&pageSize=20

# 响应
{
  "data": {
    "items": [...],
    "pagination": {
      "total": 100,        # 总记录数
      "page": 1,           # 当前页码
      "pageSize": 20,      # 每页大小
      "totalPages": 5      # 总页数
    }
  }
}

# Cursor 分页（大数据集）
GET /api/v1/projects?limit=20&cursor=eyJpZCI6MTIzfQ

# 响应
{
  "data": {
    "items": [...],
    "pagination": {
      "nextCursor": "eyJpZCI6MTQzfQ",
      "hasMore": true
    }
  }
}
```

### 7.2 过滤

```bash
# 单条件过滤
GET /api/v1/projects?status=active

# 多条件过滤（AND）
GET /api/v1/projects?status=active&owner_id=123

# 范围过滤
GET /api/v1/projects?created_after=2025-01-01&created_before=2025-12-31

# IN 查询
GET /api/v1/projects?status=active,archived

# 模糊搜索
GET /api/v1/projects?q=keyword
```

### 7.3 排序

```bash
# 单字段排序
GET /api/v1/projects?sort=created_at&order=desc

# 多字段排序
GET /api/v1/projects?sort=status,created_at&order=asc,desc
```

---

## 8. 批量操作

### 8.1 批量创建

```json
POST /api/v1/projects/batch
{
  "items": [
    {"name": "Project 1"},
    {"name": "Project 2"},
    {"name": "Project 3"}
  ]
}

// 响应
{
  "success": true,
  "data": {
    "created": [
      {"id": "1", "name": "Project 1"},
      {"id": "2", "name": "Project 2"}
    ],
    "failed": [
      {
        "index": 2,
        "error": "name already exists"
      }
    ]
  }
}
```

### 8.2 批量更新

```json
PATCH /api/v1/projects/batch
{
  "ids": ["1", "2", "3"],
  "updates": {
    "status": "archived"
  }
}

// 响应
{
  "success": true,
  "data": {
    "updated": 3,
    "failed": 0
  }
}
```

### 8.3 批量删除

```json
DELETE /api/v1/projects/batch
{
  "ids": ["1", "2", "3"]
}

// 响应
{
  "success": true,
  "data": {
    "deleted": 3
  }
}
```

---

## 9. 缓存策略

### 9.1 HTTP 缓存头

```http
# 强缓存（不变资源）
Cache-Control: public, max-age=31536000, immutable
ETag: "33a64df551425fcc55e4d42a148795d9f25f89d4"

# 协商缓存（可变资源）
Cache-Control: no-cache
ETag: "686897696a7c876b7e"
Last-Modified: Wed, 25 Dec 2025 10:00:00 GMT

# 不缓存
Cache-Control: no-store, no-cache, must-revalidate
```

### 9.2 条件请求

```http
# If-None-Match（基于 ETag）
GET /api/v1/projects/123
If-None-Match: "686897696a7c876b7e"

# 304 Not Modified（资源未变化）
HTTP/1.1 304 Not Modified
ETag: "686897696a7c876b7e"

# If-Modified-Since（基于时间）
GET /api/v1/projects/123
If-Modified-Since: Wed, 25 Dec 2025 10:00:00 GMT
```

---

## 10. 异步操作

### 10.1 长时间任务

```json
// 提交异步任务
POST /api/v1/projects/123/export
{
  "format": "json",
  "include": ["models", "files"]
}

// 响应（返回任务 ID）
{
  "success": true,
  "code": 202,
  "message": "任务已提交",
  "data": {
    "task_id": "task_123",
    "status": "pending",
    "status_url": "/api/v1/tasks/task_123"
  }
}

// 查询任务状态
GET /api/v1/tasks/task_123

// 响应
{
  "success": true,
  "data": {
    "task_id": "task_123",
    "status": "completed",
    "progress": 100,
    "result": {
      "download_url": "/api/v1/downloads/export_123.json"
    }
  }
}
```

---

## 11. WebSocket API

### 11.1 连接建立

```javascript
// 客户端连接
const ws = new WebSocket('ws://localhost:8080/ws?token=<token>');

// 订阅事件
ws.send(JSON.stringify({
  action: 'subscribe',
  events: ['user.created', 'project.updated']
}));

// 接收消息
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(data.type, data.payload);
};
```

### 11.2 消息格式

```json
// 服务端推送
{
  "type": "user.created",
  "payload": {
    "user_id": "123",
    "email": "user@example.com"
  },
  "timestamp": "2025-12-25T10:00:00Z"
}
```

---

## 12. API 文档

### 12.1 OpenAPI/Swagger

```yaml
# openapi.yaml
openapi: 3.0.0
info:
  title: apprun BaaS Platform API
  version: 1.0.0
  description: RESTful API for apprun platform

paths:
  /api/v1/projects:
    get:
      summary: List projects
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: pageSize
          in: query
          schema:
            type: integer
            default: 20
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ProjectListResponse'
```

### 12.2 示例代码生成

```bash
# 从 OpenAPI 生成客户端 SDK
openapi-generator generate \
  -i openapi.yaml \
  -g go \
  -o ./sdk/go

# 生成 Postman Collection
openapi2postmanv2 -s openapi.yaml -o postman_collection.json
```

---

## 13. 最佳实践

### 13.1 命名约定

- **字段名**: 使用 snake_case（如 `user_id`, `created_at`）
- **枚举值**: 使用 lowercase（如 `active`, `pending`）
- **布尔字段**: 使用 `is_*` 或 `has_*` 前缀（如 `is_active`, `has_permission`）

### 13.2 API 设计检查清单

- [ ] URL 使用复数名词
- [ ] HTTP 方法正确使用
- [ ] 统一响应格式
- [ ] 完整的错误处理
- [ ] 分页参数一致
- [ ] 认证和授权检查
- [ ] 幂等性保证（PUT/DELETE）
- [ ] OpenAPI 文档完整
- [ ] 示例请求和响应
- [ ] 版本控制策略

### 13.3 性能优化

- 使用 `fields` 参数减少响应大小
- 使用 `expand` 参数减少请求次数
- 合理使用 HTTP 缓存
- 避免深层嵌套资源
- 批量操作代替多次单一操作

---

## 附录

### A. API 路由清单

详见：[tech-architecture.md](../architecture/tech-architecture.md#41-路由规则)

### B. 工具推荐

- **API 测试**: Postman, Insomnia, HTTPie
- **文档生成**: Swagger UI, ReDoc
- **Mock 服务**: Prism, JSON Server
- **性能测试**: k6, Apache JMeter

---

**文档维护**: Winston (Architect Agent)  
**审核状态**: 待开发团队评审  
**下一步**: 编码规范文档 (coding-standards.md)
