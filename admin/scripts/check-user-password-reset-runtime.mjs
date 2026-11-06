import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

let options, resolvePrompt, rejectPrompt, releaseRequest
let requestError = null
const requests = [], successes = [], errors = []
const dependencies = {
  'element-plus': {
    ElMessage: { success: message => successes.push(message) },
    ElMessageBox: { prompt: (_message, _title, config) => {
      options = config
      return new Promise((resolve, reject) => { resolvePrompt = resolve; rejectPrompt = reject })
    } },
  },
  '../../api': { adminApi: { userResetPwd: async payload => {
    requests.push(payload)
    await new Promise(resolve => { releaseRequest = resolve })
    if (requestError) throw requestError
  } } },
  '../../utils/request': { showRequestError: error => errors.push(error) },
}
const source = readFileSync(new URL('../src/views/user/resetPassword.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
const module = { exports: {} }
new Function('require', 'module', 'exports', compiled)(name => dependencies[name], module, module.exports)

const flow = module.exports.resetPwd({ id: 9, name: '回归用户' })
assert.equal(options.inputType, 'password')
for (const value of ['', '      ', 'short', 'a'.repeat(73), '密'.repeat(25)]) {
  assert.notEqual(options.inputValidator(value), true)
}
assert.equal(options.inputValidator(' Fresh123! '), true)
const instance = { inputValue: 'Fresh123!', confirmButtonLoading: false }
let closed = 0
const done = () => { closed++; resolvePrompt({ value: instance.inputValue }) }
requestError = { msg: '新密码不能与当前或最近使用的密码相同' }
const rejected = options.beforeClose('confirm', instance, done)
assert.equal(instance.confirmButtonLoading, true)
await options.beforeClose('confirm', instance, done)
await options.beforeClose('cancel', instance, done)
assert.equal(requests.length, 1, 'in-flight confirmation must not submit twice')
assert.equal(closed, 0, 'cannot dismiss an unresolved reset')
releaseRequest()
await rejected
assert.equal(closed, 0, 'failed reset must retain the dialog for retry')
assert.equal(instance.confirmButtonLoading, false)
assert.equal(errors[0], requestError)
assert.equal(successes.length, 0)
requestError = null
instance.inputValue = ' Fresh456! '
const retried = options.beforeClose('confirm', instance, done)
releaseRequest()
await retried
await flow
assert.deepEqual(requests.at(-1), { id: 9, password: ' Fresh456! ' }, 'forward exact input, not the legacy default')
assert.equal(closed, 1)
assert.deepEqual(successes, ['密码已重置'])
assert.ok(!successes[0].includes(instance.inputValue), 'success messages must not disclose the password')
const canceled = module.exports.resetPwd({ id: 9, name: '回归用户' })
await options.beforeClose('cancel', { confirmButtonLoading: false }, () => rejectPrompt('cancel'))
await canceled
assert.equal(requests.length, 2)
assert.equal(errors.length, 1)
console.log('user password reset runtime passed: validation, exact input, retry, duplicate prevention and cancel')
