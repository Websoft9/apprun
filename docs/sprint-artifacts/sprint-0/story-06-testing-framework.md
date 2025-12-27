# Story 6: 测试框架与工具集
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

- [ ] 测试辅助函数正常工作
- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试通过
- [ ] Makefile 命令正常执行

---

## Related Docs

- [测试规范](../../standards/testing-standards.md)
- [testify 文档](https://github.com/stretchr/testify)

---

**Created**: 2025-12-27  
**Updated**: 2025-12-27  
**Maintainer**: Architect Agent
