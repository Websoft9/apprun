# Standards - 技术规范文档
# apprun BaaS Platform

**维护者**: Winston (Architect Agent)  
**最后更新**: 2025-12-26

---

## 📋 文档分类

| 类型 | 定义 | 示例 |
|-----|------|------|
| **Standards（技术规范）** | 定义代码编写规则和静态标准 | 命名规范、API 格式、测试方法 |
| **Processes（流程规范）** | 定义开发协作流程和动态过程 | 分支策略、PR 流程、发布步骤 |

---

## 📚 规范文档列表

| 文档 | 分类 | 适用场景 | 核心内容 |
|-----|------|---------|---------|
| **[architecture-standards.md](./architecture-standards.md)** | 技术规范 | 架构设计原则、模块化、扩展性、演进路径 | 解耦原则、分层架构、插件化、非侵入式设计、隔离策略、单体到微服务演进 |
| **[api-design.md](./api-design.md)** | 技术规范 | 设计 RESTful API、定义响应格式、错误码 | API 版本管理、URL 命名、HTTP 方法、统一响应格式、错误码规范、分页排序、认证授权 |
| **[coding-standards.md](./coding-standards.md)** | 技术规范 | 编写 Go 代码、命名变量、组织项目结构、定义 Ent Schema | 命名规范、代码结构、错误处理、注释规范、并发编程、Ent ORM 规范、代码审查清单 |
| **[testing-standards.md](./testing-standards.md)** | 技术规范 | 编写单元测试、集成测试、E2E 测试 | 测试策略（测试金字塔、覆盖率）、单元测试（AAA 模式、Mock）、集成测试、E2E 测试、性能测试、测试工具 |
| **[i18n-standards.md](./i18n-standards.md)** | 技术规范 | 国际化支持、多语言消息、API 本地化 | 语言检测、消息文件管理、go-i18n 集成、API 响应翻译、错误消息国际化 |
| **[localization-standards.md](./localization-standards.md)** | 技术规范 | 本地化支持、数据格式化、区域适配 | 货币格式化、日期时间格式化、数字格式化、度量单位转换、与 i18n 协作 |
| **[devops-process.md](./devops-process.md)** | 流程规范 | 开发流程、Git 工作流、代码审查、测试流程、发布流程 | Story 开发循环、分支策略、Commit 规范、PR 模板、Code Review 清单、CI/CD 配置、版本发布 |

---

## 🎯 快速查找

| 问题 | 查阅文档 |
|-----|---------|
| 如何设计模块架构？ | [architecture-standards.md](./architecture-standards.md) Section 1 |
| 如何实现插件化？ | [architecture-standards.md](./architecture-standards.md) Section 2 |
| 如何保证多租户隔离？ | [architecture-standards.md](./architecture-standards.md) Section 5 |
| 如何设计 API？ | [api-design.md](./api-design.md) |
| 如何命名变量/函数？ | [coding-standards.md](./coding-standards.md) Section 1 |
| 如何处理错误？ | [coding-standards.md](./coding-standards.md) Section 3 |
| 如何定义 Ent Schema？ | [coding-standards.md](./coding-standards.md) Section 12 |
| 如何实现 i18n？ | [i18n-standards.md](./i18n-standards.md) Section 5 |
| 如何翻译错误消息？ | [i18n-standards.md](./i18n-standards.md) Section 2.1 |
| 如何格式化货币/日期/数字？ | [localization-standards.md](./localization-standards.md) Section 4-5 |
| 如何编写测试？ | [testing-standards.md](./testing-standards.md) Section 2-4 |
| 如何使用 Mock？ | [testing-standards.md](./testing-standards.md) Section 2.4 |
| 如何创建分支？ | [devops-process.md](./devops-process.md) Section 2.1 |
| 如何写 Commit Message？ | [devops-process.md](./devops-process.md) Section 2.2 |
| 如何创建 PR？ | [devops-process.md](./devops-process.md) Section 3.1-3.2 |
| 如何 Code Review？ | [devops-process.md](./devops-process.md) Section 3.3 |
| 如何运行测试？ | [devops-process.md](./devops-process.md) Section 4.2 |
| 如何发布版本？ | [devops-process.md](./devops-process.md) Section 5 |

---

**相关文档**: [PRD](../prd.md) | [Epics](../epics/) | [Architecture](../architecture/) | [Sprint Artifacts](../sprint-artifacts/)
