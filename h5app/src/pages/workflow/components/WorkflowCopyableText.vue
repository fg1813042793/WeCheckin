<script setup lang="ts">
import { clipboard } from 'uview-pro'
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  value?: string
}>(), {
  value: '',
})

const displayValue = computed(() => String(props.value || '').trim() || '-')
const canCopy = computed(() => displayValue.value !== '-')

function showCopySuccess() {
  uni.showToast({ title: '流程单号已复制', icon: 'success' })
}

function copyWithUView(value: string) {
  clipboard(value, {
    showToast: false,
    success: showCopySuccess,
    fail: () => uni.showToast({ title: '复制失败，请重试', icon: 'none' }),
  })
}

async function copyValue() {
  if (!canCopy.value)
    return

  const value = displayValue.value
  // #ifdef H5
  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value)
      showCopySuccess()
      return
    }
    catch {
      // Continue with the uView fallback for restricted browser contexts.
    }
  }
  // #endif

  copyWithUView(value)
}
</script>

<template>
  <view class="workflow-copyable-text">
    <text class="workflow-copyable-text__value" :title="displayValue">
      {{ displayValue }}
    </text>
    <u-button
      v-if="canCopy"
      custom-class="workflow-copyable-text__button app-icon-button app-icon-button--small"
      title="复制流程单号"
      aria-label="复制流程单号"
      @click.stop="copyValue"
    >
      <view class="workflow-copyable-text__icon" aria-hidden="true" />
    </u-button>
  </view>
</template>

<style lang="scss" scoped>
.workflow-copyable-text {
  max-width: 100%;
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  vertical-align: middle;
}

.workflow-copyable-text__value {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-copyable-text__button,
:deep(.workflow-copyable-text__button) {
  flex: 0 0 28px;
  width: 28px;
  height: 28px;
  min-height: 28px;
  margin: 0;
  padding: 0;
  border: 0;
  color: #667085;
  background: transparent;
}

.workflow-copyable-text__button::after,
:deep(.workflow-copyable-text__button)::after {
  display: none;
}

.workflow-copyable-text__button:hover,
:deep(.workflow-copyable-text__button):hover {
  color: #2563eb;
  background: #eff6ff;
}

.workflow-copyable-text__icon {
  position: relative;
  width: 14px;
  height: 14px;
  color: currentColor;
}

.workflow-copyable-text__icon::before,
.workflow-copyable-text__icon::after {
  position: absolute;
  width: 9px;
  height: 9px;
  border: 1px solid currentColor;
  border-radius: 2px;
  box-sizing: border-box;
  content: '';
}

.workflow-copyable-text__icon::before {
  top: 1px;
  right: 1px;
}

.workflow-copyable-text__icon::after {
  bottom: 1px;
  left: 1px;
}
</style>
