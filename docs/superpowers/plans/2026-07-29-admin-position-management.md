# Admin Position Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:test-driven-development before production changes. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在管理后台增加岗位管理页面，支持岗位列表、搜索、新增、编辑、启停和删除，并接入统一权限体系。

**Architecture:** 后端复用现有 `positions` 模型，新增 admin position service/handler，并注册 `/api/v2/admin/positions` RESTful 路由。权限统一写入 `permissions` 声明，前端新增岗位管理路由、菜单权限、API 方法和页面。

**Tech Stack:** Go、GORM、Hertz、Vue 3、Element Plus、Vite。

---

### Task 1: 后端岗位接口与权限

**Files:**
- Create: `backend/internal/app/service/position/service.go`
- Create: `backend/internal/app/handler/admin/position/handler.go`
- Modify: `backend/cmd/routes_v2.go`
- Modify: `backend/internal/app/support/adminmenuperm/declarations.go`
- Modify: `backend/internal/app/support/adminrouteperm/declarations.go`
- Modify: `backend/cmd/routes_v2_swagger.go`

- [ ] 写失败测试，要求 v2 路由注册 `/api/v2/admin/positions`。
- [ ] 写失败测试，要求菜单权限包含 `position:list/add/edit/del`。
- [ ] 实现岗位 service/handler。
- [ ] 注册路由和权限声明。

### Task 2: 管理后台页面

**Files:**
- Create: `admin/src/views/position/index.vue`
- Modify: `admin/src/router/adminRoutes.ts`
- Modify: `admin/src/api/index.ts`
- Modify: `admin/scripts/check-navigation.mjs`

- [ ] 写失败检查，要求岗位路由和 API 方法存在。
- [ ] 新增岗位管理页面，使用现有 `admin-page/admin-card/admin-toolbar/admin-pagination` 样式。
- [ ] 页面接入 `hasPerm('position:*')` 控制按钮。

### Task 3: 文档与验证

**Files:**
- Modify: `docs/API_V2.md`
- Modify: `README.md`
- Modify: `backend/docs/swagger/*`

- [ ] 更新 API 文档口径。
- [ ] 重新生成 Swagger。
- [ ] 运行后端相关测试、前端检查和 admin 构建。
