import { clearPerms } from './permission'

export const ADMIN_ROUTE_TABS_STORAGE_KEY = 'admin_route_tabs'

interface StoredAdminPasswordState {
  passwordChangeRequired?: boolean
  passwordExpired?: boolean
  passwordExpiresAt?: number
  passwordExpiryWarning?: boolean
}

function readAdminInfo(): Record<string, unknown> & StoredAdminPasswordState {
  try {
    return JSON.parse(localStorage.getItem('admin_info') || '{}')
  } catch {
    return {}
  }
}

function writeAdminInfo(info: Record<string, unknown>) {
  localStorage.setItem('admin_info', JSON.stringify(info))
}

export function adminPasswordChangeRequired() {
  return readAdminInfo().passwordChangeRequired === true
}

export function markAdminPasswordChangeRequired() {
  writeAdminInfo({ ...readAdminInfo(), passwordChangeRequired: true })
}

export function clearAdminPasswordChangeRequired() {
  writeAdminInfo({
    ...readAdminInfo(),
    passwordChangeRequired: false,
    passwordExpired: false,
    passwordExpiresAt: 0,
    passwordExpiryWarning: false,
  })
}

export function resetAdminRouteTabs() {
  localStorage.removeItem(ADMIN_ROUTE_TABS_STORAGE_KEY)
}

export function clearAdminSession() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_info')
  resetAdminRouteTabs()
  clearPerms()
}
