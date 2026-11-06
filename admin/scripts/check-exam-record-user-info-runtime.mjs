import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { parse, compileScript } from '@vue/compiler-sfc'
import * as vue from 'vue'
import ts from 'typescript'

const filename = 'ExamResponses.vue'
const source = readFileSync(new URL('../src/views/exam/ExamResponses.vue', import.meta.url), 'utf8')
const script = compileScript(parse(source, { filename }).descriptor, { id: filename }).content
const questions = [
  { id: 'q-name', type: 'name', title: '<b>考生姓名</b>' },
  { id: 'q-phone', type: 'phone', title: '手机' },
  { id: 'q-email', type: 'email', title: '邮箱' },
  { id: 'q-student', type: 'studentId', title: '学号' },
  { id: 'q-employee', type: 'employeeId', title: '工号' },
  { id: 'q-class', type: 'class', title: '' },
  { id: 'q-answer', type: 'textarea', title: '普通答案' },
]
const answers = {
  'q-name': '测试考生', 'q-phone': '13800000000', 'q-email': 'exam@example.com',
  'q-student': 'S001', 'q-employee': 'E001', 'q-class': '一班', 'q-answer': '普通答案不应出现在用户信息中',
}
const record = { id: 22, userId: '', answers: JSON.stringify(answers) }
// Match the backend examDetailResponse: schema is a sibling of survey.
let detail = { survey: { id: 5, title: '用户信息回归' }, responseCount: 1, schema: JSON.stringify({ questions }) }
const dependencies = {
  vue: { ...vue, onMounted() {}, onActivated() {} },
  'vue-router': { useRoute: () => ({ query: { examId: 5 }, fullPath: '/exam/responses?examId=5' }), useRouter: () => ({}) },
  'element-plus': { ElMessage: { error(message) { throw new Error(message) } }, ElMessageBox: {} },
  '@element-plus/icons-vue': {},
  '@/api/examGrading': { hasExamGradingPermission: () => false },
  '../../api': { adminApi: {
    examDetail: async () => ({ data: detail }),
    examRecordList: async () => ({ data: { list: [record], total: 1 } }),
    examRecordDetail: async () => ({ data: { record, answers, schema: { questions }, results: [] } }),
  } },
}
const compiled = ts.transpileModule(script, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
const module = { exports: {} }
new Function('require', 'module', 'exports', compiled)(name => {
  assert.ok(name in dependencies, `unexpected dependency: ${name}`)
  return dependencies[name]
}, module, module.exports)
const originalDocument = globalThis.document
try {
  // Only label-to-plain-text conversion needs the DOM; no network or real user data.
  globalThis.document = { createElement: () => ({ innerHTML: '', get textContent() { return this.innerHTML.replace(/<[^>]*>/g, '') } }) }
  const state = vue.proxyRefs(module.exports.default.setup({}, { expose() {} }))
  await state.load()
  assert.equal(state.questions.length, questions.length, '列表必须读取详情接口顶层 schema，不能读取不存在的 survey.schema')
  const expected = '考生姓名：测试考生；手机：13800000000；邮箱：exam@example.com；学号：S001；工号：E001；班级：一班'
  assert.equal(state.formatRecordUserInfo(state.list[0]), expected, '未登录答卷应显示填写的用户信息题答案')
  await state.viewDetail(state.list[0])
  assert.equal(state.formatRecordUserInfo(state.detailDialog.record), expected, '详情与列表的用户信息应一致')
  assert.equal(state.formatRecordUserInfo({ userId: '42', answers }), `用户ID：42；${expected}`)
  assert.equal(state.formatRecordUserInfo({ userId: '0', answers: { 'q-name': '旧答卷' } }), '考生姓名：旧答卷')
  for (const raw of ['', '{invalid', '[]', 'null', JSON.stringify({ 'q-name': ' ', 'q-phone': null, 'q-answer': '无身份信息' })]) {
    assert.equal(state.formatRecordUserInfo({ userId: '', answers: raw }), '匿名')
  }
  assert.equal(state.formatRecordUserInfo({ userId: '42', answers: '' }), '用户ID：42')
  detail = { survey: { id: 5, title: '无用户信息题' }, schema: '{"questions":[]}' }
  await state.load()
  assert.equal(state.formatRecordUserInfo(state.list[0]), '匿名', '没有配置用户信息题时不得猜测普通答案中的姓名')
} finally { globalThis.document = originalDocument }
console.log('exam record user info runtime passed: actual detail DTO, list/detail, six identity types, account fallback, legacy and empty answers')
