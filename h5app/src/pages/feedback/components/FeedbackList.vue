<script setup lang="ts">
import type { UserFeedbackStatus, UserFeedbackSummary } from '@/api/user-feedback'
import { useLocale } from 'uview-pro'
import { computed } from 'vue'
import { feedbackStatusMeta } from '../feedback-status'

interface PaginationChangePayload {
  current: number
}

const props = withDefaults(defineProps<{
  error?: boolean
  items: UserFeedbackSummary[]
  loading?: boolean
  page: number
  pageSize: number
  total: number
}>(), {
  error: false,
  loading: false,
})

const emit = defineEmits<{
  open: [item: UserFeedbackSummary]
  pageChange: [page: number]
  retry: []
}>()

const { currentLocale, t } = useLocale('feedback')
const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

function statusLabel(status: UserFeedbackStatus) {
  return t(`statuses.${status}`)
}

function imageCountText(count: number) {
  return t('imageCount', { count: Math.max(0, Number(count || 0)) })
}

function formatTime(value: number) {
  const date = new Date(value)
  if (!value || Number.isNaN(date.getTime()))
    return '-'
  const locale = currentLocale.value?.name === 'en-US' ? 'en-US' : 'zh-CN'
  return date.toLocaleString(locale, { hour12: false })
}

function changePage(payload: PaginationChangePayload) {
  const page = Math.min(totalPages.value, Math.max(1, Number(payload.current || 1)))
  emit('pageChange', page)
}

function openItem(item: UserFeedbackSummary) {
  emit('open', item)
}
</script>

<template>
  <view class="feedback-list-section" :aria-busy="loading">
    <view v-if="loading" class="feedback-list__state">
      <u-loading mode="circle" size="24px" />
      <text>{{ t('loading') }}</text>
    </view>

    <view v-else-if="error" class="feedback-list__state">
      <u-icon name="error-circle" size="30px" color="var(--app-text-muted-color)" />
      <text class="feedback-list__state-title">
        {{ t('loadFailed') }}
      </text>
      <u-button custom-class="feedback-list__retry" size="small" plain @click="emit('retry')">
        <view class="feedback-list__button-content">
          <u-icon name="reload" size="14px" color="var(--u-type-primary)" />
          <text>{{ t('retry') }}</text>
        </view>
      </u-button>
    </view>

    <view v-else-if="items.length === 0" class="feedback-list__state">
      <u-icon name="chat" size="32px" color="#c9cdd4" />
      <text class="feedback-list__state-title">
        {{ t('empty') }}
      </text>
    </view>

    <template v-else>
      <view class="app-data-table feedback-list">
        <view class="app-data-table__scroller">
          <view class="app-data-table__header">
            <text>{{ t('feedbackNo') }}</text>
            <text>{{ t('summary') }}</text>
            <text>{{ t('status') }}</text>
            <text>{{ t('images') }}</text>
            <text>{{ t('lastActivity') }}</text>
            <text class="feedback-list__action-label">
              {{ t('view') }}
            </text>
          </view>

          <view
            v-for="item in items"
            :key="item.id"
            class="app-data-table__row feedback-list__row"
            role="button"
            tabindex="0"
            @click="openItem(item)"
            @keyup.enter="openItem(item)"
          >
            <text class="app-data-table__cell feedback-list__number" :data-label="t('feedbackNo')" :title="item.feedbackNo">
              {{ item.feedbackNo }}
            </text>
            <text class="app-data-table__cell feedback-list__summary" :data-label="t('summary')" :title="item.summary">
              {{ item.summary || '-' }}
            </text>
            <view class="app-data-table__cell feedback-list__status" :data-label="t('status')">
              <u-tag
                :text="statusLabel(item.status)"
                :type="feedbackStatusMeta(item.status).type"
                mode="light"
                size="mini"
              />
            </view>
            <text class="app-data-table__cell" :data-label="t('images')">
              {{ imageCountText(item.imageCount) }}
            </text>
            <text class="app-data-table__cell feedback-list__time" :data-label="t('lastActivity')">
              {{ formatTime(item.lastActivityAt) }}
            </text>
            <view class="app-data-table__cell app-data-table__actions" :data-label="t('view')">
              <u-button custom-class="feedback-list__open" size="mini" plain @click.stop="openItem(item)">
                <view class="feedback-list__button-content">
                  <u-icon name="eye" size="14px" color="var(--u-type-primary)" />
                  <text>{{ t('view') }}</text>
                </view>
              </u-button>
            </view>
          </view>
        </view>
      </view>

      <view v-if="total > pageSize" class="feedback-list__pagination">
        <u-pagination
          :model-value="page"
          custom-class="feedback-list__pagination-control"
          :total="total"
          :page-size="pageSize"
          :prev-text="t('previousPage')"
          :next-text="t('nextPage')"
          @change="changePage"
        />
      </view>
    </template>
  </view>
</template>

<style lang="scss" scoped>
.feedback-list-section {
  min-width: 0;
}

.feedback-list {
  --app-table-columns: minmax(160px, 0.9fr) minmax(220px, 1.5fr) 100px 90px minmax(170px, 1fr) 84px;
  --app-table-min-width: 880px;
}

.feedback-list__row {
  cursor: pointer;
}

.feedback-list__number {
  color: var(--app-text-color);
  font-weight: 600;
}

.feedback-list__summary,
.feedback-list__time {
  min-width: 0;
}

.feedback-list__status :deep(.u-tag) {
  width: auto;
}

.feedback-list__action-label {
  text-align: right;
}

.feedback-list__state {
  min-height: 260px;
  padding: 32px 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--app-text-muted-color);
  font-size: 13px;
  text-align: center;
  box-sizing: border-box;
}

.feedback-list__state-title {
  color: var(--app-text-color);
  font-size: 15px;
  font-weight: 600;
}

.feedback-list__retry,
.feedback-list__open,
:deep(.feedback-list__retry),
:deep(.feedback-list__open) {
  width: auto;
  min-width: 72px;
  height: 30px;
  min-height: 30px;
  margin: 0;
  padding: 0 10px;
  border-radius: 4px;
}

.feedback-list__button-content {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  font-size: 12px;
}

.feedback-list__pagination {
  padding: 18px 0 0;
  display: flex;
  justify-content: flex-end;
}

:deep(.feedback-list__pagination-control) {
  width: auto;
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.feedback-list__pagination-control .u-pagination-text),
:deep(.feedback-list__pagination-control .u-pagination-text text),
:deep(.feedback-list__pagination-control .u-pagination-text .u-text__value) {
  color: var(--app-text-secondary-color) !important;
  font-size: 12px !important;
  line-height: 28px;
  white-space: nowrap;
}

:deep(.feedback-list__pagination-control .custom-class) {
  width: auto !important;
  min-width: 52px !important;
  height: 28px !important;
  min-height: 28px !important;
  margin: 0 !important;
  padding: 0 8px !important;
  border-radius: 4px !important;
  font-size: 12px !important;
  line-height: 28px !important;
}

@media screen and (max-width: 768px) {
  .feedback-list__row {
    background: var(--app-surface-color);
  }

  .feedback-list__summary {
    display: -webkit-box;
    overflow: hidden;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
  }

  .feedback-list__pagination {
    justify-content: center;
  }
}
</style>
