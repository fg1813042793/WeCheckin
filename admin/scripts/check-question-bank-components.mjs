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
const editorDialog = read('views/question-bank/components/QuestionEditorDialog.vue')

for (const componentPath of [
  'views/question-bank/components/QuestionBankTable.vue',
  'views/question-bank/components/QuestionEditorDialog.vue',
  'views/question-bank/components/QuestionRichEditorDialog.vue',
  'views/question-bank/components/QuestionPreviewDrawer.vue',
]) {
  requireFile(componentPath)
}

for (const snippet of [
  "import QuestionBankTable from './components/QuestionBankTable.vue'",
  "import QuestionEditorDialog from './components/QuestionEditorDialog.vue'",
  "import QuestionPreviewDrawer from './components/QuestionPreviewDrawer.vue'",
  '<QuestionBankTable',
  '<QuestionEditorDialog',
  '<QuestionPreviewDrawer',
]) {
  requireSnippet(page, snippet, 'question bank page')
}

for (const snippet of [
  "import QuestionRichEditorDialog from './QuestionRichEditorDialog.vue'",
  '<QuestionRichEditorDialog',
]) {
  requireSnippet(editorDialog, snippet, 'question editor dialog')
}

for (const snippet of [
  'class="question-edit-dialog"',
  '<el-drawer v-model="preview.visible"',
  '<el-table :data="list"',
]) {
  forbidSnippet(page, snippet, 'question bank page')
}

forbidSnippet(editorDialog, 'class="bank-rich-full-dialog"', 'question editor dialog')
