<template>
  <el-drawer
    :model-value="modelValue"
    title="发布版本"
    size="640px"
    append-to-body
    destroy-on-close
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div v-loading="loading" class="version-list">
      <el-empty v-if="!loading && versions.length === 0" description="暂未发布版本" />
      <article v-for="version in versions" :key="version.id" class="version-item">
        <span class="version-item__badge">v{{ version.version }}</span>
        <div class="version-item__content">
          <div class="version-item__heading">
            <strong>{{ version.name || definition?.name || '-' }}</strong>
            <el-tag v-if="version.isCurrent" size="small" type="success">当前版本</el-tag>
            <el-tag v-if="version.rollbackFromVersion" size="small" type="warning">
              回滚自 v{{ version.rollbackFromVersion }}
            </el-tag>
          </div>
          <p class="version-item__meta">
            {{ formatTime(version.publishedAt) }} · {{ version.publishedByName || `发布人 ID ${version.publishedBy || '-'}` }}
          </p>
          <p class="version-item__summary">{{ version.changeHeadline || '暂无变更摘要' }}</p>
          <p v-if="version.publishNote" class="version-item__note">{{ version.publishNote }}</p>
          <p v-if="version.instanceCount || version.startDraftCount" class="version-item__references">
            <span v-if="version.instanceCount">{{ version.instanceCount }} 个流程实例</span>
            <span v-if="version.startDraftCount">{{ version.startDraftCount }} 份发起草稿</span>
          </p>
        </div>
        <div class="version-item__actions">
          <el-button link type="primary" :disabled="busyVersion === version.version" @click="openChanges(version)">
            查看变更
          </el-button>
          <el-button
            v-if="canPublish"
            link
            type="warning"
            :disabled="version.isCurrent || busyVersion === version.version"
            @click="rollbackVersion(version)"
          >
            回滚
          </el-button>
          <el-tooltip
            v-if="canDelete"
            :content="version.deleteBlockedReason || '删除该历史版本'"
            placement="top"
          >
            <span>
              <el-button
                link
                type="danger"
                icon="Delete"
                title="删除版本"
                :disabled="!version.canDelete || busyVersion === version.version"
                @click="deleteVersion(version)"
              />
            </span>
          </el-tooltip>
        </div>
      </article>
    </div>
  </el-drawer>

  <el-dialog
    v-model="changeDialog"
    :title="changeTarget ? `v${changeTarget.version} 版本变更` : '版本变更'"
    width="620px"
    append-to-body
    destroy-on-close
  >
    <div class="change-toolbar">
      <span>对比版本</span>
      <el-select v-model="compareTo" :disabled="changesLoading" @change="loadChanges">
        <el-option :value="0" label="发布时基准版本" />
        <el-option
          v-for="version in comparableVersions"
          :key="version.version"
          :value="version.version"
          :label="`v${version.version}`"
        />
      </el-select>
    </div>
    <div v-loading="changesLoading" class="change-content">
      <el-alert
        v-if="changeSummary"
        :title="changeSummary.headline"
        :description="changeSummary.baseVersion ? `相对于 v${changeSummary.baseVersion}，共 ${changeSummary.changeCount} 项变更` : '流程首次发布'"
        type="info"
        :closable="false"
        show-icon
      />
      <el-empty v-if="!changesLoading && !changeSummary" description="暂无变更数据" />
      <section v-for="group in groupedChanges" :key="group.category" class="change-group">
        <h4>{{ group.label }} <span>{{ group.items.length }}</span></h4>
        <div v-for="(item, index) in group.items" :key="`${item.title}-${index}`" class="change-item">
          <el-tag :type="changeActionMeta(item.action).type" size="small">
            {{ changeActionMeta(item.action).label }}
          </el-tag>
          <div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.detail }}</p>
          </div>
        </div>
      </section>
    </div>
    <template #footer>
      <el-button @click="changeDialog = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../../api'
import type {
  WorkflowDefinitionSummary,
  WorkflowVersion,
  WorkflowVersionChangeAction,
  WorkflowVersionChangeCategory,
  WorkflowVersionChangeSummary,
} from '../types'

const props = defineProps<{
  modelValue: boolean
  definition: WorkflowDefinitionSummary | null
  canPublish: boolean
  canDelete: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  changed: [version: number]
}>()

const loading = ref(false)
const versions = ref<WorkflowVersion[]>([])
const busyVersion = ref(0)
const changeDialog = ref(false)
const changesLoading = ref(false)
const changeTarget = ref<WorkflowVersion | null>(null)
const changeSummary = ref<WorkflowVersionChangeSummary | null>(null)
const compareTo = ref(0)

const categoryLabels: Record<WorkflowVersionChangeCategory, string> = {
  basic: '基础信息',
  form: '表单',
  node: '流程节点',
  route: '流转路径',
  start: '发起配置',
  notification: '通知',
  automation: '自动动作',
}
const categoryOrder = Object.keys(categoryLabels) as WorkflowVersionChangeCategory[]

const comparableVersions = computed(() => versions.value.filter(item => item.version !== changeTarget.value?.version))
const groupedChanges = computed(() => {
  const items = changeSummary.value?.items || []
  return categoryOrder
    .map(category => ({ category, label: categoryLabels[category], items: items.filter(item => item.category === category) }))
    .filter(group => group.items.length > 0)
})

watch(
  () => [props.modelValue, props.definition?.id] as const,
  ([visible]) => {
    if (visible) loadVersions()
  },
  { immediate: true },
)

async function loadVersions() {
  if (!props.definition?.id) {
    versions.value = []
    return
  }
  loading.value = true
  try {
    const response = await adminApi.workflowDefinitionVersions(props.definition.id)
    versions.value = Array.isArray(response.data) ? response.data : []
  } finally {
    loading.value = false
  }
}

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function changeActionMeta(action: WorkflowVersionChangeAction) {
  if (action === 'add') return { label: '新增', type: 'success' as const }
  if (action === 'delete') return { label: '删除', type: 'danger' as const }
  if (action === 'reorder') return { label: '排序', type: 'warning' as const }
  return { label: '修改', type: 'primary' as const }
}

async function openChanges(version: WorkflowVersion) {
  changeTarget.value = version
  compareTo.value = 0
  changeSummary.value = null
  changeDialog.value = true
  await loadChanges()
}

async function loadChanges() {
  if (!props.definition?.id || !changeTarget.value) return
  changesLoading.value = true
  try {
    const response = await adminApi.workflowDefinitionVersionChanges(
      props.definition.id,
      changeTarget.value.version,
      compareTo.value || undefined,
    )
    changeSummary.value = response.data || null
  } finally {
    changesLoading.value = false
  }
}

async function rollbackVersion(version: WorkflowVersion) {
  if (!props.definition?.id || version.isCurrent) return
  const { value } = await ElMessageBox.prompt(
    `系统会将 v${version.version} 的内容复制并发布为新版本，同时覆盖当前设计草稿；现有流程实例不受影响。`,
    '回滚流程',
    {
      type: 'warning',
      confirmButtonText: '确认回滚',
      cancelButtonText: '取消',
      inputValue: `回滚至 v${version.version}`,
      inputPlaceholder: '可选，填写回滚说明',
      inputValidator: input => !input || [...input].length <= 500 || '回滚说明不能超过 500 个字符',
    },
  )
  busyVersion.value = version.version
  try {
    const response = await adminApi.workflowDefinitionVersionRollback(props.definition.id, version.version, { note: value?.trim() })
    const newVersion = Number(response.data?.version || 0)
    ElMessage.success(`已回滚并发布为 v${newVersion}`)
    await loadVersions()
    emit('changed', newVersion)
  } finally {
    busyVersion.value = 0
  }
}

async function deleteVersion(version: WorkflowVersion) {
  if (!props.definition?.id || !version.canDelete) return
  await ElMessageBox.confirm(
    `确定物理删除 v${version.version}？该操作不可恢复，版本号不会重新使用。`,
    '删除版本',
    { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
  )
  busyVersion.value = version.version
  try {
    await adminApi.workflowDefinitionVersionDelete(props.definition.id, version.version)
    ElMessage.success(`v${version.version} 已删除`)
    await loadVersions()
    emit('changed', version.version)
  } finally {
    busyVersion.value = 0
  }
}
</script>

<style scoped>
.version-list { min-height: 180px; }
.version-item { display: grid; grid-template-columns: 48px minmax(0, 1fr) auto; gap: 12px; padding: 16px 0; border-bottom: 1px solid #e5e7eb; }
.version-item__badge { display: grid; place-items: center; width: 44px; height: 32px; border-radius: 6px; color: #0f766e; background: #e9f8f5; font-weight: 700; }
.version-item__content { min-width: 0; }
.version-item__heading { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; color: #1f2937; }
.version-item__meta,
.version-item__summary,
.version-item__note,
.version-item__references { margin: 5px 0 0; font-size: 12px; line-height: 1.5; }
.version-item__meta { color: #8492a6; }
.version-item__summary { color: #475569; }
.version-item__note { color: #64748b; overflow-wrap: anywhere; }
.version-item__references { display: flex; flex-wrap: wrap; gap: 10px; color: #b45309; }
.version-item__actions { display: flex; align-items: flex-start; gap: 2px; white-space: nowrap; }
.change-toolbar { display: flex; align-items: center; justify-content: flex-end; gap: 10px; margin-bottom: 14px; color: #64748b; font-size: 13px; }
.change-toolbar .el-select { width: 190px; }
.change-content { min-height: 180px; }
.change-group { margin-top: 18px; }
.change-group h4 { margin: 0 0 8px; color: #1f2937; font-size: 14px; }
.change-group h4 span { margin-left: 4px; color: #94a3b8; font-weight: 400; }
.change-item { display: grid; grid-template-columns: 52px minmax(0, 1fr); gap: 10px; padding: 10px 0; border-bottom: 1px solid #edf0f5; }
.change-item .el-tag { justify-self: start; }
.change-item strong { color: #334155; font-size: 13px; }
.change-item p { margin: 4px 0 0; color: #64748b; font-size: 12px; line-height: 1.55; overflow-wrap: anywhere; }
@media (max-width: 720px) {
  .version-item { grid-template-columns: 44px minmax(0, 1fr); }
  .version-item__actions { grid-column: 2; flex-wrap: wrap; }
}
</style>
