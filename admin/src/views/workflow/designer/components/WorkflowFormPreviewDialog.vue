<template>
  <el-dialog
    :model-value="modelValue"
    :title="previewTitle"
    width="min(1000px, calc(100vw - 32px))"
    append-to-body
    destroy-on-close
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="form-preview-toolbar">
      <el-radio-group v-model="previewMode" size="small" aria-label="预览设备">
        <el-radio-button value="desktop">
          <el-icon><Monitor /></el-icon>
          <span>桌面端</span>
        </el-radio-button>
        <el-radio-button value="mobile">
          <el-icon><Cellphone /></el-icon>
          <span>移动端</span>
        </el-radio-button>
      </el-radio-group>
    </div>

    <div class="form-preview-stage">
      <main
        class="form-preview-sheet"
        :class="{ 'form-preview-sheet--mobile': previewMode === 'mobile' }"
      >
        <header class="form-preview-heading">
          <h2>{{ title || draft.name || '未命名流程' }}</h2>
        </header>
        <WorkflowRuntimeForm
          v-model="previewData"
          :fields="draft.form"
          :field-access="previewFieldAccess"
          :field-actions="previewFieldActions"
          empty-text="当前流程未配置表单字段"
        />
      </main>
    </div>

    <template #footer>
      <el-button @click="emit('update:modelValue', false)">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import WorkflowRuntimeForm from '../../components/WorkflowRuntimeForm.vue'
import {
  initialWorkflowFormData,
  workflowFieldActionMap,
  workflowFieldAccessMap,
  type WorkflowFormData,
} from '../../runtimeForm'
import type { WorkflowDraft } from '../../types'

const props = defineProps<{
  modelValue: boolean
  draft: WorkflowDraft
  title?: string
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void
}>()

const previewMode = ref<'desktop' | 'mobile'>('desktop')
const previewData = ref<WorkflowFormData>({})
const previewTitle = computed(() => props.title ? `表单预览 · ${props.title}` : '表单预览')
const startNode = computed(() => props.draft.nodes.find(node => node.type === 'start'))
const permissionsByNode = computed(() => {
  const node = startNode.value
  return node ? { [node.id]: node.formPermissions || [] } : {}
})
const previewFieldAccess = computed(() => workflowFieldAccessMap(
  props.draft.form,
  permissionsByNode.value,
  startNode.value?.id,
  'write',
))
const previewFieldActions = computed(() => workflowFieldActionMap(
  props.draft.form,
  permissionsByNode.value,
  startNode.value?.id,
  { add: true, delete: true },
))

watch(() => props.modelValue, (visible) => {
  if (!visible) return
  previewMode.value = 'desktop'
  previewData.value = initialWorkflowFormData(props.draft.form)
})
</script>

<style scoped>
.form-preview-toolbar {
  display: flex;
  justify-content: flex-end;
  min-height: 32px;
  margin-bottom: 12px;
}
.form-preview-toolbar :deep(.el-radio-button__inner) {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  letter-spacing: 0;
}
.form-preview-stage {
  height: min(68vh, 720px);
  overflow: auto;
  padding: 24px;
  border: 1px solid #dfe6ee;
  background: #f4f6f8;
}
.form-preview-sheet {
  width: min(760px, 100%);
  min-height: 100%;
  margin: 0 auto;
  padding: 32px;
  background: #fff;
  box-sizing: border-box;
}
.form-preview-sheet--mobile {
  width: 390px;
  max-width: 100%;
  padding: 24px 18px;
}
.form-preview-heading {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e5eaf0;
}
.form-preview-heading h2 {
  margin: 0;
  color: #1f2937;
  font-size: 20px;
  line-height: 1.4;
  letter-spacing: 0;
  overflow-wrap: anywhere;
}
.form-preview-sheet--mobile :deep(.runtime-form-grid),
.form-preview-sheet--mobile :deep(.detail-list-columns) {
  grid-template-columns: minmax(0, 1fr);
}
.form-preview-sheet--mobile :deep(.runtime-form-grid > .el-form-item),
.form-preview-sheet--mobile :deep(.runtime-form-group),
.form-preview-sheet--mobile :deep(.runtime-form-label),
.form-preview-sheet--mobile :deep(.runtime-form-description),
.form-preview-sheet--mobile :deep(.runtime-form-button),
.form-preview-sheet--mobile :deep(.detail-list-column) {
  grid-column: 1 / -1 !important;
}
@media (max-width: 640px) {
  .form-preview-stage { height: 66vh; padding: 12px; }
  .form-preview-sheet { padding: 22px 16px; }
}
</style>
