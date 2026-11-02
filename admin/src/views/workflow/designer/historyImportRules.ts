import type { WorkflowDraft, WorkflowFormField, WorkflowHistoryImportConfig, WorkflowHistoryImportMapping } from '@/types/workflow'
import { workflowDataFieldEntries } from '../formLayout'

export interface HistoryImportFieldOption {
  value: string
  label: string
  disabled: boolean
}

const excludedTypes = new Set(['calculation', 'group', 'label', 'description', 'button'])
const families = [['text', 'textarea', 'phone', 'email'], ['number', 'amount'], ['select', 'radio'], ['multi_select', 'checkbox']]
const importableTypes = new Set([...families.flat(), 'date', 'datetime', 'time', 'date_range', 'boolean', 'user', 'user_multi', 'department', 'department_multi', 'attachment', 'detail_list'])

export function historyImportFieldReason(field: WorkflowFormField | undefined, column = false): string {
  if (!field) return '字段不存在'
  if (!field.key || field.key.length > 100) return '字段标识须为 1–100 字'
  if (excludedTypes.has(field.type) || field.calculation) return '计算或展示字段不可导入'
  if (!importableTypes.has(field.type)) return '不支持的字段类型'
  if (column && field.type === 'detail_list') return '不支持嵌套明细'
  return ''
}

export function historyImportTypesCompatible(source: WorkflowFormField | undefined, target: WorkflowFormField | undefined): boolean {
  if (!source || !target || historyImportFieldReason(source) || historyImportFieldReason(target)) return false
  return source.type === target.type || families.some(family => family.includes(source.type) && family.includes(target.type))
}

export function historyImportFields(draft: WorkflowDraft) {
  return workflowDataFieldEntries(draft.form).map(({ field, group }) => ({
    field,
    label: `${group ? `${group.label} / ` : ''}${field.label || '未命名字段'}`,
  }))
}

export function historyImportTargetReason(draft: WorkflowDraft, field: WorkflowFormField | undefined): string {
  const reason = historyImportFieldReason(field)
  if (reason || !field) return reason
  const access = draft.nodes.find(node => node.type === 'start')?.formPermissions?.find(item => item.field === field.key)?.access
  return access === 'hidden' ? '发起时隐藏' : access === 'read' ? '发起时只读' : ''
}

export function historyImportRuleErrors(config: Required<WorkflowHistoryImportConfig>, draft: WorkflowDraft): string[] {
  const errors: string[] = []
  // Disabled presets retain stale fields, while still reporting excessive payloads.
  if (config.mappings.length > 100 || config.mappings.some(mapping => (mapping.columns?.length || 0) > 100)) {
    errors.push('字段映射和每组明细列映射最多各 100 条')
  }
  if ([config.sourceMode, config.sourceStatus, config.mappingMode].some(value => value.length > 100)
    || config.mappings.some(mapping => [mapping, ...(mapping.columns || [])].some(pair => pair.sourceKey.length > 100 || pair.targetKey.length > 100))) {
    errors.push('规则选项和字段标识最多 100 字')
  }
  if (!config.enabled) return errors
  if (!['select_history', 'latest_history'].includes(config.sourceMode)) errors.push('请选择有效的来源方式')
  if (!['completed', 'all'].includes(config.sourceStatus)) errors.push('请选择有效的来源状态')
  if (!['same_key', 'mapping'].includes(config.mappingMode)) errors.push('请选择有效的字段方式')
  // Same-key imports retain the manual draft without applying its field/column rules.
  if (config.mappingMode === 'same_key') return errors
  if (config.mappingMode === 'mapping' && !config.mappings.length) errors.push('请至少添加一条字段映射')
  const fields = new Map(historyImportFields(draft).map(({ field }) => [field.key, field]))
  const seenSources = new Set<string>()
  const seenTargets = new Set<string>()
  config.mappings.forEach((mapping, index) => {
    const prefix = `第 ${index + 1} 条映射：`
    const source = fields.get(mapping.sourceKey)
    const target = fields.get(mapping.targetKey)
    const reason = historyImportFieldReason(source) || historyImportTargetReason(draft, target)
    if (reason) errors.push(prefix + reason)
    else if (!historyImportTypesCompatible(source, target)) errors.push(prefix + '字段类型不兼容')
    if (seenSources.has(mapping.sourceKey) || seenTargets.has(mapping.targetKey)) errors.push(prefix + '来源或目标字段重复')
    seenSources.add(mapping.sourceKey)
    seenTargets.add(mapping.targetKey)
    if (source?.type === 'detail_list' && target?.type === 'detail_list') errors.push(...columnErrors(mapping, source, target).map(error => prefix + error))
    else if (mapping.columns !== undefined) errors.push(prefix + '普通字段不能配置明细列映射')
  })
  return errors
}

function columnErrors(mapping: WorkflowHistoryImportMapping, source: WorkflowFormField, target: WorkflowFormField): string[] {
  const columns = mapping.columns ?? (target.columns || []).map(column => ({ sourceKey: column.key, targetKey: column.key }))
  if (mapping.columns !== undefined && (columns.length < 1 || columns.length > 100)) return ['自定义明细列映射须为 1–100 条；清除自定义映射可恢复同名列匹配']
  const errors: string[] = []
  const seenSources = new Set<string>()
  const seenTargets = new Set<string>()
  for (const column of columns) {
    const sourceColumn = source.columns?.find(item => item.key === column.sourceKey)
    const targetColumn = target.columns?.find(item => item.key === column.targetKey)
    const reason = historyImportFieldReason(sourceColumn, true) || historyImportFieldReason(targetColumn, true)
    if (reason) errors.push(`明细列${reason}${mapping.columns === undefined ? '，请配置自定义明细列映射' : ''}`)
    else if (!historyImportTypesCompatible(sourceColumn, targetColumn)) errors.push('明细列类型不兼容')
    if (seenSources.has(column.sourceKey) || seenTargets.has(column.targetKey)) errors.push('明细来源列或目标列重复')
    seenSources.add(column.sourceKey)
    seenTargets.add(column.targetKey)
  }
  return errors
}
