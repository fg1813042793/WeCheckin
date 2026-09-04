<script setup lang="ts">
import type { PerformanceReview, PerformanceReviewListPayload, PerformanceTemplate, PerformanceUser, ReviewActionRequest } from '@/types/dingtalk-h5'
import { computed, reactive, ref, watch } from 'vue'
import { getTemplate, listReviews, reviewAction } from '@/api/dingtalk-h5'
import { useDingtalkAuthStore } from '@/stores'
import { useAppContentStore } from '@/stores/appContent'
import PerformanceAdaptiveSelect from './components/PerformanceAdaptiveSelect.vue'
import PerformanceDetailPopup from './components/PerformanceDetailPopup.vue'
import PerformanceReviewDetail from './components/PerformanceReviewDetail.vue'
import { appendUniqueReviewRows, createMobilePaginationState, mobileReviewListParams, resetMobilePagination, showMobileLoadMore, showMobileNoMore, updateMobilePaginationTotal } from './components/mobilePagination'
import { normalizeReviewActionResult, statusMeta } from './constants/performanceStatus'

interface SelectOption {
  value: string
  label: string
  type?: 'group'
}

interface GradeGroup {
  label: string
  grades: string[]
}

interface PaginationChangePayload {
  current?: number
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

function addUniqueText(list: string[], value: unknown) {
  const text = firstText(value)
  if (text && !list.includes(text)) {
    list.push(text)
  }
}

function createTextSelectOptions(emptyLabel: string, values: string[]) {
  return [
    { value: '', label: emptyLabel },
    ...values.map(value => ({ value, label: value })),
  ]
}

function createManagerGradeGroups(sourceTemplate: PerformanceTemplate | null, sourceReviews: PerformanceReview[]) {
  const groups: GradeGroup[] = []
  const groupByLabel = new Map<string, GradeGroup>()
  const knownGrades = new Set<string>()

  for (const item of sourceTemplate?.gradeLevels || []) {
    const groupLabel = firstText(item.label, '其他')
    const grade = firstText(item.grade, item.label)
    if (!grade) {
      continue
    }

    let group = groupByLabel.get(groupLabel)
    if (!group) {
      group = { label: groupLabel, grades: [] }
      groupByLabel.set(groupLabel, group)
      groups.push(group)
    }
    addUniqueText(group.grades, grade)
    knownGrades.add(grade)
  }

  const otherGroup: GradeGroup = { label: '其他', grades: [] }
  for (const review of sourceReviews) {
    const grade = firstText(review.managerGrade)
    if (grade && !knownGrades.has(grade)) {
      addUniqueText(otherGroup.grades, grade)
    }
  }
  if (otherGroup.grades.length) {
    groups.push(otherGroup)
  }

  return groups
}

function createGradeSelectOptions(emptyLabel: string, groups: GradeGroup[]) {
  const options: SelectOption[] = [{ value: '', label: emptyLabel }]
  groups.forEach((group, index) => {
    options.push({ value: `__grade_group_${index}`, label: group.label, type: 'group' })
    for (const grade of group.grades) {
      options.push({ value: grade, label: grade })
    }
  })
  return options
}

function recentReviewPeriods(monthCount = 18) {
  const periods: string[] = []
  const current = new Date()
  for (let index = 0; index < monthCount; index += 1) {
    const date = new Date(current.getFullYear(), current.getMonth() - index, 1)
    periods.push(`${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`)
  }
  return periods
}

function readInputValue(event: unknown) {
  if (typeof event === 'string' || typeof event === 'number') {
    return String(event)
  }
  const candidate = event as { detail?: { value?: unknown }, target?: { value?: unknown } }
  return String(candidate?.detail?.value ?? candidate?.target?.value ?? '')
}

function reviewKey(review: PerformanceReview) {
  return review.id || review.reviewNo || `${review.employeeId || 'employee'}-${review.period || 'period'}`
}

function reviewStatusText(review: PerformanceReview) {
  return statusMeta[String(review.status || '')]?.label || String(review.status || '未知')
}

function reviewStatusType(review: PerformanceReview) {
  return statusMeta[String(review.status || '')]?.type || 'info'
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

function shouldShowFinalGrade(review: PerformanceReview | Record<string, unknown>) {
  const status = String(review.status || '').trim()
  const employeeConfirmResult = String(review.employeeConfirmResult || '').trim()
  if (employeeConfirmResult === 'disputed') {
    return false
  }
  return employeeConfirmResult === 'confirmed' || status === 'hr_final' || status === 'completed'
}

function resolveEffectiveGrade(review: PerformanceReview | Record<string, unknown>) {
  if (!shouldShowFinalGrade(review)) {
    return '-'
  }
  return firstText(review.finalGrade, review.hrbpGrade) || '-'
}

function currentAssignee(review: PerformanceReview, resolveUserName: (id: unknown) => string) {
  const explicitAssignee = firstText(review.currentAssigneeName)
  if (explicitAssignee && review.status !== 'completed') {
    return explicitAssignee
  }
  if (review.status === 'draft') {
    return firstText(review.employeeName, resolveUserName(review.employeeId))
  }
  if (review.status === 'manager_review') {
    return firstText(review.managerName, resolveUserName(review.managerId))
  }
  if (review.status === 'hrbp_review') {
    return firstText(review.hrbpReviewerName, review.hrbpName, resolveUserName(review.hrbpReviewerId || review.hrbpId))
  }
  if (review.status === 'employee_confirm') {
    return firstText(review.employeeName, resolveUserName(review.employeeId))
  }
  if (review.status === 'hr_final') {
    return firstText(review.hrbpReviewerName, review.hrbpName, resolveUserName(review.hrbpReviewerId || review.hrbpId))
  }
  return '已归档'
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
const managerPage = ref(1)
const managerMobilePagination = createMobilePaginationState()
const managerReviewTab = ref<'pending' | 'reviewed'>('pending')
const managerReviewedFiltersCollapsed = ref(true)
const selectedReviewId = ref('')
const managerReviewedFilters = reactive({
  employeeName: '',
  period: '',
  managerGrade: '',
})
const listTitle = '我要评价'
const managerPendingStatuses = ['manager_review']
const managerReviewedStatuses = ['hrbp_review', 'employee_confirm', 'hr_final', 'completed']
const managerReviewedFilterCount = computed(() => [
  managerReviewedFilters.employeeName,
  managerReviewedFilters.period,
  managerReviewedFilters.managerGrade,
].filter(Boolean).length)
const managerReviewedPeriodSelectOptions = computed<SelectOption[]>(() => {
  const periods = recentReviewPeriods()
  addUniqueText(periods, managerReviewedFilters.period)
  for (const review of reviews.value) {
    addUniqueText(periods, review.period)
  }
  return createTextSelectOptions('全部考评月份', periods.sort((left, right) => right.localeCompare(left)))
})
const managerReviewedGradeGroups = computed(() => createManagerGradeGroups(template.value, reviews.value))
const managerReviewedGradeSelectOptions = computed<SelectOption[]>(() => createGradeSelectOptions('全部上级分档', managerReviewedGradeGroups.value))
const managerRows = computed(() => {
  const statuses = managerReviewTab.value === 'pending' ? managerPendingStatuses : managerReviewedStatuses
  let rows = reviews.value.filter(review => statuses.includes(String(review.status || '')))
  const managerGrade = managerReviewedFilters.managerGrade.trim()
  if (managerReviewTab.value === 'reviewed' && managerGrade) {
    rows = rows.filter(review => String(review.managerGrade || '').trim() === managerGrade)
  }
  return rows
})
const pagedManagerRows = computed(() => {
  return isMobilePage.value ? managerRows.value : paginateRows(managerRows.value, managerPage.value)
})
const managerTotalPages = computed(() => paginationTotalPages(managerRows.value.length))
const managerPaginationText = computed(() => `${managerPage.value}/${managerTotalPages.value}`)
const showManagerPagination = computed(() => !isMobilePage.value && managerRows.value.length > 0)
const showManagerMobileLoadMore = computed(() => isMobilePage.value && showMobileLoadMore(managerMobilePagination, managerRows.value.length))
const showManagerMobileNoMore = computed(() => isMobilePage.value && showMobileNoMore(managerMobilePagination, managerRows.value.length))
const selectedReview = computed(() => {
  return managerRows.value.find(review => review.id === selectedReviewId.value) || null
})
const mobileReviewDetailVisible = computed(() => isMobilePage.value && Boolean(selectedReview.value))
const listDesc = computed(() => `共 ${reviewListTotal.value || managerRows.value.length} 条记录`)
const emptyTitle = computed(() => managerReviewTab.value === 'pending' ? '当前没有待评记录' : '当前没有已评记录')
const detailPopupVisible = computed({
  get: () => !isMobilePage.value && Boolean(selectedReviewId.value && selectedReview.value),
  set: (visible: boolean) => {
    if (!visible) {
      selectedReviewId.value = ''
    }
  },
})
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

function reviewStatusParamsForTab() {
  const statuses = managerReviewTab.value === 'reviewed' ? managerReviewedStatuses : managerPendingStatuses
  if (statuses.length === 1) {
    return { status: statuses[0] }
  }
  return { statuses: statuses.join(',') }
}

function managerReviewedFilterParams() {
  const params: Record<string, unknown> = {}
  if (managerReviewedFilters.employeeName.trim()) {
    params.employeeName = managerReviewedFilters.employeeName.trim()
  }
  if (managerReviewedFilters.period.trim()) {
    params.period = managerReviewedFilters.period.trim()
  }
  return params
}

async function loadManagerPageData(options: { append?: boolean } = {}) {
  const append = Boolean(options.append && isMobilePage.value)
  if (!append) {
    resetMobilePagination(managerMobilePagination)
  }
  const params = {
    scope: 'manager',
    skipHistory: 1,
    ...reviewStatusParamsForTab(),
    ...(managerReviewTab.value === 'reviewed' ? managerReviewedFilterParams() : {}),
    ...mobileReviewListParams(isMobilePage.value, managerMobilePagination),
  }
  if (append) {
    managerMobilePagination.loadingMore = true
  }
  else {
    loading.value = true
  }
  try {
    const [reviewRes, templateRes] = await Promise.all([
      listReviews(params),
      getTemplate(),
    ])
    const nextReviews = listFromPayload(reviewRes.data)
    reviews.value = append ? appendUniqueReviewRows(reviews.value, nextReviews) : nextReviews
    reviewListTotal.value = totalFromPayload(reviewRes.data, nextReviews.length)
    updateMobilePaginationTotal(managerMobilePagination, reviewListTotal.value, reviews.value.length)
    template.value = templateRes.data || null
  }
  finally {
    if (append) {
      managerMobilePagination.loadingMore = false
    }
    else {
      loading.value = false
    }
  }
  syncSelectedReview()
}

async function loadMoreManagerRows() {
  if (!showManagerMobileLoadMore.value || managerMobilePagination.loadingMore) {
    return
  }
  managerMobilePagination.page += 1
  try {
    await loadManagerPageData({ append: true })
  }
  catch {
    managerMobilePagination.page = Math.max(managerMobilePagination.page - 1, 1)
    uni.showToast({ title: '加载失败，请重试', icon: 'none' })
  }
}

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
  if (focusedId && managerRows.value.some(review => review.id === focusedId)) {
    selectedReviewId.value = focusedId
    appContent.clearFocusedReview()
    return
  }
  if (selectedReviewId.value && managerRows.value.some(review => review.id === selectedReviewId.value)) {
    return
  }
  selectedReviewId.value = ''
}

function refreshPageData() {
  resetManagerPage()
  return loadManagerPageData()
}

function resetManagerPage() {
  managerPage.value = 1
}

function syncManagerPage() {
  if (managerPage.value > 1 && pagedManagerRows.value.length === 0) {
    resetManagerPage()
  }
}

function handleManagerPageChange(payload: PaginationChangePayload) {
  managerPage.value = Math.max(Number(payload.current || 1), 1)
}

function employeeName(review: PerformanceReview) {
  return resolveEmployeeName(review, userName)
}

function reviewPosition(review: PerformanceReview) {
  return resolveReviewPosition(review, users.value)
}

function effectiveGrade(review: PerformanceReview) {
  return resolveEffectiveGrade(review)
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

function actionText(review: PerformanceReview) {
  return review.status === 'manager_review' ? '评价' : '查看'
}

function managerReviewTabClass(tab: 'pending' | 'reviewed') {
  return ['hrbp-review-tab', managerReviewTab.value === tab ? 'active' : ''].filter(Boolean).join(' ')
}

function managerReviewTableClass() {
  return [
    'summary-table',
    'hrbp-review-table',
    'manager-review-table',
    managerReviewTab.value === 'pending' ? 'manager-review-table--pending' : 'manager-review-table--reviewed',
  ].join(' ')
}

function reviewActionButtonClass(review: PerformanceReview) {
  return ['dt-btn', 'small', 'manager-review-action-btn', actionText(review) === '评价' ? 'dt-btn-primary' : 'dt-btn-light'].join(' ')
}

async function switchManagerReviewTab(tab: 'pending' | 'reviewed') {
  if (managerReviewTab.value === tab) {
    return
  }
  managerReviewTab.value = tab
  resetManagerPage()
  selectedReviewId.value = ''
  await loadManagerPageData()
}

function setManagerReviewedFilter(field: keyof typeof managerReviewedFilters, value: string) {
  if (Object.prototype.hasOwnProperty.call(managerReviewedFilters, field)) {
    managerReviewedFilters[field] = value.trim()
  }
}

async function searchManagerReviewedReviews() {
  resetManagerPage()
  selectedReviewId.value = ''
  await loadManagerPageData()
}

async function resetManagerReviewedFilters() {
  managerReviewedFilters.employeeName = ''
  managerReviewedFilters.period = ''
  managerReviewedFilters.managerGrade = ''
  resetManagerPage()
  selectedReviewId.value = ''
  await loadManagerPageData()
}

function selectReview(review: PerformanceReview) {
  selectedReviewId.value = review.id
}

function selectMobileReview(review: PerformanceReview) {
  if (!isMobilePage.value) {
    return
  }
  selectReview(review)
}

function closeMobileDetail() {
  selectedReviewId.value = ''
}

watch(
  () => [appContent.currentKey, appContent.refreshTick],
  () => {
    if (appContent.currentKey === 'performance:manager') {
      resetManagerPage()
      void loadManagerPageData()
    }
  },
  { immediate: true },
)

watch(managerRows, syncSelectedReview)
watch(managerRows, syncManagerPage)
</script>

<template>
  <view class="performance-page">
    <template v-if="mobileReviewDetailVisible">
      <view class="mobile-review-detail-page">
        <view class="mobile-review-detail-nav">
          <u-button custom-class="mobile-detail-back-btn" @click="closeMobileDetail">
            <u-icon name="arrow-left" size="15" color="#1677ff" />
            <text>返回待评列表</text>
          </u-button>
        </view>
        <PerformanceReviewDetail
          v-if="selectedReview"
          :review="selectedReview"
          :users="users"
          :template="template"
          :grades="grades"
          action-scope="manager"
          :detail-loading="detailLoading"
          :submit-review-action="submitReviewAction"
          @action-success="refreshPageData"
        />
      </view>
    </template>

    <template v-else>
      <view class="page-head">
        <view v-if="!isMobilePage" class="page-head__copy">
          <text class="page-title">
            {{ listTitle }}
          </text>
          <text class="page-desc">
            {{ listDesc }}
          </text>
        </view>
      </view>

      <view v-if="loading" class="panel performance-loading-panel">
        <u-loading mode="circle" />
        <text>加载中...</text>
      </view>

      <template v-else>
        <view class="hrbp-review-tabs-bar">
          <view class="hrbp-review-tabs">
            <u-button
              :custom-class="managerReviewTabClass('pending')"
              @click="switchManagerReviewTab('pending')"
            >
              待评
            </u-button>
            <u-button
              :custom-class="managerReviewTabClass('reviewed')"
              @click="switchManagerReviewTab('reviewed')"
            >
              已评
            </u-button>
          </view>
        </view>

        <view class="table-panel-stack">
          <view class="panel table-panel">
            <view class="panel-head">
              <text class="panel-title">
                评价列表
              </text>
              <text class="count-pill">
                {{ pagedManagerRows.length }} / {{ managerRows.length }}
              </text>
            </view>
            <view
              v-if="managerReviewTab === 'reviewed'"
              class="summary-filter-shell manager-reviewed-filter-shell"
              :class="{ collapsed: managerReviewedFiltersCollapsed, expanded: !managerReviewedFiltersCollapsed }"
            >
              <u-button custom-class="summary-filter-toggle" @click="managerReviewedFiltersCollapsed = !managerReviewedFiltersCollapsed">
                <text class="summary-filter-toggle-title">
                  筛选条件
                </text>
                <text v-if="managerReviewedFilterCount > 0" class="summary-filter-count">
                  已选 {{ managerReviewedFilterCount }}
                </text>
                <text class="summary-filter-arrow" :class="{ expanded: !managerReviewedFiltersCollapsed }" />
              </u-button>
              <view class="filters summary-filters manager-reviewed-filters">
                <u-input
                  custom-class="field-input"
                  :model-value="managerReviewedFilters.employeeName"
                  :border="true"
                  placeholder="员工姓名/账号"
                  @input="setManagerReviewedFilter('employeeName', readInputValue($event))"
                />
                <PerformanceAdaptiveSelect
                  :model-value="managerReviewedFilters.period"
                  custom-class="field-select"
                  title="筛选考评月份"
                  desktop-variant="month-grid"
                  desktop-placement="bottom"
                  :options="managerReviewedPeriodSelectOptions"
                  :border="true"
                  @change="setManagerReviewedFilter('period', $event)"
                />
                <PerformanceAdaptiveSelect
                  :model-value="managerReviewedFilters.managerGrade"
                  custom-class="field-select"
                  title="筛选上级分档"
                  :options="managerReviewedGradeSelectOptions"
                  :border="true"
                  @change="setManagerReviewedFilter('managerGrade', $event)"
                />
                <u-button custom-class="dt-btn dt-btn-primary" @click="searchManagerReviewedReviews">
                  查询
                </u-button>
                <u-button custom-class="dt-btn dt-btn-light" @click="resetManagerReviewedFilters">
                  重置
                </u-button>
              </view>
            </view>

            <view class="table-wrap">
              <table :class="managerReviewTableClass()">
                <thead>
                  <tr>
                    <th>员工</th>
                    <th class="manager-review-pc-only">
                      岗位
                    </th>
                    <th>考评月份</th>
                    <th class="manager-review-mobile-only manager-review-pending-only">
                      状态
                    </th>
                    <th class="manager-review-pc-only">
                      状态
                    </th>
                    <th class="manager-review-pc-only">
                      目标得分
                    </th>
                    <th class="manager-review-mobile-only manager-review-reviewed-only">
                      上级分档
                    </th>
                    <th class="manager-review-pc-only">
                      上级分档
                    </th>
                    <th class="manager-review-pc-only">
                      HRBP分档
                    </th>
                    <th class="manager-review-pc-only">
                      最终分档
                    </th>
                    <th class="manager-review-pc-only">
                      当前处理人
                    </th>
                    <th class="manager-review-pc-only">
                      操作
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="managerRows.length === 0">
                    <td :colspan="isMobilePage ? 3 : 10" class="table-empty-cell">
                      {{ emptyTitle }}
                    </td>
                  </tr>
                  <tr
                    v-for="review in pagedManagerRows"
                    v-else
                    :key="reviewKey(review)"
                    @click="selectMobileReview(review)"
                  >
                    <td>{{ employeeName(review) }}</td>
                    <td class="manager-review-pc-only">
                      {{ reviewPosition(review) }}
                    </td>
                    <td>{{ review.period || '-' }}</td>
                    <td class="manager-review-mobile-only manager-review-pending-only">
                      <u-tag
                        custom-class="mobile-status-tag"
                        :text="reviewStatusText(review)"
                        :type="reviewStatusType(review)"
                        mode="light"
                        size="mini"
                      />
                    </td>
                    <td class="manager-review-pc-only">
                      {{ reviewStatusText(review) }}
                    </td>
                    <td class="manager-review-pc-only">
                      {{ objectiveScore(review) }}
                    </td>
                    <td class="manager-review-mobile-only manager-review-reviewed-only">
                      {{ review.managerGrade || '-' }}
                    </td>
                    <td class="manager-review-pc-only">
                      {{ review.managerGrade || '-' }}
                    </td>
                    <td class="manager-review-pc-only">
                      {{ review.hrbpGrade || '-' }}
                    </td>
                    <td class="manager-review-pc-only">
                      {{ effectiveGrade(review) }}
                    </td>
                    <td class="manager-review-pc-only">
                      {{ currentAssignee(review, userName) }}
                    </td>
                    <td class="manager-review-pc-only">
                      <u-button :custom-class="reviewActionButtonClass(review)" @click="selectReview(review)">
                        {{ actionText(review) }}
                      </u-button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </view>
            <view v-if="showManagerMobileLoadMore" class="mobile-list-pagination">
              <u-button
                custom-class="dt-btn dt-btn-light mobile-list-pagination__button"
                :disabled="managerMobilePagination.loadingMore"
                @click="loadMoreManagerRows"
              >
                {{ managerMobilePagination.loadingMore ? '加载中...' : '加载更多' }}
              </u-button>
            </view>
            <view v-else-if="showManagerMobileNoMore" class="mobile-list-pagination mobile-list-pagination--done">
              没有更多了
            </view>
            <view v-if="showManagerPagination" class="pc-pagination">
              <text class="pc-pagination__total">
                共 {{ managerRows.length }} 条
              </text>
              <u-pagination
                v-model="managerPage"
                custom-class="pc-pagination-control"
                :total="managerRows.length"
                :page-size="TABLE_PAGE_SIZE"
                prev-text="上一页"
                next-text="下一页"
                @change="handleManagerPageChange"
              >
                <u-text custom-class="pc-pagination-current" :text="managerPaginationText" />
              </u-pagination>
            </view>
          </view>
        </view>

        <PerformanceDetailPopup
          v-if="!isMobilePage"
          v-model="detailPopupVisible"
          :review="selectedReview"
          :users="users"
          :template="template"
          :grades="grades"
          action-scope="manager"
          :detail-loading="detailLoading"
          :submit-review-action="submitReviewAction"
          @action-success="refreshPageData"
        />
      </template>
    </template>
  </view>
</template>

<style lang="scss" scoped src="./components/performance-page.scss"></style>
