import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const formFill = await import('../utils/formFill.js')
const currentDir = dirname(fileURLToPath(import.meta.url))

const {
  formatRemainingTime,
  getAnswerProgress,
  getQuestionInitialValue,
  isQuestionAnswered,
  parseQuestions,
  parseSettings,
} = formFill

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

assert.ok(!/\.survey-progress\s*\{[^}]*top:\s*0\b/s.test(surveyFillPage), 'survey progress must not be fixed to viewport top')
