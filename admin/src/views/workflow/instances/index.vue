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

    <el-dialog v-model="startDialog" title="发起通用流程" width="620px" destroy-on-close>
      <el-form label-width="96px" @submit.prevent>
        <el-form-item label="流程定义" required>
          <el-select v-model="startForm.definitionId" filterable placeholder="选择已发布流程" style="width: 100%" :loading="definitionsLoading">
            <el-option
              v-for="definition in publishedDefinitions"
              :key="definition.id"
              :label="`${definition.name} (${definition.key}) · v${definition.currentVersion}`"
              :value="definition.id"
            />
          </el-select>
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
            <el-descriptions-item label="开始时间">{{ formatTime(detail.instance.startTime) }}</el-descriptions-item>
            <el-descriptions-item label="结束时间">{{ formatTime(detail.instance.endTime) }}</el-descriptions-item>
          </el-descriptions>

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
import { hasPerm } from '../../../utils/permission'
import type {
  WorkflowDefinitionSummary,
  WorkflowInstanceDetail,
  WorkflowInstanceStatus,
  WorkflowInstanceSummary,
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
const loading = ref(false)
const list = ref<WorkflowInstanceSummary[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ businessType: '', businessKey: '', status: '', starterId: '' })

const startDialog = ref(false)
const starting = ref(false)
const definitionsLoading = ref(false)
const publishedDefinitions = ref<WorkflowDefinitionSummary[]>([])
const startForm = reactive({ definitionId: 0, businessType: '', businessKey: '', variablesText: '' })

const detailDialog = ref(false)
const detailLoading = ref(false)
const detail = ref<WorkflowInstanceDetail | null>(null)

function instanceStatusMeta(status: WorkflowInstanceStatus) {
  const values = {
    running: { label: '运行中', type: 'primary' as const },
    completed: { label: '已完成', type: 'success' as const },
    rejected: { label: '已驳回', type: 'danger' as const },
    cancelled: { label: '已取消', type: 'info' as const },
  }
  return values[status] || { label: status || '未知', type: 'info' as const }
}

function taskStatusLabel(status: WorkflowTaskStatus) {
  return { waiting: '等待中', pending: '待处理', approved: '已通过', rejected: '已驳回', cancelled: '已取消' }[status] || status || '-'
}

function actionLabel(action: WorkflowTaskAction | '') {
  if (action === 'approve') return '通过'
  if (action === 'reject') return '驳回'
  return '-'
}

function historyEventLabel(type: string) {
  return {
    instance_started: '流程已发起', task_created: '任务已创建', task_activated: '任务已激活',
    task_approved: '任务已通过', task_rejected: '任务已驳回', task_cancelled: '任务已取消',
    instance_completed: '流程已完成', instance_rejected: '流程已驳回',
  }[type] || type
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
    const response = await adminApi.workflowDefinitionList({ page: 1, pageSize: 100, status: 2 })
    publishedDefinitions.value = (Array.isArray(response.data?.list) ? response.data.list : [])
      .filter((item: WorkflowDefinitionSummary) => item.currentVersion > 0)
  } finally {
    definitionsLoading.value = false
  }
}

async function openStartDialog() {
  Object.assign(startForm, { definitionId: 0, businessType: '', businessKey: '', variablesText: '' })
  startDialog.value = true
  await loadPublishedDefinitions()
}

async function startInstance() {
  if (!startForm.definitionId || !startForm.businessType.trim() || !startForm.businessKey.trim()) {
    ElMessage.warning('请选择流程定义并填写业务类型、业务标识')
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
    const response = await adminApi.workflowInstanceStart({
      definitionId: startForm.definitionId,
      definitionVersion: 0,
      businessType: startForm.businessType.trim(),
      businessKey: startForm.businessKey.trim(),
      variables,
    })
    startDialog.value = false
    ElMessage.success(`流程已发起：${response.data?.instanceId || ''}`)
    await loadList()
  } finally {
    starting.value = false
  }
}

async function openDetail(row: WorkflowInstanceSummary) {
  detail.value = null
  detailDialog.value = true
  detailLoading.value = true
  try {
    const response = await adminApi.workflowInstanceDetail(row.id)
    detail.value = response.data as WorkflowInstanceDetail
  } finally {
    detailLoading.value = false
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
.instance-detail { min-height: 200px; }
.detail-section { margin-top: 22px; }
.detail-section h3 { margin: 0 0 12px; color: #303133; font-size: 15px; font-weight: 600; }
.json-viewer { max-height: 240px; margin: 0; padding: 14px; overflow: auto; border: 1px solid #e4e7ed; border-radius: 6px; background: #f8fafc; color: #475569; font-size: 12px; line-height: 1.6; }
.history-timeline { padding-top: 6px; }
.history-timeline p { margin: 4px 0 0; color: #8492a6; }
</style>
