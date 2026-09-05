<script setup lang="ts">
import type { InAppNotification, NotificationTone } from '@/api/notifications'
import { computed, ref, watch } from 'vue'
import {
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
} from '@/api/notifications'
import { NOTIFICATION_HISTORY_CONTENT_KEY } from '@/pages/notifications/notification-route-keys'
import { workflowInstanceContentKey } from '@/pages/workflow/workflow-route-keys'
import { useAppContentStore } from '@/stores'

const props = withDefaults(defineProps<{
  modelValue: boolean
  mobile?: boolean
  unreadCount?: number
}>(), {
  mobile: false,
  unreadCount: 0,
})

const emit = defineEmits<{
  'update:modelValue': [visible: boolean]
  'unread-change': [count: number]
}>()

const appContent = useAppContentStore()
const loading = ref(false)
const loadingMore = ref(false)
const markingAll = ref(false)
const errorMessage = ref('')
const notifications = ref<InAppNotification[]>([])
const selectedNotification = ref<InAppNotification | null>(null)
const page = ref(1)
const pageSize = 20
const total = ref(0)

interface NotificationTypeMeta {
  label: string
  icon: string
  color: string
  tone: NotificationTone
}

const notificationTypeMetas: Record<string, NotificationTypeMeta> = {
  task_arrived: { label: '待处理', icon: 'clock', color: '#b45309', tone: 'warning' },
  task_reminder: { label: '处理提醒', icon: 'bell', color: '#b45309', tone: 'warning' },
  approval_result_approved: { label: '审批通过', icon: 'checkmark-circle', color: '#00875a', tone: 'success' },
  approval_result_rejected: { label: '审批驳回', icon: 'error-circle', color: '#c93756', tone: 'danger' },
  approval_result_returned: { label: '审批退回', icon: 'reload', color: '#b45309', tone: 'warning' },
  node_cc: { label: '流程抄送', icon: 'share', color: '#475569', tone: 'info' },
  node_notify: { label: '流程通知', icon: 'email', color: '#2563eb', tone: 'primary' },
  instance_commented: { label: '流程评论', icon: 'chat', color: '#2563eb', tone: 'primary' },
  instance_form_revised: { label: '表单修改', icon: 'edit-pen', color: '#2563eb', tone: 'primary' },
  workflow: { label: '流程消息', icon: 'file-text', color: '#475569', tone: 'info' },
  admin_manual: { label: '系统通知', icon: 'email', color: '#2563eb', tone: 'primary' },
  scheduled_task: { label: '定时通知', icon: 'clock', color: '#475569', tone: 'info' },
  survey_stat: { label: '问卷统计', icon: 'file-text', color: '#475569', tone: 'info' },
}

const defaultNotificationTypeMeta: NotificationTypeMeta = {
  label: '系统消息',
  icon: 'email',
  color: '#475569',
  tone: 'info',
}

const notificationToneColors: Record<NotificationTone, string> = {
  primary: '#2563eb',
  success: '#00875a',
  warning: '#b45309',
  danger: '#c93756',
  info: '#475569',
}

const visible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value),
})

const hasMore = computed(() => notifications.value.length < total.value)
const unreadLabel = computed(() => props.unreadCount > 99 ? '99+' : String(props.unreadCount))

function notificationTypeMeta(notification: InAppNotification | string) {
  const type = typeof notification === 'string' ? notification : notification.type
  const fallback = notificationTypeMetas[String(type || '').trim()] || defaultNotificationTypeMeta
  if (typeof notification === 'string' || !notification.style)
    return fallback
  const style = notification.style
  const tone = notificationToneColors[style.tone] ? style.tone : fallback.tone
  return {
    label: String(style.label || '').trim() || fallback.label,
    icon: String(style.icon || '').trim() || fallback.icon,
    tone,
    color: notificationToneColors[tone],
  }
}

function notificationToneClass(notification: InAppNotification, block: string) {
  return `${block}--${notificationTypeMeta(notification).tone}`
}

watch(() => props.modelValue, (value) => {
  if (value) {
    selectedNotification.value = null
    void loadNotifications(false)
  }
})

function closePanel() {
  visible.value = false
}

function backToList() {
  selectedNotification.value = null
}

async function loadNotifications(append: boolean) {
  if (loading.value || loadingMore.value)
    return
  const nextPage = append ? page.value + 1 : 1
  if (append)
    loadingMore.value = true
  else
    loading.value = true
  errorMessage.value = ''
  try {
    const response = await listNotifications(nextPage, pageSize, true)
    const payload = response?.data
    const items = Array.isArray(payload?.list) ? payload.list : []
    notifications.value = append ? [...notifications.value, ...items] : items
    total.value = Number(payload?.total || 0)
    page.value = nextPage
  }
  catch {
    errorMessage.value = '站内信加载失败'
  }
  finally {
    loading.value = false
    loadingMore.value = false
  }
}

async function markRead(notification: InAppNotification) {
  if (notification.isRead === 1)
    return true
  try {
    await markNotificationRead(notification.id)
    notification.isRead = 1
    notifications.value = notifications.value.filter(item => item.id !== notification.id)
    total.value = Math.max(0, total.value - 1)
    emit('unread-change', Math.max(0, props.unreadCount - 1))
    return true
  }
  catch {
    uni.showToast({ title: '标记已读失败', icon: 'none' })
    return false
  }
}

function openNotification(notification: InAppNotification) {
  selectedNotification.value = notification
  void markRead(notification)
  if (notification.sourceType === 'workflow_instance' && notification.sourceId) {
    const key = workflowInstanceContentKey(notification.sourceId)
    if (!key)
      return
    appContent.openDynamicTab({
      key,
      label: notification.title || '流程详情',
      icon: 'file-text',
      path: `/pages/index/index?view=${encodeURIComponent(key)}`,
    })
    closePanel()
  }
}

function openNotificationHistory() {
  appContent.openDynamicTab({
    key: NOTIFICATION_HISTORY_CONTENT_KEY,
    label: '站内信历史',
    icon: 'email',
    path: `/pages/index/index?view=${encodeURIComponent(NOTIFICATION_HISTORY_CONTENT_KEY)}`,
  })
  closePanel()
}

async function markAllRead() {
  if (markingAll.value || props.unreadCount <= 0)
    return
  markingAll.value = true
  try {
    await markAllNotificationsRead()
    notifications.value = []
    total.value = 0
    if (selectedNotification.value)
      selectedNotification.value.isRead = 1
    emit('unread-change', 0)
    uni.showToast({ title: '已全部标记为已读', icon: 'success' })
  }
  catch {
    uni.showToast({ title: '操作失败', icon: 'none' })
  }
  finally {
    markingAll.value = false
  }
}

function formatTime(value: number) {
  if (!value)
    return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime()))
    return '-'
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}/${pad(date.getMonth() + 1)}/${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}
</script>

<template>
  <u-popup
    v-model="visible"
    mode="right"
    :width="mobile ? '100%' : '420px'"
    height="100%"
    custom-class="app-notification-popup app-pc-control-scope"
    :border-radius="0"
    :safe-area-inset-bottom="mobile"
  >
    <view class="notification-panel" :class="{ 'notification-panel--mobile': mobile }">
      <view class="notification-panel__header">
        <u-button
          v-if="selectedNotification"
          custom-class="notification-panel__icon-btn app-icon-button"
          title="返回站内信列表"
          @click="backToList"
        >
          <u-icon name="arrow-left" size="20" color="#4e5969" />
        </u-button>
        <view class="notification-panel__heading">
          <text class="notification-panel__title">
            {{ selectedNotification ? '消息详情' : '站内信' }}
          </text>
          <u-button
            v-if="!selectedNotification"
            custom-class="notification-panel__history"
            title="查看站内信历史"
            @click="openNotificationHistory"
          >
            <u-icon name="clock" size="14" color="#008f72" />
            <text>历史</text>
          </u-button>
          <text v-if="!selectedNotification && unreadCount > 0" class="notification-panel__unread-copy">
            {{ unreadLabel }} 条未读
          </text>
        </view>
        <u-button
          v-if="!selectedNotification && unreadCount > 0"
          custom-class="notification-panel__read-all"
          :loading="markingAll"
          @click="markAllRead"
        >
          全部已读
        </u-button>
        <u-button custom-class="notification-panel__icon-btn app-icon-button" title="关闭站内信" @click="closePanel">
          <u-icon name="close" size="20" color="#4e5969" />
        </u-button>
      </view>

      <view v-if="selectedNotification" class="notification-detail">
        <view class="notification-detail__meta">
          <view
            class="notification-detail__kind"
            :class="notificationToneClass(selectedNotification, 'notification-detail__kind')"
          >
            <u-icon
              :name="notificationTypeMeta(selectedNotification).icon"
              size="14"
              :color="notificationTypeMeta(selectedNotification).color"
            />
            <text>{{ notificationTypeMeta(selectedNotification).label }}</text>
          </view>
          <text class="notification-detail__title">
            {{ selectedNotification.title || '站内消息' }}
          </text>
          <text class="notification-detail__time">
            {{ formatTime(selectedNotification.addTime) }}
          </text>
        </view>
        <text class="notification-detail__content">
          {{ selectedNotification.content || '暂无内容' }}
        </text>
      </view>

      <scroll-view v-else scroll-y class="notification-panel__body">
        <view v-if="loading" class="notification-panel__state">
          <u-loading mode="circle" size="36" />
          <text>加载中...</text>
        </view>
        <view v-else-if="errorMessage" class="notification-panel__state">
          <u-icon name="error-circle" size="36" color="#86909c" />
          <text>{{ errorMessage }}</text>
          <u-button custom-class="notification-panel__retry" @click="loadNotifications(false)">
            重新加载
          </u-button>
        </view>
        <view v-else-if="notifications.length === 0" class="notification-panel__state">
          <u-icon name="email" size="42" color="#c9cdd4" />
          <text class="notification-panel__empty-title">
            暂无未读消息
          </text>
        </view>
        <view v-else class="notification-list">
          <view
            v-for="item in notifications"
            :key="item.id"
            class="notification-item"
            :class="{ 'notification-item--unread': item.isRead !== 1 }"
            @click="openNotification(item)"
          >
            <view class="notification-item__indicator" />
            <view
              class="notification-item__kind-icon"
              :class="notificationToneClass(item, 'notification-item__kind-icon')"
            >
              <u-icon
                :name="notificationTypeMeta(item).icon"
                size="17"
                :color="notificationTypeMeta(item).color"
              />
            </view>
            <view class="notification-item__content">
              <view class="notification-item__topline">
                <text
                  class="notification-item__kind"
                  :class="notificationToneClass(item, 'notification-item__kind')"
                >
                  {{ notificationTypeMeta(item).label }}
                </text>
                <text class="notification-item__title">
                  {{ item.title || '站内消息' }}
                </text>
                <text class="notification-item__time">
                  {{ formatTime(item.addTime) }}
                </text>
              </view>
              <text class="notification-item__summary">
                {{ item.content || '暂无内容' }}
              </text>
            </view>
            <u-icon name="arrow-right" size="16" color="#c9cdd4" />
          </view>
          <view class="notification-list__footer">
            <u-button
              v-if="hasMore"
              custom-class="notification-panel__load-more"
              :loading="loadingMore"
              @click="loadNotifications(true)"
            >
              加载更多
            </u-button>
            <text v-else>
              没有更多消息
            </text>
          </view>
        </view>
      </scroll-view>
    </view>
  </u-popup>
</template>

<style scoped lang="scss">
.notification-panel {
  width: 420px;
  max-width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f7f8fa;
  color: #1f2329;
}

.notification-panel--mobile {
  width: 100vw;
}

.notification-panel__header {
  flex: 0 0 64px;
  min-width: 0;
  padding: 0 16px;
  border-bottom: 1px solid #e5e6eb;
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
}

.notification-panel__heading {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.notification-panel__title {
  font-size: 18px;
  font-weight: 700;
}

.notification-panel__unread-copy,
.notification-item__time,
.notification-detail__time,
.notification-list__footer {
  color: #86909c;
  font-size: 12px;
}

.notification-panel__icon-btn,
.notification-panel__history,
.notification-panel__read-all,
.notification-panel__retry,
.notification-panel__load-more,
:deep(.notification-panel__icon-btn),
:deep(.notification-panel__history),
:deep(.notification-panel__read-all),
:deep(.notification-panel__retry),
:deep(.notification-panel__load-more) {
  margin: 0;
  box-shadow: none;
}

.notification-panel__history,
:deep(.notification-panel__history) {
  width: auto;
  min-width: 48px;
  height: 28px;
  min-height: 28px;
  margin: 0;
  padding: 0 5px;
  border: 0;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  background: transparent;
  color: #008f72;
  font-size: 12px;
}

.notification-panel__icon-btn,
:deep(.notification-panel__icon-btn) {
  width: 32px;
  height: 32px;
  min-height: 32px;
  padding: 0;
  border: 1px solid #e5e6eb;
  border-radius: 4px;
  background: #fff;
}

.notification-panel__read-all,
:deep(.notification-panel__read-all) {
  width: auto;
  min-width: 64px;
  height: 32px;
  min-height: 32px;
  padding: 0 6px;
  border: 0;
  background: transparent;
  color: #008f72;
  font-size: 13px;
}

.notification-panel__body,
.notification-detail {
  min-height: 0;
  flex: 1;
}

.notification-panel__body {
  height: 0;
}

.notification-panel__state {
  min-height: 280px;
  padding: 48px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #86909c;
  font-size: 14px;
}

.notification-panel__empty-title {
  color: #4e5969;
}

.notification-panel__retry,
.notification-panel__load-more,
:deep(.notification-panel__retry),
:deep(.notification-panel__load-more) {
  width: auto;
  min-width: 88px;
  height: 32px;
  min-height: 32px;
  padding: 0 14px;
  border-radius: 4px;
  font-size: 13px;
}

.notification-list {
  padding: 8px 0 24px;
  background: #fff;
}

.notification-item {
  position: relative;
  min-height: 88px;
  padding: 16px;
  border-bottom: 1px solid #f0f1f2;
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
}

.notification-item--unread {
  background: #f2fbf8;
}

.notification-item__indicator {
  position: absolute;
  top: 13px;
  left: 8px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: transparent;
}

.notification-item--unread .notification-item__indicator {
  background: #00a67e;
}

.notification-item__kind-icon {
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  border: 1px solid transparent;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.notification-item__content {
  min-width: 0;
  flex: 1;
}

.notification-item__topline {
  display: flex;
  align-items: center;
  gap: 8px;
}

.notification-item__kind,
.notification-detail__kind {
  width: fit-content;
  border: 1px solid transparent;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 600;
  line-height: 18px;
}

.notification-item__kind {
  flex: 0 0 auto;
  padding: 0 5px;
}

.notification-detail__kind {
  min-height: 24px;
  padding: 2px 7px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.notification-item__kind-icon--primary,
.notification-item__kind--primary,
.notification-detail__kind--primary {
  border-color: #bfdbfe;
  background: #eff6ff;
  color: #1d4ed8;
}

.notification-item__kind-icon--success,
.notification-item__kind--success,
.notification-detail__kind--success {
  border-color: #a7f3d0;
  background: #ecfdf5;
  color: #047857;
}

.notification-item__kind-icon--warning,
.notification-item__kind--warning,
.notification-detail__kind--warning {
  border-color: #fed7aa;
  background: #fff7ed;
  color: #b45309;
}

.notification-item__kind-icon--danger,
.notification-item__kind--danger,
.notification-detail__kind--danger {
  border-color: #fecdd3;
  background: #fff1f2;
  color: #be123c;
}

.notification-item__kind-icon--info,
.notification-item__kind--info,
.notification-detail__kind--info {
  border-color: #cbd5e1;
  background: #f8fafc;
  color: #475569;
}

.notification-item__title {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-item__summary {
  margin-top: 8px;
  overflow: hidden;
  display: -webkit-box;
  color: #646a73;
  font-size: 13px;
  line-height: 20px;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.notification-list__footer {
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.notification-detail {
  padding: 24px;
  overflow-y: auto;
  background: #fff;
}

.notification-detail__meta {
  padding-bottom: 18px;
  border-bottom: 1px solid #e5e6eb;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.notification-detail__title {
  font-size: 18px;
  font-weight: 700;
  line-height: 26px;
}

.notification-detail__content {
  margin-top: 20px;
  color: #4e5969;
  font-size: 14px;
  line-height: 24px;
  white-space: pre-wrap;
}

@media (max-width: 768px) {
  .notification-panel__header {
    padding: 0 12px;
  }

  .notification-detail {
    padding: 20px 16px;
  }
}
</style>
