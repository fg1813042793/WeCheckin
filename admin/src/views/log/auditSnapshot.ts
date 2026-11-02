import type { AuditLogRecord, AuditSnapshotStatus } from '@/types/auditLog'
import { safeChangeValue } from './auditLogHelpers'

const families: Record<string, string> = {
  user: '用户', role: '角色', permission: '权限', department: '部门', position: '岗位',
  dict_type: '字典类型', dict_item: '字典项', survey: '问卷', exam: '考试',
  workflow_definition: '流程定义', org_approver: '组织审批人',
}

// Match the collector's explicit columns, including its relation collections.
const fields: Record<string, string> = {
  id: '记录 ID', exists: '记录是否存在',
  user_account: '账号', user_name: '姓名', user_status: '用户状态', user_position_id: '岗位 ID',
  user_admin_enabled: '后台访问开关', user_admin_type: '管理员类型', user_role_id: '主角色 ID',
  role_name: '角色名称', role_sort: '角色排序', role_status: '角色状态',
  role_allow_admin_login: '允许后台登录', role_data_scope: '数据范围',
  permission_key: '权限编码', permission_name: '权限名称', permission_platform: '权限平台',
  permission_type: '权限类型', permission_parent_key: '上级权限编码', permission_sort: '权限排序', permission_status: '权限状态',
  dept_name: '部门名称', dept_parent_id: '上级部门 ID', dept_sort: '部门排序', dept_status: '部门状态',
  position_name: '岗位名称', position_sort: '岗位排序', position_status: '岗位状态',
  dict_type_code: '字典类型编码', dict_type_name: '字典类型名称', dict_type_status: '字典类型状态',
  dict_label: '字典项名称', dict_value: '字典项值', dict_calculation_value: '计算值', dict_sort: '字典项排序', dict_status: '字典项状态',
  survey_id: '问卷 ID', survey_title: '问卷标题', survey_category: '问卷分类', survey_status: '问卷状态',
  survey_visibility: '可见范围', survey_allow_multi: '允许多次提交', survey_start_time: '开始时间', survey_end_time: '结束时间',
  survey_max_response: '最大回收数量', survey_show_result: '展示结果', survey_anonymous: '匿名填写', survey_dept_ids: '限定部门 ID',
  survey_mode: '问卷模式', survey_order: '问卷排序',
  exam_id: '考试 ID', exam_title: '考试标题', exam_category: '考试分类', exam_status: '考试状态', exam_visibility: '可见范围',
  exam_allow_multi: '允许多次作答', exam_anonymous: '匿名作答', exam_show_result: '展示结果', exam_paper_id: '试卷 ID',
  exam_start_time: '开始时间', exam_end_time: '结束时间', exam_duration: '考试时长', exam_max_attempts: '最大作答次数',
  exam_show_score: '展示分数', exam_max_response: '最大回收数量', exam_dept_ids: '限定部门 ID', exam_mode: '考试模式',
  exam_publish_dept_ids: '发布部门 ID', exam_order: '考试排序',
  definition_key: '流程编码', definition_name: '流程名称', definition_display_name: '流程显示名称',
  definition_category: '流程分类', definition_status: '流程状态', definition_current_version: '当前发布版本',
  definition_start_config_revision: '发起配置修订号', definition_field_permissions_revision: '字段权限修订号', definition_deleted_at: '删除时间',
  subject_type: '对象类型', subject_id: '对象 ID', identity_code: '审批身份编码', user_id: '审批人用户 ID',
  assignment_sort: '审批人排序', assignment_status: '审批人配置状态',
  user_roles: '角色关联', user_depts: '部门关联', user_reporting_relations: '汇报关系',
  permission_grants: '权限授权', workflow_org_approver_assignments: '审批人配置',
  user_role_user_id: '用户 ID', user_role_role_id: '角色 ID', user_role_is_primary: '是否主角色', user_role_status: '角色关联状态',
  user_dept_user_id: '用户 ID', user_dept_dept_id: '部门 ID', employee_user_id: '员工用户 ID', manager_user_id: '上级用户 ID',
  relation_type: '汇报关系类型', is_primary: '是否主要关系', relation_sort: '关系排序', relation_status: '关系状态',
  effective_from: '生效时间', effective_to: '失效时间', grant_subject_id: '授权对象 ID', grant_permission_key: '授权权限编码',
  grant_effect: '授权效果', grant_status: '授权状态', scope: '授权范围', deptIds: '部门 ID 列表', userIds: '用户 ID 列表', setup_key: '配置键',
}

const settingLabels: Record<string, string> = {
  ADMIN_LOGIN_CAPTCHA_ENABLED: '管理员登录验证码', DINGTALK_H5_LOGIN_CAPTCHA_ENABLED: '钉钉 H5 登录验证码',
  USER_SINGLE_LOGIN: '用户单点登录', ADMIN_SINGLE_LOGIN: '管理员单点登录', DINGTALK_H5_SINGLE_LOGIN: '钉钉 H5 单点登录',
  DINGTALK_H5_SELF_BIND_ENABLED: '钉钉 H5 自助绑定', DINGTALK_H5_NOTIFY_ENABLED: '钉钉 H5 通知开关', DINGTALK_H5_NOTIFY_MODE: '钉钉 H5 通知方式',
  TOKEN_ADMIN_EXPIRE: '管理员登录有效期', TOKEN_USER_EXPIRE: '用户登录有效期', TOKEN_DINGTALK_H5_EXPIRE: '钉钉 H5 登录有效期',
  TOKEN_ADMIN_REDIS_PREFIX: '管理员会话存储前缀', TOKEN_USER_REDIS_PREFIX: '用户会话存储前缀', TOKEN_DINGTALK_H5_REDIS_PREFIX: '钉钉 H5 会话存储前缀',
  DINGTALK_H5_CORP_ID: '钉钉企业 ID', DINGTALK_H5_APP_KEY: '钉钉应用 Key', DINGTALK_H5_APP_SECRET: '钉钉应用密钥',
  DINGTALK_H5_AGENT_ID: '钉钉 Agent ID', DINGTALK_H5_UNIFIED_APP_ID: '钉钉统一应用 ID', DINGTALK_H5_ROBOT_CODE: '钉钉机器人编码',
  DINGTALK_H5_APP_URL: '钉钉 H5 应用地址', DINGTALK_H5_APP_NAME: '钉钉 H5 应用名称',
  DINGTALK_H5_LOGO_TEXT: '钉钉 H5 标识文字', DINGTALK_H5_LOGO_URL: '钉钉 H5 标识图片地址',
  STATIC_DOMAIN: '静态资源域名', SETUP_USER_FORM_FIELDS: '用户表单字段', SETUP_CONTENT_AGREEMENT: '用户协议',
  SETUP_CONTENT_PRIVACY: '隐私政策', SETUP_CONTENT_ABOUT: '关于我们', BUILTIN_TEMPLATE_PRESETS: '内置模板预设',
  HOME_PAGE_CONFIG: '首页配置', ADMIN_FRONTEND_CONFIG: '管理后台界面配置',
}
const corpLabels: Record<string, string> = {
  id: '配置 ID', corp_id: '企业 ID', corp_name: '企业名称', app_key: '应用 Key', app_secret: '应用密钥',
  agent_id: 'Agent ID', unified_app_id: '统一应用 ID', app_url: '应用地址', notify_enabled: '通知开关',
  notify_mode: '通知方式', robot_code: '机器人编码', enabled: '企业配置开关',
}
const own = <T>(map: Record<string, T>, key: string): T | undefined => Object.prototype.hasOwnProperty.call(map, key) ? map[key] : undefined
const corpField = /^settings\.corp\.([1-9]\d*)\.([^.]+)$/
const entityField = /^([^.]+)\.(.+)\.([^.]+)$/

export function snapshotFieldLabel(field: string): string {
  const corp = field.match(corpField)
  if (corp) return `钉钉企业配置（ID：${corp[1]}） · ${own(corpLabels, corp[2]) ?? corp[2]}`
  if (field.startsWith('settings.')) return own(settingLabels, field.slice(9)) ?? field
  const entity = field.match(entityField)
  if (entity && own(families, entity[1])) {
    const identifier = entity[1] === 'dict_type' ? '编码' : entity[1] === 'org_approver' ? '对象 ID' : 'ID'
    return `${families[entity[1]]}（${identifier}：${entity[2]}） · ${own(fields, entity[3]) ?? entity[3]}`
  }
  return own(fields, field) ?? field
}

const publicBooleans = new Set([
  'ADMIN_LOGIN_CAPTCHA_ENABLED', 'DINGTALK_H5_LOGIN_CAPTCHA_ENABLED', 'USER_SINGLE_LOGIN', 'ADMIN_SINGLE_LOGIN',
  'DINGTALK_H5_SINGLE_LOGIN', 'DINGTALK_H5_SELF_BIND_ENABLED', 'DINGTALK_H5_NOTIFY_ENABLED',
])
const publicDurations = new Set(['TOKEN_ADMIN_EXPIRE', 'TOKEN_USER_EXPIRE', 'TOKEN_DINGTALK_H5_EXPIRE'])
const notifyModes: Record<string, string> = {
  agent: '旧版工作通知（AgentId + OA）', agent_fallback: '旧版优先，失败兜底新版', robot: '新版机器人通知（sampleLink）',
}
function booleanValue(value: unknown): boolean | undefined {
  if (value === true || value === 1 || value === '1' || value === 'true') return true
  if (value === false || value === 0 || value === '0' || value === 'false') return false
  return undefined
}
function positiveDuration(value: unknown): value is string {
  if (typeof value !== 'string' || value.length > 128 || !/^(?:\d+(?:\.\d+)?(?:ns|us|µs|μs|ms|s|m|h))+$/.test(value)) return false
  const units: Record<string, number> = { ns: 1, us: 1e3, 'µs': 1e3, 'μs': 1e3, ms: 1e6, s: 1e9, m: 60e9, h: 3600e9 }
  const nanos = [...value.matchAll(/(\d+(?:\.\d+)?)(ns|us|µs|μs|ms|s|m|h)/g)].reduce((sum, match) => sum + Number(match[1]) * units[match[2]], 0)
  return nanos >= 1 && nanos <= 9223372036854775807
}
function publicSettingValue(field: string, value: unknown): string {
  const key = field.slice(9)
  const corp = field.match(corpField)
  if (publicBooleans.has(key) || (corp && ['enabled', 'notify_enabled'].includes(corp[2]))) {
    const flag = booleanValue(value)
    return flag === undefined ? '[已隐藏]' : flag ? '开启' : '关闭'
  }
  if (publicDurations.has(key)) return positiveDuration(value) ? value : '[已隐藏]'
  if (key === 'DINGTALK_H5_NOTIFY_MODE' || corp?.[2] === 'notify_mode') {
    return typeof value === 'string' ? own(notifyModes, value) ?? '[已隐藏]' : '[已隐藏]'
  }
  return '[已隐藏]'
}

// Explicit UI contracts: user/index, role/index, department/index, position/index,
// menu/index, dict/index, workflow/index, SurveyList and ExamList. Relation states
// also match model/account/user.go, model/permission and model/workflow/org_approver.go. Never use a
// boolean coercion or an "all other values" branch for an entity status.
const enabledStatus: Record<string, string> = { '0': '停用', '1': '启用' }
const publishedStatus: Record<string, string> = { '0': '已停用', '1': '已发布' }
const valueLabels: Record<string, Record<string, string>> = {
  user_status: { '0': '待审核', '1': '正常', '2': '禁用', '9': '管理员' },
  role_status: enabledStatus, dept_status: enabledStatus, position_status: enabledStatus,
  permission_status: enabledStatus, dict_type_status: enabledStatus, dict_status: enabledStatus,
  user_role_status: enabledStatus, relation_status: enabledStatus, grant_status: enabledStatus, assignment_status: enabledStatus,
  definition_status: { '0': '已关闭', '1': '草稿', '2': '已发布' },
  survey_status: publishedStatus, exam_status: publishedStatus,
  role_data_scope: { '1': '全部数据', '2': '本部门', '3': '本人', '4': '自定义' },
  survey_visibility: { '0': '公开链接', '1': '登录可见', '2': '部门限定' },
}
const entityBooleans = new Set([
  'user_admin_enabled', 'role_allow_admin_login', 'survey_allow_multi', 'survey_show_result', 'survey_anonymous',
  'exam_allow_multi', 'exam_anonymous', 'exam_show_result', 'exam_show_score',
])
function entityScalar(key: string, value: unknown): unknown {
  const mapping = own(valueLabels, key)
  if (mapping && (typeof value === 'string' || typeof value === 'number')) return own(mapping, String(value)) ?? value
  const flag = booleanValue(value)
  if (flag !== undefined) {
    if (key === 'exists') return flag ? '存在' : '不存在'
    if (key === 'user_role_is_primary' || key === 'is_primary') return flag ? '是' : '否'
    if (entityBooleans.has(key)) return flag ? '开启' : '关闭'
  }
  return value
}
function localizeObject(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(localizeObject)
  if (value && typeof value === 'object') {
    return Object.fromEntries(Object.entries(value).map(([key, item]) => [own(fields, key) ?? key, localizeObject(entityScalar(key, item))]))
  }
  return value
}

export function snapshotValue(field: string, value: unknown): string {
  if (value === '[已隐藏]' || value === '[已变更]') return value
  if (value === undefined) return '未提供'
  // Historical nulls describe missing observations, not an unset business value.
  if (value === null) return '未记录'
  // No public-looking setting can bypass redaction with an arbitrary string or object.
  if (field === 'settings' || field.startsWith('settings.')) return publicSettingValue(field, value)
  const safe = safeChangeValue(field, value)
  if (safe === '[已隐藏]' || safe === '[已变更]') return safe
  const entity = field.match(entityField)
  const key = entity && own(families, entity[1]) ? entity[3] : field
  if (typeof value !== 'object' && !/^\s*[\[{]/.test(safe)) return safeChangeValue(field, entityScalar(key, value))
  try { return safeChangeValue(field, localizeObject(JSON.parse(safe) as unknown)) } catch { return safe }
}

const states: Record<AuditSnapshotStatus, { title: string; description: string }> = {
  pending: { title: '等待快照完成', description: '请求已开始，快照采集尚未完成或尚未完成最终记录。' },
  recorded: { title: '已记录变更快照', description: '已记录实际字段差异。' },
  unchanged: { title: '已采集字段无差异', description: '已采集字段在请求前后未发现差异，不代表所有业务字段均未变化。' },
  not_applicable: { title: '不适用字段快照', description: '此类操作按设计不采集请求前后的字段快照。' },
  not_configured: { title: '未接入快照采集', description: '此操作或配置键尚未接入快照采集。' },
  operation_failed: { title: '业务操作未成功', description: '业务操作失败或被拒绝，未附加字段快照值。' },
  unavailable: { title: '快照采集失败', description: '本次未能完成快照采集；业务结果请查看上方记录。' },
  legacy_unknown: { title: '历史快照状态未知', description: '历史记录未提供快照采集状态，无法确认未记录差异的原因。' },
}
const reasons: Record<string, string> = {
  no_field_snapshot: '此类操作不采集字段快照', create_target_only: '创建操作仅记录新建目标', runtime_history: '运行操作通过关联运行记录追踪',
  unsupported_route: '此操作尚未接入快照采集', unsupported_setup_key: '此配置键尚未接入快照采集',
  before_read_failed: '请求前快照读取失败', after_read_failed: '请求后快照读取失败',
  before_limit: '请求前快照超过采集限制', after_limit: '请求后快照超过采集限制', diff_limit: '字段差异超过记录限制',
  encoding_limit: '快照内容处理失败或超过记录限制', target_limit: '目标数量超过采集限制', request_unavailable: '无法读取用于采集的请求信息',
  invalid_target: '无法确定有效的采集目标', result_unconfirmed: '业务结果尚未确认', business_not_successful: '业务操作未成功',
  no_collected_changes: '已采集字段未发现差异',
}

export function auditSnapshotView(record: AuditLogRecord | null) {
  const status = record?.snapshotStatus || 'legacy_unknown'
  const state = own(states, status) ?? { title: '未知快照状态', description: '无法识别此记录的快照采集状态，未推断采集结果。' }
  const changes = record?.changes ?? []
  const rows = changes.filter(change => change.field !== 'privateConfigChanged').map(change => ({
    field: change.field, label: snapshotFieldLabel(change.field),
    before: snapshotValue(change.field, change.before), after: snapshotValue(change.field, change.after),
  }))
  let reason = record?.snapshotReason ? own(reasons, record.snapshotReason) ?? '未识别的快照原因，未作推断。' : ''
  if (!reason && status === 'legacy_unknown' && record?.reasonCode?.startsWith('audit_snapshot_')) {
    const historicalReason = own(reasons, record.reasonCode.slice('audit_snapshot_'.length))
    if (historicalReason) reason = `历史原因记录：${historicalReason}。`
  }
  const hasLegacyMarker = changes.some(change => change.field === 'privateConfigChanged')
  return {
    ...state, rows, reason,
    hasRedactions: hasLegacyMarker || rows.some(row => /\[已隐藏\]|\[已变更\]/.test(row.before + row.after)),
    description: status === 'legacy_unknown' && rows.length ? '以下展示历史记录实际保存的字段差异；缺少采集状态，不能据此推断完整采集过程。' : state.description,
    legacyNotice: hasLegacyMarker
      ? '历史标记仅说明配置发生过变更，不代表配置已开启；未记录的旧值无法还原。' : '',
  }
}
