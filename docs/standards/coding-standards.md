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

### 2.1 标准布局

```
apprun/
├── cmd/                    # 可执行程序入口
│   ├── server/            # HTTP 服务器
│   │   └── main.go
│   └── cli/               # CLI 工具
│       └── main.go
├── internal/              # 私有代码（不可被外部导入）
│   ├── config/           # 配置管理
│   ├── middleware/       # HTTP 中间件
│   ├── repository/       # 数据访问层
│   ├── service/          # 业务逻辑层
│   ├── handler/          # HTTP 处理器
│   ├── cache/            # 缓存实现
│   ├── events/           # 事件总线
│   └── errors/           # 错误定义
├── pkg/                   # 公共库（可被外部导入）
│   ├── kratos/           # Kratos 集成
│   ├── temporal/         # Temporal 集成
│   └── validator/        # 验证工具
├── ent/                   # Ent ORM Schema
│   └── schema/
├── api/                   # API 定义
│   ├── openapi/          # OpenAPI 规范
│   └── proto/            # Protobuf 定义（如果使用 gRPC）
├── config/                # 配置文件
│   ├── default.yaml
│   └── conf.d/
├── docs/                  # 文档
├── tests/                 # 测试
│   ├── integration/
│   └── e2e/
├── scripts/               # 脚本
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 2.2 internal 目录结构

```
internal/
├── handler/              # HTTP 层
│   ├── user.go          # 用户处理器
│   ├── project.go       # 项目处理器
│   └── middleware.go    # 中间件
├── service/              # 业务逻辑层
│   ├── user.go
│   └── project.go
├── repository/           # 数据访问层
│   ├── user.go
│   └── project.go
└── model/                # 业务模型（DTO）
    ├── user.go
    └── project.go
```

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

```go
// ✅ 推荐：显式错误处理
func GetUser(ctx context.Context, id string) (*User, error) {
    user, err := repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    return user, nil
}

// ✅ 推荐：自定义错误类型
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

// 使用
if err != nil {
    var notFoundErr *NotFoundError
    if errors.As(err, &notFoundErr) {
        return http.StatusNotFound, notFoundErr
    }
    return http.StatusInternalServerError, err
}
```

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

### 6.1 测试文件命名

```
user.go       → user_test.go
service.go    → service_test.go
handler.go    → handler_test.go
```

### 6.2 单元测试

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

### 6.3 表格驱动测试

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

### 8.1 配置结构

```go
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

## 12. 代码审查清单

### 12.1 通用检查

- [ ] 代码遵循 Go 命名规范
- [ ] 所有导出的函数和类型有注释
- [ ] 错误处理完整
- [ ] 没有 panic（除非必要）
- [ ] Context 正确传递
- [ ] 资源正确释放（defer）
- [ ] 并发安全（使用锁或 Channel）
- [ ] 单元测试覆盖
- [ ] 无 golangci-lint 警告

### 12.2 性能检查

- [ ] 避免不必要的内存分配
- [ ] 数据库查询优化（N+1 问题）
- [ ] 合理使用缓存
- [ ] 并发数量控制
- [ ] 大文件流式处理

### 12.3 安全检查

- [ ] 输入验证
- [ ] SQL 注入防护
- [ ] XSS 防护
- [ ] 敏感信息不记录日志
- [ ] 密钥使用环境变量

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
.PHONY: fmt lint test build

fmt:
	gofmt -s -w .
	goimports -w -local github.com/websoft9/apprun .

lint:
	golangci-lint run

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

build:
	go build -o bin/server ./cmd/server
```

---

**文档维护**: Winston (Architect Agent)  
**审核状态**: 待开发团队评审  
**下一步**: 测试规范文档 (testing-standards.md)
