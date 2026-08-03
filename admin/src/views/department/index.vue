<template>
  <div class="admin-page department-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-input v-model="keyword" placeholder="搜索部门名称" clearable style="width:300px" @keyup.enter="refreshTreeView" />
          <el-button type="primary" @click="refreshTreeView">搜索</el-button>
        </div>
      </div>
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-button v-if="hasPerm('admin:menu:department:add')" type="success" @click="showAdd(0)">+ 新增顶级部门</el-button>
          <el-button @click="toggleExpand">{{ allExpanded ? '折叠全部' : '展开全部' }}</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" @click="loadTree" />
        </div>
      </div>
      <el-table :key="tableKey" :data="filteredTreeData" v-loading="loading" row-key="id" stripe :default-expand-all="allExpanded" :tree-props="{ children: 'children' }">
        <el-table-column prop="name" label="部门名称" min-width="200" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-button v-if="hasPerm('admin:menu:department:add')" size="small" type="primary" @click="showAdd(row.id)">添加子部门</el-button>
              <el-button v-if="hasPerm('admin:menu:department:edit')" size="small" @click="showEdit(row)">编辑</el-button>
              <el-popconfirm v-if="hasPerm('admin:menu:department:del')" title="确定删除该部门及其子部门？" @confirm="handleDel(row)">
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialog.visible" :title="dialog.title" width="450px">
      <el-form ref="formRef" :model="form" label-width="80px">
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="上级部门" v-if="!dialog.isTop">
          <el-tree-select v-model="form.parentId" :data="treeData" :props="{ label: 'name', value: 'id' }" placeholder="选择上级部门" check-strictly clearable />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
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
import { ref, reactive, onMounted, computed } from 'vue'
import { adminApi } from '../../api'
import { ElMessage } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const saving = ref(false)
const treeData = ref<any[]>([])
const keyword = ref('')
const allExpanded = ref(false)
const tableKey = ref(0)
const filteredTreeData = computed(() => filterDeptTree(treeData.value, keyword.value))

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
  sort: 0,
  status: 1
})

async function loadTree() {
  loading.value = true
  try {
    const res = await adminApi.deptTree()
    treeData.value = Array.isArray(res.data) ? res.data : []
  } catch {} finally {
    loading.value = false
  }
}

function filterDeptTree(list: any[], rawKeyword: string): any[] {
  const text = rawKeyword.trim().toLowerCase()
  if (!text) return list
  return list
    .map(item => {
      const children = item.children?.length ? filterDeptTree(item.children, rawKeyword) : []
      const hit = String(item.name || '').toLowerCase().includes(text)
      return hit || children.length ? { ...item, children } : null
    })
    .filter(Boolean) as any[]
}

function refreshTreeView() {
  tableKey.value++
}

function toggleExpand() {
  allExpanded.value = !allExpanded.value
  tableKey.value++
}

function showAdd(parentId: number) {
  dialog.isCreate = true
  dialog.isTop = parentId === 0
  dialog.title = parentId === 0 ? '新增顶级部门' : '新增子部门'
  form.id = 0
  form.name = ''
  form.parentId = parentId
  form.sort = 0
  form.status = 1
  dialog.visible = true
}

function showEdit(row: any) {
  dialog.isCreate = false
  dialog.title = '编辑部门'
  form.id = row.id
  form.name = row.name
  form.parentId = row.parentId || 0
  form.sort = row.sort || 0
  form.status = row.status
  dialog.visible = true
}

async function handleSave() {
  if (!form.name) {
    ElMessage.warning('请输入部门名称')
    return
  }
  saving.value = true
  try {
    if (dialog.isCreate) {
      await adminApi.deptAdd({ name: form.name, parentId: form.parentId, sort: form.sort })
      ElMessage.success('添加成功')
    } else {
      await adminApi.deptEdit({ id: form.id, name: form.name, parentId: form.parentId, sort: form.sort, status: form.status })
      ElMessage.success('保存成功')
    }
    dialog.visible = false
    await loadTree()
  } catch {} finally {
    saving.value = false
  }
}

async function handleDel(row: any) {
  try {
    await adminApi.deptDel({ id: row.id })
    ElMessage.success('删除成功')
    await loadTree()
  } catch {}
}

onMounted(() => {
  loadTree()
})
</script>
