<script setup lang="ts">
import type { WorkflowHistoryDateFilters } from '../workflow-history-filter'
import type { WorkflowSummaryColumn } from '../workflow-summary-columns'
import type {
  WorkflowInstanceSummary,
  WorkflowPublishedDefinition,
  WorkflowSummaryExportFormat,
} from '@/types/workflow'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  listWorkflowSummaryInstances,
  workflowSummaryExportUrl,
} from '@/api/workflow'
import { useDingtalkAuthStore } from '@/stores'
import { buildWorkflowHistoryTimeQuery } from '../workflow-history-filter'
import { workflowInstanceStatusMeta } from '../workflow-status'
import {
  visibleWorkflowSummaryColumns,
  workflowSummaryBaseColumns,
  workflowSummaryFormFieldGroups,
  workflowSummaryFormValue,
} from '../workflow-summary-columns'
import WorkflowDetailPanel from './WorkflowDetailPanel.vue'
import WorkflowFilterPanel from './WorkflowFilterPanel.vue'
import WorkflowHistoryDatePicker from './WorkflowHistoryDatePicker.vue'

interface SummaryFilters extends WorkflowHistoryDateFilters {
  instanceTitle: string
  definitionName: string
  starterName: string
  status: string
  definitionVersion: string
}

interface PaginationChangePayload {
  current: number
}

interface SummaryColumnSettings {
  baseHiddenKeys: string[]
  activeDefinitionId: string
  formSelectedKeysByDefinition: Record<string, string[]>
}

const props = defineProps<{
  definitions: WorkflowPublishedDefinition[]
  refreshTick?: number
}>()

const auth = useDingtalkAuthStore()
const loading = ref(false)
const loaded = ref(false)
const instances = ref<WorkflowInstanceSummary[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref<20 | 50>(20)
const filters = ref<SummaryFilters>(emptyFilters())
const appliedFilters = ref<SummaryFilters>(emptyFilters())
const selectedIds = ref<string[]>([])
const exportFormat = ref<WorkflowSummaryExportFormat>('xlsx')
const selectedInstanceId = ref('')
const detailVisible = ref(false)
const summaryColumnStorageKey = 'workflow_summary_column_settings_v2'
const legacySummaryColumnStorageKey = 'workflow_summary_column_settings_v1'
const columnSettingsOpen = ref(false)
const columnSettings = ref<SummaryColumnSettings>(loadColumnSettings())
const formColumnKeyword = ref('')
const formFieldDropdownOpen = ref(false)

const canExport = computed(() => auth.hasApiPermission('dingtalk_h5:api:workflow:export'))
const allCurrentPageSelected = computed(() => {
  return instances.value.length > 0 && instances.value.every(item => selectedIds.value.includes(item.id))
})
const filterCount = computed(() => [
  filters.value.instanceTitle.trim(),
  filters.value.definitionName.trim(),
  filters.value.starterName.trim(),
  filters.value.status,
  filters.value.definitionVersion,
  filters.value.startDateFrom,
  filters.value.startDateTo,
  filters.value.endDateFrom,
  filters.value.endDateTo,
].filter(Boolean).length)

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '审批中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '已驳回', value: 'rejected' },
  { label: '已撤回', value: 'withdrawn' },
  { label: '已取消', value: 'cancelled' },
]

const configurableBaseColumns = workflowSummaryBaseColumns.filter(column => column.configurable)
const formFieldGroups = computed(() => workflowSummaryFormFieldGroups(props.definitions))
const selectedColumnDefinitionId = computed<string>({
  get: () => columnSettings.value.activeDefinitionId,
  set: (value) => {
    columnSettings.value.activeDefinitionId = String(value || '')
    formColumnKeyword.value = ''
    formFieldDropdownOpen.value = false
    saveColumnSettings()
  },
})
const activeFormFieldGroup = computed(() => formFieldGroups.value.find(
  group => String(group.definitionId) === selectedColumnDefinitionId.value,
) || null)
const filteredActiveFormColumns = computed(() => {
  const columns = activeFormFieldGroup.value?.columns || []
  const keyword = formColumnKeyword.value.trim().toLowerCase()
  if (!keyword)
    return columns
  return columns.filter(column => (
    column.label.toLowerCase().includes(keyword)
    || String(column.field?.key || '').toLowerCase().includes(keyword)
  ))
})
const visibleBaseColumnKeys = computed<string[]>({
  get: () => configurableBaseColumns
    .map(column => column.key)
    .filter(key => !columnSettings.value.baseHiddenKeys.includes(key)),
  set: (values) => {
    const selected = new Set(values)
    columnSettings.value.baseHiddenKeys = configurableBaseColumns
      .map(column => column.key)
      .filter(key => !selected.has(key))
    saveColumnSettings()
  },
})
const selectedFormColumnKeys = computed<string[]>({
  get: () => {
    const definitionId = selectedColumnDefinitionId.value
    return definitionId ? columnSettings.value.formSelectedKeysByDefinition[definitionId] || [] : []
  },
  set: (values) => {
    const definitionId = selectedColumnDefinitionId.value
    if (!definitionId)
      return
    columnSettings.value.formSelectedKeysByDefinition = {
      ...columnSettings.value.formSelectedKeysByDefinition,
      [definitionId]: normalizeSelectedFormKeys(definitionId, values),
    }
    saveColumnSettings()
  },
})
const selectedActiveFormColumnCount = computed(() => selectedFormColumnKeys.value.length)
const formFieldSelectionText = computed(() => selectedActiveFormColumnCount.value > 0
  ? `已选 ${selectedActiveFormColumnCount.value} 项`
  : '请选择表单字段')
const summaryColumns = computed(() => visibleWorkflowSummaryColumns(
  columnSettings.value.baseHiddenKeys,
  selectedFormColumnKeys.value,
  formFieldGroups.value,
))
const summaryGridStyle = computed(() => ({
  gridTemplateColumns: summaryColumns.value.map(column => column.width).join(' '),
}))
const summaryTableStyle = computed(() => {
  const width = summaryColumns.value.reduce((total, column) => {
    const matched = column.width.match(/(\d+(?:\.\d+)?)px/)
    return total + (matched ? Number(matched[1]) : 120)
  }, 0)
  return { minWidth: `${width}px` }
})

onMounted(() => {
  // #ifdef H5
  document.addEventListener('click', closeColumnSettings)
  // #endif
  void loadSummary()
})

onBeforeUnmount(() => {
  // #ifdef H5
  document.removeEventListener('click', closeColumnSettings)
  // #endif
})

watch(
  () => props.refreshTick,
  () => void loadSummary(),
)

watch(
  () => props.definitions,
  () => {
    const normalized = normalizeColumnSettings(columnSettings.value)
    if (JSON.stringify(normalized) !== JSON.stringify(columnSettings.value)) {
      columnSettings.value = normalized
      saveColumnSettings()
    }
  },
  { deep: true },
)

function emptyFilters(): SummaryFilters {
  return {
    instanceTitle: '',
    definitionName: '',
    starterName: '',
    status: '',
    definitionVersion: '',
    startDateFrom: '',
    startDateTo: '',
    endDateFrom: '',
    endDateTo: '',
  }
}

function loadColumnSettings(): SummaryColumnSettings {
  const stored = uni.getStorageSync(summaryColumnStorageKey)
  if (stored && typeof stored === 'object' && !Array.isArray(stored))
    return normalizeStoredColumnSettings(stored)

  const legacy = uni.getStorageSync(legacySummaryColumnStorageKey)
  if (!legacy || typeof legacy !== 'object' || Array.isArray(legacy))
    return emptyColumnSettings()
  const value = legacy as { baseHiddenKeys?: unknown, selectedFormKeys?: unknown }
  const selectedFormKeys = Array.isArray(value.selectedFormKeys) ? value.selectedFormKeys.map(String) : []
  const formSelectedKeysByDefinition = selectedFormKeys.reduce<Record<string, string[]>>((result, key) => {
    const definitionId = formColumnDefinitionId(key)
    if (!definitionId)
      return result
    result[definitionId] = [...(result[definitionId] || []), key]
    return result
  }, {})
  return {
    baseHiddenKeys: Array.isArray(value.baseHiddenKeys) ? value.baseHiddenKeys.map(String) : [],
    activeDefinitionId: Object.keys(formSelectedKeysByDefinition)[0] || '',
    formSelectedKeysByDefinition,
  }
}

function saveColumnSettings() {
  uni.setStorageSync(summaryColumnStorageKey, {
    baseHiddenKeys: [...columnSettings.value.baseHiddenKeys],
    activeDefinitionId: columnSettings.value.activeDefinitionId,
    formSelectedKeysByDefinition: Object.fromEntries(
      Object.entries(columnSettings.value.formSelectedKeysByDefinition)
        .map(([definitionId, keys]) => [definitionId, [...keys]]),
    ),
  })
}

function emptyColumnSettings(): SummaryColumnSettings {
  return { baseHiddenKeys: [], activeDefinitionId: '', formSelectedKeysByDefinition: {} }
}

function normalizeStoredColumnSettings(stored: object): SummaryColumnSettings {
  const value = stored as Partial<SummaryColumnSettings>
  const selections = value.formSelectedKeysByDefinition && typeof value.formSelectedKeysByDefinition === 'object'
    && !Array.isArray(value.formSelectedKeysByDefinition)
    ? Object.fromEntries(Object.entries(value.formSelectedKeysByDefinition).map(([definitionId, keys]) => [
        String(definitionId),
        Array.isArray(keys) ? keys.map(String) : [],
      ]))
    : {}
  return {
    baseHiddenKeys: Array.isArray(value.baseHiddenKeys) ? value.baseHiddenKeys.map(String) : [],
    activeDefinitionId: String(value.activeDefinitionId || ''),
    formSelectedKeysByDefinition: selections,
  }
}

function normalizeColumnSettings(settings: SummaryColumnSettings): SummaryColumnSettings {
  const availableDefinitionIds = new Set(formFieldGroups.value.map(group => String(group.definitionId)))
  const activeDefinitionId = availableDefinitionIds.has(settings.activeDefinitionId) ? settings.activeDefinitionId : ''
  const formSelectedKeysByDefinition = Object.fromEntries(
    Object.entries(settings.formSelectedKeysByDefinition)
      .filter(([definitionId]) => availableDefinitionIds.has(definitionId))
      .map(([definitionId, keys]) => [definitionId, normalizeSelectedFormKeys(definitionId, keys)]),
  )
  return {
    baseHiddenKeys: [...settings.baseHiddenKeys],
    activeDefinitionId,
    formSelectedKeysByDefinition,
  }
}

function normalizeSelectedFormKeys(definitionId: string, values: string[]) {
  const group = formFieldGroups.value.find(item => String(item.definitionId) === definitionId)
  const available = new Set((group?.columns || []).map(column => column.key))
  return [...new Set(values.filter(key => available.has(key)))]
}

function formColumnDefinitionId(key: string) {
  const match = String(key || '').match(/^form:(\d+):/)
  return match?.[1] || ''
}

function toggleColumnSettings() {
  if (columnSettingsOpen.value) {
    closeColumnSettings()
    return
  }
  columnSettingsOpen.value = true
}

function closeColumnSettings() {
  columnSettingsOpen.value = false
  formFieldDropdownOpen.value = false
  formColumnKeyword.value = ''
}

function toggleFormFieldDropdown() {
  if (!activeFormFieldGroup.value)
    return
  formFieldDropdownOpen.value = !formFieldDropdownOpen.value
  if (!formFieldDropdownOpen.value)
    formColumnKeyword.value = ''
}

function resetColumnSettings() {
  columnSettings.value = emptyColumnSettings()
  formColumnKeyword.value = ''
  formFieldDropdownOpen.value = false
  saveColumnSettings()
}

function summaryColumnClass(column: WorkflowSummaryColumn) {
  const key = column.baseKey || 'form'
  const classKeyMap: Partial<Record<NonNullable<WorkflowSummaryColumn['baseKey']>, string>> = {
    check: 'check',
    title: 'definition',
    workflowName: 'workflow-name',
    businessKey: 'key',
    starter: 'starter',
    operator: 'operator',
    version: 'version',
    status: 'status',
    startTime: 'start-time',
    endTime: 'end-time',
    actions: 'action',
  }
  return [
    `workflow-summary__cell--${classKeyMap[key as NonNullable<WorkflowSummaryColumn['baseKey']>] || 'form'}`,
    { 'workflow-summary__cell--mobile-secondary': column.mobileHidden },
  ]
}

function summaryCellText(instance: WorkflowInstanceSummary, column: WorkflowSummaryColumn) {
  if (column.kind === 'form')
    return workflowSummaryFormValue(instance, column)
  switch (column.baseKey) {
    case 'title': return instance.instanceTitle || instance.definitionName || instance.definitionKey || '-'
    case 'workflowName': return instance.definitionName || instance.definitionKey || '-'
    case 'businessKey': return instance.businessKey || instance.id
    case 'starter': return instance.starterName || instance.starterId || '-'
    case 'operator': return instance.operatorName || instance.operatorId || '-'
    case 'version': return `v${instance.definitionVersion}`
    case 'startTime': return formatTime(instance.startTime)
    case 'endTime': return formatTime(instance.endTime)
    default: return '-'
  }
}

function validateFilters(value: SummaryFilters) {
  if (buildWorkflowHistoryTimeQuery(value) === null) {
    uni.showToast({ title: '请检查时间范围', icon: 'none' })
    return false
  }
  const version = value.definitionVersion.trim()
  if (version && (!/^\d+$/.test(version) || Number(version) <= 0)) {
    uni.showToast({ title: '流程版本必须是正整数', icon: 'none' })
    return false
  }
  return true
}

async function loadSummary() {
  if (loading.value)
    return
  const timeQuery = buildWorkflowHistoryTimeQuery(appliedFilters.value)
  if (timeQuery === null)
    return
  loading.value = true
  try {
    const response = await listWorkflowSummaryInstances({
      instanceTitle: appliedFilters.value.instanceTitle.trim() || undefined,
      definitionName: appliedFilters.value.definitionName.trim() || undefined,
      definitionVersion: Number(appliedFilters.value.definitionVersion) || undefined,
      starterName: appliedFilters.value.starterName.trim() || undefined,
      status: appliedFilters.value.status || undefined,
      ...timeQuery,
      page: page.value,
      pageSize: pageSize.value,
    })
    instances.value = Array.isArray(response?.data?.list) ? response.data.list : []
    total.value = Number(response?.data?.total || 0)
    selectedIds.value = []
    loaded.value = true
  }
  catch {
    uni.showToast({ title: '流程汇总加载失败', icon: 'none' })
  }
  finally {
    loading.value = false
  }
}

function querySummary() {
  if (!validateFilters(filters.value) || loading.value)
    return
  appliedFilters.value = { ...filters.value }
  page.value = 1
  void loadSummary()
}

function resetFilters() {
  if (loading.value)
    return
  filters.value = emptyFilters()
  appliedFilters.value = emptyFilters()
  page.value = 1
  void loadSummary()
}

function handlePageChange(payload: PaginationChangePayload) {
  page.value = Number(payload.current || 1)
  void loadSummary()
}

function handlePageSizeChange(event: Event) {
  const value = Number((event.target as HTMLSelectElement).value)
  pageSize.value = value === 50 ? 50 : 20
  page.value = 1
  void loadSummary()
}

function toggleCurrentPage() {
  selectedIds.value = allCurrentPageSelected.value ? [] : instances.value.map(item => item.id)
}

function openDetail(instanceId: string) {
  selectedInstanceId.value = instanceId
  detailVisible.value = true
}

function exportInstances(instanceIds: string[]) {
  if (!canExport.value) {
    uni.showToast({ title: '暂无流程导出权限', icon: 'none' })
    return
  }
  if (instanceIds.length === 0) {
    uni.showToast({ title: '请选择当前页需要导出的记录', icon: 'none' })
    return
  }
  const selectedDefinitionIds = [...new Set(
    instances.value
      .filter(instance => instanceIds.includes(instance.id))
      .map(instance => instance.definitionId),
  )]
  if (selectedDefinitionIds.length !== 1) {
    uni.showToast({ title: '请选择同一流程的记录批量导出', icon: 'none' })
    return
  }
  const url = workflowSummaryExportUrl(selectedDefinitionIds[0], instanceIds, exportFormat.value)
  // #ifdef H5
  if (typeof window !== 'undefined') {
    window.open(url, '_blank')
    return
  }
  // #endif
  uni.showToast({ title: '请在 H5 端下载导出文件', icon: 'none' })
}

function formatTime(timestamp?: number) {
  if (!timestamp)
    return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}
</script>

<template>
  <view class="workflow-summary">
    <WorkflowFilterPanel :active-count="filterCount">
      <view class="workflow-summary__filters">
        <view class="workflow-summary__filter workflow-summary__filter--instance-title">
          <text class="workflow-summary__filter-label">
            单据标题
          </text>
          <u-input
            v-model="filters.instanceTitle"
            :border="true"
            clearable
            :maxlength="100"
            placeholder="输入单据标题"
            @confirm="querySummary"
          />
        </view>
        <view class="workflow-summary__filter workflow-summary__filter--definition-name">
          <text class="workflow-summary__filter-label">
            流程名称
          </text>
          <u-input
            v-model="filters.definitionName"
            :border="true"
            clearable
            :maxlength="50"
            placeholder="输入流程名称"
            @confirm="querySummary"
          />
        </view>
        <view class="workflow-summary__filter workflow-summary__filter--starter">
          <text class="workflow-summary__filter-label">
            发起人
          </text>
          <u-input
            v-model="filters.starterName"
            :border="true"
            clearable
            placeholder="输入用户名模糊搜索"
            @confirm="querySummary"
          />
        </view>
        <view class="workflow-summary__filter workflow-summary__filter--status">
          <text class="workflow-summary__filter-label">
            状态
          </text>
          <select v-model="filters.status" class="workflow-summary__select" :disabled="loading" aria-label="流程状态">
            <option v-for="option in statusOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </view>
        <view class="workflow-summary__filter workflow-summary__filter--version">
          <text class="workflow-summary__filter-label">
            流程版本
          </text>
          <u-input v-model="filters.definitionVersion" :border="true" type="number" clearable placeholder="全部版本" />
        </view>
        <view class="workflow-summary__filter workflow-summary__filter--range workflow-summary__filter--start">
          <text class="workflow-summary__filter-label">
            发起时间
          </text>
          <view class="workflow-summary__date-range">
            <WorkflowHistoryDatePicker v-model="filters.startDateFrom" :disabled="loading" placeholder="开始日期" />
            <text>至</text>
            <WorkflowHistoryDatePicker v-model="filters.startDateTo" :disabled="loading" placeholder="结束日期" />
          </view>
        </view>
        <view class="workflow-summary__filter workflow-summary__filter--range workflow-summary__filter--end">
          <text class="workflow-summary__filter-label">
            完成时间
          </text>
          <view class="workflow-summary__date-range">
            <WorkflowHistoryDatePicker v-model="filters.endDateFrom" :disabled="loading" placeholder="开始日期" />
            <text>至</text>
            <WorkflowHistoryDatePicker v-model="filters.endDateTo" :disabled="loading" placeholder="结束日期" />
          </view>
        </view>
        <view class="workflow-summary__filter-actions">
          <u-button custom-class="workflow-summary__filter-action" size="small" plain :disabled="loading" @click="resetFilters">
            重置
          </u-button>
          <u-button custom-class="workflow-summary__filter-action" size="small" type="primary" :loading="loading" @click="querySummary">
            查询
          </u-button>
        </view>
      </view>
    </WorkflowFilterPanel>

    <view class="workflow-summary__toolbar">
      <view class="workflow-summary__toolbar-main">
        <u-button custom-class="workflow-summary__toolbar-button" size="small" plain :disabled="loading || instances.length === 0" @click="toggleCurrentPage">
          <u-icon :name="allCurrentPageSelected ? 'checkbox-mark' : 'grid'" size="18" color="#4e5969" />
          <text>{{ allCurrentPageSelected ? '取消本页全选' : '本页全选' }}</text>
        </u-button>
        <text class="workflow-summary__selection">
          已选 {{ selectedIds.length }} 条
        </text>
      </view>
      <view class="workflow-summary__toolbar-actions">
        <!-- #ifdef H5 -->
        <view class="workflow-summary__column-settings" @click.stop>
          <view
            class="workflow-summary__column-settings-trigger"
            role="button"
            tabindex="0"
            title="列设置"
            aria-label="列设置"
            @click="toggleColumnSettings"
            @keydown.enter.prevent="toggleColumnSettings"
            @keydown.space.prevent="toggleColumnSettings"
          >
            <u-icon name="setting" size="15px" color="#4e5969" />
            <text>列设置</text>
          </view>
          <view v-if="columnSettingsOpen" class="workflow-summary__column-settings-panel">
            <view class="workflow-summary__column-settings-head">
              <text>展示列</text>
              <text class="workflow-summary__column-settings-reset" @click="resetColumnSettings">
                恢复默认
              </text>
            </view>
            <scroll-view scroll-y class="workflow-summary__column-settings-scroll">
              <view class="workflow-summary__column-settings-group">
                <text class="workflow-summary__column-settings-group-title">
                  基础列
                </text>
                <u-checkbox-group v-model="visibleBaseColumnKeys" class="workflow-summary__column-settings-list">
                  <u-checkbox
                    v-for="column in configurableBaseColumns"
                    :key="column.key"
                    :value="column.key"
                    :label="column.label"
                  />
                </u-checkbox-group>
              </view>
              <view class="workflow-summary__column-settings-group">
                <text class="workflow-summary__column-settings-group-title">
                  流程字段
                </text>
                <select
                  v-model="selectedColumnDefinitionId"
                  class="workflow-summary__column-settings-select"
                  aria-label="选择流程字段所属流程"
                >
                  <option value="">
                    请选择流程
                  </option>
                  <option v-for="group in formFieldGroups" :key="group.definitionId" :value="String(group.definitionId)">
                    {{ group.label }}
                  </option>
                </select>

                <template v-if="activeFormFieldGroup">
                  <view class="workflow-summary__column-settings-form-select">
                    <view
                      class="workflow-summary__column-settings-form-trigger"
                      :class="{ 'workflow-summary__column-settings-form-trigger--open': formFieldDropdownOpen }"
                      role="button"
                      tabindex="0"
                      aria-label="选择表单字段"
                      :aria-expanded="formFieldDropdownOpen"
                      @click="toggleFormFieldDropdown"
                      @keydown.enter.prevent="toggleFormFieldDropdown"
                      @keydown.space.prevent="toggleFormFieldDropdown"
                    >
                      <text :class="{ 'workflow-summary__column-settings-form-placeholder': selectedActiveFormColumnCount === 0 }">
                        {{ formFieldSelectionText }}
                      </text>
                      <u-icon :name="formFieldDropdownOpen ? 'arrow-up' : 'arrow-down'" size="13px" color="#86909c" />
                    </view>
                    <view
                      v-if="formFieldDropdownOpen"
                      class="workflow-summary__column-settings-form-dropdown"
                    >
                      <view class="workflow-summary__column-settings-field-head">
                        <text>{{ activeFormFieldGroup.label }} · 表单字段</text>
                        <text>可多选</text>
                      </view>
                      <input
                        v-model="formColumnKeyword"
                        class="workflow-summary__column-settings-search"
                        type="text"
                        :maxlength="60"
                        placeholder="搜索字段名称"
                      >
                      <u-checkbox-group
                        v-if="filteredActiveFormColumns.length > 0"
                        v-model="selectedFormColumnKeys"
                        class="workflow-summary__column-settings-list workflow-summary__column-settings-list--form"
                      >
                        <u-checkbox
                          v-for="column in filteredActiveFormColumns"
                          :key="column.key"
                          :value="column.key"
                          :label="column.label"
                        />
                      </u-checkbox-group>
                      <text v-else class="workflow-summary__column-settings-empty">
                        未找到匹配字段
                      </text>
                    </view>
                  </view>
                </template>
                <text v-else class="workflow-summary__column-settings-empty">
                  请选择流程后配置表单字段
                </text>
              </view>
            </scroll-view>
          </view>
        </view>
        <!-- #endif -->
        <view v-if="canExport" class="workflow-summary__export-actions">
          <select v-model="exportFormat" class="workflow-summary__select workflow-summary__select--format" aria-label="导出格式">
            <option value="xlsx">
              Excel
            </option>
            <option value="pdf">
              PDF
            </option>
            <option value="docx">
              Word
            </option>
          </select>
          <u-button custom-class="workflow-summary__export-button" size="small" type="primary" :disabled="selectedIds.length === 0" @click="exportInstances(selectedIds)">
            <u-icon name="download" size="18" color="#ffffff" />
            <text>批量导出</text>
          </u-button>
        </view>
      </view>
    </view>

    <view v-if="loading && !loaded" class="workflow-summary__empty">
      <u-loading mode="circle" size="38" color="#0f766e" />
      <text>正在加载汇总...</text>
    </view>
    <view v-else-if="instances.length === 0" class="workflow-summary__empty">
      <u-icon name="order" size="58" color="#c8c9cc" />
      <text class="workflow-summary__empty-title">
        暂无符合条件的流程记录
      </text>
    </view>
    <scroll-view v-else scroll-x class="workflow-summary__table-scroll">
      <u-checkbox-group v-model="selectedIds">
        <view class="workflow-summary__table" :style="summaryTableStyle">
          <view class="workflow-summary__row workflow-summary__row--header" :style="summaryGridStyle">
            <text
              v-for="column in summaryColumns"
              :key="column.key"
              class="workflow-summary__cell"
              :class="summaryColumnClass(column)"
            >
              {{ column.label }}
            </text>
          </view>
          <view v-for="instance in instances" :key="instance.id" class="workflow-summary__row" :style="summaryGridStyle">
            <view
              v-for="column in summaryColumns"
              :key="column.key"
              class="workflow-summary__cell"
              :class="summaryColumnClass(column)"
            >
              <template v-if="column.baseKey === 'check'">
                <u-checkbox custom-class="workflow-summary__checkbox" :value="instance.id" label="" />
                <text class="workflow-summary__mobile-label">
                  选择
                </text>
              </template>
              <template v-else-if="column.baseKey === 'status'">
                <text class="workflow-summary__mobile-label">
                  状态
                </text>
                <u-tag :text="workflowInstanceStatusMeta(instance.status).label" :type="workflowInstanceStatusMeta(instance.status).type" size="mini" />
              </template>
              <template v-else-if="column.baseKey === 'actions'">
                <text class="workflow-summary__link" @click="openDetail(instance.id)">
                  查看
                </text>
                <text v-if="canExport" class="workflow-summary__link" @click="exportInstances([instance.id])">
                  导出
                </text>
              </template>
              <template v-else>
                <text class="workflow-summary__mobile-label">
                  {{ column.label }}
                </text>
                <text class="workflow-summary__cell-value">
                  {{ summaryCellText(instance, column) }}
                </text>
              </template>
            </view>
          </view>
        </view>
      </u-checkbox-group>
    </scroll-view>

    <view v-if="total > 0" class="workflow-summary__pagination">
      <view class="workflow-summary__page-size">
        <text>每页</text>
        <select :value="pageSize" class="workflow-summary__select workflow-summary__select--size" aria-label="每页条数" @change="handlePageSizeChange">
          <option :value="20">
            20 条
          </option>
          <option :value="50">
            50 条
          </option>
        </select>
      </view>
      <u-pagination
        v-model="page"
        custom-class="workflow-summary__pagination-control"
        :total="total"
        :page-size="pageSize"
        prev-text="上一页"
        next-text="下一页"
        @change="handlePageChange"
      />
    </view>

    <WorkflowDetailPanel
      v-model="detailVisible"
      :instance-id="selectedInstanceId"
      :definitions="definitions"
      presentation="history-drawer"
      summary-mode
    />
  </view>
</template>

<style lang="scss" scoped>
.workflow-summary {
  min-height: 0;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 12px;
  padding: 18px 22px 22px;
  overflow: auto;
  box-sizing: border-box;
  background: #f5f7fa;
  color: #1f2329;
}

:deep(.workflow-filter-panel) {
  margin-bottom: 0;
}

.workflow-summary__filters {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 12px 14px;
}

.workflow-summary__filter {
  min-width: 160px;
  max-width: 220px;
  flex: 1 1 180px;
  display: flex;
  flex-direction: column;
  gap: 7px;
}

.workflow-summary__filter--instance-title {
  max-width: 240px;
}

.workflow-summary__filter--status,
.workflow-summary__filter--version {
  min-width: 120px;
  max-width: 140px;
  flex: 0 1 140px;
}

.workflow-summary__filter--range {
  min-width: 280px;
  max-width: 340px;
  flex: 1 1 280px;
}

.workflow-summary__filter-label {
  color: #4e5969;
  font-size: 13px;
}

.workflow-summary__select {
  width: 100%;
  height: 38px;
  padding: 0 10px;
  border: 1px solid #d9e1e8;
  border-radius: 4px;
  outline: none;
  background: #ffffff;
  color: #1f2329;
  font-size: 14px;
  box-sizing: border-box;
}

.workflow-summary__select:focus {
  border-color: #0f766e;
}

.workflow-summary__date-range,
.workflow-summary__filter-actions,
.workflow-summary__toolbar,
.workflow-summary__toolbar-main,
.workflow-summary__export-actions,
.workflow-summary__pagination,
.workflow-summary__page-size,
.workflow-summary__cell--action {
  display: flex;
  align-items: center;
}

.workflow-summary__date-range {
  min-width: 0;
  gap: 8px;
}

.workflow-summary__date-range > :first-child,
.workflow-summary__date-range > :last-child {
  min-width: 0;
  flex: 1 1 0;
}

.workflow-summary__filter-actions {
  flex: 0 0 auto;
  width: fit-content;
  gap: 8px;
}

.workflow-summary__filter-action,
:deep(.workflow-summary__filter-action) {
  width: auto;
  min-width: 64px;
  height: 36px;
  min-height: 36px;
  margin: 0;
  padding: 0 14px;
  border-radius: 4px;
  font-size: 13px;
  line-height: 34px;
}

.workflow-summary__toolbar {
  min-height: 42px;
  justify-content: space-between;
  gap: 12px;
}

.workflow-summary__toolbar-main,
.workflow-summary__toolbar-actions,
.workflow-summary__export-actions {
  gap: 10px;
}

.workflow-summary__toolbar-actions {
  display: flex;
  align-items: center;
}

.workflow-summary__column-settings {
  position: relative;
  flex: 0 0 auto;
}

.workflow-summary__column-settings-trigger {
  width: 86px;
  height: 34px;
  padding: 0 10px;
  border: 1px solid #d8e0e8;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: #fff;
  color: #4e5969;
  font-size: 12px;
  cursor: pointer;
  box-sizing: border-box;
}

.workflow-summary__column-settings-trigger:hover,
.workflow-summary__column-settings-trigger:focus-visible {
  border-color: #0f766e;
  color: #0f766e;
  outline: none;
}

.workflow-summary__column-settings-panel {
  position: absolute;
  top: 40px;
  right: 0;
  z-index: 30;
  width: 320px;
  padding: 12px;
  border: 1px solid #e5e9ef;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 12px 32px rgba(31, 35, 41, 0.16);
  box-sizing: border-box;
}

.workflow-summary__column-settings-head {
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
}

.workflow-summary__column-settings-reset {
  color: #0f766e;
  font-size: 12px;
  font-weight: 400;
  cursor: pointer;
}

.workflow-summary__column-settings-scroll {
  max-height: 360px;
}

.workflow-summary__column-settings-group + .workflow-summary__column-settings-group {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid #eef1f4;
}

.workflow-summary__column-settings-group-title {
  display: block;
  margin-bottom: 8px;
  color: #86909c;
  font-size: 11px;
  font-weight: 600;
}

.workflow-summary__column-settings-list {
  display: grid;
  gap: 8px;
}

.workflow-summary__column-settings-select,
.workflow-summary__column-settings-search {
  width: 100%;
  height: 34px;
  min-width: 0;
  padding: 0 10px;
  border: 1px solid #d8e0e8;
  border-radius: 4px;
  background: #fff;
  color: #1f2329;
  font-family: inherit;
  font-size: 12px;
  box-sizing: border-box;
}

.workflow-summary__column-settings-select {
  padding-right: 28px;
  cursor: pointer;
}

.workflow-summary__column-settings-form-select {
  margin-top: 8px;
}

.workflow-summary__column-settings-form-trigger {
  width: 100%;
  height: 34px;
  min-width: 0;
  padding: 0 10px;
  border: 1px solid #d8e0e8;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  background: #fff;
  color: #1f2329;
  font-size: 12px;
  cursor: pointer;
  box-sizing: border-box;
}

.workflow-summary__column-settings-form-trigger:hover,
.workflow-summary__column-settings-form-trigger:focus-visible,
.workflow-summary__column-settings-form-trigger--open {
  border-color: #0f766e;
  outline: none;
  box-shadow: 0 0 0 2px rgba(15, 118, 110, 0.12);
}

.workflow-summary__column-settings-form-trigger > text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-summary__column-settings-form-placeholder {
  color: #a9b0bb;
}

.workflow-summary__column-settings-form-dropdown {
  margin-top: 6px;
  padding: 10px;
  border: 1px solid #e5e9ef;
  border-radius: 4px;
  background: #fff;
  box-shadow: 0 8px 20px rgba(31, 35, 41, 0.1);
}

.workflow-summary__column-settings-search {
  margin: 10px 0;
}

.workflow-summary__column-settings-select:focus,
.workflow-summary__column-settings-search:focus {
  border-color: #0f766e;
  outline: none;
  box-shadow: 0 0 0 2px rgba(15, 118, 110, 0.12);
}

.workflow-summary__column-settings-search::placeholder {
  color: #a9b0bb;
}

.workflow-summary__column-settings-field-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  color: #667085;
  font-size: 11px;
}

.workflow-summary__column-settings-field-head > text:first-child {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-summary__column-settings-field-head > text:last-child {
  flex: 0 0 auto;
  color: #86909c;
}

.workflow-summary__column-settings-list--form {
  width: 100%;
  max-height: 180px;
  padding-right: 4px;
  display: grid !important;
  grid-template-columns: minmax(0, 1fr);
  align-content: start;
  overflow-y: auto;
  box-sizing: border-box;
}

:deep(.workflow-summary__column-settings-list--form .u-checkbox) {
  width: 100%;
  min-height: 28px;
  align-items: flex-start;
  overflow: visible;
  line-height: 20px;
  box-sizing: border-box;
}

:deep(.workflow-summary__column-settings-list--form .u-checkbox__icon-wrap) {
  margin-top: 1px;
}

:deep(.workflow-summary__column-settings-list--form .u-checkbox__label) {
  min-width: 0;
  margin-right: 0 !important;
  line-height: 20px !important;
  white-space: normal;
  overflow-wrap: anywhere;
}

.workflow-summary__column-settings-empty {
  padding: 18px 0 8px;
  display: block;
  color: #98a2b3;
  font-size: 12px;
  text-align: center;
}

:deep(.workflow-summary__column-settings-list .u-checkbox) {
  margin-right: 0;
}

.workflow-summary__selection {
  color: #86909c;
  font-size: 13px;
}

.workflow-summary__select--format {
  width: 104px;
  height: 34px;
}

.workflow-summary__table-scroll {
  width: 100%;
  min-height: 0;
  border: 1px solid #e5e9ef;
  border-radius: 6px;
  background: #ffffff;
  box-sizing: border-box;
}

:deep(.workflow-summary__table-scroll .u-checkbox-group) {
  display: block;
  width: 100%;
}

.workflow-summary__table {
  width: 100%;
  min-width: 1420px;
}

.workflow-summary__row {
  display: grid;
  grid-template-columns: 66px minmax(180px, 1.3fr) minmax(140px, 1fr) minmax(190px, 1.45fr) minmax(100px, 0.8fr) minmax(100px, 0.8fr) 76px 94px 168px 168px 112px;
  min-height: 58px;
  border-bottom: 1px solid #eef1f4;
  align-items: center;
}

.workflow-summary__row:last-child {
  border-bottom: 0;
}

.workflow-summary__row--header {
  min-height: 44px;
  background: #f7f8fa;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
}

.workflow-summary__cell {
  min-width: 0;
  padding: 10px 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  box-sizing: border-box;
}

.workflow-summary__cell-value {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-summary__mobile-label {
  display: none;
}

.workflow-summary__cell--check,
.workflow-summary__cell--version,
.workflow-summary__cell--status {
  justify-content: center;
  text-align: center;
}

.workflow-summary__cell--action {
  justify-content: flex-start;
  gap: 12px;
}

.workflow-summary__link {
  color: #0f766e;
  cursor: pointer;
}

.workflow-summary__pagination {
  justify-content: space-between;
  gap: 18px;
}

:deep(.workflow-summary__pagination-control) {
  flex: 0 0 auto;
  gap: 10px;
}

.workflow-summary__page-size {
  gap: 8px;
  color: #86909c;
  font-size: 13px;
}

.workflow-summary__select--size {
  width: 88px;
  height: 34px;
}

.workflow-summary__empty {
  min-height: 260px;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #86909c;
}

.workflow-summary__empty-title {
  color: #4e5969;
  font-size: 15px;
}

@media (max-width: 768px) {
  .workflow-summary {
    width: 100%;
    gap: 10px;
    padding: 12px 10px 16px;
    overflow-x: hidden;
  }

  .workflow-summary__filters {
    display: grid;
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .workflow-summary__filter {
    width: 100%;
    min-width: 0;
    max-width: none;
  }

  .workflow-summary__filter--instance-title,
  .workflow-summary__filter--definition-name,
  .workflow-summary__filter--starter,
  .workflow-summary__filter--status,
  .workflow-summary__filter--version,
  .workflow-summary__filter--start,
  .workflow-summary__filter--end,
  .workflow-summary__filter-actions {
    grid-column: auto;
    grid-row: auto;
  }

  .workflow-summary__date-range {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
    align-items: center;
    gap: 6px;
  }

  .workflow-summary__date-range > text {
    color: #86909c;
    font-size: 12px;
  }

  .workflow-summary__filter-actions {
    width: 100%;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    justify-self: stretch;
  }

  .workflow-summary__filter-action,
  :deep(.workflow-summary__filter-action),
  .workflow-summary__toolbar-button,
  :deep(.workflow-summary__toolbar-button),
  .workflow-summary__export-button,
  :deep(.workflow-summary__export-button) {
    width: 100%;
    min-width: 0;
  }

  .workflow-summary__toolbar,
  .workflow-summary__pagination {
    width: 100%;
    padding: 12px;
    border: 1px solid #e5e9ef;
    border-radius: 6px;
    align-items: stretch;
    flex-direction: column;
    background: #ffffff;
    box-sizing: border-box;
  }

  .workflow-summary__column-settings {
    display: none;
  }

  .workflow-summary__toolbar-main,
  .workflow-summary__export-actions {
    width: 100%;
    display: grid;
    align-items: center;
    gap: 8px;
  }

  .workflow-summary__toolbar-main {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .workflow-summary__export-actions {
    grid-template-columns: 96px minmax(0, 1fr);
  }

  .workflow-summary__select--format {
    width: 96px;
  }

  .workflow-summary__table-scroll {
    border: 0;
    overflow: visible;
    background: transparent;
  }

  :deep(.workflow-summary__table-scroll .u-checkbox-group) {
    width: 100%;
  }

  .workflow-summary__table {
    min-width: 0;
    display: grid;
    gap: 10px;
  }

  .workflow-summary__row--header {
    display: none;
  }

  .workflow-summary__row:not(.workflow-summary__row--header) {
    min-height: 0;
    padding: 12px;
    border: 1px solid #e5e9ef;
    border-radius: 6px;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    align-items: start;
    gap: 12px 14px;
    background: #ffffff;
  }

  .workflow-summary__row:last-child {
    border-bottom: 1px solid #e5e9ef;
  }

  .workflow-summary__cell {
    padding: 0;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    overflow: visible;
    text-align: left;
    white-space: normal;
  }

  .workflow-summary__mobile-label {
    display: block;
    color: #86909c;
    font-size: 11px;
    line-height: 16px;
  }

  .workflow-summary__cell-value {
    width: 100%;
    color: #4e5969;
    font-size: 13px;
    line-height: 20px;
    white-space: normal;
    overflow-wrap: anywhere;
  }

  .workflow-summary__cell--check {
    grid-column: 1;
    grid-row: 1;
    flex-direction: row;
    align-items: center;
    justify-content: flex-start;
    gap: 6px;
    text-align: left;
  }

  .workflow-summary__checkbox,
  :deep(.workflow-summary__checkbox) {
    flex: 0 0 auto;
    margin: 0;
    line-height: 1;
  }

  :deep(.workflow-summary__checkbox .u-checkbox__label) {
    display: none;
  }

  .workflow-summary__cell--status {
    grid-column: 2;
    grid-row: 1;
    align-items: flex-end;
  }

  .workflow-summary__cell--definition {
    grid-column: 1 / -1;
    grid-row: 2;
  }

  .workflow-summary__cell--definition .workflow-summary__cell-value {
    color: #1f2329;
    font-size: 14px;
    font-weight: 700;
  }

  .workflow-summary__cell--key {
    grid-column: 1 / -1;
    grid-row: 3;
  }

  .workflow-summary__cell--starter {
    grid-column: 1;
    grid-row: 4;
  }

  .workflow-summary__cell--start-time {
    grid-column: 2;
    grid-row: 4;
  }

  .workflow-summary__cell--mobile-secondary {
    display: none;
  }

  .workflow-summary__cell--action {
    grid-column: 1 / -1;
    grid-row: 5;
    min-height: 32px;
    flex-direction: row;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
  }

  .workflow-summary__link {
    min-width: 44px;
    min-height: 32px;
    padding: 0 8px;
    border: 1px solid #b7d8d4;
    border-radius: 4px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
  }

  .workflow-summary__pagination {
    gap: 10px;
  }

  .workflow-summary__page-size {
    justify-content: space-between;
  }

  :deep(.workflow-summary__pagination-control) {
    width: 100%;
    max-width: 100%;
    justify-content: center;
    gap: 4px;
  }
}
</style>
