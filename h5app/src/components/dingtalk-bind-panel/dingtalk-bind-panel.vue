<script setup lang="ts">
import type { AppConfig } from '@/types/dingtalk-h5'
import { reactive } from 'vue'

defineProps<{
  appConfig: AppConfig
  bindState: {
    dingTalkUserIdMasked: string
    unionIdMasked: string
    expiresIn: number
  }
  loading: boolean
}>()

const emit = defineEmits<{
  bind: [account: string, password: string]
  retry: []
}>()

const form = reactive({
  account: '',
  password: '',
})

function submit() {
  if (!form.account.trim() || !form.password.trim()) {
    uni.showToast({
      title: '请输入账号和密码',
      icon: 'none',
    })
    return
  }
  emit('bind', form.account.trim(), form.password)
}
</script>

<template>
  <view class="bind-panel">
    <view class="bind-card">
      <view class="bind-card__head">
        <view class="bind-card__logo">
          {{ appConfig.logoText || 'OA' }}
        </view>
        <text class="bind-card__title">
          绑定系统账号
        </text>
        <text class="bind-card__desc">
          当前钉钉身份还未绑定本地系统账号，绑定后下次可直接免登。
        </text>
      </view>

      <view class="bind-meta">
        <text v-if="bindState.dingTalkUserIdMasked">
          钉钉用户：{{ bindState.dingTalkUserIdMasked }}
        </text>
        <text v-if="bindState.unionIdMasked">
          UnionID：{{ bindState.unionIdMasked }}
        </text>
        <text v-if="bindState.expiresIn">
          绑定票据有效期：{{ bindState.expiresIn }} 秒
        </text>
      </view>

      <view class="bind-form">
        <u-input v-model="form.account" placeholder="请输入系统账号" clearable :border="true" />
        <u-input v-model="form.password" type="password" placeholder="请输入系统密码" clearable :border="true" />
        <u-button type="primary" :loading="loading" @click="submit">
          绑定并登录
        </u-button>
        <u-button type="default" :disabled="loading" @click="emit('retry')">
          重新免登
        </u-button>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.bind-panel {
  min-height: 100vh;
  padding: 56rpx 32rpx;
  background: linear-gradient(180deg, rgba(var(--u-type-primary-rgb, 41, 121, 255), 0.08), transparent 420rpx);
}

.bind-card {
  padding: 36rpx 28rpx;
  border-radius: 20rpx;
  background: $u-bg-white;
  box-shadow: 0 16rpx 48rpx rgba(0, 0, 0, 0.08);

  &__head {
    display: flex;
    flex-direction: column;
    gap: 12rpx;
    align-items: center;
    text-align: center;
  }

  &__logo {
    width: 96rpx;
    height: 96rpx;
    border-radius: 24rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--u-type-primary);
    color: var(--u-white-color);
    font-size: 32rpx;
    font-weight: 700;
  }

  &__title {
    color: $u-main-color;
    font-size: 38rpx;
    font-weight: 700;
  }

  &__desc {
    color: $u-tips-color;
    font-size: 26rpx;
    line-height: 1.5;
  }
}

.bind-meta {
  margin: 28rpx 0;
  padding: 20rpx 24rpx;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  background: $u-bg-gray-light;
  color: $u-content-color;
  font-size: 24rpx;
}

.bind-form {
  display: flex;
  flex-direction: column;
  gap: 22rpx;
}
</style>
