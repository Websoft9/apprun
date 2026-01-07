# CI/CD 和代码质量工具使用指南

## 概述

本项目使用 GitHub Actions 进行持续集成（CI），使用 golangci-lint 进行代码质量检查。

## 本地开发

### 安装 golangci-lint

```bash
# macOS/Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 或使用 go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 运行 Linter

```bash
# 在项目根目录运行
cd core
golangci-lint run --config=.golangci.yml

# 自动修复部分问题
golangci-lint run --config=.golangci.yml --fix
```

### 运行测试

```bash
# 运行所有测试
cd core
go test ./...

# 运行测试并生成覆盖率报告
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## CI Pipeline

### 工作流文件

- `.github/workflows/ci.yml` - 主 CI 流水线
- `.github/workflows/docker-build.yml` - Docker 构建和发布

### CI 流程

CI 流水线包含以下步骤：

1. **Lint** - 代码质量检查
   - 运行 golangci-lint
   - 检查代码格式、潜在错误和最佳实践

2. **Test** - 单元测试
   - 运行所有单元测试
   - 生成覆盖率报告
   - 检查覆盖率阈值（> 70%）

3. **Build** - 构建验证
   - 编译 Go 二进制文件
   - 验证编译成功

4. **Docker Build** - Docker 镜像构建
   - 构建 Docker 镜像（仅测试，不推送）
   - 使用缓存加速构建

5. **Integration Test** - 集成测试（仅 PR）
   - 运行集成测试（如果存在）

### 触发条件

- **Push** 到 `main` 或 `develop` 分支
- **Pull Request** 到 `main` 或 `develop` 分支

### 状态徽章

可以在 README 中添加 CI 状态徽章：

```markdown
![CI](https://github.com/YOUR_ORG/apprun/workflows/CI%20Pipeline/badge.svg)
```

## Linter 配置

### 启用的 Linters

- `errcheck` - 检查未检查的错误
- `gosimple` - 简化代码建议
- `govet` - 静态分析
- `ineffassign` - 检测无效赋值
- `staticcheck` - 高级静态检查
- `unused` - 未使用的代码检测
- `gofmt` - 代码格式化
- `goimports` - import 格式化
- `revive` - Go 代码质量检查
- `goconst` - 常量提取建议
- `gocyclo` - 圈复杂度检查
- `dupl` - 重复代码检测
- `gosec` - 安全检查

### 排除规则

- 测试文件（`*_test.go`）排除部分检查
- `ent/` 生成的代码排除所有检查
- `docs/` 文档排除所有检查

## 常见问题

### Q: 如何修复 linter 错误？

A: 大多数错误可以通过以下方式修复：

```bash
# 自动修复格式问题
golangci-lint run --fix

# 或手动修复代码
```

### Q: 如何临时禁用某个 linter？

A: 可以在代码中使用注释：

```go
//nolint:errcheck
result := someFunction()

//nolint:all
func problematicFunction() {
    // ...
}
```

### Q: CI 失败了怎么办？

A: 检查 GitHub Actions 日志：

1. 进入 GitHub 仓库的 Actions 标签页
2. 点击失败的工作流
3. 查看详细日志
4. 根据错误信息修复问题

### Q: 测试覆盖率不够怎么办？

A: 添加更多单元测试：

```go
func TestYourFunction(t *testing.T) {
    // 测试代码
}
```

## 最佳实践

1. **提交前运行 lint**
   ```bash
   make lint  # 如果 Makefile 中有定义
   # 或
   cd core && golangci-lint run --config=.golangci.yml
   ```

2. **提交前运行测试**
   ```bash
   make test  # 如果 Makefile 中有定义
   # 或
   cd core && go test ./...
   ```

3. **保持高测试覆盖率**
   - 为新功能编写单元测试
   - 目标覆盖率 > 80%

4. **遵循代码规范**
   - 使用 `gofmt` 格式化代码
   - 为导出的函数添加注释
   - 处理所有错误

## 相关文档

- [golangci-lint 文档](https://golangci-lint.run/)
- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [编码规范](../docs/standards/coding-standards.md)

---

**最后更新**: 2026-01-06
