import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const runtimeSource = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowRuntimeForm.vue', import.meta.url),
  'utf8',
)
const startSource = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowStartPage.vue', import.meta.url),
  'utf8',
)
const detailSource = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowDetailPanel.vue', import.meta.url),
  'utf8',
)
const formDetailSource = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowFormDetailPage.vue', import.meta.url),
  'utf8',
)

for (const snippet of [
  'presentation?: \'continuous\' | \'tabs\'',
  'presentation: \'continuous\'',
  'const activeTabKey = ref(\'\')',
  'const tabScrollLeft = ref(0)',
  'const tabScrollMax = ref(0)',
  'const tabOverflow = computed',
  'const tabMode = computed',
  'const renderedFields = computed',
  'class="workflow-form__tabs"',
  'class="workflow-form__tabs-shell"',
  'class="workflow-form__tab-scroll-button workflow-form__tab-scroll-button--left"',
  'class="workflow-form__tab-scroll-button workflow-form__tab-scroll-button--right"',
  'name="arrow-left"',
  'name="arrow-right"',
  ':scroll-left="tabScrollLeft"',
  'v-if="tabOverflow"',
  '@wheel="scrollTabsHorizontally"',
  'class="workflow-form__tab"',
  ':aria-selected="activeTabKey === field.key"',
  '@click="selectTab(field.key)"',
  'v-for="field in renderedFields"',
  'firstErrorTab',
  'activeTabKey.value = target.key',
  'function syncTabOverflow()',
  'function resolveTabScrollerElement()',
  'querySelectorAll<HTMLElement>(\'.uni-scroll-view-scrollbar-hidden\')',
  'scrollers[scrollers.length - 1]',
  'function ensureActiveTabVisible()',
  'function scrollTabs(direction: -1 | 1)',
  'function scrollTabsHorizontally(event: WheelEvent)',
  'window.addEventListener(\'resize\', syncTabOverflow)',
  'watch(activeTabKey',
  'watch(tabOverflow',
]) {
  assert.ok(runtimeSource.includes(snippet), `WorkflowRuntimeForm.vue missing ${snippet}`)
}

const tabsTemplate = runtimeSource.match(/<scroll-view[\s\S]*?class="workflow-form__tabs"[\s\S]*?<\/scroll-view>/)?.[0] || ''
assert.ok(tabsTemplate, 'workflow form tabs template missing')
for (const forbidden of ['1 行', 'index + 1', '1 / 5', '共 {{']) {
  assert.ok(!tabsTemplate.includes(forbidden), `workflow form tab must not show counts or sequence: ${forbidden}`)
}

for (const snippet of [
  'const formPresentationStorageKey = \'workflow_start_form_presentation_v1\'',
  'const formPresentation = ref<WorkflowFormPresentation>',
  'const desktopFormPresentation = ref(resolveDesktopFormPresentation())',
  'const effectiveFormPresentation = computed',
  'return desktopFormPresentation.value ? formPresentation.value : \'continuous\'',
  'function toggleFormPresentation()',
  'formPresentation.value = formPresentation.value === \'tabs\' ? \'continuous\' : \'tabs\'',
  'uni.setStorageSync(formPresentationStorageKey, formPresentation.value)',
  'v-if="desktopFormPresentation"',
  '分组显示',
  '连续显示',
  ':presentation="effectiveFormPresentation"',
  'class="workflow-start-page__form-actions"',
]) {
  assert.ok(startSource.includes(snippet), `WorkflowStartPage.vue missing ${snippet}`)
}

const toolbar = startSource.match(/<view class="workflow-start-page__form-toolbar">[\s\S]*?<WorkflowRuntimeForm/)?.[0] || ''
assert.ok(toolbar.indexOf('workflow-start-page__presentation-action') < toolbar.indexOf('workflow-start-page__import-action'), 'presentation toggle must stay to the left of import')

for (const snippet of [
  'type WorkflowFormPresentation = \'continuous\' | \'tabs\'',
  'const formPresentationStorageKey = \'workflow_start_form_presentation_v1\'',
  'const formPresentation = ref<WorkflowFormPresentation>(loadFormPresentation())',
  'const desktopFormPresentation = computed(() => !mobileInteractionDialog.value)',
  'const effectiveFormPresentation = computed<WorkflowFormPresentation>',
  'return desktopFormPresentation.value ? formPresentation.value : \'continuous\'',
  'function toggleFormPresentation()',
  'v-if="desktopFormPresentation"',
  'custom-class="workflow-detail-panel__presentation-action"',
  '分组显示',
  '连续显示',
  ':presentation="effectiveFormPresentation"',
  'class="workflow-detail-panel__page-card-actions"',
]) {
  assert.ok(detailSource.includes(snippet), `WorkflowDetailPanel.vue missing approval form presentation behavior: ${snippet}`)
}

assert.equal(
  detailSource.match(/:presentation="effectiveFormPresentation"/g)?.length || 0,
  1,
  'only the approval page form should receive the switchable presentation mode',
)
const approvalHeader = detailSource.match(/<view v-if="pagePresentation" class="workflow-detail-panel__page-card-head">[\s\S]*?<\/view>\s*<WorkflowRuntimeForm/)?.[0] || ''
assert.ok(approvalHeader, 'approval page card header must include the form presentation action')
assert.ok(approvalHeader.indexOf('workflow-detail-panel__presentation-action') < approvalHeader.indexOf('workflow-detail-panel__status-tag'), 'approval presentation toggle must stay before the status tag')

for (const snippet of [
  'type WorkflowFormPresentation = \'continuous\' | \'tabs\'',
  'const formPresentationStorageKey = \'workflow_start_form_presentation_v1\'',
  'const formPresentation = ref<WorkflowFormPresentation>(loadFormPresentation())',
  'const desktopFormPresentation = ref(resolveDesktopFormPresentation())',
  'const effectiveFormPresentation = computed<WorkflowFormPresentation>',
  'return desktopFormPresentation.value ? formPresentation.value : \'continuous\'',
  'function syncStoredFormPresentation()',
  'formPresentation.value = loadFormPresentation()',
  'function toggleFormPresentation()',
  'v-if="detail && desktopFormPresentation"',
  'custom-class="workflow-form-detail-page__presentation-action"',
  '分组显示',
  '连续显示',
  ':presentation="effectiveFormPresentation"',
  'class="workflow-form-detail-page__section-head-main"',
]) {
  assert.ok(formDetailSource.includes(snippet), `WorkflowFormDetailPage.vue missing form presentation behavior: ${snippet}`)
}

assert.equal(
  formDetailSource.match(/:presentation="effectiveFormPresentation"/g)?.length || 0,
  1,
  'standalone form detail should bind the shared presentation mode once',
)

const formDetailPageHeader = formDetailSource.match(/<view class="workflow-form-detail-page__header">[\s\S]*?<view v-if="loading"/)?.[0] || ''
assert.ok(formDetailPageHeader, 'standalone form detail page header missing')
assert.ok(formDetailPageHeader.includes('<u-tag'), 'standalone form detail page header must keep the status tag')
assert.ok(!formDetailPageHeader.includes('workflow-form-detail-page__presentation-action'), 'standalone form detail page header must not contain the presentation toggle')

const formDetailSectionHeader = formDetailSource.match(/<view class="workflow-form-detail-page__section-head">[\s\S]*?<WorkflowRuntimeForm/)?.[0] || ''
assert.ok(formDetailSectionHeader, 'standalone form detail section header missing')
assert.ok(formDetailSectionHeader.includes('workflow-form-detail-page__section-head-main'), 'standalone form detail section header main content missing')
assert.ok(formDetailSectionHeader.includes('workflow-form-detail-page__presentation-action'), 'standalone form detail section header must contain the presentation toggle')
assert.ok(formDetailSectionHeader.indexOf('workflow-form-detail-page__section-head-main') < formDetailSectionHeader.indexOf('workflow-form-detail-page__presentation-action'), 'standalone form detail title must stay before the presentation toggle')

console.log('Workflow desktop form tabs checks passed')
