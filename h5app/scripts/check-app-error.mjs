import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const ts = require('typescript')

const source = readFileSync(new URL('../src/common/app-error.ts', import.meta.url), 'utf8')
const output = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2020,
    esModuleInterop: true,
  },
  fileName: 'app-error.ts',
}).outputText
const module = { exports: {} }
// eslint-disable-next-line no-new-func
new Function('module', 'exports', 'require', output)(module, module.exports, require)

const { appErrorMessage, showAppError } = module.exports
const toasts = []
globalThis.uni = {
  showToast(options) {
    toasts.push(options)
  },
}

assert.equal(appErrorMessage({ msg: '具体业务错误' }, '兜底'), '具体业务错误')
assert.equal(appErrorMessage({ message: '具体异常信息' }, '兜底'), '具体异常信息')
assert.equal(appErrorMessage({ data: { message: '嵌套错误信息' } }, '兜底'), '嵌套错误信息')
assert.equal(appErrorMessage({ response: { data: { msg: '接口错误信息' } } }, '兜底'), '接口错误信息')
assert.equal(appErrorMessage({ errMsg: '设备调用失败' }, '兜底'), '设备调用失败')
assert.equal(appErrorMessage(new Error('Error 具体信息'), '兜底'), 'Error 具体信息')
assert.equal(appErrorMessage({ message: '   ' }, '兜底'), '兜底')
assert.equal(appErrorMessage({ stack: '/Users/example/project/app.ts:12' }, '兜底'), '兜底')
assert.equal(appErrorMessage({ message: 'Bearer secret-token' }, '兜底'), '兜底')

assert.equal(showAppError({ data: { msg: '请先完善表单' } }, '操作失败'), '请先完善表单')
assert.deepEqual(toasts, [{ title: '请先完善表单', icon: 'none' }])

console.log('app error checks passed')
