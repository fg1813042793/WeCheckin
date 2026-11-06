import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'
import './check-exam-record-user-info-runtime.mjs'

const examList = readFileSync(new URL('../src/views/exam/ExamList.vue', import.meta.url), 'utf8')
const responses = readFileSync(new URL('../src/views/exam/ExamResponses.vue', import.meta.url), 'utf8')
const api = readFileSync(new URL('../src/api/index.ts', import.meta.url), 'utf8')
const request = readFileSync(new URL('../src/utils/request.ts', import.meta.url), 'utf8')
const fillPC1 = readFileSync(new URL('../src/views/exam/ExamFillPC1.vue', import.meta.url), 'utf8')
const fillPC = readFileSync(new URL('../src/views/exam/ExamFillPC.vue', import.meta.url), 'utf8')
const plain = value => JSON.parse(JSON.stringify(value))
const completionRedirect = executeTypeScriptModule(
  readFileSync(new URL('../src/utils/completionRedirect.ts', import.meta.url), 'utf8'),
  'completionRedirect.ts',
)
assert.equal(typeof completionRedirect.scheduleCompletionRedirect, 'function')
const attemptSession = executeTypeScriptModule(
  readFileSync(new URL('../src/views/exam/examAttemptSession.ts', import.meta.url), 'utf8'),
  'examAttemptSession.ts',
)
const { shouldRenewPublicExamSession, examCountdownState } = attemptSession
const questionNavigation = executeTypeScriptModule(
  readFileSync(new URL('../src/views/exam/examQuestionNavigation.ts', import.meta.url), 'utf8'),
  'examQuestionNavigation.ts',
)
const { examQuestionNavIndex } = questionNavigation

const navigationQuestions = [
  { id: 'q1', type: 'input' },
  { id: 'divider-1', type: 'divider' },
  { id: 'hidden', type: 'textarea', defaultHidden: true },
  { id: 'q2', type: 'textarea' },
  { id: 'divider-2', type: 'divider' },
]
assert.equal(examQuestionNavIndex(navigationQuestions, 0), 0)
assert.equal(examQuestionNavIndex(navigationQuestions, 1), 0, '第一题后的布局页应保持 1 / N')
assert.equal(examQuestionNavIndex(navigationQuestions, 2), 0, '隐藏题不得改变导航题号')
assert.equal(examQuestionNavIndex(navigationQuestions, 3), 1)
assert.equal(examQuestionNavIndex(navigationQuestions, 4), 1, '末尾布局页应保持最后一道真实题号')
assert.equal(examQuestionNavIndex([{ id: 'divider', type: 'divider' }, { id: 'q1', type: 'input' }], 0), 0, '开头布局页不得显示 0 / N')
assert.equal(examQuestionNavIndex([{ id: 'divider', type: 'divider' }], 0), -1)

const stalePublicSession = {
  session: 'stale-session', startAt: 1_000, visibility: 0, maxAttempts: 0, endTime: 0,
  settings: { timeLimit: 0, maxSubmitMinutes: 50, deviceLimit: 0, ipLimit: 0 },
}
assert.equal(shouldRenewPublicExamSession(stalePublicSession, 'stale-session', 3_001_000), true)
assert.equal(shouldRenewPublicExamSession(stalePublicSession, 'stale-session', 3_000_999), false)
assert.equal(shouldRenewPublicExamSession({ ...stalePublicSession, visibility: 1 }, 'stale-session', 3_001_000), false)
assert.equal(shouldRenewPublicExamSession({ ...stalePublicSession, maxAttempts: 1 }, 'stale-session', 3_001_000), false)
assert.equal(shouldRenewPublicExamSession({ ...stalePublicSession, settings: { ...stalePublicSession.settings, deviceLimit: 1 } }, 'stale-session', 3_001_000), false)
assert.equal(shouldRenewPublicExamSession(stalePublicSession, '', 3_001_000), false)
assert.equal(shouldRenewPublicExamSession(stalePublicSession, 'other-session', 3_001_000), false)

assert.deepEqual(plain(examCountdownState({
  now: 3_001_000, startAt: 1_000, timeLimit: 0, maxSubmitMinutes: 50,
  endTime: 0, submitting: false, submitted: false,
})), { remaining: 0, shouldSchedule: false, shouldAutoSubmit: true })
assert.deepEqual(plain(examCountdownState({
  now: 3_001_000, startAt: 1_000, timeLimit: 0, maxSubmitMinutes: 50,
  endTime: 0, submitting: true, submitted: false,
})), { remaining: 0, shouldSchedule: false, shouldAutoSubmit: false })
assert.deepEqual(plain(examCountdownState({
  now: 2_000, startAt: 1_000, timeLimit: 0, maxSubmitMinutes: 0,
  endTime: 0, submitting: false, submitted: false,
})), { remaining: 0, shouldSchedule: false, shouldAutoSubmit: false })
assert.deepEqual(plain(examCountdownState({
  now: 2_000, startAt: 1_000, timeLimit: 1, maxSubmitMinutes: 0,
  endTime: 0, submitting: false, submitted: false,
})), { remaining: 59_000, shouldSchedule: true, shouldAutoSubmit: false })

for (const pattern of ['查看答卷', '统计', 'goResponses(row)', 'goStatistic(row)', 'DataAnalysis', '复制考试', 'command="copy"', "command === 'copy'", 'examCopy', "hasPerm('admin:menu:exam:copy')"]) {
  assert.ok(examList.includes(pattern), `exam list missing action contract: ${pattern}`)
}
for (const pattern of ['已交答卷', 'row.responseCount || 0', 'label="答卷数"']) {
  assert.ok(examList.includes(pattern), `exam list missing response count contract: ${pattern}`)
}
assert.ok(
  examList.includes('<div class="survey-card__meta"><span><strong>{{ row.responseCount || 0 }}</strong> 份答卷</span>'),
  'exam card footer must display the submitted answer count',
)
assert.doesNotMatch(
  examList,
  /survey-card__meta[^\n]*row\.maxResponse/,
  'exam card footer must not display the maximum response limit',
)
for (const pattern of ['导出 PDF', '导出 Word', '导出 Excel', 'examRecordExport', 'exportRecord(row', 'Download']) {
  assert.ok(responses.includes(pattern), `exam responses missing export contract: ${pattern}`)
}
for (const pattern of [
  'class="record-actions"',
  'label="用户信息"',
  'formatRecordUserInfo(row)',
  "const PERSONAL_INFO_TYPES = new Set(['name', 'phone', 'email', 'studentId', 'employeeId', 'class'])",
  'label="用时" width="150"',
  'label="开始时间" width="180"',
  'label="提交时间" width="180"',
  "row.status===1?'待批改'",
  'gradeState(row).label',
  "label: '不计分'",
  'isRecordQuestionGraded(question)',
]) {
  assert.ok(responses.includes(pattern), `exam responses layout/user-info contract missing: ${pattern}`)
}
for (const pattern of ['examRecordExport(', "format: 'pdf' | 'docx' | 'xlsx'", 'getBlob(', 'examCopy(']) {
  assert.ok(api.includes(pattern), `admin API missing exam record export contract: ${pattern}`)
}
assert.ok(request.includes('export async function getBlob'), 'request utility must support authenticated blob downloads')
for (const pattern of [
  "import { examCountdownState, shouldRenewPublicExamSession } from './examAttemptSession'",
  "import { examQuestionNavIndex } from './examQuestionNavigation'",
  'const currentNavIndex = computed(() => examQuestionNavIndex(questions.value, currentIndex.value))',
  'shouldRenewPublicExamSession(res.data, requestedSession)',
  'if (tick()) countdownTimer = setInterval(tick, 1000)',
  'if (submitting.value || submitted.value) return',
]) {
  assert.ok(fillPC1.includes(pattern), `ExamFillPC1 session recovery contract missing: ${pattern}`)
}
assert.doesNotMatch(fillPC1, /watch\(\(\) => settings\.value\.timeLimit/, 'ExamFillPC1 must not start the countdown twice during load')
for (const [source, name] of [[fillPC1, 'ExamFillPC1'], [fillPC, 'ExamFillPC']]) {
  for (const pattern of [
    "import { scheduleCompletionRedirect } from '../../utils/completionRedirect'",
    '考试提交成功',
    '2 秒后自动跳转',
    '立即前往',
    'submitted.value = true',
    'redirectTimer = scheduleCompletionRedirect',
    'if (redirectTimer) clearTimeout(redirectTimer)',
  ]) {
    assert.ok(source.includes(pattern), `${name} completion redirect contract missing: ${pattern}`)
  }
}

for (const pattern of [
  'box-sizing: border-box;',
  'border: 2px solid transparent;',
  '.exam-sheet-item--cur { border-color: #3873f6; font-weight: 600; }',
]) {
  assert.ok(fillPC1.includes(pattern), `ExamFillPC1 answer sheet sizing contract missing: ${pattern}`)
}
assert.doesNotMatch(fillPC1, /\.exam-sheet-item--cur\s*\{[^}]*border:\s*2px\s+solid/, 'ExamFillPC1 selected answer number must not change its box size')
for (const pattern of [
  'box-sizing:border-box;',
  'border:2px solid transparent;',
  '.answer-sheet-cur { border-color:#409eff; font-weight:600; }',
]) {
  assert.ok(fillPC.includes(pattern), `ExamFillPC answer sheet sizing contract missing: ${pattern}`)
}
assert.doesNotMatch(fillPC, /\.answer-sheet-cur\s*\{[^}]*border:\s*2px\s+solid/, 'ExamFillPC selected answer number must not change its box size')

console.log('exam management checks passed')
