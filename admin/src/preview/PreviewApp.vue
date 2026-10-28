<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowRight, ArrowDown, DataAnalysis, Document, EditPen, Expand, Fold, Grid, HomeFilled, List, Lock, MoreFilled, Plus, Search, Setting, User, Connection, Timer, RefreshRight, Download } from '@element-plus/icons-vue'
import { AdminDialog, AdminPageHeader, AdminPageShell, AdminSearchBar, AdminTablePanel } from '../components/admin-ui'
import SurveyPreviewCard from './SurveyPreviewCard.vue'
import { createDemoSurveys, statusLabels, type DemoSurvey, type SurveyStatus } from './survey-demo'

const surveys = ref(createDemoSurveys())
const view = ref<'cards' | 'list'>('cards')
const collapsed = ref(false)
const keyword = ref('')
const category = ref('')
const applied = ref({ keyword: '', category: '' })
const activeStatus = ref<SurveyStatus | 'all'>('all')
const sort = ref('updated')
const page = ref(1)
const pageSize = 8
const categories = ['人才测评', '员工体验', '培训反馈']
const statusTabs: { value: SurveyStatus | 'all'; label: string }[] = [
  { value: 'all', label: '全部问卷' }, { value: 'published', label: '已发布' }, { value: 'draft', label: '草稿' }, { value: 'stopped', label: '已停用' },
]
const count = (status: SurveyStatus | 'all') => status === 'all' ? surveys.value.length : surveys.value.filter(item => item.status === status).length
const responseTotal = computed(() => surveys.value.reduce((sum, item) => sum + item.responses, 0))
const filtered = computed(() => surveys.value.filter(item =>
  (activeStatus.value === 'all' || item.status === activeStatus.value)
  && (!applied.value.category || item.category === applied.value.category)
  && `${item.title} ${item.id}`.toLowerCase().includes(applied.value.keyword.toLowerCase()),
).sort((a, b) => sort.value === 'responses' ? b.responses - a.responses : b.updated.localeCompare(a.updated) || b.id - a.id))
const visibleSurveys = computed(() => filtered.value.slice((page.value - 1) * pageSize, page.value * pageSize))
function search() { applied.value = { keyword: keyword.value.trim(), category: category.value }; page.value = 1 }
function reset() { keyword.value = ''; category.value = ''; activeStatus.value = 'all'; search() }
function selectStatus(status: SurveyStatus | 'all') { activeStatus.value = status; page.value = 1 }
function restoreDemo() { surveys.value = createDemoSurveys(); reset(); ElMessage.success('演示数据已重置') }

const editing = ref(false)
const editId = ref<number | null>(null)
const draft = ref({ title: '', category: '员工体验', description: '' })
const titleError = ref('')
function edit(survey?: DemoSurvey) {
  editId.value = survey?.id ?? null
  draft.value = { title: survey?.title ?? '', category: survey?.category ?? '员工体验', description: survey?.description ?? '' }
  titleError.value = ''; editing.value = true
}
function save() {
  const title = draft.value.title.trim()
  if (!title) { titleError.value = '请输入问卷名称'; return }
  const existing = surveys.value.find(item => item.id === editId.value)
  if (existing) Object.assign(existing, draft.value, { title })
  else surveys.value.unshift({ id: Math.max(1000, ...surveys.value.map(item => item.id)) + 1, ...draft.value, title, status: 'draft', questions: 0, responses: 0, updated: '2026-10-12', sampleQuestion: '暂无题目', sampleOptions: [], distribution: [] })
  editing.value = false; reset(); ElMessage.success('已保存到本次演示')
}
function duplicate(survey: DemoSurvey) {
  surveys.value.unshift({ ...structuredClone(survey), id: Math.max(...surveys.value.map(item => item.id)) + 1, title: `${survey.title}（副本）`, status: 'draft', responses: 0, distribution: [] })
  reset(); ElMessage.success('已复制问卷')
}
function toggle(survey: DemoSurvey) {
  survey.status = survey.status === 'published' ? 'stopped' : 'published'
  ElMessage.success(statusLabels[survey.status])
}
async function remove(survey: DemoSurvey) {
  try { await ElMessageBox.confirm(`确认删除“${survey.title}”？`, '删除问卷', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }) } catch { return }
  surveys.value = surveys.value.filter(item => item.id !== survey.id)
  page.value = Math.max(1, Math.min(page.value, Math.ceil(filtered.value.length / pageSize)))
  ElMessage.success('已删除演示问卷')
}
function more(command: string, survey: DemoSurvey) {
  if (command === 'copy') duplicate(survey)
  if (command === 'toggle') toggle(survey)
  if (command === 'remove') void remove(survey)
}

const details = ref<DemoSurvey | null>(null)
const detailVisible = ref(false)
const detailTab = ref('overview')
function showDetail(survey: DemoSurvey, tab = 'overview') { details.value = survey; detailTab.value = tab; detailVisible.value = true }
function percentage(count: number) { return details.value?.responses ? Math.round(count / details.value.responses * 1000) / 10 : 0 }
function exportCSV() {
  const quote = (value: string | number) => `"${String(value).replace(/"/g, '""')}"`
  const rows = [['名称', '分类', '状态', '答卷数', '题数'], ...filtered.value.map(item => [item.title, item.category, statusLabels[item.status], item.responses, item.questions])]
  const url = URL.createObjectURL(new Blob(['\uFEFF' + rows.map(row => row.map(quote).join(',')).join('\r\n')], { type: 'text/csv;charset=utf-8' }))
  const anchor = document.createElement('a'); anchor.href = url; anchor.download = '问卷列表-演示数据.csv'; anchor.click(); URL.revokeObjectURL(url)
}
</script>

<template>
  <div class="preview-app" :class="{ 'sidebar-collapsed': collapsed }">
    <aside class="preview-sidebar" aria-label="后台导航">
      <a class="preview-brand" href="/design-preview.html"><span class="brand-mark">W</span><span class="brand-copy">WeCheckin<small>管理工作台</small></span></a>
      <nav>
        <span class="nav-section-label">工作空间</span>
        <a href="/dashboard" target="_blank" rel="noopener" title="控制台"><el-icon><HomeFilled /></el-icon><span>控制台</span></a>
        <a href="/user" target="_blank" rel="noopener" title="用户管理"><el-icon><User /></el-icon><span>用户管理</span></a>
        <span class="nav-section-label">业务管理</span>
        <a href="/workflow/definitions" target="_blank" rel="noopener" title="流程管理"><el-icon><Connection /></el-icon><span>流程管理</span><el-icon class="nav-chevron"><ArrowRight /></el-icon></a>
        <a href="/design-preview.html" class="active" aria-current="page" title="问卷管理"><el-icon><Document /></el-icon><span>问卷管理</span><span class="nav-count">{{ surveys.length }}</span></a>
        <a href="/exam/list" target="_blank" rel="noopener" title="考试管理"><el-icon><EditPen /></el-icon><span>考试管理</span></a>
        <a href="/notifications" target="_blank" rel="noopener" title="通知管理"><el-icon><Connection /></el-icon><span>通知管理</span></a>
        <span class="nav-section-label">系统管理</span>
        <a href="/scheduled-task/tasks" target="_blank" rel="noopener" title="定时任务"><el-icon><Timer /></el-icon><span>定时任务</span></a>
        <a href="/menu" target="_blank" rel="noopener" title="权限设置"><el-icon><Lock /></el-icon><span>权限设置</span></a>
        <a href="/setup" target="_blank" rel="noopener" title="系统设置"><el-icon><Setting /></el-icon><span>系统设置</span></a>
      </nav>
      <div class="sidebar-footer"><span class="avatar">A</span><span class="profile-copy">管理员<small>Admin workspace</small></span></div>
    </aside>
    <div class="preview-workspace">
      <header class="workspace-header">
        <div class="breadcrumb"><el-button text :icon="collapsed ? Expand : Fold" :aria-label="collapsed ? '展开侧栏' : '折叠侧栏'" @click="collapsed = !collapsed" /><span>业务管理</span><el-icon><ArrowRight /></el-icon><strong>问卷管理</strong></div>
        <span class="preview-indicator"><i />样式预览 · 演示数据</span>
      </header>
      <div class="workspace-tabs"><span>控制台</span><span class="current">问卷管理</span></div>
      <AdminPageShell as="main" width="wide" class="survey-demo-page">
        <AdminPageHeader title="问卷管理" :divided="false">
          <template #meta><span>{{ surveys.length }} 份问卷</span><span class="meta-divider">/</span><span>累计回收 {{ responseTotal }} 份答卷</span></template>
          <template #actions><el-button :icon="Download" @click="exportCSV">导出列表</el-button><el-button type="primary" :icon="Plus" @click="edit()">新建问卷</el-button></template>
        </AdminPageHeader>
        <AdminSearchBar title="" class="preview-search" @search="search" @reset="reset">
          <el-input v-model="keyword" :prefix-icon="Search" placeholder="搜索问卷名称或 ID" aria-label="搜索问卷名称或 ID" clearable @clear="search" />
          <el-select v-model="category" placeholder="全部分类" aria-label="问卷分类" clearable><el-option v-for="item in categories" :key="item" :label="item" :value="item" /></el-select>
        </AdminSearchBar>
        <div class="collection-toolbar">
          <div class="status-tabs" role="tablist" aria-label="问卷状态"><button v-for="tab in statusTabs" :key="tab.value" type="button" role="tab" :aria-selected="activeStatus === tab.value" :class="{ active: activeStatus === tab.value }" @click="selectStatus(tab.value)">{{ tab.label }}<span>{{ count(tab.value) }}</span></button></div>
          <div class="collection-tools">
            <el-select v-model="sort" aria-label="问卷排序" class="sort-select" @change="page = 1"><el-option label="最近更新" value="updated" /><el-option label="答卷数最多" value="responses" /></el-select>
            <div class="view-segment" aria-label="展示方式"><el-tooltip content="卡片视图"><button type="button" :class="{ active: view === 'cards' }" :aria-pressed="view === 'cards'" aria-label="卡片视图" @click="view = 'cards'"><el-icon><Grid /></el-icon></button></el-tooltip><el-tooltip content="列表视图"><button type="button" :class="{ active: view === 'list' }" :aria-pressed="view === 'list'" aria-label="列表视图" @click="view = 'list'"><el-icon><List /></el-icon></button></el-tooltip></div>
            <el-tooltip content="重置演示数据"><el-button text :icon="RefreshRight" aria-label="重置演示数据" @click="restoreDemo" /></el-tooltip>
          </div>
        </div>
        <section v-if="view === 'cards'" class="survey-collection" aria-label="问卷卡片列表">
          <SurveyPreviewCard v-for="survey in visibleSurveys" :key="survey.id" :survey="survey" @edit="edit(survey)" @detail="showDetail(survey)" @statistics="showDetail(survey, 'statistics')" @copy="duplicate(survey)" @toggle="toggle(survey)" @remove="remove(survey)" />
          <el-empty v-if="!filtered.length" description="没有符合条件的问卷"><el-button @click="reset">清除筛选</el-button></el-empty>
        </section>
        <AdminTablePanel v-else :empty="!filtered.length" empty-text="没有符合条件的问卷" :min-height="0">
          <el-table :data="visibleSurveys" class="preview-table">
            <el-table-column label="问卷名称" min-width="240"><template #default="{ row }"><button class="table-title" @click="showDetail(row)">{{ row.title }}</button><small class="table-subtitle">#{{ row.id }} · {{ row.questions }} 道题目</small></template></el-table-column>
            <el-table-column prop="category" label="分类" width="110" />
            <el-table-column label="状态" width="110"><template #default="{ row }"><span class="status-label" :class="row.status"><i />{{ statusLabels[row.status as SurveyStatus] }}</span></template></el-table-column>
            <el-table-column prop="responses" label="答卷数" width="90" align="right" />
            <el-table-column prop="updated" label="最近更新" width="120" />
            <el-table-column label="操作" width="165" fixed="right"><template #default="{ row }"><div class="table-actions"><el-button text type="primary" @click="edit(row)">编辑</el-button><el-button text @click="showDetail(row, 'statistics')">统计</el-button><el-dropdown trigger="click" @command="(value: string) => more(value, row)"><el-button text :icon="MoreFilled" :aria-label="`${row.title} 更多操作`" /><template #dropdown><el-dropdown-menu><el-dropdown-item command="copy">复制问卷</el-dropdown-item><el-dropdown-item command="toggle">{{ row.status === 'published' ? '停用问卷' : '发布问卷' }}</el-dropdown-item><el-dropdown-item command="remove" divided>删除问卷</el-dropdown-item></el-dropdown-menu></template></el-dropdown></div></template></el-table-column>
          </el-table>
        </AdminTablePanel>
        <footer class="collection-footer"><span>共 {{ filtered.length }} 份问卷</span><el-pagination v-model:current-page="page" :page-size="pageSize" :total="filtered.length" layout="prev, pager, next" background /></footer>
        <p class="preview-footnote">WeCheckin · 管理工作台<span>界面方案预览 · 数据仅用于演示</span></p>
      </AdminPageShell>
    </div>
    <AdminDialog v-model="editing" :title="editId ? '编辑问卷' : '新建问卷'" width="sm" confirm-text="保存" @confirm="save">
      <el-form label-position="top" @submit.prevent="save">
        <el-form-item label="问卷名称" required :error="titleError"><el-input v-model="draft.title" placeholder="请输入问卷名称" maxlength="60" show-word-limit @input="titleError = ''" /></el-form-item>
        <el-form-item label="分类"><el-select v-model="draft.category" style="width:100%"><el-option v-for="item in categories" :key="item" :value="item" :label="item" /></el-select></el-form-item>
        <el-form-item label="问卷说明"><el-input v-model="draft.description" type="textarea" :rows="3" maxlength="160" show-word-limit /></el-form-item>
      </el-form>
    </AdminDialog>
    <AdminDialog v-model="detailVisible" :title="details?.title || '问卷详情'" width="md" :show-footer="false">
      <template v-if="details">
        <div class="detail-summary"><span class="status-label" :class="details.status"><i />{{ statusLabels[details.status] }}</span><span>{{ details.questions }} 道题目</span><span>{{ details.responses }} 份答卷</span></div>
        <el-tabs v-model="detailTab">
          <el-tab-pane label="问卷概览" name="overview"><p class="detail-description">{{ details.description || '暂无说明' }}</p><h3>{{ details.sampleQuestion }}</h3><div v-for="option in details.sampleOptions" :key="option" class="detail-option"><i />{{ option }}</div><p class="detail-date">更新于 {{ details.updated }} · {{ details.category }}</p></el-tab-pane>
          <el-tab-pane label="统计结果" name="statistics"><div v-if="details.distribution.length" class="stat-bars"><div v-for="item in details.distribution" :key="item.label"><div class="stat-bar-label"><span>{{ item.label }}</span><span>{{ item.count }} 份 · {{ percentage(item.count) }}%</span></div><el-progress :percentage="percentage(item.count)" :show-text="false" :stroke-width="8" /></div></div><el-empty v-else description="暂无答卷数据" :image-size="70" /></el-tab-pane>
        </el-tabs>
      </template>
    </AdminDialog>
  </div>
</template>
