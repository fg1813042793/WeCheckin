<template>
  <slot v-if="!ready" name="loading" />

  <slot
    v-else-if="!user && bindState.visible"
    name="bind"
    :form="bindForm"
    :bind-state="bindState"
    :loading="loading"
    :app-config="appConfig"
  />

  <slot
    v-else-if="!user && sessionAccessDenied"
    name="denied"
    :message="sessionAccessDeniedMessage"
  />

  <slot
    v-else-if="!user"
    name="login"
    :form="loginForm"
    :loading="loading"
    :app-config="appConfig"
    :auto-login-message="autoLoginMessage"
  />

  <slot v-else />
</template>

<script setup>
defineProps({
  appConfig: { type: Object, required: true },
  autoLoginMessage: { type: String, default: '' },
  bindForm: { type: Object, required: true },
  bindState: { type: Object, required: true },
  loading: { type: Boolean, default: false },
  loginForm: { type: Object, required: true },
  ready: { type: Boolean, default: false },
  sessionAccessDenied: { type: Boolean, default: false },
  sessionAccessDeniedMessage: { type: String, default: '无权限访问，请联系管理员配置钉钉 H5 权限' },
  user: { type: Object, default: null }
})
</script>
