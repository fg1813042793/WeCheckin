<script setup lang="ts">
import type { CalendarChangeDate, CalendarChangeRange } from 'uview-pro/types'
import { computed, ref } from 'vue'

const props = withDefaults(defineProps<{
  disabled?: boolean
  modelValue?: string
  placeholder?: string
}>(), {
  disabled: false,
  modelValue: '',
  placeholder: '选择日期',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const desktopOpen = ref(false)
const mobileOpen = ref(false)
const rootClass = computed(() => [
  'workflow-history-date-picker',
  desktopOpen.value ? 'is-open' : '',
  props.disabled ? 'is-disabled' : '',
].filter(Boolean).join(' '))

function resolveMobilePresentation() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    if (width > 0 && width <= 768)
      return true
  }
  catch {
    // 读取失败时按 PC 处理，避免桌面端出现底部日期弹层。
  }

  // #ifdef H5
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function')
    return window.matchMedia('(hover: none) and (pointer: coarse)').matches
  // #endif

  return false
}

function openCalendar() {
  if (props.disabled)
    return
  if (resolveMobilePresentation()) {
    desktopOpen.value = false
    mobileOpen.value = true
    return
  }
  desktopOpen.value = !desktopOpen.value
}

function closeDesktopCalendar() {
  desktopOpen.value = false
}

function selectDate(payload: CalendarChangeDate | CalendarChangeRange) {
  const value = 'result' in payload ? String(payload.result || '') : ''
  if (!value)
    return
  emit('update:modelValue', value)
  desktopOpen.value = false
  mobileOpen.value = false
}
</script>

<template>
  <view :class="rootClass">
    <view
      class="workflow-history-date-picker__field"
      role="button"
      :aria-disabled="disabled"
      @click="openCalendar"
    >
      <text :class="{ 'workflow-history-date-picker__placeholder': !modelValue }">
        {{ modelValue || placeholder }}
      </text>
      <u-icon name="calendar" size="16px" :color="disabled ? '#c8c9cc' : '#86909c'" />
    </view>

    <view v-if="desktopOpen" class="workflow-history-date-picker__mask" @click="closeDesktopCalendar" />
    <view v-if="desktopOpen" class="workflow-history-date-picker__panel" @click.stop>
      <u-calendar
        :model-value="true"
        :is-page="true"
        mode="date"
        :default-date="modelValue"
        @change="selectDate"
      />
    </view>

    <u-calendar
      v-model="mobileOpen"
      mode="date"
      :default-date="modelValue"
      @change="selectDate"
    />
  </view>
</template>

<style lang="scss" scoped>
.workflow-history-date-picker {
  position: relative;
  min-width: 0;
  width: 100%;
}

.workflow-history-date-picker.is-open {
  z-index: 40;
}

.workflow-history-date-picker__field {
  width: 100%;
  height: 36px;
  min-width: 0;
  padding: 0 10px;
  border: 1px solid #d8e0e8;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  overflow: hidden;
  background: #ffffff;
  color: #1f2329;
  font-size: 13px;
  letter-spacing: 0;
  box-sizing: border-box;
  cursor: pointer;
}

.workflow-history-date-picker__field > text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-history-date-picker__placeholder {
  color: #a9b0bb;
}

.workflow-history-date-picker.is-disabled .workflow-history-date-picker__field {
  background: #f2f3f5;
  color: #a9b0bb;
  cursor: not-allowed;
}

.workflow-history-date-picker__mask {
  position: fixed;
  inset: 0;
  z-index: 1;
  background: transparent;
}

.workflow-history-date-picker__panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 50%;
  z-index: 2;
  width: 360px;
  max-width: calc(100vw - 32px);
  padding: 8px;
  border: 1px solid #dfe5ee;
  border-radius: 6px;
  background: #ffffff;
  box-shadow: 0 12px 30px rgba(31, 35, 41, 0.16);
  box-sizing: border-box;
  transform: translateX(-50%);
}

@media screen and (max-width: 768px) {
  .workflow-history-date-picker__panel,
  .workflow-history-date-picker__mask {
    display: none;
  }
}
</style>
