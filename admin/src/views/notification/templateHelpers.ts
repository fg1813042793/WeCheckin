import type { DingTalkSendTemplateInput, SendTemplateTarget } from '@/types/notificationTemplate'

export const templateMessageTypes = [
  { label: 'OA 消息', value: 'oa' },
  { label: '链接消息', value: 'link' },
  { label: '卡片消息', value: 'action_card' },
] as const

export type TemplateErrors = Record<string, string>
const length = (value: string) => Array.from(value).length

export function templateHeadColorPickerValue(value: string) {
  const normalized = String(value || '').replace(/^#/, '').toUpperCase()
  const rgb = normalized.length === 8 ? normalized.slice(2) : normalized
  return `#${/^[\dA-F]{6}$/.test(rgb) ? rgb : '1677FF'}`
}

export function templateHeadColorFromPicker(value: string | null) {
  const normalized = String(value || '#1677FF').replace(/^#/, '').slice(-6).toUpperCase()
  return `FF${/^[\dA-F]{6}$/.test(normalized) ? normalized : '1677FF'}`
}

export function validateMessage(title: string, content: string): TemplateErrors {
  const errors: TemplateErrors = {}
  if (!title.trim()) errors.title = '请输入通知标题'
  else if (length(title.trim()) > 255) errors.title = '标题最多 255 字'
  if (!content.trim()) errors.content = '请输入通知正文'
  else if (length(content) > 5000) errors.content = '正文最多 5000 字'
  return errors
}

export function validateTemplate(form: DingTalkSendTemplateInput): TemplateErrors {
  const errors = validateMessage(form.title, form.content)
  if (!form.name.trim()) errors.name = '请输入模板名称'
  else if (length(form.name.trim()) > 80) errors.name = '模板名称最多 80 字'
  if (!templateMessageTypes.some(type => type.value === form.messageType)) errors.messageType = '请选择 OA、链接或卡片消息'
  if (typeof form.enabled !== 'boolean') errors.enabled = '请选择启用状态'
  if (form.messageType === 'oa' && !/^[\da-f]{8}$/i.test(form.headColor)) errors.headColor = '请输入 8 位十六进制颜色，例如 FF1677FF'
  if (length(form.buttonTitle.trim()) > 20 || (form.messageType === 'action_card' && !form.buttonTitle.trim())) errors.buttonTitle = '请输入不超过 20 字的卡片按钮文字'
  if (form.target.kind === 'workflow_start') {
    if (!Number.isSafeInteger(form.target.definitionId) || Number(form.target.definitionId) <= 0) errors['target.definitionId'] = '请选择已发布流程'
  } else if (form.target.kind === 'url') {
    const value = form.target.url || ''
    let valid = length(value) <= 2048 && /^https:\/\//.test(value) && !/[\u0000-\u0020\u007f-\u009f\\]/.test(value)
    try {
      valid = valid && !/[\u0000-\u001f\u007f-\u009f]/.test(decodeURIComponent(value))
      const url = new URL(value)
      valid = valid && url.protocol === 'https:' && !!url.hostname && !url.username && !url.password
    } catch { valid = false }
    if (!valid) errors['target.url'] = '请输入完整 HTTPS 地址（最多 2048 字符，不含账号、密码或控制字符）'
  } else if (form.target.kind !== 'app_home') errors['target.kind'] = '请选择跳转目标'
  return errors
}

export function templatePayload(form: DingTalkSendTemplateInput): DingTalkSendTemplateInput {
  const target: SendTemplateTarget = { kind: form.target.kind }
  if (target.kind === 'workflow_start') target.definitionId = form.target.definitionId
  if (target.kind === 'url') target.url = form.target.url
  return {
    name: form.name.trim(), enabled: form.enabled, title: form.title.trim(), content: form.content,
    messageType: form.messageType, target,
    buttonTitle: form.messageType === 'action_card' ? form.buttonTitle.trim() : '',
    headColor: form.messageType === 'oa' ? form.headColor.toUpperCase() : '',
  }
}

export function newTemplateDraft(source?: DingTalkSendTemplateInput, copy = false): DingTalkSendTemplateInput {
  if (!source) return { name: '', enabled: true, title: '', content: '', messageType: 'oa', target: { kind: 'app_home' }, buttonTitle: '打开页面', headColor: 'FF1677FF' }
  return {
    ...templatePayload(source),
    target: { ...source.target },
    name: copy ? `${Array.from(source.name).slice(0, 75).join('')}（副本）` : source.name,
    headColor: source.headColor || 'FF1677FF', buttonTitle: source.buttonTitle || '打开页面',
  }
}

export function targetSummary(target: SendTemplateTarget) {
  if (target.kind === 'app_home') return '应用首页'
  if (target.kind === 'workflow_start') return `发起流程：${target.definitionName || `流程 #${target.definitionId}`}`
  return target.url || '未设置链接'
}

export function templateError(error: unknown, fallback: string) {
  return error && typeof error === 'object' && 'msg' in error && typeof error.msg === 'string' ? error.msg : fallback
}
