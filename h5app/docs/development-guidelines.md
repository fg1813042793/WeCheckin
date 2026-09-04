# 开发规范

本文档记录当前项目的开发约定。规则以现有代码和配置为准，目标是让后续业务开发保持一致、可维护、可跨端验证。

## 上位规范与项目定位

- 本模块同时遵守 [WeCheckin 统一开发规范](../../docs/development-guidelines.md)。根规范管理跨端契约和共同工作流，本文档管理 H5App 技术栈和实现细节。
- `h5app` 是当前 WeCheckin 钉钉 H5 前端的权威源。分析或修改当前钉钉 H5 行为时以本目录为准，不以历史 `dingtalk-h5` 实现代替当前源码。
- 任务的明确要求优先级最高；本文档可具体化根规范，但不得降低安全、接口、权限和验证要求。

## 项目基线

- 技术栈：uni-app、Vue 3、TypeScript、Vite、Pinia、uView Pro、UnoCSS、SCSS。
- 包管理器：只使用 `pnpm`。项目声明为 `pnpm@10.34.5`，Node.js 要求 `>=18.0.0`。
- TypeScript：保持 `strict: true`，新增代码必须能通过 `pnpm type-check`。
- Lint：使用根目录 `eslint.config.js` 中的 `@uni-helper/eslint-config`，新增代码必须能通过 `pnpm lint`。
- 路径别名：使用 `@/` 指向 `src/`，不要新增重复别名。

## 目录职责

- `src/pages/`：uni-app 页面入口。当前钉钉 H5 采用单入口 `src/pages/index/index.vue`，业务内容通过应用菜单动态渲染；新增业务页面优先挂到应用内容路由，不要再单独注册到 `src/pages.json`，除非明确要创建独立 uni 页面。
- `src/components/`：可复用组件。组件目录使用 kebab-case，如 `app-page`，组件文件保留 `组件名.vue` 的局部结构。
- `src/composables/`：跨页面复用的组合式逻辑，命名为 `useXxx.ts`。
- `src/stores/`：Pinia store，按业务域拆分，导出统一放在 `src/stores/index.ts`。
- `src/api/`：接口访问层。按业务域拆分为 `useXxxApi.ts` 或 `xxxApi.ts`，页面不直接拼接复杂请求逻辑。
- `src/common/`：全局配置、常量、拦截器、主题等基础设施。
- `src/locale/lang/`：语言包。所有业务可见文案优先进入语言包，中文和英文 key 保持同步。
- `src/static/`：静态资源。图片、字体等资源放这里，避免散落在页面目录。

## Vue 单文件组件

- 默认使用 `<script setup lang="ts">`、`<template>`、`<style lang="scss" scoped>` 的顺序。
- 页面根节点优先使用应用壳或业务页面容器，复用全局导航、过渡和页面背景；当前钉钉 H5 不使用原生 `tabBar`。
- 组合式 API 优先使用 `ref`、`computed`、普通函数。复杂逻辑抽到 `composables` 或 store，不堆在页面文件里。
- 类型导入使用 `import type`。外部包导入在前，`@/` 内部导入在后。
- `defineProps`、`defineEmits` 要给出明确类型和默认值。复杂 prop 类型使用接口或 `PropType`。
- 模板中避免复杂表达式和业务判断，超过一行的判断或计算抽成 `computed` 或函数。
- 只在需要 uni-app 组件选项时使用 `defineOptions`，例如 `virtualHost`、`styleIsolation`。
- 页面和组件中不保留调试日志。确需临时日志时用局部 eslint disable，并在提交前移除。

## TypeScript

- 禁止新增无必要的 `any`。接口返回、表单数据、列表项、组件 prop 都应定义明确类型。
- 优先使用 `unknown` 承接外部不可信数据，再做收窄。
- 可复用类型放在对应业务模块附近；跨模块共享的类型可新增 `src/types/`。
- store 中的持久化数据必须可序列化，避免存函数、组件实例、循环引用对象。
- 异步函数必须明确处理加载态、失败态和最终收尾逻辑。

## uni-app 跨端规则

- 优先使用 `uni.*` API 和 uni-app 组件，不直接依赖浏览器 DOM、`window`、`document`。
- 平台差异使用条件编译：

```ts
// #ifdef H5
window.open(url, '_blank')
// #endif
// #ifndef H5
uni.setClipboardData({ data: url })
// #endif
```

- `window`、`document`、小程序专属 API、App 专属 API 都必须包在条件编译里。
- 布局尺寸优先使用 `rpx`，固定边框、极细线或第三方组件要求时可用 `px`。
- 跨端布局优先使用 `flex`，不要使用 `float`。
- 独立页面跳转使用 `uni.navigateTo`、`uni.redirectTo`、`uni.reLaunch`，路径必须来自 `src/pages.json`；钉钉 H5 应用内菜单跳转统一通过 `src/stores/appContent.ts` 切换内容 key。
- 跳转失败要给用户反馈，当前项目可使用 `$u.toast` 或页面内 `u-toast`。
- 页面生命周期从 `@dcloudio/uni-app` 导入，组件内部生命周期从 `vue` 导入。

## uView Pro 使用规则

- uView Pro 已在 `src/main.ts` 注册，并通过 easycom 自动引入组件。页面中直接使用 `u-*` 组件，不手动注册。
- 优先使用 uView Pro 的基础组件、反馈组件、弹窗、表单、导航等能力，再考虑自定义组件。
- 新增业务页面和组件中的按钮、输入框、文本域、选择器、表单校验、弹窗、标签、加载态和图标优先使用 uView Pro 组件，例如 `u-button`、`u-input`、`u-textarea`、`u-select`、`u-form`、`u-popup`、`u-modal`、`u-tag`、`u-loading`、`u-icon`。
- 不直接新增原生 `button`、`input`、`select`、`textarea`；确需原生能力时必须在代码旁说明原因，并保持当前 uView Pro 视觉规范。
- 主题颜色优先使用 `$u-*` SCSS 变量、`var(--u-*)` CSS 变量或 `useTheme()`，不要在业务页面散落硬编码品牌色。
- 暗黑模式和主题切换统一通过 `useTheme()`，不要绕过主题系统直接改全局样式。
- 国际化统一通过 `useLocale()` 的 `t()`，新增业务文案同步维护 `zh-CN.json` 和 `en-US.json`。
- 复杂组件样式覆盖优先使用 `custom-class` 搭配 `:deep()`；简单动态样式可使用 `custom-style`。
- 自定义组件样式应挂在页面或组件根 class 下，避免污染 uView Pro 全局样式。

## 状态管理

- 全局状态使用 Pinia setup store，命名为 `useXxxStore`。
- store 只承担跨页面共享状态和对应 action。单页面状态保留在页面或 composable 中。
- 修改状态通过 store action 完成，避免多个页面重复写同一段状态变更逻辑。
- 使用 `persist: true` 前确认数据适合持久化。敏感数据不放入普通持久化 store。
- 不要在普通业务操作里调用 `uni.clearStorage()` 清空全局缓存。除明确的“重置/退出登录”场景外，应只清理本模块 key。

## 接口与错误处理

- 请求统一走 `uview-pro` 的 `http` 和 `src/common/http.interceptor.ts`。
- 基础地址、公共 header、token、loading、toast 等公共策略放在拦截器或统一配置中。
- API 层负责请求路径、参数、返回类型。页面层只消费业务方法。
- 新增接口必须定义请求参数和响应类型，避免 `http.get<any>()` 继续扩散。
- 需要静默请求时显式传入 meta 配置，避免页面和拦截器行为不一致。
- 拦截器负责统一转换网络错误和 HTTP 非 2xx 错误，页面只处理业务可恢复错误。

### 删除语义

- `h5app` 面向钉钉 H5 普通用户，页面中的单条删除、批量删除、撤销后删除等操作必须调用 `/api/v2/dingtalk/h5/*` 软删除接口；后端只更新删除标记和审计字段，不物理移除业务记录。
- 软删除后的记录不再出现在普通列表和详情入口中，但必须保留业务数据、流程节点、任务、流转记录、评论、通知和历史版本，不能仅由前端从本地列表隐藏来代替后端删除。
- 删除权限、记录归属、当前状态和重复删除必须由后端校验；H5App 根据接口结果更新页面，不自行构造删除状态或调用 `/api/v2/admin/*` 接口。
- 管理后台 `admin` 中的删除属于物理删除，由 `/api/v2/admin/*` 接口独立实现。H5App 软删除和后台物理删除不得复用为同一个无差别删除接口。
- 新增删除能力时必须在 API 类型、权限编码和接口文档中明确 `soft delete` 或 `physical delete`，并覆盖已删除记录不可见、重复删除、无权限和关联历史保留等场景。

## WeCheckin 集成契约

- 钉钉 H5 业务接口使用 `/api/v2/dingtalk/h5/*`，新请求集中在 `src/api/` 并统一经过 `src/common/http.interceptor.ts`。
- 登录或启动 bootstrap 返回的 `menus`、`buttonPermissionKeys`、`apiPermissionKeys` 和 `permissionVersion` 是前端导航与操作权限的共享契约，不在页面另维护一套脱节的权限源。
- 菜单展示优先使用 bootstrap 返回的菜单、路径和图标。新增图标 key 时，同时核对后端默认声明、管理后台可选项和 H5App 图标映射。
- 修改 `dingtalk_h5:menu:*`、`dingtalk_h5:button:*` 或 `dingtalk_h5:api:*` 时，同步后端权限目录、SQL 迁移、管理后台配置、H5App 常量/判断和重新登录验证。完整流程见 [权限编码与前端同步](../../docs/PERMISSION_CODE_FRONTEND_SYNC.md)。
- 按钮可见性、disabled 状态和点击处理使用同一权限/业务状态条件；后端 API 权限仍是最终安全边界。
- 后端路由、请求/响应 DTO、菜单结构或权限字段变更时，同步 H5App API 类型、store/composable 和所有页面消费方，不用临时字段兼容掩盖契约漂移。

## 样式规范

H5App 的详细布局、响应式、表格、搜索栏、弹窗与验收规则见 [H5App UI 布局与样式规范](./ui-layout-style-guidelines.md)。该文档是 UI 改动的必读规范，与 `src/common/ui-layout.scss` 共同构成页面样式基线。

- 页面样式默认使用 `scoped` SCSS。
- 全局样式只放在 `src/common/style.scss` 或 `src/uni.scss`，并保持最小化。
- 页面根 class 使用页面名，如 `.settings-page`。组件内部 class 推荐 BEM 风格，如 `.section-card__header`。
- 简单间距、布局、图标类可以使用 UnoCSS；复杂结构和主题相关样式使用 SCSS。
- 颜色、背景、边框优先使用 uView Pro 主题变量，业务状态色使用 `primary`、`success`、`warning`、`error` 语义。
- 新增页面要检查 H5 和目标小程序/App 下文字是否溢出、按钮是否拥挤、底部 tabbar 是否遮挡内容。

### PC 小屏幕统一样式（强制）

- H5App 的 PC 小屏幕指视口宽度 `769px - 1023px`；`<= 768px` 继续使用手机布局，`>= 1024px` 使用标准 PC 布局。不得用视口宽度直接缩放字号，也不得通过整体 `transform: scale()` 缩放业务页面。
- PC 小屏幕的公共规则统一维护在 `src/common/ui-layout.scss` 的 `app-pc-control-scope` 下。应用壳业务内容区已挂载该 class；新增独立内容容器必须复用它，不得在多个业务页面复制同一套 uView Pro 内部选择器。
- 该断点下使用固定 `px` 控制桌面密度，避免 uView Pro 默认 `rpx` 随桌面视口放大：正文基准 `14px`，输入值和提示文字 `13px`，字数统计 `11px/16px`；普通、中号、小号、迷你按钮高度分别为 `36px`、`34px`、`32px`、`28px`，普通输入控件高度为 `36px`。
- 普通图标固定为 `18px`，按钮内图标为 `15px`，空状态等强调图标为 `28px`，加载图标使用 `18px` 或 `24px`。纯图标按钮使用 `app-icon-button`，默认 `32x32px`，紧凑操作叠加 `app-icon-button--small` 使用 `28x28px`。
- uView Pro 的输入框、文本域、placeholder、字数统计、按钮、标签、单选/复选、开关、分页和加载态都必须由公共规则覆盖完整；不能只调整外层容器而遗漏 uni-app H5 生成的内部节点或组件内联字号。
- 流程运行时表单根节点必须同时使用 `app-workflow-form` 和 `app-pc-control-scope`。发起页、任务页、详情抽屉/弹窗中的字段标签、普通输入、多行文本、明细列表、日期选择、附件列表、空状态和底部操作必须共享同一套紧凑密度。
- 顶部多页签在空间不足时必须保持单行、禁止压缩，通过横向区域左右滚动或固定的左右切换按钮访问被遮挡页签；切换按钮必须垂直居中，且不能覆盖页签文字和关闭按钮。
- 头像、品牌标识、文件缩略图必须使用专用 class 明确宽高、`flex-shrink` 和 `object-fit`，不得被业务内容的通用图标规则放大，也不得用全局 `image` 选择器统一改尺寸。
- 使用 Teleport 的 `u-popup`、`u-modal` 等业务弹层必须通过 `custom-class` 加入 `app-pc-control-scope`。小屏幕下内容无法完整展示的复杂详情抽屉应切换为居中弹窗，限制在视口内；标题和底部操作固定，只允许内容区纵向滚动。
- PC Select 打开时必须比较视口及滚动容器的上下可见空间；下方空间不足且上方更充足时向上展开。移动端继续使用适合触摸操作的选择器，不复用 PC 浮层交互。
- 修改上述尺寸、断点或覆盖范围时，必须同步 `src/common/ui-layout.scss`、[H5App UI 布局与样式规范](./ui-layout-style-guidelines.md) 和 `scripts/check-ui-style-guidelines.mjs`。至少运行 `pnpm check:ui-style`、`pnpm type-check` 和 `pnpm build:h5`，并验证一个 PC 小屏幕、一个标准 PC 和一个手机尺寸。

### PC 端选择与日期控件

- 普通单选项在 PC 端必须从字段下方或附近直接展开，不得在 PC 端使用底部弹层式选择器。`u-select` 只能用于移动端，或由自适应组件在移动端分支中使用。
- 日期和日期范围必须使用日历式交互；PC H5 优先使用锚定在输入框附近的日历，或浏览器原生 `input[type=date]`，不使用底部滚轮或普通列表代替日历。
- 当 uView Pro 组件的默认交互与 PC 端不匹配时，允许在 H5 专用界面使用原生 `select` 和 `input[type=date]`；代码中必须注明原因，并保持高度、边框、字号、禁用态和焦点态与当前主题一致。
- 共享跨端页面需通过自适应组件或条件编译提供移动端交互，不得因 PC 优化破坏 App/小程序编译。
- 交付前至少验证一次 PC 宽屏和一次移动窄屏，确认下拉层/日历位置、遮罩、键盘操作、安全区和内容滚动正常。

## 国际化

- 业务可见文本默认写入语言包，不在页面中硬编码。
- key 按模块组织，例如 `home.xxx`、`about.settingsPage.xxx`。
- `zh-CN.json` 和 `en-US.json` 必须保持相同 key 结构。
- 动态文本使用占位符，例如 `t('common.welcome', { name })`。
- 技术调试文本、一次性开发注释可以不进入语言包，但不能出现在最终用户界面。

## 资源与配置

- 静态资源放入 `src/static/`，命名使用 kebab-case。
- 新增环境变量时使用 `.env.*`，不要提交包含真实密钥的本地配置。
- `manifest.json`、`pages.json`、`vite.config.ts` 是跨端关键配置，修改后至少跑一次对应平台构建。
- `src/theme.json` 和 `src/common/uview-pro.theme.ts` 属于主题配置，修改时要同步验证亮色、暗色和至少一个业务页面。

## 提交与验证

提交前至少执行：

```bash
pnpm lint
pnpm type-check
```

修改钉钉应用入口、菜单、权限或绩效模块时再执行：

```bash
pnpm check:dingtalk-module
```

修改通用工作流页面或调用时再执行：

```bash
pnpm check:workflow-module
```

涉及页面、样式、路由或跨端配置时再执行：

```bash
pnpm check:ui-style
pnpm build:h5
```

涉及具体平台时执行对应命令，例如：

```bash
pnpm build:mp-weixin
pnpm build:app-android
pnpm build:app-ios
```

提交信息使用约定式前缀：

- `feat:` 新功能
- `fix:` 缺陷修复
- `docs:` 文档
- `style:` 样式或格式调整
- `refactor:` 重构
- `perf:` 性能优化
- `test:` 测试
- `build:` 构建或依赖
- `chore:` 维护

## 新增功能流程

1. 确认功能归属的页面、组件、store、API 和语言包。
2. 先补类型和 API 边界，再写页面交互。
3. 使用 uView Pro 组件和主题变量搭建界面，交互控件默认选 `u-*` 组件。
4. 同步维护 `pages.json`、语言包和必要静态资源。
5. 运行 `pnpm lint` 和 `pnpm type-check`。
6. 涉及 UI 时在 H5 至少做一次手动冒烟，涉及小程序/App 时补对应平台验证。

## 当前模板特别注意

- 当前部分文件仍带有示例性质，例如 `src/api/useUserApi.ts` 中的 `any` 和 `src/common/http.interceptor.ts` 中的示例 token/baseUrl。新增业务代码不要照抄这些临时写法，应按本规范补齐类型和真实配置边界。
- 根目录已有 `llms.txt`、`llms-full.txt` 未跟踪文件，除非任务明确要求，不要修改或清理它们。
