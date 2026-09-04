import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const read = path => readFileSync(resolve(root, path), 'utf8')

for (const path of [
  'src/views/scheduled-task/tasks/index.vue',
  'src/views/scheduled-task/runs/index.vue',
  'src/views/scheduled-task/workers/index.vue',
  'src/views/scheduled-task/components/TaskEditorDialog.vue',
  'src/views/workflow/components/WorkflowUserTreePicker.vue',
  'src/views/scheduled-task/handlerLabels.ts',
  'src/views/scheduled-task/types.ts',
  'src/types/scheduledTask.ts',
]) {
  if (!existsSync(resolve(root, path))) throw new Error(`scheduled task UI missing ${path}`)
}

const routes = read('src/router/adminRoutes.ts')
for (const path of ['scheduled-task/tasks', 'scheduled-task/runs', 'scheduled-task/workers']) {
  if (!routes.includes(`path: '${path}'`)) throw new Error(`scheduled task route missing ${path}`)
}

const api = read('src/api/index.ts')
for (const endpoint of [
  '/scheduled-tasks', '/scheduled-tasks/cron-preview', '/scheduled-task-handlers',
  '/scheduled-task-runs', '/scheduled-task-workers',
]) {
  if (!api.includes(endpoint)) throw new Error(`scheduled task API missing ${endpoint}`)
}

const editor = read('src/views/scheduled-task/components/TaskEditorDialog.vue')
for (const handlerType of ['go', 'workflow', 'http', 'shell', 'sql']) {
  if (!editor.includes(`'${handlerType}'`)) throw new Error(`task editor missing ${handlerType} handler`)
}
for (const snippet of ['<el-dialog', 'class="task-editor-dialog"', 'width="min(960px, 94vw)"', ':close-on-click-modal="false"', 'append-to-body', ':global(.task-editor-dialog .el-dialog__body)', 'cronPreview', 'admin:menu:scheduled-task:shell', 'admin:menu:scheduled-task:sql:write', 'WorkflowUserTreePicker', 'workflowPublishedDefinitionList', 'workflowUserOptions', 'workflowDepartmentOptions', 'handlerConfig.starterIds', ':multiple="true"', '请选择已发布流程', 'excludedUserIds', 'goJobOptions', 'selectedGoJob', 'registered-job-meta', 'notification.in_app.send', 'inAppNotificationRecipientOptions', 'notificationParams', 'notificationDepartmentModelValue', 'admin:menu:notification:send']) {
  if (!editor.includes(snippet)) throw new Error(`task editor missing ${snippet}`)
}
if (/workflow:\s*\{\s*definitionId:\s*0\b/.test(editor)) {
  throw new Error('workflow task handler must use an empty value when no definition is selected')
}
for (const snippet of ['workflow: { definitionId: undefined', 'handlerConfig.definitionId = Number.isInteger(definitionID)']) {
  if (!editor.includes(snippet)) throw new Error(`task editor missing workflow definition normalization ${snippet}`)
}
if (editor.includes('<el-drawer')) throw new Error('task editor must use a dialog instead of a drawer')

const workflowUserPicker = read('src/views/workflow/components/WorkflowUserTreePicker.vue')
for (const snippet of [
  'buildUserTree(',
  'Boolean(props.selectDepartmentRules)',
  'disabled: !selectDepartments',
  'collectDescendantUserIDs',
  "if (data.type === 'dept')",
]) {
  if (!workflowUserPicker.includes(snippet)) throw new Error(`workflow user picker missing department selection behavior ${snippet}`)
}

const tasks = read('src/views/scheduled-task/tasks/index.vue')
for (const snippet of ['admin:menu:scheduled-task:add', 'admin:menu:scheduled-task:status', 'admin:menu:scheduled-task:run']) {
  if (!tasks.includes(snippet)) throw new Error(`task list missing permission ${snippet}`)
}

const handlerLabels = read('src/views/scheduled-task/handlerLabels.ts')
for (const label of ['Go 注册任务', '发起流程', 'HTTP / Webhook 请求', '受控 Shell 命令', '受控 SQL 任务']) {
  if (!handlerLabels.includes(label)) throw new Error(`scheduled task handler label missing ${label}`)
}

const scheduledTaskTypes = read('src/types/scheduledTask.ts')
if (!scheduledTaskTypes.includes("'x-enum-labels'?: Record<string, string>")) {
  throw new Error('scheduled task metadata type missing x-enum-labels')
}
for (const [source, label] of [[editor, 'task editor'], [tasks, 'task list']]) {
  if (!source.includes('handlerTypeLabel(item.type)')) throw new Error(`${label} must display localized handler labels`)
}

const runs = read('src/views/scheduled-task/runs/index.vue')
for (const snippet of ['append-to-body', 'admin:menu:scheduled-task:run:retry', 'admin:menu:scheduled-task:run:cancel']) {
  if (!runs.includes(snippet)) throw new Error(`run list missing ${snippet}`)
}
