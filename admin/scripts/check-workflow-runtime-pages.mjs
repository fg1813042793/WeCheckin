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

if (!existsSync(instancePagePath)) {
  throw new Error('workflow instance page missing: src/views/workflow/instances/index.vue')
}
if (!existsSync(taskPagePath)) {
  throw new Error('workflow task page missing: src/views/workflow/tasks/index.vue')
}
if (!existsSync(designerPagePath)) {
  throw new Error('workflow designer page missing: src/views/workflow/designer/index.vue')
}

const api = read('api/index.ts')
const routes = read('router/adminRoutes.ts')
const instances = read('views/workflow/instances/index.vue')
const tasks = read('views/workflow/tasks/index.vue')
const designer = read('views/workflow/designer/index.vue')

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
  'workflowTaskList(',
  'workflowTaskComplete(',
  'workflowPublishedDefinitionList(',
  'workflowPublishedDefinitionDetail(',
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
  'selectedStartDefinition',
  'startFormData',
  'workflowPublishedDefinitionList',
  'workflowPublishedDefinitionDetail',
  'formData: startFormData.value',
  'detail.formData',
  '流程表单',
  'workflowInstanceList',
  'workflowInstanceStart',
  'workflowInstanceDetail',
  'class="admin-pagination"',
]) {
  requireSnippet(instances, snippet, 'workflow instance page')
}

for (const snippet of [
  "hasPerm('admin:menu:workflow:task:complete')",
  'WorkflowRuntimeForm',
  'activeInstanceDetail',
  'completeFormData',
  'writableWorkflowFormData',
  'formData: writableWorkflowFormData',
  'workflowTaskList',
  'workflowTaskComplete',
  'approve',
  'reject',
  'class="admin-pagination"',
]) {
  requireSnippet(tasks, snippet, 'workflow task page')
}
