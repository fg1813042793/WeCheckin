import type { UserFeedbackStatus } from '@/types/userFeedback'
import { hasPerm } from '@/utils/permission'

export type UserFeedbackStatusTagType = 'success' | 'warning' | 'info'

export interface UserFeedbackStatusMeta {
  label: string
  type: UserFeedbackStatusTagType
}

export const USER_FEEDBACK_LIST_PERMISSION = 'admin:menu:user-feedback:list'
export const USER_FEEDBACK_HANDLE_PERMISSION = 'admin:menu:user-feedback:handle'

const USER_FEEDBACK_STATUS_META: Record<UserFeedbackStatus, UserFeedbackStatusMeta> = {
  pending: { label: '待处理', type: 'warning' },
  processing: { label: '处理中', type: 'warning' },
  resolved: { label: '已解决', type: 'success' },
  closed: { label: '已关闭', type: 'info' },
}

const USER_FEEDBACK_STATUS_TRANSITIONS: Record<UserFeedbackStatus, readonly UserFeedbackStatus[]> = {
  pending: ['processing', 'closed'],
  processing: ['resolved'],
  resolved: ['closed', 'processing'],
  closed: ['processing'],
}

export function userFeedbackStatusMeta(status: UserFeedbackStatus): UserFeedbackStatusMeta {
  return USER_FEEDBACK_STATUS_META[status]
}

export function nextUserFeedbackStatuses(status: UserFeedbackStatus): readonly UserFeedbackStatus[] {
  return USER_FEEDBACK_STATUS_TRANSITIONS[status]
}

export function canHandleUserFeedback(): boolean {
  return hasPerm(USER_FEEDBACK_HANDLE_PERMISSION)
}
