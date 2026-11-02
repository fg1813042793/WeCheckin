<script setup lang="ts">
import { AdminPageShell, AdminTablePanel } from '@/components/admin-ui'
import AuditLogFilters from './AuditLogFilters.vue'
import AuditLogDetail from './AuditLogDetail.vue'
import { actionLabel, actorLabel, logTime, moduleLabel, outcomeMeta, sourceLabel } from './auditLogHelpers'
import { useAuditLogs } from './useAuditLogs'

const { category, canReadOperations, canReadLogin, canRead, canExport, filters, rows, total, page, pageSize,
  loading, dirty, error, filterError, exporting, exportNotice, detailOpen, detail, detailLoading, detailError,
  load, search, reset, changePage, changeSize, showDetail, retryDetail, exportData } = useAuditLogs()
</script>

<template>
  <AdminPageShell width="fluid" class="audit-log-page">
    <el-tabs v-model="category" class="audit-log-tabs" aria-label="日志类别">
      <el-tab-pane v-if="canReadOperations" label="操作日志" name="operation" />
      <el-tab-pane v-if="canReadLogin" label="登录日志" name="login" />
    </el-tabs>
    <el-empty v-if="!canRead" description="暂无查看审计日志的权限" />
    <template v-else>
      <AuditLogFilters :filters="filters" :can-export="canExport" :export-disabled="loading || dirty || !!error || !!filterError || total === 0" :exporting="exporting" :loading="loading" @search="search" @reset="reset" @refresh="load" @export="exportData" />
      <el-alert v-if="filterError" :title="filterError" type="warning" :closable="false" show-icon />
      <el-alert v-if="exportNotice" :title="exportNotice" type="info" :closable="false" show-icon />
      <AdminTablePanel :title="category === 'login' ? '登录记录' : '操作记录'" :count="total" :loading="loading" :error="error" :empty="rows.length === 0" :empty-text="dirty ? '筛选条件已修改，请点击查询' : '暂无符合条件的日志'">
        <template #actions>
          <span v-if="canExport" class="audit-export-limit">每次导出最多 1000 条</span>
          <el-button v-if="error" type="primary" plain @click="load">重试</el-button>
        </template>
        <el-table :data="rows" row-key="id" stripe style="width: 100%">
          <el-table-column label="时间" width="180"><template #default="{ row }">{{ logTime(row._createTime) }}</template></el-table-column>
          <el-table-column label="结果" width="160"><template #default="{ row }"><el-tag :type="outcomeMeta(row.outcome).type" size="small">{{ outcomeMeta(row.outcome).label }}</el-tag></template></el-table-column>
          <el-table-column label="来源" width="110"><template #default="{ row }">{{ sourceLabel(row.source) }}</template></el-table-column>
          <el-table-column label="模块 / 动作" min-width="160" show-overflow-tooltip><template #default="{ row }">{{ moduleLabel(row.module) }} / {{ actionLabel(row) }}</template></el-table-column>
          <el-table-column label="实际操作者" min-width="210" show-overflow-tooltip><template #default="{ row }">{{ actorLabel(row) }}</template></el-table-column>
          <el-table-column v-if="category === 'login'" prop="attemptedAccount" label="尝试账号（待核实）" min-width="180" show-overflow-tooltip />
          <el-table-column prop="content" label="内容" min-width="240" show-overflow-tooltip />
          <el-table-column prop="LOG_ADD_IP" label="IP 地址" min-width="150" show-overflow-tooltip />
          <el-table-column label="详情" width="80" fixed="right"><template #default="{ row }"><el-button link type="primary" :aria-label="`查看日志 ${row.id} 详情`" @click="showDetail(row)">详情</el-button></template></el-table-column>
        </el-table>
      </AdminTablePanel>
      <div class="audit-pagination">
        <el-pagination :current-page="page" :page-size="pageSize" :page-sizes="[10, 20, 50, 100]" :total="total" :disabled="loading || dirty || !!filterError" layout="total, sizes, prev, pager, next" :pager-count="7" @current-change="changePage" @size-change="changeSize" />
      </div>
      <p class="audit-note">历史记录缺少结果时显示“未知”。“已受理”仅表示请求被接受，不代表异步任务已完成。</p>
    </template>
    <AuditLogDetail v-model="detailOpen" :record="detail" :loading="detailLoading" :error="detailError" @retry="retryDetail" />
  </AdminPageShell>
</template>

<style scoped>
.audit-log-tabs :deep(.el-tabs__header) { margin: 0; }
.audit-log-tabs :deep(.el-tabs__content) { display: none; }
.audit-export-limit, .audit-note { font-size: var(--admin-ui-font-size-caption); color: var(--admin-ui-color-text-muted); }
.audit-note { margin: 0; line-height: 1.6; }
.audit-pagination { width: 100%; min-width: 0; overflow-x: auto; display: flex; justify-content: flex-end; }
@media (max-width: 767px) {
  .audit-pagination { justify-content: flex-end; }
  .audit-pagination :deep(.el-pagination) { flex-wrap: wrap; row-gap: var(--admin-ui-space-2); }
}
</style>
