import type { SurveyDesignerQuestion, SurveyTextImportPayload } from '../survey/survey-designer-types'

export type ExamDesignerQuestion = SurveyDesignerQuestion
export type ExamTextImportPayload = SurveyTextImportPayload
export type ExamCompletionAction = 'default' | 'content' | 'redirect'

export interface ExamLogicCondition {
  questionIdx?: number
  optionIdx?: number
  operator?: string
  compareValue?: string
}

export interface ExamLogicRule {
  id: string
  conditionType: string
  conditions: ExamLogicCondition[]
  action: string
  scope: 'frontend' | 'backend'
  targetQuestionIdx?: number
  targetOptionIdxs?: number[]
  branchFromIdx?: number
  branchToIdx?: number
  branchToEnd?: boolean
  formula?: string
}
