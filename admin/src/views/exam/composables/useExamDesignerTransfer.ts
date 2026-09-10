import { computed, ref, type Ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'
import { showRequestError } from '../../../utils/request'
import type { SurveyDesignerBankUploadDialogHandle, SurveyDesignerTextImportDialogHandle } from '../../survey/composables/useSurveyDesignerTransfer'
import type { SurveyBankUploadPayload } from '../../survey/survey-designer-types'
import type { ExamDesignerQuestion, ExamTextImportPayload } from '../exam-designer-types'

interface ExamDesignerTransferOptions {
  form: Record<string, any>
  questions: Ref<ExamDesignerQuestion[]>
  selected: Ref<ExamDesignerQuestion | null>
  generateId: () => string
  reloadBank: () => void
  typeName: (type: string) => string
}

export function useExamDesignerTransfer(options: ExamDesignerTransferOptions) {
  const textImportDialogRef = ref<SurveyDesignerTextImportDialogHandle | null>(null)
  const bankUploadDialogRef = ref<SurveyDesignerBankUploadDialogHandle | null>(null)
  const bankUploadQuestionId = ref('')
  const exportedJson = ref('')
  const fillTemplates: Record<string, string> = { ef: '默认模版', ef1: '新样式模版' }
  const publicUrl = computed(() => options.form.id
    ? `${window.location.origin}/${options.form.fillTemplate || 'ef'}/${options.form.id}`
    : '')

  function importTextQuestions(payload: ExamTextImportPayload) {
    if (payload.title) options.form.title = payload.title
    payload.questions.forEach(question => {
      const next: ExamDesignerQuestion = {
        id: options.generateId(), type: question.type, title: question.title,
        required: false, readOnly: false, dataType: '', placeholder: '',
        mediaType: '', mediaUrl: '', mediaWidth: '', mediaAlign: 'center',
        showDescription: true, props: {},
      }
      if (['radio', 'checkbox', 'select', 'picker', 'cascade'].includes(question.type) && question.options?.length) next.props.options = question.options
      if (question.type === 'judge') next.props.options = [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
      if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(question.type)) {
        if (question.rows?.length) next.props.rows = question.rows
        if (question.columns?.length) next.props.columns = question.columns
      }
      if (question.type === 'matrixAuto' && question.columns?.length) next.props.columns = question.columns
      if (['multiInput', 'hInput'].includes(question.type) && question.fields?.length) next.props.fields = question.fields
      options.questions.value.push(next)
      options.selected.value = next
    })
    ElMessage.success(`成功导入 ${payload.questions.length} 题`)
  }

  function openTextImport() { textImportDialogRef.value?.open() }

  function exportExam() {
    const lines = [options.form.title || '未命名试卷', '='.repeat(String(options.form.title || '未命名试卷').length), '']
    let index = 0
    options.questions.value.forEach(question => {
      if (['description', 'divider', 'pagination', 'questionSet'].includes(question.type)) {
        if (question.type === 'description') lines.push(`--- ${stripMarkup(question.title)} ---`)
        return
      }
      index++
      lines.push(`${index}. [${options.typeName(question.type)}] ${stripMarkup(question.title || '未命名')}${question.required ? '（必填）' : ''}`)
      question.props?.options?.forEach((item: any, optionIndex: number) => lines.push(`   ${String.fromCharCode(65 + optionIndex)}. ${item.label}`))
      if (question.props?.fields?.length) lines.push(`   字段: ${question.props.fields.map((item: any) => item.label).join(' / ')}`)
      if (question.props?.rows?.length) lines.push(`   行: ${question.props.rows.map((item: any) => item.title).join(' / ')}`)
      if (question.props?.columns?.length) lines.push(`   列: ${question.props.columns.map((item: any) => item.title || item.label).join(' / ')}`)
      lines.push('')
    })
    downloadBlob(new Blob([lines.join('\n')], { type: 'text/plain;charset=utf-8' }), `${options.form.title || '试卷'}.txt`)
    ElMessage.success('已导出')
  }

  function addFromBank(item: any) {
    const next: ExamDesignerQuestion = {
      id: options.generateId(), type: item.type, title: item.title || '未命名', required: false, props: {},
    }
    try { next.props = JSON.parse(item.schema) } catch {}
    options.questions.value.push(next)
    options.selected.value = next
  }

  async function onUploadBank(id: string) {
    if (!options.questions.value.some(question => question.id === id)) return
    bankUploadQuestionId.value = id
    bankUploadDialogRef.value?.open()
    try {
      const response: any = await adminApi.examQuestionBankCategories()
      bankUploadDialogRef.value?.setCategories(response.data || [])
    } catch {}
  }

  async function confirmUploadBank(payload: SurveyBankUploadPayload) {
    const question = options.questions.value.find(item => item.id === bankUploadQuestionId.value)
    if (!question) return
    bankUploadDialogRef.value?.setSaving(true)
    try {
      const { id: _id, ...schema } = question
      await adminApi.examQuestionBankInsert({
        title: stripMarkup(String(schema.title || '')).slice(0, 50),
        type: schema.type,
        schema: JSON.stringify(schema),
        category: payload.category || '',
        tags: payload.tags || '',
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

  function copyLink() { void navigator.clipboard.writeText(publicUrl.value); ElMessage.success('已复制') }
  function prepareJson() { exportedJson.value = JSON.stringify({ version: '2.0', questions: options.questions.value }, null, 2) }
  function downloadJson() { downloadBlob(new Blob([exportedJson.value], { type: 'application/json;charset=utf-8' }), `${options.form.title || 'exam'}.json`) }
  function loadJson() {
    const input = document.createElement('input')
    input.type = 'file'; input.accept = '.json'
    input.onchange = async event => {
      const file = (event.target as HTMLInputElement).files?.[0]
      if (!file) return
      try {
        const text = await file.text()
        const data = JSON.parse(text)
        if (data.questions) options.questions.value = data.questions
        exportedJson.value = text
        ElMessage.success('已导入')
      } catch { ElMessage.error('JSON 解析失败') }
    }
    input.click()
  }

  return { textImportDialogRef, bankUploadDialogRef, exportedJson, fillTemplates, publicUrl, importTextQuestions, openTextImport, exportExam, addFromBank, onUploadBank, confirmUploadBank, copyLink, prepareJson, downloadJson, loadJson }
}

function stripMarkup(value = '') { return value.replace(/<[^>]*>/g, '') }
function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url; link.download = filename; link.click()
  URL.revokeObjectURL(url)
}
