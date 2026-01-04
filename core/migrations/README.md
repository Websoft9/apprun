# Database Migrations

本目录存放数据库迁移脚本。

## 迁移脚本列表

### 001_add_configitem_audit_fields.sql
- **Story**: Story 04 - Ent Schema 配置管理
- **Date**: 2026-01-04
- **Purpose**: 为 `configitems` 表添加审计字段（status, created_at, updated_at）
- **Applies to**: 升级现有 Story 10 数据库

**执行方式**:
```bash
docker compose -f docker-compose.local.yml exec -T postgres \
  psql -U apprun -d apprun_dev < core/migrations/001_add_configitem_audit_fields.sql
```

**影响**:
- 添加 `status` 字段（默认 'active'）
- 添加 `created_at` 字段（默认当前时间）
- 添加 `updated_at` 字段（默认当前时间）
- 创建 `status` 索引
- 现有数据自动填充默认值

## 迁移策略

### 开发环境
- 使用 SQL 脚本手动迁移
- 确保 Ent Schema 与数据库结构同步

### 生产环境
- 建议使用专业迁移工具（如 golang-migrate, Flyway）
- 迁移脚本纳入版本控制
- 执行前务必备份数据库

## 命名规范

```
<序号>_<简短描述>.sql
```

示例:
- `001_add_configitem_audit_fields.sql`
- `002_add_user_roles_table.sql`
- `003_migrate_config_priorities.sql`

## 注意事项

1. **幂等性**: 脚本应使用 `IF NOT EXISTS` 确保可重复执行
2. **事务**: 使用 `BEGIN/COMMIT` 包裹变更
3. **回滚**: 考虑创建对应的 `down` 脚本
4. **测试**: 先在开发环境验证
5. **文档**: 每个脚本需注明 Story、日期、目的
