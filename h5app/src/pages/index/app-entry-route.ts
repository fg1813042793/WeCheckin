const routeQueryKeys = new Set(['view', 'key', 'page', 'reviewId', 'review_id', 'id'])

function appendRouteQuery(target: Record<string, string>, params: URLSearchParams) {
  params.forEach((value, key) => {
    if (routeQueryKeys.has(key) && value.trim())
      target[key] = value
  })
}

export function appEntryRouteQueryFromURL(value: string) {
  const result: Record<string, string> = {}
  if (!value.trim())
    return result
  try {
    const parsed = new URL(value)
    appendRouteQuery(result, parsed.searchParams)
    const queryIndex = parsed.hash.indexOf('?')
    if (queryIndex >= 0)
      appendRouteQuery(result, new URLSearchParams(parsed.hash.slice(queryIndex + 1)))
  }
  catch {
    return result
  }
  return result
}

export function mergeAppEntryRouteQuery(loadQuery: Record<string, unknown>, href = '') {
  return {
    ...appEntryRouteQueryFromURL(href),
    ...loadQuery,
  }
}
