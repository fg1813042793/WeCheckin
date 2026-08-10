#!/bin/bash

# WeCheckin 后端初始化/迁移脚本

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}===================================${NC}"
echo -e "${GREEN} WeCheckin 后端初始化/迁移脚本${NC}"
echo -e "${GREEN}===================================${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未找到 Go 环境${NC}"
    echo -e "${YELLOW}请先安装 Go 1.21 或更高版本${NC}"
    exit 1
fi

cd "$(dirname "$0")"
BACKEND_DIR="$(pwd)"
MIGRATIONS_DIR="${WECHECKIN_MIGRATIONS_DIR:-migrations}"

echo -e "${GREEN}后端目录: ${BACKEND_DIR}${NC}"
echo -e "${YELLOW}迁移目录: ${MIGRATIONS_DIR}${NC}"
echo -e "${YELLOW}执行记录表: schema_migrations，已执行过的任务会自动跳过${NC}"

if [ ! -f "config/config.prod.yaml" ]; then
    echo -e "${RED}错误: 配置文件不存在${NC}"
    echo -e "${YELLOW}请创建 config/config.prod.yaml 文件${NC}"
    exit 1
fi

mkdir -p logs
mkdir -p uploads

echo -e "${GREEN}正在执行初始化/迁移任务...${NC}"
go run ./cmd/maintenance -migrations "${MIGRATIONS_DIR}" "$@" -env prod

if [ $? -ne 0 ]; then
    echo -e "${RED}初始化/迁移失败${NC}"
    exit 1
fi

echo -e "${GREEN}初始化/迁移完成${NC}"
