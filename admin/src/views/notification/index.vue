<template>
  <div class="admin-page notification-page">
    <el-card class="admin-card" shadow="never">
      <header class="page-header">
        <div class="page-title">
          <el-icon><Bell /></el-icon>
          <div>
            <h2>通知记录管理</h2>
            <span>{{ recordSummary }}</span>
          </div>
        </div>
        <div class="page-actions">
          <el-button v-if="canStyleList" :icon="Brush" @click="styleDialogVisible = true">消息样式</el-button>
          <el-button v-if="canSend" type="primary" :icon="Promotion" @click="openSendDialog('in_app')">发送站内信</el-button>
          <el-button v-if="canSendDingTalk" type="primary" plain :icon="ChatDotRound" @click="openSendDialog('dingtalk')">发送钉钉通知</el-button>
        </div>
      </header>

      <el-tabs v-model="activeRecordView" class="record-view-tabs" @tab-change="handleRecordViewChange">
        <el-tab-pane label="站内信记录" name="in_app" />
        <el-tab-pane label="Webhook 投递" name="webhook" />
        <el-tab-pane v-if="canViewWorkflowNotifications" label="流程投递" name="workflow" />
      </el-tabs>

      <template v-if="activeRecordView === 'in_app'">
        <div class="admin-toolbar record-filters">
        <div class="admin-toolbar__left">
          <el-input v-model="filters.title" clearable placeholder="通知标题" style="width: 180px" @keyup.enter="search" />
          <el-input v-model="filters.recipientName" clearable placeholder="接收人用户名" style="width: 180px" @keyup.enter="search" />
          <el-select v-model="filters.sourceType" clearable filterable allow-create default-first-option placeholder="全部来源" style="width: 150px">
            <el-option v-for="option in sourceOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
          <el-select v-model="filters.notificationType" clearable filterable allow-create default-first-option placeholder="全部消息类型" style="width: 160px">
            <el-option v-for="option in notificationTypeOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
          <el-select v-model="filters.isRead" clearable placeholder="全部阅读状态" style="width: 140px">
            <el-option label="未读" :value="0" />
            <el-option label="已读" :value="1" />
          </el-select>
          <el-date-picker
            v-model="addTimeRange"
            type="datetimerange"
            value-format="x"
            range-separator="至"
            start-placeholder="发送开始时间"
            end-placeholder="发送结束时间"
            style="width: 340px"
          />
          <el-button type="primary" :icon="Search" @click="search">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button v-if="canDelete" type="danger" plain :icon="Delete" :disabled="inAppSelectedRows.length === 0" @click="deleteSelectedInApp">批量删除</el-button>
          <el-button circle :icon="Refresh" title="刷新" :loading="loading" @click="load" />
        </div>
        </div>

        <el-table v-loading="loading" :data="list" class="notification-table" empty-text="暂无通知投递记录" stripe @selection-change="handleInAppSelectionChange">
        <el-table-column type="selection" width="48" />
        <el-table-column prop="id" label="记录 ID" width="92" />
        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        <el-table-column label="内容" min-width="300">
          <template #default="{ row }">
            <div class="notification-content">{{ row.content }}</div>
          </template>
        </el-table-column>
        <el-table-column label="接收人" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ recipientLabel(row) }}</template>
        </el-table-column>
        <el-table-column label="消息类型" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ notificationTypeLabel(row.type) }}</template>
        </el-table-column>
        <el-table-column label="来源" width="120">
          <template #default="{ row }">{{ sourceLabel(row.sourceType) }}</template>
        </el-table-column>
        <el-table-column prop="sourceId" label="来源标识" min-width="170" show-overflow-tooltip />
        <el-table-column label="阅读状态" width="96" align="center">
          <template #default="{ row }">
            <el-tag :type="row.isRead ? 'info' : 'danger'" size="small">{{ row.isRead ? '已读' : '未读' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发送时间" width="180">
          <template #default="{ row }">{{ formatTime(row.addTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="158" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" :icon="View" @click="openDetail(row)">查看</el-button>
            <el-button
              v-if="canDelete"
              link
              type="danger"
              :icon="Delete"
              :loading="deletingID === row.id"
              @click="deleteRecord(row)"
            >删除</el-button>
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
      </template>

      <template v-else-if="activeRecordView === 'webhook'">
        <div class="admin-toolbar record-filters">
          <div class="admin-toolbar__left">
            <el-select v-model="webhookStatus" clearable placeholder="全部状态" style="width: 150px" @change="loadWebhookRecords">
              <el-option label="待发送" value="pending" />
              <el-option label="发送中" value="sending" />
              <el-option label="已发送" value="sent" />
              <el-option label="待重试" value="failed" />
              <el-option label="已终止" value="dead" />
            </el-select>
            <el-button type="primary" :icon="Search" @click="loadWebhookRecords">查询</el-button>
            <el-button @click="webhookStatus = ''; loadWebhookRecords()">重置</el-button>
          </div>
          <div class="admin-toolbar__right">
            <el-button v-if="canDelete" type="danger" plain :icon="Delete" :disabled="webhookSelectedRows.length === 0" @click="deleteSelectedWebhook">批量删除</el-button>
            <el-button circle :icon="Refresh" title="刷新" :loading="webhookLoading" @click="loadWebhookRecords" />
          </div>
        </div>

        <el-table v-loading="webhookLoading" :data="webhookList" class="notification-table" empty-text="暂无 Webhook 投递记录" stripe @selection-change="handleWebhookSelectionChange">
          <el-table-column type="selection" width="48" :selectable="selectableWebhook" />
          <el-table-column prop="id" label="记录 ID" width="92" />
          <el-table-column prop="title" label="标题" min-width="220" show-overflow-tooltip />
          <el-table-column prop="sourceId" label="来源标识" min-width="150" show-overflow-tooltip />
          <el-table-column prop="webhookType" label="类型" width="110" />
          <el-table-column prop="target" label="目标" min-width="190" show-overflow-tooltip />
          <el-table-column label="状态" width="100" align="center">
            <template #default="{ row }"><el-tag :type="webhookStatusType(row.status)" size="small">{{ webhookStatusLabel(row.status) }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="attempts" label="重试" width="70" align="center" />
          <el-table-column label="失败原因" min-width="220" show-overflow-tooltip>
            <template #default="{ row }">{{ row.lastError || '-' }}</template>
          </el-table-column>
          <el-table-column label="创建时间" width="180">
            <template #default="{ row }">{{ formatTime(row.addTime) }}</template>
          </el-table-column>
          <el-table-column label="发送时间" width="180">
            <template #default="{ row }">{{ formatTime(row.sentAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right" align="center">
            <template #default="{ row }">
              <div class="admin-table-actions">
                <el-button link type="primary" :icon="View" @click="openWebhookDetail(row)">详情</el-button>
                <el-button v-if="canSend && row.status !== 'sending'" link type="primary" :icon="RefreshRight" @click="retryWebhook(row)">重发</el-button>
                <el-button v-if="canDelete && row.status !== 'sending'" link type="danger" :icon="Delete" @click="deleteWebhook(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <div class="admin-pagination">
          <el-pagination
            v-model:current-page="webhookPage"
            v-model:page-size="webhookPageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="webhookTotal"
            layout="total, sizes, prev, pager, next"
            @current-change="loadWebhookRecords"
            @size-change="handleWebhookPageSizeChange"
          />
        </div>
      </template>

      <WorkflowDeliveryRecords
        v-else-if="activeRecordView === 'workflow' && canViewWorkflowNotifications"
        :can-retry="canRetryWorkflowNotifications"
        :can-delete="canDeleteWorkflowNotifications"
      />
    </el-card>

    <el-dialog
      v-model="detailDialogVisible"
      title="通知记录详情"
      width="min(720px, 92vw)"
      append-to-body
      destroy-on-close
    >
      <el-descriptions v-if="detailRecord" :column="2" border class="record-detail">
        <el-descriptions-item label="记录 ID">{{ detailRecord.id }}</el-descriptions-item>
        <el-descriptions-item label="发送时间">{{ formatTime(detailRecord.addTime) }}</el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ detailRecord.title }}</el-descriptions-item>
        <el-descriptions-item label="接收人">{{ recipientLabel(detailRecord) }}</el-descriptions-item>
        <el-descriptions-item label="接收人用户 ID">{{ detailRecord.recipientUserId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="消息类型">{{ notificationTypeLabel(detailRecord.type) }}</el-descriptions-item>
        <el-descriptions-item label="来源">{{ sourceLabel(detailRecord.sourceType) }}</el-descriptions-item>
        <el-descriptions-item label="来源标识" :span="2">
          <span class="record-detail__source-id">{{ detailRecord.sourceId || '-' }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="阅读状态" :span="2">
          <el-tag :type="detailRecord.isRead ? 'info' : 'danger'" size="small">
            {{ detailRecord.isRead ? '已读' : '未读' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="通知内容" :span="2">
          <div class="record-detail__content">{{ detailRecord.content }}</div>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button
          v-if="canDelete && detailRecord"
          type="danger"
          plain
          :icon="Delete"
          :loading="deletingID === detailRecord.id"
          @click="deleteRecord(detailRecord)"
        >删除记录</el-button>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="webhookDetailVisible" title="Webhook 投递详情" width="min(720px, 92vw)" append-to-body destroy-on-close>
      <el-descriptions v-if="webhookDetail" :column="2" border class="record-detail">
        <el-descriptions-item label="记录 ID">{{ webhookDetail.id }}</el-descriptions-item>
        <el-descriptions-item label="状态"><el-tag :type="webhookStatusType(webhookDetail.status)" size="small">{{ webhookStatusLabel(webhookDetail.status) }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ webhookDetail.title || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Webhook 类型">{{ webhookDetail.webhookType || '-' }}</el-descriptions-item>
        <el-descriptions-item label="目标">{{ webhookDetail.target || '-' }}</el-descriptions-item>
        <el-descriptions-item label="来源类型">{{ sourceLabel(webhookDetail.sourceType) }}</el-descriptions-item>
        <el-descriptions-item label="来源标识">{{ webhookDetail.sourceId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="重试次数">{{ webhookDetail.attempts }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(webhookDetail.addTime) }}</el-descriptions-item>
        <el-descriptions-item label="发送时间">{{ formatTime(webhookDetail.sentAt) }}</el-descriptions-item>
        <el-descriptions-item label="下次重试">{{ formatTime(webhookDetail.nextRetryAt) }}</el-descriptions-item>
        <el-descriptions-item label="失败原因" :span="2"><span class="error-text">{{ webhookDetail.lastError || '-' }}</span></el-descriptions-item>
        <el-descriptions-item label="通知正文" :span="2"><div class="record-detail__content">{{ webhookDetail.content || '-' }}</div></el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button v-if="webhookDetail && canSend && webhookDetail.status !== 'sending'" type="primary" plain :icon="RefreshRight" @click="retryWebhook(webhookDetail)">重发</el-button>
        <el-button @click="webhookDetailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

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

    <NotificationStyleDialog
      v-model="styleDialogVisible"
      :can-edit="canStyleEdit"
      :can-send="canSend"
      :can-send-ding-talk="canSendDingTalk"
      @sent="handleStyleTestSent"
    />
  </div>
</template>

<script setup lang="ts">
import { Bell, Brush, ChatDotRound, Delete, Promotion, Refresh, RefreshRight, Search, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, onMounted, reactive, ref } from 'vue'
import { adminApi } from '../../api'
import type {
  InAppNotificationItem,
  InAppNotificationRecipientOptions,
  InAppNotificationScope,
  WebhookNotificationItem,
} from '../../api/types'
import { hasPerm } from '../../utils/permission'
import WorkflowUserTreePicker from '../workflow/components/WorkflowUserTreePicker.vue'
import NotificationStyleDialog from './components/NotificationStyleDialog.vue'
import WorkflowDeliveryRecords from './components/WorkflowDeliveryRecords.vue'

const list = ref<InAppNotificationItem[]>([])
const inAppSelectedRows = ref<InAppNotificationItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const addTimeRange = ref<[number, number] | null>(null)
const sendDialogVisible = ref(false)
const styleDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const detailRecord = ref<InAppNotificationItem | null>(null)
const webhookDetailVisible = ref(false)
const webhookDetail = ref<WebhookNotificationItem | null>(null)
const deletingID = ref<number | null>(null)
const sending = ref(false)
const sendRequestID = ref('')
const sendChannel = ref<'in_app' | 'dingtalk'>('in_app')
const activeRecordView = ref<'in_app' | 'webhook' | 'workflow'>('in_app')
const webhookList = ref<WebhookNotificationItem[]>([])
const webhookTotal = ref(0)
const webhookPage = ref(1)
const webhookPageSize = ref(20)
const webhookLoading = ref(false)
const webhookStatus = ref('')
const webhookSelectedRows = ref<WebhookNotificationItem[]>([])
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
const filters = reactive({
  title: '',
  recipientName: '',
  sourceType: '',
  notificationType: '',
  isRead: '' as '' | 0 | 1,
})
const scopeOptions = [
  { label: '全部用户', value: 'all' },
  { label: '指定部门', value: 'departments' },
  { label: '指定用户', value: 'users' },
]
const sourceOptions = [
  { label: '后台发送', value: 'admin_manual' },
  { label: '流程通知', value: 'workflow_instance' },
  { label: '流程通知（历史）', value: 'workflow' },
  { label: '定时任务', value: 'scheduled_task_run' },
  { label: '问卷通知', value: 'survey' },
]
const notificationTypeOptions = [
  { label: '待处理', value: 'task_arrived' },
  { label: '处理提醒', value: 'task_reminder' },
  { label: '审批通过', value: 'approval_result_approved' },
  { label: '审批驳回', value: 'approval_result_rejected' },
  { label: '审批退回', value: 'approval_result_returned' },
  { label: '流程抄送', value: 'node_cc' },
  { label: '流程通知', value: 'node_notify' },
  { label: '流程评论', value: 'instance_commented' },
  { label: '表单修改', value: 'instance_form_revised' },
  { label: '流程消息', value: 'workflow' },
  { label: '系统通知', value: 'admin_manual' },
  { label: '定时通知', value: 'scheduled_task' },
  { label: '问卷统计', value: 'survey_stat' },
]
const canSend = computed(() => hasPerm('admin:menu:notification:send'))
const canSendDingTalk = computed(() => hasPerm('admin:menu:notification:dingtalk-send'))
const canStyleList = computed(() => hasPerm('admin:menu:notification:style:list'))
const canStyleEdit = computed(() => hasPerm('admin:menu:notification:style:edit'))
const canDelete = computed(() => hasPerm('admin:menu:notification:delete'))
const canViewWorkflowNotifications = computed(() => hasPerm('admin:menu:workflow:notification:list'))
const canRetryWorkflowNotifications = computed(() => hasPerm('admin:menu:workflow:notification:retry'))
const canDeleteWorkflowNotifications = computed(() => hasPerm('admin:menu:workflow:notification:delete'))
const recordSummary = computed(() => activeRecordView.value === 'workflow'
  ? '查看流程站内信与钉钉的完整投递状态'
  : activeRecordView.value === 'webhook'
    ? `共 ${webhookTotal.value} 条 Webhook 投递记录`
    : `共 ${total.value} 条站内信投递记录`)
const sendDialogTitle = computed(() => sendChannel.value === 'dingtalk' ? '发送钉钉通知' : '发送站内信')
const departmentModelValue = computed<number[]>({
  get: () => sendForm.departmentIds,
  set: value => { sendForm.departmentIds = normalizeIDs(value) },
})

async function load() {
  loading.value = true
  try {
    const response = await adminApi.inAppNotificationList({
      page: page.value,
      pageSize: pageSize.value,
      title: filters.title.trim() || undefined,
      recipientName: filters.recipientName.trim() || undefined,
      sourceType: filters.sourceType || undefined,
      type: filters.notificationType || undefined,
      isRead: filters.isRead === '' ? undefined : filters.isRead,
      addTimeFrom: addTimeRange.value ? Number(addTimeRange.value[0]) : undefined,
      addTimeTo: addTimeRange.value ? Number(addTimeRange.value[1]) : undefined,
    })
    list.value = Array.isArray(response.data?.list) ? response.data.list : []
    total.value = Number(response.data?.total || 0)
  } finally {
    loading.value = false
  }
}

function handleInAppSelectionChange(rows: InAppNotificationItem[]) {
  inAppSelectedRows.value = rows
}

async function deleteSelectedInApp() {
  const rows = inAppSelectedRows.value
  if (!rows.length) return
  await ElMessageBox.confirm(`将删除选中的 ${rows.length} 条站内信记录。已发送的站内信不会被撤回。`, '批量删除站内信记录', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
  const results = await Promise.allSettled(rows.map(row => adminApi.inAppNotificationDelete(row.id)))
  const failed = results.filter(result => result.status === 'rejected').length
  inAppSelectedRows.value = []
  await load()
  if (failed > 0) ElMessage.warning(`已处理 ${rows.length - failed} 条，${failed} 条删除失败`)
  else ElMessage.success(`已删除 ${rows.length} 条站内信记录`)
}

async function loadWebhookRecords() {
  webhookLoading.value = true
  try {
    const response = await adminApi.webhookNotificationList({
      page: webhookPage.value,
      pageSize: webhookPageSize.value,
      status: webhookStatus.value || undefined,
    })
    webhookList.value = Array.isArray(response.data?.list) ? response.data.list : []
    webhookTotal.value = Number(response.data?.total || 0)
  } finally {
    webhookLoading.value = false
  }
}

function selectableWebhook(row: WebhookNotificationItem) {
  return row.status !== 'sending'
}

function handleWebhookSelectionChange(rows: WebhookNotificationItem[]) {
  webhookSelectedRows.value = rows.filter(row => row.status !== 'sending')
}

async function deleteSelectedWebhook() {
  const rows = webhookSelectedRows.value
  if (!rows.length) return
  await ElMessageBox.confirm(`将删除选中的 ${rows.length} 条 Webhook 投递记录，不会撤回已经发送的请求。`, '批量删除 Webhook 记录', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
  const results = await Promise.allSettled(rows.map(row => adminApi.webhookNotificationDelete(row.id)))
  const failed = results.filter(result => result.status === 'rejected').length
  webhookSelectedRows.value = []
  await loadWebhookRecords()
  if (failed > 0) ElMessage.warning(`已处理 ${rows.length - failed} 条，${failed} 条删除失败`)
  else ElMessage.success(`已删除 ${rows.length} 条 Webhook 记录`)
}

async function openWebhookDetail(row: WebhookNotificationItem) {
  webhookDetailVisible.value = true
  webhookDetail.value = null
  const response = await adminApi.webhookNotificationDetail(row.id)
  webhookDetail.value = response.data || row
}

async function retryWebhook(row: WebhookNotificationItem) {
  await ElMessageBox.confirm('将创建一条新的 Webhook 投递记录，原记录会保留。', '重发 Webhook', { type: 'warning', confirmButtonText: '重发', cancelButtonText: '取消' })
  await adminApi.webhookNotificationRetry(row.id)
  ElMessage.success('Webhook 重发记录已创建')
  webhookDetailVisible.value = false
  await loadWebhookRecords()
}

async function deleteWebhook(row: WebhookNotificationItem) {
  await ElMessageBox.confirm('删除后只移除本地投递记录，不会撤回已经发送的 Webhook。', '删除 Webhook 记录', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
  await adminApi.webhookNotificationDelete(row.id)
  ElMessage.success('Webhook 记录已删除')
  if (webhookList.value.length === 1 && webhookPage.value > 1) webhookPage.value -= 1
  await loadWebhookRecords()
}

function handleRecordViewChange(view: string | number) {
  if (view === 'webhook') void loadWebhookRecords()
}

function handleWebhookPageSizeChange() {
  webhookPage.value = 1
  void loadWebhookRecords()
}

function openDetail(row: InAppNotificationItem) {
  detailRecord.value = row
  detailDialogVisible.value = true
}

async function deleteRecord(row: InAppNotificationItem) {
  try {
    await ElMessageBox.confirm(
      `确定删除通知记录“${row.title}”吗？该操作仅从后台记录列表移除，不会撤回接收人已收到的站内信。`,
      '删除通知记录',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  deletingID.value = row.id
  try {
    await adminApi.inAppNotificationDelete(row.id)
    if (detailRecord.value?.id === row.id) {
      detailDialogVisible.value = false
      detailRecord.value = null
    }
    if (list.value.length === 1 && page.value > 1) page.value -= 1
    await load()
    ElMessage.success('通知记录已删除')
  } finally {
    deletingID.value = null
  }
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
    if (sendChannel.value === 'in_app') {
      page.value = 1
      await load()
    }
  } finally {
    sending.value = false
  }
}

async function handleStyleTestSent(channel: 'in_app' | 'dingtalk') {
  if (channel === 'in_app') {
    page.value = 1
    await load()
  }
}

function search() {
  page.value = 1
  load()
}

function resetFilters() {
  filters.title = ''
  filters.recipientName = ''
  filters.sourceType = ''
  filters.notificationType = ''
  filters.isRead = ''
  addTimeRange.value = null
  search()
}

function handlePageSizeChange() {
  page.value = 1
  load()
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
  if (sourceType === 'workflow' || sourceType === 'workflow_instance') return '流程通知'
  if (sourceType === 'survey') return '问卷通知'
  return sourceType || '-'
}

function notificationTypeLabel(type: string) {
  return notificationTypeOptions.find(option => option.value === type)?.label || type || '-'
}

function webhookStatusLabel(status: string) {
  return ({ pending: '待发送', sending: '发送中', sent: '已发送', failed: '待重试', dead: '已终止' } as Record<string, string>)[status] || status || '-'
}

function webhookStatusType(status: string) {
  return ({ pending: 'info', sending: 'warning', sent: 'success', failed: 'warning', dead: 'danger' } as Record<string, 'info' | 'warning' | 'success' | 'danger'>)[status] || 'info'
}

function recipientLabel(row: InAppNotificationItem) {
  if (row.recipientName) return row.recipientName
  return row.recipientUserId ? `用户 #${row.recipientUserId}` : '-'
}

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

onMounted(load)
</script>

<style scoped>
.notification-page { min-height: 100%; }
.page-header { display: flex; align-items: center; justify-content: space-between; gap: 20px; padding: 0 0 18px; border-bottom: 1px solid var(--admin-border); }
.page-title { display: flex; min-width: 0; align-items: center; gap: 12px; }
.page-title > .el-icon { width: 36px; height: 36px; flex: 0 0 36px; border-radius: 6px; background: #e9f8f4; color: #07866f; font-size: 18px; }
.page-title h2 { margin: 0; color: var(--admin-text); font-size: 18px; line-height: 26px; }
.page-title span { color: var(--admin-muted); font-size: 12px; }
.page-actions { display: flex; flex: 0 0 auto; gap: 8px; }
.record-view-tabs { margin-top: 14px; }
.record-view-tabs :deep(.el-tabs__header) { margin-bottom: 16px; }
.record-filters { padding-bottom: 16px; border-bottom: 1px solid var(--admin-border); }
.notification-table { border: 1px solid var(--admin-border); border-radius: 6px; }
.notification-content { width: 100%; overflow: hidden; color: var(--admin-text-secondary); line-height: 20px; text-overflow: ellipsis; white-space: nowrap; }
.record-detail__source-id { overflow-wrap: anywhere; }
.record-detail__content { min-height: 72px; white-space: pre-wrap; overflow-wrap: anywhere; line-height: 1.7; }
@media (max-width: 720px) {
  .page-header { align-items: flex-start; }
  .page-actions { flex-wrap: wrap; justify-content: flex-end; }
}
</style>
