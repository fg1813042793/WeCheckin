import assert from 'node:assert/strict'
import './check-exam-unscored-runtime.mjs'
import { existsSync, readFileSync } from 'node:fs'

const urls = {
  designer: new URL('../src/views/exam/ExamDesigner.vue', import.meta.url),
  sidebar: new URL('../src/views/exam/components/ExamDesignerSidebar.vue', import.meta.url),
  logic: new URL('../src/views/exam/components/ExamDesignerLogicPanel.vue', import.meta.url),
  examPostStat: new URL('../src/views/exam/components/ExamPostStatSettings.vue', import.meta.url),
  settings: new URL('../src/views/exam/components/ExamDesignerSettingsPanel.vue', import.meta.url),
  document: new URL('../src/views/exam/composables/useExamDesignerDocument.ts', import.meta.url),
  transfer: new URL('../src/views/exam/composables/useExamDesignerTransfer.ts', import.meta.url),
  questionPreview: new URL('../src/views/exam/formkit/QuestionPreview.vue', import.meta.url),
  questionField: new URL('../src/views/survey/components/QuestionField.vue', import.meta.url),
  fillPC1: new URL('../src/views/exam/ExamFillPC1.vue', import.meta.url),
  styles: new URL('../src/views/exam/exam-designer.css', import.meta.url),
}

for (const [name, url] of Object.entries(urls)) {
  assert.ok(existsSync(url), `exam designer ${name} module is required`)
}

const designer = readFileSync(urls.designer, 'utf8')
const sidebar = readFileSync(urls.sidebar, 'utf8')
const logic = readFileSync(urls.logic, 'utf8')
const examPostStat = readFileSync(urls.examPostStat, 'utf8')
const settings = readFileSync(urls.settings, 'utf8')
const document = readFileSync(urls.document, 'utf8')
const transfer = readFileSync(urls.transfer, 'utf8')
const questionPreview = readFileSync(urls.questionPreview, 'utf8')
const questionField = readFileSync(urls.questionField, 'utf8')
const fillPC1 = readFileSync(urls.fillPC1, 'utf8')
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
for (const pattern of ['defineModel<ExamLogicRule[]>', 'function renderRuleDSL(', "emit('save'", 'value="postStat"', 'ExamPostStatSettings', "conditionType = 'none'"]) {
  assert.ok(logic.includes(pattern), `exam designer logic panel missing contract: ${pattern}`)
}
assert.ok(designer.includes('<el-option value="none" label="不计分" />'), 'unscored mode must be explicit')
for (const pattern of ['是否计算总分', 'calculateTotalScore', 'notifyChannel', 'webhookUrl', 'messageTemplate', '通知对象', '模版预设', 'examNotificationTemplatesGet', '{answerDownloadUrl}']) {
  assert.ok(examPostStat.includes(pattern), `exam post-stat settings missing contract: ${pattern}`)
}
for (const pattern of ["defineModel<Record<string, any>>('form'", 'function onAdminTreeCheck(', 'async function loadData(', 'async function loadDepartments(', 'async function loadManagers(', 'defineExpose({ loadData })']) {
  assert.ok(settings.includes(pattern), `exam designer settings panel missing contract: ${pattern}`)
}
for (const pattern of [
  "@update:model-value=\"emit('update:publicUrl', $event)\"",
  "@click=\"emit('downloadQr')\"",
  "'update:publicUrl': [value: string]",
  'downloadQr: []',
  'label="及格分"',
  'form.passScore',
  `<el-form-item v-if="completionAction==='content'" class="settings-full" label="提交完成页内容">`,
  `<el-form-item v-if="completionAction==='redirect'" class="settings-full" label="提交后跳转链接">`,
  'label="答卷下载格式"',
  'form.answerDownloadFormat',
  'label="下载链接有效期"',
  'form.answerDownloadExpireDays',
  'label="答卷下载服务地址"',
  'form.answerDownloadBaseUrl',
]) {
  assert.ok(settings.includes(pattern), `exam designer share settings missing contract: ${pattern}`)
}
let previousExamRuleIndex = -1
for (const pattern of [
  'label="每人考试次数"',
  'label="及格分"',
  'label="随机顺序"',
  'label="开始时间"',
  'label="结束时间"',
  'class="settings-full" label="练习模式"',
]) {
  const index = settings.indexOf(pattern)
  assert.ok(index > previousExamRuleIndex, `exam rule layout order missing contract: ${pattern}`)
  previousExamRuleIndex = index
}
for (const pattern of ['export function createExamDesignerForm', 'export function useExamDesignerDocument', 'function scheduleAutoSave(', 'async function load(', 'async function save(']) {
  assert.ok(document.includes(pattern), `exam designer document composable missing contract: ${pattern}`)
}
assert.ok(document.includes("await save(true, '逻辑规则已保存')"), 'logic save failures must remain visible')
for (const pattern of [
  "id: 0, title: '', displayTitle: '', publicUrl: ''",
  'passScore: 0',
  "answerDownloadFormat: 'pdf', answerDownloadExpireDays: 7",
  "answerDownloadBaseUrl: ''",
  'publicUrl: typeof form.publicUrl === \'string\' ? form.publicUrl.trim() : \'\'',
  'passScore: form.passScore',
  'answerDownloadFormat: form.answerDownloadFormat',
  'answerDownloadExpireDays: form.answerDownloadExpireDays',
  "const answerDownloadBaseUrl = typeof form.answerDownloadBaseUrl === 'string'",
  "form.answerDownloadBaseUrl.trim().replace(/\\/+$/, '')",
  'answerDownloadBaseUrl,',
  'passScore: survey.passScore ?? 0',
  "publicUrl: typeof settings.publicUrl === 'string' ? settings.publicUrl : ''",
  "answerDownloadFormat: settings.answerDownloadFormat || 'pdf'",
  'answerDownloadExpireDays: settings.answerDownloadExpireDays || 7',
  "answerDownloadBaseUrl: typeof settings.answerDownloadBaseUrl === 'string' ? settings.answerDownloadBaseUrl : ''",
  "messageTemplate || '').includes('{answerDownloadUrl}')",
  '请先配置答卷下载服务地址',
  'logicRules: options.logicRules.value',
  "options.logicRules.value = JSON.parse(settings.logicRules || '[]')",
]) {
  assert.ok(document.includes(pattern), `exam public URL persistence missing contract: ${pattern}`)
}
for (const pattern of ['export function useExamDesignerTransfer', 'function importTextQuestions(', 'function addFromBank(', 'function confirmUploadBank(']) {
  assert.ok(transfer.includes(pattern), `exam designer transfer composable missing contract: ${pattern}`)
}
for (const pattern of [
  'const generatedPublicUrl = computed(',
  'const publicUrl = computed({',
  'const qrUrl = computed(',
  'async function copyLink()',
  'async function downloadQR()',
]) {
  assert.ok(transfer.includes(pattern), `exam share transfer missing contract: ${pattern}`)
}
for (const pattern of [
  'v-model:public-url="publicUrl"',
  ':qr-url="qrUrl"',
  ':total-score="paperTotalScore"',
  '@download-qr="downloadQR"',
]) {
  assert.ok(designer.includes(pattern), `ExamDesigner.vue missing exam share binding: ${pattern}`)
}
for (const pattern of [
  `v-if="q.type==='divider' && dividerTitleHtml"`,
  'class="q-divider-title"',
  'v-html="dividerTitleHtml"',
  "const dividerTitleHtml = computed(() => unwrapOuterP(props.q.title || ''))",
]) {
  assert.ok(questionField.includes(pattern), `exam ef1 divider title missing contract: ${pattern}`)
}
assert.ok(
  questionField.includes(`<span v-else-if="q.type==='pagination'" class="q-type-label">分页</span>`),
  'pagination layout label must remain explicit when divider label is hidden',
)
assert.doesNotMatch(
  questionField,
  /q\.type==='divider'[^>]*class="q-type-label">分隔<\/span>/,
  'public fill pages must not show the divider type label',
)
for (const pattern of [
  "import('@/components/rich-text/AdvancedRichTextEditor.vue')",
  '<AdvancedRichTextEditor v-model="titleEditBuffer"',
  '<AdvancedRichTextEditor v-model="optionEditBuffer"',
]) {
  assert.ok(questionPreview.includes(pattern), `exam advanced rich editor missing contract: ${pattern}`)
}
for (const pattern of ['fullTitleEditorOptions', 'onFullEditorReady', 'onOptionFullEditorReady']) {
  assert.ok(!questionPreview.includes(pattern), `exam advanced rich editor still contains legacy implementation: ${pattern}`)
}
assert.ok(styles.includes('.survey-main {'), 'exam designer layout styles are required')
for (const pattern of ['resultForQuestion(i)', 'function resultForQuestion(index: number)']) {
  assert.ok(fillPC1.includes(pattern), `exam result mapping missing contract: ${pattern}`)
}
for (const pattern of [
  '<el-option value="manual" label="人工评分题" />',
  "selected.examScoreMode==='manual'",
  'label="参考答案（选填）"',
  "q.examScoreMode = t.type === 'textarea' ? 'manual' : 'single'",
]) {
  assert.ok(designer.includes(pattern), `exam manual scoring designer contract missing: ${pattern}`)
}

console.log('exam designer UX checks passed')
