import type {
  WorkflowFieldAccessMap,
  WorkflowFormData,
  WorkflowFormField,
} from '@/types/workflow'
import {
  cloneWorkflowValue,
  createWorkflowDetailRow,
  flattenWorkflowOptions,
  normalizeWorkflowFormValue,
  workflowDataFields,
} from './workflow-form'

export interface WorkflowHistoryImportMapping {
  sourceKey: string
  targetKey: string
}

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
): boolean {
  if (!sourceField || !targetField || excludedTypes.has(sourceField.type) || excludedTypes.has(targetField.type))
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
): NormalizedImportValue {
  if (!Array.isArray(value))
    return { value: [], result: 'skipped', message: '历史明细格式无效' }
  const sourceColumns = new Map((sourceField.columns || []).map(column => [column.key, column]))
  let filtered = false
  let dynamicOption = false
  const rows = value
    .filter(row => row && typeof row === 'object' && !Array.isArray(row))
    .map((row) => {
      const sourceRow = row as Record<string, unknown>
      const targetRow = createWorkflowDetailRow(targetField)
      for (const targetColumn of targetField.columns || []) {
        const sourceColumn = sourceColumns.get(targetColumn.key)
        if (!sourceColumn)
          continue
        const normalized = normalizeImportedValue(sourceColumn, targetColumn, sourceRow[sourceColumn.key])
        if (normalized.result === 'skipped') {
          filtered = true
          continue
        }
        if (normalized.result === 'filtered')
          filtered = true
        if (normalized.result === 'dynamic_option')
          dynamicOption = true
        targetRow[targetColumn.key] = cloneWorkflowValue(normalized.value)
      }
      return targetRow
    })
  if (filtered)
    return { value: rows, result: 'filtered', message: '部分明细选项或无效值已过滤' }
  if (dynamicOption)
    return { value: rows, result: 'dynamic_option', message: '明细包含动态选项，请导入后确认' }
  return { value: rows }
}

function normalizeImportedValue(
  sourceField: WorkflowFormField,
  targetField: WorkflowFormField,
  value: unknown,
): NormalizedImportValue {
  if (targetField.type === 'detail_list')
    return normalizeDetailListValue(sourceField, targetField, value)

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
  targetFields: WorkflowFormField[]
  targetData: WorkflowFormData
  targetAccess: WorkflowFieldAccessMap
  mappings: WorkflowHistoryImportMapping[]
  overwrite: boolean
}): WorkflowHistoryImportPreview {
  validateMappings(input.mappings)
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
    const sourceValue = sourceField ? input.sourceData[sourceField.key] : undefined
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
    if (!isWorkflowImportTypeCompatible(sourceField, targetField)) {
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

    const normalized = normalizeImportedValue(sourceField, targetField, sourceValue)
    if (normalized.result === 'skipped') {
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
    if (replacing)
      preview.overwriteCount += 1
    else
      preview.importCount += 1
  }
  return preview
}
