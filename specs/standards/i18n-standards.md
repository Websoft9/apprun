# 国际化 (i18n) 规范
# apprun BaaS Platform

**创建日期**: 2025-12-26  
**维护者**: Winston (Architect Agent)  
**版本**: 1.0  
**状态**: Active

---

## 概述

本文档定义 apprun 项目的国际化（i18n）规范，基于 `go-i18n` v2 库实现，确保后端服务能够为全球用户提供多语言支持。

**核心原则**：
- **用户语言优先**：根据 `Accept-Language` 自动选择语言
- **英文为默认**：所有文本默认提供英文版本
- **集中管理**：消息文件统一存放，便于翻译和维护
- **性能优先**：预加载消息，避免运行时 I/O

---

## 目录

1. [总体策略](#1-总体策略)
2. [API 层面 i18n](#2-api-层面-i18n)
3. [业务逻辑层面 i18n](#3-业务逻辑层面-i18n)
4. [系统层面 i18n](#4-系统层面-i18n)
5. [实现指南](#5-实现指南)
6. [测试规范](#6-测试规范)
7. [最佳实践](#7-最佳实践)

---

## 1. 总体策略

### 1.1 支持的语言

| 语言代码 | 语言名称 | 优先级 | 状态 |
|---------|---------|-------|------|
| `en` | English（英语） | P0 | 必须 |
| `zh-CN` | 简体中文 | P0 | 必须 |
| `zh-TW` | 繁体中文 | P1 | 推荐 |
| `ja` | 日语 | P1 | 推荐 |
| `ko` | 韩语 | P2 | 可选 |
| `de` | 德语 | P2 | 可选 |
| `fr` | 法语 | P2 | 可选 |
| `es` | 西班牙语 | P2 | 可选 |

### 1.2 默认语言和 Fallback

```
Accept-Language: zh-CN,en;q=0.9
                    ↓
            1. 检查是否支持 zh-CN
                    ↓
            2. 支持 → 使用 zh-CN
                    ↓
            3. 不支持 → Fallback to en
```

**Fallback 规则**：
- 默认语言：`en`（英语）
- 未支持的语言 → 英语
- 部分翻译缺失 → 使用英语

### 1.3 语言检测机制

#### **优先级顺序**
1. **URL 参数**：`?lang=zh-CN`（优先级最高）
2. **HTTP Header**：`Accept-Language: zh-CN,en;q=0.9`
3. **用户偏好**：数据库中存储的用户语言设置
4. **默认语言**：`en`

#### **检测流程**
```go
func detectLanguage(r *http.Request, user *User) string {
    // 1. URL 参数
    if lang := r.URL.Query().Get("lang"); lang != "" {
        if isSupportedLanguage(lang) {
            return lang
        }
    }
    
    // 2. HTTP Header
    if lang := parseAcceptLanguage(r.Header.Get("Accept-Language")); lang != "" {
        if isSupportedLanguage(lang) {
            return lang
        }
    }
    
    // 3. 用户偏好
    if user != nil && user.Language != "" {
        return user.Language
    }
    
    // 4. 默认语言
    return "en"
}
```

---

## 2. API 层面 i18n

### 2.1 错误响应国际化

#### **需要翻译的字段**

```json
{
  "success": false,
  "code": 404,
  "message": "User not found",  // ← 需要翻译
  "error": {
    "code": "USER_NOT_FOUND",    // ← 错误码（不翻译）
    "message": "User not found",  // ← 需要翻译
    "details": {
      "user_id": "123"            // ← 业务数据（不翻译）
    }
  }
}
```

#### **实现示例**

```go
// core/handlers/user.go
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    localizer := i18n.FromContext(r.Context())
    
    user, err := h.userService.GetUser(r.Context(), userID)
    if err != nil {
        // 使用 i18n 错误消息
        msg := localizer.MustLocalize(&i18n.LocalizeConfig{
            MessageID: "user_not_found",
        })
        
        response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", msg)
        return
    }
    
    response.Success(w, user)
}
```

### 2.2 响应元数据国际化

#### **需要翻译的字段**

| 字段 | 是否翻译 | 示例 |
|-----|---------|------|
| `message` | ✅ | "Operation successful" → "操作成功" |
| `status` | ✅（如果用于显示） | "active" → "活跃" |
| `code` | ❌ | 200, 404, 500 |
| `request_id` | ❌ | "req-123" |
| `timestamp` | ❌ | "2025-12-26T10:00:00Z" |
| `pagination` | ❌（数字） | `{"page": 1, "size": 20}` |

#### **实现示例**

```go
// 成功消息
msg := localizer.MustLocalize(&i18n.LocalizeConfig{
    MessageID: "project_created",
})

response.Success(w, project, msg)
```

### 2.3 请求头处理

#### **Accept-Language 解析**

```go
// core/pkg/i18n/middleware.go

func parseAcceptLanguage(acceptLang string) string {
    // 解析格式: zh-CN,zh;q=0.9,en;q=0.8
    langs := strings.Split(acceptLang, ",")
    
    for _, lang := range langs {
        // 提取语言代码（忽略 q 值）
        parts := strings.Split(lang, ";")
        langCode := strings.TrimSpace(parts[0])
        
        // 检查是否支持
        if isSupportedLanguage(langCode) {
            return langCode
        }
    }
    
    return "en" // 默认英文
}

var supportedLanguages = map[string]bool{
    "en":    true,
    "zh-CN": true,
    "zh-TW": true,
    "ja":    true,
}

func isSupportedLanguage(lang string) bool {
    return supportedLanguages[lang]
}
```

---

## 3. 业务逻辑层面 i18n

### 3.1 验证消息

#### **表单验证错误**

```go
// core/internal/validator/validator.go

func (v *Validator) ValidateEmail(email string) error {
    if email == "" {
        return v.localizer.MustLocalize(&i18n.LocalizeConfig{
            MessageID: "email_required",
        })
    }
    
    if !emailRegex.MatchString(email) {
        return v.localizer.MustLocalize(&i18n.LocalizeConfig{
            MessageID: "invalid_email",
        })
    }
    
    return nil
}
```

#### **消息文件**

```yaml
# locales/en.yaml
email_required: "Email is required"
invalid_email: "Invalid email format"
password_too_short: "Password must be at least {{.MinLength}} characters"

# locales/zh-CN.yaml
email_required: "邮箱不能为空"
invalid_email: "邮箱格式不正确"
password_too_short: "密码长度至少 {{.MinLength}} 个字符"
```

### 3.2 状态描述

#### **枚举值显示**

```go
// 数据库存储 key，应用层翻译显示
type ProjectStatus string

const (
    ProjectStatusActive   ProjectStatus = "active"
    ProjectStatusInactive ProjectStatus = "inactive"
    ProjectStatusArchived ProjectStatus = "archived"
)

func (s ProjectStatus) Localize(localizer *i18n.Localizer) string {
    return localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: fmt.Sprintf("project_status_%s", s),
    })
}
```

```yaml
# locales/en.yaml
project_status_active: "Active"
project_status_inactive: "Inactive"
project_status_archived: "Archived"

# locales/zh-CN.yaml
project_status_active: "活跃"
project_status_inactive: "未激活"
project_status_archived: "已归档"
```

### 3.3 业务规则文本

#### **配额限制提示**

```go
// core/internal/service/storage.go

func (s *StorageService) UploadFile(ctx context.Context, file *File) error {
    localizer := i18n.FromContext(ctx)
    
    // 检查配额
    if s.exceedsQuota(file.Size) {
        msg := localizer.MustLocalize(&i18n.LocalizeConfig{
            MessageID: "storage_quota_exceeded",
            TemplateData: map[string]interface{}{
                "Used":  formatSize(s.usedStorage),
                "Limit": formatSize(s.storageLimit),
            },
        })
        
        return errors.New(msg)
    }
    
    // ... 上传逻辑
}
```

```yaml
# locales/en.yaml
storage_quota_exceeded: "Storage quota exceeded ({{.Used}}/{{.Limit}})"

# locales/zh-CN.yaml
storage_quota_exceeded: "存储配额已超限（{{.Used}}/{{.Limit}}）"
```

---

## 4. 系统层面 i18n

### 4.1 日志消息

#### **结构化日志**

```go
// core/internal/logger/logger.go

type Logger struct {
    localizer *i18n.Localizer
}

func (l *Logger) InfoWithI18n(messageID string, data map[string]interface{}) {
    msg := l.localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID:    messageID,
        TemplateData: data,
    })
    
    log.Info().
        Str("message_id", messageID).
        Str("message", msg).
        Interface("data", data).
        Msg("")
}
```

**建议**：日志消息可以不翻译（内部使用），但关键用户可见的日志应支持 i18n。

### 4.2 邮件通知

#### **邮件模板**

```go
// core/internal/notification/email.go

func (n *EmailService) SendWelcomeEmail(user *User, lang string) error {
    localizer := i18n.NewLocalizer(n.bundle, lang)
    
    subject := localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "email_welcome_subject",
    })
    
    body := localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "email_welcome_body",
        TemplateData: map[string]string{
            "Name": user.Name,
        },
    })
    
    return n.send(user.Email, subject, body)
}
```

```yaml
# locales/en.yaml
email_welcome_subject: "Welcome to apprun!"
email_welcome_body: "Hi {{.Name}},\n\nWelcome to apprun BaaS Platform!"

# locales/zh-CN.yaml
email_welcome_subject: "欢迎使用 apprun！"
email_welcome_body: "你好 {{.Name}}，\n\n欢迎使用 apprun BaaS 平台！"
```

### 4.3 系统通知

#### **WebSocket 推送消息**

```go
// core/internal/notification/websocket.go

type Notification struct {
    Type      string                 `json:"type"`
    MessageID string                 `json:"message_id"`
    Data      map[string]interface{} `json:"data"`
}

// 客户端负责翻译
func (ws *WebSocketService) SendNotification(userID string, notif *Notification) {
    // 发送包含 message_id 的通知
    // 前端根据用户语言翻译
    ws.sendToUser(userID, notif)
}
```

**建议**：实时通知发送 `message_id` 和 `data`，由前端根据用户语言翻译。

---

## 5. 实现指南

### 5.1 技术选型

**推荐库**：`github.com/nicksnyder/go-i18n/v2`

**选择理由**：
- Go 社区最流行的 i18n 库
- 支持复数规则（ICU 标准）
- 支持模板变量
- 支持 YAML/JSON/TOML 多种格式
- 性能优秀（预加载）

### 5.2 项目结构

```
apprun/
├── locales/                    # 消息文件目录
│   ├── en.yaml                # 英文消息
│   ├── zh-CN.yaml             # 简体中文
│   ├── zh-TW.yaml             # 繁体中文
│   └── ja.yaml                # 日语
├── core/
│   └── pkg/
│       └── i18n/
│           ├── i18n.go        # i18n 初始化
│           ├── middleware.go   # Chi 中间件
│           └── i18n_test.go   # 单元测试
```

### 5.3 初始化代码

```go
// core/pkg/i18n/i18n.go
package i18n

import (
    "embed"
    "fmt"
    
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "golang.org/x/text/language"
    "gopkg.in/yaml.v3"
)

//go:embed ../../locales/*.yaml
var localeFS embed.FS

var Bundle *i18n.Bundle

// Init 初始化 i18n Bundle
func Init() error {
    // 创建 Bundle，默认语言为英语
    Bundle = i18n.NewBundle(language.English)
    
    // 注册 YAML 解析器
    Bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
    
    // 加载所有语言文件
    languages := []string{"en", "zh-CN", "zh-TW", "ja"}
    for _, lang := range languages {
        filename := fmt.Sprintf("locales/%s.yaml", lang)
        _, err := Bundle.LoadMessageFileFS(localeFS, filename)
        if err != nil {
            return fmt.Errorf("failed to load %s: %w", filename, err)
        }
    }
    
    return nil
}

// FromContext 从上下文获取 Localizer
func FromContext(ctx context.Context) *i18n.Localizer {
    lang, ok := ctx.Value("accept-language").(string)
    if !ok {
        lang = "en"
    }
    
    return i18n.NewLocalizer(Bundle, lang)
}
```

### 5.4 中间件集成

```go
// core/pkg/i18n/middleware.go
package i18n

import (
    "context"
    "net/http"
    "strings"
)

// Middleware Chi 中间件
func Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. 检测语言
        lang := detectLanguage(r)
        
        // 2. 存入上下文
        ctx := context.WithValue(r.Context(), "accept-language", lang)
        
        // 3. 设置响应头（可选）
        w.Header().Set("Content-Language", lang)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func detectLanguage(r *http.Request) string {
    // 1. URL 参数（优先级最高）
    if lang := r.URL.Query().Get("lang"); lang != "" {
        if isSupportedLanguage(lang) {
            return lang
        }
    }
    
    // 2. Accept-Language Header
    if lang := parseAcceptLanguage(r.Header.Get("Accept-Language")); lang != "" {
        return lang
    }
    
    // 3. 默认英文
    return "en"
}

func parseAcceptLanguage(acceptLang string) string {
    langs := strings.Split(acceptLang, ",")
    
    for _, lang := range langs {
        // 提取语言代码（忽略 q 值）
        parts := strings.Split(lang, ";")
        langCode := strings.TrimSpace(parts[0])
        
        if isSupportedLanguage(langCode) {
            return langCode
        }
    }
    
    return ""
}

var supportedLanguages = map[string]bool{
    "en":    true,
    "zh-CN": true,
    "zh-TW": true,
    "ja":    true,
}

func isSupportedLanguage(lang string) bool {
    return supportedLanguages[lang]
}
```

### 5.5 消息文件格式

#### **基本消息**

```yaml
# locales/en.yaml
user_not_found: "User not found"
invalid_email: "Invalid email format"
project_created: "Project created successfully"
```

#### **带变量的消息**

```yaml
# locales/en.yaml
welcome_user: "Welcome, {{.Name}}!"
file_uploaded: "File uploaded: {{.FileName}} ({{.FileSize}})"
storage_quota: "Storage: {{.Used}}/{{.Limit}}"
```

#### **复数规则**

```yaml
# locales/en.yaml
items_count:
  one: "{{.PluralCount}} item"
  other: "{{.PluralCount}} items"

# locales/zh-CN.yaml
items_count:
  other: "{{.PluralCount}} 个项目"
```

使用复数：
```go
msg := localizer.MustLocalize(&i18n.LocalizeConfig{
    MessageID:   "items_count",
    PluralCount: 5,
})
// en: "5 items"
// zh-CN: "5 个项目"
```

---

## 6. 测试规范

### 6.1 单元测试

```go
// core/pkg/i18n/i18n_test.go
package i18n_test

import (
    "testing"
    
    "apprun/core/pkg/i18n"
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "github.com/stretchr/testify/assert"
)

func TestLocalizer_English(t *testing.T) {
    // 初始化
    err := i18n.Init()
    assert.NoError(t, err)
    
    // 创建英文 Localizer
    localizer := i18n.NewLocalizer(i18n.Bundle, "en")
    
    // 测试基本消息
    msg := localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "user_not_found",
    })
    
    assert.Equal(t, "User not found", msg)
}

func TestLocalizer_ChineseSimplified(t *testing.T) {
    err := i18n.Init()
    assert.NoError(t, err)
    
    localizer := i18n.NewLocalizer(i18n.Bundle, "zh-CN")
    
    msg := localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "user_not_found",
    })
    
    assert.Equal(t, "用户不存在", msg)
}

func TestLocalizer_WithVariables(t *testing.T) {
    err := i18n.Init()
    assert.NoError(t, err)
    
    localizer := i18n.NewLocalizer(i18n.Bundle, "en")
    
    msg := localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "welcome_user",
        TemplateData: map[string]string{
            "Name": "Alice",
        },
    })
    
    assert.Equal(t, "Welcome, Alice!", msg)
}

func TestLocalizer_Fallback(t *testing.T) {
    err := i18n.Init()
    assert.NoError(t, err)
    
    // 不支持的语言，应 fallback 到英文
    localizer := i18n.NewLocalizer(i18n.Bundle, "fr")
    
    msg := localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "user_not_found",
    })
    
    assert.Equal(t, "User not found", msg)
}
```

### 6.2 集成测试

```go
// tests/integration/i18n_test.go
func TestAPI_I18n(t *testing.T) {
    client := setupTestClient(t)
    
    tests := []struct {
        name           string
        acceptLanguage string
        expectedMsg    string
    }{
        {
            name:           "English",
            acceptLanguage: "en",
            expectedMsg:    "User not found",
        },
        {
            name:           "Chinese Simplified",
            acceptLanguage: "zh-CN",
            expectedMsg:    "用户不存在",
        },
        {
            name:           "Chinese Traditional",
            acceptLanguage: "zh-TW",
            expectedMsg:    "使用者不存在",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/api/v1/users/999", nil)
            req.Header.Set("Accept-Language", tt.acceptLanguage)
            
            w := httptest.NewRecorder()
            client.ServeHTTP(w, req)
            
            assert.Equal(t, http.StatusNotFound, w.Code)
            
            var resp map[string]interface{}
            json.Unmarshal(w.Body.Bytes(), &resp)
            
            assert.Equal(t, tt.expectedMsg, resp["message"])
        })
    }
}
```

### 6.3 测试覆盖清单

- [ ] 所有支持的语言都有测试用例
- [ ] 测试 Fallback 机制
- [ ] 测试变量替换
- [ ] 测试复数规则
- [ ] 测试中间件语言检测
- [ ] 测试 API 响应多语言

---

## 7. 最佳实践

### 7.1 消息 ID 命名规范

```yaml
# 推荐格式: <category>_<object>_<action/state>
user_not_found          # ✅ 清晰明确
invalid_email           # ✅ 简洁
project_created         # ✅ 动作明确
storage_quota_exceeded  # ✅ 完整描述

# 避免格式
err1                    # ❌ 不明确
message                 # ❌ 太泛化
user_error              # ❌ 不够具体
```

### 7.2 变量命名规范

```yaml
# 使用 PascalCase（首字母大写）
welcome_user: "Welcome, {{.Name}}!"              # ✅
file_info: "{{.FileName}} ({{.FileSize}})"      # ✅

# 避免格式
welcome: "Welcome, {{.name}}!"                   # ❌ 小写
file: "{{.file_name}} ({{.file_size}})"         # ❌ 下划线
```

### 7.3 消息文件组织

```yaml
# 按功能模块组织
# locales/en.yaml

# === Authentication ===
auth_login_success: "Login successful"
auth_login_failed: "Invalid credentials"
auth_token_expired: "Token has expired"

# === User Management ===
user_not_found: "User not found"
user_created: "User created successfully"
user_updated: "User updated successfully"

# === Project Management ===
project_created: "Project created successfully"
project_not_found: "Project not found"
```

### 7.4 避免硬编码

```go
// ❌ 错误示例
return errors.New("User not found")

// ✅ 正确示例
localizer := i18n.FromContext(ctx)
msg := localizer.MustLocalize(&i18n.LocalizeConfig{
    MessageID: "user_not_found",
})
return errors.New(msg)
```

### 7.5 性能优化

```go
// ✅ 预加载消息（在 Init 时完成）
func Init() error {
    Bundle = i18n.NewBundle(language.English)
    Bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
    
    // 一次性加载所有语言文件
    for _, lang := range languages {
        Bundle.LoadMessageFileFS(localeFS, filename)
    }
    
    return nil
}

// ✅ 从上下文复用 Localizer
func handler(w http.ResponseWriter, r *http.Request) {
    localizer := i18n.FromContext(r.Context())
    
    // 多次使用同一个 Localizer
    msg1 := localizer.MustLocalize(...)
    msg2 := localizer.MustLocalize(...)
}

// ❌ 避免频繁创建 Localizer
func handler(w http.ResponseWriter, r *http.Request) {
    // 不要每次都创建新的 Localizer
    localizer1 := i18n.NewLocalizer(Bundle, "en")
    localizer2 := i18n.NewLocalizer(Bundle, "en")
}
```

### 7.6 翻译工作流

```
1. 开发者添加英文消息 (en.yaml)
        ↓
2. 提交 PR，CI 检查翻译完整性
        ↓
3. 翻译人员翻译其他语言 (zh-CN.yaml, ja.yaml)
        ↓
4. Review 翻译质量
        ↓
5. 合并 PR，部署
```

### 7.7 翻译完整性检查

```bash
#!/bin/bash
# scripts/check-i18n-completeness.sh

# 检查所有语言文件是否包含相同的 message ID

EN_KEYS=$(yq eval 'keys | .[]' locales/en.yaml | sort)
ZH_KEYS=$(yq eval 'keys | .[]' locales/zh-CN.yaml | sort)

DIFF=$(diff <(echo "$EN_KEYS") <(echo "$ZH_KEYS"))

if [ -n "$DIFF" ]; then
    echo "❌ Translation incomplete:"
    echo "$DIFF"
    exit 1
else
    echo "✅ All translations complete"
fi
```

---

## 附录

### A. 消息 ID 参考

常用消息 ID 模板：

```yaml
# 认证相关
auth_login_success
auth_login_failed
auth_logout_success
auth_token_invalid
auth_token_expired
auth_unauthorized

# 权限相关
perm_forbidden
perm_insufficient_role
perm_access_denied

# 资源相关
res_not_found
res_already_exists
res_created
res_updated
res_deleted

# 验证相关
validation_email_required
validation_email_invalid
validation_password_too_short
validation_field_required

# 系统相关
system_error
system_maintenance
system_unavailable
```

### B. 相关资源

- **go-i18n 文档**: https://github.com/nicksnyder/go-i18n
- **ICU 复数规则**: https://unicode-org.github.io/cldr-staging/charts/latest/supplemental/language_plural_rules.html
- **语言代码标准**: ISO 639-1 (zh, en, ja)
- **区域代码标准**: ISO 3166-1 (CN, US, JP)

### C. 相关文档

- [API 设计规范](./api-design.md) - API 响应格式
- [编码规范](./coding-standards.md) - Go 代码规范
- [DevOps 流程规范](./devops-process.md) - 开发流程

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26  
**审核状态**: Active
