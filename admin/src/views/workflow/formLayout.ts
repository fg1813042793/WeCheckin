import type { WorkflowFormField } from './types'

const layoutFieldTypes = new Set<WorkflowFormField['type']>([
  'group',
  'label',
  'description',
  'button',
])

export function isWorkflowLayoutField(field: Pick<WorkflowFormField, 'type'>): boolean {
  return layoutFieldTypes.has(field.type)
}

export function isWorkflowDataField(field: Pick<WorkflowFormField, 'type'>): boolean {
  return !isWorkflowLayoutField(field)
}

export interface WorkflowDataFieldEntry {
  field: WorkflowFormField
  group?: WorkflowFormField
}

export function workflowDataFieldEntries(fields: WorkflowFormField[]): WorkflowDataFieldEntry[] {
  const result: WorkflowDataFieldEntry[] = []

  function append(source: WorkflowFormField[], group?: WorkflowFormField) {
    for (const field of source) {
      if (field.type === 'group') {
        append(field.fields || [], field)
        continue
      }
      if (isWorkflowDataField(field)) result.push({ field, group })
    }
  }

  append(fields)
  return result
}

export function workflowDataFields(fields: WorkflowFormField[]): WorkflowFormField[] {
  return workflowDataFieldEntries(fields).map(entry => entry.field)
}

export function workflowFieldByKey(fields: WorkflowFormField[], key: string): WorkflowFormField | undefined {
  for (const field of fields) {
    if (field.key === key) return field
    if (field.type === 'group') {
      const nested = workflowFieldByKey(field.fields || [], key)
      if (nested) return nested
    }
  }
  return undefined
}

export function removeWorkflowField(fields: WorkflowFormField[], key: string): WorkflowFormField | undefined {
  const directIndex = fields.findIndex((field) => field.key === key)
  if (directIndex >= 0) return fields.splice(directIndex, 1)[0]
  for (const field of fields) {
    if (field.type !== 'group') continue
    const removed = removeWorkflowField(field.fields || [], key)
    if (removed) return removed
  }
  return undefined
}

interface WorkflowFieldLocation {
  fields: WorkflowFormField[]
  index: number
  field: WorkflowFormField
}

function workflowFieldLocation(fields: WorkflowFormField[], key: string): WorkflowFieldLocation | undefined {
  for (let index = 0; index < fields.length; index += 1) {
    const field = fields[index]
    if (field.key === key) return { fields, index, field }
    if (field.type === 'group') {
      const nested = workflowFieldLocation(field.fields || [], key)
      if (nested) return nested
    }
  }
  return undefined
}

export function insertWorkflowField(
  fields: WorkflowFormField[],
  field: WorkflowFormField,
  targetGroupKey: string | null,
  targetIndex: number,
): boolean {
  if (!field.key || workflowFieldByKey(fields, field.key)) return false

  let targetFields = fields
  if (targetGroupKey !== null) {
    const targetGroup = workflowFieldByKey(fields, targetGroupKey)
    if (!targetGroup || targetGroup.type !== 'group' || field.type === 'group') return false
    targetGroup.fields = Array.isArray(targetGroup.fields) ? targetGroup.fields : []
    targetFields = targetGroup.fields
  }

  const insertionIndex = Number.isFinite(targetIndex)
    ? Math.max(0, Math.min(Math.trunc(targetIndex), targetFields.length))
    : targetFields.length
  targetFields.splice(insertionIndex, 0, field)
  return true
}

export function moveWorkflowField(
  fields: WorkflowFormField[],
  sourceKey: string,
  targetGroupKey: string | null,
  targetIndex: number,
): boolean {
  const source = workflowFieldLocation(fields, sourceKey)
  if (!source) return false

  let targetFields = fields
  if (targetGroupKey !== null) {
    const targetGroup = workflowFieldByKey(fields, targetGroupKey)
    if (!targetGroup || targetGroup.type !== 'group' || source.field.type === 'group') return false
    targetGroup.fields = Array.isArray(targetGroup.fields) ? targetGroup.fields : []
    targetFields = targetGroup.fields
  }

  let insertionIndex = Number.isFinite(targetIndex) ? Math.trunc(targetIndex) : targetFields.length
  insertionIndex = Math.max(0, Math.min(insertionIndex, targetFields.length))
  source.fields.splice(source.index, 1)
  if (source.fields === targetFields && source.index < insertionIndex) insertionIndex -= 1
  insertionIndex = Math.max(0, Math.min(insertionIndex, targetFields.length))
  targetFields.splice(insertionIndex, 0, source.field)
  return true
}

export function moveWorkflowDetailColumn(
  columns: WorkflowFormField[],
  sourceIndex: number,
  targetIndex: number,
): boolean {
  if (!Number.isInteger(sourceIndex) || sourceIndex < 0 || sourceIndex >= columns.length) return false
  if (!Number.isInteger(targetIndex) || targetIndex < 0 || targetIndex > columns.length) return false
  if (targetIndex === sourceIndex || targetIndex === sourceIndex + 1) return false

  const [column] = columns.splice(sourceIndex, 1)
  const insertionIndex = targetIndex > sourceIndex ? targetIndex - 1 : targetIndex
  columns.splice(insertionIndex, 0, column)
  return true
}
