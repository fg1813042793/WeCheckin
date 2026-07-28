<template>
  <div class="admin-page role-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-input v-model="keyword" placeholder="搜索角色名称" clearable style="width:300px" @keyup.enter="search" />
          <el-button type="primary" @click="search">搜索</el-button>
        </div>
      </div>
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-button v-if="hasPerm('role:add')" type="success" @click="showAdd">+ 新增角色</el-button>
          <el-button v-if="hasPerm('role:del')" type="danger" :disabled="selected.length === 0" @click="delSelected">批量删除</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" @click="loadList" />
          <el-button circle icon="Download" title="导出" @click="exportData" />
        </div>
      </div>
      <el-table :data="list" v-loading="loading" stripe style="width:100%" @selection-change="selected = $event">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="角色名称" min-width="140" />
        <el-table-column prop="remark" label="备注" min-width="160" />
        <el-table-column prop="sort" label="排序" width="60" />
        <el-table-column label="数据范围" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ dataScopeLabel(row.dataScope) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="后台入口" width="100">
          <template #default="{ row }">
            <el-tag :type="row.allowAdminLogin === 0 ? 'info' : 'success'" size="small">
              {{ row.allowAdminLogin === 0 ? '未授权' : '已授权' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="后台权限" width="90">
          <template #default="{ row }">{{ (row.adminPermissionKeys || []).length }}</template>
        </el-table-column>
        <el-table-column label="接口权限" width="90">
          <template #default="{ row }">{{ (row.adminApiPermissionKeys || []).length }}</template>
        </el-table-column>
        <el-table-column label="应用权限" width="90">
          <template #default="{ row }">{{ ((row.clientMenuKeys || []).length + (row.dingtalkH5MenuKeys || []).length) || 0 }}</template>
        </el-table-column>
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-button v-if="hasPerm('role:edit')" size="small" type="primary" @click="showEdit(row)">编辑</el-button>
              <el-popconfirm v-if="hasPerm('role:del')" title="确定删除该角色？" @confirm="handleDel(row)">
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="admin-pagination">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :page-sizes="[10,20,50,100]"
          :total="total"
          layout="total,sizes,prev,pager,next"
          @current-change="loadList"
          @size-change="(val:number) => { pageSize = val; page = 1; loadList() }"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialog.visible" :title="dialog.title" width="720px">
      <el-form ref="formRef" :model="form" label-width="90px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" placeholder="角色备注" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="数据范围">
          <el-radio-group v-model="form.dataScope">
            <el-radio :value="1">全部数据</el-radio>
            <el-radio :value="2">本部门数据</el-radio>
            <el-radio :value="3">本人数据</el-radio>
            <el-radio :value="4">自定义</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="后台入口">
          <el-switch v-model="form.allowAdminLogin" active-text="授权" inactive-text="不授权" />
        </el-form-item>
        <el-form-item label="部门选择" v-if="form.dataScope === 4">
          <el-popover trigger="click" placement="bottom-start" :width="400">
            <template #reference>
              <el-input readonly :model-value="deptDisplayText" placeholder="请选择部门" suffix-icon="ArrowDown" style="width:100%" />
            </template>
            <el-tree
              ref="deptTreeRef"
              :data="deptTreeData"
              :props="{ label: 'name' }"
              show-checkbox
              check-strictly
              node-key="id"
              :default-checked-keys="form.deptIds"
              @check="onDeptCheck"
              style="max-height:300px;overflow-y:auto"
            />
          </el-popover>
        </el-form-item>
        <el-divider content-position="left">后台权限</el-divider>
        <el-form-item label="后台菜单">
          <el-tree
            ref="menuTreeRef"
            :data="menuTreeData"
            :props="{ label: 'name' }"
            show-checkbox
            check-strictly
            node-key="key"
            :default-checked-keys="form.adminPermissionKeys"
            @check="onMenuCheck"
            class="permission-tree"
          />
        </el-form-item>
        <el-form-item label="接口权限">
          <el-tree
            ref="apiTreeRef"
            :data="apiTreeData"
            :props="{ label: 'name', children: 'children' }"
            show-checkbox
            node-key="key"
            :default-checked-keys="form.adminApiPermissionKeys"
            @check="onApiCheck"
            class="permission-tree"
          />
        </el-form-item>
        <el-divider content-position="left">应用权限</el-divider>
        <el-form-item label="客户端权限">
          <el-tree
            ref="clientMenuTreeRef"
            :data="clientMenuTreeData"
            :props="{ label: 'name', children: 'children' }"
            show-checkbox
            check-strictly
            node-key="key"
            :default-checked-keys="form.clientMenuKeys"
            @check="onClientMenuCheck"
            class="permission-tree"
          />
        </el-form-item>
        <el-form-item label="钉钉 H5 权限">
          <el-tree
            ref="dingtalkH5MenuTreeRef"
            :data="dingtalkH5MenuTreeData"
            :props="{ label: 'name', children: 'children' }"
            show-checkbox
            node-key="key"
            :default-checked-keys="form.dingtalkH5MenuKeys"
            @check="onDingTalkH5MenuCheck"
            class="permission-tree"
          />
        </el-form-item>
        <el-form-item label="状态" v-if="!dialog.isCreate">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">{{ dialog.isCreate ? '确定' : '保存' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { adminApi } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const saving = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const selected = ref<any[]>([])
const menuTreeData = ref<any[]>([])
const menuTreeRef = ref<any>(null)
const apiTreeData = ref<any[]>([])
const apiTreeRef = ref<any>(null)
const deptTreeData = ref<any[]>([])
const deptTreeRef = ref<any>(null)
const clientMenuTreeData = ref<any[]>([])
const clientMenuTreeRef = ref<any>(null)
const dingtalkH5MenuTreeData = ref<any[]>([])
const dingtalkH5MenuTreeRef = ref<any>(null)

const dialog = reactive({
  visible: false,
  title: '',
  isCreate: true
})

const form = reactive({
  id: 0,
  name: '',
  remark: '',
  sort: 0,
  status: 1,
  dataScope: 1,
  allowAdminLogin: true,
  adminPermissionKeys: [] as string[],
  adminApiPermissionKeys: [] as string[],
  deptIds: [] as number[],
  clientMenuKeys: [] as string[],
  dingtalkH5MenuKeys: [] as string[]
})

function dataScopeLabel(v: number) {
  return { 1: '全部数据', 2: '本部门', 3: '本人', 4: '自定义' }[v] || '全部数据'
}

function deptNameMap() {
  const m: Record<number, string> = {}
  function walk(nodes: any[]) {
    for (const n of nodes) {
      m[n.id] = n.name
      if (n.children) walk(n.children)
    }
  }
  walk(deptTreeData.value)
  return m
}
const deptDisplayText = computed(() => {
  if (!form.deptIds || form.deptIds.length === 0) return ''
  const dmap = deptNameMap()
  return form.deptIds.map((id: number) => dmap[id] || id).join('、')
})

async function loadList() {
  loading.value = true
  try {
    const res = await adminApi.roleList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value || '' })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch { list.value = []; total.value = 0 }
  loading.value = false
}

function search() {
  page.value = 1
  loadList()
}

async function loadMenuTree() {
  try {
    const res = await adminApi.permissionTree({ platform: 'admin', types: 'directory,menu,button' })
    menuTreeData.value = Array.isArray(res.data) ? res.data : []
  } catch { menuTreeData.value = [] }
}

async function loadApiTree() {
  try {
    const res = await adminApi.permissionTree({ platform: 'admin', types: 'api_category,api' })
    apiTreeData.value = Array.isArray(res.data) ? res.data : []
  } catch { apiTreeData.value = [] }
}

async function loadDeptTree() {
  try {
    const res = await adminApi.deptTree()
    deptTreeData.value = Array.isArray(res.data) ? res.data : []
  } catch { deptTreeData.value = [] }
}

async function loadApplicationPermissionTree() {
  try {
    const res = await adminApi.appPermissionTree()
    clientMenuTreeData.value = Array.isArray(res.data?.client) ? res.data.client : []
    dingtalkH5MenuTreeData.value = Array.isArray(res.data?.dingtalkH5) ? res.data.dingtalkH5 : []
    if (clientMenuTreeData.value.length === 0 && dingtalkH5MenuTreeData.value.length === 0) {
      ElMessage.warning('应用权限配置为空，请确认后端已启动最新版本')
    }
  } catch {
    clientMenuTreeData.value = []
    dingtalkH5MenuTreeData.value = []
    ElMessage.error('应用权限加载失败')
  }
}

function showAdd() {
  dialog.isCreate = true
  dialog.title = '新增角色'
  form.id = 0
  form.name = ''
  form.remark = ''
  form.sort = 0
  form.status = 1
  form.dataScope = 1
  form.allowAdminLogin = true
  form.adminPermissionKeys = []
  form.adminApiPermissionKeys = []
  form.deptIds = []
  form.clientMenuKeys = []
  form.dingtalkH5MenuKeys = []
  dialog.visible = true
  nextTick(() => {
    menuTreeRef.value?.setCheckedKeys([])
    apiTreeRef.value?.setCheckedKeys([])
    deptTreeRef.value?.setCheckedKeys([])
    clientMenuTreeRef.value?.setCheckedKeys([])
    dingtalkH5MenuTreeRef.value?.setCheckedKeys([])
  })
}

function showEdit(row: any) {
  dialog.isCreate = false
  dialog.title = '编辑角色'
  form.id = row.id
  form.name = row.name
  form.remark = row.remark || ''
  form.sort = row.sort || 0
  form.status = row.status
  form.dataScope = row.dataScope || 1
  form.allowAdminLogin = row.allowAdminLogin !== 0
  form.adminPermissionKeys = row.adminPermissionKeys || []
  form.adminApiPermissionKeys = row.adminApiPermissionKeys || []
  form.deptIds = row.deptIds || []
  form.clientMenuKeys = row.clientMenuKeys || []
  form.dingtalkH5MenuKeys = row.dingtalkH5MenuKeys || []
  dialog.visible = true
  nextTick(() => {
    menuTreeRef.value?.setCheckedKeys(form.adminPermissionKeys)
    apiTreeRef.value?.setCheckedKeys(form.adminApiPermissionKeys)
    deptTreeRef.value?.setCheckedKeys(form.deptIds)
    clientMenuTreeRef.value?.setCheckedKeys(form.clientMenuKeys)
    dingtalkH5MenuTreeRef.value?.setCheckedKeys(form.dingtalkH5MenuKeys)
  })
}

function onMenuCheck() {
  nextTick(() => {
    form.adminPermissionKeys = menuTreeRef.value?.getCheckedKeys() || []
  })
}

function onApiCheck() {
  nextTick(() => {
    const checked = apiTreeRef.value?.getCheckedKeys() || []
    form.adminApiPermissionKeys = checked.filter((key: string) => key.startsWith('admin:api:'))
  })
}

function onDeptCheck() {
  nextTick(() => {
    form.deptIds = deptTreeRef.value?.getCheckedKeys() || []
  })
}

function onClientMenuCheck() {
  nextTick(() => {
    form.clientMenuKeys = clientMenuTreeRef.value?.getCheckedKeys() || []
  })
}

function onDingTalkH5MenuCheck() {
  nextTick(() => {
    const checked = dingtalkH5MenuTreeRef.value?.getCheckedKeys() || []
    const halfChecked = dingtalkH5MenuTreeRef.value?.getHalfCheckedKeys?.() || []
    form.dingtalkH5MenuKeys = Array.from(new Set([...checked, ...halfChecked]))
  })
}

async function handleSave() {
  if (!form.name) {
    ElMessage.warning('请输入角色名称')
    return
  }
  saving.value = true
  try {
    const payload: any = {
      name: form.name,
      remark: form.remark,
      sort: form.sort,
      dataScope: form.dataScope,
      allowAdminLogin: form.allowAdminLogin ? 1 : 0,
      adminPermissionKeys: form.adminPermissionKeys.join(','),
      adminApiPermissionKeys: form.adminApiPermissionKeys.join(',')
    }
    payload.clientMenuKeys = form.clientMenuKeys.join(',')
    payload.dingtalkH5MenuKeys = form.dingtalkH5MenuKeys.join(',')
    if (form.dataScope === 4) {
      payload.deptIds = form.deptIds.join(',')
    }
    if (!dialog.isCreate) {
      payload.id = form.id
      payload.status = form.status
    }
    if (dialog.isCreate) {
      await adminApi.roleAdd(payload)
      ElMessage.success('添加成功')
    } else {
      await adminApi.roleEdit(payload)
      ElMessage.success('保存成功')
    }
    dialog.visible = false
    await loadList()
  } catch { ElMessage.error('操作失败') }
  saving.value = false
}

async function handleDel(row: any) {
  try {
    await adminApi.roleDel({ id: row.id })
    ElMessage.success('删除成功')
    await loadList()
  } catch { ElMessage.error('删除失败') }
}

async function delSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selected.value.length} 个角色？`, '提示')
    const ids = selected.value.map((r: any) => r.id).join(',')
    await adminApi.roleDels({ ids })
    ElMessage.success('已删除')
    selected.value = []
    loadList()
  } catch {}
}

function exportData() {
  const rows = [['角色名称', '后台入口权限', '状态', '备注']]
  list.value.forEach((r: any) => {
    rows.push([r.name || '', r.allowAdminLogin === 0 ? '未授权' : '已授权', r.status === 1 ? '正常' : '停用', r.remark || ''])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = '角色列表.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

onMounted(() => { loadList(); loadMenuTree(); loadApiTree(); loadDeptTree(); loadApplicationPermissionTree() })
</script>

<style scoped>
.permission-tree {
  width: 100%;
  max-height: 260px;
  padding: 10px 12px;
  overflow-y: auto;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  background: #fbfcff;
}
</style>
