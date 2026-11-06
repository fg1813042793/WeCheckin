<script setup lang="ts">
import { reactive, ref } from 'vue'
import { changePassword } from '@/api/dingtalk-h5/profile'

const emit = defineEmits<{
  completed: []
}>()

const saving = ref(false)
const form = reactive({ currentPassword: '', newPassword: '', confirmPassword: '' })

async function submit() {
  if (!form.currentPassword || !form.newPassword) {
    uni.showToast({ title: '请填写当前密码和新密码', icon: 'none' })
    return
  }
  if (form.newPassword !== form.confirmPassword) {
    uni.showToast({ title: '两次输入的新密码不一致', icon: 'none' })
    return
  }
  saving.value = true
  try {
    const result = await changePassword({
      currentPassword: form.currentPassword,
      oldPassword: form.currentPassword,
      newPassword: form.newPassword,
      confirmPassword: form.confirmPassword,
    })
    if (!result)
      return
    uni.showToast({ title: '密码修改成功，请重新登录', icon: 'none' })
    emit('completed')
  }
  finally {
    saving.value = false
  }
}
</script>

<template>
  <view class="password-change-panel app-pc-control-scope">
    <view class="password-change-panel__body">
      <u-icon name="lock" size="52" color="var(--u-type-primary)" />
      <text class="password-change-panel__title">
        请修改登录密码
      </text>
      <text class="password-change-panel__description">
        当前账号使用临时密码或密码已过期，修改完成后才能继续使用。
      </text>
      <view class="password-change-panel__field">
        <text>当前密码</text>
        <u-input v-model="form.currentPassword" type="password" :border="true" placeholder="请输入当前密码" clearable />
      </view>
      <view class="password-change-panel__field">
        <text>新密码</text>
        <u-input v-model="form.newPassword" type="password" :border="true" placeholder="请输入新密码" clearable />
      </view>
      <view class="password-change-panel__field">
        <text>确认新密码</text>
        <u-input v-model="form.confirmPassword" type="password" :border="true" placeholder="请再次输入新密码" clearable />
      </view>
      <u-button type="primary" :loading="saving" @click="submit">
        修改密码
      </u-button>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.password-change-panel {
  min-height: 100vh;
  padding: 48rpx 28rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  background: $u-bg-color;
}

.password-change-panel__body {
  width: min(680rpx, 100%);
  padding: 42rpx 34rpx;
  display: flex;
  flex-direction: column;
  gap: 24rpx;
  border-radius: 12rpx;
  background: $u-bg-white;
  box-sizing: border-box;
}

.password-change-panel__title { color: $u-main-color; font-size: 36rpx; font-weight: 700; }
.password-change-panel__description { color: $u-tips-color; font-size: 25rpx; line-height: 1.6; }
.password-change-panel__field { display: flex; flex-direction: column; gap: 10rpx; color: $u-content-color; font-size: 26rpx; }
</style>
