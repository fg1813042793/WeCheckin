const env = import.meta.env || {}

function readEnv(key: string) {
  return String(env[key] || '').trim()
}

export const DINGTALK_H5_CONFIG = {
  APP_NAME: 'OA管理',
  BASE_URL: readEnv('VITE_API_BASE_URL'),
  CLIENT_PLATFORM: 'dingtalk-h5',
  DINGTALK_CORP_ID: readEnv('VITE_DINGTALK_CORP_ID'),
  TOKEN_KEY: 'DT_H5_TOKEN',
} as const

export const DINGTALK_H5_AUTH_EXPIRED_EVENT = 'dingtalk-h5-auth-expired'
export const DINGTALK_H5_BIND_REQUIRED_CODE = 10020
