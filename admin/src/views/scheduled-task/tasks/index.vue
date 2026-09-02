<template>
  <div class="admin-page scheduled-task-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar task-filters">
        <div class="admin-toolbar__left">
          <el-input v-model="filters.keyword" clearable placeholder="搜索任务名称或编码" style="width: 260px" @keyup.enter="search" />
          <el-select v-model="filters.handlerType" clearable placeholder="全部处理器" style="width: 150px" @change="search">
            <el-option v-for="item in handlers" :key="item.type" :label="handlerTypeLabel(item.type)" :value="item.type" />
          </el-select>
          <el-select v-model="filters.enabled" clearable placeholder="全部状态" style="width: 130px" @change="search">
            <el-option label="已启用" value="true" />
            <el-option label="已停用" value="false" />
          </el-select>
          <el-button type="primary" icon="Search" @click="search">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
      </div>

      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-button v-if="canAdd" type="primary" icon="Plus" @click="openCreate">创建任务</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" :loading="loading" @click="loadList" />
        </div>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column label="任务" min-width="230">
          <template #default="{ row }">
            <div class="task-name">
              <span class="task-name__icon"><el-icon><Timer /></el-icon></span>
              <div><strong>{{ row.name }}</strong><small>{{ row.code }}</small></div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="处理器" width="145">
          <template #default="{ row }"><el-tag size="small" :type="handlerTag(row.handlerType)">{{ handlerTypeLabel(row.handlerType) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="调度" min-width="230">
          <template #default="{ row }">
            <div class="schedule-cell"><code>{{ row.cronExpression }}</code><small>{{ row.timezone }} · {{ precisionLabel(row.cronPrecision) }}</small></div>
          </template>
        </el-table-column>
        <el-table-column label="并发 / 错过" min-width="170">
          <template #default="{ row }"><span>{{ concurrencyLabel(row.concurrencyPolicy) }}</span><span class="policy-divider">/</span><span>{{ misfireLabel(row.misfirePolicy) }}</span></template>
        </el-table-column>
        <el-table-column label="下次执行" width="180"><template #default="{ row }">{{ formatTime(row.nextRunAt) }}</template></el-table-column>
        <el-table-column label="状态" width="96" align="center">
          <template #default="{ row }">
            <el-switch v-if="canStatus" :model-value="row.enabled === 1" :loading="statusTaskId === row.id" @change="setStatusValue(row, $event)" />
            <el-tag v-else :type="row.enabled === 1 ? 'success' : 'info'" size="small">{{ row.enabled === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-button v-if="canEdit" size="small" icon="EditPen" @click="openEdit(row)">编辑</el-button>
              <el-button v-if="canRun" size="small" type="primary" plain icon="VideoPlay" :loading="runTaskId === row.id" @click="runNow(row)">运行</el-button>
              <el-button v-if="canDelete" size="small" type="danger" plain icon="Delete" @click="remove(row)" />
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="admin-pagination">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :page-sizes="[10, 20, 50, 100]" :total="total" layout="total,sizes,prev,pager,next" @current-change="loadList" @size-change="changePageSize" />
      </div>
    </el-card>

    <TaskEditorDialog v-model="editorVisible" :task="editTask" :handlers="handlers" @saved="loadList" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../../api'
import { hasPerm } from '../../../utils/permission'
import TaskEditorDialog from '../components/TaskEditorDialog.vue'
import { handlerTypeLabel } from '../handlerLabels'
import type { HandlerMetadata, HandlerType, ScheduledTask } from '../types'

const loading = ref(false)
const list = ref<ScheduledTask[]>([])
const handlers = ref<HandlerMetadata[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const statusTaskId = ref<number | null>(null)
const runTaskId = ref<number | null>(null)
const editorVisible = ref(false)
const editTask = ref<ScheduledTask | null>(null)
const filters = reactive({ keyword: '', handlerType: '' as HandlerType | '', enabled: '' })

const canAdd = computed(() => hasPerm('admin:menu:scheduled-task:add'))
const canEdit = computed(() => hasPerm('admin:menu:scheduled-task:edit'))
const canDelete = computed(() => hasPerm('admin:menu:scheduled-task:delete'))
const canStatus = computed(() => hasPerm('admin:menu:scheduled-task:status'))
const canRun = computed(() => hasPerm('admin:menu:scheduled-task:run'))

onMounted(async () => {
  await Promise.all([loadHandlers(), loadList()])
})

async function loadHandlers() {
  const response = await adminApi.scheduledTaskHandlers()
  handlers.value = Array.isArray(response.data) ? response.data : []
}

async function loadList() {
  loading.value = true
  try {
    const response = await adminApi.scheduledTaskList({ page: page.value, pageSize: pageSize.value, keyword: filters.keyword.trim(), handlerType: filters.handlerType, enabled: filters.enabled })
    list.value = Array.isArray(response.data?.list) ? response.data.list : []
    total.value = Number(response.data?.total || 0)
  } finally { loading.value = false }
}

function search() { page.value = 1; loadList() }
function resetFilters() { Object.assign(filters, { keyword: '', handlerType: '', enabled: '' }); search() }
function changePageSize() { page.value = 1; loadList() }
function openCreate() { editTask.value = null; editorVisible.value = true }
function openEdit(row: ScheduledTask) { editTask.value = row; editorVisible.value = true }

async function setStatus(row: ScheduledTask, enabled: boolean) {
  statusTaskId.value = row.id
  try {
    const response = await adminApi.scheduledTaskStatus(row.id, { enabled, version: row.version })
    Object.assign(row, response.data)
    ElMessage.success(enabled ? '任务已启用' : '任务已停用')
  } finally { statusTaskId.value = null }
}

function setStatusValue(row: ScheduledTask, value: string | number | boolean) {
  setStatus(row, Boolean(value))
}

async function runNow(row: ScheduledTask) {
  await ElMessageBox.confirm(`立即运行“${row.name}”？`, '立即运行', { type: 'warning', confirmButtonText: '运行', cancelButtonText: '取消' })
  runTaskId.value = row.id
  try {
    const response = await adminApi.scheduledTaskRun(row.id)
    ElMessage.success(response.data?.dispatchPending ? '运行记录已创建，等待调度服务恢复投递' : '运行记录已创建')
  } finally { runTaskId.value = null }
}

async function remove(row: ScheduledTask) {
  await ElMessageBox.confirm(`删除任务“${row.name}”？历史运行记录会保留。`, '删除任务', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
  await adminApi.scheduledTaskDelete(row.id)
  ElMessage.success('任务已删除')
  if (list.value.length === 1 && page.value > 1) page.value--
  await loadList()
}

function handlerTag(type: HandlerType) { return type === 'shell' || type === 'sql' ? 'danger' : type === 'http' ? 'warning' : type === 'workflow' ? 'primary' : 'success' }
function precisionLabel(value: string) { return value === 'second' ? '秒级' : '分钟级' }
function concurrencyLabel(value: string) { return ({ skip: '冲突跳过', queue_once: '合并等待', allow: '允许并行' } as Record<string, string>)[value] || value }
function misfireLabel(value: string) { return ({ skip: '错过跳过', fire_once: '补一次', catch_up: '有限补跑' } as Record<string, string>)[value] || value }
function formatTime(value: number) { return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-' }
</script>

<style scoped>
.task-name { display: flex; align-items: center; gap: 10px; min-width: 0; }
.task-name__icon { display: grid; place-items: center; flex: 0 0 34px; width: 34px; height: 34px; border-radius: 6px; color: #2563eb; background: #eff6ff; }
.task-name div, .schedule-cell { display: flex; flex-direction: column; min-width: 0; gap: 3px; }
.task-name strong { overflow: hidden; text-overflow: ellipsis; }
.task-name small, .schedule-cell small { color: var(--admin-muted); }
.schedule-cell code { color: #334155; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.policy-divider { margin: 0 5px; color: #cbd5e1; }
</style>
