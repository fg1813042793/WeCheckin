<script setup lang="ts">
import type { AppConfig, LoginRequest } from '@/types/dingtalk-h5'
import { reactive } from 'vue'

defineProps<{
  appConfig: AppConfig
  autoLoginMessage: string
  loading: boolean
}>()

const emit = defineEmits<{
  login: [form: LoginRequest]
}>()

const form = reactive<LoginRequest>({
  name: '',
  password: '',
})

function submit() {
  if (!form.name.trim() || !form.password.trim()) {
    uni.showToast({
      title: '请输入账号和密码',
      icon: 'none',
    })
    return
  }
  emit('login', {
    name: form.name.trim(),
    password: form.password,
  })
}
</script>

<template>
  <view class="auth-panel">
    <view class="auth-panel__card">
      <view class="brand">
        <image v-if="appConfig.logoUrl" class="brand__logo-image" :src="appConfig.logoUrl" mode="aspectFill" />
        <view v-else class="brand__logo-text">
          {{ appConfig.logoText || 'OA' }}
        </view>
        <text class="brand__title">
          {{ appConfig.appTitle }}
        </text>
        <text class="brand__desc">
          钉钉 H5 绩效工作台
        </text>
      </view>

      <view v-if="autoLoginMessage" class="auth-tip">
        {{ autoLoginMessage }}
      </view>

      <view class="auth-form">
        <u-input v-model="form.name" placeholder="请输入系统账号" clearable :border="true" />
        <u-input v-model="form.password" type="password" placeholder="请输入密码" clearable :border="true" />
        <u-button type="primary" :loading="loading" @click="submit">
          登录
        </u-button>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.auth-panel {
  min-height: 100vh;
  padding: 72rpx 40rpx 40rpx;
  display: flex;
  flex-direction: column;
  gap: 36rpx;
  background: linear-gradient(180deg, rgba(var(--u-type-primary-rgb, 41, 121, 255), 0.08), transparent 420rpx);

  &__card {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 36rpx;
  }
}

.brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14rpx;

  &__logo-image,
  &__logo-text {
    width: 112rpx;
    height: 112rpx;
    border-radius: 28rpx;
    overflow: hidden;
    box-shadow: 0 16rpx 36rpx rgba(0, 0, 0, 0.12);
  }

  &__logo-text {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--u-type-primary);
    color: var(--u-white-color);
    font-size: 36rpx;
    font-weight: 700;
  }

  &__title {
    margin-top: 14rpx;
    color: $u-main-color;
    font-size: 42rpx;
    font-weight: 700;
  }

  &__desc {
    color: $u-tips-color;
    font-size: 26rpx;
  }
}

.auth-tip {
  padding: 20rpx 24rpx;
  border: 1rpx solid rgba(var(--u-type-warning-rgb, 255, 153, 0), 0.24);
  border-radius: 12rpx;
  background: rgba(var(--u-type-warning-rgb, 255, 153, 0), 0.08);
  color: $u-content-color;
  font-size: 26rpx;
  line-height: 1.5;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

@media screen and (min-width: 769px) {
  .auth-panel {
    padding: 64px 24px;
    align-items: center;
    justify-content: center;
    background:
      radial-gradient(circle at 50% 0%, rgba(var(--u-type-primary-rgb, 41, 121, 255), 0.08), transparent 320px),
      $u-bg-white;
  }

  .auth-panel__card {
    width: min(420px, calc(100vw - 48px));
    padding: 38px 34px 34px;
    border: 1px solid rgba(229, 230, 235, 0.88);
    border-radius: 12px;
    background: $u-bg-white;
    box-shadow: 0 18px 48px rgba(15, 23, 42, 0.10);
    gap: 24px;
  }

  .brand {
    gap: 8px;

    &__logo-image,
    &__logo-text {
      width: 56px;
      height: 56px;
      border-radius: 14px;
      box-shadow: 0 12px 28px rgba(15, 23, 42, 0.14);
    }

    &__logo-text {
      font-size: 18px;
    }

    &__title {
      margin-top: 8px;
      font-size: 22px;
      line-height: 1.25;
    }

    &__desc {
      font-size: 13px;
      line-height: 1.5;
    }
  }

  .auth-tip {
    padding: 12px 14px;
    border-radius: 8px;
    font-size: 13px;
  }

  .auth-form {
    gap: 14px;
  }

  .auth-form :deep(.u-input) {
    min-height: 40px;
    background: $u-bg-white;
  }

  .auth-form :deep(.u-btn) {
    width: 100%;
    height: 40px;
    margin-top: 4px;
    border-radius: 6px;
  }
}
</style>
