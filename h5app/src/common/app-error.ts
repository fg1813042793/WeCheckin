const PRIVATE_ERROR_PATTERNS = [
  /\bbearer\b/i,
  /\b(?:token|authorization)\b/i,
  /(?:\/Users\/|\/home\/|[A-Za-z]:\\)/,
  /(?:\bat\s+\S+\.(?:ts|js|vue):\d+|stack\s*trace)/i,
]

function cleanMessage(value: unknown) {
  if (typeof value !== 'string')
    return ''
  const message = value.trim()
  if (!message || PRIVATE_ERROR_PATTERNS.some(pattern => pattern.test(message)))
    return ''
  return Array.from(message).slice(0, 240).join('')
}

function recordValue(record: Record<string, unknown>, key: string) {
  return cleanMessage(record[key])
}

export function appErrorMessage(error: unknown, fallback: string) {
  const fallbackMessage = cleanMessage(fallback) || '操作失败'
  if (!error || typeof error !== 'object')
    return fallbackMessage

  const source = error as Record<string, unknown>
  const directMessage = recordValue(source, 'msg') || recordValue(source, 'message')
  if (directMessage)
    return directMessage

  const data = source.data && typeof source.data === 'object' && !Array.isArray(source.data)
    ? source.data as Record<string, unknown>
    : null
  const dataMessage = data && (recordValue(data, 'msg') || recordValue(data, 'message'))
  if (dataMessage)
    return dataMessage

  const response = source.response && typeof source.response === 'object' && !Array.isArray(source.response)
    ? source.response as Record<string, unknown>
    : null
  const responseData = response?.data && typeof response.data === 'object' && !Array.isArray(response.data)
    ? response.data as Record<string, unknown>
    : null
  const responseMessage = responseData && (recordValue(responseData, 'msg') || recordValue(responseData, 'message'))
  if (responseMessage)
    return responseMessage

  return recordValue(source, 'errMsg') || fallbackMessage
}

export function showAppError(error: unknown, fallback: string) {
  const message = appErrorMessage(error, fallback)
  uni.showToast({ title: message, icon: 'none' })
  return message
}
