<template>
  <div class="history-import-rules">
    <el-form label-position="top" @submit.prevent>
      <el-form-item label="启用预设导入规则">
        <el-switch :model-value="config.enabled" :disabled="readonly" active-text="启用" inactive-text="关闭" @update:model-value="setEnabled" />
      </el-form-item>
      <p v-if="!config.enabled" class="history-import-rules__help">未启用时保留原有三步向导；关闭后保留已配置的规则草稿。</p>
      <template v-else>
        <details class="history-import-rules__section" data-panel="basic" :open="basicOpen" @toggle="togglePanel('basic', $event)">
          <summary><span class="history-import-rules__summary"><strong>导入方式</strong><span>{{ ruleSummary }}</span></span></summary>
          <div class="history-import-rules__section-body">
            <el-form-item label="来源方式">
              <el-select :model-value="config.sourceMode" :disabled="readonly" filterable @update:model-value="setSourceMode">
                <el-option label="用户选择一条历史记录后直接导入" value="select_history" />
                <el-option label="自动导入最近一条符合条件的记录" value="latest_history" />
              </el-select>
            </el-form-item>
            <el-form-item label="来源状态">
              <el-select :model-value="config.sourceStatus" :disabled="readonly" filterable @update:model-value="setSourceStatus">
                <el-option label="仅已完成（推荐）" value="completed" />
                <el-option label="所有可见状态" value="all" />
              </el-select>
              <p class="history-import-rules__help">最近记录按发起时间排序；仅已完成时，用户不能扩大状态范围。</p>
            </el-form-item>
            <el-form-item label="字段方式">
              <el-select :model-value="config.mappingMode" :disabled="readonly" filterable @update:model-value="setMappingMode">
                <el-option label="按同名、类型兼容字段匹配" value="same_key" />
                <el-option label="自定义来源 → 目标映射" value="mapping" />
              </el-select>
            </el-form-item>
            <el-form-item label="覆盖已有内容">
              <el-switch :model-value="config.overwrite" :disabled="readonly" active-text="允许覆盖" inactive-text="只填空字段" @update:model-value="setOverwrite" />
              <p class="history-import-rules__policy" role="status">
                {{ config.overwrite ? '导入将覆盖目标字段已有内容，请谨慎启用。' : '仅填充空字段，保留目标字段已有内容。' }}
              </p>
            </el-form-item>
          </div>
        </details>

        <details v-if="config.mappingMode === 'mapping'" class="history-import-rules__section" data-panel="mappings" :open="mappingsOpen" @toggle="togglePanel('mappings', $event)">
          <summary><span class="history-import-rules__summary"><strong>字段映射<span class="history-import-rules__count">{{ config.mappings.length }} 组</span></strong><span>{{ mappingSummary }}</span></span></summary>
          <div class="history-import-rules__section-body">
            <p v-if="!config.mappings.length" class="history-import-rules__help">尚未配置映射，请添加来源和目标字段。</p>
            <div v-for="(mapping, index) in config.mappings" :key="index" class="history-import-rules__mapping">
              <div class="history-import-rules__heading">
                <strong>字段映射 {{ index + 1 }}</strong>
                <el-button link type="danger" :disabled="readonly" :aria-label="`删除字段映射 ${index + 1}`" @click="deleteMapping(index)">删除</el-button>
              </div>
              <div class="history-import-rules__pair">
                <el-form-item label="来源字段">
                  <el-select :model-value="mapping.sourceKey" :disabled="readonly" filterable clearable placeholder="搜索字段名称" @update:model-value="setMapping(index, 'sourceKey', $event)">
                    <el-option v-for="option in sourceOptions(index)" :key="option.value" v-bind="option" />
                  </el-select>
                </el-form-item>
                <el-form-item label="→ 目标字段">
                  <el-select :model-value="mapping.targetKey" :disabled="readonly || !mapping.sourceKey" filterable clearable placeholder="选择可写且兼容的字段" @update:model-value="setMapping(index, 'targetKey', $event)">
                    <el-option v-for="option in targetOptions(index)" :key="option.value" v-bind="option" />
                  </el-select>
                </el-form-item>
              </div>
              <details v-if="isDetailMapping(index)" class="history-import-rules__columns" :open="columnsOpen.includes(index)" @toggle="toggleColumns(index, $event)">
                <summary><span class="history-import-rules__summary"><strong>明细列映射</strong><span>{{ mapping.columns === undefined ? '按同名列导入' : `${mapping.columns.length} 列映射` }}</span></span></summary>
                <div class="history-import-rules__section-body">
                  <p class="history-import-rules__help">未配置时按同名列导入；自定义时仅导入所选列，支持一层明细列，不支持嵌套明细。</p>
                  <div v-for="(column, columnIndex) in mapping.columns" :key="columnIndex" class="history-import-rules__column">
                    <div class="history-import-rules__pair">
                      <el-form-item :label="`来源列 ${columnIndex + 1}`">
                        <el-select :model-value="column.sourceKey" :disabled="readonly" filterable clearable placeholder="搜索来源列" @update:model-value="setColumn(index, columnIndex, 'sourceKey', $event)">
                          <el-option v-for="option in columnOptions(index, columnIndex, 'sourceKey')" :key="option.value" v-bind="option" />
                        </el-select>
                      </el-form-item>
                      <el-form-item :label="`→ 目标列 ${columnIndex + 1}`">
                        <el-select :model-value="column.targetKey" :disabled="readonly || !column.sourceKey" filterable clearable placeholder="选择兼容目标列" @update:model-value="setColumn(index, columnIndex, 'targetKey', $event)">
                          <el-option v-for="option in columnOptions(index, columnIndex, 'targetKey')" :key="option.value" v-bind="option" />
                        </el-select>
                      </el-form-item>
                    </div>
                    <el-button link type="danger" :disabled="readonly" :aria-label="`删除映射 ${index + 1} 的明细列 ${columnIndex + 1}`" @click="removeColumn(index, columnIndex)">删除此列映射</el-button>
                  </div>
                  <el-button :disabled="readonly || (mapping.columns?.length || 0) >= 100" @click="addColumnMapping(index)">添加明细列映射</el-button>
                </div>
              </details>
            </div>
            <el-button :disabled="readonly || config.mappings.length >= 100" @click="addFieldMapping">添加字段映射</el-button>
          </div>
        </details>
      </template>
      <details class="history-import-rules__section" data-panel="help">
        <summary><span class="history-import-rules__summary"><strong>使用说明</strong><span>导入范围、权限与保存规则</span></span></summary>
        <div class="history-import-rules__section-body">
          <p class="history-import-rules__help">来源仅限同一流程中本人发起且当前可见的历史申请，不跨流程、不读取他人记录。</p>
          <p class="history-import-rules__help">只导入当前可写字段，跳过隐藏、只读、计算及展示字段。历史版本字段缺失或类型不兼容时跳过并反馈结果。</p>
          <p class="history-import-rules__help">来源和目标字段各不可重复；最多 100 条字段映射，每组最多 100 条明细列映射。</p>
          <p class="history-import-rules__help">保存后立即生效。已发布流程须使用当前已发布版本中的字段；新增字段请先发布。</p>
        </div>
      </details>
      <el-alert v-if="errors.length" title="请检查导入规则" type="warning" :closable="false" show-icon>
        <ul class="history-import-rules__errors"><li v-for="error in errors" :key="error">{{ error }}</li></ul>
      </el-alert>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { WorkflowDraft } from '@/types/workflow'
import { useWorkflowHistoryImportRules } from '../useWorkflowHistoryImportRules'
import { useWorkflowHistoryImportPanels } from '../useWorkflowHistoryImportPanels'
import { historyImportFields } from '../historyImportRules'

const props = defineProps<{ draft: WorkflowDraft; readonly?: boolean }>()
const emit = defineEmits<{ change: [] }>()
const {
  config, errors, setEnabled, setSourceMode, setSourceStatus, setMappingMode, setOverwrite,
  sourceOptions, targetOptions, columnOptions, isDetailMapping,
  addMapping, removeMapping, setMapping, addColumn, removeColumn, setColumn,
} = useWorkflowHistoryImportRules(props, () => emit('change'))
const { basicOpen, mappingsOpen, columnsOpen, openColumns, removeMapping: removeMappingPanel, togglePanel, toggleColumns } = useWorkflowHistoryImportPanels(config, errors)
const ruleSummary = computed(() => [
  config.value.sourceMode === 'latest_history' ? '最近一次记录' : config.value.sourceMode === 'select_history' ? '手动选择记录' : '来源待配置',
  config.value.sourceStatus === 'completed' ? '仅已完成' : config.value.sourceStatus === 'all' ? '所有可见状态' : '状态待配置',
  config.value.mappingMode === 'mapping' ? '自定义映射' : config.value.mappingMode === 'same_key' ? '自动匹配' : '字段待配置',
  config.value.overwrite ? '覆盖已有内容' : '只填空字段',
].join(' · '))
const mappingSummary = computed(() => {
  const first = config.value.mappings[0]
  if (!first) return '尚未配置映射，展开后添加'
  const fields = historyImportFields(props.draft)
  const source = fields.find(item => item.field.key === first.sourceKey)?.label || '来源待选择'
  const target = fields.find(item => item.field.key === first.targetKey)?.label || '目标待选择'
  return `${source} → ${target}${config.value.mappings.length > 1 ? ' 等映射' : ''}`
})
function addFieldMapping() {
  const count = config.value.mappings.length
  addMapping()
  if (config.value.mappings.length > count) mappingsOpen.value = true
}
function deleteMapping(index: number) {
  if (props.readonly || !config.value.mappings[index]) return
  removeMappingPanel(index)
  removeMapping(index)
}
function addColumnMapping(index: number) {
  const count = config.value.mappings[index]?.columns?.length || 0
  addColumn(index)
  if ((config.value.mappings[index]?.columns?.length || 0) > count) openColumns(index)
}
</script>

<style scoped>
.history-import-rules { min-width: 0; }
.history-import-rules__section { border-top: 1px solid var(--admin-ui-color-border); }
.history-import-rules summary { display: flex; align-items: center; gap: var(--admin-ui-space-3); padding: var(--admin-ui-space-3) 0; cursor: pointer; list-style: none; color: var(--admin-ui-color-text); }
.history-import-rules summary::-webkit-details-marker { display: none; }
.history-import-rules summary::before { content: ''; flex: 0 0 6px; width: 6px; height: 6px; border-right: 1.5px solid currentColor; border-bottom: 1.5px solid currentColor; transform: rotate(-45deg); transition: transform 160ms ease; }
.history-import-rules details[open] > summary::before { transform: rotate(45deg); }
.history-import-rules summary:focus-visible { outline: 2px solid var(--admin-ui-color-primary); outline-offset: 2px; border-radius: var(--admin-ui-radius-sm); }
.history-import-rules__summary { display: grid; min-width: 0; gap: var(--admin-ui-space-1); overflow-wrap: anywhere; }
.history-import-rules__summary > span { color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); line-height: 1.5; }
.history-import-rules__count { margin-left: var(--admin-ui-space-2); color: var(--admin-ui-color-text-muted); font-size: var(--admin-ui-font-size-caption); font-weight: 400; }
.history-import-rules__section-body { padding: var(--admin-ui-space-2) 0 var(--admin-ui-space-4); }
.history-import-rules__help { width: 100%; margin: var(--admin-ui-space-2) 0 var(--admin-ui-space-3); color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); line-height: 1.6; overflow-wrap: anywhere; }
.history-import-rules__policy { width: 100%; margin: var(--admin-ui-space-2) 0 0; color: var(--admin-ui-color-text); line-height: 1.6; }
.history-import-rules__mapping { margin: var(--admin-ui-space-4) 0; padding: var(--admin-ui-space-3); border: 1px solid var(--admin-ui-color-border); border-radius: var(--admin-ui-radius); }
.history-import-rules__heading { display: flex; align-items: center; justify-content: space-between; gap: var(--admin-ui-space-2); margin-bottom: var(--admin-ui-space-3); }
.history-import-rules__pair { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--admin-ui-space-3); }
.history-import-rules__pair :deep(.el-form-item) { min-width: 0; margin-bottom: var(--admin-ui-space-3); }
.history-import-rules__columns { border-top: 1px solid var(--admin-ui-color-border); }
.history-import-rules__column { margin-bottom: var(--admin-ui-space-3); }
.history-import-rules__errors { margin: 0; padding-left: var(--admin-ui-space-4); }
.history-import-rules :deep(.el-select) { width: 100%; }
@media (max-width: 767px) { .history-import-rules__pair { grid-template-columns: minmax(0, 1fr); gap: 0; } }
</style>
