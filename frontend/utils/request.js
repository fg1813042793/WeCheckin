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

const request = (options) => {
  return new Promise((resolve, reject) => {
    const isAdmin = options.url.startsWith('/admin/')
    const authState = getAuthState(isAdmin)
    uni.request({
      url: BASE_URL + options.url,
      method: (options.method || 'GET').toUpperCase(),
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
  del,
  BASE_URL
}
