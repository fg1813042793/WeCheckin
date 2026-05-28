import CONFIG from '../config/index'

const BASE_URL = CONFIG.BASE_URL

const request = (options) => {
  return new Promise((resolve, reject) => {
    uni.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Authorization': uni.getStorageSync('token') || '',
        ...options.header
      },
      success: (res) => {
        if (res.statusCode === 200) {
          if (res.data.code === 0) {
            resolve(res.data)
          } else {
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
