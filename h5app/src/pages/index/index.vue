<script setup lang="ts">
import type { DingTalkUser } from '@/types/dingtalk-h5'
import type { AppNavItem } from '@/types/navigation'
import { onLoad } from '@dcloudio/uni-app'
import { computed } from 'vue'
import AppAuthGuard from '@/components/app-auth-guard/app-auth-guard.vue'
import AppShell from '@/components/app-shell/app-shell.vue'
import { resolveAppContentComponent } from '@/config/app-content-routes'
import { appPageTitle, useRegisteredAppPages } from '@/config/app-navigation'
import { useAppContentStore, useDingtalkAuthStore } from '@/stores'

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const registeredPages = useRegisteredAppPages()

const activeContentComponent = computed(() => {
  return resolveAppContentComponent(appContent.currentKey)
})
const activeDynamicTab = computed(() => appContent.dynamicTab(appContent.currentKey))

const currentUser = computed(() => auth.user as DingTalkUser)
const pageTitle = computed(() => activeDynamicTab.value?.label || appPageTitle(appContent.currentKey))

function firstRouteValue(value: unknown) {
  const candidate = Array.isArray(value) ? value[0] : value
  const text = String(candidate || '').trim()
  if (!text) {
    return ''
  }
  try {
    return decodeURIComponent(text)
  }
  catch {
    return text
  }
}

function routeViewKey(value: unknown) {
  const rawView = firstRouteValue(value)
  if (!rawView) {
    return ''
  }
  const matchedPage = registeredPages.find(page => page.key === rawView || page.contentKey === rawView)
  return matchedPage?.key || ''
}

function applyRouteQuery(query: Record<string, unknown> = {}) {
  const viewKey = routeViewKey(query.view || query.key || query.page)
  const reviewId = firstRouteValue(query.reviewId || query.review_id || query.id)
  if (viewKey) {
    appContent.switchContent(viewKey, reviewId)
    return
  }
  if (reviewId) {
    appContent.focusReview(reviewId)
  }
}

function handleAuthenticated() {
  if (!appContent.currentKey) {
    appContent.switchContent('dashboard')
  }
  appContent.requestRefresh()
}

function handleRefresh() {
  appContent.requestRefresh()
}

function handleNavigate(item: AppNavItem) {
  if (!item.key || item.key === appContent.currentKey) {
    return
  }
  appContent.switchContent(item.key)
}

onLoad((query = {}) => {
  applyRouteQuery(query as Record<string, unknown>)
})
</script>

<template>
  <AppAuthGuard @authenticated="handleAuthenticated">
    <AppShell
      :active-key="appContent.currentKey"
      :app-config="auth.appConfig"
      :app-title="auth.appTitle"
      :loading="false"
      :nav-items="auth.navItems"
      :page-title="pageTitle"
      :user="currentUser"
      @logout="auth.logoutSession"
      @navigate="handleNavigate"
      @refresh="handleRefresh"
    >
      <view v-if="auth.navItems.length === 0" class="empty-menu">
        <u-icon name="info-circle" size="72" color="#c8c9cc" />
        <text class="empty-menu__title">
          暂无可用菜单
        </text>
        <text class="empty-menu__desc">
          请联系管理员配置钉钉 H5 权限。
        </text>
      </view>
      <template v-else>
        <component :is="activeContentComponent" v-if="!activeDynamicTab" />
        <component
          :is="resolveAppContentComponent(tab.key)"
          v-for="tab in appContent.dynamicTabs"
          v-show="appContent.currentKey === tab.key"
          :key="tab.key"
          :content-key="tab.key"
        />
      </template>
    </AppShell>
  </AppAuthGuard>
</template>

<style lang="scss" scoped>
.empty-menu {
  min-height: 520rpx;
  padding: 48rpx 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 18rpx;
  color: $u-tips-color;
  font-size: 26rpx;
  text-align: center;
}

.empty-menu__title {
  color: $u-main-color;
  font-size: 32rpx;
  font-weight: 700;
}

.empty-menu__desc {
  max-width: 560rpx;
  color: $u-content-color;
  font-size: 25rpx;
  line-height: 1.5;
}
</style>
