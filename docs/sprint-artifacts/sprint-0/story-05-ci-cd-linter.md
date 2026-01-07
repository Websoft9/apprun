# Story 5: CI/CD 流水线与 Linter
# Sprint 0: Infrastructure建设

**Priority**: P0  
**Effort**: 2 天  
**Owner**: DevOps  
**Dependencies**: Story 1  
**Status**: ✅ Done  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [编码规范](../../standards/coding-standards.md), [CI/CD Guide](../../ci-cd-guide.md)

---

## User Story

作为开发者，我希望有完整的 CI/CD 流水线和安全检查，以便自动化测试、代码规范检查和容器镜像发布。

---

## Acceptance Criteria

- [x] 配置 GitHub Actions 工作流（ci.yml）
- [x] 集成 golangci-lint（含 gosec 安全检查）
- [x] 配置 lint 规则（core/.golangci.yml）
- [x] 设置自动化测试流程
- [x] 配置代码覆盖率报告（70% 阈值）
- [x] 添加依赖漏洞扫描（govulncheck）
- [x] 添加容器镜像安全扫描（Trivy）
- [x] 设置多架构镜像构建（amd64/arm64）
- [x] 配置 CI 触发 Docker 构建工作流
- [x] 设置 PR 必须通过检查
- [x] 编写 CI/CD 文档

---

## Implementation Tasks

- [x] 创建 `.github/workflows/ci.yml`（主 CI 流水线）
- [x] 创建 `.github/workflows/docker-build.yml`（镜像构建发布）
- [x] 创建 `.github/workflows/build-base.yml`（基础镜像）
- [x] 移动 `.golangci.yml` 到 `core/` 目录
- [x] 配置 golangci-lint 规则（含 gosec）
- [x] 配置单元测试步骤
- [x] 配置代码覆盖率检查（70%）
- [x] 添加 govulncheck 依赖扫描
- [x] 添加 Trivy 镜像安全扫描
- [x] 配置多平台 Docker 构建
- [x] 配置 CI 联动 Docker 构建
- [x] 更新 CONTRIBUTING.md
- [x] 编写 CI/CD Guide 文档

---

## Architecture Overview

### CI/CD 工作流架构

```
┌─────────────────────────────────────────────────────────────┐
│                       CI Pipeline (ci.yml)                   │
│  触发：Push/PR to main/develop                               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │   Lint   │  │   Test   │  │  Build   │  │  Docker  │   │
│  │          │  │          │  │          │  │  Build   │   │
│  │ gosec ✓  │  │ govuln ✓ │  │  Binary  │  │  Trivy ✓ │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│       │             │              │              │         │
│       └─────────────┴──────────────┴──────────────┘         │
│                          │                                  │
│                     ┌────▼────┐                            │
│                     │ Summary │                            │
│                     └────┬────┘                            │
│                          │                                  │
└──────────────────────────┼──────────────────────────────────┘
                           │ (CI Success on main/develop)
                           ▼
┌─────────────────────────────────────────────────────────────┐
│           Docker Build & Publish (docker-build.yml)         │
│  触发：CI Success / Manual / Tag                             │
├─────────────────────────────────────────────────────────────┤
│  • 多架构构建 (amd64/arm64)                                  │
│  • 推送到 GHCR                                               │
│  • 生成版本标签                                              │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│           Build Base Image (build-base.yml)                 │
│  触发：Weekly / go.mod 变更 / Manual                         │
├─────────────────────────────────────────────────────────────┤
│  • 独立运行，不参与 CI 联动                                   │
│  • 预构建 Go 依赖缓存                                         │
│  • 加速主镜像构建                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Technical Details

### 1. CI Pipeline (`.github/workflows/ci.yml`)

主 CI 流水线，包含 4 个并行/串行 jobs：

#### Job 1: Lint（代码质量检查）
```yaml
lint:
  runs-on: ubuntu-latest
  steps:
    - uses: golangci/golangci-lint-action@v4
      with:
        working-directory: core
        args: --timeout=5m --config=.golangci.yml
```

**检查项**：
- `gosec` - 安全漏洞检测（硬编码密钥、SQL 注入等）
- `errcheck` - 未检查错误
- `govet` - 静态分析
- `staticcheck` - 高级静态检查
- `unused` - 未使用代码
- 其他 10+ linters

#### Job 2: Test（测试与漏洞扫描）
```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - name: Run govulncheck
      run: |
        go install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck ./...
    
    - name: Run tests with coverage
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Check coverage threshold (70%)
      run: |
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$COVERAGE < 70" | bc -l) )); then
          exit 1
        fi
```

**检查项**：
- `govulncheck` - Go 依赖漏洞扫描（CVE 数据库）
- 单元测试（race 检测）
- 覆盖率检查（≥ 70%）
- Codecov 上传

#### Job 3: Build（二进制构建）
```yaml
build:
  needs: [lint, test]
  steps:
    - run: go build -o bin/apprun-core ./cmd/server
```

#### Job 4: Docker Build（镜像构建与扫描）
```yaml
docker-build:
  needs: [lint, test]
  steps:
    - uses: docker/build-push-action@v5
      with:
        platforms: linux/amd64,linux/arm64
        tags: apprun:ci-test
    
    - uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'image'
        image-ref: 'apprun:ci-test'
        format: 'sarif'
    
    - uses: github/codeql-action/upload-sarif@v3
```

**检查项**：
- 多架构构建验证
- Trivy 镜像安全扫描
- 扫描结果上传到 GitHub Security

#### Job 5: Summary（汇总与触发）
```yaml
summary:
  needs: [lint, test, build, docker-build]
  if: always()
  steps:
    - name: Check job status
    - name: Trigger Docker Build workflow
      if: |
        needs.*.result == 'success' &&
        github.event_name == 'push' &&
        (github.ref == 'refs/heads/main' || github.ref == 'refs/heads/develop')
      uses: actions/github-script@v7
```

**功能**：
- 汇总所有 job 状态
- CI 成功后，在 main/develop 分支自动触发 `docker-build.yml`

---

### 2. Docker Build Workflow (`.github/workflows/docker-build.yml`)

生产镜像构建与发布：

```yaml
on:
  workflow_dispatch:  # 由 CI 触发
  push:
    tags: ['v*']     # 版本标签发布
  pull_request:      # PR 测试

jobs:
  build-and-push:
    steps:
      - name: Extract metadata
        uses: docker/metadata-action@v5
        with:
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=sha,prefix={{branch}}-
            type=raw,value=latest,enable={{is_default_branch}}
      
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          platforms: linux/amd64,linux/arm64
          push: ${{ github.event_name != 'pull_request' }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

**特性**：
- 由 CI 成功后自动触发（main/develop）
- 多架构支持（amd64/arm64）
- 自动生成版本标签
- 推送到 GHCR
- GitHub Actions 缓存加速

---

### 3. Base Image Build (`.github/workflows/build-base.yml`)

基础镜像预构建（独立运行）：

```yaml
on:
  schedule:
    - cron: '0 2 * * 0'  # 每周日构建
  push:
    paths:
      - 'core/go.mod'
      - 'core/go.sum'
  workflow_dispatch:     # 手动触发
```

**特性**：
- 独立运行，不参与 CI 联动
- 预构建 Go 依赖缓存
- 智能重建检测（7 天或依赖变更）
- 加速主镜像构建 20-30 秒

---

### 4. golangci-lint 配置 (`core/.golangci.yml`)

```yaml
linters:
  enable:
    - errcheck      # 未检查错误
    - govet         # 静态分析
    - staticcheck   # 高级检查
    - gosec         # 安全检查 ✓
    - unused        # 未使用代码
    - goconst       # 常量提取
    - misspell      # 拼写检查
    - dupl          # 重复代码
    - unconvert     # 不必要类型转换
    - gocognit      # 认知复杂度
    - gocyclo       # 圈复杂度
    - gocritic      # 综合诊断
    - errorlint     # 错误包装
    - bodyclose     # HTTP body 关闭

linters-settings:
  gosec:
    excludes:
      - G104  # errcheck 已覆盖
      - G302  # 文件权限（测试文件可接受）
  
  gocognit:
    min-complexity: 40
  
  gocyclo:
    min-complexity: 15

issues:
  exclude-dirs:
    - ent      # 生成代码
    - docs     # 文档
    - tests/e2e
  
  exclude-rules:
    - path: _test\.go
      linters: [dupl, gosec, goconst]
    - path: 'cmd/server/main\.go'
      text: 'exitAfterDefer'
      linters: [gocritic]
```

---

## Security Enhancements

### 代码层安全检查
1. **gosec**（静态代码分析）
   - 检测：硬编码密钥、SQL 注入、弱加密、不安全文件操作
   - 集成：golangci-lint 中启用
   - 运行：每次 lint job

2. **govulncheck**（依赖漏洞扫描）
   - 检测：第三方库已知 CVE
   - 数据库：Go 官方漏洞数据库
   - 运行：每次 test job

### 容器层安全检查
3. **Trivy**（镜像漏洞扫描）
   - 检测：OS 包漏洞、应用依赖漏洞、配置问题
   - 格式：SARIF（上传到 GitHub Security）
   - 运行：每次 docker-build job

---

## Workflow Triggers

### CI Pipeline (`ci.yml`)
- **Push** 到 `main` 或 `develop`
- **Pull Request** 到 `main` 或 `develop`
- 成功后自动触发 `docker-build.yml`

### Docker Build (`docker-build.yml`)
- **Workflow Dispatch**（由 CI 触发）
- **Push** 带 `v*` 标签（版本发布）
- **Pull Request**（测试构建）

### Base Image (`build-base.yml`)
- **Schedule**（每周日 2 AM UTC）
- **Push** 修改 `go.mod`/`go.sum`
- **Manual**（workflow_dispatch）

---

## Test Cases

- [x] Lint 检查通过（含 gosec）
- [x] govulncheck 依赖扫描通过
- [x] 单元测试通过
- [x] 代码覆盖率 ≥ 70%
- [x] Docker 镜像构建成功（多架构）
- [x] Trivy 镜像扫描通过
- [x] CI 成功触发 Docker Build
- [x] PR 必须通过所有检查
- [x] 状态上传到 GitHub Security

---

## Performance Metrics

- **Lint Job**: ~2-3 分钟
- **Test Job**: ~3-4 分钟（含 govulncheck）
- **Build Job**: ~1-2 分钟
- **Docker Build Job**: ~4-5 分钟（多架构 + Trivy）
- **总耗时**: ~10-15 分钟（并行执行）

使用 base 镜像可节省 20-30 秒构建时间。

---

## Related Docs

- [CI/CD 使用指南](../../ci-cd-guide.md)
- [编码规范](../../standards/coding-standards.md)
- [DevOps 流程](../../standards/devops-process.md)
- [golangci-lint 文档](https://golangci-lint.run/)
- [Trivy 文档](https://trivy.dev/)

---

**Created**: 2025-12-27  
**Updated**: 2026-01-07  
**Maintainer**: Architect Agent
