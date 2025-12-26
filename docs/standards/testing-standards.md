# 测试规范
# apprun BaaS Platform

**创建日期**: 2025-12-25  
**维护者**: Winston (Architect Agent)  
**版本**: 1.0  
**状态**: Draft

---

## 1. 测试策略

### 1.1 测试金字塔

```
        ┌────────┐
        │  E2E   │  10%  - 端到端测试
        ├────────┤
        │ 集成测试│  30%  - 组件集成测试
        ├────────┤
        │ 单元测试│  60%  - 函数/方法测试
        └────────┘
```

### 1.2 测试目标

| 测试类型 | 覆盖率目标 | 执行频率 | 执行时间 |
|---------|-----------|---------|---------|
| **单元测试** | ≥ 70% | 每次提交 | < 1 分钟 |
| **集成测试** | ≥ 40% | 每次合并 | < 5 分钟 |
| **E2E 测试** | 核心流程 | 每日/发布前 | < 15 分钟 |

---

## 2. 单元测试

### 2.1 命名规范

```go
// 函数命名: Test<FunctionName>_<Scenario>
func TestUserService_GetUser_Success(t *testing.T) {}
func TestUserService_GetUser_NotFound(t *testing.T) {}
func TestUserService_GetUser_DatabaseError(t *testing.T) {}
```

### 2.2 测试结构

使用 **AAA 模式**（Arrange, Act, Assert）:

```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange - 准备测试数据和依赖
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    input := &CreateUserInput{
        Name:  "Alice",
        Email: "alice@example.com",
    }
    
    expectedUser := &User{
        ID:    "123",
        Name:  "Alice",
        Email: "alice@example.com",
    }
    
    mockRepo.On("Create", mock.Anything, input).Return(expectedUser, nil)
    
    // Act - 执行被测试的函数
    result, err := service.CreateUser(context.Background(), input)
    
    // Assert - 验证结果
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.ID, result.ID)
    assert.Equal(t, expectedUser.Name, result.Name)
    mockRepo.AssertExpectations(t)
}
```

### 2.3 表格驱动测试

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid email",
            email:   "user@example.com",
            wantErr: false,
        },
        {
            name:    "missing @",
            email:   "userexample.com",
            wantErr: true,
            errMsg:  "invalid email format",
        },
        {
            name:    "empty email",
            email:   "",
            wantErr: true,
            errMsg:  "email is required",
        },
        {
            name:    "no domain",
            email:   "user@",
            wantErr: true,
            errMsg:  "invalid email format",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2.4 Mock 和 Stub

#### 2.4.1 使用 testify/mock

```go
// MockUserRepository 实现 UserRepository 接口
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

// 使用示例
func TestUserService_GetUser(t *testing.T) {
    mockRepo := new(MockUserRepository)
    service := NewUserService(mockRepo)
    
    expectedUser := &User{ID: "123", Name: "Alice"}
    mockRepo.On("FindByID", mock.Anything, "123").Return(expectedUser, nil)
    
    user, err := service.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    mockRepo.AssertExpectations(t)
}
```

#### 2.4.2 使用接口 Stub

```go
// StubUserRepository 简单实现
type StubUserRepository struct {
    users map[string]*User
}

func NewStubUserRepository() *StubUserRepository {
    return &StubUserRepository{
        users: make(map[string]*User),
    }
}

func (s *StubUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    user, ok := s.users[id]
    if !ok {
        return nil, errors.New("user not found")
    }
    return user, nil
}

// 使用示例
func TestUserService_GetUser_WithStub(t *testing.T) {
    stubRepo := NewStubUserRepository()
    stubRepo.users["123"] = &User{ID: "123", Name: "Alice"}
    
    service := NewUserService(stubRepo)
    user, err := service.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

### 2.5 测试覆盖率

```bash
# 运行测试并生成覆盖率报告
go test -v -race -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html

# 查看特定包的覆盖率
go test -coverprofile=coverage.out ./internal/service
go tool cover -func=coverage.out
```

---

## 3. 集成测试

### 3.1 测试数据库

使用测试数据库或 Docker 容器:

```go
// tests/integration/setup.go

func SetupTestDB(t *testing.T) *ent.Client {
    // 使用 SQLite 内存数据库
    client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
    require.NoError(t, err)
    
    // 运行迁移
    ctx := context.Background()
    err = client.Schema.Create(ctx)
    require.NoError(t, err)
    
    // 清理函数
    t.Cleanup(func() {
        client.Close()
    })
    
    return client
}

// 使用示例
func TestUserRepository_Create(t *testing.T) {
    client := SetupTestDB(t)
    repo := NewUserRepository(client)
    
    user := &User{
        Name:  "Alice",
        Email: "alice@example.com",
    }
    
    err := repo.Create(context.Background(), user)
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### 3.2 使用 Docker 测试容器

```go
// tests/integration/testcontainers.go

import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgresContainer(t *testing.T) string {
    ctx := context.Background()
    
    req := testcontainers.ContainerRequest{
        Image:        "postgres:14-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }
    
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    require.NoError(t, err)
    
    t.Cleanup(func() {
        container.Terminate(ctx)
    })
    
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
    
    return fmt.Sprintf("postgresql://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
}
```

### 3.3 API 集成测试

```go
// tests/integration/api_test.go

func TestAPI_CreateProject(t *testing.T) {
    // Setup
    client := SetupTestDB(t)
    router := setupRouter(client)
    
    // 创建测试请求
    body := `{"name": "Test Project", "description": "Test"}`
    req := httptest.NewRequest("POST", "/api/v1/projects", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer test-token")
    
    // 执行请求
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // 验证响应
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
    assert.NotEmpty(t, response["data"].(map[string]interface{})["id"])
}
```

---

## 4. E2E 测试

### 4.1 测试场景

```go
// tests/e2e/scenarios/user_flow_test.go

func TestUserFlow_CreateProjectAndUploadFile(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }
    
    // 1. 用户注册
    user := registerUser(t, "alice@example.com", "password")
    
    // 2. 用户登录
    token := loginUser(t, "alice@example.com", "password")
    
    // 3. 创建项目
    project := createProject(t, token, "My Project")
    
    // 4. 上传文件
    file := uploadFile(t, token, project.ID, "test.txt", "Hello World")
    
    // 5. 查询文件
    files := listFiles(t, token, project.ID)
    assert.Len(t, files, 1)
    assert.Equal(t, file.ID, files[0].ID)
    
    // 6. 删除文件
    deleteFile(t, token, file.ID)
    
    // 7. 验证文件已删除
    files = listFiles(t, token, project.ID)
    assert.Len(t, files, 0)
}
```

### 4.2 使用 Docker Compose

```bash
# tests/e2e/docker-compose.yml
version: '3.8'

services:
  apprun:
    build: ../../
    environment:
      - DATABASE_URL=postgresql://test:test@postgres:5432/testdb
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - postgres
      - redis
    ports:
      - "8080:8080"
  
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: testdb
  
  redis:
    image: redis:7-alpine
```

```bash
# 运行 E2E 测试
cd tests/e2e
docker-compose up -d
go test -v ./scenarios/...
docker-compose down
```

---

## 5. 性能测试

### 5.1 基准测试

```go
// internal/cache/cache_bench_test.go

func BenchmarkCache_Get(b *testing.B) {
    cache := NewL1Cache()
    cache.Set("key", "value", 5*time.Minute)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Get("key")
    }
}

func BenchmarkCache_Set(b *testing.B) {
    cache := NewL1Cache()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
    }
}

// 并发基准测试
func BenchmarkCache_GetParallel(b *testing.B) {
    cache := NewL1Cache()
    cache.Set("key", "value", 5*time.Minute)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            cache.Get("key")
        }
    })
}
```

```bash
# 运行基准测试
go test -bench=. -benchmem ./internal/cache

# 输出示例
BenchmarkCache_Get-8          10000000    115 ns/op     0 B/op    0 allocs/op
BenchmarkCache_Set-8           5000000    243 ns/op    48 B/op    2 allocs/op
BenchmarkCache_GetParallel-8  50000000     34 ns/op     0 B/op    0 allocs/op
```

### 5.2 负载测试

使用 k6 进行负载测试:

```javascript
// tests/performance/load-test.js

import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 20 },  // 线性增加到 20 用户
        { duration: '1m', target: 20 },   // 保持 20 用户 1 分钟
        { duration: '30s', target: 0 },   // 线性减少到 0
    ],
    thresholds: {
        http_req_duration: ['p(95)<100'], // 95% 请求 < 100ms
        http_req_failed: ['rate<0.01'],   // 错误率 < 1%
    },
};

export default function() {
    // 获取项目列表
    let res = http.get('http://localhost:8080/api/v1/projects', {
        headers: { 'Authorization': 'Bearer test-token' },
    });
    
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 100ms': (r) => r.timings.duration < 100,
    });
    
    sleep(1);
}
```

```bash
# 运行负载测试
k6 run tests/performance/load-test.js
```

---

## 6. 测试辅助工具

### 6.1 测试夹具 (Fixtures)

```go
// tests/fixtures/user.go

func CreateTestUser(t *testing.T, client *ent.Client, name string) *ent.User {
    user, err := client.User.
        Create().
        SetName(name).
        SetEmail(fmt.Sprintf("%s@example.com", name)).
        Save(context.Background())
    
    require.NoError(t, err)
    return user
}

func CreateTestProject(t *testing.T, client *ent.Client, ownerID string) *ent.Project {
    project, err := client.Project.
        Create().
        SetName("Test Project").
        SetSlug("test-project").
        SetOwnerID(ownerID).
        Save(context.Background())
    
    require.NoError(t, err)
    return project
}

// 使用示例
func TestProjectService(t *testing.T) {
    client := SetupTestDB(t)
    user := CreateTestUser(t, client, "alice")
    project := CreateTestProject(t, client, user.ID)
    
    // ... 测试逻辑
}
```

### 6.2 测试工厂

```go
// tests/factory/user_factory.go

type UserFactory struct {
    client *ent.Client
    name   string
    email  string
}

func NewUserFactory(client *ent.Client) *UserFactory {
    return &UserFactory{
        client: client,
        name:   "Default User",
        email:  "default@example.com",
    }
}

func (f *UserFactory) WithName(name string) *UserFactory {
    f.name = name
    return f
}

func (f *UserFactory) WithEmail(email string) *UserFactory {
    f.email = email
    return f
}

func (f *UserFactory) Create(t *testing.T) *ent.User {
    user, err := f.client.User.
        Create().
        SetName(f.name).
        SetEmail(f.email).
        Save(context.Background())
    
    require.NoError(t, err)
    return user
}

// 使用示例
func TestUserService(t *testing.T) {
    client := SetupTestDB(t)
    
    user := NewUserFactory(client).
        WithName("Alice").
        WithEmail("alice@example.com").
        Create(t)
    
    // ... 测试逻辑
}
```

### 6.3 断言辅助函数

```go
// tests/common/assertions.go

func AssertUserEqual(t *testing.T, expected, actual *User) {
    t.Helper()
    assert.Equal(t, expected.ID, actual.ID)
    assert.Equal(t, expected.Name, actual.Name)
    assert.Equal(t, expected.Email, actual.Email)
}

func AssertProjectEqual(t *testing.T, expected, actual *Project) {
    t.Helper()
    assert.Equal(t, expected.ID, actual.ID)
    assert.Equal(t, expected.Name, actual.Name)
    assert.Equal(t, expected.Slug, actual.Slug)
}

func AssertErrorContains(t *testing.T, err error, substr string) {
    t.Helper()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), substr)
}
```

**CI/CD 流程说明**: 详细的 CI/CD 配置和测试执行流程请参考 [DevOps 流程规范](./devops-process.md#4-测试流程)。

---

## 7. 测试最佳实践

### 8.1 测试独立性

```go
// ✅ 推荐：每个测试独立运行
func TestUserService_CreateUser(t *testing.T) {
    // 创建独立的依赖
    mockRepo := new(MockUserRepository)
    service := NewUserService(mockRepo)
    
    // 独立的测试数据
    input := &CreateUserInput{Name: "Alice"}
    
    // ... 测试逻辑
}

// ❌ 避免：测试之间共享状态
var sharedService *UserService  // 不要这样做

func TestA(t *testing.T) {
    sharedService.DoSomething()  // 影响其他测试
}

func TestB(t *testing.T) {
    sharedService.DoSomething()  // 依赖 TestA 的执行
}
```

### 8.2 测试命名

```go
// ✅ 推荐：清晰描述测试场景
func TestUserService_GetUser_Success(t *testing.T) {}
func TestUserService_GetUser_NotFound(t *testing.T) {}
func TestUserService_GetUser_DatabaseError(t *testing.T) {}

// ❌ 避免：模糊的测试名
func TestGetUser(t *testing.T) {}
func TestGetUser2(t *testing.T) {}
func TestGetUserFail(t *testing.T) {}
```

### 8.3 测试可读性

```go
// ✅ 推荐：清晰的 AAA 结构
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    service := NewUserService(mockRepo)
    input := &CreateUserInput{Name: "Alice"}
    mockRepo.On("Create", mock.Anything, input).Return(&User{ID: "123"}, nil)
    
    // Act
    user, err := service.CreateUser(context.Background(), input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "123", user.ID)
}

// ❌ 避免：所有代码混在一起
func TestUserService_CreateUser(t *testing.T) {
    mockRepo := new(MockUserRepository)
    service := NewUserService(mockRepo)
    input := &CreateUserInput{Name: "Alice"}
    mockRepo.On("Create", mock.Anything, input).Return(&User{ID: "123"}, nil)
    user, err := service.CreateUser(context.Background(), input)
    assert.NoError(t, err)
    assert.Equal(t, "123", user.ID)
}
```

### 8.4 避免脆弱测试

```go
// ✅ 推荐：只验证关键行为
func TestUserService_CreateUser(t *testing.T) {
    // ... setup
    
    user, err := service.CreateUser(ctx, input)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)      // 只验证 ID 存在
    assert.Equal(t, "Alice", user.Name)
}

// ❌ 避免：过度验证实现细节
func TestUserService_CreateUser(t *testing.T) {
    // ... setup
    
    user, err := service.CreateUser(ctx, input)
    
    assert.NoError(t, err)
    assert.Equal(t, "123", user.ID)           // 硬编码 ID
    assert.Equal(t, "2025-12-25", user.CreatedAt.Format("2006-01-02"))  // 硬编码时间
    mockRepo.AssertNumberOfCalls(t, "Create", 1)  // 验证调用次数
    mockRepo.AssertCalled(t, "Create", ctx, input)  // 验证参数
}
```

---

## 8. 测试检查清单

### 8.1 单元测试检查

- [ ] 测试命名清晰（Test<Function>_<Scenario>）
- [ ] 使用 AAA 模式（Arrange, Act, Assert）
- [ ] 测试独立，不依赖执行顺序
- [ ] Mock 外部依赖
- [ ] 覆盖正常和异常情况
- [ ] 验证错误类型和消息
- [ ] 覆盖边界条件
- [ ] 单元测试覆盖率 ≥ 70%

### 9.2 集成测试检查

- [ ] 使用测试数据库
- [ ] 每个测试清理数据
- [ ] 测试真实的数据库操作
- [ ] 测试 HTTP API 端到端
- [ ] 验证事务边界
- [ ] 集成测试覆盖率 ≥ 40%

### 8.3 E2E 测试检查

- [ ] 测试完整用户流程
- [ ] 使用真实服务（Docker Compose）
- [ ] 覆盖关键业务场景
- [ ] 验证跨模块交互
- [ ] 测试执行时间 < 15 分钟

---

## 附录

### A. 测试工具推荐

| 工具 | 用途 | 链接 |
|-----|------|------|
| **testify** | 断言和 Mock | https://github.com/stretchr/testify |
| **gomock** | Mock 生成 | https://github.com/golang/mock |
| **testcontainers** | Docker 测试容器 | https://github.com/testcontainers/testcontainers-go |
| **httptest** | HTTP 测试 | Go 标准库 |
| **k6** | 负载测试 | https://k6.io/ |
| **Postman** | API 测试 | https://www.postman.com/ |

### B. Makefile 测试命令

```makefile
# Makefile

.PHONY: test test-unit test-integration test-e2e test-coverage

# 运行所有单元测试
test-unit:
	go test -v -race ./...

# 运行集成测试
test-integration:
	go test -v -tags=integration ./tests/integration/...

# 运行 E2E 测试
test-e2e:
	cd tests/e2e && docker-compose up -d
	go test -v ./tests/e2e/scenarios/...
	cd tests/e2e && docker-compose down

# 生成覆盖率报告
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 运行所有测试
test: test-unit test-integration test-e2e

# 快速测试（跳过慢速测试）
test-fast:
	go test -v -short ./...
```

---

**文档维护**: Winston (Architect Agent)  
**审核状态**: 待开发团队评审  
**完成**: 所有开发规范文档已创建完成
