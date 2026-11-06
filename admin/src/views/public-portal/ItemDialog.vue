<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { AdminDialog } from '@/components/admin-ui'
import type { PublicPortalAdminItem, PublicPortalItemInput, PublicPortalSourceOption } from '@/types/publicPortal'
import { activityCategory, canEditItem, itemInputError } from './portalModel'
import SourcePicker from './SourcePicker.vue'

const props = defineProps<{ modelValue: boolean; item: PublicPortalAdminItem | null; editable: boolean; busy: boolean; error: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; save: [input: PublicPortalItemInput] }>()
const visible = computed({ get: () => props.modelValue, set: value => { if (!props.busy) emit('update:modelValue', value) } })
const emptyDraft = (): PublicPortalItemInput => ({ kind: 'survey', sourceId: 0, title: '', description: '', enabled: true, sort: 0 })
const draft = ref<PublicPortalItemInput>(emptyDraft())
const sourceLabel = ref('')
const pickerVisible = ref(true)
const validationError = ref('')
const disabled = computed(() => !props.editable || props.busy || (props.item !== null && !canEditItem(props.editable, props.item)))
watch(() => props.modelValue, opened => {
  if (!opened) return
  const item = props.item
  draft.value = item ? { kind: item.kind, sourceId: item.sourceId, title: item.title, description: item.description, enabled: item.enabled, sort: item.sort } : emptyDraft()
  sourceLabel.value = item ? `${activityCategory(item.kind, item.category)} #${item.sourceId} · ${item.managementTitle || item.displayTitle || '来源不可用'}` : ''
  pickerVisible.value = !item
  validationError.value = ''
}, { immediate: true })
function changeKind() {
  if (disabled.value) return
  draft.value.sourceId = 0
  sourceLabel.value = ''
  pickerVisible.value = true
}
function selectSource(source: PublicPortalSourceOption) {
  if (disabled.value || source.kind !== draft.value.kind) return
  draft.value.sourceId = source.sourceId
  sourceLabel.value = `${activityCategory(source.kind, source.category)} #${source.sourceId} · ${source.managementTitle}（${source.displayTitle || '未设置显示名称'}）`
  pickerVisible.value = false
  validationError.value = ''
}
function submit() {
  if (disabled.value) return
  validationError.value = itemInputError(draft.value)
  if (props.item && !props.item.available && draft.value.sourceId === props.item.sourceId && draft.value.kind === props.item.kind && draft.value.enabled) validationError.value = '来源不可用，请先下架或更换来源后保存'
  if (!validationError.value) emit('save', { ...draft.value, title: draft.value.title.trim(), description: draft.value.description.trim() })
}
</script>

<template>
  <AdminDialog v-model="visible" :title="item ? '编辑展示内容' : '添加展示内容'" width="lg" confirm-text="保存展示内容" :confirm-loading="busy" :confirm-disabled="disabled" :close-on-press-escape="!busy" :show-close="!busy" @confirm="submit">
    <el-alert v-if="error || validationError" :title="validationError || error" type="error" :closable="false" show-icon class="dialog-notice" />
    <el-alert v-if="item && !item.available" :title="item.unavailableReason || '来源已失效，当前不会公开展示'" type="warning" :closable="false" show-icon class="dialog-notice" />
    <el-form :model="draft" label-position="top" :disabled="disabled" @submit.prevent="submit">
      <el-form-item label="内容类型"><el-radio-group v-model="draft.kind" @change="changeKind"><el-radio-button value="survey">问卷</el-radio-button><el-radio-button value="exam">考试</el-radio-button></el-radio-group></el-form-item>
      <el-form-item label="来源内容" required>
        <div class="source-choice"><span>{{ sourceLabel || '尚未选择来源内容' }}</span><el-button :disabled="disabled" @click="pickerVisible = !pickerVisible">{{ pickerVisible ? '收起候选' : '选择 / 更换来源' }}</el-button></div>
      </el-form-item>
    </el-form>
    <SourcePicker v-if="pickerVisible" :kind="draft.kind" :selected-id="draft.sourceId" :disabled="disabled" @select="selectSource" />
    <el-form :model="draft" label-position="top" :disabled="disabled" @submit.prevent="submit">
      <el-form-item label="展示标题" class="title-field"><el-input v-model="draft.title" maxlength="255" show-word-limit placeholder="留空沿用来源显示名称" /></el-form-item>
      <el-form-item label="展示简介"><el-input v-model="draft.description" type="textarea" :rows="3" maxlength="1000" show-word-limit placeholder="留空沿用来源简介，以纯文本展示" /></el-form-item>
      <el-form-item label="排序"><el-input-number v-model="draft.sort" :min="0" :max="99999" :precision="0" /><span class="field-hint">数值越小越靠前；相同时按展示记录 ID 排列。</span></el-form-item>
      <el-form-item label="上架状态"><el-switch v-model="draft.enabled" active-text="上架" inactive-text="下架" /></el-form-item>
    </el-form>
  </AdminDialog>
</template>

<style scoped>
.source-choice { width: 100%; display: flex; flex-wrap: wrap; align-items: center; gap: var(--admin-ui-space-3); overflow-wrap: anywhere; }
.field-hint { margin-left: var(--admin-ui-space-3); color: var(--admin-ui-color-text-secondary); }
.title-field { margin-top: var(--admin-ui-space-4); }
.dialog-notice { margin-bottom: var(--admin-ui-space-4); }
</style>
