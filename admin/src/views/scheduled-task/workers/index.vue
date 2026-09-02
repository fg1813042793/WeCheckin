<template>
  <div class="admin-page scheduled-worker-page">
    <el-card class="admin-card" shadow="never">
      <div class="worker-summary">
        <div><span>在线节点</span><strong>{{ onlineCount }}</strong></div>
        <div><span>总并发槽位</span><strong>{{ totalCapacity }}</strong></div>
        <div><span>正在执行</span><strong>{{ currentRuns }}</strong></div>
      </div>
      <div class="admin-toolbar">
        <div class="admin-toolbar__left"><span class="admin-muted">节点由 Redis TTL 心跳发现，HTTP 服务不会出现在此列表。</span></div>
        <div class="admin-toolbar__right">
          <el-switch v-model="autoRefresh" inline-prompt active-text="自动" inactive-text="手动" />
          <el-button circle icon="Refresh" title="刷新" :loading="loading" @click="loadWorkers" />
        </div>
      </div>
      <el-table v-loading="loading" :data="workers" stripe>
        <el-table-column label="节点" min-width="260"><template #default="{ row }"><div class="worker-name"><span :class="['worker-dot', { 'worker-dot--stale': !isOnline(row) }]" /><div><strong>{{ row.workerId }}</strong><small>{{ row.role }}</small></div></div></template></el-table-column>
        <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="isOnline(row) ? 'success' : 'danger'" size="small">{{ isOnline(row) ? '在线' : '心跳超时' }}</el-tag></template></el-table-column>
        <el-table-column label="执行负载" min-width="220"><template #default="{ row }"><div class="load-cell"><el-progress :percentage="capacityPercent(row)" :stroke-width="8" :show-text="false" /><span>{{ row.currentRuns }} / {{ row.workerCount }}</span></div></template></el-table-column>
        <el-table-column prop="version" label="版本" width="130"><template #default="{ row }">{{ row.version || '-' }}</template></el-table-column>
        <el-table-column label="启动时间" width="180"><template #default="{ row }">{{ formatTime(row.startedAt) }}</template></el-table-column>
        <el-table-column label="最近心跳" width="190"><template #default="{ row }"><span>{{ formatTime(row.lastHeartbeat) }}</span><small class="heartbeat-age">{{ heartbeatAge(row) }}</small></template></el-table-column>
      </el-table>
      <el-empty v-if="!loading && workers.length === 0" description="暂无在线 taskd worker" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { adminApi } from '../../../api'
import type { ScheduledTaskWorker } from '../types'

const loading = ref(false)
const workers = ref<ScheduledTaskWorker[]>([])
const autoRefresh = ref(true)
const now = ref(Date.now())
let clockTimer = 0
let refreshTimer = 0

const onlineCount = computed(() => workers.value.filter(isOnline).length)
const totalCapacity = computed(() => workers.value.reduce((sum, item) => sum + Number(item.workerCount || 0), 0))
const currentRuns = computed(() => workers.value.reduce((sum, item) => sum + Number(item.currentRuns || 0), 0))

onMounted(() => {
  loadWorkers()
  clockTimer = window.setInterval(() => { now.value = Date.now() }, 1000)
  startAutoRefresh()
})
onUnmounted(() => { window.clearInterval(clockTimer); window.clearInterval(refreshTimer) })
watch(autoRefresh, startAutoRefresh)

function startAutoRefresh() {
  window.clearInterval(refreshTimer)
  if (autoRefresh.value) refreshTimer = window.setInterval(loadWorkers, 15000)
}

async function loadWorkers() {
  loading.value = true
  try {
    const response = await adminApi.scheduledTaskWorkers()
    workers.value = Array.isArray(response.data) ? response.data : []
    now.value = Date.now()
  } finally { loading.value = false }
}

function isOnline(worker: ScheduledTaskWorker) { return now.value - worker.lastHeartbeat < 60000 }
function capacityPercent(worker: ScheduledTaskWorker) { return worker.workerCount > 0 ? Math.min(100, Math.round(worker.currentRuns / worker.workerCount * 100)) : 0 }
function heartbeatAge(worker: ScheduledTaskWorker) { const seconds = Math.max(0, Math.floor((now.value - worker.lastHeartbeat) / 1000)); return `${seconds} 秒前` }
function formatTime(value: number) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-' }
</script>

<style scoped>
.worker-summary { display: flex; align-items: center; gap: 0; margin: -4px 0 20px; border-bottom: 1px solid var(--admin-border); }
.worker-summary > div { display: flex; align-items: baseline; gap: 10px; min-width: 150px; padding: 14px 24px 16px 0; margin-right: 24px; border-right: 1px solid var(--admin-border); }
.worker-summary > div:last-child { border-right: 0; }
.worker-summary span { color: var(--admin-muted); font-size: 13px; }
.worker-summary strong { font-size: 22px; font-weight: 650; }
.worker-name { display: flex; align-items: center; gap: 10px; }
.worker-name div { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.worker-name small, .heartbeat-age { display: block; color: var(--admin-muted); }
.worker-dot { flex: 0 0 9px; width: 9px; height: 9px; border-radius: 50%; background: #16a34a; box-shadow: 0 0 0 3px #dcfce7; }
.worker-dot--stale { background: #dc2626; box-shadow: 0 0 0 3px #fee2e2; }
.load-cell { display: grid; grid-template-columns: minmax(100px, 1fr) 52px; align-items: center; gap: 10px; }
@media (max-width: 720px) { .worker-summary { align-items: stretch; flex-direction: column; } .worker-summary > div { border-right: 0; margin-right: 0; } }
</style>
