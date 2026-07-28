export const LAYOUT_TYPES = ['description', 'divider', 'pagination']

export function parseJsonLike(raw, fallback) {
  if (!raw) return fallback
  if (typeof raw !== 'string') return raw
  try {
    return JSON.parse(raw)
  } catch (e) {
    return fallback
  }
}

export function parseSettings(raw) {
  return parseJsonLike(raw, {}) || {}
}

export function parseQuestions(raw) {
  const schema = parseJsonLike(raw, { questions: [] }) || { questions: [] }
  return Array.isArray(schema.questions) ? schema.questions : []
}

export function getQuestionInitialValue(q) {
  const type = q.type
  if (type === 'checkbox') return []
  if (type === 'switch') return false
  if (type === 'rating') return 0
  if (type === 'nps') return 0
  if (type === 'dateRange') return ['', '']
  if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(type)) return {}
  if (type === 'matrixAuto') return []
  if (['multiInput', 'hInput'].includes(type)) return (q.props?.fields || []).map(() => '')
  if (['user', 'dept'].includes(type)) return q.multiple ? [] : ''
  if (type === 'cascade') return []
  if (type === 'picker') return ''
  if (type === 'file') return []
  return ''
}

export function isQuestionAnswered(q, val) {
  const type = q.type
  if (val === undefined || val === null) return false
  if (type === 'checkbox') return Array.isArray(val) && val.length > 0
  if (type === 'file') return typeof val === 'string' ? !!val : (Array.isArray(val) && val.length > 0)
  if (['rating', 'nps'].includes(type)) return val > 0
  if (['multiInput', 'hInput'].includes(type)) return Array.isArray(val) && val.some(v => !!v)
  if (type === 'matrixRadio') return Object.keys(val).length > 0
  if (type === 'matrixCheckbox') return Object.values(val).some(v => Array.isArray(v) && v.length > 0)
  if (type === 'matrixFillBlank') return Object.values(val).some(v => !!v)
  if (type === 'matrixAuto') return Array.isArray(val) && val.some(row => row.some(v => !!v))
  if (type === 'dateRange') return Array.isArray(val) && val.some(v => !!v)
  if (type === 'switch') return val === true
  if (type === 'cascade') return Array.isArray(val) && val.length > 0
  if (['user', 'dept'].includes(type)) return q.multiple ? (Array.isArray(val) && val.length > 0) : !!val
  if (type === 'picker') return !!val
  return !!val
}

export function getAnswerProgress(questions = [], answers = {}, options = {}) {
  const hiddenIds = new Set(options.hiddenIds || [])
  const realQuestions = (questions || []).filter(q => {
    if (!q || LAYOUT_TYPES.includes(q.type)) return false
    if (q.defaultHidden) return false
    return !hiddenIds.has(q.id)
  })
  const answered = realQuestions.filter(q => isQuestionAnswered(q, answers[q.id])).length
  const total = realQuestions.length
  return {
    answered,
    total,
    percent: total ? Math.round(answered / total * 100) : 0
  }
}

export function formatRemainingTime(ms) {
  if (ms <= 0) return '已超时'
  const t = Math.ceil(ms / 1000)
  return `${Math.floor(t / 60)}:${(t % 60).toString().padStart(2, '0')}`
}
