# Story 9: l10n Localization 本地化实施
# Sprint 0: Infrastructure建设

**Priority**: P1  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: Story 8 (i18n 基础设施)  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [API 设计规范](../../standards/api-design.md#i18n), [i18n Standards](../../standards/i18n-standards.md)

---

## User Story

作为国际化应用开发者，我需要实现本地化基础设施，使应用能够根据用户的地区偏好，自动格式化日期时间、数字货币等数据，提供符合本地习惯的用户体验。

---

## 核心问题

本地化需要解决的具体问题：

### 1. 时区处理
- **问题**：用户分布在不同时区，时间显示混乱
- **需求**：
  - 数据库使用 UTC 统一存储
  - API 返回时自动转换为用户时区
  - 支持系统级默认时区配置
  - 支持用户级时区偏好设置

### 2. 日期时间格式
- **问题**：不同地区日期格式不同
- **需求**：
  - 美国：`MM/DD/YYYY 02:30 PM`
  - 欧洲：`DD/MM/YYYY 14:30`
  - 中国：`YYYY年MM月DD日 14:30`
  - ISO 8601：`2024-01-15T14:30:00+08:00`

### 3. 数字格式
- **问题**：千位分隔符和小数点表示不同
- **需求**：
  - 美国：`1,234,567.89`
  - 欧洲：`1.234.567,89`
  - 中国：`1,234,567.89`

### 4. 货币格式
- **问题**：货币符号位置和格式不同
- **需求**：
  - 美元：`$1,234.56`
  - 欧元：`1.234,56 €`
  - 人民币：`¥1,234.56` 或 `CNY 1,234.56`

---

## Acceptance Criteria

- [ ] 实现时区中间件，从请求头或用户配置读取时区
- [ ] 在 Config 表支持系统级时区配置（`app.timezone`）
- [ ] 在 Users 表添加用户级时区字段（`timezone`）
- [ ] 提供时间格式化工具函数（根据语言和时区）
- [ ] 提供数字格式化工具函数（千位分隔符、小数点）
- [ ] 提供货币格式化工具函数（货币符号、位置）
- [ ] API 响应时间使用 RFC 3339 格式包含时区信息
- [ ] 单元测试覆盖所有格式化函数
- [ ] 文档说明时区配置和使用方法

---

## Technical Design

### 架构分层

```
┌─────────────────────────────────────────┐
│ API Layer                               │
│  • 请求：检测用户时区/语言               │
│  • 响应：格式化数据                      │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ Middleware Layer                        │
│  • timezone.Middleware                  │
│  • language.Middleware (Story 8)        │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ Business Layer                          │
│  • 使用 UTC 时间                         │
│  • 调用 l10n 工具格式化                  │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ Data Layer                              │
│  • 数据库: TIMESTAMP WITH TIME ZONE     │
│  • 存储: UTC 时间                        │
└─────────────────────────────────────────┘
```

### 包结构

```
core/pkg/l10n/
├── timezone.go       # 时区工具
├── datetime.go       # 日期时间格式化
├── number.go         # 数字格式化
├── currency.go       # 货币格式化
└── l10n_test.go      # 单元测试

core/pkg/middleware/
└── timezone.go       # 时区中间件
```

### 配置结构

```yaml
# config/default.yaml
app:
  timezone: "Asia/Shanghai"  # 系统默认时区（IANA 格式）
```

```sql
-- 用户表增加时区字段
ALTER TABLE users ADD COLUMN timezone VARCHAR(50) DEFAULT 'UTC';
```

---

## Implementation Tasks

### Task 1: 系统级时区配置
- 在 `Config.App` 添加 `Timezone` 字段
- 验证器支持 IANA 时区名称
- 启动时加载系统时区

### Task 2: 用户级时区支持
- 用户表添加 `timezone` 字段
- 用户注册时使用系统默认时区
- 提供用户时区更新 API

### Task 3: 时区中间件
- 从请求头 `Accept-Timezone` 或 `X-Timezone` 读取
- 从用户配置读取时区偏好
- 存储到请求上下文 `context.Context`

### Task 4: 格式化工具函数
- `FormatDateTime(t time.Time, lang, tz string) string`
- `FormatNumber(n float64, lang string) string`
- `FormatCurrency(amount float64, currency, lang string) string`

### Task 5: API 响应标准化
- 时间字段使用 RFC 3339 格式
- 包含时区偏移信息
- 示例：`2024-01-15T14:30:00+08:00`

---

## API Examples

### 请求示例
```http
GET /api/users/123
Accept-Language: zh-CN
Accept-Timezone: Asia/Shanghai
```

### 响应示例
```json
{
  "id": 123,
  "name": "张三",
  "created_at": "2024-01-15T14:30:00+08:00",
  "balance": {
    "amount": 1234.56,
    "formatted": "¥1,234.56",
    "currency": "CNY"
  },
  "timezone": "Asia/Shanghai",
  "language": "zh-CN"
}
```

---

## Testing Strategy

### 单元测试
- 时区转换：UTC ↔ 用户时区
- 日期格式化：多语言、多格式
- 数字格式化：千位分隔符、小数点
- 货币格式化：符号位置、格式

### 集成测试
- 中间件：请求头时区提取
- API 响应：时间字段格式正确
- 用户配置：时区设置生效

### 测试用例
```
时区：UTC, Asia/Shanghai, America/New_York, Europe/London
语言：en-US, zh-CN, de-DE, fr-FR
货币：USD, CNY, EUR, JPY
```

---

## Dependencies

- **Story 8**: i18n 基础设施（语言检测中间件）
- **Config 模块**: 系统配置支持
- **User 模块**: 用户表结构

---

## Non-Goals

本 Story 不包含：
- ❌ 复杂的地区规则（如节假日、工作日）
- ❌ 地址格式化
- ❌ 电话号码格式化
- ❌ 度量单位转换（英里 vs 公里）

---

## Benefits

1. **用户体验**：时间、数字、货币自动本地化
2. **全球化支持**：轻松支持新地区
3. **数据一致性**：统一使用 UTC 存储
4. **开发效率**：工具函数可复用

---

## Related Documentation

- [i18n Standards](../../standards/i18n-standards.md)
- [API Design - i18n Guidelines](../../standards/api-design.md#i18n)
- [Story 8 - i18n Infrastructure](./story-08-i18n.md)

---

**Created**: 2024-12-01  
**Updated**: 2025-12-31