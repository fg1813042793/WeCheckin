<script setup lang="ts">
import type {
  UserFeedbackOverview,
  UserFeedbackStatus,
  UserFeedbackSummary,
} from '@/api/user-feedback'
import { useLocale } from 'uview-pro'
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { getUserFeedbackOverview, listUserFeedbacks } from '@/api/user-feedback'
import { useAppContentStore } from '@/stores'
import {
  FEEDBACK_CONTENT_KEY,
  FEEDBACK_CREATE_CONTENT_KEY,
  feedbackDetailContentKey,
} from '../feedback-route-keys'
import FeedbackList from './FeedbackList.vue'
import FeedbackStatusOverview from './FeedbackStatusOverview.vue'

type FeedbackStatusFilter = UserFeedbackStatus | ''

const appContent = useAppContentStore()
const { t } = useLocale('feedback')
const overview = reactive<UserFeedbackOverview>({
  pending: 0,
  processing: 0,
  resolved: 0,
  closed: 0,
})
const filters = reactive({
  keyword: '',
  status: '' as FeedbackStatusFilter,
})
const appliedFilters = reactive({
  keyword: '',
  status: '' as FeedbackStatusFilter,
})
const feedbacks = ref<UserFeedbackSummary[]>([])
const overviewLoading = ref(false)
const overviewError = ref(false)
const listLoading = ref(false)
const listError = ref(false)
const page = ref(1)
const pageSize = 12
const total = ref(0)
let overviewRequestSequence = 0
let listRequestSequence = 0

const statusOptions = computed(() => [
  { value: '' as const, label: t('allStatuses') },
  ...(['pending', 'processing', 'resolved', 'closed'] as UserFeedbackStatus[]).map(status => ({
    value: status,
    label: t(`statuses.${status}`),
  })),
])
const sortedFeedbacks = computed(() => [...feedbacks.value].sort((left, right) => (
  Number(right.lastActivityAt || 0) - Number(left.lastActivityAt || 0)
)))

function nonNegativeInteger(value: unknown) {
  return Math.max(0, Math.floor(Number(value || 0)))
}

function resetOverview(next?: Partial<UserFeedbackOverview>) {
  overview.pending = nonNegativeInteger(next?.pending)
  overview.processing = nonNegativeInteger(next?.processing)
  overview.resolved = nonNegativeInteger(next?.resolved)
  overview.closed = nonNegativeInteger(next?.closed)
}

async function loadOverview() {
  const requestSequence = ++overviewRequestSequence
  overviewLoading.value = true
  overviewError.value = false
  try {
    const response = await getUserFeedbackOverview()
    if (requestSequence !== overviewRequestSequence)
      return
    resetOverview(response?.data)
  }
  catch {
    if (requestSequence === overviewRequestSequence)
      overviewError.value = true
  }
  finally {
    if (requestSequence === overviewRequestSequence)
      overviewLoading.value = false
  }
}

async function loadFeedbacks() {
  const requestSequence = ++listRequestSequence
  listLoading.value = true
  listError.value = false
  try {
    const response = await listUserFeedbacks({
      keyword: appliedFilters.keyword || undefined,
      status: appliedFilters.status || undefined,
      page: page.value,
      pageSize,
    })
    if (requestSequence !== listRequestSequence)
      return
    const payload = response?.data
    feedbacks.value = Array.isArray(payload?.list) ? payload.list : []
    total.value = nonNegativeInteger(payload?.total)
    page.value = Math.max(1, nonNegativeInteger(payload?.page) || page.value)
  }
  catch {
    if (requestSequence === listRequestSequence)
      listError.value = true
  }
  finally {
    if (requestSequence === listRequestSequence)
      listLoading.value = false
  }
}

async function refreshFeedbackCenter() {
  await Promise.all([loadOverview(), loadFeedbacks()])
}

function applyFilters() {
  appliedFilters.keyword = filters.keyword.trim()
  appliedFilters.status = filters.status
  page.value = 1
  void loadFeedbacks()
}

function resetFilters() {
  filters.keyword = ''
  filters.status = ''
  appliedFilters.keyword = ''
  appliedFilters.status = ''
  page.value = 1
  void loadFeedbacks()
}

function selectFilterStatus(status: FeedbackStatusFilter) {
  filters.status = status
}

function selectStatus(status: UserFeedbackStatus) {
  filters.status = status
  appliedFilters.status = status
  appliedFilters.keyword = filters.keyword.trim()
  page.value = 1
  void Promise.all([loadOverview(), loadFeedbacks()])
}

function changePage(nextPage: number) {
  page.value = Math.max(1, Number(nextPage || 1))
  void loadFeedbacks()
}

function openCreate() {
  appContent.openDynamicTab({
    key: FEEDBACK_CREATE_CONTENT_KEY,
    label: t('create'),
    icon: 'plus-circle',
    path: `/pages/index/index?view=${encodeURIComponent(FEEDBACK_CREATE_CONTENT_KEY)}`,
  })
}

function openFeedback(item: UserFeedbackSummary) {
  const key = feedbackDetailContentKey(String(item.id))
  if (!key)
    return
  appContent.openDynamicTab({
    key,
    label: item.feedbackNo || t('detail'),
    icon: 'chat',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  })
}

function handleVisibilityChange() {
  // #ifdef H5
  if (document.visibilityState === 'visible' && appContent.currentKey === FEEDBACK_CONTENT_KEY)
    void refreshFeedbackCenter()
  // #endif
}

watch(
  () => [appContent.currentKey, appContent.refreshTick] as const,
  ([currentKey]) => {
    if (currentKey === FEEDBACK_CONTENT_KEY)
      void refreshFeedbackCenter()
  },
  { immediate: true },
)

onMounted(() => {
  // #ifdef H5
  document.addEventListener('visibilitychange', handleVisibilityChange)
  // #endif
})

onBeforeUnmount(() => {
  overviewRequestSequence += 1
  listRequestSequence += 1
  // #ifdef H5
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  // #endif
})
</script>

<template>
  <view class="feedback-center app-layout-page">
    <view class="app-page-header">
      <view class="app-page-header__copy">
        <text class="app-page-header__title">
          {{ t('title') }}
        </text>
        <text class="app-page-header__description">
          {{ t('description') }}
        </text>
      </view>
      <view class="app-page-header__actions">
        <u-button custom-class="feedback-center__create" type="primary" size="small" @click="openCreate">
          <view class="feedback-center__button-content">
            <u-icon name="plus" size="15px" color="#ffffff" />
            <text>{{ t('create') }}</text>
          </view>
        </u-button>
      </view>
    </view>

    <FeedbackStatusOverview
      :active-status="appliedFilters.status"
      :error="overviewError"
      :loading="overviewLoading"
      :overview="overview"
      @retry="loadOverview"
      @select="selectStatus"
    />

    <view class="app-filter-bar feedback-center__filters">
      <view class="app-filter-field feedback-center__keyword-field">
        <text class="app-filter-field__label">
          {{ t('keyword') }}
        </text>
        <u-input
          v-model="filters.keyword"
          custom-class="app-filter-control feedback-center__keyword"
          type="text"
          :maxlength="100"
          :placeholder="t('keywordPlaceholder')"
          :disabled="listLoading"
          :border="true"
          @keyup.enter="applyFilters"
        />
      </view>

      <view class="app-filter-field feedback-center__status-field">
        <text class="app-filter-field__label">
          {{ t('status') }}
        </text>
        <scroll-view class="feedback-center__status-scroll" scroll-x>
          <view class="feedback-center__status-segments" role="group">
            <view
              v-for="option in statusOptions"
              :key="option.value || 'all'"
              class="feedback-center__status-option"
              :class="{ 'feedback-center__status-option--active': filters.status === option.value }"
              role="button"
              tabindex="0"
              @click="selectFilterStatus(option.value)"
              @keyup.enter="selectFilterStatus(option.value)"
            >
              {{ option.label }}
            </view>
          </view>
        </scroll-view>
      </view>

      <view class="app-filter-actions feedback-center__filter-actions">
        <u-button custom-class="feedback-center__filter-button" size="small" plain :disabled="listLoading" @click="resetFilters">
          <view class="feedback-center__button-content">
            <u-icon name="reload" size="14px" color="var(--app-text-secondary-color)" />
            <text>{{ t('reset') }}</text>
          </view>
        </u-button>
        <u-button custom-class="feedback-center__filter-button" size="small" type="primary" :loading="listLoading" @click="applyFilters">
          <view class="feedback-center__button-content">
            <u-icon name="search" size="14px" color="#ffffff" />
            <text>{{ t('search') }}</text>
          </view>
        </u-button>
      </view>
    </view>

    <view class="feedback-center__list-head">
      <text class="feedback-center__list-title">
        {{ t('listTitle') }}
      </text>
      <text class="feedback-center__list-total">
        {{ t('total', { total }) }}
      </text>
    </view>

    <FeedbackList
      :error="listError"
      :items="sortedFeedbacks"
      :loading="listLoading"
      :page="page"
      :page-size="pageSize"
      :total="total"
      @open="openFeedback"
      @page-change="changePage"
      @retry="loadFeedbacks"
    />
  </view>
</template>

<style lang="scss" scoped>
.feedback-center {
  min-width: 0;
}

.feedback-center__create,
.feedback-center__filter-button,
:deep(.feedback-center__create),
:deep(.feedback-center__filter-button) {
  width: auto;
  min-width: 88px;
  height: var(--app-control-height);
  min-height: var(--app-control-height);
  margin: 0;
  padding: 0 14px;
  border-radius: var(--app-control-radius);
}

.feedback-center__button-content {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
}

.feedback-center__filters {
  grid-template-columns: minmax(220px, 1.1fr) minmax(420px, 2fr) auto;
  margin-bottom: 18px;
}

.feedback-center__keyword-field,
.feedback-center__status-field {
  min-width: 0;
}

:deep(.feedback-center__keyword .u-input) {
  min-height: var(--app-control-height);
  padding: 0 10px !important;
  border-color: var(--app-border-strong-color) !important;
  border-radius: var(--app-control-radius) !important;
  box-sizing: border-box;
  font-size: 13px;
}

.feedback-center__status-scroll {
  width: 100%;
  min-width: 0;
  white-space: nowrap;
}

.feedback-center__status-segments {
  min-width: max-content;
  height: var(--app-control-height);
  padding: 3px;
  border: 1px solid var(--app-border-strong-color);
  border-radius: var(--app-control-radius);
  display: inline-flex;
  align-items: center;
  gap: 2px;
  background: var(--app-surface-subtle-color);
  box-sizing: border-box;
}

.feedback-center__status-option {
  height: 28px;
  padding: 0 11px;
  border-radius: 3px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--app-text-secondary-color);
  font-size: 12px;
  line-height: 28px;
  cursor: pointer;
  box-sizing: border-box;
}

.feedback-center__status-option--active {
  background: var(--app-surface-color);
  color: var(--u-type-primary);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(31, 35, 41, 0.08);
}

.feedback-center__list-head {
  min-width: 0;
  margin-bottom: 12px;
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.feedback-center__list-title {
  color: var(--app-text-color);
  font-size: 17px;
  font-weight: 700;
}

.feedback-center__list-total {
  color: var(--app-text-muted-color);
  font-size: 12px;
  white-space: nowrap;
}

@media screen and (max-width: 1023px) {
  .feedback-center__filters {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.5fr);
  }

  .feedback-center__filter-actions {
    grid-column: 1 / -1;
  }
}

@media screen and (max-width: 768px) {
  .feedback-center__create,
  :deep(.feedback-center__create) {
    width: 100%;
  }

  .feedback-center__filters {
    grid-template-columns: minmax(0, 1fr);
  }

  .feedback-center__filter-actions {
    grid-column: auto;
  }
}
</style>
