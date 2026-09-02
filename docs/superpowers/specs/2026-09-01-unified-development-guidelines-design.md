# WeCheckin 统一开发规范设计

## 目标

为 WeCheckin 建立“根级总则 + 四端专项规范”的分层规范体系，同时服务人工开发和 AI 代理修改。规范必须以当前仓库实际目录、技术栈、脚本和接口契约为准，不把尚未存在的工具链描述为已启用。

## 范围

本次生成或更新以下文件：

- `AGENTS.md`：根级 AI 执行规则和模块规范导航。
- `docs/development-guidelines.md`：项目统一开发规范。
- `backend/docs/development-guidelines.md`：Go 后端专项规范。
- `admin/docs/development-guidelines.md`：PC 管理后台专项规范。
- `frontend/docs/development-guidelines.md`：uni-app 客户端专项规范。
- `h5app/docs/development-guidelines.md`：在现有完整规范上补充项目级契约和导航。

本次不修改业务代码、构建配置、依赖、质量门禁脚本、CI 或 Git Hook。

## 分层与优先级

1. 用户或任务明确指令优先级最高。
2. 修改任一模块时，必须同时遵守根级总则和该模块专项规范。
3. 模块规范可对根级总则做技术栈相关的具体化，但不得降低安全、契约和验证要求。
4. 现有专项文档继续有效，例如 DTO/Context、API v2、权限编码、架构设计和部署文档；新规范通过链接引用，不复制整篇内容。

## 根级规范设计

### `AGENTS.md`

保持简洁且可执行，包含：

- 四端职责和对应规范路径。
- 改动聚焦、保护用户未提交修改、不擅自扩展范围。
- 新接口使用 `/api/v2` 或 `/api/v2/admin`，同步权限声明、Swagger 和前端调用。
- 数据库结构变更使用版本化迁移，不依赖主服务启动时 AutoMigrate。
- 按受影响模块运行真实存在的验证命令，并如实报告未验证项。

### `docs/development-guidelines.md`

作为团队主文档，包含：

- 项目结构与四端边界。
- 规范优先级与文档导航。
- 分支、改动范围、生成物、依赖和敏感配置的共同规则。
- API v2、权限、Swagger、响应结构和错误信息等跨端契约。
- 迁移、配置、日志、安全和兼容性的共同要求。
- 按改动类型选择验证命令的矩阵。
- 约定式提交前缀，但本次不增加强制 Hook。

## 四端专项规范

### Backend

基于 Go 1.24、Hertz、GORM、MySQL 和 Redis，明确：

- `cmd`、`internal/routes`、`handler`、`service`、`modules`、`model`、`support`、`pkg` 和 `test` 的责任边界。
- handler 只处理 HTTP 边界，业务编排进入 service/application，领域状态与持久化分离。
- 命名 DTO、Context 传递、多表写入事务、错误转换和日志边界。
- API v2 路由、权限目录、Swagger 产物必须同步。
- SQL 迁移的命名、幂等性、历史数据和回滚风险考量。
- Go 文件执行 `gofmt`；局部测试后再按风险扩大到 `go test ./...`。
- 文件行数不能太大，结构要分层

### Admin

基于 Vue 3、Vite、TypeScript 和 Element Plus，明确：

- `api`、`router`、`views`、`components`、`utils`、`constants` 和 `styles` 的职责。
- 保持 TypeScript strict，避免新增无边界的 `any`，复用类型与 API 封装。
- 页面不重复实现请求、权限和公共错误处理。
- 新后台接口只使用 `/api/v2/admin`，按钮显示与操作处理同时守权限。
- UI 变更需检查加载、空、错误、无权限、禁用和窄屏状态。
- 验证使用现有 `npm run check:all` 和 `npm run build`，不声称已有 ESLint/Prettier。
-  文件行数不能太大，结构要分层
-  UI要统一

### Frontend

基于当前根目录式 uni-app/Vue 3 客户端，明确：

- `pages`、`components`、`api`、`config`、`utils`、`mixins`、`static` 和 `uni_modules` 的职责。
- 新接口使用 `/api/v2`，统一经过请求封装，不在页面散落 base URL、token 和公共错误处理。
- 避免无条件使用 DOM 或平台专属 API，平台差异使用 uni-app 条件编译。
- 页面、路由、manifest、权限和静态资源变更需要目标端构建或真机验证。
- 验证使用现有 `npm run check:all` 与相关平台构建，不声称已有 lint 或类型门禁。
-  文件行数不能太大，结构要分层
-  UI要统一

### H5App

保留现有 uni-app + Vue 3 + TypeScript + uView Pro 完整规范，仅做以下校准：

- 声明同时受根级总则约束。
- 将 `h5app` 明确为当前钉钉 H5 的权威前端源。
- 补充 `/api/v2/dingtalk/h5` 契约、权限 key 同步和菜单/bootstrap 依赖。
- 保持 pnpm 10.34.5、Node.js 18+、strict TypeScript、ESLint、uView Pro、主题和国际化规则。
- 验证使用 `pnpm lint`、`pnpm type-check`、模块检查和按影响范围选择的平台构建。
-  文件行数不能太大，结构要分层
-  UI要统一

## 跨端契约

统一规范将强调以下联动要求：

- 接口路径、请求/响应 DTO 或前端类型、Swagger 和调用方同步。
- 后端权限声明、数据迁移、管理后台配置和调用端权限判断同步。
- 已发布流程定义与历史实例绑定，对定义、表单和运行时存储的修改不得破坏历史版本。
- 共享字段、状态、枚举、时间和分页语义必须在跨端修改时一起核对。

## 验证设计

文档生成后执行以下只读或文档级检查：

1. 确认六个目标文件存在，且根文档可导航到四端文档。
2. 检查文档中提到的脚本均存在于对应 `package.json` 或仓库脚本。
3. 搜索占位符、相互冲突的包管理器和已过时路径。
4. 运行 `git diff --check`。
5. 由于本次只修改 Markdown，不运行业务测试或构建，也不改变任何现有门禁状态。

## 完成标准

- 团队成员能从根规范快速定位到任一端的专项规范。
- AI 代理能从根 `AGENTS.md` 获知改动范围、必读文档和验证命令。
- 四端文档反映当前实际技术栈和工具能力，不伪造尚未存在的 lint、CI 或测试门禁。
- 文档不覆盖或模糊现有 API v2、权限、DTO/Context、迁移和架构专项文档。
