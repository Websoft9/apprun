# Story 1.7: Makefile 标准与命令分组
# Sprint 0: Infrastructure 建设

**Priority**: P1  
**Effort**: 0.5 天 (实际: 1 天)  
**Owner**: Dev Team  
**Dependencies**: Story 1.5 (Database Migration), Story 1.6 (CLI Framework)  
**Status**: ✅ Done  
**Module**: Infrastructure  
**Issue**: #TBD  
**Completion Date**: 2026-01-16

---

## User Story

作为开发者，我希望 Makefile 命令清晰分组且遵循统一规范，以便快速找到需要的命令并避免常见错误。

---

## Acceptance Criteria

- [x] Makefile 命令按功能清晰分组（8 个核心分组）
- [x] 命令命名遵循统一规范（kebab-case，前缀一致）
- [x] 禁止直接忽略错误（`|| true` 必须有明确注释）
- [x] 避免在 Makefile 中实现复杂逻辑（超过 5 行调用外部脚本）
- [x] 每个命令提供简洁的 echo 说明
- [x] help 输出按分组显示，突出常用命令
- [x] 所有命令在 .PHONY 中声明

---

## Command Groups

### 1. Core Development (核心开发)
```makefile
# 应用相关（app-*）：仅操作 appln 二进制/进程
app-start      # 启动 apprun（二进制 bin/server，找不到则触发 build-fast）
app-stop       # 停止 appln 进程（前台/后台均可）
app-clean      # 清理 appln 产物（rm -rf bin/*、log、tmp）

# 依赖服务（deps-*）：仅操作外部依赖（db/cache/storage 等）
deps-start     # 启动依赖（默认全部或按类型，使用 docker compose）
deps-stop      # 停止依赖服务
deps-clean     # 清理依赖持久数据（卷、临时目录）

# 开发组合命令（dev-*）：便捷组合，调用上面两个组
dev-start      # 等同于 deps-start + app-start（默认全部依赖）
dev-stop       # 等同于 app-stop + deps-stop
dev-clean      # 等同于 app-clean + deps-clean
```

### 2. Build & Generate (构建与生成)
```makefile
build            # 完整构建（generate + i18n + swagger + compile）
build-fast       # 快速构建（跳过文档生成）
gen-orm          # 生成 Ent ORM 代码
gen-apidocs      # 生成 API 文档
gen-i18n         # 处理国际化
gen-configs   # 生成配置示例
```

### 3. Testing (测试)
```makefile
test             # 运行所有测试
test-unit        # 单元测试
test-integration # 集成测试
test-cover       # 生成覆盖率报告
```

### 4. Code Quality (代码质量)
```makefile
lint             # 运行 linter + govulncheck 
lint-fix         # 自动修复问题
check            # 完整质量检查（lint + test）
```

### 5. Database (数据库)

**✅ Implemented** - All database commands delegate to `apprun migrate` CLI

```makefile
db-sync          # 自动同步 schema（开发模式）
db-migrate       # 应用迁移（apprun migrate apply）
db-diff          # 生成新迁移（需要 NAME=xxx，apprun migrate diff）
db-status        # 迁移状态（apprun migrate status）
db-validate      # 验证迁移文件（apprun migrate validate）
db-rollback      # 回滚迁移（apprun migrate rollback）
db-reset         # 重置数据库（apprun migrate reset）
db-inspect       # 检查 schema drift（apprun migrate inspect）
db-repair        # 预览修复 SQL（apprun migrate repair）
db-repair-execute # 执行修复（apprun migrate repair --execute）
```

**Implementation Details:**
- All `db-*` commands are thin wrappers around `apprun migrate` CLI
- Provides consistent interface while leveraging full migration system
- See [Story 1.5: Database Migration](1-5-database-migration.md) for details

### 6. Docker (容器化)
```makefile
docker-build     # 构建镜像
docker-up        # 启动服务
docker-down      # 停止服务
docker-logs      # 查看日志
docker-clean     # 清理资源
```

### 7. Documentation (文档)
```makefile
docs-api         # 生成 API 文档（swagger 别名）
story-validate   # 验证 Story 文档
sprint-status    # 查看 Sprint 状态
story-index      # 基于 sprint-status.yaml 生成一个 story 状态索引表（spint,name,status,epic）
```

### 8. Utilities (工具)
```makefile
clean            # 清理构建产物
install          # 安装开发工具
help             # 显示帮助
```

---

## Makefile Standards

### 命名规范
- **使用 kebab-case**：`dev-start` 而非 `dev_start`
- **前缀一致性**：相关命令使用相同前缀（如 `db-*`, `docker-*`, `test-*`）
- **简洁明确**：避免冗长名称（`dev-run` 优于 `run-application-locally`）

### 输出规范
- **使用 emoji 图标**：🚀 启动、✅ 成功、❌ 错误、🔍 检查
- **分步说明**：多步骤命令显示进度
- **结果反馈**：命令结束时显示结果或下一步操作

### 错误处理
- **禁止静默失败**：不允许 `|| true` 除非有明确注释说明原因
- **显式错误提示**：失败时使用 `|| (echo "❌ Error message" && exit 1)`
- **依赖检查**：关键工具缺失时给出安装提示

### 代码组织
- **避免复杂逻辑**：超过 5 行逻辑应提取到 `scripts/` 目录
- **使用辅助函数**：通过 `.check-*` 等内部 target 复用逻辑
- **保持可读性**：每个命令前添加注释说明用途

### 脚本提取原则

- 超过 5 行的逻辑 → 提取到 `scripts/` 目录
- 多处复用的逻辑 → 封装为独立脚本
- 复杂条件判断 → 使用 Bash/Python 脚本
- 保持 Makefile 作为"命令入口"而非"逻辑实现"

### 反模式示例
```makefile
# ❌ 不好的例子
bad-target:
	some-command || true  # 静默忽略错误
	cd dir && complex-bash-logic | grep | awk | sed  # 复杂逻辑
	echo done  # 无图标、无说明

# ✅ 好的例子
good-target:
	@echo "🔍 Checking dependencies..."
	@which required-tool > /dev/null || (echo "❌ Install: brew install required-tool" && exit 1)
	@./scripts/complex-logic.sh  # 复杂逻辑放脚本
	@echo "✅ Task completed"
```

---

## Implementation Tasks

- [x] 审查现有 Makefile 命令
- [x] 按新分组重组命令
- [x] 统一命令命名（如 `migrate-*` → `db-*`）
- [x] 移除所有不必要的 `|| true`
- [x] 提取复杂逻辑到独立脚本（数据库迁移使用 `apprun migrate` CLI）
- [x] 更新 help 输出显示分组
- [x] 添加常用快捷命令（`dev-start`, `check`）
- [x] 更新 `.PHONY` 声明

---

## Test Cases

### TC1: 命令分组验证
- **操作**：运行 `make help`
- **预期**：显示清晰的命令分组，每组包含相关命令

### TC3: 错误处理验证
- **操作**：在缺少依赖工具时运行命令（如 `make lint`）
- **预期**：显示清晰的错误提示和安装说明，不静默失败

### TC4: 脚本提取验证
- **操作**：审查 Makefile 中的所有 target
- **预期**：没有超过 15 行的复杂逻辑，已提取到 `scripts/`

---

## Definition of Done

- [x] **Makefile Structure**
  - [x] All commands organized into 8 functional groups
  - [x] Command naming follows kebab-case convention
  - [x] All commands declared in .PHONY
  - [x] help target displays grouped commands

- [x] **Code Quality**
  - [x] No `|| true` without explicit comments
  - [x] Complex logic extracted to `apprun` CLI and scripts/
  - [x] All commands have emoji + description
  - [x] Error handling provides clear guidance

- [x] **Testing**
  - [x] All Makefile commands tested manually
  - [x] dev-start/dev-stop working correctly
  - [x] Dependencies start in correct order
  - [x] Error messages helpful and actionable

- [x] **Documentation & Review**
  - [x] README.md updated with new commands
  - [x] Database commands delegate to apprun CLI
  - [x] CLI integration test suite created
  - [x] Developer workflow validated

- [x] **Backward Compatibility**
  - [x] Symlink `bin/server` preserved for compatibility
  - [x] Existing workflows maintained
  - [x] CI/CD pipelines use standardized commands

---

## Technical Notes

### 向后兼容
- 保留旧命令作为别名（标记为 deprecated）
- 在弃用命令中添加迁移提示：
  ```makefile
  old-command:
  	@echo "⚠️  Deprecated: Use 'make new-command' instead"
  	@make new-command
  ```

### Help 输出优化
```makefile
help:
	@echo "AppRun BaaS - Development Commands"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make dev       - Start development environment"
	@echo "  make test      - Run all tests"
	@echo "  make check     - Run quality checks"
	@echo ""
	@echo "🔨 Development:"
	@echo "  ..."
```


---

## Related Docs

- [Makefile](../../../Makefile) - Current Makefile implementation
- [Story 1.5: Database Migration](./1-5-database-migration.md) - Migration CLI implementation
- [Story 1.6: Unified CLI Architecture](./1-6-unified-cli-architecture.md) - CLI framework
- [Story 1.4: CI/CD Pipeline](./1-4-ci-cd-linter.md) - CI/CD integration
- [CLI Reference](../../product/api/cli-reference.md) - Complete CLI documentation
- [GNU Make Manual](https://www.gnu.org/software/make/manual/)
- [Makefile Best Practices](https://tech.davis-hansson.com/p/make/)

---

## Implementation Notes

### Database Commands - apprun migrate Integration

All database commands in the Makefile are implemented as thin wrappers around the `apprun migrate` CLI:

```makefile
# Example: db-migrate delegates to apprun CLI
db-migrate:
	@cd core && ./bin/apprun migrate apply

# Example: db-diff with parameter validation
db-diff:
ifndef NAME
	$(error NAME is required. Usage: make db-diff NAME=add_project_table)
endif
	@echo "📝 Generating migration: $(NAME)..."
	@cd core && ./bin/apprun migrate diff $(NAME)
	@echo ""
	@echo "⚠️  IMPORTANT: Review the generated SQL before committing!"
```

**Key Benefits:**
- **Single Source of Truth**: Migration logic centralized in `apprun` binary
- **Consistency**: Same behavior whether using Makefile or CLI directly
- **Maintainability**: Makefile focuses on workflow, CLI handles implementation
- **Portability**: Migration logic embedded in binary, no external tools needed

**Migration Commands Implemented:**
- `db-sync` - Auto-sync schema (development)
- `db-migrate` - Apply migrations
- `db-diff` - Generate migrations
- `db-status` - Check status
- `db-validate` - Validate files
- `db-rollback` - Rollback migrations
- `db-reset` - Reset database
- `db-inspect` - Check schema drift
- `db-repair` - Repair schema drift (declarative)
- `db-repair-execute` - Execute repair

See [Story 1.5](./1-5-database-migration.md) for complete migration system documentation.

---

## Benefits

- ✅ **易于发现**：开发者快速找到需要的命令
- ✅ **减少错误**：统一规范避免常见陷阱
- ✅ **易于维护**：清晰分组便于后续扩展
- ✅ **提升体验**：友好的输出提示和错误处理
- ✅ **团队协作**：新成员快速上手

---

## References

这是一个元数据 Story，用于定义 Makefile 的组织规范，无需具体引用外部资源。

