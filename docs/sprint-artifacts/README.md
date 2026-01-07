# Sprint Artifacts README
# apprun BaaS Platform

**Last Updated**: 2025-12-27  
**Maintainer**: Architect Agent

---

## BMad Method Workflow Hierarchy

```
Epic (Business Requirements)
    ↓
Sprint (Time-boxed Cycle)
    ↓
Story (Executable Tasks)
```

### Relationship

- **Epic**: Business-level features, defines "what to build"
  - Example: User Authentication, File Storage, Function Execution
  
- **Sprint**: Fixed time cycle (typically 2 weeks)
  - Contains multiple related Stories
  - Each Sprint focuses on partial implementation of one or more Epics
  
- **Story**: Specific executable tasks, defines "how to build"
  - Each Story belongs to a specific Sprint
  - Includes acceptance criteria, effort estimation, technical implementation

---

## Sprint-Story-Module Mapping

<!-- MAPPING_TABLE_START -->
| Sprint | Story & Description | Module | Status |
|--------|---------------------|--------|--------|
| Sprint-0 | Story 01: Docker Development & Deployment Environment | Infrastructure | Done |
| Sprint-0 | Story 02: 统一响应工具包 | Infrastructure | Done |
| Sprint-0 | Story 03: 业务错误包装框架 | Infrastructure | Done |
| Sprint-0 | Story 04: Ent Schema 配置管理 | Config | Done |
| Sprint-0 | Story 05: # Story 5a: 数据库增量迁移 | Database | ✅ Done |
| Sprint-0 | Story 05: # Story 5b: 服务端 CLI 框架 | CLI/Admin | Planning |
| Sprint-0 | Story 05: CI/CD 流水线与 Linter | Infrastructure | ✅ Done |
| Sprint-0 | Story 06: 测试框架与工具集 | Infrastructure | Planning |
| Sprint-0 | Story 07: 重构现有 Handlers | Infrastructure | Planning |
| Sprint-0 | Story 08: Database Field i18n 数据库字段国际化 | Infrastructure | Planning |
| Sprint-0 | Story 08: # Story 8b: i18n Integration Plan - i18n 集成计划 | Infrastructure | Planning |
| Sprint-0 | Story 08: i18n 国际化基础设施 | Infrastructure | Planning |
| Sprint-0 | Story 09: Localization (l10n) - 本地化支持 | pkg/l10n | Planning |
| Sprint-0 | Story 10: Configuration Center Foundation |  | Done |
| Sprint-0 | Story 11: Swagger API Documentation |  | Done |
| Sprint-0 | Story 12: 日志防腐层 | Infrastructure | Done |
| Sprint-1 | Story 13: Request Package | Infrastructure | Planning |
| Sprint-1 | Story 14: HTTP Server Package | Infrastructure | Done |
| Sprint-1 | Story 15: Environment Variable Utility Package | Infrastructure | Done |
| Sprint-1 | Story 16: Database Anti-Corruption Layer | Infrastructure | Done ✅ |
| Sprint-1 | Story 17: Go 版本升级到 1.25.5 | Infrastructure | ✅ Done |
<!-- MAPPING_TABLE_END -->

---

## Agent-Driven Workflow

### Story Creation (Architect Agent)
```
1. Analyze PRD from PM Agent
2. Create Story file: docs/sprint-artifacts/sprint-X/story-XX-title.md
3. Fill metadata and content (following Story Standards)
4. Run validation: make validate-stories
5. Commit changes: git commit -am "feat: add Story XX"
   → Git hook auto-syncs global index ✅
```

### Story Status Update (Dev Agent)
```
1. Open Story file
2. Update **Status** field: Planning → In Progress → Done
3. Add **Issue** field: #123 (when GitHub issue created)
4. Commit changes: git commit -am "chore: update Story XX status"
   → Git hook auto-syncs global index ✅
```

### Manual Sync (if needed)
```bash
make sync-index  # Manually sync global Stories index
```

**Note**: The global index (Sprint-Story-Module Mapping table) is automatically synced on every commit that modifies Story files.

---

## Story Standards

See [Story Standards](../standards/story-standards.md) for complete field definitions.

**Required fields**:
- **Priority**: P0/P1/P2
- **Effort**: X days
- **Owner**: Role name
- **Dependencies**: Story IDs or -
- **Status**: Planning/In Progress/Done/Blocked
- **Module**: Infrastructure/Auth/Storage/Functions/Management
- **Issue**: #TBD or #123

---

## Validation

```bash
# Validate all Story documents
make validate-stories
```

Agents should run this command after creating or updating Story files.

---

## Module Categories

- **Infrastructure**: Docker, CI/CD, Config, Testing
- **Auth**: Login, Permissions, Token, RBAC
- **Storage**: File upload, Object storage
- **Functions**: Execution, Scheduling
- **Management**: User admin, Monitoring

---

**Document Maintained By**: Architect Agent  
**Created**: 2025-12-26  
**Last Updated**: 2025-12-27
