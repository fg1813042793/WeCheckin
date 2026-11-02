import { ref, watch } from 'vue'
import type { ComputedRef } from 'vue'
import type { WorkflowHistoryImportConfig } from '@/types/workflow'

// Disclosure state belongs to the editor, never to the saved import rules.
export function useWorkflowHistoryImportPanels(config: ComputedRef<Required<WorkflowHistoryImportConfig>>, errors: ComputedRef<string[]>) {
  const basicOpen = ref(false)
  const mappingsOpen = ref(false)
  const columnsOpen = ref<number[]>([])

  function openColumns(index: number) {
    mappingsOpen.value = true
    if (!columnsOpen.value.includes(index)) columnsOpen.value = [...columnsOpen.value, index]
  }
  function removeMapping(index: number) {
    columnsOpen.value = columnsOpen.value.filter(value => value !== index).map(value => value > index ? value - 1 : value)
  }
  function togglePanel(panel: 'basic' | 'mappings', event: Event) {
    if (event.target !== event.currentTarget) return
    const open = Boolean((event.currentTarget as HTMLDetailsElement | null)?.open)
    if (panel === 'basic') basicOpen.value = open
    else mappingsOpen.value = open
  }
  function toggleColumns(index: number, event: Event) {
    if (event.target !== event.currentTarget || !config.value.mappings[index]) return
    const open = Boolean((event.currentTarget as HTMLDetailsElement | null)?.open)
    columnsOpen.value = open ? [...new Set([...columnsOpen.value, index])] : columnsOpen.value.filter(value => value !== index)
  }

  watch(() => config.value.enabled, (enabled, previous) => {
    if (enabled && !previous) basicOpen.value = true
  }, { flush: 'sync' })
  watch(() => config.value.mappingMode, mode => {
    if (mode === 'mapping' && config.value.enabled) mappingsOpen.value = true
  }, { flush: 'sync' })
  watch(errors, messages => {
    for (const message of messages) {
      const mapping = message.match(/^第 (\d+) 条映射：/)
      if (mapping) {
        mappingsOpen.value = true
        if (message.includes('明细')) openColumns(Number(mapping[1]) - 1)
      } else if (message.includes('映射')) {
        mappingsOpen.value = true
      } else {
        basicOpen.value = true
      }
    }
  }, { immediate: true, flush: 'sync' })

  return { basicOpen, mappingsOpen, columnsOpen, openColumns, removeMapping, togglePanel, toggleColumns }
}
