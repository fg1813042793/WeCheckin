import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import * as vue from 'vue'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

function loadModule(path, modules = {}) {
  const url = new URL(`../src/${path}`, import.meta.url)
  return executeTypeScriptModule(readFileSync(url, 'utf8'), url.pathname, modules)
}

const { resolveSurveyTitle } = loadModule('utils/surveyTitle.ts')
for (const [displayTitle, expected] of [
  ['公开问卷', '公开问卷'], ['  公开问卷  ', '公开问卷'],
  ['', '内部标题'], [' \n\t ', '内部标题'], [undefined, '内部标题'],
  [null, '内部标题'], [42, '内部标题'], [{ title: '错误数据' }, '内部标题'],
]) {
  assert.equal(resolveSurveyTitle('内部标题', displayTitle), expected)
}

// Exercise the real designer save/load path with an in-memory API boundary.
let stored
const api = loadModule('api/index.ts', {
  '../utils/request': { __esModule: true, default: {
    async post(url, payload) {
      assert.equal(url, '/api/v2/admin/surveys')
      stored = { ...JSON.parse(JSON.stringify(payload)), id: 901 }
      return { data: { id: 901 } }
    },
    async put(url, payload) {
      assert.equal(url, '/api/v2/admin/surveys/901')
      stored = JSON.parse(JSON.stringify(payload))
      return { data: null }
    },
    async get(url) {
      assert.equal(url, '/api/v2/admin/surveys/901')
      return { data: { survey: stored, schema: stored.schema } }
    },
  } },
})
const history = loadModule('views/survey/composables/useSurveyDesignerHistory.ts', { vue })
const document = loadModule('views/survey/composables/useSurveyDesignerDocument.ts', {
  vue: { ...vue, watch: () => {}, onMounted: () => {}, onBeforeUnmount: () => {} },
  'element-plus': { ElMessage: { warning: message => assert.fail(message), success: () => {} } },
  'vue-router': { onBeforeRouteLeave: () => {} },
  '../../../api': api,
  '../../../utils/request': { showRequestError: error => { throw error } },
  '../../../utils/surveyKingBridge': { normalizeQuestions: questions => questions },
  './useSurveyDesignerHistory': history,
})

function createEditor(id) {
  const form = document.createSurveyDesignerForm()
  const route = { query: id ? { id } : {} }
  const editor = document.useSurveyDesignerDocument({
    form, route, router: { replace: ({ query }) => { route.query = query } },
    questions: vue.ref([]), selected: vue.ref(null), logicRules: vue.ref([]),
    completionAction: vue.ref('default'), issues: vue.ref([]),
    locateIssue: () => {}, syncQuestionIds: () => {}, reloadResources: () => {},
  })
  return { form, editor }
}

const original = createEditor()
await original.editor.load()
original.form.title = '内部标题'
original.form.displayTitle = '  公开问卷  '
original.form.publicUrl = '  https://public.example.com/sf1/901  '
assert.equal(await original.editor.save(false), true)
assert.equal(stored.title, '内部标题')
assert.equal(JSON.parse(stored.settings).displayTitle, '公开问卷')
assert.equal(JSON.parse(stored.settings).publicUrl, 'https://public.example.com/sf1/901')

const reopened = createEditor('901')
await reopened.editor.load()
assert.equal(reopened.form.displayTitle, '公开问卷')
assert.equal(reopened.form.publicUrl, 'https://public.example.com/sf1/901')
reopened.form.description = '修改其他配置'
assert.equal(await reopened.editor.save(false), true)
assert.equal(JSON.parse(stored.settings).displayTitle, '公开问卷')
assert.equal(JSON.parse(stored.settings).publicUrl, 'https://public.example.com/sf1/901')
reopened.form.displayTitle = ''
assert.equal(await reopened.editor.save(false), true)
assert.equal(JSON.parse(stored.settings).displayTitle, '')
assert.equal(JSON.parse(stored.settings).publicUrl, 'https://public.example.com/sf1/901')
assert.equal(stored.title, '内部标题')

for (const settings of ['{}', '', '{invalid', 'null', '{"displayTitle":123}']) {
  stored.settings = settings
  await reopened.editor.load()
  assert.equal(reopened.form.displayTitle, '', 'old/invalid settings must reset display title')
  assert.equal(reopened.form.publicUrl, '', 'old/invalid settings must reset public URL')
  assert.equal(resolveSurveyTitle(reopened.form.title, reopened.form.displayTitle), '内部标题')
}
console.log('Survey display title and public URL save/reload checks passed')
