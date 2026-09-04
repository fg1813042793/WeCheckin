<script setup lang="ts">
import { computed, useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

type DrawerSize = 'sm' | 'md' | 'lg' | string
type DrawerDirection = 'rtl' | 'ltr' | 'ttb' | 'btt'

const props = withDefaults(defineProps<{
  modelValue: boolean
  title: string
  size?: DrawerSize
  direction?: DrawerDirection
  confirmText?: string
  cancelText?: string
  confirmLoading?: boolean
  confirmDisabled?: boolean
  showFooter?: boolean
  appendToBody?: boolean
  destroyOnClose?: boolean
  closeOnClickModal?: boolean
}>(), {
  size: 'md',
  direction: 'rtl',
  confirmText: '确定',
  cancelText: '取消',
  confirmLoading: false,
  confirmDisabled: false,
  showFooter: false,
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
const drawerSize = computed(() => {
  const sizes: Record<string, string> = {
    sm: 'min(var(--admin-ui-drawer-width-sm), 100vw)',
    md: 'min(var(--admin-ui-drawer-width-md), 100vw)',
    lg: 'min(var(--admin-ui-drawer-width-lg), 100vw)',
  }
  return sizes[props.size] || props.size
})

function cancel() {
  visible.value = false
  emit('cancel')
}
</script>

<template>
  <el-drawer
    v-model="visible"
    v-bind="attrs"
    class="admin-ui-drawer"
    :title="title"
    :size="drawerSize"
    :direction="direction"
    :append-to-body="appendToBody"
    :destroy-on-close="destroyOnClose"
    :close-on-click-modal="closeOnClickModal"
    @closed="emit('closed')"
  >
    <div class="admin-ui-drawer__layout">
      <el-scrollbar class="admin-ui-drawer__scrollbar">
        <div class="admin-ui-drawer__body">
          <slot />
        </div>
      </el-scrollbar>
      <footer v-if="showFooter || $slots.footer" class="admin-ui-drawer__footer">
        <slot name="footer" :cancel="cancel" :confirm="() => emit('confirm')">
          <el-button @click="cancel">{{ cancelText }}</el-button>
          <el-button
            type="primary"
            :loading="confirmLoading"
            :disabled="confirmDisabled"
            @click="emit('confirm')"
          >{{ confirmText }}</el-button>
        </slot>
      </footer>
    </div>
  </el-drawer>
</template>

<style>
.admin-ui-drawer .el-drawer__header {
  min-height: 28px;
  margin-bottom: 0;
  padding: var(--admin-ui-space-5) var(--admin-ui-space-6);
  border-bottom: 1px solid var(--admin-ui-color-border);
}

.admin-ui-drawer .el-drawer__body {
  min-height: 0;
  padding: 0;
}

.admin-ui-drawer__layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.admin-ui-drawer__scrollbar {
  flex: 1 1 auto;
  min-height: 0;
}

.admin-ui-drawer__body {
  box-sizing: border-box;
  padding: var(--admin-ui-space-5) var(--admin-ui-space-6);
}

.admin-ui-drawer__footer {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: flex-end;
  gap: var(--admin-ui-space-2);
  padding: var(--admin-ui-space-3) var(--admin-ui-space-6);
  border-top: 1px solid var(--admin-ui-color-border);
  background: var(--admin-ui-color-surface);
}

.admin-ui-drawer__footer .el-button + .el-button {
  margin-left: 0;
}

@media (max-width: 767px) {
  .admin-ui-drawer .el-drawer__header,
  .admin-ui-drawer__body,
  .admin-ui-drawer__footer {
    padding-right: var(--admin-ui-space-4);
    padding-left: var(--admin-ui-space-4);
  }
}
</style>
