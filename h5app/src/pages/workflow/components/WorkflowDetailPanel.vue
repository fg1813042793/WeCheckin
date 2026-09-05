<script setup lang="ts">
import type {
  WorkflowAttachment,
  WorkflowFieldAccessMap,
  WorkflowFieldActionsMap,
  WorkflowFormData,
  WorkflowInstanceDetail,
  WorkflowNotificationChannel,
  WorkflowPublishedDefinition,
  WorkflowReminderNode,
  WorkflowTaskSummary,
} from '@/types/workflow'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  commentWorkflowInstance,
  completeWorkflowTask,
  deleteWorkflowInstance,
  getWorkflowInstance,
  getWorkflowSummaryInstance,
  remindWorkflowInstance,
  withdrawWorkflowInstance,
} from '@/api/workflow'
import { useAppContentStore, useDingtalkAuthStore } from '@/stores'
import {
  initialWorkflowFormData,
  workflowFieldAccessMap,
  workflowFieldActionsMap,
  writableWorkflowFormData,
} from '../workflow-form'
import { workflowFormDetailContentKey, workflowFormRevisionContentKey, workflowStartContentKey } from '../workflow-route-keys'
import { workflowInstanceStatusMeta, workflowTaskStatusMeta } from '../workflow-status'
import { isWorkflowTaskAssignedToUser } from '../workflow-task'
import WorkflowImagePicker from './WorkflowImagePicker.vue'
import WorkflowNodeProgressList from './WorkflowNodeProgressList.vue'
import WorkflowParticipantSelect from './WorkflowParticipantSelect.vue'
import WorkflowReadOnlyGraph from './WorkflowReadOnlyGraph.vue'
import WorkflowRuntimeForm from './WorkflowRuntimeForm.vue'
import WorkflowTextarea from './WorkflowTextarea.vue'

type WorkflowDetailSection = 'form' | 'history' | 'graph'

const props = withDefaults(defineProps<{
  modelValue: boolean
  instanceId: string
  taskId?: string
  definitions?: WorkflowPublishedDefinition[]
  displayTitle?: string
  presentation?: 'dialog' | 'history-drawer' | 'page' | 'history-page'
  applicationActions?: boolean
  commentAction?: boolean
  formRevisionAction?: boolean
  summaryMode?: boolean
}>(), {
  taskId: '',
  definitions: () => [],
  displayTitle: '',
  presentation: 'dialog',
  applicationActions: false,
  commentAction: false,
  formRevisionAction: false,
  summaryMode: false,
})

const emit = defineEmits<{
  'update:modelValue': [visible: boolean]
  'changed': []
}>()

interface RuntimeFormExposed {
  validate: () => { valid: boolean, errors: Record<string, string> }
}

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const loading = ref(false)
const submitting = ref(false)
const commentSubmitting = ref(false)
const reminderSubmittingNodeId = ref('')
const detail = ref<WorkflowInstanceDetail | null>(null)
const formData = ref<WorkflowFormData>({})
const commentVisible = ref(false)
const commentDraft = ref('')
const commentImages = ref<WorkflowAttachment[]>([])
const commentUploading = ref(false)
const commentNotificationUserIds = ref<string[]>([])
const commentNotificationChannels = ref<WorkflowNotificationChannel[]>(['in_app'])
const rejectVisible = ref(false)
const rejectDraft = ref('')
const rejectImages = ref<WorkflowAttachment[]>([])
const rejectUploading = ref(false)
const returnVisible = ref(false)
const returnDraft = ref('')
const returnImages = ref<WorkflowAttachment[]>([])
const returnUploading = ref(false)
const returnTargetNodeId = ref('')
const actionConfirmVisible = ref(false)
const actionConfirmTitle = ref('')
const actionConfirmContent = ref('')
const applicationAction = ref('')
const activeSection = ref<WorkflowDetailSection>('form')
const historyEventsExpanded = ref(false)
const formRef = ref<RuntimeFormExposed | null>(null)
let actionConfirmResolver: ((confirmed: boolean) => void) | null = null

const HISTORY_DIALOG_BREAKPOINT = 1024
const MOBILE_INTERACTION_BREAKPOINT = 768

function resolveCompactHistoryDialog() {
  // #ifdef H5
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function')
    return window.matchMedia('(max-width: 1024px)').matches
  // #endif

  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    if (Number.isFinite(width) && width > 0)
      return width <= HISTORY_DIALOG_BREAKPOINT
  }
  catch {
    return false
  }

  return false
}

const compactHistoryDialog = ref(resolveCompactHistoryDialog())
const mobileInteractionDialog = ref(resolveMobileInteractionDialog())

function resolveMobileInteractionDialog() {
  // #ifdef H5
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function')
    return window.matchMedia(`(max-width: ${MOBILE_INTERACTION_BREAKPOINT}px)`).matches
  // #endif

  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    if (Number.isFinite(width) && width > 0)
      return width <= MOBILE_INTERACTION_BREAKPOINT
  }
  catch {
    return false
  }

  return false
}

function syncCompactHistoryDialog() {
  compactHistoryDialog.value = resolveCompactHistoryDialog()
  mobileInteractionDialog.value = resolveMobileInteractionDialog()
}

const popupVisible = computed({
  get: () => props.modelValue,
  set: (visible) => {
    if (!visible) {
      commentVisible.value = false
      rejectVisible.value = false
      returnVisible.value = false
      resolveActionConfirmation(false)
    }
    emit('update:modelValue', visible)
  },
})

const historyDrawer = computed(() => props.presentation === 'history-drawer')
const historyDialog = computed(() => historyDrawer.value && compactHistoryDialog.value)
const historyPage = computed(() => props.presentation === 'history-page')
const historyPresentation = computed(() => historyDrawer.value || historyPage.value)
const pagePresentation = computed(() => props.presentation === 'page')
const inlinePresentation = computed(() => pagePresentation.value || historyPage.value)
const popupMode = computed(() => {
  if (inlinePresentation.value)
    return 'inline'
  return historyDrawer.value && !historyDialog.value ? 'right' : 'center'
})
const popupWidth = computed(() => {
  if (inlinePresentation.value)
    return '100%'
  if (historyDialog.value)
    return '720px'
  return historyDrawer.value ? '520px' : '94%'
})
const popupHeight = computed(() => {
  if (inlinePresentation.value)
    return '100%'
  return historyDialog.value ? '760px' : '92%'
})
const popupCustomClass = computed(() => {
  if (inlinePresentation.value)
    return 'workflow-detail-popup workflow-detail-popup--page app-pc-control-scope'
  if (historyDialog.value)
    return 'workflow-detail-popup workflow-detail-popup--history-dialog app-pc-control-scope'
  if (historyDrawer.value)
    return 'workflow-detail-popup workflow-detail-popup--history-drawer app-pc-control-scope'
  return 'workflow-detail-popup app-pc-control-scope'
})
const popupBorderRadius = computed(() => {
  return inlinePresentation.value || (historyDrawer.value && !historyDialog.value) ? 0 : 8
})
const commentPopupMode = computed(() => mobileInteractionDialog.value ? 'bottom' : 'center')
const commentPopupWidth = computed(() => mobileInteractionDialog.value ? '100%' : '520px')
const rejectPopupMode = computed(() => mobileInteractionDialog.value ? 'bottom' : 'center')
const rejectPopupWidth = computed(() => mobileInteractionDialog.value ? '100%' : '480px')
const returnPopupMode = computed(() => mobileInteractionDialog.value ? 'bottom' : 'center')
const returnPopupWidth = computed(() => mobileInteractionDialog.value ? '100%' : '480px')
const pageNavigationItems: Array<{ key: WorkflowDetailSection, label: string, icon: string }> = [
  { key: 'form', label: '审批处理', icon: 'checkmark-circle' },
  { key: 'history', label: '流程记录', icon: 'clock' },
  { key: 'graph', label: '流程图', icon: 'share' },
]

const definition = computed(() => {
  const definitionId = detail.value?.instance.definitionId
  return props.definitions.find(item => item.id === definitionId) || null
})

const title = computed(() => {
  return props.displayTitle || definition.value?.name || detail.value?.instance.definitionKey || '流程详情'
})

const currentUserId = computed(() => String(auth.user?.workflowActorId || auth.user?.id || ''))

const activeTask = computed<WorkflowTaskSummary | null>(() => {
  const tasks = detail.value?.tasks || []
  if (props.taskId) {
    return tasks.find(task => task.id === props.taskId) || null
  }
  return tasks.find(task => task.status === 'pending' && task.assigneeId === currentUserId.value) || null
})

const activeNodeType = computed(() => {
  const task = activeTask.value
  return task ? detail.value?.nodeTypes?.[task.nodeId] || 'approval' : ''
})

const canHandle = computed(() => {
  return Boolean(
    activeTask.value?.status === 'pending'
    && isWorkflowTaskAssignedToUser(activeTask.value, currentUserId.value)
    && auth.hasApiPermission('dingtalk_h5:api:workflow:handle'),
  )
})

const returnTargets = computed(() => {
  const current = activeTask.value
  const currentDetail = detail.value
  if (!current || !currentDetail)
    return []

  const currentIndex = currentDetail.tasks.findIndex(task => task.id === current.id)
  if (currentIndex <= 0)
    return []
  const nodesById = new Map((currentDetail.nodes || []).map(node => [node.id, node]))
  const seen = new Set<string>()
  const targets: Array<{ label: string, value: string }> = []
  for (let index = currentIndex - 1; index >= 0; index--) {
    const task = currentDetail.tasks[index]
    if (!task || task.nodeId === current.nodeId || !['approved', 'completed'].includes(task.status) || seen.has(task.nodeId))
      continue
    const node = nodesById.get(task.nodeId)
    if (!node || !['approval', 'handle'].includes(node.type))
      continue
    seen.add(task.nodeId)
    targets.push({
      label: `${targets.length === 0 ? '上一节点' : '指定节点'}：${node.name || task.nodeName || node.id}`,
      value: node.id,
    })
  }
  return targets
})

const canReturn = computed(() => {
  const traversedParallel = (detail.value?.tokens || []).some(token => token.branchGroup || token.branchTotal > 1)
  return canHandle.value && activeNodeType.value === 'approval' && returnTargets.value.length > 0 && !traversedParallel
})

const canWithdraw = computed(() => {
  const instance = detail.value?.instance
  return Boolean(
    instance
    && instance.status === 'running'
    && instance.starterId === currentUserId.value
    && auth.hasApiPermission('dingtalk_h5:api:workflow:withdraw'),
  )
})

const applicationOwner = computed(() => {
  return Boolean(detail.value?.instance.starterId === currentUserId.value)
})

const showApplicationActions = computed(() => {
  return Boolean(historyDrawer.value && props.applicationActions && detail.value)
})

const showFormDetailAction = computed(() => {
  return Boolean(historyDrawer.value && !historyDialog.value && detail.value)
})

const showCommentAction = computed(() => {
  return Boolean(props.commentAction && detail.value)
})

const showTaskActionBar = computed(() => {
  if (pagePresentation.value && mobileInteractionDialog.value && activeSection.value === 'graph')
    return false
  return pagePresentation.value || canHandle.value || canWithdraw.value || showCommentAction.value
})

const showFormRevisionAction = computed(() => {
  const current = detail.value
  return Boolean(
    historyDrawer.value
    && props.formRevisionAction
    && current?.instance.status === 'running'
    && current.formRevision?.allowed
    && auth.hasButtonPermission('dingtalk_h5:button:workflow:form-revise')
    && auth.hasApiPermission('dingtalk_h5:api:workflow:form-revise'),
  )
})

const canModifyApplication = computed(() => {
  const instance = detail.value?.instance
  return Boolean(
    applicationOwner.value
    && definition.value
    && auth.hasApiPermission('dingtalk_h5:api:workflow:start')
    && (instance?.status !== 'running' || canWithdraw.value),
  )
})

const resubmitApplication = computed(() => {
  const status = detail.value?.instance.status || ''
  return ['completed', 'cancelled'].includes(status)
})

const modifyApplicationLabel = computed(() => {
  return resubmitApplication.value ? '再次提交' : '修改'
})

const canCommentInstance = computed(() => {
  return Boolean(
    showCommentAction.value
    && auth.hasApiPermission('dingtalk_h5:api:workflow:comment'),
  )
})

interface CommentNotificationUser {
  userId: string
  userName: string
}

interface CommentNotificationGroup {
  nodeId: string
  nodeName: string
  users: CommentNotificationUser[]
}

const commentNotificationGroups = computed<CommentNotificationGroup[]>(() => {
  const current = detail.value
  if (!current)
    return []

  const groups = new Map<string, CommentNotificationGroup>()
  const seenUsers = new Set<string>([currentUserId.value])
  const nodeNames = new Map((current.nodes || []).map(node => [node.id, node.name]))
  const addUser = (nodeId: string, fallbackNodeName: string, userId: string, userName: string) => {
    const normalizedUserId = String(userId || '').trim()
    if (!normalizedUserId || seenUsers.has(normalizedUserId))
      return
    const resolvedUserName = String(userName || current.userNames?.[normalizedUserId] || '').trim()
    if (!resolvedUserName)
      return
    const key = String(nodeId || 'start').trim() || 'start'
    const nodeName = String(nodeNames.get(key) || fallbackNodeName || '流程参与人').trim()
    const group = groups.get(key) || { nodeId: key, nodeName, users: [] }
    group.users.push({ userId: normalizedUserId, userName: resolvedUserName })
    groups.set(key, group)
    seenUsers.add(normalizedUserId)
  }

  addUser(current.startNodeId || 'start', '发起节点', current.instance.starterId, current.instance.starterName)
  for (const task of current.tasks || []) {
    addUser(task.nodeId, task.nodeName, task.assigneeId, task.assigneeName)
    addUser(task.nodeId, task.nodeName, task.handledBy, task.handledByName)
  }
  for (const event of current.history || []) {
    if (event.eventType === 'node_cc')
      addUser(event.nodeId, nodeNames.get(event.nodeId) || '抄送节点', event.actorId, event.actorName || '')
  }

  const nodeOrder = new Map((current.nodes || []).map((node, index) => [node.id, index]))
  return [...groups.values()].sort((left, right) => {
    return (nodeOrder.get(left.nodeId) ?? Number.MAX_SAFE_INTEGER) - (nodeOrder.get(right.nodeId) ?? Number.MAX_SAFE_INTEGER)
  })
})

const hasDeleteApplicationPermission = computed(() => {
  return auth.hasApiPermission('dingtalk_h5:api:workflow:delete')
})

const canDeleteApplication = computed(() => {
  const status = detail.value?.instance.status || ''
  return Boolean(
    applicationOwner.value
    && ['completed', 'rejected', 'withdrawn', 'cancelled'].includes(status),
  )
})

function deleteApplicationUnavailableMessage() {
  const instance = detail.value?.instance
  if (!instance)
    return '流程详情尚未加载完成'
  if (!applicationOwner.value)
    return '只能删除自己发起的申请'
  if (instance.status === 'running')
    return '审批中的申请请先撤销后再删除'
  return '当前状态的申请暂不支持删除'
}

const canRemindInstance = computed(() => {
  const instance = detail.value?.instance
  return Boolean(
    historyPresentation.value
    && instance?.status === 'running'
    && applicationOwner.value
    && auth.hasApiPermission('dingtalk_h5:api:workflow:remind'),
  )
})

const reminderNodes = computed(() => detail.value?.reminderNodes || [])

const applicationActionBusy = computed(() => submitting.value || commentSubmitting.value)

const fieldAccess = computed<WorkflowFieldAccessMap>(() => {
  const current = detail.value
  const task = activeTask.value
  if (!current || !task || !canHandle.value) {
    return workflowFieldAccessMap(current?.form || [], [], 'read')
  }
  return workflowFieldAccessMap(
    current.form || [],
    current.fieldPermissions?.[task.nodeId] || [],
    'read',
  )
})

const fieldActions = computed<WorkflowFieldActionsMap>(() => {
  const current = detail.value
  const task = activeTask.value
  if (!current || !task || !canHandle.value)
    return {}
  return workflowFieldActionsMap(
    current.form || [],
    current.fieldPermissions?.[task.nodeId] || [],
  )
})

const historyFieldAccess = computed<WorkflowFieldAccessMap>(() => {
  return workflowFieldAccessMap(detail.value?.form || [], [], 'read')
})

const sortedTasks = computed(() => {
  return [...(detail.value?.tasks || [])].sort((left, right) => {
    return left.sequence - right.sequence || left.id.localeCompare(right.id)
  })
})

const sortedHistory = computed(() => {
  return [...(detail.value?.history || [])].sort((left, right) => right.eventTime - left.eventTime)
})

onMounted(() => {
  syncCompactHistoryDialog()
  // #ifdef H5
  window.addEventListener('resize', syncCompactHistoryDialog)
  // #endif
})

onBeforeUnmount(() => {
  // #ifdef H5
  window.removeEventListener('resize', syncCompactHistoryDialog)
  // #endif
})

watch(
  () => [props.modelValue, props.instanceId, props.taskId],
  ([visible]) => {
    if (!visible) {
      commentVisible.value = false
      rejectVisible.value = false
      returnVisible.value = false
      resolveActionConfirmation(false)
      return
    }
    if (!props.instanceId)
      return
    if (historyPresentation.value)
      historyEventsExpanded.value = false
    void loadDetail()
  },
  { immediate: true },
)

async function loadDetail() {
  if (!props.instanceId || loading.value)
    return
  activeSection.value = 'form'
  loading.value = true
  try {
    const response = props.summaryMode
      ? await getWorkflowSummaryInstance(props.instanceId)
      : await getWorkflowInstance(props.instanceId)
    if (!response?.data)
      return
    detail.value = response.data
    formData.value = initialWorkflowFormData(response.data.form || [], response.data.formData || {})
  }
  catch {
    uni.showToast({ title: '流程详情加载失败', icon: 'none' })
  }
  finally {
    loading.value = false
  }
}

function actionLabel(action = '') {
  return {
    approve: '通过',
    reject: '驳回',
    return: '退回',
    submit: '提交',
  }[action] || action
}

function taskHandlerName(task: WorkflowTaskSummary) {
  const name = task.handledBy ? task.handledByName : task.assigneeName
  return String(name || '').trim() || '未知用户'
}

function historyLabel(eventType = '') {
  const labels: Record<string, string> = {
    instance_started: '流程已发起',
    task_created: '任务已创建',
    task_approved: '审批通过',
    task_rejected: '审批驳回',
    task_returned: '任务已退回',
    task_submitted: '办理已提交',
    node_entered: '进入节点',
    node_completed: '节点已完成',
    instance_completed: '流程已完成',
    instance_rejected: '流程已驳回',
    instance_withdrawn: '流程已撤回',
    instance_cancelled: '流程已取消',
    instance_commented: '评论',
    instance_form_revised: '表单已修改',
    instance_reminded: '已提醒处理人',
  }
  return labels[eventType] || eventType || '流程记录'
}

function historyActorName(actorName = '', actorId = '') {
  const name = String(actorName || '').trim()
  if (name)
    return name
  return actorId ? '未知用户' : ''
}

function formatTime(timestamp?: number) {
  if (!timestamp)
    return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function reminderAssigneeText(node: WorkflowReminderNode) {
  const names = node.assigneeNames.map(name => name.trim()).filter(Boolean)
  if (names.length > 0)
    return names.join('、')
  return `${node.assigneeCount} 位处理人`
}

function reminderButtonDisabled(node: WorkflowReminderNode) {
  return !canRemindInstance.value
    || reminderSubmittingNodeId.value !== ''
    || node.blockedReason === 'daily_limit'
    || node.nextAllowedAt > Date.now()
}

function reminderStatusText(node: WorkflowReminderNode) {
  if (node.blockedReason === 'daily_limit')
    return '今日提醒次数已用完'
  if (node.nextAllowedAt > Date.now())
    return `${formatTime(node.nextAllowedAt)} 后可再次提醒`
  return `今日剩余 ${node.remainingCount} 次`
}

function workflowRequestErrorMessage(error: unknown, fallback: string) {
  if (!error || typeof error !== 'object')
    return fallback
  const response = error as Record<string, unknown>
  const payload = response.data && typeof response.data === 'object' && !Array.isArray(response.data)
    ? response.data as Record<string, unknown>
    : response
  const message = payload.msg ?? payload.message
  return typeof message === 'string' && message.trim() ? message.trim() : fallback
}

async function remindNode(node: WorkflowReminderNode) {
  if (reminderButtonDisabled(node))
    return
  reminderSubmittingNodeId.value = node.nodeId
  try {
    const response = await remindWorkflowInstance(props.instanceId, { nodeId: node.nodeId })
    if (!response?.data)
      return
    await loadDetail()
    uni.showToast({ title: `已提醒 ${response.data.remindedCount} 位处理人`, icon: 'success' })
  }
  catch (error) {
    uni.showToast({ title: workflowRequestErrorMessage(error, '提醒发送失败'), icon: 'none' })
  }
  finally {
    reminderSubmittingNodeId.value = ''
  }
}

function confirmAction(title: string, content: string) {
  return new Promise<boolean>((resolve) => {
    actionConfirmResolver?.(false)
    actionConfirmResolver = resolve
    actionConfirmTitle.value = title
    actionConfirmContent.value = content
    actionConfirmVisible.value = true
  })
}

function resolveActionConfirmation(confirmed: boolean) {
  actionConfirmVisible.value = false
  const resolve = actionConfirmResolver
  actionConfirmResolver = null
  resolve?.(confirmed)
}

async function submitTask(action: 'approve' | 'submit') {
  const task = activeTask.value
  const current = detail.value
  if (!task || !current || !canHandle.value || submitting.value)
    return

  const validation = formRef.value?.validate()
  if (validation && !validation.valid) {
    uni.showToast({ title: '请检查表单必填项', icon: 'none' })
    return
  }
  const patch = writableWorkflowFormData(current.form || [], formData.value, fieldAccess.value)

  const label = activeNodeType.value === 'handle' ? '提交办理' : '通过'
  if (!await confirmAction(`${label}流程`, `确认执行“${label}”操作吗？`))
    return

  submitting.value = true
  try {
    const response = await completeWorkflowTask(task.id, {
      action,
      formData: patch,
      variables: patch,
    })
    if (!response?.data)
      return
    uni.showToast({ title: `${label}成功`, icon: 'success' })
    emit('changed')
    if (!pagePresentation.value)
      await loadDetail()
  }
  catch {
    uni.showToast({ title: `${label}失败`, icon: 'none' })
  }
  finally {
    submitting.value = false
  }
}

function openReject() {
  if (!canHandle.value || submitting.value)
    return
  rejectDraft.value = ''
  rejectImages.value = []
  rejectVisible.value = true
}

function closeReject() {
  if (submitting.value || rejectUploading.value)
    return
  rejectVisible.value = false
}

async function submitReject() {
  const task = activeTask.value
  const value = rejectDraft.value.trim()
  if (!task || !canHandle.value || submitting.value)
    return
  if (rejectUploading.value) {
    uni.showToast({ title: '请等待图片上传完成', icon: 'none' })
    return
  }
  if (!value) {
    uni.showToast({ title: '请填写驳回原因', icon: 'none' })
    return
  }
  if (Array.from(value).length > 500) {
    uni.showToast({ title: '驳回原因不能超过500个字符', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    const response = await completeWorkflowTask(task.id, {
      action: 'reject',
      comment: value,
      images: rejectImages.value,
    })
    if (!response?.data)
      return
    rejectVisible.value = false
    rejectDraft.value = ''
    rejectImages.value = []
    uni.showToast({ title: '驳回成功', icon: 'success' })
    emit('changed')
    if (!pagePresentation.value)
      await loadDetail()
  }
  catch {
    uni.showToast({ title: '驳回失败', icon: 'none' })
  }
  finally {
    submitting.value = false
  }
}

function openReturn() {
  if (!canReturn.value || submitting.value)
    return
  returnTargetNodeId.value = returnTargets.value[0]?.value || ''
  returnDraft.value = ''
  returnImages.value = []
  returnVisible.value = true
}

function closeReturn() {
  if (submitting.value || returnUploading.value)
    return
  returnVisible.value = false
}

async function submitReturn() {
  const task = activeTask.value
  const targetNodeId = returnTargetNodeId.value.trim()
  const value = returnDraft.value.trim()
  if (!task || !canReturn.value || submitting.value)
    return
  if (!targetNodeId) {
    uni.showToast({ title: '请选择退回节点', icon: 'none' })
    return
  }
  if (returnUploading.value) {
    uni.showToast({ title: '请等待图片上传完成', icon: 'none' })
    return
  }
  if (!value) {
    uni.showToast({ title: '请填写退回原因', icon: 'none' })
    return
  }
  if (Array.from(value).length > 500) {
    uni.showToast({ title: '退回原因不能超过500个字符', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    const response = await completeWorkflowTask(task.id, {
      action: 'return',
      comment: value,
      images: returnImages.value,
      returnTargetNodeId: targetNodeId,
    })
    if (!response?.data)
      return
    returnVisible.value = false
    returnDraft.value = ''
    returnImages.value = []
    returnTargetNodeId.value = ''
    uni.showToast({ title: '退回成功', icon: 'success' })
    emit('changed')
    if (!pagePresentation.value)
      await loadDetail()
  }
  catch {
    uni.showToast({ title: '退回失败', icon: 'none' })
  }
  finally {
    submitting.value = false
  }
}

async function withdraw() {
  if (!canWithdraw.value || submitting.value)
    return
  if (!await confirmAction('撤回申请', '撤回后流程将停止，确认继续吗？'))
    return
  submitting.value = true
  applicationAction.value = 'withdraw'
  try {
    const response = await withdrawWorkflowInstance(props.instanceId, '发起人主动撤回')
    if (!response?.data)
      return
    uni.showToast({ title: '流程已撤回', icon: 'success' })
    emit('changed')
    if (!pagePresentation.value)
      await loadDetail()
  }
  catch {
    uni.showToast({ title: '撤回失败', icon: 'none' })
  }
  finally {
    submitting.value = false
    applicationAction.value = ''
  }
}

function openComment() {
  if (!canCommentInstance.value || applicationActionBusy.value)
    return
  commentDraft.value = ''
  commentImages.value = []
  commentNotificationUserIds.value = []
  commentNotificationChannels.value = ['in_app']
  commentVisible.value = true
}

function closeComment() {
  if (commentSubmitting.value || commentUploading.value)
    return
  commentVisible.value = false
}

async function submitComment() {
  const value = commentDraft.value.trim()
  if (!canCommentInstance.value || commentSubmitting.value)
    return
  if (commentUploading.value) {
    uni.showToast({ title: '请等待图片上传完成', icon: 'none' })
    return
  }
  if (!value && commentImages.value.length === 0) {
    uni.showToast({ title: '请输入评论内容或上传图片', icon: 'none' })
    return
  }
  if (Array.from(value).length > 500) {
    uni.showToast({ title: '评论不能超过500个字符', icon: 'none' })
    return
  }
  if (commentNotificationUserIds.value.length > 0 && commentNotificationChannels.value.length === 0) {
    uni.showToast({ title: '请选择通知方式', icon: 'none' })
    return
  }

  commentSubmitting.value = true
  applicationAction.value = 'comment'
  try {
    await commentWorkflowInstance(props.instanceId, {
      comment: value,
      images: commentImages.value,
      notification: commentNotificationUserIds.value.length > 0
        ? {
            userIds: commentNotificationUserIds.value,
            channels: commentNotificationChannels.value,
          }
        : undefined,
    })
    commentVisible.value = false
    commentDraft.value = ''
    commentImages.value = []
    commentNotificationUserIds.value = []
    commentNotificationChannels.value = ['in_app']
    await loadDetail()
    historyEventsExpanded.value = true
    uni.showToast({ title: '评论已发布', icon: 'success' })
  }
  catch {
    uni.showToast({ title: '评论发布失败', icon: 'none' })
  }
  finally {
    commentSubmitting.value = false
    applicationAction.value = ''
  }
}

async function modifyApplication() {
  const current = detail.value
  const targetDefinition = definition.value
  if (!current || !targetDefinition || !canModifyApplication.value || applicationActionBusy.value)
    return

  const running = current.instance.status === 'running'
  const actionTitle = resubmitApplication.value ? '再次提交' : '修改申请'
  const content = running
    ? '修改前需要先撤销当前申请。撤销后将把现有表单内容复制到发起页，原流程记录保留。发起页已有未保存内容会被替换，确认继续吗？'
    : resubmitApplication.value
      ? '将把原表单内容带入发起页，确认后可创建一条新的流程申请，原流程记录保留。发起页已有未保存内容会被替换，确认继续吗？'
      : '将把现有表单内容复制到发起页作为一份新申请，原流程记录保留。发起页已有未保存内容会被替换，确认继续吗？'
  if (!await confirmAction(actionTitle, content))
    return

  submitting.value = true
  applicationAction.value = 'modify'
  try {
    if (running) {
      const response = await withdrawWorkflowInstance(props.instanceId, '发起人修改申请')
      if (!response?.data)
        throw new Error('流程撤销失败')
    }
    appContent.setWorkflowStartSeed({
      definitionId: targetDefinition.id,
      sourceInstanceId: current.instance.id,
      formData: current.formData || {},
    })
    const key = workflowStartContentKey(targetDefinition.id)
    appContent.openDynamicTab({
      key,
      label: targetDefinition.name,
      icon: 'file-text',
      path: `/pages/index/index?view=${encodeURIComponent(key)}`,
    })
    if (running)
      emit('changed')
    popupVisible.value = false
    uni.showToast({ title: '已复制到发起页', icon: 'success' })
  }
  catch {
    uni.showToast({ title: running ? '撤销并复制申请失败' : '复制申请失败', icon: 'none' })
  }
  finally {
    submitting.value = false
    applicationAction.value = ''
  }
}

function openFormRevision() {
  const current = detail.value
  if (!current || !showFormRevisionAction.value)
    return
  const key = workflowFormRevisionContentKey(current.instance.id)
  if (!key)
    return
  appContent.openDynamicTab({
    key,
    label: `修改 · ${current.instance.definitionName || title.value}`,
    icon: 'edit-pen',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  })
  popupVisible.value = false
}

function openFormDetail() {
  const current = detail.value
  if (!current || !showFormDetailAction.value)
    return
  const key = workflowFormDetailContentKey(current.instance.id)
  if (!key)
    return
  appContent.openDynamicTab({
    key,
    label: `表单 · ${current.instance.definitionName || title.value}`,
    icon: 'file-text',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  })
  popupVisible.value = false
}

async function deleteApplication() {
  if (applicationActionBusy.value)
    return
  if (!hasDeleteApplicationPermission.value) {
    uni.showToast({ title: '无申请删除权限', icon: 'none' })
    return
  }
  if (!canDeleteApplication.value) {
    uni.showToast({ title: deleteApplicationUnavailableMessage(), icon: 'none' })
    return
  }
  if (!await confirmAction('删除申请', '删除后将不再出现在“我的申请”中，后台审计记录仍会保留。确认删除吗？'))
    return

  submitting.value = true
  applicationAction.value = 'delete'
  try {
    await deleteWorkflowInstance(props.instanceId)
    popupVisible.value = false
    emit('changed')
    uni.showToast({ title: '申请已删除', icon: 'success' })
  }
  catch (error) {
    uni.showToast({ title: workflowRequestErrorMessage(error, '申请删除失败'), icon: 'none' })
  }
  finally {
    submitting.value = false
    applicationAction.value = ''
  }
}
</script>

<template>
  <u-popup
    v-model="popupVisible"
    :mode="popupMode"
    :width="popupWidth"
    :height="popupHeight"
    :custom-class="popupCustomClass"
    :border-radius="popupBorderRadius"
    :zoom="!historyDialog"
    :mask-close-able="!submitting"
    :safe-area-inset-bottom="!inlinePresentation"
  >
    <view
      class="workflow-detail-panel"
      :class="{
        'workflow-detail-panel--history-drawer': historyDrawer,
        'workflow-detail-panel--history-dialog': historyDialog,
        'workflow-detail-panel--history-page': historyPage,
        'workflow-detail-panel--page': inlinePresentation,
      }"
    >
      <view v-if="!pagePresentation" class="workflow-detail-panel__header">
        <view class="workflow-detail-panel__title-wrap">
          <view class="workflow-detail-panel__title-line">
            <text class="workflow-detail-panel__title">
              {{ title }}
            </text>
            <template v-if="historyPresentation && detail">
              <text class="workflow-detail-panel__starter">
                {{ detail.instance.starterName || '未知用户' }}
              </text>
              <u-tag
                :text="workflowInstanceStatusMeta(detail.instance.status).label"
                :type="workflowInstanceStatusMeta(detail.instance.status).type"
                custom-class="workflow-detail-panel__status-tag"
                size="mini"
              />
            </template>
          </view>
          <text
            v-if="historyPresentation && detail"
            class="workflow-detail-panel__subtitle workflow-detail-panel__subtitle--business"
          >
            业务编号：{{ detail.instance.businessKey || '-' }}
          </text>
          <text v-else-if="detail" class="workflow-detail-panel__subtitle">
            {{ detail.instance.id }} · 发起于 {{ formatTime(detail.instance.startTime) }}
          </text>
        </view>
        <u-button custom-class="workflow-detail-panel__close app-icon-button" @click="popupVisible = false">
          <u-icon name="close" size="16px" color="#5f6b7a" />
        </u-button>
      </view>

      <view v-if="pagePresentation" class="workflow-detail-panel__page-nav">
        <view
          v-for="item in pageNavigationItems"
          :key="item.key"
          class="workflow-detail-panel__page-nav-item"
          :class="{ active: activeSection === item.key }"
          @click="activeSection = item.key"
        >
          <u-icon :name="item.icon" size="16px" :color="activeSection === item.key ? '#1f2329' : '#86909c'" />
          <text>{{ item.label }}</text>
        </view>
      </view>

      <view v-if="loading" class="workflow-detail-panel__loading">
        <u-loading mode="circle" size="46" />
        <text>加载中...</text>
      </view>
      <template v-else-if="detail">
        <view v-if="!historyPresentation && !pagePresentation" class="workflow-detail-panel__summary">
          <view>
            <text class="workflow-detail-panel__summary-label">
              流程状态
            </text>
            <u-tag
              :text="workflowInstanceStatusMeta(detail.instance.status).label"
              :type="workflowInstanceStatusMeta(detail.instance.status).type"
              custom-class="workflow-detail-panel__status-tag"
              size="mini"
            />
          </view>
          <view>
            <text class="workflow-detail-panel__summary-label">
              发起人
            </text>
            <text>{{ detail.instance.starterName || '未知用户' }}</text>
          </view>
          <view>
            <text class="workflow-detail-panel__summary-label">
              业务编号
            </text>
            <text>{{ detail.instance.businessKey }}</text>
          </view>
          <view v-if="activeTask">
            <text class="workflow-detail-panel__summary-label">
              当前节点
            </text>
            <text>{{ activeTask.nodeName }}</text>
          </view>
        </view>

        <template v-if="historyPresentation">
          <view
            class="workflow-detail-panel__body"
            :class="{
              'workflow-detail-panel__body--history-drawer': historyDrawer,
              'workflow-detail-panel__body--history-page': historyPage,
            }"
          >
            <view
              class="workflow-detail-panel__history-layout"
              :class="{ 'workflow-detail-panel__history-layout--page': historyPage }"
            >
              <view class="workflow-detail-panel__history-form-section">
                <view class="workflow-detail-panel__section-heading">
                  <view class="workflow-detail-panel__section-heading-main">
                    <text class="workflow-detail-panel__section-heading-title">
                      申请表单
                    </text>
                    <text class="workflow-detail-panel__section-heading-meta">
                      发起时提交的表单内容
                    </text>
                  </view>
                  <u-button
                    v-if="showFormDetailAction"
                    custom-class="workflow-detail-panel__form-detail-action"
                    size="mini"
                    type="primary"
                    plain
                    @click="openFormDetail"
                  >
                    <u-icon name="eye" size="14px" color="#2979ff" />
                    <text>详情</text>
                  </u-button>
                </view>
                <scroll-view class="workflow-detail-panel__history-section-scroll" :scroll-y="!historyPage">
                  <WorkflowRuntimeForm
                    v-model="formData"
                    :fields="detail.form || []"
                    :field-access="historyFieldAccess"
                    :readonly="true"
                    readonly-appearance="plain"
                  />
                </scroll-view>
              </view>

              <view class="workflow-detail-panel__history-record-section">
                <view class="workflow-detail-panel__section-heading">
                  <view class="workflow-detail-panel__section-heading-main">
                    <text class="workflow-detail-panel__section-heading-title">
                      流程流转记录
                    </text>
                    <text class="workflow-detail-panel__section-heading-meta">
                      各节点处理状态与流转时间
                    </text>
                  </view>
                  <u-tag
                    :text="workflowInstanceStatusMeta(detail.instance.status).label"
                    :type="workflowInstanceStatusMeta(detail.instance.status).type"
                    custom-class="workflow-detail-panel__status-tag"
                    size="mini"
                  />
                </view>

                <scroll-view class="workflow-detail-panel__history-section-scroll" :scroll-y="!historyPage">
                  <view
                    v-if="canRemindInstance && reminderNodes.length > 0"
                    class="workflow-detail-panel__reminders"
                  >
                    <text class="workflow-detail-panel__subsection-title">
                      当前待处理
                    </text>
                    <view
                      v-for="node in reminderNodes"
                      :key="node.nodeId"
                      class="workflow-detail-panel__reminder-row"
                    >
                      <view class="workflow-detail-panel__reminder-main">
                        <text class="workflow-detail-panel__reminder-node">
                          {{ node.nodeName }}
                        </text>
                        <text class="workflow-detail-panel__reminder-assignees">
                          处理人：{{ reminderAssigneeText(node) }}
                        </text>
                        <text class="workflow-detail-panel__reminder-status">
                          {{ reminderStatusText(node) }}
                        </text>
                      </view>
                      <u-button
                        custom-class="workflow-detail-panel__reminder-button"
                        type="warning"
                        size="mini"
                        :disabled="reminderButtonDisabled(node)"
                        :loading="reminderSubmittingNodeId === node.nodeId"
                        @click="remindNode(node)"
                      >
                        <u-icon name="bell" size="14px" color="#fff" />
                        <text>提醒处理</text>
                      </u-button>
                    </view>
                  </view>

                  <WorkflowNodeProgressList
                    :node-progress="detail.nodeProgress || []"
                    :tasks="sortedTasks"
                    :history="sortedHistory"
                    :instance="detail.instance"
                  />

                  <view class="workflow-detail-panel__history-events">
                    <view
                      class="workflow-detail-panel__subsection-toggle"
                      hover-class="workflow-detail-panel__subsection-toggle--hover"
                      @click="historyEventsExpanded = !historyEventsExpanded"
                    >
                      <text class="workflow-detail-panel__subsection-title workflow-detail-panel__subsection-title--toggle">
                        流转记录
                      </text>
                      <u-icon
                        :name="historyEventsExpanded ? 'arrow-up' : 'arrow-down'"
                        size="14px"
                        color="#86909c"
                      />
                    </view>
                    <template v-if="historyEventsExpanded">
                      <view v-if="sortedHistory.length === 0" class="workflow-detail-panel__section-empty">
                        暂无流程记录
                      </view>
                      <view v-for="item in sortedHistory" :key="item.id" class="workflow-detail-panel__history-item">
                        <view class="workflow-detail-panel__history-dot" />
                        <view>
                          <text class="workflow-detail-panel__history-title">
                            {{ historyLabel(item.eventType) }}
                          </text>
                          <text class="workflow-detail-panel__history-meta">
                            {{ formatTime(item.eventTime) }}{{ historyActorName(item.actorName, item.actorId) ? ` · ${historyActorName(item.actorName, item.actorId)}` : '' }}
                          </text>
                          <text v-if="item.message" class="workflow-detail-panel__history-message">
                            {{ item.message }}
                          </text>
                          <WorkflowImagePicker
                            v-if="item.images?.length"
                            class="workflow-detail-panel__record-images"
                            :model-value="item.images"
                            readonly
                          />
                        </view>
                      </view>
                    </template>
                  </view>
                </scroll-view>
              </view>
            </view>
          </view>

          <view
            v-if="showApplicationActions || showCommentAction || showFormRevisionAction"
            class="workflow-detail-panel__application-actions"
            :class="{ 'workflow-detail-panel__application-actions--comment-only': !showApplicationActions && !showFormRevisionAction }"
          >
            <u-button
              v-if="showApplicationActions"
              custom-class="workflow-detail-panel__application-action"
              plain
              :disabled="!canWithdraw || applicationActionBusy"
              :loading="applicationAction === 'withdraw'"
              @click="withdraw"
            >
              <u-icon name="arrow-left" size="14px" :color="canWithdraw ? '#4e5969' : '#c9cdd4'" />
              <text>撤销</text>
            </u-button>
            <u-button
              v-if="showApplicationActions"
              custom-class="workflow-detail-panel__application-action"
              plain
              :disabled="!canModifyApplication || applicationActionBusy"
              :loading="applicationAction === 'modify'"
              @click="modifyApplication"
            >
              <u-icon name="edit-pen" size="14px" :color="canModifyApplication ? '#1677ff' : '#c9cdd4'" />
              <text>{{ modifyApplicationLabel }}</text>
            </u-button>
            <u-button
              v-if="showFormRevisionAction"
              custom-class="workflow-detail-panel__application-action"
              plain
              :disabled="applicationActionBusy"
              @click="openFormRevision"
            >
              <u-icon name="edit-pen" size="14px" color="#1677ff" />
              <text>修改表单</text>
            </u-button>
            <u-button
              v-if="showCommentAction"
              custom-class="workflow-detail-panel__application-action"
              plain
              :disabled="!canCommentInstance || applicationActionBusy"
              :loading="applicationAction === 'comment'"
              @click="openComment"
            >
              <u-icon name="chat" size="14px" :color="canCommentInstance ? '#0f766e' : '#c9cdd4'" />
              <text>评论</text>
            </u-button>
            <u-button
              v-if="showApplicationActions && hasDeleteApplicationPermission"
              custom-class="workflow-detail-panel__application-action workflow-detail-panel__application-action--danger"
              type="error"
              plain
              :disabled="applicationActionBusy"
              :loading="applicationAction === 'delete'"
              @click="deleteApplication"
            >
              <u-icon name="trash" size="14px" :color="applicationActionBusy ? '#c9cdd4' : '#dc2626'" />
              <text>删除</text>
            </u-button>
          </view>
        </template>

        <template v-else>
          <view v-if="!pagePresentation" class="workflow-detail-panel__tabs">
            <view :class="{ active: activeSection === 'form' }" @click="activeSection = 'form'">
              审批表单
            </view>
            <view :class="{ active: activeSection === 'history' }" @click="activeSection = 'history'">
              流程流转记录
            </view>
          </view>

          <scroll-view
            class="workflow-detail-panel__body"
            :class="{ 'workflow-detail-panel__body--page': pagePresentation }"
            scroll-y
          >
            <view
              v-if="activeSection === 'form'"
              class="workflow-detail-panel__form-section"
              :class="{ 'workflow-detail-panel__form-section--page': pagePresentation }"
            >
              <view :class="{ 'workflow-detail-panel__page-card': pagePresentation }">
                <view v-if="pagePresentation" class="workflow-detail-panel__page-card-head">
                  <view class="workflow-detail-panel__page-title-wrap">
                    <text class="workflow-detail-panel__page-title">
                      {{ title }}
                    </text>
                    <text class="workflow-detail-panel__page-meta">
                      {{ activeTask?.nodeName || '流程处理' }} · 版本 {{ detail.instance.definitionVersion }} · 发起人 {{ detail.instance.starterName || '未知用户' }}
                    </text>
                  </view>
                  <u-tag
                    :text="workflowTaskStatusMeta(canHandle ? 'pending' : activeTask?.status).label"
                    :type="workflowTaskStatusMeta(canHandle ? 'pending' : activeTask?.status).type"
                    custom-class="workflow-detail-panel__status-tag"
                    size="mini"
                  />
                </view>

                <WorkflowRuntimeForm
                  ref="formRef"
                  v-model="formData"
                  :fields="detail.form || []"
                  :field-access="fieldAccess"
                  :field-actions="fieldActions"
                  :readonly="!canHandle"
                  readonly-appearance="plain"
                />
              </view>

              <view v-if="!pagePresentation" class="workflow-detail-panel__tasks">
                <text class="workflow-detail-panel__section-title">
                  审批节点
                </text>
                <view v-for="task in detail.tasks" :key="task.id" class="workflow-detail-panel__task">
                  <view class="workflow-detail-panel__task-dot" />
                  <view class="workflow-detail-panel__task-main">
                    <view class="workflow-detail-panel__task-title">
                      <text>{{ task.nodeName }}</text>
                      <u-tag
                        :text="workflowTaskStatusMeta(task.status).label"
                        :type="workflowTaskStatusMeta(task.status).type"
                        custom-class="workflow-detail-panel__status-tag"
                        size="mini"
                      />
                    </view>
                    <text class="workflow-detail-panel__task-meta">
                      处理人：{{ taskHandlerName(task) }}
                      <template v-if="task.handledAt">
                        · {{ formatTime(task.handledAt) }}
                      </template>
                    </text>
                    <text v-if="task.action || task.comment" class="workflow-detail-panel__task-comment">
                      {{ actionLabel(task.action) }}{{ task.comment ? `：${task.comment}` : '' }}
                    </text>
                    <WorkflowImagePicker
                      v-if="task.images?.length"
                      class="workflow-detail-panel__record-images"
                      :model-value="task.images"
                      readonly
                    />
                  </view>
                </view>
              </view>
            </view>

            <view
              v-else-if="activeSection === 'history'"
              class="workflow-detail-panel__history"
              :class="{ 'workflow-detail-panel__history--page': pagePresentation }"
            >
              <WorkflowNodeProgressList
                :node-progress="detail.nodeProgress || []"
                :tasks="sortedTasks"
                :history="sortedHistory"
                :instance="detail.instance"
              />
              <view class="workflow-detail-panel__history-events workflow-detail-panel__history-events--page">
                <text class="workflow-detail-panel__subsection-title">
                  流转记录
                </text>
                <view v-if="sortedHistory.length === 0" class="workflow-detail-panel__section-empty">
                  暂无流程记录
                </view>
                <view v-for="item in sortedHistory" :key="item.id" class="workflow-detail-panel__history-item">
                  <view class="workflow-detail-panel__history-dot" />
                  <view>
                    <text class="workflow-detail-panel__history-title">
                      {{ historyLabel(item.eventType) }}
                    </text>
                    <text class="workflow-detail-panel__history-meta">
                      {{ formatTime(item.eventTime) }}{{ historyActorName(item.actorName, item.actorId) ? ` · ${historyActorName(item.actorName, item.actorId)}` : '' }}
                    </text>
                    <text v-if="item.message" class="workflow-detail-panel__history-message">
                      {{ item.message }}
                    </text>
                    <WorkflowImagePicker
                      v-if="item.images?.length"
                      class="workflow-detail-panel__record-images"
                      :model-value="item.images"
                      readonly
                    />
                  </view>
                </view>
              </view>
            </view>

            <view v-else class="workflow-detail-panel__graph">
              <WorkflowReadOnlyGraph
                :nodes="detail.nodes || []"
                :edges="detail.edges || []"
              />
            </view>
          </scroll-view>

          <view
            v-if="showTaskActionBar"
            class="workflow-detail-panel__actions"
            :class="{ 'workflow-detail-panel__actions--page': pagePresentation }"
          >
            <u-button
              v-if="pagePresentation"
              custom-class="workflow-detail-panel__action workflow-detail-panel__page-cancel"
              plain
              :disabled="submitting"
              @click="popupVisible = false"
            >
              取消
            </u-button>
            <u-button
              v-if="showCommentAction"
              custom-class="workflow-detail-panel__action workflow-detail-panel__action--comment"
              plain
              :disabled="!canCommentInstance || applicationActionBusy"
              :loading="applicationAction === 'comment'"
              @click="openComment"
            >
              <u-icon name="chat" size="14px" :color="canCommentInstance ? '#0f766e' : '#c9cdd4'" />
              <text>评论</text>
            </u-button>
            <u-button
              v-if="canWithdraw"
              custom-class="workflow-detail-panel__action"
              type="error"
              plain
              :loading="submitting"
              @click="withdraw"
            >
              撤回申请
            </u-button>
            <template v-if="canHandle">
              <u-button
                v-if="canReturn"
                custom-class="workflow-detail-panel__action"
                type="warning"
                plain
                :loading="submitting"
                @click="openReturn"
              >
                退回
              </u-button>
              <u-button
                v-if="activeNodeType === 'approval'"
                custom-class="workflow-detail-panel__action"
                type="error"
                plain
                :loading="submitting"
                @click="openReject"
              >
                驳回
              </u-button>
              <u-button
                custom-class="workflow-detail-panel__action workflow-detail-panel__action--primary"
                type="primary"
                :loading="submitting"
                @click="submitTask(activeNodeType === 'handle' ? 'submit' : 'approve')"
              >
                {{ activeNodeType === 'handle' ? '提交办理' : '同意' }}
              </u-button>
            </template>
          </view>
        </template>
      </template>
      <view v-else class="workflow-detail-panel__empty">
        流程详情不存在或无权查看
      </view>
    </view>
  </u-popup>

  <u-modal
    v-model="actionConfirmVisible"
    custom-class="app-pc-control-scope"
    :title="actionConfirmTitle"
    :content="actionConfirmContent"
    :z-index="10160"
    width="480px"
    confirm-text="确认"
    cancel-text="取消"
    :show-cancel-button="true"
    :mask-close-able="false"
    :content-style="{ whiteSpace: 'pre-wrap', textAlign: 'left', lineHeight: '1.7' }"
    @confirm="resolveActionConfirmation(true)"
    @cancel="resolveActionConfirmation(false)"
  />

  <u-popup
    v-model="returnVisible"
    :mode="returnPopupMode"
    :width="returnPopupWidth"
    custom-class="workflow-return-popup app-pc-control-scope"
    :z-index="10140"
    :border-radius="8"
    :mask-close-able="!submitting && !returnUploading"
    :safe-area-inset-bottom="true"
  >
    <view class="workflow-interaction-dialog workflow-interaction-dialog--return">
      <view class="workflow-interaction-dialog__header">
        <view>
          <text class="workflow-interaction-dialog__title">
            退回流程
          </text>
          <text class="workflow-interaction-dialog__hint">
            退回后将在目标节点重新生成待办，流程继续运行
          </text>
        </view>
        <u-button
          custom-class="workflow-interaction-dialog__close app-icon-button"
          :disabled="submitting || returnUploading"
          @click="closeReturn"
        >
          <u-icon name="close" size="16px" color="#5f6b7a" />
        </u-button>
      </view>
      <view class="workflow-return-targets">
        <text class="workflow-return-targets__label">
          退回节点
        </text>
        <u-radio-group v-model="returnTargetNodeId" wrap>
          <u-radio
            v-for="target in returnTargets"
            :key="target.value"
            :label="target.label"
            :value="target.value"
            :disabled="submitting"
          />
        </u-radio-group>
      </view>
      <WorkflowTextarea
        v-model="returnDraft"
        :disabled="submitting"
        :maxlength="500"
        :min-rows="5"
        :max-rows="10"
        placeholder="填写退回原因"
        count
      />
      <view class="workflow-interaction-dialog__images">
        <WorkflowImagePicker
          v-model="returnImages"
          :disabled="submitting"
          :max-count="9"
          @uploading-change="returnUploading = $event"
        />
        <text class="workflow-interaction-dialog__images-hint">
          最多 9 张，单张不超过 20MB
        </text>
      </view>
      <view class="workflow-interaction-dialog__actions">
        <u-button plain :disabled="submitting || returnUploading" @click="closeReturn">
          取消
        </u-button>
        <u-button type="warning" :loading="submitting" :disabled="returnUploading" @click="submitReturn">
          确认退回
        </u-button>
      </view>
    </view>
  </u-popup>

  <u-popup
    v-model="rejectVisible"
    :mode="rejectPopupMode"
    :width="rejectPopupWidth"
    custom-class="workflow-reject-popup app-pc-control-scope"
    :z-index="10140"
    :border-radius="8"
    :mask-close-able="!submitting && !rejectUploading"
    :safe-area-inset-bottom="true"
  >
    <view class="workflow-interaction-dialog workflow-interaction-dialog--reject">
      <view class="workflow-interaction-dialog__header">
        <view>
          <text class="workflow-interaction-dialog__title">
            驳回流程
          </text>
          <text class="workflow-interaction-dialog__hint">
            请填写驳回原因，可附加图片说明
          </text>
        </view>
        <u-button
          custom-class="workflow-interaction-dialog__close app-icon-button"
          :disabled="submitting || rejectUploading"
          @click="closeReject"
        >
          <u-icon name="close" size="16px" color="#5f6b7a" />
        </u-button>
      </view>
      <WorkflowTextarea
        v-model="rejectDraft"
        :disabled="submitting"
        :maxlength="500"
        :min-rows="5"
        :max-rows="10"
        placeholder="填写驳回原因"
        count
      />
      <view class="workflow-interaction-dialog__images">
        <WorkflowImagePicker
          v-model="rejectImages"
          :disabled="submitting"
          :max-count="9"
          @uploading-change="rejectUploading = $event"
        />
        <text class="workflow-interaction-dialog__images-hint">
          最多 9 张，单张不超过 20MB
        </text>
      </view>
      <view class="workflow-interaction-dialog__actions">
        <u-button plain :disabled="submitting || rejectUploading" @click="closeReject">
          取消
        </u-button>
        <u-button type="error" :loading="submitting" :disabled="rejectUploading" @click="submitReject">
          确认驳回
        </u-button>
      </view>
    </view>
  </u-popup>

  <u-popup
    v-model="commentVisible"
    :mode="commentPopupMode"
    :width="commentPopupWidth"
    custom-class="workflow-comment-popup app-pc-control-scope"
    :z-index="10140"
    :border-radius="8"
    :mask-close-able="!commentSubmitting && !commentUploading"
    :safe-area-inset-bottom="true"
  >
    <view class="workflow-interaction-dialog workflow-interaction-dialog--comment">
      <view class="workflow-interaction-dialog__header">
        <view>
          <text class="workflow-interaction-dialog__title">
            添加评论
          </text>
          <text class="workflow-interaction-dialog__hint">
            支持文字、图片或两者一起发布
          </text>
        </view>
        <u-button
          custom-class="workflow-interaction-dialog__close app-icon-button"
          :disabled="commentSubmitting || commentUploading"
          @click="closeComment"
        >
          <u-icon name="close" size="16px" color="#5f6b7a" />
        </u-button>
      </view>
      <WorkflowTextarea
        v-model="commentDraft"
        :disabled="commentSubmitting"
        :maxlength="500"
        :min-rows="5"
        :max-rows="10"
        placeholder="填写评论内容"
        count
      />
      <view class="workflow-interaction-dialog__images">
        <WorkflowImagePicker
          v-model="commentImages"
          :disabled="commentSubmitting"
          :max-count="9"
          @uploading-change="commentUploading = $event"
        />
        <text class="workflow-interaction-dialog__images-hint">
          最多 9 张，单张不超过 20MB
        </text>
      </view>
      <view v-if="commentNotificationGroups.length > 0" class="workflow-comment-notification">
        <text class="workflow-comment-notification__label">
          通知对象
        </text>
        <WorkflowParticipantSelect
          v-model="commentNotificationUserIds"
          :groups="commentNotificationGroups"
          :disabled="commentSubmitting"
          placeholder="请选择需要通知的流程参与人"
        />
        <text class="workflow-comment-notification__label workflow-comment-notification__label--channels">
          通知方式
        </text>
        <u-checkbox-group v-model="commentNotificationChannels" class="workflow-comment-notification__channels">
          <u-checkbox value="in_app" label="站内信" :disabled="commentSubmitting" />
          <u-checkbox value="dingtalk_oa" label="钉钉" :disabled="commentSubmitting" />
        </u-checkbox-group>
      </view>
      <view class="workflow-interaction-dialog__actions">
        <u-button plain :disabled="commentSubmitting || commentUploading" @click="closeComment">
          取消
        </u-button>
        <u-button type="primary" :loading="commentSubmitting" :disabled="commentUploading" @click="submitComment">
          发布评论
        </u-button>
      </view>
    </view>
  </u-popup>
</template>

<style lang="scss" scoped>
.workflow-detail-panel {
  width: 100%;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f6f8fb;
}

:global(.workflow-detail-popup .u-mode-center-box) {
  max-width: 900px;
}

:global(.workflow-detail-popup--history-drawer .u-drawer-right) {
  width: clamp(420px, 38vw, 620px) !important;
  max-width: calc(100vw - 32px);
  box-shadow: -8px 0 28px rgba(31, 35, 41, 0.16);
}

:global(.workflow-detail-popup--history-dialog .u-mode-center-box) {
  width: min(720px, calc(100vw - 32px)) !important;
  height: min(760px, calc(100vh - 32px)) !important;
  max-height: calc(100vh - 32px);
  box-shadow: 0 12px 36px rgba(31, 35, 41, 0.2);
}

:deep(.workflow-detail-popup--page) {
  width: 100% !important;
  height: 100% !important;
  min-height: 0;
}

.workflow-detail-panel--history-drawer {
  background: #fff;
}

.workflow-detail-panel--history-dialog {
  border-radius: 8px;
}

.workflow-detail-panel--page {
  min-height: 0;
}

.workflow-detail-panel__page-nav {
  flex: 0 0 auto;
  min-height: 54px;
  padding: 0 24px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: stretch;
  gap: 30px;
  overflow-x: auto;
  background: #fff;
}

.workflow-detail-panel__page-nav-item {
  flex: 0 0 auto;
  border-bottom: 2px solid transparent;
  display: flex;
  align-items: center;
  gap: 7px;
  color: #86909c;
  font-size: 14px;
  cursor: pointer;
}

.workflow-detail-panel__page-nav-item.active {
  border-bottom-color: #1f2329;
  color: #1f2329;
  font-weight: 700;
}

.workflow-detail-panel__header {
  flex: 0 0 auto;
  min-height: 72px;
  padding: 0 24px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  background: #fff;
}

.workflow-detail-panel--history-drawer .workflow-detail-panel__header {
  padding-top: 12px;
  padding-bottom: 12px;
}

.workflow-detail-panel__title-wrap {
  flex: 1 1 auto;
  min-width: 0;
  display: grid;
  gap: 4px;
}

.workflow-detail-panel__title-line {
  min-width: 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.workflow-detail-panel__title,
.workflow-detail-panel__subtitle {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-detail-panel__title {
  min-width: 0;
  color: #1f2329;
  font-size: 18px;
  font-weight: 800;
}

.workflow-detail-panel__starter {
  max-width: 45%;
  display: block;
  overflow: hidden;
  color: #4e5969;
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-detail-panel__subtitle {
  color: #86909c;
  font-size: 12px;
}

.workflow-detail-panel__subtitle--business {
  line-height: 1.4;
  overflow-wrap: anywhere;
  white-space: normal;
}

.workflow-detail-panel__close,
:deep(.workflow-detail-panel__close) {
  width: 34px;
  height: 34px;
  min-height: 34px;
  margin: 0;
  padding: 0;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  background: #fff;
}

.workflow-detail-panel__close::after,
:deep(.workflow-detail-panel__close)::after {
  display: none;
}

.workflow-detail-panel__loading,
.workflow-detail-panel__empty {
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  color: $u-tips-color;
  font-size: 13px;
}

:deep(.workflow-detail-panel__status-tag) {
  width: auto !important;
  height: 24px;
  min-height: 24px;
  padding: 0 8px !important;
  border-radius: 4px !important;
  font-size: 12px !important;
  line-height: 22px !important;
  white-space: nowrap;
  box-sizing: border-box;
}

.workflow-detail-panel--history-drawer {
  :deep(.workflow-form__field-label) {
    min-height: 21px;
    margin-bottom: 5px;
    gap: 4px;
    font-size: 13px;
  }

  :deep(.workflow-form__group-header) {
    margin-bottom: 9px;
    gap: 5px;
    font-size: 14px;
  }

  :deep(.workflow-form__label) {
    padding: 5px;
    font-size: 15px;
    line-height: 1.5;
  }

  :deep(.workflow-form__description) {
    padding: 5px;
    font-size: 12px;
  }
}

.workflow-detail-panel__summary {
  flex: 0 0 auto;
  padding: 14px 24px;
  border-bottom: 1px solid #edf0f5;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  background: #fff;
  color: #4e5969;
  font-size: 13px;
}

.workflow-detail-panel__summary > view {
  min-width: 0;
  display: grid;
  gap: 5px;
}

.workflow-detail-panel__summary-label {
  color: #86909c;
  font-size: 11px;
}

.workflow-detail-panel__tabs {
  flex: 0 0 auto;
  min-height: 48px;
  padding: 0 24px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: stretch;
  gap: 28px;
  background: #fff;
}

.workflow-detail-panel__tabs > view {
  border-bottom: 2px solid transparent;
  display: flex;
  align-items: center;
  color: #6b7785;
  font-size: 14px;
  cursor: pointer;
}

.workflow-detail-panel__tabs > view.active {
  border-bottom-color: #2979ff;
  color: #1f2329;
  font-weight: 700;
}

.workflow-detail-panel__body {
  min-height: 0;
  flex: 1 1 auto;
  height: 100%;
}

.workflow-detail-panel__body--page {
  width: min(var(--app-pc-content-max-width, 1080px), 100%);
  margin: 0 auto;
  padding: 24px;
  box-sizing: border-box;
}

.workflow-detail-panel__body--history-drawer {
  display: flex;
  overflow: hidden;
  background: #f6f8fb;
}

.workflow-detail-panel__body--history-page {
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  touch-action: pan-y;
  background: #f6f8fb;
}

.workflow-detail-panel--history-page :deep(.u-input--disabled),
.workflow-detail-panel--history-page :deep(.u-textarea--disabled),
.workflow-detail-panel--history-page :deep(.workflow-picker--disabled),
.workflow-detail-panel--history-page :deep(.u-switch--disabled) {
  pointer-events: none;
}

.workflow-detail-panel__history-layout {
  width: 100%;
  height: 100%;
  min-height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
  box-sizing: border-box;
}

.workflow-detail-panel__history-layout--page {
  width: min(var(--app-pc-content-max-width, 1080px), 100%);
  height: auto;
  min-height: 100%;
  margin: 0 auto;
  padding: 16px;
  gap: 10px;
  overflow: visible;
  background: transparent;
}

.workflow-detail-panel__history-layout--page .workflow-detail-panel__history-form-section,
.workflow-detail-panel__history-layout--page .workflow-detail-panel__history-record-section {
  flex: 0 0 auto;
  overflow: visible;
}

.workflow-detail-panel__history-layout--page .workflow-detail-panel__history-record-section {
  border-top: 0;
}

.workflow-detail-panel__history-layout--page .workflow-detail-panel__history-section-scroll {
  height: auto;
  flex: 0 0 auto;
  overflow: visible;
}

.workflow-detail-panel__history-form-section,
.workflow-detail-panel__history-record-section {
  min-height: 0;
  flex: 1 1 50%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
}

.workflow-detail-panel__history-record-section {
  border-top: 10px solid #f6f8fb;
}

.workflow-detail-panel__section-heading {
  flex: 0 0 auto;
  min-height: 44px;
  margin: 0;
  padding: 14px 20px 12px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.workflow-detail-panel__history-section-scroll {
  min-height: 0;
  height: 0;
  flex: 1 1 auto;
  padding: 16px 20px 20px;
  box-sizing: border-box;
}

.workflow-detail-panel__section-heading-main {
  min-width: 0;
}

:deep(.workflow-detail-panel__form-detail-action) {
  width: auto;
  min-width: 68px;
  height: 30px;
  margin: 0;
  padding: 0 12px;
  flex-shrink: 0;
  gap: 5px;
}

.workflow-detail-panel__section-heading-title,
.workflow-detail-panel__section-heading-meta {
  display: block;
}

.workflow-detail-panel__section-heading-title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.workflow-detail-panel__section-heading-meta {
  margin-top: 5px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.workflow-detail-panel__history-tasks,
.workflow-detail-panel__history-events {
  min-width: 0;
}

.workflow-detail-panel__history-events {
  margin-top: 4px;
  padding-top: 20px;
  border-top: 1px solid #edf0f5;
}

.workflow-detail-panel__history-events--page {
  margin-top: 24px;
}

.workflow-detail-panel__reminders {
  margin-bottom: 20px;
  padding-bottom: 4px;
  border-bottom: 1px solid #edf0f5;
}

.workflow-detail-panel__reminder-row {
  min-width: 0;
  padding: 12px 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.workflow-detail-panel__reminder-row + .workflow-detail-panel__reminder-row {
  border-top: 1px solid #f2f3f5;
}

.workflow-detail-panel__reminder-main {
  min-width: 0;
  flex: 1 1 auto;
}

.workflow-detail-panel__reminder-node,
.workflow-detail-panel__reminder-assignees,
.workflow-detail-panel__reminder-status {
  display: block;
}

.workflow-detail-panel__reminder-node {
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
}

.workflow-detail-panel__reminder-assignees,
.workflow-detail-panel__reminder-status {
  margin-top: 4px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

:deep(.workflow-detail-panel__reminder-button) {
  display: flex !important;
  align-items: center;
  justify-content: center;
  gap: 4px; /* 控制图标和文字的间距 */
  width: auto;
  min-width: 92px;
  margin: 0;
}
/* 图标和文字都设为 inline-block 并 middle 对齐 */
:deep(.workflow-detail-panel__reminder-button .u-icon) {
  vertical-align: middle !important;
  margin: 0 !important;
  display: inline-block !important;
}

:deep(.workflow-detail-panel__reminder-button text) {
  vertical-align: middle !important;
  line-height: 1 !important;
  display: inline-block !important;
}

.workflow-detail-panel__subsection-toggle {
  min-height: 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  cursor: pointer;
}

.workflow-detail-panel__subsection-toggle--hover {
  opacity: 0.72;
}

.workflow-detail-panel__subsection-title {
  display: block;
  margin-bottom: 18px;
  color: #4e5969;
  font-size: 13px;
  font-weight: 700;
}

.workflow-detail-panel__subsection-title--toggle {
  margin-bottom: 0;
}

.workflow-detail-panel__section-empty {
  padding: 20px 0;
  color: #86909c;
  font-size: 13px;
  text-align: center;
}

.workflow-detail-panel__form-section,
.workflow-detail-panel__history {
  padding: 24px;
  box-sizing: border-box;
}

.workflow-detail-panel__form-section--page {
  padding: 0;
}

.workflow-detail-panel__page-card,
.workflow-detail-panel__history--page {
  padding: 20px;
  border: 1px solid #dfe5ee;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(31, 35, 41, 0.05);
  box-sizing: border-box;
}

.workflow-detail-panel__page-card-head {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
}

.workflow-detail-panel__page-title-wrap {
  min-width: 0;
  display: grid;
  gap: 5px;
}

.workflow-detail-panel__page-title,
.workflow-detail-panel__page-meta {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-detail-panel__page-title {
  color: #1f2329;
  font-size: 17px;
  font-weight: 800;
}

.workflow-detail-panel__page-meta {
  color: #86909c;
  font-size: 12px;
}

.workflow-detail-panel__graph {
  height: 100%;
  min-height: 480px;
  box-sizing: border-box;
}

.workflow-detail-panel__tasks {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e5eaf3;
}

.workflow-detail-panel__section-title {
  display: block;
  margin-bottom: 14px;
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.workflow-detail-panel__task,
.workflow-detail-panel__history-item {
  position: relative;
  padding: 0 0 22px 24px;
}

.workflow-detail-panel__task:not(:last-child)::before,
.workflow-detail-panel__history-item:not(:last-child)::before {
  position: absolute;
  top: 12px;
  bottom: -2px;
  left: 5px;
  width: 1px;
  background: #cbd5e1;
  content: '';
}

.workflow-detail-panel__task-dot,
.workflow-detail-panel__history-dot {
  position: absolute;
  top: 6px;
  left: 0;
  width: 11px;
  height: 11px;
  border: 2px solid #fff;
  border-radius: 50%;
  background: #2979ff;
  box-shadow: 0 0 0 1px #9dbdff;
}

.workflow-detail-panel__task-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.workflow-detail-panel__task-meta,
.workflow-detail-panel__task-comment,
.workflow-detail-panel__history-meta,
.workflow-detail-panel__history-message {
  display: block;
  margin-top: 5px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.55;
}

.workflow-detail-panel__task-comment,
.workflow-detail-panel__history-message {
  color: #4e5969;
}

.workflow-detail-panel__history-title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.workflow-detail-panel__record-images {
  margin-top: 9px;
}

.workflow-detail-panel__actions {
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
  background: #fff;
  box-shadow: 0 -4px 16px rgba(31, 35, 41, 0.05);
  box-sizing: border-box;
}

.workflow-detail-panel__application-actions {
  position: relative;
  z-index: 20;
  flex: 0 0 auto;
  min-height: 64px;
  padding: 10px 12px;
  border-top: 1px solid #dfe5ee;
  display: grid;
  grid-template-columns: repeat(auto-fit, 120px);
  align-items: center;
  justify-content: start;
  gap: 8px;
  background: #fff;
  box-shadow: 0 -4px 16px rgba(31, 35, 41, 0.06);
  box-sizing: border-box;
}

.workflow-detail-panel__application-actions--comment-only {
  grid-template-columns: 120px;
  justify-content: start;
}

.workflow-detail-panel--history-drawer .workflow-detail-panel__application-actions--comment-only {
  min-height: 56px;
  padding: 10px 16px;
  grid-template-columns: 120px;
  border-top-color: #eef1f4;
  box-shadow: 0 -2px 10px rgba(31, 35, 41, 0.04);
}

.workflow-detail-panel--history-drawer
  .workflow-detail-panel__application-actions--comment-only
  :deep(.workflow-detail-panel__application-action) {
  height: 40px;
  border-color: #0f766e;
  border-radius: 4px;
  background: #fff;
  color: #0f766e;
  font-weight: 500;
  line-height: 38px;
}

.workflow-detail-panel--history-drawer
  .workflow-detail-panel__application-actions--comment-only
  :deep(.workflow-detail-panel__application-action:active) {
  background: #f0fdfa;
}

:deep(.workflow-detail-panel__application-action) {
  width: 100%;
  min-width: 0;
  height: 40px;
  margin: 0;
  padding: 0 6px;
  font-size: 13px;
  line-height: 38px;
  white-space: nowrap;
  box-sizing: border-box;
}

:deep(.workflow-detail-panel__application-action text) {
  margin-left: 4px;
}

.workflow-interaction-dialog {
  width: 480px;
  max-width: calc(100vw - 32px);
  padding: 20px;
  background: #fff;
  box-sizing: border-box;
}

.workflow-interaction-dialog__header {
  min-height: 32px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.workflow-interaction-dialog__title,
.workflow-interaction-dialog__hint {
  display: block;
}

.workflow-interaction-dialog__title {
  color: #1f2329;
  font-size: 17px;
  font-weight: 600;
}

.workflow-interaction-dialog__hint {
  margin-top: 4px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

:deep(.workflow-interaction-dialog__close) {
  width: 32px;
  min-width: 32px;
  height: 32px;
  margin: 0;
  padding: 0;
  border: 0;
  background: transparent;
}

.workflow-interaction-dialog__images {
  margin-top: 14px;
}

.workflow-interaction-dialog__images-hint {
  display: block;
  margin-top: 7px;
  color: #86909c;
  font-size: 12px;
}

.workflow-comment-notification {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #edf0f3;
}

.workflow-comment-notification__label {
  display: block;
}

.workflow-comment-notification__label {
  color: #1f2329;
  font-size: 13px;
  font-weight: 500;
}

.workflow-comment-notification__label--channels {
  margin-top: 14px;
}

.workflow-comment-notification :deep(.workflow-participant-select) {
  margin-top: 8px;
}

.workflow-comment-notification__channels {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 20px;
}

.workflow-comment-notification__channels :deep(.u-checkbox) {
  margin: 0;
}

.workflow-comment-notification__channels {
  margin-top: 9px;
}

.workflow-return-targets {
  margin-bottom: 16px;
}

.workflow-return-targets__label {
  display: block;
  margin-bottom: 9px;
  color: #1f2329;
  font-size: 13px;
  font-weight: 500;
}

.workflow-return-targets :deep(.u-radio) {
  min-width: 100%;
  margin: 0 0 8px;
}

.workflow-interaction-dialog__actions {
  margin-top: 18px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.workflow-interaction-dialog__actions :deep(.u-btn) {
  width: auto;
  min-width: 88px;
  margin: 0;
}

.workflow-detail-panel__action,
:deep(.workflow-detail-panel__action) {
  width: auto;
  min-width: 104px;
  margin: 0;
}

@media screen and (max-width: 768px) {
  :global(.workflow-detail-popup--history-drawer .u-drawer-right) {
    width: min(560px, 94vw) !important;
    max-width: 94vw;
  }

  .workflow-interaction-dialog--comment {
    width: 100%;
    max-width: none;
    max-height: 92vh;
    padding: 16px;
    border-radius: 8px 8px 0 0;
    overflow-y: auto;
  }

  .workflow-interaction-dialog--comment .workflow-interaction-dialog__actions {
    position: sticky;
    z-index: 2;
    bottom: 0;
    margin: 18px -16px -16px;
    padding: 12px 16px;
    border-top: 1px solid #edf0f3;
    background: #fff;
  }

  .workflow-interaction-dialog--comment .workflow-interaction-dialog__actions :deep(.u-btn) {
    min-width: 0;
    flex: 1 1 0;
  }

  .workflow-interaction-dialog--return {
    width: 100%;
    max-width: none;
    max-height: 92vh;
    padding: 16px;
    border-radius: 8px 8px 0 0;
    overflow-y: auto;
  }

  .workflow-interaction-dialog--return .workflow-interaction-dialog__actions {
    position: sticky;
    z-index: 2;
    bottom: 0;
    margin: 18px -16px -16px;
    padding: 12px 16px;
    border-top: 1px solid #edf0f3;
    background: #fff;
  }

  .workflow-interaction-dialog--return .workflow-interaction-dialog__actions :deep(.u-btn) {
    min-width: 0;
    flex: 1 1 0;
  }

  .workflow-interaction-dialog--reject {
    width: 100%;
    max-width: none;
    max-height: 92vh;
    padding: 16px;
    border-radius: 8px 8px 0 0;
    overflow-y: auto;
  }

  .workflow-interaction-dialog--reject .workflow-interaction-dialog__actions {
    position: sticky;
    z-index: 2;
    bottom: 0;
    margin: 18px -16px -16px;
    padding: 12px 16px;
    border-top: 1px solid #edf0f3;
    background: #fff;
  }

  .workflow-interaction-dialog--reject .workflow-interaction-dialog__actions :deep(.u-btn) {
    min-width: 0;
    flex: 1 1 0;
  }

  .workflow-detail-panel__header {
    min-height: 60px;
    padding: 0 14px 0 16px;
  }

  .workflow-detail-panel__page-nav {
    min-height: 50px;
    padding: 0 14px;
    gap: 22px;
  }

  .workflow-detail-panel__page-nav-item {
    font-size: 13px;
  }

  .workflow-detail-panel__summary {
    padding: 12px 16px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workflow-detail-panel__tabs {
    padding: 0 16px;
  }

  .workflow-detail-panel__form-section,
  .workflow-detail-panel__history {
    padding: 16px 12px;
  }

  .workflow-detail-panel__body--page {
    padding: 16px 12px;
  }

  .workflow-detail-panel__form-section--page {
    padding: 0;
  }

  .workflow-detail-panel__page-card,
  .workflow-detail-panel__history--page {
    padding: 12px;
  }

  .workflow-detail-panel__page-card-head {
    margin-bottom: 16px;
    padding-bottom: 12px;
  }

  .workflow-detail-panel__page-meta {
    white-space: normal;
  }

  .workflow-detail-panel__history-layout {
    padding: 0;
  }

  .workflow-detail-panel__history-layout--page {
    padding: 10px 0;
  }

  .workflow-detail-panel__history-form-section,
  .workflow-detail-panel__history-record-section {
    padding: 0;
  }

  .workflow-detail-panel__section-heading {
    padding: 12px 14px 10px;
  }

  .workflow-detail-panel__history-section-scroll {
    padding: 14px 12px 16px;
  }

  .workflow-detail-panel__actions {
    padding: 10px 12px;
    gap: 8px;
  }

  .workflow-detail-panel__application-actions--comment-only {
    grid-template-columns: 120px;
  }

  .workflow-detail-panel--history-drawer .workflow-detail-panel__application-actions--comment-only {
    padding: 10px 12px;
    grid-template-columns: 120px;
  }

  .workflow-detail-panel__action,
  :deep(.workflow-detail-panel__action) {
    flex: 1 1 0;
    min-width: 0;
  }

  .workflow-detail-panel__actions :deep(.workflow-detail-panel__action) {
    height: 46px;
    margin: 0;
    padding: 0 8px;
    font-size: 15px;
    line-height: 1;
    white-space: nowrap;
    box-sizing: border-box;
  }

  .workflow-detail-panel__actions :deep(.workflow-detail-panel__action text) {
    flex-shrink: 0;
    white-space: nowrap;
  }

  .workflow-detail-panel__actions :deep(.workflow-detail-panel__action--comment .u-icon) {
    display: none;
  }
}
</style>
