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
  id: number
  originalName: string
  contentType?: string
  sizeBytes: number
  url: string
  sortOrder: number
  createdAt: number
}

export interface UserFeedbackMessage {
  id: number
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
  id: number
  feedbackNo: string
  submitterId: number
  submitterName?: string
  summary: string
  imageCount: number
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

export interface AdminUserFeedbackListQuery {
  keyword?: string
  submitterId?: number
  handlerId?: number
  status?: UserFeedbackStatus
  submittedFrom?: number
  submittedTo?: number
  page?: number
  pageSize?: number
}

export interface UpdateUserFeedbackStatusInput {
  status: UserFeedbackStatus
  note: string
  notifyUser: boolean
  version: number
  requestId: string
}
