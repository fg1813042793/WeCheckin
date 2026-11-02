<script setup lang="ts">
import { Refresh } from '@element-plus/icons-vue'
import type { DingTalkSendTemplate } from '@/types/notificationTemplate'
import { targetSummary, templateMessageTypes } from '../templateHelpers'

defineProps<{
  options: DingTalkSendTemplate[]; selected: DingTalkSendTemplate | null; loading: boolean; switching: boolean
  error: string; selectionError: string; edited: boolean; disabled: boolean
}>()
const emit = defineEmits<{ select: [id: number]; retry: [] }>()
</script>

<template>
  <section class="send-template">
    <el-form-item label="发送模板">
      <div class="template-picker">
        <el-select :model-value="selected?.id || 0" aria-label="发送模板" filterable :loading="loading"
          :disabled="disabled || loading || switching || Boolean(error)" @update:model-value="emit('select', $event)">
          <el-option :value="0" label="不使用模板（沿用系统通知样式）" />
          <el-option v-if="selected && !options.some(item => item.id === selected?.id)" :value="selected.id" :label="`${selected.name}（已不可用）`" disabled />
          <el-option v-for="item in options" :key="`${item.id}-${item.version}`" :value="item.id" :label="`${item.name} · v${item.version}`" />
        </el-select>
        <el-button :icon="Refresh" aria-label="刷新模板" :loading="loading" :disabled="disabled || loading || switching" @click="emit('retry')">刷新模板</el-button>
      </div>
    </el-form-item>
    <div v-if="error || selectionError" class="template-warning" role="alert">
      <span>{{ error || selectionError }}</span>
      <el-button v-if="selectionError && !error && selected && options.some(item => item.id === selected?.id)" link type="primary" :disabled="disabled || switching || loading" @click="emit('select', selected.id)">使用最新版本</el-button>
    </div>
    <p v-if="error" class="template-help">{{ selected ? '可刷新模板重试，或选择不使用模板。当前选择和内容已保留。' : '可刷新模板重试，也可直接按系统通知样式发送。' }}</p>
    <el-button v-if="selected && (loading || error || selectionError)" aria-label="不使用模板" link type="primary" :disabled="disabled || switching" @click="emit('select', 0)">不使用模板</el-button>
    <p v-if="!loading && !error && !options.length" class="template-help">暂无启用的模板，可继续使用系统通知样式。</p>
    <div v-if="selected" class="template-summary">
      <strong>已保存模板：{{ selected.name }} · v{{ selected.version }}</strong>
      <span>{{ templateMessageTypes.find(type => type.value === selected?.messageType)?.label }} · {{ targetSummary(selected.target) }}</span>
      <span>{{ edited ? '本次内容已修改，模板保持不变。' : '本次标题和正文来自模板，可在下方修改。' }}</span>
      <el-collapse>
        <el-collapse-item title="查看模板默认内容" name="saved">
          <strong>{{ selected.title }}</strong><p class="saved-content">{{ selected.content }}</p>
        </el-collapse-item>
      </el-collapse>
    </div>
  </section>
</template>

<style scoped>
.send-template { margin-bottom: var(--admin-ui-space-4); }
.template-picker { display: flex; width: 100%; min-width: 0; gap: var(--admin-ui-space-2); }
.template-picker .el-select { flex: 1; min-width: 0; }
.template-picker .el-button { flex-shrink: 0; }
/* The dialog is teleported outside .admin-main; the wrapper retains its focus ring. */
.send-template :deep(.el-select__input:focus),
.send-template :deep(.el-select__input:focus-visible) { outline: none; }
.template-warning { color: var(--admin-ui-color-danger); overflow-wrap: anywhere; }
.template-help { color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); }
.template-summary { display: grid; gap: var(--admin-ui-space-2); padding: var(--admin-ui-space-3); border: 1px solid var(--admin-ui-color-border); border-radius: var(--admin-ui-radius); background: var(--admin-ui-color-surface-muted); color: var(--admin-ui-color-text-secondary); overflow-wrap: anywhere; }
.template-summary strong { color: var(--admin-ui-color-text); }
.saved-content { white-space: pre-wrap; margin: var(--admin-ui-space-2) 0 0; }
</style>
