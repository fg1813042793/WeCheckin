import { reactive, ref } from 'vue'
import { bindSelf } from '../../../api/auth/bind-account'
import {
  login as loginWithPassword,
  logout as logoutSession,
  publicConfig,
  ssoLogin
} from '../../../api/auth/login'
import { bootstrap as bootstrapSession } from '../../../api/workbench/dashboard'
import { isDingTalkRuntime, requestAuthCode, waitForDingTalkJSAPI } from '../../../utils/dingtalk'
import { authToken, clearAuthToken, isAuthExpiredError, isBindRequiredError, isPermissionDeniedError, setAuthToken } from '../../../utils/request'
import CONFIG from '../../../config'
import { defaultAppConfig, firstText, normalizeAppConfig } from '../../../views/performance/common/helpers'

export function usePerformanceAuth({
  state,
  selectedReviewId,
  resetRouteTabState,
  resetProfileDialog,
  closeCreateReviewDialog,
  ensureActiveMenu,
  clearContentLoading,
  refreshDataSafely,
  applyReviewDeepLinkIfNeeded,
  toast
}) {
  const ready = ref(false)
  const loading = ref(false)
  const sessionAccessDenied = ref(false)
  const sessionAccessDeniedMessage = ref('无权限访问，请联系管理员配置钉钉 H5 权限')
  const loginForm = reactive({ name: '', password: '' })
  const bindForm = reactive({ account: '', password: '' })
  const bindState = reactive({
    visible: false,
    bindTicket: '',
    corpId: '',
    dingTalkUserIdMasked: '',
    unionIdMasked: '',
    expiresIn: 0
  })
  const autoLoginTried = ref(false)
  const autoLoginMessage = ref('')
  const publicCorpId = ref('')
  let publicConfigLoaded = false

  async function login() {
    loading.value = true
    try {
      autoLoginMessage.value = ''
      resetBindState()
      resetRouteTabState()
      state.view = 'dashboard'
      const res = await loginWithPassword(loginForm)
      const payload = res.data || {}
      sessionAccessDenied.value = false
      setAuthToken(payload.token)
      applySessionAuthPayload(payload)
      if (!payloadHasSessionPermissions(payload)) {
        const bootstrapped = await loadBootstrapSafely()
        if (!bootstrapped) return
      }
      if (!(await applyReviewDeepLinkIfNeeded())) {
        await refreshDataSafely({ contentLoading: true })
      }
    } finally {
      loading.value = false
    }
  }

  function queryParam(name) {
    if (typeof window === 'undefined') return ''
    const searchParams = new URLSearchParams(window.location.search || '')
    const value = searchParams.get(name)
    if (value) return value
    const hash = window.location.hash || ''
    const queryIndex = hash.indexOf('?')
    if (queryIndex < 0) return ''
    return new URLSearchParams(hash.slice(queryIndex + 1)).get(name) || ''
  }

  function queryCorpId() {
    return queryParam('corpId') || queryParam('corpID') || queryParam('corp_id') || ''
  }

  function currentCorpId() {
    return queryCorpId() || publicCorpId.value || CONFIG.DINGTALK_CORP_ID || ''
  }

  async function shouldTryDingTalkAutoLogin() {
    if (isDingTalkRuntime()) return true
    if (!currentCorpId()) return false
    await waitForDingTalkJSAPI(1200)
    return isDingTalkRuntime()
  }

  async function ensurePublicConfig() {
    if (publicConfigLoaded) {
      return true
    }
    publicConfigLoaded = true
    try {
      const res = await publicConfig()
      const data = res.data || {}
      publicCorpId.value = String(data.corpId || '').trim()
      applyPublicAppConfig(data)
      return true
    } catch (error) {
      publicConfigLoaded = false
      return false
    }
  }

  function applyPublicAppConfig(payload = {}) {
    const config = payload.appConfig || payload
    state.appConfig = {
      appTitle: config.appTitle || payload.appTitle || state.appConfig.appTitle,
      appName: config.appName || payload.appName || state.appConfig.appName,
      logoText: config.logoText || payload.logoText || state.appConfig.logoText,
      logoUrl: config.logoUrl || payload.logoUrl || state.appConfig.logoUrl,
      appUrl: config.appUrl || payload.appUrl || state.appConfig.appUrl
    }
    state.appTitle = state.appConfig.appTitle || state.appTitle
  }

  async function tryDingTalkAutoLogin() {
    if (autoLoginTried.value || authToken()) return false
    autoLoginTried.value = true
    autoLoginMessage.value = ''
    await ensurePublicConfig()
    const shouldTry = await shouldTryDingTalkAutoLogin()
    const inDingTalk = isDingTalkRuntime()
    if (!shouldTry) {
      if (inDingTalk) {
        setDingTalkAutoLoginMessage(currentCorpId()
          ? '未检测到钉钉免登环境，请确认当前页面是在钉钉工作台应用内打开。'
          : '未配置钉钉企业 CorpId，请在后台“钉钉应用管理 / 配置选项”配置企业应用，或在应用地址增加 corpId 参数。')
      }
      return false
    }
    loading.value = true
    try {
      const corpId = currentCorpId()
      if (!corpId) {
        setDingTalkAutoLoginMessage('未配置钉钉企业 CorpId，请在后台“钉钉应用管理 / 配置选项”配置企业应用，或在应用地址增加 corpId 参数。')
        return false
      }
      const authCode = await requestAuthCode(corpId)
      if (!authCode) {
        setDingTalkAutoLoginMessage('钉钉免登未返回授权码，请确认当前应用地址配置在钉钉工作台微应用中。')
        return false
      }
      resetRouteTabState()
      state.view = 'dashboard'
      const res = await ssoLogin({ corpId, authCode })
      const payload = res.data || {}
      sessionAccessDenied.value = false
      setAuthToken(payload.token)
      applySessionAuthPayload(payload)
      if (!payloadHasSessionPermissions(payload)) {
        const bootstrapped = await loadBootstrapSafely()
        if (!bootstrapped) return false
      }
      if (!(await applyReviewDeepLinkIfNeeded())) {
        await refreshDataSafely({ contentLoading: true })
      }
      return true
    } catch (error) {
      if (isBindRequiredError(error)) {
        showBindRequired(error?.data || {})
        return true
      }
      if (isPermissionDeniedError(error)) {
        sessionAccessDenied.value = true
        sessionAccessDeniedMessage.value = error?.msg || '无权限访问，请联系管理员配置钉钉 H5 权限'
      } else {
        setDingTalkAutoLoginMessage(autoLoginErrorMessage(error))
      }
      return false
    } finally {
      loading.value = false
    }
  }

  async function retryDingTalkAutoLogin() {
    resetBindState()
    clearAuthToken()
    autoLoginTried.value = false
    await tryDingTalkAutoLogin()
  }

  async function bindDingTalkUser() {
    if (!bindState.bindTicket) {
      toast('绑定会话已过期，请重新打开应用')
      return
    }
    loading.value = true
    try {
      resetRouteTabState()
      state.view = 'dashboard'
      const res = await bindSelf({
        bindTicket: bindState.bindTicket,
        account: bindForm.account,
        password: bindForm.password
      })
      const payload = res.data || {}
      resetBindState()
      sessionAccessDenied.value = false
      setAuthToken(payload.token)
      applySessionAuthPayload(payload)
      if (!payloadHasSessionPermissions(payload)) {
        const bootstrapped = await loadBootstrapSafely()
        if (!bootstrapped) return
      }
      if (!(await applyReviewDeepLinkIfNeeded())) {
        await refreshDataSafely({ contentLoading: true })
      }
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await logoutSession()
    } finally {
      resetSessionState()
    }
  }

  function showBindRequired(data = {}) {
    clearAuthToken()
    autoLoginMessage.value = ''
    sessionAccessDenied.value = false
    sessionAccessDeniedMessage.value = '无权限访问，请联系管理员配置钉钉 H5 权限'
    state.user = null
    state.menus = []
    state.buttonPermissionKeys = []
    state.buttonPermissionReady = false
    state.apiPermissionKeys = []
    state.apiPermissionReady = false
    state.permissionVersion = 0
    bindState.visible = true
    bindState.bindTicket = String(data.bindTicket || '')
    bindState.corpId = String(data.corpId || '')
    bindState.dingTalkUserIdMasked = String(data.dingTalkUserIdMasked || '')
    bindState.unionIdMasked = String(data.unionIdMasked || '')
    bindState.expiresIn = Number(data.expiresIn || 0)
    bindForm.password = ''
    applySessionAuthPayload(data)
  }

  function resetBindState() {
    bindState.visible = false
    bindState.bindTicket = ''
    bindState.corpId = ''
    bindState.dingTalkUserIdMasked = ''
    bindState.unionIdMasked = ''
    bindState.expiresIn = 0
    bindForm.password = ''
  }

  function setDingTalkAutoLoginMessage(message) {
    if (!isDingTalkRuntime()) return
    autoLoginMessage.value = message
  }

  function autoLoginErrorMessage(error) {
    const rawMessage = String(error?.msg || error?.message || error?.errorMessage || error?.error || '').trim()
    if (!currentCorpId()) {
      return '未配置钉钉企业 CorpId，请先在后台“钉钉应用管理 / 配置选项”配置企业应用。'
    }
    if (rawMessage.includes('JSAPI') || rawMessage.includes('requestAuthCode') || rawMessage.includes('未检测到')) {
      return '未检测到钉钉免登 JSAPI，请确认页面是在钉钉工作台应用内打开，并且页面已加载钉钉 JSAPI。'
    }
    if (rawMessage.includes('notInDingTalk') || rawMessage.includes('不是钉钉端内')) {
      return '当前不是钉钉端内环境，请在钉钉工作台微应用中打开后使用免登。'
    }
    if (rawMessage) {
      return `钉钉免登失败：${rawMessage}`
    }
    return '钉钉免登失败，请确认 CorpId、应用凭证和钉钉工作台应用地址配置正确。'
  }

  function resetSessionState() {
    clearAuthToken()
    resetBindState()
    sessionAccessDenied.value = false
    sessionAccessDeniedMessage.value = '无权限访问，请联系管理员配置钉钉 H5 权限'
    clearContentLoading()
    state.user = null
    state.menus = []
    state.buttonPermissionKeys = []
    state.buttonPermissionReady = false
    state.apiPermissionKeys = []
    state.apiPermissionReady = false
    state.appConfig = defaultAppConfig()
    state.appTitle = ''
    state.permissionVersion = 0
    state.users = []
    state.reviews = []
    state.workbenchStats = []
    state.template = null
    state.view = 'dashboard'
    resetRouteTabState()
    selectedReviewId.value = ''
    resetProfileDialog()
    closeCreateReviewDialog()
  }

  async function loadBootstrap() {
    const res = await bootstrapSession()
    applySessionAuthPayload(res.data || {})
  }

  async function loadBootstrapSafely() {
    try {
      await loadBootstrap()
      return true
    } catch (error) {
      return handleSessionDataError(error)
    }
  }

  function handleSessionDataError(error) {
    if (isAuthExpiredError(error)) {
      resetSessionState()
      return false
    }
    if (isPermissionDeniedError(error)) {
      sessionAccessDenied.value = true
      sessionAccessDeniedMessage.value = error?.msg || '无权限访问，请联系管理员配置钉钉 H5 权限'
    }
    return false
  }

  function applySessionAuthPayload(payload = {}) {
    const user = payload.user || payload.userInfo
    if (user) {
      state.user = user
    }
    if (Array.isArray(payload.menus)) {
      state.menus = payload.menus
    }
    if (Array.isArray(payload.apiPermissionKeys)) {
      state.apiPermissionKeys = payload.apiPermissionKeys
    }
    if (Array.isArray(payload.buttonPermissionKeys)) {
      state.buttonPermissionKeys = payload.buttonPermissionKeys
    }
    if (Object.prototype.hasOwnProperty.call(payload, 'apiPermissionReady')) {
      state.apiPermissionReady = Boolean(payload.apiPermissionReady)
    }
    if (Object.prototype.hasOwnProperty.call(payload, 'buttonPermissionReady')) {
      state.buttonPermissionReady = Boolean(payload.buttonPermissionReady)
    }
    if (Object.prototype.hasOwnProperty.call(payload, 'permissionVersion')) {
      state.permissionVersion = Number(payload.permissionVersion || 0)
    }
    const payloadConfig = payload.appConfig && typeof payload.appConfig === 'object' ? payload.appConfig : {}
    const nextAppConfig = normalizeAppConfig({
      appTitle: firstText(payloadConfig.appTitle, payloadConfig.appName, payload.appTitle, payload.appName, payload.applicationName, state.appTitle),
      appName: firstText(payloadConfig.appName, payload.appName, payloadConfig.appTitle, payload.appTitle, state.appConfig.appName),
      logoText: firstText(payloadConfig.logoText, payload.logoText, state.appConfig.logoText),
      logoUrl: firstText(payloadConfig.logoUrl, payload.logoUrl, payload.logoURL, state.appConfig.logoUrl),
      appUrl: firstText(payloadConfig.appUrl, payload.appUrl, state.appConfig.appUrl)
    })
    if (nextAppConfig.appTitle) {
      state.appConfig = nextAppConfig
      state.appTitle = nextAppConfig.appTitle
    }
    ensureActiveMenu()
  }

  function payloadHasSessionPermissions(payload = {}) {
    return Array.isArray(payload.menus) ||
      Array.isArray(payload.apiPermissionKeys) ||
      Array.isArray(payload.buttonPermissionKeys) ||
      Object.prototype.hasOwnProperty.call(payload, 'permissionVersion')
  }

  return {
    autoLoginMessage,
    bindDingTalkUser,
    bindForm,
    bindState,
    handleSessionDataError,
    loadBootstrapSafely,
    loading,
    login,
    loginForm,
    logout,
    ready,
    resetSessionState,
    retryDingTalkAutoLogin,
    sessionAccessDenied,
    sessionAccessDeniedMessage,
    tryDingTalkAutoLogin
  }
}
