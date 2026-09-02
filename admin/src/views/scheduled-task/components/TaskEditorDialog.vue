<template>
  <el-dialog
    :model-value="modelValue"
    :title="task ? '编辑定时任务' : '创建定时任务'"
    width="min(960px, 94vw)"
    top="5vh"
    append-to-body
    destroy-on-close
    :close-on-click-modal="false"
    class="task-editor-dialog"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top" @submit.prevent>
      <div class="editor-section">
        <div class="editor-section__title">基础信息</div>
        <div class="editor-grid editor-grid--two">
          <el-form-item label="任务名称" required>
            <el-input v-model="form.name" maxlength="200" placeholder="例如：每日派发流程通知" />
          </el-form-item>
          <el-form-item label="任务编码" required>
            <el-input v-model="form.code" :disabled="Boolean(task)" maxlength="100" placeholder="workflow.notification.daily" />
          </el-form-item>
        </div>
        <el-form-item label="任务说明">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="500" show-word-limit />
        </el-form-item>
      </div>

      <div class="editor-section">
        <div class="editor-section__title">处理器</div>
        <el-form-item label="处理器类型" required>
          <el-select v-model="form.handlerType" style="width: 100%" @change="handleHandlerTypeChange">
            <el-option v-for="item in availableHandlers" :key="item.type" :label="handlerTypeLabel(item.type)" :value="item.type">
              <div class="handler-option">
                <span>{{ handlerTypeLabel(item.type) }}</span>
                <el-tag :type="riskMeta(item.riskLevel).type" size="small">{{ riskMeta(item.riskLevel).label }}</el-tag>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <template v-if="form.handlerType === 'go'">
          <el-form-item label="注册任务" required>
            <el-select v-model="handlerConfig.handlerKey" filterable style="width: 100%" placeholder="选择服务端注册任务">
              <el-option v-for="key in schemaEnum('go', 'handlerKey')" :key="key" :label="key" :value="key" />
            </el-select>
          </el-form-item>
          <el-form-item label="任务参数 JSON">
            <el-input v-model="jsonFields.params" type="textarea" :rows="4" spellcheck="false" />
          </el-form-item>
        </template>

        <template v-else-if="form.handlerType === 'workflow'">
          <div class="editor-grid editor-grid--two">
            <el-form-item label="流程定义" required>
              <el-select
                v-model="handlerConfig.definitionId"
                filterable
                :loading="workflowOptionsLoading"
                placeholder="请选择已发布流程"
                style="width: 100%"
                @change="handleWorkflowDefinitionChange"
              >
                <el-option
                  v-for="definition in workflowDefinitions"
                  :key="definition.id"
                  :label="`${definition.name} (${definition.key}) · v${definition.version}`"
                  :value="definition.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="流程发起人" required>
              <WorkflowUserTreePicker
                v-model="workflowStarterIds"
                :departments="workflowDepartments"
                :users="eligibleWorkflowUsers"
                :multiple="true"
                :disabled="workflowOptionsLoading"
                placeholder="请选择一个或多个流程发起人"
              />
            </el-form-item>
          </div>
          <el-form-item label="版本策略">
            <el-radio-group v-model="handlerConfig.versionPolicy">
              <el-radio-button value="latest_published">最新已发布版本</el-radio-button>
              <el-radio-button value="fixed_version">固定版本</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="handlerConfig.versionPolicy === 'fixed_version'" label="固定版本号" required>
            <el-input-number v-model="handlerConfig.fixedVersion" :min="1" controls-position="right" />
          </el-form-item>
          <div class="editor-grid editor-grid--two">
            <el-form-item label="流程变量 JSON">
              <el-input v-model="jsonFields.variables" type="textarea" :rows="5" spellcheck="false" />
            </el-form-item>
            <el-form-item label="初始表单 JSON">
              <el-input v-model="jsonFields.formData" type="textarea" :rows="5" spellcheck="false" />
            </el-form-item>
          </div>
        </template>

        <template v-else-if="form.handlerType === 'http'">
          <div class="editor-grid editor-grid--method">
            <el-form-item label="请求方法" required>
              <el-select v-model="handlerConfig.method">
                <el-option v-for="method in ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']" :key="method" :value="method" />
              </el-select>
            </el-form-item>
            <el-form-item label="请求地址" required>
              <el-input v-model="handlerConfig.url" placeholder="https://已加入白名单的域名/path" />
            </el-form-item>
          </div>
          <div class="editor-grid editor-grid--two">
            <el-form-item label="超时（毫秒）">
              <el-input-number v-model="handlerConfig.timeoutMillis" :min="100" :max="86400000" controls-position="right" style="width: 100%" />
            </el-form-item>
            <el-form-item label="服务端凭据引用">
              <el-input v-model="handlerConfig.credentialRef" placeholder="不在任务中填写 Token" />
            </el-form-item>
          </div>
          <div class="editor-grid editor-grid--two">
            <el-form-item label="Query JSON"><el-input v-model="jsonFields.query" type="textarea" :rows="4" spellcheck="false" /></el-form-item>
            <el-form-item label="请求头 JSON"><el-input v-model="jsonFields.headers" type="textarea" :rows="4" spellcheck="false" /></el-form-item>
          </div>
          <el-form-item label="请求 Body JSON"><el-input v-model="jsonFields.body" type="textarea" :rows="4" spellcheck="false" /></el-form-item>
          <el-form-item label="预期状态码">
            <el-input v-model="jsonFields.expectedStatus" placeholder="例如：200,201,204；留空表示 2xx" />
          </el-form-item>
        </template>

        <template v-else-if="form.handlerType === 'shell'">
          <el-alert type="warning" :closable="false" show-icon title="只能执行服务端注册的绝对路径命令，参数和环境变量受白名单限制。" />
          <el-form-item label="命令" required>
            <el-select v-model="handlerConfig.commandKey" style="width: 100%">
              <el-option v-for="key in schemaEnum('shell', 'commandKey')" :key="key" :label="key" :value="key" />
            </el-select>
          </el-form-item>
          <div class="editor-grid editor-grid--two">
            <el-form-item label="参数数组 JSON"><el-input v-model="jsonFields.args" type="textarea" :rows="5" spellcheck="false" /></el-form-item>
            <el-form-item label="环境变量 JSON"><el-input v-model="jsonFields.env" type="textarea" :rows="5" spellcheck="false" /></el-form-item>
          </div>
        </template>

        <template v-else-if="form.handlerType === 'sql'">
          <el-alert type="warning" :closable="false" show-icon title="数据源由服务端注册；普通语句仅允许单条 SELECT、INSERT、UPDATE 或 DELETE。" />
          <div class="editor-grid editor-grid--two">
            <el-form-item label="数据源键" required><el-input v-model="handlerConfig.dataSourceKey" /></el-form-item>
            <el-form-item label="执行模式">
              <el-radio-group v-model="handlerConfig.mode">
                <el-radio-button value="read">查询</el-radio-button>
                <el-radio-button value="write" :disabled="!canSQLWrite">写入</el-radio-button>
              </el-radio-group>
            </el-form-item>
          </div>
          <el-form-item label="SQL 语句">
            <el-input v-model="handlerConfig.statement" type="textarea" :rows="5" placeholder="使用 ? 占位符；调用存储过程时留空" spellcheck="false" />
          </el-form-item>
          <el-form-item v-if="handlerConfig.mode === 'write'" label="注册存储过程键">
            <el-input v-model="handlerConfig.procedureKey" placeholder="与 SQL 语句二选一" />
          </el-form-item>
          <div class="editor-grid editor-grid--three">
            <el-form-item label="参数数组 JSON"><el-input v-model="jsonFields.parameters" type="textarea" :rows="3" spellcheck="false" /></el-form-item>
            <el-form-item label="最大返回行"><el-input-number v-model="handlerConfig.maxRows" :min="1" controls-position="right" /></el-form-item>
            <el-form-item label="最大影响行"><el-input-number v-model="handlerConfig.maxAffected" :min="1" controls-position="right" /></el-form-item>
          </div>
        </template>
      </div>

      <div class="editor-section">
        <div class="editor-section__title editor-section__title--actions">
          <span>调度策略</span>
          <el-button :loading="previewing" icon="View" @click="cronPreview">预览</el-button>
        </div>
        <div class="editor-grid editor-grid--cron">
          <el-form-item label="精度">
            <el-radio-group v-model="form.cronPrecision">
              <el-radio-button value="minute">分钟</el-radio-button>
              <el-radio-button value="second">秒</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="Cron 表达式" required><el-input v-model="form.cronExpression" /></el-form-item>
          <el-form-item label="时区"><el-select v-model="form.timezone" filterable><el-option v-for="zone in timezones" :key="zone" :value="zone" /></el-select></el-form-item>
        </div>
        <div v-if="preview.length" class="cron-preview">
          <span v-for="item in preview" :key="item.utcMillis">{{ item.localTime }}</span>
        </div>
        <div class="editor-grid editor-grid--three">
          <el-form-item label="错过执行">
            <el-select v-model="form.misfirePolicy"><el-option label="跳过" value="skip" /><el-option label="补一次" value="fire_once" /><el-option label="有限补跑" value="catch_up" /></el-select>
          </el-form-item>
          <el-form-item label="最大补跑数"><el-input-number v-model="form.maxCatchUp" :min="1" :max="100" controls-position="right" /></el-form-item>
          <el-form-item label="并发策略">
            <el-select v-model="form.concurrencyPolicy"><el-option label="跳过新运行" value="skip" /><el-option label="合并等待一次" value="queue_once" /><el-option label="允许并行" value="allow" /></el-select>
          </el-form-item>
        </div>
        <div class="editor-grid editor-grid--three">
          <el-form-item label="超时（秒）"><el-input-number v-model="form.timeoutSeconds" :min="1" :max="86400" controls-position="right" /></el-form-item>
          <el-form-item label="自动重试次数"><el-input-number v-model="form.maxRetries" :min="0" :max="5" controls-position="right" /></el-form-item>
          <el-form-item label="重试间隔（秒）"><el-input-number v-model="form.retrySeconds" :min="1" :max="86400" controls-position="right" /></el-form-item>
        </div>
        <el-form-item label="创建后启用"><el-switch v-model="form.enabled" /></el-form-item>
      </div>
    </el-form>

    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'
import { hasPerm } from '../../../utils/permission'
import WorkflowUserTreePicker from '../../workflow/components/WorkflowUserTreePicker.vue'
import type { WorkflowAssigneeUser, WorkflowPublishedDefinition } from '../../workflow/types'
import { handlerTypeLabel } from '../handlerLabels'
import type { CronOccurrence, HandlerMetadata, HandlerType, ScheduledTask, ScheduledTaskPayload } from '../types'

type DepartmentNode = {
  id: number
  name: string
  children?: DepartmentNode[]
}

const props = defineProps<{ modelValue: boolean; task: ScheduledTask | null; handlers: HandlerMetadata[] }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; saved: [] }>()

const timezones = ['Asia/Shanghai', 'UTC', 'Asia/Hong_Kong', 'Asia/Tokyo', 'America/New_York', 'Europe/London']
const saving = ref(false)
const previewing = ref(false)
const preview = ref<CronOccurrence[]>([])
const workflowOptionsLoading = ref(false)
const workflowOptionsLoaded = ref(false)
const workflowDefinitions = ref<WorkflowPublishedDefinition[]>([])
const workflowUsers = ref<WorkflowAssigneeUser[]>([])
const workflowDepartments = ref<DepartmentNode[]>([])
const form = reactive({
  code: '', name: '', description: '', handlerType: 'go' as HandlerType,
  cronExpression: '0 * * * *', cronPrecision: 'minute' as const, timezone: 'Asia/Shanghai', enabled: true,
  misfirePolicy: 'skip' as const, maxCatchUp: 1, concurrencyPolicy: 'skip' as const,
  timeoutSeconds: 300, maxRetries: 0, retrySeconds: 30,
})
const handlerConfig = reactive<Record<string, any>>({})
const jsonFields = reactive({
  params: '{}', variables: '{}', formData: '{}', query: '{}', headers: '{}', body: '{}', expectedStatus: '',
  args: '[]', env: '{}', parameters: '[]',
})

const canHTTP = computed(() => hasPerm('admin:menu:scheduled-task:http'))
const canShell = computed(() => hasPerm('admin:menu:scheduled-task:shell'))
const canSQLRead = computed(() => hasPerm('admin:menu:scheduled-task:sql:read'))
const canSQLWrite = computed(() => hasPerm('admin:menu:scheduled-task:sql:write'))
const availableHandlers = computed(() => props.handlers.filter(item => {
  if (item.type === 'http') return canHTTP.value
  if (item.type === 'shell') return canShell.value
  if (item.type === 'sql') return canSQLRead.value || canSQLWrite.value
  return true
}))
const selectedWorkflowDefinition = computed(() => (
  workflowDefinitions.value.find(definition => Number(definition.id) === Number(handlerConfig.definitionId)) || null
))
const eligibleWorkflowUsers = computed(() => {
  const initiator = selectedWorkflowDefinition.value?.initiator
  if (initiator?.scope !== 'specified') return workflowUsers.value
  const allowedUserIDs = new Set((initiator.userIds || []).map(Number))
  const allowedDepartmentIDs = new Set((initiator.departmentIds || []).map(Number))
  return workflowUsers.value.filter(user => (
    allowedUserIDs.has(Number(user.id))
    || (user.deptIds || []).some(departmentID => allowedDepartmentIDs.has(Number(departmentID)))
  ))
})
const workflowStarterIds = computed<number[]>({
  get() {
    const values = Array.isArray(handlerConfig.starterIds)
      ? handlerConfig.starterIds
      : (handlerConfig.starterId ? [handlerConfig.starterId] : [])
    return normalizeIDs(values)
  },
  set(values) {
    handlerConfig.starterIds = normalizeIDs(values)
    delete handlerConfig.starterId
  },
})

watch(() => props.modelValue, async visible => {
  if (!visible) return
  loadTask(props.task)
  if (form.handlerType === 'workflow') await loadWorkflowOptions()
}, { immediate: true })

function clearObject(target: Record<string, any>) {
  Object.keys(target).forEach(key => delete target[key])
}

function loadTask(task: ScheduledTask | null) {
  preview.value = []
  Object.assign(form, task ? {
    code: task.code, name: task.name, description: task.description || '', handlerType: task.handlerType,
    cronExpression: task.cronExpression, cronPrecision: task.cronPrecision, timezone: task.timezone,
    enabled: task.enabled === 1, misfirePolicy: task.misfirePolicy, maxCatchUp: task.maxCatchUp,
    concurrencyPolicy: task.concurrencyPolicy, timeoutSeconds: task.timeoutSeconds,
    maxRetries: task.maxRetries, retrySeconds: parseJSON(task.retryBackoffJson, { seconds: 30 }).seconds || 30,
  } : {
    code: '', name: '', description: '', handlerType: availableHandlers.value[0]?.type || 'go',
    cronExpression: '0 * * * *', cronPrecision: 'minute', timezone: 'Asia/Shanghai', enabled: true,
    misfirePolicy: 'skip', maxCatchUp: 1, concurrencyPolicy: 'skip', timeoutSeconds: 300, maxRetries: 0, retrySeconds: 30,
  })
  resetHandlerConfig(task ? parseJSON(task.handlerConfigJson, {}) : undefined)
}

function resetHandlerConfig(existing?: Record<string, any>) {
  clearObject(handlerConfig)
  Object.assign(jsonFields, { params: '{}', variables: '{}', formData: '{}', query: '{}', headers: '{}', body: '{}', expectedStatus: '', args: '[]', env: '{}', parameters: '[]' })
  const defaults: Record<HandlerType, Record<string, any>> = {
    go: { handlerKey: schemaEnum('go', 'handlerKey')[0] || '' },
    workflow: { definitionId: undefined, starterIds: [], versionPolicy: 'latest_published', fixedVersion: 1 },
    http: { method: 'POST', url: '', timeoutMillis: 10000, credentialRef: '' },
    shell: { commandKey: schemaEnum('shell', 'commandKey')[0] || '' },
    sql: { dataSourceKey: '', mode: canSQLRead.value ? 'read' : 'write', statement: '', procedureKey: '', maxRows: 1000, maxAffected: 1000 },
  }
  const value = existing || defaults[form.handlerType]
  Object.assign(handlerConfig, defaults[form.handlerType], value)
  if (form.handlerType === 'workflow') {
    const definitionID = Number(handlerConfig.definitionId)
    handlerConfig.definitionId = Number.isInteger(definitionID) && definitionID > 0 ? definitionID : undefined
  }
  jsonFields.params = formatJSON(value.params ?? {})
  jsonFields.variables = formatJSON(value.variables ?? {})
  jsonFields.formData = formatJSON(value.formData ?? {})
  jsonFields.query = formatJSON(value.query ?? {})
  jsonFields.headers = formatJSON(value.headers ?? {})
  jsonFields.body = formatJSON(value.body ?? {})
  jsonFields.expectedStatus = Array.isArray(value.expectedStatus) ? value.expectedStatus.join(',') : ''
  jsonFields.args = formatJSON(value.args ?? [])
  jsonFields.env = formatJSON(value.env ?? {})
  jsonFields.parameters = formatJSON(value.parameters ?? [])
}

async function handleHandlerTypeChange() {
  resetHandlerConfig()
  if (form.handlerType === 'workflow') await loadWorkflowOptions()
}

async function loadWorkflowOptions() {
  if (workflowOptionsLoaded.value) {
    handleWorkflowDefinitionChange()
    return
  }
  workflowOptionsLoading.value = true
  try {
    const [definitionsResponse, usersResponse, departmentsResponse] = await Promise.all([
      adminApi.workflowPublishedDefinitionList({ page: 1, pageSize: 100 }),
      adminApi.workflowUserOptions({ page: 1, pageSize: 9999 }),
      adminApi.workflowDepartmentOptions(),
    ])
    workflowDefinitions.value = Array.isArray(definitionsResponse.data) ? definitionsResponse.data as WorkflowPublishedDefinition[] : []
    workflowUsers.value = Array.isArray(usersResponse.data?.list) ? usersResponse.data.list as WorkflowAssigneeUser[] : []
    workflowDepartments.value = Array.isArray(departmentsResponse.data) ? departmentsResponse.data as DepartmentNode[] : []
    workflowOptionsLoaded.value = true
    handleWorkflowDefinitionChange()
  } finally {
    workflowOptionsLoading.value = false
  }
}

function handleWorkflowDefinitionChange() {
  const eligibleUserIDs = new Set(eligibleWorkflowUsers.value.map(user => Number(user.id)))
  workflowStarterIds.value = workflowStarterIds.value.filter(userID => eligibleUserIDs.has(userID))
}

function normalizeIDs(values: Array<number | string>) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value > 0)))
}

function schemaEnum(type: HandlerType, property: string) {
  return props.handlers.find(item => item.type === type)?.configSchema?.properties?.[property]?.enum || []
}

function riskMeta(level: HandlerMetadata['riskLevel']) {
  if (level === 'critical') return { label: '极高风险', type: 'danger' as const }
  if (level === 'high') return { label: '高风险', type: 'warning' as const }
  if (level === 'medium') return { label: '中风险', type: 'primary' as const }
  return { label: '低风险', type: 'success' as const }
}

function parseJSON<T>(text: string, fallback: T): T {
  try { return JSON.parse(text) as T } catch { return fallback }
}

function formatJSON(value: unknown) {
  return JSON.stringify(value, null, 2)
}

function requiredJSON<T>(label: string, text: string, type: 'object' | 'array'): T {
  let value: unknown
  try { value = JSON.parse(text || (type === 'array' ? '[]' : '{}')) } catch { throw new Error(`${label}不是有效 JSON`) }
  if (type === 'array' ? !Array.isArray(value) : !value || Array.isArray(value) || typeof value !== 'object') {
    throw new Error(`${label}必须是 JSON ${type === 'array' ? '数组' : '对象'}`)
  }
  return value as T
}

function buildHandlerConfig() {
  const value = { ...handlerConfig }
  if (form.handlerType === 'go') value.params = requiredJSON('任务参数', jsonFields.params, 'object')
  if (form.handlerType === 'workflow') {
    value.starterIds = workflowStarterIds.value
    delete value.starterId
    value.variables = requiredJSON('流程变量', jsonFields.variables, 'object')
    value.formData = requiredJSON('初始表单', jsonFields.formData, 'object')
  }
  if (form.handlerType === 'http') {
    value.query = requiredJSON('Query', jsonFields.query, 'object')
    value.headers = requiredJSON('请求头', jsonFields.headers, 'object')
    value.body = requiredJSON('请求 Body', jsonFields.body, 'object')
    value.expectedStatus = jsonFields.expectedStatus.split(',').map(item => Number(item.trim())).filter(Number.isInteger)
  }
  if (form.handlerType === 'shell') {
    value.args = requiredJSON('参数', jsonFields.args, 'array')
    value.env = requiredJSON('环境变量', jsonFields.env, 'object')
  }
  if (form.handlerType === 'sql') value.parameters = requiredJSON('SQL 参数', jsonFields.parameters, 'array')
  return value
}

async function cronPreview() {
  if (!form.cronExpression.trim()) return ElMessage.warning('请先填写 Cron 表达式')
  previewing.value = true
  try {
    const response = await adminApi.scheduledTaskCronPreview({ expression: form.cronExpression.trim(), precision: form.cronPrecision, timezone: form.timezone, count: 5 })
    preview.value = Array.isArray(response.data?.occurrences) ? response.data.occurrences : []
  } finally { previewing.value = false }
}

async function save() {
  if (!form.name.trim() || !form.code.trim() || !form.cronExpression.trim()) return ElMessage.warning('任务名称、编码和 Cron 表达式不能为空')
  if (form.handlerType === 'workflow' && !selectedWorkflowDefinition.value) return ElMessage.warning('请选择当前已发布的流程定义')
  if (form.handlerType === 'workflow' && workflowStarterIds.value.length === 0) return ElMessage.warning('请选择流程发起人')
  if (form.handlerType === 'workflow' && workflowStarterIds.value.length > 100) return ElMessage.warning('流程发起人最多选择 100 人')
  if (form.handlerType === 'sql' && handlerConfig.mode === 'write' && !canSQLWrite.value) return ElMessage.error('缺少 SQL 写入任务权限')
  let config: Record<string, unknown>
  try { config = buildHandlerConfig() } catch (error) { return ElMessage.error((error as Error).message) }
  const payload: ScheduledTaskPayload = {
    code: form.code.trim(), name: form.name.trim(), description: form.description.trim(), handlerType: form.handlerType,
    handlerConfigJson: config, cronExpression: form.cronExpression.trim(), cronPrecision: form.cronPrecision,
    timezone: form.timezone, enabled: form.enabled, misfirePolicy: form.misfirePolicy,
    maxCatchUp: form.maxCatchUp, concurrencyPolicy: form.concurrencyPolicy,
    timeoutSeconds: form.timeoutSeconds, maxRetries: form.maxRetries,
    retryBackoffJson: { type: 'fixed', seconds: form.retrySeconds },
  }
  saving.value = true
  try {
    if (props.task) await adminApi.scheduledTaskUpdate(props.task.id, { ...payload, version: props.task.version })
    else await adminApi.scheduledTaskCreate(payload)
    ElMessage.success(props.task ? '定时任务已更新' : '定时任务已创建')
    emit('update:modelValue', false)
    emit('saved')
  } finally { saving.value = false }
}
</script>

<style scoped>
:global(.task-editor-dialog .el-dialog__body) { max-height: calc(90vh - 140px); overflow-y: auto; padding-top: 8px; }
.editor-section { padding: 0 0 20px; margin-bottom: 20px; border-bottom: 1px solid var(--el-border-color-lighter); }
.editor-section:last-child { border-bottom: 0; }
.editor-section__title { margin-bottom: 16px; font-size: 15px; font-weight: 650; color: var(--admin-text); }
.editor-section__title--actions { display: flex; align-items: center; justify-content: space-between; }
.editor-grid { display: grid; gap: 0 14px; }
.editor-grid--two { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.editor-grid--three { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.editor-grid--method { grid-template-columns: 140px minmax(0, 1fr); }
.editor-grid--cron { grid-template-columns: 160px minmax(220px, 1fr) 180px; }
.handler-option { display: flex; width: 100%; align-items: center; justify-content: space-between; gap: 12px; }
.cron-preview { display: grid; gap: 6px; padding: 10px 12px; margin: -4px 0 16px; background: #f7f8fa; border: 1px solid var(--admin-border); border-radius: 6px; color: var(--admin-muted); font-size: 13px; }
.el-alert { margin-bottom: 16px; }
@media (max-width: 720px) {
  :global(.task-editor-dialog .el-dialog__body) { padding: 8px 16px 12px; }
  .editor-grid--two, .editor-grid--three, .editor-grid--method, .editor-grid--cron { grid-template-columns: 1fr; }
}
</style>
