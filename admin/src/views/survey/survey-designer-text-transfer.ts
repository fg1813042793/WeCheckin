import { getSurveyQuestionTypeName } from './survey-designer-question-types'

export interface ParsedSurveyTextOption {
  label: string
  value: string
}

export interface ParsedSurveyTextQuestion {
  type: string
  title: string
  options: ParsedSurveyTextOption[]
  rows: Array<{ title: string }>
  columns: Array<{ title: string }>
  fields: Array<{ label: string; placeholder: string }>
}

export interface ParsedSurveyText {
  title: string
  questions: ParsedSurveyTextQuestion[]
}

interface SurveyTextQuestion {
  type: string
  title?: string
  required?: boolean
  props?: {
    options?: Array<{ label?: string; value?: string }>
    fields?: Array<{ label?: string }>
    rows?: Array<{ title?: string }>
    columns?: Array<{ title?: string; label?: string }>
  }
}

interface SurveyTextSource {
  title?: string
  description?: string
  questions: SurveyTextQuestion[]
}

const TEXT_TYPE_MAP: Record<string, string> = {
  '单选': 'radio', '单选择': 'radio', '单选题': 'radio',
  '多选': 'checkbox', '多选题': 'checkbox', '多选择': 'checkbox',
  '下拉': 'select', '下拉选择': 'select', '下拉框': 'select',
  '选择器': 'picker', picker: 'picker',
  '级联选择': 'cascade', '级联': 'cascade', cascade: 'cascade',
  '判断': 'judge', '判断题': 'judge',
  '上传文件': 'file', '文件上传': 'file', '图片上传': 'file', '文件': 'file',
  '单行文本': 'input', '文本': 'input',
  '多行': 'textarea', '多行文本': 'textarea', textarea: 'textarea',
  '数字': 'number',
  '多项填空': 'multiInput', multiInput: 'multiInput',
  '横向填空': 'hInput', hInput: 'hInput',
  '签名': 'signature', signature: 'signature',
  '扫码': 'scanCode', scanCode: 'scanCode',
  '评分': 'rating', rating: 'rating',
  'NPS评分': 'nps', nps: 'nps', NPS: 'nps',
  '手机': 'phone', '手机号': 'phone', phone: 'phone',
  '邮箱': 'email', email: 'email',
  '身份证': 'idCard', '身份证号': 'idCard', idCard: 'idCard',
  '密码': 'password', password: 'password',
  '姓名': 'name', name: 'name',
  '学号': 'studentId', studentId: 'studentId',
  '工号': 'employeeId', employeeId: 'employeeId',
  '班级': 'class', class: 'class',
  '开关': 'switch', switch: 'switch',
  '地理位置': 'location', '位置': 'location', '打卡': 'location', location: 'location',
  '日期': 'date', date: 'date',
  '时间': 'time', time: 'time',
  '日期范围': 'dateRange', dateRange: 'dateRange',
  '矩阵单选': 'matrixRadio', matrixRadio: 'matrixRadio',
  '矩阵多选': 'matrixCheckbox', matrixCheckbox: 'matrixCheckbox',
  '矩阵填空': 'matrixFillBlank', matrixFillBlank: 'matrixFillBlank',
  '表格自增': 'matrixAuto', '矩阵自增': 'matrixAuto', matrixAuto: 'matrixAuto',
  '成员': 'user', '人员': 'user', user: 'user',
  '部门': 'dept', dept: 'dept',
  '富文本': 'richText', richText: 'richText',
  '自动填充': 'autopop', '自动获取': 'autopop', autopop: 'autopop',
}

const OPTION_TYPES = ['radio', 'checkbox', 'select', 'picker', 'cascade', 'user', 'dept']
const LAYOUT_TYPES = ['description', 'divider', 'pagination', 'questionSet']

export function stripSurveyMarkup(value: string) {
  return value.replace(/<[^>]*>/g, '')
}

function createParsedQuestion(type = 'input', title = ''): ParsedSurveyTextQuestion {
  return { type, title, options: [], rows: [], columns: [], fields: [] }
}

export function parseSurveyText(text: string): ParsedSurveyText {
  const lines = text.split('\n')
  const questions: ParsedSurveyTextQuestion[] = []
  let current: ParsedSurveyTextQuestion | null = null
  let surveyTitle = ''
  let linesToSkip = 0
  const trimmed = lines.map(line => line.trim()).filter(Boolean)

  if (trimmed.length >= 2 && /^[=\-~]{3,}$/.test(trimmed[1])) {
    surveyTitle = trimmed[0]
    const nonEmptyIndexes = lines.reduce<number[]>((indexes, line, index) => {
      if (line.trim()) indexes.push(index)
      return indexes
    }, [])
    if (nonEmptyIndexes.length >= 2 && nonEmptyIndexes[0] + 1 === nonEmptyIndexes[1]
      && /^[=\-~]{3,}$/.test(lines[nonEmptyIndexes[1]].trim())) {
      linesToSkip = 2
    }
  }

  for (const rawLine of lines) {
    if (linesToSkip > 0) {
      linesToSkip--
      continue
    }
    const line = rawLine.trim()
    if (!line) {
      if (current?.title) {
        questions.push(current)
        current = null
      }
      continue
    }

    const optionMatch = line.match(/^([A-Za-z])[.、）)]\s*(.*)/)
    if (optionMatch) {
      current ||= createParsedQuestion('radio')
      const optionText = stripSurveyMarkup(optionMatch[2])
      if (optionText) {
        let label = optionText
        let value = optionMatch[1]
        const explicitValue = optionText.match(/^\[(.+?)\]\s*(.*)/)
        if (explicitValue) {
          value = explicitValue[1]
          label = explicitValue[2]
        }
        current.options.push({ label, value })
        if (current.type === 'input') current.type = 'radio'
      }
      continue
    }

    const rowMatch = line.match(/^行[:：]\s*(.+)/)
    const columnMatch = line.match(/^列[:：]\s*(.+)/)
    const fieldMatch = line.match(/^字段[:：]\s*(.+)/)
    if (rowMatch || columnMatch || fieldMatch) {
      current ||= createParsedQuestion()
      if (rowMatch) current.rows = rowMatch[1].split('/').map(title => ({ title: title.trim() }))
      if (columnMatch) current.columns = columnMatch[1].split('/').map(title => ({ title: title.trim() }))
      if (fieldMatch) current.fields = fieldMatch[1].split('/').map(label => ({ label: label.trim(), placeholder: '' }))
      continue
    }

    if (current?.title) questions.push(current)
    const typeMatch = line.match(/^\d*[.、）)]?\s*\[(.+?)\]\s*(.*)/)
    if (typeMatch) {
      const type = TEXT_TYPE_MAP[typeMatch[1].trim()] || 'input'
      const fallbackTitle = line.replace(/^\d*[.、）)]?\s*/, '').replace(/\[.*?\]/, '').trim()
      const title = stripSurveyMarkup(typeMatch[2]) || stripSurveyMarkup(fallbackTitle)
      current = createParsedQuestion(type, title)
      if (type === 'judge') current.options = [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
    } else {
      current = createParsedQuestion('input', stripSurveyMarkup(line.replace(/^\d+[.、）)]?\s*/, '')))
    }
  }

  if (current?.title) questions.push(current)
  questions.forEach(question => {
    if (question.options.length && question.type === 'input') question.type = 'radio'
    if (!question.options.length && OPTION_TYPES.includes(question.type)) question.type = 'input'
  })
  return { title: surveyTitle, questions }
}

export function serializeSurveyText(source: SurveyTextSource) {
  const lines: string[] = []
  const title = source.title || '未命名问卷'
  lines.push(title, '='.repeat(title.length))
  if (source.description) lines.push('', source.description)
  lines.push('')

  let questionIndex = 0
  source.questions.forEach(question => {
    if (LAYOUT_TYPES.includes(question.type)) {
      if (question.type === 'description') lines.push(`--- ${stripSurveyMarkup(question.title || '')} ---`)
      return
    }
    questionIndex++
    const questionTitle = stripSurveyMarkup(question.title || '未命名')
    const required = question.required ? '（必填）' : ''
    lines.push(`${questionIndex}. [${getSurveyQuestionTypeName(question.type)}] ${questionTitle}${required}`)
    question.props?.options?.forEach((option, optionIndex) => {
      const prefix = String.fromCharCode(65 + optionIndex)
      const label = stripSurveyMarkup(option.label || '')
      lines.push(option.value && option.value !== prefix
        ? `   ${prefix}. [${option.value}] ${label}`
        : `   ${prefix}. ${label}`)
    })
    if (question.props?.fields?.length) {
      lines.push(`   字段: ${question.props.fields.map(field => field.label || '').join(' / ')}`)
    }
    if (question.props?.rows?.length) {
      lines.push(`   行: ${question.props.rows.map(row => row.title || '').join(' / ')}`)
    }
    if (question.props?.columns?.length) {
      lines.push(`   列: ${question.props.columns.map(column => column.title || column.label || '').join(' / ')}`)
    }
    lines.push('')
  })
  return lines.join('\n')
}
