<template>
  <div class="survey-page">
    <el-card shadow="never" class="main-card" style="margin-top:16px">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input v-model="keyword" placeholder="搜索标题" clearable style="width:220px" @keyup.enter="load" />
          <el-input v-model="category" placeholder="分类" clearable style="width:140px" @keyup.enter="load" />
          <el-select v-model="status" placeholder="状态" clearable style="width:120px" @change="load">
            <el-option label="发布" :value="1" />
            <el-option label="停用" :value="0" />
          </el-select>
          <el-button type="primary" @click="load">搜索</el-button>
        </div>
        <div class="toolbar-right">
          <el-tooltip content="刷新"><el-button circle @click="load"><el-icon><Refresh /></el-icon></el-button></el-tooltip>
        </div>
      </div>
      <div class="action-row">
        <el-button type="primary" @click="goCreate"><el-icon style="margin-right:4px"><Plus /></el-icon>新建问卷</el-button>
        <div class="view-toggle">
          <el-button class="view-toggle__btn" :class="{ active: !gridView }" size="small" aria-label="列表视图" @click="gridView=false"><el-icon><List /></el-icon></el-button>
          <el-button class="view-toggle__btn" :class="{ active: gridView }" size="small" aria-label="卡片视图" @click="gridView=true"><el-icon><Grid /></el-icon></el-button>
        </div>
      </div>

      <div class="stat-bar">
        <div class="stat-item"><span class="stat-num">{{ stats.total }}</span> 总问卷</div>
        <div class="stat-item"><span class="stat-num active">{{ stats.published }}</span> 已发布</div>
        <div class="stat-item"><span class="stat-num muted">{{ stats.stopped }}</span> 已停用</div>
      </div>

      <div v-if="!gridView" class="table-wrap">
        <el-table class="survey-table" :data="list" v-loading="loading" stripe>
          <el-table-column prop="id" label="ID" width="70" />
          <el-table-column label="标题" min-width="200">
            <template #default="{ row }">
              <div class="cell-title">{{ row.title }}</div>
              <div class="cell-meta" v-if="row.category || row.tags">
                <el-tag v-if="row.category" size="small" round>{{ row.category }}</el-tag>
                <span v-for="t in (row.tags||'').split(',').filter(Boolean)" :key="t" class="tag-dot">{{ t }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="可见性" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="visType(row.visibility)" round>{{ visText(row.visibility) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="答卷" width="70" align="center">
            <template #default="{ row }">
              <span class="resp-count">{{ row.responseCount || 0 }}</span>
            </template>
          </el-table-column>
          <el-table-column label="时间窗" min-width="190">
            <template #default="{ row }">
              <span v-if="row.startTime || row.endTime" class="time-range">{{ fmtTime(row.startTime) }} ~ {{ fmtTime(row.endTime) }}</span>
              <span v-else class="no-limit">不限</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-switch :model-value="row.status===1" :before-change="()=>toggleStatus(row)" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="340" fixed="right">
            <template #default="{ row }">
              <div class="table-actions">
                <el-button size="small" @click="goEdit(row)">编辑</el-button>
                <el-button size="small" type="primary" @click="goDesigner(row)">设计</el-button>
                <el-button size="small" @click="goResponses(row)">数据</el-button>
                <el-button size="small" @click="goStatistic(row)">统计</el-button>
                <el-button size="small" type="success" @click="goStatReport(row)">报表</el-button>
                <el-dropdown trigger="click" @command="(cmd:string)=>handleMore(cmd,row)">
                  <el-button size="small">更多<el-icon><ArrowDown /></el-icon></el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="copy">复制</el-dropdown-item>
                      <el-dropdown-item :command="row.status===1?'disable':'enable'">
                        {{ row.status===1?'停用':'发布' }}
                      </el-dropdown-item>
                      <el-dropdown-item command="del" divided>删除</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div v-else class="card-grid">
        <div v-for="row in list" :key="row.id" class="project-card" :class="row.status===1?'is-published':'is-stopped'" @click="goDesigner(row)">
          <div class="card-head" :class="row.status===1?'pub':'stop'">
            <div class="card-type">
              <el-icon><Document /></el-icon>
              <span>{{ row.category || '问卷' }}</span>
            </div>
            <el-tag size="small" :type="row.status===1?'success':'danger'" effect="dark" round>{{ row.status===1?'已发布':'停用' }}</el-tag>
          </div>
          <div class="card-body">
            <h4 class="card-title">{{ row.title }}</h4>
            <p class="card-desc" v-if="row.description">{{ row.description.substring(0,60) }}{{ row.description.length>60?'...':'' }}</p>
            <div class="survey-preview-lines" aria-hidden="true">
              <span><i></i><b></b></span>
              <span><i></i><b></b></span>
              <span class="input-line"><b></b></span>
            </div>
          </div>
          <div class="card-foot">
            <span><el-icon><Document /></el-icon> {{ row.responseCount || 0 }} 答卷</span>
            <span v-if="row.startTime || row.endTime"><el-icon><Clock /></el-icon> {{ fmtDate(row.startTime) }}</span>
          </div>
          <div class="card-actions" @click.stop>
            <el-button size="small" @click="goEdit(row)">编辑</el-button>
            <el-button size="small" @click="goResponses(row)">数据</el-button>
            <el-button size="small" @click="goStatistic(row)">统计</el-button>
            <el-popconfirm title="确认删除?" @confirm="delRow(row)">
              <template #reference><el-button size="small" type="danger" plain>删除</el-button></template>
            </el-popconfirm>
          </div>
        </div>
        <div v-if="list.length===0 && !loading" class="empty-state">
          <el-icon :size="48" color="#ddd"><Document /></el-icon>
          <p>还没有问卷，点击右上角「新建问卷」开始</p>
        </div>
      </div>

      <div class="pagination-bar">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total,prev,pager,next" @current-change="load" background />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { adminApi } from '../../api'

const router = useRouter()
const keyword = ref('')
const category = ref('')
const status = ref<number|null>(null)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)
const gridView = ref(true)

const stats = computed(() => {
  const total = list.value.length
  const published = list.value.filter((x:any) => x.status===1).length
  const stopped = list.value.filter((x:any) => x.status===0).length
  return { total, published, stopped }
})

function visText(v: number) { return { 0:'公开', 1:'登录', 2:'部门' }[v] || '-' }
function visType(v: number) { return ({ 0:'success', 1:'primary', 2:'warning' } as any)[v] || '' }
function fmtTime(ms: number) { return ms ? new Date(ms).toLocaleDateString() : '-' }
function fmtDate(ms: number) { return ms ? new Date(ms).toLocaleDateString() : '' }

async function load() {
  loading.value = true
  try {
    const res: any = await adminApi.surveyList({
      page: page.value, pageSize: pageSize.value,
      keyword: keyword.value, category: category.value,
      status: status.value === null ? -1 : status.value
    })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally { loading.value = false }
}

async function toggleStatus(row: any) {
  const ns = row.status === 1 ? 0 : 1
  await adminApi.surveyStatus({ id: row.id, status: ns })
  row.status = ns
  ElMessage.success(ns ? '已发布' : '已停用')
  return true
}

function goCreate() { router.push('/survey/designer') }
function goEdit(row: any) { router.push({ path: '/survey/designer', query: { id: String(row.id) } }) }
function goDesigner(row: any) { router.push({ path: '/survey/designer', query: { id: String(row.id) } }) }
function goResponses(row: any) { router.push({ path: '/survey/responses', query: { surveyId: String(row.id), title: row.title } }) }
function goStatistic(row: any) { router.push({ path: '/survey/statistic', query: { surveyId: String(row.id), title: row.title } }) }
function goStatReport(row: any) { router.push({ path: '/survey/stat-report', query: { surveyId: String(row.id), title: row.title } }) }

async function delRow(row: any) {
  await adminApi.surveyDel({ id: row.id })
  ElMessage.success('已删除')
  load()
}

async function handleMore(cmd: string, row: any) {
  if (cmd === 'del') {
    await ElMessageBox.confirm(`确认删除「${row.title}」?`, '提示', { type: 'warning' })
    await adminApi.surveyDel({ id: row.id })
    ElMessage.success('已删除'); load()
  } else if (cmd === 'enable' || cmd === 'disable') {
    await adminApi.surveyStatus({ id: row.id, status: cmd === 'enable' ? 1 : 0 })
    ElMessage.success('已更新'); load()
  } else if (cmd === 'copy') {
    await adminApi.surveyCopy({ id: row.id })
    ElMessage.success('已复制'); load()
  }
}

onMounted(load)
</script>

<style scoped>
.survey-page {
  --survey-accent: #2563eb;
  --survey-accent-soft: #eff6ff;
  --survey-accent-border: #bfdbfe;
  --survey-border: #e8edf5;
  --survey-surface: #f8fafc;
  --survey-text: #1f2937;
  --survey-muted: #667085;
}

.main-card { border-radius:12px; border:1px solid var(--survey-border); }
.toolbar {
  display:flex;
  justify-content:space-between;
  align-items:flex-start;
  flex-wrap:wrap;
  gap:12px;
  margin-bottom:14px;
  padding:14px;
  background:var(--survey-surface);
  border:1px solid var(--survey-border);
  border-radius:10px;
}
.toolbar-left { display:flex; gap:8px; align-items:center; flex-wrap:wrap; }
.toolbar-right { display:flex; align-items:center; gap:8px; }

.action-row { display:flex; gap:10px; align-items:center; margin-bottom:16px; }
.view-toggle {
  display:inline-flex;
  align-items:center;
  gap:2px;
  padding:3px;
  background:#f3f5f8;
  border:1px solid var(--survey-border);
  border-radius:8px;
}
.view-toggle__btn { width:32px; height:28px; padding:0; border:none; color:#667085; background:transparent; }
.view-toggle__btn.active { color:var(--survey-accent); background:#fff; box-shadow:0 4px 10px rgba(15,23,42,0.08); }
.view-toggle :deep(.el-button + .el-button) { margin-left:0; }

.stat-bar {
  display:flex;
  align-items:center;
  flex-wrap:wrap;
  gap:24px;
  margin-bottom:16px;
  padding:12px 16px;
  background:#f8f9fc;
  border-radius:8px;
}
.stat-item {
  font-size:13px;
  color:#888;
}
.stat-num { font-weight:600; color:#333; font-size:16px; margin-right:4px; }
.stat-num.active { color:#67c23a; }
.stat-num.muted { color:#999; }

.table-wrap {
  margin-top:4px;
  overflow:hidden;
  border:1px solid var(--survey-border);
  border-radius:10px;
}
.survey-table :deep(.el-table__header th) { background:#f8fafc; color:#475467; font-weight:600; }
.survey-table :deep(.el-table__row) { height:58px; }
.survey-table :deep(.el-table__cell) { border-bottom-color:#edf1f7; }
.cell-title { font-weight:600; color:var(--survey-text); line-height:20px; }
.cell-meta { margin-top:6px; display:flex; gap:6px; align-items:center; flex-wrap:wrap; }
.tag-dot { font-size:11px; color:#667085; background:#f6f8fb; border:1px solid #edf1f7; padding:1px 8px; border-radius:4px; }
.resp-count { font-weight:700; color:var(--survey-accent); }
.time-range { font-size:12px; color:#667085; }
.no-limit { color:#b0b7c3; font-size:12px; }
.table-actions { display:flex; align-items:center; gap:8px; flex-wrap:wrap; }
.table-actions :deep(.el-button + .el-button) { margin-left:0; }

.card-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(320px,1fr)); gap:18px; margin-top:8px; }
.project-card {
  position:relative;
  display:flex;
  flex-direction:column;
  min-height:262px;
  background:
    linear-gradient(90deg, var(--survey-accent-soft) 0 8px, transparent 8px),
    #fff;
  border:1px solid var(--survey-border);
  border-radius:10px;
  overflow:hidden;
  cursor:pointer;
  box-shadow:0 6px 18px rgba(15,23,42,0.05);
  transition:border-color 0.2s, box-shadow 0.2s, transform 0.2s;
}
.project-card::before {
  content:'';
  position:absolute;
  top:18px;
  bottom:18px;
  left:18px;
  width:1px;
  background:var(--survey-accent);
  opacity:0.26;
}
.project-card.is-stopped {
  background:
    linear-gradient(90deg, #f2f4f7 0 8px, transparent 8px),
    #fff;
}
.project-card.is-stopped::before { background:#98a2b3; }
.project-card:hover { border-color:var(--survey-accent-border); box-shadow:0 12px 28px rgba(15,23,42,0.08); transform:translateY(-2px); }
.card-head,
.card-body,
.card-foot,
.card-actions { position:relative; z-index:1; }
.card-head { display:flex; justify-content:space-between; align-items:center; gap:12px; padding:18px 18px 8px 32px; }
.card-head.pub { background:transparent; }
.card-head.stop { background:transparent; }
.card-type {
  display:inline-flex;
  align-items:center;
  gap:6px;
  min-width:0;
  max-width:70%;
  overflow:hidden;
  text-overflow:ellipsis;
  white-space:nowrap;
  font-size:12px;
  color:#667085;
  font-weight:600;
}
.card-type .el-icon { color:var(--survey-accent); font-size:15px; }
.card-type span { min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.project-card.is-stopped .card-type .el-icon { color:#98a2b3; }
.card-body { flex:1; padding:8px 20px 16px 32px; min-height:126px; }
.card-title {
  display:-webkit-box;
  -webkit-box-orient:vertical;
  -webkit-line-clamp:2;
  overflow:hidden;
  margin:0;
  font-size:16px;
  font-weight:700;
  color:var(--survey-text);
  line-height:22px;
}
.card-desc {
  display:-webkit-box;
  -webkit-box-orient:vertical;
  -webkit-line-clamp:2;
  overflow:hidden;
  margin:8px 0 0;
  font-size:12px;
  color:#7a8494;
  line-height:18px;
}
.survey-preview-lines {
  display:flex;
  flex-direction:column;
  gap:9px;
  margin-top:14px;
}
.survey-preview-lines span {
  display:flex;
  align-items:center;
  gap:8px;
  height:12px;
}
.survey-preview-lines i {
  width:10px;
  height:10px;
  border:1px solid #d7dde8;
  border-radius:50%;
  background:#fff;
  flex:0 0 auto;
}
.survey-preview-lines b {
  width:62%;
  height:7px;
  border-radius:999px;
  background:#edf1f7;
  flex:0 1 auto;
}
.survey-preview-lines span:nth-child(2) b { width:46%; }
.survey-preview-lines .input-line {
  height:22px;
  margin-top:2px;
  padding-left:18px;
}
.survey-preview-lines .input-line b {
  width:78%;
  height:20px;
  border:1px dashed #dce3ee;
  background:#fbfcfe;
  border-radius:6px;
}
.card-foot {
  display:flex;
  gap:14px;
  flex-wrap:wrap;
  padding:11px 18px 11px 32px;
  border-top:1px solid #f0f3f8;
  background:#fbfcfe;
  font-size:12px;
  color:#667085;
}
.card-foot span { display:inline-flex; align-items:center; gap:4px; }
.card-actions {
  display:flex;
  justify-content:flex-end;
  flex-wrap:wrap;
  padding:10px 18px 14px 32px;
  gap:8px;
  border-top:1px solid #f0f3f8;
  background:#fff;
}
.card-actions :deep(.el-button + .el-button) { margin-left:0; }
.empty-state { grid-column:1/-1; text-align:center; padding:80px 0; color:#98a2b3; }

.pagination-bar { display:flex; justify-content:flex-start; align-items:center; gap:12px; margin-top:16px; }

@media (max-width: 768px) {
  .toolbar { align-items:stretch; }
  .toolbar-left,
  .toolbar-right { width:100%; }
  .toolbar-left :deep(.el-input),
  .toolbar-left :deep(.el-select) { width:100%!important; }
  .action-row { justify-content:space-between; }
  .stat-bar { gap:16px; }
  .card-grid { grid-template-columns:1fr; }
  .card-actions { justify-content:flex-start; }
}
</style>
