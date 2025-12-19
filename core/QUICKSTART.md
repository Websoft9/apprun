# apprun-core 快速启动指南

## ✅ 项目创建完成！

恭喜！您的 **apprun-core** 项目已经成功创建，包含：

- ✅ 完整的 GoFr 项目结构
- ✅ 数据模型层（User, Tenant, Role, Workflow）
- ✅ 中间件层（JWT 认证、多租户隔离）
- ✅ 服务层（Auth、Event、Workflow）
- ✅ 处理器层（HTTP API）
- ✅ Docker Compose 开发环境
- ✅ Makefile 开发工具

## 📁 项目结构

```
/data/cdl/apprun/core/
├── cmd/server/main.go           ✅ 主程序入口
├── internal/
│   ├── config/config.go         ✅ 配置管理
│   ├── handlers/                ✅ HTTP 处理器
│   │   ├── auth/auth.go         - 认证 API
│   │   ├── datamodel/user.go    - 用户管理 API
│   │   └── workflow/workflow.go - 工作流 API
│   ├── middleware/              ✅ 中间件
│   │   ├── auth.go              - JWT 认证
│   │   └── tenant.go            - 多租户隔离
│   ├── models/models.go         ✅ GORM 数据模型
│   └── services/                ✅ 业务逻辑层
│       ├── auth_service.go      - 认证服务
│       ├── event_service.go     - 事件服务（NATS）
│       └── workflow_service.go  - 工作流服务（Temporal）
├── configs/prometheus.yml       ✅ Prometheus 配置
├── docker-compose.yml           ✅ 开发环境
├── Dockerfile                   ✅ 生产镜像
├── Makefile                     ✅ 开发工具
├── go.mod                       ✅ Go 模块
└── README.md                    ✅ 完整文档
```

## 🚀 下一步操作

### 1. 复制环境变量文件

```bash
cd /data/cdl/apprun/core
cp .env.example .env
```

### 2. 启动开发环境（Docker Compose）

```bash
# 查看所有可用命令
make help

# 启动所有服务（推荐）
make docker-up

# 或者使用 docker-compose
docker-compose up -d
```

这将启动：
- ✅ **apprun-core**: Go 应用（端口 8080）
- ✅ **PostgreSQL**: 数据库（端口 5432）
- ✅ **NATS**: 事件总线（端口 4222）
- ✅ **Temporal**: 工作流引擎（端口 7233）
- ✅ **Temporal UI**: 工作流可视化（端口 8088）
- ✅ **Jaeger**: 分布式追踪（端口 16686）
- ✅ **Prometheus**: 监控（端口 9090）
- ✅ **Grafana**: 可视化（端口 3000）

### 3. 验证服务状态

```bash
# 健康检查
curl http://localhost:8080/.well-known/health-check

# Prometheus 指标
curl http://localhost:8080/metrics

# 简单 API 测试
curl http://localhost:8080/health
```

### 4. 测试认证 API

#### 注册用户

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "name": "Admin User",
    "password": "password123",
    "tenant_name": "my-company"
  }'
```

#### 登录

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

保存返回的 `access_token`，后续 API 调用需要使用。

### 5. 测试用户管理 API

```bash
# 获取用户列表（需要替换 YOUR_ACCESS_TOKEN）
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 6. 访问 Web UI

- **Jaeger UI** (分布式追踪): http://localhost:16686
- **Temporal UI** (工作流管理): http://localhost:8088
- **Prometheus** (监控指标): http://localhost:9090
- **Grafana** (可视化): http://localhost:3000 (admin/admin)

## 🛠️ 开发模式

如果只想本地运行应用（不使用 Docker）：

```bash
# 1. 启动依赖服务
docker-compose up -d postgres nats temporal jaeger

# 2. 等待服务就绪
sleep 5

# 3. 本地运行应用
make run

# 或者直接使用 go run
go run ./cmd/server/main.go
```

## 📝 常用开发命令

```bash
# 查看所有可用命令
make help

# 构建二进制文件
make build

# 运行测试
make test

# 代码格式化
make fmt

# 查看 Docker 日志
make docker-logs

# 只查看应用日志
make docker-logs-app

# 重启服务
make docker-restart

# 停止所有服务
make docker-down
```

## 📊 监控和可观测性

### GoFr 自动提供的功能

1. **结构化日志**
   - JSON 格式
   - 自动包含追踪 ID
   - 查看日志: `make docker-logs-app`

2. **分布式追踪**
   - OpenTelemetry 自动追踪
   - 访问 Jaeger UI: http://localhost:16686
   - 查看请求链路、延迟、错误

3. **Prometheus 指标**
   - 访问: http://localhost:8080/metrics
   - 常用指标：
     - `gofr_http_requests_total` - 请求总数
     - `gofr_http_request_duration_seconds` - 请求延迟
     - `gofr_db_queries_total` - 数据库查询数

4. **健康检查**
   - 端点: `/.well-known/health-check`
   - 自动检测：数据库、NATS、Redis 状态

## 🔧 问题排查

### 1. 服务启动失败

```bash
# 查看日志
make docker-logs

# 检查端口占用
netstat -tunlp | grep -E '8080|5432|4222|7233'

# 重启所有服务
make docker-down && make docker-up
```

### 2. 数据库连接失败

```bash
# 检查 PostgreSQL 状态
docker-compose ps postgres

# 查看数据库日志
docker-compose logs postgres

# 重启数据库
docker-compose restart postgres
```

### 3. Go 依赖问题

```bash
# 重新下载依赖
make mod-download

# 整理依赖
make mod-tidy

# 清理缓存
go clean -modcache
```

## 📚 下一步学习

1. **阅读架构文档**
   - `/data/cdl/apprun/docs/architecture/technical-architecture-apprun-gofr-2025-12-19.md`

2. **查看 GoFr 文档**
   - https://gofr.dev

3. **学习 Temporal 工作流**
   - https://docs.temporal.io

4. **了解 NATS 事件总线**
   - https://docs.nats.io

## 🎯 开发路线图

### 短期任务（接下来 1-2 周）

- [ ] 完善错误处理和日志记录
- [ ] 添加单元测试和集成测试
- [ ] 集成 Ory Kratos 认证服务
- [ ] 实现 RBAC 权限检查（Casbin）
- [ ] 添加 Swagger API 文档

### 中期任务（2-4 周）

- [ ] 完善工作流模块（Temporal）
- [ ] 添加数据模型自动生成
- [ ] 实现存储服务（MinIO/S3）
- [ ] 添加实时推送（WebSocket）
- [ ] 完善监控告警

### 长期任务（1-3 个月）

- [ ] 函数服务模块
- [ ] API 网关模块
- [ ] 多区域部署
- [ ] 性能优化
- [ ] 灰度发布

## 💡 最佳实践

1. **开发流程**
   - 使用 `make dev` 启动开发环境
   - 修改代码后自动重载（使用 air 工具）
   - 编写单元测试和集成测试

2. **Git 提交**
   - 遵循 Conventional Commits 规范
   - 每个功能一个分支
   - PR 前确保测试通过

3. **代码质量**
   - 定期运行 `make fmt` 和 `make lint`
   - 保持测试覆盖率 > 80%
   - 编写清晰的注释和文档

## 🤝 需要帮助？

- 📧 Email: support@websoft9.com
- 💬 Discord: Join our community
- 🐛 Issues: GitHub Issues
- 📖 Docs: `/data/cdl/apprun/core/README.md`

---

**祝开发愉快！🎉**
