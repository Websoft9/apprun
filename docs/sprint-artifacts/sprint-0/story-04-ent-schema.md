# Story 4: Ent Schema 配置管理
# Sprint 0: Infrastructure建设

**Priority**: P0  
**Effort**: 1 天  
**Owner**: Backend Dev  
**Dependencies**: Story 1  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [数据架构](../../architecture/data-architecture.md)

---

## User Story

作为开发者，我希望完善 Ent Schema 的配置管理表，以便实现动态配置系统。

---

## Acceptance Criteria

- [ ] 完善 `ConfigItem` Schema 定义
- [ ] 添加必要的索引（key、status、priority）
- [ ] 添加 Hooks（创建时间、更新时间）
- [ ] 实现数据验证规则
- [ ] 执行数据库迁移
- [ ] 编写种子数据脚本
- [ ] 编写单元测试

---

## Implementation Tasks

- [ ] 更新 `core/ent/schema/configitem.go`
- [ ] 添加唯一索引（key）
- [ ] 添加复合索引（status + priority）
- [ ] 实现 Hooks（CreatedAt、UpdatedAt）
- [ ] 添加字段验证（key 格式、value JSON）
- [ ] 生成迁移文件（`go generate`）
- [ ] 创建种子数据脚本
- [ ] 编写测试

---

## Technical Details

```go
// core/ent/schema/configitem.go

package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type ConfigItem struct {
    ent.Schema
}

func (ConfigItem) Fields() []ent.Field {
    return []ent.Field{
        field.String("key").
            Unique().
            NotEmpty().
            Match(regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)),
        field.JSON("value", map[string]interface{}{}),
        field.String("description").Optional(),
        field.Enum("status").
            Values("active", "inactive").
            Default("active"),
        field.Int("priority").
            Default(0),
        field.Time("created_at").
            Default(time.Now),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (ConfigItem) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("key").Unique(),
        index.Fields("status", "priority"),
    }
}
```

### 种子数据

```bash
# scripts/seed-config.sh
INSERT INTO config_items (key, value, description, status, priority)
VALUES 
('app.name', '{"value": "AppRun"}', 'Application name', 'active', 100),
('app.version', '{"value": "0.1.0"}', 'API version', 'active', 100);
```

---

## Test Cases

- [ ] Key 唯一约束生效
- [ ] Key 格式验证正确
- [ ] 索引创建成功
- [ ] Hooks 自动设置时间戳
- [ ] 种子数据插入成功

---

## Related Docs

- [数据架构](../../architecture/data-architecture.md)
- [Ent 官方文档](https://entgo.io/docs/schema-def/)

---

**Created**: 2025-12-27  
**Updated**: 2025-12-27  
**Maintainer**: Architect Agent
