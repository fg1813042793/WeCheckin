# 前端登录态读取封装 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 uni-app 源码中的客户端和管理员登录态读写收敛到统一工具，减少页面直接操作 `token` / `userInfo` / `admin_token` / `admin_info`。

**Architecture:** 新增 `frontend/utils/auth.js`，集中提供客户端登录态、管理员登录态、请求层登录态和用户 ID 获取函数。新增 `frontend/scripts/check-auth.mjs`，扫描 `frontend/App.vue`、`frontend/utils` 和 `frontend/pages`，禁止除 `utils/auth.js` 外直接读写四个登录态 storage key。页面迁移保持业务行为不变：登录页写入、退出清理、请求头读取、管理员首页鉴权、个人中心用户刷新和常见 `getUserId()` 都改为调用工具函数。

**Tech Stack:** uni-app、JavaScript、Node.js 静态检查、Bash 项目检查脚本。

---

## File Structure

- Create: `frontend/utils/auth.js`
  - 管理四个登录态 key。
  - 提供 `getClientAuth`、`setClientAuth`、`clearClientAuth`、`getAdminAuth`、`setAdminAuth`、`clearAdminAuth`、`getClientUserId`、`hasClientAuth`、`hasAnyAuth` 等函数。
- Create: `frontend/scripts/check-auth.mjs`
  - 扫描源码登录态直连写法。
  - 确认请求层使用 `utils/auth.js`。
- Modify: `frontend/utils/request.js`
  - 通过 `getRequestAuthState` 和 `clearRequestAuthState` 获取/清理登录态。
- Modify: `frontend/App.vue`
  - 通过 `hasAnyAuth()` 判断是否需要跳登录页。
- Modify: `frontend/pages/login/login.vue`
  - 登录成功后调用 `setClientAuth`。
- Modify: `frontend/pages/login/login_pwd.vue`
  - 登录成功后调用 `setClientAuth`。
- Modify: `frontend/pages/admin/admin_login.vue`
  - 进入页时调用 `clearAdminAuth`，登录成功后调用 `setAdminAuth`。
- Modify: `frontend/pages/admin/admin_home.vue`
  - 管理员鉴权和退出改用 auth 工具。
- Modify: `frontend/pages/my/my_index.vue`
  - 个人中心加载/刷新/退出改用 auth 工具。
- Modify: `frontend/pages/**/*.vue`
  - 常见用户 ID 读取和管理员上传请求头改用 auth 工具。
- Modify: `frontend/package.json`
  - 增加 `check:auth`。
- Modify: `scripts/check.sh`
  - 接入 `npm --prefix frontend run check:auth`。

---

## Tasks

- [x] 增加失败检查：`check:auth` 必须要求存在 `utils/auth.js`，并禁止页面直接读写四个登录态 key。
- [x] 运行 `npm --prefix frontend run check:auth`，确认当前直连 storage 会失败。
- [x] 新增 `frontend/utils/auth.js`，集中封装客户端/管理员登录态。
- [x] 修改 `frontend/utils/request.js`，请求层登录态来源切换到 auth 工具。
- [x] 修改登录、管理员登录、管理员首页和个人中心页面。
- [x] 修改业务页中重复的用户 ID 读取和管理员上传 Authorization 读取。
- [x] 修改 `frontend/package.json` 和 `scripts/check.sh`，纳入登录态检查。
- [x] 运行 `npm --prefix frontend run check:auth`。
- [x] 运行 `bash scripts/check.sh`。
- [x] 运行 `CHECK_FRONTEND_BUILD=1 bash scripts/check.sh`。
