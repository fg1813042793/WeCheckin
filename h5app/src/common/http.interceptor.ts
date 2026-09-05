import type { RequestConfig, RequestInterceptor, RequestMeta, RequestOptions } from 'uview-pro'
import {
  handleDingTalkH5AuthExpired,
  isApiEnvelope,
} from '@/common/dingtalk-h5-auth-expiry'
import {
  DINGTALK_H5_BIND_REQUIRED_CODE,
  DINGTALK_H5_CONFIG,
} from '@/config/dingtalk-h5'

// 全局配置
const httpRequestConfig: RequestConfig = {
  baseUrl: DINGTALK_H5_CONFIG.BASE_URL,
  header: {
    'content-type': 'application/json',
  },
  timeout: 50000,
  meta: {
    originalData: true,
    toast: true,
    loading: false,
  },
}

// 请求/响应拦截器
const httpInterceptor: RequestInterceptor = {
  // 请求拦截器
  request: (config: RequestOptions) => {
    const meta: RequestMeta = config.meta || {}
    meta.loading && showLoading()
    config.header = config.header || {}

    if (isAuthenticatedClientRequest(config.url)) {
      config.header['X-Client-Platform'] = DINGTALK_H5_CONFIG.CLIENT_PLATFORM
      const token = String(uni.getStorageSync(DINGTALK_H5_CONFIG.TOKEN_KEY) || '')
      if (token) {
        config.header.Authorization = token
      }
    }
    return config
  },
  // 响应拦截器
  response: (response: unknown) => {
    const rawResponse = response as {
      statusCode?: number
      data?: unknown
      errMsg?: string
      config?: { meta?: RequestMeta }
    }
    const meta: RequestMeta = rawResponse.config?.meta || {}
    meta.loading && hideLoading()
    const { statusCode, data: rawData, errMsg } = rawResponse
    // 网络错误
    if (errMsg && errMsg.includes('Failed to connect')) {
      meta.toast && showToast('网络错误', 'error')
      return false
    }
    if (errMsg && errMsg.includes('request:fail')) {
      meta.toast && showToast('请求错误：未知', 'error')
      return false
    }
    // 请求错误
    if (typeof statusCode !== 'number' || statusCode < 200 || statusCode >= 300) {
      const errorMessage = `请求错误[${statusCode}]`
      meta.toast && showToast(errorMessage, 'error')
      return false
    }
    if (isApiEnvelope(rawData)) {
      if (rawData.code === undefined || rawData.code === 0) {
        return rawData
      }
      handleDingTalkH5AuthExpired(rawData)
      if (Number(rawData.code) !== DINGTALK_H5_BIND_REQUIRED_CODE) {
        meta.toast && showToast(rawData.msg || rawData.message || '请求失败', 'none')
      }
      return false
    }
    return rawData
  },
}

function isDingTalkH5Request(url = '') {
  return String(url).includes('/dingtalk/h5')
}

function isAuthenticatedClientRequest(url = '') {
  const requestUrl = String(url)
  return isDingTalkH5Request(requestUrl) || requestUrl.startsWith('/api/')
}

// 显示加载中，可以替换为uview-pro的u-loading-popup组件
function showLoading() {
  uni.showLoading({
    title: '加载中...',
    mask: true,
  })
}

// 隐藏加载中，可以替换为uview-pro的u-loading-popup组件
function hideLoading() {
  uni.hideLoading()
}

// 显示toast，可以替换为uview-pro的u-toast组件
function showToast(
  title = '',
  icon: 'success' | 'error' | 'none' = 'none',
  options: { duration: number } = { duration: 2000 },
) {
  if (title.length === 0) {
    return
  }
  uni.showToast({
    title,
    icon: title.length && title.length > 7 ? 'none' : icon,
    duration: options.duration || 2000,
  })
}

// 导出
export { httpInterceptor, httpRequestConfig }
