#!/bin/bash

# 测试数据清理脚本

echo "清理测试数据..."

# 停止 Docker 环境
cd "$(dirname "$0")/../../docker"
docker compose down

# 清理数据库中的测试配置数据
echo "清理数据库配置数据..."
if docker ps | grep -q postgres; then
    docker exec docker-postgres-1 psql -U postgres -d apprun -c "DELETE FROM configitems WHERE key IN ('poc.enabled', 'poc.apikey');" 2>/dev/null || true
fi

# 清理测试文件（如果有）
echo "测试环境清理完成"