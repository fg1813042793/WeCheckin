import { computed, reactive, ref } from 'vue'
import { adminApi } from '@/api'
import type { InAppNotificationRecipientOptions, InAppNotificationScope, InAppNotificationSendResult } from '@/api/types'
import { templateError, validateMessage } from './templateHelpers'
import { useSendTemplate } from './useSendTemplate'

export type NotificationSendChannel = 'in_app' | 'dingtalk'
export function useNotificationSend(channel: () => NotificationSendChannel, permitted: () => boolean, confirmReplace: () => Promise<boolean>) {
  const form = reactive({ title: '', content: '', scope: 'all' as InAppNotificationScope, userIds: [] as number[], departmentIds: [] as number[] })
  const templates = useSendTemplate(form, confirmReplace)
  const sending = ref(false)
  const sendError = ref('')
  const errors = ref<Record<string, string>>({})
  const sendRequestID = ref('')
  const recipientOptions = ref<InAppNotificationRecipientOptions>({ users: [], departments: [] })
  const recipientOptionsLoading = ref(false)
  const recipientError = ref('')
  const recipientLoaded = ref(false)
  let epoch = 0
  const canSubmit = computed(() => permitted() && !sending.value && recipientLoaded.value && !recipientOptionsLoading.value && !recipientError.value && (channel() !== 'dingtalk' || templates.canSend.value))

  async function loadRecipientOptions() {
    if (!permitted() || recipientOptionsLoading.value) return
    const requestEpoch = epoch
    recipientOptionsLoading.value = true
    recipientError.value = ''
    try {
      const response = channel() === 'dingtalk'
        ? await adminApi.dingTalkNotificationRecipientOptions()
        : await adminApi.inAppNotificationRecipientOptions()
      if (requestEpoch !== epoch) return
      recipientOptions.value = response.data
      recipientLoaded.value = true
    } catch (cause) {
      if (requestEpoch === epoch) recipientError.value = templateError(cause, '收件范围加载失败，请重试。')
    } finally { if (requestEpoch === epoch) recipientOptionsLoading.value = false }
  }
  async function start() {
    dispose()
    Object.assign(form, { title: '', content: '', scope: 'all', userIds: [], departmentIds: [] })
    errors.value = {}
    sendError.value = ''
    sendRequestID.value = newRequestID()
    recipientOptions.value = { users: [], departments: [] }
    recipientError.value = ''
    recipientLoaded.value = false
    if (!permitted()) return
    await Promise.all([loadRecipientOptions(), ...(channel() === 'dingtalk' ? [templates.loadOptions()] : [])])
  }
  function dispose() {
    epoch++
    templates.reset()
    recipientOptionsLoading.value = false
  }
  async function submit(): Promise<InAppNotificationSendResult | null> {
    if (!canSubmit.value) return null
    errors.value = validateMessage(form.title, form.content)
    const userIds = normalizeIDs(form.userIds)
    const departmentIds = normalizeIDs(form.departmentIds)
    if (form.scope === 'users' && !userIds.length) errors.value.userIds = '请选择收件用户'
    if (form.scope === 'departments' && !departmentIds.length) errors.value.departmentIds = '请选择收件部门'
    if (Object.keys(errors.value).length) return null
    sending.value = true
    sendError.value = ''
    try {
      const payload = {
        title: form.title.trim(), content: form.content, scope: form.scope,
        ...(form.scope === 'users' ? { userIds } : {}),
        ...(form.scope === 'departments' ? { departmentIds } : {}),
      }
      const response = channel() === 'dingtalk'
        ? await adminApi.dingTalkNotificationSend({ ...templates.buildPayload(), ...payload })
        : await adminApi.inAppNotificationSend({ ...payload, requestId: sendRequestID.value })
      return response.data
    } catch (cause) {
      sendError.value = templateError(cause, '发送失败，已保留本次内容，请检查后重试。')
      return null
    } finally { sending.value = false }
  }
  return { form, templates, sending, sendError, errors, recipientOptions, recipientOptionsLoading, recipientError, canSubmit, start, dispose, loadRecipientOptions, submit }
}

function normalizeIDs(values: Array<number | string>) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isSafeInteger(value) && value > 0)))
}
function newRequestID() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}
