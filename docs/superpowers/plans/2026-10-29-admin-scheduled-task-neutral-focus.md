# Admin Scheduled Task Single Focus Border Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the double blue focus ring inside the scheduled-task editor dialog with one thin blue border without changing global Admin form behavior.

**Architecture:** Add dialog-scoped global selectors in `TaskEditorDialog.vue` because the Element Plus dialog is teleported to body. Keep one primary-color one-pixel inset border and remove the additional outer shadow and inner native outlines.

**Tech Stack:** Vue 3, Element Plus, scoped CSS, Admin static contract checks.

---

### Task 1: Add the focus-style contract

- [x] Extend `admin/scripts/check-scheduled-task.mjs` with required dialog focus selectors and neutral border style.
- [x] Run `npm run check:scheduled-task` and confirm RED.

### Task 2: Implement the scoped override

- [x] Add dialog-scoped focus rules to `admin/src/views/scheduled-task/components/TaskEditorDialog.vue`.
- [x] Run `npm run check:scheduled-task` and `npm run build`.

### Task 3: Verify rendered behavior

- [x] Open the scheduled-task creation dialog in the browser.
- [x] Verify input, textarea, and select controls use one thin blue focus border without an outer ring and retain stable dimensions.
- [x] Review the diff and preserve unrelated worktree changes.

