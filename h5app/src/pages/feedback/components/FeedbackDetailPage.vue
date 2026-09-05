<script setup lang="ts">
import type { FeedbackDraftImage, FeedbackDraftValidationError } from './feedback-editor-state'
import type { UserFeedbackDetail, UserFeedbackStatus } from '@/api/user-feedback'
import { useLocale } from 'uview-pro'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getUserFeedbackDetail, supplementUserFeedback } from '@/api/user-feedback'
import { useAppContentStore } from '@/stores'
import { feedbackDetailIdFromContentKey } from '../feedback-route-keys'
import { feedbackStatusMeta } from '../feedback-status'
import {
  canSupplementFeedback,
  createStableFeedbackRequestIdState,
  feedbackDetailIsInaccessible,
  feedbackDraftHasContent,
  validateFeedbackDraft,
  validFeedbackDraftImages,
} from './feedback-editor-state'
import FeedbackImagePicker from './FeedbackImagePicker.vue'
import FeedbackTimeline from './FeedbackTimeline.vue'

const props = defineProps<{
  contentKey: string
}>()

const appContent = useAppContentStore()
const { currentLocale, t } = useLocale('feedback')
const detail = ref<UserFeedbackDetail | null>(null)
const loading = ref(false)
const loadError = ref(false)
const inaccessible = ref(false)
const supplementContent = ref('')
const supplementImages = ref<FeedbackDraftImage[]>([])
const submitting = ref(false)
const requestIdState = createStableFeedbackRequestIdState()
let loadSequence = 0
let unregisterCloseGuard: (() => void) | undefined

const canSupplement = computed(() => Boolean(
  detail.value && canSupplementFeedback(detail.value.status, detail.value.allowsSupplement),
))
function hasUnsavedChanges() {
  return canSupplement.value && feedbackDraftHasContent(supplementContent.value, supplementImages.value)
}

function registerCloseGuard() {
  unregisterCloseGuard?.()
  unregisterCloseGuard = appContent.registerTabCloseGuard(props.contentKey, { hasUnsavedChanges })
}

function statusLabel(status: UserFeedbackStatus) {
  return t(`statuses.${status}`)
}

function formatTime(value: number) {
  const date = new Date(value)
  if (!value || Number.isNaN(date.getTime()))
    return '-'
  const locale = currentLocale.value?.name === 'en-US' ? 'en-US' : 'zh-CN'
  return date.toLocaleString(locale, { hour12: false })
}

async function loadDetail() {
  const feedbackID = feedbackDetailIdFromContentKey(props.contentKey)
  const sequence = ++loadSequence
  detail.value = null
  loading.value = false
  loadError.value = false
  inaccessible.value = false
  if (!feedbackID) {
    inaccessible.value = true
    return
  }

  loading.value = true
  try {
    const response = await getUserFeedbackDetail(feedbackID)
    if (sequence !== loadSequence)
      return
    if (!response?.data) {
      inaccessible.value = true
      return
    }
    detail.value = response.data
  }
  catch (error) {
    if (sequence === loadSequence) {
      if (feedbackDetailIsInaccessible(error))
        inaccessible.value = true
      else
        loadError.value = true
    }
  }
  finally {
    if (sequence === loadSequence)
      loading.value = false
  }
}

function validationMessage(code: FeedbackDraftValidationError) {
  const keyByCode: Record<Exclude<FeedbackDraftValidationError, ''>, string> = {
    content_required: 'detailPage.contentOrImageRequired',
    content_too_long: 'detailPage.contentTooLong',
    content_or_image_required: 'detailPage.contentOrImageRequired',
    images_too_many: 'detailPage.imagesTooMany',
    image_invalid: 'detailPage.imageInvalid',
  }
  return code ? t(keyByCode[code]) : ''
}

async function submitSupplement() {
  if (submitting.value || !detail.value || !canSupplement.value)
    return
  const validationError = validateFeedbackDraft('supplement', supplementContent.value, supplementImages.value)
  if (validationError) {
    uni.showToast({ title: validationMessage(validationError), icon: 'none' })
    return
  }

  const feedbackID = feedbackDetailIdFromContentKey(props.contentKey)
  if (!feedbackID)
    return
  submitting.value = true
  try {
    const response = await supplementUserFeedback(feedbackID, {
      content: supplementContent.value.trim(),
      requestId: requestIdState.current(),
      version: detail.value.version,
      images: validFeedbackDraftImages(supplementImages.value),
    })
    if (!response?.data)
      throw new Error('empty feedback response')
    handleSupplementSuccess(response.data)
  }
  catch {
    uni.showToast({ title: t('detailPage.supplementFailed'), icon: 'none' })
  }
  finally {
    submitting.value = false
  }
}

function handleSupplementSuccess(nextDetail: UserFeedbackDetail) {
  detail.value = nextDetail
  supplementContent.value = ''
  supplementImages.value = []
  requestIdState.rotate()
  unregisterCloseGuard?.()
  unregisterCloseGuard = undefined
  registerCloseGuard()
  appContent.requestRefresh()
  uni.showToast({ title: t('detailPage.supplementSuccess'), icon: 'success' })
}

function closeDetail() {
  appContent.requestCloseTab(props.contentKey)
}

watch(
  () => [props.contentKey, appContent.currentKey, appContent.refreshTick] as const,
  () => {
    if (appContent.currentKey === props.contentKey)
      void loadDetail()
  },
  { immediate: true },
)

onMounted(() => {
  registerCloseGuard()
})

onBeforeUnmount(() => {
  loadSequence += 1
  unregisterCloseGuard?.()
  unregisterCloseGuard = undefined
})
</script>

<template>
  <view class="feedback-detail-page app-pc-control-scope">
    <view v-if="loading" class="feedback-detail-page__state">
      <u-loading mode="circle" size="24px" />
      <text>{{ t('loading') }}</text>
    </view>

    <view v-else-if="inaccessible" class="feedback-detail-page__state">
      <u-icon name="lock" size="32px" color="var(--app-text-muted-color)" />
      <text class="feedback-detail-page__state-title">
        {{ t('detailPage.unavailable') }}
      </text>
    </view>

    <view v-else-if="loadError" class="feedback-detail-page__state">
      <u-icon name="error-circle" size="32px" color="var(--app-text-muted-color)" />
      <text class="feedback-detail-page__state-title">
        {{ t('detailPage.loadFailed') }}
      </text>
      <u-button custom-class="feedback-detail-page__retry" size="small" plain @click="loadDetail">
        <view class="feedback-button-content">
          <u-icon name="reload" size="14px" color="var(--u-type-primary)" />
          <text>{{ t('detailPage.retry') }}</text>
        </view>
      </u-button>
    </view>

    <template v-else-if="detail">
      <view class="feedback-detail-header">
        <view class="feedback-detail-header__copy">
          <view class="feedback-detail-header__title-row">
            <text class="feedback-detail-header__title">
              {{ detail.feedbackNo }}
            </text>
            <u-tag
              :text="statusLabel(detail.status)"
              :type="feedbackStatusMeta(detail.status).type"
              mode="light"
              size="mini"
            />
          </view>
          <view class="feedback-detail-header__meta">
            <text>{{ t('detailPage.createdAt') }}: {{ formatTime(detail.createdAt) }}</text>
            <text>{{ t('detailPage.lastUpdated') }}: {{ formatTime(detail.lastActivityAt || detail.updatedAt) }}</text>
            <text>{{ t('detailPage.handler') }}: {{ detail.handlerName || t('detailPage.noHandler') }}</text>
          </view>
        </view>
        <u-button custom-class="feedback-detail-page__close" size="small" plain @click="closeDetail">
          <u-icon name="close" size="16px" color="var(--app-text-secondary-color)" />
        </u-button>
      </view>

      <view class="feedback-detail-body">
        <view class="feedback-detail-section">
          <view class="feedback-detail-section__header">
            <text class="feedback-detail-section__title">
              {{ t('detailPage.timelineTitle') }}
            </text>
          </view>
          <FeedbackTimeline :messages="detail.messages" />
        </view>

        <view class="feedback-detail-section feedback-detail-section--supplement">
          <view class="feedback-detail-section__header">
            <text class="feedback-detail-section__title">
              {{ t('detailPage.supplementTitle') }}
            </text>
            <u-tag
              v-if="!canSupplement"
              :text="t('detailPage.readOnly')"
              type="info"
              mode="light"
              size="mini"
            />
          </view>

          <template v-if="canSupplement">
            <u-textarea
              v-model="supplementContent"
              custom-class="feedback-detail-page__textarea"
              :count="true"
              :maxlength="5000"
              :placeholder="t('detailPage.supplementPlaceholder')"
              :disabled="submitting"
              height="180rpx"
            />
            <view class="feedback-detail-page__images">
              <FeedbackImagePicker v-model="supplementImages" :disabled="submitting" />
            </view>
            <view class="feedback-detail-page__actions">
              <u-button
                custom-class="feedback-detail-page__submit"
                type="primary"
                :loading="submitting"
                :disabled="submitting"
                @click="submitSupplement"
              >
                <view class="feedback-button-content">
                  <u-icon name="plus-circle" size="16px" color="#fff" />
                  <text>{{ t('detailPage.submitSupplement') }}</text>
                </view>
              </u-button>
            </view>
          </template>
          <text v-else class="feedback-detail-page__readonly">
            {{ t('detailPage.readOnly') }}
          </text>
        </view>
      </view>
    </template>
  </view>
</template>

<style lang="scss" scoped>
.feedback-detail-page {
  min-width: 0;
}

.feedback-detail-page__state {
  min-height: 420px;
  padding: 36px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--app-text-muted-color);
  font-size: 13px;
  text-align: center;
}

.feedback-detail-page__state-title {
  color: var(--app-text-color);
  font-size: 15px;
  font-weight: 600;
}

.feedback-detail-header {
  padding: 18px 20px;
  border-bottom: 1px solid var(--app-border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: var(--app-surface-color);
}

.feedback-detail-header__copy {
  min-width: 0;
}

.feedback-detail-header__title-row,
.feedback-detail-header__meta,
.feedback-detail-section__header {
  display: flex;
  align-items: center;
}

.feedback-detail-header__title-row {
  min-width: 0;
  gap: 8px;
}

.feedback-detail-header__title {
  min-width: 0;
  color: var(--app-text-color);
  font-size: 18px;
  font-weight: 700;
  overflow-wrap: anywhere;
}

.feedback-detail-header__meta {
  margin-top: 6px;
  flex-wrap: wrap;
  gap: 5px 16px;
  color: var(--app-text-muted-color);
  font-size: 12px;
}

.feedback-detail-body {
  max-width: 980px;
  padding: 20px;
}

.feedback-detail-section {
  min-width: 0;
}

.feedback-detail-section + .feedback-detail-section {
  margin-top: 24px;
  padding-top: 22px;
  border-top: 1px solid var(--app-border-color);
}

.feedback-detail-section__header {
  margin-bottom: 16px;
  justify-content: space-between;
  gap: 12px;
}

.feedback-detail-section__title {
  color: var(--app-text-color);
  font-size: 16px;
  font-weight: 700;
}

.feedback-detail-page__images {
  margin-top: 16px;
}

.feedback-detail-page__actions {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.feedback-detail-page__submit,
.feedback-detail-page__retry,
.feedback-detail-page__close,
:deep(.feedback-detail-page__submit),
:deep(.feedback-detail-page__retry),
:deep(.feedback-detail-page__close) {
  width: auto;
  min-width: 96px;
  margin: 0;
  border-radius: 4px;
}

.feedback-detail-page__close,
:deep(.feedback-detail-page__close) {
  min-width: 36px;
  width: 36px;
  padding: 0;
}

.feedback-button-content {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
}

.feedback-detail-page__readonly {
  color: var(--app-text-muted-color);
  font-size: 13px;
  line-height: 1.6;
}

@media screen and (max-width: 768px) {
  .feedback-detail-header,
  .feedback-detail-body {
    padding-right: 16px;
    padding-left: 16px;
  }

  .feedback-detail-header {
    align-items: flex-start;
  }

  .feedback-detail-page__submit,
  :deep(.feedback-detail-page__submit) {
    width: 100%;
  }
}
</style>
