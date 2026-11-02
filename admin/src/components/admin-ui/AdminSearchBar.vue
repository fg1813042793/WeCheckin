<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowDown, ArrowUp } from '@element-plus/icons-vue'
import AdminFilterActions from './AdminFilterActions.vue'

const props = withDefaults(defineProps<{
  title?: string
  collapsible?: boolean
  expanded?: boolean
  defaultExpanded?: boolean
  loading?: boolean
  showActions?: boolean
  searchText?: string
  resetText?: string
}>(), {
  title: '筛选条件',
  collapsible: false,
  expanded: undefined,
  defaultExpanded: true,
  loading: false,
  showActions: true,
  searchText: '查询',
  resetText: '重置',
})

const emit = defineEmits<{
  'update:expanded': [value: boolean]
  search: []
  reset: []
}>()

const localExpanded = ref(props.defaultExpanded)

watch(() => props.expanded, (value) => {
  if (typeof value === 'boolean') localExpanded.value = value
}, { immediate: true })

const isExpanded = computed(() => !props.collapsible || localExpanded.value)

function toggle() {
  localExpanded.value = !localExpanded.value
  emit('update:expanded', localExpanded.value)
}
</script>

<template>
  <section class="admin-ui-search-bar" aria-label="筛选条件">
    <header v-if="title || collapsible" class="admin-ui-search-bar__header">
      <div class="admin-ui-search-bar__title">
        <slot name="title">{{ title }}</slot>
      </div>
      <el-button
        v-if="collapsible"
        text
        :icon="isExpanded ? ArrowUp : ArrowDown"
        :aria-expanded="isExpanded"
        @click="toggle"
      >{{ isExpanded ? '收起' : '展开' }}</el-button>
    </header>
    <el-collapse-transition>
      <form v-show="isExpanded" class="admin-ui-search-bar__content" @submit.prevent="emit('search')">
        <div class="admin-ui-search-bar__fields">
          <slot />
          <div v-if="showActions || $slots.actions" class="admin-ui-search-bar__actions">
            <slot name="actions">
              <AdminFilterActions native-type="submit" :loading="loading" :search-text="searchText" :reset-text="resetText" @reset="emit('reset')" />
            </slot>
          </div>
        </div>
      </form>
    </el-collapse-transition>
  </section>
</template>

<style scoped>
.admin-ui-search-bar {
  border: 1px solid var(--admin-ui-color-border);
  border-radius: var(--admin-ui-radius);
  background: var(--admin-ui-color-surface);
}

.admin-ui-search-bar__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--admin-ui-space-3);
  min-height: 40px;
  padding: 0 var(--admin-ui-space-4);
  border-bottom: 1px solid var(--admin-ui-color-border);
}

.admin-ui-search-bar__title {
  color: var(--admin-ui-color-text-secondary);
  font-size: var(--admin-ui-font-size-body);
  font-weight: 600;
}

.admin-ui-search-bar__content {
  display: block;
  padding: 12px 14px;
}

.admin-ui-search-bar__fields {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--admin-ui-space-2);
  min-width: 0;
}

.admin-ui-search-bar__fields :deep(.el-form-item) {
  min-width: 0;
  max-width: 100%;
  margin-bottom: 0;
}

.admin-ui-search-bar__fields :deep(.el-input),
.admin-ui-search-bar__fields :deep(.el-select) {
  width: 240px;
  max-width: 100%;
}

.admin-ui-search-bar__fields :deep(.el-date-editor) {
  width: 320px;
  max-width: 100%;
}

.admin-ui-search-bar__actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: var(--admin-ui-space-2);
}

.admin-ui-search-bar__actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

@media (max-width: 767px) {
  .admin-ui-search-bar__content {
    padding: var(--admin-ui-space-3);
  }

  .admin-ui-search-bar__fields {
    width: 100%;
  }

  .admin-ui-search-bar__fields :deep(.el-form-item),
  .admin-ui-search-bar__fields :deep(.el-input),
  .admin-ui-search-bar__fields :deep(.el-select),
  .admin-ui-search-bar__fields :deep(.el-date-editor) {
    width: 100%;
  }
}
</style>
