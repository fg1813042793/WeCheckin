# Unified Permissions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立 `permissions` + `permission_grants` 统一权限底座，并让后台登录、菜单、按钮和数据权限优先读取统一授权。

**Architecture:** 第一阶段新增统一权限模型和服务，迁移旧 `menus/role_menus/role_depts` 数据后清理旧角色授权关系表，业务读取统一授权表。角色表单继续使用菜单树交互，但保存只写统一授权；用户独立授权先提供后端能力和轻量表单字段。

**Tech Stack:** Go、GORM、MySQL、Hertz、Vue 3、Element Plus、uni-app。

---

### Task 1: 模型与迁移

**Files:**
- Create: `backend/internal/model/permission.go`
- Create: `backend/internal/model/unified_permission_test.go`
- Create: `backend/internal/bootstrap/unified_permission_migration_test.go`
- Modify: `backend/internal/bootstrap/migrate.go`
- Create: `backend/migrations/20260728170000_add_unified_permissions.sql`

- [ ] **Step 1: Write the failing model test**

```go
func TestUnifiedPermissionModelsUseExpectedTables(t *testing.T) {
	if (Permission{}).TableName() != "permissions" {
		t.Fatalf("permissions table name mismatch")
	}
	if (PermissionGrant{}).TableName() != "permission_grants" {
		t.Fatalf("permission grants table name mismatch")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/model -run TestUnifiedPermissionModelsUseExpectedTables -count=1`

Expected: FAIL with undefined `Permission` and `PermissionGrant`.

- [ ] **Step 3: Add models and AutoMigrate entries**

Create the two model structs with stable table names and add them to `autoMigrate()`.

- [ ] **Step 4: Add migration SQL**

Create tables, seed `admin:login` and `data:*`, migrate `menus` to `permissions`, migrate `role_menus` and data scope to `permission_grants`.

- [ ] **Step 5: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/model ./backend/internal/bootstrap -run 'UnifiedPermission|AutoMigrate' -count=1`

Expected: PASS.

### Task 2: 权限服务

**Files:**
- Create: `backend/internal/app/support/permission/service.go`
- Create: `backend/internal/app/support/permission/service_structure_test.go`
- Modify: `backend/internal/app/support/adminaccess/adminaccess.go`
- Modify: `backend/internal/app/service/menu/service.go`
- Modify: `backend/internal/app/support/access/access.go`

- [ ] **Step 1: Write the failing service structure test**

```go
func TestPermissionServiceExposesUnifiedAccessFunctions(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	for _, snippet := range []string{
		"AdminLoginPermissionKey",
		"SubjectHasPermissionContext",
		"AdminMenuIDsContext",
		"AdminPermCodesContext",
		"DataScopeContext",
		"SetRoleAdminPermissionsTx",
	} {
		if !strings.Contains(string(src), snippet) {
			t.Fatalf("permission service must expose %s", snippet)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/support/permission -run TestPermissionServiceExposesUnifiedAccessFunctions -count=1`

Expected: FAIL because package or functions are missing.

- [ ] **Step 3: Implement permission service**

Implement subject permission lookup with role allow, user allow, and user deny override. Add legacy fallback when unified tables are not ready.

- [ ] **Step 4: Wire admin access**

Use `admin:login` for login admission and role manager filtering, with reserved super admin fallback.

- [ ] **Step 5: Wire menu and data scope**

Use unified grants for admin menu IDs and permission codes. Data scope reads `data:*` grants first and old role fields second.

- [ ] **Step 6: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/support/permission ./backend/internal/app/support/adminaccess ./backend/internal/app/service/menu ./backend/internal/app/support/access -count=1`

Expected: PASS.

### Task 3: 角色与用户表单同步

**Files:**
- Modify: `backend/internal/app/service/role/service.go`
- Modify: `backend/internal/app/service/adminuser/service.go`
- Modify: `backend/internal/app/handler/admin/role/handler.go`
- Modify: `backend/internal/app/handler/admin/user/handler.go`
- Modify: `admin/src/views/role/index.vue`
- Modify: `admin/src/views/user/index.vue`
- Modify: `frontend/pages/admin/user/admin_user_edit.vue`

- [ ] **Step 1: Write failing structure tests**

Require role save to call `SetRoleAdminPermissionsTx` and user save to call `SetUserAdminPermissionOverridesTx`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/service/role ./backend/internal/app/service/adminuser -run 'UnifiedPermission|AdminAccess' -count=1`

Expected: FAIL because unified save calls are absent.

- [ ] **Step 3: Double-write role permissions**

Role save writes only unified grants: `admin:login`、`admin:menu:*`、`data:*`. Legacy `role_menus` and `role_depts` are migrated first and then dropped.

- [ ] **Step 4: Add user override input**

User save accepts optional permission override fields and writes user-level `allow/deny` grants without changing role binding.

- [ ] **Step 5: Adjust PC role UI labels**

Rename “后台登录” to “后台入口权限”; export text uses “有/无后台入口权限”.

- [ ] **Step 6: Run builds**

Run: `npm run build` in `admin`, and `npm run build:h5` in `frontend`.

Expected: both builds complete.

### Task 4: Final verification

**Files:**
- All modified backend and frontend files.

- [ ] **Step 1: Run backend focused tests**

Run: `GOCACHE=$PWD/.cache/go-build go test ./backend/internal/bootstrap ./backend/internal/model ./backend/internal/app/support/permission ./backend/internal/app/support/adminaccess ./backend/internal/app/service/menu ./backend/internal/app/support/access ./backend/internal/app/service/role ./backend/internal/app/service/adminauth ./backend/internal/app/service/adminuser -count=1`

Expected: PASS.

- [ ] **Step 2: Run admin build**

Run: `npm run build` from `admin`.

Expected: exit 0.

- [ ] **Step 3: Run frontend H5 build**

Run: `npm run build:h5` from `frontend`.

Expected: exit 0.
