import CONFIG from '../config/index'

const BASE_URL = CONFIG.BASE_URL

const request = (options) => {
  return new Promise((resolve, reject) => {
    const isAdmin = options.url.startsWith('/admin/')
    const token = isAdmin ? uni.getStorageSync('admin_token') : uni.getStorageSync('token')
    uni.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Authorization': token || '',
        ...options.header
      },
      success: (res) => {
        if (res.statusCode === 200) {
          if (res.data.code === 0) {
            resolve(res.data)
          } else {
            if (res.data.msg === '未登录' || res.data.msg === '登录已过期' || res.data.msg === '登录已过期或已被强制下线') {
              if (isAdmin) {
                uni.removeStorageSync('admin_token')
                uni.removeStorageSync('admin_info')
                uni.redirectTo({ url: '/pages/admin/admin_login' })
              } else {
                uni.removeStorageSync('token')
                uni.removeStorageSync('userInfo')
                uni.redirectTo({ url: '/pages/login/login' })
              }
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

export {
  request,
  get,
  post,
  put,
  del,
  BASE_URL
}
