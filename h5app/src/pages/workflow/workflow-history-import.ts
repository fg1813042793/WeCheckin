import type {
  WorkflowFieldAccessMap,
  WorkflowFormData,
  WorkflowFormField,
  WorkflowHistoryImportColumnMapping,
  WorkflowHistoryImportMapping,
} from '@/types/workflow'
import {
  cloneWorkflowValue,
  createWorkflowDetailRow,
  flattenWorkflowOptions,
  normalizeWorkflowFormValue,
  workflowDataFields,
  workflowDetailRowKey,
} from './workflow-form'

export type { WorkflowHistoryImportMapping } from '@/types/workflow'

export type WorkflowHistoryImportResult
  = 'import'
    | 'overwrite'
    | 'retain'
    | 'incompatible'
    | 'filtered'
    | 'dynamic_option'
    | 'skipped'

export interface WorkflowHistoryImportPreviewItem {
  sourceKey: string
  sourceLabel: string
  targetKey: string
  targetLabel: string
  currentValue: unknown
  importValue: unknown
  result: WorkflowHistoryImportResult
  message: string
  filteredCount?: number
  dynamicCount?: number
}

export interface WorkflowHistoryImportPreview {
  items: WorkflowHistoryImportPreviewItem[]
  patch: WorkflowFormData
  importCount: number
  overwriteCount: number
  retainCount: number
  skipCount: number
}

interface NormalizedImportValue {
  value: unknown
  result?: Extract<WorkflowHistoryImportResult, 'filtered' | 'dynamic_option' | 'skipped'>
  message?: string
  filteredCount?: number
  dynamicCount?: number
}

const stringTypes = new Set(['text', 'textarea', 'phone', 'email'])
const numberTypes = new Set(['number', 'amount'])
const singleOptionTypes = new Set(['select', 'radio'])
const multiOptionTypes = new Set(['multi_select', 'checkbox'])
const excludedTypes = new Set(['calculation', 'group', 'label', 'description', 'button'])

function inSameTypeFamily(sourceType: string, targetType: string, family: Set<string>) {
  return family.has(sourceType) && family.has(targetType)
}

export function isWorkflowImportTypeCompatible(
  sourceField: WorkflowFormField,
  targetField: WorkflowFormField,
  columns?: WorkflowHistoryImportColumnMapping[],
): boolean {
  if (!sourceField || !targetField || excludedTypes.has(sourceField.type) || excludedTypes.has(targetField.type))
    return false
  if (sourceField.calculation || targetField.calculation)
    return false
  if (columns && (sourceField.type !== 'detail_list' || targetField.type !== 'detail_list'))
    return false
  if (inSameTypeFamily(sourceField.type, targetField.type, stringTypes))
    return true
  if (inSameTypeFamily(sourceField.type, targetField.type, numberTypes))
    return true
  if (inSameTypeFamily(sourceField.type, targetField.type, singleOptionTypes))
    return true
  if (inSameTypeFamily(sourceField.type, targetField.type, multiOptionTypes))
    return true
  if (sourceField.type !== targetField.type)
    return false
  if (targetField.type !== 'detail_list')
    return true

  const sourceColumns = new Map((sourceField.columns || []).map(column => [column.key, column]))
  if (columns) {
    const targetColumns = new Map((targetField.columns || []).map(column => [column.key, column]))
    return columns.some((mapping) => {
      const source = sourceColumns.get(mapping.sourceKey)
      const target = targetColumns.get(mapping.targetKey)
      return Boolean(source && target && source.type !== 'detail_list' && target.type !== 'detail_list'
        && isWorkflowImportTypeCompatible(source, target))
    })
  }
  return (targetField.columns || []).every((targetColumn) => {
    const sourceColumn = sourceColumns.get(targetColumn.key)
    return Boolean(sourceColumn && isWorkflowImportTypeCompatible(sourceColumn, targetColumn))
  })
}

function writableTargetFields(fields: WorkflowFormField[], access: WorkflowFieldAccessMap) {
  return workflowDataFields(fields).filter(field =>
    field.type !== 'calculation' && access[field.key] === 'write',
  )
}

export function buildWorkflowImportMappings(
  sourceFields: WorkflowFormField[],
  targetFields: WorkflowFormField[],
  targetAccess: WorkflowFieldAccessMap,
): WorkflowHistoryImportMapping[] {
  const sourceByKey = new Map(workflowDataFields(sourceFields).map(field => [field.key, field]))
  return writableTargetFields(targetFields, targetAccess)
    .flatMap((targetField) => {
      const sourceField = sourceByKey.get(targetField.key)
      return sourceField && isWorkflowImportTypeCompatible(sourceField, targetField)
        ? [{ sourceKey: sourceField.key, targetKey: targetField.key }]
        : []
    })
}

export function suggestWorkflowImportMappings(
  sourceFields: WorkflowFormField[],
  targetFields: WorkflowFormField[],
  targetAccess: WorkflowFieldAccessMap,
): WorkflowHistoryImportMapping[] {
  const sourceDataFields = workflowDataFields(sourceFields)
  const targetDataFields = writableTargetFields(targetFields, targetAccess)
  const mappings = buildWorkflowImportMappings(sourceFields, targetFields, targetAccess)
  const usedSources = new Set(mappings.map(mapping => mapping.sourceKey))
  const usedTargets = new Set(mappings.map(mapping => mapping.targetKey))

  for (const sourceField of sourceDataFields) {
    if (usedSources.has(sourceField.key))
      continue
    const candidates = targetDataFields.filter(targetField =>
      !usedTargets.has(targetField.key)
      && sourceField.label.trim() !== ''
      && sourceField.label.trim() === targetField.label.trim()
      && isWorkflowImportTypeCompatible(sourceField, targetField),
    )
    if (candidates.length !== 1)
      continue
    mappings.push({ sourceKey: sourceField.key, targetKey: candidates[0].key })
    usedSources.add(sourceField.key)
    usedTargets.add(candidates[0].key)
  }
  return mappings
}

function isEmptyWorkflowImportValue(value: unknown) {
  return value === undefined
    || value === null
    || (typeof value === 'string' && value.trim() === '')
    || (Array.isArray(value) && value.length === 0)
}

function validateMappings(mappings: WorkflowHistoryImportMapping[]) {
  const sourceKeys = new Set<string>()
  const targetKeys = new Set<string>()
  for (const mapping of mappings) {
    if (sourceKeys.has(mapping.sourceKey))
      throw new Error(`源字段不能重复：${mapping.sourceKey}`)
    if (targetKeys.has(mapping.targetKey))
      throw new Error(`目标字段不能重复：${mapping.targetKey}`)
    sourceKeys.add(mapping.sourceKey)
    targetKeys.add(mapping.targetKey)
  }
}

function optionValues(field: WorkflowFormField) {
  return new Set(flattenWorkflowOptions(field.options || []).map(option => String(option.value)))
}

function normalizeDetailListValue(
  sourceField: WorkflowFormField,
  targetField: WorkflowFormField,
  value: unknown,
  columns?: WorkflowHistoryImportColumnMapping[],
): NormalizedImportValue {
  if (!Array.isArray(value))
    return { value: [], result: 'skipped', message: '历史明细格式无效' }
  const sourceColumns = new Map((sourceField.columns || []).map(column => [column.key, column]))
  const targetColumns = new Map((targetField.columns || []).map(column => [column.key, column]))
  const mappings = columns || (targetField.columns || []).map(column => ({ sourceKey: column.key, targetKey: column.key }))
  let filteredCount = 0
  let dynamicCount = 0
  const rows = value
    .filter(row => row && typeof row === 'object' && !Array.isArray(row))
    .map((row) => {
      const sourceRow = row as Record<string, unknown>
      const targetRow = createWorkflowDetailRow(targetField)
      let imported = false
      for (const mapping of mappings) {
        const targetColumn = targetColumns.get(mapping.targetKey)
        const sourceColumn = sourceColumns.get(mapping.sourceKey)
        if (!sourceColumn || !targetColumn
          || sourceColumn.key === workflowDetailRowKey(sourceField)
          || targetColumn.key === workflowDetailRowKey(targetField)
          || (columns && (sourceColumn.type === 'detail_list' || targetColumn.type === 'detail_list'))
          || !isWorkflowImportTypeCompatible(sourceColumn, targetColumn)) {
          filteredCount += 1
          continue
        }
        if (isEmptyWorkflowImportValue(sourceRow[sourceColumn.key]))
          continue
        const normalized = normalizeImportedValue(sourceColumn, targetColumn, sourceRow[sourceColumn.key])
        if (normalized.result === 'skipped' || isEmptyWorkflowImportValue(normalized.value)) {
          filteredCount += 1
          continue
        }
        if (normalized.result === 'filtered')
          filteredCount += normalized.filteredCount || 1
        if (normalized.result === 'dynamic_option')
          dynamicCount += normalized.dynamicCount || 1
        targetRow[targetColumn.key] = cloneWorkflowValue(normalized.value)
        imported = true
      }
      return imported ? targetRow : null
    })
    .filter(row => row !== null)
  if (rows.length === 0)
    return { value: [], result: 'skipped' }
  if (filteredCount > 0 || dynamicCount > 0) {
    return {
      value: rows,
      result: filteredCount > 0 ? 'filtered' : 'dynamic_option',
      message: filteredCount > 0 ? '部分明细选项或无效值已过滤' : '明细包含动态选项，请导入后确认',
      filteredCount,
      dynamicCount,
    }
  }
  return { value: rows }
}

function normalizeImportedValue(
  sourceField: WorkflowFormField,
  targetField: WorkflowFormField,
  value: unknown,
  columns?: WorkflowHistoryImportColumnMapping[],
): NormalizedImportValue {
  if (targetField.type === 'detail_list')
    return normalizeDetailListValue(sourceField, targetField, value, columns)

  const normalized = normalizeWorkflowFormValue(targetField, cloneWorkflowValue(value))
  if (singleOptionTypes.has(targetField.type)) {
    if (targetField.optionSource?.type === 'api')
      return { value: normalized, result: 'dynamic_option', message: '动态选项需在导入后确认' }
    const allowed = optionValues(targetField)
    if (!allowed.has(String(normalized)))
      return { value: undefined, result: 'skipped', message: '历史选项在当前表单中不存在' }
  }
  if (multiOptionTypes.has(targetField.type)) {
    if (targetField.optionSource?.type === 'api')
      return { value: normalized, result: 'dynamic_option', message: '动态选项需在导入后确认' }
    const values = Array.isArray(normalized) ? normalized.map(item => String(item)) : []
    const allowed = optionValues(targetField)
    const filtered = values.filter(item => allowed.has(item))
    if (filtered.length === 0 && values.length > 0)
      return { value: [], result: 'skipped', message: '历史选项在当前表单中均不存在' }
    if (filtered.length !== values.length)
      return { value: filtered, result: 'filtered', message: '部分历史选项已被当前表单过滤' }
    return { value: filtered }
  }
  return { value: normalized }
}

function previewItem(
  sourceField: WorkflowFormField | undefined,
  targetField: WorkflowFormField | undefined,
  currentValue: unknown,
  importValue: unknown,
  result: WorkflowHistoryImportResult,
  message: string,
): WorkflowHistoryImportPreviewItem {
  return {
    sourceKey: sourceField?.key || '',
    sourceLabel: sourceField?.label || sourceField?.key || '未知历史字段',
    targetKey: targetField?.key || '',
    targetLabel: targetField?.label || targetField?.key || '未知当前字段',
    currentValue: cloneWorkflowValue(currentValue),
    importValue: cloneWorkflowValue(importValue),
    result,
    message,
  }
}

export function previewWorkflowHistoryImport(input: {
  sourceFields: WorkflowFormField[]
  sourceData: WorkflowFormData
  sourceAccess?: WorkflowFieldAccessMap
  targetFields: WorkflowFormField[]
  targetData: WorkflowFormData
  targetAccess: WorkflowFieldAccessMap
  mappings: WorkflowHistoryImportMapping[]
  overwrite: boolean
}): WorkflowHistoryImportPreview {
  validateMappings(input.mappings)
  for (const mapping of input.mappings) {
    if (mapping.columns)
      validateMappings(mapping.columns)
  }
  const sourceByKey = new Map(workflowDataFields(input.sourceFields).map(field => [field.key, field]))
  const targetByKey = new Map(workflowDataFields(input.targetFields).map(field => [field.key, field]))
  const preview: WorkflowHistoryImportPreview = {
    items: [],
    patch: {},
    importCount: 0,
    overwriteCount: 0,
    retainCount: 0,
    skipCount: 0,
  }

  for (const mapping of input.mappings) {
    const sourceField = sourceByKey.get(mapping.sourceKey)
    const targetField = targetByKey.get(mapping.targetKey)
    const currentValue = targetField ? input.targetData[targetField.key] : undefined
    const sourceAllowed = !input.sourceAccess || input.sourceAccess[mapping.sourceKey] === 'write'
    const sourceValue = sourceField && sourceAllowed ? input.sourceData[sourceField.key] : undefined
    if (!sourceAllowed) {
      preview.items.push(previewItem(sourceField, targetField, currentValue, undefined, 'skipped', ''))
      preview.skipCount += 1
      continue
    }
    if (!sourceField || !targetField) {
      preview.items.push(previewItem(sourceField, targetField, currentValue, sourceValue, 'skipped', '字段已不存在'))
      preview.skipCount += 1
      continue
    }
    if (input.targetAccess[targetField.key] !== 'write' || targetField.type === 'calculation') {
      preview.items.push(previewItem(sourceField, targetField, currentValue, sourceValue, 'skipped', '当前字段不可写'))
      preview.skipCount += 1
      continue
    }
    if (!isWorkflowImportTypeCompatible(sourceField, targetField, mapping.columns)) {
      preview.items.push(previewItem(sourceField, targetField, currentValue, sourceValue, 'incompatible', '字段类型不兼容'))
      preview.skipCount += 1
      continue
    }
    if (isEmptyWorkflowImportValue(sourceValue)) {
      preview.items.push(previewItem(sourceField, targetField, currentValue, sourceValue, 'skipped', '历史字段没有可导入内容'))
      preview.skipCount += 1
      continue
    }
    if (!input.overwrite && !isEmptyWorkflowImportValue(currentValue)) {
      preview.items.push(previewItem(sourceField, targetField, currentValue, sourceValue, 'retain', '保留当前已有内容'))
      preview.retainCount += 1
      continue
    }

    const normalized = normalizeImportedValue(sourceField, targetField, sourceValue, mapping.columns)
    if (normalized.result === 'skipped' || isEmptyWorkflowImportValue(normalized.value)) {
      preview.items.push(previewItem(
        sourceField,
        targetField,
        currentValue,
        normalized.value,
        'skipped',
        normalized.message || '没有可导入内容',
      ))
      preview.skipCount += 1
      continue
    }
    const replacing = !isEmptyWorkflowImportValue(currentValue)
    const result = normalized.result || (replacing ? 'overwrite' : 'import')
    preview.items.push(previewItem(
      sourceField,
      targetField,
      currentValue,
      normalized.value,
      result,
      normalized.message || (replacing ? '将覆盖当前内容' : '将导入历史内容'),
    ))
    preview.patch[targetField.key] = cloneWorkflowValue(normalized.value)
    const item = preview.items[preview.items.length - 1]
    item.filteredCount = normalized.filteredCount || (result === 'filtered' ? 1 : 0)
    item.dynamicCount = normalized.dynamicCount || (result === 'dynamic_option' ? 1 : 0)
    if (replacing)
      preview.overwriteCount += 1
    else
      preview.importCount += 1
  }
  return preview
}
