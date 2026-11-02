import type { WorkflowSummaryColumn, WorkflowSummaryFormFieldGroup } from './workflow-summary-columns'
import type { WorkflowFormField, WorkflowInstanceSummary } from '@/types/workflow'
import { workflowSummaryFormValueInfo } from './workflow-summary-columns'

const maxColumns = 20
const maxSources = 50
const maxLabelLength = 50
const maxFieldKeyLength = 100
const maxIdLength = 80

export interface SummaryColumnSource {
  definitionId: number
  fieldKey: string
  version?: number
  fieldLabel?: string
}

export interface SummaryMappedColumn {
  id: string
  label: string
  enabled: boolean
  sources: SummaryColumnSource[]
}

export type SummaryVersionGroups = Record<string, WorkflowSummaryFormFieldGroup>

type SummaryMappedReason = 'unmapped' | 'definitionUnavailable' | 'fieldUnavailable' | 'versionUnavailable' | 'ambiguousMapping' | 'missingData'
interface SummaryMappedValue {
  text: string
  calculated: boolean
  reason?: SummaryMappedReason
}

type SummaryFieldColumn = WorkflowSummaryColumn & { field: WorkflowFormField }
type SourceResolution = { column: SummaryFieldColumn, reason?: never } | { column?: never, reason: SummaryMappedReason }

function isRecord(value: unknown): value is Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value))
    return false
  const prototype = Object.getPrototypeOf(value)
  return prototype === Object.prototype || prototype === null
}

function positiveInteger(value: unknown): value is number {
  return typeof value === 'number' && Number.isSafeInteger(value) && value > 0
}

function validIdentity(value: unknown, maxLength: number): value is string {
  return typeof value === 'string' && value.trim().length > 0 && value.length <= maxLength
}

function validSource(value: unknown): value is SummaryColumnSource {
  return isRecord(value)
    && positiveInteger(value.definitionId)
    && validIdentity(value.fieldKey, maxFieldKeyLength)
    && (value.version === undefined || positiveInteger(value.version))
}

/** Accept parsed storage data only; unavailable references remain editable. */
export function normalizeSummaryMappedColumns(raw: unknown): SummaryMappedColumn[] {
  if (!Array.isArray(raw))
    return []
  const result: SummaryMappedColumn[] = []
  const ids = new Set<string>()
  for (const item of raw.slice(0, maxColumns)) {
    if (!isRecord(item) || !validIdentity(item.id, maxIdLength) || ids.has(item.id))
      continue
    if (typeof item.label !== 'string' || !item.label.trim())
      continue
    const sources: SummaryColumnSource[] = []
    for (const candidate of Array.isArray(item.sources) ? item.sources.slice(0, maxSources) : []) {
      if (!validSource(candidate))
        continue
      // Do not trim/truncate identities or silently deduplicate competing sources.
      const source: SummaryColumnSource = { definitionId: candidate.definitionId, fieldKey: candidate.fieldKey }
      if (candidate.version !== undefined)
        source.version = candidate.version
      if (typeof candidate.fieldLabel === 'string')
        source.fieldLabel = candidate.fieldLabel.trim().slice(0, maxLabelLength)
      sources.push(source)
    }
    result.push({ id: item.id, label: item.label.trim().slice(0, maxLabelLength), enabled: item.enabled === true, sources })
    ids.add(item.id)
  }
  return result
}

/** Structural validation never depends on the currently loaded API definitions. */
export function validateSummaryMappedColumn(config: SummaryMappedColumn): string[] {
  const errors = new Set<string>()
  if (typeof config.id !== 'string' || !config.id.trim())
    errors.add('idRequired')
  else if (config.id.length > maxIdLength)
    errors.add('idTooLong')
  if (typeof config.label !== 'string' || !config.label.trim())
    errors.add('labelRequired')
  else if (config.label.trim().length > maxLabelLength)
    errors.add('labelTooLong')
  if (!Array.isArray(config.sources) || config.sources.length === 0) {
    errors.add('sourcesRequired')
    return [...errors]
  }
  if (config.sources.length > maxSources)
    errors.add('tooManySources')
  const identities = new Set<string>()
  for (const source of config.sources.slice(0, maxSources)) {
    if (!validSource(source)) {
      errors.add('sourceRequired')
      continue
    }
    if (source.fieldLabel !== undefined && (typeof source.fieldLabel !== 'string' || source.fieldLabel.length > maxLabelLength))
      errors.add('sourceRequired')
    const identity = `${source.definitionId}:${source.version ?? 'default'}`
    if (identities.has(identity))
      errors.add('duplicateSource')
    identities.add(identity)
  }
  return [...errors]
}

export function summaryMappedColumns(configs: SummaryMappedColumn[]): WorkflowSummaryColumn[] {
  return configs.slice(0, maxColumns).filter(config => config.enabled).map(config => ({
    key: `mapped:${config.id}`,
    label: config.label,
    kind: 'form',
    width: 'minmax(150px,1fr)',
    mobileHidden: true,
  }))
}

function resolveSource(
  source: SummaryColumnSource,
  groups: WorkflowSummaryFormFieldGroup[],
  versionGroups: SummaryVersionGroups,
): SourceResolution {
  let candidates: WorkflowSummaryFormFieldGroup[]
  if (source.version !== undefined) {
    if (!positiveInteger(source.version))
      return { reason: 'versionUnavailable' }
    const key = `${source.definitionId}:${source.version}`
    if (Object.prototype.hasOwnProperty.call(versionGroups, key)) {
      const historical = versionGroups[key]
      if (!historical || historical.definitionId !== source.definitionId || (historical.version !== undefined && historical.version !== source.version))
        return { reason: 'versionUnavailable' }
      candidates = [historical]
    }
    else {
      candidates = groups.filter(group => group.definitionId === source.definitionId && group.version === source.version)
    }
    if (candidates.length === 0)
      return { reason: 'versionUnavailable' }
  }
  else {
    candidates = groups.filter(group => group.definitionId === source.definitionId)
    if (candidates.length === 0)
      return { reason: 'definitionUnavailable' }
  }
  if (candidates.length > 1)
    return { reason: 'ambiguousMapping' }
  const columns = candidates[0].columns.filter((column): column is SummaryFieldColumn =>
    column.kind === 'form' && column.definitionId === source.definitionId && column.field?.key === source.fieldKey,
  )
  if (columns.length === 0)
    return { reason: 'fieldUnavailable' }
  if (columns.length > 1)
    return { reason: 'ambiguousMapping' }
  return { column: columns[0] }
}

function missingValue(reason: SummaryMappedReason): SummaryMappedValue {
  return { text: '-', calculated: false, reason }
}

export function summaryMappedValue(
  instance: WorkflowInstanceSummary,
  config: SummaryMappedColumn,
  groups: WorkflowSummaryFormFieldGroup[],
  versionGroups: SummaryVersionGroups = {},
): SummaryMappedValue {
  const matching = config.sources.filter(source => source.definitionId === instance.definitionId)
  const exact = matching.filter(source => source.version !== undefined && source.version === instance.definitionVersion)
  // Select first, then resolve: an unavailable exact schema must never use a default.
  const candidates = exact.length > 0 ? exact : matching.filter(source => source.version === undefined)
  if (candidates.length === 0)
    return missingValue('unmapped')
  if (candidates.length > 1)
    return missingValue('ambiguousMapping')
  const resolved = resolveSource(candidates[0], groups, versionGroups)
  if (!resolved.column)
    return missingValue(resolved.reason)

  // Inherited properties (e.g. constructor) are not persisted form values.
  const formData = { ...instance.formData }
  Object.setPrototypeOf(formData, null)
  const column = resolved.column
  const result = workflowSummaryFormValueInfo({ ...instance, formData }, column)
  if (result.calculated)
    return result
  const value = formData[column.field.key]
  if (value == null || (typeof value === 'string' && !value.trim()) || (Array.isArray(value) && value.length === 0))
    return missingValue('missingData')
  if (column.field.type === 'calculation' && ((typeof value !== 'number' && typeof value !== 'string') || !Number.isFinite(Number(value))))
    return missingValue('missingData')
  return result
}

function fieldTypeFamily(field: WorkflowFormField): string {
  if (field.type === 'calculation')
    return field.calculation?.resultDisplay?.mode === 'label' ? 'choice' : 'numeric'
  if (field.type === 'number' || field.type === 'amount')
    return 'numeric'
  if (field.type === 'select' || field.type === 'radio')
    return 'choice'
  if (field.type === 'checkbox' || field.type === 'multi_select')
    return 'multiChoice'
  if (['text', 'textarea', 'phone', 'email'].includes(field.type))
    return 'text'
  return field.type
}

/** A warning only: unknown/deleted sources do not block the remaining mapping. */
export function summaryMappedTypeMismatch(
  config: SummaryMappedColumn,
  groups: WorkflowSummaryFormFieldGroup[],
  versionGroups: SummaryVersionGroups = {},
): boolean {
  const families = new Set<string>()
  for (const source of config.sources) {
    const resolved = resolveSource(source, groups, versionGroups)
    if (resolved.column)
      families.add(fieldTypeFamily(resolved.column.field))
  }
  return families.size > 1
}
