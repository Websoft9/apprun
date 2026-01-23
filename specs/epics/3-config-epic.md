# Epic 3: Configuration Management

**Epic ID**: epic-3  
**Status**: ✅ Completed  
**Priority**: P0 (Critical)  
**Owner**: Backend Team  
**Timeline**: Sprint 0-1

---

## Vision

构建灵活、类型安全、易于管理的配置中心，支持多环境、动态配置和配置审计。

---

## Goals

- ✅ 基于 Ent Schema 的配置存储
- ✅ CRUD API（支持 namespace/key/value）
- ✅ 配置历史与审计
- ✅ 模块配置注册与管理
- ✅ 配置示例文件自动生成
- 🔄 多环境配置隔离（Future）
- 🔄 配置热更新通知（Future）

---

## Key Stories

### ✅ Completed (3/3)
- **Story 3.1**: Ent Schema 配置管理（基础 CRUD）- [文件](../sprint-artifacts/sprint-0/3-1-ent-schema.md)
- **Story 3.2**: Configuration Center Foundation（完整实现）- [文件](../sprint-artifacts/sprint-1/3-2-config-basic.md)
- **Story 3.3**: Module Configuration Registry（模块配置注册）- [文件](../sprint-artifacts/sprint-1/3-3-module-config-registry.md)

### 📝 Pending (0/3)
- _(All core stories completed)_

### 📌 Moved to Infrastructure Epic
- **Story 1.18**: CLI Generate - Unified Code Generation（原 Story 3.4，已扩展为统一生成工具）- 见 [1-infrastructure-epic](./1-infrastructure-epic.md)

---

## Architecture

```
┌─────────────────────────────────────────┐
│  Config API (REST)                      │
│  /config/{namespace}/{key}              │
└─────────────┬───────────────────────────┘
              │
┌─────────────▼───────────────────────────┐
│  Config Service                         │
│  - CRUD                                 │
│  - Validation                           │
│  - Change Notification (v2)             │
└─────────────┬───────────────────────────┘
              │
┌─────────────▼───────────────────────────┐
│  Ent Schema (PostgreSQL)                │
│  - namespace, key, value, type          │
│  - created_by, updated_by, updated_at   │
│  - is_encrypted, is_sensitive           │
└─────────────────────────────────────────┘
```

---

## Use Cases

### UC1: 应用配置读取
```bash
GET /config/app/smtp_host
→ { "namespace": "app", "key": "smtp_host", "value": "smtp.example.com" }
```

### UC2: 多环境配置
```bash
GET /config/app.prod/database_url
GET /config/app.dev/database_url
```

### UC3: 配置审计
- 记录每次配置变更
- 支持版本回滚
- 追踪修改人员

---

## Success Metrics

- ✅ Config API 可用性 99.9%+
- ✅ 配置读取延迟 < 10ms
- ✅ 配置示例文件自动生成
- ✅ 所有模块遵循统一配置模式
- 🔄 支持 10,000+ 配置项（Future）
- 🔄 配置变更实时通知（< 1s）（Future）

---

## Security

- 🔒 敏感配置加密存储（AES-256）
- 🔒 基于 RBAC 的配置访问控制
- 🔒 配置变更审计日志
- 🔒 API 密钥轮换机制

---

## Dependencies

- **epic-infrastructure**: 数据库、响应框架、错误处理
- **epic-auth**: 配置访问权限控制（未来）

---

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/config` | GET | List all configs |
| `/config/{namespace}` | GET | List by namespace |
| `/config/{namespace}/{key}` | GET | Get single config |
| `/config` | POST | Create config |
| `/config/{id}` | PUT | Update config |
| `/config/{id}` | DELETE | Delete config |

完整文档见：`docs/api.md`

---

## Technical Decisions

### Schema 设计
- **namespace**: 支持环境隔离（app.prod, app.dev）
- **value_type**: 支持 string, number, boolean, json
- **metadata**: 扩展字段，存储描述、标签等

### 实现方式
- Ent ORM 管理 Schema
- REST API（未来支持 gRPC）
- 基于 PostgreSQL（支持事务）

---

### Future Enhancements

- 🚀 **Story 3.5**: 配置热更新（WebSocket/SSE）
- 🚀 **Story 3.6**: 配置版本回滚
- 🚀 **Story 3.7**: 配置审计日志增强
- 🚀 **Story 3.8**: 多环境配置隔离

---

## Related Stories

- [Story 3.1: Ent Schema 配置管理](../sprint-artifacts/sprint-0/3-1-ent-schema.md)
- [Story 3.2: Configuration Center Foundation](../sprint-artifacts/sprint-1/3-2-config-basic.md)
- [Story 3.3: Module Configuration Registry](../sprint-artifacts/sprint-1/3-3-module-config-registry.md)
- [Story 1.18: CLI Generate - Unified Code Generation](../sprint-artifacts/sprint-1/1-18-cli-generate.md) (originally Story 3.4, moved to Infrastructure Epic)

---

## Notes

配置中心是 BaaS 平台的**神经中枢**，影响所有服务的行为。必须保证：
- **高可用**：单点故障不影响配置读取
- **一致性**：配置变更必须原子性生效
- **安全性**：敏感配置加密，访问可审计

**当前进度**：✅ **100% 完成** (所有核心 stories 已实现，基础能力完备)

**核心成果**：
- 配置中心 CRUD API
- 模块配置注册机制
- 配置示例自动生成（`.env.example` + `config.example`）- 见 [Story 1.18: CLI Generate](../sprint-artifacts/sprint-1/1-18-cli-generate.md)
