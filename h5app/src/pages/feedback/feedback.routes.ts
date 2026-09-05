import type { Component } from 'vue'
import {
  FEEDBACK_CONTENT_KEY,
  FEEDBACK_CREATE_CONTENT_KEY,
  feedbackDetailIdFromContentKey,
} from './feedback-route-keys'

const FeedbackContractPlaceholder: Component = () => null

export {
  FEEDBACK_CONTENT_KEY,
  FEEDBACK_CREATE_CONTENT_KEY,
  FEEDBACK_DETAIL_PREFIX,
  feedbackDetailContentKey,
  feedbackDetailIdFromContentKey,
  normalizeFeedbackDynamicContentKey,
} from './feedback-route-keys'

export const feedbackContentRoutes: Record<string, Component> = {
  [FEEDBACK_CONTENT_KEY]: FeedbackContractPlaceholder,
}

export function resolveFeedbackContentComponent(key: string): Component | undefined {
  if (key === FEEDBACK_CREATE_CONTENT_KEY || feedbackDetailIdFromContentKey(key))
    return FeedbackContractPlaceholder
  return feedbackContentRoutes[key]
}
