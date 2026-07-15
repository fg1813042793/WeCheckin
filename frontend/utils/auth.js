export const CLIENT_TOKEN_KEY = 'token'
export const CLIENT_INFO_KEY = 'userInfo'
export const ADMIN_TOKEN_KEY = 'admin_token'
export const ADMIN_INFO_KEY = 'admin_info'

export const AUTH_STORAGE_KEYS = [
  CLIENT_TOKEN_KEY,
  CLIENT_INFO_KEY,
  ADMIN_TOKEN_KEY,
  ADMIN_INFO_KEY
]

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

export function setClientUserInfo(userInfo) {
  if (userInfo) {
    writeStorage(CLIENT_INFO_KEY, userInfo)
  }
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
}

export function hasClientAuth() {
  return !!getClientToken()
}

export function getClientUserId(fallback = '') {
  if (fallback) return fallback
  const userInfo = getClientUserInfo()
  const token = getClientToken()
  return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
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

export function hasAnyAuth() {
  return hasClientAuth() || hasAdminAuth()
}

export function getRequestAuthState(isAdmin) {
  return {
    isAdmin,
    token: isAdmin ? getAdminToken() : getClientToken(),
    loginUrl: isAdmin ? '/pages/admin/admin_login' : '/pages/login/login'
  }
}

export function clearRequestAuthState(authState) {
  if (authState && authState.isAdmin) {
    clearAdminAuth()
  } else {
    clearClientAuth()
  }
}
