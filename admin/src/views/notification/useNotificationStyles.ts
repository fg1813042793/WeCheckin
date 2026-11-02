import {
  Bell,
  ChatDotRound,
  CircleCheck,
  Clock,
  Document,
  EditPen,
  InfoFilled,
  Message,
  Refresh,
  Share,
  WarningFilled,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { computed, onMounted, reactive, ref } from 'vue'
import { adminApi } from '@/api'
import type {
  DingTalkMessageType,
  DingTalkNotificationTemplate,
  InAppNotificationRecipientOptions,
  NotificationStyle,
  NotificationStyleConfig,
  NotificationTone,
} from '@/api/types'


export function useNotificationStyles(props: { canEdit: boolean; canSend: boolean; canSendDingTalk: boolean }, onSent: (channel: 'in_app' | 'dingtalk') => void) {
  const typeNames: Record<string, string> = {
    task_arrived: '任务到达',
    task_reminder: '任务催办',
    approval_result_approved: '审批通过结果',
    approval_result_rejected: '审批驳回结果',
    approval_result_returned: '审批退回结果',
    node_cc: '节点抄送',
    node_notify: '节点通知',
    instance_commented: '流程评论',
    instance_form_revised: '表单修改',
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
    { label: '编辑', value: 'edit-pen', component: EditPen },
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

  const dingTalkMessageTypeOptions: Array<{ label: string; value: DingTalkMessageType }> = [
    { label: '自动兼容（原行为）', value: 'auto' },
    { label: '文本消息', value: 'text' },
    { label: '图片消息', value: 'image' },
    { label: '语音消息', value: 'voice' },
    { label: '文件消息', value: 'file' },
    { label: '链接消息', value: 'link' },
    { label: 'OA 消息', value: 'oa' },
    { label: 'Markdown 消息', value: 'markdown' },
    { label: '卡片消息', value: 'action_card' },
  ]

  const previewChannelOptions = [
    { label: '站内信', value: 'in_app' },
    { label: '钉钉', value: 'dingtalk' },
  ]
  const loading = ref(false)
  const loadError = ref('')
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
  const dingTalkPreview = computed(() => {
    const template = selectedStyle.value?.dingTalk || defaultDingTalkTemplate()
    const result = renderDingTalkTemplate(template)
    if (result.messageType === 'auto') result.messageType = result.url ? 'oa' : 'text'
    return result
  })
  const dingTalkHeadColor = computed({
    get: () => {
      const value = String(selectedStyle.value?.dingTalk.headColor || '').replace(/^#/, '')
      return `#${value.length === 8 ? value.slice(2) : value || '1677FF'}`
    },
    set: (value: string | null) => {
      if (selectedStyle.value) selectedStyle.value.dingTalk.headColor = `FF${String(value || '#1677FF').replace(/^#/, '').slice(-6).toUpperCase()}`
    },
  })

  onMounted(loadEditor)

  async function loadEditor() {
    if (loading.value) return
    loading.value = true
    loadError.value = ''
    try {
      const requests: [ReturnType<typeof adminApi.notificationStylesGet>, Promise<unknown>?] = [adminApi.notificationStylesGet()]
      if (props.canSendDingTalk) requests.push(adminApi.dingTalkNotificationRecipientOptions())
      else if (props.canSend) requests.push(adminApi.inAppNotificationRecipientOptions())
      const [styleResponse, recipientResponse] = await Promise.all(requests)
      const config = styleResponse.data
      configVersion.value = Number(config?.version || 1)
      styles.value = Array.isArray(config?.styles) ? config.styles.map(normalizeStyle) : []
      selectedType.value = styles.value.some(style => style.type === selectedType.value)
        ? selectedType.value
        : (styles.value[0]?.type || '')
      const options = (recipientResponse as { data?: InAppNotificationRecipientOptions } | undefined)?.data
      recipientOptions.users = Array.isArray(options?.users) ? options.users : []
      recipientOptions.departments = Array.isArray(options?.departments) ? options.departments : []
      resetTestMessage()
    } catch {
      loadError.value = '业务消息样式加载失败，请重试。'
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

  async function saveStyles() {
    if (!props.canEdit || saving.value || loading.value || loadError.value) return
    if (styles.value.some(style => !style.label.trim())) {
      ElMessage.warning('请完善消息类型标签')
      return
    }
    saving.value = true
    try {
      const payload: NotificationStyleConfig = {
        version: configVersion.value,
        styles: styles.value.map(style => ({ ...style, label: style.label.trim() })),
      }
      const response = await adminApi.notificationStylesSave(payload)
      configVersion.value = Number(response.data?.version || 1)
      styles.value = Array.isArray(response.data?.styles) ? response.data.styles.map(normalizeStyle) : payload.styles
      ElMessage.success('消息样式已保存')
    } finally {
      saving.value = false
    }
  }

  async function sendTest(channel: 'in_app' | 'dingtalk') {
    if (testSending.value || (channel === 'dingtalk' ? !props.canSendDingTalk : !props.canSend)) return
    if (!selectedStyle.value) return
    if (!testForm.title.trim()) return ElMessage.warning('请输入测试标题')
    if (!testForm.content.trim()) return ElMessage.warning('请输入测试正文')
    const userIds = normalizeIDs(testForm.userIds)
    if (!userIds.length) return ElMessage.warning('请选择测试收件人')

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
      onSent(channel)
    } finally {
      testSending.value = ''
    }
  }

  function normalizeIDs(values: Array<number | string>) {
    return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value > 0)))
  }

  function normalizeStyle(style: NotificationStyle): NotificationStyle {
    return {
      ...style,
      dingTalk: { ...defaultDingTalkTemplate(), ...(style.dingTalk || {}) },
    }
  }

  function defaultDingTalkTemplate(): DingTalkNotificationTemplate {
    return {
      messageType: 'auto', title: '{{title}}', content: '{{content}}', url: '{{url}}', picUrl: '{{picUrl}}',
      sourceName: '{{sourceName}}', mediaId: '{{mediaId}}', duration: 0, buttonTitle: '查看流程', headColor: 'FF1677FF',
    }
  }

  function handleDingTalkTypeChange(value: DingTalkMessageType) {
    const template = selectedStyle.value?.dingTalk
    if (!template) return
    if (['image', 'voice', 'file'].includes(value) && template.mediaId === '{{mediaId}}') template.mediaId = ''
    if (value === 'voice' && template.duration <= 0) template.duration = 1
  }

  function dingTalkFieldVisible(field: keyof DingTalkNotificationTemplate) {
    const type = selectedStyle.value?.dingTalk.messageType || 'auto'
    const fields: Record<string, Array<keyof DingTalkNotificationTemplate>> = {
      auto: ['title', 'content', 'url', 'picUrl', 'sourceName'],
      text: ['content'],
      image: ['mediaId'],
      voice: ['mediaId', 'duration'],
      file: ['mediaId'],
      link: ['title', 'content', 'url', 'picUrl'],
      oa: ['title', 'content', 'url', 'picUrl', 'sourceName', 'headColor'],
      markdown: ['title', 'content'],
      action_card: ['title', 'content', 'url', 'buttonTitle'],
    }
    return fields[type]?.includes(field) || false
  }

  function renderDingTalkTemplate(template: DingTalkNotificationTemplate) {
    const values: Record<string, string> = {
      '{{title}}': testForm.title || '测试通知标题',
      '{{content}}': testForm.content || '测试通知正文',
      '{{url}}': 'https://example.test/workflow/detail',
      '{{sourceName}}': 'WeCheckin',
      '{{picUrl}}': 'https://example.test/logo.png',
      '{{mediaId}}': '',
      '{{duration}}': String(template.duration || 0),
    }
    const render = (value: string) => Object.entries(values).reduce((result, [token, replacement]) => result.split(token).join(replacement), value || '').trim()
    return {
      ...template,
      title: render(template.title), content: render(template.content), url: render(template.url),
      picUrl: render(template.picUrl), sourceName: render(template.sourceName), mediaId: render(template.mediaId),
      buttonTitle: render(template.buttonTitle),
    }
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

  function dingTalkMessageTypeName(type: string) {
    return dingTalkMessageTypeOptions.find(option => option.value === type)?.label || type
  }

  function tagType(tone: NotificationTone) {
    if (tone === 'danger') return 'danger'
    if (tone === 'primary') return 'primary'
    return tone
  }

  return { loading, loadError, loadEditor, saving, testSending, styles, selectedType, previewChannel, recipientOptionsLoading, recipientOptions, testForm, selectedStyle, selectedIconComponent, dingTalkPreview, dingTalkHeadColor, iconOptions, toneOptions, dingTalkMessageTypeOptions, previewChannelOptions, selectType, saveStyles, sendTest, handleDingTalkTypeChange, dingTalkFieldVisible, typeName, toneName, dingTalkMessageTypeName, tagType }
}
