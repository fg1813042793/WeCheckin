import CONFIG from '../config'

const READY_TIMEOUT = 6000
const BRIDGE_POLL_INTERVAL = 100
const DINGTALK_UA_PATTERN = /(DingTalk|AliApp\(DingTalk|DingTalkDesktop)/i
const DINGTALK_JSAPI_URL = 'https://g.alicdn.com/dingding/dingtalk-jsapi/3.0.25/dingtalk.open.js'
let dingTalkScriptPromise = null

export function getDingTalkJSAPI() {
  if (typeof window === 'undefined') return null
  return window.dd || window.DD || null
}

function isDingTalkUserAgent() {
  if (typeof navigator === 'undefined') return false
  return DINGTALK_UA_PATTERN.test(navigator.userAgent || '')
}

export function isDingTalkRuntime() {
  if (isDingTalkUserAgent()) return true

  const dd = getDingTalkJSAPI()
  const platform = String(dd?.env?.platform || '').trim().toLowerCase()
  return Boolean(platform && platform !== 'notindingtalk')
}

function loadDingTalkScript() {
  const dd = getDingTalkJSAPI()
  if (dd) return Promise.resolve(dd)
  if (typeof document === 'undefined' || !isDingTalkUserAgent()) return Promise.resolve(null)
  if (dingTalkScriptPromise) return dingTalkScriptPromise

  dingTalkScriptPromise = new Promise((resolve) => {
    const existing = document.querySelector(`script[src="${DINGTALK_JSAPI_URL}"]`)
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
  if (dd) return Promise.resolve(dd)
  if (!isDingTalkUserAgent()) return Promise.resolve(null)

  await loadDingTalkScript()
  const loaded = getDingTalkJSAPI()
  if (loaded) return Promise.resolve(loaded)

  return new Promise((resolve) => {
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

export async function initDingTalkBridge(options = {}) {
  const timeout = options.timeout || READY_TIMEOUT
  const dd = await waitForDingTalkJSAPI(timeout)

  if (!dd) {
    return Promise.resolve({
      ready: false,
      reason: 'jsapi-missing'
    })
  }

  if (typeof dd.ready !== 'function') {
    return Promise.resolve({
      ready: true,
      dd
    })
  }

  return new Promise((resolve) => {
    let settled = false
    const settle = (payload) => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      resolve(payload)
    }
    const timer = setTimeout(() => {
      settle({
        ready: false,
        reason: 'ready-timeout'
      })
    }, timeout)

    try {
      dd.ready(() => {
        settle({
          ready: true,
          dd
        })
      })
    } catch (error) {
      settle({
        ready: false,
        reason: 'ready-error',
        error
      })
      return
    }

    if (typeof dd.error === 'function') {
      dd.error((error) => {
        settle({
          ready: false,
          reason: 'jsapi-error',
          error
        })
      })
    }
  })
}

export async function requestAuthCode(corpId = CONFIG.DINGTALK_CORP_ID) {
  if (!corpId) {
    throw new Error('请先配置 VITE_DINGTALK_CORP_ID')
  }
  if (!isDingTalkRuntime()) {
    throw new Error('当前环境不是钉钉端内环境')
  }

  const bridge = await initDingTalkBridge()
  if (!bridge.ready) {
    throw new Error('当前环境未检测到钉钉免登 JSAPI')
  }

  return new Promise((resolve, reject) => {
    const dd = bridge.dd
    const requestAuthCodeApi = dd?.runtime?.permission?.requestAuthCode
    const getAuthCodeApi = dd?.getAuthCode
    let settled = false
    let timer = null
    const settleResolve = (result) => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      resolve(result?.code || result?.authCode || result?.auth_code || '')
    }
    const settleReject = (error) => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      reject(error)
    }
    timer = setTimeout(() => {
      settleReject(new Error('钉钉免登授权码获取超时'))
    }, READY_TIMEOUT)

    if (typeof requestAuthCodeApi !== 'function' && typeof getAuthCodeApi !== 'function') {
      settleReject(new Error('当前环境未检测到钉钉免登 JSAPI'))
      return
    }

    const payload = {
      corpId,
      onSuccess: settleResolve,
      onFail: settleReject,
      success: settleResolve,
      fail: settleReject
    }

    let result
    try {
      result = typeof requestAuthCodeApi === 'function'
        ? requestAuthCodeApi.call(dd.runtime.permission, payload)
        : getAuthCodeApi.call(dd, payload)
    } catch (error) {
      settleReject(error)
      return
    }
    if (result && typeof result.then === 'function') {
      result
        .then(settleResolve)
        .catch(settleReject)
    }
  })
}

export function setNavigationTitle(title) {
  const dd = getDingTalkJSAPI()
  const setTitleApi = dd?.biz?.navigation?.setTitle

  if (typeof setTitleApi === 'function') {
    setTitleApi({
      title
    })
    return
  }

  uni.setNavigationBarTitle({
    title
  })
}
