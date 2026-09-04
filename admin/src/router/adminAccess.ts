import type { RouteMeta } from 'vue-router'
import { adminApi } from '../api'
import type { AdminMenuItem } from '../api/types'
import { setPerms } from '../utils/permission'

export interface AdminAccessSnapshot {
  menus: AdminMenuItem[]
  menuPaths: Set<string>
  permissions: string[]
}

let accessPromise: Promise<AdminAccessSnapshot> | null = null
let accessToken = ''

function collectMenuPaths(items: AdminMenuItem[], result = new Set<string>()) {
  for (const item of items) {
    if (item.status !== 1) continue
    if (item.type !== 2 && item.path) result.add(item.path)
    if (Array.isArray(item.children)) collectMenuPaths(item.children, result)
  }
  return result
}

export function invalidateAdminAccessSnapshot() {
  accessPromise = null
  accessToken = ''
}

export function loadAdminAccessSnapshot(force = false): Promise<AdminAccessSnapshot> {
  const token = localStorage.getItem('admin_token') || ''
  if (force || token !== accessToken) {
    accessPromise = null
    accessToken = token
  }
  if (!accessPromise) {
    accessPromise = (async () => {
      const menuResponse = await adminApi.adminMenus()
      const permissionResponse = await adminApi.adminPerms()
      const menus = Array.isArray(menuResponse.data) ? menuResponse.data : []
      const permissions = Array.isArray(permissionResponse.data) ? permissionResponse.data : []
      setPerms(permissions)
      return { menus, permissions, menuPaths: collectMenuPaths(menus) }
    })().catch(error => {
      accessPromise = null
      throw error
    })
  }
  return accessPromise
}

export function canAccessAdminRoute(meta: RouteMeta, snapshot: AdminAccessSnapshot): boolean {
  if (meta.allowWithoutMenu) return true
  return typeof meta.menuPath === 'string' && snapshot.menuPaths.has(meta.menuPath)
}
