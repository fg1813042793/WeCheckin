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
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" :loading="loading" @click="loadList" />
        </div>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="实例 ID" min-width="190" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="mono-text">{{ row.id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="流程定义" min-width="170">
          <template #default="{ row }">
            <div class="definition-cell">
              <strong>{{ row.definitionKey || `#${row.definitionId}` }}</strong>
              <small>定义 {{ row.definitionId }} · v{{ row.definitionVersion }}</small>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="业务关联" min-width="220">
          <template #default="{ row }">
            <div class="definition-cell">
              <strong>{{ row.businessType || '-' }}</strong>
              <small>{{ row.businessKey || '-' }}</small>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="starterId" label="发起人 ID" width="120" />
        <el-table-column prop="operatorId" label="操作人 ID" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="instanceStatusMeta(row.status).type" size="small">{{ instanceStatusMeta(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="170">
          <template #default="{ row }">{{ formatTime(row.startTime) }}</template>
        </el-table-column>
        <el-table-column label="结束时间" width="170">
          <template #default="{ row }">{{ formatTime(row.endTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canViewDetail" size="small" type="primary" plain @click="openDetail(row)">详情</el-button>
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
              <el-tag :type="instanceStatusMeta(detail.instance.status).type" size="small">{{ instanceStatusMeta(detail.instance.status).label }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="流程定义">{{ detail.instance.definitionKey }} · v{{ detail.instance.definitionVersion }}</el-descriptions-item>
            <el-descriptions-item label="业务类型">{{ detail.instance.businessType }}</el-descriptions-item>
            <el-descriptions-item label="业务标识">{{ detail.instance.businessKey }}</el-descriptions-item>
            <el-descriptions-item label="发起人 ID">{{ detail.instance.starterId }}</el-descriptions-item>
            <el-descriptions-item label="操作人 ID">{{ detail.instance.operatorId }}</el-descriptions-item>
            <el-descriptions-item label="开始时间">{{ formatTime(detail.instance.startTime) }}</el-descriptions-item>
            <el-descriptions-item label="结束时间">{{ formatTime(detail.instance.endTime) }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="canStart && detail.instance.status === 'running'" class="detail-actions">
            <el-button type="primary" plain :loading="resumingTimers" @click="resumeTimers">推进到期节点</el-button>
          </div>

          <section class="detail-section">
            <h3>流程表单</h3>
            <WorkflowRuntimeForm
              :model-value="detail.formData || {}"
              :fields="detail.form || []"
              :field-access="detailFieldAccess"
              readonly
              empty-text="该实例没有流程表单数据"
            />
          </section>

          <section class="detail-section">
            <h3>流程变量</h3>
            <pre class="json-viewer">{{ formatJSON(detail.variables) }}</pre>
          </section>

          <section class="detail-section">
            <h3>任务记录</h3>
            <el-table :data="detail.tasks" size="small" border>
              <el-table-column prop="nodeName" label="节点" min-width="140" />
              <el-table-column prop="assigneeId" label="处理人 ID" width="120" />
              <el-table-column label="状态" width="100">
                <template #default="{ row }">{{ taskStatusLabel(row.status) }}</template>
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
              <el-table-column prop="recipientUserId" label="接收人 ID" width="110" />
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
                  >重试</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>

          <section class="detail-section">
            <h3>流转历史</h3>
            <el-timeline class="history-timeline">
              <el-timeline-item v-for="event in detail.history" :key="event.id" :timestamp="formatTime(event.eventTime)">
                <strong>{{ historyEventLabel(event.eventType) }}</strong>
                <span v-if="event.actorId"> · 操作人 {{ event.actorId }}</span>
                <p v-if="event.message">{{ event.message }}</p>
              </el-timeline-item>
            </el-timeline>
          </section>
        </template>
      </div>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'
import type { AdminUser } from '../../../api/types'
import { hasPerm } from '../../../utils/permission'
import WorkflowRuntimeForm from '../components/WorkflowRuntimeForm.vue'
import WorkflowUserTreePicker from '../components/WorkflowUserTreePicker.vue'
import { initialWorkflowFormData, workflowFieldActionMap, workflowFieldAccessMap, writableWorkflowFormData } from '../runtimeForm'
import type {
  WorkflowInstanceDetail,
  WorkflowInstanceStatus,
  WorkflowInstanceSummary,
  WorkflowNotificationChannel,
  WorkflowNotificationKind,
  WorkflowNotificationRecord,
  WorkflowNotificationStatus,
  WorkflowPublishedDefinition,
  WorkflowTaskAction,
  WorkflowTaskStatus,
} from '../types'

const statusOptions = [
  { label: '运行中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '已驳回', value: 'rejected' },
  { label: '已取消', value: 'cancelled' },
]

const canStart = computed(() => hasPerm('admin:menu:workflow:instance:start'))
const canViewDetail = computed(() => hasPerm('admin:menu:workflow:instance:detail'))
const canViewNotifications = computed(() => hasPerm('admin:menu:workflow:notification:list'))
const canRetryNotifications = computed(() => hasPerm('admin:menu:workflow:notification:retry'))
const loading = ref(false)
const list = ref<WorkflowInstanceSummary[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ businessType: '', businessKey: '', status: '', starterId: '' })

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
  if (initiator?.scope !== 'specified') return initiatorUsers.value
  const allowed = new Set((initiator.userIds || []).map(id => Number(id)))
  return initiatorUsers.value.filter(user => allowed.has(Number(user.id)))
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

function instanceStatusMeta(status: WorkflowInstanceStatus) {
  const values = {
    running: { label: '运行中', type: 'primary' as const },
    completed: { label: '已完成', type: 'success' as const },
    rejected: { label: '已驳回', type: 'danger' as const },
    cancelled: { label: '已取消', type: 'info' as const },
  }
  return values[status] || { label: status || '未知', type: 'info' as const }
}

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

function taskStatusLabel(status: WorkflowTaskStatus) {
  return { waiting: '等待中', pending: '待处理', completed: '已提交', approved: '已通过', rejected: '已驳回', cancelled: '已取消' }[status] || status || '-'
}

function actionLabel(action: WorkflowTaskAction | '') {
  if (action === 'approve') return '通过'
  if (action === 'reject') return '驳回'
  if (action === 'submit') return '提交'
  return '-'
}

function historyEventLabel(type: string) {
  return {
    instance_started: '流程已发起', task_created: '任务已创建', task_activated: '任务已激活',
    task_approved: '任务已通过', task_rejected: '任务已驳回', task_submitted: '任务已提交', task_cancelled: '任务已取消',
    node_cc: '已记录抄送', node_notify: '通知节点已触发', node_automated: '自动动作已执行', timer_waiting: '定时节点等待中', timer_resumed: '定时节点已推进',
    instance_completed: '流程已完成', instance_rejected: '流程已驳回',
  }[type] || type
}

function notificationStatusMeta(status: WorkflowNotificationStatus) {
  return notificationStatusOptions.find(option => option.value === status)
    || { value: status, label: status || '未知', type: 'info' as const }
}

function notificationKindLabel(kind: WorkflowNotificationKind) {
  return { node_cc: '节点抄送', node_notify: '节点通知', task_arrived: '任务到达' }[kind] || kind
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
    ElMessage.success('已执行通知重试')
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
.definition-cell strong { display: block; color: #1f2937; font-weight: 600; }
.definition-cell small { display: block; margin-top: 3px; color: #94a3b8; font-size: 12px; }
.form-help { margin-top: 6px; color: #94a3b8; font-size: 12px; line-height: 1.5; }
.dialog-section { margin-top: 6px; padding-top: 14px; border-top: 1px solid #ebeef5; }
.dialog-section h3 { margin: 0 0 12px; color: #303133; font-size: 15px; font-weight: 600; }
.instance-detail { min-height: 200px; }
.detail-actions { display: flex; justify-content: flex-end; margin-top: 14px; }
.detail-section { margin-top: 22px; }
.detail-section h3 { margin: 0 0 12px; color: #303133; font-size: 15px; font-weight: 600; }
.detail-section__heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 12px; }
.detail-section__heading h3 { margin: 0; }
.json-viewer { max-height: 240px; margin: 0; padding: 14px; overflow: auto; border: 1px solid #e4e7ed; border-radius: 6px; background: #f8fafc; color: #475569; font-size: 12px; line-height: 1.6; }
.history-timeline { padding-top: 6px; }
.history-timeline p { margin: 4px 0 0; color: #8492a6; }
</style>
