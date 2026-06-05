<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2 style="text-align:center;margin-bottom:30px">WeCheckin 管理后台</h2>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="账号" prop="name">
          <el-input v-model="form.name" placeholder="请输入管理员账号" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" @keyup.enter="login" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width:100%" :loading="loading" @click="login">登 录</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { adminApi } from '../../api'
import { setPerms } from '../../utils/permission'
import type { FormInstance } from 'element-plus'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ name: '', password: '' })
const rules = {
  name: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function login() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    const res = await adminApi.login({ name: form.name, password: form.password })
    const d = res.data || {}
    localStorage.setItem('admin_token', d.token || '')
    localStorage.setItem('admin_info', JSON.stringify(d))
    const permRes = await adminApi.adminPerms()
    setPerms(permRes.data || [])
    router.push('/dashboard')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
}
.login-card {
  width: 420px;
}
</style>
