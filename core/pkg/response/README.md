# Response Package

统一的 HTTP 响应工具包，用于生成标准化的 JSON API 响应格式。

## Features

- ✅ 统一响应格式（符合 API 设计规范）
- ✅ 标准 HTTP 状态码支持
- ✅ 标准化错误码系统
- ✅ 分页支持
- ✅ 100% 测试覆盖率

## Installation

```go
import "apprun/pkg/response"
```

## Usage

### Success Response (HTTP 200)

```go
func GetUser(w http.ResponseWriter, r *http.Request) {
    user := map[string]interface{}{
        "id":   "123",
        "name": "John Doe",
    }
    response.Success(w, user)
}
```

**Output:**
```json
{
  "success": true,
  "code": 200,
  "data": {
    "id": "123",
    "name": "John Doe"
  }
}
```

### Created Response (HTTP 201)

```go
func CreateProject(w http.ResponseWriter, r *http.Request) {
    project := map[string]string{"id": "456", "name": "New Project"}
    location := "/api/v1/projects/456"
    response.Created(w, project, location)
}
```

**Output:**
```json
{
  "success": true,
  "code": 201,
  "data": {
    "id": "456",
    "name": "New Project"
  }
}
```
**Headers:** `Location: /api/v1/projects/456`

### No Content Response (HTTP 204)

```go
func DeleteUser(w http.ResponseWriter, r *http.Request) {
    // ... delete logic ...
    response.NoContent(w)
}
```

**Output:** Empty body with HTTP 204 status.

### List Response with Pagination

```go
func ListProjects(w http.ResponseWriter, r *http.Request) {
    projects := []map[string]string{
        {"id": "1", "name": "Project 1"},
        {"id": "2", "name": "Project 2"},
    }
    
    pagination := &response.PaginationInfo{
        Total:      100,
        Page:       1,
        PageSize:   10,
        TotalPages: 10,
    }
    
    response.List(w, projects, pagination)
}
```

**Output:**
```json
{
  "success": true,
  "code": 200,
  "data": {
    "items": [
      {"id": "1", "name": "Project 1"},
      {"id": "2", "name": "Project 2"}
    ],
    "pagination": {
      "total": 100,
      "page": 1,
      "page_size": 10,
      "total_pages": 10
    }
  }
}
```

### Error Response

```go
func GetUser(w http.ResponseWriter, r *http.Request) {
    // User not found
    response.Error(w, 404, response.ErrCodeNotFound, "User not found")
}
```

**Output:**
```json
{
  "success": false,
  "code": 404,
  "error": {
    "code": "RES_NOT_FOUND_001",
    "message": "User not found"
  }
}
```

### Validation Error (HTTP 422)

```go
func CreateUser(w http.ResponseWriter, r *http.Request) {
    // Validation failed
    response.ValidationError(w, "email", "Email format is invalid")
}
```

**Output:**
```json
{
  "success": false,
  "code": 422,
  "error": {
    "code": "VAL_INVALID_PARAM_001",
    "message": "Email format is invalid",
    "details": {
      "field": "email"
    }
  }
}
```

## Error Codes

Standard error codes following the format: `<MODULE>_<ERROR_TYPE>_<NUMBER>`

### Validation Errors (VAL_*)
- `ErrCodeInvalidParam` - `VAL_INVALID_PARAM_001`
- `ErrCodeMissingField` - `VAL_MISSING_FIELD_002`
- `ErrCodeFormatError` - `VAL_FORMAT_ERROR_003`

### Resource Errors (RES_*)
- `ErrCodeNotFound` - `RES_NOT_FOUND_001`
- `ErrCodeAlreadyExists` - `RES_ALREADY_EXISTS_002`
- `ErrCodeConflict` - `RES_CONFLICT_003`

### Authentication Errors (AUTH_*)
- `ErrCodeInvalidToken` - `AUTH_INVALID_TOKEN_001`
- `ErrCodeTokenExpired` - `AUTH_TOKEN_EXPIRED_002`
- `ErrCodeUnauthorized` - `AUTH_UNAUTHORIZED_003`

### Permission Errors (PERM_*)
- `ErrCodeForbidden` - `PERM_FORBIDDEN_001`
- `ErrCodeInsufficientRole` - `PERM_INSUFFICIENT_ROLE_002`

### Business Errors (BIZ_*)
- `ErrCodeQuotaExceeded` - `BIZ_QUOTA_EXCEEDED_001`
- `ErrCodeOperationFailed` - `BIZ_OPERATION_FAILED_002`

### System Errors (SYS_*)
- `ErrCodeInternalError` - `SYS_INTERNAL_ERROR_001`
- `ErrCodeServiceUnavailable` - `SYS_SERVICE_UNAVAILABLE_002`
- `ErrCodeDatabaseError` - `SYS_DATABASE_ERROR_003`

## API Reference

### Functions

#### `Success(w http.ResponseWriter, data interface{})`
Sends a successful response with HTTP 200 status.

#### `Created(w http.ResponseWriter, data interface{}, location string)`
Sends a created response with HTTP 201 status and optional Location header.

#### `NoContent(w http.ResponseWriter)`
Sends an empty response with HTTP 204 status.

#### `List(w http.ResponseWriter, items interface{}, pagination *PaginationInfo)`
Sends a list response with optional pagination info.

#### `Error(w http.ResponseWriter, code int, errCode, message string)`
Sends an error response with specified HTTP status code and error details.

#### `ValidationError(w http.ResponseWriter, field, message string)`
Sends a validation error response (HTTP 422) with field details.

### Types

#### `Response`
```go
type Response struct {
    Success bool        `json:"success"`
    Code    int         `json:"code"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
}
```

#### `ErrorInfo`
```go
type ErrorInfo struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}
```

#### `PaginationInfo`
```go
type PaginationInfo struct {
    Total      int `json:"total"`
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    TotalPages int `json:"total_pages"`
}
```

## Testing

Run tests with coverage:
```bash
go test -v -cover
```

Current coverage: **100%**

## Standards Compliance

This package follows:
- [API Design Standards § 4.1](../../docs/standards/api-design.md#41-统一响应格式)
- [API Design Standards § 5.2](../../docs/standards/api-design.md#52-常用错误码)
- [Go Coding Standards](../../docs/standards/coding-standards.md)

## License

Part of apprun BaaS Platform
