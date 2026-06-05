import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '',
  timeout: 15000,
  transformRequest: [(data: any) => {
    if (data instanceof FormData) return data
    const params = new URLSearchParams()
    for (const key in data) {
      if (data[key] !== undefined && data[key] !== null) {
        params.append(key, String(data[key]))
      }
    }
    return params.toString()
  }],
  headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
})

request.interceptors.request.use(config => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = token
  }
  return config
})

request.interceptors.response.use(
  res => {
    if (res.data.code === 0) {
      return res.data
    }
    if (res.data.msg === '未登录' || res.data.msg === '登录已过期' || res.data.msg === '登录已过期或已被强制下线' || res.data.msg === '账号异常') {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_info')
      window.location.href = '/login'
      return Promise.reject(res.data)
    }
    ElMessage.error(res.data.msg || '请求失败')
    return Promise.reject(res.data)
  },
  err => {
    ElMessage.error('网络错误')
    return Promise.reject(err)
  }
)

export default request
