import { clearPerms } from './permission'

export const ADMIN_ROUTE_TABS_STORAGE_KEY = 'admin_route_tabs'

export function resetAdminRouteTabs() {
  localStorage.removeItem(ADMIN_ROUTE_TABS_STORAGE_KEY)
}

export function clearAdminSession() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_info')
  resetAdminRouteTabs()
  clearPerms()
}
