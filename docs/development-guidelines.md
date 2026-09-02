# WeCheckin 统一开发规范

## 1. 目的与适用范围

本文档定义 WeCheckin 全仓库的共同开发约定，适用于需求分析、代码实现、接口联调、数据迁移、测试、文档和发布前验证。各端技术细节由专项规范补充，根规范不重复子项目内容。

## 2. 项目结构与边界

| 目录 | 定位 | 主要技术栈 | 对外接口 |
| --- | --- | --- | --- |
| `backend/` | 后端服务和维护任务 | Go 1.24、Hertz、GORM、MySQL、Redis | `/api/v2`、`/api/v2/admin`、`/api/v2/dingtalk/h5` |
| `admin/` | PC 管理后台 | Vue 3、Vite、TypeScript、Element Plus | `/api/v2/admin` |
| `frontend/` | H5/App/微信小程序客户端 | uni-app、Vue 3、Vite | `/api/v2` |
| `h5app/` | 钉钉 H5 权威前端源 | uni-app、Vue 3、TypeScript、Pinia、uView Pro | `/api/v2/dingtalk/h5` |

业务应放在它所属的模块中。跨端只通过明确的 API、数据库迁移、权限编码和文档契约协作，不通过复制业务实现保持一致。

## 3. 规范层级与文档导航

规范优先级为：用户或任务的明确指令 > 根级共同规范 > 模块专项规范 > 相邻代码的现有模式。

修改子项目前必须阅读对应文档：

- [Backend 开发规范](../backend/docs/development-guidelines.md)
- [Admin 开发规范](../admin/docs/development-guidelines.md)
- [Frontend 开发规范](../frontend/docs/development-guidelines.md)
- [H5App 开发规范](../h5app/docs/development-guidelines.md)

专题约束仍以现有文档为准：

- [API v2 接口说明](API_V2.md)
- [后端 DTO 与 Context 规范](backend-dto-context-guidelines.md)
- [权限编码与前端同步](PERMISSION_CODE_FRONTEND_SYNC.md)
- [项目维护说明](project-maintenance.md)
- `docs/architecture/` 下已批准的架构设计。

## 4. 通用开发原则

### 4.1 改动边界

- 先读实际代码、配置和测试，不根据目录名或旧文档猜测当前行为。
- 使用仓库已有的工具、组件、响应结构和抽象，避免在局部功能再造一套模式。
- 一次只处理一个明确目标。无关重构、依赖升级、全局格式化应拆成独立变更。
- 工作区中既有修改视为用户资产，不覆盖、回退或删除。

### 4.2 目录与抽象

- 新文件放入已有职责目录，命名与相邻文件保持一致。
- 一个模块只承担一类职责。文件大到无法快速识别边界时，优先按业务职责拆分。
- 只有当抽象能减少实质重复、隔离外部边界或形成稳定契约时才新增。

### 4.3 依赖与生成物

- 新增依赖前确认现有依赖、标准库或局部工具无法满足需求，并评估大小、许可、维护和跨端成本。
- 依赖变更必须同步对应 lockfile，不混用 npm、pnpm 或 yarn。
- `node_modules/`、`dist/`、`unpackage/`、`.cache/`、`backend/bin/`、`backend/uploads/` 和 `backend/logs/` 不作为源码提交。
- 生成文件必须由对应生成命令更新，不手工改一份而遗漏其他产物。

## 5. 跨端接口与权限契约

### 5.1 API 版本

- 新客户端接口：`/api/v2/*`。
- 新管理后台接口：`/api/v2/admin/*`。
- 钉钉 H5 接口：`/api/v2/dingtalk/h5/*`。
- 历史路径只作兼容；新页面和新能力不得继续扩散旧路径。

接口修改时必须一次核对：

1. HTTP method 和 path。
2. path/query/body 参数的来源、类型、默认值和校验。
3. 响应 DTO、错误码和用户可见消息。
4. Swagger 声明与生成产物。
5. Admin、Frontend 和 H5App 中的调用方与类型。
6. 后端权限目录和前端权限判断。

### 5.2 权限

- 前端按钮可见性是交互层约束，后端路由权限是安全边界；两者必须同时存在。
- 修改内置权限 key 时，按“后端声明 -> 数据迁移 -> 前端筛选/常量 -> 重新登录验证”的顺序同步。
- 修改菜单或动作权限时，同时检查控件显示、事件处理和 API 访问，避免只隐藏按钮或只拦截接口。

### 5.3 数据格式

- 列表、详情、提交、登录等稳定接口使用命名 DTO，不扩散无边界动态 map。
- 分页字段、时间时区、空值、布尔值、状态码和枚举值必须保持跨端一致。
- 前端不依赖未在契约中声明的临时字段；后端不在无迁移计划时删除已发布字段。

## 6. 数据库、配置与安全

### 6.1 数据库迁移

- 结构和内置数据变更统一放入 `backend/migrations/`，按时间顺序命名。
- 迁移必须考虑重复执行、历史数据、索引成本、空值和上一版本兼容性。
- 主服务启动不运行 AutoMigrate、版本化迁移或种子数据；迁移在维护窗口执行 `backend/init.sh`。

### 6.2 配置与密钥

- 默认配置、开发覆盖和生产配置的职责保持分离。
- 密码、token、签名密钥、数据库凭证和真实企业配置不进入 Git。
- 后端敏感值优先通过 `WECHECKIN_` 环境变量注入；前端环境变量只存放可公开配置，不存放服务端密钥。

### 6.3 错误与日志

- 后端日志保留调试所需的请求和内部错误上下文，不记录密码、token 或完整敏感请求体。
- 用户可见错误要可操作，但不泄露 SQL、堆栈、内部路径或依赖细节。
- 前端公共错误处理优先由请求层承担，页面只处理可恢复的业务错误，避免同一错误重复弹出。

## 7. UI 与跨端体验

- 复用各端现有组件库、主题、布局和交互模式，不在单页引入独立的视觉语言。
- 每个异步页面或操作都要考虑加载、空数据、错误、无权限、禁用、重试和成功反馈。
- 按钮显示条件和点击处理必须使用同一权限/状态条件，不能只做视觉禁用。
- 定尺结构要有稳定尺寸约束，文字不溢出、不遮挡操作。
- 移动端和跨端功能要在目标平台检查安全区、输入法、滚动、导航、设备权限和网络异常。

## 8. 测试与验证

验证范围与变更风险匹配：纯函数或局部规则先跑聚焦测试，共享契约、权限、路由、数据库或用户关键流程变更要扩大回归范围。

| 影响范围 | 最低验证 |
| --- | --- |
| Backend 局部包 | `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./path/to/package -count=1` |
| Backend 共享契约/迁移/路由 | 相关包测试，然后 `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./...` |
| Admin | `cd admin && npm run check:all` |
| Frontend | `cd frontend && npm run check:all`；平台相关变更再跑相应 build |
| H5App | `cd h5app && pnpm lint && pnpm type-check`；页面/配置变更再跑 `pnpm build:h5` |
| 跨端变更 | 各受影响端的检查 + 关键用户流程冒烟验证 |
| 根级稳定检查 | `bash scripts/verify-local.sh` |

`scripts/verify-local.sh` 当前不包含 H5App。静态检查和构建通过不等于 UI 或真机流程已验证。

## 9. 文档、生成物与提交

- 新功能或行为变更同步更新 README、API、架构、部署或模块规范中受影响的部分。
- 修改 v2 路由或 Swagger 声明后，在 `backend` 目录执行：

  ```bash
  swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger
  ```

- 提交信息建议使用 `feat:`、`fix:`、`docs:`、`refactor:`、`perf:`、`test:`、`build:` 或 `chore:` 前缀。
- 不在未经用户要求时为了“整洁”重写历史、清理未提交文件或改变分支。

## 10. 功能交付检查清单

1. 功能边界与用户要求一致，没有夹带无关重构。
2. 后端授权、前端可见性和操作处理一致。
3. 接口、DTO/类型、Swagger 和调用方已同步。
4. 数据库变更具有版本化迁移和历史数据考量。
5. 异常、空、禁用、无权限和重试状态已处理。
6. 已运行与风险匹配的自动检查和必要的人工冒烟。
7. 交付说明列出已验证项、未验证项和剩余风险。
