import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import * as vue from 'vue'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const read = file => readFileSync(new URL(`../src/${file}`, import.meta.url), 'utf8')
const helperPath = new URL('../src/router/useFormDesignerTab.ts', import.meta.url)
assert.ok(existsSync(helperPath), '问卷和考试设计器需要独立页签生命周期')
const tabs = executeTypeScriptModule(read('router/routeTabTitles.ts'), 'routeTabTitles.ts', { vue, 'vue-router': {} })
const plain = value => JSON.parse(JSON.stringify(value))
const settle = async () => { for (let i = 0; i < 6; i++) await Promise.resolve() }
for (const kind of ['survey', 'exam']) {
  for (const state of ['active', 'background', 'aborted', 'disposed']) {
    const originalPath = `/${kind}/designer`
    const route = { path: originalPath, fullPath: originalPath, query: {} }
    const current = { value: { ...route } }
    const updated = [], replaced = [], navigated = []
    let dispose
    const router = {
      currentRoute: current,
      resolve: target => ({ ...target, fullPath: `${target.path}?id=${target.query.id}`, name: kind, meta: {} }),
      replace: async (target) => { navigated.push(target); return state === 'aborted' ? { type: 4 } : undefined },
    }
    const { useFormDesignerTab } = executeTypeScriptModule(read('router/useFormDesignerTab.ts'), 'useFormDesignerTab.ts', {
      'vue': { inject: key => key === tabs.ADMIN_ROUTE_TAB_TITLE_KEY ? (...args) => updated.push(args) : (...args) => replaced.push(args), onScopeDispose: fn => { dispose = fn } },
      'vue-router': { useRoute: () => route, useRouter: () => router },
      'element-plus': { ElMessage: { warning() {} } },
      './routeTabTitles': tabs,
    })
    const ready = useFormDesignerTab(kind)
    if (state === 'background') {
      current.value = { path: '/other', fullPath: '/other', query: {} }
      Object.assign(route, current.value)
    }
    if (state === 'disposed') dispose()
    ready(12, '管理名称', true)
    await settle()
    if (state === 'disposed') {
      assert.equal(updated.length + replaced.length + navigated.length, 0)
      continue
    }
    assert.equal(updated[0][0], originalPath, '迟到保存只能更新自身页签')
    if (state === 'background') assert.equal(navigated.length, 0, '后台新建完成不得抢占其他页签')
    else assert.equal(navigated.length, 1)
    if (state === 'aborted') assert.equal(replaced.length, 0, '取消跳转时不能删除原页签')
    else {
      assert.equal(replaced[0][0], originalPath)
      assert.equal(replaced[0][1].fullPath, `${originalPath}?id=12`)
      assert.ok(replaced[0][2].includes('管理名称'))
    }
  }

  const calls = [], readyCalls = [], saveCalls = []
  let completeSave
  const api = {
    surveyDetail: async id => { calls.push(id); return { data: { survey: { id, title: '内部管理名', settings: '{"displayTitle":"公开显示名"}' }, schema: '{"questions":[]}' } } },
    examDetail: async id => api.surveyDetail(id),
    surveyEdit: async payload => { saveCalls.push(plain(payload)); return new Promise(resolve => { completeSave = resolve }) },
    examSave: async payload => api.surveyEdit(payload),
  }
  const name = kind === 'survey' ? 'Survey' : 'Exam'
  const documentModule = executeTypeScriptModule(read(`views/${kind}/composables/use${name}DesignerDocument.ts`), 'document.ts', {
    'vue': { ...vue, watch() {}, onMounted() {}, onBeforeUnmount() {} },
    'vue-router': { onBeforeRouteLeave() {} },
    'element-plus': { ElMessage: { success() {}, warning() {}, error(message) { throw new Error(message) } }, ElMessageBox: {} },
    '../../../api': { adminApi: api },
    '../../../utils/request': { showRequestError(error) { throw error } },
    '../../../utils/surveyKingBridge': { normalizeQuestions: questions => questions },
    './useSurveyDesignerHistory': { useSurveyDesignerHistory: () => ({ canUndo: vue.ref(false), canRedo: vue.ref(false), reset() {}, undo() {}, redo() {} }) },
  })
  const form = documentModule[`create${name}DesignerForm`]()
  const route = { query: { id: 5 } }
  const document = documentModule[`use${name}DesignerDocument`]({
    form, route, router: { replace() {} }, questions: vue.ref([]), selected: vue.ref(null), logicRules: vue.ref([]),
    completionAction: vue.ref('default'), issues: vue.ref([]), locateIssue() {}, syncQuestionIds() {},
    onQuestionsLoaded() {}, reloadResources() {}, reloadSettings() {},
    onDocumentReady: (id, title, created = false) => readyCalls.push({ id, title, created }),
  })
  route.query.id = 99
  await document.load()
  assert.equal(calls[0], 5, '加载与重试必须绑定本页打开时的ID')
  assert.deepEqual(readyCalls.at(-1), { id: 5, title: '内部管理名', created: false })
  form.title = '已提交的改名'
  const saving = document.save(false)
  form.title = '请求期间尚未保存的名字'
  completeSave({ data: { id: 5 } })
  assert.equal(await saving, true)
  assert.equal(saveCalls.at(-1).id, 5)
  assert.equal(readyCalls.at(-1).title, '已提交的改名', '页签只采用本次成功保存的标题')
}
console.log('form designer tabs passed: survey/exam load, saved titles, stable IDs, creation replacement, background completion and disposal')
