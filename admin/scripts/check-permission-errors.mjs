import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const messages = []
let fulfilled, rejected
const api = { interceptors: { request: { use() {} }, response: { use(ok, fail) { fulfilled = ok; rejected = fail } } } }
const request = executeTypeScriptModule(readFileSync(fileURLToPath(new URL('../src/utils/request.ts', import.meta.url)), 'utf8'), 'request.ts', {
  axios: { create: () => api },
  'element-plus': { ElMessage: { error: message => messages.push(message) } },
  './adminSession': { clearAdminSession: () => { throw new Error('permission failure must not clear login') } },
})

const missing = { code: 1, msg: '缺少权限-admin:api:menu:list（菜单查看接口）' }
for (let i = 0; i < 2; i++) {
  await assert.rejects(fulfilled({ config: {}, data: { ...missing } }), error => {
    assert.equal(error.msg, missing.msg)
    assert.equal(request.isRequestErrorNotified(error), true)
    request.showRequestError(error, '不应重复弹出')
    return true
  })
}
assert.equal(messages.length, 2)
for (const message of messages) {
  assert.equal(message.message, missing.msg)
  assert.equal(message.grouping, true, 'concurrent failures for the same permission should group')
  assert.ok(message.duration >= 6000, 'full permission keys need time to read')
}
await assert.rejects(fulfilled({ config: {}, data: { code: 1, msg: '业务校验失败' } }))
assert.equal(messages.at(-1), '业务校验失败', 'ordinary business errors retain existing behavior')
await assert.rejects(rejected(new Error('offline')))
assert.equal(messages.at(-1), '网络错误')
api.get = async () => ({ headers: { 'content-type': 'application/json' }, data: { text: async () => JSON.stringify(missing) } })
await assert.rejects(request.getBlob('/api/v2/admin/logs/export'), error => {
  assert.equal(error.msg, missing.msg)
  assert.equal(request.isRequestErrorNotified(error), true)
  return true
})
assert.equal(messages.at(-1).message, missing.msg)
assert.equal(messages.at(-1).grouping, true, 'downloads use the same permission message behavior')
assert.deepEqual(fulfilled({ config: {}, data: { code: 0, data: [1] } }), { code: 0, data: [1] })
console.log('Permission errors: exact server message, grouped notices and login preservation passed.')
