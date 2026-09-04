<script setup lang="ts">
import type { PerformanceReview, PerformanceReviewListPayload, PerformanceTemplate } from '@/types/dingtalk-h5'
import { computed, reactive, ref, watch } from 'vue'
import { getTemplate, listReviews } from '@/api/dingtalk-h5'
import { useDingtalkAuthStore } from '@/stores'
import { useAppContentStore } from '@/stores/appContent'
import PerformanceAdaptiveSelect from './components/PerformanceAdaptiveSelect.vue'
import PerformanceDetailPopup from './components/PerformanceDetailPopup.vue'
import { appendUniqueReviewRows, createMobilePaginationState, mobileReviewListParams, resetMobilePagination, showMobileLoadMore, showMobileNoMore, updateMobilePaginationTotal } from './components/mobilePagination'
import { statusMeta } from './constants/performanceStatus'

interface SelectOption {
  value: string
  label: string
  type?: 'group'
}

interface PaginationChangePayload {
  current?: number
}

interface GradeGroup {
  label: string
  grades: string[]
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

function createGradeGroups(sourceTemplate: PerformanceTemplate | null, sourceReviews: PerformanceReview[]) {
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
    for (const key of ['finalGrade', 'hrbpGrade', 'managerGrade'] as const) {
      const grade = firstText(review[key])
      if (grade && !knownGrades.has(grade)) {
        addUniqueText(otherGroup.grades, grade)
      }
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

function createTextSelectOptions(emptyLabel: string, values: string[]) {
  return [
    { value: '', label: emptyLabel },
    ...values.map(value => ({ value, label: value })),
  ]
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

function reviewKey(review: PerformanceReview) {
  return review.id || review.reviewNo || `${review.employeeId || 'employee'}-${review.period || 'period'}`
}

function reviewStatusText(review: PerformanceReview) {
  return statusMeta[String(review.status || '')]?.label || String(review.status || '未知')
}

function reviewStatusType(review: PerformanceReview) {
  return statusMeta[String(review.status || '')]?.type || 'info'
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

function finalScore(review: PerformanceReview) {
  return firstText(
    review.finalScore,
    review.totalScore,
    review.score,
    review.objectiveScore,
    totalObjectiveScore(review),
    review.managerScore,
    review.hrbpScore,
  ) || '-'
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

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const reviews = ref<PerformanceReview[]>([])
const template = ref<PerformanceTemplate | null>(null)
const reviewListTotal = ref(0)
const loading = ref(false)
const isMobilePage = ref(resolveMobilePage())
const historyPage = ref(1)
const historyMobilePagination = createMobilePaginationState()
const historyFiltersCollapsed = ref(true)
const selectedReviewId = ref('')
const historyFilters = reactive({
  period: '',
  grade: '',
})
const listTitle = '历史绩效'
const currentReviews = computed(() => {
  return reviews.value.filter(review => (
    sameUserId(review.employeeId, auth.user?.id)
    && String(review.status || '') === 'completed'
  ))
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
const listDesc = computed(() => `共 ${reviewListTotal.value || currentReviews.value.length} 条记录`)
const historyFilterCount = computed(() => [
  historyFilters.period,
  historyFilters.grade,
].filter(Boolean).length)
const historyRows = computed(() => {
  return currentReviews.value.slice().sort((left: PerformanceReview, right: PerformanceReview) => String(right.period || '').localeCompare(String(left.period || '')))
})
const pagedHistoryRows = computed(() => {
  return isMobilePage.value ? historyRows.value : paginateRows(historyRows.value, historyPage.value)
})
const historyTotalPages = computed(() => paginationTotalPages(historyRows.value.length))
const historyPaginationText = computed(() => `${historyPage.value}/${historyTotalPages.value}`)
const showHistoryPagination = computed(() => !isMobilePage.value && historyRows.value.length > 0)
const showHistoryMobileLoadMore = computed(() => isMobilePage.value && showMobileLoadMore(historyMobilePagination, historyRows.value.length))
const showHistoryMobileNoMore = computed(() => isMobilePage.value && showMobileNoMore(historyMobilePagination, historyRows.value.length))
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
const historyPeriodSelectOptions = computed<SelectOption[]>(() => {
  const periods = recentReviewPeriods()
  addUniqueText(periods, historyFilters.period)
  for (const review of reviews.value) {
    addUniqueText(periods, review.period)
  }
  return createTextSelectOptions('全部考评月份', periods.sort((left, right) => right.localeCompare(left)))
})
const historyGradeGroups = computed(() => createGradeGroups(template.value, reviews.value))
const historyGradeSelectOptions = computed<SelectOption[]>(() => createGradeSelectOptions('全部分档', historyGradeGroups.value))

function historyFilterParams() {
  const params: Record<string, unknown> = {}
  const period = historyFilters.period.trim()
  const grade = historyFilters.grade.trim()
  if (period) {
    params.period = period
  }
  if (grade) {
    params.grade = grade
  }
  return params
}

function resetHistoryPage() {
  historyPage.value = 1
}

function syncHistoryPage() {
  if (historyPage.value > 1 && pagedHistoryRows.value.length === 0) {
    resetHistoryPage()
  }
}

function handleHistoryPageChange(payload: PaginationChangePayload) {
  historyPage.value = Math.max(Number(payload.current || 1), 1)
}

async function loadHistoryPageData(options: { append?: boolean } = {}) {
  const append = Boolean(options.append && isMobilePage.value)
  if (!append) {
    resetMobilePagination(historyMobilePagination)
  }
  if (append) {
    historyMobilePagination.loadingMore = true
  }
  else {
    loading.value = true
  }
  try {
    const [reviewRes, templateRes] = await Promise.all([
      listReviews({
        scope: 'mine',
        status: 'completed',
        skipHistory: 1,
        ...historyFilterParams(),
        ...mobileReviewListParams(isMobilePage.value, historyMobilePagination),
      }),
      getTemplate(),
    ])
    const nextReviews = listFromPayload(reviewRes.data)
    reviews.value = append ? appendUniqueReviewRows(reviews.value, nextReviews) : nextReviews
    reviewListTotal.value = totalFromPayload(reviewRes.data, nextReviews.length)
    updateMobilePaginationTotal(historyMobilePagination, reviewListTotal.value, reviews.value.length)
    template.value = templateRes.data || null
  }
  finally {
    if (append) {
      historyMobilePagination.loadingMore = false
    }
    else {
      loading.value = false
    }
  }
  syncSelectedReview()
}

async function loadMoreHistoryRows() {
  if (!showHistoryMobileLoadMore.value || historyMobilePagination.loadingMore) {
    return
  }
  historyMobilePagination.page += 1
  try {
    await loadHistoryPageData({ append: true })
  }
  catch {
    historyMobilePagination.page = Math.max(historyMobilePagination.page - 1, 1)
    uni.showToast({ title: '加载失败，请重试', icon: 'none' })
  }
}

function syncSelectedReview() {
  if (selectedReviewId.value && currentReviews.value.some(review => review.id === selectedReviewId.value)) {
    return
  }
  selectedReviewId.value = ''
}

function refreshPageData() {
  resetHistoryPage()
  return loadHistoryPageData()
}

async function setHistoryFilter(field: keyof typeof historyFilters, value: string) {
  if (!Object.prototype.hasOwnProperty.call(historyFilters, field)) {
    return
  }
  historyFilters[field] = value.trim()
  resetHistoryPage()
  selectedReviewId.value = ''
  await loadHistoryPageData()
}

async function resetHistoryFilters() {
  if (!historyFilters.period && !historyFilters.grade) {
    return
  }
  historyFilters.period = ''
  historyFilters.grade = ''
  resetHistoryPage()
  selectedReviewId.value = ''
  await loadHistoryPageData()
}

function selectReview(review: PerformanceReview) {
  selectedReviewId.value = review.id
}

function selectMobileHistoryReview(review: PerformanceReview) {
  if (!isMobilePage.value) {
    return
  }
  selectReview(review)
}

watch(
  () => [appContent.currentKey, appContent.refreshTick],
  () => {
    if (appContent.currentKey === 'performance:history') {
      resetHistoryPage()
      void loadHistoryPageData()
    }
  },
  { immediate: true },
)

watch(currentReviews, syncSelectedReview)
watch(historyRows, syncHistoryPage)
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
    </view>

    <view v-if="loading" class="panel performance-loading-panel">
      <u-loading mode="circle" />
      <text>加载中...</text>
    </view>

    <template v-else>
      <view class="table-panel-stack history-table-panel-stack">
        <view class="panel table-panel">
          <view class="panel-head">
            <text class="panel-title">
              历史记录
            </text>
            <text class="count-pill">
              {{ pagedHistoryRows.length }} / {{ historyRows.length }} 条
            </text>
          </view>
          <view class="summary-filter-shell history-filter-shell" :class="{ collapsed: historyFiltersCollapsed, expanded: !historyFiltersCollapsed }">
            <u-button custom-class="summary-filter-toggle" @click="historyFiltersCollapsed = !historyFiltersCollapsed">
              <text class="summary-filter-toggle-title">
                筛选条件
              </text>
              <text v-if="historyFilterCount > 0" class="summary-filter-count">
                已选 {{ historyFilterCount }}
              </text>
              <text class="summary-filter-arrow" :class="{ expanded: !historyFiltersCollapsed }" />
            </u-button>
            <view class="filters summary-filters history-filters">
              <PerformanceAdaptiveSelect
                :model-value="historyFilters.period"
                custom-class="field-select"
                title="筛选考评月份"
                desktop-variant="month-grid"
                desktop-placement="bottom"
                :options="historyPeriodSelectOptions"
                :border="true"
                @change="setHistoryFilter('period', $event)"
              />
              <PerformanceAdaptiveSelect
                :model-value="historyFilters.grade"
                custom-class="field-select"
                title="筛选分档"
                :options="historyGradeSelectOptions"
                :border="true"
                @change="setHistoryFilter('grade', $event)"
              />
              <u-button custom-class="dt-btn dt-btn-light" @click="resetHistoryFilters">
                重置
              </u-button>
            </view>
          </view>
          <view class="table-wrap">
            <table class="summary-table history-performance-table">
              <thead>
                <tr>
                  <th>考评月份</th>
                  <th>状态</th>
                  <th class="history-pc-only">
                    目标得分
                  </th>
                  <th class="history-mobile-only">
                    最终得分
                  </th>
                  <th class="history-pc-only">
                    上级分档
                  </th>
                  <th class="history-pc-only">
                    HRBP分档
                  </th>
                  <th class="history-pc-only">
                    最终分档
                  </th>
                  <th class="history-pc-only">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="historyRows.length === 0">
                  <td :colspan="isMobilePage ? 3 : 7" class="table-empty-cell">
                    当前没有历史记录
                  </td>
                </tr>
                <tr
                  v-for="review in pagedHistoryRows"
                  v-else
                  :key="reviewKey(review)"
                  class="history-review-row"
                  @click="selectMobileHistoryReview(review)"
                >
                  <td>{{ review.period || '-' }}</td>
                  <td>
                    <u-tag
                      v-if="isMobilePage"
                      custom-class="mobile-status-tag"
                      :text="reviewStatusText(review)"
                      :type="reviewStatusType(review)"
                      mode="light"
                      size="mini"
                    />
                    <template v-else>
                      {{ reviewStatusText(review) }}
                    </template>
                  </td>
                  <td class="history-pc-only">
                    {{ objectiveScore(review) }}
                  </td>
                  <td class="history-mobile-only">
                    {{ finalScore(review) }}
                  </td>
                  <td class="history-pc-only">
                    {{ review.managerGrade || '-' }}
                  </td>
                  <td class="history-pc-only">
                    {{ review.hrbpGrade || '-' }}
                  </td>
                  <td class="history-pc-only">
                    {{ effectiveGrade(review) }}
                  </td>
                  <td class="history-pc-only">
                    <u-button custom-class="dt-btn dt-btn-light small" @click="selectReview(review)">
                      查看
                    </u-button>
                  </td>
                </tr>
              </tbody>
            </table>
          </view>
          <view v-if="showHistoryMobileLoadMore" class="mobile-list-pagination">
            <u-button
              custom-class="dt-btn dt-btn-light mobile-list-pagination__button"
              :disabled="historyMobilePagination.loadingMore"
              @click="loadMoreHistoryRows"
            >
              {{ historyMobilePagination.loadingMore ? '加载中...' : '加载更多' }}
            </u-button>
          </view>
          <view v-else-if="showHistoryMobileNoMore" class="mobile-list-pagination mobile-list-pagination--done">
            没有更多了
          </view>
          <view v-if="showHistoryPagination" class="pc-pagination">
            <text class="pc-pagination__total">
              共 {{ historyRows.length }} 条
            </text>
            <u-pagination
              v-model="historyPage"
              custom-class="pc-pagination-control"
              :total="historyRows.length"
              :page-size="TABLE_PAGE_SIZE"
              prev-text="上一页"
              next-text="下一页"
              @change="handleHistoryPageChange"
            >
              <u-text custom-class="pc-pagination-current" :text="historyPaginationText" />
            </u-pagination>
          </view>
        </view>
      </view>

      <PerformanceDetailPopup
        v-model="detailPopupVisible"
        :review="selectedReview"
        :template="template"
        :grades="grades"
        @action-success="refreshPageData"
      />
    </template>
  </view>
</template>

<style lang="scss" scoped src="./components/performance-page.scss"></style>
