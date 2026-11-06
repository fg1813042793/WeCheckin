import type { WorkflowFormFieldSpan } from './types'

function validWorkflowFieldSpan(span: number) {
  return span === 4.8 || (Number.isInteger(span) && span >= 1 && span <= 24)
}

export function workflowFieldSpan(field: { span?: unknown }): WorkflowFormFieldSpan {
  const span = Number(field.span)
  return validWorkflowFieldSpan(span) ? span as WorkflowFormFieldSpan : 24
}

export function workflowFieldGridSpan(field: { span?: unknown }): number {
  // CSS grid lines must be integers. 120 tracks represent every supported ratio.
  return Math.round(workflowFieldSpan(field) * 5)
}

export function workflowFieldSpanLabel(value: number) {
  const span = validWorkflowFieldSpan(value) ? value : 24
  if (span === 24)
    return '24/24 · 整行'
  if (span === 4.8)
    return '4.8/24 · 1/5（历史）'
  const divisor = greatestCommonDivisor(span, 24)
  return `${span}/24 · ${span / divisor}/${24 / divisor}`
}

function greatestCommonDivisor(left: number, right: number): number {
  let a = Math.abs(left)
  let b = Math.abs(right)
  while (b > 0) {
    const remainder = a % b
    a = b
    b = remainder
  }
  return a || 1
}
