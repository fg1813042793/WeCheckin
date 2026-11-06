import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { runInNewContext } from 'node:vm'

const formFill = await import('../utils/formFill.js')
const currentDir = dirname(fileURLToPath(import.meta.url))

const {
  formatRemainingTime,
  getAnswerProgress,
  getQuestionInitialValue,
  isQuestionAnswered,
  parseQuestions,
  parseSettings,
  resolveSurveyTitle,
} = formFill

for (const [displayTitle, expected] of [
  ['公开问卷', '公开问卷'], ['  公开问卷  ', '公开问卷'],
  ['', '内部标题'], [' \n\t ', '内部标题'], [undefined, '内部标题'],
  [null, '内部标题'], [42, '内部标题'], [{ title: '错误数据' }, '内部标题'],
]) {
  assert.equal(resolveSurveyTitle('内部标题', displayTitle), expected)
}

assert.equal(typeof getAnswerProgress, 'function', 'getAnswerProgress should be exported')

assert.deepEqual(parseSettings('{"progressBar":true}'), { progressBar: true })
assert.deepEqual(parseSettings('{bad json'), {})
assert.deepEqual(parseQuestions('{"questions":[{"id":"q1","type":"radio"}]}'), [{ id: 'q1', type: 'radio' }])
assert.deepEqual(parseQuestions('{"items":[]}'), [])

assert.deepEqual(getQuestionInitialValue({ type: 'checkbox' }), [])
assert.deepEqual(getQuestionInitialValue({ type: 'matrixRadio' }), {})
assert.deepEqual(getQuestionInitialValue({ type: 'multiInput', props: { fields: [{}, {}] } }), ['', ''])

assert.equal(isQuestionAnswered({ type: 'checkbox' }, ['a']), true)
assert.equal(isQuestionAnswered({ type: 'checkbox' }, []), false)
assert.equal(isQuestionAnswered({ type: 'rating' }, 0), false)
assert.equal(isQuestionAnswered({ type: 'rating' }, 3), true)
assert.equal(isQuestionAnswered({ type: 'matrixCheckbox' }, { r1: ['a'] }), true)

const questions = [
  { id: 'layout', type: 'description' },
  { id: 'hidden', type: 'input' },
  { id: 'q1', type: 'radio' },
  { id: 'q2', type: 'checkbox' },
  { id: 'q3', type: 'rating' },
]
const answers = { q1: 'A', q2: [], q3: 4 }
assert.deepEqual(getAnswerProgress(questions, answers, { hiddenIds: ['hidden'] }), {
  answered: 2,
  total: 3,
  percent: 67,
})

assert.equal(formatRemainingTime(0), '已超时')
assert.equal(formatRemainingTime(61_000), '1:01')

const surveyFillPage = readFileSync(resolve(currentDir, '../pages/survey/fill.vue'), 'utf8')
assert.ok(!surveyFillPage.includes('$set'), 'survey fill must use Vue 3 reactive property assignment')
assert.equal((surveyFillPage.match(/\bgoNext\(\)\s*\{/g) || []).length, 1, 'survey fill must define goNext once')
for (const required of [
  `:class="{ 'survey--with-progress': settings.progressBar && !loading && survey }"`,
  'top: var(--window-top, 0px)',
  '.survey--with-progress',
  'class="survey-leave-dialog"',
  'onBackPress()',
  'shouldConfirmTimedLeave()',
  'submitFromBackPrompt()',
  'abandonTimedSurvey()',
  '放弃填写',
  'if (this.abandoning) return',
]) {
  assert.ok(surveyFillPage.includes(required), `survey fill progress layout missing: ${required}`)
}

const examFillPage = readFileSync(resolve(currentDir, '../pages/exam/fill.vue'), 'utf8')
assert.ok(!examFillPage.includes('$set'), 'exam fill must use Vue 3 reactive property assignment')

const examListPage = readFileSync(resolve(currentDir, '../pages/exam/index.vue'), 'utf8')
for (const [name, source] of [['fill', examFillPage], ['list', examListPage]]) {
  assert.match(source, /import\s*\{[^}]*resolveSurveyTitle\s+as\s+resolveExamTitle[^}]*\}\s*from\s*['"]\.\.\/\.\.\/utils\/formFill['"]/, `exam ${name} must reuse the survey title resolver`)
  const template = source.split('</template>')[0]
  assert.doesNotMatch(template, /JSON\.parse|parseSettings\(/, `exam ${name} must parse settings before rendering`)
  assert.doesNotMatch(template, /\b(?:exam|s)\.title\b/, `exam ${name} must not render the management title directly`)
}
for (const className of ['exam-login-desc', 'result-title', 'exam-title']) {
  assert.match(examFillPage, new RegExp(`<text\\b[^>]*class="${className}"[^>]*>\\s*{{\\s*displayTitle\\s*}}\\s*</text>`), `${className} must bind the display title`)
}
assert.match(examFillPage, /<text\b[^>]*class="exam-login-desc"[^>]*v-if="displayTitle"/, 'login title visibility must follow the resolved display title')
assert.match(examListPage, /<view\b[^>]*class="s-title"[^>]*>\s*{{\s*s\.displayTitle\s*}}\s*<\/view>/, 'exam cards must bind the prepared display title')

// Exercise the actual page options without starting uni-app or making network requests.
function readPageOptions(source, dependencies = {}) {
  const script = source.match(/<script>([\s\S]*?)<\/script>/)[1]
    .replace(/^import\b[\s\S]*?\sfrom\s+['"][^'"]+['"]\s*$/gm, '')
    .replace('export default', 'const page =')
  return runInNewContext(`${script}\npage`, {
    ...formFill,
    resolveExamTitle: resolveSurveyTitle,
    QuestionField: {},
    ...dependencies,
  })
}

const examTitleCases = [
  ['  考生显示名称 \n', '考生显示名称'],
  ['', '内部考试'], [' \n\t ', '内部考试'], [undefined, '内部考试'],
  [null, '内部考试'], [false, '内部考试'], [42, '内部考试'],
  [{ title: '非法名称' }, '内部考试'], [['非法名称'], '内部考试'],
].flatMap(([displayTitle, expected]) => [
  [{ displayTitle }, expected],
  [JSON.stringify({ displayTitle }), expected],
])
for (const settings of [undefined, null, '', '{bad json', 'null', '[]', '42', '"非法配置"']) {
  examTitleCases.push([settings, '内部考试'])
}

let resultSettings
const examFillOptions = readPageOptions(examFillPage, {
  examApi: {
    async getDetail() {
      return { code: 0, data: { exam: { title: '内部考试' }, settings: resultSettings } }
    },
  },
})
assert.equal(typeof examFillOptions.computed.displayTitle, 'function', 'exam fill must compute its display title')
const unscoredQuestions = [
  { id: 'contact', type: 'input', title: '联系方式', required: true, examScoreMode: 'none', examScore: 99, examCorrectAnswer: '旧答案' },
  { id: 'layout', type: 'description' },
  { id: 'graded', type: 'radio', examScoreMode: 'single', examScore: 5 },
  { id: 'legacy', type: 'input', examScore: 2 },
]
assert.equal(examFillOptions.computed.totalScore.call({ questions: unscoredQuestions }), 7, 'none must override residual score without changing legacy score fallback')
assert.equal(getAnswerProgress(unscoredQuestions, {}).total, 3, 'none remains in progress and navigation')
assert.equal(isQuestionAnswered(unscoredQuestions[0], ''), false)
assert.equal(isQuestionAnswered(unscoredQuestions[0], '用户输入'), true)

for (const [settings, expected] of examTitleCases) {
  const state = { exam: { title: '内部考试' }, settings: parseSettings(settings) }
  assert.equal(examFillOptions.computed.displayTitle.call(state), expected, 'exam fill title fallback')
  assert.equal(state.exam.title, '内部考试', 'exam fill must preserve the management title')

  resultSettings = settings
  const resultState = { ...examFillOptions.data(), examId: 1, $route: { query: { exSession: 'test-session' } } }
  await examFillOptions.methods.load.call(resultState)
  assert.equal(resultState.showResultView, true, 'result loading must still open the result view')
  assert.equal(examFillOptions.computed.displayTitle.call(resultState), expected, 'result loading must parse settings before resolving the title')
}
assert.equal(examFillOptions.computed.displayTitle.call({ exam: null, settings: {} }), '', 'unloaded exam has no title')
assert.equal(examFillOptions.computed.displayTitle.call({ exam: { title: '' }, settings: { displayTitle: ' 考生名称 ' } }), '考生名称', 'display title works without a management title')

const records = examTitleCases.map(([settings], id) => ({ id, title: '内部考试', settings }))
const originalRecords = structuredClone(records)
const listCalls = []
const examListOptions = readPageOptions(examListPage, {
  examApi: {
    async getList(params) {
      listCalls.push({ ...params })
      const list = params.page === 1 ? records : [{ id: records.length, title: '第二页考试', settings: '{bad json' }]
      return { data: { list, limits: params.page === 1 ? { 0: { deviceFull: true } } : {} } }
    },
  },
})
const listState = { ...examListOptions.data(), pageSize: records.length, isLogged: false, getDeviceId: () => 'test-device' }
await examListOptions.methods.load.call(listState)
assert.deepEqual(Array.from(listState.list, item => item.displayTitle), examTitleCases.map(([, expected]) => expected), 'exam list must resolve object and JSON settings during loading')
assert.deepEqual(Array.from(listState.list, item => item.title), records.map(item => item.title), 'exam list must retain management titles')
assert.deepEqual(records, originalRecords, 'exam list must not mutate API records')
assert.equal(listState.hasMore, true, 'a full page must still allow pagination')
const firstItem = listState.list[0]
listState.page = 2
await examListOptions.methods.load.call(listState)
assert.equal(listState.list[0], firstItem, 'pagination must preserve previous items')
assert.equal(listState.list.at(-1).displayTitle, '第二页考试', 'subsequent pages must resolve titles too')
assert.equal(listState.hasMore, false, 'a short page must still end pagination')
assert.equal(listState.limitsMap[0].deviceFull, true, 'pagination must retain existing limits')
assert.deepEqual(listCalls.map(params => params.page), [1, 2], 'title normalization must not change requested pages')

assert.ok(!/\.survey-progress\s*\{[^}]*top:\s*0\b/s.test(surveyFillPage), 'survey progress must not be fixed to viewport top')

const storedResult = { questionId: 3, correct: true, gotScore: 5, fullScore: 5 }
const resultState = { result: { results: [storedResult] } }
assert.equal(examFillOptions.methods.resultForQuestion.call(resultState, 0), null, 'unscored field must not borrow the next result')
assert.equal(examFillOptions.methods.resultForQuestion.call(resultState, 1), null, 'layout does not consume a result')
assert.equal(examFillOptions.methods.resultForQuestion.call(resultState, 2), storedResult, 'results use original schema ordinal')
assert.equal(examFillOptions.methods.examDisplayScore(unscoredQuestions[0]), 0)
assert.equal(examFillOptions.methods.examDisplayScore(unscoredQuestions[3]), 2)

// Run the real submission gate: an unscored required field still reaches server validation.
const validationCalls = [], modals = []
let modalCompletion, submissions = 0
const requiredOptions = readPageOptions(examFillPage, {
  examApi: { validate: async payload => { validationCalls.push(payload); return { data: { valid: false, errors: [{ message: '联系方式必填' }] } } } },
  uni: { showModal(options) { modals.push(options); if (options.success) modalCompletion = options.success({ confirm: true }) } },
})
const requiredState = {
  ...requiredOptions.data(), questions: unscoredQuestions, examId: 5, answers: { contact: '' },
  getDeviceUA: () => 'test', getDeviceId: () => 'device', doSubmit: async () => { submissions++ },
}
await requiredOptions.methods.onSubmit.call(requiredState)
await modalCompletion
assert.equal(validationCalls.length, 1)
assert.equal(validationCalls[0].answers, requiredState.answers, 'unscored answers must be submitted unchanged for validation')
assert.equal(submissions, 0, 'required validation failure must prevent submission')
assert.equal(requiredState.submitting, false)
assert.equal(modals.at(-1).content, '联系方式必填')
console.log('exam form runtime passed: none totals, ordinal results, legacy scores, progress and required validation gate')
