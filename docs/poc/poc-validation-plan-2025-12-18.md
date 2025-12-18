# apprun 技术验证 POC 计划

**文档版本：** 1.0  
**创建日期：** 2025-12-18  
**负责人：** 开发团队  
**计划周期：** 1周（5个工作日）  
**状态：** 待执行  

---

## 1. 验证目标

本 POC 旨在验证 apprun 轻量级架构的**核心技术可行性**，降低 MVP 开发风险，确保关键技术选型正确。

### 1.1 核心验证问题

1. ✅ Go + PostgREST 集成是否顺畅？
2. ✅ Ory Kratos + Casbin 认证授权方案是否可行？
3. ✅ WASM 函数运行时性能是否满足要求？
4. ✅ 整体资源占用是否符合轻量级目标？

**说明**：
- ~~Temporal 工作流引擎~~：**已通过POC验证**，无需重复验证
- ~~MinIO 对象存储~~：**不再使用**（已转为闭源商业产品），后续考虑其他方案

### 1.2 成功标准

| 指标 | 目标值 | 验证方式 |
|------|--------|---------|
| **API响应时间** | P95 < 200ms | 压测工具 |
| **并发能力** | 支持 500+ RPS | 负载测试 |
| **内存占用** | < 512MB (核心服务) | 监控工具 |
| **函数启动时间** | < 100ms (WASM) | 性能测试 |
| **认证性能** | > 1000 验证/秒 | 基准测试 |

---

## 2. 验证计划（3天）

**调整说明**：
- 原5天计划缩减为3天
- 移除 Day 4 Temporal验证（已完成）
- 移除 MinIO 相关验证（已弃用）

### Day 1: 环境搭建 + Go/PostgREST 集成

**时间分配：** 8小时

#### 2.1 环境准备（2小时）

**任务清单：**
```bash
# 1. 安装依赖
- [ ] Go 1.21+ 安装
- [ ] Docker & Docker Compose 安装
- [ ] PostgreSQL 15 本地环境
- [ ] PostgREST 12+ 安装

# 2. 初始化项目
- [ ] 创建 Go 项目结构
- [ ] 配置 go.mod 依赖
- [ ] 设置开发环境变量
```

**验收标准：**
- Go 版本 >= 1.21
- PostgreSQL 可连接
- PostgREST 可访问

#### 2.2 Go + PostgREST 集成（4小时）

**验证场景：**

1️⃣ **基础集成**
```go
// 目标：Go 服务代理 PostgREST 请求
package main

import (
    "github.com/gin-gonic/gin"
    "net/http/httputil"
)

func main() {
    r := gin.Default()
    
    // 代理到 PostgREST
    postgrestProxy := httputil.NewSingleHostReverseProxy(
        url.Parse("http://localhost:3000"),
    )
    
    r.Any("/data/*path", func(c *gin.Context) {
        postgrestProxy.ServeHTTP(c.Writer, c.Request)
    })
    
    r.Run(":8080")
}
```

2️⃣ **认证注入**
```go
// 目标：验证 JWT 并注入 tenant_id
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        
        // 验证 JWT
        claims, err := ValidateJWT(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // 注入租户ID到请求头
        c.Request.Header.Set("X-Tenant-ID", claims.TenantID)
        c.Next()
    }
}
```

3️⃣ **多租户隔离**
```sql
-- 目标：验证 RLS 策略
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255)
);

ALTER TABLE products ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON products
    USING (tenant_id = current_setting('request.jwt.claims')::json->>'tenant_id')::uuid);
```

**验收标准：**
- ✅ Go 可成功代理 PostgREST API
- ✅ JWT 认证中间件工作正常
- ✅ 多租户数据隔离有效
- ✅ 响应时间 < 50ms（无负载）

#### 2.3 性能基准测试（2小时）

```bash
# 使用 wrk 压测
wrk -t4 -c100 -d30s --latency http://localhost:8080/data/products

# 目标指标
# - RPS > 500
# - P95 延迟 < 100ms
# - 错误率 < 0.1%
```

**输出文档：**
- `poc-results/day1-postgrest-integration.md`

---

### Day 2: Ory Kratos + Casbin 认证授权

**时间分配：** 8小时

#### 2.1 Ory Kratos 部署（2小时）

**任务清单：**
```bash
# 1. Docker Compose 启动 Kratos
docker-compose up -d kratos kratos-migrate postgres

# 2. 配置身份架构（Identity Schema）
cat > kratos/identity.schema.json <<EOF
{
  "$id": "https://apprun.dev/identity.schema.json",
  "title": "User",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "email": {
          "type": "string",
          "format": "email"
        },
        "tenant_id": {
          "type": "string",
          "format": "uuid"
        }
      },
      "required": ["email", "tenant_id"]
    }
  }
}
EOF

# 3. 测试注册/登录
curl -X POST http://localhost:4433/self-service/registration/flows
```

**验收标准：**
- Kratos API 可访问
- 注册/登录流程正常
- 会话管理有效

#### 2.2 Casbin 集成（3小时）

**验证场景：**

1️⃣ **RBAC 模型定义**
```ini
# model.conf
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

2️⃣ **策略加载**
```go
package main

import (
    "github.com/casbin/casbin/v2"
    gormadapter "github.com/casbin/gorm-adapter/v3"
)

func InitCasbin() (*casbin.Enforcer, error) {
    // 使用 PostgreSQL 存储策略
    adapter, err := gormadapter.NewAdapter(
        "postgres",
        "host=localhost port=5432 user=apprun dbname=apprun sslmode=disable",
    )
    if err != nil {
        return nil, err
    }
    
    enforcer, err := casbin.NewEnforcer("model.conf", adapter)
    if err != nil {
        return nil, err
    }
    
    // 加载策略
    enforcer.LoadPolicy()
    
    return enforcer, nil
}
```

3️⃣ **权限检查中间件**
```go
func AuthzMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.GetString("user_id") // 从JWT获取
        resource := c.Request.URL.Path
        action := c.Request.Method
        
        // 执行权限检查
        allowed, err := enforcer.Enforce(user, resource, action)
        if err != nil || !allowed {
            c.JSON(403, gin.H{"error": "Forbidden"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

**验收标准：**
- ✅ Kratos 认证流程完整
- ✅ Casbin 权限检查正常
- ✅ 性能：> 1000 次权限检查/秒
- ✅ JWT + RBAC 集成无缝

#### 2.3 集成测试（3小时）

**测试场景：**
```go
func TestAuthFlow(t *testing.T) {
    // 1. 注册用户
    // 2. 登录获取 JWT
    // 3. 访问受保护 API
    // 4. 验证权限控制
}
```

**输出文档：**
- `poc-results/day2-auth-integration.md`

---

### Day 3: WASM 函数运行时

**时间分配：** 8小时

#### 3.1 WASM 运行时选型（2小时）

**候选方案对比：**

| 方案 | 启动时间 | 内存占用 | 多语言支持 | 推荐度 |
|------|---------|---------|-----------|--------|
| wasmer-go | < 10ms | ~10MB | Rust/C/AssemblyScript | ★★★★★ |
| wasmtime-go | < 20ms | ~15MB | Rust/C | ★★★★☆ |
| wazero | < 5ms | ~5MB | 纯Go实现 | ★★★★☆ |

**选择：wasmer-go**（生态成熟）

#### 3.2 实现 WASM 执行器（4小时）

**验证场景：**

1️⃣ **基础执行器**
```go
package function

import (
    "github.com/wasmerio/wasmer-go/wasmer"
)

type WasmRuntime struct {
    store    *wasmer.Store
    module   *wasmer.Module
    instance *wasmer.Instance
}

func NewWasmRuntime(wasmBytes []byte) (*WasmRuntime, error) {
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
    
    return &WasmRuntime{
        store:    store,
        module:   module,
        instance: instance,
    }, nil
}

func (r *WasmRuntime) Execute(functionName string, args ...interface{}) (interface{}, error) {
    fn, err := r.instance.Exports.GetFunction(functionName)
    if err != nil {
        return nil, err
    }
    
    result, err := fn(args...)
    return result, err
}
```

2️⃣ **测试函数（Rust）**
```rust
// hello.rs
#[no_mangle]
pub extern "C" fn add(a: i32, b: i32) -> i32 {
    a + b
}

// 编译为 WASM
// rustc --target wasm32-unknown-unknown -O hello.rs
```

3️⃣ **性能测试**
```go
func BenchmarkWasmExecution(b *testing.B) {
    runtime, _ := NewWasmRuntime(wasmBytes)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        runtime.Execute("add", 1, 2)
    }
}

// 目标：< 100ms/次（冷启动）
//      < 1ms/次（热执行）
```

**验收标准：**
- ✅ WASM 模块可加载
- ✅ 函数调用成功
- ✅ 启动时间 < 100ms
- ✅ 热执行时间 < 1ms
- ✅ 内存隔离有效

#### 3.3 多语言支持验证（2小时）

**测试语言：**
- Rust → WASM（最佳）
- C → WASM
- AssemblyScript → WASM

**输出文档：**
- `poc-results/day3-wasm-runtime.md`

---

### Day 3: 集成测试 + 性能基准

**时间分配：** 8小时

#### 3.1 端到端集成测试（4小时）

**测试场景：完整用户注册流程**

```go
func TestE2EUserRegistration(t *testing.T) {
    // 1. 用户注册（Ory Kratos）
    registerResp := RegisterUser("test@example.com", "password123")
    assert.Equal(t, 201, registerResp.StatusCode)
    
    // 2. 登录获取 JWT
    loginResp := Login("test@example.com", "password123")
    token := loginResp.AccessToken
    
    // 3. 访问数据 API（PostgREST + 多租户隔离）
    products := GetProducts(token)
    assert.Empty(t, products)
    
    // 4. 创建数据
    CreateProduct(token, map[string]interface{}{
        "name": "Test Product",
    })
    
    // 5. 执行 WASM 函数
    result := ExecuteFunction(token, "process_product", map[string]interface{}{
        "product_id": "xxx",
    })
    assert.Equal(t, "success", result.Status)
}
```

#### 3.2 性能基准测试（3小时）

**测试矩阵：**

| 测试项 | 工具 | 目标 | 验证方式 |
|--------|------|------|---------|
| API 响应时间 | wrk | P95 < 200ms | HTTP 压测 |
| 并发能力 | wrk | 500+ RPS | 负载测试 |
| 认证性能 | benchmark | 1000+ auth/s | Go benchmark |
| WASM 执行 | benchmark | < 100ms 冷启动 | Go benchmark |
| 内存占用 | docker stats | < 512MB 核心服务 | 监控工具 |

**压测脚本：**
```bash
#!/bin/bash

echo "=== API 响应时间测试 ==="
wrk -t4 -c100 -d30s --latency http://localhost:8080/data/products

echo "=== 认证性能测试 ==="
wrk -t4 -c100 -d30s --latency \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/workflows

echo "=== 资源占用监控 ==="
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"
```

#### 3.3 生成 POC 报告（1小时）

**报告结构：**
```markdown
# POC 验证总结报告

## 1. 验证目标回顾
## 2. 测试结果汇总
   - Go + PostgREST 集成结果
   - Ory Kratos + Casbin 认证授权结果
   - WASM 函数运行时结果
## 3. 性能指标
## 4. 风险识别
## 5. 优化建议
## 6. 下一步行动
```

**输出文档：**
- `poc-results/poc-summary-report.md`

---

## 3. 资源需求

### 3.1 人员配置

| 角色 | 人数 | 投入时间 |
|------|------|---------|
| 全栈开发 | 1-2人 | 5天全职 |
| 架构师（指导） | 1人 | 5天兼职 |

### 3.2 硬件需求

**开发环境：**
- CPU: 4核
- 内存: 8GB
- 磁盘: 50GB SSD

**测试环境：**
- CPU: 2核
- 内存: 4GB
- 磁盘: 20GB SSD

### 3.3 软件依赖

```yaml
# docker-compose.yml (POC环境)
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    ports: ["5432:5432"]
    environment:
      POSTGRES_DB: apprun_poc
      POSTGRES_USER: apprun
      POSTGRES_PASSWORD: apprun123

  postgrest:
    image: postgrest/postgrest:v12.0.2
    ports: ["3000:3000"]
    environment:
      PGRST_DB_URI: postgres://apprun:apprun123@postgres:5432/apprun_poc
      PGRST_DB_ANON_ROLE: web_anon

  kratos:
    image: oryd/kratos:latest
    ports: ["4433:4433", "4434:4434"]
    environment:
      DSN: postgres://apprun:apprun123@postgres:5432/apprun_poc
```

**说明**：
- ✅ PostgreSQL + PostgREST + Kratos 是验证核心
- ❌ Temporal 已验证，无需包含在 POC 环境
- ❌ MinIO 已弃用，后续考虑其他开源方案（如 SeaweedFS）

---

## 4. 风险管理

### 4.1 已知风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| Go 团队经验不足 | 中 | 中 | 提前学习 Go 最佳实践 |
| WASM 生态不成熟 | 低 | 中 | 备选 Docker 方案 |
| 性能不达标 | 低 | 高 | 早期性能测试 |
| 集成复杂度超预期 | 中 | 中 | 简化 MVP 范围 |

### 4.2 应急预案

**如果 WASM 方案失败：**
- 降级到 Docker 函数运行时
- 评估性能差异
- 调整架构设计

**如果性能不达标：**
- 优化数据库查询
- 增加缓存层
- 调整资源配置

**如果集成困难：**
- 寻求社区支持
- 调整技术选型
- 延长 POC 时间

---

## 5. 交付物

### 5.1 代码仓库

```
apprun-poc/
├── cmd/
│   └── server/          # POC 主程序
├── internal/
│   ├── auth/            # 认证模块
│   ├── data/            # 数据 API
│   ├── function/        # WASM 运行时
│   └── workflow/        # Temporal 集成
├── test/
│   ├── integration/     # 集成测试
│   └── benchmark/       # 性能测试
├── wasm/
│   └── examples/        # WASM 示例函数
├── docker-compose.yml   # POC 环境
└── README.md
```

### 5.2 文档输出

1. **每日验证报告**
   - `poc-results/day1-postgrest-integration.md`
   - `poc-results/day2-auth-integration.md`
   - `poc-results/day3-wasm-runtime.md` (合并到 Day 2)

2. **性能测试报告**
   - `poc-results/performance-benchmark.md`

3. **POC 总结报告**
   - `poc-results/poc-summary-report.md`

### 5.3 决策建议

基于 POC 结果，提供：
- ✅ 技术选型确认/调整建议
- ✅ 架构优化方案
- ✅ MVP 开发风险评估
- ✅ 资源规划调整建议

---

## 6. 时间表

### 6.1 甘特图（调整为3天）

```
Day 1: 环境搭建 + PostgREST     ████████
Day 2: 认证授权 + WASM          ████████████
Day 3: 集成测试 + 报告          ████████
```

### 6.2 里程碑

| 日期 | 里程碑 | 交付物 |
|------|--------|--------|
| Day 1 | PostgREST 集成完成 | 集成代码 + 测试报告 |
| Day 2 | 认证授权 + WASM 验证完成 | 认证模块 + WASM执行器 + 性能数据 |
| Day 3 | POC 总结报告 | 完整报告 + 决策建议 |

---

## 7. 评审与决策

### 7.1 POC 评审会议

**时间：** Day 3 下午  
**参会人：** 架构师、开发团队、PM  
**议程：**
1. POC 成果演示（30分钟）
2. 性能指标评审（15分钟）
3. 风险讨论（15分钟）
4. 决策表决（30分钟）

### 7.2 决策点

基于 POC 结果，需要决策：

1. **技术选型确认**
   - [ ] Go + PostgREST 方案确认
   - [ ] Ory Kratos + Casbin 方案确认
   - [ ] WASM 运行时方案确认
   - [x] Temporal 工作流方案确认（已验证）

2. **架构调整**
   - [ ] 是否需要调整资源目标
   - [ ] 是否需要增加缓存层
   - [ ] 对象存储方案选择（MinIO已弃用，考虑SeaweedFS等）

3. **进入 MVP 开发**
   - [ ] Go/No Go 决策
   - [ ] 确定 MVP 开发团队
   - [ ] 明确开发时间表

---

## 附录

### A. 参考资料

**技术文档：**
- PostgREST 官方文档: https://postgrest.org
- Ory Kratos 文档: https://www.ory.sh/kratos/docs/
- Casbin 文档: https://casbin.org/docs/overview
- Wasmer 文档: https://docs.wasmer.io/
- Temporal 文档: https://docs.temporal.io/

**最佳实践：**
- Go 项目布局: https://github.com/golang-standards/project-layout
- WASM 性能优化: https://hacks.mozilla.org/category/webassembly/

### B. 联系方式

**技术支持：**
- 架构师：Root
- 开发团队：TBD

**紧急联系：**
- Slack: #apprun-poc
- Email: dev@apprun.dev

---

**文档维护：**
- 创建者：Root
- 最后更新：2025-12-18
- 版本：1.0
