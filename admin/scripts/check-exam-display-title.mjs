import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import ts from 'typescript'

const require = createRequire(import.meta.url)
const vue = require('vue')
const cleanup = []
const saved = []
let response = { data: { survey: { id: 5, title: '内部管理名称', settings: '{"displayTitle":"考生展示名称"}' }, schema: '{"questions":[]}' } }
const source = readFileSync(new URL('../src/views/exam/composables/useExamDesignerDocument.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
const module = { exports: {} }
const deps = {
  'vue': { ...vue, onBeforeUnmount: callback => cleanup.push(callback) },
  'element-plus': { ElMessage: { success() {}, error() {}, warning() {} } },
  '../../../api': { adminApi: { examDetail: async () => response, examSave: async (payload) => { saved.push(payload); return { data: { id: 5 } } } } },
  '../../../utils/request': { showRequestError() {} },
}
new Function('require', 'module', 'exports', compiled)(name => deps[name] || require(name), module, module.exports)
const { createExamDesignerForm, useExamDesignerDocument } = module.exports
const scope = vue.effectScope()
const form = createExamDesignerForm()
assert.equal(form.displayTitle, '', 'new exams must start with an empty optional display name')
const document = scope.run(() => useExamDesignerDocument({
  form, questions: vue.ref([]), selected: vue.ref(null), logicRules: vue.ref([]),
  route: { query: { id: 5 } }, router: { replace() {} },
  onQuestionsLoaded() {}, reloadResources() {}, reloadSettings() {},
}))
try {
  await document.load()
  assert.equal(form.title, '内部管理名称')
  assert.equal(form.displayTitle, '考生展示名称')
  form.displayTitle = '  新显示名称  '
  assert.equal(await document.save(false), true)
  assert.equal(saved.at(-1).title, '内部管理名称', 'saving a display name must not replace the management name')
  assert.equal(JSON.parse(saved.at(-1).settings).displayTitle, '新显示名称')
  response = { data: { survey: { id: 5, title: saved.at(-1).title, settings: saved.at(-1).settings }, schema: saved.at(-1).schema } }
  await document.load()
  assert.equal(form.displayTitle, '新显示名称', 'save and reload must preserve the display name')
  for (const settings of ['', '{}', '{"displayTitle":null}', '{"displayTitle":42}', 'invalid']) {
    form.displayTitle = '上一个考试的显示名称'
    response = { data: { survey: { id: 5, title: '历史考试', settings }, schema: '{"questions":[]}' } }
    await document.load()
    assert.equal(form.displayTitle, '', `legacy settings must not retain a previous name: ${settings}`)
  }
  let scheduled
  const originalSetTimeout = globalThis.setTimeout
  const originalClearTimeout = globalThis.clearTimeout
  try {
    globalThis.setTimeout = (callback, delay) => { assert.equal(delay, 500); scheduled = callback; return 123 }
    globalThis.clearTimeout = () => {}
    await vue.nextTick()
    scheduled = null
    form.displayTitle = '自动保存显示名称'
    await vue.nextTick()
    assert.equal(typeof scheduled, 'function', 'editing only the display name must trigger autosave')
    scheduled()
    await vue.nextTick()
    assert.equal(JSON.parse(saved.at(-1).settings).displayTitle, '自动保存显示名称')
  } finally {
    globalThis.setTimeout = originalSetTimeout
    globalThis.clearTimeout = originalClearTimeout
  }
  form.displayTitle = '   '
  await document.save(false)
  assert.equal(JSON.parse(saved.at(-1).settings).displayTitle, '')
} finally {
  cleanup.forEach(callback => callback())
  scope.stop()
}
console.log('Exam display name: default, load/save, autosave, clearing and legacy compatibility passed')
