import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')
const helpers = executeTypeScriptModule(read('src/views/log/auditLogHelpers.ts'), 'auditLogHelpers.ts')
const { actorLabel, outcomeMeta, buildAuditQuery, emptyAuditFilters, auditCsv, safeChangeValue, exportSummary, createRequestGate, relatedLogRoute } = helpers

// Old records never acquire a successful result; accepted async work is unfinished.
assert.equal(outcomeMeta().label, '未知')
assert.equal(outcomeMeta('new-future-value').label, '未知')
assert.equal(helpers.affectedCountLabel(0), '-')
assert.equal(helpers.affectedCountLabel(undefined), '-')
assert.equal(helpers.affectedCountLabel(3), '3')
const noTargets = helpers.requestedTargetsSummary({})
assert.equal(noTargets.text, '未提供')
assert.equal(noTargets.countLabel, '-')
assert.equal(noTargets.notice, '')
const partialTargets = helpers.requestedTargetsSummary({ targetIds: ['u1', 'u2'], requestedCount: 4, affectedCount: 1 })
assert.equal(partialTargets.text, 'u1, u2')
assert.equal(partialTargets.countLabel, '4')
assert.match(partialTargets.notice, /已截断/)
assert.equal(helpers.requestedTargetsSummary({ targetIds: ['u1', 'u2'], requestedCount: 2 }).notice, '')
assert.equal(helpers.requestedTargetsSummary({ targetIds: ['u1'], affectedCount: 1 }).countLabel, '-', 'never derive requested count from target IDs or affected count')
const exportTargets = helpers.requestedTargetsSummary({ action: 'export', targetIds: Array.from({ length: 1000 }, (_, i) => String(i)), affectedCount: 1001 })
assert.equal(exportTargets.countLabel, '-')
assert.match(exportTargets.notice, /导出 1001.*1000.*不完整/)
assert.equal(helpers.requestedTargetsSummary({ targetIds: [], requestedCount: 0 }).countLabel, '0')
assert.match(helpers.requestedTargetsSummary({ requestedCount: 5 }).notice, /未提供/)
const boundedTargets = helpers.requestedTargetsSummary({ targetIds: Array.from({ length: 1001 }, (_, i) => `id-${i}`), requestedCount: 1001 })
assert.ok(!boundedTargets.text.includes('id-1000'))
assert.match(boundedTargets.notice, /展示已截断/)
const longTarget = helpers.requestedTargetsSummary({ targetIds: ['x'.repeat(20000)], requestedCount: 1 })
assert.ok(longTarget.text.length <= 16001)
assert.match(longTarget.notice, /展示已截断/)
assert.equal(helpers.moduleLabel('organization'), '组织架构')
assert.equal(helpers.actionLabel({action: 'password_login'}), '密码登录')
assert.equal(helpers.moduleOptions.length, 18)
assert.equal(outcomeMeta('accepted').label, '已受理（未完成）')
assert.notEqual(outcomeMeta('accepted').type, 'success')
assert.match(actorLabel({ adminId: 0, adminName: 'forged-admin', attemptedAccount: 'root' }), /匿名|未确认/)
assert.doesNotMatch(actorLabel({ adminId: null, adminName: 'forged-admin' }), /forged-admin/)
assert.doesNotMatch(actorLabel({ adminId: 7, actorType: 'anonymous', adminName: 'forged-admin' }), /forged-admin/)
assert.match(actorLabel({ adminId: 7, actorType: 'user', adminName: '张三' }), /用户.*张三.*7/)
assert.equal(actorLabel({ actorType: 'system' }), '系统')

const filters = { ...emptyAuditFilters(), search: ' alice ', source: 'h5app', module: ' workflow ', action: ' update ', outcome: 'failure', actorId: '7', ip: ' 127.0.0.1 ', range: ['2026-10-01T00:00:00+08:00', '2026-10-28T23:59:59+08:00'] }
const query = buildAuditQuery(filters, 2, 50)
assert.equal(query.search, 'alice')
assert.equal(query.module, 'workflow')
assert.equal(query.action, 'update')
assert.equal(query.source, 'h5app')
assert.equal(query.outcome, 'failure')
assert.equal(query.actorId, '7')
assert.equal(query.ip, '127.0.0.1')
assert.equal(query.startTime, filters.range[0])
assert.equal(query.endTime, filters.range[1])
assert.equal(query.page, 2)
assert.equal(query.pageSize, 50)
assert.equal(buildAuditQuery(emptyAuditFilters(), 1, 20).startTime, undefined)
assert.throws(() => buildAuditQuery({ ...filters, actorId: '-1' }, 1, 20), /ID/)
assert.throws(() => buildAuditQuery({ ...filters, range: [...filters.range].reverse() }, 1, 20), /时间/)
assert.throws(() => buildAuditQuery({ ...filters, range: ['invalid', 'invalid'] }, 1, 20), /时间/)

// CSV quotes and formula controls apply to every untrusted cell, including leading whitespace.
for (const payload of ['=1+1', '+cmd', '-2+3', '@SUM(A1)', ' \t=HYPERLINK("x")', '\r=1', '\tplain', '＝1+1']) {
  assert.ok(auditCsv([{ id: 1, content: payload }]).includes(`"'${payload.replaceAll('"', '""')}"`), payload)
}
assert.ok(auditCsv([{ id: 1, content: '逗号,换行\n"引号"' }]).includes('"逗号,换行\n""引号"""'))
assert.ok(auditCsv([]).startsWith('\uFEFF'))
assert.equal(safeChangeValue('password', 'do-not-show'), '[已隐藏]')
assert.equal(safeChangeValue('password', '[已变更]'), '[已变更]')
assert.match(safeChangeValue('settings', {token: '[已变更]'}), /已变更/)
assert.equal(safeChangeValue('credentials.token', 'do-not-show'), '[已隐藏]')
const changed = safeChangeValue('settings', { enabled: false, count: 0, password: 'do-not-show', nested: { accessToken: 'also-secret' } })
assert.match(changed, /false/)
assert.match(changed, /0/)
assert.doesNotMatch(changed, /do-not-show|also-secret/)
assert.doesNotMatch(safeChangeValue('settings', '{"cookie":"private"}'), /private/)
assert.doesNotMatch(safeChangeValue('settings', {apiKey: 'api-secret', nested: '{"token":"nested-secret"}'}), /api-secret|nested-secret/)
assert.match(safeChangeValue('settings', Array.from({length: 51}, (_, i) => i)), /省略/)
assert.equal(buildAuditQuery({...emptyAuditFilters(), actorId: '0'}, 1, 20).actorId, '0')
assert.ok(safeChangeValue('data', 'x'.repeat(10000)).length < 6500)
assert.doesNotMatch(auditCsv([{ id: 1, changes: [{ field: 'password', before: 'old-secret', after: 'new-secret' }] }]), /old-secret|new-secret/)

assert.equal(exportSummary({ list: Array(1000).fill({ id: 1 }), total: 1042, page: 1, pageSize: 1000 }).truncated, true)
assert.equal(exportSummary({ list: [{ id: 1 }], total: 1, page: 1, pageSize: 1000 }).truncated, false)
assert.throws(() => exportSummary({ list: Array(1001).fill({ id: 1 }), total: 1001, page: 1, pageSize: 1000 }), /1000/)
assert.throws(() => exportSummary({ list: [{ id: 1 }], total: 0, page: 1, pageSize: 1000 }), /总数/)
assert.equal(relatedLogRoute('untrusted', 'javascript:alert(1)', () => true), null)
assert.equal(relatedLogRoute('workflow_instance', '42', () => false), null)
assert.equal(relatedLogRoute('workflow_instance', '42', key => key === 'admin:menu:workflow:instance:list'), null)
assert.equal(relatedLogRoute('workflow_instance', '42', () => true)?.path, '/workflow/instances')

// Deferred responses model an old request finishing after the newer one and a cancelled request.
const gate = createRequestGate()
let rendered = ''
let finishOld
const oldTicket = gate.next()
const oldRequest = new Promise(resolve => { finishOld = resolve }).then(value => { if (gate.isCurrent(oldTicket)) rendered = value })
const newTicket = gate.next()
if (gate.isCurrent(newTicket)) rendered = 'new'
finishOld('old')
await oldRequest
assert.equal(rendered, 'new')
gate.invalidate()
assert.equal(gate.isCurrent(newTicket), false)

// Exercise the actual API wrappers: every audit endpoint is read-only with unchanged filters.
const calls = []
const api = executeTypeScriptModule(read('src/api/index.ts'), 'api/index.ts', {
  '../utils/request': { get: (url, config) => { calls.push({ url, config }); return Promise.resolve({ code: 0, data: {} }) } },
}).adminApi
await api.logList(query)
await api.loginLogList(query)
await api.logDetail('a/b')
await api.loginLogDetail('a/b')
await api.logExport(query)
await api.loginLogExport(query)
assert.deepEqual(calls.map(call => call.url), [
  '/api/v2/admin/logs', '/api/v2/admin/login-logs', '/api/v2/admin/logs/a%2Fb', '/api/v2/admin/login-logs/a%2Fb',
  '/api/v2/admin/logs/export', '/api/v2/admin/login-logs/export',
])
for (const index of [0, 1, 4, 5]) assert.equal(calls[index].config.params, query)
assert.equal(api.logClear, undefined)
assert.equal(api.logDels, undefined)

const page = read('src/views/log/index.vue')
const state = read('src/views/log/useAuditLogs.ts')
const detail = read('src/views/log/AuditLogDetail.vue')
const filterView = read('src/views/log/AuditLogFilters.vue')
const filterForm = filterView.match(/<el-form\b[\s\S]*?<\/el-form>/)?.[0] || ''
assert.doesNotMatch(filterForm, /label-position="top"|<el-form-item\s+label=/, 'audit filters must not render visible top labels')
for (const label of ['账号 / 关键词', '结果', '模块', '来源', '动作', '操作者 ID', 'IP 地址', '时间范围（本地时间）']) {
  assert.match(filterForm, new RegExp(`aria-label="${label.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}"`), `audit filter missing aria-label ${label}`)
}
assert.match(filterView, /\.audit-filter-row--secondary\s*\{[^}]*display:\s*flex;[^}]*flex-wrap:\s*wrap;/)
assert.match(filterView, /<AdminFilterActions\s[^>]*@search="\$emit\('search'\)"[^>]*@reset="\$emit\('reset'\)"/)
assert.match(filterForm, /\$emit\('search'\)/)
assert.match(filterForm, /\$emit\('reset'\)/)
assert.doesNotMatch(filterForm, /\$emit\('(?:export|refresh)'\)/, 'list actions must be separate from search filters')
assert.match(filterView, /<\/el-form>\s*<div class="audit-log-actions">/)
assert.match(filterView, /class="audit-log-actions__left">\s*<el-button[^>]*\btext\b[^>]*>导出<\/el-button>/)
assert.match(filterView, /<el-tooltip content="刷新日志"[\s\S]*?<el-button[^>]*:icon="Refresh"[^>]*aria-label="刷新日志"[^>]*\/>/)
assert.equal((filterView.match(/<el-card\b/g) || []).length, 1, 'toolbar must use one outer card')
assert.doesNotMatch(page, /type="selection"|logClear|logDels|批量删除|清空日志/)
assert.match(page, /<\/AdminTablePanel>\s*<div class="audit-pagination">/, 'pagination must be a standalone sibling below the table panel')
assert.doesNotMatch(page + detail, /v-html/)
assert.match(detail, /变更快照/)
assert.match(detail, /请求前后快照.*并发更新/)
assert.doesNotMatch(detail, /label="变更前"|label="变更后"/)
for (const permission of ['admin:menu:log:list', 'admin:menu:log:security', 'admin:menu:log:export', 'admin:menu:log:security-export']) assert.ok(state.includes(permission))
console.log('audit log contract, identity, filtering, CSV safety, redaction and stale request checks passed')

// Run the real controller with Vue reactivity and deferred read responses.
const vue = await import('vue')
function harness(permissions) {
  const requests = []
  const hooks = {}
  const confirmations = []
  const scope = vue.effectScope()
  const queue = name => (...args) => new Promise((resolve, reject) => requests.push({ name, args, resolve, reject }))
  const module = executeTypeScriptModule(state, 'useAuditLogs.ts', {
    vue: { ...vue, onMounted: fn => { hooks.mount = fn }, onBeforeUnmount: fn => { hooks.unmount = fn }, onActivated: fn => { hooks.activate = fn }, onDeactivated: fn => { hooks.deactivate = fn } },
    'element-plus': { ElMessageBox: { confirm: (...args) => new Promise((resolve, reject) => confirmations.push({ args, resolve, reject })) } },
    '@/api': { adminApi: Object.fromEntries(['logList', 'loginLogList', 'logDetail', 'loginLogDetail', 'logExport', 'loginLogExport'].map(name => [name, queue(name)])) },
    '@/utils/permission': { hasPerm: key => permissions.includes(key) },
    './auditLogHelpers': helpers,
  })
  const controller = scope.run(() => module.useAuditLogs())
  return { controller, requests, hooks, confirmations, stop: () => { hooks.unmount(); scope.stop() } }
}
const pageResult = (id, extra = {}) => ({ code: 0, data: { list: [{ id }], total: 1, page: 1, pageSize: 20, ...extra } })
const settle = async () => { await Promise.resolve(); await vue.nextTick() }

const restricted = harness(['admin:menu:log:list'])
assert.equal(restricted.controller.canReadLogin.value, false)
assert.equal(restricted.controller.canExport.value, false)
await restricted.controller.exportData()
assert.equal(restricted.requests.length, 0)
restricted.controller.category.value = 'login'
await restricted.controller.showDetail({ id: 1 })
assert.equal(restricted.requests.length, 0, 'direct handler calls must enforce read permissions')
restricted.stop()

const securityOnly = harness(['admin:menu:log:security'])
assert.equal(securityOnly.controller.category.value, 'login')
const securityLoad = securityOnly.hooks.mount()
assert.equal(securityOnly.requests[0].name, 'loginLogList')
securityOnly.requests[0].resolve(pageResult(11))
await securityLoad
securityOnly.stop()

const full = harness(['admin:menu:log:list', 'admin:menu:log:security', 'admin:menu:log:export', 'admin:menu:log:security-export'])
const c = full.controller
const initialLoad = full.hooks.mount()
c.filters.search = 'new filters'
full.requests[0].resolve(pageResult('stale'))
await initialLoad
assert.equal(c.rows.value.length, 0, 'editing filters cancels an outstanding list response')
assert.equal(c.dirty.value, true)
const newLoad = c.search()
full.requests[1].resolve(pageResult('new'))
await newLoad
assert.equal(c.rows.value[0].id, 'new')
const oldDetail = c.showDetail({ id: 1 })
c.detailOpen.value = false
full.requests[2].resolve({ code: 0, data: { id: 1 } })
await oldDetail
assert.equal(c.detail.value, null, 'closing a drawer cancels late details')
const oldExport = c.exportData()
assert.equal(full.requests[3].name, 'logExport')
assert.equal(full.requests[3].args[0].page, 1)
assert.equal(full.requests[3].args[0].pageSize, 1000)
c.category.value = 'login'
full.requests[3].resolve(pageResult(1, { total: 1001 }))
await oldExport
assert.equal(full.confirmations.length, 0, 'stale exports must not open a confirmation or download')
assert.equal(full.requests[4].name, 'loginLogList')
full.requests[4].resolve(pageResult(11))
await settle()
assert.equal(c.rows.value[0].id, 11)
const loginExport = c.exportData()
assert.equal(full.requests[5].name, 'loginLogExport')
full.requests[5].resolve(pageResult(11, { total: 1042 }))
await settle()
assert.match(full.confirmations[0].args[0], /1042.*1000.*1041/)
full.confirmations[0].reject('cancel')
await loginExport
assert.match(c.exportNotice.value, /已取消/)
assert.equal(c.exporting.value, false)
const failing = c.load()
full.requests[6].reject(new Error('failure'))
await failing
assert.match(c.error.value, /加载失败/)
assert.equal(c.rows.value.length, 0)
assert.equal(c.loading.value, false)
const pending = c.load()
full.hooks.deactivate()
full.requests[7].resolve(pageResult('late after navigation'))
await pending
assert.equal(c.rows.value.length, 0)
full.hooks.activate()
assert.equal(full.requests[8].name, 'loginLogList')
full.requests[8].resolve(pageResult(12))
await settle()
assert.equal(c.rows.value[0].id, 12)
full.stop()

const pagination = harness(['admin:menu:log:list'])
const paginationController = pagination.controller
const paginationInitial = pagination.hooks.mount()
pagination.requests[0].resolve(pageResult('page-1', { total: 45, page: 1 }))
await paginationInitial
assert.equal(paginationController.total.value, 45)
paginationController.changePage(2)
assert.equal(paginationController.page.value, 2)
assert.equal(pagination.requests[1].args[0].page, 2)
assert.equal(paginationController.total.value, 45, 'changing pages must preserve total while the next page loads')
pagination.requests[1].resolve(pageResult('page-2', { total: 45, page: 2 }))
await settle()
assert.equal(paginationController.page.value, 2)
assert.equal(paginationController.rows.value[0].id, 'page-2')
pagination.stop()
console.log('audit log controller permission, list/detail/export cancellation, retry and cached-route lifecycle checks passed')
