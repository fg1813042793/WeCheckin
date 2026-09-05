<script setup lang="ts">
import type { UserFeedbackAttachment, UserFeedbackMessage, UserFeedbackStatus } from '@/api/user-feedback'
import { useLocale } from 'uview-pro'
import { computed } from 'vue'
import { buildApiUrl } from '@/api/dingtalk-h5/base'
import { feedbackStatusMeta } from '../feedback-status'
import { sortFeedbackMessages } from './feedback-editor-state'

const props = defineProps<{
  messages: UserFeedbackMessage[]
}>()

const { currentLocale, t } = useLocale('feedback')
const sortedMessages = computed(() => sortFeedbackMessages(props.messages || []))

function messageTitle(message: UserFeedbackMessage) {
  return t(`timeline.${message.messageType}`)
}

function authorName(message: UserFeedbackMessage) {
  return message.authorName || t(`timeline.${message.authorType}`)
}

function statusLabel(status?: UserFeedbackStatus) {
  return status ? t(`statuses.${status}`) : '-'
}

function formatTime(value: number) {
  const date = new Date(value)
  if (!value || Number.isNaN(date.getTime()))
    return '-'
  const locale = currentLocale.value?.name === 'en-US' ? 'en-US' : 'zh-CN'
  return date.toLocaleString(locale, { hour12: false })
}

function attachmentUrls(attachments: UserFeedbackAttachment[]) {
  return (attachments || []).map(attachment => buildApiUrl(attachment.url)).filter(Boolean)
}

function previewAttachments(message: UserFeedbackMessage, index: number) {
  const urls = attachmentUrls(message.attachments)
  if (urls.length === 0)
    return
  uni.previewImage({ current: urls[index], urls })
}
</script>

<template>
  <view class="feedback-timeline">
    <view v-if="sortedMessages.length === 0" class="feedback-timeline__empty">
      <u-icon name="order" size="28px" color="var(--app-text-muted-color)" />
      <text>{{ t('timeline.empty') }}</text>
    </view>

    <view
      v-for="message in sortedMessages"
      v-else
      :key="String(message.id)"
      class="feedback-timeline__item"
    >
      <view class="feedback-timeline__rail">
        <view
          class="feedback-timeline__dot"
          :class="`feedback-timeline__dot--${message.messageType}`"
        />
      </view>
      <view class="feedback-timeline__content">
        <view class="feedback-timeline__header">
          <view class="feedback-timeline__heading">
            <text class="feedback-timeline__title">
              {{ messageTitle(message) }}
            </text>
            <u-tag
              v-if="message.messageType === 'status' && message.toStatus"
              :text="statusLabel(message.toStatus)"
              :type="feedbackStatusMeta(message.toStatus).type"
              mode="light"
              size="mini"
            />
          </view>
          <text class="feedback-timeline__time">
            {{ formatTime(message.createdAt) }}
          </text>
        </view>

        <view class="feedback-timeline__meta">
          <u-icon name="account" size="13px" color="var(--app-text-muted-color)" />
          <text>{{ authorName(message) }}</text>
        </view>

        <view v-if="message.messageType === 'status'" class="feedback-timeline__status-change">
          <text class="feedback-timeline__status-label">
            {{ t('timeline.statusChange') }}
          </text>
          <text>{{ statusLabel(message.fromStatus) }}</text>
          <u-icon name="arrow-right" size="13px" color="var(--app-text-muted-color)" />
          <text>{{ statusLabel(message.toStatus) }}</text>
        </view>

        <text v-if="message.content" class="feedback-timeline__message">
          {{ message.content }}
        </text>

        <view v-if="message.attachments?.length" class="feedback-timeline__images">
          <image
            v-for="(attachment, index) in message.attachments"
            :key="String(attachment.id)"
            class="feedback-timeline__image"
            :src="buildApiUrl(attachment.url)"
            mode="aspectFill"
            :aria-label="`${t('imagePicker.preview')} ${index + 1}`"
            @click="previewAttachments(message, index)"
          />
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.feedback-timeline {
  min-width: 0;
}

.feedback-timeline__empty {
  min-height: 160px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--app-text-muted-color);
  font-size: 13px;
}

.feedback-timeline__item {
  position: relative;
  min-width: 0;
  display: grid;
  grid-template-columns: 22px minmax(0, 1fr);
}

.feedback-timeline__rail {
  position: relative;
  display: flex;
  justify-content: center;
}

.feedback-timeline__rail::after {
  position: absolute;
  top: 18px;
  bottom: -4px;
  width: 1px;
  background: var(--app-border-color);
  content: '';
}

.feedback-timeline__item:last-child .feedback-timeline__rail::after {
  display: none;
}

.feedback-timeline__dot {
  position: relative;
  z-index: 1;
  width: 9px;
  height: 9px;
  margin-top: 6px;
  border: 2px solid var(--u-type-primary);
  border-radius: 50%;
  background: var(--app-surface-color);
  box-sizing: border-box;
}

.feedback-timeline__dot--status {
  border-color: var(--u-type-warning);
}

.feedback-timeline__content {
  min-width: 0;
  padding: 0 0 22px 8px;
}

.feedback-timeline__header,
.feedback-timeline__heading,
.feedback-timeline__meta,
.feedback-timeline__status-change {
  display: flex;
  align-items: center;
}

.feedback-timeline__header {
  min-width: 0;
  justify-content: space-between;
  gap: 10px;
}

.feedback-timeline__heading {
  min-width: 0;
  gap: 8px;
}

.feedback-timeline__title {
  color: var(--app-text-color);
  font-size: 14px;
  font-weight: 600;
}

.feedback-timeline__time,
.feedback-timeline__meta {
  color: var(--app-text-muted-color);
  font-size: 12px;
}

.feedback-timeline__time {
  flex-shrink: 0;
}

.feedback-timeline__meta {
  margin-top: 5px;
  gap: 4px;
}

.feedback-timeline__status-change {
  margin-top: 10px;
  gap: 6px;
  color: var(--app-text-secondary-color);
  font-size: 13px;
}

.feedback-timeline__status-label {
  color: var(--app-text-muted-color);
}

.feedback-timeline__message {
  margin-top: 10px;
  display: block;
  color: var(--app-text-secondary-color);
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.feedback-timeline__images {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.feedback-timeline__image {
  width: 68px;
  height: 68px;
  border: 1px solid var(--app-border-color);
  border-radius: 4px;
  background: var(--app-surface-subtle-color);
  cursor: pointer;
}

@media screen and (max-width: 768px) {
  .feedback-timeline__header {
    align-items: flex-start;
    flex-direction: column;
    gap: 4px;
  }
}
</style>
