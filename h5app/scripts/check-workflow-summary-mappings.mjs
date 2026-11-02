import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, resolve } from 'node:path'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const ts = require('typescript')

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../src/pages/workflow')
const cache = new Map()

// Load the actual calculation/formatting helpers, including their option schemas.
function load(name) {
  const path = resolve(root, `${name}.ts`)
  if (cache.has(path))
    return cache.get(path)
  const source = ts.transpileModule(readFileSync(path, 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
    fileName: path,
  }).outputText
  const module = { exports: {} }
  cache.set(path, module.exports)
  runInThisContext(`(function(module,exports,require){${source}\n})`)(module, module.exports, request => request.startsWith('.') ? load(request) : require(request))
  return module.exports
}

const {
  normalizeSummaryMappedColumns,
  validateSummaryMappedColumn,
  summaryMappedColumns,
  summaryMappedValue,
  summaryMappedTypeMismatch,
} = load('workflow-summary-mappings')
const { workflowSummaryFormFieldGroups } = load('workflow-summary-columns')

function freeze(value) {
  if (value && typeof value === 'object') {
    Object.values(value).forEach(freeze)
    Object.freeze(value)
  }
  return value
}

function mapped(sources, overrides = {}) {
  return { id: 'rating', label: '统一结果', enabled: true, sources, ...overrides }
}

function source(definitionId, fieldKey, version) {
  return { definitionId, fieldKey, ...(version === undefined ? {} : { version }) }
}

function group(definitionId, form, version = 2) {
  return { ...workflowSummaryFormFieldGroups([{ id: definitionId, name: '同名流程', form, version }])[0], version }
}

function row(definitionId, formData = {}, definitionVersion = 2) {
  return { id: `row-${definitionId}`, definitionId, definitionVersion, definitionName: '旧流程名称', formData }
}

function absent(reason) {
  return { text: '-', calculated: false, reason }
}

const grade = (key, expression, precision = 1) => ({ key, label: '最终得分', type: 'calculation', calculation: { expression, precision } })
const groups = freeze([
  group(11, [
    { key: 'inputs', type: 'group', fields: [{ key: 'a', label: '输入', type: 'number' }] },
    grade('first_score', '[a] * 2'),
    { key: 'note', label: '备注', type: 'text' },
    { key: 'flag', label: '确认', type: 'boolean' },
    { key: 'zero', label: '数字', type: 'number' },
  ]),
  group(22, [{ key: 'b', label: '输入', type: 'amount' }, grade('second_score', '[b] + 100', 2)]),
])
const config = freeze(mapped([source(11, 'first_score'), source(22, 'second_score')]))

test('same labels and different keys resolve only by definition identity', () => {
  assert.deepEqual(summaryMappedValue(row(11, { a: 3, b: 8 }), config, groups), { text: '6.0', calculated: true })
  assert.deepEqual(summaryMappedValue(row(22, { a: 3, b: 8 }), config, groups), { text: '108.00', calculated: true })
  assert.deepEqual(summaryMappedValue(row(33, { first_score: 9 }), config, groups), absent('unmapped'))
})

test('renames preserve mapping and no label or foreign-definition fallback exists', () => {
  const renamed = groups.map(item => ({ ...item, label: '已重命名', columns: item.columns.map(column => ({ ...column, label: '已重命名', field: { ...column.field, label: '已重命名' } })) }))
  const renamedConfig = { ...config, sources: config.sources.map(item => ({ ...item, fieldLabel: '过期显示名称' })) }
  assert.deepEqual(summaryMappedValue(row(11, { a: 3 }), renamedConfig, renamed), { text: '6.0', calculated: true })
  assert.deepEqual(summaryMappedValue(row(11, { second_score: 42 }), mapped([source(11, 'second_score')]), groups), absent('fieldUnavailable'))
  assert.deepEqual(summaryMappedValue(row(11), mapped([source(11, '最终得分')]), groups), absent('fieldUnavailable'))
})

test('each formula uses its own option values and schema', () => {
  const schemas = [11, 22].map(id => group(id, [
    { key: 'choice', type: 'select', options: [{ value: 'A', label: '一样', calculationValue: id === 11 ? 3 : 7 }] },
    grade('result', '[choice] * 2'),
  ]))
  const shared = mapped([source(11, 'result'), source(22, 'result')])
  assert.deepEqual(summaryMappedValue(row(11, { choice: 'A' }), shared, schemas), { text: '6.0', calculated: true })
  assert.deepEqual(summaryMappedValue(row(22, { choice: 'A' }), shared, schemas), { text: '14.0', calculated: true })
})

test('saved results including numeric zero win over recalculation', () => {
  assert.deepEqual(summaryMappedValue(row(11, { a: 3, first_score: 90 }), config, groups), { text: '90.0', calculated: false })
  assert.deepEqual(summaryMappedValue(row(11, { first_score: 0 }), config, groups), { text: '0.0', calculated: false })
  assert.deepEqual(summaryMappedValue(row(22, { second_score: '0' }), config, groups), { text: '0.00', calculated: false })
})

test('missing or invalid calculation inputs do not fabricate zero', () => {
  for (const a of [undefined, null, '', ' ', false, [], {}, 'invalid'])
    assert.deepEqual(summaryMappedValue(row(11, { a }), config, groups), absent('missingData'))
  assert.deepEqual(summaryMappedValue(row(11, { a: 0 }), config, groups), { text: '0.0', calculated: true })
  assert.deepEqual(summaryMappedValue(row(11, { a: 3, first_score: 'invalid' }), config, groups), absent('missingData'))
})

test('false, zero, and a literal dash remain real values', () => {
  for (const [key, value, text] of [['flag', false, '否'], ['zero', 0, '0'], ['note', '-', '-']])
    assert.deepEqual(summaryMappedValue(row(11, { [key]: value }), mapped([source(11, key)]), groups), { text, calculated: false })
  assert.deepEqual(summaryMappedValue(row(11, { note: ' ' }), mapped([source(11, 'note')]), groups), absent('missingData'))
})

test('unavailable definitions and fields have distinct reasons', () => {
  assert.deepEqual(summaryMappedValue(row(11), config, []), absent('definitionUnavailable'))
  assert.deepEqual(summaryMappedValue(row(11), mapped([source(11, 'deleted')]), groups), absent('fieldUnavailable'))
})

const historical = freeze(group(11, [{ key: 'old', type: 'number' }, grade('legacy_score', '[old] + 1', 0)], 1))
const versionGroups = freeze({ '11:1': historical })
const versionConfig = freeze(mapped([source(11, 'first_score'), source(11, 'legacy_score', 1)]))

test('exact version precedes default regardless of source order', () => {
  for (const sources of [versionConfig.sources, [...versionConfig.sources].reverse()]) {
    assert.deepEqual(summaryMappedValue(row(11, { old: 5, a: 8 }, 1), mapped(sources), groups, versionGroups), { text: '6', calculated: true })
    assert.deepEqual(summaryMappedValue(row(11, { old: 5, a: 8 }, 2), mapped(sources), groups, versionGroups), { text: '16.0', calculated: true })
  }
})

test('exact version cannot fall back when historical schema or field is missing', () => {
  const instance = row(11, { old: 5, a: 8, legacy_score: 91 }, 1)
  assert.deepEqual(summaryMappedValue(instance, versionConfig, groups), absent('versionUnavailable'))
  assert.deepEqual(summaryMappedValue(instance, versionConfig, groups, { '11:1': group(11, [{ key: 'other', type: 'number' }], 1) }), absent('fieldUnavailable'))
  assert.deepEqual(summaryMappedValue(instance, mapped([source(11, 'legacy_score', 1)]), [], versionGroups), { text: '91', calculated: false })
})

test('current schema is eligible for an exact source only when its version matches', () => {
  const exact = mapped([source(11, 'first_score', 2)])
  assert.deepEqual(summaryMappedValue(row(11, { a: 3 }), exact, groups), { text: '6.0', calculated: true })
  assert.deepEqual(summaryMappedValue(row(11, { a: 3 }, 1), exact, groups), absent('unmapped'))
  const noVersion = groups.map(({ version: _version, ...item }) => item)
  assert.deepEqual(summaryMappedValue(row(11, { a: 3 }), exact, noVersion), absent('versionUnavailable'))
})

test('historical registry cannot lend another definition or mismatched version schema', () => {
  for (const invalid of [group(22, [grade('legacy_score', '99')], 1), { ...historical, version: 2 }])
    assert.deepEqual(summaryMappedValue(row(11, { old: 5 }, 1), versionConfig, groups, { '11:1': invalid }), absent('versionUnavailable'))
  const inherited = Object.create({ '11:1': historical })
  assert.deepEqual(summaryMappedValue(row(11, { old: 5 }, 1), versionConfig, groups, inherited), absent('versionUnavailable'))
})

test('duplicate matching sources are ambiguous, while exact matches shadow default conflicts', () => {
  const duplicateDefaults = [source(11, 'first_score'), source(11, 'note')]
  assert.deepEqual(summaryMappedValue(row(11, { a: 3 }), mapped(duplicateDefaults), groups), absent('ambiguousMapping'))
  const withExact = mapped([...duplicateDefaults, source(11, 'legacy_score', 1)])
  assert.deepEqual(summaryMappedValue(row(11, { old: 5 }, 1), withExact, groups, versionGroups), { text: '6', calculated: true })
  assert.deepEqual(summaryMappedValue(row(11, { old: 5 }, 1), mapped([...versionConfig.sources, source(11, 'legacy_score', 1)]), groups, versionGroups), absent('ambiguousMapping'))
})

test('duplicate groups or matching field keys are ambiguous instead of order dependent', () => {
  assert.deepEqual(summaryMappedValue(row(11, { a: 3 }), config, [groups[0], groups[0]]), absent('ambiguousMapping'))
  const duplicateField = { ...groups[0], columns: [...groups[0].columns, groups[0].columns[1]] }
  assert.deepEqual(summaryMappedValue(row(11, { a: 3 }), config, [duplicateField]), absent('ambiguousMapping'))
})

test('normalization preserves unavailable references, detaches data, and survives storage roundtrip', () => {
  const persisted = freeze([mapped([source(999, 'removed', 44), { ...source(11, 'first_score'), fieldLabel: '旧标签' }])])
  const normalized = normalizeSummaryMappedColumns(persisted)
  assert.deepEqual(normalized, persisted)
  assert.notEqual(normalized[0], persisted[0])
  assert.notEqual(normalized[0].sources[0], persisted[0].sources[0])
  assert.deepEqual(normalizeSummaryMappedColumns(JSON.parse(JSON.stringify(normalized))), persisted)
  assert.deepEqual(validateSummaryMappedColumn(persisted[0]), [])
})

test('normalization rejects malformed references without promoting invalid versions to defaults', () => {
  for (const raw of [null, undefined, {}, 3, 'bad JSON', '[{}]'])
    assert.deepEqual(normalizeSummaryMappedColumns(raw), [])
  const invalid = [null, [], {}, source('11', 'a'), source(-1, 'a'), source(1.5, 'a'), source(11, ''), source(11, 'x'.repeat(101)), source(11, 'a', null), source(11, 'a', 0), source(11, 'a', '1'), source(11, 'a', 1.5)]
  const normalized = normalizeSummaryMappedColumns([mapped([...invalid, source(999, 'unavailable', 77)])])
  assert.deepEqual(normalized[0].sources, [source(999, 'unavailable', 77)])
  assert.deepEqual(normalizeSummaryMappedColumns([null, [], {}, mapped([], { id: '' }), mapped([], { id: 'x'.repeat(81) }), mapped([], { label: ' ' })]), [])
})

test('normalization bounds columns, sources and display labels without truncating identity', () => {
  const oversized = Array.from({ length: 25 }, (_, i) => mapped(Array.from({ length: 55 }, (_, j) => ({ ...source(j + 1, `key${j}`), fieldLabel: 'f'.repeat(80) })), { id: `column${i}`, label: 'l'.repeat(70) }))
  const normalized = normalizeSummaryMappedColumns(oversized)
  assert.equal(normalized.length, 20)
  assert.ok(normalized.every(item => item.label.length === 50 && item.sources.length === 50 && item.sources.every(field => field.fieldLabel.length <= 50)))
  assert.equal(normalizeSummaryMappedColumns([mapped([], { enabled: 'true' })])[0].enabled, false)
  assert.deepEqual(normalizeSummaryMappedColumns([config, config]), [config])
})

test('normalization preserves source ambiguity for validation and safe rendering', () => {
  const duplicate = mapped([source(11, 'first_score'), source(11, 'note')])
  assert.deepEqual(normalizeSummaryMappedColumns([duplicate]), [duplicate])
  assert.deepEqual(validateSummaryMappedColumn(duplicate), ['duplicateSource'])
})

test('malicious stored property names do not pollute objects or supply inherited form data', () => {
  const payload = JSON.parse('[{"id":"__proto__","label":"安全","enabled":true,"__proto__":{"polluted":true},"sources":[{"definitionId":11,"fieldKey":"constructor","constructor":{"polluted":true}}]}]')
  const clean = normalizeSummaryMappedColumns(payload)
  assert.deepEqual(clean, [mapped([source(11, 'constructor')], { id: '__proto__', label: '安全' })])
  assert.equal(Object.prototype.polluted, undefined)
  assert.equal(Object.hasOwn(clean[0], '__proto__'), false)
  assert.deepEqual(normalizeSummaryMappedColumns([Object.create(config)]), [])
  const constructorGroup = group(11, [{ key: 'constructor', type: 'text' }])
  assert.deepEqual(summaryMappedValue(row(11), clean[0], [constructorGroup]), absent('missingData'))
  const inheritedData = Object.create({ a: 7, first_score: 30 })
  assert.deepEqual(summaryMappedValue(row(11, inheritedData), config, groups), absent('missingData'))
})

test('validation reports stable structural errors without consulting API availability', () => {
  assert.deepEqual(validateSummaryMappedColumn(mapped([])), ['sourcesRequired'])
  assert.deepEqual(validateSummaryMappedColumn(mapped([{}], { label: ' ' })), ['labelRequired', 'sourceRequired'])
  assert.deepEqual(validateSummaryMappedColumn(mapped([source(999, 'unknown', 800)])), [])
  assert.deepEqual(validateSummaryMappedColumn(mapped([source(11, 'a'), source(11, 'b', 1), source(11, 'c', 2)])), [])
  assert.deepEqual(validateSummaryMappedColumn(mapped([source(11, 'a', 1), source(11, 'b', 1)])), ['duplicateSource'])
  assert.ok(validateSummaryMappedColumn(mapped(Array.from({ length: 51 }, (_, i) => source(i + 1, 'a')))).includes('tooManySources'))
  for (const version of [null, 0, -1, 1.2, '2', Number.NaN, Number.POSITIVE_INFINITY])
    assert.ok(validateSummaryMappedColumn(mapped([source(11, 'a', version)])).includes('sourceRequired'))
  for (const overrides of [{ id: '' }, { id: 'i'.repeat(81) }, { label: 'l'.repeat(51) }])
    assert.ok(validateSummaryMappedColumn(mapped([source(11, 'a')], overrides)).length > 0)
})

test('enabled configurations expose synthetic form columns only', () => {
  assert.deepEqual(summaryMappedColumns([config, { ...config, id: 'off', enabled: false }]), [{ key: 'mapped:rating', label: '统一结果', kind: 'form', width: 'minmax(150px,1fr)', mobileHidden: true }])
})

function typeConfig(fields) {
  return {
    config: mapped(fields.map((_, i) => source(i + 1, 'field'))),
    groups: fields.map((field, i) => group(i + 1, [{ key: 'field', label: '类型', ...field }])),
  }
}

test('type warnings group numbers, amounts and numeric calculations together', () => {
  const compatible = typeConfig([{ type: 'number' }, { type: 'amount' }, { type: 'calculation', calculation: { expression: '1' } }])
  assert.equal(summaryMappedTypeMismatch(compatible.config, compatible.groups), false)
  const mixed = typeConfig([{ type: 'number' }, { type: 'text' }])
  assert.equal(summaryMappedTypeMismatch(mixed.config, mixed.groups), true)
  assert.deepEqual(validateSummaryMappedColumn(mixed.config), [])
  assert.deepEqual(summaryMappedValue(row(1, { field: 3 }), mixed.config, mixed.groups), { text: '3', calculated: false })
})

test('selection labels and label calculations are compatible regardless of calculation editor display', () => {
  const labelled = typeConfig([{ type: 'select' }, { type: 'radio' }, { type: 'calculation', calculation: { expression: '1', display: 'field', resultDisplay: { mode: 'label' } } }])
  assert.equal(summaryMappedTypeMismatch(labelled.config, labelled.groups), false)
  const numeric = typeConfig([{ type: 'select' }, { type: 'calculation', calculation: { expression: '1', display: 'label', resultDisplay: { mode: 'number' } } }])
  assert.equal(summaryMappedTypeMismatch(numeric.config, numeric.groups), true)
})

test('text and multi-select families group appropriately; distinct temporal and entity types warn', () => {
  for (const fields of [[{ type: 'text' }, { type: 'textarea' }, { type: 'phone' }, { type: 'email' }], [{ type: 'checkbox' }, { type: 'multi_select' }]]) {
    const compatible = typeConfig(fields)
    assert.equal(summaryMappedTypeMismatch(compatible.config, compatible.groups), false)
  }
  for (const [left, right] of [['date', 'datetime'], ['date', 'date_range'], ['time', 'date'], ['user', 'department'], ['user', 'user_multi'], ['attachment', 'detail_list'], ['boolean', 'text']]) {
    const mixed = typeConfig([{ type: left }, { type: right }])
    assert.equal(summaryMappedTypeMismatch(mixed.config, mixed.groups), true, `${left} vs ${right}`)
  }
})

test('type warnings use historical schemas and ignore unavailable schemas', () => {
  const schemas = typeConfig([{ type: 'number' }, { type: 'number' }])
  const historyConfig = { ...schemas.config, sources: [source(1, 'field', 1), source(2, 'field')] }
  assert.equal(summaryMappedTypeMismatch(historyConfig, schemas.groups), false)
  assert.equal(summaryMappedTypeMismatch(historyConfig, schemas.groups, { '1:1': group(1, [{ key: 'field', type: 'text' }], 1) }), true)
  assert.equal(summaryMappedTypeMismatch(mapped([source(999, 'gone'), source(1, 'field')]), schemas.groups), false)
})

test('all operations accept frozen inputs without changing records, schemas or configuration', () => {
  const instance = freeze(row(11, { old: 5 }, 1))
  const before = JSON.stringify({ instance, config: versionConfig, groups, versionGroups })
  normalizeSummaryMappedColumns([versionConfig])
  validateSummaryMappedColumn(versionConfig)
  summaryMappedColumns([versionConfig])
  assert.deepEqual(summaryMappedValue(instance, versionConfig, groups, versionGroups), { text: '6', calculated: true })
  summaryMappedTypeMismatch(versionConfig, groups, versionGroups)
  assert.equal(JSON.stringify({ instance, config: versionConfig, groups, versionGroups }), before)
})
