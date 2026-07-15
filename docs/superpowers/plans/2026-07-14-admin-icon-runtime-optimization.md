# Admin Icon Runtime Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 减少管理后台入口文件对 Element Plus 图标库的全量导入，降低首屏运行时负担。

**Architecture:** 新增受控图标注册表，`main.ts` 只注册后台实际需要的图标；layout 动态菜单通过注册表解析图标名称。保留菜单管理页 `IconPicker` 的完整图标选择能力，因为它位于懒加载页面中。

**Tech Stack:** Vue 3、Element Plus Icons、TypeScript、Node.js 静态检查、Vite。

---

## File Structure

- Create: `admin/scripts/check-icon-runtime.mjs`
  - 防止入口文件再次全量导入 Element Plus 图标库。
- Create: `admin/src/icons.ts`
  - 受控图标注册表和解析函数。
- Modify: `admin/src/main.ts`
  - 使用 `registerAdminIcons(app)`。
- Modify: `admin/src/views/layout/index.vue`
  - 动态菜单和折叠按钮通过 `resolveAdminIcon()` 渲染。
- Modify: `admin/package.json`
  - 新增 `check:icon-runtime`。
- Modify: `scripts/check.sh`
  - 接入图标运行时检查。

## Tasks

- [x] 编写中文设计文档。
- [x] 新增图标运行时静态检查，并确认当前实现红灯。
- [x] 新增 `admin/src/icons.ts`。
- [x] 改造 `main.ts` 和 `layout/index.vue`。
- [x] 接入项目级检查。
- [x] 运行 `npm --prefix admin run check:icon-runtime`、`npm --prefix admin run build`、`bash scripts/check.sh`、`CHECK_ADMIN_BUILD=1 bash scripts/check.sh`。
