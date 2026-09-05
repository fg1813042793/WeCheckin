<script setup lang="ts">
import type { UserFeedbackOverview, UserFeedbackStatus } from '@/api/user-feedback'
import { useLocale } from 'uview-pro'
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  activeStatus?: UserFeedbackStatus | ''
  error?: boolean
  loading?: boolean
  overview: UserFeedbackOverview
}>(), {
  activeStatus: '',
  error: false,
  loading: false,
})

const emit = defineEmits<{
  retry: []
  select: [status: UserFeedbackStatus]
}>()

const { t } = useLocale('feedback')

const statusTypes = {
  pending: 'warning',
  processing: 'warning',
  resolved: 'success',
  closed: 'info',
} as const

const items = computed(() => (Object.keys(statusTypes) as UserFeedbackStatus[]).map(status => ({
  status,
  count: Math.max(0, Number(props.overview[status] || 0)),
  label: t(`statuses.${status}`),
  type: statusTypes[status],
})))

function selectItem(item: { status: UserFeedbackStatus }) {
  emit('select', item.status)
}
</script>

<template>
  <view class="feedback-overview" :aria-busy="loading">
    <view class="feedback-overview__head">
      <text class="feedback-overview__title">
        {{ t('overview') }}
      </text>
      <view v-if="loading" class="feedback-overview__loading">
        <u-loading mode="circle" size="16px" />
        <text>{{ t('loading') }}</text>
      </view>
      <view
        v-else-if="error"
        class="feedback-overview__retry"
        role="button"
        tabindex="0"
        @click="emit('retry')"
        @keyup.enter="emit('retry')"
      >
        <u-icon name="reload" size="14px" color="var(--u-type-primary)" />
        <text>{{ t('overviewFailed') }}</text>
        <text class="feedback-overview__retry-action">
          {{ t('retry') }}
        </text>
      </view>
    </view>

    <view class="feedback-overview__items">
      <view
        v-for="item in items"
        :key="item.status"
        class="feedback-overview__item"
        :class="[
          `feedback-overview__item--${item.type}`,
          { 'feedback-overview__item--active': activeStatus === item.status },
        ]"
        role="button"
        tabindex="0"
        @click="selectItem(item)"
        @keyup.enter="selectItem(item)"
      >
        <view class="feedback-overview__value-row">
          <text class="feedback-overview__value">
            {{ item.count }}
          </text>
          <u-icon name="arrow-right" size="14px" color="var(--app-text-muted-color)" />
        </view>
        <text class="feedback-overview__label">
          {{ item.label }}
        </text>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.feedback-overview {
  min-width: 0;
  padding-bottom: 18px;
  border-bottom: 1px solid var(--app-border-color);
}

.feedback-overview__head {
  min-height: 24px;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.feedback-overview__title {
  color: var(--app-text-secondary-color);
  font-size: 13px;
  font-weight: 600;
}

.feedback-overview__loading,
.feedback-overview__retry {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: var(--app-text-muted-color);
  font-size: 12px;
}

.feedback-overview__retry {
  color: var(--app-text-muted-color);
  cursor: pointer;
}

.feedback-overview__retry-action {
  color: var(--u-type-primary);
}

.feedback-overview__items {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.feedback-overview__item {
  --feedback-status-color: var(--app-text-muted-color);

  min-width: 0;
  min-height: 66px;
  padding: 10px 12px;
  border: 1px solid var(--app-border-color);
  border-left: 3px solid var(--feedback-status-color);
  border-radius: var(--app-panel-radius);
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 3px;
  background: var(--app-surface-color);
  box-sizing: border-box;
  cursor: pointer;
  transition: border-color 160ms ease, background-color 160ms ease;
}

.feedback-overview__item--warning {
  --feedback-status-color: var(--u-type-warning);
}

.feedback-overview__item--success {
  --feedback-status-color: var(--u-type-success);
}

.feedback-overview__item--info {
  --feedback-status-color: var(--u-type-info);
}

.feedback-overview__item:hover,
.feedback-overview__item--active {
  border-color: var(--feedback-status-color);
  background: var(--app-surface-subtle-color);
}

.feedback-overview__value-row {
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.feedback-overview__value {
  color: var(--app-text-color);
  font-size: 22px;
  font-weight: 700;
  line-height: 1.2;
}

.feedback-overview__label {
  overflow: hidden;
  color: var(--app-text-secondary-color);
  font-size: 12px;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media screen and (max-width: 768px) {
  .feedback-overview {
    padding-bottom: 14px;
  }

  .feedback-overview__items {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .feedback-overview__item {
    min-height: 62px;
  }
}
</style>
