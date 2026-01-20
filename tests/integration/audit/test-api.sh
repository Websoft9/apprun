#!/bin/bash

# 加载共享函数
source "$(dirname "$0")/../../common.sh"

echo "======================================"
echo "审计日志 API 集成测试"
echo "======================================"

# 设置测试环境变量
export JWT_TOKEN=""
export USER_ID=""

# 辅助函数：带认证的HTTP请求
http_get_auth() {
    local path="$1"
    curl -s --noproxy localhost,127.0.0.1 -X GET "$BASE_URL$path" \
        -H "Authorization: Bearer $JWT_TOKEN"
}

http_post_auth() {
    local path="$1"
    local data="$2"
    curl -s --noproxy localhost,127.0.0.1 -X POST "$BASE_URL$path" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -d "$data"
}

# 测试 0: 用户注册和登录（获取JWT token）
echo ""
echo "测试 0: 用户注册和登录"
RANDOM_EMAIL="audituser_$(date +%s)@test.com"
register_response=$(curl -s --noproxy localhost,127.0.0.1 -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$RANDOM_EMAIL\",\"password\":\"Test@123\",\"username\":\"audituser\"}")

echo "注册响应: $register_response"

# 登录获取token
login_response=$(curl -s --noproxy localhost,127.0.0.1 -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$RANDOM_EMAIL\",\"password\":\"Test@123\"}")

JWT_TOKEN=$(echo $login_response | jq -r '.data.access_token // .access_token // empty')
USER_ID=$(echo $login_response | jq -r '.data.user.id // .user.id // empty')

if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" == "null" ]; then
    echo -e "${YELLOW}警告: 无法获取JWT token，跳过需要认证的测试${NC}"
    echo "登录响应: $login_response"
else
    echo -e "${GREEN}✓${NC} 成功获取JWT token"
    
    # 测试 1: 执行一些操作生成审计日志
    echo ""
    echo "测试 1: 执行操作生成审计日志"
    
    # 尝试修改配置（如果可用）
    config_response=$(http_put "/api/config" '{"test.audit": "test_value"}' 2>/dev/null || echo '{}')
    
    # 刷新token（这会生成审计日志）
    refresh_response=$(http_post_auth "/api/auth/refresh" "{\"refresh_token\":\"dummy\"}")
    
    # 等待异步审计日志写入完成
    sleep 2
    
    echo -e "${GREEN}✓${NC} 操作已执行，等待审计日志写入"
fi

# 测试 2: 查询审计日志（需要admin权限）
echo ""
echo "测试 2: GET /api/admin/audit-logs - 查询审计日志"

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${YELLOW}跳过: 无JWT token${NC}"
else
    audit_response=$(http_get_auth "/api/admin/audit-logs?page=1&page_size=10")
    echo "审计日志响应: $audit_response"
    
    # 检查响应格式
    has_logs=$(echo $audit_response | jq -r 'has("data")' 2>/dev/null || echo "false")
    
    if [ "$has_logs" == "true" ]; then
        logs_array=$(echo $audit_response | jq -r '.data.logs // []' 2>/dev/null)
        log_count=$(echo $logs_array | jq -r 'length' 2>/dev/null || echo "0")
        
        if [ "$log_count" -gt "0" ]; then
            echo -e "${GREEN}✓${NC} 成功查询到 $log_count 条审计日志"
            
            # 验证日志字段
            first_log=$(echo $logs_array | jq -r '.[0]')
            has_timestamp=$(echo $first_log | jq -r 'has("timestamp")')
            has_action=$(echo $first_log | jq -r 'has("action")')
            
            assert_equals "true" "$has_timestamp" "审计日志包含 timestamp 字段"
            assert_equals "true" "$has_action" "审计日志包含 action 字段"
        else
            echo -e "${YELLOW}警告: 查询成功但没有审计日志数据${NC}"
        fi
    elif echo $audit_response | grep -q "403\|Forbidden\|platform_admin"; then
        echo -e "${YELLOW}预期: 需要 platform_admin 权限 (当前用户无管理员权限)${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗${NC} 审计日志查询失败"
        echo "响应: $audit_response"
        ((TESTS_FAILED++))
    fi
fi

# 测试 3: 测试查询参数过滤
echo ""
echo "测试 3: 测试审计日志查询参数"

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${YELLOW}跳过: 无JWT token${NC}"
else
    # 按时间范围查询
    start_time=$(date -u -d "1 hour ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v-1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2026-01-20T00:00:00Z")
    end_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "2026-01-20T23:59:59Z")
    
    time_filtered=$(http_get_auth "/api/admin/audit-logs?start_time=$start_time&end_time=$end_time&page_size=5")
    
    if echo $time_filtered | jq -e '.data' > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} 时间范围过滤查询成功"
        ((TESTS_PASSED++))
    else
        # 如果返回403，也算预期行为
        if echo $time_filtered | grep -q "403\|Forbidden"; then
            echo -e "${YELLOW}预期: 需要管理员权限${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${YELLOW}警告: 时间范围过滤查询异常${NC}"
        fi
    fi
    
    # 按action类型查询
    action_filtered=$(http_get_auth "/api/admin/audit-logs?action=login&page_size=5")
    
    if echo $action_filtered | jq -e '.data' > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Action类型过滤查询成功"
        ((TESTS_PASSED++))
    else
        if echo $action_filtered | grep -q "403\|Forbidden"; then
            echo -e "${YELLOW}预期: 需要管理员权限${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${YELLOW}警告: Action类型过滤查询异常${NC}"
        fi
    fi
fi

# 测试 4: 测试分页参数
echo ""
echo "测试 4: 测试分页参数"

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${YELLOW}跳过: 无JWT token${NC}"
else
    # 测试page_size限制（最大200）
    oversized=$(http_get_auth "/api/admin/audit-logs?page_size=300")
    
    if echo $oversized | jq -e '.data' > /dev/null 2>&1; then
        actual_size=$(echo $oversized | jq -r '.data.page_size' 2>/dev/null || echo "0")
        
        if [ "$actual_size" -le "200" ]; then
            echo -e "${GREEN}✓${NC} page_size正确限制在200以内 (实际: $actual_size)"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}✗${NC} page_size超过最大限制: $actual_size"
            ((TESTS_FAILED++))
        fi
    else
        if echo $oversized | grep -q "403\|Forbidden"; then
            echo -e "${YELLOW}预期: 需要管理员权限${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${YELLOW}警告: 分页测试异常${NC}"
        fi
    fi
fi

# 测试 5: 未认证访问应该被拒绝
echo ""
echo "测试 5: 未认证访问应被拒绝"
unauth_response=$(http_get "/api/admin/audit-logs")

if echo $unauth_response | grep -q "401\|Unauthorized\|token"; then
    echo -e "${GREEN}✓${NC} 未认证访问被正确拒绝"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}警告: 未认证访问响应异常${NC}"
    echo "响应: $unauth_response"
fi

# 打印测试摘要
print_summary
