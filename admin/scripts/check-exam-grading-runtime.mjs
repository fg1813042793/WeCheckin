import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import * as vue from 'vue'
import { parse, compileScript, compileTemplate, compileStyle } from '@vue/compiler-sfc'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const read = path => readFileSync(new URL(`../src/${path}`, import.meta.url), 'utf8')
const plain = value => JSON.parse(JSON.stringify(value))
const model = executeTypeScriptModule(read('views/exam/components/examGradingModel.ts'), 'examGradingModel.ts')
const questionSource = read('views/exam/components/ExamGradingQuestion.vue')
assert.match(
  questionSource,
  /\.grading-question__answers\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\);/s,
  '考生答案与参考答案必须上下单列显示',
)
assert.match(
  questionSource,
  /<h4 class="grading-question__reference-title">参考答案<\/h4>/,
  '参考答案标题必须提供独立样式类',
)
assert.match(
  questionSource,
  /\.grading-question__answers\s+\.grading-question__reference-title\s*\{[^}]*color:\s*var\(--admin-ui-color-success\);/s,
  '参考答案标题必须使用成功绿色',
)
let permissions = new Set(['admin:menu:exam:record-grade', 'admin:api:exam-record:grade'])
const requests = []
const api = executeTypeScriptModule(read('api/examGrading.ts'), 'examGrading.ts', {
  '@/utils/permission': { hasPerm: permission => permissions.has(permission) },
  '@/utils/request': { default: {
    get: (...args) => requests.push(['GET', ...args]), post: (...args) => requests.push(['POST', ...args]),
  }, __esModule: true },
})
api.examGradingApi.get(5, 8)
api.examGradingApi.submit(5, 8, { expectedVersion: 'v1', scores: [{ questionId: 2, score: 0 }], comment: '评语' })
assert.equal(requests[0][1], '/api/v2/admin/exams/5/records/8/grading')
assert.equal(requests[1][1], requests[0][1])
assert.equal(requests[1][3].headers['Content-Type'], 'application/json')
assert.deepEqual(JSON.parse(requests[1][3].transformRequest[0](requests[1][2])), requests[1][2])
for (const missing of [...permissions]) {
  permissions.delete(missing)
  assert.equal(api.hasExamGradingPermission(), false)
  assert.throws(() => api.examGradingApi.get(5, 8), /权限/)
  assert.throws(() => api.examGradingApi.submit(5, 8, {}), /权限/)
  permissions.add(missing)
}
assert.throws(() => api.examGradingApi.get(NaN, 8), /参数/)
assert.equal(requests.length, 2, '缺任一权限时不得发出请求')

const descriptors = new Map()
const templates = new Map()
for (const path of ['views/exam/ExamRecordGrading.vue', 'views/exam/ExamResponses.vue', 'views/exam/components/ExamGradingQuestion.vue']) {
  const source = read(path)
  const { descriptor, errors } = parse(source, { filename: path })
  assert.deepEqual(errors, [])
  const script = compileScript(descriptor, { id: path })
  const template = compileTemplate({ source: descriptor.template.content, filename: path, id: path, compilerOptions: { bindingMetadata: script.bindings } })
  assert.deepEqual(template.errors, [], `${path} template compilation`)
  for (const style of descriptor.styles) assert.deepEqual(compileStyle({ source: style.content, filename: path, id: path, scoped: style.scoped }).errors, [])
  assert.ok(!source.includes('v-html'), '答卷与评语只允许文本渲染')
  if (path.endsWith('ExamRecordGrading.vue')) assert.ok(source.split('\n').length <= 400)
  descriptors.set(path, script.content)
  templates.set(path, template.code)
}

const runtimeErrors = []
const node = (tag = '') => ({ tag, text: '', props: {}, children: [], parent: null })
const renderer = vue.createRenderer({
  createElement: tag => node(tag),
  createText: text => ({ ...node('#text'), text }), createComment: () => node('#comment'),
  setText(value, text) { value.text = text }, setElementText(value, text) { value.text = text },
  patchProp(value, key, _previous, next) { value.props[key] = next },
  parentNode: value => value.parent, nextSibling: () => null,
  insert(child, parent) { child.parent = parent; parent.children.push(child) },
  remove(child) { if (child.parent) child.parent.children = child.parent.children.filter(item => item !== child) },
})
const settle = async () => { for (let i = 0; i < 8; i++) { await vue.nextTick(); await Promise.resolve() } }
function mount(path, modules, props = {}, renderTemplate = false) {
  const component = executeTypeScriptModule(descriptors.get(path), path, { vue, ...modules }).default
  let state
  const setup = component.setup
  component.setup = (props, ctx) => { const result = setup(props, ctx); state = vue.proxyRefs(result); return result }
  component.render = renderTemplate ? executeTypeScriptModule(templates.get(path), `${path}.template`, { vue }).render : () => null
  const active = vue.ref(true)
  const other = vue.defineComponent({ render: () => null })
  const app = renderer.createApp({ render: () => vue.h(vue.KeepAlive, null, { default: () => vue.h(active.value ? component : other, props) }) })
  for (const name of ['el-tag', 'el-form-item', 'el-input-number']) {
    app.component(name, { setup: (_, { attrs, slots }) => () => vue.h('div', attrs, slots.default?.()) })
  }
  app.config.errorHandler = error => runtimeErrors.push(error)
  app.config.warnHandler = message => runtimeErrors.push(new Error(message))
  const root = node()
  app.mount(root)
  return { get state() { return state }, app, active, root }
}

const descendants = root => [root, ...root.children.flatMap(descendants)]
const renderedText = root => descendants(root).map(item => item.text).join(' ')
// 用内存渲染器检查实际卡片输出；ElImage 仅替换网络加载，保留预览参数和反馈插槽。
const questionModules = {
  './examGradingModel': model,
  'element-plus': { ElImage: {
    props: ['src', 'alt', 'previewSrcList', 'initialIndex', 'previewTeleported'],
    setup: (props, { slots }) => () => vue.h('image-preview', { ...props }, [slots.placeholder?.(), slots.error?.()]),
  } },
}
const questionProps = {
  question: { questionId: 2, title: '完整题目', answer: '作答', referenceAnswer: '参考', needManual: true, fullScore: 10, gotScore: 0 },
  index: 0, editable: false, pending: true, busy: false, score: undefined, error: '',
}
const legacyQuestion = mount('views/exam/components/ExamGradingQuestion.vue', questionModules, questionProps, true)
assert.match(renderedText(legacyQuestion.root), /完整题目/)
assert.match(renderedText(legacyQuestion.root), /作答/)
assert.equal(descendants(legacyQuestion.root).filter(item => item.tag === 'image-preview').length, 0)
legacyQuestion.app.unmount()
const unscoredQuestion = mount('views/exam/components/ExamGradingQuestion.vue', questionModules, {
  ...questionProps, editable: true, pending: false,
  question: { ...questionProps.question, title: '不计分项目经历', answer: '完整项目描述', unscored: true, needManual: false, fullScore: 0 },
}, true)
assert.match(renderedText(unscoredQuestion.root), /不计分项目经历/)
assert.match(renderedText(unscoredQuestion.root), /完整项目描述/)
assert.match(renderedText(unscoredQuestion.root), /不计分/)
assert.doesNotMatch(renderedText(unscoredQuestion.root), /得分：|参考答案|必填，0 分有效/)
unscoredQuestion.app.unmount()
const illustrated = mount('views/exam/components/ExamGradingQuestion.vue', questionModules, {
  ...questionProps,
  question: { ...questionProps.question,
    materials: [
      { label: '矩阵行说明', content: '第一行\n第二行' },
      { label: '填空项', content: '<img src=x onerror=alert(1)>' },
      { label: '选项', content: '甲\n乙' },
      { label: '描述', content: '' },
    ],
    images: [
      { url: '/uploads/question.png', alt: '题目配图' },
      { url: 'javascript:alert(1)', alt: '不安全图片' },
      { url: '', alt: '缺失图片' },
      { url: 'https://example.test/option.png', alt: '选项配图' },
    ],
  },
}, true)
const illustrationText = renderedText(illustrated.root)
for (const text of ['矩阵行说明', '第一行\n第二行', '填空项', '<img src=x onerror=alert(1)>', '选项', '描述', '（无内容）', '图片加载中', '图片加载失败', '图片地址为空或不受支持']) {
  assert.ok(illustrationText.includes(text), `卡片应显示 ${text}`)
}
const previewImages = descendants(illustrated.root).filter(item => item.tag === 'image-preview')
assert.equal(previewImages.length, 2, '空地址和非法地址不得传给图片组件')
assert.equal(descendants(illustrated.root).filter(item => item.tag === 'img').length, 0, '材料内容只渲染为文本')
assert.deepEqual(plain(previewImages[0].props.previewSrcList), ['/uploads/question.png', 'https://example.test/option.png'])
assert.equal(previewImages[1].props.initialIndex, 1, '过滤非法地址后预览仍定位到当前图片')
assert.equal(previewImages[1].props.alt, '选项配图')
illustrated.app.unmount()
const emptyImages = mount('views/exam/components/ExamGradingQuestion.vue', questionModules, {
  ...questionProps, question: { ...questionProps.question, images: [] },
}, true)
assert.match(renderedText(emptyImages.root), /暂无题目配图/)
emptyImages.app.unmount()

const initial = {
  record: { id: 8, examId: 5, status: 1, score: 6, totalScore: 20, gradingComment: '' },
  examTitle: '测试考试', version: 'v1', canGrade: true, passScore: 12, notice: '历史答卷提示',
  questions: [
    { questionId: 1, title: '客观题', fullScore: 6, gotScore: 6, needManual: false },
    { questionId: 2, title: '人工题 A', fullScore: 10, gotScore: 0, needManual: true },
    { questionId: 3, title: '人工题 B', fullScore: 4, gotScore: 0, needManual: true },
  ],
}
let allowed = true, reads = 0, confirms = 0, failure = null, cancel = false, releaseConfirm
const submissions = [], navigation = []
let server = plain(initial)
const gradingModules = {
  'vue-router': { useRoute: () => ({ params: { recordId: '8' }, query: { examId: '5', returnTo: '/exam/responses?examId=5&title=测试' } }), useRouter: () => ({ push: value => navigation.push(value) }) },
  'element-plus': { ElMessage: { success() {} }, ElMessageBox: { confirm: () => {
    confirms++
    if (cancel) return Promise.reject('cancel')
    return releaseConfirm === null ? new Promise(resolve => { releaseConfirm = resolve }) : Promise.resolve()
  } } },
  '@/components/admin-ui': {}, './components/ExamGradingQuestion.vue': {},
  './components/examGradingModel': model,
  '@/api/examGrading': { hasExamGradingPermission: () => allowed, examGradingApi: {
    get: async () => { reads++; if (failure?.load) throw failure.load; return { data: plain(server) } },
    submit: async (_exam, _record, payload) => {
      submissions.push(plain(payload))
      if (failure?.submit) throw failure.submit
      return { data: { ...plain(server), canGrade: false, record: { ...server.record, status: 2, score: 9, gradedBy: 99, gradedAt: 123, gradingComment: payload.comment } } }
    },
  } },
}
const page = mount('views/exam/ExamRecordGrading.vue', gradingModules)
await settle()
const state = page.state
assert.equal(reads, 1)
assert.deepEqual(plain(state.draft), {})
assert.equal(state.requiresReview, false, '首次加载不要求额外核对')
await state.submit()
assert.equal(confirms, 0)
assert.match(state.submitError, /全部/)
state.updateScore(2, 0); state.updateScore(3, 3); state.comment = '保留此评语'
assert.deepEqual(plain(state.preview), { score: 9, remaining: 0 })
cancel = true; await state.submit(); cancel = false
assert.equal(submissions.length, 0, '取消确认不提交')
failure = { submit: { response: { data: { msg: '网络失败，请重试' } } } }
await state.submit()
assert.match(state.submitError, /网络失败/)
assert.deepEqual(plain(state.draft), { 2: 0, 3: 3 })
assert.equal(state.comment, '保留此评语')
assert.equal(state.submitting, false)
page.active.value = false; await settle(); page.active.value = true; await settle()
assert.deepEqual(plain(state.draft), { 2: 0, 3: 3 }, '切换页签不得覆盖失败后的草稿')
failure = null; releaseConfirm = null
const pending = state.submit(); await settle(); await state.submit()
assert.equal(submissions.length, 1, '等待确认时防止重复提交')
state.updateScore(2, 7)
assert.equal(state.draft[2], 0, '提交期间锁定评分')
releaseConfirm(); await pending
assert.deepEqual(submissions[1], { expectedVersion: 'v1', scores: [{ questionId: 2, score: 0 }, { questionId: 3, score: 3 }], comment: '保留此评语' })
assert.equal(state.context.record.status, 2)
assert.equal(state.editable, false)
await state.submit(); assert.equal(submissions.length, 2)
state.goBack(); assert.equal(navigation[0], '/exam/responses?examId=5&title=测试')
page.app.unmount()

for (const patch of [{ record: { ...initial.record, status: 0 } }, { record: { ...initial.record, status: 2 } }, { canGrade: false }, { questions: [] }]) {
  server = { ...plain(initial), ...patch }
  const readonly = mount('views/exam/ExamRecordGrading.vue', gradingModules)
  await settle(); await readonly.state.submit()
  assert.equal(submissions.length, 2)
  assert.ok(readonly.state.readOnlyNotice)
  readonly.app.unmount()
}
server = plain(initial); allowed = false
const beforeReads = reads
const denied = mount('views/exam/ExamRecordGrading.vue', gradingModules)
await settle(); await denied.state.submit()
assert.equal(reads, beforeReads)
assert.equal(denied.state.permitted, false)
denied.app.unmount(); allowed = true
failure = { load: { msg: '没有此考试的数据权限' } }
const failed = mount('views/exam/ExamRecordGrading.vue', gradingModules)
await settle(); assert.match(failed.state.loadError, /数据权限/)
failure = null; await failed.state.load()
assert.equal(failed.state.editable, true)
failed.app.unmount()

// 版本冲突恢复必须运行真实页面逻辑，确保请求、草稿和提交门禁一起更新。
server = plain(initial)
server.questions.push({ questionId: 4, title: '将删除的题目', fullScore: 5, gotScore: 0, needManual: true })
const recovery = mount('views/exam/ExamRecordGrading.vue', gradingModules)
await settle()
const recoveryState = recovery.state
recoveryState.updateScore(2, 7); recoveryState.updateScore(3, 4); recoveryState.updateScore(4, 2)
recoveryState.comment = '重载后仍保留\n人工评语'
failure = { submit: { response: { data: { msg: '版本冲突，请重新加载答卷' } } } }
await recoveryState.submit()
assert.match(recoveryState.submitError, /版本冲突/)
const afterConflict = submissions.length
failure = { load: new Error('重新加载网络失败') }
await recoveryState.load()
assert.match(recoveryState.loadError, /网络失败/)
assert.equal(recoveryState.context.version, 'v1')
assert.deepEqual(plain(recoveryState.draft), { 2: 7, 3: 4, 4: 2 }, '加载失败保留所有旧评分')
assert.equal(recoveryState.comment, '重载后仍保留\n人工评语')
server.version = 'v2'
server.questions = [
  server.questions[0],
  { ...server.questions[1], title: '最新题干', answer: '最新作答', fullScore: 8 },
  { ...server.questions[2], fullScore: 3 },
  { questionId: 5, title: '新增人工题', fullScore: 4, gotScore: 0, needManual: true },
]
failure = null
await recoveryState.load()
assert.equal(recoveryState.loadError, '')
assert.equal(recoveryState.context.questions[1].answer, '最新作答')
assert.deepEqual(plain(recoveryState.draft), { 2: 7 }, '保留合法分数，丢弃消失题和超新满分的分数，不预填新增题')
assert.equal(recoveryState.comment, '重载后仍保留\n人工评语')
assert.equal(recoveryState.requiresReview, true)
assert.equal(recoveryState.reviewConfirmed, false)
recoveryState.updateScore(3, 0); recoveryState.updateScore(5, 2)
const beforeReviewConfirms = confirms
await recoveryState.submit()
assert.equal(submissions.length, afterConflict, '未明确核对时不能重试旧评分')
assert.equal(confirms, beforeReviewConfirms, '未勾选不能绕过核对门禁进入提交确认')
assert.match(recoveryState.submitError, /核对/)
await recoveryState.load()
assert.equal(recoveryState.requiresReview, true, '重复加载相同版本不得绕过待核对状态')
assert.equal(recoveryState.reviewConfirmed, false)
assert.deepEqual(plain(recoveryState.draft), { 2: 7, 3: 0, 5: 2 })
recoveryState.reviewConfirmed = true
failure = { submit: new Error('普通网络失败') }
await recoveryState.submit()
assert.equal(submissions.at(-1).expectedVersion, 'v2', '提交必须使用重新加载后的版本')
assert.equal(recoveryState.comment, '重载后仍保留\n人工评语')
assert.equal(recoveryState.reviewConfirmed, true, '网络失败无需重新确认已核对的版本')
failure = null
await recoveryState.submit()
assert.equal(submissions.length, afterConflict + 2, '普通网络失败可直接重试')
assert.deepEqual(submissions.at(-1), {
  expectedVersion: 'v2', scores: [{ questionId: 2, score: 7 }, { questionId: 3, score: 0 }, { questionId: 5, score: 2 }],
  comment: '重载后仍保留\n人工评语',
})
assert.equal(recoveryState.editable, false)
recovery.app.unmount()

server = plain(initial)
const changed = mount('views/exam/ExamRecordGrading.vue', gradingModules)
await settle()
changed.state.updateScore(2, 0); changed.state.updateScore(3, 3); changed.state.comment = ''
server.record.gradingComment = '服务端旧评语'
await changed.state.load()
assert.deepEqual(plain(changed.state.draft), { 2: 0, 3: 3 })
assert.equal(changed.state.comment, '', '用户清空的评语也必须保留')
assert.equal(changed.state.requiresReview, false, '同版本普通重载不需要核对')
server.version = 'v2'; server.questions[2].needManual = false
await changed.state.load()
assert.deepEqual(plain(changed.state.draft), { 2: 0 }, '不再人工评分的题目应丢弃旧评分')
changed.state.reviewConfirmed = true
server.version = 'v3'
await changed.state.load()
assert.equal(changed.state.reviewConfirmed, false, '再次变更版本必须重新勾选')
server.record = { ...server.record, status: 2, gradingComment: '他人已批改', score: 9 }
server.canGrade = false; server.questions[1].gotScore = 5
await changed.state.load()
assert.equal(changed.state.editable, false)
assert.deepEqual(plain(changed.state.draft), {}, '已批改时丢弃本地评分草稿')
assert.equal(changed.state.comment, '他人已批改')
assert.equal(changed.state.context.questions[1].gotScore, 5, '只读展示服务端逐题得分')
assert.equal(changed.state.requiresReview, false)
const beforeReadonlySubmit = submissions.length
await changed.state.submit()
assert.equal(submissions.length, beforeReadonlySubmit)
changed.app.unmount()

const listReads = []
const list = mount('views/exam/ExamResponses.vue', {
  'vue-router': { useRoute: () => ({ query: { examId: '5', title: '考试' }, fullPath: '/exam/responses?examId=5&title=考试' }), useRouter: () => ({ push: value => navigation.push(value) }) },
  'element-plus': { ElMessage: { error() {}, success() {} }, ElMessageBox: {} }, '@element-plus/icons-vue': {},
  '@/api/examGrading': { hasExamGradingPermission: () => allowed },
  '../../api': { adminApi: {
    examRecordList: async query => { listReads.push(plain(query)); return { data: { list: [{ id: 8, status: listReads.length === 1 ? 1 : 2 }], total: 61 } } },
    examDetail: async () => ({ data: { schema: '{"questions":[]}' } }),
  } },
})
await settle()
assert.equal(listReads.length, 1)
list.state.page = 3; list.state.pageSize = 20; list.state.keyword = 'user'
list.active.value = false; await settle(); list.active.value = true; await settle()
assert.equal(listReads.length, 2)
assert.deepEqual(listReads[1], { examId: 5, page: 3, pageSize: 20, keyword: 'user' })
assert.equal(list.state.list[0].status, 2)
assert.equal(list.state.gradeState({ grade: { gotScore: 3, fullScore: 10, correct: false } }).label, '部分得分')
Object.assign(list.state.detailDialog, {
  record: { status: 2 }, schema: null, answers: { 2: '历史答案' }, scoring: { 2: false },
  results: [{ questionId: 2, gotScore: 3, fullScore: 10, correct: false }],
})
assert.equal(list.state.gradeState(list.state.answerRows[0]).label, '部分得分', '无 schema 的历史答卷也应使用逐题分数')
assert.equal(list.state.canOpenGrading({ id: 8, status: 0 }), false)
assert.equal(list.state.canOpenGrading({ id: 8, status: 2 }), true)
allowed = false; assert.equal(list.state.canOpenGrading({ id: 8, status: 1 }), false)
list.app.unmount()
assert.deepEqual(runtimeErrors, [])
console.log('exam grading SFC/API/runtime passed: materials/images, permissions, blank/zero scores, confirmation, reload recovery, version review, retry, readonly results, cached list refresh')
