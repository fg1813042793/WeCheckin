<template>
  <view v-if="dialog.visible" class="review-action-modal" @click="$emit('close')">
    <view class="review-action-card withdraw-confirm-card" @click.stop>
      <view class="review-action-head">
        <view>
          <text class="review-action-title">{{ title }}</text>
          <text class="review-action-desc">{{ desc }}</text>
        </view>
        <button class="process-modal-close" @click="$emit('close')">×</button>
      </view>
      <view class="review-action-body">
        <text class="review-action-label">{{ reasonLabel }}</text>
        <textarea
          class="field-textarea withdraw-reason-textarea"
          v-model="dialog.reason"
          maxlength="200"
          :placeholder="placeholder"
        ></textarea>
        <text class="review-action-helper">{{ reasonLength }}/200</text>
      </view>
      <view class="review-action-actions">
        <button class="dt-btn dt-btn-light" :disabled="dialog.loading" @click="$emit('close')">取消</button>
        <button class="dt-btn dt-btn-danger" :loading="dialog.loading" @click="$emit('submit')">{{ submitText }}</button>
      </view>
    </view>
  </view>
</template>

<script setup>
defineProps({
  desc: { type: String, default: '' },
  dialog: { type: Object, required: true },
  placeholder: { type: String, default: '请填写原因' },
  reasonLabel: { type: String, default: '原因' },
  reasonLength: { type: Number, default: 0 },
  submitText: { type: String, default: '确认' },
  title: { type: String, default: '' }
})

defineEmits(['close', 'submit'])
</script>

<style>
/* 页面专属样式：从 styles/performance.css 拆分 */
.review-action-modal {
  position: fixed;
  inset: 0;
  z-index: 124;
  padding: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.42);
  backdrop-filter: blur(4px);
  overscroll-behavior: contain;
  pointer-events: auto;
  touch-action: none;
}

.review-action-card {
  width: min(460px, 100%);
  border-radius: 16px;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 28px 80px rgba(15, 23, 42, 0.2);
  overscroll-behavior: contain;
  pointer-events: auto;
  touch-action: auto;
}

.review-action-head {
  padding: 18px 20px 14px;
  border-bottom: 1px solid #f2f3f5;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.review-action-title,
.review-action-desc,
.review-action-label,
.review-action-helper {
  display: block;
}

.review-action-title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 800;
}

.review-action-desc {
  margin-top: 6px;
  color: #86909c;
  font-size: 13px;
  line-height: 1.6;
}

.review-action-body {
  padding: 18px 20px;
}

.review-action-label {
  margin-bottom: 8px;
  color: #4e5969;
  font-size: 13px;
  font-weight: 700;
}

.withdraw-reason-textarea {
  min-height: 118px;
  resize: none;
}

.review-action-helper {
  margin-top: 6px;
  color: #86909c;
  font-size: 12px;
  text-align: right;
}

.review-action-actions {
  padding: 14px 20px 18px;
  border-top: 1px solid #f2f3f5;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 960px) {
  .review-action-modal {
    align-items: flex-end;
    padding: 12px;
  }
  .review-action-card {
    width: 100%;
    max-height: calc(100vh - 32px);
    border-radius: 18px;
  }
  .review-action-head,
  .review-action-body,
  .review-action-actions {
    padding-left: 16px;
    padding-right: 16px;
  }
  .review-action-actions .dt-btn {
    flex: 1 1 0;
  }
  .withdraw-reason-textarea {
    min-height: 112px;
  }
}
</style>
