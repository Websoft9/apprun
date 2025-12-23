#!/bin/bash

# 共享的测试辅助函数
export BASE_URL="${BASE_URL:-http://localhost:8080}"
export no_proxy=localhost,127.0.0.1

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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
    curl -s --noproxy localhost,127.0.0.1 -X GET "$BASE_URL$1"
}

http_put() {
    local path="$1"
    local data="$2"
    curl -s --noproxy localhost,127.0.0.1 -X PUT "$BASE_URL$path" \
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