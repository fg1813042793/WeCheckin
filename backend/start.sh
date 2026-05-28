#!/bin/bash

# WeCheckin 后端启动脚本

# 设置颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}===================================${NC}"
echo -e "${GREEN} WeCheckin 后端服务启动脚本${NC}"
echo -e "${GREEN}===================================${NC}"

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未找到 Go 环境${NC}"
    echo -e "${YELLOW}请先安装 Go 1.21 或更高版本${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Go 环境检查通过${NC}"

# 检查 MySQL 环境
if ! command -v mysql &> /dev/null; then
    echo -e "${YELLOW}警告: 未找到 MySQL 客户端${NC}"
    echo -e "${YELLOW}请确保 MySQL 服务器已安装并运行${NC}"
else
    echo -e "${GREEN}✓ MySQL 客户端检查通过${NC}"
fi

# 检查 Redis 环境
if ! command -v redis-cli &> /dev/null; then
    echo -e "${YELLOW}警告: 未找到 Redis 客户端${NC}"
    echo -e "${YELLOW}请确保 Redis 服务器已安装并运行${NC}"
else
    echo -e "${GREEN}✓ Redis 客户端检查通过${NC}"
fi

# 进入后端目录
cd "$(dirname "$0")"
BACKEND_DIR="$(pwd)"

echo -e "${GREEN}后端目录: ${BACKEND_DIR}${NC}"

# 下载依赖
echo -e "${YELLOW}正在下载依赖...${NC}"
go mod tidy

if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 依赖下载失败${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 依赖下载完成${NC}"

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo -e "${RED}错误: 配置文件不存在${NC}"
    echo -e "${YELLOW}请创建 config.yaml 文件${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 配置文件检查通过${NC}"

# 创建必要的目录
mkdir -p logs
mkdir -p uploads

# 设置环境变量
export GIN_MODE=debug
export APP_ENV=development

# 启动服务
echo -e "${GREEN}正在启动服务...${NC}"
echo -e "${YELLOW}服务将在 http://localhost:8080 启动${NC}"
echo -e "${YELLOW}按 Ctrl+C 停止服务${NC}"

# 运行服务
go run cmd/main.go

echo -e "${GREEN}服务已停止${NC}"