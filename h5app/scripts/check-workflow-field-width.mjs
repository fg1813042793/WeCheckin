import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import ts from 'typescript'

const path = new URL('../src/pages/workflow/workflow-field-width.ts', import.meta.url)
assert.ok(existsSync(path), 'H5 必须识别新的流程列宽')
const source = ts.transpileModule(readFileSync(path, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const module = { exports: {} }
// eslint-disable-next-line no-new-func
new Function('module', 'exports', source)(module, module.exports)
const { workflowFieldSpanStyle } = module.exports
for (let span = 1; span <= 24; span++)
  assert.equal(workflowFieldSpanStyle(JSON.parse(JSON.stringify({ span }))).width, `${span * 100 / 24}%`)
assert.equal(workflowFieldSpanStyle({ span: 4.8 }).width, '20%')
for (const span of [undefined, 0, -1, 4.7, 4.9, 24.1, 25, Number.NaN])
  assert.equal(workflowFieldSpanStyle({ span }).width, '100%')
assert.equal([18, 3, 3].reduce((sum, span) => sum + Number.parseFloat(workflowFieldSpanStyle({ span }).width), 0), 100)
console.log('H5 workflow widths passed: slider steps 1-24, mixed row and legacy fifth')
