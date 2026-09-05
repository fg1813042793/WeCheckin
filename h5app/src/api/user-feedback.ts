import type { MultipartUploadFile } from '@/api/dingtalk-h5/base'
import { get, uploadMultipartFiles } from '@/api/dingtalk-h5/base'

export type UserFeedbackStatus = 'pending' | 'processing' | 'resolved' | 'closed'
export type UserFeedbackMessageType = 'initial' | 'supplement' | 'status'
export type UserFeedbackAuthorType = 'user' | 'admin'

export interface UserFeedbackOverview {
  pending: number
  processing: number
  resolved: number
  closed: number
}

export interface UserFeedbackAttachment {
  id: string
  originalName: string
  contentType?: string
  sizeBytes: number
  url: string
  sortOrder: number
  createdAt: number
}

export interface UserFeedbackMessage {
  id: string
  messageType: UserFeedbackMessageType
  authorType: UserFeedbackAuthorType
  authorId: number
  authorName?: string
  content: string
  fromStatus?: UserFeedbackStatus
  toStatus?: UserFeedbackStatus
  attachments: UserFeedbackAttachment[]
  createdAt: number
}

export interface UserFeedbackSummary {
  id: string
  feedbackNo: string
  submitterId: number
  submitterName?: string
  summary: string
  imageCount: number
  firstImageUrl?: string
  status: UserFeedbackStatus
  handlerId?: number
  handlerName?: string
  version: number
  lastActivityAt: number
  resolvedAt?: number
  closedAt?: number
  createdAt: number
  updatedAt: number
}

export interface UserFeedbackDetail extends UserFeedbackSummary {
  messages: UserFeedbackMessage[]
  allowsSupplement: boolean
}

export interface UserFeedbackList {
  list: UserFeedbackSummary[]
  total: number
  page: number
  pageSize: number
}

export interface UserFeedbackListQuery {
  keyword?: string
  status?: UserFeedbackStatus
  page?: number
  pageSize?: number
}

export interface CreateUserFeedbackInput {
  content: string
  requestId: string
  images?: readonly MultipartUploadFile[]
}

export interface SupplementUserFeedbackInput {
  content: string
  requestId: string
  version: number
  images?: readonly MultipartUploadFile[]
}

export const USER_FEEDBACK_API_PERMISSIONS = {
  overview: 'dingtalk_h5:api:feedback:list',
  list: 'dingtalk_h5:api:feedback:list',
  create: 'dingtalk_h5:api:feedback:create',
  detail: 'dingtalk_h5:api:feedback:detail',
  supplement: 'dingtalk_h5:api:feedback:supplement',
} as const

const USER_FEEDBACK_API = '/api/v2/dingtalk/h5/user-feedbacks'

function feedbackResourcePath(id: string) {
  return `${USER_FEEDBACK_API}/${encodeURIComponent(id.trim())}`
}

export function getUserFeedbackOverview() {
  return get<UserFeedbackOverview>(`${USER_FEEDBACK_API}/overview`)
}

export function listUserFeedbacks(query: UserFeedbackListQuery = {}) {
  return get<UserFeedbackList>(USER_FEEDBACK_API, query)
}

export function createUserFeedback(input: CreateUserFeedbackInput) {
  return uploadMultipartFiles<UserFeedbackDetail>(USER_FEEDBACK_API, input.images || [], {
    formData: {
      content: input.content,
      requestId: input.requestId,
    },
  })
}

export function getUserFeedbackDetail(id: string) {
  return get<UserFeedbackDetail>(feedbackResourcePath(id))
}

export function supplementUserFeedback(id: string, input: SupplementUserFeedbackInput) {
  return uploadMultipartFiles<UserFeedbackDetail>(
    `${feedbackResourcePath(id)}/supplements`,
    input.images || [],
    {
      formData: {
        content: input.content,
        requestId: input.requestId,
        version: input.version,
      },
    },
  )
}
