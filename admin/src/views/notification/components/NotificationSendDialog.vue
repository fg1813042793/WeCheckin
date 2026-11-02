<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { AdminDialog } from '@/components/admin-ui'
import WorkflowUserTreePicker from '../../workflow/components/WorkflowUserTreePicker.vue'
import SendTemplateSelect from './SendTemplateSelect.vue'
import { useNotificationSend, type NotificationSendChannel } from '../useNotificationSend'

const props = defineProps<{ modelValue: boolean; channel: NotificationSendChannel; canSend: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; sent: [channel: NotificationSendChannel] }>()
const emptyUserSelection = ref<number[]>([])
const { form, templates, sending, sendError, errors, recipientOptions, recipientOptionsLoading, recipientError, canSubmit, start, dispose, loadRecipientOptions, submit } = useNotificationSend(
  () => props.channel, () => props.canSend,
  async () => {
    try {
      await ElMessageBox.confirm('本次标题或正文已修改，切换模板会替换这些内容。是否继续？', '替换本次内容', { type: 'warning', confirmButtonText: '替换', cancelButtonText: '保留当前内容' })
      return true
    } catch { return false }
  },
)
const { options, selected, loading, switching, error, selectionError, edited, loadOptions, selectTemplate } = templates
const scopeOptions = [{ label: '全部用户', value: 'all' }, { label: '指定部门', value: 'departments' }, { label: '指定用户', value: 'users' }]
const departmentModelValue = computed({ get: () => form.departmentIds, set: value => { form.departmentIds = value } })
const formDisabled = computed(() => sending.value || switching.value || !props.canSend)
watch(() => props.modelValue, visible => {
  if (visible) { emptyUserSelection.value = []; void start() }
  else dispose()
}, { immediate: true })
onBeforeUnmount(dispose)
function close(value: boolean) { if (!sending.value && !switching.value) emit('update:modelValue', value) }
async function sendNotification() {
  const result = await submit()
  if (!result) return
  if (props.channel === 'dingtalk') {
    if (result.failedCount || result.skippedCount) ElMessage.warning(`钉钉通知发送完成：成功 ${result.sentCount} 人，未绑定或未启用 ${result.skippedCount} 人，失败 ${result.failedCount} 人`)
    else ElMessage.success(`钉钉通知已发送给 ${result.sentCount} 人`)
    if (result.failedCount > 0 && result.sentCount === 0) return
  } else ElMessage.success(`站内信已发送给 ${result.sentCount} 人`)
  emit('update:modelValue', false)
  emit('sent', props.channel)
}
</script>

<template>
  <AdminDialog :model-value="modelValue" :title="channel === 'dingtalk' ? '发送钉钉通知' : '发送站内信'" width="md"
    :show-close="!sending && !switching" :close-on-press-escape="!sending && !switching" @update:model-value="close">
    <el-alert v-if="!canSend" title="没有发送通知的权限" type="warning" :closable="false" />
    <el-alert v-if="sendError" :title="sendError" type="error" :closable="false" class="send-error" />
    <el-form label-position="top" :disabled="formDisabled" @submit.prevent="sendNotification">
      <SendTemplateSelect v-if="channel === 'dingtalk'" :options="options" :selected="selected" :loading="loading" :switching="switching"
        :error="error" :selection-error="selectionError" :edited="edited" :disabled="formDisabled" @select="selectTemplate" @retry="loadOptions" />
      <el-form-item :label="selected ? '本次通知标题' : '通知标题'" required :error="errors.title">
        <el-input v-model="form.title" maxlength="255" show-word-limit placeholder="请输入通知标题" />
      </el-form-item>
      <el-form-item :label="selected ? '本次通知正文' : '通知正文'" required :error="errors.content">
        <el-input v-model="form.content" type="textarea" :rows="6" maxlength="5000" show-word-limit resize="vertical" placeholder="请输入通知正文" />
      </el-form-item>
      <el-form-item label="收件范围" required><el-segmented v-model="form.scope" :options="scopeOptions" block /></el-form-item>
      <div v-if="recipientError" class="recipient-error" role="alert">
        {{ recipientError }} <el-button link type="primary" :disabled="recipientOptionsLoading" @click="loadRecipientOptions">重试</el-button>
      </div>
      <el-form-item v-if="form.scope === 'departments'" label="指定部门" required :error="errors.departmentIds">
        <WorkflowUserTreePicker v-model="emptyUserSelection" v-model:department-model-value="departmentModelValue"
          :departments="recipientOptions.departments" :users="[]" :multiple="true" :select-department-rules="true"
          :disabled="formDisabled || recipientOptionsLoading || Boolean(recipientError)" placeholder="请选择一个或多个部门" />
      </el-form-item>
      <el-form-item v-if="form.scope === 'users'" label="指定用户" required :error="errors.userIds">
        <WorkflowUserTreePicker v-model="form.userIds" :departments="recipientOptions.departments" :users="recipientOptions.users" :multiple="true"
          :disabled="formDisabled || recipientOptionsLoading || Boolean(recipientError)" placeholder="请选择一个或多个用户" />
      </el-form-item>
      <p v-if="recipientOptionsLoading" class="recipient-loading" role="status">正在加载收件范围…</p>
    </el-form>
    <template #footer>
      <el-button :disabled="sending || switching" @click="close(false)">取消</el-button>
      <el-button type="primary" :loading="sending" :disabled="!canSubmit" @click="sendNotification">发送</el-button>
    </template>
  </AdminDialog>
</template>

<style scoped>
.send-error { margin-bottom: var(--admin-ui-space-4); }
.recipient-error { color: var(--admin-ui-color-danger); margin-bottom: var(--admin-ui-space-3); }
.recipient-loading { color: var(--admin-ui-color-text-secondary); }
</style>
