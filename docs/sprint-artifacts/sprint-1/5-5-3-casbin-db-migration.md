# Story 5.5.3: Casbin 数据库存储迁移
# Sprint 2: 认证与授权 (Auth Epic)

**Priority**: P0 (必需)  
**Effort**: 1 天  
**Owner**: Backend Dev  
**Dependencies**: 
- Story 5.5 (RBAC 权限控制基础设施) - ✅ 已完成
- Story 1.14 (Database Package) - ✅ 已完成
- Story 1.5 (Database Migration) - ✅ 已完成

**Status**: done  
**Related Stories**: 
- Story 5.5 (RBAC 权限控制) - 父 Story
- Story 5.5.2 (RBAC API Endpoints)
**Module**: Authorization  
**Epic**: [auth-epic](../../epics/5-auth-epic.md)  
**Issue**: #TBD  

**ADR Reference**: [ADR-001: 权限机制统一到 Casbin 数据库存储](../../epics/5-auth-epic.md#adr-001-权限机制统一到-casbin-数据库存储)

---

## User Story

作为 **apprun 平台开发者**，我希望将 Casbin 权限策略从 CSV 文件迁移到数据库存储 (`casbin_rule` 表)，以实现单一数据源的权限管理，支持动态权限配置，并确保多实例部署时的策略一致性。

**业务价值**：
- 消除 CSV 文件与数据库双数据源的不一致风险
- 支持运行时动态修改权限策略（无需重启服务）
- 多实例部署时权限策略自动同步
- 为 RBAC API Endpoints (Story 5.5.2) 提供数据存储基础

---

## Acceptance Criteria

### 功能验收
- [x] 创建 `casbin_rule` Ent Schema
- [x] 生成数据库迁移脚本 (Atlas Migration)
- [x] 实现 Ent Adapter 集成 (自定义 EntAdapter，不使用官方 ent-adapter 因版本冲突)
- [x] 实现 Bootstrap 时默认策略导入（幂等）
- [x] 废弃 CSV 文件加载逻辑 (改为内存模式 fallback)
- [ ] ~~废弃 `RequirePlatformAdmin` 中间件~~ → **Deferred to Story 5.5.4** (middleware migration out of scope per ADR-001)

### 非功能验收
- [x] 单元测试覆盖 Adapter 初始化和策略加载
- [x] 集成测试验证策略 CRUD 操作
- [x] 迁移脚本向后兼容（不破坏现有数据）
- [x] 文档更新 → Story file updated; README update deferred to tech-writer

### 数据迁移验收
- [x] 现有 `default_policy.csv` 内容迁移到数据库 (通过 SeedDefaultPolicies)
- [x] 迁移过程幂等（可重复执行）
- [x] 迁移后 CSV 文件保留作为参考，但不再被加载 (仅作为内存模式 fallback)

---

## Technical Design

### 1. CasbinRule Ent Schema

```go
// core/ent/schema/casbin_rule.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type CasbinRule struct {
    ent.Schema
}

func (CasbinRule) Fields() []ent.Field {
    return []ent.Field{
        field.Int64("id"),
        field.String("ptype").MaxLen(100).NotEmpty(),  // p, g, g2
        field.String("v0").MaxLen(100).Optional(),     // sub (user/role)
        field.String("v1").MaxLen(100).Optional(),     // dom/obj
        field.String("v2").MaxLen(100).Optional(),     // obj/act
        field.String("v3").MaxLen(100).Optional(),     // act
        field.String("v4").MaxLen(100).Optional(),     // reserved
        field.String("v5").MaxLen(100).Optional(),     // reserved
    }
}

func (CasbinRule) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("ptype"),
        index.Fields("v0"),
        index.Fields("v1"),
        index.Fields("ptype", "v0", "v1", "v2", "v3"),  // 唯一策略索引
    }
}
```

### 2. Database Migration

```sql
-- migrations/00X_create_casbin_rule.sql
CREATE TABLE casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100)
);

CREATE INDEX idx_casbin_rule_ptype ON casbin_rule(ptype);
CREATE INDEX idx_casbin_rule_v0 ON casbin_rule(v0);
CREATE INDEX idx_casbin_rule_v1 ON casbin_rule(v1);
```

### 3. Ent Adapter 集成

```go
// core/internal/rbac/enforcer.go
package rbac

import (
    "github.com/casbin/casbin/v2"
    entadapter "github.com/casbin/ent-adapter"
    _ "embed"
)

//go:embed model.conf
var modelConf string

type Config struct {
    DatabaseURL string  // PostgreSQL 连接字符串
}

func InitEnforcer(cfg Config) (*casbin.Enforcer, error) {
    // 1. 创建 Ent Adapter
    adapter, err := entadapter.NewAdapter("postgres", cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to create ent adapter: %w", err)
    }

    // 2. 从内嵌 model.conf 创建 Enforcer
    m, err := model.NewModelFromString(modelConf)
    if err != nil {
        return nil, fmt.Errorf("failed to load model: %w", err)
    }

    // 3. 创建 Enforcer
    e, err := casbin.NewEnforcer(m, adapter)
    if err != nil {
        return nil, fmt.Errorf("failed to create enforcer: %w", err)
    }

    // 4. 加载策略
    if err := e.LoadPolicy(); err != nil {
        return nil, fmt.Errorf("failed to load policy: %w", err)
    }

    return e, nil
}
```

### 4. Bootstrap 默认策略导入

```go
// core/internal/bootstrap/rbac.go
package bootstrap

// DefaultPolicies 定义系统默认策略
var DefaultPolicies = [][]string{
    // Platform-level policies
    {"p", "platform_admin", "*", "*"},

    // Project Owner policies
    {"p", "owner", "config", "*"},
    {"p", "owner", "data", "*"},
    {"p", "owner", "storage", "*"},
    {"p", "owner", "function", "*"},
    {"p", "owner", "workflow", "*"},
    {"p", "owner", "member", "*"},
    {"p", "owner", "project", "*"},

    // Project Admin policies
    {"p", "admin", "config", "create"},
    {"p", "admin", "config", "read"},
    {"p", "admin", "config", "update"},
    {"p", "admin", "config", "delete"},
    {"p", "admin", "data", "*"},
    {"p", "admin", "storage", "*"},
    {"p", "admin", "function", "*"},
    {"p", "admin", "workflow", "*"},
    {"p", "admin", "member", "*"},
    {"p", "admin", "project", "read"},
    {"p", "admin", "project", "update"},

    // Project Member policies
    {"p", "member", "config", "read"},
    {"p", "member", "data", "create"},
    {"p", "member", "data", "read"},
    {"p", "member", "data", "update"},
    {"p", "member", "data", "delete"},
    {"p", "member", "storage", "*"},
    {"p", "member", "function", "read"},
    {"p", "member", "function", "execute"},
    {"p", "member", "workflow", "read"},
    {"p", "member", "workflow", "execute"},

    // Project Viewer policies (read-only)
    {"p", "viewer", "config", "read"},
    {"p", "viewer", "data", "read"},
    {"p", "viewer", "storage", "read"},
    {"p", "viewer", "function", "read"},
    {"p", "viewer", "workflow", "read"},
}

// SeedDefaultPolicies 导入默认策略（幂等）
func SeedDefaultPolicies(enforcer *casbin.Enforcer) error {
    for _, policy := range DefaultPolicies {
        // 检查策略是否已存在
        hasPolicy, err := enforcer.HasPolicy(policy[1:])
        if err != nil {
            return fmt.Errorf("check policy failed: %w", err)
        }
        if hasPolicy {
            continue  // 已存在，跳过
        }

        // 添加策略
        if _, err := enforcer.AddPolicy(policy[1:]); err != nil {
            return fmt.Errorf("add policy failed: %w", err)
        }
    }
    return nil
}
```

### 5. 废弃 RequirePlatformAdmin 中间件

**变更前**：
```go
// 使用两个不同的中间件
r.With(middleware.RequirePlatformAdmin).Get("/admin/users", ...)
r.With(middleware.RequirePermission("config", "read")).Get("/config", ...)
```

**变更后**：
```go
// 统一使用 RequirePermission
r.With(middleware.RequirePermission("user", "manage")).Get("/admin/users", ...)
r.With(middleware.RequirePermission("config", "read")).Get("/config", ...)
```

**迁移步骤**：
1. 为 `platform_admin` 角色添加相应资源权限（如 `user:manage`）
2. 替换所有 `RequirePlatformAdmin` 调用为 `RequirePermission`
3. 删除 `RequirePlatformAdmin` 中间件代码

---

## Implementation Tasks

### Phase 1: Schema & Migration (0.5 天)
1. [x] 创建 `core/ent/schema/casbin_rule.go` *(已存在)*
2. [x] 运行 `go generate ./ent/...` 生成代码 *(已完成)*
3. [x] 使用 Atlas 生成迁移脚本 *(已存在)*
4. [x] 测试迁移脚本执行 *(已验证)*

### Phase 2: Adapter Integration (0.5 天)
5. [x] 添加 `github.com/casbin/ent-adapter` 依赖 *(使用自定义 EntAdapter)*
6. [x] 修改 `InitEnforcer()` 使用 Ent Adapter
7. [x] 实现 `SeedDefaultPolicies()` Bootstrap 函数
8. [x] 修改 Server Bootstrap 调用新的初始化逻辑

### Phase 3: Middleware Migration (可选，建议拆分)
9. [ ] 识别所有 `RequirePlatformAdmin` 使用点
10. [ ] 替换为 `RequirePermission` 调用
11. [ ] 删除废弃的中间件代码

### Phase 4: Testing & Cleanup
12. [x] 编写单元测试（Adapter 初始化）→ `ent_adapter_test.go` created
13. [x] 编写集成测试（策略 CRUD）→ `ent_adapter_test.go` covers CRUD
14. [x] ~~删除或归档~~ `default_policy.csv` 保留为 fallback 模式数据源
15. [ ] 更新配置文档 → Deferred to tech-writer

---

## Dependencies

### Go Packages
```go
require (
    github.com/casbin/casbin/v2 v2.x.x
    github.com/casbin/ent-adapter v0.x.x  // ✨ 新增
)
```

### Configuration Changes
```yaml
# config/default.yaml
auth:
  casbin:
    adapter: database  # ✨ 固定为 database
    # policy_path: "./config/casbin_policy.csv"  # 已废弃
```

---

## Risks & Mitigations

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| 迁移过程数据丢失 | 高 | 备份数据库，迁移前导出 CSV |
| Ent Adapter 兼容性 | 中 | 使用官方 adapter，版本锁定 |
| 性能回归 | 低 | 数据库查询优化，必要时加索引 |
| 多实例同步延迟 | 低 | 定期 LoadPolicy，或使用 Watcher |

---

## Related Documentation

- [Epic 5: 认证与授权](../../epics/5-auth-epic.md)
- [ADR-001: 权限机制统一到 Casbin 数据库存储](../../epics/5-auth-epic.md#adr-001-权限机制统一到-casbin-数据库存储)
- [Story 5.5: RBAC 权限控制](5-5-rbac-permissions.md)
- [Casbin Ent Adapter 文档](https://github.com/casbin/ent-adapter)

---

**文档维护**: Bob (SM Agent)  
**创建日期**: 2026-01-19  
**最后更新**: 2026-01-19
