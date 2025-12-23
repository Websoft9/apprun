#!/bin/bash
# tests/scripts/run-unit-tests.sh
# 运行所有单元测试的脚本

set -e

echo "======================================"
echo "运行单元测试"
echo "======================================"

# 切换到core目录
cd core

# 运行单元测试并生成覆盖率报告
echo "运行单元测试..."
go test -v -race -coverprofile=coverage.out ./...

# 显示覆盖率摘要
echo ""
echo "覆盖率摘要:"
go tool cover -func=coverage.out

# 生成HTML覆盖率报告
echo ""
echo "生成HTML覆盖率报告..."
go tool cover -html=coverage.out -o coverage.html
echo "覆盖率报告已生成: core/coverage.html"

echo ""
echo "单元测试完成！"