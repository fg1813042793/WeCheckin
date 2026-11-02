import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const ts = require('typescript')

const read = path => readFileSync(new URL(`../src/${path}`, import.meta.url), 'utf8')
const compile = source => ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
function load(path, dependencies = {}) {
  const module = { exports: {} }
  runInThisContext(`(function(module,exports,require){${compile(read(path))}\n})`)(module, module.exports, name => dependencies[name] || require(name))
  return module.exports
}
function functionFromSFC(path, name, scope) {
  const source = read(path).match(/<script setup lang="ts">([\s\S]*?)<\/script>/)[1]
  const ast = ts.createSourceFile(path, source, ts.ScriptTarget.Latest, true)
  const declaration = ast.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === name)
  assert.ok(declaration, `${name} must exist`)
  return runInThisContext(`(function(${Object.keys(scope).join(',')}){${compile(declaration.getText(ast))}\nreturn ${name}})`)(...Object.values(scope))
}
const routes = load('pages/workflow/workflow-route-keys.ts')
const id = 'instance:测试 /1'
const normalKey = routes.workflowFormDetailContentKey(id)
const summaryKey = routes.workflowFormDetailContentKey(id, 'summary')
assert.notEqual(summaryKey, normalKey, 'summary and participant form detail tabs must not share an entry key')
for (const [key, source] of [[normalKey, 'instance'], [summaryKey, 'summary']]) {
  assert.equal(routes.workflowFormDetailInstanceIdFromContentKey(key), id)
  assert.equal(routes.workflowFormDetailSourceFromContentKey(key), source)
  assert.equal(routes.normalizeWorkflowDynamicContentKey(key), key, 'reload/deep-link normalization must preserve source')
}
assert.equal(normalKey, 'workflow:form-detail:instance%3A%E6%B5%8B%E8%AF%95%20%2F1', 'old URLs remain valid')
assert.equal(routes.workflowFormDetailInstanceIdFromContentKey('workflow:summary-form-detail:%E0%A4%A'), '')
assert.equal(routes.workflowFormDetailContentKey(' ', 'summary'), '')
assert.equal(routes.workflowFormDetailSourceFromContentKey(routes.workflowFormDetailContentKey('summary:one')), 'instance', 'an instance ID must not change the entry source')

for (const summaryMode of [true, false]) {
  const opened = []
  const popupVisible = { value: true }
  const open = functionFromSFC('pages/workflow/components/WorkflowDetailPanel.vue', 'openFormDetail', {
    props: { summaryMode },
    detail: { value: { instance: { id, definitionName: '测试流程' } } },
    showFormDetailAction: { value: true },
    title: { value: '测试流程' },
    popupVisible,
    workflowFormDetailContentKey: routes.workflowFormDetailContentKey,
    appContent: { openDynamicTab: tab => opened.push(tab) },
  })
  open()
  assert.equal(opened[0].key, summaryMode ? summaryKey : normalKey)
  assert.equal(new URL(opened[0].path, 'https://fixture.invalid').searchParams.get('view'), opened[0].key)
  assert.equal(popupVisible.value, false)
}

const calls = []
let denied = false
const api = load('api/workflow.ts', { '@/api/dingtalk-h5/base': {
  get: async (url) => {
    calls.push(url)
    if (denied)
      throw Object.assign(new Error('无权查看此汇总记录'), { msg: '无权查看此汇总记录' })
    return { data: { instance: { id }, form: [], formData: { score: 88 } } }
  },
} })
const { appErrorMessage } = load('common/app-error.ts')
for (const key of [normalKey, summaryKey]) {
  const detail = { value: null }
  const formData = { value: {} }
  const loading = { value: false }
  const loadError = { value: '' }
  const invoke = functionFromSFC('pages/workflow/components/WorkflowFormDetailPage.vue', 'loadDetail', {
    props: { contentKey: key },
    instanceId: { value: id },
    detail,
    formData,
    loading,
    loadError,
    workflowFormDetailSourceFromContentKey: routes.workflowFormDetailSourceFromContentKey,
    getWorkflowInstance: api.getWorkflowInstance,
    getWorkflowSummaryInstance: api.getWorkflowSummaryInstance,
    initialWorkflowFormData: (_fields, data) => ({ ...data }),
    appErrorMessage,
  })
  const expectedPath = `/api/v2/dingtalk/h5/workflows/${key === summaryKey ? 'summary/' : ''}instances/${encodeURIComponent(id)}`
  await invoke()
  await invoke()
  assert.ok(calls.slice(-2).every(path => path === expectedPath), 'refresh must use the same backend authorization entry')
  assert.equal(formData.value.score, 88)
  denied = true
  const count = calls.length
  await invoke()
  assert.equal(calls.length, count + 1, 'denial must not fall back to a different endpoint')
  assert.equal(detail.value, null, 'permission revocation must remove stale form detail')
  assert.equal(Object.keys(formData.value).length, 0)
  assert.equal(loadError.value, '无权查看此汇总记录')
  assert.equal(loading.value, false)
  denied = false
}
console.log('summary form detail: source-preserving links, old links, reload, endpoint choice, denial and stale-data removal passed')
