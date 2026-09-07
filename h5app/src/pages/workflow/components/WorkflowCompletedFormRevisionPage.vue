<script setup lang="ts">
import type {
  WorkflowCompletedRevisionNode,
  WorkflowFieldAccessMap,
  WorkflowFieldActionsMap,
  WorkflowFormData,
  WorkflowFormField,
  WorkflowFormRevisionPreview,
  WorkflowInstanceDetail,
} from '@/types/workflow'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  createWorkflowFormRevision,
  getWorkflowInstance,
  previewWorkflowFormRevision,
} from '@/api/workflow'
import { useAppContentStore, useDingtalkAuthStore } from '@/stores'
import {
  initialWorkflowFormData,
  workflowFieldAccessMap,
  workflowFieldActionsMap,
  writableWorkflowFormData,
} from '../workflow-form'
import {
  workflowCompletedFormRevisionInstanceIdFromContentKey,
  workflowFormRevisionDetailContentKey,
} from '../workflow-route-keys'
import WorkflowRuntimeForm from './WorkflowRuntimeForm.vue'
import WorkflowTextarea from './WorkflowTextarea.vue'

interface RuntimeFormExposed {
  validate: () => { valid: boolean, errors: Record<string, string> }
}

const props = defineProps<{
  contentKey: string
}>()

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const detail = ref<WorkflowInstanceDetail | null>(null)
const formData = ref<WorkflowFormData>({})
const originalData = ref<WorkflowFormData>({})
const selectedNodeId = ref('')
const reason = ref('')
const preview = ref<WorkflowFormRevisionPreview | null>(null)
const previewVisible = ref(false)
const formRef = ref<RuntimeFormExposed | null>(null)
const loading = ref(true)
const loadError = ref('')
const previewing = ref(false)
const submitting = ref(false)
const initialSnapshot = ref('')
let unregisterCloseGuard: (() => void) | null = null

const instanceId = computed(() => workflowCompletedFormRevisionInstanceIdFromContentKey(props.contentKey))
const revisionNodes = computed(() => detail.value?.formRevision?.completedRevisionNodes || [])
const selectedNode = computed<WorkflowCompletedRevisionNode | null>(() => (
  revisionNodes.value.find(node => node.nodeId === selectedNodeId.value) || null
))
const canCreate = computed(() => Boolean(
  detail.value?.instance.status === 'completed'
  && selectedNode.value
  && auth.hasButtonPermission('dingtalk_h5:button:workflow:form-revision-create')
  && auth.hasApiPermission('dingtalk_h5:api:workflow:form-revision-create'),
))
const title = computed(() => detail.value?.instance.instanceTitle || detail.value?.instance.definitionName || '申请表单修订')
const fieldAccess = computed<WorkflowFieldAccessMap>(() => workflowFieldAccessMap(
  detail.value?.form || [],
  selectedNode.value?.fieldPermissions || [],
  'hidden',
))
const fieldActions = computed<WorkflowFieldActionsMap>(() => workflowFieldActionsMap(
  detail.value?.form || [],
  selectedNode.value?.fieldPermissions || [],
))
const hasUnsavedChanges = computed(() => Boolean(
  initialSnapshot.value
  && (currentSnapshot() !== initialSnapshot.value || reason.value.trim()),
))
const changedFieldLabels = computed(() => {
  const labels = new Map((detail.value?.form || []).flatMap(field => flattenFieldLabels(field)))
  return Object.keys(changedPatch()).map(field => labels.get(field) || field)
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

function flattenFieldLabels(field: WorkflowFormField): Array<[string, string]> {
  return [
    [field.key, field.label || field.key],
    ...(field.fields || []).flatMap(child => flattenFieldLabels(child)),
  ]
}

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
    selectedNodeId.value = revisionNodes.value[0]?.nodeId || ''
    if (!canCreate.value) {
      loadError.value = '当前流程不允许申请表单修订，可能状态或权限已变更'
      return
    }
    resetFormForNode()
  }
  catch (error) {
    loadError.value = requestErrorMessage(error, '流程表单加载失败')
  }
  finally {
    loading.value = false
  }
}

function resetFormForNode() {
  const current = detail.value
  if (!current)
    return
  const initial = initialWorkflowFormData(current.form || [], current.formData || {})
  originalData.value = initialWorkflowFormData(current.form || [], initial)
  formData.value = initialWorkflowFormData(current.form || [], initial)
  preview.value = null
  previewVisible.value = false
  initialSnapshot.value = currentSnapshot()
}

function requestClose() {
  appContent.requestCloseTab(props.contentKey)
}

async function openPreview() {
  const current = detail.value
  const normalizedReason = reason.value.trim()
  if (!current || !selectedNode.value || !canCreate.value || previewing.value || submitting.value)
    return
  const validation = formRef.value?.validate()
  if (validation && !validation.valid) {
    uni.showToast({ title: '请检查表单填写内容', icon: 'none' })
    return
  }
  const patch = changedPatch()
  if (Object.keys(patch).length === 0) {
    uni.showToast({ title: '请至少修改一个表单字段', icon: 'none' })
    return
  }
  if (!normalizedReason) {
    uni.showToast({ title: '请填写修改原因', icon: 'none' })
    return
  }
  if (Array.from(normalizedReason).length > 500) {
    uni.showToast({ title: '修改原因不能超过500个字符', icon: 'none' })
    return
  }

  previewing.value = true
  try {
    const response = await previewWorkflowFormRevision(current.instance.id, {
      sourceNodeId: selectedNode.value.nodeId,
      expectedRevision: current.formRevision.revision,
      formData: patch,
    })
    if (!response?.data)
      throw new Error('修订预览为空')
    preview.value = response.data
    previewVisible.value = true
  }
  catch (error) {
    uni.showToast({ title: requestErrorMessage(error, '修订预览失败'), icon: 'none' })
  }
  finally {
    previewing.value = false
  }
}

async function submitRevision() {
  const current = detail.value
  const node = selectedNode.value
  if (!current || !node || !preview.value || submitting.value)
    return
  submitting.value = true
  try {
    const response = await createWorkflowFormRevision(current.instance.id, {
      sourceNodeId: node.nodeId,
      expectedRevision: current.formRevision.revision,
      formData: changedPatch(),
      reason: reason.value.trim(),
    })
    if (!response?.data)
      throw new Error('修订申请创建失败')
    const key = workflowFormRevisionDetailContentKey(response.data.id)
    appContent.removeDynamicTab(props.contentKey)
    appContent.openDynamicTab({
      key,
      label: `表单修订 · ${response.data.sourceNodeName || node.nodeName}`,
      icon: 'edit-pen',
      path: `/pages/index/index?view=${encodeURIComponent(key)}`,
    })
    appContent.requestRefresh()
    uni.showToast({
      title: response.data.status === 'applied' ? '表单修订已生效' : '修订申请已提交',
      icon: 'success',
    })
  }
  catch (error) {
    previewVisible.value = false
    uni.showToast({ title: requestErrorMessage(error, '修订申请提交失败'), icon: 'none' })
  }
  finally {
    submitting.value = false
  }
}

function confirmationAssignees() {
  return preview.value?.confirmationStages
    .flatMap(stage => stage.nodes)
    .flatMap(node => node.assigneeNames)
    .filter(Boolean)
    .join('、') || '按流程节点规则确定'
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
  <view class="workflow-completed-revision-page app-pc-control-scope">
    <view class="workflow-completed-revision-page__header">
      <view class="workflow-completed-revision-page__heading">
        <text class="workflow-completed-revision-page__title">
          {{ title }}
        </text>
        <text v-if="detail" class="workflow-completed-revision-page__meta">
          业务编号：{{ detail.instance.businessKey || '-' }} · 当前表单版本 v{{ detail.formRevision.revision }}
        </text>
      </view>
      <u-tag v-if="detail" text="已完成" type="success" size="mini" />
    </view>

    <view v-if="loading" class="workflow-completed-revision-page__state">
      <u-loading mode="circle" size="24px" />
      <text>正在加载流程表单...</text>
    </view>
    <view v-else-if="loadError" class="workflow-completed-revision-page__state workflow-completed-revision-page__state--error">
      <u-icon name="info-circle" size="28px" color="#f56c6c" />
      <text>{{ loadError }}</text>
    </view>
    <scroll-view v-else-if="detail" scroll-y class="workflow-completed-revision-page__body">
      <view class="workflow-completed-revision-page__content">
        <view class="workflow-completed-revision-page__field">
          <text class="workflow-completed-revision-page__label">
            选择来源节点
          </text>
          <select
            v-model="selectedNodeId"
            class="workflow-completed-revision-page__select"
            :disabled="previewing || submitting"
            aria-label="选择来源节点"
            @change="resetFormForNode"
          >
            <option v-for="node in revisionNodes" :key="node.nodeId" :value="node.nodeId">
              {{ node.nodeName || node.nodeId }}
            </option>
          </select>
          <text class="workflow-completed-revision-page__hint">
            以你实际办理过的节点权限展示字段，每次修订只能选择一个来源节点
          </text>
        </view>

        <view class="workflow-completed-revision-page__section-head">
          <text class="workflow-completed-revision-page__section-title">
            修改字段
          </text>
          <text class="workflow-completed-revision-page__hint">
            只会提交与当前表单不同的字段
          </text>
        </view>
        <WorkflowRuntimeForm
          ref="formRef"
          v-model="formData"
          class="workflow-completed-revision-page__form app-workflow-form app-pc-control-scope"
          :fields="detail.form || []"
          :field-access="fieldAccess"
          :field-actions="fieldActions"
          :readonly="false"
          readonly-appearance="plain"
        />

        <view class="workflow-completed-revision-page__reason">
          <text class="workflow-completed-revision-page__label">
            修改原因
          </text>
          <WorkflowTextarea
            v-model="reason"
            :disabled="previewing || submitting"
            :maxlength="500"
            :min-rows="4"
            :max-rows="8"
            count
            placeholder="说明为什么需要修改已完成流程的表单"
          />
        </view>
      </view>
    </scroll-view>

    <view v-if="detail && !loading && !loadError" class="workflow-completed-revision-page__actions">
      <u-button plain :disabled="previewing || submitting" @click="requestClose">
        取消
      </u-button>
      <u-button type="primary" :loading="previewing" :disabled="!canCreate || submitting" @click="openPreview">
        预览并提交
      </u-button>
    </view>

    <u-popup
      v-model="previewVisible"
      mode="center"
      width="560px"
      custom-class="workflow-completed-revision-popup app-pc-control-scope"
      :z-index="10140"
      :border-radius="8"
      :mask-close-able="!submitting"
    >
      <view v-if="preview" class="workflow-completed-revision-dialog">
        <view class="workflow-completed-revision-dialog__header">
          <view>
            <text class="workflow-completed-revision-dialog__title">
              确认表单修订
            </text>
            <text class="workflow-completed-revision-dialog__desc">
              以服务器计算结果为准
            </text>
          </view>
          <u-button custom-class="app-icon-button" :disabled="submitting" @click="previewVisible = false">
            <u-icon name="close" size="16px" color="#5f6b7a" />
          </u-button>
        </view>

        <view
          class="workflow-completed-revision-dialog__mode"
          :class="{ 'workflow-completed-revision-dialog__mode--review': preview.mode === 'downstream_review' }"
        >
          <u-icon :name="preview.mode === 'direct' ? 'checkmark-circle' : 'clock'" size="20px" :color="preview.mode === 'direct' ? '#059669' : '#d97706'" />
          <view>
            <text class="workflow-completed-revision-dialog__mode-title">
              {{ preview.mode === 'direct' ? '提交后立即生效' : '需要重新确认' }}
            </text>
            <text class="workflow-completed-revision-dialog__desc">
              {{ preview.mode === 'direct' ? '本次修改均为该节点配置的低风险字段' : `确认人：${confirmationAssignees()}` }}
            </text>
          </view>
        </view>

        <view>
          <text class="workflow-completed-revision-dialog__label">
            变更字段
          </text>
          <view class="workflow-completed-revision-dialog__tags">
            <u-tag v-for="field in changedFieldLabels" :key="field" :text="field" type="primary" size="mini" plain />
          </view>
        </view>

        <view v-if="preview.confirmationStages.length > 0">
          <text class="workflow-completed-revision-dialog__label">
            确认节点
          </text>
          <view v-for="stage in preview.confirmationStages" :key="stage.stage" class="workflow-completed-revision-dialog__stage">
            <text>第 {{ stage.stage }} 阶段</text>
            <text>{{ stage.nodes.map(node => node.nodeName || node.nodeId).join('、') }}</text>
          </view>
        </view>

        <view class="workflow-completed-revision-dialog__actions">
          <u-button plain :disabled="submitting" @click="previewVisible = false">
            返回修改
          </u-button>
          <u-button type="primary" :loading="submitting" @click="submitRevision">
            提交修订申请
          </u-button>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<style lang="scss" scoped>
.workflow-completed-revision-page {
  width: 100%;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f6f8fb;
  color: #1f2329;
}

.workflow-completed-revision-page__header {
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

.workflow-completed-revision-page__heading,
.workflow-completed-revision-dialog__header > view:first-child {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.workflow-completed-revision-page__title,
.workflow-completed-revision-dialog__title {
  font-size: 18px;
  line-height: 1.35;
  font-weight: 700;
}

.workflow-completed-revision-page__meta,
.workflow-completed-revision-page__hint,
.workflow-completed-revision-dialog__desc {
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.workflow-completed-revision-page__body {
  flex: 1;
  min-height: 0;
  width: min(var(--app-pc-content-max-width, 1080px), 100%);
  margin: 0 auto;
  padding: 20px 24px 32px;
  box-sizing: border-box;
}

.workflow-completed-revision-page__content {
  width: 100%;
  padding: 20px;
  border: 1px solid #dfe5ee;
  border-radius: 6px;
  background: #fff;
  box-sizing: border-box;
}

.workflow-completed-revision-page__field,
.workflow-completed-revision-page__reason,
.workflow-completed-revision-page__section-head {
  display: flex;
  flex-direction: column;
  gap: 7px;
}

.workflow-completed-revision-page__field,
.workflow-completed-revision-page__section-head {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e5eaf3;
}

.workflow-completed-revision-page__reason {
  margin-top: 20px;
  padding-top: 18px;
  border-top: 1px solid #e5eaf3;
}

.workflow-completed-revision-page__label,
.workflow-completed-revision-page__section-title,
.workflow-completed-revision-dialog__label {
  font-size: 14px;
  font-weight: 600;
}

.workflow-completed-revision-page__section-title {
  font-size: 16px;
  font-weight: 700;
}

.workflow-completed-revision-page__select {
  width: min(420px, 100%);
  height: 36px;
  padding: 0 10px;
  border: 1px solid #d8e0e8;
  border-radius: 4px;
  outline: none;
  background: #fff;
  color: #1f2329;
  font-size: 13px;
}

.workflow-completed-revision-page__select:focus {
  border-color: #2979ff;
}

.workflow-completed-revision-page__actions {
  flex-shrink: 0;
  padding: 12px 24px;
  border-top: 1px solid #e5eaf3;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background: #fff;
}

.workflow-completed-revision-page__actions :deep(.u-btn),
.workflow-completed-revision-dialog__actions :deep(.u-btn) {
  width: auto;
  min-width: 104px;
  margin: 0;
}

.workflow-completed-revision-page__state {
  flex: 1;
  min-height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #86909c;
}

.workflow-completed-revision-page__state--error {
  flex-direction: column;
  color: #4e5969;
}

.workflow-completed-revision-dialog {
  width: min(560px, calc(100vw - 32px));
  max-height: calc(100vh - 64px);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  overflow-y: auto;
  box-sizing: border-box;
  background: #fff;
}

.workflow-completed-revision-dialog__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.workflow-completed-revision-dialog__mode {
  padding: 14px;
  border: 1px solid #a7f3d0;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  background: #ecfdf5;
}

.workflow-completed-revision-dialog__mode--review {
  border-color: #fde68a;
  background: #fffbeb;
}

.workflow-completed-revision-dialog__mode > view {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.workflow-completed-revision-dialog__mode-title {
  font-size: 14px;
  font-weight: 700;
}

.workflow-completed-revision-dialog__tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.workflow-completed-revision-dialog__stage {
  margin-top: 8px;
  padding: 9px 10px;
  border-left: 3px solid #d97706;
  display: flex;
  justify-content: space-between;
  gap: 12px;
  background: #f7f8fa;
  color: #4e5969;
  font-size: 13px;
}

.workflow-completed-revision-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 768px) {
  .workflow-completed-revision-page__header,
  .workflow-completed-revision-page__actions {
    padding-left: 16px;
    padding-right: 16px;
  }

  .workflow-completed-revision-page__body {
    padding: 16px 12px;
  }

  .workflow-completed-revision-page__content {
    padding: 14px 12px;
  }

  .workflow-completed-revision-page__actions :deep(.u-btn) {
    flex: 1;
  }

  .workflow-completed-revision-dialog__stage {
    flex-direction: column;
    gap: 3px;
  }
}
</style>
