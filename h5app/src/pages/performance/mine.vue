<script setup lang="ts">
import type {
  CreateReviewPayload,
  CreateReviewRequest,
  PerformanceReview,
  PerformanceReviewListPayload,
  PerformanceTemplate,
  PerformanceUser,
  ReviewActionRequest,
} from '@/types/dingtalk-h5'
import { computed, reactive, ref, watch } from 'vue'
import { createReview, getTemplate, listReviews, listUsers, reviewAction } from '@/api/dingtalk-h5'
import { useDingtalkAuthStore } from '@/stores'
import { useAppContentStore } from '@/stores/appContent'
import { departmentLevelsFromEntity, departmentPathFromEntity } from '@/utils/departments'
import PerformanceReviewDetail from './components/PerformanceReviewDetail.vue'
import { appendUniqueReviewRows, createMobilePaginationState, mobileReviewListParams, resetMobilePagination, showMobileLoadMore, showMobileNoMore, updateMobilePaginationTotal } from './components/mobilePagination'
import { myPerformanceStatuses, myPerformanceStatusSet, normalizeReviewActionResult, statusMeta } from './constants/performanceStatus'

interface CreateTargetNode {
  key: string
  name: string
  count: number
  userIds: string[]
  childMap?: Map<string, CreateTargetNode>
  children: CreateTargetNode[]
  users: PerformanceUser[]
}

interface CreateTargetRow {
  type: 'department' | 'employee'
  key: string
  depth: number
  name?: string
  count?: number
  userIds?: string[]
  expandable?: boolean
  expanded?: boolean
  user?: PerformanceUser
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

function reviewsFromCreatePayload(payload: PerformanceReview | CreateReviewPayload | undefined) {
  if (!payload) {
    return []
  }
  const batch = payload as CreateReviewPayload
  if (Array.isArray(batch.list)) {
    return batch.list
  }
  const review = payload as PerformanceReview
  return review.id ? [review] : []
}

function failedFromCreatePayload(payload: PerformanceReview | CreateReviewPayload | undefined) {
  const batch = payload as CreateReviewPayload | undefined
  return Array.isArray(batch?.failed) ? batch.failed : []
}

function currentMonth() {
  return new Date().toISOString().slice(0, 7)
}

function nextPeriodFromPeriod(period: string) {
  const match = String(period || '').match(/^(\d{4})-(\d{1,2})$/)
  if (!match) {
    return ''
  }
  const year = Number(match[1])
  const month = Number(match[2])
  if (!year || month < 1 || month > 12) {
    return ''
  }
  const next = new Date(year, month, 1)
  return `${next.getFullYear()}-${String(next.getMonth() + 1).padStart(2, '0')}`
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

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const reviews = ref<PerformanceReview[]>([])
const users = ref<PerformanceUser[]>([])
const template = ref<PerformanceTemplate | null>(null)
const reviewListTotal = ref(0)
const loading = ref(false)
const detailLoading = ref(false)
const isMobilePage = ref(resolveMobilePage())
const mineMobilePagination = createMobilePaginationState()
const selectedReviewId = ref('')
const showCreatePopup = ref(false)
const createUserKeyword = ref('')
const createSearchFocused = ref(false)
const createExpandedDeptKeys = ref(new Set<string>())
const createMonthPickerOpen = ref(false)
const createMonthPickerYear = ref(new Date().getFullYear())
const createForm = reactive<CreateReviewRequest>({
  employeeIds: [],
  period: currentMonth(),
})
const emptySteps = ['等待创建', '流程流转', '完成归档']
const listTitle = '我的绩效'
const emptyTitle = '暂无进行中绩效单'
const emptyDesc = '属于你的绩效单只要流程未完成，就会在这里展示。'
const canCreate = computed(() => auth.hasButtonPermission('dingtalk_h5:button:review:create'))
const currentReviews = computed(() => {
  return reviews.value.filter(review => (
    sameUserId(review.employeeId, auth.user?.id)
    && myPerformanceStatusSet.has(String(review.status || ''))
  ))
})
const selectedReview = computed(() => {
  return currentReviews.value.find(review => review.id === selectedReviewId.value) || currentReviews.value[0] || null
})
const listDesc = computed(() => `共 ${reviewListTotal.value || currentReviews.value.length} 条进行中`)
const showMineMobileLoadMore = computed(() => isMobilePage.value && showMobileLoadMore(mineMobilePagination, currentReviews.value.length))
const showMineMobileNoMore = computed(() => isMobilePage.value && showMobileNoMore(mineMobilePagination, currentReviews.value.length))
const createReviewTargetUsers = computed(() => {
  const activeUsers = users.value.filter(user => user.id && Number(user.status ?? user.userStatus ?? 1) === 1)
  if (activeUsers.length > 0) {
    return activeUsers
  }
  return auth.user?.id ? [auth.user as PerformanceUser] : []
})
const createReviewSearchKeyword = computed(() => createUserKeyword.value.trim().toLowerCase())
const filteredCreateReviewTargetUsers = computed(() => {
  const keyword = createReviewSearchKeyword.value
  if (!keyword) {
    return createReviewTargetUsers.value
  }
  const terms = keyword.split(/\s+/).filter(Boolean)
  return createReviewTargetUsers.value.filter((user) => {
    const text = [
      user.id,
      user.name,
      user.account,
      user.mobile,
      user.phone,
      user.position,
      user.department,
      user.departmentLevel1,
      user.departmentLevel2,
      user.departmentLevel3,
      user.departmentLevel4,
      ...(Array.isArray(user.departmentLevels) ? user.departmentLevels : []),
      createTargetUserMeta(user),
    ].filter(Boolean).join(' ').toLowerCase()
    return terms.every(term => text.includes(term))
  })
})
const createTargetUserTree = computed(() => buildCreateTargetUserTree(filteredCreateReviewTargetUsers.value))
const createSearchActive = computed(() => createReviewSearchKeyword.value.length > 0)
const effectiveCreateExpandedDeptKeys = computed(() => {
  if (!createSearchActive.value) {
    return createExpandedDeptKeys.value
  }
  return new Set(collectCreateTargetKeys(createTargetUserTree.value))
})
const createTargetUserTreeRows = computed(() => flattenCreateTargetTree(createTargetUserTree.value, effectiveCreateExpandedDeptKeys.value))
const createTargetUserEmptyText = computed(() => {
  if (createReviewTargetUsers.value.length === 0) {
    return '暂无可创建人员'
  }
  return createReviewSearchKeyword.value ? '没有匹配的人员' : '暂无可创建人员'
})
const createReviewMonthOptions = computed(() => {
  return Array.from({ length: 12 }, (_, index) => ({
    value: index + 1,
    label: `${String(index + 1).padStart(2, '0')}月`,
  }))
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

async function loadMinePageData(options: { append?: boolean } = {}) {
  const append = Boolean(options.append && isMobilePage.value)
  if (!append) {
    resetMobilePagination(mineMobilePagination)
    loading.value = true
  }
  else {
    mineMobilePagination.loadingMore = true
  }
  try {
    const [reviewRes, templateRes] = await Promise.all([
      listReviews({
        scope: 'mine',
        skipHistory: 0,
        statuses: myPerformanceStatuses.join(','),
        ...mobileReviewListParams(isMobilePage.value, mineMobilePagination),
      }),
      getTemplate(),
    ])
    const nextReviews = listFromPayload(reviewRes.data)
    reviews.value = append ? appendUniqueReviewRows(reviews.value, nextReviews) : nextReviews
    reviewListTotal.value = totalFromPayload(reviewRes.data, nextReviews.length)
    updateMobilePaginationTotal(mineMobilePagination, reviewListTotal.value, reviews.value.length)
    template.value = templateRes.data || null
  }
  finally {
    if (append) {
      mineMobilePagination.loadingMore = false
    }
    else {
      loading.value = false
    }
  }
  syncSelectedReview()
}

async function loadMoreMineRows() {
  if (!showMineMobileLoadMore.value || mineMobilePagination.loadingMore) {
    return
  }
  mineMobilePagination.page += 1
  try {
    await loadMinePageData({ append: true })
  }
  catch {
    mineMobilePagination.page = Math.max(mineMobilePagination.page - 1, 1)
    uni.showToast({ title: '加载失败，请重试', icon: 'none' })
  }
}

async function loadCreateUsers() {
  const res = await listUsers()
  users.value = (res.data || []).map(user => ({ ...user, password: '' }))
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
  if (focusedId && currentReviews.value.some(review => review.id === focusedId)) {
    selectedReviewId.value = focusedId
    appContent.clearFocusedReview()
    return
  }
  if (selectedReviewId.value && currentReviews.value.some(review => review.id === selectedReviewId.value)) {
    return
  }
  selectedReviewId.value = currentReviews.value[0]?.id || ''
}

function refreshPageData() {
  return loadMinePageData()
}

function selectReview(review: PerformanceReview) {
  selectedReviewId.value = review.id
}

function departmentText(user: PerformanceUser) {
  return departmentPathFromEntity(user)
}

function createTargetUserMeta(user: PerformanceUser) {
  return [user.position, departmentText(user)].filter(Boolean).join(' · ') || user.id
}

function createTargetDepartmentLevels(user: PerformanceUser) {
  const levels = departmentLevelsFromEntity(user)
  return levels.length > 0 ? levels : ['未设置部门']
}

function ensureCreateTargetNode(map: Map<string, CreateTargetNode>, key: string, name: string) {
  if (!map.has(key)) {
    map.set(key, {
      key,
      name,
      count: 0,
      userIds: [],
      childMap: new Map(),
      children: [],
      users: [],
    })
  }
  return map.get(key) as CreateTargetNode
}

function finalizeCreateTargetNodes(nodes: CreateTargetNode[]): CreateTargetNode[] {
  return nodes
    .sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))
    .map(node => ({
      key: node.key,
      name: node.name,
      count: node.count,
      userIds: [...new Set(node.userIds.filter(Boolean))],
      children: finalizeCreateTargetNodes([...(node.childMap?.values() || [])]),
      users: node.users.slice().sort((left, right) => [left.department, left.name, left.id].filter(Boolean).join('\0').localeCompare([right.department, right.name, right.id].filter(Boolean).join('\0'), 'zh-CN')),
    }))
}

function buildCreateTargetUserTree(users: PerformanceUser[] = []) {
  const root = new Map<string, CreateTargetNode>()
  for (const user of users) {
    const levels = createTargetDepartmentLevels(user)
    let currentMap = root
    let currentNode: CreateTargetNode | null = null
    for (const [index, level] of levels.entries()) {
      const parentKey = currentNode?.key || 'root'
      currentNode = ensureCreateTargetNode(currentMap, `${parentKey}/l${index + 1}:${level}`, level)
      currentNode.count += 1
      if (user.id) {
        currentNode.userIds.push(user.id)
      }
      currentMap = currentNode.childMap as Map<string, CreateTargetNode>
    }
    if (currentNode) {
      currentNode.users.push(user)
    }
  }
  return finalizeCreateTargetNodes([...root.values()])
}

function collectCreateTargetKeys(nodes: CreateTargetNode[]) {
  const keys: string[] = []
  for (const node of nodes) {
    keys.push(node.key)
    keys.push(...collectCreateTargetKeys(node.children))
  }
  return keys
}

function flattenCreateTargetTree(nodes: CreateTargetNode[], expandedKeys: Set<string>, depth = 1): CreateTargetRow[] {
  const rows: CreateTargetRow[] = []
  for (const node of nodes) {
    const hasChildren = node.children.length > 0 || node.users.length > 0
    const expanded = expandedKeys.has(node.key)
    rows.push({
      type: 'department',
      key: node.key,
      depth,
      name: node.name,
      count: node.count,
      userIds: node.userIds,
      expandable: hasChildren,
      expanded,
    })
    if (!expanded) {
      continue
    }
    for (const user of node.users) {
      rows.push({ type: 'employee', key: `${node.key}/user:${user.id}`, depth: depth + 1, user })
    }
    rows.push(...flattenCreateTargetTree(node.children, expandedKeys, depth + 1))
  }
  return rows
}

function createTargetDepartmentUserIds(row: CreateTargetRow) {
  return Array.isArray(row.userIds) ? row.userIds.filter(Boolean) : []
}

function createTargetDepartmentCheckState(row: CreateTargetRow) {
  const ids = createTargetDepartmentUserIds(row)
  if (ids.length === 0) {
    return 'empty'
  }
  const selected = new Set(createForm.employeeIds || [])
  const selectedCount = ids.filter(id => selected.has(id)).length
  if (selectedCount === 0) {
    return 'unchecked'
  }
  return selectedCount === ids.length ? 'checked' : 'indeterminate'
}

function setCreateReviewEmployeeIds(ids: string[] = []) {
  const selected = new Set(ids.map(id => String(id || '').trim()).filter(Boolean))
  createForm.employeeIds = createReviewTargetUsers.value
    .map(user => user.id)
    .filter(id => selected.has(id))
  createForm.employeeId = createForm.employeeIds[0] || ''
}

function normalizeCreateReviewSelection() {
  const targetIds = new Set(createReviewTargetUsers.value.map(user => String(user.id || '').trim()).filter(Boolean))
  const selected = (createForm.employeeIds || [])
    .map(id => String(id || '').trim())
    .filter(id => targetIds.has(id))
  setCreateReviewEmployeeIds(selected)
}

function toggleCreateReviewDepartment(row: CreateTargetRow) {
  const ids = createTargetDepartmentUserIds(row)
  if (ids.length === 0) {
    return
  }
  const selected = new Set(createForm.employeeIds || [])
  const allSelected = ids.every(id => selected.has(id))
  ids.forEach((id) => {
    if (allSelected) {
      selected.delete(id)
    }
    else {
      selected.add(id)
    }
  })
  setCreateReviewEmployeeIds([...selected])
}

function toggleCreateReviewEmployee(id: string | undefined) {
  const key = String(id || '').trim()
  if (!key) {
    return
  }
  const selected = createForm.employeeIds || []
  if (selected.includes(key)) {
    setCreateReviewEmployeeIds(selected.filter(item => item !== key))
    return
  }
  setCreateReviewEmployeeIds([...selected, key])
}

function toggleCreateReviewDept(key: string) {
  const next = new Set(createExpandedDeptKeys.value)
  if (next.has(key)) {
    next.delete(key)
  }
  else {
    next.add(key)
  }
  createExpandedDeptKeys.value = next
}

function createReviewPeriodParts(period: string) {
  const match = String(period || '').match(/^(\d{4})-(\d{1,2})$/)
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

function createReviewMonthText(period: string | undefined) {
  const parts = createReviewPeriodParts(period || '')
  if (!parts) {
    return '请选择月份'
  }
  return `${parts.year}-${String(parts.month).padStart(2, '0')}`
}

function syncCreateReviewMonthPickerYear() {
  const parts = createReviewPeriodParts(createForm.period)
  createMonthPickerYear.value = parts?.year || new Date().getFullYear()
}

function toggleCreateReviewMonthPicker() {
  if (!createMonthPickerOpen.value) {
    syncCreateReviewMonthPickerYear()
  }
  createMonthPickerOpen.value = !createMonthPickerOpen.value
}

function changeCreateReviewMonthPickerYear(delta: number) {
  createMonthPickerYear.value += Number(delta || 0)
}

function selectCreateReviewMonth(month: number) {
  createForm.period = `${createMonthPickerYear.value}-${String(month).padStart(2, '0')}`
  createMonthPickerOpen.value = false
}

function isCreateReviewMonthSelected(month: number) {
  const parts = createReviewPeriodParts(createForm.period)
  return Boolean(parts && parts.year === createMonthPickerYear.value && parts.month === month)
}

function reviewItemClass(review: PerformanceReview) {
  return ['review-item', selectedReviewId.value === review.id ? 'review-item--active' : ''].filter(Boolean).join(' ')
}

function createReviewMonthTriggerClass() {
  return ['field-input', 'review-create-month-trigger', createForm.period ? 'selected' : ''].filter(Boolean).join(' ')
}

function createReviewMonthOptionClass(month: number) {
  return ['review-create-month-option', isCreateReviewMonthSelected(month) ? 'active' : ''].filter(Boolean).join(' ')
}

async function openCreatePopup() {
  if (!canCreate.value) {
    uni.showToast({
      title: '无权限创建考评单',
      icon: 'none',
    })
    return
  }
  createForm.period = currentMonth()
  createForm.employeeIds = []
  createForm.employeeId = ''
  createUserKeyword.value = ''
  createSearchFocused.value = false
  createExpandedDeptKeys.value = new Set()
  createMonthPickerOpen.value = false
  syncCreateReviewMonthPickerYear()
  if (auth.hasApiPermission('dingtalk_h5:api:user:list') && users.value.length === 0) {
    await loadCreateUsers()
  }
  normalizeCreateReviewSelection()
  showCreatePopup.value = true
}

function closeCreatePopup() {
  showCreatePopup.value = false
  createUserKeyword.value = ''
  createSearchFocused.value = false
  createMonthPickerOpen.value = false
}

async function submitCreate() {
  normalizeCreateReviewSelection()
  if (!(createForm.employeeIds || []).length) {
    uni.showToast({
      title: '请选择被考评人',
      icon: 'none',
    })
    return
  }
  if (!createForm.period.trim()) {
    uni.showToast({
      title: '请选择考评月份',
      icon: 'none',
    })
    return
  }
  const employeeIds = createForm.employeeIds || []
  const res = await createReview({
    employeeIds,
    period: createForm.period.trim(),
    nextPeriod: nextPeriodFromPeriod(createForm.period.trim()),
  })
  const created = reviewsFromCreatePayload(res.data)
  const failed = failedFromCreatePayload(res.data)
  closeCreatePopup()
  if (created[0]?.id) {
    selectedReviewId.value = created[0].id
  }
  await loadMinePageData()
  const createdCount = created.length || employeeIds.length
  const failedCount = failed.length || 0
  uni.showToast({
    title: failedCount > 0 ? `已创建 ${createdCount} 张，${failedCount} 张失败` : `已创建 ${createdCount || 1} 张`,
    icon: 'success',
  })
}

watch(
  () => [appContent.currentKey, appContent.refreshTick],
  () => {
    if (appContent.currentKey === 'performance:mine') {
      void loadMinePageData()
    }
  },
  { immediate: true },
)

watch(currentReviews, syncSelectedReview)
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
        <u-button v-if="canCreate" custom-class="dt-btn dt-btn-primary page-action-btn mine-create-btn" @click="openCreatePopup">
          创建
        </u-button>
      </view>
    </view>

    <view v-if="loading" class="panel performance-loading-panel">
      <u-loading mode="circle" />
      <text>加载中...</text>
    </view>

    <template v-else>
      <view class="review-detail-only">
        <view v-if="currentReviews.length > 1" class="panel my-performance-switcher">
          <view class="list-panel__head">
            <text class="panel-title">
              进行中绩效
            </text>
            <text class="count-pill">
              {{ currentReviews.length }} 条
            </text>
          </view>
          <view class="review-list">
            <u-button
              v-for="review in currentReviews"
              :key="reviewKey(review)"
              :custom-class="reviewItemClass(review)"
              @click="selectReview(review)"
            >
              <view class="review-item__main">
                <text class="review-item__title">
                  {{ review.period || '-' }} 月度考评
                </text>
              </view>
              <u-tag :text="reviewStatusText(review)" :type="reviewStatusType(review)" mode="light" />
            </u-button>
          </view>
          <view v-if="showMineMobileLoadMore" class="mobile-list-pagination">
            <u-button
              custom-class="dt-btn dt-btn-light mobile-list-pagination__button"
              :disabled="mineMobilePagination.loadingMore"
              @click="loadMoreMineRows"
            >
              {{ mineMobilePagination.loadingMore ? '加载中...' : '加载更多' }}
            </u-button>
          </view>
          <view v-else-if="showMineMobileNoMore" class="mobile-list-pagination mobile-list-pagination--done">
            没有更多了
          </view>
        </view>
        <PerformanceReviewDetail
          v-if="selectedReview"
          :review="selectedReview"
          :users="users"
          :template="template"
          :grades="grades"
          action-scope="mine"
          :detail-loading="detailLoading"
          :submit-review-action="submitReviewAction"
          @action-success="refreshPageData"
        />
        <view v-else class="panel detail-panel">
          <view class="performance-empty">
            <view class="performance-empty-visual">
              <view class="performance-empty-icon">
                绩
              </view>
              <view class="performance-empty-line wide" />
              <view class="performance-empty-line" />
              <view class="performance-empty-dots">
                <text />
                <text />
                <text />
              </view>
            </view>
            <text class="performance-empty-title">
              {{ emptyTitle }}
            </text>
            <text class="performance-empty-desc">
              {{ emptyDesc }}
            </text>
            <view class="performance-empty-steps">
              <text v-for="(item, index) in emptySteps" :key="item" class="performance-empty-step">
                {{ index + 1 }}. {{ item }}
              </text>
            </view>
          </view>
        </view>
      </view>
    </template>

    <view v-if="showCreatePopup" class="review-create-modal" @click="closeCreatePopup">
      <view class="review-create-card" :class="{ 'review-create-card--keyboard': createSearchFocused }" @click.stop="createMonthPickerOpen = false">
        <view class="review-create-head">
          <view>
            <text class="review-create-title">
              新建考评单
            </text>
            <text class="review-create-desc">
              选择被考评人和考评月份，支持按部门多选创建
            </text>
          </view>
          <u-button custom-class="review-create-close" @click="closeCreatePopup">
            ×
          </u-button>
        </view>

        <view class="review-create-form">
          <view class="review-create-field review-create-field-users">
            <text class="review-create-label">
              被考评人
            </text>
            <view class="create-target-search">
              <u-input
                v-model="createUserKeyword"
                custom-class="field-input create-target-search-input"
                :border="true"
                :adjust-position="false"
                :cursor-spacing="12"
                confirm-type="search"
                placeholder="搜索姓名/账号/部门/岗位"
                @focus="createSearchFocused = true; createMonthPickerOpen = false"
                @blur="createSearchFocused = false"
              />
              <u-button v-if="createUserKeyword" custom-class="create-target-search-clear" @click="createUserKeyword = ''">
                清空
              </u-button>
            </view>

            <view class="department-user-tree">
              <view
                v-for="row in createTargetUserTreeRows"
                :key="row.key"
                class="create-target-row"
                :class="[row.type, `depth-${row.depth}`]"
              >
                <view
                  v-if="row.type === 'department'"
                  class="create-target-dept-head"
                  :class="{ expanded: row.expanded }"
                  @click="toggleCreateReviewDept(row.key)"
                >
                  <view
                    class="create-target-dept-check"
                    :class="{
                      'checked': createTargetDepartmentCheckState(row) === 'checked',
                      'create-target-dept-check-indeterminate': createTargetDepartmentCheckState(row) === 'indeterminate',
                    }"
                    @click.stop="toggleCreateReviewDepartment(row)"
                  >
                    <text v-if="createTargetDepartmentCheckState(row) === 'checked'">
                      ✓
                    </text>
                    <text v-else-if="createTargetDepartmentCheckState(row) === 'indeterminate'">
                      -
                    </text>
                  </view>
                  <view class="create-target-dept-title">
                    <text class="create-target-dept-chevron" :class="{ expanded: row.expanded }" />
                    <text class="create-target-dept-name">
                      {{ row.name }}
                    </text>
                  </view>
                  <text class="create-target-dept-count">
                    {{ createTargetDepartmentUserIds(row).length }} 人
                  </text>
                </view>

                <view
                  v-else-if="row.user"
                  class="create-target-user-tree"
                  :class="{ selected: (createForm.employeeIds || []).includes(row.user.id) }"
                  @click="toggleCreateReviewEmployee(row.user.id)"
                >
                  <view class="create-target-check">
                    <text v-if="(createForm.employeeIds || []).includes(row.user.id)">
                      ✓
                    </text>
                  </view>
                  <view class="create-target-user-main">
                    <text class="create-target-user-name">
                      {{ row.user.name || row.user.id }}
                    </text>
                    <text class="create-target-user-meta">
                      {{ createTargetUserMeta(row.user) }}
                    </text>
                  </view>
                </view>
              </view>
              <view v-if="createTargetUserTreeRows.length === 0" class="create-target-empty">
                {{ createTargetUserEmptyText }}
              </view>
            </view>
          </view>

          <view class="review-create-inline-fields single">
            <view class="review-create-field">
              <text class="review-create-label">
                考评月份
              </text>
              <view class="review-create-month-picker">
                <u-button
                  :custom-class="createReviewMonthTriggerClass()"
                  @click.stop="toggleCreateReviewMonthPicker"
                >
                  <text class="review-create-month-text">
                    {{ createReviewMonthText(createForm.period) }}
                  </text>
                  <text class="review-create-month-arrow" :class="{ open: createMonthPickerOpen }" />
                </u-button>
                <view v-if="createMonthPickerOpen" class="review-create-month-dropdown" @click.stop>
                  <view class="review-create-month-head">
                    <u-button custom-class="review-create-month-nav" @click="changeCreateReviewMonthPickerYear(-1)">
                      ‹
                    </u-button>
                    <text class="review-create-month-year">
                      {{ createMonthPickerYear }}年
                    </text>
                    <u-button custom-class="review-create-month-nav" @click="changeCreateReviewMonthPickerYear(1)">
                      ›
                    </u-button>
                  </view>
                  <view class="review-create-month-grid">
                    <u-button
                      v-for="month in createReviewMonthOptions"
                      :key="month.value"
                      :custom-class="createReviewMonthOptionClass(month.value)"
                      @click="selectCreateReviewMonth(month.value)"
                    >
                      {{ month.label }}
                    </u-button>
                  </view>
                </view>
              </view>
            </view>
          </view>
        </view>

        <view class="review-create-actions">
          <text class="review-create-selected-count">
            已选 {{ (createForm.employeeIds || []).length }} 人
          </text>
          <u-button custom-class="dt-btn dt-btn-light" @click="closeCreatePopup">
            取消
          </u-button>
          <u-button custom-class="dt-btn dt-btn-primary" :disabled="loading" :loading="loading" @click="submitCreate">
            创建
          </u-button>
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped src="./components/performance-page.scss"></style>
