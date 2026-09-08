import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'

const designerUrl = new URL('../src/views/survey/SurveyDesigner.vue', import.meta.url)
const historyUrl = new URL('../src/views/survey/composables/useSurveyDesignerHistory.ts', import.meta.url)
const validationUrl = new URL('../src/views/survey/survey-designer-validation.ts', import.meta.url)
const issuesUrl = new URL('../src/views/survey/components/SurveyDesignerIssues.vue', import.meta.url)
const draggableUrl = new URL('../src/views/survey/formkit/DraggableList.vue', import.meta.url)
const responsiveUrl = new URL('../src/views/survey/survey-designer-responsive.css', import.meta.url)

assert.ok(existsSync(historyUrl), 'survey designer history composable is required')
assert.ok(existsSync(validationUrl), 'survey designer validation module is required')
assert.ok(existsSync(issuesUrl), 'survey designer issues component is required')
assert.ok(existsSync(responsiveUrl), 'survey designer responsive stylesheet is required')

const designer = readFileSync(designerUrl, 'utf8')
const history = readFileSync(historyUrl, 'utf8')
const validation = readFileSync(validationUrl, 'utf8')
const issues = readFileSync(issuesUrl, 'utf8')
const draggable = readFileSync(draggableUrl, 'utf8')
const responsive = readFileSync(responsiveUrl, 'utf8')

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
  "from './survey-designer-validation'",
  '<SurveyDesignerIssues',
  ':invalid-ids="invalidQuestionIds"',
  'const designerIssues = computed(',
  'function locateDesignerIssue(',
  'v-loading="designerLoading"',
  'v-if="designerLoadError"',
  '@click="retryLoad"',
]) {
  assert.ok(designer.includes(pattern), `SurveyDesigner.vue missing UX contract: ${pattern}`)
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
