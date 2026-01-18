# Story 1.9: 重构现有 Handlers
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

## Refactoring Scope

**Target Files**:
- `core/handlers/config.go` - Configuration management handlers (PRIMARY)
- `core/handlers/project.go` - Project management handlers (SECONDARY)
- `core/handlers/user.go` - User management handlers (FUTURE)

**Target Patterns to Refactor**:
1. ❌ Direct `json.NewEncoder(w).Encode()` → ✅ Use `pkg/response`
2. ❌ Manual HTTP status code handling → ✅ Use `pkg/response` status mapping
3. ❌ String error messages → ✅ Use `pkg/errors` with error codes
4. ❌ Duplicate JSON parsing logic → ✅ Use `pkg/request` (Story 1.11)
5. ❌ Unstructured logging → ✅ Use `pkg/logger` with structured fields

**Estimated Impact**:
- ~150 lines of duplicated code removal
- ~8 handler functions refactored
- Test coverage increase from ~40% to >80%

---

## Acceptance Criteria

### AC-001: Response Package Integration ✓
- [ ] All handlers use `response.Success()` for 200 responses
- [ ] All handlers use `response.Error()` with error codes for failures
- [ ] All list endpoints use `response.List()` with pagination
- [ ] All validation errors use `response.ValidationError()`
- [ ] Remove all direct `json.NewEncoder(w).Encode()` calls

### AC-002: Error Handling Standardization ✓
- [ ] All database errors wrapped with `pkg/errors.Wrap()`
- [ ] All handlers use error codes from `pkg/errors/codes.go`
- [ ] Error responses include proper HTTP status mapping
- [ ] No bare string error messages in responses

### AC-003: Logger Integration ✓
- [ ] All handlers use `logger.L().WithContext(r.Context())`
- [ ] Request ID automatically injected in logs
- [ ] Structured logging fields for key operations
- [ ] Remove all `log.Printf()` calls

### AC-004: Code Quality ✓
- [ ] Remove duplicate JSON parsing logic (>50 lines reduction)
- [ ] golangci-lint passes with zero new warnings
- [ ] Test coverage ≥ 80% on refactored handlers
- [ ] All existing tests continue to pass (backward compatibility)

---

## Implementation Tasks

### Phase 1: Config Handlers Refactor (4 hours)
- [ ] Analyze current `handlers/config.go` (identify 8 functions to refactor)
- [ ] Refactor `ListConfigs()` - use `response.List()` with pagination
- [ ] Refactor `GetConfig()` - use `response.Success()`/`response.Error()`
- [ ] Refactor `CreateConfig()` - use `errors.Wrap()` + error codes
- [ ] Refactor `UpdateConfig()` - add `logger.L().WithContext()` logging
- [ ] Refactor `DeleteConfig()` - standardize error handling
- [ ] Update unit tests for config handlers (add 10+ test cases)
- [ ] Run integration tests to verify backward compatibility

### Phase 2: Project Handlers Refactor (3 hours)
- [ ] Analyze current `handlers/project.go` (identify functions)
- [ ] Apply same refactoring patterns as config handlers
- [ ] Add structured logging with project context
- [ ] Update tests

### Phase 3: Testing & Documentation (1 hour)
- [ ] Run full test suite: `go test ./handlers/... -v -cover`
- [ ] Verify coverage ≥ 80%: `go tool cover -func=coverage.out`
- [ ] Run linter: `golangci-lint run ./handlers/...`
- [ ] Update handler documentation with new patterns
- [ ] Create refactoring guide for future handlers

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
    "apprun/pkg/response"
    "apprun/pkg/errors"
    "apprun/pkg/logger"
)

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
    log := logger.L().WithContext(r.Context())
    id := chi.URLParam(r, "id")
    
    // 参数验证
    if id == "" {
        log.Warn("Missing config ID", logger.Field{"path", r.URL.Path})
        response.ErrorWithRequest(w, r, http.StatusBadRequest, 
            errors.ErrCodeInvalidParam, "config ID is required")
        return
    }
    
    config, err := h.repo.GetConfig(r.Context(), id)
    if err != nil {
        appErr := errors.Wrap(err, errors.ErrCodeNotFound, "Failed to get config")
        if errors.IsNotFound(err) {
            log.Info("Config not found", logger.Field{"config_id", id})
            response.AppErrorWithRequest(w, r, appErr)
        } else {
            log.Error("Database error", logger.Field{"error", err.Error()})
            response.AppErrorWithRequest(w, r, 
                errors.Wrap(err, errors.ErrCodeInternalError, "Internal error"))
        }
        return
    }
    
    log.Info("Config retrieved", logger.Field{"config_id", id})
    response.SuccessWithRequest(w, r, config)
}
```

---

## Definition of Done

- [ ] **Code Quality**
  - [ ] All acceptance criteria (AC-001 to AC-004) met
  - [ ] golangci-lint passes with zero new warnings
  - [ ] No code duplication (DRY principle applied)
  - [ ] Code follows project coding standards

- [ ] **Testing**
  - [ ] All unit tests passing (existing + new)
  - [ ] Test coverage ≥ 80% on refactored handlers
  - [ ] Integration tests passing (backward compatibility verified)
  - [ ] Manual testing completed (health check endpoints)

- [ ] **Review & Documentation**
  - [ ] Code reviewed and approved (2 reviewers minimum)
  - [ ] Refactoring patterns documented for team reference
  - [ ] Handler documentation updated with examples
  - [ ] PR description includes before/after metrics

- [ ] **Deployment Readiness**
  - [ ] No breaking changes to API contracts
  - [ ] Database connections properly closed
  - [ ] Error messages user-friendly and informative
  - [ ] Deployed to dev environment successfully

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
