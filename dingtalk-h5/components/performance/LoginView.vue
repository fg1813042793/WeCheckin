<template>
  <view class="login-page">
    <view class="login-card">
      <view class="brand-line">
        <view class="brand-logo">
          <image class="brand-logo-img" v-if="appLogoUrl" :src="appLogoUrl" mode="aspectFill" />
          <text v-else>{{ appLogoText }}</text>
        </view>
        <view>
          <text class="brand-title">{{ appDisplayTitle }}</text>
          <text class="brand-subtitle">钉钉 OA 微应用</text>
        </view>
      </view>

      <view class="login-copy">
        <text class="login-title">登录</text>
        <text class="login-desc">钉钉内打开会自动登录；无法登录时使用账号登录。</text>
      </view>

      <view v-if="autoLoginMessage" class="login-warning">
        <text>{{ autoLoginMessage }}</text>
      </view>

      <view class="login-form">
        <input v-model="form.name" class="dt-input" placeholder="姓名或账号" />
        <input v-model="form.password" class="dt-input" password placeholder="请输入密码" />
        <button class="dt-btn dt-btn-primary dt-btn-block" :loading="loading" @click="$emit('login')">进入系统</button>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  form: {
    type: Object,
    required: true
  },
  appConfig: {
    type: Object,
    default: () => ({})
  },
  loading: {
    type: Boolean,
    default: false
  },
  autoLoginMessage: {
    type: String,
    default: ''
  }
})

defineEmits(['login'])

const appDisplayTitle = computed(() => firstText(props.appConfig.appTitle, props.appConfig.appName, 'OA管理'))
const appLogoText = computed(() => firstText(props.appConfig.logoText, 'OA').slice(0, 4))
const appLogoUrl = computed(() => firstText(props.appConfig.logoUrl, props.appConfig.logoURL))

function firstText(...values) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) return text
  }
  return ''
}
</script>
