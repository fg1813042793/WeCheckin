import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const setup = readFileSync(new URL('../src/views/setup/index.vue', import.meta.url), 'utf8')
const clientSetup = readFileSync(new URL('../src/views/client-setup/index.vue', import.meta.url), 'utf8')
const dingTalkSetup = readFileSync(new URL('../src/views/dingtalk-setup/index.vue', import.meta.url), 'utf8')
const auditSnapshot = readFileSync(new URL('../src/views/log/auditSnapshot.ts', import.meta.url), 'utf8')

for (const [source, name, key] of [
  [setup, 'system setup', 'TOKEN_ADMIN_SLIDING_EXPIRE_ENABLED'],
  [clientSetup, 'client setup', 'TOKEN_USER_SLIDING_EXPIRE_ENABLED'],
]) {
  for (const pattern of [
    'label="Token 滑动过期"',
    'tokenConfig.slidingExpireEnabled',
    `key: '${key}'`,
    '关闭（固定有效期）',
  ]) {
    assert.ok(source.includes(pattern), `${name} missing sliding expiration contract: ${pattern}`)
  }
}
for (const pattern of ['label="Token 滑动过期"', 'form.slidingExpireEnabled', 'slidingExpireEnabled: form.slidingExpireEnabled']) {
  assert.ok(dingTalkSetup.includes(pattern), `DingTalk setup missing sliding expiration contract: ${pattern}`)
}
for (const [key, label] of [
  ['TOKEN_ADMIN_SLIDING_EXPIRE_ENABLED', '管理员 Token 滑动过期'],
  ['TOKEN_USER_SLIDING_EXPIRE_ENABLED', '用户 Token 滑动过期'],
  ['TOKEN_DINGTALK_H5_SLIDING_EXPIRE_ENABLED', '钉钉 H5 Token 滑动过期'],
]) {
  assert.ok(auditSnapshot.includes(`${key}: '${label}'`), `audit snapshot must label ${key}`)
  assert.ok(auditSnapshot.includes(`'${key}'`), `audit snapshot must render ${key} as a public boolean`)
}

console.log('token sliding expiration settings checks passed')
