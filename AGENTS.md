# AGENTS.md

本文件是 WeCheckin 仓库的根级执行规则，适用于人工开发和 AI 代理修改。完整团队规范见 `docs/development-guidelines.md`。

## 项目边界

- `backend/`：Go 后端，提供管理后台、客户端和钉钉 H5 接口。
- `admin/`：Vue 3 + Vite + TypeScript + Element Plus PC 管理后台。
- `frontend/`：uni-app + Vue 3 客户端，面向 H5、App 和微信小程序。
- `h5app/`：uni-app + Vue 3 + TypeScript + uView Pro 钉钉 H5，是当前钉钉 H5 前端的权威源。

## 必读规范

修改代码前，必须阅读根规范和受影响模块的专项规范：

- Backend：`backend/docs/development-guidelines.md`
- Admin：`admin/docs/development-guidelines.md`
- Frontend：`frontend/docs/development-guidelines.md`
- H5App：`h5app/docs/development-guidelines.md`

用户或任务的明确指令优先级最高。模块规范可对根规范做技术栈相关的具体化，但不得降低安全、契约和验证要求。

## 改动规则

- 只修改完成任务必需的文件，不顺手重构无关模块。
- 工作区可能存在用户未提交修改；不覆盖、清理、回退或格式化无关文件。
- 优先沿用现有目录、类型、组件、工具和抽象；只在能明确减少复杂度时新增抽象或依赖。
- 不手工编辑生成产物，包括 `dist/`、`unpackage/` 和 Swagger 生成文件；Swagger 例外情况必须通过规定命令重新生成。
- 代码、配置、数据库和文档的行为描述必须一致；不把设计计划当成已交付实现。

## 跨端契约

- 新客户端接口使用 `/api/v2`，新管理后台接口使用 `/api/v2/admin`，钉钉 H5 使用 `/api/v2/dingtalk/h5`。
- 修改接口时同时核对路由、请求/响应 DTO、Swagger、权限声明和所有调用端。
- 权限变更必须同步后端目录、历史数据迁移、管理端配置和调用端按钮/API 判断。
- 共享字段、枚举、状态、时间、分页和错误语义不得在各端各自演化。
- 已发布流程定义和历史实例绑定到发布版本；修改定义、表单或运行时存储时不得破坏历史版本。

## 数据与安全

- 数据库结构变更使用 `backend/migrations/` 下的版本化 SQL，不依赖主服务启动时 AutoMigrate。
- 多表写入、状态流转和需要原子性的操作必须使用事务。
- 不提交密钥、token、密码、真实内网地址或用户数据。敏感配置通过环境变量或本地未跟踪配置注入。
- 用户输入、远程数据和文件上传都是不可信边界，必须在后端做权限、类型、长度和业务约束校验。

## 验证

按实际影响范围运行：

- Backend：`cd backend && GOCACHE=$PWD/../.cache/go-build go test ./...`
- Admin：`cd admin && npm run check:all`
- Frontend：`cd frontend && npm run check:all`
- H5App：`cd h5app && pnpm lint && pnpm type-check`
- 根级全量复核：`bash scripts/verify-local.sh`

`scripts/verify-local.sh` 当前只覆盖 Backend、Admin 和 Frontend，修改 `h5app` 时必须单独执行它的检查。UI、跨端或设备行为变更不能只依赖静态脚本，必须补浏览器或目标设备验证。

交付时必须说明已运行的命令、结果以及未验证的剩余风险。
