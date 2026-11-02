import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { runInThisContext } from 'node:vm'

const require = createRequire(import.meta.url)
const ts = require('typescript')

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')
function compile(source) {
  return ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
}
const apiModule = { exports: {} }
runInThisContext(`(function(module,exports,require){${compile(read('src/api/workflow.ts'))}\n})`)(apiModule, apiModule.exports, (request) => {
  assert.equal(request, '@/api/dingtalk-h5/base')
  return {
    authToken: () => 'fixture-only',
    buildApiUrl: (path, params) => `https://fixture.invalid${path}?${new URLSearchParams(params)}`,
  }
})
const { workflowSummaryExportUrl } = apiModule.exports
for (const format of ['xlsx', 'pdf', 'docx']) {
  const full = new URL(workflowSummaryExportUrl(['instance-a'], format))
  assert.equal(full.searchParams.get('content'), 'full', 'existing export defaults to full content')
  const form = new URL(workflowSummaryExportUrl(['instance-a', 'instance-b'], format, 'form'))
  assert.equal(form.searchParams.get('content'), 'form', 'form-only choice must reach the backend')
  assert.equal(form.searchParams.get('format'), format)
  assert.equal(form.searchParams.get('instanceIds'), 'instance-a,instance-b')
  assert.equal(form.pathname, '/api/v2/dingtalk/h5/workflows/summary/export')
}

// Execute the real click handler with only the browser/store boundaries replaced.
const source = read('src/pages/workflow/components/WorkflowSummarySection.vue').match(/<script setup lang="ts">([\s\S]*?)<\/script>/)[1]
const ast = ts.createSourceFile('summary.ts', source, ts.ScriptTarget.Latest, true)
const handler = ast.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === 'exportInstances')
assert.ok(handler, 'shared single and batch export handler exists')
const canExport = { value: true }
const exportFormat = { value: 'xlsx' }
const opens = []
const messages = []
const invoke = runInThisContext(`(function(canExport,exportFormat,workflowSummaryExportUrl,window,uni){${compile(handler.getText(ast))}\nreturn exportInstances})`)(
  canExport,
  exportFormat,
  workflowSummaryExportUrl,
  { open: (url, target) => opens.push({ url: new URL(url), target }) },
  { showToast: value => messages.push(value) },
)
canExport.value = false
invoke(['instance-a'], 'form')
assert.equal(opens.length, 0, 'direct calls cannot bypass the permission guard')
canExport.value = true
invoke([], 'form')
assert.equal(opens.length, 0, 'empty selection cannot start a download')
assert.equal(messages.length, 2)
for (const format of ['xlsx', 'pdf', 'docx']) {
  exportFormat.value = format
  invoke(['instance-a'], 'form')
  invoke(['instance-a', 'instance-b'], 'form')
  invoke(['instance-a'])
  const last = opens.slice(-3)
  assert.deepEqual(last.map(item => item.url.searchParams.get('content')), ['form', 'form', 'full'])
  assert.deepEqual(last.map(item => item.url.searchParams.get('instanceIds')), ['instance-a', 'instance-a,instance-b', 'instance-a'])
  assert.ok(last.every(item => item.url.searchParams.get('format') === format && item.target === '_blank'))
}
console.log('workflow summary export: formats, single/batch content, legacy default, permission and empty-selection guards passed')
