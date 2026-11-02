import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const { createPinia, setActivePinia } = require('pinia')
const ts = require('typescript')

function load(path, dependencies) {
  const source = readFileSync(new URL(`../src/${path}`, import.meta.url), 'utf8').replace(/\/\/ #ifndef H5[\s\S]*?\/\/ #endif/g, '')
  const compiled = ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
  }).outputText
  const module = { exports: {} }
  runInThisContext(`(function(module,exports,require){${compiled}\n})`)(module, module.exports, name => dependencies[name] || require(name))
  return module.exports
}
function deferred() {
  let resolve
  let reject
  const promise = new Promise((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}
const config = { TOKEN_KEY: 'DT_H5_TOKEN', TOKEN_CORP_ID_KEY: 'DT_H5_TOKEN_CORP_ID', DINGTALK_CORP_ID: '', APP_NAME: 'Test' }
const session = (id, token = `token-${id}`) => ({ data: { token, user: { id }, menus: [], apiPermissionKeys: ['test'], buttonPermissionKeys: ['test'] } })
const binding = corpId => ({ data: { code: 10020, data: { corpId, bindTicket: `ticket-${corpId}`, expiresIn: 600 } } })
function fixture({ cachedCorp, cachedToken = '', runtime = true } = {}) {
  const storage = new Map()
  if (cachedToken)
    storage.set(config.TOKEN_KEY, cachedToken)
  if (cachedCorp)
    storage.set(config.TOKEN_CORP_ID_KEY, cachedCorp)
  globalThis.uni = {
    getStorageSync: key => storage.get(key) || '',
    setStorageSync: (key, value) => storage.set(key, value),
    removeStorageSync: key => storage.delete(key),
    showToast: () => {},
  }
  const calls = { codes: [], sso: [], bootstrap: 0, logout: 0, resetContent: 0, bind: 0 }
  const responses = {
    sso: async ({ corpId }) => session(corpId),
    bootstrap: async () => session('restored', ''),
    bind: async () => session('bound'),
    login: async () => session('password'),
    logout: async () => {},
  }
  const api = {
    authToken: () => storage.get(config.TOKEN_KEY) || '',
    clearAuthToken: () => {
      storage.delete(config.TOKEN_KEY)
      storage.delete(config.TOKEN_CORP_ID_KEY)
    },
    setAuthToken: (token, corpId) => {
      storage.set(config.TOKEN_KEY, token)
      storage.set(config.TOKEN_CORP_ID_KEY, corpId || '')
    },
    publicConfig: async () => ({ data: { corpId: 'corp-a' } }),
    captcha: async () => ({ data: { enabled: false } }),
    ssoLogin: async (data) => {
      calls.sso.push(data.corpId)
      return responses.sso(data)
    },
    bootstrap: async () => {
      calls.bootstrap++
      return responses.bootstrap()
    },
    bindSelf: async (data) => {
      calls.bind++
      return responses.bind(data)
    },
    login: data => responses.login(data),
    logout: async () => {
      calls.logout++
      return responses.logout()
    },
  }
  const location = new URL('https://fixture.invalid/app?corpId=corp-a#/pages/index/index')
  globalThis.window = { location }
  setActivePinia(createPinia())
  const appShell = load('stores/appShell.ts', {})
  const { useDingtalkAuthStore } = load('stores/dingtalkAuth.ts', {
    '@/api/dingtalk-h5': api,
    '@/config/dingtalk-h5': { DINGTALK_H5_CONFIG: config, DINGTALK_H5_BIND_REQUIRED_CODE: 10020 },
    '@/config/app-navigation': { createAppNavItems: value => value, flattenAppNav: value => value, normalizeAppMenus: value => value },
    '@/utils/dingtalk': {
      isDingTalkRuntime: () => runtime,
      waitForDingTalkJSAPI: async () => {},
      requestAuthCode: async (corpId) => {
        calls.codes.push(corpId)
        return `code-${corpId}`
      },
      setBrowserFavicon: () => {},
      setNavigationTitle: () => {},
    },
    './appContent': { useAppContentStore: () => ({ resetContent: () => calls.resetContent++ }) },
    './appShell': appShell,
  })
  return {
    auth: useDingtalkAuthStore(),
    shell: appShell.useAppShellStore(),
    calls,
    storage,
    responses,
    changeCorp: (corpId) => {
      location.search = `?corpId=${corpId}`
    },
  }
}

{
  const f = fixture()
  await f.auth.init()
  f.shell.rememberOpenedTab('workflow')
  f.shell.setMenuLayout('top')
  await f.auth.logoutSession()
  assert.equal(f.auth.user, null)
  assert.deepEqual(f.shell.openedTabKeys, [], 'logout must remove ordinary menu tabs as well as dynamic tabs')
  assert.equal(f.shell.menuLayout, 'top', 'user layout preference must be retained')
  assert.equal(f.calls.sso.length, 1, 'logout must not immediately sign the user back in')
  f.responses.sso = async ({ corpId }) => Promise.reject(binding(corpId))
  await f.auth.tryDingTalkAutoLogin()
  assert.equal(f.calls.sso.length, 2, 'logout must permit another DingTalk authentication attempt')
  assert.equal(f.auth.bindState.corpId, 'corp-a')
  assert.equal(f.auth.bindState.visible, true)
  assert.deepEqual(f.auth.apiPermissionKeys, [], 'binding must not retain the previous user permissions')
}
{
  const f = fixture({ cachedToken: 'token-a', cachedCorp: 'corp-a' })
  f.changeCorp('corp-b')
  f.responses.sso = async ({ corpId }) => Promise.reject(binding(corpId))
  await f.auth.init()
  assert.equal(f.calls.bootstrap, 0, 'a new enterprise must not bootstrap the old enterprise token')
  assert.deepEqual(f.calls.sso, ['corp-b'])
  assert.equal(f.auth.bindState.corpId, 'corp-b')
  assert.equal(f.storage.has(config.TOKEN_KEY), false)
}
{
  const f = fixture({ cachedToken: 'token-a', cachedCorp: 'corp-a' })
  await f.auth.init()
  assert.equal(f.calls.bootstrap, 1, 'same-enterprise sessions can still be restored')
  assert.equal(f.calls.sso.length, 0)
  f.changeCorp('corp-b')
  f.responses.sso = async ({ corpId }) => Promise.reject(binding(corpId))
  await f.auth.syncCorpContext()
  assert.equal(f.auth.user, null)
  assert.equal(f.auth.bindState.corpId, 'corp-b')
  assert.ok(f.calls.resetContent > 0, 'switching enterprise must discard old business tabs')
  await f.auth.syncCorpContext()
  assert.equal(f.calls.sso.length, 1, 'focus and visibility events must not repeat authentication for the same enterprise')
}
{
  const f = fixture({ cachedToken: 'legacy-token' })
  await f.auth.init()
  assert.equal(f.calls.bootstrap, 0, 'unmarked legacy DingTalk tokens must be identified once')
  assert.equal(f.storage.get(config.TOKEN_CORP_ID_KEY), 'corp-a')
}
{
  const f = fixture()
  f.responses.sso = async () => Promise.reject(Object.assign(new Error('request failed'), { data: { code: 1, msg: '钉钉企业配置不可用' } }))
  await f.auth.init()
  assert.equal(f.auth.bindState.visible, false, 'authentication failure is not proof of an unbound identity')
  assert.equal(f.auth.autoLoginMessage, '钉钉企业配置不可用')
  f.responses.sso = async ({ corpId }) => Promise.reject(binding(corpId))
  await f.auth.retryDingTalkLogin()
  assert.equal(f.auth.bindState.visible, true)
  assert.equal(f.auth.autoLoginMessage, '')
  assert.equal(f.calls.sso.length, 2)
}
{
  const f = fixture()
  const pendingA = deferred()
  f.responses.sso = async ({ corpId }) => corpId === 'corp-a' ? pendingA.promise : Promise.reject(binding(corpId))
  const old = f.auth.init()
  for (let n = 0; n < 20 && !f.calls.sso.length; n++) await Promise.resolve()
  assert.equal(f.calls.sso[0], 'corp-a')
  f.changeCorp('corp-b')
  await f.auth.syncCorpContext()
  pendingA.resolve(session('corp-a'))
  await old
  assert.equal(f.auth.user, null, 'a late response from A must not sign into B')
  assert.equal(f.auth.bindState.corpId, 'corp-b')
  assert.equal(f.storage.has(config.TOKEN_KEY), false)
}
{
  const f = fixture()
  f.responses.sso = async ({ corpId }) => Promise.reject(binding(corpId))
  await f.auth.init()
  const pendingBind = deferred()
  f.responses.bind = async () => pendingBind.promise
  const old = f.auth.bindDingTalkUser('fixture', 'fixture-password')
  f.changeCorp('corp-b')
  await f.auth.syncCorpContext()
  pendingBind.resolve(session('old-bound-user'))
  await old
  assert.equal(f.auth.user, null)
  assert.equal(f.auth.bindState.bindTicket, 'ticket-corp-b', 'a late bind result must not remove the new enterprise ticket')
}
{
  const f = fixture()
  f.responses.sso = async ({ corpId }) => Promise.reject(binding(corpId))
  await f.auth.init()
  f.changeCorp('corp-b')
  await f.auth.bindDingTalkUser('fixture', 'fixture-password')
  assert.equal(f.calls.bind, 0, 'an old enterprise binding ticket must never be submitted after the URL changes')
  assert.equal(f.auth.bindState.corpId, 'corp-b')
}
{
  const f = fixture({ cachedToken: 'browser-password-token', runtime: false })
  await f.auth.init()
  assert.equal(f.calls.bootstrap, 1, 'ordinary browser password sessions remain supported')
  await f.auth.logoutSession()
  assert.equal(f.auth.user, null)
  assert.equal(f.calls.sso.length, 0)
}
{
  const f = fixture({ cachedToken: 'token-a', cachedCorp: 'corp-a' })
  const pending = deferred()
  f.responses.bootstrap = () => pending.promise
  const old = f.auth.init()
  for (let n = 0; n < 20 && !f.calls.bootstrap; n++) await Promise.resolve()
  f.changeCorp('corp-b')
  await f.auth.syncCorpContext()
  pending.resolve(session('old-bootstrap-user'))
  await old
  assert.equal(f.auth.user.id, 'corp-b')
  assert.equal(f.storage.get(config.TOKEN_CORP_ID_KEY), 'corp-b')
}
{
  const f = fixture()
  await f.auth.init()
  const pending = deferred()
  f.responses.logout = () => pending.promise
  const oldLogout = f.auth.logoutSession()
  assert.equal(f.auth.user, null, 'logout must clear local user data immediately')
  f.changeCorp('corp-b')
  await f.auth.syncCorpContext()
  pending.resolve()
  await oldLogout
  assert.equal(f.auth.user.id, 'corp-b', 'a late logout response must not remove the new enterprise user')
}
{
  const f = fixture({ cachedToken: 'expired-token', cachedCorp: 'corp-a' })
  f.responses.bootstrap = async () => Promise.reject(Object.assign(new Error('expired'), { data: { code: 1, msg: '登录已过期' } }))
  f.responses.sso = async ({ corpId }) => Promise.reject(binding(corpId))
  await f.auth.init()
  assert.equal(f.auth.bindState.visible, true, 'expired cached sessions must proceed to fresh enterprise authentication')
  assert.equal(f.auth.ready, true)
}
{
  const f = fixture()
  const pending = deferred()
  f.responses.sso = () => pending.promise
  const first = f.auth.init()
  const second = f.auth.init()
  for (let n = 0; n < 20 && !f.calls.sso.length; n++) await Promise.resolve()
  assert.equal(f.calls.sso.length, 1, 'mount/show/focus must share an in-flight initialization')
  pending.resolve(session('corp-a'))
  await Promise.all([first, second])
}
{
  const f = fixture({ cachedToken: 'token-b', cachedCorp: 'corp-b' })
  const dependencies = { '@/config/dingtalk-h5': { DINGTALK_H5_CONFIG: config, DINGTALK_H5_AUTH_EXPIRED_EVENT: 'expired' } }
  const expiry = load('common/dingtalk-h5-auth-expiry.ts', dependencies)
  const { httpInterceptor } = load('common/http.interceptor.ts', {
    ...dependencies,
    '@/common/app-error': load('common/app-error.ts', {}),
    '@/common/dingtalk-h5-auth-expiry': expiry,
  })
  globalThis.uni.$emit = () => {}
  const expired = { statusCode: 200, data: { code: 1, msg: '登录已过期' }, config: { meta: { toast: false }, header: { Authorization: 'token-a' } } }
  assert.equal(httpInterceptor.response({ ...expired, data: { code: 0, data: { id: 'old-user-a' } } }), false, 'a late successful profile response must not be passed to the new enterprise UI')
  httpInterceptor.response(expired)
  assert.equal(f.storage.get(config.TOKEN_KEY), 'token-b', 'an old request expiration must not clear the new enterprise token')
  httpInterceptor.response({ ...expired, config: { ...expired.config, header: { Authorization: 'token-b' } } })
  assert.equal(f.storage.has(config.TOKEN_KEY), false, 'current session expiration must still work')
  assert.equal(f.storage.has(config.TOKEN_CORP_ID_KEY), false)
}
{
  const f = fixture({ cachedToken: 'token-a', cachedCorp: 'corp-a' })
  const dependencies = { '@/config/dingtalk-h5': { DINGTALK_H5_CONFIG: config, DINGTALK_H5_AUTH_EXPIRED_EVENT: 'expired' } }
  const base = load('api/dingtalk-h5/base.ts', {
    ...dependencies,
    'uview-pro': { http: {}, useLocale: () => ({ t: key => key }) },
    '@/common/app-error': load('common/app-error.ts', {}),
    '@/common/dingtalk-h5-auth-expiry': load('common/dingtalk-h5-auth-expiry.ts', dependencies),
  })
  globalThis.uni.$emit = () => {}
  const originalFetch = globalThis.fetch
  try {
    for (const envelope of [{ code: 1, msg: '登录已过期' }, { code: 0, data: { id: 'old-user-a' } }]) {
      f.storage.set(config.TOKEN_KEY, 'token-a')
      globalThis.fetch = async (_url, options) => {
        assert.equal(options.headers.Authorization, 'token-a')
        f.storage.set(config.TOKEN_KEY, 'token-b')
        f.storage.set(config.TOKEN_CORP_ID_KEY, 'corp-b')
        return { ok: true, text: async () => JSON.stringify(envelope) }
      }
      await assert.rejects(base.uploadMultipartFiles('https://fixture.invalid/upload', [], {}))
      assert.equal(f.storage.get(config.TOKEN_KEY), 'token-b', 'old multipart responses must not clear the new session')
    }
    const blob = deferred()
    const fetches = []
    f.storage.set(config.TOKEN_KEY, 'token-a')
    globalThis.fetch = async (url) => {
      fetches.push(url)
      return { ok: true, blob: () => blob.promise, text: async () => JSON.stringify({ code: 0 }) }
    }
    const pending = base.uploadMultipartFiles('https://fixture.invalid/upload', [{ filePath: 'blob:fixture' }])
    await Promise.resolve()
    f.storage.set(config.TOKEN_KEY, 'token-b')
    blob.resolve(new Blob(['fixture'], { type: 'image/png' }))
    await assert.rejects(pending)
    assert.deepEqual(fetches, ['blob:fixture'], 'files prepared under A must not be uploaded using B credentials')

    let uploadOptions
    f.storage.set(config.TOKEN_KEY, 'token-a')
    globalThis.uni.uploadFile = options => uploadOptions = options
    const nativeUpload = base.uploadFile('https://fixture.invalid/avatar', 'blob:fixture')
    f.storage.set(config.TOKEN_KEY, 'token-b')
    uploadOptions.success({ statusCode: 200, data: JSON.stringify({ code: 0, data: { id: 'old-user-a' } }) })
    await assert.rejects(nativeUpload)
    assert.equal(f.storage.get(config.TOKEN_KEY), 'token-b')
  }
  finally {
    globalThis.fetch = originalFetch
  }
}
console.log('DingTalk enterprise context: logout/retry, cached sessions, switching, binding, stale responses, single initialization and browser fallback passed')
