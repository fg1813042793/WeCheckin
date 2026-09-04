import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export type AppMenuLayout = 'side' | 'top'

export const useAppShellStore = defineStore('appShell', () => {
  const menuLayout = ref<AppMenuLayout>('side')
  const sidebarCollapsed = ref(false)
  const expandedRootKeys = ref<string[]>([])
  const openedTabKeys = ref<string[]>([])

  const openedTabKeySet = computed(() => new Set(openedTabKeys.value))
  const expandedRootKeySet = computed(() => new Set(expandedRootKeys.value))

  function rememberOpenedTab(key: string, homeKey = 'dashboard') {
    const nextKey = String(key || '').trim()
    if (!nextKey || nextKey === homeKey || openedTabKeySet.value.has(nextKey)) {
      return
    }
    openedTabKeys.value = [...openedTabKeys.value, nextKey]
  }

  function removeOpenedTab(key: string) {
    openedTabKeys.value = openedTabKeys.value.filter(item => item !== key)
  }

  function isRootExpanded(key: string) {
    return expandedRootKeySet.value.has(key)
  }

  function expandRootOnly(key: string) {
    const nextKey = String(key || '').trim()
    expandedRootKeys.value = nextKey ? [nextKey] : []
  }

  function toggleRootExpansion(key: string) {
    const nextKey = String(key || '').trim()
    if (!nextKey) {
      return
    }
    expandedRootKeys.value = isRootExpanded(nextKey) ? [] : [nextKey]
  }

  function toggleSidebarCollapsed() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setSidebarCollapsed(collapsed: boolean) {
    sidebarCollapsed.value = collapsed
  }

  function setMenuLayout(layout: AppMenuLayout) {
    menuLayout.value = layout
  }

  function toggleMenuLayout() {
    menuLayout.value = menuLayout.value === 'side' ? 'top' : 'side'
  }

  return {
    expandRootOnly,
    expandedRootKeys,
    isRootExpanded,
    menuLayout,
    openedTabKeys,
    rememberOpenedTab,
    removeOpenedTab,
    setMenuLayout,
    setSidebarCollapsed,
    sidebarCollapsed,
    toggleMenuLayout,
    toggleRootExpansion,
    toggleSidebarCollapsed,
  }
}, {
  persist: {
    paths: ['menuLayout'],
  },
})
