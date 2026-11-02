<script setup lang="ts">
import type { AppConfig, LoginRequest } from '@/types/dingtalk-h5'
import { useLocale } from 'uview-pro'
import { computed, onMounted, reactive, watch } from 'vue'

const props = withDefaults(defineProps<{
  appConfig: AppConfig
  autoLoginMessage: string
  showDingtalkLogin?: boolean
  captchaEnabled?: boolean
  captchaId?: string
  captchaImage?: string
  captchaLoading?: boolean
  captchaUnavailable?: boolean
  loading: boolean
}>(), {
  captchaEnabled: false,
  showDingtalkLogin: false,
  captchaId: '',
  captchaImage: '',
  captchaLoading: false,
  captchaUnavailable: false,
})

const emit = defineEmits<{
  'login': [form: LoginRequest]
  'refresh-captcha': []
  'retry-dingtalk': []
}>()

const { t } = useLocale('auth.loginCaptcha')
const { t: authText } = useLocale('auth.dingtalk')

const form = reactive<LoginRequest>({
  name: '',
  password: '',
  captchaCode: '',
})

const captchaReady = computed(() => !props.captchaLoading && (
  !props.captchaEnabled
  || Boolean(props.captchaId && props.captchaImage && !props.captchaUnavailable)
))

watch(() => props.captchaId, () => {
  form.captchaCode = ''
})

onMounted(refreshCaptcha)

function refreshCaptcha() {
  if (props.captchaLoading)
    return
  emit('refresh-captcha')
}

function submit() {
  if (!form.name.trim() || !form.password.trim()) {
    uni.showToast({
      title: '请输入账号和密码',
      icon: 'none',
    })
    return
  }
  if (props.captchaEnabled && !String(form.captchaCode || '').trim()) {
    uni.showToast({
      title: t('required'),
      icon: 'none',
    })
    return
  }
  if (!captchaReady.value) {
    uni.showToast({
      title: t('unavailable'),
      icon: 'none',
    })
    return
  }
  emit('login', {
    name: form.name.trim(),
    password: form.password,
    captchaId: props.captchaEnabled ? props.captchaId : undefined,
    captchaCode: props.captchaEnabled ? String(form.captchaCode || '').trim().toUpperCase() : undefined,
  })
}
</script>

<template>
  <view class="auth-panel app-pc-control-scope">
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
        <view v-if="showDingtalkLogin" class="auth-tip__retry">
          {{ authText('retryHint') }}
        </view>
      </view>

      <view class="auth-form">
        <u-button v-if="showDingtalkLogin" type="primary" :loading="loading" @click="emit('retry-dingtalk')">
          {{ authText('retry') }}
        </u-button>
        <u-input v-model="form.name" placeholder="请输入系统账号" clearable :border="true" />
        <u-input v-model="form.password" type="password" placeholder="请输入密码" clearable :border="true" />
        <view v-if="captchaEnabled" class="captcha-field">
          <u-input
            v-model="form.captchaCode"
            custom-class="captcha-input"
            :placeholder="t('placeholder')"
            clearable
            :border="true"
          />
          <image
            v-if="captchaImage"
            class="captcha-image"
            :src="captchaImage"
            mode="aspectFit"
            role="button"
            tabindex="0"
            :title="t('refresh')"
            :aria-label="t('refresh')"
            :aria-disabled="captchaLoading"
            @click="refreshCaptcha"
            @keyup.enter="refreshCaptcha"
          />
          <view
            v-else
            class="captcha-image captcha-image--empty"
            role="button"
            tabindex="0"
            :title="t('refresh')"
            :aria-label="t('refresh')"
            :aria-disabled="captchaLoading"
            @click="refreshCaptcha"
            @keyup.enter="refreshCaptcha"
          >
            <u-loading v-if="captchaLoading" mode="circle" />
            <u-icon v-else name="photo" size="36" color="var(--u-tips-color)" />
          </view>
        </view>
        <view v-if="captchaEnabled && captchaUnavailable" class="captcha-error">
          {{ t('unavailable') }}
        </view>
        <u-button :type="showDingtalkLogin ? 'default' : 'primary'" :loading="loading" :disabled="!captchaReady" @click="submit">
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

.auth-tip__retry {
  margin-top: 8px;
}

.captcha-field {
  display: flex;
  min-width: 0;
  align-items: stretch;
  gap: 12rpx;
}

.captcha-field :deep(.captcha-input) {
  min-width: 0;
  flex: 1 1 auto;
}

.captcha-image {
  box-sizing: border-box;
  width: 256rpx;
  height: 88rpx;
  flex: 0 0 256rpx;
  border: 1rpx solid $u-border-color;
  border-radius: 8rpx;
  background: $u-bg-color;
  cursor: pointer;
  object-fit: contain;
}

.captcha-image--empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

.captcha-error {
  margin-top: -12rpx;
  color: $u-type-error;
  font-size: 24rpx;
  line-height: 1.5;
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

  .captcha-field {
    gap: 8px;
  }

  .captcha-image {
    width: 128px;
    height: 44px;
    flex-basis: 128px;
    border-width: 1px;
    border-radius: 6px;
  }

  .captcha-error {
    margin-top: -6px;
    font-size: 12px;
  }

  .auth-form :deep(.u-btn) {
    width: 100%;
    height: 40px;
    margin-top: 4px;
    border-radius: 6px;
  }
}

@media screen and (min-width: 769px) and (max-width: 1023px) {
  .auth-panel {
    padding: 24px;
  }

  .auth-panel__card {
    padding: 28px;
    gap: 20px;
  }

  .auth-form {
    gap: 12px;
  }

  .captcha-image {
    width: 112px;
    height: var(--app-control-height);
    flex-basis: 112px;
  }
}
</style>
