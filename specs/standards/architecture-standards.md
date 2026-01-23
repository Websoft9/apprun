# 架构设计规范
# apprun BaaS Platform

**创建日期**: 2025-12-26  
**维护者**: Winston (Architect Agent)  
**版本**: 1.0  
**状态**: Active

---

## 概述

本文档定义 apprun 项目的架构设计原则和规范，确保系统设计的一致性、可维护性和可扩展性。

**核心原则**：
- **简单优先** - 能简单解决就不复杂化
- **演进式设计** - 支持从单体到微服务的平滑演进
- **云中立** - 不绑定特定云服务商
- **安全内建** - 安全设计从一开始就考虑

---

## 1. 模块化设计

### 1.1 解耦原则

**依赖接口，而非实现**

```go
// ❌ 错误：直接依赖具体实现
type UserService struct {
    db *sql.DB  // 紧耦合
}

// ✅ 正确：依赖接口
type UserService struct {
    repo UserRepository  // 接口
}

type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id int) (*User, error)
}
```

**规范**：
- 模块间通过接口交互
- 核心业务逻辑不依赖外部库
- 使用依赖注入

### 1.2 分层架构

```
┌─────────────────────────────────┐
│  Handler (HTTP/gRPC)            │  ← 协议层
├─────────────────────────────────┤
│  Service (业务逻辑)              │  ← 业务层
├─────────────────────────────────┤
│  Repository (数据访问)           │  ← 持久层
└─────────────────────────────────┘
```

**规范**：
- 层间依赖单向（Handler → Service → Repository）
- 禁止跨层调用
- 每层职责单一

### 1.3 模块独立性

```
core/
├── auth/           # 认证模块
├── storage/        # 存储模块
├── workflow/       # 工作流模块
└── pkg/            # 共享包
    ├── response/
    └── errors/
```

**规范**：
- 每个模块可独立测试
- 模块间通过 pkg 共享通用代码
- 避免循环依赖

---

## 2. 扩展性设计

### 2.1 插件化

**策略模式实现**

```go
// 定义插件接口
type StoragePlugin interface {
    Upload(ctx context.Context, path string, data io.Reader) error
    Download(ctx context.Context, path string) (io.Reader, error)
}

// 注册插件
var storagePlugins = map[string]StoragePlugin{
    "local": &LocalStorage{},
    "s3":    &S3Storage{},
}

// 运行时选择
storage := storagePlugins[config.Type]
```

**规范**：
- 使用接口定义扩展点
- 支持配置切换
- 插件失败不影响核心

### 2.2 配置驱动

**功能开关**

```yaml
# config.yaml
features:
  realtime: true
  workflow: false
  cache: true
```

```go
if config.Features.Realtime {
    startRealtimeServer()
}
```

**规范**：
- 核心功能可配置开关
- 支持灰度发布
- 变更无需重编译

---

## 3. 非侵入式设计

### 3.1 中间件模式

```go
// 核心 Handler
func GetUser(w http.ResponseWriter, r *http.Request) {
    // 业务逻辑
}

// 通过中间件扩展
router.Use(
    middleware.Logger,
    middleware.Auth,
    middleware.RateLimit,
)
```

**规范**：
- 横切关注点使用中间件
- 中间件可组合
- 不污染业务代码

### 3.2 事件驱动

```go
// 发布事件
eventBus.Publish("user.created", event)

// 订阅事件
eventBus.Subscribe("user.created", sendWelcomeEmail)
```

**规范**：
- 模块间通过事件解耦
- 发布者不依赖订阅者
- 支持多订阅者

---

## 4. 协议设计

### 4.1 API 协议选择

| 场景 | 协议 |
|-----|------|
| CRUD 操作 | RESTful |
| 实时推送 | WebSocket |
| 内部服务 | gRPC（可选） |

**规范**：
- 外部 API 使用 RESTful + JSON
- 实时功能使用 WebSocket
- 遵循 [API 设计规范](./api-design.md)

### 4.2 数据格式

```json
// 统一响应格式
{
  "success": true,
  "data": {...}
}
```

**规范**：
- 使用 JSON 作为数据交换格式
- 字段命名使用 snake_case
- 遵循统一响应格式

---

## 5. 隔离设计

### 5.1 多租户隔离

```go
// 所有表包含 project_id
type Resource struct {
    ID        int
    ProjectID int     // 租户隔离
    Name      string
}

// 查询自动过滤
func (r *Repo) List(ctx context.Context, projectID int) ([]*Resource, error) {
    return r.db.Resource.Query().
        Where(resource.ProjectID(projectID)).  // 强制隔离
        All(ctx)
}
```

**规范**：
- 所有资源表包含 `project_id`
- 查询强制添加租户条件
- 中间件自动注入租户上下文

### 5.2 资源配额

```go
// 配额检查
func (s *Service) Upload(ctx context.Context, file File) error {
    if s.exceedsQuota(ctx, file) {
        return errors.New("quota exceeded")
    }
    return s.repo.Save(ctx, file)
}
```

**规范**：
- 每个租户有独立配额
- 操作前检查配额
- 超限返回明确错误

---

## 6. 演进路径

### 6.1 单体 → 微服务

**现在（模块化单体）**：
```
┌──────────────────────┐
│  apprun              │
│  ├── Auth Module     │
│  ├── Storage Module  │
│  └── Data Module     │
└──────────────────────┘
```

**未来（微服务）**：
```
┌─────┐  ┌─────┐  ┌─────┐
│Auth │  │Store│  │Data │
└─────┘  └─────┘  └─────┘
```

**演进步骤**：
1. 模块化单体（清晰边界）
2. 独立数据库 Schema
3. 提取为独立服务
4. 独立部署

**规范**：
- 模块间禁止直接调用内部方法
- 使用明确的接口契约
- 预留服务间通信机制

---

## 7. 设计检查清单

### 新模块设计

- [ ] 是否定义了清晰的接口？
- [ ] 是否可以独立测试？
- [ ] 是否支持配置切换？
- [ ] 是否考虑了多租户隔离？
- [ ] 是否有扩展点？

### 接口设计

- [ ] 是否遵循 RESTful 规范？
- [ ] 是否有统一的响应格式？
- [ ] 是否有完善的错误处理？
- [ ] 是否考虑了性能优化？
- [ ] 是否有权限控制？

### 数据库设计

- [ ] 是否包含 `project_id`？
- [ ] 是否有适当的索引？
- [ ] 是否考虑了数据迁移？
- [ ] 是否有软删除支持？
- [ ] 是否符合 Ent Schema 规范？

---

## 8. 反模式（避免）

### 8.1 紧耦合

```go
// ❌ 错误
func Handler(w http.ResponseWriter, r *http.Request) {
    db.Query("SELECT ...")  // Handler 直接操作数据库
}

// ✅ 正确
func Handler(w http.ResponseWriter, r *http.Request) {
    data, err := service.Get(ctx, id)  // 通过 Service 层
}
```

### 8.2 上帝类

```go
// ❌ 错误
type AppService struct {
    // 处理所有业务
}

// ✅ 正确
type UserService struct{}
type ProjectService struct{}
type StorageService struct{}
```

### 8.3 硬编码

```go
// ❌ 错误
if storageType == "s3" { ... }

// ✅ 正确
storage := factory.Create(config.Type)
storage.Upload(file)
```

---

## 相关文档

- [技术架构](../architecture/tech-architecture.md) - apprun 具体技术架构
- [部署架构](../architecture/deployment-architecture.md) - 部署方案
- [API 设计规范](./api-design.md) - API 设计规则
- [编码规范](./coding-standards.md) - Go 代码规范

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26
