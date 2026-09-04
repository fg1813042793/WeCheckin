<script setup lang="ts">
import type { WorkflowHistoryDateFilters } from '../workflow-history-filter'
import type {
  WorkflowInstanceSummary,
  WorkflowPublishedDefinition,
  WorkflowSummaryExportFormat,
} from '@/types/workflow'
import { computed, onMounted, ref } from 'vue'
import {
  listWorkflowSummaryInstances,
  workflowSummaryExportUrl,
} from '@/api/workflow'
import { useDingtalkAuthStore } from '@/stores'
import { buildWorkflowHistoryTimeQuery } from '../workflow-history-filter'
import { workflowInstanceStatusMeta } from '../workflow-status'
import WorkflowDetailPanel from './WorkflowDetailPanel.vue'
import WorkflowFilterPanel from './WorkflowFilterPanel.vue'
import WorkflowHistoryDatePicker from './WorkflowHistoryDatePicker.vue'

interface SummaryFilters extends WorkflowHistoryDateFilters {
  definitionName: string
  starterName: string
  status: string
  definitionVersion: string
}

interface PaginationChangePayload {
  current: number
}

defineProps<{
  definitions: WorkflowPublishedDefinition[]
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

const canExport = computed(() => auth.hasApiPermission('dingtalk_h5:api:workflow:export'))
const allCurrentPageSelected = computed(() => {
  return instances.value.length > 0 && instances.value.every(item => selectedIds.value.includes(item.id))
})
const filterCount = computed(() => [
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

onMounted(() => void loadSummary())

function emptyFilters(): SummaryFilters {
  return {
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
          <u-button size="small" plain :disabled="loading" @click="resetFilters">
            重置
          </u-button>
          <u-button size="small" type="primary" :loading="loading" @click="querySummary">
            查询
          </u-button>
        </view>
      </view>
    </WorkflowFilterPanel>

    <view class="workflow-summary__toolbar">
      <view class="workflow-summary__toolbar-main">
        <u-button size="small" plain :disabled="loading || instances.length === 0" @click="toggleCurrentPage">
          <u-icon :name="allCurrentPageSelected ? 'checkbox-mark' : 'grid'" size="18" color="#4e5969" />
          <text>{{ allCurrentPageSelected ? '取消本页全选' : '本页全选' }}</text>
        </u-button>
        <text class="workflow-summary__selection">
          已选 {{ selectedIds.length }} 条
        </text>
      </view>
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
        <u-button size="small" type="primary" :disabled="selectedIds.length === 0" @click="exportInstances(selectedIds)">
          <u-icon name="download" size="18" color="#ffffff" />
          <text>批量导出</text>
        </u-button>
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
        <view class="workflow-summary__table">
          <view class="workflow-summary__row workflow-summary__row--header">
            <text class="workflow-summary__cell workflow-summary__cell--check">
              选择
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--definition">
              流程名称
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--key">
              申请编号
            </text>
            <text class="workflow-summary__cell">
              发起人
            </text>
            <text class="workflow-summary__cell">
              操作人
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--version">
              版本
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--status">
              状态
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--time">
              发起时间
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--time">
              完成时间
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--action">
              操作
            </text>
          </view>
          <view v-for="instance in instances" :key="instance.id" class="workflow-summary__row">
            <view class="workflow-summary__cell workflow-summary__cell--check">
              <u-checkbox :value="instance.id" label="" />
            </view>
            <text class="workflow-summary__cell workflow-summary__cell--definition">
              {{ instance.definitionName || instance.definitionKey || '-' }}
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--key">
              {{ instance.businessKey || instance.id }}
            </text>
            <text class="workflow-summary__cell">
              {{ instance.starterName || instance.starterId || '-' }}
            </text>
            <text class="workflow-summary__cell">
              {{ instance.operatorName || instance.operatorId || '-' }}
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--version">
              v{{ instance.definitionVersion }}
            </text>
            <view class="workflow-summary__cell workflow-summary__cell--status">
              <u-tag :text="workflowInstanceStatusMeta(instance.status).label" :type="workflowInstanceStatusMeta(instance.status).type" size="mini" />
            </view>
            <text class="workflow-summary__cell workflow-summary__cell--time">
              {{ formatTime(instance.startTime) }}
            </text>
            <text class="workflow-summary__cell workflow-summary__cell--time">
              {{ formatTime(instance.endTime) }}
            </text>
            <view class="workflow-summary__cell workflow-summary__cell--action">
              <text class="workflow-summary__link" @click="openDetail(instance.id)">
                查看
              </text>
              <text v-if="canExport" class="workflow-summary__link" @click="exportInstances([instance.id])">
                导出
              </text>
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
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: 12px 14px;
  align-items: end;
}

.workflow-summary__filter {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 7px;
}

.workflow-summary__filter--definition-name {
  grid-column: 1 / 4;
  grid-row: 1;
}

.workflow-summary__filter--starter {
  grid-column: 4 / 7;
  grid-row: 1;
}

.workflow-summary__filter--status {
  grid-column: 7 / 8;
  grid-row: 1;
}

.workflow-summary__filter--version {
  grid-column: 8 / 9;
  grid-row: 1;
}

.workflow-summary__filter--start {
  grid-column: 1 / 7;
  grid-row: 2;
}

.workflow-summary__filter--end {
  grid-column: 7 / 13;
  grid-row: 2;
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
  grid-column: 11 / 13;
  grid-row: 3;
  width: fit-content;
  justify-self: end;
  justify-content: flex-end;
  gap: 8px;
}

.workflow-summary__toolbar {
  min-height: 42px;
  justify-content: space-between;
  gap: 12px;
}

.workflow-summary__toolbar-main,
.workflow-summary__export-actions {
  gap: 10px;
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
  min-width: 1280px;
}

.workflow-summary__row {
  display: grid;
  grid-template-columns: 66px minmax(140px, 1fr) minmax(190px, 1.45fr) minmax(100px, 0.8fr) minmax(100px, 0.8fr) 76px 94px 168px 168px 112px;
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

@media (max-width: 1200px) {
  .workflow-summary__filters {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }

  .workflow-summary__filter--definition-name {
    grid-column: 1 / 3;
  }

  .workflow-summary__filter--starter {
    grid-column: 3 / 5;
  }

  .workflow-summary__filter--status {
    grid-column: 5 / 6;
  }

  .workflow-summary__filter--version {
    grid-column: 6 / 7;
  }

  .workflow-summary__filter-actions {
    grid-column: 5 / 7;
    grid-row: 3;
  }

  .workflow-summary__filter--start {
    grid-column: 1 / 4;
  }

  .workflow-summary__filter--end {
    grid-column: 4 / 7;
  }
}

@media (max-width: 768px) {
  .workflow-summary {
    padding: 14px 12px 18px;
  }

  .workflow-summary__filters {
    grid-template-columns: 1fr;
  }

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
    align-items: stretch;
    flex-direction: column;
  }

  .workflow-summary__toolbar,
  .workflow-summary__pagination {
    align-items: stretch;
    flex-direction: column;
  }

  .workflow-summary__toolbar-main,
  .workflow-summary__export-actions {
    justify-content: space-between;
  }
}
</style>
