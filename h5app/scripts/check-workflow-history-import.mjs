import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const ts = require('typescript')

function evaluateTypeScript(relativePath, dependencies = {}) {
  const url = new URL(relativePath, import.meta.url)
  const source = readFileSync(url, 'utf8')
  const compiled = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2022,
      esModuleInterop: true,
    },
    fileName: url.pathname,
  }).outputText
  const module = { exports: {} }
  const localRequire = (request) => {
    if (request in dependencies)
      return dependencies[request]
    return require(request)
  }
  const execute = runInThisContext(`(function (module, exports, require) {${compiled}\n})`)
  execute(module, module.exports, localRequire)
  return module.exports
}

const calculationModule = evaluateTypeScript('../src/pages/workflow/workflow-calculation.ts')
const formModule = evaluateTypeScript('../src/pages/workflow/workflow-form.ts', {
  './workflow-calculation': calculationModule,
})
const importModule = evaluateTypeScript('../src/pages/workflow/workflow-history-import.ts', {
  './workflow-form': formModule,
})

const {
  buildWorkflowImportMappings,
  isWorkflowImportTypeCompatible,
  previewWorkflowHistoryImport,
  suggestWorkflowImportMappings,
} = importModule

const mappedCalculationOptions = formModule.normalizeWorkflowOptions([
  { label: 'A+', value: '1.5', scoreValue: 1.5 },
], {
  calculationValueField: 'scoreValue',
})
assert.equal(mappedCalculationOptions[0]?.calculationValue, 1.5)

const specificValidationMessage = '与上级分档不一致，请确认！！！'
assert.equal(
  formModule.workflowFormValidationMessage({ result: specificValidationMessage }, '请检查表单填写内容'),
  specificValidationMessage,
)
assert.equal(
  formModule.workflowFormValidationMessage({}, '请检查表单填写内容'),
  '请检查表单填写内容',
)

const field = (key, type, extra = {}) => ({ key, label: key, type, ...extra })

assert.equal(
  isWorkflowImportTypeCompatible(field('source', 'text'), field('target', 'textarea')),
  true,
)
assert.equal(
  isWorkflowImportTypeCompatible(field('source', 'number'), field('target', 'amount')),
  true,
)
assert.equal(
  isWorkflowImportTypeCompatible(field('source', 'date'), field('target', 'datetime')),
  false,
)
assert.equal(
  isWorkflowImportTypeCompatible(field('source', 'calculation'), field('target', 'number')),
  false,
)

assert.deepEqual(
  buildWorkflowImportMappings(
    [field('remark', 'text'), field('readOnly', 'text')],
    [field('remark', 'textarea'), field('readOnly', 'text')],
    { remark: 'write', readOnly: 'read' },
  ),
  [{ sourceKey: 'remark', targetKey: 'remark' }],
)

assert.deepEqual(
  suggestWorkflowImportMappings(
    [field('oldRemark', 'text', { label: '备注' })],
    [field('remark', 'textarea', { label: '备注' })],
    { remark: 'write' },
  ),
  [{ sourceKey: 'oldRemark', targetKey: 'remark' }],
)

const preview = previewWorkflowHistoryImport({
  sourceFields: [
    field('emptyTarget', 'text'),
    field('filledTarget', 'text'),
    field('zeroTarget', 'number'),
    field('falseTarget', 'boolean'),
    field('readonlyTarget', 'text'),
    field('calculatedTarget', 'number'),
  ],
  sourceData: {
    emptyTarget: '导入内容',
    filledTarget: '历史内容',
    zeroTarget: 8,
    falseTarget: true,
    readonlyTarget: '不可写',
    calculatedTarget: 99,
  },
  targetFields: [
    field('emptyTarget', 'textarea'),
    field('filledTarget', 'text'),
    field('zeroTarget', 'amount'),
    field('falseTarget', 'boolean'),
    field('readonlyTarget', 'text'),
    field('calculatedTarget', 'calculation'),
  ],
  targetData: {
    emptyTarget: '   ',
    filledTarget: '当前内容',
    zeroTarget: 0,
    falseTarget: false,
    readonlyTarget: '',
    calculatedTarget: undefined,
  },
  targetAccess: {
    emptyTarget: 'write',
    filledTarget: 'write',
    zeroTarget: 'write',
    falseTarget: 'write',
    readonlyTarget: 'read',
    calculatedTarget: 'read',
  },
  mappings: [
    { sourceKey: 'emptyTarget', targetKey: 'emptyTarget' },
    { sourceKey: 'filledTarget', targetKey: 'filledTarget' },
    { sourceKey: 'zeroTarget', targetKey: 'zeroTarget' },
    { sourceKey: 'falseTarget', targetKey: 'falseTarget' },
    { sourceKey: 'readonlyTarget', targetKey: 'readonlyTarget' },
    { sourceKey: 'calculatedTarget', targetKey: 'calculatedTarget' },
  ],
  overwrite: false,
})

assert.deepEqual(preview.patch, { emptyTarget: '导入内容' })
assert.equal(preview.importCount, 1)
assert.equal(preview.overwriteCount, 0)
assert.equal(preview.retainCount, 3)
assert.equal(preview.skipCount, 2)

const overwritePreview = previewWorkflowHistoryImport({
  sourceFields: [field('mapped', 'text'), field('unmapped', 'text')],
  sourceData: { mapped: '历史内容', unmapped: '不应导入' },
  targetFields: [field('mapped', 'text'), field('unmapped', 'text')],
  targetData: { mapped: '当前内容', unmapped: '保留内容' },
  targetAccess: { mapped: 'write', unmapped: 'write' },
  mappings: [{ sourceKey: 'mapped', targetKey: 'mapped' }],
  overwrite: true,
})
assert.deepEqual(overwritePreview.patch, { mapped: '历史内容' })
assert.equal(overwritePreview.overwriteCount, 1)

const optionPreview = previewWorkflowHistoryImport({
  sourceFields: [
    field('single', 'select'),
    field('multiple', 'multi_select'),
    field('remote', 'select'),
  ],
  sourceData: {
    single: 'removed',
    multiple: ['kept', 'removed'],
    remote: 'remote-value',
  },
  targetFields: [
    field('single', 'radio', { options: [{ label: '保留', value: 'kept' }] }),
    field('multiple', 'checkbox', { options: [{ label: '保留', value: 'kept' }] }),
    field('remote', 'select', { optionSource: { type: 'api', url: '/api/options' } }),
  ],
  targetData: { single: '', multiple: [], remote: '' },
  targetAccess: { single: 'write', multiple: 'write', remote: 'write' },
  mappings: [
    { sourceKey: 'single', targetKey: 'single' },
    { sourceKey: 'multiple', targetKey: 'multiple' },
    { sourceKey: 'remote', targetKey: 'remote' },
  ],
  overwrite: false,
})
assert.deepEqual(optionPreview.patch, {
  multiple: ['kept'],
  remote: 'remote-value',
})
assert.equal(optionPreview.items.find(item => item.targetKey === 'single')?.result, 'skipped')
assert.equal(optionPreview.items.find(item => item.targetKey === 'multiple')?.result, 'filtered')
assert.equal(optionPreview.items.find(item => item.targetKey === 'remote')?.result, 'dynamic_option')

assert.throws(() => previewWorkflowHistoryImport({
  sourceFields: [field('first', 'text'), field('second', 'text')],
  sourceData: { first: 'A', second: 'B' },
  targetFields: [field('target', 'text')],
  targetData: { target: '' },
  targetAccess: { target: 'write' },
  mappings: [
    { sourceKey: 'first', targetKey: 'target' },
    { sourceKey: 'second', targetKey: 'target' },
  ],
  overwrite: false,
}), /目标字段不能重复/)

const detailPreview = previewWorkflowHistoryImport({
  sourceFields: [field('items', 'detail_list', {
    rowKey: 'sourceRowId',
    columns: [field('name', 'text'), field('score', 'number'), field('legacy', 'text')],
  })],
  sourceData: {
    items: [{ sourceRowId: 'old-row', name: '目标一', score: 90, legacy: '丢弃' }],
  },
  targetFields: [field('items', 'detail_list', {
    rowKey: 'targetRowId',
    columns: [field('name', 'textarea'), field('score', 'amount')],
  })],
  targetData: { items: [] },
  targetAccess: { items: 'write' },
  mappings: [{ sourceKey: 'items', targetKey: 'items' }],
  overwrite: false,
})
assert.equal(detailPreview.importCount, 1)
assert.equal(detailPreview.patch.items.length, 1)
assert.equal(detailPreview.patch.items[0].name, '目标一')
assert.equal(detailPreview.patch.items[0].score, 90)
assert.ok(String(detailPreview.patch.items[0].targetRowId).startsWith('row_'))
assert.equal('sourceRowId' in detailPreview.patch.items[0], false)
assert.equal('legacy' in detailPreview.patch.items[0], false)

const dialogUrl = new URL('../src/pages/workflow/components/WorkflowHistoryImportDialog.vue', import.meta.url)
assert.ok(existsSync(dialogUrl), 'WorkflowHistoryImportDialog.vue is required')
const dialogSource = readFileSync(dialogUrl, 'utf8')
for (const snippet of [
  'listWorkflowInstances',
  'getWorkflowInstance',
  'scope: \'started\'',
  'definitionId: props.definition.id',
  '全部导入',
  '字段映射',
  '覆盖已有内容',
  'previewWorkflowHistoryImport',
  'WorkflowHistoryImportMapping',
  'emit(\'apply\'',
  'instanceTitle',
  'statusFilter',
  '<u-pagination',
  'listLoading',
  'listError',
  '暂无历史申请',
  '重新加载',
  'type HistoryImportStep = \'select\' | \'configure\' | \'preview\'',
  'const currentStep = ref<HistoryImportStep>(\'select\')',
  'currentStep.value = \'configure\'',
  'v-if="currentStep === \'select\'"',
  ':class="{ \'workflow-history-import-dialog__actions--mobile-hidden\': mobile && currentStep === \'select\' }"',
  'workflow-history-import-dialog__cancel',
  'v-if="!mobile"',
  'v-else-if="currentStep === \'configure\'"',
  'goToPreviousStep',
  'goToPreview',
  'workflow-history-import-dialog__close',
  'title="关闭"',
  'role="button"',
  'tabindex="0"',
  '@keydown.enter.prevent',
  '@keydown.space.prevent',
  'width: 32px',
  'height: 32px',
  'workflow-history-import-dialog__mapping-remove',
  'title="删除映射"',
  'grid-template-columns: minmax(0, 1fr) 20px minmax(0, 1fr) 32px',
  '<u-search',
  '@search="submitInstanceSearch"',
  '@custom="submitInstanceSearch"',
  '@clear="clearInstanceSearch"',
  ':show-action="false"',
  '@search="submitInstanceSearch"',
  '@custom="submitInstanceSearch"',
  '@focus="keepInstanceSearchVisible"',
  ':scroll-top="dialogScrollTop"',
  'workflow-history-import-dialog__filter-secondary',
  'function mappingOptionLabel',
  '{{ mappingOptionLabel(sourceField) }}',
  '{{ mappingOptionLabel(targetField) }}',
  'function submitInstanceSearch',
  'uni.hideKeyboard()',
  'function keepInstanceSearchVisible',
  'workflow-history-import-dialog__filter-button--search',
  '.workflow-history-import-dialog__search :deep(.u-content)',
  'height: 36px !important',
  ':adjust-position="false"',
  'onKeyboardHeightChange',
  'offKeyboardHeightChange',
  'visualViewport',
  'position: sticky',
  'searchFocusTimer',
  'workflowFieldAccessMap',
  'const sourceAccess = computed',
  'sourceDetail.value.fieldPermissions?.[sourceDetail.value.startNodeId || \'start\']',
  'sourceAccess.value[field.key] === \'write\'',
  'sourceFields.value,',
  'const mobileBatchSize = 10',
  'const mobileVisibleRows = 3',
  'const desktopPageSize = 6',
  'const pageSize = computed(() => mobile.value ? mobileBatchSize : desktopPageSize)',
  'pageSize: pageSize.value',
  ':scroll-y="mobile"',
  '@scrolltolower="loadMoreInstances"',
  'function loadMoreInstances',
  'mobileLoadingMore',
  'v-if="!mobile && total > pageSize"',
  'mobileListStyle',
  '--mobile-history-visible-rows',
]) {
  assert.ok(dialogSource.includes(snippet), `WorkflowHistoryImportDialog.vue missing ${snippet}`)
}
const searchFilterButtonIndex = dialogSource.indexOf('workflow-history-import-dialog__filter-button--search')
const resetFilterButtonIndex = dialogSource.indexOf('@click="resetFilters"')
assert.ok(searchFilterButtonIndex >= 0 && searchFilterButtonIndex < resetFilterButtonIndex, 'search button must appear before reset')
assert.equal(
  dialogSource.includes('custom-class="app-icon-button app-icon-button--small"'),
  false,
  'mapping remove action must not rely on u-button intrinsic width',
)
assert.equal(
  dialogSource.includes('<u-button type="primary" size="small" :loading="listLoading" @click="queryInstances">'),
  false,
  'history selection must use the search action built into u-search',
)
for (const legacyHeading of [
  '1. 选择历史申请',
  '2. 配置导入方式',
  '3. 导入预览',
]) {
  assert.equal(dialogSource.includes(legacyHeading), false, `WorkflowHistoryImportDialog.vue must remove ${legacyHeading}`)
}

const startPageSource = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowStartPage.vue', import.meta.url),
  'utf8',
)
const validationPageSources = [
  startPageSource,
  readFileSync(new URL('../src/pages/workflow/components/WorkflowDetailPanel.vue', import.meta.url), 'utf8'),
  readFileSync(new URL('../src/pages/workflow/components/WorkflowFormRevisionPage.vue', import.meta.url), 'utf8'),
  readFileSync(new URL('../src/pages/workflow/components/WorkflowCompletedFormRevisionPage.vue', import.meta.url), 'utf8'),
]
for (const pageSource of validationPageSources) {
  assert.ok(pageSource.includes('workflowFormValidationMessage'), 'workflow submit page must prefer the specific validation message')
}
for (const genericMessage of ['请检查表单必填项', '请检查表单填写内容']) {
  for (const pageSource of validationPageSources) {
    assert.equal(pageSource.includes(`title: '${genericMessage}'`), false, `workflow submit page must not hardcode ${genericMessage}`)
  }
}
for (const snippet of [
  'dingtalk_h5:button:workflow:history-import',
  'const canImportHistory = computed',
  'starterAllowed.value',
  'auth.hasApiPermission(\'dingtalk_h5:api:workflow:start\')',
  'auth.hasApiPermission(\'dingtalk_h5:api:workflow:view\')',
  'WorkflowHistoryImportDialog',
  'v-model="historyImportVisible"',
  '@apply="applyHistoryImport"',
  '>导入<',
  'function applyHistoryImport',
  'initialWorkflowFormData(value.form || [], {',
  '...formData.value,',
  '...payload.patch,',
]) {
  assert.ok(startPageSource.includes(snippet), `WorkflowStartPage.vue missing ${snippet}`)
}

console.log('workflow history import checks passed')
