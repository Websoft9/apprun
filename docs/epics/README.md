````markdown
# Epics

**📌 Single Source of Truth for Epic Definitions**

Epic 设计文档 - BMad Method 业务级功能定义

> **重要**: 此文件夹是所有 Epic 的唯一事实来源。每个 Epic 都有独立的文件，便于版本控制和协作。

---

## 📁 Epic File Structure

```
docs/epics/
├── README.md                      ← Epic 索引和状态总览
├── 1-infrastructure-epic.md       ← Epic 1: Platform Infrastructure
├── 2-i18n-epic.md                 ← Epic 2: Internationalization
├── 3-config-epic.md               ← Epic 3: Configuration Management
├── 4-docs-epic.md                 ← Epic 4: API Documentation
├── 5-auth-epic.md                 ← Epic 5: Authentication & Authorization
├── 6-storage-epic.md              ← Epic 6: File Storage Service
├── 7-functions-epic.md            ← Epic 7: Serverless Functions
└── ...                            ← More epics to be added
```

**Naming Convention**: `N-name-epic.md` where N is the epic number

---

## Active Epics

| Epic | Status | Priority | Progress | Description |
|------|--------|----------|----------|-------------|
| [Infrastructure & Foundation](./1-infrastructure-epic.md) | 🔄 In Progress | P0 | 82% (14/17) | 平台基础设施与核心工具包 |
| [Configuration Management](./3-config-epic.md) | 🔄 In Progress | P0 | 50% (1/2) | 配置中心与多环境管理 |
| [Authentication & Authorization](./5-auth-epic.md) | 🔄 In Progress | P1 | 33% (1/3) | 用户认证、授权与权限管理 |
| [Internationalization & Localization](./2-i18n-epic.md) | 📝 Planning | P1 | 0% (0/4) | 多语言支持与本地化 |
| [File Storage Service](./6-storage-epic.md) | 📝 Planning | P2 | 0% (0/0) | 对象存储与文件管理 |
| [Serverless Functions](./7-functions-epic.md) | 📝 Planning | P2 | 0% (0/0) | 函数计算与任务调度 |

---

## Epic Lifecycle

```
Planning → In Progress → Done → Archived
```

- **Planning**: 需求分析，Story 定义
- **In Progress**: 开发中，至少 1 个 Story 进行中
- **Done**: 所有 Story 完成，功能可用
- **Archived**: 已归档，不再维护

---

## Related Docs

- [Sprint Status](../sprint-artifacts/sprint-status.yaml) - Epic-Story 关系追踪
- [Story Index](../sprint-artifacts/story-index.md) - Story 状态总览
