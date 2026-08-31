<template>
  <div class="workflow-designer-page" v-loading="loading">
    <header class="designer-header">
      <div class="designer-header__main">
        <el-button circle icon="ArrowLeft" title="返回流程列表" @click="backToList" />
        <span class="designer-logo"><el-icon><Share /></el-icon></span>
        <div class="designer-title">
          <el-input
            v-if="detail"
            v-model="detail.name"
            class="designer-title__input"
            maxlength="80"
            @input="markDirty"
          />
          <div class="designer-title__meta">
            <code>{{ detail?.key || '-' }}</code>
            <el-tag v-if="detail" :type="statusMeta.type" size="small">{{ statusMeta.label }}</el-tag>
            <span v-if="detail?.currentVersion">当前 v{{ detail.currentVersion }}</span>
            <span v-if="dirty" class="dirty-state"><i />未保存</span>
          </div>
        </div>
      </div>
      <div class="designer-header__actions">
        <el-button @click="openMetadata">流程信息</el-button>
        <el-button @click="openVersions">版本</el-button>
        <el-button :loading="validating" @click="validateDraft">校验</el-button>
        <el-button v-if="canEdit" type="primary" plain :loading="saving" @click="saveDraft">保存</el-button>
        <el-button v-if="canPublish" type="success" :loading="publishing" @click="publishDraft">发布</el-button>
      </div>
    </header>

    <div v-if="detail" class="designer-body">
      <WorkflowCanvas
        :draft="detail.draft"
        :selected-node-id="selectedNodeId"
        :readonly="!canEdit"
        @select="selectNode"
        @insert="insertNode"
        @add-branch="addGatewayBranch"
      />
    </div>

    <div v-if="validationErrors.length" class="validation-panel">
      <div class="validation-panel__heading">
        <span><el-icon><WarningFilled /></el-icon>发现 {{ validationErrors.length }} 个问题</span>
        <el-button link @click="validationErrors = []">收起</el-button>
      </div>
      <button
        v-for="error in validationErrors"
        :key="`${error.code}-${error.nodeId || error.edgeId}`"
        type="button"
        @click="focusValidationError(error)"
      >
        <code>{{ error.code }}</code>
        <span>{{ error.message }}</span>
      </button>
    </div>

    <el-drawer v-model="nodeDrawer" :title="selectedNodeTitle" size="420px" append-to-body destroy-on-close>
      <NodeInspector
        v-if="detail"
        :draft="detail.draft"
        :selected-node-id="selectedNodeId"
        :readonly="!canEdit"
        @change="markDirty"
        @delete="deleteNode"
        @delete-branch="deleteGatewayBranch"
        @add-branch="addGatewayBranch"
      />
    </el-drawer>

    <el-drawer v-model="metadataDrawer" title="流程信息" size="440px" append-to-body>
      <el-form v-if="detail" label-position="top">
        <el-form-item label="流程名称" required>
          <el-input v-model="detail.name" maxlength="80" @input="markDirty" />
        </el-form-item>
        <el-form-item label="流程编码">
          <el-input :model-value="detail.key" disabled />
        </el-form-item>
        <el-form-item label="流程分类">
          <el-input v-model="detail.category" maxlength="50" placeholder="例如：行政审批" @input="markDirty" />
        </el-form-item>
        <el-form-item label="流程说明">
          <el-input v-model="detail.description" type="textarea" :rows="5" maxlength="300" show-word-limit @input="markDirty" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="metadataDrawer = false">关闭</el-button>
        <el-button v-if="canEdit" type="primary" :loading="saving" @click="saveMetadata">保存</el-button>
      </template>
    </el-drawer>

    <el-drawer v-model="versionDrawer" title="发布版本" size="520px" append-to-body>
      <div v-loading="versionsLoading" class="version-list">
        <el-empty v-if="!versionsLoading && versions.length === 0" description="暂未发布版本" />
        <div v-for="version in versions" :key="version.id" class="version-item">
          <span class="version-item__badge">v{{ version.version }}</span>
          <div>
            <strong>{{ detail?.name }}</strong>
            <p>{{ formatTime(version.publishedAt) }} · 发布人 ID {{ version.publishedBy || '-' }}</p>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../../api'
import { hasPerm } from '../../../utils/permission'
import WorkflowCanvas from './components/WorkflowCanvas.vue'
import NodeInspector from './components/NodeInspector.vue'
import { addBranch, insertNodeAtEdge, removeBranch, removeNode } from './graph'
import type {
  WorkflowDefinitionDetail,
  WorkflowNodeType,
  WorkflowValidationError,
  WorkflowValidationResult,
  WorkflowVersion,
} from '../types'
import { cloneDraft } from '../types'

const route = useRoute()
const router = useRouter()
const definitionId = computed(() => Number(route.params.id || 0))
const loading = ref(false)
const saving = ref(false)
const validating = ref(false)
const publishing = ref(false)
const detail = ref<WorkflowDefinitionDetail | null>(null)
const selectedNodeId = ref('')
const dirty = ref(false)
const validationErrors = ref<WorkflowValidationError[]>([])
const nodeDrawer = ref(false)
const metadataDrawer = ref(false)
const versionDrawer = ref(false)
const versionsLoading = ref(false)
const versions = ref<WorkflowVersion[]>([])

const canEdit = computed(() => hasPerm('admin:menu:workflow:edit'))
const canPublish = computed(() => hasPerm('admin:menu:workflow:publish'))
const selectedNode = computed(() => detail.value?.draft.nodes.find(item => item.id === selectedNodeId.value))
const selectedNodeTitle = computed(() => {
  if (!selectedNode.value) return '节点配置'
  if (selectedNode.value.type === 'start') return '发起人配置'
  if (selectedNode.value.type === 'approval') return '审批人配置'
  if (selectedNode.value.gatewayMode === 'split') return selectedNode.value.type === 'exclusive' ? '条件分支配置' : '并行分支配置'
  return '节点配置'
})
const statusMeta = computed(() => {
  if (detail.value?.status === 2) return { label: '已发布', type: 'success' as const }
  if (detail.value?.status === 0) return { label: '已停用', type: 'info' as const }
  return { label: '草稿', type: 'warning' as const }
})

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

async function loadDetail() {
  if (!definitionId.value) {
    ElMessage.error('流程定义 ID 无效')
    await router.replace('/workflow/definitions')
    return
  }
  loading.value = true
  try {
    const response = await adminApi.workflowDefinitionDetail(definitionId.value)
    const data = response.data as WorkflowDefinitionDetail
    data.draft = cloneDraft(data.draft)
    detail.value = data
    selectedNodeId.value = ''
    nodeDrawer.value = false
    dirty.value = false
    validationErrors.value = []
  } finally {
    loading.value = false
  }
}

function markDirty() {
  dirty.value = true
  validationErrors.value = []
}

function selectNode(nodeId: string) {
  selectedNodeId.value = nodeId
  nodeDrawer.value = Boolean(nodeId)
}

function insertNode(payload: {
  edgeId: string
  type: Extract<WorkflowNodeType, 'approval' | 'exclusive' | 'parallel'>
}) {
  if (!detail.value) return
  const node = insertNodeAtEdge(detail.value.draft, payload.edgeId, payload.type)
  if (!node) {
    ElMessage.warning('该流程位置暂时无法插入节点')
    return
  }
  selectNode(node.id)
  markDirty()
}

async function deleteNode(nodeId: string) {
  if (!detail.value) return
  const node = detail.value.draft.nodes.find(item => item.id === nodeId)
  if (!node) return
  try {
    await ElMessageBox.confirm(`确定删除“${node.name}”？关联的连线将同步调整。`, '删除节点', { type: 'warning' })
  } catch {
    return
  }
  if (!removeNode(detail.value.draft, nodeId)) {
    ElMessage.warning('该节点当前无法删除；分支至少需要保留两条路径')
    return
  }
  selectNode('')
  markDirty()
}

async function deleteGatewayBranch(splitId: string, edgeId: string) {
  if (!detail.value) return
  const branch = detail.value.draft.edges.find(item => item.id === edgeId && item.source === splitId)
  if (!branch) return
  try {
    await ElMessageBox.confirm(`确定删除“${branch.name || '该分支'}”？分支内的节点将一并删除。`, '删除分支', { type: 'warning' })
  } catch {
    return
  }
  if (!removeBranch(detail.value.draft, splitId, edgeId)) {
    ElMessage.warning('分支至少需要保留两条路径')
    return
  }
  markDirty()
}

function addGatewayBranch(splitId: string) {
  if (!detail.value) return
  const node = addBranch(detail.value.draft, splitId)
  if (!node) {
    ElMessage.warning('未找到对应的汇聚节点')
    return
  }
  selectNode(splitId)
  markDirty()
}

async function saveDraft(showMessage = true) {
  if (!detail.value) return false
  const name = detail.value.name.trim()
  if (!name) {
    ElMessage.warning('流程名称不能为空')
    return false
  }
  saving.value = true
  try {
    detail.value.draft.key = detail.value.key
    detail.value.draft.name = name
    const response = await adminApi.workflowDefinitionUpdate(detail.value.id, {
      name,
      category: detail.value.category.trim(),
      description: detail.value.description.trim(),
      status: detail.value.status,
      draft: detail.value.draft,
    })
    const saved = response.data as WorkflowDefinitionDetail
    saved.draft = cloneDraft(saved.draft)
    detail.value = saved
    dirty.value = false
    if (showMessage) ElMessage.success('草稿已保存')
    return true
  } finally {
    saving.value = false
  }
}

async function validateDraft(showSuccess = true) {
  if (!detail.value) return false
  validating.value = true
  try {
    if (!(await saveDraft(false))) return false
    const response = await adminApi.workflowDefinitionValidate(detail.value.id)
    const result = response.data as WorkflowValidationResult
    validationErrors.value = Array.isArray(result.errors) ? result.errors : []
    if (result.valid) {
      if (showSuccess) ElMessage.success('流程结构校验通过')
      return true
    }
    ElMessage.error(`流程存在 ${validationErrors.value.length} 个问题`)
    focusValidationError(validationErrors.value[0])
    return false
  } finally {
    validating.value = false
  }
}

async function publishDraft() {
  if (!detail.value) return
  if (!(await validateDraft(false))) return
  try {
    await ElMessageBox.confirm('发布后将生成不可变更的新版本，后续修改会进入下一版本草稿。确定继续？', '发布流程', { type: 'warning' })
  } catch {
    return
  }
  publishing.value = true
  try {
    const response = await adminApi.workflowDefinitionPublish(detail.value.id)
    ElMessage.success(`流程已发布为 v${response.data?.version}`)
    await loadDetail()
  } finally {
    publishing.value = false
  }
}

function focusValidationError(error?: WorkflowValidationError) {
  if (!error) return
  if (error.nodeId) selectNode(error.nodeId)
  else if (error.edgeId && detail.value) {
    const target = detail.value.draft.edges.find(item => item.id === error.edgeId)?.source
    if (target) selectNode(target)
  }
}

function openMetadata() {
  metadataDrawer.value = true
}

async function saveMetadata() {
  if (await saveDraft()) metadataDrawer.value = false
}

async function openVersions() {
  if (!detail.value) return
  versions.value = []
  versionDrawer.value = true
  versionsLoading.value = true
  try {
    const response = await adminApi.workflowDefinitionVersions(detail.value.id)
    versions.value = Array.isArray(response.data) ? response.data : []
  } finally {
    versionsLoading.value = false
  }
}

function backToList() {
  router.push('/workflow/definitions')
}

onBeforeRouteLeave(() => {
  if (!dirty.value) return true
  return window.confirm('当前流程还有未保存的修改，确定离开吗？')
})

onMounted(loadDetail)
</script>

<style scoped>
.workflow-designer-page { position: relative; display: flex; flex-direction: column; min-height: calc(100vh - 138px); margin: -20px; background: #f5f7fa; }
.designer-header { display: flex; align-items: center; justify-content: space-between; gap: 20px; min-height: 72px; padding: 0 20px; border-bottom: 1px solid #dfe6ee; background: #fff; }
.designer-header__main, .designer-header__actions { display: flex; align-items: center; gap: 10px; }
.designer-header__main { min-width: 0; }
.designer-header__actions { flex: 0 0 auto; }
.designer-logo { display: grid; place-items: center; width: 38px; height: 38px; border-radius: 8px; color: #fff; background: #0f9f8f; font-size: 19px; }
.designer-title { min-width: 260px; }
.designer-title__input { width: min(420px, 32vw); }
.designer-title__input :deep(.el-input__wrapper) { padding-left: 0; box-shadow: none; }
.designer-title__input :deep(.el-input__inner) { height: 28px; color: #1f2937; font-size: 18px; font-weight: 650; }
.designer-title__meta { display: flex; align-items: center; gap: 8px; min-height: 22px; color: #94a3b8; font-size: 11px; }
.designer-title__meta code { color: #64748b; }
.dirty-state { display: inline-flex; align-items: center; gap: 5px; color: #d97706; }
.dirty-state i { width: 6px; height: 6px; border-radius: 50%; background: #f59e0b; }
.designer-body { min-height: 0; flex: 1; display: flex; height: calc(100vh - 210px); }
.validation-panel { position: absolute; left: 18px; bottom: 18px; z-index: 5; width: min(560px, calc(100vw - 36px)); max-height: 230px; overflow-y: auto; border: 1px solid #f4c7c3; border-radius: 8px; background: #fff; box-shadow: 0 14px 35px rgb(15 23 42 / 14%); }
.validation-panel__heading { position: sticky; top: 0; display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; border-bottom: 1px solid #f7dfdc; color: #b42318; background: #fff8f7; font-size: 13px; font-weight: 600; }
.validation-panel__heading span { display: flex; align-items: center; gap: 6px; }
.validation-panel > button { display: flex; width: 100%; gap: 10px; padding: 10px 12px; border: 0; border-bottom: 1px solid #f1f3f6; background: #fff; color: #475569; text-align: left; cursor: pointer; }
.validation-panel > button:hover { background: #fafbfc; }
.validation-panel > button code { color: #b42318; font-size: 11px; }
.version-item { display: flex; align-items: center; gap: 12px; padding: 14px 0; border-bottom: 1px solid #edf0f5; }
.version-item__badge { display: grid; place-items: center; width: 42px; height: 32px; border-radius: 6px; color: #0f766e; background: #e9f8f5; font-weight: 700; }
.version-item p { margin: 4px 0 0; color: #8492a6; font-size: 12px; }
@media (max-width: 1180px) {
  .designer-header { align-items: flex-start; flex-direction: column; padding: 12px 16px; }
  .designer-header__actions { width: 100%; overflow-x: auto; padding-bottom: 2px; }
  .designer-title__input { width: 360px; }
  .designer-body { height: calc(100vh - 270px); }
}
</style>
