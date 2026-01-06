# Story 9: Database Field i18n 数据库字段国际化
# Sprint 0: Infrastructure 建设

**Priority**: P2  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: Story 8 (i18n 基础设施)  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [i18n Standards](../../standards/i18n-standards.md)

---

## User Story

作为开发者，我希望为 ent 生成的数据库模型字段（如 Description、Title）提供多语言支持，以便用户根据语言偏好查看本地化内容，无需修改主表 schema。

---

## 核心问题

数据库字段国际化需要解决：

### 1. 灵活扩展
- 添加新语言无需修改主表结构
- 支持任意字段的国际化

### 2. 查询效率
- 获取翻译内容时性能优化
- 支持批量查询避免 N+1 问题

### 3. 数据完整性
- 主表删除时级联删除翻译
- 必填语言的数据校验

---

## Acceptance Criteria

- [ ] 创建通用 `Translation` schema 支持任意实体翻译
- [ ] 实现翻译 CRUD API（Create/Read/Update/Delete）
- [ ] 集成现有 i18n 中间件，自动返回用户语言版本
- [ ] 提供批量翻译查询方法（避免 N+1）
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 编写使用文档 `docs/standards/db-i18n.md`

---

## Technical Design

### Schema 设计

```go
// ent/schema/translation.go
type Translation struct {
    ent.Schema
}

func (Translation) Fields() []ent.Field {
    return []ent.Field{
        field.String("entity_type").Comment("实体类型：product, category"),
        field.Int("entity_id").Comment("实体ID"),
        field.String("field_name").Comment("字段名：description, title"),
        field.String("language").Comment("语言代码：en-US, zh-CN"),
        field.Text("value").Comment("翻译内容"),
    }
}

func (Translation) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("entity_type", "entity_id", "field_name", "language").Unique(),
    }
}
```

---

## Implementation Tasks

### Phase 1: Schema & API（Day 1）
- [ ] 创建 `ent/schema/translation.go`
- [ ] 运行 `go generate` 生成代码
- [ ] 创建 `modules/translation/service.go`（CRUD 逻辑）
- [ ] 创建 `modules/translation/handler.go`（HTTP API）

### Phase 2: Integration & Helpers（Day 2）
- [ ] 实现 `GetTranslation(ctx, entityType, entityID, fieldName)` 辅助函数
- [ ] 实现批量查询 `GetTranslations(ctx, entityType, entityIDs, fieldNames)`
- [ ] 编写单元测试（repository、service、handler）
- [ ] 编写文档 `docs/standards/db-i18n.md`

---

## API Design

### 创建/更新翻译
```http
PUT /api/translations
Content-Type: application/json

{
  "entity_type": "product",
  "entity_id": 123,
  "field_name": "description",
  "translations": {
    "en-US": "High quality product",
    "zh-CN": "高质量产品"
  }
}
```

### 获取翻译（自动根据用户语言）
```http
GET /api/products/123
Accept-Language: zh-CN

{
  "id": 123,
  "name": "产品A",
  "description": "高质量产品"  // 自动返回中文
}
```

---

## Example Usage

```go
// 创建产品时设置翻译
product := client.Product.Create().
    SetName("Product A").
    SetDescription("Default description").
    SaveX(ctx)

// 添加翻译
translationService.Set(ctx, "product", product.ID, "description", map[string]string{
    "en-US": "High quality product",
    "zh-CN": "高质量产品",
})

// 查询时自动获取翻译
lang := i18n.GetLanguage(ctx)
desc := translationService.Get(ctx, "product", product.ID, "description", lang)
```

---

## Testing Strategy

- **Unit Tests**: 翻译 CRUD 逻辑
- **Integration Tests**: API 端点测试
- **Performance Tests**: 批量查询性能

---

## Success Metrics

- 翻译查询延迟 < 10ms
- 批量查询支持 100+ 实体
- 测试覆盖率 ≥ 80%

---

## Future Enhancements

- 支持翻译版本历史
- 集成翻译管理后台
- AI 自动翻译建议
