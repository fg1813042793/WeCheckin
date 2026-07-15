# Docker Deployment Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 WeCheckin 的 Docker Compose 部署具备环境样板、健康依赖、备份恢复脚本和项目级防回归检查。

**Architecture:** 保留现有 `backend/` 容器部署结构，在 Compose 层补 env_file、健康检查和 depends_on 条件；在 `backend/scripts/` 提供 Docker 场景备份/恢复脚本；在根级 `scripts/check.sh` 接入静态检查，避免部署配置回退。

**Tech Stack:** Docker Compose、Nginx、MySQL 8、Redis 7、Go 后端、Node.js 静态检查、Bash 脚本。

---

## File Structure

- Create: `backend/.env.example`
  - Docker Compose 环境变量样板，覆盖 MySQL、Redis、后端端口、CORS、迁移开关。
- Modify: `backend/docker-compose.yml`
  - 使用 env_file 和变量默认值，新增 MySQL/Redis healthcheck，backend/Nginx 改为 healthy 依赖。
- Create: `backend/scripts/docker-backup.sh`
  - Docker Compose 数据库备份脚本，输出到 `backend/backups/`。
- Create: `backend/scripts/docker-restore.sh`
  - Docker Compose 数据库恢复脚本，要求显式确认。
- Create: `scripts/check-deploy-config.mjs`
  - 静态检查 Docker/env/备份/恢复/文档关键约束。
- Modify: `scripts/check.sh`
  - 接入部署配置检查。
- Modify: `README.md`
  - 补充 Docker env 示例和备份恢复入口。
- Modify: `docs/DEPLOYMENT_TROUBLESHOOTING.md`
  - 补充 `.env`、健康依赖、备份恢复说明。
- Modify: `docs/superpowers/plans/2026-07-14-full-stack-optimization-roadmap.md`
  - 同步本批次进度。

## Tasks

- [x] 新增部署配置静态检查，并确认当前实现红灯。
- [x] 新增 `backend/.env.example`。
- [x] 强化 `backend/docker-compose.yml` 健康检查和健康依赖。
- [x] 新增 Docker 备份和恢复脚本。
- [x] 将部署检查接入 `scripts/check.sh`。
- [x] 更新中文 README、部署排障文档和路线图。
- [x] 运行 `node scripts/check-deploy-config.mjs`、`docker compose -f backend/docker-compose.yml config`、`bash scripts/check.sh`、`CHECK_ADMIN_BUILD=1 bash scripts/check.sh`、`CHECK_FRONTEND_BUILD=1 bash scripts/check.sh`。
