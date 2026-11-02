/** 与考试记录 DTO 一致；批改上下文同时返回人工批改信息。 */
export interface ExamRecord {
  id: number
  examId: number
  paperId: number
  userId: string
  answers: string
  score: number
  totalScore: number
  pass: number
  status: number
  startTime: number
  submitTime: number
  timeSpent: number
  addIp: string
  isAutoSubmit: number
  device: string
  browser: string
  deviceType: string
  platformType: string
  deviceId: string
  session: string
  result: string
  gradedBy: number
  gradedAt: number
  gradingComment: string
}

export interface ExamGradingQuestion {
  questionId: number
  title: string
  type: string
  answer: string
  referenceAnswer: string
  fullScore: number
  gotScore: number
  needManual: boolean
  materials?: Array<{ label: string; content: string }>
  images?: Array<{ url: string; alt: string }>
}

export interface ExamGradingContext {
  record: ExamRecord
  examTitle: string
  passScore: number
  version: string
  canGrade: boolean
  notice: string
  questions: ExamGradingQuestion[]
}

export interface ExamGradingSubmission {
  expectedVersion: string
  scores: { questionId: number; score: number }[]
  comment: string
}

export type ExamGradingDraft = Record<number, number | null | undefined>
