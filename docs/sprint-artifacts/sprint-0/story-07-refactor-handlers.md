# Story 7: 重构现有 Handlers
# Sprint 0: Infrastructure建设

**Priority**: P1  
**Effort**: 1 天  
**Owner**: Backend Dev  
**Dependencies**: Story 2, Story 3  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [API 设计规范](../../standards/api-design.md)

---

## User Story

作为开发者，我希望重构现有的 Handler 代码，以便使用统一的响应工具包和错误处理框架。

---

## Acceptance Criteria

- [ ] 重构 `handlers/config.go`
- [ ] 使用统一响应工具包
- [ ] 使用错误处理框架
- [ ] 移除重复代码
- [ ] 添加请求参数验证
- [ ] 更新单元测试
- [ ] 确保向后兼容

---

## Implementation Tasks

- [ ] 分析现有 config.go 代码
- [ ] 重构 ListConfigs（使用 response.List）
- [ ] 重构 GetConfig（使用 response.Success/Error）
- [ ] 重构 CreateConfig（使用 errors 包）
- [ ] 重构 UpdateConfig
- [ ] 重构 DeleteConfig
- [ ] 添加请求参数验证
- [ ] 更新单元测试

---

## Technical Details

### 重构前（示例）

```go
// handlers/config.go (旧代码)

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    config, err := h.repo.GetConfig(r.Context(), id)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(config)
}
```

### 重构后（示例）

```go
// handlers/config.go (新代码)

import (
    "github.com/yourusername/apprun/core/pkg/response"
    "github.com/yourusername/apprun/core/pkg/errors"
)

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    // 参数验证
    if id == "" {
        response.Error(w, http.StatusBadRequest, errors.ErrInvalidRequest, "config ID is required")
        return
    }
    
    config, err := h.repo.GetConfig(r.Context(), id)
    if err != nil {
        if errors.IsNotFound(err) {
            response.Error(w, http.StatusNotFound, errors.ErrResourceNotFound, "config not found")
        } else {
            response.Error(w, http.StatusInternalServerError, errors.ErrInternalError, "failed to get config")
        }
        return
    }
    
    response.Success(w, config)
}
```

---

## 重构清单

- [ ] GetConfig
- [ ] ListConfigs（添加分页）
- [ ] CreateConfig（添加验证）
- [ ] UpdateConfig（添加验证）
- [ ] DeleteConfig
- [ ] 移除重复的 JSON 编码代码
- [ ] 统一错误响应格式

---

## Test Cases

- [ ] 所有 Handler 测试通过
- [ ] 响应格式符合规范
- [ ] 错误处理正确
- [ ] 参数验证生效

---

## Related Docs

- [API 设计规范](../../standards/api-design.md)
- [Story 2: 响应工具包](./story-02-response-package.md)
- [Story 3: 错误处理](./story-03-error-handling.md)

---

**Created**: 2025-12-27  
**Updated**: 2025-12-27  
**Maintainer**: Architect Agent
