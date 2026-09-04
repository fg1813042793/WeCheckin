import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { clearAdminSession } from './adminSession'

export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

export type ApiRequest = Omit<AxiosInstance, 'get' | 'post' | 'put' | 'delete' | 'patch'> & {
  get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  post<T = unknown, D = unknown>(url: string, data?: D, config?: AxiosRequestConfig<D>): Promise<ApiResponse<T>>
  put<T = unknown, D = unknown>(url: string, data?: D, config?: AxiosRequestConfig<D>): Promise<ApiResponse<T>>
  patch<T = unknown, D = unknown>(url: string, data?: D, config?: AxiosRequestConfig<D>): Promise<ApiResponse<T>>
}

const REQUEST_ERROR_NOTIFIED = '__adminRequestErrorNotified'

function markRequestErrorNotified(error: unknown) {
  if (error && typeof error === 'object') {
    try {
      Object.defineProperty(error, REQUEST_ERROR_NOTIFIED, { value: true, configurable: true })
    } catch {
      // Preserve the original request error even when a library freezes it.
    }
  }
  return error
}

export function isRequestErrorNotified(error: unknown): boolean {
  return Boolean(error && typeof error === 'object' && REQUEST_ERROR_NOTIFIED in error)
}

export function showRequestError(error: unknown, fallback: string) {
  if (isRequestErrorNotified(error)) return
  const message = error && typeof error === 'object' && 'msg' in error && typeof error.msg === 'string'
    ? error.msg
    : fallback
  ElMessage.error(message)
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

function encodeFormBody(data: unknown) {
  const params = new URLSearchParams()
  if (!data || typeof data !== 'object') return params.toString()
  for (const [key, value] of Object.entries(data)) {
    if (value !== undefined && value !== null) {
      params.append(key, String(value))
    }
  }
  return params.toString()
}

const axiosInstance = axios.create({
  baseURL: '',
  timeout: 15000,
  transformRequest: [(data: unknown, headers) => {
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
    return Promise.reject(markRequestErrorNotified(res.data))
  },
  err => {
    ElMessage.error('网络错误')
    return Promise.reject(markRequestErrorNotified(err))
  }
)

const request = axiosInstance as ApiRequest

export default request
