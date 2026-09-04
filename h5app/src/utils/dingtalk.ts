import { DINGTALK_H5_CONFIG } from '@/config/dingtalk-h5'

interface DingTalkBridgeResult {
  ready: boolean
  dd?: DingTalkJSAPI
  reason?: string
  error?: unknown
}

interface DingTalkJSAPI {
  env?: {
    platform?: string
  }
  ready?: (callback: () => void) => void
  error?: (callback: (error: unknown) => void) => void
  getAuthCode?: (payload: AuthCodePayload) => Promise<AuthCodeResult> | void
  runtime?: {
    permission?: {
      requestAuthCode?: (payload: AuthCodePayload) => Promise<AuthCodeResult> | void
    }
  }
  biz?: {
    navigation?: {
      setTitle?: (payload: { title: string }) => void
    }
  }
}

interface AuthCodeResult {
  code?: string
  authCode?: string
  auth_code?: string
}

interface AuthCodePayload {
  corpId: string
  onSuccess: (result: AuthCodeResult) => void
  onFail: (error: unknown) => void
  success: (result: AuthCodeResult) => void
  fail: (error: unknown) => void
}

declare global {
  interface Window {
    dd?: DingTalkJSAPI
    DD?: DingTalkJSAPI
  }
}

const READY_TIMEOUT = 6000
const BRIDGE_POLL_INTERVAL = 100
const DINGTALK_UA_PATTERN = /DingTalk/i
const DINGTALK_JSAPI_URL = 'https://g.alicdn.com/dingding/dingtalk-jsapi/3.0.25/dingtalk.open.js'

let dingTalkScriptPromise: Promise<DingTalkJSAPI | null> | null = null

export function getDingTalkJSAPI() {
  if (typeof window === 'undefined') {
    return null
  }
  return window.dd || window.DD || null
}

function isDingTalkUserAgent() {
  if (typeof navigator === 'undefined') {
    return false
  }
  return DINGTALK_UA_PATTERN.test(navigator.userAgent || '')
}

export function isDingTalkRuntime() {
  if (isDingTalkUserAgent()) {
    return true
  }

  const dd = getDingTalkJSAPI()
  const platform = String(dd?.env?.platform || '').trim().toLowerCase()
  return Boolean(platform && platform !== 'notindingtalk')
}

function loadDingTalkScript() {
  const dd = getDingTalkJSAPI()
  if (dd) {
    return Promise.resolve(dd)
  }
  if (typeof document === 'undefined' || !isDingTalkUserAgent()) {
    return Promise.resolve(null)
  }
  if (dingTalkScriptPromise) {
    return dingTalkScriptPromise
  }

  dingTalkScriptPromise = new Promise((resolve) => {
    const existing = document.querySelector<HTMLScriptElement>(`script[src="${DINGTALK_JSAPI_URL}"]`)
    const script = existing || document.createElement('script')
    script.src = DINGTALK_JSAPI_URL
    script.async = true
    script.onload = () => resolve(getDingTalkJSAPI())
    script.onerror = () => resolve(null)
    if (!existing) {
      document.head.appendChild(script)
    }
  })

  return dingTalkScriptPromise
}

export async function waitForDingTalkJSAPI(timeout = READY_TIMEOUT) {
  const dd = getDingTalkJSAPI()
  if (dd) {
    return dd
  }
  if (!isDingTalkUserAgent()) {
    return null
  }

  await loadDingTalkScript()
  const loaded = getDingTalkJSAPI()
  if (loaded) {
    return loaded
  }

  return new Promise<DingTalkJSAPI | null>((resolve) => {
    const start = Date.now()
    const timer = setInterval(() => {
      const current = getDingTalkJSAPI()
      if (current) {
        clearInterval(timer)
        resolve(current)
        return
      }
      if (Date.now() - start >= timeout) {
        clearInterval(timer)
        resolve(null)
      }
    }, BRIDGE_POLL_INTERVAL)
  })
}

export async function initDingTalkBridge(options: { timeout?: number } = {}) {
  const timeout = options.timeout || READY_TIMEOUT
  const dd = await waitForDingTalkJSAPI(timeout)

  if (!dd) {
    return {
      ready: false,
      reason: 'jsapi-missing',
    } satisfies DingTalkBridgeResult
  }

  if (typeof dd.ready !== 'function') {
    return {
      ready: true,
      dd,
    } satisfies DingTalkBridgeResult
  }

  return new Promise<DingTalkBridgeResult>((resolve) => {
    let settled = false
    const timer = setTimeout(() => {
      settle({
        ready: false,
        reason: 'ready-timeout',
      })
    }, timeout)

    function settle(payload: DingTalkBridgeResult) {
      if (settled) {
        return
      }
      settled = true
      clearTimeout(timer)
      resolve(payload)
    }

    try {
      dd.ready?.(() => {
        settle({
          ready: true,
          dd,
        })
      })
    }
    catch (error) {
      settle({
        ready: false,
        reason: 'ready-error',
        error,
      })
      return
    }

    dd.error?.((error) => {
      settle({
        ready: false,
        reason: 'jsapi-error',
        error,
      })
    })
  })
}

export async function requestAuthCode(corpId = DINGTALK_H5_CONFIG.DINGTALK_CORP_ID) {
  if (!corpId) {
    throw new Error('请先配置钉钉企业 CorpId')
  }
  if (!isDingTalkRuntime()) {
    throw new Error('当前环境不是钉钉端内环境')
  }

  const bridge = await initDingTalkBridge()
  if (!bridge.ready || !bridge.dd) {
    throw new Error('当前环境未检测到钉钉免登 JSAPI')
  }

  return new Promise<string>((resolve, reject) => {
    const dd = bridge.dd
    const requestAuthCodeApi = dd?.runtime?.permission?.requestAuthCode
    const getAuthCodeApi = dd?.getAuthCode
    let settled = false
    const timer = setTimeout(() => {
      settleReject(new Error('钉钉免登授权码获取超时'))
    }, READY_TIMEOUT)

    function settleResolve(result: AuthCodeResult = {}) {
      if (settled) {
        return
      }
      settled = true
      clearTimeout(timer)
      resolve(result.code || result.authCode || result.auth_code || '')
    }

    function settleReject(error: unknown) {
      if (settled) {
        return
      }
      settled = true
      clearTimeout(timer)
      reject(error)
    }

    if (typeof requestAuthCodeApi !== 'function' && typeof getAuthCodeApi !== 'function') {
      settleReject(new Error('当前环境未检测到钉钉免登 JSAPI'))
      return
    }

    const payload: AuthCodePayload = {
      corpId,
      onSuccess: settleResolve,
      onFail: settleReject,
      success: settleResolve,
      fail: settleReject,
    }

    try {
      const result = typeof requestAuthCodeApi === 'function'
        ? requestAuthCodeApi.call(dd?.runtime?.permission, payload)
        : getAuthCodeApi?.call(dd, payload)
      if (result && typeof result.then === 'function') {
        result.then(settleResolve).catch(settleReject)
      }
    }
    catch (error) {
      settleReject(error)
    }
  })
}

export function setNavigationTitle(title: string) {
  const dd = getDingTalkJSAPI()
  const setTitleApi = dd?.biz?.navigation?.setTitle

  if (typeof setTitleApi === 'function') {
    setTitleApi({ title })
    return
  }

  uni.setNavigationBarTitle({ title })
}
