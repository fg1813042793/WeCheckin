<template>
  <div class="admin-page scheduled-run-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar run-filters">
        <div class="admin-toolbar__left">
          <el-input v-model="filters.taskId" clearable placeholder="任务 ID" style="width: 120px" @keyup.enter="search" />
          <el-select v-model="filters.status" clearable placeholder="全部状态" style="width: 140px" @change="search">
            <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-select v-model="filters.triggerType" clearable placeholder="全部触发方式" style="width: 140px" @change="search">
            <el-option label="计划调度" value="scheduled" /><el-option label="错过补跑" value="misfire" /><el-option label="手工运行" value="manual" /><el-option label="手工重试" value="manual_retry" />
          </el-select>
          <el-input v-model="filters.workerId" clearable placeholder="Worker ID" style="width: 180px" @keyup.enter="search" />
          <el-date-picker v-model="filters.range" type="datetimerange" range-separator="至" start-placeholder="开始时间" end-placeholder="结束时间" style="width: 340px" />
          <el-button type="primary" icon="Search" @click="search">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
      </div>
      <div class="admin-toolbar">
        <div class="admin-toolbar__left"><span class="admin-muted">运行记录由 MySQL 保存，Redis 仅用于投递与节点心跳。</span></div>
        <div class="admin-toolbar__right"><el-button circle icon="Refresh" title="刷新" :loading="loading" @click="loadList" /></div>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column label="运行 ID" min-width="230"><template #default="{ row }"><button class="run-link" type="button" @click="openDetail(row)">{{ row.id }}</button><small class="run-parent" v-if="row.parentRunId">重试自 {{ row.parentRunId }}</small></template></el-table-column>
        <el-table-column prop="taskId" label="任务 ID" width="100" />
        <el-table-column label="触发方式" width="110"><template #default="{ row }">{{ triggerLabel(row.triggerType) }}</template></el-table-column>
        <el-table-column label="状态" width="115"><template #default="{ row }"><el-tag :type="statusMeta(row.status).type" size="small">{{ statusMeta(row.status).label }}</el-tag></template></el-table-column>
        <el-table-column label="计划时间" width="175"><template #default="{ row }">{{ formatTime(row.scheduledAt) }}</template></el-table-column>
        <el-table-column label="执行信息" min-width="210"><template #default="{ row }"><div class="execution-cell"><span>{{ row.workerId || '未分配 Worker' }}</span><small>尝试 {{ row.attempt + 1 }} 次<span v-if="row.coalescedCount > 1"> · 合并 {{ row.coalescedCount }} 次</span></small></div></template></el-table-column>
        <el-table-column label="结果" min-width="220" show-overflow-tooltip><template #default="{ row }"><span :class="{ 'error-text': row.errorSummary }">{{ row.errorSummary || row.resultSummary || '-' }}</span></template></el-table-column>
        <el-table-column label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-button size="small" icon="View" @click="openDetail(row)">详情</el-button>
              <el-button v-if="canRetry && row.status === 'failed'" size="small" type="primary" plain icon="RefreshRight" @click="retry(row)">重试</el-button>
              <el-button v-if="canCancel && cancelable(row.status)" size="small" type="danger" plain icon="Close" @click="cancelRun(row)">取消</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="admin-pagination"><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :page-sizes="[10, 20, 50, 100]" :total="total" layout="total,sizes,prev,pager,next" @current-change="loadList" @size-change="changePageSize" /></div>
    </el-card>

    <el-drawer v-model="detailVisible" title="运行详情" size="min(720px, 94vw)" append-to-body destroy-on-close>
      <div v-loading="detailLoading">
        <template v-if="detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="运行 ID" :span="2"><code>{{ detail.run.id }}</code></el-descriptions-item>
            <el-descriptions-item label="任务 ID">{{ detail.run.taskId }}</el-descriptions-item>
            <el-descriptions-item label="状态"><el-tag :type="statusMeta(detail.run.status).type" size="small">{{ statusMeta(detail.run.status).label }}</el-tag></el-descriptions-item>
            <el-descriptions-item label="触发方式">{{ triggerLabel(detail.run.triggerType) }}</el-descriptions-item>
            <el-descriptions-item label="Worker">{{ detail.run.workerId || '-' }}</el-descriptions-item>
            <el-descriptions-item label="计划时间">{{ formatTime(detail.run.scheduledAt) }}</el-descriptions-item>
            <el-descriptions-item label="开始时间">{{ formatTime(detail.run.startedAt) }}</el-descriptions-item>
            <el-descriptions-item label="结束时间">{{ formatTime(detail.run.finishedAt) }}</el-descriptions-item>
            <el-descriptions-item label="下次重试">{{ formatTime(detail.run.nextRetryAt) }}</el-descriptions-item>
            <el-descriptions-item label="结果" :span="2">{{ detail.run.resultSummary || '-' }}</el-descriptions-item>
            <el-descriptions-item label="错误" :span="2"><span class="error-text">{{ detail.run.errorSummary || '-' }}</span></el-descriptions-item>
          </el-descriptions>
          <div class="log-heading"><strong>分段日志</strong><span>{{ detail.logs.length }} 条</span></div>
          <el-empty v-if="detail.logs.length === 0" description="暂无运行日志" />
          <div v-else class="run-logs">
            <div v-for="log in detail.logs" :key="log.id" class="run-log">
              <div class="run-log__meta"><span>#{{ log.sequence }} · {{ formatTime(log.logTime) }}</span><el-tag :type="logLevelType(log.level)" size="small">{{ log.level }}</el-tag><span>{{ log.stage || '-' }}</span></div>
              <pre>{{ log.content }}</pre>
            </div>
          </div>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../../api'
import { hasPerm } from '../../../utils/permission'
import type { ScheduledTaskRun, ScheduledTaskRunDetail } from '../types'

const statusOptions = [
  { label: '等待中', value: 'waiting' }, { label: '已入队', value: 'queued' }, { label: '运行中', value: 'running' },
  { label: '等待重试', value: 'retry_wait' }, { label: '成功', value: 'success' }, { label: '失败', value: 'failed' },
  { label: '已取消', value: 'canceled' }, { label: '已跳过', value: 'skipped' },
]
const loading = ref(false)
const list = ref<ScheduledTaskRun[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<ScheduledTaskRunDetail | null>(null)
const filters = reactive<{ taskId: string; status: string; triggerType: string; workerId: string; range: [Date, Date] | null }>({ taskId: '', status: '', triggerType: '', workerId: '', range: null })
const canRetry = computed(() => hasPerm('admin:menu:scheduled-task:run:retry'))
const canCancel = computed(() => hasPerm('admin:menu:scheduled-task:run:cancel'))

onMounted(loadList)

async function loadList() {
  loading.value = true
  try {
    const response = await adminApi.scheduledTaskRunList({
      page: page.value, pageSize: pageSize.value, taskId: filters.taskId, status: filters.status,
      triggerType: filters.triggerType, workerId: filters.workerId.trim(),
      startTime: filters.range?.[0]?.getTime(), endTime: filters.range?.[1]?.getTime(),
    })
    list.value = Array.isArray(response.data?.list) ? response.data.list : []
    total.value = Number(response.data?.total || 0)
  } finally { loading.value = false }
}

function search() { page.value = 1; loadList() }
function resetFilters() { Object.assign(filters, { taskId: '', status: '', triggerType: '', workerId: '', range: null }); search() }
function changePageSize() { page.value = 1; loadList() }

async function openDetail(row: ScheduledTaskRun) {
  detailVisible.value = true
  detailLoading.value = true
  detail.value = null
  try {
    const response = await adminApi.scheduledTaskRunDetail(row.id)
    detail.value = response.data as ScheduledTaskRunDetail
  } finally { detailLoading.value = false }
}

async function retry(row: ScheduledTaskRun) {
  await ElMessageBox.confirm('将创建一条关联原运行的新手工重试记录。', '重试运行', { type: 'warning', confirmButtonText: '重试', cancelButtonText: '取消' })
  const response = await adminApi.scheduledTaskRunRetry(row.id)
  ElMessage.success(response.data?.dispatchPending ? '重试记录已创建，等待恢复投递' : '重试记录已创建')
  await loadList()
}

async function cancelRun(row: ScheduledTaskRun) {
  await ElMessageBox.confirm(row.status === 'running' ? '将请求 worker 中断当前处理器。' : '该运行尚未执行，将直接取消。', '取消运行', { type: 'warning', confirmButtonText: '取消运行', cancelButtonText: '返回' })
  await adminApi.scheduledTaskRunCancel(row.id)
  ElMessage.success(row.status === 'running' ? '已提交取消请求' : '运行已取消')
  await loadList()
}

function cancelable(status: string) { return ['waiting', 'queued', 'running', 'retry_wait'].includes(status) }
function triggerLabel(value: string) { return ({ scheduled: '计划调度', misfire: '错过补跑', manual: '手工运行', manual_retry: '手工重试' } as Record<string, string>)[value] || value }
function statusMeta(value: string) {
  const map: Record<string, { label: string; type: 'success' | 'warning' | 'danger' | 'info' | 'primary' }> = {
    waiting: { label: '等待中', type: 'info' }, queued: { label: '已入队', type: 'primary' }, running: { label: '运行中', type: 'warning' },
    retry_wait: { label: '等待重试', type: 'warning' }, success: { label: '成功', type: 'success' }, failed: { label: '失败', type: 'danger' },
    canceled: { label: '已取消', type: 'info' }, skipped: { label: '已跳过', type: 'info' },
  }
  return map[value] || { label: value, type: 'info' as const }
}
function logLevelType(value: string) { return value === 'error' ? 'danger' : value === 'warn' ? 'warning' : value === 'debug' ? 'info' : 'primary' }
function formatTime(value: number) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-' }
</script>

<style scoped>
.run-link { display: block; max-width: 100%; padding: 0; overflow: hidden; color: #2563eb; background: none; border: 0; text-overflow: ellipsis; white-space: nowrap; cursor: pointer; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.run-parent { display: block; margin-top: 4px; color: var(--admin-muted); }
.execution-cell { display: flex; flex-direction: column; gap: 4px; }
.execution-cell small { color: var(--admin-muted); }
.error-text { color: #dc2626; }
.log-heading { display: flex; align-items: center; justify-content: space-between; margin: 24px 0 12px; }
.log-heading span { color: var(--admin-muted); font-size: 13px; }
.run-logs { display: grid; gap: 10px; }
.run-log { border: 1px solid var(--admin-border); border-radius: 6px; overflow: hidden; }
.run-log__meta { display: flex; align-items: center; gap: 8px; padding: 8px 10px; background: #f7f8fa; color: var(--admin-muted); font-size: 12px; }
.run-log pre { margin: 0; padding: 10px; overflow: auto; white-space: pre-wrap; word-break: break-word; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
</style>
