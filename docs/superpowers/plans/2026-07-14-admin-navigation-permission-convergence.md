# 管理后台路由和权限配置收敛 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将管理后台静态路由配置集中管理，并增加自动化检查防止导航配置再次分叉。

**Architecture:** 新增 `admin/src/router/adminRoutes.ts`，集中导出后台子路由 `adminChildRoutes`。`admin/src/router/index.ts` 只负责装配路由和守卫，`admin/src/views/layout/index.vue` 只展示后端返回的权限菜单。新增 `admin/scripts/check-navigation.mjs`，在项目级检查中验证路由配置和布局引用保持收敛。

> 当前状态：旧默认全量菜单兜底已废弃，后台菜单只由统一权限接口 `/api/v2/admin/me/menus` 控制。

**Tech Stack:** Vue 3、Vue Router、Element Plus、Node.js 脚本检查。

---

## File Structure

- Create: `admin/src/router/adminRoutes.ts`
  - 保存管理后台主布局下的静态子路由。
- Modify: `admin/src/router/index.ts`
  - 使用 `adminChildRoutes` 替代内联 children 数组。
- Modify: `admin/src/views/layout/index.vue`
  - 使用接口菜单渲染侧栏。
  - 移除模板中的硬编码兜底菜单块。
- Create: `admin/scripts/check-navigation.mjs`
  - 检查路由和菜单配置集中化。
  - 检查布局不再导入或展示默认全量菜单。
- Modify: `admin/package.json`
  - 增加 `check:navigation`。
- Modify: `scripts/check.sh`
  - 纳入 `npm --prefix admin run check:navigation`。

---

## Tasks

- [x] 增加失败检查：`check:navigation` 必须要求存在 `adminRoutes.ts`、`adminChildRoutes`，并要求布局通过 `displayMenuTree` 渲染授权菜单。
- [x] 运行 `npm --prefix admin run check:navigation`，确认缺少集中配置时失败。
- [x] 新增 `admin/src/router/adminRoutes.ts`，迁移原有后台 children 路由。
- [x] 修改 `admin/src/router/index.ts`，从 `./adminRoutes` 引入 `adminChildRoutes`。
- [x] 修改 `admin/src/views/layout/index.vue`，以 `displayMenuTree` 渲染接口授权菜单。
- [x] 修改 `admin/package.json` 和 `scripts/check.sh`，纳入导航配置检查。
- [x] 运行 `npm --prefix admin run check:navigation`。
- [x] 运行 `bash scripts/check.sh`。
