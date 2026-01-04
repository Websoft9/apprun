# Story 3: 业务错误包装框架
# Sprint 0: Infrastructure建设

**Priority**: P0  
**Effort**: 1 天  
**Owner**: Backend Dev  
**Dependencies**: 无  
**Status**: Done  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [API 设计规范](../../standards/api-design.md#42-错误处理)

**Completed**: 2026-01-04  
**Coverage**: 97.6% (errors), 100% (httpmap)  
**Files**: `pkg/errors/errors.go`, `pkg/errors/codes.go`, `pkg/errors/httpmap/httpmap.go`

---

## User Story

作为开发者，我希望有统一的错误包装工具，以便在业务模块中标准化错误定义、传播和处理，支持错误链和上下文注入。

---

## Architecture Decision

### 设计原则
1. **高内聚低耦合**：errors 包自包含，定义自己的错误码，不依赖其他业务包
2. **stdlib 优先**：基于 Go 1.13+ stdlib `errors`，支持 `Is`/`As`/`Unwrap`
3. **简单实用**：提供核心功能，避免过度设计

### 职责范围
- 定义业务错误码常量
- 提供错误包装和上下文管理
- 映射错误码到 HTTP 状态码（供 HTTP 层使用）

---

## Acceptance Criteria

- [ ] 创建 `core/pkg/errors` 包
- [ ] 定义错误码常量（模块前缀 + 类型 + 编号）
- [ ] 实现 `AppError` 结构体（Code、Message、Err、Context）
- [ ] 实现 `New`、`Newf`、`Wrap`、`Wrapf`、`WithContext` 方法
- [ ] 实现 `Category()` 方法和 `IsValidation`、`IsNotFound` 等 helper 函数
- [ ] 创建 `core/pkg/errors/httpmap` 子包，实现 `ToHTTPStatus` 映射函数
- [ ] 编写单元测试（覆盖率 > 90%）
- [ ] 编写 README.md（含错误码注册规则）

---

## Implementation Tasks

- [x] 创建 `core/pkg/errors/errors.go`
- [x] 定义错误码常量（含前缀定义和命名规则）
- [x] 实现 `AppError` 结构体（Error、Unwrap、WithContext、Category）
- [x] 实现工厂函数（New、Newf、Wrap、Wrapf）
- [x] 实现 helper 函数（IsValidation、IsNotFound、IsAuth、IsSystem）
- [x] 创建 `core/pkg/errors/httpmap/httpmap.go`（ToHTTPStatus 适配器）
- [x] 编写 `errors_test.go` 和 `httpmap_test.go`
- [x] 编写 `README.md`（含错误码注册流程）

---

## Technical Details

### 错误码设计

**格式**：`<MODULE>_<CATEGORY>_<NAME>_<NNN>`

```go
// core/pkg/errors/errors.go
package errors

// 类别前缀（预定义）
const (
    CategoryValidation = "VAL"   // 验证错误
    CategoryResource   = "RES"   // 资源错误
    CategoryAuth       = "AUTH"  // 认证错误
    CategoryPermission = "PERM"  // 权限错误
    CategoryBusiness   = "BIZ"   // 业务错误
    CategorySystem     = "SYS"   // 系统错误
)

// Context key 常量
const (
    ContextKeyUserID    = "user_id"
    ContextKeyProjectID = "project_id"
    ContextKeyRequestID = "request_id"
)

// 错误码常量（示例）
const (
    ErrCodeNotFound      = "CORE_RES_NOT_FOUND_001"
    ErrCodeInvalidParam  = "CORE_VAL_INVALID_PARAM_001"
    ErrCodeUnauthorized  = "CORE_AUTH_UNAUTHORIZED_001"
    ErrCodeInternalError = "CORE_SYS_INTERNAL_ERROR_001"
)

// AppError 业务错误
type AppError struct {
    Code    string
    Message string
    Err     error
    Context map[string]interface{}
}

func (e *AppError) Error() string
func (e *AppError) Unwrap() error
func (e *AppError) WithContext(key string, value interface{}) *AppError
func (e *AppError) Category() string  // 返回类别（如 "VAL"）

func New(code, message string) *AppError
func Newf(code, format string, args ...interface{}) *AppError
func Wrap(err error, code, message string) *AppError
func Wrapf(err error, code, format string, args ...interface{}) *AppError

// Helper 函数
func IsValidation(err error) bool
func IsNotFound(err error) bool
func IsAuth(err error) bool
func IsSystem(err error) bool
```

### HTTP 映射适配器

```go
// core/pkg/errors/httpmap/httpmap.go
package httpmap

import "apprun/pkg/errors"

// ToHTTPStatus 映射错误码到 HTTP 状态码（粗粒度）
func ToHTTPStatus(err error) int {
    var appErr *errors.AppError
    if !errors.As(err, &appErr) {
        return 500
    }
    
    switch appErr.Category() {
    case errors.CategoryValidation:
        return 400
    case errors.CategoryAuth:
        return 401
    case errors.CategoryPermission:
        return 403
    case errors.CategoryResource:
        if strings.Contains(appErr.Code, "NOT_FOUND") {
            return 404
        }
        return 409
    case errors.CategorySystem:
        return 500
    default:
        return 500
    }
}
```

### 使用示例

**业务层**：
```go
// 创建错误
err := errors.New(errors.ErrCodeNotFound, "Project not found")

// 格式化消息
err := errors.Newf(errors.ErrCodeInvalidParam, "Invalid field: %s", field)

// 包装错误
err := errors.Wrap(dbErr, errors.ErrCodeInternalError, "DB query failed").
    WithContext(errors.ContextKeyUserID, userID)

// 判断错误类型
if errors.IsNotFound(err) {
    // 处理 not found
}
```

**Handler 层**：
```go
import "apprun/pkg/errors/httpmap"

if err != nil {
    status := httpmap.ToHTTPStatus(err)
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        response.Error(w, status, appErr.Code, appErr.Message)
    }
}
```

---

## Test Cases

- [ ] New/Newf/Wrap/Wrapf 创建 AppError 正确
- [ ] Unwrap 保留错误链，支持 errors.Is/As
- [ ] WithContext 链式调用有效
- [ ] Category() 返回正确类别
- [ ] IsValidation/IsNotFound 等 helper 函数正确
- [ ] httpmap.ToHTTPStatus 映射正确
- [ ] Context key 常量化使用

---

## Error Code Naming Rules

1. **格式**：`<MODULE>_<CATEGORY>_<NAME>_<NNN>`
   - MODULE: 模块前缀（如 CORE、CONFIG、AUTH）
   - CATEGORY: 类别前缀（VAL、RES、AUTH、PERM、BIZ、SYS）
   - NAME: 语义名称（如 NOT_FOUND、INVALID_PARAM）
   - NNN: 三位数编号（001-999）

2. **注册流程**：
   - 新增错误码需添加到 `errors.go` 常量组
   - 命名避免重复（通过 PR review 校验）
   - 编号递增，同类错误相邻

3. **Context 使用规则**：
   - 使用预定义常量（ContextKey*）
   - 仅存储可序列化值（string、int、bool）
   - 敏感信息需脱敏

---

## Implementation Notes

1. **Category() 实现**：解析错误码字符串（按 `_` 分割取第二个字段），在 `New/Newf` 时校验格式合法性
2. **IsNotFound() 实现**：判断 `Category() == "RES"` 且 `strings.Contains(Code, "NOT_FOUND")`
3. **httpmap 兜底**：未知 Category 或解析失败时默认返回 `500`

---

## Migration History

### 2026-01-04: Error Handling Migration
**Developer**: Amelia (Dev Agent)

#### Migrated Modules

1. **Logger Package** (`pkg/logger/zap_adapter.go`)
   - Replaced `fmt.Errorf` with `errors.New()` and `errors.Wrap()`
   - Updated error codes: `ErrCodeLogInvalidConfig`, `ErrCodeLogFileOpenFailed`
   - **Tests Updated**: Fixed 3 failing tests in `zap_adapter_test.go` to match new error message format
     - `TestNewZapLogger_InvalidTarget`: Now expects "Invalid logger config"
     - `TestNewZapLogger_DuplicateTarget`: Now expects "Invalid logger config"
     - `TestNewZapLogger_PartialMultiTargetFailure`: Now expects "Failed to parse output targets"

2. **Config Module** (`modules/config/service.go`)
   - Migrated `UpdateConfig()` to use AppError with proper error codes
   - Migrated `DeleteDynamicConfig()` to use AppError
   - **New Error Code**: Added `ErrCodeConfigNotAllowedDB = "CONFIG_BIZ_NOT_ALLOWED_DB_004"`
   - **Tests Updated**: Fixed handler and service tests to expect correct HTTP status codes
     - Validation errors (VAL): 400 Bad Request
     - Business logic errors (BIZ): 422 Unprocessable Entity
   - **Files Changed**:
     - `modules/config/service.go`: Error handling migration
     - `modules/config/handler_test.go`: Test assertions updated
     - `modules/config/service_test.go`: Error message assertions updated
     - `pkg/errors/codes.go`: New error code added

#### HTTP Status Code Mapping
- **VAL** (Validation) → 400 Bad Request
- **RES** (Resource) → 404 Not Found (if contains "NOT_FOUND"), else 400
- **BIZ** (Business) → 422 Unprocessable Entity
- **SYS** (System) → 500 Internal Server Error

#### Test Results
- All logger tests passing (22/22)
- All config module tests passing (38/38)
- Error message assertions updated to match new error wrapping format

---

## Related Docs

- [API 设计规范](../../standards/api-design.md)
- [Go errors 最佳实践](https://go.dev/blog/go1.13-errors)

---

**Created**: 2025-12-27  
**Updated**: 2026-01-04  
**Maintainer**: Winston (Architect Agent)
