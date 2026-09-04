<script setup lang="ts">
import type { WorkflowHistoryDateFilters } from '../workflow-history-filter'
import type {
  WorkflowInstanceSummary,
  WorkflowPublishedDefinition,
  WorkflowTaskSummary,
} from '@/types/workflow'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  getWorkflowInstance,
  listWorkflowCategories,
  listWorkflowDefinitions,
  listWorkflowInstances,
  listWorkflowTasks,
} from '@/api/workflow'
import { useAppContentStore, useDingtalkAuthStore } from '@/stores'
import { buildWorkflowHistoryTimeQuery } from '../workflow-history-filter'
import {
  workflowInstanceContentKey,
  workflowStartContentKey,
  workflowTaskContentKey,
} from '../workflow-route-keys'
import { workflowInstanceStatusMeta, workflowTaskStatusMeta } from '../workflow-status'
import { workflowTaskTabTitle } from '../workflow-task'
import WorkflowDetailPanel from './WorkflowDetailPanel.vue'
import WorkflowFilterPanel from './WorkflowFilterPanel.vue'
import WorkflowHistoryDatePicker from './WorkflowHistoryDatePicker.vue'
import WorkflowRecordTable from './WorkflowRecordTable.vue'
import WorkflowSummaryPage from './WorkflowSummaryPage.vue'

type WorkflowCenterTab = 'start' | 'pending' | 'handled' | 'started' | 'copied' | 'summary'
type WorkflowListTab = Exclude<WorkflowCenterTab, 'start' | 'summary'>
type WorkflowCategory = 'all' | 'recent' | string

interface WorkflowPaginationChangePayload {
  current: number
}

interface WorkflowRecordFilters extends WorkflowHistoryDateFilters {
  definitionName: string
  definitionCategory: string
  starterName: string
  status: string
}

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const activeTab = ref<WorkflowCenterTab>('start')
const definitions = ref<WorkflowPublishedDefinition[]>([])
const workflowCategories = ref<string[]>([])
const failedDefinitionLogoIds = ref<number[]>([])
const definitionsLoading = ref(false)
const categoriesLoading = ref(false)
const listLoading = ref(false)
const countsLoading = ref(false)
const keyword = ref('')
const selectedCategory = ref<WorkflowCategory>('all')
const workflowListTabs: WorkflowListTab[] = ['pending', 'handled', 'started', 'copied']
const recordFilters = reactive(createRecordFilters())
const appliedRecordFilters = reactive(createRecordFilters())
const instances = ref<WorkflowInstanceSummary[]>([])
const tasks = ref<WorkflowTaskSummary[]>([])
const taskInstances = reactive<Record<string, WorkflowInstanceSummary>>({})
const page = ref(1)
const pageSize = 12
const total = ref(0)
const workflowLoadingStyle = {
  width: '22px',
  height: '22px',
}
const counts = reactive({
  pending: 0,
  handled: 0,
  started: 0,
  copied: 0,
})

const detailVisible = ref(false)
const selectedInstanceId = ref('')

const hasViewPermission = computed(() => auth.hasApiPermission('dingtalk_h5:api:workflow:view'))
const canStart = computed(() => auth.hasApiPermission('dingtalk_h5:api:workflow:start'))
const canSummary = computed(() => (
  auth.hasButtonPermission('dingtalk_h5:button:workflow:summary')
  && auth.hasApiPermission('dingtalk_h5:api:workflow:summary')
))
const hasWorkflowAccess = computed(() => canStart.value || hasViewPermission.value || canSummary.value)

const navigationTabs = computed(() => [
  ...(canStart.value
    ? [{
        key: 'start' as const,
        label: '发起审批',
        icon: 'plus-circle',
      }]
    : []),
  ...(hasViewPermission.value
    ? [
        { key: 'pending' as const, label: '我的待办', icon: 'clock' },
        { key: 'started' as const, label: '我的申请', icon: 'file-text' },
        { key: 'copied' as const, label: '抄送我的', icon: 'email' },
        { key: 'handled' as const, label: '已处理', icon: 'checkmark-circle' },
      ]
    : []),
  ...(canSummary.value
    ? [{ key: 'summary' as const, label: '汇总', icon: 'order' }]
    : []),
])

const definitionMap = computed(() => new Map(definitions.value.map(item => [item.id, item])))
const activeListTab = computed<WorkflowListTab>(() => (
  activeTab.value === 'start' || activeTab.value === 'summary'
    ? 'pending'
    : activeTab.value
))
const activeRecordFilters = computed(() => recordFilters[activeListTab.value])
const activeAppliedRecordFilters = computed(() => appliedRecordFilters[activeListTab.value])
const showStarterNameFilter = computed(() => activeTab.value === 'pending' || activeTab.value === 'handled')
const showStatusFilter = computed(() => activeTab.value === 'started' || activeTab.value === 'copied')
const recordCategoryOptions = computed(() => [
  { value: '', label: '全部分类' },
  ...workflowCategories.value.map(value => ({ value, label: value })),
])

const historyStatusOptions = [
  { label: '全部状态', value: '' },
  { label: '审批中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '已驳回', value: 'rejected' },
  { label: '已撤回', value: 'withdrawn' },
  { label: '已取消', value: 'cancelled' },
]
const applicationFilterCount = computed(() => {
  const filters = activeRecordFilters.value
  return [
    filters.definitionName.trim(),
    filters.definitionCategory,
    filters.startDateFrom,
    filters.startDateTo,
    showStarterNameFilter.value ? filters.starterName.trim() : '',
    showStatusFilter.value ? filters.status : '',
  ].filter(Boolean).length
})

const recentDefinitionIds = ref<number[]>(loadRecentDefinitionIds())
const catalogFilterCount = computed(() => [
  keyword.value.trim(),
  selectedCategory.value === 'all' ? '' : selectedCategory.value,
].filter(Boolean).length)

function loadRecentDefinitionIds() {
  const value = uni.getStorageSync('workflow_recent_definition_ids')
  return Array.isArray(value) ? value.map(Number).filter(Number.isFinite).slice(0, 8) : []
}

const categories = computed(() => {
  const values = [...new Set(definitions.value.map(item => item.category.trim()).filter(Boolean))]
  return [
    ...(recentDefinitionIds.value.length > 0 ? [{ value: 'recent', label: '最近使用' }] : []),
    { value: 'all', label: '全部' },
    ...values.map(value => ({ value, label: value })),
  ]
})

const filteredDefinitions = computed(() => {
  const query = keyword.value.trim().toLowerCase()
  let list = definitions.value
  if (selectedCategory.value === 'recent') {
    const order = recentDefinitionIds.value
    list = list
      .filter(item => order.includes(item.id))
      .sort((left, right) => order.indexOf(left.id) - order.indexOf(right.id))
  }
  else if (selectedCategory.value !== 'all') {
    list = list.filter(item => item.category === selectedCategory.value)
  }
  if (!query)
    return list
  return list.filter((item) => {
    return [item.name, item.key, item.category, item.description]
      .some(value => String(value || '').toLowerCase().includes(query))
  })
})

const groupedDefinitions = computed(() => {
  if (selectedCategory.value !== 'all') {
    return [{
      category: selectedCategory.value === 'recent' ? '最近使用' : selectedCategory.value,
      list: filteredDefinitions.value,
    }]
  }
  const groups = new Map<string, WorkflowPublishedDefinition[]>()
  for (const definition of filteredDefinitions.value) {
    const category = definition.category.trim() || '其他'
    if (!groups.has(category))
      groups.set(category, [])
    groups.get(category)?.push(definition)
  }
  return [...groups.entries()].map(([category, list]) => ({ category, list }))
})

const showStarterColumn = computed(() => ['pending', 'handled', 'copied'].includes(activeTab.value))
const showCurrentProgressColumns = computed(() => activeTab.value === 'started')

const recordColumns = computed(() => [
  { key: 'name', label: '流程名称', width: 'minmax(140px, 1.2fr)' },
  { key: 'businessKey', label: '流程单号', width: 'minmax(180px, 1.35fr)', mobileHidden: true },
  ...(showStarterColumn.value
    ? [{ key: 'starterName', label: '发起人', width: 'minmax(100px, 0.8fr)', mobileHidden: true }]
    : []),
  {
    key: 'context',
    label: activeTab.value === 'pending' ? '当前节点' : '流程分类',
    width: 'minmax(100px, 0.8fr)',
    mobileHidden: true,
  },
  ...(activeTab.value === 'pending'
    ? [{ key: 'assigneeName', label: '节点处理人', width: 'minmax(110px, 0.9fr)', mobileHidden: true }]
    : []),
  ...(showCurrentProgressColumns.value
    ? [
        { key: 'currentNode', label: '当前节点', width: 'minmax(100px, 0.85fr)', mobileHidden: true },
        { key: 'currentAssignees', label: '节点处理人', width: 'minmax(110px, 0.9fr)', mobileHidden: true },
      ]
    : []),
  { key: 'status', label: '审批状态', width: '92px' },
  { key: 'submittedAt', label: '提交时间', width: 'minmax(150px, 1fr)' },
  { key: 'actions', label: '操作', width: '84px' },
])

const recordRows = computed(() => {
  if (activeTab.value === 'pending') {
    return tasks.value.map((task) => {
      const instance = taskInstances[task.instanceId]
      return {
        id: task.id,
        cells: {
          name: taskDefinitionName(task),
          businessKey: instance?.businessKey || task.instanceId,
          starterName: taskStarterDisplayName(task),
          context: task.nodeName || '-',
          assigneeName: task.assigneeName.trim() || '-',
          submittedAt: formatTime(instance?.startTime),
        },
        status: workflowTaskStatusMeta(task.status),
      }
    })
  }

  return instances.value.map(instance => ({
    id: instance.id,
    cells: {
      name: definitionName(instance),
      businessKey: instance.businessKey || '-',
      starterName: starterDisplayName(instance),
      context: definitionCategory(instance),
      currentNode: currentNodeDisplay(instance),
      currentAssignees: currentAssigneeDisplay(instance),
      submittedAt: formatTime(instance.startTime),
    },
    status: workflowInstanceStatusMeta(instance.status),
  }))
})

watch(
  () => appContent.refreshTick,
  () => {
    if (appContent.currentKey === 'workflow')
      void refreshAll()
  },
)

watch(
  () => auth.permissionVersion,
  () => {
    if (!navigationTabs.value.some(tab => tab.key === activeTab.value))
      activeTab.value = defaultWorkflowTab()
  },
)

onMounted(() => {
  const focusedTabOpened = openFocusedWorkflowTab()
  if (!focusedTabOpened)
    activeTab.value = defaultWorkflowTab()
  recentDefinitionIds.value = loadRecentDefinitionIds()
  openFocusedWorkflowInstance()
  void refreshAll()
})

function defaultWorkflowTab(): WorkflowCenterTab {
  if (canStart.value)
    return 'start'
  if (canSummary.value)
    return 'summary'
  return 'pending'
}

function openFocusedWorkflowTab() {
  const requestedTab = appContent.focusedWorkflowTab as WorkflowCenterTab
  if (!requestedTab)
    return false
  appContent.clearFocusedWorkflowTab()
  if (!navigationTabs.value.some(tab => tab.key === requestedTab))
    return false
  activeTab.value = requestedTab
  return true
}

async function refreshAll() {
  if (!hasWorkflowAccess.value)
    return
  if (canStart.value || hasViewPermission.value)
    await loadDefinitions()
  if (hasViewPermission.value)
    await Promise.all([loadWorkflowCategories(), loadCounts(), loadCurrentList()])
}

async function loadDefinitions() {
  if ((!canStart.value && !hasViewPermission.value) || definitionsLoading.value)
    return
  definitionsLoading.value = true
  try {
    const response = await listWorkflowDefinitions()
    definitions.value = Array.isArray(response?.data) ? response.data : []
    workflowCategories.value = [...new Set(definitions.value.map(item => item.category.trim()).filter(Boolean))]
    failedDefinitionLogoIds.value = []
    if (selectedCategory.value === 'recent' && recentDefinitionIds.value.length === 0)
      selectedCategory.value = 'all'
  }
  catch {
    uni.showToast({ title: '流程列表加载失败', icon: 'none' })
  }
  finally {
    definitionsLoading.value = false
  }
}

async function loadWorkflowCategories() {
  if (!hasViewPermission.value || categoriesLoading.value)
    return
  categoriesLoading.value = true
  try {
    const response = await listWorkflowCategories()
    const values = Array.isArray(response?.data)
      ? response.data.map(value => String(value || '').trim()).filter(Boolean)
      : []
    workflowCategories.value = [...new Set(values)]
  }
  catch {
    uni.showToast({ title: '流程分类加载失败', icon: 'none' })
  }
  finally {
    categoriesLoading.value = false
  }
}

async function loadCounts() {
  if (!hasViewPermission.value || countsLoading.value)
    return
  countsLoading.value = true
  try {
    const [pending, handled, started, copied] = await Promise.all([
      listWorkflowTasks({ status: 'pending', page: 1, pageSize: 1 }),
      listWorkflowInstances({ scope: 'handled', page: 1, pageSize: 1 }),
      listWorkflowInstances({ scope: 'started', page: 1, pageSize: 1 }),
      listWorkflowInstances({ scope: 'copied', page: 1, pageSize: 1 }),
    ])
    counts.pending = Number(pending?.data?.total || 0)
    counts.handled = Number(handled?.data?.total || 0)
    counts.started = Number(started?.data?.total || 0)
    counts.copied = Number(copied?.data?.total || 0)
  }
  finally {
    countsLoading.value = false
  }
}

async function loadCurrentList() {
  if (!hasViewPermission.value || !workflowListTabs.includes(activeTab.value as WorkflowListTab) || listLoading.value)
    return
  listLoading.value = true
  try {
    const listTab = activeListTab.value
    const filters = activeAppliedRecordFilters.value
    const timeQuery = buildWorkflowHistoryTimeQuery(filters) || {}
    const definitionName = filters.definitionName.trim() || undefined
    const definitionCategory = filters.definitionCategory || undefined
    if (listTab === 'pending') {
      const response = await listWorkflowTasks({
        status: 'pending',
        definitionName,
        definitionCategory,
        starterName: filters.starterName.trim() || undefined,
        startTimeFrom: timeQuery.startTimeFrom,
        startTimeTo: timeQuery.startTimeTo,
        page: page.value,
        pageSize,
      })
      tasks.value = Array.isArray(response?.data?.list) ? response.data.list : []
      instances.value = []
      total.value = Number(response?.data?.total || 0)
      await loadTaskInstances(tasks.value)
    }
    else {
      const response = await listWorkflowInstances({
        scope: listTab,
        definitionName,
        definitionCategory,
        starterName: listTab === 'handled' ? filters.starterName.trim() || undefined : undefined,
        status: showStatusFilter.value ? filters.status || undefined : undefined,
        ...timeQuery,
        page: page.value,
        pageSize,
      })
      instances.value = Array.isArray(response?.data?.list) ? response.data.list : []
      tasks.value = []
      total.value = Number(response?.data?.total || 0)
    }
  }
  catch {
    uni.showToast({ title: '流程数据加载失败', icon: 'none' })
  }
  finally {
    listLoading.value = false
  }
}

async function loadTaskInstances(list: WorkflowTaskSummary[]) {
  const missing = [...new Set(list.map(task => task.instanceId))]
    .filter(instanceId => !taskInstances[instanceId])
  await Promise.all(missing.map(async (instanceId) => {
    try {
      const response = await getWorkflowInstance(instanceId)
      if (response?.data?.instance)
        taskInstances[instanceId] = response.data.instance
    }
    catch {
      // Task cards remain usable with task metadata when the detail lookup fails.
    }
  }))
}

function switchTab(tab: WorkflowCenterTab) {
  if (tab === activeTab.value)
    return
  activeTab.value = tab
  page.value = 1
  if (workflowListTabs.includes(tab as WorkflowListTab))
    void loadCurrentList()
}

function emptyRecordFilters(): WorkflowRecordFilters {
  return {
    definitionName: '',
    definitionCategory: '',
    starterName: '',
    status: '',
    startDateFrom: '',
    startDateTo: '',
  }
}

function createRecordFilters() {
  return Object.fromEntries(
    workflowListTabs.map(tab => [tab, emptyRecordFilters()]),
  ) as Record<WorkflowListTab, WorkflowRecordFilters>
}

function queryRecords() {
  if (listLoading.value)
    return
  if (buildWorkflowHistoryTimeQuery(activeRecordFilters.value) === null) {
    uni.showToast({ title: '请检查提交时间范围', icon: 'none' })
    return
  }
  Object.assign(appliedRecordFilters[activeListTab.value], activeRecordFilters.value)
  page.value = 1
  void loadCurrentList()
}

function resetRecordFilters() {
  if (listLoading.value)
    return
  Object.assign(recordFilters[activeListTab.value], emptyRecordFilters())
  Object.assign(appliedRecordFilters[activeListTab.value], emptyRecordFilters())
  page.value = 1
  void loadCurrentList()
}

function openWorkflowStartTab(definition: WorkflowPublishedDefinition) {
  if (!canStart.value) {
    uni.showToast({ title: '无流程发起权限', icon: 'none' })
    return
  }
  const key = workflowStartContentKey(definition.id)
  appContent.openDynamicTab({
    key,
    label: definition.name,
    icon: 'file-text',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  })
}

function definitionStartMeta(definition: WorkflowPublishedDefinition) {
  const version = `版本 ${definition.version}`
  if (definition.startLimit?.mode !== 'limited')
    return version
  const status = definition.startLimitStatus
  if (!status)
    return version
  if (!status.allowed)
    return `${version} · 当前周期已达上限`
  return `${version} · 当前周期剩余 ${status.remainingCount} 次`
}

function hasDefinitionLogo(definition: WorkflowPublishedDefinition) {
  return Boolean(String(definition.logoUrl || '').trim())
    && !failedDefinitionLogoIds.value.includes(definition.id)
}

function markDefinitionLogoFailed(definitionId: number) {
  if (failedDefinitionLogoIds.value.includes(definitionId))
    return
  failedDefinitionLogoIds.value = [...failedDefinitionLogoIds.value, definitionId]
}

function openFocusedWorkflowInstance() {
  const instanceId = appContent.focusedWorkflowInstanceId
  if (!instanceId)
    return
  appContent.clearFocusedWorkflowInstance()
  openInstance(instanceId)
}

function openWorkflowTaskTab(task: WorkflowTaskSummary) {
  const key = workflowTaskContentKey(task.id, task.instanceId)
  if (!key)
    return
  appContent.openDynamicTab({
    key,
    label: workflowTaskTabTitle(taskStarterDisplayName(task), taskDefinitionName(task)),
    icon: 'checkmark-circle',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  })
}

function openWorkflowInstanceTab(instanceId: string) {
  const key = workflowInstanceContentKey(instanceId)
  if (!key)
    return
  const instance = instances.value.find(item => item.id === instanceId)
  appContent.openDynamicTab({
    key,
    label: instance ? definitionName(instance) : '流程详情',
    icon: 'eye',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  })
}

function resolveMobilePage() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    if (width > 0)
      return width <= 768
  }
  catch {
    // 读取失败时继续按 H5 指针类型判断。
  }

  // #ifdef H5
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function')
    return window.matchMedia('(hover: none) and (pointer: coarse)').matches
  // #endif

  return false
}

function openInstance(instanceId: string) {
  if (resolveMobilePage()) {
    openWorkflowInstanceTab(instanceId)
    return
  }
  selectedInstanceId.value = instanceId
  detailVisible.value = true
}

function openRecord(recordId: string) {
  if (activeTab.value === 'pending') {
    const task = tasks.value.find(item => item.id === recordId)
    if (task)
      openWorkflowTaskTab(task)
    return
  }
  openInstance(recordId)
}

function definitionCategory(instance: WorkflowInstanceSummary) {
  return definitionMap.value.get(instance.definitionId)?.category?.trim() || '其他'
}

function currentNodeDisplay(instance: WorkflowInstanceSummary) {
  return (instance.currentNodeNames || []).map(name => name.trim()).filter(Boolean).join('、') || '-'
}

function currentAssigneeDisplay(instance: WorkflowInstanceSummary) {
  return (instance.currentAssigneeNames || []).map(name => name.trim()).filter(Boolean).join('、') || '-'
}

async function handleDetailChanged() {
  await Promise.all([loadCounts(), loadCurrentList()])
}

function handlePageChange(payload: WorkflowPaginationChangePayload) {
  page.value = Number(payload.current || 1)
  void loadCurrentList()
}

function definitionName(instance?: WorkflowInstanceSummary) {
  if (!instance)
    return '流程审批'
  return definitionMap.value.get(instance.definitionId)?.name || instance.definitionKey || '流程审批'
}

function taskDefinitionName(task: WorkflowTaskSummary) {
  return String(task.definitionName || '').trim() || definitionName(taskInstances[task.instanceId])
}

function taskStarterDisplayName(task: WorkflowTaskSummary) {
  return String(task.starterName || '').trim() || starterDisplayName(taskInstances[task.instanceId])
}

function starterDisplayName(instance?: WorkflowInstanceSummary) {
  return String(instance?.starterName || '').trim() || '未知用户'
}

function formatTime(timestamp?: number) {
  if (!timestamp)
    return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function activeListTitle() {
  const titles: Record<WorkflowCenterTab, string> = {
    start: '发起审批',
    pending: '我的待办',
    handled: '已处理',
    started: '我的申请',
    copied: '抄送我的',
    summary: '汇总',
  }
  return titles[activeTab.value]
}
</script>

<template>
  <view class="workflow-center">
    <view v-if="!hasWorkflowAccess" class="workflow-center__permission-empty">
      <u-icon name="lock" size="32px" color="#c8c9cc" />
      <text class="workflow-center__empty-title">
        暂无流程查看权限
      </text>
      <text>请联系管理员配置工作流菜单和接口权限。</text>
    </view>

    <template v-else>
      <view class="workflow-center__nav">
        <view
          v-for="tab in navigationTabs"
          :key="tab.key"
          class="workflow-center__nav-item"
          :class="{ active: activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          <u-icon :name="tab.icon" size="16px" :color="activeTab === tab.key ? '#1f2329' : '#86909c'" />
          <text>{{ tab.label }}</text>
        </view>
      </view>

      <view v-if="hasViewPermission" class="workflow-center__quick-links">
        <view class="workflow-center__quick-link workflow-center__quick-link--amber" @click="switchTab('pending')">
          <view>
            <text class="workflow-center__quick-value">
              {{ counts.pending }}
            </text>
            <text>待我处理</text>
          </view>
          <u-icon name="arrow-right" size="14px" color="#b7791f" />
        </view>
        <view class="workflow-center__quick-link workflow-center__quick-link--teal" @click="switchTab('handled')">
          <view>
            <text class="workflow-center__quick-value">
              {{ counts.handled }}
            </text>
            <text>已处理</text>
          </view>
          <u-icon name="arrow-right" size="14px" color="#0f766e" />
        </view>
        <view class="workflow-center__quick-link workflow-center__quick-link--blue" @click="switchTab('started')">
          <view>
            <text class="workflow-center__quick-value">
              {{ counts.started }}
            </text>
            <text>我发起的</text>
          </view>
          <u-icon name="arrow-right" size="14px" color="#2563eb" />
        </view>
        <view class="workflow-center__quick-link workflow-center__quick-link--violet" @click="switchTab('copied')">
          <view>
            <text class="workflow-center__quick-value">
              {{ counts.copied }}
            </text>
            <text>抄送我的</text>
          </view>
          <u-icon name="arrow-right" size="14px" color="#7c3aed" />
        </view>
      </view>

      <view v-if="activeTab === 'start'" class="workflow-center__catalog">
        <WorkflowFilterPanel :active-count="catalogFilterCount">
          <view class="workflow-center__catalog-filters">
            <view class="workflow-center__search">
              <u-search
                v-model="keyword"
                custom-class="workflow-center__search-control"
                placeholder="搜索流程名称、分类或说明"
                :show-action="false"
                bg-color="#ffffff"
              />
            </view>
            <scroll-view class="workflow-center__categories" scroll-x>
              <view class="workflow-center__category-row">
                <view
                  v-for="category in categories"
                  :key="category.value"
                  class="workflow-center__category"
                  :class="{ active: selectedCategory === category.value }"
                  @click="selectedCategory = category.value"
                >
                  {{ category.label }}
                </view>
              </view>
            </scroll-view>
          </view>
        </WorkflowFilterPanel>

        <view v-if="definitionsLoading" class="workflow-center__loading">
          <u-loading custom-class="workflow-center__loading-icon" :custom-style="workflowLoadingStyle" mode="circle" size="42" />
          <text>加载流程...</text>
        </view>
        <view v-else-if="filteredDefinitions.length === 0" class="workflow-center__empty">
          <u-icon name="file-text" size="32px" color="#c8c9cc" />
          <text class="workflow-center__empty-title">
            暂无可发起流程
          </text>
          <text>没有匹配当前搜索和发起范围的流程。</text>
        </view>
        <view v-else class="workflow-center__groups">
          <section v-for="group in groupedDefinitions" :key="group.category" class="workflow-center__group">
            <view class="workflow-center__group-title">
              <text>{{ group.category }}</text>
              <text>{{ group.list.length }}</text>
            </view>
            <view class="workflow-center__definition-grid">
              <view
                v-for="(definition, index) in group.list"
                :key="definition.id"
                class="workflow-definition"
                @click="openWorkflowStartTab(definition)"
              >
                <view class="workflow-definition__icon" :class="`tone-${index % 5}`">
                  <image
                    v-if="hasDefinitionLogo(definition)"
                    class="workflow-definition__logo"
                    :src="definition.logoUrl"
                    :alt="definition.name"
                    mode="aspectFill"
                    @error="markDefinitionLogoFailed(definition.id)"
                  />
                  <u-icon v-else name="file-text" size="18px" color="#ffffff" />
                </view>
                <view class="workflow-definition__copy">
                  <text class="workflow-definition__name">
                    {{ definition.name }}
                  </text>
                  <text class="workflow-definition__desc">
                    {{ definitionStartMeta(definition) }}
                  </text>
                </view>
                <u-icon name="arrow-right" size="14px" color="#c3cad4" />
              </view>
            </view>
          </section>
        </view>
      </view>

      <WorkflowSummaryPage v-else-if="activeTab === 'summary'" />

      <view v-else class="workflow-center__list-section">
        <view class="workflow-center__list-head workflow-center__list-head--mobile-hidden">
          <view>
            <text class="workflow-center__list-title">
              {{ activeListTitle() }}
            </text>
            <text class="workflow-center__list-total">
              共 {{ total }} 条
            </text>
          </view>
        </view>

        <WorkflowFilterPanel :active-count="applicationFilterCount">
          <view class="workflow-center__record-filters">
            <view class="workflow-center__filter-field">
              <text class="workflow-center__filter-label">
                流程名称
              </text>
              <input
                v-model="activeRecordFilters.definitionName"
                class="workflow-center__filter-input"
                type="text"
                :maxlength="50"
                placeholder="输入流程名称"
                :disabled="listLoading"
                @keyup.enter="queryRecords"
              >
            </view>
            <view class="workflow-center__filter-field">
              <text class="workflow-center__filter-label">
                流程分类
              </text>
              <!-- H5 PC 使用原生选择控件，避免 u-select 在桌面端弹出底部选择层。 -->
              <select
                v-model="activeRecordFilters.definitionCategory"
                class="workflow-center__filter-select"
                :disabled="listLoading || categoriesLoading"
                aria-label="流程分类"
              >
                <option v-for="option in recordCategoryOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </view>
            <view class="workflow-center__filter-field workflow-center__filter-field--date">
              <text class="workflow-center__filter-label">
                提交时间
              </text>
              <view class="workflow-center__filter-date-range">
                <WorkflowHistoryDatePicker
                  v-model="activeRecordFilters.startDateFrom"
                  :disabled="listLoading"
                  placeholder="开始日期"
                />
                <text class="workflow-center__filter-separator">
                  至
                </text>
                <WorkflowHistoryDatePicker
                  v-model="activeRecordFilters.startDateTo"
                  :disabled="listLoading"
                  placeholder="结束日期"
                />
              </view>
            </view>
            <view v-if="showStarterNameFilter" class="workflow-center__filter-field">
              <text class="workflow-center__filter-label">
                发起人
              </text>
              <input
                v-model="activeRecordFilters.starterName"
                class="workflow-center__filter-input"
                type="text"
                :maxlength="50"
                placeholder="输入发起人用户名"
                :disabled="listLoading"
                @keyup.enter="queryRecords"
              >
            </view>
            <view v-if="showStatusFilter" class="workflow-center__filter-field">
              <text class="workflow-center__filter-label">
                审批状态
              </text>
              <select
                v-model="activeRecordFilters.status"
                class="workflow-center__filter-select"
                :disabled="listLoading"
                aria-label="审批状态"
              >
                <option v-for="option in historyStatusOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </view>
            <view class="workflow-center__filter-actions">
              <u-button custom-class="workflow-center__filter-action" size="small" plain :disabled="listLoading" @click="resetRecordFilters">
                重置
              </u-button>
              <u-button custom-class="workflow-center__filter-action" size="small" type="primary" :loading="listLoading" @click="queryRecords">
                查询
              </u-button>
            </view>
          </view>
        </WorkflowFilterPanel>

        <view v-if="listLoading" class="workflow-center__loading">
          <u-loading custom-class="workflow-center__loading-icon" :custom-style="workflowLoadingStyle" mode="circle" size="42" />
          <text>加载中...</text>
        </view>
        <view v-else-if="activeTab === 'pending' && tasks.length === 0" class="workflow-center__empty">
          <u-icon name="checkmark-circle" size="32px" color="#9ca3af" />
          <text class="workflow-center__empty-title">
            暂无待办
          </text>
          <text>当前没有需要处理的流程任务。</text>
        </view>
        <view v-else-if="activeTab !== 'pending' && instances.length === 0" class="workflow-center__empty">
          <u-icon name="order" size="32px" color="#9ca3af" />
          <text class="workflow-center__empty-title">
            暂无流程记录
          </text>
        </view>
        <WorkflowRecordTable
          v-else
          :columns="recordColumns"
          :rows="recordRows"
        >
          <template #actions="{ row }">
            <u-button
              custom-class="workflow-record__action workflow-record__action--view"
              size="mini"
              plain
              @click.stop="openRecord(row.id)"
            >
              <view class="workflow-record__action-content">
                <u-icon v-if="activeTab === 'pending'" name="arrow-right" size="14px" color="#2563eb" />
                <u-icon v-else name="eye" size="14px" color="#2563eb" />
                <text>{{ activeTab === 'pending' ? '办理' : '查看' }}</text>
              </view>
            </u-button>
          </template>
        </WorkflowRecordTable>

        <view v-if="total > pageSize" class="workflow-center__pagination">
          <u-pagination
            v-model="page"
            custom-class="workflow-center__pagination-control"
            :total="total"
            :page-size="pageSize"
            prev-text="上一页"
            next-text="下一页"
            @change="handlePageChange"
          />
        </view>
      </view>
    </template>

    <WorkflowDetailPanel
      v-model="detailVisible"
      :instance-id="selectedInstanceId"
      :definitions="definitions"
      presentation="history-drawer"
      :application-actions="activeTab === 'started'"
      comment-action
      @changed="handleDetailChanged"
    />
  </view>
</template>

<style lang="scss" scoped>
.workflow-center {
  min-height: 100%;
  padding: 0 0 36px;
  background: #f5f7fa;
  color: #1f2329;
  box-sizing: border-box;
}

.workflow-center__nav {
  min-height: 54px;
  padding: 0 24px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: stretch;
  gap: 30px;
  overflow-x: auto;
  background: #fff;
}

.workflow-center__nav-item {
  flex: 0 0 auto;
  border-bottom: 2px solid transparent;
  display: flex;
  align-items: center;
  gap: 7px;
  color: #86909c;
  font-size: 14px;
  cursor: pointer;
}

.workflow-center__nav-item.active {
  border-bottom-color: #1f2329;
  color: #1f2329;
  font-weight: 700;
}

.workflow-center__quick-links {
  padding: 18px 24px;
  border-bottom: 1px solid #e5eaf3;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  background: #fff;
}

.workflow-center__quick-link {
  min-width: 0;
  min-height: 62px;
  padding: 12px 14px;
  border: 1px solid #e5eaf3;
  border-left-width: 4px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  background: #fff;
  color: #4e5969;
  font-size: 12px;
  cursor: pointer;
  box-sizing: border-box;
}

.workflow-center__quick-link > view {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.workflow-center__quick-value {
  color: #1f2329;
  font-size: 23px;
  font-weight: 800;
}

.workflow-center__quick-link--amber { border-left-color: #d69e2e; }
.workflow-center__quick-link--teal { border-left-color: #0f766e; }
.workflow-center__quick-link--blue { border-left-color: #2563eb; }
.workflow-center__quick-link--violet { border-left-color: #7c3aed; }

.workflow-center__catalog,
.workflow-center__list-section {
  padding: 22px 24px 0;
}

.workflow-center__search {
  width: min(560px, 100%);
  margin-bottom: 16px;
}

:deep(.workflow-center__search-control .u-content) {
  height: 36px !important;
  padding: 0 10px;
  border: 1px solid #d8e0e8 !important;
  border-radius: 4px !important;
  box-sizing: border-box;
}

:deep(.workflow-center__search-control .u-input) {
  margin: 0 7px;
  font-size: 13px;
}

:deep(.workflow-center__search-control .u-icon__icon) {
  font-size: 15px !important;
}

.workflow-center__catalog-filters {
  min-width: 0;
}

.workflow-center__categories {
  width: 100%;
  margin-bottom: 0;
  white-space: nowrap;
}

.workflow-center__category-row {
  display: inline-flex;
  align-items: center;
  gap: 24px;
}

.workflow-center__category {
  padding: 0 0 9px;
  border-bottom: 2px solid transparent;
  color: #86909c;
  font-size: 13px;
  cursor: pointer;
}

.workflow-center__category.active {
  border-bottom-color: #1f2329;
  color: #1f2329;
  font-weight: 700;
}

.workflow-center__groups {
  display: grid;
  gap: 28px;
}

.workflow-center__group-title {
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #4e5969;
  font-size: 13px;
  font-weight: 700;
}

.workflow-center__group-title > text:last-child {
  color: #9ca3af;
  font-weight: 500;
}

.workflow-center__definition-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.workflow-definition {
  min-width: 0;
  min-height: 78px;
  padding: 14px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 12px;
  background: #fff;
  cursor: pointer;
  box-sizing: border-box;
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.workflow-definition:hover {
  border-color: #a8b4c5;
  box-shadow: 0 5px 16px rgba(15, 23, 42, 0.07);
}

.workflow-definition__icon {
  flex: 0 0 auto;
  width: 42px;
  height: 42px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.workflow-definition__icon.tone-0 { background: #0f766e; }
.workflow-definition__icon.tone-1 { background: #2563eb; }
.workflow-definition__icon.tone-2 { background: #d97706; }
.workflow-definition__icon.tone-3 { background: #7c3aed; }
.workflow-definition__icon.tone-4 { background: #dc5a46; }

.workflow-definition__logo {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  display: block;
}

.workflow-definition__copy {
  min-width: 0;
  flex: 1 1 auto;
}

.workflow-definition__name,
.workflow-definition__desc {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-definition__name {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.workflow-definition__desc {
  margin-top: 5px;
  color: #86909c;
  font-size: 11px;
}

.workflow-center__list-head {
  margin-bottom: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.workflow-center__list-head > view:first-child {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.workflow-center__list-title {
  font-size: 17px;
  font-weight: 800;
}

.workflow-center__list-total {
  color: #86909c;
  font-size: 12px;
}

.workflow-center__record-filters {
  display: grid;
  grid-template-columns: minmax(180px, 240px) minmax(150px, 180px) minmax(300px, 430px) minmax(160px, 220px) auto;
  align-items: end;
  justify-content: start;
  gap: 12px;
}

.workflow-center__filter-field {
  min-width: 0;
  display: grid;
  gap: 7px;
}

.workflow-center__filter-label {
  color: #4e5969;
  font-size: 12px;
  font-weight: 600;
}

.workflow-center__filter-select,
.workflow-center__filter-input {
  width: 100%;
  height: 36px;
  min-width: 0;
  padding: 0 11px;
  border: 1px solid #d8e0e8;
  border-radius: 4px;
  background: #ffffff;
  color: #1f2329;
  font-family: inherit;
  font-size: 13px;
  letter-spacing: 0;
  box-sizing: border-box;
}

.workflow-center__filter-select {
  cursor: pointer;
}

.workflow-center__filter-input {
  cursor: text;
}

.workflow-center__filter-select:focus,
.workflow-center__filter-input:focus {
  border-color: #0f766e;
  outline: none;
  box-shadow: 0 0 0 2px rgba(15, 118, 110, 0.12);
}

.workflow-center__filter-select:disabled,
.workflow-center__filter-input:disabled {
  background: #f2f3f5;
  color: #a9b0bb;
  cursor: not-allowed;
}

.workflow-center__filter-input::placeholder {
  color: #a9b0bb;
}

.workflow-center__filter-date-range {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 7px;
}

.workflow-center__filter-separator {
  color: #86909c;
  font-size: 12px;
}

.workflow-center__filter-actions {
  grid-column: -2 / -1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.workflow-center__filter-action,
:deep(.workflow-center__filter-action) {
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

.workflow-record__action,
:deep(.workflow-record__action) {
  width: auto;
  min-width: 70px;
  height: 30px;
  min-height: 30px;
  margin: 0;
  padding: 0 10px;
  border-radius: 4px;
}

.workflow-record__action-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 12px;
}

:deep(.workflow-center__loading-icon) {
  width: 22px !important;
  height: 22px !important;
}

.workflow-record__action--view,
:deep(.workflow-record__action--view) {
  border-color: #bfd0f7;
  color: #2563eb;
}

.workflow-center__pagination {
  padding: 18px 0;
  display: flex;
  justify-content: flex-end;
}

:deep(.workflow-center__pagination-control) {
  width: auto;
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.workflow-center__pagination-control .u-pagination-text),
:deep(.workflow-center__pagination-control .u-pagination-text text),
:deep(.workflow-center__pagination-control .u-pagination-text .u-text__value) {
  color: #5f6b7a !important;
  font-size: 12px !important;
  line-height: 28px;
  white-space: nowrap;
}

:deep(.workflow-center__pagination-control .custom-class) {
  width: auto !important;
  min-width: 48px !important;
  height: 28px !important;
  min-height: 28px !important;
  margin: 0 !important;
  padding: 0 8px !important;
  border-radius: 4px !important;
  font-size: 12px !important;
  line-height: 28px !important;
}

.workflow-center__loading,
.workflow-center__empty,
.workflow-center__permission-empty {
  min-height: 320px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #86909c;
  font-size: 13px;
  text-align: center;
}

.workflow-center__permission-empty {
  min-height: 520px;
}

.workflow-center__empty-title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

@media screen and (max-width: 1200px) {
  .workflow-center__definition-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media screen and (max-width: 900px) {
  .workflow-center__definition-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workflow-center__record-filters {
    grid-template-columns: minmax(150px, 180px) minmax(0, 1fr);
  }

  .workflow-center__filter-field--date {
    grid-column: 1 / -1;
  }

  .workflow-center__filter-actions {
    grid-column: auto;
    justify-content: flex-end;
  }

}

@media screen and (max-width: 768px) {
  .workflow-center__list-head--mobile-hidden {
    display: none;
  }

  .workflow-center__nav {
    min-height: 50px;
    padding: 0 14px;
    gap: 22px;
  }

  .workflow-center__nav-item {
    font-size: 13px;
  }

  .workflow-center__quick-links {
    padding: 12px 14px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .workflow-center__quick-link {
    min-height: 54px;
    padding: 9px 10px;
  }

  .workflow-center__quick-value {
    font-size: 19px;
  }

  .workflow-center__catalog,
  .workflow-center__list-section {
    padding: 16px 12px 0;
  }

  .workflow-center__definition-grid {
    grid-template-columns: 1fr;
  }

  .workflow-center__record-filters {
    grid-template-columns: 1fr;
  }

  .workflow-center__filter-field--date {
    grid-column: auto;
  }

  .workflow-center__filter-actions {
    grid-column: auto;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workflow-center__filter-action,
  :deep(.workflow-center__filter-action) {
    width: 100%;
  }

  .workflow-definition {
    min-height: 72px;
  }

  .workflow-center__pagination {
    justify-content: center;
  }

}
</style>
