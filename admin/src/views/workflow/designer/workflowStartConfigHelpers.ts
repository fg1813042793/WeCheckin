import type { WorkflowDraft } from '@/types/workflow'

export function normalizeIDs(values: Array<number | string>) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value > 0)))
}

export function normalizeDays(values: unknown[], maximum: number) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value >= 1 && value <= maximum))).sort((left, right) => left - right)
}

export function normalizeInteger(value: unknown, minimum: number, maximum: number, fallback: number) {
  const normalized = Math.trunc(Number(value))
  return Number.isInteger(normalized) && normalized >= minimum && normalized <= maximum ? normalized : fallback
}

export function normalizeClock(value: unknown, fallback: string) {
  return typeof value === 'string' && /^([01]\d|2[0-3]):[0-5]\d$/.test(value) ? value : fallback
}

export function flattenFormFields(fields: WorkflowDraft['form']): WorkflowDraft['form'] {
  return fields.flatMap(field => field.type === 'group' ? flattenFormFields(field.fields || []) : [field])
}
