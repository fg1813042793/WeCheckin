export const CLIENT_TOKEN_KEY = 'token'
export const CLIENT_INFO_KEY = 'userInfo'
export const CLIENT_PASSWORD_STATE_KEY = 'client_password_state'
export const ADMIN_TOKEN_KEY = 'admin_token'
export const ADMIN_INFO_KEY = 'admin_info'

export const AUTH_STORAGE_KEYS = [
  CLIENT_TOKEN_KEY,
  CLIENT_INFO_KEY,
  CLIENT_PASSWORD_STATE_KEY,
  ADMIN_TOKEN_KEY,
  ADMIN_INFO_KEY
]

let pendingAuthNavigation = ''

function readStorage(key, fallback = '') {
  try {
    const value = uni.getStorageSync(key)
    return value === undefined ? fallback : value
  } catch (e) {
    return fallback
  }
}

function writeStorage(key, value) {
  try {
    uni.setStorageSync(key, value)
  } catch (e) {
    // ignore storage write failures; callers keep their current UI flow
  }
}

function removeStorage(key) {
  try {
    uni.removeStorageSync(key)
  } catch (e) {
    // ignore storage cleanup failures
  }
}

export function getClientToken() {
  return readStorage(CLIENT_TOKEN_KEY, '') || ''
}

export function getClientUserInfo() {
  return readStorage(CLIENT_INFO_KEY, null)
}

export function getClientAuth() {
  return {
    token: getClientToken(),
    info: getClientUserInfo()
  }
}

export function getClientPasswordState() {
  return readStorage(CLIENT_PASSWORD_STATE_KEY, {}) || {}
}

export function clientPasswordChangeRequired() {
  return getClientPasswordState().passwordChangeRequired === true
}

export function markClientPasswordChangeRequired() {
  writeStorage(CLIENT_PASSWORD_STATE_KEY, { ...getClientPasswordState(), passwordChangeRequired: true })
}

export function clearClientPasswordChangeRequired() {
  writeStorage(CLIENT_PASSWORD_STATE_KEY, {
    passwordChangeRequired: false,
    passwordExpired: false,
    passwordExpiresAt: 0,
    passwordExpiryWarning: false
  })
}

export function setClientUserInfo(userInfo) {
  if (userInfo) {
    writeStorage(CLIENT_INFO_KEY, userInfo)
  }
  writeStorage(CLIENT_PASSWORD_STATE_KEY, {
    passwordChangeRequired: data.passwordChangeRequired === true,
    passwordExpired: data.passwordExpired === true,
    passwordExpiresAt: Number(data.passwordExpiresAt || 0),
    passwordExpiryWarning: data.passwordExpiryWarning === true
  })
}

export function setClientAuth(data = {}) {
  const token = data.token || ''
  const userInfo = data.userInfo || data
  if (token) {
    writeStorage(CLIENT_TOKEN_KEY, token)
  }
  if (userInfo) {
    writeStorage(CLIENT_INFO_KEY, userInfo)
  }
}

export function clearClientAuth() {
  removeStorage(CLIENT_TOKEN_KEY)
  removeStorage(CLIENT_INFO_KEY)
  removeStorage(CLIENT_PASSWORD_STATE_KEY)
  removeStorage('clientMenuPermissionKeys')
  removeStorage('clientApiPermissionKeys')
  removeStorage('clientMenus')
  removeStorage('clientPermissionSnapshot')
}

export function hasClientAuth() {
  return !!getClientToken()
}

export function getClientUserId(fallback = '') {
  if (fallback) return fallback
  const userInfo = getClientUserInfo()
  return (userInfo && (userInfo.miniOpenID || userInfo.id)) || ''
}

export function getAdminToken() {
  return readStorage(ADMIN_TOKEN_KEY, '') || ''
}

export function getAdminInfo() {
  return readStorage(ADMIN_INFO_KEY, null)
}

export function getAdminAuth() {
  return {
    token: getAdminToken(),
    info: getAdminInfo()
  }
}

export function adminPasswordChangeRequired() {
  const info = getAdminInfo() || {}
  return info.passwordChangeRequired === true
}

export function markAdminPasswordChangeRequired() {
  writeStorage(ADMIN_INFO_KEY, { ...(getAdminInfo() || {}), passwordChangeRequired: true })
}

export function clearAdminPasswordChangeRequired() {
  writeStorage(ADMIN_INFO_KEY, {
    ...(getAdminInfo() || {}),
    passwordChangeRequired: false,
    passwordExpired: false,
    passwordExpiresAt: 0,
    passwordExpiryWarning: false
  })
}

export function setAdminAuth(data = {}) {
  if (data.token) {
    writeStorage(ADMIN_TOKEN_KEY, data.token)
  }
  writeStorage(ADMIN_INFO_KEY, data)
}

export function clearAdminAuth() {
  removeStorage(ADMIN_TOKEN_KEY)
  removeStorage(ADMIN_INFO_KEY)
}

export function hasAdminAuth() {
  return !!getAdminToken()
}

export function getRequestAuthState(isAdmin) {
  return {
    isAdmin,
    token: isAdmin ? getAdminToken() : getClientToken(),
    loginUrl: isAdmin ? '/pages/admin/admin_login' : '/pages/login/login',
    passwordChangeUrl: isAdmin ? '/pages/admin/mgr/admin_mgr_pwd' : '/pages/my/change_password'
  }
}

export function clearRequestAuthState(authState) {
  if (authState && authState.isAdmin) {
    clearAdminAuth()
  } else {
    clearClientAuth()
  }
}

export function navigateToAuthPage(loginUrl, method = 'redirectTo') {
  const target = `/${String(loginUrl || '').replace(/^\/+/, '')}`
  if (target === '/' || pendingAuthNavigation === target) return false

  const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []
  const currentRoute = pages.length ? `/${String(pages[pages.length - 1].route || '').replace(/^\/+/, '')}` : ''
  if (currentRoute === target) return false

  pendingAuthNavigation = target
  const navigate = method === 'reLaunch' ? uni.reLaunch : uni.redirectTo
  navigate({
    url: target,
    complete: () => {
      if (pendingAuthNavigation === target) pendingAuthNavigation = ''
    }
  })
  return true
}
