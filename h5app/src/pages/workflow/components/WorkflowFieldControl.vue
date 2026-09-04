<script setup lang="ts">
import type { WorkflowFormField, WorkflowFormOption } from '@/types/workflow'
import { computed, nextTick, ref, watch } from 'vue'
import { requestWorkflowOptionSource } from '@/api/workflow'
import {
  flattenWorkflowOptions,
  normalizeWorkflowOptions,
  validWorkflowOptionUrl,
  workflowOptionResponsePayload,
} from '../workflow-form'
import { resolveWorkflowSelectPlacement } from '../workflow-select-placement'
import WorkflowAttachmentControl from './WorkflowAttachmentControl.vue'
import WorkflowTextarea from './WorkflowTextarea.vue'

const props = withDefaults(defineProps<{
  field: WorkflowFormField
  modelValue: unknown
  readonly?: boolean
  readonlyAppearance?: 'disabled' | 'plain'
  textareaDefaultMinRows?: number
  textareaDefaultMaxRows?: number
}>(), {
  readonly: false,
  readonlyAppearance: 'disabled',
  textareaDefaultMinRows: 3,
  textareaDefaultMaxRows: 8,
})

const emit = defineEmits<{
  'update:modelValue': [value: unknown]
}>()

const remoteOptions = ref<WorkflowFormOption[]>([])
const optionLoading = ref(false)
const desktopSelectOpen = ref(false)
const desktopSelectPlacement = ref<'top' | 'bottom'>('bottom')
const desktopSelectRootRef = ref<unknown>(null)
const desktopSelectPanelRef = ref<unknown>(null)
const mobileSelectVisible = ref(false)

const plainReadonlyEmpty = computed(() => {
  if (!props.readonly || props.readonlyAppearance !== 'plain')
    return false
  const value = props.modelValue
  if (value === undefined || value === null)
    return true
  if (typeof value === 'string')
    return value.trim() === ''
  if (Array.isArray(value)) {
    return value.length === 0 || value.every((item) => {
      return item === undefined || item === null || (typeof item === 'string' && item.trim() === '')
    })
  }
  return false
})

const options = computed(() => {
  const source = props.field.optionSource
  const list = source?.type === 'api' && remoteOptions.value.length > 0
    ? remoteOptions.value
    : normalizeWorkflowOptions(props.field.options || [], source)
  return flattenWorkflowOptions(list)
})

const stringModel = computed({
  get: () => typeof props.modelValue === 'string' ? props.modelValue : '',
  set: value => emit('update:modelValue', value),
})

const numericModel = computed({
  get: () => props.modelValue === undefined || props.modelValue === null ? '' : String(props.modelValue),
  set: (value: string | number) => {
    const text = String(value || '').trim()
    emit('update:modelValue', text === '' ? undefined : Number(text))
  },
})

const booleanModel = computed({
  get: () => Boolean(props.modelValue),
  set: value => emit('update:modelValue', Boolean(value)),
})

const arrayModel = computed<string[]>({
  get: () => Array.isArray(props.modelValue) ? props.modelValue.map(item => String(item)) : [],
  set: value => emit('update:modelValue', value),
})

const arrayTextModel = computed({
  get: () => arrayModel.value.join('\n'),
  set: (value: string) => emit('update:modelValue', value.split(/\r?\n/).map(item => item.trim()).filter(Boolean)),
})

const selectedLabel = computed(() => {
  const value = String(props.modelValue || '')
  return options.value.find(option => option.value === value)?.label || value
})
const mobileDefaultValue = computed(() => {
  const selectedIndex = options.value.findIndex(option => option.value === stringModel.value)
  return [selectedIndex >= 0 ? selectedIndex : 0]
})

const dateValue = computed(() => stringModel.value.slice(0, 10))
const timeValue = computed(() => {
  const value = stringModel.value
  if (props.field.type === 'datetime')
    return value.slice(11, 16)
  return value.slice(0, 5)
})

watch(
  () => JSON.stringify(props.field.optionSource || {}),
  () => void loadRemoteOptions(),
  { immediate: true },
)

async function loadRemoteOptions() {
  const source = props.field.optionSource
  if (source?.type !== 'api' || !validWorkflowOptionUrl(source.url)) {
    remoteOptions.value = []
    return
  }
  optionLoading.value = true
  try {
    const response = await requestWorkflowOptionSource(source)
    remoteOptions.value = normalizeWorkflowOptions(workflowOptionResponsePayload(response, source), source)
  }
  catch {
    remoteOptions.value = []
  }
  finally {
    optionLoading.value = false
  }
}

function resolveMobileSelect() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    if (width > 0 && width <= 768)
      return true
  }
  catch {
    // 无法读取设备信息时按 PC 处理，避免宽屏出现底部弹层。
  }

  // #ifdef H5
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function')
    return window.matchMedia('(hover: none) and (pointer: coarse)').matches
  // #endif

  return false
}

function resolveH5Element(value: unknown): HTMLElement | null {
  // #ifdef H5
  if (typeof HTMLElement === 'undefined')
    return null
  if (value instanceof HTMLElement)
    return value
  if (value && typeof value === 'object' && '$el' in value) {
    const element = (value as { $el?: unknown }).$el
    if (element instanceof HTMLElement)
      return element
  }
  // #endif
  return null
}

// #ifdef H5
function resolveVisibleVerticalBounds(element: HTMLElement) {
  let visibleTop = 0
  let visibleBottom = Math.max(window.innerHeight, document.documentElement.clientHeight)
  let ancestor = element.parentElement

  while (ancestor) {
    const style = window.getComputedStyle(ancestor)
    if (['auto', 'scroll', 'hidden', 'clip'].includes(style.overflowY)) {
      const rect = ancestor.getBoundingClientRect()
      visibleTop = Math.max(visibleTop, rect.top)
      visibleBottom = Math.min(visibleBottom, rect.bottom)
    }
    ancestor = ancestor.parentElement
  }

  return { visibleTop, visibleBottom }
}
// #endif

function estimatedDesktopSelectPanelHeight() {
  return Math.min(260, Math.max(44, options.value.length * 36 + 8))
}

function syncDesktopSelectPlacement() {
  // #ifdef H5
  if (typeof window === 'undefined' || typeof document === 'undefined')
    return
  const control = resolveH5Element(desktopSelectRootRef.value)
  if (!control)
    return
  const panel = resolveH5Element(desktopSelectPanelRef.value)
  const controlRect = control.getBoundingClientRect()
  const { visibleTop, visibleBottom } = resolveVisibleVerticalBounds(control)
  const measuredPanelHeight = panel?.scrollHeight || 0
  const panelHeight = measuredPanelHeight > 0
    ? Math.min(260, measuredPanelHeight)
    : estimatedDesktopSelectPanelHeight()

  desktopSelectPlacement.value = resolveWorkflowSelectPlacement({
    controlTop: controlRect.top,
    controlBottom: controlRect.bottom,
    visibleTop,
    visibleBottom,
    panelHeight,
  })
  // #endif
}

function openSelect() {
  if (props.readonly || options.value.length === 0)
    return
  if (resolveMobileSelect()) {
    desktopSelectOpen.value = false
    mobileSelectVisible.value = true
    return
  }
  if (desktopSelectOpen.value) {
    closeDesktopSelect()
    return
  }
  syncDesktopSelectPlacement()
  desktopSelectOpen.value = true
  void nextTick(syncDesktopSelectPlacement)
}

function closeDesktopSelect() {
  desktopSelectOpen.value = false
}

function selectDesktopOption(option: WorkflowFormOption) {
  emit('update:modelValue', option.value)
  closeDesktopSelect()
}

function clearSelect() {
  if (props.readonly)
    return
  emit('update:modelValue', '')
  desktopSelectOpen.value = false
  mobileSelectVisible.value = false
}

function confirmSelect(items: Array<{ value?: string }>) {
  const selected = items[0]
  if (selected?.value !== undefined) {
    emit('update:modelValue', String(selected.value))
    mobileSelectVisible.value = false
  }
}

function updateDate(event: { detail?: { value?: string } }) {
  const value = String(event.detail?.value || '')
  if (props.field.type === 'datetime') {
    emit('update:modelValue', `${value} ${timeValue.value || '00:00'}:00`)
  }
  else {
    emit('update:modelValue', value)
  }
}

function updateTime(event: { detail?: { value?: string } }) {
  const value = String(event.detail?.value || '')
  if (props.field.type === 'datetime') {
    emit('update:modelValue', `${dateValue.value || new Date().toISOString().slice(0, 10)} ${value}:00`)
  }
  else {
    emit('update:modelValue', value ? `${value}:00` : '')
  }
}

function updateRange(index: number, event: { detail?: { value?: string } }) {
  const next = [...arrayModel.value]
  while (next.length < 2) next.push('')
  next[index] = String(event.detail?.value || '')
  emit('update:modelValue', next)
}
</script>

<template>
  <view class="workflow-control">
    <view v-if="plainReadonlyEmpty" class="workflow-control__empty">
      暂无填写
    </view>
    <WorkflowAttachmentControl
      v-else-if="field.type === 'attachment'"
      :field="field"
      :model-value="modelValue"
      :readonly="readonly"
      @update:model-value="emit('update:modelValue', $event)"
    />
    <u-input
      v-else-if="['text', 'phone', 'email', 'user', 'department'].includes(field.type)"
      v-model="stringModel"
      :border="true"
      :clearable="!readonly"
      :disabled="readonly"
      :maxlength="field.maxLength || undefined"
      :placeholder="field.placeholder || '请输入'"
    />
    <WorkflowTextarea
      v-else-if="field.type === 'textarea'"
      v-model="stringModel"
      :disabled="readonly"
      :maxlength="field.maxLength || undefined"
      :min-rows="field.minVisibleRows || textareaDefaultMinRows"
      :max-rows="field.maxVisibleRows || textareaDefaultMaxRows"
      :placeholder="field.placeholder || '请输入内容'"
      count
    />
    <u-input
      v-else-if="field.type === 'number' || field.type === 'amount'"
      v-model="numericModel"
      :border="true"
      :disabled="readonly"
      type="number"
      :placeholder="field.placeholder || (field.type === 'amount' ? '请输入金额' : '请输入数字')"
    />
    <view
      v-else-if="field.type === 'select'"
      ref="desktopSelectRootRef"
      class="workflow-control__select"
      :class="{ 'is-open': desktopSelectOpen }"
    >
      <u-input
        :model-value="selectedLabel"
        :border="true"
        :clearable="!readonly && Boolean(stringModel)"
        :disabled="readonly"
        :select-open="desktopSelectOpen"
        type="select"
        :placeholder="optionLoading ? '选项加载中' : field.placeholder || '请选择'"
        @click="openSelect"
        @clear="clearSelect"
      />
      <view v-if="desktopSelectOpen" class="workflow-control__select-mask" @click="closeDesktopSelect" />
      <view
        v-if="desktopSelectOpen"
        ref="desktopSelectPanelRef"
        class="workflow-control__select-panel"
        :class="{ 'workflow-control__select-panel--top': desktopSelectPlacement === 'top' }"
        @click.stop
      >
        <view
          v-for="option in options"
          :key="option.value"
          class="workflow-control__select-option"
          :class="{ 'is-selected': option.value === stringModel }"
          role="option"
          :aria-selected="option.value === stringModel"
          tabindex="0"
          @click="selectDesktopOption(option)"
          @keydown.enter.prevent="selectDesktopOption(option)"
          @keydown.space.prevent="selectDesktopOption(option)"
        >
          <text class="workflow-control__select-option-label">
            {{ option.label }}
          </text>
          <u-icon v-if="option.value === stringModel" name="checkbox-mark" size="15" color="#1677ff" />
        </view>
      </view>
      <u-select
        v-model="mobileSelectVisible"
        mode="single-column"
        :default-value="mobileDefaultValue"
        :list="options"
        :preserve-selection="false"
        @confirm="confirmSelect"
      />
    </view>
    <u-radio-group
      v-else-if="field.type === 'radio'"
      v-model="stringModel"
      :disabled="readonly"
      wrap
    >
      <u-radio
        v-for="option in options"
        :key="option.value"
        :label="option.label"
        :value="option.value"
      />
    </u-radio-group>
    <u-checkbox-group
      v-else-if="field.type === 'multi_select' || field.type === 'checkbox'"
      v-model="arrayModel"
      :disabled="readonly"
      wrap
    >
      <u-checkbox
        v-for="option in options"
        :key="option.value"
        :label="option.label"
        :value="option.value"
      />
    </u-checkbox-group>
    <view v-else-if="field.type === 'boolean'" class="workflow-control__switch">
      <u-switch v-model="booleanModel" :disabled="readonly" />
      <text>{{ booleanModel ? '是' : '否' }}</text>
    </view>
    <picker
      v-else-if="field.type === 'date'"
      :disabled="readonly"
      mode="date"
      :value="dateValue"
      @change="updateDate"
    >
      <view class="workflow-picker" :class="{ 'workflow-picker--disabled': readonly }">
        <text :class="{ 'workflow-picker__placeholder': !dateValue }">
          {{ dateValue || field.placeholder || '请选择日期' }}
        </text>
        <u-icon name="calendar" size="28" color="#909399" />
      </view>
    </picker>
    <view v-else-if="field.type === 'datetime'" class="workflow-control__datetime">
      <picker :disabled="readonly" mode="date" :value="dateValue" @change="updateDate">
        <view class="workflow-picker" :class="{ 'workflow-picker--disabled': readonly }">
          <text :class="{ 'workflow-picker__placeholder': !dateValue }">
            {{ dateValue || '选择日期' }}
          </text>
        </view>
      </picker>
      <picker :disabled="readonly" mode="time" :value="timeValue" @change="updateTime">
        <view class="workflow-picker" :class="{ 'workflow-picker--disabled': readonly }">
          <text :class="{ 'workflow-picker__placeholder': !timeValue }">
            {{ timeValue || '选择时间' }}
          </text>
        </view>
      </picker>
    </view>
    <picker
      v-else-if="field.type === 'time'"
      :disabled="readonly"
      mode="time"
      :value="timeValue"
      @change="updateTime"
    >
      <view class="workflow-picker" :class="{ 'workflow-picker--disabled': readonly }">
        <text :class="{ 'workflow-picker__placeholder': !timeValue }">
          {{ timeValue || field.placeholder || '请选择时间' }}
        </text>
        <u-icon name="clock" size="28" color="#909399" />
      </view>
    </picker>
    <view v-else-if="field.type === 'date_range'" class="workflow-control__date-range">
      <picker :disabled="readonly" mode="date" :value="arrayModel[0] || ''" @change="updateRange(0, $event)">
        <view class="workflow-picker" :class="{ 'workflow-picker--disabled': readonly }">
          <text :class="{ 'workflow-picker__placeholder': !arrayModel[0] }">
            {{ arrayModel[0] || '开始日期' }}
          </text>
        </view>
      </picker>
      <text class="workflow-control__range-separator">
        至
      </text>
      <picker :disabled="readonly" mode="date" :value="arrayModel[1] || ''" @change="updateRange(1, $event)">
        <view class="workflow-picker" :class="{ 'workflow-picker--disabled': readonly }">
          <text :class="{ 'workflow-picker__placeholder': !arrayModel[1] }">
            {{ arrayModel[1] || '结束日期' }}
          </text>
        </view>
      </picker>
    </view>
    <WorkflowTextarea
      v-else-if="['user_multi', 'department_multi'].includes(field.type)"
      v-model="arrayTextModel"
      :disabled="readonly"
      :placeholder="field.placeholder || '每行填写一项'"
    />
    <u-input
      v-else
      v-model="stringModel"
      :border="true"
      :disabled="readonly"
      :placeholder="field.placeholder || '请输入'"
    />
  </view>
</template>

<style lang="scss" scoped>
.workflow-control {
  width: 100%;
}

.workflow-control__empty {
  min-height: 72rpx;
  padding: 0 22rpx;
  border: 1px solid $u-border-color;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  background: #fff;
  color: $u-content-color;
  font-size: 26rpx;
  box-sizing: border-box;
}

.workflow-control__select {
  position: relative;
  width: 100%;
}

.workflow-control__select.is-open {
  z-index: 320;
}

.workflow-control__select-mask {
  position: fixed;
  inset: 0;
  z-index: 1;
  background: transparent;
}

.workflow-control__select-panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  z-index: 2;
  max-height: 260px;
  padding: 4px;
  overflow-y: auto;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 8px 20px rgba(31, 35, 41, 0.12);
  box-sizing: border-box;
}

.workflow-control__select-panel--top {
  top: auto;
  bottom: calc(100% + 6px);
  box-shadow: 0 -8px 20px rgba(31, 35, 41, 0.12);
}

.workflow-control__select-option {
  width: 100%;
  min-height: 36px;
  padding: 0 10px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  color: #1f2329;
  font-size: 13px;
  line-height: 20px;
  cursor: pointer;
  box-sizing: border-box;
}

.workflow-control__select-option:hover,
.workflow-control__select-option:focus-visible,
.workflow-control__select-option.is-selected {
  outline: none;
  background: #f2f6ff;
  color: #1677ff;
}

.workflow-control__select-option-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-control__switch {
  min-height: 72rpx;
  display: flex;
  align-items: center;
  gap: 18rpx;
  color: $u-content-color;
  font-size: 25rpx;
}

.workflow-control__datetime,
.workflow-control__date-range {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 12rpx;
}

.workflow-control__date-range {
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
}

.workflow-control__range-separator {
  color: $u-tips-color;
  font-size: 24rpx;
}

.workflow-picker {
  min-height: 72rpx;
  padding: 0 22rpx;
  border: 1px solid $u-border-color;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12rpx;
  background: #fff;
  color: $u-main-color;
  font-size: 26rpx;
  box-sizing: border-box;
}

.workflow-picker--disabled {
  background: #f5f7fa;
  color: $u-tips-color;
}

.workflow-picker__placeholder {
  color: $u-light-color;
}
</style>
