<template>
  <view class="container">
    <view class="login-box">
      <text class="title">管理后台</text>
      <text class="subtitle">WeCheckin Administrator</text>

      <view class="form">
        <view class="input-group">
          <text class="input-label">账号</text>
          <input
            class="input"
            v-model="name"
            placeholder="请输入管理员账号"
            placeholder-class="placeholder"
          />
        </view>
        <view class="input-group">
          <text class="input-label">密码</text>
          <input
            class="input"
            v-model="pwd"
            type="password"
            placeholder="请输入密码"
            placeholder-class="placeholder"
          />
        </view>

        <button class="login-btn" @click="handleLogin">登 录</button>

        <view class="back-link" @click="goBack">
          <text>返回用户端</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../api/admin'

export default {
  data() {
    return {
      name: '',
      pwd: ''
    }
  },

  onLoad() {
    uni.removeStorageSync('admin_token')
    uni.removeStorageSync('admin_info')
  },

  methods: {
    async handleLogin() {
      if (!this.name.trim()) {
        uni.showToast({ title: '请输入账号', icon: 'none' })
        return
      }
      if (!this.pwd.trim()) {
        uni.showToast({ title: '请输入密码', icon: 'none' })
        return
      }

      try {
        uni.showLoading({ title: '登录中...' })
        const res = await adminApi.login({ name: this.name.trim(), pwd: this.pwd.trim() })
        uni.hideLoading()

        if (res.data && res.data.token) {
          uni.setStorageSync('admin_token', res.data.token)
          uni.setStorageSync('admin_info', res.data)
          uni.showToast({ title: '登录成功', icon: 'success' })
          uni.redirectTo({ url: '/pages/admin/admin_home' })
        } else {
          uni.showToast({ title: res.msg || '登录失败', icon: 'none' })
        }
      } catch (e) {
        uni.hideLoading()
        console.error('管理员登录失败', e)
      }
    },

    goBack() {
      uni.switchTab({ url: '/pages/my/my_index' })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background: linear-gradient(135deg, #c0392b 0%, #e74c3c 50%, #c0392b 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx;
}

.login-box {
  width: 100%;
  background-color: #fff;
  border-radius: 24rpx;
  padding: 60rpx 50rpx;
}

.title {
  display: block;
  text-align: center;
  font-size: 44rpx;
  font-weight: bold;
  color: #c0392b;
  margin-bottom: 12rpx;
}

.subtitle {
  display: block;
  text-align: center;
  font-size: 24rpx;
  color: #999;
  margin-bottom: 60rpx;
  letter-spacing: 2rpx;
}

.form {
  display: flex;
  flex-direction: column;
}

.input-group {
  margin-bottom: 36rpx;
}

.input-label {
  display: block;
  font-size: 28rpx;
  color: #333;
  margin-bottom: 16rpx;
  font-weight: 500;
}

.input {
  height: 88rpx;
  background-color: #f8f8f8;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333;
}

.placeholder {
  color: #ccc;
}

.login-btn {
  margin-top: 20rpx;
  height: 88rpx;
  line-height: 88rpx;
  background: linear-gradient(135deg, #c0392b 0%, #e74c3c 100%);
  color: #fff;
  font-size: 32rpx;
  border-radius: 12rpx;
  border: none;
}

.login-btn::after {
  border: none;
}

.back-link {
  text-align: center;
  margin-top: 40rpx;
  font-size: 26rpx;
  color: #999;
  text-decoration: underline;
}
</style>
