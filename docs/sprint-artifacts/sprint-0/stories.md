# Sprint 0: 基础设施建设
# apprun BaaS Platform

**Sprint 周期**: 2025-12-26 ~ 2026-01-09 (2 周)  
**Sprint 目标**: 搭建开发基础设施，建立编码规范和工具链  
**负责人**: Dev Team Lead  
**状态**: Planning

---

## Sprint 目标

### 核心目标
实现通用技术规范的基础代码，为后续业务 Epic 开发提供标准化工具和框架。

### 验收标准
- [ ] 统一响应工具包可用
- [ ] 错误处理框架可用
- [ ] Ent Schema 规范配置完成
- [ ] CI/CD Linter 检查通过
- [ ] 测试框架就绪
- [ ] i18n 基础设施就绪
- [ ] l10n 基础设施就绪
- [ ] 所有代码通过 golangci-lint 检查

### Stories 总览

| Story | 描述 | 优先级 | 工期 | 状态 |
|-------|------|--------|------|------|
| Story 1 | 统一响应工具包 | P0 | 2 天 | Planning |
| Story 2 | 错误处理框架 | P0 | 2 天 | Planning |
| Story 3 | Ent Schema 规范配置 | P0 | 1 天 | Planning |
| Story 4 | CI/CD Linter 配置 | P0 | 1 天 | Planning |
| Story 5 | 测试框架工具包 | P1 | 2 天 | Planning |
| Story 6 | 重构现有 Handler | P1 | 1 天 | Planning |
| Story 7 | i18n 基础设施 | P1 | 2 天 | Planning |
| Story 8 | l10n 基础设施 | P1 | 2 天 | Planning |

**总工期**: 13 天（P0: 6 天，P1: 7 天）  
**依赖关系**: Story 2 依赖 Story 1，Story 6 依赖 Story 1-2，Story 8 依赖 Story 7

---

## Stories

### Story 1: 统一响应工具包

**优先级**: P0  
**工作量**: 2 天  
**负责人**: Backend Dev  
**关联规范**: [API 设计规范](../../standards/api-design.md#41-统一响应格式)

#### 用户故事
作为开发者，我希望有统一的响应工具包，以便快速实现标准化的 API 响应格式。

#### 验收标准
- [ ] 创建 `core/pkg/response` 包
- [ ] 实现 `Success()` 函数（成功响应）
- [ ] 实现 `Error()` 函数（错误响应）
- [ ] 实现 `List()` 函数（列表响应含分页）
- [ ] 编写单元测试（覆盖率 > 90%）
- [ ] 编写使用文档和示例

#### 实现任务
- [ ] 创建 `core/pkg/response/response.go`
- [ ] 定义响应结构体（Response、ErrorInfo、PaginationInfo）
- [ ] 实现 Success 函数
- [ ] 实现 Error 函数
- [ ] 实现 List 函数（含分页）
- [ ] 编写单元测试
- [ ] 编写 README.md（使用示例）
- [ ] 更新现有 Handler（config.go）使用新工具包

#### 技术细节
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

#### 测试用例
- 成功响应格式正确
- 错误响应包含完整错误信息
- 列表响应包含分页信息
- JSON 序列化正确

---

### Story 2: 错误处理框架

**优先级**: P0  
**工作量**: 2 天  
**负责人**: Backend Dev  
**关联规范**: [API 设计规范](../../standards/api-design.md#5-错误码规范)

#### 用户故事
作为开发者，我希望有标准化的错误处理框架，以便统一管理错误码和错误消息。

#### 验收标准
- [ ] 创建 `core/pkg/errors` 包
- [ ] 定义标准错误码（认证、权限、资源、验证、系统）
- [ ] 实现自定义错误类型
- [ ] 实现错误码映射 HTTP 状态码
- [ ] 编写单元测试（覆盖率 > 90%）
- [ ] 编写错误码文档

#### 实现任务
- [ ] 创建 `core/pkg/errors/errors.go`
- [ ] 创建 `core/pkg/errors/codes.go`
- [ ] 定义 AppError 结构体
- [ ] 实现错误构造函数（New, Wrap）
- [ ] 实现 HTTP 状态码映射
- [ ] 定义所有错误码常量
- [ ] 编写单元测试
- [ ] 编写错误码文档（README.md）

#### 技术细节
```go
// core/pkg/errors/errors.go

package errors

type AppError struct {
    Code       string                 // 错误码
    Message    string                 // 错误消息
    HTTPStatus int                    // HTTP 状态码
    Details    map[string]interface{} // 详细信息
    Err        error                  // 原始错误
}

func (e *AppError) Error() string {
    return e.Message
}

func New(code, message string, httpStatus int) *AppError {
    // 实现
}

func Wrap(err error, code, message string, httpStatus int) *AppError {
    // 实现
}
```

```go
// core/pkg/errors/codes.go

package errors

// 认证错误
const (
    ErrAuthInvalidToken   = "AUTH_INVALID_TOKEN"
    ErrAuthTokenExpired   = "AUTH_TOKEN_EXPIRED"
    ErrAuthUnauthorized   = "AUTH_UNAUTHORIZED"
)

// 权限错误
const (
    ErrPermForbidden        = "PERM_FORBIDDEN"
    ErrPermInsufficientRole = "PERM_INSUFFICIENT_ROLE"
)

// 资源错误
const (
    ErrResNotFound      = "RES_NOT_FOUND"
    ErrResAlreadyExists = "RES_ALREADY_EXISTS"
)

// ... 更多错误码
```

#### 测试用例
- AppError 正确创建和包装
- HTTP 状态码映射正确
- 错误信息包含完整上下文

---

### Story 3: Ent Schema 规范配置

**优先级**: P0  
**工作量**: 1 天  
**负责人**: Backend Dev  
**关联规范**: [编码规范 - Ent ORM](../../standards/coding-standards.md#12-ent-orm-规范)

#### 用户故事
作为开发者，我希望 Ent Schema 遵循统一规范，以便 API 响应字段格式一致。

#### 验收标准
- [ ] 现有 Ent Schema 添加 JSON tag（snake_case）
- [ ] 创建 Ent Schema 模板
- [ ] 编写 Ent Schema 检查脚本
- [ ] 检查脚本集成到开发流程
- [ ] 所有 Schema 通过规范检查

#### 实现任务
- [ ] 更新 `ent/schema/users.go`（添加 JSON tag）
- [ ] 更新 `ent/schema/servers.go`（添加 JSON tag）
- [ ] 更新 `ent/schema/configitem.go`（添加 JSON tag）
- [ ] 创建 `scripts/check-ent-json-tags.sh`
- [ ] 添加执行权限
- [ ] 在 Makefile 中添加 `ent-check` 目标
- [ ] 运行 `go generate ./ent` 重新生成代码
- [ ] 验证 API 响应字段格式

#### 技术细节
```go
// ent/schema/users.go 示例

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").
            StorageKey("id").
            StructTag(`json:"user_id"`),
        
        field.String("email").
            StorageKey("email").
            StructTag(`json:"email"`),
        
        field.Time("created_at").
            StorageKey("created_at").
            StructTag(`json:"created_at"`).
            Default(time.Now),
    }
}
```

#### 测试用例
- 检查脚本正确识别缺少 JSON tag 的字段
- 检查脚本正确识别 CamelCase 的 JSON tag
- 所有现有 Schema 通过检查

---

### Story 4: CI/CD Linter 检查配置

**优先级**: P0  
**工作量**: 1 天  
**负责人**: DevOps/Backend Dev  
**关联规范**: [编码规范 - 工具配置](../../standards/coding-standards.md#a-工具配置)

#### 用户故事
作为开发团队，我希望 CI/CD 自动检查代码规范，以便及早发现代码质量问题。

#### 验收标准
- [ ] golangci-lint 配置完成
- [ ] GitHub Actions CI 配置完成
- [ ] Ent Schema 检查集成到 CI
- [ ] PR 自动触发检查
- [ ] 所有检查通过

#### 实现任务
- [ ] 创建 `.golangci.yml` 配置文件
- [ ] 创建 `.github/workflows/ci.yml`
- [ ] 配置 golangci-lint job
- [ ] 配置 ent-check job
- [ ] 配置单元测试 job
- [ ] 配置代码覆盖率上传
- [ ] 在 README 中添加 CI 状态徽章
- [ ] 测试 CI 流程

#### 技术细节
```yaml
# .github/workflows/ci.yml

name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --config=.golangci.yml
    
    - name: Check Ent Schema JSON tags
      run: |
        chmod +x scripts/check-ent-json-tags.sh
        ./scripts/check-ent-json-tags.sh

  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

#### 测试用例
- Push 到 main/develop 触发 CI
- PR 创建触发 CI
- Linter 检查失败时 CI 失败
- 所有测试通过时 CI 成功

---

### Story 5: 测试框架与工具包

**优先级**: P1  
**工作量**: 2 天  
**负责人**: Backend Dev  
**关联规范**: [编码规范 - 测试规范](../../standards/coding-standards.md#6-测试规范)

#### 用户故事
作为开发者，我希望有统一的测试工具包，以便快速编写高质量的测试用例。

#### 验收标准
- [ ] 创建 `pkg/testutil` 测试工具包
- [ ] 实现 Mock HTTP 请求辅助函数
- [ ] 实现数据库测试辅助函数（基于 Ent）
- [ ] 实现断言辅助函数
- [ ] 编写测试示例
- [ ] 编写测试指南文档

#### 实现任务
- [ ] 创建 `pkg/testutil` 包
- [ ] 实现 HTTP 测试辅助函数
- [ ] 实现数据库测试辅助函数
- [ ] 实现 Mock 工具
- [ ] 创建测试示例（example_test.go）
- [ ] 编写测试指南（docs/standards/testing-guide.md）
- [ ] 为现有代码添加示例测试

#### 技术细节
```go
// pkg/testutil/http.go

package testutil

import (
    "net/http"
    "net/http/httptest"
)

// NewRequest 创建测试请求
func NewRequest(method, path string, body interface{}) *http.Request {
    // 实现
}

// NewRecorder 创建响应记录器
func NewRecorder() *httptest.ResponseRecorder {
    return httptest.NewRecorder()
}

// AssertJSON 断言 JSON 响应
func AssertJSON(t *testing.T, w *httptest.ResponseRecorder, expected interface{}) {
    // 实现
}
```

```go
// pkg/testutil/db.go

package testutil

import (
    "context"
    "testing"
    "apprun/ent"
)

// SetupTestDB 创建测试数据库
func SetupTestDB(t *testing.T) *ent.Client {
    // 实现
}

// TeardownTestDB 清理测试数据库
func TeardownTestDB(t *testing.T, client *ent.Client) {
    // 实现
}
```

#### 测试用例
- HTTP 测试辅助函数正常工作
- 数据库测试辅助函数可创建和清理测试数据
- 示例测试通过

---

### Story 6: 更新现有代码使用新工具

**优先级**: P1  
**工作量**: 1 天  
**负责人**: Backend Dev

#### 用户故事
作为开发者，我希望现有代码使用新的工具包，以便验证工具包的可用性。

#### 验收标准
- [ ] `core/handlers/config.go` 使用 response 包
- [ ] 错误处理使用 errors 包
- [ ] 所有 API 响应格式统一
- [ ] 现有测试通过
- [ ] 编写集成测试

#### 实现任务
- [ ] 重构 `core/handlers/config.go`
  - [ ] 使用 `response.Success()`
  - [ ] 使用 `response.Error()`
  - [ ] 使用 `errors` 包定义错误
- [ ] 更新 `core/routes/router.go`
  - [ ] 健康检查使用 response 包
- [ ] 编写集成测试
  - [ ] 测试 GET /api/config
  - [ ] 测试 PUT /api/config
  - [ ] 测试 GET /api/config/{key}
- [ ] 运行所有测试确保通过

#### 测试用例
- 配置 API 响应格式符合规范
- 错误响应包含完整错误信息
- 集成测试通过

---

### Story 7: i18n 基础设施

**优先级**: P1  
**工作量**: 2 天  
**负责人**: Backend Dev  
**关联规范**: [i18n 规范](../../standards/i18n-standards.md)

#### 用户故事
作为开发者，我希望有国际化（i18n）基础设施，以便支持多语言用户。

#### 验收标准
- [ ] 创建 `core/pkg/i18n` 包
- [ ] 集成 go-i18n v2 库
- [ ] 实现语言检测中间件
- [ ] 创建英文和中文消息文件
- [ ] 编写单元测试（覆盖率 > 80%）
- [ ] 编写使用文档

#### 实现任务
- [ ] 安装 go-i18n 依赖
- [ ] 创建 `core/pkg/i18n/i18n.go`（初始化）
- [ ] 创建 `core/pkg/i18n/middleware.go`（Chi 中间件）
- [ ] 创建消息文件目录 `locales/`
- [ ] 创建 `locales/en.yaml`（英文）
- [ ] 创建 `locales/zh-CN.yaml`（中文）
- [ ] 实现 `FromContext()` 辅助函数
- [ ] 编写单元测试
- [ ] 编写 README.md（使用示例）
- [ ] 更新 Router 集成中间件

#### 技术细节
```go
// core/pkg/i18n/i18n.go

package i18n

import (
    "embed"
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "golang.org/x/text/language"
    "gopkg.in/yaml.v3"
)

//go:embed ../../locales/*.yaml
var localeFS embed.FS

var Bundle *i18n.Bundle

func Init() error {
    Bundle = i18n.NewBundle(language.English)
    Bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
    
    // 加载语言文件
    languages := []string{"en", "zh-CN"}
    for _, lang := range languages {
        _, err := Bundle.LoadMessageFileFS(localeFS, 
            fmt.Sprintf("locales/%s.yaml", lang))
        if err != nil {
            return err
        }
    }
    
    return nil
}

func FromContext(ctx context.Context) *i18n.Localizer {
    lang := ctx.Value("accept-language").(string)
    return i18n.NewLocalizer(Bundle, lang)
}
```

```go
// core/pkg/i18n/middleware.go

package i18n

import (
    "context"
    "net/http"
)

func Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 检测语言
        lang := detectLanguage(r)
        
        // 存入上下文
        ctx := context.WithValue(r.Context(), "accept-language", lang)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func detectLanguage(r *http.Request) string {
    // 1. URL 参数
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
```

```yaml
# locales/en.yaml
user_not_found: "User not found"
invalid_email: "Invalid email format"
project_created: "Project created successfully"
welcome_user: "Welcome, {{.Name}}!"

# locales/zh-CN.yaml
user_not_found: "用户不存在"
invalid_email: "邮箱格式不正确"
project_created: "项目创建成功"
welcome_user: "欢迎，{{.Name}}！"
```

#### 测试用例
- 英文消息加载正确
- 中文消息加载正确
- 语言检测从 URL 参数
- 语言检测从 Accept-Language Header
- Fallback 到英文
- 变量替换正常工作
- 中间件正确设置上下文

---

## Story 8: 本地化（l10n）基础设施

**优先级**: P1  
**工期**: 2 天  
**依赖**: Story 7（i18n 基础设施）

### 目标
建立本地化基础设施，支持货币、日期、数字的区域化格式化，与 i18n 松耦合协作。

### 任务清单
- [ ] 创建 `core/pkg/localization` 包
  - [ ] `localization.go` - 主 Localizer
  - [ ] `currency.go` - 货币格式化
  - [ ] `datetime.go` - 日期时间格式化
  - [ ] `number.go` - 数字格式化
  - [ ] `units.go` - 度量单位转换
  - [ ] `config.go` - 配置加载
- [ ] 创建 `config/localization.yaml` 配置文件
- [ ] 创建中间件 `core/internal/middleware/localization.go`
- [ ] 编写单元测试（覆盖率 > 80%）
- [ ] 集成测试（验证 API 响应格式化）
- [ ] 更新 `docs/standards/localization-standards.md`（如需补充）

### 验收标准
1. **货币格式化**
   - 支持 USD、CNY、JPY、EUR、GBP
   - 正确显示货币符号位置（前缀 vs 后缀）
   - 千分位和小数点符合区域规则
   
2. **日期时间格式化**
   - 支持 5+ 种区域的日期格式
   - 支持 12/24 小时制切换
   - 时区转换正确
   
3. **数字格式化**
   - 千分位分隔符正确（逗号、点、空格）
   - 小数点符号正确
   - 百分比格式化
   
4. **度量单位转换**
   - 支持长度单位（米、千米、英里）
   - 支持重量单位（克、千克、磅）
   - 文件大小格式化（B、KB、MB、GB）
   
5. **架构要求**
   - 与 i18n 共享语言检测
   - 独立的 Localizer 上下文
   - 缓存机制（避免重复创建 Localizer）

### 代码示例

#### Localizer 主入口

```go
// core/pkg/localization/localization.go

package localization

import (
    "context"
    "time"
    "golang.org/x/text/language"
)

type Localizer struct {
    locale            string
    tag               language.Tag
    currencyFormatter *CurrencyFormatter
    dateTimeFormatter *DateTimeFormatter
    numberFormatter   *NumberFormatter
    unitConverter     *UnitConverter
}

func NewLocalizer(locale string) *Localizer {
    tag := language.MustParse(locale)
    
    return &Localizer{
        locale:            locale,
        tag:               tag,
        currencyFormatter: NewCurrencyFormatter(locale, getDefaultCurrency(locale)),
        dateTimeFormatter: NewDateTimeFormatter(locale),
        numberFormatter:   NewNumberFormatter(locale),
        unitConverter:     NewUnitConverter(locale),
    }
}

func FromContext(ctx context.Context) *Localizer {
    locale, ok := ctx.Value("locale").(string)
    if !ok {
        locale = "en-US"
    }
    
    return NewLocalizer(locale)
}

func (l *Localizer) FormatCurrency(amount float64, currency string) string {
    formatter := NewCurrencyFormatter(l.locale, currency)
    return formatter.FormatWithSymbol(amount)
}

func (l *Localizer) FormatDate(t time.Time) string {
    return l.dateTimeFormatter.FormatDate(t)
}

func (l *Localizer) FormatDateTime(t time.Time) string {
    return l.dateTimeFormatter.FormatDateTime(t)
}

func (l *Localizer) FormatNumber(n float64) string {
    return l.numberFormatter.FormatDecimal(n, 2)
}

func (l *Localizer) FormatBytes(bytes int64) string {
    return l.unitConverter.FormatBytes(bytes)
}
```

#### 货币格式化

```go
// core/pkg/localization/currency.go

package localization

import (
    "golang.org/x/text/currency"
    "golang.org/x/text/language"
    "golang.org/x/text/message"
)

type CurrencyFormatter struct {
    locale   language.Tag
    currency currency.Unit
    printer  *message.Printer
}

func NewCurrencyFormatter(locale, currencyCode string) *CurrencyFormatter {
    tag := language.MustParse(locale)
    curr := currency.MustParseISO(currencyCode)
    
    return &CurrencyFormatter{
        locale:   tag,
        currency: curr,
        printer:  message.NewPrinter(tag),
    }
}

func (f *CurrencyFormatter) FormatWithSymbol(amount float64) string {
    symbol := f.getCurrencySymbol()
    formatted := f.printer.Sprintf("%.2f", amount)
    
    if f.isSymbolPrefix() {
        return fmt.Sprintf("%s%s", symbol, formatted)
    }
    return fmt.Sprintf("%s %s", formatted, symbol)
}

func (f *CurrencyFormatter) getCurrencySymbol() string {
    symbols := map[string]string{
        "USD": "$",
        "CNY": "¥",
        "JPY": "¥",
        "EUR": "€",
        "GBP": "£",
    }
    return symbols[f.currency.String()]
}
```

#### 中间件

```go
// core/internal/middleware/localization.go

package middleware

import (
    "context"
    "net/http"
    "apprun/core/pkg/i18n"
)

func LocalizationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. 检测语言（复用 i18n 逻辑）
        lang := i18n.DetectLanguage(r)
        
        // 2. 映射到 Locale
        locale := mapLanguageToLocale(lang)
        
        // 3. 检查用户偏好（如果已登录）
        if user := getUserFromContext(r.Context()); user != nil {
            if user.PreferredLocale != "" {
                locale = user.PreferredLocale
            }
        }
        
        // 4. 存入上下文
        ctx := context.WithValue(r.Context(), "locale", locale)
        ctx = context.WithValue(ctx, "accept-language", lang) // i18n 使用
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func mapLanguageToLocale(lang string) string {
    localeMap := map[string]string{
        "en":    "en-US",
        "zh-CN": "zh-CN",
        "zh-TW": "zh-TW",
        "ja":    "ja-JP",
    }
    
    if locale, ok := localeMap[lang]; ok {
        return locale
    }
    
    return "en-US"
}
```

#### 配置文件

```yaml
# config/localization.yaml

localization:
  default_locale: en-US
  
  locales:
    en-US:
      currency: USD
      date_format: "01/02/2006"
      time_format: "3:04 PM"
      timezone: "America/New_York"
      
    zh-CN:
      currency: CNY
      date_format: "2006-01-02"
      time_format: "15:04"
      timezone: "Asia/Shanghai"
      
    ja-JP:
      currency: JPY
      date_format: "2006/01/02"
      time_format: "15:04"
      timezone: "Asia/Tokyo"
  
  currencies:
    USD:
      symbol: "$"
      decimal_places: 2
      symbol_prefix: true
      
    CNY:
      symbol: "¥"
      decimal_places: 2
      symbol_prefix: true
```

#### 使用示例

```go
// Handler 中使用

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
    // i18n: 消息翻译
    i18nLocalizer := i18n.FromContext(r.Context())
    message := i18nLocalizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "product_detail",
    })
    
    // l10n: 数据格式化
    l10nLocalizer := localization.FromContext(r.Context())
    
    product := h.getProduct(productID)
    
    response.Success(w, map[string]interface{}{
        "message":    message,                                      // i18n
        "name":       product.Name,
        "price":      l10nLocalizer.FormatCurrency(product.Price, "USD"), // l10n
        "created_at": l10nLocalizer.FormatDate(product.CreatedAt),        // l10n
        "size":       l10nLocalizer.FormatBytes(product.Size),            // l10n
    })
}
```

#### 测试用例
- 货币格式化（USD、CNY、JPY、EUR）
- 日期格式化（5+ 种区域）
- 日期时间格式化（12/24 小时制）
- 数字格式化（千分位、小数点）
- 文件大小格式化（B、KB、MB、GB）
- Locale 检测（URL 参数、用户偏好、语言映射）
- 与 i18n 协作（共享语言检测，独立上下文）
- 缓存机制（避免重复创建 Localizer）

---

## Sprint 依赖

### 外部依赖
- GitHub Actions (CI/CD)
- Go 1.21+
- golangci-lint
- go-i18n v2
- golang.org/x/text

### 工具依赖
- Ent ORM
- Testify (测试框架)
- httptest (HTTP 测试)
- go-i18n (国际化)
- golang.org/x/text (本地化)

---

## Sprint 风险

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| CI/CD 配置复杂 | 中 | 使用标准 GitHub Actions，参考最佳实践 |
| Ent 代码重新生成问题 | 低 | 先备份现有代码，使用版本控制 |
| 现有代码重构工作量 | 中 | 优先重构核心 Handler，其他逐步迁移 |

---

## Sprint 监控指标

- [ ] 代码覆盖率 > 80%
- [ ] golangci-lint 零告警
- [ ] CI 构建时间 < 5 分钟
- [ ] 所有 PR 检查通过率 100%

---

## Sprint 交付物

1. **代码**
   - `core/pkg/response` 包（含测试）
   - `core/pkg/errors` 包（含测试）
   - `core/pkg/i18n` 包（含测试）
   - `core/pkg/localization` 包（含测试）
   - `pkg/testutil` 包（含示例）
   - 更新后的 Ent Schema
   - 更新后的 Handler 代码

2. **配置**
   - `.golangci.yml`
   - `.github/workflows/ci.yml`
   - `scripts/check-ent-json-tags.sh`
   - `config/localization.yaml`
   - 更新后的 Makefile

3. **国际化/本地化资源**
   - `locales/en.yaml` (英文消息)
   - `locales/zh-CN.yaml` (简体中文消息)
   - `locales/zh-TW.yaml` (繁体中文消息)
   - `locales/ja.yaml` (日文消息)

4. **文档**
   - `core/pkg/response/README.md`
   - `core/pkg/errors/README.md`
   - `core/pkg/i18n/README.md`
   - `core/pkg/localization/README.md`
   - `docs/standards/testing-guide.md`（可选）

---

## Sprint 回顾准备

### 需要讨论的问题
- 工具包 API 设计是否合理？
- CI/CD 流程是否满足需求？
- 测试框架是否易用？
- i18n/l10n 架构设计是否满足业务需求？
- 是否需要调整开发流程？

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26
