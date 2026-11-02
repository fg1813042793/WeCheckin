import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import * as vue from 'vue'
import { parse, compileScript } from '@vue/compiler-sfc'
import { renderToString } from '@vue/server-renderer'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = path => readFileSync(resolve(root, path), 'utf8')
const load = (path, modules = {}) => {
  assert.ok(existsSync(resolve(root, path)), `missing implementation: ${path}`)
  const source = (modules['node:url'] ? "const URL = require('node:url').URL;\n" : '') + read(path)
  return executeTypeScriptModule(source, path, { vue, ...modules })
}
const helpers = load('src/views/notification/templateHelpers.ts', { 'node:url': { URL } })
// Render the real selector without a browser; only Element Plus controls are stubbed.
async function renderSelector(props) {
  const fileName = 'src/views/notification/components/SendTemplateSelect.vue'
  const { descriptor } = parse(read(fileName), { filename: fileName })
  const compiled = compileScript(descriptor, { id: 'send-template-test', inlineTemplate: true })
  const component = executeTypeScriptModule(compiled.content, fileName + '.ts', {
    vue, '../templateHelpers': helpers, '@element-plus/icons-vue': { Refresh: () => vue.h('i') },
  }).default
  const buttons = []
  const app = vue.createSSRApp(component, props)
  for (const name of ['ElFormItem', 'ElSelect', 'ElOption', 'ElCollapse', 'ElCollapseItem']) {
    app.component(name, { setup: (_props, { slots }) => () => vue.h('div', slots.default?.()) })
  }
  app.component('ElButton', {
    inheritAttrs: false,
    setup: (_props, { attrs, slots }) => { buttons.push(attrs); return () => vue.h('button', attrs, slots.default?.()) },
  })
  return { html: await renderToString(app), buttons }
}
const { newTemplateDraft, templatePayload, validateTemplate, targetSummary, validateMessage, templateHeadColorPickerValue, templateHeadColorFromPicker } = helpers
assert.equal(templateHeadColorPickerValue('FF1677FF'), '#1677FF')
assert.equal(templateHeadColorPickerValue('#AABBCC'), '#AABBCC')
assert.equal(templateHeadColorFromPicker('#aabbcc'), 'FFAABBCC')
assert.equal(templateHeadColorFromPicker(null), 'FF1677FF')
const editorSource = read('src/views/notification/components/NotificationTemplateEditor.vue')
assert.match(editorSource, /<el-color-picker v-model="headColorPicker"/)
assert.doesNotMatch(editorSource, /<el-input v-model="form\.headColor"/)
assert.match(editorSource, /按钮颜色由钉钉客户端固定，不支持自定义/)
assert.match(read('src/views/notification/components/BusinessNotificationStyles.vue'), /按钮颜色由钉钉客户端固定，不支持自定义/)
const draft = () => ({ ...newTemplateDraft(), name: '日常提醒', title: '请填写', content: '今日事项' })
assert.deepEqual(Object.keys(validateTemplate(draft())), [])
for (const [field, value] of [['name', ' '], ['name', '字'.repeat(81)], ['title', ''], ['title', '字'.repeat(256)], ['content', '\n'], ['content', '字'.repeat(5001)], ['messageType', 'text'], ['headColor', '12345'], ['buttonTitle', '字'.repeat(21)]]) {
  assert.ok(validateTemplate({ ...draft(), [field]: value })[field], `validate ${field}`)
}
assert.equal(validateMessage('😀'.repeat(255), '正文').title, undefined)
assert.ok(validateTemplate({ ...draft(), messageType: 'action_card', buttonTitle: ' ' }).buttonTitle)
assert.ok(validateTemplate({ ...draft(), target: { kind: 'workflow_start' } })['target.definitionId'])
for (const url of ['http://example.test', 'https://user:pass@example.test', 'https://example.test/\npath', 'https://example.test/%0aevil', 'https://example.test/\u0085', 'https://example.test/%C2%85', 'https://example.test/%broken', 'https:example.test', 'https://example.test/' + 'a'.repeat(2048)]) {
  assert.ok(validateTemplate({ ...draft(), target: { kind: 'url', url } })['target.url'], url)
}
assert.equal(validateTemplate({ ...draft(), target: { kind: 'url', url: 'https://example.test/a?b=1#start' } })['target.url'], undefined)
const selected = { id: 7, name: '入职', enabled: true, title: '模板标题', content: '模板正文', messageType: 'oa', target: { kind: 'workflow_start', definitionId: 8, definitionName: '入职流程' }, buttonTitle: '查看', headColor: 'FF1677FF', version: 3, updatedAt: 1 }
const copied = newTemplateDraft(selected, true)
assert.ok(copied.name !== selected.name)
assert.equal(copied.id, undefined)
assert.equal(copied.version, undefined)
const payload = templatePayload(selected)
assert.equal(payload.target.definitionName, undefined)
assert.equal(payload.id, undefined)
assert.equal(payload.version, undefined)
assert.match(targetSummary(selected.target), /入职流程/)

const deferred = () => { let resolve, reject; const promise = new Promise((yes, no) => { resolve = yes; reject = no }); return { promise, resolve, reject } }
let candidates = [selected]
let optionFailure = false
let confirmCount = 0
let accept = true
const apiStub = { options: async () => { if (optionFailure) throw Error('offline'); return { data: { list: candidates, total: candidates.length } } } }
const { useSendTemplate } = load('src/views/notification/useSendTemplate.ts', { './templateHelpers': helpers, '@/api/notificationTemplates': { notificationTemplatesApi: apiStub } })
const form = vue.reactive({ title: '', content: '', scope: 'all', userIds: [], departmentIds: [] })
const chooser = useSendTemplate(form, async () => { confirmCount++; return accept })
assert.equal(chooser.canSend.value, true, 'untemplated sends do not depend on option loading')
optionFailure = true
await chooser.loadOptions()
assert.equal(chooser.canSend.value, true, 'untemplated sends survive option failures')
assert.equal(chooser.buildPayload().templateId, undefined)
optionFailure = false
await chooser.loadOptions()
assert.equal(chooser.canSend.value, true)
assert.equal(chooser.buildPayload().templateId, undefined)
await chooser.selectTemplate(7)
assert.equal(form.title, selected.title)
assert.equal(chooser.buildPayload().templateVersion, 3)
form.title = '仅本次修改'
accept = false
await chooser.selectTemplate(0)
assert.equal(chooser.selected.value.id, 7)
assert.equal(form.title, '仅本次修改')
assert.equal(confirmCount, 1)
optionFailure = true
await chooser.loadOptions()
assert.equal(chooser.canSend.value, false)
assert.equal(chooser.selected.value.id, 7)
assert.throws(() => chooser.buildPayload())
optionFailure = false
candidates = [{ ...selected, version: 4, title: '新版标题' }]
await chooser.loadOptions()
assert.equal(chooser.canSend.value, false)
assert.equal(chooser.selected.value.version, 3)
accept = true
await chooser.selectTemplate(7)
assert.equal(form.title, '新版标题')
form.content = '本次正文'
assert.equal(chooser.buildPayload().content, '本次正文')
assert.equal(chooser.selected.value.content, '模板正文')
assert.equal(chooser.buildPayload().templateVersion, 4)
candidates = []
await chooser.loadOptions()
assert.equal(chooser.canSend.value, false)
assert.equal(chooser.selected.value.id, 7)
await chooser.selectTemplate(0)
assert.equal(chooser.canSend.value, true)
assert.equal(chooser.buildPayload().templateId, undefined)
assert.equal(form.content, '')

// Candidate pagination must never silently omit templates beyond the first page.
const optionQueries = []
apiStub.options = async query => {
  optionQueries.push(query)
  return { data: { list: query.page === 1 ? [selected] : [{ ...selected, id: 9 }], total: 2 } }
}
await chooser.loadOptions()
assert.equal(chooser.options.value.length, 2)
assert.deepEqual(optionQueries.map(q => q.page), [1, 2])
await chooser.selectTemplate(9)
apiStub.options = async query => {
  if (query.page === 2) throw Error('second page offline')
  return { data: { list: [selected], total: 2 } }
}
await chooser.loadOptions()
assert.equal(chooser.canSend.value, false)
assert.equal(chooser.selected.value.id, 9)
assert.equal(chooser.options.value.length, 2)
let boundedCalls = 0
apiStub.options = async () => { boundedCalls++; return { data: { list: [selected], total: 2 } } }
await chooser.loadOptions()
assert.equal(chooser.canSend.value, false)
assert.ok(boundedCalls <= 100)
apiStub.options = async () => ({ data: { list: [selected], total: 1 } })
const confirmation = deferred()
const raceForm = vue.reactive({ title: '已修改', content: '内容', scope: 'all' })
const raceChooser = useSendTemplate(raceForm, () => confirmation.promise)
await raceChooser.loadOptions()
const pendingSelection = raceChooser.selectTemplate(7)
assert.equal(raceChooser.canSend.value, false, 'switch confirmation blocks even an untemplated send')
await raceChooser.loadOptions()
confirmation.resolve(true)
await pendingSelection
assert.equal(raceChooser.selected.value, null)
assert.equal(raceForm.title, '已修改')
assert.equal(raceChooser.switching.value, false)

// Explicitly abandoning a template remains possible after a candidate failure, with confirmation.
let fallbackAccepted = false
const fallbackForm = vue.reactive({ title: '', content: '', scope: 'all' })
const fallbackChooser = useSendTemplate(fallbackForm, async () => fallbackAccepted)
await fallbackChooser.loadOptions()
await fallbackChooser.selectTemplate(7)
fallbackForm.title = '保留我的草稿'
apiStub.options = async () => { throw Error('offline') }
await fallbackChooser.loadOptions()
let fallbackPromise
const fallbackSelector = await renderSelector({
  options: fallbackChooser.options.value, selected: fallbackChooser.selected.value, loading: false, switching: false,
  error: fallbackChooser.error.value, selectionError: '', edited: true, disabled: false,
  onSelect: id => { fallbackPromise = fallbackChooser.selectTemplate(id) },
})
const fallbackButton = fallbackSelector.buttons.find(button => button['aria-label'] === '不使用模板')
assert.ok(fallbackButton)
fallbackButton.onClick()
await fallbackPromise
assert.equal(fallbackChooser.selected.value.id, 7)
assert.equal(fallbackForm.title, '保留我的草稿')
assert.equal(fallbackChooser.canSend.value, false)
fallbackAccepted = true
fallbackButton.onClick()
await fallbackPromise
assert.equal(fallbackChooser.selected.value, null)
assert.equal(fallbackChooser.canSend.value, true)
apiStub.options = async () => ({ data: { list: [selected], total: 1 } })

// Execute send controller, capturing submissions without sending messages.
const deliveries = []
let permitted = true
let channel = 'dingtalk'
let sendPending
const adminApi = {
  dingTalkNotificationRecipientOptions: async () => ({ data: { users: [], departments: [] } }),
  inAppNotificationRecipientOptions: async () => ({ data: { users: [], departments: [] } }),
  dingTalkNotificationSend: async value => { deliveries.push(value); return sendPending ? sendPending.promise : { data: { sentCount: 1 } } },
  inAppNotificationSend: async value => { deliveries.push(value); return { data: { sentCount: 1 } } },
}
const { useNotificationSend } = load('src/views/notification/useNotificationSend.ts', { '@/api': { adminApi }, './templateHelpers': helpers, './useSendTemplate': { useSendTemplate } })
let replaceAccepted = true
const sender = useNotificationSend(() => channel, () => permitted, async () => replaceAccepted)
await sender.start()
assert.equal(deliveries.length, 0)
await sender.submit()
assert.ok(sender.errors.value.title)
await sender.templates.selectTemplate(7)
sender.form.title = '自定义标题'
sender.form.scope = 'departments'
sender.form.departmentIds = [8, 8, -1]
sendPending = deferred()
const sending = sender.submit()
await sender.submit()
assert.equal(deliveries.length, 1)
assert.equal(deliveries[0].templateId, 7)
assert.equal(deliveries[0].templateVersion, 3)
assert.equal(deliveries[0].title, '自定义标题')
assert.deepEqual(Array.from(deliveries[0].departmentIds), [8])
assert.equal(deliveries[0].target, undefined)
sendPending.reject({ msg: '所选模板已停用' })
await sending
assert.match(sender.sendError.value, /停用/)
assert.equal(sender.form.title, '自定义标题')
assert.equal(sender.templates.selected.value.id, 7)
sendPending = undefined

// Server-side version rejection must expose refresh even before the local cache knows it is stale.
const successfulSend = adminApi.dingTalkNotificationSend
adminApi.dingTalkNotificationSend = async () => { throw { msg: '模板版本已更新，请刷新后重试' } }
await sender.submit()
assert.match(sender.sendError.value, /版本/)
assert.equal(sender.templates.selectionError.value, '')
assert.equal(sender.templates.error.value, '')
let refreshPromise
const selector = await renderSelector({
  options: sender.templates.options.value, selected: sender.templates.selected.value, loading: false, switching: false,
  error: '', selectionError: '', edited: true, disabled: false,
  onRetry: () => { refreshPromise = sender.templates.loadOptions() },
})
const refresh = selector.buttons.find(button => button['aria-label'] === '刷新模板')
assert.ok(refresh, 'refresh is always rendered, including after a server-only version rejection')
assert.equal(refresh.disabled, false)
const freshOptions = deferred()
apiStub.options = () => freshOptions.promise
refresh.onClick()
assert.equal(sender.templates.loading.value, true)
assert.equal(sender.canSubmit.value, false, 'selected templates stay blocked while refreshing')
assert.equal(sender.form.title, '自定义标题')
freshOptions.resolve({ data: { list: [{ ...selected, version: 4, title: '新版标题', content: '新版正文' }], total: 1 } })
await refreshPromise
assert.equal(sender.templates.selected.value.version, 3)
assert.equal(sender.form.title, '自定义标题')
assert.equal(sender.form.content, '模板正文')
assert.equal(sender.canSubmit.value, false)
replaceAccepted = false
await sender.templates.selectTemplate(7)
assert.equal(sender.templates.selected.value.version, 3)
assert.equal(sender.form.title, '自定义标题')
assert.equal(sender.form.content, '模板正文')
replaceAccepted = true
await sender.templates.selectTemplate(7)
assert.equal(sender.templates.selected.value.version, 4)
assert.equal(sender.form.title, '新版标题')
assert.equal(sender.form.content, '新版正文')
adminApi.dingTalkNotificationSend = successfulSend
await sender.submit()
assert.equal(deliveries.at(-1).templateVersion, 4)

await sender.templates.selectTemplate(0)
sender.form.title = '无模板'
sender.form.content = '正文'
sender.form.scope = 'all'
const defaultOptions = deferred()
apiStub.options = () => defaultOptions.promise
const pendingDefaultOptions = sender.templates.loadOptions()
assert.equal(sender.canSubmit.value, true, 'default sends work during option loading')
await sender.submit()
assert.equal(deliveries.at(-1).title, '无模板')
assert.equal(deliveries.at(-1).templateId, undefined)
defaultOptions.reject(Error('options offline'))
await pendingDefaultOptions
assert.equal(sender.canSubmit.value, true, 'default sends work after option failure')
assert.equal(sender.form.title, '无模板')
assert.equal(sender.form.content, '正文')
const defaultSelector = await renderSelector({ options: [], selected: null, loading: false, switching: false, error: sender.templates.error.value, selectionError: '', edited: true, disabled: false })
assert.match(defaultSelector.html, /可直接按系统通知样式发送/)
assert.ok(defaultSelector.buttons.find(button => button['aria-label'] === '刷新模板'))
await sender.submit()
assert.equal(deliveries.at(-1).templateId, undefined)
assert.equal(deliveries.at(-1).templateVersion, undefined)
assert.equal(deliveries.at(-1).departmentIds, undefined)
channel = 'in_app'
await sender.start()
sender.form.title = '站内信'
sender.form.content = '正文'
sender.form.scope = 'users'
sender.form.userIds = [5]
await sender.submit()
assert.ok(deliveries.at(-1).requestId)
const deliveryCount = deliveries.length
permitted = false
await sender.submit()
assert.equal(deliveries.length, deliveryCount)

// Published workflow searches and selected-ID resolution ignore stale responses.
let workflowRequests = []
const workflowApi = { workflowOptions: query => { const pending = deferred(); workflowRequests.push({ query, pending }); return pending.promise } }
const { useTemplateWorkflowOptions } = load('src/views/notification/useTemplateWorkflowOptions.ts', { '@/api/notificationTemplates': { notificationTemplatesApi: workflowApi }, './templateHelpers': helpers })
const workflows = useTemplateWorkflowOptions()
const oldSearch = workflows.search('old')
const currentSearch = workflows.search('current', 2)
workflowRequests[1].pending.resolve({ data: { list: [{ id: 8, name: 'current' }], total: 40, page: 2, pageSize: 20 } })
await currentSearch
workflowRequests[0].pending.resolve({ data: { list: [{ id: 7, name: 'old' }], total: 1, page: 1, pageSize: 20 } })
await oldSearch
assert.equal(workflows.options.value[0].id, 8)
assert.equal(workflows.page.value, 2)
const resolveWorkflow = workflows.resolveSelected(8)
assert.equal(workflowRequests.at(-1).query.id, 8)
workflowRequests.at(-1).pending.resolve({ data: { list: [], total: 0, page: 1, pageSize: 20 } })
await resolveWorkflow
assert.match(workflows.selectedError.value, /未发布/)
const retryWorkflow = workflows.resolveSelected(8)
workflowRequests.at(-1).pending.resolve({ data: { list: [{ id: 8, name: 'published' }], total: 1, page: 1, pageSize: 20 } })
await retryWorkflow
assert.equal(workflows.selected.value.name, 'published')
assert.equal(workflows.selectedError.value, '')

const calls = []
let canEdit = true
let updatePending
const managerApi = {
  list: async query => { calls.push(['list', query]); return { data: { list: [selected], total: 21, page: query.page, pageSize: query.pageSize } } },
  create: async data => { calls.push(['create', data]); return { data: { ...data, id: 9, version: 1 } } },
  update: async (id, data) => { calls.push(['update', id, data]); if (updatePending) return updatePending.promise; return { data: { ...data, id, version: data.version + 1 } } },
  remove: async (id, version) => { calls.push(['remove', id, version]); return { data: null } },
}
const { useTemplateManager } = load('src/views/notification/useTemplateManager.ts', { './templateHelpers': helpers, '@/api/notificationTemplates': { notificationTemplatesApi: managerApi } })
const manager = useTemplateManager(() => canEdit, async () => accept)
await manager.load()
assert.equal(manager.items.value.length, 1)
assert.equal(manager.total.value, 21)
manager.page.value = 2
manager.keyword.value = '入职'
await manager.load()
assert.equal(calls.at(-1)[1].page, 2)
assert.equal(calls.at(-1)[1].keyword, '入职')
manager.openEditor(selected, true)
await manager.save(draft())
assert.equal(calls.find(c => c[0] === 'create')[0], 'create')
manager.openEditor(selected)
updatePending = deferred()
const firstSave = manager.save(draft())
await manager.save(draft())
assert.equal(calls.filter(c => c[0] === 'update').length, 1)
assert.equal(calls.at(-1)[2].version, 3)
updatePending.reject({ msg: '模板版本已更新' })
await firstSave
assert.equal(manager.editorVisible.value, true)
assert.match(manager.mutationError.value, /版本/)
updatePending = undefined
accept = false
await manager.remove(selected)
assert.equal(calls.filter(c => c[0] === 'remove').length, 0)
accept = true
await manager.remove(selected)
assert.deepEqual(calls.find(c => c[0] === 'remove'), ['remove', 7, 3])
await manager.toggle(selected)
assert.equal(calls.filter(c => c[0] === 'update').at(-1)[2].enabled, false)
assert.equal(calls.filter(c => c[0] === 'update').at(-1)[2].version, 3)
canEdit = false
const callCount = calls.length
manager.openEditor()
await manager.save(draft())
await manager.toggle(selected)
await manager.remove(selected)
assert.equal(calls.length, callCount)

// Execute the real API adapter against a request recorder: paths, JSON bodies and versions.
const requests = []
const request = Object.fromEntries(['get', 'post', 'put', 'delete'].map(method => [method, async (...args) => { requests.push([method, ...args]); return { data: {} } }]))
const { notificationTemplatesApi } = load('src/api/notificationTemplates.ts', { '../utils/request': { default: request, __esModule: true } })
await notificationTemplatesApi.options({ page: 1, pageSize: 100 })
await notificationTemplatesApi.workflowOptions({ keyword: '审批', page: 2, pageSize: 20, id: 8 })
await notificationTemplatesApi.create(draft())
await notificationTemplatesApi.update(7, { ...draft(), version: 3 })
await notificationTemplatesApi.remove(7, 3)
assert.equal(requests[0][1], '/api/v2/admin/dingtalk-notification-templates/options')
assert.equal(requests[0][2].params.pageSize, 100)
assert.equal(requests[1][1], '/api/v2/admin/dingtalk-notification-templates/workflow-options')
assert.equal(requests[1][2].params.id, 8)
assert.equal(requests[2][3].headers['Content-Type'], 'application/json')
assert.equal(requests[3][2].version, 3)
assert.equal(requests[4][2].params.version, 3)
console.log('Notification template runtime checks passed: validation, complete pagination, selection/cancellation, stale versions, send scopes/payloads, workflow searches, CRUD permissions, duplicate submits and API contracts')
