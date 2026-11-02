import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const ts = require('typescript')
const vue = require('vue')

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../src')
const plain = value => JSON.parse(JSON.stringify(value))
const compile = source => ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
let readDetail
const cache = new Map()
function load(path) {
  if (cache.has(path))
    return cache.get(path)
  const module = { exports: {} }
  const localRequire = (request) => {
    if (request === '@/api/workflow')
      return { getWorkflowSummaryInstance: id => readDetail(id) }
    if (request.startsWith('@/'))
      return load(resolve(root, `${request.slice(2)}.ts`))
    if (request.startsWith('.'))
      return load(resolve(dirname(path), `${request}.ts`))
    return require(request)
  }
  runInThisContext(`(function(module,exports,require){${compile(readFileSync(path, 'utf8'))}\n})`)(module, module.exports, localRequire)
  cache.set(path, module.exports)
  return module.exports
}
const mappings = load(resolve(root, 'pages/workflow/workflow-summary-mappings.ts'))
const { workflowSummaryFormFieldGroups } = load(resolve(root, 'pages/workflow/workflow-summary-columns.ts'))
const { useWorkflowSummaryVersionFields } = load(resolve(root, 'composables/useWorkflowSummaryVersionFields.ts'))
const definitions = [{ id: 1, name: '管理名称', version: 3, form: [{ key: 'score', label: '当前得分', type: 'number' }, { key: 'note', label: '备注', type: 'text' }] }]
const mapped = { id: 'rating', label: '最终分档', enabled: true, sources: [{ definitionId: 1, fieldKey: 'score' }, { definitionId: 2, fieldKey: 'other_score' }] }

// Run the actual storage handlers, preserving existing v2 and legacy selections.
const source = readFileSync(resolve(root, 'pages/workflow/components/WorkflowSummarySection.vue'), 'utf8').match(/<script setup lang="ts">([\s\S]*?)<\/script>/)[1]
const ast = ts.createSourceFile('summary.ts', source, ts.ScriptTarget.Latest, true)
const names = ['loadColumnSettings', 'saveColumnSettings', 'emptyColumnSettings', 'normalizeStoredColumnSettings', 'normalizeColumnSettings', 'normalizeSelectedFormKeys', 'formColumnDefinitionId', 'saveMappedColumn', 'resetColumnSettings']
const functions = ast.statements.filter(node => ts.isFunctionDeclaration(node) && names.includes(node.name?.text)).map(node => node.getText(ast)).join('\n')
const storage = new Map()
const columnSettings = { value: {} }
const scope = {
  columnSettings,
  formFieldGroups: { value: workflowSummaryFormFieldGroups(definitions) },
  summaryColumnStorageKey: 'v2',
  legacySummaryColumnStorageKey: 'v1',
  normalizeSummaryMappedColumns: mappings.normalizeSummaryMappedColumns,
  formColumnKeyword: { value: '' },
  formFieldDropdownOpen: { value: false },
  t: key => key,
  uni: { getStorageSync: key => storage.get(key), setStorageSync: (key, value) => storage.set(key, plain(value)), showToast: () => {} },
}
const handlers = runInThisContext(`(function(${Object.keys(scope).join(',')}){${compile(functions)}\nreturn {${names.join(',')}}})`)(...Object.values(scope))
storage.set('v2', { baseHiddenKeys: ['version'], activeDefinitionId: '1', formSelectedKeysByDefinition: { 1: ['form:1:score', 'form:1:note'] } })
columnSettings.value = handlers.loadColumnSettings()
assert.equal(columnSettings.value.mappedColumns.length, 0)
handlers.saveMappedColumn(mapped, true)
assert.deepEqual(columnSettings.value.formSelectedKeysByDefinition[1], ['form:1:note'])
assert.deepEqual(plain(handlers.loadColumnSettings().mappedColumns), [mapped])
handlers.saveMappedColumn({ ...mapped, label: '改名后的分档' }, false)
assert.equal(columnSettings.value.mappedColumns.length, 1)
assert.equal(handlers.loadColumnSettings().mappedColumns[0].label, '改名后的分档')
scope.formFieldGroups.value = []
assert.equal(handlers.normalizeColumnSettings(columnSettings.value).mappedColumns[0].sources.length, 2, 'temporarily unavailable definitions must not erase mappings')
handlers.resetColumnSettings()
assert.equal(handlers.loadColumnSettings().mappedColumns.length, 0)
assert.equal(handlers.loadColumnSettings().baseHiddenKeys.length, 0)
storage.delete('v2')
storage.set('v1', { baseHiddenKeys: ['operator'], selectedFormKeys: ['form:1:score'] })
assert.deepEqual(plain(handlers.loadColumnSettings().formSelectedKeysByDefinition), { 1: ['form:1:score'] })

const effect = vue.effectScope()
const rows = vue.ref([{ id: 'history-1', definitionId: 1, definitionVersion: 1 }])
const permission = vue.ref(1)
const fields = effect.run(() => useWorkflowSummaryVersionFields(() => definitions, () => rows.value, () => permission.value))
let calls = 0
let finish
readDetail = async () => {
  calls++
  return await new Promise(resolve => finish = resolve)
}
await fields.ensure(1, 3)
assert.equal(calls, 0, 'current published schema needs no detail request')
const pendingA = fields.ensure(1, 1)
const pendingB = fields.ensure(1, 1)
assert.equal(pendingA, pendingB, 'duplicate historical schema loads share a request')
assert.equal(calls, 1)
finish({ data: { instance: rows.value[0], form: [{ key: 'old_score', type: 'number', label: '历史得分' }], formData: { secretValue: 'not-schema-data' } } })
await pendingA
assert.equal(fields.groups.value['1:1'].columns[0].field.key, 'old_score')
assert.equal(JSON.stringify(fields.groups.value).includes('not-schema-data'), false)
await fields.ensure(1, 1)
assert.equal(calls, 1, 'already fetched version is cached')
await fields.ensure(1, 2)
assert.equal(fields.errors.value['1:2'], 'versionNotOnPage')
rows.value = [{ id: 'history-2', definitionId: 1, definitionVersion: 2 }]
readDetail = async () => {
  throw new Error('denied')
}
await fields.ensure(1, 2)
assert.equal(fields.errors.value['1:2'], 'versionLoadFailed')
assert.equal(fields.groups.value['1:2'], undefined)
readDetail = async () => ({ data: { instance: { ...rows.value[0], definitionVersion: 8 }, form: [] } })
await fields.ensure(1, 2)
assert.equal(fields.errors.value['1:2'], 'versionLoadFailed', 'mismatched historical responses are rejected')
readDetail = async () => await new Promise(resolve => finish = resolve)
const stale = fields.ensure(1, 2)
permission.value++
await vue.nextTick()
finish({ data: { instance: rows.value[0], form: [{ key: 'stale', type: 'number' }] } })
await stale
assert.equal(Object.keys(fields.groups.value).length, 0, 'permission changes clear caches and reject stale responses')
effect.stop()
// Refresh responses can arrive in either order. A newly published version makes
// an exact mapping historical even if the list response arrived first.
const versionWatch = ast.statements.find(node => ts.isExpressionStatement(node) && node.getText(ast).startsWith('watch(') && node.getText(ast).includes('versionFields.ensure'))
assert.ok(versionWatch)
for (const order of ['list-first', 'definitions-first']) {
  const refreshScope = vue.effectScope()
  const props = vue.reactive({ definitions: plain(definitions) })
  const records = vue.ref(order === 'list-first' ? [{ id: 'v3', definitionId: 1, definitionVersion: 3, formData: { score: 7 } }] : [])
  const settings = vue.ref({ mappedColumns: [{ ...mapped, sources: [{ definitionId: 1, fieldKey: 'score', version: 3 }] }] })
  const auth = vue.reactive({ permissionVersion: 1 })
  let requests = 0
  readDetail = async () => {
    requests++
    return { data: { instance: records.value[0], form: definitions[0].form } }
  }
  let versionFields
  refreshScope.run(() => {
    versionFields = useWorkflowSummaryVersionFields(() => props.definitions, () => records.value, () => auth.permissionVersion)
    runInThisContext(`(function(watch,instances,columnSettings,auth,props,versionFields){${compile(versionWatch.getText(ast))}\n})`)(vue.watch, records, settings, auth, props, versionFields)
  })
  if (order === 'list-first') {
    records.value = [...records.value]
    await vue.nextTick()
  }
  props.definitions = [{ ...definitions[0], version: 4, form: [{ key: 'new_score', label: '新版得分', type: 'number' }] }]
  await vue.nextTick()
  if (order === 'definitions-first')
    records.value = [{ id: 'v3', definitionId: 1, definitionVersion: 3, formData: { score: 7 } }]
  await vue.nextTick()
  await Promise.resolve()
  assert.equal(requests, 1, `${order}: exact-version schemas must reload after definition version changes`)
  assert.equal(mappings.summaryMappedValue(records.value[0], settings.value.mappedColumns[0], workflowSummaryFormFieldGroups(props.definitions), versionFields.groups.value).text, '7')
  refreshScope.stop()
}
console.log('mapped settings: v1/v2 compatibility, persistence, replacement, reset, version cache, scope denial and race checks passed')
