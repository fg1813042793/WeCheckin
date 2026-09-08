import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'

const designerUrl = new URL('../src/views/survey/SurveyDesigner.vue', import.meta.url)
const historyUrl = new URL('../src/views/survey/composables/useSurveyDesignerHistory.ts', import.meta.url)
const validationUrl = new URL('../src/views/survey/survey-designer-validation.ts', import.meta.url)
const issuesUrl = new URL('../src/views/survey/components/SurveyDesignerIssues.vue', import.meta.url)
const outlineUrl = new URL('../src/views/survey/components/SurveyDesignerOutline.vue', import.meta.url)
const panelTitleUrl = new URL('../src/views/survey/components/SurveyDesignerPanelTitle.vue', import.meta.url)
const paletteUrl = new URL('../src/views/survey/components/SurveyDesignerQuestionPalette.vue', import.meta.url)
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
const draggable = readFileSync(draggableUrl, 'utf8')
const responsive = readFileSync(responsiveUrl, 'utf8')
const designerStyle = readFileSync(designerStyleUrl, 'utf8')
const questionTypes = readFileSync(questionTypesUrl, 'utf8')
const textTransfer = readFileSync(textTransferUrl, 'utf8')

for (const pattern of [
  "import { useSurveyDesignerHistory } from './composables/useSurveyDesignerHistory'",
  ':disabled="!canUndo"',
  '@click="undoDesignerChange"',
  ':disabled="!canRedo"',
  '@click="redoDesignerChange"',
  '@click="showShortcutHelp"',
  'class="toolbar-save-status"',
  'const designerSnapshot = computed(',
  'watch(designerSnapshot, scheduleAutoSave)',
  'onBeforeRouteLeave(',
  "window.addEventListener('beforeunload', handleBeforeUnload)",
  "import SurveyDesignerIssues from './components/SurveyDesignerIssues.vue'",
  "import SurveyDesignerOutline from './components/SurveyDesignerOutline.vue'",
  "import SurveyDesignerPanelTitle from './components/SurveyDesignerPanelTitle.vue'",
  "import SurveyDesignerQuestionPalette from './components/SurveyDesignerQuestionPalette.vue'",
  "getSurveyQuestionTypeName as typeName",
  "from './survey-designer-text-transfer'",
  "from './survey-designer-validation'",
  '<SurveyDesignerIssues',
  '<SurveyDesignerOutline',
  '<SurveyDesignerPanelTitle',
  '<SurveyDesignerQuestionPalette',
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
assert.ok(designerStyle.includes('.survey-main {'), 'survey designer layout styles must live in the scoped stylesheet')
assert.ok(questionTypes.includes('export function getSurveyQuestionTypeName'), 'question type label helper is required')

for (const pattern of [
  'export interface ParsedSurveyTextQuestion',
  'export function parseSurveyText',
  'export function serializeSurveyText',
  'export function stripSurveyMarkup',
]) {
  assert.ok(textTransfer.includes(pattern), `survey text transfer missing contract: ${pattern}`)
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
