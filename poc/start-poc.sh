#!/bin/bash

# apprun POC 环境启动脚本
# 用途：一键启动所有 POC 服务并验证状态

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}"
echo "╔════════════════════════════════════════════════╗"
echo "║        apprun POC Environment Setup            ║"
echo "║        技术验证环境启动                          ║"
echo "╚════════════════════════════════════════════════╝"
echo -e "${NC}"

# 检查依赖
echo -e "${YELLOW}[1/6] 检查依赖...${NC}"
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Error: docker 未安装${NC}"; exit 1; }
command -v docker compose >/dev/null 2>&1 || { echo -e "${RED}Error: docker compose 未安装${NC}"; exit 1; }
echo -e "${GREEN}✓ 依赖检查通过${NC}"

# 清理旧容器
echo -e "${YELLOW}[2/6] 清理旧环境（如有）...${NC}"
docker compose down -v 2>/dev/null || true
echo -e "${GREEN}✓ 清理完成${NC}"

# 启动服务
echo -e "${YELLOW}[3/6] 启动 Docker 服务...${NC}"
docker compose up -d

# 等待服务就绪
echo -e "${YELLOW}[4/6] 等待服务启动...${NC}"
sleep 5

# 健康检查
echo -e "${YELLOW}[5/6] 健康检查...${NC}"

check_service() {
    local name=$1
    local url=$2
    local max_attempts=30
    local attempt=0
    
    echo -n "  检查 $name ... "
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC}"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    
    echo -e "${RED}✗ (超时)${NC}"
    return 1
}

check_service "PostgreSQL" "http://localhost:5432" || echo "  (使用 pg_isready 检查)"
docker exec apprun-poc-postgres pg_isready -U apprun && echo -e "  PostgreSQL ${GREEN}✓${NC}"

check_service "PostgREST" "http://localhost:3000"
check_service "Kratos (Public)" "http://localhost:4433/health/alive"
check_service "Kratos (Admin)" "http://localhost:4434/health/ready"

# 显示服务信息
echo -e "${YELLOW}[6/6] POC 环境信息${NC}"
echo ""
echo "╔════════════════════════════════════════════════╗"
echo "║           服务访问地址                          ║"
echo "╠════════════════════════════════════════════════╣"
echo "║ PostgreSQL      : localhost:5432               ║"
echo "║   - Database    : apprun_poc                   ║"
echo "║   - User        : apprun                       ║"
echo "║   - Password    : apprun123                    ║"
echo "║                                                ║"
echo "║ PostgREST API   : http://localhost:3000        ║"
echo "║   - 示例        : GET /products                ║"
echo "║                                                ║"
echo "║ Ory Kratos                                     ║"
echo "║   - Public API  : http://localhost:4433        ║"
echo "║   - Admin API   : http://localhost:4434        ║"
echo "╚════════════════════════════════════════════════╝"
echo ""

echo -e "${GREEN}✓ POC 环境启动成功！${NC}"
echo ""
echo "快速测试："
echo "  1. 测试 PostgREST API:"
echo "     curl http://localhost:3000/products"
echo ""
echo "  2. 测试 Kratos 健康状态:"
echo "     curl http://localhost:4433/health/alive"
echo ""
echo "查看日志："
echo "  docker compose logs -f [service-name]"
echo ""
echo "停止环境："
echo "  docker compose down"
echo ""
echo "完整清理（包括数据）："
echo "  docker compose down -v"
echo ""
