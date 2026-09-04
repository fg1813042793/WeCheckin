# Workflow Start Form Card Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Present the workflow start form inside one responsive card and swap the submit/draft button positions.

**Architecture:** Keep `WorkflowRuntimeForm` and all workflow APIs unchanged. Add one page-level wrapper in `WorkflowStartPage.vue`, style it responsively, and protect the wrapper and action order through the existing workflow module structure check.

**Tech Stack:** Vue 3, uni-app, uView Pro, SCSS, Node.js structure checks

---

### Task 1: Add the form card and reorder actions

**Files:**
- Modify: `h5app/scripts/check-workflow-module.mjs`
- Modify: `h5app/src/pages/workflow/components/WorkflowStartPage.vue`

- [ ] **Step 1: Write the failing structure checks**

Require `class="workflow-start-page__form-card"`, its white card style, and the ordered action handlers:

```js
patterns: [
  'class="workflow-start-page__form-card"',
  '.workflow-start-page__form-card {',
  'background: #ffffff;',
]

patterns: [
  '@click="cancelStart"',
  '@click="submitStart"',
  '@click="saveDraft"',
]
```

- [ ] **Step 2: Run the check and verify it fails**

Run: `cd h5app && node scripts/check-workflow-module.mjs`

Expected: FAIL because the form card class is missing and the current action order places `saveDraft` before `submitStart`.

- [ ] **Step 3: Implement the minimal page change**

Wrap only the runtime form:

```vue
<view class="workflow-start-page__form-card">
  <WorkflowRuntimeForm ... />
</view>
```

Style the wrapper as a single responsive white card with a 6px radius, border, subtle shadow, and reduced mobile padding. Move the existing submit button block before the existing draft button block without changing their props or handlers.

- [ ] **Step 4: Run focused and full verification**

Run:

```bash
cd h5app
node scripts/check-workflow-module.mjs
corepack pnpm exec eslint src/pages/workflow/components/WorkflowStartPage.vue scripts/check-workflow-module.mjs
corepack pnpm type-check
corepack pnpm build:h5
```

Expected: all commands exit with status 0 and the H5 production build completes.

- [ ] **Step 5: Review the diff**

Run: `git -C h5app diff --check`

Expected: no whitespace errors; only the page and workflow structure check contain implementation changes for this request.
