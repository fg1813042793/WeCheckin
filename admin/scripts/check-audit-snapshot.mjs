import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import * as vue from 'vue'
import { parse, compileScript } from '@vue/compiler-sfc'
import { renderToString } from '@vue/server-renderer'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')
const helpers = executeTypeScriptModule(read('src/views/log/auditLogHelpers.ts'), 'auditLogHelpers.ts')
const snapshotPath = 'src/views/log/auditSnapshot.ts'
const snapshot = existsSync(new URL(`../${snapshotPath}`, import.meta.url))
  ? executeTypeScriptModule(read(snapshotPath), 'auditSnapshot.ts', { './auditLogHelpers': helpers }) : {}
const source = read('src/views/log/AuditLogDetail.vue')
const { descriptor } = parse(source, { filename: 'AuditLogDetail.vue' })
const compiled = compileScript(descriptor, { id: 'audit-snapshot-test', inlineTemplate: true })
const passThrough = { setup: (_, { slots }) => () => vue.h('div', slots.default?.()) }
const detail = executeTypeScriptModule(compiled.content, 'AuditLogDetail.vue.ts', {
  vue, '@/components/admin-ui': { AdminDrawer: passThrough },
  '@/utils/permission': { hasPerm: () => false },
  './auditLogHelpers': helpers, './auditSnapshot': snapshot,
}).default

async function render(record, overrides = {}) {
  const app = vue.createSSRApp({ render: () => vue.h(detail, { modelValue: true, loading: false, error: '', record: { id: 1, ...record }, ...overrides }) })
  for (const name of ['ElDescriptions', 'ElTag', 'ElButton', 'RouterLink']) app.component(name, passThrough)
  app.component('ElDescriptionsItem', { props: ['label'], setup: (props, { slots }) => () => vue.h('div', [props.label, slots.default?.()]) })
  app.component('ElAlert', { props: ['title'], setup: props => () => vue.h('p', { role: 'status' }, props.title) })
  app.component('ElEmpty', { props: ['description'], setup: props => () => vue.h('p', props.description) })
  app.component('ElTooltip', { props: ['content'], setup: (props, { slots }) => () => vue.h('span', { title: props.content }, slots.default?.()) })
  app.component('ElTable', { props: ['data'], setup: (props, { slots }) => {
    vue.provide('audit-test-rows', props.data)
    return () => vue.h('div', { 'data-test': 'snapshot-table' }, slots.default?.())
  } })
  app.component('ElTableColumn', { props: ['label', 'prop'], setup: (props, { slots }) => {
    const rows = vue.inject('audit-test-rows')
    return () => vue.h('section', [vue.h('h4', props.label), ...rows.map(row => vue.h('div', slots.default?.({ row }) ?? String(row[props.prop] ?? '')))])
  } })
  app.directive('loading', { getSSRProps: () => ({}) })
  return renderToString(app)
}

// Compile and render the actual SFC. Only shell/Element Plus components and permissions are mocked.
const { snapshotFieldLabel: label, snapshotValue: value, auditSnapshotView: view } = snapshot
for (const field of ['settings.STATIC_DOMAIN', 'settings.ADMIN_LOGIN_CAPTCHA_ENABLED', 'settings.TOKEN_ADMIN_EXPIRE', 'user.7.user_name', 'password']) {
  assert.equal(value(field, null), '未记录', 'null is a missing observation, never proof of an unset value')
}
for (const [field, after, expectedAfter] of [
  ['settings.STATIC_DOMAIN', '[已变更]', '[已变更]'],
  ['settings.ADMIN_LOGIN_CAPTCHA_ENABLED', true, '开启'],
]) {
  const changes = [{ field, before: null, after }]
  const row = view({ changes }).rows[0]
  assert.equal(row.before, '未记录')
  assert.equal(row.after, expectedAfter)
  const html = await render({ changes })
  assert.match(html, /<pre>未记录<\/pre>/)
  assert.ok(html.includes(`<pre>${expectedAfter}</pre>`))
  assert.doesNotMatch(html, /未设置/)
  assert.match(html, /请求前后快照.*并发更新/)
}
const safeCaptchaBoolean = { field: 'settings.ADMIN_LOGIN_CAPTCHA_ENABLED', before: false, after: true }
const captchaHTML = await render({ snapshotStatus: 'recorded', changes: [safeCaptchaBoolean] })
assert.match(captchaHTML, /管理员登录验证码/)
assert.match(captchaHTML, /关闭/)
assert.match(captchaHTML, /开启/)
assert.match(captchaHTML, /请求前/)
assert.match(captchaHTML, /请求后/)
assert.match(captchaHTML, /请求前后快照.*并发更新/)
assert.doesNotMatch(captchaHTML, /脱敏|保护标记/)
assert.doesNotMatch(source, /v-html/)

const booleanFields = ['ADMIN_LOGIN_CAPTCHA_ENABLED', 'DINGTALK_H5_LOGIN_CAPTCHA_ENABLED', 'USER_SINGLE_LOGIN', 'ADMIN_SINGLE_LOGIN', 'DINGTALK_H5_SINGLE_LOGIN', 'DINGTALK_H5_SELF_BIND_ENABLED', 'DINGTALK_H5_NOTIFY_ENABLED', 'TOKEN_ADMIN_SLIDING_EXPIRE_ENABLED', 'TOKEN_USER_SLIDING_EXPIRE_ENABLED', 'TOKEN_DINGTALK_H5_SLIDING_EXPIRE_ENABLED'].map(key => `settings.${key}`)
booleanFields.push('settings.corp.12.enabled', 'settings.corp.12.notify_enabled')
for (const field of booleanFields) {
  for (const input of [true, 1, '1', 'true']) assert.equal(value(field, input), '开启', field)
  for (const input of [false, 0, '0', 'false']) assert.equal(value(field, input), '关闭', field)
  for (const input of ['rawsecretblocked', 'TRUE', 'yes', 2, {}, [], { value: true }]) assert.equal(value(field, input), '[已隐藏]', field)
  assert.equal(value(field, null), '未记录')
}
for (const field of ['settings.TOKEN_ADMIN_EXPIRE', 'settings.TOKEN_USER_EXPIRE', 'settings.TOKEN_DINGTALK_H5_EXPIRE']) {
  for (const duration of ['24h0m0s', '1h30m', '1.5s', '500µs']) assert.equal(value(field, duration), duration)
  for (const input of ['rawsecretblocked', '0s', '-1h', '1h secret', 'Infinityh', '999999999999h', 3600, {}, []]) assert.equal(value(field, input), '[已隐藏]')
}
for (const field of ['settings.DINGTALK_H5_NOTIFY_MODE', 'settings.corp.12.notify_mode']) {
  assert.match(value(field, 'agent'), /旧版工作通知/)
  assert.match(value(field, 'agent_fallback'), /旧版优先/)
  assert.match(value(field, 'robot'), /新版机器人通知/)
  assert.equal(value(field, 'rawsecretblocked'), '[已隐藏]')
}
for (const field of ['settings.ADMIN_LOGIN_CAPTCHA_ENABLED_EXTRA', 'settings.ADMIN_LOGIN_CAPTCHA_ENABLED.token', 'settings.TOKEN_ADMIN_EXPIRE.extra', 'settings.corp.0.enabled', 'settings.corp.-1.enabled', 'settings.corp.1.app_secret', 'settings.corp.1.app_key', 'settings.STATIC_DOMAIN', 'password', 'credentials.token', 'apiKey', 'user.3.user_password']) {
  assert.equal(value(field, 'rawsecretblocked'), '[已隐藏]', field)
}
for (const field of ['settings.ADMIN_LOGIN_CAPTCHA_ENABLED', 'settings.TOKEN_ADMIN_EXPIRE', 'password', 'user.1.user_name']) {
  for (const marker of ['[已隐藏]', '[已变更]']) assert.equal(value(field, marker), marker)
  assert.equal(value(field, undefined), '未提供', 'do not invent missing old values')
}
const mixed = await render({ snapshotStatus: 'recorded', changes: [safeCaptchaBoolean,
  { field: 'settings.TOKEN_ADMIN_EXPIRE', before: '1h0m0s', after: '2h0m0s' },
  { field: 'settings.corp.3.app_secret', before: 'rawsecretblocked', after: 'another-raw-secret' },
  { field: 'settings.ADMIN_LOGIN_CAPTCHA_ENABLED', before: 'rawsecretblocked', after: { secret: 'another-raw-secret' } },
  { field: 'user.7.user_name', before: '旧名字', after: '<img src=x onerror=alert(1)>' },
] })
assert.match(mixed, /管理员登录有效期/)
assert.match(mixed, /1h0m0s/)
assert.match(mixed, /旧名字/)
assert.match(mixed, /&lt;img/)
assert.doesNotMatch(mixed, /rawsecretblocked|another-raw-secret|<img src=x/)
assert.match(mixed, /敏感值已脱敏/)
assert.equal(value('user.7.user_status', 0), '待审核')
assert.equal(value('user.7.user_status', '1'), '正常')
assert.equal(value('user.7.user_status', 2), '禁用')
assert.equal(value('user.7.user_status', 9), '管理员')
for (const family of ['role', 'department', 'position', 'permission', 'dict_type', 'dict_item']) {
  const key = { department: 'dept', dict_type: 'dict_type', dict_item: 'dict' }[family] || family
  assert.equal(value(`${family}.7.${key}_status`, '0'), '停用')
  assert.equal(value(`${family}.7.${key}_status`, '1'), '启用')
  assert.equal(value(`${family}.7.${key}_status`, '8'), '8')
}
for (const [status, expected] of [[0, '已关闭'], [1, '草稿'], [2, '已发布']]) assert.equal(value('workflow_definition.7.definition_status', status), expected)
for (const family of ['survey', 'exam']) {
  assert.equal(value(`${family}.7.${family}_status`, 0), '已停用')
  assert.equal(value(`${family}.7.${family}_status`, 1), '已发布')
  assert.equal(value(`${family}.7.${family}_status`, 2), '2')
}
assert.equal(value('user.7.user_status', 'future-state'), 'future-state')
assert.equal(value('user.7.user_status', false), 'false')
assert.match(label('user.7.user_name'), /用户.*7.*姓名/)
assert.match(label('dict_type.team.kind.dict_type_name'), /字典类型.*team.kind.*名称/)
assert.match(label('settings.corp.12.notify_enabled'), /企业.*12.*通知/)
assert.equal(label('future.1.unknown_field'), 'future.1.unknown_field')
assert.equal(label('user.1.unknown_field'), '用户（ID：1） · unknown_field')
const relations = value('user.7.user_roles', [{ user_role_role_id: '2', user_role_status: '1', user_role_is_primary: '1', password: 'rawsecretblocked' }])
assert.match(relations, /角色 ID/)
assert.match(relations, /启用/)
assert.doesNotMatch(relations, /rawsecretblocked/)
for (const [collection, key] of [['user_reporting_relations', 'relation_status'], ['permission_grants', 'grant_status'], ['workflow_org_approver_assignments', 'assignment_status']]) {
  const field = `user.7.${collection}`
  assert.match(value(field, [{ [key]: '1' }]), /启用/)
  assert.match(value(field, [{ [key]: '0' }]), /停用/)
  assert.match(value(field, [{ [key]: 'future-state' }]), /future-state/)
}

const statuses = {
  pending: /采集尚未完成/,
  recorded: /已记录/,
  unchanged: /已采集字段.*未发现差异/,
  not_applicable: /不适用/,
  not_configured: /未接入快照采集/,
  operation_failed: /业务操作失败或被拒绝/,
  unavailable: /快照采集失败/,
  legacy_unknown: /历史记录.*无法确认/,
}
for (const [snapshotStatus, expected] of Object.entries(statuses)) {
  const html = await render({ snapshotStatus, changes: [], outcome: 'success' })
  assert.match(html, expected, snapshotStatus)
  assert.doesNotMatch(html, /data-test="snapshot-table"/)
  assert.doesNotMatch(html, /策略|序列化|保护标记/)
}
for (const outcome of ['success', 'failure', 'blocked', 'accepted', undefined]) {
  const state = view({ outcome, changes: [] })
  assert.match(state.description, /历史记录.*无法确认/)
  assert.doesNotMatch(state.description, /采集失败|操作失败|未配置|未发现差异/)
}
const historical = await render({ changes: [{ field: 'user.7.user_name', before: '历史姓名甲', after: '历史姓名乙' }] })
assert.match(historical, /历史姓名甲/)
assert.match(historical, /历史姓名乙/)
assert.match(historical, /data-test="snapshot-table"/)
assert.doesNotMatch(historical, /脱敏|保护标记/)
const marker = { field: 'privateConfigChanged', after: true }
for (const changes of [[marker], [marker, safeCaptchaBoolean], [marker, { field: 'settings.ADMIN_LOGIN_CAPTCHA_ENABLED', after: '[已变更]' }]]) {
  const html = await render({ changes })
  assert.match(html, /配置发生过变更/)
  assert.match(html, /不代表.*开启/)
  assert.match(html, /敏感值已脱敏/)
  assert.doesNotMatch(html, /privateConfigChanged|>true</)
  assert.equal(view({ changes }).rows.length, changes.length - 1)
}
assert.equal(view({ changes: [marker] }).rows.length, 0)
const oldReason = view({ reasonCode: 'audit_snapshot_before_read_failed' })
assert.match(oldReason.reason, /历史原因记录.*请求前.*读取失败/)
assert.match(oldReason.description, /无法确认/)
assert.equal(view({ reasonCode: 'business_not_successful' }).reason, '')
assert.equal(view({ reasonCode: 'audit_snapshot_rawsecretblocked' }).reason, '')
for (const reason of ['no_field_snapshot', 'create_target_only', 'runtime_history', 'unsupported_route', 'unsupported_setup_key', 'before_read_failed', 'after_read_failed', 'before_limit', 'after_limit', 'diff_limit', 'encoding_limit', 'target_limit', 'request_unavailable', 'invalid_target', 'result_unconfirmed', 'business_not_successful', 'no_collected_changes']) {
  const state = view({ snapshotStatus: 'unavailable', snapshotReason: reason })
  assert.ok(state.reason && /[\u4e00-\u9fff]/u.test(state.reason), reason)
  assert.doesNotMatch(state.reason, /未识别/, reason)
  assert.doesNotMatch(state.reason, /策略|序列化/, reason)
}
assert.match(view({ snapshotStatus: 'unavailable', snapshotReason: 'before_read_failed' }).reason, /请求前.*读取失败/)
assert.match(view({ snapshotStatus: 'unavailable', snapshotReason: 'after_limit' }).reason, /请求后.*超过采集限制/)
for (const change of [
  { field: 'settings.STATIC_DOMAIN', before: null, after: '[已变更]' },
  { field: 'password', before: '[已隐藏]', after: '[已变更]' },
  { field: 'user.7.user_roles', after: [{ password: 'rawsecretblocked' }] },
]) {
  const html = await render({ changes: [change] })
  assert.match(html, /敏感值已脱敏/)
  assert.doesNotMatch(html, /rawsecretblocked/)
}
const unavailable = await render({ outcome: 'success', snapshotStatus: 'unavailable', snapshotReason: 'after_read_failed' })
assert.match(unavailable, /请求后.*读取失败/)
assert.match(unavailable, /成功/)
const future = await render({ snapshotStatus: 'future_state', snapshotReason: 'rawsecretblocked' })
assert.match(future, /未知快照状态/)
assert.doesNotMatch(future, /rawsecretblocked/)
assert.match(view({ snapshotStatus: 'unavailable', snapshotReason: 'future_reason' }).reason, /未识别/)
assert.match(await render({}, { loading: true }), /正在加载日志详情/)
assert.match(await render({}, { error: '加载失败' }), /重试/)
console.log('audit snapshot real helper and rendered detail checks passed: public scalars, redaction, labels, historical records and all capture states')
