# Admin 接口文档菜单 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 Admin 后台增加受菜单权限控制的“接口文档”入口，并在后台页签内可靠地嵌入后端 Swagger UI。

**Architecture:** 后端权限目录和版本化 SQL 共同声明 `admin:menu:swagger`。Admin 新增本地路由和专用 iframe 页面，统一通过同源 `/swagger` 访问文档；开发 Vite 与两套生产 Nginx 配置负责将该路径代理到后端。

**Tech Stack:** Go 1.24、GORM/MySQL 版本化迁移、Vue 3、TypeScript strict、Element Plus、Vite、Nginx、Node.js 结构检查

---

### Task 1: 锁定菜单与迁移契约

**Files:**
- Create: `backend/internal/support/adminmenuperm/swagger_structure_test.go`
- Create: `backend/test/internal/bootstrap/migrations/admin_swagger_menu_migration_test.go`
- Modify: `backend/internal/support/adminmenuperm/declarations.go`
- Create: `backend/migrations/20260902143000_add_admin_swagger_menu.sql`

- [ ] **Step 1: 写失败测试**

断言 `Declarations(false)` 包含键 `admin:menu:swagger`、路径 `/swagger-docs`、权限 `swagger:view`、图标 `Document` 和菜单类型；断言迁移使用 `INSERT INTO permissions`、`ON DUPLICATE KEY UPDATE` 且不自动扩大普通角色授权。

- [ ] **Step 2: 运行测试并确认因声明和迁移缺失而失败**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/adminmenuperm ./test/internal/bootstrap/migrations -run 'TestAdminSwagger' -count=1
```

- [ ] **Step 3: 实现菜单声明与迁移**

在顶级菜单声明中增加：

```go
{Key: "admin:menu:swagger", Name: "接口文档", Path: "/swagger-docs", Perms: "swagger:view", Icon: "Document", Sort: 19, Type: TypeMenu}
```

迁移以同一字段插入 `permissions`，冲突时同步名称、路径、图标、权限标识、排序和状态，不写入 `permission_grants`。

- [ ] **Step 4: 重跑定向测试并确认通过**

### Task 2: 锁定 Admin 页面与代理契约

**Files:**
- Create: `admin/scripts/check-swagger-docs.mjs`
- Modify: `admin/package.json`
- Modify: `admin/src/router/adminRoutes.ts`
- Create: `admin/src/views/swagger-docs/index.vue`
- Modify: `admin/vite.config.ts`
- Modify: `admin/nginx/default.conf.template`
- Modify: `nginx/default.conf.template`

- [ ] **Step 1: 写失败结构检查**

检查脚本要求路由 `/swagger-docs` 懒加载页面；页面包含 `/swagger/doc.json` 可用性探测、`/swagger/index.html` iframe、刷新和新窗口动作；Vite 与两套 Nginx 均代理 `/swagger`。

- [ ] **Step 2: 运行检查并确认因页面和代理缺失而失败**

```bash
cd admin
node scripts/check-swagger-docs.mjs
```

- [ ] **Step 3: 实现路由、页面和代理**

页面使用 `fetch('/swagger/doc.json', { cache: 'no-store' })` 探测文档；成功后通过以 key 重建 iframe 的方式刷新，失败时使用 `el-result` 显示错误与重试。新窗口动作调用：

```ts
window.open('/swagger/index.html', '_blank', 'noopener,noreferrer')
```

Vite 和 Nginx 都按原 URI 将 `/swagger` 转发给现有后端上游。

- [ ] **Step 4: 将检查接入 `npm run check:all` 并确认通过**

### Task 3: 全量验证

- [ ] **Step 1: 格式化并运行后端全量测试**

```bash
cd backend
gofmt -w internal/support/adminmenuperm/declarations.go internal/support/adminmenuperm/swagger_structure_test.go test/internal/bootstrap/migrations/admin_swagger_menu_migration_test.go
GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
```

- [ ] **Step 2: 运行 Admin 全量门禁**

```bash
cd admin
npm run check:all
```

- [ ] **Step 3: 检查补丁格式和受影响文件**

```bash
git diff --check -- backend/internal/support/adminmenuperm backend/migrations/20260902143000_add_admin_swagger_menu.sql backend/test/internal/bootstrap/migrations/admin_swagger_menu_migration_test.go admin/src/router/adminRoutes.ts admin/src/views/swagger-docs admin/vite.config.ts admin/nginx/default.conf.template nginx/default.conf.template admin/package.json admin/scripts/check-swagger-docs.mjs
git status --short
```

- [ ] **Step 4: 记录未执行项**

不自动运行浏览器验证；不自动连接数据库执行迁移。交付时明确需要执行 `cd backend && bash init.sh` 并重启部署 Nginx/Admin。
