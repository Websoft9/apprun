# Story 9: l10n 本地化实施
# Sprint 0: Infrastructure建设

**Priority**: P1  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: Story 8  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [API 设计规范](../../standards/api-design.md#i18n)

---

## User Story

作为开发者，我希望在现有代码中应用本地化（l10n），以便所有用户可见消息支持多语言。

---

## Acceptance Criteria

- [ ] 更新 response 包支持 i18n
- [ ] 更新 errors 包支持 i18n
- [ ] 翻译所有错误消息
- [ ] 翻译所有成功消息
- [ ] 更新 Handlers 使用 i18n
- [ ] 编写 l10n 测试
- [ ] 更新 API 文档

---

## Implementation Tasks

- [ ] 修改 `pkg/response` 包（接受 lang 参数）
- [ ] 修改 `pkg/errors` 包（支持翻译）
- [ ] 翻译所有错误码消息
- [ ] 更新 Handlers（从 context 获取 lang）
- [ ] 添加翻译覆盖率检查
- [ ] 编写 l10n 测试
- [ ] 更新 API 文档（Accept-Language）

---

## Technical Details

### 更新 Response 包

```go
// core/pkg/response/response.go

import "github.com/yourusername/apprun/core/pkg/i18n"

func Success(w http.ResponseWriter, data interface{}, lang string) {
    resp := Response{
        Success: true,
        Code:    http.StatusOK,
        Message: i18n.Translate(lang, "success.ok", nil),
        Data:    data,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func Error(w http.ResponseWriter, code int, errCode, messageID string, lang string) {
    resp := Response{
        Success: false,
        Code:    code,
        Error: &ErrorInfo{
            Code:    errCode,
            Message: i18n.Translate(lang, messageID, nil),
        },
    }
    
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(resp)
}
```

### 更新 Handlers

```go
// handlers/config.go

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
    lang := r.Context().Value("lang").(string)
    
    id := chi.URLParam(r, "id")
    if id == "" {
        response.Error(w, http.StatusBadRequest, errors.ErrInvalidRequest, "errors.invalid_request", lang)
        return
    }
    
    config, err := h.repo.GetConfig(r.Context(), id)
    if err != nil {
        response.Error(w, http.StatusNotFound, errors.ErrResourceNotFound, "errors.not_found", lang)
        return
    }
    
    response.Success(w, config, lang)
}
```

### 翻译文件更新

```toml
# core/locales/active.zh-CN.toml

[success.ok]
other = "操作成功"

[success.created]
other = "创建成功"

[success.updated]
other = "更新成功"

[success.deleted]
other = "删除成功"

[errors.unauthorized]
other = "未授权访问"

[errors.forbidden]
other = "禁止访问"

[errors.not_found]
other = "资源未找到"

[errors.invalid_request]
other = "请求参数无效"

[errors.internal_error]
other = "服务器内部错误"

[config.key_required]
other = "配置键不能为空"

[config.value_required]
other = "配置值不能为空"
```

```toml
# core/locales/active.en-US.toml

[success.ok]
other = "Operation successful"

[success.created]
other = "Created successfully"

[success.updated]
other = "Updated successfully"

[success.deleted]
other = "Deleted successfully"

[errors.unauthorized]
other = "Unauthorized access"

[errors.forbidden]
other = "Forbidden"

[errors.not_found]
other = "Resource not found"

[errors.invalid_request]
other = "Invalid request parameters"

[errors.internal_error]
other = "Internal server error"

[config.key_required]
other = "Config key is required"

[config.value_required]
other = "Config value is required"
```

---

## 翻译覆盖率检查

```bash
# scripts/check-translations.sh

#!/bin/bash
# 检查所有语言的翻译文件是否包含相同的 messageID

LANG_FILES="core/locales/active.*.toml"

# 提取所有 messageID 并比较
```

---

## Test Cases

- [ ] zh-CN 响应消息正确
- [ ] en-US 响应消息正确
- [ ] 所有错误消息已翻译
- [ ] 所有成功消息已翻译
- [ ] 翻译覆盖率 100%

---

## Related Docs

- [API 设计规范](../../standards/api-design.md#i18n)
- [Story 8: i18n Infrastructure](./story-08-i18n.md)

---

**Created**: 2025-12-27  
**Updated**: 2025-12-27  
**Maintainer**: Architect Agent
