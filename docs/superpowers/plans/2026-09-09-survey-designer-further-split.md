# Survey Designer Further Split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce `SurveyDesigner.vue` from 958 lines to roughly 550-650 lines without changing survey designer behavior.

**Architecture:** Keep `SurveyDesigner.vue` as the workspace composition root. Move import/export and question-bank orchestration into one typed composable, and move document persistence, auto-save, history, validation status, and route-leave protection into another typed composable. Preserve existing child component contracts and API payloads.

**Tech Stack:** Vue 3 Composition API, TypeScript strict, Element Plus, Vue Router, Node contract checks.

---

### Task 1: Lock the new ownership boundaries

**Files:**
- Modify: `admin/scripts/check-survey-designer-ux.mjs`
- Modify: `admin/scripts/check-component-complexity.mjs`

- [ ] Add assertions requiring `useSurveyDesignerTransfer.ts` and `useSurveyDesignerDocument.ts`.
- [ ] Assert that transfer, persistence, auto-save, and load implementations no longer live in `SurveyDesigner.vue`.
- [ ] Run `cd admin && npm run check:survey-designer-ux` and confirm it fails because the new composables do not exist.

### Task 2: Extract transfer and question-bank orchestration

**Files:**
- Create: `admin/src/views/survey/composables/useSurveyDesignerTransfer.ts`
- Modify: `admin/src/views/survey/SurveyDesigner.vue`

- [ ] Move text import/export, SurveyKing JSON conversion, JSON download, QR/link helpers, question-bank insertion, and upload-to-bank orchestration into the composable.
- [ ] Give the composable typed refs and callbacks for form, questions, selected question, ID generation, and dialog handles.
- [ ] Keep user messages and generated payloads behaviorally identical.
- [ ] Run `cd admin && npm run check:survey-designer-ux` and `npm run build`.

### Task 3: Extract document lifecycle

**Files:**
- Create: `admin/src/views/survey/composables/useSurveyDesignerDocument.ts`
- Modify: `admin/src/views/survey/SurveyDesigner.vue`

- [ ] Move settings/payload serialization, completion-action normalization, load/save, auto-save scheduling, dirty state, history restoration, keyboard shortcuts, before-unload, and route-leave confirmation into the composable.
- [ ] Pass issue validation and issue-location callbacks into the composable rather than importing presentation concerns into it.
- [ ] Return only the state and commands required by the parent template.
- [ ] Preserve the existing 800 ms auto-save delay and pending-save retry behavior.
- [ ] Run `cd admin && npm run check:survey-designer-ux` and `npm run build`.

### Task 4: Tighten budgets and verify

**Files:**
- Modify: `admin/scripts/check-component-complexity.mjs`

- [ ] Set the `SurveyDesigner.vue` budget to the achieved line count rounded up to a small maintenance margin.
- [ ] Add budgets for both new composables so complexity cannot simply migrate.
- [ ] Run `cd admin && npm run check:all`.
- [ ] Run `git diff --check -- admin/src/views/survey admin/scripts/check-survey-designer-ux.mjs admin/scripts/check-component-complexity.mjs`.
- [ ] Record final line counts and any remaining manual UI verification gap.
