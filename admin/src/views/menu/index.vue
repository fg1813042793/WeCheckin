<template>
  <div class="admin-page permission-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-select v-model="permissionScope" style="width: 220px" @change="loadTree">
            <el-option v-for="item in permissionScopeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-input v-model="keyword" placeholder="搜索权限名称/编码/路径" clearable style="width:320px" @keyup.enter="refreshTreeView" />
          <el-button type="primary" @click="refreshTreeView">搜索</el-button>
        </div>
      </div>
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-button v-if="hasPerm('admin:menu:permission:add')" type="success" @click="showAdd('')">新增权限</el-button>
          <el-button @click="toggleExpand">{{ allExpanded ? '折叠全部' : '展开全部' }}</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" @click="loadTree" />
        </div>
      </div>

      <el-table
        :key="tableKey"
        :data="filteredTreeData"
        v-loading="loading"
        row-key="key"
        stripe
        style="width:100%"
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
        <el-table-column prop="resourcePath" :label="resourcePathColumnLabel" min-width="180" show-overflow-tooltip />
        <el-table-column label="图标" width="132">
          <template #default="{ row }">
            <span v-if="isDingTalkH5PermissionIcon(row)" class="permission-h5-icon-table" :title="dingtalkH5IconTitle(row.icon)">
              <i class="permission-h5-real-icon u-iconfont" :class="dingtalkH5IconClass(row.icon)" aria-hidden="true"></i>
              <span class="permission-h5-icon-value">{{ resolveDingTalkH5IconValue(row.icon) }}</span>
            </span>
            <el-icon v-else-if="resolveAdminIcon(row.icon)" :size="18"><component :is="resolveAdminIcon(row.icon)" /></el-icon>
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
              <el-button v-if="hasPerm('admin:menu:permission:add')" size="small" type="primary" @click="showAdd(row.key)">添加子项</el-button>
              <el-button v-if="hasPerm('admin:menu:permission:edit')" size="small" @click="showEdit(row)">编辑</el-button>
              <el-popconfirm v-if="hasPerm('admin:menu:permission:del')" title="确定删除该权限及其子权限？" @confirm="handleDel(row)">
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
          <el-input v-model="form.permissionKey" placeholder="如 admin:menu:user 或 user:list" />
        </el-form-item>
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入权限名称" />
        </el-form-item>
        <el-form-item label="所属平台">
          <el-select v-model="form.platform" style="width:100%" @change="normalizePermissionTypeForForm">
            <el-option v-for="item in availablePermissionPlatformOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="权限类型">
          <el-select v-model="form.type" style="width:100%">
            <el-option v-for="item in availablePermissionTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
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
        <el-form-item :label="resourcePathFormLabel">
          <el-input v-model="form.resourcePath" :placeholder="resourcePathPlaceholder" />
        </el-form-item>
        <el-form-item label="图标" v-if="showIconPicker">
          <el-popover v-if="isDingTalkH5MenuIcon" placement="bottom-start" :width="460" trigger="click">
            <template #reference>
              <el-input :model-value="dingtalkH5IconInputText" placeholder="选择 uView 图标" readonly clearable @clear="form.icon = ''">
                <template #prefix>
                  <i v-if="form.icon" class="permission-h5-real-icon permission-h5-real-icon--prefix u-iconfont" :class="dingtalkH5IconClass(form.icon)" aria-hidden="true"></i>
                </template>
              </el-input>
            </template>
            <el-input v-model="dingtalkH5IconKeyword" class="permission-h5-icon-search" placeholder="搜索图标名称/编码" clearable />
            <div class="permission-h5-icon-grid">
              <button
                v-for="item in dingtalkH5FilteredIconOptions"
                :key="item.value"
                type="button"
                class="permission-h5-icon-cell"
                :class="{ active: isDingTalkH5IconActive(item) }"
                :title="`${item.label} ${item.value}`"
                @click="selectDingTalkH5Icon(item.value)"
              >
                <i class="permission-h5-real-icon permission-h5-real-icon--picker u-iconfont" :class="dingtalkH5IconClass(item.value)" aria-hidden="true"></i>
                <span class="permission-h5-icon-name">{{ item.label }}</span>
                <span class="permission-h5-icon-code">{{ item.value }}</span>
              </button>
            </div>
          </el-popover>
          <IconPicker v-else v-model="form.icon" />
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
import { showRequestError } from '../../utils/request'
import '../../styles/uview-iconfont.css'
import {
  DINGTALK_H5_ICON_OPTIONS,
  type DingTalkH5IconOption,
  dingtalkH5IconLabel,
  resolveDingTalkH5IconValue,
} from '../../constants/dingtalkH5Icons'

const loading = ref(false)
const saving = ref(false)
const treeData = ref<any[]>([])
const keyword = ref('')
const permissionScope = ref('admin_menu')
const allExpanded = ref(false)
const tableKey = ref(0)
const originalPermissionKey = ref('')
const dingtalkH5IconKeyword = ref('')
const filteredTreeData = computed(() => filterPermissionTree(treeData.value, keyword.value))
const permissionScopeOptions = [
  { label: '菜单权限/后台管理', value: 'admin_menu', platform: 'admin', types: 'directory,menu,button,login', defaultType: 'menu', defaultPlatform: 'admin' },
  { label: '菜单权限/客户端', value: 'client', platform: 'client', types: 'menu', defaultType: 'menu', defaultPlatform: 'client' },
  { label: '菜单权限/钉钉H5', value: 'dingtalk_h5', platform: 'dingtalk_h5', types: 'directory,menu,button', defaultType: 'menu', defaultPlatform: 'dingtalk_h5' },
  { label: '接口权限', value: 'admin_api', platform: '', types: 'api_category,api', defaultType: 'api_category', defaultPlatform: 'admin' },
  { label: '数据权限', value: 'data', platform: 'data', types: '', defaultType: 'data', defaultPlatform: 'data' }
]
const permissionPlatformOptions = [
  { label: '后台管理', value: 'admin' },
  { label: '客户端', value: 'client' },
  { label: '钉钉 H5', value: 'dingtalk_h5' },
  { label: '数据权限', value: 'data' }
]
const permissionTypeOptions = [
  { label: '目录', value: 'directory' },
  { label: '菜单', value: 'menu' },
  { label: '按钮', value: 'button' },
  { label: '接口分组', value: 'api_category' },
  { label: '接口权限点', value: 'api' },
  { label: '登录入口', value: 'login' },
  { label: '数据权限', value: 'data' }
]
const activePermissionScope = computed(() => permissionScopeOptions.find(item => item.value === permissionScope.value) || permissionScopeOptions[0])
const availablePermissionPlatformOptions = computed(() => {
  const platformSet = new Set(allowedPermissionPlatformsForScope())
  return permissionPlatformOptions.filter(item => platformSet.has(item.value))
})
const dingtalkH5IconInputText = computed(() => {
  const value = resolveDingTalkH5IconValue(form.icon)
  if (!value) return ''
  return `${dingtalkH5IconLabel(form.icon)} / ${value}`
})
const dingtalkH5FilteredIconOptions = computed(() => {
  const keyword = dingtalkH5IconKeyword.value.trim().toLowerCase()
  if (!keyword) return DINGTALK_H5_ICON_OPTIONS
  return DINGTALK_H5_ICON_OPTIONS.filter(item => {
    const aliases = item.aliases?.join(' ') || ''
    return `${item.label} ${item.value} ${aliases}`.toLowerCase().includes(keyword)
  })
})
const availablePermissionTypeOptions = computed(() => {
  const typeSet = new Set(allowedPermissionTypesForForm())
  return permissionTypeOptions.filter(item => typeSet.has(item.value))
})

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
  icon: '',
  sort: 0,
  status: 1
})
const resourcePathColumnLabel = computed(() => permissionScope.value === 'admin_api' ? 'API路径' : '路径')
const resourcePathFormLabel = computed(() => (form.type === 'api' || form.type === 'api_category') ? 'API路径' : '路径')
const resourcePathPlaceholder = computed(() => {
  if (form.type === 'api') return '如 /api/v2/admin/users'
  if (form.type === 'api_category') return '接口分组可留空'
  return '菜单路由或资源路径'
})
const isDingTalkH5MenuIcon = computed(() => form.platform === 'dingtalk_h5' && (form.type === 'menu' || form.type === 'directory'))
const showIconPicker = computed(() => {
  if (form.type !== 'menu' && form.type !== 'directory') return false
  return form.platform === 'admin' || form.platform === 'dingtalk_h5'
})

async function loadTree() {
  loading.value = true
  try {
    const params = {
      types: activePermissionScope.value.types
    } as { platform?: string; types: string }
    if (activePermissionScope.value.platform) params.platform = activePermissionScope.value.platform
    const res = await adminApi.permissionTree(params)
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
      const hit = [item.name, item.key, item.resourcePath, item.path, item.icon]
        .some(value => String(value || '').toLowerCase().includes(text))
      return hit || children.length ? { ...item, children } : null
    })
    .filter(Boolean) as any[]
}

function refreshTreeView() {
  tableKey.value++
}

function resetForm(parentKey: string) {
  originalPermissionKey.value = ''
  form.permissionKey = ''
  form.name = ''
  form.platform = permissionPlatformFromKey(parentKey) || activePermissionScope.value.defaultPlatform
  form.type = defaultPermissionTypeForParent(parentKey)
  form.parentKey = parentKey
  form.resourcePath = ''
  form.icon = ''
  form.sort = 0
  form.status = 1
  normalizePermissionTypeForForm()
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
  originalPermissionKey.value = row.permissionKey || row.key || ''
  form.permissionKey = originalPermissionKey.value
  form.name = row.name || ''
  form.platform = row.platform || 'admin'
  form.type = row.type || 'menu'
  form.parentKey = row.parentKey || ''
  form.resourcePath = row.resourcePath || row.path || ''
  form.icon = row.icon || ''
  form.sort = row.sort || 0
  form.status = row.status
  normalizePermissionTypeForForm()
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
    icon: normalizedPermissionIcon(),
    sort: form.sort,
    status: form.status
  }
  try {
    if (dialog.isCreate) {
      await adminApi.permissionAdd(payload)
      ElMessage.success('添加成功')
    } else {
      await adminApi.permissionEdit({ ...payload, originalKey: originalPermissionKey.value })
      ElMessage.success('保存成功')
    }
    dialog.visible = false
    permissionScope.value = scopeValueForPermission(form.platform, form.type)
    await loadTree()
  } catch (error) { showRequestError(error, '操作失败') }
  saving.value = false
}

function scopeValueForPermission(platform: string, type: string) {
  if (type === 'api_category' || type === 'api') return 'admin_api'
  if (platform === 'admin') return 'admin_menu'
  return permissionScopeOptions.find(item => item.platform === platform)?.value || 'admin_menu'
}

function permissionPlatformFromKey(parentKey: string) {
  if (parentKey.startsWith('dingtalk_h5:')) return 'dingtalk_h5'
  if (parentKey.startsWith('client:')) return 'client'
  if (parentKey.startsWith('data:')) return 'data'
  if (parentKey.startsWith('admin:')) return 'admin'
  return ''
}

function defaultPermissionTypeForParent(parentKey: string) {
  if (parentKey.includes(':api-category:')) return 'api'
  return activePermissionScope.value.defaultType
}

function allowedPermissionTypesForForm() {
  if (permissionScope.value === 'admin_api') {
    return ['api_category', 'api']
  }
  if (form.platform === 'data') {
    return ['data']
  }
  if (form.platform === 'admin') {
    return ['directory', 'menu', 'button', 'login']
  }
  if (form.platform === 'client') {
    return ['menu']
  }
  if (form.platform === 'dingtalk_h5') {
    return ['directory', 'menu', 'button']
  }
  return [activePermissionScope.value.defaultType]
}

function allowedPermissionPlatformsForScope() {
  if (permissionScope.value === 'admin_api') {
    return ['admin', 'client', 'dingtalk_h5']
  }
  if (permissionScope.value === 'data') {
    return ['data']
  }
  if (activePermissionScope.value.platform) {
    return [activePermissionScope.value.platform]
  }
  return ['admin']
}

function normalizePermissionPlatformForForm() {
  const allowed = allowedPermissionPlatformsForScope()
  if (!allowed.includes(form.platform)) {
    form.platform = allowed[0] || activePermissionScope.value.defaultPlatform
  }
}

function normalizePermissionTypeForForm() {
  normalizePermissionPlatformForForm()
  const allowed = allowedPermissionTypesForForm()
  if (!allowed.includes(form.type)) {
    form.type = allowed[0] || activePermissionScope.value.defaultType
  }
}

async function handleDel(row: any) {
  try {
    await adminApi.permissionDel({ key: row.key })
    ElMessage.success('删除成功')
    await loadTree()
  } catch (error) { showRequestError(error, '删除失败') }
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
    api_category: '接口分组',
    api: '接口权限点',
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

function isDingTalkH5PermissionIcon(row: any) {
  return row?.platform === 'dingtalk_h5' && Boolean(row.icon) && (row.type === 'menu' || row.type === 'directory')
}

function dingtalkH5IconClass(icon: string) {
  const value = resolveDingTalkH5IconValue(icon)
  return value ? `uicon-${value}` : ''
}

function dingtalkH5IconTitle(icon: string) {
  const rawValue = String(icon || '').trim()
  const value = resolveDingTalkH5IconValue(rawValue)
  if (!value) return '-'
  const label = dingtalkH5IconLabel(rawValue)
  return rawValue && rawValue !== value ? `${label} / ${value}（兼容旧值：${rawValue}）` : `${label} / ${value}`
}

function isDingTalkH5IconActive(item: DingTalkH5IconOption) {
  return resolveDingTalkH5IconValue(form.icon) === item.value
}

function selectDingTalkH5Icon(icon: string) {
  form.icon = icon
}

function normalizedPermissionIcon() {
  if (form.platform === 'dingtalk_h5' && (form.type === 'menu' || form.type === 'directory')) {
    return resolveDingTalkH5IconValue(form.icon)
  }
  return form.icon
}

onMounted(loadTree)
</script>

<style scoped>
.permission-h5-icon-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  max-height: 360px;
  overflow: auto;
  padding-right: 2px;
}

.permission-h5-icon-search {
  margin-bottom: 10px;
}

.permission-h5-icon-cell {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 78px;
  padding: 8px 6px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: #fff;
  cursor: pointer;
  transition: all 0.18s ease;
}

.permission-h5-icon-cell:hover,
.permission-h5-icon-cell.active {
  border-color: #409eff;
  background: #ecf5ff;
}

.permission-h5-icon-table {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 112px;
  padding: 3px 8px;
  border: 1px solid #dbeafe;
  border-radius: 6px;
  background: #f8fbff;
  color: #1f6feb;
  vertical-align: middle;
}

.permission-h5-real-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  flex: 0 0 22px;
  color: #1f6feb;
  font-size: 18px;
  line-height: 1;
}

.permission-h5-real-icon--prefix {
  width: 18px;
  height: 18px;
  flex-basis: 18px;
  font-size: 16px;
}

.permission-h5-real-icon--picker {
  width: 28px;
  height: 28px;
  flex-basis: 28px;
  border-radius: 8px;
  background: #eff6ff;
  color: #1677ff;
  font-size: 22px;
}

.permission-h5-icon-value,
.permission-h5-icon-code {
  overflow: hidden;
  color: #6b7280;
  font-size: 12px;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.permission-h5-icon-value {
  color: #1f6feb;
}

.permission-h5-icon-name {
  margin-top: 6px;
  color: #1f2937;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
}

.permission-h5-icon-code {
  margin-top: 3px;
  max-width: 100%;
}
</style>
