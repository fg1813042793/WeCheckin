import type { Component } from 'vue'
import FeedbackCenter from './components/FeedbackCenter.vue'
import FeedbackCreatePage from './components/FeedbackCreatePage.vue'
import FeedbackDetailPage from './components/FeedbackDetailPage.vue'
import {
  FEEDBACK_CONTENT_KEY,
  FEEDBACK_CREATE_CONTENT_KEY,
  feedbackDetailIdFromContentKey,
} from './feedback-route-keys'

export {
  FEEDBACK_CONTENT_KEY,
  FEEDBACK_CREATE_CONTENT_KEY,
  FEEDBACK_DETAIL_PREFIX,
  feedbackDetailContentKey,
  feedbackDetailIdFromContentKey,
  normalizeFeedbackDynamicContentKey,
} from './feedback-route-keys'

export const feedbackContentRoutes: Record<string, Component> = {
  [FEEDBACK_CONTENT_KEY]: FeedbackCenter,
}

export function resolveFeedbackContentComponent(key: string): Component | undefined {
  if (key === FEEDBACK_CREATE_CONTENT_KEY)
    return FeedbackCreatePage
  if (feedbackDetailIdFromContentKey(key))
    return FeedbackDetailPage
  return feedbackContentRoutes[key]
}
