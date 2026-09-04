<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  title?: string
  count?: number
  loading?: boolean
  empty?: boolean
  emptyText?: string
  error?: string
  minHeight?: number | string
}>(), {
  title: '',
  count: undefined,
  loading: false,
  empty: false,
  emptyText: '暂无数据',
  error: '',
  minHeight: 240,
})

const bodyStyle = computed(() => ({
  '--admin-ui-table-min-height': typeof props.minHeight === 'number' ? `${props.minHeight}px` : props.minHeight,
}))
</script>

<template>
  <section class="admin-ui-table-panel">
    <header v-if="title || $slots.toolbar || $slots.actions" class="admin-ui-table-panel__toolbar">
      <div class="admin-ui-table-panel__heading">
        <slot name="toolbar">
          <h2 v-if="title">{{ title }}</h2>
          <span v-if="typeof count === 'number'" class="admin-ui-table-panel__count">共 {{ count }} 条</span>
        </slot>
      </div>
      <div v-if="$slots.actions" class="admin-ui-table-panel__actions">
        <slot name="actions" />
      </div>
    </header>
    <div v-loading="loading" class="admin-ui-table-panel__body" :style="bodyStyle">
      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
      <el-empty v-else-if="empty && !loading" :description="emptyText" :image-size="88" />
      <slot v-else />
    </div>
    <footer v-if="$slots.footer" class="admin-ui-table-panel__footer">
      <slot name="footer" />
    </footer>
  </section>
</template>

<style scoped>
.admin-ui-table-panel {
  overflow: hidden;
  border: 1px solid var(--admin-ui-color-border);
  border-radius: var(--admin-ui-radius);
  background: var(--admin-ui-color-surface);
}

.admin-ui-table-panel__toolbar,
.admin-ui-table-panel__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--admin-ui-space-3);
  min-width: 0;
  padding: var(--admin-ui-space-3) var(--admin-ui-space-4);
}

.admin-ui-table-panel__toolbar {
  min-height: var(--admin-ui-table-header-height);
  border-bottom: 1px solid var(--admin-ui-color-border);
}

.admin-ui-table-panel__footer {
  border-top: 1px solid var(--admin-ui-color-border);
}

.admin-ui-table-panel__heading,
.admin-ui-table-panel__actions {
  display: flex;
  align-items: center;
  gap: var(--admin-ui-space-2);
  min-width: 0;
}

.admin-ui-table-panel__actions {
  flex: 0 0 auto;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.admin-ui-table-panel__actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.admin-ui-table-panel h2 {
  overflow-wrap: anywhere;
  margin: 0;
  font-size: var(--admin-ui-font-size-section);
  line-height: 24px;
  letter-spacing: 0;
}

.admin-ui-table-panel__count {
  color: var(--admin-ui-color-text-muted);
  font-size: var(--admin-ui-font-size-caption);
}

.admin-ui-table-panel__body {
  min-height: var(--admin-ui-table-min-height);
  min-width: 0;
}

.admin-ui-table-panel__body :deep(.el-table__header-wrapper th.el-table__cell) {
  height: var(--admin-ui-table-header-height);
  background: var(--admin-ui-color-surface-muted);
}

.admin-ui-table-panel__body :deep(.el-table__body-wrapper td.el-table__cell) {
  min-height: var(--admin-ui-table-row-height);
}

@media (max-width: 767px) {
  .admin-ui-table-panel__toolbar,
  .admin-ui-table-panel__footer {
    align-items: stretch;
    flex-direction: column;
  }

  .admin-ui-table-panel__actions {
    justify-content: flex-start;
  }
}
</style>
