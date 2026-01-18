# apprun-core

**apprun BaaS 平台核心服务** - 基于 GoFr + CNCF 生态的轻量级后端即服务平台

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![GoFr](https://img.shields.io/badge/GoFr-1.26+-7C3AED?style=flat)](https://gofr.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 📋 项目简介

apprun-core 是 apprun BaaS 平台的核心服务，采用 **GoFr 框架** 和 **CNCF 生态组件**，提供：

- ✅ **开箱即用的企业级特性**：零配置可观测性、自动健康检查、内置指标收集
- ✅ **减少 90% 基础代码**：专注业务逻辑，摆脱基础设施重复工作
- ✅ **轻量级部署**：单二进制，最小 512MB 内存即可运行核心服务
- ✅ **多租户架构**：原生支持租户隔离，自动数据过滤
- ✅ **事件驱动**：NATS 事件总线 + Temporal 工作流引擎
- ✅ **云原生**：基于 CNCF 项目，天然适配 Kubernetes

## 🏗️ 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| **框架** | [GoFr 1.26+](https://gofr.dev) | 企业级微服务框架 |
| **可观测性** | OpenTelemetry | 自动追踪、日志、指标 |
| **事件总线** | NATS | 轻量级消息队列 |
| **数据库** | PostgreSQL 15+ | 主数据库 |
| **ORM** | GORM 1.25+ | Go ORM 框架 |
| **认证** | Ory Kratos | 身份认证服务 |
| **授权** | Casbin | RBAC 引擎 |
| **工作流** | Temporal | 工作流编排引擎 |
| **监控** | Prometheus + Grafana | 监控和可视化 |
| **追踪** | Jaeger | 分布式追踪 |

## 📁 项目结构

```
core/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── internal/
│   ├── config/                  # 配置管理
│   ├── handlers/                # HTTP 处理器
│   │   ├── auth/                # 认证 API
│   │   ├── datamodel/           # 数据模型 API
│   │   └── workflow/            # 工作流 API
│   ├── middleware/              # 中间件
│   │   ├── auth.go              # JWT 认证
│   │   └── tenant.go            # 多租户隔离
│   ├── models/                  # 数据模型 (GORM)
│   └── services/                # 业务逻辑层
├── pkg/                         # 公共库
│   ├── kratos/                  # Kratos 客户端封装
│   └── temporal/                # Temporal 客户端封装
├── configs/                     # 配置文件
├── docker-compose.yml           # 本地开发环境
├── Dockerfile                   # 生产镜像
├── Makefile                     # 开发工具
└── go.mod                       # Go 模块定义
```

## 🚀 快速开始

### 前置要求

- Go 1.24+ 
- Docker & Docker Compose
- Make (可选)
- **Atlas CLI** (必需) - 用于数据库迁移

#### 安装 Atlas CLI

```bash
# Linux / macOS
curl -sSf https://atlasgo.sh | sh

# 或使用 Homebrew (macOS)
brew install ariga/tap/atlas

# 验证安装
atlas version
```

> **重要**: apprun 的数据库迁移系统基于 Atlas 工具链，需要安装 atlas CLI 二进制文件。
> 详见：https://atlasgo.io/getting-started

### 1. 克隆项目

```bash
git clone https://github.com/Websoft9/apprun.git
cd apprun/core
```

### 2. 复制环境变量

```bash
cp .env.example .env
```

### 3. 启动开发环境

#### 方式 A：使用 Docker Compose（推荐）

```bash
# 启动所有服务（包括数据库、NATS、Temporal、Jaeger 等）
make docker-up

# 或者直接使用 docker-compose
docker-compose up -d

# 查看日志
make docker-logs-app
```

#### 方式 B：本地运行

```bash
# 启动依赖服务（数据库、NATS、Temporal 等）
make dev

# 在另一个终端运行应用
make run
```

### 4. 验证服务

```bash
# 健康检查
curl http://localhost:8080/.well-known/health-check

# Prometheus 指标
curl http://localhost:8080/metrics

# API 测试
curl http://localhost:8080/health
```

### 5. 访问 Web UI

- **apprun API**: http://localhost:8080
- **Jaeger UI** (追踪): http://localhost:16686
- **Temporal UI** (工作流): http://localhost:8088
- **Prometheus** (监控): http://localhost:9090
- **Grafana** (可视化): http://localhost:3000 (admin/admin)

## 📖 API 文档

### 认证 API

```bash
# 用户注册
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "Test User",
    "password": "password123",
    "tenant_name": "my-company"
  }'

# 用户登录
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 用户管理 API

```bash
# 获取用户列表（需要认证）
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "name": "New User"
  }'
```

### 工作流 API

```bash
# 启动工作流
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "user-onboarding-001",
    "workflow_type": "OnboardingWorkflow",
    "input": {
      "user_id": 123,
      "email": "user@example.com"
    }
  }'
```

## 🛠️ 开发指南

### 常用命令

```bash
# 查看所有可用命令
make help

# 构建二进制文件
make build

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# 下载依赖
make mod-download

# 整理依赖
make mod-tidy

# 查看 Docker 日志
make docker-logs
```

### 添加新的 API 端点

1. 在 `internal/handlers/` 中创建处理器
2. 在 `cmd/server/main.go` 中注册路由
3. （可选）在 `internal/services/` 中添加业务逻辑
4. （可选）在 `internal/models/` 中定义数据模型

### 数据库迁移

GoFr 支持自动迁移：

```go
// 在 main.go 中
app.Migrate(func(gApp *gofr.Gofr) error {
    return models.AutoMigrate(gApp.GORM())
})
```

## 📊 可观测性

### 自动获得的功能（GoFr 内置）

- ✅ **结构化日志**：JSON 格式，自动包含追踪 ID
- ✅ **分布式追踪**：OpenTelemetry 自动追踪所有请求
- ✅ **Prometheus 指标**：HTTP 请求、数据库查询、NATS 消息等
- ✅ **健康检查**：自动检测数据库、NATS、Redis 等服务状态

### 查看追踪

访问 Jaeger UI: http://localhost:16686

### 查看指标

访问 Prometheus: http://localhost:9090

常用查询：
- `gofr_http_requests_total` - HTTP 请求总数
- `gofr_http_request_duration_seconds` - 请求延迟
- `gofr_db_queries_total` - 数据库查询数

## 🐳 Docker 部署

### 构建镜像

```bash
make docker-build
```

### 推送到镜像仓库

```bash
docker tag apprun-core:latest your-registry/apprun-core:latest
docker push your-registry/apprun-core:latest
```

### 生产环境部署

参考 `docker-compose.yml` 或使用 Kubernetes 部署（见 `/deployments` 目录）

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage
```

## 📝 配置说明

所有配置通过环境变量管理，支持的配置项：

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `APP_PORT` | 8080 | HTTP 服务端口 |
| `DB_HOST` | localhost | 数据库主机 |
| `DB_PORT` | 5432 | 数据库端口 |
| `PUBSUB_NATS_URL` | nats://localhost:4222 | NATS 连接地址 |
| `TEMPORAL_HOST` | localhost:7233 | Temporal 服务地址 |
| `JWT_SECRET` | (必填) | JWT 签名密钥 |
| `LOG_LEVEL` | info | 日志级别 |

完整配置见 `.env.example`

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🔗 相关链接

- [GoFr 文档](https://gofr.dev)
- [Temporal 文档](https://docs.temporal.io)
- [NATS 文档](https://docs.nats.io)
- [Ory Kratos 文档](https://www.ory.sh/docs/kratos)

## 💬 支持

- 📧 Email: support@websoft9.com
- 💬 Discord: [Join our community](https://discord.gg/websoft9)
- 🐛 Issues: [GitHub Issues](https://github.com/Websoft9/apprun/issues)

---

**Built with ❤️ by Websoft9 Team**
