import type { WorkflowHistoryImportPreview } from './workflow-history-import'
import type {
  WorkflowHistoryImportColumnMapping,
  WorkflowHistoryImportConfig,
  WorkflowHistoryImportMapping,
} from '@/types/workflow'

export type WorkflowHistoryImportPreset = Required<WorkflowHistoryImportConfig>

// Matches workflowcore's field-key contract, including its 100-character limit.
const fieldKeyPattern = /^[A-Z][\w.-]{0,99}$/i

function invalidConfig(): never {
  throw new Error('workflowHistoryImport.invalidConfig')
}

function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value))
    return invalidConfig()
  return value as Record<string, unknown>
}

function mappingList(value: unknown, columns = false): WorkflowHistoryImportMapping[] {
  if (!Array.isArray(value) || value.length < 1 || value.length > 100)
    return invalidConfig()
  const sources = new Set<string>()
  const targets = new Set<string>()
  return value.map((item) => {
    const raw = record(item)
    const { sourceKey, targetKey } = raw
    if (typeof sourceKey !== 'string' || typeof targetKey !== 'string'
      || !fieldKeyPattern.test(sourceKey) || !fieldKeyPattern.test(targetKey)
      || sources.has(sourceKey) || targets.has(targetKey)) {
      return invalidConfig()
    }
    sources.add(sourceKey)
    targets.add(targetKey)
    const result: WorkflowHistoryImportMapping = { sourceKey, targetKey }
    if (raw.columns !== undefined) {
      if (columns)
        return invalidConfig()
      result.columns = mappingList(raw.columns, true) as WorkflowHistoryImportColumnMapping[]
    }
    return result
  })
}

function checkMappingDraftSize(value: unknown, columns = false) {
  if (value === undefined || value === null)
    return
  if (!Array.isArray(value) || value.length > 100)
    return invalidConfig()
  for (const item of value) {
    const raw = record(item)
    for (const key of ['sourceKey', 'targetKey']) {
      if (raw[key] !== undefined && (typeof raw[key] !== 'string' || (raw[key] as string).length > 100))
        return invalidConfig()
    }
    if (!columns)
      checkMappingDraftSize(raw.columns, true)
  }
}

export function normalizeWorkflowHistoryImportConfig(value: unknown): WorkflowHistoryImportPreset | null {
  if (value === undefined || value === null)
    return null
  const raw = record(value)
  if (raw.enabled === false)
    return null
  if (raw.enabled !== true)
    return invalidConfig()
  const sourceMode = raw.sourceMode === undefined ? 'select_history' : raw.sourceMode
  const sourceStatus = raw.sourceStatus === undefined ? 'completed' : raw.sourceStatus
  const mappingMode = raw.mappingMode === undefined ? 'same_key' : raw.mappingMode
  const overwrite = raw.overwrite === undefined ? false : raw.overwrite
  if ((sourceMode !== 'select_history' && sourceMode !== 'latest_history')
    || (sourceStatus !== 'completed' && sourceStatus !== 'all')
    || (mappingMode !== 'same_key' && mappingMode !== 'mapping') || typeof overwrite !== 'boolean') {
    return invalidConfig()
  }
  // Saved manual drafts are inactive in same_key mode, matching live start-config validation.
  checkMappingDraftSize(raw.mappings)
  const mappings = mappingMode === 'mapping' ? mappingList(raw.mappings) : []
  return { enabled: true, sourceMode, sourceStatus, mappingMode, overwrite, mappings }
}

export function workflowHistoryImportSummary(preview: WorkflowHistoryImportPreview) {
  return {
    imported: preview.importCount + preview.overwriteCount,
    overwritten: preview.overwriteCount,
    retained: preview.retainCount,
    skipped: preview.skipCount,
    filtered: preview.items.reduce((count, item) => count + (item.filteredCount || 0), 0),
    dynamic: preview.items.reduce((count, item) => count + (item.dynamicCount || 0), 0),
  }
}

export type WorkflowHistoryImportSummary = ReturnType<typeof workflowHistoryImportSummary>

export function workflowHistoryImportNeedsAttention(summary: WorkflowHistoryImportSummary) {
  return summary.skipped > 0 || summary.filtered > 0 || summary.dynamic > 0
}

export interface WorkflowHistoryImportApplied {
  sourceInstanceId: string
  patch: WorkflowHistoryImportPreview['patch']
  importCount: number
  summary: WorkflowHistoryImportSummary
}
