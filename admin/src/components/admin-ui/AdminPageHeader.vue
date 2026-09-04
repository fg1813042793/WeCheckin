<script setup lang="ts">
import { ArrowLeft } from '@element-plus/icons-vue'

withDefaults(defineProps<{
  title: string
  description?: string
  back?: boolean
  divided?: boolean
}>(), {
  description: '',
  back: false,
  divided: true,
})

const emit = defineEmits<{
  back: []
}>()
</script>

<template>
  <header class="admin-ui-page-header" :class="{ 'admin-ui-page-header--divided': divided }">
    <div class="admin-ui-page-header__identity">
      <el-button
        v-if="back"
        class="admin-ui-page-header__back"
        circle
        :icon="ArrowLeft"
        aria-label="返回"
        title="返回"
        @click="emit('back')"
      />
      <div class="admin-ui-page-header__copy">
        <div class="admin-ui-page-header__title-row">
          <h1>{{ title }}</h1>
          <slot name="status" />
        </div>
        <p v-if="description">{{ description }}</p>
        <div v-if="$slots.meta" class="admin-ui-page-header__meta">
          <slot name="meta" />
        </div>
      </div>
    </div>
    <div v-if="$slots.actions" class="admin-ui-page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.admin-ui-page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--admin-ui-space-4);
  min-width: 0;
  padding-bottom: var(--admin-ui-space-4);
}

.admin-ui-page-header--divided {
  border-bottom: 1px solid var(--admin-ui-color-border);
}

.admin-ui-page-header__identity,
.admin-ui-page-header__title-row,
.admin-ui-page-header__actions,
.admin-ui-page-header__meta {
  display: flex;
  align-items: center;
}

.admin-ui-page-header__identity {
  gap: var(--admin-ui-space-3);
  min-width: 0;
}

.admin-ui-page-header__copy {
  min-width: 0;
}

.admin-ui-page-header__title-row {
  flex-wrap: wrap;
  gap: var(--admin-ui-space-2);
}

.admin-ui-page-header h1 {
  overflow-wrap: anywhere;
  margin: 0;
  font-size: var(--admin-ui-font-size-title);
  line-height: var(--admin-ui-line-height-title);
  font-weight: 650;
  letter-spacing: 0;
}

.admin-ui-page-header p {
  margin: var(--admin-ui-space-1) 0 0;
  color: var(--admin-ui-color-text-secondary);
  font-size: var(--admin-ui-font-size-body);
  line-height: var(--admin-ui-line-height-body);
}

.admin-ui-page-header__meta {
  flex-wrap: wrap;
  gap: var(--admin-ui-space-3);
  margin-top: var(--admin-ui-space-2);
  color: var(--admin-ui-color-text-muted);
  font-size: var(--admin-ui-font-size-caption);
}

.admin-ui-page-header__actions {
  flex: 0 0 auto;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--admin-ui-space-2);
}

.admin-ui-page-header__actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

@media (max-width: 767px) {
  .admin-ui-page-header {
    align-items: stretch;
    flex-direction: column;
  }

  .admin-ui-page-header__actions {
    justify-content: flex-start;
  }
}
</style>
