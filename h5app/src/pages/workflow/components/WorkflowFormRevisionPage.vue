<script setup lang="ts">
import type {
  WorkflowFieldAccessMap,
  WorkflowFieldActionsMap,
  WorkflowFormData,
  WorkflowInstanceDetail,
  WorkflowNotificationChannel,
} from '@/types/workflow'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { getWorkflowInstance, reviseWorkflowInstanceForm } from '@/api/workflow'
import { useAppContentStore, useDingtalkAuthStore } from '@/stores'
import {
  initialWorkflowFormData,
  workflowFieldAccessMap,
  workflowFieldActionsMap,
  writableWorkflowFormData,
} from '../workflow-form'
import { workflowFormRevisionInstanceIdFromContentKey } from '../workflow-route-keys'
import WorkflowParticipantSelect from './WorkflowParticipantSelect.vue'
import WorkflowRuntimeForm from './WorkflowRuntimeForm.vue'
import WorkflowTextarea from './WorkflowTextarea.vue'

interface RuntimeFormExposed {
  validate: () => { valid: boolean, errors: Record<string, string> }
}

interface ParticipantGroup {
  nodeId: string
  nodeName: string
  users: Array<{ userId: string, userName: string }>
}

const props = defineProps<{
  contentKey: string
}>()

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const detail = ref<WorkflowInstanceDetail | null>(null)
const formData = ref<WorkflowFormData>({})
const originalData = ref<WorkflowFormData>({})
const formRef = ref<RuntimeFormExposed | null>(null)
const loading = ref(true)
const loadError = ref('')
const submitting = ref(false)
const notificationVisible = ref(false)
const notifyEnabled = ref(true)
const notificationUserIds = ref<string[]>([])
const notificationChannels = ref<WorkflowNotificationChannel[]>(['in_app'])
const reason = ref('')
const initialSnapshot = ref('')
let unregisterCloseGuard: (() => void) | null = null

const instanceId = computed(() => workflowFormRevisionInstanceIdFromContentKey(props.contentKey))
const currentUserId = computed(() => String(auth.user?.workflowActorId || auth.user?.id || ''))
const canRevise = computed(() => Boolean(
  detail.value?.instance.status === 'running'
  && detail.value.formRevision?.allowed
  && auth.hasButtonPermission('dingtalk_h5:button:workflow:form-revise')
  && auth.hasApiPermission('dingtalk_h5:api:workflow:form-revise'),
))
const title = computed(() => detail.value?.instance.definitionName || '修改流程表单')
const fieldAccess = computed<WorkflowFieldAccessMap>(() => workflowFieldAccessMap(
  detail.value?.form || [],
  detail.value?.formRevision?.fieldPermissions || [],
  'hidden',
))
const fieldActions = computed<WorkflowFieldActionsMap>(() => workflowFieldActionsMap(
  detail.value?.form || [],
  detail.value?.formRevision?.fieldPermissions || [],
))
const hasUnsavedChanges = computed(() => Boolean(
  initialSnapshot.value && currentSnapshot() !== initialSnapshot.value,
))

const notificationGroups = computed<ParticipantGroup[]>(() => {
  const current = detail.value
  if (!current)
    return []
  const groups = new Map<string, ParticipantGroup>()
  const seenUsers = new Set<string>([currentUserId.value])
  const nodeNames = new Map((current.nodes || []).map(node => [node.id, node.name]))
  const addUser = (nodeId: string, fallbackNodeName: string, userId: string, userName: string) => {
    const normalizedUserId = String(userId || '').trim()
    if (!normalizedUserId || seenUsers.has(normalizedUserId))
      return
    const resolvedUserName = String(userName || current.userNames?.[normalizedUserId] || '').trim()
    if (!resolvedUserName)
      return
    const key = String(nodeId || current.startNodeId || 'start').trim()
    const group = groups.get(key) || {
      nodeId: key,
      nodeName: String(nodeNames.get(key) || fallbackNodeName || '流程参与人').trim(),
      users: [],
    }
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
  return [...groups.values()].sort((left, right) => (
    (nodeOrder.get(left.nodeId) ?? Number.MAX_SAFE_INTEGER)
    - (nodeOrder.get(right.nodeId) ?? Number.MAX_SAFE_INTEGER)
  ))
})

onMounted(() => {
  unregisterCloseGuard = appContent.registerTabCloseGuard(props.contentKey, {
    hasUnsavedChanges: () => hasUnsavedChanges.value,
  })
  void loadPage()
})

onBeforeUnmount(() => {
  unregisterCloseGuard?.()
})

function writableData(values: WorkflowFormData) {
  return writableWorkflowFormData(detail.value?.form || [], values, fieldAccess.value)
}

function currentSnapshot() {
  return JSON.stringify(writableData(formData.value))
}

function changedPatch() {
  const current = writableData(formData.value)
  const original = writableData(originalData.value)
  const patch: WorkflowFormData = {}
  for (const [field, value] of Object.entries(current)) {
    if (JSON.stringify(value) !== JSON.stringify(original[field]))
      patch[field] = value
  }
  return patch
}

async function loadPage() {
  if (!instanceId.value) {
    loadError.value = '流程实例无效'
    loading.value = false
    return
  }
  loading.value = true
  loadError.value = ''
  try {
    const response = await getWorkflowInstance(instanceId.value)
    if (!response?.data)
      throw new Error('流程详情不存在')
    detail.value = response.data
    if (!canRevise.value) {
      loadError.value = '当前流程不可修改，可能已结束或权限已变更'
      return
    }
    const initial = initialWorkflowFormData(response.data.form || [], response.data.formData || {})
    originalData.value = initialWorkflowFormData(response.data.form || [], initial)
    formData.value = initialWorkflowFormData(response.data.form || [], initial)
    initialSnapshot.value = currentSnapshot()
  }
  catch (error) {
    loadError.value = requestErrorMessage(error, '流程表单加载失败')
  }
  finally {
    loading.value = false
  }
}

function requestClose() {
  appContent.requestCloseTab(props.contentKey)
}

function openSubmitDialog() {
  if (!canRevise.value || submitting.value)
    return
  const validation = formRef.value?.validate()
  if (validation && !validation.valid) {
    uni.showToast({ title: '请检查表单填写内容', icon: 'none' })
    return
  }
  if (Object.keys(changedPatch()).length === 0) {
    uni.showToast({ title: '请至少修改一个表单字段', icon: 'none' })
    return
  }
  const pendingAssignees = new Set(
    (detail.value?.tasks || [])
      .filter(task => task.status === 'pending' && task.assigneeId !== currentUserId.value)
      .map(task => task.assigneeId),
  )
  const selectable = new Set(notificationGroups.value.flatMap(group => group.users.map(user => user.userId)))
  notificationUserIds.value = [...pendingAssignees].filter(userId => selectable.has(userId))
  notifyEnabled.value = notificationUserIds.value.length > 0
  reason.value = ''
  notificationChannels.value = ['in_app']
  notificationVisible.value = true
}

async function submitRevision() {
  const current = detail.value
  const normalizedReason = reason.value.trim()
  if (!current || submitting.value)
    return
  if (!normalizedReason) {
    uni.showToast({ title: '请填写修改原因', icon: 'none' })
    return
  }
  if (Array.from(normalizedReason).length > 500) {
    uni.showToast({ title: '修改原因不能超过500个字符', icon: 'none' })
    return
  }
  if (notifyEnabled.value && notificationUserIds.value.length === 0) {
    uni.showToast({ title: '请选择通知对象', icon: 'none' })
    return
  }
  if (notifyEnabled.value && notificationChannels.value.length === 0) {
    uni.showToast({ title: '请选择通知方式', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    await reviseWorkflowInstanceForm(current.instance.id, {
      expectedRevision: current.formRevision.revision,
      formData: changedPatch(),
      reason: normalizedReason,
      notification: notifyEnabled.value
        ? { userIds: notificationUserIds.value, channels: notificationChannels.value }
        : undefined,
    })
    notificationVisible.value = false
    appContent.removeDynamicTab(props.contentKey)
    appContent.focusWorkflowTab('handled')
    appContent.switchContent('workflow')
    appContent.requestRefresh()
    uni.showToast({ title: '表单修改已提交', icon: 'success' })
  }
  catch (error) {
    uni.showToast({ title: requestErrorMessage(error, '表单修改提交失败'), icon: 'none' })
  }
  finally {
    submitting.value = false
  }
}

function requestErrorMessage(error: unknown, fallback: string) {
  if (!error || typeof error !== 'object')
    return fallback
  const response = error as Record<string, unknown>
  const payload = response.data && typeof response.data === 'object' && !Array.isArray(response.data)
    ? response.data as Record<string, unknown>
    : response
  const message = payload.msg ?? payload.message
  return typeof message === 'string' && message.trim() ? message.trim() : fallback
}
</script>

<template>
  <view class="workflow-form-revision-page app-pc-control-scope">
    <view class="workflow-form-revision-page__header">
      <view class="workflow-form-revision-page__heading">
        <text class="workflow-form-revision-page__title">
          {{ title }}
        </text>
        <text v-if="detail" class="workflow-form-revision-page__meta">
          业务编号：{{ detail.instance.businessKey || '-' }} · 表单版本 {{ detail.formRevision.revision }}
        </text>
      </view>
      <u-tag v-if="detail" text="审批中" type="warning" size="mini" />
    </view>

    <view v-if="loading" class="workflow-form-revision-page__state">
      <u-loading mode="circle" size="24px" />
      <text>正在加载流程表单...</text>
    </view>
    <view v-else-if="loadError" class="workflow-form-revision-page__state workflow-form-revision-page__state--error">
      <u-icon name="info-circle" size="28px" color="#f56c6c" />
      <text>{{ loadError }}</text>
    </view>
    <scroll-view v-else-if="detail" scroll-y class="workflow-form-revision-page__body">
      <view class="workflow-form-revision-page__content">
        <view class="workflow-form-revision-page__section-head">
          <text class="workflow-form-revision-page__section-title">
            修改申请表单
          </text>
          <text class="workflow-form-revision-page__section-desc">
            展示你已办理节点可见的字段，仅可修改配置为可写的字段
          </text>
        </view>
        <WorkflowRuntimeForm
          ref="formRef"
          v-model="formData"
          class="workflow-form-revision-page__form app-workflow-form app-pc-control-scope"
          :fields="detail.form || []"
          :field-access="fieldAccess"
          :field-actions="fieldActions"
          :readonly="false"
          readonly-appearance="plain"
        />
      </view>
    </scroll-view>

    <view v-if="detail && !loading && !loadError" class="workflow-form-revision-page__actions">
      <u-button plain :disabled="submitting" @click="requestClose">
        取消
      </u-button>
      <u-button type="primary" :loading="submitting" :disabled="!canRevise" @click="openSubmitDialog">
        提交修改
      </u-button>
    </view>

    <u-popup
      v-model="notificationVisible"
      mode="center"
      width="540px"
      custom-class="workflow-form-revision-popup app-pc-control-scope"
      :z-index="10140"
      :border-radius="8"
      :mask-close-able="!submitting"
    >
      <view class="workflow-form-revision-dialog">
        <view class="workflow-form-revision-dialog__header">
          <view>
            <text class="workflow-form-revision-dialog__title">
              提交表单修改
            </text>
            <text class="workflow-form-revision-dialog__desc">
              记录修改原因，并按需通知流程内用户
            </text>
          </view>
          <u-button custom-class="app-icon-button" :disabled="submitting" @click="notificationVisible = false">
            <u-icon name="close" size="16px" color="#5f6b7a" />
          </u-button>
        </view>

        <text class="workflow-form-revision-dialog__label">
          修改原因
        </text>
        <WorkflowTextarea v-model="reason" :disabled="submitting" :maxlength="500" :min-rows="4" :max-rows="8" count placeholder="说明本次修改原因" />

        <view class="workflow-form-revision-dialog__switch-row">
          <view>
            <text class="workflow-form-revision-dialog__label">
              发送通知
            </text>
            <text class="workflow-form-revision-dialog__hint">
              可暂不通知，默认选择当前待处理人
            </text>
          </view>
          <u-switch v-model="notifyEnabled" :disabled="submitting || notificationGroups.length === 0" />
        </view>

        <template v-if="notifyEnabled">
          <text class="workflow-form-revision-dialog__label">
            通知对象
          </text>
          <WorkflowParticipantSelect
            v-model="notificationUserIds"
            :groups="notificationGroups"
            :disabled="submitting"
            placeholder="请选择流程内用户"
          />
          <text class="workflow-form-revision-dialog__label workflow-form-revision-dialog__label--channels">
            通知渠道
          </text>
          <u-checkbox-group v-model="notificationChannels" class="workflow-form-revision-dialog__channels">
            <u-checkbox value="in_app" label="站内信" :disabled="submitting" />
            <u-checkbox value="dingtalk_oa" label="钉钉 OA" :disabled="submitting" />
          </u-checkbox-group>
        </template>

        <view class="workflow-form-revision-dialog__actions">
          <u-button plain :disabled="submitting" @click="notificationVisible = false">
            取消
          </u-button>
          <u-button type="primary" :loading="submitting" @click="submitRevision">
            确认提交
          </u-button>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<style lang="scss" scoped>
.workflow-form-revision-page {
  width: 100%;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f6f8fb;
  color: #1f2329;
}

.workflow-form-revision-page__header {
  min-height: 72px;
  padding: 14px 24px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  box-sizing: border-box;
  background: #fff;
}

.workflow-form-revision-page__heading,
.workflow-form-revision-dialog__header > view:first-child {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.workflow-form-revision-page__title,
.workflow-form-revision-dialog__title {
  color: #1f2329;
  font-size: 18px;
  line-height: 1.35;
  font-weight: 700;
}

.workflow-form-revision-page__meta,
.workflow-form-revision-page__section-desc,
.workflow-form-revision-dialog__desc,
.workflow-form-revision-dialog__hint {
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.workflow-form-revision-page__body {
  flex: 1;
  min-height: 0;
  width: min(var(--app-pc-content-max-width, 1080px), 100%);
  margin: 0 auto;
  padding: 20px 24px 32px;
  box-sizing: border-box;
}

.workflow-form-revision-page__section-head {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.workflow-form-revision-page__section-title {
  font-size: 16px;
  font-weight: 700;
}

.workflow-form-revision-page__content {
  width: 100%;
  padding: 20px;
  border: 1px solid #dfe5ee;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(31, 35, 41, 0.05);
  box-sizing: border-box;
}

.workflow-form-revision-page__form {
  width: 100%;
}

.workflow-form-revision-page__actions {
  flex-shrink: 0;
  padding: 12px 24px;
  border-top: 1px solid #e5eaf3;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background: #fff;
}

.workflow-form-revision-page__actions :deep(.u-btn) {
  width: auto;
  min-width: 96px;
  margin: 0;
}

.workflow-form-revision-page__state {
  flex: 1;
  min-height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #86909c;
}

.workflow-form-revision-page__state--error {
  flex-direction: column;
  color: #4e5969;
}

.workflow-form-revision-dialog {
  width: min(540px, calc(100vw - 32px));
  max-height: calc(100vh - 64px);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
  box-sizing: border-box;
  background: #fff;
}

.workflow-form-revision-dialog__header,
.workflow-form-revision-dialog__switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.workflow-form-revision-dialog__switch-row {
  padding-top: 4px;
}

.workflow-form-revision-dialog__switch-row > view:first-child {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.workflow-form-revision-dialog__label {
  color: #1f2329;
  font-size: 13px;
  font-weight: 600;
}

.workflow-form-revision-dialog__label--channels {
  margin-top: 4px;
}

.workflow-form-revision-dialog__channels {
  display: flex;
  gap: 18px;
}

.workflow-form-revision-dialog__actions {
  padding-top: 6px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.workflow-form-revision-dialog__actions :deep(.u-btn) {
  width: auto;
  min-width: 88px;
  margin: 0;
}

@media (max-width: 768px) {
  .workflow-form-revision-page__header,
  .workflow-form-revision-page__actions {
    padding-left: 16px;
    padding-right: 16px;
  }

  .workflow-form-revision-page__body {
    padding: 16px 12px;
  }

  .workflow-form-revision-page__content {
    padding: 12px;
  }

  .workflow-form-revision-page__section-head {
    margin-bottom: 16px;
    padding-bottom: 12px;
  }

  .workflow-form-revision-page__actions :deep(.u-btn) {
    flex: 1;
  }
}
</style>
