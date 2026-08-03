<template>
  <div class="admin-page position-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-input v-model="keyword" placeholder="搜索岗位名称" clearable style="width:300px" @keyup.enter="search" />
          <el-button type="primary" @click="search">搜索</el-button>
        </div>
      </div>
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-tooltip :disabled="canAddPosition" content="缺少岗位新增权限" placement="top">
            <span>
              <el-button type="success" :disabled="!canAddPosition" @click="showAdd">创建岗位</el-button>
            </span>
          </el-tooltip>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" @click="loadList" />
        </div>
      </div>

      <el-table :data="list" v-loading="loading" stripe style="width:100%">
        <el-table-column prop="name" label="岗位名称" min-width="180" />
        <el-table-column prop="sort" label="排序" width="90" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ fmtTime(row.addTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-tooltip :disabled="canEditPosition" content="缺少岗位编辑权限" placement="top">
                <span>
                  <el-button size="small" type="primary" :disabled="!canEditPosition" @click="showEdit(row)">编辑</el-button>
                </span>
              </el-tooltip>
              <el-popconfirm v-if="canDeletePosition" title="确定删除该岗位？" @confirm="handleDel(row)">
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
              <el-tooltip v-else content="缺少岗位删除权限" placement="top">
                <span>
                  <el-button size="small" type="danger" :disabled="!canDeletePosition">删除</el-button>
                </span>
              </el-tooltip>
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

    <el-dialog v-model="dialog.visible" :title="dialog.title" width="460px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="岗位名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入岗位名称" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item v-if="!dialog.isCreate" label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">{{ dialog.isCreate ? '创建' : '保存' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../api'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const saving = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const canAddPosition = computed(() => hasPerm('admin:menu:position:add'))
const canEditPosition = computed(() => hasPerm('admin:menu:position:edit'))
const canDeletePosition = computed(() => hasPerm('admin:menu:position:del'))

const dialog = reactive({
  visible: false,
  title: '',
  isCreate: true
})

const form = reactive({
  id: 0,
  name: '',
  sort: 0,
  status: 1
})

function fmtTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function loadList() {
  loading.value = true
  try {
    const res = await adminApi.positionList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value })
    list.value = Array.isArray(res.data?.list) ? res.data.list : []
    total.value = Number(res.data?.total || 0)
  } catch {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  loadList()
}

function showAdd() {
  dialog.isCreate = true
  dialog.title = '创建岗位'
  form.id = 0
  form.name = ''
  form.sort = 0
  form.status = 1
  dialog.visible = true
}

function showEdit(row: any) {
  dialog.isCreate = false
  dialog.title = '编辑岗位'
  form.id = row.id
  form.name = row.name || ''
  form.sort = row.sort || 0
  form.status = row.status === 0 ? 0 : 1
  dialog.visible = true
}

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入岗位名称')
    return
  }
  saving.value = true
  try {
    if (dialog.isCreate) {
      await adminApi.positionAdd({ name: form.name.trim(), sort: form.sort })
      ElMessage.success('创建成功')
    } else {
      await adminApi.positionEdit({ id: form.id, name: form.name.trim(), sort: form.sort, status: form.status })
      ElMessage.success('保存成功')
    }
    dialog.visible = false
    await loadList()
  } finally {
    saving.value = false
  }
}

async function handleDel(row: any) {
  try {
    await adminApi.positionDel({ id: row.id })
    ElMessage.success('删除成功')
    await loadList()
  } catch {}
}

onMounted(() => {
  loadList()
})
</script>
