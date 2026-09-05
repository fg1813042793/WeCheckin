<script lang="ts">
export type UserFeedbackFilterStatus = '' | 'pending' | 'processing' | 'resolved' | 'closed'

export interface UserFeedbackFilterState {
  keyword: string
  status: UserFeedbackFilterStatus
  handlerId?: number
  submittedDates: [string, string] | null
}

function localDateStart(value: string): number | undefined {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) return undefined
  const timestamp = new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3])).getTime()
  return Number.isFinite(timestamp) ? timestamp : undefined
}

export function buildUserFeedbackQuery(filters: UserFeedbackFilterState, page: number, pageSize: number) {
  const keyword = filters.keyword.trim()
  const handlerId = Number(filters.handlerId)
  const submittedFrom = filters.submittedDates ? localDateStart(filters.submittedDates[0]) : undefined
  const submittedStartTo = filters.submittedDates ? localDateStart(filters.submittedDates[1]) : undefined
  return {
    ...(keyword ? { keyword } : {}),
    ...(filters.status ? { status: filters.status } : {}),
    ...(Number.isInteger(handlerId) && handlerId > 0 ? { handlerId } : {}),
    ...(submittedFrom !== undefined ? { submittedFrom } : {}),
    ...(submittedStartTo !== undefined ? { submittedTo: submittedStartTo + 86_400_000 - 1 } : {}),
    page,
    pageSize,
  }
}

export function buildUserFeedbackOverviewQuery(filters: UserFeedbackFilterState) {
  const { status: _status, page: _page, pageSize: _pageSize, ...query } = buildUserFeedbackQuery(filters, 1, 1)
  return query
}
</script>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture, Refresh, View } from '@element-plus/icons-vue'
import { AdminPageHeader, AdminPageShell, AdminSearchBar, AdminTablePanel } from '@/components/admin-ui'
import { adminApi } from '@/api'
import type { UserFeedbackDetail, UserFeedbackOverview as OverviewData, UserFeedbackStatus, UserFeedbackSummary } from '@/types/userFeedback'
import { showRequestError } from '@/utils/request'
import { canHandleUserFeedback, userFeedbackStatusMeta } from './userFeedbackStatus'
import UserFeedbackOverview from './components/UserFeedbackOverview.vue'
import UserFeedbackDetailDrawer from './components/UserFeedbackDetailDrawer.vue'
import UserFeedbackStatusDialog from './components/UserFeedbackStatusDialog.vue'

interface DetailDrawerExposed {
  refresh: () => Promise<void>
}

interface StatusDialogExposed {
  open: (detail: UserFeedbackDetail) => void
}

const filters = reactive<UserFeedbackFilterState>({
  keyword: '',
  status: '',
  handlerId: undefined,
  submittedDates: null,
})
const overview = ref<OverviewData | null>(null)
const overviewLoading = ref(false)
const overviewError = ref('')
const rows = ref<UserFeedbackSummary[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const listLoading = ref(false)
const listError = ref('')
const detailVisible = ref(false)
const selectedFeedbackId = ref<number | null>(null)
const detailDrawerRef = ref<DetailDrawerExposed>()
const statusDialogRef = ref<StatusDialogExposed>()
let overviewRequestSequence = 0
let listRequestSequence = 0

const canHandle = computed(() => canHandleUserFeedback())
const tableEmpty = computed(() => !listLoading.value && !listError.value && rows.value.length === 0)

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function personLabel(name: string | undefined, id: number | undefined, fallback: string) {
  if (name) return name
  if (id) return `#${id}`
  return fallback
}

async function loadOverview() {
  const sequence = ++overviewRequestSequence
  overviewLoading.value = true
  overviewError.value = ''
  try {
    const response = await adminApi.userFeedbackOverview(buildUserFeedbackOverviewQuery(filters))
    if (sequence === overviewRequestSequence) overview.value = response.data
  } catch (error) {
    if (sequence !== overviewRequestSequence) return
    overviewError.value = '反馈统计加载失败，请重试'
    showRequestError(error, '反馈统计加载失败')
  } finally {
    if (sequence === overviewRequestSequence) overviewLoading.value = false
  }
}

async function loadList() {
  const sequence = ++listRequestSequence
  listLoading.value = true
  listError.value = ''
  try {
    const response = await adminApi.userFeedbackList(buildUserFeedbackQuery(filters, page.value, pageSize.value))
    if (sequence !== listRequestSequence) return
    rows.value = response.data.list
    total.value = response.data.total
    page.value = response.data.page
    pageSize.value = response.data.pageSize
  } catch (error) {
    if (sequence !== listRequestSequence) return
    listError.value = '反馈列表加载失败，请重试'
    showRequestError(error, '反馈列表加载失败')
  } finally {
    if (sequence === listRequestSequence) listLoading.value = false
  }
}

async function refreshOverviewAndList() {
  await Promise.allSettled([loadOverview(), loadList()])
}

function search() {
  page.value = 1
  void refreshOverviewAndList()
}

function resetFilters() {
  filters.keyword = ''
  filters.status = ''
  filters.handlerId = undefined
  filters.submittedDates = null
  search()
}

function handleOverviewSelect(status: UserFeedbackStatus) {
  filters.status = status
  page.value = 1
  void refreshOverviewAndList()
}

function handlePageSizeChange() {
  page.value = 1
  void loadList()
}

function openDetail(row: UserFeedbackSummary) {
  selectedFeedbackId.value = row.id
  detailVisible.value = true
}

function openStatusDialog(detail: UserFeedbackDetail) {
  if (!canHandleUserFeedback()) {
    ElMessage.warning('当前账号没有反馈处理权限')
    return
  }
  statusDialogRef.value?.open(detail)
}

async function handleStatusSuccess(_detail: UserFeedbackDetail) {
  await Promise.allSettled([
    loadOverview(),
    loadList(),
    detailDrawerRef.value?.refresh() ?? Promise.resolve(),
  ])
}

function handleVisibilityChange() {
  if (document.visibilityState === 'visible') void refreshOverviewAndList()
}

function mountPage() {
  document.addEventListener('visibilitychange', handleVisibilityChange)
  void refreshOverviewAndList()
}

function unmountPage() {
  document.removeEventListener('visibilitychange', handleVisibilityChange)
}

onMounted(mountPage)
onBeforeUnmount(unmountPage)
</script>

<template>
  <AdminPageShell width="wide" density="compact">
    <AdminPageHeader title="用户反馈" description="查看用户提交内容并跟踪处理进度">
      <template #actions>
        <el-button
          circle
          :icon="Refresh"
          :loading="overviewLoading || listLoading"
          title="刷新"
          aria-label="刷新"
          @click="refreshOverviewAndList"
        />
      </template>
    </AdminPageHeader>

    <UserFeedbackOverview
      :overview="overview"
      :active-status="filters.status"
      :loading="overviewLoading"
      :error="overviewError"
      @select="handleOverviewSelect"
      @retry="loadOverview"
    />

    <AdminSearchBar collapsible :loading="listLoading" @search="search" @reset="resetFilters">
      <el-form-item label="编号 / 用户 / 消息">
        <el-input v-model="filters.keyword" clearable placeholder="输入反馈编号、用户名或消息关键词" />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="filters.status" clearable placeholder="全部状态">
          <el-option v-for="status in ['pending', 'processing', 'resolved', 'closed'] as UserFeedbackStatus[]" :key="status" :label="userFeedbackStatusMeta(status).label" :value="status" />
        </el-select>
      </el-form-item>
      <el-form-item label="处理人 ID">
        <el-input-number v-model="filters.handlerId" :min="1" :step="1" :precision="0" controls-position="right" placeholder="全部处理人" />
      </el-form-item>
      <el-form-item label="提交时间">
        <el-date-picker
          v-model="filters.submittedDates"
          type="daterange"
          value-format="YYYY-MM-DD"
          unlink-panels
          clearable
          start-placeholder="开始日期"
          end-placeholder="结束日期"
        />
      </el-form-item>
    </AdminSearchBar>

    <AdminTablePanel
      title="反馈列表"
      :count="total"
      :loading="listLoading"
      :empty="tableEmpty"
      empty-text="暂无符合条件的反馈"
      :error="listError"
      min-height="360px"
    >
      <template #actions>
        <el-button :icon="Refresh" :loading="listLoading" @click="loadList">重试 / 刷新</el-button>
      </template>
      <el-table :data="rows" row-key="id" table-layout="fixed">
        <el-table-column prop="feedbackNo" label="反馈编号" width="174" show-overflow-tooltip />
        <el-table-column label="内容摘要 / 首图" min-width="260">
          <template #default="{ row }">
            <div class="feedback-summary-cell">
              <div v-if="row.imageCount > 0" class="feedback-summary-cell__image" :title="`包含 ${row.imageCount} 张图片`">
                <el-icon><Picture /></el-icon>
              </div>
              <el-tooltip :content="row.summary || '暂无文字内容'" placement="top" :show-after="400">
                <span>{{ row.summary || '暂无文字内容' }}</span>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="提交人" width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ personLabel(row.submitterName, row.submitterId, '未知用户') }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="userFeedbackStatusMeta(row.status).type">{{ userFeedbackStatusMeta(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="处理人" width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ personLabel(row.handlerName, row.handlerId, '未分配') }}</template>
        </el-table-column>
        <el-table-column prop="imageCount" label="图片数" width="82" align="center" />
        <el-table-column label="最近活动" width="180">
          <template #default="{ row }">{{ formatTime(row.lastActivityAt) }}</template>
        </el-table-column>
        <el-table-column label="提交时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="92" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" :icon="View" @click="openDetail(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <span />
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadList"
          @size-change="handlePageSizeChange"
        />
      </template>
    </AdminTablePanel>

    <UserFeedbackDetailDrawer
      ref="detailDrawerRef"
      v-model="detailVisible"
      :feedback-id="selectedFeedbackId"
      :can-handle="canHandle"
      :format-time="formatTime"
      @request-status-update="openStatusDialog"
    />
    <UserFeedbackStatusDialog ref="statusDialogRef" @success="handleStatusSuccess" />
  </AdminPageShell>
</template>

<style scoped>
.feedback-summary-cell {
  display: flex;
  align-items: center;
  gap: var(--admin-ui-space-3);
  min-width: 0;
}

.feedback-summary-cell > span {
  overflow: hidden;
  min-width: 0;
  color: var(--admin-ui-color-text-secondary);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feedback-summary-cell__image {
  display: grid;
  flex: 0 0 36px;
  width: 36px;
  height: 36px;
  place-items: center;
  border: 1px solid var(--admin-ui-color-border);
  border-radius: var(--admin-ui-radius-sm);
  background: var(--admin-ui-color-surface-muted);
  color: var(--admin-ui-color-text-muted);
}

@media (max-width: 767px) {
  :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: flex-end;
  }
}
</style>
