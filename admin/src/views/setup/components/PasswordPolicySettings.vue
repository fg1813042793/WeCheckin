<template>
  <section class="password-policy-settings" aria-label="密码策略">
    <el-divider content-position="left">密码策略</el-divider>
    <el-alert
      title="策略默认关闭。开启后仅校验新建、重置和修改的密码，不会立即要求现有账号改密。"
      type="info"
      :closable="false"
      show-icon
    />
    <el-form v-loading="loading" label-width="180px">
      <el-form-item label="启用密码策略">
        <el-switch v-model="policy.enabled" />
      </el-form-item>
      <el-form-item label="最小密码长度">
        <el-input-number v-model="policy.minLength" :min="6" :max="64" :disabled="!policy.enabled" />
      </el-form-item>
      <el-form-item label="字符要求">
        <el-checkbox v-model="policy.requireLetter" :disabled="!policy.enabled">包含字母</el-checkbox>
        <el-checkbox v-model="policy.requireNumber" :disabled="!policy.enabled">包含数字</el-checkbox>
        <el-checkbox v-model="policy.requireUppercase" :disabled="!policy.enabled">包含大写字母</el-checkbox>
        <el-checkbox v-model="policy.requireLowercase" :disabled="!policy.enabled">包含小写字母</el-checkbox>
        <el-checkbox v-model="policy.requireSpecial" :disabled="!policy.enabled">包含特殊字符</el-checkbox>
      </el-form-item>
      <el-form-item label="密码有效期">
        <el-input-number v-model="policy.maxAgeDays" :min="0" :max="3650" :disabled="!policy.enabled" />
        <span class="field-help">天，0 表示永不过期</span>
      </el-form-item>
      <el-form-item label="到期前提醒">
        <el-input-number v-model="policy.expiryWarningDays" :min="0" :max="Math.max(policy.maxAgeDays, 0)" :disabled="!policy.enabled || policy.maxAgeDays === 0" />
        <span class="field-help">天</span>
      </el-form-item>
      <el-form-item label="历史密码">
        <el-input-number v-model="policy.historyCount" :min="0" :max="20" :disabled="!policy.enabled" />
        <span class="field-help">个，0 表示不限制重复</span>
      </el-form-item>
      <el-form-item label="新账号首次登录改密">
        <el-switch v-model="policy.forceChangeOnCreate" :disabled="!policy.enabled" />
      </el-form-item>
      <el-form-item label="重置后强制改密">
        <el-switch v-model="policy.forceChangeOnReset" :disabled="!policy.enabled" />
      </el-form-item>
      <el-form-item label="允许弱临时密码">
        <el-switch v-model="policy.allowWeakTemporaryPassword" :disabled="!policy.enabled" />
        <span class="field-help">仅限管理员创建或重置，用户首次登录仍须修改</span>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save">保存密码策略</el-button>
      </el-form-item>
    </el-form>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getPasswordPolicy, savePasswordPolicy } from '../../../api/passwordPolicy'
import type { PasswordPolicy } from '../../../types/passwordPolicy'

const loading = ref(false)
const saving = ref(false)
const policy = reactive<PasswordPolicy>({
  version: 1,
  enabled: false,
  minLength: 8,
  requireLetter: true,
  requireNumber: true,
  requireUppercase: false,
  requireLowercase: false,
  requireSpecial: false,
  maxAgeDays: 90,
  expiryWarningDays: 7,
  historyCount: 5,
  forceChangeOnCreate: true,
  forceChangeOnReset: true,
  allowWeakTemporaryPassword: true,
  effectiveAt: 0,
})

async function load() {
  loading.value = true
  try {
    const response = await getPasswordPolicy()
    Object.assign(policy, response.data)
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const response = await savePasswordPolicy({ ...policy })
    Object.assign(policy, response.data)
    ElMessage.success('密码策略已保存')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.password-policy-settings { margin-top: 22px; }
.password-policy-settings .el-alert { margin-bottom: 18px; }
.field-help { margin-left: 10px; color: #8492a6; font-size: 12px; }
</style>
