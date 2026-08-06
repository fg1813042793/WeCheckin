import { reactive } from 'vue'
import { listReviews } from '../../../../api/performance/common/reviews'
import { myPerformanceStatuses } from '../constants'

export function usePerformanceReviewList({
  state,
  contentView,
  selectedReviewId,
  managerReviewTab,
  hrbpReviewTab,
  ensureSelectedReview,
  refreshDataSafely
}) {
  const historyFilters = reactive({
    year: '',
    month: '',
    grade: ''
  })
  const managerReviewedFilters = reactive({
    employeeName: '',
    period: '',
    objectiveScore: ''
  })
  const hrbpReviewedFilters = reactive({
    employeeName: '',
    period: '',
    grade: ''
  })

  async function loadReviews(params = {}, options = {}) {
    const res = await listReviews({ ...reviewQueryParamsForContentView(), ...params })
    const payload = res.data || {}
    if (Array.isArray(payload)) {
      state.reviews = payload
      state.reviewListTotal = payload.length
      state.reviewPage = 1
      state.reviewPageSize = payload.length || 20
    } else {
      state.reviews = Array.isArray(payload.list) ? payload.list : []
      state.reviewListTotal = Number(payload.total || state.reviews.length)
      state.reviewPage = Number(payload.page || 1)
      state.reviewPageSize = Number(payload.pageSize || 20)
    }
    ensureSelectedReview(options)
  }

  function reviewQueryParamsForContentView() {
    const view = contentView.value
    const params = {
      scope: reviewScopeForContentView(),
      skipHistory: view === 'mine' ? 0 : 1
    }
    if (view === 'manager') {
      Object.assign(params, reviewStatusParamsForTab(managerReviewTab.value, ['manager_review'], ['hrbp_review', 'employee_confirm', 'hr_final', 'completed']))
      if (managerReviewTab.value === 'reviewed') {
        Object.assign(params, managerReviewedFilterParams())
      }
    } else if (view === 'hrbp') {
      Object.assign(params, reviewStatusParamsForTab(hrbpReviewTab.value, ['hrbp_review'], ['employee_confirm', 'hr_final', 'completed']))
      if (hrbpReviewTab.value === 'reviewed') {
        Object.assign(params, hrbpReviewedFilterParams())
      }
    } else if (view === 'mine') {
      params.statuses = myPerformanceStatuses.join(',')
    } else if (view === 'history') {
      params.status = 'completed'
      Object.assign(params, reviewHistoryFilterParams())
    } else if (view === 'summary') {
      params.status = 'completed'
    }
    return params
  }

  function reviewScopeForContentView() {
    if (contentView.value === 'history') return 'mine'
    if (contentView.value === 'mine') return 'mine'
    if (contentView.value === 'hrbp') return 'hrbp'
    if (contentView.value === 'summary') return 'summary'
    return contentView.value || 'mine'
  }

  function managerReviewedFilterParams() {
    const employeeName = String(managerReviewedFilters.employeeName || '').trim()
    const period = String(managerReviewedFilters.period || '').trim()
    const objectiveScore = String(managerReviewedFilters.objectiveScore || '').trim()
    const params = {}
    if (employeeName) params.employeeName = employeeName
    if (period) params.period = period
    if (objectiveScore) params.objectiveScore = objectiveScore
    return params
  }

  function hrbpReviewedFilterParams() {
    const employeeName = String(hrbpReviewedFilters.employeeName || '').trim()
    const period = String(hrbpReviewedFilters.period || '').trim()
    const grade = String(hrbpReviewedFilters.grade || '').trim()
    const params = {}
    if (employeeName) params.employeeName = employeeName
    if (period) params.period = period
    if (grade) params.grade = grade
    return params
  }

  function reviewHistoryFilterParams() {
    const year = String(historyFilters.year || '').trim()
    const month = String(historyFilters.month || '').trim()
    const grade = String(historyFilters.grade || '').trim()
    const params = {}
    if (year && month) {
      params.period = `${year}-${month}`
    } else {
      if (year) params.year = year
      if (month) params.month = month
    }
    if (grade) params.grade = grade
    return params
  }

  function reviewStatusParamsForTab(tab, pendingStatuses, reviewedStatuses) {
    const statuses = tab === 'reviewed' ? reviewedStatuses : pendingStatuses
    if (statuses.length === 1) return { status: statuses[0] }
    return { statuses: statuses.join(',') }
  }

  async function switchManagerReviewTab(tab) {
    if (managerReviewTab.value === tab) return
    managerReviewTab.value = tab
    selectedReviewId.value = ''
    await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
  }

  async function switchHrbpReviewTab(tab) {
    if (hrbpReviewTab.value === tab) return
    hrbpReviewTab.value = tab
    selectedReviewId.value = ''
    await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
  }

  async function setHistoryFilter(field, value) {
    if (!Object.prototype.hasOwnProperty.call(historyFilters, field)) return
    const nextValue = String(value || '').trim()
    if (historyFilters[field] === nextValue) return
    historyFilters[field] = nextValue
    selectedReviewId.value = ''
    await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
  }

  async function resetHistoryFilters() {
    if (!historyFilters.year && !historyFilters.month && !historyFilters.grade) return
    Object.assign(historyFilters, { year: '', month: '', grade: '' })
    selectedReviewId.value = ''
    await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
  }

  function setManagerReviewedFilter(field, value) {
    if (!Object.prototype.hasOwnProperty.call(managerReviewedFilters, field)) return
    managerReviewedFilters[field] = String(value || '').trim()
  }

  async function searchManagerReviewedReviews() {
    selectedReviewId.value = ''
    await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
  }

  async function resetManagerReviewedFilters() {
    if (!managerReviewedFilters.employeeName && !managerReviewedFilters.period && !managerReviewedFilters.objectiveScore) return
    Object.assign(managerReviewedFilters, { employeeName: '', period: '', objectiveScore: '' })
    selectedReviewId.value = ''
    await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
  }

  function setHrbpReviewedFilter(field, value) {
    if (!Object.prototype.hasOwnProperty.call(hrbpReviewedFilters, field)) return
    hrbpReviewedFilters[field] = String(value || '').trim()
  }

  async function searchHrbpReviewedReviews() {
    selectedReviewId.value = ''
    await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
  }

  async function resetHrbpReviewedFilters() {
    if (!hrbpReviewedFilters.employeeName && !hrbpReviewedFilters.period && !hrbpReviewedFilters.grade) return
    Object.assign(hrbpReviewedFilters, { employeeName: '', period: '', grade: '' })
    selectedReviewId.value = ''
    await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
  }

  return {
    historyFilters,
    managerReviewedFilters,
    hrbpReviewedFilters,
    loadReviews,
    reviewQueryParamsForContentView,
    reviewScopeForContentView,
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
  }
}
