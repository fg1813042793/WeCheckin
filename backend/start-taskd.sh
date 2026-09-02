#!/bin/bash

# WeCheckin 定时任务服务启动脚本

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}===================================${NC}"
echo -e "${GREEN} WeCheckin 定时任务服务启动脚本${NC}"
echo -e "${GREEN}===================================${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未找到 Go 环境${NC}"
    echo -e "${YELLOW}请先安装 Go 1.24 或更高版本${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Go 环境检查通过${NC}"

if ! command -v mysql &> /dev/null; then
    echo -e "${YELLOW}警告: 未找到 MySQL 客户端${NC}"
    echo -e "${YELLOW}请确保 MySQL 服务器已安装并运行${NC}"
else
    echo -e "${GREEN}✓ MySQL 客户端检查通过${NC}"
fi

if ! command -v redis-cli &> /dev/null; then
    echo -e "${YELLOW}警告: 未找到 Redis 客户端${NC}"
    echo -e "${YELLOW}请确保 Redis 服务器已安装并运行${NC}"
else
    echo -e "${GREEN}✓ Redis 客户端检查通过${NC}"
fi

cd "$(dirname "$0")"
BACKEND_DIR="$(pwd)"

echo -e "${GREEN}后端目录: ${BACKEND_DIR}${NC}"
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
export GOCACHE="${GOCACHE:-${BACKEND_DIR}/../.cache/go-build}"
mkdir -p "${GOCACHE}"

echo -e "${YELLOW}正在检查依赖...${NC}"
go mod download

if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 依赖检查失败${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 依赖检查完成${NC}"

if [ ! -f "config/config.yaml" ]; then
    echo -e "${RED}错误: 配置文件不存在${NC}"
    echo -e "${YELLOW}请创建 config/config.yaml 文件${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 配置文件检查通过${NC}"

mkdir -p logs
mkdir -p uploads

export APP_ENV="${APP_ENV:-development}"

TASKD_ROLE="${TASKD_ROLE:-all}"
TASKD_ARGS=("$@")
HAS_ROLE=false
for arg in "${TASKD_ARGS[@]}"; do
    case "$arg" in
        --role|--role=*)
            HAS_ROLE=true
            break
            ;;
    esac
done

if [ "$HAS_ROLE" = false ]; then
    TASKD_ARGS=("--role=${TASKD_ROLE}" "${TASKD_ARGS[@]}")
fi

echo -e "${GREEN}正在启动定时任务服务...${NC}"
echo -e "${YELLOW}该进程不监听 HTTP 端口${NC}"
echo -e "${YELLOW}按 Ctrl+C 停止服务${NC}"

go run ./cmd/taskd "${TASKD_ARGS[@]}"
TASKD_EXIT_CODE=$?

if [ "${TASKD_EXIT_CODE}" -ne 0 ]; then
    echo -e "${RED}错误: 定时任务服务启动或运行失败${NC}"
    exit "${TASKD_EXIT_CODE}"
fi

echo -e "${GREEN}定时任务服务已停止${NC}"
