import type { ExamGradingContext, ExamGradingDraft, ExamGradingSubmission } from '@/types/examGrading'

export function scoreError(value: unknown, fullScore: number): string {
  if (value === undefined || value === null) return '请填写评分，0 分也需明确填写'
  if (typeof value !== 'number' || !Number.isInteger(value) || value < 0 || value > fullScore) {
    return `请输入 0 至 ${fullScore} 的整数分数`
  }
  return ''
}

export function retainGradingDraft(context: ExamGradingContext, draft: ExamGradingDraft): ExamGradingDraft {
  const retained: ExamGradingDraft = {}
  for (const question of context.questions) {
    const value = draft[question.questionId]
    if (question.needManual && !scoreError(value, question.fullScore)) retained[question.questionId] = value
  }
  return retained
}

export function gradingImageUrl(value: unknown): string {
  if (typeof value !== 'string') return ''
  const url = value.trim()
  // 禁止协议相对路径、反斜杠和浏览器可能忽略的控制字符。
  if (/[\u0000-\u0020\u007f\\]/.test(url)) return ''
  return /^\/(?!\/)/.test(url) || /^https?:\/\/[^/?#]+(?:[/?#]|$)/i.test(url) ? url : ''
}

export function gradingPreview(context: ExamGradingContext, draft: ExamGradingDraft) {
  let score = 0
  let remaining = 0
  for (const question of context.questions) {
    if (!question.needManual) score += question.gotScore
    else if (scoreError(draft[question.questionId], question.fullScore)) remaining++
    else score += draft[question.questionId] as number
  }
  return { score, remaining }
}

export function buildSubmission(context: ExamGradingContext, draft: ExamGradingDraft, comment: string): ExamGradingSubmission {
  if (context.record.status !== 1 || !context.canGrade) throw new Error('当前答卷不可批改')
  if (!context.version) throw new Error('缺少答卷版本，请重新加载')
  const manualQuestions = context.questions.filter(question => question.needManual)
  if (!manualQuestions.length) throw new Error('没有需要人工批改的题目')
  const scores = manualQuestions.map(question => {
    const score = draft[question.questionId]
    if (scoreError(score, question.fullScore)) throw new Error('请为全部人工题填写有效评分')
    return { questionId: question.questionId, score: score as number }
  })
  return { expectedVersion: context.version, scores, comment }
}

export function gradingErrorMessage(error: unknown, fallback: string): string {
  if (!error || typeof error !== 'object') return fallback
  if ('response' in error && error.response && typeof error.response === 'object' && 'data' in error.response) {
    return gradingErrorMessage(error.response.data, fallback)
  }
  if ('msg' in error && typeof error.msg === 'string' && error.msg) return error.msg
  if ('message' in error && typeof error.message === 'string' && error.message) return error.message
  return fallback
}
