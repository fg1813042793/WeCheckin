# Workflow Detail Preview Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep every workflow detail-column label directly above its own input preview while preserving the configured 24-column layout.

**Architecture:** Replace the two independent label/control grids in both designer preview surfaces with one 24-column grid of field cells. Each cell owns its label and control preview, so the pair wraps as one unit without changing workflow form data, runtime rendering, or backend contracts.

**Tech Stack:** Vue 3 SFC, TypeScript, Element Plus, CSS Grid, Node.js structural checks, Vite, in-app Browser QA.

---

## File Map

- `admin/scripts/check-workflow-form-designer.mjs`: structural regression contract for both detail preview surfaces.
- `admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue`: root-level detail field preview.
- `admin/src/views/workflow/designer/components/WorkflowFormFieldPreview.vue`: detail preview nested inside a form group.

### Task 1: Add the Failing Structural Regression Check

**Files:**

- Modify: `admin/scripts/check-workflow-form-designer.mjs:126`
- Test: `admin/scripts/check-workflow-form-designer.mjs`

- [ ] **Step 1: Add the detail preview cell-contract check before `requirements`**

Insert this block immediately before `const requirements = [`:

```js
const detailPreviewRequirements = [
  ['detail-preview__grid', '缺少统一明细列网格'],
  ['detail-preview__cell', '明细字段名称与输入预览未绑定到同一网格单元'],
  ['detail-preview__label', '明细字段单元缺少名称'],
  ['detail-preview__control', '明细字段单元缺少输入预览'],
]

for (const [source, surface] of [
  [formDesignerSource, '根级明细预览'],
  [fieldPreviewSource, '分组内明细预览'],
]) {
  for (const [snippet, message] of detailPreviewRequirements) {
    if (!source.includes(snippet)) throw new Error(`${surface}${message}`)
  }
  for (const legacyClass of ['detail-preview__head', 'detail-preview__row']) {
    if (source.includes(legacyClass)) throw new Error(`${surface}仍使用独立名称和输入网格：${legacyClass}`)
  }
}
```

- [ ] **Step 2: Run the structural check and verify RED**

Run from `admin/`:

```bash
npm run check:workflow-form-designer
```

Expected: FAIL with `根级明细预览缺少统一明细列网格`. The failure must come from the new contract, not a syntax or dependency error.

### Task 2: Bind Labels and Controls in Both Preview Surfaces

**Files:**

- Modify: `admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue:255`
- Modify: `admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue:1238`
- Modify: `admin/src/views/workflow/designer/components/WorkflowFormFieldPreview.vue:6`
- Modify: `admin/src/views/workflow/designer/components/WorkflowFormFieldPreview.vue:91`
- Test: `admin/scripts/check-workflow-form-designer.mjs`

- [ ] **Step 1: Replace the root-level detail preview's two grids with field cells**

Replace the `detail-preview__head` and `detail-preview__row` blocks inside `WorkflowFormDesigner.vue` with:

```vue
<div class="detail-preview__grid">
  <div
    v-for="column in detailColumns(field)"
    :key="column.key"
    class="detail-preview__cell"
    :style="{ gridColumn: `span ${fieldSpan(column)}` }"
  >
    <span class="detail-preview__label">{{ column.label || column.key }}</span>
    <span class="detail-preview__control">{{ detailColumnPlaceholder(column) }}</span>
  </div>
</div>
```

Keep the existing disabled `新增行` button immediately after this grid.

- [ ] **Step 2: Replace the root-level detail preview styles**

Remove `.detail-preview__head`, `.detail-preview__row`, and their descendant rules. Add:

```css
.detail-preview__grid {
  display: grid;
  grid-template-columns: repeat(24, minmax(0, 1fr));
  gap: 8px;
}
.detail-preview__cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}
.detail-preview__label {
  min-width: 0;
  overflow: hidden;
  color: #64748b;
  font-size: 11px;
  font-weight: 650;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.detail-preview__control {
  display: block;
  min-height: 28px;
  padding: 6px 8px;
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 5px;
  color: #64748b;
  background: #fff;
  font-size: 11px;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}
```

- [ ] **Step 3: Apply the same cell structure to the grouped-field preview**

Replace the `detail-preview__head` and `detail-preview__row` blocks inside `WorkflowFormFieldPreview.vue` with:

```vue
<div class="detail-preview__grid">
  <div
    v-for="column in field.columns || []"
    :key="column.key"
    class="detail-preview__cell"
    :style="{ gridColumn: `span ${columnSpan(column)}` }"
  >
    <span class="detail-preview__label">{{ column.label || column.key }}</span>
    <span class="detail-preview__control">{{ column.placeholder || '请输入' }}</span>
  </div>
</div>
```

Keep the existing disabled `新增行` button immediately after this grid.

- [ ] **Step 4: Apply the same cell styles to the grouped-field preview**

Remove `.detail-preview__head`, `.detail-preview__row`, and their descendant rules from `WorkflowFormFieldPreview.vue`. Add:

```css
.detail-preview__grid {
  display: grid;
  grid-template-columns: repeat(24, minmax(0, 1fr));
  gap: 8px;
}
.detail-preview__cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}
.detail-preview__label {
  min-width: 0;
  overflow: hidden;
  color: #64748b;
  font-size: 11px;
  font-weight: 650;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.detail-preview__control {
  display: block;
  min-height: 28px;
  padding: 6px 8px;
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 5px;
  color: #64748b;
  background: #fff;
  font-size: 11px;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}
```

Keep `.detail-preview` and its existing media query unchanged.

- [ ] **Step 5: Run the focused structural check and verify GREEN**

Run from `admin/`:

```bash
npm run check:workflow-form-designer
```

Expected: exit code 0 and `workflow form designer checks passed`.

- [ ] **Step 6: Check formatting and commit the focused change**

Run from the repository root:

```bash
git diff --check -- \
  admin/scripts/check-workflow-form-designer.mjs \
  admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue \
  admin/src/views/workflow/designer/components/WorkflowFormFieldPreview.vue
git add \
  admin/scripts/check-workflow-form-designer.mjs \
  admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue \
  admin/src/views/workflow/designer/components/WorkflowFormFieldPreview.vue
git commit -m "fix: align workflow detail preview fields"
```

Expected: diff check exits 0 and the commit contains only these three files.

### Task 3: Run Full Admin and Rendered QA

**Files:**

- Verify: `admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue`
- Verify: `admin/src/views/workflow/designer/components/WorkflowFormFieldPreview.vue`

- [ ] **Step 1: Run the complete Admin gate**

Run from `admin/`:

```bash
npm run check:all
```

Expected: all structural checks, TypeScript build, Vite build, and bundle budget checks pass.

- [ ] **Step 2: Start or reuse the Admin development server**

Check port 5173:

```bash
lsof -nP -iTCP:5173 -sTCP:LISTEN
```

If no Admin server is listening, run from `admin/` and keep the process active:

```bash
npm run dev -- --host 127.0.0.1 --port 5173
```

If another process owns 5173, start this Admin instance on 5174:

```bash
npm run dev -- --host 127.0.0.1 --port 5174
```

- [ ] **Step 3: Validate the workflow designer in the in-app Browser**

Use the Browser runtime required by `browser:control-in-app-browser` and keep one tab bound throughout the check:

1. Name the session `WeCheckin workflow detail preview alignment`.
2. Reuse a signed-in workflow designer tab when available; otherwise navigate to `http://127.0.0.1:5173/workflow/definitions` or the selected fallback port.
3. Open `/workflow/definitions/:id/designer`, select `表单设计`, and inspect a detail field configured as `24, 12, 12, 24`.
4. Confirm URL/title identity, meaningful DOM content, no Vite/framework error overlay, and no relevant console errors or warnings.
5. Confirm the rendered DOM contains four `.detail-preview__cell` elements and each contains one `.detail-preview__label` followed by one `.detail-preview__control`.
6. Capture a desktop screenshot around 1440x900 showing `目标`, `权重`, `完成度`, and `结果` directly above their controls.
7. Repeat at a narrow desktop viewport around 960x900 and confirm no overlap, clipping, or detached labels.
8. Exercise one interaction by selecting the detail field or changing a column width and verify the label/control pair moves together after the UI updates.

Expected: both viewport checks pass and screenshot evidence shows label/control pairing for every configured span.

- [ ] **Step 4: Run final source and patch checks**

Run from the repository root:

```bash
if rg -n 'detail-preview__(head|row)' \
  admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue \
  admin/src/views/workflow/designer/components/WorkflowFormFieldPreview.vue; then exit 1; fi
git diff --check
git status --short
```

Expected: no legacy class match; `git diff --check` exits 0; status preserves unrelated user changes.
