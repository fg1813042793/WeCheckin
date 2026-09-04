import type {
  ApiEnvelope,
  AuthSessionPayload,
  CreateReviewPayload,
  CreateReviewRequest,
  PerformanceReview,
  PerformanceReviewListPayload,
  PerformanceTemplate,
  PerformanceUser,
  ReviewActionRequest,
  WorkbenchStat,
} from '@/types/dingtalk-h5'
import { authToken, buildApiUrl, del, DINGTALK_H5_API, get, post, put } from './base'

export function bootstrap() {
  return get<AuthSessionPayload>(`${DINGTALK_H5_API}/bootstrap`)
}

export function workbench() {
  return get<{ stats?: WorkbenchStat[] }>(`${DINGTALK_H5_API}/workbench`)
}

export function listReviews(params: Record<string, unknown> = {}) {
  return get<PerformanceReviewListPayload | PerformanceReview[]>(`${DINGTALK_H5_API}/reviews`, params)
}

export function reviewDetail(id: string) {
  return get<PerformanceReview>(`${DINGTALK_H5_API}/reviews/${encodeURIComponent(id)}`)
}

export function reviewAction(id: string, action: string, data: ReviewActionRequest = { action }) {
  return post<PerformanceReview>(`${DINGTALK_H5_API}/reviews/${encodeURIComponent(id)}/${action}`, data)
}

export function deleteReview(id: string) {
  return del<ApiEnvelope>(`${DINGTALK_H5_API}/reviews/${encodeURIComponent(id)}`)
}

export function exportReviewsUrl(params: Record<string, string | number | boolean | null | undefined> = {}) {
  return buildApiUrl(`${DINGTALK_H5_API}/reviews/export`, {
    ...params,
    token: authToken(),
  })
}

export function createReview(data: CreateReviewRequest) {
  return post<PerformanceReview | CreateReviewPayload>(`${DINGTALK_H5_API}/reviews`, data)
}

export function listUsers() {
  return get<PerformanceUser[]>(`${DINGTALK_H5_API}/users`)
}

export function createUser(data: PerformanceUser) {
  return post<{ user?: PerformanceUser, users?: PerformanceUser[] }>(`${DINGTALK_H5_API}/users`, data)
}

export function updateUser(id: string, data: Partial<PerformanceUser>) {
  return put<{ user?: PerformanceUser, users?: PerformanceUser[] }>(`${DINGTALK_H5_API}/users/${encodeURIComponent(id)}`, data)
}

export function deleteUser(id: string) {
  return del<{ users?: PerformanceUser[] }>(`${DINGTALK_H5_API}/users/${encodeURIComponent(id)}`)
}

export function getTemplate() {
  return get<PerformanceTemplate>(`${DINGTALK_H5_API}/template`)
}

export function saveTemplate(data: PerformanceTemplate) {
  return put<PerformanceTemplate>(`${DINGTALK_H5_API}/template`, data)
}
