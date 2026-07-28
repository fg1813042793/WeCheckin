<template>
  <div class="admin-page permission-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-page__title">权限管理</div>
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-select v-model="permissionScope" style="width: 150px" @change="loadTree">
            <el-option v-for="item in permissionScopeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-input v-model="keyword" placeholder="搜索权限名称/编码/路径" clearable style="width:320px" @keyup.enter="refreshTreeView" />
        </div>
        <div class="admin-toolbar__right">
          <el-button v-if="hasPerm('menu:add')" type="primary" @click="showAdd('')">新增权限</el-button>
          <el-button @click="toggleExpand">{{ allExpanded ? '折叠全部' : '展开全部' }}</el-button>
          <el-button circle icon="Refresh" title="刷新" @click="loadTree" />
        </div>
      </div>

      <el-table
        :key="tableKey"
        :data="filteredTreeData"
        v-loading="loading"
        row-key="key"
        stripe
        :default-expand-all="allExpanded"
        :tree-props="{ children: 'children' }"
      >
        <el-table-column label="权限名称" min-width="220">
          <template #default="{ row }">
            <span>{{ row.name }}</span>
            <el-tag size="small" :type="typeTagType(row.type)" style="margin-left:6px">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="key" label="权限编码" min-width="220" />
        <el-table-column prop="resourcePath" label="路径" min-width="160" />
        <el-table-column prop="perms" label="兼容标识" min-width="180" />
        <el-table-column label="图标" width="80">
          <template #default="{ row }">
            <el-icon v-if="resolveAdminIcon(row.icon)" :size="18"><component :is="resolveAdminIcon(row.icon)" /></el-icon>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="70" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="290" fixed="right">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-button v-if="hasPerm('menu:add')" size="small" type="primary" @click="showAdd(row.key)">添加子项</el-button>
              <el-button v-if="hasPerm('menu:edit')" size="small" @click="showEdit(row)">编辑</el-button>
              <el-popconfirm v-if="hasPerm('menu:del')" title="确定删除该权限及其子权限？" @confirm="handleDel(row)">
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialog.visible" :title="dialog.title" width="560px">
      <el-form ref="formRef" :model="form" label-width="96px">
        <el-form-item label="权限编码" prop="permissionKey">
          <el-input v-model="form.permissionKey" :disabled="!dialog.isCreate" placeholder="如 admin:menu:user 或 user:list" />
        </el-form-item>
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入权限名称" />
        </el-form-item>
        <el-form-item label="所属平台">
          <el-select v-model="form.platform" style="width:100%">
            <el-option label="后台管理" value="admin" />
            <el-option label="客户端" value="client" />
            <el-option label="钉钉 H5" value="dingtalk_h5" />
            <el-option label="数据权限" value="data" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限类型">
          <el-select v-model="form.type" style="width:100%">
            <el-option label="目录" value="directory" />
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
            <el-option label="接口类别" value="api_category" />
            <el-option label="接口" value="api" />
            <el-option label="登录入口" value="login" />
            <el-option label="数据权限" value="data" />
          </el-select>
        </el-form-item>
        <el-form-item label="上级权限">
          <el-tree-select
            v-model="form.parentKey"
            :data="treeData"
            :props="{ label: 'name', value: 'key', children: 'children' }"
            node-key="key"
            placeholder="选择上级权限"
            check-strictly
            clearable
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="路径">
          <el-input v-model="form.resourcePath" placeholder="菜单路由或资源路径" />
        </el-form-item>
        <el-form-item label="兼容标识">
          <el-input v-model="form.perms" placeholder="旧权限码，多个用英文逗号分隔" />
        </el-form-item>
        <el-form-item label="图标" v-if="form.platform === 'admin' && (form.type === 'menu' || form.type === 'directory')">
          <IconPicker v-model="form.icon" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
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
import { ref, reactive, onMounted, computed } from 'vue'
import { adminApi } from '../../api'
import { ElMessage } from 'element-plus'
import IconPicker from '../../components/IconPicker.vue'
import { resolveAdminIcon } from '../../icons'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const saving = ref(false)
const treeData = ref<any[]>([])
const keyword = ref('')
const permissionScope = ref('admin_menu')
const allExpanded = ref(false)
const tableKey = ref(0)
const filteredTreeData = computed(() => filterPermissionTree(treeData.value, keyword.value))
const permissionScopeOptions = [
  { label: '后台管理', value: 'admin_menu', platform: 'admin', types: 'directory,menu,button,login', defaultType: 'menu' },
  { label: '接口类别', value: 'admin_api', platform: 'admin', types: 'api_category,api', defaultType: 'api_category' },
  { label: '客户端', value: 'client', platform: 'client', types: '', defaultType: 'menu' },
  { label: '钉钉 H5', value: 'dingtalk_h5', platform: 'dingtalk_h5', types: '', defaultType: 'menu' },
  { label: '数据权限', value: 'data', platform: 'data', types: '', defaultType: 'data' }
]
const activePermissionScope = computed(() => permissionScopeOptions.find(item => item.value === permissionScope.value) || permissionScopeOptions[0])

const dialog = reactive({
  visible: false,
  title: '',
  isCreate: true
})

const form = reactive({
  permissionKey: '',
  name: '',
  platform: 'admin',
  type: 'menu',
  parentKey: '',
  resourcePath: '',
  perms: '',
  icon: '',
  sort: 0,
  status: 1
})

async function loadTree() {
  loading.value = true
  try {
    const res = await adminApi.permissionTree({
      platform: activePermissionScope.value.platform,
      types: activePermissionScope.value.types
    })
    treeData.value = Array.isArray(res.data) ? res.data : []
  } catch { treeData.value = [] }
  loading.value = false
}

function filterPermissionTree(list: any[], rawKeyword: string): any[] {
  const text = rawKeyword.trim().toLowerCase()
  if (!text) return list
  return list
    .map(item => {
      const children = item.children?.length ? filterPermissionTree(item.children, rawKeyword) : []
      const hit = [item.name, item.key, item.resourcePath, item.path, item.perms, item.icon]
        .some(value => String(value || '').toLowerCase().includes(text))
      return hit || children.length ? { ...item, children } : null
    })
    .filter(Boolean) as any[]
}

function refreshTreeView() {
  tableKey.value++
}

function resetForm(parentKey: string) {
  form.permissionKey = ''
  form.name = ''
  form.platform = activePermissionScope.value.platform
  form.type = activePermissionScope.value.defaultType
  form.parentKey = parentKey
  form.resourcePath = ''
  form.perms = ''
  form.icon = ''
  form.sort = 0
  form.status = 1
}

function showAdd(parentKey: string) {
  dialog.isCreate = true
  dialog.title = parentKey ? '新增子权限' : '新增权限'
  resetForm(parentKey)
  dialog.visible = true
}

function showEdit(row: any) {
  dialog.isCreate = false
  dialog.title = '编辑权限'
  form.permissionKey = row.permissionKey || row.key || ''
  form.name = row.name || ''
  form.platform = row.platform || 'admin'
  form.type = row.type || 'menu'
  form.parentKey = row.parentKey || ''
  form.resourcePath = row.resourcePath || row.path || ''
  form.perms = row.perms || ''
  form.icon = row.icon || ''
  form.sort = row.sort || 0
  form.status = row.status
  dialog.visible = true
}

async function handleSave() {
  if (!form.permissionKey.trim()) {
    ElMessage.warning('请输入权限编码')
    return
  }
  if (!form.name.trim()) {
    ElMessage.warning('请输入权限名称')
    return
  }
  saving.value = true
  const payload = {
    key: form.permissionKey,
    name: form.name,
    platform: form.platform,
    type: form.type,
    parentKey: form.parentKey,
    resourcePath: form.resourcePath,
    path: form.resourcePath,
    perms: form.perms,
    icon: form.icon,
    sort: form.sort,
    status: form.status
  }
  try {
    if (dialog.isCreate) {
      await adminApi.permissionAdd(payload)
    ElMessage.success('添加成功')
    } else {
      await adminApi.permissionEdit(payload)
      ElMessage.success('保存成功')
    }
    dialog.visible = false
    permissionScope.value = scopeValueForPermission(form.platform, form.type)
    await loadTree()
  } catch { ElMessage.error('操作失败') }
  saving.value = false
}

function scopeValueForPermission(platform: string, type: string) {
  if (platform === 'admin' && (type === 'api_category' || type === 'api')) return 'admin_api'
  if (platform === 'admin') return 'admin_menu'
  return permissionScopeOptions.find(item => item.platform === platform)?.value || 'admin_menu'
}

async function handleDel(row: any) {
  try {
    await adminApi.permissionDel({ key: row.key })
    ElMessage.success('删除成功')
    await loadTree()
  } catch { ElMessage.error('删除失败') }
}

function toggleExpand() {
  allExpanded.value = !allExpanded.value
  tableKey.value++
}

function typeLabel(type: string) {
  const labels: Record<string, string> = {
    menu: '菜单',
    directory: '目录',
    button: '按钮',
    api_category: '接口类别',
    api: '接口',
    login: '入口',
    data: '数据'
  }
  return labels[type] || type || '-'
}

function typeTagType(type: string) {
  if (type === 'button') return 'info'
  if (type === 'api_category') return 'success'
  if (type === 'api') return 'warning'
  if (type === 'data') return 'success'
  return 'primary'
}

onMounted(loadTree)
</script>
