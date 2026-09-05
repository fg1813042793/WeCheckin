import type {
  WorkflowAttachment,
  WorkflowFieldAccess,
  WorkflowFieldAccessMap,
  WorkflowFieldActionsMap,
  WorkflowFieldPermission,
  WorkflowFormData,
  WorkflowFormField,
  WorkflowFormOption,
  WorkflowOptionSource,
} from '@/types/workflow'
import { calculateWorkflowFormData } from './workflow-calculation'

const arrayFieldTypes = new Set([
  'multi_select',
  'checkbox',
  'date_range',
  'user_multi',
  'department_multi',
  'attachment',
  'detail_list',
])

const layoutFieldTypes = new Set(['label', 'description', 'button'])

export function workflowDataFields(fields: WorkflowFormField[]): WorkflowFormField[] {
  const result: WorkflowFormField[] = []
  for (const field of fields || []) {
    if (!field?.key)
      continue
    if (field.type === 'group') {
      result.push(...workflowDataFields(field.fields || []))
    }
    else if (!layoutFieldTypes.has(field.type)) {
      result.push(field)
    }
  }
  return result
}

export function cloneWorkflowValue(value: unknown): unknown {
  if (Array.isArray(value))
    return value.map(item => cloneWorkflowValue(item))
  if (value && typeof value === 'object')
    return JSON.parse(JSON.stringify(value))
  return value
}

export function emptyWorkflowFieldValue(field: Pick<WorkflowFormField, 'type'>): unknown {
  if (arrayFieldTypes.has(field.type))
    return []
  if (field.type === 'boolean')
    return false
  if (field.type === 'number' || field.type === 'amount' || field.type === 'calculation')
    return undefined
  return ''
}

function generatedRowId() {
  return `row_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
}

export function workflowDetailRowKey(field: Pick<WorkflowFormField, 'rowKey'>) {
  return String(field.rowKey || '').trim() || 'id'
}

export function createWorkflowDetailRow(field: WorkflowFormField): Record<string, unknown> {
  const row: Record<string, unknown> = {
    [workflowDetailRowKey(field)]: generatedRowId(),
  }
  for (const column of field.columns || []) {
    row[column.key] = emptyWorkflowFieldValue(column)
  }
  return row
}

export function normalizeWorkflowFormValue(field: WorkflowFormField, value: unknown): unknown {
  if (field.type === 'detail_list') {
    if (!Array.isArray(value))
      return []
    const rowKey = workflowDetailRowKey(field)
    return value
      .filter(item => item && typeof item === 'object' && !Array.isArray(item))
      .map((item) => {
        const source = item as Record<string, unknown>
        const row: Record<string, unknown> = {
          [rowKey]: String(source[rowKey] || '').trim() || generatedRowId(),
        }
        for (const column of field.columns || []) {
          const raw = Object.prototype.hasOwnProperty.call(source, column.key)
            ? source[column.key]
            : emptyWorkflowFieldValue(column)
          row[column.key] = normalizeWorkflowFormValue(column, raw)
        }
        return row
      })
  }
  if (field.type === 'attachment')
    return normalizeWorkflowAttachments(value)
  if (!arrayFieldTypes.has(field.type))
    return value
  if (Array.isArray(value))
    return value.map(item => String(item)).filter(Boolean)
  if (typeof value === 'string') {
    return value.split(/\r?\n/).map(item => item.trim()).filter(Boolean)
  }
  return []
}

export function normalizeWorkflowAttachments(value: unknown): WorkflowAttachment[] {
  const source = Array.isArray(value)
    ? value
    : typeof value === 'string'
      ? value.split(/\r?\n/)
      : []
  const result: WorkflowAttachment[] = []
  for (const item of source) {
    if (typeof item === 'string') {
      const url = item.trim()
      if (!url)
        continue
      result.push({ id: url, name: workflowAttachmentName(url), url, mimeType: '', size: 0 })
      continue
    }
    if (!item || typeof item !== 'object' || Array.isArray(item))
      continue
    const record = item as Record<string, unknown>
    const url = String(record.url || '').trim()
    if (!url)
      continue
    const size = Number(record.size)
    result.push({
      id: String(record.id || url).trim() || url,
      name: String(record.name || '').trim() || workflowAttachmentName(url),
      url,
      mimeType: String(record.mimeType || '').trim(),
      size: Number.isFinite(size) && size >= 0 ? size : 0,
    })
  }
  return result
}

function workflowAttachmentName(url: string) {
  const path = url.split(/[?#]/, 1)[0]
  const rawName = path.slice(path.lastIndexOf('/') + 1) || '附件'
  try {
    return decodeURIComponent(rawName)
  }
  catch {
    return rawName
  }
}

export function initialWorkflowFormData(
  fields: WorkflowFormField[],
  existing: WorkflowFormData = {},
): WorkflowFormData {
  const result: WorkflowFormData = {}
  for (const field of workflowDataFields(fields)) {
    if (Object.prototype.hasOwnProperty.call(existing, field.key)) {
      result[field.key] = normalizeWorkflowFormValue(field, cloneWorkflowValue(existing[field.key]))
    }
    else if (Object.prototype.hasOwnProperty.call(field, 'default')) {
      result[field.key] = normalizeWorkflowFormValue(field, cloneWorkflowValue(field.default))
    }
    else if (field.type === 'detail_list' && Number(field.minRows || 0) > 0) {
      result[field.key] = Array.from({ length: Number(field.minRows) }, () => createWorkflowDetailRow(field))
    }
    else {
      result[field.key] = emptyWorkflowFieldValue(field)
    }
  }
  return calculateWorkflowFormData(fields, result)
}

export function workflowFieldAccessMap(
  fields: WorkflowFormField[],
  permissions: WorkflowFieldPermission[] | undefined,
  defaultAccess: WorkflowFieldAccess,
): WorkflowFieldAccessMap {
  const result: WorkflowFieldAccessMap = {}
  const dataFields = workflowDataFields(fields)
  const fieldTypes = new Map(dataFields.map(field => [field.key, field.type]))
  for (const field of dataFields) result[field.key] = field.type === 'calculation' ? 'read' : defaultAccess
  for (const permission of permissions || []) {
    if (permission.field in result && ['hidden', 'read', 'write'].includes(permission.access)
      && (fieldTypes.get(permission.field) !== 'calculation' || permission.access !== 'write')) {
      result[permission.field] = permission.access
    }
  }
  return result
}

export function workflowFieldActionsMap(
  fields: WorkflowFormField[],
  permissions: WorkflowFieldPermission[] | undefined,
  defaults = { add: false, delete: false },
): WorkflowFieldActionsMap {
  const result: WorkflowFieldActionsMap = {}
  for (const field of workflowDataFields(fields)) {
    if (field.type === 'detail_list') {
      result[field.key] = [
        ...(defaults.add ? ['add' as const] : []),
        ...(defaults.delete ? ['delete' as const] : []),
      ]
    }
  }
  for (const permission of permissions || []) {
    if (permission.field in result && permission.access === 'write') {
      if (!Array.isArray(permission.actions))
        continue
      result[permission.field] = (permission.actions || []).filter(action => action === 'add' || action === 'delete')
    }
  }
  return result
}

export function visibleWorkflowFormFields(
  fields: WorkflowFormField[],
  accessMap: WorkflowFieldAccessMap,
): WorkflowFormField[] {
  const result: WorkflowFormField[] = []
  for (const field of fields || []) {
    if (field.type === 'group') {
      const children = visibleWorkflowFormFields(field.fields || [], accessMap)
      if (children.length > 0)
        result.push({ ...field, fields: children })
    }
    else if (layoutFieldTypes.has(field.type) || accessMap[field.key] !== 'hidden') {
      result.push(field)
    }
  }
  return result
}

export function writableWorkflowFormData(
  fields: WorkflowFormField[],
  values: WorkflowFormData,
  accessMap: WorkflowFieldAccessMap,
): WorkflowFormData {
  const result: WorkflowFormData = {}
  for (const field of workflowDataFields(fields)) {
    if (field.type === 'calculation' || accessMap[field.key] !== 'write')
      continue
    const value = normalizeWorkflowFormValue(field, values[field.key])
    if (value !== undefined)
      result[field.key] = cloneWorkflowValue(value)
  }
  return result
}

export function normalizeWorkflowOptions(
  options: unknown,
  source?: Partial<WorkflowOptionSource>,
): WorkflowFormOption[] {
  if (!Array.isArray(options))
    return []
  const labelField = String(source?.labelField || '').trim() || 'label'
  const valueField = String(source?.valueField || '').trim() || 'value'
  const childrenField = String(source?.childrenField || '').trim() || 'children'
  const result: WorkflowFormOption[] = []
  for (const item of options) {
    if (!item || typeof item !== 'object' || Array.isArray(item))
      continue
    const record = item as Record<string, unknown>
    const label = String(optionPathValue(record, labelField) || '').trim()
    const value = String(optionPathValue(record, valueField) || '').trim()
    if (!label || !value)
      continue
    const children = normalizeWorkflowOptions(optionPathValue(record, childrenField), source)
    result.push(children.length > 0 ? { label, value, children } : { label, value })
  }
  return result
}

export function flattenWorkflowOptions(options: WorkflowFormOption[] = []): WorkflowFormOption[] {
  const result: WorkflowFormOption[] = []
  for (const option of options) {
    result.push({ label: option.label, value: option.value })
    if (option.children?.length)
      result.push(...flattenWorkflowOptions(option.children))
  }
  return result
}

export function optionPathValue(record: Record<string, unknown>, path: string): unknown {
  let current: unknown = record
  for (const segment of path.split('.').map(item => item.trim()).filter(Boolean)) {
    if (!current || typeof current !== 'object' || Array.isArray(current))
      return undefined
    current = (current as Record<string, unknown>)[segment]
  }
  return current
}

export function workflowOptionResponsePayload(response: unknown, source: WorkflowOptionSource) {
  const path = String(source.responsePath || '').trim() || 'data'
  if (response && typeof response === 'object' && !Array.isArray(response)) {
    const record = response as Record<string, unknown>
    const direct = optionPathValue(record, path)
    if (direct !== undefined)
      return direct
    if (record.data && typeof record.data === 'object' && !Array.isArray(record.data)) {
      const nested = optionPathValue(record.data as Record<string, unknown>, path)
      if (nested !== undefined)
        return nested
    }
    if (record.data !== undefined)
      return record.data
  }
  return response
}

export function validWorkflowOptionUrl(value?: string) {
  const url = String(value || '').trim()
  return url.startsWith('/api/') && !url.startsWith('//') && !url.includes('://') && !/\s/.test(url)
}

export function validateWorkflowFormData(
  fields: WorkflowFormField[],
  values: WorkflowFormData,
  accessMap?: WorkflowFieldAccessMap,
) {
  const errors: Record<string, string> = {}
  for (const field of workflowDataFields(fields)) {
    if (accessMap && accessMap[field.key] !== 'write')
      continue
    const message = validateWorkflowField(field, values[field.key], values)
    if (message) {
      errors[field.key] = message
      continue
    }
    if (field.type !== 'detail_list' || !Array.isArray(values[field.key]))
      continue
    const rows = values[field.key] as Array<Record<string, unknown>>
    for (let index = 0; index < rows.length; index += 1) {
      for (const column of field.columns || []) {
        const cellError = validateWorkflowField(column, rows[index]?.[column.key], rows[index] || {})
        if (cellError) {
          errors[field.key] = `${field.label || field.key}第${index + 1}行：${cellError}`
          break
        }
      }
      if (errors[field.key])
        break
    }
  }
  return errors
}

function validateWorkflowField(field: WorkflowFormField, value: unknown, values: WorkflowFormData): string {
  const label = field.label || field.key
  if (field.required && isEmptyWorkflowValue(value))
    return `${label}不能为空`
  if (isEmptyWorkflowValue(value)) {
    const conditional = (field.rules || []).find(rule =>
      rule.type === 'conditional_required' && (!rule.when || workflowConditionMatches(rule.when, values)),
    )
    return conditional ? String(conditional.message || '').trim() || `${label}不能为空` : ''
  }
  if (typeof value === 'string') {
    if (field.maxLength && Array.from(value).length > field.maxLength)
      return `${label}长度不能超过${field.maxLength}`
    if (field.type === 'phone' && !/^\+?\d[0-9 -]{5,19}$/.test(value.trim()))
      return `${label}格式无效`
    if (field.type === 'email' && !/^[^\s@]+@[^\s@][^\s.@]*\.[^\s@]+$/.test(value.trim()))
      return `${label}格式无效`
  }
  if (field.type === 'number' || field.type === 'amount') {
    const number = Number(value)
    if (!Number.isFinite(number))
      return `${label}类型无效`
    if (field.min !== undefined && number < field.min)
      return `${label}不能小于${field.min}`
    if (field.max !== undefined && number > field.max)
      return `${label}不能大于${field.max}`
  }
  if (field.type === 'date_range') {
    const range = Array.isArray(value) ? value.map(item => String(item)) : []
    if (range.length !== 2 || range[0] > range[1])
      return `${label}日期区间无效`
  }
  if (field.type === 'detail_list') {
    if (!Array.isArray(value))
      return `${label}类型无效`
    if (field.minRows && value.length < field.minRows)
      return `${label}至少需要${field.minRows}行`
    if (field.maxRows && value.length > field.maxRows)
      return `${label}最多允许${field.maxRows}行`
  }
  for (const rule of field.rules || []) {
    if (rule.when && !workflowConditionMatches(rule.when, values))
      continue
    if (rule.type === 'min_length' && typeof value === 'string' && rule.min !== undefined && Array.from(value).length < rule.min) {
      return String(rule.message || '').trim() || `${label}长度不能少于${rule.min}`
    }
    if (rule.type === 'max_length' && typeof value === 'string' && rule.max !== undefined && Array.from(value).length > rule.max) {
      return String(rule.message || '').trim() || `${label}长度不能超过${rule.max}`
    }
    if (rule.type === 'pattern' && typeof value === 'string') {
      try {
        if (!new RegExp(rule.pattern || '').test(value))
          return String(rule.message || '').trim() || `${label}格式不正确`
      }
      catch {
        return String(rule.message || '').trim() || `${label}格式不正确`
      }
    }
    if (rule.type === 'selection_count' && Array.isArray(value)) {
      if (rule.min !== undefined && value.length < rule.min)
        return String(rule.message || '').trim() || `${label}至少选择${rule.min}项`
      if (rule.max !== undefined && value.length > rule.max)
        return String(rule.message || '').trim() || `${label}最多选择${rule.max}项`
    }
    if (rule.type === 'compare_field' && rule.field && !workflowOperatorMatches(value, values[rule.field], rule.operator || 'eq')) {
      return String(rule.message || '').trim() || `${label}与${rule.field}的关系不符合要求`
    }
  }
  return ''
}

function workflowConditionMatches(condition: { field: string, operator: string, value?: unknown }, values: WorkflowFormData) {
  const value = values[condition.field]
  if (condition.operator === 'empty')
    return isEmptyWorkflowValue(value)
  if (condition.operator === 'not_empty')
    return !isEmptyWorkflowValue(value)
  return workflowOperatorMatches(value, condition.value, condition.operator)
}

function workflowOperatorMatches(left: unknown, right: unknown, operator: string) {
  const leftNumber = typeof left === 'number' ? left : Number.NaN
  const rightNumber = typeof right === 'number' ? right : Number.NaN
  let comparison: number | undefined
  if (Number.isFinite(leftNumber) && Number.isFinite(rightNumber))
    comparison = Math.sign(leftNumber - rightNumber)
  else if (typeof left === 'string' && typeof right === 'string')
    comparison = left === right ? 0 : left < right ? -1 : 1
  else if (typeof left === 'boolean' && typeof right === 'boolean')
    comparison = left === right ? 0 : left ? 1 : -1
  if (comparison === undefined)
    return operator === 'ne'
  if (operator === 'eq')
    return comparison === 0
  if (operator === 'ne')
    return comparison !== 0
  if (operator === 'gt')
    return comparison > 0
  if (operator === 'gte')
    return comparison >= 0
  if (operator === 'lt')
    return comparison < 0
  if (operator === 'lte')
    return comparison <= 0
  return false
}

function isEmptyWorkflowValue(value: unknown) {
  return value === undefined || value === null || (typeof value === 'string' && value.trim() === '') || (Array.isArray(value) && value.length === 0)
}
