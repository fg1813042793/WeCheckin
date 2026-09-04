import type { ApiResponse } from '../utils/request'

export interface PublicLoginData {
  token: string
  userInfo?: {
    id?: number
    deptId?: number
    name?: string
  }
}

export interface PublicValidationData {
  valid: boolean
  errors: Array<{ questionId?: string; message: string }>
}

export interface PublicSurveyDetail {
  id: number
  title: string
  schema?: { questions?: Array<Record<string, unknown>> }
  settings?: Record<string, unknown>
  session?: string
  startAt?: number
  [key: string]: unknown
}

export interface PublicExamDetail {
  id?: number
  title?: string
  exam?: Record<string, unknown>
  paper?: Record<string, unknown>
  record?: Record<string, unknown>
  questions?: Array<Record<string, unknown>>
  answers?: Record<string, unknown>
  results?: Array<Record<string, unknown>>
  schema?: { questions?: Array<Record<string, unknown>> }
  settings?: Record<string, unknown>
  session?: string
  startAt?: number
  [key: string]: unknown
}

export type PublicSubmitData = Record<string, unknown>
export type PublicRequestData = Record<string, unknown>

const API_BASE = import.meta.env.VITE_API_BASE || ''

function isApiResponse(value: unknown): value is ApiResponse<unknown> {
  return Boolean(
    value
    && typeof value === 'object'
    && typeof (value as { code?: unknown }).code === 'number'
    && typeof (value as { msg?: unknown }).msg === 'string',
  )
}

async function publicRequest<T>(path: string, init?: RequestInit): Promise<ApiResponse<T>> {
  const token = localStorage.getItem('user_token')
  const headers = new Headers(init?.headers)
  if (token) headers.set('Authorization', token)
  if (init?.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')

  const response = await fetch(`${API_BASE}${path}`, { ...init, headers })
  const payload: unknown = await response.json()
  if (!isApiResponse(payload)) {
    throw new Error('接口响应格式错误')
  }
  return payload as ApiResponse<T>
}

function post<T>(path: string, data: PublicRequestData) {
  return publicRequest<T>(path, { method: 'POST', body: JSON.stringify(data) })
}

function normalizeValidation(response: ApiResponse<PublicValidationData & { ok?: boolean }>): ApiResponse<PublicValidationData> {
  const data = response.data || { valid: false, errors: [] }
  return {
    ...response,
    data: {
      valid: typeof data.valid === 'boolean' ? data.valid : data.ok === true,
      errors: Array.isArray(data.errors) ? data.errors : [],
    },
  }
}

export const publicFormApi = {
  login(name: string, password: string) {
    return post<PublicLoginData>('/api/v2/auth/password-login', { name, pwd: password })
  },
  surveyDetail(id: string | number, session: string) {
    const query = new URLSearchParams({ session })
    return publicRequest<PublicSurveyDetail>(`/api/v2/surveys/${encodeURIComponent(String(id))}?${query}`)
  },
  surveyValidate(data: PublicRequestData) {
    return post<PublicValidationData>('/api/v2/survey/validate', data)
  },
  surveySubmit(id: string | number, data: PublicRequestData) {
    return post<PublicSubmitData>(`/api/v2/surveys/${encodeURIComponent(String(id))}/responses`, data)
  },
  examDetail(id: string | number, session: string) {
    const query = new URLSearchParams({ session })
    return publicRequest<PublicExamDetail>(`/api/v2/exams/${encodeURIComponent(String(id))}?${query}`)
  },
  examResult(session: string) {
    const query = new URLSearchParams({ session })
    return publicRequest<PublicExamDetail>(`/api/v2/exam-results?${query}`)
  },
  async examValidate(id: string | number, data: PublicRequestData) {
    const response = await post<PublicValidationData & { ok?: boolean }>(`/api/v2/exams/${encodeURIComponent(String(id))}/validation`, data)
    return normalizeValidation(response)
  },
  examSubmit(id: string | number, data: PublicRequestData) {
    return post<PublicSubmitData>(`/api/v2/exams/${encodeURIComponent(String(id))}/submissions`, data)
  },
}
