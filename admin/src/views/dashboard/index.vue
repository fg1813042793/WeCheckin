<template>
  <div v-loading="loading" class="admin-page dashboard-page">
    <section class="dashboard-toolbar">
      <div>
        <p class="dashboard-kicker">WeCheckin 管理后台</p>
        <h2>运营控制台</h2>
        <p>聚合当前权限范围内的业务趋势、待办事项和运行异常。</p>
      </div>
      <div class="dashboard-toolbar__actions">
        <span class="updated-at">更新于 {{ updatedAtText }}</span>
        <el-radio-group v-model="trendDays" size="small" @change="loadData">
          <el-radio-button :value="7">近 7 天</el-radio-button>
          <el-radio-button :value="30">近 30 天</el-radio-button>
        </el-radio-group>
        <el-button icon="Refresh" circle title="刷新" :loading="loading" @click="loadData" />
      </div>
    </section>

    <el-alert
      v-if="loadError"
      :title="loadError"
      type="error"
      show-icon
      :closable="false"
    >
      <template #default>
        <el-button link type="primary" @click="loadData">重新加载</el-button>
      </template>
    </el-alert>

    <section class="metric-grid" aria-label="核心指标">
      <el-card v-for="item in metrics" :key="item.label" class="metric-card admin-card" shadow="never" @click="goRoute(item.path)">
        <div class="metric-icon" :class="item.tone">
          <el-icon><component :is="item.icon" /></el-icon>
        </div>
        <div class="metric-content">
          <div class="metric-value">{{ item.value }}</div>
          <div class="metric-label">{{ item.label }}</div>
          <div class="metric-meta">{{ item.meta }}</div>
        </div>
        <el-icon class="metric-arrow"><ArrowRight /></el-icon>
      </el-card>
    </section>

    <section class="dashboard-main-grid">
      <el-card class="admin-card dashboard-trend" shadow="never">
        <template #header>
          <div class="card-header">
            <div>
              <strong>业务趋势</strong>
              <span>按天统计新增与提交量</span>
            </div>
            <el-tag size="small" type="info" effect="plain">{{ trendDays }} 天</el-tag>
          </div>
        </template>
        <DashboardTrendChart
          v-if="trendSeries.length && overview.trend.length"
          class="trend-chart"
          :option="trendOption"
          :description="`近 ${trendDays} 天业务趋势图`"
        />
        <el-empty v-else :image-size="64" description="当前权限下暂无趋势数据" />
        <div v-if="latestTrend" class="trend-summary" aria-label="最近一天趋势摘要">
          <span v-for="item in latestTrendSummary" :key="item.label">
            {{ item.label }} <strong>{{ item.value }}</strong>
          </span>
        </div>
      </el-card>

      <el-card class="admin-card dashboard-todos" shadow="never">
        <template #header>
          <div class="card-header">
            <div>
              <strong>待办与异常</strong>
              <span>优先处理需要人工介入的事项</span>
            </div>
            <el-tag size="small" :type="todoTotal > 0 ? 'warning' : 'success'" effect="plain">
              {{ todoTotal > 0 ? `${todoTotal} 项` : '已清零' }}
            </el-tag>
          </div>
        </template>
        <div v-if="todoItems.length" class="todo-list">
          <button v-for="item in todoItems" :key="item.label" class="todo-item" type="button" @click="goRoute(item.path)">
            <span class="todo-icon" :class="item.tone"><el-icon><component :is="item.icon" /></el-icon></span>
            <span class="todo-copy">
              <strong>{{ item.label }}</strong>
              <em>{{ item.desc }}</em>
            </span>
            <span class="todo-count" :class="{ 'has-value': item.count > 0 }">{{ item.count }}</span>
            <el-icon><ArrowRight /></el-icon>
          </button>
        </div>
        <el-empty v-else :image-size="64" description="当前角色没有可展示的待办模块" />
      </el-card>
    </section>

    <section class="dashboard-bottom-grid">
      <el-card class="admin-card dashboard-health" shadow="never">
        <template #header>
          <div class="card-header">
            <div>
              <strong>运行状态</strong>
              <span>状态来自真实异常计数，不使用静态健康结论</span>
            </div>
          </div>
        </template>
        <div v-if="healthItems.length" class="health-grid">
          <button v-for="item in healthItems" :key="item.label" class="health-item" type="button" @click="goRoute(item.path)">
            <span class="health-indicator" :class="item.tone" />
            <span>
              <strong>{{ item.label }}</strong>
              <em>{{ item.desc }}</em>
            </span>
            <el-tag size="small" :type="item.tagType" effect="light">{{ item.status }}</el-tag>
          </button>
        </div>
        <el-empty v-else :image-size="64" description="当前角色没有可展示的运行模块" />
      </el-card>

      <el-card class="admin-card quick-entry-panel" shadow="never">
        <template #header>
          <div class="card-header">
            <div>
              <strong>常用入口</strong>
              <span>仅展示当前角色有权访问的模块</span>
            </div>
          </div>
        </template>
        <div class="quick-actions">
          <el-button v-for="item in quickActions" :key="item.path" plain @click="goRoute(item.path)">
            <el-icon><component :is="item.icon" /></el-icon>
            {{ item.label }}
          </el-button>
        </div>
      </el-card>
    </section>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminApi } from '../../api'
import type { AdminHomeData, AdminHomeTrendPoint } from '../../api/types'
import { hasPerm } from '../../utils/permission'
import DashboardTrendChart from './DashboardTrendChart.vue'

type TrendDays = 7 | 30
type Tone = 'blue' | 'green' | 'orange' | 'purple' | 'red' | 'cyan'

const emptyOverview: AdminHomeData = {
  userCnt: 0,
  enrollCnt: 0,
  newsCnt: 0,
  joinCnt: 0,
  eventCnt: 0,
  eventUserCnt: 0,
  mgrCnt: 0,
  surveyCnt: 0,
  todaySurveyResponseCnt: 0,
  examCnt: 0,
  todayExamRecordCnt: 0,
  pendingUserCnt: 0,
  pendingWorkflowTaskCnt: 0,
  failedWorkflowNotificationCnt: 0,
  failedScheduledRunCnt: 0,
  pendingFeedbackCnt: 0,
  unboundDingTalkUserCnt: 0,
  unreadNotificationCnt: 0,
  trendDays: 7,
  trend: [],
  updatedAt: 0,
  capabilities: {},
}

const router = useRouter()
const overview = ref<AdminHomeData>({ ...emptyOverview })
const loading = ref(false)
const loadError = ref('')
const trendDays = ref<TrendDays>(7)

function canAccess(capability: string, ...permissions: string[]) {
  return overview.value.capabilities[capability] !== false && permissions.some(permission => hasPerm(permission))
}

const metrics = computed(() => [
  { label: '可见用户', value: overview.value.userCnt, meta: `${overview.value.pendingUserCnt} 人待审核`, path: '/user', icon: 'User', tone: 'blue' as Tone, visible: canAccess('users', 'admin:menu:user') },
  { label: '今日打卡', value: overview.value.joinCnt, meta: `${overview.value.enrollCnt} 个项目`, path: '/enroll', icon: 'List', tone: 'green' as Tone, visible: canAccess('enrollments', 'admin:menu:enroll') },
  { label: '问卷总数', value: overview.value.surveyCnt, meta: `今日 ${overview.value.todaySurveyResponseCnt} 份答卷`, path: '/survey', icon: 'DocumentChecked', tone: 'orange' as Tone, visible: canAccess('surveys', 'admin:menu:survey:list') },
  { label: '考试总数', value: overview.value.examCnt, meta: `今日 ${overview.value.todayExamRecordCnt} 次提交`, path: '/exam/list', icon: 'EditPen', tone: 'purple' as Tone, visible: canAccess('exams', 'admin:menu:exam', 'admin:menu:exam:list-page') },
  { label: '我的流程待办', value: overview.value.pendingWorkflowTaskCnt, meta: '仅统计当前管理员可执行任务', path: '/workflow/tasks', icon: 'Share', tone: 'red' as Tone, visible: canAccess('workflowTasks', 'admin:menu:workflow:tasks') },
  { label: '未读通知', value: overview.value.unreadNotificationCnt, meta: '当前管理员站内通知', path: '/notifications', icon: 'Bell', tone: 'cyan' as Tone, visible: canAccess('notifications', 'admin:menu:notification') },
].filter(item => item.visible))

const todoItems = computed(() => [
  { label: '待审核用户', desc: '处理新用户审核和账号状态', count: overview.value.pendingUserCnt, path: '/user', icon: 'User', tone: 'warning', visible: canAccess('users', 'admin:menu:user') },
  { label: '流程任务', desc: '处理分配给当前管理员的任务', count: overview.value.pendingWorkflowTaskCnt, path: '/workflow/tasks', icon: 'Share', tone: 'primary', visible: canAccess('workflowTasks', 'admin:menu:workflow:tasks') },
  { label: '流程通知异常', desc: '检查失败或终止的流程通知', count: overview.value.failedWorkflowNotificationCnt, path: '/workflow/instances', icon: 'Warning', tone: 'danger', visible: canAccess('workflowNotifications', 'admin:menu:workflow:instances') },
  { label: '定时任务失败', desc: '查看最近 24 小时失败运行', count: overview.value.failedScheduledRunCnt, path: '/scheduled-task/runs', icon: 'Timer', tone: 'danger', visible: canAccess('scheduledRuns', 'admin:menu:scheduled-task:runs') },
  { label: '待处理反馈', desc: '跟进待受理和处理中的反馈', count: overview.value.pendingFeedbackCnt, path: '/user-feedbacks', icon: 'ChatDotRound', tone: 'warning', visible: canAccess('feedback', 'admin:menu:user-feedback') },
  { label: '钉钉未绑定用户', desc: '补充本地用户与钉钉账号绑定', count: overview.value.unboundDingTalkUserCnt, path: '/dingtalk/bindings', icon: 'Connection', tone: 'warning', visible: canAccess('dingtalkBindings', 'admin:menu:dingtalk:bindings') },
].filter(item => item.visible))

const todoTotal = computed(() => todoItems.value.reduce((total, item) => total + item.count, 0))

const healthItems = computed(() => [
  statusItem('流程通知', overview.value.failedWorkflowNotificationCnt, '通知投递正常', '存在失败或终止记录', '/workflow/instances', canAccess('workflowNotifications', 'admin:menu:workflow:instances')),
  statusItem('定时任务', overview.value.failedScheduledRunCnt, '最近 24 小时无失败', '最近 24 小时存在失败', '/scheduled-task/runs', canAccess('scheduledRuns', 'admin:menu:scheduled-task:runs')),
  statusItem('钉钉绑定', overview.value.unboundDingTalkUserCnt, '可见用户均已绑定', '存在未绑定用户', '/dingtalk/bindings', canAccess('dingtalkBindings', 'admin:menu:dingtalk:bindings')),
  statusItem('站内通知', overview.value.unreadNotificationCnt, '没有未读通知', '存在未读通知', '/notifications', canAccess('notifications', 'admin:menu:notification')),
].filter(item => item.visible))

const quickActions = computed(() => [
  { label: '用户管理', path: '/user', icon: 'User', visible: canAccess('users', 'admin:menu:user') },
  { label: '问卷管理', path: '/survey', icon: 'List', visible: canAccess('surveys', 'admin:menu:survey:list') },
  { label: '流程任务', path: '/workflow/tasks', icon: 'Share', visible: canAccess('workflowTasks', 'admin:menu:workflow:tasks') },
  { label: '定时任务', path: '/scheduled-task/runs', icon: 'Timer', visible: canAccess('scheduledRuns', 'admin:menu:scheduled-task:runs') },
  { label: '用户反馈', path: '/user-feedbacks', icon: 'ChatDotRound', visible: canAccess('feedback', 'admin:menu:user-feedback') },
  { label: '系统配置', path: '/setup', icon: 'Setting', visible: canAccess('setup', 'admin:menu:setup') },
].filter(item => item.visible))

const trendSeriesDefs = computed(() => [
  { name: '新增用户', key: 'userCount' as const, color: '#2563eb', visible: canAccess('users', 'admin:menu:user') },
  { name: '打卡参与', key: 'enrollJoinCount' as const, color: '#059669', visible: canAccess('enrollments', 'admin:menu:enroll') },
  { name: '问卷答卷', key: 'surveyResponseCount' as const, color: '#d97706', visible: canAccess('surveyResponses', 'admin:menu:survey:responses') },
  { name: '考试提交', key: 'examRecordCount' as const, color: '#7c3aed', visible: canAccess('exams', 'admin:menu:exam', 'admin:menu:exam:list-page') },
].filter(item => item.visible))

const trendSeries = computed(() => trendSeriesDefs.value.map(item => ({
  name: item.name,
  type: 'line',
  smooth: true,
  symbolSize: 7,
  data: overview.value.trend.map(point => point[item.key]),
  lineStyle: { width: 2, color: item.color },
  itemStyle: { color: item.color },
  emphasis: { focus: 'series' },
})))

const trendOption = computed(() => ({
  color: trendSeriesDefs.value.map(item => item.color),
  tooltip: { trigger: 'axis' },
  legend: { top: 0, left: 0, data: trendSeriesDefs.value.map(item => item.name) },
  grid: { left: 42, right: 18, top: 48, bottom: 32 },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: overview.value.trend.map(point => point.date.slice(5)),
    axisTick: { show: false },
  },
  yAxis: { type: 'value', minInterval: 1, splitLine: { lineStyle: { color: '#eef2f7' } } },
  series: trendSeries.value,
}))

const latestTrend = computed<AdminHomeTrendPoint | undefined>(() => overview.value.trend[overview.value.trend.length - 1])
const latestTrendSummary = computed(() => trendSeriesDefs.value.map(item => ({
  label: item.name,
  value: latestTrend.value?.[item.key] || 0,
})))

const updatedAtText = computed(() => {
  if (!overview.value.updatedAt) return '尚未更新'
  return new Date(overview.value.updatedAt).toLocaleString('zh-CN', { hour12: false })
})

function statusItem(label: string, count: number, normalText: string, warningText: string, path: string, visible: boolean) {
  return {
    label,
    count,
    desc: count > 0 ? warningText : normalText,
    status: count > 0 ? `${count} 项异常` : '正常',
    tone: count > 0 ? 'warning' : 'success',
    tagType: count > 0 ? 'warning' as const : 'success' as const,
    path,
    visible,
  }
}

function goRoute(path: string) {
  void router.push(path)
}

async function loadData() {
  if (loading.value) return
  loading.value = true
  loadError.value = ''
  try {
    const response = await adminApi.home({ days: trendDays.value })
    overview.value = { ...emptyOverview, ...(response.data || {}) }
  } catch {
    loadError.value = '控制台数据加载失败，请检查网络或稍后重试。'
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.dashboard-page {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
  box-sizing: border-box;
}

.dashboard-page .admin-card {
  border: 1px solid var(--admin-border, #e5e7eb);
  border-radius: 8px;
  box-shadow: none;
}

.dashboard-toolbar {
  min-height: 72px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 4px 0 12px;
  border-bottom: 1px solid #e5e7eb;
}

.dashboard-kicker {
  margin: 0 0 4px;
  color: #2563eb;
  font-size: 12px;
  font-weight: 600;
}

.dashboard-toolbar h2 {
  margin: 0;
  color: #111827;
  font-size: 22px;
  line-height: 1.35;
}

.dashboard-toolbar p:not(.dashboard-kicker) {
  margin: 5px 0 0;
  color: #64748b;
  font-size: 13px;
}

.dashboard-toolbar__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  flex-wrap: wrap;
}

.updated-at {
  color: #64748b;
  font-size: 12px;
  white-space: nowrap;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
  gap: 12px;
}

.metric-card {
  cursor: pointer;
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.metric-card:hover {
  border-color: #93c5fd;
  box-shadow: 0 4px 14px rgb(15 23 42 / 7%);
}

.metric-card :deep(.el-card__body) {
  min-height: 92px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  box-sizing: border-box;
}

.metric-icon,
.todo-icon {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.metric-icon {
  width: 42px;
  height: 42px;
  font-size: 20px;
}

.metric-icon.blue { color: #2563eb; background: #eff6ff; }
.metric-icon.green { color: #059669; background: #ecfdf5; }
.metric-icon.orange { color: #d97706; background: #fff7ed; }
.metric-icon.purple { color: #7c3aed; background: #f5f3ff; }
.metric-icon.red { color: #dc2626; background: #fef2f2; }
.metric-icon.cyan { color: #0891b2; background: #ecfeff; }

.metric-content { min-width: 0; }
.metric-value { color: #111827; font-size: 26px; line-height: 1; font-weight: 700; }
.metric-label { margin-top: 5px; color: #334155; font-size: 13px; font-weight: 600; }
.metric-meta { margin-top: 3px; overflow: hidden; color: #94a3b8; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.metric-arrow { margin-left: auto; color: #cbd5e1; }

.dashboard-main-grid,
.dashboard-bottom-grid {
  display: grid;
  gap: 16px;
}

.dashboard-main-grid { grid-template-columns: minmax(0, 1.6fr) minmax(330px, 0.9fr); }
.dashboard-bottom-grid { grid-template-columns: minmax(0, 1.4fr) minmax(300px, 0.6fr); }

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.card-header strong { display: block; color: #1e293b; font-size: 14px; }
.card-header > div > span { display: block; margin-top: 3px; color: #94a3b8; font-size: 12px; font-weight: 400; }

.trend-chart { width: 100%; height: 300px; }
.trend-summary { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; padding-top: 10px; border-top: 1px solid #f1f5f9; color: #64748b; font-size: 12px; }
.trend-summary strong { margin-left: 3px; color: #1e293b; }

.todo-list { display: flex; flex-direction: column; gap: 8px; }
.todo-item,
.health-item {
  width: 100%;
  border: 1px solid #edf1f5;
  background: #fff;
  color: inherit;
  text-align: left;
  cursor: pointer;
  box-sizing: border-box;
}

.todo-item {
  min-height: 58px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: 6px;
}

.todo-item:hover,
.health-item:hover { border-color: #bfdbfe; background: #f8fbff; }
.todo-icon { width: 32px; height: 32px; }
.todo-icon.warning { color: #d97706; background: #fff7ed; }
.todo-icon.primary { color: #2563eb; background: #eff6ff; }
.todo-icon.danger { color: #dc2626; background: #fef2f2; }
.todo-copy { min-width: 0; flex: 1; }
.todo-copy strong,
.health-item strong { display: block; color: #1e293b; font-size: 13px; }
.todo-copy em,
.health-item em { display: block; margin-top: 3px; overflow: hidden; color: #94a3b8; font-size: 11px; font-style: normal; text-overflow: ellipsis; white-space: nowrap; }
.todo-count { min-width: 28px; color: #94a3b8; font-size: 14px; font-weight: 700; text-align: right; }
.todo-count.has-value { color: #dc2626; }
.todo-item > .el-icon { color: #cbd5e1; }

.health-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
.health-item { min-height: 68px; display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 10px; padding: 12px; border-radius: 6px; }
.health-indicator { width: 9px; height: 9px; border-radius: 50%; }
.health-indicator.success { background: #10b981; }
.health-indicator.warning { background: #f59e0b; }

.quick-actions { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
.quick-actions :deep(.el-button) { width: 100%; min-height: 38px; margin: 0; justify-content: flex-start; }

@media (max-width: 1100px) {
  .dashboard-main-grid,
  .dashboard-bottom-grid { grid-template-columns: 1fr; }
}

@media (max-width: 768px) {
  .dashboard-toolbar { align-items: flex-start; flex-direction: column; }
  .dashboard-toolbar__actions { width: 100%; justify-content: flex-start; }
  .metric-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .health-grid { grid-template-columns: 1fr; }
}

@media (max-width: 520px) {
  .metric-grid,
  .quick-actions { grid-template-columns: 1fr; }
}
</style>
