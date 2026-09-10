# Exam Designer Split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce `ExamDesigner.vue` from 2847 lines to below 1000 lines while preserving exam scoring, question-bank, logic, settings, import/export, and persistence behavior.

**Architecture:** Keep `ExamDesigner.vue` as the workspace composition root. Extract stable visual regions into exam-specific components, move stateful workflows into typed composables, and move SFC styles into dedicated stylesheets. Reuse survey helpers only where their contracts are already generic; keep exam scoring behavior isolated.

**Tech Stack:** Vue 3 Composition API, TypeScript strict, Element Plus, Vue Router, Node contract checks.

---

### Task 1: Add exam designer contracts

**Files:**
- Create: `admin/scripts/check-exam-designer-ux.mjs`
- Modify: `admin/package.json`

- [ ] Assert the required exam components, composables, and stylesheets exist.
- [ ] Assert extracted implementations no longer live in `ExamDesigner.vue`.
- [ ] Add `check:exam-designer-ux` to `check:all`.
- [ ] Run the new check and confirm it fails before implementation.

### Task 2: Extract dialogs and transfer workflow

**Files:**
- Create: `admin/src/views/exam/components/ExamDesignerFormulaDialog.vue`
- Create: `admin/src/views/exam/components/ExamDesignerTextImportDialog.vue`
- Create: `admin/src/views/exam/components/ExamDesignerBankUploadDialog.vue`
- Create: `admin/src/views/exam/composables/useExamDesignerTransfer.ts`
- Modify: `admin/src/views/exam/ExamDesigner.vue`

- [ ] Move dialog markup and local state out of the parent.
- [ ] Preserve text parsing, exam answer/score import, JSON export/import, and question-bank upload payloads.
- [ ] Keep parent interaction through typed refs and emits.

### Task 3: Extract sidebar and logic panel

**Files:**
- Create: `admin/src/views/exam/components/ExamDesignerSidebar.vue`
- Create: `admin/src/views/exam/components/ExamDesignerLogicPanel.vue`
- Modify: `admin/src/views/exam/ExamDesigner.vue`

- [ ] Move type palette, question-bank tree, outline, and appearance resources into the sidebar.
- [ ] Move logic rule editor state, DSL rendering, and persistence trigger into the logic panel.
- [ ] Preserve question selection and add-from-bank callbacks.

### Task 4: Extract persistence and styles

**Files:**
- Create: `admin/src/views/exam/composables/useExamDesignerDocument.ts`
- Create: `admin/src/views/exam/exam-designer.css`
- Modify: `admin/src/views/exam/ExamDesigner.vue`

- [ ] Move form initialization, completion action, load/save, auto-save, and payload serialization into the document composable.
- [ ] Preserve the existing 500 ms auto-save delay and save messages.
- [ ] Move both style blocks into the dedicated stylesheet without selector changes.

### Task 5: Tighten budgets and verify

**Files:**
- Modify: `admin/scripts/check-component-complexity.mjs`

- [ ] Set the parent budget below 1000 lines and add budgets for new files.
- [ ] Run `npm run check:exam-designer-ux`.
- [ ] Run `npm run check:all`.
- [ ] Run `git diff --check` for the scoped files and record final line counts.
