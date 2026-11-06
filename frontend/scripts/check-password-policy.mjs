import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')
const auth = read('utils/auth.js')
for (const snippet of ['CLIENT_PASSWORD_STATE_KEY', 'clientPasswordChangeRequired', 'markClientPasswordChangeRequired', 'clearClientPasswordChangeRequired']) {
  assert.ok(auth.includes(snippet), `frontend auth missing ${snippet}`)
}
const request = read('utils/request.js')
for (const snippet of ['PASSWORD_CHANGE_REQUIRED', 'markClientPasswordChangeRequired', 'authState.passwordChangeUrl']) {
  assert.ok(request.includes(snippet), `frontend request missing ${snippet}`)
}
const api = read('api/index.js')
assert.ok(api.includes("patch(`${API_V2}/me/password`"), 'frontend API missing password change endpoint')
const pages = read('pages.json')
assert.ok(pages.includes('pages/my/change_password'), 'frontend pages.json missing password change page')
for (const file of ['pages/login/login.vue', 'pages/login/login_pwd.vue']) {
  assert.ok(read(file).includes('passwordChangeRequired'), `${file} must redirect restricted login`)
}
const page = read('pages/my/change_password.vue')
for (const snippet of ['当前密码', '新密码', '确认新密码', 'passportApi.changePassword', 'clearClientAuth', "uni.reLaunch({ url: '/pages/login/login_pwd' })"]) {
  assert.ok(page.includes(snippet), `frontend password page missing ${snippet}`)
}
const adminLogin = read('pages/admin/admin_login.vue')
assert.ok(adminLogin.includes('passwordChangeRequired'), 'mobile admin login must redirect restricted sessions')
const adminPassword = read('pages/admin/mgr/admin_mgr_pwd.vue')
for (const snippet of ['adminPasswordChangeRequired', 'clearAdminAuth', 'onBackPress', "uni.reLaunch({ url: '/pages/admin/admin_login' })"]) {
  assert.ok(adminPassword.includes(snippet), `mobile admin password page missing ${snippet}`)
}
console.log('Frontend password policy contracts passed')
