# Admin 开发规范

## 1. 适用范围

本文档适用于 `admin/` PC 管理后台。同时遵守 [WeCheckin 统一开发规范](../../docs/development-guidelines.md)、[API v2 接口说明](../../docs/API_V2.md) 和 [权限编码与前端同步](../../docs/PERMISSION_CODE_FRONTEND_SYNC.md)。

## 2. 技术与工具基线

- Vue 3、Vite、TypeScript strict、Vue Router、Element Plus、Axios。
- 包管理使用 npm，依赖变更同步 `package-lock.json`，不在 `admin/` 中使用 pnpm 或 yarn。
- `@/` 映射到 `src/`，不新增重复别名。
- 当前构建通过 `vue-tsc -b && vite build` 完成，项目尚未配置 ESLint/Prettier，不得把它们声称为已有质量门禁。

## 3. 目录职责

- `src/api/`：后端 API 访问、请求/响应类型和业务接口分组。页面不直接拼接复杂请求。
- `src/router/`：路由表、导航守卫和页面 meta。新页面必须同步路由和菜单/标签行为。
- `src/views/`：业务页面，按业务域分目录。大页面的子组件和纯逻辑应拆到当前业务目录。
- `src/components/`：跨页面可复用组件，不放入只有一个页面使用的业务细节。
- `src/utils/`：请求层、纯函数和稳定基础工具。
- `src/constants/`：跨页面枚举、常量和契约 key；不放会随请求变化的运行时数据。
- `src/styles/`：全局样式、变量和可复用样式基础，业务页面样式留在组件内。
- `src/styles/admin-ui-tokens.css`：新 Admin UI 基础层的颜色、间距、控件高度、容器尺寸、圆角和断点契约。
- `src/components/admin-ui/`：新页面共用的页面壳、标题、搜索栏、表格容器、弹窗和抽屉；不承载业务字段、权限和请求。
- `src/examples/admin-ui/`：不注册路由、不挂菜单、不请求接口的 Admin UI 样板页。

## 4. Vue 与 TypeScript

- 新 Vue SFC 默认使用 Composition API 和 `<script setup lang="ts">`，但修改旧页面时优先保持局部风格，不为统一语法重写整页。
- 保持 `tsconfig.json` 的 `strict: true`，不通过放宽编译选项解决局部类型问题。
- props、emits、API 参数/响应、表单、表格行、树节点和组件 expose 必须有明确类型。
- 外部不可信数据先用 `unknown` 承接并收窄，不新增无边界 `any`。
- 复杂派生状态放在 `computed`，副作用保持在显式函数或生命周期中，模板不承载多层业务判断。
- 组件只在能形成稳定输入/输出契约时拆分，不为拆文件而制造大量透传 props。
- 已登记的遗留大组件执行“只减不增”预算，`npm run check:component-complexity` 会阻止继续增长；新增 API、类型、纯逻辑或独立交互应拆入 `src/api`、`src/types`、同业务目录组件或 composable。

## 5. API 与错误处理

- 新后台请求只使用 `/api/v2/admin/*`，历史路径不作为新能力入口。
- 请求统一经过 `src/utils/request.ts` 和 `src/api/`，页面不重复处理 base URL、token、登录过期和通用错误弹窗。
- 修改后端 DTO 时同步 API 类型和所有调用页面，不在页面用多组兼容字段掩盖契约漂移。
- 通用网络/HTTP 错误由请求层反馈，页面只处理可恢复的业务错误，避免同一错误弹出两次。
- 重复提交、搜索和数据刷新根据场景使用禁用态、请求取消、去重或幂等 key，不只依赖用户不会连点。

## 6. 路由、菜单与权限

- 新页面同时核对路由、菜单、面包屑、页签标题、刷新恢复和深链接访问。
- 新业务路由必须声明 `meta.adminUi`，并使用新 Admin UI 公共组件：

```ts
meta: {
  title: '示例列表',
  menuPath: '/example',
  adminUi: {
    version: 1,
    pattern: 'filter-list',
  },
}
```

- `adminUi.pattern` 只允许 `list`、`filter-list`、`form`、`detail`、`workspace`。
- 新路由页面必须从 `@/components/admin-ui` 导入并渲染 `AdminPageShell`。
- `scripts/admin-ui-legacy-routes.json` 只记录当前旧路由基线；不得将新路由加入基线以绕过 UI 契约检查。
- 菜单/按钮权限与 API 权限是不同契约。控件可见性、disabled 状态和事件处理使用同一业务权限条件。
- 页面不认为“按钮隐藏”就完成授权，后端路由权限必须同步。
- 权限 key 或前缀变更时，同步后端声明、SQL 迁移、权限树筛选和页面硬编码判断。

## 7. Element Plus 与 UI

- Admin UI 的完整结构、尺寸和交互规则见 [Admin UI 开发规范](./ui-guidelines.md)，新页面必须同时遵守本章和该文档。
- 表单、表格、弹窗、消息、加载和常见操作优先复用 Element Plus 和项目已有组件。
- 新页面使用 `AdminPageShell`、`AdminPageHeader`、`AdminSearchBar`、`AdminTablePanel`、`AdminDialog` 和 `AdminDrawer` 组合页面结构，不重写页面级容器 CSS。
- 公共组件负责布局、间距、高度、滚动和响应式；业务页面只负责字段、列、权限、请求、业务校验和业务状态。
- 颜色、间距、控件高度、表头高度、圆角、页面边距和容器宽度优先使用 `admin-ui-tokens.css` 中的变量，不在业务页面重复声明。
- 断点契约为：`>= 1200px` 标准 PC、`768px - 1199px` 窄屏 PC/平板、`< 768px` 移动端。
- 页面必须覆盖加载、空数据、请求失败、无权限、表单校验、提交中和成功反馈。
- 表格稳定列宽、操作列和滚动范围；长文本使用折行、截断加 tooltip 或详情查看，不覆盖相邻内容。
- 搜索栏的查询与重置行为由 `AdminSearchBar` 统一；普通列表优先 1-3 个条件，复杂筛选可折叠。
- 短表单和小型详情使用 `AdminDialog`，较长详情和需保留列表上下文的操作使用 `AdminDrawer`；移动端复杂交互优先独立路由页。
- 设计器、表单编辑器和图形工具需要稳定容器尺寸，动态内容不得造成工具栏或画布跳动。
- 涉及重要数据的删除、发布、撤回和完成操作需要明确确认和可理解反馈。
- UI 改动至少检查主流桌面宽度和窄窗口，复杂交互使用浏览器手动冒烟。
- 参考 `src/examples/admin-ui/SimpleListExample.vue` 和 `AdvancedFilterListExample.vue`；样板只用于参考，不得挂载到路由或菜单。

## 8. 性能与依赖

- 大库、编辑器、图表、地图和工作流图形优先按路由或功能懒加载。
- 新增依赖前检查现有 Element Plus、Vue Router、项目工具和原生 Web API 是否已能满足需求。
- 保持 Vite 分包策略和 bundle 预算，不因为一个局部组件将大依赖并入首屏。

## 9. 验证命令

完整 Admin 门禁：

```bash
cd admin
npm run check:all
```

`check:all` 包含请求/API v2、构建配置、导航、权限、组件复杂度、重要页面结构、TypeScript 构建和 bundle 预算等检查。局部改动可先运行对应 `check:*`，交付前根据风险扩大到 `check:all`。

新页面或路由的 Admin UI 契约可单独检查：

```bash
cd admin
npm run check:admin-ui-contract
```

该检查只对比旧路由基线之后的新路由，不扫描或要求改造旧业务页面。新路由缺少 `meta.adminUi` 或未使用 `AdminPageShell` 时，`check:all` 失败。

单独构建：

```bash
cd admin
npm run build
```

改动页面、样式、路由、设计器或复杂交互时，构建/静态检查之外还要在浏览器验证实际流程。

## 10. 交付检查

1. API 调用集中在 `src/api` 和请求层，不扩散历史路径。
2. TypeScript strict 下通过构建，没有用新 `any` 掉过契约问题。
3. 路由、菜单、按钮、事件处理和后端权限已联动。
4. 新路由已声明 `meta.adminUi`，页面已使用 `AdminPageShell` 和适用的 Admin UI 公共组件。
5. 页面级布局和尺寸来自公共组件与 token，业务页面未重复实现基础结构。
6. 加载、空、错误、禁用、无权限和成功状态已处理。
7. 已运行相关 `check:*`、构建和必要的浏览器冒烟，结果如实报告。
