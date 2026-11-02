import type { AuditLogFilters, AuditLogOutcome, AuditLogPage, AuditLogQuery, AuditLogRecord } from '@/types/auditLog'

export const EXPORT_LIMIT = 1000
export const outcomeOptions: { value: AuditLogOutcome; label: string; type: 'success' | 'danger' | 'warning' | 'primary' | 'info' }[] = [
  { value: 'success', label: '成功', type: 'success' },
  { value: 'failure', label: '失败', type: 'danger' },
  { value: 'blocked', label: '已阻止', type: 'danger' },
  { value: 'partial', label: '部分完成', type: 'warning' },
  { value: 'accepted', label: '已受理（未完成）', type: 'primary' },
  { value: 'unknown', label: '未知', type: 'info' },
]
export const sourceOptions = [
  { value: 'admin', label: '管理后台' }, { value: 'client', label: '客户端' },
  { value: 'h5app', label: '钉钉 H5' }, { value: 'system', label: '系统' },
]
export const moduleOptions = [
  { value: 'auth', label: '身份认证' }, { value: 'user', label: '用户' },
  { value: 'role', label: '角色' }, { value: 'permission', label: '权限' },
  { value: 'organization', label: '组织架构' }, { value: 'dictionary', label: '数据字典' },
  { value: 'survey', label: '问卷' }, { value: 'exam', label: '考试' },
  { value: 'workflow', label: '流程' }, { value: 'scheduler', label: '任务调度' },
  { value: 'notification', label: '通知' }, { value: 'settings', label: '系统设置' },
  { value: 'feedback', label: '用户反馈' }, { value: 'content', label: '内容' },
  { value: 'audit', label: '审计' }, { value: 'file', label: '文件' },
  { value: 'performance', label: '绩效' }, { value: 'other', label: '其他' },
]
export const actionOptions = [
  { value: 'password_login', label: '密码登录' }, { value: 'sso_login', label: '单点登录' },
  { value: 'bind', label: '绑定' }, { value: 'logout', label: '退出登录' },
  { value: 'create', label: '新增' }, { value: 'update', label: '修改' },
  { value: 'delete', label: '删除' }, { value: 'status', label: '状态变更' },
  { value: 'publish', label: '发布' }, { value: 'rollback', label: '回滚' },
  { value: 'run', label: '执行' }, { value: 'retry', label: '重试' },
  { value: 'cancel', label: '取消' }, { value: 'export', label: '导出' },
  { value: 'password_change', label: '修改密码' }, { value: 'force_offline', label: '强制下线' },
  { value: 'send', label: '发送' },
]

export function moduleLabel(value?: string) {
  return moduleOptions.find(item => item.value === value)?.label || value || '未记录'
}

export function affectedCountLabel(value?: number) {
  return value != null && Number.isSafeInteger(value) && value > 0 ? String(value) : '-'
}

export function requestedTargetsSummary(record?: Pick<AuditLogRecord, 'targetIds' | 'requestedCount' | 'action' | 'affectedCount'> | null) {
  const provided = Array.isArray(record?.targetIds)
  const ids = provided ? record!.targetIds! : []
  const count = record?.requestedCount
  const knownCount = count != null && Number.isSafeInteger(count) && count >= 0
  const joined = ids.slice(0, 1000).join(', ')
  const displayTruncated = ids.length > 1000 || joined.length > 16000
  const notices: string[] = []
  if (knownCount && count > ids.length) {
    notices.push(provided
      ? `请求 ${count} 个目标，仅返回 ${ids.length} 个目标 ID，列表已截断。`
      : `请求 ${count} 个目标，但目标 ID 列表未提供，无法展示完整列表。`)
  }
  if (record?.action === 'export' && provided && record.affectedCount != null && record.affectedCount > ids.length) {
    notices.push(`导出 ${record.affectedCount} 条记录，仅记录 ${ids.length} 个目标 ID，目标列表不完整。`)
  }
  if (displayTruncated) notices.push('目标 ID 展示已截断：最多显示 1000 个 ID、16000 个字符。')
  return {
    text: provided ? (joined.slice(0, 16000) || '-') + (joined.length > 16000 ? '…' : '') : '未提供',
    countLabel: knownCount ? String(count) : '-',
    notice: notices.join(' '),
  }
}

export function outcomeMeta(value?: string) {
  return outcomeOptions.find(item => item.value === value) || outcomeOptions[5]!
}

export function sourceLabel(value?: string) {
  return sourceOptions.find(item => item.value === value)?.label || value || '未知'
}

export function actorLabel(row: AuditLogRecord) {
  if (row.actorType === 'system') return '系统'
  if (row.actorType === 'anonymous' || !row.adminId || row.adminId <= 0) return '匿名 / 未确认身份'
  const kind = row.actorType === 'user' ? '用户' : '管理员'
  return `${kind}：${row.adminName || row.adminDesc || '未记录名称'}（ID ${row.adminId}）`
}

export function actionLabel(row: AuditLogRecord) {
  const legacy: Record<string, string> = { '1': '登录', '2': '添加', '3': '删除', '4': '修改', '5': '其他' }
  return actionOptions.find(item => item.value === row.action)?.label || row.action || legacy[String(row.type)] || '未记录'
}

export function logTime(value: AuditLogRecord['_createTime']) {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : date.toLocaleString('zh-CN', { hour12: false })
}

export function emptyAuditFilters(): AuditLogFilters {
  return { search: '', source: '', module: '', action: '', outcome: '', actorId: '', ip: '', range: null }
}

export function buildAuditQuery(filters: AuditLogFilters, page: number, pageSize: number): AuditLogQuery {
  const actorId = filters.actorId.trim()
  if (actorId && !/^(0|[1-9]\d*)$/.test(actorId)) throw new Error('操作者 ID 必须为非负整数；0 表示匿名身份')
  if (filters.range && (filters.range.length !== 2 || filters.range.some(value => !value || !Number.isFinite(Date.parse(value))) || Date.parse(filters.range[0]) > Date.parse(filters.range[1]))) {
    throw new Error('请选择完整且起始时间不晚于结束时间的时间范围')
  }
  return {
    search: filters.search.trim() || undefined, source: filters.source || undefined,
    module: filters.module.trim() || undefined, action: filters.action.trim() || undefined,
    outcome: filters.outcome || undefined, actorId: actorId || undefined, ip: filters.ip.trim() || undefined,
    startTime: filters.range?.[0], endTime: filters.range?.[1], page, pageSize,
  }
}

// Logical cancellation also covers transports that finish after cancellation. It
// avoids the shared Axios interceptor treating an intentional abort as a network error.
export function createRequestGate() {
  let generation = 0
  return { next: () => ++generation, invalidate: () => { generation++ }, isCurrent: (ticket: number) => ticket === generation }
}

const sensitiveField = /password|passwd|pwd|token|secret|cookie|authorization|credential|(?:private|api|access|signing).?key|captcha|密码|密钥|验证码/i
const MAX_VALUE_LENGTH = 6000

function redactedValue(value: unknown) {
  return value === '[已变更]' ? value : '[已隐藏]'
}

function redact(value: unknown, depth = 0): unknown {
  if (depth > 5) return '[内容过深，已省略]'
  if (typeof value === 'string') {
    if (/^\s*[\[{]/.test(value)) {
      try { return redact(JSON.parse(value) as unknown, depth + 1) } catch { /* Not JSON; keep inert text. */ }
    }
    return value.length > MAX_VALUE_LENGTH ? `${value.slice(0, MAX_VALUE_LENGTH)}…（已截断）` : value
  }
  if (Array.isArray(value)) {
    const items = value.slice(0, 50).map(item => redact(item, depth + 1))
    return value.length > 50 ? [...items, '[其余内容已省略]'] : items
  }
  if (value && typeof value === 'object') {
    const entries = Object.entries(value)
    const items = entries.slice(0, 50).map(([key, item]) => [key, sensitiveField.test(key) ? redactedValue(item) : redact(item, depth + 1)])
    if (entries.length > 50) items.push(['…', '[其余字段已省略]'])
    return Object.fromEntries(items)
  }
  return value
}

export function safeChangeValue(field: string, value: unknown): string {
  if (sensitiveField.test(field)) return redactedValue(value)
  if (value === undefined) return '未提供'
  if (typeof value === 'string' && /^[\s]*[\[{]/.test(value)) {
    try { value = JSON.parse(value) as unknown } catch { /* Plain text is rendered as text only. */ }
  }
  const clean = redact(value)
  const text = typeof clean === 'string' ? clean : JSON.stringify(clean, null, 2) ?? '未提供'
  return text.length > MAX_VALUE_LENGTH ? `${text.slice(0, MAX_VALUE_LENGTH)}…（已截断）` : text
}

export function exportSummary(data: AuditLogPage) {
  if (!Array.isArray(data?.list) || !Number.isSafeInteger(data.total) || data.total < data.list.length) throw new Error('导出响应的总数无效，请重试')
  if (data.list.length > EXPORT_LIMIT) throw new Error('导出响应超过 1000 条上限，请缩小筛选范围')
  return { count: data.list.length, total: data.total, truncated: data.total > data.list.length }
}

function csvCell(value: unknown) {
  let text = value == null ? '' : String(value)
  // Quotes alone do not stop spreadsheets evaluating a formula, including one
  // hidden behind whitespace/control characters or full-width formula prefixes.
  if (/^[\s\u0000-\u001f\u007f-\u009f]*[=+\-@＝＋－＠]/u.test(text) || /^[\t\r\n]/.test(text)) text = `'${text}`
  return `"${text.replace(/"/g, '""')}"`
}

export function auditCsv(rows: AuditLogRecord[]) {
  const columns: [string, (row: AuditLogRecord) => unknown][] = [
    ['日志 ID', row => row.id], ['事件 ID', row => row.eventId], ['类别', row => row.category],
    ['时间', row => logTime(row._createTime)], ['来源', row => sourceLabel(row.source)],
    ['模块', row => moduleLabel(row.module)], ['动作', actionLabel], ['结果', row => outcomeMeta(row.outcome).label],
    ['内容', row => row.content], ['原因代码', row => row.reasonCode], ['原因', row => row.reason],
    ['实际操作者', actorLabel], ['操作者类型', row => row.actorType],
    ['尝试账号（不代表已验证身份）', row => row.attemptedAccount],
    ['目标类型', row => row.targetType], ['目标 ID', row => row.targetId], ['目标名称', row => row.targetName],
    ['影响数量', row => affectedCountLabel(row.affectedCount)], ['IP', row => row.LOG_ADD_IP], ['请求 ID', row => row.requestId],
    ['浏览器', row => row.browser], ['设备类型', row => row.deviceType], ['平台', row => row.platformType],
    ['方法', row => row.method], ['路径', row => row.path], ['耗时 ms', row => row.durationMs], ['User Agent', row => row.userAgent],
    ['变更快照（请求前后，可能包含并发更新）', row => (row.changes || []).map(change => `${change.field}: ${safeChangeValue(change.field, change.before)} → ${safeChangeValue(change.field, change.after)}`).join('\n')],
    ['关联类型', row => row.relatedType], ['关联 ID', row => row.relatedId],
  ]
  return '\uFEFF' + [columns.map(([label]) => label), ...rows.map(row => columns.map(([, value]) => value(row)))].map(row => row.map(csvCell).join(',')).join('\r\n')
}

export function relatedLogRoute(type: string | undefined, id: string | number | undefined, permitted: (key: string) => boolean) {
  if (!id) return null
  // These existing routes are list pages; never manufacture unsupported detail URLs.
  const routes: Record<string, { path: string; menu: string; permission: string; label: string }> = {
    workflow_instance: { path: '/workflow/instances', menu: 'admin:menu:workflow:instances', permission: 'admin:menu:workflow:instance:list', label: '查看流程实例列表' },
    scheduled_task_run: { path: '/scheduled-task/runs', menu: 'admin:menu:scheduled-task:runs', permission: 'admin:menu:scheduled-task:run:list', label: '查看运行记录列表' },
  }
  const route = Object.prototype.hasOwnProperty.call(routes, type || '') ? routes[type!] : undefined
  return route && permitted(route.menu) && permitted(route.permission) ? route : null
}
