# Migrate Inspect & Repair 使用指南

**新增命令**: `inspect` 和 `repair`  
**日期**: 2026-01-16  
**用途**: 处理数据库结构损坏和 schema drift 场景  

---

## 背景：两种迁移场景

### 场景 1: 新迁移文件生成（正常开发）✅ 已支持

**情况**:
```bash
# 开发者修改了 Ent schema
vim core/ent/schema/user.go
```

**解决方案**: Versioned Migrations
```bash
./core/bin/apprun migrate diff add_user_field
./core/bin/apprun migrate apply
```

### 场景 2: 数据库结构破坏（修复场景）✅ 新增支持

**情况**:
```sql
-- 手动执行了 SQL（意外或测试）
ALTER TABLE users DROP COLUMN email;

-- 或者迁移失败，表结构不一致
-- 或者迁移文件丢失/损坏
```

**之前的问题**:
```bash
$ ./core/bin/apprun migrate status
✅ All migrations applied  # ← 误报！

$ ./core/bin/apprun migrate diff fix
Error: no schema changes detected  # ← 对比的是 schema 和 migrations 目录
```

**新的解决方案**: Declarative Migrations
```bash
./core/bin/apprun migrate inspect          # 检测问题
./core/bin/apprun migrate repair --execute # 修复
```

---

## 新命令详解

### 1. `migrate inspect` - 检测 Schema Drift

**用途**: 检查数据库实际结构，对比 Ent schema

**与 `migrate status` 的区别**:

| 命令 | 检查内容 | 用途 |
|------|---------|------|
| `status` | 迁移历史（哪些文件已应用） | 查看迁移进度 |
| `inspect` | **数据库实际结构** vs Ent schema | 检测手动修改 |

**示例 1: 正常情况**
```bash
$ ./core/bin/apprun migrate inspect

🔍 Inspecting database schema...

✅ Database schema is in sync with Ent schema

No schema drift detected:
  - All tables exist as defined in Ent schema
  - All columns match their definitions
  - All indexes are correctly created
```

**示例 2: 检测到差异**
```bash
$ ./core/bin/apprun migrate inspect

🔍 Inspecting database schema...

⚠️  Schema drift detected!

Differences between database and Ent schema:
-- Drop column "old_field" from table "users"
ALTER TABLE `users` DROP COLUMN `old_field`;

-- Add column "email" to table "users"
ALTER TABLE `users` ADD COLUMN `email` varchar(255) NOT NULL;

💡 Possible causes:
   - Manual database changes (ALTER TABLE, DROP COLUMN, etc.)
   - Failed migration that left database inconsistent
   - Migration files corrupted or lost

🔧 To fix:
   apprun migrate repair          # Preview SQL
   apprun migrate repair --execute # Apply fix

⚠️  Or reset and re-apply migrations:
   apprun migrate reset
   apprun migrate sync
```

### 2. `migrate repair` - 修复 Schema

**用途**: 使用声明式迁移修复数据库结构

**安全机制**:
- ✅ 默认 dry-run（只显示 SQL）
- ✅ 需要 `--execute` 才执行
- ✅ 生产环境禁用
- ✅ 遵循 diff policy（不删除表/列）

**示例 1: Dry-run（默认）**
```bash
$ ./core/bin/apprun migrate repair

🔧 Repairing database schema...

🔍 Dry-run mode: SQL will be displayed but not executed

📋 Detected schema differences:
-- Planned Changes:
-- Add column "email" to table "users"
ALTER TABLE `users` ADD COLUMN `email` varchar(255) NOT NULL;

💡 To apply these changes:
   apprun migrate repair --execute
```

**示例 2: 执行修复**
```bash
$ ./core/bin/apprun migrate repair --execute

🔧 Repairing database schema...

⚠️  WARNING: This will modify your database!

Recommendations before proceeding:
  1. Backup your database
  2. Review the SQL that will be executed

📋 Detected schema differences:
-- Apply SQL...
ALTER TABLE `users` ADD COLUMN `email` varchar(255) NOT NULL;

✅ Database schema repaired successfully!

⚠️  Important notes:
   - This was a declarative repair (no migration file created)
   - Migration history may be out of sync
   - Consider regenerating migrations if needed:
     1. Reset: apprun migrate reset
     2. Generate: apprun migrate diff baseline
     3. Apply: apprun migrate apply
```

**示例 3: 生产环境被阻止**
```bash
$ APP_ENV=production ./core/bin/apprun migrate repair

❌ ERROR: 'migrate repair' is not allowed in production

Production environments must use versioned migrations:
  1. Generate migration: apprun migrate diff fix_schema
  2. Review SQL file: cat migrations/*.sql
  3. Apply migration: apprun migrate apply
```

---

## 使用场景

### 场景 A: 开发时意外修改数据库

**情况**:
```sql
-- 开发者不小心手动执行了 SQL
psql -d apprun -c "DROP TABLE posts;"
```

**修复流程**:
```bash
# 1. 检查问题
$ ./core/bin/apprun migrate inspect
⚠️  Schema drift detected!
-- Missing table "posts"

# 2. 预览修复 SQL
$ ./core/bin/apprun migrate repair
-- Will create table "posts" ...

# 3. 执行修复
$ ./core/bin/apprun migrate repair --execute
✅ Database schema repaired successfully!

# 4. 验证
$ ./core/bin/apprun migrate inspect
✅ Database schema is in sync
```

### 场景 B: 迁移文件损坏

**情况**:
```bash
# 迁移文件被修改或删除
rm core/migrations/20260115_add_users.sql

# Atlas checksum 验证失败
$ ./core/bin/apprun migrate status
Error: checksum mismatch
```

**修复流程**:

**选项 1: 使用 repair（快速修复）**
```bash
# 忽略迁移历史，直接修复数据库
./core/bin/apprun migrate repair --execute
```

**选项 2: 重置迁移历史（完整解决）**
```bash
# 1. 备份数据（如果需要）
pg_dump apprun > backup.sql

# 2. 重置数据库
./core/bin/apprun migrate reset

# 3. 重新生成基线迁移
./core/bin/apprun migrate diff baseline

# 4. 应用
./core/bin/apprun migrate apply
```

### 场景 C: 测试后快速恢复

**情况**:
```bash
# 测试时修改了很多表结构
# 想快速恢复到 Ent schema 定义的状态
```

**修复流程**:
```bash
# 不用重置，直接修复
./core/bin/apprun migrate repair --execute
```

### 场景 D: 生产环境检测（只读）

**情况**:
```bash
# 怀疑生产数据库被手动修改
```

**检查流程**:
```bash
# 1. 检查（只读操作，安全）
$ APP_ENV=production ./core/bin/apprun migrate inspect
⚠️  Schema drift detected!

# 2. 不能直接 repair（生产禁用）
$ APP_ENV=production ./core/bin/apprun migrate repair
❌ ERROR: repair not allowed in production

# 3. 正确做法：生成迁移文件
$ ./core/bin/apprun migrate diff fix_production_schema
$ cat core/migrations/20260116_fix_production_schema.sql
# 审查 SQL...

# 4. 在生产应用
$ APP_ENV=production ./core/bin/apprun migrate apply
```

---

## 最佳实践

### ✅ DO（推荐）

**开发环境**:
```bash
# 正常开发：使用 Versioned
./core/bin/apprun migrate sync

# 数据库损坏：使用 Declarative
./core/bin/apprun migrate repair --execute

# 定期检查
./core/bin/apprun migrate inspect
```

**生产环境**:
```bash
# 只使用 Versioned Migrations
./core/bin/apprun migrate apply

# 只读检查（不修复）
./core/bin/apprun migrate inspect
```

### ❌ DON'T（禁止）

```bash
# ❌ 生产环境使用 repair
APP_ENV=production ./core/bin/apprun migrate repair --execute

# ❌ 不检查直接修复
./core/bin/apprun migrate repair --execute  # 先 inspect！

# ❌ 忽略 repair 后的警告
# Repair 后应该：
# 1. 验证数据完整性
# 2. 考虑重新生成迁移历史
# 3. 提交新的 baseline 迁移
```

---

## 工作流对比

### Versioned Migrations（主流程）

```
修改 Ent Schema
    ↓
migrate diff
    ↓
生成 .sql 文件
    ↓
审查 SQL
    ↓
migrate apply
    ↓
提交到 Git
```

**特点**: 完整审计追踪，团队协作

### Declarative Migrations（修复流程）

```
数据库损坏/drift
    ↓
migrate inspect
    ↓
显示差异
    ↓
migrate repair
    ↓
直接修复数据库（无文件）
```

**特点**: 快速修复，无历史记录

---

## 技术实现

### Atlas API 使用

**Versioned Mode（当前的 apply/diff/status）**:
```go
// 使用 migrate 包
import "ariga.io/atlas/sql/migrate"

dir := migrate.NewDir("migrations")
executor := migrate.NewExecutor(drv, dir, table)
executor.Execute(ctx, migrate.PlanOptions{})
```

**Declarative Mode（新的 inspect/repair）**:
```go
// 使用 schema 包
import "ariga.io/atlas/sql/schema"

// 读取实际数据库
actual, _ := drv.InspectRealm(ctx, &schema.InspectRealmOption{})

// 读取期望状态（Ent）
desired, _ := ent.LoadSchema(ctx)

// 计算差异
diff, _ := drv.SchemaDiff(actual, desired)

// 生成 SQL
changes, _ := drv.PlanChanges(ctx, "repair", diff)
```

### 为什么需要 Atlas CLI？

Declarative 模式调用 `atlas schema diff/apply` 命令：

```bash
# Atlas CLI 能解析 ent:// URL
atlas schema apply \
  --url "postgres://localhost/db" \
  --to "ent://ent/schema" \        # ← CLI 能加载 Ent schema
  --dev-url "docker://postgres/15"  # ← CLI 能启动临时容器
```

Go SDK 不支持这些高级功能，所以需要 CLI。

---

## 常见问题

### Q: inspect 和 status 有什么区别？

**A**: 
- `status`: 检查**迁移历史**（哪些 .sql 文件已执行）
- `inspect`: 检查**数据库实际结构**（表/列是否符合 Ent 定义）

### Q: repair 会创建迁移文件吗？

**A**: 不会！repair 是声明式迁移，直接修改数据库，不生成 .sql 文件。

### Q: repair 后迁移历史会乱吗？

**A**: 可能会。解决方案：
1. 修复后重新 baseline：`migrate diff baseline`
2. 或者接受不一致（开发环境可接受）

### Q: 生产环境能用 repair 吗？

**A**: 不能！代码会阻止。生产必须：
1. 生成迁移：`migrate diff fix`
2. 审查 SQL
3. 应用迁移：`migrate apply`

### Q: 什么时候用 repair，什么时候用 reset?

**A**:
- `repair`: 数据库有数据，想保留数据，只修复结构
- `reset`: 数据不重要，想完全清空重来

---

## 总结

### Atlas 的两种模式

| 模式 | 依赖 | 审计 | 团队协作 | 适用场景 |
|------|------|------|---------|---------|
| **Versioned** | 迁移文件 | ✅ 完整 | ✅ Git | 生产/正常开发 |
| **Declarative** | 数据库连接 | ❌ 无 | ❌ 无法共享 | 开发修复 |

### 新命令摘要

```bash
# 检查数据库和 schema 差异
./core/bin/apprun migrate inspect

# 修复差异（dry-run）
./core/bin/apprun migrate repair

# 执行修复
./core/bin/apprun migrate repair --execute
```

### 何时使用

| 场景 | 命令 |
|------|------|
| 正常开发 | `sync` 或 `diff` + `apply` |
| 数据库损坏 | `inspect` + `repair` |
| 生产部署 | `apply` |
| 检查差异 | `inspect` |
| 完全重置 | `reset` + `sync` |

---

## 相关文档

- [Atlas 迁移模式详解](ATLAS-MIGRATION-MODES.md)
- [迁移文件维护指南](MIGRATIONS-MAINTENANCE.md)
- [Story 1.5 实现细节](../sprint-artifacts/sprint-0/1-5-database-migration.md)
