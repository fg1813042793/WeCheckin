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
  'components/rich-text/AdvancedRichTextEditor.vue',
  'components/rich-text/RichTextToolbar.vue',
  'components/rich-text/richTextConfig.ts',
  'components/rich-text/sanitizeRichText.ts',
  'components/rich-text/rich-text-content.css',
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
  "hasPerm('admin:menu:bank:add')",
  "hasPerm('admin:menu:bank:edit')",
  "hasPerm('admin:menu:bank:del')",
  "hasPerm('admin:menu:question-bank:add')",
  "hasPerm('admin:menu:question-bank:edit')",
  "hasPerm('admin:menu:question-bank:del')",
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
requireSnippet(canvas, '.bank-rich-editor:focus-within', 'question editor canvas')
requireSnippet(canvas, 'box-shadow:none;', 'question editor canvas')

const editorDialog = read('views/question-bank/components/QuestionEditorDialog.vue')
requireSnippet(editorDialog, '.bank-rich-editor:focus-within', 'question editor dialog')
requireSnippet(editorDialog, 'box-shadow: none;', 'question editor dialog')

const advancedEditor = read('components/rich-text/AdvancedRichTextEditor.vue')
for (const snippet of [
  'RichTextToolbar',
  'sanitizeRichText',
  "getModule('table')",
  'insertTable',
  'sourceMode',
  'toggleFullScreen',
  'clearFormatting',
]) {
  requireSnippet(advancedEditor, snippet, 'advanced rich text editor')
}

const richTextToolbar = read('components/rich-text/RichTextToolbar.vue')
for (const snippet of [
  'rich-text-toolbar__row--primary',
  'rich-text-toolbar__row--secondary',
  'toolbar-color-control',
  '清除格式',
  '代码块',
  '插入表格',
  '源码',
  'aria-pressed',
  'is-active',
]) {
  requireSnippet(richTextToolbar, snippet, 'rich text toolbar')
}
forbidSnippet(richTextToolbar, 'toolbar-button__mini-label', 'rich text toolbar')

const sanitizer = read('components/rich-text/sanitizeRichText.ts')
for (const snippet of ['DOMPurify', 'FORBID_TAGS', 'FORBID_ATTR', 'serializeRichText', "querySelectorAll('.ql-ui')"]) {
  requireSnippet(sanitizer, snippet, 'rich text sanitizer')
}

const richTextConfig = read('components/rich-text/richTextConfig.ts')
for (const snippet of ['highlight.js/lib/core', 'Quill.register', 'RICH_TEXT_LANGUAGES', 'RICH_TEXT_FONT_OPTIONS', 'RICH_TEXT_SIZE_OPTIONS', 'syntax:', 'imageResize:']) {
  requireSnippet(richTextConfig, snippet, 'rich text config')
}

const richTextStyles = read('components/rich-text/rich-text-content.css')
for (const snippet of ['.ql-code-block-container', '.ql-code-block', 'content: "代码"', 'background: #111827', 'box-shadow: inset 0 32px', 'table', 'blockquote']) {
  requireSnippet(richTextStyles, snippet, 'rich text content styles')
}

const richEditorDialog = read('views/question-bank/components/QuestionRichEditorDialog.vue')
for (const snippet of [
  'defineAsyncComponent',
  "import('@/components/rich-text/AdvancedRichTextEditor.vue')",
  '<AdvancedRichTextEditor',
]) {
  requireSnippet(richEditorDialog, snippet, 'question rich editor dialog')
}

const properties = read('views/question-bank/components/QuestionPropertiesPanel.vue')
for (const snippet of ['题目属性', '所属题库', "activeScope === 'exam'", '正确答案', '答案解析']) {
  requireSnippet(properties, snippet, 'question properties panel')
}
