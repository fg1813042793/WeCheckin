<script setup lang="ts">
import { computed } from 'vue'
import type { UserFeedbackMessage, UserFeedbackMessageType } from '@/types/userFeedback'
import { userFeedbackStatusMeta } from '../userFeedbackStatus'

const props = defineProps<{
  messages: UserFeedbackMessage[]
  formatTime: (timestamp: number) => string
}>()

const orderedMessages = computed(() => [...props.messages].sort((left, right) => left.createdAt - right.createdAt))

function messageTypeLabel(type: UserFeedbackMessageType) {
  switch (type) {
    case 'initial': return '提交反馈'
    case 'supplement': return '用户补充'
    case 'status': return '状态更新'
  }
}

function timelineType(type: UserFeedbackMessageType) {
  if (type === 'status') return 'primary'
  if (type === 'initial') return 'success'
  return 'warning'
}

function authorLabel(message: UserFeedbackMessage) {
  if (message.authorName) return message.authorName
  return `${message.authorType === 'admin' ? '管理员' : '用户'} #${message.authorId}`
}

function previewSources(message: UserFeedbackMessage) {
  return message.attachments.map(attachment => attachment.url).filter(Boolean)
}
</script>

<template>
  <el-empty v-if="orderedMessages.length === 0" description="暂无流转记录" :image-size="72" />
  <el-timeline v-else class="feedback-timeline">
    <el-timeline-item
      v-for="message in orderedMessages"
      :key="message.id"
      :timestamp="formatTime(message.createdAt)"
      :type="timelineType(message.messageType)"
      placement="top"
    >
      <div class="feedback-timeline__heading">
        <strong>{{ messageTypeLabel(message.messageType) }}</strong>
        <span>{{ authorLabel(message) }}</span>
      </div>
      <div v-if="message.messageType === 'status' && message.toStatus" class="feedback-timeline__transition">
        <el-tag v-if="message.fromStatus" :type="userFeedbackStatusMeta(message.fromStatus).type" size="small">
          {{ userFeedbackStatusMeta(message.fromStatus).label }}
        </el-tag>
        <span aria-hidden="true">→</span>
        <el-tag :type="userFeedbackStatusMeta(message.toStatus).type" size="small">
          {{ userFeedbackStatusMeta(message.toStatus).label }}
        </el-tag>
      </div>
      <p v-if="message.content" class="feedback-timeline__content">{{ message.content }}</p>
      <p v-else-if="message.attachments.length === 0" class="feedback-timeline__empty">暂无文字内容</p>
      <div v-if="message.attachments.length > 0" class="feedback-timeline__images">
        <el-image
          v-for="(attachment, imageIndex) in message.attachments"
          :key="attachment.id"
          :src="attachment.url"
          :alt="attachment.originalName || `反馈图片 ${imageIndex + 1}`"
          :preview-src-list="previewSources(message)"
          :initial-index="imageIndex"
          fit="cover"
          preview-teleported
        />
      </div>
    </el-timeline-item>
  </el-timeline>
</template>

<style scoped>
.feedback-timeline {
  margin: 0;
  padding-left: var(--admin-ui-space-2);
}

.feedback-timeline__heading,
.feedback-timeline__transition {
  display: flex;
  align-items: center;
}

.feedback-timeline__heading {
  flex-wrap: wrap;
  gap: var(--admin-ui-space-2);
  min-width: 0;
}

.feedback-timeline__heading strong {
  color: var(--admin-ui-color-text);
  font-size: var(--admin-ui-font-size-body);
}

.feedback-timeline__heading span,
.feedback-timeline__empty {
  color: var(--admin-ui-color-text-muted);
  font-size: var(--admin-ui-font-size-caption);
}

.feedback-timeline__transition {
  gap: var(--admin-ui-space-2);
  margin-top: var(--admin-ui-space-2);
}

.feedback-timeline__content {
  margin: var(--admin-ui-space-2) 0 0;
  color: var(--admin-ui-color-text-secondary);
  line-height: 1.7;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.feedback-timeline__empty {
  margin: var(--admin-ui-space-2) 0 0;
}

.feedback-timeline__images {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(88px, 112px));
  gap: var(--admin-ui-space-2);
  margin-top: var(--admin-ui-space-3);
}

.feedback-timeline__images :deep(.el-image) {
  width: 100%;
  aspect-ratio: 1;
  overflow: hidden;
  border: 1px solid var(--admin-ui-color-border);
  border-radius: var(--admin-ui-radius-sm);
  background: var(--admin-ui-color-surface-muted);
}
</style>
