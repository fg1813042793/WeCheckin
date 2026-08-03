<template>
  <div class="admin-page dingtalk-perf-histories-page" aria-label="钉钉 H5 绩效流转记录管理">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-input
            v-model="filters.keyword"
            clearable
            placeholder="搜索考评单 / 操作人 / 动作"
            style="width:320px"
            @keyup.enter="handleFilter"
            @clear="handleFilter"
          />
          <el-input
            v-model="filters.reviewNo"
            clearable
            placeholder="考评单号"
            style="width:220px"
            @keyup.enter="handleFilter"
            @clear="handleFilter"
          />
          <el-input
            v-model="filters.byAccount"
            clearable
            placeholder="操作人账号"
            style="width:180px"
            @keyup.enter="handleFilter"
            @clear="handleFilter"
          />
          <el-input
            v-model="filters.action"
            clearable
            placeholder="动作"
            style="width:180px"
            @keyup.enter="handleFilter"
            @clear="handleFilter"
          />
          <el-button type="primary" @click="handleFilter">搜索</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
      </div>
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-button
            v-if="canDelete"
            type="danger"
            :disabled="selected.length === 0"
            @click="deleteSelectedHistories"
          >
            批量删除
          </el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" @click="loadList" />
        </div>
      </div>

      <el-table
        :data="list"
        v-loading="loading"
        stripe
        row-key="id"
        style="width:100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column v-if="canDelete" type="selection" width="48" />
        <el-table-column prop="reviewNo" label="考评单号" min-width="190" show-overflow-tooltip />
        <el-table-column prop="action" label="动作" min-width="260" show-overflow-tooltip />
        <el-table-column label="操作人" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="history-cell-main">{{ row.byName || row.byAccount || '-' }}</div>
            <div class="history-cell-sub">{{ row.byAccount || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="reviewId" label="考评单ID" width="100" />
        <el-table-column label="创建人ID" width="100">
          <template #default="{ row }">{{ row.createBy || '-' }}</template>
        </el-table-column>
        <el-table-column label="更新人ID" width="100">
          <template #default="{ row }">{{ row.updateBy || '-' }}</template>
        </el-table-column>
        <el-table-column label="记录时间" width="170">
          <template #default="{ row }">{{ formatTime(row.addTime) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">{{ formatTime(row.editTime) }}</template>
        </el-table-column>
        <el-table-column v-if="canDelete" label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="danger" plain @click="deleteHistory(row)">删除</el-button>
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
          @current-change="loadList"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../../utils/request'
import { hasPerm } from '../../utils/permission'

type HistoryRow = {
  id: number
  reviewId: number
  reviewNo: string
  byAccount: string
  byName: string
  action: string
  addTime: number
  editTime: number
  createBy: number
  updateBy: number
}

type HistoryListResponse = {
  list: HistoryRow[]
  total: number
}

const loading = ref(false)
const list = ref<HistoryRow[]>([])
const selected = ref<HistoryRow[]>([])
const total = ref(0)
const canDelete = computed(() => hasPerm('admin:menu:dingtalk:perf-histories:del'))

const filters = reactive({
  page: 1,
  pageSize: 20,
  keyword: '',
  reviewNo: '',
  byAccount: '',
  action: ''
})

function formatTime(value: number) {
  if (!value) return '-'
  const date = new Date(value)
  const pad = (num: number) => String(num).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

async function loadList() {
  loading.value = true
  try {
    const res = await request.get<HistoryListResponse>('/api/v2/admin/dingtalk/perf-histories', {
      params: {
        page: filters.page,
        pageSize: filters.pageSize,
        keyword: filters.keyword.trim(),
        reviewNo: filters.reviewNo.trim(),
        byAccount: filters.byAccount.trim(),
        action: filters.action.trim()
      }
    })
    const data = res.data || {}
    list.value = Array.isArray(data.list) ? data.list : []
    selected.value = []
    total.value = Number(data.total || 0)
  } finally {
    loading.value = false
  }
}

function handleFilter() {
  filters.page = 1
  loadList()
}

function handleSizeChange() {
  filters.page = 1
  loadList()
}

function resetFilters() {
  filters.keyword = ''
  filters.reviewNo = ''
  filters.byAccount = ''
  filters.action = ''
  handleFilter()
}

function handleSelectionChange(rows: HistoryRow[]) {
  selected.value = rows
}

async function deleteHistory(row: HistoryRow) {
  await ElMessageBox.confirm(`确认删除绩效流转记录「${row.action || row.id}」？删除后将从数据库中移除。`, '删除确认', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消'
  })
  await request.delete(`/api/v2/admin/dingtalk/perf-histories/${row.id}`)
  ElMessage.success('删除成功')
  await loadList()
}

async function deleteSelectedHistories() {
  if (!canDelete.value || selected.value.length === 0) return
  const ids = selected.value.map(row => row.id).filter(id => Number(id) > 0)
  if (ids.length === 0) return
  await ElMessageBox.confirm(`确认删除选中的 ${ids.length} 条绩效流转记录？删除后将从数据库中移除。`, '批量删除确认', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消'
  })
  await request.delete('/api/v2/admin/dingtalk/perf-histories', {
    data: { ids: ids.join(',') }
  })
  ElMessage.success('批量删除成功')
  selected.value = []
  await loadList()
}

onMounted(loadList)
</script>

<style scoped>
.history-cell-main {
  color: #1f2937;
  font-weight: 600;
}

.history-cell-sub {
  margin-top: 2px;
  color: #8a98ab;
  font-size: 12px;
}
</style>
