import CONFIG from '../config/index'
import { clearRequestAuthState, getRequestAuthState } from './auth'

const BASE_URL = CONFIG.BASE_URL
const LOGIN_EXPIRED_MESSAGES = new Set([
  '未登录',
  '登录已过期',
  '登录已过期或已被强制下线',
  '账号异常'
])

let redirectingToLogin = false
const inflightRequests = new Map()

function getAuthState(isAdmin) {
  return getRequestAuthState(isAdmin)
}

function clearAuthState(authState) {
  clearRequestAuthState(authState)
}

function redirectToLogin(authState) {
  if (redirectingToLogin) return
  redirectingToLogin = true
  uni.redirectTo({
    url: authState.loginUrl,
    complete: () => {
      redirectingToLogin = false
    }
  })
}

function normalizeRequestData(value) {
  if (Array.isArray(value)) {
    return value.map((item) => normalizeRequestData(item))
  }
  if (value && typeof value === 'object') {
    return Object.keys(value).sort().reduce((result, key) => {
      result[key] = normalizeRequestData(value[key])
      return result
    }, {})
  }
  return value
}

function buildRequestKey(options, method) {
  const data = normalizeRequestData(options.data || {})
  return `${method} ${options.url} ${JSON.stringify(data)}`
}

const request = (options) => {
  const method = (options.method || 'GET').toUpperCase()
  const shouldDedupe = method === 'GET' && options.dedupe !== false
  const requestKey = shouldDedupe ? buildRequestKey(options, method) : ''
  if (shouldDedupe && inflightRequests.has(requestKey)) {
    return inflightRequests.get(requestKey)
  }

  const promise = new Promise((resolve, reject) => {
    const isAdmin = options.url.startsWith('/admin/') || options.url.startsWith('/api/v2/admin/')
    const authState = getAuthState(isAdmin)
    uni.request({
      url: BASE_URL + options.url,
      method,
      data: options.data || {},
      timeout: options.timeout || 15000,
      header: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Authorization': authState.token || '',
        ...options.header
      },
      success: (res) => {
        if (res.statusCode === 200) {
          if (res.data.code === 0) {
            resolve(res.data)
          } else {
            if (LOGIN_EXPIRED_MESSAGES.has(res.data.msg)) {
              clearAuthState(authState)
              redirectToLogin(authState)
              reject(res.data)
              return
            }
            uni.showToast({
              title: res.data.msg || '请求失败',
              icon: 'none'
            })
            reject(res.data)
          }
        } else {
          uni.showToast({
            title: '网络错误',
            icon: 'none'
          })
          reject(res)
        }
      },
      fail: (err) => {
        uni.showToast({
          title: '网络连接失败',
          icon: 'none'
        })
        reject(err)
      }
    })
  })
  if (shouldDedupe) {
    inflightRequests.set(requestKey, promise)
    promise.then(
      () => inflightRequests.delete(requestKey),
      () => inflightRequests.delete(requestKey)
    )
  }
  return promise
}

const get = (url, data = {}) => {
  return request({ url, method: 'GET', data })
}

const post = (url, data = {}) => {
  return request({ url, method: 'POST', data })
}

const put = (url, data = {}) => {
  return request({ url, method: 'PUT', data })
}

const patch = (url, data = {}) => {
  return request({ url, method: 'PATCH', data })
}

const del = (url, data = {}) => {
  return request({ url, method: 'DELETE', data })
}

const postJSON = (url, data = {}) => {
  return request({
    url,
    method: 'POST',
    data,
    header: { 'Content-Type': 'application/json' }
  })
}

export {
  request,
  get,
  post,
  postJSON,
  put,
  patch,
  del,
  BASE_URL
}
