# Docker Observability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐 Docker 部署的日志轮转、请求 ID 追踪和日志查看入口。

**Architecture:** 在 Compose 层使用统一 logging 锚点；在 Nginx 层增加 request id 日志和转发头；通过 Bash 脚本提供只读日志查看；通过部署静态检查防止配置回退。

**Tech Stack:** Docker Compose、Nginx、Bash、Node.js 静态检查、中文部署文档。

---

## File Structure

- Modify: `scripts/check-deploy-config.mjs`
  - 扩展 env、compose、nginx、日志脚本和部署文档检查。
- Modify: `backend/.env.example`
  - 新增 `WECHECKIN_LOG_MAX_SIZE` 和 `WECHECKIN_LOG_MAX_FILE`。
- Modify: `backend/docker-compose.yml`
  - 新增 `x-logging`，四个服务复用 `logging: *default-logging`。
- Modify: `backend/nginx.conf`
  - 访问日志记录 `request_id`，代理透传 `X-Request-ID`。
- Create: `backend/scripts/docker-logs.sh`
  - 提供 Docker Compose 服务日志跟踪入口。
- Modify: `docs/DEPLOYMENT_TROUBLESHOOTING.md`
  - 新增日志轮转、请求 ID 和日志查看说明。
- Modify: `README.md`
  - 在 Docker 部署段落补充日志查看命令。

## Tasks

- [x] 新增部署静态检查，并确认当前实现红灯。
- [x] 增加 Compose 日志轮转环境变量和 logging 锚点。
- [x] 增加 Nginx request id 日志和 `X-Request-ID` 透传。
- [x] 新增 `backend/scripts/docker-logs.sh`。
- [x] 更新中文部署文档。
- [x] 运行 `node scripts/check-deploy-config.mjs`、`docker compose -f backend/docker-compose.yml config`、`CHECK_ADMIN_BUILD=1 bash scripts/check.sh`。
