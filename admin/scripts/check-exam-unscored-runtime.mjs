import assert from 'node:assert/strict'
import { readFileSync, existsSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createRequire } from 'node:module'
import { parse, compileScript, compileTemplate } from '@vue/compiler-sfc'
import ts from 'typescript'
import * as vue from 'vue'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../src/views/exam')
const require = createRequire(import.meta.url)
const plain = value => JSON.parse(JSON.stringify(value))
const saved = []
let response = { data: { survey: { id: 5, title: '不计分回归' }, schema: '{"questions":[]}' } }
const messages = { success() {}, warning() {}, error(message) { throw new Error(message) } }
const api = { examDetail: async () => response, examSave: async payload => { saved.push(plain(payload)); return { data: { id: 5 } } }, formkitTypes: async () => ({ data: [] }) }
const dependencies = {
  vue,
  'element-plus': { ElMessage: messages, ElMessageBox: {} },
  'vue-router': { useRoute: () => ({ query: { id: 5 } }), useRouter: () => ({ replace() {} }) },
}
const cache = new Map()
function load(file) {
  if (cache.has(file)) return cache.get(file)
  let source = readFileSync(file, 'utf8')
  if (file.endsWith('.vue')) source = compileScript(parse(source, { filename: file }).descriptor, { id: file }).content
  const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, esModuleInterop: true } }).outputText
  const module = { exports: {} }
  new Function('require', 'module', 'exports', compiled)(name => {
    if (dependencies[name]) return dependencies[name]
    if (name.endsWith('.vue')) return {}
    if (name.endsWith('/api')) return { adminApi: api }
    if (name.endsWith('/utils/request')) return { showRequestError: error => { throw error } }
    if (name.endsWith('survey-designer-question-types')) return { FALLBACK_TYPES: [], getSurveyQuestionTypeName: type => type }
    const local = resolve(dirname(file), `${name}.ts`)
    if (name.startsWith('.') && existsSync(local)) return load(local)
    return require(name)
  }, module, module.exports)
  cache.set(file, module.exports)
  return module.exports
}
const node = () => ({ children: [] })
const renderer = vue.createRenderer({
  createElement: node, createText: node, createComment: node,
  setText() {}, setElementText() {}, patchProp() {}, parentNode: () => null, nextSibling: () => null,
  insert(child, parent) { parent.children.push(child) }, remove() {},
})
const component = load(resolve(root, 'ExamDesigner.vue')).default
let state
const setup = component.setup
component.setup = (props, ctx) => { const result = setup(props, ctx); state = vue.proxyRefs(result); return result }
component.render = () => null
const app = renderer.createApp(component)
app.config.errorHandler = error => { throw error }
app.mount(node())
const settle = async () => { for (let i = 0; i < 6; i++) await vue.nextTick() }
try {
  await settle()
  const question = vue.reactive({ id: 'q1', type: 'multiInput', title: '必填但不计分', required: true, examScoreMode: 'manual', examScore: 12, examCorrectAnswer: '旧答案', props: { fields: [{ label: '填空', examScore: 8, examCorrect: true, examCorrectAnswer: '旧字段答案' }], options: [{ value: 'A', examScore: 9, examCorrect: true }] } })
  state.questions = [question]
  state.selected = question
  await settle()
  assert.equal(state.paperTotalScore, 12)
  question.examScoreMode = 'none'
  await settle()
  assert.equal(question.examScore, 0, 'manual -> none must reset the question score, including old-client reduce')
  assert.equal(state.paperTotalScore, 0)
  assert.equal(state.partialTotalScore, 0, 'stale field scores must not affect none mode')
  assert.equal(state.perOptionTotalScore, 0)
  assert.equal(question.examCorrectAnswer, undefined)
  question.props.fields[0].examScore = 50
  question.props.fields[0].examCorrectAnswer = '延迟更新'
  question.props.options[0].examScore = 50
  await settle()
  assert.equal(question.examScore, 0, 'option/field watchers must not resurrect score')
  assert.equal(question.examCorrectAnswer, undefined, 'field watchers must not resurrect correct answer')
  assert.equal(state.paperTotalScore, 0)
  state.populateFieldsFromCorrectAnswer()
  state.syncFieldsCorrectAnswer()
  assert.equal(question.examCorrectAnswer, undefined)
  question.examScoreMode = 'manual'
  await settle()
  question.examScore = 7
  question.examScoreMode = ''
  await settle()
  assert.equal(question.examScoreMode, 'none', 'clear must persist explicit none instead of legacy empty')
  assert.equal(question.examScore, 0)
  const legacy = { id: 'legacy', type: 'input', title: '历史空模式', required: false, examScore: 4, props: {} }
  state.questions.push(legacy)
  assert.equal(load(resolve(root, 'examScore.ts')).questionScore(legacy), 0, 'designer legacy empty mode behavior stays unchanged')
  assert.equal(await state.save(false), true)
  const snapshot = JSON.parse(saved.at(-1).schema)
  assert.equal(snapshot.questions[0].examScoreMode, 'none')
  assert.equal(snapshot.questions[0].required, true)
  assert.equal(snapshot.questions[1].examScoreMode, undefined)
  response = { data: { survey: { id: 5, title: '不计分回归' }, schema: saved.at(-1).schema } }
  await state.load()
  await settle()
  assert.equal(state.questions[0].examScoreMode, 'none', 'save/reload must preserve explicit none')
  assert.equal(state.questions[0].examScore, 0)
  const scoring = load(resolve(root, 'examScore.ts'))
  for (const type of ['radio', 'multiInput']) {
    for (const mode of ['single', 'manual', 'perOption', 'allCorrect', 'partialCorrect']) {
      const item = { type, examScoreMode: mode, examScore: 7, props: { options: [{ examScore: 3, examCorrect: true }, { examScore: 4 }], fields: [{ examScore: 3, examCorrect: true }, { examScore: 4 }] } }
      assert.equal(scoring.questionScore(item), mode === 'partialCorrect' ? 3 : 7, `legacy ${mode} score stays unchanged`)
      item.examScoreMode = 'none'
      assert.equal(scoring.questionScore(item), 0, `${type}: none overrides every score source`)
      assert.equal(scoring.examDisplayScore(item), 0)
    }
  }
  state.prepareJson()
  const exported = JSON.parse(state.exportedJson)
  assert.equal(exported.questions[0].examScoreMode, 'none', 'JSON export retains the contract')
  const originalDocument = globalThis.document
  let fileInput
  try {
    globalThis.document = { createElement() { fileInput = { click() {} }; return fileInput } }
    state.loadJson()
    await fileInput.onchange({ target: { files: [{ text: async () => JSON.stringify(exported) }] } })
    assert.equal(state.questions[0].examScoreMode, 'none', 'JSON import retains explicit none')
    state.selected = state.questions[0]
    await settle()
    assert.equal(state.paperTotalScore, 0)
    assert.equal(state.questions[0].examCorrectAnswer, undefined)
  } finally { globalThis.document = originalDocument }
} finally {
  app.unmount()
}

// Exercise actual PC page computed state and title generation without network lifecycles.
dependencies.vue = { ...vue, onMounted() {}, onUnmounted() {}, onBeforeUnmount() {}, onActivated() {} }
dependencies['html5-qrcode'] = {}
dependencies['@vueup/vue-quill'] = {}
dependencies['@vueup/vue-quill/dist/vue-quill.snow.css'] = {}
dependencies['../../api/public'] = { publicFormApi: {} }
for (const file of ['ExamFillPC.vue', 'ExamFillPC1.vue']) {
  const page = load(resolve(root, file)).default
  let pageState
  const pageSetup = page.setup
  page.setup = (props, ctx) => { const result = pageSetup(props, ctx); pageState = vue.proxyRefs(result); return result }
  page.render = () => null
  const mounted = renderer.createApp(page)
  mounted.mount(node())
  try {
    const unscored = { id: 'none', type: 'input', title: '不计分题', required: true, examScoreMode: 'none', examScore: 99, props: {} }
    pageState.questions = [unscored, { id: 'legacy', type: 'input', examScore: 4 }]
    pageState.exam = { showScore: 1 }
    assert.equal(pageState.totalScore, 4, `${file}: residual score must be excluded from total`)
    assert.equal(pageState.realQuestions.length, 2, `${file}: none question remains answerable`)
    if (file === 'ExamFillPC.vue') assert.doesNotMatch(pageState.questionTitleHtml(unscored, 0), /99分/)
    pageState.handleSubmitSuccess({ record: { score: 4, totalScore: 4 }, results: [{ questionId: 2, correct: true }] })
    assert.equal(pageState.result.score, 4, `${file}: submission must read the stored record score`)
    assert.equal(pageState.result.correctCnt, 1, `${file}: submission must count actual graded results`)
    assert.equal(pageState.result.fullScore, 4)
    pageState.handleSubmitSuccess({ record: { score: 0, totalScore: 0 }, results: [] })
    assert.equal(pageState.result.correctCnt, 0)
    assert.deepEqual(plain(pageState.result.results), [])

  } finally { mounted.unmount() }
}

dependencies['@element-plus/icons-vue'] = {}
dependencies['@/api/examGrading'] = { hasExamGradingPermission: () => true }
const responses = load(resolve(root, 'ExamResponses.vue')).default
let rowsState
const rowsSetup = responses.setup
responses.setup = (props, ctx) => { const result = rowsSetup(props, ctx); rowsState = vue.proxyRefs(result); return result }
responses.render = () => null
const rowsApp = renderer.createApp(responses)
rowsApp.mount(node())
const originalDocument = globalThis.document
try {
  globalThis.document = { createElement() { return { innerHTML: '', get textContent() { return this.innerHTML.replace(/<[^>]*>/g, '') } } } }
  const questions = [
    { id: 'name', type: 'name', title: '姓名' },
    { id: 'layout', type: 'description' },
    { id: 'q3', type: 'input', title: '不计分字段', examScoreMode: 'none', examScore: 99 },
    { id: 'legacy', type: 'input', title: '旧零分字段' },
  ]
  Object.assign(rowsState.detailDialog, { record: { status: 2 }, schema: { questions }, answers: { name: '甲', q3: '完整保留的答案', legacy: '乙' }, scoring: { name: false, legacy: false }, results: [] })
  assert.equal(rowsState.answerRows[1].value, '完整保留的答案')
  assert.equal(rowsState.gradeState(rowsState.answerRows[1]).label, '不计分')
  assert.equal(rowsState.gradeState(rowsState.answerRows[0]).label, '不计分', 'legacy personal info rule unchanged')
  assert.equal(rowsState.gradeState(rowsState.answerRows[2]).label, '不计分', 'legacy zero-score rule unchanged')
  for (const [grade, expected] of [
    [{ questionId: 3, correct: false, gotScore: 3, fullScore: 10 }, '部分得分'],
    [{ questionId: 3, correct: false, gotScore: 0, fullScore: 10 }, '错误'],
    [{ questionId: 3, correct: true, gotScore: 10, fullScore: 10 }, '正确'],
  ]) {
    rowsState.detailDialog.results = [grade]
    assert.equal(rowsState.gradeState(rowsState.answerRows[1]).label, expected, 'stored grade wins over later none and uses original schema ordinal')
    assert.deepEqual(plain(rowsState.answerRows[1].grade), grade)
  }
  rowsState.detailDialog.results = []
  rowsState.detailDialog.scoring.q3 = false
  assert.equal(rowsState.gradeState(rowsState.answerRows[1]).label, '错误', 'stored legacy scoring also survives a later none definition')
} finally { globalThis.document = originalDocument; rowsApp.unmount() }
console.log('exam unscored runtime passed: designer watchers, clear, stale fields/options, PC totals/titles, save/reload, JSON transfer and historical grades')

// Compile the real statistics element from each template, including the uni-app page.
for (const file of [resolve(root, 'ExamFillPC.vue'), resolve(root, 'ExamFillPC1.vue'), resolve(root, '../../../../frontend/pages/exam/fill.vue')]) {
  const source = readFileSync(file, 'utf8')
  const element = source.match(/<(el-tag|text)\b[^>]*>正确 \{\{ result\.correctCnt \}\}[^<]*<\/\1>/)?.[0]
  assert.ok(element, `${file}: locate the real result statistics element`)
  const template = compileTemplate({ source: element, filename: file, id: 'result-stats', compilerOptions: { isCustomElement: () => true } })
  assert.deepEqual(template.errors, [])
  const module = { exports: {} }
  const compiled = ts.transpileModule(template.code, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
  new Function('require', 'module', 'exports', compiled)(() => vue, module, module.exports)
  const empty = module.exports.render({ result: { correctCnt: 0, results: [] } }, [])
  assert.equal(empty.type, vue.Comment, `${file}: all-none results must hide 0/0 correct counts`)
  const scored = module.exports.render({ result: { correctCnt: 1, results: [{ questionId: 3, correct: true }] } }, [])
  assert.match(scored.children, /正确 1\/1/, `${file}: mixed results use the actual result count`)
}
console.log('exam result templates passed: all-none hides correct counts; mixed results show the actual denominator')
