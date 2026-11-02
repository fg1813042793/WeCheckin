<template>
  <div class="dingtalk-delivery-records">
    <div class="admin-toolbar record-filters">
      <div class="admin-toolbar__left">
        <el-input v-model="filters.title" clearable placeholder="通知标题" style="width: 210px" @keyup.enter="search" />
        <el-select v-model="filters.sourceType" clearable placeholder="全部来源" style="width: 150px">
          <el-option label="后台发送" value="admin_manual_dingtalk" />
          <el-option label="定时任务" value="scheduled_task_run" />
          <el-option label="用户反馈" value="user_feedback" />
        </el-select>
        <el-select v-model="filters.status" clearable placeholder="全部状态" style="width: 150px">
          <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <AdminFilterActions @search="search" @reset="resetFilters" />
      </div>
      <div class="admin-toolbar__right">
        <el-button circle :icon="Refresh" title="刷新" :loading="loading" @click="load" />
      </div>
    </div>

    <el-table v-loading="loading" :data="list" class="notification-table" empty-text="暂无钉钉投递记录" stripe>
      <el-table-column prop="id" label="记录 ID" width="92" />
      <el-table-column prop="title" label="标题" min-width="210" show-overflow-tooltip />
      <el-table-column label="发送模板" min-width="190" show-overflow-tooltip>
        <template #default="{ row }">{{ templateLabel(row) }}</template>
      </el-table-column>
      <el-table-column label="来源" width="110">
        <template #default="{ row }">{{ sourceLabel(row.sourceType) }}</template>
      </el-table-column>
      <el-table-column prop="sourceId" label="来源标识" min-width="170" show-overflow-tooltip />
      <el-table-column label="状态" width="105" align="center">
        <template #default="{ row }"><el-tag :type="statusMeta(row.status).type" size="small">{{ statusMeta(row.status).label }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="plannedCount" label="计划人数" width="86" align="center" />
      <el-table-column prop="sentCount" label="成功人数" width="86" align="center" />
      <el-table-column prop="skippedCount" label="跳过人数" width="86" align="center" />
      <el-table-column prop="failedCount" label="失败人数" width="86" align="center" />
      <el-table-column label="发送时间" width="180">
        <template #default="{ row }">{{ formatTime(row.sentAt || row.addTime) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right" align="center">
        <template #default="{ row }">
          <el-button link type="primary" :icon="View" @click="openDetail(row)">查看</el-button>
          <el-button v-if="canDelete" link type="danger" :icon="Delete" :loading="deletingId === row.id" @click="deleteRecord(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="admin-pagination">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @current-change="load"
        @size-change="handlePageSizeChange"
      />
    </div>
  </div>

  <el-dialog v-model="detailVisible" title="钉钉投递详情" width="min(760px, 92vw)" append-to-body destroy-on-close>
    <el-descriptions v-if="detailRecord" :column="2" border class="record-detail">
      <el-descriptions-item label="记录 ID">{{ detailRecord.id }}</el-descriptions-item>
      <el-descriptions-item label="状态"><el-tag :type="statusMeta(detailRecord.status).type" size="small">{{ statusMeta(detailRecord.status).label }}</el-tag></el-descriptions-item>
      <el-descriptions-item label="标题" :span="2">{{ detailRecord.title }}</el-descriptions-item>
      <el-descriptions-item label="发送模板" :span="2">{{ templateLabel(detailRecord) }}</el-descriptions-item>
      <el-descriptions-item label="来源">{{ sourceLabel(detailRecord.sourceType) }}</el-descriptions-item>
      <el-descriptions-item label="来源标识"><span class="breakable">{{ detailRecord.sourceId }}</span></el-descriptions-item>
      <el-descriptions-item label="计划人数">{{ detailRecord.plannedCount }}</el-descriptions-item>
      <el-descriptions-item label="成功人数">{{ detailRecord.sentCount }}</el-descriptions-item>
      <el-descriptions-item label="跳过人数">{{ detailRecord.skippedCount }}</el-descriptions-item>
      <el-descriptions-item label="失败人数">{{ detailRecord.failedCount }}</el-descriptions-item>
      <el-descriptions-item label="执行次数">{{ detailRecord.attempts }}</el-descriptions-item>
      <el-descriptions-item label="发送时间">{{ formatTime(detailRecord.sentAt || detailRecord.addTime) }}</el-descriptions-item>
      <el-descriptions-item label="失败摘要" :span="2">{{ detailRecord.lastError || '-' }}</el-descriptions-item>
      <el-descriptions-item label="通知正文" :span="2"><div class="detail-content">{{ detailRecord.content }}</div></el-descriptions-item>
    </el-descriptions>
    <template #footer>
      <el-button v-if="detailRecord && canDelete" type="danger" plain :icon="Delete" :loading="deletingId === detailRecord.id" @click="deleteRecord(detailRecord)">删除记录</el-button>
      <el-button @click="detailVisible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { Delete, Refresh, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { AdminFilterActions } from '@/components/admin-ui'
import { onMounted, reactive, ref } from 'vue'
import { adminApi } from '../../../api'
import type { DingTalkDeliveryRecordItem, DingTalkDeliveryStatus } from '../../../api/types'

defineProps<{ canDelete?: boolean }>()

type TagType = 'success' | 'warning' | 'danger' | 'info'

const list = ref<DingTalkDeliveryRecordItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const deletingId = ref<number | null>(null)
const detailVisible = ref(false)
const detailRecord = ref<DingTalkDeliveryRecordItem | null>(null)
const filters = reactive({ title: '', sourceType: '', status: '' as '' | DingTalkDeliveryStatus })
const statusOptions: Array<{ label: string; value: DingTalkDeliveryStatus; type: TagType }> = [
  { label: '已发送', value: 'sent', type: 'success' },
  { label: '部分成功', value: 'partial', type: 'warning' },
  { label: '发送失败', value: 'failed_final', type: 'danger' },
  { label: '全部跳过', value: 'skipped', type: 'info' },
]

onMounted(load)

async function load() {
  loading.value = true
  try {
    const response = await adminApi.dingTalkNotificationRecordList({
      page: page.value, pageSize: pageSize.value,
      title: filters.title.trim() || undefined,
      sourceType: filters.sourceType || undefined,
      status: filters.status || undefined,
    })
    list.value = Array.isArray(response.data?.list) ? response.data.list : []
    total.value = Number(response.data?.total || 0)
  } finally {
    loading.value = false
  }
}

function search() { page.value = 1; void load() }
function resetFilters() { filters.title = ''; filters.sourceType = ''; filters.status = ''; search() }
function handlePageSizeChange() { page.value = 1; void load() }

async function openDetail(row: DingTalkDeliveryRecordItem) {
  detailVisible.value = true
  const response = await adminApi.dingTalkNotificationRecordDetail(row.id)
  detailRecord.value = response.data || row
}

async function deleteRecord(row: DingTalkDeliveryRecordItem) {
  try {
    await ElMessageBox.confirm('删除后只移除后台投递记录，不会撤回已经发送的钉钉通知。', '删除钉钉投递记录', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
  } catch { return }
  deletingId.value = row.id
  try {
    await adminApi.dingTalkNotificationRecordDelete(row.id)
    if (detailRecord.value?.id === row.id) { detailVisible.value = false; detailRecord.value = null }
    if (list.value.length === 1 && page.value > 1) page.value -= 1
    await load()
    ElMessage.success('钉钉投递记录已删除')
  } finally { deletingId.value = null }
}

function statusMeta(status: DingTalkDeliveryStatus) {
  return statusOptions.find(option => option.value === status) || { label: status || '-', type: 'info' as const }
}

function sourceLabel(sourceType: string) {
  if (sourceType === 'scheduled_task_run') return '定时任务'
  if (sourceType === 'admin_manual_dingtalk') return '后台发送'
  if (sourceType === 'user_feedback') return '用户反馈'
  return sourceType || '-'
}

function templateLabel(row: DingTalkDeliveryRecordItem) {
  if (!row.templateId) return '系统通知样式'
  return `${row.templateName || `模板 #${row.templateId}`} · v${row.templateVersion || '-'}`
}

function formatTime(value?: number) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'
}
</script>

<style scoped>
.dingtalk-delivery-records { min-width: 0; }
.breakable { overflow-wrap: anywhere; }
.detail-content { max-height: 220px; overflow: auto; white-space: pre-wrap; overflow-wrap: anywhere; }
</style>
