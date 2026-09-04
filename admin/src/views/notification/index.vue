<template>
  <div class="notification-page">
    <header class="page-header">
      <div class="page-title">
        <el-icon><Bell /></el-icon>
        <div>
          <h2>通知管理</h2>
          <span>{{ unreadCount > 0 ? `${unreadCount} 条未读` : '暂无未读消息' }}</span>
        </div>
      </div>
      <div class="page-actions">
        <el-button v-if="unreadCount > 0 && canRead" :icon="CircleCheck" @click="markAllRead">全部已读</el-button>
        <el-button v-if="canSend" type="primary" :icon="Promotion" @click="openSendDialog('in_app')">发送站内信</el-button>
        <el-button v-if="canSendDingTalk" type="primary" plain :icon="ChatDotRound" @click="openSendDialog('dingtalk')">发送钉钉通知</el-button>
      </div>
    </header>

    <el-table v-loading="loading" :data="list" class="notification-table" empty-text="暂无站内信">
      <el-table-column label="标题" min-width="220">
        <template #default="{ row }">
          <div class="notification-title">
            <span v-if="!row.isRead" class="unread-dot" aria-label="未读" />
            <strong>{{ row.title }}</strong>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="内容" min-width="360">
        <template #default="{ row }">
          <div class="notification-content">{{ row.content }}</div>
        </template>
      </el-table-column>
      <el-table-column label="来源" width="120">
        <template #default="{ row }">{{ sourceLabel(row.sourceType) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="88" align="center">
        <template #default="{ row }">
          <el-tag :type="row.isRead ? 'info' : 'danger'" size="small">{{ row.isRead ? '已读' : '未读' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="发送时间" width="180">
        <template #default="{ row }">{{ formatTime(row.addTime) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" align="center">
        <template #default="{ row }">
          <el-button v-if="!row.isRead && canRead" link type="primary" @click="markRead(row.id)">标为已读</el-button>
          <span v-else class="operation-placeholder">-</span>
        </template>
      </el-table-column>
    </el-table>

    <footer v-if="total > pageSize" class="page-footer">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="load"
      />
    </footer>

    <el-dialog
      v-model="sendDialogVisible"
      :title="sendDialogTitle"
      width="min(640px, 92vw)"
      append-to-body
      destroy-on-close
      :close-on-click-modal="false"
    >
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="通知标题" required>
          <el-input v-model="sendForm.title" maxlength="255" show-word-limit placeholder="请输入通知标题" />
        </el-form-item>
        <el-form-item label="通知正文" required>
          <el-input
            v-model="sendForm.content"
            type="textarea"
            :rows="7"
            maxlength="5000"
            show-word-limit
            resize="vertical"
            placeholder="请输入通知正文"
          />
        </el-form-item>
        <el-form-item label="收件范围" required>
          <el-segmented v-model="sendForm.scope" :options="scopeOptions" block />
        </el-form-item>
        <el-form-item v-if="sendForm.scope === 'departments'" label="指定部门" required>
          <WorkflowUserTreePicker
            v-model="emptyUserSelection"
            v-model:department-model-value="departmentModelValue"
            :departments="recipientOptions.departments"
            :users="[]"
            :multiple="true"
            :select-department-rules="true"
            :disabled="recipientOptionsLoading"
            placeholder="请选择一个或多个部门"
          />
        </el-form-item>
        <el-form-item v-if="sendForm.scope === 'users'" label="指定用户" required>
          <WorkflowUserTreePicker
            v-model="sendForm.userIds"
            :departments="recipientOptions.departments"
            :users="recipientOptions.users"
            :multiple="true"
            :disabled="recipientOptionsLoading"
            placeholder="请选择一个或多个用户"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="sendDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="sending" @click="sendNotification">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { Bell, ChatDotRound, CircleCheck, Promotion } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { computed, onMounted, reactive, ref } from 'vue'
import { adminApi } from '../../api'
import type {
  InAppNotificationItem,
  InAppNotificationRecipientOptions,
  InAppNotificationScope,
} from '../../api/types'
import { hasPerm } from '../../utils/permission'
import WorkflowUserTreePicker from '../workflow/components/WorkflowUserTreePicker.vue'

const list = ref<InAppNotificationItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const unreadCount = ref(0)
const sendDialogVisible = ref(false)
const sending = ref(false)
const sendRequestID = ref('')
const sendChannel = ref<'in_app' | 'dingtalk'>('in_app')
const recipientOptionsLoading = ref(false)
const recipientOptionsLoadedFor = ref<'in_app' | 'dingtalk' | ''>('')
const emptyUserSelection = ref<number[]>([])
const recipientOptions = reactive<InAppNotificationRecipientOptions>({ users: [], departments: [] })
const sendForm = reactive({
  title: '',
  content: '',
  scope: 'all' as InAppNotificationScope,
  userIds: [] as number[],
  departmentIds: [] as number[],
})
const scopeOptions = [
  { label: '全部用户', value: 'all' },
  { label: '指定部门', value: 'departments' },
  { label: '指定用户', value: 'users' },
]
const canRead = computed(() => hasPerm('admin:menu:notification:read'))
const canSend = computed(() => hasPerm('admin:menu:notification:send'))
const canSendDingTalk = computed(() => hasPerm('admin:menu:notification:dingtalk-send'))
const sendDialogTitle = computed(() => sendChannel.value === 'dingtalk' ? '发送钉钉通知' : '发送站内信')
const departmentModelValue = computed<number[]>({
  get: () => sendForm.departmentIds,
  set: value => { sendForm.departmentIds = normalizeIDs(value) },
})

async function load() {
  loading.value = true
  try {
    const response = await adminApi.inAppNotificationList({ page: page.value, pageSize })
    list.value = Array.isArray(response.data?.list) ? response.data.list : []
    total.value = Number(response.data?.total || 0)
  } finally {
    loading.value = false
  }
}

async function loadUnreadCount() {
  const response = await adminApi.inAppNotificationUnreadCount()
  unreadCount.value = Number(response.data?.count || 0)
}

async function loadRecipientOptions() {
  if (recipientOptionsLoadedFor.value === sendChannel.value) return
  recipientOptionsLoading.value = true
  try {
    const response = sendChannel.value === 'dingtalk'
      ? await adminApi.dingTalkNotificationRecipientOptions()
      : await adminApi.inAppNotificationRecipientOptions()
    recipientOptions.users = Array.isArray(response.data?.users) ? response.data.users : []
    recipientOptions.departments = Array.isArray(response.data?.departments) ? response.data.departments : []
    recipientOptionsLoadedFor.value = sendChannel.value
  } finally {
    recipientOptionsLoading.value = false
  }
}

async function openSendDialog(channel: 'in_app' | 'dingtalk') {
  sendChannel.value = channel
  resetSendForm()
  sendDialogVisible.value = true
  await loadRecipientOptions()
}

async function sendNotification() {
  if (!sendForm.title.trim()) return ElMessage.warning('请输入通知标题')
  if (!sendForm.content.trim()) return ElMessage.warning('请输入通知正文')
  if (sendForm.scope === 'departments' && sendForm.departmentIds.length === 0) return ElMessage.warning('请选择收件部门')
  if (sendForm.scope === 'users' && sendForm.userIds.length === 0) return ElMessage.warning('请选择收件用户')
  sending.value = true
  try {
    const payload = {
      title: sendForm.title.trim(),
      content: sendForm.content,
      scope: sendForm.scope,
      userIds: sendForm.scope === 'users' ? normalizeIDs(sendForm.userIds) : undefined,
      departmentIds: sendForm.scope === 'departments' ? normalizeIDs(sendForm.departmentIds) : undefined,
    }
    const response = sendChannel.value === 'dingtalk'
      ? await adminApi.dingTalkNotificationSend(payload)
      : await adminApi.inAppNotificationSend({ ...payload, requestId: sendRequestID.value })
    const sentCount = Number(response.data?.sentCount || 0)
    const skippedCount = Number(response.data?.skippedCount || 0)
    const failedCount = Number(response.data?.failedCount || 0)
    if (sendChannel.value === 'dingtalk') {
      const detail = `成功 ${sentCount} 人，未绑定或未启用 ${skippedCount} 人，失败 ${failedCount} 人`
      if (failedCount > 0 || skippedCount > 0) ElMessage.warning(`钉钉通知发送完成：${detail}`)
      else ElMessage.success(`钉钉通知已发送给 ${sentCount} 人`)
      if (failedCount > 0 && sentCount === 0) return
    } else {
      ElMessage.success(`站内信已发送给 ${sentCount} 人`)
    }
    sendDialogVisible.value = false
    if (sendChannel.value === 'in_app') await Promise.all([load(), loadUnreadCount()])
  } finally {
    sending.value = false
  }
}

async function markRead(id: number) {
  await adminApi.inAppNotificationMarkRead(id)
  await Promise.all([load(), loadUnreadCount()])
}

async function markAllRead() {
  await adminApi.inAppNotificationMarkAllRead()
  await Promise.all([load(), loadUnreadCount()])
}

function resetSendForm() {
  sendRequestID.value = newRequestID()
  sendForm.title = ''
  sendForm.content = ''
  sendForm.scope = 'all'
  sendForm.userIds = []
  sendForm.departmentIds = []
  emptyUserSelection.value = []
}

function normalizeIDs(values: Array<number | string>) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value > 0)))
}

function newRequestID() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function sourceLabel(sourceType: string) {
  if (sourceType === 'admin_manual') return '后台发送'
  if (sourceType === 'scheduled_task_run') return '定时任务'
  if (sourceType === 'workflow') return '流程通知'
  return '系统通知'
}

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

onMounted(() => Promise.all([load(), loadUnreadCount()]))
</script>

<style scoped>
.notification-page { min-height: 100%; padding: 20px; background: var(--admin-bg); }
.page-header { display: flex; align-items: center; justify-content: space-between; gap: 20px; padding: 0 0 18px; border-bottom: 1px solid var(--admin-border); }
.page-title { display: flex; min-width: 0; align-items: center; gap: 12px; }
.page-title > .el-icon { width: 36px; height: 36px; flex: 0 0 36px; border-radius: 6px; background: #e9f8f4; color: #07866f; font-size: 18px; }
.page-title h2 { margin: 0; color: var(--admin-text); font-size: 18px; line-height: 26px; }
.page-title span { color: var(--admin-muted); font-size: 12px; }
.page-actions { display: flex; flex: 0 0 auto; gap: 8px; }
.notification-table { margin-top: 18px; border: 1px solid var(--admin-border); border-radius: 6px; }
.notification-title { display: flex; min-width: 0; align-items: center; gap: 9px; }
.notification-title strong { overflow: hidden; color: var(--admin-text); text-overflow: ellipsis; white-space: nowrap; }
.unread-dot { width: 7px; height: 7px; flex: 0 0 7px; border-radius: 50%; background: #e5484d; }
.notification-content { display: -webkit-box; overflow: hidden; color: var(--admin-text-secondary); line-height: 20px; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
.operation-placeholder { color: var(--admin-muted); }
.page-footer { display: flex; justify-content: flex-end; padding-top: 16px; }
@media (max-width: 720px) {
  .notification-page { padding: 14px; }
  .page-header { align-items: flex-start; }
  .page-actions { flex-direction: column; }
}
</style>
