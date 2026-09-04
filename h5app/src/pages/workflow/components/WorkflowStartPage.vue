<script setup lang="ts">
import type { WorkflowHistoryDateFilters } from '../workflow-history-filter'
import type {
  WorkflowFieldAccessMap,
  WorkflowFieldActionsMap,
  WorkflowFormData,
  WorkflowInstanceSummary,
  WorkflowPublishedDefinition,
  WorkflowStartDraft,
} from '@/types/workflow'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  deleteWorkflowStartDraft,
  getWorkflowDefinition,
  getWorkflowStartDraft,
  listWorkflowInstances,
  saveWorkflowStartDraft,
  startWorkflowInstance,
} from '@/api/workflow'
import { useAppContentStore, useDingtalkAuthStore } from '@/stores'
import {
  initialWorkflowFormData,
  workflowFieldAccessMap,
  workflowFieldActionsMap,
  writableWorkflowFormData,
} from '../workflow-form'
import { buildWorkflowHistoryTimeQuery } from '../workflow-history-filter'
import { workflowDefinitionIdFromContentKey, workflowInstanceContentKey } from '../workflow-route-keys'
import { workflowInstanceStatusMeta } from '../workflow-status'
import WorkflowDetailPanel from './WorkflowDetailPanel.vue'
import WorkflowFilterPanel from './WorkflowFilterPanel.vue'
import WorkflowHistoryDatePicker from './WorkflowHistoryDatePicker.vue'
import WorkflowReadOnlyGraph from './WorkflowReadOnlyGraph.vue'
import WorkflowRuntimeForm from './WorkflowRuntimeForm.vue'

type WorkflowStartSection = 'start' | 'history' | 'drafts' | 'flow'

interface RuntimeFormExposed {
  validate: () => { valid: boolean, errors: Record<string, string> }
}

interface WorkflowPaginationChangePayload {
  current: number
}

interface WorkflowHistoryFilters extends WorkflowHistoryDateFilters {
  status: string
}

const props = defineProps<{
  contentKey: string
}>()

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const definition = ref<WorkflowPublishedDefinition | null>(null)
const activeSection = ref<WorkflowStartSection>('start')
const formData = ref<WorkflowFormData>({})
const formRef = ref<RuntimeFormExposed | null>(null)
const loading = ref(true)
const loadError = ref('')
const draftSaving = ref(false)
const draftDeleting = ref(false)
const submitting = ref(false)
const savedDraft = ref<WorkflowStartDraft | null>(null)
const starterAllowed = ref(false)
const savedSnapshot = ref('')
const historyLoading = ref(false)
const historyLoaded = ref(false)
const historyInstances = ref<WorkflowInstanceSummary[]>([])
const historyPage = ref(1)
const historyPageSize = 10
const historyTotal = ref(0)
const historyFilters = ref<WorkflowHistoryFilters>(emptyHistoryFilters())
const appliedHistoryFilters = ref<WorkflowHistoryFilters>(emptyHistoryFilters())
const detailVisible = ref(false)
const selectedInstanceId = ref('')
let unregisterCloseGuard: (() => void) | null = null

const definitionId = computed(() => workflowDefinitionIdFromContentKey(props.contentKey))
const canStart = computed(() => auth.hasApiPermission('dingtalk_h5:api:workflow:start'))
const canView = computed(() => auth.hasApiPermission('dingtalk_h5:api:workflow:view'))
const busy = computed(() => loading.value || draftSaving.value || draftDeleting.value || submitting.value)
const startLimitExceeded = computed(() => (
  definition.value?.startLimit?.mode === 'limited'
  && definition.value.startLimitStatus?.allowed === false
))
const startLimitMessage = computed(() => {
  const value = definition.value
  if (!value || value.startLimit?.mode !== 'limited')
    return ''
  const status = value.startLimitStatus
  if (!status)
    return ''
  if (!status.allowed)
    return '当前周期的发起次数已用完'
  return `当前周期还可发起 ${status.remainingCount} 次`
})
const navigationItems = computed<Array<{ key: WorkflowStartSection, label: string, icon: string }>>(() => [
  ...(starterAllowed.value ? [{ key: 'start' as const, label: '发起审批', icon: 'plus-circle' }] : []),
  ...(canView.value ? [{ key: 'history' as const, label: '历史记录', icon: 'clock' }] : []),
  ...(starterAllowed.value ? [{ key: 'drafts' as const, label: '草稿箱', icon: 'file-text' }] : []),
  { key: 'flow' as const, label: '流程图', icon: 'share' },
])
const historyStatusOptions = [
  { label: '全部状态', value: '' },
  { label: '审批中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '已驳回', value: 'rejected' },
  { label: '已撤回', value: 'withdrawn' },
  { label: '已取消', value: 'cancelled' },
]
const historyFilterCount = computed(() => [
  historyFilters.value.status,
  historyFilters.value.startDateFrom,
  historyFilters.value.startDateTo,
  historyFilters.value.endDateFrom,
  historyFilters.value.endDateTo,
].filter(Boolean).length)
const fieldAccess = computed<WorkflowFieldAccessMap>(() => {
  const value = definition.value
  if (!value)
    return {}
  return workflowFieldAccessMap(
    value.form || [],
    value.fieldPermissions?.[value.startNodeId || 'start'] || [],
    'write',
  )
})
const fieldActions = computed<WorkflowFieldActionsMap>(() => {
  const value = definition.value
  if (!value)
    return {}
  return workflowFieldActionsMap(
    value.form || [],
    value.fieldPermissions?.[value.startNodeId || 'start'] || [],
    { add: true, delete: true },
  )
})
const hasUnsavedChanges = computed(() => {
  return Boolean(definition.value && savedSnapshot.value && currentSnapshot() !== savedSnapshot.value)
})
const hasSavedDraft = computed(() => Boolean(savedDraft.value))
const draftCount = computed(() => hasSavedDraft.value ? 1 : 0)
const draftUpdatedAt = computed(() => Number(savedDraft.value?.updatedAt || 0))
const draftIsLegacy = computed(() => Boolean(
  savedDraft.value
  && definition.value
  && savedDraft.value.definitionVersion !== definition.value.version,
))
onMounted(() => {
  unregisterCloseGuard = appContent.registerTabCloseGuard(props.contentKey, {
    hasUnsavedChanges: () => hasUnsavedChanges.value,
    saveDraft,
  })
  void loadStartPage()
})

onBeforeUnmount(() => {
  unregisterCloseGuard?.()
})

watch(
  () => appContent.workflowStartSeedTick,
  () => applyWorkflowStartSeed(),
)

function currentWritableData() {
  const value = definition.value
  if (!value)
    return {}
  return writableWorkflowFormData(value.form || [], formData.value, fieldAccess.value)
}

function currentSnapshot() {
  return JSON.stringify(currentWritableData())
}

function businessKey(value: WorkflowPublishedDefinition) {
  return `h5-${value.key}-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`
}

function rememberDefinition(id: number) {
  const stored = uni.getStorageSync('workflow_recent_definition_ids')
  const current = Array.isArray(stored) ? stored.map(Number).filter(Number.isFinite) : []
  const ids = [id, ...current.filter(item => item !== id)].slice(0, 8)
  uni.setStorageSync('workflow_recent_definition_ids', ids)
}

async function loadStartPage() {
  loading.value = true
  loadError.value = ''
  savedDraft.value = null
  starterAllowed.value = false
  try {
    if (!definitionId.value)
      throw new Error('流程定义参数无效')

    let loadedDefinition: WorkflowPublishedDefinition | null = null
    if (canStart.value) {
      try {
        const response = await getWorkflowDefinition(definitionId.value)
        loadedDefinition = response?.data || null
        starterAllowed.value = Boolean(loadedDefinition)
      }
      catch {
        starterAllowed.value = false
      }
    }
    if (!loadedDefinition)
      throw new Error('流程定义不存在或已停用')

    definition.value = loadedDefinition
    formData.value = initialWorkflowFormData(loadedDefinition.form || [])

    if (starterAllowed.value) {
      try {
        const draftResponse = await getWorkflowStartDraft(loadedDefinition.id)
        const draft = draftResponse?.data
        savedDraft.value = draft || null
      }
      catch {
        uni.showToast({ title: '草稿加载失败', icon: 'none' })
      }
    }
    savedSnapshot.value = currentSnapshot()
    applyWorkflowStartSeed()
  }
  catch (error) {
    loadError.value = error instanceof Error ? error.message : '流程表单加载失败'
  }
  finally {
    loading.value = false
  }
}

function applyWorkflowStartSeed() {
  const value = definition.value
  if (!value)
    return false
  const seed = appContent.takeWorkflowStartSeed(value.id)
  if (!seed)
    return false
  formData.value = initialWorkflowFormData(value.form || [], seed.formData)
  activeSection.value = 'start'
  uni.showToast({ title: '已复制原申请内容，请修改后提交', icon: 'none' })
  return true
}

function continueSavedDraft() {
  const value = definition.value
  const draft = savedDraft.value
  if (!value || !draft || busy.value)
    return

  formData.value = initialWorkflowFormData(value.form || [], draft.formData || {})
  savedSnapshot.value = currentSnapshot()
  activeSection.value = 'start'
  if (draft.definitionVersion !== value.version) {
    uni.showToast({ title: '已按当前版本恢复草稿，请检查表单变化', icon: 'none' })
  }
}

function confirmDeleteSavedDraft() {
  if (!savedDraft.value || busy.value)
    return
  uni.showModal({
    title: '删除草稿',
    content: '删除后无法恢复，确认删除这份草稿吗？',
    confirmText: '删除',
    confirmColor: '#ef4444',
    success: (result) => {
      if (result.confirm)
        void deleteSavedDraft()
    },
  })
}

async function deleteSavedDraft() {
  const value = definition.value
  if (!value || !savedDraft.value || draftDeleting.value)
    return
  draftDeleting.value = true
  try {
    await deleteWorkflowStartDraft(value.id)
    savedDraft.value = null
    uni.showToast({ title: '草稿已删除', icon: 'success' })
  }
  catch (error) {
    uni.showToast({ title: workflowRequestErrorMessage(error, '草稿删除失败'), icon: 'none' })
  }
  finally {
    draftDeleting.value = false
  }
}

function switchSection(section: WorkflowStartSection) {
  activeSection.value = section
  if (section === 'history' && !historyLoaded.value)
    void loadHistory()
}

function emptyHistoryFilters(): WorkflowHistoryFilters {
  return {
    status: '',
    startDateFrom: '',
    startDateTo: '',
    endDateFrom: '',
    endDateTo: '',
  }
}

function queryHistory() {
  if (historyLoading.value)
    return
  if (buildWorkflowHistoryTimeQuery(historyFilters.value) === null) {
    uni.showToast({ title: '请检查时间范围', icon: 'none' })
    return
  }
  appliedHistoryFilters.value = { ...historyFilters.value }
  historyPage.value = 1
  void loadHistory()
}

function resetHistoryFilters() {
  if (historyLoading.value)
    return
  historyFilters.value = emptyHistoryFilters()
  appliedHistoryFilters.value = emptyHistoryFilters()
  historyPage.value = 1
  void loadHistory()
}

async function loadHistory() {
  const value = definition.value
  if (!value || !canView.value || historyLoading.value)
    return
  historyLoading.value = true
  try {
    const timeQuery = buildWorkflowHistoryTimeQuery(appliedHistoryFilters.value) || {}
    const response = await listWorkflowInstances({
      definitionId: value.id,
      scope: 'started',
      status: appliedHistoryFilters.value.status || undefined,
      ...timeQuery,
      page: historyPage.value,
      pageSize: historyPageSize,
    })
    historyInstances.value = Array.isArray(response?.data?.list) ? response.data.list : []
    historyTotal.value = Number(response?.data?.total || 0)
    historyLoaded.value = true
  }
  catch {
    uni.showToast({ title: '历史记录加载失败', icon: 'none' })
  }
  finally {
    historyLoading.value = false
  }
}

function handleHistoryPageChange(payload: WorkflowPaginationChangePayload) {
  historyPage.value = Number(payload.current || 1)
  void loadHistory()
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

function openHistoryInstance(instanceId: string) {
  if (resolveMobilePage()) {
    const key = workflowInstanceContentKey(instanceId)
    if (!key)
      return
    appContent.openDynamicTab({
      key,
      label: definition.value?.name || '流程详情',
      icon: 'eye',
      path: `/pages/index/index?view=${encodeURIComponent(key)}`,
    })
    return
  }

  selectedInstanceId.value = instanceId
  detailVisible.value = true
}

function formatTime(timestamp?: number) {
  if (!timestamp)
    return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function workflowRequestErrorMessage(error: unknown, fallback: string) {
  if (!error || typeof error !== 'object')
    return fallback
  const response = error as Record<string, unknown>
  const payload = response.data && typeof response.data === 'object' && !Array.isArray(response.data)
    ? response.data as Record<string, unknown>
    : response
  const message = payload.msg ?? payload.message
  if (typeof message !== 'string' || !message.trim())
    return fallback
  return message.trim()
}

async function saveDraft() {
  const value = definition.value
  if (!value || loading.value || submitting.value || draftSaving.value)
    return false

  draftSaving.value = true
  try {
    const response = await saveWorkflowStartDraft(value.id, {
      definitionVersion: value.version,
      formData: currentWritableData(),
    })
    if (response?.data?.formData) {
      formData.value = initialWorkflowFormData(value.form || [], response.data.formData)
    }
    savedSnapshot.value = currentSnapshot()
    savedDraft.value = response?.data || {
      definitionId: value.id,
      definitionVersion: value.version,
      formData: currentWritableData(),
      updatedAt: Date.now(),
    }
    rememberDefinition(value.id)
    uni.showToast({ title: '草稿已保存', icon: 'success' })
    return true
  }
  catch (error) {
    uni.showToast({ title: workflowRequestErrorMessage(error, '草稿保存失败'), icon: 'none' })
    return false
  }
  finally {
    draftSaving.value = false
  }
}

async function submitStart() {
  const value = definition.value
  if (!value || busy.value)
    return

  if (startLimitExceeded.value) {
    uni.showToast({ title: startLimitMessage.value || '发起次数已用完', icon: 'none' })
    return
  }

  const validation = formRef.value?.validate()
  if (validation && !validation.valid) {
    uni.showToast({ title: '请检查表单必填项', icon: 'none' })
    return
  }

  const data = currentWritableData()
  submitting.value = true
  try {
    const response = await startWorkflowInstance({
      definitionId: value.id,
      definitionVersion: value.version,
      businessType: value.key,
      businessKey: businessKey(value),
      formData: data,
      variables: data,
    })
    const instanceId = String(response?.data?.instanceId || '')
    if (!instanceId)
      throw new Error('流程实例创建失败')

    savedSnapshot.value = currentSnapshot()
    rememberDefinition(value.id)
    appContent.switchContent('workflow')
    appContent.removeDynamicTab(props.contentKey)
    appContent.requestRefresh()
    uni.showToast({ title: '提交成功', icon: 'success' })
  }
  catch (error) {
    uni.showToast({ title: workflowRequestErrorMessage(error, '流程发起失败'), icon: 'none' })
  }
  finally {
    submitting.value = false
  }
}

function cancelStart() {
  if (busy.value)
    return
  appContent.requestCloseTab(props.contentKey)
}
</script>

<template>
  <view class="workflow-start-page">
    <view v-if="loading" class="workflow-start-page__state">
      <u-loading mode="circle" color="#0f766e" size="42" />
      <text>正在加载流程表单...</text>
    </view>

    <view v-else-if="loadError" class="workflow-start-page__state">
      <u-icon name="info-circle" size="60" color="#c8c9cc" />
      <text class="workflow-start-page__state-title">
        {{ loadError }}
      </text>
      <u-button size="small" plain @click="loadStartPage">
        重新加载
      </u-button>
    </view>

    <template v-else-if="definition">
      <view class="workflow-start-page__nav">
        <view
          v-for="item in navigationItems"
          :key="item.key"
          class="workflow-start-page__nav-item"
          :class="{ active: activeSection === item.key }"
          @click="switchSection(item.key)"
        >
          <u-icon :name="item.icon" size="26" :color="activeSection === item.key ? '#1f2329' : '#86909c'" />
          <text>{{ item.label }}</text>
          <text
            v-if="item.key === 'drafts' && draftCount > 0"
            class="workflow-start-page__nav-badge"
          >
            {{ draftCount }}
          </text>
        </view>
      </view>

      <template v-if="activeSection === 'start'">
        <view class="workflow-start-page__body">
          <!-- <view class="workflow-start-page__process">
            <view class="workflow-start-page__heading">
              <view class="workflow-start-page__icon">
                <u-icon name="file-text" size="30" color="#ffffff" />
              </view>
              <view class="workflow-start-page__title-group">
                <text class="workflow-start-page__title">
                  {{ definition.name }}
                </text>
                <text class="workflow-start-page__meta">
                  {{ definition.category || '流程审批' }} · 版本 {{ definition.version }}
                </text>
              </view>
            </view>
            <view class="workflow-start-page__status" :class="{ dirty: hasUnsavedChanges }">
              <view class="workflow-start-page__status-dot" />
              <text>{{ saveStatus }}</text>
            </view>
          </view>
          <view v-if="definition.description" class="workflow-start-page__description">
            {{ definition.description }}
          </view> -->
          <view
            v-if="startLimitMessage"
            class="workflow-start-page__limit-status"
            :class="{ exhausted: startLimitExceeded }"
          >
            <u-icon :name="startLimitExceeded ? 'info-circle' : 'checkmark-circle'" size="22" :color="startLimitExceeded ? '#b45309' : '#0f766e'" />
            <text>{{ startLimitMessage }}</text>
          </view>
          <view class="workflow-start-page__form-card">
            <WorkflowRuntimeForm
              ref="formRef"
              v-model="formData"
              :fields="definition.form || []"
              :field-access="fieldAccess"
              :field-actions="fieldActions"
            />
          </view>
        </view>

        <view class="workflow-start-page__actions">
          <u-button custom-class="workflow-start-page__action" plain :disabled="busy" @click="cancelStart">
            取消
          </u-button>
          <u-button
            custom-class="workflow-start-page__action"
            type="primary"
            :loading="submitting"
            :disabled="loading || draftSaving || startLimitExceeded"
            @click="submitStart"
          >
            提交申请
          </u-button>
          <u-button
            custom-class="workflow-start-page__action"
            type="primary"
            plain
            :loading="draftSaving"
            :disabled="loading || submitting"
            @click="saveDraft"
          >
            保存草稿
          </u-button>
        </view>
      </template>

      <view v-else-if="activeSection === 'history'" class="workflow-start-page__section">
        <!-- <view class="workflow-start-page__section-head">
          <view>
            <text class="workflow-start-page__section-title">
              历史记录
            </text>
            <text class="workflow-start-page__section-total">
              共 {{ historyTotal }} 条
            </text>
          </view>
          <u-button custom-class="workflow-start-page__refresh" size="small" plain :disabled="historyLoading || !canView" @click="loadHistory">
            <u-icon name="reload" size="23" color="#4e5969" />
          </u-button>
        </view> -->
        <WorkflowFilterPanel v-if="canView" :active-count="historyFilterCount">
          <view class="workflow-start-page__history-filters">
            <view class="workflow-start-page__filter-field">
              <text class="workflow-start-page__filter-label">
                审批状态
              </text>
              <!-- H5 PC 使用原生选择控件，避免 u-select 在桌面端弹出底部选择层。 -->
              <select
                v-model="historyFilters.status"
                class="workflow-start-page__filter-select"
                :disabled="historyLoading"
                aria-label="审批状态"
              >
                <option v-for="option in historyStatusOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </view>
            <view class="workflow-start-page__filter-field">
              <text class="workflow-start-page__filter-label">
                发起时间
              </text>
              <view class="workflow-start-page__filter-date-range">
                <WorkflowHistoryDatePicker
                  v-model="historyFilters.startDateFrom"
                  :disabled="historyLoading"
                  placeholder="开始日期"
                />
                <text class="workflow-start-page__filter-separator">
                  至
                </text>
                <WorkflowHistoryDatePicker
                  v-model="historyFilters.startDateTo"
                  :disabled="historyLoading"
                  placeholder="结束日期"
                />
              </view>
            </view>
            <view class="workflow-start-page__filter-field">
              <text class="workflow-start-page__filter-label">
                完成时间
              </text>
              <view class="workflow-start-page__filter-date-range">
                <WorkflowHistoryDatePicker
                  v-model="historyFilters.endDateFrom"
                  :disabled="historyLoading"
                  placeholder="开始日期"
                />
                <text class="workflow-start-page__filter-separator">
                  至
                </text>
                <WorkflowHistoryDatePicker
                  v-model="historyFilters.endDateTo"
                  :disabled="historyLoading"
                  placeholder="结束日期"
                />
              </view>
            </view>
            <view class="workflow-start-page__filter-actions">
              <u-button custom-class="workflow-start-page__filter-action" size="small" plain :disabled="historyLoading" @click="resetHistoryFilters">
                重置
              </u-button>
              <u-button custom-class="workflow-start-page__filter-action" size="small" type="primary" :loading="historyLoading" @click="queryHistory">
                查询
              </u-button>
            </view>
          </view>
        </WorkflowFilterPanel>
        <view v-if="!canView" class="workflow-start-page__empty">
          <u-icon name="lock" size="58" color="#c8c9cc" />
          <text class="workflow-start-page__empty-title">
            暂无流程查看权限
          </text>
        </view>
        <view v-else-if="historyLoading" class="workflow-start-page__empty">
          <u-loading mode="circle" color="#0f766e" size="38" />
          <text>正在加载历史记录...</text>
        </view>
        <view v-else-if="historyInstances.length === 0" class="workflow-start-page__empty">
          <u-icon name="order" size="58" color="#c8c9cc" />
          <text class="workflow-start-page__empty-title">
            暂无历史记录
          </text>
        </view>
        <scroll-view v-else scroll-x class="workflow-start-page__records-scroll">
          <view class="workflow-start-page__records-table">
            <view class="workflow-start-page__record workflow-start-page__record--header">
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--name">
                流程名称
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--key">
                申请编号
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--version">
                流程版本
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--start">
                发起时间
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--end">
                完成时间
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--status">
                审批状态
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--action">
                操作
              </text>
            </view>
            <view
              v-for="instance in historyInstances"
              :key="instance.id"
              class="workflow-start-page__record"
              @click="openHistoryInstance(instance.id)"
            >
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--name">
                {{ definition.name }}
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--key">
                {{ instance.businessKey }}
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--version">
                v{{ instance.definitionVersion }}
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--start">
                {{ formatTime(instance.startTime) }}
              </text>
              <text class="workflow-start-page__record-cell workflow-start-page__record-cell--end">
                {{ formatTime(instance.endTime) }}
              </text>
              <view class="workflow-start-page__record-cell workflow-start-page__record-cell--status">
                <u-tag :text="workflowInstanceStatusMeta(instance.status).label" :type="workflowInstanceStatusMeta(instance.status).type" size="mini" />
              </view>
              <view class="workflow-start-page__record-cell workflow-start-page__record-cell--action">
                <text>查看</text>
                <u-icon name="arrow-right" size="22" color="#0f766e" />
              </view>
            </view>
          </view>
        </scroll-view>
        <view v-if="historyTotal > historyPageSize" class="workflow-start-page__pagination">
          <u-pagination
            v-model="historyPage"
            :total="historyTotal"
            :page-size="historyPageSize"
            prev-text="上一页"
            next-text="下一页"
            @change="handleHistoryPageChange"
          />
        </view>
      </view>

      <view v-else-if="activeSection === 'drafts'" class="workflow-start-page__section">
        <!-- <view class="workflow-start-page__section-head">
          <view>
            <text class="workflow-start-page__section-title">
              草稿箱
            </text>
            <text class="workflow-start-page__section-total">
              共 {{ draftCount }} 份
            </text>
          </view>
        </view> -->
        <view v-if="savedDraft" class="workflow-start-page__draft-grid">
          <view class="workflow-start-page__draft-card">
            <view class="workflow-start-page__draft-cover-head">
              <view class="workflow-start-page__draft-icon">
                <u-icon name="file-text" size="22" color="#0f766e" />
              </view>
              <view class="workflow-start-page__draft-cover-actions">
                <text class="workflow-start-page__draft-label">
                  {{ draftIsLegacy ? '旧版本草稿' : '流程草稿' }}
                </text>
                <view
                  class="workflow-start-page__draft-delete app-icon-button app-icon-button--small"
                  role="button"
                  aria-label="删除草稿"
                  :class="{ disabled: draftDeleting }"
                  @click.stop="confirmDeleteSavedDraft"
                >
                  <u-loading v-if="draftDeleting" mode="circle" size="17" color="#ef4444" />
                  <u-icon v-else name="trash" size="18" color="#ef4444" />
                </view>
              </view>
            </view>
            <view class="workflow-start-page__draft-main">
              <text class="workflow-start-page__draft-title">
                {{ definition.name }}
              </text>
              <text class="workflow-start-page__draft-version">
                版本 {{ savedDraft.definitionVersion }}
              </text>
              <text v-if="draftIsLegacy" class="workflow-start-page__draft-current-version">
                当前流程版本 {{ definition.version }}
              </text>
            </view>
            <view class="workflow-start-page__draft-footer">
              <text class="workflow-start-page__record-time">
                最后更新：{{ formatTime(draftUpdatedAt) }}
              </text>
              <u-button custom-class="workflow-start-page__draft-action" size="small" type="primary" @click="continueSavedDraft">
                继续填写
              </u-button>
            </view>
          </view>
        </view>
        <view v-else class="workflow-start-page__empty">
          <u-icon name="file-text" size="58" color="#c8c9cc" />
          <text class="workflow-start-page__empty-title">
            暂无已保存草稿
          </text>
        </view>
      </view>

      <view v-else class="workflow-start-page__section workflow-start-page__section--graph">
        <view class="workflow-start-page__section-head">
          <view>
            <text class="workflow-start-page__section-title">
              流程图
            </text>
            <text class="workflow-start-page__section-total">
              版本 {{ definition.version }}
            </text>
          </view>
        </view>
        <WorkflowReadOnlyGraph
          :nodes="definition.nodes || []"
          :edges="definition.edges || []"
        />
      </view>
    </template>

    <WorkflowDetailPanel
      v-if="definition"
      v-model="detailVisible"
      :instance-id="selectedInstanceId"
      :definitions="[definition]"
      presentation="history-drawer"
      comment-action
      @changed="loadHistory"
    />
  </view>
</template>

<style lang="scss" scoped>
.workflow-start-page {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f5f7fa;
  color: #1f2329;
  box-sizing: border-box;
}

.workflow-start-page__state {
  min-height: 0;
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: #86909c;
  font-size: 13px;
}

.workflow-start-page__state-title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.workflow-start-page__nav {
  flex: 0 0 auto;
  min-height: 54px;
  padding: 0 24px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: stretch;
  gap: 30px;
  overflow-x: auto;
  background: #ffffff;
}

.workflow-start-page__nav-item {
  flex: 0 0 auto;
  border-bottom: 2px solid transparent;
  display: flex;
  align-items: center;
  gap: 7px;
  color: #86909c;
  font-size: 14px;
  cursor: pointer;
}

.workflow-start-page__nav-item.active {
  border-bottom-color: #1f2329;
  color: #1f2329;
  font-weight: 700;
}

.workflow-start-page__nav-badge {
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f59e0b;
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
  line-height: 18px;
  box-sizing: border-box;
}

.workflow-start-page__process {
  margin-bottom: 18px;
  padding-bottom: 18px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.workflow-start-page__heading {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.workflow-start-page__icon {
  width: 42px;
  height: 42px;
  flex: 0 0 42px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0f766e;
}

.workflow-start-page__title-group {
  min-width: 0;
  display: grid;
  gap: 4px;
}

.workflow-start-page__title,
.workflow-start-page__meta {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-start-page__title {
  font-size: 18px;
  font-weight: 800;
}

.workflow-start-page__meta {
  color: #86909c;
  font-size: 12px;
}

.workflow-start-page__status {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 7px;
  color: #6b7785;
  font-size: 12px;
}

.workflow-start-page__status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #10b981;
}

.workflow-start-page__status.dirty {
  color: #b7791f;
}

.workflow-start-page__status.dirty .workflow-start-page__status-dot {
  background: #f59e0b;
}

.workflow-start-page__body {
  min-height: 0;
  flex: 1 1 auto;
  width: min(var(--app-pc-content-max-width, 1080px), 100%);
  margin: 0 auto;
  padding: 24px;
  overflow-y: auto;
  box-sizing: border-box;
}

.workflow-start-page__form-card {
  padding: 20px;
  border: 1px solid #dfe5ee;
  border-radius: 6px;
  background: #ffffff;
  box-shadow: 0 2px 8px rgba(31, 35, 41, 0.05);
  box-sizing: border-box;
}

.workflow-start-page__limit-status {
  min-height: 40px;
  margin-bottom: 12px;
  padding: 9px 12px;
  border: 1px solid #bfe5dc;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 8px;
  background: #eefaf7;
  color: #0f766e;
  font-size: 12px;
  box-sizing: border-box;
}

.workflow-start-page__limit-status.exhausted {
  border-color: #f4d7a1;
  background: #fff8e8;
  color: #b45309;
}

.workflow-start-page__description {
  margin-bottom: 18px;
  padding: 12px 14px;
  border-left: 3px solid #0f766e;
  background: #eaf8f5;
  color: #3f5f5b;
  font-size: 12px;
  line-height: 1.6;
}

.workflow-start-page__section {
  min-height: 0;
  flex: 1 1 auto;
  width: min(var(--app-pc-content-max-width, 1080px), 100%);
  margin: 0 auto;
  padding: 24px;
  overflow-y: auto;
  box-sizing: border-box;
}

.workflow-start-page__section--graph {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.workflow-start-page__section--graph .workflow-start-page__section-head {
  flex: 0 0 auto;
}

.workflow-start-page__section--graph :deep(.workflow-graph__viewport) {
  min-height: 0;
  flex: 1 1 auto;
}

.workflow-start-page__section-head {
  min-height: 44px;
  margin-bottom: 14px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.workflow-start-page__section-head > view:first-child {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.workflow-start-page__section-title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.workflow-start-page__section-total {
  color: #86909c;
  font-size: 12px;
}

.workflow-start-page__history-filters {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr) minmax(0, 1fr) auto;
  align-items: end;
  gap: 12px;
}

.workflow-start-page__filter-field {
  min-width: 0;
  display: grid;
  gap: 7px;
}

.workflow-start-page__filter-label {
  color: #4e5969;
  font-size: 12px;
  font-weight: 600;
}

.workflow-start-page__filter-select {
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
  cursor: pointer;
}

.workflow-start-page__filter-select:focus {
  border-color: #0f766e;
  outline: none;
  box-shadow: 0 0 0 2px rgba(15, 118, 110, 0.12);
}

.workflow-start-page__filter-select:disabled {
  background: #f2f3f5;
  color: #a9b0bb;
  cursor: not-allowed;
}

.workflow-start-page__filter-date-range {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 7px;
}

.workflow-start-page__filter-separator {
  color: #86909c;
  font-size: 12px;
}

.workflow-start-page__filter-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.workflow-start-page__filter-action,
:deep(.workflow-start-page__filter-action) {
  width: auto;
  min-width: 64px;
  height: 36px;
  margin: 0;
}

.workflow-start-page__refresh,
:deep(.workflow-start-page__refresh) {
  width: 34px;
  height: 34px;
  min-height: 34px;
  margin: -6px 0 0;
  padding: 0;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  background: #ffffff;
}

.workflow-start-page__refresh::after,
:deep(.workflow-start-page__refresh)::after {
  display: none;
}

.workflow-start-page__empty {
  min-height: 300px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #86909c;
  font-size: 13px;
  text-align: center;
}

.workflow-start-page__empty-title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
}

.workflow-start-page__records-scroll {
  width: 100%;
  min-width: 0;
}

.workflow-start-page__records-table {
  width: 100%;
  min-width: 960px;
  border: 1px solid #e5eaf3;
  border-radius: 6px;
  overflow: hidden;
  background: #ffffff;
  box-sizing: border-box;
}

.workflow-start-page__record {
  min-width: 0;
  min-height: 58px;
  border-bottom: 1px solid #edf0f5;
  display: grid;
  grid-template-columns: minmax(140px, 1.2fr) minmax(220px, 1.8fr) minmax(96px, 0.75fr) minmax(155px, 1.25fr) minmax(155px, 1.25fr) minmax(96px, 0.75fr) 76px;
  background: #ffffff;
  box-sizing: border-box;
  cursor: pointer;
}

.workflow-start-page__record:last-child {
  border-bottom: 0;
}

.workflow-start-page__record:hover {
  background: #f7fbfa;
}

.workflow-start-page__record--header {
  min-height: 44px;
  background: #f7f8fa;
  color: #4e5969;
  cursor: default;
  font-size: 12px;
  font-weight: 600;
}

.workflow-start-page__record--header:hover {
  background: #f7f8fa;
}

.workflow-start-page__record-cell {
  min-width: 0;
  padding: 12px 14px;
  border-right: 1px solid #edf0f5;
  display: flex;
  align-items: center;
  color: #4e5969;
  font-size: 13px;
  line-height: 20px;
  box-sizing: border-box;
}

.workflow-start-page__record-cell:last-child {
  border-right: 0;
}

.workflow-start-page__record-cell--name,
.workflow-start-page__record-cell--key,
.workflow-start-page__record-cell--version,
.workflow-start-page__record-cell--start,
.workflow-start-page__record-cell--end {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-start-page__record:not(.workflow-start-page__record--header) .workflow-start-page__record-cell--name {
  color: #1f2329;
  font-weight: 600;
}

.workflow-start-page__record:not(.workflow-start-page__record--header) .workflow-start-page__record-cell--key,
.workflow-start-page__record:not(.workflow-start-page__record--header) .workflow-start-page__record-cell--start,
.workflow-start-page__record:not(.workflow-start-page__record--header) .workflow-start-page__record-cell--end {
  color: #86909c;
}

.workflow-start-page__record-cell--status,
.workflow-start-page__record-cell--action {
  justify-content: center;
}

.workflow-start-page__record-cell--action {
  gap: 3px;
  color: #0f766e;
}

.workflow-start-page__draft-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 160px));
  gap: 16px;
  align-items: start;
}

.workflow-start-page__draft-card {
  width: 160px;
  aspect-ratio: 210 / 297;
  padding: 12px;
  border: 1px solid #d8e0e8;
  border-radius: 2px;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  box-shadow: 0 6px 18px rgba(31, 35, 41, 0.12);
  box-sizing: border-box;
}

.workflow-start-page__draft-cover-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.workflow-start-page__draft-cover-actions {
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
  gap: 5px;
}

.workflow-start-page__draft-label {
  min-width: 0;
  padding-top: 2px;
  overflow: hidden;
  color: #0f766e;
  font-size: 10px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-start-page__draft-delete {
  width: 26px;
  height: 26px;
  flex: 0 0 26px;
  border: 1px solid #fecaca;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff5f5;
  box-sizing: border-box;
  cursor: pointer;
}

.workflow-start-page__draft-delete.disabled {
  cursor: wait;
  opacity: 0.65;
}

.workflow-start-page__draft-main {
  min-width: 0;
  flex: 1 1 auto;
  padding: 8px 0 10px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
}

.workflow-start-page__draft-title {
  overflow: hidden;
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.45;
  display: -webkit-box;
  text-overflow: ellipsis;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.workflow-start-page__draft-version {
  color: #4e5969;
  font-size: 10px;
}

.workflow-start-page__draft-current-version {
  color: #b7791f;
  font-size: 10px;
}

.workflow-start-page__draft-footer {
  padding-top: 8px;
  border-top: 1px solid #eef1f5;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 7px;
}

.workflow-start-page__draft-footer .workflow-start-page__record-time {
  font-size: 9px;
  line-height: 1.45;
  white-space: normal;
}

.workflow-start-page__draft-action,
:deep(.workflow-start-page__draft-action) {
  width: 100%;
  min-width: 0;
  margin: 0;
}

.workflow-start-page__record-title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.workflow-start-page__record-key,
.workflow-start-page__record-time {
  color: #86909c;
  font-size: 12px;
}

.workflow-start-page__draft-icon {
  width: 32px;
  height: 32px;
  flex: 0 0 32px;
  border-radius: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #eaf8f5;
}

.workflow-start-page__pagination {
  margin-top: 18px;
  display: flex;
  justify-content: flex-end;
}

.workflow-start-page__actions {
  position: relative;
  z-index: 20;
  flex: 0 0 auto;
  min-height: 68px;
  padding: 12px 24px;
  border-top: 1px solid #dfe5ee;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  background: #ffffff;
  box-sizing: border-box;
  box-shadow: 0 -4px 16px rgba(31, 35, 41, 0.05);
}

.workflow-start-page__action,
:deep(.workflow-start-page__action) {
  width: auto;
  min-width: 112px;
  margin: 0;
}

@media screen and (max-width: 1000px) {
  .workflow-start-page__history-filters {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workflow-start-page__filter-actions {
    grid-column: 1 / -1;
    justify-content: flex-end;
  }
}

@media screen and (max-width: 768px) {
  .workflow-start-page {
    height: auto;
    min-height: 100%;
    display: block;
    overflow: visible;
  }

  .workflow-start-page__nav {
    min-height: 50px;
    padding: 0 14px;
    gap: 22px;
  }

  .workflow-start-page__nav-item {
    font-size: 13px;
  }

  .workflow-start-page__icon {
    width: 38px;
    height: 38px;
    flex-basis: 38px;
  }

  .workflow-start-page__body {
    padding: 16px 12px;
    overflow: visible;
  }

  .workflow-start-page__form-card {
    padding: 12px;
  }

  .workflow-start-page__process {
    align-items: flex-start;
  }

  .workflow-start-page__section {
    padding: 16px 12px;
    overflow: visible;
  }

  .workflow-start-page__history-filters {
    grid-template-columns: 1fr;
    padding-top: 4px;
  }

  .workflow-start-page__filter-actions {
    grid-column: auto;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workflow-start-page__filter-action,
  :deep(.workflow-start-page__filter-action) {
    width: 100%;
  }

  .workflow-start-page__records-table {
    min-width: 0;
    border: 0;
    border-radius: 0;
    overflow: visible;
    background: transparent;
  }

  .workflow-start-page__record--header {
    display: none;
  }

  .workflow-start-page__record:not(.workflow-start-page__record--header) {
    min-height: 0;
    margin-bottom: 10px;
    padding: 12px;
    border: 1px solid #e5eaf3;
    border-radius: 6px;
    grid-template-columns: auto minmax(0, 1fr) auto;
    grid-template-rows: auto auto;
    align-items: center;
    gap: 10px 12px;
  }

  .workflow-start-page__record:not(.workflow-start-page__record--header):last-child {
    margin-bottom: 0;
    border-bottom: 1px solid #e5eaf3;
  }

  .workflow-start-page__record-cell {
    padding: 0;
    border-right: 0;
  }

  .workflow-start-page__record-cell--key,
  .workflow-start-page__record-cell--version,
  .workflow-start-page__record-cell--end {
    display: none;
  }

  .workflow-start-page__record-cell--name {
    grid-column: 1 / -1;
    grid-row: 1;
  }

  .workflow-start-page__record-cell--status {
    grid-column: 1;
    grid-row: 2;
    justify-content: flex-start;
  }

  .workflow-start-page__record-cell--start {
    grid-column: 2;
    grid-row: 2;
  }

  .workflow-start-page__record-cell--action {
    grid-column: 3;
    grid-row: 2;
    justify-content: flex-end;
  }

  .workflow-start-page__section--graph {
    display: block;
  }

  .workflow-start-page__draft-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .workflow-start-page__draft-card {
    width: 160px;
    max-width: 100%;
    margin: 0 auto;
  }

  .workflow-start-page__pagination {
    justify-content: center;
  }

  .workflow-start-page__actions {
    position: sticky;
    bottom: 0;
    min-height: 64px;
    padding: 10px 12px;
  }

  .workflow-start-page__action,
  :deep(.workflow-start-page__action) {
    flex: 1 1 0;
    min-width: 0;
  }
}
</style>
