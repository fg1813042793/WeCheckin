import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { clearAdminSession } from './adminSession'

export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

export type ApiRequest = Omit<AxiosInstance, 'get' | 'post' | 'put' | 'delete' | 'patch'> & {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  post<T = any, D = any>(url: string, data?: D, config?: AxiosRequestConfig<D>): Promise<ApiResponse<T>>
  put<T = any, D = any>(url: string, data?: D, config?: AxiosRequestConfig<D>): Promise<ApiResponse<T>>
  patch<T = any, D = any>(url: string, data?: D, config?: AxiosRequestConfig<D>): Promise<ApiResponse<T>>
}

const LOGIN_EXPIRED_MESSAGES = new Set([
  '未登录',
  '登录已过期',
  '登录已过期或已被强制下线',
  '账号异常'
])

let redirectingToLogin = false

function redirectToLogin() {
  if (redirectingToLogin || window.location.pathname === '/login') return
  redirectingToLogin = true
  window.location.href = '/login'
}

function encodeFormBody(data: any) {
  const params = new URLSearchParams()
  if (!data) return params.toString()
  for (const key in data) {
    if (data[key] !== undefined && data[key] !== null) {
      params.append(key, String(data[key]))
    }
  }
  return params.toString()
}

const axiosInstance = axios.create({
  baseURL: '',
  timeout: 15000,
  transformRequest: [(data: any, headers: any) => {
    if (data instanceof FormData) {
      if (headers && typeof headers.delete === 'function') {
        headers.delete('Content-Type')
      } else if (headers) {
        delete headers['Content-Type']
        delete headers['content-type']
      }
      return data
    }
    return encodeFormBody(data)
  }],
  headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
})

axiosInstance.interceptors.request.use(config => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = token
  }
  return config
})

axiosInstance.interceptors.response.use(
  res => {
    if (res.data.code === 0) {
      return res.data
    }
    if (LOGIN_EXPIRED_MESSAGES.has(res.data.msg)) {
      clearAdminSession()
      redirectToLogin()
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

const request = axiosInstance as ApiRequest

export default request
