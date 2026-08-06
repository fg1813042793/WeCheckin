import { computed, reactive, ref } from 'vue'
import { createReview as createPerformanceReview } from '../../../../api/performance/mine'
import {
  buildCreateTargetUserTree,
  createTargetUserMatchesKeyword,
  currentMonth,
  flattenCreateTargetTree,
  nextMonthFromPeriod,
  normalizeSearchKeyword
} from '../../common/helpers'

export function usePerformanceReviewCreation({
  state,
  selectedReviewId,
  reviewTargetUsers,
  loadUsers,
  loadReviews,
  hasApiPermission,
  canCreateReview,
  toast
}) {
  const newReview = reactive({ employeeId: '' })
  const createReviewDialog = reactive({ visible: false, loading: false })
  const createReviewForm = reactive({ employeeIds: [], period: currentMonth() })
  const createReviewUserKeyword = ref('')
  const createReviewExpandedDeptKeys = ref(new Set())
  const createReviewMonthPickerOpen = ref(false)
  const createReviewMonthPickerYear = ref(new Date().getFullYear())

  const createReviewTargetUsers = computed(() => {
    if (reviewTargetUsers.value.length > 0) return reviewTargetUsers.value
    return state.user?.id ? [state.user] : []
  })
  const createReviewSearchKeyword = computed(() => normalizeSearchKeyword(createReviewUserKeyword.value))
  const filteredCreateReviewTargetUsers = computed(() => {
    const keyword = createReviewSearchKeyword.value
    if (!keyword) return createReviewTargetUsers.value
    return createReviewTargetUsers.value.filter((user) => createTargetUserMatchesKeyword(user, keyword))
  })
  const createTargetUserTree = computed(() => buildCreateTargetUserTree(filteredCreateReviewTargetUsers.value))
  const createTargetUserTreeRows = computed(() => flattenCreateTargetTree(
    createTargetUserTree.value,
    createReviewExpandedDeptKeys.value,
    1,
    Boolean(createReviewSearchKeyword.value)
  ))
  const createTargetUserEmptyText = computed(() => {
    if (createReviewTargetUsers.value.length === 0) return '暂无可创建人员'
    return createReviewSearchKeyword.value ? '没有匹配的人员' : '暂无可创建人员'
  })
  const createReviewMonthOptions = computed(() => Array.from({ length: 12 }, (_, index) => ({
    value: index + 1,
    label: `${String(index + 1).padStart(2, '0')}月`
  })))

  function ensureNewReviewEmployee() {
    if (!newReview.employeeId) {
      newReview.employeeId = reviewTargetUsers.value[0]?.id || ''
    }
  }

  async function openCreateReviewDialog() {
    if (!canCreateReview()) {
      toast('无权限创建考评单')
      return
    }
    createReviewForm.period = currentMonth()
    createReviewForm.employeeIds = []
    createReviewUserKeyword.value = ''
    createReviewExpandedDeptKeys.value = new Set()
    syncCreateReviewMonthPickerYear()
    createReviewMonthPickerOpen.value = false
    if (hasApiPermission('dingtalk_h5:api:user:list') && state.users.length === 0) {
      await loadUsers()
    }
    normalizeCreateReviewSelection()
    createReviewDialog.visible = true
  }

  function closeCreateReviewDialog() {
    createReviewDialog.visible = false
    createReviewDialog.loading = false
    createReviewUserKeyword.value = ''
    createReviewMonthPickerOpen.value = false
  }

  function createReviewPeriodParts(period) {
    const match = String(period || '').match(/^(\d{4})-(\d{1,2})$/)
    if (!match) return null
    const year = Number(match[1])
    const month = Number(match[2])
    if (!year || month < 1 || month > 12) return null
    return { year, month }
  }

  function createReviewMonthText(period) {
    const parts = createReviewPeriodParts(period)
    if (!parts) return '请选择月份'
    return `${parts.year}-${String(parts.month).padStart(2, '0')}`
  }

  function syncCreateReviewMonthPickerYear() {
    const parts = createReviewPeriodParts(createReviewForm.period)
    createReviewMonthPickerYear.value = parts?.year || new Date().getFullYear()
  }

  function toggleCreateReviewMonthPicker() {
    if (!createReviewMonthPickerOpen.value) {
      syncCreateReviewMonthPickerYear()
    }
    createReviewMonthPickerOpen.value = !createReviewMonthPickerOpen.value
  }

  function changeCreateReviewMonthPickerYear(delta) {
    createReviewMonthPickerYear.value += Number(delta || 0)
  }

  function selectCreateReviewMonth(month) {
    createReviewForm.period = `${createReviewMonthPickerYear.value}-${String(month).padStart(2, '0')}`
    createReviewMonthPickerOpen.value = false
  }

  function isCreateReviewMonthSelected(month) {
    const parts = createReviewPeriodParts(createReviewForm.period)
    return Boolean(parts && parts.year === createReviewMonthPickerYear.value && parts.month === month)
  }

  function normalizeCreateReviewSelection() {
    const targetIds = new Set(createReviewTargetUsers.value.map((item) => item.id).filter(Boolean))
    let selected = createReviewForm.employeeIds.filter((id) => targetIds.has(id))
    if (selected.length === 0) {
      const currentID = state.user?.id
      if (currentID && targetIds.has(currentID)) {
        selected = [currentID]
      } else if (createReviewTargetUsers.value[0]?.id) {
        selected = [createReviewTargetUsers.value[0].id]
      }
    }
    createReviewForm.employeeIds = selected
    newReview.employeeId = selected[0] || ''
  }

  function setCreateReviewEmployeeIds(ids = []) {
    const selected = new Set(ids.map((id) => String(id || '').trim()).filter(Boolean))
    createReviewForm.employeeIds = createReviewTargetUsers.value
      .map((item) => item.id)
      .filter((id) => selected.has(id))
    newReview.employeeId = createReviewForm.employeeIds[0] || ''
  }

  function createTargetDepartmentUserIds(row) {
    return Array.isArray(row?.userIds) ? row.userIds.filter(Boolean) : []
  }

  function createTargetDepartmentCheckState(row) {
    const ids = createTargetDepartmentUserIds(row)
    if (ids.length === 0) return 'empty'
    const selected = new Set(createReviewForm.employeeIds)
    const selectedCount = ids.filter((id) => selected.has(id)).length
    if (selectedCount === 0) return 'unchecked'
    return selectedCount === ids.length ? 'checked' : 'indeterminate'
  }

  function toggleCreateReviewDepartment(row) {
    const ids = createTargetDepartmentUserIds(row)
    if (ids.length === 0) return
    const selected = new Set(createReviewForm.employeeIds)
    const allSelected = ids.every((id) => selected.has(id))
    ids.forEach((id) => {
      if (allSelected) {
        selected.delete(id)
      } else {
        selected.add(id)
      }
    })
    setCreateReviewEmployeeIds([...selected])
  }

  function toggleCreateReviewEmployee(id) {
    id = String(id || '').trim()
    if (!id) return
    if (createReviewForm.employeeIds.includes(id)) {
      setCreateReviewEmployeeIds(createReviewForm.employeeIds.filter((item) => item !== id))
      return
    }
    setCreateReviewEmployeeIds([...createReviewForm.employeeIds, id])
  }

  function toggleCreateReviewDept(key) {
    key = String(key || '').trim()
    if (!key) return
    const next = new Set(createReviewExpandedDeptKeys.value)
    if (next.has(key)) {
      next.delete(key)
    } else {
      next.add(key)
    }
    createReviewExpandedDeptKeys.value = next
  }

  async function createReview() {
    if (!canCreateReview()) {
      toast('无权限创建考评单')
      return
    }
    normalizeCreateReviewSelection()
    if (createReviewForm.employeeIds.length === 0) {
      toast('请选择被考评人')
      return
    }
    createReviewDialog.loading = true
    try {
      const res = await createPerformanceReview({
        employeeIds: createReviewForm.employeeIds,
        period: createReviewForm.period,
        nextPeriod: nextMonthFromPeriod(createReviewForm.period)
      })
      const data = res.data || {}
      const created = Array.isArray(data.list) ? data.list : (data.id ? [data] : [])
      const failed = Array.isArray(data.failed) ? data.failed : []
      await loadReviews()
      if (created[0]?.id) {
        selectedReviewId.value = created[0].id
      }
      closeCreateReviewDialog()
      toast(failed.length > 0 ? `已创建 ${created.length} 张，${failed.length} 张失败` : `已创建 ${created.length || 1} 张`)
    } catch (error) {
      toast(error?.msg || '创建失败')
    } finally {
      createReviewDialog.loading = false
    }
  }

  return {
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
  }
}
