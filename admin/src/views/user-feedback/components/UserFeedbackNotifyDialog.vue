<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { AdminDialog } from '@/components/admin-ui'
import { adminApi } from '@/api'
import type { UserFeedbackDetail, UserFeedbackNotificationChannel } from '@/types/userFeedback'
import { showRequestError } from '@/utils/request'
import { canHandleUserFeedback, userFeedbackStatusMeta } from '../userFeedbackStatus'

const emit = defineEmits<{ success: [detail: UserFeedbackDetail] }>()
const visible = ref(false)
const submitting = ref(false)
const detail = ref<UserFeedbackDetail | null>(null)
const content = ref('')
const channel = ref<UserFeedbackNotificationChannel>('internal')
const requestId = ref('')
const contentError = ref('')
const submitError = ref('')
let attemptedPayload: string | null = null

const confirmDisabled = computed(() => submitting.value || !content.value.trim() || Array.from(content.value.trim()).length > 5000)
const recipient = computed(() => detail.value?.submitterName
  ? `${detail.value.submitterName} (#${detail.value.submitterId})`
  : `用户 #${detail.value?.submitterId}`)

function newRequestId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  return `feedback-notify-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function open(feedback: UserFeedbackDetail) {
  if (!canHandleUserFeedback()) {
    ElMessage.warning('当前账号没有反馈处理权限')
    return
  }
  if (submitting.value || visible.value) return
  detail.value = feedback
  content.value = ''
  channel.value = 'internal'
  requestId.value = newRequestId()
  attemptedPayload = null
  contentError.value = ''
  submitError.value = ''
  visible.value = true
}

function handleVisibleChange(value: boolean) {
  if (submitting.value) return
  visible.value = value
}

function beforeClose(done: () => void) {
  if (!submitting.value) done()
}

async function submit() {
  if (!canHandleUserFeedback()) {
    ElMessage.warning('当前账号没有反馈处理权限')
    return
  }
  if (submitting.value || !visible.value || !detail.value) return
  const text = content.value.trim()
  contentError.value = !text ? '请填写通知正文' : Array.from(text).length > 5000 ? '通知正文不能超过 5000 字' : ''
  if (contentError.value) return
  if (channel.value !== 'internal' && channel.value !== 'dingtalk') {
    submitError.value = '请选择通知渠道'
    return
  }
  // 相同内容与渠道重试复用幂等键；编辑为另一条通知时不能重放旧结果。
  const payload = JSON.stringify([text, channel.value])
  if (attemptedPayload !== null && attemptedPayload !== payload) requestId.value = newRequestId()
  attemptedPayload = payload
  submitting.value = true
  submitError.value = ''
  try {
    const response = await adminApi.userFeedbackNotify(detail.value.id, {
      content: text,
      channel: channel.value,
      version: detail.value.version,
      requestId: requestId.value,
    })
    visible.value = false
    ElMessage.success('通知已提交发送')
    emit('success', response.data)
  } catch (error) {
    if (error && typeof error === 'object' && 'msg' in error && error.msg === '反馈已更新，请刷新后重试') {
      submitError.value = '反馈已更新，请关闭弹窗并刷新详情后再通知；当前填写内容已保留。'
    } else {
      submitError.value = error && typeof error === 'object' && 'msg' in error && typeof error.msg === 'string'
        ? error.msg : '通知提交失败，请重试'
      showRequestError(error, '通知提交失败')
    }
  } finally {
    submitting.value = false
  }
}

defineExpose({ open, visible })
</script>

<template>
  <AdminDialog
    :model-value="visible"
    title="通知提交人"
    width="sm"
    :confirm-loading="submitting"
    :confirm-disabled="confirmDisabled"
    confirm-text="确认发送"
    append-to-body
    :destroy-on-close="false"
    :before-close="beforeClose"
    :show-close="!submitting"
    :close-on-press-escape="!submitting"
    :close-on-click-modal="false"
    @update:model-value="handleVisibleChange"
    @confirm="submit"
  >
    <el-alert v-if="submitError" :title="submitError" type="warning" show-icon :closable="false" />
    <el-form v-if="detail" class="feedback-notify-form" label-position="top" :disabled="submitting" @submit.prevent="submit">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="反馈编号">{{ detail.feedbackNo }}</el-descriptions-item>
        <el-descriptions-item label="接收人">{{ recipient }}</el-descriptions-item>
        <el-descriptions-item label="当前状态">
          <el-tag :type="userFeedbackStatusMeta(detail.status).type">{{ userFeedbackStatusMeta(detail.status).label }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>
      <el-form-item label="通知渠道" required>
        <el-radio-group v-model="channel" aria-label="通知渠道">
          <el-radio value="internal">站内通知</el-radio>
          <el-radio value="dingtalk">钉钉通知</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="通知正文" required :error="contentError">
        <el-input
          v-model="content"
          aria-label="通知正文"
          type="textarea"
          :rows="6"
          maxlength="5000"
          show-word-limit
          placeholder="填写给反馈提交人的通知内容"
          @input="contentError = ''"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button :disabled="submitting" @click="handleVisibleChange(false)">取消</el-button>
      <el-button type="primary" :loading="submitting" :disabled="confirmDisabled" @click="submit">确认发送</el-button>
    </template>
  </AdminDialog>
</template>

<style scoped>
.feedback-notify-form {
  margin-top: var(--admin-ui-space-4);
}

.feedback-notify-form :deep(.el-descriptions) {
  margin-bottom: var(--admin-ui-space-4);
  overflow-wrap: anywhere;
}
</style>
