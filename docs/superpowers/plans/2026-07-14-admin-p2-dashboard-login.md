# Admin P2 Dashboard Login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 继续优化 WeCheckin 管理后台 P2 级体验，完善控制台、登录页和 fallback 菜单入口。

**Architecture:** 保留已有 Vue 3、Element Plus、admin 全局样式和接口层。新增一条静态检查防回归，局部调整三个低风险界面文件，不改变接口和权限流程。

**Tech Stack:** Vue 3、Vue Router、Element Plus、TypeScript、Vite、Node.js 静态检查。

---

## File Structure

- Create: `admin/scripts/check-p2-ui.mjs`
  - 检查控制台是否包含欢迎区、指标网格、快捷入口和待办区域。
  - 检查登录页是否包含品牌面板、图标输入和响应式样式。
  - 检查 fallback 菜单是否补全考试记录、考试统计、考试表单入口。
- Modify: `admin/package.json`
  - 新增 `check:p2-ui`。
- Modify: `scripts/check.sh`
  - 接入 `npm --prefix admin run check:p2-ui`。
- Modify: `admin/src/views/dashboard/index.vue`
  - 从两块简单数字卡片升级为运营概览。
- Modify: `admin/src/views/login/index.vue`
  - 从单卡片登录页升级为双栏品牌登录页。
- Modify: `admin/src/router/adminRoutes.ts`
  - 补全考试分组 fallback 菜单。

## Tasks

- [x] 编写中文设计文档。
- [x] 新增 P2 UI 静态检查，并确认当前实现红灯。
- [x] 优化控制台运营概览。
- [x] 优化登录页品牌和表单体验。
- [x] 补全 fallback 考试菜单入口。
- [x] 接入项目级检查。
- [x] 运行 `npm --prefix admin run check:p2-ui`、`bash scripts/check.sh`、`npm --prefix admin run build`、`CHECK_ADMIN_BUILD=1 bash scripts/check.sh`，并用浏览器验证 `/login` 渲染。
