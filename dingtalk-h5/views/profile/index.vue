<template>
  <view v-if="dialog.visible" class="profile-center-modal" @click="$emit('close')">
    <view class="profile-center-card" @click.stop>
      <view class="profile-center-head">
        <view>
          <text class="profile-center-title">个人中心</text>
          <text class="profile-center-desc">维护当前登录账号、头像和登录密码。</text>
        </view>
        <button class="process-modal-close" @click="$emit('close')">×</button>
      </view>
      <view class="profile-center-body">
        <view class="profile-center-avatar-row">
          <view
            class="profile-center-avatar-preview"
            :class="{ uploading: dialog.uploading }"
            @click="$emit('choose-avatar')"
          >
            <image v-if="avatarPreview" class="profile-center-avatar-image" :src="avatarPreview" mode="aspectFill" />
            <view v-else class="profile-center-avatar-text">{{ initial }}</view>
            <text class="profile-center-avatar-mask">{{ dialog.uploading ? '上传中' : '更换' }}</text>
          </view>
          <view class="profile-center-avatar-meta">
            <text class="profile-center-avatar-title">{{ displayName }}</text>
            <text class="profile-center-avatar-desc">上传头像后会同步到顶部头像展示。</text>
            <view class="profile-center-avatar-actions">
              <button
                class="dt-btn dt-btn-light small"
                :disabled="dialog.loading || dialog.uploading"
                :loading="dialog.uploading"
                @click="$emit('choose-avatar')"
              >上传头像</button>
              <button
                v-if="dialog.avatar"
                class="dt-btn dt-btn-light small"
                :disabled="dialog.loading || dialog.uploading"
                @click="$emit('clear-avatar')"
              >移除头像</button>
            </view>
            <text v-if="dialog.avatar" class="profile-center-avatar-path">保存后生效</text>
          </view>
        </view>
        <view class="profile-center-form">
          <view class="profile-center-field">
            <text class="profile-center-label">账号</text>
            <input class="field-input" v-model="dialog.account" placeholder="请输入账号" />
          </view>
          <view class="profile-center-field">
            <text class="profile-center-label">当前密码</text>
            <input class="field-input" v-model="dialog.currentPassword" password placeholder="修改账号或密码时必填" />
          </view>
          <view class="profile-center-grid">
            <view class="profile-center-field">
              <text class="profile-center-label">新密码</text>
              <input class="field-input" v-model="dialog.newPassword" password placeholder="不修改可留空" />
            </view>
            <view class="profile-center-field">
              <text class="profile-center-label">确认新密码</text>
              <input class="field-input" v-model="dialog.confirmPassword" password placeholder="再次输入新密码" />
            </view>
          </view>
        </view>
      </view>
      <view class="profile-center-actions">
        <button class="dt-btn dt-btn-light" :disabled="dialog.loading" @click="$emit('close')">取消</button>
        <button class="dt-btn dt-btn-primary" :loading="dialog.loading" @click="$emit('submit')">保存</button>
      </view>
    </view>
  </view>
</template>

<script setup>
defineProps({
  avatarPreview: { type: String, default: '' },
  dialog: { type: Object, required: true },
  displayName: { type: String, default: '当前用户' },
  initial: { type: String, default: 'U' }
})

defineEmits(['choose-avatar', 'clear-avatar', 'close', 'submit'])
</script>

<style>
/* 页面专属样式：从 styles/performance.css 拆分 */
.profile-center-modal {
  position: fixed;
  inset: 0;
  z-index: 126;
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

.profile-center-card {
  width: min(520px, 100%);
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

.profile-center-head {
  padding: 18px 20px 14px;
  border-bottom: 1px solid #f2f3f5;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.profile-center-title,
.profile-center-desc,
.profile-center-label,
.profile-center-avatar-title,
.profile-center-avatar-desc {
  display: block;
}

.profile-center-title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 800;
}

.profile-center-desc,
.profile-center-avatar-desc {
  margin-top: 6px;
  color: #86909c;
  font-size: 13px;
  line-height: 1.6;
}

.profile-center-body {
  min-height: 0;
  padding: 18px 20px;
  display: grid;
  gap: 18px;
  overflow: auto;
}

.profile-center-avatar-row {
  padding: 12px;
  border: 1px solid #e5eaf3;
  border-radius: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
  background: #f8fbff;
}

.profile-center-avatar-preview {
  position: relative;
  width: 58px;
  height: 58px;
  border-radius: 14px;
  flex: 0 0 58px;
  overflow: hidden;
  background: #eef3fb;
  cursor: pointer;
}

.profile-center-avatar-image,
.profile-center-avatar-text {
  width: 100%;
  height: 100%;
  border-radius: 14px;
}

.profile-center-avatar-image {
  display: block;
}

.profile-center-avatar-text {
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1677ff 0%, #4c9aff 100%);
  color: #fff;
  font-size: 18px;
  font-weight: 800;
}

.profile-center-avatar-mask {
  position: absolute;
  inset: auto 0 0;
  padding: 4px 0;
  background: rgba(15, 23, 42, 0.62);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.2;
  text-align: center;
}

.profile-center-avatar-preview.uploading .profile-center-avatar-mask {
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.profile-center-avatar-meta {
  min-width: 0;
  flex: 1 1 auto;
}

.profile-center-avatar-title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 800;
}

.profile-center-avatar-actions {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.profile-center-avatar-path {
  display: block;
  margin-top: 8px;
  color: #86909c;
  font-size: 12px;
}

.profile-center-form {
  display: grid;
  gap: 14px;
}

.profile-center-field {
  min-width: 0;
  display: grid;
  gap: 8px;
}

.profile-center-label {
  color: #4e5969;
  font-size: 13px;
  font-weight: 700;
}

.profile-center-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.profile-center-actions {
  padding: 14px 20px 18px;
  border-top: 1px solid #f2f3f5;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 960px) {
  .profile-center-modal {
    align-items: center;
    justify-content: center;
    padding: 12px;
  }
  .profile-center-card {
    width: 100%;
    max-height: calc(100vh - 32px);
    border-radius: 18px;
  }
  .profile-center-head,
  .profile-center-body,
  .profile-center-actions {
    padding-left: 16px;
    padding-right: 16px;
  }
  .profile-center-grid {
    grid-template-columns: 1fr;
  }
  .profile-center-actions .dt-btn {
    flex: 1 1 0;
  }
}
</style>
