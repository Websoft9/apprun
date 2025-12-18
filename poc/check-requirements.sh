#!/bin/bash

# apprun POC - 启动前检查脚本

echo "🔍 检查 POC 环境前置条件..."
echo ""

# 检查 Docker
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
    echo "✅ Docker 已安装: $DOCKER_VERSION"
else
    echo "❌ Docker 未安装，请先安装 Docker"
    echo "   安装命令: curl -fsSL https://get.docker.com | sh"
    exit 1
fi

# 检查 Docker Compose
if command -v docker-compose &> /dev/null; then
    COMPOSE_VERSION=$(docker-compose --version | awk '{print $4}' | sed 's/,//')
    echo "✅ Docker Compose 已安装: $COMPOSE_VERSION"
else
    echo "❌ Docker Compose 未安装，请先安装"
    exit 1
fi

# 检查 Docker 服务
if docker info &> /dev/null; then
    echo "✅ Docker 服务运行正常"
else
    echo "❌ Docker 服务未运行，请启动 Docker"
    exit 1
fi

# 检查端口占用
echo ""
echo "🔍 检查端口占用情况..."
PORTS=(5432 3000 4433 4434 7233 8233 9000 9001)
PORT_OK=true

for PORT in "${PORTS[@]}"; do
    if netstat -tuln 2>/dev/null | grep -q ":$PORT "; then
        echo "⚠️  端口 $PORT 已被占用"
        PORT_OK=false
    fi
done

if [ "$PORT_OK" = true ]; then
    echo "✅ 所有必需端口可用"
else
    echo ""
    echo "❌ 部分端口已被占用，请先关闭占用端口的服务"
    echo "   或修改 docker-compose.yml 中的端口映射"
    exit 1
fi

# 检查磁盘空间
echo ""
echo "🔍 检查磁盘空间..."
AVAILABLE_SPACE=$(df -BG . | tail -1 | awk '{print $4}' | sed 's/G//')
if [ "$AVAILABLE_SPACE" -ge 20 ]; then
    echo "✅ 磁盘空间充足: ${AVAILABLE_SPACE}GB 可用"
else
    echo "⚠️  磁盘空间不足: ${AVAILABLE_SPACE}GB 可用（建议至少 20GB）"
fi

# 检查内存
echo ""
echo "🔍 检查可用内存..."
AVAILABLE_MEM=$(free -g | awk '/^Mem:/{print $7}')
if [ "$AVAILABLE_MEM" -ge 4 ]; then
    echo "✅ 可用内存充足: ${AVAILABLE_MEM}GB"
else
    echo "⚠️  可用内存不足: ${AVAILABLE_MEM}GB（建议至少 4GB）"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 环境检查完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📝 下一步："
echo "   1. 启动 POC 环境:  ./start-poc.sh"
echo "   2. 查看服务状态:   docker-compose ps"
echo "   3. 查看日志:       docker-compose logs -f"
echo ""
