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

### 2.1 模块化单体架构（推荐）

**按功能模块垂直切分，每个模块高内聚低耦合**

```
apprun/
├── cmd/                    # 可执行程序入口
│   └── server/main.go
│
├── modules/                # 业务模块（模块化单体）
│   ├── config/            # 配置管理模块
│   │   ├── handler.go     # HTTP API
│   │   ├── service.go     # 业务逻辑
│   │   ├── repository.go  # 数据访问
│   │   └── types.go       # 领域模型
│   │
│   ├── user/              # 用户模块
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── types.go
│   │
│   └── app/               # 应用管理模块
│       ├── handler.go
│       ├── service.go
│       ├── repository.go
│       └── types.go
│
├── internal/              # 内部基础设施（非业务模块）
│   ├── config/           # 全局配置加载器
│   ├── middleware/       # 中间件
│   ├── validator/        # 验证器
│   └── database/         # 数据库连接
│
├── pkg/                   # 可复用工具包
│   ├── logger/
│   └── errors/
│
├── ent/                   # Ent ORM
│   └── schema/
│
├── config/                # 配置文件
│   ├── default.yaml
│   └── conf_d/
│
├── docs/                  # 文档
├── tests/                 # 测试
├── Makefile
└── README.md
```

**优势**：
- ✅ 模块边界清晰，易于理解和维护
- ✅ 便于独立测试和部署
- ✅ 未来可无缝拆分为微服务

### 2.2 模块内部结构

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

---

**文档维护**: Winston (Architect Agent)  
**审核状态**: 待开发团队评审  
**下一步**: 测试规范文档 (testing-standards.md)
