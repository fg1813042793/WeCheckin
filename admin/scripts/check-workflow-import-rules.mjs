import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import * as vue from 'vue'
import { parse, compileScript } from '@vue/compiler-sfc'
import { renderToString } from '@vue/server-renderer'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = path => readFileSync(resolve(root, path), 'utf8')
const load = (path, modules = {}) => {
  assert.ok(existsSync(resolve(root, path)), `missing implementation: ${path}`)
  return executeTypeScriptModule(read(path), path, { vue, ...modules })
}
const plain = value => JSON.parse(JSON.stringify(value))
const types = load('src/types/workflow.ts')
assert.equal(typeof types.normalizeWorkflowHistoryImportConfig, 'function', 'history-import defaults need a shared normalizer')
const layout = load('src/views/workflow/formLayout.ts')
const helpers = load('src/views/workflow/designer/historyImportRules.ts', { '../formLayout': layout })
const { useWorkflowHistoryImportRules } = load('src/views/workflow/designer/useWorkflowHistoryImportRules.ts', {
  '@/types/workflow': types, './historyImportRules': helpers,
})
const { useWorkflowHistoryImportPanels } = load('src/views/workflow/designer/useWorkflowHistoryImportPanels.ts')
const field = (key, type = 'text', extra = {}) => ({ key, label: key, type, ...extra })
const draft = () => ({
  schemaVersion: 1, key: 'fixture', name: '离线规则测试', edges: [],
  nodes: [{ id: 'start', type: 'start', name: '开始', formPermissions: [
    { field: 'locked', access: 'read' }, { field: 'hidden', access: 'hidden' },
  ] }],
  form: [field('group', 'group', { fields: [field('previous'), field('current', 'textarea')] }),
    field('number', 'number'), field('amount', 'amount'), field('locked'), field('hidden'),
    field('result', 'calculation'), field('description', 'description'),
    field('previous_rows', 'detail_list', { columns: [field('old_goal'), field('old_score', 'number'), field('total', 'calculation'), field('nested', 'detail_list')] }),
    field('current_rows', 'detail_list', { columns: [field('goal', 'textarea'), field('score', 'amount'), field('total', 'calculation')] })],
})
const defaults = { enabled: false, sourceMode: 'select_history', sourceStatus: 'completed', mappingMode: 'same_key', overwrite: false, mappings: [] }
assert.deepEqual(plain(types.normalizeWorkflowHistoryImportConfig()), defaults)
assert.equal(types.cloneDraft(draft()).historyImport, undefined, 'old drafts must not acquire an enabled preset')
assert.deepEqual(plain(types.cloneDraft({ ...draft(), historyImport: { enabled: true } }).historyImport), { ...defaults, enabled: true })

const props = vue.reactive({ draft: draft(), readonly: false })
let changes = 0
const editor = useWorkflowHistoryImportRules(props, () => { changes++ })
assert.equal(props.draft.historyImport, undefined, 'opening the editor must not dirty old drafts')
editor.setEnabled(true)
assert.equal(changes, 1)
assert.deepEqual(plain(props.draft.historyImport), { ...defaults, enabled: true })
editor.setMappingMode('mapping')
assert.match(editor.errors.value.join(' '), /至少.*映射/)
editor.addMapping()
editor.setMapping(0, 'sourceKey', 'previous')
editor.setMapping(0, 'targetKey', 'current')
assert.deepEqual(plain(editor.config.value.mappings), [{ sourceKey: 'previous', targetKey: 'current' }])
assert.equal(editor.errors.value.length, 0)
assert.equal(editor.targetOptions(0).find(option => option.value === 'locked').disabled, true)
assert.equal(editor.targetOptions(0).find(option => option.value === 'hidden').disabled, true)
assert.equal(editor.targetOptions(0).find(option => option.value === 'number').disabled, true)
assert.equal(editor.targetOptions(0).find(option => option.value === 'result').disabled, true)
assert.equal(editor.sourceOptions(0).some(option => option.value === 'description'), false)
const beforeInvalid = plain(props.draft)
editor.setMapping(0, 'targetKey', 'locked')
editor.setMapping(0, 'targetKey', 'number')
assert.deepEqual(plain(props.draft), beforeInvalid, 'handlers must enforce disabled choices')

editor.addMapping()
assert.equal(editor.sourceOptions(1).find(option => option.value === 'previous').disabled, true)
editor.setMapping(1, 'sourceKey', 'previous_rows')
editor.setMapping(1, 'targetKey', 'current_rows')
assert.equal(editor.isDetailMapping(1), true)
assert.equal(editor.config.value.mappings[1].columns, undefined, 'omitted columns mean same-key columns')
assert.match(editor.errors.value.join(' '), /明细列/, 'different detail schemas need a custom column map')
editor.addColumn(1)
editor.setColumn(1, 0, 'sourceKey', 'old_goal')
editor.setColumn(1, 0, 'targetKey', 'goal')
assert.equal(editor.columnOptions(1, 0, 'targetKey').find(option => option.value === 'score').disabled, true)
assert.equal(editor.columnOptions(1, 0, 'sourceKey').find(option => option.value === 'nested').disabled, true)
assert.equal(editor.columnOptions(1, 0, 'targetKey').find(option => option.value === 'total').disabled, true)
editor.addColumn(1)
editor.setColumn(1, 1, 'sourceKey', 'old_score')
editor.setColumn(1, 1, 'targetKey', 'score')
assert.equal(editor.errors.value.length, 0)
editor.setSourceMode('latest_history')
editor.setSourceStatus('all')
editor.setOverwrite(true)
const configured = plain(props.draft)
const roundtrip = types.cloneDraft(JSON.parse(JSON.stringify(props.draft)))
assert.deepEqual(plain(roundtrip), configured, 'all draft fields and nested mappings survive save payload/readback')
roundtrip.historyImport.mappings[1].columns[0].targetKey = 'changed'
assert.equal(props.draft.historyImport.mappings[1].columns[0].targetKey, 'goal', 'clone must isolate nested columns')
editor.setEnabled(false)
assert.deepEqual(plain(props.draft.historyImport.mappings), configured.historyImport.mappings)
props.draft.form = []
assert.equal(editor.errors.value.length, 0, 'disabled stale mappings must not block other edits')
editor.setEnabled(true)
assert.match(editor.errors.value.join(' '), /不存在/)
props.draft = types.cloneDraft(configured)
editor.setMapping(1, 'sourceKey', 'number')
assert.equal(editor.config.value.mappings[1].columns, undefined, 'changing a detail field removes stale columns')
assert.equal(editor.config.value.mappings[1].targetKey, '', 'incompatible target must be cleared')

props.draft = types.cloneDraft(configured)
props.readonly = true
const beforeReadonly = plain(props.draft)
const eventsBeforeReadonly = changes
editor.setEnabled(false)
editor.setSourceMode('select_history')
editor.setSourceStatus('completed')
editor.setMappingMode('same_key')
editor.setOverwrite(false)
editor.addMapping()
editor.removeMapping(0)
editor.setMapping(0, 'targetKey', 'previous')
editor.addColumn(1)
editor.removeColumn(1, 0)
editor.setColumn(1, 0, 'targetKey', 'score')
assert.deepEqual(plain(props.draft), beforeReadonly)
assert.equal(changes, eventsBeforeReadonly, 'readonly handlers emit no changes')
props.readonly = false
editor.removeColumn(1, 1)
editor.removeColumn(1, 0)
assert.equal(editor.config.value.mappings[1].columns, undefined, 'removing final custom column restores same-key behavior')
props.draft = types.cloneDraft(configured)
editor.setColumn(1, 0, 'sourceKey', undefined)
assert.equal(editor.config.value.mappings[1].columns[0].sourceKey, '', 'Element Plus clear events normalize to an empty key')
assert.equal(editor.config.value.mappings[1].columns[0].targetKey, '')
editor.setMapping(0, 'sourceKey', undefined)
assert.equal(editor.config.value.mappings[0].sourceKey, '')
assert.equal(editor.config.value.mappings[0].targetKey, '')
assert.ok(editor.errors.value.length > 0)
const manualDraft = plain(editor.config.value.mappings)
const manualErrors = plain(editor.errors.value)
editor.setMappingMode('same_key')
assert.equal(editor.errors.value.length, 0, 'same-key mode ignores incomplete stored manual mappings')
props.draft = types.cloneDraft(JSON.parse(JSON.stringify(props.draft)))
assert.deepEqual(plain(editor.config.value.mappings), manualDraft, 'same-key save roundtrip retains manual mapping drafts')
editor.setMappingMode('mapping')
assert.deepEqual(plain(editor.config.value.mappings), manualDraft, 'switching back restores the unchanged mapping draft')
assert.deepEqual(plain(editor.errors.value), manualErrors, 'manual mapping validation resumes when switching back')

for (const [source, target, compatible] of [['text', 'email', true], ['phone', 'textarea', true], ['number', 'amount', true], ['select', 'radio', true], ['multi_select', 'checkbox', true], ['date', 'datetime', false], ['number', 'text', false], ['calculation', 'calculation', false], ['description', 'description', false]]) {
  assert.equal(helpers.historyImportTypesCompatible(field('s', source), field('t', target)), compatible, `${source} -> ${target}`)
}
assert.equal(helpers.historyImportTypesCompatible(field('s', 'number', { calculation: { expression: '1+1' } }), field('t', 'number')), false, 'a calculation-backed field is never an import source')
assert.equal(helpers.historyImportTypesCompatible(field('s', 'unknown_type'), field('t', 'unknown_type')), false, 'unknown remote field types fail closed')
const invalid = (patch, mappings = []) => helpers.historyImportRuleErrors({ ...defaults, enabled: true, mappingMode: 'mapping', mappings, ...patch }, draft())
const unusedMappings = [
  { sourceKey: 'deleted', targetKey: 'deleted' },
  { sourceKey: 'deleted', targetKey: 'deleted' },
  { sourceKey: 'previous', targetKey: 'number', columns: [{ sourceKey: '', targetKey: '' }] },
  { sourceKey: 'previous_rows', targetKey: 'current_rows', columns: [] },
]
assert.equal(invalid({ mappingMode: 'same_key' }, unusedMappings).length, 0, 'unused missing, duplicate, incompatible and invalid column mappings are ignored')
assert.ok(invalid({ mappingMode: 'same_key', sourceMode: 'invalid' }, unusedMappings).some(error => /来源方式/.test(error)))
assert.ok(invalid({ mappingMode: 'same_key', sourceStatus: 'invalid' }, unusedMappings).some(error => /来源状态/.test(error)))
assert.ok(invalid({ mappingMode: 'same_key' }, Array.from({ length: 101 }, () => unusedMappings[0])).some(error => /100/.test(error)), 'unused mapping count is still bounded')
assert.ok(invalid({ mappingMode: 'same_key' }, [{ sourceKey: 'a'.repeat(101), targetKey: 'deleted' }]).some(error => /100/.test(error)), 'unused mapping keys are still bounded')
assert.ok(invalid({ mappingMode: 'same_key' }, [{ sourceKey: 'deleted', targetKey: 'deleted', columns: Array.from({ length: 101 }, () => ({ sourceKey: '', targetKey: '' })) }]).some(error => /100/.test(error)), 'unused column mappings are still bounded')
assert.ok(invalid({}, [{ sourceKey: 'previous', targetKey: 'current' }, { sourceKey: 'previous', targetKey: 'current' }]).some(error => /重复/.test(error)))
assert.ok(invalid({}, [{ sourceKey: 'previous', targetKey: 'current', columns: [{ sourceKey: 'a', targetKey: 'b' }] }]).some(error => /普通字段/.test(error)))
assert.ok(invalid({}, [{ sourceKey: 'previous_rows', targetKey: 'current_rows', columns: [] }]).some(error => /1.*100/.test(error)))
assert.ok(invalid({ sourceMode: 'invalid' }).some(error => /来源方式/.test(error)))
assert.ok(invalid({ enabled: false }, [{ sourceKey: 'a'.repeat(101), targetKey: 'missing' }]).some(error => /100/), 'disabled stale drafts still enforce payload bounds')
const sameKeyRows = { ...draft(), form: [field('rows', 'detail_list', { columns: [field('goal'), field('score', 'number')] })] }
assert.equal(helpers.historyImportRuleErrors({ ...defaults, enabled: true, mappingMode: 'mapping', mappings: [{ sourceKey: 'rows', targetKey: 'rows' }] }, sameKeyRows).length, 0)

// Disclosure state must never enter the saved payload or dirty the draft.
const panelScope = vue.effectScope()
panelScope.run(() => {
  const panelProps = vue.reactive({ draft: types.cloneDraft(configured), readonly: false })
  let panelChanges = 0
  const controls = useWorkflowHistoryImportRules(panelProps, () => panelChanges++)
  const panels = useWorkflowHistoryImportPanels(controls.config, controls.errors)
  assert.equal(panels.basicOpen.value, false, 'existing valid settings start compact')
  assert.equal(panels.mappingsOpen.value, false)
  assert.deepEqual(plain(panels.columnsOpen.value), [])
  const toggle = open => { const target = { open }; return { target, currentTarget: target } }
  panels.togglePanel('basic', toggle(true))
  panels.togglePanel('mappings', toggle(true))
  panels.toggleColumns(1, toggle(true))
  panels.togglePanel('mappings', { target: { open: false }, currentTarget: { open: true } })
  assert.equal(panels.mappingsOpen.value, true, 'nested toggle must not change its parent')
  panels.togglePanel('basic', toggle(false))
  panels.togglePanel('mappings', toggle(false))
  assert.equal(panelChanges, 0, 'folding panels does not dirty the form')
  assert.deepEqual(plain(panelProps.draft), configured)
  panels.toggleColumns(1, toggle(false))
  controls.setColumn(1, 0, 'sourceKey', undefined)
  assert.equal(panels.mappingsOpen.value, true, 'invalid mapping is revealed')
  assert.deepEqual(plain(panels.columnsOpen.value), [1], 'invalid detail columns are revealed')
  panels.removeMapping(0)
  controls.removeMapping(0)
  assert.deepEqual(plain(panels.columnsOpen.value), [0], 'expanded rows follow deletion indices')
  panelProps.draft = types.cloneDraft(configured)
  panels.togglePanel('basic', toggle(false))
  controls.setEnabled(false)
  controls.setEnabled(true)
  assert.equal(panels.basicOpen.value, true, 'newly enabled rules show their settings')
  controls.setMappingMode('same_key')
  panels.togglePanel('mappings', toggle(false))
  controls.setMappingMode('mapping')
  assert.equal(panels.mappingsOpen.value, true, 'switching to custom mapping reveals its editor')
  panelProps.readonly = true
  const readonlyEvents = panelChanges
  panels.togglePanel('basic', toggle(false))
  panels.togglePanel('basic', toggle(true))
  assert.equal(panels.basicOpen.value, true, 'readonly settings can still be inspected')
  assert.equal(panelChanges, readonlyEvents)
  assert.deepEqual(plain(panelProps.draft), configured)
})
panelScope.stop()

// Compile/render the real SFC without browser or network. Capture control callbacks
// so readonly and ordinary interactions exercise the template/controller wiring.
async function renderEditor(model, readonly = false) {
  const fileName = 'src/views/workflow/designer/components/WorkflowHistoryImportRules.vue'
  const { descriptor } = parse(read(fileName), { filename: fileName })
  const compiled = compileScript(descriptor, { id: 'history-import-test', inlineTemplate: true })
  const component = executeTypeScriptModule(compiled.content, fileName + '.ts', {
    vue, '../useWorkflowHistoryImportRules': { useWorkflowHistoryImportRules },
    '../useWorkflowHistoryImportPanels': { useWorkflowHistoryImportPanels }, '../historyImportRules': helpers,
  }).default
  const controls = []
  const app = vue.createSSRApp(component, { draft: model, readonly })
  for (const name of ['ElForm', 'ElFormItem', 'ElSwitch', 'ElSelect', 'ElOption', 'ElButton', 'ElAlert']) {
    app.component(name, {
      inheritAttrs: false,
      setup: (_props, { attrs, slots }) => {
        controls.push({ name, ...attrs })
        return () => vue.h('div', { 'data-control': name }, [String(attrs.label || attrs.title || ''), slots.default?.()])
      },
    })
  }
  return { html: await renderToString(app), controls }
}
const legacy = draft()
const legacyView = await renderEditor(legacy)
assert.match(legacyView.html, /本人.*历史/)
assert.match(legacyView.html, /三步向导/)
assert.equal(legacy.historyImport, undefined)
legacyView.controls.find(control => control.name === 'ElSwitch')['onUpdate:modelValue'](true)
assert.equal(legacy.historyImport.enabled, true)
const configuredView = await renderEditor(types.cloneDraft(configured))
assert.match(configuredView.html, /覆盖已有内容/)
assert.match(configuredView.html, /明细列映射/)
assert.ok(configuredView.controls.filter(control => control.name === 'ElSelect').every(control => control.filterable !== undefined), 'selectors are searchable')
const readonlyModel = types.cloneDraft(configured)
const readonlyView = await renderEditor(readonlyModel, true)
assert.ok(readonlyView.controls.filter(control => ['ElSelect', 'ElSwitch', 'ElButton'].includes(control.name)).every(control => control.disabled === true))
readonlyView.controls.find(control => control.name === 'ElSwitch')['onUpdate:modelValue'](false)
assert.deepEqual(plain(readonlyModel), configured)
const host = read('src/views/workflow/designer/components/WorkflowStartConfig.vue')
assert.match(host, /config-section__index">05</)
assert.match(host, /导入规则设置/)
assert.match(host, /<WorkflowHistoryImportRules[\s\S]*?:draft="draft"[\s\S]*?:readonly="readonly"[\s\S]*?@change="emit\('change'\)"/)
assert.ok(host.trimEnd().split('\n').length <= 808, 'legacy host must not grow')
console.log('workflow history import rules: defaults, roundtrip, fields, columns, disclosure state, readonly controller and SFC render passed')
