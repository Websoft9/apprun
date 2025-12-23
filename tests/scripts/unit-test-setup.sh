#!/bin/bash
# tests/scripts/unit-test-setup.sh
# 单元测试环境准备脚本

set -e

echo "======================================"
echo "准备单元测试环境"
echo "======================================"

# 检查Go环境
echo "检查Go环境..."
if ! command -v go &> /dev/null; then
    echo "错误: Go未安装或不在PATH中"
    exit 1
fi

go_version=$(go version)
echo "✓ Go版本: $go_version"

# 检查测试依赖
echo ""
echo "检查测试依赖..."

# 检查testify
if ! go list -m github.com/stretchr/testify &> /dev/null; then
    echo "安装测试依赖..."
    cd core
    go mod tidy
    cd ..
fi

echo "✓ 测试依赖已准备"

# 创建测试输出目录
echo ""
echo "创建测试输出目录..."
mkdir -p core/test-results

echo "✓ 单元测试环境准备完成"

echo ""
echo "运行单元测试:"
echo "  tests/scripts/run-unit-tests.sh"
echo "  或: make test-unit"