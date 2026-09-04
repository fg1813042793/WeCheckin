<script setup lang="ts">
import type { PerformanceReview, PerformanceReviewListPayload, PerformanceTemplate, PerformanceUser, ReviewActionRequest } from '@/types/dingtalk-h5'
import { computed, reactive, ref, watch } from 'vue'
import { deleteReview, exportReviewsUrl, getTemplate, listReviews, listUsers, reviewAction } from '@/api/dingtalk-h5'
import { useDingtalkAuthStore } from '@/stores'
import { useAppContentStore } from '@/stores/appContent'
import { departmentLevelsFromEntity, departmentPathFromEntity } from '@/utils/departments'
import PerformanceDetailPopup from './components/PerformanceDetailPopup.vue'
import { appendUniqueReviewRows, createMobilePaginationState, mobileReviewListParams, resetMobilePagination, showMobileLoadMore, showMobileNoMore, updateMobilePaginationTotal } from './components/mobilePagination'
import { normalizeReviewActionResult, statusMeta } from './constants/performanceStatus'

interface PaginationChangePayload {
  current?: number
}

interface SummaryDepartmentNode {
  key: string
  name: string
  path: string
  count: number
  childMap?: Map<string, SummaryDepartmentNode>
  children: SummaryDepartmentNode[]
}

interface SummaryDepartmentRow {
  key: string
  name: string
  path: string
  count: number
  depth: number
  expandable: boolean
  expanded: boolean
}

interface SummaryPeriodMonth {
  value: string
  label: string
  available: boolean
  selected: boolean
}

const TABLE_PAGE_SIZE = 10

function resolveMobilePage() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    const deviceType = String(info.deviceType || '').toLowerCase()
    const platform = String(info.platform || '').toLowerCase()
    return (width > 0 && width <= 768) || deviceType === 'phone' || (['android', 'ios'].includes(platform) && width <= 1024)
  }
  catch {
    return false
  }
}

function paginateRows<T>(rows: T[], page: number) {
  const start = (Math.max(page, 1) - 1) * TABLE_PAGE_SIZE
  return rows.slice(start, start + TABLE_PAGE_SIZE)
}

function paginationTotalPages(total: number) {
  return Math.max(Math.ceil(total / TABLE_PAGE_SIZE), 1)
}

function sameUserId(left: unknown, right: unknown) {
  const leftText = String(left || '').trim()
  const rightText = String(right || '').trim()
  return Boolean(leftText && rightText && leftText === rightText)
}

function listFromPayload(payload: PerformanceReviewListPayload | PerformanceReview[] | undefined) {
  if (Array.isArray(payload)) {
    return payload
  }
  return payload?.list || payload?.rows || payload?.items || []
}

function totalFromPayload(payload: PerformanceReviewListPayload | PerformanceReview[] | undefined, fallback: number) {
  if (Array.isArray(payload)) {
    return payload.length
  }
  return Number(payload?.total ?? fallback)
}

function firstText(...values: Array<unknown>) {
  for (const value of values) {
    const text = String(value ?? '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

function reviewKey(review: PerformanceReview) {
  return review.id || review.reviewNo || `${review.employeeId || 'employee'}-${review.period || 'period'}`
}

function reviewStatusText(review: PerformanceReview) {
  return statusMeta[String(review.status || '')]?.label || String(review.status || '未知')
}

function resolveEmployeeName(review: PerformanceReview, resolveUserName: (id: unknown) => string) {
  return firstText(review.employeeName, resolveUserName(review.employeeId), review.employeeId) || '-'
}

function resolveReviewPosition(review: PerformanceReview, reviewUsers: PerformanceUser[] = []) {
  return firstText(review.position, reviewUsers.find(user => user.id === review.employeeId)?.position) || '-'
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === 'object' && !Array.isArray(value))
}

function roundScore(value: number) {
  return Math.round(value * 10) / 10
}

function reviewObjectiveItems(review: PerformanceReview) {
  return (Array.isArray(review.objectives) ? review.objectives : []).filter(isRecord)
}

function objectiveItemScore(item: Record<string, unknown>) {
  const completion = Number(item.completion)
  const weight = Number(item.weight)
  if (!Number.isFinite(completion) || !Number.isFinite(weight)) {
    return 0
  }
  return roundScore(weight * completion / 100)
}

function totalObjectiveScore(review: PerformanceReview) {
  const objectives = reviewObjectiveItems(review)
  if (!objectives.length) {
    return ''
  }
  return roundScore(objectives.reduce((sum, item) => sum + objectiveItemScore(item), 0))
}

function objectiveScore(review: PerformanceReview) {
  return firstText(review.objectiveScore, totalObjectiveScore(review), review.managerScore) || '-'
}

function reviewDepartmentLevels(review: PerformanceReview) {
  const levels = departmentLevelsFromEntity(review)
  return levels.length > 0 ? levels : ['未设置部门']
}

function reviewDepartmentPath(review: PerformanceReview) {
  return reviewDepartmentLevels(review).join(' / ')
}

function reviewDepartmentText(review: PerformanceReview) {
  return departmentPathFromEntity(review)
}

function ensureSummaryDepartmentNode(map: Map<string, SummaryDepartmentNode>, key: string, name: string, path: string) {
  if (!map.has(key)) {
    map.set(key, {
      key,
      name,
      path,
      count: 0,
      childMap: new Map(),
      children: [],
    })
  }
  return map.get(key) as SummaryDepartmentNode
}

function finalizeSummaryDepartmentNodes(nodes: SummaryDepartmentNode[]): SummaryDepartmentNode[] {
  return nodes
    .sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))
    .map(node => ({
      key: node.key,
      name: node.name,
      path: node.path,
      count: node.count,
      children: finalizeSummaryDepartmentNodes([...(node.childMap?.values() || [])]),
    }))
}

function buildSummaryDepartmentTree(items: PerformanceReview[]) {
  const root = new Map<string, SummaryDepartmentNode>()
  for (const review of items) {
    const levels = reviewDepartmentLevels(review)
    let currentMap = root
    let currentPath = ''
    for (const [index, level] of levels.entries()) {
      currentPath = currentPath ? `${currentPath} / ${level}` : level
      const key = `dept:${currentPath}`
      const node = ensureSummaryDepartmentNode(currentMap, key, level, currentPath)
      node.count += 1
      currentMap = node.childMap as Map<string, SummaryDepartmentNode>
      if (index === levels.length - 1) {
        continue
      }
    }
  }
  return finalizeSummaryDepartmentNodes([...root.values()])
}

function flattenSummaryDepartmentTree(nodes: SummaryDepartmentNode[], expandedKeys: Set<string>, depth = 1): SummaryDepartmentRow[] {
  const rows: SummaryDepartmentRow[] = []
  for (const node of nodes) {
    const expandable = node.children.length > 0
    const expanded = expandedKeys.has(node.key)
    rows.push({
      key: node.key,
      name: node.name,
      path: node.path,
      count: node.count,
      depth,
      expandable,
      expanded,
    })
    if (expandable && expanded) {
      rows.push(...flattenSummaryDepartmentTree(node.children, expandedKeys, depth + 1))
    }
  }
  return rows
}

function periodParts(period: unknown) {
  const match = String(period || '').trim().match(/^(\d{4})-(\d{1,2})$/)
  if (!match) {
    return null
  }
  const year = Number(match[1])
  const month = Number(match[2])
  if (!year || month < 1 || month > 12) {
    return null
  }
  return { year, month }
}

function formatPeriod(year: number, month: number) {
  return `${year}-${String(month).padStart(2, '0')}`
}

function shouldShowFinalGrade(review: PerformanceReview | Record<string, unknown>) {
  const status = String(review.status || '').trim()
  const employeeConfirmResult = String(review.employeeConfirmResult || '').trim()
  if (employeeConfirmResult === 'disputed') {
    return false
  }
  return employeeConfirmResult === 'confirmed' || status === 'hr_final' || status === 'completed'
}

function effectiveGrade(review: PerformanceReview | Record<string, unknown>) {
  if (!shouldShowFinalGrade(review)) {
    return '-'
  }
  return firstText(review.finalGrade, review.hrbpGrade) || '-'
}

function employeeConfirmText(review: PerformanceReview | Record<string, unknown>) {
  const value = String(review.employeeConfirmResult || '').trim()
  const textMap: Record<string, string> = {
    confirmed: '已确认',
    confirm: '已确认',
    disputed: '有异议',
    dispute: '有异议',
    pending: '待确认',
    unconfirmed: '待确认',
  }
  return textMap[value] || value || '待确认'
}

function employeeConfirmTagType(review: PerformanceReview | Record<string, unknown>) {
  const text = employeeConfirmText(review)
  if (text === '已确认') {
    return 'success'
  }
  if (text === '有异议') {
    return 'error'
  }
  return 'warning'
}

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const reviews = ref<PerformanceReview[]>([])
const users = ref<PerformanceUser[]>([])
const template = ref<PerformanceTemplate | null>(null)
const reviewListTotal = ref(0)
const loading = ref(false)
const detailLoading = ref(false)
const isMobilePage = ref(resolveMobilePage())
const summaryPage = ref(1)
const summaryMobilePagination = createMobilePaginationState()
const summaryFiltersCollapsed = ref(true)
const selectedReviewId = ref('')
const selectedDepartmentPaths = ref<string[]>([])
const selectedPeriods = ref<string[]>([])
const summaryDepartmentExpandedKeys = ref<Set<string>>(new Set())
const summaryDepartmentPickerOpen = ref(false)
const summaryPeriodPickerOpen = ref(false)
const summaryPeriodPickerYear = ref(new Date().getFullYear())
const summaryFilters = reactive({
  employeeName: '',
})
const listTitle = 'HRBP汇总'
const listDesc = '按人员和月份查看进度、分档并导出结果'
const currentReviews = computed(() => {
  return reviews.value.filter(review => String(review.status || '') === 'completed')
})
const selectedReview = computed(() => {
  return currentReviews.value.find(review => review.id === selectedReviewId.value) || null
})
const detailPopupVisible = computed({
  get: () => Boolean(selectedReviewId.value && selectedReview.value),
  set: (visible: boolean) => {
    if (!visible) {
      selectedReviewId.value = ''
    }
  },
})
const summaryFilterCount = computed(() => [
  summaryFilters.employeeName.trim(),
  selectedDepartmentPaths.value.length > 0 ? 'department' : '',
  selectedPeriods.value.length > 0 ? 'period' : '',
].filter(Boolean).length)
const summaryDepartmentTree = computed(() => buildSummaryDepartmentTree(currentReviews.value))
const summaryDepartmentTreeRows = computed(() => flattenSummaryDepartmentTree(summaryDepartmentTree.value, summaryDepartmentExpandedKeys.value))
const summaryDepartmentLabel = computed(() => {
  if (selectedDepartmentPaths.value.length === 0) {
    return '部门名称'
  }
  if (selectedDepartmentPaths.value.length === 1) {
    return selectedDepartmentPaths.value[0]
  }
  return `已选 ${selectedDepartmentPaths.value.length} 个部门`
})
const summaryPeriods = computed(() => {
  return [...new Set(currentReviews.value.map(review => String(review.period || '').trim()).filter(Boolean))]
    .sort((left, right) => right.localeCompare(left))
})
const summaryPeriodYears = computed(() => {
  const years = summaryPeriods.value
    .map(period => periodParts(period)?.year)
    .filter((year): year is number => Boolean(year))
  return [...new Set(years)].sort((left, right) => right - left)
})
const summaryPeriodSet = computed(() => new Set(summaryPeriods.value))
const summaryPeriodMonths = computed<SummaryPeriodMonth[]>(() => {
  return Array.from({ length: 12 }, (_, index) => {
    const month = index + 1
    const value = formatPeriod(summaryPeriodPickerYear.value, month)
    return {
      value,
      label: `${month}月`,
      available: summaryPeriodSet.value.has(value),
      selected: selectedPeriods.value.includes(value),
    }
  })
})
const summaryPeriodLabel = computed(() => {
  if (selectedPeriods.value.length === 0) {
    return '考评月份'
  }
  if (selectedPeriods.value.length === 1) {
    return selectedPeriods.value[0]
  }
  return `已选 ${selectedPeriods.value.length} 个月份`
})
const summaryRows = computed(() => {
  const employeeKeyword = summaryFilters.employeeName.trim().toLowerCase()
  const departmentPaths = selectedDepartmentPaths.value
  const periods = selectedPeriods.value
  return currentReviews.value.filter((review: PerformanceReview) => {
    const employeeText = [
      review.employeeName,
      userName(review.employeeId),
      review.employeeId,
    ].filter(Boolean).join(' ').toLowerCase()
    if (employeeKeyword && !employeeText.includes(employeeKeyword)) {
      return false
    }
    const departmentPath = reviewDepartmentPath(review)
    if (departmentPaths.length > 0 && !departmentPaths.some(path => departmentPath === path || departmentPath.startsWith(`${path} / `))) {
      return false
    }
    if (periods.length > 0 && !periods.includes(String(review.period || ''))) {
      return false
    }
    return true
  })
})
const pagedSummaryRows = computed(() => {
  return isMobilePage.value ? summaryRows.value : paginateRows(summaryRows.value, summaryPage.value)
})
const summaryTotalPages = computed(() => paginationTotalPages(summaryRows.value.length))
const summaryPaginationText = computed(() => `${summaryPage.value}/${summaryTotalPages.value}`)
const showSummaryPagination = computed(() => !isMobilePage.value && summaryRows.value.length > 0)
const showSummaryMobileLoadMore = computed(() => isMobilePage.value && showMobileLoadMore(summaryMobilePagination, summaryRows.value.length))
const showSummaryMobileNoMore = computed(() => isMobilePage.value && showMobileNoMore(summaryMobilePagination, summaryRows.value.length))

async function loadSummaryPageData(options: { append?: boolean } = {}) {
  const append = Boolean(options.append && isMobilePage.value)
  if (!append) {
    resetMobilePagination(summaryMobilePagination)
  }
  if (append) {
    summaryMobilePagination.loadingMore = true
  }
  else {
    loading.value = true
  }
  try {
    const [reviewRes, usersRes, templateRes] = await Promise.all([
      listReviews({
        scope: 'summary',
        status: 'completed',
        skipHistory: 1,
        ...mobileReviewListParams(isMobilePage.value, summaryMobilePagination),
      }),
      listUsers(),
      getTemplate(),
    ])
    const nextReviews = listFromPayload(reviewRes.data)
    reviews.value = append ? appendUniqueReviewRows(reviews.value, nextReviews) : nextReviews
    reviewListTotal.value = totalFromPayload(reviewRes.data, nextReviews.length)
    updateMobilePaginationTotal(summaryMobilePagination, reviewListTotal.value, reviews.value.length)
    users.value = (usersRes.data || []).map(user => ({ ...user, password: '' }))
    template.value = templateRes.data || null
    syncSummaryPeriodPickerYear()
  }
  finally {
    if (append) {
      summaryMobilePagination.loadingMore = false
    }
    else {
      loading.value = false
    }
  }
  syncSelectedReview()
}

async function loadMoreSummaryRows() {
  if (!showSummaryMobileLoadMore.value || summaryMobilePagination.loadingMore) {
    return
  }
  summaryMobilePagination.page += 1
  try {
    await loadSummaryPageData({ append: true })
  }
  catch {
    summaryMobilePagination.page = Math.max(summaryMobilePagination.page - 1, 1)
    uni.showToast({ title: '加载失败，请重试', icon: 'none' })
  }
}

const grades = computed(() => {
  const values = new Set<string>()
  for (const item of template.value?.gradeLevels || []) {
    if (item.grade) {
      values.add(String(item.grade))
    }
    if (item.label) {
      values.add(String(item.label))
    }
  }
  for (const review of reviews.value) {
    for (const key of ['finalGrade', 'hrbpGrade', 'managerGrade']) {
      const value = review[key]
      if (value) {
        values.add(String(value))
      }
    }
  }
  return [...values]
})

async function submitReviewAction(id: string, action: string, data: ReviewActionRequest) {
  detailLoading.value = true
  try {
    const res = await reviewAction(id, action, data)
    const normalizedReview = normalizeReviewActionResult(res.data, action, data)
    if (normalizedReview) {
      upsertReview(normalizedReview)
    }
    return normalizedReview
  }
  finally {
    detailLoading.value = false
  }
}

function upsertReview(review: PerformanceReview) {
  if (!review.id) {
    return
  }
  const index = reviews.value.findIndex(item => item.id === review.id)
  if (index >= 0) {
    reviews.value[index] = review
    return
  }
  reviews.value.unshift(review)
}

function syncSelectedReview() {
  const focusedId = appContent.focusedReviewId
  if (focusedId && currentReviews.value.some(review => review.id === focusedId)) {
    selectedReviewId.value = focusedId
    appContent.clearFocusedReview()
    return
  }
  if (selectedReviewId.value && currentReviews.value.some(review => review.id === selectedReviewId.value)) {
    return
  }
  selectedReviewId.value = ''
}

function refreshPageData() {
  resetSummaryPage()
  return loadSummaryPageData()
}

function resetSummaryPage() {
  summaryPage.value = 1
}

function syncSummaryPage() {
  if (summaryPage.value > 1 && pagedSummaryRows.value.length === 0) {
    resetSummaryPage()
  }
}

function handleSummaryPageChange(payload: PaginationChangePayload) {
  summaryPage.value = Math.max(Number(payload.current || 1), 1)
}

function closeSummaryFilterPickers() {
  summaryDepartmentPickerOpen.value = false
  summaryPeriodPickerOpen.value = false
}

function toggleSummaryDepartmentPicker() {
  summaryDepartmentPickerOpen.value = !summaryDepartmentPickerOpen.value
  if (summaryDepartmentPickerOpen.value) {
    summaryPeriodPickerOpen.value = false
  }
}

function toggleSummaryPeriodPicker() {
  summaryPeriodPickerOpen.value = !summaryPeriodPickerOpen.value
  if (summaryPeriodPickerOpen.value) {
    summaryDepartmentPickerOpen.value = false
    syncSummaryPeriodPickerYear()
  }
}

function toggleSummaryDepartmentExpand(key: string) {
  const next = new Set(summaryDepartmentExpandedKeys.value)
  if (next.has(key)) {
    next.delete(key)
  }
  else {
    next.add(key)
  }
  summaryDepartmentExpandedKeys.value = next
}

function summaryDepartmentSelected(path: string) {
  return selectedDepartmentPaths.value.includes(path)
}

function toggleSummaryDepartmentSelection(row: SummaryDepartmentRow) {
  const selected = summaryDepartmentSelected(row.path)
  selectedDepartmentPaths.value = selected
    ? selectedDepartmentPaths.value.filter(path => path !== row.path)
    : [...selectedDepartmentPaths.value, row.path]
}

function clearSummaryDepartments() {
  selectedDepartmentPaths.value = []
}

function summaryDepartmentRowClass(row: SummaryDepartmentRow) {
  return [
    'summary-department-row',
    row.expanded ? 'expanded' : '',
    summaryDepartmentSelected(row.path) ? 'selected' : '',
  ].filter(Boolean).join(' ')
}

function summaryDepartmentTreeRowClass(row: SummaryDepartmentRow) {
  return [
    'summary-department-tree-row',
    `depth-${row.depth}`,
    row.expanded ? 'expanded' : '',
    summaryDepartmentSelected(row.path) ? 'selected' : '',
  ].filter(Boolean).join(' ')
}

function summaryDepartmentTreeRowStyle(row: SummaryDepartmentRow) {
  return { paddingLeft: `${Math.max(row.depth - 1, 0) * 18}px` }
}

function syncSummaryPeriodPickerYear() {
  const years = summaryPeriodYears.value
  if (years.length > 0 && !years.includes(summaryPeriodPickerYear.value)) {
    summaryPeriodPickerYear.value = years[0]
  }
}

function changeSummaryPeriodPickerYear(offset: number) {
  summaryPeriodPickerYear.value += offset
}

function summaryPeriodMonthClass(month: SummaryPeriodMonth) {
  return [
    'summary-period-month',
    month.selected ? 'active' : '',
    month.available ? '' : 'disabled',
  ].filter(Boolean).join(' ')
}

function toggleSummaryPeriodSelection(month: SummaryPeriodMonth) {
  if (!month.available) {
    return
  }
  const selected = selectedPeriods.value.includes(month.value)
  selectedPeriods.value = selected
    ? selectedPeriods.value.filter(period => period !== month.value)
    : [...selectedPeriods.value, month.value].sort((left, right) => right.localeCompare(left))
}

function employeeName(review: PerformanceReview) {
  return resolveEmployeeName(review, userName)
}

function reviewPosition(review: PerformanceReview) {
  return resolveReviewPosition(review, users.value)
}

function userName(id: unknown) {
  const key = String(id || '').trim()
  if (!key) {
    return '-'
  }
  const matched = users.value.find(user => user.id === key)
  if (matched?.name) {
    return matched.name
  }
  if (sameUserId(auth.user?.id, key)) {
    return auth.user?.name || key
  }
  return key
}

function selectReview(review: PerformanceReview) {
  selectedReviewId.value = review.id
}

function handleSummaryRowClick(review: PerformanceReview) {
  if (!isMobilePage.value) {
    return
  }
  selectReview(review)
}

function confirmDelete(review: PerformanceReview) {
  uni.showModal({
    title: '删除考评单',
    content: `确认删除 ${review.period || ''} 月度考评？`,
    confirmText: '删除',
    success: async (res) => {
      if (!res.confirm) {
        return
      }
      await deleteReview(review.id)
      reviews.value = reviews.value.filter(item => item.id !== review.id)
      if (selectedReviewId.value === review.id) {
        selectedReviewId.value = ''
      }
      uni.showToast({
        title: '已删除',
        icon: 'success',
      })
    },
  })
}

function exportSummary() {
  openSummaryExport(false)
}

function exportSummaryDetail() {
  openSummaryExport(true)
}

function openSummaryExport(detail: boolean) {
  const url = exportReviewsUrl({
    scope: 'summary',
    status: 'completed',
    employeeName: summaryFilters.employeeName,
    departmentNames: selectedDepartmentPaths.value.join(','),
    periods: selectedPeriods.value.join(','),
    detail: detail ? 1 : undefined,
  })
  if (typeof window !== 'undefined') {
    window.open(url, '_blank')
    return
  }
  uni.showToast({
    title: '请在 H5 端打开导出链接',
    icon: 'none',
  })
}

function resetSummaryFilters() {
  summaryFilters.employeeName = ''
  selectedDepartmentPaths.value = []
  selectedPeriods.value = []
  closeSummaryFilterPickers()
  resetSummaryPage()
}

watch(
  () => [summaryFilters.employeeName, selectedDepartmentPaths.value.join('\0'), selectedPeriods.value.join('\0')],
  () => {
    resetSummaryPage()
    selectedReviewId.value = ''
  },
)

watch(
  () => [appContent.currentKey, appContent.refreshTick],
  () => {
    if (appContent.currentKey === 'performance:summary') {
      resetSummaryPage()
      void loadSummaryPageData()
    }
  },
  { immediate: true },
)

watch(currentReviews, syncSelectedReview)
watch(summaryRows, syncSummaryPage)
</script>

<template>
  <view class="performance-page">
    <view class="page-head">
      <view v-if="!isMobilePage" class="page-head__copy">
        <text class="page-title">
          {{ listTitle }}
        </text>
        <text class="page-desc">
          {{ listDesc }}
        </text>
      </view>
      <view class="head-actions">
        <u-button
          v-if="auth.hasButtonPermission('dingtalk_h5:button:review:export')"
          custom-class="dt-btn dt-btn-primary page-action-btn summary-export-btn"
          @click="exportSummary"
        >
          导出当前筛选
        </u-button>
        <u-button
          v-if="auth.hasButtonPermission('dingtalk_h5:button:review:export')"
          custom-class="dt-btn dt-btn-light page-action-btn summary-export-btn"
          @click="exportSummaryDetail"
        >
          导出表单详情
        </u-button>
      </view>
    </view>

    <view v-if="loading" class="panel performance-loading-panel">
      <u-loading mode="circle" />
      <text>加载中...</text>
    </view>

    <template v-else>
      <view class="table-panel-stack summary-table-panel-stack">
        <view class="panel table-panel">
          <view class="panel-head">
            <text class="panel-title">
              汇总列表
            </text>
            <text class="count-pill">
              {{ pagedSummaryRows.length }} / {{ summaryRows.length }}
            </text>
          </view>
          <view class="summary-filter-shell" :class="{ collapsed: summaryFiltersCollapsed, expanded: !summaryFiltersCollapsed }">
            <u-button custom-class="summary-filter-toggle" @click="summaryFiltersCollapsed = !summaryFiltersCollapsed">
              <text class="summary-filter-toggle-title">
                筛选条件
              </text>
              <text v-if="summaryFilterCount > 0" class="summary-filter-count">
                已选 {{ summaryFilterCount }}
              </text>
              <text class="summary-filter-arrow" :class="{ expanded: !summaryFiltersCollapsed }" />
            </u-button>
            <view class="filters summary-filters">
              <u-input v-model="summaryFilters.employeeName" custom-class="field-input" :border="true" placeholder="员工姓名" />
              <view class="summary-department-picker">
                <u-button
                  custom-class="field-input summary-filter-picker-trigger"
                  @click.stop="toggleSummaryDepartmentPicker"
                >
                  <text class="summary-filter-picker-text" :class="{ placeholder: selectedDepartmentPaths.length === 0 }">
                    {{ summaryDepartmentLabel }}
                  </text>
                  <u-icon name="arrow-down" size="14" color="#a8b0bd" />
                </u-button>
                <view v-if="summaryDepartmentPickerOpen" class="summary-filter-picker-mask" @click="closeSummaryFilterPickers" />
                <view v-if="summaryDepartmentPickerOpen" class="summary-department-dropdown" @click.stop>
                  <view class="summary-picker-head">
                    <text class="summary-picker-title">
                      部门选择
                    </text>
                    <u-button custom-class="summary-picker-clear" @click="clearSummaryDepartments">
                      清空
                    </u-button>
                  </view>
                  <view class="summary-department-tree">
                    <view v-if="summaryDepartmentTreeRows.length === 0" class="summary-picker-empty">
                      暂无部门
                    </view>
                    <view
                      v-for="row in summaryDepartmentTreeRows"
                      v-else
                      :key="row.key"
                      :class="summaryDepartmentTreeRowClass(row)"
                      :style="summaryDepartmentTreeRowStyle(row)"
                    >
                      <u-button
                        v-if="row.expandable"
                        custom-class="summary-department-expand"
                        @click.stop="toggleSummaryDepartmentExpand(row.key)"
                      >
                        <u-icon
                          name="arrow-right"
                          size="13"
                          color="#86909c"
                          :class="{ 'summary-department-chevron--open': row.expanded }"
                        />
                      </u-button>
                      <view v-else class="summary-department-expand summary-department-expand--placeholder" />
                      <u-button
                        :custom-class="summaryDepartmentRowClass(row)"
                        @click="toggleSummaryDepartmentSelection(row)"
                      >
                        <text class="summary-department-check" :class="{ selected: selectedDepartmentPaths.includes(row.path) }" />
                        <text class="summary-department-name">
                          {{ row.name }}
                        </text>
                        <text class="summary-department-count">
                          {{ row.count }}条
                        </text>
                      </u-button>
                    </view>
                  </view>
                </view>
              </view>
              <view class="summary-period-picker">
                <u-button
                  custom-class="field-input summary-filter-picker-trigger"
                  @click.stop="toggleSummaryPeriodPicker"
                >
                  <text class="summary-filter-picker-text" :class="{ placeholder: selectedPeriods.length === 0 }">
                    {{ summaryPeriodLabel }}
                  </text>
                  <u-icon name="arrow-down" size="14" color="#a8b0bd" />
                </u-button>
                <view v-if="summaryPeriodPickerOpen" class="summary-filter-picker-mask" @click="closeSummaryFilterPickers" />
                <view v-if="summaryPeriodPickerOpen" class="summary-period-dropdown" @click.stop>
                  <view class="summary-picker-head summary-picker-head--period">
                    <u-button custom-class="summary-period-nav summary-period-nav--prev" @click="changeSummaryPeriodPickerYear(-1)">
                      ‹
                    </u-button>
                    <text class="summary-picker-title">
                      {{ summaryPeriodPickerYear }}年
                    </text>
                    <u-button custom-class="summary-period-nav summary-period-nav--next" @click="changeSummaryPeriodPickerYear(1)">
                      ›
                    </u-button>
                  </view>
                  <view class="summary-period-grid">
                    <u-button
                      v-for="month in summaryPeriodMonths"
                      :key="month.value"
                      :custom-class="summaryPeriodMonthClass(month)"
                      :disabled="!month.available"
                      @click="toggleSummaryPeriodSelection(month)"
                    >
                      {{ month.label }}
                    </u-button>
                  </view>
                </view>
              </view>
              <u-button custom-class="dt-btn dt-btn-light" @click="resetSummaryFilters">
                重置
              </u-button>
            </view>
          </view>
          <view class="table-wrap">
            <table class="summary-table summary-report-table">
              <thead>
                <tr>
                  <th>员工</th>
                  <th class="summary-report-pc-only">
                    部门
                  </th>
                  <th class="summary-report-pc-only">
                    岗位
                  </th>
                  <th>考评月份</th>
                  <th class="summary-report-mobile-only">
                    最终分档
                  </th>
                  <th class="summary-report-pc-only">
                    状态
                  </th>
                  <th class="summary-report-pc-only">
                    目标得分
                  </th>
                  <th class="summary-report-pc-only">
                    上级分档
                  </th>
                  <th class="summary-report-pc-only">
                    HRBP分档
                  </th>
                  <th class="summary-report-pc-only">
                    员工确认
                  </th>
                  <th class="summary-report-pc-only">
                    最终分档
                  </th>
                  <th class="summary-report-pc-only">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="summaryRows.length === 0">
                  <td :colspan="isMobilePage ? 3 : 11" class="table-empty-cell">
                    当前没有汇总记录
                  </td>
                </tr>
                <tr
                  v-for="review in pagedSummaryRows"
                  v-else
                  :key="reviewKey(review)"
                  :class="{ 'summary-report-row--clickable': isMobilePage }"
                  @click="handleSummaryRowClick(review)"
                >
                  <td>{{ employeeName(review) }}</td>
                  <td class="summary-report-pc-only">
                    {{ reviewDepartmentText(review) }}
                  </td>
                  <td class="summary-report-pc-only">
                    {{ reviewPosition(review) }}
                  </td>
                  <td>{{ review.period || '-' }}</td>
                  <td class="summary-report-mobile-only">
                    {{ effectiveGrade(review) }}
                  </td>
                  <td class="summary-report-pc-only">
                    {{ reviewStatusText(review) }}
                  </td>
                  <td class="summary-report-pc-only">
                    {{ objectiveScore(review) }}
                  </td>
                  <td class="summary-report-pc-only">
                    {{ review.managerGrade || '-' }}
                  </td>
                  <td class="summary-report-pc-only">
                    {{ review.hrbpGrade || '-' }}
                  </td>
                  <td class="summary-report-pc-only">
                    <u-tag
                      :text="employeeConfirmText(review)"
                      :type="employeeConfirmTagType(review)"
                      mode="light"
                      size="mini"
                    />
                  </td>
                  <td class="summary-report-pc-only">
                    {{ effectiveGrade(review) }}
                  </td>
                  <td class="summary-action-buttons summary-report-pc-only">
                    <u-button custom-class="dt-btn dt-btn-light small" @click="selectReview(review)">
                      查看
                    </u-button>
                    <u-button v-if="auth.hasButtonPermission('dingtalk_h5:button:review:delete')" custom-class="dt-btn dt-btn-danger-light small" @click="confirmDelete(review)">
                      删除
                    </u-button>
                  </td>
                </tr>
              </tbody>
            </table>
          </view>
          <view v-if="showSummaryMobileLoadMore" class="mobile-list-pagination">
            <u-button
              custom-class="dt-btn dt-btn-light mobile-list-pagination__button"
              :disabled="summaryMobilePagination.loadingMore"
              @click="loadMoreSummaryRows"
            >
              {{ summaryMobilePagination.loadingMore ? '加载中...' : '加载更多' }}
            </u-button>
          </view>
          <view v-else-if="showSummaryMobileNoMore" class="mobile-list-pagination mobile-list-pagination--done">
            没有更多了
          </view>
          <view v-if="showSummaryPagination" class="pc-pagination">
            <text class="pc-pagination__total">
              共 {{ summaryRows.length }} 条
            </text>
            <u-pagination
              v-model="summaryPage"
              custom-class="pc-pagination-control"
              :total="summaryRows.length"
              :page-size="TABLE_PAGE_SIZE"
              prev-text="上一页"
              next-text="下一页"
              @change="handleSummaryPageChange"
            >
              <u-text custom-class="pc-pagination-current" :text="summaryPaginationText" />
            </u-pagination>
          </view>
        </view>
      </view>

      <PerformanceDetailPopup
        v-model="detailPopupVisible"
        :review="selectedReview"
        :users="users"
        :template="template"
        :grades="grades"
        :popup-width="isMobilePage ? '94%' : '90%'"
        :detail-loading="detailLoading"
        :submit-review-action="submitReviewAction"
        @action-success="refreshPageData"
      />
    </template>
  </view>
</template>

<style lang="scss" scoped src="./components/performance-page.scss"></style>
