# Architecture Design Questions Checklist
# apprun BaaS Platform

**创建日期**: 2025-12-25  
**架构师**: Winston (Architect Agent)  
**目标**: 在开始架构设计前，明确所有关键技术决策点

---

## 📋 问题清单使用说明

本问题清单基于 PRD 定义的 13 个核心模块，旨在帮助架构师在设计阶段明确所有关键决策点。请逐一回答每个问题，您的答案将指导后续的架构设计工作。

**符号说明**:
- 🔴 **必答题** - 必须在架构设计前明确
- 🟡 **重要题** - 显著影响架构设计
- 🟢 **可选题** - 可在实施阶段决策

---

## 1️⃣ 总体架构决策

### 1.1 系统架构风格 🔴

**Q1.1.1**: 您希望采用什么样的整体架构风格？
- [ ] 单体架构 (Monolithic) - 所有模块在一个进程
- [ ] 微服务架构 (Microservices) - 每个模块独立服务
- [x] 模块化单体 (Modular Monolith) - 单进程但模块清晰分离
- [ ] 混合架构 (Hybrid) - 核心单体 + 部分微服务

**理由**: MVP 阶段优先轻量级部署（单机 1C2G），模块间清晰分离便于后续演进为微服务，降低运维复杂度

**Q1.1.2**: 对于 MVP 阶段，您期望的部署复杂度是？
- [x] 简单（单机 docker-compose 一键启动）
- [ ] 中等（多容器编排，但无 Kubernetes）
- [ ] 复杂（完整云原生，Kubernetes + Helm）

**约束考虑**: PRD NFR-007 要求单机 4C8G 可运行核心功能

---

### 1.2 技术栈基础 🔴

**Q1.2.1**: 您在以下技术领域的偏好和经验如何？

**后端语言**:
- [x] Go (已使用) - 当前项目已有 Go 代码
- [ ] Rust - 性能优先
- [ ] Node.js/TypeScript - JavaScript 生态
- [ ] Java/Kotlin - 企业级生态

**前端框架** (如需管理界面):
- [ ] React
- [ ] Vue.js
- [ ] Svelte
- [ ] Angular
- [x] 无需前端（纯 API）

**Q1.2.2**: 您的团队对以下技术的熟悉程度？(1-5分，5最熟悉)
- Golang: 5 分
- Docker/容器化: 5 分
- 分布式系统: 5 分
- 云原生技术: 5 分
- DevOps/CI/CD: 5 分

---

### 1.3 基础设施偏好 🔴

**Q1.3.1**: 您计划在什么环境运行 apprun？
- [ ] 本地开发环境（笔记本/台式机）
- [ ] 私有数据中心（自建机房）
- [ ] 公有云（AWS/阿里云/腾讯云等）
- [x] 混合云（私有云 + 公有云）

**Q1.3.2**: 您是否有特定的云服务商偏好或限制？
- [x] 无限制，可选任意云服务商
- [ ] 必须使用: _____________ (指定云服务商)
- [ ] 必须避免: _____________ (指定云服务商)

**Q1.3.3**: 您是否希望最大化利用云服务商的托管服务？
- [ ] 是 - 优先使用 RDS、Redis、对象存储等托管服务
- [ ] 否 - 优先使用自部署开源方案（可移植性优先）
- [x] 混合 - 核心自建，周边托管

---

## 2️⃣ 数据层架构

### 2.1 数据库选型 🔴

**Q2.1.1**: 您对以下数据库类型的偏好？

**关系型数据库**:
- [x] PostgreSQL (推荐 - 功能最全)
- [ ] MySQL/MariaDB
- [ ] SQLite (仅开发/小规模)
- [ ] 其他: _____________

**理由**: 功能最全（JSON、全文搜索、地理位置、数组）、与 Ent ORM 完美集成、ACID 事务、社区活跃、可扩展性好

**Q2.1.2**: 是否需要 NoSQL 数据库？
- [ ] 不需要（关系型足够）
- [ ] 需要文档数据库（MongoDB、CouchDB）
- [x] 需要键值数据库（Redis、Etcd）
- [ ] 需要图数据库（Neo4j、ArangoDB）

**用途**: 会话存储、配置缓存、分布式锁、队列等
**注**: Redis 是对于事件中心来说是必要组件，对于缓存来说是备选（L2级）

**Q2.1.3**: 您希望如何管理数据库 Schema？
- [ ] ORM 自动迁移（如 Ent、GORM）
- [ ] 迁移工具（如 golang-migrate、Flyway）
- [ ] 手动 SQL 脚本
- [x] 声明式 Schema 管理（如 Atlas）

**现状**: 项目已使用 Ent (ent/schema/)
**理由**: Atlas 是 Ent 官方推荐的 Schema 管理工具，支持声明式迁移、版本控制、Dry-run 预览、自动生成迁移脚本、团队协作友好

---

### 2.2 缓存策略 🟡

**Q2.2.1**: 您计划如何实现缓存？
- [ ] 进程内缓存（如 Go sync.Map、bigcache）
- [ ] 分布式缓存（Redis、Memcached）
- [x] 多层缓存（本地 + 分布式），使用 Write-through 模式
- [ ] 不需要缓存（MVP 阶段）

**理由**: L1 缓存（进程内）用于热点数据和配置，极低延迟；L2 缓存（Redis）用于会话存储和跨实例共享；无 Redis 时自动降级到 L1；性能最优且灵活

**Q2.2.2**: 缓存的主要用途是？(可多选)
- [x] API 响应缓存
- [x] 会话存储（Session）
- [x] 配置缓存
- [x] 热点数据缓存
- [ ] 分布式锁

**注**: apprun 目前为单机部署，使用进程内锁即可
---

### 2.3 数据建模 (FR-DATA-001) 🔴

**Q2.3.1**: 您希望如何实现 DSL 定义的数据模型？

**方案A - 配置化 ORM (推荐)**:
- 用户通过 YAML/JSON 定义模型
- 系统解析配置并生成 Ent Schema
- 运行 `go generate` 生成 CRUD 代码
- 优点: 类型安全、性能好
- 缺点: 需编译部署

**方案B - 动态 Schema**:
- 用户通过 API 定义模型
- 使用 JSON Schema 验证
- 动态生成数据库表（ALTER TABLE）
- 优点: 无需重启
- 缺点: 性能较低、类型安全弱

**方案C - 混合方案**:
- 核心模型静态（Ent）
- 用户模型动态（JSON + 通用表）

**您的选择**: 方案A 配置化 ORM
**理由**: apprun 是面向开发者的平台，遵循 code as service 的原则

**Q2.3.2**: 您对自动生成 RESTful API 的期望？
- [ ] 完全自动生成（CRUD + 过滤 + 分页）
- [x] 生成基础 API，手动扩展
- [ ] 手动实现所有 API

**现状**: 项目已使用 Ent  
**注**: 后期要考虑业务校验和业务逻辑处理与 ent，handler 层的分工

---

## 3️⃣ 认证与权限 (FR-AUTH-001)

### 3.1 认证服务集成 🔴

**Q3.1.1**: 您希望如何实现认证服务？

**方案A - 集成 Ory Kratos (推荐)**:
- 优点: 生产级、功能完整、已有 POC
- 缺点: 额外依赖
- 现状: poc/kratos/ 已有配置

**方案B - 自研认证服务**:
- 优点: 完全可控
- 缺点: 安全风险高、开发周期长

**方案C - 第三方 SaaS**:
- Auth0、Clerk、Supabase Auth
- 优点: 快速集成
- 缺点: 供应商锁定、成本

**您的选择**: 方案A -集成 Ory Kratos
**理由**: 认证是比较复杂的模块，自研周期太长

**Q3.1.2**: 如果集成 Ory Kratos，您希望如何集成？
- [ ] 独立服务（通过 API 调用）
- [x] 独立运行+共享数据库（仅读取 Kratos 数据表）
- [ ] 事件订阅（监听 Kratos Webhooks）

---

### 3.2 授权模型 🔴

**Q3.2.1**: 您希望实现什么级别的权限控制？

- [ ] **基础 RBAC** - 用户 → 角色 → 权限
- [x] **Project-based RBAC** - 用户在不同 Project 有不同角色
- [ ] **ABAC (属性)** - 基于资源属性动态授权
- [ ] **ReBAC (关系)** - 基于资源关系图（如 Google Zanzibar）

**PRD 要求**: Project-based RBAC（用户可加入多个 Project）+ 平台级全局权限

**Q3.2.2**: 您希望如何实现权限检查？
- [ ] 代码中硬编码（if user.HasRole("admin")）
- [ ] 中间件拦截（基于路由规则）
- [x] 策略引擎（如 Casbin、OPA）
- [ ] 混合方案

**Q3.2.3**: 是否需要细粒度权限（行级安全）？
- [x] 不需要（资源级权限足够）
- [ ] 需要（如用户只能看到自己的数据）

**示例**: 用户 A 在 Project X 是 Admin，但在 Project Y 是 Viewer

---

## 4️⃣ API 网关 (FR-GATEWAY-001)

### 4.1 网关选型 🔴

**Q4.1.1**: 您希望使用什么样的 API 网关？

**方案A - 自研简化网关 (Go)**:
- 基于 Chi Router + httputil.ReverseProxy + middleware
- 优点: 轻量、可控
- 缺点: 功能有限

**方案B - 开源网关**:
- Kong (Lua/插件丰富)
- Traefik (云原生/自动化)
- Nginx + OpenResty (高性能)
- Envoy (Istio 核心)
- 其他: _____________

**方案C - 云服务网关**:
- AWS API Gateway
- 阿里云 API Gateway
- 腾讯云 API Gateway

**您的选择**: 方案A - 自研简化网关 (Go)
**理由**: 路由本身是核心模块，在它的基础上开发更快

**Q4.1.2**: PRD 明确 MVP 不包含的高级功能，您是否同意？
- [x] ✅ 同意 - 限流、熔断、服务发现 MVP 不实现
- [ ] ❌ 不同意 - 我需要: _____________

---

### 4.2 路由设计 🟡

**Q4.2.1**: 您希望使用什么样的 API 风格？
- [ ] RESTful (资源导向)
- [ ] GraphQL (查询语言)
- [ ] gRPC (高性能 RPC)
- [x] 混合（不同模块不同风格）

**Q4.2.2**: 您希望如何组织 API 路由？

**方案A - 按模块分组**:
```
/api/v1/auth/*
/api/v1/data/*
/api/v1/storage/*
```

**方案B - 按资源类型**:
```
/api/v1/users/*
/api/v1/projects/*
/api/v1/models/*
```

**方案C - 混合**:

核心资源用方案B，功能模块用方案A

**您的选择**:  方案C - 混合。 例如
```
模块分组：/api/v1/auth/*, /api/v1/storage/*
资源分组：/api/v1/users/*, /api/v1/projects/*
```

---

## 5️⃣ 配置中心 (FR-CONFIG-001)

### 5.1 配置管理 🟡

**Q5.1.1**: 您希望如何存储配置？

**方案A - 文件 + 数据库混合 (推荐)**:
- 系统配置: YAML 文件（config/default.yaml）
- 动态配置: 数据库（configitem 表）
- 现状: 项目已有此实现

**方案B - 配置中心服务**:
- Consul、Etcd、Nacos
- 优点: 分布式、Watch 机制
- 缺点: 额外依赖

**方案C - 云服务**:
- AWS Parameter Store、阿里云 ACM

**您的选择**: 方案A - 文件 + 数据库混合

**Q5.1.2**: 您希望配置如何动态更新？
- [x] 进程内 Watch（定期轮询）
- [ ] 事件驱动（配置变更推送事件）
- [ ] 重启生效（MVP 可接受）

**现状**: 使用 Viper 已实现基础配置管理

---

## 6️⃣ 文件存储 (FR-STORAGE-001)

### 6.1 存储方案 🔴

**Q6.1.1**: 您希望使用什么样的文件存储？

**方案A - 对象存储 (推荐)**:
- MinIO (自部署、S3 兼容)
- AWS S3
- 阿里云 OSS
- 腾讯云 COS
- 优点: 可扩展、CDN 友好
- 缺点: 无原生文件夹概念

**方案B - 本地文件系统**:
- 直接存储到服务器磁盘
- 优点: 简单
- 缺点: 不支持水平扩展

**方案C - 网络文件系统**:
- NFS、GlusterFS、Ceph
- 优点: 保留文件夹语义
- 缺点: 性能和复杂度

**您的选择**: 混合方案：本地文件系统（优先）+ S3支持（接口抽象）
**理由**: 类似于一个“虚拟文件系统”（VFS），用户无需关心底层存储类型。虚拟文件系统包有 github.com/spf13/afero

**Q6.1.2**: 您希望如何模拟文件夹结构？(如选择对象存储)
- [ ] 虚拟文件夹（在 Key 中用 `/` 分隔）
- [ ] 元数据表（数据库记录文件夹关系）
- [x] 两者结合（路径 + 元数据）

**Q6.1.3**: 您对文件访问权限的要求？
- [ ] 公开访问（生成公共 URL）
- [x] 私有访问（临时签名 URL）
- [x] 基于用户权限（继承 RBAC）

---

### 6.2 文件处理 🟡

**Q6.2.1**: 您是否需要以下文件处理能力？(可多选)
- [ ] 图片压缩/缩放
- [ ] 文档预览（PDF、Office）
- [ ] 视频转码
- [ ] 病毒扫描
- [ ] 内容审核（AI 检测）
- [x] 全部都不需要

**注**: PRD 明确 MVP 不包含高级图片/视频处理

---

## 7️⃣ 函数服务 (FR-FUNC-001)

### 7.1 函数执行方案 🔴

**Q7.1.1**: 您希望如何实现函数执行隔离？

**方案A - 进程隔离**:
- 每个函数启动独立进程
- 优点: 简单
- 缺点: 启动慢、资源占用高

**方案B - 容器隔离**:
- 每个函数一个 Docker 容器
- 优点: 安全、资源可限制
- 缺点: 需 Docker API、冷启动慢

**方案C - WebAssembly (WASM)**:
- 函数编译为 WASM 模块
- 优点: 快速启动、安全沙箱
- 缺点: 生态较新、语言限制

**方案D - Plugin 机制**:
- 基于 Go Plugin 或 gRPC
- 优点: 高性能
- 缺点: 灵活性有限

**您的选择**: 方案A - 进程隔离
**理由**: 快速开发

**Q7.1.2**: 您希望支持哪些编程语言？(PRD 要求 Golang)
- [x] 仅 Go
- [ ] Go + JavaScript/TypeScript (Node.js)
- [ ] Go + Python
- [ ] 多语言（WASM 支持）

**Q7.1.3**: 您希望函数如何触发？(可多选)
- [x] HTTP 请求（Webhook）
- [x] 事件驱动（订阅事件总线）
- [ ] 定时任务（Cron）
- [ ] 数据库触发器（CDC）

---

## 8️⃣ 插件扩展 (FR-PLUG-001)

### 8.1 插件架构 🟡

**Q8.1.1**: 您希望使用什么插件协议？

**方案A - Go Plugin**:
- 优点: 原生、高性能
- 缺点: 版本兼容性差、仅 Linux/macOS

**方案B - gRPC Plugin (推荐)**:
- 参考 HashiCorp go-plugin
- 优点: 跨语言、进程隔离
- 缺点: 略复杂

**方案C - WebAssembly**:
- 优点: 安全、跨平台
- 缺点: 生态较新

**您的选择**: 方案B - gRPC Plugin  
**注**: 插件用于扩展系统能力（如新存储后端），函数服务用于用户自定义业务逻辑  

**Q8.1.2**: 您希望插件扩展哪些能力？(可多选)
- [x] 认证方式（LDAP、OAuth）
- [x] 存储后端（新存储类型）
- [x] API 拦截器（中间件）
- [x] 工作流节点（自定义节点）
- [x] 数据库类型（新 DB 适配器）

---

## 9️⃣ 工作流服务 (FR-WORKFLOW-001)

> 工作流的架构和开发已经在 https://github.com/websoft9/waterflow 单独处理。

### 9.1 工作流引擎选型 🔴

**Q9.1.1**: 您希望使用什么工作流引擎？

**方案A - Temporal (推荐)**:
- 优点: 生产级、容错性强、社区活跃
- 缺点: 架构复杂、资源占用高
- 现状: 项目已有 pkg/temporal/

**方案B - Cadence**:
- Temporal 的前身
- 优点: 成熟稳定
- 缺点: 社区活跃度低于 Temporal

**方案C - Airflow**:
- 优点: 生态丰富、适合数据管道
- 缺点: Python 生态、重量级

**方案D - 自研轻量引擎**:
- 优点: 轻量、可控
- 缺点: 功能有限、可靠性待验证

**方案E - Waterflow (Websoft9)**:
- 优点: 自家产品、可深度定制
- 缺点: 社区小
- 参考: docs/prd.md 提到此项目

**您的选择**: Waterflow (Websoft9) 
**理由**: Waterflow 是一个基于 Temporal 的平台。  
**注**: Waterflow 作为独立服务运行，apprun 通过 API 调用/事件驱动触发工作流"  

**Q9.1.2**: 您希望如何定义工作流？
- [x] YAML/JSON 配置
- [ ] 代码定义（DSL）
- [ ] 可视化编辑器（Drag & Drop）
- [ ] 混合（配置 + 代码）

**注**: PRD 明确 MVP 不包含可视化编排器

---

### 9.2 工作流触发 🟡

**Q9.2.1**: 您希望支持哪些触发方式？(可多选)
- [x] 手动触发（API 调用）
- [x] 定时触发（Cron）
- [x] 事件触发（订阅事件总线）
- [x] Webhook 触发（外部系统）
- [x] 数据库触发（CDC）

---

## 🔟 事件中心 (FR-EVENT-001)

### 10.1 消息中间件选型 🔴

**Q10.1.1**: 您希望使用什么消息队列/事件总线？

**方案A - NATS/NATS Streaming**:
- 优点: 轻量、云原生、Go 原生
- 缺点: 功能相对简单

**方案B - Kafka**:
- 优点: 高吞吐、持久化、生态丰富
- 缺点: 重量级、Java 生态

**方案C - RabbitMQ**:
- 优点: 功能丰富、AMQP 标准
- 缺点: Erlang 生态、性能一般

**方案D - Redis Streams**:
- 优点: 轻量、已有 Redis 可复用
- 缺点: 功能有限

**方案E - 云服务**:
- AWS SQS/SNS、阿里云 MNS

**您的选择**: 方案D - Redis Streams
**理由**: 项目已有 Redis 作为可选依赖，无需额外部署新服务，降低复杂度

**Q10.1.2**: 您对事件持久化的要求？
- [ ] 内存即可（重启丢失）
- [x] 短期持久化（1-7天）
- [ ] 长期持久化（需事件溯源）

**Q10.1.3**: 您对事件顺序的要求？
- [ ] 不保证顺序（性能优先）
- [x] 分区内有序（如 Kafka）
- [ ] 全局严格有序

---

## 1️⃣1️⃣ 实时推送 (FR-REALTIME-001)

### 11.1 实时通信方案 🟡

**Q11.1.1**: 您希望如何实现实时推送？

**方案A - WebSocket (原生)**:
- 自己实现 WebSocket 服务器
- 优点: 完全可控
- 缺点: 需处理连接管理、分布式问题

**方案B - WebSocket 框架**:
- nhooyr.io/websocket、Gorilla WebSocket、ws
- 优点: 开箱即用
- 缺点: 框架依赖

**方案C - Server-Sent Events (SSE)**:
- 优点: 简单、单向推送足够
- 缺点: 仅服务端到客户端

**方案D - 云服务**:
- AWS AppSync、阿里云 MQTT

**您的选择**: WebSocket 框架 - coder/websocket（原为 nhooyr.io/websocket）

**Q11.1.2**: 您希望如何实现 Database CDC（变更捕获）？
- [ ] 数据库触发器 + Webhook
- [ ] 轮询数据库（定期查询）
- [ ] CDC 工具（Debezium、Maxwell）
- [x] 应用层拦截（ORM Hooks）

**Q11.1.3**: 您是否需要支持多客户端同步？
- [x] 不需要（单向推送即可）
- [ ] 需要（协作编辑场景）

---

## 1️⃣2️⃣ 国际化 (FR-I18N-001)

### 12.1 国际化方案 🟢

**Q12.1.1**: 您希望如何存储翻译数据？
- [ ] JSON/YAML 文件
- [x] 数据库（Key-Value 表）
- [ ] 专业 i18n 服务（Crowdin、Lokalise）

**Q12.1.2**: 您希望支持哪些语言？(可多选)
- [x] 中文（简体/繁体）
- [x] 英文
- [ ] 日文
- [ ] 韩文
- [ ] 其他: _____________

**Q12.1.3**: 您希望如何检测用户语言？(优先级排序)
1. 用户设置
2. Cookie
3. HTTP Header
4. IP 地理位置

选项: HTTP Header (Accept-Language)、Cookie、用户设置、IP 地理位置

---

## 1️⃣3️⃣ 日志与监控 (FR-LOG-001)

### 13.1 日志方案 🔴

**Q13.1.1**: 您希望使用什么日志架构？

**方案A - ELK Stack**:
- Elasticsearch + Logstash + Kibana
- 优点: 功能强大、生态丰富
- 缺点: 资源占用高

**方案B - Loki + Grafana**:
- 优点: 轻量、云原生
- 缺点: 查询能力弱于 ES

**方案C - 云服务**:
- AWS CloudWatch、阿里云 SLS

**方案D - 简化方案 (MVP)**:
- 日志写文件 + 简单查询 API
- 优点: 极简
- 缺点: 功能有限

**您的选择**: 方案D - 简化方案 (MVP)

**Q13.1.2**: 您希望如何实现分布式追踪（Trace ID）？
- [ ] 自定义 Trace ID（UUID）
- [ ] OpenTelemetry 标准
- [ ] 云服务（X-Ray、SkyWalking）
- [x] MVP 不实现

**注**: PRD 明确 MVP 不包含 Distributed Tracing

---

### 13.2 监控方案 🟡

**Q13.2.1**: 您希望使用什么监控系统？

**方案A - Prometheus + Grafana (推荐)**:
- 优点: 云原生标准、社区丰富
- 缺点: 需额外部署

**方案B - 云服务监控**:
- AWS CloudWatch、阿里云 ARMS

**方案C - 简化方案 (MVP)**:
- 基础指标 API（/metrics 端点）
- 优点: 轻量
- 缺点: 无可视化

**您的选择**: 方案A - Prometheus + Grafana

**Q13.2.2**: 您需要监控哪些指标？(可多选)
- [x] Linux 和 容器的系统资源（CPU、内存、磁盘）
- [x] API 性能（响应时间、QPS）
- [x] 业务指标（用户数、请求数）
- [x] 错误率和告警
- [x] 数据库性能

**Q13.2.3**: 您希望如何实现告警？
- [x] 邮件通知
- [x] Webhook（钉钉、Slack）
- [ ] 短信通知
- [ ] MVP 不实现

**注**: 通知使用监控工具自带功能，不做开发

---

## 1️⃣4️⃣ License 管理 (FR-LICENSE-001)

### 14.1 License 方案 🟢

**Q14.1.1**: 您希望实现什么级别的 License 控制？
- [x] 简单开关（功能启用/禁用）
- [ ] 用量限制（用户数、API 调用数）
- [x] 时间限制（试用期、订阅到期）
- [x] 离线验证（本地 License 文件）
- [ ] 在线验证（服务器验证）

**Q14.1.2**: 您希望如何加密 License？
- [ ] 简单签名（HMAC）
- [x] 非对称加密（RSA/ECDSA）
- [ ] 商业 License 方案（如 Cryptlex）

---

## 1️⃣5️⃣ 开发工具链

### 15.1 CI/CD 🟡

**Q15.1.1**: 您计划使用什么 CI/CD 工具？
- [x] GitHub Actions (推荐)
- [ ] GitLab CI/CD
- [ ] Jenkins
- [ ] 云服务（CodePipeline 等）
- [ ] 其他: _____________

**现状**: 项目在 GitHub (Websoft9/apprun)

**Q15.1.2**: 您希望 CI/CD 包含哪些阶段？(可多选)
- [x] 代码检查（Linter）
- [x] 单元测试
- [x] 集成测试
- [x] 镜像构建和推送
- [x] 自动部署（Dev/Staging 环境）

---

### 15.2 测试策略 🟡

**Q15.2.1**: 您对测试覆盖率的目标？
- 单元测试: 70% (PRD 要求 >70%)
- 集成测试: 40%
- E2E 测试: 20%

**现状**: tests/ 目录已有测试框架

**Q15.2.2**: 您希望使用什么测试工具？
- [ ] Go 原生 testing
- [x] Testify (断言库)
- [ ] Ginkgo/Gomega (BDD)
- [ ] 其他: _____________

---

### 15.3 代码质量 🟡

**Q15.3.1**: 您希望使用哪些代码质量工具？(可多选)
- [x] golangci-lint (多 Linter 集成)
- [x] gofmt/goimports (格式化)
- [x] go vet (静态分析)
- [x] staticcheck
- [ ] SonarQube (代码质量平台)

---

## 1️⃣6️⃣ 安全性

### 16.1 数据安全 🔴

**Q16.1.1**: 您对以下数据的加密要求？

- **传输层**:
  - [x] HTTPS/TLS 1.3
  - [ ] 双向 TLS (mTLS)
  
- **存储层**:
  - [x] 密码加密（bcrypt）
  - [x] 敏感数据加密（AES）
  - [ ] 数据库透明加密（TDE）

**Q16.1.2**: 您希望如何管理加密密钥？
- [x] 环境变量
- [ ] 配置文件（加密）
- [ ] 密钥管理服务（KMS）
- [ ] HashiCorp Vault

---

### 16.2 安全防护 🟡

**Q16.2.1**: 您需要哪些安全防护？(可多选)
- [x] SQL 注入防护（参数化查询）
- [x] XSS 防护（输入过滤）
- [x] CSRF 防护（Token 验证）
- [ ] 请求频率限制（Rate Limiting）
- [ ] DDoS 防护

**注**: PRD 明确 Rate Limiting 不在 MVP 范围

---

## 1️⃣7️⃣ 性能优化

### 17.1 性能目标 🔴

**Q17.1.1**: 您是否接受 PRD 定义的性能目标？
- [x] ✅ 接受 - API P95 < 100ms, QPS > 10,000
- [ ] ❌ 需调整 - 我的目标: _____________

**Q17.1.2**: 您计划如何实现性能目标？(可多选)
- [x] 数据库索引优化
- [x] 缓存（Redis）
- [x] 连接池（DB、Redis）
- [x] 异步处理（队列）
- [ ] CDN（静态资源）
- [ ] 负载均衡（水平扩展）

---

## 1️⃣8️⃣ 开发规范

### 18.1 代码规范 🟡

**Q18.1.1**: 您希望使用什么项目结构？

**方案A - 标准 Go 布局**:
```
cmd/          # 可执行程序入口
internal/     # 私有代码
pkg/          # 公共库
api/          # API 定义（Proto/OpenAPI）
```

**方案B - DDD 分层**:
```
domain/       # 领域模型
application/  # 应用服务
infrastructure/ # 基础设施
```

**方案C - 微服务分离**:
```
services/
  auth/
  data/
  storage/
```

**现状**: 项目使用方案 A（已有 cmd/、internal/、pkg/）

**您的选择**: 混合方案

---

### 18.2 API 规范 🟡

**Q18.2.1**: 您希望使用什么 API 文档工具？

- [x] Swagger/OpenAPI (推荐)
- [ ] API Blueprint
- [ ] Postman Collection
- [ ] 手写文档

**现状**: docs/api.md 已有部分文档

**Q18.2.2**: 您希望如何定义统一响应格式？

**示例A**: Success  
```json
{
  "success": true,
  "code": 200,
  "message": "操作成功",
  "data": {"user": {"id": 1, "name": "Alice"}}
}
```

**示例B**: Failed  
```json
{
  "success": false,
  "code": 400,
  "message": "参数错误",
  "error": {"details": "用户名不能为空"}
}
```

**示例C**:  Pagination Success
```json
{
  "success": true,
  "code": 200,
  "message": "查询成功",
  "data": {
    "items": [
      {"id": 1, "name": "Alice"},
      {"id": 2, "name": "Bob"}
    ],
    "pagination": {
      "total": 100,        // 总记录数
      "page": 1,           // 当前页码
      "pageSize": 10,      // 每页大小
      "totalPages": 10     // 总页数
    }
  }
}
```

**示例D**:  Pagination failed  

```json
{
  "success": false,
  "code": 400,
  "message": "参数错误",
  "error": {
    "code": "INVALID_PAGE",
    "message": "页码超出范围",
    "details": "page 必须在 1-10 之间"
  }
}
```

**您的偏好**: A+B+C+D

---

## 📊 决策优先级总结

请在回答完所有问题后，总结您的核心决策：

### 🔴 必答题决策（立即生效）

1. **系统架构风格**:模块化单体 (Modular Monolith)，随时演进为微服务。
2. **主要编程语言**:go
3. **数据库选型**: PostgreSQL - 功能最全，与 Ent ORM 完美集成
4. **认证服务**: 集成 Ory Kratos - 生产级认证，共享数据库集成
5. **文件存储**: 混合方案：本地文件系统（优先）+ S3支持（接口抽象） - 虚拟文件系统，支持本地和云存储。
6. **工作流引擎**: Waterflow (Websoft9) - 基于 Temporal 的平台，自家产品可深度定制。
7. **消息队列**: Redis Streams - 轻量、可复用已有 Redis，短期持久化，分区内有序。
8. **API 网关**: 自研简化网关 (Go) - 基于 Chi Router + httputil.ReverseProxy，轻量可控。
9. **日志方案**: 日志写文件 + 简单查询 API，极简部署

### 🟡 重要决策（影响架构设计）

1. **缓存方案**:多层缓存（本地 + 分布式） - L1 进程内缓存 + L2 Redis，用于 API 响应、会话、配置。
2. **授权模型**: Project-based RBAC - 用户在不同 Project 有不同角色，继承 RBAC，策略引擎 Casbin。
3. **函数执行隔离**: 进程隔离 - 每个函数独立进程，简单快速，支持 Go 语言。
4. **实时推送方案**: WebSocket 框架 - coder/websocket（原为 nhooyr.io/websocket），应用层拦截 ORM Hooks 实现 CDC。
5. **监控系统**:Prometheus + Grafana - 监控系统资源、API 性能、业务指标，邮件/Webhook 告警。

### 🟢 可选决策（可延迟到实施阶段）

1. **国际化存储**: 数据库（Key-Value 表） - 支持中文（简体/繁体）和英文，检测优先级：用户设置 > Cookie > HTTP Header > IP 地理位置。
2. **License 方案**: 简单开关 + 时间限制 + 离线验证 + 非对称加密（RSA/ECDSA） 

---

## 技术栈清单

- **语言**: Go 1.24+, golangci-lint 1.64.8, gosec 2.22.7+
- **数据库**: PostgreSQL 14+, Redis 7+ (可选)
- **ORM**: Ent + Atlas
- **认证**: Ory Kratos
- **工作流**: Waterflow (Temporal)
- **路由**: Chi Router
- **WebSocket**: coder/websocket
- **文件系统**: spf13/afero
- **监控**: Prometheus + Grafana
- **容器**: Docker 20.10+
- **Node**: Node.js 18+
- **Git**: Git 2.30+

## 📝 开放式问题

**Q1**: 除了 PRD 列出的 13 个模块，您是否有其他需求？
_________________________

**Q2**: 您预期的 MVP 发布时间线？(仅供参考，不作承诺)
_________________________

**Q3**: 您是否有特定的技术债务或历史包袱需要考虑？
_________________________

**Q4**: 您对开源 vs 商业闭源的态度？
- [x] 完全开源（如当前项目）
- [ ] 核心开源，插件闭源
- [ ] 完全闭源

---

## 🎯 下一步行动

完成此问题清单后，架构师将：

1. **创建技术架构文档** (docs/architecture/tech-architecture.md)
   - 基于您的回答做出技术选型
   - 设计模块间交互
   - 定义接口规范

2. **创建部署架构文档** (docs/architecture/deployment-architecture.md)
   - Docker Compose 配置
   - 服务依赖关系图
   - 网络和安全配置

3. **创建数据架构文档** (docs/architecture/data-architecture.md)
   - 数据模型设计
   - 数据流图
   - 缓存策略

4. **更新项目规范** (docs/standards/)
   - API 设计规范
   - 代码规范
   - 测试规范

---

**文档版本**: 1.0  
**最后更新**: 2025-12-25  
**状态**: 待填写
