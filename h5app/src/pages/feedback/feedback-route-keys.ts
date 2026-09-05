export const FEEDBACK_CONTENT_KEY = 'feedback'
export const FEEDBACK_CREATE_CONTENT_KEY = 'feedback:create'
export const FEEDBACK_DETAIL_PREFIX = 'feedback:detail:'

export function feedbackDetailContentKey(id: string | number) {
  const value = String(id).trim()
  return value ? `${FEEDBACK_DETAIL_PREFIX}${encodeURIComponent(value)}` : ''
}

export function feedbackDetailIdFromContentKey(key: string) {
  if (!key.startsWith(FEEDBACK_DETAIL_PREFIX))
    return ''
  const encodedID = key.slice(FEEDBACK_DETAIL_PREFIX.length)
  if (!encodedID)
    return ''
  try {
    const id = decodeURIComponent(encodedID).trim()
    return id && feedbackDetailContentKey(id) === key ? id : ''
  }
  catch {
    return ''
  }
}

export function normalizeFeedbackDynamicContentKey(key: string) {
  if (key === FEEDBACK_CREATE_CONTENT_KEY)
    return FEEDBACK_CREATE_CONTENT_KEY
  const id = feedbackDetailIdFromContentKey(key)
  return id ? feedbackDetailContentKey(id) : ''
}
