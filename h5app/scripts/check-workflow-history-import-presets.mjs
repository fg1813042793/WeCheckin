import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const ts = require('typescript')
const { effectScope, reactive } = require('vue')

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../src')
const cache = new Map()
let api
function load(path) {
  if (cache.has(path))
    return cache.get(path)
  const compiled = ts.transpileModule(readFileSync(path, 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
    fileName: path,
  }).outputText
  const module = { exports: {} }
  const localRequire = (request) => {
    if (request === '@/api/workflow') {
      return {
        listWorkflowInstances: query => api.list(query),
        getWorkflowInstance: id => api.detail(id),
      }
    }
    if (request.startsWith('@/'))
      return load(resolve(root, `${request.slice(2)}.ts`))
    if (request.startsWith('.'))
      return load(resolve(dirname(path), `${request}.ts`))
    return require(request)
  }
  runInThisContext(`(function(module,exports,require){${compiled}\n})`)(module, module.exports, localRequire)
  cache.set(path, module.exports)
  return module.exports
}
const helper = load(resolve(root, 'pages/workflow/workflow-history-import-preset.ts'))
const {
  normalizeWorkflowHistoryImportConfig: normalize,
  workflowHistoryImportNeedsAttention,
} = helper
assert.equal(normalize(undefined), null)
assert.equal(normalize(null), null)
assert.equal(normalize({ enabled: false, mappings: [{ sourceKey: 'deleted' }] }), null)
assert.deepEqual(normalize({ enabled: true }), {
  enabled: true,
  sourceMode: 'select_history',
  sourceStatus: 'completed',
  mappingMode: 'same_key',
  overwrite: false,
  mappings: [],
})
assert.equal(workflowHistoryImportNeedsAttention({ imported: 1, overwritten: 0, retained: 0, skipped: 0, filtered: 0, dynamic: 0 }), false)
assert.equal(workflowHistoryImportNeedsAttention({ imported: 1, overwritten: 0, retained: 2, skipped: 0, filtered: 0, dynamic: 0 }), false)
assert.equal(workflowHistoryImportNeedsAttention({ imported: 1, overwritten: 0, retained: 0, skipped: 1, filtered: 0, dynamic: 0 }), true)
assert.equal(workflowHistoryImportNeedsAttention({ imported: 1, overwritten: 0, retained: 0, skipped: 0, filtered: 1, dynamic: 0 }), true)
assert.equal(workflowHistoryImportNeedsAttention({ imported: 1, overwritten: 0, retained: 0, skipped: 0, filtered: 0, dynamic: 1 }), true)
const staleManualDraft = [
  { sourceKey: '', targetKey: '', columns: [] },
  { sourceKey: '', targetKey: '', columns: [] },
]
assert.deepEqual(normalize({ enabled: true, mappingMode: 'same_key', mappings: staleManualDraft }).mappings, [], 'same-key mode ignores saved manual drafts')
assert.throws(() => normalize({ enabled: true, mappingMode: 'same_key', mappings: Array.from({ length: 101 }, () => ({})) }), /invalidConfig/)
assert.throws(() => normalize({ enabled: true, mappingMode: 'same_key', mappings: [{ sourceKey: 'a'.repeat(101) }] }), /invalidConfig/)
for (const config of [
  { mappingMode: 'mapping', mappings: [{ sourceKey: 'invalid key', targetKey: 'valid' }] },
  { mappingMode: 'mapping', mappings: [{ sourceKey: 'valid', targetKey: 'a'.repeat(101) }] },
  { sourceMode: 'unknown' },
  { sourceStatus: 'running' },
  { mappingMode: 'unknown' },
  { overwrite: 'false' },
  { mappingMode: 'mapping', mappings: [] },
  { mappingMode: 'mapping', mappings: [{ sourceKey: 'a', targetKey: 'b', columns: [] }] },
  { mappingMode: 'mapping', mappings: [{ sourceKey: 'a', targetKey: 'b' }, { sourceKey: 'a', targetKey: 'c' }] },
  { mappingMode: 'mapping', mappings: [{ sourceKey: 'a', targetKey: 'b', columns: [{ sourceKey: 'c', targetKey: 'd', columns: [] }] }] },
]) {
  assert.throws(() => normalize({ enabled: true, ...config }), /invalidConfig/)
}
const { useWorkflowHistoryImportPreset } = load(resolve(root, 'composables/useWorkflowHistoryImportPreset.ts'))
const field = (key, type = 'text', extra = {}) => ({ key, type, label: key, ...extra })
const summary = { id: 'recent', definitionId: 7, status: 'completed' }
const detail = {
  instance: summary,
  form: [field('old'), field('same')],
  formData: { old: 'history', same: 'same value' },
  startNodeId: 'start',
  fieldPermissions: {},
}
function deferred() {
  let resolve
  let reject
  const promise = new Promise((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}
function harness(config = { enabled: true }) {
  const state = reactive({
    visible: true,
    definition: { id: 7, version: 2, form: [field('new'), field('same')], historyImport: config, fieldPermissions: {} },
    data: {},
    access: { new: 'write', same: 'write' },
  })
  const calls = []
  const applied = []
  api = {
    list: async (query) => {
      calls.push(query)
      return { data: { list: [summary], total: 1 } }
    },
    detail: async () => ({ data: structuredClone(detail) }),
  }
  const scope = effectScope()
  const runtime = scope.run(() => useWorkflowHistoryImportPreset({
    visible: () => state.visible,
    definition: () => state.definition,
    targetData: () => state.data,
    targetAccess: () => state.access,
    translate: (key, params) => params ? `${key}:${JSON.stringify(params)}` : key,
    apply: payload => applied.push(payload),
    contextChanged: () => { state.visible = false },
  }))
  return { state, calls, applied, runtime, stop: () => scope.stop() }
}
for (const config of [undefined, null, { enabled: false }]) {
  const h = harness(null)
  h.state.definition.historyImport = config
  assert.equal(h.runtime.enabled.value, false, 'legacy configurations must not become presets')
  await h.runtime.select(summary)
  assert.equal(h.applied.length, 0)
  h.stop()
}
{
  const h = harness({ enabled: true, mappingMode: 'mapping', mappings: [{ sourceKey: 'old', targetKey: 'new' }] })
  await h.runtime.select(summary)
  assert.deepEqual(h.applied[0].patch, { new: 'history' })
  await h.runtime.select(summary)
  assert.equal(h.applied.length, 1, 'a session applies at most once')
  h.stop()
}
{
  const h = harness({ enabled: true, sourceMode: 'latest_history' })
  await h.runtime.latest()
  assert.deepEqual(h.calls, [{ definitionId: 7, scope: 'started', status: 'completed', page: 1, pageSize: 1 }])
  assert.deepEqual(h.applied[0].patch, { same: 'same value' })
  h.stop()
}
{
  const h = harness({ enabled: true, sourceMode: 'latest_history', sourceStatus: 'all' })
  api.list = async (query) => {
    h.calls.push(query)
    return { data: { list: [{ ...summary, status: 'running' }] } }
  }
  api.detail = async () => ({ data: { ...structuredClone(detail), instance: { ...summary, status: 'running' } } })
  await h.runtime.latest()
  assert.equal(h.calls[0].status, undefined)
  assert.equal(h.applied.length, 1)
  h.stop()
}
for (const cancel of ['close', 'reopen', 'definition', 'config', 'permissions', 'unmount']) {
  const h = harness()
  const pending = deferred()
  api.detail = () => pending.promise
  const work = h.runtime.select(summary)
  assert.equal(h.runtime.busy.value, true)
  if (cancel === 'close')
    h.state.visible = false
  if (cancel === 'reopen') {
    h.state.visible = false
    h.state.visible = true
  }
  if (cancel === 'definition')
    h.state.definition.id = 8
  if (cancel === 'config')
    h.state.definition.historyImport.overwrite = true
  if (cancel === 'permissions')
    h.state.access.same = 'hidden'
  if (cancel === 'unmount')
    h.stop()
  pending.resolve({ data: structuredClone(detail) })
  await work
  assert.equal(h.applied.length, 0, `${cancel} must invalidate a pending detail`)
  h.stop()
}
{
  const h = harness({ enabled: true, sourceMode: 'latest_history' })
  const pending = deferred()
  api.list = () => pending.promise
  let detailReads = 0
  api.detail = async () => {
    detailReads++
    return { data: detail }
  }
  const work = h.runtime.latest()
  h.state.visible = false
  pending.resolve({ data: { list: [summary] } })
  await work
  assert.equal(detailReads, 0, 'cancelled latest lookup must not read or apply the detail')
  h.stop()
}
{
  const h = harness()
  const pending = deferred()
  api.detail = () => pending.promise
  const first = h.runtime.select(summary)
  const second = h.runtime.select(summary)
  pending.resolve({ data: detail })
  await Promise.all([first, second])
  assert.equal(h.applied.length, 1, 'double clicks must not apply twice')
  h.stop()
}
for (const failure of ['empty', 'network', 'status', 'definition', 'id', 'zero', 'source-read', 'invalid']) {
  const h = harness({ enabled: true, sourceMode: 'latest_history' })
  if (failure === 'empty')
    api.list = async () => ({ data: { list: [] } })
  if (failure === 'network')
    api.list = async () => { throw new Error('offline') }
  if (failure === 'status')
    api.detail = async () => ({ data: { ...detail, instance: { ...summary, status: 'running' } } })
  if (failure === 'definition')
    api.detail = async () => ({ data: { ...detail, instance: { ...summary, definitionId: 8 } } })
  if (failure === 'id')
    api.detail = async () => ({ data: { ...detail, instance: { ...summary, id: 'different-record' } } })
  if (failure === 'zero')
    h.state.data.same = 'keep me'
  if (failure === 'source-read')
    api.detail = async () => ({ data: { ...detail, fieldPermissions: { start: [{ field: 'same', access: 'read' }] } } })
  if (failure === 'invalid') {
    h.stop()
    const invalid = harness({ enabled: true, sourceMode: 'invalid' })
    await invalid.runtime.latest()
    assert.ok(invalid.runtime.error.value)
    assert.equal(invalid.applied.length, 0)
    invalid.stop()
    continue
  }
  await h.runtime.latest()
  assert.equal(h.applied.length, 0, `${failure} must not report success`)
  assert.ok(h.runtime.error.value, `${failure} needs visible feedback`)
  h.stop()
}
console.log('workflow history import preset runtime checks passed')

// Execute the real SFC setup with Vue reactivity; only platform hooks and HTTP are replaced.
// No browser, mounted DOM, or live endpoint is used.
function dialogHarness(config, width = 390) {
  const vue = require('vue')
  const props = reactive({
    modelValue: false,
    definition: { id: 7, version: 2, form: [field('same')], historyImport: config, fieldPermissions: {} },
    targetData: {},
    targetAccess: { same: 'write' },
  })
  const events = []
  const calls = []
  api = {
    list: async (query) => {
      calls.push(query)
      return { data: { list: [summary], total: 1 } }
    },
    detail: async () => ({ data: structuredClone(detail) }),
  }
  const path = resolve(root, 'pages/workflow/components/WorkflowHistoryImportDialog.vue')
  const script = readFileSync(path, 'utf8').match(/<script setup lang="ts">([\s\S]*?)<\/script>/)[1]
  const compiled = ts.transpileModule(`${script}\nexport { currentStep, instances, listLoading, listError, mobileLoadingMore, sourceDetail, presetError, statusFilter, page, openDialog, selectInstance, closeDialog, goToPreview, applyImport, queryInstances, loadInstances, loadMoreInstances };`, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
    fileName: path,
  }).outputText
  const module = { exports: {} }
  const localRequire = (request) => {
    if (request === 'vue')
      return { ...vue, onMounted: () => {}, onBeforeUnmount: vue.onScopeDispose }
    if (request === 'uview-pro')
      return { useLocale: () => ({ t: key => key }) }
    if (request === '@/api/workflow')
      return { listWorkflowInstances: query => api.list(query), getWorkflowInstance: id => api.detail(id) }
    if (request.startsWith('@/'))
      return load(resolve(root, `${request.slice(2)}.ts`))
    return load(resolve(dirname(path), `${request}.ts`))
  }
  const scope = effectScope()
  scope.run(() => runInThisContext(`(function(module,exports,require,defineProps,defineEmits,uni){${compiled}\n})`)(
    module,
    module.exports,
    localRequire,
    () => props,
    () => (name, value) => {
      events.push([name, value])
      if (name === 'update:modelValue')
        props.modelValue = value
    },
    { getSystemInfoSync: () => ({ windowWidth: width }), showToast: () => {}, hideKeyboard: () => {} },
  ))
  return { props, events, calls, dialog: module.exports, stop: () => scope.stop() }
}
async function settle() {
  for (let index = 0; index < 8; index++)
    await require('vue').nextTick()
}
for (const config of [undefined, null, { enabled: false }]) {
  const h = dialogHarness(config)
  h.props.modelValue = true
  await settle()
  assert.equal(h.dialog.currentStep.value, 'select')
  assert.equal(h.calls[0].pageSize, 10, 'keep the mobile list batch size')
  assert.equal(h.calls[0].status, undefined, 'legacy wizard keeps all statuses by default')
  await h.dialog.selectInstance(summary)
  assert.equal(h.dialog.currentStep.value, 'configure')
  assert.equal(h.events.some(([name]) => name === 'apply'), false)
  h.dialog.goToPreview()
  assert.equal(h.dialog.currentStep.value, 'preview')
  h.dialog.applyImport()
  assert.deepEqual(h.events.find(([name]) => name === 'apply')[1].patch, { same: 'same value' })
  h.dialog.applyImport()
  assert.equal(h.events.filter(([name]) => name === 'apply').length, 1)
  h.stop()
}
{
  const h = dialogHarness({ enabled: true }, 1280)
  h.props.modelValue = true
  await settle()
  assert.equal(h.calls[0].status, 'completed')
  assert.equal(h.calls[0].pageSize, 6, 'keep desktop pagination')
  h.dialog.statusFilter.value = 'running'
  h.dialog.queryInstances()
  await settle()
  assert.equal(h.calls[1].status, 'completed', 'filters must not widen a completed-only preset')
  await h.dialog.selectInstance(summary)
  assert.equal(h.dialog.currentStep.value, 'select', 'preset never visits configure or preview')
  assert.equal(h.events.filter(([name]) => name === 'apply').length, 1)
  assert.equal(h.props.modelValue, false)
  h.stop()
}
{
  const h = dialogHarness({ enabled: true, sourceMode: 'latest_history' })
  h.props.modelValue = true
  await settle()
  assert.equal(h.calls.length, 1)
  assert.equal(h.calls[0].pageSize, 1)
  assert.equal(h.events.filter(([name]) => name === 'apply').length, 1)
  h.stop()
}
for (const config of [null, { enabled: true }]) {
  const h = dialogHarness(config)
  const pending = deferred()
  api.list = () => pending.promise
  h.props.modelValue = true
  await settle()
  h.dialog.closeDialog()
  pending.resolve({ data: { list: [summary], total: 1 } })
  await settle()
  assert.deepEqual(h.dialog.instances.value, [], 'late list response cannot repopulate a closed dialog')
  assert.equal(h.events.some(([name]) => name === 'apply'), false)
  h.stop()
}
for (const config of [null, { enabled: true }]) {
  const h = dialogHarness(config)
  h.props.modelValue = true
  await settle()
  const pending = deferred()
  api.detail = () => pending.promise
  const work = h.dialog.selectInstance(summary)
  h.dialog.closeDialog()
  h.props.modelValue = true
  await settle()
  pending.resolve({ data: detail })
  await work
  assert.equal(h.dialog.sourceDetail.value, null)
  assert.equal(h.dialog.currentStep.value, 'select')
  assert.equal(h.events.some(([name]) => name === 'apply'), false, 'old detail cannot apply in a new dialog session')
  h.stop()
}
for (const config of [null, { enabled: true }]) {
  const h = dialogHarness(config)
  h.props.modelValue = true
  await settle()
  api.list = async () => {
    throw new Error('temporary list failure')
  }
  await h.dialog.loadInstances()
  assert.equal(h.dialog.listError.value, 'temporary list failure')

  const pending = deferred()
  api.list = () => pending.promise
  const recovery = h.dialog.loadInstances({ type: 'click' })
  assert.equal(h.dialog.listError.value, '', 'retry must clear the previous list error before requesting')
  assert.equal(h.dialog.listLoading.value, true, 'a click event must trigger replacement loading')
  assert.equal(h.dialog.mobileLoadingMore.value, false, 'a click event is not an append request')
  const recovered = { ...summary, id: 'recovered' }
  pending.resolve({ data: { list: [recovered], total: 3 } })
  await recovery
  assert.deepEqual(h.dialog.instances.value.map(item => item.id), ['recovered'])

  const replacement = { ...summary, id: 'replacement' }
  api.list = async () => ({ data: { list: [replacement], total: 3 } })
  await h.dialog.loadInstances({ type: 'click' })
  assert.deepEqual(h.dialog.instances.value.map(item => item.id), ['replacement'], 'reload replaces existing records')

  const more = { ...summary, id: 'more' }
  const morePending = deferred()
  let moreQuery
  api.list = (query) => {
    moreQuery = query
    return morePending.promise
  }
  h.dialog.loadMoreInstances()
  assert.equal(h.dialog.mobileLoadingMore.value, true)
  assert.equal(h.dialog.listLoading.value, false)
  morePending.resolve({ data: { list: [replacement, more], total: 2 } })
  await settle()
  assert.equal(moreQuery.page, 2)
  assert.deepEqual(h.dialog.instances.value.map(item => item.id), ['replacement', 'more'], 'load more appends and deduplicates')
  assert.equal(h.dialog.listError.value, '')
  assert.equal(h.dialog.mobileLoadingMore.value, false)
  h.stop()
}
const dialogTemplate = readFileSync(resolve(root, 'pages/workflow/components/WorkflowHistoryImportDialog.vue'), 'utf8')
assert.equal((dialogTemplate.match(/@click="loadInstances\(\)"/g) || []).length, 2, 'both error and empty-state reload buttons must omit the click event')
assert.equal(dialogTemplate.includes('@click="loadInstances"'), false)
console.log('workflow history import dialog runtime checks passed')
