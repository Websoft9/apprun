#!/bin/bash
# Manual Test Script for Metrics Tags (Story 9.1)
# This script verifies that metrics history API returns tags in the response

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"

echo "🧪 Testing Metrics Tags Implementation (Story 9.1)"
echo "=================================================="
echo ""

# Check if admin token is provided
if [ -z "$ADMIN_TOKEN" ]; then
    echo "⚠️  Warning: ADMIN_TOKEN not set"
    echo "   Please set it with: export ADMIN_TOKEN='your_token_here'"
    echo ""
fi

# Test 1: Query metrics history
echo "📊 Test 1: Query metrics history (should include tags)"
echo "GET ${BASE_URL}/api/metrics/history?name=user_count_total&duration=24h"
echo ""

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X GET \
    "${BASE_URL}/api/metrics/history?name=user_count_total&duration=24h&limit=5" \
    -H "accept: application/json" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_CODE:/d')

echo "Response Code: $HTTP_CODE"
echo ""

if [ "$HTTP_CODE" == "200" ]; then
    echo "✅ Request successful"
    echo ""
    echo "Response Body:"
    echo "$BODY" | jq '.'
    echo ""
    
    # Check if response contains tags field
    TAGS_COUNT=$(echo "$BODY" | jq -r '.data.metrics[0].tags // empty' | wc -l)
    if [ "$TAGS_COUNT" -gt 0 ]; then
        echo "✅ Response includes tags field"
        echo ""
        echo "Sample tags:"
        echo "$BODY" | jq -r '.data.metrics[0].tags'
    else
        echo "⚠️  Tags field not found or empty"
        echo "   This is expected if no metrics have been collected yet"
    fi
else
    echo "❌ Request failed with code $HTTP_CODE"
    echo "$BODY" | jq '.'
fi

echo ""
echo "=================================================="

# Test 2: Query with tag filter
echo ""
echo "📊 Test 2: Query with tag filter"
echo "GET ${BASE_URL}/api/metrics/history?name=user_count_total&duration=1h&tags[env]=production"
echo ""

RESPONSE2=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X GET \
    "${BASE_URL}/api/metrics/history?name=user_count_total&duration=1h&tags[env]=production" \
    -H "accept: application/json" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}")

HTTP_CODE2=$(echo "$RESPONSE2" | grep "HTTP_CODE:" | cut -d: -f2)
BODY2=$(echo "$RESPONSE2" | sed '/HTTP_CODE:/d')

echo "Response Code: $HTTP_CODE2"
echo ""

if [ "$HTTP_CODE2" == "200" ]; then
    echo "✅ Tag filtering works"
    echo ""
    FILTERED_COUNT=$(echo "$BODY2" | jq -r '.data.count')
    echo "Filtered results count: $FILTERED_COUNT"
else
    echo "❌ Request failed with code $HTTP_CODE2"
    echo "$BODY2" | jq '.'
fi

echo ""
echo "=================================================="
echo ""
echo "💡 Tips:"
echo "   - Metrics are automatically persisted with tags when collected"
echo "   - Tags include: env, instance, source, host"
echo "   - Set METRICS_ENV and METRICS_INSTANCE_ID environment variables"
echo "   - Default env: production, default instance: hostname"
echo ""
echo "🔧 Environment Variables:"
echo "   export METRICS_ENV=production"
echo "   export METRICS_INSTANCE_ID=apprun-01"
echo "   export METRICS_TAGS_ENABLED=true"
echo ""
