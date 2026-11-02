<template>
  <el-dialog v-model="visible" title="升级流程实例版本" width="min(620px, 92vw)" append-to-body destroy-on-close :close-on-click-modal="false">
    <div v-loading="loading" class="instance-version-upgrade">
      <template v-if="preview">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="实例 ID" :span="2"><code>{{ preview.instanceId }}</code></el-descriptions-item>
          <el-descriptions-item label="当前版本">v{{ preview.currentVersion }}</el-descriptions-item>
          <el-descriptions-item label="目标版本">v{{ preview.targetVersion }}</el-descriptions-item>
        </el-descriptions>

        <el-alert
          v-if="preview.compatible"
          class="upgrade-alert"
          type="success"
          title="兼容性检查通过"
          description="升级只切换实例绑定版本，不修改现有 Token、待办、表单数据和审批历史。"
          show-icon
          :closable="false"
        />
        <el-alert
          v-else-if="preview.alreadyLatest"
          class="upgrade-alert"
          type="info"
          title="当前实例已经使用最新发布版本"
          show-icon
          :closable="false"
        />
        <el-alert
          v-else
          class="upgrade-alert"
          type="error"
          title="当前实例不能原地升级"
          description="需要先处理以下兼容性问题，系统不会修改实例运行数据。"
          show-icon
          :closable="false"
        />

        <ul v-if="preview.blockers.length" class="upgrade-blockers">
          <li v-for="blocker in preview.blockers" :key="`${blocker.code}:${blocker.message}`">{{ blocker.message }}</li>
        </ul>
      </template>
    </div>

    <template #footer>
      <el-button :disabled="upgrading" @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="upgrading" :disabled="!preview?.compatible" @click="upgrade">确认升级</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api'
import type {
  WorkflowInstanceSummary,
  WorkflowInstanceVersionUpgradePreview,
  WorkflowInstanceVersionUpgradeResult,
} from '@/types/workflow'

const props = defineProps<{
  modelValue: boolean
  instance: WorkflowInstanceSummary | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'upgraded': [result: WorkflowInstanceVersionUpgradeResult]
}>()

const visible = computed({ get: () => props.modelValue, set: value => emit('update:modelValue', value) })
const loading = ref(false)
const upgrading = ref(false)
const preview = ref<WorkflowInstanceVersionUpgradePreview | null>(null)

watch(
  () => [props.modelValue, props.instance?.id] as const,
  ([open]) => {
    if (open) void loadPreview()
    else preview.value = null
  },
  { immediate: true },
)

async function loadPreview() {
  if (!props.instance?.id) return
  loading.value = true
  preview.value = null
  try {
    const response = await adminApi.workflowInstanceVersionUpgradePreview(props.instance.id)
    preview.value = response.data as WorkflowInstanceVersionUpgradePreview
  } finally {
    loading.value = false
  }
}

async function upgrade() {
  if (!props.instance || !preview.value?.compatible || upgrading.value) return
  await ElMessageBox.confirm(
    `确认将该实例从 v${preview.value.currentVersion} 升级至 v${preview.value.targetVersion}？`,
    '确认升级版本',
    { type: 'warning', confirmButtonText: '升级', cancelButtonText: '取消' },
  )
  upgrading.value = true
  try {
    const response = await adminApi.workflowInstanceVersionUpgrade(props.instance.id, {
      expectedFromVersion: preview.value.currentVersion,
    })
    const result = response.data as WorkflowInstanceVersionUpgradeResult
    ElMessage.success(`流程实例已升级至 v${result.currentVersion}`)
    emit('upgraded', result)
    visible.value = false
  } finally {
    upgrading.value = false
  }
}
</script>

<style scoped>
.instance-version-upgrade { min-height: 160px; }
.upgrade-alert { margin-top: 16px; }
.upgrade-blockers { margin: 14px 0 0; padding-left: 22px; color: #dc2626; line-height: 1.8; }
.instance-version-upgrade code { color: #475569; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
</style>
