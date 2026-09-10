export const DINGTALK_OAUTH_STATE_KEY = 'dingtalkOAuthAttempt'
export const DINGTALK_OAUTH_CONSUMED_STATE_KEY = 'dingtalkOAuthConsumedState'

const OAUTH_ATTEMPT_TTL = 10 * 60 * 1000

function configuredRedirectURI() {
  return String(import.meta.env.VITE_DINGTALK_OAUTH_REDIRECT_URI || '').trim()
}

export function getDingTalkOAuthRedirectURI() {
  const configured = configuredRedirectURI()
  if (configured) return configured
  // #ifdef H5
  return `${window.location.origin}${window.location.pathname}`
  // #endif
  return ''
}

export function getDingTalkNativeRedirectURI() {
  const configured = String(import.meta.env.VITE_DINGTALK_NATIVE_REDIRECT_URI || '').trim()
  if (configured) return webRedirectURI(configured)
  return webRedirectURI(String(import.meta.env.VITE_API_BASE_URL || '').trim())
}

function webRedirectURI(value) {
  const target = String(value || '').trim()
  const match = target.match(/^https?:\/\/([^/?#\s]+)(?:[/?#]|$)/i)
  if (!match || !match[1] || match[1].includes('@')) return ''
  return target
}

export function beginDingTalkOAuth({ authorizationUrl, state, corpId }) {
  if (!authorizationUrl || !state || !corpId) {
    throw new Error('钉钉登录参数不完整')
  }
  prepareDingTalkOAuthAttempt({ state, corpId })

  // #ifdef H5
  window.location.assign(authorizationUrl)
  return
  // #endif

  uni.removeStorageSync(DINGTALK_OAUTH_STATE_KEY)
  throw new Error('当前平台暂不支持网页钉钉登录')
}

export function prepareDingTalkOAuthAttempt({ state, corpId }) {
  if (!state || !corpId) {
    throw new Error('钉钉登录参数不完整')
  }
  uni.setStorageSync(DINGTALK_OAUTH_STATE_KEY, {
    state,
    corpId,
    createdAt: Date.now()
  })
}

export function readDingTalkOAuthCallback(pageOptions = {}) {
  const callback = callbackParams(pageOptions)
  if (!callback.authCode && !callback.error) return null

  return validateDingTalkOAuthResult(callback)
}

export function validateDingTalkOAuthResult(callback = {}) {
  if (!callback.authCode && !callback.error) return null

  const consumed = uni.getStorageSync(DINGTALK_OAUTH_CONSUMED_STATE_KEY)
  if (
    callback.state &&
    consumed &&
    consumed.state === callback.state &&
    Date.now() - Number(consumed.consumedAt || 0) <= OAUTH_ATTEMPT_TTL
  ) {
    return null
  }

  const attempt = uni.getStorageSync(DINGTALK_OAUTH_STATE_KEY)
  if (!attempt || !attempt.state || !attempt.corpId) {
    throw new Error('钉钉登录会话已失效，请重新发起登录')
  }
  if (!callback.state || callback.state !== attempt.state) {
    throw new Error('钉钉登录状态校验失败，请重新发起登录')
  }
  uni.setStorageSync(DINGTALK_OAUTH_CONSUMED_STATE_KEY, {
    state: callback.state,
    consumedAt: Date.now()
  })
  if (Date.now() - Number(attempt.createdAt || 0) > OAUTH_ATTEMPT_TTL) {
    throw new Error('钉钉登录会话已过期，请重新发起登录')
  }
  if (callback.error) {
    throw new Error(callback.errorDescription || '钉钉授权未完成')
  }
  return {
    corpId: attempt.corpId,
    authCode: callback.authCode
  }
}

export function clearDingTalkOAuthAttempt() {
  uni.removeStorageSync(DINGTALK_OAUTH_STATE_KEY)

  // #ifdef H5
  const cleanURL = `${window.location.pathname}${window.location.hash || ''}`
  window.history.replaceState({}, document.title, cleanURL)
  // #endif
}

function callbackParams(pageOptions) {
  const params = {}
  appendParams(params, pageOptions)

  // #ifdef H5
  appendParams(params, parseDingTalkOAuthQuery(window.location.search))
  // #endif

  return {
    authCode: firstParam(params, ['authCode', 'code']),
    state: firstParam(params, ['state']),
    error: firstParam(params, ['error']),
    errorDescription: firstParam(params, ['error_description', 'errorDescription'])
  }
}

function appendParams(target, source) {
  if (!source) return
  Object.keys(source).forEach((key) => {
    const value = source[key]
    if (!(key in target) && value !== undefined && value !== null) {
      target[key] = String(value)
    }
  })
}

function firstParam(params, keys) {
  for (const key of keys) {
    const value = String(params[key] || '').trim()
    if (value) return value
  }
  return ''
}

export function parseDingTalkOAuthQuery(rawQuery) {
  const result = {}
  let query = String(rawQuery || '').trim()
  const queryIndex = query.indexOf('?')
  if (queryIndex >= 0) query = query.slice(queryIndex + 1)
  if (query.startsWith('?')) query = query.slice(1)
  const hashIndex = query.indexOf('#')
  if (hashIndex >= 0) query = query.slice(0, hashIndex)
  if (!query) return result

  query.split('&').forEach((part) => {
    if (!part) return
    const separatorIndex = part.indexOf('=')
    const rawKey = separatorIndex >= 0 ? part.slice(0, separatorIndex) : part
    const rawValue = separatorIndex >= 0 ? part.slice(separatorIndex + 1) : ''
    const key = decodeQueryPart(rawKey)
    if (!key || key in result) return
    result[key] = decodeQueryPart(rawValue)
  })
  return result
}

function decodeQueryPart(value) {
  try {
    return decodeURIComponent(String(value || '').replace(/\+/g, ' '))
  } catch (e) {
    return String(value || '')
  }
}
