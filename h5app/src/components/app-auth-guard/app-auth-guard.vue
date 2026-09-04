<script setup lang="ts">
import type { LoginRequest } from '@/types/dingtalk-h5'
import { computed, onMounted, ref, watch } from 'vue'
import DingtalkBindPanel from '@/components/dingtalk-bind-panel/dingtalk-bind-panel.vue'
import DingtalkLoginPanel from '@/components/dingtalk-login-panel/dingtalk-login-panel.vue'
import { useDingtalkAuthStore } from '@/stores'

const emit = defineEmits<{
  authenticated: []
}>()

const auth = useDingtalkAuthStore()
const initializing = ref(false)

const userId = computed(() => auth.user?.id || '')

async function ensureReady() {
  if (auth.ready) {
    if (auth.user) {
      emit('authenticated')
    }
    return
  }
  if (initializing.value) {
    return
  }
  initializing.value = true
  try {
    await auth.init()
    if (auth.user) {
      emit('authenticated')
    }
  }
  finally {
    initializing.value = false
  }
}

async function handleLogin(form: LoginRequest) {
  const loggedIn = await auth.loginWithPassword(form)
  if (loggedIn) {
    emit('authenticated')
  }
}

async function handleBind(account: string, password: string) {
  const loggedIn = await auth.bindDingTalkUser(account, password)
  if (loggedIn) {
    emit('authenticated')
  }
}

async function handleRetryAutoLogin() {
  auth.resetSession()
  const loggedIn = await auth.tryDingTalkAutoLogin()
  if (loggedIn) {
    emit('authenticated')
  }
}

onMounted(() => {
  ensureReady()
})

watch(userId, (nextUserId) => {
  if (nextUserId) {
    emit('authenticated')
  }
})
</script>

<template>
  <app-page hide-nav>
    <view v-if="!auth.ready" class="loading-page">
      <u-loading mode="circle" />
      <text>加载中...</text>
    </view>

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
      :loading="auth.loading"
      @login="handleLogin"
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
