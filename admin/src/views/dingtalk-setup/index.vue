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
                    <div class="corp-config-item__actions">
                      <el-button
                        size="small"
                        :loading="testingCorpKey === corpConfig.localKey"
                        :disabled="!canTestNotification(corpConfig)"
                        @click="testCorpNotification(corpConfig)"
                      >
                        测试通知
                      </el-button>
                      <el-button
                        v-if="form.corpConfigs.length > 1"
                        text
                        type="danger"
                        @click="removeCorpConfig(index)"
                      >
                        删除
                      </el-button>
                    </div>
                  </div>
                  <el-row :gutter="12">
                    <el-col :span="12">
                      <div class="corp-config-field">
                        <div class="corp-config-field__label">企业 CorpId</div>
                        <el-input v-model="corpConfig.corpId" placeholder="钉钉企业 CorpId" />
                      </div>
                    </el-col>
                    <el-col :span="12">
                      <div class="corp-config-field">
                        <div class="corp-config-field__label">企业名称</div>
                        <el-input v-model="corpConfig.corpName" placeholder="企业名称（选填）" />
                      </div>
                    </el-col>
                  </el-row>
                  <el-row :gutter="12">
                    <el-col :span="12">
                      <div class="corp-config-field">
                        <div class="corp-config-field__label">应用 AppKey</div>
                        <el-input v-model="corpConfig.appKey" placeholder="钉钉内部应用 AppKey" />
                      </div>
                    </el-col>
                    <el-col :span="12">
                      <div class="corp-config-field">
                        <div class="corp-config-field__label">应用 AppSecret</div>
                        <el-input
                          v-model="corpConfig.appSecret"
                          type="password"
                          show-password
                          :placeholder="corpConfig.appSecretSet ? '已保存，留空表示不修改' : '请输入 AppSecret'"
                        />
                      </div>
                    </el-col>
                  </el-row>
                  <div class="corp-config-mode">
                    <div class="corp-config-mode__head">
                      <span>默认通知方式</span>
                      <span class="settings-help">App ID 只用于通知点击打开应用，不决定发送通道。</span>
                    </div>
                    <el-row :gutter="12">
                      <el-col :span="6">
                        <div class="corp-config-field">
                          <div class="corp-config-field__label">通知方式</div>
                          <el-select v-model="corpConfig.notifyMode" placeholder="默认通知方式">
                            <el-option label="旧版工作通知（AgentId + OA）" value="agent" />
                            <el-option label="旧版优先，失败兜底新版" value="agent_fallback" />
                            <el-option label="新版机器人通知（sampleLink）" value="robot" />
                          </el-select>
                        </div>
                      </el-col>
                      <el-col v-if="corpConfig.notifyMode !== 'robot'" :span="6">
                        <div class="corp-config-field">
                          <div class="corp-config-field__label">AgentId</div>
                          <el-input v-model="corpConfig.agentId" placeholder="旧版 AgentId（数字）" />
                        </div>
                      </el-col>
                      <el-col v-if="corpConfig.notifyMode !== 'agent'" :span="6">
                        <div class="corp-config-field">
                          <div class="corp-config-field__label">RobotCode</div>
                          <el-input v-model="corpConfig.robotCode" placeholder="留空默认使用 AppKey" />
                        </div>
                      </el-col>
                      <el-col :span="corpConfig.notifyMode === 'agent_fallback' ? 6 : 12">
                        <div class="corp-config-field">
                          <div class="corp-config-field__label">App ID</div>
                          <el-input v-model="corpConfig.unifiedAppId" placeholder="新版 App ID，用于通知跳转" />
                        </div>
                      </el-col>
                    </el-row>
                    <div class="settings-help">
                      选择旧版工作通知时只发送 AgentId + OA 消息；如需旧版失败后自动尝试新版 sampleLink，请选择旧版优先兜底。
                    </div>
                  </div>
                  <el-row :gutter="12">
                    <el-col :span="24">
                      <div class="corp-config-field">
                        <div class="corp-config-field__label">H5 应用地址</div>
                        <el-input
                          v-model="corpConfig.appUrl"
                          placeholder="例如 https://...，通知点击时优先打开该企业应用地址"
                        />
                      </div>
                    </el-col>
                  </el-row>
                  <div class="corp-config-notify">
                    <div class="corp-config-notify__main">
                      <span class="corp-config-notify__label">钉钉通知</span>
                      <el-switch
                        v-model="corpConfig.notifyEnabled"
                        :active-value="1"
                        :inactive-value="0"
                        active-text="开启"
                        inactive-text="关闭"
                        inline-prompt
                      />
                    </div>
                    <span class="settings-help">开启后，可通过该企业应用发送流程提醒和管理后台手动通知。</span>
                  </div>
                  <div class="corp-config-enabled">
                    <span class="corp-config-enabled__label">企业应用状态</span>
                    <el-switch
                      v-model="corpConfig.enabled"
                      :active-value="1"
                      :inactive-value="0"
                      active-text="启用"
                      inactive-text="停用"
                      inline-prompt
                    />
                  </div>
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
            <el-form-item label="H5 应用地址">
              <el-input v-model="form.appUrl" placeholder="https://...，用于绩效通知点击跳转" />
              <div class="settings-help">通知会自动追加 reviewNo、view 等参数，点击后进入对应考评单。</div>
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

      <el-dialog
        v-model="diagnosisDialogVisible"
        title="通知诊断结果"
        width="860px"
        class="dingtalk-diagnosis-dialog"
      >
        <template v-if="diagnosisResult">
          <el-alert
            :type="diagnosisResult.success ? 'success' : 'error'"
            :title="diagnosisResult.success ? '测试通知发送成功' : '测试通知发送失败'"
            :description="diagnosisResult.conclusion"
            show-icon
            :closable="false"
          />
          <el-descriptions class="diagnosis-summary" :column="2" border>
            <el-descriptions-item label="企业">{{ diagnosisResult.corpName || diagnosisResult.corpId }}</el-descriptions-item>
            <el-descriptions-item label="通知方式">{{ diagnosisResult.notifyMode }}</el-descriptions-item>
            <el-descriptions-item label="AppKey">{{ diagnosisResult.appKeyMasked }}</el-descriptions-item>
            <el-descriptions-item label="AppSecret">{{ diagnosisResult.appSecretSet ? '已配置' : '未配置' }}</el-descriptions-item>
            <el-descriptions-item label="AgentId">{{ diagnosisResult.agentId || '-' }}</el-descriptions-item>
            <el-descriptions-item label="接收人">{{ diagnosisResult.recipientUserId || '-' }}</el-descriptions-item>
          </el-descriptions>

          <div class="diagnosis-section-title">调用链路</div>
          <div class="diagnosis-steps">
            <div
              v-for="(step, index) in diagnosisResult.steps"
              :key="`${step.name}-${index}`"
              class="diagnosis-step"
            >
              <div class="diagnosis-step__head">
                <span>{{ index + 1 }}. {{ step.name }}</span>
                <el-tag
                  size="small"
                  :type="step.status === 'success' ? 'success' : step.status === 'skipped' ? 'info' : 'danger'"
                >
                  {{ step.status }}
                </el-tag>
              </div>
              <div v-if="step.endpoint" class="diagnosis-step__endpoint">
                {{ step.method }} {{ step.endpoint }}
                <span v-if="step.durationMs">耗时 {{ step.durationMs }}ms</span>
              </div>
              <div class="diagnosis-step__grid">
                <div>
                  <div class="diagnosis-step__label">请求</div>
                  <pre>{{ formatJSON(step.request) }}</pre>
                </div>
                <div>
                  <div class="diagnosis-step__label">响应</div>
                  <pre>{{ formatJSON(step.response || (step.error ? { error: step.error } : {})) }}</pre>
                </div>
              </div>
            </div>
          </div>
        </template>
        <template #footer>
          <el-button @click="diagnosisDialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="copyDiagnosisEvidence">复制证据</el-button>
        </template>
      </el-dialog>
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

type DiagnosisStep = {
  name: string
  status: string
  method?: string
  endpoint?: string
  durationMs?: number
  request?: Record<string, unknown>
  response?: Record<string, unknown>
  error?: string
}

type DiagnosisResult = {
  success: boolean
  checkedAt: string
  corpId: string
  corpName: string
  notifyEnabled: number
  notifyMode: string
  appKeyMasked: string
  appSecretSet: boolean
  agentId: string
  unifiedAppId: string
  robotCodeMasked: string
  recipientUserId: string
  conclusion: string
  steps: DiagnosisStep[]
}

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
  logoUrl: '',
  appUrl: ''
})
const loading = ref(false)
const saving = ref(false)
const testingCorpKey = ref('')
const diagnosisDialogVisible = ref(false)
const diagnosisResult = ref<DiagnosisResult | null>(null)

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
    form.appUrl = data.appUrl || ''
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

function canTestNotification(corpConfig: CorpConfigForm) {
  const hasButtonPerm = hasPerm('admin:menu:dingtalk:config:test') || hasPerm('admin:menu:dingtalk:config:edit')
  return hasButtonPerm && Boolean(corpConfig.corpId.trim()) && testingCorpKey.value === ''
}

async function testCorpNotification(corpConfig: CorpConfigForm) {
  const corpId = corpConfig.corpId.trim()
  if (!corpId) {
    ElMessage.warning('请先填写企业 CorpId')
    return
  }
  testingCorpKey.value = corpConfig.localKey
  try {
    const res = await request.post<DiagnosisResult>('/api/v2/admin/dingtalk/settings/notification-test', { corpId })
    diagnosisResult.value = res.data
    diagnosisDialogVisible.value = true
    if (res.data.success) {
      ElMessage.success('测试通知发送成功')
    } else {
      ElMessage.warning('测试通知失败，已生成诊断结果')
    }
  } finally {
    testingCorpKey.value = ''
  }
}

function formatJSON(value?: Record<string, unknown>) {
  return JSON.stringify(value || {}, null, 2)
}

async function copyDiagnosisEvidence() {
  if (!diagnosisResult.value) return
  const text = JSON.stringify(diagnosisResult.value, null, 2)
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
    } else {
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    ElMessage.success('已复制诊断证据')
  } catch {
    ElMessage.warning('复制失败，请手动复制诊断内容')
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
        logoUrl: form.logoUrl.trim(),
        appUrl: form.appUrl.trim()
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

.corp-config-item__actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.corp-config-add {
  margin-top: 12px;
}

.corp-config-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.corp-config-field__label,
.corp-config-enabled__label {
  color: #475569;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
}

.corp-config-mode {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 2px 0 4px;
}

.corp-config-mode__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: #374151;
  font-size: 13px;
  font-weight: 600;
}

.corp-config-notify,
.corp-config-notify__main {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 32px;
}

.corp-config-notify {
  flex-wrap: wrap;
  justify-content: space-between;
  padding: 4px 2px;
}

.corp-config-enabled {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 32px;
  padding: 2px 2px 0;
}

.corp-config-notify__label {
  color: #374151;
  font-size: 13px;
  font-weight: 600;
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

.diagnosis-summary {
  margin-top: 14px;
}

.diagnosis-section-title {
  margin: 18px 0 10px;
  color: #1f2937;
  font-size: 15px;
  font-weight: 700;
}

.diagnosis-steps {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 440px;
  overflow: auto;
}

.diagnosis-step {
  padding: 12px;
  border: 1px solid #e5eaf3;
  border-radius: 8px;
  background: #f8fafc;
}

.diagnosis-step__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: #1f2937;
  font-size: 14px;
  font-weight: 700;
}

.diagnosis-step__endpoint {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 8px;
  color: #64748b;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  word-break: break-all;
}

.diagnosis-step__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 12px;
  margin-top: 10px;
}

.diagnosis-step__label {
  margin-bottom: 6px;
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
}

.diagnosis-step pre {
  min-height: 64px;
  max-height: 180px;
  margin: 0;
  overflow: auto;
  padding: 10px;
  border: 1px solid #e5eaf3;
  border-radius: 6px;
  background: #fff;
  color: #334155;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}

:deep(.dingtalk-diagnosis-dialog .el-dialog__body) {
  padding-top: 12px;
}
</style>
