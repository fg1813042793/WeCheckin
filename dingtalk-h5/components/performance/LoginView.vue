<template>
  <view class="login-page">
    <view class="login-card">
      <view class="brand-line">
        <view class="brand-logo">OA</view>
        <view>
          <text class="brand-title">OA管理</text>
          <text class="brand-subtitle">钉钉 OA 微应用</text>
        </view>
      </view>

      <view class="login-copy">
        <text class="login-title">登录</text>
        <text class="login-desc">使用绩效系统账号进入对应流程，钉钉免登接口可在此基础上继续接入。</text>
      </view>

      <view class="login-form">
        <input v-model="form.name" class="dt-input" placeholder="姓名或账号，例如 Nick" />
        <input v-model="form.password" class="dt-input" password placeholder="密码，默认 123456" />
        <button class="dt-btn dt-btn-primary dt-btn-block" :loading="loading" @click="$emit('login')">进入系统</button>
      </view>

      <text class="login-tip">Lip / Cube / HRBP / Nick 等账号可使用 123456 体验。</text>
    </view>
  </view>
</template>

<script setup>
defineProps({
  form: {
    type: Object,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  }
})

defineEmits(['login'])
</script>
