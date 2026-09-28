import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowRuntimeForm.vue', import.meta.url),
  'utf8',
)

for (const snippet of [
  '\'workflow-form--readonly\': readonly',
  '.workflow-form:not(.workflow-form--embedded):not(.workflow-form--readonly) {',
  'background: #f6f8fb;',
  '.workflow-form--embedded {',
  'background: transparent;',
  'background: #f1f4f8;',
  '.workflow-form:not(.workflow-form--readonly) {',
  'background-color: #fff !important;',
]) {
  assert.ok(source.includes(snippet), `WorkflowRuntimeForm.vue missing ${snippet}`)
}

assert.match(
  source,
  /\.workflow-form:not\(\.workflow-form--embedded\):not\(\.workflow-form--readonly\)\s*\{[^}]*padding:\s*20rpx;[^}]*border-radius:\s*10rpx;[^}]*background:\s*#f6f8fb;[^}]*box-sizing:\s*border-box;/,
  'editable top-level workflow forms must use the subtle canvas background',
)
assert.doesNotMatch(
  source,
  /\.workflow-form:not\(\.workflow-form--embedded\):not\(\.workflow-form--readonly\)\s*\{[^}]*border:/,
  'editable workflow canvas must not add another border',
)
assert.match(
  source,
  /\.workflow-form__group\s*\{[^}]*background:\s*#f1f4f8;/,
  'workflow groups must be visually distinct from the top-level canvas',
)

console.log('Workflow form canvas checks passed')
