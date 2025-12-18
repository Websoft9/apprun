# apprun POC 快速参考指南

## 📋 核心文档索引

| 文档 | 路径 | 用途 |
|------|------|------|
| **POC 验证计划** | `docs/poc/poc-validation-plan-2025-12-18.md` | 完整的 5 天验证计划 |
| **POC 环境说明** | `poc/README.md` | 环境使用指南和测试场景 |
| **技术架构文档** | `docs/architecture/technical-architecture-apprun-lightweight-2025-12-18.md` | 完整架构设计 |

---

## 🚀 5 天 POC 时间表

```
Day 1 ████████  环境搭建 + PostgREST 集成验证
Day 2 ████████  Ory Kratos + Casbin 认证授权验证
Day 3 ████████  WASM 函数运行时验证
Day 4 ████████  Temporal 工作流集成验证
Day 5 ████████  集成测试 + 性能基准 + 报告
```

---

## 🎯 成功标准

| 指标 | 目标值 | 验证方式 |
|------|--------|---------|
| **API响应时间** | P95 < 200ms | wrk 压测 |
| **并发能力** | 支持 500+ RPS | 负载测试 |
| **内存占用** | < 1GB (总计) | docker stats |
| **函数启动时间** | < 100ms (WASM) | Go benchmark |
| **认证性能** | > 1000 验证/秒 | Go benchmark |

---

## 📦 快速开始（3 步）

### 第 1 步：启动 POC 环境

```bash
cd poc
./start-poc.sh
```

**预期输出**：所有服务启动成功，显示访问地址

### 第 2 步：验证服务

```bash
# 测试 PostgREST
curl http://localhost:3000/products

# 测试 Kratos
curl http://localhost:4433/health/alive

# 测试 Temporal Web UI
open http://localhost:8233
```

### 第 3 步：开始开发验证

```bash
# 创建 Go 项目
cd ..
mkdir -p apprun-core/cmd/server
mkdir -p apprun-core/internal/{auth,data,function,workflow}

# 初始化 Go 模块
cd apprun-core
go mod init github.com/apprun/core

# 安装依赖
go get github.com/gin-gonic/gin
go get github.com/casbin/casbin/v2
go get github.com/wasmerio/wasmer-go/wasmer
go get go.temporal.io/sdk
```

---

## 🧪 核心验证场景

### 场景 1️⃣: PostgREST 集成 (Day 1)

**目标**：Go 代理 PostgREST，注入认证和租户隔离

```go
// main.go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http/httputil"
    "net/url"
)

func main() {
    r := gin.Default()
    
    // 代理到 PostgREST
    target, _ := url.Parse("http://localhost:3000")
    proxy := httputil.NewSingleHostReverseProxy(target)
    
    r.Any("/data/*path", func(c *gin.Context) {
        // TODO: 验证 JWT
        // TODO: 注入 tenant_id
        proxy.ServeHTTP(c.Writer, c.Request)
    })
    
    r.Run(":8080")
}
```

**测试**：
```bash
curl http://localhost:8080/data/products
```

---

### 场景 2️⃣: 认证授权集成 (Day 2)

**目标**：Ory Kratos 认证 + Casbin 授权

```go
// internal/auth/casbin.go
package auth

import (
    "github.com/casbin/casbin/v2"
)

func InitCasbin() (*casbin.Enforcer, error) {
    enforcer, err := casbin.NewEnforcer("model.conf", "policy.csv")
    if err != nil {
        return nil, err
    }
    return enforcer, nil
}

func CheckPermission(enforcer *casbin.Enforcer, user, resource, action string) (bool, error) {
    return enforcer.Enforce(user, resource, action)
}
```

**测试**：
```bash
go test -bench=BenchmarkCasbinEnforce
```

---

### 场景 3️⃣: WASM 函数执行 (Day 3)

**目标**：加载并执行 WASM 模块

```go
// internal/function/wasm.go
package function

import (
    "github.com/wasmerio/wasmer-go/wasmer"
)

func ExecuteWasm(wasmBytes []byte, functionName string, args ...interface{}) (interface{}, error) {
    engine := wasmer.NewEngine()
    store := wasmer.NewStore(engine)
    
    module, err := wasmer.NewModule(store, wasmBytes)
    if err != nil {
        return nil, err
    }
    
    instance, err := wasmer.NewInstance(module, wasmer.NewImportObject())
    if err != nil {
        return nil, err
    }
    
    fn, err := instance.Exports.GetFunction(functionName)
    if err != nil {
        return nil, err
    }
    
    return fn(args...)
}
```

**测试**：
```bash
go test -bench=BenchmarkWasmExecution
```

---

### 场景 4️⃣: Temporal 工作流 (Day 4)

**目标**：定义并执行工作流

```go
// internal/workflow/registration.go
package workflow

import (
    "go.temporal.io/sdk/workflow"
)

func UserRegistrationWorkflow(ctx workflow.Context, email string) error {
    // Step 1: 发送欢迎邮件
    err := workflow.ExecuteActivity(ctx, SendWelcomeEmail, email).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // Step 2: 创建默认项目
    err = workflow.ExecuteActivity(ctx, CreateDefaultProject, email).Get(ctx, nil)
    if err != nil {
        return err
    }
    
    return nil
}
```

**测试**：
```bash
# 启动 Worker
go run cmd/worker/main.go

# 触发工作流
curl -X POST http://localhost:8080/api/v1/workflows/execute \
  -d '{"workflow":"user_registration","input":{"email":"test@example.com"}}'
```

---

## 📊 性能测试工具

### wrk (HTTP 压测)

```bash
# 安装
sudo apt-get install wrk

# 基础压测
wrk -t4 -c100 -d30s http://localhost:8080/data/products

# 带认证的压测
wrk -t4 -c100 -d30s \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/workflows
```

### Go Benchmark

```go
func BenchmarkAuthCheck(b *testing.B) {
    enforcer, _ := InitCasbin()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        enforcer.Enforce("alice", "/api/workflows", "GET")
    }
}

func BenchmarkWasmExecution(b *testing.B) {
    runtime, _ := NewWasmRuntime(wasmBytes)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        runtime.Execute("add", 1, 2)
    }
}
```

**运行**：
```bash
go test -bench=. -benchmem
```

---

## 🔍 监控命令

### 资源监控

```bash
# 实时监控
docker stats

# 导出 CSV
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" > resources.csv
```

### 日志查看

```bash
# 所有服务
docker-compose logs -f

# 特定服务
docker-compose logs -f postgres
docker-compose logs -f postgrest
docker-compose logs -f kratos
docker-compose logs -f temporal
```

### 数据库查询

```sql
-- 租户统计
SELECT * FROM tenant_stats;

-- 工作流统计
SELECT * FROM workflow_execution_stats;

-- 性能分析
EXPLAIN ANALYZE SELECT * FROM products WHERE tenant_id = '11111111-1111-1111-1111-111111111111';
```

---

## 🐛 常见问题

### Q1: 服务启动失败？

```bash
# 检查端口占用
sudo netstat -tulpn | grep -E '(5432|3000|4433|7233|9000)'

# 清理旧容器
docker-compose down -v
./start-poc.sh
```

### Q2: PostgREST 返回 401？

```bash
# 检查 JWT secret 配置
docker-compose logs postgrest | grep JWT

# 使用匿名访问测试
curl http://localhost:3000/products
```

### Q3: Temporal Web UI 无法访问？

```bash
# 检查 Temporal 健康状态
docker-compose logs temporal | grep -i error

# 重启 Temporal
docker-compose restart temporal
```

### Q4: 内存占用超过 1GB？

```bash
# 查看详细占用
docker stats --no-stream

# 优化建议
# - PostgreSQL: 调整 shared_buffers
# - Temporal: 使用 SQLite 模式
# - 关闭不必要的服务
```

---

## 📝 验证清单

### ✅ Day 1 - PostgREST 集成
- [ ] PostgREST API 可访问
- [ ] Go 代理工作正常
- [ ] JWT 认证注入成功
- [ ] 多租户隔离有效
- [ ] 响应时间 < 100ms
- [ ] 生成 `day1-postgrest-integration.md`

### ✅ Day 2 - 认证授权
- [ ] Kratos 注册/登录流程
- [ ] JWT 生成和验证
- [ ] Casbin 权限检查
- [ ] 性能 > 1000 auth/s
- [ ] 生成 `day2-auth-integration.md`

### ✅ Day 3 - WASM 运行时
- [ ] WASM 模块加载
- [ ] 函数执行成功
- [ ] 启动时间 < 100ms
- [ ] 热执行 < 1ms
- [ ] 生成 `day3-wasm-runtime.md`

### ✅ Day 4 - Temporal 集成
- [ ] Worker 注册
- [ ] 工作流执行
- [ ] 失败重试机制
- [ ] Web UI 可用
- [ ] 生成 `day4-temporal-integration.md`

### ✅ Day 5 - 集成测试
- [ ] 端到端流程测试
- [ ] 性能基准测试
- [ ] 资源占用验证
- [ ] 生成 `poc-summary-report.md`

---

## 📚 参考资料

### 官方文档
- [PostgREST](https://postgrest.org/en/stable/)
- [Ory Kratos](https://www.ory.sh/kratos/docs/)
- [Casbin](https://casbin.org/docs/overview)
- [Wasmer](https://docs.wasmer.io/)
- [Temporal](https://docs.temporal.io/)

### 代码示例
- [PostgREST + Go](https://github.com/PostgREST/postgrest/tree/main/test)
- [Kratos + Go](https://github.com/ory/kratos-client-go)
- [Casbin + Gin](https://github.com/casbin/casbin/tree/master/examples)
- [Temporal + Go](https://github.com/temporalio/samples-go)

### 性能优化
- [Go 性能优化](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)
- [PostgreSQL 调优](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [WASM 性能](https://hacks.mozilla.org/category/webassembly/)

---

## 🎓 学习路径

### Week 1: 基础准备
1. Go 语言基础复习
2. Docker Compose 使用
3. PostgreSQL 基础
4. RESTful API 设计

### Week 2: POC 验证（本周）
按照 5 天计划执行验证

### Week 3: MVP 开发准备
1. 代码结构设计
2. CI/CD 搭建
3. 测试框架选型
4. 文档体系建立

---

**维护信息**
- 创建者: Root
- 最后更新: 2025-12-18
- 版本: 1.0
- 反馈: dev@apprun.dev
