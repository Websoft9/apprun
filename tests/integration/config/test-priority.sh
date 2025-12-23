#!/bin/bash

# 加载共享函数
source "$(dirname "$0")/../../common.sh"

echo "=========================================="
echo "测试配置优先级"
echo "=========================================="

# 清理之前的测试数据
echo "清理之前的测试数据..."
response=$(http_put "/config" '{"poc.apikey": "poc-api-key-12345678901234", "poc.enabled": true}')
echo "测试数据清理完成"

# 测试 1: 查看当前配置（default.yaml + 环境变量）
echo ""
echo "1. 查看当前配置（default.yaml + 环境变量）"
response=$(http_get "/config")
db_host=$(echo $response | jq -r '.[] | select(.path=="database.host") | .value')
poc_apikey=$(echo $response | jq -r '.[] | select(.path=="poc.apikey") | .value')
assert_equals "postgres" "$db_host" "database.host 使用环境变量值"
assert_equals "poc-api-key-12345678901234" "$poc_apikey" "poc.apikey 使用环境变量值"

# 测试 2: 修改 DB 中的配置（poc.apikey）
echo ""
echo "2. 修改 DB 中的配置（poc.apikey）"
response=$(http_put "/config" '{"poc.apikey": "db-stored-key-789"}')
poc_apikey=$(echo $response | jq -r '.[] | select(.path=="poc.apikey") | .value')
assert_equals "db-stored-key-789" "$poc_apikey" "DB 配置更新成功"

# 测试 3: 验证 DB 配置是否生效（重启容器）
echo ""
echo "3. 验证 DB 配置是否生效（重启容器）"
docker compose -f ../../docker/docker-compose.yml stop app > /dev/null 2>&1
docker compose -f ../../docker/docker-compose.yml start app > /dev/null 2>&1
sleep 8
echo "调用API..."
response=$(http_get "/config")
echo "API响应: $response"
poc_apikey=$(echo $response | jq -r '.[] | select(.path=="poc.apikey") | .value')
echo "解析后的poc.apikey: $poc_apikey"
assert_equals "poc-api-key-12345678901234" "$poc_apikey" "环境变量覆盖 DB 配置"

# 测试 4: 验证 database.host（环境变量覆盖文件）
echo ""
echo "4. 验证 database.host（环境变量覆盖文件）"
response=$(http_get "/config")
db_host=$(echo $response | jq -r '.[] | select(.path=="database.host") | .value')
assert_equals "postgres" "$db_host" "database.host 使用环境变量值"

echo ""
echo "=========================================="
echo "优先级验证完成"
echo "=========================================="

# 打印测试摘要
print_summary
