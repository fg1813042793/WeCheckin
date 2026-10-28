<template>
  <AdminPageShell width="fluid" density="compact" class="survey-page">
    <section class="survey-toolbar-card" aria-label="问卷工具栏">
      <AdminSearchBar title="" class="survey-search" @search="search" @reset="resetSearch">
        <el-input v-model="keyword" :prefix-icon="Search" placeholder="搜索问卷名称或 ID" clearable @keyup.enter="search" />
        <el-input v-model="category" placeholder="按分类筛选" clearable @keyup.enter="search" />
      </AdminSearchBar>
      <div class="survey-actions-row">
        <div class="survey-actions-row__left">
          <el-button type="primary" :icon="Plus" @click="goCreate">新建问卷</el-button>
        </div>
        <div class="survey-actions-row__right">
          <el-button text :icon="RefreshRight" aria-label="刷新问卷列表" title="刷新" :loading="loading" @click="load" />
        </div>
      </div>
    </section>

    <div class="collection-toolbar">
      <div class="status-tabs" role="tablist" aria-label="问卷状态">
        <button v-for="tab in statusTabs" :key="String(tab.value)" type="button" role="tab" :aria-selected="status === tab.value" :class="{ active: status === tab.value }" @click="selectStatus(tab.value)">
          {{ tab.label }}<span>{{ statusCount(tab.value) }}</span>
        </button>
      </div>
      <div class="collection-tools">
        <div class="view-segment" aria-label="展示方式">
          <el-tooltip content="卡片视图"><button type="button" :class="{ active: gridView }" :aria-pressed="gridView" aria-label="卡片视图" @click="gridView = true"><el-icon><Grid /></el-icon></button></el-tooltip>
          <el-tooltip content="列表视图"><button type="button" :class="{ active: !gridView }" :aria-pressed="!gridView" aria-label="列表视图" @click="gridView = false"><el-icon><List /></el-icon></button></el-tooltip>
        </div>
      </div>
    </div>

    <section v-if="gridView" class="survey-collection" aria-label="问卷卡片列表">
      <article v-for="row in list" :key="row.id" class="survey-card" @click="goDesigner(row)">
        <div class="survey-card__topline">
          <span class="survey-card__category"><el-icon><Document /></el-icon>{{ row.category || '问卷' }}</span>
          <span class="status-label" :class="row.status === 1 ? 'published' : 'stopped'"><i />{{ row.status === 1 ? '已发布' : '已停用' }}</span>
        </div>
        <button class="survey-card__content" type="button" :aria-label="`编辑 ${row.title}`" @click.stop="goDesigner(row)">
          <h2>{{ row.title }}</h2>
          <p>{{ row.description || '暂无问卷说明' }}</p>
          <div class="question-sample" aria-hidden="true"><span>01</span><b></b><i></i><b class="short"></b><i></i><b class="shorter"></b></div>
        </button>
        <div class="survey-card__meta"><span><strong>{{ row.responseCount || 0 }}</strong> 份答卷</span><time>{{ fmtDate(row.addTime) }} 更新</time></div>
        <footer class="survey-card__actions" @click.stop>
          <el-button text type="primary" :icon="EditPen" @click="goEdit(row)">编辑</el-button>
          <el-button text :icon="Document" @click="goResponses(row)">数据</el-button>
          <el-button text :icon="DataAnalysis" @click="goStatistic(row)">统计</el-button>
          <el-dropdown trigger="click" @command="(command: string) => handleMore(command, row)">
            <el-button text :icon="MoreFilled" :aria-label="`${row.title} 更多操作`" title="更多操作" />
            <template #dropdown><el-dropdown-menu>
              <el-dropdown-item command="report">查看报表</el-dropdown-item>
              <el-dropdown-item command="copy">复制问卷</el-dropdown-item>
              <el-dropdown-item :command="row.status === 1 ? 'disable' : 'enable'">{{ row.status === 1 ? '停用问卷' : '发布问卷' }}</el-dropdown-item>
              <el-dropdown-item command="del" divided>删除问卷</el-dropdown-item>
            </el-dropdown-menu></template>
          </el-dropdown>
        </footer>
      </article>
      <el-empty v-if="!list.length && !loading" description="暂无问卷数据"><el-button type="primary" @click="goCreate">新建问卷</el-button></el-empty>
    </section>

    <AdminTablePanel v-else title="问卷列表" :count="total" :loading="loading" :empty="!list.length" empty-text="暂无问卷数据" :min-height="320">
      <el-table :data="list" class="survey-table" stripe>
        <el-table-column label="问卷名称" min-width="240"><template #default="{ row }"><button class="table-title" @click="goDesigner(row)">{{ row.title }}</button><small class="table-subtitle">#{{ row.id }} · {{ row.category || '问卷' }}</small></template></el-table-column>
        <el-table-column label="可见性" width="100"><template #default="{ row }"><el-tag size="small" :type="visType(row.visibility)">{{ visText(row.visibility) }}</el-tag></template></el-table-column>
        <el-table-column prop="responseCount" label="答卷数" width="90" align="right" />
        <el-table-column label="最近更新" width="120"><template #default="{ row }">{{ fmtDate(row.addTime) }}</template></el-table-column>
        <el-table-column label="状态" width="100"><template #default="{ row }"><span class="status-label" :class="row.status === 1 ? 'published' : 'stopped'"><i />{{ row.status === 1 ? '已发布' : '已停用' }}</span></template></el-table-column>
        <el-table-column label="操作" width="190" fixed="right"><template #default="{ row }"><div class="table-actions"><el-button text type="primary" @click="goEdit(row)">编辑</el-button><el-button text @click="goResponses(row)">数据</el-button><el-button text @click="goStatistic(row)">统计</el-button><el-dropdown trigger="click" @command="(command: string) => handleMore(command, row)"><el-button text :icon="MoreFilled" :aria-label="`${row.title} 更多操作`" /></el-dropdown></div></template></el-table-column>
      </el-table>
      <template #footer><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" background @current-change="load" /></template>
    </AdminTablePanel>

    <footer v-if="gridView" class="collection-footer"><span>共 {{ total }} 份问卷</span><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="prev, pager, next" background @current-change="load" /></footer>
  </AdminPageShell>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { DataAnalysis, Document, EditPen, Grid, List, MoreFilled, Plus, RefreshRight, Search } from '@element-plus/icons-vue'
import { AdminPageShell, AdminSearchBar, AdminTablePanel } from '../../components/admin-ui'
import { adminApi } from '../../api'

const router = useRouter()
const keyword = ref('')
const category = ref('')
const status = ref<number | null>(null)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)
const gridView = ref(true)
const statusTabs = [{ value: null, label: '全部问卷' }, { value: 1, label: '已发布' }, { value: 0, label: '已停用' }]
function statusCount(value: number | null) { return value === null ? total.value : list.value.filter(item => item.status === value).length }
function visText(value: number) { return { 0: '公开', 1: '登录', 2: '部门' }[value] || '-' }
function visType(value: number) { return ({ 0: 'success', 1: 'primary', 2: 'warning' } as any)[value] || '' }
function fmtDate(value: number) { return value ? new Date(value).toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }) : '-' }
async function load() { loading.value = true; try { const res: any = await adminApi.surveyList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value.trim(), category: category.value.trim(), status: status.value === null ? -1 : status.value }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } finally { loading.value = false } }
function search() { page.value = 1; void load() }
function resetSearch() { keyword.value = ''; category.value = ''; status.value = null; page.value = 1; void load() }
function selectStatus(value: number | null) { status.value = value; page.value = 1; void load() }
function goCreate() { router.push('/survey/designer') }
function goEdit(row: any) { router.push({ path: '/survey/designer', query: { id: String(row.id) } }) }
function goDesigner(row: any) { router.push({ path: '/survey/designer', query: { id: String(row.id) } }) }
function goResponses(row: any) { router.push({ path: '/survey/responses', query: { surveyId: String(row.id), title: row.title } }) }
function goStatistic(row: any) { router.push({ path: '/survey/statistic', query: { surveyId: String(row.id), title: row.title } }) }
function goStatReport(row: any) { router.push({ path: '/survey/stat-report', query: { surveyId: String(row.id), title: row.title } }) }
async function toggleStatus(row: any) { const next = row.status === 1 ? 0 : 1; await adminApi.surveyStatus({ id: row.id, status: next }); row.status = next; ElMessage.success(next === 1 ? '已发布' : '已停用') }
async function handleMore(command: string, row: any) {
  if (command === 'report') return goStatReport(row)
  if (command === 'copy') { await adminApi.surveyCopy({ id: row.id }); ElMessage.success('已复制'); return void load() }
  if (command === 'enable' || command === 'disable') { await toggleStatus(row); return void load() }
  if (command === 'del') { try { await ElMessageBox.confirm(`确认删除「${row.title}」?`, '删除问卷', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }) } catch { return }; await adminApi.surveyDel({ id: row.id }); ElMessage.success('已删除'); void load() }
}
onMounted(load)
</script>

<style scoped>
.survey-page { --survey-border:var(--admin-ui-color-border); --survey-text:var(--admin-ui-color-text); --survey-muted:var(--admin-ui-color-text-secondary); }
.survey-toolbar-card { overflow:hidden; border:1px solid var(--survey-border); border-radius:8px; background:#fff; }
.survey-search.admin-ui-search-bar { border:0; border-radius:0; background:transparent; }
.survey-search .admin-ui-search-bar__content { padding:12px 14px; align-items:center; }
.survey-search .admin-ui-search-bar__fields { display:flex; flex:initial; gap:10px; }
.survey-search .el-input { width:270px; }
.survey-search .admin-ui-search-bar__actions { gap:8px; }
.survey-actions-row { display:flex; align-items:center; justify-content:space-between; gap:12px; min-height:54px; padding:0 14px; border-top:1px solid var(--survey-border); }
.survey-actions-row__left, .survey-actions-row__right { display:flex; align-items:center; gap:8px; }
.survey-actions-row__right .el-button { width:32px; height:32px; margin:0; }
.collection-toolbar { display:flex; align-items:center; justify-content:space-between; gap:12px; min-height:48px; border-bottom:1px solid var(--survey-border); }
.status-tabs { display:flex; align-self:stretch; gap:22px; }
.status-tabs button { display:flex; align-items:center; gap:8px; padding:0; border:0; border-bottom:2px solid transparent; background:transparent; color:var(--survey-muted); font-size:13px; cursor:pointer; }
.status-tabs button.active { color:var(--admin-ui-color-primary); border-bottom-color:var(--admin-ui-color-primary); font-weight:600; }
.status-tabs button > span { min-width:18px; padding:1px 5px; border-radius:4px; background:#e9edf1; color:#737d8b; font-size:11px; text-align:center; }
.status-tabs button.active > span { background:#e5effc; color:#2878d0; }
.collection-tools { display:flex; align-items:center; gap:10px; padding-bottom:8px; }
.view-segment { display:flex; padding:2px; border:1px solid var(--survey-border); border-radius:6px; background:#e9edf1; }
.view-segment button { display:grid; place-items:center; width:30px; height:26px; padding:0; border:0; border-radius:4px; background:transparent; color:var(--survey-muted); cursor:pointer; }
.view-segment button.active { background:#fff; color:var(--admin-ui-color-primary); box-shadow:0 1px 3px #20262e12; }
.survey-collection { display:grid; grid-template-columns:repeat(auto-fill,minmax(235px,1fr)); gap:16px; }
.survey-collection > .el-empty { grid-column:1 / -1; padding:60px 0; }
.survey-card { display:flex; min-width:0; height:290px; flex-direction:column; overflow:hidden; border:1px solid var(--survey-border); border-radius:8px; background:#fff; transition:border-color 160ms, box-shadow 160ms; }
.survey-card:hover { border-color:#b8c5d6; box-shadow:0 4px 12px #20262e08; }
.survey-card__topline { display:flex; align-items:center; justify-content:space-between; gap:8px; padding:16px 16px 0; }
.survey-card__category { display:flex; align-items:center; gap:6px; overflow:hidden; color:var(--survey-muted); font-size:11px; text-overflow:ellipsis; white-space:nowrap; }
.survey-card__category .el-icon { color:#8792a2; font-size:15px; }
.status-label { display:inline-flex; align-items:center; gap:5px; color:#7d8794; font-size:11px; white-space:nowrap; }
.status-label i { width:5px; height:5px; border-radius:50%; background:currentColor; }
.status-label.published { color:var(--admin-ui-color-success); }
.survey-card__content { flex:1; min-height:0; width:100%; padding:12px 16px 8px; border:0; background:transparent; color:inherit; text-align:left; cursor:pointer; }
.survey-card__content h2 { overflow:hidden; margin:0; color:var(--survey-text); font-size:15px; font-weight:650; line-height:23px; text-overflow:ellipsis; white-space:nowrap; }
.survey-card__content p { display:-webkit-box; height:36px; margin:6px 0 12px; overflow:hidden; color:var(--survey-muted); font-size:11px; line-height:18px; -webkit-box-orient:vertical; -webkit-line-clamp:2; }
.question-sample { display:grid; grid-template-columns:18px 1fr; align-items:center; gap:8px 5px; padding-top:10px; border-top:1px solid #eef0f3; color:#7a8492; font-size:10px; }
.question-sample span { grid-column:1 / -1; color:#a2aab5; }
.question-sample i { width:9px; height:9px; border:1px solid #bfc8d3; border-radius:50%; }
.question-sample b { height:7px; border-radius:999px; background:#edf1f7; }
.question-sample b.short { width:65%; }
.question-sample b.shorter { width:45%; }
.survey-card__meta { display:flex; align-items:center; justify-content:space-between; gap:8px; padding:10px 16px; color:var(--survey-muted); font-size:10px; }
.survey-card__meta strong { color:var(--survey-text); font-size:14px; font-weight:600; font-variant-numeric:tabular-nums; }
.survey-card__meta time { white-space:nowrap; }
.survey-card__actions { display:flex; align-items:center; gap:2px; padding:5px 8px; border-top:1px solid #edf0f4; }
.survey-card__actions .el-button { height:30px; padding:7px 8px; font-size:12px; }
.survey-card__actions .el-button + .el-button { margin-left:0; }
.survey-card__actions > .el-dropdown { margin-left:auto; }
.collection-footer { display:flex; align-items:center; justify-content:space-between; color:var(--survey-muted); font-size:12px; }
.survey-table .el-table__cell { padding:13px 0; font-size:12px; }
.table-title { padding:0; border:0; background:transparent; color:var(--survey-text); font-size:13px; text-align:left; cursor:pointer; }
.table-title:hover { color:var(--admin-ui-color-primary); }
.table-subtitle { display:block; margin-top:4px; color:#8792a2; font-size:11px; }
.table-actions { display:flex; align-items:center; gap:0; }
.table-actions .el-button { margin:0; padding:6px; font-size:12px; }
@media (max-width:767px) { .survey-search .admin-ui-search-bar__content { padding:12px; } .survey-search .admin-ui-search-bar__fields { display:grid; grid-template-columns:1fr; width:100%; } .survey-search .el-input { width:100%; } .survey-actions-row { min-height:52px; padding:0 12px; } .collection-toolbar { align-items:stretch; flex-direction:column; } .status-tabs { min-height:40px; justify-content:space-between; gap:8px; } .status-tabs button { font-size:12px; } .collection-tools { justify-content:flex-end; } .survey-collection { grid-template-columns:minmax(0,1fr); } .survey-card { height:282px; } }
</style>
