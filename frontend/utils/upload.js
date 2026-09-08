import CONFIG from '../config/index'
import { getAdminToken, getClientToken } from './auth'

function parseUploadResponse(response) {
  if (response.statusCode !== 200) {
    throw new Error(response.statusCode === 413 ? '上传文件过大' : `上传失败（状态 ${response.statusCode}）`)
  }

  let payload
  try {
    payload = typeof response.data === 'string' ? JSON.parse(response.data) : response.data
  } catch {
    throw new Error('上传响应格式错误')
  }
  if (!payload || payload.code !== 0 || !payload.data) {
    throw new Error(payload?.msg || '上传失败')
  }

  const data = payload.data
  const rawURL = typeof data === 'string' ? data : data.url
  if (!rawURL) throw new Error('上传响应缺少文件地址')
  const domain = typeof data === 'object' ? (data.domain || '') : ''
  const fullURL = /^https?:\/\//i.test(rawURL) ? rawURL : `${domain || CONFIG.BASE_URL}${rawURL}`
  return typeof data === 'object' ? { ...data, url: fullURL } : { url: fullURL }
}

function resolveUploadTarget(scope) {
  if (scope === 'admin') {
    return { path: '/api/v2/admin/uploads', token: getAdminToken() }
  }
  if (scope === 'client' || getClientToken()) {
    return { path: '/api/v2/uploads', token: getClientToken() }
  }
  // Anonymous surveys and exams keep the compatibility endpoint until scoped upload tokens are available.
  return { path: '/upload', token: '' }
}

export function uploadFile(filePath, options = {}) {
  const { path, token } = resolveUploadTarget(options.scope || 'auto')
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: CONFIG.BASE_URL + path,
      filePath,
      name: options.name || 'file',
      header: token ? { Authorization: token } : {},
      success: (response) => {
        try {
          resolve(parseUploadResponse(response))
        } catch (error) {
          reject(error)
        }
      },
      fail: () => reject(new Error('网络异常，上传失败'))
    })
  })
}

export { parseUploadResponse }
