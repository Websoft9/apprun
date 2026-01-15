# Story 1.6.1: Admin Management Commands (管理命令集)
# Sprint 0→3: Infrastructure Enhancement

**Priority**: P2  
**Effort**: 2-3 天  
**Owner**: Backend Dev  
**Dependencies**: Story 1.6 (Unified CLI Architecture) - **REQUIRED**  
**Status**: 📋 **Ready for Dev** (依赖Story 1.6的CLI框架)  
**Module**: CLI/Admin  
**Issue**: #TBD  
**Related**: [Story 1.6 - Unified CLI Architecture](./1-6-unified-cli-architecture.md)

---

## 🔄 Story Relationship with 1.6 (2026-01-15)

**This story is a CONTINUATION of [Story 1.6: Unified CLI Architecture](./1-6-unified-cli-architecture.md)**

### Scope Division:

| Story | Scope | Status |
|-------|-------|--------|
| **Story 1.6** | CLI 框架基础 + `migrate` + `serve` + `version` | ✅ **Completed** |
| **Story 1.6.1** | 管理命令实施：`backup` + `cache` + `system` | 📋 **Ready for Dev** |

### Why Keep Story 1.6.1?

1. **功能互补** - Story 1.6 完成了CLI框架，Story 1.6.1 负责具体管理命令实现
2. **明确分工** - 1.6聚焦架构（已完成），1.6.1聚焦运维功能（待实施）
3. **渐进交付** - 避免单个Story过于庞大（1.6已3-4天，1.6.1再2-3天）

### Implementation Constraint:
- ✅ **必须使用** Story 1.6 的 Cobra 框架
- ✅ **必须遵循** 扁平化 `cmd/` 目录结构
- ✅ **命令格式**: `apprun <command>` (不是 `apprun-admin`)

---

## User Story (Updated)

---

## User Story (Original - For Reference Only)

作为运维工程师，我希望通过 `apprun` CLI 执行平台管理任务（备份恢复、缓存管理、系统诊断），以便在无法访问Web界面时完成运维工作。

---

## Scope (Updated for Story 1.6 Integration)

**本Story专注实现以下管理命令**（基于Story 1.6的CLI框架）：

### 1. backup - 备份管理命令组 ✨
```bash
apprun backup create [--output <path>]     # 创建数据库备份
apprun backup restore <file>               # 恢复备份
apprun backup list [--limit 10]            # 列出备份文件
```

### 2. cache - 缓存管理命令组 ✨
```bash
apprun cache clear [--pattern <key>]       # 清理缓存（支持通配符）
apprun cache stats                         # 缓存统计信息
apprun cache get <key>                     # 查看缓存值（调试用）
```

### 3. system - 系统诊断命令组 ✨
```bash
apprun system healthcheck                  # 完整健康检查（DB/Redis/Disk）
apprun system info                         # 系统信息（版本/配置/资源）
apprun system diagnose                     # 诊断常见问题
```

**不包含的命令**（已在Story 1.6实现）：
- ❌ `apprun migrate` - 已在Story 1.6实现
- ❌ `apprun serve` - 已在Story 1.6实现
- ❌ `apprun version` - 已在Story 1.6实现

---

## Acceptance Criteria

### AC-001: Backup Commands ✓
- [ ] `apprun backup create` 创建PostgreSQL备份（pg_dump格式）
- [ ] `apprun backup restore <file>` 恢复备份（带确认提示）
- [ ] `apprun backup list` 列出备份目录的所有备份文件
- [ ] 备份文件命名规范：`backup-YYYYMMDD-HHmmss.sql`
- [ ] 支持 `--output` 指定备份路径（默认：`./backups/`）

### AC-002: Cache Commands ✓
- [ ] `apprun cache clear` 清理所有缓存（需确认）
- [ ] `apprun cache clear --pattern "user:*"` 清理匹配模式的缓存
- [ ] `apprun cache stats` 显示Redis内存使用、命中率、键数量
- [ ] `apprun cache get <key>` 查看缓存值（调试用）
- [ ] 集成pkg/cache包（Story 1.16）

### AC-003: System Commands ✓
- [ ] `apprun system healthcheck` 检查：数据库连接、Redis连接、磁盘空间
- [ ] `apprun system info` 显示：Go版本、构建信息、配置路径、进程PID
- [ ] `apprun system diagnose` 诊断：端口占用、配置问题、依赖可达性
- [ ] 输出格式友好（彩色、对齐、状态图标）

### AC-004: Implementation Standards ✓
- [ ] 所有命令在 `core/cmd/` 目录下（backup.go, cache.go, system.go）
- [ ] 遵循Story 1.6的扁平化Cobra结构
- [ ] 使用 `pkg/database`, `pkg/cache` 等基础设施包
- [ ] 错误处理使用 `pkg/errors` 标准错误码
- [ ] 日志记录使用 `pkg/logger` 结构化日志

---

## Implementation Tasks

### Phase 1: Backup Commands (1 天)
- [ ] 创建 `core/cmd/backup.go` - backup 根命令
- [ ] 实现 `backup_create.go` - 调用 pg_dump
- [ ] 实现 `backup_restore.go` - 调用 psql
- [ ] 实现 `backup_list.go` - 列出备份目录
- [ ] 创建 `core/pkg/backup/manager.go` - 备份管理逻辑
- [ ] 单元测试：`backup_test.go`（mock文件系统）

### Phase 2: Cache Commands (0.5 天)
- [ ] 创建 `core/cmd/cache.go` - cache 根命令
- [ ] 实现 `cache_clear.go` - 调用 pkg/cache.FlushDB()
- [ ] 实现 `cache_stats.go` - 调用 pkg/cache.Info()
- [ ] 实现 `cache_get.go` - 调用 pkg/cache.Get()
- [ ] 单元测试：`cache_test.go`（使用testcontainers）

### Phase 3: System Commands (1 天)
- [ ] 创建 `core/cmd/system.go` - system 根命令
- [ ] 实现 `system_healthcheck.go` - 检查DB/Redis/Disk
- [ ] 实现 `system_info.go` - 显示系统信息
- [ ] 实现 `system_diagnose.go` - 诊断常见问题
- [ ] 创建 `core/pkg/system/diagnostic.go` - 诊断逻辑
- [ ] 单元测试：`system_test.go`

### Phase 4: Testing & Documentation (0.5 天)
- [ ] 集成测试：完整命令执行流程
- [ ] 更新 `docs/cli-reference.md` 添加新命令
- [ ] 更新 README.md 添加管理命令示例
- [ ] 创建运维手册：`docs/operations/cli-management.md`

---

## Technical Details

### Backup Implementation

使用 `os/exec` 调用 PostgreSQL 工具：
```go
// core/pkg/backup/manager.go
package backup

import (
    "context"
    "fmt"
    "os/exec"
    "path/filepath"
    "time"
)

type Manager struct {
    dbHost     string
    dbPort     string
    dbUser     string
    dbPassword string
    dbName     string
    backupDir  string
}

func (m *Manager) Create(ctx context.Context) (string, error) {
    timestamp := time.Now().Format("20060102-150405")
    filename := fmt.Sprintf("backup-%s.sql", timestamp)
    filepath := filepath.Join(m.backupDir, filename)
    
    cmd := exec.CommandContext(ctx, "pg_dump",
        "-h", m.dbHost,
        "-p", m.dbPort,
        "-U", m.dbUser,
        "-F", "c", // Custom format (compressed)
        "-f", filepath,
        m.dbName,
    )
    cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", m.dbPassword))
    
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("pg_dump failed: %w", err)
    }
    
    return filepath, nil
}
```

### Cache Implementation

复用 Story 1.16 的 `pkg/cache` 包：

```go
// core/cmd/cache_clear.go
package cmd

import (
    "github.com/spf13/cobra"
    "apprun/pkg/cache"
    "apprun/pkg/logger"
)

var cacheClearCmd = &cobra.Command{
    Use:   "clear",
    Short: "Clear cache entries",
    RunE:  runCacheClear,
}

func init() {
    cacheCmd.AddCommand(cacheClearCmd)
    cacheClearCmd.Flags().String("pattern", "", "Key pattern to clear (e.g., 'user:*')")
}

func runCacheClear(cmd *cobra.Command, args []string) error {
    log := logger.L()
    pattern, _ := cmd.Flags().GetString("pattern")
    
    cacheClient, err := cache.NewClient(&cache.Config{
        // Load from environment
    })
    if err != nil {
        return fmt.Errorf("failed to connect to cache: %w", err)
    }
    defer cacheClient.Close()
    
    if pattern != "" {
        count, err := cacheClient.DeletePattern(pattern)
        log.Info("Cache cleared by pattern", 
            logger.Field{"pattern", pattern}, 
            logger.Field{"count", count})
    } else {
        if err := cacheClient.FlushDB(); err != nil {
            return fmt.Errorf("failed to flush cache: %w", err)
        }
        log.Info("All cache cleared")
    }
    
    return nil
}
```

### System Healthcheck Implementation

```go
// core/pkg/system/diagnostic.go
package system

import (
    "context"
    "fmt"
    "apprun/pkg/database"
    "apprun/pkg/cache"
)

type HealthStatus struct {
    Database bool   `json:"database"`
    Cache    bool   `json:"cache"`
    Disk     bool   `json:"disk"`
    Message  string `json:"message,omitempty"`
}

func (d *Diagnostic) HealthCheck(ctx context.Context) (*HealthStatus, error) {
    status := &HealthStatus{}
    
    // Check database
    if err := d.db.Ping(ctx); err == nil {
        status.Database = true
    }
    
    // Check cache
    if err := d.cache.Ping(ctx); err == nil {
        status.Cache = true
    }
    
    // Check disk space
    if diskFree() > 1*GB {
        status.Disk = true
    }
    
    if status.Database && status.Cache && status.Disk {
        status.Message = "All systems operational"
    } else {
        status.Message = "Some systems require attention"
    }
    
    return status, nil
}
```

---

## Definition of Done

- [ ] **Code Implementation**
  - [ ] 3 command files created (backup.go, cache.go, system.go)
  - [ ] All subcommands implemented with proper error handling
  - [ ] Integration with pkg/database, pkg/cache, pkg/logger
  - [ ] Follows Story 1.6 Cobra patterns

- [ ] **Testing**
  - [ ] Unit tests for all commands (>80% coverage)
  - [ ] Integration tests with real database/redis (testcontainers)
  - [ ] Manual testing on dev environment
  - [ ] All tests passing: `go test ./cmd/... -v`

- [ ] **Documentation**
  - [ ] docs/cli-reference.md updated with new commands
  - [ ] README.md updated with management command examples
  - [ ] docs/operations/cli-management.md created (运维手册)
  - [ ] Godoc comments for all public functions

- [ ] **Code Quality**
  - [ ] golangci-lint passes (zero warnings)
  - [ ] Code reviewed and approved (2 reviewers)
  - [ ] Consistent with Story 1.6 patterns
  - [ ] Error messages user-friendly

- [ ] **Operational Readiness**
  - [ ] Commands tested in Docker container
  - [ ] Makefile targets added (if needed)
  - [ ] No breaking changes to existing CLI
  - [ ] Deployed to dev environment successfully

---

## Test Cases

### Backup Commands
- [ ] `apprun backup create` 创建备份成功（文件存在）
- [ ] 备份文件命名符合 `backup-YYYYMMDD-HHmmss.sql` 格式
- [ ] `apprun backup list` 正确列出所有备份文件
- [ ] `apprun backup restore <file>` 恢复备份成功（需确认提示）
- [ ] 缺少 pg_dump 工具时给出友好错误提示

### Cache Commands
- [ ] `apprun cache clear` 清理所有缓存（需确认）
- [ ] `apprun cache clear --pattern "test:*"` 清理匹配键
- [ ] `apprun cache stats` 显示内存使用、命中率、键数量
- [ ] `apprun cache get <key>` 正确返回缓存值
- [ ] Redis连接失败时给出友好错误提示

### System Commands
- [ ] `apprun system healthcheck` 显示所有组件状态（✅/❌图标）
- [ ] `apprun system info` 显示Go版本、构建信息、配置路径
- [ ] `apprun system diagnose` 诊断并给出修复建议
- [ ] 输出格式美观（彩色、对齐、表格）

---

## Dependencies

### External Tools
- `pg_dump` / `pg_restore` - PostgreSQL备份工具（需预安装）
- Redis server - 缓存服务（Story 1.16）

### Internal Dependencies
- ✅ **Story 1.6** - Unified CLI Architecture（CLI框架基础）
- ✅ **Story 1.16** - Redis Cache Package（cache命令依赖）
- ✅ **Story 1.14** - Database Package（backup/system命令依赖）
- ✅ **Story 1.10** - Logger Package（结构化日志）
- ✅ **Story 1.3** - Error Handling（错误处理）

---

## Related Stories

- **Story 1.6** - Unified CLI Architecture (依赖：必须先完成)
- **Story 1.16** - Redis Cache Package (依赖：cache命令需要)
- **Story 1.5** - Database Migration (相关：migrate命令已在1.6实现)
- **Story 1.17** - Platform Initialization (相关：init命令)

---

## Makefile Integration (Optional)

如需要，可添加便捷目标：

```makefile
# Backup shortcuts
backup-create:
	@bin/apprun backup create

backup-restore:
	@bin/apprun backup restore $(FILE)

# Cache shortcuts
cache-clear:
	@bin/apprun cache clear --pattern "$(PATTERN)"

# System diagnostics
healthcheck:
	@bin/apprun system healthcheck
```

---

## Related Docs

- [Story 1.6 - Unified CLI Architecture](./1-6-unified-cli-architecture.md)
- [Cobra Documentation](https://cobra.dev/)
- [PostgreSQL Backup Documentation](https://www.postgresql.org/docs/current/backup.html)
- [运维最佳实践](../../standards/operations-best-practices.md)

---

**Created**: 2026-01-05  
**Updated**: 2026-01-15  
**Maintainer**: Scrum Master (Bob)

**Story Relationship**: 本Story是Story 1.6的功能扩展，专注于实现管理命令（backup/cache/system），必须基于Story 1.6的CLI框架实现。