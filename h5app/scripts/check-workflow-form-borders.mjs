import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowRuntimeForm.vue', import.meta.url),
  'utf8',
)
const layoutStyle = readFileSync(new URL('../src/common/ui-layout.scss', import.meta.url), 'utf8')

assert.match(
  source,
  /\.workflow-detail\s*\{[^}]*padding:\s*0;[^}]*border:\s*0;[^}]*background:\s*transparent;/,
  'detail list must not add another outer frame',
)
assert.match(
  source,
  /\.workflow-detail__row\s*\{[^}]*border:\s*1px solid \$u-border-color;/,
  'each detail row must retain one clear boundary',
)
assert.match(
  source,
  /\.workflow-form--plain-readonly\s*\{[\s\S]*?:deep\(\.u-input--disabled\)[\s\S]*?border-color:\s*#e5eaf3 !important;[\s\S]*?background-color:\s*#fafbfc !important;[\s\S]*?box-shadow:\s*none !important;/,
  'plain readonly controls must keep one subtle input boundary',
)
assert.match(
  source,
  /\.workflow-form--plain-readonly\s*\{[\s\S]*?:deep\(\.u-textarea__count\)[\s\S]*?display:\s*none !important;/,
  'plain readonly textareas must hide character counters',
)
assert.match(
  layoutStyle,
  /\.app-workflow-form \.workflow-detail\)[^{]*\{[^}]*padding:\s*0 !important;[^}]*border:\s*0 !important;/,
  'compact PC styles must keep the detail list outer frame removed',
)

console.log('Workflow form border checks passed')
