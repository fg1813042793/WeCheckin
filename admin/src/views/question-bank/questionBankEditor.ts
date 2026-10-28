import type { BankScope } from './utils/importExport'

export const questionBankTypeMap: Record<string, string> = {
  radio: '单选题', checkbox: '多选题', select: '下拉题', picker: '选择器', cascade: '级联选择', judge: '判断题',
  file: '文件/图片', input: '单行文本', textarea: '多行文本', number: '数字', multiInput: '多项填空', hInput: '横向填空',
  signature: '签名', scanCode: '扫码', rating: '评分', nps: 'NPS评分', matrixRadio: '矩阵单选', matrixCheckbox: '矩阵多选',
  matrixFillBlank: '矩阵填空', matrixAuto: '表格自增', divider: '分割线', description: '文字描述', questionSet: '问题组',
  pagination: '分页', user: '成员', dept: '部门', richText: '富文本', autopop: '自动填充', name: '姓名', studentId: '学号',
  employeeId: '工号', class: '班级', phone: '手机', email: '邮箱', idCard: '身份证', password: '密码', date: '日期',
  time: '时间', dateRange: '日期范围', switch: '开关', location: '地理位置',
}

export const questionBankTypeOptions = Object.entries(questionBankTypeMap).map(([value, label]) => ({ value, label }))
export const richTitleEditorOptions = {
  theme: 'snow', placeholder: '请输入题目标题',
  modules: { toolbar: [['bold', 'italic', 'underline', 'strike'], [{ color: [] }, { background: [] }], [{ list: 'ordered' }, { list: 'bullet' }], ['clean']] },
}
export const richOptionEditorOptions = {
  theme: 'snow', placeholder: '选项文本',
  modules: { toolbar: [['bold', 'italic', 'underline'], [{ color: [] }, { background: [] }], ['clean']] },
}
export const richFullEditorOptions = {
  theme: 'snow', placeholder: '请输入富文本内容',
  modules: { toolbar: [[{ header: [1, 2, 3, false] }], ['bold', 'italic', 'underline', 'strike'], [{ color: [] }, { background: [] }], [{ list: 'ordered' }, { list: 'bullet' }], [{ indent: '-1' }, { indent: '+1' }], [{ align: [] }], ['blockquote', 'code-block'], ['link'], ['clean']] },
}

const optionTypes = new Set(['radio', 'checkbox', 'select', 'picker', 'cascade', 'judge'])
const fieldTypes = new Set(['multiInput', 'hInput'])
const matrixTypes = new Set(['matrixRadio', 'matrixCheckbox', 'matrixFillBlank', 'matrixAuto'])
const layoutTypes = new Set(['description', 'questionSet', 'pagination', 'divider'])

export function typeName(type: string) {
  return questionBankTypeMap[type] || type || '-'
}

export function stripQuestionHtml(value: string) {
  return String(value || '').replace(/<[^>]*>/g, '').replace(/&nbsp;/g, ' ').trim()
}

export function usesOptions(type: string) { return optionTypes.has(type) }
export function usesFields(type: string) { return fieldTypes.has(type) }
export function usesMatrix(type: string) { return matrixTypes.has(type) }

function clonePlain<T>(value: T): T {
  return JSON.parse(JSON.stringify(value ?? {}))
}

function optionValue(index: number) {
  return String.fromCharCode(65 + index)
}

function normalizeOption(option: any, index: number) {
  return { ...clonePlain(option || {}), label: String(option?.label ?? option?.title ?? option?.text ?? `选项${optionValue(index)}`), value: String(option?.value ?? optionValue(index)) }
}

function normalizeField(field: any, index: number) {
  return { ...clonePlain(field || {}), label: String(field?.label ?? field?.title ?? `填空${index + 1}`), placeholder: String(field?.placeholder ?? '') }
}

function normalizeMatrixItem(item: any, index: number, prefix: string) {
  return { ...clonePlain(item || {}), id: String(item?.id || `${prefix}${index + 1}`), title: String(item?.title ?? item?.label ?? `${prefix === 'r' ? '行' : '列'}${index + 1}`) }
}

export function ensureQuestionBankQuestionShape(question: any, fallback: { type: string; title: string; scope: BankScope }) {
  const q = clonePlain(question || {})
  q.type = String(q.type || fallback.type || 'radio')
  q.title = String(q.title || fallback.title || typeName(q.type) || '未命名题目')
  q.required = Boolean(q.required)
  q.readOnly = Boolean(q.readOnly)
  q.placeholder = String(q.placeholder || '')
  q.description = String(q.description || '')
  q.showDescription = q.showDescription ?? !layoutTypes.has(q.type)
  q.props = q.props && typeof q.props === 'object' && !Array.isArray(q.props) ? q.props : {}
  if (usesOptions(q.type)) {
    const defaults = q.type === 'judge' ? [{ label: '对', value: 'true' }, { label: '错', value: 'false' }] : [{ label: '选项A', value: 'A' }, { label: '选项B', value: 'B' }]
    q.props.options = Array.isArray(q.props.options) && q.props.options.length ? q.props.options.map(normalizeOption) : defaults
  }
  if (usesFields(q.type)) {
    const source = Array.isArray(q.props.fields) && q.props.fields.length ? q.props.fields : q.props.options
    q.props.fields = Array.isArray(source) && source.length ? source.map(normalizeField) : [{ label: '填空1', placeholder: '' }, { label: '填空2', placeholder: '' }]
  }
  if (usesMatrix(q.type)) {
    q.props.rows = Array.isArray(q.props.rows) && q.props.rows.length ? q.props.rows.map((item: any, index: number) => normalizeMatrixItem(item, index, 'r')) : [{ id: 'r1', title: '行1' }, { id: 'r2', title: '行2' }]
    q.props.columns = Array.isArray(q.props.columns) && q.props.columns.length ? q.props.columns.map((item: any, index: number) => normalizeMatrixItem(item, index, 'c')) : [{ id: 'c1', title: '列A' }, { id: 'c2', title: '列B' }]
    if (q.type === 'matrixAuto') q.props.rows = Array.isArray(q.props.rows) ? q.props.rows : []
  }
  if (q.type === 'file') {
    q.fileTypes = Array.isArray(q.fileTypes) ? q.fileTypes : ['image']
    q.maxFileSize = Number(q.maxFileSize || 10)
    q.maxFileCount = Number(q.maxFileCount || 1)
  }
  if (q.type === 'rating') q.props.maxRating = Number(q.props.maxRating || 5)
  if (q.type === 'nps') q.props.maxRating = Number(q.props.maxRating || 10)
  if (fallback.scope === 'exam') {
    q.examScore = Number(q.examScore || 0)
    q.examCorrectAnswer = String(q.examCorrectAnswer || '')
    q.examAnalysis = String(q.examAnalysis || '')
  }
  return q
}

export function parseQuestionBankSchema(schema: any, fallback: any = {}) {
  let parsed = schema
  if (typeof schema === 'string' && schema.trim()) parsed = JSON.parse(schema)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return fallback
  if (Array.isArray(parsed.questions)) return parsed.questions[0] || fallback
  return parsed
}

export function nextOptionValue(index: number) {
  return optionValue(index)
}
