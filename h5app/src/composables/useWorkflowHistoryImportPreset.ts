import type { WorkflowHistoryImportApplied } from '@/pages/workflow/workflow-history-import-preset'
import type {
  WorkflowFieldAccessMap,
  WorkflowFormData,
  WorkflowInstanceDetail,
  WorkflowInstanceSummary,
  WorkflowPublishedDefinition,
} from '@/types/workflow'
import { computed, onScopeDispose, ref, watch } from 'vue'
import { getWorkflowInstance, listWorkflowInstances } from '@/api/workflow'
import { appErrorMessage } from '@/common/app-error'
import { workflowDataFields, workflowFieldAccessMap } from '@/pages/workflow/workflow-form'
import { previewWorkflowHistoryImport } from '@/pages/workflow/workflow-history-import'
import { normalizeWorkflowHistoryImportConfig, workflowHistoryImportSummary } from '@/pages/workflow/workflow-history-import-preset'

interface PresetOptions {
  visible: () => boolean
  definition: () => WorkflowPublishedDefinition
  targetData: () => WorkflowFormData
  targetAccess: () => WorkflowFieldAccessMap
  translate: (key: string, params?: Record<string, string | number>) => string
  apply: (payload: WorkflowHistoryImportApplied) => void
  contextChanged: () => void
}

/** Owns one open dialog session, including legacy requests using capture/isCurrent. */
export function useWorkflowHistoryImportPreset(options: PresetOptions) {
  const busy = ref(false)
  const error = ref('')
  let session = 0
  let active = false
  let applied = false
  const configState = computed(() => {
    try {
      return { config: normalizeWorkflowHistoryImportConfig(options.definition().historyImport), invalid: false }
    }
    catch {
      return { config: null, invalid: true }
    }
  })
  const config = computed(() => configState.value.config)
  const enabled = computed(() => configState.value.invalid || config.value !== null)
  const latestMode = computed(() => config.value?.sourceMode === 'latest_history')
  const statusLocked = computed(() => config.value?.sourceStatus === 'completed')
  const configError = computed(() => configState.value.invalid ? options.translate('workflowHistoryImport.invalidConfig') : '')

  function invalidate() {
    session += 1
    active = false
    busy.value = false
    error.value = ''
  }
  function capture() {
    return session
  }
  function isCurrent(token: number) {
    return active && options.visible() && token === session && !applied
  }
  watch(options.visible, (visible) => {
    invalidate()
    active = visible
    applied = false
  }, { immediate: true, flush: 'sync' })
  watch([
    options.definition,
    () => JSON.stringify([options.definition(), options.targetAccess()]),
  ], () => {
    invalidate()
    if (options.visible())
      options.contextChanged()
  }, { flush: 'sync' })
  onScopeDispose(invalidate)

  function assertSource(instance: WorkflowInstanceSummary) {
    if (instance.definitionId !== options.definition().id
      || (statusLocked.value && instance.status !== 'completed')) {
      throw new Error(options.translate('workflowHistoryImport.sourceChanged'))
    }
  }

  function applyDetail(detail: WorkflowInstanceDetail, token: number) {
    if (!isCurrent(token) || !config.value)
      return
    assertSource(detail.instance)
    const sourceAccess = workflowFieldAccessMap(detail.form || [], detail.fieldPermissions?.[detail.startNodeId || 'start'] || [], 'write')
    const targetFields = options.definition().form || []
    const mappings = config.value.mappingMode === 'mapping'
      ? config.value.mappings
      : workflowDataFields(targetFields).filter(field => field.type !== 'calculation').map(field => ({ sourceKey: field.key, targetKey: field.key }))
    const preview = previewWorkflowHistoryImport({
      sourceFields: detail.form || [],
      sourceData: detail.formData || {},
      sourceAccess,
      targetFields,
      targetData: options.targetData(),
      targetAccess: options.targetAccess(),
      mappings,
      overwrite: config.value.overwrite,
    })
    const summary = workflowHistoryImportSummary(preview)
    if (summary.imported === 0) {
      error.value = `${options.translate('workflowHistoryImport.noValues')} ${options.translate('workflowHistoryImport.result', summary)}`
      return
    }
    if (!isCurrent(token))
      return
    applied = true
    options.apply({ sourceInstanceId: detail.instance.id, patch: preview.patch, importCount: summary.imported, summary })
  }

  async function execute(instance?: WorkflowInstanceSummary) {
    const token = capture()
    if (!isCurrent(token) || busy.value || !enabled.value)
      return
    error.value = ''
    if (configError.value || !config.value) {
      error.value = configError.value
      return
    }
    busy.value = true
    try {
      let source = instance
      if (!source) {
        const response = await listWorkflowInstances({
          definitionId: options.definition().id,
          scope: 'started',
          status: statusLocked.value ? 'completed' : undefined,
          page: 1,
          pageSize: 1,
        })
        if (!isCurrent(token))
          return
        source = response?.data?.list?.[0]
        if (!source) {
          error.value = options.translate('workflowHistoryImport.empty')
          return
        }
      }
      assertSource(source)
      const response = await getWorkflowInstance(source.id)
      if (!isCurrent(token))
        return
      if (!response?.data || response.data.instance.id !== source.id)
        throw new Error(options.translate('workflowHistoryImport.readFailed'))
      applyDetail(response.data, token)
    }
    catch (cause) {
      if (isCurrent(token))
        error.value = appErrorMessage(cause, options.translate('workflowHistoryImport.readFailed'))
    }
    finally {
      if (token === session)
        busy.value = false
    }
  }

  return {
    config,
    enabled,
    latestMode,
    statusLocked,
    configError,
    busy,
    error,
    capture,
    isCurrent,
    invalidate,
    select: (instance: WorkflowInstanceSummary) => execute(instance),
    latest: () => execute(),
  }
}
