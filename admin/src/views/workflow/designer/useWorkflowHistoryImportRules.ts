import { computed } from 'vue'
import type { WorkflowDraft, WorkflowHistoryImportColumnMapping, WorkflowHistoryImportConfig } from '@/types/workflow'
import { normalizeWorkflowHistoryImportConfig } from '@/types/workflow'
import {
  historyImportFieldReason, historyImportFields, historyImportRuleErrors,
  historyImportTargetReason, historyImportTypesCompatible,
} from './historyImportRules'
import type { HistoryImportFieldOption } from './historyImportRules'

type MappingSide = keyof WorkflowHistoryImportColumnMapping
type RulesProps = { draft: WorkflowDraft; readonly?: boolean }

export function useWorkflowHistoryImportRules(props: RulesProps, changed: () => void) {
  const config = computed(() => normalizeWorkflowHistoryImportConfig(props.draft.historyImport))
  const fields = computed(() => historyImportFields(props.draft))
  const errors = computed(() => historyImportRuleErrors(config.value, props.draft))
  const fieldByKey = (key: string) => fields.value.find(item => item.field.key === key)?.field

  function commit(next: Required<WorkflowHistoryImportConfig>) {
    if (props.readonly) return
    props.draft.historyImport = next
    changed()
  }

  function patch(patch: Partial<WorkflowHistoryImportConfig>) {
    commit({ ...config.value, ...patch })
  }

  function sourceOptions(index: number): HistoryImportFieldOption[] {
    return fields.value.map(({ field, label }) => {
      const reason = historyImportFieldReason(field)
        || (config.value.mappings.some((mapping, i) => i !== index && mapping.sourceKey === field.key) ? '已用作来源' : '')
      return { value: field.key, label: reason ? `${label} · ${reason}` : label, disabled: !!reason }
    })
  }

  function targetOptions(index: number): HistoryImportFieldOption[] {
    const source = fieldByKey(config.value.mappings[index]?.sourceKey || '')
    return fields.value.map(({ field, label }) => {
      const reason = historyImportTargetReason(props.draft, field)
        || (!source ? '请先选择来源' : !historyImportTypesCompatible(source, field) ? '类型不兼容' : '')
        || (config.value.mappings.some((mapping, i) => i !== index && mapping.targetKey === field.key) ? '已用作目标' : '')
      return { value: field.key, label: reason ? `${label} · ${reason}` : label, disabled: !!reason }
    })
  }

  function isDetailMapping(index: number) {
    const mapping = config.value.mappings[index]
    return !!mapping && fieldByKey(mapping.sourceKey)?.type === 'detail_list' && fieldByKey(mapping.targetKey)?.type === 'detail_list'
  }

  function columnOptions(index: number, columnIndex: number, side: MappingSide): HistoryImportFieldOption[] {
    const mapping = config.value.mappings[index]
    if (!mapping || !isDetailMapping(index)) return []
    const parent = fieldByKey(mapping[side])
    const source = fieldByKey(mapping.sourceKey)?.columns?.find(column => column.key === mapping.columns?.[columnIndex]?.sourceKey)
    return (parent?.columns || []).map((field) => {
      const reason = historyImportFieldReason(field, true)
        || (side === 'targetKey' ? historyImportTargetReason(props.draft, parent) : '')
        || (side === 'targetKey' && !historyImportTypesCompatible(source, field) ? '请先选择兼容的来源列' : '')
        || (mapping.columns?.some((column, i) => i !== columnIndex && column[side] === field.key) ? '该列已使用' : '')
      const label = field.label || '未命名列'
      return { value: field.key, label: reason ? `${label} · ${reason}` : label, disabled: !!reason }
    })
  }

  function addMapping() {
    if (!config.value.enabled || config.value.mappingMode !== 'mapping' || config.value.mappings.length >= 100) return
    const next = normalizeWorkflowHistoryImportConfig(props.draft.historyImport)
    next.mappings.push({ sourceKey: '', targetKey: '' })
    commit(next)
  }

  function removeMapping(index: number) {
    const next = normalizeWorkflowHistoryImportConfig(props.draft.historyImport)
    if (!next.mappings[index]) return
    next.mappings.splice(index, 1)
    commit(next)
  }

  function setMapping(index: number, side: MappingSide, input: unknown) {
    const value = input ?? ''
    if (typeof value !== 'string') return
    const options = side === 'sourceKey' ? sourceOptions(index) : targetOptions(index)
    if (value && !options.some(option => option.value === value && !option.disabled)) return
    const next = normalizeWorkflowHistoryImportConfig(props.draft.historyImport)
    const mapping = next.mappings[index]
    if (!mapping || mapping[side] === value) return
    mapping[side] = value
    delete mapping.columns
    if (side === 'sourceKey' && !historyImportTypesCompatible(fieldByKey(value), fieldByKey(mapping.targetKey))) mapping.targetKey = ''
    commit(next)
  }

  function addColumn(index: number) {
    if (!isDetailMapping(index)) return
    const next = normalizeWorkflowHistoryImportConfig(props.draft.historyImport)
    const mapping = next.mappings[index]!
    if ((mapping.columns?.length || 0) >= 100) return
    mapping.columns = [...(mapping.columns || []), { sourceKey: '', targetKey: '' }]
    commit(next)
  }

  function removeColumn(index: number, columnIndex: number) {
    const next = normalizeWorkflowHistoryImportConfig(props.draft.historyImport)
    const mapping = next.mappings[index]
    if (!mapping?.columns?.[columnIndex]) return
    mapping.columns.splice(columnIndex, 1)
    if (!mapping.columns.length) delete mapping.columns
    commit(next)
  }

  function setColumn(index: number, columnIndex: number, side: MappingSide, input: unknown) {
    const value = input ?? ''
    if (typeof value !== 'string') return
    if (value && !columnOptions(index, columnIndex, side).some(option => option.value === value && !option.disabled)) return
    const next = normalizeWorkflowHistoryImportConfig(props.draft.historyImport)
    const mapping = next.mappings[index]
    const column = mapping?.columns?.[columnIndex]
    if (!mapping || !column || column[side] === value) return
    column[side] = value
    if (side === 'sourceKey') {
      const source = fieldByKey(mapping.sourceKey)?.columns?.find(item => item.key === value)
      const target = fieldByKey(mapping.targetKey)?.columns?.find(item => item.key === column.targetKey)
      if (!historyImportTypesCompatible(source, target)) column.targetKey = ''
    }
    commit(next)
  }

  return {
    config, errors, sourceOptions, targetOptions, columnOptions, isDetailMapping,
    addMapping, removeMapping, setMapping, addColumn, removeColumn, setColumn,
    setEnabled: (value: string | number | boolean) => patch({ enabled: value === true }),
    setOverwrite: (value: string | number | boolean) => patch({ overwrite: value === true }),
    setSourceMode: (value: unknown) => { if (value === 'select_history' || value === 'latest_history') patch({ sourceMode: value }) },
    setSourceStatus: (value: unknown) => { if (value === 'completed' || value === 'all') patch({ sourceStatus: value }) },
    setMappingMode: (value: unknown) => { if (value === 'same_key' || value === 'mapping') patch({ mappingMode: value }) },
  }
}
