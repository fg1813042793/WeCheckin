<template>
  <div>
    <el-card>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
        <div style="display:flex;gap:10px;align-items:center">
          <span style="line-height:32px;font-size:16px;font-weight:600">菜单权限管理</span>
          <el-button v-if="hasPerm('menu:add')" type="success" @click="showAdd(0)">+ 新增顶级菜单</el-button>
          <el-button @click="toggleExpand">{{ allExpanded ? '折叠全部' : '展开全部' }}</el-button>
        </div>
        <div>
          <el-button circle icon="Refresh" title="刷新" @click="loadTree" />
        </div>
      </div>
      <el-table :key="tableKey" :data="treeData" v-loading="loading" row-key="id" stripe :default-expand-all="allExpanded" :tree-props="{ children: 'children' }">
        <el-table-column label="菜单名称" min-width="200">
          <template #default="{ row }">
            <span>{{ row.name }}</span>
            <el-tag v-if="row.type === 2" size="small" type="info" style="margin-left:6px">按钮</el-tag>
            <el-tag v-else-if="row.type === 0" size="small" type="warning" style="margin-left:6px">目录</el-tag>
            <el-tag v-else size="small" type="success" style="margin-left:6px">菜单</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路由/路径" width="160" />
        <el-table-column prop="perms" label="权限标识" width="200" />
        <el-table-column label="图标" width="80">
          <template #default="{ row }">
            <el-icon v-if="row.icon" :size="18"><component :is="row.icon" /></el-icon>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="60" />
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="320">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button v-if="hasPerm('menu:add')" size="small" type="primary" @click="showAdd(row.id)">添加子项</el-button>
              <el-button v-if="hasPerm('menu:edit')" size="small" @click="showEdit(row)">编辑</el-button>
              <el-popconfirm v-if="hasPerm('menu:del')" title="确定删除该菜单及其子项？" @confirm="handleDel(row)">
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialog.visible" :title="dialog.title" width="500px">
      <el-form ref="formRef" :model="form" label-width="90px">
        <el-form-item label="菜单类型">
          <el-radio-group v-model="form.type">
            <el-radio :value="0">目录</el-radio>
            <el-radio :value="1">菜单</el-radio>
            <el-radio :value="2">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="菜单名称" prop="name">
          <el-input v-model="form.name" placeholder="请输菜单名称" />
        </el-form-item>
        <el-form-item label="上级菜单" v-if="!dialog.isTop || form.type === 2">
          <el-tree-select v-model="form.parentId" :data="treeData" :props="{ label: 'name', value: 'id' }" placeholder="选择上级菜单" check-strictly clearable />
        </el-form-item>
        <el-form-item label="路由路径" v-if="form.type === 1">
          <el-input v-model="form.path" placeholder="如 /user" />
        </el-form-item>
        <el-form-item label="权限标识" v-if="form.type === 1 || form.type === 2">
          <el-input v-model="form.perms" placeholder="多个逗号分隔，如 user:list,user:add" />
        </el-form-item>
        <el-form-item label="图标" v-if="form.type === 0 || form.type === 1">
          <IconPicker v-model="form.icon" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态" v-if="form.type !== 2 && !dialog.isCreate">
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
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '../../api'
import { ElMessage } from 'element-plus'
import IconPicker from '../../components/IconPicker.vue'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const saving = ref(false)
const treeData = ref<any[]>([])
const allExpanded = ref(true)
const tableKey = ref(0)

const dialog = reactive({
  visible: false,
  title: '',
  isCreate: true,
  isTop: false
})

const form = reactive({
  id: 0,
  name: '',
  parentId: 0,
  path: '',
  perms: '',
  icon: '',
  sort: 0,
  status: 1,
  type: 1
})

async function loadTree() {
  loading.value = true
  try {
    const res = await adminApi.menuTree()
    treeData.value = Array.isArray(res.data) ? res.data : []
  } catch { treeData.value = [] }
  loading.value = false
}

function showAdd(parentId: number) {
  dialog.isCreate = true
  dialog.isTop = parentId === 0
  dialog.title = parentId === 0 ? '新增顶级菜单' : '新增子菜单'
  form.id = 0
  form.name = ''
  form.parentId = parentId
  form.path = ''
  form.perms = ''
  form.icon = ''
  form.sort = 0
  form.status = 1
  form.type = 1
  dialog.visible = true
}

function showEdit(row: any) {
  dialog.isCreate = false
  dialog.isTop = false
  dialog.title = '编辑菜单'
  form.id = row.id
  form.name = row.name
  form.parentId = row.parentId || 0
  form.path = row.path || ''
  form.perms = row.perms || ''
  form.icon = row.icon || ''
  form.sort = row.sort || 0
  form.status = row.status
  form.type = row.type || 1
  dialog.visible = true
}

async function handleSave() {
  if (!form.name) {
    ElMessage.warning('请输入菜单名称')
    return
  }
  saving.value = true
  try {
    if (dialog.isCreate) {
      await adminApi.menuAdd({ name: form.name, parentId: form.parentId, path: form.path, perms: form.perms, icon: form.icon, sort: form.sort, type: form.type })
      ElMessage.success('添加成功')
    } else {
      await adminApi.menuEdit({ id: form.id, name: form.name, parentId: form.parentId, path: form.path, perms: form.perms, icon: form.icon, sort: form.sort, status: form.status, type: form.type })
      ElMessage.success('保存成功')
    }
    dialog.visible = false
    await loadTree()
  } catch { ElMessage.error('操作失败') }
  saving.value = false
}

async function handleDel(row: any) {
  try {
    await adminApi.menuDel({ id: row.id })
    ElMessage.success('删除成功')
    await loadTree()
  } catch { ElMessage.error('删除失败') }
}

function toggleExpand() {
  allExpanded.value = !allExpanded.value
  tableKey.value++
}

onMounted(loadTree)
</script>
