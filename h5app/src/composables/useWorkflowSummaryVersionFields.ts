import type { WorkflowSummaryFormFieldGroup } from '@/pages/workflow/workflow-summary-columns'
import type { WorkflowInstanceSummary, WorkflowPublishedDefinition } from '@/types/workflow'
import { onScopeDispose, ref, watch } from 'vue'
import { getWorkflowSummaryInstance } from '@/api/workflow'
import { workflowSummaryFormFieldGroups } from '@/pages/workflow/workflow-summary-columns'

export function useWorkflowSummaryVersionFields(
  definitions: () => WorkflowPublishedDefinition[],
  instances: () => WorkflowInstanceSummary[],
  permissionVersion: () => unknown,
) {
  const groups = ref<Record<string, WorkflowSummaryFormFieldGroup>>({})
  const loading = ref<Record<string, boolean>>({})
  const errors = ref<Record<string, string>>({})
  let generation = 0
  const pending = new Map<string, Promise<void>>()

  function reset() {
    generation++
    groups.value = {}
    loading.value = {}
    errors.value = {}
    pending.clear()
  }
  watch(permissionVersion, reset)
  onScopeDispose(reset)

  function ensure(definitionId: number, version: number): Promise<void> {
    const key = `${definitionId}:${version}`
    if (!Number.isInteger(definitionId) || definitionId <= 0 || !Number.isInteger(version) || version <= 0)
      return Promise.resolve()
    const current = definitions().find(item => item.id === definitionId)
    if (current?.version === version || groups.value[key])
      return Promise.resolve()
    if (pending.has(key))
      return pending.get(key)!
    const instance = instances().find(item => item.definitionId === definitionId && item.definitionVersion === version)
    if (!instance) {
      errors.value = { ...errors.value, [key]: 'versionNotOnPage' }
      return Promise.resolve()
    }
    const requestGeneration = generation
    loading.value = { ...loading.value, [key]: true }
    errors.value = { ...errors.value, [key]: '' }
    const request = (async () => {
      try {
        // Historical schemas are only read through an authorized summary record.
        const response = await getWorkflowSummaryInstance(instance.id)
        if (requestGeneration !== generation)
          return
        const detail = response.data
        if (!detail || detail.instance.id !== instance.id || detail.instance.definitionId !== definitionId || detail.instance.definitionVersion !== version)
          throw new Error('version mismatch')
        const definition = { ...current, id: definitionId, name: current?.name || detail.instance.definitionName, version, form: detail.form } as WorkflowPublishedDefinition
        const group = workflowSummaryFormFieldGroups([definition])[0]
        groups.value = { ...groups.value, [key]: group || { definitionId, version, label: definition.name, columns: [] } }
      }
      catch {
        if (requestGeneration === generation)
          errors.value = { ...errors.value, [key]: 'versionLoadFailed' }
      }
      finally {
        if (requestGeneration === generation) {
          loading.value = { ...loading.value, [key]: false }
          pending.delete(key)
        }
      }
    })()
    pending.set(key, request)
    return request
  }

  return { groups, loading, errors, ensure }
}
