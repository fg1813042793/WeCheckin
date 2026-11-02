# H5App Workflow Feedback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep workflow start-limit information visible while reducing history-import feedback to a transient success toast or a compact expandable warning.

**Architecture:** Reuse the existing `WorkflowStartPage.vue` status areas. Add a pure predicate beside the existing import-summary helper so warning visibility is testable, then render only anomalous import results as an inline notice with expandable details.

**Tech Stack:** Vue 3, TypeScript, uni-app, uView Pro, project Node contract checks.

---

### Task 1: Define import-feedback severity

**Files:**
- Modify: `h5app/src/pages/workflow/workflow-history-import-preset.ts`
- Test: `h5app/scripts/check-workflow-history-import-presets.mjs`

- [x] Add failing assertions proving skipped, filtered, and dynamic-option counts require attention while clean imports do not.
- [x] Add the minimal pure predicate and rerun the focused preset check.

### Task 2: Render compact expandable feedback

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowStartPage.vue`
- Modify: `h5app/src/locale/lang/zh-CN.json`
- Modify: `h5app/src/locale/lang/en-US.json`
- Test: `h5app/scripts/check-workflow-history-import.mjs`

- [x] Add failing source-contract checks for warning-only persistence, compact summary text, and a details toggle.
- [x] Replace the long always-visible import result with warning-only structured state.
- [x] Keep clean success feedback transient through the existing toast.
- [x] Add an accessible details toggle and responsive warning styling.

### Task 3: Verify the H5App change

**Files:**
- Verify only; do not edit generated output.

- [x] Run the focused history-import checks.
- [x] Run type checking and focused lint; record the repository-wide pre-existing lint failures separately.
- [x] Run the workflow-module contract and H5 build; record the pre-existing `WorkflowCenter.vue` contract failure separately.
- [x] Run `git diff --check` on the changed files and report the authenticated import-state browser validation gap.
