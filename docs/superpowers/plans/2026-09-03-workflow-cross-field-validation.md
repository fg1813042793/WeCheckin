# Workflow Cross-Field Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enlarge the workflow form property panel and expose type-safe validation between different fields in the same form.

**Architecture:** Keep the existing `compare_field` JSON contract. Add a small Admin-side compatibility helper shared by the validation-rule editor and Admin runtime validation, and mirror the same compatibility, operator, and typed comparison rules in `workflowcore` so saved definitions and submitted data cannot bypass the designer restrictions.

**Tech Stack:** Vue 3, TypeScript, Element Plus, Go, existing Node structure checks and Go tests.

---

### Task 1: Lock down the behavior with failing tests

**Files:**
- Modify: `admin/scripts/check-workflow-form-designer.mjs`
- Modify: `backend/internal/workflowcore/form_rules_test.go`

- [x] Assert the `440px`, `380px`, and `320px` property-panel widths.
- [x] Assert compatible and incompatible field type pairs.
- [x] Assert equality-only and ordered operator sets.
- [x] Run `node scripts/check-workflow-form-designer.mjs` and the focused Go test; confirm they fail because the compatibility helper and backend support are missing.

### Task 2: Implement the Admin designer behavior

**Files:**
- Create: `admin/src/views/workflow/workflowValidationRules.ts`
- Modify: `admin/src/views/workflow/designer/components/WorkflowValidationRulesEditor.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue`

- [x] Implement the shared field-family compatibility matrix.
- [x] Limit text, choice, identity, department, and boolean comparisons to `eq` and `ne`.
- [x] Keep all six operators for numbers, amounts, dates, datetimes, and times.
- [x] Reset an invalid operator when the target field changes.
- [x] Resize the property panel to the three tested responsive widths.
- [x] Run the Admin designer structure check and confirm it passes.

### Task 3: Enforce the same contract in the backend

**Files:**
- Modify: `backend/internal/workflowcore/validation.go`
- Modify: `backend/internal/workflowcore/form.go`
- Modify: `admin/src/views/workflow/runtimeForm.ts`

- [x] Extend comparable field families to choices, booleans, users, and departments.
- [x] Validate the operator against the owner and target field types.
- [x] Compare choice and identity values as exact strings instead of coercing numeric-looking values.
- [x] Run the focused Admin runtime and Go tests and confirm they pass.

### Task 4: Verify the affected applications

**Files:**
- No production changes expected.

- [x] Run `cd admin && npm run check:all`; the affected checks pass, then the existing workflow-version-history check blocks the suite.
- [x] Run `cd backend && GOCACHE=$PWD/../.cache/go-build go test ./... -count=1`; `internal/workflowcore` passes while unrelated workflow-version-history and sandbox listener failures block the suite.
- [x] Run the H5 workflow regression check, type check, lint check for the new script, and production build.
