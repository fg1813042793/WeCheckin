# H5App Workflow History Import Wizard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 H5App 历史表单导入弹窗改为选择、配置、预览三步向导，并优化顶部关闭按钮。

**Architecture:** 保留现有历史查询、详情读取、映射和预览纯函数。组件新增步骤状态并用互斥模板渲染当前步骤；详情加载成功后自动推进，底部操作根据步骤切换。关闭按钮使用组件内稳定尺寸的图标控件。

**Tech Stack:** Vue 3、uni-app、TypeScript、uView Pro、Node 结构检查、SCSS。

---

### Task 1: Add failing wizard structure checks

**Files:**
- Modify: `h5app/scripts/check-workflow-history-import.mjs`

- [ ] **Step 1: Add assertions for wizard state and exclusive pages**

Require these snippets:

```text
type HistoryImportStep = 'select' | 'configure' | 'preview'
const currentStep = ref<HistoryImportStep>('select')
currentStep.value = 'configure'
v-if="currentStep === 'select'"
v-else-if="currentStep === 'configure'"
v-else
goToPreviousStep
goToPreview
```

Reject the old numbered section headings `1. 选择历史申请`, `2. 配置导入方式`, and `3. 导入预览`.

- [ ] **Step 2: Add assertions for the compact close button**

Require `workflow-history-import-dialog__close`, `title="关闭"`, `role="button"`, `tabindex="0"`, Enter and Space keyboard handlers, and `width: 32px` / `height: 32px`.

- [ ] **Step 3: Run the check and verify RED**

```bash
cd h5app
pnpm check:workflow-history-import
```

Expected: FAIL because the dialog still renders all three sections together.

### Task 2: Implement the three-step wizard

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowHistoryImportDialog.vue`

- [ ] **Step 1: Add step state and metadata**

Add:

```ts
type HistoryImportStep = 'select' | 'configure' | 'preview'
const currentStep = ref<HistoryImportStep>('select')
const stepNumber = computed(() => ({ select: 1, configure: 2, preview: 3 })[currentStep.value])
const stepTitle = computed(() => ({
  select: '选择历史申请',
  configure: '配置导入方式',
  preview: '导入预览',
})[currentStep.value])
```

Reset `currentStep` to `select` in `openDialog`.

- [ ] **Step 2: Advance after successful source loading**

At the end of the successful `selectInstance` branch, after mappings are initialized, set:

```ts
currentStep.value = 'configure'
```

Keep detail errors on the selection step.

- [ ] **Step 3: Add navigation functions**

Implement:

```ts
function goToPreviousStep() {
  currentStep.value = currentStep.value === 'preview' ? 'configure' : 'select'
}

function goToPreview() {
  if (!sourceDetail.value || previewError.value)
    return
  currentStep.value = 'preview'
}
```

- [ ] **Step 4: Render only the active step**

Move the existing selection, configuration, and preview markup into three mutually exclusive containers. The header shows `stepTitle` and `步骤 {{ stepNumber }}/3`; it does not repeat numbered section headings.

- [ ] **Step 5: Switch footer actions by step**

- Selection: “取消”。No next button because successful selection advances automatically.
- Configuration: “上一步”“下一步”。
- Preview: “上一步”“确认导入”。

Keep `canApply` as the confirmation guard.

### Task 3: Optimize close button and responsive styles

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowHistoryImportDialog.vue`

- [ ] **Step 1: Replace the u-button close control**

Use a `view` with close icon, accessible attributes, click and keyboard handlers.

- [ ] **Step 2: Add stable close dimensions and focus states**

Set the close control to `32px × 32px`, prevent shrinking, and add hover/focus-visible states without changing the dialog dimensions.

- [ ] **Step 3: Keep one scrollable step body**

Ensure header and footer remain fixed while the active step body is the only scroll container on PC and mobile.

### Task 4: Verify the wizard

**Files:** None.

- [ ] **Step 1: Run focused checks**

```bash
cd h5app
pnpm check:workflow-history-import
pnpm exec eslint src/pages/workflow/components/WorkflowHistoryImportDialog.vue scripts/check-workflow-history-import.mjs
pnpm type-check
pnpm build:h5
```

Expected: PASS.

- [ ] **Step 2: Run diff checks**

```bash
git diff --check
git status --short
```

Confirm unrelated dirty files remain untouched. Browser interaction remains for user acceptance.
