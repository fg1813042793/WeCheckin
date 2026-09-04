<script setup lang="ts">
import { computed } from 'vue'

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
const textModel = computed({
  get: () => String(props.modelValue ?? ''),
  set: value => emit('update:modelValue', String(value ?? '')),
})
const normalizedMinRows = computed(() => normalizeRows(props.minRows, 3))
const normalizedMaxRows = computed(() => Math.max(normalizedMinRows.value, normalizeRows(props.maxRows, 8)))
const minHeight = computed(() => normalizedMinRows.value * rowHeight)
const textareaStyle = computed(() => ({
  '--workflow-textarea-max-height': `${normalizedMaxRows.value * rowHeight}px`,
}))

function normalizeRows(value: number, fallback: number) {
  const rows = Number(value)
  return Number.isInteger(rows) && rows > 0 ? Math.min(30, rows) : fallback
}
</script>

<template>
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
</template>

<style lang="scss" scoped>
:deep(.workflow-textarea__control .u-textarea__field) {
  max-height: var(--workflow-textarea-max-height) !important;
  overflow-y: auto !important;
}

:deep(.workflow-textarea__control .u-textarea__field .uni-textarea-textarea) {
  overflow-y: auto !important;
}
</style>
