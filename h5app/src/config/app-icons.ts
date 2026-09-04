const appMenuIconAliases: Record<string, string> = {
  dashboard: 'home',
  performance: 'list',
  mine: 'file-text',
  history: 'order',
  manager: 'account-fill',
  hrbp: 'man-add',
  summary: 'order',
  org: 'setting',
  template: 'file-text',
  account: 'account-fill',
  workflow: 'checkmark-circle',
}

function firstText(...values: Array<unknown>) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

export function resolveAppMenuIcon(icon: unknown, fallback = 'grid') {
  const iconKey = firstText(icon)
  const fallbackKey = firstText(fallback, 'grid')
  if (!iconKey) {
    return appMenuIconAliases[fallbackKey] || fallbackKey
  }
  return appMenuIconAliases[iconKey] || iconKey
}
