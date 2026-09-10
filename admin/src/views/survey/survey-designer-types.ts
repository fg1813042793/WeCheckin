export interface SurveyDesignerQuestion {
  id: string
  type: string
  title: string
  description?: string
  required: boolean
  placeholder?: string
  props: Record<string, any>
  validate?: any[]
  logic?: any[]
  calcValue?: any
  readOnly?: boolean
  dataType?: string
  examScore?: number
  examCorrectAnswer?: string
  examAnalysis?: string
  mediaType?: string
  mediaUrl?: string
  mediaWidth?: string
  mediaAlign?: string
  [key: string]: any
}

export interface SurveyTextImportPayload {
  title: string
  questions: import('./survey-designer-text-transfer').ParsedSurveyTextQuestion[]
}

export interface SurveyFormulaPayload {
  type: string
  text: string
}

export interface SurveyBankUploadPayload {
  category: string
  tags: string
}

export interface SurveyLogicCondition {
  questionIdx?: number
  optionIdx?: number
  operator?: string
  compareValue?: string
}

export interface SurveyLogicRule {
  id: string
  conditionType: string
  conditions: SurveyLogicCondition[]
  action: string
  scope: 'frontend' | 'backend'
  targetQuestionIdx?: number
  targetOptionIdxs?: number[]
  branchFromIdx?: number
  branchToIdx?: number
  branchToEnd?: boolean
  formula?: string
  statField?: string
  statScope?: string
  notifyAdmin?: boolean
  notifyUserIds?: string
  notifyChannel?: string
  webhookType?: string
  webhookUrl?: string
  messageTemplate?: string
}

export type SurveyCompletionAction = 'default' | 'content' | 'redirect'
