<template>
  <view class="password-page">
    <view class="password-panel">
      <text class="title">请修改登录密码</text>
      <text class="description">当前账号使用临时密码或密码已过期，修改后才能继续使用。</text>
      <view class="form-item">
        <text class="label">当前密码</text>
        <input v-model="form.currentPassword" class="input" type="password" placeholder="请输入当前密码" />
      </view>
      <view class="form-item">
        <text class="label">新密码</text>
        <input v-model="form.newPassword" class="input" type="password" placeholder="请输入新密码" />
      </view>
      <view class="form-item">
        <text class="label">确认新密码</text>
        <input v-model="form.confirmPassword" class="input" type="password" placeholder="请再次输入新密码" />
      </view>
      <button class="submit" :loading="saving" @click="submit">修改密码</button>
    </view>
  </view>
</template>

<script>
import { passportApi } from '../../api/index'
import { clearClientAuth, getClientToken } from '../../utils/auth'

export default {
  data() {
    return { form: { currentPassword: '', newPassword: '', confirmPassword: '' }, saving: false }
  },
  onLoad() {
    if (!getClientToken()) uni.reLaunch({ url: '/pages/login/login_pwd' })
  },
  methods: {
    async submit() {
      if (!this.form.currentPassword || !this.form.newPassword) {
        uni.showToast({ title: '请填写当前密码和新密码', icon: 'none' })
        return
      }
      if (this.form.newPassword !== this.form.confirmPassword) {
        uni.showToast({ title: '两次输入的新密码不一致', icon: 'none' })
        return
      }
      this.saving = true
      try {
        await passportApi.changePassword(this.form)
        clearClientAuth()
        uni.showToast({ title: '密码修改成功，请重新登录', icon: 'none' })
        setTimeout(() => uni.reLaunch({ url: '/pages/login/login_pwd' }), 600)
      } finally {
        this.saving = false
      }
    }
  }
}
</script>

<style scoped>
.password-page { min-height: 100vh; padding: 48rpx 32rpx; background: #f5f7fb; box-sizing: border-box; }
.password-panel { padding: 36rpx 30rpx; border-radius: 12rpx; background: #fff; }
.title { display: block; color: #20242a; font-size: 40rpx; font-weight: 600; }
.description { display: block; margin: 16rpx 0 36rpx; color: #7b8490; font-size: 26rpx; line-height: 1.6; }
.form-item { margin-bottom: 26rpx; }
.label { display: block; margin-bottom: 12rpx; color: #4a5563; font-size: 28rpx; }
.input { height: 80rpx; padding: 0 24rpx; border: 1px solid #d9e0e8; border-radius: 8rpx; box-sizing: border-box; }
.submit { margin-top: 16rpx; color: #fff; background: #2f7bdc; }
</style>
