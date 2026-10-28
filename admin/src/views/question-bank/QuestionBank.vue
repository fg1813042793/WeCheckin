<template>
  <div class="admin-page question-bank-page">
    <section class="question-bank-mobile-notice" aria-label="移动端编辑提示">
      <strong>建议在桌面端维护题库</strong>
      <span>题库工作台需要同时显示题库、编辑区和属性面板，请使用宽度不小于 768px 的设备。</span>
    </section>

    <section class="question-bank-workspace">
      <header class="question-bank-workspace__toolbar">
        <div class="question-bank-workspace__identity">
          <strong>题库管理</strong>
          <div class="question-bank-workspace__scope" role="tablist" aria-label="题库类型">
            <button type="button" :class="{ active: activeScope === 'survey' }" @click="handleScopeSelect('survey')">问卷题库</button>
            <button type="button" :class="{ active: activeScope === 'exam' }" @click="handleScopeSelect('exam')">考试题库</button>
          </div>
        </div>

        <div class="question-bank-workspace__actions">
          <input ref="importFileInput" class="bank-import-input" type="file" accept=".json,application/json" @change="handleImportFile" />
          <el-button :icon="Upload" :loading="importing" :disabled="!canAddQuestion" @click="triggerImport">导入</el-button>
          <el-button :icon="Download" :loading="exporting" @click="exportQuestions">导出</el-button>
          <el-button :icon="Refresh" title="刷新题目" @click="load" />
          <el-button :icon="Menu" title="显示或隐藏题库导航" @click="toggleLeftPanel" />
          <el-button :icon="Setting" title="显示或隐藏题目属性" @click="toggleRightPanel" />
          <span v-if="editDialog.visible" class="question-bank-workspace__save-state" :class="`is-${saveState}`">{{ saveStatusLabel }}</span>
          <el-button
            v-if="editDialog.visible"
            type="primary"
            :icon="Check"
            :loading="saving"
            :disabled="!canModifyCurrentQuestion"
            @click="save(false)"
          >立即保存</el-button>
        </div>
      </header>

      <div
        class="question-bank-workspace__body"
        :class="{ 'is-left-collapsed': leftCollapsed, 'is-right-collapsed': rightCollapsed }"
      >
        <QuestionBankNavigator
          v-show="!leftCollapsed"
          :categories="categories"
          :active-category="category"
          :keyword="keyword"
          :question-type="questionType"
          :type-options="typeOptions"
          :list="list"
          :selected-id="form.id"
          :loading="loading"
          :error="listError"
          :total="total"
          :page="page"
          :page-size="pageSize"
          :can-add="canAddQuestion"
          :question-title="questionTitle"
          :type-name="typeName"
          @select-category="selectCategory"
          @select-question="selectQuestion"
          @add-question="openCreate"
          @update:keyword="keyword = $event"
          @update:question-type="questionType = $event"
          @update:page="page = $event; load()"
          @search="search"
          @retry="load"
        />

        <QuestionEditorCanvas
          v-model:editor-mode="editorMode"
          :active="editDialog.visible"
          :can-add="canAddQuestion"
          :readonly="!canModifyCurrentQuestion"
          :form="form"
          :visual-question="visualQuestion"
          :rich-title-editor-options="richTitleEditorOptions"
          :rich-option-editor-options="richOptionEditorOptions"
          :rich-full-editor-options="richFullEditorOptions"
          :rich-edit-dialog="richEditDialog"
          :rich-edit-title="richEditTitle"
          :bind-compact-rich-editor="bindCompactRichEditor"
          :sync-plain-title-from-visual="syncPlainTitleFromVisual"
          :uses-options="usesOptions"
          :uses-fields="usesFields"
          :uses-matrix="usesMatrix"
          :add-visual-option="addVisualOption"
          :remove-visual-option="removeVisualOption"
          :add-visual-field="addVisualField"
          :remove-visual-field="removeVisualField"
          :add-matrix-row="addMatrixRow"
          :remove-matrix-row="removeMatrixRow"
          :add-matrix-column="addMatrixColumn"
          :remove-matrix-column="removeMatrixColumn"
          :format-schema="formatSchema"
          :apply-json-to-visual="applyJsonToVisual"
          :confirm-rich-full-edit="confirmRichFullEdit"
          @add-question="openCreate"
        />

        <QuestionPropertiesPanel
          v-show="!rightCollapsed"
          :active="editDialog.visible"
          :active-scope="activeScope"
          :form="form"
          :visual-question="visualQuestion"
          :categories="categories"
          :type-options="typeOptions"
          :readonly="!canModifyCurrentQuestion"
          :can-delete="canDeleteQuestion && editDialog.isEdit"
          :type-name="typeName"
          @type-change="handleTypeChange"
          @delete-question="deleteCurrentQuestion"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Download, Menu, Refresh, Setting, Upload } from '@element-plus/icons-vue'
import { adminApi } from '../../api'
import { hasPerm } from '../../utils/permission'
import QuestionBankNavigator from './components/QuestionBankNavigator.vue'
import QuestionEditorCanvas from './components/QuestionEditorCanvas.vue'
import QuestionPropertiesPanel from './components/QuestionPropertiesPanel.vue'
import {
  ensureQuestionBankQuestionShape,
  nextOptionValue,
  parseQuestionBankSchema,
  questionBankTypeOptions,
  richFullEditorOptions,
  richOptionEditorOptions,
  richTitleEditorOptions,
  stripQuestionHtml as stripHtml,
  typeName,
  usesFields,
  usesMatrix,
  usesOptions,
} from './questionBankEditor'
import {
  buildQuestionBankExportPayload,
  downloadJson,
  exportFilename,
  normalizeExportQuestion,
  normalizeSchemaValue,
  parseImportQuestions,
  type BankScope,
  type QuestionPayload
} from './utils/importExport'
type RichEditTarget = 'title' | 'option'

const activeScope = ref<BankScope>('survey')
const keyword = ref('')
const category = ref('')
const questionType = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const listError = ref('')
const saving = ref(false)
const importing = ref(false)
const exporting = ref(false)
const list = ref<any[]>([])
const categories = ref<string[]>([])
const importFileInput = ref<HTMLInputElement | null>(null)
const editorMode = ref<'visual' | 'preview' | 'json'>('visual')
const visualQuestion = ref<any>({ props: {} })
const saveState = ref<'idle' | 'dirty' | 'saving' | 'saved' | 'error'>('idle')
const leftCollapsed = ref(false)
const rightCollapsed = ref(false)
const canAddQuestion = computed(() => hasPerm('admin:menu:question-bank:add'))
const canEditQuestion = computed(() => hasPerm('admin:menu:question-bank:edit'))
const canDeleteQuestion = computed(() => hasPerm('admin:menu:question-bank:del'))
const canModifyCurrentQuestion = computed(() => editDialog.isEdit ? canEditQuestion.value : canAddQuestion.value)

const editDialog = reactive({ visible: false, isEdit: false })
const form = reactive({
  id: 0,
  title: '',
  type: '',
  category: '',
  tags: '',
  schema: ''
})
const richEditDialog = reactive({
  visible: false,
  target: 'title' as RichEditTarget,
  optionIndex: -1,
  content: ''
})
let autoSaveTimer: ReturnType<typeof setTimeout> | null = null
let activeSavePromise: Promise<boolean> | null = null
let suppressDraftWatch = false
let lastSavedSnapshot = ''

const saveStatusLabel = computed(() => ({
  idle: '选择题目后开始编辑',
  dirty: editDialog.isEdit ? '有未保存修改' : '新题目需首次保存',
  saving: '保存中...',
  saved: '已保存',
  error: '保存失败，修改已保留',
})[saveState.value])

const typeOptions = computed(() => questionBankTypeOptions)
const richEditTitle = computed(() => richEditDialog.target === 'title' ? '完整编辑题目标题' : `完整编辑选项 ${richEditDialog.optionIndex + 1}`)

function scopeName(scope: BankScope) {
  return scope === 'survey' ? '问卷题库' : '考试题库'
}

function bindCompactRichEditor(quill: any, target: RichEditTarget, optionIndex = -1) {
  const toolbar = quill?.getModule?.('toolbar')
  const container = toolbar?.container as HTMLElement | undefined
  if (!container) return
  if (container.querySelector('.ql-bank-full')) return
  const button = document.createElement('button')
  button.type = 'button'
  button.className = 'ql-bank-full'
  button.title = '高级编辑'
  button.textContent = '高级编辑'
  button.addEventListener('click', (event) => {
    event.preventDefault()
    event.stopPropagation()
    openRichFullEdit(target, optionIndex)
  })
  container.appendChild(button)
}

function openRichFullEdit(target: RichEditTarget, optionIndex = -1) {
  richEditDialog.target = target
  richEditDialog.optionIndex = optionIndex
  richEditDialog.content = target === 'title'
    ? String(visualQuestion.value.title || '')
    : String(visualQuestion.value.props?.options?.[optionIndex]?.label || '')
  richEditDialog.visible = true
}

function confirmRichFullEdit() {
  if (richEditDialog.target === 'title') {
    visualQuestion.value.title = richEditDialog.content
    syncPlainTitleFromVisual()
  } else {
    const option = visualQuestion.value.props?.options?.[richEditDialog.optionIndex]
    if (option) option.label = richEditDialog.content
  }
  richEditDialog.visible = false
}

function questionTitle(row: any) {
  return stripHtml(row?.title || '') || '未命名题目'
}

function formatTime(value: number) {
  if (!value) return '-'
  const ms = value < 1000000000000 ? value * 1000 : value
  const date = new Date(ms)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN', { hour12: false })
}

function prettySchema(schema: string) {
  if (!schema) return '-'
  try {
    return JSON.stringify(JSON.parse(schema), null, 2)
  } catch {
    return schema
  }
}

function ensureVisualQuestionShape(question: any) {
  return ensureQuestionBankQuestionShape(question, {
    type: form.type || 'radio', title: form.title || typeName(form.type || 'radio'), scope: activeScope.value,
  })
}

function parseSchemaQuestion(schema: any, fallback: any = {}) {
  return parseQuestionBankSchema(schema, fallback)
}

function syncPlainTitleFromVisual() {
  form.title = stripHtml(visualQuestion.value.title || '') || typeName(visualQuestion.value.type)
}

function syncSchemaFromVisual() {
  syncPlainTitleFromVisual()
  const schema = ensureVisualQuestionShape({
    ...visualQuestion.value,
    title: visualQuestion.value.title || form.title.trim(),
    type: form.type
  })
  delete schema.id
  visualQuestion.value = schema
  form.schema = JSON.stringify(schema, null, 2)
}

function draftSnapshot() {
  return JSON.stringify({
    title: form.title,
    type: form.type,
    category: form.category,
    tags: form.tags,
    visualQuestion: visualQuestion.value,
  })
}

function clearAutoSaveTimer() {
  if (!autoSaveTimer) return
  clearTimeout(autoSaveTimer)
  autoSaveTimer = null
}

function scheduleAutoSave() {
  if (suppressDraftWatch || !editDialog.visible || !canModifyCurrentQuestion.value) return
  const snapshot = draftSnapshot()
  if (snapshot === lastSavedSnapshot) return
  saveState.value = 'dirty'
  clearAutoSaveTimer()
  if (!editDialog.isEdit) return
  autoSaveTimer = setTimeout(() => {
    autoSaveTimer = null
    void save(true)
  }, 800)
}

async function flushPendingSave() {
  clearAutoSaveTimer()
  if (!editDialog.visible || saveState.value === 'idle' || saveState.value === 'saved') return true
  if (!editDialog.isEdit) return false
  return await save(true)
}

async function confirmWorkspaceChange() {
  if (await flushPendingSave()) return true
  try {
    await ElMessageBox.confirm(
      editDialog.isEdit ? '当前题目保存失败，继续切换将放弃未保存修改。' : '新增题目尚未保存，继续切换将放弃当前内容。',
      '存在未保存修改',
      { type: 'warning', confirmButtonText: '放弃修改', cancelButtonText: '继续编辑' },
    )
    return true
  } catch {
    return false
  }
}

function applyVisualFromSchema(showMessage = false) {
  try {
    const parsed = parseSchemaQuestion(form.schema, {})
    visualQuestion.value = ensureVisualQuestionShape({
      ...parsed,
      type: parsed.type || form.type || 'radio',
      title: parsed.title || form.title || typeName(parsed.type || form.type || 'radio')
    })
    form.type = visualQuestion.value.type
    form.title = stripHtml(visualQuestion.value.title)
    if (showMessage) ElMessage.success('已同步到可视化')
    return true
  } catch {
    if (showMessage) ElMessage.warning('当前 JSON 不是有效题目配置')
    return false
  }
}

function applyJsonToVisual() {
  applyVisualFromSchema(true)
}

function handleEditorModeChange(name: string | number) {
  if (name === 'json') syncSchemaFromVisual()
  if (name === 'visual') applyVisualFromSchema()
}

function handleTypeChange(type: string) {
  form.type = type
  visualQuestion.value = ensureVisualQuestionShape({
    ...visualQuestion.value,
    type,
    title: visualQuestion.value.title || form.title || typeName(type)
  })
  syncSchemaFromVisual()
}

function addVisualOption() {
  const options = visualQuestion.value.props.options
  options.push({ label: `选项${nextOptionValue(options.length)}`, value: nextOptionValue(options.length) })
}

function removeVisualOption(index: number) {
  visualQuestion.value.props.options.splice(index, 1)
}

function addVisualField() {
  const fields = visualQuestion.value.props.fields
  fields.push({ label: `填空${fields.length + 1}`, placeholder: '' })
}

function removeVisualField(index: number) {
  visualQuestion.value.props.fields.splice(index, 1)
}

function addMatrixRow() {
  const rows = visualQuestion.value.props.rows
  rows.push({ id: `r${rows.length + 1}`, title: `行${rows.length + 1}` })
}

function removeMatrixRow(index: number) {
  visualQuestion.value.props.rows.splice(index, 1)
}

function addMatrixColumn() {
  const columns = visualQuestion.value.props.columns
  columns.push({ id: `c${columns.length + 1}`, title: `列${columns.length + 1}` })
}

function removeMatrixColumn(index: number) {
  visualQuestion.value.props.columns.splice(index, 1)
}

function buildListParams(extra: Record<string, any> = {}) {
  return {
    page: page.value,
    pageSize: pageSize.value,
    keyword: keyword.value,
    category: category.value === '__uncategorized__' ? '' : category.value,
    type: questionType.value,
    ...extra
  }
}

function listFromResponse(res: any) {
  return Array.isArray(res.data?.list) ? res.data.list : []
}

async function requestQuestionBankList(scope: BankScope, params: Record<string, any>) {
  return scope === 'survey'
    ? await adminApi.surveyQuestionBankList(params)
    : await adminApi.examQuestionBankList(params)
}

async function insertQuestion(scope: BankScope, payload: QuestionPayload) {
  if (scope === 'survey') await adminApi.surveyQuestionBankInsert(payload)
  else await adminApi.examQuestionBankInsert(payload)
}

async function loadCategories() {
  const api = activeScope.value === 'survey' ? adminApi.surveyQuestionBankCategories : adminApi.examQuestionBankCategories
  try {
    const res: any = await api()
    categories.value = Array.isArray(res.data) ? res.data : []
  } catch {
    categories.value = []
  }
}

async function load() {
  loading.value = true
  listError.value = ''
  try {
    if (category.value === '__uncategorized__') {
      const rows = await fetchAllQuestionRows(activeScope.value, { keyword: keyword.value, type: questionType.value })
      const uncategorized = rows.filter(row => !String(row?.category || '').trim())
      total.value = uncategorized.length
      const offset = (page.value - 1) * pageSize.value
      list.value = uncategorized.slice(offset, offset + pageSize.value)
    } else {
      const res: any = await requestQuestionBankList(activeScope.value, buildListParams())
      list.value = listFromResponse(res)
      total.value = Number(res.data?.total || 0)
    }
  } catch (error: any) {
    list.value = []
    total.value = 0
    listError.value = error?.message || error?.msg || '题目加载失败'
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

async function handleScopeSelect(scope: BankScope) {
  if (scope === activeScope.value) return
  if (!await confirmWorkspaceChange()) return
  activeScope.value = scope
  page.value = 1
  keyword.value = ''
  category.value = ''
  questionType.value = ''
  list.value = []
  total.value = 0
  clearCurrentQuestion()
  await loadCategories()
  await load()
}

function resetForm() {
  suppressDraftWatch = true
  form.id = 0
  form.title = '未命名题目'
  form.type = 'radio'
  form.category = category.value === '__uncategorized__' ? '' : category.value
  form.tags = ''
  visualQuestion.value = ensureVisualQuestionShape({ title: form.title, type: form.type })
  syncSchemaFromVisual()
  editorMode.value = 'visual'
  nextTick(() => {
    suppressDraftWatch = false
    lastSavedSnapshot = ''
    saveState.value = 'dirty'
  })
}

async function openCreate() {
  if (!canAddQuestion.value || !await confirmWorkspaceChange()) return
  resetForm()
  editDialog.isEdit = false
  editDialog.visible = true
}

async function selectQuestion(row: any) {
  if (Number(row?.id || 0) === form.id && editDialog.visible) return
  if (!await confirmWorkspaceChange()) return
  openEdit(row)
}

function openEdit(row: any) {
  suppressDraftWatch = true
  form.id = Number(row.id || 0)
  form.title = stripHtml(row.title || '')
  form.type = row.type || ''
  form.category = row.category || ''
  form.tags = row.tags || ''
  form.schema = row.schema || ''
  try {
    const parsed = parseSchemaQuestion(form.schema, {})
    visualQuestion.value = ensureVisualQuestionShape({
      ...parsed,
      type: parsed.type || form.type || 'radio',
      title: parsed.title || form.title || typeName(parsed.type || form.type || 'radio')
    })
    form.type = visualQuestion.value.type
    form.title = stripHtml(visualQuestion.value.title)
    syncSchemaFromVisual()
  } catch {
    visualQuestion.value = ensureVisualQuestionShape({ title: form.title || '未命名题目', type: form.type || 'radio' })
  }
  editorMode.value = 'visual'
  editDialog.isEdit = true
  editDialog.visible = true
  nextTick(() => {
    suppressDraftWatch = false
    lastSavedSnapshot = draftSnapshot()
    saveState.value = 'saved'
  })
}

function clearCurrentQuestion() {
  clearAutoSaveTimer()
  editDialog.visible = false
  editDialog.isEdit = false
  form.id = 0
  lastSavedSnapshot = ''
  saveState.value = 'idle'
}

async function selectCategory(value: string) {
  if (value === category.value) return
  if (!await confirmWorkspaceChange()) return
  category.value = value
  page.value = 1
  clearCurrentQuestion()
  await load()
}

function formatSchema() {
  if (!form.schema.trim()) return
  try {
    form.schema = JSON.stringify(JSON.parse(form.schema), null, 2)
  } catch {
    ElMessage.warning('当前内容不是有效 JSON')
  }
}

function save(automatic = false): Promise<boolean> {
  if (activeSavePromise) return activeSavePromise
  activeSavePromise = persistQuestion(automatic).finally(() => {
    activeSavePromise = null
  })
  return activeSavePromise
}

async function persistQuestion(automatic = false): Promise<boolean> {
  if (!canModifyCurrentQuestion.value) return false
  clearAutoSaveTimer()
  suppressDraftWatch = true
  if (editorMode.value !== 'json') syncSchemaFromVisual()
  const plainTitle = stripHtml(form.title || visualQuestion.value.title || '').trim()
  if (!plainTitle) {
    if (!automatic) ElMessage.warning('请输入题目标题')
    suppressDraftWatch = false
    saveState.value = 'error'
    return false
  }
  if (!form.type) {
    if (!automatic) ElMessage.warning('请选择题型')
    suppressDraftWatch = false
    saveState.value = 'error'
    return false
  }
  saving.value = true
  saveState.value = 'saving'
  try {
    const payload = {
      id: form.id,
      title: plainTitle,
      type: form.type,
      category: form.category || '',
      tags: form.tags || '',
      schema: form.schema || ''
    }
    let response: any
    if (activeScope.value === 'survey') {
      response = editDialog.isEdit
        ? await adminApi.surveyQuestionBankEdit(payload)
        : await adminApi.surveyQuestionBankInsert(payload)
    } else {
      response = editDialog.isEdit
        ? await adminApi.examQuestionBankEdit(payload)
        : await adminApi.examQuestionBankInsert(payload)
    }
    if (!editDialog.isEdit) {
      form.id = Number(response?.data?.id || form.id || 0)
      editDialog.isEdit = form.id > 0
    }
    lastSavedSnapshot = draftSnapshot()
    saveState.value = 'saved'
    if (!automatic) ElMessage.success('已保存')
    await loadCategories()
    await load()
    return true
  } catch {
    saveState.value = 'error'
    return false
  } finally {
    saving.value = false
    suppressDraftWatch = false
  }
}

async function fetchAllQuestionRows(scope: BankScope, filters: Record<string, any> = {}) {
  const pageSizeForExport = 500
  const first: any = await requestQuestionBankList(scope, {
    page: 1,
    pageSize: pageSizeForExport,
    keyword: filters.keyword ?? keyword.value,
    category: filters.category ?? (category.value === '__uncategorized__' ? '' : category.value),
    type: filters.type ?? questionType.value,
  })
  const items = [...listFromResponse(first)]
  const allTotal = Number(first.data?.total || items.length)
  const pageCount = Math.ceil(allTotal / pageSizeForExport)
  for (let currentPage = 2; currentPage <= pageCount; currentPage++) {
    const res: any = await requestQuestionBankList(scope, {
      page: currentPage,
      pageSize: pageSizeForExport,
      keyword: filters.keyword ?? keyword.value,
      category: filters.category ?? (category.value === '__uncategorized__' ? '' : category.value),
      type: filters.type ?? questionType.value,
    })
    items.push(...listFromResponse(res))
  }
  return items
}

async function fetchExportQuestions(scope: BankScope) {
  return (await fetchAllQuestionRows(scope)).map(normalizeExportQuestion)
}

async function exportQuestions() {
  const scope = activeScope.value
  exporting.value = true
  try {
    const items = await fetchExportQuestions(scope)
    if (!items.length) {
      ElMessage.warning('暂无可导出的题目')
      return
    }
    downloadJson(exportFilename(scope), buildQuestionBankExportPayload({
      scope,
      scopeName: scopeName(scope),
      filters: {
        keyword: keyword.value,
        category: category.value,
        type: questionType.value
      },
      items
    }))
    ElMessage.success(`已导出 ${items.length} 道题`)
  } catch {
    ElMessage.error('导出失败，请稍后重试')
  } finally {
    exporting.value = false
  }
}

function triggerImport() {
  if (importing.value) return
  if (importFileInput.value) importFileInput.value.value = ''
  importFileInput.value?.click()
}

async function handleImportFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  let items: QuestionPayload[] = []
  try {
    items = parseImportQuestions(await file.text(), category.value)
  } catch (err: any) {
    ElMessage.error(err?.message || '导入文件解析失败')
    return
  }
  if (!items.length) {
    ElMessage.warning('文件中没有可导入的题目')
    return
  }
  const scope = activeScope.value
  try {
    await ElMessageBox.confirm(`将 ${items.length} 道题导入到「${scopeName(scope)}」，是否继续？`, '导入题目', { type: 'warning' })
  } catch {
    return
  }
  importing.value = true
  let successCount = 0
  try {
    for (const item of items) {
      await insertQuestion(scope, item)
      successCount++
    }
    ElMessage.success(`已导入 ${items.length} 道题`)
    await loadCategories()
    await load()
  } catch {
    ElMessage.error(`导入失败，已成功导入 ${successCount}/${items.length} 道题`)
  } finally {
    importing.value = false
  }
}

async function deleteRow(row: any) {
  if (!canDeleteQuestion.value) return
  try {
    await ElMessageBox.confirm(`确认删除「${questionTitle(row)}」?`, '删除题目', { type: 'warning' })
  } catch {
    return
  }
  clearAutoSaveTimer()
  if (activeScope.value === 'survey') {
    await adminApi.surveyQuestionBankDel({ id: row.id })
  } else {
    await adminApi.examQuestionBankDel({ id: row.id })
  }
  ElMessage.success('已删除')
  if (list.value.length === 1 && page.value > 1) page.value--
  if (Number(row.id || 0) === form.id) clearCurrentQuestion()
  await loadCategories()
  await load()
}

async function deleteCurrentQuestion() {
  if (!form.id) return
  const row = list.value.find(item => Number(item.id) === form.id) || { id: form.id, title: form.title }
  await deleteRow(row)
}

function handleWorkspaceKeydown(event: KeyboardEvent) {
  if (!editDialog.visible || !(event.metaKey || event.ctrlKey) || event.key.toLowerCase() !== 's') return
  event.preventDefault()
  void save(false)
}

function applyResponsivePanelDefaults() {
  rightCollapsed.value = window.innerWidth < 1200
  leftCollapsed.value = window.innerWidth < 900
}

function toggleLeftPanel() {
  const opening = leftCollapsed.value
  leftCollapsed.value = !leftCollapsed.value
  if (opening && window.innerWidth < 1200) rightCollapsed.value = true
}

function toggleRightPanel() {
  const opening = rightCollapsed.value
  rightCollapsed.value = !rightCollapsed.value
  if (opening && window.innerWidth < 1200) leftCollapsed.value = true
}

watch(form, scheduleAutoSave, { deep: true })
watch(visualQuestion, scheduleAutoSave, { deep: true })
watch(editorMode, value => handleEditorModeChange(value))

onMounted(async () => {
  window.addEventListener('keydown', handleWorkspaceKeydown)
  applyResponsivePanelDefaults()
  await loadCategories()
  await load()
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleWorkspaceKeydown)
  clearAutoSaveTimer()
})
</script>

<style scoped src="./question-bank-workspace.css"></style>
