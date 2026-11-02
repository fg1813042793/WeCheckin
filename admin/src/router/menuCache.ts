import type { RouteLocationNormalizedLoaded } from 'vue-router'
import type { AdminMenuItem } from './adminRoutes'

export const ADMIN_MENU_CONFIG_CHANGED_EVENT = 'admin-menu-config-changed'

export function adminRouteCacheEnabled(
  route: Pick<RouteLocationNormalizedLoaded, 'path' | 'meta'>,
  menuTree: AdminMenuItem[],
): boolean {
  const menuPath = typeof route.meta.menuPath === 'string' ? route.meta.menuPath : route.path
  const menu = findAdminMenuByPath(menuTree, menuPath)
  return menu !== undefined && menu.cacheEnabled !== 0
}

export function notifyAdminMenuConfigChanged() {
  window.dispatchEvent(new CustomEvent(ADMIN_MENU_CONFIG_CHANGED_EVENT))
}

function findAdminMenuByPath(items: AdminMenuItem[], path: string): AdminMenuItem | undefined {
  for (const item of items) {
    if (item.path === path) return item
    const nested = item.children?.length ? findAdminMenuByPath(item.children, path) : undefined
    if (nested) return nested
  }
  return undefined
}
