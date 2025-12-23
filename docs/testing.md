# 测试指南

## 概述

apprun 项目采用分层测试策略，确保代码质量和功能正确性。

## 测试目录结构

```
tests/
├── common.sh                    # 共享测试工具函数
├── integration/                 # 集成测试
│   └── config/                  # 配置模块集成测试
│       ├── test-api.sh         # API 功能测试
│       └── test-priority.sh    # 优先级测试
├── e2e/                        # 端到端测试
│   └── scenarios/              # 测试场景
└── scripts/                    # 测试辅助脚本
    ├── setup-test-db.sh       # 测试环境准备
    └── cleanup.sh             # 清理脚本
```

## 测试类型

### 1. 单元测试 (Unit Tests)
- **位置**: `core/internal/*/`
- **工具**: Go testing framework
- **覆盖**: 核心逻辑、工具函数
- **命令**: `make test-unit`

### 2. 集成测试 (Integration Tests)
- **位置**: `tests/integration/{module}/`
- **工具**: Shell 脚本 + Docker
- **覆盖**: API 接口、数据库交互
- **命令**: `make test-integration`

### 3. 端到端测试 (E2E Tests)
- **位置**: `tests/e2e/`
- **工具**: 待实现
- **覆盖**: 完整用户流程

## 运行测试

### 快速测试
```bash
# 运行所有测试
make test-all

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 运行配置模块测试（推荐）
make test-config
```

### 开发环境测试
```bash
# 启动开发环境
make dev

# 在另一个终端运行测试
make test-config
```

### 手动测试
```bash
# 启动测试环境
tests/scripts/setup-test-db.sh

# 运行单个测试
tests/integration/config/test-api.sh

# 清理环境
tests/scripts/cleanup.sh
```

## 编写测试

### 单元测试示例

#### 基本单元测试

```go
// core/internal/config/config_test.go
package config

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "apprun/ent"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
    // Setup test database
    client := setupTestDB(t)
    defer client.Close()

    // Test InitConfig
    err := InitConfig(client)
    assert.NoError(t, err)
    assert.NotNil(t, entClient)
    assert.NotNil(t, validate)
    assert.NotNil(t, viperInstance)
}

func TestLoadConfig(t *testing.T) {
    // Setup test database
    client := setupTestDB(t)
    defer client.Close()

    // Initialize config
    err := InitConfig(client)
    require.NoError(t, err)

    // Create temporary config directory
    tempDir := t.TempDir()
    configDir := filepath.Join(tempDir, "config")
    err = os.MkdirAll(configDir, 0755)
    require.NoError(t, err)

    // Create default.yaml
    defaultConfig := `app:
  name: "test-app"
poc:
  enabled: true
`
    err = os.WriteFile(filepath.Join(configDir, "default.yaml"), []byte(defaultConfig), 0644)
    require.NoError(t, err)

    // Change to temp directory for config loading
    oldWd, _ := os.Getwd()
    os.Chdir(tempDir)
    defer os.Chdir(oldWd)

    // Test LoadConfig
    cfg, err := LoadConfig()
    assert.NoError(t, err)
    assert.NotNil(t, cfg)
    assert.Equal(t, "test-app", cfg.App.Name)
    assert.True(t, cfg.POC.Enabled)
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *ent.Client {
    client, err := ent.Open("sqlite3", "file:memdb?mode=memory&cache=shared&_fk=1")
    require.NoError(t, err)

    // Run schema migration
    err = client.Schema.Create(context.Background())
    require.NoError(t, err)

    return client
}
```

#### HTTP处理器测试

```go
// core/cmd/server/main_test.go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestHandleGetConfig(t *testing.T) {
    // Setup test database and config
    // ... setup code ...

    // Create request
    req := httptest.NewRequest(http.MethodGet, "/config", nil)
    w := httptest.NewRecorder()

    // Call handler
    handleGetConfig(w, req)

    // Check response
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

    var items []config.ConfigItem
    err := json.NewDecoder(w.Body).Decode(&items)
    assert.NoError(t, err)
    assert.NotEmpty(t, items)
}

func TestHandlePutConfig(t *testing.T) {
    // Setup test database and config
    // ... setup code ...

    t.Run("successful update", func(t *testing.T) {
        updates := map[string]interface{}{
            "poc.enabled": false,
        }
        body, _ := json.Marshal(updates)

        req := httptest.NewRequest(http.MethodPut, "/config", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()

        handlePutConfig(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
    })
}
```

#### 结构体验证测试

```go
// core/internal/config/types_test.go
package config

import (
    "testing"

    "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
    validate := validator.New()

    tests := []struct {
        name        string
        config      Config
        expectError bool
    }{
        {
            name: "valid config",
            config: Config{
                App: struct {
                    Name    string `validate:"required,min=1"`
                    Version string `validate:"required"`
                }{
                    Name:    "test-app",
                    Version: "1.0.0",
                },
                Database: struct {
                    Driver   string `validate:"required,oneof=postgres mysql"`
                    Host     string `validate:"required"`
                    Port     int    `validate:"required,min=1,max=65535"`
                    User     string `validate:"required"`
                    Password string `validate:"required,min=8"`
                    DBName   string `validate:"required"`
                }{
                    Driver:   "postgres",
                    Host:     "localhost",
                    Port:     5432,
                    User:     "testuser",
                    Password: "testpassword",
                    DBName:   "testdb",
                },
                POC: struct {
                    Enabled  bool   `validate:""`
                    Database string `validate:"required,url"`
                    APIKey   string `validate:"required,min=10"`
                }{
                    Enabled:  true,
                    Database: "postgres://user:pass@localhost:5432/test",
                    APIKey:   "test-api-key-12345",
                },
            },
            expectError: false,
        },
        {
            name: "invalid config - missing app name",
            config: Config{
                App: struct {
                    Name    string `validate:"required,min=1"`
                    Version string `validate:"required"`
                }{
                    Name:    "",
                    Version: "1.0.0",
                },
            },
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validate.Struct(tt.config)
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 测试工具

### 共享函数 (`tests/common.sh`)

- `assert_equals expected actual message` - 断言相等
- `http_get path` - GET 请求
- `http_put path data` - PUT 请求
- `print_summary` - 打印测试结果

### 环境变量

- `BASE_URL`: API 基础 URL (默认: http://localhost:8080)
- `no_proxy`: 代理排除列表

## CI/CD 集成

### GitHub Actions 示例

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Unit Tests
        run: make test-unit

      - name: Start Test Environment
        run: make docker-up

      - name: Integration Tests
        run: make test-integration

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
```

## 测试覆盖率

```bash
# 生成覆盖率报告
cd core && go test -coverprofile=coverage.out ./...
cd core && go tool cover -html=coverage.out -o coverage.html

# 查看覆盖率
cd core && go tool cover -func=coverage.out
```

## 最佳实践

### 1. 测试命名
- 文件: `{feature}_test.go` 或 `test-{feature}.sh`
- 函数: `Test{FunctionName}` 或 `test_{function_name}`

### 2. 测试组织
- **Given-When-Then**: 清晰的测试结构
- **独立性**: 每个测试独立运行
- **可重复**: 结果一致性

### 3. 错误处理
- 使用断言函数记录失败
- 提供清晰的错误信息
- 自动清理测试数据

### 4. 性能考虑
- 单元测试快速执行
- 集成测试使用 Docker 隔离
- 并行执行提高效率

## 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 检查端口占用
   lsof -i :8080
   # 停止冲突服务
   docker compose down
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker compose logs postgres
   # 重启数据库
   docker compose restart postgres
   ```

3. **测试超时**
   ```bash
   # 增加等待时间
   sleep 10
   ```

## 扩展测试

### 添加新模块测试
```bash
# 1. 创建目录
mkdir -p tests/integration/auth

# 2. 创建测试脚本
cat > tests/integration/auth/test-login.sh << 'EOF'
#!/bin/bash
source "$(dirname "$0")/../../common.sh"

# 登录测试逻辑
EOF

# 3. 更新 Makefile
echo "test-auth:" >> Makefile
echo "	@tests/integration/auth/test-login.sh" >> Makefile
```

### 添加单元测试
```bash
# 1. 创建测试文件
touch core/internal/auth/auth_test.go

# 2. 编写测试
cat > core/internal/auth/auth_test.go << 'EOF'
package auth

import "testing"

func TestValidateToken(t *testing.T) {
    // 测试逻辑
}
EOF
```

## 贡献指南

1. **提交前**: 运行 `make test-all`
2. **新功能**: 添加相应测试
3. **重构**: 确保测试通过
4. **文档**: 更新测试文档

---

**快速开始**:
```bash
make test-config  # 推荐的测试入口
```