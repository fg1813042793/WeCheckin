<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import {
  AdminDrawer,
  AdminPageHeader,
  AdminPageShell,
  AdminSearchBar,
  AdminTablePanel,
} from '@/components/admin-ui'

interface ExampleOrder {
  id: string
  subject: string
  category: string
  owner: string
  status: '处理中' | '已完成' | '已关闭'
  createdAt: string
}

const rows: ExampleOrder[] = [
  { id: 'WF-20260904-001', subject: '市场活动费用申请', category: '费用', owner: '陈明', status: '处理中', createdAt: '2026-09-04 08:45' },
  { id: 'WF-20260903-006', subject: '新员工设备领用', category: '行政', owner: '李青', status: '已完成', createdAt: '2026-09-03 15:10' },
  { id: 'WF-20260902-012', subject: '供应商合同审核', category: '法务', owner: '周宁', status: '已关闭', createdAt: '2026-09-02 10:30' },
]

const filters = reactive({
  keyword: '',
  category: '',
  owner: '',
  status: '',
  dates: [] as string[],
})
const appliedFilters = ref({ ...filters })
const drawerVisible = ref(false)
const selectedRow = ref<ExampleOrder | null>(null)

const filteredRows = computed(() => rows.filter((row) => {
  const keywordMatches = !appliedFilters.value.keyword
    || row.subject.includes(appliedFilters.value.keyword)
    || row.id.includes(appliedFilters.value.keyword)
  const categoryMatches = !appliedFilters.value.category || row.category === appliedFilters.value.category
  const ownerMatches = !appliedFilters.value.owner || row.owner === appliedFilters.value.owner
  const statusMatches = !appliedFilters.value.status || row.status === appliedFilters.value.status
  return keywordMatches && categoryMatches && ownerMatches && statusMatches
}))

function search() {
  appliedFilters.value = { ...filters }
}

function reset() {
  filters.keyword = ''
  filters.category = ''
  filters.owner = ''
  filters.status = ''
  filters.dates = []
  search()
}

function openDetail(row: ExampleOrder) {
  selectedRow.value = row
  drawerVisible.value = true
}

function statusType(status: ExampleOrder['status']) {
  if (status === '已完成') return 'success'
  if (status === '已关闭') return 'info'
  return 'warning'
}
</script>

<template>
  <AdminPageShell width="wide">
    <AdminPageHeader title="流程实例" description="查看与跟踪业务流程" />

    <AdminSearchBar collapsible @search="search" @reset="reset">
      <el-form-item label="关键词">
        <el-input v-model="filters.keyword" clearable placeholder="流程单号或主题" />
      </el-form-item>
      <el-form-item label="流程分类">
        <el-select v-model="filters.category" clearable placeholder="全部分类">
          <el-option label="费用" value="费用" />
          <el-option label="行政" value="行政" />
          <el-option label="法务" value="法务" />
        </el-select>
      </el-form-item>
      <el-form-item label="发起人">
        <el-select v-model="filters.owner" clearable filterable placeholder="全部人员">
          <el-option label="陈明" value="陈明" />
          <el-option label="李青" value="李青" />
          <el-option label="周宁" value="周宁" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="filters.status" clearable placeholder="全部状态">
          <el-option label="处理中" value="处理中" />
          <el-option label="已完成" value="已完成" />
          <el-option label="已关闭" value="已关闭" />
        </el-select>
      </el-form-item>
      <el-form-item label="发起日期">
        <el-date-picker
          v-model="filters.dates"
          type="daterange"
          value-format="YYYY-MM-DD"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
        />
      </el-form-item>
    </AdminSearchBar>

    <AdminTablePanel title="流程列表" :count="filteredRows.length" :empty="filteredRows.length === 0">
      <el-table :data="filteredRows" row-key="id">
        <el-table-column prop="subject" label="流程主题" min-width="240" show-overflow-tooltip />
        <el-table-column prop="id" label="流程单号" width="180" />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column prop="owner" label="发起人" width="120" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="发起时间" width="180" />
        <el-table-column label="操作" width="100" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <span />
        <el-pagination layout="total, prev, pager, next" :total="filteredRows.length" :page-size="10" />
      </template>
    </AdminTablePanel>

    <AdminDrawer v-model="drawerVisible" title="流程详情" size="md">
      <el-descriptions v-if="selectedRow" :column="1" border>
        <el-descriptions-item label="流程主题">{{ selectedRow.subject }}</el-descriptions-item>
        <el-descriptions-item label="流程单号">{{ selectedRow.id }}</el-descriptions-item>
        <el-descriptions-item label="分类">{{ selectedRow.category }}</el-descriptions-item>
        <el-descriptions-item label="发起人">{{ selectedRow.owner }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ selectedRow.status }}</el-descriptions-item>
        <el-descriptions-item label="发起时间">{{ selectedRow.createdAt }}</el-descriptions-item>
      </el-descriptions>
    </AdminDrawer>
  </AdminPageShell>
</template>
