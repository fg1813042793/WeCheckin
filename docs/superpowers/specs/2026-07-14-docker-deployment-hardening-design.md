# Docker 部署韧性优化设计

## 背景

当前后端 Dockerfile、Compose 和 Nginx 已经统一到 `8083`，并提供 `/health`、`/ready`。下一步需要让容器部署更适合真实环境：配置从示例环境文件进入 Compose，服务启动顺序基于健康状态，数据库备份和恢复有固定脚本，项目级检查能阻止部署配置回退。

## 目标

- 提供 `backend/.env.example`，集中说明 Docker 部署所需数据库、Redis、后端端口、迁移和 CORS 配置。
- 为 MySQL、Redis、backend 和 Nginx 增加可检查的健康依赖关系。
- 增加 Docker 场景下的数据库备份和恢复脚本。
- 将 Docker 部署配置纳入 `bash scripts/check.sh`，避免端口、健康检查、备份脚本和 env 示例漂移。
- 更新中文部署文档和路线图。

## 非目标

- 不引入 Kubernetes、Traefik、证书自动签发或云厂商 Secret 管理。
- 不改变后端 API、数据库模型和登录逻辑。
- 不在本批次实际执行 Docker 容器启动、数据库备份或恢复。

## 设计

### 配置

新增 `backend/.env.example`，作为 `backend/docker-compose.yml` 的环境变量样板。Compose 使用 `${VAR:-default}` 保留本地可启动能力，但文档明确生产环境必须复制为 `.env` 并修改默认密码。

### 健康依赖

MySQL 使用 `mysqladmin ping`，Redis 使用 `redis-cli ping`。backend 依赖 MySQL 和 Redis 的 healthy 状态，Nginx 依赖 backend 的 healthy 状态。backend 的 healthcheck 继续请求 `http://127.0.0.1:8083/health`，避免依赖数据库；数据库可用性由 `/ready` 单独表达。

### 备份和恢复

新增两个脚本：

- `backend/scripts/docker-backup.sh`：通过 `docker compose exec -T mysql mysqldump` 导出数据库到 `backend/backups/`，并可选打包 `uploads/`。
- `backend/scripts/docker-restore.sh`：从指定 SQL 文件恢复数据库，恢复前要求用户显式确认。

脚本只面向 Docker Compose 场景，不处理远程对象存储和跨主机同步。

### 检查

新增根级 `scripts/check-deploy-config.mjs`。检查项覆盖：

- `backend/.env.example` 是否存在并包含关键变量。
- `backend/docker-compose.yml` 是否包含 MySQL/Redis/backend/Nginx healthcheck 和健康依赖。
- 端口仍为 `8083`。
- 备份/恢复脚本存在并包含关键命令。
- 部署文档引用 env 示例、健康检查和备份恢复脚本。

## 风险

- Compose 的 `depends_on.condition: service_healthy` 依赖 Docker Compose v2；旧版 docker-compose 可能不支持。
- Redis 启用密码后，本地临时调试需要同步更新 backend `.env`。
- 恢复脚本会覆盖数据库内容，所以必须要求显式确认。
