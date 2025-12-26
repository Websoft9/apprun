# Sprint Artifacts 索引
# apprun BaaS Platform

**最后更新**: 2025-12-26  
**维护者**: Architect Agent

---

## 📋 Sprint 总览

| Sprint | 周期 | 状态 | 主要模块 | 进度 | 关键 Stories |
|--------|------|------|---------|------|-------------|
| [Sprint 0](./sprint-0/) | 12/26-01/09 | Planning | 基础设施 | 0% | 响应工具包、错误处理、CI/CD |
| Sprint 1 | 待定 | 待开始 | 认证模块 | 0% | 待规划 |
| Sprint 2 | 待定 | 待开始 | 存储模块 | 0% | 待规划 |

---

## 🎯 按模块查找 Sprint

### 基础设施
- **Sprint 0** (2025-12-26 ~ 2026-01-09)
  - Story 1: 统一响应工具包
  - Story 2: 错误处理框架
  - Story 3: Ent Schema 规范配置
  - Story 4: CI/CD Linter 检查配置
  - Story 5: 测试框架与工具包
  - Story 6: 更新现有代码使用新工具

### 认证与授权模块
- **Sprint 1** (待规划)
  - Story 1: Kratos 集成与 Session 验证
  - Story 2: JWT Token 管理
  - Story 3: 认证中间件
  - Story 4: RBAC 权限控制
  - Story 5: 用户信息接口

### 存储服务模块
- **Sprint 2** (待规划)
  - Story 1: 存储后端抽象层
  - Story 2: 文件上传功能
  - Story 3: 文件下载功能
  - Story 4: 文件列表与删除
  - Story 5: 文件夹管理
  - Story 6: 存储配额管理

### 函数服务模块
- **Sprint 3** (待规划)
  - Story 1: 函数管理基础
  - Story 2: 函数编译引擎
  - Story 3: 函数执行引擎
  - Story 4: HTTP 触发功能
  - Story 5: 执行日志管理
  - Story 6: 事件触发功能

---

## 🔍 按关键词查找

### 响应处理
- Sprint 0 > Story 1: 统一响应工具包

### 错误处理
- Sprint 0 > Story 2: 错误处理框架

### 数据库 / ORM
- Sprint 0 > Story 3: Ent Schema 规范配置

### CI/CD
- Sprint 0 > Story 4: CI/CD Linter 检查配置

### 测试
- Sprint 0 > Story 5: 测试框架与工具包

### 认证 / 授权
- Sprint 1 > Story 1-5: 认证与权限相关

### 文件存储
- Sprint 2 > Story 1-6: 存储服务相关

### 函数执行
- Sprint 3 > Story 1-6: 函数服务相关

---

## 📊 Sprint 统计

### 总体进度
- **已完成 Sprint**: 0
- **进行中 Sprint**: 0
- **待开始 Sprint**: 4
- **总 Stories**: 22 (已计划)
- **已完成 Stories**: 0

### 按优先级统计
- **P0 (必需)**: 16 Stories
- **P1 (重要)**: 6 Stories
- **P2 (可选)**: 0 Stories

### 按模块统计
- **基础设施**: 6 Stories (Sprint 0)
- **认证模块**: 5 Stories (Sprint 1)
- **存储模块**: 6 Stories (Sprint 2)
- **函数模块**: 6 Stories (Sprint 3)

---

## 📅 Sprint 时间线

```
2025-12-26 ━━━━━━━━━━━━━━━━━━━━ 2026-01-09  Sprint 0 (基础设施)
2026-01-10 ━━━━━━━━━━━━━━━━━━━━ 2026-01-23  Sprint 1 (认证模块)
2026-01-24 ━━━━━━━━━━━━━━━━━━━━ 2026-02-06  Sprint 2 (存储模块)
2026-02-07 ━━━━━━━━━━━━━━━━━━━━ 2026-02-20  Sprint 3 (函数模块)
```

---

## 🔗 快速导航

### Sprint 0: 基础设施建设
- 📄 [Stories 文档](./sprint-0/stories.md)
- 📈 [进度报告](./sprint-0/progress.md)
- 📝 [Sprint 总结](./sprint-0/summary.md)

### 相关文档
- 📖 [PRD](../prd.md)
- 🏗️ [Epics](../epics/)
  - [认证 Epic](../epics/auth-epic.md)
  - [存储 Epic](../epics/storage-epic.md)
  - [函数 Epic](../epics/functions-epic.md)
- 📐 [技术规范](../standards/)
  - [API 设计规范](../standards/api-design.md)
  - [编码规范](../standards/coding-standards.md)
  - [测试规范](../standards/testing-standards.md)
- 🏛️ [架构文档](../architecture/)
  - [技术架构](../architecture/tech-architecture.md)
  - [部署架构](../architecture/deployment-architecture.md)

---

## 📊 Epic 与 Sprint 映射

### 认证与授权 Epic → Sprint 分配
- **Sprint 1** (核心功能): Kratos 集成、JWT 管理、基础中间件
- **Sprint 2** (权限控制): RBAC 实现、高级权限策略
- **Sprint 3** (完善优化): 性能优化、安全加固

### 存储服务 Epic → Sprint 分配
- **Sprint 2** (核心功能): 文件上传下载、基础管理
- **Sprint 3** (高级功能): 配额管理、S3 集成、缓存优化

### 函数服务 Epic → Sprint 分配
- **Sprint 3** (核心功能): 函数管理、编译执行、HTTP 触发
- **Sprint 4** (高级功能): 事件触发、性能优化、监控

---

## 📈 进度追踪

### 里程碑
- [ ] **M0**: 基础设施就绪 (Sprint 0 完成)
- [ ] **M1**: 用户认证可用 (Sprint 1 完成)
- [ ] **M2**: 文件存储可用 (Sprint 2 完成)
- [ ] **M3**: 函数执行可用 (Sprint 3 完成)
- [ ] **M4**: MVP 发布 (Sprint 4 完成)

### 关键指标
- **代码覆盖率目标**: > 80%
- **CI 构建时间目标**: < 5 分钟
- **API 响应时间目标**: P95 < 100ms
- **单元测试通过率目标**: 100%

---

## 🏆 Sprint 最佳实践

### Story 大小
- **理想**: 1-3 天完成
- **最大**: 不超过 5 天
- **拆分标准**: INVEST 原则

### Definition of Done
1. 代码实现完成
2. 单元测试通过（覆盖率 > 90%）
3. 代码审查通过
4. 文档编写完成
5. 验收标准全部满足

### Sprint 节奏
- **Planning**: Sprint 第一天
- **Daily Sync**: 每日更新 progress.md
- **Review**: Sprint 最后一天
- **Retrospective**: Sprint 结束后

---

## 🔄 文档更新记录

| 日期 | 更新内容 | 更新者 |
|------|---------|--------|
| 2025-12-26 | 创建索引文档，添加 Sprint 0 | Architect Agent |
| 待定 | Sprint 1 规划 | 待定 |
| 待定 | Sprint 2 规划 | 待定 |

---

## 📝 使用指南

### 查找 Story
1. **按时间**: 查看 Sprint 时间线，选择对应 Sprint
2. **按模块**: 使用"按模块查找 Sprint"部分
3. **按关键词**: 使用"按关键词查找"部分

### 跟踪进度
1. 查看 Sprint 总览表中的"进度"列
2. 访问具体 Sprint 的 `progress.md` 查看详细进度
3. 查看"Sprint 统计"了解整体情况

### 查看总结
- 每个 Sprint 完成后，访问对应的 `summary.md` 查看总结和经验教训

---

**文档维护**: Architect Agent  
**创建日期**: 2025-12-26  
**最后更新**: 2025-12-26  
**下次更新**: Sprint 0 完成后
