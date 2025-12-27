# Story 2: 统一响应工具包
# Sprint 0: Infrastructure建设

**Priority**: P0  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: Story 1  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [API 设计规范](../../standards/api-design.md#41-统一响应格式)

---

## User Story

作为开发者，我希望有统一的响应工具包，以便快速实现标准化的 API 响应格式。

---

## Acceptance Criteria

- [ ] 创建 `core/pkg/response` 包
- [ ] 实现 `Success()` 函数（成功响应）
- [ ] 实现 `Error()` 函数（错误响应）
- [ ] 实现 `List()` 函数（列表响应含分页）
- [ ] 编写单元测试（覆盖率 > 90%）
- [ ] 编写使用文档和示例

---

## Implementation Tasks

- [ ] 创建 `core/pkg/response/response.go`
- [ ] 定义响应结构体（Response、ErrorInfo、PaginationInfo）
- [ ] 实现 Success 函数
- [ ] 实现 Error 函数
- [ ] 实现 List 函数（含分页）
- [ ] 编写单元测试
- [ ] 编写 README.md（使用示例）
- [ ] 更新现有 Handler（config.go）使用新工具包

---

## Technical Details

```go
// core/pkg/response/response.go

package response

import (
    "encoding/json"
    "net/http"
)

type Response struct {
    Success bool        `json:"success"`
    Code    int         `json:"code"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type PaginationInfo struct {
    Total      int `json:"total"`
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    TotalPages int `json:"total_pages"`
}

func Success(w http.ResponseWriter, data interface{}) {
    // 实现
}

func Error(w http.ResponseWriter, code int, errCode, message string) {
    // 实现
}

func List(w http.ResponseWriter, items interface{}, pagination *PaginationInfo) {
    // 实现
}
```

---

## Test Cases

- [ ] 成功响应格式正确
- [ ] 错误响应包含完整错误信息
- [ ] 列表响应包含分页信息
- [ ] JSON 序列化正确

---

## Related Docs

- [API 设计规范](../../standards/api-design.md)
- [编码规范](../../standards/coding-standards.md)

---

**Created**: 2025-12-27  
**Updated**: 2025-12-27  
**Maintainer**: Architect Agent
