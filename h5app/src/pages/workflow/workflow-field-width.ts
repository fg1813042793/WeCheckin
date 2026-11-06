import type { WorkflowFormField } from '@/types/workflow'

export function workflowFieldSpanStyle(field: Pick<WorkflowFormField, 'span'>) {
  const value = Number(field.span || 24)
  const span = value === 4.8 || (Number.isInteger(value) && value >= 1 && value <= 24) ? value : 24
  return { width: `${span * 100 / 24}%` }
}
