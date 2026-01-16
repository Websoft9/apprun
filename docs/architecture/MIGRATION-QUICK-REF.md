# Atlas 迁移系统快速参考

## 命令对比

| 命令 | 模式 | 用途 | 生成文件 | 生产可用 |
|------|------|------|---------|---------|
| `diff` | Versioned | 生成迁移文件 | ✅ .sql | ✅ |
| `apply` | Versioned | 应用迁移文件 | ❌ | ✅ |
| `sync` | Versioned | 自动生成+应用 | ✅ .sql | ⚠️ 慎用 |
| `status` | Versioned | 检查迁移历史 | ❌ | ✅ |
| `rollback` | Versioned | 回滚上次迁移 | ❌ | ✅ |
| `reset` | Versioned | 清空数据库 | ❌ | ❌ |
| **`inspect`** | **Declarative** | **检查结构差异** | ❌ | ✅（只读）|
| **`repair`** | **Declarative** | **修复结构差异** | ❌ | ❌ |
| **`clean`** | **Maintenance** | **清理失败迁移** | ❌ | ❌ |

## 使用场景

### 场景1: 正常开发流程
```bash
# 修改 Ent schema
vim core/ent/schema/user.go

# 生成并应用
./core/bin/apprun migrate sync

# 或者分步骤
./core/bin/apprun migrate diff add_user_field
./core/bin/apprun migrate apply
```

### 场景2: 数据库损坏修复
```bash
# 1. 检查问题
./core/bin/apprun migrate inspect

# 2. 预览修复 SQL
./core/bin/apprun migrate repair

# 3. 执行修复
./core/bin/apprun migrate repair --execute

# 4. 验证
./core/bin/apprun migrate inspect
```

### 场景3: repair后误生成迁移文件（你遇到的问题）
```bash
# 问题：用repair创建表后，又用diff生成了迁移文件，导致apply失败
# 原因：迁移文件试图创建已存在的表

# 解决方法：
# 1. 清理迁移历史记录
./core/bin/apprun migrate clean --all          # 清除所有失败记录
# 或
./core/bin/apprun migrate clean 20260116093904 # 清除特定版本

# 2. 删除冗余的迁移文件
rm core/migrations/20260116093904_*.sql

# 3. 重新计算校验和
cd core && atlas migrate hash --dir file://migrations

# 4. 验证状态
./core/bin/apprun migrate status
```

### 场景4: 完全重建数据库
```bash
# 方法A: 使用 repair（推荐，空数据库场景）
./core/bin/apprun migrate reset               # 清空数据库
./core/bin/apprun migrate repair --execute    # 直接创建所有表
# 注意：不会生成迁移文件，不会记录迁移历史

# 方法B: 使用迁移文件（推荐，有迁移文件时）
./core/bin/apprun migrate reset               # 清空数据库
./core/bin/apprun migrate apply               # 应用所有迁移文件

# 方法C: 使用 sync（开发环境）
./core/bin/apprun migrate reset               # 清空数据库
./core/bin/apprun migrate sync                # 自动生成并应用
```

### 场景5: 迁移文件丢失
```bash
# 选项 A: 快速修复（保留数据）
./core/bin/apprun migrate repair --execute

# 选项 B: 完全重置（清空数据）
./core/bin/apprun migrate reset
./core/bin/apprun migrate diff baseline
./core/bin/apprun migrate apply
```

### 生产环境检查
```bash
# 只检查，不修复
APP_ENV=production ./core/bin/apprun migrate inspect

# 如果有差异，生成迁移文件
./core/bin/apprun migrate diff fix_production

# 审查 SQL
cat core/migrations/20260116_fix_production.sql

# 应用
APP_ENV=production ./core/bin/apprun migrate apply
```

## Makefile 快捷命令

```bash
# Versioned Migrations
make db-sync              # 自动同步（开发）
make db-diff NAME=xxx     # 生成迁移
make db-migrate           # 应用迁移
make db-status            # 查看状态
make db-rollback          # 回滚
make db-reset             # 重置

# Declarative Migrations（新增）
make db-inspect           # 检查差异
make db-repair            # 预览修复
make db-repair-execute    # 执行修复
```

## 两种模式的本质区别

### Versioned: 基于迁移文件

```
┌─────────────┐     ┌──────────────┐     ┌──────────┐
│ Ent Schema  │────▶│ Migrations   │────▶│ Database │
└─────────────┘     │ (历史文件)    │     └──────────┘
                    └──────────────┘
                           │
                           ▼
                    Git 版本控制
```

**优点**: 完整历史，可审查，可回滚  
**缺点**: 依赖文件完整性

### Declarative: 直接对比

```
┌─────────────┐
│ Ent Schema  │◀────┐
└─────────────┘     │
                    │ Atlas
                    │ Diff
┌──────────┐        │
│ Database │◀───────┘
└──────────┘
```

**优点**: 不依赖历史，能修复损坏  
**缺点**: 无历史记录，无法共享

## 何时使用哪种模式？

| 场景 | 使用模式 | 理由 |
|------|---------|------|
| 日常开发 | Versioned | 需要 Git 追踪 |
| 团队协作 | Versioned | 需要共享迁移 |
| 生产部署 | Versioned | 需要审查 SQL |
| 数据库损坏 | Declarative | 快速修复 |
| 测试后恢复 | Declarative | 不需要历史 |
| 检查差异 | Declarative | 验证一致性 |

## 安全守则

### ✅ 安全操作

```bash
# 任何环境都安全
./core/bin/apprun migrate status
./core/bin/apprun migrate inspect
./core/bin/apprun migrate validate

# 开发环境安全
./core/bin/apprun migrate sync
./core/bin/apprun migrate repair --execute
./core/bin/apprun migrate reset
```

### ⚠️ 需要谨慎

```bash
# 生产环境需要审查
./core/bin/apprun migrate apply

# 回滚需要验证
./core/bin/apprun migrate rollback
```

### ❌ 禁止操作

```bash
# 生产环境禁止
APP_ENV=production ./core/bin/apprun migrate reset        # ❌
APP_ENV=production ./core/bin/apprun migrate repair       # ❌
APP_ENV=production ./core/bin/apprun migrate sync         # ❌
APP_ENV=production ./core/bin/apprun migrate clean        # ❌
```

## ⚙️ 配置说明

### Atlas 配置嵌入

**设计优势**：
- ✅ **单一二进制**：atlas.hcl 嵌入到 apprun 二进制中
- ✅ **零配置依赖**：无需外部配置文件，开箱即用
- ✅ **模块化统一**：配置与命令代码在 cmd 目录下
- ✅ **运行时动态**：根据数据库 URL 动态生成配置

**实现原理**：
```go
// core/cmd/migrate.go
//go:embed atlas.hcl
var atlasConfigTemplate string

// 运行时生成临时配置
func generateAtlasConfig(dbURL string) (configPath string, cleanup func(), err error) {
    // 替换模板中的占位符
    config := strings.ReplaceAll(atlasConfigTemplate, "{{db_url}}", dbURL)
    
    // 创建临时文件
    tmpFile, _ := os.CreateTemp("", "atlas-*.hcl")
    tmpFile.WriteString(config)
    
    // 返回路径和清理函数
    return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }, nil
}
```

### Atlas 元数据表过滤

项目已配置自动过滤 Atlas 元数据表，避免误报 schema drift：

**代码层面**（已实现）：
```go
// core/pkg/database/migrate.go
func filterAtlasMetadataTables(output string) string {
    // 自动过滤 atlas_schema_revisions 表
    // inspect 和 repair 命令不会报告此表的差异
}
```

**Atlas 配置**（嵌入式）：
```hcl
# core/cmd/atlas.hcl (嵌入到二进制)
env "local" {
  src = "ent://ent/schema"
  schemas = ["public"]
  
  # 排除 Atlas 元数据表
  exclude = [
    "atlas_schema_revisions",
  ]
}
```

**效果**：
- ✅ `migrate inspect` 不再报告 `atlas_schema_revisions` 为 drift
- ✅ `migrate diff` 不会生成删除此表的 SQL
- ✅ `migrate repair` 不会尝试删除此表
- ✅ 保持应用 schema 与框架元数据分离
- ✅ 配置随二进制分发，无需手动管理

## 🎯 最佳实践

### ✅ 推荐的工作流程

**开发环境：**
```bash
# 1. 修改 schema
vim core/ent/schema/user.go

# 2. 自动生成并应用（快速迭代）
./core/bin/apprun migrate sync

# 3. 提交到 git
git add core/migrations/*.sql
git commit -m "feat: add user phone field"
```

**生产环境：**
```bash
# 1. 拉取代码（包含迁移文件）
git pull origin main

# 2. 审查迁移文件
cat core/migrations/20260116_*.sql

# 3. 应用迁移
./core/bin/apprun migrate apply

# 4. 验证
./core/bin/apprun migrate status
```

### ❌ 常见错误

**错误1：repair 后又 diff**
```bash
# ❌ 错误流程
./core/bin/apprun migrate repair --execute  # 创建表
./core/bin/apprun migrate diff baseline     # 生成重复的迁移文件
./core/bin/apprun migrate apply            # ❌ 表已存在错误

# ✅ 正确做法
./core/bin/apprun migrate repair --execute  # 创建表
# 不需要再生成迁移文件！表已经创建好了
```

**错误2：reset 后 apply 失败**
```bash
# ❌ 问题原因
./core/bin/apprun migrate reset            # 清空数据库和历史
./core/bin/apprun migrate apply            # ❌ 没有 pending migrations

# ✅ 正确做法
./core/bin/apprun migrate reset            
cd core && atlas migrate hash --dir file://migrations  # 重新计算
./core/bin/apprun migrate apply            # ✅ 现在可以应用
```

**错误3：手动删除迁移文件后 checksum 错误**
```bash
# ❌ 错误流程
rm core/migrations/20260116_*.sql          # 手动删除
./core/bin/apprun migrate apply            # ❌ checksum mismatch

# ✅ 正确做法
./core/bin/apprun migrate clean 20260116093904  # 先清理历史
rm core/migrations/20260116_*.sql               # 再删除文件
cd core && atlas migrate hash --dir file://migrations  # 重新hash
```

### 🚨 repair vs apply 使用指南

| 场景 | 使用命令 | 原因 |
|------|---------|------|
| 空数据库初始化 | `repair --execute` | 无需迁移历史，快速创建 |
| 有迁移文件的项目 | `apply` | 保持版本化历史 |
| 数据库被手动修改 | `repair --execute` | 修复漂移 |
| 开发中快速迭代 | `sync` | 自动生成+应用 |
| 生产部署 | `apply` | 可审计的版本化迁移 |
| 迁移文件丢失 | `repair --execute` 然后重新生成 baseline | 恢复数据库结构 |

## 故障排查

### 问题1: "relation already exists"

**场景**: 使用 repair 后又 diff，apply 时报错

**原因**: 迁移文件试图创建已存在的表

**解决**: 
```bash
# 查看失败的迁移
./core/bin/apprun migrate status

# 清理失败记录
./core/bin/apprun migrate clean --all

# 删除冗余文件
rm core/migrations/20260116093904_*.sql

# 重新hash
cd core && atlas migrate hash --dir file://migrations
```

### 问题2: "no schema changes detected"

**场景**: reset 后 sync 失败

**原因**: Atlas 对比的是 schema 和 migrations 目录

**解决**: 
```bash
# 新版本已修复！sync 会先检查 pending migrations
./core/bin/apprun migrate sync  # 现在可以正常工作
```

### 问题3: "checksum mismatch"

**场景**: 迁移文件被修改

**解决**:
```bash
# 选项 1: 恢复原文件（推荐）
git checkout core/migrations/*.sql

# 选项 2: 使用 repair 修复
./core/bin/apprun migrate repair --execute
```

### 问题: Schema drift detected

**场景**: 数据库被手动修改

**解决**:
```bash
# 1. 检查差异
./core/bin/apprun migrate inspect

# 2. 决定处理方式
# 方式 A: 修复数据库（不保留修改）
./core/bin/apprun migrate repair --execute

# 方式 B: 保留修改（生成迁移）
./core/bin/apprun migrate diff capture_manual_changes
```

## 相关文档

- [完整技术说明](ATLAS-MIGRATION-MODES.md)
- [使用指南](MIGRATE-INSPECT-REPAIR.md)
- [维护策略](MIGRATIONS-MAINTENANCE.md)
