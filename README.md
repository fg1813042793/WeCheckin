# MY打卡 - 多功能打卡应用

一个包含移动端、小程序端、PC 管理台和 Go 后端的打卡/活动/问卷/考试系统。

## 项目结构

```text
WeCheckin/
├── backend/                 # Go 后端服务
│   ├── cmd/main.go          # 服务入口
│   ├── cmd/maintenance/     # 初始化/迁移维护入口
│   ├── config/              # 运行配置，默认端口 8083
│   ├── internal/            # 业务模块、模型、处理器、中间件
│   ├── docs/swagger/        # Swagger 文档产物
│   ├── migrations/          # 版本化 SQL 迁移
│   ├── scripts/             # 数据和运维脚本
│   ├── Dockerfile
│   └── docker-compose.yml
├── admin/                   # PC 管理台，Vue 3 + Vite + Element Plus
│   ├── src/api/
│   ├── src/router/
│   ├── src/views/
│   └── package.json
├── frontend/                # uni-app 客户端，支持 H5/App/微信小程序
│   ├── pages/
│   ├── components/
│   ├── api/
│   ├── config/
│   ├── pages.json
│   ├── manifest.json
│   └── package.json
├── docs/                    # 使用和调试文档
├── openspec/                # 规格驱动变更记录
├── scripts/                 # 本地质量门禁和部署配置检查
├── go.mod                   # 后端 Go module
└── README.md
```

## 功能模块

- 用户认证、管理员认证、JWT/Redis token 管理
- 打卡项目、打卡记录、报名表单和统计导出
- 通知公告、收藏、首页配置
- 赛事活动、参与用户、动态和成绩管理
- 问卷系统、答卷、统计报表、题库和资源管理
- 在线考试、考试记录、题库和资源管理
- PC 管理台的用户、部门、岗位、角色和统一权限管理

## 技术栈

### 后端

- Go 1.24 module
- Hertz (CloudWeGo)
- GORM + MySQL
- Redis
- Swagger

### PC 管理台

- Vue 3
- Vite
- TypeScript
- Element Plus
- Axios
- ECharts / vue-echarts

### 移动端 / 小程序端

- uni-app
- Vue 3
- Vite
- HBuilderX / DCloud 工具链

## 快速开始

### 后端服务

默认配置文件位于 `backend/config/config.yaml`，默认监听端口为 `8083`。

首次部署或版本升级需要先执行一次初始化/迁移脚本。脚本会按 `backend/migrations` 文件名顺序执行 SQL，并通过 `schema_migrations` 记录已执行任务，重复执行时会自动跳过历史任务。

```bash
cd backend
bash init.sh
```

```bash
cd backend
go mod download
go run ./cmd
```

也可以使用启动脚本：

```bash
cd backend
bash start.sh
```

Swagger 入口：

```text
http://localhost:8083/swagger/index.html
```

### PC 管理台

管理台开发服务默认端口为 `3000`，请求通过 `admin/vite.config.ts` 代理到 `http://localhost:8083`。

```bash
cd admin
npm install
npm run dev
```

构建：

```bash
cd admin
npm run build
```

### uni-app 客户端

客户端 API 地址在 `frontend/config/index.js` 中配置。

```bash
cd frontend
npm install
npm run dev:h5
```

常用脚本：

```bash
npm run dev:h5
npm run dev:app
npm run dev:mp-weixin
npm run build:h5
npm run build:app
npm run build:mp-weixin
```

## 配置说明

- `backend/config/config.yaml`：后端默认配置。
- `backend/config/config.dev.yaml`：开发环境覆盖配置，可在 `backend` 目录通过 `go run ./cmd -env dev` 合并读取。
- `backend/config/config.example.yaml`：安全示例配置，适合复制为新环境的起点。
- `frontend/config/index.js`：uni-app 客户端 API 地址、版本和缓存配置。
- `admin/vite.config.ts`：管理台开发代理配置。
- `docs/API_V2.md`：当前 `/api/v2` RESTful 接口、调用入口和 Swagger 更新说明。

uni-app 客户端默认读取 `frontend/.env` 中的 `VITE_API_BASE_URL` 作为后端 API 地址。可以复制 `frontend/.env.example` 为 `frontend/.env` 后按环境修改：

```bash
cd frontend
cp .env.example .env
```

本地 H5 调试可使用 `http://localhost:8083`；真机或小程序调试需要填写设备可访问的局域网、测试环境或生产环境地址。

后端支持 `WECHECKIN_` 环境变量覆盖 YAML 配置，环境变量优先级高于配置文件。敏感值建议通过环境变量注入，例如：

```bash
cd backend
WECHECKIN_DATABASE_PASSWORD='your-db-password' \
WECHECKIN_REDIS_PASSWORD='your-redis-password' \
go run ./cmd
```

常用覆盖项包括：

- `WECHECKIN_SERVER_PORT`
- `WECHECKIN_DATABASE_HOST`
- `WECHECKIN_DATABASE_PORT`
- `WECHECKIN_DATABASE_USER`
- `WECHECKIN_DATABASE_PASSWORD`
- `WECHECKIN_DATABASE_DBNAME`
- `WECHECKIN_REDIS_HOST`
- `WECHECKIN_REDIS_PORT`
- `WECHECKIN_REDIS_PASSWORD`
- `WECHECKIN_REDIS_DB`
- `WECHECKIN_CORS_ALLOW_ORIGINS`

`WECHECKIN_CORS_ALLOW_ORIGINS`、`WECHECKIN_CORS_ALLOW_METHODS` 和 `WECHECKIN_CORS_ALLOW_HEADERS` 使用英文逗号分隔多个值。

后端主服务启动不再执行 GORM AutoMigrate、版本化迁移或种子数据。需要补齐表结构、权限菜单或基础配置时，统一在维护窗口执行 `backend/init.sh`。

密码写入已使用 bcrypt。历史 MD5 密码仍可登录，登录成功后会自动升级为 bcrypt 哈希。

后端健康检查：

```text
http://localhost:8083/health
http://localhost:8083/ready
```

用户扩展表单字段通过 `setups` 表中的 `SETUP_USER_FORM_FIELDS` 配置项保存。后端启动不会清理旧 `user_form_fields` 表；如需迁移历史数据，应使用单独迁移脚本处理。

## API v2

当前 PC 管理后台、uni-app 客户端和 uni-app 移动端管理页已迁移到 `/api/v2`：

- 客户端接口：`/api/v2`
- 后台管理接口：`/api/v2/admin`
- 前端调用入口：`admin/src/api/index.ts`、`frontend/api/index.js`、`frontend/api/admin.js`

后台管理只使用 `/api/v2/admin/*` 路由和统一权限体系；旧版 `/admin/*` 后台接口已不再作为兼容入口。`/passport/*`、`/home/*`、`/survey/*`、`/exam/*` 等历史客户端路径如仍存在，仅用于兼容旧页面和小程序旧代码。新增接口和新增页面应使用 v2 路由，并同步更新 Swagger。

详见：[API v2 接口说明](docs/API_V2.md)。

Swagger 重新生成命令：

```bash
cd backend
swag init -g main.go --dir ./cmd --parseDependency --output docs/swagger
```

## 测试

推荐使用项目级检查脚本验证当前关键回归检查：

```bash
bash scripts/check.sh
```

该脚本会使用项目内 `.cache/go-build` 作为 Go 构建缓存，并在结束时自动清理 `.cache/`。默认覆盖范围包括：

- 后端启动入口、Token、handler、service、配置和 formkit 测试。
- uni-app 客户端 API 配置、v2 接口迁移、请求层、登录态、生产日志和 FormRender 逻辑静态检查。
- 管理后台 v2 接口迁移、请求层、导航配置、基础 UI 壳、用户列表、图标运行时和 Vite 构建分包静态检查。

如需在同一条命令中追加构建验证，可以使用以下开关：

```bash
CHECK_FRONTEND_BUILD=1 bash scripts/check.sh
CHECK_ADMIN_BUILD=1 bash scripts/check.sh
CHECK_BUILDS=1 bash scripts/check.sh
```

其中 `CHECK_BUILDS=1` 会同时运行前端 H5 构建和管理后台构建。

如需单独运行 formkit 测试，也可以执行：

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/app/formkit/...
```

如需覆盖当前所有稳定检查，建议使用：

```bash
bash scripts/verify-local.sh
```

### 性能回归排查

项目根目录提供关键接口性能基线脚本，默认只输出告警，不阻断命令：

```bash
npm run check:performance
```

常用环境变量：

- `WECHECKIN_PERF_BASE_URL`：后端地址，默认 `http://127.0.0.1:8083`。
- `WECHECKIN_ADMIN_TOKEN`：后台接口 token，按项目前端格式传原始 token；脚本会兼容去掉可选的 `Bearer ` 前缀。
- `WECHECKIN_USER_TOKEN`：客户端接口 token，缺省时仍会尝试访问公开列表；业务返回 `code != 0` 会视为失败。
- `WECHECKIN_DINGTALK_TOKEN`：钉钉 H5 接口 token，按项目前端格式传原始 token。
- `WECHECKIN_PERF_STRICT=1`：开启严格模式，接口超阈值或不可访问时退出非 0。

发布前建议在服务热启动后执行：

```bash
WECHECKIN_PERF_BASE_URL=http://127.0.0.1:8083 \
WECHECKIN_ADMIN_TOKEN='your-admin-token' \
WECHECKIN_USER_TOKEN='your-user-token' \
WECHECKIN_DINGTALK_TOKEN='your-dingtalk-token' \
WECHECKIN_PERF_STRICT=1 \
npm run check:performance
```

## 文档

- [中文部署和排障指南](docs/DEPLOYMENT_TROUBLESHOOTING.md)
- [单点 MySQL 部署兼容升级说明](docs/SINGLE_NODE_MYSQL_UPGRADE.md)
- [API v2 接口说明](docs/API_V2.md)
- [HBuilderX Android 调试指南](docs/HBUILDER_DEBUG.md)
- [测试数据说明](docs/TEST_DATA.md)
- [项目维护说明](docs/project-maintenance.md)
- [性能排查说明](docs/performance/README.md)
- [后端 DTO 与 Context 规范](docs/backend-dto-context-guidelines.md)
- [后端版本化迁移说明](backend/migrations/README.md)
- `docs/CC打卡小程序安装使用手册.docx`

## 部署

完整部署步骤、环境变量清单、Nginx 示例和常见问题请查看：

- [中文部署和排障指南](docs/DEPLOYMENT_TROUBLESHOOTING.md)
- [单点 MySQL 部署兼容升级说明](docs/SINGLE_NODE_MYSQL_UPGRADE.md)

后端提供 Dockerfile 和 docker-compose 示例，可作为容器化部署起点：

```bash
cd backend
cp .env.example .env
docker-compose up -d
```

Dockerfile、Compose 和 Nginx 示例已统一到后端端口 `8083`。`backend/.env.example` 提供 Docker 部署环境变量样板，复制为 `.env` 后必须修改 MySQL、Redis 密码、域名和 CORS。

如果 MySQL、Redis 和 Nginx 已由外部环境提供，只需要单独部署后端容器，可使用后端独立 Compose：

```bash
cd backend
cp .env.backend.example .env
docker compose -f docker-compose.backend.yml up -d --build
```

钉钉 H5 也支持单独部署为静态站点容器，默认把 `/api/v2` 同源代理到外部后端：

```bash
cd dingtalk-h5
cp .env.docker.example .env
docker compose -f docker-compose.h5.yml up -d --build
```

已有 MySQL 单点部署升级时，接口层以 `/api/v2` 为准，数据库结构升级仍建议先备份，并在备份库或维护窗口执行 `backend/init.sh`。迁移完成后，常态运行只启动服务，不再夹带初始化任务。

Docker Compose 中 MySQL、Redis、后端和 Nginx 均配置了 healthcheck，backend 依赖 MySQL/Redis healthy，Nginx 依赖 backend healthy。该配置依赖 Docker Compose v2 的 `condition: service_healthy`。

Compose 已配置 Docker 日志轮转，Nginx 会记录并透传 `X-Request-ID`。查看容器日志：

```bash
cd backend
bash scripts/docker-logs.sh backend
bash scripts/docker-logs.sh nginx
```

Docker 场景数据库备份和恢复脚本：

```bash
cd backend
bash scripts/docker-backup.sh
bash scripts/docker-restore.sh backups/wecheckin-YYYYMMDD-HHMMSS.sql
```
