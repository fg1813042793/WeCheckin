<template>
  <AdminPageShell width="fluid" density="compact" class="survey-page exam-page">
    <section class="survey-toolbar-card" aria-label="考试工具栏">
      <AdminSearchBar title="" class="survey-search" @search="search" @reset="resetSearch">
        <el-input v-model="keyword" :prefix-icon="Search" placeholder="搜索考试名称或 ID" clearable @keyup.enter="search" />
        <el-input v-model="category" placeholder="按分类筛选" clearable @keyup.enter="search" />
        <el-select v-model="statusFilter" placeholder="全部状态" clearable @change="search">
          <el-option label="已发布" :value="1" />
          <el-option label="已停用" :value="0" />
        </el-select>
      </AdminSearchBar>
      <div class="survey-actions-row">
        <div class="survey-actions-row__left">
          <el-button type="primary" :icon="Plus" @click="goCreate">新建考试</el-button>
        </div>
        <div class="survey-actions-row__right">
          <el-button text :icon="RefreshRight" aria-label="刷新考试列表" title="刷新" :loading="loading" @click="load" />
        </div>
      </div>
    </section>

    <div class="collection-toolbar">
      <div class="status-tabs" role="tablist" aria-label="考试状态">
        <button v-for="tab in statusTabs" :key="String(tab.value)" type="button" role="tab" :aria-selected="statusFilter === tab.value" :class="{ active: statusFilter === tab.value }" @click="selectStatus(tab.value)">
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

    <section v-if="gridView" class="survey-collection" aria-label="考试卡片列表">
      <article v-for="row in list" :key="row.id" class="survey-card" @click="goDesigner(row)">
        <div class="survey-card__topline">
          <span class="survey-card__category"><el-icon><Document /></el-icon>{{ row.category || '考试' }}</span>
          <span class="status-label" :class="row.status === 1 ? 'published' : 'stopped'"><i />{{ row.status === 1 ? '已发布' : '已停用' }}</span>
        </div>
        <button class="survey-card__content" type="button" :aria-label="`编辑 ${row.title}`" @click.stop="goDesigner(row)">
          <h2>{{ row.title }}</h2>
          <p>{{ row.description || '暂无考试说明' }}</p>
          <div class="exam-card__preview" aria-hidden="true">
            <div><span>最长交卷</span><strong>{{ maxSubmitMinutes(row) }}{{ maxSubmitMinutes(row) !== '-' ? ' 分钟' : '' }}</strong></div>
            <div><span>答题次数</span><strong>{{ row.maxResponse || 0 }}{{ row.maxResponse ? ' 次' : ' 不限' }}</strong></div>
          </div>
        </button>
        <div class="survey-card__meta"><span><strong>{{ row.maxResponse || 0 }}</strong> {{ row.maxResponse ? '次限制' : '不限次数' }}</span><time>{{ fmtDate(row.addTime) }} 更新</time></div>
        <footer class="survey-card__actions" @click.stop>
          <el-button text type="primary" @click="goDesigner(row)">设计</el-button>
          <el-dropdown trigger="click" @command="(command: string) => handleMore(command, row)">
            <el-button text :icon="MoreFilled" :aria-label="`${row.title} 更多操作`" title="更多操作" />
            <template #dropdown><el-dropdown-menu>
              <el-dropdown-item command="edit">编辑考试</el-dropdown-item>
              <el-dropdown-item :command="row.status === 1 ? 'disable' : 'enable'">{{ row.status === 1 ? '停用考试' : '发布考试' }}</el-dropdown-item>
              <el-dropdown-item command="del" divided>删除考试</el-dropdown-item>
            </el-dropdown-menu></template>
          </el-dropdown>
        </footer>
      </article>
      <el-empty v-if="!list.length && !loading" description="暂无考试数据"><el-button type="primary" @click="goCreate">新建考试</el-button></el-empty>
    </section>

    <AdminTablePanel v-else title="考试列表" :count="total" :loading="loading" :empty="!list.length" empty-text="暂无考试数据" :min-height="320">
      <el-table :data="list" class="survey-table" stripe>
        <el-table-column label="考试名称" min-width="240"><template #default="{ row }"><button class="table-title" @click="goDesigner(row)">{{ row.title }}</button><small class="table-subtitle">#{{ row.id }} · {{ row.category || '考试' }}</small></template></el-table-column>
        <el-table-column label="可见性" width="100"><template #default="{ row }"><el-tag size="small" :type="visType(row.visibility)">{{ visText(row.visibility) }}</el-tag></template></el-table-column>
        <el-table-column label="最长交卷" width="110" align="center"><template #default="{ row }">{{ maxSubmitMinutes(row) }}{{ maxSubmitMinutes(row) !== '-' ? ' 分钟' : '' }}</template></el-table-column>
        <el-table-column label="答题次数" width="100" align="center"><template #default="{ row }">{{ row.maxResponse || 0 }}{{ row.maxResponse ? ' 次' : ' 不限' }}</template></el-table-column>
        <el-table-column label="最近更新" width="120"><template #default="{ row }">{{ fmtDate(row.addTime) }}</template></el-table-column>
        <el-table-column label="状态" width="100"><template #default="{ row }"><span class="status-label" :class="row.status === 1 ? 'published' : 'stopped'"><i />{{ row.status === 1 ? '已发布' : '已停用' }}</span></template></el-table-column>
        <el-table-column label="操作" width="150" fixed="right"><template #default="{ row }"><div class="table-actions"><el-button text type="primary" @click="goDesigner(row)">设计</el-button><el-dropdown trigger="click" @command="(command: string) => handleMore(command, row)"><el-button text :icon="MoreFilled" :aria-label="`${row.title} 更多操作`" /></el-dropdown></div></template></el-table-column>
      </el-table>
      <template #footer><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total, prev, pager, next" background @current-change="load" /></template>
    </AdminTablePanel>

    <footer v-if="gridView" class="collection-footer"><span>共 {{ total }} 份考试</span><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="prev, pager, next" background @current-change="load" /></footer>
  </AdminPageShell>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { Document, Grid, List, MoreFilled, Plus, RefreshRight, Search } from '@element-plus/icons-vue'
import { AdminPageShell, AdminSearchBar, AdminTablePanel } from '../../components/admin-ui'
import { adminApi } from '../../api'

const router = useRouter()
const keyword = ref('')
const category = ref('')
const statusFilter = ref<number | null>(null)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)
const gridView = ref(true)
const statusTabs = [{ value: null, label: '全部考试' }, { value: 1, label: '已发布' }, { value: 0, label: '已停用' }]

function statusCount(value: number | null) { return value === null ? total.value : list.value.filter(item => item.status === value).length }
function visText(value: number) { return { 0: '公开', 1: '登录', 2: '部门' }[value] || '-' }
function visType(value: number) { return ({ 0: 'success', 1: 'primary', 2: 'warning' } as any)[value] || '' }
function fmtDate(value: number) { return value ? new Date(value).toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }) : '-' }
function maxSubmitMinutes(row: any) {
  try {
    const settings = typeof row.settings === 'string' ? JSON.parse(row.settings) : (row.settings || {})
    return settings.maxSubmitMinutes || '-'
  } catch { return '-' }
}
async function load() {
  loading.value = true
  try {
    const response: any = await adminApi.examList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value.trim(), category: category.value.trim(), status: statusFilter.value === null ? -1 : statusFilter.value })
    list.value = response.data?.list || []
    total.value = response.data?.total || 0
  } finally { loading.value = false }
}
function search() { page.value = 1; void load() }
function resetSearch() { keyword.value = ''; category.value = ''; statusFilter.value = null; page.value = 1; void load() }
function selectStatus(value: number | null) { statusFilter.value = value; page.value = 1; void load() }
async function toggleStatus(row: any) {
  const next = row.status === 1 ? 0 : 1
  await adminApi.examStatus({ id: row.id, status: next })
  row.status = next
  ElMessage.success(next === 1 ? '已发布' : '已停用')
}
async function handleMore(command: string, row: any) {
  if (command === 'edit') return goEdit(row)
  if (command === 'enable' || command === 'disable') { await toggleStatus(row); return void load() }
  if (command === 'del') {
    try { await ElMessageBox.confirm(`确认删除「${row.title}」?`, '删除考试', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }) } catch { return }
    await adminApi.examDelete({ id: row.id })
    ElMessage.success('已删除')
    void load()
  }
}
function goEdit(row: any) { router.push({ path: '/exam/designer', query: { id: String(row.id) } }) }
function goDesigner(row: any) { goEdit(row) }
function goCreate() { router.push('/exam/designer') }
onMounted(load)
</script>

<style scoped>
.exam-page { --survey-border:var(--admin-ui-color-border); --survey-text:var(--admin-ui-color-text); --survey-muted:var(--admin-ui-color-text-secondary); }
.exam-card__preview { display:flex; flex-direction:column; gap:8px; margin-top:10px; padding-top:10px; border-top:1px solid #eef0f3; }
.exam-card__preview > div { display:flex; align-items:center; justify-content:space-between; gap:10px; color:var(--survey-muted); font-size:11px; }
.exam-card__preview strong { color:var(--survey-text); font-size:12px; font-weight:600; }
</style>
