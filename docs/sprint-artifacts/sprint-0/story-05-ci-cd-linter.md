# Story 5: CI/CD 流水线与 Linter
# Sprint 0: Infrastructure建设

**Priority**: P0  
**Effort**: 1 天  
**Owner**: DevOps  
**Dependencies**: Story 1  
**Status**: Planning  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [编码规范](../../standards/coding-standards.md)

---

## User Story

作为开发者，我希望有 CI/CD 流水线和代码质量检查，以便自动化测试和代码规范检查。

---

## Acceptance Criteria

- [ ] 配置 GitHub Actions 工作流
- [ ] 集成 golangci-lint
- [ ] 配置 lint 规则（.golangci.yml）
- [ ] 设置自动化测试流程
- [ ] 配置代码覆盖率报告
- [ ] 设置 PR 必须通过检查
- [ ] 编写 CI/CD 文档

---

## Implementation Tasks

- [ ] 创建 `.github/workflows/ci.yml`
- [ ] 创建 `.golangci.yml` 配置
- [ ] 配置 golangci-lint 规则
- [ ] 配置单元测试步骤
- [ ] 配置代码覆盖率检查（80%）
- [ ] 配置 Docker 镜像构建测试
- [ ] 更新 CONTRIBUTING.md

---

## Technical Details

### GitHub Actions 工作流

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          working-directory: core

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run tests
        run: |
          cd core
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      - name: Build Docker image
        run: docker build -f docker/Dockerfile -t apprun:test .
```

### golangci-lint 配置

```yaml
# .golangci.yml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign

linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/yourusername/apprun

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

---

## Test Cases

- [ ] Lint 检查通过
- [ ] 单元测试通过
- [ ] 代码覆盖率 > 80%
- [ ] Docker 镜像构建成功
- [ ] PR 必须通过所有检查

---

## Related Docs

- [编码规范](../../standards/coding-standards.md)
- [golangci-lint 文档](https://golangci-lint.run/)

---

**Created**: 2025-12-27  
**Updated**: 2025-12-27  
**Maintainer**: Architect Agent
