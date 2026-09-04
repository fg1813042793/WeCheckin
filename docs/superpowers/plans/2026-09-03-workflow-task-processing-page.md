# Workflow Task Processing Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make pending workflow tasks open as a full processing page that matches the workflow start page while preserving node permissions and approval behavior.

**Architecture:** Keep task completion, form validation, and permission logic in `WorkflowDetailPanel.vue`, and specialize only its page presentation. Return the version-pinned graph with instance details so approvers can view the correct flow without requiring permission to start that definition.

**Tech Stack:** Go, GORM, Vue 3, TypeScript, uni-app, uView Pro, SCSS.

---

### Task 1: Lock the expected contracts

**Files:**
- Modify: `backend/internal/modules/workflow/infrastructure/mapping_test.go`
- Modify: `h5app/scripts/check-workflow-module.mjs`

- [x] Add a failing backend mapping test for instance graph nodes and edges.
- [x] Add failing H5 structure checks for the processing tabs, graph component, form card, and page actions.
- [x] Run both focused checks and confirm they fail for missing behavior.

### Task 2: Return the instance-version graph

**Files:**
- Modify: `backend/internal/modules/workflow/application/types.go`
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/infrastructure/gorm_store.go`
- Modify: `backend/internal/modules/workflow/application/service_test.go`

- [x] Add `nodes` and `edges` to `InstanceDetail`.
- [x] Build graph data from the definition version bound to the instance.
- [x] Resolve dynamic assignee display names against the instance starter.
- [x] Run focused workflow application and infrastructure tests.

### Task 3: Build the full task processing page

**Files:**
- Modify: `h5app/src/types/workflow.ts`
- Modify: `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowTaskPage.vue`

- [x] Add instance graph fields to the H5 type contract.
- [x] Add `审批处理`, `流程记录`, and `流程图` tabs for page presentation.
- [x] Render the editable form and comment in one centered card with the existing node permission maps.
- [x] Keep reject/approve/submit behavior and place cancel before the task actions.
- [x] Render the version-pinned read-only graph.

### Task 4: Verify

**Files:**
- Verify only.

- [x] Run backend focused tests.
- [x] Run workflow checks, targeted lint, type checking, and the H5 build. Full lint remains blocked by pre-existing unrelated errors.
- [x] Verify the clean development server loads without an overlay; authenticated task-page interaction remains pending because no login session was available.
