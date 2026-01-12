# Epic 2: 国际化基础设施 (i18n)
# apprun BaaS Platform

**Epic ID**: epic-2  
**关联 PRD**: 基础设施需求  
**负责人**: Architect Agent  
**状态**: Planning  
**优先级**: P1 (基础设施)  
**预估工作量**: 1.5-2 周

---

## 1. Epic 概述

### 1.1 业务目标

提供独立的国际化（i18n）基础设施，支持多语言响应消息，提升全球用户体验。

### 1.2 核心价值

- 用户可使用母语查看系统消息
- 翻译文本与代码分离，便于维护
- 轻松添加新语言支持
- 为后续本地化（l10n）奠定基础

### 1.3 验收标准

- [ ] 支持中文（zh-CN）和英文（en-US）
- [ ] 用户可通过 Query/Header/Cookie 切换语言
- [ ] 翻译缺失时系统不崩溃（fallback 机制）
- [ ] 翻译函数调用延迟 < 1ms (P99)
- [ ] 单元测试覆盖率 ≥ 70%

---

## 2. 技术规范

### 2.1 架构设计

```
HTTP Request → 语言检测中间件 → Handler
                      ↓                ↓
               Context (lang)    pkg/i18n (翻译)
                                      ↓
                              翻译文件 (TOML)
```

**职责分离**：
- **中间件**：检测语言，注入 Context（不翻译）
- **pkg/i18n**：提供翻译服务（独立于 HTTP）
- **Handler**：调用翻译，返回本地化响应

### 2.2 核心组件

#### pkg/i18n 包
```go
// 初始化（启动时调用）
func Init(defaultLang string, supportedLangs []string) error

// 翻译（推荐使用）
func TranslateContext(ctx context.Context, messageID string, data map[string]interface{}) string

// 基础翻译
func Translate(lang, messageID string, data map[string]interface{}) string

// Context 辅助
func WithLanguage(ctx context.Context, lang string) context.Context
func GetLanguage(ctx context.Context) string
```

#### 中间件
```go
// 语言检测中间件
func LanguageDetector() func(next http.Handler) http.Handler
```

### 2.3 翻译文件格式

**locales/active.zh-CN.toml**
```toml
[errors.invalid_param]
other = "请求参数无效"

[errors.not_found]
other = "资源未找到"

[success.created]
other = "创建成功"
```

**locales/active.en-US.toml**（结构相同）

### 2.4 语言检测优先级

1. **URL Query**: `?lang=zh-CN` (最高优先级)
2. **Cookie**: `lang=zh-CN`
3. **Accept-Language Header**: `Accept-Language: zh-CN,en;q=0.9`
4. **系统默认**: `en-US` (最低优先级)

### 2.5 错误处理策略

| 场景 | 处理方式 |
|------|---------|
| 翻译文件缺失 | 警告日志 + 继续运行 |
| 翻译键不存在 | 返回 messageID 作为 fallback |
| 不支持的语言 | 降级到默认语言 |
| 初始化失败 | 程序退出 |

---

## 3. Stories 拆分

### Story 2.1: i18n 核心基础设施
**优先级**: P1  
**工作量**: 2-3 天  
**文件**: [2-1-i18n.md](../sprint-artifacts/sprint-0/2-1-i18n.md)

- [ ] 实现 `pkg/i18n` 包
- [ ] 集成 go-i18n v2
- [ ] 实现语言检测中间件
- [ ] 实现翻译函数
- [ ] 翻译文件管理（toml）
- [ ] 单元测试 + 集成测试

**交付**：
- i18n 包可用
- 中英文翻译示例
- API 文档

### Story 2.2: Database Field i18n 数据库字段国际化
**优先级**: P2  
**工作量**: 1 天  
**文件**: [2-2-db-field-i18n.md](../sprint-artifacts/sprint-0/2-2-db-field-i18n.md)

### Story 2.3: i18n Integration Plan
**优先级**: P2  
**工作量**: 0.5 天  
**文件**: [2-3-i18n-integration-plan.md](../sprint-artifacts/sprint-0/2-3-i18n-integration-plan.md)

### Story 2.4: l10n 本地化支持
**优先级**: P2  
**工作量**: 1 天  
**文件**: [2-4-l10n.md](../sprint-artifacts/sprint-0/2-4-l10n.md)

- [ ] 时区处理
- [ ] 货币格式化
- [ ] 日期/时间格式化
- [ ] 数字格式化
- [ ] 单元测试

---

## 4. 依赖关系

### 技术依赖
- `github.com/nicksnyder/go-i18n/v2` - 国际化库
- `github.com/BurntSushi/toml` - TOML 解析
- `golang.org/x/text` - 标准库支持

### 模块依赖
- 无外部模块依赖（独立运行）

### 配置依赖
- 默认语言配置
- 支持的语言列表

---

## 5. 使用示例

### 初始化（main.go）
```go
func main() {
    // 初始化 i18n
    if err := i18n.Init("en-US", []string{"en-US", "zh-CN"}); err != nil {
        log.Fatal("i18n init failed:", err)
    }
    
    // 注册中间件
    r := chi.NewRouter()
    r.Use(middleware.LanguageDetector())
    
    // 注册 Handler
    r.Get("/api/demo", demoHandler)
    http.ListenAndServe(":8080", r)
}
```

### Handler 中使用
```go
func demoHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 自动获取用户语言并翻译
    msg := i18n.TranslateContext(ctx, "success.created", nil)
    
    w.Write([]byte(msg))
}
```

### 测试
```bash
# 中文
curl "http://localhost:8080/api/demo?lang=zh-CN"
# 响应: 创建成功

# 英文
curl "http://localhost:8080/api/demo?lang=en-US"
# 响应: Created successfully
```

---

## 6. 风险与挑战

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| 翻译文件缺失 | 低 | fallback 机制，系统继续运行 |
| 性能开销 | 低 | 启动时加载到内存，运行时无 I/O |
| 翻译质量 | 中 | 建立翻译审核流程 |

---

## 7. 测试策略

### 单元测试
- 翻译函数正确性
- Context 存取语言
- 语言检测优先级
- Fallback 机制

### 集成测试
- 端到端语言切换
- 中间件与 Handler 协作
- 并发请求语言隔离

### 性能测试
- 翻译函数调用 < 1ms (P99)
- 启动时间增加 < 100ms
- 并发翻译性能

---

## 8. 监控指标

- `i18n_translations_total` - 翻译调用次数
- `i18n_fallback_total` - fallback 次数
- `i18n_translate_duration_seconds` - 翻译耗时

---

## 附录

### A. 相关文档

- [Story 2.1: i18n 基础设施](../sprint-artifacts/sprint-0/2-1-i18n.md)
- [Story 2.2: DB Field i18n](../sprint-artifacts/sprint-0/2-2-db-field-i18n.md)
- [Story 2.3: i18n Integration Plan](../sprint-artifacts/sprint-0/2-3-i18n-integration-plan.md)
- [Story 2.4: Localization (l10n)](../sprint-artifacts/sprint-0/2-4-l10n.md)
- [go-i18n 官方文档](https://github.com/nicksnyder/go-i18n)

### B. 扩展计划

- **Phase 2**: 动态翻译加载（数据库/Redis）
- **Phase 3**: 翻译管理 Web UI
- **Phase 4**: 更多语言支持（日语、韩语等）

---

**文档维护**: Winston (Architect Agent)  
**创建日期**: 2026-01-05  
**最后更新**: 2026-01-05
