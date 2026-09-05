import { del, get, patch } from '@/api/dingtalk-h5/base'

const NOTIFICATION_API = '/api/v2/dingtalk/h5/notifications'

export type NotificationTone = 'primary' | 'success' | 'warning' | 'danger' | 'info'
export type DingTalkMessageType = 'auto' | 'text' | 'image' | 'voice' | 'file' | 'link' | 'oa' | 'markdown' | 'action_card'

export interface DingTalkNotificationTemplate {
  messageType: DingTalkMessageType
  title: string
  content: string
  url: string
  picUrl: string
  sourceName: string
  mediaId: string
  duration: number
  buttonTitle: string
  headColor: string
}

export interface NotificationStyle {
  type: string
  label: string
  icon: string
  tone: NotificationTone
  dingTalk?: DingTalkNotificationTemplate
}

export interface InAppNotification {
  id: number
  title: string
  content: string
  type: string
  sourceType: string
  sourceId: string
  isRead: number
  addTime: number
  style?: NotificationStyle
}

export interface InAppNotificationList {
  list: InAppNotification[]
  total: number
  page: number
  pageSize: number
}

export interface InAppNotificationUnreadCount {
  count: number
}

export function listNotifications(page = 1, pageSize = 20, unreadOnly = false) {
  return get<InAppNotificationList>(NOTIFICATION_API, {
    page,
    pageSize,
    unreadOnly: unreadOnly ? 1 : undefined,
  })
}

export function getNotificationUnreadCount() {
  return get<InAppNotificationUnreadCount>(`${NOTIFICATION_API}/unread-count`)
}

export function markNotificationRead(id: number) {
  return patch<void>(`${NOTIFICATION_API}/${id}/read`)
}

export function markAllNotificationsRead() {
  return patch<void>(`${NOTIFICATION_API}/read-all`)
}

export function deleteNotification(id: number) {
  return del<void>(`${NOTIFICATION_API}/${id}`)
}
