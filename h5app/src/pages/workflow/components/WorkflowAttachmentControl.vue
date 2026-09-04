<script setup lang="ts">
import type { WorkflowAttachment, WorkflowFormField } from '@/types/workflow'
import { computed, ref } from 'vue'
import { buildApiUrl } from '@/api/dingtalk-h5/base'
import { uploadWorkflowAttachment } from '@/api/workflow'
import { normalizeWorkflowAttachments } from '../workflow-form'

const props = withDefaults(defineProps<{
  field: WorkflowFormField
  modelValue: unknown
  readonly?: boolean
}>(), {
  readonly: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: WorkflowAttachment[]]
}>()

const uploading = ref(false)
const attachments = computed(() => normalizeWorkflowAttachments(props.modelValue))
const maximumCount = computed(() => {
  const maximum = (props.field.rules || [])
    .filter(rule => rule.type === 'selection_count')
    .map(rule => Number(rule.max))
    .find(value => Number.isInteger(value) && value > 0)
  return maximum || 20
})

function chooseAttachments() {
  if (props.readonly || uploading.value)
    return
  const remaining = maximumCount.value - attachments.value.length
  if (remaining <= 0) {
    uni.showToast({ title: `最多上传${maximumCount.value}个附件`, icon: 'none' })
    return
  }
  uni.chooseFile({
    count: Math.min(9, remaining),
    success: (result) => {
      const files = Array.isArray(result.tempFiles) ? result.tempFiles : []
      void uploadChosenFiles(files.map(normalizeChosenFile))
    },
    fail: (error) => {
      if (!String(error.errMsg || '').toLowerCase().includes('cancel'))
        uni.showToast({ title: '选择附件失败', icon: 'none' })
    },
  })
}

function normalizeChosenFile(file: unknown) {
  if (!file || typeof file !== 'object' || Array.isArray(file))
    return { path: '', name: '', size: 0 }
  const record = file as Record<string, unknown>
  const path = String(record.path || record.tempFilePath || '')
  return {
    path,
    name: String(record.name || '').trim() || path.slice(path.lastIndexOf('/') + 1),
    size: Number(record.size || 0),
  }
}

async function uploadChosenFiles(files: Array<{ path: string, name: string, size: number }>) {
  const selected = files.filter(file => file.path)
  if (selected.length === 0)
    return
  const oversized = selected.find(file => file.size > 20 * 1024 * 1024)
  if (oversized) {
    uni.showToast({ title: `${oversized.name || '附件'}超过20MB`, icon: 'none' })
    return
  }
  uploading.value = true
  const next = [...attachments.value]
  try {
    for (const file of selected) {
      const response = await uploadWorkflowAttachment(file.path)
      const attachment = response.data
      if (!attachment?.url)
        throw new Error('附件上传响应缺少地址')
      if (!next.some(item => item.id === attachment.id || item.url === attachment.url))
        next.push(attachment)
    }
    emit('update:modelValue', next)
    uni.showToast({ title: '附件已上传', icon: 'success' })
  }
  catch (error) {
    uni.showToast({ title: uploadErrorMessage(error), icon: 'none' })
  }
  finally {
    uploading.value = false
  }
}

function removeAttachment(index: number) {
  if (props.readonly || uploading.value)
    return
  const next = [...attachments.value]
  next.splice(index, 1)
  emit('update:modelValue', next)
}

function previewAttachment(attachment: WorkflowAttachment) {
  const url = buildApiUrl(attachment.url)
  if (isImageAttachment(attachment)) {
    const imageUrls = attachments.value.filter(isImageAttachment).map(item => buildApiUrl(item.url))
    uni.previewImage({ current: url, urls: imageUrls })
    return
  }

  // #ifdef H5
  const opened = window.open(url, '_blank', 'noopener,noreferrer')
  if (!opened)
    uni.showToast({ title: '浏览器阻止了附件窗口', icon: 'none' })
  // #endif

  // #ifndef H5
  openDownloadedAttachment(url)
  // #endif
}

function downloadAttachment(attachment: WorkflowAttachment) {
  const url = buildApiUrl(attachment.url)
  // #ifdef H5
  const link = document.createElement('a')
  link.href = url
  link.download = attachment.name
  link.rel = 'noopener noreferrer'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  // #endif

  // #ifndef H5
  openDownloadedAttachment(url)
  // #endif
}

function openDownloadedAttachment(url: string) {
  uni.downloadFile({
    url,
    success: (result) => {
      if (result.statusCode !== 200) {
        uni.showToast({ title: '附件下载失败', icon: 'none' })
        return
      }
      uni.openDocument({
        filePath: result.tempFilePath,
        showMenu: true,
        fail: () => uni.showToast({ title: '当前设备无法打开此附件', icon: 'none' }),
      })
    },
    fail: () => uni.showToast({ title: '附件下载失败', icon: 'none' }),
  })
}

function isImageAttachment(attachment: WorkflowAttachment) {
  if (attachment.mimeType.toLowerCase().startsWith('image/'))
    return true
  return /\.(?:jpe?g|png|gif|webp)(?:[?#].*)?$/i.test(attachment.url)
}

function formatAttachmentSize(size: number) {
  if (!Number.isFinite(size) || size <= 0)
    return ''
  if (size < 1024)
    return `${size} B`
  if (size < 1024 * 1024)
    return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

function uploadErrorMessage(error: unknown) {
  if (!error || typeof error !== 'object' || Array.isArray(error))
    return '附件上传失败'
  const record = error as Record<string, unknown>
  return String(record.msg || record.message || '').trim() || '附件上传失败'
}
</script>

<template>
  <view class="workflow-attachment">
    <view v-if="attachments.length > 0" class="workflow-attachment__list">
      <view
        v-for="(attachment, index) in attachments"
        :key="attachment.id || `${attachment.url}-${index}`"
        class="workflow-attachment__item"
      >
        <view class="workflow-attachment__file" role="button" :aria-label="`预览${attachment.name}`" @click="previewAttachment(attachment)">
          <view class="workflow-attachment__icon">
            <u-icon :name="isImageAttachment(attachment) ? 'photo' : 'file-text'" size="28" color="#0f766e" />
          </view>
          <view class="workflow-attachment__meta">
            <text class="workflow-attachment__name">
              {{ attachment.name }}
            </text>
            <text v-if="formatAttachmentSize(attachment.size)" class="workflow-attachment__size">
              {{ formatAttachmentSize(attachment.size) }}
            </text>
          </view>
        </view>
        <view class="workflow-attachment__actions">
          <view class="app-icon-button app-icon-button--small" role="button" :aria-label="`下载${attachment.name}`" @click.stop="downloadAttachment(attachment)">
            <u-icon name="download" size="28" color="#2979ff" />
          </view>
          <view v-if="!readonly" class="app-icon-button app-icon-button--small" role="button" :aria-label="`移除${attachment.name}`" @click.stop="removeAttachment(index)">
            <u-icon name="trash" size="28" color="#fa3534" />
          </view>
        </view>
      </view>
    </view>
    <view v-else-if="readonly" class="workflow-attachment__empty">
      暂无附件
    </view>
    <u-button
      v-if="!readonly"
      custom-class="workflow-attachment__upload"
      size="small"
      plain
      :loading="uploading"
      :disabled="uploading"
      @click="chooseAttachments"
    >
      <u-icon v-if="!uploading" name="plus" size="24" color="#2979ff" />
      <text>
        {{ uploading ? '上传中' : '上传附件' }}
      </text>
    </u-button>
  </view>
</template>

<style lang="scss" scoped>
.workflow-attachment {
  width: 100%;
}

.workflow-attachment__list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-bottom: 14rpx;
}

.workflow-attachment__item {
  min-height: 78rpx;
  padding: 12rpx 16rpx;
  border: 1px solid $u-border-color;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  background: #fff;
  box-sizing: border-box;
}

.workflow-attachment__file {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  gap: 14rpx;
  cursor: pointer;
}

.workflow-attachment__icon {
  width: 52rpx;
  height: 52rpx;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 118, 110, 0.08);
  flex: 0 0 auto;
}

.workflow-attachment__meta {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.workflow-attachment__name {
  overflow-wrap: anywhere;
  color: $u-main-color;
  font-size: 26rpx;
  line-height: 1.45;
}

.workflow-attachment__size {
  color: $u-tips-color;
  font-size: 22rpx;
}

.workflow-attachment__actions {
  min-width: 74rpx;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 22rpx;
  flex: 0 0 auto;
}

.workflow-attachment__actions > view {
  width: 42rpx;
  height: 42rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.workflow-attachment__empty {
  min-height: 72rpx;
  padding: 0 22rpx;
  border: 1px solid $u-border-color;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  color: $u-content-color;
  font-size: 26rpx;
  box-sizing: border-box;
}

:deep(.workflow-attachment__upload) {
  width: auto;
  min-width: 190rpx;
  margin: 0;
}

:deep(.workflow-attachment__upload .u-button__text) {
  display: inline-flex;
  align-items: center;
  gap: 8rpx;
}
</style>
