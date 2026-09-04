<template>
  <el-dialog
    :model-value="modelValue"
    title="消息样式"
    width="min(1080px, 96vw)"
    append-to-body
    destroy-on-close
    :close-on-click-modal="false"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div v-loading="loading" class="style-editor">
      <nav class="style-types" aria-label="消息类型">
        <button
          v-for="style in styles"
          :key="style.type"
          type="button"
          :class="['style-type', { 'is-active': style.type === selectedType }]"
          @click="selectType(style.type)"
        >
          <span :class="['tone-dot', `tone-dot--${style.tone}`]" />
          <span class="style-type__text">
            <strong>{{ typeName(style.type) }}</strong>
            <small>{{ style.type }}</small>
          </span>
        </button>
      </nav>

      <div v-if="selectedStyle" class="style-workspace">
        <section class="style-form">
          <div class="section-heading">
            <div>
              <h3>{{ typeName(selectedStyle.type) }}</h3>
              <code>{{ selectedStyle.type }}</code>
            </div>
            <el-tag :type="tagType(selectedStyle.tone)" effect="light">{{ toneName(selectedStyle.tone) }}</el-tag>
          </div>

          <el-form label-position="top" @submit.prevent>
            <div class="form-grid">
              <el-form-item label="类型标签" required>
                <el-input v-model="selectedStyle.label" maxlength="20" show-word-limit :disabled="!canEdit" />
              </el-form-item>
              <el-form-item label="图标" required>
                <el-select v-model="selectedStyle.icon" :disabled="!canEdit" class="full-width">
                  <el-option v-for="option in iconOptions" :key="option.value" :label="option.label" :value="option.value">
                    <div class="icon-option">
                      <el-icon><component :is="option.component" /></el-icon>
                      <span>{{ option.label }}</span>
                    </div>
                  </el-option>
                </el-select>
              </el-form-item>
            </div>
            <el-form-item label="语义色" required>
              <el-radio-group v-model="selectedStyle.tone" :disabled="!canEdit" class="tone-selector">
                <el-radio-button v-for="option in toneOptions" :key="option.value" :value="option.value">
                  <span :class="['tone-dot', `tone-dot--${option.value}`]" />{{ option.label }}
                </el-radio-button>
              </el-radio-group>
            </el-form-item>
          </el-form>

          <div class="preview-heading">
            <h3>实时预览</h3>
            <el-segmented v-model="previewChannel" :options="previewChannelOptions" size="small" />
          </div>
          <div v-if="previewChannel === 'in_app'" :class="['message-preview', `message-preview--${selectedStyle.tone}`]">
            <div :class="['message-preview__icon', `message-preview__icon--${selectedStyle.tone}`]">
              <el-icon><component :is="selectedIconComponent" /></el-icon>
            </div>
            <div class="message-preview__body">
              <div class="message-preview__meta">
                <el-tag :type="tagType(selectedStyle.tone)" size="small" effect="light">{{ selectedStyle.label }}</el-tag>
                <span>刚刚</span>
              </div>
              <strong>{{ testForm.title || '测试通知标题' }}</strong>
              <p>{{ testForm.content || '测试通知正文' }}</p>
            </div>
          </div>
          <div v-else class="dingtalk-preview">
            <div class="dingtalk-preview__brand">钉钉工作通知</div>
            <strong>{{ testForm.title || '测试通知标题' }}</strong>
            <p>{{ testForm.content || '测试通知正文' }}</p>
          </div>
        </section>

        <section v-if="canSend || canSendDingTalk" class="test-section">
          <h3>发送测试</h3>
          <el-form label-position="top" @submit.prevent>
            <el-form-item label="测试标题" required>
              <el-input v-model="testForm.title" maxlength="255" show-word-limit />
            </el-form-item>
            <el-form-item label="测试正文" required>
              <el-input v-model="testForm.content" type="textarea" :rows="3" maxlength="5000" show-word-limit resize="vertical" />
            </el-form-item>
            <el-form-item label="测试收件人" required>
              <WorkflowUserTreePicker
                v-model="testForm.userIds"
                :departments="recipientOptions.departments"
                :users="recipientOptions.users"
                :multiple="true"
                :disabled="recipientOptionsLoading"
                placeholder="请选择测试收件人"
              />
            </el-form-item>
          </el-form>
          <div class="test-actions">
            <el-button
              v-if="canSend"
              type="primary"
              :icon="Promotion"
              :loading="testSending === 'in_app'"
              :disabled="Boolean(testSending)"
              @click="sendTest('in_app')"
            >保存并发送站内信测试</el-button>
            <el-button
              v-if="canSendDingTalk"
              type="primary"
              plain
              :icon="ChatDotRound"
              :loading="testSending === 'dingtalk'"
              :disabled="Boolean(testSending)"
              @click="sendTest('dingtalk')"
            >保存并发送钉钉测试</el-button>
          </div>
        </section>
      </div>
    </div>

    <template #footer>
      <el-button @click="emit('update:modelValue', false)">关闭</el-button>
      <el-button v-if="canEdit" type="primary" :icon="Check" :loading="saving" @click="saveStyles(true)">保存样式</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import {
  Bell,
  ChatDotRound,
  Check,
  CircleCheck,
  Clock,
  Document,
  InfoFilled,
  Message,
  Promotion,
  Refresh,
  Share,
  WarningFilled,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { computed, reactive, ref, watch } from 'vue'
import { adminApi } from '../../../api'
import type {
  InAppNotificationRecipientOptions,
  NotificationStyle,
  NotificationStyleConfig,
  NotificationTone,
} from '../../../api/types'
import WorkflowUserTreePicker from '../../workflow/components/WorkflowUserTreePicker.vue'

const props = defineProps<{
  modelValue: boolean
  canEdit: boolean
  canSend: boolean
  canSendDingTalk: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'sent': [channel: 'in_app' | 'dingtalk']
}>()

const typeNames: Record<string, string> = {
  task_arrived: '任务到达',
  task_reminder: '任务催办',
  approval_result_approved: '审批通过结果',
  approval_result_rejected: '审批驳回结果',
  approval_result_returned: '审批退回结果',
  node_cc: '节点抄送',
  node_notify: '节点通知',
  instance_commented: '流程评论',
  workflow: '兼容流程消息',
  admin_manual: '后台手动通知',
  scheduled_task: '定时任务通知',
  survey_stat: '问卷统计通知',
}

const iconOptions = [
  { label: '时钟', value: 'clock', component: Clock },
  { label: '铃铛', value: 'bell', component: Bell },
  { label: '完成', value: 'checkmark-circle', component: CircleCheck },
  { label: '警告', value: 'error-circle', component: WarningFilled },
  { label: '重试', value: 'reload', component: Refresh },
  { label: '分享', value: 'share', component: Share },
  { label: '邮件', value: 'email', component: Message },
  { label: '对话', value: 'chat', component: ChatDotRound },
  { label: '文档', value: 'file-text', component: Document },
  { label: '信息', value: 'info-circle', component: InfoFilled },
]

const toneOptions: Array<{ label: string; value: NotificationTone }> = [
  { label: '主要', value: 'primary' },
  { label: '成功', value: 'success' },
  { label: '提醒', value: 'warning' },
  { label: '危险', value: 'danger' },
  { label: '信息', value: 'info' },
]

const previewChannelOptions = [
  { label: '站内信', value: 'in_app' },
  { label: '钉钉', value: 'dingtalk' },
]
const loading = ref(false)
const saving = ref(false)
const testSending = ref<'' | 'in_app' | 'dingtalk'>('')
const styles = ref<NotificationStyle[]>([])
const configVersion = ref(1)
const selectedType = ref('')
const previewChannel = ref<'in_app' | 'dingtalk'>('in_app')
const recipientOptionsLoading = ref(false)
const recipientOptions = reactive<InAppNotificationRecipientOptions>({ users: [], departments: [] })
const testForm = reactive({ title: '', content: '', userIds: [] as number[] })

const selectedStyle = computed(() => styles.value.find(style => style.type === selectedType.value))
const selectedIconComponent = computed(() => iconOptions.find(option => option.value === selectedStyle.value?.icon)?.component || Message)

watch(() => props.modelValue, (visible) => {
  if (visible) void loadEditor()
})

async function loadEditor() {
  loading.value = true
  try {
    const requests: [ReturnType<typeof adminApi.notificationStylesGet>, Promise<unknown>?] = [adminApi.notificationStylesGet()]
    if (props.canSendDingTalk) requests.push(adminApi.dingTalkNotificationRecipientOptions())
    else if (props.canSend) requests.push(adminApi.inAppNotificationRecipientOptions())
    const [styleResponse, recipientResponse] = await Promise.all(requests)
    const config = styleResponse.data
    configVersion.value = Number(config?.version || 1)
    styles.value = Array.isArray(config?.styles) ? config.styles.map(style => ({ ...style })) : []
    selectedType.value = styles.value.some(style => style.type === selectedType.value)
      ? selectedType.value
      : (styles.value[0]?.type || '')
    const options = (recipientResponse as { data?: InAppNotificationRecipientOptions } | undefined)?.data
    recipientOptions.users = Array.isArray(options?.users) ? options.users : []
    recipientOptions.departments = Array.isArray(options?.departments) ? options.departments : []
    resetTestMessage()
  } finally {
    loading.value = false
  }
}

function selectType(type: string) {
  selectedType.value = type
  resetTestMessage()
}

function resetTestMessage() {
  const label = selectedStyle.value?.label || '系统通知'
  testForm.title = `【${label}】测试消息`
  testForm.content = `这是一条${label}测试消息。`
}

async function saveStyles(showSuccess: boolean) {
  if (!props.canEdit) return true
  if (styles.value.some(style => !style.label.trim())) {
    ElMessage.warning('请完善消息类型标签')
    return false
  }
  saving.value = true
  try {
    const payload: NotificationStyleConfig = {
      version: configVersion.value,
      styles: styles.value.map(style => ({ ...style, label: style.label.trim() })),
    }
    const response = await adminApi.notificationStylesSave(payload)
    configVersion.value = Number(response.data?.version || 1)
    styles.value = Array.isArray(response.data?.styles) ? response.data.styles.map(style => ({ ...style })) : payload.styles
    if (showSuccess) ElMessage.success('消息样式已保存')
    return true
  } finally {
    saving.value = false
  }
}

async function sendTest(channel: 'in_app' | 'dingtalk') {
  if (!selectedStyle.value) return
  if (!testForm.title.trim()) return ElMessage.warning('请输入测试标题')
  if (!testForm.content.trim()) return ElMessage.warning('请输入测试正文')
  const userIds = normalizeIDs(testForm.userIds)
  if (!userIds.length) return ElMessage.warning('请选择测试收件人')
  if (!await saveStyles(false)) return

  testSending.value = channel
  try {
    const payload = {
      requestId: newRequestID(),
      title: testForm.title.trim(),
      content: testForm.content,
      userIds,
      notificationType: selectedStyle.value.type,
    }
    const response = channel === 'dingtalk'
      ? await adminApi.notificationStyleDingTalkTest(payload)
      : await adminApi.notificationStyleInAppTest(payload)
    const sentCount = Number(response.data?.sentCount || 0)
    const skippedCount = Number(response.data?.skippedCount || 0)
    const failedCount = Number(response.data?.failedCount || 0)
    if (channel === 'dingtalk' && (skippedCount > 0 || failedCount > 0)) {
      ElMessage.warning(`钉钉测试完成：成功 ${sentCount} 人，跳过 ${skippedCount} 人，失败 ${failedCount} 人`)
    } else {
      ElMessage.success(`测试消息已发送给 ${sentCount} 人`)
    }
    emit('sent', channel)
  } finally {
    testSending.value = ''
  }
}

function normalizeIDs(values: Array<number | string>) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value > 0)))
}

function newRequestID() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function typeName(type: string) {
  return typeNames[type] || type
}

function toneName(tone: NotificationTone) {
  return toneOptions.find(option => option.value === tone)?.label || tone
}

function tagType(tone: NotificationTone) {
  if (tone === 'danger') return 'danger'
  if (tone === 'primary') return 'primary'
  return tone
}
</script>

<style scoped>
.style-editor { display: grid; grid-template-columns: 248px minmax(0, 1fr); min-height: 620px; border: 1px solid var(--admin-border); }
.style-types { overflow-y: auto; border-right: 1px solid var(--admin-border); background: #f7f9fc; }
.style-type { display: flex; width: 100%; min-height: 54px; align-items: center; gap: 10px; padding: 8px 12px; border: 0; border-bottom: 1px solid var(--admin-border); background: transparent; color: var(--admin-text-secondary); text-align: left; cursor: pointer; }
.style-type:hover { background: #eef5ff; }
.style-type.is-active { background: #eaf3ff; color: #2563eb; box-shadow: inset 3px 0 #2563eb; }
.style-type__text { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.style-type__text strong { font-size: 13px; font-weight: 600; }
.style-type__text small { overflow: hidden; color: var(--admin-muted); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.style-workspace { min-width: 0; overflow-y: auto; }
.style-form, .test-section { padding: 20px 22px; }
.test-section { border-top: 8px solid var(--admin-bg); }
.section-heading, .preview-heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.section-heading h3, .preview-heading h3, .test-section h3 { margin: 0; color: var(--admin-text); font-size: 16px; }
.section-heading code { display: block; margin-top: 3px; color: var(--admin-muted); font-size: 11px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.full-width { width: 100%; }
.icon-option { display: flex; align-items: center; gap: 8px; }
.tone-selector { display: flex; flex-wrap: wrap; }
.tone-selector :deep(.el-radio-button__inner) { display: inline-flex; align-items: center; gap: 6px; }
.tone-dot { width: 8px; height: 8px; flex: 0 0 8px; border-radius: 50%; }
.tone-dot--primary { background: #2563eb; }
.tone-dot--success { background: #00875a; }
.tone-dot--warning { background: #b45309; }
.tone-dot--danger { background: #c93756; }
.tone-dot--info { background: #64748b; }
.preview-heading { margin-top: 6px; }
.message-preview { display: flex; min-height: 116px; gap: 14px; padding: 16px; border: 1px solid var(--admin-border); border-left-width: 3px; border-radius: 6px; background: #fff; }
.message-preview--primary { border-left-color: #2563eb; }
.message-preview--success { border-left-color: #00875a; }
.message-preview--warning { border-left-color: #b45309; }
.message-preview--danger { border-left-color: #c93756; }
.message-preview--info { border-left-color: #64748b; }
.message-preview__icon { display: flex; width: 34px; height: 34px; flex: 0 0 34px; align-items: center; justify-content: center; border-radius: 6px; font-size: 18px; }
.message-preview__icon--primary { background: #eaf2ff; color: #2563eb; }
.message-preview__icon--success { background: #e7f7f1; color: #00875a; }
.message-preview__icon--warning { background: #fff4df; color: #b45309; }
.message-preview__icon--danger { background: #fff0f3; color: #c93756; }
.message-preview__icon--info { background: #eef2f6; color: #64748b; }
.message-preview__body { min-width: 0; flex: 1; }
.message-preview__meta { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-bottom: 8px; color: var(--admin-muted); font-size: 12px; }
.message-preview__body > strong { display: block; overflow: hidden; color: var(--admin-text); font-size: 14px; text-overflow: ellipsis; white-space: nowrap; }
.message-preview__body > p, .dingtalk-preview p { display: -webkit-box; overflow: hidden; margin: 7px 0 0; color: var(--admin-text-secondary); line-height: 20px; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
.dingtalk-preview { min-height: 116px; padding: 16px; border: 1px solid var(--admin-border); border-radius: 6px; background: #fff; }
.dingtalk-preview__brand { margin-bottom: 12px; color: #1476ff; font-size: 12px; font-weight: 600; }
.dingtalk-preview > strong { color: var(--admin-text); font-size: 14px; }
.test-section h3 { margin-bottom: 16px; }
.test-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; }
@media (max-width: 760px) {
  .style-editor { grid-template-columns: 1fr; }
  .style-types { display: flex; overflow-x: auto; border-right: 0; border-bottom: 1px solid var(--admin-border); }
  .style-type { width: 170px; flex: 0 0 170px; border-right: 1px solid var(--admin-border); border-bottom: 0; }
  .style-workspace { max-height: 64vh; }
  .form-grid { grid-template-columns: 1fr; gap: 0; }
}
</style>
