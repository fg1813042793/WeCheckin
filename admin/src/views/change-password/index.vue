<template>
  <AdminPageShell as="div" width="fluid" class="password-change-page">
    <section class="password-change-panel">
      <div class="password-change-mark">W</div>
      <h1>请修改登录密码</h1>
      <p>当前账号使用临时密码或密码已过期，修改完成后才能继续使用系统。</p>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="submit">
        <el-form-item label="当前密码" prop="currentPassword">
          <el-input v-model="form.currentPassword" type="password" show-password autocomplete="current-password" />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="form.newPassword" type="password" show-password autocomplete="new-password" />
        </el-form-item>
        <el-form-item label="确认新密码" prop="confirmPassword">
          <el-input v-model="form.confirmPassword" type="password" show-password autocomplete="new-password" />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="saving" class="submit-button">修改密码</el-button>
      </el-form>
    </section>
  </AdminPageShell>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { adminApi } from '../../api'
import { AdminPageShell } from '@/components/admin-ui'
import { clearAdminSession } from '../../utils/adminSession'
import { invalidateAdminAccessSnapshot } from '../../router/adminAccess'

const router = useRouter()
const formRef = ref<FormInstance>()
const saving = ref(false)
const form = reactive({ currentPassword: '', newPassword: '', confirmPassword: '' })
const rules: FormRules = {
  currentPassword: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  newPassword: [{ required: true, message: '请输入新密码', trigger: 'blur' }],
  confirmPassword: [{
    validator: (_rule, value, callback) => value === form.newPassword ? callback() : callback(new Error('两次输入的新密码不一致')),
    trigger: 'blur',
  }],
}

async function submit() {
  if (!(await formRef.value?.validate().catch(() => false))) return
  saving.value = true
  try {
    await adminApi.mgrPwd({ oldPassword: form.currentPassword, password: form.newPassword })
    clearAdminSession()
    invalidateAdminAccessSnapshot()
    ElMessage.success('密码修改成功，请重新登录')
    await router.replace('/login')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.password-change-page { min-height: 100vh; display: grid; place-items: center; padding: 24px; background: #f5f7fb; }
.password-change-panel { width: min(420px, 100%); padding: 34px; border: 1px solid #dfe5ec; border-radius: 8px; background: #fff; box-shadow: 0 16px 40px rgb(15 23 42 / 8%); }
.password-change-mark { display: grid; place-items: center; width: 44px; height: 44px; margin-bottom: 20px; border-radius: 8px; color: #fff; background: #2f7bdc; font-size: 24px; font-weight: 700; }
h1 { margin: 0; color: #1f2937; font-size: 24px; }
p { margin: 10px 0 24px; color: #667085; line-height: 1.7; }
.submit-button { width: 100%; }
</style>
