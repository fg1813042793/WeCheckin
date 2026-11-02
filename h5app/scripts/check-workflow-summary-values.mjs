import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const ts = require('typescript')

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../src/pages/workflow')
const cache = new Map()
function load(name) {
  const path = resolve(root, `${name}.ts`)
  if (cache.has(path))
    return cache.get(path)
  const source = ts.transpileModule(readFileSync(path, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
  const module = { exports: {} }
  cache.set(path, module.exports)
  runInThisContext(`(function(module,exports,require){${source}\n})`)(module, module.exports, request => request.startsWith('.') ? load(request) : require(request))
  return module.exports
}
const helpers = load('workflow-summary-columns')
const grade = { key: 'final', label: '最终分档', type: 'calculation', calculation: { expression: '([leader] + [hrbp]) / 2', precision: 2 } }
const options = [{ value: 'A', label: '优秀', calculationValue: 1 }, { value: 'B', label: '良好', calculationValue: 3 }]
const form = [{ key: 'group', type: 'group', fields: [{ key: 'leader', label: '上级分档', type: 'select', options }, { key: 'hrbp', label: 'HRBP分档', type: 'select', options }] }, grade]
const column = helpers.workflowSummaryFormFieldGroups([{ id: 1, name: '考评单', form }])[0].columns.find(item => item.field.key === 'final')
const legacy = { id: 'old', definitionId: 1, definitionVersion: 16, formData: { leader: 'A', hrbp: 'B' } }
const before = JSON.stringify(legacy)
assert.equal(helpers.workflowSummaryFormValue(legacy, column), '2.00', 'historic records without a saved calculation must use available source values')
assert.equal(JSON.stringify(legacy), before, 'summary computation must not change original data')
assert.equal(helpers.workflowSummaryFormValueInfo(legacy, column).calculated, true)
const persisted = { ...legacy, formData: { ...legacy.formData, final: 4 } }
assert.equal(helpers.workflowSummaryFormValue(persisted, column), '4.00', 'saved historical results take priority')
assert.equal(helpers.workflowSummaryFormValueInfo(persisted, column).calculated, false)
assert.equal(helpers.workflowSummaryFormValue({ ...legacy, formData: { final: 0 } }, column), '0.00')
for (const data of [{}, { leader: 'A' }, { leader: null, hrbp: 'B' }, { leader: '', hrbp: 'B' }, { leader: '   ', hrbp: 'B' }, { leader: 'invalid', hrbp: 'B' }])
  assert.equal(helpers.workflowSummaryFormValue({ ...legacy, formData: data }, column), '-', 'missing/invalid inputs cannot silently count as zero')
assert.equal(helpers.workflowSummaryFormValue({ ...legacy, definitionId: 2 }, column), '-', 'another definition cannot borrow this schema')
const labelled = { ...column, field: { ...grade, calculation: { ...grade.calculation, resultDisplay: { mode: 'label', rules: [{ min: 1, max: 2, label: '达标' }] } } } }
assert.equal(helpers.workflowSummaryFormValue(legacy, labelled), '达标')
const gradeOptions = [
  { value: '1.1', label: 'C+', calculationValue: 1.1 },
  { value: '1.0', label: 'C', calculationValue: 1 },
  { value: '0.9', label: 'C-', calculationValue: 0.9 },
]
const legacyLabelForm = [
  { key: 'manager', type: 'select', options: gradeOptions },
  { key: 'hrbp', type: 'select', options: gradeOptions },
  { key: 'final', type: 'calculation', calculation: { expression: '([manager] + [hrbp]) / 2', precision: 2 } },
]
const legacyLabelColumn = helpers.workflowSummaryFormFieldGroups([{ id: 1, form: legacyLabelForm }])[0].columns.at(-1)
assert.equal(
  helpers.workflowSummaryFormValue({ ...legacy, formData: { manager: 'C', hrbp: 'C' } }, legacyLabelColumn),
  '1.00',
  'summary backfill must resolve a unique historical option label against the current numeric option value',
)
const ambiguousLegacyLabelForm = legacyLabelForm.map(field => field.type === 'select'
  ? { ...field, options: [...gradeOptions, { value: '2.0', label: 'C', calculationValue: 2 }] }
  : field)
const ambiguousLegacyLabelColumn = helpers.workflowSummaryFormFieldGroups([{ id: 1, form: ambiguousLegacyLabelForm }])[0].columns.at(-1)
assert.equal(
  helpers.workflowSummaryFormValue({ ...legacy, formData: { manager: 'C', hrbp: 'C' } }, ambiguousLegacyLabelColumn),
  '-',
  'duplicate labels must stay unresolved instead of choosing an arbitrary option',
)
const numericForm = [{ key: 'a', type: 'number' }, { key: 'b', type: 'number' }, { ...grade, calculation: { expression: '[a] / [b]', precision: 1 } }]
const numericColumn = helpers.workflowSummaryFormFieldGroups([{ id: 1, form: numericForm }])[0].columns.at(-1)
assert.equal(helpers.workflowSummaryFormValue({ ...legacy, formData: { a: 0, b: 2 } }, numericColumn), '0.0')
assert.equal(helpers.workflowSummaryFormValue({ ...legacy, formData: { a: 1, b: 0 } }, numericColumn), '-')
assert.equal(helpers.workflowSummaryFormValue({ ...legacy, formData: { a: true, b: 2 } }, numericColumn), '-')
assert.equal(helpers.workflowSummaryFormValue({ ...legacy, formData: { a: [], b: 2 } }, numericColumn), '-')
const detailForm = [{ key: 'items', type: 'detail_list', columns: [{ key: 'score', type: 'number' }] }, { ...grade, calculation: { expression: 'SUM([items.score])', precision: 0 } }]
const detailColumn = helpers.workflowSummaryFormFieldGroups([{ id: 1, form: detailForm }])[0].columns.at(-1)
assert.equal(helpers.workflowSummaryFormValue({ ...legacy, formData: { items: [{ score: 3 }, { score: 0 }] } }, detailColumn), '3')
assert.equal(helpers.workflowSummaryFormValue({ ...legacy, formData: { items: [{ score: 3 }, {}] } }, detailColumn), '-')
assert.equal(helpers.workflowSummaryFormValue({ ...legacy, formData: { items: [] } }, detailColumn), '0')
const ordinary = { ...column, field: { key: 'missing', label: '备注', type: 'text' } }
assert.equal(helpers.workflowSummaryFormValue(legacy, ordinary), '-')
const calculation = load('workflow-calculation')
assert.equal(calculation.evaluateWorkflowCalculation(grade, { leader: 'A' }, form).value, 0.5, 'ordinary form evaluation keeps its existing empty-input behavior')
console.log('workflow summary values: historical calculation, saved results, labels, complete inputs, isolation and immutable display passed')
