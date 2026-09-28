import assert from 'node:assert/strict'
import { Buffer } from 'node:buffer'
import { existsSync, readFileSync } from 'node:fs'
import ts from 'typescript'

const helperURL = new URL('../src/pages/workflow/workflow-history-split.ts', import.meta.url)
const component = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowDetailPanel.vue', import.meta.url),
  'utf8',
)

assert.ok(existsSync(helperURL), 'workflow history split helper is required')

const helperSource = readFileSync(helperURL, 'utf8')
const compiled = ts.transpileModule(helperSource, {
  compilerOptions: {
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ES2022,
  },
}).outputText
const helper = await import(`data:text/javascript;base64,${Buffer.from(compiled).toString('base64')}`)

assert.equal(helper.clampWorkflowHistorySplit(10), 25)
assert.equal(helper.clampWorkflowHistorySplit(50), 50)
assert.equal(helper.clampWorkflowHistorySplit(90), 75)
assert.equal(helper.workflowHistorySplitFromPointer(500, 100, 800), 50)
assert.equal(helper.workflowHistorySplitFromKey(50, 'ArrowUp', false), 46)
assert.equal(helper.workflowHistorySplitFromKey(50, 'ArrowDown', true), 60)
assert.equal(helper.workflowHistorySplitFromKey(50, 'Home', false), 25)
assert.equal(helper.workflowHistorySplitFromKey(50, 'End', false), 75)
assert.equal(helper.workflowHistorySplitFromKey(50, 'Enter', false), null)

for (const snippet of [
  'const historySplitPercent = ref(WORKFLOW_HISTORY_SPLIT_DEFAULT)',
  'const historyResizeEnabled = computed(() => historyDrawer.value && !historyDialog.value)',
  'class="workflow-detail-panel__history-resizer"',
  'role="separator"',
  'aria-orientation="horizontal"',
  ':aria-valuenow="Math.round(historySplitPercent)"',
  '@mousedown.prevent="beginHistoryResize"',
  '@dblclick="resetHistorySplit"',
  '@keydown="handleHistoryResizeKey"',
  'workflow-detail-panel__history-layout--resizable',
  'cursor: row-resize;',
]) {
  assert.ok(component.includes(snippet), `WorkflowDetailPanel.vue missing ${snippet}`)
}

console.log('Workflow history resizer checks passed')
