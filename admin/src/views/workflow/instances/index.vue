<template>
  <div class="admin-page workflow-instance-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar workflow-runtime-filters">
        <div class="admin-toolbar__left">
          <el-input v-model="filters.businessType" placeholder="业务类型" clearable style="width: 180px" @keyup.enter="search" />
          <el-input v-model="filters.businessKey" placeholder="业务标识" clearable style="width: 220px" @keyup.enter="search" />
          <el-select v-model="filters.status" placeholder="全部状态" clearable style="width: 140px" @change="search">
            <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
          <el-input v-model="filters.starterId" placeholder="发起人 ID" clearable style="width: 150px" @keyup.enter="search" />
          <el-button type="primary" @click="search">搜索</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
      </div>

      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-button v-if="canStart" type="primary" icon="Plus" @click="openStartDialog">发起流程</el-button>
          <el-button
            v-if="canDelete"
            type="danger"
            plain
            icon="Delete"
            :disabled="selectedInstances.length === 0"
            :loading="batchDeleting"
            @click="deleteSelectedInstances"
          >批量删除</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" :loading="loading" @click="loadList" />
        </div>
      </div>

      <el-table v-loading="loading" :data="list" row-key="id" stripe @selection-change="handleSelectionChange">
        <el-table-column v-if="canDelete" type="selection" width="48" :selectable="canSelectInstance" />
        <el-table-column prop="id" label="实例 ID" min-width="190" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="mono-text">{{ row.id }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="definitionName" label="流程名称" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.definitionName || row.definitionKey || `#${row.definitionId}` }}
          </template>
        </el-table-column>
        <el-table-column prop="definitionKey" label="流程编码" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="mono-text">{{ row.definitionKey || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="版本号" width="90" align="center">
          <template #default="{ row }">v{{ row.definitionVersion }}</template>
        </el-table-column>
        <el-table-column prop="businessType" label="业务类型" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.businessType || '-' }}</template>
        </el-table-column>
        <el-table-column prop="businessKey" label="业务标识" min-width="210" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="mono-text">{{ row.businessKey || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="发起人" width="120">
          <template #default="{ row }">{{ userDisplay(row.starterName, row.starterId) }}</template>
        </el-table-column>
        <el-table-column label="操作人" width="120">
          <template #default="{ row }">{{ userDisplay(row.operatorName, row.operatorId) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="workflowInstanceStatusMeta(row.status).type" size="small">{{ workflowInstanceStatusMeta(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="170">
          <template #default="{ row }">{{ formatTime(row.startTime) }}</template>
        </el-table-column>
        <el-table-column label="结束时间" width="170">
          <template #default="{ row }">{{ formatTime(row.endTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canViewDetail" link type="primary" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="canDelete"
              link
              type="danger"
              :disabled="!canSelectInstance(row)"
              :loading="deletingInstanceId === row.id"
              :title="canSelectInstance(row) ? '删除实例' : '审批中的实例不能删除'"
              @click="deleteInstance(row)"
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
          layout="total,sizes,prev,pager,next"
          @current-change="loadList"
          @size-change="handlePageSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="startDialog" title="发起通用流程" width="720px" destroy-on-close>
      <el-form label-width="108px" @submit.prevent>
        <el-form-item label="流程定义" required>
          <el-select
            v-model="startForm.definitionId"
            filterable
            placeholder="选择已发布流程"
            style="width: 100%"
            :loading="definitionsLoading"
            @change="handleStartDefinitionChange"
          >
            <el-option
              v-for="definition in publishedDefinitions"
              :key="definition.id"
              :label="`${definition.name} (${definition.key}) · v${definition.version} · ${availabilityStatusLabel(definition.availabilityStatus)}`"
              :value="definition.id"
              :disabled="!canStartDefinition(definition)"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="实例创建方式" required>
          <el-radio-group v-model="startMode" @change="handleStartModeChange">
            <el-radio-button value="single">代一名用户发起</el-radio-button>
            <el-radio-button value="per_user">为多名用户分别发起</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="业务发起人" required>
          <WorkflowUserTreePicker
            :model-value="starterUserIds"
            :departments="initiatorDepartments"
            :users="eligibleInitiatorUsers"
            :multiple="startMode === 'per_user'"
            :disabled="initiatorUsersLoading"
            :placeholder="startMode === 'per_user' ? '请选择一个或多个业务发起人' : '请选择一名业务发起人'"
            @update:model-value="starterUserIds = $event"
          />
          <div v-if="startMode === 'per_user'" class="form-help">将分别以每位所选用户为业务发起人创建独立流程实例。</div>
        </el-form-item>
        <el-form-item label="业务类型" required>
          <el-input v-model="startForm.businessType" maxlength="80" placeholder="例如：purchase_order" />
        </el-form-item>
        <el-form-item label="业务标识" required>
          <el-input v-model="startForm.businessKey" maxlength="120" placeholder="例如：PO-20260829-001" />
        </el-form-item>
        <el-form-item label="流程变量">
          <el-input
            v-model="startForm.variablesText"
            type="textarea"
            :rows="8"
            placeholder="可选，填写 JSON 对象，例如：{&quot;amount&quot;: 1200, &quot;managerId&quot;: &quot;18&quot;}"
          />
          <div class="form-help">变量用于审批人解析和网关条件判断，不承载具体业务表单。</div>
        </el-form-item>
      </el-form>
      <section v-if="selectedStartDefinition" class="dialog-section">
        <h3>流程表单</h3>
        <WorkflowRuntimeForm
          ref="startRuntimeForm"
          v-model="startFormData"
          :fields="selectedStartDefinition.form || []"
          :field-access="startFieldAccess"
          :field-actions="startFieldActions"
          empty-text="该流程未配置表单"
        />
      </section>
      <template #footer>
        <el-button @click="startDialog = false">取消</el-button>
        <el-button type="primary" :loading="starting" @click="startInstance">确认发起</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialog" title="流程实例详情" width="920px" destroy-on-close>
      <div v-loading="detailLoading" class="instance-detail">
        <template v-if="detail">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="实例 ID" :span="2"><span class="mono-text">{{ detail.instance.id }}</span></el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="workflowInstanceStatusMeta(detail.instance.status).type" size="small">{{ workflowInstanceStatusMeta(detail.instance.status).label }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="流程定义">{{ detail.instance.definitionKey }} · v{{ detail.instance.definitionVersion }}</el-descriptions-item>
            <el-descriptions-item label="业务类型">{{ detail.instance.businessType }}</el-descriptions-item>
            <el-descriptions-item label="业务标识">{{ detail.instance.businessKey }}</el-descriptions-item>
            <el-descriptions-item label="发起人">{{ userDisplay(detail.instance.starterName, detail.instance.starterId) }}</el-descriptions-item>
            <el-descriptions-item label="操作人">{{ userDisplay(detail.instance.operatorName, detail.instance.operatorId) }}</el-descriptions-item>
            <el-descriptions-item label="开始时间">{{ formatTime(detail.instance.startTime) }}</el-descriptions-item>
            <el-descriptions-item label="结束时间">{{ formatTime(detail.instance.endTime) }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="canStart && detail.instance.status === 'running'" class="detail-actions">
            <el-button type="primary" plain :loading="resumingTimers" @click="resumeTimers">推进到期节点</el-button>
          </div>

          <section class="detail-section">
            <el-collapse v-model="detailExpandedSections.form" class="detail-collapse">
              <el-collapse-item name="form">
                <template #title><span class="detail-collapse__title">流程表单</span></template>
                <WorkflowRuntimeForm
                  :model-value="detail.formData || {}"
                  :fields="detail.form || []"
                  :field-access="detailFieldAccess"
                  :user-name-map="detail.userNames || {}"
                  readonly
                  empty-text="该实例没有流程表单数据"
                />
              </el-collapse-item>
            </el-collapse>
          </section>

          <section class="detail-section">
            <el-collapse v-model="detailExpandedSections.variables" class="detail-collapse">
              <el-collapse-item name="variables">
                <template #title><span class="detail-collapse__title">流程变量</span></template>
                <pre class="json-viewer">{{ formatJSON(detail.variables) }}</pre>
              </el-collapse-item>
            </el-collapse>
          </section>

          <section class="detail-section">
            <h3>任务记录</h3>
            <el-table :data="detail.tasks" size="small" border>
              <el-table-column prop="nodeName" label="节点" min-width="140" />
              <el-table-column label="处理人" width="120">
                <template #default="{ row }">{{ taskUserDisplay(row) }}</template>
              </el-table-column>
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="workflowTaskStatusMeta(row.status).type" size="small">{{ workflowTaskStatusMeta(row.status).label }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="动作" width="90">
                <template #default="{ row }">{{ actionLabel(row.action) }}</template>
              </el-table-column>
              <el-table-column prop="comment" label="处理意见" min-width="180" show-overflow-tooltip />
              <el-table-column label="处理时间" width="170">
                <template #default="{ row }">{{ formatTime(row.handledAt) }}</template>
              </el-table-column>
            </el-table>
          </section>

          <section v-if="canViewNotifications" class="detail-section">
            <div class="detail-section__heading">
              <h3>通知投递</h3>
              <el-button
                v-if="canRetryNotifications"
                size="small"
                type="primary"
                plain
                :loading="dispatchingDueNotifications"
                @click="dispatchDueNotifications"
              >投递到期通知</el-button>
            </div>
            <el-table v-loading="notificationsLoading" :data="notifications" size="small" border empty-text="该实例暂无通知投递记录">
              <el-table-column label="接收人" width="110">
                <template #default="{ row }">{{ userDisplay(row.recipientUserName, row.recipientUserId) }}</template>
              </el-table-column>
              <el-table-column label="事件" width="110">
                <template #default="{ row }">{{ notificationKindLabel(row.kind) }}</template>
              </el-table-column>
              <el-table-column label="渠道" width="105">
                <template #default="{ row }">{{ notificationChannelLabel(row.channel) }}</template>
              </el-table-column>
              <el-table-column label="状态" width="95">
                <template #default="{ row }">
                  <el-tag :type="notificationStatusMeta(row.status).type" size="small">{{ notificationStatusMeta(row.status).label }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="attempts" label="尝试" width="70" align="center" />
              <el-table-column prop="lastError" label="最近错误" min-width="180" show-overflow-tooltip>
                <template #default="{ row }">{{ row.lastError || '-' }}</template>
              </el-table-column>
              <el-table-column label="发送时间" width="170">
                <template #default="{ row }">{{ formatTime(row.sentAt) }}</template>
              </el-table-column>
              <el-table-column v-if="canRetryNotifications" label="操作" width="80" fixed="right">
                <template #default="{ row }">
                  <el-button
                    v-if="row.status === 'failed' || row.status === 'dead'"
                    link
                    type="primary"
                    :loading="retryingNotificationId === row.id"
                    @click="retryNotification(row)"
                  >重发</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>

          <section class="detail-section">
            <el-collapse v-model="detailExpandedSections.history" class="detail-collapse">
              <el-collapse-item name="history">
                <template #title><span class="detail-collapse__title">流转历史</span></template>
                <el-timeline class="history-timeline">
                  <el-timeline-item v-for="event in detail.history" :key="event.id" :timestamp="formatTime(event.eventTime)">
                    <strong>{{ historyEventLabel(event.eventType) }}</strong>
                    <span v-if="event.actorId"> · 操作人 {{ userDisplay(event.actorName, event.actorId) }}</span>
                    <p v-if="event.message">{{ event.message }}</p>
                  </el-timeline-item>
                </el-timeline>
              </el-collapse-item>
            </el-collapse>
          </section>
        </template>
      </div>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../../api'
import type { AdminUser } from '../../../api/types'
import { hasPerm } from '../../../utils/permission'
import WorkflowRuntimeForm from '../components/WorkflowRuntimeForm.vue'
import WorkflowUserTreePicker from '../components/WorkflowUserTreePicker.vue'
import { initialWorkflowFormData, workflowFieldActionMap, workflowFieldAccessMap, writableWorkflowFormData } from '../runtimeForm'
import type {
  WorkflowInstanceDetail,
  WorkflowInstanceSummary,
  WorkflowNotificationChannel,
  WorkflowNotificationKind,
  WorkflowNotificationRecord,
  WorkflowNotificationStatus,
  WorkflowPublishedDefinition,
  WorkflowTaskAction,
  WorkflowTaskSummary,
} from '../types'
import { workflowInstanceStatusMeta, workflowTaskStatusMeta } from '../workflowStatus'

const statusOptions = [
  { label: '审批中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '已驳回', value: 'rejected' },
  { label: '已取消', value: 'cancelled' },
]

const canStart = computed(() => hasPerm('admin:menu:workflow:instance:start'))
const canViewDetail = computed(() => hasPerm('admin:menu:workflow:instance:detail'))
const canDelete = computed(() => hasPerm('admin:menu:workflow:instance:delete'))
const canViewNotifications = computed(() => hasPerm('admin:menu:workflow:notification:list'))
const canRetryNotifications = computed(() => hasPerm('admin:menu:workflow:notification:retry'))
const loading = ref(false)
const list = ref<WorkflowInstanceSummary[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ businessType: '', businessKey: '', status: '', starterId: '' })
const selectedInstances = ref<WorkflowInstanceSummary[]>([])
const deletingInstanceId = ref('')
const batchDeleting = ref(false)

const startDialog = ref(false)
const starting = ref(false)
const definitionsLoading = ref(false)
const publishedDefinitions = ref<WorkflowPublishedDefinition[]>([])
const startFormData = ref<Record<string, unknown>>({})
const startRuntimeForm = ref<{ validate: () => boolean; resetValidation: () => void } | null>(null)
const startForm = reactive<{
  definitionId: number | undefined
  businessType: string
  businessKey: string
  variablesText: string
}>({
  definitionId: undefined,
  businessType: '',
  businessKey: '',
  variablesText: '',
})
const startMode = ref<'single' | 'per_user'>('single')
const starterUserIds = ref<number[]>([])
const initiatorUsers = ref<AdminUser[]>([])
const initiatorDepartments = ref<any[]>([])
const initiatorUsersLoading = ref(false)

const detailDialog = ref(false)
const detailLoading = ref(false)
const resumingTimers = ref(false)
const detail = ref<WorkflowInstanceDetail | null>(null)
const notifications = ref<WorkflowNotificationRecord[]>([])
const notificationsLoading = ref(false)
const dispatchingDueNotifications = ref(false)
const retryingNotificationId = ref('')
const detailExpandedSections = reactive<{ form: string[]; variables: string[]; history: string[] }>({
  form: [],
  variables: [],
  history: [],
})

const notificationStatusOptions = [
  { value: 'pending' as const, label: '待投递', type: 'info' as const },
  { value: 'sending' as const, label: '投递中', type: 'warning' as const },
  { value: 'sent' as const, label: '已发送', type: 'success' as const },
  { value: 'failed' as const, label: '待重试', type: 'danger' as const },
  { value: 'dead' as const, label: '已停止', type: 'danger' as const },
]

const selectedStartDefinition = computed(() => (
  publishedDefinitions.value.find((definition) => definition.id === startForm.definitionId) || null
))

const eligibleInitiatorUsers = computed(() => {
  const initiator = selectedStartDefinition.value?.initiator
  const excludedUserIds = new Set((initiator?.excludedUserIds || []).map(Number))
  const allowedUserIds = new Set((initiator?.userIds || []).map(Number))
  const allowedDepartmentIds = new Set((initiator?.departmentIds || []).map(Number))
  return initiatorUsers.value.filter((user) => {
    if (excludedUserIds.has(Number(user.id))) return false
    if (initiator?.scope !== 'specified') return true
    return allowedUserIds.has(Number(user.id))
      || (user.deptIds || []).some(departmentId => allowedDepartmentIds.has(Number(departmentId)))
  })
})

const startFieldAccess = computed(() => {
  const definition = selectedStartDefinition.value
  if (!definition) return {}
  return workflowFieldAccessMap(
    definition.form || [],
    definition.fieldPermissions || {},
    definition.startNodeId || 'start',
    'write',
  )
})

const startFieldActions = computed(() => {
  const definition = selectedStartDefinition.value
  if (!definition) return {}
  return workflowFieldActionMap(
    definition.form || [],
    definition.fieldPermissions || {},
    definition.startNodeId || 'start',
    { add: true, delete: true },
  )
})

const detailFieldAccess = computed(() => {
  const fields = detail.value?.form || []
  return Object.fromEntries(fields.map((field) => [field.key, 'read' as const]))
})

function availabilityStatusLabel(status: WorkflowPublishedDefinition['availabilityStatus']) {
  const values = {
    available: '可发起',
    not_started: '尚未开放',
    expired: '已结束',
    outside_window: '当前未开放',
  }
  return values[status] || '当前未开放'
}

function canStartDefinition(definition: WorkflowPublishedDefinition) {
  return definition.availabilityStatus === 'available'
}

function actionLabel(action: WorkflowTaskAction | '') {
  if (action === 'approve') return '通过'
  if (action === 'reject') return '驳回'
  if (action === 'return') return '退回'
  if (action === 'submit') return '提交'
  return '-'
}

function userDisplay(name: string | undefined, id: string | undefined) {
  return name?.trim() || id?.trim() || '-'
}

function taskUserDisplay(task: WorkflowTaskSummary) {
  if (task.handledBy || task.handledByName) return userDisplay(task.handledByName, task.handledBy)
  return userDisplay(task.assigneeName, task.assigneeId)
}

function canSelectInstance(instance: WorkflowInstanceSummary) {
  return instance.status !== 'running'
}

function handleSelectionChange(rows: WorkflowInstanceSummary[]) {
  selectedInstances.value = rows
}

function historyEventLabel(type: string) {
  return {
    instance_started: '流程已发起', task_created: '任务已创建', task_activated: '任务已激活',
    task_approved: '任务已通过', task_rejected: '任务已驳回', task_returned: '任务已退回', task_submitted: '任务已提交', task_cancelled: '任务已取消',
    node_cc: '已记录抄送', node_notify: '通知节点已触发', node_automated: '自动动作已执行', timer_waiting: '定时节点等待中', timer_resumed: '定时节点已推进',
    instance_completed: '流程已完成', instance_rejected: '流程已驳回',
  }[type] || type
}

function notificationStatusMeta(status: WorkflowNotificationStatus) {
  return notificationStatusOptions.find(option => option.value === status)
    || { value: status, label: status || '未知', type: 'info' as const }
}

function notificationKindLabel(kind: WorkflowNotificationKind) {
  return {
    node_cc: '节点抄送',
    node_notify: '节点通知',
    task_arrived: '任务到达',
    task_reminder: '处理提醒',
    instance_commented: '流程评论',
    approval_result_approved: '审批通过结果',
    approval_result_rejected: '审批驳回结果',
  }[kind] || kind
}

function notificationChannelLabel(channel: WorkflowNotificationChannel) {
  return { in_app: '站内通知', dingtalk_oa: '钉钉 OA' }[channel] || channel
}

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function formatJSON(value: unknown) {
  return JSON.stringify(value || {}, null, 2)
}

function parseVariables(text: string) {
  if (!text.trim()) return {}
  const value = JSON.parse(text)
  if (!value || Array.isArray(value) || typeof value !== 'object') throw new Error('流程变量必须是 JSON 对象')
  return value as Record<string, unknown>
}

async function loadList() {
  loading.value = true
  try {
    const response = await adminApi.workflowInstanceList({
      page: page.value, pageSize: pageSize.value,
      businessType: filters.businessType.trim(), businessKey: filters.businessKey.trim(),
      status: filters.status, starterId: filters.starterId.trim(),
    })
    list.value = Array.isArray(response.data?.list) ? response.data.list : []
    total.value = Number(response.data?.total || 0)
    selectedInstances.value = []
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  loadList()
}

function resetFilters() {
  Object.assign(filters, { businessType: '', businessKey: '', status: '', starterId: '' })
  search()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  page.value = 1
  loadList()
}

async function loadPublishedDefinitions() {
  definitionsLoading.value = true
  try {
    const response = await adminApi.workflowPublishedDefinitionList({ page: 1, pageSize: 100 })
    publishedDefinitions.value = Array.isArray(response.data) ? response.data as WorkflowPublishedDefinition[] : []
  } finally {
    definitionsLoading.value = false
  }
}

async function loadInitiatorOptions() {
  if (initiatorUsers.value.length > 0 && initiatorDepartments.value.length > 0) return
  initiatorUsersLoading.value = true
  try {
    const [usersResponse, departmentsResponse] = await Promise.all([
      adminApi.workflowUserOptions({ page: 1, pageSize: 9999 }),
      adminApi.workflowDepartmentOptions(),
    ])
    initiatorUsers.value = Array.isArray(usersResponse.data?.list) ? usersResponse.data.list : []
    initiatorDepartments.value = Array.isArray(departmentsResponse.data) ? departmentsResponse.data : []
  } finally {
    initiatorUsersLoading.value = false
  }
}

function handleStartModeChange(mode: string | number | boolean | undefined) {
  if (mode === 'single') starterUserIds.value = starterUserIds.value.slice(0, 1)
}

function buildTargetBusinessKey(base: string, targetUserId: number) {
  const suffix = `:user:${targetUserId}`
  const normalizedBase = base.trim().slice(0, Math.max(1, 120 - suffix.length))
  return `${normalizedBase}${suffix}`
}

async function openStartDialog() {
  Object.assign(startForm, { definitionId: undefined, businessType: '', businessKey: '', variablesText: '' })
  startMode.value = 'single'
  starterUserIds.value = []
  startFormData.value = {}
  startDialog.value = true
  await Promise.all([loadPublishedDefinitions(), loadInitiatorOptions()])
}

async function handleStartDefinitionChange(definitionID?: number) {
  if (!definitionID) {
    startFormData.value = {}
    return
  }
  const response = await adminApi.workflowPublishedDefinitionDetail(definitionID)
  const definition = response.data as WorkflowPublishedDefinition
  const index = publishedDefinitions.value.findIndex((item) => item.id === definition.id)
  if (index >= 0) {
    publishedDefinitions.value.splice(index, 1, definition)
  } else {
    publishedDefinitions.value.push(definition)
  }
  startFormData.value = initialWorkflowFormData(definition.form || [])
  startRuntimeForm.value?.resetValidation()
  starterUserIds.value = starterUserIds.value.filter(id => eligibleInitiatorUsers.value.some(user => Number(user.id) === Number(id)))
}

async function startInstance() {
  const definitionID = startForm.definitionId
  if (!definitionID || !startForm.businessType.trim() || !startForm.businessKey.trim()) {
    ElMessage.warning('请选择流程定义并填写业务类型、业务标识')
    return
  }
  if (starterUserIds.value.length === 0) {
    ElMessage.warning('请选择业务发起人')
    return
  }
  if (selectedStartDefinition.value && !canStartDefinition(selectedStartDefinition.value)) {
    ElMessage.warning('当前不在流程允许发起时间内')
    return
  }
  if (selectedStartDefinition.value && startRuntimeForm.value && !startRuntimeForm.value.validate()) {
    ElMessage.warning('请检查流程表单中的校验提示')
    return
  }
  let variables: Record<string, unknown>
  try {
    variables = parseVariables(startForm.variablesText)
  } catch (error) {
    ElMessage.warning(error instanceof Error ? error.message : '流程变量 JSON 格式无效')
    return
  }
  starting.value = true
  try {
    const basePayload = {
      definitionId: definitionID,
      definitionVersion: 0,
      businessType: startForm.businessType.trim(),
      businessKey: startForm.businessKey.trim(),
      variables,
      formData: writableWorkflowFormData(selectedStartDefinition.value?.form || [], startFormData.value, startFieldAccess.value),
    }
    if (startMode.value === 'single') {
      const response = await adminApi.workflowInstanceStart({ ...basePayload, starterId: String(starterUserIds.value[0]) })
      startDialog.value = false
      ElMessage.success(`流程已发起：${response.data?.instanceId || ''}`)
    } else {
      const requestedUserIds = [...starterUserIds.value]
      const results = await Promise.allSettled(requestedUserIds.map((starterUserId) => adminApi.workflowInstanceStart({
        ...basePayload,
        starterId: String(starterUserId),
        businessKey: buildTargetBusinessKey(basePayload.businessKey, starterUserId),
        variables: { ...variables, targetUserId: String(starterUserId) },
        formData: { ...basePayload.formData },
      })))
      const failedUserIds = requestedUserIds.filter((_, index) => results[index]?.status === 'rejected')
      const successCount = requestedUserIds.length - failedUserIds.length
      if (failedUserIds.length === 0) {
        startDialog.value = false
        ElMessage.success(`已为 ${successCount} 位用户分别发起流程`)
      } else {
        starterUserIds.value = failedUserIds
        ElMessage.warning(`已成功发起 ${successCount} 个实例，${failedUserIds.length} 个失败；已保留失败用户供重试`)
      }
    }
    await loadList()
  } finally {
    starting.value = false
  }
}

async function openDetail(row: WorkflowInstanceSummary) {
  detail.value = null
  notifications.value = []
  detailDialog.value = true
  detailExpandedSections.form = []
  detailExpandedSections.variables = []
  detailExpandedSections.history = []
  detailLoading.value = true
  try {
    const [response] = await Promise.all([
      adminApi.workflowInstanceDetail(row.id),
      canViewNotifications.value ? loadNotifications(row.id) : Promise.resolve(),
    ])
    detail.value = response.data as WorkflowInstanceDetail
  } finally {
    detailLoading.value = false
  }
}

async function deleteInstance(instance: WorkflowInstanceSummary) {
  if (!canSelectInstance(instance)) {
    ElMessage.warning('审批中的流程实例不能删除，请先取消或等待流程结束')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确认删除流程实例「${instance.businessKey || instance.id}」？删除后将从管理列表隐藏，审计数据仍会保留。`,
      '删除流程实例',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  deletingInstanceId.value = instance.id
  try {
    await adminApi.workflowInstanceDelete(instance.id)
    ElMessage.success('流程实例已删除')
    if (list.value.length === 1 && page.value > 1) page.value -= 1
    await loadList()
  } finally {
    deletingInstanceId.value = ''
  }
}

async function deleteSelectedInstances() {
  const ids = selectedInstances.value.filter(canSelectInstance).map(instance => instance.id)
  if (ids.length === 0) return
  try {
    await ElMessageBox.confirm(
      `确认删除选中的 ${ids.length} 个流程实例？删除后将从管理列表隐藏，审计数据仍会保留。`,
      '批量删除流程实例',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  batchDeleting.value = true
  try {
    const response = await adminApi.workflowInstanceBatchDelete(ids)
    ElMessage.success(`已删除 ${Number(response.data?.deleted || ids.length)} 个流程实例`)
    if (ids.length >= list.value.length && page.value > 1) page.value -= 1
    await loadList()
  } finally {
    batchDeleting.value = false
  }
}

async function loadNotifications(instanceId = detail.value?.instance.id || '') {
  if (!canViewNotifications.value || !instanceId) {
    notifications.value = []
    return
  }
  notificationsLoading.value = true
  try {
    const response = await adminApi.workflowNotificationList({ instanceId, page: 1, pageSize: 100 })
    notifications.value = Array.isArray(response.data?.list) ? response.data.list as WorkflowNotificationRecord[] : []
  } finally {
    notificationsLoading.value = false
  }
}

async function retryNotification(notification: WorkflowNotificationRecord) {
  retryingNotificationId.value = notification.id
  try {
    await adminApi.workflowNotificationRetry(notification.id)
    ElMessage.success('通知已重发')
    await loadNotifications(notification.instanceId)
  } finally {
    retryingNotificationId.value = ''
  }
}

async function dispatchDueNotifications() {
  dispatchingDueNotifications.value = true
  try {
    const response = await adminApi.workflowNotificationDispatchDue({ limit: 100 })
    ElMessage.success(`已处理 ${Number(response.data?.dispatched || 0)} 条到期通知`)
    await loadNotifications()
  } finally {
    dispatchingDueNotifications.value = false
  }
}

async function resumeTimers() {
  if (!detail.value) return
  resumingTimers.value = true
  try {
    const instanceID = detail.value.instance.id
    const response = await adminApi.workflowInstanceResume(instanceID)
    const advanced = Number(response.data?.advanced || 0)
    if (advanced > 0) {
      ElMessage.success(`已推进 ${advanced} 个到期定时节点`)
      const detailResponse = await adminApi.workflowInstanceDetail(instanceID)
      detail.value = detailResponse.data as WorkflowInstanceDetail
      await loadList()
    } else {
      ElMessage.info('当前没有已到期的定时节点')
    }
  } finally {
    resumingTimers.value = false
  }
}

onMounted(loadList)
</script>

<style scoped>
.workflow-runtime-filters { margin-bottom: 4px; }
.mono-text { color: #475569; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.form-help { margin-top: 6px; color: #94a3b8; font-size: 12px; line-height: 1.5; }
.dialog-section { margin-top: 6px; padding-top: 14px; border-top: 1px solid #ebeef5; }
.dialog-section h3 { margin: 0 0 12px; color: #303133; font-size: 15px; font-weight: 600; }
.instance-detail { min-height: 200px; }
.detail-actions { display: flex; justify-content: flex-end; margin-top: 14px; }
.detail-section { margin-top: 22px; }
.detail-section h3 { margin: 0 0 12px; color: #303133; font-size: 15px; font-weight: 600; }
.detail-section__heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 12px; }
.detail-section__heading h3 { margin: 0; }
.detail-collapse { border-top: 0; }
.detail-collapse__title { color: #303133; font-size: 15px; font-weight: 600; }
.json-viewer { max-height: 240px; margin: 0; padding: 14px; overflow: auto; border: 1px solid #e4e7ed; border-radius: 6px; background: #f8fafc; color: #475569; font-size: 12px; line-height: 1.6; }
.history-timeline { padding-top: 6px; }
.history-timeline p { margin: 4px 0 0; color: #8492a6; }
</style>
