# Epic: 函数服务
# apprun BaaS Platform

**关联 PRD**: [FR-FUNC-001](../prd.md#24-函数服务)  
**负责人**: Architect Agent  
**状态**: Planning  
**优先级**: P1 (重要)  
**预估工作量**: 3-4 周

---

## 1. Epic 概述

### 1.1 业务目标

提供用户自定义函数的部署和执行能力，支持 HTTP 触发和事件触发，实现服务端业务逻辑扩展。

### 1.2 核心价值

- 用户可编写和部署自定义函数
- 函数执行相互隔离
- 支持 HTTP 和事件双触发方式
- 自动资源管理和日志收集

### 1.3 验收标准

- [ ] 用户可创建和部署 Go 函数
- [ ] 函数可通过 HTTP 调用
- [ ] 函数执行时间 < 30s (默认)
- [ ] 函数执行相互隔离
- [ ] 函数日志可查询
- [ ] 资源限制生效（内存、超时）

---

## 2. 技术规范

> 📖 **通用规范参考**：[API 设计规范](../standards/api-design.md) | [编码规范](../standards/coding-standards.md)

### 2.1 架构设计

#### 执行模型
```
用户代码 → 编译为二进制 → 独立进程执行
                              ↓
                       stdin/stdout 通信
```

#### 函数生命周期
```
创建 → 编译 → 就绪 → 执行 → 完成
                    ↓
                  失败/超时
```

### 2.2 API 端点

| 端点 | 方法 | 功能 | 认证 |
|-----|------|------|------|
| `/api/v1/functions` | POST | 创建函数 | JWT |
| `/api/v1/functions` | GET | 列出函数 | JWT |
| `/api/v1/functions/{id}` | GET | 获取函数详情 | JWT |
| `/api/v1/functions/{id}` | PUT | 更新函数 | JWT |
| `/api/v1/functions/{id}` | DELETE | 删除函数 | JWT |
| `/api/v1/functions/{id}/invoke` | POST | 同步执行函数 | JWT |
| `/api/v1/functions/{id}/invoke-async` | POST | 异步执行函数 | JWT |
| `/api/v1/functions/executions/{id}` | GET | 查询执行结果 | JWT |
| `/api/v1/functions/{id}/logs` | GET | 获取执行日志 | JWT |

#### 示例：创建函数

**请求**：
```http
POST /api/v1/functions
Authorization: Bearer <token>
Content-Type: application/json

{
  "project_id": "proj-123",
  "name": "send-email",
  "description": "发送邮件通知",
  "runtime": "go1.21",
  "code": "package main\n\nimport \"encoding/json\"\n\nfunc Handler(input map[string]interface{}) (map[string]interface{}, error) {\n    return map[string]interface{}{\"status\": \"ok\"}, nil\n}",
  "trigger": "http",
  "timeout": 30,
  "memory": 128
}
```

**响应**：
```json
{
  "success": true,
  "code": 201,
  "data": {
    "function_id": "func-456",
    "name": "send-email",
    "status": "pending",
    "version": "1",
    "invoke_url": "/api/v1/functions/func-456/invoke",
    "created_at": "2025-12-26T10:00:00Z"
  }
}
```

#### 示例：执行函数

**请求**：
```http
POST /api/v1/functions/func-456/invoke
Authorization: Bearer <token>
Content-Type: application/json

{
  "input": {
    "email": "user@example.com",
    "subject": "Welcome"
  }
}
```

**响应**：
```json
{
  "success": true,
  "code": 200,
  "data": {
    "execution_id": "exec-789",
    "output": {
      "status": "sent",
      "message_id": "msg-123"
    },
    "duration": 150,
    "memory_used": 64,
    "logs": "2025-12-26T10:00:00Z [INFO] Sending email...\n"
  }
}
```

### 2.3 数据模型

#### 函数表（Ent Schema）
```go
// ent/schema/function.go
func (Function) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").StorageKey("id").StructTag(`json:"function_id"`),
        field.String("project_id").StorageKey("project_id").StructTag(`json:"project_id"`),
        field.String("name").StorageKey("name").StructTag(`json:"name"`),
        field.String("description").StorageKey("description").StructTag(`json:"description"`),
        field.String("runtime").StorageKey("runtime").StructTag(`json:"runtime"`),
        field.Text("code").StorageKey("code").StructTag(`json:"code"`),
        field.String("trigger").StorageKey("trigger").StructTag(`json:"trigger"`), // "http" or "event"
        field.String("status").StorageKey("status").StructTag(`json:"status"`),   // "pending", "active", "failed"
        field.Int("timeout").StorageKey("timeout").StructTag(`json:"timeout"`),
        field.Int("memory").StorageKey("memory").StructTag(`json:"memory"`),
        field.Int("version").StorageKey("version").StructTag(`json:"version"`),
        field.Time("created_at").StorageKey("created_at").StructTag(`json:"created_at"`),
    }
}
```

#### 执行记录表（Ent Schema）
```go
// ent/schema/execution.go
func (Execution) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").StorageKey("id").StructTag(`json:"execution_id"`),
        field.String("function_id").StorageKey("function_id").StructTag(`json:"function_id"`),
        field.String("status").StorageKey("status").StructTag(`json:"status"`), // "pending", "running", "completed", "failed"
        field.JSON("input", map[string]interface{}{}).StorageKey("input").StructTag(`json:"input"`),
        field.JSON("output", map[string]interface{}{}).StorageKey("output").StructTag(`json:"output"`),
        field.String("error").Optional().StorageKey("error").StructTag(`json:"error,omitempty"`),
        field.Int64("duration").StorageKey("duration").StructTag(`json:"duration"`), // 毫秒
        field.Int("memory_used").StorageKey("memory_used").StructTag(`json:"memory_used"`),
        field.Text("logs").StorageKey("logs").StructTag(`json:"logs"`),
        field.Time("started_at").StorageKey("started_at").StructTag(`json:"started_at"`),
        field.Time("completed_at").Optional().StorageKey("completed_at").StructTag(`json:"completed_at,omitempty"`),
    }
}
```

### 2.4 执行引擎设计

#### 函数编译（伪代码）
```go
func (e *ExecutionEngine) CompileFunction(ctx context.Context, fn *Function) error {
    // 1. 创建临时目录
    tmpDir, _ := os.MkdirTemp("", "func-"+fn.ID)
    defer os.RemoveAll(tmpDir)
    
    // 2. 写入函数代码
    codeFile := filepath.Join(tmpDir, "main.go")
    os.WriteFile(codeFile, []byte(fn.Code), 0644)
    
    // 3. 编译为二进制
    binaryPath := filepath.Join("/var/apprun/functions", fn.ID)
    cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, codeFile)
    
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("compile failed: %s", output)
    }
    
    return nil
}
```

#### 函数执行（伪代码）
```go
func (e *ExecutionEngine) ExecuteFunction(ctx context.Context, fn *Function, input map[string]interface{}) (*Execution, error) {
    // 1. 创建超时 Context
    execCtx, cancel := context.WithTimeout(ctx, time.Duration(fn.Timeout)*time.Second)
    defer cancel()
    
    // 2. 序列化输入
    inputJSON, _ := json.Marshal(input)
    
    // 3. 启动进程
    binaryPath := filepath.Join("/var/apprun/functions", fn.ID)
    cmd := exec.CommandContext(execCtx, binaryPath)
    cmd.Stdin = bytes.NewReader(inputJSON)
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    // 4. 执行并记录
    startTime := time.Now()
    err := cmd.Run()
    duration := time.Since(startTime)
    
    // 5. 解析输出
    var output map[string]interface{}
    json.Unmarshal(stdout.Bytes(), &output)
    
    return &Execution{
        Status:   "completed",
        Output:   output,
        Duration: duration.Milliseconds(),
        Logs:     stderr.String(),
    }, nil
}
```

### 2.5 函数代码模板

```go
// 用户函数模板
package main

import (
    "encoding/json"
    "os"
)

// Handler 是函数入口
func Handler(input map[string]interface{}) (map[string]interface{}, error) {
    // 用户业务逻辑
    email := input["email"].(string)
    
    // 返回结果
    return map[string]interface{}{
        "status":  "success",
        "message": "Email sent to " + email,
    }, nil
}

// main 函数（系统自动生成）
func main() {
    var input map[string]interface{}
    json.NewDecoder(os.Stdin).Decode(&input)
    
    output, err := Handler(input)
    if err != nil {
        os.Stderr.WriteString(err.Error())
        os.Exit(1)
    }
    
    json.NewEncoder(os.Stdout).Encode(output)
}
```

### 2.6 资源限制

| 限制项 | 免费用户 | 付费用户 |
|--------|---------|---------|
| 单函数超时 | 30 秒 | 300 秒 |
| 最大内存 | 128 MB | 512 MB |
| 并发执行数 | 10 | 100 |
| 函数数量 | 10 | 100 |

### 2.7 权限控制

| 操作 | 项目角色 | 说明 |
|-----|---------|------|
| 创建函数 | member+ | 项目成员及以上 |
| 执行函数 | viewer+ | 查看者及以上 |
| 更新函数 | member+ | 函数创建者或管理员 |
| 删除函数 | admin+ | 管理员及以上 |

### 2.8 配置

```yaml
# config/functions.yaml
functions:
  runtime:
    go_version: "1.21"
    binary_dir: "/var/apprun/functions"
    temp_dir: "/tmp/apprun-functions"
  
  limits:
    default_timeout: 30
    max_timeout: 300
    default_memory: 128
    max_memory: 512
    max_concurrent: 100
  
  logging:
    retention_days: 30
```

---

## 3. Stories 拆分

### Story 1: 函数管理基础
**优先级**: P0  
**工作量**: 3 天
- [ ] 定义函数数据模型（Ent Schema）
- [ ] 实现函数 CRUD API
- [ ] 函数状态管理
- [ ] 编写单元测试

### Story 2: 函数编译引擎
**优先级**: P0  
**工作量**: 4 天
- [ ] 实现 Go 函数编译逻辑
- [ ] 错误处理和日志收集
- [ ] 编译缓存机制
- [ ] 编写编译测试

### Story 3: 函数执行引擎
**优先级**: P0  
**工作量**: 5 天
- [ ] 实现进程隔离执行
- [ ] 超时控制
- [ ] 输入输出处理
- [ ] 资源限制（内存）
- [ ] 编写执行测试

### Story 4: HTTP 触发功能
**优先级**: P0  
**工作量**: 2 天
- [ ] 实现同步执行接口
- [ ] 实现异步执行接口
- [ ] 执行结果查询
- [ ] 编写 HTTP 触发测试

### Story 5: 执行日志管理
**优先级**: P1  
**工作量**: 2 天
- [ ] 日志收集和存储
- [ ] 日志查询接口
- [ ] 日志清理策略
- [ ] 编写日志测试

### Story 6: 事件触发功能
**优先级**: P2  
**工作量**: 3 天
- [ ] 集成事件总线
- [ ] 函数事件订阅
- [ ] 事件触发逻辑
- [ ] 编写事件触发测试

---

## 4. 依赖关系

### 技术依赖
- Go 编译器 (1.21+)
- os/exec (进程管理)
- Ent ORM (数据库)

### 模块依赖
- 认证模块（权限验证）
- 配置模块（函数配置）
- 事件模块（事件触发，可选）

### 外部依赖
- 文件系统（存储编译后的二进制）
- PostgreSQL（元数据）

---

## 5. 风险与挑战

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| 函数执行超时 | 中 | 严格超时控制，自动杀死进程 |
| 内存泄漏 | 高 | 进程隔离，执行完自动清理 |
| 恶意代码执行 | 高 | 沙箱环境，资源限制 |
| 编译失败 | 中 | 详细错误提示，代码模板 |

---

## 6. 测试策略

### 单元测试
- 函数编译逻辑
- 函数执行逻辑
- 输入输出处理
- 超时控制

### 集成测试
- 完整函数生命周期（创建 → 编译 → 执行）
- HTTP 触发流程
- 权限验证场景

### 性能测试
- 函数执行延迟 < 100ms（简单函数）
- 并发执行 10 个函数
- 内存使用监控

---

## 7. 监控指标

- `functions_invocations_total` - 函数调用总数
- `functions_duration_seconds` - 函数执行时长
- `functions_errors_total` - 函数执行失败次数
- `functions_compile_duration_seconds` - 函数编译耗时

---

## 附录

### A. 错误码定义

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| `FUNC_NOT_FOUND` | 404 | 函数不存在 |
| `FUNC_COMPILE_FAILED` | 500 | 函数编译失败 |
| `FUNC_EXEC_TIMEOUT` | 504 | 函数执行超时 |
| `FUNC_EXEC_FAILED` | 500 | 函数执行失败 |
| `FUNC_QUOTA_EXCEEDED` | 429 | 函数执行配额超限 |

### B. 相关文档

- [PRD - 函数服务](../prd.md#24-函数服务)
- [API 设计规范](../standards/api-design.md)

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26
