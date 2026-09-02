import type {
  WorkflowDetailRowAction,
  WorkflowFieldAccess,
  WorkflowFieldPermission,
  WorkflowFormField,
  WorkflowFormFieldType,
  WorkflowFormOption,
  WorkflowOptionSource,
  WorkflowValidationCondition,
  WorkflowValidationOperator,
  WorkflowValidationRule,
} from './types'
import { isWorkflowDataField, workflowDataFields } from './formLayout'

export type WorkflowFormData = Record<string, unknown>
export type WorkflowFieldAccessMap = Record<string, WorkflowFieldAccess>
export type WorkflowFieldActionMap = Record<string, { add: boolean; delete: boolean }>
export type WorkflowNodeFieldPermissions = Record<string, WorkflowFieldPermission[] | undefined>
export type WorkflowDefaultFieldActions = Partial<Record<WorkflowDetailRowAction, boolean>>
export type WorkflowFormValidationErrors = Record<string, string>

const arrayFieldTypes = new Set<WorkflowFormFieldType>([
  'multi_select',
  'checkbox',
  'date_range',
  'user_multi',
  'department_multi',
  'attachment',
  'detail_list',
])

export function normalizeWorkflowOptions(options: unknown, source?: Partial<WorkflowOptionSource>): WorkflowFormOption[] {
  if (!Array.isArray(options)) return []
  const labelField = source?.labelField?.trim() || 'label'
  const valueField = source?.valueField?.trim() || 'value'
  const childrenField = source?.childrenField?.trim() || 'children'
  const result: WorkflowFormOption[] = []
  for (const item of options) {
    if (!item || typeof item !== 'object' || Array.isArray(item)) continue
    const record = item as Record<string, unknown>
    const rawLabel = optionPathValue(record, labelField)
    const rawValue = optionPathValue(record, valueField)
    const label = rawLabel === undefined || rawLabel === null ? '' : String(rawLabel).trim()
    const value = rawValue === undefined || rawValue === null ? '' : String(rawValue).trim()
    if (!label || !value) continue
    const children = normalizeWorkflowOptions(optionPathValue(record, childrenField), source)
    result.push(children.length > 0 ? { label, value, children } : { label, value })
  }
  return result
}

export function flattenWorkflowOptions(options: WorkflowFormOption[] = []): WorkflowFormOption[] {
  const result: WorkflowFormOption[] = []
  for (const option of options) {
    result.push({ label: option.label, value: option.value })
    if (Array.isArray(option.children) && option.children.length > 0) {
      result.push(...flattenWorkflowOptions(option.children))
    }
  }
  return result
}

export function hasWorkflowOptionChildren(options: WorkflowFormOption[] = []): boolean {
  return options.some((option) => Array.isArray(option.children) && option.children.length > 0)
}

export function optionPathValue(record: Record<string, unknown>, path: string): unknown {
  const segments = path.split('.').map((item) => item.trim()).filter(Boolean)
  if (segments.length === 0) return undefined
  let current: unknown = record
  for (const segment of segments) {
    if (!current || typeof current !== 'object' || Array.isArray(current)) return undefined
    current = (current as Record<string, unknown>)[segment]
  }
  return current
}

function cloneWorkflowValue(value: unknown): unknown {
  if (Array.isArray(value)) return value.map((item) => cloneWorkflowValue(item))
  if (value && typeof value === 'object') return JSON.parse(JSON.stringify(value))
  return value
}

function generatedDetailRowID() {
  return `row_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
}

export function isArrayWorkflowField(field: Pick<WorkflowFormField, 'type'>): boolean {
  return arrayFieldTypes.has(field.type)
}

export function emptyWorkflowFieldValue(field: Pick<WorkflowFormField, 'type'>): unknown {
  if (isArrayWorkflowField(field)) return []
  if (field.type === 'boolean') return false
  if (field.type === 'number' || field.type === 'amount') return undefined
  return ''
}

export function initialWorkflowFormData(fields: WorkflowFormField[], existing: WorkflowFormData = {}): WorkflowFormData {
  const result: WorkflowFormData = {}
  for (const field of workflowDataFields(fields)) {
    if (!field?.key) continue
    if (Object.prototype.hasOwnProperty.call(existing, field.key)) {
      result[field.key] = normalizeWorkflowFormValue(field, cloneWorkflowValue(existing[field.key]))
      continue
    }
    if (Object.prototype.hasOwnProperty.call(field, 'default')) {
      result[field.key] = normalizeWorkflowFormValue(field, cloneWorkflowValue(field.default))
      continue
    }
    result[field.key] = emptyWorkflowFieldValue(field)
  }
  return result
}

export function workflowFieldAccessMap(
  fields: WorkflowFormField[],
  permissionsByNode: WorkflowNodeFieldPermissions | undefined,
  nodeId: string | undefined,
  defaultAccess: WorkflowFieldAccess,
): WorkflowFieldAccessMap {
  const accessByField: WorkflowFieldAccessMap = {}
  for (const field of workflowDataFields(fields)) {
    if (field?.key) accessByField[field.key] = defaultAccess
  }
  const permissions = nodeId ? permissionsByNode?.[nodeId] : undefined
  if (!Array.isArray(permissions)) return accessByField
  for (const permission of permissions) {
    if (!permission?.field || !(permission.field in accessByField)) continue
    if (permission.access === 'hidden' || permission.access === 'read' || permission.access === 'write') {
      accessByField[permission.field] = permission.access
    }
  }
  return accessByField
}

export function workflowFieldActionMap(
  fields: WorkflowFormField[],
  permissionsByNode: WorkflowNodeFieldPermissions | undefined,
  nodeId: string | undefined,
  defaultActions: WorkflowDefaultFieldActions = {},
): WorkflowFieldActionMap {
  const actionByField: WorkflowFieldActionMap = {}
  const detailFields = new Set<string>()
  for (const field of workflowDataFields(fields)) {
    if (field?.key && field.type === 'detail_list') {
      detailFields.add(field.key)
      actionByField[field.key] = {
        add: Boolean(defaultActions.add),
        delete: Boolean(defaultActions.delete),
      }
    }
  }
  const permissions = nodeId ? permissionsByNode?.[nodeId] : undefined
  if (!Array.isArray(permissions)) return actionByField
  for (const permission of permissions) {
    if (!permission?.field || !detailFields.has(permission.field) || permission.access !== 'write') continue
    if (!Array.isArray(permission.actions) || permission.actions.length === 0) continue
    const actions = new Set<WorkflowDetailRowAction>((permission.actions || []).filter((action): action is WorkflowDetailRowAction => action === 'add' || action === 'delete'))
    actionByField[permission.field] = {
      add: actions.has('add'),
      delete: actions.has('delete'),
    }
  }
  return actionByField
}

export function visibleWorkflowFormFields(fields: WorkflowFormField[], accessByField: WorkflowFieldAccessMap): WorkflowFormField[] {
  const result: WorkflowFormField[] = []
  for (const field of fields) {
    if (!field?.key) continue
    if (field.type === 'group') {
      const visibleChildren = visibleWorkflowFormFields(field.fields || [], accessByField)
      if (visibleChildren.length > 0) result.push({ ...field, fields: visibleChildren })
      continue
    }
    if (!isWorkflowDataField(field) || accessByField[field.key] !== 'hidden') result.push(field)
  }
  return result
}

export function normalizeWorkflowFormValue(field: WorkflowFormField, value: unknown): unknown {
  if (value === undefined) return value
  if (field.type === 'detail_list') return normalizeDetailRows(field, value)
  if (!isArrayWorkflowField(field)) return value
  if (Array.isArray(value)) return value.map((item) => String(item)).filter((item) => item.trim() !== '')
  if (typeof value === 'string') {
    return value
      .split(/\r?\n/)
      .map((item) => item.trim())
      .filter(Boolean)
  }
  return []
}

export function workflowDetailRowKey(field?: Pick<WorkflowFormField, 'rowKey'>): string {
  const rowKey = String(field?.rowKey || '').trim()
  return rowKey || 'id'
}

export function createWorkflowDetailRow(field?: WorkflowFormField): Record<string, unknown> {
  const row: Record<string, unknown> = { [workflowDetailRowKey(field)]: generatedDetailRowID() }
  for (const column of field?.columns || []) {
    if (!column?.key) continue
    row[column.key] = emptyWorkflowFieldValue(column)
  }
  return row
}

function normalizeDetailRows(field: WorkflowFormField, value: unknown): Record<string, unknown>[] {
  if (!Array.isArray(value)) return []
  const rowKey = workflowDetailRowKey(field)
  return value
    .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object' && !Array.isArray(item))
    .map((item) => {
      const row: Record<string, unknown> = {}
      const rawRowID = item[rowKey]
      row[rowKey] = typeof rawRowID === 'string' && rawRowID.trim() ? rawRowID : generatedDetailRowID()
      for (const column of field.columns || []) {
        if (!column?.key) continue
        const value = Object.prototype.hasOwnProperty.call(item, column.key)
          ? item[column.key]
          : emptyWorkflowFieldValue(column)
        row[column.key] = normalizeWorkflowFormValue(column, value)
      }
      return row
    })
}

export function writableWorkflowFormData(
  fields: WorkflowFormField[],
  values: WorkflowFormData,
  accessByField: WorkflowFieldAccessMap,
): WorkflowFormData {
  const result: WorkflowFormData = {}
  for (const field of workflowDataFields(fields)) {
    if (!field?.key || accessByField[field.key] !== 'write') continue
    const value = normalizeWorkflowFormValue(field, values[field.key])
    if (value !== undefined) result[field.key] = cloneWorkflowValue(value)
  }
  return result
}

export function validateWorkflowFormData(fields: WorkflowFormField[], values: WorkflowFormData): WorkflowFormValidationErrors {
  const errors: WorkflowFormValidationErrors = {}
  for (const field of workflowDataFields(fields)) {
    const value = values[field.key]
    const isDetailList = field.type === 'detail_list'
    const message = validateWorkflowField(field, value, values, !isDetailList)
    if (message) {
      errors[field.key] = message
      continue
    }
    if (!isDetailList) continue
    const rows = Array.isArray(value) ? value as Array<Record<string, unknown>> : []
    for (let rowIndex = 0; rowIndex < rows.length; rowIndex += 1) {
      const row = rows[rowIndex]
      if (!row || typeof row !== 'object' || Array.isArray(row)) {
        errors[field.key] = `${field.label || field.key}第${rowIndex + 1}行数据无效`
        break
      }
      for (const column of field.columns || []) {
        const columnMessage = validateWorkflowField(column, row[column.key], row)
        if (columnMessage) {
          errors[field.key] = `${field.label || field.key}第${rowIndex + 1}行：${columnMessage}`
          break
        }
      }
      if (errors[field.key]) break
    }
    if (errors[field.key]) continue
    const ruleMessage = validateWorkflowRules(field, value, values)
    if (ruleMessage) errors[field.key] = ruleMessage
  }
  return errors
}

export function workflowFieldIsRequired(field: WorkflowFormField, values: WorkflowFormData): boolean {
  if (field.required) return true
  return (field.rules || []).some(rule => rule.type === 'conditional_required' && (!rule.when || workflowConditionMatches(rule.when, values)))
}

function validateWorkflowField(field: WorkflowFormField, value: unknown, values: WorkflowFormData, includeRules = true): string {
  const label = field.label || field.key
  if (field.required && isEmptyWorkflowValue(value)) return `${label}不能为空`
  if (isEmptyWorkflowValue(value)) {
    if (field.type === 'detail_list') {
      const rowCount = Array.isArray(value) ? value.length : 0
      if (field.minRows && rowCount < field.minRows) return `${label}至少需要${field.minRows}行`
      if (field.maxRows && rowCount > field.maxRows) return `${label}最多允许${field.maxRows}行`
    }
    return includeRules ? validateWorkflowRules(field, value, values) : ''
  }
  if (typeof value === 'string') {
    if (field.maxLength && Array.from(value).length > field.maxLength) return `${label}长度不能超过${field.maxLength}`
    if (field.type === 'phone' && !/^\+?[0-9][0-9 -]{5,19}$/.test(value.trim())) return `${label}格式无效`
    if (field.type === 'email' && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim())) return `${label}格式无效`
    if (field.type === 'time' && !/^([01]\d|2[0-3]):[0-5]\d(?::[0-5]\d)?$/.test(value.trim())) return `${label}格式无效`
  }
  if (field.type === 'number' || field.type === 'amount') {
    const number = workflowNumber(value)
    if (number === undefined) return `${label}类型无效`
    if (field.min !== undefined && number < field.min) return `${label}不能小于${field.min}`
    if (field.max !== undefined && number > field.max) return `${label}不能大于${field.max}`
  }
  if (field.type === 'date_range') {
    const range = Array.isArray(value) ? value.map(item => String(item)) : []
    if (range.length !== 2 || range[0] > range[1]) return `${label}日期区间无效`
  }
  if (field.type === 'detail_list') {
    if (!Array.isArray(value)) return `${label}类型无效`
    if (field.minRows && value.length < field.minRows) return `${label}至少需要${field.minRows}行`
    if (field.maxRows && value.length > field.maxRows) return `${label}最多允许${field.maxRows}行`
  }
  return includeRules ? validateWorkflowRules(field, value, values) : ''
}

function validateWorkflowRules(field: WorkflowFormField, value: unknown, values: WorkflowFormData): string {
  for (const rule of field.rules || []) {
    if (rule.when && !workflowConditionMatches(rule.when, values)) continue
    if (isEmptyWorkflowValue(value) && !['conditional_required', 'selection_count', 'column_sum'].includes(rule.type)) continue
    const message = validateWorkflowRule(field, value, rule, values)
    if (message) return message
  }
  return ''
}

function validateWorkflowRule(field: WorkflowFormField, value: unknown, rule: WorkflowValidationRule, values: WorkflowFormData): string {
  const label = field.label || field.key
  let failed = false
  let defaultMessage = `${label}校验不通过`
  switch (rule.type) {
    case 'conditional_required':
      failed = isEmptyWorkflowValue(value)
      defaultMessage = `${label}不能为空`
      break
    case 'min_length':
      failed = typeof value !== 'string' || rule.min === undefined || Array.from(value).length < rule.min
      defaultMessage = `${label}长度不能少于${rule.min ?? 0}`
      break
    case 'max_length':
      failed = typeof value !== 'string' || rule.max === undefined || Array.from(value).length > rule.max
      defaultMessage = `${label}长度不能超过${rule.max ?? 0}`
      break
    case 'pattern':
      try {
        failed = typeof value !== 'string' || !new RegExp(rule.pattern || '').test(value)
      } catch {
        failed = true
      }
      defaultMessage = `${label}格式不正确`
      break
    case 'number_range': {
      const number = workflowNumber(value)
      failed = number === undefined || (rule.min !== undefined && number < rule.min) || (rule.max !== undefined && number > rule.max)
      if (number !== undefined && rule.min !== undefined && number < rule.min) defaultMessage = `${label}不能小于${rule.min}`
      else if (number !== undefined && rule.max !== undefined && number > rule.max) defaultMessage = `${label}不能大于${rule.max}`
      break
    }
    case 'decimal_places': {
      const places = workflowDecimalPlaces(value)
      failed = places === undefined || rule.precision === undefined || places > rule.precision
      defaultMessage = `${label}小数位不能超过${rule.precision ?? 0}位`
      break
    }
    case 'selection_count': {
      const count = Array.isArray(value) ? value.length : -1
      failed = count < 0 || (rule.min !== undefined && count < rule.min) || (rule.max !== undefined && count > rule.max)
      if (rule.min !== undefined && count < rule.min) defaultMessage = `${label}至少选择${rule.min}项`
      else if (rule.max !== undefined && count > rule.max) defaultMessage = `${label}最多选择${rule.max}项`
      break
    }
    case 'compare_field': {
      const other = rule.field ? values[rule.field] : undefined
      if (isEmptyWorkflowValue(value) || isEmptyWorkflowValue(other)) return ''
      failed = !workflowOperatorMatches(value, other, rule.operator || 'eq')
      defaultMessage = `${label}与${rule.field || '目标字段'}的关系不符合要求`
      break
    }
    case 'column_sum': {
      const column = (field.columns || []).find(item => item.key === rule.column)
      const target = workflowNumber(rule.value)
      let sum = 0
      failed = !Array.isArray(value) || !column || target === undefined
      if (!failed) {
        for (const row of value as Array<Record<string, unknown>>) {
          const item = row?.[column!.key]
          if (isEmptyWorkflowValue(item)) continue
          const number = workflowNumber(item)
          if (number === undefined) {
            failed = true
            break
          }
          sum += number
        }
      }
      if (!failed) failed = !workflowNumberOperatorMatches(sum, target!, rule.operator || 'eq')
      defaultMessage = `${label}“${column?.label || rule.column || '目标列'}”列合计必须${workflowOperatorLabel(rule.operator || 'eq')}${rule.value ?? 0}`
      break
    }
  }
  return failed ? rule.message?.trim() || defaultMessage : ''
}

function workflowConditionMatches(condition: WorkflowValidationCondition, values: WorkflowFormData): boolean {
  const value = values[condition.field]
  if (condition.operator === 'empty') return isEmptyWorkflowValue(value)
  if (condition.operator === 'not_empty') return !isEmptyWorkflowValue(value)
  return workflowOperatorMatches(value, condition.value, condition.operator)
}

function workflowOperatorMatches(left: unknown, right: unknown, operator: WorkflowValidationOperator): boolean {
  const comparison = compareWorkflowValues(left, right)
  if (comparison === undefined) return operator === 'ne'
  if (operator === 'eq') return comparison === 0
  if (operator === 'ne') return comparison !== 0
  if (operator === 'gt') return comparison > 0
  if (operator === 'gte') return comparison >= 0
  if (operator === 'lt') return comparison < 0
  if (operator === 'lte') return comparison <= 0
  return false
}

function workflowNumberOperatorMatches(left: number, right: number, operator: WorkflowValidationOperator): boolean {
  if (!Number.isFinite(left) || !Number.isFinite(right)) return false
  const tolerance = 1e-9 * Math.max(1, Math.abs(left), Math.abs(right))
  const equal = Math.abs(left - right) <= tolerance
  if (operator === 'eq') return equal
  if (operator === 'ne') return !equal
  if (operator === 'gt') return left > right && !equal
  if (operator === 'gte') return left > right || equal
  if (operator === 'lt') return left < right && !equal
  if (operator === 'lte') return left < right || equal
  return false
}

function workflowOperatorLabel(operator: WorkflowValidationOperator): string {
  if (operator === 'eq') return '等于'
  if (operator === 'ne') return '不等于'
  if (operator === 'gt') return '大于'
  if (operator === 'gte') return '大于等于'
  if (operator === 'lt') return '小于'
  if (operator === 'lte') return '小于等于'
  return ''
}

function compareWorkflowValues(left: unknown, right: unknown): number | undefined {
  const leftNumber = workflowNumber(left)
  const rightNumber = workflowNumber(right)
  if (leftNumber !== undefined && rightNumber !== undefined) return Math.sign(leftNumber - rightNumber)
  if (typeof left === 'string' && typeof right === 'string') return left === right ? 0 : left < right ? -1 : 1
  if (typeof left === 'boolean' && typeof right === 'boolean') return left === right ? 0 : left ? 1 : -1
  return undefined
}

function workflowNumber(value: unknown): number | undefined {
  if (typeof value === 'number') return Number.isFinite(value) ? value : undefined
  if (typeof value !== 'string' || value.trim() === '') return undefined
  const number = Number(value)
  return Number.isFinite(number) ? number : undefined
}

function workflowDecimalPlaces(value: unknown): number | undefined {
  const number = workflowNumber(value)
  if (number === undefined) return undefined
  const text = number.toLocaleString('en-US', { useGrouping: false, maximumFractionDigits: 20 })
  return text.includes('.') ? text.split('.')[1].replace(/0+$/, '').length : 0
}

function isEmptyWorkflowValue(value: unknown): boolean {
  if (value === undefined || value === null) return true
  if (typeof value === 'string') return value.trim() === ''
  if (Array.isArray(value)) return value.length === 0
  return false
}
