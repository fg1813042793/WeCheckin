# Workflow Starter Display Name Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Return and display the workflow starter's user name instead of exposing the user ID in H5 workflow approval views.

**Architecture:** Extend the shared workflow instance summary DTO with `starterName`. Resolve names in the GORM store with one user query per instance page, reuse the same mapping for detail responses, and render a stable unknown-user fallback in H5 while retaining `starterId` for authorization only.

**Tech Stack:** Go 1.24, GORM, Hertz, Vue 3, TypeScript, uni-app, uView Pro

---

### Task 1: Lock the backend DTO and name resolution contract

**Files:**
- Modify: `backend/internal/modules/workflow/application/types.go`
- Modify: `backend/internal/modules/workflow/infrastructure/mapping_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpclient/handler_test.go`

- [ ] **Step 1: Add failing tests**

Add tests that require `InstanceSummary.StarterName`, prefer `user_name`, fall back to `user_account`, return an empty name for missing users, and serialize `"starterName":"张三"` in the H5 detail response.

- [ ] **Step 2: Verify the tests fail**

Run:

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/infrastructure ./internal/modules/workflow/transport/httpclient -count=1
```

Expected: compilation or assertion failure because `starterName` is not yet part of the instance contract.

### Task 2: Resolve starter names without N+1 queries

**Files:**
- Modify: `backend/internal/modules/workflow/infrastructure/gorm_store.go`

- [ ] **Step 1: Implement a batch name resolver**

Collect numeric user IDs, query `id`, `user_name`, and `user_account` once, and return a map keyed by the original string ID. Keep unknown IDs absent from the map.

- [ ] **Step 2: Populate list and detail summaries**

Apply the map to every list item and the single detail item. Do not change notification payload fallback behavior.

- [ ] **Step 3: Format and verify backend packages**

Run:

```bash
gofmt -w internal/modules/workflow/application/types.go internal/modules/workflow/infrastructure/gorm_store.go internal/modules/workflow/infrastructure/mapping_test.go internal/modules/workflow/transport/httpclient/handler_test.go
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/infrastructure ./internal/modules/workflow/transport/httpclient -count=1
```

Expected: both packages pass.

### Task 3: Display the starter name in H5

**Files:**
- Modify: `h5app/src/types/workflow.ts`
- Modify: `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue`
- Modify: `h5app/scripts/check-workflow-module.mjs`

- [ ] **Step 1: Add a failing structure check**

Require `starterName: string`, require the “发起人” value to use `detail.instance.starterName`, and forbid `#{{ detail.instance.starterId }}`.

- [ ] **Step 2: Verify the check fails**

Run `cd h5app && node scripts/check-workflow-module.mjs`.

Expected: failure because the H5 type and template still use only `starterId`.

- [ ] **Step 3: Implement the H5 contract**

Add `starterName` to `WorkflowInstanceSummary` and render `detail.instance.starterName || '未知用户'` in the detail summary. Leave authorization checks on `starterId` unchanged.

- [ ] **Step 4: Verify H5**

Run:

```bash
cd h5app
pnpm check:workflow-module
pnpm exec eslint src/types/workflow.ts src/pages/workflow/components/WorkflowDetailPanel.vue scripts/check-workflow-module.mjs
pnpm type-check
pnpm build:h5
```

Expected: all scoped checks and the H5 build pass.

### Task 4: Run final regression checks

**Files:**
- Review all files listed above.

- [ ] **Step 1: Run workflow backend regression tests**

Run `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/... -count=1`.

Expected: all workflow module packages pass.

- [ ] **Step 2: Confirm no ID is rendered as the starter name**

Run `rg -n "发起人|starterName|#\\{\\{ detail.instance.starterId" h5app/src/pages/workflow h5app/src/types/workflow.ts`.

Expected: the visible starter value uses `starterName`; `starterId` remains only in authorization logic.
