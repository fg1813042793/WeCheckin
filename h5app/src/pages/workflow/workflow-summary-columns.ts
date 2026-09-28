import type {
  WorkflowFormField,
  WorkflowInstanceSummary,
  WorkflowPublishedDefinition,
} from '@/types/workflow'
import { flattenWorkflowOptions } from './workflow-form'

const workflowLayoutFieldTypes = new Set(['label', 'description', 'button'])

export type WorkflowSummaryBaseColumnKey
  = | 'check'
    | 'title'
    | 'workflowName'
    | 'businessKey'
    | 'starter'
    | 'operator'
    | 'version'
    | 'status'
    | 'startTime'
    | 'endTime'
    | 'actions'

export interface WorkflowSummaryColumn {
  key: string
  label: string
  width: string
  kind: 'base' | 'form'
  baseKey?: WorkflowSummaryBaseColumnKey
  definitionId?: number
  field?: WorkflowFormField
  configurable?: boolean
  mobileHidden?: boolean
}

export interface WorkflowSummaryFormFieldGroup {
  definitionId: number
  label: string
  columns: WorkflowSummaryColumn[]
}

interface WorkflowSummaryDataField {
  field: WorkflowFormField
  groupPath: string[]
}

export const workflowSummaryBaseColumns: WorkflowSummaryColumn[] = [
  { key: 'check', label: '选择', width: '66px', kind: 'base', baseKey: 'check' },
  { key: 'title', label: '单据标题', width: 'minmax(180px, 1.3fr)', kind: 'base', baseKey: 'title' },
  { key: 'workflowName', label: '流程名称', width: 'minmax(140px, 1fr)', kind: 'base', baseKey: 'workflowName', configurable: true, mobileHidden: true },
  { key: 'businessKey', label: '申请编号', width: 'minmax(190px, 1.45fr)', kind: 'base', baseKey: 'businessKey' },
  { key: 'starter', label: '发起人', width: 'minmax(100px, 0.8fr)', kind: 'base', baseKey: 'starter' },
  { key: 'operator', label: '操作人', width: 'minmax(100px, 0.8fr)', kind: 'base', baseKey: 'operator', configurable: true, mobileHidden: true },
  { key: 'version', label: '版本', width: '76px', kind: 'base', baseKey: 'version', configurable: true, mobileHidden: true },
  { key: 'status', label: '状态', width: '94px', kind: 'base', baseKey: 'status' },
  { key: 'startTime', label: '发起时间', width: '168px', kind: 'base', baseKey: 'startTime' },
  { key: 'endTime', label: '完成时间', width: '168px', kind: 'base', baseKey: 'endTime', configurable: true, mobileHidden: true },
  { key: 'actions', label: '操作', width: '112px', kind: 'base', baseKey: 'actions' },
]

export function workflowSummaryFormFieldGroups(definitions: WorkflowPublishedDefinition[]): WorkflowSummaryFormFieldGroup[] {
  return definitions.map(definition => ({
    definitionId: definition.id,
    label: definition.displayName || definition.name || definition.key,
    columns: workflowSummaryDataFields(definition.form || []).map(({ field, groupPath }) => ({
      key: `form:${definition.id}:${field.key}`,
      label: [...groupPath, field.label || field.key].join('.'),
      width: 'minmax(130px, 1fr)',
      kind: 'form' as const,
      definitionId: definition.id,
      field,
      configurable: true,
      mobileHidden: true,
    })),
  })).filter(group => group.columns.length > 0)
}

function workflowSummaryDataFields(fields: WorkflowFormField[], groupPath: string[] = []): WorkflowSummaryDataField[] {
  const result: WorkflowSummaryDataField[] = []
  for (const field of fields || []) {
    if (!field?.key)
      continue
    if (field.type === 'group') {
      const nextGroupPath = [...groupPath, field.label || field.key]
      result.push(...workflowSummaryDataFields(field.fields || [], nextGroupPath))
    }
    else if (!workflowLayoutFieldTypes.has(field.type)) {
      result.push({ field, groupPath })
    }
  }
  return result
}

export function visibleWorkflowSummaryColumns(
  baseHiddenKeys: string[],
  selectedFormKeys: string[],
  formGroups: WorkflowSummaryFormFieldGroup[],
) {
  const hidden = new Set(baseHiddenKeys)
  const selected = new Set(selectedFormKeys)
  const base = workflowSummaryBaseColumns.filter(column => !column.configurable || !hidden.has(column.key))
  const actionIndex = base.findIndex(column => column.baseKey === 'actions')
  const formColumns = formGroups.flatMap(group => group.columns).filter(column => selected.has(column.key))
  if (actionIndex < 0)
    return [...base, ...formColumns]
  return [...base.slice(0, actionIndex), ...formColumns, ...base.slice(actionIndex)]
}

export function workflowSummaryFormValue(instance: WorkflowInstanceSummary, column: WorkflowSummaryColumn) {
  if (column.kind !== 'form' || instance.definitionId !== column.definitionId || !column.field)
    return '-'
  return formatWorkflowFormValue(instance.formData?.[column.field.key], column.field)
}

function formatWorkflowFormValue(value: unknown, field: WorkflowFormField): string {
  if (value === undefined || value === null || value === '')
    return '-'
  if (field.type === 'boolean')
    return value ? '是' : '否'
  if (field.type === 'select' || field.type === 'radio')
    return optionLabel(field, value)
  if (field.type === 'multi_select' || field.type === 'checkbox') {
    const values = Array.isArray(value) ? value : [value]
    return values.map(item => optionLabel(field, item)).join('、') || '-'
  }
  if (field.type === 'date_range' && Array.isArray(value))
    return value.map(item => String(item)).join(' 至 ') || '-'
  if (field.type === 'attachment' && Array.isArray(value)) {
    return value.map((item) => {
      if (item && typeof item === 'object' && !Array.isArray(item)) {
        const attachment = item as Record<string, unknown>
        return String(attachment.name || attachment.url || '')
      }
      return String(item || '')
    }).filter(Boolean).join('、') || '-'
  }
  if (field.type === 'detail_list' && Array.isArray(value))
    return value.length > 0 ? `共 ${value.length} 项` : '-'
  if (Array.isArray(value))
    return value.map(item => String(item)).join('、') || '-'
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value)
    }
    catch {
      return '-'
    }
  }
  return String(value)
}

function optionLabel(field: WorkflowFormField, value: unknown) {
  const text = String(value)
  return flattenWorkflowOptions(field.options || []).find(option => String(option.value) === text)?.label || text
}
