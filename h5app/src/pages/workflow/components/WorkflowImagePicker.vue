<script setup lang="ts">
import type { WorkflowAttachment } from '@/types/workflow'
import { computed, ref } from 'vue'
import { buildApiUrl } from '@/api/dingtalk-h5/base'
import { uploadWorkflowAttachment } from '@/api/workflow'

const props = withDefaults(defineProps<{
  modelValue: WorkflowAttachment[]
  readonly?: boolean
  disabled?: boolean
  maxCount?: number
}>(), {
  readonly: false,
  disabled: false,
  maxCount: 9,
})

const emit = defineEmits<{
  'update:modelValue': [value: WorkflowAttachment[]]
  'uploading-change': [uploading: boolean]
}>()

const uploading = ref(false)
const images = computed(() => Array.isArray(props.modelValue) ? props.modelValue : [])
const canAdd = computed(() => !props.readonly && !props.disabled && !uploading.value && images.value.length < props.maxCount)

function chooseImages() {
  if (!canAdd.value)
    return
  const remaining = props.maxCount - images.value.length
  uni.chooseImage({
    count: Math.min(9, remaining),
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (result) => {
      void uploadChosenImages(normalizeChosenImages(result))
    },
    fail: (error) => {
      if (!String(error.errMsg || '').toLowerCase().includes('cancel'))
        uni.showToast({ title: '选择图片失败', icon: 'none' })
    },
  })
}

function normalizeChosenImages(result: unknown) {
  if (!result || typeof result !== 'object' || Array.isArray(result))
    return []
  const record = result as Record<string, unknown>
  const paths = Array.isArray(record.tempFilePaths) ? record.tempFilePaths : []
  const files = Array.isArray(record.tempFiles) ? record.tempFiles : []
  return paths.map((value, index) => {
    const path = String(value || '')
    const file = files[index]
    const size = file && typeof file === 'object' && !Array.isArray(file)
      ? Number((file as Record<string, unknown>).size || 0)
      : 0
    return { path, size }
  }).filter(item => item.path)
}

async function uploadChosenImages(files: Array<{ path: string, size: number }>) {
  if (files.length === 0)
    return
  if (files.some(file => file.size > 20 * 1024 * 1024)) {
    uni.showToast({ title: '单张图片不能超过20MB', icon: 'none' })
    return
  }

  setUploading(true)
  const next = [...images.value]
  try {
    for (const file of files) {
      const response = await uploadWorkflowAttachment(file.path)
      const image = response.data
      if (!image?.url || !image.mimeType?.toLowerCase().startsWith('image/'))
        throw new Error('上传文件不是有效图片')
      if (!next.some(item => item.id === image.id || item.url === image.url))
        next.push(image)
    }
    emit('update:modelValue', next.slice(0, props.maxCount))
  }
  catch (error) {
    uni.showToast({ title: uploadErrorMessage(error), icon: 'none' })
  }
  finally {
    setUploading(false)
  }
}

function setUploading(value: boolean) {
  uploading.value = value
  emit('uploading-change', value)
}

function removeImage(index: number) {
  if (props.readonly || props.disabled || uploading.value)
    return
  const next = [...images.value]
  next.splice(index, 1)
  emit('update:modelValue', next)
}

function previewImage(index: number) {
  const urls = images.value.map(item => buildApiUrl(item.url))
  if (urls.length === 0)
    return
  uni.previewImage({ current: urls[index], urls })
}

function uploadErrorMessage(error: unknown) {
  if (!error || typeof error !== 'object' || Array.isArray(error))
    return '图片上传失败'
  const record = error as Record<string, unknown>
  return String(record.msg || record.message || '').trim() || '图片上传失败'
}
</script>

<template>
  <view v-if="images.length > 0 || !readonly" class="workflow-image-picker">
    <view
      v-for="(image, index) in images"
      :key="image.id || `${image.url}-${index}`"
      class="workflow-image-picker__item"
      role="button"
      :aria-label="`预览图片${index + 1}`"
      @click="previewImage(index)"
    >
      <image class="workflow-image-picker__image" :src="buildApiUrl(image.url)" mode="aspectFill" />
      <view
        v-if="!readonly"
        class="workflow-image-picker__remove"
        role="button"
        :aria-label="`移除图片${index + 1}`"
        @click.stop="removeImage(index)"
      >
        <u-icon name="close" size="12px" color="#fff" />
      </view>
    </view>

    <view
      v-if="canAdd"
      class="workflow-image-picker__add"
      role="button"
      aria-label="上传图片"
      @click="chooseImages"
    >
      <u-icon name="plus" size="22px" color="#86909c" />
      <text>图片</text>
    </view>
    <view v-else-if="uploading" class="workflow-image-picker__add workflow-image-picker__add--disabled">
      <u-loading mode="circle" size="20px" />
      <text>上传中</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.workflow-image-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.workflow-image-picker__item,
.workflow-image-picker__add {
  position: relative;
  width: 72px;
  height: 72px;
  border-radius: 4px;
  overflow: hidden;
  box-sizing: border-box;
}

.workflow-image-picker__item {
  background: #f2f3f5;
  cursor: pointer;
}

.workflow-image-picker__image {
  width: 100%;
  height: 100%;
  display: block;
}

.workflow-image-picker__remove {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(31, 35, 41, 0.72);
}

.workflow-image-picker__add {
  border: 1px dashed #c9cdd4;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  color: #86909c;
  font-size: 12px;
  background: #fafbfc;
  cursor: pointer;
}

.workflow-image-picker__add--disabled {
  cursor: default;
}

@media screen and (max-width: 768px) {
  .workflow-image-picker__item,
  .workflow-image-picker__add {
    width: 64px;
    height: 64px;
  }
}
</style>
