import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'

const designerUrl = new URL('../src/views/survey/SurveyDesigner.vue', import.meta.url)
const historyUrl = new URL('../src/views/survey/composables/useSurveyDesignerHistory.ts', import.meta.url)

assert.ok(existsSync(historyUrl), 'survey designer history composable is required')

const designer = readFileSync(designerUrl, 'utf8')
const history = readFileSync(historyUrl, 'utf8')

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
]) {
  assert.ok(designer.includes(pattern), `SurveyDesigner.vue missing UX contract: ${pattern}`)
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
