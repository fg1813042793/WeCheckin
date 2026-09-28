# Workflow Calculation Values Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow workflow calculations to use select/radio option calculation values, map numeric results to labels, and configure the same calculation value in dictionary management while preserving old definitions.

**Architecture:** Keep `label` and `value` semantics unchanged. Add optional `calculationValue` to form options and dictionary items, using numeric legacy `value` as fallback. Add optional result label rules under calculation configuration while keeping the existing `display` field meaning. Share calculation behavior across backend, Admin runtime/designer, and H5 runtime.

**Tech Stack:** Go workflowcore, MySQL versioned SQL migrations, Vue 3 + TypeScript + Element Plus Admin, uni-app Vue H5App.

---

### Task 1: Add failing tests and contracts

**Files:**
- Modify: `backend/internal/workflowcore/calculation_test.go`
- Modify: `admin/scripts/check-workflow-form-designer.mjs`
- Modify: `admin/scripts/check-dict-management.mjs`

- [x] Add backend tests for numeric legacy select values, nonnumeric values with `calculationValue`, missing calculation values, and range/exact result labels.
- [x] Add frontend contracts for option calculation values, result mapping controls, and dictionary field/payload support.
- [x] Run focused tests/checks and confirm the new behavior fails before implementation.

### Task 2: Extend shared workflow contracts and calculation engines

**Files:**
- Modify: `backend/internal/workflowcore/types.go`
- Modify: `backend/internal/workflowcore/calculation.go`
- Modify: `admin/src/types/workflow.ts`
- Modify: `admin/src/views/workflow/workflowCalculation.ts`
- Modify: `h5app/src/types/workflow.ts`
- Modify: `h5app/src/pages/workflow/workflow-calculation.ts`

- [x] Add optional option `calculationValue` and calculation result display rules.
- [x] Permit single `select`/`radio` references when options are numeric through `calculationValue` or legacy `value`.
- [x] Keep multi-select/checkbox arithmetic unsupported with an explicit error.
- [x] Resolve numeric results and mapped labels without changing stored form data.
- [x] Preserve old `display` and numeric/amount calculation behavior.

### Task 3: Update workflow designer and runtime UI

**Files:**
- Modify: `admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowCalculationEditor.vue`
- Modify: `admin/src/views/workflow/runtimeForm.ts`
- Modify: `admin/src/views/workflow/components/WorkflowRuntimeForm.vue`
- Modify: `h5app/src/pages/workflow/workflow-form.ts`
- Modify: `h5app/src/pages/workflow/components/WorkflowRuntimeForm.vue`

- [x] Add calculation value input to static select/radio options and preserve it through JSON/remote normalization.
- [x] Add optional result label mapping configuration to the calculation editor.
- [x] Render mapped labels in Admin and H5 while old calculations still render formatted numbers.

### Task 4: Extend dictionary storage and management

**Files:**
- Create: `backend/migrations/20260912100000_add_dictionary_calculation_value.sql`
- Modify: `backend/internal/model/system/dict.go`
- Modify: `backend/internal/handler/admin/dict/handler.go`
- Modify: `backend/internal/service/admin/dict/service.go`
- Modify: `admin/src/api/types.ts`
- Modify: `admin/src/views/dict/index.vue`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`

- [x] Add a nullable decimal column and pass optional calculation values through admin/public dictionary APIs.
- [x] Add dictionary management input and table display support.
- [x] Preserve existing method compatibility through wrapper functions and unchanged old rows.

### Task 5: Verify all affected contracts

- [x] Run backend workflowcore and dictionary tests.
- [x] Run Admin workflow/dictionary checks and TypeScript build.
- [x] Run H5App lint and type-check.
- [x] Run migration/Swagger checks and `git diff --check`.
- [x] Report that browser/device visual verification was not performed unless explicitly requested.
