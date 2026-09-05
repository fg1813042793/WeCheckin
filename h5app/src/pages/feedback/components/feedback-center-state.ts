export interface FeedbackListRequestParameters {
  keyword?: string
  page: number
  pageSize: number
  status?: string
}

export interface FeedbackDynamicTab {
  icon: string
  key: string
  label: string
  path: string
}

interface FeedbackVisibilityDocument {
  visibilityState: string
  addEventListener: (event: 'visibilitychange', listener: () => void) => void
  removeEventListener: (event: 'visibilitychange', listener: () => void) => void
}

export function resolveFeedbackListPage(requestedPage: number, total: number, pageSize: number) {
  const normalizedPageSize = Math.max(1, Math.floor(Number(pageSize || 0)))
  const normalizedTotal = Math.max(0, Math.floor(Number(total || 0)))
  const normalizedRequestedPage = Math.max(1, Math.floor(Number(requestedPage || 0)))
  const page = Math.min(normalizedRequestedPage, Math.max(1, Math.ceil(normalizedTotal / normalizedPageSize)))
  return {
    page,
    shouldReload: page !== normalizedRequestedPage,
  }
}

export function feedbackListRequestKey(parameters: FeedbackListRequestParameters) {
  return JSON.stringify([
    String(parameters.keyword || '').trim(),
    String(parameters.status || '').trim(),
    Math.max(1, Math.floor(Number(parameters.page || 0))),
    Math.max(1, Math.floor(Number(parameters.pageSize || 0))),
  ])
}

export function createInFlightRequestDeduper() {
  const requests = new Map<string, Promise<unknown>>()

  function run<T>(key: string, request: () => Promise<T>): Promise<T> {
    const normalizedKey = String(key || '').trim()
    const existing = requests.get(normalizedKey)
    if (existing)
      return existing as Promise<T>

    let pending: Promise<T>
    try {
      pending = Promise.resolve(request())
    }
    catch (error) {
      pending = Promise.reject(error)
    }
    requests.set(normalizedKey, pending)

    const removeSettledRequest = () => {
      if (requests.get(normalizedKey) === pending)
        requests.delete(normalizedKey)
    }
    pending.then(removeSettledRequest, removeSettledRequest)
    return pending
  }

  function clear() {
    requests.clear()
  }

  return { clear, run }
}

export function registerFeedbackVisibilityRefresh(
  visibilityDocument: FeedbackVisibilityDocument,
  isFeedbackActive: () => boolean,
  refresh: () => void,
) {
  const handleVisibilityChange = () => {
    if (visibilityDocument.visibilityState === 'visible' && isFeedbackActive())
      refresh()
  }
  visibilityDocument.addEventListener('visibilitychange', handleVisibilityChange)
  return () => visibilityDocument.removeEventListener('visibilitychange', handleVisibilityChange)
}

export function createFeedbackDynamicTab(key: string, label: string, icon: string): FeedbackDynamicTab | null {
  const normalizedKey = String(key || '').trim()
  if (!normalizedKey)
    return null
  return {
    key: normalizedKey,
    label: String(label || normalizedKey).trim() || normalizedKey,
    icon: String(icon || 'chat').trim() || 'chat',
    path: `/pages/index/index?view=${encodeURIComponent(normalizedKey)}`,
  }
}
