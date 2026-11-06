<script setup lang="ts">
import {
  ArrowLeftBold,
  ArrowRightBold,
  Bottom,
  Brush,
  Delete,
  Document,
  FullScreen,
  Grid,
  Link,
  Picture,
  RefreshLeft,
  RefreshRight,
  Top,
} from '@element-plus/icons-vue'
import { computed, ref } from 'vue'
import {
  RICH_TEXT_FONT_OPTIONS,
  RICH_TEXT_HEADER_OPTIONS,
  RICH_TEXT_LANGUAGES,
  RICH_TEXT_SIZE_OPTIONS,
} from './richTextConfig'

type FormatValue = string | number | boolean
type TableCommand = 'insert-row-above' | 'insert-row-below' | 'insert-column-left' | 'insert-column-right' | 'delete-row' | 'delete-column' | 'delete-table'

const props = defineProps<{
  activeFormats: Record<string, unknown>
  canUndo: boolean
  canRedo: boolean
  inTable: boolean
  sourceMode: boolean
  fullscreen: boolean
}>()

const emit = defineEmits<{
  format: [name: string, value: FormatValue]
  undo: []
  redo: []
  clear: []
  link: []
  image: [type: 'file' | 'url']
  insertTable: [rows: number, columns: number]
  tableCommand: [command: TableCommand]
  toggleSource: []
  toggleFullscreen: []
}>()

const tablePopoverVisible = ref(false)
const tableRows = ref(2)
const tableColumns = ref(2)

const disabled = computed(() => props.sourceMode)
const headerValue = computed(() => String(props.activeFormats.header || ''))
const fontValue = computed(() => String(props.activeFormats.font || ''))
const sizeValue = computed(() => String(props.activeFormats.size || ''))
const alignValue = computed(() => String(props.activeFormats.align || ''))
const codeLanguage = computed(() => {
  const value = props.activeFormats['code-block']
  return typeof value === 'string' && value ? value : 'plain'
})
const textColor = computed(() => typeof props.activeFormats.color === 'string' ? props.activeFormats.color : '#1f2329')
const backgroundColor = computed(() => typeof props.activeFormats.background === 'string' ? props.activeFormats.background : '#ffffff')

function isActive(name: string, expected: FormatValue = true) {
  const value = props.activeFormats[name]
  return expected === true ? Boolean(value) : String(value || '') === String(expected)
}

function toggleFormat(name: string, value: FormatValue = true) {
  emit('format', name, isActive(name, value) ? false : value)
}

function setFormat(name: string, value: string | number | boolean | undefined) {
  emit('format', name, value || false)
}

function setSelectFormat(name: string, value: unknown) {
  setFormat(name, typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean' ? value : '')
}

function setColorFormat(name: 'color' | 'background', value: string | null) {
  setFormat(name, value || '')
}

function handleHeaderChange(value: unknown) { setSelectFormat('header', value) }
function handleFontChange(value: unknown) { setSelectFormat('font', value) }
function handleSizeChange(value: unknown) { setSelectFormat('size', value) }
function handleAlignChange(value: unknown) { setSelectFormat('align', value) }
function handleCodeLanguageChange(value: unknown) { setSelectFormat('code-block', value) }
function handleTextColorChange(value: string | null) { setColorFormat('color', value) }
function handleBackgroundColorChange(value: string | null) { setColorFormat('background', value) }

function handleImageDropdown(command: unknown) {
  if (command === 'file' || command === 'url')
    emit('image', command)
}

function handleTableDropdown(command: unknown) {
  const validCommands: TableCommand[] = ['insert-row-above', 'insert-row-below', 'insert-column-left', 'insert-column-right', 'delete-row', 'delete-column', 'delete-table']
  if (typeof command === 'string' && validCommands.includes(command as TableCommand))
    emit('tableCommand', command as TableCommand)
}

function insertTable() {
  emit('insertTable', tableRows.value, tableColumns.value)
  tablePopoverVisible.value = false
}
</script>

<template>
  <div class="rich-text-toolbar" role="toolbar" aria-label="富文本编辑工具栏">
    <div class="rich-text-toolbar__row rich-text-toolbar__row--primary">
      <div class="rich-text-toolbar__group rich-text-toolbar__selects">
      <el-select :model-value="headerValue" class="toolbar-select toolbar-select--heading" size="small" placeholder="正文" aria-label="段落格式" :disabled="disabled" @change="handleHeaderChange">
        <el-option v-for="item in RICH_TEXT_HEADER_OPTIONS" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-select :model-value="fontValue" class="toolbar-select toolbar-select--font" size="small" placeholder="默认字体" aria-label="字体" :disabled="disabled" @change="handleFontChange">
        <el-option v-for="item in RICH_TEXT_FONT_OPTIONS" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      <el-select :model-value="sizeValue" class="toolbar-select toolbar-select--size" size="small" placeholder="默认字号" aria-label="字号" :disabled="disabled" @change="handleSizeChange">
        <el-option v-for="item in RICH_TEXT_SIZE_OPTIONS" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      </div>

      <div class="rich-text-toolbar__group rich-text-toolbar__group--inline">
      <el-tooltip content="粗体" placement="top">
        <el-button class="toolbar-button toolbar-button--text" :class="{ 'is-active': isActive('bold') }" :aria-pressed="isActive('bold')" aria-label="粗体" :disabled="disabled" @mousedown.prevent @click="toggleFormat('bold')"><strong>B</strong></el-button>
      </el-tooltip>
      <el-tooltip content="斜体" placement="top">
        <el-button class="toolbar-button toolbar-button--text" :class="{ 'is-active': isActive('italic') }" :aria-pressed="isActive('italic')" aria-label="斜体" :disabled="disabled" @mousedown.prevent @click="toggleFormat('italic')"><em>I</em></el-button>
      </el-tooltip>
      <el-tooltip content="下划线" placement="top">
        <el-button class="toolbar-button toolbar-button--text" :class="{ 'is-active': isActive('underline') }" :aria-pressed="isActive('underline')" aria-label="下划线" :disabled="disabled" @mousedown.prevent @click="toggleFormat('underline')"><u>U</u></el-button>
      </el-tooltip>
      <el-tooltip content="删除线" placement="top">
        <el-button class="toolbar-button toolbar-button--text" :class="{ 'is-active': isActive('strike') }" :aria-pressed="isActive('strike')" aria-label="删除线" :disabled="disabled" @mousedown.prevent @click="toggleFormat('strike')"><s>S</s></el-button>
      </el-tooltip>
      <el-tooltip content="文字颜色" placement="top">
        <span class="toolbar-trigger toolbar-color-control"><strong>A</strong><el-color-picker :model-value="textColor" size="small" aria-label="文字颜色" :disabled="disabled" @change="handleTextColorChange" /></span>
      </el-tooltip>
      <el-tooltip content="背景颜色" placement="top">
        <span class="toolbar-trigger toolbar-color-control"><strong>底</strong><el-color-picker :model-value="backgroundColor" size="small" aria-label="背景颜色" :disabled="disabled" @change="handleBackgroundColorChange" /></span>
      </el-tooltip>
      </div>
    </div>

    <div class="rich-text-toolbar__row rich-text-toolbar__row--secondary">
      <div class="rich-text-toolbar__group rich-text-toolbar__group--paragraph">
      <el-tooltip content="有序列表" placement="top">
        <el-button class="toolbar-button toolbar-button--text" :class="{ 'is-active': isActive('list', 'ordered') }" :aria-pressed="isActive('list', 'ordered')" aria-label="有序列表" :disabled="disabled" @mousedown.prevent @click="toggleFormat('list', 'ordered')">1.</el-button>
      </el-tooltip>
      <el-tooltip content="无序列表" placement="top">
        <el-button class="toolbar-button toolbar-button--text" :class="{ 'is-active': isActive('list', 'bullet') }" :aria-pressed="isActive('list', 'bullet')" aria-label="无序列表" :disabled="disabled" @mousedown.prevent @click="toggleFormat('list', 'bullet')">•</el-button>
      </el-tooltip>
      <el-tooltip content="减少缩进" placement="top">
        <el-button class="toolbar-button" aria-label="减少缩进" :disabled="disabled" @mousedown.prevent @click="emit('format', 'indent', '-1')"><el-icon><ArrowLeftBold /></el-icon></el-button>
      </el-tooltip>
      <el-tooltip content="增加缩进" placement="top">
        <el-button class="toolbar-button" aria-label="增加缩进" :disabled="disabled" @mousedown.prevent @click="emit('format', 'indent', '+1')"><el-icon><ArrowRightBold /></el-icon></el-button>
      </el-tooltip>
      <el-select :model-value="alignValue" class="toolbar-select toolbar-select--align" size="small" placeholder="左对齐" aria-label="文本对齐" :disabled="disabled" @change="handleAlignChange">
        <el-option label="左对齐" value="" />
        <el-option label="居中" value="center" />
        <el-option label="右对齐" value="right" />
        <el-option label="两端对齐" value="justify" />
      </el-select>
      <el-tooltip content="引用" placement="top">
        <el-button class="toolbar-button toolbar-button--text" :class="{ 'is-active': isActive('blockquote') }" :aria-pressed="isActive('blockquote')" aria-label="引用" :disabled="disabled" @mousedown.prevent @click="toggleFormat('blockquote')">“”</el-button>
      </el-tooltip>
      <el-tooltip content="代码块" placement="top">
        <el-button class="toolbar-button toolbar-button--code" :class="{ 'is-active': isActive('code-block') }" :aria-pressed="isActive('code-block')" aria-label="代码块" :disabled="disabled" @mousedown.prevent @click="toggleFormat('code-block', codeLanguage)">&lt;/&gt;</el-button>
      </el-tooltip>
      <el-select v-if="isActive('code-block')" :model-value="codeLanguage" class="toolbar-select toolbar-select--language" size="small" aria-label="代码语言" :disabled="disabled" @change="handleCodeLanguageChange">
        <el-option v-for="item in RICH_TEXT_LANGUAGES" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
      </div>

      <div class="rich-text-toolbar__group rich-text-toolbar__group--insert">
      <el-tooltip content="插入链接" placement="top">
        <el-button class="toolbar-button" aria-label="插入链接" :disabled="disabled" @mousedown.prevent @click="emit('link')"><el-icon><Link /></el-icon></el-button>
      </el-tooltip>
      <el-dropdown trigger="click" :disabled="disabled" @command="handleImageDropdown">
        <el-button class="toolbar-button" aria-label="插入图片" title="插入图片" :disabled="disabled"><el-icon><Picture /></el-icon></el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="file">选择本地图片</el-dropdown-item>
            <el-dropdown-item command="url">插入图片地址</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <el-popover v-model:visible="tablePopoverVisible" placement="bottom" :width="250" trigger="click">
        <template #reference>
          <el-button class="toolbar-button" aria-label="插入表格" title="插入表格" :disabled="disabled"><el-icon><Grid /></el-icon></el-button>
        </template>
        <div class="table-insert-panel">
          <label>行数 <el-input-number v-model="tableRows" :min="1" :max="10" size="small" /></label>
          <label>列数 <el-input-number v-model="tableColumns" :min="1" :max="10" size="small" /></label>
          <el-button type="primary" size="small" @click="insertTable">插入表格</el-button>
        </div>
      </el-popover>

      <el-dropdown trigger="click" :disabled="disabled || !inTable" @command="handleTableDropdown">
        <el-button class="toolbar-button" aria-label="表格行列操作" title="表格行列操作" :disabled="disabled || !inTable"><el-icon><Grid /></el-icon></el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="insert-row-above"><el-icon><Top /></el-icon>上方插入行</el-dropdown-item>
            <el-dropdown-item command="insert-row-below"><el-icon><Bottom /></el-icon>下方插入行</el-dropdown-item>
            <el-dropdown-item command="insert-column-left"><el-icon><ArrowLeftBold /></el-icon>左侧插入列</el-dropdown-item>
            <el-dropdown-item command="insert-column-right"><el-icon><ArrowRightBold /></el-icon>右侧插入列</el-dropdown-item>
            <el-dropdown-item command="delete-row" divided><el-icon><Delete /></el-icon>删除当前行</el-dropdown-item>
            <el-dropdown-item command="delete-column"><el-icon><Delete /></el-icon>删除当前列</el-dropdown-item>
            <el-dropdown-item command="delete-table"><el-icon><Delete /></el-icon>删除整张表格</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      </div>

      <div class="rich-text-toolbar__group rich-text-toolbar__group--end">
      <el-tooltip content="撤销" placement="top">
        <el-button class="toolbar-button" aria-label="撤销" :disabled="disabled || !canUndo" @mousedown.prevent @click="emit('undo')"><el-icon><RefreshLeft /></el-icon></el-button>
      </el-tooltip>
      <el-tooltip content="重做" placement="top">
        <el-button class="toolbar-button" aria-label="重做" :disabled="disabled || !canRedo" @mousedown.prevent @click="emit('redo')"><el-icon><RefreshRight /></el-icon></el-button>
      </el-tooltip>
      <el-tooltip content="清除所选内容格式" placement="top">
        <el-button class="toolbar-button toolbar-button--labeled" aria-label="清除格式" :disabled="disabled" @mousedown.prevent @click="emit('clear')"><el-icon><Brush /></el-icon><span>清除格式</span></el-button>
      </el-tooltip>
      <el-tooltip :content="sourceMode ? '返回可视化编辑' : '编辑 HTML 源码'" placement="top">
        <el-button class="toolbar-button" :class="{ 'is-active': sourceMode }" :aria-pressed="sourceMode" aria-label="源码" @click="emit('toggleSource')"><el-icon><Document /></el-icon></el-button>
      </el-tooltip>
      <el-tooltip :content="fullscreen ? '退出全屏' : '全屏编辑'" placement="top">
        <el-button class="toolbar-button" :class="{ 'is-active': fullscreen }" :aria-pressed="fullscreen" aria-label="全屏" @click="emit('toggleFullscreen')"><el-icon><FullScreen /></el-icon></el-button>
      </el-tooltip>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rich-text-toolbar {
  display: flex;
  align-items: stretch;
  flex-direction: column;
  gap: 5px;
  min-height: 70px;
  padding: 7px 9px;
  border-bottom: 1px solid #dcdfe6;
  background: #fff;
}

.rich-text-toolbar__row {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 6px;
}

.rich-text-toolbar__row--primary {
  min-height: 30px;
}

.rich-text-toolbar__row--secondary {
  min-height: 30px;
}

.rich-text-toolbar__group {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  min-width: 0;
  padding-right: 6px;
  border-right: 1px solid #e5e6eb;
}

.rich-text-toolbar__row > .rich-text-toolbar__group:last-child {
  border-right: 0;
}

.rich-text-toolbar__group--end {
  margin-left: auto;
  padding-right: 0;
  border-right: 0;
}

.toolbar-button {
  width: 30px;
  height: 30px;
  min-height: 30px;
  margin: 0;
  padding: 0;
  border-color: transparent;
  color: #4e5969;
}

.toolbar-button:hover,
.toolbar-button:focus-visible,
.toolbar-button.is-active {
  border-color: #91caff;
  background: #eaf3ff;
  color: #1677ff;
}

.toolbar-button.is-active {
  box-shadow: 0 0 0 1px #1677ff inset;
}

.toolbar-button--text,
.toolbar-button--code {
  font-size: 14px;
  font-weight: 600;
}

.toolbar-button--code {
  width: 38px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.toolbar-button--labeled {
  width: auto;
  padding: 0 8px;
  gap: 4px;
  font-size: 12px;
}

.toolbar-trigger {
  display: inline-flex;
  align-items: center;
}

.toolbar-color-control {
  gap: 2px;
  color: #4e5969;
  font-size: 11px;
}

.toolbar-color-control strong {
  min-width: 12px;
  text-align: center;
}

.toolbar-color-control :deep(.el-color-picker__trigger) {
  width: 24px;
  height: 24px;
  padding: 3px;
}

.toolbar-select {
  flex: 0 0 auto;
}

.toolbar-select--heading { width: 96px; }
.toolbar-select--font { width: 118px; }
.toolbar-select--size { width: 92px; }
.toolbar-select--align { width: 92px; }
.toolbar-select--language { width: 116px; }

.table-insert-panel {
  display: grid;
  gap: 10px;
}

.table-insert-panel label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: #4e5969;
  font-size: 13px;
}

@media (max-width: 900px) {
  .rich-text-toolbar__row {
    flex-wrap: wrap;
  }

  .rich-text-toolbar__group--end {
    margin-left: auto;
  }
}
</style>
