<template>
  <div class="admin-page workflow-task-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar workflow-runtime-filters">
        <div class="admin-toolbar__left">
          <el-input v-model="filters.instanceId" placeholder="流程实例 ID" clearable style="width: 240px" @keyup.enter="search" />
          <el-input v-model="filters.assigneeId" placeholder="处理人 ID" clearable style="width: 160px" @keyup.enter="search" />
          <el-select v-model="filters.status" placeholder="全部状态" clearable style="width: 140px" @change="search">
            <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
          <el-button type="primary" @click="search">搜索</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" :loading="loading" @click="loadList" />
        </div>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="任务 ID" min-width="190" show-overflow-tooltip>
          <template #default="{ row }"><span class="mono-text">{{ row.id }}</span></template>
        </el-table-column>
        <el-table-column prop="instanceId" label="实例 ID" min-width="190" show-overflow-tooltip>
          <template #default="{ row }"><span class="mono-text">{{ row.instanceId }}</span></template>
        </el-table-column>
        <el-table-column label="任务节点" min-width="170">
          <template #default="{ row }">
            <div class="task-node">
              <strong>{{ row.nodeName || row.nodeId }}</strong>
              <small>{{ approvalModeLabel(row.approvalMode) }}<template v-if="row.total > 1"> · {{ row.sequence }}/{{ row.total }}</template></small>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="assigneeId" label="处理人 ID" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="taskStatusMeta(row.status).type" size="small">{{ taskStatusMeta(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="处理结果" width="100">
          <template #default="{ row }">{{ actionLabel(row.action) }}</template>
        </el-table-column>
        <el-table-column prop="comment" label="处理意见" min-width="180" show-overflow-tooltip />
        <el-table-column label="处理时间" width="170">
          <template #default="{ row }">{{ formatTime(row.handledAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="canComplete && row.status === 'pending'"
              size="small"
              type="primary"
              @click="openCompleteDialog(row)"
            >处理</el-button>
            <span v-else class="muted-text">-</span>
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

    <el-dialog v-model="completeDialog" title="处理流程任务" width="700px" destroy-on-close>
      <el-descriptions v-if="activeTask" :column="2" border class="task-summary">
        <el-descriptions-item label="任务节点">{{ activeTask.nodeName || activeTask.nodeId }}</el-descriptions-item>
        <el-descriptions-item label="处理人 ID">{{ activeTask.assigneeId }}</el-descriptions-item>
        <el-descriptions-item label="实例 ID" :span="2"><span class="mono-text">{{ activeTask.instanceId }}</span></el-descriptions-item>
      </el-descriptions>
      <section v-loading="taskDetailLoading" class="complete-runtime-form">
        <h3>流程表单</h3>
        <WorkflowRuntimeForm
          ref="completeRuntimeForm"
          v-model="completeFormData"
          :fields="activeInstanceDetail?.form || []"
          :field-access="activeTaskFieldAccess"
          :field-actions="activeTaskFieldActions"
          empty-text="该任务没有可处理的流程表单"
        />
      </section>
      <el-form label-width="86px" @submit.prevent>
        <el-form-item label="处理结果" required>
          <el-tag v-if="activeTaskNodeType === 'handle'" type="primary">提交</el-tag>
          <el-radio-group v-else v-model="completeForm.action">
            <el-radio-button value="approve">通过</el-radio-button>
            <el-radio-button value="reject">驳回</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="activeTaskNodeType === 'handle' ? '办理说明' : '处理意见'">
          <el-input v-model="completeForm.comment" type="textarea" :rows="4" maxlength="500" show-word-limit :placeholder="activeTaskNodeType === 'handle' ? '填写办理说明' : '填写审批意见'" />
        </el-form-item>
        <el-form-item label="更新变量">
          <el-input
            v-model="completeForm.variablesText"
            type="textarea"
            :rows="6"
            placeholder="可选，填写需要合并到流程上下文的 JSON 对象"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="completeDialog = false">取消</el-button>
        <el-button :type="completeForm.action === 'reject' ? 'danger' : 'primary'" :loading="completing" :disabled="taskDetailLoading" @click="completeTask">
          确认{{ completeActionText }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../../api'
import { hasPerm } from '../../../utils/permission'
import WorkflowRuntimeForm from '../components/WorkflowRuntimeForm.vue'
import { initialWorkflowFormData, workflowFieldActionMap, workflowFieldAccessMap, writableWorkflowFormData } from '../runtimeForm'
import type { WorkflowInstanceDetail, WorkflowTaskAction, WorkflowTaskStatus, WorkflowTaskSummary } from '../types'

const statusOptions = [
  { label: '待激活', value: 'waiting' },
  { label: '待处理', value: 'pending' },
  { label: '已提交', value: 'completed' },
  { label: '已通过', value: 'approved' },
  { label: '已驳回', value: 'rejected' },
  { label: '已取消', value: 'cancelled' },
]

const canComplete = computed(() => hasPerm('admin:menu:workflow:task:complete'))
const loading = ref(false)
const list = ref<WorkflowTaskSummary[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ instanceId: '', assigneeId: '', status: '' })

const completeDialog = ref(false)
const completing = ref(false)
const taskDetailLoading = ref(false)
const activeTask = ref<WorkflowTaskSummary | null>(null)
const activeInstanceDetail = ref<WorkflowInstanceDetail | null>(null)
const completeFormData = ref<Record<string, unknown>>({})
const completeRuntimeForm = ref<{ validate: () => boolean; resetValidation: () => void } | null>(null)
const completeForm = reactive<{ action: WorkflowTaskAction; comment: string; variablesText: string }>({
  action: 'approve', comment: '', variablesText: '',
})

const activeTaskNodeType = computed(() => {
  if (!activeTask.value || !activeInstanceDetail.value) return 'approval'
  return activeInstanceDetail.value.nodeTypes?.[activeTask.value.nodeId] || 'approval'
})

const completeActionText = computed(() => {
  if (completeForm.action === 'submit') return '提交'
  if (completeForm.action === 'reject') return '驳回'
  return '通过'
})

const activeTaskFieldAccess = computed(() => {
  if (!activeTask.value || !activeInstanceDetail.value) return {}
  return workflowFieldAccessMap(
    activeInstanceDetail.value.form || [],
    activeInstanceDetail.value.fieldPermissions || {},
    activeTask.value.nodeId,
    'read',
  )
})

const activeTaskFieldActions = computed(() => {
  if (!activeTask.value || !activeInstanceDetail.value) return {}
  return workflowFieldActionMap(
    activeInstanceDetail.value.form || [],
    activeInstanceDetail.value.fieldPermissions || {},
    activeTask.value.nodeId,
  )
})

function taskStatusMeta(status: WorkflowTaskStatus) {
  const values = {
    waiting: { label: '待激活', type: 'info' as const }, pending: { label: '待处理', type: 'warning' as const },
    completed: { label: '已提交', type: 'success' as const },
    approved: { label: '已通过', type: 'success' as const }, rejected: { label: '已驳回', type: 'danger' as const },
    cancelled: { label: '已取消', type: 'info' as const },
  }
  return values[status] || { label: status || '未知', type: 'info' as const }
}

function actionLabel(action: WorkflowTaskAction | '') {
  if (action === 'approve') return '通过'
  if (action === 'reject') return '驳回'
  if (action === 'submit') return '提交'
  return '-'
}

function approvalModeLabel(mode: string) {
  return { single: '单人审批', sequential: '依次审批', parallel: '并行审批', countersign: '会签审批' }[mode] || mode || '审批'
}

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function parseVariables(text: string) {
  if (!text.trim()) return {}
  const value = JSON.parse(text)
  if (!value || Array.isArray(value) || typeof value !== 'object') throw new Error('更新变量必须是 JSON 对象')
  return value as Record<string, unknown>
}

async function loadList() {
  loading.value = true
  try {
    const response = await adminApi.workflowTaskList({
      page: page.value, pageSize: pageSize.value,
      instanceId: filters.instanceId.trim(), assigneeId: filters.assigneeId.trim(), status: filters.status,
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
  Object.assign(filters, { instanceId: '', assigneeId: '', status: '' })
  search()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  page.value = 1
  loadList()
}

async function openCompleteDialog(row: WorkflowTaskSummary) {
  activeTask.value = row
  activeInstanceDetail.value = null
  completeFormData.value = {}
  Object.assign(completeForm, { action: 'approve', comment: '', variablesText: '' })
  completeDialog.value = true
  taskDetailLoading.value = true
  try {
    const response = await adminApi.workflowInstanceDetail(row.instanceId)
    activeInstanceDetail.value = response.data as WorkflowInstanceDetail
    completeForm.action = activeInstanceDetail.value.nodeTypes?.[row.nodeId] === 'handle' ? 'submit' : 'approve'
    completeFormData.value = initialWorkflowFormData(
      activeInstanceDetail.value.form || [],
      activeInstanceDetail.value.formData || {},
    )
    completeRuntimeForm.value?.resetValidation()
  } finally {
    taskDetailLoading.value = false
  }
}

async function completeTask() {
  if (!activeTask.value) return
  if (completeRuntimeForm.value && !completeRuntimeForm.value.validate()) {
    ElMessage.warning('请检查流程表单中的校验提示')
    return
  }
  let variables: Record<string, unknown>
  try {
    variables = parseVariables(completeForm.variablesText)
  } catch (error) {
    ElMessage.warning(error instanceof Error ? error.message : '更新变量 JSON 格式无效')
    return
  }
  const actionText = completeActionText.value
  try {
    await ElMessageBox.confirm(`确定${actionText}任务“${activeTask.value.nodeName || activeTask.value.nodeId}”？`, `确认${actionText}`, {
      type: completeForm.action === 'reject' ? 'warning' : 'info',
      confirmButtonText: `确认${actionText}`,
    })
  } catch {
    return
  }
  completing.value = true
  try {
    await adminApi.workflowTaskComplete(activeTask.value.id, {
      action: completeForm.action,
      comment: completeForm.comment.trim(),
      variables,
      formData: writableWorkflowFormData(activeInstanceDetail.value?.form || [], completeFormData.value, activeTaskFieldAccess.value),
    })
    completeDialog.value = false
    ElMessage.success(`任务已${actionText}`)
    await loadList()
  } finally {
    completing.value = false
  }
}

onMounted(loadList)
</script>

<style scoped>
.workflow-runtime-filters { margin-bottom: 4px; }
.mono-text { color: #475569; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.task-node strong { display: block; color: #1f2937; font-weight: 600; }
.task-node small { display: block; margin-top: 3px; color: #94a3b8; font-size: 12px; }
.muted-text { color: #c0c4cc; }
.task-summary { margin-bottom: 20px; }
.complete-runtime-form { margin-bottom: 18px; padding-bottom: 2px; }
.complete-runtime-form h3 { margin: 0 0 12px; color: #303133; font-size: 15px; font-weight: 600; }
</style>
