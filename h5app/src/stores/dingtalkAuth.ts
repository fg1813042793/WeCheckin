import type {
  ApiEnvelope,
  AppConfig,
  AuthSessionPayload,
  BindRequiredPayload,
  DingTalkMenu,
  DingTalkUser,
  LoginRequest,
  PublicConfigPayload,
} from '@/types/dingtalk-h5'
import { defineStore } from 'pinia'
import { computed, reactive, ref } from 'vue'
import {
  authToken,
  bindSelf,
  bootstrap,
  clearAuthToken,
  captcha as fetchLoginCaptcha,
  login,
  logout,
  publicConfig,
  setAuthToken,
  ssoLogin,
} from '@/api/dingtalk-h5'
import { createAppNavItems, flattenAppNav, normalizeAppMenus } from '@/config/app-navigation'
import { DINGTALK_H5_BIND_REQUIRED_CODE, DINGTALK_H5_CONFIG } from '@/config/dingtalk-h5'
import { isDingTalkRuntime, requestAuthCode, setBrowserFavicon, setNavigationTitle, waitForDingTalkJSAPI } from '@/utils/dingtalk'
import { useAppContentStore } from './appContent'
import { useAppShellStore } from './appShell'

function defaultAppConfig(): AppConfig {
  return {
    appTitle: 'OA管理',
    appName: 'OA管理',
    logoText: 'OA',
    logoUrl: '',
    appUrl: '',
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === 'object')
}

function firstText(...values: Array<unknown>) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

function resolveErrorEnvelope(error: unknown): ApiEnvelope {
  const candidate = error as { data?: unknown, code?: unknown, msg?: unknown, message?: unknown }
  if (isApiEnvelope(candidate?.data)) {
    return candidate.data
  }
  if (isApiEnvelope(candidate)) {
    return candidate
  }
  return {
    msg: firstText(candidate?.msg, candidate?.message) || '请求失败',
  }
}

function isApiEnvelope(value: unknown): value is ApiEnvelope {
  return Boolean(value && typeof value === 'object' && (
    'code' in value
    || 'msg' in value
    || 'message' in value
    || 'data' in value
  ))
}

function hasSessionFields(value: unknown) {
  return isRecord(value) && (
    'token' in value
    || 'user' in value
    || 'userInfo' in value
    || 'menus' in value
    || 'apiPermissionKeys' in value
    || 'buttonPermissionKeys' in value
    || 'appConfig' in value
  )
}

function resolveSessionPayload(source: unknown): AuthSessionPayload {
  if (!isRecord(source)) {
    return {}
  }
  if (hasSessionFields(source)) {
    return source as AuthSessionPayload
  }

  const nested = source.data
  if (isApiEnvelope(source) && hasSessionFields(nested)) {
    return nested as AuthSessionPayload
  }
  if (isApiEnvelope(nested) && hasSessionFields(nested.data)) {
    return nested.data as AuthSessionPayload
  }
  if (hasSessionFields(nested)) {
    return nested as AuthSessionPayload
  }
  return source as AuthSessionPayload
}

export const useDingtalkAuthStore = defineStore('dingtalkAuth', () => {
  const ready = ref(false)
  const loading = ref(false)
  const user = ref<DingTalkUser | null>(null)
  const menus = ref<DingTalkMenu[]>([])
  const buttonPermissionKeys = ref<string[]>([])
  const apiPermissionKeys = ref<string[]>([])
  const buttonPermissionReady = ref(false)
  const apiPermissionReady = ref(false)
  const permissionVersion = ref(0)
  const appConfig = ref<AppConfig>(defaultAppConfig())
  const publicCorpId = ref('')
  const autoLoginTried = ref(false)
  const autoLoginMessage = ref('')
  const canRetryDingTalkLogin = ref(isDingTalkRuntime())
  const captchaEnabled = ref(false)
  const captchaId = ref('')
  const captchaImage = ref('')
  const captchaLoading = ref(false)
  const captchaUnavailable = ref(false)
  const sessionAccessDenied = ref(false)
  const sessionAccessDeniedMessage = ref('无权限访问，请联系管理员配置钉钉 H5 权限')
  const passwordChangeRequired = ref(false)
  const passwordExpired = ref(false)
  const passwordExpiresAt = ref(0)
  const passwordExpiryWarning = ref(false)
  const bindState = reactive({
    visible: false,
    bindTicket: '',
    corpId: '',
    dingTalkUserIdMasked: '',
    unionIdMasked: '',
    expiresIn: 0,
  })

  let publicConfigLoaded = false
  let publicConfigRequest: Promise<boolean> | null = null
  let contextCorpId: string | null = null
  let authRevision = 0
  let passwordExpiryWarningShown = false
  let initRequest: { corpId: string, promise: Promise<void> } | null = null

  const isLoggedIn = computed(() => Boolean(user.value))
  const normalizedMenus = computed(() => normalizeAppMenus(menus.value))
  const flatMenus = computed(() => flattenAppNav(normalizedMenus.value))
  const navItems = computed(() => createAppNavItems(normalizedMenus.value))

  const appTitle = computed(() => firstText(appConfig.value.appTitle, appConfig.value.appName, DINGTALK_H5_CONFIG.APP_NAME))

  function hasButtonPermission(key: string) {
    const permissionKey = String(key || '').trim()
    return Boolean(permissionKey && buttonPermissionReady.value && buttonPermissionKeys.value.includes(permissionKey))
  }

  function hasApiPermission(key: string) {
    const permissionKey = String(key || '').trim()
    return Boolean(permissionKey && apiPermissionReady.value && apiPermissionKeys.value.includes(permissionKey))
  }

  function hasMenuPermission(key: string) {
    return flatMenus.value.some(item => item.key === key)
  }

  function queryParam(name: string) {
    if (typeof window === 'undefined') {
      return ''
    }
    const searchParams = new URLSearchParams(window.location.search || '')
    const value = searchParams.get(name)
    if (value) {
      return value
    }
    const hash = window.location.hash || ''
    const queryIndex = hash.indexOf('?')
    if (queryIndex < 0) {
      return ''
    }
    return new URLSearchParams(hash.slice(queryIndex + 1)).get(name) || ''
  }

  function currentCorpId() {
    return firstText(queryParam('corpId'), queryParam('corpID'), queryParam('corp_id'), publicCorpId.value, DINGTALK_H5_CONFIG.DINGTALK_CORP_ID)
  }

  function applyAppConfig(payload: Partial<AppConfig> = {}) {
    appConfig.value = {
      appTitle: firstText(payload.appTitle, payload.appName, appConfig.value.appTitle, DINGTALK_H5_CONFIG.APP_NAME),
      appName: firstText(payload.appName, payload.appTitle, appConfig.value.appName, DINGTALK_H5_CONFIG.APP_NAME),
      logoText: firstText(payload.logoText, appConfig.value.logoText, 'OA'),
      logoUrl: firstText(payload.logoUrl, appConfig.value.logoUrl),
      appUrl: firstText(payload.appUrl, appConfig.value.appUrl),
    }
    setNavigationTitle(appConfig.value.appTitle)
    setBrowserFavicon(appConfig.value.logoUrl)
  }

  function applyPublicConfig(payload: PublicConfigPayload = {}) {
    const nested = typeof payload.appConfig === 'object' && payload.appConfig ? payload.appConfig as Partial<AppConfig> : {}
    publicCorpId.value = firstText(payload.corpId)
    applyAppConfig({
      appTitle: firstText(nested.appTitle, payload.appTitle),
      appName: firstText(nested.appName, payload.appName),
      logoText: firstText(nested.logoText, payload.logoText),
      logoUrl: firstText(nested.logoUrl, payload.logoUrl),
      appUrl: firstText(nested.appUrl, payload.appUrl),
    })
  }

  function applySessionPayload(payload: unknown = {}) {
    const sessionPayload = resolveSessionPayload(payload)
    const nextUser = sessionPayload.user || sessionPayload.userInfo
    if (nextUser) {
      user.value = nextUser
    }
    if (Array.isArray(sessionPayload.menus)) {
      menus.value = sessionPayload.menus
    }
    const nextApiPermissionKeys = Array.isArray(sessionPayload.apiPermissionKeys)
      ? sessionPayload.apiPermissionKeys
      : null
    const nextButtonPermissionKeys = Array.isArray(sessionPayload.buttonPermissionKeys)
      ? sessionPayload.buttonPermissionKeys
      : null
    if (nextApiPermissionKeys) {
      apiPermissionKeys.value = nextApiPermissionKeys
    }
    if (nextButtonPermissionKeys) {
      buttonPermissionKeys.value = nextButtonPermissionKeys
    }
    if (nextApiPermissionKeys && !Object.prototype.hasOwnProperty.call(sessionPayload, 'apiPermissionReady')) {
      apiPermissionReady.value = true
    }
    if (nextButtonPermissionKeys && !Object.prototype.hasOwnProperty.call(sessionPayload, 'buttonPermissionReady')) {
      buttonPermissionReady.value = true
    }
    if (Object.prototype.hasOwnProperty.call(sessionPayload, 'apiPermissionReady')) {
      apiPermissionReady.value = Boolean(sessionPayload.apiPermissionReady)
    }
    if (Object.prototype.hasOwnProperty.call(sessionPayload, 'buttonPermissionReady')) {
      buttonPermissionReady.value = Boolean(sessionPayload.buttonPermissionReady)
    }
    if (Object.prototype.hasOwnProperty.call(sessionPayload, 'permissionVersion')) {
      permissionVersion.value = Number(sessionPayload.permissionVersion || 0)
    }
    if (Object.prototype.hasOwnProperty.call(sessionPayload, 'passwordChangeRequired')) {
      passwordChangeRequired.value = sessionPayload.passwordChangeRequired === true
      passwordExpired.value = sessionPayload.passwordExpired === true
      passwordExpiresAt.value = Number(sessionPayload.passwordExpiresAt || 0)
      passwordExpiryWarning.value = sessionPayload.passwordExpiryWarning === true
      if (passwordExpiryWarning.value && !passwordChangeRequired.value && !passwordExpiryWarningShown) {
        passwordExpiryWarningShown = true
        uni.showToast({ title: '登录密码即将到期，请及时修改', icon: 'none' })
      }
    }
    const payloadConfig = sessionPayload.appConfig || {}
    applyAppConfig({
      appTitle: firstText(payloadConfig.appTitle, payloadConfig.appName, sessionPayload.appTitle, sessionPayload.appName, sessionPayload.applicationName),
      appName: firstText(payloadConfig.appName, sessionPayload.appName, payloadConfig.appTitle, sessionPayload.appTitle),
      logoText: firstText(payloadConfig.logoText, sessionPayload.logoText),
      logoUrl: firstText(payloadConfig.logoUrl, sessionPayload.logoUrl, sessionPayload.logoURL),
      appUrl: firstText(payloadConfig.appUrl, sessionPayload.appUrl),
    })
  }

  function updateCurrentUser(payload: Partial<DingTalkUser> = {}) {
    if (!user.value || !isRecord(payload)) {
      return
    }
    const previousUser = user.value
    const nextId = firstText(payload.id, payload.account, previousUser.id)
    const nextAvatar = firstText(payload.avatar, payload.avatarUrl, payload.pic, payload.userPic, previousUser.avatar, previousUser.avatarUrl)
    user.value = {
      ...previousUser,
      ...payload,
      id: nextId,
      account: firstText(payload.account, nextId),
      avatar: nextAvatar,
    }
  }

  function ensurePublicConfig(): Promise<boolean> {
    if (publicConfigLoaded) {
      return Promise.resolve(true)
    }
    if (publicConfigRequest)
      return publicConfigRequest
    publicConfigRequest = publicConfig().then((res) => {
      applyPublicConfig(res.data || {})
      publicConfigLoaded = true
      return true
    }).catch(() => false).finally(() => {
      publicConfigRequest = null
    })
    return publicConfigRequest
  }

  function resetBindState() {
    bindState.visible = false
    bindState.bindTicket = ''
    bindState.corpId = ''
    bindState.dingTalkUserIdMasked = ''
    bindState.unionIdMasked = ''
    bindState.expiresIn = 0
  }

  async function refreshLoginCaptcha() {
    const revision = authRevision
    captchaLoading.value = true
    captchaUnavailable.value = false
    captchaId.value = ''
    captchaImage.value = ''
    try {
      const res = await fetchLoginCaptcha()
      if (revision !== authRevision)
        return false
      const payload = res.data || { enabled: false }
      captchaEnabled.value = payload.enabled === true
      if (!captchaEnabled.value) {
        return true
      }
      captchaId.value = String(payload.captchaId || '')
      captchaImage.value = String(payload.imageData || '')
      if (!captchaId.value || !captchaImage.value) {
        captchaUnavailable.value = true
        return false
      }
      return true
    }
    catch {
      if (revision !== authRevision)
        return false
      captchaEnabled.value = true
      captchaUnavailable.value = true
      return false
    }
    finally {
      if (revision === authRevision)
        captchaLoading.value = false
    }
  }

  function showBindRequired(payload: BindRequiredPayload = {}) {
    clearAuthToken()
    user.value = null
    menus.value = []
    bindState.visible = true
    bindState.bindTicket = String(payload.bindTicket || '')
    bindState.corpId = String(payload.corpId || '')
    bindState.dingTalkUserIdMasked = String(payload.dingTalkUserIdMasked || '')
    bindState.unionIdMasked = String(payload.unionIdMasked || '')
    bindState.expiresIn = Number(payload.expiresIn || 0)
    applySessionPayload(payload)
  }

  function clearSessionState() {
    clearAuthToken()
    resetBindState()
    autoLoginTried.value = false
    autoLoginMessage.value = ''
    loading.value = false
    user.value = null
    menus.value = []
    buttonPermissionKeys.value = []
    apiPermissionKeys.value = []
    buttonPermissionReady.value = false
    apiPermissionReady.value = false
    permissionVersion.value = 0
    sessionAccessDenied.value = false
    sessionAccessDeniedMessage.value = '无权限访问，请联系管理员配置钉钉 H5 权限'
    passwordChangeRequired.value = false
    passwordExpired.value = false
    passwordExpiresAt.value = 0
    passwordExpiryWarning.value = false
    passwordExpiryWarningShown = false
    captchaEnabled.value = false
    captchaId.value = ''
    captchaImage.value = ''
    captchaLoading.value = false
    captchaUnavailable.value = false
    useAppContentStore().resetContent()
    useAppShellStore().resetSessionTabs()
  }

  function resetSession() {
    authRevision++
    initRequest = null
    clearSessionState()
    ready.value = true
  }

  function isCurrentAuthRequest(revision: number, corpId: string) {
    return revision === authRevision && corpId === currentCorpId()
  }

  function handleAuthError(error: unknown) {
    const envelope = resolveErrorEnvelope(error)
    if (Number(envelope.code) === DINGTALK_H5_BIND_REQUIRED_CODE) {
      showBindRequired((envelope.data || {}) as BindRequiredPayload)
      return true
    }
    const message = firstText(envelope.msg, envelope.message)
    if (message === '无权限访问') {
      sessionAccessDenied.value = true
      sessionAccessDeniedMessage.value = message
      return true
    }
    if (['未登录', '登录已过期', '登录已过期或已被强制下线', '账号异常'].includes(message)) {
      resetSession()
      return true
    }
    if (message) {
      autoLoginMessage.value = message
    }
    return false
  }

  async function loadBootstrap(revision = authRevision, corpId = currentCorpId()) {
    const token = authToken()
    const res = await bootstrap()
    if (!isCurrentAuthRequest(revision, corpId) || token !== authToken())
      return false
    applySessionPayload(res)
    return true
  }

  async function loginWithPassword(form: LoginRequest) {
    if (contextCorpId !== currentCorpId()) {
      await syncCorpContext()
      return false
    }
    if (loading.value)
      return false
    const revision = authRevision
    const corpId = currentCorpId()
    loading.value = true
    let succeeded = false
    try {
      autoLoginMessage.value = ''
      resetBindState()
      const res = await login(form)
      if (!isCurrentAuthRequest(revision, corpId))
        return false
      const payload = resolveSessionPayload(res)
      if (payload.token) {
        setAuthToken(payload.token, corpId)
      }
      applySessionPayload(payload)
      if (!payload.menus) {
        await loadBootstrap(revision, corpId)
      }
      if (!isCurrentAuthRequest(revision, corpId))
        return false
      succeeded = true
      return true
    }
    finally {
      if (isCurrentAuthRequest(revision, corpId) && captchaEnabled.value && !succeeded) {
        await refreshLoginCaptcha()
      }
      if (isCurrentAuthRequest(revision, corpId))
        loading.value = false
    }
  }

  async function shouldTryDingTalkAutoLogin() {
    if (isDingTalkRuntime()) {
      return true
    }
    if (!currentCorpId()) {
      return false
    }
    await waitForDingTalkJSAPI(1200)
    return isDingTalkRuntime()
  }

  async function tryDingTalkAutoLogin() {
    if (autoLoginTried.value || authToken()) {
      return false
    }
    autoLoginTried.value = true
    autoLoginMessage.value = ''
    const revision = authRevision
    await ensurePublicConfig()
    if (revision !== authRevision)
      return false
    const corpId = currentCorpId()

    const shouldTry = await shouldTryDingTalkAutoLogin()
    if (isCurrentAuthRequest(revision, corpId))
      canRetryDingTalkLogin.value = shouldTry
    if (!shouldTry || !isCurrentAuthRequest(revision, corpId)) {
      return false
    }

    if (!corpId) {
      autoLoginMessage.value = '未配置钉钉企业 CorpId，请在后台配置企业应用，或在应用地址增加 corpId 参数。'
      return false
    }

    loading.value = true
    try {
      const authCode = await requestAuthCode(corpId)
      if (!isCurrentAuthRequest(revision, corpId))
        return false
      if (!authCode) {
        autoLoginMessage.value = '钉钉免登未返回授权码，请确认当前应用地址配置在钉钉工作台微应用中。'
        return false
      }
      const res = await ssoLogin({ corpId, authCode })
      if (!isCurrentAuthRequest(revision, corpId))
        return false
      const payload = resolveSessionPayload(res)
      if (payload.token) {
        setAuthToken(payload.token, corpId)
      }
      applySessionPayload(payload)
      if (!payload.menus) {
        await loadBootstrap(revision, corpId)
      }
      return isCurrentAuthRequest(revision, corpId) && Boolean(user.value)
    }
    catch (error) {
      if (isCurrentAuthRequest(revision, corpId))
        handleAuthError(error)
      return false
    }
    finally {
      if (isCurrentAuthRequest(revision, corpId))
        loading.value = false
    }
  }

  async function bindDingTalkUser(account: string, password: string) {
    const corpId = currentCorpId()
    if (contextCorpId !== corpId || (bindState.corpId && bindState.corpId !== corpId)) {
      await syncCorpContext()
      return false
    }
    if (loading.value)
      return false
    if (!bindState.bindTicket) {
      uni.showToast({ title: '绑定会话已过期，请重新打开应用', icon: 'none' })
      return false
    }
    const revision = authRevision
    loading.value = true
    try {
      const res = await bindSelf({
        bindTicket: bindState.bindTicket,
        account,
        password,
      })
      if (!isCurrentAuthRequest(revision, corpId))
        return false
      const payload = resolveSessionPayload(res)
      resetBindState()
      if (payload.token) {
        setAuthToken(payload.token, corpId)
      }
      applySessionPayload(payload)
      if (!payload.menus) {
        await loadBootstrap(revision, corpId)
      }
      return isCurrentAuthRequest(revision, corpId) && Boolean(user.value)
    }
    catch (error) {
      if (isCurrentAuthRequest(revision, corpId))
        handleAuthError(error)
      return false
    }
    finally {
      if (isCurrentAuthRequest(revision, corpId))
        loading.value = false
    }
  }

  async function restoreSession(revision: number, corpId: string) {
    if (!authToken()) {
      return false
    }
    try {
      return await loadBootstrap(revision, corpId)
    }
    catch (error) {
      if (!isCurrentAuthRequest(revision, corpId))
        return false
      const envelope = resolveErrorEnvelope(error)
      const message = firstText(envelope.msg, envelope.message)
      if (['未登录', '登录已过期', '登录已过期或已被强制下线', '账号异常'].includes(message))
        clearSessionState()
      else
        handleAuthError(error)
      return false
    }
  }

  function init(): Promise<void> {
    const requestedCorpId = currentCorpId()
    if (initRequest?.corpId === requestedCorpId)
      return initRequest.promise
    const revision = ++authRevision
    ready.value = false
    const promise = (async () => {
      await ensurePublicConfig()
      if (revision !== authRevision)
        return
      const corpId = currentCorpId()
      if (initRequest)
        initRequest.corpId = corpId
      const tokenCorpId = String(uni.getStorageSync(DINGTALK_H5_CONFIG.TOKEN_CORP_ID_KEY) || '')
      const changedCorp = contextCorpId !== null && contextCorpId !== corpId
      const wrongTokenCorp = authToken() && tokenCorpId !== corpId && (tokenCorpId || isDingTalkRuntime())
      // Old unmarked DingTalk sessions must identify their enterprise once after upgrading.
      const unknownTokenCorp = authToken() && !tokenCorpId && isDingTalkRuntime()
      if (changedCorp || wrongTokenCorp || unknownTokenCorp)
        clearSessionState()
      contextCorpId = corpId
      const restored = await restoreSession(revision, corpId)
      if (isCurrentAuthRequest(revision, corpId) && !restored)
        await tryDingTalkAutoLogin()
    })().finally(() => {
      if (revision === authRevision)
        ready.value = true
      if (initRequest?.promise === promise)
        initRequest = null
    })
    initRequest = { corpId: requestedCorpId, promise }
    return promise
  }

  async function syncCorpContext() {
    const corpId = currentCorpId()
    const tokenCorpId = String(uni.getStorageSync(DINGTALK_H5_CONFIG.TOKEN_CORP_ID_KEY) || '')
    if (!ready.value || contextCorpId !== corpId || (authToken() && tokenCorpId && tokenCorpId !== corpId))
      await init()
  }

  async function retryDingTalkLogin() {
    if (loading.value)
      return false
    resetSession()
    await init()
    return Boolean(user.value)
  }

  async function logoutSession() {
    const request = logout()
    // Clear immediately so a late logout response cannot erase a new enterprise session.
    resetSession()
    try {
      await request
    }
    catch {
      // The request layer reports the failure; local logout remains effective.
    }
  }

  async function completePasswordChange() {
    resetSession()
  }

  uni.$on('password-change-required', () => {
    passwordChangeRequired.value = true
  })

  return {
    apiPermissionKeys,
    apiPermissionReady,
    appConfig,
    appTitle,
    autoLoginMessage,
    canRetryDingTalkLogin,
    captchaEnabled,
    captchaId,
    captchaImage,
    captchaLoading,
    captchaUnavailable,
    bindDingTalkUser,
    bindState,
    buttonPermissionKeys,
    buttonPermissionReady,
    currentCorpId,
    flatMenus,
    handleAuthError,
    hasApiPermission,
    hasButtonPermission,
    hasMenuPermission,
    init,
    isLoggedIn,
    loadBootstrap,
    loading,
    loginWithPassword,
    logoutSession,
    menus,
    navItems,
    normalizedMenus,
    permissionVersion,
    passwordChangeRequired,
    passwordExpired,
    passwordExpiresAt,
    passwordExpiryWarning,
    completePasswordChange,
    ready,
    refreshLoginCaptcha,
    resetSession,
    retryDingTalkLogin,
    sessionAccessDenied,
    sessionAccessDeniedMessage,
    tryDingTalkAutoLogin,
    syncCorpContext,
    updateCurrentUser,
    user,
  }
}, {
  persist: false,
})
