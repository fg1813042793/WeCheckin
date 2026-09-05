import type { ApiEnvelope } from '@/types/dingtalk-h5'
import { http } from 'uview-pro'
import {
  handleDingTalkH5AuthExpired,
  isApiEnvelope,
} from '@/common/dingtalk-h5-auth-expiry'
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

export interface MultipartUploadFile {
  filePath: string
  name?: string
  mimeType?: string
}

export interface UploadMultipartFilesOptions {
  fileFieldName?: string
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
    return isApiEnvelope(raw) ? raw as ApiEnvelope<T> : null
  }
  if (typeof raw !== 'string') {
    return null
  }
  try {
    const data: unknown = JSON.parse(raw)
    return isApiEnvelope(data) ? data as ApiEnvelope<T> : null
  }
  catch {
    return null
  }
}

const SAFE_IMAGE_MIME_TYPES = new Set(['image/jpeg', 'image/png', 'image/webp'])

function safeImageMimeType(file: MultipartUploadFile, blobType: string) {
  const blobMimeType = String(blobType || '').trim().toLowerCase().split(';')[0]
  if (blobMimeType) {
    if (SAFE_IMAGE_MIME_TYPES.has(blobMimeType)) {
      return blobMimeType
    }
    throw new Error('不支持的图片类型')
  }

  const fallbackMimeType = String(file.mimeType || '').trim().toLowerCase().split(';')[0]
  if (SAFE_IMAGE_MIME_TYPES.has(fallbackMimeType)) {
    return fallbackMimeType
  }
  throw new Error('不支持的图片类型')
}

function imageExtension(mimeType: string) {
  if (mimeType === 'image/jpeg')
    return '.jpg'
  if (mimeType === 'image/png')
    return '.png'
  if (mimeType === 'image/webp')
    return '.webp'
  return ''
}

function safeUploadFilename(file: MultipartUploadFile, index: number, mimeType: string) {
  const source = String(file.name || '').split(/[?#]/)[0] || ''
  const segments = source.replace(/\\/g, '/').split('/')
  let filename = segments.at(-1) || ''
  try {
    filename = decodeURIComponent(filename)
  }
  catch {
    filename = ''
  }
  filename = Array.from(filename, (character) => {
    const codePoint = character.codePointAt(0) || 0
    return codePoint <= 31 || codePoint === 127 ? '_' : character
  }).join('')
  filename = filename.replace(/[<>:"/\\|?*]/g, '_').replace(/^\.+/, '')
  filename = filename.trim()
  const extensionIndex = filename.lastIndexOf('.')
  if (extensionIndex > 0) {
    filename = filename.slice(0, extensionIndex)
  }
  const extension = imageExtension(mimeType)
  filename = filename.replace(/\.+$/, '').trim().slice(0, 120 - extension.length)
  return `${filename || `image-${index + 1}`}${extension}`
}

function isAbortError(error: unknown) {
  return Boolean(error && typeof error === 'object' && 'name' in error && error.name === 'AbortError')
}

// #ifdef H5
async function uploadMultipartFilesInH5<T>(
  url: string,
  files: readonly MultipartUploadFile[],
  options: UploadMultipartFilesOptions,
) {
  const controller = new AbortController()
  const timeout = Math.max(1, options.timeout || 30000)
  let timeoutTriggered = false
  const timeoutID = setTimeout(() => {
    timeoutTriggered = true
    controller.abort()
  }, timeout)
  try {
    const form = new FormData()
    for (const [key, value] of Object.entries(options.formData || {})) {
      form.append(key, String(value))
    }

    for (const [index, file] of files.entries()) {
      const response = await fetch(file.filePath, { signal: controller.signal })
      if (!response.ok) {
        throw new Error('读取待上传图片失败')
      }
      const blob = await response.blob()
      if (blob.size === 0) {
        throw new Error('待上传图片不能为空')
      }
      const mimeType = safeImageMimeType(file, blob.type)
      const uploadBlob = blob.type === mimeType ? blob : blob.slice(0, blob.size, mimeType)
      form.append(options.fileFieldName || 'images', uploadBlob, safeUploadFilename(file, index, mimeType))
    }

    const headers: Record<string, string> = {
      ...options.header,
      'Authorization': authToken(),
      'X-Client-Platform': DINGTALK_H5_CONFIG.CLIENT_PLATFORM,
    }
    for (const key of Object.keys(headers)) {
      if (key.toLowerCase() === 'content-type') {
        delete headers[key]
      }
    }

    const response = await fetch(buildApiUrl(url), {
      method: 'POST',
      body: form,
      headers,
      signal: controller.signal,
    })
    const responseText = await response.text()
    const data = parseUploadResponseData<T>(responseText)
    if (data && data.code !== undefined && data.code !== 0) {
      handleDingTalkH5AuthExpired(data)
    }
    if (!response.ok) {
      throw data || new Error(`上传请求失败[${response.status}]`)
    }
    if (!data) {
      throw new Error('上传响应异常')
    }
    if (data.code === undefined || data.code === 0) {
      return data
    }
    throw data
  }
  catch (error) {
    if (timeoutTriggered && isAbortError(error)) {
      throw new Error('上传超时，请检查网络后重试')
    }
    throw error
  }
  finally {
    clearTimeout(timeoutID)
  }
}
// #endif

export function uploadMultipartFiles<T>(
  url: string,
  files: readonly MultipartUploadFile[],
  options: UploadMultipartFilesOptions = {},
) {
  let task: Promise<ApiEnvelope<T>>
  // #ifdef H5
  task = uploadMultipartFilesInH5<T>(url, files, options)
  // #endif
  // #ifndef H5
  task = Promise.reject(new Error('当前上传仅支持 H5'))
  // #endif
  return task
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
