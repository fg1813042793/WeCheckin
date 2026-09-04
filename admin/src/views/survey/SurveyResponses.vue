<template>
  <div class="responses-page">
    <el-card class="resp-card" shadow="never">
      <div class="page-head">
        <div class="page-head-main">
          <el-button class="back-btn" @click="goBack">‹ 返回</el-button>
          <div class="title-block">
            <h3>答卷管理</h3>
            <div class="title-meta">
              <span class="survey-title">{{ surveyTitle }}</span>
              <el-tag v-if="surveyId" type="info" size="small">ID: {{ surveyId }}</el-tag>
            </div>
          </div>
        </div>
        <div class="page-head-actions">
          <el-button @click="load">刷新</el-button>
          <el-button type="primary" @click="exportCSV">导出 CSV</el-button>
        </div>
      </div>

      <div class="stat-grid">
        <div class="stat-card">
          <span class="stat-label">答卷总数</span>
          <strong>{{ total }}</strong>
        </div>
        <div class="stat-card">
          <span class="stat-label">本页完成</span>
          <strong class="stat-success">{{ responseStats.completed }}</strong>
        </div>
        <div class="stat-card">
          <span class="stat-label">本页草稿</span>
          <strong class="stat-muted">{{ responseStats.draft }}</strong>
        </div>
        <div class="stat-card">
          <span class="stat-label">本页平均用时</span>
          <strong>{{ formatDuration(responseStats.avgDuration) }}</strong>
        </div>
      </div>

      <div class="toolbar">
        <div class="toolbar-left">
          <el-input v-model="keyword" class="keyword-input" placeholder="搜索昵称/用户ID/设备" clearable @clear="search" @keyup.enter="search" />
          <el-button type="primary" @click="search">搜索</el-button>
        </div>
        <div class="toolbar-actions">
          <el-button size="small" type="danger" :disabled="!selectedIds.length" @click="batchDel">批量删除</el-button>
          <span class="selected-tip">已选 {{ selectedIds.length }} 条</span>
        </div>
      </div>

      <div class="table-shell">
        <el-table
          :data="list"
          v-loading="loading"
          class="response-table"
          stripe
          row-key="id"
          empty-text="暂无答卷数据"
          @selection-change="onSelectionChange"
        >
          <el-table-column type="selection" width="44" />
          <el-table-column prop="id" label="ID" width="76" />
          <el-table-column label="答卷用户" min-width="180">
            <template #default="{ row }">
              <div class="user-cell">
                <div class="user-name">{{ row.nickname || '匿名用户' }}</div>
                <div class="user-meta">用户ID：{{ formatUserId(row.userId) }}</div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status===1 ? 'success' : 'info'" size="small" round>
                {{ row.status===1 ? '已完成' : '草稿' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="用时" width="96">
            <template #default="{ row }"><span class="duration-text">{{ formatDuration(row.duration) }}</span></template>
          </el-table-column>
          <el-table-column label="设备信息" min-width="230" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="device-cell">
                <div class="device-tags">
                  <el-tag v-if="row.deviceType" size="small" effect="plain">{{ row.deviceType }}</el-tag>
                  <el-tag v-if="row.platformType" size="small" type="info" effect="plain">{{ row.platformType }}</el-tag>
                  <el-tag v-if="row.browser" size="small" type="info" effect="plain">{{ row.browser }}</el-tag>
                </div>
                <div class="device-text">{{ row.device || '-' }}</div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="开始时间" min-width="150">
            <template #default="{ row }">{{ formatTime(row.startTime) }}</template>
          </el-table-column>
          <el-table-column label="提交时间" min-width="150">
            <template #default="{ row }">{{ formatTime(row.submitTime) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="140" fixed="right">
            <template #default="{ row }">
              <div class="table-actions">
                <el-button size="small" @click="viewDetail(row)">查看</el-button>
                <el-button size="small" type="danger" plain @click="del(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <div class="admin-pagination">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total,prev,pager,next" background @current-change="load" />
      </div>
    </el-card>

    <el-dialog v-model="detailDialog.visible" class="response-detail-dialog" :title="`答卷详情 #${detailDialog.response?.id}`" width="860px">
      <el-descriptions :column="2" border size="small" class="response-desc">
        <el-descriptions-item label="用户">{{ detailDialog.response?.userId || '匿名' }}</el-descriptions-item>
        <el-descriptions-item label="用时">{{ formatDuration(detailDialog.response?.duration) }}</el-descriptions-item>
        <el-descriptions-item label="设备">{{ detailDialog.response?.device || '-' }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ detailDialog.response?.ip || '-' }}</el-descriptions-item>
        <el-descriptions-item label="开始">{{ formatTime(detailDialog.response?.startTime) }}</el-descriptions-item>
        <el-descriptions-item label="提交">{{ formatTime(detailDialog.response?.submitTime) }}</el-descriptions-item>
      </el-descriptions>
      <el-table :data="answerRows" class="answer-table" stripe size="small">
        <el-table-column prop="questionId" label="题号" width="80" />
        <el-table-column prop="title" label="题目" min-width="200" show-overflow-tooltip />
        <el-table-column prop="value" label="答案" min-width="200" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'
import { showRequestError } from '../../utils/request'

const route = useRoute()
const router = useRouter()
const surveyId = Number(route.query.surveyId || 0)
const surveyTitle = String(route.query.title || `问卷 #${surveyId}`)

const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)
const questions = ref<any[]>([])
const skipTypes = ['divider', 'description', 'pagination', 'questionSet']
const keyword = ref('')
const selectedIds = ref<number[]>([])

const responseStats = computed(() => {
  const completed = list.value.filter((r: any) => r.status === 1).length
  const draft = list.value.filter((r: any) => r.status !== 1).length
  const durations = list.value.map((r: any) => Number(r.duration || 0)).filter((d: number) => d > 0)
  const avgDuration = durations.length ? Math.round(durations.reduce((sum: number, d: number) => sum + d, 0) / durations.length) : 0
  return { completed, draft, avgDuration }
})

function onSelectionChange(rows: any[]) {
  selectedIds.value = rows.map((r: any) => r.id)
}

function formatTime(ms: number) {
  if (!ms) return '-'
  return new Date(ms).toLocaleString()
}
function formatDuration(sec: number) {
  if (!sec) return '-'
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return m > 0 ? `${m}分${s}秒` : `${s}秒`
}
function formatUserId(userId: any) {
  return (userId === 0 || userId === '0' || userId === undefined || userId === null) ? '-' : userId
}

async function load() {
  if (!surveyId) { ElMessage.error('缺少 surveyId'); return }
  loading.value = true
  try {
    const [res, detailRes]: any = await Promise.all([
      adminApi.surveyResponseList({ surveyId, page: page.value, pageSize: pageSize.value, keyword: keyword.value || undefined }),
      adminApi.surveyDetail(surveyId)
    ])
    list.value = res.data?.list || res.list || []
    total.value = res.data?.total || res.total || 0
    const detail = (detailRes as any).data || detailRes
    const raw = detail?.schema
    const sch = raw ? (typeof raw === 'string' ? JSON.parse(raw) : raw) : { questions: [] }
    questions.value = (sch.questions || []).filter((q: any) => !skipTypes.includes(q.type))
  } finally { loading.value = false }
}

function search() {
  page.value = 1
  load()
}

async function batchDel() {
  if (!selectedIds.value.length) return
  await ElMessageBox.confirm(`确认删除选中的 ${selectedIds.value.length} 条答卷？`, '提示', { type: 'warning' })
  await adminApi.surveyResponseBatchDel({ ids: selectedIds.value.join(',') })
  ElMessage.success('已删除')
  selectedIds.value = []
  load()
}

const detailDialog = reactive({ visible: false, response: null as any, answers: {} as any, schema: null as any })
const answerRows = computed(() => {
  if (!detailDialog.response) return []
  const sch = detailDialog.schema
  const ans = detailDialog.answers || {}
  if (!sch) {
    return Object.entries(ans).map(([k, v]) => ({ questionId: k, title: k, value: formatVal(v) }))
  }
  return sch.questions.filter((q: any) => !skipTypes.includes(q.type)).map((q: any, i: number) => ({
    questionId: i + 1,
    title: stripHtml(q.title || ''),
    value: formatVal(ans[q.id])
  }))
})
function stripHtml(s: string) {
  const d = document.createElement('div')
  d.innerHTML = s
  return d.textContent || d.innerText || ''
}
function formatVal(v: any) {
  if (v == null) return '(未填)'
  if (Array.isArray(v)) return v.join(', ')
  if (typeof v === 'object') return JSON.stringify(v)
  return v
}

async function viewDetail(row: any) {
  const raw: any = await adminApi.surveyResponseDetail(row.id)
  const data = raw.data || raw
  detailDialog.response = data.response || data
  detailDialog.answers = data.answers || {}
  detailDialog.schema = data.schema || null
  detailDialog.visible = true
}

async function exportCSV() {
  try {
    const res: any = await adminApi.surveyResponseExport(surveyId)
    const url = URL.createObjectURL(new Blob([res], { type: 'text/csv;charset=utf-8' }))
    const a = document.createElement('a')
    a.href = url; a.download = `survey_${surveyId}.csv`; a.click()
    URL.revokeObjectURL(url)
  } catch (error) { showRequestError(error, '导出失败') }
}

async function del(row: any) {
  await ElMessageBox.confirm(`确认删除答卷 #${row.id}?`, '提示', { type: 'warning' })
  await adminApi.surveyResponseDel({ id: row.id })
  ElMessage.success('已删除')
  load()
}

function goBack() { router.push('/survey') }

onMounted(load)
</script>

<style scoped>
.responses-page {
  --resp-accent:#2563eb;
  --resp-accent-soft:#eff6ff;
  --resp-border:#e8edf5;
  --resp-surface:#f8fafc;
  --resp-text:#1f2937;
  --resp-muted:#667085;
}

.resp-card {
  overflow:visible;
  border:1px solid var(--resp-border);
  border-radius:12px;
}

.page-head {
  display:flex;
  align-items:flex-start;
  justify-content:space-between;
  gap:16px;
  padding-bottom:16px;
  border-bottom:1px solid var(--resp-border);
}
.page-head-main {
  display:flex;
  align-items:flex-start;
  gap:12px;
  min-width:0;
}
.back-btn {
  flex-shrink:0;
}
.title-block {
  min-width:0;
}
.title-block h3 {
  margin:0;
  color:var(--resp-text);
  font-size:18px;
  font-weight:700;
  line-height:28px;
}
.title-meta {
  display:flex;
  align-items:center;
  gap:8px;
  margin-top:4px;
  min-width:0;
  color:var(--resp-muted);
  font-size:12px;
}
.survey-title {
  max-width:520px;
  overflow:hidden;
  text-overflow:ellipsis;
  white-space:nowrap;
}
.page-head-actions {
  display:flex;
  align-items:center;
  justify-content:flex-end;
  gap:8px;
  flex-shrink:0;
}

.stat-grid {
  display:grid;
  grid-template-columns:repeat(4,minmax(0,1fr));
  gap:12px;
  margin:16px 0;
}
.stat-card {
  min-height:66px;
  padding:12px 14px;
  background:#fff;
  border:1px solid var(--resp-border);
  border-radius:10px;
  box-shadow:0 1px 2px rgba(15,23,42,0.03);
}
.stat-label {
  display:block;
  color:var(--resp-muted);
  font-size:12px;
  line-height:18px;
}
.stat-card strong {
  display:block;
  margin-top:6px;
  color:var(--resp-text);
  font-size:20px;
  line-height:24px;
}
.stat-card .stat-success {
  color:#16a34a;
}
.stat-card .stat-muted {
  color:#98a2b3;
}

.toolbar {
  display:flex;
  align-items:center;
  justify-content:space-between;
  gap:12px;
  margin-bottom:16px;
  padding:14px;
  background:var(--resp-surface);
  border:1px solid var(--resp-border);
  border-radius:10px;
}
.toolbar-left,
.toolbar-actions {
  display:flex;
  align-items:center;
  gap:8px;
  flex-wrap:wrap;
}
.keyword-input {
  width:280px;
}
.selected-tip {
  color:var(--resp-muted);
  font-size:12px;
}

.table-shell {
  overflow:hidden;
  border:1px solid var(--resp-border);
  border-radius:10px;
}
.response-table :deep(.el-table__header th) {
  background:#f8fafc;
  color:#475467;
  font-weight:600;
}
.response-table :deep(.el-table__row) {
  height:64px;
}
.response-table :deep(.el-table__cell) {
  border-bottom-color:#edf1f7;
}
.user-cell,
.device-cell {
  min-width:0;
}
.user-name {
  color:var(--resp-text);
  font-weight:600;
  line-height:20px;
}
.user-meta,
.device-text {
  margin-top:3px;
  color:#98a2b3;
  font-size:12px;
  line-height:18px;
  overflow:hidden;
  text-overflow:ellipsis;
  white-space:nowrap;
}
.device-tags {
  display:flex;
  align-items:center;
  gap:4px;
  flex-wrap:wrap;
}
.duration-text {
  color:var(--resp-accent);
  font-weight:600;
}
.table-actions {
  display:flex;
  align-items:center;
  gap:8px;
}
.table-actions :deep(.el-button + .el-button) {
  margin-left:0;
}
.admin-pagination {
  display:flex;
  justify-content:flex-start;
  margin-top:16px;
}

.response-desc {
  margin-bottom:14px;
}
.answer-table {
  border:1px solid var(--resp-border);
  border-radius:8px;
  overflow:hidden;
}
.answer-table :deep(.el-table__header th) {
  background:#f8fafc;
  color:#475467;
  font-weight:600;
}

@media (max-width: 900px) {
  .page-head,
  .toolbar {
    align-items:stretch;
    flex-direction:column;
  }
  .page-head-actions,
  .toolbar-actions {
    justify-content:flex-start;
  }
  .stat-grid {
    grid-template-columns:repeat(2,minmax(0,1fr));
  }
  .keyword-input {
    width:100%!important;
  }
}

@media (max-width: 560px) {
  .page-head-main {
    flex-direction:column;
  }
  .stat-grid {
    grid-template-columns:1fr;
  }
}
</style>
