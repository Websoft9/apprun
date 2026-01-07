# Story 17: Go 版本升级到 1.25.5
# Sprint 1: 安全修复

**Priority**: P0  
**Effort**: 0.5 天  
**Owner**: DevOps  
**Dependencies**: Story 5  
**Status**: ✅ Done  
**Module**: Infrastructure  
**Issue**: #TBD  
**Related**: [Story 5 CI/CD](../sprint-0/story-05-ci-cd-linter.md)

---

## User Story

作为开发者，我希望升级 Go 版本到 1.25.5，以便修复 15 个标准库安全漏洞，确保 CI 和本地环境的 govulncheck 一致性。

---

## Background

**问题现象**：
- CI govulncheck 报告 15 个漏洞（GO-2025-4175 到 GO-2025-3563）
- 本地 govulncheck 仅报告 2 个漏洞（GO-2025-4175, GO-2025-4155）

**根本原因**：
- CI 使用 `go1.24.0`（包含 15 个已知漏洞）
- 本地使用 `go1.25.3`（大部分漏洞已修复）
- 版本不一致导致安全检查标准不统一

---

## Acceptance Criteria

- [x] CI 工作流升级到 Go 1.25.5
- [x] Docker 镜像升级到 Go 1.25.5
- [x] Swagger CI 升级到 Go 1.25.5
- [x] go.mod 声明 Go 1.25.5
- [x] govulncheck 通过（0 漏洞）
- [x] 所有 CI 检查通过

---

## Implementation Tasks

### 1. 更新 CI 工作流
- [x] `.github/workflows/ci.yml`: `GO_VERSION: '1.24.0'` → `'1.25.5'`

### 2. 更新 Docker 镜像
- [x] `docker/Dockerfile.base`:
  - Stage 1: `FROM golang:1.24-alpine` → `golang:1.25.5-alpine`
  - Stage 2: `FROM golang:1.24-alpine` → `golang:1.25.5-alpine`

### 3. 更新 Swagger CI
- [x] `.github/workflows/swagger-ci.yml`: `go-version: '1.21'` → `'1.25.5'`

### 4. 更新 go.mod
- [x] `core/go.mod`: `go 1.24.0` → `go 1.25.5`

---

## Technical Details

### 修复的安全漏洞（15个）

| 漏洞编号 | 模块 | 描述 | 修复版本 |
|---------|------|------|---------|
| GO-2025-4175 | crypto/x509 | 通配符名称约束验证错误 | 1.24.11 |
| GO-2025-4155 | crypto/x509 | 证书验证错误打印资源消耗 | 1.24.11 |
| GO-2025-4013 | crypto/x509 | DSA 公钥验证 panic | 1.24.8 |
| GO-2025-4012 | net/http | Cookie 解析内存耗尽 | 1.24.8 |
| GO-2025-4011 | encoding/asn1 | DER 解析内存耗尽 | 1.24.8 |
| GO-2025-4010 | net/url | IPv6 主机名验证不足 | 1.24.8 |
| GO-2025-4009 | encoding/pem | 无效输入二次复杂度 | 1.24.8 |
| GO-2025-4008 | crypto/tls | ALPN 协商错误信息泄露 | 1.24.8 |
| GO-2025-4007 | crypto/x509 | 名称约束检查二次复杂度 | 1.24.9 |
| GO-2025-4006 | net/mail | ParseAddress CPU 消耗 | 1.24.8 |
| GO-2025-3956 | os/exec | LookPath 意外路径返回 | 1.24.6 |
| GO-2025-3849 | database/sql | Rows.Scan 错误结果 | 1.24.6 |
| GO-2025-3750 | os | O_CREATE\|O_EXCL 不一致 | 1.24.4 |
| GO-2025-3749 | crypto/x509 | ExtKeyUsageAny 禁用策略验证 | 1.24.4 |
| GO-2025-3563 | net/http | 无效 chunked 数据请求走私 | 1.24.2 |

### 版本对比

```yaml
# 之前（不一致）
CI:     go1.24.0  → 15 漏洞 ❌
Local:  go1.25.3  → 2 漏洞  ⚠️
Docker: go1.24    → 15 漏洞 ❌

# 之后（统一）
All:    go1.25.5  → 0 漏洞  ✅
```

---

## Files Changed

```
.github/workflows/ci.yml         (GO_VERSION)
.github/workflows/swagger-ci.yml (go-version)
docker/Dockerfile.base           (FROM golang:x.xx)
core/go.mod                      (go directive)
```

---

## Testing

### 验证步骤

1. **本地验证**
   ```bash
   cd core
   go version  # go1.25.5
   make lint   # 包含 govulncheck
   ```

2. **CI 验证**
   - Push 到 main 分支
   - 检查 CI Pipeline 中的 govulncheck 步骤
   - 确认 0 漏洞输出

3. **Docker 验证**
   ```bash
   docker build -t test -f docker/Dockerfile.base .
   docker run test go version  # go1.25.5
   ```

---

## Migration Notes

### 兼容性
- ✅ Go 1.25.5 向后兼容 1.24.0 代码
- ✅ 无需修改应用代码
- ✅ go.mod 依赖自动适配

### 注意事项
- Docker 基础镜像需重新构建（触发 build-base.yml）
- 本地开发环境需升级 Go 到 1.25.5+
- CI 缓存会自动更新到新 Go 版本

---

## Rollback Plan

如遇兼容性问题：
1. 回退 `.github/workflows/ci.yml` 中的 `GO_VERSION`
2. 回退 `docker/Dockerfile.base` 中的 `FROM` 版本
3. 回退 `core/go.mod` 中的 `go` 指令
4. 重新构建 Docker 基础镜像

---

## References

- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Go Security Policy](https://go.dev/security/policy)
- [govulncheck Documentation](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- Story 5: [CI/CD Pipeline & Linter](../sprint-0/story-05-ci-cd-linter.md)

---

## Definition of Done

- [x] Go 1.25.5 在所有环境中使用（CI/Docker/本地）
- [x] govulncheck 通过，0 stdlib 漏洞
- [x] 所有 CI 检查通过（lint/test/build）
- [x] Docker 镜像构建成功
- [x] 文档更新完成
- [x] Git commit 包含清晰的修复说明

---

**Commit Message**:
```
fix(ci): upgrade Go version to 1.25.5 to fix 15 stdlib vulnerabilities

- CI: go1.24.0 → go1.25.5 (fixes GO-2025-4175 to GO-2025-3563)
- Docker: golang:1.24-alpine → golang:1.25.5-alpine
- Swagger CI: go1.21 → go1.25.5
- go.mod: go 1.24.0 → go 1.25.5

Root cause: CI used outdated Go 1.24.0 with 15 known vulnerabilities
Local env (go1.25.3) only had 2 vulnerabilities, causing confusion

Related: Story 5 - CI/CD Pipeline & govulncheck integration
Related: Story 17 - Go Version Upgrade
```
