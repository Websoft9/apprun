#!/bin/bash

# 测试环境准备脚本
# 启动测试数据库和应用

set -e

echo "准备测试环境..."

# 启动 Docker 环境
cd "$(dirname "$0")/../../docker"
docker compose up -d

# 等待服务启动
echo "等待服务启动..."
sleep 5

# 在容器内验证服务状态
if docker exec docker-app-1 wget -qO- http://localhost:8080/ | grep -q "apprun"; then
    echo "✓ 应用服务启动成功"
else
    echo "✗ 应用服务启动失败"
    exit 1
fi

echo "测试环境准备完成"