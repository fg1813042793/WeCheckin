<script setup lang="ts">
import { computed, useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

type DialogWidth = 'sm' | 'md' | 'lg' | string

const props = withDefaults(defineProps<{
  modelValue: boolean
  title: string
  width?: DialogWidth
  confirmText?: string
  cancelText?: string
  confirmLoading?: boolean
  confirmDisabled?: boolean
  showFooter?: boolean
  appendToBody?: boolean
  destroyOnClose?: boolean
  closeOnClickModal?: boolean
}>(), {
  width: 'md',
  confirmText: '确定',
  cancelText: '取消',
  confirmLoading: false,
  confirmDisabled: false,
  showFooter: true,
  appendToBody: true,
  destroyOnClose: true,
  closeOnClickModal: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: []
  cancel: []
  closed: []
}>()

const attrs = useAttrs()
const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})
const dialogWidth = computed(() => {
  const widths: Record<string, string> = {
    sm: 'min(var(--admin-ui-dialog-width-sm), calc(100vw - 32px))',
    md: 'min(var(--admin-ui-dialog-width-md), calc(100vw - 32px))',
    lg: 'min(var(--admin-ui-dialog-width-lg), calc(100vw - 32px))',
  }
  return widths[props.width] || props.width
})

function cancel() {
  visible.value = false
  emit('cancel')
}
</script>

<template>
  <el-dialog
    v-model="visible"
    v-bind="attrs"
    class="admin-ui-dialog"
    :title="title"
    :width="dialogWidth"
    :append-to-body="appendToBody"
    :destroy-on-close="destroyOnClose"
    :close-on-click-modal="closeOnClickModal"
    @closed="emit('closed')"
  >
    <div class="admin-ui-dialog__body">
      <slot />
    </div>
    <template v-if="showFooter || $slots.footer" #footer>
      <slot name="footer" :cancel="cancel" :confirm="() => emit('confirm')">
        <el-button @click="cancel">{{ cancelText }}</el-button>
        <el-button
          type="primary"
          :loading="confirmLoading"
          :disabled="confirmDisabled"
          @click="emit('confirm')"
        >{{ confirmText }}</el-button>
      </slot>
    </template>
  </el-dialog>
</template>

<style>
.admin-ui-dialog.el-dialog {
  overflow: hidden;
  border-radius: var(--admin-ui-radius-lg);
  box-shadow: var(--admin-ui-shadow-dialog);
}

.admin-ui-dialog .el-dialog__header {
  min-height: 24px;
  margin-right: 0;
  padding: var(--admin-ui-space-5) var(--admin-ui-space-6) var(--admin-ui-space-4);
  border-bottom: 1px solid var(--admin-ui-color-border);
}

.admin-ui-dialog .el-dialog__body {
  padding: 0;
}

.admin-ui-dialog__body {
  box-sizing: border-box;
  max-height: min(68vh, 720px);
  overflow: auto;
  padding: var(--admin-ui-space-5) var(--admin-ui-space-6);
}

.admin-ui-dialog .el-dialog__footer {
  padding: var(--admin-ui-space-3) var(--admin-ui-space-6);
  border-top: 1px solid var(--admin-ui-color-border);
}

.admin-ui-dialog .el-dialog__footer .el-button + .el-button {
  margin-left: var(--admin-ui-space-2);
}

@media (max-width: 767px) {
  .admin-ui-dialog .el-dialog__header,
  .admin-ui-dialog__body,
  .admin-ui-dialog .el-dialog__footer {
    padding-right: var(--admin-ui-space-4);
    padding-left: var(--admin-ui-space-4);
  }
}
</style>
