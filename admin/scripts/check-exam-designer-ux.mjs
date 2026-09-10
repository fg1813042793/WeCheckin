import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'

const urls = {
  designer: new URL('../src/views/exam/ExamDesigner.vue', import.meta.url),
  sidebar: new URL('../src/views/exam/components/ExamDesignerSidebar.vue', import.meta.url),
  logic: new URL('../src/views/exam/components/ExamDesignerLogicPanel.vue', import.meta.url),
  settings: new URL('../src/views/exam/components/ExamDesignerSettingsPanel.vue', import.meta.url),
  document: new URL('../src/views/exam/composables/useExamDesignerDocument.ts', import.meta.url),
  transfer: new URL('../src/views/exam/composables/useExamDesignerTransfer.ts', import.meta.url),
  styles: new URL('../src/views/exam/exam-designer.css', import.meta.url),
}

for (const [name, url] of Object.entries(urls)) {
  assert.ok(existsSync(url), `exam designer ${name} module is required`)
}

const designer = readFileSync(urls.designer, 'utf8')
const sidebar = readFileSync(urls.sidebar, 'utf8')
const logic = readFileSync(urls.logic, 'utf8')
const settings = readFileSync(urls.settings, 'utf8')
const document = readFileSync(urls.document, 'utf8')
const transfer = readFileSync(urls.transfer, 'utf8')
const styles = readFileSync(urls.styles, 'utf8')

for (const pattern of [
  "import ExamDesignerSidebar from './components/ExamDesignerSidebar.vue'",
  "import ExamDesignerLogicPanel from './components/ExamDesignerLogicPanel.vue'",
  "import ExamDesignerSettingsPanel from './components/ExamDesignerSettingsPanel.vue'",
  "import { createExamDesignerForm, useExamDesignerDocument } from './composables/useExamDesignerDocument'",
  "import { useExamDesignerTransfer } from './composables/useExamDesignerTransfer'",
  "import SurveyDesignerFormulaDialog from '../survey/components/SurveyDesignerFormulaDialog.vue'",
  "import SurveyDesignerTextImportDialog from '../survey/components/SurveyDesignerTextImportDialog.vue'",
  "import SurveyDesignerBankUploadDialog from '../survey/components/SurveyDesignerBankUploadDialog.vue'",
  '<ExamDesignerSidebar',
  '<ExamDesignerLogicPanel',
  '<ExamDesignerSettingsPanel',
  '<style scoped src="./exam-designer.css"></style>',
]) {
  assert.ok(designer.includes(pattern), `ExamDesigner.vue missing contract: ${pattern}`)
}

for (const [pattern, message] of [
  [/function parseTextToQuestions\(/, 'text parsing must stay outside ExamDesigner.vue'],
  [/function loadBank\(/, 'question bank loading must stay outside ExamDesigner.vue'],
  [/function renderRuleDSL\(/, 'logic DSL rendering must stay outside ExamDesigner.vue'],
  [/async function load\(/, 'document loading must stay outside ExamDesigner.vue'],
  [/async function save\(/, 'document persistence must stay outside ExamDesigner.vue'],
  [/<style scoped>/, 'exam designer styles must stay outside ExamDesigner.vue'],
]) {
  assert.doesNotMatch(designer, pattern, message)
}

for (const pattern of ['examQuestionBankList', 'examResourceList', 'defineExpose', "emit('addBank'", ":class=\"{ compact: middleTab === 'logic' }\""]) {
  assert.ok(sidebar.includes(pattern), `exam designer sidebar missing contract: ${pattern}`)
}
for (const pattern of ['defineModel<ExamLogicRule[]>', 'function renderRuleDSL(', "emit('save'"]) {
  assert.ok(logic.includes(pattern), `exam designer logic panel missing contract: ${pattern}`)
}
for (const pattern of ["defineModel<Record<string, any>>('form'", 'function onAdminTreeCheck(', 'async function loadData(', 'async function loadDepartments(', 'async function loadManagers(', 'defineExpose({ loadData })']) {
  assert.ok(settings.includes(pattern), `exam designer settings panel missing contract: ${pattern}`)
}
for (const pattern of ['export function createExamDesignerForm', 'export function useExamDesignerDocument', 'function scheduleAutoSave(', 'async function load(', 'async function save(']) {
  assert.ok(document.includes(pattern), `exam designer document composable missing contract: ${pattern}`)
}
assert.ok(document.includes("await save(true, '逻辑规则已保存')"), 'logic save failures must remain visible')
for (const pattern of ['export function useExamDesignerTransfer', 'function importTextQuestions(', 'function addFromBank(', 'function confirmUploadBank(']) {
  assert.ok(transfer.includes(pattern), `exam designer transfer composable missing contract: ${pattern}`)
}
assert.ok(styles.includes('.survey-main {'), 'exam designer layout styles are required')

console.log('exam designer UX checks passed')
