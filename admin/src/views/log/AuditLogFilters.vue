<script setup lang="ts">
import { Refresh } from '@element-plus/icons-vue'
import { AdminFilterActions } from '@/components/admin-ui'
import type { AuditLogFilters } from '@/types/auditLog'
import { actionOptions, moduleOptions, outcomeOptions, sourceOptions } from './auditLogHelpers'

defineProps<{ filters: AuditLogFilters; canExport: boolean; exportDisabled: boolean; exporting: boolean; loading: boolean }>()
defineEmits<{ search: []; reset: []; refresh: []; export: [] }>()
</script>

<template>
  <el-card class="audit-search-card" shadow="never">
    <el-form @submit.prevent="$emit('search')">
      <div class="audit-filter-row audit-filter-row--primary">
        <el-form-item>
          <el-input v-model="filters.search" aria-label="账号 / 关键词" placeholder="账号、内容、请求 ID" clearable maxlength="200" @keyup.enter="$emit('search')" />
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.outcome" aria-label="结果" clearable placeholder="全部结果">
            <el-option v-for="item in outcomeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.module" aria-label="模块" clearable filterable placeholder="全部模块">
            <el-option v-for="item in moduleOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.source" aria-label="来源" clearable placeholder="全部来源">
            <el-option v-for="item in sourceOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model="filters.action" aria-label="动作" clearable filterable allow-create default-first-option placeholder="全部动作 / 动作标识">
            <el-option v-for="item in actionOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
      </div>
      <div class="audit-filter-row audit-filter-row--secondary">
        <el-form-item>
          <el-input v-model="filters.actorId" aria-label="操作者 ID" inputmode="numeric" placeholder="ID，0 表示匿名" clearable maxlength="20" @keyup.enter="$emit('search')" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="filters.ip" aria-label="IP 地址" placeholder="IPv4 / IPv6" clearable maxlength="64" @keyup.enter="$emit('search')" />
        </el-form-item>
        <el-form-item class="audit-date-field">
          <el-date-picker v-model="filters.range" class="audit-date-range" aria-label="时间范围（本地时间）" type="datetimerange" format="YYYY-MM-DD HH:mm:ss" value-format="YYYY-MM-DDTHH:mm:ssZ" start-placeholder="开始时间" end-placeholder="结束时间" range-separator="至" :default-time="[new Date(2000, 0, 1, 0, 0, 0), new Date(2000, 0, 1, 23, 59, 59)]" />
        </el-form-item>
        <AdminFilterActions :loading="loading" @search="$emit('search')" @reset="$emit('reset')" />
      </div>
    </el-form>
    <div class="audit-log-actions">
      <div class="audit-log-actions__left">
        <el-button v-if="canExport" text :loading="exporting" :disabled="exportDisabled" @click="$emit('export')">导出</el-button>
      </div>
      <el-tooltip content="刷新日志" placement="top">
        <span class="audit-log-actions__refresh">
          <el-button text :icon="Refresh" :disabled="loading" aria-label="刷新日志" @click="$emit('refresh')" />
        </span>
      </el-tooltip>
    </div>
  </el-card>
</template>

<style scoped>
.audit-search-card { border-color: var(--admin-ui-color-border); border-radius: var(--admin-ui-radius); }
.audit-search-card :deep(.el-card__body) { padding: 12px 14px; }
.audit-filter-row { display: grid; gap: var(--admin-ui-space-3); align-items: end; min-width: 0; }
.audit-filter-row--primary { grid-template-columns: minmax(200px, 1.5fr) repeat(4, minmax(130px, 1fr)); }
.audit-filter-row--secondary { display: flex; flex-wrap: wrap; gap: var(--admin-ui-space-2); margin-top: var(--admin-ui-space-3); }
.audit-filter-row--secondary :deep(.el-form-item) { flex: 0 1 var(--audit-filter-width, 180px); max-width: 100%; }
.audit-filter-row--secondary :deep(.el-form-item:nth-child(2)) { --audit-filter-width: 220px; }
.audit-filter-row--secondary :deep(.audit-date-field) { --audit-filter-width: 400px; }
.audit-filter-row :deep(.el-form-item) { margin-bottom: 0; min-width: 0; }
.audit-filter-row :deep(.el-select) { width: 100%; }
.audit-date-field :deep(.el-form-item__content) { min-width: 0; }
.audit-date-field :deep(.audit-date-range) { width: 400px; max-width: 100%; box-sizing: border-box; }
.audit-date-field :deep(.el-range-input) { flex: 1; min-width: 0; font-size: 12px; }
.audit-date-field :deep(.el-range-separator) { flex: 0 0 auto; padding: 0 4px; }
.audit-log-actions { display: flex; align-items: center; justify-content: space-between; gap: var(--admin-ui-space-3); margin-top: var(--admin-ui-space-3); padding-top: var(--admin-ui-space-3); border-top: 1px solid var(--admin-ui-color-border); }
.audit-log-actions__left { display: flex; align-items: center; flex-wrap: wrap; gap: var(--admin-ui-space-2); min-width: 0; }
.audit-log-actions__refresh { display: inline-flex; flex: 0 0 auto; }
@media (max-width: 1199px) {
  .audit-filter-row--primary { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (max-width: 767px) {
  .audit-filter-row--primary { grid-template-columns: minmax(0, 1fr); }
  .audit-filter-row--secondary :deep(.el-form-item) { flex-basis: 100%; }
  .audit-date-field :deep(.audit-date-range) { height: auto; min-height: 72px; padding: 6px 24px; flex-direction: column; position: relative; }
  .audit-date-field :deep(.el-range-input) { flex: none; width: 100%; height: 22px; }
  .audit-date-field :deep(.el-range__icon) { position: absolute; left: 5px; top: 22px; }
  .audit-date-field :deep(.el-range__close-icon) { position: absolute; right: 5px; top: 22px; }
  .audit-date-field :deep(.el-range-separator) { line-height: 16px; height: 16px; }
}
</style>
