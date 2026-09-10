import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'

const designerUrl = new URL('../src/views/survey/SurveyDesigner.vue', import.meta.url)
const historyUrl = new URL('../src/views/survey/composables/useSurveyDesignerHistory.ts', import.meta.url)
const validationUrl = new URL('../src/views/survey/survey-designer-validation.ts', import.meta.url)
const issuesUrl = new URL('../src/views/survey/components/SurveyDesignerIssues.vue', import.meta.url)
const outlineUrl = new URL('../src/views/survey/components/SurveyDesignerOutline.vue', import.meta.url)
const panelTitleUrl = new URL('../src/views/survey/components/SurveyDesignerPanelTitle.vue', import.meta.url)
const paletteUrl = new URL('../src/views/survey/components/SurveyDesignerQuestionPalette.vue', import.meta.url)
const responsesPanelUrl = new URL('../src/views/survey/components/SurveyDesignerResponsesPanel.vue', import.meta.url)
const reportPanelUrl = new URL('../src/views/survey/components/SurveyDesignerReportPanel.vue', import.meta.url)
const logicPanelUrl = new URL('../src/views/survey/components/SurveyDesignerLogicPanel.vue', import.meta.url)
const settingsPanelUrl = new URL('../src/views/survey/components/SurveyDesignerSettingsPanel.vue', import.meta.url)
const resultSettingsUrl = new URL('../src/views/survey/components/SurveyDesignerResultSettings.vue', import.meta.url)
const questionPropertiesUrl = new URL('../src/views/survey/components/SurveyDesignerQuestionProperties.vue', import.meta.url)
const sidebarUrl = new URL('../src/views/survey/components/SurveyDesignerSidebar.vue', import.meta.url)
const formulaDialogUrl = new URL('../src/views/survey/components/SurveyDesignerFormulaDialog.vue', import.meta.url)
const textImportDialogUrl = new URL('../src/views/survey/components/SurveyDesignerTextImportDialog.vue', import.meta.url)
const bankUploadDialogUrl = new URL('../src/views/survey/components/SurveyDesignerBankUploadDialog.vue', import.meta.url)
const presetDialogUrl = new URL('../src/views/survey/components/SurveyDesignerPresetDialog.vue', import.meta.url)
const responsesUrl = new URL('../src/views/survey/composables/useSurveyResponses.ts', import.meta.url)
const questionsUrl = new URL('../src/views/survey/composables/useSurveyDesignerQuestions.ts', import.meta.url)
const transferUrl = new URL('../src/views/survey/composables/useSurveyDesignerTransfer.ts', import.meta.url)
const documentUrl = new URL('../src/views/survey/composables/useSurveyDesignerDocument.ts', import.meta.url)
const draggableUrl = new URL('../src/views/survey/formkit/DraggableList.vue', import.meta.url)
const responsiveUrl = new URL('../src/views/survey/survey-designer-responsive.css', import.meta.url)
const designerStyleUrl = new URL('../src/views/survey/survey-designer.css', import.meta.url)
const questionTypesUrl = new URL('../src/views/survey/survey-designer-question-types.ts', import.meta.url)
const textTransferUrl = new URL('../src/views/survey/survey-designer-text-transfer.ts', import.meta.url)

assert.ok(existsSync(historyUrl), 'survey designer history composable is required')
assert.ok(existsSync(validationUrl), 'survey designer validation module is required')
assert.ok(existsSync(issuesUrl), 'survey designer issues component is required')
assert.ok(existsSync(outlineUrl), 'survey designer outline component is required')
assert.ok(existsSync(panelTitleUrl), 'survey designer property panel title component is required')
assert.ok(existsSync(paletteUrl), 'survey designer question palette component is required')
assert.ok(existsSync(responsesPanelUrl), 'survey designer responses panel is required')
assert.ok(existsSync(reportPanelUrl), 'survey designer report panel is required')
assert.ok(existsSync(logicPanelUrl), 'survey designer logic panel is required')
assert.ok(existsSync(settingsPanelUrl), 'survey designer settings panel is required')
assert.ok(existsSync(resultSettingsUrl), 'survey designer result settings is required')
assert.ok(existsSync(questionPropertiesUrl), 'survey designer question properties panel is required')
assert.ok(existsSync(sidebarUrl), 'survey designer sidebar is required')
assert.ok(existsSync(formulaDialogUrl), 'survey designer formula dialog is required')
assert.ok(existsSync(textImportDialogUrl), 'survey designer text import dialog is required')
assert.ok(existsSync(bankUploadDialogUrl), 'survey designer bank upload dialog is required')
assert.ok(existsSync(presetDialogUrl), 'survey designer preset dialog is required')
assert.ok(existsSync(responsesUrl), 'survey responses composable is required')
assert.ok(existsSync(questionsUrl), 'survey designer questions composable is required')
assert.ok(existsSync(transferUrl), 'survey designer transfer composable is required')
assert.ok(existsSync(documentUrl), 'survey designer document composable is required')
assert.ok(existsSync(responsiveUrl), 'survey designer responsive stylesheet is required')
assert.ok(existsSync(designerStyleUrl), 'survey designer scoped stylesheet is required')
assert.ok(existsSync(textTransferUrl), 'survey designer text transfer module is required')

const designer = readFileSync(designerUrl, 'utf8')
const history = readFileSync(historyUrl, 'utf8')
const validation = readFileSync(validationUrl, 'utf8')
const issues = readFileSync(issuesUrl, 'utf8')
const outline = readFileSync(outlineUrl, 'utf8')
const panelTitle = readFileSync(panelTitleUrl, 'utf8')
const palette = readFileSync(paletteUrl, 'utf8')
const responsesPanel = readFileSync(responsesPanelUrl, 'utf8')
const reportPanel = readFileSync(reportPanelUrl, 'utf8')
const logicPanel = readFileSync(logicPanelUrl, 'utf8')
const settingsPanel = readFileSync(settingsPanelUrl, 'utf8')
const resultSettings = readFileSync(resultSettingsUrl, 'utf8')
const questionProperties = readFileSync(questionPropertiesUrl, 'utf8')
const sidebar = readFileSync(sidebarUrl, 'utf8')
const formulaDialog = readFileSync(formulaDialogUrl, 'utf8')
const textImportDialog = readFileSync(textImportDialogUrl, 'utf8')
const bankUploadDialog = readFileSync(bankUploadDialogUrl, 'utf8')
const presetDialog = readFileSync(presetDialogUrl, 'utf8')
const responses = readFileSync(responsesUrl, 'utf8')
const questions = readFileSync(questionsUrl, 'utf8')
const transfer = readFileSync(transferUrl, 'utf8')
const document = readFileSync(documentUrl, 'utf8')
const draggable = readFileSync(draggableUrl, 'utf8')
const responsive = readFileSync(responsiveUrl, 'utf8')
const designerStyle = readFileSync(designerStyleUrl, 'utf8')
const questionTypes = readFileSync(questionTypesUrl, 'utf8')
const textTransfer = readFileSync(textTransferUrl, 'utf8')

for (const pattern of [
  ':disabled="!canUndo"',
  '@click="undoDesignerChange"',
  ':disabled="!canRedo"',
  '@click="redoDesignerChange"',
  '@click="showShortcutHelp"',
  'class="toolbar-save-status"',
  "import SurveyDesignerIssues from './components/SurveyDesignerIssues.vue'",
  "import SurveyDesignerSidebar from './components/SurveyDesignerSidebar.vue'",
  "import SurveyDesignerResponsesPanel from './components/SurveyDesignerResponsesPanel.vue'",
  "import SurveyDesignerReportPanel from './components/SurveyDesignerReportPanel.vue'",
  "import SurveyDesignerLogicPanel from './components/SurveyDesignerLogicPanel.vue'",
  "import SurveyDesignerSettingsPanel from './components/SurveyDesignerSettingsPanel.vue'",
  "import SurveyDesignerResultSettings from './components/SurveyDesignerResultSettings.vue'",
  "import SurveyDesignerQuestionProperties from './components/SurveyDesignerQuestionProperties.vue'",
  "import { useSurveyDesignerQuestions } from './composables/useSurveyDesignerQuestions'",
  "import { useSurveyDesignerTransfer } from './composables/useSurveyDesignerTransfer'",
  "import { createSurveyDesignerForm, useSurveyDesignerDocument } from './composables/useSurveyDesignerDocument'",
  "import SurveyDesignerFormulaDialog from './components/SurveyDesignerFormulaDialog.vue'",
  "import SurveyDesignerTextImportDialog from './components/SurveyDesignerTextImportDialog.vue'",
  "import SurveyDesignerBankUploadDialog from './components/SurveyDesignerBankUploadDialog.vue'",
  "getSurveyQuestionTypeName as typeName",
  "from './survey-designer-validation'",
  '<SurveyDesignerIssues',
  '<SurveyDesignerSidebar',
  '<SurveyDesignerResponsesPanel',
  '<SurveyDesignerReportPanel',
  '<SurveyDesignerLogicPanel',
  '<SurveyDesignerSettingsPanel',
  '<SurveyDesignerResultSettings',
  '<SurveyDesignerQuestionProperties',
  '<SurveyDesignerFormulaDialog',
  '<SurveyDesignerTextImportDialog',
  '<SurveyDesignerBankUploadDialog',
  ':invalid-ids="invalidQuestionIds"',
  'const designerIssues = computed(',
  'function locateDesignerIssue(',
  'v-loading="designerLoading"',
  'v-if="designerLoadError"',
  '@click="retryLoad"',
  '<style scoped src="./survey-designer.css"></style>',
]) {
  assert.ok(designer.includes(pattern), `SurveyDesigner.vue missing UX contract: ${pattern}`)
}

assert.doesNotMatch(designer, /function typeName\(/, 'question type labels must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /tabsBarRef/, 'question palette scrolling must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function parseTextToQuestions\(/, 'text parsing must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function surveyToText\(/, 'text serialization must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function loadResponses\(/, 'response loading must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function loadReport\(/, 'report loading must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /const ruleForm = ref/, 'logic rule editor state must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function renderRuleDSL\(/, 'logic DSL rendering must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function loadAdminTree\(/, 'settings organization tree loading must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function addQuestion\(/, 'question state and operations must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function syncLogicRulesAfterQuestionChange\(/, 'question rule remapping must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function importTextQuestions\(/, 'text transfer orchestration must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function addFromBank\(/, 'question bank orchestration must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function exportSurveyKingJson\(/, 'SurveyKing transfer must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function buildSettingsPayload\(/, 'document serialization must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /function scheduleAutoSave\(/, 'auto-save orchestration must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /async function load\(/, 'document loading must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /async function save\(/, 'document persistence must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /class="survey-setting-panel"/, 'question properties implementation must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /class="logic-panel logic-full-panel"/, 'result settings implementation must stay outside SurveyDesigner.vue')
assert.doesNotMatch(designer, /class="survey-sidebar-panel"/, 'sidebar implementation must stay outside SurveyDesigner.vue')
for (const obsoleteDialogState of ['showFormulaDialog', 'showTextImportDialog', 'bankDialog', 'showSavePresetDialog']) {
  assert.ok(!designer.includes(obsoleteDialogState), `${obsoleteDialogState} must stay outside SurveyDesigner.vue`)
}
assert.ok(designerStyle.includes('.survey-main {'), 'survey designer layout styles must live in the scoped stylesheet')
assert.ok(questionTypes.includes('export function getSurveyQuestionTypeName'), 'question type label helper is required')
assert.ok(logicPanel.includes('defineModel<SurveyLogicRule[]>'), 'logic panel must expose typed v-model rules')
assert.ok(logicPanel.includes("emit('save'"), 'logic panel must delegate persistence to the designer')
assert.ok(settingsPanel.includes('defineModel<SurveyCompletionAction>'), 'settings panel must expose completion action v-model')
assert.ok(logicPanel.includes("import SurveyDesignerPresetDialog from './SurveyDesignerPresetDialog.vue'"), 'logic panel must own preset dialog')
assert.ok(logicPanel.includes('<SurveyDesignerPresetDialog'), 'logic panel must render preset dialog')
assert.ok(settingsPanel.includes('adminApi.deptTree()'), 'settings panel must own department loading')
assert.ok(resultSettings.includes('form.resultConfig.statType'), 'result settings must own statistic fields')
assert.ok(sidebar.includes('adminApi.surveyQuestionBankList'), 'sidebar must own question bank loading')
assert.ok(sidebar.includes('adminApi.surveyResourceList'), 'sidebar must own appearance resource loading')
assert.ok(questionProperties.includes("import SurveyDesignerPanelTitle from './SurveyDesignerPanelTitle.vue'"), 'question properties must own its panel title')

for (const pattern of [
  'defineModel<SurveyDesignerQuestion | null>',
  "defineModel<number>('selectedOptIdx'",
  "emit('remove')",
  "emit('openFormula'",
  'class="survey-setting-panel"',
]) {
  assert.ok(questionProperties.includes(pattern), `survey question properties missing contract: ${pattern}`)
}

for (const pattern of [
  'export function useSurveyDesignerQuestions',
  'function addQuestion(',
  'function replaceQuestions(',
  'function syncLogicRulesAfterQuestionChange(',
  'function removeSelected(',
]) {
  assert.ok(questions.includes(pattern), `survey question state missing contract: ${pattern}`)
}

for (const pattern of [
  'export function useSurveyDesignerTransfer',
  'function importTextQuestions(',
  'function addFromBank(',
  'function confirmUploadBank(',
  'function exportSurveyKingJson(',
  'const publicUrl = computed(',
]) {
  assert.ok(transfer.includes(pattern), `survey designer transfer composable missing contract: ${pattern}`)
}

for (const pattern of [
  'export function createSurveyDesignerForm',
  'export function useSurveyDesignerDocument',
  'function buildSettingsPayload(',
  'function scheduleAutoSave(',
  'async function load(',
  'async function save(',
  'onBeforeRouteLeave(',
  "window.addEventListener('beforeunload', handleBeforeUnload)",
  'useSurveyDesignerHistory(',
]) {
  assert.ok(document.includes(pattern), `survey designer document composable missing contract: ${pattern}`)
}

for (const pattern of [
  'export interface ParsedSurveyTextQuestion',
  'export function parseSurveyText',
  'export function serializeSurveyText',
  'export function stripSurveyMarkup',
]) {
  assert.ok(textTransfer.includes(pattern), `survey text transfer missing contract: ${pattern}`)
}

for (const pattern of ['useSurveyResponses(', 'loadResponses', 'deleteResponse', 'batchDeleteResponses']) {
  assert.ok(responses.includes(pattern), `survey responses composable missing contract: ${pattern}`)
}

for (const [source, name, patterns] of [
  [responsesPanel, 'responses panel', ['useSurveyResponses(', 'el-pagination', 'responseKeyword']],
  [reportPanel, 'report panel', ['surveyStatistic', 'chart-type-btns', 'fieldStats']],
  [formulaDialog, 'formula dialog', ['defineExpose', 'insertFormulaTag', "emit('save'"]],
  [textImportDialog, 'text import dialog', ['parseSurveyText', 'defineExpose', "emit('import'"]],
  [bankUploadDialog, 'bank upload dialog', ['defineExpose', "emit('confirm'"]],
  [presetDialog, 'preset dialog', ['defineExpose', "emit('save'"]],
]) {
  for (const pattern of patterns) {
    assert.ok(source.includes(pattern), `survey designer ${name} missing contract: ${pattern}`)
  }
}

for (const pattern of [
  'const categories = computed(',
  'const activeCategory = ref(',
  "emit('add', type)",
  'aria-label="上一个题型分类"',
  'aria-label="下一个题型分类"',
]) {
  assert.ok(palette.includes(pattern), `survey question palette missing contract: ${pattern}`)
}

assert.doesNotMatch(designer, /class="outline-tree"/, 'outline implementation must stay outside SurveyDesigner.vue')

for (const pattern of [
  'v-model="keyword"',
  'v-model="onlyIssues"',
  'const filteredChildren = computed(',
  "emit('select', questionId)",
  '没有匹配题目',
  'WarningFilled',
]) {
  assert.ok(outline.includes(pattern), `survey outline navigation missing contract: ${pattern}`)
}

for (const pattern of [
  'class="props-panel-title"',
  'class="props-panel-title-context"',
  'position:sticky',
]) {
  assert.ok(panelTitle.includes(pattern), `survey property panel title missing contract: ${pattern}`)
}

for (const pattern of [':class="{ selected: q.id === selectedId, invalid:', 'class="card-invalid-label"']) {
  assert.ok(draggable.includes(pattern), `survey question error state missing contract: ${pattern}`)
}

for (const pattern of ['@media (min-width: 768px) and (max-width: 1199px)', '.survey-editor:not(.has-open-settings)', '@media (max-width: 767px)']) {
  assert.ok(responsive.includes(pattern), `survey responsive layout missing contract: ${pattern}`)
}

for (const pattern of [
  'export interface SurveyDesignerIssue',
  'export function collectSurveyDesignerIssues',
  "'question_title_required'",
  "'question_id_duplicate'",
  "'question_options_required'",
]) {
  assert.ok(validation.includes(pattern), `survey validation missing contract: ${pattern}`)
}

for (const pattern of [
  "defineEmits<{ locate: [issue: SurveyDesignerIssue] }>()",
  "emit('locate', issue)",
  'class="designer-issues-summary"',
  '@media (max-width: 767px)',
]) {
  assert.ok(issues.includes(pattern), `survey issues component missing contract: ${pattern}`)
}

assert.doesNotMatch(designer, /class="toolbar-btn" disabled/)

for (const pattern of [
  'export function useSurveyDesignerHistory',
  'const canUndo = computed(',
  'const canRedo = computed(',
  'function reset(',
  'function record(',
  'function undo(',
  'function redo(',
  'function dispose(',
]) {
  assert.ok(history.includes(pattern), `history composable missing contract: ${pattern}`)
}

console.log('survey designer UX checks passed')
