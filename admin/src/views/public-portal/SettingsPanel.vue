<script setup lang="ts">
import { ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { PublicPortalSettings } from '@/types/publicPortal'
import { settingsError, settingsWritePayload } from './portalModel'

const props = defineProps<{ settings: PublicPortalSettings; editable: boolean; busy: boolean }>()
const emit = defineEmits<{ save: [settings: PublicPortalSettings] }>()
const draft = ref<PublicPortalSettings>({ ...props.settings })
const form = ref<FormInstance>()
const validationError = ref('')
watch(() => props.settings, value => { draft.value = { ...value }; validationError.value = ''; form.value?.clearValidate() })
const rules: FormRules<PublicPortalSettings> = {
  title: [{ required: true, whitespace: true, message: '请输入页面标题', trigger: 'blur' }, { max: 100, message: '最多 100 字', trigger: 'blur' }],
  description: [{ max: 1000, message: '最多 1000 字', trigger: 'blur' }],
}
async function submit() {
  if (!props.editable || props.busy) return
  if (form.value && !await form.value.validate().catch(() => false)) return
  validationError.value = settingsError(draft.value)
  if (!validationError.value) emit('save', settingsWritePayload(draft.value))
}
</script>

<template>
  <el-form ref="form" :model="draft" :rules="rules" label-position="top" :disabled="!editable || busy" class="portal-settings" @submit.prevent="submit">
    <el-alert v-if="validationError" :title="validationError" type="error" show-icon :closable="false" />
    <el-form-item label="开启公共入口">
      <el-switch v-model="draft.enabled" active-text="开启" inactive-text="关闭" />
      <p class="field-hint">关闭后公众无法浏览。预览可查看已保存的展示内容。</p>
    </el-form-item>
    <el-form-item label="页面标题" prop="title"><el-input v-model="draft.title" maxlength="100" show-word-limit placeholder="例如：问卷与考试" /></el-form-item>
    <el-form-item label="页面简介" prop="description"><el-input v-model="draft.description" type="textarea" :rows="3" maxlength="1000" show-word-limit placeholder="向参与者说明本页用途" /></el-form-item>
    <el-form-item label="Logo URL" prop="logoUrl"><el-input v-model="draft.logoUrl" placeholder="https://… 或 /uploads/…" /><p class="field-hint">支持 HTTP(S) 地址或站内绝对路径，留空使用默认标识。</p></el-form-item>
    <el-form-item label="结束后的展示"><el-switch v-model="draft.showEnded" active-text="保留已结束活动" /><p class="field-hint">关闭时自动隐藏已结束内容。</p></el-form-item>
    <el-button type="primary" native-type="submit" :loading="busy" :disabled="!editable || busy">保存页面设置</el-button>
    <span class="version-note">当前版本 {{ settings.version }}</span>
  </el-form>
</template>

<style scoped>
.portal-settings { max-width: var(--admin-ui-page-max-form); padding: var(--admin-ui-space-5); }
.field-hint { width: 100%; margin: var(--admin-ui-space-1) 0 0; color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); line-height: 1.6; }
.version-note { margin-left: var(--admin-ui-space-3); color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); }
</style>
