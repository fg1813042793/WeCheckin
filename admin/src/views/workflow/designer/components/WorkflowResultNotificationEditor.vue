<template>
  <section class="result-notification-editor">
    <div class="section-title-row">
      <h4>任务处理结果通知</h4>
      <el-switch :model-value="enabled" :disabled="readonly" @change="updateEnabled" />
    </div>

    <template v-if="enabled">
      <label class="field-label">通知渠道</label>
      <el-checkbox-group :model-value="config.channels" :disabled="readonly" class="notification-channels" @change="updateChannels">
        <el-checkbox value="in_app">站内通知</el-checkbox>
        <el-checkbox value="dingtalk_oa">钉钉 OA</el-checkbox>
      </el-checkbox-group>

      <label class="field-label spacing">通知标题</label>
      <el-input
        :model-value="config.title"
        maxlength="256"
        :disabled="readonly"
        @input="updateTitle"
      />

      <label class="field-label spacing">通知正文</label>
      <el-input
        :model-value="config.content"
        type="textarea"
        :rows="5"
        maxlength="2000"
        show-word-limit
        :disabled="readonly"
        @input="updateContent"
      />
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { WorkflowNotificationChannel, WorkflowNotificationConfig } from '../../types'
import { defaultResultNotificationConfig } from '../graph'

const props = withDefaults(defineProps<{
  modelValue?: WorkflowNotificationConfig
  readonly?: boolean
}>(), {
  readonly: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: WorkflowNotificationConfig]
}>()

const defaultConfig = defaultResultNotificationConfig()
const enabled = computed(() => props.modelValue?.enabled === true)
const config = computed(() => ({
  ...defaultConfig,
  ...(props.modelValue || {}),
  channels: [...(props.modelValue?.channels || defaultConfig.channels)],
}))

function update(patch: Partial<WorkflowNotificationConfig>) {
  emit('update:modelValue', { ...config.value, ...patch })
}

function updateEnabled(value: string | number | boolean) {
  update({ enabled: Boolean(value) })
}

function updateChannels(value: unknown) {
  const allowed = new Set<WorkflowNotificationChannel>(['in_app', 'dingtalk_oa'])
  const channels = Array.from(new Set(Array.isArray(value) ? value : []))
    .filter((item): item is WorkflowNotificationChannel => allowed.has(item as WorkflowNotificationChannel))
  update({ channels })
}

function updateTitle(value: string) {
  update({ title: String(value || '') })
}

function updateContent(value: string) {
  update({ content: String(value || '') })
}
</script>

<style scoped>
.result-notification-editor { padding: 18px 0; border-bottom: 1px solid #edf0f5; }
.section-title-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.section-title-row h4 { margin: 0; color: #334155; font-size: 13px; }
.field-label { display: block; margin-bottom: 7px; color: #64748b; font-size: 12px; font-weight: 500; }
.field-label.spacing { margin-top: 14px; }
.notification-channels { display: flex; flex-wrap: wrap; gap: 4px 18px; }
.notification-channels :deep(.el-checkbox) { margin-right: 0; }
</style>
