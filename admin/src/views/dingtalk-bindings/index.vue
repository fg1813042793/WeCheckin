<template>
  <div class="admin-page dingtalk-bindings-page" aria-label="钉钉用户绑定管理">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-select v-model="filters.corpId" clearable placeholder="全部企业" style="width:220px" @change="handleFilter">
            <el-option
              v-for="corp in corpOptions"
              :key="corp.corpId"
              :label="corpLabel(corp)"
              :value="corp.corpId"
            />
          </el-select>
          <el-input
            v-model="filters.keyword"
            clearable
            placeholder="搜索钉钉 UserId / UnionId / 本地用户"
            style="width:380px"
            @keyup.enter="handleFilter"
            @clear="handleFilter"
          />
          <el-select v-model="filters.enabled" clearable placeholder="全部状态" style="width:150px" @change="handleFilter">
            <el-option label="启用" value="1" />
            <el-option label="停用" value="0" />
          </el-select>
          <el-button type="primary" @click="handleFilter">搜索</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
      </div>

      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-tooltip :disabled="canCreateBinding" :content="createDisabledTip" placement="top">
            <span>
              <el-button type="success" :disabled="!canCreateBinding" @click="openCreate">+ 新增绑定</el-button>
            </span>
          </el-tooltip>
          <el-button plain @click="fieldHelpVisible = true">说明文件</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" @click="loadBindings" />
        </div>
      </div>

      <el-table :data="list" v-loading="loading" stripe style="width:100%">
        <el-table-column label="企业" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="binding-cell-main">{{ row.corpName || row.corpId }}</div>
            <div class="binding-cell-sub">{{ row.corpId }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="dingTalkUserId" label="钉钉 UserId" min-width="180" show-overflow-tooltip />
        <el-table-column prop="unionId" label="UnionId" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.unionId || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="本地用户" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="binding-cell-main">{{ row.userName || `用户 ${row.userId}` }}</div>
            <div class="binding-cell-sub">{{ row.userAccount || row.userMiniOpenId || row.userId }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ row.enabled === 1 ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.editTime || row.addTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-tooltip :disabled="canEdit" content="缺少钉钉用户绑定维护权限" placement="top">
                <span>
                  <el-button size="small" type="primary" :disabled="!canEdit" @click="openEdit(row)">编辑</el-button>
                </span>
              </el-tooltip>
              <el-tooltip :disabled="canEdit" content="缺少钉钉用户绑定维护权限" placement="top">
                <span>
                  <el-button
                    size="small"
                    :type="row.enabled === 1 ? 'warning' : 'success'"
                    :disabled="!canEdit"
                    @click="toggleStatus(row)"
                  >
                    {{ row.enabled === 1 ? '停用' : '启用' }}
                  </el-button>
                </span>
              </el-tooltip>
              <el-tooltip :disabled="canEdit" content="缺少钉钉用户绑定维护权限" placement="top">
                <span>
                  <el-button size="small" type="danger" plain :disabled="!canEdit" @click="deleteBinding(row)">
                    删除
                  </el-button>
                </span>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="admin-pagination">
        <el-pagination
          v-model:current-page="filters.page"
          v-model:page-size="filters.pageSize"
          background
          layout="total, sizes, prev, pager, next"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          @current-change="loadBindings"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑绑定' : '新增绑定'" width="560px">
      <el-form label-width="120px" class="binding-form">
        <el-form-item label="钉钉企业">
          <el-select v-model="form.corpId" filterable placeholder="请选择钉钉企业">
            <el-option
              v-for="corp in corpOptions"
              :key="corp.corpId"
              :label="corpLabel(corp)"
              :value="corp.corpId"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="钉钉 UserId">
          <el-input v-model="form.dingTalkUserId" maxlength="160" placeholder="请输入钉钉 UserId" />
        </el-form-item>
        <el-form-item label="UnionId">
          <el-input v-model="form.unionId" maxlength="160" placeholder="可选" />
        </el-form-item>
        <el-form-item label="本地用户">
          <el-tree-select
            v-model="form.userId"
            :data="userTreeOptions"
            :props="userTreeProps"
            node-key="value"
            filterable
            check-strictly
            clearable
            :render-after-expand="false"
            :filter-node-method="filterUserTreeNode"
            placeholder="请选择本地用户"
            @change="handleLocalUserChange"
          >
            <template #default="{ data }">
              <div :class="['user-tree-node', `user-tree-node--${data.type || 'dept'}`]">
                <span class="user-tree-node__label">{{ data.label }}</span>
                <small v-if="data.type === 'user'" class="user-tree-node__meta">
                  {{ data.account || data.miniOpenId || data.userId }}
                </small>
                <small v-else-if="data.count" class="user-tree-node__meta">{{ data.count }} 人</small>
              </div>
            </template>
          </el-tree-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch
            v-model="form.enabled"
            :active-value="1"
            :inactive-value="0"
            active-text="启用"
            inactive-text="停用"
            inline-prompt
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="!canEdit" @click="saveBinding">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="fieldHelpVisible" title="钉钉用户绑定字段说明" width="720px" class="binding-field-help-dialog">
      <div class="binding-field-help">
        <section class="binding-field-help__section">
          <h3>企业</h3>
          <p>对应“钉钉应用管理 / 配置选项”里的企业应用配置。绑定记录必须选择用户所属企业，同一员工在不同企业下的钉钉身份可能不同。</p>
        </section>
        <section class="binding-field-help__section">
          <h3>钉钉 UserId</h3>
          <p>员工在当前企业通讯录中的用户 UserId。可在钉钉管理后台通讯录成员详情、通讯录导出数据中查看，也可通过钉钉免登录返回的用户信息获取。</p>
        </section>
        <section class="binding-field-help__section">
          <h3>UnionId</h3>
          <p>钉钉开放平台的用户统一标识。可通过钉钉用户详情接口或免登录用户信息接口获取；如果暂时没有 UnionId，可以先使用钉钉 UserId 完成绑定。</p>
        </section>
        <section class="binding-field-help__section">
          <h3>本地用户</h3>
          <p>对应本系统“用户管理”里的账号。免登录时系统会根据所选企业下的钉钉 UserId / UnionId 找到绑定的本地用户，并以该用户身份进入系统。</p>
        </section>
        <div class="binding-field-help__note">
          保存前请确认钉钉 UserId、UnionId 与所选企业一致；跨企业或迁移企业后，需要为新的企业应用重新维护绑定关系。
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="fieldHelpVisible = false">知道了</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../../utils/request'
import { hasPerm } from '../../utils/permission'

type BindingRow = {
  id: number
  corpId: string
  corpName: string
  dingTalkUserId: string
  unionId: string
  userId: number
  userName: string
  userAccount: string
  userMiniOpenId: string
  enabled: number
  addTime: number
  editTime: number
}

type CorpOption = {
  corpId: string
  corpName: string
  enabled: number
}

type UserOption = {
  id: number
  name: string
  account: string
  miniOpenId: string
  status: number
}

type UserTreeNode = {
  value: number | string
  label: string
  type: 'dept' | 'user'
  userId?: number
  account?: string
  miniOpenId?: string
  status?: number
  disabled?: boolean
  count?: number
  searchText?: string
  children?: UserTreeNode[]
}

type BindingListResponse = {
  list: BindingRow[]
  total: number
  corpOptions: CorpOption[]
  userOptions: UserOption[]
  userTreeOptions: UserTreeNode[]
}

const editPermissionKeys = ['admin:menu:dingtalk:bindings:edit']
const canEdit = computed(() => editPermissionKeys.some(key => hasPerm(key)))
const canCreateBinding = computed(() => canEdit.value && corpOptions.value.length > 0)
const createDisabledTip = computed(() => {
  if (!canEdit.value) return '缺少钉钉用户绑定维护权限'
  if (corpOptions.value.length === 0) return '请先配置钉钉企业'
  return ''
})
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const fieldHelpVisible = ref(false)
const list = ref<BindingRow[]>([])
const total = ref(0)
const corpOptions = ref<CorpOption[]>([])
const userOptions = ref<UserOption[]>([])
const userTreeOptions = ref<UserTreeNode[]>([])
const userTreeProps = {
  label: 'label',
  value: 'value',
  children: 'children',
  disabled: 'disabled'
}

const filters = reactive({
  page: 1,
  pageSize: 20,
  corpId: '',
  keyword: '',
  enabled: ''
})

const form = reactive({
  id: 0,
  corpId: '',
  dingTalkUserId: '',
  unionId: '',
  userId: undefined as number | undefined,
  enabled: 1
})

function corpLabel(corp: CorpOption) {
  const name = corp.corpName || corp.corpId
  return corp.enabled === 1 ? name : `${name}（停用）`
}

function formatTime(value: number) {
  if (!value) return '-'
  const date = new Date(value)
  const pad = (num: number) => String(num).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

async function loadBindings() {
  loading.value = true
  try {
    const res = await request.get<BindingListResponse>('/api/v2/admin/dingtalk/user-bindings', {
      params: {
        page: filters.page,
        pageSize: filters.pageSize,
        corpId: filters.corpId,
        keyword: filters.keyword.trim(),
        enabled: filters.enabled
      }
    })
    const data = res.data || {}
    list.value = Array.isArray(data.list) ? data.list : []
    total.value = Number(data.total || 0)
    corpOptions.value = Array.isArray(data.corpOptions) ? data.corpOptions : []
    userOptions.value = Array.isArray(data.userOptions) ? data.userOptions : []
    userTreeOptions.value = Array.isArray(data.userTreeOptions) ? data.userTreeOptions : []
  } finally {
    loading.value = false
  }
}

function handleFilter() {
  filters.page = 1
  loadBindings()
}

function handleSizeChange() {
  filters.page = 1
  loadBindings()
}

function resetFilters() {
  filters.corpId = ''
  filters.keyword = ''
  filters.enabled = ''
  handleFilter()
}

function resetForm() {
  form.id = 0
  form.corpId = corpOptions.value[0]?.corpId || ''
  form.dingTalkUserId = ''
  form.unionId = ''
  form.userId = undefined
  form.enabled = 1
}

function openCreate() {
  resetForm()
  dialogVisible.value = true
}

function openEdit(row: BindingRow) {
  form.id = row.id
  form.corpId = row.corpId
  form.dingTalkUserId = row.dingTalkUserId
  form.unionId = row.unionId || ''
  form.userId = row.userId
  form.enabled = row.enabled === 1 ? 1 : 0
  dialogVisible.value = true
}

function filterUserTreeNode(keyword: string, data: UserTreeNode) {
  const text = keyword.trim().toLowerCase()
  if (!text) return true
  return String(data.searchText || data.label || '').toLowerCase().includes(text)
}

function handleLocalUserChange(value: number | string | undefined) {
  const id = Number(value)
  form.userId = Number.isFinite(id) && id > 0 ? id : undefined
}

async function saveBinding() {
  if (!form.corpId) {
    ElMessage.warning('请选择钉钉企业')
    return
  }
  if (!form.dingTalkUserId.trim()) {
    ElMessage.warning('请输入钉钉 UserId')
    return
  }
  if (!form.userId) {
    ElMessage.warning('请选择本地用户')
    return
  }
  saving.value = true
  try {
    await request.post('/api/v2/admin/dingtalk/user-bindings', {
      id: form.id || undefined,
      corpId: form.corpId,
      dingTalkUserId: form.dingTalkUserId.trim(),
      unionId: form.unionId.trim(),
      userId: form.userId,
      enabled: form.enabled
    })
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await loadBindings()
  } finally {
    saving.value = false
  }
}

async function toggleStatus(row: BindingRow) {
  const enabled = row.enabled === 1 ? 0 : 1
  const action = enabled === 1 ? '启用' : '停用'
  await ElMessageBox.confirm(`确认${action}该钉钉用户绑定？`, '提示', { type: 'warning' })
  await request.patch(`/api/v2/admin/dingtalk/user-bindings/${row.id}/status`, { enabled })
  ElMessage.success(`${action}成功`)
  await loadBindings()
}

async function deleteBinding(row: BindingRow) {
  if (!canEdit.value) return
  await ElMessageBox.confirm('确认删除该钉钉用户绑定？删除后该钉钉用户将无法通过此绑定免登。', '删除绑定', {
    type: 'warning',
    confirmButtonText: '删除',
    confirmButtonClass: 'el-button--danger'
  })
  await request.delete(`/api/v2/admin/dingtalk/user-bindings/${row.id}`)
  ElMessage.success('删除成功')
  if (list.value.length === 1 && filters.page > 1) {
    filters.page -= 1
  }
  await loadBindings()
}

onMounted(loadBindings)
</script>

<style scoped>
.dingtalk-bindings-page {
  padding: 20px;
}

.binding-cell-main {
  color: #1f2937;
  font-weight: 600;
}

.binding-cell-sub {
  margin-top: 2px;
  color: #8a98ab;
  font-size: 12px;
}

.binding-form :deep(.el-select),
.binding-form :deep(.el-tree-select) {
  width: 100%;
}

.user-tree-node {
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.user-tree-node__label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-tree-node__meta {
  flex: none;
  color: #8a98ab;
  font-size: 12px;
}

.user-tree-node--dept .user-tree-node__label {
  color: #334155;
  font-weight: 600;
}

.binding-field-help {
  display: grid;
  gap: 12px;
  color: #334155;
}

.binding-field-help__section {
  padding: 14px 16px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fff;
}

.binding-field-help__section h3 {
  margin: 0 0 8px;
  color: #1f2937;
  font-size: 15px;
  font-weight: 700;
}

.binding-field-help__section p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
  line-height: 1.7;
}

.binding-field-help__note {
  padding: 10px 12px;
  border: 1px solid #bfdbfe;
  border-radius: 6px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 13px;
  line-height: 1.7;
}

</style>
