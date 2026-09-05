export interface FeedbackNotificationOpenInput {
  sourceType: string
  sourceId: string
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
  openDynamicTab: (tab: FeedbackNotificationDynamicTab) => void
  closePanel: () => void
  fallbackLabel: string
}

export async function openFeedbackNotification(
  notification: FeedbackNotificationOpenInput,
  dependencies: FeedbackNotificationOpenDependencies,
) {
  if (notification.sourceType !== 'user_feedback')
    return false

  const marked = await dependencies.markRead()
  if (!marked)
    return true

  const key = dependencies.feedbackDetailContentKey(notification.sourceId)
  if (!key)
    return true

  dependencies.openDynamicTab({
    key,
    label: String(notification.title || '').trim() || dependencies.fallbackLabel,
    icon: 'chat',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  })
  dependencies.closePanel()
  return true
}
