<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, watch } from 'vue'
import { useTemplateWorkflowOptions } from '../useTemplateWorkflowOptions'

const props = defineProps<{ modelValue?: number; displayName?: string; disabled?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: number | undefined]; ready: [value: boolean] }>()
const { options, selected, loading, resolving, error, selectedError, page, total, keyword, pageSize, search, resolveSelected, dispose } = useTemplateWorkflowOptions()
const displayedOptions = computed(() => {
  if (!props.modelValue || options.value.some(item => item.id === props.modelValue)) return options.value
  return [{ id: props.modelValue, name: selected.value?.name || props.displayName || `流程 #${props.modelValue}` }, ...options.value]
})
watch(() => props.modelValue, id => void resolveSelected(id), { immediate: true })
watch(() => Boolean(props.modelValue && selected.value?.id === props.modelValue && !resolving.value && !selectedError.value), ready => emit('ready', ready), { immediate: true })
onMounted(() => void search())
onBeforeUnmount(dispose)
</script>

<template>
  <div class="workflow-select">
    <el-select :model-value="modelValue" filterable remote clearable :remote-method="search" :loading="loading || resolving"
      :disabled="disabled" placeholder="搜索已发布流程" no-data-text="暂无匹配的已发布流程"
      @update:model-value="emit('update:modelValue', $event || undefined)">
      <el-option v-for="item in displayedOptions" :key="item.id" :value="item.id" :label="item.name" />
    </el-select>
    <div v-if="error || selectedError" class="workflow-error" role="alert">
      <span>{{ selectedError || error }}</span>
      <el-button link type="primary" :disabled="disabled || loading || resolving" @click="search(keyword, page); resolveSelected(modelValue)">重试</el-button>
    </div>
    <el-pagination v-if="total > pageSize" small layout="prev, pager, next" :pager-count="5"
      :current-page="page" :page-size="pageSize" :total="total" :disabled="disabled || loading"
      @current-change="search(keyword, $event)" />
    <span v-if="!loading && !error && !options.length" class="workflow-help">暂无匹配的已发布流程，可更换关键词。</span>
  </div>
</template>

<style scoped>
.workflow-select { display: grid; width: 100%; gap: var(--admin-ui-space-2); }
.workflow-select :deep(.el-pagination) { justify-self: end; }
.workflow-error { color: var(--admin-ui-color-danger); overflow-wrap: anywhere; }
.workflow-help { color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); }
</style>
