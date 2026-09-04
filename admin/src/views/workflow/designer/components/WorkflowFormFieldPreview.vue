<template>
  <div class="nested-field-preview">
    <h3 v-if="field.type === 'label'">{{ field.label }}</h3>
    <p v-else-if="field.type === 'description'">{{ field.content || '请输入说明内容' }}</p>
    <el-button v-else-if="field.type === 'button'" size="small" disabled>{{ field.label }}</el-button>
    <div v-else-if="field.type === 'detail_list'" class="detail-preview">
      <div class="detail-preview__grid">
        <div
          v-for="column in field.columns || []"
          :key="column.key"
          class="detail-preview__cell"
          :style="{ gridColumn: `span ${columnSpan(column)}` }"
        >
          <span class="detail-preview__label">{{ column.label || column.key }}</span>
          <span class="detail-preview__control">{{ column.placeholder || '请输入' }}</span>
        </div>
      </div>
      <el-button size="small" icon="Plus" disabled>新增行</el-button>
    </div>
    <el-input-number
      v-else-if="field.type === 'number' || field.type === 'amount'"
      :model-value="typeof field.default === 'number' ? field.default : undefined"
      :precision="field.type === 'amount' ? 2 : undefined"
      controls-position="right"
      disabled
    />
    <el-tree-select
      v-else-if="field.type === 'select' || field.type === 'multi_select'"
      :model-value="field.default"
      :data="options"
      :multiple="field.type === 'multi_select'"
      :props="optionTreeProps"
      node-key="value"
      check-strictly
      :placeholder="field.placeholder || '请选择'"
      disabled
    />
    <el-radio-group v-else-if="field.type === 'radio'" :model-value="field.default" disabled>
      <el-radio v-for="option in flatOptions" :key="option.value" :value="option.value">{{ option.label }}</el-radio>
    </el-radio-group>
    <el-checkbox-group v-else-if="field.type === 'checkbox'" :model-value="arrayDefault" disabled>
      <el-checkbox v-for="option in flatOptions" :key="option.value" :value="option.value">{{ option.label }}</el-checkbox>
    </el-checkbox-group>
    <el-switch v-else-if="field.type === 'boolean'" :model-value="Boolean(field.default)" disabled />
    <el-button v-else-if="field.type === 'attachment'" icon="Upload" disabled>选择附件</el-button>
    <el-input
      v-else-if="field.type === 'textarea'"
      :model-value="stringDefault"
      :placeholder="field.placeholder || '请输入内容'"
      type="textarea"
      :autosize="textareaAutosize(field, 3, 8)"
      resize="none"
      disabled
    />
    <el-input v-else :model-value="stringDefault" :placeholder="field.placeholder || '请选择或输入'" disabled />
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import type { WorkflowFormField } from '../../types'
import { flattenWorkflowOptions, normalizeWorkflowOptions, workflowTextareaAutosize as textareaAutosize } from '../../runtimeForm'

const props = defineProps<{ field: WorkflowFormField }>()
const optionTreeProps = { label: 'label', value: 'value', children: 'children' }
const options = computed(() => normalizeWorkflowOptions(props.field.options || [], props.field.optionSource))
const flatOptions = computed(() => flattenWorkflowOptions(options.value))
const stringDefault = computed(() => typeof props.field.default === 'string' ? props.field.default : '')
const arrayDefault = computed(() => Array.isArray(props.field.default) ? props.field.default : [])

function columnSpan(column: WorkflowFormField) {
  const span = Number(column.span || 24)
  return [6, 8, 12, 24].includes(span) ? span : 24
}
</script>

<style scoped>
.nested-field-preview { max-width: 720px; margin: 12px 0 0 45px; pointer-events: none; }
.nested-field-preview > h3 { margin: 0; color: #273548; font-size: 15px; }
.nested-field-preview > p { margin: 0; color: #64748b; font-size: 12px; line-height: 1.7; white-space: pre-wrap; }
.nested-field-preview :deep(.el-select),
.nested-field-preview :deep(.el-input-number),
.nested-field-preview :deep(.el-date-editor) { width: 100%; }
.nested-field-preview :deep(.el-radio-group),
.nested-field-preview :deep(.el-checkbox-group) { display: flex; flex-wrap: wrap; gap: 8px 18px; }
.nested-field-preview :deep(.el-radio),
.nested-field-preview :deep(.el-checkbox) { margin-right: 0; }
.detail-preview { display: flex; flex-direction: column; gap: 8px; padding: 10px; border: 1px solid #e6ebf1; border-radius: 6px; background: #f8fafc; }
.detail-preview__grid {
  display: grid;
  grid-template-columns: repeat(24, minmax(0, 1fr));
  gap: 8px;
}
.detail-preview__cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}
.detail-preview__label {
  min-width: 0;
  overflow: hidden;
  color: #64748b;
  font-size: 11px;
  font-weight: 650;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.detail-preview__control {
  display: block;
  min-height: 28px;
  padding: 6px 8px;
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 5px;
  color: #64748b;
  background: #fff;
  font-size: 11px;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}
@media (max-width: 1120px) {
  .nested-field-preview { margin-left: 0; }
}
</style>
