# 本地化 (l10n) 规范
# apprun BaaS Platform

**创建日期**: 2025-12-26  
**维护者**: Winston (Architect Agent)  
**版本**: 1.0  
**状态**: Active

---

## 概述

本文档定义 apprun 项目的本地化（Localization, l10n）规范，基于 `golang.org/x/text` 库实现，确保后端服务能够适应不同地区的文化习惯和数据格式。

**核心原则**：
- **独立但协作**：与 i18n 共享语言检测，但保持独立实现
- **数据为中心**：关注数据格式化，而非文本翻译
- **性能优先**：预编译格式化规则，避免运行时解析
- **可扩展性**：支持动态添加新的本地化规则

**与 i18n 的关系**：
- **i18n**：解决"软件能支持多语言"（消息翻译）
- **l10n**：解决"如何适应当地习惯"（数据格式化）

---

## 目录

1. [总体策略](#1-总体策略)
2. [货币和价格本地化](#2-货币和价格本地化)
3. [日期和时间本地化](#3-日期和时间本地化)
4. [数字和度量本地化](#4-数字和度量本地化)
5. [实现指南](#5-实现指南)
6. [测试规范](#6-测试规范)
7. [最佳实践](#7-最佳实践)

---

## 1. 总体策略

### 1.1 支持的区域 (Locale)

| Locale | 区域名称 | 优先级 | 货币 | 日期格式 | 状态 |
|--------|---------|-------|------|---------|------|
| `en-US` | 美国英语 | P0 | USD | MM/DD/YYYY | 必须 |
| `zh-CN` | 中国简体中文 | P0 | CNY | YYYY-MM-DD | 必须 |
| `zh-TW` | 中国繁体中文 | P1 | TWD | YYYY/MM/DD | 推荐 |
| `ja-JP` | 日本语 | P1 | JPY | YYYY年MM月DD日 | 推荐 |
| `en-GB` | 英国英语 | P2 | GBP | DD/MM/YYYY | 可选 |
| `de-DE` | 德语 | P2 | EUR | DD.MM.YYYY | 可选 |
| `fr-FR` | 法语 | P2 | EUR | DD/MM/YYYY | 可选 |

### 1.2 Locale 检测机制

#### **与 i18n 协作**

```go
// 共享语言检测中间件
func LocaleMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. 检测语言（复用 i18n 逻辑）
        lang := detectLanguage(r)
        
        // 2. 映射到 Locale
        locale := mapLanguageToLocale(lang)
        
        // 3. 存入上下文（独立 key，避免耦合）
        ctx := context.WithValue(r.Context(), "locale", locale)
        ctx = context.WithValue(ctx, "accept-language", lang) // i18n 使用
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 语言到 Locale 映射
func mapLanguageToLocale(lang string) string {
    localeMap := map[string]string{
        "en":    "en-US",
        "zh-CN": "zh-CN",
        "zh-TW": "zh-TW",
        "ja":    "ja-JP",
        "de":    "de-DE",
        "fr":    "fr-FR",
    }
    
    if locale, ok := localeMap[lang]; ok {
        return locale
    }
    
    return "en-US" // 默认
}
```

#### **优先级顺序**
1. **URL 参数**：`?locale=zh-CN`（覆盖语言检测）
2. **用户偏好**：数据库存储的用户本地化设置
3. **语言映射**：从 i18n 的语言检测结果映射
4. **默认 Locale**：`en-US`

### 1.3 与 i18n 的架构关系

```
HTTP 请求
    ↓
LocaleMiddleware (共享)
    ├── 检测语言 (detectLanguage)
    ├── 设置 i18n 上下文 (accept-language)
    └── 设置 l10n 上下文 (locale)
    ↓
业务逻辑
    ├── i18n.FromContext() → 消息翻译
    └── localization.FromContext() → 数据格式化
```

**耦合度**：低（仅共享语言检测，上下文独立）

---

## 2. 货币和价格本地化

### 2.1 货币格式化

#### **不同区域的货币格式**

| Locale | 货币 | 格式示例 |
|--------|------|---------|
| `en-US` | USD | $1,234.56 |
| `zh-CN` | CNY | ¥1,234.56 |
| `ja-JP` | JPY | ¥1,235 (无小数) |
| `en-GB` | GBP | £1,234.56 |
| `de-DE` | EUR | 1.234,56 € |
| `fr-FR` | EUR | 1 234,56 € |

#### **实现示例**

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

func (f *CurrencyFormatter) Format(amount float64) string {
    // 使用 golang.org/x/text/message 格式化
    return f.printer.Sprintf("%.2f", amount)
}

// 带货币符号
func (f *CurrencyFormatter) FormatWithSymbol(amount float64) string {
    symbol := f.getCurrencySymbol()
    formatted := f.Format(amount)
    
    // 根据 locale 决定符号位置
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

func (f *CurrencyFormatter) isSymbolPrefix() bool {
    // 美元、人民币等前置
    prefixLocales := []string{"en-US", "zh-CN", "ja-JP"}
    for _, loc := range prefixLocales {
        if f.locale.String() == loc {
            return true
        }
    }
    return false
}
```

#### **使用示例**

```go
// API Handler
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
    localizer := localization.FromContext(r.Context())
    
    product := h.getProduct(productID)
    
    // 格式化价格
    formattedPrice := localizer.FormatCurrency(product.Price, "USD")
    
    response.Success(w, map[string]interface{}{
        "id":    product.ID,
        "name":  product.Name,
        "price": formattedPrice, // "$1,234.56" or "¥1,234.56"
        "price_raw": product.Price, // 原始数值（供客户端使用）
    })
}
```

### 2.2 货币转换

#### **实时汇率支持**

```go
// core/pkg/localization/exchange.go

type ExchangeRateService struct {
    rates map[string]float64 // 相对 USD 的汇率
}

func (s *ExchangeRateService) Convert(amount float64, from, to string) float64 {
    // 1. 转换为 USD 基准
    usdAmount := amount / s.rates[from]
    
    // 2. 转换为目标货币
    return usdAmount * s.rates[to]
}

// 用户偏好货币
type UserCurrencyPreference struct {
    UserID   string
    Currency string // "USD", "CNY", "EUR"
}
```

---

## 3. 日期和时间本地化

### 3.1 日期格式化

#### **不同区域的日期格式**

| Locale | 短格式 | 长格式 | 时间格式 |
|--------|--------|--------|---------|
| `en-US` | 12/25/2025 | December 25, 2025 | 3:30 PM |
| `zh-CN` | 2025-12-25 | 2025年12月25日 | 15:30 |
| `ja-JP` | 2025/12/25 | 2025年12月25日 | 15:30 |
| `en-GB` | 25/12/2025 | 25 December 2025 | 15:30 |
| `de-DE` | 25.12.2025 | 25. Dezember 2025 | 15:30 |

#### **实现示例**

```go
// core/pkg/localization/datetime.go
package localization

import (
    "time"
    "golang.org/x/text/language"
)

type DateTimeFormatter struct {
    locale language.Tag
}

func NewDateTimeFormatter(locale string) *DateTimeFormatter {
    return &DateTimeFormatter{
        locale: language.MustParse(locale),
    }
}

func (f *DateTimeFormatter) FormatDate(t time.Time) string {
    formats := map[string]string{
        "en-US": "01/02/2006",
        "zh-CN": "2006-01-02",
        "ja-JP": "2006/01/02",
        "en-GB": "02/01/2006",
        "de-DE": "02.01.2006",
    }
    
    format := formats[f.locale.String()]
    if format == "" {
        format = "2006-01-02" // ISO 8601
    }
    
    return t.Format(format)
}

func (f *DateTimeFormatter) FormatDateTime(t time.Time) string {
    dateStr := f.FormatDate(t)
    timeStr := f.FormatTime(t)
    return fmt.Sprintf("%s %s", dateStr, timeStr)
}

func (f *DateTimeFormatter) FormatTime(t time.Time) string {
    // 12 小时制 vs 24 小时制
    if f.is12HourFormat() {
        return t.Format("3:04 PM")
    }
    return t.Format("15:04")
}

func (f *DateTimeFormatter) is12HourFormat() bool {
    twelveHourLocales := []string{"en-US"}
    for _, loc := range twelveHourLocales {
        if f.locale.String() == loc {
            return true
        }
    }
    return false
}

// 相对时间（"2 hours ago"）
func (f *DateTimeFormatter) FormatRelative(t time.Time) string {
    duration := time.Since(t)
    
    // 使用 i18n 翻译相对时间描述
    // 这里需要配合 i18n 包
    return f.formatDuration(duration)
}
```

### 3.2 时区处理

#### **时区转换**

```go
// core/pkg/localization/timezone.go

type TimezoneConverter struct {
    userTimezone *time.Location
}

func NewTimezoneConverter(timezone string) (*TimezoneConverter, error) {
    loc, err := time.LoadLocation(timezone)
    if err != nil {
        return nil, err
    }
    
    return &TimezoneConverter{
        userTimezone: loc,
    }, nil
}

func (c *TimezoneConverter) ConvertToUserTime(t time.Time) time.Time {
    return t.In(c.userTimezone)
}

// 用户偏好时区
type UserTimezonePreference struct {
    UserID   string
    Timezone string // "Asia/Shanghai", "America/New_York"
}
```

#### **使用示例**

```go
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    localizer := localization.FromContext(r.Context())
    
    user := h.getUser(userID)
    
    // 格式化注册时间（用户本地时间）
    formattedDate := localizer.FormatDateTime(user.CreatedAt)
    
    response.Success(w, map[string]interface{}{
        "id":         user.ID,
        "name":       user.Name,
        "created_at": formattedDate, // "12/25/2025 3:30 PM" or "2025-12-25 15:30"
        "created_at_iso": user.CreatedAt.Format(time.RFC3339), // 原始 ISO 格式
    })
}
```

---

## 4. 数字和度量本地化

### 4.1 数字格式化

#### **不同区域的数字格式**

| Locale | 千分位 | 小数点 | 示例 |
|--------|--------|--------|------|
| `en-US` | , | . | 1,234.56 |
| `zh-CN` | , | . | 1,234.56 |
| `de-DE` | . | , | 1.234,56 |
| `fr-FR` | ␣ (空格) | , | 1 234,56 |

#### **实现示例**

```go
// core/pkg/localization/number.go
package localization

import (
    "golang.org/x/text/language"
    "golang.org/x/text/message"
    "golang.org/x/text/number"
)

type NumberFormatter struct {
    printer *message.Printer
}

func NewNumberFormatter(locale string) *NumberFormatter {
    tag := language.MustParse(locale)
    return &NumberFormatter{
        printer: message.NewPrinter(tag),
    }
}

func (f *NumberFormatter) FormatInteger(n int64) string {
    return f.printer.Sprintf("%d", n)
}

func (f *NumberFormatter) FormatDecimal(n float64, precision int) string {
    format := fmt.Sprintf("%%.%df", precision)
    return f.printer.Sprintf(format, n)
}

func (f *NumberFormatter) FormatPercent(n float64) string {
    return f.printer.Sprintf("%.2f%%", n*100)
}
```

### 4.2 度量单位本地化

#### **长度单位**

```go
// core/pkg/localization/units.go

type UnitConverter struct {
    locale string
}

func (c *UnitConverter) FormatDistance(meters float64) string {
    // 美国使用英制单位
    if c.locale == "en-US" {
        miles := meters / 1609.34
        return fmt.Sprintf("%.2f mi", miles)
    }
    
    // 其他国家使用公制
    if meters >= 1000 {
        km := meters / 1000
        return fmt.Sprintf("%.2f km", km)
    }
    return fmt.Sprintf("%.0f m", meters)
}

func (c *UnitConverter) FormatWeight(grams float64) string {
    if c.locale == "en-US" {
        pounds := grams / 453.592
        return fmt.Sprintf("%.2f lb", pounds)
    }
    
    if grams >= 1000 {
        kg := grams / 1000
        return fmt.Sprintf("%.2f kg", kg)
    }
    return fmt.Sprintf("%.0f g", grams)
}

func (c *UnitConverter) FormatTemperature(celsius float64) string {
    if c.locale == "en-US" {
        fahrenheit := celsius*9/5 + 32
        return fmt.Sprintf("%.1f°F", fahrenheit)
    }
    return fmt.Sprintf("%.1f°C", celsius)
}
```

### 4.3 文件大小格式化

```go
func (c *UnitConverter) FormatBytes(bytes int64) string {
    units := []string{"B", "KB", "MB", "GB", "TB"}
    
    size := float64(bytes)
    unitIndex := 0
    
    for size >= 1024 && unitIndex < len(units)-1 {
        size /= 1024
        unitIndex++
    }
    
    return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}
```

---

## 5. 实现指南

### 5.1 技术选型

**推荐库**：
- `golang.org/x/text/language` - 语言标签和 Locale
- `golang.org/x/text/message` - 格式化和复数规则
- `golang.org/x/text/currency` - 货币处理
- `golang.org/x/text/number` - 数字格式化

### 5.2 项目结构

```
apprun/
├── config/
│   └── localization.yaml      # 本地化配置
├── core/
│   └── pkg/
│       ├── i18n/               # i18n 包（消息翻译）
│       └── localization/       # l10n 包（数据格式化）✨
│           ├── localization.go # 主入口和 Localizer
│           ├── currency.go     # 货币格式化
│           ├── datetime.go     # 日期时间格式化
│           ├── number.go       # 数字格式化
│           ├── units.go        # 度量单位转换
│           ├── config.go       # 配置加载
│           └── localization_test.go
```

### 5.3 核心代码实现

#### **5.3.1 主 Localizer**

```go
// core/pkg/localization/localization.go
package localization

import (
    "context"
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

// 从上下文获取 Localizer
func FromContext(ctx context.Context) *Localizer {
    locale, ok := ctx.Value("locale").(string)
    if !ok {
        locale = "en-US"
    }
    
    return NewLocalizer(locale)
}

// 货币格式化
func (l *Localizer) FormatCurrency(amount float64, currency string) string {
    formatter := NewCurrencyFormatter(l.locale, currency)
    return formatter.FormatWithSymbol(amount)
}

// 日期格式化
func (l *Localizer) FormatDate(t time.Time) string {
    return l.dateTimeFormatter.FormatDate(t)
}

// 日期时间格式化
func (l *Localizer) FormatDateTime(t time.Time) string {
    return l.dateTimeFormatter.FormatDateTime(t)
}

// 数字格式化
func (l *Localizer) FormatNumber(n float64) string {
    return l.numberFormatter.FormatDecimal(n, 2)
}

// 文件大小格式化
func (l *Localizer) FormatBytes(bytes int64) string {
    return l.unitConverter.FormatBytes(bytes)
}

func getDefaultCurrency(locale string) string {
    currencyMap := map[string]string{
        "en-US": "USD",
        "zh-CN": "CNY",
        "ja-JP": "JPY",
        "en-GB": "GBP",
        "de-DE": "EUR",
        "fr-FR": "EUR",
    }
    
    if curr, ok := currencyMap[locale]; ok {
        return curr
    }
    return "USD"
}
```

#### **5.3.2 配置管理**

```yaml
# config/localization.yaml
localization:
  default_locale: en-US
  
  # Locale 配置
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
  
  # 货币配置
  currencies:
    USD:
      symbol: "$"
      decimal_places: 2
      symbol_prefix: true
      
    CNY:
      symbol: "¥"
      decimal_places: 2
      symbol_prefix: true
      
    EUR:
      symbol: "€"
      decimal_places: 2
      symbol_prefix: false
```

```go
// core/pkg/localization/config.go
package localization

type Config struct {
    DefaultLocale string                 `yaml:"default_locale"`
    Locales       map[string]LocaleConfig `yaml:"locales"`
    Currencies    map[string]CurrencyConfig `yaml:"currencies"`
}

type LocaleConfig struct {
    Currency   string `yaml:"currency"`
    DateFormat string `yaml:"date_format"`
    TimeFormat string `yaml:"time_format"`
    Timezone   string `yaml:"timezone"`
}

type CurrencyConfig struct {
    Symbol        string `yaml:"symbol"`
    DecimalPlaces int    `yaml:"decimal_places"`
    SymbolPrefix  bool   `yaml:"symbol_prefix"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

#### **5.3.3 中间件集成**

```go
// core/internal/middleware/localization.go
package middleware

import (
    "context"
    "net/http"
    "apprun/core/pkg/i18n"
    "apprun/core/pkg/localization"
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
        "de":    "de-DE",
        "fr":    "fr-FR",
    }
    
    if locale, ok := localeMap[lang]; ok {
        return locale
    }
    
    return "en-US"
}
```

---

## 6. 测试规范

### 6.1 单元测试

```go
// core/pkg/localization/localization_test.go
package localization_test

import (
    "testing"
    "time"
    
    "apprun/core/pkg/localization"
    "github.com/stretchr/testify/assert"
)

func TestLocalizer_FormatCurrency(t *testing.T) {
    tests := []struct {
        locale   string
        amount   float64
        currency string
        expected string
    }{
        {"en-US", 1234.56, "USD", "$1,234.56"},
        {"zh-CN", 1234.56, "CNY", "¥1,234.56"},
        {"ja-JP", 1235, "JPY", "¥1,235"},
        {"de-DE", 1234.56, "EUR", "1.234,56 €"},
    }
    
    for _, tt := range tests {
        t.Run(tt.locale, func(t *testing.T) {
            localizer := localization.NewLocalizer(tt.locale)
            result := localizer.FormatCurrency(tt.amount, tt.currency)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestLocalizer_FormatDate(t *testing.T) {
    date := time.Date(2025, 12, 25, 15, 30, 0, 0, time.UTC)
    
    tests := []struct {
        locale   string
        expected string
    }{
        {"en-US", "12/25/2025"},
        {"zh-CN", "2025-12-25"},
        {"ja-JP", "2025/12/25"},
        {"en-GB", "25/12/2025"},
        {"de-DE", "25.12.2025"},
    }
    
    for _, tt := range tests {
        t.Run(tt.locale, func(t *testing.T) {
            localizer := localization.NewLocalizer(tt.locale)
            result := localizer.FormatDate(date)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestLocalizer_FormatBytes(t *testing.T) {
    tests := []struct {
        name     string
        bytes    int64
        expected string
    }{
        {"Bytes", 512, "512.00 B"},
        {"Kilobytes", 1536, "1.50 KB"},
        {"Megabytes", 1572864, "1.50 MB"},
        {"Gigabytes", 1610612736, "1.50 GB"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            localizer := localization.NewLocalizer("en-US")
            result := localizer.FormatBytes(tt.bytes)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 6.2 集成测试

```go
// tests/integration/localization_test.go
func TestAPI_Localization(t *testing.T) {
    tests := []struct {
        locale        string
        acceptLang    string
        expectedPrice string
        expectedDate  string
    }{
        {
            locale:        "en-US",
            acceptLang:    "en",
            expectedPrice: "$1,234.56",
            expectedDate:  "12/25/2025",
        },
        {
            locale:        "zh-CN",
            acceptLang:    "zh-CN",
            expectedPrice: "¥1,234.56",
            expectedDate:  "2025-12-25",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.locale, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/api/v1/products/1", nil)
            req.Header.Set("Accept-Language", tt.acceptLang)
            
            w := httptest.NewRecorder()
            app.ServeHTTP(w, req)
            
            var resp map[string]interface{}
            json.Unmarshal(w.Body.Bytes(), &resp)
            
            assert.Equal(t, tt.expectedPrice, resp["price"])
            assert.Equal(t, tt.expectedDate, resp["created_at"])
        })
    }
}
```

---

## 7. 最佳实践

### 7.1 与 i18n 协作

```go
// ✅ 推荐：独立但协作
func handler(w http.ResponseWriter, r *http.Request) {
    // i18n: 消息翻译
    i18nLocalizer := i18n.FromContext(r.Context())
    message := i18nLocalizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "welcome",
    })
    
    // l10n: 数据格式化
    l10nLocalizer := localization.FromContext(r.Context())
    price := l10nLocalizer.FormatCurrency(1234.56, "USD")
    date := l10nLocalizer.FormatDate(time.Now())
    
    response.Success(w, map[string]interface{}{
        "message":    message,  // i18n
        "price":      price,    // l10n
        "created_at": date,     // l10n
    })
}

// ❌ 避免：混合使用
func handler(w http.ResponseWriter, r *http.Request) {
    // 不要在 i18n 消息中混入格式化逻辑
    // ❌ message: "Price: $1,234.56"
}
```

### 7.2 缓存优化

```go
// ✅ 推荐：缓存 Localizer
var localizerCache sync.Map

func GetLocalizer(locale string) *Localizer {
    if cached, ok := localizerCache.Load(locale); ok {
        return cached.(*Localizer)
    }
    
    localizer := NewLocalizer(locale)
    localizerCache.Store(locale, localizer)
    return localizer
}
```

### 7.3 用户偏好存储

```go
// 用户本地化偏好
type UserLocalizationPreference struct {
    UserID   string `json:"user_id"`
    Locale   string `json:"locale"`     // "en-US", "zh-CN"
    Currency string `json:"currency"`   // "USD", "CNY"
    Timezone string `json:"timezone"`   // "Asia/Shanghai"
}

// 更新用户偏好
func (s *UserService) UpdateLocalizationPreference(
    ctx context.Context,
    userID string,
    pref *UserLocalizationPreference,
) error {
    // 保存到数据库
    return s.repo.UpdateLocalizationPreference(ctx, userID, pref)
}
```

---

## 附录

### A. Locale 参考

常用 Locale 代码：

| Locale | 语言 | 国家/地区 | 货币 |
|--------|------|----------|------|
| en-US | 英语 | 美国 | USD |
| en-GB | 英语 | 英国 | GBP |
| zh-CN | 中文 | 中国大陆 | CNY |
| zh-TW | 中文 | 中国台湾 | TWD |
| zh-HK | 中文 | 中国香港 | HKD |
| ja-JP | 日语 | 日本 | JPY |
| ko-KR | 韩语 | 韩国 | KRW |
| de-DE | 德语 | 德国 | EUR |
| fr-FR | 法语 | 法国 | EUR |
| es-ES | 西班牙语 | 西班牙 | EUR |

### B. 相关资源

- **golang.org/x/text**: https://pkg.go.dev/golang.org/x/text
- **CLDR (Unicode Common Locale Data Repository)**: https://cldr.unicode.org/
- **ISO 4217 (货币代码)**: https://www.iso.org/iso-4217-currency-codes.html
- **ISO 8601 (日期时间)**: https://www.iso.org/iso-8601-date-and-time-format.html
- **IANA Time Zone Database**: https://www.iana.org/time-zones

### C. 相关文档

- [i18n 规范](./i18n-standards.md) - 国际化（消息翻译）
- [API 设计规范](./api-design.md) - API 响应格式
- [编码规范](./coding-standards.md) - Go 代码规范

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26  
**审核状态**: Active
