<template>
  <view class="container">
    <view class="logo-area">
      <image src="/static/logo.png" mode="aspectFit" class="logo"></image>
      <text class="app-name">MY打卡</text>
      <text class="app-desc">使用钉钉验证身份后直接登录</text>
    </view>

    <view class="form-area">
      <button class="login-btn" :loading="loading" :disabled="loading" @click="handleDingTalkLogin">
        钉钉登录
      </button>

      <view class="agreement">
        <checkbox :checked="agreed" @click="agreed = !agreed" color="#1677ff" />
        <text class="agree-text">我已阅读并同意</text>
        <view class="agree-link" @click="goAgreement">《用户协议》</view>
        <text class="agree-text">和</text>
        <view class="agree-link" @click="goPrivacy">《隐私政策》</view>
      </view>

      <view v-if="configLoaded && !dingtalkEnabled" class="login-notice">
        钉钉登录尚未配置，请联系管理员
      </view>
    </view>

    <view class="other-login" @click="goPasswordLogin">
      <text class="other-text">账号密码登录</text>
    </view>

    <view class="admin-login" @click="handleAdminLogin">
      <text class="admin-login-text">管理员登录</text>
    </view>
  </view>
</template>

<script>
import { passportApi } from '../../api/index'
import { setClientAuth } from '../../utils/auth'
import { ensureClientPermissionSnapshot } from '../../utils/clientPermission'
import {
  beginDingTalkOAuth,
  clearDingTalkOAuthAttempt,
  getDingTalkNativeRedirectURI,
  getDingTalkOAuthRedirectURI,
  prepareDingTalkOAuthAttempt,
  readDingTalkOAuthCallback,
  validateDingTalkOAuthResult
} from '../../utils/dingtalkOAuth'
// #ifdef APP-PLUS
import {
  consumeDingTalkAuthResult,
  startDingTalkNativeAuth
} from '@/uni_modules/wecheckin-dingtalk-auth'
// #endif

export default {
  data() {
    return {
      loading: false,
      callbackHandling: false,
      configLoaded: false,
      dingtalkEnabled: false,
      corpOptions: [],
      agreed: false,
      loadOptions: {}
    }
  },

  async onLoad(options = {}) {
    this.loadOptions = options
    // #ifdef APP-PLUS
    await this.completeDingTalkNativeAuth()
    // #endif
    // #ifdef H5
    await this.completeDingTalkOAuth()
    // #endif
    await this.loadDingTalkConfig()
  },

  async onShow() {
    // #ifdef APP-PLUS
    if (await this.completeDingTalkNativeAuth()) return
    // #endif
    // #ifdef H5
    await this.completeDingTalkOAuth()
    // #endif
  },

  methods: {
    async loadDingTalkConfig() {
      if (this.configLoaded) return
      try {
        const res = await passportApi.dingtalkLoginConfig()
        const data = res.data || {}
        this.corpOptions = Array.isArray(data.corpOptions) ? data.corpOptions : []
        this.dingtalkEnabled = data.enabled === true && this.corpOptions.length > 0
      } catch (e) {
        this.dingtalkEnabled = false
        this.corpOptions = []
      } finally {
        this.configLoaded = true
      }
    },

    async handleDingTalkLogin() {
      if (!this.agreed) {
        uni.showToast({ title: '请先同意用户协议', icon: 'none' })
        return
      }
      await this.loadDingTalkConfig()
      if (!this.dingtalkEnabled) {
        uni.showToast({ title: '钉钉登录尚未配置，请联系管理员', icon: 'none' })
        return
      }
      const corp = await this.selectCorp()
      if (!corp) return
      let redirectUri = ''
      // #ifdef APP-PLUS
      redirectUri = getDingTalkNativeRedirectURI()
      // #endif
      // #ifdef H5
      redirectUri = getDingTalkOAuthRedirectURI()
      // #endif
      if (!redirectUri) {
        uni.showToast({ title: '当前平台未配置钉钉登录回调地址', icon: 'none' })
        return
      }

      this.loading = true
      try {
        const res = await passportApi.dingtalkAuthorization({ corpId: corp.corpId, redirectUri })
        const responseData = res && res.data && typeof res.data === 'object' ? res.data : {}
        const data = responseData.clientId || responseData.redirectUri || responseData.state
          ? responseData
          : (responseData.data || {})

        // #ifdef APP-PLUS
        const clientId = String(data.clientId || '').trim()
        const nativeRedirectUri = String(data.redirectUri || '').trim()
        const state = String(data.state || '').trim()
        if (!clientId || !nativeRedirectUri || !state) {
          this.loading = false
          uni.showToast({ title: '钉钉授权参数不完整，请重启后端服务', icon: 'none' })
          return
        }
        prepareDingTalkOAuthAttempt({ state, corpId: corp.corpId })
        const nativeResult = startDingTalkNativeAuth({
          clientId,
          redirectUri: nativeRedirectUri,
          state,
          scope: 'openid',
          nonce: state
        })
        this.loading = false
        if (!nativeResult || nativeResult.started !== true) {
          clearDingTalkOAuthAttempt()
          uni.showToast({ title: nativeResult?.error || '无法启动钉钉授权', icon: 'none' })
        }
        return
        // #endif

        // #ifdef H5
        beginDingTalkOAuth({
          authorizationUrl: data.authorizationUrl,
          state: data.state,
          corpId: corp.corpId
        })
        return
        // #endif

        this.loading = false
        uni.showToast({ title: '当前平台暂不支持钉钉原生登录', icon: 'none' })
      } catch (e) {
        this.loading = false
        if (e && e.message) {
          uni.showToast({ title: e.message, icon: 'none' })
        }
      }
    },

    selectCorp() {
      if (this.corpOptions.length === 1) {
        return Promise.resolve(this.corpOptions[0])
      }
      return new Promise((resolve) => {
        uni.showActionSheet({
          itemList: this.corpOptions.map((item) => item.corpName || item.corpId),
          success: ({ tapIndex }) => resolve(this.corpOptions[tapIndex] || null),
          fail: () => resolve(null)
        })
      })
    },

    async completeDingTalkOAuth() {
      if (this.callbackHandling) return
      let callback
      try {
        callback = readDingTalkOAuthCallback(this.loadOptions)
      } catch (e) {
        clearDingTalkOAuthAttempt()
        uni.showToast({ title: e.message || '钉钉登录失败', icon: 'none' })
        return
      }
      if (!callback) return

      await this.finishDingTalkLogin(callback)
    },

    async completeDingTalkNativeAuth() {
      // #ifdef APP-PLUS
      if (this.callbackHandling) return false
      const result = consumeDingTalkAuthResult()
      if (!result) return false

      try {
        const nativeError = this.dingTalkNativeErrorMessage(result.error)
        const callback = validateDingTalkOAuthResult({
          authCode: result.authCode,
          state: result.state,
          error: nativeError,
          errorDescription: nativeError
        })
        if (callback) await this.finishDingTalkLogin(callback)
      } catch (e) {
        clearDingTalkOAuthAttempt()
        uni.showToast({ title: e.message || '钉钉登录失败', icon: 'none' })
      }
      return true
      // #endif

      return false
    },

    dingTalkNativeErrorMessage(error) {
      const code = String(error || '').trim()
      if (!code || code === '0' || code === 'ERR_OK') return ''
      if (code === '-2' || code === 'ERR_USER_CANCEL') return '已取消钉钉授权'
      if (code === '-4' || code === 'ERR_AUTH_DENIED') return '钉钉授权未通过'
      return '钉钉授权失败'
    },

    async finishDingTalkLogin(callback) {
      if (this.callbackHandling) return

      this.callbackHandling = true
      this.loading = true
      try {
        const res = await passportApi.loginByDingTalk(callback)
        if (!res.data || !res.data.token) {
          throw new Error('钉钉登录结果异常')
        }
        setClientAuth(res.data)
        await ensureClientPermissionSnapshot()
        clearDingTalkOAuthAttempt()
        uni.showToast({ title: '登录成功', icon: 'success' })
        setTimeout(() => {
          uni.switchTab({ url: '/pages/index/index' })
        }, 500)
      } catch (e) {
        clearDingTalkOAuthAttempt()
        if (e && e.message) {
          uni.showToast({ title: e.message, icon: 'none' })
        }
      } finally {
        this.callbackHandling = false
        this.loading = false
      }
    },

    goPasswordLogin() {
      uni.navigateTo({ url: '/pages/login/login_pwd' })
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
  box-sizing: border-box;
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
.app-desc {
  font-size: 28rpx;
  color: #777;
}
.form-area {
  padding: 0 20rpx;
}
.login-btn {
  margin-top: 60rpx;
  width: 100%;
  height: 96rpx;
  line-height: 96rpx;
  border-radius: 48rpx;
  background-color: #1677ff;
  color: #fff;
  font-size: 32rpx;
}
.login-btn[disabled] {
  opacity: 0.65;
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
.agree-text,
.agree-link {
  font-size: 24rpx;
}
.agree-text {
  color: #777;
}
.agree-link {
  color: #1677ff;
}
.login-notice {
  margin-top: 30rpx;
  color: #909399;
  font-size: 24rpx;
  text-align: center;
}
.other-login {
  margin-top: 100rpx;
  text-align: center;
}
.other-text,
.admin-login-text {
  display: inline;
  padding-bottom: 4rpx;
  font-size: 26rpx;
  border-bottom: 1rpx solid currentColor;
}
.other-text {
  color: #666;
}
.admin-login {
  text-align: center;
  margin-top: 40rpx;
  padding-bottom: 40rpx;
}
.admin-login-text {
  color: #aaa;
}
</style>
