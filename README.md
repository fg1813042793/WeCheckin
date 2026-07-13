# MY打卡 - 多功能打卡应用

一个包含移动端、小程序端、PC 管理台和 Go 后端的打卡/活动/问卷/考试系统。

## 项目结构

```text
WeCheckin/
├── backend/                 # Go 后端服务
│   ├── cmd/main.go          # 服务入口
│   ├── config/              # 运行配置，默认端口 8083
│   ├── internal/            # 业务模块、模型、处理器、中间件
│   ├── docs/swagger/        # Swagger 文档产物
│   ├── scripts/             # 数据和迁移脚本
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
- PC 管理台的用户、部门、角色、菜单和权限管理

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

```bash
cd backend
go mod tidy
go run cmd/main.go
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
- `backend/config/config.dev.yaml`：开发环境覆盖配置，可通过 `go run cmd/main.go -env dev` 合并读取。
- `frontend/config/index.js`：uni-app 客户端 API 地址、版本和缓存配置。
- `admin/vite.config.ts`：管理台开发代理配置。

后端启动时会自动执行 GORM AutoMigrate，并初始化部分系统配置和菜单数据。

## 测试

当前项目测试主要集中在后端 formkit 子系统：

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/formkit/...
```

如果测试后生成 `.cache/`，可以删除该目录；它已加入 `.gitignore`。

## 文档

- [HBuilderX Android 调试指南](docs/HBUILDER_DEBUG.md)
- [测试数据说明](docs/TEST_DATA.md)
- `docs/CC打卡小程序安装使用手册.docx`

## 部署

后端提供 Dockerfile 和 docker-compose 示例：

```bash
cd backend
docker-compose up -d
```

部署前请根据目标环境调整 MySQL、Redis、端口、上传目录和反向代理配置。
