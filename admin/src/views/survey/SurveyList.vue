<template>
  <div class="survey-page">
    <div class="page-banner">
      <div class="banner-content">
        <h2>问卷管理</h2>
        <p>创建和管理调查问卷、投票与表单</p>
      </div>
      <el-button type="primary" size="large" @click="goCreate" class="create-btn">
        <el-icon style="margin-right:6px"><Plus /></el-icon>新建问卷
      </el-button>
    </div>

    <el-card shadow="never" class="main-card">
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

      <div class="stat-bar">
        <div class="stat-item"><span class="stat-num">{{ stats.total }}</span> 总问卷</div>
        <div class="stat-item"><span class="stat-num active">{{ stats.published }}</span> 已发布</div>
        <div class="stat-item"><span class="stat-num muted">{{ stats.stopped }}</span> 已停用</div>
      </div>

      <div v-if="!gridView" class="table-wrap">
        <el-table :data="list" v-loading="loading" stripe>
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
              <el-button size="small" @click="goEdit(row)">编辑</el-button>
              <el-button size="small" type="primary" @click="goDesigner(row)">设计</el-button>
              <el-button size="small" @click="goResponses(row)">数据</el-button>
              <el-button size="small" @click="goStatistic(row)">统计</el-button>
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
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div v-else class="card-grid">
        <div v-for="row in list" :key="row.id" class="project-card" @click="goDesigner(row)">
          <div class="card-head" :class="row.status===1?'pub':'stop'">
            <div class="card-type">{{ row.category || '问卷' }}</div>
            <el-tag size="small" :type="row.status===1?'success':'danger'" effect="dark" round>{{ row.status===1?'发布':'停用' }}</el-tag>
          </div>
          <div class="card-body">
            <h4 class="card-title">{{ row.title }}</h4>
            <p class="card-desc" v-if="row.description">{{ row.description.substring(0,60) }}{{ row.description.length>60?'...':'' }}</p>
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
        <div class="view-toggle">
          <el-button :type="gridView?'':'default'" size="small" @click="gridView=false"><el-icon><List /></el-icon></el-button>
          <el-button :type="gridView?'default':''" size="small" @click="gridView=true"><el-icon><Grid /></el-icon></el-button>
        </div>
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
    list.value = res.list || []
    total.value = res.total || 0
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
.survey-page { max-width:1400px; margin:0 auto; }
.page-banner { display:flex; align-items:center; justify-content:space-between; padding:24px 0 16px; }
.banner-content h2 { margin:0; font-size:22px; font-weight:600; color:#1a1a2e; }
.banner-content p { margin:4px 0 0; color:#888; font-size:13px; }
.create-btn { border-radius:8px; padding:10px 24px; font-size:14px; box-shadow:0 4px 12px rgba(251,69,76,0.3); }
.main-card { border-radius:12px; }
.toolbar { display:flex; justify-content:space-between; align-items:center; margin-bottom:16px; flex-wrap:wrap; gap:8px; }
.toolbar-left { display:flex; gap:8px; align-items:center; flex-wrap:wrap; }
.stat-bar { display:flex; gap:24px; margin-bottom:16px; padding:12px 16px; background:#f8f9fc; border-radius:8px; }
.stat-item { font-size:13px; color:#888; }
.stat-num { font-weight:600; color:#333; font-size:16px; margin-right:4px; }
.stat-num.active { color:#67c23a; }
.stat-num.muted { color:#999; }

.table-wrap { margin-top:4px; }
.cell-title { font-weight:500; color:#333; }
.cell-meta { margin-top:4px; display:flex; gap:6px; align-items:center; flex-wrap:wrap; }
.tag-dot { font-size:11px; color:#888; background:#f0f0f0; padding:1px 8px; border-radius:3px; }
.resp-count { font-weight:600; color:#fb454c; }
.time-range { font-size:12px; color:#888; }
.no-limit { color:#bbb; font-size:12px; }

.card-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(300px,1fr)); gap:16px; margin-top:8px; }
.project-card { background:#fff; border:1px solid #f0f0f0; border-radius:12px; overflow:hidden; cursor:pointer; transition:all 0.2s; }
.project-card:hover { border-color:#fb454c; box-shadow:0 4px 20px rgba(251,69,76,0.08); transform:translateY(-2px); }
.card-head { display:flex; justify-content:space-between; align-items:center; padding:14px 18px; }
.card-head.pub { background:linear-gradient(135deg,#f0f9f0,#e8f5e8); }
.card-head.stop { background:linear-gradient(135deg,#f8f9fc,#f0f0f0); }
.card-type { font-size:12px; color:#888; font-weight:500; }
.card-body { padding:14px 18px; min-height:60px; }
.card-title { margin:0; font-size:15px; font-weight:600; color:#1a1a2e; }
.card-desc { margin:6px 0 0; font-size:12px; color:#999; line-height:1.4; }
.card-foot { display:flex; gap:16px; padding:10px 18px; border-top:1px solid #f5f5f5; font-size:12px; color:#999; }
.card-actions { display:none; padding:10px 18px; gap:8px; border-top:1px solid #f0f0f0; background:#fafafa; }
.project-card:hover .card-actions { display:flex; }
.empty-state { grid-column:1/-1; text-align:center; padding:80px 0; color:#aaa; }

.pagination-bar { display:flex; justify-content:space-between; align-items:center; margin-top:16px; }
.view-toggle { display:flex; gap:4px; }
</style>
