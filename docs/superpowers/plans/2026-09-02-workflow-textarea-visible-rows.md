# Workflow Textarea Visible Rows Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add designer-configurable textarea visible row bounds and consume them consistently in Admin and H5 runtime forms.

**Architecture:** Extend the workflow definition field contract with presentation-only row bounds. Validate the contract in workflowcore, configure it in the Admin designer, and normalize omitted values at each renderer so historical definitions remain unchanged.

**Tech Stack:** Go, Vue 3, TypeScript, Element Plus, uni-app, uView Pro, Node structure checks.

---

### Task 1: Workflow Definition Contract

**Files:**
- Modify: `backend/internal/workflowcore/types.go`
- Modify: `backend/internal/workflowcore/validation.go`
- Test: `backend/internal/workflowcore/definition_test.go`

- [ ] Add failing schema tests for textarea bounds and detail textarea columns.
- [ ] Run `go test ./internal/workflowcore -run 'TestValidateDefinition' -count=1` and confirm the new cases fail.
- [ ] Add `MinVisibleRows` and `MaxVisibleRows` JSON fields and validate positive ordered bounds up to 30 on textarea fields only.
- [ ] Rerun the focused workflowcore test and confirm it passes.

### Task 2: Admin Designer and Runtime

**Files:**
- Modify: `admin/src/views/workflow/types.ts`
- Modify: `admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowFormFieldPreview.vue`
- Modify: `admin/src/views/workflow/components/WorkflowRuntimeForm.vue`
- Test: `admin/scripts/check-workflow-form-designer.mjs`
- Test: `admin/scripts/check-workflow-runtime-form.mjs`

- [ ] Add failing source assertions for both property editors, defaults, cleanup on type changes, preview rows, and runtime autosize.
- [ ] Run the two Admin workflow checks and confirm the assertions fail.
- [ ] Add typed row properties, designer controls and defaults (`3-8`, detail `2-6`).
- [ ] Bind preview and runtime Element Plus textareas to normalized row bounds.
- [ ] Rerun checks, focused ESLint, and `pnpm type-check` in `admin`.

### Task 3: H5 Runtime

**Files:**
- Modify: `h5app/src/types/workflow.ts`
- Modify: `h5app/src/pages/workflow/components/WorkflowFieldControl.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowRuntimeForm.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowResizableTextarea.vue`
- Test: `h5app/scripts/check-workflow-module.mjs`

- [ ] Replace resize-oriented assertions with failing autosize, row-bound, compact-detail, and native-count assertions.
- [ ] Run `pnpm check:workflow-module` and confirm failure.
- [ ] Restore `u-textarea`, add min/max row props, cap auto-height with CSS, and remove the custom count and resize handle.
- [ ] Pass compact defaults for detail columns and configured bounds for all textarea fields.
- [ ] Rerun the workflow check, focused ESLint, `pnpm type-check`, and `pnpm build:h5`.

### Task 4: Final Verification

- [ ] Run `gofmt` on changed Go files.
- [ ] Run focused backend tests, Admin checks, and H5 checks.
- [ ] Run `git diff --check` in the main repository and H5 submodule.
- [ ] Inspect the compiled H5 CSS for auto-height bounds and absence of `resize: vertical`.
