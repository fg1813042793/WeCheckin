import CONFIG from '../config'

const READY_TIMEOUT = 6000

export function getDingTalkJSAPI() {
  if (typeof window === 'undefined') return null
  return window.dd || window.DD || null
}

export function isDingTalkRuntime() {
  if (typeof navigator === 'undefined') return false
  return /DingTalk/i.test(navigator.userAgent)
}

export function initDingTalkBridge(options = {}) {
  const timeout = options.timeout || READY_TIMEOUT
  const dd = getDingTalkJSAPI()

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

    dd.ready(() => {
      settle({
        ready: true,
        dd
      })
    })

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

export function requestAuthCode(corpId = CONFIG.DINGTALK_CORP_ID) {
  return new Promise((resolve, reject) => {
    const dd = getDingTalkJSAPI()
    const requestAuthCodeApi = dd?.runtime?.permission?.requestAuthCode

    if (!corpId) {
      reject(new Error('请先配置 VITE_DINGTALK_CORP_ID'))
      return
    }
    if (typeof requestAuthCodeApi !== 'function') {
      reject(new Error('当前环境未检测到钉钉免登 JSAPI'))
      return
    }

    requestAuthCodeApi({
      corpId,
      onSuccess: (result) => resolve(result?.code || ''),
      onFail: (error) => reject(error)
    })
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
