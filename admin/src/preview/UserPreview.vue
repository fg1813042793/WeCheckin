<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowRight, Download, EditPen, Grid, List, MoreFilled, Plus, RefreshRight, Search, User, View } from '@element-plus/icons-vue'
import { AdminDialog, AdminPageHeader, AdminPageShell, AdminSearchBar, AdminTablePanel } from '../components/admin-ui'

type UserStatus = 'normal' | 'pending' | 'disabled'
interface DemoUser { id: number; name: string; mobile: string; status: UserStatus; department: string; position: string; manager: string; roles: string[]; loginCount: number; lastLogin: string; createdAt: string }
const statusLabels: Record<string, string> = { normal: '正常', pending: '待审核', disabled: '已禁用' }
const users = ref<DemoUser[]>([
  { id: 101, name: 'Foster', mobile: '138****2601', status: 'normal', department: '米视科技 / 研发部 / 运维组', position: '运维', manager: 'David', roles: ['钉钉H5-数据权限-员工', '系统管理员'], loginCount: 38, lastLogin: '2026-10-13 09:10', createdAt: '2026-07-28' },
  { id: 102, name: 'Lip', mobile: '138****1228', status: 'normal', department: '米视科技 / 研发部 / Java开发组', position: 'Java开发', manager: 'David', roles: ['钉钉H5-数据权限-员工'], loginCount: 16, lastLogin: '2026-10-12 18:42', createdAt: '2026-08-04' },
  { id: 103, name: 'Steven', mobile: '137****8391', status: 'normal', department: '米视科技 / M/H业务 / 综合部', position: '业务负责人', manager: '-', roles: ['HRBP专员'], loginCount: 12, lastLogin: '2026-10-12 16:18', createdAt: '2026-08-03' },
  { id: 104, name: 'Nick', mobile: '136****4412', status: 'pending', department: '米视科技 / 产品部', position: '产品经理', manager: 'Amy', roles: [], loginCount: 0, lastLogin: '-', createdAt: '2026-10-12' },
  { id: 105, name: 'Amy', mobile: '135****9082', status: 'normal', department: '米视科技 / M/H业务 / 综合部', position: '-', manager: '-', roles: ['HRBP专员'], loginCount: 21, lastLogin: '2026-10-11 14:32', createdAt: '2026-07-28' },
  { id: 106, name: 'Cherry', mobile: '139****0306', status: 'disabled', department: '米视科技 / 市场部', position: '市场专员', manager: 'Foster', roles: [], loginCount: 4, lastLogin: '2026-09-28 10:05', createdAt: '2026-07-22' },
  { id: 107, name: 'David', mobile: '158****7710', status: 'normal', department: '米视科技 / 研发部 / Java开发组', position: '研发经理', manager: '-', roles: ['超级管理员'], loginCount: 72, lastLogin: '2026-10-13 08:56', createdAt: '2026-06-19' },
  { id: 108, name: 'Monica', mobile: '151****2290', status: 'normal', department: '米视科技 / M/H业务 / 综合部', position: 'HRBP', manager: 'David', roles: ['HRBP专员'], loginCount: 25, lastLogin: '2026-10-10 11:44', createdAt: '2026-07-28' },
  { id: 109, name: 'Paul', mobile: '150****5188', status: 'normal', department: '米视科技 / 研发部 / 安卓开发组', position: 'Android开发', manager: '-', roles: ['钉钉H5-数据权限-员工'], loginCount: 19, lastLogin: '2026-10-09 15:20', createdAt: '2026-07-28' },
  { id: 110, name: 'Lucky', mobile: '159****8662', status: 'normal', department: '米视科技 / HRBP', position: 'HRBP', manager: '-', roles: [], loginCount: 9, lastLogin: '2026-10-08 12:11', createdAt: '2026-07-28' },
])
const keyword = ref('')
const department = ref('')
const status = ref<UserStatus | ''>('')
const role = ref('')
const applied = ref({ keyword: '', department: '', status: '', role: '' })
const selectedIds = ref<number[]>([])
const view = ref<'table' | 'cards'>('table')
const detail = ref<DemoUser | null>(null)
const detailVisible = ref(false)
const filteredUsers = computed(() => users.value.filter(user => {
  const text = (user.name + ' ' + user.mobile).toLowerCase()
  return (!applied.value.keyword || text.includes(applied.value.keyword.toLowerCase()))
    && (!applied.value.department || user.department.includes(applied.value.department))
    && (!applied.value.status || user.status === applied.value.status)
    && (!applied.value.role || user.roles.includes(applied.value.role))
}))
const stats = computed(() => ({ total: users.value.length, normal: users.value.filter(u => u.status === 'normal').length, pending: users.value.filter(u => u.status === 'pending').length, disabled: users.value.filter(u => u.status === 'disabled').length }))
function search() { applied.value = { keyword: keyword.value.trim(), department: department.value, status: status.value, role: role.value }; selectedIds.value = [] }
function reset() { keyword.value = ''; department.value = ''; status.value = ''; role.value = ''; search() }
function selectAll() { selectedIds.value = selectedIds.value.length === filteredUsers.value.length ? [] : filteredUsers.value.map(user => user.id) }
function toggleSelected(id: number) { selectedIds.value = selectedIds.value.includes(id) ? selectedIds.value.filter(item => item !== id) : [...selectedIds.value, id] }
function showDetail(user: DemoUser) { detail.value = user; detailVisible.value = true }
function statusClass(value: UserStatus) { return value }
function exportUsers() { window.alert('演示：已生成用户列表导出文件') }
function createUser() { window.alert('演示：打开新增用户表单') }
</script>

<template>
  <div class="preview-app user-preview">
    <aside class="preview-sidebar">
      <a class="preview-brand" href="/design-preview.html"><span class="brand-mark">W</span><span class="brand-copy">WeCheckin<small>管理工作台</small></span></a>
      <nav>
        <span class="nav-section-label">工作空间</span>
        <a href="/design-preview.html"><el-icon><Grid /></el-icon><span>控制台</span></a>
        <a class="active" href="/user-preview.html"><el-icon><User /></el-icon><span>用户管理</span><span class="nav-count">{{ stats.total }}</span></a>
        <a href="/design-preview.html"><el-icon><ArrowRight /></el-icon><span>问卷管理</span></a>
        <a href="/notification-preview.html"><el-icon><View /></el-icon><span>通知管理</span></a>
        <span class="nav-section-label">系统管理</span>
        <a href="/design-preview.html"><el-icon><EditPen /></el-icon><span>权限设置</span></a>
      </nav>
      <div class="sidebar-footer"><span class="avatar">A</span><span class="profile-copy">管理员<small>Admin workspace</small></span></div>
    </aside>
    <div class="preview-workspace">
      <header class="workspace-header"><div class="breadcrumb"><span>工作空间</span><el-icon>›</el-icon><strong>用户管理</strong></div><span class="preview-indicator"><i />样式预览 · 演示数据</span></header>
      <div class="workspace-tabs"><span>控制台</span><span>角色管理</span><span>权限设置</span></div>
      <AdminPageShell as="main" width="wide" density="compact" class="user-demo-page">
        <AdminPageHeader title="用户管理" description="管理账号、组织关系和角色权限" :divided="false">
          <template #meta><span>{{ stats.total }} 位用户</span><span class="meta-divider">/</span><span>今日活跃 {{ stats.normal - 1 }} 位</span></template>
        </AdminPageHeader>
        <AdminSearchBar title="" class="user-search" @search="search" @reset="reset">
          <el-input v-model="keyword" :prefix-icon="Search" placeholder="搜索用户名或手机号" clearable />
          <el-select v-model="department" placeholder="全部部门" clearable><el-option label="研发部" value="研发部" /><el-option label="M/H业务" value="M/H业务" /><el-option label="综合部" value="综合部" /><el-option label="产品部" value="产品部" /></el-select>
          <el-select v-model="role" placeholder="全部角色" clearable><el-option label="超级管理员" value="超级管理员" /><el-option label="HRBP专员" value="HRBP专员" /><el-option label="钉钉H5-数据权限-员工" value="钉钉H5-数据权限-员工" /></el-select>
          <el-select v-model="status" placeholder="全部状态" clearable><el-option v-for="(label, value) in { normal: '正常', pending: '待审核', disabled: '已禁用' }" :key="value" :label="label" :value="value" /></el-select>
        </AdminSearchBar>
        <div class="collection-toolbar user-action-toolbar">
          <div class="collection-tools__left"><el-button type="primary" :icon="Plus" @click="createUser">新增用户</el-button><el-button type="danger" plain :disabled="!selectedIds.length" @click="selectedIds = []">批量删除</el-button></div>
          <div class="collection-tools__right"><el-button text :icon="RefreshRight" aria-label="刷新用户" title="刷新" /><el-button text :icon="Download" @click="exportUsers" aria-label="导出用户" title="导出" /><div class="view-segment"><button type="button" :class="{ active: view === 'table' }" aria-label="表格视图" @click="view = 'table'"><List /></button><button type="button" :class="{ active: view === 'cards' }" aria-label="卡片视图" @click="view = 'cards'"><Grid /></button></div></div>
        </div>
        <AdminTablePanel title="用户列表" :count="filteredUsers.length" :empty="!filteredUsers.length" empty-text="没有符合条件的用户" :min-height="440">
          <template v-if="view === 'table'">
            <el-table :data="filteredUsers" class="user-preview-table" stripe @selection-change="selectedIds = $event.map((user: DemoUser) => user.id)">
              <el-table-column type="selection" width="48" />
              <el-table-column label="用户" min-width="190"><template #default="{ row }"><div class="user-cell"><span class="user-initial">{{ row.name.slice(0, 1).toUpperCase() }}</span><button class="user-name" @click="showDetail(row)">{{ row.name }}</button></div></template></el-table-column>
              <el-table-column prop="mobile" label="手机号" width="140" />
              <el-table-column label="状态" width="100"><template #default="{ row }"><span class="user-status" :class="statusClass(row.status)"><i />{{ statusLabels[row.status] }}</span></template></el-table-column>
              <el-table-column label="所属部门" min-width="230"><template #default="{ row }"><span class="department-pill">{{ row.department }}</span></template></el-table-column>
              <el-table-column prop="position" label="岗位" width="120" />
              <el-table-column prop="manager" label="直属上级" width="110" />
              <el-table-column label="绑定角色" min-width="190"><template #default="{ row }"><span v-if="row.roles.length" class="role-pills"><span v-for="item in row.roles.slice(0, 1)" :key="item">{{ item }}</span><b v-if="row.roles.length > 1">+{{ row.roles.length - 1 }}</b></span><span v-else class="muted">未绑定</span></template></el-table-column>
              <el-table-column prop="loginCount" label="登录次数" width="90" align="right" />
              <el-table-column prop="lastLogin" label="最后登录" width="155" />
              <el-table-column label="操作" width="105" fixed="right"><template #default="{ row }"><div class="user-actions"><el-button text type="primary" @click="showDetail(row)">详情</el-button><el-dropdown trigger="click"><el-button text :icon="MoreFilled" aria-label="更多操作" /></el-dropdown></div></template></el-table-column>
            </el-table>
          </template>
          <template v-else><div class="user-card-grid"><article v-for="user in filteredUsers" :key="user.id" class="user-card"><div class="user-card-top"><span class="user-initial user-initial-lg">{{ user.name.slice(0, 1).toUpperCase() }}</span><span class="user-status" :class="statusClass(user.status)"><i />{{ statusLabels[user.status] }}</span></div><h3>{{ user.name }}</h3><p>{{ user.department }}</p><div class="user-card-meta">{{ user.position || '未设置岗位' }} · {{ user.loginCount }} 次登录</div><footer><el-button text type="primary" @click="showDetail(user)">查看详情</el-button></footer></article></div></template>
        </AdminTablePanel>
      </AdminPageShell>
    </div>
    <AdminDialog v-model="detailVisible" title="用户详情" width="md" :show-footer="false"><template v-if="detail"><div class="user-detail-head"><span class="user-initial user-initial-lg">{{ detail.name.slice(0, 1).toUpperCase() }}</span><div><h2>{{ detail.name }}</h2><span class="user-status" :class="statusClass(detail.status)"><i />{{ statusLabels[detail.status] }}</span></div></div><el-descriptions :column="2" border size="small"><el-descriptions-item label="手机号">{{ detail.mobile }}</el-descriptions-item><el-descriptions-item label="岗位">{{ detail.position }}</el-descriptions-item><el-descriptions-item label="所属部门">{{ detail.department }}</el-descriptions-item><el-descriptions-item label="直属上级">{{ detail.manager }}</el-descriptions-item><el-descriptions-item label="注册时间">{{ detail.createdAt }}</el-descriptions-item><el-descriptions-item label="最后登录">{{ detail.lastLogin }}</el-descriptions-item><el-descriptions-item label="绑定角色" :span="2">{{ detail.roles.join('、') || '未绑定' }}</el-descriptions-item></el-descriptions></template></AdminDialog>
  </div>
</template>
