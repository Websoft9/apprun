# 部署架构文档
# apprun BaaS Platform

**创建日期**: 2025-12-25  
**架构师**: Winston (Architect Agent)  
**版本**: 1.0  
**状态**: Draft

---

## 1. 部署概览

### 1.1 MVP 部署架构

**目标**: 单机 1C2G，docker-compose 一键启动

```
┌────────────────────────────────────────────────────────┐
│                  Host Machine (4C8G)                   │
├────────────────────────────────────────────────────────┤
│                    Docker Network                      │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │
│  │   apprun     │  │ PostgreSQL   │  │   Redis     │ │
│  │   :8080      │──│   :5432      │  │   :6379     │ │
│  │  (Go Binary) │  │  (数据持久化) │  │ (事件+缓存)  │ │
│  └──────────────┘  └──────────────┘  └─────────────┘ │
│         │                                              │
│         ├──────────────┬──────────────┬──────────────┐│
│  ┌──────▼─────┐ ┌─────▼──────┐ ┌────▼──────┐ ┌─────▼┤
│  │Ory Kratos  │ │ Waterflow  │ │Prometheus │ │Grafana│
│  │  :4433     │ │   :7233    │ │  :9090    │ │ :3000 │ │
│  │ (认证服务)  │ │ (工作流)    │ │ (监控)     │ │(可视化)│
│  └────────────┘ └────────────┘ └───────────┘ └──────┘│
│                                                        │
│  Volumes:                                              │
│  - postgres-data  (数据库持久化)                        │
│  - redis-data     (Redis 持久化)                        │
│  - apprun-storage (本地文件存储)                        │
│  - apprun-logs    (日志文件)                            │
└────────────────────────────────────────────────────────┘
         │
         ▼
    HTTPS :443 (Nginx/Traefik 反向代理)
```

### 1.2 服务依赖关系

```
apprun 启动依赖顺序:
1. PostgreSQL    (必需 - 主数据库)
2. Ory Kratos    (必需 - 认证服务)
3. Redis         (可选 - 事件中心 + 缓存)
4. Waterflow     (可选 - 工作流引擎)
5. apprun        (核心服务)
6. Prometheus    (可选 - 监控)
7. Grafana       (可选 - 可视化)
```

---

## 2. Docker Compose 配置

### 2.1 核心服务配置

```yaml
version: '3.8'

services:
  # ==================== 核心服务 ====================
  apprun:
    image: websoft9/apprun:latest
    container_name: apprun
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://apprun:password@postgres:5432/apprun?sslmode=disable
      - REDIS_URL=redis://redis:6379/0
      - KRATOS_PUBLIC_URL=http://kratos:4433
      - KRATOS_ADMIN_URL=http://kratos:4434
      - WATERFLOW_URL=http://waterflow:7233
      - LOG_LEVEL=info
      - STORAGE_PATH=/data/storage
    volumes:
      - apprun-storage:/data/storage
      - apprun-logs:/var/log/apprun
      - ./config:/app/config:ro
    depends_on:
      postgres:
        condition: service_healthy
      kratos:
        condition: service_started
      redis:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - apprun-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ==================== 数据库 ====================
  postgres:
    image: postgres:14-alpine
    container_name: apprun-postgres
    environment:
      - POSTGRES_USER=apprun
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=apprun
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped
    networks:
      - apprun-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U apprun"]
      interval: 10s
      timeout: 5s
      retries: 5

  # ==================== Redis (可选) ====================
  redis:
    image: redis:7-alpine
    container_name: apprun-redis
    command: redis-server --appendonly yes --maxmemory 512mb --maxmemory-policy allkeys-lru
    volumes:
      - redis-data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped
    networks:
      - apprun-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # ==================== 认证服务 ====================
  kratos:
    image: oryd/kratos:latest
    container_name: apprun-kratos
    command: serve -c /etc/config/kratos/kratos.yml --dev --watch-courier
    environment:
      - DSN=postgresql://apprun:password@postgres:5432/apprun?sslmode=disable
    volumes:
      - ./config/kratos:/etc/config/kratos:ro
    ports:
      - "4433:4433"  # Public API
      - "4434:4434"  # Admin API
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - apprun-network

  # ==================== 工作流引擎 (可选) ====================
  waterflow:
    image: websoft9/waterflow:latest
    container_name: apprun-waterflow
    environment:
      - DATABASE_URL=postgresql://apprun:password@postgres:5432/apprun?sslmode=disable
    ports:
      - "7233:7233"  # Temporal Frontend
      - "7234:7234"  # Waterflow API
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - apprun-network

  # ==================== 监控 (可选) ====================
  prometheus:
    image: prom/prometheus:latest
    container_name: apprun-prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
    volumes:
      - ./config/prometheus:/etc/prometheus:ro
      - prometheus-data:/prometheus
    ports:
      - "9090:9090"
    restart: unless-stopped
    networks:
      - apprun-network

  grafana:
    image: grafana/grafana:latest
    container_name: apprun-grafana
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_INSTALL_PLUGINS=redis-datasource
    volumes:
      - grafana-data:/var/lib/grafana
      - ./config/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./config/grafana/datasources:/etc/grafana/provisioning/datasources:ro
    ports:
      - "3000:3000"
    depends_on:
      - prometheus
    restart: unless-stopped
    networks:
      - apprun-network

# ==================== 网络 ====================
networks:
  apprun-network:
    driver: bridge

# ==================== 持久化卷 ====================
volumes:
  postgres-data:
  redis-data:
  apprun-storage:
  apprun-logs:
  prometheus-data:
  grafana-data:
```

### 2.2 配置文件结构

```
config/
├── default.yaml              # apprun 主配置
├── conf.d/                   # 额外配置目录
│   └── custom-database.yaml
├── kratos/
│   ├── kratos.yml           # Kratos 配置
│   └── identity.schema.json # 用户身份 Schema
├── prometheus/
│   └── prometheus.yml       # Prometheus 配置
└── grafana/
    ├── dashboards/          # Grafana 仪表盘
    │   └── apprun.json
    └── datasources/         # 数据源配置
        └── prometheus.yml
```

---

## 3. 网络架构

### 3.1 内部网络

```
Docker Network: apprun-network (bridge)

服务通信:
- apprun        → postgres:5432
- apprun        → redis:6379
- apprun        → kratos:4433 (Public API)
- apprun        → kratos:4434 (Admin API)
- apprun        → waterflow:7234
- kratos        → postgres:5432 (共享数据库)
- waterflow     → postgres:5432
- prometheus    → apprun:8080/metrics
- grafana       → prometheus:9090
```

### 3.2 外部访问

```
公网访问流程:

Internet (HTTPS :443)
    │
    ▼
Nginx/Traefik (反向代理 + TLS 终止)
    │
    ├─→ apprun:8080       (主 API)
    ├─→ kratos:4433       (认证 API)
    ├─→ grafana:3000      (监控面板)
    └─→ waterflow:7234    (工作流 API - 可选)
```

**Nginx 配置示例**:
```nginx
# /etc/nginx/conf.d/apprun.conf
upstream apprun_backend {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate /etc/ssl/certs/apprun.crt;
    ssl_certificate_key /etc/ssl/private/apprun.key;
    ssl_protocols TLSv1.3;

    location / {
        proxy_pass http://apprun_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket 支持
    location /ws {
        proxy_pass http://apprun_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

---

## 4. 存储架构

### 4.1 持久化存储

| 卷名称 | 用途 | 容量建议 | 备份策略 |
|--------|------|----------|----------|
| `postgres-data` | PostgreSQL 数据 | 20GB+ | 每日全量 + WAL 归档 |
| `redis-data` | Redis AOF 持久化 | 2GB | 可选（事件可丢失） |
| `apprun-storage` | 文件存储 | 50GB+ | 每日增量 |
| `apprun-logs` | 日志文件 | 5GB | 轮转保留 7 天 |
| `prometheus-data` | 监控数据 | 10GB | 保留 30 天 |
| `grafana-data` | Grafana 配置 | 1GB | 每周备份 |

### 4.2 文件存储

```
本地存储目录结构:
/data/storage/
├── projects/
│   ├── project-1/
│   │   ├── files/
│   │   └── uploads/
│   └── project-2/
└── temp/               # 临时文件（定期清理）

切换到 S3 存储:
- 修改 apprun 环境变量: STORAGE_TYPE=s3
- 添加 S3 配置: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, S3_BUCKET
- 通过 afero.S3Fs 无缝切换
```

---

## 5. 安全配置

### 5.1 密码管理

```bash
# 生成强密码
openssl rand -base64 32

# 环境变量方式（推荐）
cat > .env <<EOF
POSTGRES_PASSWORD=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 32)
KRATOS_SECRET=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 32)
EOF

# Docker Compose 引用
docker-compose --env-file .env up -d
```

### 5.2 网络隔离

```yaml
# 生产环境：使用独立网络
networks:
  frontend:  # 对外服务
    driver: bridge
  backend:   # 内部服务
    driver: bridge
    internal: true

services:
  apprun:
    networks:
      - frontend
      - backend
  
  postgres:
    networks:
      - backend  # 仅内部访问

  redis:
    networks:
      - backend
```

### 5.3 防火墙规则

```bash
# 仅开放必要端口
ufw allow 22/tcp    # SSH
ufw allow 443/tcp   # HTTPS
ufw allow 80/tcp    # HTTP (重定向到 HTTPS)

# 拒绝直接访问内部服务
ufw deny 5432/tcp   # PostgreSQL
ufw deny 6379/tcp   # Redis
ufw deny 8080/tcp   # apprun (通过 Nginx 代理)

ufw enable
```

---

## 6. 启动与停止

### 6.1 启动流程

```bash
# 1. 克隆配置
git clone https://github.com/websoft9/apprun.git
cd apprun

# 2. 配置环境变量
cp .env.example .env
vim .env  # 修改密码和配置

# 3. 启动所有服务
docker-compose up -d

# 4. 查看日志
docker-compose logs -f apprun

# 5. 验证服务健康
docker-compose ps
curl http://localhost:8080/health

# 6. 初始化数据库
docker-compose exec apprun ./apprun migrate up

# 7. 创建管理员用户
docker-compose exec apprun ./apprun admin create \
  --email admin@example.com \
  --password Admin@123
```

### 6.2 停止与清理

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据卷（危险！）
docker-compose down -v

# 仅重启 apprun
docker-compose restart apprun

# 查看资源占用
docker stats
```

---

## 7. 扩展部署

### 7.1 多实例部署

```yaml
# docker-compose.scale.yml
services:
  apprun:
    deploy:
      replicas: 3
    environment:
      - REDIS_URL=redis://redis:6379/0  # 必需 Redis 做会话共享
    
  nginx:
    image: nginx:alpine
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    ports:
      - "443:443"
    depends_on:
      - apprun

# 启动 3 个 apprun 实例
docker-compose -f docker-compose.yml -f docker-compose.scale.yml up -d --scale apprun=3
```

**Nginx 负载均衡配置**:
```nginx
upstream apprun_cluster {
    least_conn;
    server apprun-1:8080;
    server apprun-2:8080;
    server apprun-3:8080;
}

server {
    listen 443 ssl http2;
    location / {
        proxy_pass http://apprun_cluster;
    }
}
```

### 7.2 外部数据库

```yaml
# docker-compose.external-db.yml
services:
  apprun:
    environment:
      # 使用外部 PostgreSQL (如 RDS)
      - DATABASE_URL=postgresql://user:pass@rds.aws.com:5432/apprun
      # 使用外部 Redis (如 ElastiCache)
      - REDIS_URL=redis://redis.aws.com:6379/0

  # 移除 postgres 和 redis 服务
  # postgres: (删除)
  # redis: (删除)
```

---

## 8. 备份与恢复

### 8.1 数据库备份

```bash
# 手动备份
docker-compose exec postgres pg_dump -U apprun apprun > backup-$(date +%Y%m%d).sql

# 定期备份脚本
cat > /usr/local/bin/backup-apprun.sh <<'EOF'
#!/bin/bash
BACKUP_DIR="/backups/apprun"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份 PostgreSQL
docker-compose exec -T postgres pg_dump -U apprun apprun | gzip > $BACKUP_DIR/db-$DATE.sql.gz

# 备份文件存储
tar czf $BACKUP_DIR/storage-$DATE.tar.gz -C /var/lib/docker/volumes/apprun-storage/_data .

# 删除 7 天前的备份
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
EOF

chmod +x /usr/local/bin/backup-apprun.sh

# 添加到 crontab
crontab -e
# 0 2 * * * /usr/local/bin/backup-apprun.sh
```

### 8.2 数据恢复

```bash
# 恢复数据库
gunzip < backup-20251225.sql.gz | docker-compose exec -T postgres psql -U apprun apprun

# 恢复文件存储
tar xzf storage-20251225.tar.gz -C /var/lib/docker/volumes/apprun-storage/_data

# 重启服务
docker-compose restart apprun
```

---

## 9. 监控与告警

### 9.1 健康检查

```bash
# 服务健康检查端点
curl http://localhost:8080/health
# Response: {"status":"ok","database":"connected","redis":"connected"}

# Prometheus 指标
curl http://localhost:8080/metrics
```

### 9.2 告警规则

```yaml
# config/prometheus/alert.rules.yml
groups:
  - name: apprun_alerts
    interval: 30s
    rules:
      - alert: ApprunDown
        expr: up{job="apprun"} == 0
        for: 1m
        annotations:
          summary: "apprun service is down"
        
      - alert: HighErrorRate
        expr: rate(http_errors_total[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate: {{ $value }}"
      
      - alert: DatabaseConnectionPoolExhausted
        expr: db_connections_used / db_connections_max > 0.9
        for: 2m
        annotations:
          summary: "Database connection pool almost full"
```

---

## 10. 故障排查

### 10.1 常见问题

**问题 1: apprun 无法连接数据库**
```bash
# 检查网络
docker-compose exec apprun ping postgres

# 检查数据库日志
docker-compose logs postgres

# 验证连接
docker-compose exec postgres psql -U apprun -d apprun -c "SELECT 1"
```

**问题 2: Kratos 认证失败**
```bash
# 检查 Kratos 健康
curl http://localhost:4433/health/ready

# 查看 Kratos 日志
docker-compose logs kratos

# 验证数据库表
docker-compose exec postgres psql -U apprun -d apprun -c "\dt kratos*"
```

**问题 3: Redis 内存不足**
```bash
# 检查 Redis 内存使用
docker-compose exec redis redis-cli INFO memory

# 调整 maxmemory 策略
docker-compose exec redis redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

### 10.2 日志查看

```bash
# 查看所有日志
docker-compose logs

# 实时跟踪 apprun 日志
docker-compose logs -f apprun

# 查看最近 100 行
docker-compose logs --tail=100 apprun

# 导出日志
docker-compose logs apprun > apprun-logs-$(date +%Y%m%d).log
```

---

## 11. 性能优化

### 11.1 资源限制

```yaml
services:
  apprun:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 1G

  postgres:
    deploy:
      resources:
        limits:
          cpus: '1.5'
          memory: 2G
```

### 11.2 数据库优化

```sql
-- PostgreSQL 性能配置
ALTER SYSTEM SET shared_buffers = '512MB';
ALTER SYSTEM SET effective_cache_size = '2GB';
ALTER SYSTEM SET maintenance_work_mem = '128MB';
ALTER SYSTEM SET max_connections = 200;

-- 重启生效
SELECT pg_reload_conf();
```

---

## 附录

### A. 系统要求

**最低配置 (开发)**:
- CPU: 2 核
- 内存: 4GB
- 磁盘: 20GB SSD
- 网络: 10Mbps

**推荐配置 (MVP)**:
- CPU: 4 核
- 内存: 8GB
- 磁盘: 100GB SSD
- 网络: 100Mbps

**生产配置**:
- CPU: 8 核+
- 内存: 16GB+
- 磁盘: 500GB SSD (RAID 1/10)
- 网络: 1Gbps

### B. 端口清单

| 服务 | 端口 | 协议 | 公开 | 用途 |
|------|------|------|------|------|
| apprun | 8080 | HTTP | ❌ | API 服务 |
| PostgreSQL | 5432 | TCP | ❌ | 数据库 |
| Redis | 6379 | TCP | ❌ | 缓存/事件 |
| Kratos | 4433 | HTTP | ✅ | 认证 API |
| Kratos Admin | 4434 | HTTP | ❌ | 管理 API |
| Waterflow | 7234 | HTTP | ❌ | 工作流 API |
| Prometheus | 9090 | HTTP | ❌ | 监控 |
| Grafana | 3000 | HTTP | ✅ | 监控面板 |
| Nginx | 443 | HTTPS | ✅ | 反向代理 |

### C. 环境变量清单

```bash
# apprun 核心配置
DATABASE_URL=postgresql://user:pass@host:port/db
REDIS_URL=redis://host:port/db
KRATOS_PUBLIC_URL=http://kratos:4433
LOG_LEVEL=info|debug|warn|error
STORAGE_TYPE=local|s3
STORAGE_PATH=/data/storage

# S3 配置 (如使用)
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx
AWS_REGION=us-east-1
S3_BUCKET=apprun-storage

# 安全配置
JWT_SECRET=xxx
ENCRYPTION_KEY=xxx

# 性能配置
DB_MAX_CONNECTIONS=100
REDIS_MAX_IDLE=10
```

---

**文档维护**: Winston (Architect Agent)  
**审核状态**: 待运维团队评审  
**下一步**: 创建数据架构文档 (data-architecture.md)
