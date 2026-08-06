import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { providePerformanceContext } from '../../views/performance/common/context'
import { reviewDetail as fetchReviewDetail } from '../../api/performance/common/reviews'
import { setNavigationTitle } from '../../utils/dingtalk'
import { AUTH_EXPIRED_EVENT, authToken } from '../../utils/request'
import {
  historyMonthOptions,
  menuPageKeys,
  myPerformanceStatusSet,
  statusMeta
} from '../../views/performance/common/constants'
import {
  createTargetUserMeta,
  defaultAppConfig,
  firstText,
  flattenMenuTree,
  menuContentKey,
  menuIcon,
  menuLabel,
  normalizeAppConfig,
  normalizeMenuTree,
  titleForContent,
  unique
} from '../../views/performance/common/helpers'
import {
  usePerformanceNavigation
} from './composables/usePerformanceNavigation'
import { normalizeReviewDeepLinkTab, normalizeReviewDeepLinkView } from '../../router/index'
import { usePerformanceAuth } from './composables/usePerformanceAuth'
import { usePerformanceData } from './composables/usePerformanceData'
import { usePerformancePermissions } from './composables/usePerformancePermissions'
import { usePerformanceProfile } from '../../views/profile/composables/usePerformanceProfile'
import { usePerformanceReviewActions } from '../../views/performance/common/composables/usePerformanceReviewActions'
import { usePerformanceReviewCreation } from '../../views/performance/mine/composables/usePerformanceReviewCreation'
import { usePerformanceReviewList } from '../../views/performance/common/composables/usePerformanceReviewList'

export function usePerformanceApp() {
  const refreshing = ref(false)
  const contentLoading = ref(false)

  const summaryFilters = reactive({
    employeeName: '',
    departmentName: '',
    departmentNames: [],
    period: '',
    status: ''
  })
  let contentLoadingSeq = 0
  let reviewDeepLinkApplied = false

  const state = reactive({
    user: null,
    menus: [],
    users: [],
    reviews: [],
    reviewListTotal: 0,
    reviewPage: 1,
    reviewPageSize: 20,
    workbenchStats: [],
    template: null,
    buttonPermissionKeys: [],
    buttonPermissionReady: false,
    apiPermissionKeys: [],
    apiPermissionReady: false,
    appConfig: defaultAppConfig(),
    appTitle: '',
    permissionVersion: 0,
    view: 'dashboard'
  })

  const menuTreeItems = computed(() => normalizeMenuTree(state.menus))
  const flatMenuItems = computed(() => flattenMenuTree(menuTreeItems.value))

  const {
    canAddNextObjective,
    canCreateReview,
    canDeleteNextObjective,
    canDeleteReview,
    canEditNextObjectives,
    canEditObjectiveDimension,
    canEditTemplate,
    canEditUsers,
    canEmployeeConfirm,
    canExportReviews,
    canFinal,
    canHrbpHandle,
    canManager,
    canPerformReviewAction,
    canSelf,
    canWithdraw,
    hasApiPermission,
    hasButtonPermission,
    hasMenuPermission
  } = usePerformancePermissions({
    state,
    flatMenuItems,
    sameUserId: (left, right) => sameUserId(left, right)
  })

  const navItems = computed(() => {
    if (menuTreeItems.value.length > 0) {
      const items = menuTreeItems.value
        .filter((item) => ['dashboard', 'performance'].includes(item.key))
        .map((item) => ({
          key: item.key,
          label: menuLabel(item),
          icon: menuIcon(item),
          children: (item.children || [])
            .filter((child) => menuPageKeys.has(child.key))
            .map((child) => ({ key: child.key, label: menuLabel(child), icon: menuIcon(child) }))
        }))
      if (items.length > 0) return items
    }
    return []
  })

  const performanceTabs = computed(() => {
    const performanceMenu = menuTreeItems.value.find((item) => item.key === 'performance')
    return (performanceMenu?.children || [])
      .filter((item) => menuPageKeys.has(item.key))
      .map((item) => ({ key: item.key, label: menuLabel(item), icon: menuIcon(item) }))
  })

  const {
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
  } = usePerformanceNavigation({
    state,
    navItems,
    flatMenuItems,
    performanceTabs,
    refreshDataSafely: (options) => refreshDataSafely(options),
    updateReview: (review) => updateReview(review),
    canSelf: (review) => canSelf(review),
    canManager: (review) => canManager(review),
    canHrbpHandle: (review) => canHrbpHandle(review),
    canEmployeeConfirm: (review) => canEmployeeConfirm(review),
    canFinal: (review) => canFinal(review)
  })

  const activeView = computed(() => String(state.view || '').startsWith('performance:') ? 'performance' : state.view)
  const activeMenuItem = computed(() => {
    return flatMenuItems.value.find((item) => item.key === state.view) ||
      navItems.value.find((item) => item.key === activeView.value) ||
      null
  })
  const contentView = computed(() => menuContentKey(state.view))
  const appConfig = computed(() => normalizeAppConfig(state.appConfig))
  const appTitle = computed(() => firstText(appConfig.value.appTitle, state.appTitle, defaultAppConfig().appTitle))

  const {
    profileDialog,
    profileAvatarPreview,
    profileDisplayName,
    profileInitial,
    chooseProfileAvatar,
    clearProfileAvatar,
    closeProfileDialog,
    openProfileDialog,
    resetProfileDialog,
    submitProfileDialog
  } = usePerformanceProfile({
    state,
    sameUserId: (left, right) => sameUserId(left, right),
    sanitizeUsers: (users) => sanitizeUsers(users),
    toast: (title) => toast(title)
  })

  function syncDingTalkPageTitle(title) {
    const text = firstText(title, appConfig.value.appName, defaultAppConfig().appTitle)
    if (!text) return
    if (typeof document !== 'undefined') {
      document.title = text
    }
    try {
      setNavigationTitle(text)
    } catch (error) {
      // Some non-uni preview containers do not expose the navigation bridge.
    }
  }

  watch(appTitle, (title) => {
    syncDingTalkPageTitle(title)
  }, { immediate: true })

  const pageTitle = computed(() => {
    return activeMenuItem.value?.label || titleForContent(contentView.value)
  })

  const sectionTitle = computed(() => {
    return activeMenuItem.value?.label || titleForContent(contentView.value)
  })

  const selectedReview = computed(() => {
    if (contentView.value === 'mine') {
      return currentReviews.value.find((item) => item.id === selectedReviewId.value) || currentReviews.value[0] || null
    }
    return state.reviews.find((item) => item.id === selectedReviewId.value) || currentReviews.value[0] || null
  })

  const currentReviews = computed(() => {
    let list = [...state.reviews]
    const view = contentView.value
    if (view === 'mine') {
      list = myPerformanceReviews()
    } else if (view === 'history') {
      list = list.filter((item) => sameUserId(item.employeeId, state.user?.id) && item.status === 'completed')
    } else if (view === 'manager') {
      list = list.filter((item) => sameUserId(item.managerId, state.user?.id) && ['manager_review', 'hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(item.status))
    } else if (view === 'hrbp') {
      list = list.filter((item) => ['hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(item.status))
    } else if (view === 'dashboard') {
      list = queueReviews()
    }
    return list
  })

  const summaryReviews = computed(() => {
    const employeeName = summaryFilters.employeeName.trim().toLowerCase()
    return state.reviews.filter((item) => {
      if (item.status !== 'completed') return false
      const employee = userName(item.employeeId)
      const employeeText = [employee, item.employeeName, item.employeeId].filter(Boolean).join(' ').toLowerCase()
      if (employeeName && !employeeText.includes(employeeName)) return false
      if (!summaryDepartmentMatches(item.department, summaryFilters.departmentName, summaryFilters.departmentNames)) return false
      if (summaryFilters.period && item.period !== summaryFilters.period) return false
      return true
    })
  })

  const statCards = computed(() => {
    const all = state.reviews.length
    const draft = state.reviews.filter((item) => item.status === 'draft').length
    const reviewing = state.reviews.filter((item) => ['manager_review', 'hrbp_review', 'employee_confirm', 'hr_final'].includes(item.status)).length
    const completed = state.reviews.filter((item) => item.status === 'completed').length
    return [
      ['queue', '我的待办', queueReviews().length],
      ['all', '全部考评单', all],
      ['draft', '员工填写', draft],
      ['reviewing', '流转中', reviewing],
      ['completed', '已完成', completed]
    ]
  })

  const workbenchCards = computed(() => state.workbenchStats.map((item) => [item.key, item.label, item.value]))

  const departments = computed(() => unique(state.reviews.map((item) => item.department)))
  const managerIds = computed(() => unique(state.reviews.map((item) => item.managerId).filter(Boolean)))
  const hrbpIds = computed(() => unique(state.reviews.map((item) => item.hrbpId).filter(Boolean)))
  const grades = computed(() => (state.template?.gradeLevels || []).map((item) => item.grade))
  const historyYearOptions = computed(() => {
    const nowYear = new Date().getFullYear()
    const years = new Set()
    for (let year = nowYear; year >= nowYear - 5; year -= 1) {
      years.add(String(year))
    }
    for (const review of state.reviews || []) {
      const match = String(review.period || '').match(/^(\d{4})-/)
      if (match) years.add(match[1])
    }
    return [...years].sort((left, right) => Number(right) - Number(left))
  })
  const reviewTargetUsers = computed(() => state.users.filter((item) => item && item.id && Number(item.status || 1) === 1))

  const {
    historyFilters,
    managerReviewedFilters,
    hrbpReviewedFilters,
    loadReviews,
    switchManagerReviewTab,
    switchHrbpReviewTab,
    setHistoryFilter,
    resetHistoryFilters,
    setManagerReviewedFilter,
    searchManagerReviewedReviews,
    resetManagerReviewedFilters,
    setHrbpReviewedFilter,
    searchHrbpReviewedReviews,
    resetHrbpReviewedFilters
  } = usePerformanceReviewList({
    state,
    contentView,
    selectedReviewId,
    managerReviewTab,
    hrbpReviewTab,
    ensureSelectedReview: (options) => ensureSelectedReview(options),
    refreshDataSafely: (options) => refreshDataSafely(options)
  })

  const {
    deleteReview,
    deleteUser,
    ensureSelectedReview,
    exportSummary,
    loadUsers,
    refreshData,
    sanitizeUsers,
    saveTemplate,
    saveUser,
    updateReview
  } = usePerformanceData({
    state,
    contentView,
    currentReviews,
    navItems,
    selectedReviewId,
    summaryFilters,
    ensureActiveMenu: () => ensureActiveMenu(),
    ensureNewReviewEmployee: () => ensureNewReviewEmployee(),
    loadReviews: (params, options) => loadReviews(params, options),
    hasApiPermission: (key) => hasApiPermission(key),
    canDeleteReview: () => canDeleteReview(),
    canEditTemplate: () => canEditTemplate(),
    canExportReviews: () => canExportReviews(),
    confirmReviewAction: (action) => confirmReviewAction(action),
    toast: (title) => toast(title)
  })

  const {
    newReview,
    createReviewDialog,
    createReviewForm,
    createReviewUserKeyword,
    createReviewTargetUsers,
    createTargetUserTree,
    createTargetUserTreeRows,
    createTargetUserEmptyText,
    createTargetDepartmentCheckState,
    createTargetDepartmentUserIds,
    createReviewMonthText,
    createReviewMonthPickerOpen,
    createReviewMonthPickerYear,
    createReviewMonthOptions,
    isCreateReviewMonthSelected,
    closeCreateReviewDialog,
    ensureNewReviewEmployee,
    openCreateReviewDialog,
    toggleCreateReviewDept,
    toggleCreateReviewDepartment,
    toggleCreateReviewEmployee,
    toggleCreateReviewMonthPicker,
    changeCreateReviewMonthPickerYear,
    selectCreateReviewMonth,
    createReview
  } = usePerformanceReviewCreation({
    state,
    selectedReviewId,
    reviewTargetUsers,
    loadUsers: () => loadUsers(),
    loadReviews: (params, options) => loadReviews(params, options),
    hasApiPermission: (key) => hasApiPermission(key),
    canCreateReview: () => canCreateReview(),
    toast: (title) => toast(title)
  })

  const {
    withdrawDialog,
    withdrawReasonLength,
    closeWithdrawDialog,
    withdrawReview,
    submitWithdrawReview,
    returnDialog,
    returnReasonLength,
    closeReturnDialog,
    returnReview,
    submitReturnReview,
    disputeDialog,
    disputeReasonLength,
    closeDisputeDialog,
    submitDisputeReview,
    confirmReviewAction,
    performReviewAction
  } = usePerformanceReviewActions({
    selectedReview,
    canPerformReviewAction: (action) => canPerformReviewAction(action),
    loadReviews: (params, options) => loadReviews(params, options),
    updateReview: (review) => updateReview(review),
    toast: (title) => toast(title)
  })

  const {
    autoLoginMessage,
    bindDingTalkUser,
    bindForm,
    bindState,
    handleSessionDataError,
    loadBootstrapSafely,
    loading,
    login,
    loginForm,
    logout,
    ready,
    resetSessionState,
    retryDingTalkAutoLogin,
    sessionAccessDenied,
    sessionAccessDeniedMessage,
    tryDingTalkAutoLogin
  } = usePerformanceAuth({
    state,
    selectedReviewId,
    resetRouteTabState: () => resetRouteTabState(),
    resetProfileDialog: () => resetProfileDialog(),
    closeCreateReviewDialog: () => closeCreateReviewDialog(),
    ensureActiveMenu: () => ensureActiveMenu(),
    clearContentLoading: () => {
      contentLoading.value = false
      contentLoadingSeq += 1
    },
    refreshDataSafely: (options) => refreshDataSafely(options),
    applyReviewDeepLinkIfNeeded: () => applyReviewDeepLinkIfNeeded(),
    toast: (title) => toast(title)
  })

  function statusText(status) {
    return statusMeta[status]?.label || status
  }

  function latestReviewAction(review) {
    const latestAction = String(review?.latestAction || '').trim()
    if (latestAction) return latestAction
    const histories = Array.isArray(review?.history) ? review.history : []
    return String(histories[histories.length - 1]?.action || '').trim()
  }

  function returnStatusTextFromAction(action) {
    const title = String(action || '').split(/[：:]/)[0].trim()
    if (title.startsWith('退回员工')) return '退回员工'
    if (title.startsWith('退回上级')) return '退回上级'
    if (title.startsWith('退回 HRBP') || title.startsWith('退回HRBP')) return '退回HRBP'
    return ''
  }

  function reviewReturnStatusText(review) {
    return returnStatusTextFromAction(latestReviewAction(review))
  }

  function reviewStatusText(review) {
    return reviewReturnStatusText(review) || statusText(review?.status)
  }

  function sameUserId(left, right) {
    const leftText = String(left || '').trim()
    const rightText = String(right || '').trim()
    return Boolean(leftText && rightText && leftText === rightText)
  }

  function userName(id) {
    return reviewPersonNameFromReviews(id) || state.users.find((item) => sameUserId(item.id, id))?.name || id || '无'
  }

  function reviewPersonName(review, id) {
    if (!review || !id) return ''
    const pairs = [
      [review.employeeId, review.employeeName],
      [review.managerId, review.managerName],
      [review.hrbpId, review.hrbpName],
      [review.hrbpReviewerId, review.hrbpReviewerName]
    ]
    const found = pairs.find(([account, name]) => sameUserId(account, id) && firstText(name))
    return found ? firstText(found[1]) : ''
  }

  function reviewPersonNameFromReviews(id) {
    if (!id) return ''
    for (const review of state.reviews) {
      const name = reviewPersonName(review, id)
      if (name) return name
    }
    return sameUserId(state.user?.id, id) ? firstText(state.user.name) : ''
  }

  function userOptionText(user) {
    return [user.name, user.position, user.department].filter(Boolean).join(' · ')
  }

  function queueReviews() {
    return state.reviews.filter((item) => canSelf(item) || canManager(item) || canHrbpHandle(item) || canEmployeeConfirm(item) || canFinal(item))
  }

  function myPerformanceReviews() {
    return state.reviews.filter((item) => sameUserId(item.employeeId, state.user?.id) && myPerformanceStatusSet.has(item.status))
  }

  function queryParam(name) {
    if (typeof window === 'undefined') return ''
    const searchParams = new URLSearchParams(window.location.search || '')
    const value = searchParams.get(name)
    if (value) return value
    const hash = window.location.hash || ''
    const queryIndex = hash.indexOf('?')
    if (queryIndex < 0) return ''
    return new URLSearchParams(hash.slice(queryIndex + 1)).get(name) || ''
  }

  function initialReviewDeepLink() {
    const reviewNo = firstText(queryParam('reviewNo'), queryParam('review_no'), queryParam('reviewId'), queryParam('id'))
    if (!reviewNo) return null
    return {
      reviewNo,
      view: normalizeReviewDeepLinkView(firstText(queryParam('view'), queryParam('menu'))),
      reviewTab: normalizeReviewDeepLinkTab(firstText(queryParam('reviewTab'), queryParam('tab'))),
      period: queryParam('period')
    }
  }

  async function applyReviewDeepLinkIfNeeded() {
    if (reviewDeepLinkApplied) return false
    const deepLink = initialReviewDeepLink()
    if (!deepLink) return false
    reviewDeepLinkApplied = true
    try {
      const res = await fetchReviewDetail(deepLink.reviewNo)
      const detail = res.data || {}
      if (!detail?.id) return false
      await openWorkbenchTodo(detail, { preferredView: deepLink.view, reviewTab: deepLink.reviewTab, replaceRouteTabs: true })
      return true
    } catch (error) {
      handleSessionDataError(error)
      return false
    }
  }

  async function refreshDataSafely(options = {}) {
    const useContentLoading = Boolean(options.contentLoading)
    const loadingSeq = useContentLoading ? ++contentLoadingSeq : 0
    if (useContentLoading) {
      contentLoading.value = true
    }
    try {
      await refreshData(options)
      return true
    } catch (error) {
      return handleSessionDataError(error)
    } finally {
      if (useContentLoading && loadingSeq === contentLoadingSeq) {
        contentLoading.value = false
      }
    }
  }

  async function refreshSessionAndDataSafely() {
    const bootstrapped = await loadBootstrapSafely()
    if (!bootstrapped) return false
    return refreshDataSafely({ forceReference: true, contentLoading: true })
  }

  async function refreshWithUserFeedback() {
    if (refreshing.value) return false
    refreshing.value = true
    showRefreshLoading()
    try {
      const refreshed = await refreshSessionAndDataSafely()
      hideRefreshLoading()
      if (refreshed) {
        toast('已刷新')
      } else {
        toast('刷新失败，请稍后重试')
      }
      return refreshed
    } catch (error) {
      handleSessionDataError(error)
      hideRefreshLoading()
      toast('刷新失败，请稍后重试')
      return false
    } finally {
      refreshing.value = false
    }
  }

  function effectiveGrade(review) {
    return review.finalGrade || review.hrbpGrade || ''
  }

  function objectiveScore(item) {
    const completion = Number(item.completion)
    const weight = Number(item.weight)
    if (!Number.isFinite(completion) || !Number.isFinite(weight)) return 0
    return Math.round((weight * completion / 100) * 10) / 10
  }

  function totalObjectiveScore(review) {
    const total = (review.objectives || []).reduce((sum, item) => sum + objectiveScore(item), 0)
    return Math.round(total * 10) / 10
  }

  function valueTotal(review, field) {
    const nums = (review.values || []).map((item) => Number(item[field])).filter(Number.isFinite)
    if (!nums.length) return '-'
    return Math.round(nums.reduce((sum, item) => sum + item, 0) * 10) / 10
  }

  function gradeOptions() {
    return [''].concat(grades.value)
  }

  function statusClass(status) {
    return `chip chip-${statusMeta[status]?.tone || 'blue'}`
  }

  function reviewStatusClass(review) {
    return reviewReturnStatusText(review) ? 'chip chip-danger' : statusClass(review?.status)
  }

  function resetFilters() {
    Object.assign(summaryFilters, { employeeName: '', departmentName: '', departmentNames: [], period: '', status: '' })
  }

  function summaryDepartmentMatches(department, departmentName, departmentNames) {
    const departmentText = String(department || '').toLowerCase()
    const selectedDepartments = Array.isArray(departmentNames) ? departmentNames : []
    if (selectedDepartments.length > 0) {
      return selectedDepartments.some((name) => {
        const text = String(name || '').trim().toLowerCase()
        return text && departmentText.includes(text)
      })
    }
    const keyword = String(departmentName || '').trim().toLowerCase()
    return !keyword || departmentText.includes(keyword)
  }

  function toast(title) {
    uni.showToast({ title, icon: 'none' })
  }

  function showRefreshLoading() {
    if (typeof uni !== 'undefined' && uni.showLoading) {
      uni.showLoading({ title: '刷新中...', mask: true })
    }
  }

  function hideRefreshLoading() {
    if (typeof uni !== 'undefined' && uni.hideLoading) {
      uni.hideLoading()
    }
  }

  providePerformanceContext({
    canEditObjectiveDimension,
    canAddNextObjective,
    canDeleteNextObjective,
    canEditNextObjectives,
    canEmployeeConfirm,
    canCreateReview,
    canDeleteReview,
    canEditUsers,
    canEditTemplate,
    canExportReviews,
    canFinal,
    canHrbpHandle,
    canManager,
    canSelf,
    canWithdraw,
    contentLoading,
    contentView,
    createReview,
    createReviewDialog,
    createReviewForm,
    createReviewTargetUsers,
    createTargetUserTree,
    currentReviews,
    dashboardFilter,
    deleteReview,
    deleteUser,
    departments,
    effectiveGrade,
    exportSummary,
    gradeOptions,
    grades,
    hasApiPermission,
    hasButtonPermission,
    hasMenuPermission,
    historyFilters,
    historyMonthOptions,
    historyYearOptions,
    hrbpReviewTab,
    hrbpReviewedFilters,
    hrbpIds,
    refreshData: refreshWithUserFeedback,
    managerReviewTab,
    managerReviewedFilters,
    managerIds,
    newReview,
    objectiveScore,
    pageTitle,
    performReviewAction,
    openCreateReviewDialog,
    openWorkbenchTodo,
    queueReviews,
    resetFilters,
    returnReview,
    reviewTab,
    reviewTargetUsers,
    resetHrbpReviewedFilters,
    resetHistoryFilters,
    resetManagerReviewedFilters,
    saveTemplate,
    saveUser,
    selectReview,
    selectedReview,
    selectedReviewId,
    sectionTitle,
    statCards,
    state,
    statusClass,
    statusMeta,
    statusText,
    reviewStatusClass,
    reviewStatusText,
    returnStatusTextFromAction,
    searchHrbpReviewedReviews,
    searchManagerReviewedReviews,
    setHrbpReviewedFilter,
    setHistoryFilter,
    setManagerReviewedFilter,
    switchHrbpReviewTab,
    switchManagerReviewTab,
    summaryFilters,
    summaryReviews,
    totalObjectiveScore,
    userName,
    userOptionText,
    valueTotal,
    workbenchCards,
    withdrawReview
  })

  onMounted(async () => {
    uni.$on(AUTH_EXPIRED_EVENT, resetSessionState)
    if (!authToken()) {
      await tryDingTalkAutoLogin()
      ready.value = true
      return
    }
    const bootstrapped = await loadBootstrapSafely()
    if (bootstrapped) {
      if (!(await applyReviewDeepLinkIfNeeded())) {
        await refreshDataSafely({ contentLoading: true })
      }
    } else if (!authToken()) {
      await tryDingTalkAutoLogin()
    }
    ready.value = true
  })

  onUnmounted(() => {
    uni.$off(AUTH_EXPIRED_EVENT, resetSessionState)
  })

  return {
    ready,
    state,
    bindState,
    bindForm,
    loading,
    appConfig,
    sessionAccessDenied,
    sessionAccessDeniedMessage,
    loginForm,
    autoLoginMessage,
    bindDingTalkUser,
    retryDingTalkAutoLogin,
    resetSessionState,
    login,
    activeView,
    activePerformanceTab,
    appTitle,
    navItems,
    pageTitle,
    routeTabs,
    activateRouteTab,
    closeRouteTab,
    logout,
    openProfileDialog,
    switchView,
    contentView,
    contentLoading,
    sectionTitle,
    profileDialog,
    profileAvatarPreview,
    profileDisplayName,
    profileInitial,
    chooseProfileAvatar,
    clearProfileAvatar,
    closeProfileDialog,
    submitProfileDialog,
    createReviewDialog,
    createReviewForm,
    createReviewUserKeyword,
    createTargetUserTree,
    createTargetUserTreeRows,
    createTargetUserEmptyText,
    createTargetDepartmentCheckState,
    createTargetDepartmentUserIds,
    createTargetUserMeta,
    createReviewMonthText,
    createReviewMonthPickerOpen,
    createReviewMonthPickerYear,
    createReviewMonthOptions,
    isCreateReviewMonthSelected,
    closeCreateReviewDialog,
    toggleCreateReviewDept,
    toggleCreateReviewDepartment,
    toggleCreateReviewEmployee,
    toggleCreateReviewMonthPicker,
    changeCreateReviewMonthPickerYear,
    selectCreateReviewMonth,
    createReview,
    withdrawDialog,
    withdrawReasonLength,
    closeWithdrawDialog,
    submitWithdrawReview,
    returnDialog,
    returnReasonLength,
    closeReturnDialog,
    submitReturnReview,
    disputeDialog,
    disputeReasonLength,
    closeDisputeDialog,
    submitDisputeReview
  }
}
