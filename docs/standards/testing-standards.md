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

## 9. 测试评审流程

### 9.1 评审触发条件

| 触发条件 | 责任人 | 时机 |
|---------|--------|------|
| **Story 完成** | TEA Agent / QA Lead | PR 合并前 |
| **覆盖率 < 70%** | Developer / TEA | CI 自动触发 |
| **生产发布前** | Release Manager | 发布前 24h |
| **关键功能变更** | Architect / TEA | 设计评审后 |

### 9.2 评审清单

基于**第 8 章测试检查清单**进行评审，重点关注：

- **覆盖率**: 单元测试 ≥ 70%，集成测试 ≥ 40%
- **测试结构**: AAA 模式、命名规范、表格驱动测试
- **边缘案例**: 错误路径、边界值、空值处理、并发安全
- **Mock/Stub**: 依赖隔离、接口设计、返回值验证
- **集成测试**: API 端点、数据库操作、完整流程
- **可维护性**: 测试意图清晰、失败信息有用、避免重复

详细检查项参见第 8 章。

### 9.3 质量评分标准

#### 9.3.1 评分规则

```
起始分数: 100
扣分: P0(-10) | P1(-5) | P2(-2) | P3(-1)
加分: 优秀实践 (最高 +30, 每项 +5)
最终: max(0, min(100, 起始分 - 扣分 + 加分))
```

#### 9.3.2 等级标准

| 分数 | 等级 | 评价 | 行动 |
|-----|------|------|------|
| **90-100** | A+ | 优秀 | ✅ 直接批准 |
| **80-89** | A | 良好 | ✅ 批准 + 建议改进 |
| **70-79** | B | 可接受 | ⚠️ 有条件批准 + 技术债工单 |
| **60-69** | C | 需改进 | ❌ 要求修改关键问题 |
| **< 60** | F | 不可接受 | 🚫 阻止合并 |

#### 9.3.3 违规严重程度

**P0 (Critical)**: 覆盖率 < 50%、缺少集成测试、nil 依赖、数据竞争  
**P1 (High)**: 覆盖率 50-70%、缺少边缘案例、无并发测试、DRY 违规  
**P2 (Medium)**: 缺少 fixture、错误消息验证不足、文档缺失  
**P3 (Low)**: 命名不清晰、辅助函数可提取、测试顺序优化

#### 9.3.2 等级标准

| 分数范围 | 等级 | 评价 | 行动建议 |
|---------|------|------|---------|
| **90-100** | A+ | 优秀 (Excellent) | 直接批准合并 ✅ |
| **80-89** | A | 良好 (Good) | 批准合并，建议改进 ✅ |
| **70-79** | B | 可接受 (Acceptable) | 有条件批准，记录技术债 ⚠️ |
| **60-69** | C | 需改进 (Needs Improvement) | 要求修改关键问题 ❌ |
| **< 60** | F | 不可接受 (Critical Issues) | 阻止合并，重新设计 🚫 |

#### 9.3.3 违规严重程度定义

**Critical (P0) - 阻止生产部署**:
- 测试覆盖率 < 50%
- 缺少集成测试
- 测试使用 nil 依赖并期望失败
- 存在数据竞争 (race condition)
- 硬编码延迟 (sleep, waitFor 无理由)
- 共享状态导致测试不隔离

**High (P1) - 影响可维护性**:
- 测试覆盖率 50-70%
- 缺少边缘案例测试（5+ 缺失）
- 缺少并发安全测试
- DRY 违规严重（重复代码 > 30%）
- 缺少 Mock/Stub 导致测试脆弱

**Medium (P2) - 技术债**:
- Fixture 模式缺失
- 错误消息验证不足
- 验证规则测试不完整
- 测试文档缺失

**Low (P3) - 优化建议**:
- 测试命名不够清晰
- 辅助函数可以提取
- 测试顺序可以优化

### 9.4 评审文档管理

**存储路径**: `docs/sprint-artifacts/sprint-{n}/story-{n}-TEST-REVIEW.md`

**命名规范**: `story-10-TEST-REVIEW.md` (大写 TEST-REVIEW 便于识别)

**文档关联**: 在 sprint README 中建立三位一体链接
```markdown
| 10 | Configuration Center | ✅ | [Story](story-10.md) · [Implementation](story-10-IMPLEMENTATION-SUMMARY.md) · [Test Review](story-10-TEST-REVIEW.md) |
```

**必需章节**:
1. Executive Summary (质量分数、等级、关键优缺点、推荐决策)
2. Quality Criteria Assessment (评审清单状态表)
3. Quality Score Breakdown (违规统计、扣分计算)
4. Critical Issues (P0/P1 详情：位置、描述、修复建议、代码示例)
5. Recommendations (P2/P3 改进建议)
6. Best Practices Found (优秀模式示例)
7. Test File Analysis (元数据、结构、覆盖率)
8. Next Steps (即时行动 + 后续改进)
9. Decision (批准/有条件批准/要求修改/阻止，附理由)

**示例**: [story-10-TEST-REVIEW.md](../sprint-artifacts/sprint-0/story-10-TEST-REVIEW.md)

**保留策略**:
- 永久保留: 关键功能、生产部署前、严重缺陷修复、首次实现
- 归档 1 年: 常规 story、重构、技术债修复
- 保留 3 个月: 小修改、重复评审

### 9.5 批准标准

| 质量分数 | 覆盖率 | P0 违规 | 决策 | 条件 |
|---------|--------|---------|------|------|
| ≥ 80 | ≥ 70% | 0 | ✅ **直接批准** | 无 |
| 70-79 | ≥ 60% | 0 | ⚠️ **有条件批准** | 创建技术债工单，下 sprint 修复 P1 |
| 60-69 | ≥ 50% | 1-2 | ❌ **要求修改** | 修复 P0 后重审 |
| < 60 | < 50% | 3+ | 🚫 **阻止合并** | 全面重构测试 |

**决策流程**:
```
覆盖率 ≥ 70% + 所有测试通过 + 无 P0 违规 + 分数 ≥ 80 → ✅ 直接批准
覆盖率 ≥ 60% + 无 P0 违规 + 分数 70-79 → ⚠️ 有条件批准 (技术债工单)
存在 P0 违规 或 覆盖率 50-60% → ❌ 要求修改
覆盖率 < 50% 或 多个 P0 违规 → 🚫 阻止合并
```

### 9.6 评审执行流程

**准备阶段** (5-10 min):
- 阅读 story 定义和验收标准
- 运行: `make test && make test-coverage`
- 打开覆盖率报告: `coverage.html`

**执行阶段** (30-60 min):
1. **快速扫描** (10 min): 覆盖率总览、未测试文件、命名结构
2. **深度分析** (20-30 min): 检查评审清单、识别违规、记录问题
3. **评分总结** (10-20 min): 统计违规、计算分数、编写报告

**输出**: 评审文档 + 质量分数/等级 + 批准决策 + 行动计划

### 9.7 持续改进

**跟踪指标**:
- 平均测试覆盖率 (目标 ≥ 70%, 每周)
- 平均质量分数 (目标 ≥ 80, 每 sprint)
- P0 违规率 (目标 < 5%, 每 sprint)
- 评审周期时间 (目标 < 2 天, 每月)

**每 Sprint 回顾**: 评审是否及时？建议是否有价值？是否捕获生产缺陷？流程需要优化？

**知识积累**: 建立测试知识库（常见违规模式、优秀案例、反模式、框架最佳实践）

---

## 10. 测试评审示例

### 10.1 完整评审案例: Story 10 配置中心

参考实际评审文档: [story-10-TEST-REVIEW.md](../sprint-artifacts/sprint-0/story-10-TEST-REVIEW.md)

**关键学习点**:
- 质量分数 72/100 (B - Acceptable)
- 覆盖率 42.7% (低于 70% 标准)
- 3 个 P0 违规，4 个 P1 违规
- 决策: Approve with Comments (有条件批准)
- 条件: 在下一个 sprint 修复 P0 违规

**最佳实践**:
- ✅ 优秀的分层测试结构
- ✅ 清晰的 AAA 模式
- ✅ 正确使用 `t.TempDir()` 隔离

**需改进**:
- ❌ 缺少集成测试
- ❌ main_test.go 使用 nil 数据库
- ❌ 边缘案例覆盖不足

---

**文档维护**: Murat (TEA Agent) + Winston (Architect Agent)  
**最后更新**: 2025-12-29  
**审核状态**: 待开发团队评审  
**版本**: 1.1 - 添加测试评审流程章节
