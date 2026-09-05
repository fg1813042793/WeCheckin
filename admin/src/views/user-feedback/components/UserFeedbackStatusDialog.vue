<script lang="ts">
export function requiresStatusNote(fromStatus: string, toStatus: string): boolean {
  return !(fromStatus === 'pending' && toStatus === 'processing')
}

export function isVersionConflict(error: unknown): boolean {
  return Boolean(
    error
    && typeof error === 'object'
    && 'msg' in error
    && error.msg === '反馈已更新，请刷新后重试',
  )
}
</script>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { AdminDialog } from '@/components/admin-ui'
import { adminApi } from '@/api'
import type { UserFeedbackDetail, UserFeedbackStatus } from '@/types/userFeedback'
import { showRequestError } from '@/utils/request'
import { canHandleUserFeedback, nextUserFeedbackStatuses, userFeedbackStatusMeta } from '../userFeedbackStatus'

const emit = defineEmits<{
  success: [detail: UserFeedbackDetail]
}>()

const visible = ref(false)
const submitting = ref(false)
const detail = ref<UserFeedbackDetail | null>(null)
const requestId = ref('')
const submitError = ref('')
const form = reactive<{
  status: UserFeedbackStatus | ''
  note: string
  notifyUser: boolean
}>({
  status: '',
  note: '',
  notifyUser: true,
})

const statusOptions = computed(() => detail.value ? nextUserFeedbackStatuses(detail.value.status) : [])
const noteRequired = computed(() => Boolean(detail.value && form.status && requiresStatusNote(detail.value.status, form.status)))
const confirmDisabled = computed(() => !form.status || (noteRequired.value && !form.note.trim()))

function newRequestId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  return `feedback-status-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function open(feedback: UserFeedbackDetail) {
  if (!canHandleUserFeedback()) {
    ElMessage.warning('当前账号没有反馈处理权限')
    return
  }
  const nextStatuses = nextUserFeedbackStatuses(feedback.status)
  detail.value = feedback
  form.status = nextStatuses[0] || ''
  form.note = ''
  form.notifyUser = true
  requestId.value = newRequestId()
  submitError.value = ''
  visible.value = true
}

function handleVisibleChange(value: boolean) {
  if (submitting.value && !value) return
  visible.value = value
}

async function submit() {
  if (!canHandleUserFeedback()) {
    ElMessage.warning('当前账号没有反馈处理权限')
    return
  }
  if (!detail.value || !form.status) return
  if (!nextUserFeedbackStatuses(detail.value.status).includes(form.status)) {
    submitError.value = '当前状态已不允许执行该操作，请关闭后刷新详情'
    return
  }
  if (requiresStatusNote(detail.value.status, form.status) && !form.note.trim()) {
    submitError.value = '请填写处理说明'
    return
  }

  submitting.value = true
  submitError.value = ''
  try {
    const response = await adminApi.userFeedbackUpdateStatus(detail.value.id, {
      status: form.status,
      note: form.note.trim(),
      notifyUser: form.notifyUser,
      version: detail.value.version,
      requestId: requestId.value,
    })
    visible.value = false
    ElMessage.success('反馈状态已更新')
    emit('success', response.data)
  } catch (error) {
    if (isVersionConflict(error)) {
      submitError.value = '反馈已更新，请刷新详情后再处理；当前填写内容已保留。'
    } else {
      submitError.value = error && typeof error === 'object' && 'msg' in error && typeof error.msg === 'string'
        ? error.msg
        : '反馈状态更新失败，请重试'
      showRequestError(error, '反馈状态更新失败')
    }
  } finally {
    submitting.value = false
  }
}

defineExpose({ open })
</script>

<template>
  <AdminDialog
    :model-value="visible"
    title="更新反馈状态"
    width="sm"
    :confirm-loading="submitting"
    :confirm-disabled="confirmDisabled"
    confirm-text="确认更新"
    append-to-body
    :destroy-on-close="false"
    @update:model-value="handleVisibleChange"
    @confirm="submit"
  >
    <el-alert v-if="submitError" :title="submitError" type="warning" show-icon :closable="false" />
    <el-form v-if="detail" class="feedback-status-form" label-position="top" @submit.prevent>
      <el-form-item label="目标状态" required>
        <el-select v-model="form.status" placeholder="请选择目标状态">
          <el-option
            v-for="status in statusOptions"
            :key="status"
            :label="userFeedbackStatusMeta(status).label"
            :value="status"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="处理说明" :required="noteRequired">
        <el-input
          v-model="form.note"
          type="textarea"
          :rows="5"
          maxlength="5000"
          show-word-limit
          placeholder="填写本次处理情况"
          @input="submitError = ''"
        />
      </el-form-item>
      <el-form-item label="通知用户">
        <div class="feedback-status-form__switch">
          <el-switch v-model="form.notifyUser" />
          <span>{{ form.notifyUser ? '发送站内信' : '不发送站内信' }}</span>
        </div>
      </el-form-item>
    </el-form>
  </AdminDialog>
</template>

<style scoped>
.feedback-status-form {
  margin-top: var(--admin-ui-space-4);
}

.feedback-status-form :deep(.el-select) {
  width: 100%;
}

.feedback-status-form__switch {
  display: flex;
  align-items: center;
  gap: var(--admin-ui-space-2);
  color: var(--admin-ui-color-text-secondary);
}
</style>
