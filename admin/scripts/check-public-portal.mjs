import assert from 'node:assert/strict'
import { readFileSync, existsSync } from 'node:fs'
import vm from 'node:vm'
import { createRequire } from 'node:module'
import ts from 'typescript'
import * as vue from 'vue'
import { parse, compileScript, compileTemplate, compileStyle } from '@vue/compiler-sfc'

const require = createRequire(import.meta.url)
const plain = value => JSON.parse(JSON.stringify(value))

const read = path => readFileSync(new URL(`../src/${path}`, import.meta.url), 'utf8')
function execute(path, modules = {}, globals = {}) {
  return executeSource(read(path), path, modules, globals)
}
function executeSource(input, path, modules = {}, globals = {}) {
  const source = input.replaceAll('import.meta.env.VITE_API_BASE', "''")
  const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020, esModuleInterop: true } }).outputText
  const module = { exports: {} }
  vm.runInNewContext(compiled, {
    exports: module.exports, module, Error, URL, URLSearchParams, Headers, AbortController, setTimeout, clearTimeout,
    require: name => { assert.ok(name in modules, `unexpected import: ${name}`); return modules[name] },
    ...globals,
  }, { filename: path })
  return module.exports
}

assert.ok(existsSync(new URL('../src/views/public-portal/portalModel.ts', import.meta.url)), '公共入口必须提供真实的 URL、版本提交与状态规则')
const browserWindow = { location: { origin: 'http://localhost:5173' } }
const copiedLinks = []
const model = execute('views/public-portal/portalModel.ts', {}, {
  window: browserWindow,
  navigator: { clipboard: { writeText: async value => copiedLinks.push(value) } },
})
for (const url of ['javascript:alert(1)', 'data:text/html,x', '//evil.example', 'https://user:pass@example.com/sf/1', 'https://example.com/\npath']) {
  assert.equal(model.safeTargetUrl(url), '', `reject unsafe target ${url}`)
}
assert.equal(model.safeTargetUrl('https://forms.example.com/sf/1'), 'https://forms.example.com/sf/1')
const siteOrigins = ['http://localhost:5173', 'http://127.0.0.1:8080', 'http://192.168.2.1:5173', 'http://10.0.0.1', 'http://172.16.0.1', 'http://[::1]:5173', 'https://forms.example.com']
for (const origin of siteOrigins) {
  assert.equal(model.publicActivitiesUrl(origin), `${origin}/public/activities`, `use current site ${origin}`)
}
const invalidOrigins = [
  '', 'null', '//evil.example', 'javascript:alert(1)', 'ftp://example.com', 'https://user:pass@example.com',
  'https://@example.com', 'https://example.com/a', 'https://example.com?q=1', 'https://example.com/#x',
  'https://example.com?', 'https://example.com#', 'https://example.com\\evil', 'https://example.com\n',
  'https:///example.com', 'https://example.com/.', 'https://example.com/path/..',
]
for (const origin of invalidOrigins) {
  assert.equal(model.publicActivitiesUrl(origin), '', `reject invalid site origin ${origin}`)
}
assert.equal(model.publicActivitiesUrl(), 'http://localhost:5173/public/activities', 'default to browser origin without settings')
assert.equal(model.publicActivitiesUrl('https://forms.example.com/'), 'https://forms.example.com/public/activities')
assert.equal(model.publicActivitiesUrl('https://[2606:4700:4700::1111]'), 'https://[2606:4700:4700::1111]/public/activities')
for (const origin of siteOrigins) {
  for (const path of ['/sf/9', '/sf1/9', '/ef/12', '/ef1/12']) {
    assert.equal(model.publicActivityUrl(path, origin), `${origin}${path}`)
    for (const legacyOrigin of ['https://old.example.com', 'http://old.example.com:9000', 'https://evil.example']) {
      assert.equal(model.publicActivityUrl(`${legacyOrigin}${path}`, origin), `${origin}${path}`, 'legacy hosts never control the navigation origin')
    }
  }
}
const unsafeActivityUrls = [
  '', 'sf/9', '//evil.example/sf/9', '///evil.example/sf/9', 'javascript:alert(1)', 'data:text/html,x',
  'https://user:pass@evil.example/sf/9', 'https://@evil.example/sf/9', 'https://evil.example', 'https://evil.example/login',
  'http:evil.example/sf/9', 'https:///evil.example/sf/9', '/sf/9?next=https://evil.example', '/sf/9#x', '/sf/9?', '/sf/9#',
  'https://evil.example/sf/9?next=/login', 'https://evil.example/sf/9#x',
  '/sf/9\\evil', 'https://evil.example\\@forms.example.com/sf/9', '/sf/9\n', ' /sf/9',
  '/sf/0', '/sf/-1', '/sf/01', '/sf/1.2', '/sf/NaN', '/sf/9/', '/SF/9', '/admin',
  '/sf/%39', '/sf%2f9', '/sf/9%3fnext=evil', '/sf/../sf/9', '/sf/%2e%2e/sf/9',
  'https://evil.example/anything/../sf/9', 'https://evil.example/%2e%2e/sf/9',
]
for (const url of unsafeActivityUrls) assert.equal(model.publicActivityUrl(url), '', `reject unexpected filling URL ${url}`)
for (const origin of invalidOrigins) assert.equal(model.publicActivityUrl('/sf/9', origin), '', `reject invalid origin ${origin}`)
assert.equal(execute('views/public-portal/portalModel.ts').publicActivityUrl('/sf/9'), '', 'missing browser origin fails closed')
assert.match(model.formatActivityTime(Date.UTC(2026, 10, 3, 2)), /2026/)
assert.equal(model.safeLogoUrl('/uploads/logo.png'), '/uploads/logo.png')
assert.equal(model.safeLogoUrl('//evil.example/logo.png'), '')
assert.equal(model.safeLogoUrl('/\\evil.example/logo.png'), '')
assert.equal(model.isPreviewQuery('1'), true)
for (const value of [undefined, '', '0', 'true', ['1']]) assert.equal(model.isPreviewQuery(value), false)
const item = { id: 7, kind: 'survey', sourceId: 9, title: '  公开名称  ', description: '说明', enabled: true, sort: 2, version: 3, managementTitle: '内部名', displayTitle: '展示名', available: true, unavailableReason: '', canManage: true }
assert.deepEqual(JSON.parse(JSON.stringify(model.itemWritePayload(item))), { kind: 'survey', sourceId: 9, title: '公开名称', description: '说明', enabled: true, sort: 2, version: 3 })
assert.equal(model.canEditItem(true, item), true)
assert.equal(model.canEditItem(false, item), false)
assert.equal(model.canEditItem(true, { ...item, canManage: false }), false)
const deletedItem = { ...item, available: false, unavailableReason: '来源已删除', canManage: true }
assert.equal(model.canEditItem(true, deletedItem), false)
assert.equal(model.canRemoveItem(true, deletedItem), true)
assert.equal(model.canToggleItem(true, deletedItem), false)
assert.equal(model.canToggleItem(true, { ...item, available: false, enabled: true }), true)
assert.equal(model.canToggleItem(true, { ...item, available: false, enabled: false }), false)
const publicItem = { ...item, state: 'active', url: 'http://localhost:5173/sf/9', startTime: 0, endTime: 0, loginRequired: true, passwordRequired: true }
assert.equal(model.canOpenActivity(publicItem), true)
assert.equal(model.canOpenActivity({ ...publicItem, state: 'upcoming' }), false)
assert.equal(model.canOpenActivity({ ...publicItem, state: 'ended' }), false)
assert.equal(model.canOpenActivity({ ...publicItem, url: 'javascript:alert(1)' }), false)
assert.equal(model.filterActivities([publicItem], 'category:考试', '').length, 0)
assert.equal(model.filterActivities([publicItem], 'all', '公开').length, 1)
assert.equal(model.filterActivities([publicItem], 'all', '内部名').length, 1)
const sameNameItems = [
  { ...publicItem, id: 201, sourceId: 9, managementTitle: 'DISC-研发岗位', title: 'DISC测试', description: '同一说明' },
  { ...publicItem, id: 202, sourceId: 10, url: '/sf/10', managementTitle: 'DISC-销售岗位', title: 'DISC测试', description: '同一说明' },
  { ...publicItem, id: 203, kind: 'exam', sourceId: 9, category: '面试题', managementTitle: 'DISC-能力测验', title: 'DISC测试', description: '同一说明', url: '/ef/9' },
]
assert.equal(model.activityIdentity('survey', 9), '问卷编号：9')
assert.equal(model.activityIdentity('exam', 9), '考试编号：9')
assert.equal(model.activityIdentity('exam', 9, ' 面试题 '), '面试题编号：9')
assert.equal(model.activityIdentity('exam', 9, '  '), '考试编号：9')
assert.deepEqual(plain(model.filterActivities(sameNameItems, 'all', '面试题').map(item => item.id)), [203])
assert.deepEqual(plain(model.filterActivities(sameNameItems, 'category:问卷', '面试题').map(item => item.id)), [])
assert.deepEqual(plain(model.filterActivities(sameNameItems, 'all', '研发岗位').map(item => item.id)), [201])
assert.deepEqual(plain(model.filterActivities(sameNameItems, 'all', '问卷编号：9').map(item => item.id)), [201])
assert.deepEqual(plain(model.filterActivities(sameNameItems, 'all', '#10').map(item => item.id)), [202])
assert.deepEqual(plain(model.filterActivities(sameNameItems, 'category:面试题', '9').map(item => item.id)), [203])
assert.equal(typeof model.activityFilters, 'function', '公共页筛选必须使用当前内容的分类')
assert.deepEqual(plain(model.activityFilters(sameNameItems)), [
  { value: 'all', label: '全部' }, { value: 'category:问卷', label: '问卷' }, { value: 'category:面试题', label: '面试题' },
])
const multipleCategories = [...sameNameItems, { ...sameNameItems[2], id: 204, category: '培训测验' }, { ...sameNameItems[2], id: 205, category: '  ' }]
assert.deepEqual(plain(model.activityFilters(multipleCategories).map(option => option.label)), ['全部', '问卷', '面试题', '培训测验', '考试'])
assert.deepEqual(plain(model.filterActivities(multipleCategories, 'category:面试题', '').map(item => item.id)), [203])
assert.deepEqual(plain(model.filterActivities(multipleCategories, 'category:培训测验', '').map(item => item.id)), [204])
assert.deepEqual(plain(model.filterActivities(multipleCategories, 'category:考试', '').map(item => item.id)), [205])
assert.deepEqual(plain(model.activityFilters([])), [{ value: 'all', label: '全部' }])
assert.equal(model.filterActivities([{ ...publicItem, category: 'all' }], 'category:all', '').length, 1, '分类值不能与全部选项冲突')
assert.deepEqual(plain(model.activityFilters([{ ...publicItem, category: ' 面试题 ' }, sameNameItems[2]]).map(option => option.label)), ['全部', '面试题'], '同名分类合并展示')
assert.match(model.safeQrFilename('../A:/\\B\u0000', 'survey', 9), /^[^/\\:\u0000]+\.png$/u)

const requests = []
const validPublicData = { enabled: true, title: '活动', description: '', logoUrl: '', items: [{ ...publicItem, url: '/sf/9' }] }
let publicReply = { code: 0, msg: '', data: validPublicData }
const api = execute('api/publicPortal.ts', { '@/views/public-portal/portalModel': model }, {
  fetch: async (url, init) => { requests.push({ url, init }); return { ok: true, status: 200, json: async () => publicReply } },
  localStorage: { getItem() { throw new Error('匿名请求不得读取任何 token') } },
})
const activities = await api.fetchPublicActivities(undefined, 'https://api.example.com/')
assert.equal(activities.items[0].url, 'http://localhost:5173/sf/9', 'API base must not become the filling page origin')
assert.equal(validPublicData.items[0].url, '/sf/9', 'boundary mapping must not mutate the response')
publicReply = { code: 0, msg: '', data: { ...validPublicData, items: [{ ...validPublicData.items[0], category: '面试题' }] } }
assert.equal((await api.fetchPublicActivities()).items[0].category, '面试题', 'API boundary must retain source category')
publicReply = { code: 0, msg: '', data: { ...validPublicData, items: [{ ...validPublicData.items[0], category: {} }] } }
await assert.rejects(api.fetchPublicActivities(), /响应/, 'reject malformed category values')
publicReply = { code: 0, msg: '', data: validPublicData }
requests.splice(1)
assert.equal(requests[0].url, 'https://api.example.com/api/v2/public/activities')
assert.equal(requests[0].init.credentials, 'omit')
assert.equal(requests[0].init.cache, 'no-store')
assert.equal(new Headers(requests[0].init.headers).has('Authorization'), false)
await assert.rejects(api.fetchPublicPortalPreview('', undefined, ''), /管理员登录/)
assert.equal(requests.length, 1)
const previewData = await api.fetchPublicPortalPreview('test-admin-token', undefined, '')
assert.equal(previewData.items[0].url, activities.items[0].url)
assert.equal(requests.length, 2, 'preview performs exactly one request without configuration prerequisites')
assert.equal(requests[1].url, '/api/v2/admin/public-portal/preview')
assert.equal(new Headers(requests[1].init.headers).get('Authorization'), 'test-admin-token')
for (const origin of siteOrigins) {
  browserWindow.location.origin = origin
  publicReply = { code: 0, msg: '', data: { ...validPublicData, items: [
    ...['/sf/9', '/sf1/9', '/ef/12', '/ef1/12', 'https://old.example.com/sf/9'].map(url => ({ ...publicItem, url })),
    ...unsafeActivityUrls.map(url => ({ ...publicItem, url })),
  ] } }
  for (const result of [await api.fetchPublicActivities(), await api.fetchPublicPortalPreview('test-admin-token')]) {
    assert.deepEqual(plain(result.items.slice(0, 5).map(row => row.url)), ['/sf/9', '/sf1/9', '/ef/12', '/ef1/12', '/sf/9'].map(path => `${origin}${path}`))
    assert.ok(result.items.slice(5).every(row => row.url === ''), 'unsafe API targets disable every action')
  }
}
browserWindow.location.origin = 'http://localhost:5173'
publicReply = { code: 1, msg: '登录已过期', data: null }
await assert.rejects(api.fetchPublicPortalPreview('test-admin-token'), /登录已过期/)
publicReply = { code: 0, msg: '', data: { enabled: true, items: 'invalid' } }
await assert.rejects(api.fetchPublicActivities(), /响应/)

const apiWrites = []
const adminApi = execute('api/publicPortal.ts', {
  '@/views/public-portal/portalModel': model,
  '@/utils/request': { __esModule: true, default: Object.fromEntries(['get', 'post', 'put', 'delete'].map(method => [method, async (...args) => { apiWrites.push([method, ...args]); return { data: {} } }])) },
}).publicPortalApi
await adminApi.update(7, model.itemWritePayload(item))
assert.equal(apiWrites[0][1], '/api/v2/admin/public-portal/items/7')
assert.equal(apiWrites[0][3].headers['Content-Type'], 'application/json')
assert.equal(JSON.parse(apiWrites[0][3].transformRequest[0](apiWrites[0][2])).version, 3)
await adminApi.remove(7, 3)
assert.equal(apiWrites[1][2].params.version, 3)
await adminApi.options({ kind: 'exam', keyword: '公开', page: 2, pageSize: 10 })
assert.equal(apiWrites[2][2].params.kind, 'exam')
assert.equal(apiWrites[2][2].params.page, 2)
const defaults = { enabled: false, title: '公共入口', description: '', logoUrl: '', showEnded: false, version: 4 }
assert.equal(model.settingsError({ ...defaults, enabled: true }), '', 'enabling requires no domain')
await adminApi.saveSettings({ ...defaults, enabled: true, publicBaseUrl: 'https://obsolete.example.com' })
assert.deepEqual(plain(apiWrites.at(-1)[2]), { ...defaults, enabled: true }, 'settings payload drops legacy server fields')
assert.deepEqual(JSON.parse(apiWrites.at(-1)[3].transformRequest[0](apiWrites.at(-1)[2])), { ...defaults, enabled: true })

// Execute real Vue setup/composable code without a browser or production API.
const compiledScripts = new Map()
for (const file of ['public-portal/index.vue', 'public-portal/SettingsPanel.vue', 'public-portal/SourcePicker.vue', 'public-portal/ItemDialog.vue', 'public/Activities.vue', 'public/ActivityCard.vue', 'public/ActivityQr.vue']) {
  const path = `views/${file}`
  const source = read(path)
  const { descriptor, errors } = parse(source, { filename: path })
  assert.deepEqual(errors, [])
  const script = compileScript(descriptor, { id: path })
  compiledScripts.set(path, script.content)
  assert.deepEqual(compileTemplate({ source: descriptor.template.content, filename: path, id: path, compilerOptions: { bindingMetadata: script.bindings } }).errors, [])
  for (const style of descriptor.styles) assert.deepEqual(compileStyle({ source: style.content, filename: path, id: path, scoped: style.scoped }).errors, [])
  assert.ok(!source.includes('v-html'), `${path} must render untrusted copy as text`)
}
const errors = []
const node = (tag = '', text = '') => ({ tag, text, props: {}, children: [], parent: null })
const renderer = vue.createRenderer({
  createElement: tag => node(tag), createText: text => node('#text', text), createComment: () => node('#comment'),
  setText(value, text) { value.text = text }, setElementText(value, text) { value.text = text; value.children = [] },
  patchProp(value, key, _previous, next) { value.props[key] = next },
  parentNode: value => value.parent, nextSibling: () => null,
  insert(child, parent) { child.parent = parent; parent.children.push(child) },
  remove(child) { if (child.parent) child.parent.children = child.parent.children.filter(value => value !== child) },
})
const settle = async () => { for (let i = 0; i < 12; i++) { await vue.nextTick(); await Promise.resolve() } }
function mountSetup(setup, props = {}, emit = () => {}) {
  let state
  const app = renderer.createApp({ setup: () => { state = vue.proxyRefs(setup(props, { expose() {}, emit })); return () => null } })
  app.config.errorHandler = cause => errors.push(cause)
  app.config.warnHandler = cause => errors.push(cause)
  app.mount(node())
  return { get state() { return state }, app }
}
function mountSfc(path, modules, props = {}, globals = {}, emit = () => {}) {
  const component = executeSource(compiledScripts.get(path), path, { vue, ...modules }, globals).default
  return mountSetup(component.setup, props, emit)
}
const messages = []
const element = { ElMessage: { success: message => messages.push(message), error: message => messages.push(message) }, ElMessageBox: { confirm: async () => {} } }
const settingsEvents = []
const settingsPanel = mountSfc('views/public-portal/SettingsPanel.vue', { './portalModel': model }, { settings: defaults, editable: true, busy: false }, {}, (...args) => settingsEvents.push(plain(args)))
settingsPanel.state.draft.enabled = true
await settingsPanel.state.submit()
assert.deepEqual(settingsEvents, [['save', { ...defaults, enabled: true }]], 'the form can enable and save without a domain field')
assert.equal(settingsPanel.state.validationError, '')
settingsPanel.app.unmount()
for (const path of ['types/publicPortal.ts', 'views/public-portal/SettingsPanel.vue', 'views/public-portal/index.vue']) {
  assert.doesNotMatch(read(path), /publicBaseUrl|公共域名/, `${path} must remove the obsolete domain UI/contract`)
}
let allowed = false
let configReads = 0
let writeError = null
let readError = null
const mutations = []
let server = { settings: { ...defaults }, items: [plain(item)] }
const managerApi = {
  get: async () => { configReads++; if (readError) throw readError; return { data: plain(server) } },
  saveSettings: async data => { mutations.push(['settings', plain(data)]); if (writeError) throw writeError; server.settings = { ...data, version: data.version + 1 } },
  create: async data => { mutations.push(['create', plain(data)]); if (writeError) throw writeError },
  update: async (id, data) => { mutations.push(['update', id, plain(data)]); if (writeError) throw writeError; server.items = server.items.map(row => row.id === id ? { ...row, ...data, version: row.version + 1 } : row) },
  remove: async (id, version) => { mutations.push(['remove', id, version]); if (writeError) throw writeError },
}
const { usePublicPortalAdmin } = execute('views/public-portal/usePublicPortalAdmin.ts', {
  vue, 'element-plus': element, '@/api/publicPortal': { publicPortalApi: managerApi },
  '@/utils/permission': { hasPerm: key => key === model.PORTAL_EDIT_PERMISSION && allowed }, './portalModel': model,
})
const manager = mountSetup(usePublicPortalAdmin)
await settle()
assert.equal(configReads, 1, 'read-only admins still load config')
assert.equal(manager.state.settings.title, '公共入口')
assert.equal(manager.state.publicUrl, 'http://localhost:5173/public/activities')
manager.state.openCreate(); manager.state.openEdit(item)
await manager.state.saveSettings(defaults); await manager.state.saveItem(item)
await manager.state.changeItem(item, { enabled: false }); await manager.state.removeItem(item)
assert.equal(mutations.length, 0, 'all handlers enforce edit permission')
assert.equal(manager.state.dialogVisible, false)
allowed = true
const unmanageable = { ...item, canManage: false }
manager.state.openEdit(unmanageable)
await manager.state.changeItem(unmanageable, { sort: 5 }); await manager.state.removeItem(unmanageable)
assert.equal(mutations.length, 0, 'canManage false blocks mutation independently of menu permission')
manager.state.openEdit(deletedItem)
await manager.state.changeItem(deletedItem, { enabled: false })
await manager.state.changeItem(deletedItem, { sort: 15 })
assert.equal(mutations.length, 0, 'deleted sources cannot be edited, sorted or toggled')
await manager.state.removeItem(deletedItem)
assert.deepEqual(mutations.pop(), ['remove', 7, 3], 'creator/superadmin may remove a deleted source')
await manager.state.changeItem({ ...item, available: false, enabled: false }, { enabled: true })
assert.equal(mutations.length, 0, 'unavailable content cannot be published')
manager.state.openEdit(item)
writeError = { msg: '配置已更新，请刷新后重试' }
await manager.state.saveItem(model.itemWritePayload(item))
assert.equal(manager.state.dialogVisible, true)
assert.equal(manager.state.editing.version, 3)
assert.match(manager.state.actionError, /配置已更新/)
writeError = null
await manager.state.saveItem({ ...model.itemWritePayload(item), title: '新标题' })
assert.equal(mutations.at(-1)[2].version, 3)
assert.equal(manager.state.dialogVisible, false)
assert.equal(manager.state.sortedItems[0].version, 4)
await manager.state.changeItem(manager.state.sortedItems[0], { sort: 25 })
assert.equal(mutations.at(-1)[2].version, 4)
assert.equal(mutations.at(-1)[2].sort, 25)
await manager.state.changeItem(manager.state.sortedItems[0], { enabled: false })
assert.equal(mutations.at(-1)[2].version, 5)
assert.equal(mutations.at(-1)[2].enabled, false)
await manager.state.removeItem(manager.state.sortedItems[0])
assert.deepEqual(mutations.at(-1), ['remove', 7, 6])
await manager.state.saveSettings({ ...defaults, enabled: true, title: ' 公共入口 ', publicBaseUrl: 'https://obsolete.example.com/' })
assert.deepEqual(mutations.at(-1), ['settings', { ...defaults, enabled: true }], 'enable without forwarding old config fields')
assert.equal(manager.state.settings.enabled, true)
assert.equal(manager.state.settings.version, 5)
readError = { msg: '请先执行数据库迁移' }
await manager.state.load()
assert.match(manager.state.loadError, /数据库迁移/)
const beforeBlocked = mutations.length
await manager.state.changeItem(item, { sort: 10 })
assert.equal(mutations.length, beforeBlocked)
readError = null; await manager.state.load()
assert.equal(manager.state.loadError, '')
manager.app.unmount()

const optionRequests = []
const pendingOptions = []
const sourceProps = vue.reactive({ kind: 'survey', selectedId: 0, disabled: false })
const picker = mountSfc('views/public-portal/SourcePicker.vue', {
  '@/components/admin-ui': {}, './portalModel': model,
  '@/api/publicPortal': { publicPortalApi: { options: query => { optionRequests.push(plain(query)); return new Promise(resolve => pendingOptions.push(resolve)) } } },
}, sourceProps)
await settle()
assert.equal(optionRequests[0].kind, 'survey')
sourceProps.kind = 'exam'; await settle()
pendingOptions[1]({ data: { list: [{ kind: 'exam', sourceId: 12 }], total: 31 } }); await settle()
pendingOptions[0]({ data: { list: [{ kind: 'survey', sourceId: 11 }], total: 1 } }); await settle()
assert.equal(picker.state.list[0].kind, 'exam', 'stale source responses cannot overwrite newer kind results')
picker.state.keyword = '安全培训'; picker.state.page = 3
const searching = picker.state.load(); await settle()
assert.deepEqual(optionRequests.at(-1), { kind: 'exam', keyword: '安全培训', page: 3, pageSize: 10 })
pendingOptions.at(-1)({ data: { list: [], total: 31 } }); await searching
assert.equal(picker.state.total, 31)
picker.app.unmount()

let anonymousReads = 0, previewReads = 0, tokenReads = 0
const route = vue.reactive({ query: {} })
const publicData = { enabled: true, title: '<img onerror=alert(1)>', description: '<script>unsafe</script>', logoUrl: 'javascript:alert(1)', items: [publicItem, sameNameItems[2]] }
const doc = { title: 'Admin' }
const publicPage = mountSfc('views/public/Activities.vue', {
  'vue-router': { useRoute: () => route }, '../public-portal/portalModel': model, './ActivityCard.vue': {},
  '@/api/publicPortal': { fetchPublicActivities: async () => { anonymousReads++; return publicData }, fetchPublicPortalPreview: async token => { previewReads++; if (!token) throw new Error('预览需要管理员登录'); return publicData } },
}, {}, { document: doc, localStorage: { getItem: () => { tokenReads++; return '' } } })
await settle()
assert.equal(anonymousReads, 1)
assert.equal(tokenReads, 0)
assert.equal(previewReads, 0)
assert.equal(publicPage.state.logo, '')
assert.equal(doc.title, publicData.title)
assert.equal(publicPage.state.categorySummary, '问卷 · 面试题')
publicPage.state.category = 'category:面试题'; await settle()
assert.deepEqual(plain(publicPage.state.filtered.map(item => item.id)), [203])
await publicPage.state.load(); await settle()
assert.equal(publicPage.state.category, 'category:面试题', '刷新时仍存在的选中分类应保留')
publicData.items = [publicItem]
await publicPage.state.load(); await settle()
assert.equal(publicPage.state.category, 'all', '分类移除或改名后回到全部，避免列表卡在失效筛选')
assert.equal(publicPage.state.categorySummary, '问卷')
route.query.preview = '0'; await settle(); assert.equal(previewReads, 0)
route.query.preview = '1'; await settle()
assert.equal(tokenReads, 1)
assert.match(publicPage.state.error, /管理员登录/)
route.query.preview = undefined; await settle()
assert.equal(anonymousReads, 4)
assert.equal(publicPage.state.error, '')
publicPage.app.unmount()
assert.equal(doc.title, 'Admin')

// Encode with the installed library, inspect actual PNG pixels and independently decode them.
const qr = execute('views/public-portal/portalQr.ts', { './portalModel': model, qrcode: require('qrcode') })
const { PNG } = require('pngjs')
const ZXing = require('html5-qrcode/third_party/zxing-js.umd.js')
const pngUrl = await qr.generateActivityQr(publicItem.url)
const png = PNG.sync.read(Buffer.from(pngUrl.split(',')[1], 'base64'))
assert.equal(png.width, 1024); assert.equal(png.height, 1024)
assert.equal(qr.QR_OPTIONS.margin, 4)
const pixelCount = png.width * png.height
for (let i = 0; i < pixelCount; i++) {
  const offset = i * 4
  assert.ok(png.data[offset] === 0 || png.data[offset] === 255, 'black and white PNG only')
  assert.equal(png.data[offset], png.data[offset + 1]); assert.equal(png.data[offset], png.data[offset + 2]); assert.equal(png.data[offset + 3], 255)
  if (i < 1024 || i >= pixelCount - 1024 || i % 1024 === 0 || i % 1024 === 1023) assert.equal(png.data[offset], 255, 'quiet zone reaches every edge')
}
function decodeQr(dataUrl) {
  const image = PNG.sync.read(Buffer.from(dataUrl.split(',')[1], 'base64'))
  const rgb = new Int32Array(image.width * image.height)
  for (let i = 0; i < rgb.length; i++) rgb[i] = (image.data[i * 4] << 16) | (image.data[i * 4 + 1] << 8) | image.data[i * 4 + 2]
  const bitmap = new ZXing.BinaryBitmap(new ZXing.HybridBinarizer(new ZXing.RGBLuminanceSource(rgb, image.width, image.height)))
  return new ZXing.QRCodeReader().decode(bitmap, new Map([[ZXing.DecodeHintType.PURE_BARCODE, true]])).getText()
}
assert.equal(decodeQr(pngUrl), publicItem.url)
for (const url of ['/sf/9', '//evil.example/sf/9', 'javascript:alert(1)', 'https://user:pass@example.com/sf/9']) {
  await assert.rejects(qr.generateActivityQr(url), /链接无效/, 'QR encoder requires an absolute HTTP(S) URL')
}

// Render the real card and QR templates; only the Element Plus shell is replaced.
// Exercise their actual event handlers and decode the image passed to download.
function inlineSfc(path, modules) {
  const { descriptor } = parse(read(path), { filename: path })
  const script = compileScript(descriptor, { id: path, inlineTemplate: true })
  return executeSource(script.content, path, { vue, ...modules }).default
}
function descendants(root) { return [root, ...root.children.flatMap(descendants)] }
function textContent(root) { return root.text + root.children.map(textContent).join('') }
const cardDownloads = []
const qrTasks = []
const renderedQr = inlineSfc('views/public/ActivityQr.vue', {
  'element-plus': element, '../public-portal/portalModel': model,
  '../public-portal/portalQr': {
    ...qr,
    generateActivityQr(url) { const task = qr.generateActivityQr(url); qrTasks.push(task.catch(() => {})); return task },
    downloadActivityQr: (dataUrl, filename) => cardDownloads.push({ dataUrl, filename }),
  },
})
const renderedCard = inlineSfc('views/public/ActivityCard.vue', {
  'element-plus': element, '../public-portal/portalModel': model, './ActivityQr.vue': renderedQr,
})
function mountCard(item) {
  const root = node()
  const app = renderer.createApp(renderedCard, { item })
  for (const [name, tag] of [['ElButton', 'button'], ['ElTag', 'span'], ['ElDialog', 'section']]) {
    app.component(name, { setup: (_, { attrs, slots }) => () => vue.h(tag, attrs, [slots.default?.(), slots.footer?.()]) })
  }
  app.config.errorHandler = cause => errors.push(cause)
  app.config.warnHandler = cause => errors.push(cause)
  app.mount(root)
  return { root, app }
}
for (const origin of ['http://localhost:5173', 'http://192.168.2.1:8080', 'https://forms.example.com']) {
  browserWindow.location.origin = origin
  publicReply = { code: 0, msg: '', data: validPublicData }
  const normalizedItem = (await api.fetchPublicActivities(undefined, 'https://api.example.com')).items[0]
  const card = mountCard(normalizedItem)
  await Promise.all(qrTasks.splice(0)); await settle()
  const nodes = descendants(card.root)
  const link = nodes.find(value => value.tag === 'a')
  assert.equal(link.props.href, `${origin}/sf/9`)
  let prevented = false
  link.props.onClick({ preventDefault() { prevented = true } })
  assert.equal(prevented, false)
  const copyButton = nodes.find(value => value.tag === 'button' && textContent(value) === '复制链接')
  assert.equal(copyButton.props.disabled, false)
  await copyButton.props.onClick()
  assert.equal(copiedLinks.at(-1), link.props.href)
  const qrImage = descendants(card.root).find(value => value.tag === 'img')
  assert.equal(decodeQr(qrImage.props.src), link.props.href, 'rendered QR, copy and entry share the current-origin URL')
  const downloadButton = nodes.find(value => value.tag === 'button' && textContent(value) === '下载 PNG')
  await downloadButton.props.onClick()
  assert.equal(decodeQr(cardDownloads.at(-1).dataUrl), link.props.href)
  card.app.unmount()
}
browserWindow.location.origin = 'http://localhost:5173'
for (const sameNameItem of sameNameItems) {
  const card = mountCard({ ...sameNameItem, description: sameNameItem.category ? '' : sameNameItem.description })
  await Promise.all(qrTasks.splice(0)); await settle()
  const identity = model.activityIdentity(sameNameItem.kind, sameNameItem.sourceId, sameNameItem.category)
  const nodes = descendants(card.root)
  assert.ok(textContent(card.root).includes(identity), 'same-name cards display the source identity')
  assert.equal(textContent(nodes.find(value => value.tag === 'h2')), 'DISC测试', 'display title remains the configured public title')
  assert.ok(textContent(card.root).includes(sameNameItem.managementTitle), 'management name distinguishes same display titles')
  assert.equal(textContent(nodes.find(value => value.props.class === 'kind-label')), sameNameItem.category || model.kindLabels[sameNameItem.kind], 'card badge must follow the source category with a type fallback')
  if (sameNameItem.category) {
    const description = textContent(nodes.find(value => value.props.class === 'description'))
    assert.doesNotMatch(description, /考试/, '自定义分类不得继续显示固定的考试说明')
    assert.match(description, /按页面提示/)
  }
  const thumbnail = nodes.find(value => value.tag === 'button' && value.props['aria-label']?.startsWith('放大'))
  assert.ok(thumbnail.props['aria-label'].includes(identity), 'same-name QR controls have distinct labels')
  thumbnail.props.onClick(); await settle()
  assert.ok(descendants(card.root).some(value => value.props.title?.includes(identity)), 'QR preview identifies the exact source')
  const downloadButton = nodes.find(value => value.tag === 'button' && textContent(value) === '下载 PNG')
  await downloadButton.props.onClick()
  assert.equal(cardDownloads.at(-1).filename, `${sameNameItem.kind}-${sameNameItem.sourceId}-${sameNameItem.managementTitle}.png`)
  assert.equal(decodeQr(cardDownloads.at(-1).dataUrl), model.publicActivityUrl(sameNameItem.url), 'same-name downloads must target their own source')
  card.app.unmount()
}
for (const url of ['https://evil.example/login', '//evil.example/sf/9', 'javascript:alert(1)']) {
  const card = mountCard({ ...publicItem, url })
  await Promise.all(qrTasks.splice(0)); await settle()
  const nodes = descendants(card.root)
  assert.ok(!nodes.some(value => value.tag === 'a'), 'unsafe links cannot navigate even if the card bypasses the API')
  assert.ok(!nodes.some(value => value.tag === 'img'), 'unsafe links cannot produce a QR code')
  const copyButton = nodes.find(value => value.tag === 'button' && textContent(value) === '复制链接')
  assert.equal(copyButton.props.disabled, true)
  const beforeCopy = copiedLinks.length
  await copyButton.props.onClick()
  assert.equal(copiedLinks.length, beforeCopy)
  card.app.unmount()
}
let qrCalls = 0, downloads = 0
const qrPanel = mountSfc('views/public/ActivityQr.vue', {
  'element-plus': element, '../public-portal/portalModel': model,
  '../public-portal/portalQr': { generateActivityQr: async () => { qrCalls++; if (qrCalls === 1) throw new Error('canvas failure'); return pngUrl }, downloadActivityQr: () => { downloads++ } },
}, { url: publicItem.url, title: publicItem.title, kind: 'survey', sourceId: 9 })
await settle()
assert.match(qrPanel.state.error, /canvas failure/)
assert.equal(qrPanel.state.dataUrl, '')
assert.equal(downloads, 0)
await qrPanel.state.generate()
assert.equal(qrPanel.state.dataUrl, pngUrl)
assert.equal(qrPanel.state.error, '')
await qrPanel.state.download()
assert.equal(downloads, 1, 'upcoming activities can use the same download path without a state gate')
qrPanel.app.unmount()
assert.deepEqual(errors, [])
console.log('public portal passed: domain-free settings, current-origin URL allowlist, single-request preview, DTO writes/version conflicts, permissions/readonly, source search races/pagination, anonymous isolation, schema errors, rendered entry/copy/QR/download agreement, real 1024px black/white PNG decode and retry')
