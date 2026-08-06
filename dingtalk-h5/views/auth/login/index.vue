<template>
  <view class="login-page">
    <view class="login-card">
      <AuthBrandHeader :app-config="appConfig" />

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
import AuthBrandHeader from '../../../components/auth/AuthBrandHeader.vue'

defineProps({
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
</script>
