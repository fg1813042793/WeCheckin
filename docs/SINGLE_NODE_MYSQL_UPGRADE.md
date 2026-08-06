# 单点 MySQL 部署兼容升级说明

最后更新：2026-07-28

本文档用于已有单机部署升级到当前版本时参考。这里的“单点部署”指 Nginx、Go 后端、MySQL、Redis 在同一台服务器或同一台服务器上的 Docker Compose 环境中运行。

## 兼容结论

当前版本可以兼容已有 MySQL 单点部署，但生产升级建议按保守流程执行：

- 后端数据库驱动为 MySQL，推荐 MySQL 8.0。
- 新接口统一走 `/api/v2` 和 `/api/v2/admin`。
- 后台管理只使用 `/api/v2/admin/*`；旧版 `/admin/*` 后台路由已不再作为兼容入口。
- `/passport/*`、`/home/*`、`/survey/*`、`/exam/*` 等历史客户端路径如仍存在，仅用于兼容旧页面和小程序旧代码。
- Nginx 需要代理 `/api/`；如仍承载旧客户端入口，再代理对应旧路径。`/api/` 的 `proxy_pass` 不能带结尾斜杠，否则会截掉 `/api/v2` 前缀。
- 历史 MD5 密码仍可登录，登录成功后会自动升级为 bcrypt。
- 后端启动不再执行迁移或种子数据；初始化/迁移统一通过 `backend/init.sh` 手动执行。
- 迁移执行记录写入 `schema_migrations`，已执行过的 SQL 文件和初始化任务不会重复执行。

## 主要风险

风险主要来自数据库结构升级，而不是接口路由：

- `backend/init.sh` 会补齐表结构、执行版本化 SQL 和基础权限种子。
- 老生产库数据量较大时，DDL 可能锁表或超时。
- 如果跳过初始化/迁移，新代码访问旧表结构时也可能因为缺字段而报错。

因此生产环境不要在没有备份和验证的情况下直接放量。

## 推荐升级流程

1. 备份数据库和上传文件。

```bash
mysqldump -h 127.0.0.1 -P 3306 -u wecheckin -p --single-transaction --routines --triggers wecheckin > wecheckin-before-upgrade.sql
tar -czf uploads-before-upgrade.tar.gz backend/uploads
```

Docker Compose 部署可使用项目脚本：

```bash
cd backend
bash scripts/docker-backup.sh
```

2. 记录旧版本运行配置。

```bash
env | grep '^WECHECKIN_'
```

至少确认 MySQL、Redis、端口、CORS、Token 前缀和上传目录配置。

3. 首次启动新后端时先只验证连接、健康检查和路由，服务启动不会执行数据库迁移。

```bash
WECHECKIN_DATABASE_HOST=127.0.0.1 \
WECHECKIN_DATABASE_PORT=3306 \
WECHECKIN_DATABASE_USER=wecheckin \
WECHECKIN_DATABASE_PASSWORD='change-me' \
WECHECKIN_DATABASE_DBNAME=wecheckin \
./bin/wecheckin
```

4. 验证健康检查。

```bash
curl -I http://127.0.0.1:8083/health
curl -I http://127.0.0.1:8083/ready
```

5. 在备份库、测试库或维护窗口执行一次初始化/迁移。

当前项目通过独立脚本执行迁移和种子数据，已执行记录保存在 `schema_migrations`：

```bash
cd backend
bash init.sh
```

`cmd/maintenance` 的配置文件读取规则和环境变量覆盖方式见 [后端版本化迁移说明](../backend/migrations/README.md#配置读取规则)。

迁移成功并验证业务后，生产常态运行只启动后端服务，不需要再执行初始化脚本。

6. 验证新旧接口。

```bash
curl -sS -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8083/api/v2/home
curl -sS -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8083/home/list
curl -I http://127.0.0.1:8083/swagger/index.html
```

管理后台登录接口应使用：

```text
POST /api/v2/admin/auth/login
```

旧页面或旧小程序如果仍在使用历史路径，需要同时验证对应旧接口。

## Nginx 兼容配置

生产 Nginx 至少需要保留 `/api/` 代理：

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8083;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    client_max_body_size 32m;
}
```

如果仍需兼容旧前端或旧小程序，再保留旧路径代理：

```nginx
location ~ ^/(passport|home|upload|uploads|user_form_fields|survey|exam|dict|geo|fav|news|enroll|event|swagger|test)(/|$) {
    proxy_pass http://127.0.0.1:8083;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    client_max_body_size 32m;
}
```

注意：`location /api/` 中应使用 `proxy_pass http://127.0.0.1:8083;`，不要写成 `proxy_pass http://127.0.0.1:8083/;`。

## Docker Compose 单点部署

项目的 `backend/docker-compose.yml` 已包含 MySQL、Redis、backend 和 Nginx，适合单机容器化部署：

```bash
cd backend
cp .env.example .env
docker-compose up -d
```

旧库升级时，`.env` 只保留服务运行所需的 MySQL、Redis、端口和 CORS 配置。数据库结构升级请在维护窗口进入容器或宿主机执行 `backend/init.sh`。

## 回滚策略

- 如果没有执行 `backend/init.sh` 或手工 DDL，回滚通常只需要切回旧后端、旧前端和旧 Nginx 配置。
- 如果已经执行 `backend/init.sh` 或手工 DDL，回滚前先评估是否需要恢复 MySQL 备份。
- 回滚时注意保留升级期间新增的上传文件，避免恢复数据库后文件记录和磁盘文件不一致。

## 上线前检查清单

- [ ] MySQL 已完整备份，并确认备份文件可恢复。
- [ ] `backend/uploads` 或 Docker 上传卷已备份。
- [ ] 新版本后端使用正确的 `WECHECKIN_DATABASE_*` 和 `WECHECKIN_REDIS_*`。
- [ ] 迁移前已备份，且初始化/迁移只在维护窗口执行 `backend/init.sh`。
- [ ] Nginx 已代理 `/api/`，并且 `proxy_pass` 没有结尾斜杠。
- [ ] 如需兼容旧入口，Nginx 已保留旧路径代理。
- [ ] `/health` 和 `/ready` 均返回 200。
- [ ] 新管理后台 `/api/v2/admin/*` 接口可用。
- [ ] 旧前端或旧小程序仍使用的历史接口可用。
- [ ] 已准备好旧版本程序、旧 Nginx 配置和数据库恢复方案。
