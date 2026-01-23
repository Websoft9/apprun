# Story 2.3: i18n Integration Plan - i18n 集成计划
# Sprint 0: Infrastructure 建设

**Priority**: P2  
**Effort**: 0.5 天（规划）  
**Owner**: Architect + Dev  
**Dependencies**: Story 8 (i18n 基础设施)  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [Story 8: i18n](./story-08-i18n.md)

---

## User Story

作为架构师和开发者，我需要明确哪些现有和未来的 Story 需要集成 i18n，以便制定清晰的集成路线图，确保所有用户可见消息都支持多语言。

---

## Scope Definition

**本计划专注于后端 API 的 i18n**，覆盖以下类型：
- ✅ **错误消息**（验证、权限、系统错误）
- ✅ **成功消息**（操作确认）
- ✅ **验证消息**（字段级别）
- 🔄 **日期时间格式**（l10n，Story 09）
- 🔄 **数字货币格式**（l10n，Story 09）
- ❌ **前端 UI 标签**（前端独立实现，不在此计划范围）
- ❌ **邮件通知内容**（暂不考虑，待未来 Epic）
- ❌ **日志消息**（保持英文）

---

## Integration Checklist

### 🔥 **高优先级（必须集成）**

#### 1. Story 02: Response Package ✅ **已完成**
- **状态**: Done
- **集成点**: 
  - `response.Error()` - 错误响应消息
  - `response.ValidationError()` - 验证错误消息
  - `response.AppError()` - 应用错误消息
- **工作量**: 中等（10+ 调用点）
- **集成方式**:
  ```go
  msg := i18n.TranslateContext(r.Context(), "errors.not_found", nil)
  response.Error(w, r, errors.ErrCodeNotFound, msg)
  ```

#### 2. Story 03: Error Handling ✅ **已完成**
- **状态**: Done
- **集成点**: 
  - 114+ 预定义错误码的消息翻译
  - `errors.New()` 的 message 参数
- **工作量**: 大（需翻译 114 个错误码）
- **集成方式**:
  ```go
  errors.New(
      errors.VAL_CONFIG_INVALID_PARAM_001,
      i18n.Translate(lang, "errors.config.invalid_param", nil),
  )
  ```

#### 3. Story 13: Request Package 📋 **计划中**
- **状态**: Planning (Sprint 1)
- **集成点**:
  - `request.ParseAndValidate()` 验证错误
  - 字段级别验证消息
- **工作量**: 小
- **集成方式**:
  ```go
  // 验证失败时返回本地化消息
  response.ValidationError(w, r, i18n.TranslateContext(ctx, "validation.required", nil))
  ```

---

### 🟡 **中优先级（建议集成）**

#### 4. Story 09: l10n (Localization) 📋 **计划中**
- **状态**: Planning
- **依赖**: 直接依赖 Story 8
- **集成点**:
  - 日期时间格式化
  - 数字货币格式化
  - 时区转换
- **工作量**: 大（新功能模块）

#### 5. Story 10: Config Module ✅ **已完成**
- **状态**: Done
- **集成点**:
  - 配置验证错误消息
  - 配置项描述（可选）
- **工作量**: 小
- **集成方式**: 配置验证失败时使用 i18n 翻译错误

#### 6. Story 11: Swagger Docs ✅ **已完成**
- **状态**: Done
- **集成点**:
  - API 描述文本（可选）
  - 错误响应示例
- **工作量**: 小（可选）
- **备注**: Swagger 描述可保持英文，仅错误响应需要 i18n

---

### ⚪ **低优先级（可选集成）**

#### 7. Story 04: Ent Schema ✅ **已完成**
- **状态**: Done
- **集成点**: 数据库字段验证错误（Hook 层）
- **工作量**: 小
- **备注**: 通常在业务层处理，Schema 层保持英文

#### 8. Story 12: Logger Package ✅ **已完成**
- **状态**: Done
- **集成点**: 无需集成
- **备注**: 日志消息建议保持英文，便于运维和搜索

---

## Implementation Roadmap

### Phase 1: 核心错误消息国际化（当前 Sprint）
- [x] Story 8: i18n 基础设施
- [ ] 为 Story 02/03 创建基础翻译文件
- [ ] 翻译 114 个错误码（en-US, zh-CN）
- [ ] 更新 Response/Errors 包集成示例

### Phase 2: 业务模块集成（Sprint 1）
- [ ] Story 13: Request 包验证消息
- [ ] Story 10: Config 模块错误消息
- [ ] 创建模块级翻译文件组织规范

### Phase 3: 本地化功能（Sprint 1/2）
- [ ] Story 09: l10n 完整实现
- [ ] 日期时间格式化
- [ ] 数字货币格式化

---

## Translation File Structure

### 推荐组织方式

```
core/locales/
├── active.en-US.toml          # 英文主文件
├── active.zh-CN.toml          # 中文主文件
├── template.toml              # 自动生成的模板（开发时）
└── README.md                  # 翻译维护指南
```

### 翻译键命名规范

```toml
# 通用消息
[common.welcome]
other = "Welcome"

[common.success]
other = "Operation successful"

# 错误消息（按模块分组）
[errors.validation.required]
other = "Field is required"

[errors.config.invalid_param]
other = "Invalid configuration parameter"

[errors.auth.unauthorized]
other = "Unauthorized access"

# 成功消息
[success.created]
other = "Resource created successfully"

[success.updated]
other = "Resource updated successfully"
```

---

## Developer Guidelines

### 1. 何时使用 i18n
- ✅ **API 响应消息**（错误、成功、验证）
- ✅ **用户可见的所有文本**（提示、帮助）
- 🔄 **日期时间格式**（需结合 l10n）
- 🔄 **数字货币格式**（需结合 l10n）
- ❌ **日志消息**（保持英文）
- ❌ **调试信息**（保持英文）
- ❌ **配置键名**（代码常量）
- ❌ **数据库字段名**（内部使用）

### 2. 使用示例

**在 Handler 中**:
```go
func CreateUser(w http.ResponseWriter, r *http.Request) {
    user, err := parseUser(r)
    if err != nil {
        msg := i18n.TranslateContext(r.Context(), "errors.validation.invalid_input", nil)
        response.ValidationError(w, r, msg)
        return
    }
    
    msg := i18n.TranslateContext(r.Context(), "success.user.created", nil)
    response.Success(w, map[string]interface{}{
        "message": msg,
        "user": user,
    })
}
```

**在 Service 层**:
```go
func (s *UserService) Create(ctx context.Context, user *User) error {
    if err := s.validate(user); err != nil {
        lang := i18n.GetLanguage(ctx)
        msg := i18n.Translate(lang, "errors.validation.email_exists", nil)
        return errors.New(errors.VAL_USER_EMAIL_EXISTS, msg)
    }
    // ...
}
```

### 3. 添加新翻译键流程
1. 在代码中使用 `i18n.TranslateContext(ctx, "new.key", nil)`
2. 运行 `make i18n` 自动提取键
3. 在 `active.en-US.toml` 和 `active.zh-CN.toml` 中填写翻译
4. 提交代码和翻译文件

---

## Success Metrics

- [ ] 所有 API 错误响应支持 zh-CN 和 en-US
- [ ] 所有成功消息支持多语言
- [ ] 114 个错误码全部翻译
- [ ] 验证消息支持字段级别翻译
- [ ] 新功能开发默认支持 i18n
- [ ] 翻译覆盖率 100%（通过 `make i18n` 验证）

---

## Future Considerations (待评估)

以下内容暂不在当前 Story 范围，待未来评估：
- **前端 UI 国际化**：如有 Web/移动端界面，需单独规划
- **邮件通知国际化**：密码重置、系统通知等邮件模板
- **帮助文档国际化**：用户手册、FAQ 等
- **法律合规文本**：隐私政策、服务条款（需法律审核）
- **多租户语言偏好**：每个租户/用户的语言设置存储

---

## Related Documentation

- [Story 8: i18n 基础设施](./story-08-i18n.md)
- [Story 9: l10n 本地化](./story-09-l10n.md)
- [API 设计规范 - i18n](../../standards/api-design.md#i18n)
- [i18n Package README](../../../core/pkg/i18n/README.md)

---

**Created**: 2026-01-06  
**Maintainer**: Winston (Architect Agent)  
**Status**: Active Planning Document
