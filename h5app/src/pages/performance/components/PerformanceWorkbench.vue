<script setup lang="ts">
import type { ReviewActionKey } from '../constants/performancePermissions'
import type { PerformanceReview, PerformanceReviewListPayload } from '@/types/dingtalk-h5'
import type { WorkflowTaskSummary } from '@/types/workflow'
import { computed, ref, watch } from 'vue'
import { listReviews } from '@/api/dingtalk-h5'
import { listWorkflowTasks } from '@/api/workflow'
import { useAppContentStore } from '@/stores/appContent'
import { useDingtalkAuthStore } from '@/stores/dingtalkAuth'
import { reviewActionApiPermissions, reviewActionButtonPermissions } from '../constants/performancePermissions'
import { statusMeta } from '../constants/performanceStatus'

type TodoFilter = 'all' | 'workflow' | 'performance'

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

function firstText(...values: Array<unknown>) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

function reviewStatusText(review: PerformanceReview) {
  return statusMeta[String(review.status || '')]?.label || String(review.status || '未知')
}

function reviewStatusType(review: PerformanceReview) {
  return statusMeta[String(review.status || '')]?.type || 'info'
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
const workflowTasks = ref<WorkflowTaskSummary[]>([])
const workflowPendingCount = ref(0)
const loading = ref(false)
const todoFilter = ref<TodoFilter>('all')

const todoStatuses = new Set(['draft', 'manager_review', 'hrbp_review', 'employee_confirm', 'hr_final'])
const todoStatusMenuKeys: Record<string, string> = {
  draft: 'performance:mine',
  manager_review: 'performance:manager',
  hrbp_review: 'performance:hrbp',
  employee_confirm: 'performance:mine',
  hr_final: 'performance:hrbp',
}
const todoStatusActions: Record<string, ReviewActionKey[]> = {
  draft: ['save-self', 'submit-self'],
  manager_review: ['return-employee', 'submit-manager'],
  hrbp_review: ['return-manager', 'submit-hrbp', 'withdraw'],
  employee_confirm: ['confirm-result', 'dispute-result', 'withdraw'],
  hr_final: ['return-hrbp', 'finalize', 'withdraw'],
}
const currentUserTodos = computed(() => reviews.value.filter(isCurrentUserTodo))
const todos = computed(() => currentUserTodos.value.filter(canOpenTodo))
const canViewWorkflowTodos = computed(() => (
  auth.hasMenuPermission('workflow')
  && auth.hasApiPermission('dingtalk_h5:api:workflow:view')
))
const totalTodoCount = computed(() => todos.value.length + workflowPendingCount.value)
const hasTodos = computed(() => totalTodoCount.value > 0)
const todoToneType = computed(() => hasTodos.value ? 'warning' : 'success')
const todoToneColor = computed(() => hasTodos.value ? 'var(--u-type-warning)' : 'var(--u-type-success)')
const todoFilters = computed<Array<{ key: TodoFilter, label: string, count: number }>>(() => [
  { key: 'all', label: '全部', count: totalTodoCount.value },
  { key: 'workflow', label: '流程审批', count: workflowPendingCount.value },
  { key: 'performance', label: '绩效任务', count: todos.value.length },
])
const visibleWorkflowTasks = computed(() => todoFilter.value === 'performance' ? [] : workflowTasks.value)
const visiblePerformanceTodos = computed(() => todoFilter.value === 'workflow' ? [] : todos.value)
const hasVisibleTodos = computed(() => visibleWorkflowTasks.value.length > 0 || visiblePerformanceTodos.value.length > 0)

function isCurrentUserTodo(review: PerformanceReview) {
  const status = String(review.status || '')
  const currentUserId = auth.user?.id
  if (!todoStatuses.has(status) || !currentUserId) {
    return false
  }
  const currentAssigneeId = firstText(
    review.currentAssigneeId,
    review.assigneeId,
    review.currentHandlerId,
    review.currentHandlerUserId,
  )
  if (currentAssigneeId) {
    return sameUserId(currentAssigneeId, currentUserId)
  }
  if (status === 'draft' || status === 'employee_confirm') {
    return sameUserId(review.employeeId, currentUserId)
  }
  if (status === 'manager_review') {
    return sameUserId(review.managerId, currentUserId)
  }
  if (status === 'hrbp_review' || status === 'hr_final') {
    return sameUserId(review.hrbpReviewerId, currentUserId) || sameUserId(review.hrbpId, currentUserId)
  }
  return false
}

function userName(id: unknown) {
  const key = String(id || '').trim()
  if (!key) {
    return '-'
  }
  if (sameUserId(auth.user?.id, key)) {
    return auth.user?.name || key
  }
  return key
}

function reviewEmployeeName(review: PerformanceReview) {
  const resolvedName = userName(review.employeeId)
  return firstText(review.employeeName, resolvedName === '-' ? '' : resolvedName, '该员工')
}

function todoTitlePrefix(review: PerformanceReview) {
  const period = review.period || '-'
  return `${reviewEmployeeName(review)} ${period} 月度考评`
}

function todoTitle(review: PerformanceReview) {
  const prefix = todoTitlePrefix(review)
  if (review.status === 'draft') {
    return `${prefix}待填写`
  }
  if (review.status === 'manager_review') {
    return `${prefix}待上级评价`
  }
  if (review.status === 'hrbp_review') {
    return `${prefix}待 HRBP 评价`
  }
  if (review.status === 'employee_confirm') {
    return `${prefix}待确认`
  }
  if (review.status === 'hr_final') {
    return `${prefix}待归档`
  }
  return `${prefix}待处理`
}

function todoHandlerText(review: PerformanceReview) {
  return `当前处理人 ${currentAssignee(review, userName)}`
}

function todoMeta(review: PerformanceReview) {
  return `${review.department || '-'} · ${todoHandlerText(review)}`
}

function todoAction(review: PerformanceReview) {
  if (review.status === 'draft') {
    return '填写'
  }
  if (review.status === 'manager_review' || review.status === 'hrbp_review') {
    return '评价'
  }
  if (review.status === 'employee_confirm') {
    return '确认'
  }
  if (review.status === 'hr_final') {
    return '归档'
  }
  return '查看'
}

function canUseReviewAction(action: ReviewActionKey) {
  const apiPermission = reviewActionApiPermissions[action]
  const buttonPermission = reviewActionButtonPermissions[action]
  return auth.hasApiPermission(apiPermission) && auth.hasButtonPermission(buttonPermission)
}

function canOpenTodo(review: PerformanceReview) {
  const targetKey = todoStatusMenuKeys[String(review.status || '')]
  if (!targetKey || !auth.hasMenuPermission(targetKey)) {
    return false
  }
  return (todoStatusActions[String(review.status || '')] || []).some(canUseReviewAction)
}

function showToast(title: string) {
  uni.showToast({
    title,
    icon: 'none',
  })
}

function openTodo(review: PerformanceReview) {
  if (!canOpenTodo(review)) {
    showToast('无权限访问')
    return
  }
  const nextView = todoStatusMenuKeys[String(review.status || '')] || 'performance:mine'
  appContent.switchContent(nextView, review.id)
}

function openWorkflowTodos() {
  if (!canViewWorkflowTodos.value) {
    showToast('无权限访问')
    return
  }
  appContent.focusWorkflowTab('pending')
  appContent.switchContent('workflow')
}

function workflowTodoTitle(task: WorkflowTaskSummary) {
  return firstText(task.definitionName, task.nodeName, '流程审批')
}

function workflowTodoStarter(task: WorkflowTaskSummary) {
  return firstText(task.starterName, task.starterId, '未知用户')
}

function workflowTodoMeta(task: WorkflowTaskSummary) {
  return `发起人：${workflowTodoStarter(task)} · 当前节点：${firstText(task.nodeName, '待处理')}`
}

function setTodoFilter(filter: TodoFilter) {
  todoFilter.value = filter
}

async function loadPerformanceTodos() {
  const res = await listReviews({
    scope: 'dashboard',
    pageSize: 100,
    skipHistory: 1,
  })
  reviews.value = listFromPayload(res.data)
}

async function loadWorkflowTodos() {
  workflowTasks.value = []
  workflowPendingCount.value = 0
  if (!canViewWorkflowTodos.value)
    return
  const res = await listWorkflowTasks({ status: 'pending', page: 1, pageSize: 100 })
  workflowTasks.value = Array.isArray(res.data?.list) ? res.data.list : []
  workflowPendingCount.value = Math.max(0, Number(res.data?.total || 0))
}

async function loadWorkbenchData() {
  loading.value = true
  try {
    await Promise.allSettled([loadPerformanceTodos(), loadWorkflowTodos()])
  }
  finally {
    loading.value = false
  }
}

watch(
  () => [appContent.currentKey, appContent.refreshTick],
  () => {
    if (appContent.currentKey === 'dashboard') {
      void loadWorkbenchData()
    }
  },
  { immediate: true },
)
</script>

<template>
  <view class="workbench">
    <view class="workbench__inner">
      <view class="workbench__header">
        <view class="workbench__heading">
          <view class="workbench__heading-icon">
            <u-icon name="clock" size="18px" :color="todoToneColor" />
          </view>
          <text class="workbench__title">
            我的待办
          </text>
          <u-tag custom-class="workbench__total-tag" :text="`${totalTodoCount} 条`" :type="todoToneType" mode="light" size="mini" />
        </view>

        <view class="workbench__filters">
          <view
            v-for="filter in todoFilters"
            :key="filter.key"
            class="workbench__filter"
            :class="{ 'workbench__filter--active': todoFilter === filter.key }"
            hover-class="workbench__filter--hover"
            @click="setTodoFilter(filter.key)"
          >
            <text>{{ filter.label }}</text>
            <text v-if="filter.count > 0" class="workbench__filter-count">
              {{ filter.count }}
            </text>
          </view>
        </view>
      </view>

      <view class="workbench__content">
        <view v-if="loading" class="loading">
          <u-loading mode="circle" size="24px" />
          <text>加载中...</text>
        </view>

        <view v-else-if="hasVisibleTodos" class="todo-list">
          <view v-for="task in visibleWorkflowTasks" :key="task.id" class="todo-item todo-item--workflow" @click="openWorkflowTodos">
            <view class="todo-item__icon todo-item__icon--workflow">
              <u-icon name="file-text" size="20px" color="#0f766e" />
            </view>
            <view class="todo-item__main">
              <view class="todo-item__title-row">
                <text class="todo-item__title">
                  {{ workflowTodoTitle(task) }}
                </text>
                <u-tag custom-class="todo-source-tag" text="流程审批" type="primary" mode="light" size="mini" />
                <u-tag custom-class="todo-status-tag" text="待处理" type="warning" mode="light" size="mini" />
              </view>
              <text class="todo-item__meta">
                {{ workflowTodoMeta(task) }}
              </text>
            </view>
            <view class="todo-item__side">
              <u-button custom-class="todo-item__action-button" type="primary" size="mini" plain @click.stop="openWorkflowTodos">
                <text>处理</text>
                <u-icon name="arrow-right" size="14px" color="var(--u-type-primary)" />
              </u-button>
            </view>
          </view>
          <view v-for="review in visiblePerformanceTodos" :key="review.id" class="todo-item todo-item--performance" @click="openTodo(review)">
            <view class="todo-item__icon todo-item__icon--performance">
              <u-icon name="order" size="20px" color="#b45309" />
            </view>
            <view class="todo-item__main">
              <view class="todo-item__title-row">
                <text class="todo-item__title">
                  {{ todoTitle(review) }}
                </text>
                <u-tag custom-class="todo-source-tag" text="绩效任务" type="success" mode="light" size="mini" />
                <u-tag
                  custom-class="todo-status-tag"
                  :text="reviewStatusText(review)"
                  :type="reviewStatusType(review)"
                  mode="light"
                  size="mini"
                />
                <text class="todo-item__handler todo-item__handler--mobile">
                  {{ todoHandlerText(review) }}
                </text>
              </view>
              <text class="todo-item__meta todo-item__meta--desktop">
                {{ todoMeta(review) }}
              </text>
            </view>
            <view class="todo-item__side">
              <u-button custom-class="todo-item__action-button" type="primary" size="mini" plain @click.stop="openTodo(review)">
                <text>{{ todoAction(review) }}</text>
                <u-icon name="arrow-right" size="14px" color="var(--u-type-primary)" />
              </u-button>
            </view>
          </view>
        </view>

        <view v-else class="empty">
          <u-icon name="checkmark-circle" size="36px" color="var(--u-type-success)" />
          <text class="empty__title">
            {{ hasTodos ? '当前分类没有待办' : '当前没有待办' }}
          </text>
          <text class="empty__desc">
            需要处理的绩效单或流程审批会自动出现在这里。
          </text>
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.workbench {
  width: 100%;
}

.workbench__inner {
  width: min(var(--app-pc-content-max-width, 1080px), 100%);
  margin: 0 auto;
}

.workbench__header {
  min-height: 72px;
  padding: 0 4px;
  border-bottom: 1px solid $u-border-color;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 28px;
}

.workbench__heading {
  min-height: 72px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.workbench__heading-icon {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff7e6;
}

.workbench__title {
  color: $u-main-color;
  font-size: 20px;
  font-weight: 700;
}

.workbench__filters {
  display: flex;
  align-items: flex-end;
  gap: 28px;
}

.workbench__filter {
  position: relative;
  min-height: 48px;
  padding: 0 2px;
  display: flex;
  align-items: center;
  gap: 7px;
  color: $u-content-color;
  font-size: 14px;
  cursor: pointer;

  &::after {
    position: absolute;
    right: 0;
    bottom: -1px;
    left: 0;
    height: 2px;
    background: var(--u-type-primary);
    content: '';
    opacity: 0;
  }
}

.workbench__filter--hover,
.workbench__filter--active {
  color: $u-main-color;
}

.workbench__filter--active {
  font-weight: 600;

  &::after {
    opacity: 1;
  }
}

.workbench__filter-count {
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: $u-bg-gray-light;
  color: $u-tips-color;
  font-size: 12px;
  line-height: 20px;
}

.workbench__filter--active .workbench__filter-count {
  background: var(--u-type-primary-light);
  color: var(--u-type-primary);
}

.workbench__content {
  padding-top: 20px;
}

.loading,
.empty {
  min-height: 280px;
  padding: 48px 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: $u-tips-color;
  font-size: 13px;
  text-align: center;
}

.empty__title {
  color: $u-main-color;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.4;
}

.empty__desc {
  max-width: 420px;
  color: $u-tips-color;
  font-size: 14px;
  line-height: 1.6;
}

.todo-list {
  overflow: hidden;
  border: 1px solid $u-border-color;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  background: $u-bg-white;
}

.todo-item {
  min-height: 92px;
  padding: 18px 20px;
  border-bottom: 1px solid $u-border-color;
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) auto;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: background-color 0.16s ease;

  &:last-child {
    border-bottom: 0;
  }

  &:hover,
  &:active {
    background: $u-bg-gray-light;
  }
}

.todo-item__icon {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.todo-item__icon--workflow {
  background: #e8f7f4;
}

.todo-item__icon--performance {
  background: #fff4df;
}

.todo-item__main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.todo-item__title-row {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.todo-item__title {
  color: $u-main-color;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.35;
}

.workbench__total-tag,
.todo-source-tag,
.todo-status-tag,
:deep(.workbench__total-tag),
:deep(.todo-source-tag),
:deep(.todo-status-tag) {
  flex: 0 0 auto;
  width: auto;
  height: 24px;
  min-height: 24px;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 22px;
  white-space: nowrap;
  box-sizing: border-box;
}

.todo-item__handler {
  display: none;
  color: $u-tips-color;
  font-size: 23rpx;
  line-height: 1.4;
}

.todo-item__meta {
  color: $u-tips-color;
  font-size: 13px;
  line-height: 1.4;
}

.todo-item__meta--mobile {
  display: none;
}

.todo-item__side {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
}

.todo-item__action-button,
:deep(.todo-item__action-button) {
  width: 88px;
  height: 34px;
  min-height: 34px;
  margin: 0;
  padding: 0 12px;
  border-radius: 4px;
  gap: 4px;
  font-size: 13px;
  line-height: 32px;
}

@media (max-width: 768px) {
  .workbench__header {
    min-height: 0;
    padding: 0;
    flex-direction: column;
    align-items: stretch;
    gap: 0;
  }

  .workbench__heading {
    min-height: 96rpx;
  }

  .workbench__title {
    font-size: 32rpx;
  }

  .workbench__heading-icon {
    width: 60rpx;
    height: 60rpx;
  }

  .workbench__filters {
    gap: 36rpx;
    overflow-x: auto;
  }

  .workbench__filter {
    min-height: 76rpx;
    flex: 0 0 auto;
    font-size: 26rpx;
  }

  .workbench__content {
    padding-top: 24rpx;
  }

  .todo-list {
    border-radius: 8rpx;
  }

  .todo-item {
    min-height: 148rpx;
    padding: 24rpx;
    grid-template-columns: 72rpx minmax(0, 1fr);
    gap: 18rpx;
  }

  .todo-item__icon {
    width: 72rpx;
    height: 72rpx;
  }

  .todo-item__title {
    font-size: 28rpx;
  }

  .todo-item__meta {
    font-size: 23rpx;
  }

  .todo-item__side {
    grid-column: 2;
    justify-content: flex-end;
  }

  .todo-item__action-button,
  :deep(.todo-item__action-button) {
    width: 144rpx;
    height: 56rpx;
    min-height: 56rpx;
    padding: 0 16rpx;
    font-size: 24rpx;
    line-height: 56rpx;
  }

  .todo-item__meta--desktop {
    display: none;
  }

  .todo-item__handler--mobile {
    display: inline;
  }
}
</style>
