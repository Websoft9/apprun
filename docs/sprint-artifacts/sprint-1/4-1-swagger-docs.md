# Story 4.1: Swagger API Documentation

**Epic**: Sprint-0 基础设施  
**Priority**: Medium  
**Points**: 2  
**Status**: Done  
**Sprint**: Sprint-0

---

## 📋 User Story

**As a** API Consumer (Frontend Developer / Third-party Developer)  
**I want** interactive API documentation with Swagger UI  
**So that** I can easily understand, test, and integrate with REST APIs without reading source code

---

## 🎯 Acceptance Criteria

### 1. Swagger UI 访问
- [x] Swagger UI 可通过 `/api/docs/` 访问（自定义路径）
- [x] OpenAPI spec 可通过 `/api/docs/doc.json` 获取
- [x] 页面加载完整，样式正常，可交互
- [x] 支持"Try it out"功能在线测试 API（直接调用真实端点）

### 2. API 文档自动生成
- [x] 使用 Swaggo 自动生成 OpenAPI 3.0 规范
- [x] 通过代码注解自动更新文档（无需手动维护）
- [x] 包含所有现有 REST API 端点（当前：配置模块 5 个端点）

### 3. 文档完整性
- [x] 每个端点包含：
  - 请求方法（GET/POST/PUT/DELETE）
  - 路径参数、查询参数、请求体
  - 响应状态码（200/400/404/500）
  - 响应示例（JSON）
- [x] 包含数据模型定义（Request/Response 结构体）
- [x] 包含错误响应示例

### 4. 开发体验
- [x] `make swagger` 命令生成/更新文档
- [x] CI 流程验证文档与代码同步
- [x] 文档自动部署（开发/测试环境）

---

## 📦 Deliverables

### 1. Swagger 配置文件
- `docs/docs.go` - 自动生成的嵌入式文档（编译进二进制）
- `docs/swagger.yaml` - OpenAPI 规范（YAML 格式）
- `docs/swagger.json` - OpenAPI 规范（JSON 格式）

### 2. Handler 注解
- `core/modules/config/handler.go` - 添加 Swaggo 注解
  ```go
  // @Summary      Get configuration item
  // @Description  Query a single configuration item by key
  // @Tags         config
  // @Accept       json
  // @Produce      json
  // @Param        key  query  string  true  "Configuration key"
  // @Success      200  {object}  GetConfigResponse
  // @Failure      400  {object}  ErrorResponse
  // @Router       /config [get]
  ```

### 3. Swagger 中间件
- `core/routes/swagger.go` - Swagger UI 路由注册（挂载到 `/api/docs/`）
- `cmd/server/main.go` - 导入生成的文档包（`import _ "apprun/docs"`）

### 4. 文档与脚本
- `Makefile` - 添加 `swagger` target
- `docs/standards/api-design.md` - 更新 API 文档规范
- `README.md` - 添加 Swagger 访问说明

---

## 🔧 Technical Design

### 工具选型：Swaggo

**理由**:
- ✅ 轻量级，与 Go 生态集成好
- ✅ 注解驱动，代码即文档
- ✅ 支持 OpenAPI 3.0
- ✅ 与 chi router 兼容

**依赖包**:
```go
github.com/swaggo/swag         // CLI 工具
github.com/swaggo/http-swagger // HTTP handler
github.com/swaggo/files        // 静态文件
```

### 架构集成

```
core/
├── cmd/server/
│   └── main.go                 # 导入 docs，Swagger 初始化
├── routes/
│   ├── router.go              # 原有路由
│   └── swagger.go             # Swagger 路由（新增）
├── modules/config/
│   └── handler.go             # 添加注解
└── docs/                       # 自动生成（swag init）
    ├── docs.go                # 嵌入式文档（编译进二进制）
    ├── swagger.yaml           # OpenAPI 规范（YAML）
    └── swagger.json           # OpenAPI 规范（JSON）
```

### 静态资源嵌入机制

**Swaggo 自动嵌入原理**（零部署依赖）：

```
编译时: swag init → 生成 docs/docs.go（OpenAPI spec 转为 Go 常量）
运行时: import _ "apprun/docs" → init() 注册到内存 → 单一二进制文件
访问:   /api/docs/ → http-swagger 从内存提供 UI 和 spec
```

**关键优势**：
- ✅ 单一可执行文件部署
- ✅ 容器镜像无需静态文件
- ✅ 代码注解自动同步文档

### 注解规范

**端点注解模板**：
```go
// @Summary      <Short description>
// @Description  <Detailed explanation>
// @Tags         <Group name>
// @Param        <param_name>  <location>  <type>  <required>  "<description>"
// @Success      200  {object}  <ResponseStruct>
// @Failure      400  {object}  ErrorResponse
// @Router       /<path> [<method>]
```

**示例**（配置模块）：
```go
// @Summary      Get configuration item
// @Tags         config
// @Param        key  query  string  true  "Configuration key"
// @Success      200  {object}  GetConfigResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /config [get]
```

### 开发流程

```bash
# 1. 添加注解
vim core/modules/config/handler.go

# 2. 生成文档
make swagger

# 3. 访问测试
curl http://localhost:8080/api/docs/
```

---

## 🧪 Testing Strategy

### 验证清单
- [x] 所有 API 端点在 Swagger UI 中可见
- [x] "Try it out"功能正常工作
- [x] 请求/响应示例准确
- [x] 错误响应文档完整
- [x] 本地生成的文档与代码同步

### 集成测试
```bash
# CI 验证流程
make swagger                    # 生成文档
git diff --exit-code docs/      # 确保已提交
curl /api/docs/doc.json | jq    # 验证 spec
```

---

## 📝 Notes

### 文档规范
- **注解位置**：紧邻函数定义
- **Tags 分组**：config、user、server 等模块化分组
- **错误响应**：统一使用 ErrorResponse 结构
- **必填字段**：明确标注 true/false

### 提交规范
- 每次修改 API 必须同步更新注解
- 提交前运行 `make swagger` 并提交 `docs/` 目录
- CI 流程自动验证文档同步

---

## ✅ Definition of Done

- [x] Swaggo 依赖包安装完成
- [x] 配置模块 5 个端点添加完整注解
- [x] Swagger UI 可通过 `/api/docs/` 访问
- [x] OpenAPI spec 可通过 `/api/docs/doc.json` 获取
- [x] `make swagger` 命令正常工作（生成 `docs/docs.go`）
- [x] 文档嵌入二进制，单一可执行文件部署
- [x] 文档包含请求/响应示例
- [x] 错误响应文档完整（400/404/500）
- [x] `docs/` 目录生成并提交到 Git
- [x] `README.md` 更新 API 文档访问说明
- [x] CI 流程验证文档同步
- [x] 本地测试所有端点"Try it out"功能

---

## 🔗 Dependencies

**依赖**:
- Story 10: Configuration Center Foundation（已完成）

**被依赖**:
- Story 12-20: 后续所有 API 开发（需遵循文档规范）

---

**Created**: 2025-12-29  
**Author**: Analyst Agent  
**Estimated Effort**: 4-6 hours
