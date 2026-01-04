# Story 4: Ent Schema 配置管理
# Sprint 0: Infrastructure建设

**Priority**: P0  
**Effort**: 1 天  
**Owner**: Backend Dev  
**Dependencies**: Story 1  
**Status**: Done  
**Module**: Config  
**Issue**: #TBD  
**Related**: [数据架构](../../architecture/data-architecture.md)

---

## User Story

作为开发者，我希望完善 Ent Schema 的配置管理表，以便实现动态配置系统。

---

## Acceptance Criteria

- [x] 完善 `ConfigItem` Schema 定义
- [x] 添加必要的索引（key、status）
- [x] 添加 Hooks（创建时间、更新时间）
- [x] 实现数据验证规则
- [x] 执行数据库迁移
- [x] 编写单元测试

---

## Implementation Tasks

- [x] 更新 `core/ent/schema/configitem.go`（添加审计和状态字段）
- [x] 添加唯一索引（key）
- [x] 实现自动时间戳 Hooks（created_at、updated_at）
- [x] 数据库自动迁移（通过 `database.Connect()`）
- [x] 编写 Schema 测试（验证约束、索引）

---

## Technical Details

### Schema Definition

```go
// core/ent/schema/configitem.go
package schema

import (
    "time"
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type Configitem struct {
    ent.Schema
}

func (Configitem) Fields() []ent.Field {
    return []ent.Field{
        // 核心字段
        field.String("key").
            Unique().
            NotEmpty().
            Comment("配置项的键，如 app.name"),
        field.String("value").
            Comment("配置项的值（字符串）"),
        field.Bool("is_dynamic").
            Default(false).
            Comment("是否为动态配置（db:true）"),
        
        // 状态管理
        field.Enum("status").
            Values("active", "inactive").
            Default("active").
            Comment("配置项状态，支持软删除"),
        
        // 审计字段（自动时间戳）
        field.Time("created_at").
            Default(time.Now).
            Immutable().
            Comment("创建时间"),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now).
            Comment("更新时间"),
    }
}

func (Configitem) Indexes() []ent.Index {
    return []ent.Index{
        // key 唯一索引（主查询字段）
        index.Fields("key").Unique(),
        
        // status 索引（支持按状态筛选）
        index.Fields("status"),
    }
}
```

### 自动迁移

迁移通过 `pkg/database.Connect()` 自动执行：

```go
// core/pkg/database/postgres.go (已实现)
func Connect(ctx context.Context, cfg *Config) (Client, error) {
    client, err := ent.Open(cfg.Driver, dsn)
    if err != nil {
        return nil, err
    }
    
    // 自动创建/更新表结构
    if err := client.Schema.Create(ctx); err != nil {
        client.Close()
        return nil, err
    }
    
    return &entClient{client: client}, nil
}
```

**执行时机**: 应用启动时在 `main.go` 中调用 `database.Connect()`

---

## Test Cases

- [x] Key 唯一约束生效（重复 key 插入失败）
- [x] Status 枚举验证（仅接受 active/inactive）
- [x] 索引创建成功（key 唯一索引、status 索引）
- [x] 自动时间戳生效（created_at、updated_at）
- [x] 软删除功能验证（status=inactive）

---

## Implementation Notes

### 字段设计决策

1. **保留字段**:
   - `key`, `value`, `is_dynamic`: Story 10 核心功能
   - `status`: 支持软删除，便于审计
   - `created_at`, `updated_at`: 审计追踪

2. **移除字段**:
   - ❌ `priority`: 当前系统基于 key 精确匹配，无排序需求
   - ❌ `description`: 等 i18n 标准化后再考虑（避免多语言复杂度）

3. **索引策略**:
   - ✅ `key` 唯一索引: 主查询字段，必需
   - ✅ `status` 索引: 支持按状态筛选
   - ❌ 复合索引: 无明确查询模式，避免过度优化

### 数据库迁移流程

#### 初次部署（新数据库）

```bash
# 1. 修改 Schema
vim core/ent/schema/configitem.go

# 2. 生成 Ent 代码
go generate ./core/ent

# 3. 启动应用（自动迁移）
make run-local
```

#### 升级现有数据库

如果数据库已有旧版 `configitems` 表（无 status/created_at/updated_at 字段）：

```bash
# 1. 执行迁移脚本（添加字段并填充默认值）
docker compose -f docker-compose.local.yml exec -T postgres \
  psql -U apprun -d apprun_dev < core/migrations/001_add_configitem_audit_fields.sql

# 2. 重新生成 Ent 代码
go generate ./core/ent

# 3. 启动应用
make run-local
```

**迁移脚本位置**: `core/migrations/001_add_configitem_audit_fields.sql`

**迁移安全性**:
- ✅ 使用事务（BEGIN/COMMIT）确保原子性
- ✅ 先添加 nullable 字段，填充默认值后再设置 NOT NULL
- ✅ 现有数据自动设置 `status='active'`, `created_at/updated_at=NOW()`
- ✅ Schema.Create() 是幂等操作，不会删除数据

### 未来扩展点

当需要以下功能时再扩展：

1. **Description 字段**: 等 i18n 标准化后添加
   ```go
   field.JSON("description", map[string]string{}).Optional()
   // {"en": "App name", "zh-CN": "应用名称"}
   ```

2. **Priority 字段**: 如有排序需求
   ```go
   field.Int("priority").Default(0)
   index.Fields("status", "priority")
   ```

---

## Related Docs

- [Story 10: Config System](story-10-config-basic.md)
- [数据架构](../../architecture/data-architecture.md)
- [Ent Schema 文档](https://entgo.io/docs/schema-def/)
- [Ent Hooks 文档](https://entgo.io/docs/hooks/)

---

**Created**: 2025-12-27  
**Updated**: 2026-01-04  
**Maintainer**: Amelia (Dev Agent)
