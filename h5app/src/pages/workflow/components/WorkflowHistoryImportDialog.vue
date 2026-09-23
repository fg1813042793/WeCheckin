<script setup lang="ts">
import type { WorkflowHistoryImportMapping } from '../workflow-history-import'
import type {
  WorkflowFieldAccessMap,
  WorkflowFormData,
  WorkflowInstanceDetail,
  WorkflowInstanceSummary,
  WorkflowPublishedDefinition,
} from '@/types/workflow'
import { computed, ref, watch } from 'vue'
import { getWorkflowInstance, listWorkflowInstances } from '@/api/workflow'
import { workflowDataFields } from '../workflow-form'
import {
  buildWorkflowImportMappings,
  isWorkflowImportTypeCompatible,
  previewWorkflowHistoryImport,
  suggestWorkflowImportMappings,
} from '../workflow-history-import'
import { workflowInstanceStatusMeta } from '../workflow-status'

type ImportMode = 'all' | 'mapping'

interface PaginationChangePayload {
  current: number
}

interface SelectChangeEvent {
  detail?: { value?: string }
  target?: { value?: string }
}

const props = defineProps<{
  modelValue: boolean
  definition: WorkflowPublishedDefinition
  targetData: WorkflowFormData
  targetAccess: WorkflowFieldAccessMap
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'apply': [payload: { sourceInstanceId: string, patch: WorkflowFormData, importCount: number }]
}>()

const pageSize = 6
const popupVisible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value),
})
const mobile = ref(false)
const initialDefinitionVersion = ref(0)
const listLoading = ref(false)
const listError = ref('')
const instances = ref<WorkflowInstanceSummary[]>([])
const page = ref(1)
const total = ref(0)
const instanceTitle = ref('')
const statusFilter = ref('')
const selectedInstanceId = ref('')
const detailLoading = ref(false)
const detailError = ref('')
const sourceDetail = ref<WorkflowInstanceDetail | null>(null)
const importMode = ref<ImportMode>('all')
const manualMappings = ref<WorkflowHistoryImportMapping[]>([])
const overwrite = ref(false)

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '审批中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '已驳回', value: 'rejected' },
  { label: '已撤回', value: 'withdrawn' },
  { label: '已取消', value: 'cancelled' },
]

const popupMode = computed(() => mobile.value ? 'bottom' : 'center')
const popupWidth = computed(() => mobile.value ? '100%' : '760px')
const popupHeight = computed(() => mobile.value ? '92vh' : '86vh')
const sourceFields = computed(() => workflowDataFields(sourceDetail.value?.form || [])
  .filter(field => field.type !== 'calculation'))
const targetFields = computed(() => workflowDataFields(props.definition.form || [])
  .filter(field => field.type !== 'calculation' && props.targetAccess[field.key] === 'write'))
const allMappings = computed(() => buildWorkflowImportMappings(
  sourceDetail.value?.form || [],
  props.definition.form || [],
  props.targetAccess,
))
const effectiveMappings = computed(() => importMode.value === 'all'
  ? allMappings.value
  : manualMappings.value.filter(mapping => mapping.sourceKey && mapping.targetKey))
const previewState = computed(() => {
  if (!sourceDetail.value) {
    return {
      preview: emptyPreview(),
      error: '',
    }
  }
  try {
    return {
      preview: previewWorkflowHistoryImport({
        sourceFields: sourceDetail.value.form || [],
        sourceData: sourceDetail.value.formData || {},
        targetFields: props.definition.form || [],
        targetData: props.targetData,
        targetAccess: props.targetAccess,
        mappings: effectiveMappings.value,
        overwrite: overwrite.value,
      }),
      error: '',
    }
  }
  catch (error) {
    return {
      preview: emptyPreview(),
      error: requestErrorMessage(error, '字段映射无效'),
    }
  }
})
const preview = computed(() => previewState.value.preview)
const previewError = computed(() => previewState.value.error)
const applicableCount = computed(() => preview.value.importCount + preview.value.overwriteCount)
const canApply = computed(() => Boolean(
  sourceDetail.value
  && !detailLoading.value
  && !previewError.value
  && applicableCount.value > 0,
))

function emptyPreview() {
  return {
    items: [],
    patch: {},
    importCount: 0,
    overwriteCount: 0,
    retainCount: 0,
    skipCount: 0,
  }
}

watch(
  () => props.modelValue,
  (visible) => {
    if (visible)
      void openDialog()
  },
)

watch(
  () => props.definition.version,
  (version) => {
    if (!props.modelValue || !initialDefinitionVersion.value || version === initialDefinitionVersion.value)
      return
    popupVisible.value = false
    uni.showToast({ title: '流程版本已更新，请重新打开导入', icon: 'none' })
  },
)

async function openDialog() {
  mobile.value = resolveMobilePage()
  initialDefinitionVersion.value = props.definition.version
  page.value = 1
  total.value = 0
  instanceTitle.value = ''
  statusFilter.value = ''
  selectedInstanceId.value = ''
  sourceDetail.value = null
  detailError.value = ''
  importMode.value = 'all'
  manualMappings.value = []
  overwrite.value = false
  await loadInstances()
}

async function loadInstances() {
  if (listLoading.value)
    return
  listLoading.value = true
  listError.value = ''
  try {
    const response = await listWorkflowInstances({
      definitionId: props.definition.id,
      scope: 'started',
      instanceTitle: instanceTitle.value.trim() || undefined,
      status: statusFilter.value || undefined,
      page: page.value,
      pageSize,
    })
    instances.value = Array.isArray(response?.data?.list) ? response.data.list : []
    total.value = Number(response?.data?.total || 0)
    if (!instances.value.some(item => item.id === selectedInstanceId.value)) {
      selectedInstanceId.value = ''
      sourceDetail.value = null
      manualMappings.value = []
    }
  }
  catch (error) {
    instances.value = []
    total.value = 0
    listError.value = requestErrorMessage(error, '历史申请加载失败')
  }
  finally {
    listLoading.value = false
  }
}

function queryInstances() {
  page.value = 1
  void loadInstances()
}

function resetFilters() {
  instanceTitle.value = ''
  statusFilter.value = ''
  page.value = 1
  void loadInstances()
}

function handlePageChange(payload: PaginationChangePayload) {
  page.value = Number(payload.current || 1)
  void loadInstances()
}

async function selectInstance(instance: WorkflowInstanceSummary) {
  if (detailLoading.value)
    return
  selectedInstanceId.value = instance.id
  sourceDetail.value = null
  detailError.value = ''
  detailLoading.value = true
  try {
    const response = await getWorkflowInstance(instance.id)
    const detail = response?.data || null
    if (!detail || detail.instance.definitionId !== props.definition.id)
      throw new Error('历史申请不属于当前流程')
    sourceDetail.value = detail
    importMode.value = 'all'
    manualMappings.value = suggestWorkflowImportMappings(
      detail.form || [],
      props.definition.form || [],
      props.targetAccess,
    )
  }
  catch (error) {
    detailError.value = requestErrorMessage(error, '历史申请详情加载失败')
  }
  finally {
    detailLoading.value = false
  }
}

function retryDetail() {
  const instance = instances.value.find(item => item.id === selectedInstanceId.value)
  if (instance)
    void selectInstance(instance)
}

function setImportMode(mode: ImportMode) {
  importMode.value = mode
  if (mode === 'mapping' && sourceDetail.value && manualMappings.value.length === 0) {
    manualMappings.value = suggestWorkflowImportMappings(
      sourceDetail.value.form || [],
      props.definition.form || [],
      props.targetAccess,
    )
  }
}

function addMapping() {
  manualMappings.value.push({ sourceKey: '', targetKey: '' })
}

function removeMapping(index: number) {
  manualMappings.value.splice(index, 1)
}

function updateMapping(index: number, key: keyof WorkflowHistoryImportMapping, event: SelectChangeEvent) {
  const mapping = manualMappings.value[index]
  if (!mapping)
    return
  mapping[key] = String(event.detail?.value ?? event.target?.value ?? '')
  if (key === 'sourceKey' && mapping.targetKey) {
    const sourceField = sourceFields.value.find(field => field.key === mapping.sourceKey)
    const targetField = targetFields.value.find(field => field.key === mapping.targetKey)
    if (!sourceField || !targetField || !isWorkflowImportTypeCompatible(sourceField, targetField))
      mapping.targetKey = ''
  }
}

function compatibleTargetFields(sourceKey: string) {
  const sourceField = sourceFields.value.find(field => field.key === sourceKey)
  if (!sourceField)
    return []
  return targetFields.value.filter(targetField => isWorkflowImportTypeCompatible(sourceField, targetField))
}

function sourceDisabled(sourceKey: string, rowIndex: number) {
  return manualMappings.value.some((mapping, index) => index !== rowIndex && mapping.sourceKey === sourceKey)
}

function targetDisabled(targetKey: string, rowIndex: number) {
  return manualMappings.value.some((mapping, index) => index !== rowIndex && mapping.targetKey === targetKey)
}

function applyImport() {
  if (!canApply.value || !sourceDetail.value)
    return
  emit('apply', {
    sourceInstanceId: sourceDetail.value.instance.id,
    patch: preview.value.patch,
    importCount: applicableCount.value,
  })
  popupVisible.value = false
}

function formatTime(timestamp?: number) {
  if (!timestamp)
    return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function previewValue(value: unknown) {
  if (value === undefined || value === null || value === '')
    return '暂无填写'
  if (typeof value === 'string')
    return value.length > 48 ? `${value.slice(0, 48)}...` : value
  if (typeof value === 'number' || typeof value === 'boolean')
    return String(value)
  try {
    const text = JSON.stringify(value)
    return text.length > 48 ? `${text.slice(0, 48)}...` : text
  }
  catch {
    return '复杂内容'
  }
}

function previewResultLabel(result: string) {
  return {
    import: '将导入',
    overwrite: '将覆盖',
    retain: '保留现有',
    incompatible: '类型不兼容',
    filtered: '选项已过滤',
    dynamic_option: '动态选项待确认',
    skipped: '已跳过',
  }[result] || '已跳过'
}

function previewResultType(result: string): 'success' | 'warning' | 'error' | 'info' {
  if (result === 'import' || result === 'overwrite')
    return 'success'
  if (result === 'filtered' || result === 'dynamic_option' || result === 'retain')
    return 'warning'
  if (result === 'incompatible')
    return 'error'
  return 'info'
}

function requestErrorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message.trim())
    return error.message.trim()
  if (!error || typeof error !== 'object')
    return fallback
  const response = error as Record<string, unknown>
  const payload = response.data && typeof response.data === 'object' && !Array.isArray(response.data)
    ? response.data as Record<string, unknown>
    : response
  const message = payload.msg ?? payload.message
  return typeof message === 'string' && message.trim() ? message.trim() : fallback
}

function resolveMobilePage() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    return width > 0 && width <= 768
  }
  catch {
    return false
  }
}
</script>

<template>
  <u-popup
    v-model="popupVisible"
    :mode="popupMode"
    :width="popupWidth"
    :height="popupHeight"
    custom-class="workflow-history-import-popup app-pc-control-scope"
    :border-radius="8"
    :safe-area-inset-bottom="mobile"
    :mask-close-able="!detailLoading"
  >
    <view class="workflow-history-import-dialog" :class="{ 'workflow-history-import-dialog--mobile': mobile }">
      <view class="workflow-history-import-dialog__header">
        <view>
          <text class="workflow-history-import-dialog__title">
            导入历史表单
          </text>
          <text class="workflow-history-import-dialog__subtitle">
            仅显示当前流程中由我发起的历史申请
          </text>
        </view>
        <u-button custom-class="app-icon-button" :disabled="detailLoading" @click="popupVisible = false">
          <u-icon name="close" size="16px" color="#5f6b7a" />
        </u-button>
      </view>

      <scroll-view scroll-y class="workflow-history-import-dialog__body">
        <section class="workflow-history-import-dialog__section">
          <view class="workflow-history-import-dialog__section-head">
            <text class="workflow-history-import-dialog__section-title">
              1. 选择历史申请
            </text>
            <text class="workflow-history-import-dialog__section-note">
              共 {{ total }} 条
            </text>
          </view>
          <view class="workflow-history-import-dialog__filters">
            <u-input v-model="instanceTitle" clearable placeholder="搜索单据标题" @confirm="queryInstances" />
            <select v-model="statusFilter" class="workflow-history-import-dialog__select" aria-label="审批状态">
              <option v-for="option in statusOptions" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
            <view class="workflow-history-import-dialog__filter-actions">
              <u-button plain size="small" :disabled="listLoading" @click="resetFilters">
                重置
              </u-button>
              <u-button type="primary" size="small" :loading="listLoading" @click="queryInstances">
                查询
              </u-button>
            </view>
          </view>

          <view v-if="listLoading" class="workflow-history-import-dialog__state">
            <u-loading mode="circle" color="#0f766e" size="28" />
            <text>正在加载历史申请...</text>
          </view>
          <view v-else-if="listError" class="workflow-history-import-dialog__state workflow-history-import-dialog__state--error">
            <u-icon name="error-circle" size="30" color="#ef4444" />
            <text>{{ listError }}</text>
            <u-button plain size="small" @click="loadInstances">
              重新加载
            </u-button>
          </view>
          <view v-else-if="instances.length === 0" class="workflow-history-import-dialog__state">
            <u-icon name="order" size="34" color="#a8b0bd" />
            <text>暂无历史申请</text>
            <u-button plain size="small" @click="loadInstances">
              重新加载
            </u-button>
          </view>
          <view v-else class="workflow-history-import-dialog__instances">
            <view
              v-for="instance in instances"
              :key="instance.id"
              class="workflow-history-import-dialog__instance"
              :class="{ 'workflow-history-import-dialog__instance--selected': selectedInstanceId === instance.id }"
              @click="selectInstance(instance)"
            >
              <view class="workflow-history-import-dialog__instance-main">
                <text class="workflow-history-import-dialog__instance-title">
                  {{ instance.instanceTitle || instance.definitionName }}
                </text>
                <text class="workflow-history-import-dialog__instance-key">
                  {{ instance.businessKey || instance.id }}
                </text>
              </view>
              <view class="workflow-history-import-dialog__instance-meta">
                <text>v{{ instance.definitionVersion }}</text>
                <text>{{ formatTime(instance.startTime) }}</text>
                <text v-if="instance.endTime">
                  完成 {{ formatTime(instance.endTime) }}
                </text>
                <u-tag
                  :text="workflowInstanceStatusMeta(instance.status).label"
                  :type="workflowInstanceStatusMeta(instance.status).type"
                  size="mini"
                />
              </view>
            </view>
          </view>
          <view v-if="total > pageSize" class="workflow-history-import-dialog__pagination">
            <u-pagination
              v-model="page"
              :total="total"
              :page-size="pageSize"
              prev-text="上一页"
              next-text="下一页"
              @change="handlePageChange"
            />
          </view>
        </section>

        <section class="workflow-history-import-dialog__section">
          <view class="workflow-history-import-dialog__section-head">
            <text class="workflow-history-import-dialog__section-title">
              2. 配置导入方式
            </text>
            <text v-if="sourceDetail" class="workflow-history-import-dialog__section-note">
              来源版本 v{{ sourceDetail.instance.definitionVersion }}
            </text>
          </view>
          <view v-if="detailLoading" class="workflow-history-import-dialog__state">
            <u-loading mode="circle" color="#0f766e" size="28" />
            <text>正在读取历史表单...</text>
          </view>
          <view v-else-if="detailError" class="workflow-history-import-dialog__state workflow-history-import-dialog__state--error">
            <u-icon name="error-circle" size="30" color="#ef4444" />
            <text>{{ detailError }}</text>
            <u-button plain size="small" @click="retryDetail">
              重新加载
            </u-button>
          </view>
          <view v-else-if="!sourceDetail" class="workflow-history-import-dialog__state workflow-history-import-dialog__state--compact">
            <text>请先选择一条历史申请</text>
          </view>
          <template v-else>
            <view class="workflow-history-import-dialog__mode-row">
              <view class="workflow-history-import-dialog__segmented" role="tablist" aria-label="导入方式">
                <view
                  class="workflow-history-import-dialog__segment"
                  :class="{ 'workflow-history-import-dialog__segment--active': importMode === 'all' }"
                  role="tab"
                  :aria-selected="importMode === 'all'"
                  @click="setImportMode('all')"
                >
                  全部导入
                </view>
                <view
                  class="workflow-history-import-dialog__segment"
                  :class="{ 'workflow-history-import-dialog__segment--active': importMode === 'mapping' }"
                  role="tab"
                  :aria-selected="importMode === 'mapping'"
                  @click="setImportMode('mapping')"
                >
                  字段映射
                </view>
              </view>
              <view class="workflow-history-import-dialog__overwrite">
                <view>
                  <text class="workflow-history-import-dialog__overwrite-title">
                    覆盖已有内容
                  </text>
                  <text class="workflow-history-import-dialog__overwrite-note">
                    默认只填充空字段
                  </text>
                </view>
                <u-switch v-model="overwrite" />
              </view>
            </view>

            <view v-if="importMode === 'all'" class="workflow-history-import-dialog__all-summary">
              将按相同字段 key 导入 {{ allMappings.length }} 个兼容且可写字段。
            </view>
            <view v-else class="workflow-history-import-dialog__mappings">
              <view
                v-for="(mapping, index) in manualMappings"
                :key="`${index}-${mapping.sourceKey}-${mapping.targetKey}`"
                class="workflow-history-import-dialog__mapping-row"
              >
                <select
                  :value="mapping.sourceKey"
                  class="workflow-history-import-dialog__select"
                  aria-label="历史字段"
                  @change="updateMapping(index, 'sourceKey', $event)"
                >
                  <option value="">
                    选择历史字段
                  </option>
                  <option
                    v-for="sourceField in sourceFields"
                    :key="sourceField.key"
                    :value="sourceField.key"
                    :disabled="sourceDisabled(sourceField.key, index)"
                  >
                    {{ sourceField.label }}（{{ sourceField.key }}）
                  </option>
                </select>
                <u-icon name="arrow-right" size="18px" color="#87909f" />
                <select
                  :value="mapping.targetKey"
                  class="workflow-history-import-dialog__select"
                  aria-label="当前字段"
                  :disabled="!mapping.sourceKey"
                  @change="updateMapping(index, 'targetKey', $event)"
                >
                  <option value="">
                    选择当前字段
                  </option>
                  <option
                    v-for="targetField in compatibleTargetFields(mapping.sourceKey)"
                    :key="targetField.key"
                    :value="targetField.key"
                    :disabled="targetDisabled(targetField.key, index)"
                  >
                    {{ targetField.label }}（{{ targetField.key }}）
                  </option>
                </select>
                <u-button custom-class="app-icon-button app-icon-button--small" @click="removeMapping(index)">
                  <u-icon name="trash" size="15px" color="#ef4444" />
                </u-button>
              </view>
              <u-button plain size="small" icon="plus" @click="addMapping">
                增加映射
              </u-button>
            </view>
          </template>
        </section>

        <section class="workflow-history-import-dialog__section">
          <view class="workflow-history-import-dialog__section-head">
            <text class="workflow-history-import-dialog__section-title">
              3. 导入预览
            </text>
          </view>
          <view class="workflow-history-import-dialog__stats">
            <text>导入 {{ preview.importCount }}</text>
            <text>覆盖 {{ preview.overwriteCount }}</text>
            <text>保留 {{ preview.retainCount }}</text>
            <text>跳过 {{ preview.skipCount }}</text>
          </view>
          <text v-if="previewError" class="workflow-history-import-dialog__preview-error">
            {{ previewError }}
          </text>
          <view v-else-if="preview.items.length === 0" class="workflow-history-import-dialog__state workflow-history-import-dialog__state--compact">
            <text>选择历史申请并配置字段后可查看预览</text>
          </view>
          <view v-else class="workflow-history-import-dialog__preview-list">
            <view v-for="item in preview.items" :key="`${item.sourceKey}-${item.targetKey}`" class="workflow-history-import-dialog__preview-item">
              <view class="workflow-history-import-dialog__preview-head">
                <text class="workflow-history-import-dialog__preview-title">
                  {{ item.sourceLabel }} → {{ item.targetLabel }}
                </text>
                <u-tag :text="previewResultLabel(item.result)" :type="previewResultType(item.result)" size="mini" />
              </view>
              <text class="workflow-history-import-dialog__preview-value">
                当前：{{ previewValue(item.currentValue) }}
              </text>
              <text class="workflow-history-import-dialog__preview-value">
                导入：{{ previewValue(item.importValue) }}
              </text>
              <text class="workflow-history-import-dialog__preview-message">
                {{ item.message }}
              </text>
            </view>
          </view>
        </section>
      </scroll-view>

      <view class="workflow-history-import-dialog__actions">
        <u-button plain :disabled="detailLoading" @click="popupVisible = false">
          取消
        </u-button>
        <u-button type="primary" :disabled="!canApply" @click="applyImport">
          确认导入
        </u-button>
      </view>
    </view>
  </u-popup>
</template>

<style lang="scss" scoped>
.workflow-history-import-dialog {
  width: 100%;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f6f8fb;
  color: #1f2329;
}

.workflow-history-import-dialog__header,
.workflow-history-import-dialog__actions {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  border-color: #e5eaf3;
  background: #fff;
}

.workflow-history-import-dialog__header {
  min-height: 68px;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 18px;
  border-bottom: 1px solid #e5eaf3;
}

.workflow-history-import-dialog__title,
.workflow-history-import-dialog__subtitle {
  display: block;
}

.workflow-history-import-dialog__title {
  font-size: 18px;
  font-weight: 700;
}

.workflow-history-import-dialog__subtitle {
  margin-top: 4px;
  color: #87909f;
  font-size: 13px;
}

.workflow-history-import-dialog__body {
  flex: 1 1 auto;
  min-height: 0;
  padding: 14px 18px;
  box-sizing: border-box;
}

.workflow-history-import-dialog__section {
  padding: 14px 0;
  border-bottom: 1px solid #e5eaf3;
}

.workflow-history-import-dialog__section:last-child {
  border-bottom: 0;
}

.workflow-history-import-dialog__section-head,
.workflow-history-import-dialog__preview-head,
.workflow-history-import-dialog__instance-meta,
.workflow-history-import-dialog__mode-row,
.workflow-history-import-dialog__overwrite,
.workflow-history-import-dialog__filters,
.workflow-history-import-dialog__mapping-row,
.workflow-history-import-dialog__stats {
  display: flex;
  align-items: center;
}

.workflow-history-import-dialog__section-head,
.workflow-history-import-dialog__preview-head {
  justify-content: space-between;
  gap: 12px;
}

.workflow-history-import-dialog__section-title {
  font-size: 15px;
  font-weight: 650;
}

.workflow-history-import-dialog__section-note {
  color: #87909f;
  font-size: 12px;
}

.workflow-history-import-dialog__filters {
  gap: 8px;
  margin-top: 12px;
}

.workflow-history-import-dialog__filters :deep(.u-input) {
  flex: 1 1 auto;
}

.workflow-history-import-dialog__select {
  min-width: 0;
  height: 36px;
  padding: 0 30px 0 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  box-sizing: border-box;
  background: #fff;
  color: #303133;
  font: inherit;
}

.workflow-history-import-dialog__select:focus-visible {
  border-color: #0f766e;
  outline: 2px solid rgb(15 118 110 / 18%);
}

.workflow-history-import-dialog__filter-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(64px, auto));
  gap: 8px;
}

.workflow-history-import-dialog__state {
  min-height: 116px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 10px;
  color: #87909f;
  font-size: 13px;
}

.workflow-history-import-dialog__state--compact {
  min-height: 72px;
}

.workflow-history-import-dialog__state--error {
  color: #b42318;
}

.workflow-history-import-dialog__instances,
.workflow-history-import-dialog__preview-list,
.workflow-history-import-dialog__mappings {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
}

.workflow-history-import-dialog__instance,
.workflow-history-import-dialog__preview-item {
  border: 1px solid #e1e6ef;
  border-radius: 6px;
  background: #fff;
}

.workflow-history-import-dialog__instance {
  min-height: 58px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  box-sizing: border-box;
  cursor: pointer;
}

.workflow-history-import-dialog__instance:hover,
.workflow-history-import-dialog__instance--selected {
  border-color: #0f766e;
  background: #f0fdfa;
}

.workflow-history-import-dialog__instance-main {
  min-width: 0;
  flex: 1 1 auto;
}

.workflow-history-import-dialog__instance-title,
.workflow-history-import-dialog__instance-key {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-history-import-dialog__instance-title {
  font-size: 14px;
  font-weight: 600;
}

.workflow-history-import-dialog__instance-key,
.workflow-history-import-dialog__instance-meta,
.workflow-history-import-dialog__preview-value,
.workflow-history-import-dialog__preview-message {
  color: #87909f;
  font-size: 12px;
}

.workflow-history-import-dialog__instance-key {
  margin-top: 4px;
}

.workflow-history-import-dialog__instance-meta {
  flex: 0 0 auto;
  gap: 10px;
}

.workflow-history-import-dialog__pagination {
  display: flex;
  justify-content: center;
  margin-top: 12px;
}

.workflow-history-import-dialog__mode-row {
  justify-content: space-between;
  gap: 16px;
  margin-top: 12px;
}

.workflow-history-import-dialog__segmented {
  display: inline-grid;
  grid-template-columns: repeat(2, minmax(96px, 1fr));
  padding: 3px;
  border: 1px solid #dfe4ec;
  border-radius: 6px;
  background: #f2f4f7;
}

.workflow-history-import-dialog__segment {
  min-height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 14px;
  border-radius: 4px;
  color: #667085;
  font-size: 13px;
  cursor: pointer;
}

.workflow-history-import-dialog__segment--active {
  background: #fff;
  color: #0f766e;
  font-weight: 600;
  box-shadow: 0 1px 3px rgb(16 24 40 / 10%);
}

.workflow-history-import-dialog__overwrite {
  gap: 12px;
}

.workflow-history-import-dialog__overwrite-title,
.workflow-history-import-dialog__overwrite-note {
  display: block;
}

.workflow-history-import-dialog__overwrite-title {
  font-size: 13px;
  font-weight: 600;
}

.workflow-history-import-dialog__overwrite-note {
  margin-top: 2px;
  color: #87909f;
  font-size: 12px;
}

.workflow-history-import-dialog__all-summary {
  margin-top: 12px;
  padding: 10px 12px;
  border-left: 3px solid #0f766e;
  background: #f0fdfa;
  color: #47605e;
  font-size: 13px;
}

.workflow-history-import-dialog__mapping-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 20px minmax(0, 1fr) 30px;
  gap: 8px;
}

.workflow-history-import-dialog__stats {
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.workflow-history-import-dialog__stats text {
  padding: 4px 8px;
  border-radius: 4px;
  background: #eef2f6;
  color: #475467;
  font-size: 12px;
}

.workflow-history-import-dialog__preview-error {
  display: block;
  margin-top: 10px;
  color: #b42318;
  font-size: 13px;
}

.workflow-history-import-dialog__preview-item {
  padding: 10px 12px;
}

.workflow-history-import-dialog__preview-title {
  min-width: 0;
  overflow: hidden;
  color: #344054;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-history-import-dialog__preview-value,
.workflow-history-import-dialog__preview-message {
  display: block;
  margin-top: 5px;
  word-break: break-word;
}

.workflow-history-import-dialog__preview-message {
  color: #667085;
}

.workflow-history-import-dialog__actions {
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 18px;
  border-top: 1px solid #e5eaf3;
}

.workflow-history-import-dialog__actions :deep(.u-button) {
  width: 96px;
  margin: 0;
}

@media (max-width: 768px) {
  .workflow-history-import-dialog__header {
    padding: 12px 14px;
  }

  .workflow-history-import-dialog__body {
    padding: 10px 14px;
  }

  .workflow-history-import-dialog__filters,
  .workflow-history-import-dialog__mode-row,
  .workflow-history-import-dialog__instance {
    align-items: stretch;
    flex-direction: column;
  }

  .workflow-history-import-dialog__filter-actions {
    grid-template-columns: repeat(2, 1fr);
  }

  .workflow-history-import-dialog__instance-meta {
    flex-wrap: wrap;
  }

  .workflow-history-import-dialog__segmented {
    width: 100%;
    box-sizing: border-box;
  }

  .workflow-history-import-dialog__overwrite {
    justify-content: space-between;
  }

  .workflow-history-import-dialog__mapping-row {
    grid-template-columns: minmax(0, 1fr) 18px minmax(0, 1fr) 30px;
  }

  .workflow-history-import-dialog__mapping-row .workflow-history-import-dialog__select {
    padding-right: 20px;
  }

  .workflow-history-import-dialog__actions {
    padding: 10px 14px calc(10px + env(safe-area-inset-bottom));
  }

  .workflow-history-import-dialog__actions :deep(.u-button) {
    width: auto;
    flex: 1 1 0;
  }
}
</style>
