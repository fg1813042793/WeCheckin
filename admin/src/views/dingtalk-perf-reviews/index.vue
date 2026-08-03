<template>
  <div class="admin-page dingtalk-perf-reviews-page" aria-label="钉钉 H5 绩效考评单管理">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <el-input
            v-model="filters.keyword"
            clearable
            placeholder="搜索编号 / 员工 / 部门 / 上级 / HRBP"
            style="width:360px"
            @keyup.enter="handleFilter"
            @clear="handleFilter"
          />
          <el-date-picker
            v-model="filters.period"
            type="month"
            value-format="YYYY-MM"
            placeholder="考评月份"
            style="width:160px"
            @change="handleFilter"
          />
          <el-select v-model="filters.status" clearable placeholder="全部状态" style="width:160px" @change="handleFilter">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
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
            @click="deleteSelectedReviews"
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
        <el-table-column label="考评单" min-width="190" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="perf-cell-main">{{ row.reviewNo || `#${row.id}` }}</div>
            <div class="perf-cell-sub">{{ row.employeeAccount || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="employeeAccount" label="员工" min-width="120" show-overflow-tooltip />
        <el-table-column prop="department" label="部门" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">{{ row.department || '-' }}</template>
        </el-table-column>
        <el-table-column prop="period" label="考评月份" width="110" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.statusLabel || row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="数据状态" width="110">
          <template #default="{ row }">
            <el-tag :type="isDeleted(row) ? 'danger' : 'success'" size="small">
              {{ isDeleted(row) ? '软删除' : '正常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="targetScore" label="目标得分" width="90" />
        <el-table-column label="上级分档" width="100">
          <template #default="{ row }">{{ row.managerGrade || '-' }}</template>
        </el-table-column>
        <el-table-column label="HRBP分档" width="100">
          <template #default="{ row }">{{ row.hrbpGrade || '-' }}</template>
        </el-table-column>
        <el-table-column label="最终分档" width="100">
          <template #default="{ row }">{{ row.finalGrade || '-' }}</template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">{{ formatTime(row.editTime || row.addTime) }}</template>
        </el-table-column>
        <el-table-column label="删除时间" width="170">
          <template #default="{ row }">{{ formatTime(row.deletedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <div class="admin-table-actions">
              <el-button
                size="small"
                type="primary"
                :disabled="!canViewDetail"
                @click="openDetail(row)"
              >
                详情
              </el-button>
              <el-button
                v-if="canDelete"
                size="small"
                type="danger"
                plain
                @click="deleteReview(row)"
              >
                删除
              </el-button>
            </div>
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

    <el-dialog v-model="detailVisible" title="绩效考评单详情" width="min(920px, 92vw)" class="perf-detail-dialog">
      <template v-if="detail">
        <el-descriptions :column="3" border class="perf-descriptions">
          <el-descriptions-item label="考评单号">{{ detail.reviewNo || '-' }}</el-descriptions-item>
          <el-descriptions-item label="员工">{{ detail.employeeAccount || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ detail.statusLabel || detail.status || '-' }}</el-descriptions-item>
          <el-descriptions-item label="考评月份">{{ detail.period || '-' }}</el-descriptions-item>
          <el-descriptions-item label="目标月份">{{ detail.nextPeriod || '-' }}</el-descriptions-item>
          <el-descriptions-item label="部门">{{ detail.department || '-' }}</el-descriptions-item>
          <el-descriptions-item label="数据状态">
            <el-tag :type="isDeleted(detail) ? 'danger' : 'success'" size="small">
              {{ isDeleted(detail) ? '软删除' : '正常' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="删除时间">{{ formatTime(detail.deletedAt) }}</el-descriptions-item>
          <el-descriptions-item label="直属上级">{{ detail.managerAccount || '-' }}</el-descriptions-item>
          <el-descriptions-item label="HRBP">{{ detail.hrbpAccount || '-' }}</el-descriptions-item>
          <el-descriptions-item label="实际HRBP">{{ detail.hrbpReviewerAccount || '-' }}</el-descriptions-item>
          <el-descriptions-item label="上级分档">{{ detail.managerGrade || '-' }}</el-descriptions-item>
          <el-descriptions-item label="HRBP分档">{{ detail.hrbpGrade || '-' }}</el-descriptions-item>
          <el-descriptions-item label="最终分档">{{ detail.finalGrade || '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-tabs v-model="detailTab" class="perf-detail-tabs">
          <el-tab-pane label="表单内容" name="content">
            <section class="perf-detail-section">
              <h4>本月目标</h4>
              <pre class="json-block">{{ prettyJSON(detail.objectivesJson) }}</pre>
            </section>
            <section class="perf-detail-section">
              <h4>思考总结</h4>
              <div class="text-block">{{ detail.selfSummary || '暂无' }}</div>
            </section>
            <section class="perf-detail-section">
              <h4>价值观自评</h4>
              <pre class="json-block">{{ prettyJSON(detail.valuesJson) }}</pre>
            </section>
            <section class="perf-detail-section">
              <h4>上级评价</h4>
              <div class="text-block">{{ detail.managerComment || '暂无' }}</div>
            </section>
            <section class="perf-detail-section">
              <h4>HRBP评价</h4>
              <div class="text-block">{{ detail.hrbpComment || '暂无' }}</div>
            </section>
            <section class="perf-detail-section">
              <h4>下月目标</h4>
              <pre class="json-block">{{ prettyJSON(detail.nextObjectivesJson) }}</pre>
            </section>
          </el-tab-pane>
          <el-tab-pane label="流转记录" name="histories">
            <el-table :data="detail.histories || []" stripe>
              <el-table-column prop="action" label="动作" min-width="220" show-overflow-tooltip />
              <el-table-column label="操作人" min-width="160" show-overflow-tooltip>
                <template #default="{ row }">
                  <div class="perf-cell-main">{{ row.byName || row.byAccount || '-' }}</div>
                  <div class="perf-cell-sub">{{ row.byAccount || '-' }}</div>
                </template>
              </el-table-column>
              <el-table-column label="时间" width="170">
                <template #default="{ row }">{{ formatTime(row.addTime) }}</template>
              </el-table-column>
            </el-table>
          </el-tab-pane>
          <el-tab-pane label="审计字段" name="audit">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="创建人ID">{{ detail.createBy || '-' }}</el-descriptions-item>
              <el-descriptions-item label="更新人ID">{{ detail.updateBy || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建部门ID">{{ detail.createDeptId || '-' }}</el-descriptions-item>
              <el-descriptions-item label="更新部门ID">{{ detail.updateDeptId || '-' }}</el-descriptions-item>
              <el-descriptions-item label="删除人ID">{{ detail.deleteBy || '-' }}</el-descriptions-item>
              <el-descriptions-item label="删除部门ID">{{ detail.deleteDeptId || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(detail.addTime) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatTime(detail.editTime) }}</el-descriptions-item>
              <el-descriptions-item label="删除时间">{{ formatTime(detail.deletedAt) }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
        </el-tabs>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../../utils/request'
import { hasPerm } from '../../utils/permission'

type StatusOption = {
  value: string
  label: string
}

type ReviewRow = {
  id: number
  reviewNo: string
  employeeAccount: string
  managerAccount: string
  hrbpAccount: string
  hrbpReviewerAccount?: string
  department: string
  period: string
  nextPeriod: string
  status: string
  statusLabel: string
  targetScore: number
  managerGrade: string
  hrbpGrade: string
  finalGrade: string
  employeeConfirm: string
  addTime: number
  editTime: number
  createBy: number
  updateBy: number
  createDeptId: number
  updateDeptId: number
  deleteBy: number
  deleteDeptId: number
  deletedAt: number
}

type HistoryRow = {
  id: number
  reviewId: number
  reviewNo: string
  byAccount: string
  byName: string
  action: string
  addTime: number
  editTime: number
}

type ReviewDetail = ReviewRow & {
  objectivesJson: string
  nextObjectivesJson: string
  valuesJson: string
  selfSummary: string
  managerComment: string
  hrbpComment: string
  employeeConfirmComment: string
  finalNote: string
  histories: HistoryRow[]
  departmentLevel1: string
  departmentLevel2: string
  departmentLevel3: string
}

type ReviewListResponse = {
  list: ReviewRow[]
  total: number
  statusOptions: StatusOption[]
}

const loading = ref(false)
const detailLoading = ref(false)
const list = ref<ReviewRow[]>([])
const selected = ref<ReviewRow[]>([])
const total = ref(0)
const statusOptions = ref<StatusOption[]>([])
const detail = ref<ReviewDetail | null>(null)
const detailVisible = ref(false)
const detailTab = ref('content')

const filters = reactive({
  page: 1,
  pageSize: 20,
  keyword: '',
  period: '',
  status: ''
})

const canViewDetail = computed(() => hasPerm('admin:menu:dingtalk:perf-reviews:detail') || hasPerm('admin:menu:dingtalk:perf-reviews:list'))
const canDelete = computed(() => hasPerm('admin:menu:dingtalk:perf-reviews:del'))

function statusTagType(status: string) {
  if (status === 'completed') return 'success'
  if (status === 'employee_draft') return 'info'
  if (status === 'employee_fill') return 'warning'
  return 'primary'
}

function isDeleted(row: Pick<ReviewRow, 'deletedAt'> | null | undefined) {
  return Number(row?.deletedAt || 0) > 0
}

function formatTime(value: number) {
  if (!value) return '-'
  const date = new Date(value)
  const pad = (num: number) => String(num).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function prettyJSON(raw: string) {
  const text = String(raw || '').trim()
  if (!text) return '暂无'
  try {
    return JSON.stringify(JSON.parse(text), null, 2)
  } catch {
    return text
  }
}

async function loadList() {
  loading.value = true
  try {
    const res = await request.get<ReviewListResponse>('/api/v2/admin/dingtalk/perf-reviews', {
      params: {
        page: filters.page,
        pageSize: filters.pageSize,
        keyword: filters.keyword.trim(),
        period: filters.period,
        status: filters.status
      }
    })
    const data = res.data || {}
    list.value = Array.isArray(data.list) ? data.list : []
    selected.value = []
    total.value = Number(data.total || 0)
    statusOptions.value = Array.isArray(data.statusOptions) ? data.statusOptions : statusOptions.value
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
  filters.period = ''
  filters.status = ''
  handleFilter()
}

function handleSelectionChange(rows: ReviewRow[]) {
  selected.value = rows
}

async function openDetail(row: ReviewRow) {
  if (!canViewDetail.value) {
    ElMessage.warning('缺少绩效考评单详情权限')
    return
  }
  detailVisible.value = true
  detailTab.value = 'content'
  detail.value = null
  detailLoading.value = true
  try {
    const res = await request.get<ReviewDetail>(`/api/v2/admin/dingtalk/perf-reviews/${row.id}`)
    detail.value = res.data
  } finally {
    detailLoading.value = false
  }
}

async function deleteReview(row: ReviewRow) {
  await ElMessageBox.confirm(`确认删除绩效考评单「${row.reviewNo || row.id}」？删除后将从数据库中移除。`, '删除确认', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消'
  })
  await request.delete(`/api/v2/admin/dingtalk/perf-reviews/${row.id}`)
  ElMessage.success('删除成功')
  await loadList()
}

async function deleteSelectedReviews() {
  if (!canDelete.value || selected.value.length === 0) return
  const ids = selected.value.map(row => row.id).filter(id => Number(id) > 0)
  if (ids.length === 0) return
  await ElMessageBox.confirm(`确认删除选中的 ${ids.length} 个绩效考评单？删除后将从数据库中移除。`, '批量删除确认', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消'
  })
  await request.delete('/api/v2/admin/dingtalk/perf-reviews', {
    data: { ids: ids.join(',') }
  })
  ElMessage.success('批量删除成功')
  selected.value = []
  await loadList()
}

onMounted(loadList)
</script>

<style scoped>
.perf-cell-main {
  color: #1f2937;
  font-weight: 600;
}

.perf-cell-sub {
  margin-top: 2px;
  color: #8a98ab;
  font-size: 12px;
}

.perf-detail-tabs {
  margin-top: 18px;
}

.perf-detail-section {
  padding: 12px 0;
  border-bottom: 1px solid #eef2f7;
}

.perf-detail-section:last-child {
  border-bottom: 0;
}

.perf-detail-section h4 {
  margin: 0 0 10px;
  color: #1f2937;
  font-size: 14px;
}

.json-block,
.text-block {
  margin: 0;
  padding: 12px;
  min-height: 48px;
  max-height: 240px;
  overflow: auto;
  border: 1px solid #e5eaf3;
  border-radius: 6px;
  background: #f8fafc;
  color: #334155;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
}

.perf-detail-dialog :deep(.el-dialog__body) {
  max-height: min(720px, 72vh);
  overflow: auto;
}
</style>
