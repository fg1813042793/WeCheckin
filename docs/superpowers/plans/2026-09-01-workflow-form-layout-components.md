# Workflow Form Layout Components Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为通用工作流表单增加一级固定分组、标签、静态说明、说明按钮和字段级说明弹窗，同时保持现有表单数据与节点权限语义。

**Architecture:** 继续使用 `Definition.Form` 作为唯一表单定义，`group.fields` 只建立布局树，组内业务字段值仍位于 `formData[field.key]`。后端递归校验并展平业务字段；admin 提供纯函数处理展平、可见性和拖拽定位，设计器与运行时只负责交互和渲染。

**Tech Stack:** Go 1.24.5, Hertz/GORM workflow module, Vue 3, TypeScript, Element Plus, Vite, Node static regression scripts.

---

## File Map

- Modify `backend/internal/workflow/types.go`: add layout field types, `FormHelp`, `FormField.Fields/Content/Help`, and validation codes.
- Modify `backend/internal/workflow/validation.go`: recursively validate root/group/detail contexts and return only permission-addressable business fields.
- Modify `backend/internal/workflow/form.go`: flatten group business fields for start data validation and task patch validation.
- Modify `backend/internal/modules/workflow/infrastructure/gorm_store.go`: deep-clone group fields and help configuration in published definitions.
- Modify `backend/internal/workflow/definition_test.go`: schema, form-data, task-patch, and invalid-layout tests.
- Modify `backend/internal/modules/workflow/application/service_test.go`: application-level group field start/complete regression.
- Modify `admin/src/views/workflow/types.ts`: mirror the new schema.
- Create `admin/src/views/workflow/formLayout.ts`: pure recursive field classification, lookup, flatten, removal, and cross-container movement helpers.
- Modify `admin/src/views/workflow/runtimeForm.ts`: recursive initial values, access/actions, visible layout, normalization, and remote-option traversal support.
- Modify `admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue`: palette, group canvas, nested drag/drop, properties, deletion confirmation, and help editing.
- Modify `admin/src/views/workflow/designer/components/WorkflowFieldPermissions.vue`: show flattened business fields only.
- Modify `admin/src/views/workflow/designer/index.vue`: recursively maintain node permissions after form changes.
- Modify `admin/src/views/workflow/components/WorkflowRuntimeForm.vue`: recursively render groups/display components and shared plain-text help dialog.
- Modify `admin/scripts/check-workflow-form-designer.mjs`: pure helper and designer source checks.
- Modify `admin/scripts/check-workflow-runtime-form.mjs`: recursive form/runtime checks.
- Modify `admin/scripts/check-workflow-runtime-pages.mjs`: page integration regression snippets when required.

## Task 1: Backend Layout Schema

**Files:**
- Modify: `backend/internal/workflow/types.go`
- Modify: `backend/internal/workflow/validation.go`
- Test: `backend/internal/workflow/definition_test.go`

- [ ] **Step 1: Write failing schema tests**

Add a valid definition containing a group, label, description, button, data-field help, and a detail list inside the group:

```go
func TestValidateDefinitionAcceptsFormLayoutComponents(t *testing.T) {
	definition := validLinearDefinition()
	definition.Form = []FormField{{
		Key: "basic_group", Label: "基本信息", Type: FormFieldTypeGroup,
		Help: &FormHelp{ButtonText: "查看说明", Title: "基本信息说明", Content: "请如实填写。"},
		Fields: []FormField{
			{Key: "basic_label", Label: "申请信息", Type: FormFieldTypeLabel},
			{Key: "basic_tip", Label: "填写提示", Type: FormFieldTypeDescription, Content: "请核对后提交。"},
			{Key: "reason", Label: "申请事由", Type: FormFieldTypeTextarea, Required: true,
				Help: &FormHelp{Title: "事由说明", Content: "请输入完整事由。"}},
			{Key: "rules", Label: "查看填写规则", Type: FormFieldTypeButton,
				Help: &FormHelp{Title: "填写规则", Content: "仅填写本次申请。"}},
		},
	}}
	definition.Nodes[0].FormPermissions = []FieldPermission{{Field: "reason", Access: FieldAccessWrite}}
	if errors := ValidateDefinition(definition); len(errors) != 0 {
		t.Fatalf("expected layout form to be valid, got %#v", errors)
	}
}
```

Extend the invalid-schema table with nested groups, duplicate root/group keys, empty description, button without help, oversized help, `fields` on text, and `content` on group. Assert the new codes `form_field_layout_invalid` or `form_field_help_invalid`.

- [ ] **Step 2: Run schema tests and verify RED**

Run:

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/workflow -run 'TestValidateDefinitionAcceptsFormLayoutComponents|TestValidateDefinitionRejectsInvalidFormSchema' -count=1
```

Expected: compile failure for missing `FormFieldTypeGroup`, `FormHelp`, `Fields`, `Content`, and `Help`.

- [ ] **Step 3: Add schema types and validation**

Add these definitions in `types.go`:

```go
const (
	FormFieldTypeGroup       = "group"
	FormFieldTypeLabel       = "label"
	FormFieldTypeDescription = "description"
	FormFieldTypeButton      = "button"
)

const (
	ValidationFormFieldLayout = "form_field_layout_invalid"
	ValidationFormFieldHelp   = "form_field_help_invalid"
)

type FormHelp struct {
	ButtonText string `json:"buttonText,omitempty"`
	Title      string `json:"title"`
	Content    string `json:"content"`
}
```

Extend `FormField`:

```go
Fields  []FormField `json:"fields,omitempty"`
Content string      `json:"content,omitempty"`
Help    *FormHelp   `json:"help,omitempty"`
```

Replace the current two-state nested flag with explicit root/group/detail traversal. Use one shared `layoutKeys` map for root and group children, but a local map for each detail-list column set. Add only real data fields to the map returned to `validateFieldPermissions`:

```go
func isFormLayoutFieldType(fieldType string) bool {
	return fieldType == FormFieldTypeGroup || fieldType == FormFieldTypeLabel ||
		fieldType == FormFieldTypeDescription || fieldType == FormFieldTypeButton
}

func validFormHelp(help *FormHelp) bool {
	if help == nil {
		return true
	}
	return utf8.RuneCountInString(strings.TrimSpace(help.ButtonText)) <= 30 &&
		strings.TrimSpace(help.Title) != "" && utf8.RuneCountInString(strings.TrimSpace(help.Title)) <= 100 &&
		strings.TrimSpace(help.Content) != "" && utf8.RuneCountInString(strings.TrimSpace(help.Content)) <= 2000
}
```

For `group`, require label, at least one child, reject child groups, recurse with the shared key map, and merge child business fields into the returned permission map. For `label`, require label length at most 100. For `description`, require non-empty content length at most 2000. For `button`, require label length at most 30 and non-nil valid help. Reject `Fields` outside groups, `Content` outside descriptions, and `Help` outside group/button/data fields.

- [ ] **Step 4: Run schema tests and verify GREEN**

Run the command from Step 2. Expected: PASS.

## Task 2: Backend Recursive Data, Patch, and Clone

**Files:**
- Modify: `backend/internal/workflow/form.go`
- Modify: `backend/internal/modules/workflow/infrastructure/gorm_store.go`
- Test: `backend/internal/workflow/definition_test.go`
- Test: `backend/internal/modules/workflow/application/service_test.go`

- [ ] **Step 1: Write failing recursive behavior tests**

Add tests proving:

```go
func TestValidateFormDataReadsBusinessFieldsInsideGroup(t *testing.T) {
	fields := []FormField{{Key: "group", Label: "分组", Type: FormFieldTypeGroup, Fields: []FormField{
		{Key: "tip", Label: "提示", Type: FormFieldTypeDescription, Content: "请填写"},
		{Key: "reason", Label: "原因", Type: FormFieldTypeText, Required: true},
	}}}
	if err := ValidateFormData(fields, map[string]interface{}{}, false); !errors.Is(err, ErrFormDataInvalid) {
		t.Fatalf("missing grouped field error = %v", err)
	}
	if err := ValidateFormData(fields, map[string]interface{}{"reason": "出差"}, false); err != nil {
		t.Fatalf("valid grouped data error = %v", err)
	}
}
```

Add a task-patch test where `reason` is in a group and writable; add a published-definition clone test that mutates draft `Fields[0].Help.Content` after cloning and verifies the published clone is unchanged.

- [ ] **Step 2: Run behavior tests and verify RED**

Run:

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/workflow ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure -run 'Group|Grouped|CloneFormFields' -count=1
```

Expected: grouped required fields are skipped or clone assertions fail.

- [ ] **Step 3: Implement recursive business-field traversal**

Add an internal helper in `form.go` and use it in `ValidateFormData`, `ValidateNodeFormPatch`, and field lookup:

```go
func dataFormFields(fields []FormField) []FormField {
	result := make([]FormField, 0, len(fields))
	for _, field := range fields {
		if field.Type == FormFieldTypeGroup {
			result = append(result, dataFormFields(field.Fields)...)
			continue
		}
		if isFormLayoutFieldType(field.Type) {
			continue
		}
		result = append(result, field)
	}
	return result
}
```

Keep detail-list columns inside their parent field. Do not create `formData[group.key]`.

Update `cloneFormFields`:

```go
cloned[index].Fields = cloneFormFields(field.Fields)
if field.Help != nil {
	help := *field.Help
	cloned[index].Help = &help
}
```

- [ ] **Step 4: Run recursive behavior tests and verify GREEN**

Run the command from Step 2. Expected: PASS.

## Task 3: Frontend Pure Form-Layout Helpers

**Files:**
- Modify: `admin/src/views/workflow/types.ts`
- Create: `admin/src/views/workflow/formLayout.ts`
- Modify: `admin/src/views/workflow/runtimeForm.ts`
- Modify: `admin/scripts/check-workflow-form-designer.mjs`
- Modify: `admin/scripts/check-workflow-runtime-form.mjs`

- [ ] **Step 1: Write failing pure-helper checks**

In the static scripts, load `formLayout.ts` and assert:

```js
const layout = loadTypeScriptModule('src/views/workflow/formLayout.ts')
const groupedFields = [{
  key: 'basic_group', label: '基本信息', type: 'group', fields: [
    { key: 'tip', label: '提示', type: 'description', content: '请填写' },
    { key: 'reason', label: '原因', type: 'text', required: true },
  ],
}]
assert(layout.workflowDataFields(groupedFields).map(item => item.key).join(',') === 'reason', '应只展平组内业务字段')
assert(layout.moveWorkflowField(groupedFields, 'reason', null, 0), '组内字段应可移到根级')
assert(!layout.moveWorkflowField(groupedFields, 'basic_group', 'another_group', 0), '表单组不应嵌套')
```

Extend runtime checks so `initialWorkflowFormData(groupedFields, { reason: '出差' })` returns only `{ reason: '出差' }`, access/action maps include grouped data fields, and `visibleWorkflowFormFields` preserves a group while filtering hidden children.

- [ ] **Step 2: Run helper checks and verify RED**

Run:

```bash
cd admin
npm run check:workflow-form-designer
npm run check:workflow-runtime-form
```

Expected: missing module/functions/types.

- [ ] **Step 3: Implement schema and pure helpers**

In the existing `WorkflowFormFieldType` literal union in `types.ts`, append `'group' | 'label' | 'description' | 'button'`, then add:

```ts
export interface WorkflowFormHelp {
  buttonText?: string
  title: string
  content: string
}
```

Extend `WorkflowFormField` with `fields?: WorkflowFormField[]`, `content?: string`, and `help?: WorkflowFormHelp`.

Create `formLayout.ts` exporting:

```ts
export function isWorkflowLayoutField(field: Pick<WorkflowFormField, 'type'>): boolean
export function isWorkflowDataField(field: Pick<WorkflowFormField, 'type'>): boolean
export function workflowDataFields(fields: WorkflowFormField[]): WorkflowFormField[]
export function workflowFieldByKey(fields: WorkflowFormField[], key: string): WorkflowFormField | undefined
export function removeWorkflowField(fields: WorkflowFormField[], key: string): WorkflowFormField | undefined
export function moveWorkflowField(fields: WorkflowFormField[], sourceKey: string, targetGroupKey: string | null, targetIndex: number): boolean
```

`moveWorkflowField` must restore the source field when the target is invalid and reject moving `group` into any group. Update runtime helpers to use `workflowDataFields`; keep `visibleWorkflowFormFields` layout-aware and return a shallow-cloned group with only visible children.

- [ ] **Step 4: Run helper checks and verify GREEN**

Run the commands from Step 2. Expected: PASS.

## Task 4: Designer Group, Display Components, and Help Properties

**Files:**
- Modify: `admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowFieldPermissions.vue`
- Modify: `admin/src/views/workflow/designer/index.vue`
- Modify: `admin/scripts/check-workflow-form-designer.mjs`

- [ ] **Step 1: Add failing designer source checks**

Require snippets for the four palette items, group child drop zone, `moveWorkflowField`, deletion confirmation, help switch, help title/content controls, and flattened permissions:

```js
[formDesignerSource, "type: 'group'", '表单设计器缺少表单组'],
[formDesignerSource, "type: 'label'", '表单设计器缺少标签'],
[formDesignerSource, "type: 'description'", '表单设计器缺少说明'],
[formDesignerSource, "type: 'button'", '表单设计器缺少说明按钮'],
[formDesignerSource, 'group.fields', '表单组缺少子字段画布'],
[permissionSource, 'workflowDataFields', '字段权限未展平分组业务字段'],
[designerIndexSource, 'workflowDataFields', '表单变更未递归维护节点权限'],
```

- [ ] **Step 2: Run designer check and verify RED**

Run `cd admin && npm run check:workflow-form-designer`. Expected: first missing snippet fails.

- [ ] **Step 3: Implement designer behavior**

Add the four layout types to `fieldTypes` and a `布局与说明` palette group. `buildField` must create stable unique keys and these defaults:

```ts
if (type === 'group') return { key, label: '新建分组', type, span: 24, fields: [] }
if (type === 'label') return { key, label: '标签文本', type, span: 24 }
if (type === 'description') return { key, label: '填写提示', type, span: 24, content: '请输入说明内容' }
if (type === 'button') return { key, label: '查看说明', type, span: 24, help: { title: '说明', content: '请输入说明内容' } }
```

Replace numeric drag state with `{ containerKey: string | null; index: number; fieldKey: string }`, call `moveWorkflowField` on drop, and render one-level child cards inside `group.fields`. Group drop targets reject a group source. Change selected-field lookup/removal to pure recursive helpers. Before deleting a non-empty group, call:

```ts
await ElMessageBox.confirm('删除分组将同时删除组内组件，确定继续？', '删除分组', { type: 'warning' })
```

Add property sections for group title, label text, description content, button help, and a help switch for group/data fields. Use plain textareas, no `v-html`.

In `WorkflowFieldPermissions.vue`, iterate `permissionFields = computed(() => workflowDataFields(props.draft.form))`. In designer `index.vue`, use the same helper to preserve/initialize node permissions for grouped data fields only.

- [ ] **Step 4: Run designer check and verify GREEN**

Run `npm run check:workflow-form-designer`. Expected: PASS.

## Task 5: Runtime Recursive Rendering and Shared Help Dialog

**Files:**
- Modify: `admin/src/views/workflow/components/WorkflowRuntimeForm.vue`
- Modify: `admin/src/views/workflow/runtimeForm.ts`
- Modify: `admin/scripts/check-workflow-runtime-form.mjs`
- Modify: `admin/scripts/check-workflow-runtime-pages.mjs`

- [ ] **Step 1: Add failing runtime source checks**

Require group recursion, label/description/button branches, label-slot help button, and plain-text dialog:

```js
[runtimeSource, "field.type === 'group'", '运行时缺少表单组'],
[runtimeSource, "field.type === 'description'", '运行时缺少静态说明'],
[runtimeSource, 'openFieldHelp(field)', '运行时缺少说明入口'],
[runtimeSource, '<el-dialog', '运行时缺少共用说明对话框'],
[runtimeSource, 'white-space: pre-wrap', '说明对话框未保留换行'],
```

Add runtime pure assertions that hidden group children remove an empty group, while display children keep the group visible.

- [ ] **Step 2: Run runtime checks and verify RED**

Run:

```bash
cd admin
npm run check:workflow-runtime-form
npm run check:workflow-runtime-pages
```

Expected: missing group/help snippets or recursive visibility assertions fail.

- [ ] **Step 3: Implement runtime layout and help**

Make `WorkflowRuntimeForm.vue` recursively invoke itself for `field.fields`, passing the same `modelValue`, access map, action map, and readonly state. Use an embedded class so a nested instance renders a grid without another empty state.

Before data controls, branch on layout types:

```vue
<section v-if="field.type === 'group'" class="runtime-form-group">
  <header>
    <strong>{{ field.label }}</strong>
    <el-button v-if="field.help" link type="primary" @click="openFieldHelp(field)">
      {{ field.help.buttonText || '查看说明' }}
    </el-button>
  </header>
  <WorkflowRuntimeForm
    :fields="field.fields || []"
    :model-value="modelValue"
    :field-access="fieldAccess"
    :field-actions="fieldActions"
    :readonly="readonly"
    embedded
    @update:model-value="$emit('update:modelValue', $event)"
  />
</section>
<h3 v-else-if="field.type === 'label'" class="runtime-form-label">{{ field.label }}</h3>
<p v-else-if="field.type === 'description'" class="runtime-form-description">{{ field.content }}</p>
<el-button v-else-if="field.type === 'button'" @click="openFieldHelp(field)">{{ field.label }}</el-button>
```

For data controls, use the `el-form-item` label slot to place the text help button immediately after the field label. Store the active help in a ref and render content with interpolation inside a `white-space: pre-wrap` container. Do not use `v-html`.

Update remote-option collection so recursive component instances load options for their own direct data/detail fields without treating group/display nodes as option fields.

- [ ] **Step 4: Run runtime checks and verify GREEN**

Run commands from Step 2. Expected: PASS.

## Task 6: Full Verification and Review

**Files:**
- Verify all files listed above.

- [ ] **Step 1: Format Go files**

Run:

```bash
cd backend
gofmt -w internal/workflow/types.go internal/workflow/validation.go internal/workflow/form.go internal/workflow/definition_test.go internal/modules/workflow/application/service_test.go internal/modules/workflow/infrastructure/gorm_store.go
```

- [ ] **Step 2: Run focused backend tests**

```bash
cd backend
GOCACHE=$PWD/.cache/go-build go test ./internal/workflow ./internal/service/admin/workflow ./internal/modules/workflow/... -count=1
```

Expected: all packages PASS.

- [ ] **Step 3: Run admin workflow checks**

```bash
cd admin
npm run check:workflow-form-designer
npm run check:workflow-runtime-form
npm run check:workflow-runtime-pages
npm run check:workflow-tree
```

Expected: all checks PASS.

- [ ] **Step 4: Run admin build and bundle check**

```bash
cd admin
npm run build
npm run check:bundle
```

Expected: build and bundle check exit 0; pre-existing dependency PURE-comment warnings may remain non-fatal.

- [ ] **Step 5: Check scoped diff formatting**

Run:

```bash
git diff --check -- \
  backend/internal/workflow/types.go \
  backend/internal/workflow/validation.go \
  backend/internal/workflow/form.go \
  backend/internal/workflow/definition_test.go \
  backend/internal/modules/workflow/application/service_test.go \
  backend/internal/modules/workflow/infrastructure/gorm_store.go \
  admin/src/views/workflow/types.ts \
  admin/src/views/workflow/formLayout.ts \
  admin/src/views/workflow/runtimeForm.ts \
  admin/src/views/workflow/designer/components/WorkflowFormDesigner.vue \
  admin/src/views/workflow/designer/components/WorkflowFieldPermissions.vue \
  admin/src/views/workflow/designer/index.vue \
  admin/src/views/workflow/components/WorkflowRuntimeForm.vue \
  admin/scripts/check-workflow-form-designer.mjs \
  admin/scripts/check-workflow-runtime-form.mjs \
  admin/scripts/check-workflow-runtime-pages.mjs
```

Expected: no whitespace errors.

- [ ] **Step 6: Review scope**

Confirm the diff does not touch existing performance workflow handlers, `h5app`, or unrelated admin modules. Report user-owned baseline failures separately if any broader check is run.

## Execution Note

The implementation files overlap existing uncommitted workflow work. Do not create implementation commits automatically: staging any whole overlapping file could include prior user-owned changes. Use the RED/GREEN commands and scoped diffs above as review checkpoints, then leave final integration to the user.
