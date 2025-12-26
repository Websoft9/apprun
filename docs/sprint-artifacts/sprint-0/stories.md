# Sprint 0: 基础设施建设
# apprun BaaS Platform

**Sprint 周期**: 2025-12-26 ~ 2026-01-09 (2 周)  
**Sprint 目标**: 搭建开发基础设施，建立编码规范和工具链  
**负责人**: Dev Team Lead  
**状态**: Planning

---

## Sprint 目标

### 核心目标
实现通用技术规范的基础代码，为后续业务 Epic 开发提供标准化工具和框架。

### 验收标准
- [ ] 统一响应工具包可用
- [ ] 错误处理框架可用
- [ ] Ent Schema 规范配置完成
- [ ] CI/CD Linter 检查通过
- [ ] 测试框架就绪
- [ ] 所有代码通过 golangci-lint 检查

---

## Stories

### Story 1: 统一响应工具包

**优先级**: P0  
**工作量**: 2 天  
**负责人**: Backend Dev  
**关联规范**: [API 设计规范](../../standards/api-design.md#41-统一响应格式)

#### 用户故事
作为开发者，我希望有统一的响应工具包，以便快速实现标准化的 API 响应格式。

#### 验收标准
- [ ] 创建 `core/pkg/response` 包
- [ ] 实现 `Success()` 函数（成功响应）
- [ ] 实现 `Error()` 函数（错误响应）
- [ ] 实现 `List()` 函数（列表响应含分页）
- [ ] 编写单元测试（覆盖率 > 90%）
- [ ] 编写使用文档和示例

#### 实现任务
- [ ] 创建 `core/pkg/response/response.go`
- [ ] 定义响应结构体（Response、ErrorInfo、PaginationInfo）
- [ ] 实现 Success 函数
- [ ] 实现 Error 函数
- [ ] 实现 List 函数（含分页）
- [ ] 编写单元测试
- [ ] 编写 README.md（使用示例）
- [ ] 更新现有 Handler（config.go）使用新工具包

#### 技术细节
```go
// core/pkg/response/response.go

package response

import (
    "encoding/json"
    "net/http"
)

type Response struct {
    Success bool        `json:"success"`
    Code    int         `json:"code"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type PaginationInfo struct {
    Total      int `json:"total"`
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    TotalPages int `json:"total_pages"`
}

func Success(w http.ResponseWriter, data interface{}) {
    // 实现
}

func Error(w http.ResponseWriter, code int, errCode, message string) {
    // 实现
}

func List(w http.ResponseWriter, items interface{}, pagination *PaginationInfo) {
    // 实现
}
```

#### 测试用例
- 成功响应格式正确
- 错误响应包含完整错误信息
- 列表响应包含分页信息
- JSON 序列化正确

---

### Story 2: 错误处理框架

**优先级**: P0  
**工作量**: 2 天  
**负责人**: Backend Dev  
**关联规范**: [API 设计规范](../../standards/api-design.md#5-错误码规范)

#### 用户故事
作为开发者，我希望有标准化的错误处理框架，以便统一管理错误码和错误消息。

#### 验收标准
- [ ] 创建 `core/pkg/errors` 包
- [ ] 定义标准错误码（认证、权限、资源、验证、系统）
- [ ] 实现自定义错误类型
- [ ] 实现错误码映射 HTTP 状态码
- [ ] 编写单元测试（覆盖率 > 90%）
- [ ] 编写错误码文档

#### 实现任务
- [ ] 创建 `core/pkg/errors/errors.go`
- [ ] 创建 `core/pkg/errors/codes.go`
- [ ] 定义 AppError 结构体
- [ ] 实现错误构造函数（New, Wrap）
- [ ] 实现 HTTP 状态码映射
- [ ] 定义所有错误码常量
- [ ] 编写单元测试
- [ ] 编写错误码文档（README.md）

#### 技术细节
```go
// core/pkg/errors/errors.go

package errors

type AppError struct {
    Code       string                 // 错误码
    Message    string                 // 错误消息
    HTTPStatus int                    // HTTP 状态码
    Details    map[string]interface{} // 详细信息
    Err        error                  // 原始错误
}

func (e *AppError) Error() string {
    return e.Message
}

func New(code, message string, httpStatus int) *AppError {
    // 实现
}

func Wrap(err error, code, message string, httpStatus int) *AppError {
    // 实现
}
```

```go
// core/pkg/errors/codes.go

package errors

// 认证错误
const (
    ErrAuthInvalidToken   = "AUTH_INVALID_TOKEN"
    ErrAuthTokenExpired   = "AUTH_TOKEN_EXPIRED"
    ErrAuthUnauthorized   = "AUTH_UNAUTHORIZED"
)

// 权限错误
const (
    ErrPermForbidden        = "PERM_FORBIDDEN"
    ErrPermInsufficientRole = "PERM_INSUFFICIENT_ROLE"
)

// 资源错误
const (
    ErrResNotFound      = "RES_NOT_FOUND"
    ErrResAlreadyExists = "RES_ALREADY_EXISTS"
)

// ... 更多错误码
```

#### 测试用例
- AppError 正确创建和包装
- HTTP 状态码映射正确
- 错误信息包含完整上下文

---

### Story 3: Ent Schema 规范配置

**优先级**: P0  
**工作量**: 1 天  
**负责人**: Backend Dev  
**关联规范**: [编码规范 - Ent ORM](../../standards/coding-standards.md#12-ent-orm-规范)

#### 用户故事
作为开发者，我希望 Ent Schema 遵循统一规范，以便 API 响应字段格式一致。

#### 验收标准
- [ ] 现有 Ent Schema 添加 JSON tag（snake_case）
- [ ] 创建 Ent Schema 模板
- [ ] 编写 Ent Schema 检查脚本
- [ ] 检查脚本集成到开发流程
- [ ] 所有 Schema 通过规范检查

#### 实现任务
- [ ] 更新 `ent/schema/users.go`（添加 JSON tag）
- [ ] 更新 `ent/schema/servers.go`（添加 JSON tag）
- [ ] 更新 `ent/schema/configitem.go`（添加 JSON tag）
- [ ] 创建 `scripts/check-ent-json-tags.sh`
- [ ] 添加执行权限
- [ ] 在 Makefile 中添加 `ent-check` 目标
- [ ] 运行 `go generate ./ent` 重新生成代码
- [ ] 验证 API 响应字段格式

#### 技术细节
```go
// ent/schema/users.go 示例

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").
            StorageKey("id").
            StructTag(`json:"user_id"`),
        
        field.String("email").
            StorageKey("email").
            StructTag(`json:"email"`),
        
        field.Time("created_at").
            StorageKey("created_at").
            StructTag(`json:"created_at"`).
            Default(time.Now),
    }
}
```

#### 测试用例
- 检查脚本正确识别缺少 JSON tag 的字段
- 检查脚本正确识别 CamelCase 的 JSON tag
- 所有现有 Schema 通过检查

---

### Story 4: CI/CD Linter 检查配置

**优先级**: P0  
**工作量**: 1 天  
**负责人**: DevOps/Backend Dev  
**关联规范**: [编码规范 - 工具配置](../../standards/coding-standards.md#a-工具配置)

#### 用户故事
作为开发团队，我希望 CI/CD 自动检查代码规范，以便及早发现代码质量问题。

#### 验收标准
- [ ] golangci-lint 配置完成
- [ ] GitHub Actions CI 配置完成
- [ ] Ent Schema 检查集成到 CI
- [ ] PR 自动触发检查
- [ ] 所有检查通过

#### 实现任务
- [ ] 创建 `.golangci.yml` 配置文件
- [ ] 创建 `.github/workflows/ci.yml`
- [ ] 配置 golangci-lint job
- [ ] 配置 ent-check job
- [ ] 配置单元测试 job
- [ ] 配置代码覆盖率上传
- [ ] 在 README 中添加 CI 状态徽章
- [ ] 测试 CI 流程

#### 技术细节
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
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
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
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

#### 测试用例
- Push 到 main/develop 触发 CI
- PR 创建触发 CI
- Linter 检查失败时 CI 失败
- 所有测试通过时 CI 成功

---

### Story 5: 测试框架与工具包

**优先级**: P1  
**工作量**: 2 天  
**负责人**: Backend Dev  
**关联规范**: [编码规范 - 测试规范](../../standards/coding-standards.md#6-测试规范)

#### 用户故事
作为开发者，我希望有统一的测试工具包，以便快速编写高质量的测试用例。

#### 验收标准
- [ ] 创建 `pkg/testutil` 测试工具包
- [ ] 实现 Mock HTTP 请求辅助函数
- [ ] 实现数据库测试辅助函数（基于 Ent）
- [ ] 实现断言辅助函数
- [ ] 编写测试示例
- [ ] 编写测试指南文档

#### 实现任务
- [ ] 创建 `pkg/testutil` 包
- [ ] 实现 HTTP 测试辅助函数
- [ ] 实现数据库测试辅助函数
- [ ] 实现 Mock 工具
- [ ] 创建测试示例（example_test.go）
- [ ] 编写测试指南（docs/standards/testing-guide.md）
- [ ] 为现有代码添加示例测试

#### 技术细节
```go
// pkg/testutil/http.go

package testutil

import (
    "net/http"
    "net/http/httptest"
)

// NewRequest 创建测试请求
func NewRequest(method, path string, body interface{}) *http.Request {
    // 实现
}

// NewRecorder 创建响应记录器
func NewRecorder() *httptest.ResponseRecorder {
    return httptest.NewRecorder()
}

// AssertJSON 断言 JSON 响应
func AssertJSON(t *testing.T, w *httptest.ResponseRecorder, expected interface{}) {
    // 实现
}
```

```go
// pkg/testutil/db.go

package testutil

import (
    "context"
    "testing"
    "apprun/ent"
)

// SetupTestDB 创建测试数据库
func SetupTestDB(t *testing.T) *ent.Client {
    // 实现
}

// TeardownTestDB 清理测试数据库
func TeardownTestDB(t *testing.T, client *ent.Client) {
    // 实现
}
```

#### 测试用例
- HTTP 测试辅助函数正常工作
- 数据库测试辅助函数可创建和清理测试数据
- 示例测试通过

---

### Story 6: 更新现有代码使用新工具

**优先级**: P1  
**工作量**: 1 天  
**负责人**: Backend Dev

#### 用户故事
作为开发者，我希望现有代码使用新的工具包，以便验证工具包的可用性。

#### 验收标准
- [ ] `core/handlers/config.go` 使用 response 包
- [ ] 错误处理使用 errors 包
- [ ] 所有 API 响应格式统一
- [ ] 现有测试通过
- [ ] 编写集成测试

#### 实现任务
- [ ] 重构 `core/handlers/config.go`
  - [ ] 使用 `response.Success()`
  - [ ] 使用 `response.Error()`
  - [ ] 使用 `errors` 包定义错误
- [ ] 更新 `core/routes/router.go`
  - [ ] 健康检查使用 response 包
- [ ] 编写集成测试
  - [ ] 测试 GET /api/config
  - [ ] 测试 PUT /api/config
  - [ ] 测试 GET /api/config/{key}
- [ ] 运行所有测试确保通过

#### 测试用例
- 配置 API 响应格式符合规范
- 错误响应包含完整错误信息
- 集成测试通过

---

## Sprint 依赖

### 外部依赖
- GitHub Actions (CI/CD)
- Go 1.21+
- golangci-lint

### 工具依赖
- Ent ORM
- Testify (测试框架)
- httptest (HTTP 测试)

---

## Sprint 风险

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| CI/CD 配置复杂 | 中 | 使用标准 GitHub Actions，参考最佳实践 |
| Ent 代码重新生成问题 | 低 | 先备份现有代码，使用版本控制 |
| 现有代码重构工作量 | 中 | 优先重构核心 Handler，其他逐步迁移 |

---

## Sprint 监控指标

- [ ] 代码覆盖率 > 80%
- [ ] golangci-lint 零告警
- [ ] CI 构建时间 < 5 分钟
- [ ] 所有 PR 检查通过率 100%

---

## Sprint 交付物

1. **代码**
   - `core/pkg/response` 包（含测试）
   - `core/pkg/errors` 包（含测试）
   - `pkg/testutil` 包（含示例）
   - 更新后的 Ent Schema
   - 更新后的 Handler 代码

2. **配置**
   - `.golangci.yml`
   - `.github/workflows/ci.yml`
   - `scripts/check-ent-json-tags.sh`
   - 更新后的 Makefile

3. **文档**
   - `core/pkg/response/README.md`
   - `core/pkg/errors/README.md`
   - `docs/standards/testing-guide.md`（可选）

---

## Sprint 回顾准备

### 需要讨论的问题
- 工具包 API 设计是否合理？
- CI/CD 流程是否满足需求？
- 测试框架是否易用？
- 是否需要调整开发流程？

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26
