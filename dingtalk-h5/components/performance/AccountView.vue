<template>
  <view class="account-page">
    <view class="page-head">
      <view>
        <text class="page-title">{{ ctx.sectionTitle.value }}</text>
        <text class="page-desc">{{ [ctx.state.user.name, ctx.state.user.position].filter(Boolean).join(' · ') }}</text>
      </view>
    </view>

    <section class="panel account-panel">
      <view class="panel-head">
        <text class="panel-title">修改密码</text>
      </view>
      <view class="password-form">
        <input v-model="ctx.accountForm.currentPassword" class="field-input" password placeholder="当前密码" />
        <input v-model="ctx.accountForm.newPassword" class="field-input" password placeholder="新密码" />
        <input v-model="ctx.accountForm.confirmPassword" class="field-input" password placeholder="确认新密码" />
        <button class="dt-btn dt-btn-primary" @click="ctx.changePassword">保存密码</button>
      </view>
    </section>
  </view>
</template>

<script setup>
import { usePerformanceContext } from './context'

const ctx = usePerformanceContext()
</script>
