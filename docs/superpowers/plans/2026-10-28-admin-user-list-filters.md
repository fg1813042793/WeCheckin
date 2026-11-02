# Admin User List Filters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Add real server-side status, mobile, position, direct-manager, and role filters to the Admin user list.

**Architecture:** A typed `UserListFilter` is applied before count and pagination. Relationship filters use `EXISTS` subqueries, while a dedicated Vue component owns the expanded search UI so the existing oversized user page does not absorb another presentation responsibility.

**Tech Stack:** Go 1.24, GORM/MySQL, Hertz, Vue 3, TypeScript, Element Plus, Vite.

---

### Task 1: Lock the backend filter contract

**Files:**
- Create: `backend/internal/service/admin/adminuser/user_list_filter_test.go`
- Modify: `backend/internal/service/admin/adminuser/service.go`

- [x] Add DryRun tests that construct `UserListFilter` values and assert generated SQL contains `user_status`, `user_mobile`, `user_position_id`, active `user_reporting_relations`, and `user_roles` predicates.
- [x] Run `GOCACHE=$PWD/../.cache/go-build go test ./internal/service/admin/adminuser -run 'TestApplyUserListFilters' -count=1` and confirm the tests fail because the filter type/helper does not exist.
- [x] Add `UserListFilter`, `HasAny`, and `applyUserListFilters`; add `GetUserListByFilterContext` while preserving the old keyword wrapper.
- [x] Re-run the focused package tests and confirm they pass.

### Task 2: Extend the HTTP query contract

**Files:**
- Create: `backend/internal/handler/admin/user/list_filter_structure_test.go`
- Modify: `backend/internal/handler/admin/user/handler.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`
- Regenerate: `backend/docs/swagger/docs.go`
- Regenerate: `backend/docs/swagger/swagger.json`
- Regenerate: `backend/docs/swagger/swagger.yaml`

- [x] Add a handler contract test requiring `mobile`, `status`, `positionId`, `managerUserId`, and `roleId` parsing and a `GetUserListByFilterContext` call.
- [x] Run the focused handler test and confirm it fails on the missing parameters.
- [x] Parse optional IDs and status with explicit validation, build `UserListFilter`, and call the filtered service.
- [x] Add the five Swagger query annotations to both the handler and centralized v2 Swagger declaration.
- [x] Regenerate Swagger with the repository command and run focused handler/Swagger tests.

### Task 3: Add the typed filter component

**Files:**
- Create: `admin/src/views/user/components/UserListFilters.vue`
- Modify: `admin/src/api/types.ts`
- Modify: `admin/src/api/index.ts`
- Modify: `admin/scripts/check-user-list-ui.mjs`

- [x] Extend the UI contract check with required component import/rendering, five fields, search/reset events, and typed API query snippets.
- [x] Run `npm run check:user-list-ui` and confirm it fails because the component and query type do not exist.
- [x] Add `AdminUserListQuery extends PageQuery` and use it in `adminApi.userList`.
- [x] Implement `UserListFilters.vue` with two text inputs, four filterable selects, and search/reset controls in a responsive grid.
- [x] Re-run the UI contract check and keep it failing until page integration is added.

### Task 4: Integrate filters without growing page responsibilities

**Files:**
- Modify: `admin/src/views/user/index.vue`

- [x] Replace the inline keyword row with `UserListFilters` and add one reactive filter object.
- [x] Pass existing position, user, and role option arrays into the component.
- [x] Send trimmed, non-empty filters to `adminApi.userList`; preserve sort, page and page size.
- [x] Implement search/reset handlers that reset `page` to 1.
- [x] Run `npm run check:user-list-ui`, `npm run check:component-complexity`, and `npm run build`.

### Task 5: End-to-end verification

**Files:**
- Verify only; do not add committed browser artifacts.

- [x] Run focused backend user service and handler tests.
- [x] Run `npm run check:user-list-ui`, `npm run build`, and `npm run check:bundle`.
- [x] In the browser, verify status, mobile, position, manager and role filters individually and in combination; verify reset and pagination retain correct results.
- [x] Run `git diff --check` and report unrelated repository-wide baseline failures separately.

No commits are created because this is a shared dirty worktree and unrelated user changes must remain untouched.
