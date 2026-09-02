# Workflow Core Package Rename Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rename the workflow definition core from `internal/workflow` and `package workflow` to `internal/workflowcore` and `package workflowcore` without changing behavior.

**Architecture:** Keep the existing boundary intact: `internal/workflowcore` owns definition types, form validation, availability validation, and BPMN compilation, while `internal/modules/workflow` continues to own runtime state and orchestration. This is a mechanical internal import migration with no API, persistence, or workflow-version compatibility layer.

**Tech Stack:** Go 1.24, Go internal packages, Go test, Markdown architecture documentation.

---

## File Map

Core files move together because they form one Go package:

- `backend/internal/workflowcore/types.go`: workflow definition schema and constants.
- `backend/internal/workflowcore/validation.go`: definition validation.
- `backend/internal/workflowcore/form.go`: form data and node patch validation.
- `backend/internal/workflowcore/start_availability.go`: start-window evaluation.
- `backend/internal/workflowcore/compiler.go`: BPMN compatibility export.
- `backend/internal/workflowcore/definition_test.go`: definition, form, availability, and compiler tests.
- `backend/internal/workflowcore/form_rules_test.go`: structured form rule tests.

Runtime and definition-management callers keep their existing responsibilities; only their import path changes. The current architecture documents change to reflect the new name. Historical plans and specifications keep the path that was correct when they were written.

### Task 1: Move and Rename the Core Package

**Files:**

- Move: `backend/internal/workflow/types.go` -> `backend/internal/workflowcore/types.go`
- Move: `backend/internal/workflow/validation.go` -> `backend/internal/workflowcore/validation.go`
- Move: `backend/internal/workflow/form.go` -> `backend/internal/workflowcore/form.go`
- Move: `backend/internal/workflow/start_availability.go` -> `backend/internal/workflowcore/start_availability.go`
- Move: `backend/internal/workflow/compiler.go` -> `backend/internal/workflowcore/compiler.go`
- Move: `backend/internal/workflow/definition_test.go` -> `backend/internal/workflowcore/definition_test.go`
- Move: `backend/internal/workflow/form_rules_test.go` -> `backend/internal/workflowcore/form_rules_test.go`

- [ ] **Step 1: Confirm the characterization-test baseline**

Run from `backend/`:

```bash
GOCACHE=$PWD/../.cache/go-build go test ./internal/workflow ./internal/modules/workflow/... ./internal/service/admin/workflow -count=1
```

Expected: all listed packages print `ok`. This baseline passed before the plan was written.

- [ ] **Step 2: Move each package file and change its package declaration**

Use `apply_patch` move operations so existing tracked and untracked file contents are preserved without staging user changes:

```diff
*** Begin Patch
*** Update File: backend/internal/workflow/types.go
*** Move to: backend/internal/workflowcore/types.go
@@
-package workflow
+package workflowcore
*** Update File: backend/internal/workflow/validation.go
*** Move to: backend/internal/workflowcore/validation.go
@@
-package workflow
+package workflowcore
*** Update File: backend/internal/workflow/form.go
*** Move to: backend/internal/workflowcore/form.go
@@
-package workflow
+package workflowcore
*** Update File: backend/internal/workflow/start_availability.go
*** Move to: backend/internal/workflowcore/start_availability.go
@@
-package workflow
+package workflowcore
*** Update File: backend/internal/workflow/compiler.go
*** Move to: backend/internal/workflowcore/compiler.go
@@
-package workflow
+package workflowcore
*** Update File: backend/internal/workflow/definition_test.go
*** Move to: backend/internal/workflowcore/definition_test.go
@@
-package workflow
+package workflowcore
*** Update File: backend/internal/workflow/form_rules_test.go
*** Move to: backend/internal/workflowcore/form_rules_test.go
@@
-package workflow
+package workflowcore
*** End Patch
```

- [ ] **Step 3: Verify that old imports now fail closed**

Run from `backend/`:

```bash
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/workflow/domain -count=1
```

Expected: FAIL with `package wecheckin/backend/internal/workflow is not in std` or an equivalent missing-package error. This proves callers still need explicit migration.

### Task 2: Migrate All Go Callers

**Files:**

- Modify: `backend/internal/modules/workflow/application/notification_test.go`
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/application/service_test.go`
- Modify: `backend/internal/modules/workflow/application/types.go`
- Modify: `backend/internal/modules/workflow/domain/engine.go`
- Modify: `backend/internal/modules/workflow/domain/engine_test.go`
- Modify: `backend/internal/modules/workflow/domain/runtime.go`
- Modify: `backend/internal/modules/workflow/infrastructure/assignee_resolver.go`
- Modify: `backend/internal/modules/workflow/infrastructure/assignee_resolver_test.go`
- Modify: `backend/internal/modules/workflow/infrastructure/form_clone_test.go`
- Modify: `backend/internal/modules/workflow/infrastructure/gorm_store.go`
- Modify: `backend/internal/modules/workflow/infrastructure/mapping_test.go`
- Modify: `backend/internal/modules/workflow/infrastructure/notification_channels.go`
- Modify: `backend/internal/modules/workflow/infrastructure/notification_dispatcher_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpclient/handler_test.go`
- Modify: `backend/internal/service/admin/workflow/service.go`
- Modify: `backend/internal/service/admin/workflow/service_test.go`

- [ ] **Step 1: Replace the old aliased import in every direct caller**

Apply the same exact import replacement to every file listed above:

```diff
-workflowcore "wecheckin/backend/internal/workflow"
+"wecheckin/backend/internal/workflowcore"
```

For single-line imports, the complete result is:

```go
import "wecheckin/backend/internal/workflowcore"
```

For import blocks, the complete entry is:

```go
"wecheckin/backend/internal/workflowcore"
```

No selector changes are required because existing call sites already use the `workflowcore` identifier.

- [ ] **Step 2: Format all changed Go files**

Run from the repository root:

```bash
gofmt -w backend/internal/workflowcore/*.go \
  backend/internal/modules/workflow/application/notification_test.go \
  backend/internal/modules/workflow/application/service.go \
  backend/internal/modules/workflow/application/service_test.go \
  backend/internal/modules/workflow/application/types.go \
  backend/internal/modules/workflow/domain/engine.go \
  backend/internal/modules/workflow/domain/engine_test.go \
  backend/internal/modules/workflow/domain/runtime.go \
  backend/internal/modules/workflow/infrastructure/assignee_resolver.go \
  backend/internal/modules/workflow/infrastructure/assignee_resolver_test.go \
  backend/internal/modules/workflow/infrastructure/form_clone_test.go \
  backend/internal/modules/workflow/infrastructure/gorm_store.go \
  backend/internal/modules/workflow/infrastructure/mapping_test.go \
  backend/internal/modules/workflow/infrastructure/notification_channels.go \
  backend/internal/modules/workflow/infrastructure/notification_dispatcher_test.go \
  backend/internal/modules/workflow/transport/httpadmin/handler_test.go \
  backend/internal/modules/workflow/transport/httpclient/handler_test.go \
  backend/internal/service/admin/workflow/service.go \
  backend/internal/service/admin/workflow/service_test.go
```

Expected: command exits with status 0 and produces no output.

- [ ] **Step 3: Run focused package tests**

Run from `backend/`:

```bash
GOCACHE=$PWD/../.cache/go-build go test ./internal/workflowcore ./internal/modules/workflow/... ./internal/service/admin/workflow -count=1
```

Expected: all listed packages print `ok`.

### Task 3: Update Current Architecture Documentation

**Files:**

- Modify: `backend/docs/development-guidelines.md:50`
- Modify: `docs/architecture/go-workflow-engine-v1.md:25`

- [ ] **Step 1: Update the Backend package-boundary rule**

Change the package name while preserving the separation rule:

```diff
-- `internal/workflow` 承载流程定义核心、表单校验和 BPMN 编译；`internal/modules/workflow` 承载运行时状态机与应用服务，不应将两者合并为同一个包。
+- `internal/workflowcore` 承载流程定义核心、表单校验和 BPMN 编译；`internal/modules/workflow` 承载运行时状态机与应用服务，不应将两者合并为同一个包。
```

- [ ] **Step 2: Update the workflow engine architecture map**

```diff
-- `internal/workflow`：流程定义、校验和 BPMN 导出。
+- `internal/workflowcore`：流程定义、校验和 BPMN 导出。
```

- [ ] **Step 3: Confirm historical records were not rewritten**

Run from the repository root:

```bash
git diff -- docs/superpowers/plans docs/superpowers/specs
```

Expected: no implementation-time changes under historical plans or specifications. The already committed design and this implementation plan are excluded from the working diff once committed separately.

### Task 4: Verify the Complete Rename

**Files:**

- Verify: `backend/internal/workflowcore/`
- Verify: `backend/internal/modules/workflow/`
- Verify: `backend/internal/service/admin/workflow/`
- Verify: `backend/docs/development-guidelines.md`
- Verify: `docs/architecture/go-workflow-engine-v1.md`

- [ ] **Step 1: Prove the old source directory is gone**

Run from the repository root:

```bash
test ! -d backend/internal/workflow
```

Expected: command exits with status 0 and produces no output.

- [ ] **Step 2: Prove current Go source has no old import or package declaration**

Run from the repository root:

```bash
if rg -n --glob '*.go' 'wecheckin/backend/internal/workflow"|^package workflow$' backend; then exit 1; fi
```

Expected: command exits with status 0 and produces no matches. The quote after `workflow` prevents the new `workflowcore` path from matching.

- [ ] **Step 3: Prove current architecture docs use the new path**

Run from the repository root:

```bash
rg -n 'internal/workflowcore' backend/docs/development-guidelines.md docs/architecture/go-workflow-engine-v1.md
```

Expected: one match in each file.

- [ ] **Step 4: Run the complete Backend test suite**

Run from `backend/`:

```bash
GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
```

Expected: all Backend packages pass.

- [ ] **Step 5: Check patch formatting and ownership boundaries**

Run from the repository root:

```bash
git diff --check
git status --short
```

Expected: `git diff --check` exits with status 0. `git status --short` shows the workflow core rename, direct import updates, and two current documentation updates alongside pre-existing user changes; no unrelated files are newly modified by this implementation.

Implementation files must remain uncommitted because several moved and importing files already contain user-owned uncommitted changes. A separate implementation commit would absorb those changes; create one only after the user explicitly reviews and authorizes that staging scope.
