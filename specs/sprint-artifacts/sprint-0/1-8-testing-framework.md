# Story 1.8: 测试框架与工具集
# Sprint 0: Infrastructure建设

**Priority**: P1  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: Story 1  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [测试规范](../../standards/testing-standards.md)

---

## User Story

作为开发者，我希望有完善的测试框架和工具，以便编写单元测试、集成测试和 E2E 测试。

---

## Acceptance Criteria

- [ ] 集成 testify 断言库
- [ ] 配置测试数据库（Docker）
- [ ] 创建测试辅助函数
- [ ] 编写单元测试示例
- [ ] 编写集成测试示例
- [ ] 配置测试脚本（Makefile）
- [ ] 编写测试文档

---

## Implementation Tasks

- [ ] 添加依赖（testify、sqlmock）
- [ ] 创建 `core/internal/testutil` 包
- [ ] 实现测试数据库辅助函数
- [ ] 实现 HTTP 测试辅助函数
- [ ] 编写单元测试示例（response、errors）
- [ ] 编写集成测试示例（config API）
- [ ] 更新 Makefile（test、test-coverage）
- [ ] 编写测试指南文档

---

## Technical Details

### 测试辅助函数

```go
// core/internal/testutil/database.go

package testutil

import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/yourusername/apprun/core/ent"
)

func SetupTestDB(t *testing.T) *ent.Client {
    client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
    require.NoError(t, err)
    
    err = client.Schema.Create(context.Background())
    require.NoError(t, err)
    
    t.Cleanup(func() {
        client.Close()
    })
    
    return client
}
```

```go
// core/internal/testutil/http.go

package testutil

import (
    "net/http/httptest"
    "testing"
)

func NewTestRequest(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
    // 实现
}
```

### 单元测试示例

```go
// core/pkg/response/response_test.go

func TestSuccess(t *testing.T) {
    w := httptest.NewRecorder()
    response.Success(w, map[string]string{"message": "ok"})
    
    assert.Equal(t, http.StatusOK, w.Code)
    // 更多断言...
}
```

### Makefile 测试命令

```makefile
# 单元测试
test:
	cd core && go test -v -race ./...

# 代码覆盖率
test-coverage:
	cd core && go test -v -race -coverprofile=coverage.out ./...
	cd core && go tool cover -html=coverage.out -o coverage.html

# 集成测试
test-integration:
	cd tests && ./scripts/run-integration-tests.sh
```

---

## Test Cases

### Unit Tests
- [ ] 测试辅助函数正常工作（SetupTestDB, NewTestRequest）
- [ ] response package 覆盖率 > 80%
- [ ] errors package 覆盖率 > 80%
- [ ] Mock 测试正常工作

### Integration Tests
- [ ] HTTP Handler 端到端测试通过
- [ ] 数据库集成测试通过
- [ ] Redis 集成测试通过（如有）
- [ ] Makefile test 命令正常执行

---

## Test Templates

### Unit Test Template
```go
// Example: core/pkg/response/response_test.go
package response_test

import (
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "apprun/pkg/response"
)

func TestSuccess(t *testing.T) {
    // Arrange
    w := httptest.NewRecorder()
    data := map[string]string{"message": "ok"}
    
    // Act
    response.Success(w, data)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "\"success\":true")
    assert.Contains(t, w.Body.String(), "\"message\":\"ok\"")
}
```

### Integration Test Template
```go
// Example: tests/integration/config_api_test.go
package integration

import (
    "context"
    "testing"
    "github.com/stretchr/testify/suite"
    "apprun/internal/testutil"
)

type ConfigAPITestSuite struct {
    suite.Suite
    db     *ent.Client
    router http.Handler
}

func (s *ConfigAPITestSuite) SetupTest() {
    s.db = testutil.SetupTestDB(s.T())
    s.router = setupRouter(s.db)
}

func (s *ConfigAPITestSuite) TestCreateConfig_Success() {
    // Arrange
    payload := `{"key":"test","value":"value1"}`
    
    // Act
    resp := testutil.DoRequest(s.T(), s.router, "POST", 
        "/api/v1/configs", strings.NewReader(payload))
    
    // Assert
    s.Equal(http.StatusCreated, resp.Code)
    s.Contains(resp.Body.String(), "\"success\":true")
}

func TestConfigAPISuite(t *testing.T) {
    suite.Run(t, new(ConfigAPITestSuite))
}
```

### Table-Driven Test Template
```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errCode string
    }{
        {
            name:    "valid input",
            input:   "valid@example.com",
            wantErr: false,
        },
        {
            name:    "invalid email",
            input:   "invalid-email",
            wantErr: true,
            errCode: "VAL_INVALID_EMAIL_001",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errCode)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## Definition of Done

- [ ] **Core Package Implementation**
  - [ ] `testutil` package created with helper functions
  - [ ] SetupTestDB() working with in-memory SQLite
  - [ ] HTTP test helpers (NewTestRequest, etc.) implemented
  - [ ] Mock interfaces created for external dependencies

- [ ] **Test Coverage**
  - [ ] Unit tests for response package (>80% coverage)
  - [ ] Unit tests for errors package (>80% coverage)
  - [ ] Integration test examples created (config API)
  - [ ] All tests passing: `go test ./... -v`

- [ ] **Tools & CI Integration**
  - [ ] testify dependency added to go.mod
  - [ ] Makefile targets: test, test-unit, test-integration, test-cover
  - [ ] CI pipeline updated to run tests
  - [ ] Coverage report generated and uploaded (codecov)

- [ ] **Documentation**
  - [ ] README.md with test templates and examples
  - [ ] Testing standards document updated
  - [ ] Team training session completed (optional)
  - [ ] Code reviewed and approved (2 reviewers)

- [ ] **Quality Gates**
  - [ ] Zero test failures in CI
  - [ ] Test execution time < 2 minutes (unit tests)
  - [ ] No flaky tests identified
  - [ ] golangci-lint passes on test files

---

## Related Docs

- [测试规范](../../standards/testing-standards.md)
- [testify 文档](https://github.com/stretchr/testify)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

**Created**: 2025-12-27  
**Updated**: 2026-01-15  
**Maintainer**: Scrum Master (Bob)
