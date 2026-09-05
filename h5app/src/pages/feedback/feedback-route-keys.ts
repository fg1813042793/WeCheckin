export const FEEDBACK_CONTENT_KEY = 'feedback'
export const FEEDBACK_CREATE_CONTENT_KEY = 'feedback:create'
export const FEEDBACK_DETAIL_PREFIX = 'feedback:detail:'
const MAX_UINT64 = 18446744073709551615n

function normalizeFeedbackDetailID(id: string) {
  const value = String(id).trim()
  if (!/^[1-9]\d*$/.test(value)) {
    return ''
  }
  try {
    return BigInt(value) <= MAX_UINT64 ? value : ''
  }
  catch {
    return ''
  }
}

export function feedbackDetailContentKey(id: string) {
  const value = normalizeFeedbackDetailID(id)
  return value ? `${FEEDBACK_DETAIL_PREFIX}${encodeURIComponent(value)}` : ''
}

export function feedbackDetailIdFromContentKey(key: string) {
  if (!key.startsWith(FEEDBACK_DETAIL_PREFIX))
    return ''
  const encodedID = key.slice(FEEDBACK_DETAIL_PREFIX.length)
  if (!encodedID)
    return ''
  try {
    const id = normalizeFeedbackDetailID(decodeURIComponent(encodedID))
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
