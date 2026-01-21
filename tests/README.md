# Test Framework Documentation
# apprun BaaS Platform

**Created**: 2026-01-20  
**Status**: Active  
**Test Design**: [_bmad-output/test-design-apprun-system.md](../_bmad-output/test-design-apprun-system.md)

---

## 📋 Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Test Utilities](#test-utilities)
- [Best Practices](#best-practices)
- [CI/CD Integration](#cicd-integration)

---

## Overview

apprun使用**Go原生测试框架**，配合以下工具：

| 工具 | 用途 |
|------|------|
| **testing** | Go标准测试包 |
| **testify** | 断言库 |
| **httptest** | HTTP测试服务器 |
| **Testcontainers** | 集成测试容器（可选） |
| **k6** | 性能负载测试（独立） |

### 测试分层

```
tests/
├── unit/          # 单元测试 (40%) - 业务逻辑
├── integration/   # 黑盒集成测试 (50%) - API + 数据库
├── e2e/           # 端到端测试 (10%) - 完整流程
├── performance/   # 性能测试 (k6脚本)
├── fixtures/      # 测试数据工厂
└── testutils/     # 测试工具函数
```

> 模块的代码侧存放的是白盒集成测试
---

## Quick Start

### 1. 安装依赖

```bash
# 安装Go依赖
cd tests
go mod init github.com/Websoft9/apprun/tests
go get github.com/stretchr/testify
go get github.com/lib/pq
go get github.com/google/uuid
go get golang.org/x/crypto/bcrypt
```

### 2. 设置测试数据库

```bash
# 启动PostgreSQL
make deps-start

# 创建测试数据库
createdb -U postgres apprun_test

# 或使用环境变量
export TEST_DATABASE_URL="postgresql://postgres:postgres@localhost:5432/apprun_test?sslmode=disable"
```

### 3. 运行测试

```bash
# 运行所有单元测试
make test-unit

# 运行集成测试
make test-integration

# 运行E2E测试（需要启动服务器）
make test-e2e

# 运行所有测试
make test

# 生成覆盖率报告
make test-cover
```

---

## Test Structure

### 单元测试 (`tests/unit/`)

**特点**:
- 快速执行（无外部依赖）
- 测试纯函数和业务逻辑
- 不依赖数据库或HTTP

**示例**:
```go
// tests/unit/auth/password_test.go
func TestHashPassword(t *testing.T) {
    password := "SecurePass123!"
    
    hash, err := HashPassword(password)
    require.NoError(t, err)
    assert.NotEmpty(t, hash)
    
    // 验证bcrypt格式
    cost, _ := bcrypt.Cost([]byte(hash))
    assert.GreaterOrEqual(t, cost, 12)
}
```

### 集成测试 (`tests/integration/`)

**特点**:
- 涉及数据库、API、中间件
- 使用真实PostgreSQL（或Testcontainers）
- 测试组件交互

**示例**:
```go
// tests/integration/api/auth_test.go
func TestUserRegistration(t *testing.T) {
    // Setup测试数据库
    db := testutils.SetupTestDB(t)
    ctx := context.Background()
    
    // 使用工厂创建测试数据
    userFactory := fixtures.NewUserFactory(db)
    defer userFactory.Cleanup(ctx)
    
    // 测试API端点
    router := setupTestRouter(db)
    client := testutils.NewHTTPTestClient(router)
    
    resp := client.POST("/api/auth/register", map[string]string{
        "email": "test@example.com",
        "password": "SecurePass123!",
    })
    
    testutils.AssertStatusCode(t, resp, 200)
}
```

### E2E测试 (`tests/e2e/`)

**特点**:
- 测试完整用户流程
- 需要运行中的服务器
- 最接近真实使用场景

**示例**:
```go
// tests/e2e/scenarios/auth_flow_test.go
func TestCompleteAuthFlow(t *testing.T) {
    apiClient := NewAPIClient("http://localhost:8080")
    
    // Step 1: 注册
    registerResp := apiClient.POST("/api/auth/register", userData)
    // Step 2: 登录
    loginResp := apiClient.POST("/api/auth/login", credentials)
    // Step 3: 访问受保护资源
    apiClient.SetAuthToken(token)
    meResp := apiClient.GET("/api/users/me")
    // ...
}
```

---

## Running Tests

### Makefile命令

```bash
# 单元测试
make test-unit                    # 运行所有单元测试
make test-unit-cover              # 带覆盖率
make test-cover                   # 生成HTML覆盖率报告

# 集成测试
make test-integration             # 运行所有集成测试

# E2E测试
make test-e2e                     # 需要手动启动服务器
make test-e2e-auto                # 自动启动服务器

# 运行特定包的测试
make test-pkg PKG=unit/auth
make test-pkg PKG=integration/api

# 监听文件变化自动运行测试（需要entr）
make test-watch PKG=unit/auth

# 运行所有测试
make test
```

### Go命令

```bash
# 运行单个测试文件
cd tests
go test -v ./unit/auth/password_test.go

# 运行特定测试函数
go test -v -run TestHashPassword ./unit/auth/

# 带race detector
go test -race ./unit/...

# 短模式（跳过集成测试）
go test -short ./...

# 详细输出
go test -v ./...

# 并行运行（默认GOMAXPROCS）
go test -parallel 4 ./...

# 超时控制
go test -timeout 5m ./integration/...
```

---

## Writing Tests

### 测试模板

#### 单元测试模板

```go
package mypackage

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMyFunction(t *testing.T) {
	// 使用表格驱动测试
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"valid input", "test", "TEST", false},
		{"empty input", "", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MyFunction(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
```

#### 集成测试模板

```go
package api

import (
	"context"
	"testing"
	
	"github.com/Websoft9/apprun/tests/fixtures"
	"github.com/Websoft9/apprun/tests/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMyAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	
	// Setup
	db := testutils.SetupTestDB(t)
	ctx := context.Background()
	
	factory := fixtures.NewUserFactory(db)
	defer factory.Cleanup(ctx)
	
	t.Run("my test case", func(t *testing.T) {
		// Create test data
		user, err := factory.CreateUser(ctx)
		assert.NoError(t, err)
		
		// Test your API
		// ...
	})
}
```

---

## Test Utilities

### 数据库辅助工具

```go
// Setup test database with automatic cleanup
db := testutils.SetupTestDB(t)

// Manual truncate if needed
testutils.TruncateAllTables(t, db)
```

### HTTP辅助工具

```go
// Create HTTP test client
client := testutils.NewHTTPTestClient(router)

// Set JWT token
client.SetAuthToken("your-jwt-token")

// Make requests
resp := client.GET("/api/users/me")
resp := client.POST("/api/auth/login", loginData)

// Assert responses
testutils.AssertStatusCode(t, resp, 200)
testutils.AssertSuccessResponse(t, resp)
testutils.AssertErrorResponse(t, resp, "ERROR_CODE")
```

### 断言辅助工具

```go
// Time assertions
testutils.AssertTimeRecent(t, timestamp, 5) // Within 5 seconds

// Format assertions
testutils.AssertEmailFormat(t, "user@example.com")
testutils.AssertUUIDFormat(t, "550e8400-e29b-41d4-a716-446655440000")
testutils.AssertJWTFormat(t, "header.payload.signature")
testutils.AssertPasswordHashed(t, "$2a$12$...")

// Map assertions
testutils.AssertMapHasKeys(t, myMap, "key1", "key2")
```

### 测试数据工厂

```go
// User factory
userFactory := fixtures.NewUserFactory(db)
defer userFactory.Cleanup(ctx)

// Create user with defaults
user, err := userFactory.CreateUser(ctx)

// Create user with custom fields
user, err := userFactory.CreateUser(ctx,
	fixtures.WithEmail("custom@example.com"),
	fixtures.WithName("Custom User"),
	fixtures.WithPassword("CustomPass123!"),
)

// Create user with credentials (returns password)
user, password, err := userFactory.CreateUserWithCredentials(ctx, 
	"test@example.com", 
	"TestPass123!",
)

// Create multiple users
users, err := userFactory.CreateUsers(ctx, 5)

// Project factory
projectFactory := fixtures.NewProjectFactory(db)
defer projectFactory.Cleanup(ctx)

project, err := projectFactory.CreateProject(ctx,
	fixtures.WithProjectName("Test Project"),
	fixtures.WithProjectOwner(user.ID),
)

// Assign user to project
err = projectFactory.AssignUserToProject(ctx, 
	userID, 
	projectID, 
	"member",
)
```

---

## Best Practices

### ✅ DO

1. **使用表格驱动测试**
   ```go
   tests := []struct {
       name     string
       input    int
       expected int
   }{
       {"positive", 5, 25},
       {"negative", -5, 25},
       {"zero", 0, 0},
   }
   ```

2. **使用Subtests**
   ```go
   t.Run("valid input", func(t *testing.T) {
       // ...
   })
   ```

3. **清理资源**
   ```go
   t.Cleanup(func() {
       factory.Cleanup(ctx)
   })
   ```

4. **使用testutils辅助函数**
   ```go
   testutils.AssertStatusCode(t, resp, 200)
   ```

5. **测试隔离**
   - 每个测试独立
   - 使用工厂创建数据
   - 自动清理

6. **使用require vs assert**
   ```go
   require.NoError(t, err)  // 失败立即停止
   assert.Equal(t, a, b)    // 失败继续执行
   ```

### ❌ DON'T

1. **❌ 测试依赖顺序**
   ```go
   // 错误：依赖t1先运行
   func t1() { createUser() }
   func t2() { loginUser() }  // 依赖t1
   ```

2. **❌ 硬编码测试数据**
   ```go
   // 错误
   email := "test@example.com"  // 多个测试可能冲突
   
   // 正确
   email := fmt.Sprintf("test-%d@example.com", time.Now().Unix())
   ```

3. **❌ 跳过错误检查**
   ```go
   // 错误
   user, _ := factory.CreateUser(ctx)
   
   // 正确
   user, err := factory.CreateUser(ctx)
   require.NoError(t, err)
   ```

4. **❌ 测试实现细节**
   ```go
   // 错误：测试内部实现
   assert.Contains(t, sql, "SELECT * FROM users")
   
   // 正确：测试行为
   users, err := repo.GetAllUsers(ctx)
   assert.NoError(t, err)
   assert.Len(t, users, 5)
   ```

---

## CI/CD Integration

### GitHub Actions示例

```yaml
# .github/workflows/test.yml
name: Test Suite

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run unit tests
        run: make test-unit
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./tests/coverage-unit.out

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      
      - name: Run integration tests
        run: make test-integration
        env:
          TEST_DATABASE_URL: postgresql://postgres:postgres@localhost:5432/apprun_test?sslmode=disable
```

---

## Troubleshooting

### 常见问题

#### 1. 数据库连接失败

```bash
# 检查PostgreSQL是否运行
psql -U postgres -c "SELECT 1"

# 检查测试数据库是否存在
psql -U postgres -l | grep apprun_test

# 创建测试数据库
createdb -U postgres apprun_test
```

#### 2. 测试超时

```bash
# 增加超时时间
go test -timeout 10m ./integration/...

# 或在Makefile中设置
cd tests && go test -timeout 5m ./integration/...
```

#### 3. 并发测试冲突

```go
// 使用唯一标识符
email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
```

#### 4. 端口冲突

```bash
# E2E测试端口被占用
lsof -i :8080
kill -9 <PID>
```

---

## Next Steps

1. **实现Auth API handlers** - 当前测试使用TODO标记
2. **配置CI/CD** - 添加GitHub Actions workflow
3. **添加更多测试场景** - 参考[test-design-apprun-system.md](../_bmad-output/test-design-apprun-system.md)
4. **性能测试** - 配置k6脚本
5. **安全测试** - 集成gosec和govulncheck

---

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Test Design Document](../_bmad-output/test-design-apprun-system.md)
- [BMad Test Architect Workflows](.bmad/bmm/workflows/testarch/)

---

**维护者**: Websoft9  
**最后更新**: 2026-01-20
