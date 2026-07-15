# WeCheckin 中文部署和排障指南

本文档用于从一台干净服务器或本地环境部署 WeCheckin。当前项目包含 Go 后端、Vue 3 管理后台和 uni-app 客户端，推荐先完成后端部署，再部署管理后台和客户端。

## 版本和依赖

推荐环境：

- Go 1.24.x
- Node.js 20.x 或更高的当前 LTS 版本
- npm 10.x 或更高版本
- MySQL 8.0
- Redis 7.x
- Nginx 1.24 或更高版本

端口约定：

- 后端默认端口：`8083`
- 管理后台开发端口：`3000`
- uni-app H5 开发端口：由 uni CLI 分配，通常为 `5173`
- 生产环境建议由 Nginx 对外暴露 `80` / `443`

## 部署拓扑

推荐生产拓扑：

```text
用户浏览器 / 小程序 / App
        |
      Nginx
        |
  静态资源 admin/dist 或 frontend/dist
        |
  反向代理 API 到 Go 后端 :8083
        |
   MySQL 8 + Redis 7
```

管理后台请求层使用相对路径访问后端，例如 `/admin/login`、`/passport/login`、`/upload`。因此生产环境需要让管理后台静态站点和后端 API 处在同一域名下，或在 Nginx 中把这些路径反向代理到后端。

uni-app 客户端使用 `frontend/.env` 中的 `VITE_API_BASE_URL` 作为后端地址。H5 可以使用同域名地址，小程序和 App 需要配置为设备可访问的 HTTPS 域名。

## 后端配置

首次部署建议从安全示例配置开始：

```bash
cp backend/config/config.example.yaml backend/config/config.yaml
```

至少检查以下配置：

- `server.port`：默认使用 `8083`。
- `database.host` / `database.port` / `database.user` / `database.password` / `database.dbname`。
- `redis.host` / `redis.port` / `redis.password` / `redis.db`。
- `cors.allow_origins`：生产环境不要使用 `*`，建议填写实际前端域名。
- `log.dir`：确认进程有写入权限。
- `oss.local.path`：本地上传目录，默认 `./uploads`。

后端支持 `WECHECKIN_` 环境变量覆盖 YAML 配置。生产环境推荐用环境变量注入敏感值：

```bash
WECHECKIN_SERVER_PORT=8083 \
WECHECKIN_DATABASE_HOST=127.0.0.1 \
WECHECKIN_DATABASE_PORT=3306 \
WECHECKIN_DATABASE_USER=wecheckin \
WECHECKIN_DATABASE_PASSWORD='change-me' \
WECHECKIN_DATABASE_DBNAME=wecheckin \
WECHECKIN_REDIS_HOST=127.0.0.1 \
WECHECKIN_REDIS_PORT=6379 \
WECHECKIN_REDIS_PASSWORD='change-me' \
WECHECKIN_REDIS_DB=0 \
WECHECKIN_CORS_ALLOW_ORIGINS='https://your-domain.example' \
WECHECKIN_AUTO_MIGRATE=false \
./bin/wecheckin
```

多个 CORS 值使用英文逗号分隔：

```bash
WECHECKIN_CORS_ALLOW_ORIGINS='https://admin.example.com,https://h5.example.com'
```

## 后端部署

从仓库根目录构建：

```bash
go mod download
mkdir -p backend/bin
go build -o backend/bin/wecheckin ./backend/cmd
```

准备运行目录：

```bash
mkdir -p backend/logs backend/uploads
```

启动后端：

```bash
cd backend
./bin/wecheckin
```

如果要使用环境覆盖：

```bash
cd backend
WECHECKIN_DATABASE_PASSWORD='change-me' \
WECHECKIN_REDIS_PASSWORD='change-me' \
./bin/wecheckin
```

如需启用在线考试菜单：

```bash
cd backend
./bin/wecheckin -exam
```

如需合并读取 `backend/config/config.dev.yaml`：

```bash
cd backend
./bin/wecheckin -env dev
```

启动校验：

```bash
curl -I http://127.0.0.1:8083/health
curl -I http://127.0.0.1:8083/ready
curl -I http://127.0.0.1:8083/swagger/index.html
```

`/health` 用于进程存活检查；`/ready` 会检查数据库连接。看到 `HTTP/1.1 200 OK` 或可访问 Swagger 页面，说明后端进程已经对外服务。

默认情况下后端启动会执行 GORM AutoMigrate。生产环境如需把迁移纳入发布流程，可设置 `WECHECKIN_AUTO_MIGRATE=false` 关闭启动迁移。

密码存储已升级为 bcrypt。历史 MD5 密码仍可兼容登录，登录成功后后端会自动写回 bcrypt 哈希。

## 管理后台部署

安装依赖并构建：

```bash
npm --prefix admin install
npm --prefix admin run build
```

构建产物位于：

```text
admin/dist
```

生产部署建议：

1. 将 `admin/dist` 发布到 Nginx 静态目录。
2. 为管理后台配置 history fallback，未命中静态文件时返回 `index.html`。
3. 将 `/admin`、`/passport`、`/home`、`/upload`、`/uploads`、`/user_form_fields`、`/survey`、`/exam`、`/dict`、`/geo` 等 API 路径反向代理到后端 `http://127.0.0.1:8083`。
4. 上传文件较大时设置 `client_max_body_size 32m` 或更高。

Nginx 示例：

```nginx
server {
    listen 80;
    server_name admin.example.com;

    root /var/www/wecheckin-admin;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location ~ ^/(admin|passport|home|upload|uploads|user_form_fields|survey|exam|dict|geo)(/|$) {
        proxy_pass http://127.0.0.1:8083;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        client_max_body_size 32m;
    }
}
```

如果管理后台的问卷、考试公开填写页需要生成绝对链接，可在构建前设置：

```bash
VITE_API_BASE=https://admin.example.com npm --prefix admin run build
```

## uni-app 客户端部署

复制环境变量示例：

```bash
cp frontend/.env.example frontend/.env
```

修改 `frontend/.env`：

```env
VITE_API_BASE_URL=https://api.example.com
```

本地 H5 调试：

```bash
npm --prefix frontend install
npm --prefix frontend run dev:h5
```

H5 构建：

```bash
npm --prefix frontend run build:h5
```

微信小程序构建：

```bash
npm --prefix frontend run build:mp-weixin
```

App 构建：

```bash
npm --prefix frontend run build:app
```

小程序和 App 真机调试时，`VITE_API_BASE_URL` 必须是设备可访问的地址。微信小程序生产环境还需要在微信公众平台配置合法 request 域名。

## Docker 示例说明

`backend/Dockerfile`、`backend/docker-compose.yml` 和 `backend/nginx.conf` 已按当前后端端口 `8083` 对齐，可作为容器化部署起点：

```bash
cd backend
cp .env.example .env
docker-compose up -d
```

`backend/.env.example` 是 Docker Compose 环境变量样板。复制为 `.env` 后至少修改：

- `MYSQL_ROOT_PASSWORD`
- `WECHECKIN_DATABASE_PASSWORD`
- `WECHECKIN_REDIS_PASSWORD`
- `WECHECKIN_CORS_ALLOW_ORIGINS`
- `WECHECKIN_AUTO_MIGRATE`

当前 Compose 示例会把后端映射到 `8083:8083`，并通过 Nginx 代理 `/admin`、`/passport`、`/home`、`/upload`、`/survey`、`/exam`、`/health`、`/ready` 等路径。MySQL、Redis、backend 和 Nginx 均配置了 healthcheck，backend 使用 `condition: service_healthy` 等待 MySQL/Redis，Nginx 使用 `condition: service_healthy` 等待 backend。该能力需要 Docker Compose v2。

Compose 已为 MySQL、Redis、backend 和 Nginx 配置 Docker `json-file` 日志轮转。默认值来自 `backend/.env.example`：

```env
WECHECKIN_LOG_MAX_SIZE=20m
WECHECKIN_LOG_MAX_FILE=5
```

Nginx 会在访问日志中记录 `request_id=$request_id`，并向后端透传 `X-Request-ID`。排查单次请求时，可以先从网关日志找到 `request_id`，再按时间窗口对照后端日志。

查看 Docker 服务日志：

```bash
cd backend
bash scripts/docker-logs.sh backend
bash scripts/docker-logs.sh nginx
TAIL=500 bash scripts/docker-logs.sh backend
```

Docker 场景数据库备份：

```bash
cd backend
bash scripts/docker-backup.sh
```

备份文件会写入 `backend/backups/`。如果存在 `backend/uploads/`，脚本也会打包上传目录。

Docker 场景数据库恢复：

```bash
cd backend
bash scripts/docker-restore.sh backups/wecheckin-YYYYMMDD-HHMMSS.sql
```

恢复脚本会要求输入 `RESTORE` 二次确认。恢复前建议先执行一次 `docker-backup.sh`，并确认目标数据库和备份文件匹配。

## 验证命令

基础项目检查：

```bash
bash scripts/check.sh
```

同时验证管理后台构建：

```bash
CHECK_ADMIN_BUILD=1 bash scripts/check.sh
```

同时验证 uni-app H5 构建和管理后台构建：

```bash
CHECK_BUILDS=1 bash scripts/check.sh
```

单独验证后端：

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/cmd ./backend/pkg/tokenutil ./backend/internal/app/handler ./backend/internal/app/service ./backend/internal/config ./backend/internal/app/formkit/...
```

单独验证管理后台：

```bash
npm --prefix admin run build
```

单独验证前端配置：

```bash
npm --prefix frontend run check:config
npm --prefix frontend run check:request
npm --prefix frontend run check:logs
npm --prefix frontend run check:auth
npm --prefix frontend run check:formkit-logic
```

单独验证 Docker 部署配置：

```bash
node scripts/check-deploy-config.mjs
```

## 上线检查清单

- MySQL 数据库和账号已创建，后端配置可连接。
- Redis 可连接，密码和 DB 与配置一致。
- `backend/logs`、`backend/uploads` 可写。
- 生产 CORS 已限制为真实域名。
- `/health` 和 `/ready` 均返回 200。
- Docker 部署已从 `backend/.env.example` 复制 `.env` 并修改默认密码。
- Docker Compose 的 healthcheck 均为 healthy。
- 上线前已执行 `backend/scripts/docker-backup.sh` 或已有其他备份方案。
- 管理后台静态文件已发布，并配置 history fallback。
- 管理后台 API 路径已反向代理到后端。
- uni-app 的 `VITE_API_BASE_URL` 指向真实可访问域名。
- 小程序合法域名已在平台侧配置。
- `bash scripts/check.sh` 通过。
- 管理后台上线前建议额外跑 `CHECK_ADMIN_BUILD=1 bash scripts/check.sh`。

## 常见问题和排障

### 后端启动提示数据库连接失败

检查 MySQL 地址、端口、账号、密码和数据库名：

```bash
mysql -h 127.0.0.1 -P 3306 -u wecheckin -p wecheckin
```

确认 `WECHECKIN_DATABASE_*` 环境变量没有覆盖成错误值：

```bash
env | grep '^WECHECKIN_DATABASE_'
```

### 后端启动提示 Redis 连接失败

检查 Redis 是否可访问：

```bash
redis-cli -h 127.0.0.1 -p 6379 ping
```

如果 Redis 设置了密码：

```bash
redis-cli -h 127.0.0.1 -p 6379 -a 'change-me' ping
```

### 管理后台接口 404

管理后台生产环境使用相对路径请求 API。确认 Nginx 已代理 `/admin`、`/passport`、`/home`、`/upload`、`/uploads`、`/survey`、`/exam` 等路径到后端。

### 刷新管理后台页面后 404

这是前端 history 路由没有 fallback 到 `index.html`。确认 Nginx 中存在：

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

### 登录后很快失效或被踢下线

检查 Redis 是否稳定、Token 前缀是否一致、后台系统设置中的单点登录配置是否符合预期。也可以检查：

```bash
env | grep '^WECHECKIN_TOKEN_'
```

### 上传失败或大文件失败

检查后端 `uploads` 目录权限和 Nginx 上传大小限制：

```bash
ls -ld backend/uploads
```

Nginx 中设置：

```nginx
client_max_body_size 32m;
```

### 小程序或 App 无法访问接口

确认 `frontend/.env` 中的 `VITE_API_BASE_URL` 是手机可以访问的 HTTPS 地址。微信小程序还需要在微信公众平台配置合法域名，本地 `localhost` 和电脑局域网地址不能用于生产小程序。

### CORS 报错

生产环境需要把前端域名加入 `WECHECKIN_CORS_ALLOW_ORIGINS` 或 `backend/config/config.yaml` 的 `cors.allow_origins`。多个域名用英文逗号分隔。

### 管理后台构建出现 chunk 体积 warning

当前管理后台已配置 Vite `manualChunks`，将 Vue、Element Plus、图表、编辑器、扫码/地图等依赖拆分到独立 vendor chunk，并把项目级检查接入 `CHECK_ADMIN_BUILD=1 bash scripts/check.sh`。如果仍看到第三方库的 PURE 注释提示，通常是依赖包注释位置导致的 Rollup 提示，不会阻止产物生成。
