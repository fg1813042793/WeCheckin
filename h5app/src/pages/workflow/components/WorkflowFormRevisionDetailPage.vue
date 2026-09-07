<script setup lang="ts">
import type {
  WorkflowAttachment,
  WorkflowFormField,
  WorkflowFormRevisionDetail,
  WorkflowFormRevisionTaskSummary,
  WorkflowInstanceDetail,
} from '@/types/workflow'
import { computed, onMounted, ref } from 'vue'
import {
  cancelWorkflowFormRevision,
  completeWorkflowFormRevisionTask,
  getWorkflowFormRevision,
  getWorkflowInstance,
} from '@/api/workflow'
import { useAppContentStore, useDingtalkAuthStore } from '@/stores'
import { workflowDataFields } from '../workflow-form'
import {
  workflowFormRevisionDetailIdFromContentKey,
  workflowInstanceContentKey,
} from '../workflow-route-keys'
import { workflowFormRevisionStatusMeta, workflowTaskStatusMeta } from '../workflow-status'
import WorkflowImagePicker from './WorkflowImagePicker.vue'
import WorkflowTextarea from './WorkflowTextarea.vue'

type RevisionAction = 'approve' | 'reject'

const props = defineProps<{
  contentKey: string
}>()

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const revision = ref<WorkflowFormRevisionDetail | null>(null)
const instance = ref<WorkflowInstanceDetail | null>(null)
const loading = ref(true)
const loadError = ref('')
const submitting = ref(false)
const cancelling = ref(false)
const actionVisible = ref(false)
const action = ref<RevisionAction>('approve')
const comment = ref('')
const images = ref<WorkflowAttachment[]>([])
const uploading = ref(false)

const revisionId = computed(() => workflowFormRevisionDetailIdFromContentKey(props.contentKey))
const currentUserId = computed(() => String(auth.user?.workflowActorId || auth.user?.id || '').trim())
const currentTask = computed(() => (
  revision.value?.tasks.find(task => task.status === 'pending' && task.assigneeId === currentUserId.value) || null
))
const canHandle = computed(() => Boolean(
  revision.value?.status === 'pending'
  && currentTask.value
  && auth.hasButtonPermission('dingtalk_h5:button:workflow:form-revision-handle')
  && auth.hasApiPermission('dingtalk_h5:api:workflow:form-revision-handle'),
))
const canCancel = computed(() => Boolean(
  revision.value?.status === 'pending'
  && revision.value.requesterId === currentUserId.value
  && revision.value.tasks.every(task => !task.handledAt && !task.action)
  && auth.hasButtonPermission('dingtalk_h5:button:workflow:form-revision-create')
  && auth.hasApiPermission('dingtalk_h5:api:workflow:form-revision-create'),
))
const title = computed(() => (
  instance.value?.instance.instanceTitle
  || instance.value?.instance.definitionName
  || revision.value?.sourceNodeName
  || '表单修订详情'
))
const fieldDiffRows = computed(() => {
  const fields = workflowDataFields(instance.value?.form || [])
  const fieldMap = new Map(fields.map(field => [field.key, field]))
  return Object.keys(revision.value?.patch || {}).map(key => ({
    key,
    field: fieldMap.get(key),
    label: fieldMap.get(key)?.label || key,
    before: revision.value?.beforeFormData?.[key],
    after: revision.value?.proposedFormData?.[key],
  }))
})

onMounted(() => {
  void loadPage()
})

async function loadPage() {
  if (!revisionId.value) {
    loadError.value = '表单修订请求无效'
    loading.value = false
    return
  }
  loading.value = true
  loadError.value = ''
  try {
    const response = await getWorkflowFormRevision(revisionId.value)
    if (!response?.data)
      throw new Error('表单修订记录不存在')
    revision.value = response.data
    const instanceResponse = await getWorkflowInstance(response.data.sourceInstanceId)
    instance.value = instanceResponse?.data || null
  }
  catch (error) {
    loadError.value = requestErrorMessage(error, '表单修订详情加载失败')
  }
  finally {
    loading.value = false
  }
}

function requestClose() {
  appContent.requestCloseTab(props.contentKey)
}

function openOriginalInstance() {
  const current = revision.value
  if (!current)
    return
  const key = workflowInstanceContentKey(current.sourceInstanceId)
  if (!key)
    return
  appContent.openDynamicTab({
    key,
    label: `流程详情 · ${title.value}`,
    icon: 'eye',
    path: `/pages/index/index?view=${encodeURIComponent(key)}`,
  })
}

function openAction(nextAction: RevisionAction) {
  if (!canHandle.value || submitting.value)
    return
  action.value = nextAction
  comment.value = ''
  images.value = []
  actionVisible.value = true
}

async function submitAction() {
  const task = currentTask.value
  const normalizedComment = comment.value.trim()
  if (!task || !canHandle.value || submitting.value)
    return
  if (uploading.value) {
    uni.showToast({ title: '请等待图片上传完成', icon: 'none' })
    return
  }
  if (action.value === 'reject' && !normalizedComment) {
    uni.showToast({ title: '请填写驳回原因', icon: 'none' })
    return
  }
  if (Array.from(normalizedComment).length > 500) {
    uni.showToast({ title: '处理意见不能超过500个字符', icon: 'none' })
    return
  }

  submitting.value = true
  try {
    const response = await completeWorkflowFormRevisionTask(task.id, {
      action: action.value,
      comment: normalizedComment,
      images: images.value,
    })
    if (!response?.data)
      throw new Error('修订处理结果为空')
    revision.value = response.data
    actionVisible.value = false
    appContent.requestRefresh()
    uni.showToast({ title: action.value === 'approve' ? '已通过修订' : '已驳回修订', icon: 'success' })
  }
  catch (error) {
    uni.showToast({ title: requestErrorMessage(error, '修订处理失败'), icon: 'none' })
  }
  finally {
    submitting.value = false
  }
}

async function cancelRevision() {
  if (!revision.value || !canCancel.value || cancelling.value)
    return
  if (!await confirmCancel())
    return
  cancelling.value = true
  try {
    const response = await cancelWorkflowFormRevision(revision.value.id)
    if (!response?.data)
      throw new Error('取消修订结果为空')
    revision.value = response.data
    appContent.requestRefresh()
    uni.showToast({ title: '修订申请已取消', icon: 'success' })
  }
  catch (error) {
    uni.showToast({ title: requestErrorMessage(error, '取消修订失败'), icon: 'none' })
  }
  finally {
    cancelling.value = false
  }
}

function confirmCancel() {
  return new Promise<boolean>((resolve) => {
    uni.showModal({
      title: '取消修订',
      content: '取消后待确认任务将一并终止，确认继续吗？',
      confirmText: '确认取消',
      success: result => resolve(Boolean(result.confirm)),
      fail: () => resolve(false),
    })
  })
}

function taskUserDisplay(task: WorkflowFormRevisionTaskSummary) {
  const userId = task.handledBy || task.assigneeId
  return task.assigneeName || instance.value?.userNames?.[userId] || userId || '未知用户'
}

function requesterDisplay() {
  const requesterId = revision.value?.requesterId || ''
  return instance.value?.userNames?.[requesterId] || requesterId || '未知用户'
}

function displayValue(value: unknown, field?: WorkflowFormField) {
  if (value === undefined || value === null || value === '')
    return '暂无填写'
  if (typeof value === 'boolean')
    return value ? '是' : '否'
  if (typeof value === 'string' || typeof value === 'number')
    return String(value)
  if (field?.type === 'attachment' && Array.isArray(value)) {
    return value.map((item) => {
      if (item && typeof item === 'object' && !Array.isArray(item))
        return String((item as Record<string, unknown>).name || (item as Record<string, unknown>).url || '')
      return String(item || '')
    }).filter(Boolean).join('、') || '暂无填写'
  }
  try {
    return JSON.stringify(value)
  }
  catch {
    return String(value)
  }
}

function formatTime(timestamp?: number) {
  if (!timestamp)
    return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
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
  <view class="workflow-revision-detail-page app-pc-control-scope">
    <view class="workflow-revision-detail-page__header">
      <view class="workflow-revision-detail-page__heading">
        <view class="workflow-revision-detail-page__title-line">
          <text class="workflow-revision-detail-page__title">
            {{ title }}
          </text>
          <u-tag
            v-if="revision"
            :text="workflowFormRevisionStatusMeta(revision.status).label"
            :type="workflowFormRevisionStatusMeta(revision.status).type"
            size="mini"
          />
        </view>
        <text v-if="revision" class="workflow-revision-detail-page__meta">
          {{ revision.sourceNodeName || revision.sourceNodeId }} · {{ requesterDisplay() }} · {{ formatTime(revision.createdAt) }}
        </text>
      </view>
      <view class="workflow-revision-detail-page__header-actions">
        <u-button v-if="revision" size="mini" plain @click="openOriginalInstance">
          <u-icon name="eye" size="14px" color="#2979ff" />
          <text>原流程</text>
        </u-button>
        <u-button custom-class="app-icon-button" @click="requestClose">
          <u-icon name="close" size="16px" color="#5f6b7a" />
        </u-button>
      </view>
    </view>

    <view v-if="loading" class="workflow-revision-detail-page__state">
      <u-loading mode="circle" size="24px" />
      <text>正在加载修订详情...</text>
    </view>
    <view v-else-if="loadError" class="workflow-revision-detail-page__state workflow-revision-detail-page__state--error">
      <u-icon name="info-circle" size="28px" color="#f56c6c" />
      <text>{{ loadError }}</text>
    </view>

    <scroll-view v-else-if="revision" scroll-y class="workflow-revision-detail-page__body">
      <view class="workflow-revision-detail-page__content">
        <view class="workflow-revision-detail-page__meta-grid">
          <view><text>来源节点</text><text>{{ revision.sourceNodeName || revision.sourceNodeId }}</text></view>
          <view><text>生效方式</text><text>{{ revision.mode === 'direct' ? '直接生效' : '下游确认' }}</text></view>
          <view><text>基准版本</text><text>v{{ revision.baseFormRevision }}</text></view>
          <view><text>生效版本</text><text>{{ revision.appliedFormRevision ? `v${revision.appliedFormRevision}` : '待生效' }}</text></view>
        </view>

        <view class="workflow-revision-detail-page__section">
          <text class="workflow-revision-detail-page__section-title">
            修改原因
          </text>
          <text class="workflow-revision-detail-page__reason">
            {{ revision.reason || '暂无填写' }}
          </text>
        </view>

        <view class="workflow-revision-detail-page__section">
          <text class="workflow-revision-detail-page__section-title">
            字段差异
          </text>
          <view v-if="fieldDiffRows.length === 0" class="workflow-revision-detail-page__empty">
            暂无字段差异
          </view>
          <view v-for="row in fieldDiffRows" :key="row.key" class="workflow-revision-detail-page__diff">
            <text class="workflow-revision-detail-page__diff-label">
              {{ row.label }}
            </text>
            <view class="workflow-revision-detail-page__diff-values">
              <view>
                <text class="workflow-revision-detail-page__diff-caption">
                  修改前
                </text>
                <text class="workflow-revision-detail-page__diff-value">
                  {{ displayValue(row.before, row.field) }}
                </text>
              </view>
              <u-icon name="arrow-right" size="16px" color="#86909c" />
              <view>
                <text class="workflow-revision-detail-page__diff-caption">
                  修改后
                </text>
                <text class="workflow-revision-detail-page__diff-value">
                  {{ displayValue(row.after, row.field) }}
                </text>
              </view>
            </view>
          </view>
        </view>

        <view class="workflow-revision-detail-page__section">
          <text class="workflow-revision-detail-page__section-title">
            确认记录
          </text>
          <view v-if="revision.tasks.length === 0" class="workflow-revision-detail-page__empty">
            直接生效，无需确认
          </view>
          <view v-else class="workflow-revision-detail-page__timeline">
            <view v-for="task in revision.tasks" :key="task.id" class="workflow-revision-detail-page__task">
              <view class="workflow-revision-detail-page__task-marker" :class="`is-${task.status}`" />
              <view class="workflow-revision-detail-page__task-content">
                <view class="workflow-revision-detail-page__task-heading">
                  <text>{{ task.nodeName || task.nodeId }}</text>
                  <u-tag :text="workflowTaskStatusMeta(task.status).label" :type="workflowTaskStatusMeta(task.status).type" size="mini" plain />
                </view>
                <text>处理人：{{ taskUserDisplay(task) }}</text>
                <text v-if="task.comment">
                  处理意见：{{ task.comment }}
                </text>
                <text v-if="task.handledAt">
                  处理时间：{{ formatTime(task.handledAt) }}
                </text>
                <WorkflowImagePicker v-if="task.images?.length" :model-value="task.images" readonly />
              </view>
            </view>
          </view>
        </view>
      </view>
    </scroll-view>

    <view v-if="revision && (canCancel || canHandle)" class="workflow-revision-detail-page__actions">
      <u-button v-if="canCancel" type="error" plain :loading="cancelling" :disabled="submitting" @click="cancelRevision">
        取消修订
      </u-button>
      <template v-if="canHandle">
        <u-button type="error" plain :disabled="submitting || cancelling" @click="openAction('reject')">
          驳回
        </u-button>
        <u-button type="primary" :disabled="submitting || cancelling" @click="openAction('approve')">
          通过
        </u-button>
      </template>
    </view>

    <u-popup
      v-model="actionVisible"
      mode="center"
      width="520px"
      custom-class="workflow-revision-action-popup app-pc-control-scope"
      :z-index="10140"
      :border-radius="8"
      :mask-close-able="!submitting"
    >
      <view class="workflow-revision-action-dialog">
        <view class="workflow-revision-action-dialog__header">
          <view>
            <text class="workflow-revision-action-dialog__title">
              {{ action === 'approve' ? '通过表单修订' : '驳回表单修订' }}
            </text>
            <text class="workflow-revision-action-dialog__desc">
              {{ action === 'approve' ? '处理意见可选' : '请填写驳回原因' }}
            </text>
          </view>
          <u-button custom-class="app-icon-button" :disabled="submitting" @click="actionVisible = false">
            <u-icon name="close" size="16px" color="#5f6b7a" />
          </u-button>
        </view>
        <WorkflowTextarea
          v-model="comment"
          :disabled="submitting"
          :maxlength="500"
          :min-rows="4"
          :max-rows="8"
          count
          :placeholder="action === 'approve' ? '填写处理意见（可选）' : '填写驳回原因'"
        />
        <WorkflowImagePicker v-model="images" :disabled="submitting" @uploading-change="uploading = $event" />
        <view class="workflow-revision-action-dialog__actions">
          <u-button plain :disabled="submitting" @click="actionVisible = false">
            取消
          </u-button>
          <u-button :type="action === 'approve' ? 'primary' : 'error'" :loading="submitting" @click="submitAction">
            确认{{ action === 'approve' ? '通过' : '驳回' }}
          </u-button>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<style lang="scss" scoped>
.workflow-revision-detail-page {
  width: 100%;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f6f8fb;
  color: #1f2329;
}

.workflow-revision-detail-page__header {
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

.workflow-revision-detail-page__heading,
.workflow-revision-action-dialog__header > view:first-child {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.workflow-revision-detail-page__title-line,
.workflow-revision-detail-page__header-actions,
.workflow-revision-detail-page__task-heading,
.workflow-revision-action-dialog__header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.workflow-revision-detail-page__title,
.workflow-revision-action-dialog__title {
  font-size: 18px;
  line-height: 1.35;
  font-weight: 700;
}

.workflow-revision-detail-page__meta,
.workflow-revision-action-dialog__desc {
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.workflow-revision-detail-page__body {
  flex: 1;
  min-height: 0;
  width: min(var(--app-pc-content-max-width, 1080px), 100%);
  margin: 0 auto;
  padding: 20px 24px 32px;
  box-sizing: border-box;
}

.workflow-revision-detail-page__content {
  padding: 20px;
  border: 1px solid #dfe5ee;
  border-radius: 6px;
  background: #fff;
}

.workflow-revision-detail-page__meta-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border-top: 1px solid #e5eaf3;
  border-left: 1px solid #e5eaf3;
}

.workflow-revision-detail-page__meta-grid > view {
  min-width: 0;
  border-right: 1px solid #e5eaf3;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  flex-direction: column;
}

.workflow-revision-detail-page__meta-grid text {
  min-height: 36px;
  padding: 8px 10px;
  overflow-wrap: anywhere;
  box-sizing: border-box;
}

.workflow-revision-detail-page__meta-grid text:first-child {
  background: #f7f8fa;
  color: #86909c;
  font-size: 12px;
}

.workflow-revision-detail-page__section {
  margin-top: 24px;
}

.workflow-revision-detail-page__section-title {
  display: block;
  margin-bottom: 10px;
  font-size: 15px;
  font-weight: 700;
}

.workflow-revision-detail-page__reason,
.workflow-revision-detail-page__empty {
  display: block;
  padding: 12px;
  background: #f7f8fa;
  color: #4e5969;
  line-height: 1.6;
  white-space: pre-wrap;
}

.workflow-revision-detail-page__diff {
  border: 1px solid #e5eaf3;
}

.workflow-revision-detail-page__diff + .workflow-revision-detail-page__diff {
  margin-top: 10px;
}

.workflow-revision-detail-page__diff-label {
  display: block;
  padding: 8px 12px;
  border-bottom: 1px solid #e5eaf3;
  background: #f7f8fa;
  font-weight: 600;
}

.workflow-revision-detail-page__diff-values {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 24px minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  padding: 12px;
}

.workflow-revision-detail-page__diff-values > view {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.workflow-revision-detail-page__diff-caption {
  color: #86909c;
  font-size: 12px;
}

.workflow-revision-detail-page__diff-value {
  min-height: 22px;
  overflow-wrap: anywhere;
  line-height: 1.6;
  white-space: pre-wrap;
}

.workflow-revision-detail-page__timeline {
  padding-left: 8px;
}

.workflow-revision-detail-page__task {
  position: relative;
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  gap: 10px;
  padding-bottom: 18px;
}

.workflow-revision-detail-page__task::before {
  position: absolute;
  top: 12px;
  bottom: 0;
  left: 5px;
  width: 1px;
  background: #dfe5ee;
  content: '';
}

.workflow-revision-detail-page__task:last-child::before {
  display: none;
}

.workflow-revision-detail-page__task-marker {
  z-index: 1;
  width: 10px;
  height: 10px;
  margin-top: 4px;
  border: 2px solid #c9cdd4;
  border-radius: 50%;
  background: #fff;
}

.workflow-revision-detail-page__task-marker.is-pending { border-color: #d97706; }
.workflow-revision-detail-page__task-marker.is-approved { border-color: #059669; }
.workflow-revision-detail-page__task-marker.is-rejected { border-color: #dc2626; }

.workflow-revision-detail-page__task-content {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
  color: #4e5969;
  font-size: 13px;
}

.workflow-revision-detail-page__task-heading {
  justify-content: space-between;
  color: #1f2329;
  font-weight: 600;
}

.workflow-revision-detail-page__actions {
  flex-shrink: 0;
  padding: 12px 24px;
  border-top: 1px solid #e5eaf3;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background: #fff;
}

.workflow-revision-detail-page__actions :deep(.u-btn),
.workflow-revision-action-dialog__actions :deep(.u-btn) {
  width: auto;
  min-width: 96px;
  margin: 0;
}

.workflow-revision-detail-page__state {
  flex: 1;
  min-height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #86909c;
}

.workflow-revision-detail-page__state--error {
  flex-direction: column;
  color: #4e5969;
}

.workflow-revision-action-dialog {
  width: min(520px, calc(100vw - 32px));
  max-height: calc(100vh - 64px);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
  box-sizing: border-box;
  background: #fff;
}

.workflow-revision-action-dialog__header {
  justify-content: space-between;
}

.workflow-revision-action-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 768px) {
  .workflow-revision-detail-page__header,
  .workflow-revision-detail-page__actions {
    padding-left: 16px;
    padding-right: 16px;
  }

  .workflow-revision-detail-page__body {
    padding: 16px 12px;
  }

  .workflow-revision-detail-page__content {
    padding: 14px 12px;
  }

  .workflow-revision-detail-page__meta-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workflow-revision-detail-page__diff-values {
    grid-template-columns: 1fr;
  }

  .workflow-revision-detail-page__diff-values > :deep(.u-icon) {
    transform: rotate(90deg);
    justify-self: center;
  }

  .workflow-revision-detail-page__actions :deep(.u-btn) {
    flex: 1;
  }
}
</style>
