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
        <text class="login-title">绑定系统账号</text>
        <text class="login-desc">已识别当前钉钉身份，请输入绩效系统账号和密码完成首次绑定。</text>
      </view>

      <view class="bind-summary">
        <view class="bind-summary-row">
          <text class="bind-summary-label">企业</text>
          <text class="bind-summary-value">{{ bindState.corpId || '-' }}</text>
        </view>
        <view class="bind-summary-row">
          <text class="bind-summary-label">钉钉用户</text>
          <text class="bind-summary-value">{{ bindState.dingTalkUserIdMasked || '-' }}</text>
        </view>
      </view>

      <view class="login-form">
        <input v-model="form.account" class="dt-input" placeholder="系统账号 / 姓名 / 手机号" />
        <input v-model="form.password" class="dt-input" password placeholder="系统密码" />
        <button class="dt-btn dt-btn-primary dt-btn-block" :loading="loading" @click="$emit('bind')">绑定并进入</button>
        <button class="dt-btn dt-btn-light dt-btn-block" :disabled="loading" @click="$emit('retry')">重新获取钉钉身份</button>
      </view>

      <text class="login-tip">绑定成功后，下次在钉钉中打开会自动进入系统。</text>
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
  bindState: {
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
  }
})

defineEmits(['bind', 'retry'])

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
