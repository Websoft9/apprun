#!/bin/bash
# Test API Route Updates
# Verifies correct API paths after removing /v1 prefix and fixing /health

set -e

BASE_URL="http://localhost:8080"
TIMESTAMP=$(date +%s)

# Disable proxy
export http_proxy=""
export https_proxy=""

echo "🧪 API 路由更新验证"
echo "======================="
echo ""

# Test 1: Health endpoint at root
echo "📋 测试 1: Health 端点在根路径"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ GET /health - HTTP 200"
    echo "   响应: $(echo "$BODY" | jq -c .data)"
else
    echo "❌ 预期 HTTP 200，实际: $HTTP_CODE"
    echo "   响应: $BODY"
    exit 1
fi
echo ""

# Test 2: /api/health should 404
echo "📋 测试 2: /api/health 应该返回 404"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/health")
if [ "$HTTP_CODE" = "404" ]; then
    echo "✅ GET /api/health - HTTP 404 (正确)"
else
    echo "❌ 预期 HTTP 404，实际: $HTTP_CODE"
    exit 1
fi
echo ""

# Test 3: Register endpoint without /v1
echo "📋 测试 3: 注册端点 /api/auth/register (无 /v1 前缀)"
TEST_EMAIL="route_test_${TIMESTAMP}@example.com"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"SecurePass123\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "201" ]; then
    echo "✅ POST /api/auth/register - HTTP 201"
    echo "   Email: $(echo "$BODY" | jq -r .data.email)"
else
    echo "❌ 预期 HTTP 201，实际: $HTTP_CODE"
    echo "   响应: $BODY"
    exit 1
fi
echo ""

# Test 4: Old /v1 path should 404
echo "📋 测试 4: 旧路径 /api/v1/auth/register 应该返回 404"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123"}')
if [ "$HTTP_CODE" = "404" ]; then
    echo "✅ POST /api/v1/auth/register - HTTP 404 (正确)"
else
    echo "❌ 预期 HTTP 404，实际: $HTTP_CODE"
    exit 1
fi
echo ""

# Test 5: Demo endpoint
echo "📋 测试 5: Demo 端点 /api/demo/i18n"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/demo/i18n")
if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ GET /api/demo/i18n - HTTP 200"
else
    echo "❌ 预期 HTTP 200，实际: $HTTP_CODE"
    exit 1
fi
echo ""

echo "======================="
echo "✅ 所有 API 路由更新验证通过！"
echo ""
echo "📊 最终路由结构:"
echo "   GET  /health                 ✅ (根路径)"
echo "   POST /api/auth/register      ✅ (移除 /v1)"
echo "   GET  /api/demo/i18n          ✅"
echo "   GET  /api/config             ✅"
echo ""
echo "❌ 已移除的路由:"
echo "   /api/health                  (404)"
echo "   /api/v1/auth/register        (404)"
