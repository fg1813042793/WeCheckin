import type { ApiEnvelope } from '@/types/dingtalk-h5'
import { http } from 'uview-pro'
import { DINGTALK_H5_CONFIG } from '@/config/dingtalk-h5'

export const API_V2 = '/api/v2'
export const DINGTALK_H5_API = `${API_V2}/dingtalk/h5`

type QueryValue = string | number | boolean | null | undefined
type UploadFormValue = string | number | boolean

export interface UploadFileOptions {
  name?: string
  formData?: Record<string, UploadFormValue>
  timeout?: number
  header?: Record<string, string>
}

function trimRightSlash(value: string) {
  return value.replace(/\/+$/, '')
}

function isAbsoluteUrl(url: string) {
  return /^https?:\/\//i.test(url)
}

function withBaseUrl(url: string) {
  if (isAbsoluteUrl(url)) {
    return url
  }
  const baseUrl = trimRightSlash(DINGTALK_H5_CONFIG.BASE_URL)
  if (!baseUrl) {
    return url
  }
  return `${baseUrl}${url.startsWith('/') ? url : `/${url}`}`
}

export function authToken() {
  return String(uni.getStorageSync(DINGTALK_H5_CONFIG.TOKEN_KEY) || '')
}

export function setAuthToken(token: string) {
  uni.setStorageSync(DINGTALK_H5_CONFIG.TOKEN_KEY, token || '')
}

export function clearAuthToken() {
  uni.removeStorageSync(DINGTALK_H5_CONFIG.TOKEN_KEY)
}

export function buildApiUrl(url: string, query: Record<string, QueryValue> = {}) {
  const fullUrl = withBaseUrl(url)
  const params = Object.entries(query)
    .filter(([, value]) => value !== undefined && value !== null && String(value) !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)
    .join('&')

  if (!params) {
    return fullUrl
  }
  return `${fullUrl}${fullUrl.includes('?') ? '&' : '?'}${params}`
}

export function get<T>(url: string, data: unknown = {}) {
  return http.get<ApiEnvelope<T>>(url, data)
}

export function post<T>(url: string, data: unknown = {}) {
  return http.post<ApiEnvelope<T>>(url, data)
}

export function put<T>(url: string, data: unknown = {}) {
  return http.put<ApiEnvelope<T>>(url, data)
}

export function patch<T>(url: string, data: unknown = {}) {
  return http.request<ApiEnvelope<T>>({
    url,
    method: 'PATCH' as 'POST',
    data,
    header: {},
  })
}

export function del<T>(url: string, data: unknown = {}) {
  return http.delete<ApiEnvelope<T>>(url, data)
}

function parseUploadResponseData<T>(raw: unknown): ApiEnvelope<T> | null {
  if (raw && typeof raw === 'object') {
    return raw as ApiEnvelope<T>
  }
  if (typeof raw !== 'string') {
    return null
  }
  try {
    return JSON.parse(raw) as ApiEnvelope<T>
  }
  catch {
    return null
  }
}

export function uploadFile<T>(url: string, filePath: string, options: UploadFileOptions = {}) {
  return new Promise<ApiEnvelope<T>>((resolve, reject) => {
    uni.uploadFile({
      url: buildApiUrl(url),
      filePath,
      name: options.name || 'file',
      formData: options.formData || {},
      timeout: options.timeout || 30000,
      header: {
        'Authorization': authToken(),
        'X-Client-Platform': 'dingtalk-h5',
        ...options.header,
      },
      success: (res) => {
        const statusCode = Number(res.statusCode || 0)
        if (statusCode < 200 || statusCode >= 300) {
          uni.showToast({ title: '上传失败', icon: 'none' })
          reject(res)
          return
        }

        const data = parseUploadResponseData<T>(res.data)
        if (!data) {
          uni.showToast({ title: '上传响应异常', icon: 'none' })
          reject(res)
          return
        }
        if (data.code === undefined || data.code === 0) {
          resolve(data)
          return
        }
        reject(data)
      },
      fail: (error) => {
        uni.showToast({ title: '上传失败', icon: 'none' })
        reject(error)
      },
    })
  })
}
