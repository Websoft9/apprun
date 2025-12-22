#!/bin/bash

# 测试配置中心 API

BASE_URL="http://localhost:8080"
# 禁用代理
export no_proxy=localhost,127.0.0.1

echo "======================================"
echo "测试 1: GET /config - 获取所有配置项"
echo "======================================"
curl -s -X GET "$BASE_URL/config" | python3 -m json.tool

echo ""
echo "======================================"
echo "测试 2: PUT /config - 修改可存储的配置项"
echo "======================================"
curl -s -X PUT "$BASE_URL/config" \
  -H "Content-Type: application/json" \
  -d '{
    "poc.enabled": false,
    "poc.apikey": "new-secret-key-456"
  }' | python3 -m json.tool

echo ""
echo "======================================"
echo "测试 3: GET /config - 验证修改后的配置"
echo "======================================"
curl -s -X GET "$BASE_URL/config" | python3 -m json.tool | grep -A3 "poc"

echo ""
echo "======================================"
echo "测试 4: PUT /config - 尝试修改不可存储的配置项 (应该失败)"
echo "======================================"
curl -s -X PUT "$BASE_URL/config" \
  -H "Content-Type: application/json" \
  -d '{
    "app.name": "hacked"
  }'

echo ""
echo "======================================"
echo "测试完成"
echo "======================================"
