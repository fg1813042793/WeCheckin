# Admin P1 UI Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 优化 WeCheckin 管理后台 P1 级界面体验，让整体壳层、列表页结构和用户管理操作区更适合日常后台管理。

**Architecture:** 保留 Vue 3、Element Plus 和现有路由/权限体系。新增轻量全局后台样式和静态检查；局部改造 layout 与用户管理页，不触碰后端接口和复杂设计器业务逻辑。

**Tech Stack:** Vue 3、Vue Router、Element Plus、TypeScript、Vite、Node.js 静态检查。

---

## File Structure

- Create: `admin/scripts/check-ui-shell.mjs`
  - 检查后台壳层包含折叠侧栏、面包屑、菜单状态和全局样式导入。
- Create: `admin/scripts/check-user-list-ui.mjs`
  - 检查用户管理列表使用统一页面类，操作列已收敛为主操作和更多菜单。
- Modify: `admin/package.json`
  - 增加 `check:ui-shell` 和 `check:user-list-ui`。
- Modify: `scripts/check.sh`
  - 接入新的后台 UI 静态检查。
- Create: `admin/src/styles/admin.css`
  - 提供后台页面、卡片、工具栏、表格操作和分页的统一样式。
- Modify: `admin/src/main.ts`
  - 导入 `admin/src/styles/admin.css`。
- Modify: `admin/src/views/layout/index.vue`
  - 增加折叠侧栏、面包屑、菜单加载/失败/空状态和更稳的主内容区域。
- Modify: `admin/src/views/user/index.vue`
  - 使用统一列表结构；操作列从大宽度按钮堆叠改为主按钮 + 更多菜单。

## Tasks

- [x] 编写中文设计文档。
- [x] 新增后台 UI 静态检查，并确认当前实现红灯。
- [x] 新增全局后台样式并导入。
- [x] 优化后台 layout 壳层。
- [x] 优化用户管理列表页结构和操作列。
- [x] 接入项目级检查。
- [x] 运行 `npm --prefix admin run check:ui-shell`、`npm --prefix admin run check:user-list-ui`、`bash scripts/check.sh`、`npm --prefix admin run build`、`CHECK_ADMIN_BUILD=1 bash scripts/check.sh`。
