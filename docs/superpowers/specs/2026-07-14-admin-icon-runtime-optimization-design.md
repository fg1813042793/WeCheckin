# Admin 图标运行时优化设计

## 背景

管理后台当前在 `admin/src/main.ts` 中使用 `import * as ElementPlusIconsVue from '@element-plus/icons-vue'`，并把图标库中的所有图标注册为全局组件。这样实现简单，但会让后台入口承担不必要的图标库加载成本。后台已经通过路由懒加载和 Vite 分包做了部分优化，下一步应继续减少首屏入口的全量导入。

## 目标

1. 入口文件不再全量导入 Element Plus 图标库。
2. 提供一个受控图标注册表，只注册当前后台导航、布局、列表页和登录页实际需要的常用图标。
3. 动态菜单图标通过注册表解析，未识别图标优雅降级为不显示图标，不影响菜单文字和路由。
4. 保留 `IconPicker` 的完整图标选择能力；它只在菜单管理页懒加载时使用，不进入后台首屏入口。

## 非目标

- 不改 Element Plus 组件注册方式。
- 不引入自动按需导入插件。
- 不重构所有页面的图标写法。
- 不改变后端菜单接口和图标字段。

## 方案

新增 `admin/src/icons.ts`：

- 显式导入后台常用图标。
- 导出 `adminIconMap`、`resolveAdminIcon()` 和 `registerAdminIcons()`。
- `main.ts` 调用 `registerAdminIcons(app)`，不再遍历整个图标库。
- `layout/index.vue` 的动态菜单图标通过 `resolveAdminIcon(item.icon)` 渲染，避免未知字符串组件警告。

新增 `admin/scripts/check-icon-runtime.mjs`：

- 禁止 `main.ts` 出现 `import * as ElementPlusIconsVue`。
- 要求 `main.ts` 使用 `registerAdminIcons(app)`。
- 要求 `icons.ts` 暴露 `ADMIN_ICON_NAMES`、`resolveAdminIcon` 和 `registerAdminIcons`。

## 验证

- `npm --prefix admin run check:icon-runtime`
- `bash scripts/check.sh`
- `CHECK_ADMIN_BUILD=1 bash scripts/check.sh`
