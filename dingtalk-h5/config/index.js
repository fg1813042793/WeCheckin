const env = import.meta.env || {}

const CONFIG = {
  APP_NAME: 'OA管理',
  BASE_URL: env.VITE_API_BASE_URL || '',
  DINGTALK_CORP_ID: env.VITE_DINGTALK_CORP_ID || '',
  TOKEN_KEY: 'DT_H5_TOKEN'
}

export default CONFIG
