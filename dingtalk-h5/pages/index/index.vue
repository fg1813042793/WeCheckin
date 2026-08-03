<template>
  <view class="dt-page">
    <view v-if="!ready" class="loading-page">加载中...</view>

    <BindAccountView
      v-else-if="!state.user && bindState.visible"
      :form="bindForm"
      :bind-state="bindState"
      :loading="loading"
      :app-config="appConfig"
      @bind="bindDingTalkUser"
      @retry="retryDingTalkAutoLogin"
    />

    <view v-else-if="!state.user && sessionAccessDenied" class="loading-page no-permission-page">
      <text>{{ sessionAccessDeniedMessage }}</text>
      <button class="dt-btn dt-btn-light" @click="resetSessionState">重新登录</button>
    </view>

    <LoginView
      v-else-if="!state.user"
      :form="loginForm"
      :loading="loading"
      :app-config="appConfig"
      :auto-login-message="autoLoginMessage"
      @login="login"
    />

    <AppShell
      v-else
      :active-view="activeView"
      :active-performance-tab="activePerformanceTab"
      :active-route-tab="state.view"
      :app-config="appConfig"
      :app-title="appTitle"
      :nav-items="navItems"
      :page-title="pageTitle"
      :route-tabs="routeTabs"
      :user="state.user"
      @activate-route-tab="activateRouteTab"
      @close-route-tab="closeRouteTab"
      @logout="logout"
      @open-profile="openProfileDialog"
      @switch-view="switchView"
    >
      <view v-if="navItems.length === 0" class="empty no-permission">暂无可用菜单，请联系管理员配置钉钉 H5 权限</view>
      <view v-else-if="contentLoading" class="content-loading-view">
        <view class="page-head">
          <view>
            <text class="page-title">{{ sectionTitle }}</text>
            <text class="page-desc">加载中...</text>
          </view>
        </view>
        <section class="panel performance-loading-panel">
          <view class="performance-loading-head">
            <view class="performance-loading-title"></view>
            <view class="performance-loading-action"></view>
          </view>
          <view class="performance-loading-table">
            <view v-for="row in 6" :key="`loading-row-${row}`" class="performance-loading-row">
              <view
                v-for="cell in 6"
                :key="`loading-cell-${row}-${cell}`"
                :class="['performance-loading-cell', cell === 1 ? 'wide' : '', row === 1 ? 'head' : '']"
              ></view>
            </view>
          </view>
        </section>
      </view>
      <SummaryView v-else-if="contentView === 'summary'" />
      <OrgView v-else-if="contentView === 'org'" />
      <TemplateView v-else-if="contentView === 'template'" />
      <WorkbenchView v-else />

      <view v-if="profileDialog.visible" class="profile-center-modal" @click="closeProfileDialog">
        <view class="profile-center-card" @click.stop>
          <view class="profile-center-head">
            <view>
              <text class="profile-center-title">个人中心</text>
              <text class="profile-center-desc">维护当前登录账号、头像和登录密码。</text>
            </view>
            <button class="process-modal-close" @click="closeProfileDialog">×</button>
          </view>
          <view class="profile-center-body">
            <view class="profile-center-avatar-row">
              <image v-if="profileAvatarPreview" class="profile-center-avatar-image" :src="profileAvatarPreview" mode="aspectFill" />
              <view v-else class="profile-center-avatar-text">{{ profileInitial }}</view>
              <view class="profile-center-avatar-meta">
                <text class="profile-center-avatar-title">{{ profileDisplayName }}</text>
                <text class="profile-center-avatar-desc">填写头像地址后会同步到顶部头像展示。</text>
              </view>
            </view>
            <view class="profile-center-form">
              <view class="profile-center-field">
                <text class="profile-center-label">头像地址</text>
                <input
                  class="field-input"
                  v-model="profileDialog.avatar"
                  placeholder="http(s) 地址或 / 开头的站内路径"
                />
              </view>
              <view class="profile-center-field">
                <text class="profile-center-label">账号</text>
                <input class="field-input" v-model="profileDialog.account" placeholder="请输入账号" />
              </view>
              <view class="profile-center-field">
                <text class="profile-center-label">当前密码</text>
                <input class="field-input" v-model="profileDialog.currentPassword" password placeholder="修改账号或密码时必填" />
              </view>
              <view class="profile-center-grid">
                <view class="profile-center-field">
                  <text class="profile-center-label">新密码</text>
                  <input class="field-input" v-model="profileDialog.newPassword" password placeholder="不修改可留空" />
                </view>
                <view class="profile-center-field">
                  <text class="profile-center-label">确认新密码</text>
                  <input class="field-input" v-model="profileDialog.confirmPassword" password placeholder="再次输入新密码" />
                </view>
              </view>
            </view>
          </view>
          <view class="profile-center-actions">
            <button class="dt-btn dt-btn-light" :disabled="profileDialog.loading" @click="closeProfileDialog">取消</button>
            <button class="dt-btn dt-btn-primary" :loading="profileDialog.loading" @click="submitProfileDialog">保存</button>
          </view>
        </view>
      </view>

      <view v-if="createReviewDialog.visible" class="review-create-modal" @click="closeCreateReviewDialog">
        <view class="review-create-card" @click.stop="createReviewMonthPickerOpen = false">
          <view class="review-create-head">
            <view>
              <text class="review-create-title">新建考评单</text>
              <text class="review-create-desc">选择被考评人和考评月份，支持按部门多选创建</text>
            </view>
            <button class="review-create-close" @click="closeCreateReviewDialog">×</button>
          </view>
          <view class="review-create-form">
            <view class="review-create-field review-create-field-users">
              <text class="review-create-label">被考评人</text>
              <view class="create-target-search">
                <input
                  class="field-input create-target-search-input"
                  v-model="createReviewUserKeyword"
                  placeholder="搜索姓名/账号/部门/岗位"
                />
                <button
                  v-if="createReviewUserKeyword"
                  class="create-target-search-clear"
                  @click="createReviewUserKeyword = ''"
                >清空</button>
              </view>
              <view class="department-user-tree">
                <view
                  v-for="row in createTargetUserTreeRows"
                  :key="row.key"
                  :class="['create-target-row', row.type, `depth-${row.depth}`]"
                >
                  <view
                    v-if="row.type === 'department'"
                    class="create-target-dept-head"
                    :class="{ expanded: row.expanded }"
                    @click="toggleCreateReviewDept(row.key)"
                  >
                    <view
                      class="create-target-dept-check"
                      :class="[
                        createTargetDepartmentCheckState(row) === 'checked' ? 'checked' : '',
                        createTargetDepartmentCheckState(row) === 'indeterminate' ? 'create-target-dept-check-indeterminate' : ''
                      ]"
                      @click.stop="toggleCreateReviewDepartment(row)"
                    >
                      <text v-if="createTargetDepartmentCheckState(row) === 'checked'">✓</text>
                      <text v-else-if="createTargetDepartmentCheckState(row) === 'indeterminate'">-</text>
                    </view>
                    <view class="create-target-dept-title">
                      <text class="create-target-dept-chevron" :class="{ expanded: row.expanded }"></text>
                      <text class="create-target-dept-name">{{ row.name }}</text>
                    </view>
                    <text class="create-target-dept-count">{{ createTargetDepartmentUserIds(row).length }} 人</text>
                  </view>
                  <view
                    v-else-if="row.type === 'employee'"
                    class="create-target-user-tree"
                    :class="{ selected: createReviewForm.employeeIds.includes(row.user.id) }"
                    @click="toggleCreateReviewEmployee(row.user.id)"
                  >
                    <view class="create-target-check">
                      <text v-if="createReviewForm.employeeIds.includes(row.user.id)">✓</text>
                    </view>
                    <view class="create-target-user-main">
                      <text class="create-target-user-name">{{ row.user.name || row.user.id }}</text>
                      <text class="create-target-user-meta">{{ createTargetUserMeta(row.user) }}</text>
                    </view>
                  </view>
                </view>
                <view v-if="createTargetUserTree.length === 0" class="create-target-empty">{{ createTargetUserEmptyText }}</view>
              </view>
            </view>
            <view class="review-create-inline-fields single">
              <view class="review-create-field">
                <text class="review-create-label">考评月份</text>
                <view class="review-create-month-picker">
                  <button
                    class="field-input review-create-month-trigger"
                    :class="{ selected: createReviewForm.period }"
                    @click.stop="toggleCreateReviewMonthPicker"
                  >
                    <text class="review-create-month-text">{{ createReviewMonthText(createReviewForm.period) }}</text>
                    <text class="review-create-month-arrow" :class="{ open: createReviewMonthPickerOpen }"></text>
                  </button>
                  <view v-if="createReviewMonthPickerOpen" class="review-create-month-dropdown" @click.stop>
                    <view class="review-create-month-head">
                      <button class="review-create-month-nav" @click="changeCreateReviewMonthPickerYear(-1)">‹</button>
                      <text class="review-create-month-year">{{ createReviewMonthPickerYear }}年</text>
                      <button class="review-create-month-nav" @click="changeCreateReviewMonthPickerYear(1)">›</button>
                    </view>
                    <view class="review-create-month-grid">
                      <button
                        v-for="month in createReviewMonthOptions"
                        :key="month.value"
                        class="review-create-month-option"
                        :class="{ active: isCreateReviewMonthSelected(month.value) }"
                        @click="selectCreateReviewMonth(month.value)"
                      >{{ month.label }}</button>
                    </view>
                  </view>
                </view>
              </view>
            </view>
          </view>
          <view class="review-create-actions">
            <button class="dt-btn dt-btn-light" @click="closeCreateReviewDialog">取消</button>
            <button class="dt-btn dt-btn-primary" :loading="createReviewDialog.loading" @click="createReview">创建</button>
          </view>
        </view>
      </view>

      <view v-if="withdrawDialog.visible" class="review-action-modal" @click="closeWithdrawDialog">
        <view class="review-action-card withdraw-confirm-card" @click.stop>
          <view class="review-action-head">
            <view>
              <text class="review-action-title">撤回提交</text>
              <text class="review-action-desc">撤回后考评单将回到上一流程节点，流程记录会保存撤回理由。</text>
            </view>
            <button class="process-modal-close" @click="closeWithdrawDialog">×</button>
          </view>
          <view class="review-action-body">
            <text class="review-action-label">撤回理由</text>
            <textarea
              class="field-textarea withdraw-reason-textarea"
              v-model="withdrawDialog.reason"
              maxlength="200"
              placeholder="请填写撤回原因"
            ></textarea>
            <text class="review-action-helper">{{ withdrawReasonLength }}/200</text>
          </view>
          <view class="review-action-actions">
            <button class="dt-btn dt-btn-light" :disabled="withdrawDialog.loading" @click="closeWithdrawDialog">取消</button>
            <button class="dt-btn dt-btn-danger" :loading="withdrawDialog.loading" @click="submitWithdrawReview">确认撤回</button>
          </view>
        </view>
      </view>

      <view v-if="returnDialog.visible" class="review-action-modal" @click="closeReturnDialog">
        <view class="review-action-card withdraw-confirm-card" @click.stop>
          <view class="review-action-head">
            <view>
              <text class="review-action-title">{{ returnDialog.title }}</text>
              <text class="review-action-desc">{{ returnDialog.desc }}</text>
            </view>
            <button class="process-modal-close" @click="closeReturnDialog">×</button>
          </view>
          <view class="review-action-body">
            <text class="review-action-label">{{ returnDialog.reasonLabel }}</text>
            <textarea
              class="field-textarea withdraw-reason-textarea"
              v-model="returnDialog.reason"
              maxlength="200"
              :placeholder="returnDialog.placeholder"
            ></textarea>
            <text class="review-action-helper">{{ returnReasonLength }}/200</text>
          </view>
          <view class="review-action-actions">
            <button class="dt-btn dt-btn-light" :disabled="returnDialog.loading" @click="closeReturnDialog">取消</button>
            <button class="dt-btn dt-btn-danger" :loading="returnDialog.loading" @click="submitReturnReview">确认退回</button>
          </view>
        </view>
      </view>
    </AppShell>
  </view>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import AppShell from '../../components/performance/AppShell.vue'
import BindAccountView from '../../components/performance/BindAccountView.vue'
import LoginView from '../../components/performance/LoginView.vue'
import OrgView from '../../components/performance/OrgView.vue'
import SummaryView from '../../components/performance/SummaryView'
import TemplateView from '../../components/performance/TemplateView'
import WorkbenchView from '../../components/performance/WorkbenchView'
import { providePerformanceContext } from '../../components/performance/context'
import { dingTalkAuthApi, dingTalkPerformanceApi } from '../../services/dingtalkH5Api'
import { isDingTalkRuntime, requestAuthCode, waitForDingTalkJSAPI } from '../../utils/dingtalk'
import { AUTH_EXPIRED_EVENT, authToken, clearAuthToken, isAuthExpiredError, isBindRequiredError, isPermissionDeniedError, setAuthToken } from '../../utils/request'
import CONFIG from '../../config'

const statusMeta = {
  draft: { label: '员工填写', tone: 'warning', step: 0 },
  manager_review: { label: '上级评价', tone: 'purple', step: 1 },
  hrbp_review: { label: 'HRBP评价', tone: 'blue', step: 2 },
  employee_confirm: { label: '员工确认', tone: 'warning', step: 3 },
  hr_final: { label: 'HRBP归档', tone: 'orange', step: 4 },
  completed: { label: '已完成', tone: 'green', step: 5 }
}

const myPerformanceStatuses = ['draft', 'manager_review', 'hrbp_review', 'employee_confirm', 'hr_final']
const myPerformanceStatusSet = new Set(myPerformanceStatuses)

const reviewActionApiPermissions = {
  'save-self': 'dingtalk_h5:api:review:self_save',
  'submit-self': 'dingtalk_h5:api:review:self_submit',
  'submit-manager': 'dingtalk_h5:api:review:manager_submit',
  'submit-hrbp': 'dingtalk_h5:api:review:hrbp_submit',
  'confirm-result': 'dingtalk_h5:api:review:confirm',
  'dispute-result': 'dingtalk_h5:api:review:dispute',
  withdraw: 'dingtalk_h5:api:review:withdraw',
  'return-employee': 'dingtalk_h5:api:review:return_employee',
  'return-manager': 'dingtalk_h5:api:review:return_manager',
  'return-hrbp': 'dingtalk_h5:api:review:return_hrbp',
  finalize: 'dingtalk_h5:api:review:finalize'
}

const reviewActionButtonPermissions = {
  'save-self': 'dingtalk_h5:button:review:self_save',
  'submit-self': 'dingtalk_h5:button:review:self_submit',
  'submit-manager': 'dingtalk_h5:button:review:manager_submit',
  'submit-hrbp': 'dingtalk_h5:button:review:hrbp_submit',
  'confirm-result': 'dingtalk_h5:button:review:confirm',
  'dispute-result': 'dingtalk_h5:button:review:dispute',
  withdraw: 'dingtalk_h5:button:review:withdraw',
  'return-employee': 'dingtalk_h5:button:review:return_employee',
  'return-manager': 'dingtalk_h5:button:review:return_manager',
  'return-hrbp': 'dingtalk_h5:button:review:return_hrbp',
  finalize: 'dingtalk_h5:button:review:finalize'
}

const reviewActionConfirmCopy = {
  'submit-self': {
    title: '提交自评',
    content: '确认提交当前绩效？提交后将进入上级评价流程。',
    confirmText: '提交'
  },
  'submit-manager': {
    title: '提交给HRBP',
    content: '确认提交给 HRBP？提交后将进入 HRBP 评价流程。',
    confirmText: '提交'
  },
  'submit-hrbp': {
    title: '提交给员工',
    content: '确认提交给员工确认？提交后员工将查看并确认绩效结果。',
    confirmText: '提交'
  },
  'confirm-result': {
    title: '确认结果',
    content: '确认绩效结果无误？确认后将进入 HRBP 归档流程。',
    confirmText: '确认'
  },
  'dispute-result': {
    title: '提出异议',
    content: '确认提交异议？提交后将返回 HRBP 处理。',
    confirmText: '提交',
    confirmColor: '#e34d59'
  },
  'return-employee': {
    title: '退回员工',
    content: '确认退回员工修改？退回后员工可重新编辑自评内容。',
    confirmText: '退回',
    confirmColor: '#e34d59'
  },
  'return-manager': {
    title: '退回上级',
    content: '确认退回上级修改？退回后上级可重新调整评价。',
    confirmText: '退回',
    confirmColor: '#e34d59'
  },
  'return-hrbp': {
    title: '退回HRBP',
    content: '确认退回 HRBP 修改？退回后 HRBP 可重新处理评价。',
    confirmText: '退回',
    confirmColor: '#e34d59'
  },
  finalize: {
    title: '绩效归档',
    content: '确认归档绩效结果？归档后流程将完成。',
    confirmText: '归档'
  },
  'delete-review': {
    title: '删除考评单',
    content: '确认删除这张考评单？删除后将不再展示。',
    confirmText: '删除',
    confirmColor: '#e34d59'
  }
}

const ready = ref(false)
const loading = ref(false)
const refreshing = ref(false)
const sessionAccessDenied = ref(false)
const sessionAccessDeniedMessage = ref('无权限访问，请联系管理员配置钉钉 H5 权限')
const contentLoading = ref(false)
const reviewTab = ref('currentTargets')
const dashboardFilter = ref('queue')
const selectedReviewId = ref('')
const activePerformanceTab = ref('mine')
const routeTabs = ref([])
const managerReviewTab = ref('pending')
const hrbpReviewTab = ref('pending')

const loginForm = reactive({ name: '', password: '' })
const bindForm = reactive({ account: '', password: '' })
const bindState = reactive({
  visible: false,
  bindTicket: '',
  corpId: '',
  dingTalkUserIdMasked: '',
  unionIdMasked: '',
  expiresIn: 0
})
const autoLoginTried = ref(false)
const autoLoginMessage = ref('')
const publicCorpId = ref('')
const profileDialog = reactive({
  visible: false,
  loading: false,
  account: '',
  avatar: '',
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})
const newReview = reactive({ employeeId: '' })
const createReviewDialog = reactive({ visible: false, loading: false })
const createReviewForm = reactive({ employeeIds: [], period: currentMonth() })
const createReviewUserKeyword = ref('')
const createReviewExpandedDeptKeys = ref(new Set())
const createReviewMonthPickerOpen = ref(false)
const createReviewMonthPickerYear = ref(new Date().getFullYear())
const withdrawDialog = reactive({ visible: false, loading: false, reason: '' })
const returnDialog = reactive({
  visible: false,
  loading: false,
  action: '',
  title: '退回',
  desc: '',
  reasonLabel: '退回原因',
  placeholder: '请填写退回原因',
  reason: ''
})
const summaryFilters = reactive({
  employeeName: '',
  departmentName: '',
  departmentNames: [],
  period: '',
  status: ''
})
let contentLoadingSeq = 0
let publicConfigLoaded = false

const state = reactive({
  user: null,
  menus: [],
  users: [],
  reviews: [],
  reviewListTotal: 0,
  reviewPage: 1,
  reviewPageSize: 20,
  workbenchStats: [],
  template: null,
  buttonPermissionKeys: [],
  buttonPermissionReady: false,
  apiPermissionKeys: [],
  apiPermissionReady: false,
  appConfig: defaultAppConfig(),
  appTitle: 'OA管理',
  permissionVersion: 0,
  view: 'dashboard'
})

const menuPageKeys = new Set([
  'dashboard',
  'performance',
  'performance:mine',
  'performance:history',
  'performance:manager',
  'performance:hrbp',
  'performance:summary',
  'performance:org',
  'performance:template'
])

const menuTreeItems = computed(() => normalizeMenuTree(state.menus))
const flatMenuItems = computed(() => flattenMenuTree(menuTreeItems.value))

const navItems = computed(() => {
  if (menuTreeItems.value.length > 0) {
    const items = menuTreeItems.value
      .filter((item) => ['dashboard', 'performance'].includes(item.key))
      .map((item) => ({
        key: item.key,
        label: menuLabel(item),
        icon: menuIcon(item),
        children: (item.children || [])
          .filter((child) => menuPageKeys.has(child.key))
          .map((child) => ({ key: child.key, label: menuLabel(child), icon: menuIcon(child) }))
      }))
    if (items.length > 0) return items
  }
  return []
})

const performanceTabs = computed(() => {
  const performanceMenu = menuTreeItems.value.find((item) => item.key === 'performance')
  return (performanceMenu?.children || [])
    .filter((item) => menuPageKeys.has(item.key))
    .map((item) => ({ key: item.key, label: menuLabel(item), icon: menuIcon(item) }))
})

const activeView = computed(() => String(state.view || '').startsWith('performance:') ? 'performance' : state.view)
const activeMenuItem = computed(() => {
  return flatMenuItems.value.find((item) => item.key === state.view) ||
    navItems.value.find((item) => item.key === activeView.value) ||
    null
})
const contentView = computed(() => menuContentKey(state.view))
const appConfig = computed(() => normalizeAppConfig(state.appConfig))
const appTitle = computed(() => firstText(appConfig.value.appTitle, state.appTitle, 'OA管理'))
const profileAvatarPreview = computed(() => firstText(profileDialog.avatar))
const profileDisplayName = computed(() => firstText(state.user?.name, state.user?.account, state.user?.id, '当前用户'))
const profileInitial = computed(() => firstText(profileDisplayName.value).slice(0, 1).toUpperCase() || 'U')

const pageTitle = computed(() => {
  return activeMenuItem.value?.label || titleForContent(contentView.value)
})

const sectionTitle = computed(() => {
  return activeMenuItem.value?.label || titleForContent(contentView.value)
})

const selectedReview = computed(() => {
  if (contentView.value === 'mine') {
    return currentReviews.value.find((item) => item.id === selectedReviewId.value) || currentReviews.value[0] || null
  }
  return state.reviews.find((item) => item.id === selectedReviewId.value) || currentReviews.value[0] || null
})

const currentReviews = computed(() => {
  let list = [...state.reviews]
  const view = contentView.value
  if (view === 'mine') {
    list = myPerformanceReviews()
  } else if (view === 'history') {
    list = list.filter((item) => sameUserId(item.employeeId, state.user?.id) && item.period !== currentMonth())
  } else if (view === 'manager') {
    list = list.filter((item) => sameUserId(item.managerId, state.user?.id) && ['manager_review', 'hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(item.status))
  } else if (view === 'hrbp') {
    list = list.filter((item) => ['hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(item.status))
  } else if (view === 'dashboard') {
    list = queueReviews()
  }
  return list
})

const summaryReviews = computed(() => {
  const employeeName = summaryFilters.employeeName.trim().toLowerCase()
  return state.reviews.filter((item) => {
    const employee = userName(item.employeeId)
    const employeeText = [employee, item.employeeName, item.employeeId].filter(Boolean).join(' ').toLowerCase()
    if (employeeName && !employeeText.includes(employeeName)) return false
    if (!summaryDepartmentMatches(item.department, summaryFilters.departmentName, summaryFilters.departmentNames)) return false
    if (summaryFilters.period && item.period !== summaryFilters.period) return false
    if (summaryFilters.status && item.status !== summaryFilters.status) return false
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
const reviewTargetUsers = computed(() => state.users.filter((item) => item && item.id && Number(item.status || 1) === 1))
const createReviewTargetUsers = computed(() => {
  if (reviewTargetUsers.value.length > 0) return reviewTargetUsers.value
  return state.user?.id ? [state.user] : []
})
const createReviewSearchKeyword = computed(() => normalizeSearchKeyword(createReviewUserKeyword.value))
const filteredCreateReviewTargetUsers = computed(() => {
  const keyword = createReviewSearchKeyword.value
  if (!keyword) return createReviewTargetUsers.value
  return createReviewTargetUsers.value.filter((user) => createTargetUserMatchesKeyword(user, keyword))
})
const createTargetUserTree = computed(() => buildCreateTargetUserTree(filteredCreateReviewTargetUsers.value))
const createTargetUserTreeRows = computed(() => flattenCreateTargetTree(createTargetUserTree.value, createReviewExpandedDeptKeys.value, 1, Boolean(createReviewSearchKeyword.value)))
const createTargetUserEmptyText = computed(() => {
  if (createReviewTargetUsers.value.length === 0) return '暂无可创建人员'
  return createReviewSearchKeyword.value ? '没有匹配的人员' : '暂无可创建人员'
})
const createReviewMonthOptions = computed(() => Array.from({ length: 12 }, (_, index) => ({
  value: index + 1,
  label: `${index + 1}月`
})))
const withdrawReasonLength = computed(() => [...String(withdrawDialog.reason || '')].length)
const returnReasonLength = computed(() => [...String(returnDialog.reason || '')].length)

function statusText(status) {
  return statusMeta[status]?.label || status
}

function sameUserId(left, right) {
  const leftText = String(left || '').trim()
  const rightText = String(right || '').trim()
  return Boolean(leftText && rightText && leftText === rightText)
}

function userName(id) {
  return reviewPersonNameFromReviews(id) || state.users.find((item) => sameUserId(item.id, id))?.name || id || '无'
}

function reviewPersonName(review, id) {
  if (!review || !id) return ''
  const pairs = [
    [review.employeeId, review.employeeName],
    [review.managerId, review.managerName],
    [review.hrbpId, review.hrbpName],
    [review.hrbpReviewerId, review.hrbpReviewerName]
  ]
  const found = pairs.find(([account, name]) => sameUserId(account, id) && firstText(name))
  return found ? firstText(found[1]) : ''
}

function reviewPersonNameFromReviews(id) {
  if (!id) return ''
  for (const review of state.reviews) {
    const name = reviewPersonName(review, id)
    if (name) return name
  }
  return sameUserId(state.user?.id, id) ? firstText(state.user.name) : ''
}

function userOptionText(user) {
  return [user.name, user.position, user.department].filter(Boolean).join(' · ')
}

function createTargetUserMeta(user) {
  return [user.position, user.department].filter(Boolean).join(' · ') || user.id
}

function normalizeSearchKeyword(value) {
  return String(value || '').trim().toLowerCase()
}

function createTargetUserMatchesKeyword(user, keyword) {
  const text = [
    user.id,
    user.name,
    user.account,
    user.mobile,
    user.phone,
    user.position,
    user.department,
    user.departmentLevel1,
    user.departmentLevel2,
    user.departmentLevel3,
    createTargetUserMeta(user)
  ].filter(Boolean).join(' ').toLowerCase()
  const terms = keyword.split(/\s+/).filter(Boolean)
  return terms.every((term) => text.includes(term))
}

function buildCreateTargetUserTree(users = []) {
  const root = new Map()
  for (const user of users || []) {
    const levels = createTargetDepartmentLevels(user)
    let currentMap = root
    let currentNode = null
    for (const [index, level] of levels.entries()) {
      const parentKey = currentNode?.key || 'root'
      currentNode = ensureCreateTargetNode(currentMap, `${parentKey}/l${index + 1}:${level}`, level)
      currentNode.count += 1
      if (user.id) currentNode.userIds.push(user.id)
      currentMap = currentNode.childMap
    }
    if (currentNode) {
      currentNode.users.push(user)
    }
  }
  return finalizeCreateTargetNodes([...root.values()])
}

function ensureCreateTargetNode(map, key, name) {
  if (!map.has(key)) {
    map.set(key, { key, name, count: 0, userIds: [], childMap: new Map(), children: [], users: [] })
  }
  return map.get(key)
}

function finalizeCreateTargetNodes(nodes = []) {
  return nodes
    .sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))
    .map((node) => ({
      key: node.key,
      name: node.name,
      count: node.count,
      userIds: [...new Set(node.userIds.filter(Boolean))],
      children: finalizeCreateTargetNodes([...node.childMap.values()]),
      users: node.users.slice().sort((left, right) => userSortText(left).localeCompare(userSortText(right), 'zh-CN'))
    }))
}

function flattenCreateTargetTree(nodes = [], expandedKeys, depth = 1, forceExpanded = false) {
  const rows = []
  for (const node of nodes) {
    const hasChildren = node.children.length > 0 || node.users.length > 0
    const expanded = forceExpanded || expandedKeys.has(node.key)
    rows.push({
      type: 'department',
      key: node.key,
      depth,
      name: node.name,
      count: node.count,
      userIds: node.userIds,
      expandable: hasChildren,
      expanded
    })
    if (!expanded) continue
    for (const user of node.users) {
      rows.push({ type: 'employee', key: `${node.key}/user:${user.id}`, depth: depth + 1, user })
    }
    rows.push(...flattenCreateTargetTree(node.children, expandedKeys, depth + 1, forceExpanded))
  }
  return rows
}

function createTargetDepartmentLevels(user) {
  const parts = String(user.department || '').split('/').map((item) => item.trim()).filter(Boolean)
  const levels = [
    firstText(user.departmentLevel1, parts[0]),
    firstText(user.departmentLevel2, parts[1]),
    firstText(user.departmentLevel3, parts[2])
  ].filter(Boolean)
  return levels.length > 0 ? levels : ['未设置部门']
}

function userSortText(user) {
  return [user.departmentLevel1, user.departmentLevel2, user.departmentLevel3, user.name, user.id].filter(Boolean).join('\x00')
}

function unique(items) {
  return [...new Set(items.filter(Boolean))]
}

function queueReviews() {
  return state.reviews.filter((item) => canSelf(item) || canManager(item) || canHrbpHandle(item) || canEmployeeConfirm(item) || canFinal(item))
}

function myPerformanceReviews() {
  return state.reviews.filter((item) => sameUserId(item.employeeId, state.user?.id) && myPerformanceStatusSet.has(item.status))
}

async function switchView(view) {
  const nextView = view === 'performance' ? preferredPerformanceView() : view
  const sameView = state.view === nextView
  state.view = nextView
  syncActivePerformanceTab()
  ensureRouteTab()
  reviewTab.value = 'currentTargets'
  if (!sameView) {
    selectedReviewId.value = ''
  }
  await refreshDataSafely({ contentLoading: true })
}

async function activateRouteTab(view) {
  if (!view || view === state.view) return
  await switchView(view)
}

async function closeRouteTab(view) {
  const key = String(view || '')
  if (!key) return
  const tabs = routeTabs.value
  const index = tabs.findIndex((item) => item.key === key)
  if (index < 0) return
  const nextTabs = tabs.filter((item) => item.key !== key)
  routeTabs.value = nextTabs
  if (state.view !== key) return
  const nextTab = tabs[index + 1] || tabs[index - 1] || nextTabs[0] || null
  if (nextTab) {
    await switchView(nextTab.key)
    return
  }
  ensureActiveMenu()
  await refreshDataSafely({ contentLoading: true })
}

function ensureActiveMenu() {
  const items = navItems.value
  if (!items.length) {
    state.view = ''
    routeTabs.value = []
    return
  }
  if (state.view === 'performance' && performanceTabs.value.length > 0) {
    state.view = preferredPerformanceView()
    syncActivePerformanceTab()
    syncRouteTabLabels()
    ensureRouteTab()
    return
  }
  if (flatMenuItems.value.some((item) => item.key === state.view)) {
    syncActivePerformanceTab()
    syncRouteTabLabels()
    ensureRouteTab()
    return
  }
  const first = items[0]
  state.view = first.key === 'performance' ? preferredPerformanceView() : first.key
  syncActivePerformanceTab()
  syncRouteTabLabels()
  ensureRouteTab()
}

function ensureRouteTab(view = state.view) {
  const key = String(view || '')
  if (!key) return
  const label = routeTabLabel(key)
  const nextTab = {
    key,
    label,
    closable: key !== 'dashboard'
  }
  const index = routeTabs.value.findIndex((item) => item.key === key)
  if (index >= 0) {
    routeTabs.value = routeTabs.value.map((item, itemIndex) => itemIndex === index ? { ...item, ...nextTab } : item)
    return
  }
  routeTabs.value = [...routeTabs.value, nextTab]
}

function syncRouteTabLabels() {
  if (!routeTabs.value.length) return
  routeTabs.value = routeTabs.value
    .filter((item) => item && item.key && flatMenuItems.value.some((menu) => menu.key === item.key))
    .map((item) => ({ ...item, label: routeTabLabel(item.key), closable: item.key !== 'dashboard' }))
}

function routeTabLabel(key) {
  const menu = flatMenuItems.value.find((item) => item.key === key) ||
    navItems.value.find((item) => item.key === key)
  return menu ? menuLabel(menu) : titleForContent(menuContentKey(key))
}

function selectReview(id) {
  selectedReviewId.value = id
  reviewTab.value = 'currentTargets'
}

function workbenchTodoTarget(review) {
  if (canSelf(review)) {
    return { view: 'performance:mine', reviewTab: 'currentTargets' }
  }
  if (canManager(review)) {
    return { view: 'performance:manager', managerTab: 'pending', reviewTab: 'manager' }
  }
  if (canHrbpHandle(review)) {
    return { view: 'performance:hrbp', hrbpTab: 'pending', reviewTab: 'hrbp' }
  }
  if (canEmployeeConfirm(review)) {
    return { view: 'performance:mine', reviewTab: 'manager' }
  }
  if (canFinal(review)) {
    return { view: 'performance:hrbp', hrbpTab: 'reviewed', reviewTab: 'hrbp' }
  }
  return { view: 'performance:mine', reviewTab: 'currentTargets' }
}

async function openWorkbenchTodo(review) {
  if (!review?.id) return
  const target = workbenchTodoTarget(review)
  if (target.managerTab) managerReviewTab.value = target.managerTab
  if (target.hrbpTab) hrbpReviewTab.value = target.hrbpTab
  state.view = target.view
  syncActivePerformanceTab()
  ensureRouteTab()
  selectedReviewId.value = review.id
  reviewTab.value = target.reviewTab || 'currentTargets'
  await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
  selectedReviewId.value = review.id
  reviewTab.value = target.reviewTab || 'currentTargets'
}

async function login() {
  loading.value = true
  try {
    autoLoginMessage.value = ''
    resetBindState()
    resetRouteTabState()
    state.view = 'dashboard'
    const res = await dingTalkAuthApi.login(loginForm)
    const payload = res.data || {}
    sessionAccessDenied.value = false
    setAuthToken(payload.token)
    applySessionAuthPayload(payload)
    if (!payloadHasSessionPermissions(payload)) {
      const bootstrapped = await loadBootstrapSafely()
      if (!bootstrapped) return
    }
    await refreshDataSafely({ contentLoading: true })
  } finally {
    loading.value = false
  }
}

function queryParam(name) {
  if (typeof window === 'undefined') return ''
  const searchParams = new URLSearchParams(window.location.search || '')
  const value = searchParams.get(name)
  if (value) return value
  const hash = window.location.hash || ''
  const queryIndex = hash.indexOf('?')
  if (queryIndex < 0) return ''
  return new URLSearchParams(hash.slice(queryIndex + 1)).get(name) || ''
}

function queryCorpId() {
  return queryParam('corpId') || queryParam('corpID') || queryParam('corp_id') || ''
}

function currentCorpId() {
  return queryCorpId() || publicCorpId.value || CONFIG.DINGTALK_CORP_ID || ''
}

async function shouldTryDingTalkAutoLogin() {
  if (isDingTalkRuntime()) return true
  if (!currentCorpId()) return false
  await waitForDingTalkJSAPI(1200)
  return isDingTalkRuntime()
}

async function ensurePublicConfig() {
  if (publicConfigLoaded) {
    return true
  }
  publicConfigLoaded = true
  try {
    const res = await dingTalkAuthApi.publicConfig()
    const data = res.data || {}
    publicCorpId.value = String(data.corpId || '').trim()
    applyPublicAppConfig(data)
    return true
  } catch (error) {
    publicConfigLoaded = false
    return false
  }
}

function applyPublicAppConfig(payload = {}) {
  const config = payload.appConfig || payload
  state.appConfig = {
    appTitle: config.appTitle || payload.appTitle || state.appConfig.appTitle,
    appName: config.appName || payload.appName || state.appConfig.appName,
    logoText: config.logoText || payload.logoText || state.appConfig.logoText,
    logoUrl: config.logoUrl || payload.logoUrl || state.appConfig.logoUrl
  }
  state.appTitle = state.appConfig.appTitle || state.appTitle
}

async function tryDingTalkAutoLogin() {
  if (autoLoginTried.value || authToken()) return false
  autoLoginTried.value = true
  autoLoginMessage.value = ''
  await ensurePublicConfig()
  const shouldTry = await shouldTryDingTalkAutoLogin()
  const inDingTalk = isDingTalkRuntime()
  if (!shouldTry) {
    if (inDingTalk) {
      setDingTalkAutoLoginMessage(currentCorpId()
        ? '未检测到钉钉免登环境，请确认当前页面是在钉钉工作台应用内打开。'
        : '未配置钉钉企业 CorpId，请在后台“钉钉应用管理 / 配置选项”配置企业应用，或在应用地址增加 corpId 参数。')
    }
    return false
  }
  loading.value = true
  try {
    const corpId = currentCorpId()
    if (!corpId) {
      setDingTalkAutoLoginMessage('未配置钉钉企业 CorpId，请在后台“钉钉应用管理 / 配置选项”配置企业应用，或在应用地址增加 corpId 参数。')
      return false
    }
    const authCode = await requestAuthCode(corpId)
    if (!authCode) {
      setDingTalkAutoLoginMessage('钉钉免登未返回授权码，请确认当前应用地址配置在钉钉工作台微应用中。')
      return false
    }
    resetRouteTabState()
    state.view = 'dashboard'
    const res = await dingTalkAuthApi.ssoLogin({ corpId, authCode })
    const payload = res.data || {}
    sessionAccessDenied.value = false
    setAuthToken(payload.token)
    applySessionAuthPayload(payload)
    if (!payloadHasSessionPermissions(payload)) {
      const bootstrapped = await loadBootstrapSafely()
      if (!bootstrapped) return false
    }
    await refreshDataSafely({ contentLoading: true })
    return true
  } catch (error) {
    if (isBindRequiredError(error)) {
      showBindRequired(error?.data || {})
      return true
    }
    if (isPermissionDeniedError(error)) {
      sessionAccessDenied.value = true
      sessionAccessDeniedMessage.value = error?.msg || '无权限访问，请联系管理员配置钉钉 H5 权限'
    } else {
      setDingTalkAutoLoginMessage(autoLoginErrorMessage(error))
    }
    return false
  } finally {
    loading.value = false
  }
}

async function retryDingTalkAutoLogin() {
  resetBindState()
  clearAuthToken()
  autoLoginTried.value = false
  await tryDingTalkAutoLogin()
}

async function bindDingTalkUser() {
  if (!bindState.bindTicket) {
    toast('绑定会话已过期，请重新打开应用')
    return
  }
  loading.value = true
  try {
    resetRouteTabState()
    state.view = 'dashboard'
    const res = await dingTalkAuthApi.bindSelf({
      bindTicket: bindState.bindTicket,
      account: bindForm.account,
      password: bindForm.password
    })
    const payload = res.data || {}
    resetBindState()
    sessionAccessDenied.value = false
    setAuthToken(payload.token)
    applySessionAuthPayload(payload)
    if (!payloadHasSessionPermissions(payload)) {
      const bootstrapped = await loadBootstrapSafely()
      if (!bootstrapped) return
    }
    await refreshDataSafely({ contentLoading: true })
  } finally {
    loading.value = false
  }
}

async function logout() {
  try {
    await dingTalkAuthApi.logout()
  } finally {
    resetSessionState()
  }
}

function currentProfileAccount() {
  return firstText(state.user?.account, state.user?.id)
}

function currentProfileAvatar() {
  return firstText(state.user?.avatar, state.user?.avatarUrl, state.user?.pic, state.user?.userPic)
}

function resetProfileDialog() {
  Object.assign(profileDialog, {
    visible: false,
    loading: false,
    account: '',
    avatar: '',
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
}

function openProfileDialog() {
  if (!state.user) return
  Object.assign(profileDialog, {
    visible: true,
    loading: false,
    account: currentProfileAccount(),
    avatar: currentProfileAvatar(),
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
}

function closeProfileDialog() {
  if (profileDialog.loading) return
  resetProfileDialog()
}

async function submitProfileDialog() {
  if (profileDialog.loading) return
  const account = String(profileDialog.account || '').trim()
  const avatar = String(profileDialog.avatar || '').trim()
  const currentAccount = currentProfileAccount()
  const accountChanged = account !== currentAccount
  const avatarChanged = avatar !== currentProfileAvatar()
  const newPassword = String(profileDialog.newPassword || '').trim()
  const confirmPassword = String(profileDialog.confirmPassword || '').trim()
  const currentPassword = String(profileDialog.currentPassword || '').trim()
  const passwordChanging = Boolean(newPassword || confirmPassword)
  if (!account) {
    toast('请填写账号')
    return
  }
  if ((accountChanged || passwordChanging) && !currentPassword) {
    toast('修改账号或密码时请填写当前密码')
    return
  }
  if (passwordChanging) {
    if (newPassword.length < 6) {
      toast('新密码至少 6 位')
      return
    }
    if (newPassword !== confirmPassword) {
      toast('两次输入的新密码不一致')
      return
    }
  }
  if (!accountChanged && !avatarChanged && !passwordChanging) {
    toast('没有需要保存的修改')
    return
  }
  profileDialog.loading = true
  try {
    if (accountChanged || avatarChanged) {
      const res = await dingTalkAuthApi.updateProfile({ account, avatar, currentPassword })
      applyProfileUser(res.data?.user || res.data, currentAccount)
    }
    if (passwordChanging) {
      await dingTalkAuthApi.changePassword({
        currentPassword,
        newPassword,
        confirmPassword
      })
    }
    resetProfileDialog()
    toast('个人中心已保存')
  } catch (error) {
    toast(error?.msg || '保存失败')
  } finally {
    profileDialog.loading = false
  }
}

function applyProfileUser(user, oldAccount = '') {
  if (!user || !state.user) return
  const previousAccount = firstText(oldAccount, currentProfileAccount())
  const nextAccount = firstText(user.account, user.id, previousAccount)
  const nextUser = {
    ...state.user,
    ...user,
    id: nextAccount,
    account: nextAccount,
    avatar: firstText(user.avatar, user.avatarUrl, user.pic, user.userPic)
  }
  state.user = nextUser
  if (previousAccount && nextAccount && previousAccount !== nextAccount) {
    replaceLocalAccountReferences(previousAccount, nextAccount)
  }
  upsertProfileUser(nextUser, previousAccount)
}

function upsertProfileUser(user, oldAccount = '') {
  const next = sanitizeUsers([user])[0]
  if (!next?.id) return
  let replaced = false
  state.users = state.users.map((item) => {
    if (sameUserId(item.id, next.id) || sameUserId(item.id, oldAccount)) {
      replaced = true
      return { ...item, ...next }
    }
    return item
  })
  if (!replaced) {
    state.users.push(next)
  }
}

function replaceLocalAccountReferences(oldAccount, nextAccount) {
  for (const review of state.reviews) {
    for (const key of ['employeeId', 'managerId', 'hrbpId', 'hrbpReviewerId']) {
      if (sameUserId(review[key], oldAccount)) {
        review[key] = nextAccount
      }
    }
    for (const history of review.histories || []) {
      if (sameUserId(history.byAccount, oldAccount)) {
        history.byAccount = nextAccount
      }
    }
  }
}

function showBindRequired(data = {}) {
  clearAuthToken()
  autoLoginMessage.value = ''
  sessionAccessDenied.value = false
  sessionAccessDeniedMessage.value = '无权限访问，请联系管理员配置钉钉 H5 权限'
  state.user = null
  state.menus = []
  state.buttonPermissionKeys = []
  state.buttonPermissionReady = false
  state.apiPermissionKeys = []
  state.apiPermissionReady = false
  state.permissionVersion = 0
  bindState.visible = true
  bindState.bindTicket = String(data.bindTicket || '')
  bindState.corpId = String(data.corpId || '')
  bindState.dingTalkUserIdMasked = String(data.dingTalkUserIdMasked || '')
  bindState.unionIdMasked = String(data.unionIdMasked || '')
  bindState.expiresIn = Number(data.expiresIn || 0)
  bindForm.password = ''
  applySessionAuthPayload(data)
}

function resetBindState() {
  bindState.visible = false
  bindState.bindTicket = ''
  bindState.corpId = ''
  bindState.dingTalkUserIdMasked = ''
  bindState.unionIdMasked = ''
  bindState.expiresIn = 0
  bindForm.password = ''
}

function setDingTalkAutoLoginMessage(message) {
  if (!isDingTalkRuntime()) return
  autoLoginMessage.value = message
}

function autoLoginErrorMessage(error) {
  const rawMessage = String(error?.msg || error?.message || error?.errorMessage || error?.error || '').trim()
  if (!currentCorpId()) {
    return '未配置钉钉企业 CorpId，请先在后台“钉钉应用管理 / 配置选项”配置企业应用。'
  }
  if (rawMessage.includes('JSAPI') || rawMessage.includes('requestAuthCode') || rawMessage.includes('未检测到')) {
    return '未检测到钉钉免登 JSAPI，请确认页面是在钉钉工作台应用内打开，并且页面已加载钉钉 JSAPI。'
  }
  if (rawMessage.includes('notInDingTalk') || rawMessage.includes('不是钉钉端内')) {
    return '当前不是钉钉端内环境，请在钉钉工作台微应用中打开后使用免登。'
  }
  if (rawMessage) {
    return `钉钉免登失败：${rawMessage}`
  }
  return '钉钉免登失败，请确认 CorpId、应用凭证和钉钉工作台应用地址配置正确。'
}

function resetRouteTabState() {
  routeTabs.value = []
  activePerformanceTab.value = 'mine'
  reviewTab.value = 'currentTargets'
  managerReviewTab.value = 'pending'
  hrbpReviewTab.value = 'pending'
  dashboardFilter.value = 'queue'
}

function resetSessionState() {
  clearAuthToken()
  resetBindState()
  sessionAccessDenied.value = false
  sessionAccessDeniedMessage.value = '无权限访问，请联系管理员配置钉钉 H5 权限'
  contentLoading.value = false
  contentLoadingSeq += 1
  state.user = null
  state.menus = []
  state.buttonPermissionKeys = []
  state.buttonPermissionReady = false
  state.apiPermissionKeys = []
  state.apiPermissionReady = false
  state.appConfig = defaultAppConfig()
  state.appTitle = 'OA管理'
  state.permissionVersion = 0
  state.users = []
  state.reviews = []
  state.workbenchStats = []
  state.template = null
  state.view = 'dashboard'
  resetRouteTabState()
  selectedReviewId.value = ''
  resetProfileDialog()
  closeCreateReviewDialog()
}

async function loadBootstrap() {
  const res = await dingTalkPerformanceApi.bootstrap()
  applySessionAuthPayload(res.data || {})
}

async function loadBootstrapSafely() {
  try {
    await loadBootstrap()
    return true
  } catch (error) {
    return handleSessionDataError(error)
  }
}

async function refreshDataSafely(options = {}) {
  const useContentLoading = Boolean(options.contentLoading)
  const loadingSeq = useContentLoading ? ++contentLoadingSeq : 0
  if (useContentLoading) {
    contentLoading.value = true
  }
  try {
    await refreshData(options)
    return true
  } catch (error) {
    return handleSessionDataError(error)
  } finally {
    if (useContentLoading && loadingSeq === contentLoadingSeq) {
      contentLoading.value = false
    }
  }
}

async function refreshSessionAndDataSafely() {
  const bootstrapped = await loadBootstrapSafely()
  if (!bootstrapped) return false
  return refreshDataSafely({ forceReference: true, contentLoading: true })
}

async function refreshWithUserFeedback() {
  if (refreshing.value) return false
  refreshing.value = true
  showRefreshLoading()
  try {
    const refreshed = await refreshSessionAndDataSafely()
    hideRefreshLoading()
    if (refreshed) {
      toast('已刷新')
    } else {
      toast('刷新失败，请稍后重试')
    }
    return refreshed
  } catch (error) {
    handleSessionDataError(error)
    hideRefreshLoading()
    toast('刷新失败，请稍后重试')
    return false
  } finally {
    refreshing.value = false
  }
}

function handleSessionDataError(error) {
  if (isAuthExpiredError(error)) {
    resetSessionState()
    return false
  }
  if (isPermissionDeniedError(error)) {
    sessionAccessDenied.value = true
    sessionAccessDeniedMessage.value = error?.msg || '无权限访问，请联系管理员配置钉钉 H5 权限'
  }
  return false
}

function applySessionAuthPayload(payload = {}) {
  const user = payload.user || payload.userInfo
  if (user) {
    state.user = user
  }
  if (Array.isArray(payload.menus)) {
    state.menus = payload.menus
  }
  if (Array.isArray(payload.apiPermissionKeys)) {
    state.apiPermissionKeys = payload.apiPermissionKeys
  }
  if (Array.isArray(payload.buttonPermissionKeys)) {
    state.buttonPermissionKeys = payload.buttonPermissionKeys
  }
  if (Object.prototype.hasOwnProperty.call(payload, 'apiPermissionReady')) {
    state.apiPermissionReady = Boolean(payload.apiPermissionReady)
  }
  if (Object.prototype.hasOwnProperty.call(payload, 'buttonPermissionReady')) {
    state.buttonPermissionReady = Boolean(payload.buttonPermissionReady)
  }
  if (Object.prototype.hasOwnProperty.call(payload, 'permissionVersion')) {
    state.permissionVersion = Number(payload.permissionVersion || 0)
  }
  const payloadConfig = payload.appConfig && typeof payload.appConfig === 'object' ? payload.appConfig : {}
  const nextAppConfig = normalizeAppConfig({
    appTitle: firstText(payloadConfig.appTitle, payloadConfig.appName, payload.appTitle, payload.appName, payload.applicationName, state.appTitle),
    appName: firstText(payloadConfig.appName, payload.appName, payloadConfig.appTitle, payload.appTitle, state.appConfig.appName),
    logoText: firstText(payloadConfig.logoText, payload.logoText, state.appConfig.logoText),
    logoUrl: firstText(payloadConfig.logoUrl, payload.logoUrl, payload.logoURL, state.appConfig.logoUrl)
  })
  if (nextAppConfig.appTitle) {
    state.appConfig = nextAppConfig
    state.appTitle = nextAppConfig.appTitle
  }
  ensureActiveMenu()
}

function normalizeMenuTree(menus = [], isRoot = true) {
  if (!Array.isArray(menus)) return []
  const items = menus
    .filter((item) => item && menuPageKeys.has(item.key))
    .map((item) => ({
      ...item,
      label: menuLabel(item),
      children: normalizeMenuTree(item.children || [], false)
    }))
  const performanceMenu = items.find((item) => item.key === 'performance')
  if (performanceMenu && performanceMenu.children.length === 0) {
    performanceMenu.children = items
      .filter((item) => String(item.key || '').startsWith('performance:'))
      .map((item) => ({ ...item, children: [] }))
  }
  return isRoot ? items.filter((item) => !String(item.key || '').startsWith('performance:')) : items
}

function flattenMenuTree(menus = []) {
  return menus.flatMap((item) => [item, ...flattenMenuTree(item.children || [])])
}

function payloadHasSessionPermissions(payload = {}) {
  return Array.isArray(payload.menus) ||
    Array.isArray(payload.apiPermissionKeys) ||
    Array.isArray(payload.buttonPermissionKeys) ||
    Object.prototype.hasOwnProperty.call(payload, 'permissionVersion')
}

function hasApiPermission(key) {
  if (!state.apiPermissionReady) return false
  return state.apiPermissionKeys.includes(key)
}

function hasButtonPermission(key) {
  if (!state.buttonPermissionReady) return false
  return state.buttonPermissionKeys.includes(key)
}

function hasMenuPermission(key) {
  return flatMenuItems.value.some((item) => item.permissionKey === key)
}

function canCreateReview() {
  return hasButtonPermission('dingtalk_h5:button:review:create') &&
    hasApiPermission('dingtalk_h5:api:review:create')
}

function canDeleteReview() {
  return hasButtonPermission('dingtalk_h5:button:review:delete') &&
    hasApiPermission('dingtalk_h5:api:review:delete')
}

function canExportReviews() {
  return hasButtonPermission('dingtalk_h5:button:review:export') &&
    hasApiPermission('dingtalk_h5:api:review:export')
}

function canEditTemplate() {
  return hasButtonPermission('dingtalk_h5:button:template:edit') &&
    hasApiPermission('dingtalk_h5:api:template:save')
}

function canEditUsers() {
  return hasButtonPermission('dingtalk_h5:button:user:config') &&
    hasApiPermission('dingtalk_h5:api:user:edit')
}

function canEditNextObjectives(review) {
  return canSelf(review) &&
    hasButtonPermission('dingtalk_h5:button:review:next_objective_edit')
}

function canAddNextObjective(review) {
  return canEditNextObjectives(review) &&
    hasButtonPermission('dingtalk_h5:button:review:next_objective_add')
}

function canDeleteNextObjective(review) {
  return canEditNextObjectives(review) &&
    hasButtonPermission('dingtalk_h5:button:review:next_objective_delete')
}

function sanitizeUsers(users) {
  return (users || []).map((item) => ({ ...item, password: '' }))
}

function upsertUser(user) {
  const sanitized = sanitizeUsers([user])[0]
  if (!sanitized?.id) return
  const index = state.users.findIndex((item) => item.id === sanitized.id)
  if (index >= 0) {
    state.users[index] = sanitized
  } else {
    state.users.push(sanitized)
  }
}

function shouldAutoSelectReview(options = {}) {
  if (options.autoSelectReview === false) return false
  return contentView.value === 'mine'
}

function ensureSelectedReview(options = {}) {
  if (!newReview.employeeId) newReview.employeeId = reviewTargetUsers.value[0]?.id || ''
  if (!shouldAutoSelectReview(options)) {
    if (selectedReviewId.value && !state.reviews.some((item) => item.id === selectedReviewId.value)) {
      selectedReviewId.value = ''
    }
    return
  }
  if (!selectedReviewId.value || !currentReviews.value.some((item) => item.id === selectedReviewId.value)) {
    selectedReviewId.value = currentReviews.value[0]?.id || ''
  }
}

async function loadUsers() {
  if (!hasApiPermission('dingtalk_h5:api:user:list')) {
    state.users = []
    ensureSelectedReview()
    return
  }
  const res = await dingTalkPerformanceApi.users()
  state.users = sanitizeUsers(res.data || [])
  ensureSelectedReview()
}

async function openCreateReviewDialog() {
  if (!canCreateReview()) {
    toast('无权限创建考评单')
    return
  }
  createReviewForm.period = currentMonth()
  createReviewForm.employeeIds = []
  createReviewUserKeyword.value = ''
  createReviewExpandedDeptKeys.value = new Set()
  syncCreateReviewMonthPickerYear()
  createReviewMonthPickerOpen.value = false
  if (hasApiPermission('dingtalk_h5:api:user:list') && state.users.length === 0) {
    await loadUsers()
  }
  normalizeCreateReviewSelection()
  createReviewDialog.visible = true
}

function closeCreateReviewDialog() {
  createReviewDialog.visible = false
  createReviewDialog.loading = false
  createReviewUserKeyword.value = ''
  createReviewMonthPickerOpen.value = false
}

function openWithdrawDialog() {
  if (!selectedReview.value) return
  if (!canPerformReviewAction('withdraw')) {
    toast('无权限操作')
    return
  }
  withdrawDialog.reason = ''
  withdrawDialog.loading = false
  withdrawDialog.visible = true
}

function closeWithdrawDialog() {
  if (withdrawDialog.loading) return
  withdrawDialog.visible = false
  withdrawDialog.reason = ''
}

function returnDialogCopy(action, label) {
  const titleMap = {
    'return-employee': '退回员工',
    'return-manager': '退回上级',
    'return-hrbp': '退回 HRBP'
  }
  const descMap = {
    'return-employee': '退回后员工可重新编辑自评内容，流程记录会保存退回原因。',
    'return-manager': '退回后上级可重新调整评价内容，流程记录会保存退回原因。',
    'return-hrbp': '退回后 HRBP 可重新处理评价内容，流程记录会保存退回原因。'
  }
  return {
    title: titleMap[action] || label || '退回',
    desc: descMap[action] || '退回后流程会返回上一节点，流程记录会保存退回原因。',
    reasonLabel: '退回原因',
    placeholder: '请填写退回原因'
  }
}

function openReturnDialog(action, label) {
  const copy = returnDialogCopy(action, label)
  returnDialog.visible = true
  returnDialog.loading = false
  returnDialog.action = action
  returnDialog.title = copy.title
  returnDialog.desc = copy.desc
  returnDialog.reasonLabel = copy.reasonLabel
  returnDialog.placeholder = copy.placeholder
  returnDialog.reason = ''
}

function closeReturnDialog() {
  if (returnDialog.loading) return
  returnDialog.visible = false
  returnDialog.action = ''
  returnDialog.reason = ''
}

function createReviewPeriodParts(period) {
  const match = String(period || '').match(/^(\d{4})-(\d{1,2})$/)
  if (!match) return null
  const year = Number(match[1])
  const month = Number(match[2])
  if (!year || month < 1 || month > 12) return null
  return { year, month }
}

function createReviewMonthText(period) {
  const parts = createReviewPeriodParts(period)
  if (!parts) return '请选择月份'
  return `${parts.year}-${String(parts.month).padStart(2, '0')}`
}

function syncCreateReviewMonthPickerYear() {
  const parts = createReviewPeriodParts(createReviewForm.period)
  createReviewMonthPickerYear.value = parts?.year || new Date().getFullYear()
}

function toggleCreateReviewMonthPicker() {
  if (!createReviewMonthPickerOpen.value) {
    syncCreateReviewMonthPickerYear()
  }
  createReviewMonthPickerOpen.value = !createReviewMonthPickerOpen.value
}

function changeCreateReviewMonthPickerYear(delta) {
  createReviewMonthPickerYear.value += Number(delta || 0)
}

function selectCreateReviewMonth(month) {
  createReviewForm.period = `${createReviewMonthPickerYear.value}-${String(month).padStart(2, '0')}`
  createReviewMonthPickerOpen.value = false
}

function isCreateReviewMonthSelected(month) {
  const parts = createReviewPeriodParts(createReviewForm.period)
  return Boolean(parts && parts.year === createReviewMonthPickerYear.value && parts.month === month)
}

function normalizeCreateReviewSelection() {
  const targetIds = new Set(createReviewTargetUsers.value.map((item) => item.id).filter(Boolean))
  let selected = createReviewForm.employeeIds.filter((id) => targetIds.has(id))
  if (selected.length === 0) {
    const currentID = state.user?.id
    if (currentID && targetIds.has(currentID)) {
      selected = [currentID]
    } else if (createReviewTargetUsers.value[0]?.id) {
      selected = [createReviewTargetUsers.value[0].id]
    }
  }
  createReviewForm.employeeIds = selected
  newReview.employeeId = selected[0] || ''
}

function setCreateReviewEmployeeIds(ids = []) {
  const selected = new Set(ids.map((id) => String(id || '').trim()).filter(Boolean))
  createReviewForm.employeeIds = createReviewTargetUsers.value
    .map((item) => item.id)
    .filter((id) => selected.has(id))
  newReview.employeeId = createReviewForm.employeeIds[0] || ''
}

function createTargetDepartmentUserIds(row) {
  return Array.isArray(row?.userIds) ? row.userIds.filter(Boolean) : []
}

function createTargetDepartmentCheckState(row) {
  const ids = createTargetDepartmentUserIds(row)
  if (ids.length === 0) return 'empty'
  const selected = new Set(createReviewForm.employeeIds)
  const selectedCount = ids.filter((id) => selected.has(id)).length
  if (selectedCount === 0) return 'unchecked'
  return selectedCount === ids.length ? 'checked' : 'indeterminate'
}

function toggleCreateReviewDepartment(row) {
  const ids = createTargetDepartmentUserIds(row)
  if (ids.length === 0) return
  const selected = new Set(createReviewForm.employeeIds)
  const allSelected = ids.every((id) => selected.has(id))
  ids.forEach((id) => {
    if (allSelected) {
      selected.delete(id)
    } else {
      selected.add(id)
    }
  })
  setCreateReviewEmployeeIds([...selected])
}

function toggleCreateReviewEmployee(id) {
  id = String(id || '').trim()
  if (!id) return
  if (createReviewForm.employeeIds.includes(id)) {
    setCreateReviewEmployeeIds(createReviewForm.employeeIds.filter((item) => item !== id))
    return
  }
  setCreateReviewEmployeeIds([...createReviewForm.employeeIds, id])
}

function toggleCreateReviewDept(key) {
  key = String(key || '').trim()
  if (!key) return
  const next = new Set(createReviewExpandedDeptKeys.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  createReviewExpandedDeptKeys.value = next
}

async function loadReviews(params = {}, options = {}) {
  const res = await dingTalkPerformanceApi.reviews({ ...reviewQueryParamsForContentView(), ...params })
  const payload = res.data || {}
  if (Array.isArray(payload)) {
    state.reviews = payload
    state.reviewListTotal = payload.length
    state.reviewPage = 1
    state.reviewPageSize = payload.length || 20
  } else {
    state.reviews = Array.isArray(payload.list) ? payload.list : []
    state.reviewListTotal = Number(payload.total || state.reviews.length)
    state.reviewPage = Number(payload.page || 1)
    state.reviewPageSize = Number(payload.pageSize || 20)
  }
  ensureSelectedReview(options)
}

function reviewQueryParamsForContentView() {
  const view = contentView.value
  const params = {
    scope: reviewScopeForContentView(),
    skipHistory: view === 'mine' ? 0 : 1
  }
  if (view === 'manager') {
    Object.assign(params, reviewStatusParamsForTab(managerReviewTab.value, ['manager_review'], ['hrbp_review', 'employee_confirm', 'hr_final', 'completed']))
  } else if (view === 'hrbp') {
    Object.assign(params, reviewStatusParamsForTab(hrbpReviewTab.value, ['hrbp_review'], ['employee_confirm', 'hr_final', 'completed']))
  } else if (view === 'mine') {
    params.statuses = myPerformanceStatuses.join(',')
  } else if (view === 'history') {
    params.notPeriod = currentMonth()
  }
  return params
}

function reviewStatusParamsForTab(tab, pendingStatuses, reviewedStatuses) {
  const statuses = tab === 'reviewed' ? reviewedStatuses : pendingStatuses
  if (statuses.length === 1) return { status: statuses[0] }
  return { statuses: statuses.join(',') }
}

async function switchManagerReviewTab(tab) {
  if (managerReviewTab.value === tab) return
  managerReviewTab.value = tab
  selectedReviewId.value = ''
  await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
}

async function switchHrbpReviewTab(tab) {
  if (hrbpReviewTab.value === tab) return
  hrbpReviewTab.value = tab
  selectedReviewId.value = ''
  await refreshDataSafely({ contentLoading: true, autoSelectReview: false })
}

async function loadTemplate() {
  if (!hasApiPermission('dingtalk_h5:api:template:view')) {
    state.template = null
    return
  }
  const res = await dingTalkPerformanceApi.template()
  state.template = res.data
}

async function saveTemplate(data) {
  if (!canEditTemplate()) {
    toast('无权限保存模板')
    return
  }
  const res = await dingTalkPerformanceApi.saveTemplate(data)
  state.template = res.data
  toast('模板已保存')
}

function needsUserDirectoryForContentView() {
  return contentView.value === 'summary'
}

async function ensureReferenceData(options = {}) {
  const force = Boolean(options.force)
  const needsUsers = Boolean(options.users)
  const needsTemplate = options.template !== false
  const tasks = []
  if (needsUsers && (force || !state.users.length) && hasApiPermission('dingtalk_h5:api:user:list')) tasks.push(loadUsers())
  if (needsTemplate && (force || !state.template) && hasApiPermission('dingtalk_h5:api:template:view')) tasks.push(loadTemplate())
  if (tasks.length > 0) {
    await Promise.all(tasks)
  }
}

async function refreshData(options = {}) {
  if (!state.user) return
  if (navItems.value.length === 0) return
  const forceReference = Boolean(options.forceReference)
  ensureActiveMenu()
  if (state.view === 'dashboard') {
    await loadReviews({ pageSize: 100 }, { autoSelectReview: false })
    return
  }
  if (contentView.value === 'org') {
    await loadUsers()
    return
  }
  if (contentView.value === 'template') {
    await loadTemplate()
    return
  }
  await loadReviews({}, options)
  await ensureReferenceData({ force: forceReference, users: needsUserDirectoryForContentView(), template: true })
}

function reviewScopeForContentView() {
  if (contentView.value === 'history') return 'mine'
  if (contentView.value === 'mine') return 'mine'
  if (contentView.value === 'hrbp') return 'hrbp'
  if (contentView.value === 'summary') return 'summary'
  return contentView.value || 'mine'
}

function canSelf(review) {
  return sameUserId(review?.employeeId, state.user?.id) &&
    review.status === 'draft' &&
    (canPerformReviewAction('save-self') || canPerformReviewAction('submit-self'))
}

function canManager(review) {
  return sameUserId(review?.managerId, state.user?.id) &&
    review.status === 'manager_review' &&
    canPerformReviewAction('submit-manager')
}

function isHrbpActor(review) {
  if (!review) return false
  if (review.hrbpReviewerId) return sameUserId(review.hrbpReviewerId, state.user?.id)
  return sameUserId(review.hrbpId, state.user?.id)
}

function canHrbpHandle(review) {
  return review?.status === 'hrbp_review' &&
    isHrbpActor(review) &&
    canPerformReviewAction('submit-hrbp')
}

function canEmployeeConfirm(review) {
  return sameUserId(review?.employeeId, state.user?.id) &&
    review.status === 'employee_confirm' &&
    (canPerformReviewAction('confirm-result') || canPerformReviewAction('dispute-result'))
}

function canFinal(review) {
  if (!review || !['hr_final', 'completed'].includes(review.status)) return false
  if (!canPerformReviewAction('finalize')) return false
  if (review.hrbpReviewerId) return sameUserId(review.hrbpReviewerId, state.user?.id)
  return sameUserId(review.hrbpId, state.user?.id)
}

function canWithdraw(review) {
  if (!canPerformReviewAction('withdraw')) return false
  if (!review) return false
  if (review.status === 'manager_review' && sameUserId(review.employeeId, state.user?.id)) return true
  if (review.status === 'hrbp_review' && sameUserId(review.managerId, state.user?.id)) return true
  if (review.status === 'employee_confirm' && isHrbpActor(review)) return true
  if (review.status === 'hr_final' && sameUserId(review.employeeId, state.user?.id) && !review.finalGrade) return true
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

function hasFilledRequiredValue(value) {
  if (value === null || value === undefined) return false
  if (typeof value === 'number') return Number.isFinite(value)
  return String(value).trim() !== ''
}

function hasRequiredCurrentObjectives(review) {
  const objectives = Array.isArray(review?.objectives) ? review.objectives : []
  if (!objectives.length) return false
  return objectives.every((item) =>
    String(item?.target || '').trim() !== '' &&
    hasFilledRequiredValue(item?.completion) &&
    String(item?.result || '').trim() !== ''
  )
}

function hasRequiredSelfValues(review) {
  const values = Array.isArray(review?.values) ? review.values : []
  if (!values.length) return false
  return values.every((item) => hasFilledRequiredValue(item?.self))
}

function hasRequiredNextObjectives(review) {
  const nextObjectives = Array.isArray(review?.nextObjectives) ? review.nextObjectives : []
  if (!nextObjectives.length) return false
  return nextObjectives.every((item) => String(item?.target || '').trim() !== '')
}

function selfSubmitRequiredMessage(review) {
  const missing = []
  if (!hasRequiredCurrentObjectives(review)) missing.push('本月目标')
  if (String(review?.selfSummary || '').trim() === '') missing.push('思考总结')
  if (!hasRequiredSelfValues(review)) missing.push('价值观自评')
  if (!hasRequiredNextObjectives(review)) missing.push('下月目标')
  return missing.length ? `请完善：${missing.join('、')}` : ''
}

function validateSelfSubmitReview(review) {
  return selfSubmitRequiredMessage(review)
}

function hasRequiredManagerValues(review) {
  const values = Array.isArray(review?.values) ? review.values : []
  if (!values.length) return false
  return values.every((item) => hasFilledRequiredValue(item?.manager))
}

function managerSubmitRequiredMessage(review) {
  const missing = []
  if (!hasFilledRequiredValue(review?.managerGrade)) missing.push('上级分档')
  if (String(review?.managerComment || '').trim() === '') missing.push('评价内容')
  if (!hasRequiredManagerValues(review)) missing.push('上级价值观评分')
  return missing.length ? `请完善：${missing.join('、')}` : ''
}

function validateManagerSubmitReview(review) {
  return managerSubmitRequiredMessage(review)
}

function hasRequiredHrbpValues(review) {
  const values = Array.isArray(review?.values) ? review.values : []
  if (!values.length) return false
  return values.every((item) => hasFilledRequiredValue(item?.hrbp))
}

function hrbpSubmitRequiredMessage(review) {
  const missing = []
  if (!hasFilledRequiredValue(review?.hrbpGrade)) missing.push('HRBP分档')
  if (String(review?.hrbpComment || '').trim() === '') missing.push('评价内容')
  if (!hasRequiredHrbpValues(review)) missing.push('HRBP价值观评分')
  return missing.length ? `请完善：${missing.join('、')}` : ''
}

function hrbpSubmitGradeMismatchMessage(review) {
  const managerGrade = String(review?.managerGrade || '').trim()
  const hrbpGrade = String(review?.hrbpGrade || '').trim()
  if (!managerGrade) return '上级分档为空，不能提交给员工确认。'
  if (hrbpGrade && hrbpGrade !== managerGrade) {
    return `HRBP分档需与上级分档一致，当前上级分档为「${managerGrade}」，HRBP分档为「${hrbpGrade}」。`
  }
  return ''
}

function validateHrbpSubmitReview(review) {
  const requiredMessage = hrbpSubmitRequiredMessage(review)
  if (requiredMessage) return { message: requiredMessage, modal: false }
  const gradeMessage = hrbpSubmitGradeMismatchMessage(review)
  if (gradeMessage) return { title: '分档不一致', message: gradeMessage, modal: true }
  return null
}

async function confirmReviewAction(action) {
  const copy = reviewActionConfirmCopy[action]
  if (!copy) return true
  return new Promise((resolve) => {
    if (typeof uni !== 'undefined' && uni.showModal) {
      uni.showModal({
        title: copy.title,
        content: copy.content,
        confirmText: copy.confirmText || '确定',
        confirmColor: copy.confirmColor || '#1677ff',
        success: (res) => resolve(Boolean(res.confirm)),
        fail: () => resolve(false)
      })
      return
    }
    if (typeof window !== 'undefined' && window.confirm) {
      resolve(window.confirm(copy.content))
      return
    }
    resolve(false)
  })
}

async function showValidationModal(title, content) {
  return new Promise((resolve) => {
    if (typeof uni !== 'undefined' && uni.showModal) {
      uni.showModal({
        title,
        content,
        showCancel: false,
        confirmText: '知道了',
        confirmColor: '#1677ff',
        success: () => resolve(true),
        fail: () => resolve(false)
      })
      return
    }
    if (typeof window !== 'undefined' && window.alert) {
      window.alert(content)
      resolve(true)
      return
    }
    resolve(false)
  })
}

async function performReviewAction(action, successText) {
  if (!selectedReview.value) return
  if (!canPerformReviewAction(action)) {
    toast('无权限操作')
    return
  }
  if (action === 'submit-self') {
    const validationMessage = validateSelfSubmitReview(selectedReview.value)
    if (validationMessage) {
      toast(validationMessage)
      return
    }
  }
  if (action === 'submit-manager') {
    const managerValidationMessage = validateManagerSubmitReview(selectedReview.value)
    if (managerValidationMessage) {
      toast(managerValidationMessage)
      return
    }
  }
  if (action === 'submit-hrbp') {
    const hrbpValidationMessage = validateHrbpSubmitReview(selectedReview.value)
    if (hrbpValidationMessage) {
      if (hrbpValidationMessage.modal) {
        await showValidationModal(hrbpValidationMessage.title || '提示', hrbpValidationMessage.message)
      } else {
        toast(hrbpValidationMessage.message || hrbpValidationMessage)
      }
      return
    }
  }
  const confirmed = await confirmReviewAction(action)
  if (!confirmed) return
  const res = await dingTalkPerformanceApi.reviewAction(selectedReview.value.id, action, reviewPayload(selectedReview.value))
  updateReview(res.data)
  await loadReviews()
  toast(successText)
}

async function returnReview(action, label) {
  if (!selectedReview.value) return
  if (!canPerformReviewAction(action)) {
    toast('无权限操作')
    return
  }
  openReturnDialog(action, label)
}

async function submitReturnReview() {
  if (!selectedReview.value) {
    closeReturnDialog()
    return
  }
  const action = returnDialog.action
  if (!canPerformReviewAction(action)) {
    toast('无权限操作')
    return
  }
  const reason = String(returnDialog.reason || '').trim()
  if (!reason) {
    toast('请填写退回原因')
    return
  }
  returnDialog.loading = true
  try {
    const res = await dingTalkPerformanceApi.reviewAction(selectedReview.value.id, action, {
      ...reviewPayload(selectedReview.value),
      returnReason: reason
    })
    updateReview(res.data)
    await loadReviews()
    returnDialog.loading = false
    closeReturnDialog()
    toast('已退回')
  } catch (error) {
    toast(error?.msg || '退回失败')
  } finally {
    returnDialog.loading = false
  }
}

async function withdrawReview() {
  openWithdrawDialog()
}

async function submitWithdrawReview() {
  if (!selectedReview.value) {
    closeWithdrawDialog()
    return
  }
  if (!canPerformReviewAction('withdraw')) {
    toast('无权限操作')
    return
  }
  const reason = String(withdrawDialog.reason || '').trim()
  if (!reason) {
    toast('请填写撤回理由')
    return
  }
  withdrawDialog.loading = true
  try {
    const res = await dingTalkPerformanceApi.reviewAction(selectedReview.value.id, 'withdraw', {
      returnReason: reason
    })
    updateReview(res.data)
    await loadReviews()
    withdrawDialog.loading = false
    closeWithdrawDialog()
    toast('已撤回')
  } catch (error) {
    toast(error?.msg || '撤回失败')
  } finally {
    withdrawDialog.loading = false
  }
}

function updateReview(review) {
  const index = state.reviews.findIndex((item) => item.id === review.id)
  if (index >= 0) state.reviews[index] = review
  selectedReviewId.value = review.id
}

async function createReview() {
  if (!canCreateReview()) {
    toast('无权限创建考评单')
    return
  }
  normalizeCreateReviewSelection()
  if (createReviewForm.employeeIds.length === 0) {
    toast('请选择被考评人')
    return
  }
  createReviewDialog.loading = true
  try {
    const res = await dingTalkPerformanceApi.createReview({
      employeeIds: createReviewForm.employeeIds,
      period: createReviewForm.period,
      nextPeriod: nextMonthFromPeriod(createReviewForm.period)
    })
    const data = res.data || {}
    const created = Array.isArray(data.list) ? data.list : (data.id ? [data] : [])
    const failed = Array.isArray(data.failed) ? data.failed : []
    await loadReviews()
    if (created[0]?.id) {
      selectedReviewId.value = created[0].id
    }
    closeCreateReviewDialog()
    toast(failed.length > 0 ? `已创建 ${created.length} 张，${failed.length} 张失败` : `已创建 ${created.length || 1} 张`)
  } catch (error) {
    toast(error?.msg || '创建失败')
  } finally {
    createReviewDialog.loading = false
  }
}

async function deleteReview(id) {
  if (!canDeleteReview()) {
    toast('无权限删除考评单')
    return
  }
  const confirmed = await confirmReviewAction('delete-review')
  if (!confirmed) return
  try {
    await dingTalkPerformanceApi.deleteReview(id)
    await loadReviews()
    toast('已删除')
  } catch (error) {
    toast(error?.msg || '删除失败')
  }
}

async function exportSummary() {
  if (!canExportReviews()) {
    toast('无权限导出')
    return
  }
  window.location.href = dingTalkPerformanceApi.exportUrl({ scope: 'summary', ...summaryFilters })
}

async function saveUser(user) {
  if (!hasApiPermission('dingtalk_h5:api:user:edit')) {
    toast('无权限保存人员')
    return
  }
  const res = await dingTalkPerformanceApi.updateUser(user.id, normalizeUserPayload(user))
  if (res.data?.user) {
    upsertUser(res.data.user)
  } else if (Array.isArray(res.data?.users)) {
    state.users = sanitizeUsers(res.data.users)
  }
  ensureSelectedReview()
  toast('人员已保存')
}

async function deleteUser(id) {
  if (!hasApiPermission('dingtalk_h5:api:user:delete')) {
    toast('无权限删除人员')
    return
  }
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
    position: user.position || '',
    departmentLevel1: user.departmentLevel1 || '',
    departmentLevel2: user.departmentLevel2 || '',
    departmentLevel3: user.departmentLevel3 || '',
    managerId: user.managerId || '',
    hrbpId: user.hrbpId || ''
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
  Object.assign(summaryFilters, { employeeName: '', departmentName: '', departmentNames: [], period: '', status: '' })
}

function summaryDepartmentMatches(department, departmentName, departmentNames) {
  const departmentText = String(department || '').toLowerCase()
  const selectedDepartments = Array.isArray(departmentNames) ? departmentNames : []
  if (selectedDepartments.length > 0) {
    return selectedDepartments.some((name) => {
      const text = String(name || '').trim().toLowerCase()
      return text && departmentText.includes(text)
    })
  }
  const keyword = String(departmentName || '').trim().toLowerCase()
  return !keyword || departmentText.includes(keyword)
}

function canPerformReviewAction(action) {
  const apiPermission = reviewActionApiPermissions[action]
  const buttonPermission = reviewActionButtonPermissions[action]
  return Boolean(apiPermission && buttonPermission && hasApiPermission(apiPermission) && hasButtonPermission(buttonPermission))
}

function menuContentKey(key) {
  if (String(key || '').startsWith('performance:')) {
    return String(key).replace('performance:', '')
  }
  return key
}

function preferredPerformanceView() {
  const preferred = performanceTabs.value.find((item) => menuContentKey(item.key) === activePerformanceTab.value)
  return preferred?.key || performanceTabs.value[0]?.key || 'performance'
}

function syncActivePerformanceTab() {
  if (String(state.view || '').startsWith('performance:')) {
    activePerformanceTab.value = menuContentKey(state.view)
  }
}

function menuIcon(item) {
  const key = menuContentKey(item.key)
  return item.icon || key || item.key
}

function menuLabel(item = {}) {
  return firstText(item.label, item.name, item.title, item.menuName) || titleForContent(menuContentKey(item.key))
}

function firstText(...values) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) return text
  }
  return ''
}

function defaultAppConfig() {
  return {
    appTitle: 'OA管理',
    appName: 'OA管理',
    logoText: 'OA',
    logoUrl: ''
  }
}

function normalizeAppConfig(config = {}) {
  const fallback = defaultAppConfig()
  const appName = firstText(config.appName, config.appTitle, fallback.appName)
  const appTitle = firstText(config.appTitle, appName, fallback.appTitle)
  const logoText = firstText(config.logoText, fallback.logoText).slice(0, 4)
  const logoUrl = firstText(config.logoUrl, config.logoURL)
  return {
    appTitle,
    appName,
    logoText,
    logoUrl
  }
}

function titleForContent(view) {
  const titles = {
    dashboard: '工作台',
    performance: '绩效管理',
    mine: '我的绩效',
    history: '历史绩效',
    manager: '上级评价',
    hrbp: 'HRBP评价',
    summary: 'HRBP汇总',
    org: '流程执行',
    template: '绩效模版'
  }
  return titles[view] || '工作台'
}

function currentMonth() {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

function nextMonthFromPeriod(period) {
  const [yearText, monthText] = String(period || '').split('-')
  const year = Number(yearText)
  const month = Number(monthText)
  if (!year || month < 1 || month > 12) return nextMonth()
  const date = new Date(year, month, 1)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
}

function nextMonth() {
  const now = new Date()
  now.setMonth(now.getMonth() + 1)
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

function toast(title) {
  uni.showToast({ title, icon: 'none' })
}

function showRefreshLoading() {
  if (typeof uni !== 'undefined' && uni.showLoading) {
    uni.showLoading({ title: '刷新中...', mask: true })
  }
}

function hideRefreshLoading() {
  if (typeof uni !== 'undefined' && uni.hideLoading) {
    uni.hideLoading()
  }
}

providePerformanceContext({
  canEditObjectiveDimension,
  canAddNextObjective,
  canDeleteNextObjective,
  canEditNextObjectives,
  canEmployeeConfirm,
  canCreateReview,
  canDeleteReview,
  canEditUsers,
  canEditTemplate,
  canExportReviews,
  canFinal,
  canHrbpHandle,
  canManager,
  canSelf,
  canWithdraw,
  contentLoading,
  contentView,
  createReview,
  createReviewDialog,
  createReviewForm,
  createReviewTargetUsers,
  createTargetUserTree,
  currentReviews,
  dashboardFilter,
  deleteReview,
  deleteUser,
  departments,
  effectiveGrade,
  exportSummary,
  gradeOptions,
  grades,
  hasApiPermission,
  hasButtonPermission,
  hasMenuPermission,
  hrbpReviewTab,
  hrbpIds,
  refreshData: refreshWithUserFeedback,
  managerReviewTab,
  managerIds,
  newReview,
  objectiveScore,
  pageTitle,
  performReviewAction,
  openCreateReviewDialog,
  openWorkbenchTodo,
  queueReviews,
  resetFilters,
  returnReview,
  reviewTab,
  reviewTargetUsers,
  saveTemplate,
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
  switchHrbpReviewTab,
  switchManagerReviewTab,
  summaryFilters,
  summaryReviews,
  totalObjectiveScore,
  userName,
  userOptionText,
  valueTotal,
  workbenchCards,
  withdrawReview
})

onMounted(async () => {
  uni.$on(AUTH_EXPIRED_EVENT, resetSessionState)
  if (!authToken()) {
    await tryDingTalkAutoLogin()
    ready.value = true
    return
  }
  const bootstrapped = await loadBootstrapSafely()
  if (bootstrapped) {
    await refreshDataSafely({ contentLoading: true })
  } else if (!authToken()) {
    await tryDingTalkAutoLogin()
  }
  ready.value = true
})

onUnmounted(() => {
  uni.$off(AUTH_EXPIRED_EVENT, resetSessionState)
})
</script>
