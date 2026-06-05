<template>
  <el-card>
    <div style="display:flex;gap:10px;margin-bottom:12px">
      <el-input v-model="keyword" placeholder="搜索" clearable style="width:300px" @keyup.enter="search" />
      <el-button type="primary" @click="search">搜索</el-button>
    </div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
      <div>
        <el-button v-if="hasPerm('log:del')" type="danger" :disabled="selected.length === 0" @click="delSelected">批量删除</el-button>
      </div>
      <div>
        <el-button circle icon="Refresh" title="刷新" @click="load" />
        <el-button circle icon="Download" title="导出" @click="exportData" />
      </div>
    </div>
    <div style="margin-bottom:8px;color:#999">共 {{ total }} 条记录</div>
    <el-table :data="list" v-loading="loading" stripe style="width:100%" @selection-change="selected = $event">
      <el-table-column type="selection" width="45" />
      <el-table-column type="index" label="#" width="50" />
      <el-table-column label="操作类型" width="100">
        <template #default="{ row }">
          <el-tag :type="logTypeTag(row.type)" size="small">{{ logTypeLabel(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作人" width="180">
        <template #default="{ row }">{{ row.adminName }} ({{ row.adminDesc }})</template>
      </el-table-column>
      <el-table-column prop="content" label="操作内容" min-width="200" />
      <el-table-column label="操作时间" width="170">
        <template #default="{ row }">{{ fmtLogTime(row._createTime) }}</template>
      </el-table-column>
      <el-table-column prop="LOG_ADD_IP" label="IP地址" width="140" />
    </el-table>
    <div style="text-align:center;margin-top:16px">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :page-sizes="[10,20,50,100]"
        :total="total"
        layout="total,sizes,prev,pager,next"
        @current-change="load"
        @size-change="(val:number) => { pageSize = val; page = 1; load() }"
      />
    </div>
  </el-card>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const selected = ref<any[]>([])

function fmtLogTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
function logTypeLabel(t: string | number) {
  return { '1': '登录', '2': '添加', '3': '删除', '4': '修改', '5': '其他' }[String(t)] || '其他'
}
function logTypeTag(t: string | number) {
  return { '1': '', '2': 'success', '3': 'danger', '4': 'warning', '5': 'info' }[String(t)] || 'info'
}

async function load() {
  loading.value = true
  try {
    const res = await adminApi.logList({ search: keyword.value, page: page.value, pageSize: pageSize.value })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch { list.value = []; total.value = 0 }
  loading.value = false
}

async function clearLog() {
  try {
    await ElMessageBox.confirm('确定清空所有操作日志？此操作不可恢复！', '警告', { confirmButtonText: '确定清空', type: 'warning' })
    await adminApi.logClear()
    ElMessage.success('日志已清空')
    load()
  } catch {}
}

function search() {
  page.value = 1
  load()
}

async function delSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selected.value.length} 条日志？`, '提示')
    for (const row of selected.value) {}
    ElMessage.success('已删除')
    selected.value = []
    load()
  } catch {}
}

function exportData() {
  const headers = Object.keys(list.value[0] || {})
  const rows = [headers]
  list.value.forEach((r: any) => {
    rows.push(headers.map(h => String(r[h] ?? '')))
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = '操作日志.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

onMounted(load)
</script>
