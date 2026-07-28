export type ID = string | number

export type QueryPrimitive = string | number | boolean | null | undefined
export type QueryParams = Record<string, QueryPrimitive>
export type FormPayload = Record<string, unknown>

export interface PageQuery extends QueryParams {
  page?: number
  pageSize?: number
  keyword?: string
  sort?: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page?: number
  size?: number
}

export interface AdminUser {
  id: number
  name: string
  mobile?: string
  avatar?: string
  pic?: string
  status: number
  forms?: string
  loginCnt?: number
  addTime?: number
  loginTime?: number
  deptIds?: number[]
  account?: string
  adminEnabled?: number
  adminType?: number
  roleId?: number
  roleName?: string
  allowPermissionKeys?: string[]
  denyPermissionKeys?: string[]
  adminDeptIds?: number[]
  positionId?: number
  positionName?: string
}

export interface SurveyItem {
  id: number
  title: string
  description?: string
  category?: string
  status: number
  mode?: string
  addTime?: number
  responseCount?: number
}

export interface ExamItem {
  id: number
  title: string
  description?: string
  category?: string
  status: number
  duration?: number
  addTime?: number
}

export interface QuestionBankItem {
  id: number
  title: string
  type: string
  schema: string
  category?: string
  tags?: string
  status?: number
  deptId?: number
  createBy?: number
  addTime?: number
}

export interface ResourceItem {
  id: number
  surveyId?: number
  examId?: number
  type: string
  url: string
  filename: string
  path: string
  domain?: string
  addTime?: number
}
