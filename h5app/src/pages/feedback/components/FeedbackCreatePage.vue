<script setup lang="ts">
import type { FeedbackDraftImage, FeedbackDraftValidationError } from './feedback-editor-state'
import type { UserFeedbackDetail } from '@/api/user-feedback'
import { useLocale } from 'uview-pro'
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { createUserFeedback } from '@/api/user-feedback'
import { useAppContentStore } from '@/stores'
import {
  FEEDBACK_CREATE_CONTENT_KEY,
  feedbackDetailContentKey,
} from '../feedback-route-keys'
import { createFeedbackDynamicTab } from './feedback-center-state'
import {
  createFeedbackComponentLifecycle,
  createStableFeedbackRequestIdState,
  feedbackDraftHasContent,
  runFeedbackAsyncOperation,
  validateFeedbackDraft,
  validFeedbackDraftImages,
} from './feedback-editor-state'
import FeedbackImagePicker from './FeedbackImagePicker.vue'

const props = withDefaults(defineProps<{
  contentKey?: string
}>(), {
  contentKey: FEEDBACK_CREATE_CONTENT_KEY,
})

const appContent = useAppContentStore()
const { t } = useLocale('feedback')
const content = ref('')
const images = ref<FeedbackDraftImage[]>([])
const submitting = ref(false)
const requestIdState = createStableFeedbackRequestIdState()
const componentLifecycle = createFeedbackComponentLifecycle()
let unregisterCloseGuard: (() => void) | undefined

function hasUnsavedChanges() {
  return feedbackDraftHasContent(content.value, images.value)
}

function canClose() {
  return !submitting.value
}

function registerCloseGuard() {
  unregisterCloseGuard?.()
  unregisterCloseGuard = appContent.registerTabCloseGuard(props.contentKey, { canClose, hasUnsavedChanges })
}

function validationMessage(code: FeedbackDraftValidationError) {
  const keyByCode: Record<Exclude<FeedbackDraftValidationError, ''>, string> = {
    content_required: 'createPage.contentRequired',
    content_too_long: 'createPage.contentTooLong',
    content_or_image_required: 'createPage.contentRequired',
    images_too_many: 'createPage.imagesTooMany',
    image_invalid: 'createPage.imageInvalid',
  }
  return code ? t(keyByCode[code]) : ''
}

async function submitFeedback() {
  if (submitting.value)
    return
  const validationError = validateFeedbackDraft('create', content.value, images.value)
  if (validationError) {
    uni.showToast({ title: validationMessage(validationError), icon: 'none' })
    return
  }

  submitting.value = true
  await runFeedbackAsyncOperation({
    lifecycle: componentLifecycle,
    request: async () => {
      const response = await createUserFeedback({
        content: content.value.trim(),
        requestId: requestIdState.current(),
        images: validFeedbackDraftImages(images.value),
      })
      if (!response?.data)
        throw new Error('empty feedback response')
      return response.data
    },
    success: detail => handleCreateSuccess(detail),
    failure: () => uni.showToast({ title: t('createPage.failed'), icon: 'none' }),
    settled: () => {
      submitting.value = false
    },
  })
}

function handleCreateSuccess(detail: UserFeedbackDetail) {
  const detailKey = feedbackDetailContentKey(detail.id)
  const detailTab = createFeedbackDynamicTab(detailKey, detail.feedbackNo || t('detail'), 'chat')
  content.value = ''
  images.value = []
  requestIdState.rotate()
  unregisterCloseGuard?.()
  unregisterCloseGuard = undefined
  appContent.requestRefresh()
  appContent.removeDynamicTab(props.contentKey)
  if (detailTab)
    appContent.openDynamicTab(detailTab)
  uni.showToast({ title: t('createPage.success'), icon: 'success' })
}

function cancelCreate() {
  if (submitting.value)
    return
  appContent.requestCloseTab(props.contentKey)
}

onMounted(() => {
  registerCloseGuard()
})

onBeforeUnmount(() => {
  componentLifecycle.invalidate()
  unregisterCloseGuard?.()
  unregisterCloseGuard = undefined
})
</script>

<template>
  <view class="feedback-create-page app-pc-control-scope">
    <view class="feedback-page-header">
      <view class="feedback-page-header__copy">
        <text class="feedback-page-header__title">
          {{ t('createPage.title') }}
        </text>
        <text class="feedback-page-header__description">
          {{ t('createPage.description') }}
        </text>
      </view>
      <u-button custom-class="feedback-create-page__cancel" size="small" plain :disabled="submitting" @click="cancelCreate">
        <view class="feedback-button-content">
          <u-icon name="close" size="14px" color="var(--app-text-secondary-color)" />
          <text>{{ t('createPage.cancel') }}</text>
        </view>
      </u-button>
    </view>

    <view class="feedback-editor">
      <view class="feedback-editor__field">
        <text class="feedback-editor__label">
          {{ t('createPage.contentLabel') }}
        </text>
        <u-textarea
          v-model="content"
          custom-class="feedback-editor__textarea"
          :count="true"
          :maxlength="5000"
          :placeholder="t('createPage.contentPlaceholder')"
          :disabled="submitting"
          height="220rpx"
        />
      </view>

      <view class="feedback-editor__field">
        <text class="feedback-editor__label">
          {{ t('createPage.imageLabel') }}
        </text>
        <FeedbackImagePicker v-model="images" :disabled="submitting" />
      </view>

      <view class="feedback-editor__actions">
        <u-button custom-class="feedback-create-page__cancel" plain :disabled="submitting" @click="cancelCreate">
          {{ t('createPage.cancel') }}
        </u-button>
        <u-button
          custom-class="feedback-create-page__submit"
          type="primary"
          :loading="submitting"
          :disabled="submitting"
          @click="submitFeedback"
        >
          <view class="feedback-button-content">
            <u-icon name="checkmark" size="16px" color="#fff" />
            <text>{{ t('createPage.submit') }}</text>
          </view>
        </u-button>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.feedback-create-page {
  min-width: 0;
}

.feedback-page-header {
  padding: 18px 20px;
  border-bottom: 1px solid var(--app-border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: var(--app-surface-color);
}

.feedback-page-header__copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.feedback-page-header__title {
  color: var(--app-text-color);
  font-size: 18px;
  font-weight: 700;
}

.feedback-page-header__description {
  color: var(--app-text-muted-color);
  font-size: 12px;
  line-height: 1.5;
}

.feedback-editor {
  width: 100%;
  max-width: 860px;
  margin: 0 auto;
  box-sizing: border-box;
  padding: 22px 20px 28px;
}

.feedback-editor__field + .feedback-editor__field {
  margin-top: 20px;
}

.feedback-editor__label {
  margin-bottom: 8px;
  display: block;
  color: var(--app-text-color);
  font-size: 14px;
  font-weight: 600;
}

.feedback-editor__actions {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.feedback-create-page__cancel,
.feedback-create-page__submit,
:deep(.feedback-create-page__cancel),
:deep(.feedback-create-page__submit) {
  width: auto;
  min-width: 96px;
  margin: 0;
  border-radius: 4px;
}

.feedback-button-content {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
}

@media screen and (max-width: 768px) {
  .feedback-page-header,
  .feedback-editor {
    padding-right: 16px;
    padding-left: 16px;
  }

  .feedback-page-header {
    align-items: flex-start;
  }

  .feedback-page-header > :deep(.feedback-create-page__cancel) {
    display: none;
  }

  .feedback-editor__actions > :deep(.feedback-create-page__cancel),
  .feedback-editor__actions > :deep(.feedback-create-page__submit) {
    flex: 1;
  }
}
</style>
