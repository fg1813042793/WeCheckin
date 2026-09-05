export interface FeedbackNotificationOpenInput {
  sourceType: string
  sourceId: string
  type?: string
  title: string
}

export interface FeedbackNotificationDynamicTab {
  key: string
  label: string
  icon: string
  path: string
}

export interface FeedbackNotificationOpenDependencies {
  markRead: () => Promise<boolean>
  feedbackDetailContentKey: (sourceID: string) => string
  openDynamicTab: (tab: FeedbackNotificationDynamicTab) => boolean | Promise<boolean>
  closePanel: () => void
  fallbackLabel: string
}

export function isFeedbackNotification(
  notification: Pick<FeedbackNotificationOpenInput, 'sourceType' | 'type'>,
) {
  return notification.sourceType === 'user_feedback' || notification.type === 'feedback_status'
}

export async function openFeedbackNotification(
  notification: FeedbackNotificationOpenInput,
  dependencies: FeedbackNotificationOpenDependencies,
) {
  if (!isFeedbackNotification(notification))
    return false

  const marked = await dependencies.markRead()
  if (!marked)
    return true

  const key = dependencies.feedbackDetailContentKey(notification.sourceId)
  if (!key)
    return true

  const opened = await dependencies.openDynamicTab({
    key,
    label: String(notification.title || '').trim() || dependencies.fallbackLabel,
    icon: 'chat',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  })
  if (!opened)
    return true
  dependencies.closePanel()
  return true
}
