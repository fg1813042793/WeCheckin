import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const root = resolve(currentDir, '..')
const read = path => readFileSync(resolve(root, path), 'utf8')

const pagePath = 'src/views/notification/index.vue'
if (!existsSync(resolve(root, pagePath))) throw new Error(`in-app notification UI missing ${pagePath}`)

const routes = read('src/router/adminRoutes.ts')
for (const snippet of ["path: 'notifications'", "menuPath: '/notifications'", "../views/notification/index.vue"]) {
  if (!routes.includes(snippet)) throw new Error(`in-app notification route missing ${snippet}`)
}
if (!routes.includes("path: 'survey/notify', name: 'SurveyNotify', component: () => import('../views/notification/index.vue')")) {
  throw new Error('legacy survey notification route must reuse the canonical in-app notification page')
}

const api = read('src/api/index.ts')
for (const method of ['inAppNotificationList', 'inAppNotificationUnreadCount', 'inAppNotificationMarkRead', 'inAppNotificationMarkAllRead', 'inAppNotificationSend', 'dingTalkNotificationRecipientOptions', 'dingTalkNotificationSend']) {
  if (!api.includes(`${method}(`)) throw new Error(`in-app notification API missing ${method}`)
}
if (!api.includes('/in-app-notifications')) throw new Error('in-app notification API must use the canonical admin endpoint')
if (!api.includes('/dingtalk-notifications')) throw new Error('DingTalk notification API must use the canonical admin endpoint')

const page = read(pagePath)
for (const snippet of [
  '通知管理',
  '发送站内信',
  '发送钉钉通知',
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
]) {
  if (!page.includes(snippet)) throw new Error(`in-app notification page missing ${snippet}`)
}
