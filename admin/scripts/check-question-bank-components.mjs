import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')

function read(path) {
  return readFileSync(resolve(srcDir, path), 'utf8')
}

function requireSnippet(source, snippet, label) {
  if (!source.includes(snippet)) {
    throw new Error(`${label} missing required snippet: ${snippet}`)
  }
}

function forbidSnippet(source, snippet, label) {
  if (source.includes(snippet)) {
    throw new Error(`${label} still contains old inline snippet: ${snippet}`)
  }
}

function requireFile(path) {
  const fullPath = resolve(srcDir, path)
  if (!existsSync(fullPath)) {
    throw new Error(`missing component file: ${path}`)
  }
}

const page = read('views/question-bank/QuestionBank.vue')

for (const componentPath of [
  'views/question-bank/components/QuestionBankNavigator.vue',
  'views/question-bank/components/QuestionEditorCanvas.vue',
  'views/question-bank/components/QuestionPropertiesPanel.vue',
  'views/question-bank/components/QuestionRichEditorDialog.vue',
]) {
  requireFile(componentPath)
}

for (const snippet of [
  "import QuestionBankNavigator from './components/QuestionBankNavigator.vue'",
  "import QuestionEditorCanvas from './components/QuestionEditorCanvas.vue'",
  "import QuestionPropertiesPanel from './components/QuestionPropertiesPanel.vue'",
  '<QuestionBankNavigator',
  '<QuestionEditorCanvas',
  '<QuestionPropertiesPanel',
  'question-bank-workspace',
  'question-bank-mobile-notice',
  'saveStatusLabel',
  'scheduleAutoSave',
  'flushPendingSave',
  'handleWorkspaceKeydown',
]) {
  requireSnippet(page, snippet, 'question bank page')
}

for (const snippet of [
  'class="question-edit-dialog"',
  "import QuestionBankTable from './components/QuestionBankTable.vue'",
  "import QuestionEditorDialog from './components/QuestionEditorDialog.vue'",
  "import QuestionPreviewDrawer from './components/QuestionPreviewDrawer.vue'",
  '<QuestionBankTable',
  '<QuestionEditorDialog',
  '<QuestionPreviewDrawer',
]) {
  forbidSnippet(page, snippet, 'question bank page')
}

const navigator = read('views/question-bank/components/QuestionBankNavigator.vue')
for (const snippet of ['全部题目', '未分类', "emit('select-question'", "emit('add-question'", 'question-bank-navigator__list']) {
  requireSnippet(navigator, snippet, 'question bank navigator')
}

const canvas = read('views/question-bank/components/QuestionEditorCanvas.vue')
for (const snippet of ['编辑', '预览', 'JSON', 'QuestionRichEditorDialog', "emit('add-question'"]) {
  requireSnippet(canvas, snippet, 'question editor canvas')
}

const properties = read('views/question-bank/components/QuestionPropertiesPanel.vue')
for (const snippet of ['题目属性', '所属题库', "activeScope === 'exam'", '正确答案', '答案解析']) {
  requireSnippet(properties, snippet, 'question properties panel')
}
