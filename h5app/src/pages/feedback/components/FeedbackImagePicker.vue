<script setup lang="ts">
import type { FeedbackDraftImage, FeedbackDraftImageError } from './feedback-editor-state'
import { useLocale } from 'uview-pro'
import { computed, ref } from 'vue'
import { normalizeFeedbackChosenImages } from './feedback-editor-state'

const props = withDefaults(defineProps<{
  modelValue: FeedbackDraftImage[]
  disabled?: boolean
  maxCount?: number
}>(), {
  disabled: false,
  maxCount: 6,
})

const emit = defineEmits<{
  'update:modelValue': [value: FeedbackDraftImage[]]
}>()

const { t } = useLocale('feedback.imagePicker')
const batchSequence = ref(0)
const images = computed(() => Array.isArray(props.modelValue) ? props.modelValue : [])
const canAdd = computed(() => !props.disabled && images.value.length < props.maxCount)

function chooseImages() {
  if (!canAdd.value)
    return
  const remaining = props.maxCount - images.value.length
  uni.chooseImage({
    count: Math.min(remaining, 6),
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (result) => {
      batchSequence.value += 1
      const batchID = `feedback-${Date.now().toString(36)}-${batchSequence.value}`
      const chosen = normalizeFeedbackChosenImages(result, batchID)
      emit('update:modelValue', [...images.value, ...chosen].slice(0, props.maxCount))
    },
    fail: (error) => {
      if (!String(error.errMsg || '').toLowerCase().includes('cancel'))
        uni.showToast({ title: t('chooseFailed'), icon: 'none' })
    },
  })
}

function previewImage(index: number) {
  const urls = images.value.map(image => image.filePath).filter(Boolean)
  if (urls.length === 0)
    return
  uni.previewImage({ current: images.value[index]?.filePath, urls })
}

function removeImage(index: number) {
  if (props.disabled)
    return
  const next = [...images.value]
  next.splice(index, 1)
  emit('update:modelValue', next)
}

function imageErrorText(error?: FeedbackDraftImageError) {
  if (error === 'too_large')
    return t('tooLarge')
  if (error === 'unsupported_type')
    return t('unsupportedType')
  return ''
}
</script>

<template>
  <view class="feedback-image-picker">
    <view
      v-for="(image, index) in images"
      :key="image.clientId"
      class="feedback-image-picker__entry"
    >
      <view
        class="feedback-image-picker__item"
        :class="{ 'feedback-image-picker__item--error': image.error }"
        role="button"
        :aria-label="`${t('preview')} ${index + 1}`"
        @click="previewImage(index)"
      >
        <image class="feedback-image-picker__image" :src="image.filePath" mode="aspectFill" />
        <view
          v-if="!disabled"
          class="feedback-image-picker__remove"
          role="button"
          :aria-label="`${t('remove')} ${index + 1}`"
          @click.stop="removeImage(index)"
        >
          <u-icon name="close" size="12px" color="#fff" />
        </view>
      </view>
      <text v-if="image.error" class="feedback-image-picker__error">
        {{ imageErrorText(image.error) }}
      </text>
    </view>

    <view
      v-if="canAdd"
      class="feedback-image-picker__add"
      role="button"
      :aria-label="t('add')"
      @click="chooseImages"
    >
      <u-icon name="camera" size="22px" color="var(--app-text-muted-color)" />
      <text>{{ t('add') }}</text>
    </view>
  </view>
  <text class="feedback-image-picker__hint">
    {{ t('limitHint') }}
  </text>
</template>

<style lang="scss" scoped>
.feedback-image-picker {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 10px;
}

.feedback-image-picker__entry {
  width: 72px;
  min-width: 0;
}

.feedback-image-picker__item,
.feedback-image-picker__add {
  width: 72px;
  height: 72px;
  border-radius: 4px;
  box-sizing: border-box;
  overflow: hidden;
}

.feedback-image-picker__item {
  position: relative;
  border: 1px solid var(--app-border-color);
  background: var(--app-surface-subtle-color);
  cursor: pointer;
}

.feedback-image-picker__item--error {
  border-color: var(--u-type-error);
}

.feedback-image-picker__image {
  width: 100%;
  height: 100%;
  display: block;
}

.feedback-image-picker__remove {
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

.feedback-image-picker__add {
  border: 1px dashed var(--app-border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: var(--app-text-muted-color);
  background: var(--app-surface-subtle-color);
  font-size: 12px;
  cursor: pointer;
}

.feedback-image-picker__error {
  width: 100%;
  margin-top: 4px;
  display: block;
  color: var(--u-type-error);
  font-size: 11px;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.feedback-image-picker__hint {
  margin-top: 8px;
  display: block;
  color: var(--app-text-muted-color);
  font-size: 12px;
  line-height: 1.5;
}

@media screen and (max-width: 768px) {
  .feedback-image-picker__entry,
  .feedback-image-picker__item,
  .feedback-image-picker__add {
    width: 64px;
  }

  .feedback-image-picker__item,
  .feedback-image-picker__add {
    height: 64px;
  }
}
</style>
