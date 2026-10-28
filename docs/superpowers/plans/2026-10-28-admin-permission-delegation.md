# Admin Permission Delegation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prevent child administrators from viewing or granting permissions, roles, departments, or users beyond their own effective authority.

**Architecture:** Add a reusable delegation ceiling under `backend/internal/support/permissiondelegation`. Permission-tree reads filter full catalogs through this ceiling, while role and user writes validate submitted assignments transactionally before persistence. Existing Admin Vue pages keep their API shapes and receive already-filtered trees and role lists.

**Tech Stack:** Go 1.24, Hertz, GORM/MySQL, Vue 3, TypeScript, Element Plus.

---

### Task 1: Delegation Ceiling

**Files:**
- Create: `backend/internal/support/permissiondelegation/ceiling.go`
- Test: `backend/internal/support/permissiondelegation/ceiling_test.go`

- [x] Write failing tests for permission-key subset validation, data-scope ordering, visible department/user validation, role grant validation, and reserved-super-admin bypass.
- [x] Run `GOCACHE=$PWD/../.cache/go-build go test ./internal/support/permissiondelegation -count=1` and confirm missing implementation failures.
- [x] Implement `Ceiling`, `ResolveContext`, `ValidatePermissionKeys`, `ValidateRoleGrant`, `ValidateRoleIDs`, `ValidateDepartmentIDs`, and `ValidateUserIDs`.
- [x] Re-run the package tests and confirm they pass.

### Task 2: Filter Permission Trees

**Files:**
- Modify: `backend/internal/handler/admin/permission/handler.go`
- Modify: `backend/internal/service/admin/adminpermission/service.go`
- Modify: `backend/internal/handler/admin/role/handler.go`
- Modify: `backend/internal/service/admin/role/service.go`
- Test: `backend/internal/service/admin/adminpermission/delegation_test.go`
- Test: `backend/internal/service/admin/role/delegation_tree_test.go`

- [x] Write failing tests showing child administrators only receive allowed leaves plus required ancestors.
- [x] Add current-admin-aware tree/list functions and application-tree filtering.
- [x] Require reserved super administrator for permission catalog add/edit/delete.
- [x] Run the affected handler and service tests.

### Task 3: Enforce Role Delegation

**Files:**
- Modify: `backend/internal/handler/admin/role/handler.go`
- Modify: `backend/internal/service/admin/role/service.go`
- Test: `backend/internal/service/admin/role/delegation_write_test.go`

- [x] Write failing tests for over-privileged role creation/edit, hidden role ID edit/delete, reserved role management, and role-list filtering.
- [x] Validate requested role grants before creating or updating records.
- [x] Add Admin-scoped delete and batch-delete functions.
- [x] Filter role list totals and pagination after delegability checks.
- [x] Run `go test ./internal/service/admin/role ./internal/handler/admin/role -count=1`.

### Task 4: Enforce User Delegation

**Files:**
- Modify: `backend/internal/handler/admin/user/handler.go`
- Modify: `backend/internal/service/admin/adminuser/service.go`
- Test: `backend/internal/service/admin/adminuser/delegation_write_test.go`

- [x] Write failing tests for assigning hidden roles, direct permissions, outside departments/users, and reserved roles.
- [x] Pass current administrator ID through user creation.
- [x] Validate user departments, manager, role IDs, direct permission overrides, and extra data scopes inside the write transaction.
- [x] Run `go test ./internal/service/admin/adminuser ./internal/handler/admin/user -count=1`.

### Task 5: Frontend Contract and Regression Verification

**Files:**
- Modify only if required by response typing: `admin/src/api/types.ts`
- Test existing Admin contract scripts and backend route/permission guards.

- [x] Run focused Backend permission, role, user, middleware, and route tests.
- [x] Run `cd admin && npm run check:all`.
- [x] Run `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./... -count=1` and report unrelated baseline failures separately.
- [x] Run scoped `git diff --check` and verify unrelated dirty files and `h5app/` are untouched.
