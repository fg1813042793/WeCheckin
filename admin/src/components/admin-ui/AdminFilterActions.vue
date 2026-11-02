<script setup lang="ts">
const props = withDefaults(defineProps<{
  loading?: boolean
  disabled?: boolean
  nativeType?: 'button' | 'submit'
  searchText?: string
  resetText?: string
}>(), {
  loading: false,
  disabled: false,
  nativeType: 'button',
  searchText: '查询',
  resetText: '重置',
})

const emit = defineEmits<{ search: []; reset: [] }>()
function search() {
  if (!props.disabled && !props.loading && props.nativeType === 'button') emit('search')
}
function reset() {
  if (!props.disabled) emit('reset')
}
</script>

<template>
  <div class="admin-filter-actions" role="group" aria-label="筛选操作">
    <el-button type="primary" size="default" :native-type="nativeType" :loading="loading" :disabled="disabled" @click="search">{{ searchText }}</el-button>
    <el-button size="default" native-type="button" :disabled="disabled" @click="reset">{{ resetText }}</el-button>
  </div>
</template>

<style scoped>
.admin-filter-actions {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: flex-start;
  justify-self: start;
  gap: var(--admin-ui-space-2);
  white-space: nowrap;
}
.admin-filter-actions :deep(.el-button) {
  height: var(--admin-ui-control-height);
  min-width: 60px;
  margin: 0;
  padding: 0 var(--admin-ui-space-3);
  border-radius: var(--admin-ui-radius);
  font-size: var(--admin-ui-font-size-body);
}
</style>
