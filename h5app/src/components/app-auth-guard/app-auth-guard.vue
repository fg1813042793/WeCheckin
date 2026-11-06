<script setup lang="ts">
import type { LoginRequest } from '@/types/dingtalk-h5'
import { computed, onBeforeUnmount, onMounted, watch } from 'vue'
import DingtalkBindPanel from '@/components/dingtalk-bind-panel/dingtalk-bind-panel.vue'
import DingtalkLoginPanel from '@/components/dingtalk-login-panel/dingtalk-login-panel.vue'
import PasswordChangePanel from '@/components/password-change-panel/password-change-panel.vue'
import { useDingtalkAuthStore } from '@/stores'

const emit = defineEmits<{
  authenticated: []
}>()

const auth = useDingtalkAuthStore()
const userId = computed(() => auth.user?.id || '')

async function ensureReady() {
  await auth.syncCorpContext()
}

async function handleLogin(form: LoginRequest) {
  await auth.loginWithPassword(form)
}

async function handleBind(account: string, password: string) {
  await auth.bindDingTalkUser(account, password)
}

async function handleRetryAutoLogin() {
  await auth.retryDingTalkLogin()
}

function handleVisibilityChange() {
  // #ifdef H5
  if (document.visibilityState === 'visible')
    void ensureReady()
  // #endif
}

onMounted(() => {
  void ensureReady()
  // #ifdef H5
  window.addEventListener('focus', ensureReady)
  window.addEventListener('pageshow', ensureReady)
  window.addEventListener('hashchange', ensureReady)
  window.addEventListener('popstate', ensureReady)
  document.addEventListener('visibilitychange', handleVisibilityChange)
  // #endif
})

onBeforeUnmount(() => {
  // #ifdef H5
  window.removeEventListener('focus', ensureReady)
  window.removeEventListener('pageshow', ensureReady)
  window.removeEventListener('hashchange', ensureReady)
  window.removeEventListener('popstate', ensureReady)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  // #endif
})

watch(userId, (nextUserId) => {
  if (nextUserId) {
    emit('authenticated')
  }
}, { immediate: true })
</script>

<template>
  <app-page hide-nav>
    <view v-if="!auth.ready" class="loading-page">
      <u-loading mode="circle" />
      <text>加载中...</text>
    </view>

    <PasswordChangePanel
      v-else-if="auth.passwordChangeRequired"
      @completed="auth.completePasswordChange"
    />

    <DingtalkBindPanel
      v-else-if="!auth.user && auth.bindState.visible"
      :app-config="auth.appConfig"
      :bind-state="auth.bindState"
      :loading="auth.loading"
      @bind="handleBind"
      @retry="handleRetryAutoLogin"
    />

    <view v-else-if="!auth.user && auth.sessionAccessDenied" class="loading-page denied-page">
      <u-icon name="lock" size="72" color="var(--u-type-error)" />
      <text class="denied-page__title">
        暂无访问权限
      </text>
      <text class="denied-page__desc">
        {{ auth.sessionAccessDeniedMessage }}
      </text>
      <u-button type="primary" @click="auth.resetSession">
        重新登录
      </u-button>
    </view>

    <DingtalkLoginPanel
      v-else-if="!auth.user"
      :app-config="auth.appConfig"
      :auto-login-message="auth.autoLoginMessage"
      :show-dingtalk-login="auth.canRetryDingTalkLogin"
      :captcha-enabled="auth.captchaEnabled"
      :captcha-id="auth.captchaId"
      :captcha-image="auth.captchaImage"
      :captcha-loading="auth.captchaLoading"
      :captcha-unavailable="auth.captchaUnavailable"
      :loading="auth.loading"
      @refresh-captcha="auth.refreshLoginCaptcha"
      @login="handleLogin"
      @retry-dingtalk="handleRetryAutoLogin"
    />

    <slot v-else />
  </app-page>
</template>

<style lang="scss" scoped>
.loading-page {
  min-height: 100vh;
  padding: 48rpx 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 18rpx;
  background: $u-bg-white;
  color: $u-tips-color;
  font-size: 26rpx;
  text-align: center;
}

.denied-page__title {
  color: $u-main-color;
  font-size: 32rpx;
  font-weight: 700;
}

.denied-page__desc {
  max-width: 560rpx;
  color: $u-content-color;
  font-size: 25rpx;
  line-height: 1.5;
}
</style>
