# H5App Workflow History Form Import Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 H5App 流程发起页增加历史表单导入能力，支持当前流程历史申请选择、全部导入、字段映射、差异预览和可选覆盖已有内容。

**Architecture:** 使用现有历史实例列表和详情接口读取实例绑定版本的 Schema 与 `formData`。新增纯 TypeScript 映射模块负责兼容性、建议、预览和补丁；独立弹窗负责交互；发起页只负责权限、显隐和把补丁合并到当前表单。后端仅新增按钮权限目录和迁移，不新增导入 API 或业务表。

**Tech Stack:** Vue 3、uni-app、TypeScript、Pinia、uView Pro、Node 结构/纯函数检查、Go 权限目录、MySQL migrations。

---

## File Map

- Create: `h5app/src/pages/workflow/workflow-history-import.ts` — 字段展平、类型兼容、自动映射、手工映射校验、预览和补丁生成。
- Create: `h5app/scripts/check-workflow-history-import.mjs` — 运行纯函数用例并检查组件接线。
- Modify: `h5app/package.json` — 增加历史导入检查命令。
- Create: `h5app/src/pages/workflow/components/WorkflowHistoryImportDialog.vue` — 历史实例选择、模式切换、映射编辑、覆盖开关和预览。
- Modify: `h5app/src/pages/workflow/components/WorkflowStartPage.vue` — 导入按钮、权限判断、弹窗接线、补丁合并和重算。
- Modify: `backend/internal/support/appmenuperm/catalog.go` — 增加“导入历史表单”按钮权限。
- Create: `backend/internal/support/appmenuperm/workflow_history_import_test.go` — 校验按钮权限目录。
- Create: `backend/migrations/20260923100000_add_dingtalk_h5_workflow_history_import_permission.sql` — 注册按钮权限并从 start 权限回填 allow 授权。
- Create: `backend/test/internal/bootstrap/migrations/workflow_history_import_permission_migration_test.go` — 校验迁移。
- Modify: `docs/PERMISSION_CODE_FRONTEND_SYNC.md` — 记录导入按钮权限组合。

### Task 1: Build history import mapping as pure functions

**Files:**
- Create: `h5app/scripts/check-workflow-history-import.mjs`
- Create: `h5app/src/pages/workflow/workflow-history-import.ts`
- Modify: `h5app/package.json`

- [ ] **Step 1: Write the failing pure-function check**

Create a script that transpiles `workflow-calculation.ts`, `workflow-form.ts` and `workflow-history-import.ts` to CommonJS with the project TypeScript dependency and executes real functions. Evaluate the three modules in memory with a custom `require` that returns the already evaluated calculation module for `./workflow-calculation` and the form module for `./workflow-form`; type-only `@/types/workflow` imports disappear during transpilation. This avoids adding a test runner or modifying production imports. Cover at least these cases:

```js
assert.equal(isWorkflowImportTypeCompatible(textField, textareaField), true)
assert.equal(isWorkflowImportTypeCompatible(numberField, amountField), true)
assert.equal(isWorkflowImportTypeCompatible(dateField, datetimeField), false)
assert.equal(isWorkflowImportTypeCompatible(calculationField, numberField), false)

assert.deepEqual(
  buildWorkflowImportMappings(sourceFields, targetFields, targetAccess),
  [{ sourceKey: 'remark', targetKey: 'remark' }],
)

assert.deepEqual(
  suggestWorkflowImportMappings(renamedSourceFields, renamedTargetFields, targetAccess),
  [{ sourceKey: 'oldRemark', targetKey: 'remark' }],
)
```

Add preview tests for:

- Default mode keeps non-empty target values.
- Overwrite mode replaces mapped targets only.
- `0` and `false` count as non-empty.
- Static option values missing from the current definition are filtered or skipped.
- Calculation and read-only targets never produce a patch.
- Duplicate target mappings are rejected.
- Compatible detail lists rebuild target rows and drop unknown columns.

Add `check:workflow-history-import` to `h5app/package.json`.

- [ ] **Step 2: Run the check and verify RED**

```bash
cd h5app
pnpm check:workflow-history-import
```

Expected: FAIL because the TypeScript module does not exist.

- [ ] **Step 3: Implement types and field helpers**

Define:

```ts
export interface WorkflowHistoryImportMapping {
  sourceKey: string
  targetKey: string
}

export type WorkflowHistoryImportResult
  = 'import' | 'overwrite' | 'retain' | 'incompatible' | 'filtered' | 'dynamic_option' | 'skipped'

export interface WorkflowHistoryImportPreviewItem {
  sourceKey: string
  sourceLabel: string
  targetKey: string
  targetLabel: string
  currentValue: unknown
  importValue: unknown
  result: WorkflowHistoryImportResult
  message: string
}

export interface WorkflowHistoryImportPreview {
  items: WorkflowHistoryImportPreviewItem[]
  patch: WorkflowFormData
  importCount: number
  overwriteCount: number
  retainCount: number
  skipCount: number
}
```

Reuse `workflowDataFields`, `cloneWorkflowValue`, `emptyWorkflowFieldValue`, `normalizeWorkflowFormValue` and `workflowDetailRowKey` from `workflow-form.ts` rather than duplicating form normalization.

- [ ] **Step 4: Implement compatibility and mapping**

Implement explicit type families:

```ts
const stringTypes = new Set(['text', 'textarea', 'phone', 'email'])
const numberTypes = new Set(['number', 'amount'])
const singleOptionTypes = new Set(['select', 'radio'])
const multiOptionTypes = new Set(['multi_select', 'checkbox'])
```

Exact-only types are date/time, user, department, boolean, attachment and detail list. Reject layout fields and calculations. `buildWorkflowImportMappings` matches same keys only; `suggestWorkflowImportMappings` first uses same key, then a unique same-label compatible target. Both functions accept the current start-node access map and include only `write` targets.

- [ ] **Step 5: Implement preview and patch generation**

Implement `previewWorkflowHistoryImport` with this signature:

```ts
export function previewWorkflowHistoryImport(input: {
  sourceFields: WorkflowFormField[]
  sourceData: WorkflowFormData
  targetFields: WorkflowFormField[]
  targetData: WorkflowFormData
  targetAccess: WorkflowFieldAccessMap
  mappings: WorkflowHistoryImportMapping[]
  overwrite: boolean
}): WorkflowHistoryImportPreview
```

Normalize values against target fields. For static options, filter against current options. For API-backed options, retain type-correct values and mark `dynamic_option`. Detail lists require compatible target columns and create fresh target row IDs through existing helpers.

- [ ] **Step 6: Run pure-function check and verify GREEN**

```bash
cd h5app
pnpm check:workflow-history-import
pnpm exec eslint src/pages/workflow/workflow-history-import.ts scripts/check-workflow-history-import.mjs
```

Expected: PASS.

- [ ] **Step 7: Commit Task 1**

```bash
git add h5app/src/pages/workflow/workflow-history-import.ts \
  h5app/scripts/check-workflow-history-import.mjs \
  h5app/package.json
git commit -m "feat: 增加流程历史表单导入规则"
```

### Task 2: Add the history import button permission

**Files:**
- Create: `backend/internal/support/appmenuperm/workflow_history_import_test.go`
- Modify: `backend/internal/support/appmenuperm/catalog.go`
- Create: `backend/test/internal/bootstrap/migrations/workflow_history_import_permission_migration_test.go`
- Create: `backend/migrations/20260923100000_add_dingtalk_h5_workflow_history_import_permission.sql`

- [ ] **Step 1: Write failing permission catalog test**

Assert a declaration exists with:

```go
Declaration{
    Key:       "dingtalk_h5:button:workflow:history-import",
    Name:      "导入历史表单",
    Platform:  "dingtalk_h5",
    Path:      "workflow:history-import",
    ParentKey: "dingtalk_h5:menu:workflow",
}
```

- [ ] **Step 2: Write failing migration test**

The migration contract test must assert:

- The new button key and display name.
- Source permission `dingtalk_h5:api:workflow:start`.
- Only active `allow` grants are copied.
- `ON DUPLICATE KEY UPDATE` is present.
- The migration does not add a new API permission.

- [ ] **Step 3: Run tests and verify RED**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/appmenuperm ./test/internal/bootstrap/migrations -run 'WorkflowHistoryImport' -count=1
```

Expected: FAIL because the declaration and migration do not exist.

- [ ] **Step 4: Implement catalog declaration and migration**

Add the button after the four workflow task action buttons and before form-revision actions. The migration must insert the permission idempotently and copy active allow grants from `dingtalk_h5:api:workflow:start` for the same subject type and ID.

Do not copy deny grants: a start deny already prevents entering/submitting the start page, while this button is an additional UI capability.

- [ ] **Step 5: Run tests and verify GREEN**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/appmenuperm ./test/internal/bootstrap/migrations -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit Task 2**

```bash
git add backend/internal/support/appmenuperm/catalog.go \
  backend/internal/support/appmenuperm/workflow_history_import_test.go \
  backend/migrations/20260923100000_add_dingtalk_h5_workflow_history_import_permission.sql \
  backend/test/internal/bootstrap/migrations/workflow_history_import_permission_migration_test.go
git commit -m "feat: 增加流程历史表单导入权限"
```

### Task 3: Build the history import dialog

**Files:**
- Create: `h5app/src/pages/workflow/components/WorkflowHistoryImportDialog.vue`
- Modify: `h5app/scripts/check-workflow-history-import.mjs`

- [ ] **Step 1: Extend the structure check and verify RED**

Require the component to contain:

```text
listWorkflowInstances
getWorkflowInstance
scope: 'started'
definitionId
全部导入
字段映射
覆盖已有内容
previewWorkflowHistoryImport
WorkflowHistoryImportMapping
emit('apply'
```

Also require pagination, title search, status filter, loading, empty and retry states.

Run `pnpm check:workflow-history-import` and confirm it fails because the component is missing.

- [ ] **Step 2: Define component contract**

Use props and emits:

```ts
const props = defineProps<{
  modelValue: boolean
  definition: WorkflowPublishedDefinition
  targetData: WorkflowFormData
  targetAccess: WorkflowFieldAccessMap
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'apply': [payload: { sourceInstanceId: string, patch: WorkflowFormData, importCount: number }]
}>()
```

The component may call workflow list/detail APIs but must not access Pinia or submit/start workflows.

- [ ] **Step 3: Implement historical instance selection**

On open, reset transient state and call:

```ts
listWorkflowInstances({
  definitionId: props.definition.id,
  scope: 'started',
  instanceTitle: filters.instanceTitle.trim() || undefined,
  status: filters.status || undefined,
  page,
  pageSize,
})
```

Selecting a row calls `getWorkflowInstance(instance.id)`. Store the returned source `form`, `formData`, version and instance ID. If the current target definition version changes while open, close and require reopening.

- [ ] **Step 4: Implement mode, mapping and preview UI**

- Use a segmented control for `全部导入` and `字段映射`.
- Generate same-key mappings in all mode.
- Initialize manual mode from `suggestWorkflowImportMappings`.
- Provide source/target selects, add/remove row commands and prevent duplicate selections.
- Add the default-off overwrite switch.
- Recompute preview whenever source, mode, mappings or overwrite changes.
- Render preview result labels and summary counts.
- Disable confirmation when `preview.importCount + preview.overwriteCount === 0`.

- [ ] **Step 5: Implement responsive dialog states**

PC uses a centered popup with a constrained width and scrollable content. Mobile uses bottom placement, full width, safe-area padding and fixed actions. Include loading, empty, load-error/retry and source-detail-error states.

- [ ] **Step 6: Verify component structure and types**

```bash
cd h5app
pnpm check:workflow-history-import
pnpm exec eslint src/pages/workflow/components/WorkflowHistoryImportDialog.vue
pnpm type-check
```

Expected: PASS.

- [ ] **Step 7: Commit Task 3**

```bash
git add h5app/src/pages/workflow/components/WorkflowHistoryImportDialog.vue \
  h5app/scripts/check-workflow-history-import.mjs
git commit -m "feat: 增加历史表单导入弹窗"
```

### Task 4: Integrate import into the workflow start page

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowStartPage.vue`
- Modify: `h5app/scripts/check-workflow-history-import.mjs`
- Modify: `h5app/scripts/check-workflow-module.mjs`

- [ ] **Step 1: Add failing start-page structure assertions**

Require:

```text
dingtalk_h5:button:workflow:history-import
const canImportHistory = computed
WorkflowHistoryImportDialog
v-model="historyImportVisible"
@apply="applyHistoryImport"
导入
```

Require `canImportHistory` to combine button, start, view and `starterAllowed`. Require `applyHistoryImport` to merge the patch with current data and call `initialWorkflowFormData` using the current definition form.

Run the check and verify RED.

- [ ] **Step 2: Add permission and dialog state**

Add:

```ts
const historyImportVisible = ref(false)
const canImportHistory = computed(() => Boolean(
  starterAllowed.value
  && auth.hasButtonPermission('dingtalk_h5:button:workflow:history-import')
  && auth.hasApiPermission('dingtalk_h5:api:workflow:start')
  && auth.hasApiPermission('dingtalk_h5:api:workflow:view'),
))
```

Import and render `WorkflowHistoryImportDialog` only when the definition exists.

- [ ] **Step 3: Add the form-card toolbar**

Add a compact toolbar above `WorkflowRuntimeForm`. Keep the form as the primary surface; use an icon plus “导入” text button, no marketing copy or nested card. Disable it while the start page is busy.

- [ ] **Step 4: Merge an applied import**

Implement:

```ts
function applyHistoryImport(payload: { patch: WorkflowFormData, importCount: number }) {
  const value = definition.value
  if (!value)
    return
  formData.value = initialWorkflowFormData(value.form || [], {
    ...formData.value,
    ...payload.patch,
  })
  activeSection.value = 'start'
  uni.showToast({ title: `已导入 ${payload.importCount} 个字段`, icon: 'success' })
}
```

Do not update `savedSnapshot`; the existing unsaved-change guard must detect the import.

- [ ] **Step 5: Verify start-page integration**

```bash
cd h5app
pnpm check:workflow-history-import
pnpm check:workflow-module
pnpm exec eslint src/pages/workflow/components/WorkflowStartPage.vue \
  src/pages/workflow/components/WorkflowHistoryImportDialog.vue \
  src/pages/workflow/workflow-history-import.ts \
  scripts/check-workflow-history-import.mjs
pnpm type-check
```

Expected: PASS, or report unrelated concurrent workflow-page checks separately.

- [ ] **Step 6: Commit Task 4**

```bash
git add h5app/src/pages/workflow/components/WorkflowStartPage.vue \
  h5app/scripts/check-workflow-history-import.mjs \
  h5app/scripts/check-workflow-module.mjs
git commit -m "feat: 接入流程历史表单导入"
```

### Task 5: Document and verify the feature

**Files:**
- Modify: `docs/PERMISSION_CODE_FRONTEND_SYNC.md`

- [ ] **Step 1: Document the permission combination**

Add a section explaining that the import button requires the new button permission plus existing workflow start and view APIs. State that no import API exists and formal submission remains the server validation boundary.

- [ ] **Step 2: Run backend focused verification**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/appmenuperm ./test/internal/bootstrap/migrations -count=1
```

Expected: PASS.

- [ ] **Step 3: Run H5App focused verification**

```bash
cd h5app
pnpm check:workflow-history-import
pnpm type-check
pnpm build:h5
```

Expected: PASS.

- [ ] **Step 4: Run broader verification**

```bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./...

cd ../h5app
pnpm lint
pnpm check:workflow-module
```

Report existing unrelated failures separately. Do not fix concurrent Admin, workflow-definition lifecycle, history resizer or other unrelated changes.

- [ ] **Step 5: Run diff checks**

```bash
git diff --check
git status --short
```

Confirm only intended files were committed and all unrelated worktree changes remain untouched.

- [ ] **Step 6: Commit documentation**

```bash
git add docs/PERMISSION_CODE_FRONTEND_SYNC.md
git commit -m "docs: 说明流程历史表单导入权限"
```

### Task 6: Manual acceptance matrix

**Files:** None.

- [ ] **Step 1: Apply the permission migration and re-login**

Run the repository migration command for the target environment. Re-login to refresh `buttonPermissionKeys`.

- [ ] **Step 2: Verify permission combinations**

- Start + view + history-import button: import button is visible.
- Start + view without history-import button: import button is hidden.
- Start + history-import without view: import button is hidden and history API remains denied.

- [ ] **Step 3: Verify all-import behavior**

Use same-version and older-version historical forms. Confirm same-key compatible fields import, calculations recompute, removed fields are skipped and current-only fields remain unchanged.

- [ ] **Step 4: Verify field mapping behavior**

Map renamed text and number fields, reject incompatible mappings, prevent duplicate targets, and verify manual same-label suggestions remain editable.

- [ ] **Step 5: Verify overwrite behavior**

With overwrite disabled, populated values including `0` and `false` remain unchanged. With overwrite enabled, only mapped fields are replaced.

- [ ] **Step 6: Verify complex fields and responsive UI**

Check attachments, static/dynamic options, compatible detail lists, PC pagination/search and mobile bottom-popup scrolling/fixed actions.

- [ ] **Step 7: Verify draft and submission flows**

After import, confirm the page reports unsaved changes, saved drafts restore imported values, and formal submission creates a new instance without modifying the source history record.
