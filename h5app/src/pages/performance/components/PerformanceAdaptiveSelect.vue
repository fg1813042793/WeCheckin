<script setup lang="ts">
import { computed, ref } from 'vue'

interface SelectOption {
  value: string
  label: string
  type?: 'group'
}

interface MonthGridMonth {
  value: string
  label: string
  available: boolean
  selected: boolean
}

const props = withDefaults(defineProps<{
  border?: boolean
  customClass?: string
  customStyle?: Record<string, string | number>
  desktopPlacement?: 'bottom' | 'top'
  desktopVariant?: 'list' | 'month-grid'
  disabled?: boolean
  modelValue?: string | number
  mobileOptions?: SelectOption[]
  options: SelectOption[]
  placeholder?: string
  title?: string
}>(), {
  border: true,
  customClass: '',
  customStyle: () => ({}),
  desktopPlacement: 'bottom',
  desktopVariant: 'list',
  disabled: false,
  modelValue: '',
  placeholder: '请选择',
  title: '请选择',
})

const emit = defineEmits<{
  'change': [value: string]
  'update:modelValue': [value: string]
}>()

const desktopOpen = ref(false)
const mobileSelectVisible = ref(false)
const monthGridYear = ref(new Date().getFullYear())

const selectedValue = computed(() => String(props.modelValue ?? ''))
const selectableOptions = computed(() => props.options.filter(option => option.type !== 'group'))
const mobileSelectableOptions = computed(() => (props.mobileOptions || props.options).filter(option => option.type !== 'group'))
const monthGridPeriodOptions = computed(() => selectableOptions.value.filter(option => Boolean(parseMonthGridValue(option.value))))
const monthGridValueSet = computed(() => new Set(monthGridPeriodOptions.value.map(option => option.value)))
const monthGridYears = computed(() => {
  const years = monthGridPeriodOptions.value
    .map(option => parseMonthGridValue(option.value)?.year)
    .filter((year): year is number => Boolean(year))
  const selectedYear = parseMonthGridValue(selectedValue.value)?.year
  if (selectedYear) {
    years.push(selectedYear)
  }
  return [...new Set(years)].sort((left, right) => right - left)
})
const monthGridMonths = computed<MonthGridMonth[]>(() => {
  return Array.from({ length: 12 }, (_, index) => {
    const month = index + 1
    const value = formatMonthGridValue(monthGridYear.value, month)
    return {
      value,
      label: `${String(month).padStart(2, '0')}月`,
      available: monthGridValueSet.value.has(value),
      selected: selectedValue.value === value,
    }
  })
})
const selectedLabel = computed(() => {
  return selectableOptions.value.find(item => item.value === selectedValue.value)?.label || props.placeholder
})
const canClearSelection = computed(() => !props.disabled && selectedValue.value !== '')
const mobileDefaultValue = computed(() => {
  const index = mobileSelectableOptions.value.findIndex(item => item.value === selectedValue.value)
  return [index >= 0 ? index : 0]
})
const rootClass = computed(() => [
  'performance-adaptive-select',
  desktopOpen.value ? 'is-open' : '',
  props.disabled ? 'is-disabled' : '',
  props.desktopVariant === 'month-grid' ? 'is-month-grid' : '',
].filter(Boolean).join(' '))
const fieldClass = computed(() => ['performance-adaptive-select__field', props.customClass].filter(Boolean).join(' '))
const desktopPanelClass = computed(() => [
  'performance-adaptive-select__panel',
  `is-placement-${props.desktopPlacement}`,
  props.desktopVariant === 'month-grid' ? 'performance-adaptive-select__month-panel' : '',
].filter(Boolean).join(' '))

function resolveMobileSelect() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    if (width > 0 && width <= 768) {
      return true
    }
  }
  catch {
    // 无法读取设备信息时按 PC 处理，避免 H5 桌面端出现底部弹层。
  }

  // #ifdef H5
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function') {
    return window.matchMedia('(hover: none) and (pointer: coarse)').matches
  }
  // #endif

  return false
}

function openSelect() {
  if (props.disabled) {
    return
  }
  if (resolveMobileSelect()) {
    desktopOpen.value = false
    mobileSelectVisible.value = true
    return
  }
  if (!desktopOpen.value && props.desktopVariant === 'month-grid') {
    syncMonthGridYear()
  }
  desktopOpen.value = !desktopOpen.value
}

function closeDesktopSelect() {
  desktopOpen.value = false
}

function setValue(value: string) {
  emit('update:modelValue', value)
  emit('change', value)
}

function selectDesktopOption(option: SelectOption) {
  if (option.type === 'group') {
    return
  }
  setValue(option.value)
  closeDesktopSelect()
}

function clearSelect() {
  setValue('')
  desktopOpen.value = false
  mobileSelectVisible.value = false
}

function confirmMobileSelect(items: unknown) {
  const selected = Array.isArray(items) ? items[0] : items
  const value = String((selected as { value?: unknown } | undefined)?.value ?? '')
  setValue(value)
}

function optionKey(option: SelectOption) {
  return option.type === 'group' ? `group-${option.label}` : `option-${option.value}`
}

function optionButtonClass(option: SelectOption) {
  return [
    'performance-adaptive-select__option',
    option.value === selectedValue.value ? 'is-selected' : '',
  ].filter(Boolean).join(' ')
}

function parseMonthGridValue(value: unknown) {
  const match = String(value || '').trim().match(/^(\d{4})-(\d{1,2})$/)
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

function formatMonthGridValue(year: number, month: number) {
  return `${year}-${String(month).padStart(2, '0')}`
}

function syncMonthGridYear() {
  const selectedParts = parseMonthGridValue(selectedValue.value)
  if (selectedParts) {
    monthGridYear.value = selectedParts.year
    return
  }
  monthGridYear.value = monthGridYears.value[0] || new Date().getFullYear()
}

function changeMonthGridYear(delta: number) {
  monthGridYear.value += Number(delta || 0)
}

function monthGridOptionClass(month: MonthGridMonth) {
  return [
    'performance-adaptive-select__month-option',
    month.selected ? 'is-selected' : '',
    month.available ? '' : 'is-disabled',
  ].filter(Boolean).join(' ')
}

function selectMonthGridValue(month: MonthGridMonth) {
  if (!month.available) {
    return
  }
  setValue(month.value)
  closeDesktopSelect()
}
</script>

<template>
  <view :class="rootClass">
    <u-input
      type="select"
      :custom-class="fieldClass"
      :custom-style="customStyle"
      :model-value="selectedLabel"
      :border="border"
      :clearable="canClearSelection"
      :disabled="disabled"
      :select-open="desktopOpen"
      :placeholder="placeholder"
      @click="openSelect"
      @clear="clearSelect"
    />

    <view v-if="desktopOpen" class="performance-adaptive-select__mask" @click="closeDesktopSelect" />
    <view v-if="desktopOpen" :class="desktopPanelClass" @click.stop>
      <template v-if="desktopVariant === 'month-grid'">
        <view class="performance-adaptive-select__month-head">
          <u-button custom-class="performance-adaptive-select__month-nav" @click="changeMonthGridYear(-1)">
            ‹
          </u-button>
          <text class="performance-adaptive-select__month-year">
            {{ monthGridYear }}年
          </text>
          <u-button custom-class="performance-adaptive-select__month-nav" @click="changeMonthGridYear(1)">
            ›
          </u-button>
        </view>
        <view class="performance-adaptive-select__month-grid">
          <u-button
            v-for="month in monthGridMonths"
            :key="month.value"
            :custom-class="monthGridOptionClass(month)"
            :disabled="!month.available"
            @click="selectMonthGridValue(month)"
          >
            {{ month.label }}
          </u-button>
        </view>
      </template>
      <view v-else class="performance-adaptive-select__option-list">
        <template v-for="option in options" :key="optionKey(option)">
          <view v-if="option.type === 'group'" class="performance-adaptive-select__group">
            {{ option.label }}
          </view>
          <u-button
            v-else
            :custom-class="optionButtonClass(option)"
            @click="selectDesktopOption(option)"
          >
            <text class="performance-adaptive-select__option-label">
              {{ option.label }}
            </text>
            <u-icon v-if="option.value === selectedValue" name="checkbox-mark" size="15" color="#1677ff" />
          </u-button>
        </template>
      </view>
    </view>

    <u-select
      v-model="mobileSelectVisible"
      mode="single-column"
      :title="title"
      :list="mobileSelectableOptions"
      :default-value="mobileDefaultValue"
      :preserve-selection="false"
      @confirm="confirmMobileSelect"
    />
  </view>
</template>

<style lang="scss" scoped>
.performance-adaptive-select {
  position: relative;
  min-width: 180px;
  flex: 0 0 auto;

  &.is-open {
    z-index: 320;
  }
}

.performance-adaptive-select__mask {
  position: fixed;
  inset: 0;
  z-index: 1;
  background: transparent;
}

.performance-adaptive-select__panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  z-index: 2;
  min-width: 100%;
  max-height: 260px;
  overflow-y: auto;
  border: 1px solid #dfe5ef;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 10px 28px rgba(31, 35, 41, 0.14);
}

.performance-adaptive-select__option-list {
  padding: 4px;
}

.performance-adaptive-select__group {
  margin: 6px 4px 2px;
  padding: 8px 6px 4px;
  border-top: 1px solid #edf0f5;
  color: #86909c;
  font-size: 12px;
  font-weight: 600;
  line-height: 18px;

  &:first-child {
    margin-top: 0;
    border-top: 0;
  }
}

.performance-adaptive-select__option,
:deep(.performance-adaptive-select__option) {
  width: 100%;
  min-height: 34px;
  height: auto;
  margin: 0;
  padding: 0 10px;
  border: 0;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  color: #1f2329;
  font-size: 13px;
  line-height: 34px;
  text-align: left;

  &::after {
    display: none;
  }
}

.performance-adaptive-select__option.is-selected,
:deep(.performance-adaptive-select__option.is-selected) {
  background: #eef5ff;
  color: #1677ff;
  font-weight: 600;
}

.performance-adaptive-select__option-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.performance-adaptive-select__month-panel {
  width: 300px;
  max-width: calc(100vw - 48px);
  max-height: none;
  padding: 14px;
  border-color: #e5eaf3;
  border-radius: 10px;
  overflow: visible;
  box-shadow: 0 16px 40px rgba(31, 35, 41, 0.14);
}

.performance-adaptive-select__month-panel.is-placement-bottom {
  top: calc(100% + 6px);
  right: auto;
  bottom: auto;
}

.performance-adaptive-select__month-panel.is-placement-top {
  top: auto;
  right: auto;
  bottom: calc(100% + 6px);
}

.performance-adaptive-select__month-head {
  margin-bottom: 12px;
  padding: 0 2px;
  display: grid;
  grid-template-columns: 32px minmax(0, 1fr) 32px;
  align-items: center;
  gap: 10px;
}

.performance-adaptive-select__month-year {
  color: #1f2329;
  font-size: 15px;
  font-weight: 800;
  line-height: 32px;
  text-align: center;
}

.performance-adaptive-select__month-nav,
:deep(.performance-adaptive-select__month-nav) {
  width: 32px;
  height: 32px;
  min-height: 32px;
  margin: 0;
  padding: 0;
  border: 1px solid #e5eaf3;
  border-radius: 8px;
  background: #f8fafc;
  color: #4e5969;
  font-size: 18px;
  font-weight: 800;
  line-height: 30px;
}

.performance-adaptive-select__month-nav::after,
:deep(.performance-adaptive-select__month-nav)::after {
  display: none;
}

.performance-adaptive-select__month-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.performance-adaptive-select__month-option,
:deep(.performance-adaptive-select__month-option) {
  width: 100%;
  height: 38px;
  min-height: 38px;
  margin: 0;
  padding: 0 8px;
  border: 1px solid #e5eaf3;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8fafc;
  color: #4e5969;
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
}

.performance-adaptive-select__month-option::after,
:deep(.performance-adaptive-select__month-option)::after {
  display: none;
}

.performance-adaptive-select__month-option.is-selected,
:deep(.performance-adaptive-select__month-option.is-selected) {
  border-color: #1677ff;
  background: #1677ff;
  color: #fff;
  box-shadow: 0 6px 14px rgba(22, 119, 255, 0.22);
}

.performance-adaptive-select__month-option.is-disabled,
:deep(.performance-adaptive-select__month-option.is-disabled) {
  border-color: #edf0f5;
  background: #f7f8fa;
  color: #c0c6d0;
  opacity: 1;
}

@media (max-width: 768px), (hover: none) and (pointer: coarse) {
  .performance-adaptive-select {
    min-width: 0;
    flex: 1 1 auto;
  }
}
</style>
