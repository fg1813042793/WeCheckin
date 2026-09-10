import { computed, ref, type Ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'
import { showRequestError } from '../../../utils/request'
import { exportToSurveyKing, importFromSurveyKing, normalizeQuestions } from '../../../utils/surveyKingBridge'
import { serializeSurveyText } from '../survey-designer-text-transfer'
import type {
  SurveyBankUploadPayload,
  SurveyDesignerQuestion,
  SurveyTextImportPayload,
} from '../survey-designer-types'

export interface SurveyDesignerTextImportDialogHandle {
  open: () => void
}

export interface SurveyDesignerBankUploadDialogHandle {
  open: (categories?: string[]) => void
  setCategories: (categories: string[]) => void
  setSaving: (saving: boolean) => void
  close: () => void
}

interface SurveyDesignerTransferOptions {
  form: Record<string, any>
  questions: Ref<SurveyDesignerQuestion[]>
  selected: Ref<SurveyDesignerQuestion | null>
  selectedOptIdx: Ref<number>
  generateId: () => string
  replaceQuestions: (questions: SurveyDesignerQuestion[]) => void
  syncIdCounter: () => void
  reloadBank: () => void
}

export function useSurveyDesignerTransfer(options: SurveyDesignerTransferOptions) {
  const textImportDialogRef = ref<SurveyDesignerTextImportDialogHandle | null>(null)
  const bankUploadDialogRef = ref<SurveyDesignerBankUploadDialogHandle | null>(null)
  const bankUploadQuestionId = ref('')
  const exportedJson = ref('')
  const fillTemplates: Record<string, string> = { sf: '默认模版', sf1: '新样式模版' }
  const publicUrl = computed(() => options.form.id
    ? `${window.location.origin}/${options.form.fillTemplate || 'sf'}/${options.form.id}`
    : '')
  const qrUrl = computed(() => publicUrl.value
    ? `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(publicUrl.value)}`
    : '')
  const embedCode = computed(() => options.form.id
    ? `<iframe src="${publicUrl.value}" width="100%" height="600" frameborder="0"></iframe>`
    : '')

  function importTextQuestions(payload: SurveyTextImportPayload) {
    if (payload.title) options.form.title = payload.title
    payload.questions.forEach(question => {
      const newQuestion: SurveyDesignerQuestion = {
        id: options.generateId(),
        type: question.type,
        title: question.title,
        required: false,
        readOnly: false,
        dataType: '',
        placeholder: '',
        mediaType: '',
        mediaUrl: '',
        mediaWidth: '',
        mediaAlign: 'center',
        showDescription: true,
        props: {},
      }
      if (['radio', 'checkbox', 'select', 'picker', 'cascade'].includes(question.type) && question.options?.length) {
        newQuestion.props.options = question.options
      }
      if (question.type === 'judge') {
        newQuestion.props.options = [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
      }
      if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(question.type)) {
        if (question.rows?.length) newQuestion.props.rows = question.rows
        if (question.columns?.length) newQuestion.props.columns = question.columns
      }
      if (question.type === 'matrixAuto' && question.columns?.length) newQuestion.props.columns = question.columns
      if (['multiInput', 'hInput'].includes(question.type) && question.fields?.length) {
        newQuestion.props.fields = question.fields
      }
      options.questions.value.push(newQuestion)
      options.selected.value = newQuestion
    })
    ElMessage.success(`成功导入 ${payload.questions.length} 题`)
  }

  function openTextImport() {
    textImportDialogRef.value?.open()
  }

  function exportSurvey() {
    const text = serializeSurveyText({
      title: options.form.title,
      description: options.form.description,
      questions: options.questions.value,
    })
    downloadBlob(new Blob([text], { type: 'text/plain;charset=utf-8' }), `${options.form.title || '问卷'}.txt`)
    ElMessage.success('已导出')
  }

  function addFromBank(item: any) {
    const source = clonePlain(getBankQuestionSchema(item)) as SurveyDesignerQuestion
    const type = source.type || item.type || 'input'
    const newQuestion: SurveyDesignerQuestion = {
      ...source,
      id: options.generateId(),
      type,
      title: source.title || item.title || '未命名',
      required: source.required ?? false,
      readOnly: source.readOnly ?? false,
      dataType: source.dataType ?? '',
      placeholder: source.placeholder ?? '',
      mediaType: source.mediaType ?? '',
      mediaUrl: source.mediaUrl ?? '',
      mediaWidth: source.mediaWidth ?? '',
      mediaAlign: source.mediaAlign ?? 'center',
      showDescription: source.showDescription ?? !['description', 'questionSet', 'pagination', 'divider'].includes(type),
      props: source.props ? clonePlain(source.props) : {},
    }
    delete newQuestion.schema
    options.questions.value.push(newQuestion)
    options.selected.value = newQuestion
    options.selectedOptIdx.value = -1
  }

  async function onUploadBank(id: string) {
    if (!options.questions.value.some(question => question.id === id)) return
    bankUploadQuestionId.value = id
    bankUploadDialogRef.value?.open()
    try {
      const response: any = await adminApi.surveyQuestionBankCategories()
      bankUploadDialogRef.value?.setCategories(response.data || [])
    } catch {}
  }

  async function confirmUploadBank(dialogPayload: SurveyBankUploadPayload) {
    const question = options.questions.value.find(item => item.id === bankUploadQuestionId.value)
    if (!question) return
    bankUploadDialogRef.value?.setSaving(true)
    try {
      const { id: _id, ...schema } = question
      await adminApi.formkitSaveToBank({
        title: stripHtml(String(schema.title || '')).slice(0, 50),
        type: schema.type,
        schema: JSON.stringify(schema),
        category: dialogPayload.category,
        tags: dialogPayload.tags,
      })
      ElMessage.success('已上传到题库')
      bankUploadDialogRef.value?.close()
      options.reloadBank()
    } catch (error) {
      showRequestError(error, '上传失败')
    } finally {
      bankUploadDialogRef.value?.setSaving(false)
    }
  }

  function genQR() {
    if (!options.form.id) {
      ElMessage.warning('请先保存问卷')
      return
    }
    ElMessage.success('二维码已生成')
  }

  function copyLink() {
    void navigator.clipboard.writeText(publicUrl.value)
    ElMessage.success('已复制')
  }

  async function downloadQR() {
    try {
      const response = await fetch(qrUrl.value)
      downloadBlob(await response.blob(), `问卷二维码_${options.form.id}.png`)
    } catch {
      ElMessage.error('下载失败')
    }
  }

  function exportSurveyKingJson() {
    exportedJson.value = JSON.stringify({
      version: '2.0',
      questions: exportToSurveyKing(options.questions.value),
      setting: {},
    }, null, 2)
    ElMessage.success('已转换为 SurveyKing JSON')
  }

  function loadSurveyKingJson() {
    try {
      const text = prompt('请粘贴 SurveyKing JSON：')
      if (!text) return
      const parsed = JSON.parse(text)
      const converted = importFromSurveyKing(parsed.questions || parsed)
      if (!converted.length) {
        ElMessage.warning('未识别到有效题目')
        return
      }
      options.replaceQuestions(converted)
      options.syncIdCounter()
      options.selected.value = converted[0] || null
      ElMessage.success(`已导入 ${converted.length} 题`)
    } catch {
      ElMessage.error('JSON 解析失败')
    }
  }

  function downloadJson() {
    downloadBlob(new Blob([exportedJson.value], { type: 'application/json' }), 'survey.json')
  }

  return {
    textImportDialogRef,
    bankUploadDialogRef,
    exportedJson,
    fillTemplates,
    publicUrl,
    qrUrl,
    embedCode,
    importTextQuestions,
    openTextImport,
    exportSurvey,
    addFromBank,
    onUploadBank,
    confirmUploadBank,
    genQR,
    copyLink,
    downloadQR,
    exportSurveyKingJson,
    loadSurveyKingJson,
    downloadJson,
  }
}

function clonePlain<T>(value: T): T {
  return JSON.parse(JSON.stringify(value ?? {}))
}

function getBankQuestionSchema(item: any) {
  let parsed: any = {}
  try {
    parsed = typeof item.schema === 'string' ? JSON.parse(item.schema) : (item.schema || {})
  } catch {
    parsed = {}
  }
  const raw = Array.isArray(parsed?.questions) ? parsed.questions[0] : parsed
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return {}
  return normalizeQuestions([raw])[0] || raw
}

function stripHtml(value: string) {
  return value.replace(/<[^>]*>/g, '')
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}
