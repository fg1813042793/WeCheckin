<template>
  <view class="container">
    <view class="logo-area">
      <image src="/static/logo.png" mode="aspectFit" class="logo"></image>
      <text class="app-name">MY打卡</text>
      <text class="app-desc">养成好习惯，记录每一天</text>
    </view>

    <view class="form-area">
      <view class="form-item">
        <input 
          type="text" 
          v-model="form.user_id" 
          placeholder="请输入用户ID"
          class="form-input"
        />
      </view>

      <button class="login-btn" :loading="loading" @click="handleLogin">登 录</button>
      
      <view class="agreement">
        <checkbox :checked="agreed" @click="agreed = !agreed" color="#fb454c" />
        <text class="agree-text">我已阅读并同意</text>
        <view class="agree-link" @click="goAgreement">《用户协议》</view>
        <text class="agree-text">和</text>
        <view class="agree-link" @click="goPrivacy">《隐私政策》</view>
      </view>
    </view>

    <view class="other-login" @click="showOtherMethods">
      <text class="other-text">其他登录方式 ›</text>
    </view>

    <view class="admin-login" @click="handleAdminLogin">
      <text class="admin-login-text">管理员登录</text>
    </view>
  </view>
</template>

<script>
import { passportApi } from '../../api/index'

export default {
  data() {
    return {
      form: {
        user_id: ''
      },
      loading: false,
      agreed: false,
      isIOS: false
    }
  },

  onLoad() {
    const systemInfo = uni.getSystemInfoSync()
    this.isIOS = systemInfo.platform === 'ios'
  },

  methods: {
    async handleLogin() {
      if (!this.agreed) {
        uni.showToast({ title: '请先同意用户协议', icon: 'none' })
        return
      }

      if (!this.form.user_id) {
        uni.showToast({ title: '请输入用户ID', icon: 'none' })
        return
      }

      this.loading = true
      try {
        const res = await passportApi.login(this.form)
        if (res.data) {
          uni.setStorageSync('token', res.data.token || '')
          uni.setStorageSync('userInfo', res.data.userInfo || res.data)
        }
        uni.showToast({ title: '登录成功', icon: 'success' })
        setTimeout(() => {
          uni.switchTab({ url: '/pages/index/index' })
        }, 1500)
      } catch (e) {
        console.error('登录失败', e)
      } finally {
        this.loading = false
      }
    },

    showOtherMethods() {
      const itemList = ['账号密码登录']
      // #ifdef MP-WEIXIN
      itemList.push('微信登录')
      // #endif
      uni.showActionSheet({
        itemList,
        success: (res) => {
          const idx = res.tapIndex
          // #ifdef MP-WEIXIN
          if (idx === 0) {
            uni.navigateTo({ url: '/pages/login/login_pwd' })
          } else if (idx === 1) {
            this.wechatLogin()
          }
          // #endif
          // #ifndef MP-WEIXIN
          if (idx === 0) {
            uni.navigateTo({ url: '/pages/login/login_pwd' })
          }
          // #endif
        }
      })
    },

    wechatLogin() {
      // #ifdef MP-WEIXIN
      wx.login({
        success: async (loginRes) => {
          try {
            const res = await passportApi.login({ user_id: loginRes.code })
            if (res.data) {
              uni.setStorageSync('token', res.data.token || '')
              uni.setStorageSync('userInfo', res.data.userInfo || res.data)
            }
            uni.showToast({ title: '登录成功', icon: 'success' })
            setTimeout(() => {
              uni.switchTab({ url: '/pages/index/index' })
            }, 1500)
          } catch (e) {
            console.error('微信登录失败', e)
          }
        }
      })
      // #endif
    },

    goAgreement() {
      uni.navigateTo({ url: '/pages/about/agreement' })
    },

    goPrivacy() {
      uni.navigateTo({ url: '/pages/about/privacy' })
    },

    handleAdminLogin() {
      uni.navigateTo({ url: '/pages/admin/admin_login' })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #fff;
  padding: 60rpx 40rpx;
}
.logo-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 80rpx;
  margin-bottom: 80rpx;
}
.logo {
  width: 160rpx;
  height: 160rpx;
  border-radius: 32rpx;
  margin-bottom: 30rpx;
}
.app-name {
  font-size: 44rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 16rpx;
}
.app-name {
  font-size: 44rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 16rpx;
}
.app-desc {
  font-size: 28rpx;
  color: #999;
}
.form-area {
  padding: 0 20rpx;
}
.form-item {
  display: flex;
  align-items: center;
  background-color: #f5f5f5;
  border-radius: 16rpx;
  padding: 0 30rpx;
  height: 96rpx;
}
.form-input {
  flex: 1;
  font-size: 30rpx;
  color: #333;
  height: 100%;
}
.login-btn {
  margin-top: 60rpx;
  background-color: #fb454c;
  color: #fff;
  font-size: 32rpx;
  border-radius: 48rpx;
  height: 96rpx;
  line-height: 96rpx;
  width: 100%;
}
.login-btn::after {
  border: none;
}
.agreement {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 30rpx;
  flex-wrap: wrap;
}
.agree-text {
  font-size: 24rpx;
  color: #999;
}
.agree-link {
  font-size: 24rpx;
  color: #fb454c;
  display: inline;
}
.other-login {
  margin-top: 100rpx;
  text-align: center;
}
.other-text {
  font-size: 26rpx;
  color: #999;
  display: inline;
  border-bottom: 1rpx solid #999;
  padding-bottom: 4rpx;
}
.admin-login {
  text-align: center;
  margin-top: 40rpx;
  padding-bottom: 40rpx;
}
.admin-login-text {
  font-size: 26rpx;
  color: #bbb;
  border-bottom: 1rpx solid #bbb;
}
</style>
