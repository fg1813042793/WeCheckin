# H5App Workflow Summary Cross-Version Export Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow any authorized workflow instances, including different definitions and published versions, to be exported together while keeping one instance per Excel sheet.

**Architecture:** Remove the single-definition parameter from the H5 export URL, handler request, and summary service. Authorization remains based on selected instance IDs plus the current user's visibility scope; each detail continues loading its own bound definition version before rendering.

**Tech Stack:** Go 1.24, Hertz, workflow application service, Vue 3, uni-app, TypeScript, existing XLSX/PDF/DOCX renderers.

**Execution note:** Preserve all unrelated dirty work. Do not commit, reset, clean, or push.

---

### Task 1: Prove backend cross-definition authorization

**Files:**
- Modify: `backend/internal/service/dingtalkh5/workflowsummary/service_test.go`
- Modify: `backend/internal/service/dingtalkh5/workflowsummary/service.go`

- [x] Add a failing service test with two authorized instances using different definition IDs and versions.
- [x] Run `GOCACHE=$PWD/../.cache/go-build go test ./internal/service/dingtalkh5/workflowsummary -run TestExportSupportsCrossDefinitionVersions -count=1` and confirm RED.
- [x] Remove `DefinitionID` from `ExportRequest`, skip definition validation, and authorize export with definition ID zero.
- [x] Update existing partial-authorization test and run the workflowsummary package GREEN.

### Task 2: Remove the handler and API definition parameter

**Files:**
- Modify: `backend/internal/modules/workflow/transport/httpclient/summary_handler_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpclient/summary_handler.go`
- Modify: `backend/internal/routes/v2/swagger/h5app.go`
- Modify: `h5app/src/api/workflow.ts`

- [x] Add a failing handler test asserting export parses instance IDs and format without a definition ID.
- [x] Remove the handler query parsing and update the typed H5 URL builder signature.
- [x] Remove the Swagger export `definitionId` parameter and regenerate Swagger artifacts.
- [x] Run focused handler and Swagger tests.

### Task 3: Remove the H5 selection restriction

**Files:**
- Modify: `h5app/scripts/check-workflow-module.mjs`
- Modify: `h5app/src/pages/workflow/components/WorkflowSummarySection.vue`

- [x] Change the static contract to require direct selected-instance export and forbid the old same-definition toast.
- [x] Run `pnpm check:workflow-module` and confirm the new contract fails before implementation.
- [x] Delete definition ID grouping and call the new URL builder with selected IDs and format.
- [x] Re-run the workflow module check and targeted ESLint.

### Task 4: Verify rendering and regression boundaries

**Files:**
- Modify only if required by tests: `backend/internal/service/dingtalkh5/workflowsummary/export_test.go`

- [x] Run workflowsummary and HTTP client tests.
- [x] Run H5 `pnpm lint`, `pnpm type-check`, `pnpm check:workflow-module`, and `pnpm build:h5`.
- [x] Run backend `go test ./...` and report unrelated baseline failures separately.
- [x] Review the final diff to confirm instance-bound version loading and one-instance-per-sheet rendering remain unchanged.

