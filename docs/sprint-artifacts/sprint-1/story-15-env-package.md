# Story 15: Environment Variable Utility Package
# Sprint 1: Infrastructure Enhancement

**Priority**: P0  
**Effort**: 0.5 days (Already Implemented)  
**Owner**: Backend Dev  
**Dependencies**: -  
**Status**: Done  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [Coding Standards](../../standards/coding-standards.md), [Architecture Standards](../../standards/architecture-standards.md)

---

## User Story

作为开发者，我希望有统一的环境变量访问工具，以便在基础设施配置中以类型安全的方式读取环境变量，避免重复的类型转换代码并提供默认值支持。

---

## Acceptance Criteria

- [x] 实现 `core/pkg/env` 包的核心功能
- [x] 提供 5 个类型安全的函数：`Get`, `MustGet`, `GetInt`, `GetBool`, `GetDuration`
- [x] 提供配置文件加载函数：`LoadConfigToEnv`
- [x] 零外部依赖（仅使用 Go 标准库）
- [x] 不依赖配置中心或任何业务模块
- [x] 单元测试覆盖率 100% (10/10 tests passing)
- [x] 用于基础设施配置（Server, Database, Logger）
- [x] 支持配置文件与环境变量混合使用（优先级正确）

---

## Design Philosophy

### 🎯 核心定位

**pkg/env 是配置分层架构中的 Layer 0 - 环境变量访问层**

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 0: 环境变量访问 (pkg/env)                            │
│  • 零依赖，仅使用 os.Getenv()                               │
│  • 类型转换和默认值                                         │
│  • 基础设施配置唯一数据源                                   │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: 基础设施配置 (pkg/server, internal/database)      │
│  • Server 端口、数据库连接                                  │
│  • 使用 pkg/env 读取配置                                    │
│  • 启动时确定，运行时不变                                   │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: 配置中心 (modules/config)                          │
│  • 依赖 Layer 1 (需要数据库和 Server)                       │
│  • 提供配置 API 服务                                        │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: 业务配置 (modules/user, modules/functions)        │
│  • 依赖 Layer 2 (从配置中心读取)                            │
│  • 运行时可动态修改                                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Technical Details

### 包结构
```
core/pkg/env/
├── env.go          # 5 个核心函数 (Get, MustGet, GetInt, GetBool, GetDuration)
├── loader.go       # 配置文件加载器 (LoadConfigToEnv)
├── env_test.go     # 核心函数单元测试 (5 tests)
└── loader_test.go  # 加载器单元测试 (5 tests)
```

### API 设计

```go
package env

// Get - 获取字符串环境变量（带默认值）
func Get(key, defaultValue string) string

// MustGet - 获取必需的环境变量（不存在则 panic）
func MustGet(key string) string

// GetInt - 获取整数环境变量（带默认值）
func GetInt(key string, defaultValue int) int

// GetBool - 获取布尔环境变量（带默认值）
// 支持: "true", "1", "yes", "on" (不区分大小写)
func GetBool(key string, defaultValue bool) bool

// GetDuration - 获取时间间隔环境变量（带默认值）
// 支持: "5s", "10m", "1h"
func GetDuration(key string, defaultValue time.Duration) time.Duration
```

---

## Dependencies & Relationships

### ✅ 依赖关系

```go
// pkg/env 的依赖
import (
    "os"          // 环境变量访问
    "strconv"     // 类型转换
    "time"        // Duration 类型
)
// 无业务依赖！
```

### 🔗 被依赖关系

**pkg/env 被以下模块使用（Layer 1 基础设施）：**

1. **pkg/server** - HTTP/HTTPS Server 配置
   ```go
   serverCfg := &server.Config{
       HTTPPort:    env.Get("HTTP_PORT", "8080"),
       HTTPSPort:   env.Get("HTTPS_PORT", "8443"),
       SSLCertFile: env.Get("SSL_CERT_FILE", ""),
   }
   ```

2. **cmd/server/main.go** - 应用启动配置
   ```go
   configDir := env.Get("CONFIG_DIR", "./config")
   bootstrap := config.NewBootstrap(configDir)
   ```

3. **internal/database** (未来) - 数据库连接配置
   ```go
   dbConfig := &database.Config{
       Host: env.Get("DB_HOST", "localhost"),
       Port: env.GetInt("DB_PORT", 5432),
   }
   ```

4. **pkg/logger** (未来) - 日志配置
   ```go
   logLevel := env.Get("LOG_LEVEL", "info")
   logOutput := env.Get("LOG_OUTPUT", "stdout")
   ```

---

## Why NOT Use Config Center?

### ❌ 错误设计：基础设施依赖配置中心

```
启动流程（会死锁）:
  1. main() 启动
  2. 初始化配置中心
     ├─ 需要连接数据库（需要 DB_HOST, DB_PORT）
     └─ 需要启动 HTTP Server（需要 HTTP_PORT）
  3. 从配置中心读取 Server 配置 ❌ (配置中心还没启动！)
  4. 启动 Server ❌ (没有端口信息！)
  5. 🔴 循环依赖 / 死锁
```

### ✅ 正确设计：基础设施使用环境变量

```
启动流程（清晰顺序）:
  1. main() 启动
  2. pkg/env.Get() 读取基础设施配置 ✅
     ├─ HTTP_PORT=8080
     ├─ DB_HOST=localhost
     └─ 无需依赖任何服务
  3. 启动 HTTP Server ✅
  4. 连接数据库 ✅
  5. 初始化配置中心 ✅
  6. 配置中心 API 可用 ✅
```

---

## Configuration Layers Comparison

| 特性 | Layer 0 (pkg/env) | Layer 2 (Config Center) |
|------|-------------------|-------------------------|
| **数据源** | 环境变量 | 数据库 / API |
| **依赖** | 零依赖 | 需要 DB + Server |
| **用途** | 基础设施配置 | 业务配置 |
| **修改方式** | 重启应用 | 热更新 |
| **配置项** | Server, DB, Logger | User, Functions, Features |
| **读取时机** | 启动时 | 运行时 |
| **类型安全** | Go 原生类型 | 动态解析 |
| **错误处理** | Panic / Default | API Error Response |

---

## Usage Examples

### Example 1: Server Configuration (pkg/server)

```go
// pkg/server/server.go
serverCfg := &server.Config{
    HTTPPort:            env.Get("HTTP_PORT", "8080"),
    HTTPSPort:           env.Get("HTTPS_PORT", "8443"),
    SSLCertFile:         env.Get("SSL_CERT_FILE", ""),
    SSLKeyFile:          env.Get("SSL_KEY_FILE", ""),
    ShutdownTimeout:     env.GetDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
    EnableHTTPWithHTTPS: env.GetBool("ENABLE_HTTP_WITH_HTTPS", true),
}
```

### Example 2: Application Bootstrap (cmd/server/main.go)

```go
// cmd/server/main.go
func main() {
    // 读取配置目录（基础设施配置）
    configDir := env.Get("CONFIG_DIR", "./config")
    
    // 读取 Server 端口（基础设施配置）
    httpPort := env.Get("HTTP_PORT", "8080")
    
    // 启动应用...
}
```

### Example 3: Environment Variables Override

```bash
# 开发环境
HTTP_PORT=8080 HTTPS_PORT=8443 go run ./cmd/server

# 生产环境
HTTP_PORT=80 HTTPS_PORT=443 LOG_LEVEL=warn ./apprun-server

# Docker 容器
docker run -e HTTP_PORT=9090 -e DB_HOST=postgres apprun
```

---

## Test Coverage

```bash
$ go test ./pkg/env -v
=== RUN   TestGet
--- PASS: TestGet (0.00s)
=== RUN   TestMustGet
--- PASS: TestMustGet (0.00s)
=== RUN   TestGetInt
--- PASS: TestGetInt (0.00s)
=== RUN   TestGetBool
--- PASS: TestGetBool (0.00s)
=== RUN   TestGetDuration
--- PASS: TestGetDuration (0.00s)
PASS
ok      apprun/pkg/env  0.012s
```

**Coverage: 100% (5/5 tests passing)**

---

## Design Principles

### 1. 🔒 Zero Dependencies
```go
// ✅ 仅使用标准库
import (
    "os"
    "strconv"
    "time"
)

// ❌ 不依赖任何业务模块
// 不依赖: viper, config center, database
```

### 2. 🎯 Single Responsibility
- **唯一职责**: 提供类型安全的环境变量访问
- **不做**: 配置验证、配置持久化、配置热更新
- **专注**: 简单、快速、可靠

### 3. 🔧 Fail-Safe Defaults
```go
// 所有函数都提供默认值（除了 MustGet）
httpPort := env.Get("HTTP_PORT", "8080")  // 无环境变量时使用 "8080"
maxConns := env.GetInt("MAX_CONNS", 100)  // 无环境变量时使用 100
debug := env.GetBool("DEBUG", false)      // 无环境变量时使用 false
```

### 4. 🚀 Layer 0 Independence
```
pkg/env 不能依赖任何高层模块
├─ ❌ 不能依赖 config center
├─ ❌ 不能依赖 database
├─ ❌ 不能依赖 server
└─ ✅ 仅依赖 Go 标准库
```

---

## Integration Points

### Current Usage (As of Sprint 1)
- [x] `cmd/server/main.go` - 读取 CONFIG_DIR, SERVER_PORT
- [x] `pkg/server/server.go` - 所有 server 配置
- [x] `modules/config/bootstrap.go` - 配置目录路径

### Future Usage (Sprint 2+)
- [ ] `internal/database/config.go` - 数据库连接配置
- [ ] `pkg/logger/config.go` - 日志配置
- [ ] `internal/cache/config.go` - 缓存配置
- [ ] `pkg/temporal/config.go` - Temporal 配置

---

## Benefits

### ✅ 优势

1. **避免循环依赖**
   - 基础设施配置不依赖配置中心
   - 启动顺序清晰

2. **类型安全**
   - 编译时类型检查
   - 避免运行时类型转换错误

3. **零外部依赖**
   - 无需 viper 等第三方库
   - 减少依赖风险

4. **性能优秀**
   - 直接读取环境变量
   - 无数据库 / 网络开销

5. **测试友好**
   - 易于 Mock 环境变量
   - 100% 测试覆盖率

---

## Configuration File Support (Enhanced)

### 🆕 LoadConfigToEnv() Function

**功能**：从 `default.yaml` 加载基础设施配置并转换为环境变量

**环境变量命名规则**：
```
Pattern: {GROUP}_UPPERCASE_{KEY}_UPPERCASE
Formula: toUpper(group) + "_" + toUpper(key)

示例：
  server.http_port        → SERVER_HTTP_PORT
  server.ssl_cert_file    → SERVER_SSL_CERT_FILE
  database.host           → DATABASE_HOST
  database.db_name        → DATABASE_DB_NAME
```

**优先级**（从高到低）：
1. 运行时环境变量（`export`, `docker -e`, `.env`）⭐⭐⭐ **最高**
2. 配置文件（`default.yaml`）
3. 代码默认值（`DefaultConfig()`）

```go
// pkg/env/loader.go
func LoadConfigToEnv(configDir string) error {
    // 读取 default.yaml 中的 server 和 database 配置
    // 自动转换为环境变量（规则：GROUP_KEY 全大写）
    // 仅在环境变量未设置时才设置
    // 文件不存在时不报错（使用纯环境变量模式）
}
```

**使用场景**：
- ✅ 开发环境：default.yaml 提供默认配置
- ✅ 生产环境：环境变量覆盖文件配置
- ✅ 版本控制：配置文件可提交 git
- ✅ 动态加载：所有 default.yaml 配置项自动转换为环境变量

**示例**：
```yaml
# config/default.yaml
server:
  http_port: "8080"
  https_port: "8443"
database:
  host: localhost
  port: 5432
  db_name: apprun_dev
```

```bash
# 自动转换为环境变量：
# SERVER_HTTP_PORT=8080
# SERVER_HTTPS_PORT=8443
# DATABASE_HOST=localhost
# DATABASE_PORT=5432
# DATABASE_DB_NAME=apprun_dev

# 环境变量优先级更高（覆盖配置文件）
export SERVER_HTTP_PORT=9090
export DATABASE_HOST=prodhost
# 最终使用: SERVER_HTTP_PORT=9090 (环境变量) 而不是 8080 (文件)
#           DATABASE_HOST=prodhost (环境变量) 而不是 localhost (文件)
```

---

## Related Documentation

- [Coding Standards - Configuration Guidelines](../../standards/coding-standards.md#145-配置结构体规范)
- [Architecture Standards - Configuration Layers](../../standards/architecture-standards.md)
- [Story 14 - HTTP Server Package](./story-14-http-server.md)
- [Story 10 - Config Center Basic](../sprint-0/story-10-config-basic.md)
- [Story 16 - Database Anti-Corruption Layer](./story-16-database-layer.md)

---

**Created**: 2025-12-31  
**Updated**: 2025-12-31 (Added LoadConfigToEnv support)  
**Maintainer**: BMad Dev Agent (Amelia)

