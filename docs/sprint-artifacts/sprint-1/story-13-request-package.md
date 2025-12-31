# Story 13: Request Package
# Sprint 1: Infrastructure Enhancement

**Priority**: P1  
**Effort**: 2-3 天  
**Owner**: Backend Dev  
**Dependencies**: Story 2 (Response Package), Story 12 (Logger Package)  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [API 设计规范](../../standards/api-design.md)

---

## User Story

作为开发者，我希望有统一的请求处理工具包，以便在所有 API Handler 中快速解析、验证请求数据，减少重复代码并确保输入安全性。

---

## Acceptance Criteria

- [ ] 实现 `core/pkg/request` 包的核心功能
- [ ] 提供泛型函数 `ParseAndValidate[T](r *http.Request) (T, error)`
- [ ] 集成 go-playground/validator 验证框架
- [ ] 自动返回 `response.ValidationError` 格式错误
- [ ] 支持请求体大小限制（默认 1MB，可配置）
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 编写 `README.md` 使用文档

---

## Implementation Tasks

### Phase 1: 核心实现（Day 1）
- [ ] 创建 `core/pkg/request/request.go` - 核心解析函数
- [ ] 创建 `core/pkg/request/validator.go` - 验证逻辑
- [ ] 实现 `ParseAndValidate[T]()` 泛型函数
- [ ] 实现 `Parse[T]()` 仅解析函数（不验证）
- [ ] 实现 `Validate[T]()` 仅验证函数（已有结构体）
- [ ] 添加请求体大小限制中间件

### Phase 2: 测试与文档（Day 2）
- [ ] 编写单元测试（覆盖正常、异常、边界场景）
- [ ] 编写 `README.md` 文档（含使用示例）
- [ ] 创建 demo handler 演示集成
- [ ] golangci-lint 检查通过

### Phase 3: 集成验证（Day 3 可选）
- [ ] 重构 config handler 使用新 request 包
- [ ] 验证与 response 包的协同工作
- [ ] 性能测试（确保无明显开销）

---

## Technical Design

### 包结构
```
core/pkg/request/
├── request.go       # 核心解析函数
├── validator.go     # 验证配置与辅助函数
├── request_test.go  # 单元测试
└── README.md        # 使用文档
```

### 核心 API

```go
package request

// ParseAndValidate 解析请求体并验证（推荐）
func ParseAndValidate[T any](r *http.Request) (T, error)

// Parse 仅解析请求体（不验证）
func Parse[T any](r *http.Request) (T, error)

// Validate 仅验证结构体（已解析）
func Validate[T any](data T) error

// WithMaxBodySize 配置请求体最大大小（默认 1MB）
func WithMaxBodySize(size int64)
```

### 使用示例

```go
// Before (手动处理)
func UpdateConfig(w http.ResponseWriter, r *http.Request) {
    var req UpdateConfigRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.ValidationErrorWithRequest(w, r, "Invalid JSON format", err)
        return
    }
    if req.Key == "" {
        response.ValidationErrorWithRequest(w, r, "key", "missing 'key' field")
        return
    }
    // ... business logic
}

// After (使用 request 包)
func UpdateConfig(w http.ResponseWriter, r *http.Request) {
    req, err := request.ParseAndValidate[UpdateConfigRequest](r)
    if err != nil {
        response.ValidationErrorWithRequest(w, r, "validation failed", err)
        return
    }
    // ... business logic
}
```

### 验证标签示例

```go
type UpdateConfigRequest struct {
    Key   string `json:"key" validate:"required,min=1,max=100"`
    Value string `json:"value" validate:"required"`
}

type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=32,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```

---

## Dependencies

### Go 模块依赖
```bash
go get github.com/go-playground/validator/v10
```

### 已有依赖（无需安装）
- `apprun/pkg/response` - 错误响应格式
- `apprun/pkg/logger` - 日志记录
- `encoding/json` - JSON 解析

---

## Testing Strategy

### 单元测试覆盖
- ✅ 正常 JSON 解析与验证
- ✅ 无效 JSON 格式错误处理
- ✅ 验证失败场景（缺少字段、格式错误）
- ✅ 请求体过大限制
- ✅ 空请求体处理
- ✅ Content-Type 校验

### 集成测试
- ✅ 与 response 包协同工作
- ✅ 与 config handler 集成测试

---

## Definition of Done

- [ ] 所有代码通过 golangci-lint 检查
- [ ] 单元测试覆盖率 ≥ 80%，全部通过
- [ ] README 文档完整，含使用示例
- [ ] Demo handler 演示集成
- [ ] Code Review 通过
- [ ] 与 response/logger 包无冲突

---

## Technical Notes

### 与 Response 包的关系
```
HTTP Request
    ↓
[pkg/request] ← 解析 + 验证
    ↓
[Business Handler]
    ↓
[pkg/response] ← 统一响应
    ↓
HTTP Response
```

### 设计原则
- **最小化**: 仅提供核心解析与验证，避免过度抽象
- **泛型优先**: 使用 Go 1.18+ 泛型避免反射开销
- **可选使用**: 不强制使用，保持与 Chi 框架的兼容性
- **与 Response 对称**: 保持 API 风格一致

### 安全考虑
- 默认限制请求体大小为 1MB（防止 DoS）
- 验证所有输入字段（防止注入攻击）
- 敏感字段过滤（后续与 logger 集成）

---

## References

- [API 设计规范](../../standards/api-design.md)
- [Story 2: Response Package](../sprint-0/story-02-response-package.md)
- [Story 12: Logger Package](../sprint-0/story-12-logger-package.md)
- [go-playground/validator](https://github.com/go-playground/validator)
