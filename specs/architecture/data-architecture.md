# 数据架构文档
# apprun BaaS Platform

**创建日期**: 2025-12-25  
**架构师**: Winston (Architect Agent)  
**版本**: 1.0  
**状态**: Draft

---

## 1. 数据架构概览

### 1.1 数据分层

```
┌─────────────────────────────────────────────────┐
│            Application Layer (Go)               │
│  ┌─────────────────────────────────────────┐   │
│  │  Handler Layer (HTTP/Validation)        │   │
│  └──────────────────┬──────────────────────┘   │
│  ┌──────────────────▼──────────────────────┐   │
│  │  Service Layer (Business Logic)         │   │
│  └──────────────────┬──────────────────────┘   │
│  ┌──────────────────▼──────────────────────┐   │
│  │  Repository Layer (Data Access)         │   │
│  └──────────────────┬──────────────────────┘   │
└───────────────────┬─┴──────────────────────────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
  ┌─────────┐ ┌─────────┐ ┌─────────┐
  │  Ent    │ │  Cache  │ │  Redis  │
  │  ORM    │ │ (L1/L2) │ │ Streams │
  └────┬────┘ └────┬────┘ └────┬────┘
       │           │           │
       ▼           ▼           ▼
  PostgreSQL    Redis      Redis
  (主数据)      (缓存)     (事件)
```

---

## 2. 核心数据模型

### 2.1 认证与用户 (Auth & Users)

#### 2.1.1 Kratos 表（只读）

```go
// Ory Kratos 管理的表（apprun 只读访问）

// identities - 用户身份
type Identity struct {
    ID              string    `json:"id"`              // UUID
    SchemaID        string    `json:"schema_id"`       // identity.default
    Traits          JSON      `json:"traits"`          // 用户属性（email, name）
    State           string    `json:"state"`           // active, inactive
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// sessions - 会话
type Session struct {
    ID              string    `json:"id"`
    IdentityID      string    `json:"identity_id"`     // FK -> identities.id
    Active          bool      `json:"active"`
    ExpiresAt       time.Time `json:"expires_at"`
    AuthenticatedAt time.Time `json:"authenticated_at"`
    IssuedAt        time.Time `json:"issued_at"`
}
```

#### 2.1.2 apprun 用户表

```go
// ent/schema/user.go

// User - 用户扩展信息
type User struct {
    ent.Schema
}

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("identity_id").Unique(),      // FK -> Kratos identities.id
        field.String("username").Optional(),        // 用户名
        field.String("display_name").Optional(),    // 显示名称
        field.String("avatar_url").Optional(),      // 头像 URL
        field.JSON("metadata", map[string]interface{}{}).Optional(), // 扩展元数据
        field.Enum("status").Values("active", "suspended", "deleted").Default("active"),
        field.Time("last_login_at").Optional(),
        field.Time("created_at").Immutable().Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (User) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("projects", UserProject.Type),     // 用户项目关系
    }
}

func (User) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("identity_id").Unique(),
        index.Fields("status"),
    }
}
```

---

### 2.2 项目与权限 (Projects & RBAC)

```go
// ent/schema/project.go

// Project - 项目
type Project struct {
    ent.Schema
}

func (Project) Fields() []ent.Field {
    return []ent.Field{
        field.String("name"),
        field.String("slug").Unique(),              // URL-safe 标识符
        field.Text("description").Optional(),
        field.String("owner_id"),                   // FK -> users.id
        field.JSON("settings", map[string]interface{}{}).Optional(),
        field.Enum("status").Values("active", "archived", "deleted").Default("active"),
        field.Time("created_at").Immutable().Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Project) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("owner", User.Type).Ref("owned_projects").Unique().Required().Field("owner_id"),
        edge.To("members", UserProject.Type),       // 项目成员
        edge.To("models", Model.Type),              // 数据模型
        edge.To("files", File.Type),                // 文件
        edge.To("functions", Function.Type),        // 函数
    }
}

func (Project) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("slug").Unique(),
        index.Fields("owner_id"),
        index.Fields("status"),
    }
}
```

```go
// ent/schema/user_project.go

// UserProject - 用户项目关系（RBAC）
type UserProject struct {
    ent.Schema
}

func (UserProject) Fields() []ent.Field {
    return []ent.Field{
        field.String("user_id"),                    // FK -> users.id
        field.String("project_id"),                 // FK -> projects.id
        field.Enum("role").Values("owner", "admin", "editor", "viewer").Default("viewer"),
        field.JSON("permissions", []string{}).Optional(), // 自定义权限列表
        field.Time("joined_at").Immutable().Default(time.Now),
    }
}

func (UserProject) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("user", User.Type).Unique().Required().Field("user_id"),
        edge.To("project", Project.Type).Unique().Required().Field("project_id"),
    }
}

func (UserProject) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("user_id", "project_id").Unique(), // 唯一约束
        index.Fields("project_id"),
    }
}
```

---

### 2.3 数据建模 (Data Modeling)

```go
// ent/schema/model.go

// Model - 用户定义的数据模型
type Model struct {
    ent.Schema
}

func (Model) Fields() []ent.Field {
    return []ent.Field{
        field.String("project_id"),                 // FK -> projects.id
        field.String("name"),                       // 模型名称（如 Product）
        field.String("table_name"),                 // 数据库表名
        field.Text("description").Optional(),
        field.JSON("schema", map[string]interface{}{}), // DSL Schema 定义
        field.Bool("is_system").Default(false),     // 是否系统模型
        field.Int("version").Default(1),            // Schema 版本号
        field.Time("created_at").Immutable().Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Model) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).Ref("models").Unique().Required().Field("project_id"),
        edge.To("fields", ModelField.Type),         // 模型字段
    }
}

func (Model) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("project_id", "name").Unique(),
        index.Fields("table_name").Unique(),
    }
}
```

```go
// ent/schema/model_field.go

// ModelField - 模型字段定义
type ModelField struct {
    ent.Schema
}

func (ModelField) Fields() []ent.Field {
    return []ent.Field{
        field.String("model_id"),                   // FK -> models.id
        field.String("name"),                       // 字段名
        field.Enum("type").Values("string", "int", "float", "bool", "datetime", "json", "text"),
        field.Bool("required").Default(false),
        field.Bool("unique").Default(false),
        field.String("default_value").Optional(),
        field.JSON("constraints", map[string]interface{}{}).Optional(), // 验证规则
        field.Int("order").Default(0),              // 显示顺序
        field.Time("created_at").Immutable().Default(time.Now),
    }
}

func (ModelField) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("model", Model.Type).Ref("fields").Unique().Required().Field("model_id"),
    }
}

func (ModelField) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("model_id", "name").Unique(),
    }
}
```

---

### 2.4 文件存储 (File Storage)

```go
// ent/schema/file.go

// File - 文件元数据
type File struct {
    ent.Schema
}

func (File) Fields() []ent.Field {
    return []ent.Field{
        field.String("project_id"),                 // FK -> projects.id
        field.String("name"),                       // 文件名
        field.String("path"),                       // 虚拟路径 /project-1/docs/file.pdf
        field.String("storage_path"),               // 实际存储路径
        field.Enum("storage_type").Values("local", "s3").Default("local"),
        field.String("mime_type").Optional(),
        field.Int64("size").Default(0),             // 字节数
        field.String("hash").Optional(),            // SHA256 哈希
        field.String("uploader_id"),                // FK -> users.id
        field.JSON("metadata", map[string]interface{}{}).Optional(),
        field.Time("created_at").Immutable().Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (File) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).Ref("files").Unique().Required().Field("project_id"),
        edge.From("uploader", User.Type).Ref("uploaded_files").Unique().Required().Field("uploader_id"),
        edge.From("folder", Folder.Type).Ref("files").Unique().Optional().Field("folder_id"),
    }
}

func (File) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("project_id", "path").Unique(),
        index.Fields("storage_path"),
        index.Fields("uploader_id"),
    }
}
```

```go
// ent/schema/folder.go

// Folder - 文件夹（虚拟）
type Folder struct {
    ent.Schema
}

func (Folder) Fields() []ent.Field {
    return []ent.Field{
        field.String("project_id"),                 // FK -> projects.id
        field.String("name"),
        field.String("path"),                       // /project-1/docs/
        field.String("parent_id").Optional(),       // FK -> folders.id (自引用)
        field.JSON("metadata", map[string]interface{}{}).Optional(),
        field.Time("created_at").Immutable().Default(time.Now),
    }
}

func (Folder) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).Ref("folders").Unique().Required().Field("project_id"),
        edge.To("children", Folder.Type).From("parent").Unique().Field("parent_id"),
        edge.To("files", File.Type),
    }
}

func (Folder) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("project_id", "path").Unique(),
    }
}
```

---

### 2.5 函数服务 (Functions)

```go
// ent/schema/function.go

// Function - 函数定义
type Function struct {
    ent.Schema
}

func (Function) Fields() []ent.Field {
    return []ent.Field{
        field.String("project_id"),                 // FK -> projects.id
        field.String("name"),
        field.Text("description").Optional(),
        field.Text("code"),                         // Go 代码
        field.Enum("runtime").Values("go1.24").Default("go1.24"),
        field.Enum("trigger").Values("http", "event", "cron").Default("http"),
        field.JSON("trigger_config", map[string]interface{}{}).Optional(), // 触发器配置
        field.JSON("env_vars", map[string]string{}).Optional(), // 环境变量
        field.Int("timeout").Default(30),           // 超时（秒）
        field.Int("memory_limit").Default(128),     // 内存限制（MB）
        field.Bool("enabled").Default(true),
        field.Time("created_at").Immutable().Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Function) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).Ref("functions").Unique().Required().Field("project_id"),
        edge.To("executions", FunctionExecution.Type),
    }
}

func (Function) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("project_id", "name").Unique(),
        index.Fields("enabled"),
    }
}
```

```go
// ent/schema/function_execution.go

// FunctionExecution - 函数执行记录
type FunctionExecution struct {
    ent.Schema
}

func (FunctionExecution) Fields() []ent.Field {
    return []ent.Field{
        field.String("function_id"),                // FK -> functions.id
        field.JSON("input", map[string]interface{}{}).Optional(),
        field.JSON("output", map[string]interface{}{}).Optional(),
        field.Enum("status").Values("running", "success", "failed", "timeout"),
        field.Text("error").Optional(),
        field.Int("duration_ms").Default(0),        // 执行时长（毫秒）
        field.Time("started_at").Immutable().Default(time.Now),
        field.Time("finished_at").Optional(),
    }
}

func (FunctionExecution) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("function", Function.Type).Ref("executions").Unique().Required().Field("function_id"),
    }
}

func (FunctionExecution) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("function_id"),
        index.Fields("started_at"),
    }
}
```

---

### 2.6 配置管理 (Configuration)

```go
// ent/schema/config_item.go

// ConfigItem - 配置项（已存在）
type ConfigItem struct {
    ent.Schema
}

func (ConfigItem) Fields() []ent.Field {
    return []ent.Field{
        field.String("key").Unique(),               // 配置键
        field.Text("value"),                        // 配置值
        field.Enum("type").Values("string", "int", "bool", "json").Default("string"),
        field.Text("description").Optional(),
        field.Bool("is_secret").Default(false),     // 是否敏感信息
        field.String("namespace").Default("default"), // 命名空间
        field.Time("created_at").Immutable().Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (ConfigItem) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("namespace", "key").Unique(),
    }
}
```

---

## 3. 缓存架构

### 3.1 多层缓存策略

```go
// internal/cache/multilevel.go

// L1: 进程内缓存
type L1Cache struct {
    data sync.Map                   // 并发安全的 Map
    ttl  time.Duration
}

func (c *L1Cache) Get(key string) (interface{}, bool) {
    if item, ok := c.data.Load(key); ok {
        cached := item.(*CacheItem)
        if time.Now().Before(cached.ExpiresAt) {
            return cached.Value, true
        }
        c.data.Delete(key)  // 过期删除
    }
    return nil, false
}

func (c *L1Cache) Set(key string, value interface{}, ttl time.Duration) {
    c.data.Store(key, &CacheItem{
        Value:     value,
        ExpiresAt: time.Now().Add(ttl),
    })
}

// L2: Redis 缓存
type L2Cache struct {
    client *redis.Client
}

func (c *L2Cache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}

func (c *L2Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}

// 多层缓存管理器
type MultiLevelCache struct {
    l1 *L1Cache
    l2 *L2Cache
    db *ent.Client
}

// Get - 读取数据（L1 → L2 → DB）
func (c *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, error) {
    // 1. 查询 L1
    if val, ok := c.l1.Get(key); ok {
        return val, nil
    }
    
    // 2. 查询 L2
    if c.l2 != nil {
        if val, err := c.l2.Get(ctx, key); err == nil {
            c.l1.Set(key, val, 5*time.Minute)  // 回写 L1
            return val, nil
        }
    }
    
    // 3. 查询数据库
    val, err := c.queryDB(ctx, key)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    if c.l2 != nil {
        c.l2.Set(ctx, key, val, 30*time.Minute)
    }
    c.l1.Set(key, val, 5*time.Minute)
    
    return val, nil
}

// Set - 写入数据（Write-through 模式）
func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}) error {
    // 1. 更新数据库
    if err := c.updateDB(ctx, key, value); err != nil {
        return err
    }
    
    // 2. 更新 L2
    if c.l2 != nil {
        c.l2.Set(ctx, key, value, 30*time.Minute)
    }
    
    // 3. 更新 L1
    c.l1.Set(key, value, 5*time.Minute)
    
    return nil
}

// Delete - 删除缓存
func (c *MultiLevelCache) Delete(ctx context.Context, key string) error {
    c.l1.data.Delete(key)
    if c.l2 != nil {
        return c.l2.client.Del(ctx, key).Err()
    }
    return nil
}
```

### 3.2 缓存键命名规范

```go
// 缓存键格式
const (
    // 用户缓存: user:{identity_id}
    CacheKeyUser = "user:%s"
    
    // 项目缓存: project:{project_id}
    CacheKeyProject = "project:%s"
    
    // 配置缓存: config:{namespace}:{key}
    CacheKeyConfig = "config:%s:%s"
    
    // 会话缓存: session:{session_id}
    CacheKeySession = "session:%s"
    
    // API 响应缓存: api:{method}:{path}:{query_hash}
    CacheKeyAPIResponse = "api:%s:%s:%s"
)

// TTL 配置
const (
    TTLUserCache     = 10 * time.Minute   // 用户信息
    TTLProjectCache  = 30 * time.Minute   // 项目信息
    TTLConfigCache   = 1 * time.Hour      // 配置项
    TTLSessionCache  = 24 * time.Hour     // 会话
    TTLAPICache      = 5 * time.Minute    // API 响应
)
```

### 3.3 缓存失效策略

```go
// 缓存失效场景

// 1. 用户更新 → 删除用户缓存
func (s *UserService) UpdateUser(ctx context.Context, id string, input *UpdateUserInput) error {
    if err := s.repo.Update(ctx, id, input); err != nil {
        return err
    }
    
    // 删除缓存
    cacheKey := fmt.Sprintf(CacheKeyUser, id)
    return s.cache.Delete(ctx, cacheKey)
}

// 2. 项目更新 → 删除项目相关缓存
func (s *ProjectService) UpdateProject(ctx context.Context, id string, input *UpdateProjectInput) error {
    if err := s.repo.Update(ctx, id, input); err != nil {
        return err
    }
    
    // 删除项目缓存
    cacheKey := fmt.Sprintf(CacheKeyProject, id)
    if err := s.cache.Delete(ctx, cacheKey); err != nil {
        return err
    }
    
    // 删除关联的成员缓存
    members, _ := s.repo.GetMembers(ctx, id)
    for _, member := range members {
        userKey := fmt.Sprintf(CacheKeyUser, member.UserID)
        s.cache.Delete(ctx, userKey)
    }
    
    return nil
}

// 3. 配置更新 → Viper Watch 自动失效
```

---

## 4. 事件架构 (Redis Streams)

### 4.1 事件定义

```go
// internal/events/types.go

// Event - 事件结构
type Event struct {
    ID        string                 `json:"id"`         // 事件 ID
    Type      string                 `json:"type"`       // 事件类型
    Source    string                 `json:"source"`     // 事件源
    Data      map[string]interface{} `json:"data"`       // 事件数据
    Timestamp time.Time              `json:"timestamp"`
}

// 事件类型常量
const (
    EventUserCreated      = "user.created"
    EventUserUpdated      = "user.updated"
    EventUserDeleted      = "user.deleted"
    
    EventProjectCreated   = "project.created"
    EventProjectUpdated   = "project.updated"
    
    EventFileUploaded     = "file.uploaded"
    EventFileDeleted      = "file.deleted"
    
    EventFunctionExecuted = "function.executed"
    EventModelCreated     = "model.created"
)
```

### 4.2 事件发布/订阅

```go
// internal/events/bus.go

// EventBus - Redis Streams 事件总线
type EventBus struct {
    client *redis.Client
}

// Publish - 发布事件
func (eb *EventBus) Publish(ctx context.Context, eventType string, data map[string]interface{}) error {
    event := Event{
        ID:        uuid.New().String(),
        Type:      eventType,
        Source:    "apprun",
        Data:      data,
        Timestamp: time.Now(),
    }
    
    // 序列化为 JSON
    payload, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    // 发布到 Redis Stream
    streamKey := fmt.Sprintf("events:%s", eventType)
    return eb.client.XAdd(ctx, &redis.XAddArgs{
        Stream: streamKey,
        Values: map[string]interface{}{
            "payload": payload,
        },
    }).Err()
}

// Subscribe - 订阅事件
func (eb *EventBus) Subscribe(ctx context.Context, eventType string, handler func(Event)) error {
    streamKey := fmt.Sprintf("events:%s", eventType)
    consumerGroup := "apprun-consumers"
    
    // 创建消费者组
    eb.client.XGroupCreate(ctx, streamKey, consumerGroup, "0")
    
    // 持续读取
    for {
        streams, err := eb.client.XReadGroup(ctx, &redis.XReadGroupArgs{
            Group:    consumerGroup,
            Consumer: "apprun-1",
            Streams:  []string{streamKey, ">"},
            Count:    10,
            Block:    time.Second,
        }).Result()
        
        if err != nil {
            continue
        }
        
        for _, stream := range streams {
            for _, message := range stream.Messages {
                var event Event
                if err := json.Unmarshal([]byte(message.Values["payload"].(string)), &event); err != nil {
                    continue
                }
                
                // 调用处理器
                handler(event)
                
                // 确认消息
                eb.client.XAck(ctx, streamKey, consumerGroup, message.ID)
            }
        }
    }
}
```

### 4.3 事件使用示例

```go
// 发布事件
eventBus.Publish(ctx, EventUserCreated, map[string]interface{}{
    "user_id": user.ID,
    "email":   user.Email,
})

// 订阅事件
eventBus.Subscribe(ctx, EventUserCreated, func(event Event) {
    userID := event.Data["user_id"].(string)
    log.Printf("New user created: %s", userID)
    
    // 触发工作流
    workflowService.TriggerOnUserCreated(ctx, userID)
    
    // 发送欢迎邮件
    emailService.SendWelcome(ctx, userID)
})
```

---

## 5. 数据迁移

### 5.1 Atlas 迁移流程

```bash
# 1. 修改 Ent Schema
vim ent/schema/user.go

# 2. 生成 Ent 代码
go generate ./ent

# 3. 生成迁移文件（Dry-run）
atlas migrate diff migration_name \
  --dir "file://ent/migrate/migrations" \
  --to "ent://ent/schema" \
  --dev-url "docker://postgres/14/test"

# 4. 查看迁移 SQL
cat ent/migrate/migrations/20251225120000_migration_name.sql

# 5. 应用迁移
atlas migrate apply \
  --dir "file://ent/migrate/migrations" \
  --url "postgres://apprun:password@localhost:5432/apprun?sslmode=disable"

# 6. 验证 Schema
atlas schema inspect \
  --url "postgres://apprun:password@localhost:5432/apprun?sslmode=disable"
```

### 5.2 迁移最佳实践

```go
// ent/migrate/migrations/20251225120000_add_user_status.sql

-- 示例迁移 SQL
BEGIN;

-- 添加字段
ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active';

-- 创建索引
CREATE INDEX idx_users_status ON users(status);

-- 数据迁移
UPDATE users SET status = 'active' WHERE status IS NULL;

-- 添加约束
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

COMMIT;
```

---

## 6. 数据备份策略

### 6.1 PostgreSQL 备份

```bash
# 全量备份
pg_dump -U apprun apprun > backup-full-$(date +%Y%m%d).sql

# 仅 Schema
pg_dump -U apprun --schema-only apprun > backup-schema.sql

# 仅数据
pg_dump -U apprun --data-only apprun > backup-data.sql

# 压缩备份
pg_dump -U apprun apprun | gzip > backup-$(date +%Y%m%d).sql.gz

# WAL 归档（持续备份）
archive_mode = on
archive_command = 'cp %p /backups/wal/%f'
```

### 6.2 数据保留策略

| 数据类型 | 保留期 | 备份频率 | 存储位置 |
|----------|--------|----------|----------|
| 核心业务数据 | 永久 | 每日全量 | S3/OSS |
| 函数执行日志 | 30 天 | 不备份 | 数据库 |
| 文件存储 | 永久 | 每日增量 | S3/OSS |
| Redis 缓存 | 7 天 | AOF 持久化 | 本地 |
| 监控数据 | 30 天 | 不备份 | Prometheus |

---

## 7. 数据安全

### 7.1 敏感数据加密

```go
// internal/crypto/encryption.go

// 加密敏感字段
type EncryptionService struct {
    key []byte  // AES-256 密钥
}

func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(s.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }
    
    block, err := aes.NewCipher(s.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonceSize := gcm.NonceSize()
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
    return string(plaintext), nil
}

// 使用示例
func (s *UserService) SaveAPIKey(ctx context.Context, userID string, apiKey string) error {
    // 加密 API Key
    encrypted, err := s.crypto.Encrypt(apiKey)
    if err != nil {
        return err
    }
    
    // 存储加密后的值
    return s.repo.UpdateAPIKey(ctx, userID, encrypted)
}
```

### 7.2 行级安全 (可选)

```sql
-- PostgreSQL Row Level Security (RLS)
-- 注: MVP 不实现，仅供参考

-- 启用 RLS
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

-- 创建策略：用户只能访问自己的项目
CREATE POLICY project_access ON projects
    FOR SELECT
    USING (
        owner_id = current_setting('app.current_user_id')::text
        OR EXISTS (
            SELECT 1 FROM user_projects
            WHERE project_id = projects.id
            AND user_id = current_setting('app.current_user_id')::text
        )
    );
```

---

## 8. 性能优化

### 8.1 索引策略

```sql
-- 用户表索引
CREATE INDEX idx_users_identity_id ON users(identity_id);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_last_login ON users(last_login_at DESC);

-- 项目表索引
CREATE INDEX idx_projects_slug ON projects(slug);
CREATE INDEX idx_projects_owner ON projects(owner_id);
CREATE INDEX idx_projects_status ON projects(status);

-- 用户项目关系索引
CREATE UNIQUE INDEX idx_user_projects_unique ON user_projects(user_id, project_id);
CREATE INDEX idx_user_projects_project ON user_projects(project_id);

-- 文件表索引
CREATE INDEX idx_files_project_path ON files(project_id, path);
CREATE INDEX idx_files_uploader ON files(uploader_id);
CREATE INDEX idx_files_created ON files(created_at DESC);

-- 函数执行记录索引（分区表）
CREATE INDEX idx_function_executions_function ON function_executions(function_id);
CREATE INDEX idx_function_executions_started ON function_executions(started_at DESC);
```

### 8.2 查询优化

```go
// 使用 Ent 预加载（避免 N+1 查询）

// ❌ 错误：N+1 查询
projects, _ := client.Project.Query().All(ctx)
for _, project := range projects {
    owner, _ := project.QueryOwner().Only(ctx)  // N 次查询
    fmt.Println(owner.Name)
}

// ✅ 正确：使用 WithOwner 预加载
projects, _ := client.Project.Query().
    WithOwner().  // 预加载 owner
    All(ctx)
for _, project := range projects {
    fmt.Println(project.Edges.Owner.Name)  // 无额外查询
}

// 分页查询
projects, _ := client.Project.Query().
    Where(project.StatusEQ("active")).
    Order(ent.Desc(project.FieldCreatedAt)).
    Limit(10).
    Offset(0).
    All(ctx)
```

### 8.3 连接池配置

```go
// internal/database/pool.go

import (
    "database/sql"
    _ "github.com/lib/pq"
)

func NewDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    // 连接池配置
    db.SetMaxOpenConns(100)        // 最大连接数
    db.SetMaxIdleConns(10)         // 最大空闲连接
    db.SetConnMaxLifetime(1 * time.Hour)  // 连接最大生命周期
    db.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接超时
    
    return db, nil
}
```

---

## 附录

### A. 数据库 Schema 总览

```
核心表:
├── identities (Kratos - 只读)
├── sessions (Kratos - 只读)
├── users
├── projects
├── user_projects
├── models
├── model_fields
├── files
├── folders
├── functions
├── function_executions
├── config_items
└── servers (已存在)

估算数据量 (100 个项目):
- users: ~1,000 行
- projects: ~100 行
- user_projects: ~500 行
- models: ~500 行
- files: ~10,000 行
- function_executions: ~100,000 行/月（需定期清理）
```

### B. Ent 代码生成命令

```bash
# 生成 Ent 代码
go generate ./ent

# 生成带 GraphQL 的代码
go run -mod=mod entgo.io/contrib/entgql/cmd/entgql generate ./ent/schema

# 生成 OpenAPI Spec
go run -mod=mod github.com/ogen-go/ogen/cmd/ogen --target ./api/openapi ./ent/openapi.json
```

### C. 数据字典

详细的数据字典请参阅：`docs/database-schema.md`（待创建）

---

**文档维护**: Winston (Architect Agent)  
**审核状态**: 待技术团队评审  
**下一步**: 
- 创建开发规范文档 (docs/standards/api-design.md, coding-standards.md, testing-standards.md)
- 或开始 MVP 原型开发
