# Question Bank Workspace Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Replace the question-bank table and large edit dialog with a persistent three-column workspace that supports category navigation, continuous question editing, properties, preview/JSON modes, and mixed autosave.

**Architecture:** Keep `QuestionBank.vue` as the API/state coordinator and extract three presentational components for navigation, editor canvas, and properties. Reuse existing schema normalization and rich-text helpers, preserve current backend contracts, and add a small save-state/dirty guard in the page coordinator.

**Tech Stack:** Vue 3 Composition API, TypeScript, Element Plus, Vue Quill, existing Admin request layer and question-bank APIs.

---

### Task 1: Lock the workspace contract

**Files:**
- Modify: `admin/scripts/check-question-bank-components.mjs`

- [x] Require `QuestionBankNavigator.vue`, `QuestionEditorCanvas.vue`, and `QuestionPropertiesPanel.vue`.
- [x] Require the page to render the three components, save-state labels, autosave scheduling, keyboard save, and desktop-only notice.
- [x] Forbid `QuestionBankTable` and `QuestionEditorDialog` as page-level editing entry points.
- [x] Run `npm run check:question-bank-components`; expect failure because workspace components do not exist.

### Task 2: Build the navigation and properties columns

**Files:**
- Create: `admin/src/views/question-bank/components/QuestionBankNavigator.vue`
- Create: `admin/src/views/question-bank/components/QuestionPropertiesPanel.vue`

- [x] Implement category entries for all, uncategorized, and current API categories.
- [x] Implement internal-scroll question list, keyword/type filters, loading/error/empty states, and add/select events.
- [x] Implement properties for type, category, tags, required, read-only, placeholder, description, upload/rating settings, and exam score/answer/analysis.
- [x] Keep props/emits typed and avoid API calls inside these components.

### Task 3: Build the center editor canvas

**Files:**
- Create: `admin/src/views/question-bank/components/QuestionEditorCanvas.vue`
- Reuse: `admin/src/views/question-bank/components/QuestionRichEditorDialog.vue`

- [x] Move visual title, option, multi-field, matrix, upload, rating, and live-preview UI out of the old dialog.
- [x] Add edit/preview/JSON segmented modes.
- [x] Emit content mutations through the shared reactive draft and expose format/sync actions supplied by the page.
- [x] Provide a no-selection state with a primary add-question action.

### Task 4: Convert the page into a workspace and add mixed saving

**Files:**
- Modify: `admin/src/views/question-bank/QuestionBank.vue`
- Remove from page usage: `admin/src/views/question-bank/components/QuestionBankTable.vue`
- Remove from page usage: `admin/src/views/question-bank/components/QuestionEditorDialog.vue`
- Remove from page usage: `admin/src/views/question-bank/components/QuestionPreviewDrawer.vue`

- [x] Replace the card/table/dialog template with a fixed-height workspace shell and top toolbar.
- [x] Preserve the existing API, schema, import/export, normalization, rich editor, delete, and scope-switch logic.
- [x] Add explicit new-question drafts and existing-question selection.
- [x] Add `dirty`, `saveState`, 800ms autosave timer, immediate save, and `Ctrl/Cmd + S`.
- [x] Flush pending saves before switching; on failure keep the draft selected and expose retry/discard/cancel behavior.
- [x] Apply add/edit/delete permissions to controls and read-only state.
- [x] Add desktop, narrow desktop, and mobile-notice CSS without changing the route.

### Task 5: Verify behavior and layout

**Files:**
- Modify as needed: `admin/scripts/check-question-bank-components.mjs`

- [x] Run `npm run check:question-bank-components` and confirm it passes.
- [x] Run `npm run check:component-complexity` and correct any new budget violation.
- [x] Run `npm run check:all`; separate unrelated baseline failures from this feature.
- [x] Start the Admin dev server and use browser screenshots at 1440px, 1024px, and 390px.
- [x] Exercise survey/exam switching, category selection, search, add, edit, autosave, immediate save, delete, preview, JSON, and disabled-permission states.
- [x] Run `git diff --check` on the changed question-bank files.

## Plan Review

- The plan covers the approved three-column layout, current-category data model, top-level scope switch, mixed saving, permissions, error retention, responsive behavior, static checks, build checks, and browser verification.
- No backend or database change is required.
- Existing unrelated dirty work remains untouched; no commit is created unless explicitly requested.
- Execution result: focused question-bank, shared FormKit type, Admin UI contract, build, and bundle checks passed. `check:all` reached the pre-existing navigation assertion failure in `AdminMenuNode.vue` before the question-bank checks.
