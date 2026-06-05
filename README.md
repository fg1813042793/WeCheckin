# MY打卡 - 多功能打卡应用

一个基于uni-app开发的多平台打卡应用，支持Android、iOS、HarmonyOS和微信小程序。

## 项目结构

```
WeCheckin/
├── frontend/                 # 前端代码
│   ├── pages/              # 页面组件
│   ├── components/         # 通用组件
│   ├── utils/              # 工具函数
│   ├── api/                # API接口
│   ├── config/             # 配置文件
│   ├── static/             # 静态资源
│   ├── store/              # 状态管理
│   ├── App.vue             # 根组件
│   ├── main.js             # 入口文件
│   ├── pages.json          # 页面配置
│   ├── manifest.json       # 应用配置
│   └── package.json        # 依赖管理
├── backend/                # 后端代码
│   ├── cmd/                # 命令行入口
│   ├── internal/           # 内部包
│   │   ├── config/         # 配置管理
│   │   ├── model/          # 数据模型
│   │   ├── service/        # 业务逻辑
│   │   └── handler/        # 请求处理
│   ├── api/                # API路由
│   ├── config.yaml         # 配置文件
│   ├── go.mod             # Go模块
│   ├── docker-compose.yml  # Docker配置
│   └── Dockerfile         # Docker构建
├── docs/                  # 文档
│   ├── HBUILDER_DEBUG.md  # HBuilder调试指南
│   └── MY打卡小程序安装使用手册.docx  # 安装手册
├── project.config.json    # 项目配置
├── project.private.config.json  # 项目私有配置
└── README.md             # 项目说明
```

## 功能特性

- 📱 多平台支持：Android、iOS、HarmonyOS、微信小程序
- 🔐 用户认证：JWT token认证
- 📝 打卡管理：创建、编辑、删除打卡活动
- 📊 数据统计：打卡统计、用户分析
- 🔔 消息通知：实时通知推送
- 🎨 现代化UI：基于uni-ui的精美界面

## 技术栈

### 前端
- **框架**：uni-app (Vue 2)
- **UI库**：uni-ui
- **状态管理**：Vuex
- **HTTP客户端**：Axios
- **构建工具**：HBuilderX

### 后端
- **语言**：Go 1.19+
- **框架**：Hertz (CloudWeGo)
- **数据库**：MySQL + GORM
- **缓存**：Redis
- **认证**：JWT
- **容器化**：Docker

## 快速开始

### 前端开发

1. 进入前端目录：
```bash
cd frontend
```

2. 安装依赖：
```bash
npm install
```

3. 开发调试：
```bash
npm run dev
```

4. 构建项目：
```bash
npm run build
```

### 后端开发

1. 进入后端目录：
```bash
cd backend
```

2. 安装依赖：
```bash
go mod tidy
```

3. 运行服务：
```bash
go run cmd/main.go
```

4. 使用Docker运行：
```bash
docker-compose up -d
```

## 配置说明

### 前端配置

- `config/index.js`：API地址配置
- `manifest.json`：应用配置
- `pages.json`：页面路由配置

### 后端配置

- `config.yaml`：服务配置
- 数据库配置：MySQL连接信息
- Redis配置：缓存服务配置

## 开发指南

### 添加新页面

1. 在 `frontend/pages/` 下创建新页面目录
2. 在 `frontend/pages.json` 中添加页面配置
3. 创建对应的Vue组件

### 添加新API

1. 在 `frontend/api/` 下添加API接口文件
2. 在 `backend/api/` 下添加对应的API处理函数
3. 在 `backend/internal/service/` 下添加业务逻辑

## 部署说明

### 前端部署

1. 构建项目：
```bash
cd frontend && npm run build
```

2. 将 `dist` 目录部署到Web服务器

### 后端部署

1. 构建Docker镜像：
```bash
cd backend && docker build -t wecheckin-backend .
```

2. 使用Docker Compose部署：
```bash
docker-compose up -d
```

## 许可证

MIT License

## 联系方式

如有问题，请提交Issue或联系开发团队。