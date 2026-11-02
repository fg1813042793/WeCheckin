<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { AdminDrawer } from '@/components/admin-ui'
import type { DingTalkSendTemplateInput } from '@/types/notificationTemplate'
import { newTemplateDraft, templateHeadColorFromPicker, templateHeadColorPickerValue, templateMessageTypes, validateTemplate } from '../templateHelpers'
import TemplateWorkflowSelect from './TemplateWorkflowSelect.vue'

const props = defineProps<{
  modelValue: boolean; draft: DingTalkSendTemplateInput; editing: boolean; canEdit: boolean; saving: boolean; error: string
}>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; save: [value: DingTalkSendTemplateInput] }>()
const form = reactive(newTemplateDraft())
const headColorPicker = computed({
  get: () => templateHeadColorPickerValue(form.headColor),
  set: value => { form.headColor = templateHeadColorFromPicker(value) },
})
const formRef = ref<FormInstance>()
const validating = ref(false)
const workflowReady = ref(false)
const locked = computed(() => props.saving || validating.value || !props.canEdit)
const fields = ['name', 'title', 'content', 'messageType', 'enabled', 'headColor', 'buttonTitle', 'target.kind', 'target.definitionId', 'target.url']
const rules: FormRules = Object.fromEntries(fields.map(field => [field, [{
  validator: (_rule: unknown, _value: unknown, callback: (error?: Error) => void) => {
    const error = validateTemplate(form)[field]
    callback(error ? new Error(error) : undefined)
  }, trigger: ['blur', 'change'],
}]]))
watch(() => props.modelValue, async visible => {
  if (!visible) return
  Object.assign(form, newTemplateDraft(props.draft))
  workflowReady.value = false
  await nextTick()
  formRef.value?.clearValidate()
}, { immediate: true })
function setTarget(kind: DingTalkSendTemplateInput['target']['kind']) {
  form.target = { kind }
  workflowReady.value = false
  formRef.value?.clearValidate(['target.definitionId', 'target.url'])
}
async function save() {
  if (locked.value) return
  validating.value = true
  try {
    const valid = await formRef.value?.validate().catch(() => false)
    if (!valid || !props.canEdit || props.saving || (form.target.kind === 'workflow_start' && !workflowReady.value)) return
    emit('save', newTemplateDraft(form))
  } finally { validating.value = false }
}
function close(value: boolean) { if ((!props.saving && !validating.value) || value) emit('update:modelValue', value) }
</script>

<template>
  <AdminDrawer :model-value="modelValue" :title="editing ? '编辑发送模板' : '新建发送模板'" size="lg"
    :show-close="!saving" :close-on-press-escape="!saving" @update:model-value="close">
    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon class="editor-alert" />
    <el-alert v-if="!canEdit" title="没有编辑模板的权限" type="warning" :closable="false" />
    <el-form ref="formRef" class="template-editor-form" :model="form" :rules="rules" label-position="top" :disabled="locked" @submit.prevent="save">
      <el-form-item label="模板名称" prop="name" required><el-input v-model="form.name" maxlength="80" show-word-limit /></el-form-item>
      <el-form-item label="默认标题" prop="title" required><el-input v-model="form.title" maxlength="255" show-word-limit /></el-form-item>
      <el-form-item label="默认正文" prop="content" required><el-input v-model="form.content" type="textarea" :rows="4" maxlength="5000" show-word-limit /></el-form-item>
      <el-form-item label="消息类型" prop="messageType" required>
        <el-select v-model="form.messageType"><el-option v-for="type in templateMessageTypes" :key="type.value" :value="type.value" :label="type.label" /></el-select>
      </el-form-item>
      <el-form-item label="跳转目标" prop="target.kind" required>
        <el-select :model-value="form.target.kind" @update:model-value="setTarget">
          <el-option value="app_home" label="应用首页" /><el-option value="workflow_start" label="发起指定流程" /><el-option value="url" label="自定义 HTTPS 地址" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="form.target.kind === 'workflow_start'" label="已发布流程" prop="target.definitionId" required>
        <TemplateWorkflowSelect v-model="form.target.definitionId" :display-name="form.target.definitionName" :disabled="locked" @ready="workflowReady = $event" />
        <span class="editor-help">收件人打开后，仍需具备该流程的发起权限。</span>
      </el-form-item>
      <el-form-item v-if="form.target.kind === 'url'" label="HTTPS 地址" prop="target.url" required>
        <el-input v-model="form.target.url" maxlength="2048" placeholder="https://example.com/page" />
      </el-form-item>
      <el-form-item v-if="form.messageType === 'oa'" label="OA 顶部颜色" prop="headColor" required>
        <el-color-picker v-model="headColorPicker" :disabled="locked" />
        <span class="editor-help">选择后自动保存为钉钉要求的不透明 8 位颜色，例如 FF1677FF。</span>
      </el-form-item>
      <el-form-item v-if="form.messageType === 'action_card'" label="卡片按钮文字" prop="buttonTitle" required>
        <el-input v-model="form.buttonTitle" maxlength="20" show-word-limit />
        <span class="editor-help">按钮颜色由钉钉客户端固定，不支持自定义。</span>
      </el-form-item>
      <el-form-item label="启用模板" prop="enabled"><el-switch v-model="form.enabled" /><span class="editor-help">停用后不再出现在发送候选中</span></el-form-item>
    </el-form>
    <template #footer>
      <el-button :disabled="saving || validating" @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="saving || validating" :disabled="!canEdit || (form.target.kind === 'workflow_start' && !workflowReady)" @click="save">保存模板</el-button>
    </template>
  </AdminDrawer>
</template>

<style scoped>
.editor-alert { margin-bottom: var(--admin-ui-space-4); }
.editor-help { margin: var(--admin-ui-space-1) 0 0 var(--admin-ui-space-1); color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); }
/* Keep Element Plus wrapper focus styling without the global inner tabindex outline. */
.template-editor-form :deep(.el-select__input:focus),
.template-editor-form :deep(.el-select__input:focus-visible) { outline: none; }
</style>
