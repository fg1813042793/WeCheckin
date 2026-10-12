import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/pages/performance/components/PerformanceWorkbench.vue', import.meta.url),
  'utf8',
)
const workflowTypes = readFileSync(
  new URL('../src/types/workflow.ts', import.meta.url),
  'utf8',
)

assert.match(
  source,
  /\.todo-item\s*\{[^}]*grid-template-columns:\s*48px minmax\(0, 1fr\) auto;/,
  'desktop todo cards must keep their three-column layout',
)

const mobileStyles = source.match(/@media \(max-width: 768px\) \{([\s\S]*)\}\s*<\/style>/)?.[1] || ''

assert.match(
  mobileStyles,
  /\.todo-item\s*\{[^}]*grid-template-columns:\s*72rpx minmax\(0, 1fr\) auto;/,
  'mobile todo cards must keep the action in a dedicated third column',
)
assert.match(
  mobileStyles,
  /\.todo-item__side\s*\{[^}]*grid-column:\s*3;[^}]*grid-row:\s*1;/,
  'mobile todo actions must stay on the first row',
)
assert.match(
  mobileStyles,
  /\.todo-item__action-button,[\s\S]*?:deep\(\.todo-item__action-button\)\s*\{[^}]*white-space:\s*nowrap;/,
  'mobile todo action labels must not wrap',
)

for (const snippet of [
  'workflowFormRevisionDetailContentKey',
  'workflowTaskContentKey',
  'workflowTaskTabTitle',
  'function openWorkflowTodo(task: WorkflowTaskSummary)',
  'appContent.openDynamicTab({',
  'path: `/pages/index/index?view=$' + '{encodeURIComponent(key)}`',
  'task.taskType === \'form_revision\'',
]) {
  assert.ok(source.includes(snippet), `workbench direct workflow task navigation missing ${snippet}`)
}

assert.equal(
  source.match(/@click(?:\.stop)?="openWorkflowTodo\(task\)"/g)?.length || 0,
  2,
  'workflow todo card and action button must both open the task page directly',
)
assert.ok(!source.includes('function openWorkflowTodos()'), 'workbench must not route workflow todo actions through the pending list')
assert.ok(workflowTypes.includes('instanceTitle: string'), 'workflow task contract must expose the instance title')
assert.match(
  source,
  /function workflowTodoCardTitle\(task: WorkflowTaskSummary\) \{\s*return firstText\(task\.instanceTitle, workflowTodoTitle\(task\)\)\s*\}/,
  'workbench workflow cards must prefer the instance title and keep the existing task title as fallback',
)
assert.match(
  source,
  /function workflowTodoTitle\(task: WorkflowTaskSummary\) \{\s*return firstText\(task\.definitionName, task\.nodeName, '流程审批'\)\s*\}/,
  'workflow task tab titles must retain their existing definition-name behavior',
)
assert.ok(source.includes('{{ workflowTodoCardTitle(task) }}'), 'workflow todo cards must render the card-specific title')

console.log('Workbench mobile todo layout checks passed')
