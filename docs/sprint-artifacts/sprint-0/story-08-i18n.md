# Story 8: i18n 国际化基础设施
# Sprint 0: Infrastructure建设

**Priority**: P1  
**Effort**: 1.5 天  
**Owner**: Backend Dev  
**Dependencies**: Story 1  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  

---

## User Story

作为开发者，我希望有独立的国际化（i18n）基础设施，以便支持多语言响应消息。

---

## Business Value

- 用户可使用母语查看系统消息
- 翻译文本与代码分离，便于维护
- 轻松添加新语言支持

---

## Acceptance Criteria

### 功能要求
- [ ] 集成 `go-i18n` 库 (v2)
- [ ] 创建翻译文件（TOML 格式）
- [ ] 创建语言元数据文件（`languages.yaml`）
- [ ] 实现语言元数据管理（加载、查询、API）
- [ ] 实现语言检测中间件
- [ ] 实现翻译函数（支持 Context）
- [ ] 支持 zh-CN、en-US 两种语言

### 架构要求
- [ ] pkg/i18n 包独立运行，无外部依赖
- [ ] 翻译文件在启动时从指定路径加载
- [ ] 支持翻译缺失的降级处理
- [ ] 语言检测失败时使用默认语言

### 质量要求
- [ ] 单元测试覆盖率 ≥ 70%
- [ ] 编写使用文档
- [ ] 性能：翻译函数调用 < 1ms

---

## Implementation Tasks

### Phase 1: 依赖和文件结构 (30 分钟)
- [ ] 添加依赖：`go get github.com/nicksnyder/go-i18n/v2 github.com/BurntSushi/toml`
- [ ] 创建 `core/locales/` 目录
- [ ] 创建 `core/locales/languages.yaml`（语言元数据配置）
- [ ] 创建 `core/pkg/i18n/` 目录
- [ ] 创建 `core/pkg/middleware/` 目录

### Phase 2: 翻译文件 (45 分钟)
- [ ] 创建 `active.zh-CN.toml`（中文翻译）
- [ ] 创建 `active.en-US.toml`（英文翻译）
- [ ] 添加常用错误和成功消息

### Phase 3: pkg/i18n 核心包 (2 小时)
- [ ] 实现 `i18n.go`：Init、Translate、IsSupported 等函数
- [ ] 实现 `languages.go`：LoadLanguageMetadata、GetSupportedLanguages 等函数
- [ ] 实现 `context.go`：Context 辅助函数
- [ ] 编写单元测试

### Phase 4: 语言检测中间件 (1.5 小时)
- [ ] 实现 `middleware/language.go`
- [ ] 支持 Query、Header、Cookie 检测
- [ ] 编写中间件测试

### Phase 5: 集成和文档 (1.5 小时)
- [ ] 在 `main.go` 中初始化 i18n
- [ ] 在路由中注册中间件
- [ ] 编写使用文档和示例
- [ ] 运行所有测试

---

## Technical Details

### 目录结构

```
core/
├── locales/
│   ├── languages.yaml        # 语言元数据（新增）
│   ├── active.zh-CN.toml
│   └── active.en-US.toml
├── pkg/
│   ├── i18n/
│   │   ├── i18n.go
│   │   ├── languages.go      # 语言元数据管理（新增）
│   │   ├── context.go
│   │   └── i18n_test.go
│   └── middleware/
│       ├── language.go
│       └── language_test.go
└── main.go
```

### 语言元数据管理

**目的**：维护支持的语言列表及其显示名称，用于：
- 前端语言选择器（下拉菜单）
- API 返回支持的语言列表
- 语言验证和国际化信息展示

**维护位置**：`core/locales/languages.yaml`
```yaml
# 语言元数据配置文件
languages:
  en-US:
    code: "en-US"           # BCP 47 标准代码
    name: "English"         # 原生名称
    display_name: "English (United States)"  # 显示名称
    direction: "ltr"        # 文字方向 (ltr/rtl)
    enabled: true           # 是否启用
  zh-CN:
    code: "zh-CN"
    name: "中文"
    display_name: "中文（简体）"
    direction: "ltr"
    enabled: true
  zh-TW:
    code: "zh-TW"
    name: "中文"
    display_name: "中文（繁體）"
    direction: "ltr"
    enabled: false          # 预留，暂未启用
  ar-SA:
    code: "ar-SA"
    name: "العربية"
    display_name: "Arabic (Saudi Arabia)"
    direction: "rtl"        # 阿拉伯语从右到左
    enabled: false
```

**API 设计**：
```go
// pkg/i18n/languages.go
type LanguageMetadata struct {
    Code       string `json:"code" yaml:"code"`
    NativeName string `json:"native_name" yaml:"native_name"` // 原生名称 (English, 中文)
    Icon       string `json:"icon" yaml:"icon"`               // 图标/Emoji (🇺🇸, 🇨🇳)
    Direction  string `json:"direction" yaml:"direction"`
    Enabled    bool   `json:"enabled" yaml:"enabled"`
    IsDefault  bool   `json:"is_default" yaml:"is_default"`   // 是否为系统默认语言
}

// GetSupportedLanguages 获取所有已启用的语言
func GetSupportedLanguages() []LanguageMetadata

// GetLanguageMetadata 获取指定语言的元数据
func GetLanguageMetadata(code string) *LanguageMetadata
```

**配置示例**：
```yaml
# languages.yaml
languages:
  en-US:
    code: "en-US"
    native_name: "English"
    icon: "🇺🇸"
    direction: "ltr"
    enabled: true
    is_default: true
  zh-CN:
    code: "zh-CN"
    native_name: "中文"
    icon: "🇨🇳"
    direction: "ltr"
    enabled: true
    is_default: false
```

**使用场景**：
```go
// 1. API: 获取支持的语言列表
GET /api/languages
Response: [
  {"code": "en-US", "native_name": "English", "icon": "🇺🇸", "direction": "ltr", "is_default": true},
  {"code": "zh-CN", "native_name": "中文", "icon": "🇨🇳", "direction": "ltr", "is_default": false}
]

// 2. 前端语言选择器
// 使用 native_name 确保用户能识别自己的语言
<select>
  <option value="en-US">🇺🇸 English</option>
  <option value="zh-CN">🇨🇳 中文</option>
</select>
```

---

### 翻译文件示例

**locales/active.zh-CN.toml**
```toml
[errors.invalid_param]
other = "请求参数无效"

[errors.not_found]
other = "资源未找到"

[success.created]
other = "创建成功"
```

**locales/active.en-US.toml**: 英文翻译（结构相同）

### 核心函数签名

**pkg/i18n/i18n.go**
```go
// 初始化（启动时调用一次，从指定路径加载翻译文件）
func Init(defaultLang string, supportedLangs []string, translationsPath string) error

// 翻译（基础版本）
func Translate(lang, messageID string, data map[string]interface{}) string

// 获取支持的语言列表（已启用）
func GetSupportedLanguages() []LanguageMetadata

// 检查语言是否支持
func IsSupported(lang string) bool
```

**pkg/i18n/languages.go** (新增)
```go
// 语言元数据
type LanguageMetadata struct {
    Code        string `json:"code" yaml:"code"`
    Name        string `json:"name" yaml:"name"`
    DisplayName string `json:"display_name" yaml:"display_name"`
    Direction   string `json:"direction" yaml:"direction"`
    Enabled     bool   `json:"enabled" yaml:"enabled"`
}

// 加载语言元数据（从 languages.yaml）
func LoadLanguageMetadata(path string) error

// 获取所有已启用的语言元数据
func GetSupportedLanguages() []LanguageMetadata

// 获取指定语言的元数据
func GetLanguageMetadata(code string) *LanguageMetadata

// 获取语言的显示名称
func GetLanguageName(code string) string
```

**pkg/i18n/context.go**
```go
// 将语言存入 Context
func WithLanguage(ctx context.Context, lang string) context.Context

// 从 Context 获取语言
func GetLanguage(ctx context.Context) string

// 从 Context 自动获取语言并翻译（推荐使用）
func TranslateContext(ctx context.Context, messageID string, data map[string]interface{}) string
```

**pkg/middleware/language.go**
```go
// 语言检测中间件
func LanguageDetector() func(next http.Handler) http.Handler
```

### 语言检测优先级

1. URL Query: `?lang=zh-CN` （最高优先级）
2. Cookie: `lang=zh-CN`
3. Accept-Language Header: `Accept-Language: zh-CN,en;q=0.9`
4. 系统默认: `en-US` （最低优先级）

---

## Architecture Guidelines

### 职责分离

**pkg/i18n 包职责**（核心翻译引擎）：
- **初始化与加载**：启动时从指定路径加载所有翻译文件到内存（TOML 格式）
- **翻译服务**：提供 `Translate()` 和 `TranslateContext()` 函数
- **语言管理**：管理支持的语言列表，验证语言代码
- **错误处理**：翻译键不存在时返回 messageID 作为 fallback
- **独立性**：不依赖 HTTP 或其他业务模块，可被任何组件调用

**中间件职责**（HTTP 层）：
- 从请求中检测用户语言（Query > Cookie > Header > Default）
- 验证语言是否支持
- 将语言存入 Context
- **不执行翻译**（职责分离）

**Handler 使用方式**：
- 从 Context 获取语言（自动）
- 调用 `i18n.TranslateContext()` 获取本地化文本
- 返回响应给用户

### 错误处理策略

| 错误场景 | 处理方式 |
|---------|---------|
| 翻译文件缺失 | 警告日志 + 继续运行 |
| 翻译键不存在 | 返回 messageID 作为 fallback |
| 不支持的语言 | 降级到默认语言 |
| 初始化失败 | 程序退出（启动时必须成功）|

### 性能要点

- ✅ 启动时加载所有翻译到内存
- ✅ 运行时无文件 I/O
- ✅ Bundle 是线程安全的，支持并发
- ❌ 避免每次请求重新创建 Bundle
- ❌ 避免在中间件中执行翻译逻辑

---

## Test Cases

### 单元测试

**pkg/i18n 测试**
- [ ] 初始化成功，加载所有语言文件
- [ ] 语言元数据加载和查询正常
- [ ] 翻译返回正确的本地化文本
- [ ] 翻译键不存在时返回 fallback
- [ ] 支持模板变量插值
- [ ] Context 函数正确存取语言
- [ ] `GetSupportedLanguages()` 返回正确的元数据列表

**中间件测试**
- [ ] Query 参数优先级最高
- [ ] Cookie 检测正常工作
- [ ] Accept-Language Header 解析正确
- [ ] 不支持的语言降级到默认语言

### 集成测试

- [ ] 端到端语言切换（Query）
- [ ] 中文和英文响应正确
- [ ] 并发请求语言隔离

### 性能测试

- [ ] 翻译函数调用 < 1ms (P99)
- [ ] 启动时间增加 < 100ms

---

## Usage Example

### 初始化（main.go）
```go
// 1. 初始化 i18n（加载翻译文件和语言元数据）
i18n.Init("en-US", []string{"en-US", "zh-CN"}, "locales")
i18n.LoadLanguageMetadata("locales/languages.yaml")

// 2. 注册中间件
r.Use(middleware.LanguageDetector())

// 3. 提供语言列表 API
r.Get("/api/languages", func(w http.ResponseWriter, r *http.Request) {
    languages := i18n.GetSupportedLanguages()
    json.NewEncoder(w).Encode(languages)
})
```

### Handler 中使用
```go
func demoHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    msg := i18n.TranslateContext(ctx, "success.created", nil)
    w.Write([]byte(msg))
}
```

### 测试语言切换
```bash
# 测试翻译
curl "http://localhost:8080/api/demo?lang=zh-CN"  # 中文
curl "http://localhost:8080/api/demo?lang=en-US"  # 英文
curl -H "Accept-Language: zh-CN" "http://localhost:8080/api/demo"  # Header

# 获取支持的语言列表
curl "http://localhost:8080/api/languages"
# Response: [
#   {"code": "en-US", "native_name": "English", "icon": "🇺🇸", "direction": "ltr", "enabled": true, "is_default": true},
#   {"code": "zh-CN", "native_name": "中文", "icon": "🇨🇳", "direction": "ltr", "enabled": true, "is_default": false}
# ]
```

---

## Developer Notes

### 常见错误

1. **忘记初始化**：在使用翻译前必须调用 `i18n.Init()`
2. **硬编码语言**：应从 Context 获取语言，而非硬编码
3. **中间件位置**：语言检测中间件应在业务 Handler 之前注册
4. **日志语言**：日志应使用英文，翻译仅用于用户响应

### 开发检查清单

- [ ] `i18n.Init()` 在 `main.go` 中调用
- [ ] 所有消息都有 zh-CN 和 en-US 翻译
- [ ] Handler 使用 `TranslateContext()` 获取翻译
- [ ] 中间件已在路由中注册
- [ ] 单元测试覆盖率 ≥ 70%
- [ ] 不支持的语言能优雅降级

---

## Related Docs

- [go-i18n 官方文档](https://github.com/nicksnyder/go-i18n)
- [Story 9: l10n 本地化](./story-09-l10n.md)

---

## 与 Story 09 (l10n) 的关系

**Story 08 (i18n)** 和 **Story 09 (l10n)** 是互补关系，各司其职：

| 维度 | Story 08 (i18n) | Story 09 (l10n) |
|------|-----------------|-----------------|
| **职责** | 文本翻译和语言管理 | 数字/货币/时间格式化 |
| **核心功能** | 翻译消息、错误码 | 格式化显示值 |
| **元数据** | 语言名称、方向 | 国家格式规则、时区 |
| **配置文件** | `languages.yaml`, `active.*.toml` | `countries.yaml` (可选) |
| **用户配置** | `user.language` | `user.country`, `user.timezone` |
| **依赖关系** | 独立 | 依赖 Story 08 的语言信息 |

**集成示例**：
```go
func GetProduct(w http.ResponseWriter, r *http.Request) {
    // Story 08: 获取语言（用于翻译）
    lang := i18n.GetLanguage(r.Context())
    
    // Story 09: 获取国家和时区（用于格式化）
    l10nInfo := l10n.FromContext(r.Context())
    
    product := productRepo.FindByID(123)
    
    response.Success(w, map[string]interface{}{
        "id": product.ID,
        "name": i18n.TranslateContext(r.Context(), "product.name", nil), // Story 08
        "price": l10n.FormatCurrency(product.Price, "USD", l10nInfo.Country), // Story 09
        "created_at": l10n.FormatTime(product.CreatedAt, l10nInfo.Timezone, lang), // Story 09
    })
}
```

---

## Success Criteria

### 功能验收
- [ ] 用户可通过 `?lang=zh-CN` 切换语言
- [ ] 错误消息根据用户语言显示
- [ ] 不支持的语言自动降级
- [ ] 翻译缺失时系统不崩溃

### 性能验收
- [ ] 启动时间增加 < 100ms
- [ ] 翻译调用延迟 < 1ms (P99)
- [ ] 并发请求无性能问题

### 质量验收
- [ ] 单元测试覆盖率 ≥ 70%
- [ ] 代码审查通过
- [ ] 文档完整

---

**Created**: 2025-12-27  
**Updated**: 2026-01-05 (简化版)  
**Maintainer**: Winston (Architect Agent)

---

## Changelog

### 2026-01-05 - Simplified Version
- 移除所有外部依赖（Story 7, Story 10）
- 移除伪代码，仅保留函数签名
- 简化为独立运行的 i18n 模块
- 文档精简到 ~280 行
- 专注于核心功能和实用性
