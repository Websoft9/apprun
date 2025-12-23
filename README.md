# apprun

一个轻量级的配置中心服务，使用 Go 开发。

## 快速开始

### 开发环境

```bash
# 克隆项目
git clone <repository-url>
cd apprun

# 启动开发环境
make dev

# 访问服务
curl http://localhost:8080/health
```

### 测试

```bash
# 运行所有测试
make test-all

# 运行配置模块测试（推荐）
make test-config

# 查看测试覆盖率
make test-unit
```

## 项目结构

```
apprun/
├── core/                    # Go 核心应用
│   ├── cmd/server/         # 主程序入口
│   ├── internal/           # 内部包
│   │   ├── config/         # 配置管理
│   │   ├── handlers/       # HTTP 处理
│   │   └── services/       # 业务逻辑
│   └── ent/                # 数据库模型 (Ent ORM)
├── tests/                  # 测试套件
│   ├── common.sh          # 共享测试工具
│   ├── integration/       # 集成测试
│   └── scripts/           # 测试辅助脚本
├── docker/                 # Docker 配置
├── docs/                   # 文档
└── Makefile               # 构建脚本
```

## 核心特性

- **配置优先级**: 环境变量 > 数据库 > 配置文件
- **RESTful API**: 完整的 CRUD 操作
- **Docker 支持**: 容器化部署
- **测试驱动**: 完整的测试套件

## API 接口

### 配置管理

```bash
# 获取所有配置
GET /config

# 获取特定配置
GET /config/{path}

# 更新配置
PUT /config
Content-Type: application/json
{
  "path": "value"
}
```

### 健康检查

```bash
GET /health
```

## 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `W9_APP_NAME` | 应用名称 | apprun |
| `W9_HTTP_PORT` | HTTP 端口 | 8080 |
| `W9_DATABASE_URL` | 数据库连接 | postgres://... |
| `W9_LOG_LEVEL` | 日志级别 | info |

## 开发

### 构建

```bash
# 构建应用
make build

# 构建 Docker 镜像
make docker-build
```

### 测试

详细的测试指南请参考 [docs/testing.md](docs/testing.md)。

### 代码规范

- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 提交前运行 `make test-all`

## 部署

### Docker Compose

```bash
# 启动服务
docker compose up -d

# 查看日志
docker compose logs -f apprun
```

### 生产环境

```bash
# 构建生产镜像
make docker-build-prod

# 部署
docker run -p 8080:8080 apprun:latest
```

## 贡献

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 文档

- [API 文档](docs/api.md)
- [测试指南](docs/testing.md)
- [技术架构](docs/architecture/)
- [产品需求](docs/prd.md)