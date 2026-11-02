import type { AdminMenuItem } from '../api/types'

export function visibleAdminMenuTree(items: AdminMenuItem[]): AdminMenuItem[] {
  const result: AdminMenuItem[] = []
  for (const item of items) {
    if (item.hidden === 1 && item.type !== 2) continue
    const children = Array.isArray(item.children) ? visibleAdminMenuTree(item.children) : []
    result.push({
      ...item,
      children: children.length > 0 ? children : undefined,
    })
  }
  return result
}
