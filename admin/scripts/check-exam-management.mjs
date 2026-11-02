import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const examList = readFileSync(new URL('../src/views/exam/ExamList.vue', import.meta.url), 'utf8')
const responses = readFileSync(new URL('../src/views/exam/ExamResponses.vue', import.meta.url), 'utf8')
const api = readFileSync(new URL('../src/api/index.ts', import.meta.url), 'utf8')
const request = readFileSync(new URL('../src/utils/request.ts', import.meta.url), 'utf8')

for (const pattern of ['查看答卷', '统计', 'goResponses(row)', 'goStatistic(row)', 'DataAnalysis', '复制考试', 'command="copy"', "command === 'copy'", 'examCopy', "hasPerm('admin:menu:exam:copy')"]) {
  assert.ok(examList.includes(pattern), `exam list missing action contract: ${pattern}`)
}
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

console.log('exam management checks passed')
