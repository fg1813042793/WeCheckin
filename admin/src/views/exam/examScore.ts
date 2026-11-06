interface ScoreItem { examScore?: number; examCorrect?: boolean }
interface ScoreQuestion {
  type?: string
  examScoreMode?: string
  examScore?: number
  props?: { fields?: ScoreItem[]; options?: ScoreItem[] }
}

// Keep the legacy display fallback; only explicit none overrides stored schema scores.
export function examDisplayScore(question: ScoreQuestion): number {
  return question.examScoreMode === 'none' ? 0 : Number(question.examScore) || 0
}

interface ExamSubmissionResult {
  record?: { score?: number; totalScore?: number }
  results?: { questionId: number; correct?: boolean }[]
  score?: number
  fullScore?: number
  correctCnt?: number
}

export function examResultSummary(data: ExamSubmissionResult) {
  if (!data?.record) return data
  return {
    score: data.record.score, fullScore: data.record.totalScore,
    correctCnt: data.results?.filter(item => item.correct).length, results: data.results,
  }
}

export function questionScore(q: ScoreQuestion): number {
  if (!q.examScoreMode || q.examScoreMode === 'none') return 0
  if (q.examScoreMode === 'manual') return Number(q.examScore) || 0
  if (q.examScoreMode === 'single') return Number(q.examScore) || 0
  if (q.examScoreMode === 'perOption') {
    const items = ['multiInput','hInput'].includes(q.type || '') ? (q.props?.fields||[]) : (q.props?.options||[])
    return items.reduce((s: number, o: ScoreItem) => s + (Number(o.examScore) || 0), 0)
  }
  if (q.examScoreMode === 'allCorrect') return Number(q.examScore) || 0
  if (q.examScoreMode === 'partialCorrect') {
    const items = ['multiInput','hInput'].includes(q.type || '') ? (q.props?.fields||[]) : (q.props?.options||[])
    return items.reduce((s: number, o: ScoreItem) => s + (o.examCorrect ? (Number(o.examScore) || 0) : 0), 0)
  }
  return 0
}
