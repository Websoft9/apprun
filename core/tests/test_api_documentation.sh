#!/bin/bash
# Test API Documentation Improvements
# Verifies correct API paths and health endpoint

set -e

echo "🧪 API 文档改进验证"
echo "======================="
echo ""

SWAGGER_FILE="/data/cdl/apprun/core/docs/swagger.json"

if [ ! -f "$SWAGGER_FILE" ]; then
    echo "❌ Swagger 文档不存在: $SWAGGER_FILE"
    exit 1
fi

echo "📋 验证 1: BasePath 设置"
BASE_PATH=$(cat "$SWAGGER_FILE" | jq -r '.basePath')
if [ "$BASE_PATH" = "/api" ]; then
    echo "✅ BasePath 正确: $BASE_PATH"
else
    echo "❌ BasePath 错误: $BASE_PATH (期望: /api)"
    exit 1
fi
echo ""

echo "📋 验证 2: API 路径结构"
echo "期望的路径 (相对于 /api):"
EXPECTED_PATHS=(
    "/health"
    "/demo/i18n"
    "/v1/auth/register"
    "/config"
    "/config/list"
    "/config/allowed"
)

for path in "${EXPECTED_PATHS[@]}"; do
    if cat "$SWAGGER_FILE" | jq -e ".paths.\"$path\"" > /dev/null 2>&1; then
        # 获取完整的 URL
        FULL_URL="/api${path}"
        echo "  ✅ $path → $FULL_URL"
    else
        echo "  ❌ 缺少路径: $path"
        exit 1
    fi
done
echo ""

echo "📋 验证 3: 确保没有重复的 /api 前缀"
INVALID_PATHS=$(cat "$SWAGGER_FILE" | jq -r '.paths | keys[]' | grep "^/api/" || true)
if [ -z "$INVALID_PATHS" ]; then
    echo "✅ 没有重复的 /api 前缀"
else
    echo "❌ 发现重复的 /api 前缀:"
    echo "$INVALID_PATHS"
    exit 1
fi
echo ""

echo "📋 验证 4: Health 端点定义"
HEALTH_TAG=$(cat "$SWAGGER_FILE" | jq -r '.paths."/health".get.tags[0]')
HEALTH_SUMMARY=$(cat "$SWAGGER_FILE" | jq -r '.paths."/health".get.summary')
if [ "$HEALTH_TAG" = "system" ] && [ "$HEALTH_SUMMARY" = "Health Check" ]; then
    echo "✅ Health 端点正确定义"
    echo "   Tag: $HEALTH_TAG"
    echo "   Summary: $HEALTH_SUMMARY"
else
    echo "❌ Health 端点定义错误"
    exit 1
fi
echo ""

echo "📋 验证 5: 注册端点路径"
REGISTER_PATH="/v1/auth/register"
REGISTER_METHOD=$(cat "$SWAGGER_FILE" | jq -r ".paths.\"$REGISTER_PATH\" | keys[0]")
if [ "$REGISTER_METHOD" = "post" ]; then
    echo "✅ 注册端点路径正确: POST /api$REGISTER_PATH"
else
    echo "❌ 注册端点路径错误"
    exit 1
fi
echo ""

echo "📋 验证 6: 必填字段标注"
REQUIRED_FIELDS=$(cat "$SWAGGER_FILE" | jq -r '.definitions."service.RegisterRequest".required | join(", ")')
if [ "$REQUIRED_FIELDS" = "email, password" ]; then
    echo "✅ 必填字段正确标注: $REQUIRED_FIELDS"
else
    echo "❌ 必填字段错误: $REQUIRED_FIELDS"
    exit 1
fi
echo ""

echo "======================="
echo "✅ 所有 API 文档改进验证通过！"
echo ""
echo "📊 改进总结:"
echo "   1. BasePath = /api"
echo "   2. 路径不包含重复的 /api 前缀"
echo "   3. 添加了 /health 端点"
echo "   4. 注册端点: /api/v1/auth/register"
echo "   5. Demo 端点: /api/demo/i18n"
echo "   6. 必填字段正确标注"
echo ""
echo "🌐 实际访问 URL:"
for path in "${EXPECTED_PATHS[@]}"; do
    echo "   http://localhost:8080/api${path}"
done
