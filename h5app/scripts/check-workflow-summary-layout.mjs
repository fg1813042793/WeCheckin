import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowSummarySection.vue', import.meta.url),
  'utf8',
)

assert.match(source, /workflow-summary__filter--status/)
assert.match(source, /workflow-summary__filter--version/)
assert.match(source, /workflow-summary__filter--start/)
assert.match(source, /workflow-summary__filter--end/)
assert.match(source, /grid-template-columns:\s*repeat\(12,\s*minmax\(0,\s*1fr\)\)/)
assert.match(source, /\.workflow-summary__filter-actions\s*\{[\s\S]*?grid-column:\s*7\s*\/\s*13;[\s\S]*?grid-row:\s*1/)
assert.match(source, /\.workflow-summary__filter--start\s*\{[\s\S]*?grid-column:\s*1\s*\/\s*7;[\s\S]*?grid-row:\s*2/)
assert.match(source, /:deep\(\.workflow-filter-panel\)[\s\S]*?margin-bottom:\s*0/)
assert.match(source, /:deep\(\.workflow-summary__table-scroll \.u-checkbox-group\)[\s\S]*?width:\s*100%/)
assert.match(source, /\.workflow-summary__table\s*\{[\s\S]*?width:\s*100%/)

console.log('workflow summary layout checks passed')
