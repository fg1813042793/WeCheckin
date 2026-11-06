<template>
  <div class="post-stat-settings">
    <el-divider />
    <el-form-item label="是否计算总分">
      <el-switch v-model="rule.calculateTotalScore" active-text="计算总分" inactive-text="不展示总分" />
      <div class="field-help">关闭后考试仍正常阅卷并保存成绩，只是不在交卷后统计通知中展示得分与满分。</div>
    </el-form-item>
    <el-form-item label="通知方式">
      <el-radio-group v-model="rule.notifyChannel">
        <el-radio value="internal">站内通知</el-radio>
        <el-radio value="webhook">Webhook</el-radio>
        <el-radio value="both">全部</el-radio>
      </el-radio-group>
    </el-form-item>
    <template v-if="['webhook', 'both'].includes(rule.notifyChannel || '')">
      <el-form-item label="Webhook 类型">
        <el-select v-model="rule.webhookType" class="full-width">
          <el-option label="钉钉" value="dingtalk" /><el-option label="企业微信" value="wecom" />
          <el-option label="飞书" value="lark" /><el-option label="自定义" value="custom" />
        </el-select>
      </el-form-item>
      <el-form-item label="Webhook URL"><el-input v-model="rule.webhookUrl" placeholder="https://example.com/webhook" clearable /></el-form-item>
    </template>
    <el-form-item label="通知消息模板">
      <el-select v-model="selectedTemplate" class="full-width template-select" placeholder="模版预设" clearable @change="applyTemplate">
        <el-option-group label="内置模版"><el-option v-for="item in builtinPresets" :key="item.key" :label="item.label" :value="item.key" /></el-option-group>
        <el-option-group v-if="customPresets.length" label="自定义模版"><el-option v-for="item in customPresets" :key="item.key" :label="item.label" :value="item.key" /></el-option-group>
      </el-select>
      <el-input v-model="rule.messageTemplate" type="textarea" :rows="4" :placeholder="defaultTemplate" />
      <div class="field-help">变量：<code>{title}</code>、<code>{submitter}</code>、<code>{date}</code>、<code>{total}</code>、<code>{score}</code>、<code>{fullScore}</code>、<code>{result}</code>、<code>{answerDownloadUrl}</code> 答卷下载链接</div>
    </el-form-item>
    <el-form-item label="通知对象">
      <el-checkbox v-model="rule.notifyAdmin" class="notify-admin">通知管理员</el-checkbox>
      <el-input v-model="rule.notifyUserIds" placeholder="指定成员 ID（多个用逗号隔开）" clearable />
      <div v-if="destinationMissing" class="field-help field-help-danger">请配置站内通知对象或 Webhook 地址。</div>
    </el-form-item>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { adminApi, type TemplatePreset } from '../../../api'
import type { ExamLogicRule } from '../exam-designer-types'

const rule = defineModel<ExamLogicRule>({ required: true })
const defaultTemplate = '📝 考试「{title}」收到新答卷\n提交人：{submitter}　时间：{date}\n共 {total} 份提交\n\n{result}'
const selectedTemplate = ref('')
const builtin = ref<TemplatePreset[]>([])
const custom = ref<TemplatePreset[]>([])
const builtinPresets = computed(() => builtin.value.map((item, index) => ({ ...item, key: `builtin:${index}` })))
const customPresets = computed(() => custom.value.map((item, index) => ({ ...item, key: `custom:${index}` })))
const destinationMissing = computed(() => {
  const channel = rule.value.notifyChannel || ''
  const internal = ['internal', 'both'].includes(channel) && (rule.value.notifyAdmin || String(rule.value.notifyUserIds || '').trim())
  const webhook = ['webhook', 'both'].includes(channel) && String(rule.value.webhookUrl || '').trim()
  return !internal && !webhook
})
function applyTemplate(key: string) {
  const source = key.startsWith('custom:') ? custom.value : builtin.value
  const index = Number(key.split(':')[1])
  rule.value.messageTemplate = source[index]?.value || ''
}
onMounted(async () => {
  try {
    const response = await adminApi.examNotificationTemplatesGet()
    builtin.value = response.data?.builtin || []
    custom.value = response.data?.custom || []
  } catch { builtin.value = []; custom.value = [] }
})
</script>

<style scoped>
.post-stat-settings { width:100%; }.full-width { width:100%; }.template-select { margin-bottom:8px; }.notify-admin { display:block; margin-bottom:8px; }
.field-help { margin-top:4px; color:#98a2b3; font-size:11px; line-height:1.6; }.field-help-danger { color:#d93025; }
</style>
