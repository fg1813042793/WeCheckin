<script setup lang="ts">
import type { UserFeedbackOverview, UserFeedbackStatus } from '@/types/userFeedback'
import { userFeedbackStatusMeta } from '../userFeedbackStatus'

const props = withDefaults(defineProps<{
  overview?: UserFeedbackOverview | null
  activeStatus?: UserFeedbackStatus | ''
  loading?: boolean
  error?: string
}>(), {
  overview: null,
  activeStatus: '',
  loading: false,
  error: '',
})

const emit = defineEmits<{
  select: [status: UserFeedbackStatus]
  retry: []
}>()

const overviewItems: Array<{ status: UserFeedbackStatus; countKey: keyof UserFeedbackOverview }> = [
  { status: 'pending', countKey: 'pending' },
  { status: 'processing', countKey: 'processing' },
  { status: 'resolved', countKey: 'resolved' },
  { status: 'closed', countKey: 'closed' },
]
</script>

<template>
  <section class="feedback-overview" aria-label="反馈状态统计">
    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false">
      <template #default>
        <el-button link type="primary" @click="emit('retry')">重新加载</el-button>
      </template>
    </el-alert>
    <el-skeleton v-else-if="loading && !overview" :rows="1" animated />
    <div v-else class="feedback-overview__items">
      <button
        v-for="item in overviewItems"
        :key="item.status"
        type="button"
        class="feedback-overview__item"
        :class="[
          `feedback-overview__item--${userFeedbackStatusMeta(item.status).type}`,
          { 'is-active': activeStatus === item.status },
        ]"
        :aria-pressed="activeStatus === item.status"
        @click="emit('select', item.status)"
      >
        <span class="feedback-overview__count">{{ overview?.[item.countKey] ?? 0 }}</span>
        <span>{{ userFeedbackStatusMeta(item.status).label }}</span>
      </button>
    </div>
  </section>
</template>

<style scoped>
.feedback-overview {
  min-width: 0;
}

.feedback-overview__items {
  display: grid;
  overflow: hidden;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border: 1px solid var(--admin-ui-color-border);
  border-radius: var(--admin-ui-radius);
  background: var(--admin-ui-color-surface);
}

.feedback-overview__item {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--admin-ui-space-2);
  min-width: 0;
  min-height: 58px;
  padding: var(--admin-ui-space-3) var(--admin-ui-space-4);
  border: 0;
  border-right: 1px solid var(--admin-ui-color-border);
  background: transparent;
  color: var(--admin-ui-color-text-secondary);
  cursor: pointer;
  font: inherit;
  letter-spacing: 0;
  text-align: left;
}

.feedback-overview__item:last-child {
  border-right: 0;
}

.feedback-overview__item::before {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 3px;
  background: var(--feedback-overview-accent);
  content: '';
}

.feedback-overview__item:hover,
.feedback-overview__item.is-active {
  background: var(--admin-ui-color-surface-muted);
}

.feedback-overview__item:focus-visible {
  z-index: 1;
  outline: 2px solid var(--admin-ui-color-focus);
  outline-offset: -2px;
}

.feedback-overview__item--warning {
  --feedback-overview-accent: var(--admin-ui-color-warning);
}

.feedback-overview__item--success {
  --feedback-overview-accent: var(--admin-ui-color-success);
}

.feedback-overview__item--info {
  --feedback-overview-accent: var(--admin-ui-color-text-muted);
}

.feedback-overview__count {
  flex: 0 0 auto;
  color: var(--admin-ui-color-text);
  font-size: 22px;
  font-weight: 650;
  line-height: 28px;
}

@media (max-width: 767px) {
  .feedback-overview__items {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .feedback-overview__item:nth-child(2) {
    border-right: 0;
  }

  .feedback-overview__item:nth-child(-n + 2) {
    border-bottom: 1px solid var(--admin-ui-color-border);
  }
}
</style>
