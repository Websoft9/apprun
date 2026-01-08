# Database Migrations Guide

## 概述

AppRun 使用 **Atlas** 作为数据库迁移工具，通过 Docker 执行，无需本地安装。

### 为什么选择 Atlas？
- ✅ **声明式 + 版本化**: 支持两种迁移模式
- ✅ **多数据库支持**: PostgreSQL, MySQL, SQLite, MariaDB
- ✅ **Schema 验证**: 自动检测 schema 漂移
- ✅ **回滚安全**: 自动生成回滚 SQL
- ✅ **Ent 集成**: 与 Ent ORM 完美配合

### 为什么使用 Docker？
- 🚀 **无需安装**: 避免下载 GB 级别的依赖
- 🔒 **版本锁定**: 使用 `arigaio/atlas:latest` 镜像
- 🐧 **跨平台**: Linux/macOS/Windows 统一体验
- 📦 **CI/CD 友好**: 容易集成到流水线

## 快速开始

### 前置条件
```bash
# 1. 确保 Docker 正在运行
docker --version

# 2. 确保数据库容器启动
docker ps | grep postgres

# 3. 启动开发数据库（如果未启动）
docker-compose -f docker-compose.dev.yml up -d postgres
```

### 基本工作流

```bash
# 1. 检查当前迁移状态
make migrate-status

# 2. 应用待处理的迁移
make migrate-apply

# 3. 验证迁移成功
make migrate-status
```

## 迁移命令详解

### 1. 检查迁移状态

```bash
make migrate-status
```

**输出示例**:
```
Migration Status: OK
  -- Current Version: 003
  -- Next Version: Already at latest version
  -- Executed Files: 3
  -- Pending Files: 0
```

**状态说明**:
- `OK`: 数据库与迁移文件同步
- `PENDING`: 有待应用的迁移
- `ERROR`: 发现 schema 漂移或冲突

### 2. 应用迁移

```bash
make migrate-apply
```

**示例输出**:
```
Migrating to version 003 from 002 (1 migrations in total):
  -- migrating version 003
    -> ALTER TABLE users ADD COLUMN uuid UUID NOT NULL DEFAULT gen_random_uuid();
    -> CREATE UNIQUE INDEX users_uuid_key ON users(uuid);
  -- ok (538ms)
  -- 1 migration, 5 sql statements
```

**注意事项**:
- ⚠️ 生产环境建议使用 `--dry-run` 先预览
- ⚠️ 大表迁移可能耗时较长
- ⚠️ 确保已备份数据库

### 3. 生成新迁移

```bash
# 修改 Ent schema 后生成迁移
make migrate-diff NAME=add_user_avatar
```

**工作原理**:
1. Atlas 连接数据库获取当前 schema
2. 读取 `core/ent/schema/` 中的 Ent 定义
3. 计算差异并生成 SQL
4. 创建 `core/migrations/00X_add_user_avatar.sql`

**示例输出**:
```sql
-- Modify "users" table
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(255);
```

### 4. 验证迁移文件

```bash
make migrate-validate
```

**检查项**:
- ✅ SQL 语法正确
- ✅ 迁移顺序合法
- ✅ Checksum 匹配
- ✅ 文件命名规范

### 5. Lint 检查

```bash
make migrate-lint
```

**检测内容**:
- 数据丢失风险（DROP TABLE/COLUMN）
- 并发安全问题（ADD COLUMN NOT NULL）
- 性能问题（大表 ALTER）
- 最佳实践违反

### 6. 生成 Checksum

```bash
make migrate-hash
```

**用途**:
- 生成 `core/migrations/atlas.sum`
- 防止迁移文件被篡改
- CI/CD 中验证文件完整性

### 7. 设置 Baseline（现有数据库）

```bash
make migrate-baseline
```

**使用场景**:
- 🔄 将现有数据库纳入版本控制
- 🔄 跳过已应用的历史迁移
- 🔄 团队成员加入时同步

**示例**:
```bash
# 假设数据库已有 schema，从版本 002 开始
make migrate-baseline

# 结果：001, 002 标记为已执行，003 待应用
```

## 目录结构

```
core/migrations/
├── 001_init_schema.sql          # 初始 schema
├── 002_add_auth_tables.sql      # 认证表
├── 003_add_user_uuid.sql        # UUID 字段
└── atlas.sum                    # Checksum 文件
```

**命名规范**:
- `00X_` 前缀：三位数字序号
- `descriptive_name` 主体：描述性名称
- `.sql` 后缀：SQL 文件

## Docker 实现细节

### Makefile 配置

```makefile
ATLAS_IMAGE := arigaio/atlas:latest
POSTGRES_URL := postgres://apprun:dev_password_123@host.docker.internal:5432/apprun_dev

migrate-status:
	docker run --rm \
		-v $(PWD)/core:/app \
		--add-host=host.docker.internal:host-gateway \
		$(ATLAS_IMAGE) \
		migrate status \
		--dir file:///app/migrations \
		--url "$(POSTGRES_URL)"
```

### 关键参数说明

| 参数 | 说明 |
|------|------|
| `--rm` | 运行后自动删除容器 |
| `-v $(PWD)/core:/app` | 挂载 core 目录到容器 /app |
| `--add-host=host.docker.internal:host-gateway` | 容器访问宿主机网络 |
| `--dir file:///app/migrations` | 迁移文件目录（容器内路径） |
| `--url "$(POSTGRES_URL)"` | 数据库连接 URL |

### 网络配置

```
[Host Machine]
    ↑
    | host.docker.internal:5432
    |
[Docker Container: Atlas]
    ↓
[Docker Container: PostgreSQL]
```

**为什么使用 `host.docker.internal`？**
- ✅ 跨平台兼容（Docker Desktop 自动支持）
- ✅ 简化配置（无需手动查找宿主机 IP）
- ✅ 容器可以访问宿主机上的 PostgreSQL 容器

## 常见场景

### 场景 1: 添加新字段

```bash
# 1. 修改 Ent schema
vim core/ent/schema/user.go
# 添加: field.String("phone").Optional()

# 2. 生成迁移
make migrate-diff NAME=add_user_phone

# 3. 检查生成的 SQL
cat core/migrations/004_add_user_phone.sql

# 4. 应用迁移
make migrate-apply

# 5. 验证
make migrate-status
```

### 场景 2: 创建新表

```bash
# 1. 创建 Ent schema
vim core/ent/schema/post.go

# 2. 生成迁移
make migrate-diff NAME=create_posts_table

# 3. 检查 SQL
cat core/migrations/005_create_posts_table.sql

# 4. 应用
make migrate-apply
```

### 场景 3: 现有数据库导入

```bash
# 1. 检查当前状态
make migrate-status
# 输出: PENDING (Current: No migration, Next: 001, Pending: 3)

# 2. 设置 baseline（假设 001, 002 已在 DB）
make migrate-baseline

# 3. 再次检查
make migrate-status
# 输出: PENDING (Current: 002, Next: 003, Pending: 1)

# 4. 应用新迁移
make migrate-apply
```

### 场景 4: CI/CD 集成

```yaml
# .github/workflows/test.yml
- name: Run Migrations
  run: |
    make migrate-status
    make migrate-apply
    make migrate-validate
```

## 故障排除

### 问题 1: "checksum file not found"

**原因**: 缺少 `atlas.sum` 文件

**解决**:
```bash
make migrate-hash
```

### 问题 2: "database is not clean"

**原因**: 数据库已有 schema，但未设置 baseline

**解决**:
```bash
make migrate-baseline
```

### 问题 3: "dial tcp: lookup host.docker.internal"

**原因**: Docker Desktop 未启动或版本过低

**解决**:
```bash
# 检查 Docker 版本
docker version

# 重启 Docker Desktop
# 或使用宿主机 IP 替换 host.docker.internal
```

### 问题 4: "migration checksum mismatch"

**原因**: 迁移文件被修改

**解决**:
```bash
# 重新生成 checksum
make migrate-hash

# 或回滚到正确版本
git checkout core/migrations/
```

## 最佳实践

### ✅ DO

1. **小步迭代**: 每次迁移只做一件事
2. **向前兼容**: 优先 ADD 而非 ALTER/DROP
3. **测试先行**: 在开发环境充分测试
4. **代码审查**: 迁移 SQL 必须 review
5. **备份优先**: 生产环境先备份
6. **Checksum 验证**: 提交前运行 `make migrate-hash`
7. **CI 验证**: 自动运行 `migrate-validate` 和 `migrate-lint`

### ❌ DON'T

1. **不要手动修改已应用的迁移**: 创建新迁移来修正
2. **不要在生产直接 ALTER**: 使用 online schema change
3. **不要跳过版本**: 按顺序应用
4. **不要删除迁移文件**: 即使已应用
5. **不要忽略 lint 警告**: 可能导致数据丢失

## 生产环境部署

### 部署前检查清单

- [ ] 数据库已备份
- [ ] 迁移已在 staging 测试
- [ ] 评估迁移耗时（大表）
- [ ] 准备回滚方案
- [ ] 通知相关团队
- [ ] 监控系统就绪

### 部署步骤

```bash
# 1. 连接生产数据库
export POSTGRES_URL="postgres://user:pass@prod-db:5432/apprun"

# 2. Dry-run 预览
docker run --rm \
  -v $(PWD)/core:/app \
  arigaio/atlas:latest \
  migrate apply \
  --dir file:///app/migrations \
  --url "$POSTGRES_URL" \
  --dry-run

# 3. 确认无误后执行
make migrate-apply

# 4. 验证
make migrate-status

# 5. 检查应用日志
kubectl logs -f deployment/apprun-core
```

### 回滚策略

```bash
# 方式 1: 应用回滚迁移（推荐）
# 创建 004_rollback_add_user_uuid.sql
ALTER TABLE users DROP COLUMN uuid;

make migrate-apply

# 方式 2: Atlas 回滚（需要企业版）
atlas migrate down \
  --dir file:///app/migrations \
  --url "$POSTGRES_URL" \
  --to-version 002
```

## 参考资料

- [Atlas 官方文档](https://atlasgo.io/docs)
- [Atlas Versioned Migrations](https://atlasgo.io/versioned/intro)
- [Ent Schema Migration](https://entgo.io/docs/versioned-migrations)
- [PostgreSQL Migration Best Practices](https://www.postgresql.org/docs/15/ddl-alter.html)

## 附录

### A. Atlas 配置文件

```hcl
# core/atlas.hcl
env "dev" {
  src = "ent://core/ent/schema"
  dev = "docker://postgres/15/dev"
  url = getenv("POSTGRES_URL")
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
```

### B. 支持的数据库

| 数据库 | URL 格式 | Docker 镜像 |
|--------|---------|------------|
| PostgreSQL | `postgres://user:pass@host:5432/db` | `postgres:15` |
| MySQL | `mysql://user:pass@host:3306/db` | `mysql:8` |
| SQLite | `sqlite://file.db` | N/A |
| MariaDB | `maria://user:pass@host:3306/db` | `mariadb:10` |

### C. 迁移文件模板

```sql
-- Modify "users" table
-- Purpose: Add phone number field for 2FA
-- Author: dev@apprun.com
-- Date: 2025-01-08

ALTER TABLE users 
  ADD COLUMN phone VARCHAR(20) NULL;

-- Create index for faster lookup
CREATE INDEX idx_users_phone ON users(phone) 
  WHERE phone IS NOT NULL;

-- Add comment
COMMENT ON COLUMN users.phone IS 'User phone number for 2FA, E.164 format';
```

---

**文档版本**: v1.0  
**最后更新**: 2025-01-08  
**维护者**: AppRun Dev Team
