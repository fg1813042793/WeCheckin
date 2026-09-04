<script setup lang="ts">
import type { PerformanceReview, PerformanceTemplate, PerformanceUser, ReviewActionRequest } from '@/types/dingtalk-h5'
import { computed } from 'vue'
import PerformanceReviewDetail from './PerformanceReviewDetail.vue'

type ReviewActionScope = 'mine' | 'manager' | 'hrbp' | 'readonly'

const props = defineProps<{
  actionScope?: ReviewActionScope
  detailLoading?: boolean
  grades?: string[]
  modelValue: boolean
  popupWidth?: string
  review: PerformanceReview | null
  submitReviewAction?: (id: string, action: string, data: ReviewActionRequest) => Promise<PerformanceReview | void>
  template?: PerformanceTemplate | null
  users?: PerformanceUser[]
}>()

const emit = defineEmits<{
  'action-success': [review?: PerformanceReview | void]
  'update:modelValue': [visible: boolean]
}>()

const popupVisible = computed({
  get: () => props.modelValue,
  set: (visible: boolean) => emit('update:modelValue', visible),
})

const reviewTitle = computed(() => {
  if (!props.review) {
    return '绩效详情'
  }
  return `${props.review.period || '-'} 月度考评`
})
const detailPopupWidth = computed(() => props.popupWidth || '86%')

function closePopup() {
  popupVisible.value = false
}
</script>

<template>
  <u-popup
    v-model="popupVisible"
    mode="center"
    :width="detailPopupWidth"
    height="88%"
    custom-class="app-pc-control-scope"
    border-radius="8"
    :mask-close-able="true"
    :safe-area-inset-bottom="true"
  >
    <view class="performance-detail-popup">
      <view class="performance-detail-popup__head">
        <view class="performance-detail-popup__copy">
          <text class="performance-detail-popup__title">
            {{ reviewTitle }}
          </text>
          <text class="performance-detail-popup__desc">
            查看绩效表单、流程进度和评价结果
          </text>
        </view>
        <u-button custom-class="performance-detail-popup__close app-icon-button" @click="closePopup">
          <u-icon name="close" size="18" color="#5f6b7a" />
        </u-button>
      </view>

      <scroll-view class="performance-detail-popup__body" scroll-y>
        <PerformanceReviewDetail
          v-if="review"
          :review="review"
          :users="users"
          :template="template"
          :grades="grades"
          :action-scope="actionScope"
          :detail-loading="detailLoading"
          :submit-review-action="submitReviewAction"
          @action-success="emit('action-success', $event)"
        />
      </scroll-view>
    </view>
  </u-popup>
</template>

<style lang="scss" scoped>
.performance-detail-popup {
  width: 100%;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f7f9fc;
}

.performance-detail-popup__head {
  flex: 0 0 auto;
  min-height: 64px;
  padding: 0 20px 0 24px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: #fff;
}

.performance-detail-popup__copy {
  min-width: 0;
  display: grid;
  gap: 2px;
}

.performance-detail-popup__title,
.performance-detail-popup__desc {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.performance-detail-popup__title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 800;
  line-height: 1.35;
}

.performance-detail-popup__desc {
  color: #86909c;
  font-size: 12px;
}

.performance-detail-popup__close,
:deep(.performance-detail-popup__close) {
  width: 34px;
  height: 34px;
  min-height: 34px;
  margin: 0;
  padding: 0;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #fff;
}

.performance-detail-popup__close::after,
:deep(.performance-detail-popup__close)::after {
  display: none;
}

.performance-detail-popup__body {
  min-height: 0;
  flex: 1 1 auto;
  height: 100%;
  padding: 22px 24px 26px;
  box-sizing: border-box;
}

@media (max-width: 768px), (hover: none) and (pointer: coarse) {
  .performance-detail-popup__head {
    min-height: 56px;
    padding: 0 14px 0 16px;
  }

  .performance-detail-popup__title {
    font-size: 16px;
  }

  .performance-detail-popup__body {
    padding: 14px 12px 18px;
  }
}
</style>
