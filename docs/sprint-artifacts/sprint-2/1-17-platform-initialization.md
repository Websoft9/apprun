# Story 1.17: Platform Initialization (平台初始化)
# Sprint 2: Infrastructure Enhancement

**Priority**: P0  
**Effort**: 2 天  
**Owner**: Backend Dev + DevOps  
**Dependencies**: Story 1.5 (Database Migration), Story 1.6 (CLI Framework), Story 5.x (Auth)  
**Status**: 📋 Ready for Dev  
**Module**: Infrastructure / Bootstrap  
**Related**: [Architecture](../../architecture/)  
**Blocked By**: Story 1.6 (需要 CLI 框架支持 `apprun init` 命令)

---

## User Story

作为**运维人员**，我希望通过单一命令完成平台首次部署的初始化工作（数据库迁移、管理员账号创建、默认配置导入、平台项目创建），使平台能快速投入使用。

作为**开发者**，我希望平台启动时能自动检测初始化状态，在未初始化时提供友好提示或自动完成初始化（开发环境）。

---

## 关于初始化

软件初始化不仅仅是“启动时跑个脚本”，而是一个多阶段、多场景的生命周期管理过程。它涉及数据库表结构、数据导入、用户创建等核心操作，这些操作需要在不同启动阶段（如首次启动、后续重启、版本升级）中智能适配，以避免数据不一致、性能瓶颈或安全风险。

### 场景分类

初始化场景可以分为以下几类（基于实际项目经验，我会结合“什么能工作”与“什么能扩展”来分类）：

1. 首次启动（First-Time Setup）：

- 场景描述：软件安装后第一次运行。此时系统是“空白状态”，需要从零开始创建基础环境。
- 典型操作：导入数据库表结构、导入种子数据（seed data，如默认配置或示例数据）、创建初始用户（admin用户或系统用户）。
- 我的理解：这是“奠基”阶段，强调自动化和无干预。用户期望“一键安装”，但架构师需要考虑云环境下的弹性伸缩

2. 后续启动/重启（Routine Restart）：

- 场景描述：系统已初始化过，只是重启或容器重启。此时数据库已存在，但需要验证完整性。
- 典型操作：检查表结构是否存在、验证数据完整性、执行轻量级健康检查（health checks），可能导入增量数据或更新缓存。
- 我的理解：这是“维护”阶段，重点是快速恢复和容错。重启不应重新导入所有数据（避免浪费时间），而是增量式。如果是分布式系统中，还涉及服务发现和状态同步。

3. 版本升级启动（Version Upgrade）：

- 场景描述：软件版本更新后首次启动。可能涉及数据库迁移（schema changes）、数据转换或新功能的启用。
- 典型操作：检测版本差异、运行迁移脚本（migrations）、导入新数据、更新用户权限或角色。
- 我的理解：这是“演进”阶段，最复杂。架构师需要设计“向后兼容”的迁移策略，避免破坏现有数据。

4. 其他边缘场景：

- 多环境部署：如开发/测试/生产环境，初始化逻辑需环境感知（environment-aware），避免生产数据被测试数据覆盖。
- 灾难恢复：从备份恢复时，初始化需重建状态。
- 横向扩展：新实例加入集群时，初始化需同步状态，而非重新创建。

> 这些场景不是孤立的——好的架构会将它们抽象成一个“初始化框架”，使用状态机或工作流来管理（例如，基于事件驱动的初始化管道）。

### 架构原则

1. apprun作为产品包（二进制或容器），初始化应完全内置于代码中，不依赖外部工具如Ansible。这意味着初始化逻辑需嵌入Go代码，通过启动时的钩子或服务来执行。  

2. 初始化模块，使用状态机管理不同阶段（首次/重启/升级）
3. 数据一致性和完整性
4. 失败时提供清晰日志和恢复选项（e.g., 重试机制或手动干预）。避免部分初始化导致系统不稳定。升级失败，考虑有自动回滚到上一版本.
5. 设计迁移链（migration chains），支持跳跃升级（e.g., 从v1直接到v3）
6. 断路器（circuit breaker）模式进行容错设计



## Acceptance Criteria

### AC-001: 手动初始化命令（生产环境）
- [x] `make init` 执行完整初始化：数据库迁移 → 创建超级管理员账号 → 创建平台项目
- [x] 幂等性保证：多次执行 `make init` 安全（已存在则跳过）
- [x] 初始化完成后输出管理员凭据和下一步操作提示（随机密码时显示在控制台框内）

### AC-002: 超级管理员配置 (InitConfig)
- [x] 支持配置文件 `auth.init` 配置项：username, email, password
- [x] **随机密码生成**：password 为空时自动生成 16 位安全密码（包含大小写、数字、特殊字符）
- [x] 生成的密码在控制台美化输出（带边框警告信息）
- [x] 密码使用 bcrypt 加密存储
- [x] **保留用户名支持**：超级管理员初始化时可使用 admin 等保留用户名

### AC-003: 平台项目初始化
- [x] 创建 Platform Project（固定 UUID: `00000000-0000-0000-0000-000000000000`）
- [x] 超级管理员自动成为 Platform Project owner
- [x] **分配平台管理员角色**：超级管理员自动获得 `platform_admin` 全局角色
- [x] 通过 RBAC 系统赋予所有权限（`p, platform_admin, *, *`）

### AC-004: 启动时自动检测（程序内）
- [x] `bootstrap/server.go` 启动时检查超级管理员是否存在
- [x] 已存在 → 继续启动
- [x] 不存在 → 自动创建（使用配置文件设置）
- [x] 创建失败时记录错误但不阻止启动（warning level）

---

## Implementation Tasks

### Task 1: Bootstrap Package（核心逻辑）
- [ ] 创建 `core/pkg/bootstrap/initializer.go`
  - [ ] `CheckInitialized(ctx)` - 查询 `platform_meta.initialized` 状态
  - [ ] `Initialize(ctx, config)` - 执行初始化流程（事务）
    - Create system user (ID=1, username="system")
    - Create admin account (from config)
    - Create platform project with system user as owner
    - Insert `platform_meta.initialized = true`
  - [ ] `Config` 结构体（AdminEmail, AdminPassword, AutoInit）

### Task 2: Migration File（数据库结构）
- [ ] 创建 `migrations/00X_platform_meta.sql`
  ```sql
  CREATE TABLE platform_meta (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```

### Task 3: CLI Command（Cobra 子命令）
- [ ] 创建 `core/cmd/cli/init.go` - Cobra 子命令定义
  - [ ] 实现 `initCmd` 命令，读取环境变量并调用 `bootstrap.Initialize()`
  - [ ] 添加 flags：`--admin-email`, `--admin-password`, `--skip-confirmation`
  - [ ] 在 `init()` 函数中注册到 rootCmd：`rootCmd.AddCommand(initCmd)`
- [ ] 添加 `Makefile` target:
  ```makefile
  init: db-migrate
      @bin/apprun init
  ```
- [ ] **依赖 Story 1.6**：需要 Cobra 框架和 `cmd/cli/root.go` 先完成

### Task 4: Startup Integration（main.go）
- [ ] 在 `Phase 2.8` 添加初始化检查
  - 调用 `bootstrap.CheckInitialized()`
  - 使用 `bootstrap.DefaultConfig().AutoInit` 环境变量决定行为
  - 已创建 `core/internal/bootstrap/config.go` 定义 AUTO_INIT 配置结构

### Task 5: Data Migration Framework（为未来扩展）
- [ ] 扩展 `platform_meta` 表支持迁移记录（key=migration_xxx, value=timestamp）
- [ ] 创建 `migrations/data/` 目录（存放数据迁移脚本）
- [ ] 添加 `Makefile` target: `migrate-data`（执行数据迁移）
- [ ] 文档说明数据迁移与 Schema 迁移的区别

### Task 6: Testing & Documentation
- [ ] 单元测试：`bootstrap/initializer_test.go`（幂等性、事务回滚）
- [ ] 集成测试：完整初始化流程测试
- [ ] 更新 `.env.example` 添加 `PLATFORM_ADMIN_EMAIL/PASSWORD/AUTO_INIT`
- [ ] 更新 `README.md` 添加初始化步骤说明

---

## Technical Design

### Architecture Decision: 职责边界与分层

**程序外（Infrastructure Layer - 由部署脚本/CI 完成）**：
- 数据库 Schema 迁移（`make db-migrate` - Atlas CLI）
- 容器镜像构建和环境变量注入
- Secret 管理（DB密码、JWT密钥）

**程序内（Application Layer - 由应用启动完成）**：
- 初始化状态检查（Phase 2.8）
- 条件初始化（根据 `AUTO_INIT` 环境变量）
- 运行时配置加载和服务初始化

### Data Generation Flow

**初始化数据来源**：

| 数据类型 | 来源 | 示例 |
|---------|------|------|
| 系统用户 | 硬编码常量 | ID=1, username="system" |
| 管理员邮箱 | 环境变量 | $PLATFORM_ADMIN_EMAIL |
| 管理员密码 | 环境变量 + bcrypt | Hash($PLATFORM_ADMIN_PASSWORD) |
| UUID | 程序生成 | uuid.New() |
| 时间戳 | 程序生成 | time.Now() |
| 平台项目 | 硬编码常量 | name="Platform" |

**生成流程**：
```
环境变量 → 读取配置 → 动态生成（UUID/Hash/时间）→ Ent Client 插入
```

### Environment Strategy Matrix

| 环境 | DB Migration | AUTO_INIT | Admin Creation | Behavior |
|------|-------------|-----------|----------------|----------|
| **Local Dev** | Auto (`make dev-start`) | `true` | Auto-generate | 首次启动自动完成所有初始化 |
| **CI Tests** | Auto (test setup) | `true` | `admin@test.com` | 每个测试独立数据库 + seed |
| **Staging** | Manual (`make db-migrate`) | `false` | Manual (`make init`) | 显式初始化，模拟生产流程 |
| **Production** | CI/CD Pipeline | `false` | Manual (`make init`) | 运维手动执行，安全可控 |

### Two-Stage Initialization Flow

```
┌────────────── Deployment Phase (程序外) ──────────────┐
│                                                       │
│  Step 1: make db-migrate  (Schema Migration)        │
│  Step 2: make init        (Data Initialization)     │
│          ├─ Create system user (ID=1)               │
│          ├─ Create admin account                    │
│          ├─ Create platform project                 │
│          └─ Mark initialized (platform_meta)        │
│                                                       │
└───────────────────────────────────────────────────────┘
         │
         ▼
┌────────────── Runtime Phase (程序内) ─────────────────┐
│                                                       │
│  main.go Phase 2.8: Initialization Check            │
│  ┌──────────────────────────────────────────┐       │
│  │ if platform_meta.initialized exists      │       │
│  │   → Continue startup                     │       │
│  │ else if AUTO_INIT=true                   │       │
│  │   → Auto-initialize (dev mode)           │       │
│  │ else                                     │       │
│  │   → FATAL: "Run make init"               │       │
│  └──────────────────────────────────────────┘       │
│                                                       │
└───────────────────────────────────────────────────────┘
```

### Module Structure

```
core/
├── cmd/
│   ├── main.go                 # 程序入口
│   └── cli/
│       ├── root.go             # Cobra 根命令（来自 Story 1.6）
│       ├── start.go            # start 子命令（来自 Story 1.6）
│       ├── init.go             # NEW: init 子命令（本 Story）
│       └── ...                 # 其他子命令
├── pkg/
│   └── bootstrap/              # NEW: 初始化模块
│       ├── initializer.go      # 核心初始化逻辑
│       └── initializer_test.go
└── migrations/
    ├── 00X_platform_meta.sql   # NEW: platform_meta 表
    └── data/                   # NEW: 数据迁移目录 (预留)
        └── README.md           # 说明数据迁移用法
```

**架构说明：**
- `cmd/cli/init.go` 遵循 Story 1.6 的 Cobra 架构
- 统一的命令入口：`apprun init`（而非独立工具）
- 所有子命令都在 `cmd/cli/` 目录下，保持一致性

---

## Implementation Approach

### Key Implementation Points

**1. Idempotency Pattern**
```go
// bootstrap/initializer.go
func (i *Initializer) Initialize(ctx context.Context) error {
    // Check if already initialized (query platform_meta table)
    if isInitialized, _ := i.CheckInitialized(ctx); isInitialized {
        return nil // Safe to call multiple times
    }
    
    // Execute in transaction
    return i.db.WithTx(ctx, func(tx *ent.Tx) error {
        // 1. Create system user (ID=1)
        // 2. Create admin account
        // 3. Create platform project
        // 4. Mark initialized
    })
}
```

**2. Environment Detection in main.go**
```go
// Phase 2.8: Initialization Check
initializer := bootstrap.NewInitializer(dbClient, bootstrap.Config{
    AdminEmail:    env.Get("PLATFORM_ADMIN_EMAIL", "admin@example.com"),
    AdminPassword: env.Get("PLATFORM_ADMIN_PASSWORD", ""),
    AutoInit:      env.GetBool("PLATFORM_AUTO_INIT", false),
})

if isInit, _ := initializer.CheckInitialized(ctx); !isInit {
    if initializer.Config.AutoInit {
        // Dev mode: auto-initialize
        initializer.Initialize(ctx)
    } else {
        // Prod mode: fail fast
        log.Fatal("Platform not initialized. Run: make init")
    }
}
```

**3. CLI Command Structure**
```bash
# Makefile
init: db-migrate
    @echo "🚀 Initializing platform..."
    @cd core && go run cmd/init/main.go
    @echo "✅ Admin: $$PLATFORM_ADMIN_EMAIL"
```

---

## Testing Strategy

### Unit Tests
- [ ] `bootstrap/initializer_test.go`
  - TestCheckInitialized (已初始化 vs 未初始化)
  - TestInitialize_Success (完整流程)
  - TestInitialize_Idempotent (重复执行安全)
  - TestInitialize_TransactionRollback (失败回滚)

### Integration Tests
- [ ] 测试环境自动初始化（`AUTO_INIT=true`）
- [ ] 生产模式拒绝启动（`AUTO_INIT=false` + 未初始化）
- [ ] `make init` 命令端到端测试

### Manual Test Checklist
- [ ] 全新数据库执行 `make init` → 成功创建系统用户、管理员、平台项目
- [ ] 重复执行 `make init` → 幂等，不报错
- [ ] 启动应用（已初始化） → 正常启动
- [ ] 启动应用（未初始化 + `AUTO_INIT=false`） → FATAL 退出

---

## Definition of Done

- [ ] 代码实现并通过单元测试（覆盖率 ≥80%）
- [ ] `make init` 命令工作正常
- [ ] `main.go` 集成初始化检查逻辑
- [ ] 测试环境和生产环境行为符合预期
- [ ] 文档更新（README + .env.example）
- [ ] Code Review 通过
- [ ] Product Owner 验收通过

---

## Notes

**Security**
- 管理员密码通过环境变量配置，避免硬编码
- 生产环境禁用 `AUTO_INIT`，强制手动初始化

**Scalability**
- `platform_meta` 表可扩展存储其他元数据（版本号、特性开关等）
- 初始化流程可扩展支持更多默认数据导入

**Tool Responsibilities: Atlas vs Ent Client**
- **Atlas CLI** (Schema Only):
  - 表结构定义（CREATE TABLE）
  - 索引和约束（CREATE INDEX, ALTER TABLE）
  - Schema 版本控制
  - ❌ 不负责数据插入
- **Ent Client** (Data + Logic):
  - 业务数据初始化（INSERT via Go）
  - 动态数据生成（UUID, Hash, 时间戳）
  - 环境变量配置读取
  - 复杂业务逻辑（幂等性、事务、条件判断）
  - 错误处理和日志记录

**Data Migration vs Schema Migration**
- **Schema Migration** (Atlas): 表结构变更（`make db-migrate`）
- **Data Migration** (Go): 业务数据转换（`make migrate-data`）
- **Seed Data** (Bootstrap): 初始默认数据（`make init`）

**Testing Isolation**
- 每个集成测试使用独立数据库（`test_xxx` schema）
- 测试自动执行迁移和初始化，无需手动 setup

---

## References

- [Story 1.5: Database Migration](./1-5-database-migration.md)
- [PRD - Authentication](../../prd.md#21-authentication--authorization)
- [Deployment Standards](../../standards/deployment.md)
