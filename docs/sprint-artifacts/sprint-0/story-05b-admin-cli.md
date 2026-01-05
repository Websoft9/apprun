# Story 5b: 服务端 CLI 框架
# Sprint 0: Infrastructure建设

**Priority**: P1  
**Effort**: 1 天  
**Owner**: Backend Dev  
**Dependencies**: Story 1 (Docker Environment)  
**Status**: Planning  
**Module**: CLI/Admin  
**Issue**: #TBD  
**Related**: [架构文档](../../architecture/tech-architecture.md)

---

## User Story

作为运维工程师，我希望有一套服务端 CLI 工具，以便在服务器本地执行运维管理任务（如数据库迁移、备份、健康检查等）。

---

## Design Principles

1. **职责分离**：服务端 CLI 专注运维管理，不依赖 API 服务
2. **本地执行**：直接访问数据库/文件系统，无需 HTTP 调用
3. **可扩展性**：清晰的命令分组结构，便于添加新功能
4. **安全优先**：需要管理员权限，支持审计日志

---

## Acceptance Criteria

- [ ] 创建 `cmd/admin/` 目录结构
- [ ] 集成 CLI 框架（cobra 或 urfave/cli）
- [ ] 实现基础命令分组（migrate、backup、system）
- [ ] 添加全局选项（--config、--verbose、--dry-run）
- [ ] 实现命令执行日志记录
- [ ] 构建独立二进制文件 `apprun-admin`
- [ ] 编写 CLI 使用文档

---

## Implementation Tasks

- [ ] 选择并集成 CLI 框架（推荐 cobra）
- [ ] 创建 `cmd/admin/main.go` 入口
- [ ] 创建 `cmd/admin/root.go` 根命令（包含 PersistentPreRun）
- [ ] 实现全局配置加载逻辑（复用 `internal/config`）
- [ ] 实现数据库连接初始化（复用 `pkg/database`）
- [ ] 创建命令分组结构（migrate、backup、system、cache）
- [ ] 实现统一错误处理和输出格式
- [ ] 添加 Makefile 构建目标
- [ ] 更新 `docs/product/administration/cli-reference.md`

---

## Technical Details

### CLI 框架选择

**推荐：Cobra**
- 成熟稳定（Kubernetes、Docker 使用）
- 自动生成帮助文档
- 支持子命令嵌套
- 丰富的生态系统

**安装：**
```bash
go get -u github.com/spf13/cobra@latest
```

---

### 目录结构

```
core/
├── cmd/
│   ├── server/          # HTTP 服务器
│   └── admin/           # 服务端 CLI（新增）
│       ├── main.go      # CLI 入口
│       ├── root.go      # 根命令
│       ├── migrate/     # 数据库迁移命令组
│       ├── backup/      # 备份恢复命令组
│       ├── system/      # 系统管理命令组
│       └── cache/       # 缓存管理命令组
```

---

### 命令分组设计

**1. migrate（数据库迁移）**
- `apprun-admin migrate up` - 应用待迁移
- `apprun-admin migrate down` - 回滚迁移
- `apprun-admin migrate status` - 查看迁移状态
- `apprun-admin migrate generate <name>` - 生成迁移文件

> 说明：migrate 子命令不实现迁移逻辑本身，只是调用核心迁移模块（见 Story 5a）。

**2. backup（备份恢复）**
- `apprun-admin backup create` - 创建数据库备份
- `apprun-admin backup restore <file>` - 恢复备份
- `apprun-admin backup list` - 列出备份文件

**3. system（系统管理）**
- `apprun-admin system healthcheck` - 系统健康检查
- `apprun-admin system info` - 显示系统信息
- `apprun-admin system version` - 显示版本信息

**4. cache（缓存管理）**
- `apprun-admin cache clear` - 清理缓存
- `apprun-admin cache stats` - 缓存统计信息

---

### 全局选项

**通用参数：**
- `--config <path>` - 指定配置文件路径（默认：`./config/default.yaml`）
- `--verbose` / `-v` - 详细输出模式
- `--dry-run` - 模拟执行，不实际操作
- `--log-file <path>` - 指定日志文件路径

**环境变量：**
- 支持通过环境变量覆盖配置
- 命令行参数优先级最高

---

### 数据库连接管理

**配置加载策略：**
- 复用 `config/default.yaml`（与 HTTP 服务器共享配置源）
- 支持 `--config` 指定自定义配置路径
- 环境变量 `DATABASE_*` 优先级最高（便于容器环境）

**连接初始化：**
- 在根命令（`root.go`）的 `PersistentPreRun` 钩子中建立连接
- 使用 `pkg/database/Connect()` 复用现有数据库逻辑
- 连接失败则提前退出，显示友好错误信息
- 所有子命令自动继承已初始化的连接

**子命令访问模式：**
- 通过全局变量或 Cobra Context 共享数据库连接
- 子命令无需显式传递连接参数
- 专注业务逻辑实现，降低耦合

**优势：**
- 配置统一管理（避免重复）
- 连接池复用（性能优化）
- 易于测试（可 mock 全局连接）

---

### 构建和部署

**Makefile 目标：**
```makefile
build-admin:
    go build -o bin/apprun-admin ./cmd/admin

install-admin:
    go install ./cmd/admin
```

**Docker 集成：**
- 二进制文件内置在容器中
- Entrypoint 脚本可调用 CLI 命令
- 示例：`docker exec apprun apprun-admin migrate up`

**独立分发：**
- 编译为静态二进制文件
- 支持多平台（Linux、macOS）
- 可通过包管理器分发（未来）

---

## Test Cases

- [ ] 帮助文档正确显示（`--help`）
- [ ] 全局选项正常工作（`--config`、`--verbose`）
- [ ] 数据库连接成功初始化
- [ ] 连接失败时友好错误提示
- [ ] 命令分组结构清晰
- [ ] 错误处理友好（无效命令提示）
- [ ] 日志记录正确输出

---

## Future Enhancements

**Sprint 1+：**
- 添加用户管理命令（`user create`、`user reset-password`）
- 集成配置管理命令（`config export`、`config import`）
- 支持交互式模式（prompt 输入）
- 命令自动补全（bash/zsh）

**与客户端 CLI 的关系：**
- 客户端 CLI（`apprun`）：通过 API 远程调用
- 服务端 CLI（`apprun-admin`）：本地直接操作
- 两者功能互补，不重叠

---

## Related Stories

- **Story 5a**：数据库增量迁移（CLI 将作为调用入口之一）
- **Future Story**：客户端 CLI 框架（API 调用封装）

---

## Related Docs

- [Cobra Documentation](https://cobra.dev/)
- [CLI 设计最佳实践](https://clig.dev/)
- [技术架构](../../architecture/tech-architecture.md)

---

## Notes

- 服务端 CLI 主要用于运维场景，不是用户日常工具
- 需要高权限（数据库访问、文件系统操作）
- 未来可扩展为容器内调试工具（troubleshooting）
- 迁移能力属于核心模块（Story 5a）；本 CLI 仅负责提供运维入口（命令封装）
- 数据库连接通过 `PersistentPreRun` 全局初始化，子命令无需关心连接管理
- 配置与 HTTP 服务器共享，确保一致性

---

**Created**: 2026-01-05  
**Updated**: 2026-01-05  
**Maintainer**: PM Agent (John)
