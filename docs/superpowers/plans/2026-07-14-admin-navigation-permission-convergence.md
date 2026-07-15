# 管理后台路由和权限配置收敛 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将管理后台静态路由和侧栏兜底菜单收敛到同一配置文件，并增加自动化检查防止配置再次分叉。

**Architecture:** 新增 `admin/src/router/adminRoutes.ts`，集中导出后台子路由 `adminChildRoutes` 与兜底菜单 `fallbackMenuItems`。`admin/src/router/index.ts` 只负责装配路由和守卫，`admin/src/views/layout/index.vue` 从同一配置读取兜底菜单；后端返回的角色菜单仍优先展示。新增 `admin/scripts/check-navigation.mjs`，在项目级检查中验证路由配置、兜底菜单和布局引用保持收敛。

**Tech Stack:** Vue 3、Vue Router、Element Plus、Node.js 脚本检查。

---

## File Structure

- Create: `admin/src/router/adminRoutes.ts`
  - 保存管理后台主布局下的静态子路由。
  - 保存接口不可用时使用的兜底侧栏菜单。
- Modify: `admin/src/router/index.ts`
  - 使用 `adminChildRoutes` 替代内联 children 数组。
- Modify: `admin/src/views/layout/index.vue`
  - 使用 `fallbackMenuItems` 和接口菜单统一渲染侧栏。
  - 移除模板中的硬编码兜底菜单块。
- Create: `admin/scripts/check-navigation.mjs`
  - 检查路由和菜单配置集中化。
  - 检查兜底菜单包含关键入口并且布局不再硬编码菜单。
- Modify: `admin/package.json`
  - 增加 `check:navigation`。
- Modify: `scripts/check.sh`
  - 纳入 `npm --prefix admin run check:navigation`。

---

## Tasks

- [x] 增加失败检查：`check:navigation` 必须要求存在 `adminRoutes.ts`、`adminChildRoutes`、`fallbackMenuItems`，并要求布局通过 `displayMenuTree` 渲染。
- [x] 运行 `npm --prefix admin run check:navigation`，确认缺少集中配置时失败。
- [x] 新增 `admin/src/router/adminRoutes.ts`，迁移原有后台 children 路由并定义兜底菜单。
- [x] 修改 `admin/src/router/index.ts`，从 `./adminRoutes` 引入 `adminChildRoutes`。
- [x] 修改 `admin/src/views/layout/index.vue`，以 `displayMenuTree` 统一渲染接口菜单与兜底菜单。
- [x] 修改 `admin/package.json` 和 `scripts/check.sh`，纳入导航配置检查。
- [x] 运行 `npm --prefix admin run check:navigation`。
- [x] 运行 `bash scripts/check.sh`。
