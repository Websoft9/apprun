# 资源管理模块设计文档

## 概述

资源管理模块是一个基于DSL（Domain Specific Language）的动态资源管理系统，专门为DevOps流程设计，支持抽象和管理各种云资源和SaaS服务。该模块采用插件化架构，通过YAML DSL定义资源类型，实现完全动态扩展，无需修改代码即可支持新资源类型。

## 核心特性

- **DSL驱动**：通过YAML文件定义资源类型结构，无代码修改即可扩展
- **插件化架构**：每个资源提供商实现统一插件接口
- **动态验证**：运行时基于DSL验证资源配置
- **统一存储**：所有资源实例存储在Ent数据库中
- **依赖管理**：支持资源间的依赖关系

## 架构设计

### 核心组件

```
DSL Parser → Resource Manager → Plugins → Ent Storage
     ↓              ↓              ↓         ↓
类型定义    资源生命周期    提供商API    数据库持久化
```

### 组件职责

1. **DSL Parser**：解析YAML定义，构建运行时类型信息
2. **Resource Manager**：统一管理资源CRUD操作
3. **Resource Plugin**：实现具体提供商的资源操作
4. **Ent Storage**：持久化存储资源实例和状态

## DSL定义规范

### 资源类型定义

```yaml
type: server                    # 资源类型标识
name: 云服务器                  # 显示名称
description: 云平台虚拟服务器实例  # 描述
version: "1.0"                 # 版本号

fields:                        # 字段定义
  - name: region               # 字段名
    type: string              # 字段类型
    required: true            # 是否必填
    description: 部署地域      # 字段描述
    options: ["us-east-1", "cn-beijing"]  # 枚举选项
    validation: "required"    # 验证规则

  - name: instance_type
    type: string
    required: true
    description: 实例规格
    validation: "required,min=1,max=50"

providers:                     # 支持的提供商
  - aws
  - alibaba
  - tencent

lifecycle:                     # 生命周期钩子
  create: "create_server"
  update: "update_server"
  delete: "delete_server"
  status: "get_server_status"
```

### 资源实例定义

```yaml
type: server                   # 资源类型
name: web-server-01           # 实例名称
provider: aws                 # 提供商
config:                       # 配置数据
  region: us-east-1
  instance_type: t3.micro
  ami: ami-12345678
```

## Ent Schema设计

### 设计原则：资源信息与操作分离

资源管理采用**关注点分离**原则，将资源信息（静态）和资源操作（动态）存储在不同的表中：

- **Resource表**：存储资源的当前状态、配置和属性（静态数据）
- **ResourceOperation表**：记录所有操作历史、审计追踪（动态数据）

这种设计的优势：
1. **清晰的职责分离**：资源"是什么"与"做了什么"独立管理
2. **完整的审计追踪**：保留所有操作历史，支持回溯和分析
3. **支持异步操作**：操作记录独立于资源状态，便于异步任务管理
4. **查询性能优化**：不同查询模式使用不同索引，避免相互干扰

### 资源信息表

```go
type Resource struct {
    ent.Schema
}

func (Resource) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").Unique(),                    // 资源唯一ID
        field.String("type").NotEmpty(),                // 资源类型
        field.String("name").NotEmpty(),                // 资源名称
        field.String("provider").NotEmpty(),            // 提供商
        field.JSON("config", map[string]interface{}{}), // 配置数据
        field.JSON("state", map[string]interface{}{}),  // 运行时状态
        field.String("status").Default("pending"),      // 当前状态
        field.String("external_id"),                    // 云平台资源ID
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now),
    }
}

func (Resource) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("dependencies", Resource.Type).From("dependents"),
        edge.To("operations", ResourceOperation.Type),  // 关联操作记录
    }
}
```

### 资源操作表

```go
type ResourceOperation struct {
    ent.Schema
}

func (ResourceOperation) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").Unique(),
        field.String("resource_id").NotEmpty(),         // 关联资源
        field.String("operation").NotEmpty(),           // 操作类型: create, update, delete, start, stop
        field.String("status").Default("pending"),      // pending, running, success, failed
        field.String("operator").Optional(),            // 操作者
        field.JSON("input", map[string]interface{}{}),  // 输入参数
        field.JSON("output", map[string]interface{}{}), // 输出结果
        field.String("error_message").Optional(),       // 错误信息
        field.Time("started_at").Default(time.Now),
        field.Time("completed_at").Optional(),
        field.String("job_id").Optional(),              // 异步任务ID
    }
}

func (ResourceOperation) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("resource", Resource.Type).
            Ref("operations").
            Unique().
            Required(),
    }
}
```

## 核心接口

### ResourcePlugin接口

```go
type ResourcePlugin interface {
    Create(ctx context.Context, config map[string]interface{}) (*ResourceInstance, error)
    Update(ctx context.Context, id string, config map[string]interface{}) error
    Delete(ctx context.Context, id string) error
    Status(ctx context.Context, id string) (string, error)
    
    // 生命周期操作
    Start(ctx context.Context, id string) error
    Stop(ctx context.Context, id string) error
    Restart(ctx context.Context, id string) error
}

type ResourceInstance struct {
    ID         string
    ExternalID string
    Status     string
    State      map[string]interface{}
}
```

### DSLParser接口

```go
type DSLParser struct {
    types     map[string]*ResourceType
    validator *validator.Validate
}

func (p *DSLParser) LoadType(filePath string) error
func (p *DSLParser) ValidateConfig(resourceType string, config map[string]interface{}) error
```

### ResourceManager接口

```go
type Manager struct {
    client    *ent.Client
    parser    *DSLParser
    plugins   map[string]ResourcePlugin
}

func (m *Manager) CreateResource(typeName, name, provider string, config map[string]interface{}) (*ent.Resource, error)
func (m *Manager) GetResource(id string) (*ent.Resource, error)
func (m *Manager) ListResources(typeFilter, providerFilter string) ([]*ent.Resource, error)
func (m *Manager) GetResourceOperations(id string) ([]*ent.ResourceOperation, error)  // 查询操作历史
func (m *Manager) StartResource(id, operator string) (*ent.ResourceOperation, error)  // 启动资源
func (m *Manager) StopResource(id, operator string) (*ent.ResourceOperation, error)   // 停止资源
```

## 使用流程

### 1. 定义资源类型
创建 `resources/types/{type}.yaml` 文件定义新资源类型。

### 2. 实现插件
为每个提供商实现 `ResourcePlugin` 接口。

### 3. 注册组件
```go
// 注册插件
resourceManager.RegisterPlugin("aws", &AWSPlugin{})

// 加载类型定义
resourceManager.LoadResourceTypes("./resources/types")
```

### 4. 创建资源实例
```go
resource, err := resourceManager.CreateResource("server", "web-01", "aws", config)
```

### 5. API集成
提供REST API进行资源管理：
- `POST /api/resources` - 创建资源
- `GET /api/resources` - 列出资源
- `GET /api/resources/{id}` - 获取资源详情
- `PUT /api/resources/{id}` - 更新资源
- `DELETE /api/resources/{id}` - 删除资源
- `GET /api/resources/{id}/operations` - 获取资源操作历史
- `POST /api/resources/{id}/start` - 启动资源
- `POST /api/resources/{id}/stop` - 停止资源
- `POST /api/resources/{id}/restart` - 重启资源

## 支持的资源类型

### 当前支持
1. **云服务器 (server)** - AWS EC2, Alibaba ECS, Tencent CVM
2. **云数据库 (database)** - RDS, Cloud DB
3. **SMTP服务 (smtp)** - SendGrid, AWS SES, Alibaba DirectMail
4. **云平台 (platform)** - Kubernetes, Docker Swarm

### 扩展新资源
1. 创建DSL定义文件
2. 实现对应插件
3. 注册到管理器

## 优势

1. **完全动态**：无需代码修改即可添加新资源类型
2. **类型安全**：运行时验证确保配置正确性
3. **统一管理**：所有资源通过统一接口管理
4. **插件扩展**：新提供商只需实现插件接口
5. **依赖管理**：支持资源间的依赖关系建模
6. **可观测性**：完整的状态跟踪和生命周期管理
7. **审计追踪**：资源操作历史完整记录，支持回溯和分析
8. **异步支持**：操作与状态分离，天然支持异步任务管理

## 部署架构

```
DSL Files ──→ DSL Parser ──→ Resource Manager
                    │                │
                    └───────────────→ Plugins
                                     │
                                     └───────────────→ Ent Database
```

## 安全考虑

- **配置加密**：敏感配置使用加密存储
- **访问控制**：基于角色的资源访问权限
- **审计日志**：记录所有资源操作
- **网络安全**：插件与云提供商通信使用HTTPS

## 监控和运维

- **状态监控**：定期检查资源状态
- **告警机制**：资源异常时发送告警
- **日志记录**：详细的操作日志
- **性能监控**：资源创建/更新耗时统计

## 未来扩展

- **模板系统**：预定义资源模板
- **工作流引擎**：支持复杂部署流程
- **多租户支持**：资源隔离和配额管理
- **事件驱动**：基于事件的资源自动化