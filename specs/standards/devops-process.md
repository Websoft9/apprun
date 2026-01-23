# DevOps 流程规范
# apprun BaaS Platform

**创建日期**: 2025-12-26  
**维护者**: Winston (Architect Agent)  
**版本**: 1.0  
**状态**: Active

---

## 概述

本文档定义 apprun 项目的完整开发流程，涵盖从需求到发布的全生命周期。适用于 AI 辅助开发场景，优化了传统敏捷流程。

**核心原则**：
- **Story 驱动开发**：最小可交付单元
- **TDD 优先**：测试先行，质量内建
- **小步快跑**：频繁提交，持续集成
- **AI 协作**：充分利用 AI Agent 审查和生成

---

## 目录

1. [开发流程](#1-开发流程)
2. [Git 工作流](#2-git-工作流)
3. [代码审查流程](#3-代码审查流程)
4. [测试流程](#4-测试流程)
5. [发布流程](#5-发布流程)
6. [工具配置](#6-工具配置)
7. [常见问题](#7-常见问题)

---

## 1. 开发流程

### 1.1 分支策略

```
main (生产)
  └── develop (开发主线)
       └── sprint-{N}-story-{M}-{brief-description}
```

#### 分支规则
- **main**: 生产就绪代码，受保护，只接受来自 `develop` 的 PR
- **develop**: 开发主线，集成所有功能分支
- **feature**: 功能分支，从 `develop` 创建，命名格式：`sprint-{N}-story-{M}-{description}`

#### 分支示例
```bash
sprint-0-story-1-response-package
sprint-0-story-2-errors-framework
sprint-1-story-1-auth-session
```

---

### 1.2 Story 开发循环

#### **Step 1: 准备阶段**

```bash
# 1. 拉取最新代码
git checkout develop
git pull origin develop

# 2. 创建功能分支
git checkout -b sprint-0-story-1-response-package

# 3. 在 progress.md 中标记 Story 为 In Progress
vim docs/sprint-artifacts/sprint-0/progress.md
```

**确认清单**：
- [ ] 阅读 Story 验收标准（`stories.md`）
- [ ] 理解实现任务列表
- [ ] 检查依赖的 Story 是否完成
- [ ] 准备测试数据和 Mock 对象

---

#### **Step 2: TDD 开发**

```bash
# 1. 先写测试（Red）
touch core/pkg/response/response_test.go
vim core/pkg/response/response_test.go

# 2. 运行测试（应该失败）
cd core
go test ./pkg/response -v

# 3. 实现功能代码（Green）
touch core/pkg/response/response.go
vim core/pkg/response/response.go

# 4. 运行测试（应该通过）
go test ./pkg/response -v -race

# 5. 重构代码（Refactor）
# 优化实现，确保测试仍然通过
go test ./pkg/response -v -race

# 6. 检查覆盖率
go test -coverprofile=coverage.out ./pkg/response
go tool cover -func=coverage.out
# 目标：> 90% (P0) 或 > 80% (P1)
```

**TDD 循环**：
```
Red (写失败的测试) → Green (最小实现通过测试) → Refactor (优化代码) → Repeat
```

---

#### **Step 3: 代码检查**

```bash
# 1. 运行 Linter（需要先完成 Story 4）
cd core
golangci-lint run ./pkg/response

# 2. 运行所有测试（确保没有破坏现有功能）
go test -v -race ./...

# 3. 检查 Ent Schema（如果修改了 Schema）
cd ..
./scripts/check-ent-json-tags.sh

# 4. 生成覆盖率报告（可选）
cd core
go tool cover -html=coverage.out -o coverage.html
```

---

#### **Step 4: 提交代码**

```bash
# 1. 查看修改
git status
git diff

# 2. 添加文件
git add core/pkg/response/

# 3. 提交（使用规范的 commit message）
git commit -m "feat(response): implement Success/Error/List functions

- Add Response struct with success/error fields
- Implement Success() for successful responses
- Implement Error() for error responses  
- Implement List() for paginated list responses
- Add unit tests with 95% coverage
- Add README with usage examples

Story: Sprint-0 Story-1
Tests: All passing, coverage 95%
"
```

**Commit Message 格式**：
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型**：
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档更新
- `refactor`: 重构（不改变功能）
- `test`: 测试相关
- `chore`: 构建/工具配置

---

#### **Step 5: 推送和创建 PR**

```bash
# 1. 推送分支
git push origin sprint-0-story-1-response-package

# 2. 在 GitHub 创建 Pull Request
# 目标分支: develop
# 使用 PR 模板（见 3.2 节）

# 3. 请求 AI Agent Review
# 在 PR 评论中使用：
# @architect 请审查 API 设计
# @dev 请检查代码实现和测试
```

---

#### **Step 6: 修复 Review 意见**

```bash
# 1. 根据 Review 意见修改代码
vim core/pkg/response/response.go

# 2. 运行测试
go test ./pkg/response -v

# 3. 提交修复
git add .
git commit -m "fix(response): address review comments

- Improve error message formatting
- Add nil pointer check
- Update test cases
"

# 4. 推送修复
git push origin sprint-0-story-1-response-package
```

---

#### **Step 7: 合并和清理**

```bash
# 1. 所有检查通过后，合并 PR（Squash Merge）
# 在 GitHub Web 界面操作

# 2. 切回 develop 分支
git checkout develop
git pull origin develop

# 3. 删除本地功能分支（可选）
git branch -d sprint-0-story-1-response-package

# 4. 更新进度文档
vim docs/sprint-artifacts/sprint-0/progress.md
# - 标记 Story 为 Completed
# - 记录实际工作量
# - 记录遇到的问题和解决方案
```

---

### 1.3 开发最佳实践

#### ✅ **推荐做法**

```bash
# 1. 小步提交
git commit -m "feat(response): add Response struct"
git commit -m "feat(response): implement Success function"
git commit -m "test(response): add Success function tests"

# 2. 频繁推送（避免丢失代码）
git push origin sprint-0-story-1-response-package

# 3. 测试先行
# 先写 response_test.go
# 后写 response.go

# 4. 及时更新文档
vim core/pkg/response/README.md
```

#### ❌ **避免做法**

```bash
# 1. 避免大批量提交
# ❌ git commit -m "完成 Story 1"（包含 1000+ 行修改）

# 2. 避免长期不推送
# ❌ 本地开发 3 天不推送

# 3. 避免跳过测试
# ❌ go test ./... 失败但继续开发

# 4. 避免没有文档
# ❌ 只有代码，没有 README 和注释
```

---

## 2. Git 工作流

### 2.1 分支管理

#### **创建分支**

```bash
# 从 develop 创建功能分支
git checkout develop
git pull origin develop
git checkout -b sprint-0-story-1-response-package
```

#### **同步 develop 分支**

```bash
# 在功能分支中合并最新的 develop
git checkout sprint-0-story-1-response-package
git fetch origin
git merge origin/develop

# 或使用 rebase（保持线性历史）
git rebase origin/develop
```

#### **解决冲突**

```bash
# 1. 发现冲突
git merge origin/develop
# CONFLICT (content): Merge conflict in core/pkg/response/response.go

# 2. 手动解决冲突
vim core/pkg/response/response.go
# 编辑冲突标记：<<<<<<< HEAD 和 >>>>>>> origin/develop

# 3. 标记为已解决
git add core/pkg/response/response.go

# 4. 完成合并
git commit -m "merge: resolve conflicts with develop"

# 5. 运行测试确保没有问题
go test ./...
```

---

### 2.2 Commit 规范

#### **Commit Message 模板**

```
<type>(<scope>): <subject>

<body>

Story: Sprint-{N} Story-{M}
Tests: <test status>
```

#### **Type 类型**

| Type | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(auth): add JWT token validation` |
| `fix` | 修复 Bug | `fix(response): handle nil data correctly` |
| `docs` | 文档 | `docs(readme): update installation guide` |
| `refactor` | 重构 | `refactor(handler): simplify error handling` |
| `test` | 测试 | `test(response): add edge case tests` |
| `chore` | 构建/工具 | `chore(ci): update golangci-lint version` |
| `perf` | 性能优化 | `perf(cache): optimize cache key generation` |
| `style` | 代码格式 | `style(handler): fix indentation` |

#### **Scope 范围**

常用 scope：
- `response`, `errors`, `auth`, `storage`, `functions`
- `handler`, `middleware`, `repository`, `service`
- `ent`, `schema`, `migration`
- `ci`, `docker`, `config`

#### **Commit 示例**

```bash
# 好的 Commit
git commit -m "feat(response): implement Success/Error/List functions

- Add Response struct with success/error fields
- Implement Success() for 200 OK responses
- Implement Error() for error responses with error codes
- Implement List() for paginated responses
- Add unit tests with 95% coverage

Story: Sprint-0 Story-1
Tests: 15 tests passing, coverage 95%
"

# 简单的 Commit
git commit -m "fix(response): handle nil data in Success function"

# 文档 Commit
git commit -m "docs(response): add usage examples to README"
```

---

### 2.3 合并策略

#### **feature → develop: Squash Merge**

```bash
# 在 GitHub PR 中选择 "Squash and merge"
# 优点：保持 develop 分支历史简洁
# 结果：多个功能分支 commit 合并为 1 个 commit
```

#### **develop → main: Merge Commit**

```bash
# 在 GitHub PR 中选择 "Create a merge commit"
# 优点：保留完整的 develop 历史
# 结果：记录 Sprint 完整交付
```

#### **快速修复: Cherry-pick**

```bash
# 场景：production 紧急修复
git checkout main
git pull origin main
git checkout -b hotfix-auth-bug

# 修复代码并提交
git commit -m "fix(auth): handle expired token correctly"

# 合并到 main
git push origin hotfix-auth-bug
# 创建 PR: hotfix-auth-bug → main

# Cherry-pick 到 develop
git checkout develop
git cherry-pick <commit-hash>
git push origin develop
```

---

### 2.4 标签管理

#### **版本标签**

```bash
# 发布版本时打标签
git checkout main
git pull origin main
git tag -a v1.0.0 -m "Release v1.0.0

Features:
- Authentication module
- Storage module
- Functions module

Sprint: Sprint 1-3
"

git push origin v1.0.0
```

#### **标签规范**

```bash
# 格式: v{major}.{minor}.{patch}
v1.0.0  # 主版本发布
v1.1.0  # 新功能发布
v1.1.1  # Bug 修复

# 预发布版本
v1.0.0-alpha
v1.0.0-beta
v1.0.0-rc.1
```

---

## 3. 代码审查流程

### 3.1 创建 Pull Request

#### **PR 标题格式**

```
[Sprint-{N}] Story-{M}: <Brief Description>
```

**示例**：
```
[Sprint-0] Story-1: Unified Response Package
[Sprint-1] Story-2: JWT Token Authentication
```

---

### 3.2 PR 描述模板

```markdown
## Story 信息
- **Story**: Sprint-{N} Story-{M} - <Story Title>
- **优先级**: P0/P1/P2
- **工作量估算**: X 天
- **实际工作量**: Y 天

## 变更说明

### 新增
- [ ] `core/pkg/response` 包
- [ ] `Success()` 函数
- [ ] `Error()` 函数
- [ ] `List()` 函数（含分页）

### 修改
- [ ] 更新 `core/handlers/config.go` 使用新响应包

### 删除
- [ ] 删除旧的响应处理代码

## 测试

### 单元测试
- **覆盖率**: 95%
- **测试数量**: 15 个
- **测试结果**: ✅ All Passing

```bash
go test -v -race ./pkg/response
=== RUN   TestSuccess
--- PASS: TestSuccess (0.00s)
=== RUN   TestError
--- PASS: TestError (0.00s)
...
PASS
coverage: 95.0% of statements
```

### 集成测试
- [ ] 配置 API 测试通过
- [ ] 响应格式符合规范

## 验收标准检查
- [ ] 创建 `core/pkg/response` 包
- [ ] 实现 `Success()` 函数（成功响应）
- [ ] 实现 `Error()` 函数（错误响应）
- [ ] 实现 `List()` 函数（列表响应含分页）
- [ ] 编写单元测试（覆盖率 > 90%）
- [ ] 编写使用文档和示例

## Review 重点
- [ ] API 设计是否符合规范？
- [ ] 错误处理是否完善？
- [ ] 测试用例是否充分？
- [ ] 文档是否清晰？
- [ ] 性能是否有问题？

## 相关文档
- [API 设计规范](../../docs/standards/api-design.md)
- [编码规范](../../docs/standards/coding-standards.md)
- [Story 详情](../../docs/sprint-artifacts/sprint-{N}/stories.md)

---

**AI Agent Review**:  
@architect 请审查 API 设计和架构  
@dev 请检查代码实现和测试覆盖率  
@tea 请检查测试策略和边界情况
```

---

### 3.3 Code Review 清单

#### **功能审查**
- [ ] 代码实现符合 Story 验收标准
- [ ] 所有实现任务完成
- [ ] 功能逻辑正确，没有 Bug
- [ ] 边界条件处理完善

#### **代码质量**
- [ ] 符合编码规范（参考 `docs/standards/coding-standards.md`）
- [ ] 变量命名清晰（驼峰命名）
- [ ] 函数职责单一（SRP 原则）
- [ ] 没有硬编码（使用常量或配置）
- [ ] 错误处理完善（返回有意义的错误）
- [ ] 没有代码重复（DRY 原则）

#### **测试质量**
- [ ] 单元测试覆盖率 > 90%（P0）或 > 80%（P1）
- [ ] 测试用例覆盖正常和异常场景
- [ ] 测试命名清晰（`Test{Function}_{Scenario}`）
- [ ] 使用 Table-Driven Tests（多场景）
- [ ] Mock 外部依赖
- [ ] 测试独立，不依赖执行顺序

#### **文档质量**
- [ ] 公开函数有注释（描述、参数、返回值）
- [ ] README.md 包含使用示例
- [ ] 复杂逻辑有说明注释
- [ ] 更新相关文档（PRD、Epic、Architecture）

#### **性能和安全**
- [ ] 没有明显性能问题（大循环、内存泄漏）
- [ ] 输入验证完善
- [ ] 没有 SQL 注入、XSS 等安全隐患
- [ ] 敏感信息不输出到日志

#### **依赖和兼容性**
- [ ] 没有引入不必要的依赖
- [ ] 依赖版本明确（go.mod）
- [ ] 向后兼容（如果是 API 修改）

---

### 3.4 AI Agent Review

#### **使用方法**

在 PR 评论中使用 `@agent-name` 请求 AI 审查：

```markdown
@architect 请审查这个响应包的 API 设计：
- 是否符合 RESTful 规范？
- 结构体设计是否合理？
- 是否有改进建议？

@dev 请检查代码实现：
- 测试覆盖率是否足够？
- 是否有边界条件遗漏？
- 错误处理是否完善？

@tea 请审查测试策略：
- 测试用例是否充分？
- 是否需要增加集成测试？
- Mock 使用是否合理？
```

#### **Agent 职责分工**

| Agent | 审查重点 | 示例 |
|-------|---------|------|
| **@architect** | 架构设计、API 接口、模块划分 | API 设计是否合理？是否符合 SOLID 原则？ |
| **@dev** | 代码实现、算法逻辑、性能优化 | 代码质量如何？是否有重复代码？ |
| **@tea** | 测试策略、测试覆盖率、边界条件 | 测试是否充分？是否有遗漏场景？ |
| **@pm** | 功能完整性、用户体验 | 功能是否满足 PRD 要求？ |

---

### 3.5 Review 反馈处理

#### **处理流程**

```bash
# 1. 阅读 Review 意见
# 在 GitHub PR 页面查看评论

# 2. 逐条回复
# - 同意：✅ 感谢建议，已修复
# - 不同意：💬 说明原因，讨论方案

# 3. 修改代码
vim core/pkg/response/response.go

# 4. 运行测试
go test ./pkg/response -v

# 5. 提交修复
git add .
git commit -m "fix: address review comments

- Improve error message formatting
- Add nil pointer check in Success function
- Update test cases for edge conditions
"

# 6. 推送修复
git push origin sprint-0-story-1-response-package

# 7. 回复评论
# 在 PR 中回复：已修复，请重新审查
```

---

## 4. 测试流程

### 4.1 测试执行流程

```
开发时 → 本地测试 → PR 时 → CI 自动测试 → 合并时 → 集成测试 → 发布前 → E2E 测试
```

---

### 4.2 本地测试（开发时）

#### **Step 1: 单元测试**

```bash
# 运行单个包的测试
cd core
go test -v ./pkg/response

# 运行所有测试
go test -v ./...

# 运行测试并检测数据竞争
go test -v -race ./...

# 运行快速测试（跳过慢速测试）
go test -v -short ./...
```

#### **Step 2: 覆盖率检查**

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./pkg/response

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html

# 在浏览器中打开
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

**覆盖率目标**：
- P0 Story: ≥ 90%
- P1 Story: ≥ 80%
- P2 Story: ≥ 70%

#### **Step 3: 代码检查**

```bash
# 运行 golangci-lint
golangci-lint run ./pkg/response

# 自动修复（如果支持）
golangci-lint run --fix ./pkg/response

# 运行所有检查
golangci-lint run ./...
```

#### **Step 4: Ent Schema 检查**

```bash
# 检查 JSON tag 规范
./scripts/check-ent-json-tags.sh

# 重新生成 Ent 代码
cd core
go generate ./ent
```

---

### 4.3 CI 自动化测试（PR 时）

#### **触发条件**

```yaml
# .github/workflows/ci.yml

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
```

#### **测试 Job**

```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
      
      - name: Check Ent Schema JSON tags
        run: ./scripts/check-ent-json-tags.sh
  
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
  
  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:14
      redis:
        image: redis:7
    steps:
      - name: Run integration tests
        run: go test -v -tags=integration ./tests/integration/...
```

#### **检查状态**

```bash
# PR 中查看检查状态
✅ lint / golangci-lint
✅ lint / ent-check
✅ test / unit-tests
✅ test / integration-tests
✅ codecov / coverage (95%)
```

---

### 4.4 集成测试（合并时）

#### **测试范围**

- 数据库操作测试
- API 端到端测试
- 跨模块交互测试

#### **执行命令**

```bash
# 运行集成测试
cd tests/integration
./config/test-api.sh

# 或使用 Make
cd /data/cdl/apprun
make test-integration
```

---

### 4.5 E2E 测试（发布前）

#### **测试场景**

```bash
# 使用 Docker Compose 启动完整环境
cd tests/e2e
docker-compose up -d

# 运行 E2E 测试
go test -v ./scenarios/...

# 清理环境
docker-compose down
```

#### **测试覆盖**

- 用户注册、登录流程
- 项目创建、管理流程
- 文件上传、下载流程
- 函数部署、执行流程

---

### 4.6 测试失败处理

#### **本地测试失败**

```bash
# 1. 查看失败详情
go test -v ./pkg/response

# 2. 运行单个测试
go test -v -run TestSuccess ./pkg/response

# 3. 调试测试
go test -v -run TestSuccess ./pkg/response -test.v=true

# 4. 修复代码
vim response.go

# 5. 重新测试
go test -v ./pkg/response
```

#### **CI 测试失败**

```bash
# 1. 在 GitHub PR 页面查看失败日志

# 2. 本地复现
git pull origin sprint-0-story-1-response-package
go test -v -race ./...

# 3. 修复并推送
git commit -m "fix: resolve CI test failures"
git push origin sprint-0-story-1-response-package
```

---

## 5. 发布流程

### 5.1 发布准备

#### **Step 1: Sprint 完成检查**

```bash
# 检查所有 Story 是否完成
vim docs/sprint-artifacts/sprint-{N}/progress.md

# 所有 Story 状态应为 Completed
- [x] Story 1: Completed
- [x] Story 2: Completed
- [x] Story 3: Completed
```

#### **Step 2: 测试验证**

```bash
# 1. 运行所有单元测试
make test-unit

# 2. 运行集成测试
make test-integration

# 3. 运行 E2E 测试
make test-e2e

# 4. 检查覆盖率
# 目标：整体覆盖率 > 70%
```

#### **Step 3: 代码审查**

```bash
# 确保所有 PR 已合并到 develop
git checkout develop
git pull origin develop
git log --oneline -10
```

---

### 5.2 版本发布

#### **Step 1: 创建 Release PR**

```bash
# 1. 从 develop 创建 release 分支
git checkout develop
git pull origin develop
git checkout -b release-v1.0.0

# 2. 更新版本号
vim core/version.go
# const Version = "1.0.0"

# 3. 更新 CHANGELOG
vim CHANGELOG.md

# 4. 提交
git add .
git commit -m "chore(release): prepare v1.0.0 release

- Update version to 1.0.0
- Update CHANGELOG with Sprint 1-3 changes
"

# 5. 推送
git push origin release-v1.0.0

# 6. 创建 PR: release-v1.0.0 → main
```

#### **Step 2: 合并到 main**

```bash
# 在 GitHub 创建 PR
# 标题: [Release] v1.0.0
# 目标: main
# 合并方式: Merge Commit（保留完整历史）
```

#### **Step 3: 打标签**

```bash
# 1. 拉取最新 main
git checkout main
git pull origin main

# 2. 创建标签
git tag -a v1.0.0 -m "Release v1.0.0

Features:
- Authentication module (JWT, Session, RBAC)
- Storage module (Local, S3, File metadata)
- Functions module (Go function execution)

Sprint: Sprint 1-3
Coverage: 75%
"

# 3. 推送标签
git push origin v1.0.0
```

#### **Step 4: 创建 GitHub Release**

在 GitHub Release 页面：
- Tag: `v1.0.0`
- Title: `apprun v1.0.0`
- Description: 参考 CHANGELOG
- Assets: 构建产物（可选）

---

### 5.3 发布验证

#### **Step 1: 部署到测试环境**

```bash
# 使用 Docker 部署
docker build -t apprun:v1.0.0 .
docker-compose up -d
```

#### **Step 2: 冒烟测试**

```bash
# 检查服务健康
curl http://localhost:8080/health

# 测试关键接口
curl http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer test-token"
```

#### **Step 3: 回滚计划**

```bash
# 如果发现问题，回滚到上一个版本
git revert v1.0.0
git tag -a v1.0.1 -m "Rollback v1.0.0"
git push origin v1.0.1
```

---

### 5.4 发布后任务

#### **Step 1: 合并 main 到 develop**

```bash
# 保持 develop 和 main 同步
git checkout develop
git merge main
git push origin develop
```

#### **Step 2: 完成 Sprint 文档**

```bash
# 填写 Sprint Summary
vim docs/sprint-artifacts/sprint-{N}/summary.md

# 总结经验教训
# 记录 AI 协作效果
# 提出改进建议
```

#### **Step 3: 规划下一个 Sprint**

```bash
# 创建下一个 Sprint 文档
mkdir docs/sprint-artifacts/sprint-{N+1}
vim docs/sprint-artifacts/sprint-{N+1}/stories.md
```

---

## 6. 工具配置

### 6.1 golangci-lint

#### **安装**

```bash
# 方式 1: 使用脚本安装
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b $(go env GOPATH)/bin v1.64.8

# 方式 2: 使用 Homebrew (macOS)
brew install golangci-lint

# 方式 3: 使用 go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

#### **配置文件**

```yaml
# .golangci.yml

run:
  timeout: 5m
  tests: true
  skip-dirs:
    - ent
    - vendor

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - gocyclo
    - misspell

linters-settings:
  gocyclo:
    min-complexity: 15
```

#### **运行**

```bash
# 运行 Linter
golangci-lint run

# 自动修复
golangci-lint run --fix

# 指定目录
golangci-lint run ./pkg/response
```

---

### 6.2 Git Hooks

#### **安装 pre-commit**

```bash
# .git/hooks/pre-commit

#!/bin/bash

echo "Running pre-commit checks..."

# 1. Run tests
echo "Running tests..."
cd core
go test -short ./...
if [ $? -ne 0 ]; then
    echo "❌ Tests failed"
    exit 1
fi

# 2. Run linter
echo "Running linter..."
golangci-lint run ./...
if [ $? -ne 0 ]; then
    echo "❌ Linter failed"
    exit 1
fi

echo "✅ Pre-commit checks passed"
exit 0
```

```bash
# 添加执行权限
chmod +x .git/hooks/pre-commit
```

---

### 6.3 Makefile 快捷命令

```makefile
# Makefile

.PHONY: test lint dev commit-check

# 运行所有测试
test:
	cd core && go test -v -race ./...

# 运行 Linter
lint:
	golangci-lint run ./...

# 覆盖率报告
coverage:
	cd core && go test -coverprofile=coverage.out ./...
	cd core && go tool cover -html=coverage.out -o coverage.html

# 提交前检查
commit-check: test lint
	@echo "✅ Commit checks passed"
```

---

## 7. 常见问题

### 7.1 测试相关

**Q: 测试覆盖率不够怎么办？**

```bash
# 1. 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# 2. 查看未覆盖代码（红色部分）

# 3. 为未覆盖代码添加测试
vim pkg/response/response_test.go
```

---

**Q: 如何跳过慢速测试？**

```go
// 在测试中添加
func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping slow test in short mode")
    }
    // ... 测试逻辑
}
```

```bash
# 运行快速测试
go test -short ./...
```

---

### 7.2 Git 相关

**Q: 如何撤销上一次 commit？**

```bash
# 撤销 commit，保留修改
git reset --soft HEAD~1

# 撤销 commit 和修改
git reset --hard HEAD~1

# 修改 commit message
git commit --amend -m "new message"
```

---

**Q: 如何合并多个 commit？**

```bash
# 交互式 rebase（合并最近 3 个 commit）
git rebase -i HEAD~3

# 在编辑器中将 pick 改为 squash
pick abc123 first commit
squash def456 second commit
squash ghi789 third commit

# 保存并退出
```

---

**Q: 如何解决合并冲突？**

```bash
# 1. 查看冲突文件
git status

# 2. 编辑冲突文件
vim core/pkg/response/response.go

# 3. 标记为已解决
git add core/pkg/response/response.go

# 4. 完成合并
git commit -m "merge: resolve conflicts"
```

---

### 7.3 开发相关

**Q: 如何请求 AI Agent Review？**

在 PR 评论中：
```markdown
@architect 请审查 API 设计
@dev 请检查代码质量
@tea 请审查测试覆盖率
```

---

**Q: 如何快速创建新 Story？**

```bash
# 1. 从 stories.md 复制 Story 模板
vim docs/sprint-artifacts/sprint-{N}/stories.md

# 2. 填写 Story 信息
# 3. 创建功能分支
git checkout -b sprint-{N}-story-{M}-{description}
```

---

## 附录

### A. 快速参考

#### **开发循环**

```
准备 → TDD → 检查 → 提交 → 推送 → PR → Review → 合并
```

#### **TDD 循环**

```
Red (写测试) → Green (实现) → Refactor (优化) → Repeat
```

#### **常用命令**

```bash
# 测试
go test -v ./...
go test -v -race ./...
go test -coverprofile=coverage.out ./...

# Linter
golangci-lint run ./...

# Git
git checkout -b sprint-{N}-story-{M}-{desc}
git commit -m "feat(scope): description"
git push origin sprint-{N}-story-{M}-{desc}
```

---

### B. 相关文档

- [编码规范](./coding-standards.md)
- [API 设计规范](./api-design.md)
- [测试规范](./testing-standards.md)
- [Sprint Artifacts](../sprint-artifacts/)

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26  
**审核状态**: Active
