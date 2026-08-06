import { deleteReview as deletePerformanceReview, exportReviewsUrl } from '../../../api/performance/common/reviews'
import {
  deleteUser as deletePerformanceUser,
  listUsers,
  updateUser as updatePerformanceUser
} from '../../../api/performance/flow-config'
import { getTemplate, saveTemplate as savePerformanceTemplate } from '../../../api/performance/parameters'

export function usePerformanceData({
  state,
  contentView,
  currentReviews,
  navItems,
  selectedReviewId,
  summaryFilters,
  ensureActiveMenu,
  ensureNewReviewEmployee,
  loadReviews,
  hasApiPermission,
  canDeleteReview,
  canEditTemplate,
  canExportReviews,
  confirmReviewAction,
  toast
}) {
  function sanitizeUsers(users) {
    return (users || []).map((item) => ({ ...item, password: '' }))
  }

  function upsertUser(user) {
    const sanitized = sanitizeUsers([user])[0]
    if (!sanitized?.id) return
    const index = state.users.findIndex((item) => item.id === sanitized.id)
    if (index >= 0) {
      state.users[index] = sanitized
    } else {
      state.users.push(sanitized)
    }
  }

  function shouldAutoSelectReview(options = {}) {
    if (options.autoSelectReview === false) return false
    return contentView.value === 'mine'
  }

  function ensureSelectedReview(options = {}) {
    ensureNewReviewEmployee()
    if (!shouldAutoSelectReview(options)) {
      if (selectedReviewId.value && !state.reviews.some((item) => item.id === selectedReviewId.value)) {
        selectedReviewId.value = ''
      }
      return
    }
    if (!selectedReviewId.value || !currentReviews.value.some((item) => item.id === selectedReviewId.value)) {
      selectedReviewId.value = currentReviews.value[0]?.id || ''
    }
  }

  async function loadUsers() {
    if (!hasApiPermission('dingtalk_h5:api:user:list')) {
      state.users = []
      ensureSelectedReview()
      return
    }
    const res = await listUsers()
    state.users = sanitizeUsers(res.data || [])
    ensureSelectedReview()
  }

  async function loadTemplate() {
    if (!hasApiPermission('dingtalk_h5:api:template:view')) {
      state.template = null
      return
    }
    const res = await getTemplate()
    state.template = res.data
  }

  async function saveTemplate(data) {
    if (!canEditTemplate()) {
      toast('无权限保存模板')
      return
    }
    const res = await savePerformanceTemplate(data)
    state.template = res.data
    toast('模板已保存')
  }

  function needsUserDirectoryForContentView() {
    return contentView.value === 'summary'
  }

  async function ensureReferenceData(options = {}) {
    const force = Boolean(options.force)
    const needsUsers = Boolean(options.users)
    const needsTemplate = options.template !== false
    const tasks = []
    if (needsUsers && (force || !state.users.length) && hasApiPermission('dingtalk_h5:api:user:list')) tasks.push(loadUsers())
    if (needsTemplate && (force || !state.template) && hasApiPermission('dingtalk_h5:api:template:view')) tasks.push(loadTemplate())
    if (tasks.length > 0) {
      await Promise.all(tasks)
    }
  }

  async function refreshData(options = {}) {
    if (!state.user) return
    if (navItems.value.length === 0) return
    const forceReference = Boolean(options.forceReference)
    ensureActiveMenu()
    if (state.view === 'dashboard') {
      await loadReviews({ pageSize: 100 }, { autoSelectReview: false })
      return
    }
    if (contentView.value === 'org') {
      await loadUsers()
      return
    }
    if (contentView.value === 'template') {
      await loadTemplate()
      return
    }
    await loadReviews({}, options)
    await ensureReferenceData({ force: forceReference, users: needsUserDirectoryForContentView(), template: true })
  }

  function updateReview(review) {
    if (!review?.id) return
    const index = state.reviews.findIndex((item) => item.id === review.id)
    if (index >= 0) {
      state.reviews[index] = review
    } else {
      state.reviews.push(review)
    }
    selectedReviewId.value = review.id
  }

  async function deleteReview(id) {
    if (!canDeleteReview()) {
      toast('无权限删除考评单')
      return
    }
    const confirmed = await confirmReviewAction('delete-review')
    if (!confirmed) return
    try {
      await deletePerformanceReview(id)
      await loadReviews()
      toast('已删除')
    } catch (error) {
      toast(error?.msg || '删除失败')
    }
  }

  async function exportSummary() {
    if (!canExportReviews()) {
      toast('无权限导出')
      return
    }
    window.location.href = exportReviewsUrl({ scope: 'summary', ...summaryFilters, status: 'completed' })
  }

  async function saveUser(user) {
    if (!hasApiPermission('dingtalk_h5:api:user:edit')) {
      toast('无权限保存人员')
      return
    }
    const res = await updatePerformanceUser(user.id, normalizeUserPayload(user))
    if (res.data?.user) {
      upsertUser(res.data.user)
    } else if (Array.isArray(res.data?.users)) {
      state.users = sanitizeUsers(res.data.users)
    }
    ensureSelectedReview()
    toast('人员已保存')
  }

  async function deleteUser(id) {
    if (!hasApiPermission('dingtalk_h5:api:user:delete')) {
      toast('无权限删除人员')
      return
    }
    if (!window.confirm(`确认删除账号 ${id}？`)) return
    const res = await deletePerformanceUser(id)
    state.users = sanitizeUsers(res.data.users || [])
    ensureSelectedReview()
    toast('人员已删除')
  }

  function normalizeUserPayload(user) {
    return {
      id: user.id,
      name: user.name,
      password: user.password || '',
      position: user.position || '',
      departmentLevel1: user.departmentLevel1 || '',
      departmentLevel2: user.departmentLevel2 || '',
      departmentLevel3: user.departmentLevel3 || '',
      managerId: user.managerId || '',
      hrbpId: user.hrbpId || ''
    }
  }

  return {
    deleteReview,
    deleteUser,
    ensureReferenceData,
    ensureSelectedReview,
    exportSummary,
    loadTemplate,
    loadUsers,
    refreshData,
    sanitizeUsers,
    saveTemplate,
    saveUser,
    updateReview,
    upsertUser
  }
}
