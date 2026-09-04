<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  activeCount?: number
}>()

const expanded = ref(!resolveMobilePresentation())

function resolveMobilePresentation() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    if (width > 0)
      return width <= 768
  }
  catch {
    // 读取失败时使用 H5 媒体查询判断。
  }

  // #ifdef H5
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function')
    return window.matchMedia('(max-width: 768px)').matches
  // #endif

  return false
}
</script>

<template>
  <view class="workflow-filter-panel" :class="{ 'is-expanded': expanded }">
    <u-button custom-class="workflow-filter-panel__toggle" @click="expanded = !expanded">
      <view class="workflow-filter-panel__toggle-content">
        <view class="workflow-filter-panel__heading">
          <u-icon name="search" size="14px" color="#4e5969" />
          <text class="workflow-filter-panel__title">
            筛选条件
          </text>
          <text v-if="activeCount" class="workflow-filter-panel__count">
            已选 {{ activeCount }}
          </text>
        </view>
        <u-icon v-if="expanded" name="arrow-up" size="14px" color="#86909c" />
        <u-icon v-else name="arrow-down" size="14px" color="#86909c" />
      </view>
    </u-button>
    <view v-show="expanded" class="workflow-filter-panel__content">
      <slot />
    </view>
  </view>
</template>

<style lang="scss" scoped>
.workflow-filter-panel {
  margin-bottom: 20px;
  border-bottom: 1px solid #e5eaf3;
}

.workflow-filter-panel__toggle,
:deep(.workflow-filter-panel__toggle) {
  width: 100%;
  height: 38px;
  min-height: 38px;
  margin: 0;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
}

.workflow-filter-panel__toggle::after,
:deep(.workflow-filter-panel__toggle)::after {
  display: none;
}

.workflow-filter-panel__toggle-content {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.workflow-filter-panel__heading {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 7px;
}

.workflow-filter-panel__title {
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
}

.workflow-filter-panel__count {
  color: #86909c;
  font-size: 11px;
  font-weight: 400;
}

.workflow-filter-panel__content {
  padding: 12px 0 18px;
}

@media screen and (max-width: 768px) {
  .workflow-filter-panel {
    margin-bottom: 14px;
  }

  .workflow-filter-panel__toggle,
  :deep(.workflow-filter-panel__toggle) {
    height: 36px;
    min-height: 36px;
  }

  .workflow-filter-panel__content {
    padding: 10px 0 14px;
  }
}
</style>
