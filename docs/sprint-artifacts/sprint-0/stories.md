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
- [ ] 本地开发环境可一键启动
- [ ] CI/CD 自动化流程就绪
- [ ] 生产部署方案可用
- [ ] 统一响应工具包可用
- [ ] 错误处理框架可用
- [ ] Ent Schema 规范配置完成
- [ ] CI/CD Linter 检查通过
- [ ] 测试框架就绪
- [ ] i18n 基础设施就绪
- [ ] l10n 基础设施就绪
- [ ] 所有代码通过 golangci-lint 检查

### Stories 总览

| Story | 描述 | 优先级 | 工期 | 状态 |
|-------|------|--------|------|------|
| Story 9 | 本地开发环境搭建 | P0 | 1 天 | Planning |
| Story 10 | 生产部署配置 | P0 | 2 天 | Planning |
| Story 1 | 统一响应工具包 | P0 | 2 天 | Planning |
| Story 2 | 错误处理框架 | P0 | 2 天 | Planning |
| Story 3 | Ent Schema 规范配置 | P0 | 1 天 | Planning |
| Story 4 | CI/CD Linter 配置 | P0 | 1 天 | Planning |
| Story 5 | 测试框架工具包 | P1 | 2 天 | Planning |
| Story 6 | 重构现有 Handler | P1 | 1 天 | Planning |
| Story 7 | i18n 基础设施 | P1 | 2 天 | Planning |
| Story 8 | l10n 基础设施 | P1 | 2 天 | Planning |

**总工期**: 16 天（P0: 9 天，P1: 7 天）  
**依赖关系**: Story 10 依赖 Story 9，Story 2 依赖 Story 1，Story 6 依赖 Story 1-2，Story 8 依赖 Story 7

---

## Stories

### Story 9: 本地开发环境搭建

**优先级**: P0 ⚡ **最高优先级**  
**工作量**: 1 天  
**负责人**: DevOps/Backend Dev  
**依赖**: 无（第一个完成）

#### 用户故事
作为开发者，我希望能快速搭建本地开发环境，以便在 5 分钟内开始编码。

#### 验收标准
- [ ] 创建 `docker-compose.dev.yml`（开发环境）
- [ ] PostgreSQL 容器配置
- [ ] Redis 容器配置（可选）
- [ ] 环境变量模板 `.env.example`
- [ ] 快速启动脚本 `scripts/dev-setup.sh`
- [ ] 开发文档 `docs/development.md`
- [ ] 开发者能在 5 分钟内启动环境

#### 实现任务
- [ ] 创建 `docker-compose.dev.yml`
- [ ] 创建 `.env.example`（包含所有必需的环境变量）
- [ ] 创建 `scripts/dev-setup.sh`（一键启动脚本）
- [ ] 创建 `docs/development.md`（开发指南）
- [ ] 更新根目录 `Makefile`（添加 `make dev-up` 命令）
- [ ] 添加数据库初始化脚本
- [ ] 测试完整的启动流程

#### 技术细节

```yaml
# docker-compose.dev.yml
version: '3.9'

services:
  postgres:
    image: postgres:16-alpine
    container_name: apprun-postgres-dev
    environment:
      POSTGRES_DB: apprun_dev
      POSTGRES_USER: apprun
      POSTGRES_PASSWORD: apprun_dev_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U apprun"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: apprun-redis-dev
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  postgres_data:
  redis_data:
```

```bash
# .env.example
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=apprun_dev
DB_USER=apprun
DB_PASSWORD=apprun_dev_password
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Server
SERVER_PORT=8080
SERVER_ENV=development

# JWT
JWT_SECRET=your-jwt-secret-here-min-32-chars

# Encryption
ENCRYPTION_KEY=your-32-byte-encryption-key-here
```

```bash
#!/bin/bash
# scripts/dev-setup.sh

set -e

echo "🚀 Starting apprun development environment..."

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker first."
    exit 1
fi

# 检查 Docker Compose
if ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose not found. Please install Docker Compose V2."
    exit 1
fi

# 复制环境变量文件
if [ ! -f .env ]; then
    echo "📝 Creating .env from .env.example..."
    cp .env.example .env
    echo "⚠️  Please update .env with your settings"
fi

# 启动 Docker 容器
echo "🐳 Starting Docker containers..."
docker compose -f docker-compose.dev.yml up -d

# 等待数据库就绪
echo "⏳ Waiting for PostgreSQL..."
until docker exec apprun-postgres-dev pg_isready -U apprun > /dev/null 2>&1; do
    sleep 1
done

echo "✅ Development environment is ready!"
echo ""
echo "📚 Next steps:"
echo "  1. cd core"
echo "  2. go run cmd/server/main.go"
echo "  3. Visit http://localhost:8080"
echo ""
echo "🛠️  Useful commands:"
echo "  - make dev-up      # Start containers"
echo "  - make dev-down    # Stop containers"
echo "  - make dev-logs    # View logs"
echo "  - make dev-clean   # Remove all data"
```

```makefile
# Makefile (根目录)
.PHONY: dev-up dev-down dev-logs dev-clean

# 启动开发环境
dev-up:
	@chmod +x scripts/dev-setup.sh
	@./scripts/dev-setup.sh

# 停止开发环境
dev-down:
	docker compose -f docker-compose.dev.yml down

# 查看日志
dev-logs:
	docker compose -f docker-compose.dev.yml logs -f

# 清理开发环境（包括数据）
dev-clean:
	docker compose -f docker-compose.dev.yml down -v
	rm -f .env
```

#### 测试用例
- 执行 `make dev-up` 成功启动所有容器
- PostgreSQL 健康检查通过
- Redis 健康检查通过
- 可以连接数据库
- 可以运行 Go 应用

---

### Story 10: 生产部署配置

**优先级**: P0 ⚡ **第二优先级**  
**工作量**: 2 天  
**负责人**: DevOps/Backend Dev  
**依赖**: Story 9（本地环境）

#### 用户故事
作为运维人员，我希望有一键部署方案，以便快速在云服务器上部署生产环境。

#### 验收标准
- [ ] 创建 `docker-compose.prod.yml`（生产环境）
- [ ] 创建生产环境 Dockerfile
- [ ] CI/CD 自动构建 Docker 镜像
- [ ] 部署脚本 `scripts/deploy.sh`
- [ ] HTTPS/TLS 配置（Nginx 反向代理）
- [ ] 健康检查和自动重启
- [ ] 部署文档 `docs/deployment.md`
- [ ] 能在 15 分钟内完成生产部署

#### 实现任务
- [ ] 创建 `Dockerfile`（多阶段构建）
- [ ] 创建 `docker-compose.prod.yml`
- [ ] 创建 Nginx 配置 `docker/nginx/nginx.conf`
- [ ] 创建 `.env.prod.example`（生产环境变量模板）
- [ ] 创建 `scripts/deploy.sh`（一键部署脚本）
- [ ] 创建 `docs/deployment.md`（部署指南）
- [ ] 配置 GitHub Actions 自动构建镜像
- [ ] 测试完整的部署流程

#### 技术细节

```dockerfile
# Dockerfile (多阶段构建)
# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git make

# 复制 go mod 文件
COPY core/go.mod core/go.sum ./
RUN go mod download

# 复制源代码
COPY core/ ./

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o server ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/server .
COPY --from=builder /build/config ./config

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行
CMD ["./server"]
```

```yaml
# docker-compose.prod.yml
version: '3.9'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: apprun-app
    restart: unless-stopped
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=${DB_NAME}
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - SERVER_PORT=8080
      - SERVER_ENV=production
      - JWT_SECRET=${JWT_SECRET}
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - apprun-network

  postgres:
    image: postgres:16-alpine
    container_name: apprun-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - apprun-network

  redis:
    image: redis:7-alpine
    container_name: apprun-redis
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--no-auth-warning", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - apprun-network

  nginx:
    image: nginx:alpine
    container_name: apprun-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./docker/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./docker/nginx/ssl:/etc/nginx/ssl:ro
      - nginx_logs:/var/log/nginx
    depends_on:
      - app
    networks:
      - apprun-network

volumes:
  postgres_data:
  redis_data:
  nginx_logs:

networks:
  apprun-network:
    driver: bridge
```

```nginx
# docker/nginx/nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream apprun {
        server app:8080;
    }

    server {
        listen 80;
        server_name _;

        # HTTP 重定向到 HTTPS
        return 301 https://$host$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name _;

        # SSL 证书（使用 Let's Encrypt 或自签名）
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;

        # SSL 配置
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;

        # 日志
        access_log /var/log/nginx/access.log;
        error_log /var/log/nginx/error.log;

        # 代理配置
        location / {
            proxy_pass http://apprun;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;

            # 超时配置
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }

        # 健康检查（不记录日志）
        location /health {
            proxy_pass http://apprun;
            access_log off;
        }
    }
}
```

```bash
#!/bin/bash
# scripts/deploy.sh

set -e

echo "🚀 Deploying apprun to production..."

# 检查环境变量文件
if [ ! -f .env.prod ]; then
    echo "❌ .env.prod not found. Please create it from .env.prod.example"
    exit 1
fi

# 加载环境变量
export $(grep -v '^#' .env.prod | xargs)

# 停止旧容器
echo "🛑 Stopping old containers..."
docker compose -f docker-compose.prod.yml down

# 拉取最新代码
echo "📥 Pulling latest code..."
git pull origin main

# 构建新镜像
echo "🔨 Building Docker images..."
docker compose -f docker-compose.prod.yml build --no-cache

# 启动新容器
echo "🐳 Starting containers..."
docker compose -f docker-compose.prod.yml up -d

# 等待服务就绪
echo "⏳ Waiting for services..."
sleep 10

# 健康检查
echo "🏥 Running health check..."
until curl -f http://localhost/health > /dev/null 2>&1; do
    echo "Waiting for app..."
    sleep 5
done

echo "✅ Deployment successful!"
echo ""
echo "📊 Service status:"
docker compose -f docker-compose.prod.yml ps
echo ""
echo "🌐 Application is running at:"
echo "  - HTTP: http://your-domain.com"
echo "  - HTTPS: https://your-domain.com"
echo ""
echo "📝 View logs: docker compose -f docker-compose.prod.yml logs -f"
```

```yaml
# .github/workflows/docker-build.yml
name: Build and Push Docker Image

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - name: Checkout code
      uses: actions/checkout@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v4
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}

    - name: Build and push
      uses: docker/build-push-action@v4
      with:
        context: .
        push: ${{ github.event_name != 'pull_request' }}
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
```

#### 测试用例
- Docker 镜像成功构建
- 执行 `scripts/deploy.sh` 成功部署
- 所有容器健康检查通过
- HTTPS 访问正常
- 健康检查端点响应正常
- 容器自动重启工作正常

---

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

### Story 7: i18n 基础设施

**优先级**: P1  
**工作量**: 2 天  
**负责人**: Backend Dev  
**关联规范**: [i18n 规范](../../standards/i18n-standards.md)

#### 用户故事
作为开发者，我希望有国际化（i18n）基础设施，以便支持多语言用户。

#### 验收标准
- [ ] 创建 `core/pkg/i18n` 包
- [ ] 集成 go-i18n v2 库
- [ ] 实现语言检测中间件
- [ ] 创建英文和中文消息文件
- [ ] 编写单元测试（覆盖率 > 80%）
- [ ] 编写使用文档

#### 实现任务
- [ ] 安装 go-i18n 依赖
- [ ] 创建 `core/pkg/i18n/i18n.go`（初始化）
- [ ] 创建 `core/pkg/i18n/middleware.go`（Chi 中间件）
- [ ] 创建消息文件目录 `locales/`
- [ ] 创建 `locales/en.yaml`（英文）
- [ ] 创建 `locales/zh-CN.yaml`（中文）
- [ ] 实现 `FromContext()` 辅助函数
- [ ] 编写单元测试
- [ ] 编写 README.md（使用示例）
- [ ] 更新 Router 集成中间件

#### 技术细节
```go
// core/pkg/i18n/i18n.go

package i18n

import (
    "embed"
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "golang.org/x/text/language"
    "gopkg.in/yaml.v3"
)

//go:embed ../../locales/*.yaml
var localeFS embed.FS

var Bundle *i18n.Bundle

func Init() error {
    Bundle = i18n.NewBundle(language.English)
    Bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
    
    // 加载语言文件
    languages := []string{"en", "zh-CN"}
    for _, lang := range languages {
        _, err := Bundle.LoadMessageFileFS(localeFS, 
            fmt.Sprintf("locales/%s.yaml", lang))
        if err != nil {
            return err
        }
    }
    
    return nil
}

func FromContext(ctx context.Context) *i18n.Localizer {
    lang := ctx.Value("accept-language").(string)
    return i18n.NewLocalizer(Bundle, lang)
}
```

```go
// core/pkg/i18n/middleware.go

package i18n

import (
    "context"
    "net/http"
)

func Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 检测语言
        lang := detectLanguage(r)
        
        // 存入上下文
        ctx := context.WithValue(r.Context(), "accept-language", lang)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func detectLanguage(r *http.Request) string {
    // 1. URL 参数
    if lang := r.URL.Query().Get("lang"); lang != "" {
        if isSupportedLanguage(lang) {
            return lang
        }
    }
    
    // 2. Accept-Language Header
    if lang := parseAcceptLanguage(r.Header.Get("Accept-Language")); lang != "" {
        return lang
    }
    
    // 3. 默认英文
    return "en"
}
```

```yaml
# locales/en.yaml
user_not_found: "User not found"
invalid_email: "Invalid email format"
project_created: "Project created successfully"
welcome_user: "Welcome, {{.Name}}!"

# locales/zh-CN.yaml
user_not_found: "用户不存在"
invalid_email: "邮箱格式不正确"
project_created: "项目创建成功"
welcome_user: "欢迎，{{.Name}}！"
```

#### 测试用例
- 英文消息加载正确
- 中文消息加载正确
- 语言检测从 URL 参数
- 语言检测从 Accept-Language Header
- Fallback 到英文
- 变量替换正常工作
- 中间件正确设置上下文

---

## Story 8: 本地化（l10n）基础设施

**优先级**: P1  
**工期**: 2 天  
**依赖**: Story 7（i18n 基础设施）

### 目标
建立本地化基础设施，支持货币、日期、数字的区域化格式化，与 i18n 松耦合协作。

### 任务清单
- [ ] 创建 `core/pkg/localization` 包
  - [ ] `localization.go` - 主 Localizer
  - [ ] `currency.go` - 货币格式化
  - [ ] `datetime.go` - 日期时间格式化
  - [ ] `number.go` - 数字格式化
  - [ ] `units.go` - 度量单位转换
  - [ ] `config.go` - 配置加载
- [ ] 创建 `config/localization.yaml` 配置文件
- [ ] 创建中间件 `core/internal/middleware/localization.go`
- [ ] 编写单元测试（覆盖率 > 80%）
- [ ] 集成测试（验证 API 响应格式化）
- [ ] 更新 `docs/standards/localization-standards.md`（如需补充）

### 验收标准
1. **货币格式化**
   - 支持 USD、CNY、JPY、EUR、GBP
   - 正确显示货币符号位置（前缀 vs 后缀）
   - 千分位和小数点符合区域规则
   
2. **日期时间格式化**
   - 支持 5+ 种区域的日期格式
   - 支持 12/24 小时制切换
   - 时区转换正确
   
3. **数字格式化**
   - 千分位分隔符正确（逗号、点、空格）
   - 小数点符号正确
   - 百分比格式化
   
4. **度量单位转换**
   - 支持长度单位（米、千米、英里）
   - 支持重量单位（克、千克、磅）
   - 文件大小格式化（B、KB、MB、GB）
   
5. **架构要求**
   - 与 i18n 共享语言检测
   - 独立的 Localizer 上下文
   - 缓存机制（避免重复创建 Localizer）

### 代码示例

#### Localizer 主入口

```go
// core/pkg/localization/localization.go

package localization

import (
    "context"
    "time"
    "golang.org/x/text/language"
)

type Localizer struct {
    locale            string
    tag               language.Tag
    currencyFormatter *CurrencyFormatter
    dateTimeFormatter *DateTimeFormatter
    numberFormatter   *NumberFormatter
    unitConverter     *UnitConverter
}

func NewLocalizer(locale string) *Localizer {
    tag := language.MustParse(locale)
    
    return &Localizer{
        locale:            locale,
        tag:               tag,
        currencyFormatter: NewCurrencyFormatter(locale, getDefaultCurrency(locale)),
        dateTimeFormatter: NewDateTimeFormatter(locale),
        numberFormatter:   NewNumberFormatter(locale),
        unitConverter:     NewUnitConverter(locale),
    }
}

func FromContext(ctx context.Context) *Localizer {
    locale, ok := ctx.Value("locale").(string)
    if !ok {
        locale = "en-US"
    }
    
    return NewLocalizer(locale)
}

func (l *Localizer) FormatCurrency(amount float64, currency string) string {
    formatter := NewCurrencyFormatter(l.locale, currency)
    return formatter.FormatWithSymbol(amount)
}

func (l *Localizer) FormatDate(t time.Time) string {
    return l.dateTimeFormatter.FormatDate(t)
}

func (l *Localizer) FormatDateTime(t time.Time) string {
    return l.dateTimeFormatter.FormatDateTime(t)
}

func (l *Localizer) FormatNumber(n float64) string {
    return l.numberFormatter.FormatDecimal(n, 2)
}

func (l *Localizer) FormatBytes(bytes int64) string {
    return l.unitConverter.FormatBytes(bytes)
}
```

#### 货币格式化

```go
// core/pkg/localization/currency.go

package localization

import (
    "golang.org/x/text/currency"
    "golang.org/x/text/language"
    "golang.org/x/text/message"
)

type CurrencyFormatter struct {
    locale   language.Tag
    currency currency.Unit
    printer  *message.Printer
}

func NewCurrencyFormatter(locale, currencyCode string) *CurrencyFormatter {
    tag := language.MustParse(locale)
    curr := currency.MustParseISO(currencyCode)
    
    return &CurrencyFormatter{
        locale:   tag,
        currency: curr,
        printer:  message.NewPrinter(tag),
    }
}

func (f *CurrencyFormatter) FormatWithSymbol(amount float64) string {
    symbol := f.getCurrencySymbol()
    formatted := f.printer.Sprintf("%.2f", amount)
    
    if f.isSymbolPrefix() {
        return fmt.Sprintf("%s%s", symbol, formatted)
    }
    return fmt.Sprintf("%s %s", formatted, symbol)
}

func (f *CurrencyFormatter) getCurrencySymbol() string {
    symbols := map[string]string{
        "USD": "$",
        "CNY": "¥",
        "JPY": "¥",
        "EUR": "€",
        "GBP": "£",
    }
    return symbols[f.currency.String()]
}
```

#### 中间件

```go
// core/internal/middleware/localization.go

package middleware

import (
    "context"
    "net/http"
    "apprun/core/pkg/i18n"
)

func LocalizationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. 检测语言（复用 i18n 逻辑）
        lang := i18n.DetectLanguage(r)
        
        // 2. 映射到 Locale
        locale := mapLanguageToLocale(lang)
        
        // 3. 检查用户偏好（如果已登录）
        if user := getUserFromContext(r.Context()); user != nil {
            if user.PreferredLocale != "" {
                locale = user.PreferredLocale
            }
        }
        
        // 4. 存入上下文
        ctx := context.WithValue(r.Context(), "locale", locale)
        ctx = context.WithValue(ctx, "accept-language", lang) // i18n 使用
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func mapLanguageToLocale(lang string) string {
    localeMap := map[string]string{
        "en":    "en-US",
        "zh-CN": "zh-CN",
        "zh-TW": "zh-TW",
        "ja":    "ja-JP",
    }
    
    if locale, ok := localeMap[lang]; ok {
        return locale
    }
    
    return "en-US"
}
```

#### 配置文件

```yaml
# config/localization.yaml

localization:
  default_locale: en-US
  
  locales:
    en-US:
      currency: USD
      date_format: "01/02/2006"
      time_format: "3:04 PM"
      timezone: "America/New_York"
      
    zh-CN:
      currency: CNY
      date_format: "2006-01-02"
      time_format: "15:04"
      timezone: "Asia/Shanghai"
      
    ja-JP:
      currency: JPY
      date_format: "2006/01/02"
      time_format: "15:04"
      timezone: "Asia/Tokyo"
  
  currencies:
    USD:
      symbol: "$"
      decimal_places: 2
      symbol_prefix: true
      
    CNY:
      symbol: "¥"
      decimal_places: 2
      symbol_prefix: true
```

#### 使用示例

```go
// Handler 中使用

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
    // i18n: 消息翻译
    i18nLocalizer := i18n.FromContext(r.Context())
    message := i18nLocalizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "product_detail",
    })
    
    // l10n: 数据格式化
    l10nLocalizer := localization.FromContext(r.Context())
    
    product := h.getProduct(productID)
    
    response.Success(w, map[string]interface{}{
        "message":    message,                                      // i18n
        "name":       product.Name,
        "price":      l10nLocalizer.FormatCurrency(product.Price, "USD"), // l10n
        "created_at": l10nLocalizer.FormatDate(product.CreatedAt),        // l10n
        "size":       l10nLocalizer.FormatBytes(product.Size),            // l10n
    })
}
```

#### 测试用例
- 货币格式化（USD、CNY、JPY、EUR）
- 日期格式化（5+ 种区域）
- 日期时间格式化（12/24 小时制）
- 数字格式化（千分位、小数点）
- 文件大小格式化（B、KB、MB、GB）
- Locale 检测（URL 参数、用户偏好、语言映射）
- 与 i18n 协作（共享语言检测，独立上下文）
- 缓存机制（避免重复创建 Localizer）

---

## Sprint 依赖

### 外部依赖
- GitHub Actions (CI/CD)
- Go 1.21+
- golangci-lint
- go-i18n v2
- golang.org/x/text

### 工具依赖
- Ent ORM
- Testify (测试框架)
- httptest (HTTP 测试)
- go-i18n (国际化)
- golang.org/x/text (本地化)

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
   - `core/pkg/i18n` 包（含测试）
   - `core/pkg/localization` 包（含测试）
   - `pkg/testutil` 包（含示例）
   - 更新后的 Ent Schema
   - 更新后的 Handler 代码

2. **配置**
   - `.golangci.yml`
   - `.github/workflows/ci.yml`
   - `scripts/check-ent-json-tags.sh`
   - `config/localization.yaml`
   - 更新后的 Makefile

3. **国际化/本地化资源**
   - `locales/en.yaml` (英文消息)
   - `locales/zh-CN.yaml` (简体中文消息)
   - `locales/zh-TW.yaml` (繁体中文消息)
   - `locales/ja.yaml` (日文消息)

4. **文档**
   - `core/pkg/response/README.md`
   - `core/pkg/errors/README.md`
   - `core/pkg/i18n/README.md`
   - `core/pkg/localization/README.md`
   - `docs/standards/testing-guide.md`（可选）

---

## Sprint 回顾准备

### 需要讨论的问题
- 工具包 API 设计是否合理？
- CI/CD 流程是否满足需求？
- 测试框架是否易用？
- i18n/l10n 架构设计是否满足业务需求？
- 是否需要调整开发流程？

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26
