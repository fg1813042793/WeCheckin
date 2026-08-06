<template>
  <view class="dt-page">
    <AuthGate
      :ready="ready"
      :user="state.user"
      :bind-state="bindState"
      :bind-form="bindForm"
      :loading="loading"
      :app-config="appConfig"
      :session-access-denied="sessionAccessDenied"
      :session-access-denied-message="sessionAccessDeniedMessage"
      :login-form="loginForm"
      :auto-login-message="autoLoginMessage"
    >
      <template #loading>
        <view class="loading-page">加载中...</view>
      </template>

      <template #bind>
        <BindAccountView
          :form="bindForm"
          :bind-state="bindState"
          :loading="loading"
          :app-config="appConfig"
          @bind="bindDingTalkUser"
          @retry="retryDingTalkAutoLogin"
        />
      </template>

      <template #denied>
        <view class="loading-page no-permission-page">
          <text>{{ sessionAccessDeniedMessage }}</text>
          <button class="dt-btn dt-btn-light" @click="resetSessionState">重新登录</button>
        </view>
      </template>

      <template #login>
        <LoginView
          :form="loginForm"
          :loading="loading"
          :app-config="appConfig"
          :auto-login-message="autoLoginMessage"
          @login="login"
        />
      </template>

      <AppShell
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
        <AppContentOutlet
          :content-view="contentView"
          :content-loading="contentLoading"
          :nav-items-length="navItems.length"
          :section-title="sectionTitle"
        />

        <ProfileDialog
          :dialog="profileDialog"
          :avatar-preview="profileAvatarPreview"
          :display-name="profileDisplayName"
          :initial="profileInitial"
          @choose-avatar="chooseProfileAvatar"
          @clear-avatar="clearProfileAvatar"
          @close="closeProfileDialog"
          @submit="submitProfileDialog"
        />

        <CreateReviewDialog
          :dialog="createReviewDialog"
          :form="createReviewForm"
          :user-keyword="createReviewUserKeyword"
          :tree-rows="createTargetUserTreeRows"
          :tree-empty="createTargetUserTree.length === 0"
          :empty-text="createTargetUserEmptyText"
          :department-check-state="createTargetDepartmentCheckState"
          :department-user-ids="createTargetDepartmentUserIds"
          :user-meta="createTargetUserMeta"
          :month-text="createReviewMonthText"
          :month-picker-open="createReviewMonthPickerOpen"
          :month-picker-year="createReviewMonthPickerYear"
          :month-options="createReviewMonthOptions"
          :is-month-selected="isCreateReviewMonthSelected"
          @update:user-keyword="createReviewUserKeyword = $event"
          @close="closeCreateReviewDialog"
          @close-month-picker="createReviewMonthPickerOpen = false"
          @toggle-dept="toggleCreateReviewDept"
          @toggle-department="toggleCreateReviewDepartment"
          @toggle-employee="toggleCreateReviewEmployee"
          @toggle-month-picker="toggleCreateReviewMonthPicker"
          @change-month-year="changeCreateReviewMonthPickerYear"
          @select-month="selectCreateReviewMonth"
          @create="createReview"
        />

        <ReviewActionDialog
          :dialog="withdrawDialog"
          title="撤回提交"
          desc="撤回后考评单将回到上一流程节点，流程记录会保存撤回理由。"
          reason-label="撤回理由"
          placeholder="请填写撤回原因"
          submit-text="确认撤回"
          :reason-length="withdrawReasonLength"
          @close="closeWithdrawDialog"
          @submit="submitWithdrawReview"
        />

        <ReviewActionDialog
          :dialog="returnDialog"
          :title="returnDialog.title"
          :desc="returnDialog.desc"
          :reason-label="returnDialog.reasonLabel"
          :placeholder="returnDialog.placeholder"
          submit-text="确认退回"
          :reason-length="returnReasonLength"
          @close="closeReturnDialog"
          @submit="submitReturnReview"
        />

        <ReviewActionDialog
          :dialog="disputeDialog"
          title="提出异议"
          desc="提交后将返回 HRBP 处理，流程记录会保存异议原因。"
          reason-label="异议原因"
          placeholder="请填写异议原因"
          submit-text="提交异议"
          :reason-length="disputeReasonLength"
          @close="closeDisputeDialog"
          @submit="submitDisputeReview"
        />
      </AppShell>
    </AuthGate>
  </view>
</template>

<script setup>
import AppShell from '../../components/app/AppShell.vue'
import AppContentOutlet from '../../components/app/AppContentOutlet'
import AuthGate from '../../components/auth/AuthGate.vue'
import BindAccountView from '../../views/auth/bind-account/index.vue'
import CreateReviewDialog from '../../views/performance/mine/components/CreateReviewDialog.vue'
import LoginView from '../../views/auth/login/index.vue'
import ProfileDialog from '../../views/profile/index.vue'
import ReviewActionDialog from './components/ReviewActionDialog.vue'
import { usePerformanceApp } from './usePerformanceApp'

const {
  ready,
  state,
  bindState,
  bindForm,
  loading,
  appConfig,
  sessionAccessDenied,
  sessionAccessDeniedMessage,
  loginForm,
  autoLoginMessage,
  bindDingTalkUser,
  retryDingTalkAutoLogin,
  resetSessionState,
  login,
  activeView,
  activePerformanceTab,
  appTitle,
  navItems,
  pageTitle,
  routeTabs,
  activateRouteTab,
  closeRouteTab,
  logout,
  openProfileDialog,
  switchView,
  contentView,
  contentLoading,
  sectionTitle,
  profileDialog,
  profileAvatarPreview,
  profileDisplayName,
  profileInitial,
  chooseProfileAvatar,
  clearProfileAvatar,
  closeProfileDialog,
  submitProfileDialog,
  createReviewDialog,
  createReviewForm,
  createReviewUserKeyword,
  createTargetUserTree,
  createTargetUserTreeRows,
  createTargetUserEmptyText,
  createTargetDepartmentCheckState,
  createTargetDepartmentUserIds,
  createTargetUserMeta,
  createReviewMonthText,
  createReviewMonthPickerOpen,
  createReviewMonthPickerYear,
  createReviewMonthOptions,
  isCreateReviewMonthSelected,
  closeCreateReviewDialog,
  toggleCreateReviewDept,
  toggleCreateReviewDepartment,
  toggleCreateReviewEmployee,
  toggleCreateReviewMonthPicker,
  changeCreateReviewMonthPickerYear,
  selectCreateReviewMonth,
  createReview,
  withdrawDialog,
  withdrawReasonLength,
  closeWithdrawDialog,
  submitWithdrawReview,
  returnDialog,
  returnReasonLength,
  closeReturnDialog,
  submitReturnReview,
  disputeDialog,
  disputeReasonLength,
  closeDisputeDialog,
  submitDisputeReview
} = usePerformanceApp()
</script>
