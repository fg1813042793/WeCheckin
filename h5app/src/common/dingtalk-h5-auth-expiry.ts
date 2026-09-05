import type { ApiEnvelope } from '@/types/dingtalk-h5'
import {
  DINGTALK_H5_AUTH_EXPIRED_EVENT,
  DINGTALK_H5_CONFIG,
} from '@/config/dingtalk-h5'

const LOGIN_EXPIRED_MESSAGES = new Set([
  '未登录',
  '登录已过期',
  '登录已过期或已被强制下线',
  '账号异常',
])

export function isApiEnvelope(value: unknown): value is ApiEnvelope {
  return Boolean(value && typeof value === 'object' && (
    'code' in value
    || 'msg' in value
    || 'message' in value
    || 'data' in value
  ))
}

export function isDingTalkH5AuthExpired(payload: ApiEnvelope) {
  const message = String(payload.msg || payload.message || '')
  return LOGIN_EXPIRED_MESSAGES.has(message)
}

export function handleDingTalkH5AuthExpired(payload: ApiEnvelope) {
  if (!isDingTalkH5AuthExpired(payload)) {
    return false
  }
  uni.removeStorageSync(DINGTALK_H5_CONFIG.TOKEN_KEY)
  uni.$emit(DINGTALK_H5_AUTH_EXPIRED_EVENT, payload)
  return true
}
