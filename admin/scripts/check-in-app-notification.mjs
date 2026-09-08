import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const read = path => readFileSync(resolve(root, path), 'utf8')

const pagePath = 'src/views/notification/index.vue'
if (!existsSync(resolve(root, pagePath))) throw new Error(`in-app notification UI missing ${pagePath}`)
const workflowRecordsPath = 'src/views/notification/components/WorkflowDeliveryRecords.vue'
if (!existsSync(resolve(root, workflowRecordsPath))) throw new Error(`workflow delivery record UI missing ${workflowRecordsPath}`)

const routes = read('src/router/adminRoutes.ts')
for (const snippet of ["path: 'notifications'", "title: '通知记录管理'", "menuPath: '/notifications'", "../views/notification/index.vue"]) {
  if (!routes.includes(snippet)) throw new Error(`in-app notification route missing ${snippet}`)
}
if (!routes.includes("path: 'survey/notify', name: 'SurveyNotify', component: () => import('../views/notification/index.vue')")) {
  throw new Error('legacy survey notification route must reuse the canonical in-app notification page')
}

const api = read('src/api/index.ts')
for (const method of ['inAppNotificationList', 'inAppNotificationDelete', 'inAppNotificationSend', 'dingTalkNotificationRecipientOptions', 'dingTalkNotificationSend', 'notificationStylesGet', 'notificationStylesSave', 'notificationStyleInAppTest', 'notificationStyleDingTalkTest', 'workflowNotificationSend', 'workflowNotificationDelete']) {
  if (!api.includes(`${method}(`)) throw new Error(`in-app notification API missing ${method}`)
}
if (!api.includes('/in-app-notifications')) throw new Error('in-app notification API must use the canonical admin endpoint')
if (!api.includes('/dingtalk-notifications')) throw new Error('DingTalk notification API must use the canonical admin endpoint')

const page = read(pagePath)
for (const snippet of [
  '通知记录管理',
  '接收人',
  'recipientName',
  'sourceType',
  'notificationType',
  'isRead',
  'addTimeRange',
  '发送站内信',
  '发送钉钉通知',
	'通知记录详情',
	'查看',
	'删除',
	'ElMessageBox.confirm',
	"hasPerm('admin:menu:notification:delete')",
  "hasPerm('admin:menu:notification:send')",
  "hasPerm('admin:menu:notification:dingtalk-send')",
  'WorkflowUserTreePicker',
  "value: 'all'",
  "value: 'departments'",
  "value: 'users'",
  'departmentModelValue',
  'const sendRequestID = ref',
  'sendRequestID.value = newRequestID()',
  'requestId: sendRequestID.value',
  "sendChannel.value === 'dingtalk'",
  'NotificationStyleDialog',
  'WorkflowDeliveryRecords',
  "name=\"workflow\"",
  "hasPerm('admin:menu:workflow:notification:list')",
	"hasPerm('admin:menu:workflow:notification:delete')",
  "消息样式",
  "hasPerm('admin:menu:notification:style:list')",
]) {
  if (!page.includes(snippet)) throw new Error(`in-app notification page missing ${snippet}`)
}

const workflowRecords = read(workflowRecordsPath)
for (const snippet of [
  '流程实例 ID',
  '接收人 ID',
  '全部事件',
  '全部渠道',
  '全部状态',
  '最近错误',
  '流程单号',
  '投递详情',
  '处理待投递通知',
  'adminApi.workflowNotificationList',
  'adminApi.workflowNotificationRetry',
  'adminApi.workflowNotificationSend',
  'adminApi.workflowNotificationDispatchDue',
	'adminApi.workflowNotificationDelete',
	'canDelete',
	'deleteRecord(row)',
	'删除待投递记录后，该通知将不会再自动发送',
  'manualSend(row)',
  "row.status === 'failed' || row.status === 'dead'",
]) {
  if (!workflowRecords.includes(snippet)) throw new Error(`workflow delivery record UI missing ${snippet}`)
}
for (const obsolete of ['unreadCount', 'loadUnreadCount', 'markAllRead', 'markRead(row.id)', '全部已读', '标为已读']) {
  if (page.includes(obsolete)) throw new Error(`notification record management must not expose inbox action ${obsolete}`)
}
const notificationContentStyle = page.match(/\.notification-content\s*\{([^}]*)\}/)?.[1] || ''
for (const declaration of ['overflow: hidden', 'white-space: nowrap', 'text-overflow: ellipsis']) {
  if (!notificationContentStyle.includes(declaration)) throw new Error(`notification content cell missing ${declaration}`)
}
if (notificationContentStyle.includes('line-clamp')) throw new Error('notification content cell must remain single-line')

const styleDialog = read('src/views/notification/components/NotificationStyleDialog.vue')
for (const snippet of [
  '实时预览',
  '测试收件人',
  '发送站内信测试',
  '发送钉钉测试',
  'notificationType: selectedStyle.value.type',
  'adminApi.notificationStylesSave',
  'adminApi.notificationStyleInAppTest',
  'adminApi.notificationStyleDingTalkTest',
  '钉钉消息模板',
  "value: 'text'",
  "value: 'image'",
  "value: 'voice'",
  "value: 'file'",
  "value: 'link'",
  "value: 'oa'",
  "value: 'markdown'",
  "value: 'action_card'",
  "'mediaId'",
  "'duration'",
  "'headColor'",
  '已发布流程定义的 Logo',
  'DINGTALK_H5_LOGO_URL',
  'dingtalk://dingtalkclient/action/openapp',
  'redirect_url',
  '0_&lt;AgentID&gt;',
]) {
  if (!styleDialog.includes(snippet)) throw new Error(`notification style dialog missing ${snippet}`)
}
if (styleDialog.includes('保存并发送')) throw new Error('notification style test buttons should use send-test wording')
if (styleDialog.includes('await saveStyles(false)')) throw new Error('sending a style test must not persist unsaved style changes')
if ((styleDialog.match(/saveStyles\(/g) || []).length !== 2) {
  throw new Error('notification styles must only be persisted by the explicit save action')
}
