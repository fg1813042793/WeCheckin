import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import './check-user-password-reset-runtime.mjs'

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')

const types = read('src/types/passwordPolicy.ts')
for (const snippet of ['interface PasswordPolicy', 'minLength', 'maxAgeDays', 'historyCount', 'allowWeakTemporaryPassword']) {
  assert.ok(types.includes(snippet), `password policy types missing ${snippet}`)
}

const api = read('src/api/passwordPolicy.ts')
for (const snippet of ['/api/v2/admin/password-policy', 'getPasswordPolicy', 'savePasswordPolicy']) {
  assert.ok(api.includes(snippet), `password policy API missing ${snippet}`)
}

const setup = read('src/views/setup/index.vue')
const loginTabStart = setup.indexOf('<el-tab-pane label="登录配置" name="LOGIN_CONFIG">')
const passwordPolicyTabStart = setup.indexOf('<el-tab-pane label="密码策略" name="PASSWORD_POLICY">')
assert.ok(loginTabStart >= 0, 'setup must render login config tab')
assert.ok(passwordPolicyTabStart > loginTabStart, 'setup must render password policy as a separate tab')
assert.ok(
  !setup.slice(loginTabStart, passwordPolicyTabStart).includes('PasswordPolicySettings'),
  'login config tab must not render PasswordPolicySettings',
)
assert.ok(
  setup.slice(passwordPolicyTabStart).includes('<PasswordPolicySettings />'),
  'password policy tab must render PasswordPolicySettings',
)

const router = read('src/router/index.ts')
for (const snippet of ['/change-password', 'passwordChangeRequired', 'adminPasswordChangeRequired']) {
  assert.ok(router.includes(snippet), `router forced-password contract missing ${snippet}`)
}

const request = read('src/utils/request.ts')
for (const snippet of ['PASSWORD_CHANGE_REQUIRED', 'markAdminPasswordChangeRequired', '/change-password']) {
  assert.ok(request.includes(snippet), `request password restriction handling missing ${snippet}`)
}

const login = read('src/views/login/index.vue')
assert.ok(login.includes('passwordChangeRequired'), 'login must redirect restricted sessions')

const loginTypes = read('src/api/types.ts')
for (const snippet of ['passwordChangeRequired', 'passwordExpired', 'passwordExpiresAt', 'passwordExpiryWarning']) {
  assert.ok(loginTypes.includes(snippet), `admin login type missing ${snippet}`)
}

const changePage = read('src/views/change-password/index.vue')
for (const snippet of ['当前密码', '新密码', '确认新密码', 'adminApi.mgrPwd', 'clearAdminSession', "router.replace('/login')"]) {
  assert.ok(changePage.includes(snippet), `forced password page missing ${snippet}`)
}

const managerPage = read('src/views/mgr/index.vue')
for (const snippet of ['adminApi.mgrPwd', 'clearAdminSession', "router.replace('/login')"]) {
  assert.ok(managerPage.includes(snippet), `manager password dialog missing revoked-session handling ${snippet}`)
}

console.log('password policy Admin contracts passed')
