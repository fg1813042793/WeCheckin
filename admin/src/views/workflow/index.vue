<template>
  <div class="admin-page workflow-list-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar workflow-filters">
        <div class="admin-toolbar__left">
          <el-input v-model="filters.keyword" placeholder="搜索流程名称或编码" clearable style="width: 280px" @keyup.enter="search" />
          <el-input v-model="filters.category" placeholder="流程分类" clearable style="width: 180px" @keyup.enter="search" />
          <el-select v-model="filters.status" placeholder="全部状态" clearable style="width: 140px" @change="search">
            <el-option label="草稿" :value="1" />
            <el-option label="已发布" :value="2" />
            <el-option label="已停用" :value="0" />
          </el-select>
          <el-button type="primary" @click="search">搜索</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
      </div>

      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-button v-if="canAdd" type="primary" icon="Plus" @click="openCreate">创建流程</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" :loading="loading" @click="loadList" />
        </div>
      </div>

      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="name" label="流程名称" min-width="190">
          <template #default="{ row }">
            <div class="workflow-name">
              <span class="workflow-name__icon">
                <img v-if="row.logoUrl" :src="row.logoUrl" :alt="row.name" />
                <el-icon v-else><Share /></el-icon>
              </span>
              <div>
                <strong>{{ row.name }}</strong>
                <small>{{ row.key }}</small>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" min-width="120">
          <template #default="{ row }">{{ row.category || '-' }}</template>
        </el-table-column>
        <el-table-column prop="description" label="说明" min-width="240" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusMeta(row.status).type" size="small">{{ statusMeta(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="当前版本" width="100" align="center">
          <template #default="{ row }">{{ row.currentVersion ? `v${row.currentVersion}` : '-' }}</template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">{{ formatTime(row.editTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="390" fixed="right">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-button v-if="canEdit" size="small" icon="EditPen" @click="openEdit(row)">修改</el-button>
              <el-button v-if="canEdit" size="small" type="primary" @click="openDesigner(row)">设计</el-button>
              <el-button size="small" @click="openVersions(row)">版本</el-button>
              <el-button v-if="canPublish" size="small" type="success" plain @click="publish(row)">发布</el-button>
              <el-button v-if="canDelete" size="small" type="danger" plain :disabled="row.currentVersion > 0" @click="remove(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="admin-pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total,sizes,prev,pager,next"
          @current-change="loadList"
          @size-change="handlePageSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="editDialog" title="修改流程信息" width="520px" destroy-on-close>
      <el-form label-width="86px" @submit.prevent>
        <el-form-item label="流程名称" required>
          <el-input
            v-model="editForm.name"
            maxlength="80"
            placeholder="请输入流程名称"
            @keyup.enter="updateDefinitionMetadata"
          />
        </el-form-item>
        <el-form-item label="流程编码">
          <el-input :model-value="editTarget?.key" disabled />
          <div class="form-help">流程编码和已发布版本保持不变。</div>
        </el-form-item>
        <el-form-item label="流程分类">
          <el-input v-model="editForm.category" maxlength="50" placeholder="例如：行政审批" />
        </el-form-item>
        <el-form-item label="流程说明">
          <el-input v-model="editForm.description" type="textarea" :rows="3" maxlength="300" show-word-limit />
        </el-form-item>
        <el-form-item label="流程 Logo">
          <WorkflowLogoPicker
            :image-url="editLogoRemoved ? '' : editTarget?.logoUrl"
            :file="editLogoFile"
            :disabled="editing"
            @update:file="handleEditLogoFile"
            @clear="editLogoRemoved = true"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialog = false">取消</el-button>
        <el-button type="primary" :loading="editing" @click="updateDefinitionMetadata">保存</el-button>
      </template>
    </el-dialog>

    <WorkflowPublishDialog
      v-model="publishDialog"
      :definition="publishTarget"
      @published="loadList"
    />

    <el-dialog v-model="createDialog" title="创建流程" width="520px" destroy-on-close>
      <el-form label-width="86px" @submit.prevent>
        <el-form-item label="流程名称" required>
          <el-input v-model="createForm.name" maxlength="80" placeholder="例如：采购申请审批" />
        </el-form-item>
        <el-form-item label="流程编码" required>
          <el-input v-model="createForm.key" maxlength="80" placeholder="例如：purchase_approval" />
          <div class="form-help">发布后编码用于生成 BPMN process id，不建议再调整。</div>
        </el-form-item>
        <el-form-item label="流程分类">
          <el-input v-model="createForm.category" maxlength="50" placeholder="例如：行政审批" />
        </el-form-item>
        <el-form-item label="流程说明">
          <el-input v-model="createForm.description" type="textarea" :rows="3" maxlength="300" show-word-limit />
        </el-form-item>
        <el-form-item label="流程 Logo">
          <WorkflowLogoPicker v-model:file="createLogoFile" :disabled="creating" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialog = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="createDefinition">创建并设计</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="versionDrawer" title="发布版本" size="520px">
      <div v-loading="versionsLoading" class="version-list">
        <el-empty v-if="!versionsLoading && versions.length === 0" description="暂未发布版本" />
        <div v-for="version in versions" :key="version.id" class="version-item">
          <span class="version-item__badge">v{{ version.version }}</span>
          <div>
            <strong>{{ activeDefinition?.name }}</strong>
            <p>{{ formatTime(version.publishedAt) }} · 发布人 ID {{ version.publishedBy || '-' }}</p>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'
import { hasPerm } from '../../utils/permission'
import WorkflowLogoPicker from './components/WorkflowLogoPicker.vue'
import WorkflowPublishDialog from './components/WorkflowPublishDialog.vue'
import type { WorkflowDefinitionDetail, WorkflowDefinitionSummary, WorkflowVersion } from './types'

const router = useRouter()
const loading = ref(false)
const list = ref<WorkflowDefinitionSummary[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive<{ keyword: string; category: string; status: number | '' }>({ keyword: '', category: '', status: '' })

const canAdd = computed(() => hasPerm('admin:menu:workflow:add'))
const canEdit = computed(() => hasPerm('admin:menu:workflow:edit'))
const canPublish = computed(() => hasPerm('admin:menu:workflow:publish'))
const canDelete = computed(() => hasPerm('admin:menu:workflow:del'))

const createDialog = ref(false)
const creating = ref(false)
const createForm = reactive({ name: '', key: '', category: '', description: '' })
const createLogoFile = ref<File | null>(null)
const editDialog = ref(false)
const editing = ref(false)
const editTarget = ref<WorkflowDefinitionSummary | null>(null)
const editForm = reactive({ name: '', category: '', description: '' })
const editLogoFile = ref<File | null>(null)
const editLogoRemoved = ref(false)
const publishDialog = ref(false)
const publishTarget = ref<WorkflowDefinitionSummary | null>(null)
const versionDrawer = ref(false)
const versionsLoading = ref(false)
const versions = ref<WorkflowVersion[]>([])
const activeDefinition = ref<WorkflowDefinitionSummary | null>(null)

function statusMeta(status: number) {
  if (status === 2) return { label: '已发布', type: 'success' as const }
  if (status === 0) return { label: '已停用', type: 'info' as const }
  return { label: '草稿', type: 'warning' as const }
}

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

async function loadList() {
  loading.value = true
  try {
    const response = await adminApi.workflowDefinitionList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: filters.keyword.trim(),
      category: filters.category.trim(),
      status: filters.status,
    })
    list.value = Array.isArray(response.data?.list) ? response.data.list : []
    total.value = Number(response.data?.total || 0)
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  loadList()
}

function resetFilters() {
  filters.keyword = ''
  filters.category = ''
  filters.status = ''
  search()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  page.value = 1
  loadList()
}

function openCreate() {
  Object.assign(createForm, { name: '', key: '', category: '', description: '' })
  createLogoFile.value = null
  createDialog.value = true
}

function openEdit(row: WorkflowDefinitionSummary) {
  editTarget.value = row
  Object.assign(editForm, {
    name: row.name,
    category: row.category || '',
    description: row.description || '',
  })
  editLogoFile.value = null
  editLogoRemoved.value = false
  editDialog.value = true
}

function handleEditLogoFile(file: File | null) {
  editLogoFile.value = file
  if (file) editLogoRemoved.value = false
}

function buildWorkflowDefinitionFormData(values: Record<string, string>, logoFile: File | null, removeLogo = false) {
  const formData = new FormData()
  Object.entries(values).forEach(([key, value]) => formData.append(key, value))
  if (logoFile) formData.append('logo', logoFile)
  if (removeLogo) formData.append('removeLogo', 'true')
  return formData
}

async function updateDefinitionMetadata() {
  if (!editTarget.value) return
  const name = editForm.name.trim()
  const category = editForm.category.trim()
  const description = editForm.description.trim()
  if (!name) {
    ElMessage.warning('流程名称不能为空')
    return
  }
  if (
    name === editTarget.value.name
    && category === editTarget.value.category
    && description === editTarget.value.description
    && !editLogoFile.value
    && !editLogoRemoved.value
  ) {
    editDialog.value = false
    return
  }
  editing.value = true
  try {
    const formData = buildWorkflowDefinitionFormData({
      name,
      description: editForm.description.trim(),
      category: editForm.category.trim(),
    }, editLogoFile.value, editLogoRemoved.value)
    await adminApi.workflowDefinitionUpdate(editTarget.value.id, formData)
    editDialog.value = false
    ElMessage.success('流程信息修改成功')
    await loadList()
  } finally {
    editing.value = false
  }
}

async function createDefinition() {
  const name = createForm.name.trim()
  const key = createForm.key.trim()
  if (!name || !key) {
    ElMessage.warning('请填写流程名称和流程编码')
    return
  }
  if (!/^[A-Za-z][A-Za-z0-9_\-]*$/.test(key)) {
    ElMessage.warning('流程编码需以字母开头，仅支持字母、数字、下划线和短横线')
    return
  }
  creating.value = true
  try {
    const formData = buildWorkflowDefinitionFormData({
      name,
      key,
      category: createForm.category.trim(),
      description: createForm.description.trim(),
    }, createLogoFile.value)
    const response = await adminApi.workflowDefinitionCreate(formData)
    const detail = response.data as WorkflowDefinitionDetail
    createDialog.value = false
    ElMessage.success('流程已创建')
    await router.push(`/workflow/definitions/${detail.id}/designer`)
  } finally {
    creating.value = false
  }
}

function openDesigner(row: WorkflowDefinitionSummary) {
  router.push(`/workflow/definitions/${row.id}/designer`)
}

function publish(row: WorkflowDefinitionSummary) {
  publishTarget.value = row
  publishDialog.value = true
}

async function remove(row: WorkflowDefinitionSummary) {
  await ElMessageBox.confirm(`删除草稿“${row.name}”？该操作不可恢复。`, '删除流程', { type: 'warning' })
  await adminApi.workflowDefinitionDelete(row.id)
  ElMessage.success('删除成功')
  await loadList()
}

async function openVersions(row: WorkflowDefinitionSummary) {
  activeDefinition.value = row
  versions.value = []
  versionDrawer.value = true
  versionsLoading.value = true
  try {
    const response = await adminApi.workflowDefinitionVersions(row.id)
    versions.value = Array.isArray(response.data) ? response.data : []
  } finally {
    versionsLoading.value = false
  }
}

onMounted(loadList)
</script>

<style scoped>
.workflow-name { display: flex; align-items: center; gap: 10px; }
.workflow-name__icon { display: grid; place-items: center; width: 32px; height: 32px; overflow: hidden; border-radius: 7px; color: #0f766e; background: #e9f8f5; }
.workflow-name__icon img { width: 100%; height: 100%; object-fit: cover; }
.workflow-name strong { display: block; color: #1f2937; font-weight: 600; }
.workflow-name small { display: block; margin-top: 3px; color: #94a3b8; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.form-help { margin-top: 6px; color: #94a3b8; font-size: 12px; line-height: 1.5; }
.version-list { min-height: 180px; }
.version-item { display: flex; align-items: center; gap: 12px; padding: 14px 0; border-bottom: 1px solid #edf0f5; }
.version-item__badge { display: grid; place-items: center; width: 42px; height: 32px; border-radius: 6px; color: #0f766e; background: #e9f8f5; font-weight: 700; }
.version-item p { margin: 4px 0 0; color: #8492a6; font-size: 12px; }
</style>
