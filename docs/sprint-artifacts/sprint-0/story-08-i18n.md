# Story 8: i18n 国际化Infrastructure
# Sprint 0: Infrastructure建设

**Priority**: P1  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: Story 1  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [API 设计规范](../../standards/api-design.md#i18n)

---

## User Story

作为开发者，我希望有国际化（i18n）Infrastructure，以便支持多语言响应消息。

---

## Acceptance Criteria

- [ ] 集成 `go-i18n` 库
- [ ] 创建翻译文件结构（TOML）
- [ ] 实现语言检测中间件
- [ ] 实现翻译辅助函数
- [ ] 支持 zh-CN、en-US
- [ ] 编写单元测试
- [ ] 编写 i18n 使用文档

---

## Implementation Tasks

- [ ] 添加 `go-i18n` 依赖
- [ ] 创建 `core/locales` 目录结构
- [ ] 创建翻译文件（active.zh-CN.toml、active.en-US.toml）
- [ ] 实现 `core/pkg/i18n` 包
- [ ] 实现语言检测中间件
- [ ] 实现 Translate 函数
- [ ] 更新错误消息支持 i18n
- [ ] 编写单元测试

---

## Technical Details

### 目录结构

```
core/
├── locales/
│   ├── active.zh-CN.toml
│   └── active.en-US.toml
└── pkg/
    └── i18n/
        ├── i18n.go
        └── middleware.go
```

### 翻译文件

```toml
# core/locales/active.zh-CN.toml

[errors.unauthorized]
other = "未授权访问"

[errors.not_found]
other = "资源未找到"

[errors.invalid_request]
other = "请求参数无效"

[success.created]
other = "创建成功"
```

```toml
# core/locales/active.en-US.toml

[errors.unauthorized]
other = "Unauthorized access"

[errors.not_found]
other = "Resource not found"

[errors.invalid_request]
other = "Invalid request parameters"

[success.created]
other = "Created successfully"
```

### i18n 包

```go
// core/pkg/i18n/i18n.go

package i18n

import (
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "golang.org/x/text/language"
)

var bundle *i18n.Bundle

func Init() error {
    bundle = i18n.NewBundle(language.English)
    bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
    
    bundle.LoadMessageFile("locales/active.en-US.toml")
    bundle.LoadMessageFile("locales/active.zh-CN.toml")
    
    return nil
}

func Translate(lang, messageID string, data map[string]interface{}) string {
    localizer := i18n.NewLocalizer(bundle, lang)
    
    msg, err := localizer.Localize(&i18n.LocalizeConfig{
        MessageID:    messageID,
        TemplateData: data,
    })
    
    if err != nil {
        return messageID // fallback
    }
    
    return msg
}
```

### 中间件

```go
// core/pkg/i18n/middleware.go

func LanguageDetector(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. 检查 Accept-Language header
        // 2. 检查查询参数 ?lang=zh-CN
        // 3. 默认 en-US
        
        lang := detectLanguage(r)
        ctx := context.WithValue(r.Context(), "lang", lang)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Test Cases

- [ ] 语言检测正确（header、query）
- [ ] 翻译函数返回正确文本
- [ ] 不存在的 messageID 返回 fallback
- [ ] 支持模板数据插值

---

## Related Docs

- [go-i18n 文档](https://github.com/nicksnyder/go-i18n)
- [Story 9: l10n 本地化](./story-09-l10n.md)

---

**Created**: 2025-12-27  
**Updated**: 2025-12-27  
**Maintainer**: Architect Agent
