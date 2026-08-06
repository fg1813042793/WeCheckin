import { ref } from 'vue'
import { normalizeReviewDeepLinkTab, normalizeReviewDeepLinkView } from '../../../router/index'
import { menuPageKeys } from '../../../views/performance/common/constants'
import { menuContentKey, menuLabel, titleForContent } from '../../../views/performance/common/helpers'

export function usePerformanceNavigation({
  state,
  navItems,
  flatMenuItems,
  performanceTabs,
  refreshDataSafely,
  updateReview,
  canSelf,
  canManager,
  canHrbpHandle,
  canEmployeeConfirm,
  canFinal
}) {
  const reviewTab = ref('currentTargets')
  const dashboardFilter = ref('queue')
  const selectedReviewId = ref('')
  const activePerformanceTab = ref('mine')
  const routeTabs = ref([])
  const managerReviewTab = ref('pending')
  const hrbpReviewTab = ref('pending')

  function preferredPerformanceView() {
    const preferred = performanceTabs.value.find((item) => menuContentKey(item.key) === activePerformanceTab.value)
    return preferred?.key || performanceTabs.value[0]?.key || 'performance'
  }

  function syncActivePerformanceTab() {
    if (String(state.view || '').startsWith('performance:')) {
      activePerformanceTab.value = menuContentKey(state.view)
    }
  }

  function routeTabLabel(key) {
    const menu = flatMenuItems.value.find((item) => item.key === key) ||
      navItems.value.find((item) => item.key === key)
    return menu ? menuLabel(menu) : titleForContent(menuContentKey(key))
  }

  function ensureRouteTab(view = state.view) {
    const key = String(view || '')
    if (!key) return
    const nextTab = {
      key,
      label: routeTabLabel(key),
      closable: key !== 'dashboard'
    }
    const index = routeTabs.value.findIndex((item) => item.key === key)
    if (index >= 0) {
      routeTabs.value = routeTabs.value.map((item, itemIndex) => itemIndex === index ? { ...item, ...nextTab } : item)
      return
    }
    routeTabs.value = [...routeTabs.value, nextTab]
  }

  function syncRouteTabLabels() {
    if (!routeTabs.value.length) return
    routeTabs.value = routeTabs.value
      .filter((item) => item && item.key && flatMenuItems.value.some((menu) => menu.key === item.key))
      .map((item) => ({ ...item, label: routeTabLabel(item.key), closable: item.key !== 'dashboard' }))
  }

  function ensureActiveMenu() {
    const items = navItems.value
    if (!items.length) {
      state.view = ''
      routeTabs.value = []
      return
    }
    if (state.view === 'performance' && performanceTabs.value.length > 0) {
      state.view = preferredPerformanceView()
      syncActivePerformanceTab()
      syncRouteTabLabels()
      ensureRouteTab()
      return
    }
    if (flatMenuItems.value.some((item) => item.key === state.view)) {
      syncActivePerformanceTab()
      syncRouteTabLabels()
      ensureRouteTab()
      return
    }
    const first = items[0]
    state.view = first.key === 'performance' ? preferredPerformanceView() : first.key
    syncActivePerformanceTab()
    syncRouteTabLabels()
    ensureRouteTab()
  }

  function resetRouteTabState() {
    routeTabs.value = []
    activePerformanceTab.value = 'mine'
    reviewTab.value = 'currentTargets'
    managerReviewTab.value = 'pending'
    hrbpReviewTab.value = 'pending'
    dashboardFilter.value = 'queue'
  }

  async function switchView(view) {
    const nextView = view === 'performance' ? preferredPerformanceView() : view
    const sameView = state.view === nextView
    state.view = nextView
    syncActivePerformanceTab()
    ensureRouteTab()
    reviewTab.value = 'currentTargets'
    if (!sameView) {
      selectedReviewId.value = ''
    }
    await refreshDataSafely({ contentLoading: true })
  }

  async function activateRouteTab(view) {
    if (!view || view === state.view) return
    await switchView(view)
  }

  async function closeRouteTab(view) {
    const key = String(view || '')
    if (!key) return
    const tabs = routeTabs.value
    const index = tabs.findIndex((item) => item.key === key)
    if (index < 0) return
    const nextTabs = tabs.filter((item) => item.key !== key)
    routeTabs.value = nextTabs
    if (state.view !== key) return
    const nextTab = tabs[index + 1] || tabs[index - 1] || nextTabs[0] || null
    if (nextTab) {
      await switchView(nextTab.key)
      return
    }
    ensureActiveMenu()
    await refreshDataSafely({ contentLoading: true })
  }

  function selectReview(id) {
    selectedReviewId.value = id
    reviewTab.value = 'currentTargets'
  }

  function workbenchTodoTarget(review) {
    if (canSelf(review)) {
      return { view: 'performance:mine', reviewTab: 'currentTargets' }
    }
    if (canManager(review)) {
      return { view: 'performance:manager', managerTab: 'pending', reviewTab: 'manager' }
    }
    if (canHrbpHandle(review)) {
      return { view: 'performance:hrbp', hrbpTab: 'pending', reviewTab: 'hrbp' }
    }
    if (canEmployeeConfirm(review)) {
      return { view: 'performance:mine', reviewTab: 'manager' }
    }
    if (canFinal(review)) {
      return { view: 'performance:hrbp', hrbpTab: 'reviewed', reviewTab: 'hrbp' }
    }
    return { view: 'performance:mine', reviewTab: 'currentTargets' }
  }

  async function openWorkbenchTodo(review, options = {}) {
    if (!review?.id) return
    const target = workbenchTodoTarget(review)
    const preferredView = normalizeReviewDeepLinkView(options.preferredView)
    if (preferredView) {
      target.view = preferredView
      if (preferredView === 'performance:manager') {
        target.managerTab = review.status === 'manager_review' ? 'pending' : 'reviewed'
        target.reviewTab = 'manager'
      } else if (preferredView === 'performance:hrbp') {
        target.hrbpTab = review.status === 'hrbp_review' ? 'pending' : 'reviewed'
        target.reviewTab = 'hrbp'
      } else if (preferredView === 'performance:mine' && !target.reviewTab) {
        target.reviewTab = 'currentTargets'
      }
    }
    const preferredReviewTab = normalizeReviewDeepLinkTab(options.reviewTab)
    if (preferredReviewTab) {
      target.reviewTab = preferredReviewTab
    }
    if (target.managerTab) managerReviewTab.value = target.managerTab
    if (target.hrbpTab) hrbpReviewTab.value = target.hrbpTab
    state.view = target.view
    syncActivePerformanceTab()
    if (options.replaceRouteTabs) {
      routeTabs.value = []
    }
    ensureRouteTab()
    updateReview(review)
    reviewTab.value = target.reviewTab || 'currentTargets'
    await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
    updateReview(review)
    reviewTab.value = target.reviewTab || 'currentTargets'
  }

  return {
    activePerformanceTab,
    routeTabs,
    dashboardFilter,
    reviewTab,
    selectedReviewId,
    managerReviewTab,
    hrbpReviewTab,
    activateRouteTab,
    closeRouteTab,
    ensureActiveMenu,
    openWorkbenchTodo,
    resetRouteTabState,
    selectReview,
    switchView
  }
}
