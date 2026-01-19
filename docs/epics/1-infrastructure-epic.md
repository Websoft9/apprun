# Epic 1: Infrastructure & Foundation

**Epic ID**: epic-1  
**Status**: In Progress  
**Priority**: P0 (Critical)  
**Owner**: Platform Team  
**Timeline**: Sprint 0-1

---

## Vision

建立稳定、可维护、可扩展的 BaaS 平台基础设施，为上层业务功能提供坚实支撑。

---

## Goals

- ✅ 容器化开发与部署环境
- ✅ 统一的响应、错误处理、日志框架
- ✅ 数据库防腐层与增量迁移
- ✅ CI/CD 流水线与代码质量保障
- ✅ 核心工具包（Server、Env、Database、Redis Cache）
- 🔄 测试框架与 CLI 工具
- 🔄 重构现有 Handlers

---

## Key Stories

### ✅ Completed (15/19)
- **Story 1.1**: Docker 开发部署环境
- **Story 1.2**: 统一响应工具包
- **Story 1.3**: 业务错误包装框架
- **Story 1.4**: CI/CD 流水线与 Linter
- **Story 1.5**: 数据库增量迁移
- **Story 1.6**: Unified CLI Architecture (统一 CLI 命令架构)
- **Story 1.7**: Makefile 标准与命令分组
- **Story 1.10**: 日志防腐层
- **Story 1.12**: HTTP Server Package
- **Story 1.13**: Environment Variable Utility
- **Story 1.14**: Database Anti-Corruption Layer
- **Story 1.15**: Go 1.25.5 升级
- **Story 1.16**: Redis Cache Package

### 🔄 In Progress (1/19)
- **Story 1.18**: CLI Generate - Unified Code Generation (统一代码生成工具) - [文件](../sprint-artifacts/sprint-1/1-18-cli-generate.md)

### 📝 Pending (3/19)
- **Story 1.8**: 测试框架与工具集
- **Story 1.9**: 重构现有 Handlers
- **Story 1.11**: Request Package
- **Story 1.17**: Platform Initialization (平台初始化)

---

## Success Metrics

- ✅ 开发环境一键启动（make dev-start）
- ✅ 代码质量门禁通过率 100%
- ✅ 核心包测试覆盖率 > 80%
- 🔄 端到端测试覆盖关键流程
- 🔄 部署流水线自动化

---

## Dependencies

- Docker & Docker Compose
- PostgreSQL 16+
- Redis 7+
- Go 1.25.5

---

## Technical Decisions

### 架构选择
- **Ent ORM**: 类型安全的 ORM 框架
- **Kratos**: HTTP/gRPC 服务框架
- **防腐层设计**: 隔离外部依赖，便于替换

### 工具链
- **golangci-lint**: 代码质量检查
- **Atlas**: 数据库 Schema 管理
- **Swagger**: API 文档生成

### 开发规范
- Makefile 命令分组（Story 05c）
- 错误码统一管理
- 结构化日志（slog）

---

## Related Epics

- **epic-config**: 配置管理（依赖基础设施）
- **epic-auth**: 认证授权（依赖基础设施）
- **epic-i18n**: 国际化（依赖基础设施）

---

## Notes

Infrastructure Epic 是平台的根基，**优先级最高**。所有上层功能都依赖这里的基础能力。

当前进度：**79% 完成** (15/19 stories done, 1 in progress)
