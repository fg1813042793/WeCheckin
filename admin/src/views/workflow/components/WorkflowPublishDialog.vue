<template>
  <el-dialog
    :model-value="modelValue"
    title="发布流程"
    width="500px"
    destroy-on-close
    :close-on-click-modal="false"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <el-form label-width="110px" @submit.prevent>
      <el-form-item label="流程名称">
        <el-input :model-value="publishTarget?.name" disabled />
      </el-form-item>
      <el-form-item label="发布版本">
        <el-input :model-value="publishTarget ? `v${publishTarget.currentVersion + 1}` : ''" disabled />
      </el-form-item>
      <el-form-item label="配置来源">
        <div class="publish-config-source">
          <el-icon><Setting /></el-icon>
          <span>流程配置将随本次版本一起发布</span>
        </div>
      </el-form-item>
      <el-form-item label="发布说明">
        <el-input
          v-model="publishNote"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          placeholder="可选，说明本次流程调整内容"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button :disabled="publishing" @click="$emit('update:modelValue', false)">取消</el-button>
      <el-button type="success" :loading="publishing" @click="confirmPublish">发布</el-button>
    </template>
  </el-dialog>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'
import type {
  WorkflowDefinitionSummary,
} from '../types'

const props = defineProps<{
  modelValue: boolean
  definition: WorkflowDefinitionSummary | null
}>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  published: [version: number]
}>()

const publishing = ref(false)
const publishNote = ref('')
const publishTarget = computed(() => props.definition)

watch(() => props.modelValue, (visible) => {
  if (visible) publishNote.value = ''
})

async function confirmPublish() {
  if (!publishTarget.value) return
  publishing.value = true
  try {
    const response = await adminApi.workflowDefinitionPublish(publishTarget.value.id, { note: publishNote.value.trim() })
    const version = Number(response.data?.version || 0)
    ElMessage.success(`已发布 v${version}`)
    emit('update:modelValue', false)
    emit('published', version)
  } finally {
    publishing.value = false
  }
}
</script>

<style scoped>
.publish-config-source { display: flex; align-items: center; gap: 8px; color: #475569; }
.publish-config-source .el-icon { color: #0f766e; }
</style>
