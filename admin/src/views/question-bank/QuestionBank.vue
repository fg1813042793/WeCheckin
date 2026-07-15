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

      <el-table :data="list" v-loading="loading" stripe style="width:100%" empty-text="暂无题目">
        <el-table-column prop="id" label="ID" width="76" />
        <el-table-column label="题目" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="bank-question-title">{{ questionTitle(row) }}</div>
            <div v-if="row.tags || row.category" class="bank-question-meta">
              <el-tag v-if="row.category" size="small" round>{{ row.category }}</el-tag>
              <span v-for="tag in tagList(row.tags)" :key="tag" class="bank-tag">{{ tag }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="题型" width="150">
          <template #default="{ row }">
            <span class="bank-type-cell">
              <question-icon :type="row.type" class="bank-type-icon" />
              <span>{{ typeName(row.type) }}</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="100">
          <template #default>{{ activeScope === 'survey' ? '问卷' : '考试' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 1 ? 'success' : 'info'" round>{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ formatTime(row.addTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-button size="small" @click="openPreview(row)">预览</el-button>
              <el-button size="small" type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="danger" plain @click="deleteRow(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="admin-pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10,20,50,100]"
          :total="total"
          layout="total,sizes,prev,pager,next"
          @current-change="load"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="editDialog.visible" :title="editDialog.isEdit ? '编辑题目' : '新增题目'" width="900px" :close-on-click-modal="false" class="question-edit-dialog">
      <el-form label-position="top" class="question-edit-form">
        <el-form-item label="题目标题">
          <div class="bank-rich-editor bank-rich-title">
            <QuillEditor
              v-model:content="visualQuestion.title"
              content-type="html"
              :options="richTitleEditorOptions"
              @update:content="syncPlainTitleFromVisual"
              @ready="(quill: any) => bindCompactRichEditor(quill, 'title')"
            />
          </div>
        </el-form-item>
        <el-form-item label="题型">
          <el-select v-model="form.type" placeholder="选择题型" filterable style="width:100%" @change="handleTypeChange">
            <el-option v-for="item in typeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.category" placeholder="选择或输入分类" filterable allow-create clearable style="width:100%">
            <el-option v-for="item in categories" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.tags" placeholder="多个标签用逗号分隔" />
        </el-form-item>
        <el-form-item label="题目配置">
          <el-tabs v-model="editorMode" class="schema-tabs" @tab-change="handleEditorModeChange">
            <el-tab-pane label="可视化编辑" name="visual">
              <div class="visual-editor">
                <div class="visual-main">
                  <div class="visual-section">
                    <div class="visual-section-title">基础设置</div>
                    <div class="visual-grid">
                      <div class="visual-field">
                        <span>必填</span>
                        <el-switch v-model="visualQuestion.required" />
                      </div>
                      <div class="visual-field">
                        <span>只读</span>
                        <el-switch v-model="visualQuestion.readOnly" />
                      </div>
                      <div class="visual-field wide">
                        <span>提示语</span>
                        <el-input v-model="visualQuestion.placeholder" placeholder="请输入提示语" />
                      </div>
                      <div class="visual-field wide">
                        <span>说明</span>
                        <el-input v-model="visualQuestion.description" type="textarea" :rows="2" placeholder="请输入题目说明" />
                      </div>
                    </div>
                  </div>

                  <div v-if="usesOptions(visualQuestion.type)" class="visual-section">
                    <div class="visual-section-title">选项设置</div>
                    <div class="visual-option-list">
                      <div v-for="(opt, idx) in visualQuestion.props.options" :key="idx" class="visual-option-row">
                        <span class="option-index">{{ idx + 1 }}</span>
                        <div class="bank-rich-editor bank-rich-option">
                          <QuillEditor
                            v-model:content="opt.label"
                            content-type="html"
                            :options="richOptionEditorOptions"
                            @ready="(quill: any) => bindCompactRichEditor(quill, 'option', idx)"
                          />
                        </div>
                        <el-input v-model="opt.value" placeholder="选项值" />
                        <el-button icon="Delete" circle plain type="danger" @click="removeVisualOption(idx)" />
                      </div>
                    </div>
                    <el-button size="small" @click="addVisualOption">+ 添加选项</el-button>
                  </div>

                  <div v-if="usesFields(visualQuestion.type)" class="visual-section">
                    <div class="visual-section-title">字段设置</div>
                    <div class="visual-option-list">
                      <div v-for="(field, idx) in visualQuestion.props.fields" :key="idx" class="visual-option-row">
                        <span class="option-index">{{ idx + 1 }}</span>
                        <el-input v-model="field.label" placeholder="字段名称" />
                        <el-input v-model="field.placeholder" placeholder="提示语" />
                        <el-button icon="Delete" circle plain type="danger" @click="removeVisualField(idx)" />
                      </div>
                    </div>
                    <el-button size="small" @click="addVisualField">+ 添加字段</el-button>
                  </div>

                  <div v-if="usesMatrix(visualQuestion.type)" class="visual-section">
                    <div class="visual-section-title">矩阵设置</div>
                    <div class="matrix-editor-grid">
                      <div>
                        <div class="matrix-subtitle">行</div>
                        <div class="visual-option-list">
                          <div v-for="(row, idx) in visualQuestion.props.rows" :key="idx" class="visual-option-row compact">
                            <span class="option-index">{{ idx + 1 }}</span>
                            <el-input v-model="row.title" placeholder="行标题" />
                            <el-button icon="Delete" circle plain type="danger" @click="removeMatrixRow(idx)" />
                          </div>
                        </div>
                        <el-button size="small" @click="addMatrixRow">+ 添加行</el-button>
                      </div>
                      <div>
                        <div class="matrix-subtitle">列</div>
                        <div class="visual-option-list">
                          <div v-for="(col, idx) in visualQuestion.props.columns" :key="idx" class="visual-option-row compact">
                            <span class="option-index">{{ idx + 1 }}</span>
                            <el-input v-model="col.title" placeholder="列标题" />
                            <el-button icon="Delete" circle plain type="danger" @click="removeMatrixColumn(idx)" />
                          </div>
                        </div>
                        <el-button size="small" @click="addMatrixColumn">+ 添加列</el-button>
                      </div>
                    </div>
                  </div>

                  <div v-if="visualQuestion.type === 'file'" class="visual-section">
                    <div class="visual-section-title">上传设置</div>
                    <div class="visual-grid">
                      <div class="visual-field wide">
                        <span>文件类型</span>
                        <el-select v-model="visualQuestion.fileTypes" multiple clearable style="width:100%">
                          <el-option label="图片" value="image" />
                          <el-option label="文档" value="document" />
                          <el-option label="视频" value="video" />
                          <el-option label="音频" value="audio" />
                        </el-select>
                      </div>
                      <div class="visual-field">
                        <span>大小 MB</span>
                        <el-input-number v-model="visualQuestion.maxFileSize" :min="1" :max="200" controls-position="right" style="width:100%" />
                      </div>
                      <div class="visual-field">
                        <span>数量</span>
                        <el-input-number v-model="visualQuestion.maxFileCount" :min="1" :max="20" controls-position="right" style="width:100%" />
                      </div>
                    </div>
                  </div>

                  <div v-if="visualQuestion.type === 'rating' || visualQuestion.type === 'nps'" class="visual-section">
                    <div class="visual-section-title">评分设置</div>
                    <div class="visual-grid">
                      <div class="visual-field">
                        <span>最大分值</span>
                        <el-input-number v-model="visualQuestion.props.maxRating" :min="1" :max="20" controls-position="right" style="width:100%" />
                      </div>
                    </div>
                  </div>

                  <div v-if="activeScope === 'exam'" class="visual-section">
                    <div class="visual-section-title">考试设置</div>
                    <div class="visual-grid">
                      <div class="visual-field">
                        <span>分值</span>
                        <el-input-number v-model="visualQuestion.examScore" :min="0" :step="1" controls-position="right" style="width:100%" />
                      </div>
                      <div class="visual-field wide">
                        <span>正确答案</span>
                        <el-input v-model="visualQuestion.examCorrectAnswer" placeholder="请输入正确答案" />
                      </div>
                      <div class="visual-field wide">
                        <span>解析</span>
                        <el-input v-model="visualQuestion.examAnalysis" type="textarea" :rows="2" placeholder="请输入答案解析" />
                      </div>
                    </div>
                  </div>
                </div>

                <div class="visual-preview">
                  <div class="visual-preview-title">
                    <question-icon :type="visualQuestion.type" class="bank-type-icon" />
                    <span v-html="visualQuestion.title || typeName(visualQuestion.type)"></span>
                  </div>
                  <div class="visual-preview-body">
                    <template v-if="usesOptions(visualQuestion.type)">
                      <div v-for="(opt, idx) in visualQuestion.props.options" :key="idx" class="preview-option">
                        <span class="preview-dot" />
                        <span v-html="opt.label || `选项${idx + 1}`"></span>
                      </div>
                    </template>
                    <template v-else-if="usesFields(visualQuestion.type)">
                      <div v-for="(field, idx) in visualQuestion.props.fields" :key="idx" class="preview-field">
                        <span>{{ field.label || `字段${idx + 1}` }}</span>
                        <el-input :placeholder="field.placeholder || '请输入'" disabled />
                      </div>
                    </template>
                    <template v-else-if="usesMatrix(visualQuestion.type)">
                      <div class="preview-matrix">
                        <div class="preview-matrix-row header">
                          <span>行/列</span>
                          <span v-for="(col, idx) in visualQuestion.props.columns" :key="idx">{{ col.title || `列${idx + 1}` }}</span>
                        </div>
                        <div v-for="(row, idx) in visualQuestion.props.rows" :key="idx" class="preview-matrix-row">
                          <span>{{ row.title || `行${idx + 1}` }}</span>
                          <span v-for="(_col, cidx) in visualQuestion.props.columns" :key="cidx">-</span>
                        </div>
                      </div>
                    </template>
                    <el-input v-else-if="visualQuestion.type === 'textarea'" type="textarea" :placeholder="visualQuestion.placeholder || '请输入'" disabled />
                    <el-input v-else :placeholder="visualQuestion.placeholder || previewPlaceholder(visualQuestion.type)" disabled />
                  </div>
                </div>
              </div>
            </el-tab-pane>
            <el-tab-pane label="高级 JSON" name="json">
              <div class="schema-editor">
                <el-input v-model="form.schema" type="textarea" :rows="14" placeholder="题目 schema JSON，可从设计器上传后编辑" />
                <div class="schema-actions">
                  <el-button size="small" @click="formatSchema">格式化 JSON</el-button>
                  <el-button size="small" @click="applyJsonToVisual">同步到可视化</el-button>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialog.visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="richEditDialog.visible"
      :title="richEditTitle"
      width="760px"
      append-to-body
      :close-on-click-modal="false"
      class="bank-rich-full-dialog"
      @closed="richEditDialog.content=''"
    >
      <div class="bank-rich-full-editor">
        <QuillEditor
          v-model:content="richEditDialog.content"
          content-type="html"
          :options="richFullEditorOptions"
        />
      </div>
      <template #footer>
        <el-button @click="richEditDialog.visible=false">取消</el-button>
        <el-button type="primary" @click="confirmRichFullEdit">确定</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="preview.visible" title="题目预览" size="520px">
      <div v-if="preview.row" class="preview-panel">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="ID">{{ preview.row.id }}</el-descriptions-item>
          <el-descriptions-item label="标题">{{ questionTitle(preview.row) }}</el-descriptions-item>
          <el-descriptions-item label="题型">{{ typeName(preview.row.type) }}</el-descriptions-item>
          <el-descriptions-item label="分类">{{ preview.row.category || '-' }}</el-descriptions-item>
          <el-descriptions-item label="标签">{{ preview.row.tags || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(preview.row.addTime) }}</el-descriptions-item>
        </el-descriptions>
        <div class="preview-schema-title">题目配置</div>
        <pre class="preview-schema">{{ prettySchema(preview.row.schema) }}</pre>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'
import QuestionIcon from '../survey/formkit/QuestionIcon.vue'

type BankScope = 'survey' | 'exam'
type QuestionPayload = {
  title: string
  type: string
  category: string
  tags: string
  schema: string
}
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

function normalizeSchemaValue(value: any) {
  if (value === undefined || value === null || value === '') return ''
  if (typeof value === 'string') return value
  return JSON.stringify(value)
}

function normalizeExportQuestion(row: any): QuestionPayload {
  return {
    title: String(row.title || ''),
    type: String(row.type || ''),
    category: String(row.category || ''),
    tags: String(row.tags || ''),
    schema: normalizeSchemaValue(row.schema)
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

function downloadJson(filename: string, data: any) {
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

function exportFilename(scope: BankScope) {
  const stamp = new Date().toISOString().replace(/[:.]/g, '-')
  return `${scope === 'survey' ? 'survey' : 'exam'}-question-bank-${stamp}.json`
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
    downloadJson(exportFilename(scope), {
      version: '1.0',
      module: 'question-bank',
      scope,
      scopeName: scopeName(scope),
      exportedAt: new Date().toISOString(),
      filters: {
        keyword: keyword.value,
        category: category.value,
        type: questionType.value
      },
      total: items.length,
      items
    })
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

function extractImportRows(raw: any) {
  if (Array.isArray(raw)) return raw
  const candidates = [raw?.items, raw?.list, raw?.questions, raw?.data?.items, raw?.data?.list]
  const rows = candidates.find(Array.isArray)
  if (!rows) throw new Error('未找到题目数组，请选择题库导出的 JSON 文件')
  return rows
}

function normalizeImportQuestion(row: any, index: number): QuestionPayload {
  const title = String(row?.title ?? row?.label ?? row?.question ?? row?.name ?? '').trim()
  const type = String(row?.type ?? row?.qType ?? row?.questionType ?? '').trim()
  if (!title || !type) throw new Error(`第 ${index + 1} 题缺少标题或题型`)
  return {
    title,
    type,
    category: String(row?.category || category.value || ''),
    tags: String(row?.tags || ''),
    schema: normalizeSchemaValue(row?.schema ?? row)
  }
}

function parseImportQuestions(text: string) {
  let raw: any
  try {
    raw = JSON.parse(text)
  } catch {
    throw new Error('JSON 文件格式不正确')
  }
  return extractImportRows(raw).map((row: any, index: number) => normalizeImportQuestion(row, index))
}

async function handleImportFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  let items: QuestionPayload[] = []
  try {
    items = parseImportQuestions(await file.text())
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

.bank-question-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #1f2937;
  font-weight: 600;
}

.bank-question-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 5px;
  min-width: 0;
}

.bank-tag {
  overflow: hidden;
  max-width: 96px;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #64748b;
  font-size: 12px;
}

.bank-type-cell {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
}

.bank-type-icon {
  width: 16px;
  height: 16px;
  color: #94a3b8;
}

.schema-editor {
  width: 100%;
}

.schema-editor :deep(.el-textarea__inner),
.preview-schema {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 12px;
  line-height: 1.6;
}

.schema-actions {
  margin-top: 8px;
}

.schema-tabs {
  width: 100%;
}

.bank-rich-editor {
  display: block;
  width: 100%;
  min-width: 0;
  overflow: hidden;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #fff;
  transition: border-color 0.18s, box-shadow 0.18s;
}

.bank-rich-editor:hover {
  border-color: #c0c4cc;
}

.bank-rich-editor:focus-within {
  border-color: var(--bank-accent);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08);
}

.bank-rich-editor :deep(.ql-toolbar) {
  display: flex;
  flex-wrap: wrap;
  box-sizing: border-box;
  width: 100%;
  max-height: 0;
  overflow: hidden;
  padding: 0 8px;
  border: 0;
  border-bottom: 0 solid #eef2f7;
  background: #fff;
  opacity: 0;
  transition: max-height 0.18s ease, padding 0.18s ease, opacity 0.18s ease;
}

.bank-rich-editor:focus-within :deep(.ql-toolbar) {
  max-height: 74px;
  padding: 6px 8px;
  border-bottom-width: 1px;
  opacity: 1;
}

.bank-rich-editor :deep(.ql-toolbar .ql-formats) {
  margin-right: 6px;
}

.bank-rich-editor :deep(.ql-toolbar button),
.bank-rich-editor :deep(.ql-toolbar .ql-picker-label) {
  border-radius: 5px;
}

.bank-rich-editor :deep(.ql-toolbar .ql-bank-full) {
  width: 68px;
  padding: 0 6px;
  color: var(--bank-accent);
  font-size: 12px;
  font-weight: 600;
}

.bank-rich-editor :deep(.ql-toolbar button:hover),
.bank-rich-editor :deep(.ql-toolbar button.ql-active),
.bank-rich-editor :deep(.ql-toolbar .ql-picker-label:hover) {
  background: #f1f5f9;
}

.bank-rich-editor :deep(.ql-container) {
  display: block;
  box-sizing: border-box;
  width: 100%;
  height: auto;
  border: 0;
  border-radius: 8px;
  font-size: 13px;
  font-family: inherit;
}

.bank-rich-title :deep(.ql-editor) {
  min-height: 44px;
  padding: 10px 12px;
}

.bank-rich-option :deep(.ql-editor) {
  min-height: 38px;
  padding: 8px 10px;
}

.bank-rich-title :deep(.ql-editor p),
.bank-rich-option :deep(.ql-editor p) {
  margin: 0;
}

.bank-rich-title :deep(.ql-editor.ql-blank::before),
.bank-rich-option :deep(.ql-editor.ql-blank::before) {
  left: 12px;
  color: #a8b0bd;
  font-style: normal;
}

:global(.bank-rich-full-dialog .el-dialog__body) {
  padding-top: 8px;
}

:global(.bank-rich-full-editor) {
  display: flex;
  flex-direction: column;
  height: min(58vh, 560px);
  min-height: 420px;
}

:global(.bank-rich-full-editor .ql-toolbar) {
  flex-shrink: 0;
  border-color: #dcdfe6;
  border-radius: 8px 8px 0 0;
  background: #fff;
}

:global(.bank-rich-full-editor .ql-container) {
  flex: 1;
  min-height: 0;
  border-color: #dcdfe6;
  border-radius: 0 0 8px 8px;
  font-size: 14px;
}

:global(.bank-rich-full-editor .ql-editor) {
  min-height: 100%;
  padding: 14px 16px;
}

.question-edit-form :deep(.el-form-item__content) {
  min-width: 0;
}

.question-edit-form :deep(.el-form-item) {
  margin-bottom: 14px;
}

.question-edit-form :deep(.el-form-item__label) {
  margin-bottom: 6px;
  color: #475467;
  font-weight: 600;
  line-height: 1.3;
}

.visual-editor {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 14px;
  width: 100%;
}

.visual-main {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.visual-section,
.visual-preview {
  border: 1px solid var(--bank-border);
  border-radius: 8px;
  background: #fff;
}

.visual-section {
  padding: 12px;
}

.visual-section-title {
  margin-bottom: 10px;
  color: #344054;
  font-size: 13px;
  font-weight: 650;
}

.visual-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.visual-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.visual-field > span,
.matrix-subtitle {
  color: #667085;
  font-size: 12px;
}

.visual-field.wide {
  grid-column: 1 / -1;
}

.visual-option-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 8px;
}

.visual-option-row {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) minmax(0, 120px) 34px;
  gap: 8px;
  align-items: start;
}

.visual-option-row.compact {
  grid-template-columns: 28px minmax(0, 1fr) 34px;
}

.option-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  margin-top: 7px;
  border-radius: 999px;
  background: #f1f5f9;
  color: #64748b;
  font-size: 12px;
}

.visual-option-row > .el-button {
  margin-top: 4px;
}

.matrix-editor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.matrix-subtitle {
  margin-bottom: 8px;
}

.visual-preview {
  align-self: start;
  overflow: hidden;
  background: #f8fafc;
}

.visual-preview-title {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid var(--bank-border);
  color: #1f2937;
  font-weight: 650;
}

.visual-preview-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
}

.preview-option {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 28px;
  color: #344054;
  font-size: 13px;
}

.preview-dot {
  width: 12px;
  height: 12px;
  border: 1px solid #cbd5e1;
  border-radius: 999px;
  background: #fff;
}

.preview-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #344054;
  font-size: 12px;
}

.preview-matrix {
  overflow-x: auto;
  border: 1px solid var(--bank-border);
  border-radius: 6px;
  background: #fff;
}

.preview-matrix-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(72px, 1fr));
}

.preview-matrix-row > span {
  padding: 7px 8px;
  border-right: 1px solid var(--bank-border);
  border-bottom: 1px solid var(--bank-border);
  color: #475467;
  font-size: 12px;
}

.preview-matrix-row.header > span {
  background: #f1f5f9;
  color: #334155;
  font-weight: 650;
}

.preview-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preview-schema-title {
  color: #344054;
  font-size: 13px;
  font-weight: 650;
}

.preview-schema {
  overflow: auto;
  max-height: 420px;
  margin: 0;
  padding: 12px;
  border: 1px solid var(--bank-border);
  border-radius: 8px;
  background: #f8fafc;
  color: #344054;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 768px) {
  .bank-tabs {
    width: 100%;
  }

  .visual-editor,
  .matrix-editor-grid,
  .visual-grid {
    grid-template-columns: 1fr;
  }

  .visual-option-row,
  .visual-option-row.compact {
    grid-template-columns: 24px minmax(0, 1fr) 34px;
  }

  .visual-option-row:not(.compact) :deep(.el-input:nth-of-type(2)) {
    grid-column: 2 / 3;
  }
}
</style>
