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
- [ ] 创建 `core/pkg/i18n/` 目录
- [ ] 创建 `core/pkg/middleware/` 目录

### Phase 2: 翻译文件 (45 分钟)
- [ ] 创建 `active.zh-CN.toml`（中文翻译）
- [ ] 创建 `active.en-US.toml`（英文翻译）
- [ ] 添加常用错误和成功消息

### Phase 3: pkg/i18n 核心包 (2 小时)
- [ ] 实现 `i18n.go`：Init、Translate、GetLanguage 等函数
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
│   ├── active.zh-CN.toml
│   └── active.en-US.toml
├── pkg/
│   ├── i18n/
│   │   ├── i18n.go
│   │   ├── context.go
│   │   └── i18n_test.go
│   └── middleware/
│       ├── language.go
│       └── language_test.go
└── main.go
```

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

// 获取支持的语言列表
func GetSupportedLanguages() []string

// 检查语言是否支持
func IsSupported(lang string) bool
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
- [ ] 翻译返回正确的本地化文本
- [ ] 翻译键不存在时返回 fallback
- [ ] 支持模板变量插值
- [ ] Context 函数正确存取语言

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
// 1. 初始化 i18n
i18n.Init("en-US", []string{"en-US", "zh-CN"})

// 2. 注册中间件
r.Use(middleware.LanguageDetector())
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
curl "http://localhost:8080/api/demo?lang=zh-CN"  # 中文
curl "http://localhost:8080/api/demo?lang=en-US"  # 英文
curl -H "Accept-Language: zh-CN" "http://localhost:8080/api/demo"  # Header
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
