import fs from 'node:fs'
import path from 'node:path'
import { createRequire } from 'node:module'
import ts from 'typescript'

const root = process.cwd()
const require = createRequire(import.meta.url)
const moduleCache = new Map()
const runtimeSource = fs.readFileSync(path.join(root, 'src/views/workflow/components/WorkflowRuntimeForm.vue'), 'utf8')

function loadTypeScriptModule(relativePath) {
  const filename = path.join(root, relativePath)
  if (moduleCache.has(filename)) return moduleCache.get(filename)
  const source = fs.readFileSync(filename, 'utf8')
  const output = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2020,
      esModuleInterop: true,
    },
    fileName: filename,
  }).outputText
  const module = { exports: {} }
  moduleCache.set(filename, module.exports)
  const localRequire = specifier => {
    if (!specifier.startsWith('.')) return require(specifier)
    const resolved = path.resolve(path.dirname(filename), specifier)
    const dependency = path.extname(resolved) ? resolved : `${resolved}.ts`
    return loadTypeScriptModule(path.relative(root, dependency))
  }
  const runner = new Function('module', 'exports', 'require', output)
  runner(module, module.exports, localRequire)
  moduleCache.set(filename, module.exports)
  return module.exports
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

const {
  createWorkflowDetailRow,
  calculateWorkflowFormData,
  evaluateWorkflowCalculation,
  flattenWorkflowOptions,
  initialWorkflowFormData,
  normalizeWorkflowOptions,
  normalizeWorkflowAttachments,
  validateWorkflowFormData,
  visibleWorkflowFormFields,
  workflowFieldActionMap,
  workflowFieldAccessMap,
  writableWorkflowFormData,
  workflowTextareaAutosize,
} = loadTypeScriptModule('src/views/workflow/runtimeForm.ts')

assert(
  JSON.stringify(workflowTextareaAutosize({ minVisibleRows: 4, maxVisibleRows: 10 })) === JSON.stringify({ minRows: 4, maxRows: 10 }),
  '运行时应读取多行文本显示行数配置',
)
assert(
  JSON.stringify(workflowTextareaAutosize({}, 2, 6)) === JSON.stringify({ minRows: 2, maxRows: 6 }),
  '旧明细多行文本应使用兼容默认显示行数',
)

const fields = [
  { key: 'reason', label: '申请原因', type: 'textarea', required: true, default: '出差' },
  { key: 'amount', label: '报销金额', type: 'amount', default: 100 },
  {
    key: 'region',
    label: '所属区域',
    type: 'select',
    options: [
      { label: '华东', value: 'east', children: [{ label: '上海', value: 'shanghai' }] },
    ],
  },
  {
    key: 'department',
    label: '所属部门',
    type: 'multi_select',
    optionSource: {
      type: 'api',
      url: '/api/v2/admin/departments/tree',
      method: 'GET',
      responsePath: 'data',
      labelField: 'name',
      valueField: 'id',
      childrenField: 'children',
    },
  },
  { key: 'attachments', label: '附件', type: 'attachment' },
  { key: 'internalNote', label: '内部备注', type: 'text' },
  {
    key: 'objectives',
    label: '我的目标',
    type: 'detail_list',
    rowKey: 'id',
    columns: [
      { key: 'target', label: '目标', type: 'textarea', required: true },
      { key: 'weight', label: '权重', type: 'number' },
      { key: 'result', label: '结果', type: 'textarea' },
    ],
  },
]

const calculationFields = [
  { key: 'quantity', label: '数量', type: 'number' },
  { key: 'price', label: '单价', type: 'amount' },
  {
    key: 'items', label: '明细', type: 'detail_list',
    columns: [
      { key: 'quantity', label: '数量', type: 'number' },
      { key: 'price', label: '单价', type: 'amount' },
    ],
  },
  { key: 'scalarTotal', label: '普通合计', type: 'calculation', calculation: { expression: '[quantity] * [price]', display: 'field', precision: 2 } },
  { key: 'detailTotal', label: '明细合计', type: 'calculation', calculation: { expression: 'SUM([items.quantity] * [items.price])', display: 'label', precision: 2 } },
  { key: 'negative', label: '负数舍入', type: 'calculation', calculation: { expression: '-1.005', display: 'field', precision: 2 } },
]
const calculated = calculateWorkflowFormData(calculationFields, {
  quantity: 3,
  price: 12.345,
  items: [
    { id: 'row-1', quantity: 2, price: 10 },
    { id: 'row-2', quantity: 3, price: 4.5 },
  ],
  scalarTotal: 999,
})
assert(calculated.scalarTotal === 37.04, '普通计算字段应按配置精度实时计算')
assert(calculated.detailTotal === 33.5, '明细字段应支持不同列逐行组合后合计')
assert(calculated.negative === -1.01, '负数临界小数应与后端按相同规则四舍五入')
assert(evaluateWorkflowCalculation(calculationFields[4], calculated).error === '', '有效公式不应返回错误')
const calculationAccess = workflowFieldAccessMap(calculationFields, {}, 'start', 'write')
assert(calculationAccess.scalarTotal === 'read' && calculationAccess.detailTotal === 'read', '计算字段必须固定为只读')
const calculationPayload = writableWorkflowFormData(calculationFields, calculated, calculationAccess)
assert(!('scalarTotal' in calculationPayload) && !('detailTotal' in calculationPayload), '客户端不得提交计算字段')
const permissions = {
  start: [
    { field: 'internalNote', access: 'hidden' },
    { field: 'amount', access: 'write' },
  ],
  manager: [
    { field: 'reason', access: 'read' },
    { field: 'amount', access: 'write' },
    { field: 'attachments', access: 'write' },
    { field: 'internalNote', access: 'hidden' },
    { field: 'objectives', access: 'write', actions: ['add', 'delete'] },
  ],
}

const initial = initialWorkflowFormData(fields, { reason: '客户拜访' })
assert(initial.reason === '客户拜访', 'existing form data should win over field defaults')
assert(initial.amount === 100, 'field defaults should seed missing form values')
assert(initial.region === '', 'tree select fields should default to an empty string')
assert(Array.isArray(initial.department), 'remote multi-select fields should default to an empty array')
assert(Array.isArray(initial.attachments), 'attachment fields should default to an empty array')
const structuredAttachments = normalizeWorkflowAttachments([
  { id: 'uploads/receipt.pdf', name: '报销凭证.pdf', url: '/uploads/receipt.pdf', mimeType: 'application/pdf', size: 1024 },
  '/uploads/legacy.pdf',
])
assert(structuredAttachments[0].name === '报销凭证.pdf', '结构化附件元数据应保持不变')
assert(structuredAttachments[1].url === '/uploads/legacy.pdf', '历史字符串附件应转换为可预览对象')
assert(Array.isArray(initial.objectives), 'detail list fields should default to an empty row array')

const normalizedTreeOptions = normalizeWorkflowOptions([
  {
    name: '总部',
    id: 'hq',
    children: [{ name: '研发部', id: 12 }],
  },
], {
  type: 'api',
  url: '/api/v2/admin/departments/tree',
  labelField: 'name',
  valueField: 'id',
  childrenField: 'children',
})
assert(normalizedTreeOptions[0].label === '总部', 'option source label mapping should be supported')
assert(normalizedTreeOptions[0].children?.[0]?.value === '12', 'option source value mapping should stringify child values')
assert(flattenWorkflowOptions(normalizedTreeOptions).some(option => option.value === '12'), 'tree options should be flattenable for non-dropdown controls')

const detailRow = createWorkflowDetailRow(fields.find(field => field.key === 'objectives'))
assert(typeof detailRow.id === 'string' && detailRow.id.length > 0, 'detail rows should include a stable generated row key')
assert(detailRow.target === '', 'detail row text columns should be initialized')
assert(detailRow.weight === undefined, 'detail row number columns should be initialized as undefined')

const startAccess = workflowFieldAccessMap(fields, permissions, 'start', 'write')
const startActions = workflowFieldActionMap(fields, permissions, 'start', { add: true, delete: true })
assert(startAccess.reason === 'write', 'start node should default to writable fields')
assert(startAccess.internalNote === 'hidden', 'explicit hidden permission should be preserved')
assert(startActions.objectives.add === true && startActions.objectives.delete === true, 'start form should allow detail row edits by default')
assert(visibleWorkflowFormFields(fields, startAccess).length === 6, 'hidden fields should not render')
const startPayload = writableWorkflowFormData(fields, initial, startAccess)
assert(!('internalNote' in startPayload), 'start form payload must not include hidden fields')
const permissionValidationFields = [
  { key: 'reason', label: '申请原因', type: 'textarea', required: true },
  { key: 'managerComment', label: '主管意见', type: 'textarea', required: true },
  { key: 'internalCode', label: '内部编码', type: 'text', required: true },
]
const permissionValidationAccess = {
  reason: 'write',
  managerComment: 'read',
  internalCode: 'hidden',
}
const permissionErrors = validateWorkflowFormData(
  permissionValidationFields,
  { reason: '', managerComment: '', internalCode: '' },
  permissionValidationAccess,
)
assert(permissionErrors.reason === '申请原因不能为空', 'writable required fields must still be validated')
assert(!permissionErrors.managerComment, 'read-only required fields must not be validated')
assert(!permissionErrors.internalCode, 'hidden required fields must not be validated')
assert(
  Object.keys(validateWorkflowFormData(permissionValidationFields, { reason: '出差' }, permissionValidationAccess)).length === 0,
  'only writable fields should participate in form validation',
)
assert(
  runtimeSource.includes('validateWorkflowFormData(props.fields || [], props.modelValue || {}, accessMap.value)'),
  'runtime form must pass the effective field access map into validation',
)
assert(runtimeSource.includes('userNameMap?: Record<string, string>'), '运行时表单应支持用户名映射')
assert(runtimeSource.includes('userDisplayName'), '只读用户字段应显示用户名')

const managerAccess = workflowFieldAccessMap(fields, permissions, 'manager', 'read')
const managerActions = workflowFieldActionMap(fields, permissions, 'manager')
assert(managerActions.objectives.add === true, 'detail list add action should be exposed to runtime form')
assert(managerActions.objectives.delete === true, 'detail list delete action should be exposed to runtime form')
const payload = writableWorkflowFormData(fields, {
  reason: '篡改原因',
  amount: 120,
  attachments: ['receipt.pdf'],
  internalNote: 'secret',
  objectives: [{ id: 'obj-1', target: '提升续费率', weight: 40, result: '已完成' }],
}, managerAccess)
assert(!('reason' in payload), 'read-only fields must not be submitted as a task form patch')
assert(payload.amount === 120, 'writable scalar field should be submitted')
assert(Array.isArray(payload.attachments) && payload.attachments[0].url === 'receipt.pdf', 'writable attachment field should submit normalized metadata')
assert(runtimeSource.includes('attachmentTextModel'), '运行时表单应以附件名称或地址显示结构化附件')
assert(!('internalNote' in payload), 'hidden fields must not be submitted')
assert(Array.isArray(payload.objectives) && payload.objectives[0].target === '提升续费率', 'detail list payload should preserve row objects')

const groupedFields = [{
  key: 'basic_group',
  label: '基本信息',
  type: 'group',
  fields: [
    { key: 'tip', label: '填写提示', type: 'description', content: '请如实填写' },
    { key: 'reason', label: '申请原因', type: 'text', required: true },
    { key: 'secret', label: '内部字段', type: 'text' },
  ],
}]
const groupedPermissions = { start: [{ field: 'secret', access: 'hidden' }] }
const groupedInitial = initialWorkflowFormData(groupedFields, { reason: '出差', tip: '篡改' })
assert(Object.keys(groupedInitial).join(',') === 'reason,secret', '布局组件不应进入初始表单数据')
assert(groupedInitial.reason === '出差', '组内业务字段应读取根级表单数据')
const groupedAccess = workflowFieldAccessMap(groupedFields, groupedPermissions, 'start', 'write')
assert(!('basic_group' in groupedAccess) && !('tip' in groupedAccess), '布局组件不应进入字段权限映射')
assert(groupedAccess.reason === 'write' && groupedAccess.secret === 'hidden', '组内业务字段应进入字段权限映射')
const groupedVisible = visibleWorkflowFormFields(groupedFields, groupedAccess)
assert(groupedVisible.length === 1 && groupedVisible[0].fields.length === 2, '分组应保留展示组件并过滤隐藏业务字段')
const hiddenOnlyGroup = visibleWorkflowFormFields([{
  key: 'hidden_group',
  label: '隐藏分组',
  type: 'group',
  fields: [{ key: 'secret', label: '内部字段', type: 'text' }],
}], { secret: 'hidden' })
assert(hiddenOnlyGroup.length === 0, '只包含隐藏业务字段的空分组不应渲染')

const ruleFields = [
  { key: 'expenseType', label: '报销类型', type: 'select', options: [{ label: '其他', value: 'other' }] },
  {
    key: 'reason', label: '具体说明', type: 'textarea',
    rules: [{ id: 'reason_required', type: 'conditional_required', when: { field: 'expenseType', operator: 'eq', value: 'other' }, message: '选择其他类型时，请填写具体说明' }],
  },
  { key: 'code', label: '单据编号', type: 'text', rules: [{ id: 'code_pattern', type: 'pattern', pattern: '^[A-Z]{2}[0-9]{4}$', message: '单据编号格式不正确' }] },
  { key: 'amount', label: '金额', type: 'amount', rules: [{ id: 'amount_scale', type: 'decimal_places', precision: 2 }] },
  { key: 'attachments', label: '附件', type: 'attachment', rules: [{ id: 'attachment_count', type: 'selection_count', min: 1 }] },
  { key: 'startDate', label: '开始日期', type: 'date' },
  { key: 'endDate', label: '结束日期', type: 'date', rules: [{ id: 'date_order', type: 'compare_field', field: 'startDate', operator: 'gte', message: '结束日期不能早于开始日期' }] },
]
const validRuleData = { expenseType: 'other', reason: '临时采购', code: 'BX1024', amount: 1280.5, attachments: ['receipt.pdf'], startDate: '2026-09-01', endDate: '2026-09-03' }
assert(Object.keys(validateWorkflowFormData(ruleFields, validRuleData)).length === 0, '有效复杂规则数据不应产生错误')
assert(validateWorkflowFormData(ruleFields, { ...validRuleData, reason: '' }).reason === '选择其他类型时，请填写具体说明', '条件必填规则应生效')
assert(validateWorkflowFormData(ruleFields, { ...validRuleData, code: 'bx1024' }).code === '单据编号格式不正确', '格式规则应生效')
assert(validateWorkflowFormData(ruleFields, { ...validRuleData, amount: 100.123 }).amount === '金额小数位不能超过2位', '小数位规则应生效')
assert(validateWorkflowFormData(ruleFields, { ...validRuleData, attachments: [] }).attachments === '附件至少选择1项', '数量规则应生效')
assert(validateWorkflowFormData(ruleFields, { ...validRuleData, endDate: '2026-08-31' }).endDate === '结束日期不能早于开始日期', '字段比较规则应生效')
assert(Object.keys(validateWorkflowFormData(ruleFields, { ...validRuleData, expenseType: 'travel', reason: '' })).length === 0, '未满足条件时不应执行条件必填')

const choiceCompareFields = [
  { key: 'grade', label: '上级分档', type: 'radio', options: [{ label: '01', value: '01' }, { label: '1', value: '1' }] },
  {
    key: 'result', label: '复核结果', type: 'select',
    options: [{ label: '01', value: '01' }, { label: '1', value: '1' }],
    rules: [{ id: 'same_grade', type: 'compare_field', field: 'grade', operator: 'eq' }],
  },
]
assert(Boolean(validateWorkflowFormData(choiceCompareFields, { grade: '1', result: '01' }).result), '选项值 01 与 1 不应按数值判定为相等')
assert(Object.keys(validateWorkflowFormData(choiceCompareFields, { grade: '01', result: '01' })).length === 0, '完全相同的选项值应通过字段比较')

const detailSumField = {
  key: 'objectives',
  label: '我的目标',
  type: 'detail_list',
  columns: [
    { key: 'weight', label: '权重', type: 'number' },
    { key: 'budget', label: '预算', type: 'amount' },
    { key: 'target', label: '目标', type: 'text' },
  ],
  rules: [{ id: 'weight_sum', type: 'column_sum', column: 'weight', operator: 'eq', value: 100 }],
}
const detailSumRows = [
  { id: 'row_1', weight: 40, budget: 0.1, target: '完成方案' },
  { id: 'row_2', weight: 60, budget: 0.2, target: '完成上线' },
  { id: 'row_3', weight: '', budget: undefined, target: '持续优化' },
]
assert(Object.keys(validateWorkflowFormData([detailSumField], { objectives: detailSumRows })).length === 0, '列合计应将空单元格按零计算')

const floatingSumField = {
  ...detailSumField,
  rules: [{ id: 'budget_sum', type: 'column_sum', column: 'budget', operator: 'eq', value: 0.3 }],
}
assert(Object.keys(validateWorkflowFormData([floatingSumField], { objectives: detailSumRows })).length === 0, '列合计应容忍正常浮点运算误差')

for (const test of [
  { operator: 'eq', target: 99 },
  { operator: 'ne', target: 100 },
  { operator: 'gt', target: 100 },
  { operator: 'gte', target: 101 },
  { operator: 'lt', target: 100 },
  { operator: 'lte', target: 99 },
]) {
  const current = {
    ...detailSumField,
    rules: [{ id: `weight_sum_${test.operator}`, type: 'column_sum', column: 'weight', operator: test.operator, value: test.target }],
  }
  assert(Boolean(validateWorkflowFormData([current], { objectives: detailSumRows }).objectives), `列合计应支持 ${test.operator} 比较关系`)
}

const invalidDetailSum = {
  ...detailSumField,
  rules: [{ id: 'weight_sum', type: 'column_sum', column: 'weight', operator: 'eq', value: 99 }],
}
assert(validateWorkflowFormData([invalidDetailSum], { objectives: detailSumRows }).objectives === '我的目标“权重”列合计必须等于99', '列合计应生成明确的默认错误提示')
invalidDetailSum.rules[0].message = '权重合计必须为99'
assert(validateWorkflowFormData([invalidDetailSum], { objectives: detailSumRows }).objectives === '权重合计必须为99', '列合计应优先使用自定义错误提示')

const requiredDetailSum = {
  ...detailSumField,
  columns: detailSumField.columns.map(column => column.key === 'weight' ? { ...column, required: true } : column),
  rules: [{ id: 'weight_sum', type: 'column_sum', column: 'weight', operator: 'eq', value: 100, message: '不应优先显示的合计错误' }],
}
const requiredRows = [
  { id: 'row_1', weight: 40, budget: 0.1, target: '完成方案' },
  { id: 'row_2', weight: '', budget: 0.2, target: '完成上线' },
]
assert(validateWorkflowFormData([requiredDetailSum], { objectives: requiredRows }).objectives === '我的目标第2行：权重不能为空', '明细列必填错误应优先于列合计错误')

for (const [snippet, message] of [
  ["field.type === 'group'", '运行时缺少表单组'],
  ["field.type === 'label'", '运行时缺少标签'],
  ["field.type === 'description'", '运行时缺少静态说明'],
  ["field.type === 'button'", '运行时缺少说明按钮'],
  ['openFieldHelp(field)', '运行时缺少说明入口'],
  ['<el-dialog', '运行时缺少共用说明对话框'],
  ['white-space: pre-wrap', '说明对话框未保留换行'],
  ['defineExpose({ validate', '运行时表单缺少提交前校验入口'],
  [':error="fieldError(field)"', '运行时表单缺少字段行内错误提示'],
  [':autosize="workflowTextareaAutosize(field, 3, 8)"', '普通多行文本未应用显示行数配置'],
  [':autosize="workflowTextareaAutosize(column, 2, 6)"', '明细多行文本未应用显示行数配置'],
  [':show-word-limit="Boolean(field.maxLength)"', '普通多行文本未显示字数统计'],
  [':show-word-limit="Boolean(column.maxLength)"', '明细多行文本未显示字数统计'],
]) {
  assert(runtimeSource.includes(snippet), message)
}
assert(!runtimeSource.includes('v-html'), '运行时说明内容必须使用纯文本渲染')

console.log('workflow runtime form checks passed')
