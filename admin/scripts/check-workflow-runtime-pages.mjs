import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')

function read(path) {
  return readFileSync(resolve(srcDir, path), 'utf8')
}

function requireSnippet(source, snippet, label) {
  if (!source.includes(snippet)) {
    throw new Error(`${label} missing required snippet: ${snippet}`)
  }
}

const instancePagePath = resolve(srcDir, 'views/workflow/instances/index.vue')
const taskPagePath = resolve(srcDir, 'views/workflow/tasks/index.vue')
const designerPagePath = resolve(srcDir, 'views/workflow/designer/index.vue')
const definitionPagePath = resolve(srcDir, 'views/workflow/index.vue')
const publishDialogPath = resolve(srcDir, 'views/workflow/components/WorkflowPublishDialog.vue')
const logoPickerPath = resolve(srcDir, 'views/workflow/components/WorkflowLogoPicker.vue')
const userTreePickerPath = resolve(srcDir, 'views/workflow/components/WorkflowUserTreePicker.vue')
const flowConfigPath = resolve(srcDir, 'views/workflow/designer/components/WorkflowStartConfig.vue')
const formPreviewDialogPath = resolve(srcDir, 'views/workflow/designer/components/WorkflowFormPreviewDialog.vue')

if (!existsSync(instancePagePath)) {
  throw new Error('workflow instance page missing: src/views/workflow/instances/index.vue')
}
if (!existsSync(taskPagePath)) {
  throw new Error('workflow task page missing: src/views/workflow/tasks/index.vue')
}
if (!existsSync(designerPagePath)) {
  throw new Error('workflow designer page missing: src/views/workflow/designer/index.vue')
}
if (!existsSync(definitionPagePath)) {
  throw new Error('workflow definition page missing: src/views/workflow/index.vue')
}
if (!existsSync(publishDialogPath)) {
  throw new Error('workflow publish dialog missing: src/views/workflow/components/WorkflowPublishDialog.vue')
}
if (!existsSync(logoPickerPath)) {
  throw new Error('workflow logo picker missing: src/views/workflow/components/WorkflowLogoPicker.vue')
}
if (!existsSync(userTreePickerPath)) {
  throw new Error('workflow user tree picker missing: src/views/workflow/components/WorkflowUserTreePicker.vue')
}
if (!existsSync(flowConfigPath)) {
  throw new Error('workflow start config missing: src/views/workflow/designer/components/WorkflowStartConfig.vue')
}
if (!existsSync(formPreviewDialogPath)) {
  throw new Error('workflow form preview dialog missing: src/views/workflow/designer/components/WorkflowFormPreviewDialog.vue')
}

const api = read('api/index.ts')
const workflowTypes = read('views/workflow/types.ts')
const routes = read('router/adminRoutes.ts')
const instances = read('views/workflow/instances/index.vue')
const tasks = read('views/workflow/tasks/index.vue')
const designer = read('views/workflow/designer/index.vue')
const definitions = read('views/workflow/index.vue')
const publishDialog = read('views/workflow/components/WorkflowPublishDialog.vue')
const logoPicker = read('views/workflow/components/WorkflowLogoPicker.vue')
const userTreePicker = read('views/workflow/components/WorkflowUserTreePicker.vue')
const flowConfig = read('views/workflow/designer/components/WorkflowStartConfig.vue')
const formPreviewDialog = read('views/workflow/designer/components/WorkflowFormPreviewDialog.vue')

requireSnippet(designer, "if (selectedNode.value.type === 'start') return '开始节点'", 'workflow start node inspector title')

for (const snippet of [
  '<WorkflowFormPreviewDialog',
  ':draft="detail.draft"',
  'icon="View"',
  '@click="openFormPreview"',
]) {
  requireSnippet(designer, snippet, 'workflow designer form preview')
}

for (const snippet of [
  '<WorkflowRuntimeForm',
  'initialWorkflowFormData',
  'workflowFieldAccessMap',
  'workflowFieldActionMap',
  ':field-access="previewFieldAccess"',
  ':field-actions="previewFieldActions"',
  "previewMode === 'mobile'",
]) {
  requireSnippet(formPreviewDialog, snippet, 'workflow form preview dialog')
}

const designerDrawers = designer.match(/<el-drawer\b[^>]*>/gs) || []
assert.ok(designerDrawers.length > 0, 'workflow designer must include at least one drawer')
assert.equal(
  designerDrawers.filter(drawer => drawer.includes('append-to-body')).length,
  designerDrawers.length,
  'workflow designer drawers must append to body so the fixed admin header cannot cover them',
)

for (const snippet of [
  'workflowInstanceList(',
  'workflowInstanceStart(',
  'workflowInstanceDetail(',
  'workflowInstanceResume(',
  'workflowTaskList(',
  'workflowTaskComplete(',
  'workflowPublishedDefinitionList(',
  'workflowPublishedDefinitionDetail(',
  'workflowUserOptions(',
  'workflowDepartmentOptions(',
  'workflowNotificationList(',
  'workflowNotificationRetry(',
  'workflowNotificationDispatchDue(',
]) {
  requireSnippet(api, snippet, 'workflow runtime API')
}

for (const snippet of [
  "path: 'workflow/instances'",
  "name: 'WorkflowInstances'",
  "views/workflow/instances/index.vue",
  "path: 'workflow/tasks'",
  "name: 'WorkflowTasks'",
  "views/workflow/tasks/index.vue",
]) {
  requireSnippet(routes, snippet, 'workflow runtime route')
}

for (const snippet of [
  "hasPerm('admin:menu:workflow:instance:start')",
  "hasPerm('admin:menu:workflow:instance:detail')",
  'WorkflowRuntimeForm',
  'workflowFieldActionMap',
  'selectedStartDefinition',
  'startFormData',
  'startFieldActions',
  ':field-actions="startFieldActions"',
  'ref="startRuntimeForm"',
  'startRuntimeForm.value.validate()',
  'workflowPublishedDefinitionList',
  'workflowPublishedDefinitionDetail',
  'workflowUserOptions',
  'workflowDepartmentOptions',
  'WorkflowUserTreePicker',
  'formData: writableWorkflowFormData',
  'detail.formData',
  '流程表单',
  'workflowInstanceList',
  'workflowInstanceStart',
  'startMode',
  '实例创建方式',
  '代一名用户发起',
  '为多名用户分别发起',
  '业务发起人',
  'starterUserIds',
  'starterId:',
  'targetUserId',
  'buildTargetBusinessKey',
  'Promise.allSettled',
  ':disabled="!canStartDefinition(definition)"',
  'availabilityStatusLabel',
  ':user:',
  'failedUserIds',
  'workflowInstanceResume',
  'workflowInstanceDetail',
  "hasPerm('admin:menu:workflow:notification:list')",
  "hasPerm('admin:menu:workflow:notification:retry')",
  'workflowNotificationList',
  'workflowNotificationRetry',
  'workflowNotificationDispatchDue',
  'notificationStatusOptions',
  "'pending'",
  "'sending'",
  "'sent'",
  "'failed'",
  "'dead'",
  '通知投递',
  '投递到期通知',
  '重试',
  "node_cc: '已记录抄送'",
  "node_notify: '通知节点已触发'",
  'class="admin-pagination"',
]) {
  requireSnippet(instances, snippet, 'workflow instance page')
}

assert.ok(!/\bdefinitionId:\s*0\b/.test(instances), 'workflow start form must use an empty value when no definition is selected')
requireSnippet(instances, 'definitionId: undefined', 'workflow start form empty definition value')

for (const snippet of [
  "hasPerm('admin:menu:workflow:task:complete')",
  'WorkflowRuntimeForm',
  'workflowFieldActionMap',
  'activeInstanceDetail',
  'completeFormData',
  'activeTaskFieldActions',
  ':field-actions="activeTaskFieldActions"',
  'ref="completeRuntimeForm"',
  'completeRuntimeForm.value.validate()',
  'writableWorkflowFormData',
  'formData: writableWorkflowFormData',
  'workflowTaskList',
  'workflowTaskComplete',
  'approve',
  'reject',
  'submit',
  'activeTaskNodeType',
  'class="admin-pagination"',
]) {
  requireSnippet(tasks, snippet, 'workflow task page')
}

for (const snippet of [
  '@click="openEdit(row)"',
  'title="修改流程信息"',
  'v-model="editForm.name"',
  'v-model="editForm.category"',
  'v-model="editForm.description"',
  ':model-value="editTarget?.key"',
  'workflowDefinitionUpdate',
  'description: editForm.description.trim()',
  'category: editForm.category.trim()',
  '流程信息修改成功',
  '<WorkflowLogoPicker',
  'row.logoUrl',
  'buildWorkflowDefinitionFormData',
  "formData.append('logo', logoFile)",
  "formData.append('removeLogo', 'true')",
]) {
  requireSnippet(definitions, snippet, 'workflow definition metadata edit')
}

for (const snippet of ['accept="image/png,image/jpeg,image/webp"', 'icon="Upload"', 'icon="Delete"', 'workflow-logo-picker__preview']) {
  requireSnippet(logoPicker, snippet, 'workflow logo picker')
}

for (const snippet of [
  'workflow-user-picker__selection',
  'visibleSelectedUsers',
  'hiddenSelectedUsers',
  'workflow-user-picker__overflow',
  '<el-tooltip',
  'ResizeObserver',
]) {
  requireSnippet(userTreePicker, snippet, 'workflow user tree picker compact selection')
}
assert.ok(
  !userTreePicker.includes('workflow-user-picker__selected'),
  'workflow user tree picker must not render selected users below the input',
)

requireSnippet(workflowTypes, 'logoUrl: string', 'workflow definition logo type')

for (const snippet of [
  'workflowDefinitionPublish(id: ID, data:',
  'initiator?: WorkflowInitiatorConfig',
  'departmentIds?: number[]',
]) {
  requireSnippet(api, snippet, 'workflow publish API')
}

requireSnippet(workflowTypes, 'departmentIds?: number[]', 'workflow initiator type')
requireSnippet(workflowTypes, 'export interface WorkflowStartAvailabilityConfig', 'workflow start availability type')
requireSnippet(workflowTypes, 'availability?: WorkflowStartAvailabilityConfig', 'workflow start availability type')

for (const snippet of [
  '<el-tab-pane label="流程配置" name="config" />',
  '<WorkflowStartConfig',
  "activeDesignerTab === 'config'",
  "error.code === 'initiator_invalid'",
  "error.code === 'start_availability_invalid'",
]) {
  requireSnippet(designer, snippet, 'workflow designer flow config tab')
}

for (const snippet of [
  '允许发起范围',
  '允许发起时间',
  '全部用户',
  '指定范围',
  '允许发起部门',
  '额外允许用户',
  '<el-tree-select',
  'multiple',
  'check-strictly',
  '长期有效',
  '指定时间段',
  '每周周期开放',
  '每月周期开放',
  'Asia/Shanghai',
  'lastDayOfMonth',
  '最后一天',
  '.config-section {',
  'border-radius: 8px',
  'background: #fff',
  'box-shadow: 0 2px 10px',
  'grid-template-columns: repeat(2, minmax(0, 1fr))',
  'width: min(1040px, 100%)',
  'justify-content: start',
  '@media (max-width: 1100px)',
]) {
  requireSnippet(flowConfig, snippet, 'workflow start config')
}

for (const snippet of [
  'title="发布流程"',
  '发布版本',
  '流程配置将随本次版本一起发布',
  'workflowDefinitionPublish(publishTarget.value.id)',
]) {
  requireSnippet(publishDialog, snippet, 'workflow publish initiator dialog')
}

for (const forbidden of ['WorkflowUserTreePicker', '允许发起部门', '额外允许用户', '<el-tree-select']) {
  assert.ok(!publishDialog.includes(forbidden), `workflow publish dialog must not duplicate flow config: ${forbidden}`)
}

for (const [source, label] of [
  [definitions, 'workflow definition page'],
  [designer, 'workflow designer page'],
]) {
  requireSnippet(source, '<WorkflowPublishDialog', label)
  requireSnippet(source, '@published=', label)
}
