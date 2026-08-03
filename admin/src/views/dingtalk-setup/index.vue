<template>
  <div class="admin-page dingtalk-setup-page">
    <el-card class="admin-card" shadow="never" v-loading="loading">
      <template #header>
        <div class="admin-card__header">
          <div>
            <div class="admin-card__title">钉钉应用配置</div>
          </div>
          <el-button circle icon="Refresh" title="刷新" @click="loadSettings" />
        </div>
      </template>

      <el-tabs v-model="activeTab" class="dingtalk-settings-tabs">
        <el-tab-pane label="企业应用" name="corp">
          <el-form label-width="120px" class="dingtalk-settings-form dingtalk-settings-form--corp">
            <el-form-item label="企业应用">
              <div class="corp-config-list">
                <div
                  v-for="(corpConfig, index) in form.corpConfigs"
                  :key="corpConfig.localKey"
                  class="corp-config-item"
                >
                  <div class="corp-config-item__head">
                    <span>企业 {{ index + 1 }}</span>
                    <el-button
                      v-if="form.corpConfigs.length > 1"
                      text
                      type="danger"
                      @click="removeCorpConfig(index)"
                    >
                      删除
                    </el-button>
                  </div>
                  <el-row :gutter="12">
                    <el-col :span="12">
                      <el-input v-model="corpConfig.corpId" placeholder="钉钉企业 CorpId" />
                    </el-col>
                    <el-col :span="12">
                      <el-input v-model="corpConfig.corpName" placeholder="企业名称（选填）" />
                    </el-col>
                  </el-row>
                  <el-row :gutter="12">
                    <el-col :span="12">
                      <el-input v-model="corpConfig.appKey" placeholder="钉钉内部应用 AppKey" />
                    </el-col>
                    <el-col :span="12">
                      <el-input
                        v-model="corpConfig.appSecret"
                        type="password"
                        show-password
                        :placeholder="corpConfig.appSecretSet ? '已保存，留空表示不修改' : '请输入 AppSecret'"
                      />
                    </el-col>
                  </el-row>
                  <el-row :gutter="12">
                    <el-col :span="12">
                      <el-input v-model="corpConfig.agentId" placeholder="钉钉内部应用 AgentId" />
                    </el-col>
                  </el-row>
                  <el-switch
                    v-model="corpConfig.enabled"
                    :active-value="1"
                    :inactive-value="0"
                    active-text="启用"
                    inactive-text="停用"
                    inline-prompt
                  />
                </div>
                <el-button class="corp-config-add" @click="addCorpConfig">新增企业应用</el-button>
              </div>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="登录配置" name="login">
          <el-form label-width="160px" class="dingtalk-settings-form">
            <el-form-item label="Token 过期时间">
              <el-input v-model="form.tokenExpire" placeholder="168h">
                <template #append>例: 168h / 24h</template>
              </el-input>
            </el-form-item>
            <el-form-item label="Redis Key 前缀">
              <el-input v-model="form.redisPrefix" placeholder="dingtalk_h5_token:" />
            </el-form-item>
            <el-form-item label="单点登录">
              <el-switch
                v-model="form.singleLogin"
                :active-value="1"
                :inactive-value="0"
                active-text="开启"
                inactive-text="关闭"
                inline-prompt
                style="--el-switch-on-color: #f56c6c; --el-switch-off-color: #67c23a"
              />
            </el-form-item>
            <el-form-item label="首次自助绑定">
              <el-switch
                v-model="form.selfBind"
                :active-value="1"
                :inactive-value="0"
                active-text="开启"
                inactive-text="关闭"
                inline-prompt
              />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="配置" name="app">
          <el-form label-width="160px" class="dingtalk-settings-form">
            <el-form-item label="应用名称">
              <el-input v-model="form.appName" maxlength="24" placeholder="例如 OA管理" />
            </el-form-item>
            <el-form-item label="Logo 文字">
              <el-input v-model="form.logoText" maxlength="4" placeholder="未配置时默认 OA" />
            </el-form-item>
            <el-form-item label="Logo 图片地址">
              <el-input v-model="form.logoUrl" placeholder="https://...，留空时显示 Logo 文字" />
            </el-form-item>
            <el-form-item label="绩效流程通知">
              <div class="settings-switch-line">
                <el-switch
                  v-model="form.notifyEnabled"
                  :active-value="1"
                  :inactive-value="0"
                  active-text="开启"
                  inactive-text="关闭"
                  inline-prompt
                />
                <span class="settings-help">开启后，员工提交自评会通过钉钉工作通知提醒直属上级。</span>
              </div>
            </el-form-item>
            <el-form-item label="显示预览">
              <div class="brand-preview">
                <div class="brand-preview__logo">
                  <img v-if="previewLogoUrl" :src="previewLogoUrl" alt="logo" />
                  <span v-else>{{ previewLogoText }}</span>
                </div>
                <div class="brand-preview__name">{{ previewAppName }}</div>
              </div>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <div class="dingtalk-settings-actions">
        <el-button
          type="primary"
          :loading="saving"
          :disabled="!hasPerm('admin:menu:dingtalk:config:edit')"
          @click="saveSettings"
        >
          保存
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../../utils/request'
import { hasPerm } from '../../utils/permission'

const defaultAppName = 'OA管理'
const defaultLogoText = 'OA'

type CorpConfigForm = {
  localKey: string
  corpId: string
  corpName: string
  appKey: string
  appSecret: string
  agentId: string
  appSecretSet: boolean
  enabled: number
}

function newCorpConfig(data: Partial<CorpConfigForm> = {}): CorpConfigForm {
  return {
    localKey: data.localKey || `${Date.now()}-${Math.random().toString(16).slice(2)}`,
    corpId: data.corpId || '',
    corpName: data.corpName || '',
    appKey: data.appKey || '',
    agentId: data.agentId || '',
    appSecret: '',
    appSecretSet: Boolean(data.appSecretSet),
    enabled: Number(data.enabled) === 0 ? 0 : 1
  }
}

const activeTab = ref('corp')
const form = reactive({
  corpId: '',
  appKey: '',
  appSecret: '',
  appSecretSet: false,
  corpConfigs: [newCorpConfig()],
  tokenExpire: '168h',
  redisPrefix: 'dingtalk_h5_token:',
  singleLogin: 0,
  selfBind: 1,
  notifyEnabled: 0,
  appName: defaultAppName,
  logoText: defaultLogoText,
  logoUrl: ''
})
const loading = ref(false)
const saving = ref(false)

const previewAppName = computed(() => form.appName.trim() || defaultAppName)
const previewLogoText = computed(() => (form.logoText.trim() || defaultLogoText).slice(0, 4))
const previewLogoUrl = computed(() => form.logoUrl.trim())

function normalizeCorpConfigs(data: any): CorpConfigForm[] {
  const source = Array.isArray(data) ? data : []
  const items = source
    .map(item =>
      newCorpConfig({
        corpId: item?.corpId || '',
        corpName: item?.corpName || '',
        appKey: item?.appKey || '',
        agentId: item?.agentId || '',
        appSecretSet: Boolean(item?.appSecretSet),
        enabled: Number(item?.enabled) === 0 ? 0 : 1
      })
    )
    .filter(item => item.corpId || item.appKey || item.corpName)
  return items.length > 0 ? items : [newCorpConfig()]
}

function addCorpConfig() {
  form.corpConfigs.push(newCorpConfig())
}

async function removeCorpConfig(index: number) {
  try {
    await ElMessageBox.confirm('确定删除该企业应用？', '提示', { type: 'warning' })
  } catch {
    return
  }
  form.corpConfigs.splice(index, 1)
  if (form.corpConfigs.length === 0) {
    form.corpConfigs.push(newCorpConfig())
  }
}

async function loadSettings() {
  loading.value = true
  try {
    const res = await request.get('/api/v2/admin/dingtalk/settings')
    const data = res.data || {}
    form.corpId = data.corpId || ''
    form.appKey = data.appKey || ''
    form.appSecret = ''
    form.appSecretSet = Boolean(data.appSecretSet)
    form.corpConfigs = normalizeCorpConfigs(
      data.corpConfigs && data.corpConfigs.length
        ? data.corpConfigs
        : [{ corpId: data.corpId || '', corpName: data.corpId || '', appKey: data.appKey || '', agentId: data.agentId || '', appSecretSet: data.appSecretSet }]
    )
    form.tokenExpire = data.tokenExpire || '168h'
    form.redisPrefix = data.redisPrefix || 'dingtalk_h5_token:'
    form.singleLogin = Number(data.singleLogin) === 1 ? 1 : 0
    form.selfBind = Number(data.selfBind) === 0 ? 0 : 1
    form.notifyEnabled = Number(data.notifyEnabled) === 1 ? 1 : 0
    form.appName = data.appName || defaultAppName
    form.logoText = data.logoText || defaultLogoText
    form.logoUrl = data.logoUrl || ''
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  const tokenExpire = form.tokenExpire.trim()
  const redisPrefix = form.redisPrefix.trim()
  const appName = previewAppName.value
  const logoText = previewLogoText.value
  if (!tokenExpire) {
    ElMessage.warning('请输入 Token 过期时间')
    return
  }
  if (!redisPrefix) {
    ElMessage.warning('请输入 Redis Key 前缀')
    return
  }
  const corpConfigs = form.corpConfigs
    .map(item => ({
      corpId: item.corpId.trim(),
      corpName: item.corpName.trim(),
      appKey: item.appKey.trim(),
      appSecret: item.appSecret.trim(),
      agentId: item.agentId.trim(),
      enabled: item.enabled
    }))
    .filter(item => item.corpId || item.corpName || item.appKey || item.appSecret)
  for (const item of corpConfigs) {
    if (!item.corpId) {
      ElMessage.warning('请填写企业 CorpId')
      return
    }
    if (!item.appKey) {
      ElMessage.warning('请填写企业 AppKey')
      return
    }
  }
  const firstCorpConfig = corpConfigs[0] || { corpId: '', appKey: '', appSecret: '', agentId: '' }
  saving.value = true
  try {
    await request.put('/api/v2/admin/dingtalk/settings', {
      corpId: firstCorpConfig.corpId,
      appKey: firstCorpConfig.appKey,
      appSecret: firstCorpConfig.appSecret,
      agentId: firstCorpConfig.agentId,
      corpConfigs: JSON.stringify(corpConfigs),
      tokenExpire,
      redisPrefix,
      singleLogin: form.singleLogin,
      selfBind: form.selfBind,
      notifyEnabled: form.notifyEnabled,
      appName,
      logoText,
      logoUrl: form.logoUrl.trim()
    })
    ElMessage.success('保存成功')
    await loadSettings()
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.dingtalk-setup-page {
  padding: 20px;
}

.dingtalk-settings-tabs {
  max-width: 760px;
}

.dingtalk-settings-form {
  max-width: 620px;
  padding-top: 12px;
}

.dingtalk-settings-form--corp {
  max-width: 720px;
}

.dingtalk-settings-actions {
  max-width: 620px;
  padding-left: 160px;
}

.corp-config-list {
  width: 100%;
}

.corp-config-item {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  border: 1px solid #e5eaf3;
  border-radius: 8px;
  background: #f8fafc;
}

.corp-config-item + .corp-config-item {
  margin-top: 12px;
}

.corp-config-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #1f2937;
  font-size: 14px;
  font-weight: 600;
}

.corp-config-add {
  margin-top: 12px;
}

.settings-switch-line {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 32px;
}

.settings-help {
  color: #6b7280;
  font-size: 13px;
}

.admin-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.admin-card__title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.brand-preview {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-width: 220px;
  padding: 12px 14px;
  border: 1px solid #e5eaf3;
  border-radius: 8px;
  background: #f8fafc;
}

.brand-preview__logo {
  display: flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 8px;
  background: #1677ff;
  color: #fff;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0;
}

.brand-preview__logo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.brand-preview__name {
  color: #1f2937;
  font-size: 15px;
  font-weight: 700;
}
</style>
