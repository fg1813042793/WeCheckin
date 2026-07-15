const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083'

export default {
  BASE_URL: API_BASE_URL,
  VER: 'build 2026.05.28',
  COMPANY: 'MY打卡',

  IS_DEMO: false,
  MOBILE_CHECK: false,

  IMG_UPLOAD_SIZE: 20,

  CACHE_IS_LIST: true,
  CACHE_LIST_TIME: 60 * 30
}
