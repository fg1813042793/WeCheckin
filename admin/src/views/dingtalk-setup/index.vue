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
          <el-form class="dingtalk-settings-form dingtalk-settings-form--corp">
            <div class="corp-config-list">
              <div
                v-for="(corpConfig, index) in form.corpConfigs"
                :key="corpConfig.localKey"
                class="corp-config-item"
              >
                <div class="corp-config-item__head">
                  <div class="corp-config-item__identity">
                    <span class="corp-config-item__index">{{ index + 1 }}</span>
                    <div>
                      <div class="corp-config-item__title">
                        {{ corpConfig.corpName.trim() || `企业应用 ${index + 1}` }}
                      </div>
                      <div class="corp-config-item__meta">
                        {{ corpConfig.corpId.trim() || '未填写 CorpId' }}
                      </div>
                    </div>
                  </div>
                  <div class="corp-config-item__actions">
                    <div class="corp-config-switch">
                      <span>应用状态</span>
                      <el-switch
                        v-model="corpConfig.enabled"
                        :active-value="1"
                        :inactive-value="0"
                        active-text="启用"
                        inactive-text="停用"
                        inline-prompt
                      />
                    </div>
                    <div class="corp-config-switch">
                      <span>钉钉通知</span>
                      <el-switch
                        v-model="corpConfig.notifyEnabled"
                        :active-value="1"
                        :inactive-value="0"
                        active-text="开启"
                        inactive-text="关闭"
                        inline-prompt
                      />
                    </div>
                    <el-tooltip v-if="form.corpConfigs.length > 1" content="删除企业应用" placement="top">
                      <el-button circle icon="Delete" type="danger" plain @click="removeCorpConfig(index)" />
                    </el-tooltip>
                  </div>
                </div>

                <section class="corp-config-section">
                  <div class="corp-config-section__title">基本信息</div>
                  <div class="corp-config-grid">
                    <div class="corp-config-field">
                      <div class="corp-config-field__label">企业 CorpId</div>
                      <el-input v-model="corpConfig.corpId" placeholder="钉钉企业 CorpId" />
                    </div>
                    <div class="corp-config-field">
                      <div class="corp-config-field__label">企业名称</div>
                      <el-input v-model="corpConfig.corpName" placeholder="企业名称（选填）" />
                    </div>
                    <div class="corp-config-field">
                      <div class="corp-config-field__label">应用 AppKey</div>
                      <el-input v-model="corpConfig.appKey" placeholder="钉钉内部应用 AppKey" />
                    </div>
                    <div class="corp-config-field">
                      <div class="corp-config-field__label">应用 AppSecret</div>
                      <el-input
                        v-model="corpConfig.appSecret"
                        type="password"
                        show-password
                        :placeholder="corpConfig.appSecretSet ? '已保存，留空表示不修改' : '请输入 AppSecret'"
                      />
                    </div>
                  </div>
                </section>

                <section class="corp-config-section">
                  <div class="corp-config-section__title">通知配置</div>
                  <div class="corp-config-grid corp-config-grid--notification">
                    <div class="corp-config-field">
                      <div class="corp-config-field__label">默认通知方式</div>
                      <el-select v-model="corpConfig.notifyMode" placeholder="默认通知方式">
                        <el-option label="旧版工作通知（AgentId + OA）" value="agent" />
                        <el-option label="旧版优先，失败兜底新版" value="agent_fallback" />
                        <el-option label="新版机器人通知（sampleLink）" value="robot" />
                      </el-select>
                    </div>
                    <div v-if="corpConfig.notifyMode !== 'robot'" class="corp-config-field">
                      <div class="corp-config-field__label">AgentId</div>
                      <el-input v-model="corpConfig.agentId" placeholder="旧版 AgentId（数字）" />
                    </div>
                    <div v-if="corpConfig.notifyMode !== 'agent'" class="corp-config-field">
                      <div class="corp-config-field__label">RobotCode</div>
                      <el-input v-model="corpConfig.robotCode" placeholder="留空时使用 AppKey" />
                    </div>
                    <div class="corp-config-field">
                      <div class="corp-config-field__label">App ID（通知跳转）</div>
                      <el-input v-model="corpConfig.unifiedAppId" placeholder="新版应用 App ID" />
                    </div>
                  </div>
                </section>

                <section class="corp-config-section corp-config-section--last">
                  <div class="corp-config-section__title">应用入口</div>
                  <div class="corp-config-field">
                    <div
                      class="corp-config-field__label"
                      :class="{ 'corp-config-field__label--required': corpConfig.enabled === 1 }"
                    >
                      H5 应用地址
                    </div>
                    <el-input v-model="corpConfig.appUrl" placeholder="例如 https://.../?corpId=..." />
                  </div>
                </section>
              </div>
              <el-button class="corp-config-add" icon="Plus" @click="addCorpConfig">新增企业应用</el-button>
            </div>
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

      <div
        class="dingtalk-settings-actions"
        :class="{ 'dingtalk-settings-actions--corp': activeTab === 'corp' }"
      >
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
import type { DingTalkSettings } from '../../api/types'
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
  unifiedAppId: string
  appUrl: string
  notifyEnabled: number
  notifyMode: string
  robotCode: string
  appSecretSet: boolean
  enabled: number
}

type SettingsPayload = Record<string, string | number>

function normalizeNotifyMode(mode?: string) {
  const value = String(mode || '').trim()
  if (value === 'robot' || value === 'agent_fallback' || value === 'agent') {
    return value
  }
  return 'agent'
}

function newCorpConfig(data: Partial<CorpConfigForm> = {}): CorpConfigForm {
  return {
    localKey: data.localKey || `${Date.now()}-${Math.random().toString(16).slice(2)}`,
    corpId: data.corpId || '',
    corpName: data.corpName || '',
    appKey: data.appKey || '',
    agentId: data.agentId || '',
    unifiedAppId: data.unifiedAppId || '',
    appUrl: data.appUrl || '',
    notifyEnabled: Number(data.notifyEnabled) === 1 ? 1 : 0,
    notifyMode: normalizeNotifyMode(data.notifyMode),
    robotCode: data.robotCode || '',
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
        unifiedAppId: item?.unifiedAppId || '',
        appUrl: item?.appUrl || '',
        notifyEnabled: Number(item?.notifyEnabled) === 1 ? 1 : 0,
        notifyMode: item?.notifyMode || '',
        robotCode: item?.robotCode || '',
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
    const res = await request.get<DingTalkSettings>('/api/v2/admin/dingtalk/settings')
    const data = res.data || {}
    form.corpId = data.corpId || ''
    form.appKey = data.appKey || ''
    form.appSecret = ''
    form.appSecretSet = Boolean(data.appSecretSet)
    form.corpConfigs = normalizeCorpConfigs(
      data.corpConfigs && data.corpConfigs.length
        ? data.corpConfigs
        : [{
            corpId: data.corpId || '',
            corpName: data.corpId || '',
            appKey: data.appKey || '',
            agentId: data.agentId || '',
            unifiedAppId: data.unifiedAppId || '',
            appUrl: data.appUrl || '',
            notifyEnabled: Number(data.notifyEnabled) === 1 ? 1 : 0,
            notifyMode: data.notifyMode || '',
            robotCode: data.robotCode || '',
            appSecretSet: data.appSecretSet
          }]
    )
    form.tokenExpire = data.tokenExpire || '168h'
    form.redisPrefix = data.redisPrefix || 'dingtalk_h5_token:'
    form.singleLogin = Number(data.singleLogin) === 1 ? 1 : 0
    form.selfBind = Number(data.selfBind) === 0 ? 0 : 1
    form.appName = data.appName || defaultAppName
    form.logoText = data.logoText || defaultLogoText
    form.logoUrl = data.logoUrl || ''
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  const payload = buildSettingsPayload()
  if (!payload) {
    return
  }
  saving.value = true
  try {
    await request.put('/api/v2/admin/dingtalk/settings', payload)
    ElMessage.success('保存成功')
    await loadSettings()
  } finally {
    saving.value = false
  }
}

function buildSettingsPayload(): SettingsPayload | null {
  switch (activeTab.value) {
    case 'corp':
      return buildCorpSettingsPayload()
    case 'login': {
      const tokenExpire = form.tokenExpire.trim()
      const redisPrefix = form.redisPrefix.trim()
      if (!tokenExpire) {
        ElMessage.warning('请输入 Token 过期时间')
        return null
      }
      if (!redisPrefix) {
        ElMessage.warning('请输入 Redis Key 前缀')
        return null
      }
      return {
        scope: 'login',
        tokenExpire,
        redisPrefix,
        singleLogin: form.singleLogin,
        selfBind: form.selfBind
      }
    }
    case 'app':
      return {
        scope: 'app',
        appName: previewAppName.value,
        logoText: previewLogoText.value,
        logoUrl: form.logoUrl.trim()
      }
    default:
      return buildCorpSettingsPayload()
  }
}

function buildCorpSettingsPayload(): SettingsPayload | null {
  const corpConfigs = form.corpConfigs
    .map(item => ({
      corpId: item.corpId.trim(),
      corpName: item.corpName.trim(),
      appKey: item.appKey.trim(),
      appSecret: item.appSecret.trim(),
      agentId: item.agentId.trim(),
      unifiedAppId: item.unifiedAppId.trim(),
      appUrl: item.appUrl.trim(),
      notifyEnabled: item.notifyEnabled,
      notifyMode: normalizeNotifyMode(item.notifyMode),
      robotCode: item.robotCode.trim(),
      enabled: item.enabled
    }))
    .filter(item => item.corpId || item.corpName || item.appKey || item.appSecret)
  for (const item of corpConfigs) {
    if (!item.corpId) {
      ElMessage.warning('请填写企业 CorpId')
      return null
    }
    if (!item.appKey) {
      ElMessage.warning('请填写企业 AppKey')
      return null
    }
    if (item.enabled === 1 && !item.appUrl) {
      ElMessage.warning(`请填写企业“${item.corpName || item.corpId}”的 H5 应用地址`)
      return null
    }
  }
  const firstCorpConfig = corpConfigs[0] || {
    corpId: '',
    appKey: '',
    appSecret: '',
    agentId: '',
    unifiedAppId: '',
    appUrl: '',
    notifyEnabled: 0,
    notifyMode: 'agent',
    robotCode: ''
  }
  return {
    scope: 'corp',
    corpId: firstCorpConfig.corpId,
    appKey: firstCorpConfig.appKey,
    appSecret: firstCorpConfig.appSecret,
    agentId: firstCorpConfig.agentId,
    unifiedAppId: firstCorpConfig.unifiedAppId,
    appUrl: firstCorpConfig.appUrl,
    notifyEnabled: firstCorpConfig.notifyEnabled,
    notifyMode: firstCorpConfig.notifyMode,
    robotCode: firstCorpConfig.robotCode,
    corpConfigs: JSON.stringify(corpConfigs)
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
  max-width: 1120px;
}

.dingtalk-settings-form {
  max-width: 620px;
  padding-top: 12px;
}

.dingtalk-settings-form--corp {
  max-width: 1080px;
}

.dingtalk-settings-actions {
  display: flex;
  max-width: 620px;
  padding-left: 160px;
}

.dingtalk-settings-actions--corp {
  justify-content: flex-end;
  max-width: 1080px;
  padding-left: 0;
}

.corp-config-list {
  width: 100%;
}

.corp-config-item {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--admin-ui-color-border);
  border-radius: var(--admin-ui-radius-lg);
  background: var(--admin-ui-color-surface);
}

.corp-config-item + .corp-config-item {
  margin-top: 12px;
}

.corp-config-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--admin-ui-space-4);
  min-height: 64px;
  padding: var(--admin-ui-space-3) var(--admin-ui-space-5);
  border-bottom: 1px solid var(--admin-ui-color-border);
  background: var(--admin-ui-color-surface-muted);
}

.corp-config-item__identity {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--admin-ui-space-3);
}

.corp-config-item__index {
  display: inline-flex;
  width: 28px;
  height: 28px;
  flex: 0 0 28px;
  align-items: center;
  justify-content: center;
  border-radius: var(--admin-ui-radius);
  background: var(--admin-ui-color-primary-soft);
  color: var(--admin-ui-color-primary);
  font-size: var(--admin-ui-font-size-caption);
  font-weight: 700;
}

.corp-config-item__title {
  overflow: hidden;
  color: var(--admin-ui-color-text);
  font-size: var(--admin-ui-font-size-body);
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.corp-config-item__meta {
  overflow: hidden;
  margin-top: 2px;
  color: var(--admin-ui-color-text-muted);
  font-size: var(--admin-ui-font-size-caption);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.corp-config-item__actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: var(--admin-ui-space-4);
}

.corp-config-switch {
  display: inline-flex;
  align-items: center;
  gap: var(--admin-ui-space-2);
  color: var(--admin-ui-color-text-secondary);
  font-size: var(--admin-ui-font-size-caption);
}

.corp-config-add {
  margin-top: var(--admin-ui-space-3);
}

.corp-config-section {
  padding: var(--admin-ui-space-4) var(--admin-ui-space-5) var(--admin-ui-space-5);
  border-bottom: 1px solid var(--admin-ui-color-border);
}

.corp-config-section--last {
  border-bottom: 0;
}

.corp-config-section__title {
  margin-bottom: var(--admin-ui-space-3);
  color: var(--admin-ui-color-text);
  font-size: var(--admin-ui-font-size-body);
  font-weight: 700;
}

.corp-config-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--admin-ui-space-4) var(--admin-ui-space-5);
}

.corp-config-grid--notification {
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
}

.corp-config-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.corp-config-field__label {
  color: var(--admin-ui-color-text-secondary);
  font-size: var(--admin-ui-font-size-caption);
  font-weight: 600;
  line-height: var(--admin-ui-line-height-body);
}

.corp-config-field__label--required::before {
  margin-right: 4px;
  color: var(--el-color-danger);
  content: '*';
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

@media (max-width: 767px) {
  .dingtalk-setup-page {
    padding: var(--admin-ui-page-gutter-mobile);
  }

  .dingtalk-settings-form,
  .dingtalk-settings-actions {
    max-width: none;
  }

  .dingtalk-settings-actions {
    justify-content: flex-end;
    padding-left: 0;
  }

  .corp-config-item__head {
    align-items: flex-start;
    flex-direction: column;
    padding: var(--admin-ui-space-3);
  }

  .corp-config-item__actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .corp-config-item__actions :deep(.el-button.is-circle) {
    margin-left: auto;
  }

  .corp-config-section {
    padding: var(--admin-ui-space-4) var(--admin-ui-space-3);
  }

  .corp-config-grid,
  .corp-config-grid--notification {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
