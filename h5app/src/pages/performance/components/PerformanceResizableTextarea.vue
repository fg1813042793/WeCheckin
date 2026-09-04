<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  modelValue?: string | number
  placeholder?: string
  disabled?: boolean
  height?: number | string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '',
  disabled: false,
  height: 84,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const textValue = computed(() => String(props.modelValue ?? ''))
const textareaStyle = computed(() => ({
  '--template-textarea-height': typeof props.height === 'number' ? `${props.height}px` : props.height,
}))

function readInputValue(event: unknown) {
  const inputEvent = event as {
    detail?: { value?: unknown }
    target?: { value?: unknown }
  }
  return String(inputEvent.detail?.value ?? inputEvent.target?.value ?? '')
}

function handleInput(event: unknown) {
  emit('update:modelValue', readInputValue(event))
}
</script>

<template>
  <textarea
    class="template-editor-textarea template-native-textarea"
    :style="textareaStyle"
    :value="textValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :maxlength="-1"
    :auto-height="false"
    @input="handleInput"
  />
</template>
