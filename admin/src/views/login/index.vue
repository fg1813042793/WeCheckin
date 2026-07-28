<template>
  <div class="login-shell">
    <section class="login-brand-panel">
      <div class="brand-mark">W</div>
      <h1>WeCheckin 管理后台</h1>
      <p>统一管理用户、活动、问卷、考试和系统配置，让日常运营更清晰。</p>
      <div class="login-feature-list">
        <span>活动打卡</span>
        <span>问卷统计</span>
        <span>在线考试</span>
        <span>权限配置</span>
      </div>
    </section>

    <section class="login-form-panel">
      <el-card class="login-card" shadow="never">
        <div class="login-card__header">
          <h2>登录后台</h2>
          <p>请输入姓名/手机号和密码</p>
        </div>
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
          <el-form-item label="姓名/手机号" prop="name">
            <el-input v-model="form.name" prefix-icon="User" placeholder="请输入姓名或手机号" size="large" />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              prefix-icon="Lock"
              type="password"
              show-password
              placeholder="请输入密码"
              size="large"
              @keyup.enter="login"
            />
          </el-form-item>
          <el-button type="primary" class="login-submit" size="large" :loading="loading" @click="login">登 录</el-button>
        </el-form>
        <p class="login-tip">登录后将根据账号权限展示对应菜单。</p>
      </el-card>
    </section>
  </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminApi } from '../../api'
import { setPerms } from '../../utils/permission'
import type { FormInstance } from 'element-plus'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ name: '', password: '' })
const rules = {
  name: [{ required: true, message: '请输入姓名或手机号', trigger: 'blur' }],
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
.login-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(360px, 0.95fr) minmax(420px, 1.05fr);
  background: #f5f7fb;
}

.login-brand-panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 64px;
  color: #fff;
  background:
    linear-gradient(135deg, rgba(31, 45, 61, 0.92), rgba(37, 99, 235, 0.88)),
    radial-gradient(circle at 20% 20%, rgba(255, 255, 255, 0.2), transparent 32%);
}

.brand-mark {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.22);
  font-size: 26px;
  font-weight: 800;
}

.login-brand-panel h1 {
  margin: 24px 0 12px;
  font-size: 34px;
  line-height: 1.25;
}

.login-brand-panel p {
  max-width: 440px;
  margin: 0;
  color: rgba(255, 255, 255, 0.78);
  font-size: 15px;
  line-height: 1.8;
}

.login-feature-list {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 28px;
}

.login-feature-list span {
  padding: 7px 12px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.18);
  font-size: 13px;
}

.login-form-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.login-card {
  width: 420px;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.login-card__header {
  margin-bottom: 24px;
}

.login-card__header h2 {
  margin: 0;
  color: #111827;
  font-size: 24px;
}

.login-card__header p,
.login-tip {
  margin: 8px 0 0;
  color: #6b7280;
  font-size: 13px;
}

.login-submit {
  width: 100%;
  margin-top: 4px;
}

.login-tip {
  text-align: center;
  margin-top: 18px;
}

@media (max-width: 768px) {
  .login-shell {
    grid-template-columns: 1fr;
  }

  .login-brand-panel {
    display: none;
  }

  .login-form-panel {
    padding: 20px;
  }

  .login-card {
    width: 100%;
  }
}
</style>
