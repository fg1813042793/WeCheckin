import { passportApi } from '../api/index'

export const CLIENT_MENU_KEY = 'clientMenuPermissionKeys'
export const CLIENT_API_KEY = 'clientApiPermissionKeys'
export const CLIENT_MENUS_KEY = 'clientMenus'
export const CLIENT_PERMISSION_SNAPSHOT_KEY = 'clientPermissionSnapshot'

const TAB_PAGES = new Set([
  '/pages/index/index',
  '/pages/news/news_index',
  '/pages/enroll/enroll_index',
  '/pages/event/event_index',
  '/pages/my/my_index'
])

function readStorage(key, fallback) {
  try {
    const value = uni.getStorageSync(key)
    return value === undefined || value === '' ? fallback : value
  } catch (e) {
    return fallback
  }
}

function writeStorage(key, value) {
  try {
    uni.setStorageSync(key, value)
  } catch (e) {
    // storage failures should not make the UI crash
  }
}

function normalizeKeys(values) {
  if (!Array.isArray(values)) return []
  const seen = new Set()
  const result = []
  values.forEach((value) => {
    const key = String(value || '').trim()
    if (key && !seen.has(key)) {
      seen.add(key)
      result.push(key)
    }
  })
  return result
}

function menuKeysFromMenus(menus) {
  if (!Array.isArray(menus)) return []
  return normalizeKeys(menus.map((item) => item.permissionKey || item.key))
}

export function setClientPermissionSnapshot(payload = {}) {
  const menus = Array.isArray(payload.menus) ? payload.menus : []
  const menuKeys = normalizeKeys(payload.menuPermissionKeys || payload.menuKeys || menuKeysFromMenus(menus))
  const apiKeys = normalizeKeys(payload.apiPermissionKeys || [])
  const snapshot = {
    menuReady: !!payload.menuPermissionReady,
    apiReady: !!payload.apiPermissionReady,
    permissionVersion: payload.permissionVersion || 0
  }
  writeStorage(CLIENT_MENU_KEY, menuKeys)
  writeStorage(CLIENT_API_KEY, apiKeys)
  writeStorage(CLIENT_MENUS_KEY, menus)
  writeStorage(CLIENT_PERMISSION_SNAPSHOT_KEY, snapshot)
  return {
    menus,
    menuPermissionKeys: menuKeys,
    apiPermissionKeys: apiKeys,
    ...snapshot
  }
}

export async function ensureClientPermissionSnapshot() {
  const res = await passportApi.bootstrap()
  return setClientPermissionSnapshot(res.data || {})
}

export function getClientMenuPermissionKeys() {
  return normalizeKeys(readStorage(CLIENT_MENU_KEY, []))
}

export function getClientMenus() {
  const menus = readStorage(CLIENT_MENUS_KEY, [])
  return Array.isArray(menus) ? menus : []
}

export function hasClientMenuPermission(permissionKey) {
  const key = String(permissionKey || '').trim()
  if (!key) return true
  return getClientMenuPermissionKeys().includes(key)
}

export function filterClientMenus(menus = []) {
  if (!Array.isArray(menus)) return []
  return menus.filter((item) => hasClientMenuPermission(item.permissionKey || item.key))
}

export function openClientMenu(menu) {
  const url = menu && (menu.url || menu.path)
  if (!url) return
  if (TAB_PAGES.has(url)) {
    uni.switchTab({ url })
    return
  }
  uni.navigateTo({ url })
}

export async function guardClientMenuPage(permissionKey) {
  try {
    await ensureClientPermissionSnapshot()
  } catch (e) {
    return false
  }
  if (hasClientMenuPermission(permissionKey)) {
    return true
  }
  uni.showToast({ title: '无权限访问', icon: 'none' })
  const firstMenu = getClientMenus().find((item) => hasClientMenuPermission(item.permissionKey))
  if (firstMenu && firstMenu.path) {
    setTimeout(() => openClientMenu(firstMenu), 300)
  }
  return false
}
