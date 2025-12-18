# apprun POC 环境

这是 apprun 的技术验证 POC 环境，用于验证核心技术选型的可行性。

## 📋 验证目标

本 POC 旨在验证以下核心技术：

1. ✅ **Go + PostgREST 集成** - 数据 API 自动生成
2. ✅ **Ory Kratos + Casbin** - 认证与授权方案
3. ✅ **WASM 函数运行时** - 轻量级函数执行

**已验证/不包含：**
- ~~Temporal 工作流~~：已通过独立POC验证
- ~~MinIO 对象存储~~：已转为闭源商业产品，后续考虑 SeaweedFS 等开源替代方案

## 🚀 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 8GB 可用内存
- 磁盘空间 20GB

### 一键启动

```bash
# 进入 POC 目录
cd poc

# 启动所有服务
./start-poc.sh
```

启动成功后，你将看到所有服务的访问地址。

## 🔍 服务列表

### 1. PostgreSQL (数据库)
- **端口**: 5432
- **数据库**: apprun_poc
- **用户名**: apprun
- **密码**: apprun123
- **测试连接**:
  ```bash
  psql -h localhost -U apprun -d apprun_poc
  ```

### 2. PostgREST (数据 API)
- **端口**: 3000
- **API 文档**: http://localhost:3000
- **测试查询**:
  ```bash
  # 获取所有产品
  curl http://localhost:3000/products
  
  # 筛选查询
  curl "http://localhost:3000/products?name=eq.Product%20A"
  
  # 创建产品（需要认证）
  curl -X POST http://localhost:3000/products \
    -H "Content-Type: application/json" \
    -d '{"tenant_id": "11111111-1111-1111-1111-111111111111", "name": "New Product", "price": 49.99}'
  ```

### 3. Ory Kratos (认证服务)
- **Public API**: http://localhost:4433
- **Admin API**: http://localhost:4434
- **健康检查**:
  ```bash
  curl http://localhost:4433/health/alive
  curl http://localhost:4434/health/ready
  ```



## 📚 测试数据

POC 环境已预置测试数据：

### 租户
```sql
SELECT * FROM tenants;
```
| ID | Name | Plan |
|----|------|------|
| 11111111-1111-1111-1111-111111111111 | Test Tenant 1 | free |
| 22222222-2222-2222-2222-222222222222 | Test Tenant 2 | pro |

### 用户
```sql
SELECT * FROM users;
```
| Email | Tenant | Role |
|-------|--------|------|
| alice@test.com | Test Tenant 1 | admin |
| bob@test.com | Test Tenant 1 | user |
| charlie@test.com | Test Tenant 2 | admin |

### 产品
```sql
SELECT * FROM products;
```
| Name | Tenant | Price |
|------|--------|-------|
| Product A | Test Tenant 1 | $19.99 |
| Product B | Test Tenant 1 | $29.99 |
| Product C | Test Tenant 2 | $39.99 |

## 🧪 验证场景

### 场景 1: PostgREST API 测试

```bash
# 1. 查询所有产品
curl http://localhost:3000/products

# 2. 分页查询
curl "http://localhost:3000/products?limit=10&offset=0"

# 3. 排序
curl "http://localhost:3000/products?order=price.desc"

# 4. 筛选
curl "http://localhost:3000/products?price=gt.20"

# 5. 关联查询（需要设置外键）
curl "http://localhost:3000/products?select=*,tenants(*)"
```

### 场景 2: 多租户隔离测试

```sql
-- 设置当前租户（模拟 JWT claims）
SET request.jwt.claims = '{"tenant_id": "11111111-1111-1111-1111-111111111111"}';

-- 查询产品（应该只返回该租户的数据）
SELECT * FROM products;
```

### 场景 3: Ory Kratos 注册/登录

```bash
# 1. 初始化注册流程
curl -X GET http://localhost:4433/self-service/registration/api

# 2. 提交注册（需要从上一步获取 flow ID）
curl -X POST http://localhost:4433/self-service/registration?flow=<flow_id> \
  -H "Content-Type: application/json" \
  -d '{
    "traits": {
      "email": "newuser@test.com",
      "name": "New User",
      "tenant_id": "11111111-1111-1111-1111-111111111111",
      "role": "user"
    },
    "password": "secure-password-123"
  }'

# 3. 登录流程类似
curl -X GET http://localhost:4433/self-service/login/api
```



## 🛠️ 开发指南

### 连接 PostgreSQL

```bash
# CLI 连接
psql -h localhost -p 5432 -U apprun -d apprun_poc

# Go 连接字符串
postgres://apprun:apprun123@localhost:5432/apprun_poc?sslmode=disable
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务
docker-compose logs -f postgres
docker-compose logs -f postgrest
docker-compose logs -f kratos
docker-compose logs -f temporal
docker-compose logs -f minio
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart postgrest
```

### 停止和清理

```bash
# 停止服务（保留数据）
docker-compose stop

# 停止并删除容器（保留数据卷）
docker-compose down

# 完全清理（包括数据）
docker-compose down -v
```

## 📊 监控与调试

### 资源监控

```bash
# 查看容器资源占用
docker stats

# 预期结果（约 400MB 总内存）
# CONTAINER              CPU %   MEM USAGE / LIMIT
# apprun-poc-postgres    ~2%     256MB / 8GB
# apprun-poc-postgrest   ~0.5%   50MB / 8GB
# apprun-poc-kratos      ~1%     50MB / 8GB
# apprun-core (待开发)   ~1%     100MB / 8GB (目标)
```

### 数据库调试

```sql
-- 查看租户统计
SELECT * FROM tenant_stats;

-- 查看工作流执行统计
SELECT * FROM workflow_execution_stats;

-- 查看审计日志
SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 10;

-- 查看活跃连接
SELECT * FROM pg_stat_activity;
```

## 🔐 安全注意

⚠️ **警告**：这是 POC 环境，使用默认密码和不安全配置。

**生产环境必须修改：**
- PostgreSQL 密码
- Kratos secrets (cookie & cipher)
- JWT secret
- MinIO credentials

## 📝 验证清单（3天计划）

### Day 1: PostgREST 集成
- [ ] PostgREST API 可访问
- [ ] CRUD 操作正常
- [ ] 多租户隔离有效
- [ ] 性能满足要求 (< 100ms)

### Day 2: Kratos + Casbin + WASM
- [ ] 用户注册/登录成功
- [ ] JWT 生成和验证
- [ ] Casbin 权限检查
- [ ] 性能满足要求 (> 1000 auth/s)
- [ ] WASM 模块加载成功
- [ ] 函数执行正常
- [ ] 启动时间 < 100ms

### Day 3: 集成测试
- [ ] 端到端流程通过
- [ ] 性能基准达标
- [ ] 资源占用 < 512MB (核心服务)
- [ ] 生成 POC 报告

## 📖 参考文档

- [PostgREST 文档](https://postgrest.org/en/stable/)
- [Ory Kratos 文档](https://www.ory.sh/kratos/docs/)
- [Temporal 文档](https://docs.temporal.io/)
- [MinIO 文档](https://min.io/docs/minio/linux/index.html)

## 🤝 获取帮助

如遇问题：
1. 查看日志: `docker-compose logs -f`
2. 检查服务状态: `docker-compose ps`
3. 重启服务: `docker-compose restart`
4. 联系架构师团队

---

**文档维护**
- 创建者: Root
- 最后更新: 2025-12-18
- 版本: 1.0
