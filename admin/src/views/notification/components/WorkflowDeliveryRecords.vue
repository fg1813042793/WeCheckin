<template>
  <div class="workflow-delivery-records">
    <div class="admin-toolbar delivery-filters">
      <div class="admin-toolbar__left">
        <el-input v-model="filters.instanceId" clearable placeholder="流程实例 ID" style="width: 210px" @keyup.enter="search" />
        <el-input v-model="filters.recipientUserId" clearable placeholder="接收人 ID" style="width: 150px" @keyup.enter="search" />
        <el-select v-model="filters.kind" clearable placeholder="全部事件" style="width: 160px">
          <el-option v-for="option in kindOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-select v-model="filters.channel" clearable placeholder="全部渠道" style="width: 140px">
          <el-option v-for="option in channelOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-select v-model="filters.status" clearable placeholder="全部状态" style="width: 140px">
          <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="search">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
      <div class="admin-toolbar__right">
        <el-button
          v-if="canRetry"
          type="primary"
          plain
          :icon="RefreshRight"
          :loading="dispatching"
          @click="dispatchDue"
        >处理待投递通知</el-button>
        <el-button circle :icon="Refresh" title="刷新" :loading="loading" @click="load" />
      </div>
    </div>

    <el-table v-loading="loading" :data="list" class="delivery-table" empty-text="暂无流程通知投递记录" stripe>
      <el-table-column label="通知标题" min-width="190" show-overflow-tooltip>
        <template #default="{ row }">{{ row.payload.title || '-' }}</template>
      </el-table-column>
      <el-table-column label="流程单号" min-width="210" show-overflow-tooltip>
        <template #default="{ row }">{{ row.businessKey || '-' }}</template>
      </el-table-column>
      <el-table-column prop="instanceId" label="流程实例" min-width="210" show-overflow-tooltip />
      <el-table-column label="接收人" min-width="130" show-overflow-tooltip>
        <template #default="{ row }">{{ recipientLabel(row) }}</template>
      </el-table-column>
      <el-table-column label="事件" width="130">
        <template #default="{ row }">
          <el-tag :type="kindMeta(row.kind).type" effect="plain" size="small">{{ kindMeta(row.kind).label }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="渠道" width="105">
        <template #default="{ row }">{{ channelLabel(row.channel) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusMeta(row.status).type" size="small">{{ statusMeta(row.status).label }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="attempts" label="尝试" width="70" align="center" />
      <el-table-column label="最近错误" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">{{ row.lastError || '-' }}</template>
      </el-table-column>
      <el-table-column label="发送时间" width="180">
        <template #default="{ row }">{{ formatTime(row.sentAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right" align="center">
        <template #default="{ row }">
          <el-button link type="primary" :icon="View" @click="openDetail(row)">查看</el-button>
          <el-button
            v-if="canRetry && row.status !== 'sending'"
            link
            type="primary"
            :icon="Promotion"
            :loading="sendingId === row.id"
            @click="manualSend(row)"
          >{{ isRetryable(row) ? '重发' : '发送' }}</el-button>
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

  <el-dialog v-model="detailVisible" title="投递详情" width="min(760px, 92vw)" append-to-body destroy-on-close>
    <el-descriptions v-if="detailRecord" :column="2" border class="delivery-detail">
      <el-descriptions-item label="投递记录 ID" :span="2"><span class="breakable">{{ detailRecord.id }}</span></el-descriptions-item>
      <el-descriptions-item label="流程单号" :span="2"><span class="breakable">{{ detailRecord.businessKey || '-' }}</span></el-descriptions-item>
      <el-descriptions-item label="流程实例" :span="2"><span class="breakable">{{ detailRecord.instanceId }}</span></el-descriptions-item>
      <el-descriptions-item label="节点 ID">{{ detailRecord.nodeId || '-' }}</el-descriptions-item>
      <el-descriptions-item label="任务 ID">{{ detailRecord.taskId || '-' }}</el-descriptions-item>
      <el-descriptions-item label="接收人">{{ recipientLabel(detailRecord) }}</el-descriptions-item>
      <el-descriptions-item label="接收人 ID">{{ detailRecord.recipientUserId || '-' }}</el-descriptions-item>
      <el-descriptions-item label="事件">{{ kindMeta(detailRecord.kind).label }}</el-descriptions-item>
      <el-descriptions-item label="渠道">{{ channelLabel(detailRecord.channel) }}</el-descriptions-item>
      <el-descriptions-item label="状态"><el-tag :type="statusMeta(detailRecord.status).type" size="small">{{ statusMeta(detailRecord.status).label }}</el-tag></el-descriptions-item>
      <el-descriptions-item label="尝试次数">{{ detailRecord.attempts }}</el-descriptions-item>
      <el-descriptions-item label="创建时间">{{ formatTime(detailRecord.addTime) }}</el-descriptions-item>
      <el-descriptions-item label="发送时间">{{ formatTime(detailRecord.sentAt) }}</el-descriptions-item>
      <el-descriptions-item label="下次重试">{{ formatTime(detailRecord.nextRetryAt) }}</el-descriptions-item>
      <el-descriptions-item label="平台消息 ID">{{ detailRecord.providerMessageId || '-' }}</el-descriptions-item>
      <el-descriptions-item label="通知标题" :span="2">{{ detailRecord.payload.title || '-' }}</el-descriptions-item>
      <el-descriptions-item label="通知内容" :span="2"><div class="detail-content">{{ detailRecord.payload.content || '-' }}</div></el-descriptions-item>
      <el-descriptions-item label="最近错误" :span="2"><div class="detail-error">{{ detailRecord.lastError || '-' }}</div></el-descriptions-item>
    </el-descriptions>
    <template #footer>
      <el-button
        v-if="detailRecord && canRetry && detailRecord.status !== 'sending'"
        type="primary"
        plain
        :icon="Promotion"
        :loading="sendingId === detailRecord.id"
        @click="manualSend(detailRecord)"
      >{{ isRetryable(detailRecord) ? '重发通知' : '发送通知' }}</el-button>
      <el-button @click="detailVisible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { Promotion, Refresh, RefreshRight, Search, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import { adminApi } from '../../../api'
import type {
  WorkflowNotificationChannel,
  WorkflowNotificationKind,
  WorkflowNotificationRecord,
  WorkflowNotificationStatus,
} from '../../workflow/types'

defineProps<{ canRetry?: boolean }>()

type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

const kindOptions: Array<{ label: string; value: WorkflowNotificationKind; type: TagType }> = [
  { label: '任务到达', value: 'task_arrived', type: 'warning' },
  { label: '处理提醒', value: 'task_reminder', type: 'warning' },
  { label: '审批通过结果', value: 'approval_result_approved', type: 'success' },
  { label: '审批驳回结果', value: 'approval_result_rejected', type: 'danger' },
  { label: '审批退回结果', value: 'approval_result_returned', type: 'warning' },
  { label: '节点抄送', value: 'node_cc', type: 'info' },
  { label: '节点通知', value: 'node_notify', type: 'primary' },
  { label: '流程评论', value: 'instance_commented', type: 'primary' },
  { label: '表单修改', value: 'instance_form_revised', type: 'primary' },
]
const channelOptions: Array<{ label: string; value: WorkflowNotificationChannel }> = [
  { label: '站内通知', value: 'in_app' },
  { label: '钉钉通知', value: 'dingtalk_oa' },
]
const statusOptions: Array<{ label: string; value: WorkflowNotificationStatus; type: TagType }> = [
  { label: '待发送', value: 'pending', type: 'warning' },
  { label: '发送中', value: 'sending', type: 'primary' },
  { label: '已发送', value: 'sent', type: 'success' },
  { label: '待重试', value: 'failed', type: 'warning' },
  { label: '已停止', value: 'dead', type: 'danger' },
]

const list = ref<WorkflowNotificationRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const dispatching = ref(false)
const sendingId = ref('')
const detailVisible = ref(false)
const detailRecord = ref<WorkflowNotificationRecord | null>(null)
const filters = reactive({
  instanceId: '',
  recipientUserId: '',
  kind: '' as '' | WorkflowNotificationKind,
  channel: '' as '' | WorkflowNotificationChannel,
  status: '' as '' | WorkflowNotificationStatus,
})

async function load() {
  loading.value = true
  try {
    const response = await adminApi.workflowNotificationList({
      page: page.value,
      pageSize: pageSize.value,
      instanceId: filters.instanceId.trim() || undefined,
      recipientUserId: filters.recipientUserId.trim() || undefined,
      kind: filters.kind || undefined,
      channel: filters.channel || undefined,
      status: filters.status || undefined,
    })
    list.value = Array.isArray(response.data?.list) ? response.data.list : []
    total.value = Number(response.data?.total || 0)
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

function resetFilters() {
  filters.instanceId = ''
  filters.recipientUserId = ''
  filters.kind = ''
  filters.channel = ''
  filters.status = ''
  search()
}

function handlePageSizeChange() {
  page.value = 1
  load()
}

function openDetail(row: WorkflowNotificationRecord) {
  detailRecord.value = row
  detailVisible.value = true
}

function isRetryable(row: WorkflowNotificationRecord) {
  return row.status === 'failed' || row.status === 'dead'
}

async function manualSend(row: WorkflowNotificationRecord) {
  const actionLabel = isRetryable(row) ? '重新发送' : '发送'
  try {
    await ElMessageBox.confirm(
      `确定${actionLabel}“${row.payload.title || row.id}”给${recipientLabel(row)}吗？`,
      `${actionLabel}流程通知`,
      { type: 'warning', confirmButtonText: actionLabel, cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  sendingId.value = row.id
  try {
    if (isRetryable(row)) {
      await adminApi.workflowNotificationRetry(row.id)
    } else {
      await adminApi.workflowNotificationSend(row.id)
    }
    ElMessage.success('已执行发送，请查看最新投递状态')
    await load()
    if (detailRecord.value?.id === row.id) {
      const refreshedRecord = list.value.find(item => item.id === row.id)
      if (refreshedRecord) {
        detailRecord.value = refreshedRecord
      } else {
        detailVisible.value = false
        detailRecord.value = null
      }
    }
  } finally {
    sendingId.value = ''
  }
}

async function dispatchDue() {
  dispatching.value = true
  try {
    const response = await adminApi.workflowNotificationDispatchDue({ limit: 100 })
    ElMessage.success(`本次处理 ${Number(response.data?.dispatched || 0)} 条待投递通知`)
    await load()
  } finally {
    dispatching.value = false
  }
}

function recipientLabel(row: WorkflowNotificationRecord) {
  return row.recipientUserName?.trim() || (row.recipientUserId ? `用户 #${row.recipientUserId}` : '-')
}

function kindMeta(kind: WorkflowNotificationKind) {
  return kindOptions.find(option => option.value === kind) || { label: kind || '-', type: 'info' as TagType }
}

function channelLabel(channel: WorkflowNotificationChannel) {
  return channelOptions.find(option => option.value === channel)?.label || channel || '-'
}

function statusMeta(status: WorkflowNotificationStatus) {
  return statusOptions.find(option => option.value === status) || { label: status || '-', type: 'info' as TagType }
}

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

onMounted(load)
</script>

<style scoped>
.workflow-delivery-records { min-width: 0; }
.delivery-filters { margin-top: 0; padding-bottom: 16px; border-bottom: 1px solid var(--admin-border); }
.delivery-table { border: 1px solid var(--admin-border); border-radius: 6px; }
.breakable, .detail-content, .detail-error { overflow-wrap: anywhere; word-break: break-word; }
.detail-content, .detail-error { min-height: 54px; white-space: pre-wrap; line-height: 1.7; }
.detail-error { color: var(--el-color-danger); }
</style>
