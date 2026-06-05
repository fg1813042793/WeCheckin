<template>
  <div>
    <el-card>
      <div style="display:flex;gap:10px;margin-bottom:16px">
        <span style="line-height:32px;font-size:16px;font-weight:600">部门管理</span>
        <el-button v-if="hasPerm('dept:add')" type="success" @click="showAdd(0)">+ 新增顶级部门</el-button>
      </div>
      <el-table :data="treeData" v-loading="loading" row-key="id" stripe default-expand-all :tree-props="{ children: 'children' }">
        <el-table-column prop="name" label="部门名称" min-width="200" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button v-if="hasPerm('dept:add')" size="small" type="primary" @click="showAdd(row.id)">添加子部门</el-button>
              <el-button v-if="hasPerm('dept:edit')" size="small" @click="showEdit(row)">编辑</el-button>
              <el-popconfirm v-if="hasPerm('dept:del')" title="确定删除该部门及其子部门？" @confirm="handleDel(row)">
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
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '../../api'
import { ElMessage } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const saving = ref(false)
const treeData = ref<any[]>([])

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
  } catch (e) {
    console.error('加载部门树失败', e)
  } finally {
    loading.value = false
  }
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
  } catch (e) {
    console.error('操作失败', e)
  } finally {
    saving.value = false
  }
}

async function handleDel(row: any) {
  try {
    await adminApi.deptDel({ id: row.id })
    ElMessage.success('删除成功')
    await loadTree()
  } catch (e) {
    console.error('删除失败', e)
  }
}

onMounted(() => {
  loadTree()
})
</script>
