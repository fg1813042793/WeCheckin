import fs from 'node:fs'
import path from 'node:path'
import { createRequire } from 'node:module'
import ts from 'typescript'

const root = process.cwd()
const read = relativePath => fs.readFileSync(path.join(root, relativePath), 'utf8')
const require = createRequire(import.meta.url)

function loadTypeScriptModule(relativePath) {
  const filename = path.join(root, relativePath)
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
  const runner = new Function('module', 'exports', 'require', output)
  runner(module, module.exports, require)
  return module.exports
}

const typeSource = read('src/types/workflow.ts')
const designerSource = read('src/views/workflow/designer/index.vue')
const formDesignerSource = read('src/views/workflow/designer/components/WorkflowFormDesigner.vue')
const fieldPreviewSource = read('src/views/workflow/designer/components/WorkflowFormFieldPreview.vue')
const validationRulesSource = read('src/views/workflow/designer/components/WorkflowValidationRulesEditor.vue')
const permissionSource = read('src/views/workflow/designer/components/WorkflowFieldPermissions.vue')
const validationRuleUtilsPath = 'src/views/workflow/workflowValidationRules.ts'
if (!fs.existsSync(path.join(root, validationRuleUtilsPath))) {
  throw new Error('表单设计器缺少跨字段校验兼容矩阵')
}
const {
  insertWorkflowField,
  moveWorkflowDetailColumn,
  moveWorkflowField,
  removeWorkflowField,
  workflowDataFieldEntries,
  workflowDataFields,
  workflowFieldByKey,
} = loadTypeScriptModule('src/views/workflow/formLayout.ts')
const {
  workflowCompareFieldCompatible,
  workflowCompareOperators,
} = loadTypeScriptModule(validationRuleUtilsPath)

for (const [left, right] of [
  ['number', 'amount'],
  ['text', 'email'],
  ['select', 'radio'],
  ['boolean', 'boolean'],
  ['user', 'user'],
  ['department', 'department'],
  ['date', 'date'],
]) {
  if (!workflowCompareFieldCompatible(left, right)) {
    throw new Error(`应允许 ${left} 与 ${right} 字段比较`)
  }
}

for (const [left, right] of [
  ['user', 'department'],
  ['select', 'text'],
  ['date', 'datetime'],
  ['attachment', 'attachment'],
  ['multi_select', 'multi_select'],
  ['detail_list', 'detail_list'],
]) {
  if (workflowCompareFieldCompatible(left, right)) {
    throw new Error(`不应允许 ${left} 与 ${right} 字段比较`)
  }
}

if (workflowCompareOperators('select').join(',') !== 'eq,ne') {
  throw new Error('单选和下拉字段只应支持等于、不等于')
}
if (workflowCompareOperators('boolean').join(',') !== 'eq,ne') {
  throw new Error('布尔字段只应支持等于、不等于')
}
if (workflowCompareOperators('amount').join(',') !== 'eq,ne,gt,gte,lt,lte') {
  throw new Error('数字与金额字段应支持全部大小比较关系')
}
if (workflowCompareOperators('date').join(',') !== 'eq,ne,gt,gte,lt,lte') {
  throw new Error('日期与时间字段应支持全部大小比较关系')
}

for (const [snippet, message] of [
  ['grid-template-columns: 320px minmax(480px, 1fr) 440px', '宽屏下字段属性面板应扩大为 440px'],
  ['grid-template-columns: 200px minmax(420px, 1fr) 380px', '中等屏幕字段属性面板应保留 380px'],
  ['grid-template-columns: 176px minmax(360px, 1fr) 320px', '窄屏字段属性面板应保留 320px'],
  ['workflowCompareFieldCompatible', '高级校验未按字段类型筛选可比较字段'],
  ['fieldCompareOperators', '高级校验未按当前字段限制比较关系'],
]) {
  const source = snippet.startsWith('grid-template-columns') ? formDesignerSource : validationRulesSource
  if (!source.includes(snippet)) throw new Error(message)
}

const groupedFields = [{
  key: 'basic_group',
  label: '基本信息',
  type: 'group',
  fields: [
    { key: 'tip', label: '提示', type: 'description', content: '请填写' },
    { key: 'reason', label: '原因', type: 'text', required: true },
  ],
}, {
  key: 'another_group',
  label: '其他信息',
  type: 'group',
  fields: [],
}]

if (workflowDataFields(groupedFields).map(item => item.key).join(',') !== 'reason') {
  throw new Error('表单布局工具应只展平组内业务字段')
}
const groupedFieldEntry = workflowDataFieldEntries(groupedFields)[0]
if (groupedFieldEntry?.field.key !== 'reason' || groupedFieldEntry?.group?.label !== '基本信息') {
  throw new Error('字段权限展平组内字段时应保留所属表单组')
}
if (workflowFieldByKey(groupedFields, 'reason')?.label !== '原因') {
  throw new Error('表单布局工具应递归查找组内组件')
}
if (!moveWorkflowField(groupedFields, 'reason', null, 0) || groupedFields[0]?.key !== 'reason') {
  throw new Error('组内字段应可移动到根级')
}
if (moveWorkflowField(groupedFields, 'basic_group', 'another_group', 0)) {
  throw new Error('表单组不应嵌套')
}
if (removeWorkflowField(groupedFields, 'tip')?.key !== 'tip') {
  throw new Error('表单布局工具应递归删除组内组件')
}

const insertionFields = [{ key: 'group', label: '分组', type: 'group', fields: [] }]
if (!insertWorkflowField(insertionFields, { key: 'first', label: '组件一', type: 'text' }, 'group', 0)) {
  throw new Error('应支持向空分组插入组件')
}
if (!insertWorkflowField(insertionFields, { key: 'second', label: '组件二', type: 'description', content: '说明' }, 'group', 1)) {
  throw new Error('应支持向已有组件的分组继续插入组件')
}
if (insertionFields[0].fields.map(item => item.key).join(',') !== 'first,second') {
  throw new Error('分组应保留连续插入的多个组件')
}
if (!insertWorkflowField(insertionFields, { key: 'last', label: '末尾组件', type: 'text' }, null, insertionFields.length)) {
  throw new Error('应支持向表单末尾插入组件')
}
if (insertionFields.at(-1)?.key !== 'last') {
  throw new Error('表单末尾落点应把组件放在最后')
}

const moveFields = [
  { key: 'group', label: '分组', type: 'group', fields: [] },
  { key: 'first', label: '组件一', type: 'text' },
  { key: 'second', label: '组件二', type: 'text' },
  { key: 'last', label: '最后组件', type: 'text' },
]
if (!moveWorkflowField(moveFields, 'first', 'group', 0) || !moveWorkflowField(moveFields, 'second', 'group', 1)) {
  throw new Error('应支持连续把多个现有组件拖入同一分组')
}
if (moveFields[0].fields.map(item => item.key).join(',') !== 'first,second') {
  throw new Error('连续拖入分组的组件顺序不正确')
}
if (!moveWorkflowField(moveFields, 'last', null, 0) || moveFields[0]?.key !== 'last') {
  throw new Error('最后一个组件应可拖动到表单其他位置')
}

const detailColumnOrder = [
  { key: 'target', label: '目标', type: 'text' },
  { key: 'weight', label: '权重', type: 'number' },
  { key: 'result', label: '结果', type: 'textarea' },
]
if (!moveWorkflowDetailColumn(detailColumnOrder, 0, 3) || detailColumnOrder.map(item => item.key).join(',') !== 'weight,result,target') {
  throw new Error('明细列应支持拖动到列表末尾')
}
if (!moveWorkflowDetailColumn(detailColumnOrder, 2, 0) || detailColumnOrder.map(item => item.key).join(',') !== 'target,weight,result') {
  throw new Error('明细列应支持拖动到列表开头')
}
if (moveWorkflowDetailColumn(detailColumnOrder, 1, 2)) {
  throw new Error('明细列拖入原位置不应触发变更')
}

for (const hiddenRangeDefault of [
  'field.min = 0',
  "field.max = type === 'amount' ? 1000000 : 100",
  'column.min ??= 0',
  "column.max ??= column.type === 'amount' ? 1000000 : 100",
]) {
  if (formDesignerSource.includes(hiddenRangeDefault)) {
    throw new Error(`数字字段不应自动写入隐藏范围：${hiddenRangeDefault}`)
  }
}

const detailColumnFactorySource = formDesignerSource.slice(
  formDesignerSource.indexOf('function buildDetailColumn'),
  formDesignerSource.indexOf('function addDetailColumn'),
)
if (detailColumnFactorySource.includes('maxLength')) {
  throw new Error('新增明细列不应自动写入最大长度')
}
if (formDesignerSource.includes('column.maxLength ??=')) {
  throw new Error('切换明细列类型不应自动写入最大长度')
}

for (const [snippet, message] of [
  ['class="detail-column-range"', '明细数字列缺少范围设置区域'],
  ["updateDetailColumnRange(column, 'min', $event)", '明细数字列缺少最小值更新入口'],
  ["updateDetailColumnRange(column, 'max', $event)", '明细数字列缺少最大值更新入口'],
  ['delete column[key]', '清空明细数字范围时应删除对应属性'],
  ['class="detail-column-max-length"', '明细文本列缺少最大长度设置'],
  ['updateDetailColumnMaxLength(column, $event)', '明细文本列缺少最大长度清空入口'],
  ['textarea-visible-rows', '多行文本字段缺少显示行数设置区域'],
  ['.textarea-visible-rows { grid-column: 1 / -1;', '明细多行文本显示行数设置未占满属性区域'],
  ['.textarea-visible-rows :deep(.el-form-item) { min-width: 0;', '多行文本显示行数输入框缺少收缩保护'],
  ['label="最小显示行数"', '多行文本字段缺少最小显示行数设置'],
  ['label="最大显示行数"', '多行文本字段缺少最大显示行数设置'],
  ["updateTextareaVisibleRows(selectedField, 'minVisibleRows', $event, 3, 8)", '普通多行文本缺少最小显示行数更新逻辑'],
  ["updateTextareaVisibleRows(column, 'maxVisibleRows', $event, 2, 6)", '明细多行文本缺少最大显示行数更新逻辑'],
  ['field.minVisibleRows = 3', '新建多行文本缺少默认最小显示行数'],
  ['field.maxVisibleRows = 8', '新建多行文本缺少默认最大显示行数'],
  ['minVisibleRows: 2, maxVisibleRows: 6', '默认明细多行列缺少显示行数配置'],
  ['delete column.minVisibleRows', '明细列切换为非多行类型时应清理最小显示行数'],
  ['delete column.maxVisibleRows', '明细列切换为非多行类型时应清理最大显示行数'],
]) {
  if (!formDesignerSource.includes(snippet)) throw new Error(message)
}

for (const property of ['minVisibleRows?: number', 'maxVisibleRows?: number']) {
  if (!typeSource.includes(property)) throw new Error(`流程字段类型缺少属性：${property}`)
}

for (const [source, snippet, message] of [
  [formDesignerSource, ':autosize="textareaAutosize(field, 3, 8)"', '根级多行文本预览未读取显示行数'],
  [fieldPreviewSource, ':autosize="textareaAutosize(field, 3, 8)"', '分组内多行文本预览未读取显示行数'],
]) {
  if (!source.includes(snippet)) throw new Error(message)
}

const detailPreviewRequirements = [
  ['detail-preview__grid', '缺少统一明细列网格'],
  ['detail-preview__cell', '明细字段名称与输入预览未绑定到同一网格单元'],
  ['detail-preview__label', '明细字段单元缺少名称'],
  ['detail-preview__control', '明细字段单元缺少输入预览'],
]

for (const [source, surface] of [
  [formDesignerSource, '根级明细预览'],
  [fieldPreviewSource, '分组内明细预览'],
]) {
  for (const [snippet, message] of detailPreviewRequirements) {
    if (!source.includes(snippet)) throw new Error(`${surface}${message}`)
  }
  for (const legacyClass of ['detail-preview__head', 'detail-preview__row']) {
    if (source.includes(legacyClass)) throw new Error(`${surface}仍使用独立名称和输入网格：${legacyClass}`)
  }
}

const requirements = [
  [typeSource, "export type WorkflowFormFieldType", '缺少流程表单字段类型'],
  [typeSource, 'form: WorkflowFormField[]', '流程草稿缺少 form 字段'],
  [typeSource, 'span?: WorkflowFormFieldSpan', '流程字段缺少布局宽度'],
  [typeSource, "'detail_list'", '流程字段类型缺少明细列表'],
  [typeSource, 'children?: WorkflowFormOption[]', '流程字段选项缺少树形 children 支持'],
  [typeSource, "export type WorkflowOptionSourceType = 'static' | 'api'", '流程字段缺少选项来源类型'],
  [typeSource, 'optionSource?: WorkflowOptionSource', '流程字段缺少后端接口选项来源配置'],
  [typeSource, 'columns?: WorkflowFormField[]', '明细列表字段缺少列定义'],
  [typeSource, 'fields?: WorkflowFormField[]', '流程表单缺少分组子组件定义'],
  [typeSource, 'help?: WorkflowFormHelp', '流程表单缺少说明配置'],
  [typeSource, 'rules?: WorkflowValidationRule[]', '流程字段缺少复杂校验规则'],
  [typeSource, "'column_sum'", '流程校验规则缺少明细列合计类型'],
  [typeSource, 'column?: string', '流程校验规则缺少明细列编码'],
  [typeSource, 'value?: number', '流程校验规则缺少列合计目标值'],
  [typeSource, "export type WorkflowDetailRowAction = 'add' | 'delete'", '缺少明细列表行级动作类型'],
  [typeSource, 'actions?: WorkflowDetailRowAction[]', '字段权限缺少行级动作配置'],
  [typeSource, 'formPermissions?: WorkflowFieldPermission[]', '流程节点缺少字段权限'],
  [designerSource, 'WorkflowFormDesigner', '流程设计器未接入表单设计'],
  [designerSource, 'WorkflowFieldPermissions', '流程设计器未接入字段权限'],
  [designerSource, '表单设计', '流程设计器缺少表单设计页签'],
  [designerSource, '流程设计', '流程设计器缺少流程设计页签'],
  [designerSource, '字段权限', '流程设计器缺少字段权限页签'],
  [formDesignerSource, "type: 'textarea'", '表单设计器缺少多行文本字段'],
  [formDesignerSource, "type: 'user'", '表单设计器缺少人员字段'],
  [formDesignerSource, "type: 'department'", '表单设计器缺少部门字段'],
  [formDesignerSource, "type: 'attachment'", '表单设计器缺少附件字段'],
  [formDesignerSource, "type: 'amount'", '表单设计器缺少金额字段'],
  [formDesignerSource, "type: 'phone'", '表单设计器缺少手机号字段'],
  [formDesignerSource, "type: 'email'", '表单设计器缺少邮箱字段'],
  [formDesignerSource, "type: 'radio'", '表单设计器缺少单选框字段'],
  [formDesignerSource, "type: 'checkbox'", '表单设计器缺少复选框字段'],
  [formDesignerSource, "type: 'time'", '表单设计器缺少时间字段'],
  [formDesignerSource, "type: 'date_range'", '表单设计器缺少日期区间字段'],
  [formDesignerSource, "type: 'user_multi'", '表单设计器缺少多人字段'],
  [formDesignerSource, "type: 'department_multi'", '表单设计器缺少多部门字段'],
  [formDesignerSource, "type: 'detail_list'", '表单设计器缺少明细列表字段'],
  [formDesignerSource, "type: 'group'", '表单设计器缺少表单组'],
  [formDesignerSource, "type: 'label'", '表单设计器缺少标签'],
  [formDesignerSource, "type: 'description'", '表单设计器缺少说明'],
  [formDesignerSource, "type: 'button'", '表单设计器缺少说明按钮'],
  [formDesignerSource, 'group.fields', '表单组缺少子组件画布'],
  [formDesignerSource, 'moveWorkflowField', '表单设计器缺少跨容器拖拽'],
  [formDesignerSource, 'insertWorkflowField', '表单设计器缺少调色板拖入逻辑'],
  [formDesignerSource, '@dragstart="handlePaletteDragStart', '左侧字段组件缺少拖拽开始事件'],
  [formDesignerSource, ':draggable="!readonly"', '表单组件卡片未提供稳定拖拽区域'],
  [formDesignerSource, 'class="field-drag-handle"\n                      draggable="true"', '拖动图标本身未配置为拖拽源'],
  [formDesignerSource, 'empty-canvas-drop', '空表单缺少调色板拖入落点'],
  [formDesignerSource, 'root-tail-drop-zone', '表单末尾缺少稳定拖放区域'],
  [formDesignerSource, '删除分组将同时删除组内组件，确定继续？', '非空表单组删除前缺少确认'],
  [formDesignerSource, '说明设置', '字段属性缺少说明设置'],
  [formDesignerSource, '说明标题', '字段属性缺少说明标题'],
  [formDesignerSource, '说明内容', '字段属性缺少说明内容'],
  [formDesignerSource, '明细列', '明细列表缺少列配置区域'],
  [formDesignerSource, 'addDetailColumn', '明细列表缺少增加列能力'],
  [formDesignerSource, 'removeDetailColumn', '明细列表缺少删除列能力'],
  [formDesignerSource, 'detail-column-drag-handle', '明细列表列配置缺少拖拽手柄'],
  [formDesignerSource, 'handleDetailColumnDragStart', '明细列表缺少列拖拽开始逻辑'],
  [formDesignerSource, 'handleDetailColumnDragOver', '明细列表缺少列拖拽落点计算'],
  [formDesignerSource, 'handleDetailColumnDrop', '明细列表缺少列拖拽排序逻辑'],
  [formDesignerSource, 'detail-column-tail-drop', '明细列表缺少拖放到末尾的明确落点'],
  [formDesignerSource, 'moveDetailColumnByOffset', '明细列表缺少精确上移和下移能力'],
  [formDesignerSource, 'title="上移一位"', '明细列缺少上移按钮'],
  [formDesignerSource, 'title="下移一位"', '明细列缺少下移按钮'],
  [formDesignerSource, 'detail-column-layout', '明细列表列配置缺少宽度设置'],
  [formDesignerSource, 'updateDetailColumnSpan', '明细列表缺少列宽更新逻辑'],
  [formDesignerSource, 'fieldSpan(column)', '明细列表画布预览未应用列宽'],
  [fieldPreviewSource, 'columnSpan(column)', '分组内明细列表预览未应用列宽'],
  [fieldPreviewSource, 'repeat(24, minmax(0, 1fr))', '分组内明细列表预览缺少 24 栅格'],
  [formDesignerSource, 'rowKey', '明细列表缺少行标识配置'],
  [formDesignerSource, '选项来源', '选项配置缺少来源切换'],
  [formDesignerSource, 'JSON', '选项配置缺少 JSON 输入'],
  [formDesignerSource, 'optionJsonText', '选项配置缺少 JSON 编辑状态'],
  [formDesignerSource, 'applyOptionsJson', '选项配置缺少 JSON 应用逻辑'],
  [formDesignerSource, '接口配置', '选项配置缺少接口配置区域'],
  [formDesignerSource, '后端接口', '选项配置缺少后端接口地址配置'],
  [formDesignerSource, 'responsePath', '选项接口配置缺少响应路径映射'],
  [formDesignerSource, 'labelField', '选项接口配置缺少名称字段映射'],
  [formDesignerSource, 'valueField', '选项接口配置缺少值字段映射'],
  [formDesignerSource, 'childrenField', '选项接口配置缺少子级字段映射'],
  [formDesignerSource, 'el-tree-select', '表单设计器缺少树形下拉预览'],
  [formDesignerSource, 'fieldGroups', '字段组件缺少分类展示'],
  [formDesignerSource, 'field-preview', '表单画布缺少字段控件预览'],
  [formDesignerSource, 'property-section', '字段属性缺少分组结构'],
  [formDesignerSource, 'field-drag-handle', '表单字段缺少拖拽手柄'],
  [formDesignerSource, '@dragstart="handleDragStart', '表单字段缺少拖拽开始事件'],
  [formDesignerSource, 'field-drop-zone--tail', '表单字段缺少置底拖放区域'],
  [formDesignerSource, '字段宽度', '表单字段缺少宽度配置'],
  [validationRulesSource, '添加规则', '表单字段缺少新增校验规则入口'],
  [validationRulesSource, 'addValidationRule', '表单字段缺少新增校验规则逻辑'],
  [validationRulesSource, 'removeValidationRule', '表单字段缺少删除校验规则逻辑'],
  [validationRulesSource, "rule.type === 'column_sum'", '明细列表缺少列合计规则配置'],
  [validationRulesSource, "label: '列合计'", '明细列表缺少列合计规则类型'],
  [validationRulesSource, 'summableColumns', '列合计规则缺少数字和金额列筛选'],
  [validationRulesSource, 'workflowCompareFieldCompatible', '普通字段比较未使用兼容类型矩阵'],
  [validationRulesSource, "rule.field = ''", '新增字段比较规则不应默认关联目标字段'],
  [formDesignerSource, 'WorkflowValidationRulesEditor', '表单设计器未接入高级校验规则编辑器'],
  [formDesignerSource, 'grid-template-columns: repeat(24, minmax(0, 1fr))', '表单画布缺少 24 列栅格'],
  [formDesignerSource, 'fieldSpan(field)', '表单字段未应用布局宽度'],
  [permissionSource, "value=\"hidden\"", '字段权限缺少隐藏选项'],
  [permissionSource, "value=\"read\"", '字段权限缺少只读选项'],
  [permissionSource, "value=\"write\"", '字段权限缺少可编辑选项'],
  [permissionSource, "value=\"add\"", '字段权限缺少新增行选项'],
  [permissionSource, "value=\"delete\"", '字段权限缺少删除行选项'],
  [permissionSource, '新增行', '字段权限缺少新增行文案'],
  [permissionSource, '删除行', '字段权限缺少删除行文案'],
  [permissionSource, 'updateFieldActions', '字段权限缺少行级动作更新逻辑'],
  [permissionSource, "node.type === 'handle'", '字段权限缺少办理节点支持'],
  [permissionSource, 'class="row-action-slot"', '明细字段权限未预留稳定的行操作区域'],
  [permissionSource, 'vertical-align: top', '字段权限表格单元格未保持顶部对齐'],
  [permissionSource, 'padding: 0 20px 24px', '字段权限滚动容器顶部不应保留可透视内容的内边距'],
  [permissionSource, 'margin-top: 16px', '字段权限表格缺少初始顶部间距'],
  [permissionSource, 'field-column-header', '字段权限首列表头缺少独立样式'],
  [permissionSource, 'field-group-title', '字段权限未显示字段所属表单组标题'],
  [permissionSource, 'workflowPermissionNodes', '字段权限未按照流程树顺序排列节点'],
  [permissionSource, 'workflowDataFieldEntries', '字段权限未保留分组上下文并展平业务字段'],
  [designerSource, 'workflowDataFields', '表单变更未递归维护节点权限'],
]

for (const [source, snippet, message] of requirements) {
  if (!source.includes(snippet)) throw new Error(message)
}

if (validationRulesSource.includes('<el-empty')) {
  throw new Error('高级校验规则为空时不应显示占位图片')
}

const forbiddenSnippets = [
  [formDesignerSource, '@click="moveField', '表单字段已支持拖拽排序，不应保留上移/下移点击处理'],
  [formDesignerSource, 'function moveField', '表单字段已支持拖拽排序，不应保留上移/下移函数'],
  [validationRulesSource, "rule.field = compareFields.value[0]?.key || ''", '新增字段比较规则不应默认关联第一个字段'],
  [permissionSource, 'padding: 16px 20px 24px', '字段权限滚动容器顶部内边距会导致内容从粘性表头上方透出'],
]

for (const [source, snippet, message] of forbiddenSnippets) {
  if (source.includes(snippet)) throw new Error(message)
}

console.log('workflow form designer structure checks passed')
