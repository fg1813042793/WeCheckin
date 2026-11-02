import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import * as vue from 'vue'
import { compileScript, compileTemplate, parse } from '@vue/compiler-sfc'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../src/views/user-feedback')
const read = path => readFileSync(resolve(root, path), 'utf8')
const plain = value => JSON.parse(JSON.stringify(value))
const deferred = () => {
  let resolve, reject
  const promise = new Promise((yes, no) => { resolve = yes; reject = no })
  return { promise, resolve, reject }
}
const settle = async () => { for (let i = 0; i < 5; i++) await vue.nextTick() }
const empty = { render: () => null }
const passthrough = { setup: (_, { attrs, slots }) => () => vue.h('div', attrs, [slots.default?.(), slots.footer?.()]) }
const node = (tag = '') => ({ tag, text: '', props: {}, children: [], parent: null })
const renderer = vue.createRenderer({
  createElement: tag => node(tag), createText: text => ({ ...node('#text'), text }), createComment: () => node('#comment'),
  setText: (node, text) => { node.text = text },
  setElementText: (node, text) => { node.text = text; node.children = [] },
  patchProp: (node, key, _old, next) => { node.props[key] = next },
  parentNode: node => node.parent,
  nextSibling: node => node.parent?.children[node.parent.children.indexOf(node) + 1] || null,
  insert(child, parent, anchor = null) {
    if (child.parent) child.parent.children = child.parent.children.filter(item => item !== child)
    child.parent = parent
    const index = anchor ? parent.children.indexOf(anchor) : -1
    if (index < 0) parent.children.push(child)
    else parent.children.splice(index, 0, child)
  },
  remove(child) { if (child.parent) child.parent.children = child.parent.children.filter(item => item !== child) },
})
const descendants = node => [node, ...node.children.flatMap(descendants)]
const renderedText = root => descendants(root).map(node => node.text).join(' ')
const button = (root, label) => descendants(root).find(node => node.tag === 'button' && renderedText(node).trim() === label)

function harness() {
  const allowed = vue.ref(true)
  const messages = [], errors = [], runtimeErrors = [], requests = []
  const api = {
    userFeedbackOverview: async () => ({ data: {} }),
    userFeedbackList: async () => ({ data: { list: [], total: 0, page: 1, pageSize: 20 } }),
    userFeedbackDetail: id => { const request = deferred(); requests.push({ id, ...request }); return request.promise },
  }
  const status = executeTypeScriptModule(read('userFeedbackStatus.ts'), 'userFeedbackStatus.ts', {
    '@/utils/permission': { hasPerm: () => allowed.value },
  })
  const modules = {
    vue,
    '@/api': { adminApi: api },
    '@/utils/request': { showRequestError: error => errors.push(error) },
    'element-plus': { ElMessage: { warning: text => messages.push(text), success: text => messages.push(text) } },
    '@/components/admin-ui': { AdminDialog: passthrough, AdminDrawer: passthrough, AdminPageShell: passthrough, AdminPageHeader: passthrough, AdminSearchBar: passthrough, AdminTablePanel: passthrough },
    '@element-plus/icons-vue': { Picture: empty, Refresh: empty, View: empty },
    './userFeedbackStatus': status,
    '../userFeedbackStatus': status,
    './components/UserFeedbackOverview.vue': { default: empty, __esModule: true },
    './components/UserFeedbackDetailDrawer.vue': { default: empty, __esModule: true },
    './components/UserFeedbackStatusDialog.vue': { default: empty, __esModule: true },
    './components/UserFeedbackNotifyDialog.vue': { default: empty, __esModule: true },
    './UserFeedbackTimeline.vue': { default: empty, __esModule: true },
    'test:document': { document: { addEventListener() {}, removeEventListener() {} } },
  }
  if (existsSync(resolve(root, 'composables/useUserFeedbackActions.ts'))) {
    modules['./composables/useUserFeedbackActions'] = executeTypeScriptModule(read('composables/useUserFeedbackActions.ts'), 'useUserFeedbackActions.ts', modules)
  }
  function mount(path, props = {}, render = false) {
    const filename = resolve(root, path)
    const { descriptor, errors } = parse(read(path), { filename })
    assert.deepEqual(errors, [])
    const compiled = compileScript(descriptor, { id: path })
    const component = executeTypeScriptModule(`import { document } from 'test:document';\n${compiled.content}`, `${path}.ts`, modules).default
    let state, exposed
    const setup = component.setup
    component.setup = (props, context) => {
      const result = setup(props, { ...context, expose(value) { exposed = value; context.expose(value) } })
      state = vue.proxyRefs(result)
      return result
    }
    const template = compileTemplate({ source: descriptor.template.content, filename, id: path, compilerOptions: { bindingMetadata: compiled.bindings } })
    assert.deepEqual(template.errors, [])
    component.render = render ? executeTypeScriptModule(template.code, `${path}.render.ts`, { vue }).render : () => null
    const app = renderer.createApp(component, props)
    for (const tag of ['el-alert', 'el-form', 'el-form-item', 'el-input', 'el-select', 'el-option', 'el-switch', 'el-radio-group', 'el-radio', 'el-tag', 'el-descriptions', 'el-descriptions-item', 'el-timeline', 'el-timeline-item', 'el-empty', 'el-image']) app.component(tag, passthrough)
    app.component('el-button', { setup: (_, { attrs, slots }) => () => vue.h('button', attrs, slots.default?.()) })
    app.directive('loading', {})
    app.config.errorHandler = error => runtimeErrors.push(error)
    app.config.warnHandler = message => runtimeErrors.push(new Error(message))
    const container = node()
    app.mount(container)
    assert.deepEqual(runtimeErrors, [])
    return { state, exposed, app, root: container }
  }
  return { allowed, messages, errors, runtimeErrors, requests, api, mount }
}

const feedback = (id = '1', version = 1, status = 'pending') => ({
  id, feedbackNo: `FB-${id}`, submitterId: 9, submitterName: '测试提交人', summary: '反馈', imageCount: 0,
  handlerId: 7, handlerName: '处理人', status, version, messages: [], allowsSupplement: true,
  createdAt: 1, updatedAt: 1, lastActivityAt: 1,
})

async function checkPageActions() {
  const h = harness()
  const page = h.mount('index.vue')
  const refreshes = []
  h.api.userFeedbackOverview = async () => { refreshes.push('overview'); return { data: {} } }
  h.api.userFeedbackList = async () => { refreshes.push('list'); return { data: { list: [], total: 0, page: 1, pageSize: 20 } } }
  page.state.detailDrawerRef = { refresh: async () => { refreshes.push('detail') } }
  await page.state.handleStatusSuccess(feedback('1', 2))
  assert.deepEqual(refreshes, ['overview', 'list', 'detail'], '操作成功应刷新统计、列表和已打开详情')
  const opened = []
  const statusDialog = { visible: false, open: detail => opened.push(['status', plain(detail)]) }
  const notifyDialog = { visible: false, open: detail => opened.push(['notify', plain(detail)]) }
  page.state.statusDialogRef = statusDialog
  page.state.notifyDialogRef = notifyDialog
  assert.equal(typeof page.state.openNotifyDialog, 'function', '列表和抽屉必须有独立通知入口')

  h.allowed.value = false
  await page.state.openStatusDialog(feedback())
  await page.state.openNotifyDialog(feedback())
  assert.equal(h.requests.length, 0, '无权限时直接调用 handler 也不得读取操作详情')
  h.allowed.value = true
  const first = page.state.openStatusDialog(feedback('1'))
  assert.equal(h.requests.length, 1)
  assert.equal(opened.length, 0, '读取最新详情前不得打开操作弹窗')
  assert.equal(page.state.actionLoading.id, '1')
  await page.state.openStatusDialog(feedback('1'))
  assert.equal(h.requests.length, 1, '同一操作加载中不得重复获取详情')
  const second = page.state.openNotifyDialog(feedback('2'))
  h.requests[0].resolve({ data: feedback('1', 5) })
  await first
  assert.equal(opened.length, 0, '先前行的迟到响应不得打开旧弹窗')
  assert.equal(page.state.actionLoading.id, '2', '旧 finally 不得清除新行的 loading')
  h.requests[1].resolve({ data: feedback('2', 8, 'closed') })
  await second
  assert.equal(opened[0][0], 'notify')
  assert.equal(opened[0][1].version, 8)
  assert.equal(opened[0][1].status, 'closed', '所有状态均可独立通知')
  assert.equal(page.state.actionLoading, null)

  const failed = page.state.openStatusDialog(feedback('3'))
  h.requests.at(-1).reject(new Error('读取失败'))
  await failed
  assert.equal(opened.length, 1, '读取失败不得复用旧 detail')
  assert.equal(page.state.actionLoading, null)
  assert.match(page.state.actionError, /加载失败/)
  const retry = page.state.openStatusDialog(feedback('3'))
  h.requests.at(-1).resolve({ data: feedback('3', 12, 'processing') })
  await retry
  assert.equal(opened.at(-1)[1].version, 12)
  assert.equal(page.state.actionError, '')

  const olderFailure = page.state.openStatusDialog(feedback('stale'))
  const staleRequest = h.requests.at(-1)
  const newer = page.state.openNotifyDialog(feedback('new'))
  const errorCount = h.errors.length
  staleRequest.reject(new Error('旧行加载失败'))
  await olderFailure
  assert.equal(h.errors.length, errorCount)
  assert.equal(page.state.actionError, '')
  assert.equal(page.state.actionLoading.id, 'new')
  page.state.cancelPendingAction()
  h.requests.at(-1).resolve({ data: feedback('new', 3) })
  await newer
  assert.equal(opened.length, 2)

  const revoked = page.state.openNotifyDialog(feedback('4'))
  h.allowed.value = false
  h.requests.at(-1).resolve({ data: feedback('4', 10) })
  await revoked
  assert.equal(opened.length, 2, '详情返回前权限撤销，不得打开弹窗')
  h.allowed.value = true
  notifyDialog.visible = true
  const count = h.requests.length
  await page.state.openStatusDialog(feedback('5'))
  assert.equal(h.requests.length, count, '已有草稿弹窗时不得覆盖或叠加操作')
  notifyDialog.visible = false

  const cancelled = page.state.openNotifyDialog(feedback('6'))
  page.state.openDetail(feedback('7'))
  h.requests.at(-1).resolve({ data: feedback('6', 10) })
  await cancelled
  assert.equal(opened.length, 2, '改为查看详情后应取消待打开的操作')

  const pending = page.state.openNotifyDialog(feedback('8'))
  const errorsBefore = h.errors.length
  page.app.unmount()
  h.requests.at(-1).reject(new Error('过期请求失败'))
  await pending
  assert.equal(h.errors.length, errorsBefore, '卸载后旧请求不得提示错误')
  assert.equal(opened.length, 2)
  assert.deepEqual(h.runtimeErrors, [])
}

async function checkNotifyDialog() {
  const h = harness(), calls = [], successes = []
  h.api.userFeedbackNotify = (id, body) => { const request = deferred(); calls.push({ id, body: plain(body), ...request }); return request.promise }
  const dialog = h.mount('components/UserFeedbackNotifyDialog.vue', { onSuccess: detail => successes.push(plain(detail)) }, true)
  const s = dialog.state
  h.allowed.value = false
  dialog.exposed.open(feedback())
  assert.equal(s.visible, false)
  h.allowed.value = true
  dialog.exposed.open(feedback('9', 4, 'closed'))
  await settle()
  for (const label of ['FB-9', '测试提交人', '已关闭']) assert.ok(renderedText(dialog.root).includes(label), `通知确认应显示 ${label}`)
  for (const label of ['站内通知', '钉钉通知']) assert.ok(renderedText(dialog.root).includes(label), `通知渠道应可选择 ${label}`)
  const radioOptions = descendants(dialog.root).filter(node => ['internal', 'dingtalk'].includes(node.props.value))
  assert.deepEqual(radioOptions.map(node => node.props.value), ['internal', 'dingtalk'], '应绑定两个真实单选值')
  const channelControl = descendants(dialog.root).find(node => node.props['aria-label'] === '通知渠道')
  channelControl.props['onUpdate:modelValue']('dingtalk')
  assert.equal(s.channel, 'dingtalk', '单选交互应更新提交渠道')
  channelControl.props['onUpdate:modelValue']('internal')
  assert.equal(s.channel, 'internal', '通知渠道默认站内通知')
  assert.ok(button(dialog.root, '确认发送'), '通知必须由确认按钮发送')
  s.content = ' \n '
  await s.submit()
  assert.equal(calls.length, 0)
  assert.match(s.contentError, /填写/)
  s.content = '文'.repeat(5001)
  await s.submit()
  assert.equal(calls.length, 0)
  assert.match(s.contentError, /5000/)
  const contentControl = descendants(dialog.root).find(node => node.props['aria-label'] === '通知正文')
  contentControl.props['onUpdate:modelValue']('  已收到您的补充，请查看  \n')
  assert.equal(s.content, '  已收到您的补充，请查看  \n', '正文输入应绑定可编辑草稿')
  s.channel = 'unsupported'
  await s.submit()
  assert.equal(calls.length, 0, '非法渠道不应发请求')
  s.channel = 'internal'
  const first = s.submit()
  await settle()
  assert.equal(s.submitting, true)
  assert.equal(button(dialog.root, '取消').props.disabled, true)
  assert.equal(button(dialog.root, '确认发送').props.loading, true)
  const dialogNode = descendants(dialog.root).find(node => node.props.title === '通知提交人')
  assert.equal(dialogNode.props['show-close'], false)
  assert.equal(dialogNode.props['close-on-press-escape'], false)
  assert.equal(dialogNode.props['close-on-click-modal'], false)
  assert.equal(calls.length, 1)
  assert.deepEqual(Object.keys(calls[0].body).sort(), ['channel', 'content', 'requestId', 'version'])
  assert.equal(calls[0].body.channel, 'internal')
  assert.equal(calls[0].body.content, '已收到您的补充，请查看')
  assert.equal(calls[0].body.version, 4)
  assert.equal(calls[0].id, '9')
  const requestId = calls[0].body.requestId
  assert.ok(requestId.length > 0 && requestId.length <= 64)
  await s.submit()
  s.handleVisibleChange(false)
  let closed = false
  s.beforeClose(() => { closed = true })
  dialog.exposed.open(feedback('10'))
  assert.equal(calls.length, 1, '提交 handler 必须阻止双击')
  assert.equal(closed, false)
  assert.equal(s.visible, true)
  assert.equal(s.detail.id, '9', '发送中不得切换接收人')
  calls[0].reject(new Error('网络失败'))
  await first
  assert.equal(s.visible, true)
  assert.equal(s.submitting, false)
  assert.equal(s.content, '  已收到您的补充，请查看  \n')
  assert.equal(s.requestId, requestId)
  h.allowed.value = false
  await s.submit()
  assert.equal(calls.length, 1, '发送前再次检查权限')
  h.allowed.value = true
  const retry = s.submit()
  assert.deepEqual(calls[1].body, calls[0].body, '重试必须发送同一 requestId 和 payload')
  const result = feedback('9', 5, 'closed')
  result.messages = [{ id: 'm1', messageType: 'notification', authorType: 'admin', authorId: 7, content: '已收到您的补充，请查看', attachments: [], createdAt: 2 }]
  calls[1].resolve({ data: result })
  await retry
  assert.equal(s.visible, false)
  assert.deepEqual(successes, [result])
  assert.ok(h.messages.includes('通知已提交发送'))
  assert.equal(result.handlerId, 7)
  await s.submit()
  assert.equal(calls.length, 2, '成功关闭后不能再次调用提交发送')

  dialog.exposed.open(feedback('9', 5))
  s.content = '字'.repeat(5000)
  const boundary = s.submit()
  assert.equal(calls.at(-1).body.content.length, 5000, '5000字正文应允许提交')
  calls.at(-1).reject(new Error('网络失败'))
  await boundary
  s.channel = 'dingtalk'
  s.content = '保留冲突草稿'
  const conflict = s.submit()
  const conflictKey = calls.at(-1).body.requestId
  assert.equal(calls.at(-1).body.channel, 'dingtalk')
  calls.at(-1).reject({ msg: '反馈已更新，请刷新后重试' })
  await conflict
  assert.equal(s.visible, true)
  assert.equal(s.content, '保留冲突草稿')
  assert.equal(s.channel, 'dingtalk', '失败保留所选渠道')
  assert.equal(s.requestId, conflictKey)
  assert.match(s.submitError, /刷新详情/)
  const conflictRetry = s.submit()
  assert.equal(calls.at(-1).body.requestId, conflictKey)
  calls.at(-1).reject({ msg: '反馈已更新，请刷新后重试' })
  await conflictRetry
  s.content = '不同通知正文'
  const edited = s.submit()
  assert.notEqual(calls.at(-1).body.requestId, conflictKey, '编辑为不同正文后不得重放旧通知')
  calls.at(-1).reject(new Error('网络失败'))
  await edited
  const editedKey = calls.at(-1).body.requestId
  s.channel = 'internal'
  const switched = s.submit()
  assert.notEqual(calls.at(-1).body.requestId, editedKey, '切换渠道应使用新的 requestId')
  assert.equal(calls.at(-1).body.channel, 'internal')
  calls.at(-1).reject(new Error('网络失败'))
  await switched
  s.channel = 'dingtalk'
  const unbound = s.submit()
  calls.at(-1).reject({ msg: '提交人未绑定钉钉' })
  await unbound
  assert.equal(s.content, '不同通知正文')
  assert.match(s.submitError, /未绑定钉钉/)
  s.beforeClose(() => { closed = true })
  assert.equal(closed, true, '失败后允许关闭')
  dialog.app.unmount()
  assert.deepEqual(h.runtimeErrors, [])
}

async function checkStatusDialog() {
  const h = harness(), calls = []
  h.api.userFeedbackUpdateStatus = (id, body) => { const request = deferred(); calls.push({ id, body: plain(body), ...request }); return request.promise }
  const dialog = h.mount('components/UserFeedbackStatusDialog.vue')
  h.allowed.value = false
  dialog.exposed.open(feedback())
  assert.equal(dialog.state.visible, false)
  h.allowed.value = true
  dialog.exposed.open(feedback('1', 8))
  assert.deepEqual(plain(dialog.state.statusOptions), ['processing', 'closed'])
  const saving = dialog.state.submit()
  await dialog.state.submit()
  assert.equal(calls.length, 1)
  assert.equal(calls[0].body.status, 'processing')
  assert.equal(calls[0].body.version, 8)
  dialog.state.handleVisibleChange(false)
  assert.equal(dialog.state.visible, true)
  calls[0].reject(new Error('失败'))
  await saving
  const retry = dialog.state.submit()
  assert.deepEqual(calls[1].body, calls[0].body)
  calls[1].resolve({ data: feedback('1', 9, 'processing') })
  await retry
  dialog.exposed.open(feedback('1', 9, 'processing'))
  assert.deepEqual(plain(dialog.state.statusOptions), ['resolved'])
  await dialog.state.submit()
  assert.equal(calls.length, 2, '完成处理仍必须填写处理说明')
  dialog.state.form.note = '已解决'
  h.allowed.value = false
  await dialog.state.submit()
  assert.equal(calls.length, 2)
  dialog.app.unmount()
  assert.deepEqual(h.runtimeErrors, [])
}

async function checkDrawerAndTimeline() {
  const h = harness(), statusEvents = [], notifyEvents = []
  const drawer = h.mount('components/UserFeedbackDetailDrawer.vue', {
    modelValue: true, feedbackId: '1', canHandle: true, formatTime: String,
    onRequestStatusUpdate: detail => statusEvents.push(detail), onRequestNotify: detail => notifyEvents.push(detail),
  }, true)
  drawer.state.detail = feedback()
  await settle()
  assert.ok(button(drawer.root, '处理'))
  assert.ok(button(drawer.root, '通知提交人'))
  button(drawer.root, '处理').props.onClick()
  button(drawer.root, '通知提交人').props.onClick()
  assert.equal(statusEvents.length, 1)
  assert.equal(notifyEvents.length, 1)
  h.allowed.value = false
  drawer.state.requestStatusUpdate()
  drawer.state.requestNotify()
  assert.equal(statusEvents.length, 1)
  assert.equal(notifyEvents.length, 1)
  h.allowed.value = true
  drawer.state.loading = true
  drawer.state.requestStatusUpdate()
  drawer.state.requestNotify()
  assert.equal(statusEvents.length, 1, '读取过程中不得使用旧 detail')
  assert.equal(notifyEvents.length, 1)
  drawer.app.unmount()
  const message = { id: '1', messageType: 'notification', authorType: 'admin', authorId: 1, content: '<b>通知正文</b>', attachments: [], createdAt: 1 }
  const timeline = h.mount('components/UserFeedbackTimeline.vue', { messages: [message], formatTime: String }, true)
  assert.match(renderedText(timeline.root), /通知提交人/)
  assert.match(renderedText(timeline.root), /管理员 #1/)
  assert.match(renderedText(timeline.root), /<b>通知正文<\/b>/)
  assert.ok(!descendants(timeline.root).some(node => node.tag === 'b'), '通知正文只能作为纯文本渲染')
  timeline.app.unmount()
  assert.deepEqual(h.runtimeErrors, [])
}

await checkPageActions()
await checkNotifyDialog()
await checkStatusDialog()
await checkDrawerAndTimeline()
console.log('Admin user feedback actions: real Vue setup, notifications, permissions, retries and races passed.')
