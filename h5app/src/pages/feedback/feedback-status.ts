import type { UserFeedbackStatus } from '@/api/user-feedback'

export type UserFeedbackStatusTagType = 'success' | 'warning' | 'info'

export interface UserFeedbackStatusMeta {
  label: string
  type: UserFeedbackStatusTagType
}

const USER_FEEDBACK_STATUS_META: Record<UserFeedbackStatus, UserFeedbackStatusMeta> = {
  pending: { label: '待处理', type: 'warning' },
  processing: { label: '处理中', type: 'warning' },
  resolved: { label: '已解决', type: 'success' },
  closed: { label: '已关闭', type: 'info' },
}

export function feedbackStatusMeta(status: UserFeedbackStatus): UserFeedbackStatusMeta {
  return USER_FEEDBACK_STATUS_META[status]
}
