#!/bin/bash

# 加载共享函数
source "$(dirname "$0")/../../common.sh"

echo "======================================"
echo "配置中心 API 测试"
echo "======================================"

# 测试 1: GET /config
echo ""
echo "测试 1: GET /config - 获取所有配置项"
response=$(http_get "/config")
assert_equals "true" "$(echo $response | jq -r 'length > 0')" "返回配置项列表"

# 测试 2: PUT /config - 修改可存储的配置项
echo ""
echo "测试 2: PUT /config - 修改可存储的配置项"
response=$(http_put "/config" '{"poc.enabled": false, "poc.apikey": "new-secret-key-456"}')
assert_equals "false" "$(echo $response | jq -r '.[] | select(.path=="poc.enabled") | .value')" "poc.enabled 更新成功"
assert_equals "new-secret-key-456" "$(echo $response | jq -r '.[] | select(.path=="poc.apikey") | .value')" "poc.apikey 更新成功"

# 测试 3: GET /config - 验证修改后的配置
echo ""
echo "测试 3: GET /config - 验证修改后的配置"
response=$(http_get "/config")
poc_enabled=$(echo $response | jq -r '.[] | select(.path=="poc.enabled") | .value')
poc_apikey=$(echo $response | jq -r '.[] | select(.path=="poc.apikey") | .value')
assert_equals "false" "$poc_enabled" "poc.enabled 保持修改后的值"
assert_equals "new-secret-key-456" "$poc_apikey" "poc.apikey 保持修改后的值"

# 测试 4: PUT /config - 尝试修改不可存储的配置项 (应该失败)
echo ""
echo "测试 4: PUT /config - 尝试修改不可存储的配置项 (应该失败)"
response=$(http_put "/config" '{"app.name": "hacked"}')
assert_equals "true" "$(echo $response | grep -q 'not allowed to be modified' && echo 'true' || echo 'false')" "拒绝修改不可存储的配置项"

# 打印测试摘要
print_summary
