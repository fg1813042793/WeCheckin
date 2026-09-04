<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import {
  AdminDialog,
  AdminPageHeader,
  AdminPageShell,
  AdminSearchBar,
  AdminTablePanel,
} from '@/components/admin-ui'

interface ExampleUser {
  id: number
  name: string
  department: string
  status: '启用' | '停用'
  updatedAt: string
}

const rows: ExampleUser[] = [
  { id: 1, name: '陈明', department: '产品部', status: '启用', updatedAt: '2026-09-04 09:30' },
  { id: 2, name: '李青', department: '运营部', status: '启用', updatedAt: '2026-09-03 16:20' },
  { id: 3, name: '周宁', department: '财务部', status: '停用', updatedAt: '2026-09-02 11:05' },
]

const keyword = ref('')
const appliedKeyword = ref('')
const dialogVisible = ref(false)
const editingName = ref('')

const filteredRows = computed(() => {
  const value = appliedKeyword.value.trim()
  if (!value) return rows
  return rows.filter((row) => row.name.includes(value) || row.department.includes(value))
})

function search() {
  appliedKeyword.value = keyword.value
}

function reset() {
  keyword.value = ''
  appliedKeyword.value = ''
}

function openCreateDialog() {
  editingName.value = ''
  dialogVisible.value = true
}
</script>

<template>
  <AdminPageShell width="wide">
    <AdminPageHeader title="用户示例" description="管理用户基础信息与启用状态">
      <template #actions>
        <el-button type="primary" :icon="Plus" @click="openCreateDialog">新增用户</el-button>
      </template>
    </AdminPageHeader>

    <AdminSearchBar :title="''" @search="search" @reset="reset">
      <el-form-item label="关键词">
        <el-input v-model="keyword" clearable placeholder="搜索姓名或部门" />
      </el-form-item>
    </AdminSearchBar>

    <AdminTablePanel title="用户列表" :count="filteredRows.length" :empty="filteredRows.length === 0">
      <el-table :data="filteredRows" row-key="id">
        <el-table-column prop="name" label="姓名" min-width="140" />
        <el-table-column prop="department" label="部门" min-width="160" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === '启用' ? 'success' : 'info'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" width="180" />
        <el-table-column label="操作" width="100" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="editingName = row.name; dialogVisible = true">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <span />
        <el-pagination layout="prev, pager, next" :total="filteredRows.length" :page-size="10" />
      </template>
    </AdminTablePanel>

    <AdminDialog v-model="dialogVisible" :title="editingName ? '编辑用户' : '新增用户'" @confirm="dialogVisible = false">
      <el-form label-position="top">
        <el-form-item label="姓名" required>
          <el-input v-model="editingName" placeholder="请输入姓名" />
        </el-form-item>
      </el-form>
    </AdminDialog>
  </AdminPageShell>
</template>
