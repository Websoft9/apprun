# 编码规范
# apprun BaaS Platform

**创建日期**: 2025-12-25  
**维护者**: Winston (Architect Agent)  
**版本**: 1.0  
**状态**: Draft

---

## 1. Go 编码规范

### 1.1 基本原则

- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 和 `goimports` 格式化代码
- 使用 `golangci-lint` 进行静态检查
- 所有的代码和注释都采用英文
- **统一使用公共包**：
  - **错误处理**: `apprun/pkg/errors` - 统一错误码和错误包装
  - **API 响应**: `apprun/pkg/response` - 统一 HTTP 响应格式
  - **日志处理**: `apprun/pkg/logger` - 统一日志输出

### 1.2 命名规范

#### 1.2.1 包名

```go
// ✅ 推荐：小写单词，简短有意义
package user
package storage
package cache

// ❌ 避免：下划线、大写、复数
package user_service  // 使用 package userservice
package User          // 使用 package user
package users         // 使用 package user
```

#### 1.2.2 变量和函数名

```go
// ✅ 推荐：驼峰命名
var userName string
var userID int
func getUserByID(id int) (*User, error)

// 导出的变量/函数使用大写开头
var DefaultTimeout = 30 * time.Second
func NewUserService() *UserService

// ❌ 避免：下划线分隔
var user_name string      // 使用 userName
func get_user_by_id()     // 使用 getUserByID
```

#### 1.2.3 常量

```go
// ✅ 推荐：驼峰或全大写（根据上下文）
const (
    MaxRetries = 3
    DefaultPageSize = 20
)

const (
    StatusActive   = "active"
    StatusInactive = "inactive"
)

// 枚举类型常量
type Status string

const (
    StatusPending   Status = "pending"
    StatusRunning   Status = "running"
    StatusCompleted Status = "completed"
)
```

#### 1.2.4 接口名

```go
// ✅ 推荐：以 -er 结尾
type Reader interface {
    Read(p []byte) (n int, err error)
}

type UserRepository interface {
    FindByID(id string) (*User, error)
    Save(user *User) error
}

// 单方法接口直接使用方法名 + er
type Closer interface {
    Close() error
}
```

---

## 2. 项目结构

### 2.1 模块化单体架构（推荐）

**按功能模块垂直切分，每个模块高内聚低耦合**

```
apprun/
├── core/                       # Go 应用核心代码
│   ├── main.go                # 主入口（package main）
│   ├── cmd/                   # Cobra 命令实现（package cmd）
│   │   ├── root.go           # 根命令 + 全局 flags
│   │   ├── serve.go          # apprun serve（启动服务）
│   │   ├── migrate.go        # apprun migrate（数据库迁移）
│   │   └── version.go        # apprun version（版本信息）
│   │
│   ├── modules/              # 业务模块（模块化单体）
│   │   ├── config/          # 配置管理模块
│   │   │   ├── handler.go   # HTTP API
│   │   │   ├── service.go   # 业务逻辑
│   │   │   ├── repository.go # 数据访问
│   │   │   └── types.go     # 领域模型
│   │   │
│   │   ├── auth/            # 认证授权模块
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   └── config.go    # 模块配置
│   │   │
│   │   └── user/            # 用户模块
│   │       ├── handler.go
│   │       ├── service.go
│   │       ├── repository.go
│   │       └── types.go
│   │
│   ├── internal/            # 内部基础设施（非业务模块）
│   │   ├── bootstrap/       # 启动编排逻辑
│   │   │   └── server.go   # 服务器启动 + Swagger 注释
│   │   ├── middleware/      # 中间件
│   │   ├── jwt/            # JWT 认证
│   │   ├── rbac/           # RBAC 权限
│   │   └── password/       # 密码加密
│   │
│   ├── pkg/                 # 可复用工具包（通用库）
│   │   ├── version/        # 版本管理
│   │   ├── database/       # 数据库客户端
│   │   ├── cache/          # 缓存客户端
│   │   ├── logger/         # 日志库
│   │   ├── errors/         # 错误处理
│   │   ├── response/       # 统一响应
│   │   └── i18n/           # 国际化
│   │
│   ├── ent/                # Ent ORM
│   │   └── schema/
│   │
│   ├── routes/             # 路由配置
│   ├── handlers/           # HTTP 处理器
│   ├── docs/               # Swagger 文档（自动生成）
│   ├── config/             # 配置文件
│   │   ├── default.yaml
│   │   └── conf_d/
│   └── bin/                # 编译产物（.gitignore）
│       ├── apprun         # CLI 可执行文件
│       └── server         # 向后兼容符号链接
│
├── docker/                 # Docker 配置
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── docs/                   # 项目文档
│   ├── architecture/      # 架构文档
│   ├── standards/         # 编码规范
│   └── sprint-artifacts/  # 迭代产物
│
├── tests/                  # 测试（E2E/集成测试）
│   ├── e2e/
│   └── integration/
│
├── scripts/                # 辅助脚本
├── examples/               # 示例配置
├── Makefile               # 构建入口（根目录唯一）
└── README.md
```

**优势**:
- ✅ 模块边界清晰，易于理解和维护
- ✅ CLI 扁平化结构，符合 Go 标准项目布局
- ✅ 启动逻辑在 `internal/bootstrap/`，符合分层规范
- ✅ 便于独立测试和部署
- ✅ 未来可无缝拆分为微服务

**关键设计决策**：
- **main.go 在 core/ 根目录**：避免包冲突，Go 标准做法
- **cmd/ 包含 Cobra 命令**：扁平化结构，不使用 cmd/cli/ 子目录
- **internal/bootstrap/**：应用启动编排，依赖业务模块
- **pkg/**：通用可复用库，不依赖业务逻辑
- **modules/**：业务模块，垂直切分

### 2.3 常量组织规范 (Constants Organization)

**决策日期**: 2026-01-12  
**决策背景**: BMad Method 强调"业务内聚优于文件类型分离"

#### 2.3.1 基本原则

**所有模块常量统一定义在各自的 `config.go` 文件中**，而不是单独创建 `constants.go`。

**原因**:
1. **业务内聚**: 常量与配置在语义上相关（验证规则、默认值、业务枚举）
2. **可发现性**: 新开发者在一个文件找到所有配置相关项
3. **维护性**: 修改验证规则时只需编辑一个文件
4. **Config Center 兼容**: 常量和配置自然共存

#### 2.3.2 文件组织结构

```go
// modules/auth/config.go
package auth

import "time"

// ============================================================================
// Module Constants (Validation Rules & Enums)
// ============================================================================

const (
    // Password validation rules
    MinPasswordLength = 8
    MaxPasswordLength = 128
    
    // Username validation rules
    MinUsernameLength = 3
    MaxUsernameLength = 64
    
    // Account status codes
    StatusActive   int8 = 1
    StatusDisabled int8 = 2
    StatusPending  int8 = 3
    
    // Gender codes
    GenderUnknown int8 = 0
    GenderMale    int8 = 1
    GenderFemale  int8 = 2
    
    // Default values
    DefaultBcryptCost          = 10
    DefaultMaxFailedAttempts   = 5
    DefaultFailedLoginCacheTTL = 5 * time.Minute
)

// ============================================================================
// Runtime Configuration (Config Center Managed)
// ============================================================================

type Config struct {
    JWT      jwt.Config     `yaml:"jwt"`
    Security SecurityConfig `yaml:"security"`
}

type SecurityConfig struct {
    BcryptCost        int           `yaml:"bcrypt_cost" default:"10" db:"true" validate:"min=4,max=31"`
    MaxFailedAttempts int           `yaml:"max_failed_attempts" default:"5" db:"true"`
    // ...
}

func DefaultConfig() *Config {
    return &Config{
        Security: SecurityConfig{
            BcryptCost:        DefaultBcryptCost,  // ← References constant
            MaxFailedAttempts: DefaultMaxFailedAttempts,
        },
    }
}
```

#### 2.3.3 命名约定

| 常量类型 | 命名模式 | 示例 |
|---------|---------|------|
| **最小值** | `Min<Name>` | `MinPasswordLength`, `MinUserAge` |
| **最大值** | `Max<Name>` | `MaxPasswordLength`, `MaxRetries` |
| **默认值** | `Default<Name>` | `DefaultTimeout`, `DefaultBcryptCost` |
| **状态码** | `Status<Name>` | `StatusActive`, `StatusPending` |
| **错误码** | `Err<Name>` | `ErrInvalidEmail`, `ErrUserNotFound` |
| **类型码** | `Type<Name>` | `TypeAdmin`, `TypeGuest` |

#### 2.3.4 何时使用独立的 constants.go

**仅在以下情况使用独立文件**:
1. 模块有 **50+ 个常量**（极少见）
2. 常量需要 **跨多个子包共享**
3. 常量需要 **复杂的初始化逻辑**（如计算、组合）

**示例** (大型模块，需要独立文件):
```go
// modules/workflow/constants.go (>50 常量)
package workflow

const (
    // 状态机状态 (20+ 个状态)
    StateInit       = "init"
    StatePending    = "pending"
    StateRunning    = "running"
    // ... 50+ more states
    
    // 动作类型 (20+ 个动作)
    ActionCreate    = "create"
    ActionUpdate    = "update"
    // ... 50+ more actions
)
```

#### 2.3.5 使用示例

```go
// ✅ 推荐：从 config 包导入常量
import "apprun/modules/auth"

func ValidatePassword(pwd string) error {
    if len(pwd) < auth.MinPasswordLength {
        return errors.New("password too short")
    }
    if len(pwd) > auth.MaxPasswordLength {
        return errors.New("password too long")
    }
    return nil
}

// ✅ 推荐：使用状态常量
user.Status = auth.StatusActive

// ❌ 避免：硬编码魔术数字
user.Status = 1  // 应使用 auth.StatusActive
```

#### 2.3.6 实际项目应用

**已应用此规范的模块**:
- ✅ `pkg/jwt/config.go` - JWT 常量（MinSecretLength, DefaultExpiry, DefaultWhitelistPaths）
- ✅ `modules/auth/config.go` - Auth 模块常量（密码规则、状态码、性别码、bcrypt配置）

**未来模块应遵循此模式**:
- `modules/user/config.go` - 用户模块常量
- `modules/project/config.go` - 项目模块常量
- `pkg/logger/config.go` - 日志模块常量

---

## 3. 代码风格

```
modules/config/
├── handler.go         # HTTP 层：路由和请求处理
├── service.go         # 业务逻辑层：核心业务规则
├── repository.go      # 数据访问层：数据库 CRUD
├── types.go           # 领域模型：请求/响应结构
└── config_test.go     # 模块测试
```

**分层职责**：
- `handler` - 处理 HTTP 请求，调用 service
- `service` - 实现业务逻辑，调用 repository
- `repository` - 封装数据访问，使用 Ent Client
- `types` - 定义领域模型和 DTO

---

## 3. 代码风格

### 3.1 函数设计

```go
// ✅ 推荐：函数参数不超过 3-4 个
func CreateUser(ctx context.Context, name, email string) (*User, error) {
    // ...
}

// ❌ 避免：过多参数
func CreateUser(ctx context.Context, name, email, phone, address, city, country string) (*User, error) {
    // ...
}

// ✅ 推荐：使用 struct 封装多个参数
type CreateUserInput struct {
    Name    string
    Email   string
    Phone   string
    Address string
    City    string
    Country string
}

func CreateUser(ctx context.Context, input *CreateUserInput) (*User, error) {
    // ...
}
```

### 3.2 错误处理

**必须使用统一的错误处理包**: `apprun/pkg/errors`

```go
import "apprun/pkg/errors"

// ✅ 推荐：使用 pkg/errors 创建业务错误
func GetUser(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, errors.New(errors.ErrCodeInvalidParam, "User ID required")
    }
    
    user, err := repo.FindByID(ctx, id)
    if err != nil {
        // 包装底层错误，添加上下文
        return nil, errors.Wrap(err, errors.ErrCodeInternalError, "Failed to find user").
            WithContext(errors.ContextKeyUserID, id)
    }
    
    if user == nil {
        return nil, errors.New(errors.ErrCodeNotFound, "User not found").
            WithContext(errors.ContextKeyUserID, id)
    }
    
    return user, nil
}

// ✅ 推荐：检查错误类型
if errors.IsNotFound(err) {
    // 处理资源不存在
}

if errors.IsValidation(err) {
    // 处理验证错误
}

// ✅ 推荐：HTTP 层映射状态码
import "apprun/pkg/errors/httpmap"

func HandleError(w http.ResponseWriter, err error) {
    status := httpmap.ToHTTPStatus(err)
    w.WriteHeader(status)
    
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error":   appErr.Code,
            "message": appErr.Message,
        })
    }
}

// ❌ 避免：直接使用 fmt.Errorf 或 errors.New (stdlib)
func GetUser(id string) error {
    return fmt.Errorf("user not found")  // 不推荐
}
```

**错误处理规范**：
- 所有业务错误必须使用 `pkg/errors` 创建
- 使用预定义错误码常量（见 `pkg/errors/codes.go`）
- 底层错误必须用 `Wrap/Wrapf` 包装，保留错误链
- 添加有用的上下文信息（user_id, request_id 等）
- HTTP 层使用 `httpmap.ToHTTPStatus` 映射状态码

### 3.3 上下文使用

```go
// ✅ 推荐：Context 作为第一个参数
func ProcessData(ctx context.Context, data []byte) error {
    // ...
}

// ✅ 推荐：从 Context 获取值
func GetUserFromContext(ctx context.Context) (*User, error) {
    user, ok := ctx.Value("user").(*User)
    if !ok {
        return nil, errors.New("user not found in context")
    }
    return user, nil
}

// ✅ 推荐：Context 超时控制
func FetchData(ctx context.Context) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // 执行操作
    return data, nil
}
```

### 3.4 并发编程

```go
// ✅ 推荐：使用 sync.WaitGroup
func ProcessItems(items []Item) {
    var wg sync.WaitGroup
    
    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            processItem(i)
        }(item)  // 注意：传递副本避免闭包问题
    }
    
    wg.Wait()
}

// ✅ 推荐：使用 Channel 通信
func Producer(ch chan<- int) {
    defer close(ch)
    for i := 0; i < 10; i++ {
        ch <- i
    }
}

func Consumer(ch <-chan int) {
    for val := range ch {
        fmt.Println(val)
    }
}

// ✅ 推荐：使用 sync.Once 保证单次执行
var (
    instance *Singleton
    once     sync.Once
)

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}
```

---

## 4. 分层架构

### 4.1 Handler 层

```go
// internal/handler/user.go

type UserHandler struct {
    service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
    return &UserHandler{service: service}
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // 1. 解析参数
    userID := chi.URLParam(r, "id")
    if userID == "" {
        render.JSON(w, r, ErrorResponse(400, "user_id is required"))
        return
    }
    
    // 2. 调用 Service
    user, err := h.service.GetUser(r.Context(), userID)
    if err != nil {
        render.JSON(w, r, ErrorResponse(500, err.Error()))
        return
    }
    
    // 3. 返回响应
    render.JSON(w, r, SuccessResponse(user))
}
```

### 4.2 Service 层

```go
// internal/service/user.go

type UserService struct {
    repo  *repository.UserRepository
    cache *cache.MultiLevelCache
}

func NewUserService(repo *repository.UserRepository, cache *cache.MultiLevelCache) *UserService {
    return &UserService{
        repo:  repo,
        cache: cache,
    }
}

// GetUser 获取用户（带缓存）
func (s *UserService) GetUser(ctx context.Context, id string) (*model.User, error) {
    // 1. 查询缓存
    cacheKey := fmt.Sprintf("user:%s", id)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        return cached.(*model.User), nil
    }
    
    // 2. 查询数据库
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    
    // 3. 写入缓存
    s.cache.Set(ctx, cacheKey, user)
    
    return user, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, input *model.CreateUserInput) (*model.User, error) {
    // 1. 业务校验
    if err := s.validateUser(input); err != nil {
        return nil, err
    }
    
    // 2. 创建用户
    user, err := s.repo.Create(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // 3. 发布事件
    s.eventBus.Publish(ctx, "user.created", map[string]interface{}{
        "user_id": user.ID,
        "email":   user.Email,
    })
    
    return user, nil
}

func (s *UserService) validateUser(input *model.CreateUserInput) error {
    if input.Name == "" {
        return errors.New("name is required")
    }
    if !isValidEmail(input.Email) {
        return errors.New("invalid email format")
    }
    return nil
}
```

### 4.3 Repository 层

```go
// internal/repository/user.go

type UserRepository struct {
    client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepository {
    return &UserRepository{client: client}
}

// FindByID 根据 ID 查询用户
func (r *UserRepository) FindByID(ctx context.Context, id string) (*ent.User, error) {
    return r.client.User.
        Query().
        Where(user.IDEQ(id)).
        Only(ctx)
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, input *model.CreateUserInput) (*ent.User, error) {
    return r.client.User.
        Create().
        SetName(input.Name).
        SetEmail(input.Email).
        Save(ctx)
}

// List 列表查询（带分页）
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]*ent.User, int, error) {
    // 查询总数
    total, err := r.client.User.Query().Count(ctx)
    if err != nil {
        return nil, 0, err
    }
    
    // 分页查询
    users, err := r.client.User.Query().
        Limit(pageSize).
        Offset((page - 1) * pageSize).
        Order(ent.Desc(user.FieldCreatedAt)).
        All(ctx)
    
    return users, total, err
}
```

---

## 5. 注释规范

### 5.1 包注释

```go
// Package user provides user management functionality.
// It includes user CRUD operations, authentication, and authorization.
package user
```

### 5.2 函数注释

```go
// GetUser retrieves a user by ID.
// It returns an error if the user is not found or if there's a database error.
//
// Example:
//
//	user, err := service.GetUser(ctx, "123")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    // ...
}
```

### 5.3 类型注释

```go
// User represents a user in the system.
// It contains basic user information and authentication details.
type User struct {
    ID        string    `json:"id"`         // Unique user identifier
    Name      string    `json:"name"`       // User's full name
    Email     string    `json:"email"`      // User's email address
    CreatedAt time.Time `json:"created_at"` // Account creation timestamp
}
```

### 5.4 TODO 注释

```go
// TODO(username): Add input validation
// TODO: Implement retry logic
// FIXME: Memory leak in this function
// HACK: Temporary workaround for issue #123
```

---

## 6. 测试规范

### 6.1 测试文件位置

**原则：所有测试与代码同目录**

这是 Go 官方推荐的最佳实践，遵循 Go 标准库和所有知名开源项目（Kubernetes、Docker、Prometheus）的做法。

```
modules/auth/
├── handler/
│   ├── login.go
│   └── login_test.go              # 单元测试
├── service/
│   ├── auth_service.go
│   ├── auth_service_test.go       # 单元测试
│   └── auth_integration_test.go   # 集成测试
└── repository/
    ├── user_repo.go
    └── user_repo_test.go          # 单元测试
```

**区分单元测试和集成测试**：

- **单元测试**：`xxx_test.go` - 测试单个函数/方法，使用 mock
- **集成测试**：`xxx_integration_test.go` - 测试完整流程，使用真实数据库

**运行方式**：

```bash
# 只运行单元测试（快速，CI 默认）
go test -short ./...

# 只运行集成测试
go test -run Integration ./...

# 运行所有测试
go test ./...
```

**集成测试标准模板**：

```go
//go:build integration
// +build integration

package auth_test  // 使用 _test 后缀，只测试公开 API

func TestLoginIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    // 测试代码
}
```

**优势**：
- ✅ Go 官方推荐，符合社区标准
- ✅ 可以测试包内私有函数/方法
- ✅ `go test ./...` 自动发现所有测试
- ✅ 测试覆盖率统计自动关联
- ✅ IDE 智能提示和跳转完美支持
- ✅ 测试和代码版本同步演进

**不推荐**：将测试放在独立的 `tests/` 目录（这会导致无法测试私有函数，违背 Go 惯例）

### 6.2 测试文件命名

```
user.go       → user_test.go
service.go    → service_test.go
handler.go    → handler_test.go
```

### 6.3 单元测试

```go
// internal/service/user_test.go

func TestUserService_GetUser(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    mockCache := &MockCache{}
    service := NewUserService(mockRepo, mockCache)
    
    ctx := context.Background()
    expectedUser := &User{ID: "123", Name: "Alice"}
    
    mockRepo.On("FindByID", ctx, "123").Return(expectedUser, nil)
    
    // Execute
    user, err := service.GetUser(ctx, "123")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    mockRepo.AssertExpectations(t)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{}
    mockCache := &MockCache{}
    service := NewUserService(mockRepo, mockCache)
    
    ctx := context.Background()
    mockRepo.On("FindByID", ctx, "999").Return(nil, errors.New("not found"))
    
    // Execute
    user, err := service.GetUser(ctx, "999")
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
}
```

### 6.4 表格驱动测试

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty email", "", true},
        {"no domain", "user@", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## 7. 依赖注入

### 7.1 构造函数注入

```go
// ✅ 推荐：依赖通过构造函数注入
type UserService struct {
    repo      UserRepository
    cache     Cache
    eventBus  EventBus
}

func NewUserService(
    repo UserRepository,
    cache Cache,
    eventBus EventBus,
) *UserService {
    return &UserService{
        repo:     repo,
        cache:    cache,
        eventBus: eventBus,
    }
}
```

### 7.2 接口依赖

```go
// ✅ 推荐：依赖接口而非具体实现
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Create(ctx context.Context, user *User) error
}

type Cache interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}) error
}
```

---

## 8. 配置管理

### 8.1 配置中心架构

apprun 使用 **Config Center Registry Pattern**，模块配置独立定义在各自包内，通过注册表统一管理。

**设计原则**：
- **Business Cohesion**: 模块配置定义在模块包内（如 `pkg/i18n/config.go`，`internal/jwt/config.go`）
- **Centralized Management**: Config Center 通过 Registry 统一管理所有模块配置
- **Three-Layer Model**: Business Structs (Source) → Config Center (Mapper) → Data Sources (YAML/DB/Env)

### 8.2 模块配置标准结构

**每个需要配置的模块必须包含 `config.go` 文件**，定义模块配置结构和工厂函数。

#### 8.2.1 Config.go 标准模板

```go
// pkg/yourmodule/config.go 或 internal/yourmodule/config.go
package yourmodule

import "time"

// Config defines the configuration for YourModule
// Tags: yaml (YAML key), default (default value), db (allow database storage), validate (validation rules)
type Config struct {
    // Add your configuration fields here
    Enabled bool   `yaml:"enabled" default:"true" db:"true" validate:""`
    Timeout string `yaml:"timeout" default:"30s" db:"false" validate:"required"`
    
    // Nested configuration
    Advanced AdvancedConfig `yaml:"advanced"`
}

// AdvancedConfig defines advanced configuration options
type AdvancedConfig struct {
    MaxRetries int `yaml:"max_retries" default:"3" db:"false" validate:"min=0,max=10"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
    return Config{
        Enabled: true,
        Timeout: "30s",
        Advanced: AdvancedConfig{
            MaxRetries: 3,
        },
    }
}

// ToRuntimeConfig converts Config to runtime configuration
// Use this when you need to parse or transform config values
func (c *Config) ToRuntimeConfig() (*RuntimeConfig, error) {
    timeout, err := time.ParseDuration(c.Timeout)
    if err != nil {
        return nil, err
    }
    
    return &RuntimeConfig{
        Enabled:    c.Enabled,
        Timeout:    timeout, // Parsed duration
        MaxRetries: c.Advanced.MaxRetries,
    }, nil
}

// RuntimeConfig is the internal configuration used by module at runtime
// Use parsed/transformed types (time.Duration, *url.URL, etc.)
type RuntimeConfig struct {
    Enabled    bool
    Timeout    time.Duration // Parsed from string
    MaxRetries int
}
```

#### 8.2.2 配置注册（main.go）

```go
// cmd/server/main.go

import (
    "apprun/modules/config"
    "apprun/pkg/i18n"
    "apprun/internal/jwt"
    "apprun/pkg/yourmodule"
)

func initializeConfigRegistry() *config.ConfigRegistry {
    registry := config.NewRegistry()
    
    // Register module configurations
    if err := registry.Register("logger", &logger.Config{}); err != nil {
        log.Fatalf("Failed to register logger config: %v", err)
    }
    
    if err := registry.Register("i18n", &i18n.Config{}); err != nil {
        log.Fatalf("Failed to register i18n config: %v", err)
    }
    
    if err := registry.Register("jwt", &jwt.Config{}); err != nil {
        log.Fatalf("Failed to register jwt config: %v", err)
    }
    
    // Register your module
    if err := registry.Register("yourmodule", &yourmodule.Config{}); err != nil {
        log.Fatalf("Failed to register yourmodule config: %v", err)
    }
    
    return registry
}
```

#### 8.2.3 配置文件（YAML）

```yaml
# config/default.yaml

# Your module configuration
yourmodule:
  enabled: true
  timeout: 30s
  advanced:
    max_retries: 3

# JWT configuration example
jwt:
  secret: ${JWT_SECRET}  # Environment variable
  expiry: 24h
  issuer: "apprun"
  whitelist_paths:
    - /api/v1/auth/register
    - /api/v1/auth/login
    - /health
```

### 8.3 工厂函数作为"配置连接器"

**工厂函数是模块与配置中心的连接桥梁**，提供优雅的初始化方式。

#### 8.3.1 工厂函数标准模板

```go
// pkg/yourmodule/yourmodule.go

// NewServiceFromConfig creates a service instance from configuration (recommended)
// This is the "Config Connector" - bridges module and Config Center
func NewServiceFromConfig(cfg *Config) (*Service, error) {
    // Convert to runtime config
    runtimeCfg, err := cfg.ToRuntimeConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to convert config: %w", err)
    }
    
    // Initialize service with runtime config
    return NewService(runtimeCfg), nil
}

// NewService creates a service instance from runtime configuration
// Use this for direct initialization (testing, advanced use cases)
func NewService(runtimeCfg *RuntimeConfig) *Service {
    return &Service{
        enabled:    runtimeCfg.Enabled,
        timeout:    runtimeCfg.Timeout,
        maxRetries: runtimeCfg.MaxRetries,
    }
}
```

#### 8.3.2 使用模式

**模式 1: 配置中心模式（推荐生产环境）**
```go
// Application startup - load from Config Center
registry := config.NewRegistry()
registry.Register("yourmodule", &yourmodule.Config{})

bootstrap := config.NewBootstrapWithRegistry("./config", registry)
configService, _ := bootstrap.CreateService(ctx, dbClient)

// Get module config from registry and initialize
yourmoduleCfg := &yourmodule.Config{
    // Config loaded from YAML/Env/DB by Config Center
}
service, err := yourmodule.NewServiceFromConfig(yourmoduleCfg)
```

**模式 2: 直接配置模式（测试/简单场景）**
```go
// Testing - direct configuration without Config Center
cfg := &yourmodule.Config{
    Enabled: true,
    Timeout: "5s",
}
service, err := yourmodule.NewServiceFromConfig(cfg)
```

**模式 3: Runtime Config 模式（高级场景）**
```go
// Advanced - bypass config conversion for performance
runtimeCfg := &yourmodule.RuntimeConfig{
    Enabled:    true,
    Timeout:    5 * time.Second,
    MaxRetries: 3,
}
service := yourmodule.NewService(runtimeCfg)
```

### 8.4 配置优先级（6-Layer System）

Config Center 使用 6 层优先级系统，**从低到高**：

1. **Struct Tag Defaults** - `default:"value"` tag
2. **default.yaml** - Base configuration file
3. **Specialized Files** - `database.yaml`, `server.yaml`, etc.
4. **conf_d/ Directory** - Additional config files
5. **Database** - Only for `db:"true"` fields
6. **Environment Variables** - Highest priority

```bash
# Example: Override JWT secret via environment
export JWT_SECRET="production-secret-key"
export JWT_EXPIRY="1h"
```

### 8.5 配置标签说明

| Tag | Description | Example | Required |
|-----|-------------|---------|----------|
| `yaml` | YAML key name | `yaml:"timeout"` | Yes |
| `default` | Default value | `default:"30s"` | Recommended |
| `db` | Allow database storage | `db:"true"` or `db:"false"` | Yes |
| `validate` | Validation rules | `validate:"required,min=1"` | Optional |
| `json` | JSON key (for API) | `json:"timeout"` | Optional |

**Validation Rules** (using go-playground/validator):
- `required` - Field must be non-zero
- `min=N` / `max=N` - Min/max value for numbers
- `len=N` - Exact length for strings/slices
- `oneof=A B C` - Value must be one of the options
- `url` / `email` - Format validation

### 8.6 实际案例对比

#### JWT Module (已实现)

```go
// internal/jwt/config.go
type Config struct {
    Secret         string   `yaml:"secret" default:"" db:"false" validate:"required,min=32"`
    Expiry         string   `yaml:"expiry" default:"24h" db:"false" validate:"required"`
    Issuer         string   `yaml:"issuer" default:"apprun" db:"false"`
    WhitelistPaths []string `yaml:"whitelist_paths" default:"[...]" db:"false"`
}

func (c *Config) ToRuntimeConfig() (*RuntimeConfig, error) {
    duration, err := time.ParseDuration(c.Expiry)
    if err != nil {
        return nil, err
    }
    
    whitelist := make(map[string]bool, len(c.WhitelistPaths))
    for _, path := range c.WhitelistPaths {
        whitelist[path] = true
    }
    
    return &RuntimeConfig{
        Secret:    c.Secret,
        Expiry:    duration,
        Issuer:    c.Issuer,
        Whitelist: whitelist,
    }, nil
}

// modules/auth/middleware/auth.go
func NewJWTMiddlewareFromConfig(cfg *jwt.Config) (*JWTMiddleware, error) {
    runtimeCfg, err := cfg.ToRuntimeConfig()
    if err != nil {
        return nil, err
    }
    return NewJWTMiddleware(runtimeCfg), nil
}
```

#### i18n Module (参考实现)

```go
// pkg/i18n/config.go
type Config struct {
    DefaultLanguage     string   `yaml:"default_language" default:"en-US" db:"false"`
    SupportedLanguages  []string `yaml:"supported_languages" db:"false"`
    TranslationsPath    string   `yaml:"translations_path" default:"./locales" db:"false"`
}

// pkg/i18n/i18n.go
func InitWithConfig(cfg *Config) error {
    // Load translations based on config
    return loadTranslations(cfg.TranslationsPath, cfg.SupportedLanguages)
}
```

### 8.7 最佳实践

**✅ DO**:
- 每个模块定义独立的 `config.go`
- 使用 `DefaultConfig()` 提供合理默认值
- 提供 `ToRuntimeConfig()` 处理类型转换
- 创建工厂函数 `NewXXXFromConfig(cfg *Config)` 作为配置连接器
- 在 `main.go` 中注册所有模块配置
- 使用环境变量覆盖敏感配置（如密钥）

**❌ DON'T**:
- ❌ 在 `internal/config/types.go` 中定义业务模块配置（违反 Registry Pattern）
- ❌ 在模块内部直接读取配置文件（破坏解耦）
- ❌ 硬编码配置值（如白名单路径、超时时间）
- ❌ 使用全局变量存储配置（不利于测试）
- ❌ 跳过配置验证（validate tag）

### 8.8 旧配置结构（已废弃）

```go
// ❌ OLD - 不符合 Registry Pattern
// internal/config/config.go

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    Kratos   KratosConfig   `mapstructure:"kratos"`
}

type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
    URL         string `mapstructure:"url"`
    MaxOpenConn int    `mapstructure:"max_open_conn"`
    MaxIdleConn int    `mapstructure:"max_idle_conn"`
}

// Load 加载配置
func Load(path string) (*Config, error) {
    viper.SetConfigFile(path)
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

**迁移说明**: 
- 旧的全局 Config 结构已被 Registry Pattern 替代
- 每个模块现在独立定义配置
- 参考 `pkg/i18n/config.go` 和 `internal/jwt/config.go` 作为标准实现

---

## 9. 日志规范

### 9.1 结构化日志

```go
import "github.com/sirupsen/logrus"

// ✅ 推荐：结构化日志
log.WithFields(logrus.Fields{
    "user_id":    userID,
    "project_id": projectID,
    "action":     "create_file",
}).Info("File created successfully")

// ✅ 推荐：错误日志包含上下文
log.WithError(err).WithFields(logrus.Fields{
    "user_id": userID,
}).Error("Failed to create user")

// ❌ 避免：非结构化日志
log.Println("User", userID, "created file in project", projectID)
```

### 9.2 日志级别

```go
// DEBUG - 调试信息
log.Debug("Cache hit for key: user:123")

// INFO - 常规信息
log.Info("User logged in successfully")

// WARN - 警告（不影响功能）
log.Warn("Cache miss, fetching from database")

// ERROR - 错误（影响功能）
log.WithError(err).Error("Failed to connect to database")

// FATAL - 致命错误（程序退出）
log.Fatal("Failed to start server")
```

---

## 10. 安全规范

### 10.1 输入验证

```go
// ✅ 推荐：验证所有用户输入
func CreateUser(input *CreateUserInput) error {
    if input.Name == "" {
        return errors.New("name is required")
    }
    
    if len(input.Name) > 100 {
        return errors.New("name too long")
    }
    
    if !isValidEmail(input.Email) {
        return errors.New("invalid email format")
    }
    
    return nil
}

// ✅ 推荐：使用白名单验证
func ValidateFileType(mimeType string) bool {
    allowedTypes := []string{
        "image/jpeg",
        "image/png",
        "application/pdf",
    }
    
    for _, allowed := range allowedTypes {
        if mimeType == allowed {
            return true
        }
    }
    return false
}
```

### 10.2 SQL 注入防护

```go
// ✅ 推荐：使用参数化查询（Ent 自动处理）
users, err := client.User.Query().
    Where(user.NameEQ(name)).  // 安全的参数化查询
    All(ctx)

// ❌ 避免：字符串拼接 SQL
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)
```

### 10.3 敏感信息处理

```go
// ✅ 推荐：不在日志中输出敏感信息
type User struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Password string `json:"-"`  // JSON 序列化时忽略
    APIKey   string `json:"-"`
}

// ✅ 推荐：使用环境变量存储密钥
dbPassword := os.Getenv("DB_PASSWORD")
jwtSecret := os.Getenv("JWT_SECRET")
```

### 10.4 数据加密

#### **传输加密**

```go
// ✅ 推荐：强制使用 HTTPS
func main() {
    // 生产环境只允许 HTTPS
    if config.Env == "production" {
        log.Fatal(http.ListenAndServeTLS(":443", "cert.pem", "key.pem", router))
    }
    
    // 开发环境可以使用 HTTP
    log.Fatal(http.ListenAndServe(":8080", router))
}

// ✅ 推荐：禁用不安全的 TLS 版本
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,  // 最低 TLS 1.2
    CipherSuites: []uint16{
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
    },
}
```

#### **存储加密**

```go
// ✅ 推荐：加密敏感字段
import "golang.org/x/crypto/bcrypt"

// 密码加密
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}

func VerifyPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}

// API Key 加密（使用 AES）
import "crypto/aes"
import "crypto/cipher"

func EncryptAPIKey(key string, secret []byte) ([]byte, error) {
    block, err := aes.NewCipher(secret)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    // 生产环境使用 crypto/rand
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    return gcm.Seal(nonce, nonce, []byte(key), nil), nil
}
```

### 10.5 密钥管理

#### **环境变量管理**

```bash
# .env.example（提交到版本控制）
DB_HOST=localhost
DB_PORT=5432
DB_NAME=apprun
DB_USER=postgres
DB_PASSWORD=         # 不填写实际值
JWT_SECRET=          # 不填写实际值
ENCRYPTION_KEY=      # 不填写实际值

# .env（不提交，添加到 .gitignore）
DB_PASSWORD=actual_password_here
JWT_SECRET=actual_secret_here
ENCRYPTION_KEY=actual_key_here
```

```go
// ✅ 推荐：密钥加载和验证
func LoadConfig() (*Config, error) {
    config := &Config{
        DBPassword:    os.Getenv("DB_PASSWORD"),
        JWTSecret:     os.Getenv("JWT_SECRET"),
        EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
    }
    
    // 验证必需的密钥
    if config.JWTSecret == "" {
        return nil, errors.New("JWT_SECRET is required")
    }
    
    if len(config.EncryptionKey) != 32 {
        return nil, errors.New("ENCRYPTION_KEY must be 32 bytes")
    }
    
    return config, nil
}
```

#### **密钥轮换**

```go
// ✅ 推荐：支持多密钥验证（密钥轮换）
type KeyManager struct {
    currentKey  string
    previousKey string  // 旧密钥，用于验证
}

func (km *KeyManager) Sign(data string) (string, error) {
    // 使用当前密钥签名
    return signWithKey(data, km.currentKey)
}

func (km *KeyManager) Verify(data, signature string) bool {
    // 先用当前密钥验证
    if verifyWithKey(data, signature, km.currentKey) {
        return true
    }
    
    // 如果失败，用旧密钥验证（支持轮换期）
    if km.previousKey != "" {
        return verifyWithKey(data, signature, km.previousKey)
    }
    
    return false
}
```

### 10.6 安全日志

```go
// ✅ 推荐：记录安全相关事件
func AuditLog(ctx context.Context, action string, details map[string]interface{}) {
    user := getUserFromContext(ctx)
    
    log.Info().
        Str("user_id", user.ID).
        Str("action", action).
        Fields(details).
        Str("ip", getIPFromContext(ctx)).
        Time("timestamp", time.Now()).
        Msg("security_audit")
}

// 使用示例
AuditLog(ctx, "user.login", map[string]interface{}{
    "method": "password",
    "success": true,
})

AuditLog(ctx, "file.delete", map[string]interface{}{
    "file_id": fileID,
    "project_id": projectID,
})

AuditLog(ctx, "permission.denied", map[string]interface{}{
    "resource": "project:123",
    "action": "delete",
})
```

```go
// ✅ 推荐：敏感操作失败次数限制
type RateLimiter struct {
    attempts map[string]int
    mu       sync.Mutex
}

func (rl *RateLimiter) CheckLoginAttempts(userID string) error {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    if rl.attempts[userID] >= 5 {
        return errors.New("too many failed login attempts, account locked")
    }
    
    return nil
}

func (rl *RateLimiter) RecordFailedLogin(userID string) {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    rl.attempts[userID]++
    
    // 记录安全日志
    log.Warn().
        Str("user_id", userID).
        Int("attempts", rl.attempts[userID]).
        Msg("failed_login_attempt")
}
```

---

## 11. 性能优化

### 11.1 避免不必要的分配

```go
// ✅ 推荐：预分配切片容量
users := make([]*User, 0, expectedSize)

// ✅ 推荐：使用 strings.Builder
var sb strings.Builder
sb.WriteString("Hello")
sb.WriteString(" ")
sb.WriteString("World")
result := sb.String()

// ❌ 避免：频繁字符串拼接
result := ""
for _, word := range words {
    result += word + " "  // 每次都会分配新内存
}
```

### 11.2 并发控制

```go
// ✅ 推荐：使用 Worker Pool 限制并发
func ProcessItems(items []Item) {
    const maxWorkers = 10
    semaphore := make(chan struct{}, maxWorkers)
    var wg sync.WaitGroup
    
    for _, item := range items {
        wg.Add(1)
        semaphore <- struct{}{}  // 获取令牌
        
        go func(i Item) {
            defer wg.Done()
            defer func() { <-semaphore }()  // 释放令牌
            processItem(i)
        }(item)
    }
    
    wg.Wait()
}
```

---

## 12. Ent ORM 规范

### 12.1 字段定义规范

**所有 Ent Schema 字段必须显式定义 JSON tag，使用 snake_case 格式**

```go
// ent/schema/user.go

func (User) Fields() []ent.Field {
    return []ent.Field{
        // ✅ 推荐：显式定义 JSON tag 和 StorageKey
        field.String("name").
            StorageKey("name").
            StructTag(`json:"name"`),
        
        field.String("email").
            StorageKey("email").
            StructTag(`json:"email"`),
        
        field.Time("created_at").
            StorageKey("created_at").
            StructTag(`json:"created_at"`).
            Default(time.Now),
        
        // 敏感字段：不在 JSON 中输出
        field.String("password_hash").
            StorageKey("password_hash").
            StructTag(`json:"-"`),
        
        // 可选字段：使用 omitempty
        field.String("phone").
            Optional().
            StorageKey("phone").
            StructTag(`json:"phone,omitempty"`),
    }
}
```

### 12.2 关系字段规范

```go
func (Project) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("owner", User.Type).
            Ref("projects").
            Unique().
            StructTag(`json:"owner"`),
        
        edge.To("members", User.Type).
            StructTag(`json:"members"`),
    }
}
```

### 12.3 Ent Schema 检查清单

- [ ] 所有字段有显式的 `json` tag
- [ ] JSON tag 使用 snake_case 格式
- [ ] StorageKey 与数据库列名一致
- [ ] 敏感字段使用 `json:"-"`
- [ ] 可选字段使用 `omitempty`
- [ ] 关系字段有适当的 JSON tag

---

## 13. 代码审查清单

### 13.1 通用检查

- [ ] 代码遵循 Go 命名规范
- [ ] 所有导出的函数和类型有注释
- [ ] 错误处理完整
- [ ] 没有 panic（除非必要）
- [ ] Context 正确传递
- [ ] 资源正确释放（defer）
- [ ] 并发安全（使用锁或 Channel）
- [ ] 单元测试覆盖
- [ ] 无 golangci-lint 警告
- [ ] Ent Schema 字段符合 JSON tag 规范

### 13.2 性能检查

- [ ] 避免不必要的内存分配
- [ ] 数据库查询优化（N+1 问题）
- [ ] 合理使用缓存
- [ ] 并发数量控制
- [ ] 大文件流式处理

### 13.3 安全检查

- [ ] 输入验证（长度、格式、白名单）
- [ ] SQL 注入防护（使用 ORM 参数化查询）
- [ ] XSS 防护（输出转义）
- [ ] 敏感信息不记录日志
- [ ] 密钥使用环境变量
- [ ] 传输加密（HTTPS/TLS 1.2+）
- [ ] 敏感数据存储加密（密码、API Key）
- [ ] 安全日志记录（登录、权限、敏感操作）
- [ ] 失败重试限制（防暴力破解）
- [ ] 权限验证（认证 + 授权）

---


## 附录

### A. 工具配置

#### golangci-lint 配置

```yaml
# .golangci.yml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - staticcheck
    - errcheck
    - gosec
    - ineffassign
    - unused

linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/websoft9/apprun
```

#### Ent Schema JSON Tag 检查脚本

```bash
#!/bin/bash
# scripts/check-ent-json-tags.sh

set -e

echo "🔍 检查 Ent Schema JSON tag 规范..."

schema_files=$(find ent/schema -name "*.go" 2>/dev/null || true)

if [ -z "$schema_files" ]; then
    echo "⚠️  未找到 Ent Schema 文件，跳过检查"
    exit 0
fi

errors=0

for file in $schema_files; do
    # 检查是否有未定义 JSON tag 的字段
    if grep -q "field\." "$file" && ! grep -q 'StructTag.*json:' "$file"; then
        echo "❌ $file: 发现字段缺少 JSON tag 定义"
        errors=$((errors + 1))
    fi
    
    # 检查 JSON tag 格式（应为 snake_case）
    if grep -P 'StructTag.*json:"[^"]*[A-Z][^"]*"' "$file" > /dev/null 2>&1; then
        echo "❌ $file: JSON tag 应使用 snake_case 格式"
        errors=$((errors + 1))
    fi
done

if [ $errors -eq 0 ]; then
    echo "✅ 所有 Ent Schema JSON tag 检查通过"
else
    echo "❌ 发现 $errors 个 JSON tag 规范问题"
    exit 1
fi
```

#### CI/CD GitHub Actions 配置

```yaml
# .github/workflows/ci.yml

name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --config=.golangci.yml
    
    - name: Check Ent Schema JSON tags
      run: |
        chmod +x scripts/check-ent-json-tags.sh
        ./scripts/check-ent-json-tags.sh

  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
```

#### EditorConfig

```ini
# .editorconfig
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{yaml,yml}]
indent_style = space
indent_size = 2
```

### B. Makefile 示例

```makefile
.PHONY: fmt lint ent-check test build

fmt:
	gofmt -s -w .
	goimports -w -local github.com/websoft9/apprun .

lint:
	golangci-lint run

ent-check:
	./scripts/check-ent-json-tags.sh

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

check: lint ent-check test

build:
	go build -o bin/server ./cmd/server
```

---

## 12. Docker 规范

### 12.1 Docker Compose 命令

**必须使用 Docker Compose V2 语法**:

```bash
# 正确 ✅
docker compose up -d
docker compose down

# 错误 ❌ (deprecated)
docker-compose up -d
docker-compose down
```

### 12.2 Docker Compose 文件格式

- **不使用 `version` 字段**（Docker Compose V2 已废弃）
- 文件直接以 `services:` 开头

```yaml
# 正确 ✅
services:
  app:
    image: myapp:latest

# 错误 ❌
version: '3.8'
services:
  app:
    image: myapp:latest
```

### 12.3 Dockerfile 最佳实践

- 使用多阶段构建减小镜像体积
- 使用非 root 用户运行应用
- 添加健康检查 (HEALTHCHECK)
- 静态编译 Go 二进制文件

### 12.4 Docker Compose 文件命名

- `docker-compose.yml` - 生产部署配置（默认）
- `docker-compose.dev.yml` - 本地开发依赖服务
- `docker-compose.local.yml` - 本地集成测试

---

## 13. 项目结构规范

### 13.1 Makefile 位置

**规则**: Makefile **必须**放在项目根目录，且**只能有一个**。

```
apprun/
├── Makefile          ✅ 唯一的构建入口
├── core/
│   ├── cmd/
│   └── pkg/
├── docker/
└── tests/
```

**禁止**: 在子目录创建独立的 Makefile（如 `core/Makefile`）

**原因**:
- 符合用户期望：开发者习惯在根目录执行构建命令
- 简化 CI/CD：GitHub Actions 默认在根目录执行
- 统一入口：所有构建、测试、部署命令集中管理
- 避免混淆：防止不同目录下的命令冲突

**使用方式**:
```bash
# 查看所有可用命令
make help

# 常用命令
make build              # 构建应用
make test-all           # 运行所有测试
make dev-up             # 启动开发环境
make clean              # 清理构建产物
```

### 13.2 目录组织原则

- `core/` - Go 应用核心代码
- `docker/` - Docker 相关配置（Dockerfile、compose 文件）
- `docs/` - 项目文档
- `tests/` - 测试脚本和数据
- `scripts/` - 辅助脚本
- `examples/` - 示例配置

---

## 14. 配置管理规范

### 14.1 配置优先级（6层）

从高到低：

1. **环境变量**（无前缀，最高优先级）
2. **数据库配置**（`configitems` 表，`db:"true"` 字段）
3. **用户配置**（`config/conf_d/*.yaml`，按字母顺序）
4. **领域配置**（`config/database.yaml` 等）
5. **默认配置**（`config/default.yaml`）
6. **结构体默认值**（`default` tag，最低优先级）

### 14.2 环境变量映射

**规则**: 无前缀，`section.field` → `SECTION_FIELD`

```bash
# 示例
database.host     → DATABASE_HOST
app.name          → APP_NAME
server.http_port  → SERVER_HTTP_PORT
```

### 14.3 结构体标签

```go
type Config struct {
    Database struct {
        Host string `yaml:"host" default:"localhost" db:"false"` // 不从DB加载
        Port int    `yaml:"port" default:"5432" db:"false"`
    } `yaml:"database"`
    
    App struct {
        Name string `yaml:"name" default:"apprun" db:"true"` // 可从DB加载
    } `yaml:"app"`
}
```

**标签说明**:
- `default` - 程序内置默认值
- `validate` - 校验规则（如 `required`, `min=1`）
- `db:"false"` - 禁止从数据库加载（防止循环依赖）
- `db:"true"` - 允许从数据库动态配置

### 14.4 数据库配置保护

**强制规则**: 数据库连接配置**必须**标记 `db:"false"`

```go
// ✅ 正确：防止循环依赖
type Config struct {
    Database struct {
        Host     string `db:"false"`
        Password string `db:"false"`
    }
}

// ❌ 错误：会导致循环依赖
type Config struct {
    Database struct {
        Host string `db:"true"` // 危险！
    }
}
```

### 14.5 模块配置结构体规范

**所有模块的配置结构体必须遵循统一的标签标准**，参考 `internal/config/types.go`。

#### **标准标签格式**

```go
type ModuleConfig struct {
    FieldName  Type  `validate:"..." default:"..." db:"true|false"`
}
```

**标签说明**：

| 标签 | 说明 | 必需 | 示例 |
|------|------|------|------|
| `validate` | 验证规则（go-playground/validator） | 推荐 | `validate:"required,min=1"` |
| `default` | 默认值（文档说明） | 推荐 | `default:"8080"` |
| `db` | 是否允许从配置中心加载 | **必需** | `db:"true"` 或 `db:"false"` |

#### **基础设施配置 vs 业务配置**

**1. 基础设施配置** - 标记 `db:"false"`

适用场景：
- 服务器启动配置（端口、SSL证书）
- 数据库连接配置
- 日志级别（影响全局）
- 启动时必需的参数

```go
// pkg/server/server.go
type Config struct {
    HTTPPort            string        `validate:"required,min=1,max=5" default:"8080" db:"false"`
    HTTPSPort           string        `validate:"required,min=1,max=5" default:"8443" db:"false"`
    SSLCertFile         string        `validate:"omitempty,file" default:"" db:"false"`
    SSLKeyFile          string        `validate:"omitempty,file" default:"" db:"false"`
    ShutdownTimeout     time.Duration `validate:"required,min=1s" default:"30s" db:"false"`
    EnableHTTPWithHTTPS bool          `default:"true" db:"false"`
}
```

**理由**：
- ✅ 服务器是最早启动的组件，不能依赖配置中心
- ✅ 避免循环依赖（配置中心需要服务器）
- ✅ 基础设施配置通过环境变量提供

**2. 业务配置** - 标记 `db:"true"`

适用场景：
- 业务功能开关
- API 密钥
- 业务规则参数
- 运行时可变的配置

```go
// internal/config/types.go
type Config struct {
    POC struct {
        Enabled  bool   `default:"true" db:"true"`
        Database string `validate:"required,url" default:"..." db:"true"`
        APIKey   string `validate:"required,min=10" db:"true"`
    } `yaml:"poc" validate:"required"`
}
```

**理由**：
- ✅ 支持运行时动态修改
- ✅ 可通过配置中心管理
- ✅ 不影响服务器启动

#### **验证规则示例**

```go
// 常用验证规则
type ExampleConfig struct {
    // 必填字段
    Name string `validate:"required" default:"example" db:"true"`
    
    // 字符串长度限制
    Password string `validate:"required,min=8,max=32" default:"" db:"false"`
    
    // 数字范围
    Port int `validate:"required,min=1,max=65535" default:"8080" db:"false"`
    
    // 枚举值
    LogLevel string `validate:"required,oneof=debug info warn error" default:"info" db:"true"`
    
    // URL 格式
    DatabaseURL string `validate:"required,url" default:"..." db:"false"`
    
    // 文件路径（可选）
    SSLCertFile string `validate:"omitempty,file" default:"" db:"false"`
    
    // 时间段
    Timeout time.Duration `validate:"required,min=1s" default:"30s" db:"false"`
}
```

#### **完整示例：Logger 模块**

```go
// pkg/logger/config.go
package logger

type Config struct {
    // 日志级别：debug, info, warn, error
    Level string `validate:"required,oneof=debug info warn error" default:"info" db:"true"`
    
    // 输出目标：stdout, stderr, file
    Output struct {
        Targets []string `validate:"required,min=1" default:"[\"stdout\"]" db:"true"`
        File    string   `validate:"omitempty,file" default:"" db:"true"`
    } `yaml:"output"`
    
    // 日志格式：json, text
    Format string `validate:"required,oneof=json text" default:"json" db:"true"`
    
    // 启用调用位置记录
    EnableCaller bool `default:"true" db:"true"`
}
```

#### **配置结构体检查清单**

- [ ] 每个字段都有 `yaml` 标签（**必需**，复杂驼峰命名必须显式指定）
- [ ] 每个字段都有 `validate` 标签（推荐）
- [ ] 每个字段都有 `default` 标签（文档作用）
- [ ] 每个字段都有 `db` 标签（**必需**）
- [ ] 基础设施配置标记 `db:"false"`
- [ ] 业务配置标记 `db:"true"`
- [ ] 验证规则符合业务逻辑
- [ ] 默认值合理且安全
- [ ] 提供 `DefaultConfig()` 函数用于实际默认值赋值

#### **yaml 标签命名规范**

**规则**: 使用 `snake_case` 命名，复杂驼峰字段必须显式指定

```go
// ✅ 正确：复杂驼峰命名显式指定 yaml 标签
type Config struct {
    HTTPPort            string `yaml:"http_port" ...`      // HTTP → http
    HTTPSPort           string `yaml:"https_port" ...`     // HTTPS → https
    SSLCertFile         string `yaml:"ssl_cert_file" ...`  // SSL → ssl
    EnableHTTPWithHTTPS bool   `yaml:"enable_http_with_https" ...`
}

// ❌ 错误：缺少 yaml 标签
type BadConfig struct {
    HTTPPort string `default:"8080" db:"false"`  // 缺少 yaml tag
}
```

**环境变量自动映射**（Viper AutomaticEnv）:
```bash
# yaml tag → 环境变量（自动大写下划线）
http_port              → HTTP_PORT
https_port             → HTTPS_PORT
ssl_cert_file          → SSL_CERT_FILE
enable_http_with_https → ENABLE_HTTP_WITH_HTTPS
```

#### **DefaultConfig() 函数规范**

**为什么需要**：
- ❌ `default` 标签不会自动赋值（仅作文档说明）
- ✅ `DefaultConfig()` 提供实际可用的默认值
- ✅ 支持复杂类型（如 `time.Duration`）
- ✅ 防御性编程（nil 检查）

**标准模式**:
```go
// 1. 定义配置结构体（包含 default 标签）
type Config struct {
    HTTPPort        string        `yaml:"http_port" validate:"required" default:"8080" db:"false"`
    ShutdownTimeout time.Duration `yaml:"shutdown_timeout" validate:"required" default:"30s" db:"false"`
}

// 2. 提供 DefaultConfig() 函数（实际赋值）
func DefaultConfig() *Config {
    return &Config{
        HTTPPort:        "8080",              // 与 default 标签保持一致
        ShutdownTimeout: 30 * time.Second,    // 类型安全
    }
}

// 3. 在 API 中使用
func Start(router http.Handler, cfg *Config) error {
    if cfg == nil {
        cfg = DefaultConfig()  // 防御性编程
    }
    // ...
}
```

**标签 vs 函数对比**:

| 特性 | default 标签 | DefaultConfig() 函数 |
|------|-------------|---------------------|
| 自动赋值 | ❌ 否 | ✅ 是 |
| 类型安全 | ❌ 仅字符串 | ✅ 完全类型安全 |
| 复杂类型 | ❌ 不支持 time.Duration | ✅ 支持 |
| 文档作用 | ✅ 是 | ✅ 是 |
| 可测试 | ❌ 不可测试 | ✅ 可单元测试 |
| 职责 | 元数据/文档 | 实际初始化 |

**最佳实践**:
- ✅ 同时保留两者（标签 + 函数）
- ✅ 保持默认值一致
- ✅ 所有配置结构体都应提供 `DefaultConfig()`

#### **反例：不规范的配置结构体**

```go
// ❌ 错误 1：缺少 yaml 标签
type BadConfig struct {
    HTTPPort string `default:"8080" db:"false"`  // 缺少 yaml tag
}

// ❌ 错误 2：缺少 DefaultConfig() 函数
type BadConfig2 struct {
    Port string `yaml:"port" default:"8080" db:"false"`
}
// 直接使用会得到零值！
cfg := &BadConfig2{}
fmt.Println(cfg.Port)  // 输出: "" (空字符串，不是 "8080")

// ❌ 错误 3：基础设施配置标记 db:"true"
type BadServerConfig struct {
    Port int `yaml:"port" db:"true"`  // 服务器端口不应该从配置中心加载
}

// ❌ 错误 4：缺少验证规则
type BadAuthConfig struct {
    Password string `yaml:"password" default:"123456" db:"false"`  // 缺少 min 验证
}
```

---

**文档维护**: Winston (Architect Agent) & Amelia (Dev Agent)  
**审核状态**: 待开发团队评审  
**下一步**: 测试规范文档 (testing-standards.md)
