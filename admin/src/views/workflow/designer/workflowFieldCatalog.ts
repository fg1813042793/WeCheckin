import type { WorkflowFormFieldType } from '../types'

export interface WorkflowFieldTypeMeta {
  type: WorkflowFormFieldType
  label: string
  icon: string
}

export const workflowFieldTypes: WorkflowFieldTypeMeta[] = [
  { type: 'text', label: '单行文本', icon: 'EditPen' },
  { type: 'textarea', label: '多行文本', icon: 'Document' },
  { type: 'number', label: '数字', icon: 'Odometer' },
  { type: 'amount', label: '金额', icon: 'Money' },
  { type: 'phone', label: '手机号', icon: 'Cellphone' },
  { type: 'email', label: '邮箱', icon: 'Message' },
  { type: 'boolean', label: '开关', icon: 'Switch' },
  { type: 'select', label: '单选下拉', icon: 'Select' },
  { type: 'multi_select', label: '多选下拉', icon: 'Finished' },
  { type: 'radio', label: '单选框', icon: 'CircleCheck' },
  { type: 'checkbox', label: '复选框', icon: 'Checked' },
  { type: 'date', label: '日期', icon: 'Calendar' },
  { type: 'datetime', label: '日期时间', icon: 'Clock' },
  { type: 'time', label: '时间', icon: 'Timer' },
  { type: 'date_range', label: '日期区间', icon: 'Calendar' },
  { type: 'user', label: '人员', icon: 'User' },
  { type: 'user_multi', label: '多人', icon: 'UserFilled' },
  { type: 'department', label: '部门', icon: 'OfficeBuilding' },
  { type: 'department_multi', label: '多部门', icon: 'CopyDocument' },
  { type: 'attachment', label: '附件', icon: 'Paperclip' },
  { type: 'detail_list', label: '明细列表', icon: 'Tickets' },
  { type: 'calculation', label: '计算组件', icon: 'Operation' },
  { type: 'group', label: '表单组', icon: 'FolderOpened' },
  { type: 'label', label: '标签', icon: 'CollectionTag' },
  { type: 'description', label: '说明', icon: 'InfoFilled' },
  { type: 'button', label: '说明按钮', icon: 'Pointer' },
]

export const workflowFieldGroups = [
  { label: '布局与说明', items: workflowFieldTypes.filter(item => ['group', 'label', 'description', 'button'].includes(item.type)) },
  { label: '基础字段', items: workflowFieldTypes.filter(item => ['text', 'textarea', 'number', 'amount', 'calculation', 'phone', 'email', 'boolean'].includes(item.type)) },
  { label: '选择与时间', items: workflowFieldTypes.filter(item => ['select', 'multi_select', 'radio', 'checkbox', 'date', 'datetime', 'time', 'date_range'].includes(item.type)) },
  { label: '组织与附件', items: workflowFieldTypes.filter(item => ['user', 'user_multi', 'department', 'department_multi', 'attachment'].includes(item.type)) },
  { label: '明细字段', items: workflowFieldTypes.filter(item => item.type === 'detail_list') },
]

export const workflowDetailColumnTypes = workflowFieldTypes.filter(item => ![
  'detail_list', 'calculation', 'attachment', 'user', 'user_multi', 'department', 'department_multi',
  'group', 'label', 'description', 'button',
].includes(item.type))
