# Workflow Instance Compatible Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a permission-controlled Admin action that upgrades only structurally compatible running workflow instances to the latest published version.

**Architecture:** Put compatibility analysis and transactional upgrade orchestration in the workflow application layer. Infrastructure supplies locked state, latest definition loading, conditional version update and history persistence; HTTP Admin and Vue only adapt the contract.

**Tech Stack:** Go/Hertz/GORM/MySQL, Vue 3/TypeScript/Element Plus, versioned SQL permissions.

---

### Task 1: Compatibility analysis and transactional upgrade

**Files:**
- Create: `backend/internal/modules/workflow/application/instance_version_upgrade.go`
- Create: `backend/internal/modules/workflow/application/instance_version_upgrade_test.go`
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/domain/runtime.go`

- [x] Add failing tests for compatible, latest, terminal, node-missing, node-type, form-key and active-revision cases.
- [x] Add preview and upgrade DTOs plus a pure compatibility analyzer.
- [x] Add a transactional upgrade method that rechecks compatibility and appends an audit event.

### Task 2: Persistence and HTTP Admin contract

**Files:**
- Modify: `backend/internal/modules/workflow/infrastructure/runtime_store.go`
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler.go`
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler_test.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`

- [x] Add a conditional version update persistence method and focused store tests.
- [x] Add preview/upgrade handlers with scope checks and authenticated administrator identity.
- [x] Register GET preview and POST upgrade routes and Swagger declarations.

### Task 3: Permission synchronization

**Files:**
- Modify: `backend/internal/support/adminmenuperm/declarations.go`
- Modify: `backend/internal/support/adminmenuperm/workflow_structure_test.go`
- Modify: `backend/internal/support/adminrouteperm/catalog.go`
- Modify: `backend/internal/support/adminrouteperm/workflow_runtime_structure_test.go`
- Create: `backend/migrations/20261102100000_add_workflow_instance_upgrade_permission.sql`
- Create: `backend/migrations/workflow_instance_upgrade_permission_test.go`

- [x] Add failing permission catalog and migration checks.
- [x] Add menu/API permissions using business permission `workflow:instance:upgrade` without broad automatic grants.

### Task 4: Admin preview and confirmation UI

**Files:**
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/types/workflow.ts`
- Create: `admin/src/views/workflow/instances/InstanceVersionUpgradeDialog.vue`
- Modify: `admin/src/views/workflow/instances/index.vue`
- Modify: `admin/scripts/check-workflow-runtime-pages.mjs`

- [x] Add failing Admin contracts for API types, permission, button, preview dialog and refresh.
- [x] Add preview/upgrade API methods and typed response.
- [x] Add the permission-controlled list/detail action and compatibility dialog.

### Task 5: Verification

**Files:**
- Verify all modified files; do not edit generated output manually.

- [x] Run focused workflow application, infrastructure, HTTP, permission and migration tests.
- [x] Regenerate Swagger through the repository command and run route/catalog tests.
- [x] Run Admin workflow contracts and `npm run build`.
- [x] Run `git diff --check` and report browser/deployment validation gaps.
