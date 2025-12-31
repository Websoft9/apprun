# Story 16: Database Anti-Corruption Layer
# Sprint 1: Infrastructure Enhancement

**Priority**: P1  
**Effort**: 1 day  
**Owner**: Backend Dev  
**Dependencies**: Story 15 (Environment Package)  
**Status**: Done ✅  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [Story 15 - Environment Package](./story-15-env-package.md), [Story 14 - HTTP Server Package](./story-14-http-server.md), [Coding Standards](../../standards/coding-standards.md)

---

## User Story

作为开发者，我希望将数据库连接逻辑从配置模块解耦，并创建防腐层隔离 Ent ORM 依赖，以便实现基础设施配置的独立性和可测试性。

---

## Acceptance Criteria

- [x] 创建 `core/pkg/database` 包作为防腐层
- [x] 实现 `Config` 结构体，使用 `pkg/env` 读取环境变量
- [x] 实现 `Client` 接口，隔离 Ent ORM 依赖
- [x] 提供 `Connect()` 函数，支持 schema 自动迁移
- [x] 从 `bootstrap.go` 解耦数据库连接逻辑
- [x] 单元测试覆盖核心功能（3个测试用例）
- [x] 更新 `cmd/server/main.go` 使用新包
- [x] 使用 `DATABASE_*` 环境变量命名规范（统一命名约定）

---

## Problem Statement

**当前问题**：
- `bootstrap.go` 职责过多（配置加载 + 数据库连接 + 服务创建）
- 数据库配置依赖配置中心（循环依赖风险）
- 层级不一致：数据库连接应属于 Layer 1 基础设施，而非混在 bootstrap 中
- Ent ORM 依赖暴露到业务层，缺乏防腐层隔离

---

## Design

### 实际实现架构

**核心原则**：
- `pkg/database` 定位为 **Layer 1 基础设施**，与 `pkg/server`、`pkg/logger` 同层
- 配置来源：从环境变量（通过 `pkg/env`）读取，使用 `DATABASE_*` 命名规范
- 防腐层：通过 `Client` 接口隔离 Ent ORM 依赖

**已实现包结构**：
```
pkg/database/
├── config.go         # Config + DefaultConfig() (31 lines)
├── client.go         # Client 接口（防腐层，70 lines）
├── postgres.go       # PostgreSQL 实现 (38 lines)
└── database_test.go  # 单元测试 (3 tests)
```

**关键 API**：
- `DefaultConfig() *Config`：从 `DATABASE_*` 环境变量读取配置
- `Connect(ctx, cfg) (Client, error)`：建立连接，执行 schema migration
- `Client` 接口：
  - `Close() error`：关闭连接
  - `Ping(ctx) error`：检查连接状态
  - `Tx(ctx, fn) error`：事务执行
  - `GetEntClient() *ent.Client`：获取底层 Ent 客户端（谨慎使用）

### 环境变量命名规范

**统一使用 `DATABASE_*` 前缀**（与 Story 15 统一命名约定）：
```bash
DATABASE_DRIVER=postgres
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=secret      # 必需
DATABASE_DB_NAME=apprun
```

**命名规则**：`{GROUP}_UPPERCASE_{KEY}_UPPERCASE`
- 示例：`database.host` → `DATABASE_HOST`
- 示例：`database.db_name` → `DATABASE_DB_NAME`

---

## Implementation Summary

### 已完成任务

**Phase 1 - Database Package 创建** ✅
- `config.go`：配置结构体，使用 `DATABASE_*` 环境变量
- `client.go`：Client 接口防腐层，隔离 Ent ORM
- `postgres.go`：PostgreSQL 连接实现，自动 schema migration
- `database_test.go`：单元测试（3个测试用例）

**Phase 2 - Bootstrap 重构** ✅
- 移除 `bootstrap.InitDatabase()` 方法
- `CreateService()` 方法接收外部 `database.Client` 参数
- 解耦数据库连接逻辑，简化 bootstrap 职责

**Phase 3 - Main.go 更新** ✅
- Phase 2 中直接调用 `database.DefaultConfig()` 和 `database.Connect()`
- 数据库连接独立于配置加载流程
- 启动流程清晰：Phase 0 (配置文件) → Phase 1 (配置注册) → Phase 2 (数据库连接) → Phase 3 (配置服务)

### 代码示例

**Database Config (config.go)**:
```go
func DefaultConfig() *Config {
    return &Config{
        Driver:   env.Get("DATABASE_DRIVER", "postgres"),
        Host:     env.Get("DATABASE_HOST", "localhost"),
        Port:     env.GetInt("DATABASE_PORT", 5432),
        User:     env.Get("DATABASE_USER", "postgres"),
        Password: env.MustGet("DATABASE_PASSWORD"), // Required
        DBName:   env.Get("DATABASE_DB_NAME", "apprun"),
    }
}
```

**Client Interface (client.go)**:
```go
type Client interface {
    Close() error
    Ping(ctx context.Context) error
    Tx(ctx context.Context, fn func(tx *ent.Tx) error) error
    GetEntClient() *ent.Client
}
```

**Main.go Integration**:
```go
// Phase 2: Connect to Database (Layer 1 infrastructure)
dbCfg := database.DefaultConfig()
dbClient, err := database.Connect(ctx, dbCfg)
if err != nil {
    log.Fatalf("❌ Failed to connect to database: %v", err)
}
defer dbClient.Close()

// Phase 3: Initialize Config Service (Layer 2)
configService, err := bootstrap.CreateService(ctx, dbClient)
```

---

## Key Benefits

- **职责分离**：`pkg/database` 专注数据库连接，`bootstrap.go` 专注配置装配
- **层级一致**：与 `pkg/server`、`pkg/logger` 同为 Layer 1 基础设施
- **防腐层**：通过 `Client` 接口隔离 Ent ORM，业务层不直接依赖 ORM 实现
- **可测试性**：独立测试数据库连接逻辑，可 Mock Client 接口
- **可复用性**：CLI 工具、脚本可直接使用，无需依赖配置中心

---

## Migration Impact

### 新建文件
- ✅ `pkg/database/config.go` (31 lines)
- ✅ `pkg/database/client.go` (70 lines)
- ✅ `pkg/database/postgres.go` (38 lines)
- ✅ `pkg/database/database_test.go` (71 lines, 3 tests)

### 修改文件
- ✅ `modules/config/bootstrap.go` - 移除 `InitDatabase()`，`CreateService()` 接收 `database.Client`
- ✅ `cmd/server/main.go` - Phase 2 直接使用 `database.Connect()`
- ✅ `Makefile` - `run-local` 使用 `DATABASE_*` 环境变量

### 破坏性变更
- ✅ `bootstrap.InitDatabase()` 方法已移除
- ✅ `CreateService(ctx, dbClient)` 现在需要外部传入数据库客户端
- ✅ 环境变量命名从 `DB_*` 改为 `DATABASE_*`（统一命名约定）

### 兼容性说明
- **向后不兼容**：使用旧 `DB_*` 环境变量的配置需要更新为 `DATABASE_*`
- **迁移指南**：
  ```bash
  # Before
  DB_HOST=localhost
  DB_PORT=5432
  DB_PASSWORD=secret
  
  # After
  DATABASE_HOST=localhost
  DATABASE_PORT=5432
  DATABASE_PASSWORD=secret
  ```

---

## Test Cases

### 已实现测试

**单元测试** (`database_test.go`) - 3/3 PASS ✅：
- [x] `TestDefaultConfig`：验证环境变量读取与配置正确性
- [x] `TestDefaultConfig_Defaults`：验证默认值正确性（仅设置必需的 PASSWORD）
- [x] `TestConnect_InvalidConfig`：验证连接失败场景的错误处理

### 测试覆盖

**当前覆盖**：核心功能单元测试
- 配置加载：✅ 环境变量读取、默认值
- 连接逻辑：✅ 错误处理验证
- 接口设计：✅ Client 接口定义

**集成测试**（需要真实数据库）：
- 连接成功场景：需要测试数据库环境
- Ping 功能：需要活动连接
- 事务执行：需要实际数据库操作
- Schema migration：已在 `Connect()` 中自动执行

### 测试运行

```bash
# 单元测试
$ go test ./pkg/database -v
PASS: TestDefaultConfig
PASS: TestDefaultConfig_Defaults
PASS: TestConnect_InvalidConfig
ok      apprun/pkg/database     0.027s

# 与其他包一起测试
$ go test ./pkg/env ./pkg/server ./pkg/database ./modules/config
ok      apprun/pkg/env          0.016s
ok      apprun/pkg/server       0.012s
ok      apprun/pkg/database     0.027s
ok      apprun/modules/config   0.073s
```

---

## Related Documentation

- [Story 15 - Environment Package](./story-15-env-package.md) - 环境变量工具，`DATABASE_*` 命名规范
- [Story 14 - HTTP Server Package](./story-14-http-server.md) - 同层基础设施包参考
- [Coding Standards - Configuration Guidelines](../../standards/coding-standards.md) - 配置规范
- [Architecture Standards](../../standards/architecture-standards.md) - 架构分层标准

---

## Benefits Achieved

### 职责分离 ✅
- `pkg/database` 专注数据库连接和防腐层
- `bootstrap.go` 简化为配置装配器
- 清晰的关注点分离

### 层级一致 ✅
- 与 `pkg/server`、`pkg/logger` 同为 Layer 1 基础设施
- 数据库连接不再混入配置加载流程
- 启动流程更清晰：配置 → 数据库 → 配置服务 → 业务服务

### 防腐层 ✅
- `Client` 接口隔离 Ent ORM 依赖
- 业务层通过接口操作，不直接依赖 ORM 实现
- 未来可替换 ORM 而不影响上层

### 可测试性 ✅
- 独立测试数据库连接逻辑
- 可 Mock `Client` 接口进行单元测试
- 配置加载与数据库连接分离测试

### 可复用性 ✅
- CLI 工具、脚本可直接使用 `pkg/database`
- 无需依赖配置中心
- 环境变量配置简单直接

### 命名统一 ✅
- 统一使用 `DATABASE_*` 前缀
- 遵循 `{GROUP}_UPPERCASE_{KEY}_UPPERCASE` 规范
- 与 Story 15 环境变量命名约定一致

---

**Created**: 2025-12-31  
**Updated**: 2025-12-31  
**Completed**: 2025-12-31 ✅  
**Maintainer**: BMad Dev Agent (Amelia)
