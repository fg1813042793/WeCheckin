import CONFIG from '../config'

const LOGIN_EXPIRED_MESSAGES = new Set([
  '未登录',
  '登录已过期',
  '登录已过期或已被强制下线',
  '账号异常'
])

const PERMISSION_DENIED_MESSAGES = new Set([
  '无权限访问'
])

export const AUTH_EXPIRED_EVENT = 'dingtalk-h5-auth-expired'

function trimRightSlash(value) {
  return String(value || '').replace(/\/+$/, '')
}

function buildUrl(url) {
  if (/^https?:\/\//i.test(url)) return url
  const baseUrl = trimRightSlash(CONFIG.BASE_URL)
  if (!baseUrl) return url
  return `${baseUrl}${url}`
}

function getAuthToken() {
  return uni.getStorageSync(CONFIG.TOKEN_KEY) || ''
}

export function setAuthToken(token) {
  uni.setStorageSync(CONFIG.TOKEN_KEY, token || '')
}

export function clearAuthToken() {
  uni.removeStorageSync(CONFIG.TOKEN_KEY)
}

export function authToken() {
  return getAuthToken()
}

export function isAuthExpiredError(error) {
  return LOGIN_EXPIRED_MESSAGES.has(error?.msg)
}

export function isPermissionDeniedError(error) {
  return PERMISSION_DENIED_MESSAGES.has(error?.msg)
}

export function isBindRequiredError(error) {
  return Number(error?.code) === 10020
}

function handleBusinessError(data, reject) {
  if (isAuthExpiredError(data)) {
    uni.removeStorageSync(CONFIG.TOKEN_KEY)
    uni.$emit(AUTH_EXPIRED_EVENT, data)
  }

  if (isBindRequiredError(data)) {
    reject(data)
    return
  }

  uni.showToast({
    title: data?.msg || '请求失败',
    icon: 'none'
  })
  reject(data)
}

export function request(options = {}) {
  return new Promise((resolve, reject) => {
    uni.request({
      url: buildUrl(options.url),
      method: (options.method || 'GET').toUpperCase(),
      data: options.data || {},
      timeout: options.timeout || 15000,
      header: {
        'Content-Type': 'application/json',
        'Authorization': getAuthToken(),
        'X-Client-Platform': 'dingtalk-h5',
        ...options.header
      },
      success: (res) => {
        if (res.statusCode < 200 || res.statusCode >= 300) {
          uni.showToast({
            title: '网络错误',
            icon: 'none'
          })
          reject(res)
          return
        }

        if (res.data?.code === undefined || res.data?.code === 0) {
          resolve(res.data)
          return
        }

        handleBusinessError(res.data, reject)
      },
      fail: (error) => {
        uni.showToast({
          title: '网络连接失败',
          icon: 'none'
        })
        reject(error)
      }
    })
  })
}

function parseUploadResponseData(raw) {
  if (typeof raw !== 'string') return raw
  try {
    return JSON.parse(raw)
  } catch (error) {
    return null
  }
}

export function uploadFile(url, filePath, options = {}) {
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: buildUrl(url),
      filePath,
      name: options.name || 'file',
      formData: options.formData || {},
      timeout: options.timeout || 30000,
      header: {
        'Authorization': getAuthToken(),
        'X-Client-Platform': 'dingtalk-h5',
        ...options.header
      },
      success: (res) => {
        if (res.statusCode < 200 || res.statusCode >= 300) {
          uni.showToast({
            title: '上传失败',
            icon: 'none'
          })
          reject(res)
          return
        }

        const data = parseUploadResponseData(res.data)
        if (!data) {
          uni.showToast({
            title: '上传响应异常',
            icon: 'none'
          })
          reject(res)
          return
        }
        if (data.code === undefined || data.code === 0) {
          resolve(data)
          return
        }

        handleBusinessError(data, reject)
      },
      fail: (error) => {
        uni.showToast({
          title: '上传失败',
          icon: 'none'
        })
        reject(error)
      }
    })
  })
}

export function buildApiUrl(url, query = {}) {
  const fullUrl = buildUrl(url)
  const params = Object.entries(query)
    .filter(([, value]) => value !== undefined && value !== null && String(value) !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)
    .join('&')
  if (!params) return fullUrl
  return `${fullUrl}${fullUrl.includes('?') ? '&' : '?'}${params}`
}

export function get(url, data = {}, options = {}) {
  return request({
    ...options,
    url,
    method: 'GET',
    data
  })
}

export function post(url, data = {}, options = {}) {
  return request({
    ...options,
    url,
    method: 'POST',
    data
  })
}

export function put(url, data = {}, options = {}) {
  return request({
    ...options,
    url,
    method: 'PUT',
    data
  })
}

export function patch(url, data = {}, options = {}) {
  return request({
    ...options,
    url,
    method: 'PATCH',
    data
  })
}

export function del(url, data = {}, options = {}) {
  return request({
    ...options,
    url,
    method: 'DELETE',
    data
  })
}
