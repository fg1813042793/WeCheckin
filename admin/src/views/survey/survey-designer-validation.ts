export type SurveyDesignerIssueSeverity = 'error' | 'warning'
export type SurveyDesignerIssueSection = 'edit' | 'setting'

export interface SurveyDesignerIssue {
  code: string
  message: string
  severity: SurveyDesignerIssueSeverity
  section: SurveyDesignerIssueSection
  questionId?: string
  questionIndex?: number
}

interface SurveyDesignerValidationInput {
  form: Record<string, any>
  questions: Array<Record<string, any>>
}

const pureLayoutTypes = new Set(['divider', 'description', 'pagination', 'questionSet'])
const optionTypes = new Set(['select', 'radio', 'checkbox', 'picker', 'cascade', 'judge'])
const matrixTypes = new Set(['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'])
const multiInputTypes = new Set(['multiInput', 'hInput'])

export function isValidRedirectUrl(value: unknown) {
  const raw = String(value || '').trim()
  if (!raw || raw.startsWith('/')) return true
  try {
    const url = new URL(raw)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

function optionLabel(option: any) {
  return String(option?.label ?? option?.title ?? '').replace(/<[^>]*>/g, '').trim()
}

export function collectSurveyDesignerIssues({ form, questions }: SurveyDesignerValidationInput): SurveyDesignerIssue[] {
  const issues: SurveyDesignerIssue[] = []
  const pushSurveyIssue = (code: string, message: string, section: SurveyDesignerIssueSection = 'setting') => {
    issues.push({ code, message, severity: 'error', section })
  }

  if (!String(form.title || '').trim()) pushSurveyIssue('survey_title_required', '请填写问卷标题', 'edit')
  if (!questions.some(question => !pureLayoutTypes.has(question.type))) {
    issues.push({ code: 'survey_question_required', message: '至少添加一道可作答题目', severity: 'warning', section: 'edit' })
  }
  if (Number(form.visibility) === 2 && !String(form.deptIds || '').split(',').map(item => item.trim()).some(Boolean)) {
    pushSurveyIssue('survey_department_required', '部门限定需要至少选择一个部门')
  }
  if (form.startDate && form.endDate && Number(form.endDate) < Number(form.startDate)) {
    pushSurveyIssue('survey_time_range_invalid', '结束时间不能早于开始时间')
  }
  if (!isValidRedirectUrl(form.redirectUrl)) pushSurveyIssue('survey_redirect_url_invalid', '请输入有效的完成后跳转链接')

  const idCounts = new Map<string, number>()
  const duplicateIdsReported = new Set<string>()
  questions.forEach((question) => {
    const id = String(question.id || '').trim()
    if (id) idCounts.set(id, (idCounts.get(id) || 0) + 1)
  })

  questions.forEach((question, index) => {
    const questionId = String(question.id || '').trim()
    const add = (code: string, message: string, severity: SurveyDesignerIssueSeverity = 'error') => {
      issues.push({ code, message, severity, section: 'edit', questionId: questionId || undefined, questionIndex: index })
    }
    if (!questionId) add('question_id_required', `第 ${index + 1} 项缺少题目 ID`)
    else if ((idCounts.get(questionId) || 0) > 1 && !duplicateIdsReported.has(questionId)) {
      duplicateIdsReported.add(questionId)
      add('question_id_duplicate', `题目 ID “${questionId}” 重复`)
    }

    if (!pureLayoutTypes.has(question.type) && !String(question.title || '').replace(/<[^>]*>/g, '').trim()) {
      add('question_title_required', `第 ${index + 1} 题缺少标题`)
    }

    if (optionTypes.has(question.type)) {
      const options = Array.isArray(question.props?.options) ? question.props.options : []
      if (!options.length) add('question_options_required', `第 ${index + 1} 题至少需要一个选项`)
      else if (options.some((option: any) => !optionLabel(option))) add('question_option_label_required', `第 ${index + 1} 题存在空选项`)
    }

    if (multiInputTypes.has(question.type)) {
      const fields = Array.isArray(question.props?.fields) ? question.props.fields : []
      if (!fields.length) add('question_fields_required', `第 ${index + 1} 题至少需要一个填写项`)
      else if (fields.some((field: any) => !String(field?.label || '').trim())) add('question_field_label_required', `第 ${index + 1} 题存在未命名填写项`)
    }

    if (matrixTypes.has(question.type)) {
      if (!Array.isArray(question.props?.rows) || !question.props.rows.length) add('question_matrix_rows_required', `第 ${index + 1} 题至少需要一行`)
      if (!Array.isArray(question.props?.columns) || !question.props.columns.length) add('question_matrix_columns_required', `第 ${index + 1} 题至少需要一列`)
    }
    if (question.type === 'matrixAuto' && (!Array.isArray(question.props?.columns) || !question.props.columns.length)) {
      add('question_matrix_columns_required', `第 ${index + 1} 题至少需要一列`)
    }

    const min = question.props?.min
    const max = question.props?.max
    if (question.type === 'number' && min !== undefined && max !== undefined && Number(min) > Number(max)) {
      add('question_number_range_invalid', `第 ${index + 1} 题的最小值不能大于最大值`)
    }
  })

  return issues
}
