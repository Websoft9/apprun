# Story 3: 错误处理框架
# Sprint 0: Infrastructure建设

**Priority**: P0  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: Story 2  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [API 设计规范](../../standards/api-design.md#42-错误处理)

---

## User Story

作为开发者，我希望有统一的错误处理框架，以便标准化系统错误定义和处理流程。

---

## Acceptance Criteria

- [ ] 创建 `core/pkg/errors` 包
- [ ] 定义错误码常量（业务错误、系统错误）
- [ ] 实现 `AppError` 类型
- [ ] 实现错误码到 HTTP 状态码的映射
- [ ] 编写单元测试（覆盖率 > 90%）
- [ ] 编写使用文档和示例

---

## Implementation Tasks

- [ ] 创建 `core/pkg/errors/errors.go`
- [ ] 定义错误码常量（E1000-E4999）
- [ ] 定义 AppError 结构体
- [ ] 实现 New 函数
- [ ] 实现 Wrap 函数
- [ ] 实现 ToHTTPStatus 函数
- [ ] 编写单元测试
- [ ] 编写 README.md（使用示例）

---

## Technical Details

### 错误码规范

- `E1xxx`: 认证/授权错误
- `E2xxx`: 请求参数错误
- `E3xxx`: 业务逻辑错误
- `E4xxx`: 系统/服务错误

### 代码示例

```go
// core/pkg/errors/errors.go

package errors

import "errors"

// 错误码常量
const (
    ErrUnauthorized      = "E1001"
    ErrForbidden         = "E1003"
    ErrInvalidRequest    = "E2001"
    ErrResourceNotFound  = "E3001"
    ErrInternalError     = "E4001"
)

type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    return e.Message
}

func New(code, message string) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
    }
}

func Wrap(err error, code, message string) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}

func ToHTTPStatus(code string) int {
    // E1xxx -> 401/403
    // E2xxx -> 400
    // E3xxx -> 404/422
    // E4xxx -> 500
}
```

---

## Test Cases

- [ ] 创建 AppError 正确
- [ ] Wrap 保留原始错误
- [ ] 错误码正确映射到 HTTP 状态码
- [ ] Error() 方法返回正确消息

---

## Related Docs

- [API 设计规范](../../standards/api-design.md)
- [错误处理最佳实践](../../standards/coding-standards.md#error-handling)

---

**Created**: 2025-12-27  
**Updated**: 2025-12-27  
**Maintainer**: Architect Agent
