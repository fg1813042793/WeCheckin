<script setup lang="ts">
import type { WorkflowHistoryImportMapping } from '../workflow-history-import'
import type {
  WorkflowFieldAccessMap,
  WorkflowFormData,
  WorkflowInstanceDetail,
  WorkflowInstanceSummary,
  WorkflowPublishedDefinition,
} from '@/types/workflow'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getWorkflowInstance, listWorkflowInstances } from '@/api/workflow'
import { workflowDataFields, workflowFieldAccessMap } from '../workflow-form'
import {
  buildWorkflowImportMappings,
  isWorkflowImportTypeCompatible,
  previewWorkflowHistoryImport,
  suggestWorkflowImportMappings,
} from '../workflow-history-import'
import { workflowInstanceStatusMeta } from '../workflow-status'

type ImportMode = 'all' | 'mapping'
type HistoryImportStep = 'select' | 'configure' | 'preview'

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

const mobileBatchSize = 10
const mobileVisibleRows = 3
const desktopPageSize = 6
const popupVisible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value),
})
const mobile = ref(false)
const pageSize = computed(() => mobile.value ? mobileBatchSize : desktopPageSize)
const initialDefinitionVersion = ref(0)
const listLoading = ref(false)
const mobileLoadingMore = ref(false)
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
const currentStep = ref<HistoryImportStep>('select')
const importMode = ref<ImportMode>('all')
const manualMappings = ref<WorkflowHistoryImportMapping[]>([])
const overwrite = ref(false)
const dialogScrollTop = ref(0)
let searchFocusTimer: ReturnType<typeof setTimeout> | null = null
let keyboardHeightHandler: ((result: { height: number }) => void) | null = null
let viewportResizeHandler: (() => void) | null = null

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
const mobileListStyle = computed(() => mobile.value
  ? { '--mobile-history-visible-rows': mobileVisibleRows }
  : {})
const sourceAccess = computed<WorkflowFieldAccessMap>(() => {
  if (!sourceDetail.value)
    return {}
  return workflowFieldAccessMap(
    sourceDetail.value.form || [],
    sourceDetail.value.fieldPermissions?.[sourceDetail.value.startNodeId || 'start'] || [],
    'write',
  )
})
const stepNumber = computed(() => ({ select: 1, configure: 2, preview: 3 })[currentStep.value])
const stepTitle = computed(() => ({
  select: '选择历史申请',
  configure: '配置导入方式',
  preview: '导入预览',
})[currentStep.value])
const sourceFields = computed(() => workflowDataFields(sourceDetail.value?.form || [])
  .filter(field => field.type !== 'calculation' && sourceAccess.value[field.key] === 'write'))
const targetFields = computed(() => workflowDataFields(props.definition.form || [])
  .filter(field => field.type !== 'calculation' && props.targetAccess[field.key] === 'write'))
const allMappings = computed(() => buildWorkflowImportMappings(
  sourceFields.value,
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
        sourceFields: sourceFields.value,
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
const canGoToPreview = computed(() => Boolean(
  sourceDetail.value
  && !detailLoading.value
  && !previewError.value
  && effectiveMappings.value.length > 0,
))

function mappingOptionLabel(field: { label?: string, key: string }) {
  const label = String(field.label || field.key)
  return mobile.value ? label : `${label}（${field.key}）`
}

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

onMounted(() => {
  keyboardHeightHandler = (result) => {
    if (result.height > 0)
      scheduleSearchVisibilityReset(60)
  }
  if (typeof uni.onKeyboardHeightChange === 'function')
    uni.onKeyboardHeightChange(keyboardHeightHandler)

  // #ifdef H5
  if (typeof window !== 'undefined' && window.visualViewport) {
    viewportResizeHandler = () => {
      const viewport = window.visualViewport
      if (viewport && window.innerHeight - viewport.height > 120)
        scheduleSearchVisibilityReset(60)
    }
    window.visualViewport.addEventListener('resize', viewportResizeHandler)
  }
  // #endif
})

onBeforeUnmount(() => {
  if (searchFocusTimer) {
    clearTimeout(searchFocusTimer)
    searchFocusTimer = null
  }
  if (keyboardHeightHandler && typeof uni.offKeyboardHeightChange === 'function')
    uni.offKeyboardHeightChange(keyboardHeightHandler)

  // #ifdef H5
  if (viewportResizeHandler && typeof window !== 'undefined' && window.visualViewport)
    window.visualViewport.removeEventListener('resize', viewportResizeHandler)
  // #endif
})

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
  currentStep.value = 'select'
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

async function loadInstances(append = false) {
  if (listLoading.value || mobileLoadingMore.value)
    return
  if (append)
    mobileLoadingMore.value = true
  else
    listLoading.value = true
  if (!append)
    listError.value = ''
  try {
    const response = await listWorkflowInstances({
      definitionId: props.definition.id,
      scope: 'started',
      instanceTitle: instanceTitle.value.trim() || undefined,
      status: statusFilter.value || undefined,
      page: page.value,
      pageSize: pageSize.value,
    })
    const nextInstances = Array.isArray(response?.data?.list) ? response.data.list : []
    instances.value = append
      ? [...instances.value, ...nextInstances.filter(next => !instances.value.some(current => current.id === next.id))]
      : nextInstances
    total.value = Number(response?.data?.total || 0)
    if (!append && !instances.value.some(item => item.id === selectedInstanceId.value)) {
      selectedInstanceId.value = ''
      sourceDetail.value = null
      manualMappings.value = []
    }
  }
  catch (error) {
    if (append) {
      page.value = Math.max(1, page.value - 1)
      uni.showToast({ title: requestErrorMessage(error, '更多历史申请加载失败'), icon: 'none' })
    }
    else {
      instances.value = []
      total.value = 0
      listError.value = requestErrorMessage(error, '历史申请加载失败')
    }
  }
  finally {
    if (append)
      mobileLoadingMore.value = false
    else
      listLoading.value = false
  }
}

function loadMoreInstances() {
  if (!mobile.value || listLoading.value || mobileLoadingMore.value || instances.value.length >= total.value)
    return
  page.value += 1
  void loadInstances(true)
}

function queryInstances() {
  page.value = 1
  void loadInstances()
}

function submitInstanceSearch() {
  if (mobile.value)
    uni.hideKeyboard()
  queryInstances()
}

function keepInstanceSearchVisible() {
  if (!mobile.value)
    return
  resetDialogScrollTop()
  scheduleSearchVisibilityReset(280)
}

function resetDialogScrollTop() {
  dialogScrollTop.value = 1
  nextTick(() => {
    dialogScrollTop.value = 0
  })
}

function scheduleSearchVisibilityReset(delay: number) {
  if (!mobile.value || !popupVisible.value || currentStep.value !== 'select')
    return
  if (searchFocusTimer)
    clearTimeout(searchFocusTimer)
  searchFocusTimer = setTimeout(() => {
    resetDialogScrollTop()
    searchFocusTimer = null
  }, delay)
}

function clearInstanceSearch() {
  if (mobile.value)
    uni.hideKeyboard()
  instanceTitle.value = ''
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
      sourceFields.value,
      props.definition.form || [],
      props.targetAccess,
    )
    currentStep.value = 'configure'
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
      sourceFields.value,
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

function closeDialog() {
  popupVisible.value = false
}

function goToPreviousStep() {
  currentStep.value = currentStep.value === 'preview' ? 'configure' : 'select'
}

function goToPreview() {
  if (!canGoToPreview.value)
    return
  currentStep.value = 'preview'
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
        <view class="workflow-history-import-dialog__heading">
          <text class="workflow-history-import-dialog__title">
            导入历史表单
          </text>
          <view class="workflow-history-import-dialog__step-meta">
            <text class="workflow-history-import-dialog__subtitle">
              {{ stepTitle }}
            </text>
            <text class="workflow-history-import-dialog__step-number">
              步骤 {{ stepNumber }}/3
            </text>
          </view>
        </view>
        <view
          class="workflow-history-import-dialog__close"
          role="button"
          tabindex="0"
          title="关闭"
          aria-label="关闭"
          @click="closeDialog"
          @keydown.enter.prevent="closeDialog"
          @keydown.space.prevent="closeDialog"
        >
          <u-icon name="close" size="16px" color="#5f6b7a" />
        </view>
      </view>

      <scroll-view scroll-y class="workflow-history-import-dialog__body" :scroll-top="dialogScrollTop">
        <section v-if="currentStep === 'select'" class="workflow-history-import-dialog__section workflow-history-import-dialog__step-page">
          <view class="workflow-history-import-dialog__section-head">
            <text class="workflow-history-import-dialog__section-title">
              仅显示当前流程中由我发起的历史申请
            </text>
            <text class="workflow-history-import-dialog__section-note">
              共 {{ total }} 条
            </text>
          </view>
          <view class="workflow-history-import-dialog__filters">
            <view class="workflow-history-import-dialog__search">
              <u-search
                v-model="instanceTitle"
                shape="square"
                bg-color="#ffffff"
                border-color="#dcdfe6"
                placeholder="搜索单据标题"
                :adjust-position="false"
                :show-action="false"
                :animation="false"
                :disabled="listLoading || detailLoading"
                @search="submitInstanceSearch"
                @custom="submitInstanceSearch"
                @clear="clearInstanceSearch"
                @focus="keepInstanceSearchVisible"
              />
            </view>
            <view class="workflow-history-import-dialog__filter-secondary">
              <select v-model="statusFilter" class="workflow-history-import-dialog__select" aria-label="审批状态">
                <option v-for="option in statusOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
              <u-button
                custom-class="workflow-history-import-dialog__filter-button workflow-history-import-dialog__filter-button--search"
                type="primary"
                size="small"
                :loading="listLoading"
                @click="submitInstanceSearch"
              >
                搜索
              </u-button>
              <u-button custom-class="workflow-history-import-dialog__filter-button" plain size="small" :disabled="listLoading" @click="resetFilters">
                重置
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
          <view v-else-if="detailLoading" class="workflow-history-import-dialog__state workflow-history-import-dialog__state--compact">
            <u-loading mode="circle" color="#0f766e" size="28" />
            <text>正在读取历史表单...</text>
          </view>
          <view v-else-if="detailError" class="workflow-history-import-dialog__state workflow-history-import-dialog__state--error workflow-history-import-dialog__state--compact">
            <u-icon name="error-circle" size="30" color="#ef4444" />
            <text>{{ detailError }}</text>
            <u-button plain size="small" @click="retryDetail">
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
          <scroll-view
            v-else
            :scroll-y="mobile"
            class="workflow-history-import-dialog__instances-scroll"
            :class="{ 'workflow-history-import-dialog__instances-scroll--mobile': mobile }"
            :style="mobileListStyle"
            :lower-threshold="40"
            @scrolltolower="loadMoreInstances"
          >
            <view class="workflow-history-import-dialog__instances">
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
              <view v-if="mobileLoadingMore" class="workflow-history-import-dialog__load-more">
                <u-loading mode="circle" color="#0f766e" size="20" />
                <text>正在加载更多...</text>
              </view>
            </view>
          </scroll-view>
          <view v-if="!mobile && total > pageSize" class="workflow-history-import-dialog__pagination">
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

        <section v-else-if="currentStep === 'configure'" class="workflow-history-import-dialog__section workflow-history-import-dialog__step-page">
          <view class="workflow-history-import-dialog__section-head">
            <text class="workflow-history-import-dialog__section-title">
              {{ sourceDetail?.instance.instanceTitle || sourceDetail?.instance.businessKey || '历史申请' }}
            </text>
            <text v-if="sourceDetail" class="workflow-history-import-dialog__section-note">
              来源版本 v{{ sourceDetail.instance.definitionVersion }}
            </text>
          </view>
          <view v-if="!sourceDetail" class="workflow-history-import-dialog__state workflow-history-import-dialog__state--compact">
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
                    {{ mappingOptionLabel(sourceField) }}
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
                    {{ mappingOptionLabel(targetField) }}
                  </option>
                </select>
                <view
                  class="workflow-history-import-dialog__mapping-remove"
                  role="button"
                  tabindex="0"
                  title="删除映射"
                  aria-label="删除映射"
                  @click="removeMapping(index)"
                  @keydown.enter.prevent="removeMapping(index)"
                  @keydown.space.prevent="removeMapping(index)"
                >
                  <u-icon name="trash" size="15px" color="#ef4444" />
                </view>
              </view>
              <u-button plain size="small" icon="plus" @click="addMapping">
                增加映射
              </u-button>
            </view>
          </template>
        </section>

        <section v-else class="workflow-history-import-dialog__section workflow-history-import-dialog__step-page">
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

      <view
        class="workflow-history-import-dialog__actions"
        :class="{ 'workflow-history-import-dialog__actions--mobile-hidden': mobile && currentStep === 'select' }"
      >
        <template v-if="currentStep === 'select'">
          <u-button v-if="!mobile" class="workflow-history-import-dialog__cancel" plain :disabled="detailLoading" @click="closeDialog">
            取消
          </u-button>
        </template>
        <template v-else-if="currentStep === 'configure'">
          <u-button plain :disabled="detailLoading" @click="goToPreviousStep">
            上一步
          </u-button>
          <u-button type="primary" :disabled="!canGoToPreview" @click="goToPreview">
            下一步
          </u-button>
        </template>
        <template v-else>
          <u-button plain @click="goToPreviousStep">
            上一步
          </u-button>
          <u-button type="primary" :disabled="!canApply" @click="applyImport">
            确认导入
          </u-button>
        </template>
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

.workflow-history-import-dialog__actions--mobile-hidden {
  display: none;
}

.workflow-history-import-dialog__header {
  min-height: 68px;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 18px;
  border-bottom: 1px solid #e5eaf3;
}

.workflow-history-import-dialog__heading {
  min-width: 0;
}

.workflow-history-import-dialog__step-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.workflow-history-import-dialog__step-number {
  color: #98a2b3;
  font-size: 12px;
}

.workflow-history-import-dialog__close {
  width: 32px;
  height: 32px;
  flex: 0 0 32px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  cursor: pointer;
  transition: background-color 0.15s ease, box-shadow 0.15s ease;
}

.workflow-history-import-dialog__close:hover {
  background: #f2f4f7;
}

.workflow-history-import-dialog__close:focus-visible {
  outline: none;
  background: #f2f4f7;
  box-shadow: 0 0 0 2px rgb(15 118 110 / 22%);
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
  color: #87909f;
  font-size: 13px;
}

.workflow-history-import-dialog__body {
  flex: 1 1 auto;
  min-height: 0;
  padding: 14px 18px;
  box-sizing: border-box;
}

.workflow-history-import-dialog__step-page {
  min-height: 100%;
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

.workflow-history-import-dialog__search {
  min-width: 0;
  flex: 1 1 auto;
}

.workflow-history-import-dialog__search :deep(.u-search) {
  width: 100%;
  min-height: 36px;
}

.workflow-history-import-dialog__search :deep(.u-content) {
  height: 36px !important;
  border-radius: 4px !important;
  box-sizing: border-box;
}

.workflow-history-import-dialog__search :deep(.u-input) {
  font-size: 13px !important;
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

.workflow-history-import-dialog__filter-secondary {
  flex: 0 0 auto;
  display: grid;
  grid-template-columns: 140px 64px 64px;
  align-items: center;
  gap: 8px;
}

:deep(.workflow-history-import-dialog__filter-button) {
  width: 100%;
  margin: 0;
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

.workflow-history-import-dialog__instances-scroll {
  margin-top: 12px;
}

.workflow-history-import-dialog__instances-scroll .workflow-history-import-dialog__instances {
  margin-top: 0;
}

.workflow-history-import-dialog__load-more {
  min-height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #87909f;
  font-size: 12px;
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
  grid-template-columns: minmax(0, 1fr) 20px minmax(0, 1fr) 32px;
  gap: 8px;
}

.workflow-history-import-dialog__mapping-remove {
  width: 32px;
  height: 32px;
  align-self: center;
  border: 1px solid #fda4a4;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  background: #fff;
  cursor: pointer;
}

.workflow-history-import-dialog__mapping-remove:hover,
.workflow-history-import-dialog__mapping-remove:focus-visible {
  outline: none;
  border-color: #ef4444;
  background: #fff1f2;
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

  .workflow-history-import-dialog__filters {
    position: sticky;
    z-index: 5;
    top: 0;
    padding: 8px 0 10px;
    background: #f6f8fb;
    box-shadow: 0 6px 10px rgb(15 23 42 / 4%);
  }

  .workflow-history-import-dialog__filter-secondary {
    width: 100%;
    grid-template-columns: minmax(0, 1fr) 64px 64px;
  }

  .workflow-history-import-dialog__instance-meta {
    flex-wrap: wrap;
  }

  .workflow-history-import-dialog__instances-scroll--mobile {
    --mobile-history-row-height: 116px;

    height: calc(
      var(--mobile-history-row-height)
      * var(--mobile-history-visible-rows)
      + 16px
    );
  }

  .workflow-history-import-dialog__instances-scroll--mobile .workflow-history-import-dialog__instance {
    height: var(--mobile-history-row-height);
    min-height: var(--mobile-history-row-height);
    box-sizing: border-box;
  }

  .workflow-history-import-dialog__segmented {
    width: 100%;
    box-sizing: border-box;
  }

  .workflow-history-import-dialog__overwrite {
    justify-content: space-between;
  }

  .workflow-history-import-dialog__mapping-row {
    grid-template-columns: minmax(0, 1fr) 18px minmax(0, 1fr) 32px;
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
