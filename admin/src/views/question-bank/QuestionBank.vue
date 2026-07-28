<template>
  <div class="admin-page question-bank-page">
    <el-card class="admin-card" shadow="never">
      <div class="bank-header">
        <el-tabs v-model="activeScope" class="bank-tabs" @tab-change="handleScopeChange">
          <el-tab-pane label="问卷题库" name="survey" />
          <el-tab-pane label="考试题库" name="exam" />
        </el-tabs>
      </div>

      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-input v-model="keyword" placeholder="搜索题目标题/题型" clearable style="width:300px" @keyup.enter="search" />
          <el-select v-model="category" placeholder="分类" clearable filterable allow-create style="width:160px" @change="search">
            <el-option v-for="item in categories" :key="item" :label="item" :value="item" />
          </el-select>
          <el-select v-model="questionType" placeholder="题型" clearable filterable style="width:150px" @change="search">
            <el-option v-for="item in typeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-button type="primary" @click="search">搜索</el-button>
        </div>
      </div>

      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-button type="primary" @click="openCreate">+ 新增题目</el-button>
          <el-button icon="Upload" :loading="importing" @click="triggerImport">导入</el-button>
          <el-button icon="Download" :loading="exporting" @click="exportQuestions">导出</el-button>
          <input ref="importFileInput" class="bank-import-input" type="file" accept=".json,application/json" @change="handleImportFile" />
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" @click="load" />
        </div>
      </div>

      <div class="bank-stat-bar">
        <div class="bank-stat"><strong>{{ total }}</strong> 总题目</div>
        <div class="bank-stat"><strong>{{ list.length }}</strong> 当前页</div>
        <div class="bank-stat"><strong>{{ categories.length }}</strong> 分类</div>
      </div>

      <QuestionBankTable
        v-model:page="page"
        v-model:page-size="pageSize"
        :list="list"
        :loading="loading"
        :active-scope="activeScope"
        :total="total"
        :question-title="questionTitle"
        :tag-list="tagList"
        :type-name="typeName"
        :format-time="formatTime"
        @load="load"
        @size-change="handleSizeChange"
        @preview="openPreview"
        @edit="openEdit"
        @delete="deleteRow"
      />
    </el-card>

    <QuestionEditorDialog
      v-model:visible="editDialog.visible"
      v-model:editor-mode="editorMode"
      :is-edit="editDialog.isEdit"
      :form="form"
      :visual-question="visualQuestion"
      :categories="categories"
      :type-options="typeOptions"
      :active-scope="activeScope"
      :saving="saving"
      :rich-title-editor-options="richTitleEditorOptions"
      :rich-option-editor-options="richOptionEditorOptions"
      :rich-full-editor-options="richFullEditorOptions"
      :rich-edit-dialog="richEditDialog"
      :rich-edit-title="richEditTitle"
      :bind-compact-rich-editor="bindCompactRichEditor"
      :sync-plain-title-from-visual="syncPlainTitleFromVisual"
      :handle-type-change="handleTypeChange"
      :handle-editor-mode-change="handleEditorModeChange"
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
      :type-name="typeName"
      :preview-placeholder="previewPlaceholder"
      :format-schema="formatSchema"
      :apply-json-to-visual="applyJsonToVisual"
      :confirm-rich-full-edit="confirmRichFullEdit"
      :save="save"
    />

    <QuestionPreviewDrawer
      v-model:visible="preview.visible"
      :row="preview.row"
      :question-title="questionTitle"
      :type-name="typeName"
      :format-time="formatTime"
      :pretty-schema="prettySchema"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'
import QuestionBankTable from './components/QuestionBankTable.vue'
import QuestionEditorDialog from './components/QuestionEditorDialog.vue'
import QuestionPreviewDrawer from './components/QuestionPreviewDrawer.vue'
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
const saving = ref(false)
const importing = ref(false)
const exporting = ref(false)
const list = ref<any[]>([])
const categories = ref<string[]>([])
const importFileInput = ref<HTMLInputElement | null>(null)
const editorMode = ref<'visual' | 'json'>('visual')
const visualQuestion = ref<any>({ props: {} })

const editDialog = reactive({ visible: false, isEdit: false })
const form = reactive({
  id: 0,
  title: '',
  type: '',
  category: '',
  tags: '',
  schema: ''
})
const preview = reactive<{ visible: boolean; row: any | null }>({ visible: false, row: null })
const richEditDialog = reactive({
  visible: false,
  target: 'title' as RichEditTarget,
  optionIndex: -1,
  content: ''
})

const typeMap: Record<string, string> = {
  radio: '单选题',
  checkbox: '多选题',
  select: '下拉题',
  picker: '选择器',
  cascade: '级联选择',
  judge: '判断题',
  file: '文件/图片',
  input: '单行文本',
  textarea: '多行文本',
  number: '数字',
  multiInput: '多项填空',
  hInput: '横向填空',
  signature: '签名',
  scanCode: '扫码',
  rating: '评分',
  nps: 'NPS评分',
  matrixRadio: '矩阵单选',
  matrixCheckbox: '矩阵多选',
  matrixFillBlank: '矩阵填空',
  matrixAuto: '表格自增',
  divider: '分割线',
  description: '文字描述',
  questionSet: '问题组',
  pagination: '分页',
  user: '成员',
  dept: '部门',
  richText: '富文本',
  autopop: '自动填充',
  name: '姓名',
  studentId: '学号',
  employeeId: '工号',
  class: '班级',
  phone: '手机',
  email: '邮箱',
  idCard: '身份证',
  password: '密码',
  date: '日期',
  time: '时间',
  dateRange: '日期范围',
  switch: '开关',
  location: '地理位置'
}

const typeOptions = computed(() => Object.entries(typeMap).map(([value, label]) => ({ value, label })))
const optionTypes = new Set(['radio', 'checkbox', 'select', 'picker', 'cascade', 'judge'])
const fieldTypes = new Set(['multiInput', 'hInput'])
const matrixTypes = new Set(['matrixRadio', 'matrixCheckbox', 'matrixFillBlank', 'matrixAuto'])
const layoutTypes = new Set(['description', 'questionSet', 'pagination', 'divider'])
const richTitleEditorOptions = {
  theme: 'snow',
  placeholder: '请输入题目标题',
  modules: {
    toolbar: [
      ['bold', 'italic', 'underline', 'strike'],
      [{ color: [] }, { background: [] }],
      [{ list: 'ordered' }, { list: 'bullet' }],
      ['clean']
    ]
  }
}
const richOptionEditorOptions = {
  theme: 'snow',
  placeholder: '选项文本',
  modules: {
    toolbar: [
      ['bold', 'italic', 'underline'],
      [{ color: [] }, { background: [] }],
      ['clean']
    ]
  }
}
const richFullEditorOptions = {
  theme: 'snow',
  placeholder: '请输入富文本内容',
  modules: {
    toolbar: [
      [{ header: [1, 2, 3, false] }],
      ['bold', 'italic', 'underline', 'strike'],
      [{ color: [] }, { background: [] }],
      [{ list: 'ordered' }, { list: 'bullet' }],
      [{ indent: '-1' }, { indent: '+1' }],
      [{ align: [] }],
      ['blockquote', 'code-block'],
      ['link'],
      ['clean']
    ]
  }
}
const richEditTitle = computed(() => richEditDialog.target === 'title' ? '完整编辑题目标题' : `完整编辑选项 ${richEditDialog.optionIndex + 1}`)

function scopeName(scope: BankScope) {
  return scope === 'survey' ? '问卷题库' : '考试题库'
}

function typeName(type: string) {
  return typeMap[type] || type || '-'
}

function stripHtml(value: string) {
  return String(value || '').replace(/<[^>]*>/g, '').replace(/&nbsp;/g, ' ').trim()
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

function tagList(tags: string) {
  return String(tags || '').split(',').map(item => item.trim()).filter(Boolean)
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

function clonePlain<T>(value: T): T {
  return JSON.parse(JSON.stringify(value ?? {}))
}

function usesOptions(type: string) {
  return optionTypes.has(type)
}

function usesFields(type: string) {
  return fieldTypes.has(type)
}

function usesMatrix(type: string) {
  return matrixTypes.has(type)
}

function previewPlaceholder(type: string) {
  if (type === 'date') return '选择日期'
  if (type === 'time') return '选择时间'
  if (type === 'file') return '上传文件'
  if (type === 'signature') return '签名'
  if (type === 'scanCode') return '扫码'
  return '请输入'
}

function optionValue(index: number) {
  return String.fromCharCode(65 + index)
}

function normalizeOption(option: any, index: number) {
  return {
    ...clonePlain(option || {}),
    label: String(option?.label ?? option?.title ?? option?.text ?? `选项${optionValue(index)}`),
    value: String(option?.value ?? optionValue(index))
  }
}

function normalizeField(field: any, index: number) {
  return {
    ...clonePlain(field || {}),
    label: String(field?.label ?? field?.title ?? `填空${index + 1}`),
    placeholder: String(field?.placeholder ?? '')
  }
}

function normalizeMatrixItem(item: any, index: number, prefix: string) {
  return {
    ...clonePlain(item || {}),
    id: String(item?.id || `${prefix}${index + 1}`),
    title: String(item?.title ?? item?.label ?? `${prefix === 'r' ? '行' : '列'}${index + 1}`)
  }
}

function ensureVisualQuestionShape(question: any) {
  const q = clonePlain(question || {})
  q.type = String(q.type || form.type || 'radio')
  q.title = String(q.title || form.title || typeName(q.type) || '未命名题目')
  q.required = Boolean(q.required)
  q.readOnly = Boolean(q.readOnly)
  q.placeholder = String(q.placeholder || '')
  q.description = String(q.description || '')
  q.showDescription = q.showDescription ?? !layoutTypes.has(q.type)
  q.props = q.props && typeof q.props === 'object' && !Array.isArray(q.props) ? q.props : {}
  if (usesOptions(q.type)) {
    const defaults = q.type === 'judge'
      ? [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
      : [{ label: '选项A', value: 'A' }, { label: '选项B', value: 'B' }]
    q.props.options = Array.isArray(q.props.options) && q.props.options.length
      ? q.props.options.map(normalizeOption)
      : defaults
  }
  if (usesFields(q.type)) {
    const source = Array.isArray(q.props.fields) && q.props.fields.length ? q.props.fields : q.props.options
    q.props.fields = Array.isArray(source) && source.length
      ? source.map(normalizeField)
      : [{ label: '填空1', placeholder: '' }, { label: '填空2', placeholder: '' }]
  }
  if (usesMatrix(q.type)) {
    q.props.rows = Array.isArray(q.props.rows) && q.props.rows.length
      ? q.props.rows.map((item: any, index: number) => normalizeMatrixItem(item, index, 'r'))
      : [{ id: 'r1', title: '行1' }, { id: 'r2', title: '行2' }]
    q.props.columns = Array.isArray(q.props.columns) && q.props.columns.length
      ? q.props.columns.map((item: any, index: number) => normalizeMatrixItem(item, index, 'c'))
      : [{ id: 'c1', title: '列A' }, { id: 'c2', title: '列B' }]
    if (q.type === 'matrixAuto') q.props.rows = Array.isArray(q.props.rows) ? q.props.rows : []
  }
  if (q.type === 'file') {
    q.fileTypes = Array.isArray(q.fileTypes) ? q.fileTypes : ['image']
    q.maxFileSize = Number(q.maxFileSize || 10)
    q.maxFileCount = Number(q.maxFileCount || 1)
  }
  if (q.type === 'rating') q.props.maxRating = Number(q.props.maxRating || 5)
  if (q.type === 'nps') q.props.maxRating = Number(q.props.maxRating || 10)
  if (activeScope.value === 'exam') {
    q.examScore = Number(q.examScore || 0)
    q.examCorrectAnswer = String(q.examCorrectAnswer || '')
    q.examAnalysis = String(q.examAnalysis || '')
  }
  return q
}

function parseSchemaQuestion(schema: any, fallback: any = {}) {
  let parsed = schema
  if (typeof schema === 'string' && schema.trim()) parsed = JSON.parse(schema)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return fallback
  if (Array.isArray(parsed.questions)) return parsed.questions[0] || fallback
  return parsed
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
  options.push({ label: `选项${optionValue(options.length)}`, value: optionValue(options.length) })
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
    category: category.value,
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
  try {
    const res: any = await requestQuestionBankList(activeScope.value, buildListParams())
    list.value = listFromResponse(res)
    total.value = Number(res.data?.total || 0)
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

function handleSizeChange() {
  page.value = 1
  load()
}

async function handleScopeChange() {
  page.value = 1
  keyword.value = ''
  category.value = ''
  questionType.value = ''
  list.value = []
  total.value = 0
  await loadCategories()
  await load()
}

function resetForm() {
  form.id = 0
  form.title = '未命名题目'
  form.type = 'radio'
  form.category = category.value || ''
  form.tags = ''
  visualQuestion.value = ensureVisualQuestionShape({ title: form.title, type: form.type })
  syncSchemaFromVisual()
  editorMode.value = 'visual'
}

function openCreate() {
  resetForm()
  editDialog.isEdit = false
  editDialog.visible = true
}

function openEdit(row: any) {
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
}

function openPreview(row: any) {
  preview.row = row
  preview.visible = true
}

function formatSchema() {
  if (!form.schema.trim()) return
  try {
    form.schema = JSON.stringify(JSON.parse(form.schema), null, 2)
  } catch {
    ElMessage.warning('当前内容不是有效 JSON')
  }
}

async function save() {
  if (editorMode.value === 'visual') syncSchemaFromVisual()
  const plainTitle = stripHtml(form.title || visualQuestion.value.title || '').trim()
  if (!plainTitle) {
    ElMessage.warning('请输入题目标题')
    return
  }
  if (!form.type) {
    ElMessage.warning('请选择题型')
    return
  }
  saving.value = true
  try {
    const payload = {
      id: form.id,
      title: plainTitle,
      type: form.type,
      category: form.category || '',
      tags: form.tags || '',
      schema: form.schema || ''
    }
    if (activeScope.value === 'survey') {
      if (editDialog.isEdit) await adminApi.surveyQuestionBankEdit(payload)
      else await adminApi.surveyQuestionBankInsert(payload)
    } else {
      if (editDialog.isEdit) await adminApi.examQuestionBankEdit(payload)
      else await adminApi.examQuestionBankInsert(payload)
    }
    ElMessage.success('已保存')
    editDialog.visible = false
    await loadCategories()
    await load()
  } finally {
    saving.value = false
  }
}

async function fetchExportQuestions(scope: BankScope) {
  const pageSizeForExport = 500
  const first: any = await requestQuestionBankList(scope, buildListParams({ page: 1, pageSize: pageSizeForExport }))
  const items = [...listFromResponse(first)]
  const allTotal = Number(first.data?.total || items.length)
  const pageCount = Math.ceil(allTotal / pageSizeForExport)
  for (let currentPage = 2; currentPage <= pageCount; currentPage++) {
    const res: any = await requestQuestionBankList(scope, buildListParams({ page: currentPage, pageSize: pageSizeForExport }))
    items.push(...listFromResponse(res))
  }
  return items.map(normalizeExportQuestion)
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
  try {
    await ElMessageBox.confirm(`确认删除「${questionTitle(row)}」?`, '删除题目', { type: 'warning' })
  } catch {
    return
  }
  if (activeScope.value === 'survey') {
    await adminApi.surveyQuestionBankDel({ id: row.id })
  } else {
    await adminApi.examQuestionBankDel({ id: row.id })
  }
  ElMessage.success('已删除')
  if (list.value.length === 1 && page.value > 1) page.value--
  await loadCategories()
  await load()
}

onMounted(async () => {
  await loadCategories()
  await load()
})
</script>

<style scoped>
.question-bank-page {
  --bank-accent: #2563eb;
  --bank-accent-soft: #eff6ff;
  --bank-border: #e5e7eb;
}

.bank-header {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--bank-border);
}

.bank-tabs {
  width: 260px;
  max-width: 100%;
}

.bank-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
}

.bank-tabs :deep(.el-tabs__nav-wrap::after) {
  display: none;
}

.bank-import-input {
  display: none;
}

.bank-stat-bar {
  display: flex;
  align-items: center;
  gap: 22px;
  height: 40px;
  margin-bottom: 14px;
  padding: 0 12px;
  border-radius: 8px;
  background: #f8fafc;
  color: #667085;
  font-size: 13px;
}

.bank-stat strong {
  margin-right: 4px;
  color: var(--bank-accent);
  font-size: 17px;
}


@media (max-width: 768px) {
  .bank-tabs {
    width: 100%;
  }
}
</style>
