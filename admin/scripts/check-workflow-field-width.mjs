import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const path = new URL('../src/views/workflow/fieldWidth.ts', import.meta.url)
assert.ok(existsSync(path), '流程布局需要统一列宽契约，支持 24 级滑动宽度')
const widths = executeTypeScriptModule(readFileSync(path, 'utf8'), 'fieldWidth.ts')
for (let span = 1; span <= 24; span++) {
  const field = JSON.parse(JSON.stringify({ key: 'field', span }))
  assert.equal(widths.workflowFieldSpan(field), span, `${span}/24 save/reload`)
  assert.equal(widths.workflowFieldGridSpan(field), span * 5, `${span}/24 preview/runtime grid`)
}
assert.equal(widths.workflowFieldSpan({ span: 4.8 }), 4.8, 'historical exact fifth remains readable')
assert.equal(widths.workflowFieldGridSpan({ span: 4.8 }), 24)
assert.equal(widths.workflowFieldSpanLabel(18), '18/24 · 3/4')
assert.equal(widths.workflowFieldSpanLabel(24), '24/24 · 整行')
assert.equal([18, 3, 3].reduce((sum, span) => sum + widths.workflowFieldGridSpan({ span }), 0), 120)
for (const span of [undefined, 0, null, -1, 4.7, 4.9, 24.1, 25, NaN, Infinity]) {
  assert.equal(widths.workflowFieldSpan({ span }), 24, '历史默认值和非法宽度回退整行')
}
console.log('workflow field widths passed: slider steps 1-24, 6/8+1/8+1/8, JSON reload and legacy fifth')
