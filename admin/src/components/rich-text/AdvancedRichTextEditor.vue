<script setup lang="ts">
import type Quill from 'quill'
import { QuillEditor } from '@vueup/vue-quill'
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, nextTick, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import '@/utils/quill-image-resize'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import RichTextToolbar from './RichTextToolbar.vue'
import { createRichTextModules } from './richTextConfig'
import { sanitizeRichText, serializeRichText } from './sanitizeRichText'
import './rich-text-content.css'

type FormatValue = string | number | boolean
type TableCommand = 'insert-row-above' | 'insert-row-below' | 'insert-column-left' | 'insert-column-right' | 'delete-row' | 'delete-column' | 'delete-table'

interface RichTextRange {
  index: number
  length: number
}

interface RichTextHistoryModule {
  stack?: {
    undo?: unknown[]
    redo?: unknown[]
  }
  undo: () => void
  redo: () => void
}

interface RichTextTableModule {
  insertTable: (rows: number, columns: number) => void
  insertRowAbove: () => void
  insertRowBelow: () => void
  insertColumnLeft: () => void
  insertColumnRight: () => void
  deleteRow: () => void
  deleteColumn: () => void
  deleteTable: () => void
}

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  minHeight?: number
}>(), {
  placeholder: '请输入富文本内容',
  minHeight: 360,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const quill = shallowRef<Quill | null>(null)
const content = ref(props.modelValue || '')
const activeFormats = ref<Record<string, unknown>>({})
const savedRange = ref<RichTextRange>({ index: 0, length: 0 })
const canUndo = ref(false)
const canRedo = ref(false)
const inTable = ref(false)
const sourceMode = ref(false)
const sourceDraft = ref('')
const fullscreen = ref(false)
const imageInput = ref<HTMLInputElement | null>(null)
let lastEmittedContent = ''

const editorOptions = computed(() => ({
  theme: 'snow',
  placeholder: props.placeholder,
  modules: createRichTextModules(),
}))

const editorStyle = computed(() => ({
  '--advanced-editor-min-height': `${Math.max(240, props.minHeight)}px`,
}))

watch(() => props.modelValue, (value) => {
  const next = String(value || '')
  if (next !== lastEmittedContent && next !== serializeRichText(content.value))
    content.value = next
})

watch(content, (value) => {
  const next = serializeRichText(String(value || ''))
  lastEmittedContent = next
  if (next !== props.modelValue)
    emit('update:modelValue', next)
})

function historyModule() {
  return quill.value?.getModule('history') as RichTextHistoryModule | undefined
}

function tableModule() {
  return quill.value?.getModule('table') as RichTextTableModule | undefined
}

function syncHistoryState() {
  const history = historyModule()
  canUndo.value = Boolean(history?.stack?.undo?.length)
  canRedo.value = Boolean(history?.stack?.redo?.length)
}

function syncActiveFormats(range?: RichTextRange | null) {
  const editor = quill.value
  if (!editor)
    return
  const currentRange = range || editor.getSelection() || savedRange.value
  if (currentRange)
    savedRange.value = { index: currentRange.index, length: currentRange.length }
  activeFormats.value = currentRange ? editor.getFormat(currentRange.index, currentRange.length) : {}
  inTable.value = Boolean(activeFormats.value.table)
  syncHistoryState()
}

function handleSelectionChange(range: RichTextRange | null) {
  if (range)
    syncActiveFormats(range)
}

function handleTextChange() {
  syncActiveFormats()
}

function handleReady(instance: Quill) {
  if (quill.value) {
    quill.value.off('selection-change', handleSelectionChange)
    quill.value.off('text-change', handleTextChange)
  }
  quill.value = instance
  instance.on('selection-change', handleSelectionChange)
  instance.on('text-change', handleTextChange)
  syncActiveFormats(instance.getSelection())
}

onBeforeUnmount(() => {
  quill.value?.off('selection-change', handleSelectionChange)
  quill.value?.off('text-change', handleTextChange)
})

function restoreSelection() {
  const editor = quill.value
  if (!editor)
    return null
  const length = Math.max(0, editor.getLength() - 1)
  const range = {
    index: Math.min(savedRange.value.index, length),
    length: Math.min(savedRange.value.length, Math.max(0, length - savedRange.value.index)),
  }
  editor.focus()
  editor.setSelection(range.index, range.length, 'silent')
  return range
}

function applyFormat(name: string, value: FormatValue) {
  const editor = quill.value
  if (!editor || sourceMode.value)
    return
  restoreSelection()
  editor.format(name, value, 'user')
  syncActiveFormats(editor.getSelection())
}

function clearFormatting() {
  const editor = quill.value
  if (!editor || sourceMode.value)
    return
  const range = restoreSelection()
  if (!range)
    return
  const formats = editor.getFormat(range.index, range.length)
  for (const name of Object.keys(formats)) {
    if (name !== 'table')
      editor.format(name, false, 'user')
  }
  syncActiveFormats(editor.getSelection())
}

function undo() {
  historyModule()?.undo()
  nextTick(syncHistoryState)
}

function redo() {
  historyModule()?.redo()
  nextTick(syncHistoryState)
}

function safeLink(value: string) {
  const url = value.trim()
  return /^(https?:\/\/|mailto:|tel:)/i.test(url) ? url : ''
}

async function editLink() {
  const editor = quill.value
  if (!editor || sourceMode.value)
    return
  const range = restoreSelection()
  if (!range)
    return
  const current = String(editor.getFormat(range.index, range.length).link || '')
  try {
    const result = await ElMessageBox.prompt('请输入以 http://、https://、mailto: 或 tel: 开头的地址', '插入链接', {
      inputValue: current,
      confirmButtonText: '应用',
      cancelButtonText: '取消',
      inputPlaceholder: 'https://example.com',
    })
    const url = safeLink(result.value)
    if (!url) {
      ElMessage.warning('请输入有效链接地址')
      return
    }
    if (range.length === 0) {
      editor.insertText(range.index, url, { link: url }, 'user')
      editor.setSelection(range.index + url.length, 0, 'silent')
    }
    else {
      editor.format('link', url, 'user')
    }
    syncActiveFormats(editor.getSelection())
  }
  catch {
    // Cancel keeps the current link unchanged.
  }
}

function insertImage(url: string) {
  const editor = quill.value
  const range = restoreSelection()
  if (!editor || !range)
    return
  editor.insertEmbed(range.index, 'image', url, 'user')
  editor.setSelection(range.index + 1, 0, 'silent')
}

async function insertImageUrl() {
  try {
    const result = await ElMessageBox.prompt('请输入图片的 http:// 或 https:// 地址', '插入图片', {
      confirmButtonText: '插入',
      cancelButtonText: '取消',
      inputPlaceholder: 'https://example.com/image.png',
    })
    const url = result.value.trim()
    if (!/^https?:\/\//i.test(url)) {
      ElMessage.warning('请输入有效图片地址')
      return
    }
    insertImage(url)
  }
  catch {
    // Cancel keeps the document unchanged.
  }
}

function handleImageCommand(type: 'file' | 'url') {
  if (type === 'url') {
    void insertImageUrl()
    return
  }
  imageInput.value?.click()
}

function handleImageFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file)
    return
  if (!file.type.startsWith('image/')) {
    ElMessage.warning('请选择图片文件')
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.warning('图片不能超过 5MB')
    return
  }
  const reader = new FileReader()
  reader.onload = () => insertImage(String(reader.result || ''))
  reader.onerror = () => ElMessage.error('图片读取失败')
  reader.readAsDataURL(file)
}

function insertTable(rows: number, columns: number) {
  const module = tableModule()
  if (!module || sourceMode.value)
    return
  restoreSelection()
  module.insertTable(Math.min(10, Math.max(1, rows)), Math.min(10, Math.max(1, columns)))
  syncActiveFormats(quill.value?.getSelection())
}

function runTableCommand(command: TableCommand) {
  const module = tableModule()
  if (!module || sourceMode.value)
    return
  restoreSelection()
  const actions: Record<TableCommand, () => void> = {
    'insert-row-above': () => module.insertRowAbove(),
    'insert-row-below': () => module.insertRowBelow(),
    'insert-column-left': () => module.insertColumnLeft(),
    'insert-column-right': () => module.insertColumnRight(),
    'delete-row': () => module.deleteRow(),
    'delete-column': () => module.deleteColumn(),
    'delete-table': () => module.deleteTable(),
  }
  actions[command]()
  syncActiveFormats(quill.value?.getSelection())
}

function openSourceMode() {
  sourceDraft.value = serializeRichText(quill.value?.root.innerHTML || content.value || '')
  sourceMode.value = true
}

function cancelSourceMode() {
  sourceDraft.value = ''
  sourceMode.value = false
}

async function applySource() {
  const editor = quill.value
  if (!editor)
    return
  const safeHtml = sanitizeRichText(sourceDraft.value)
  sourceMode.value = false
  await nextTick()
  editor.clipboard.dangerouslyPasteHTML(safeHtml, 'user')
  content.value = serializeRichText(editor.root.innerHTML)
  syncActiveFormats(editor.getSelection())
  ElMessage.success('源码已安全应用')
}

function toggleSourceMode() {
  if (sourceMode.value)
    cancelSourceMode()
  else
    openSourceMode()
}

function toggleFullScreen() {
  fullscreen.value = !fullscreen.value
}
</script>

<template>
  <section
    class="advanced-rich-text-editor"
    :class="{ 'is-fullscreen': fullscreen, 'is-source-mode': sourceMode }"
    :style="editorStyle"
  >
    <RichTextToolbar
      :active-formats="activeFormats"
      :can-undo="canUndo"
      :can-redo="canRedo"
      :in-table="inTable"
      :source-mode="sourceMode"
      :fullscreen="fullscreen"
      @format="applyFormat"
      @undo="undo"
      @redo="redo"
      @clear="clearFormatting"
      @link="editLink"
      @image="handleImageCommand"
      @insert-table="insertTable"
      @table-command="runTableCommand"
      @toggle-source="toggleSourceMode"
      @toggle-fullscreen="toggleFullScreen"
    />

    <div v-show="!sourceMode" class="advanced-rich-text-editor__visual">
      <QuillEditor
        v-model:content="content"
        content-type="html"
        :options="editorOptions"
        toolbar=""
        @ready="handleReady"
      />
    </div>

    <div v-show="sourceMode" class="advanced-rich-text-editor__source">
      <div class="source-editor__header">
        <div>
          <strong>HTML 源码</strong>
          <span>应用时会移除脚本、事件属性和危险链接。</span>
        </div>
        <div class="source-editor__actions">
          <el-button size="small" @click="cancelSourceMode">取消源码修改</el-button>
          <el-button type="primary" size="small" @click="applySource">应用源码</el-button>
        </div>
      </div>
      <el-input
        v-model="sourceDraft"
        class="source-editor__input"
        type="textarea"
        resize="none"
        spellcheck="false"
        aria-label="HTML 源码"
      />
    </div>

    <input ref="imageInput" class="advanced-rich-text-editor__file-input" type="file" accept="image/*" @change="handleImageFile">
  </section>
</template>

<style scoped>
.advanced-rich-text-editor {
  display: flex;
  min-height: var(--advanced-editor-min-height);
  overflow: hidden;
  flex-direction: column;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  background: #fff;
}

.advanced-rich-text-editor.is-fullscreen {
  position: fixed;
  z-index: 4000;
  inset: 12px;
  height: calc(100vh - 24px);
  min-height: 0;
  box-shadow: 0 18px 60px rgba(31, 35, 41, 0.24);
}

.advanced-rich-text-editor__visual,
.advanced-rich-text-editor__source {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
}

.advanced-rich-text-editor__visual :deep(.ql-toolbar) {
  display: none;
}

.advanced-rich-text-editor__visual :deep(.ql-container) {
  min-height: var(--advanced-editor-min-height);
  flex: 1;
  border: 0;
  font-size: 14px;
}

.is-fullscreen .advanced-rich-text-editor__visual :deep(.ql-container) {
  min-height: 0;
}

.advanced-rich-text-editor__visual :deep(.ql-editor) {
  min-height: var(--advanced-editor-min-height);
  padding: 16px 18px;
}

.is-fullscreen .advanced-rich-text-editor__visual :deep(.ql-editor) {
  min-height: calc(100vh - 90px);
}

.advanced-rich-text-editor__source {
  padding: 12px;
  gap: 10px;
  background: #f7f8fa;
}

.source-editor__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.source-editor__header > div:first-child {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px;
}

.source-editor__header strong {
  color: #1f2329;
  font-size: 14px;
}

.source-editor__header span {
  color: #86909c;
  font-size: 12px;
}

.source-editor__actions {
  display: flex;
  flex-shrink: 0;
  gap: 8px;
}

.source-editor__input {
  flex: 1;
}

.source-editor__input :deep(.el-textarea),
.source-editor__input :deep(.el-textarea__inner) {
  height: 100%;
  min-height: var(--advanced-editor-min-height) !important;
}

.source-editor__input :deep(.el-textarea__inner) {
  padding: 14px 16px;
  border-radius: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 13px;
  line-height: 1.6;
  tab-size: 2;
}

.advanced-rich-text-editor__file-input {
  display: none;
}

@media (max-width: 768px) {
  .source-editor__header {
    align-items: stretch;
    flex-direction: column;
  }

  .source-editor__actions {
    justify-content: flex-end;
  }
}
</style>
