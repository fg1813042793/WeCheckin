import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')
const types = read('src/types/dingtalk-h5.ts')
for (const snippet of ['passwordChangeRequired', 'passwordExpired', 'passwordExpiresAt', 'passwordExpiryWarning']) {
  assert.ok(types.includes(snippet), `H5 auth types missing ${snippet}`)
}
const store = read('src/stores/dingtalkAuth.ts')
for (const snippet of ['passwordChangeRequired', 'applySessionPayload', 'completePasswordChange', 'password-change-required']) {
  assert.ok(store.includes(snippet), `H5 auth store missing ${snippet}`)
}
const completePasswordChangeStart = store.indexOf('async function completePasswordChange()')
const completePasswordChangeEnd = store.indexOf('uni.$on(\'password-change-required\'', completePasswordChangeStart)
const completePasswordChange = store.slice(completePasswordChangeStart, completePasswordChangeEnd)
assert.ok(completePasswordChange.includes('resetSession()'), 'H5 password change completion must clear the revoked local session')
assert.ok(!completePasswordChange.includes('loadBootstrap()'), 'H5 password change completion must not reuse the revoked token')
const guard = read('src/components/app-auth-guard/app-auth-guard.vue')
for (const snippet of ['PasswordChangePanel', 'auth.passwordChangeRequired', '@completed="auth.completePasswordChange"']) {
  assert.ok(guard.includes(snippet), `H5 auth guard missing ${snippet}`)
}
const forcedChangeBranch = guard.indexOf('v-else-if="auth.passwordChangeRequired"')
const unauthenticatedBranch = guard.indexOf('v-else-if="!auth.user')
assert.ok(forcedChangeBranch >= 0, 'H5 auth guard must render forced password change branch')
assert.ok(
  forcedChangeBranch < unauthenticatedBranch,
  'H5 auth guard must prioritize forced password change over unauthenticated panels',
)
const panel = read('src/components/password-change-panel/password-change-panel.vue')
const shell = read('src/components/app-shell/app-shell.vue')
assert.ok(shell.includes('await auth.completePasswordChange()'), 'personal-center password changes must clear the same revoked session as forced changes')
for (const snippet of ['当前密码', '新密码', '确认新密码', 'changePassword', 'emit(\'completed\')']) {
  assert.ok(panel.includes(snippet), `H5 forced password panel missing ${snippet}`)
}
const interceptor = read('src/common/http.interceptor.ts')
for (const snippet of ['PASSWORD_CHANGE_REQUIRED', 'uni.$emit(\'password-change-required\')']) {
  assert.ok(interceptor.includes(snippet), `H5 interceptor missing ${snippet}`)
}
console.log('H5App password policy contracts passed')
