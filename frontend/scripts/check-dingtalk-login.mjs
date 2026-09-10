import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const apiSource = readFileSync(resolve(root, 'api/index.js'), 'utf8')
const loginSource = readFileSync(resolve(root, 'pages/login/login.vue'), 'utf8')
const passwordSource = readFileSync(resolve(root, 'pages/login/login_pwd.vue'), 'utf8')
const oauthSource = readFileSync(resolve(root, 'utils/dingtalkOAuth.js'), 'utf8')
const oauthModule = await import('../utils/dingtalkOAuth.js')

for (const snippet of [
  'dingtalkLoginConfig()',
  'dingtalkAuthorization(data)',
  'loginByDingTalk(data)',
  '`${API_V2}/auth/dingtalk-config`',
  '`${API_V2}/auth/dingtalk-authorization`',
  '`${API_V2}/auth/dingtalk-login`',
]) {
  if (!apiSource.includes(snippet)) {
    throw new Error(`DingTalk client API missing: ${snippet}`)
  }
}
for (const snippet of [
  'return postJSON(`${API_V2}/auth/dingtalk-authorization`, data)',
  'return postJSON(`${API_V2}/auth/dingtalk-login`, data)',
]) {
  if (!apiSource.includes(snippet)) {
    throw new Error(`DingTalk client API must use the documented JSON contract: ${snippet}`)
  }
}

for (const snippet of [
  '钉钉登录',
  'beginDingTalkOAuth',
  'completeDingTalkOAuth',
  'ensureClientPermissionSnapshot',
  'await ensureClientPermissionSnapshot()',
]) {
  if (!loginSource.includes(snippet)) {
    throw new Error(`DingTalk login page missing: ${snippet}`)
  }
}
for (const forbidden of ['v-model="form.user_id"', 'wechatLogin()', 'wx.login(', 'passportApi.login(']) {
  if (loginSource.includes(forbidden) || passwordSource.includes(forbidden)) {
    throw new Error(`legacy public identity login must be removed from frontend: ${forbidden}`)
  }
}
if (!passwordSource.includes('钉钉登录')) {
  throw new Error('password login page must offer DingTalk login as the primary alternative')
}
for (const snippet of [
  'DINGTALK_OAUTH_STATE_KEY',
  'DINGTALK_OAUTH_CONSUMED_STATE_KEY',
  'VITE_DINGTALK_OAUTH_REDIRECT_URI',
]) {
  if (!oauthSource.includes(snippet)) {
    throw new Error(`DingTalk OAuth helper missing: ${snippet}`)
  }
}
for (const forbidden of ['new URLSearchParams(', 'new URL(']) {
  if (oauthSource.includes(forbidden)) {
    throw new Error(`DingTalk OAuth helper must not depend on unavailable App Web API: ${forbidden}`)
  }
}
if (typeof oauthModule.parseDingTalkOAuthQuery !== 'function') {
  throw new Error('DingTalk OAuth helper must expose its App-compatible query parser')
}
const parsedQuery = oauthModule.parseDingTalkOAuthQuery('?authCode=code%2B1&state=state+1')
if (parsedQuery.authCode !== 'code+1' || parsedQuery.state !== 'state 1') {
  throw new Error(`DingTalk OAuth query parser decoded values incorrectly: ${JSON.stringify(parsedQuery)}`)
}
for (const forbidden of ['dingtalk://dingtalkclient/page/link', 'plus.runtime.arguments']) {
  if (oauthSource.includes(forbidden)) {
    throw new Error(`H5 DingTalk OAuth helper must not retain the retired App bridge: ${forbidden}`)
  }
}

console.log('DingTalk client login checks passed.')
