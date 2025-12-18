# POC 计划调整说明

**调整日期**: 2025-12-18  
**调整原因**: 
1. Temporal 工作流引擎已通过独立 POC 验证
2. MinIO 已转为闭源商业产品，不再使用

---

## 📋 调整内容

### ✅ 计划缩减：5天 → 3天

**原计划（5天）**：
```
Day 1: 环境搭建 + PostgREST
Day 2: Ory Kratos + Casbin
Day 3: WASM 函数运行时
Day 4: Temporal 工作流        ← 已验证，移除
Day 5: 集成测试 + 报告
```

**新计划（3天）**：
```
Day 1: 环境搭建 + PostgREST
Day 2: Ory Kratos + Casbin + WASM
Day 3: 集成测试 + 报告
```

### ✅ 验证目标调整

**原目标（5项）**：
1. ✅ Go + PostgREST 集成
2. ✅ Ory Kratos + Casbin 认证授权
3. ✅ WASM 函数运行时
4. ~~Temporal 工作流引擎~~
5. ~~整体资源占用 ~1GB~~

**新目标（3项）**：
1. ✅ Go + PostgREST 集成
2. ✅ Ory Kratos + Casbin 认证授权
3. ✅ WASM 函数运行时

### ✅ 服务组件调整

**原 POC 环境（5个服务）**：
- PostgreSQL 15
- PostgREST 12
- Ory Kratos
- ~~Temporal~~ ← 移除
- ~~MinIO~~ ← 移除

**新 POC 环境（3个服务）**：
- PostgreSQL 15
- PostgREST 12
- Ory Kratos

### ✅ 资源目标调整

**原目标**：
- 总内存 < 1GB
- PostgreSQL: 256MB
- PostgREST: 50MB
- Kratos: 50MB
- Temporal: 150MB
- MinIO: 50MB

**新目标**：
- 总内存 < 400MB（核心服务）
- PostgreSQL: 256MB
- PostgREST: 50MB
- Kratos: 50MB
- apprun-core: ~100MB（目标）

---

## 🔄 对象存储方案调整

### MinIO 弃用原因
- **闭源转型**：MinIO 已转为 AGPL + 商业双授权模式
- **开源风险**：不再适合作为长期依赖
- **法律风险**：可能涉及商业许可问题

### 替代方案考虑

| 方案 | 许可证 | 特性 | 推荐度 |
|------|--------|------|--------|
| **SeaweedFS** | Apache 2.0 | 轻量、高性能、S3兼容 | ★★★★★ |
| **Rook/Ceph** | Apache 2.0 | 企业级、功能强大 | ★★★☆☆ |
| **本地文件系统** | - | 简单、无依赖 | ★★★★☆ |

**推荐**：
- MVP 阶段：本地文件系统 + 简单 S3 API 包装
- 生产阶段：SeaweedFS（开源、轻量、S3 兼容）

---

## 📊 Temporal 验证状态

### 已验证内容
- [x] Temporal Server 部署（SQLite 模式）
- [x] 工作流定义和执行
- [x] Activity 注册和调用
- [x] 失败重试机制
- [x] Web UI 监控

### 验证结论
- ✅ **技术可行**：Temporal 完全满足工作流需求
- ✅ **性能良好**：SQLite 模式下资源占用 ~150MB
- ✅ **成熟稳定**：生产级工作流引擎，久经考验
- ✅ **集成简单**：Go SDK 易用，文档完善

### 无需重复验证
Temporal 已通过独立 POC 验证，无需在本次 POC 中重复。

---

## 📝 文档更新清单

- [x] `docs/poc/poc-validation-plan-2025-12-18.md` - 更新为3天计划
- [x] `poc/docker-compose.yml` - 移除 Temporal 和 MinIO
- [x] `poc/start-poc.sh` - 更新服务列表
- [x] `poc/README.md` - 更新验证目标和服务说明
- [x] `docs/poc/POC-ADJUSTMENT-NOTICE.md` - 本文档（调整说明）

---

## 🎯 下一步行动

### 立即可执行
```bash
cd poc/
./start-poc.sh
```

**启动的服务**：
- PostgreSQL（数据库）
- PostgREST（数据 API）
- Ory Kratos（认证服务）

### 3天验证计划

**Day 1（8小时）**：
- 环境搭建
- Go + PostgREST 集成
- 多租户隔离验证
- 性能基准测试

**Day 2（8小时）**：
- Ory Kratos 部署配置
- Casbin RBAC 集成
- WASM 运行时实现
- 性能测试

**Day 3（8小时）**：
- 端到端集成测试
- 性能基准测试
- 生成 POC 报告

### 评审时间
**Day 3 下午**：POC 评审会议，Go/No Go 决策

---

## 💡 优势说明

### 缩减后的优势
1. **时间更紧凑**：3天 vs 5天，节省 40% 时间
2. **目标更聚焦**：专注核心集成（PostgREST + Auth + WASM）
3. **资源更轻量**：400MB vs 1GB，减少 60% 内存占用
4. **风险更低**：Temporal 已验证，无不确定性

### 不影响功能
- ✅ Temporal 工作流：已验证可用，直接集成
- ✅ 对象存储：后续选择开源替代方案
- ✅ 核心能力：PostgREST + Auth + WASM 完整验证

---

## ⚠️ 注意事项

### 对象存储方案待定
- 本次 POC **不包含**对象存储验证
- MVP 阶段可使用本地文件系统
- 生产阶段推荐 SeaweedFS

### Temporal 集成计划
- 架构设计中已包含 Temporal
- MVP 开发时直接集成（已验证）
- 无需额外 POC 验证

---

**文档维护**
- 创建者: Root (架构师)
- 创建日期: 2025-12-18
- 版本: 1.0
- 状态: ✅ 已完成

---

🎯 **调整后的 POC 计划更加高效！** 🚀
