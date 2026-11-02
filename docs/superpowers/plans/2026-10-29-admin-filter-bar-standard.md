# Admin Filter Bar Standard Implementation Plan

> **For agentic workers:** Execute inline in the current workspace. Preserve unrelated dirty changes and do not commit.

**Goal:** Make every Admin list filter follow the approved spacing, inline action and label-free presentation rules.

**Architecture:** Strengthen the existing `AdminFilterActions` and `AdminSearchBar` shared primitives, then migrate only real list-filter pages that still violate the contract. Extend the existing static check so future changes cannot reintroduce visible filter labels, right-pushed query/reset actions or inconsistent search-bar padding.

**Tech Stack:** Vue 3, TypeScript, Element Plus, Node-based SFC contract checks, Vite.

---

### Task 1: Add failing filter-bar contract coverage

**Files:**
- Modify: `admin/scripts/check-filter-actions.mjs`

- [x] Assert `AdminSearchBar` uses `12px 14px` content padding and keeps actions inside the field flow.
- [x] Assert known list-filter components do not render visible `el-form-item` labels.
- [x] Assert query/reset action groups are not placed in `admin-toolbar__right`.
- [x] Run `npm run check:filter-actions` and confirm the new assertions fail before implementation.

### Task 2: Standardize shared search primitives

**Files:**
- Modify: `admin/src/components/admin-ui/AdminSearchBar.vue`
- Modify: `admin/src/components/admin-ui/AdminFilterActions.vue`

- [x] Change search content padding to `12px 14px`.
- [x] Keep action buttons as an indivisible inline group immediately after the final filter.
- [x] Preserve Enter-to-search, loading, disabled and reset behavior.
- [x] Run `npm run check:filter-actions` and confirm shared-component assertions pass.

### Task 3: Remove remaining visible list-filter labels

**Files:**
- Modify: `admin/src/views/user-feedback/index.vue`
- Modify only additional list-filter components identified by the contract check.

- [x] Remove visible filter labels while retaining placeholders.
- [x] Add `aria-label` to each affected input, select, number input and date picker.
- [x] Keep all current models, events and query semantics unchanged.
- [x] Run `npm run check:filter-actions` and affected page checks.

### Task 4: Normalize legacy filter layout

**Files:**
- Modify only violating files under `admin/src/views/**`.
- Modify shared Admin styles only when a page-neutral rule is sufficient.

- [x] Keep filter controls and `AdminFilterActions` in the same left-aligned flex flow.
- [x] Preserve page actions such as add, export and refresh outside the filter action group.
- [x] Verify no direct hand-written query/reset pairs remain.

### Task 5: Verify rendered behavior

**Files:**
- No production file changes unless browser QA finds a regression.

- [x] Run `npm run check:filter-actions` and affected focused checks.
- [x] Run `npm run build`.
- [x] Run `npm run check:all`; it stops at the existing navigation assertion `props.depth > 0`, unrelated to filter changes.
- [x] Browser-check one shared `AdminSearchBar` page and representative legacy toolbar pages at desktop and narrow widths.
- [x] Confirm controls do not overlap, action buttons follow the last filter, visible labels are absent and reset clears entered values.

## Execution Result

- Focused filter, user, audit, feedback, survey/exam, scheduled-task, notification and workflow checks passed.
- Admin production build passed.
- Browser verification covered `/user-feedbacks`, `/log` and `/user` at 1440px and 900px widths. Layout and reset interaction passed; local API requests reported connection refused because the backend service was not running for this visual check.
- No commit, push, reset or cleanup was performed. Existing unrelated worktree changes were preserved.

