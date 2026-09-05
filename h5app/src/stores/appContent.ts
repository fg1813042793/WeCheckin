import type { AppNavItem } from '@/types/navigation'
import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface AppTabCloseGuard {
  hasUnsavedChanges: () => boolean
  saveDraft?: () => Promise<boolean>
}

export interface WorkflowStartSeed {
  definitionId: number
  sourceInstanceId: string
  formData: Record<string, unknown>
}

const tabCloseGuards = new Map<string, AppTabCloseGuard>()

export const useAppContentStore = defineStore('appContent', () => {
  const currentKey = ref('dashboard')
  const focusedReviewId = ref('')
  const focusedWorkflowInstanceId = ref('')
  const focusedWorkflowTab = ref('')
  const workflowStartSeed = ref<WorkflowStartSeed | null>(null)
  const workflowStartSeedTick = ref(0)
  const refreshTick = ref(0)
  const dynamicTabs = ref<AppNavItem[]>([])
  const closeRequestKey = ref('')
  const closeRequestTick = ref(0)

  function switchContent(key: string, reviewId = '') {
    const nextKey = String(key || '').trim()
    currentKey.value = nextKey || 'dashboard'
    focusedReviewId.value = String(reviewId || '').trim()
  }

  function resetContent() {
    currentKey.value = 'dashboard'
    focusedReviewId.value = ''
    focusedWorkflowInstanceId.value = ''
    focusedWorkflowTab.value = ''
    workflowStartSeed.value = null
    dynamicTabs.value = []
    tabCloseGuards.clear()
  }

  function openDynamicTab(tab: Omit<AppNavItem, 'children'> & { children?: AppNavItem[] }) {
    const key = String(tab.key || '').trim()
    if (!key)
      return
    const normalized: AppNavItem = {
      key,
      label: String(tab.label || key).trim() || key,
      icon: String(tab.icon || 'file-text').trim() || 'file-text',
      path: tab.path,
      children: tab.children || [],
    }
    const index = dynamicTabs.value.findIndex(item => item.key === key)
    if (index >= 0)
      dynamicTabs.value.splice(index, 1, normalized)
    else
      dynamicTabs.value.push(normalized)
    currentKey.value = key
    focusedReviewId.value = ''
  }

  function removeDynamicTab(key: string) {
    const normalized = String(key || '').trim()
    dynamicTabs.value = dynamicTabs.value.filter(item => item.key !== normalized)
    tabCloseGuards.delete(normalized)
  }

  function dynamicTab(key: string) {
    return dynamicTabs.value.find(item => item.key === key)
  }

  function registerTabCloseGuard(key: string, guard: AppTabCloseGuard) {
    const normalized = String(key || '').trim()
    if (!normalized)
      return () => {}
    tabCloseGuards.set(normalized, guard)
    return () => {
      if (tabCloseGuards.get(normalized) === guard)
        tabCloseGuards.delete(normalized)
    }
  }

  function hasUnsavedTabChanges(key: string) {
    return Boolean(tabCloseGuards.get(key)?.hasUnsavedChanges())
  }

  function canSaveTabDraft(key: string) {
    return typeof tabCloseGuards.get(key)?.saveDraft === 'function'
  }

  async function saveTabDraft(key: string) {
    const guard = tabCloseGuards.get(key)
    return guard?.saveDraft ? guard.saveDraft() : false
  }

  function requestCloseTab(key: string) {
    closeRequestKey.value = String(key || '').trim()
    closeRequestTick.value += 1
  }

  function focusReview(id: string) {
    focusedReviewId.value = String(id || '').trim()
  }

  function clearFocusedReview() {
    focusedReviewId.value = ''
  }

  function focusWorkflowInstance(id: string) {
    focusedWorkflowInstanceId.value = String(id || '').trim()
  }

  function clearFocusedWorkflowInstance() {
    focusedWorkflowInstanceId.value = ''
  }

  function focusWorkflowTab(tab: string) {
    focusedWorkflowTab.value = String(tab || '').trim()
  }

  function clearFocusedWorkflowTab() {
    focusedWorkflowTab.value = ''
  }

  function setWorkflowStartSeed(seed: WorkflowStartSeed) {
    workflowStartSeed.value = {
      definitionId: Number(seed.definitionId || 0),
      sourceInstanceId: String(seed.sourceInstanceId || '').trim(),
      formData: JSON.parse(JSON.stringify(seed.formData || {})) as Record<string, unknown>,
    }
    workflowStartSeedTick.value += 1
  }

  function takeWorkflowStartSeed(definitionId: number) {
    const seed = workflowStartSeed.value
    if (!seed || seed.definitionId !== definitionId)
      return null
    workflowStartSeed.value = null
    return seed
  }

  function requestRefresh() {
    refreshTick.value += 1
  }

  return {
    clearFocusedWorkflowInstance,
    clearFocusedWorkflowTab,
    clearFocusedReview,
    closeRequestKey,
    closeRequestTick,
    currentKey,
    dynamicTab,
    dynamicTabs,
    focusedReviewId,
    focusedWorkflowInstanceId,
    focusedWorkflowTab,
    focusWorkflowInstance,
    focusWorkflowTab,
    focusReview,
    canSaveTabDraft,
    hasUnsavedTabChanges,
    openDynamicTab,
    registerTabCloseGuard,
    removeDynamicTab,
    refreshTick,
    requestCloseTab,
    requestRefresh,
    resetContent,
    saveTabDraft,
    setWorkflowStartSeed,
    switchContent,
    takeWorkflowStartSeed,
    workflowStartSeed,
    workflowStartSeedTick,
  }
}, {
  persist: false,
})
