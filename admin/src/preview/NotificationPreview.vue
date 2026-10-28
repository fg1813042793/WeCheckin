<script setup lang="ts">
import { computed, ref } from 'vue'
import { Bell, Check, Clock, EditPen, Filter, MoreFilled, RefreshRight, Search, View } from '@element-plus/icons-vue'
import { AdminDialog, AdminPageHeader, AdminPageShell, AdminSearchBar, AdminTablePanel } from '../components/admin-ui'

type NoticeStatus = 'sent' | 'pending' | 'failed'
interface Notice { id: number; title: string; receiver: string; channel: string; event: string; status: NoticeStatus; attempts: number; sentAt: string; error: string }
const statuses: Record<string, string> = { sent: '已发送', pending: '待处理', failed: '发送失败' }
const notices = ref<Notice[]>([
  { id: 19, title: '问卷「DISC测试」收到新答卷', receiver: 'Foster', channel: '站内信', event: '问卷提交', status: 'sent', attempts: 0, sentAt: '2026-10-13 09:28:14', error: '-' },
  { id: 18, title: '问卷「DISC测试」收到新答卷', receiver: 'foster', channel: 'Webhook', event: '问卷提交', status: 'sent', attempts: 0, sentAt: '2026-10-13 09:28:12', error: '-' },
  { id: 17, title: '季度培训反馈提醒', receiver: '团队成员', channel: '钉钉通知', event: '定时任务', status: 'pending', attempts: 1, sentAt: '-', error: '-' },
  { id: 16, title: '流程审批结果通知', receiver: 'Nick', channel: 'Webhook', event: '审批完成', status: 'failed', attempts: 3, sentAt: '-', error: '连接目标服务器超时，请检查 Webhook 地址' },
  { id: 15, title: '新员工入职问卷邀请', receiver: 'David', channel: '站内信', event: '手动发送', status: 'sent', attempts: 0, sentAt: '2026-10-12 16:42:09', error: '-' },
  { id: 14, title: '问卷「办公环境改善建议」收到新答卷', receiver: 'Lip', channel: 'Webhook', event: '问卷提交', status: 'failed', attempts: 2, sentAt: '-', error: '返回 HTTP 500，目标服务暂时不可用' },
])
const keyword = ref('')
const channel = ref('')
const status = ref<NoticeStatus | ''>('')
const applied = ref({ keyword: '', channel: '', status: '' })
const detail = ref<Notice | null>(null)
const detailVisible = ref(false)
const filtered = computed(() => notices.value.filter(item => {
  const text = (item.title + ' ' + item.receiver + ' ' + item.event).toLowerCase()
  return (!applied.value.keyword || text.includes(applied.value.keyword.toLowerCase()))
    && (!applied.value.channel || item.channel === applied.value.channel)
    && (!applied.value.status || item.status === applied.value.status)
}))
const stat = computed(() => ({ total: notices.value.length, sent: notices.value.filter(i => i.status === 'sent').length, pending: notices.value.filter(i => i.status === 'pending').length, failed: notices.value.filter(i => i.status === 'failed').length }))
function search() { applied.value = { keyword: keyword.value.trim(), channel: channel.value, status: status.value } }
function reset() { keyword.value = ''; channel.value = ''; status.value = ''; search() }
function showDetail(item: Notice) { detail.value = item; detailVisible.value = true }
function retry(item: Notice) { item.status = 'pending'; item.attempts += 1; item.error = '-' }
</script>

<template>
  <div class="preview-app notification-preview">
    <aside class="preview-sidebar">
      <a class="preview-brand" href="/design-preview.html"><span class="brand-mark">W</span><span class="brand-copy">WeCheckin<small>管理工作台</small></span></a>
      <nav>
        <span class="nav-section-label">工作空间</span>
        <a href="/design-preview.html"><el-icon><Bell /></el-icon><span>问卷管理</span></a>
        <a class="active" href="/notification-preview.html"><el-icon><Bell /></el-icon><span>通知管理</span><span class="nav-count">{{ stat.total }}</span></a>
        <a href="/design-preview.html"><el-icon><Clock /></el-icon><span>定时任务</span></a>
        <span class="nav-section-label">系统管理</span>
        <a href="/design-preview.html"><el-icon><Filter /></el-icon><span>权限设置</span></a>
      </nav>
      <div class="sidebar-footer"><span class="avatar">A</span><span class="profile-copy">管理员<small>Admin workspace</small></span></div>
    </aside>
    <div class="preview-workspace">
      <header class="workspace-header"><div class="breadcrumb"><span>工作空间</span><el-icon>›</el-icon><strong>通知管理</strong></div><span class="preview-indicator"><i />样式预览 · 演示数据</span></header>
      <div class="workspace-tabs"><span>问卷管理</span><span class="current">通知管理</span><span>定时任务</span></div>
      <AdminPageShell as="main" width="wide" density="compact" class="notification-demo-page">
        <AdminPageHeader title="通知管理" description="查看站内信、Webhook 和钉钉通知的投递状态" :divided="false">
          <template #actions><el-button text :icon="RefreshRight" aria-label="刷新通知" title="刷新" /><el-button type="primary" :icon="EditPen">发送通知</el-button></template>
        </AdminPageHeader>
        <div class="notice-metrics"><div><span class="metric-icon"><Bell /></span><span><strong>{{ stat.total }}</strong><small>全部通知</small></span></div><div><span class="metric-icon success"><Check /></span><span><strong>{{ stat.sent }}</strong><small>已发送</small></span></div><div><span class="metric-icon warning"><Clock /></span><span><strong>{{ stat.pending }}</strong><small>待处理</small></span></div><div><span class="metric-icon danger">!</span><span><strong>{{ stat.failed }}</strong><small>发送失败</small></span></div></div>
        <AdminSearchBar title="" class="notice-search" @search="search" @reset="reset">
          <el-input v-model="keyword" :prefix-icon="Search" placeholder="搜索通知标题、接收人或事件" clearable />
          <el-select v-model="channel" placeholder="全部渠道" clearable><el-option label="站内信" value="站内信" /><el-option label="Webhook" value="Webhook" /><el-option label="钉钉通知" value="钉钉通知" /></el-select>
          <el-select v-model="status" placeholder="全部状态" clearable><el-option v-for="(label, value) in statuses" :key="value" :label="label" :value="value" /></el-select>
        </AdminSearchBar>
        <AdminTablePanel title="通知记录" :count="filtered.length" :empty="!filtered.length" empty-text="没有符合条件的通知" :min-height="360">
          <el-table :data="filtered" class="notice-table" stripe>
            <el-table-column label="通知内容" min-width="300"><template #default="{ row }"><button class="notice-title" @click="showDetail(row)">{{ row.title }}</button><small>{{ row.event }} · {{ row.receiver }}</small></template></el-table-column>
            <el-table-column prop="channel" label="渠道" width="110" />
            <el-table-column label="状态" width="120"><template #default="{ row }"><span class="notice-status" :class="row.status"><i />{{ statuses[row.status] }}</span></template></el-table-column>
            <el-table-column label="尝试" width="70" align="center"><template #default="{ row }">{{ row.attempts }}</template></el-table-column>
            <el-table-column prop="sentAt" label="发送时间" width="155" />
            <el-table-column label="操作" width="150" fixed="right"><template #default="{ row }"><div class="notice-actions"><el-button text type="primary" :icon="View" @click="showDetail(row)">查看</el-button><el-button v-if="row.status === 'failed'" text type="warning" @click="retry(row)">重试</el-button><el-dropdown trigger="click"><el-button text :icon="MoreFilled" aria-label="更多操作" /></el-dropdown></div></template></el-table-column>
          </el-table>
          <template #footer><el-pagination :total="filtered.length" :page-size="20" layout="total, prev, pager, next" background /></template>
        </AdminTablePanel>
      </AdminPageShell>
    </div>
    <AdminDialog v-model="detailVisible" title="通知详情" width="md" :show-footer="false">
      <template v-if="detail"><div class="notice-detail-head"><span class="notice-status" :class="detail.status"><i />{{ statuses[detail.status] }}</span><span>{{ detail.channel }}</span><span>尝试 {{ detail.attempts }} 次</span></div><el-descriptions :column="1" border size="small"><el-descriptions-item label="通知内容">{{ detail.title }}</el-descriptions-item><el-descriptions-item label="接收人">{{ detail.receiver }}</el-descriptions-item><el-descriptions-item label="事件">{{ detail.event }}</el-descriptions-item><el-descriptions-item label="发送时间">{{ detail.sentAt }}</el-descriptions-item><el-descriptions-item label="失败原因"><span :class="{ 'error-text': detail.error !== '-' }">{{ detail.error }}</span></el-descriptions-item></el-descriptions></template>
    </AdminDialog>
  </div>
</template>
