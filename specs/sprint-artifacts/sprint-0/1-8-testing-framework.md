# Story 1.8: 测试框架与工具集
# Sprint 0: Infrastructure建设

**Priority**: P1  
**Effort**: 2 天  
**Owner**: Backend Dev  
**Dependencies**: Story 1  
**Status**: Completed  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [测试规范](../../standards/testing-decisions.md)

**Completed**: 2026-01-23

---

## User Story

作为开发者，我希望有完善的测试框架和工具，以便编写单元测试、集成测试和 E2E 测试。

---

## Acceptance Criteria

- [x] 集成 testify 断言库
- [x] 配置测试数据库（Docker）
- [x] 创建测试辅助函数
- [x] 编写单元测试示例
- [x] 编写集成测试示例
- [x] 配置测试脚本（Makefile）
- [x] 编写测试文档

---

## Implementation Tasks

- [x] 添加依赖（testify、sqlmock）
- [x] 创建 `tests/testutils` 包
- [x] 实现测试数据库辅助函数
- [x] 实现 HTTP 测试辅助函数
- [x] 编写单元测试示例（response、errors）
- [x] 编写集成测试示例（config API）
- [x] 更新 Makefile（test、test-coverage）
- [x] 编写测试指南文档

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

> **Note**: 测试命令已集成到项目根目录的 Makefile 中，使用 `make help` 查看所有可用的测试命令。

---

## Test Cases

### Unit Tests
- [x] 测试辅助函数正常工作（SetupTestDB, NewTestRequest）
- [x] response package 覆盖率 > 80%
- [x] errors package 覆盖率 > 80%
- [x] Mock 测试正常工作

### Integration Tests
- [x] HTTP Handler 端到端测试通过
- [x] 数据库集成测试通过
- [x] Redis 集成测试通过（如有）
- [x] Makefile test 命令正常执行

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

- [x] **Core Package Implementation**
  - [x] `testutil` package created with helper functions (`tests/testutils/`)
  - [x] SetupTestDB() working with PostgreSQL test database
  - [x] HTTP test helpers (HTTPTestClient, etc.) implemented
  - [x] Mock interfaces created for external dependencies

- [x] **Test Coverage**
  - [x] Unit tests for response package (437 lines, comprehensive coverage)
  - [x] Unit tests for errors package (301 lines, comprehensive coverage)
  - [x] Integration test examples created (auth, config, api, db, audit, cli)
  - [x] All tests passing: 76 test files created

- [x] **Tools & CI Integration**
  - [x] testify dependency added to go.mod (v1.11.1)
  - [x] Makefile targets: test, test-unit, test-integration, test-cover, test-e2e
  - [x] CI pipeline ready for test execution
  - [x] Coverage report generation configured

- [x] **Documentation**
  - [x] tests/README.md with test templates and examples (616 lines)
  - [x] Testing standards document created (testing-decisions.md)
  - [x] Test design document created (_bmad-output/test-design-apprun-system.md)
  - [x] Code reviewed and approved

- [x] **Quality Gates**
  - [x] Zero test failures in test runs
  - [x] Test execution time optimized
  - [x] Test framework stable and reliable
  - [x] golangci-lint integration configured

---

## Related Docs

- [测试决策文档](../../standards/testing-decisions.md)
- [测试框架 README](../../../tests/README.md)
- [测试设计文档](../../../_bmad-output/test-design-apprun-system.md)
- [项目 Makefile](../../../Makefile) - 查看所有测试命令
- [testify 文档](https://github.com/stretchr/testify)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

**Created**: 2025-12-27  
**Updated**: 2026-01-23  
**Completed**: 2026-01-23  
**Maintainer**: Scrum Master (Bob)
