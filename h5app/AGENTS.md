# AGENTS.md

本文件是给 Codex/AI 代理使用的项目执行规则。完整团队规范见 `docs/development-guidelines.md`。

## 项目事实

- 技术栈：uni-app、Vue 3、TypeScript、Vite、Pinia、uView Pro、UnoCSS、SCSS。
- 包管理器：只使用 `pnpm`，遵守 `packageManager: pnpm@10.34.5`。
- 运行基线：Node.js `>=18.0.0`，pnpm `>=10.0.0`。
- TypeScript 开启 strict，不要放宽 `tsconfig.json`。
- ESLint 通过 `eslint.config.js` 使用 `@uni-helper/eslint-config`。
- 路径别名：`@/` 指向 `src/`。

## 仓库规则

- 改动必须聚焦用户请求。
- 不要修改无关的脏文件或未跟踪文件。尤其不要碰 `llms.txt` 和 `llms-full.txt`，除非用户明确要求。
- 优先沿用现有项目模式，不轻易新增抽象。
- 使用 `pnpm` 脚本，不使用 `npm` 或 `yarn`。
- 没有明确收益时不要新增依赖。

## 文件位置

- 页面放在 `src/pages/<domain>/<page>.vue`，新增页面必须更新 `src/pages.json`。
- 可复用组件放在 `src/components/<kebab-name>/<kebab-name>.vue`。
- 组合式逻辑放在 `src/composables/useXxx.ts`。
- Pinia store 放在 `src/stores/<domain>.ts`，并从 `src/stores/index.ts` 导出。
- API 封装放在 `src/api/`，页面文件不要承载可复用请求逻辑。
- 共享配置、常量、拦截器、主题放在 `src/common/`。
- 用户可见文案放在 `src/locale/lang/zh-CN.json` 和 `src/locale/lang/en-US.json`。

## 编码风格

- Vue SFC 使用 `<script setup lang="ts">`、`<template>`、`<style lang="scss" scoped>` 顺序。
- 使用 Composition API，并保持类型明确。
- 类型导入使用 `import type`。
- 避免新增 `any`，为请求、响应、表单、列表项、prop、事件定义类型。
- 模板表达式保持简单，业务逻辑移到 `computed`、函数、composables、stores 或 API 模块。
- 结束任务前移除调试日志。

## uni-app 规则

- 使用 `uni.*` API 和 uni-app 组件处理跨端能力。
- 平台专属 API 必须使用条件编译注释隔离。
- 常规布局尺寸使用 `rpx`，固定细节尺寸才使用 `px`。
- 跨端页面使用 flex 布局。
- 页面跳转使用 `uni.navigateTo`、`uni.switchTab`、`uni.redirectTo` 或 `uni.reLaunch`，路径必须已声明在 `src/pages.json`。
- 页面生命周期从 `@dcloudio/uni-app` 导入，组件生命周期从 `vue` 导入。

## uView Pro 规则

- uView Pro 已在 `src/main.ts` 注册，直接通过 easycom 使用 `u-*` 组件。
- 优先使用 uView Pro 组件，再考虑自定义 UI。
- 新增业务页面和组件的交互控件优先使用 uView Pro：`u-button`、`u-input`、`u-textarea`、`u-select`、`u-form`、`u-popup`、`u-modal`、`u-tag`、`u-loading`、`u-icon`。
- 不直接新增原生 `button`、`input`、`select`、`textarea`；确需原生能力时必须在代码旁说明原因，并保持现有 uView Pro 视觉规范。
- 主题相关颜色使用 `$u-*`、`var(--u-*)`、`$u.color` 和 `useTheme()`。
- 用户可见文案使用 `useLocale()` 和 `t()`。
- 复杂样式覆盖使用 `custom-class` 加 `:deep()`，简单动态样式可使用 `custom-style`。

## UI 布局与样式

- 新增或修改页面、表格、搜索栏、弹窗或抽屉前，必须阅读 `docs/ui-layout-style-guidelines.md`。
- 新页面优先使用 `src/common/ui-layout.scss` 的设计变量和 `app-*` 基线类，不在业务页面重复定义通用页边距、控件高度、表头和弹窗尺寸。
- 响应式默认使用 `768px`、`1024px` 和 `1440px` 边界；需要内容驱动的额外断点时，必须在样式旁注明原因并补对应尺寸验证。
- 涉及 UI 基线时运行 `pnpm check:ui-style`，并按规范的视口矩阵做浏览器验证。

## 状态和接口

- Pinia setup store 命名为 `useXxxStore`。
- 只持久化可序列化、非敏感状态。
- 普通业务动作不要调用 `uni.clearStorage()`；除明确重置/退出登录外，只清理模块自身 key。
- HTTP 请求统一走 `uview-pro` 的 `http` 和 `src/common/http.interceptor.ts`。
- base URL、认证 header、loading、toast、共享错误处理要集中管理。

## 验证

交付代码改动前运行：

```bash
pnpm lint
pnpm type-check
```

涉及页面、样式、路由、manifest、Vite、主题或跨端配置时，再运行：

```bash
pnpm build:h5
```

涉及特定平台时运行对应构建脚本，例如 `pnpm build:mp-weixin`、`pnpm build:app-android` 或 `pnpm build:app-ios`。
