export const ADMIN_FRONTEND_CONFIG_SETUP_KEY = 'ADMIN_FRONTEND_CONFIG'
export const ADMIN_FRONTEND_CONFIG_STORAGE_KEY = 'admin_frontend_config'
export const ADMIN_FRONTEND_CONFIG_CHANGED_EVENT = 'admin-frontend-config-changed'

export interface AdminFrontendConfig {
  tabCacheEnabled: 0 | 1
}

export const defaultAdminFrontendConfig: AdminFrontendConfig = {
  tabCacheEnabled: 1,
}

function isDisabledValue(value: unknown) {
  return value === 0 || value === '0' || value === false || value === 'false'
}

export function normalizeAdminFrontendConfig(input: unknown): AdminFrontendConfig {
  let source: any = input
  if (typeof input === 'string') {
    try {
      source = JSON.parse(input)
    } catch {
      source = { tabCacheEnabled: input }
    }
  }

  const rawValue = source?.tabCacheEnabled ?? source?.adminTabCacheEnabled ?? source?.ADMIN_TAB_CACHE_ENABLED
  return {
    tabCacheEnabled: isDisabledValue(rawValue) ? 0 : 1,
  }
}

export function loadCachedAdminFrontendConfig(): AdminFrontendConfig {
  try {
    const raw = localStorage.getItem(ADMIN_FRONTEND_CONFIG_STORAGE_KEY)
    return raw ? normalizeAdminFrontendConfig(raw) : { ...defaultAdminFrontendConfig }
  } catch {
    return { ...defaultAdminFrontendConfig }
  }
}

export function cacheAdminFrontendConfig(config: AdminFrontendConfig) {
  try {
    localStorage.setItem(ADMIN_FRONTEND_CONFIG_STORAGE_KEY, JSON.stringify(config))
  } catch {
    // Ignore storage failures, the backend setup value remains the source of truth.
  }
}

export function notifyAdminFrontendConfigChanged(config: AdminFrontendConfig) {
  window.dispatchEvent(new CustomEvent(ADMIN_FRONTEND_CONFIG_CHANGED_EVENT, { detail: config }))
}
