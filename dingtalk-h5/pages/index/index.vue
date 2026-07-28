<template>
  <view class="dt-page">
    <view v-if="!ready" class="loading-page">加载中...</view>

    <LoginView v-else-if="!state.user" :form="loginForm" :loading="loading" @login="login" />

    <AppShell
      v-else
      :active-view="state.view"
      :active-performance-tab="activePerformanceTab"
      :nav-items="navItems"
      :page-title="pageTitle"
      :performance-tabs="performanceTabs"
      :role-text="roleText"
      :user="state.user"
      @logout="logout"
      @switch-performance-tab="switchPerformanceTab"
      @switch-view="switchView"
    >
      <SummaryView v-if="contentView === 'summary'" />
      <OrgView v-else-if="contentView === 'org'" />
      <TemplateView v-else-if="contentView === 'template'" />
      <WorkbenchView v-else />
    </AppShell>
  </view>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import AppShell from '../../components/performance/AppShell.vue'
import LoginView from '../../components/performance/LoginView.vue'
import OrgView from '../../components/performance/OrgView.vue'
import SummaryView from '../../components/performance/SummaryView'
import TemplateView from '../../components/performance/TemplateView'
import WorkbenchView from '../../components/performance/WorkbenchView'
import { providePerformanceContext } from '../../components/performance/context'
import { dingTalkAuthApi, dingTalkPerformanceApi } from '../../services/dingtalkH5Api'
import { authToken, clearAuthToken, setAuthToken } from '../../utils/request'

const statusMeta = {
  draft: { label: '员工填写', tone: 'warning', step: 0 },
  manager_review: { label: '上级评价', tone: 'purple', step: 1 },
  hrbp_review: { label: 'HRBP评价', tone: 'blue', step: 2 },
  employee_confirm: { label: '员工确认', tone: 'warning', step: 3 },
  hr_final: { label: 'HRBP归档', tone: 'orange', step: 4 },
  completed: { label: '已完成', tone: 'green', step: 5 }
}

const roleName = {
  director: '总监',
  manager: '经理',
  supervisor: '主管',
  employee: '员工',
  hrbp: 'HRBP',
  hrbp_manager: 'HRBP主管',
  admin: 'HRBP管理员'
}

const navByRole = {
  employee: [
    ['dashboard', '工作台', 'dashboard'],
    ['performance', '绩效管理', 'performance']
  ],
  supervisor: [
    ['dashboard', '工作台', 'dashboard'],
    ['performance', '绩效管理', 'performance']
  ],
  manager: [
    ['dashboard', '工作台', 'dashboard'],
    ['performance', '绩效管理', 'performance']
  ],
  director: [
    ['dashboard', '工作台', 'dashboard'],
    ['performance', '绩效管理', 'performance']
  ],
  hrbp: [
    ['dashboard', '工作台', 'dashboard'],
    ['performance', '绩效管理', 'performance']
  ],
  hrbp_manager: [
    ['dashboard', '工作台', 'dashboard'],
    ['performance', '绩效管理', 'performance']
  ],
  admin: [
    ['dashboard', '工作台', 'dashboard'],
    ['performance', '绩效管理', 'performance']
  ]
}

const performanceTabsByRole = {
  employee: [
    ['mine', '本月绩效'],
    ['history', '历史绩效']
  ],
  supervisor: [
    ['mine', '本月绩效'],
    ['history', '历史绩效'],
    ['summary', 'HRBP汇总']
  ],
  manager: [
    ['mine', '本月绩效'],
    ['history', '历史绩效'],
    ['summary', 'HRBP汇总']
  ],
  director: [
    ['mine', '本月绩效'],
    ['history', '历史绩效'],
    ['summary', 'HRBP汇总']
  ],
  hrbp: [
    ['mine', '本月绩效'],
    ['history', '历史绩效'],
    ['hrbp', 'HRBP评价'],
    ['summary', 'HRBP汇总']
  ],
  hrbp_manager: [
    ['mine', '本月绩效'],
    ['history', '历史绩效'],
    ['hrbp', 'HRBP评价'],
    ['summary', 'HRBP汇总']
  ],
  admin: [
    ['hrbp', 'HRBP评价'],
    ['summary', 'HRBP汇总'],
    ['org', '流程执行'],
    ['template', '绩效模版']
  ]
}

const ready = ref(false)
const loading = ref(false)
const reviewTab = ref('current')
const dashboardFilter = ref('queue')
const activePerformanceTab = ref('mine')
const selectedReviewId = ref('')

const loginForm = reactive({ name: 'Nick', password: '123456' })
const newReview = reactive({ employeeId: '', period: currentMonth(), nextPeriod: nextMonth() })
const newUser = reactive({
  id: '',
  name: '',
  password: '123456',
  role: 'employee',
  position: '',
  departmentLevel1: 'M/H业务',
  departmentLevel2: '',
  departmentLevel3: '',
  managerId: '',
  hrbpId: '',
  responsibleDepartments: ''
})

const summaryFilters = reactive({
  keyword: '',
  department: '',
  period: '',
  nextPeriod: '',
  status: '',
  managerId: '',
  hrbpId: '',
  grade: ''
})

const state = reactive({
  user: null,
  menus: [],
  users: [],
  reviews: [],
  workbenchStats: [],
  template: null,
  view: 'dashboard'
})

const navItems = computed(() => {
  if (Array.isArray(state.menus) && state.menus.length > 0) {
    const items = state.menus
      .filter((item) => item.key === 'dashboard' || item.key === 'performance')
      .map((item) => ({ key: item.key, label: item.label, icon: item.icon || item.key }))
    if (items.length > 0) return items
  }
  const role = state.user?.role || 'employee'
  return [...(navByRole[role] || navByRole.employee)].map(([key, label, icon]) => ({ key, label, icon }))
})

const performanceTabs = computed(() => {
  if (Array.isArray(state.menus) && state.menus.length > 0) {
    return state.menus
      .filter((item) => String(item.key || '').startsWith('performance:'))
      .map((item) => ({ key: String(item.key).replace('performance:', ''), label: item.label }))
  }
  const role = state.user?.role || 'employee'
  return [...(performanceTabsByRole[role] || performanceTabsByRole.employee)].map(([key, label]) => ({ key, label }))
})

const contentView = computed(() => state.view === 'performance' ? activePerformanceTab.value : state.view)

const pageTitle = computed(() => {
  if (state.view === 'performance') return '绩效管理'
  return '工作台'
})

const sectionTitle = computed(() => {
  const titles = {
    dashboard: '工作台',
    mine: '本月绩效',
    history: '历史绩效',
    hrbp: 'HRBP评价',
    summary: 'HRBP汇总',
    org: '流程执行',
    template: '绩效模版'
  }
  return titles[contentView.value] || pageTitle.value
})

const selectedReview = computed(() => state.reviews.find((item) => item.id === selectedReviewId.value) || currentReviews.value[0] || null)

const currentReviews = computed(() => {
  let list = [...state.reviews]
  const view = contentView.value
  if (view === 'mine') {
    list = list.filter((item) => item.employeeId === state.user.id && item.period === currentMonth())
  } else if (view === 'history') {
    list = list.filter((item) => item.employeeId === state.user.id && item.period !== currentMonth())
  } else if (view === 'manager') {
    list = list.filter((item) => item.managerId === state.user.id && item.status === 'manager_review')
  } else if (view === 'hrbp') {
    list = list.filter((item) => canHrbpHandle(item) || canFinal(item))
  } else if (view === 'dashboard') {
    if (dashboardFilter.value === 'queue') list = queueReviews()
    if (dashboardFilter.value === 'draft') list = list.filter((item) => item.status === 'draft')
    if (dashboardFilter.value === 'reviewing') list = list.filter((item) => ['manager_review', 'hrbp_review', 'employee_confirm', 'hr_final'].includes(item.status))
    if (dashboardFilter.value === 'completed') list = list.filter((item) => item.status === 'completed')
  }
  return list
})

const summaryReviews = computed(() => {
  const keyword = summaryFilters.keyword.trim().toLowerCase()
  return state.reviews.filter((item) => {
    const employee = userName(item.employeeId)
    const grade = effectiveGrade(item)
    const haystack = [item.id, employee, item.employeeId, item.department, item.managerId, item.hrbpId, item.period, item.nextPeriod, grade].join(' ').toLowerCase()
    if (keyword && !haystack.includes(keyword)) return false
    if (summaryFilters.department && item.department !== summaryFilters.department) return false
    if (summaryFilters.period && item.period !== summaryFilters.period) return false
    if (summaryFilters.nextPeriod && item.nextPeriod !== summaryFilters.nextPeriod) return false
    if (summaryFilters.status && item.status !== summaryFilters.status) return false
    if (summaryFilters.managerId && item.managerId !== summaryFilters.managerId) return false
    if (summaryFilters.hrbpId && item.hrbpId !== summaryFilters.hrbpId) return false
    if (summaryFilters.grade && grade !== summaryFilters.grade) return false
    return true
  })
})

const statCards = computed(() => {
  const all = state.reviews.length
  const draft = state.reviews.filter((item) => item.status === 'draft').length
  const reviewing = state.reviews.filter((item) => ['manager_review', 'hrbp_review', 'employee_confirm', 'hr_final'].includes(item.status)).length
  const completed = state.reviews.filter((item) => item.status === 'completed').length
  return [
    ['queue', '我的待办', queueReviews().length],
    ['all', '全部考评单', all],
    ['draft', '员工填写', draft],
    ['reviewing', '流转中', reviewing],
    ['completed', '已完成', completed]
  ]
})

const workbenchCards = computed(() => state.workbenchStats.map((item) => [item.key, item.label, item.value]))

const departments = computed(() => unique(state.reviews.map((item) => item.department)))
const managerIds = computed(() => unique(state.reviews.map((item) => item.managerId).filter(Boolean)))
const hrbpIds = computed(() => unique(state.reviews.map((item) => item.hrbpId).filter(Boolean)))
const grades = computed(() => (state.template?.gradeLevels || []).map((item) => item.grade))
const reviewTargetUsers = computed(() => state.users.filter((item) => ['employee', 'supervisor', 'manager', 'director', 'hrbp', 'hrbp_manager', 'admin'].includes(item.role)))

function roleText(role) {
  return roleName[role] || role || ''
}

function statusText(status) {
  return statusMeta[status]?.label || status
}

function userName(id) {
  return state.users.find((item) => item.id === id)?.name || id || '无'
}

function userOptionText(user) {
  return `${user.name} · ${roleText(user.role)} · ${user.department || ''}`
}

function unique(items) {
  return [...new Set(items.filter(Boolean))]
}

function queueReviews() {
  return state.reviews.filter((item) => canSelf(item) || canManager(item) || canHrbpHandle(item) || canEmployeeConfirm(item) || canFinal(item))
}

function switchView(view) {
  state.view = view
  if (view === 'performance') {
    ensureActivePerformanceTab()
  }
  reviewTab.value = 'current'
  const first = currentReviews.value[0]
  selectedReviewId.value = first?.id || ''
  void refreshData()
}

function switchPerformanceTab(tab) {
  activePerformanceTab.value = tab
  reviewTab.value = 'current'
  selectedReviewId.value = ''
  void refreshData()
}

function ensureActivePerformanceTab() {
  const tabs = performanceTabs.value
  if (!tabs.length) {
    activePerformanceTab.value = ''
    return
  }
  if (!tabs.some((item) => item.key === activePerformanceTab.value)) {
    activePerformanceTab.value = tabs[0].key
  }
}

function selectReview(id) {
  selectedReviewId.value = id
  reviewTab.value = 'current'
}

async function login() {
  loading.value = true
  try {
    const res = await dingTalkAuthApi.login(loginForm)
    setAuthToken(res.data.token)
    state.user = res.data.userInfo
    await loadBootstrap()
    await refreshData()
  } finally {
    loading.value = false
  }
}

async function logout() {
  try {
    await dingTalkAuthApi.logout()
  } finally {
    clearAuthToken()
    state.user = null
    state.menus = []
    state.users = []
    state.reviews = []
    state.template = null
    selectedReviewId.value = ''
  }
}

async function loadBootstrap() {
  const res = await dingTalkPerformanceApi.bootstrap()
  state.user = res.data.user
  state.menus = Array.isArray(res.data.menus) ? res.data.menus : []
  const topMenus = state.menus.filter((item) => item.key === 'dashboard' || item.key === 'performance')
  if (topMenus.length > 0 && !topMenus.some((item) => item.key === state.view)) {
    state.view = topMenus[0].key
  }
  ensureActivePerformanceTab()
}

function sanitizeUsers(users) {
  return (users || []).map((item) => ({ ...item, password: '' }))
}

function ensureSelectedReview() {
  if (!newReview.employeeId) newReview.employeeId = reviewTargetUsers.value[0]?.id || ''
  if (!selectedReviewId.value || !state.reviews.some((item) => item.id === selectedReviewId.value)) {
    selectedReviewId.value = currentReviews.value[0]?.id || state.reviews[0]?.id || ''
  }
}

async function loadUsers() {
  const res = await dingTalkPerformanceApi.users()
  state.users = sanitizeUsers(res.data || [])
  ensureSelectedReview()
}

async function loadReviews(params = {}) {
  const res = await dingTalkPerformanceApi.reviews({ scope: reviewScopeForContentView(), ...params })
  state.reviews = res.data || []
  ensureSelectedReview()
}

async function loadWorkbenchStats() {
  const res = await dingTalkPerformanceApi.workbench()
  state.workbenchStats = res.data?.cards || []
  state.reviews = []
  selectedReviewId.value = ''
}

async function loadTemplate() {
  const res = await dingTalkPerformanceApi.template()
  state.template = res.data
}

async function refreshData() {
  if (!state.user) return
  if (state.view === 'dashboard') {
    await loadWorkbenchStats()
    return
  }
  ensureActivePerformanceTab()
  if (!activePerformanceTab.value) return
  if (contentView.value === 'org') {
    await loadUsers()
    return
  }
  if (contentView.value === 'template') {
    await loadTemplate()
    return
  }
  await Promise.all([loadReviews(), loadUsers(), loadTemplate()])
}

function reviewScopeForContentView() {
  if (contentView.value === 'history') return 'mine'
  if (contentView.value === 'mine') return 'mine'
  if (contentView.value === 'hrbp') return 'hrbp'
  if (contentView.value === 'summary') return 'summary'
  return contentView.value || 'mine'
}

function canSelf(review) {
  return review?.employeeId === state.user?.id && review.status === 'draft'
}

function canManager(review) {
  return review?.managerId === state.user?.id && review.status === 'manager_review'
}

function canHrbpHandle(review) {
  if (!review || review.status !== 'hrbp_review') return false
  if (review.hrbpReviewerId) return review.hrbpReviewerId === state.user?.id
  return ['admin', 'hrbp_manager'].includes(state.user?.role) || (state.user?.role === 'hrbp' && review.hrbpId === state.user?.id)
}

function canEmployeeConfirm(review) {
  return review?.employeeId === state.user?.id && review.status === 'employee_confirm'
}

function canFinal(review) {
  if (!review || !['hr_final', 'completed'].includes(review.status)) return false
  if (review.hrbpReviewerId) return review.hrbpReviewerId === state.user?.id
  return review.hrbpId === state.user?.id || ['admin', 'hrbp_manager'].includes(state.user?.role)
}

function canWithdraw(review) {
  if (!review) return false
  if (review.status === 'manager_review' && review.employeeId === state.user.id) return true
  if (review.status === 'hrbp_review' && review.managerId === state.user.id) return true
  if (review.status === 'employee_confirm' && canHrbpHandle({ ...review, status: 'hrbp_review' })) return true
  if (review.status === 'hr_final' && review.employeeId === state.user.id && !review.finalGrade) return true
  if (review.status === 'hr_final' && canFinal(review) && !review.finalGrade) return true
  return false
}

function canEditObjectiveDimension(review) {
  return canSelf(review) && !review.objectiveSourceReviewId
}

function reviewPayload(review) {
  return {
    objectives: review.objectives,
    nextObjectives: review.nextObjectives,
    values: review.values,
    selfSummary: review.selfSummary,
    managerComment: review.managerComment,
    managerGrade: review.managerGrade,
    hrbpComment: review.hrbpComment,
    hrbpGrade: review.hrbpGrade,
    employeeConfirmComment: review.employeeConfirmComment,
    finalGrade: review.finalGrade,
    finalNote: review.finalNote
  }
}

async function performReviewAction(action, successText) {
  if (!selectedReview.value) return
  const res = await dingTalkPerformanceApi.reviewAction(selectedReview.value.id, action, reviewPayload(selectedReview.value))
  updateReview(res.data)
  await loadReviews()
  toast(successText)
}

async function returnReview(action, label) {
  if (!selectedReview.value) return
  const reason = window.prompt(`${label}原因`)
  if (reason === null) return
  if (!reason.trim()) {
    toast('请填写退回原因')
    return
  }
  const res = await dingTalkPerformanceApi.reviewAction(selectedReview.value.id, action, {
    ...reviewPayload(selectedReview.value),
    returnReason: reason.trim()
  })
  updateReview(res.data)
  await loadReviews()
  toast('已退回')
}

async function withdrawReview() {
  if (!selectedReview.value || !window.confirm('确认撤销提交？撤销后会回到上一阶段，可重新编辑。')) return
  const res = await dingTalkPerformanceApi.reviewAction(selectedReview.value.id, 'withdraw', {})
  updateReview(res.data)
  await loadReviews()
  toast('已撤销')
}

function updateReview(review) {
  const index = state.reviews.findIndex((item) => item.id === review.id)
  if (index >= 0) state.reviews[index] = review
  selectedReviewId.value = review.id
}

async function createReview() {
  const res = await dingTalkPerformanceApi.createReview(newReview)
  state.view = 'performance'
  activePerformanceTab.value = 'summary'
  await loadReviews({ scope: 'summary' })
  selectedReviewId.value = res.data.id
  toast('考评单已创建')
}

async function deleteReview(id) {
  if (!window.confirm('确认删除这张考评单？删除后不能恢复。')) return
  await dingTalkPerformanceApi.deleteReview(id)
  await loadReviews()
  toast('已删除')
}

async function exportSummary() {
  window.location.href = dingTalkPerformanceApi.exportUrl({ scope: 'summary', ...summaryFilters })
}

async function createUser() {
  const payload = normalizeUserPayload(newUser)
  const res = await dingTalkPerformanceApi.createUser(payload)
  state.users = sanitizeUsers(res.data.users || [])
  Object.assign(newUser, { id: '', name: '', password: '123456', role: 'employee', position: '', departmentLevel1: 'M/H业务', departmentLevel2: '', departmentLevel3: '', managerId: '', hrbpId: '', responsibleDepartments: '' })
  ensureSelectedReview()
  toast('人员已创建')
}

async function saveUser(user) {
  const res = await dingTalkPerformanceApi.updateUser(user.id, normalizeUserPayload(user))
  state.users = sanitizeUsers(res.data.users || [])
  ensureSelectedReview()
  toast('人员已保存')
}

async function deleteUser(id) {
  if (!window.confirm(`确认删除账号 ${id}？`)) return
  const res = await dingTalkPerformanceApi.deleteUser(id)
  state.users = sanitizeUsers(res.data.users || [])
  ensureSelectedReview()
  toast('人员已删除')
}

function normalizeUserPayload(user) {
  return {
    id: user.id,
    name: user.name,
    password: user.password || '',
    role: user.role,
    position: user.position || '',
    departmentLevel1: user.departmentLevel1 || '',
    departmentLevel2: user.departmentLevel2 || '',
    departmentLevel3: user.departmentLevel3 || '',
    managerId: user.managerId || '',
    hrbpId: user.hrbpId || '',
    responsibleDepartments: Array.isArray(user.responsibleDepartments) ? user.responsibleDepartments.join(',') : (user.responsibleDepartments || '')
  }
}

function effectiveGrade(review) {
  return review.finalGrade || review.hrbpGrade || ''
}

function objectiveScore(item) {
  const completion = Number(item.completion)
  const weight = Number(item.weight)
  if (!Number.isFinite(completion) || !Number.isFinite(weight)) return 0
  return Math.round((weight * completion / 100) * 10) / 10
}

function totalObjectiveScore(review) {
  const total = (review.objectives || []).reduce((sum, item) => sum + objectiveScore(item), 0)
  return Math.round(total * 10) / 10
}

function valueTotal(review, field) {
  const nums = (review.values || []).map((item) => Number(item[field])).filter(Number.isFinite)
  if (!nums.length) return '-'
  return Math.round(nums.reduce((sum, item) => sum + item, 0) * 10) / 10
}

function gradeOptions() {
  return [''].concat(grades.value)
}

function statusClass(status) {
  return `chip chip-${statusMeta[status]?.tone || 'blue'}`
}

function resetFilters() {
  Object.assign(summaryFilters, { keyword: '', department: '', period: '', nextPeriod: '', status: '', managerId: '', hrbpId: '', grade: '' })
}

function currentMonth() {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

function nextMonth() {
  const now = new Date()
  now.setMonth(now.getMonth() + 1)
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

function toast(title) {
  uni.showToast({ title, icon: 'none' })
}

providePerformanceContext({
  activePerformanceTab,
  canEditObjectiveDimension,
  canEmployeeConfirm,
  canFinal,
  canHrbpHandle,
  canManager,
  canSelf,
  canWithdraw,
  contentView,
  createReview,
  createUser,
  currentReviews,
  dashboardFilter,
  deleteReview,
  deleteUser,
  departments,
  effectiveGrade,
  exportSummary,
  gradeOptions,
  grades,
  hrbpIds,
  refreshData,
  managerIds,
  newReview,
  newUser,
  objectiveScore,
  pageTitle,
  performanceTabs,
  performReviewAction,
  queueReviews,
  resetFilters,
  returnReview,
  reviewTab,
  reviewTargetUsers,
  roleName,
  roleText,
  saveUser,
  selectReview,
  selectedReview,
  selectedReviewId,
  sectionTitle,
  statCards,
  state,
  statusClass,
  statusMeta,
  statusText,
  summaryFilters,
  summaryReviews,
  switchPerformanceTab,
  totalObjectiveScore,
  userName,
  userOptionText,
  valueTotal,
  workbenchCards,
  withdrawReview
})

onMounted(async () => {
  if (!authToken()) {
    ready.value = true
    return
  }
  try {
    await loadBootstrap()
    await refreshData()
  } catch (error) {
    clearAuthToken()
  } finally {
    ready.value = true
  }
})
</script>
