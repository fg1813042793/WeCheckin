<script setup lang="ts">
import type { InAppNotification } from '@/api/notifications'
import type { FeedbackNotificationDynamicTab } from '@/pages/notifications/feedback-notification-open'
import { useLocale } from 'uview-pro'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import {
  deleteNotification,
  listNotifications,
  markNotificationRead,
} from '@/api/notifications'
import { confirmUnsavedNavigation, navigateWithUnsavedGuard } from '@/components/app-shell/app-shell-navigation-guard'
import { feedbackDetailContentKey } from '@/pages/feedback/feedback-route-keys'
import { isFeedbackNotification, openFeedbackNotification } from '@/pages/notifications/feedback-notification-open'
import { runNotificationMarkRead } from '@/pages/notifications/notification-mark-read'
import {
  isWorkflowNotification,
  openWorkflowNotification,
  workflowNotificationLabel,
} from '@/pages/notifications/workflow-notification-open'
import { workflowFormRevisionDetailContentKey, workflowInstanceContentKey } from '@/pages/workflow/workflow-route-keys'
import { useAppContentStore } from '@/stores'

defineProps<{
  contentKey?: string
}>()

const appContent = useAppContentStore()
const { t: navigationT } = useLocale('appShell.unsavedNavigation')
const { t: notificationT } = useLocale('notifications')
const loading = ref(false)
const loadingMore = ref(false)
const deletingId = ref(0)
const errorMessage = ref('')
const notifications = ref<InAppNotification[]>([])
const selectedNotification = ref<InAppNotification | null>(null)
const pendingDelete = ref<InAppNotification | null>(null)
const detailVisible = ref(false)
const deleteConfirmVisible = ref(false)
const mobile = ref(resolveMobile())
const page = ref(1)
const pageSize = 20
const total = ref(0)

const hasMore = computed(() => notifications.value.length < total.value)
const detailPopupWidth = computed(() => mobile.value ? '92%' : '720px')

function resolveMobile() {
  try {
    const systemInfo = uni.getSystemInfoSync()
    const width = Number(systemInfo.windowWidth || systemInfo.screenWidth || 0)
    return width > 0 && width <= 768
  }
  catch {
    return typeof window !== 'undefined' && window.innerWidth <= 768
  }
}

function syncMobile() {
  mobile.value = resolveMobile()
}

function notificationLabel(notification: InAppNotification) {
  if (isFeedbackNotification(notification))
    return String(notification.style?.label || '').trim() || notificationT('feedbackStatus')
  if (isWorkflowNotification(notification))
    return String(notification.style?.label || '').trim() || workflowNotificationLabel(notification)
  return String(notification.style?.label || '').trim() || '系统消息'
}

function notificationIcon(notification: InAppNotification) {
  return String(notification.style?.icon || '').trim() || 'email'
}

function notificationTone(notification: InAppNotification) {
  const tone = String(notification.style?.tone || '').trim()
  return ['primary', 'success', 'warning', 'danger', 'info'].includes(tone) ? tone : 'info'
}

function notificationToneColor(notification: InAppNotification) {
  return {
    primary: '#2563eb',
    success: '#00875a',
    warning: '#b45309',
    danger: '#c93756',
    info: '#475569',
  }[notificationTone(notification)] || '#475569'
}

function notificationPreview(content: string) {
  const firstLine = String(content || '')
    .split(/\r?\n/)
    .map(line => line.trim())
    .find(Boolean)
  return firstLine || '暂无内容'
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
    const response = await listNotifications(nextPage, pageSize)
    const payload = response?.data
    const items = Array.isArray(payload?.list) ? payload.list : []
    notifications.value = append ? [...notifications.value, ...items] : items
    total.value = Math.max(0, Number(payload?.total || 0))
    page.value = nextPage
  }
  catch {
    errorMessage.value = '站内信历史加载失败'
  }
  finally {
    loading.value = false
    loadingMore.value = false
  }
}

async function markRead(notification: InAppNotification, options: { requireExplicitSuccess?: boolean } = {}) {
  if (notification.isRead === 1)
    return true
  return runNotificationMarkRead({
    requireExplicitSuccess: options.requireExplicitSuccess,
    request: () => markNotificationRead(notification.id),
    commit: () => {
      notification.isRead = 1
      appContent.requestRefresh()
    },
    fail: () => uni.showToast({ title: '标记已读失败', icon: 'none' }),
  })
}

function confirmNavigationDiscard() {
  return confirmUnsavedNavigation({
    title: navigationT('title'),
    content: navigationT('content'),
    confirm: navigationT('confirm'),
    cancel: navigationT('cancel'),
  })
}

function openNotificationTab(tab: FeedbackNotificationDynamicTab) {
  return navigateWithUnsavedGuard({
    activeKey: appContent.currentKey,
    targetKey: tab.key,
    hasUnsavedChanges: key => appContent.hasUnsavedTabChanges(key),
    confirmLeave: () => confirmNavigationDiscard(),
    navigate: () => appContent.openDynamicTab(tab),
  })
}

function closeFeedbackMessage() {
  detailVisible.value = false
  selectedNotification.value = null
}

function handleFeedbackNotification(notification: InAppNotification) {
  selectedNotification.value = notification
  detailVisible.value = true
  void openFeedbackNotification(notification, {
    markRead: () => markRead(notification, { requireExplicitSuccess: true }),
    feedbackDetailContentKey: sourceID => feedbackDetailContentKey(sourceID),
    openDynamicTab: tab => openNotificationTab(tab),
    closePanel: closeFeedbackMessage,
    fallbackLabel: notificationT('feedbackDetail'),
  })
}

function handleWorkflowNotification(notification: InAppNotification) {
  selectedNotification.value = notification
  detailVisible.value = true
  void openWorkflowNotification(notification, {
    markRead: () => markRead(notification, { requireExplicitSuccess: true }),
    workflowInstanceContentKey,
    workflowFormRevisionDetailContentKey,
    openDynamicTab: tab => openNotificationTab(tab),
    closePanel: closeFeedbackMessage,
  })
}

function openNotification(notification: InAppNotification) {
  if (isFeedbackNotification(notification)) {
    handleFeedbackNotification(notification)
    return
  }
  if (isWorkflowNotification(notification)) {
    handleWorkflowNotification(notification)
    return
  }
  void markRead(notification)
  selectedNotification.value = notification
  detailVisible.value = true
}

function requestDelete(notification: InAppNotification) {
  if (deletingId.value)
    return
  pendingDelete.value = notification
  deleteConfirmVisible.value = true
}

function cancelDelete() {
  pendingDelete.value = null
}

async function confirmDelete() {
  const notification = pendingDelete.value
  pendingDelete.value = null
  if (!notification || deletingId.value)
    return
  deletingId.value = notification.id
  try {
    await deleteNotification(notification.id)
    detailVisible.value = false
    selectedNotification.value = null
    await loadNotifications(false)
    appContent.requestRefresh()
    uni.showToast({ title: '已删除', icon: 'success' })
  }
  catch {
    uni.showToast({ title: '删除失败', icon: 'none' })
  }
  finally {
    deletingId.value = 0
  }
}

onMounted(() => {
  syncMobile()
  void loadNotifications(false)
  if (typeof window !== 'undefined')
    window.addEventListener('resize', syncMobile)
})

onBeforeUnmount(() => {
  if (typeof window !== 'undefined')
    window.removeEventListener('resize', syncMobile)
})
</script>

<template>
  <view class="notification-history app-pc-control-scope">
    <view class="notification-history__head">
      <view>
        <text class="notification-history__title">
          站内信历史
        </text>
        <text class="notification-history__count">
          共 {{ total }} 条
        </text>
      </view>
    </view>

    <view v-if="loading" class="notification-history__state">
      <u-loading mode="circle" size="36px" />
      <text>加载中...</text>
    </view>
    <view v-else-if="errorMessage" class="notification-history__state">
      <u-icon name="error-circle" size="38px" color="#86909c" />
      <text>{{ errorMessage }}</text>
      <u-button custom-class="notification-history__state-action" @click="loadNotifications(false)">
        重新加载
      </u-button>
    </view>
    <view v-else-if="notifications.length === 0" class="notification-history__state">
      <u-icon name="email" size="44px" color="#c9cdd4" />
      <text class="notification-history__empty-title">
        暂无历史消息
      </text>
    </view>
    <view v-else class="notification-history__list">
      <view class="notification-history__columns">
        <text>消息</text>
        <text>状态</text>
        <text>时间</text>
        <text>操作</text>
      </view>
      <view
        v-for="item in notifications"
        :key="item.id"
        class="notification-history__item"
        :class="{ 'notification-history__item--unread': item.isRead !== 1 }"
      >
        <view class="notification-history__message" @click="openNotification(item)">
          <view class="notification-history__icon">
            <u-icon :name="notificationIcon(item)" size="20px" :color="notificationToneColor(item)" />
          </view>
          <view class="notification-history__message-copy">
            <view class="notification-history__message-title-row">
              <text class="notification-history__message-title">
                {{ item.title || '站内消息' }}
              </text>
              <text class="notification-history__kind">
                {{ notificationLabel(item) }}
              </text>
            </view>
            <text class="notification-history__summary">
              {{ notificationPreview(item.content) }}
            </text>
          </view>
        </view>
        <view class="notification-history__status">
          <u-tag
            :text="item.isRead === 1 ? '已读' : '未读'"
            :type="item.isRead === 1 ? 'info' : 'success'"
            mode="light"
            size="mini"
          />
        </view>
        <text class="notification-history__time">
          {{ formatTime(item.addTime) }}
        </text>
        <view class="notification-history__actions">
          <u-button
            custom-class="notification-history__action"
            title="查看消息"
            @click="openNotification(item)"
          >
            <u-icon name="eye" size="16px" color="#2563eb" />
            <text>查看</text>
          </u-button>
          <u-button
            custom-class="notification-history__action notification-history__action--danger"
            title="删除消息"
            :disabled="deletingId > 0"
            @click="requestDelete(item)"
          >
            <u-icon name="trash" size="16px" color="#dc2626" />
            <text>删除</text>
          </u-button>
        </view>
      </view>
      <view class="notification-history__footer">
        <u-button
          v-if="hasMore"
          custom-class="notification-history__load-more"
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

    <u-popup
      v-model="detailVisible"
      mode="center"
      :width="detailPopupWidth"
      custom-class="notification-history-detail-popup app-pc-control-scope"
      :z-index="10140"
      :border-radius="6"
      :mask-close-able="true"
    >
      <view v-if="selectedNotification" class="notification-history-detail">
        <view class="notification-history-detail__head">
          <view>
            <text class="notification-history-detail__kind">
              {{ notificationLabel(selectedNotification) }}
            </text>
            <text class="notification-history-detail__title">
              {{ selectedNotification.title || '站内消息' }}
            </text>
          </view>
          <u-button
            custom-class="notification-history-detail__close app-icon-button"
            title="关闭消息详情"
            @click="detailVisible = false"
          >
            <u-icon name="close" size="20px" color="#4e5969" />
          </u-button>
        </view>
        <text class="notification-history-detail__time">
          {{ formatTime(selectedNotification.addTime) }}
        </text>
        <scroll-view scroll-y class="notification-history-detail__body">
          <text>{{ selectedNotification.content || '暂无内容' }}</text>
        </scroll-view>
      </view>
    </u-popup>

    <u-modal
      v-model="deleteConfirmVisible"
      custom-class="app-pc-control-scope"
      title="删除站内信"
      content="确认删除这条站内信吗？删除后将不再显示，是否继续？"
      confirm-text="删除"
      cancel-text="取消"
      :show-cancel-button="true"
      :mask-close-able="false"
      :z-index="10160"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
  </view>
</template>

<style scoped lang="scss">
.notification-history {
  min-height: 100%;
  padding: 24px;
  background: #f5f7fa;
  color: #1f2329;
}

.notification-history__head {
  max-width: 1120px;
  min-height: 56px;
  margin: 0 auto;
  border-bottom: 1px solid #dfe3e8;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.notification-history__title {
  font-size: 18px;
  font-weight: 700;
}

.notification-history__count,
.notification-history__summary,
.notification-history__time,
.notification-history__footer,
.notification-history-detail__time {
  color: #86909c;
  font-size: 12px;
}

.notification-history__count {
  margin-left: 10px;
}

.notification-history__state {
  max-width: 1120px;
  min-height: 360px;
  margin: 16px auto 0;
  border: 1px solid #e5e6eb;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background: #fff;
  color: #86909c;
}

.notification-history__empty-title {
  color: #4e5969;
  font-size: 14px;
}

.notification-history__state-action,
.notification-history__load-more,
:deep(.notification-history__state-action),
:deep(.notification-history__load-more) {
  width: auto;
  min-width: 88px;
  height: 32px;
  min-height: 32px;
  margin: 0;
  padding: 0 14px;
  border-radius: 4px;
  box-shadow: none;
  font-size: 13px;
}

.notification-history__list {
  max-width: 1120px;
  margin: 16px auto 0;
  border: 1px solid #e5e6eb;
  overflow: hidden;
  background: #fff;
}

.notification-history__columns,
.notification-history__item {
  display: grid;
  grid-template-columns: minmax(320px, 1fr) 88px 168px 148px;
  align-items: center;
}

.notification-history__columns {
  min-height: 44px;
  padding: 0 16px;
  border-bottom: 1px solid #e5e6eb;
  background: #f7f8fa;
  color: #646a73;
  font-size: 13px;
  font-weight: 600;
}

.notification-history__item {
  min-height: 82px;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f1f2;
  overflow: hidden;
}

.notification-history__item--unread {
  box-shadow: inset 3px 0 0 #00a67e;
  background: #f5fcfa;
}

.notification-history__message {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.notification-history__icon {
  width: 36px;
  height: 36px;
  flex: 0 0 36px;
  border: 1px solid #dfe3e8;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
}

.notification-history__message-copy {
  width: 100%;
  min-width: 0;
  overflow: hidden;
}

.notification-history__message-title-row {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 8px;
}

.notification-history__message-title {
  min-width: 0;
  overflow: hidden;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-history__kind {
  flex: 0 0 auto;
  color: #646a73;
  font-size: 11px;
}

.notification-history__summary {
  width: 100%;
  max-width: 100%;
  margin-top: 6px;
  overflow: hidden;
  display: block;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-history__columns > :not(:first-child),
.notification-history__status,
.notification-history__time,
.notification-history__actions {
  padding-left: 12px;
}

.notification-history__actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.notification-history__action,
:deep(.notification-history__action) {
  width: auto;
  min-width: 58px;
  height: 30px;
  min-height: 30px;
  margin: 0;
  padding: 0 7px;
  border: 1px solid #bfdbfe;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  background: #fff;
  color: #2563eb;
  box-shadow: none;
  font-size: 12px;
}

.notification-history__action--danger,
:deep(.notification-history__action--danger) {
  border-color: #fecaca;
  color: #dc2626;
}

.notification-history__footer {
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.notification-history-detail {
  width: 100%;
  height: min(78vh, 680px);
  max-height: calc(100vh - 32px);
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
}

.notification-history-detail__head {
  flex: 0 0 auto;
  min-height: 72px;
  padding: 14px 16px;
  box-sizing: border-box;
  border-bottom: 1px solid #e5e6eb;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.notification-history-detail__head > view:first-child {
  min-width: 0;
}

.notification-history-detail__kind {
  display: block;
  color: #008f72;
  font-size: 12px;
}

.notification-history-detail__title {
  margin-top: 4px;
  display: block;
  overflow-wrap: anywhere;
  font-size: 18px;
  font-weight: 700;
  line-height: 26px;
}

.notification-history-detail__close,
:deep(.notification-history-detail__close) {
  width: 32px;
  height: 32px;
  min-height: 32px;
  margin: 0;
  padding: 0;
  border: 1px solid #e5e6eb;
  border-radius: 4px;
  background: #fff;
  box-shadow: none;
}

.notification-history-detail__time {
  flex: 0 0 auto;
  padding: 12px 16px 0;
}

.notification-history-detail__body {
  min-height: 0;
  max-height: none;
  height: 100%;
  flex: 1 1 auto;
  padding: 16px;
  box-sizing: border-box;
  color: #4e5969;
  font-size: 14px;
  line-height: 24px;
  white-space: pre-wrap;
}

@media (max-width: 768px) {
  .notification-history {
    padding: 16px 12px;
  }

  .notification-history__head {
    min-height: 48px;
  }

  .notification-history__columns,
  .notification-history__summary,
  .notification-history__kind {
    display: none;
  }

  .notification-history__list {
    margin-top: 12px;
    border: 0;
    overflow: visible;
    background: transparent;
  }

  .notification-history__item {
    min-height: 96px;
    margin-bottom: 10px;
    padding: 12px;
    border: 1px solid #e5e6eb;
    border-radius: 6px;
    grid-template-columns: 1fr auto;
    grid-template-areas:
      "message message"
      "status time"
      "actions actions";
    row-gap: 9px;
    background: #fff;
  }

  .notification-history__item--unread {
    box-shadow: inset 3px 0 0 #00a67e;
    background: #f5fcfa;
  }

  .notification-history__message {
    grid-area: message;
  }

  .notification-history__icon {
    width: 40px;
    height: 40px;
    flex: 0 0 40px;
    border-radius: 7px;
  }

  .notification-history__status {
    grid-area: status;
    padding-left: 52px;
  }

  .notification-history__time {
    grid-area: time;
    padding-left: 0;
    text-align: right;
  }

  .notification-history__actions {
    grid-area: actions;
    padding-left: 52px;
    justify-content: flex-end;
  }

  .notification-history__action,
  :deep(.notification-history__action) {
    min-width: 66px;
    height: 32px;
    min-height: 32px;
    padding: 0 9px;
    gap: 4px;
  }

  .notification-history-detail {
    max-height: calc(100vh - 24px);
  }
}
</style>
