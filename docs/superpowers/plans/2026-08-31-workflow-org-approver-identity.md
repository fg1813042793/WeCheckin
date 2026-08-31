# Workflow Org Approver Identity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add workflow-specific organization approver identities and let normal users link to a direct manager for approval routing.

**Architecture:** Keep permission roles separate from workflow approver identities. Store direct managers on users, store department-scoped workflow approver assignments in dedicated workflow tables, and resolve approval nodes to one or more user IDs according to the node approval mode.

**Tech Stack:** Go 1.24, GORM, Hertz, MySQL migrations, Vue 3, Vite, Element Plus.

---

### Task 1: RED Checks

**Files:**
- Modify: `admin/scripts/check-workflow-form-designer.mjs`
- Create: `backend/internal/modules/workflow/infrastructure/org_assignee_resolver_structure_test.go`
- Create: `backend/internal/service/admin/adminuser/manager_structure_test.go`

- [ ] Add a frontend structure check requiring organization identity UI, user-tree approval selection, and manager picker wiring.
- [ ] Add backend structure tests requiring `manager_user_id`, manager DTO fields, and organization approver assignment models.
- [ ] Run the targeted checks and confirm they fail before implementation.

### Task 2: User Direct Manager

**Files:**
- Modify: `backend/internal/model/account/user.go`
- Modify: `backend/internal/service/admin/adminuser/service.go`
- Modify: `backend/internal/handler/admin/user/handler.go`
- Create: `backend/migrations/20260831113000_add_user_manager_and_workflow_org_approvers.sql`
- Modify: `admin/src/api/types.ts`
- Modify: `admin/src/views/user/index.vue`

- [ ] Add `manager_user_id` to `users` through model and migration.
- [ ] Expose `managerUserId` and `managerUserName` in user list/detail DTOs.
- [ ] Accept `managerUserId` in user add/edit handlers and service methods, rejecting self-manager assignment.
- [ ] Add a department tree user picker in the admin user form for direct manager.

### Task 3: Organization Approver Identities

**Files:**
- Create: `backend/internal/model/workflow/org_approver.go`
- Create: `backend/internal/service/admin/workflow/org_approver.go`
- Modify: `backend/internal/modules/workflow/infrastructure/assignee_resolver.go`
- Modify: `backend/internal/workflow/types.go`
- Modify: `backend/internal/workflow/validation.go`
- Modify: `admin/src/views/workflow/types.ts`
- Modify: `admin/src/api/index.ts`

- [ ] Add workflow organization identity and assignment models.
- [ ] Add lightweight admin lookup APIs for active identities and department assignments.
- [ ] Add `org_identity` assignee type and resolve `starter_department:<identityCode>` plus `department:<deptId>:<identityCode>`.
- [ ] Preserve existing `department_leader` by resolving it as `starter_department:department_leader`.

### Task 4: Designer UI

**Files:**
- Modify: `admin/src/views/workflow/designer/index.vue`
- Modify: `admin/src/views/workflow/designer/components/NodeInspector.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowNodeCard.vue`
- Modify: `admin/scripts/check-workflow-form-designer.mjs`

- [ ] Load users, departments, and workflow organization identities into the workflow designer.
- [ ] Show a user tree selector for specified users and a department/identity selector for organization identities.
- [ ] Keep single-user mode limited to one selected user; keep sequential, parallel, and countersign as ordered/multiple user lists.
- [ ] Clarify that single approval with multiple resolved organization members means any one member may approve.

### Task 5: Verification

**Files:**
- Check only changed backend and admin workflow/user files.

- [ ] Run backend structure tests for manager and organization approvers.
- [ ] Run workflow validation/resolver tests.
- [ ] Run admin workflow/user structure checks.
- [ ] Run `git diff --check`.
