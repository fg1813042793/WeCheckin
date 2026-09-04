import type { WorkflowFormFieldType, WorkflowValidationOperator } from './types'

const orderedOperators: WorkflowValidationOperator[] = ['eq', 'ne', 'gt', 'gte', 'lt', 'lte']
const equalityOperators: WorkflowValidationOperator[] = ['eq', 'ne']

function comparisonFamily(fieldType: WorkflowFormFieldType) {
  if (fieldType === 'number' || fieldType === 'amount') return 'number'
  if (['text', 'textarea', 'phone', 'email'].includes(fieldType)) return 'text'
  if (fieldType === 'select' || fieldType === 'radio') return 'choice'
  if (fieldType === 'date' || fieldType === 'datetime' || fieldType === 'time') return `temporal:${fieldType}`
  if (fieldType === 'boolean' || fieldType === 'user' || fieldType === 'department') return fieldType
  return ''
}

export function workflowCompareFieldCompatible(left: WorkflowFormFieldType, right: WorkflowFormFieldType) {
  const leftFamily = comparisonFamily(left)
  return leftFamily !== '' && leftFamily === comparisonFamily(right)
}

export function workflowCompareOperators(fieldType: WorkflowFormFieldType): WorkflowValidationOperator[] {
  if (fieldType === 'number' || fieldType === 'amount' || fieldType === 'date' || fieldType === 'datetime' || fieldType === 'time') {
    return [...orderedOperators]
  }
  return comparisonFamily(fieldType) ? [...equalityOperators] : []
}
