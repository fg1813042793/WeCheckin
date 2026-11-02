import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const ts = require('typescript')

const read = path => readFileSync(new URL(`../src/${path}`, import.meta.url), 'utf8')
function compile(source) {
  return ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
  }).outputText
}
function load(path, dependencies = {}) {
  const module = { exports: {} }
  runInThisContext(`(function(module,exports,require){${compile(read(path))}\n})`)(
    module,
    module.exports,
    name => dependencies[name] || require(name),
  )
  return module.exports
}
const appErrors = load('common/app-error.ts')
const dependencies = {
  '@/common/app-error': appErrors,
  '@/config/dingtalk-h5': { DINGTALK_H5_CONFIG: {}, DINGTALK_H5_BIND_REQUIRED_CODE: 10002 },
  '@/common/dingtalk-h5-auth-expiry': {
    isApiEnvelope: value => Boolean(value && typeof value === 'object' && 'code' in value),
    handleDingTalkH5AuthExpired: () => {},
  },
}
const { httpInterceptor, httpRequestConfig } = load('common/http.interceptor.ts', dependencies)
const toasts = []
globalThis.uni = { showToast: options => toasts.push(options), getStorageSync: () => '' }
let reply
let calls = 0
const base = load('api/dingtalk-h5/base.ts', {
  ...dependencies,
  'uview-pro': { http: {
    async post(url, data, options = {}) {
      calls++
      assert.equal(url, '/api/v2/dingtalk/h5/workflows/tasks/task-1/complete')
      assert.equal(data.formData.score, 35)
      const response = { ...reply, config: { meta: { ...httpRequestConfig.meta, ...options.meta } } }
      const result = httpInterceptor.response(response)
      if (result === false)
        throw response
      return result
    },
  } },
})
const api = load('api/workflow.ts', { '@/api/dingtalk-h5/base': base })
const source = read('pages/workflow/components/WorkflowDetailPanel.vue').match(/<script setup lang="ts">([\s\S]*?)<\/script>/)[1]
const ast = ts.createSourceFile('WorkflowDetailPanel.ts', source, ts.ScriptTarget.Latest, true)
const functions = ast.statements.filter(node => ts.isFunctionDeclaration(node)
  && ['submitTask', 'workflowRequestErrorMessage', 'beginSubmission', 'finishSubmission'].includes(node.name?.text))
function submitter(scope) {
  return runInThisContext(`(function(${Object.keys(scope).join(',')}){${compile(functions.map(node => node.getText(ast)).join('\n'))}\nreturn submitTask})`)(...Object.values(scope))
}

for (const action of ['approve', 'submit']) {
  const formData = { value: { score: 35, comment: '已填写的评价' } }
  const submitting = { value: false }
  const submittingAction = { value: '' }
  let changed = 0
  let reloads = 0
  const submit = submitter({
    activeTask: { value: { id: 'task-1' } },
    detail: { value: { form: [] } },
    canApprove: { value: true },
    canSubmit: { value: true },
    submitting,
    submittingAction,
    formRef: { value: { validate: () => ({ valid: true }) } },
    formData,
    fieldAccess: { value: {} },
    writableWorkflowFormData: (_fields, data) => ({ ...data }),
    activeNodeType: { value: action === 'approve' ? 'approval' : 'handle' },
    confirmAction: async () => true,
    completeWorkflowTask: api.completeWorkflowTask,
    appErrorMessage: appErrors.appErrorMessage,
    pagePresentation: { value: false },
    emit: () => changed++,
    loadDetail: async () => reloads++,
  })
  for (const failure of [
    { statusCode: 200, data: { code: 1, msg: '流程节点未解析到处理人，请联系管理员检查审批人配置' } },
    { statusCode: 403, data: { code: 1, msg: '无权执行该流程操作' } },
    { errMsg: 'request:fail timeout' },
  ]) {
    reply = failure
    toasts.length = 0
    const before = calls
    await submit(action)
    assert.equal(calls, before + 1, 'submission must not retry automatically')
    assert.equal(toasts.at(-1)?.title, failure.data?.msg || failure.errMsg, 'the backend/network reason must survive the page catch')
    assert.equal(toasts.length, 1, 'one failed submission must show one error notification')
    assert.equal(changed, 0, 'failure must not emit success or close the task')
    assert.equal(reloads, 0, 'failure must not reload and erase unsaved input')
    assert.deepEqual(formData.value, { score: 35, comment: '已填写的评价' })
    assert.equal(submitting.value, false, 'failure must restore the retry button')
    assert.equal(submittingAction.value, '')
  }
  reply = { statusCode: 200, data: { code: 0, data: { instance: { id: 'instance-1' } } } }
  toasts.length = 0
  await submit(action)
  assert.equal(changed, 1, 'a later successful retry must still update the parent')
  assert.equal(reloads, 1)
  assert.equal(toasts[0]?.title, action === 'approve' ? '通过成功' : '提交办理成功')
}

// Existing API callers retain the shared interceptor's default feedback.
reply = { statusCode: 200, data: { code: 1, msg: '工作流任务已处理' } }
toasts.length = 0
await assert.rejects(api.completeWorkflowTask('task-1', { action: 'approve', formData: { score: 35 } }))
assert.equal(toasts[0]?.title, '工作流任务已处理')
console.log('workflow submission errors: reason, single toast, retained input, retry, success and default API feedback passed')
