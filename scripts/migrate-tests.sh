#!/bin/bash

# 测试脚本迁移脚本
# 执行此脚本将现有测试脚本迁移到标准目录结构

set -e

echo "开始迁移测试脚本到标准目录..."

# 1. 创建目录结构
mkdir -p tests/{integration/config,e2e/scenarios,performance,scripts}

# 2. 移动现有测试脚本（如果存在）
if [ -f "core/test-api.sh" ]; then
    mv core/test-api.sh tests/integration/config/
    echo "✓ 移动 test-api.sh"
fi

if [ -f "core/test-priority.sh" ]; then
    mv core/test-priority.sh tests/integration/config/
    echo "✓ 移动 test-priority.sh"
fi

# 3. 创建共享函数文件（如果不存在）
if [ ! -f "tests/common.sh" ]; then
    cat > tests/common.sh << 'EOF'
#!/bin/bash
# 共享测试工具函数
export BASE_URL="${BASE_URL:-http://localhost:8080}"
export no_proxy=localhost,127.0.0.1

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 测试结果记录
TESTS_PASSED=0
TESTS_FAILED=0

# 断言函数
assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="$3"

    if [ "$expected" == "$actual" ]; then
        echo -e "${GREEN}✓${NC} $message"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗${NC} $message"
        echo "  Expected: $expected"
        echo "  Actual:   $actual"
        ((TESTS_FAILED++))
        return 1
    fi
}

# HTTP 请求包装
http_get() {
    curl -s -X GET "$BASE_URL$1"
}

http_put() {
    local path="$1"
    local data="$2"
    curl -s -X PUT "$BASE_URL$path" \
        -H "Content-Type: application/json" \
        -d "$data"
}

# 打印测试摘要
print_summary() {
    echo ""
    echo "======================================"
    echo "测试摘要"
    echo "======================================"
    echo -e "通过: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "失败: ${RED}$TESTS_FAILED${NC}"
    echo "总计: $((TESTS_PASSED + TESTS_FAILED))"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}所有测试通过！${NC}"
        return 0
    else
        echo -e "\n${RED}有测试失败！${NC}"
        return 1
    fi
}
EOF
    echo "✓ 创建 tests/common.sh"
fi

# 4. 创建辅助脚本
if [ ! -f "tests/scripts/setup-test-db.sh" ]; then
    cat > tests/scripts/setup-test-db.sh << 'EOF'
#!/bin/bash
# 测试环境准备脚本

set -e

echo "准备测试环境..."

cd "$(dirname "$0")/../../docker"
docker compose up -d

echo "等待服务启动..."
sleep 5

if curl -s http://localhost:8080/ | grep -q "apprun"; then
    echo "✓ 应用服务启动成功"
else
    echo "✗ 应用服务启动失败"
    exit 1
fi

echo "测试环境准备完成"
EOF
    echo "✓ 创建 tests/scripts/setup-test-db.sh"
fi

if [ ! -f "tests/scripts/cleanup.sh" ]; then
    cat > tests/scripts/cleanup.sh << 'EOF'
#!/bin/bash
# 测试数据清理脚本

echo "清理测试数据..."

cd "$(dirname "$0")/../../docker"
docker compose down

echo "测试环境清理完成"
EOF
    echo "✓ 创建 tests/scripts/cleanup.sh"
fi

# 5. 创建 Makefile（如果不存在）
if [ ! -f "Makefile" ]; then
    cat > Makefile << 'EOF'
# apprun Makefile

.PHONY: help build test test-all test-unit test-integration test-e2e clean docker-build docker-up docker-down

help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  test-all       - Run all tests"
	@echo "  test-unit      - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e       - Run end-to-end tests"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-up      - Start Docker services"
	@echo "  docker-down    - Stop Docker services"
	@echo "  clean          - Clean build artifacts"

build:
	cd core && go build -o bin/server ./cmd/server

test-all: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	cd core && go test -v -race -coverprofile=coverage.out ./...

test-integration:
	@echo "Running integration tests..."
	@tests/scripts/setup-test-db.sh
	@tests/integration/config/test-api.sh
	@tests/integration/config/test-priority.sh
	@tests/scripts/cleanup.sh

test-e2e:
	@echo "Running E2E tests..."
	@echo "E2E tests not implemented yet"

docker-build:
	cd docker && docker compose build

docker-up:
	cd docker && docker compose up -d

docker-down:
	cd docker && docker compose down

clean:
	cd core && rm -rf bin/ coverage.out coverage.html
	find . -name "*.log" -delete

dev: docker-up
	@echo "Development environment started"
	@echo "App: http://localhost:8080"
	@echo "Config API: http://localhost:8080/config"

test-config: test-unit
	@echo "Running config module tests..."
	@tests/scripts/setup-test-db.sh
	@tests/integration/config/test-api.sh
	@tests/scripts/cleanup.sh
EOF
    echo "✓ 创建 Makefile"
fi

# 6. 设置执行权限
chmod +x tests/integration/config/*.sh
chmod +x tests/scripts/*.sh

echo ""
echo "✓ 迁移完成！"
echo ""
echo "新目录结构："
echo "tests/"
echo "├── common.sh                    # 共享测试函数"
echo "├── integration/"
echo "│   └── config/"
echo "│       ├── test-api.sh         # API 功能测试"
echo "│       └── test-priority.sh    # 优先级测试"
echo "└── scripts/"
echo "    ├── setup-test-db.sh       # 测试环境准备"
echo "    └── cleanup.sh             # 清理脚本"
echo ""
echo "运行测试命令："
echo "  make test-all        # 运行所有测试"
echo "  make test-unit       # 运行单元测试"
echo "  make test-integration # 运行集成测试"
echo "  make test-config     # 运行配置模块测试"