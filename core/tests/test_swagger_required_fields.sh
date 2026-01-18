#!/bin/bash
# Test Story 18 Swagger Required Fields
# Tests that email and password are properly marked as required

set -e

BASE_URL="http://localhost:8080"
TIMESTAMP=$(date +%s)

echo "🧪 Story 18 - Swagger 必填字段测试"
echo "=================================="
echo ""

# Test 1: Missing email field
echo "📋 测试 1: 缺少 email 字段 (应返回 400)"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"password":"SecurePass123"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "400" ]; then
    echo "✅ HTTP 400 - 正确拒绝缺少 email 的请求"
    echo "   响应: $(echo "$BODY" | jq -c .)"
else
    echo "❌ 预期 HTTP 400，实际: $HTTP_CODE"
    echo "   响应: $BODY"
    exit 1
fi
echo ""

# Test 2: Missing password field
echo "📋 测试 2: 缺少 password 字段 (应返回 400)"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "400" ]; then
    echo "✅ HTTP 400 - 正确拒绝缺少 password 的请求"
    echo "   响应: $(echo "$BODY" | jq -c .)"
else
    echo "❌ 预期 HTTP 400，实际: $HTTP_CODE"
    echo "   响应: $BODY"
    exit 1
fi
echo ""

# Test 3: Missing both fields
echo "📋 测试 3: 缺少所有必填字段 (应返回 400)"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "400" ]; then
    echo "✅ HTTP 400 - 正确拒绝空请求"
    echo "   响应: $(echo "$BODY" | jq -c .)"
else
    echo "❌ 预期 HTTP 400，实际: $HTTP_CODE"
    echo "   响应: $BODY"
    exit 1
fi
echo ""

# Test 4: Valid request with required fields only
echo "📋 测试 4: 仅包含必填字段的有效请求 (应返回 201)"
TEST_EMAIL="required_test_${TIMESTAMP}@example.com"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"SecurePass123\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "201" ]; then
    echo "✅ HTTP 201 - 成功注册 (仅必填字段)"
    echo "   响应: $(echo "$BODY" | jq -c .data)"
else
    echo "❌ 预期 HTTP 201，实际: $HTTP_CODE"
    echo "   响应: $BODY"
    exit 1
fi
echo ""

# Test 5: Valid request with all optional fields
echo "📋 测试 5: 包含所有可选字段的完整请求 (应返回 201)"
TEST_EMAIL_2="full_test_${TIMESTAMP}@example.com"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\":\"$TEST_EMAIL_2\",
    \"password\":\"SecurePass123\",
    \"username\":\"testuser_$TIMESTAMP\",
    \"nickname\":\"Test User\",
    \"phone\":\"+86-13800138000\",
    \"gender\":1,
    \"timezone\":\"Asia/Shanghai\",
    \"language\":\"zh-CN\"
  }")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "201" ]; then
    echo "✅ HTTP 201 - 成功注册 (包含所有字段)"
    echo "   响应: $(echo "$BODY" | jq -c .data)"
else
    echo "❌ 预期 HTTP 201，实际: $HTTP_CODE"
    echo "   响应: $BODY"
    exit 1
fi
echo ""

echo "=================================="
echo "✅ 所有 Swagger 必填字段测试通过！"
echo ""
echo "📊 测试摘要:"
echo "   - email 字段正确标记为必填"
echo "   - password 字段正确标记为必填"
echo "   - API 正确验证必填字段"
echo "   - 可选字段可以省略"
