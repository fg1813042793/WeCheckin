<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'

const props = withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  disabled?: boolean
  maxlength?: number
  count?: boolean
  minRows?: number
  maxRows?: number
}>(), {
  modelValue: '',
  placeholder: '',
  disabled: false,
  maxlength: -1,
  count: false,
  minRows: 3,
  maxRows: 8,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const rowHeight = 24
const nativeVerticalPadding = 20
const nativeTextareaTag = 'textarea'
const nativeTextareaRef = ref<unknown>(null)
const composing = ref(false)
const nativeValue = ref(String(props.modelValue ?? ''))
const lastCommittedValue = ref(nativeValue.value)
const textModel = computed({
  get: () => String(props.modelValue ?? ''),
  set: value => emit('update:modelValue', String(value ?? '')),
})
const normalizedMinRows = computed(() => normalizeRows(props.minRows, 3))
const normalizedMaxRows = computed(() => Math.max(normalizedMinRows.value, normalizeRows(props.maxRows, 8)))
const minHeight = computed(() => normalizedMinRows.value * rowHeight)
const nativeMinHeight = computed(() => minHeight.value + nativeVerticalPadding)
const nativeMaxHeight = computed(() => normalizedMaxRows.value * rowHeight + nativeVerticalPadding)
const nativeMaxlength = computed(() => props.maxlength >= 0 ? props.maxlength : undefined)
const countText = computed(() => props.maxlength >= 0
  ? `${nativeValue.value.length}/${props.maxlength}`
  : String(nativeValue.value.length))
const textareaStyle = computed(() => ({
  '--workflow-textarea-max-height': `${normalizedMaxRows.value * rowHeight}px`,
  '--workflow-native-textarea-min-height': `${nativeMinHeight.value}px`,
  '--workflow-native-textarea-max-height': `${nativeMaxHeight.value}px`,
}))

watch(
  () => props.modelValue,
  (value) => {
    const nextValue = String(value ?? '')
    lastCommittedValue.value = nextValue
    if (composing.value || nativeValue.value === nextValue)
      return
    nativeValue.value = nextValue
    void nextTick(resizeNativeTextarea)
  },
)

watch([normalizedMinRows, normalizedMaxRows], () => void nextTick(resizeNativeTextarea))

onMounted(() => void nextTick(resizeNativeTextarea))

function normalizeRows(value: number, fallback: number) {
  const rows = Number(value)
  return Number.isInteger(rows) && rows > 0 ? Math.min(30, rows) : fallback
}

function resolveNativeTextarea(value: unknown): HTMLTextAreaElement | null {
  // #ifdef H5
  if (typeof HTMLTextAreaElement === 'undefined')
    return null
  if (value instanceof HTMLTextAreaElement)
    return value
  if (value && typeof value === 'object' && '$el' in value) {
    const element = (value as { $el?: unknown }).$el
    if (element instanceof HTMLTextAreaElement)
      return element
  }
  // #endif
  return null
}

function nativeEventTextarea(event: unknown) {
  if (!event || typeof event !== 'object')
    return null
  const currentTarget = (event as { currentTarget?: unknown }).currentTarget
  const target = (event as { target?: unknown }).target
  return resolveNativeTextarea(currentTarget) || resolveNativeTextarea(target)
}

function nativeInputValue(event: unknown) {
  const textarea = nativeEventTextarea(event)
  if (textarea)
    return textarea.value
  if (event && typeof event === 'object' && 'detail' in event) {
    const detail = (event as { detail?: { value?: unknown } }).detail
    if (detail && detail.value !== undefined)
      return String(detail.value ?? '')
  }
  return nativeValue.value
}

function nativeInputComposing(event: unknown) {
  if (!event || typeof event !== 'object')
    return false
  if ('isComposing' in event)
    return Boolean((event as { isComposing?: boolean }).isComposing)
  const detail = (event as { detail?: { isComposing?: boolean } }).detail
  return Boolean(detail?.isComposing)
}

function commitNativeValue(value: string) {
  if (value === lastCommittedValue.value)
    return
  lastCommittedValue.value = value
  emit('update:modelValue', value)
}

function resizeNativeTextarea(value: unknown = nativeTextareaRef.value) {
  const textarea = resolveNativeTextarea(value)
  if (!textarea)
    return
  textarea.style.height = 'auto'
  const scrollHeight = textarea.scrollHeight
  const height = Math.max(nativeMinHeight.value, Math.min(nativeMaxHeight.value, scrollHeight))
  textarea.style.height = `${height}px`
  textarea.style.overflowY = scrollHeight > nativeMaxHeight.value ? 'auto' : 'hidden'
}

function handleCompositionStart() {
  composing.value = true
}

function handleCompositionEnd(event: unknown) {
  composing.value = false
  const value = nativeInputValue(event)
  nativeValue.value = value
  resizeNativeTextarea(nativeEventTextarea(event))
  commitNativeValue(value)
}

function handleNativeInput(event: unknown) {
  const value = nativeInputValue(event)
  nativeValue.value = value
  resizeNativeTextarea(nativeEventTextarea(event))
  if (composing.value || nativeInputComposing(event))
    return
  commitNativeValue(value)
}

function handleNativeBlur(event: unknown) {
  if (!composing.value)
    return
  handleCompositionEnd(event)
}
</script>

<template>
  <!-- #ifdef H5 -->
  <view
    class="workflow-textarea__native"
    :class="{
      'workflow-textarea__native--count': count,
      'workflow-textarea__native--disabled': disabled,
    }"
    :style="textareaStyle"
  >
    <component
      :is="nativeTextareaTag"
      ref="nativeTextareaRef"
      :value="nativeValue"
      class="workflow-textarea__native-control"
      :disabled="disabled"
      :maxlength="nativeMaxlength"
      :placeholder="placeholder"
      @blur="handleNativeBlur"
      @compositionstart="handleCompositionStart"
      @compositionend="handleCompositionEnd"
      @input="handleNativeInput"
    />
    <text v-if="count" class="workflow-textarea__native-count">
      {{ countText }}
    </text>
  </view>
  <!-- #endif -->
  <!-- #ifndef H5 -->
  <u-textarea
    v-model="textModel"
    custom-class="workflow-textarea__control"
    :custom-style="textareaStyle"
    :border="true"
    :disabled="disabled"
    :maxlength="maxlength"
    :placeholder="placeholder"
    :height="minHeight"
    :auto-height="true"
    :count="count"
  />
  <!-- #endif -->
</template>

<style lang="scss" scoped>
.workflow-textarea__native {
  position: relative;
  width: 100%;
  border: 1px solid var(--u-border-color);
  border-radius: 4px;
  background: var(--u-bg-white);
  box-sizing: border-box;
  overflow: hidden;
}

.workflow-textarea__native-control {
  display: block;
  width: 100%;
  min-height: var(--workflow-native-textarea-min-height);
  max-height: var(--workflow-native-textarea-max-height);
  margin: 0;
  padding: 10px 12px;
  border: 0;
  outline: 0;
  color: var(--u-main-color);
  background: transparent;
  font: inherit;
  font-size: 28rpx;
  line-height: 24px;
  resize: none;
  box-sizing: border-box;
  overflow-y: hidden;
}

.workflow-textarea__native--count .workflow-textarea__native-control {
  padding-bottom: 28px;
}

.workflow-textarea__native-control::placeholder {
  color: var(--u-light-color);
}

.workflow-textarea__native-count {
  position: absolute;
  right: 10px;
  bottom: 6px;
  color: var(--u-light-color);
  background: var(--u-bg-white);
  font-size: 24rpx;
  line-height: 18px;
  pointer-events: none;
}

.workflow-textarea__native--disabled {
  background: var(--u-bg-color);
}

.workflow-textarea__native--disabled .workflow-textarea__native-control {
  color: var(--u-main-color);
  -webkit-text-fill-color: var(--u-main-color);
  cursor: not-allowed;
}

.workflow-textarea__native--disabled .workflow-textarea__native-control::placeholder {
  color: var(--u-light-color);
  -webkit-text-fill-color: var(--u-light-color);
}

.workflow-textarea__native--disabled .workflow-textarea__native-count {
  background: transparent;
}

:deep(.workflow-textarea__control .u-textarea__field) {
  max-height: var(--workflow-textarea-max-height) !important;
  overflow-y: auto !important;
}

:deep(.workflow-textarea__control .u-textarea__field .uni-textarea-textarea) {
  overflow-y: auto !important;
}

:deep(.workflow-textarea__control.u-textarea--disabled .u-textarea__field) {
  color: var(--u-main-color) !important;
  -webkit-text-fill-color: var(--u-main-color) !important;
}

:deep(.workflow-textarea__control.u-textarea--disabled .u-textarea__field::placeholder) {
  color: var(--u-light-color) !important;
  -webkit-text-fill-color: var(--u-light-color) !important;
}

@media screen and (min-width: 769px) and (max-width: 1023px) {
  .workflow-textarea__native-control {
    font-size: 13px;
    line-height: 20px;
  }

  .workflow-textarea__native-count {
    font-size: 11px;
    line-height: 16px;
  }
}
</style>
