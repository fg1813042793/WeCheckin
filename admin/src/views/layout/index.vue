<template>
  <el-container class="admin-shell">
    <el-aside class="admin-sidebar" :class="{ 'is-collapsed': sidebarCollapsed }" :width="sidebarWidth">
      <div class="admin-brand" :class="{ 'is-collapsed': sidebarCollapsed }">
        <div class="admin-brand__mark">W</div>
        <span v-show="!sidebarCollapsed" class="admin-brand__name">WeCheckin 管理</span>
      </div>
      <el-scrollbar class="admin-menu-scroll">
        <div v-if="menuLoading" class="menu-state">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span v-show="!sidebarCollapsed">菜单加载中</span>
        </div>
        <el-alert
          v-else-if="menuError && !sidebarCollapsed"
          :title="menuError"
          type="warning"
          show-icon
          :closable="false"
          class="menu-error"
        />
        <el-empty
          v-if="!menuLoading && displayMenuTree.length === 0"
          class="menu-empty"
          description="暂无菜单"
          :image-size="64"
        />
      <el-menu
        v-else
        :default-active="route.path"
        :collapse="sidebarCollapsed"
        class="layout-menu"
        router
        background-color="#1f2d3d"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <AdminMenuNode
          v-for="item in displayMenuTree"
          :key="item.path || item.id"
          :item="item"
        />
      </el-menu>
      </el-scrollbar>
    </el-aside>
    <el-container class="admin-workspace">
      <el-header class="admin-header">
        <div class="admin-header__left">
          <el-button text class="sidebar-toggle" :aria-label="sidebarCollapsed ? '展开侧栏' : '折叠侧栏'" @click="toggleSidebar">
            <el-icon><component :is="sidebarCollapsed ? resolveAdminIcon('Expand') : resolveAdminIcon('Fold')" /></el-icon>
          </el-button>
          <div class="admin-title-stack">
            <el-breadcrumb separator="/">
              <el-breadcrumb-item
                v-for="item in breadcrumbItems"
                :key="item.name + (item.path || '')"
                :to="item.path ? { path: item.path } : undefined"
              >
                {{ item.name }}
              </el-breadcrumb-item>
            </el-breadcrumb>
          </div>
        </div>
        <el-dropdown @command="handleCommand">
          <span class="admin-profile">
            <el-avatar :src="adminInfo?.pic" size="small">{{ adminInfo?.name?.[0] }}</el-avatar>
            {{ adminInfo?.name || '管理员' }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <div class="admin-route-tabs">
        <button
          v-if="routeTabsOverflowing"
          class="admin-route-tabs__nav admin-route-tabs__nav--left"
          type="button"
          aria-label="向左翻动页签"
          :disabled="!routeTabsCanScrollLeft"
          @click="scrollRouteTabs('left')"
        >
          <el-icon><component :is="resolveAdminIcon('ArrowLeft')" /></el-icon>
        </button>
        <el-scrollbar ref="routeTabsScrollbarRef" class="admin-route-tabs__scroll" @scroll="handleRouteTabsScroll">
          <div class="admin-route-tabs__inner">
            <button
              v-for="tab in visitedTabs"
              :key="tab.fullPath"
              class="admin-route-tab"
              :class="{ 'is-active': tab.fullPath === route.fullPath }"
              type="button"
              @click="router.push(tab.fullPath)"
            >
              <span class="admin-route-tab__dot" />
              <span class="admin-route-tab__title">{{ tab.title }}</span>
              <span
                v-if="!isAffixTab(tab)"
                class="admin-route-tab__close"
                title="关闭"
                @click.stop="closeVisitedTab(tab)"
              >
                <el-icon><component :is="resolveAdminIcon('Close')" /></el-icon>
              </span>
            </button>
          </div>
        </el-scrollbar>
        <button
          v-if="routeTabsOverflowing"
          class="admin-route-tabs__nav admin-route-tabs__nav--right"
          type="button"
          aria-label="向右翻动页签"
          :disabled="!routeTabsCanScrollRight"
          @click="scrollRouteTabs('right')"
        >
          <el-icon><component :is="resolveAdminIcon('ArrowRight')" /></el-icon>
        </button>
      </div>
      <el-main class="admin-main">
        <router-view v-if="permsReady" v-slot="{ Component }">
          <keep-alive v-if="tabCacheEnabled" :include="cachedRouteNames">
            <component :is="getCachedRouteComponent(Component, route)" :key="route.fullPath" />
          </keep-alive>
          <component v-else :is="Component" :key="route.fullPath" />
        </router-view>
        <div v-else class="admin-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>权限加载中</span>
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<script lang="ts" setup>
import { useRoute, useRouter } from 'vue-router'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import { ref, onMounted, onBeforeUnmount, computed, watch, nextTick, defineComponent, h, markRaw } from 'vue'
import type { Component } from 'vue'
import { adminApi } from '../../api'
import AdminMenuNode from './AdminMenuNode.vue'
import { ADMIN_ROUTE_TABS_STORAGE_KEY, clearAdminSession } from '../../utils/adminSession'
import {
  canAccessAdminRoute,
  invalidateAdminAccessSnapshot,
  loadAdminAccessSnapshot,
} from '../../router/adminAccess'
import {
  ADMIN_FRONTEND_CONFIG_CHANGED_EVENT,
  ADMIN_FRONTEND_CONFIG_SETUP_KEY,
  ADMIN_FRONTEND_CONFIG_STORAGE_KEY,
  cacheAdminFrontendConfig,
  loadCachedAdminFrontendConfig,
  normalizeAdminFrontendConfig,
} from '../../utils/frontendConfig'
import type { AdminMenuItem } from '../../router/adminRoutes'
import { resolveAdminIcon } from '../../icons'

interface VisitedTab {
  title: string
  path: string
  fullPath: string
  name: string
}

const affixTabPaths = new Set(['/dashboard'])

const route = useRoute()
const router = useRouter()
const adminInfo = ref(JSON.parse(localStorage.getItem('admin_info') || '{}'))
const menuTree = ref<AdminMenuItem[]>([])
const permsReady = ref(false)
const menuLoading = ref(false)
const menuError = ref('')
const sidebarCollapsed = ref(localStorage.getItem('admin_sidebar_collapsed') === '1')
const visitedTabs = ref<VisitedTab[]>(loadVisitedTabs())
const routeTabsScrollbarRef = ref<any>()
const routeTabsOverflowing = ref(false)
const routeTabsCanScrollLeft = ref(false)
const routeTabsCanScrollRight = ref(false)
const displayMenuTree = computed(() => menuTree.value)
const sidebarWidth = computed(() => sidebarCollapsed.value ? '64px' : '220px')
const pageTitle = computed(() => String(route.meta.title || '控制台'))
const tabCacheEnabled = ref(loadCachedAdminFrontendConfig().tabCacheEnabled === 1)
let routeTabsResizeObserver: ResizeObserver | null = null
const routeCacheComponentMap = new Map<string, { name: string; source: Component; component: Component }>()
const breadcrumbItems = computed(() => {
  const trail = findMenuTrail(displayMenuTree.value, route.path)
  if (trail.length > 0) {
    return trail.map(item => ({ name: item.name, path: item.path }))
  }
  if (route.path === '/dashboard') return [{ name: '后台首页', path: '/dashboard' }]
  return [{ name: '后台首页', path: '/dashboard' }, { name: pageTitle.value }]
})
const cachedRouteNames = computed(() => {
  if (!tabCacheEnabled.value) return []
  return visitedTabs.value.map(tab => getRouteCacheName(tab.fullPath, tab.name))
})

function hashRoutePath(value: string) {
  let hash = 0
  for (let i = 0; i < value.length; i += 1) {
    hash = ((hash << 5) - hash + value.charCodeAt(i)) | 0
  }
  return Math.abs(hash).toString(36)
}

function safeRouteName(value: string) {
  return value.replace(/[^a-zA-Z0-9_]/g, '_') || 'Route'
}

function getRouteCacheName(fullPath: string, name: string) {
  return `AdminRouteCache_${safeRouteName(name)}_${hashRoutePath(fullPath)}`
}

function getCachedRouteComponent(component: Component, target: RouteLocationNormalizedLoaded) {
  const cacheName = getRouteCacheName(target.fullPath, String(target.name || target.path))
  const cached = routeCacheComponentMap.get(target.fullPath)
  if (cached?.source === component && cached.name === cacheName) {
    return cached.component
  }

  const wrapped = markRaw(defineComponent({
    name: cacheName,
    setup() {
      return () => h(component)
    },
  }))
  routeCacheComponentMap.set(target.fullPath, { name: cacheName, source: component, component: wrapped })
  return wrapped
}

function pruneClosedRouteCaches() {
  const activeFullPaths = new Set(visitedTabs.value.map(tab => tab.fullPath))
  for (const fullPath of routeCacheComponentMap.keys()) {
    if (!activeFullPaths.has(fullPath)) {
      routeCacheComponentMap.delete(fullPath)
    }
  }
}

function findMenuTrail(items: AdminMenuItem[], path: string, parents: AdminMenuItem[] = []): AdminMenuItem[] {
  for (const item of items) {
    const current = [...parents, item]
    if (item.path === path) return current
    if (item.children?.length) {
      const childTrail = findMenuTrail(item.children, path, current)
      if (childTrail.length > 0) return childTrail
    }
  }
  return []
}

function loadVisitedTabs(): VisitedTab[] {
  try {
    const parsed = JSON.parse(localStorage.getItem(ADMIN_ROUTE_TABS_STORAGE_KEY) || '[]')
    if (!Array.isArray(parsed)) return []
    const tabs = parsed.filter((item: any) => item?.title && item?.path && item?.fullPath)
    return normalizeVisitedTabs(tabs)
  } catch {
    return []
  }
}

function normalizeVisitedTabs(tabs: VisitedTab[]) {
  const seen = new Set<string>()
  return tabs.filter(tab => {
    if (seen.has(tab.fullPath)) return false
    seen.add(tab.fullPath)
    return true
  })
}

function persistVisitedTabs() {
  localStorage.setItem(ADMIN_ROUTE_TABS_STORAGE_KEY, JSON.stringify(visitedTabs.value))
}

function routeToVisitedTab(target: RouteLocationNormalizedLoaded): VisitedTab | null {
  if (target.path === '/login' || target.path === '/forbidden') return null
  const title = String(target.meta?.title || '')
  if (!title) return null
  return {
    title,
    path: target.path,
    fullPath: target.fullPath,
    name: String(target.name || target.path),
  }
}

function addVisitedTab(target: RouteLocationNormalizedLoaded) {
  const tab = routeToVisitedTab(target)
  if (!tab) return
  const index = visitedTabs.value.findIndex(item => item.fullPath === tab.fullPath)
  if (index >= 0) {
    visitedTabs.value[index] = tab
  } else {
    visitedTabs.value.push(tab)
  }
  visitedTabs.value = normalizeVisitedTabs(visitedTabs.value)
  persistVisitedTabs()
  syncRouteTabsAfterRender(true)
}

function isAffixTab(tab: VisitedTab) {
  return affixTabPaths.has(tab.path)
}

function closeVisitedTab(tab: VisitedTab) {
  if (isAffixTab(tab)) return
  const closingIndex = visitedTabs.value.findIndex(item => item.fullPath === tab.fullPath)
  visitedTabs.value = normalizeVisitedTabs(visitedTabs.value.filter(item => item.fullPath !== tab.fullPath))
  routeCacheComponentMap.delete(tab.fullPath)
  persistVisitedTabs()
  syncRouteTabsAfterRender(false)
  if (tab.fullPath !== route.fullPath) return

  const nextTab = visitedTabs.value[closingIndex] || visitedTabs.value[closingIndex - 1]
  router.push(nextTab?.fullPath || '/')
}

function getRouteTabsScrollWrap(): HTMLElement | null {
  const scrollbar = routeTabsScrollbarRef.value
  const wrap = scrollbar?.wrapRef as HTMLElement | undefined
  if (wrap) return wrap
  return (scrollbar?.$el?.querySelector?.('.el-scrollbar__wrap') || null) as HTMLElement | null
}

function updateRouteTabsScrollState() {
  const wrap = getRouteTabsScrollWrap()
  if (!wrap) {
    routeTabsOverflowing.value = false
    routeTabsCanScrollLeft.value = false
    routeTabsCanScrollRight.value = false
    return
  }

  const maxScrollLeft = Math.max(wrap.scrollWidth - wrap.clientWidth, 0)
  routeTabsOverflowing.value = maxScrollLeft > 1
  routeTabsCanScrollLeft.value = routeTabsOverflowing.value && wrap.scrollLeft > 1
  routeTabsCanScrollRight.value = routeTabsOverflowing.value && wrap.scrollLeft < maxScrollLeft - 1
}

function handleRouteTabsScroll() {
  updateRouteTabsScrollState()
}

function scrollRouteTabs(direction: 'left' | 'right') {
  const wrap = getRouteTabsScrollWrap()
  if (!wrap) return
  const maxScrollLeft = Math.max(wrap.scrollWidth - wrap.clientWidth, 0)
  const distance = Math.max(180, Math.floor(wrap.clientWidth * 0.72))
  const target = Math.min(
    maxScrollLeft,
    Math.max(0, wrap.scrollLeft + (direction === 'left' ? -distance : distance)),
  )
  wrap.scrollTo({ left: target, behavior: 'smooth' })
  window.setTimeout(updateRouteTabsScrollState, 220)
}

function scrollActiveRouteTabIntoView() {
  const wrap = getRouteTabsScrollWrap()
  const activeTab = wrap?.querySelector('.admin-route-tab.is-active') as HTMLElement | null
  activeTab?.scrollIntoView({ block: 'nearest', inline: 'nearest', behavior: 'smooth' })
  window.setTimeout(updateRouteTabsScrollState, 220)
}

function syncRouteTabsAfterRender(scrollActive: boolean) {
  nextTick(() => {
    updateRouteTabsScrollState()
    if (!scrollActive) return
    nextTick(() => {
      scrollActiveRouteTabIntoView()
      updateRouteTabsScrollState()
    })
  })
}

function setupRouteTabsResizeObserver() {
  const wrap = getRouteTabsScrollWrap()
  if (!wrap || typeof ResizeObserver === 'undefined') return
  routeTabsResizeObserver?.disconnect()
  routeTabsResizeObserver = new ResizeObserver(updateRouteTabsScrollState)
  routeTabsResizeObserver.observe(wrap)
  const inner = wrap.querySelector('.admin-route-tabs__inner')
  if (inner) routeTabsResizeObserver.observe(inner)
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
  localStorage.setItem('admin_sidebar_collapsed', sidebarCollapsed.value ? '1' : '0')
  syncRouteTabsAfterRender(false)
  window.setTimeout(updateRouteTabsScrollState, 240)
}

async function loadAccess() {
  menuLoading.value = true
  menuError.value = ''
  try {
    const snapshot = await loadAdminAccessSnapshot()
    menuTree.value = snapshot.menus.filter(item => item.type !== 2)
    visitedTabs.value = normalizeVisitedTabs(visitedTabs.value.filter(tab => {
      const resolved = router.resolve(tab.fullPath)
      return canAccessAdminRoute(resolved.meta, snapshot)
    }))
    persistVisitedTabs()
    return snapshot
  } catch {
    menuTree.value = []
    menuError.value = '菜单加载失败，请稍后重试'
    return null
  } finally {
    menuLoading.value = false
    permsReady.value = true
  }
}

async function loadFrontendConfig() {
  const cached = loadCachedAdminFrontendConfig()
  tabCacheEnabled.value = cached.tabCacheEnabled === 1
  try {
    const res = await adminApi.setupGetContent(ADMIN_FRONTEND_CONFIG_SETUP_KEY)
    const config = normalizeAdminFrontendConfig(res.data)
    cacheAdminFrontendConfig(config)
    tabCacheEnabled.value = config.tabCacheEnabled === 1
  } catch {
    // Keep local/default config when backend setup is unavailable.
  }
}

function applyFrontendConfig(input: unknown) {
  const config = normalizeAdminFrontendConfig(input)
  cacheAdminFrontendConfig(config)
  tabCacheEnabled.value = config.tabCacheEnabled === 1
}

function handleFrontendConfigChanged(event: Event) {
  applyFrontendConfig((event as CustomEvent).detail)
}

function handleFrontendConfigStorage(event: StorageEvent) {
  if (event.key !== ADMIN_FRONTEND_CONFIG_STORAGE_KEY) return
  applyFrontendConfig(event.newValue)
}

async function handleCommand(cmd: string) {
  if (cmd === 'logout') {
    try { await adminApi.adminLogout() } catch { /* ignore */ }
    clearAdminSession()
    invalidateAdminAccessSnapshot()
    router.push('/login')
  }
}

onMounted(() => {
  void loadAccess().then(snapshot => {
    if (snapshot?.permissions.includes('admin:api:setup:list')) {
      void loadFrontendConfig()
    }
  })
  nextTick(() => {
    setupRouteTabsResizeObserver()
    updateRouteTabsScrollState()
    window.addEventListener('resize', updateRouteTabsScrollState)
    window.addEventListener(ADMIN_FRONTEND_CONFIG_CHANGED_EVENT, handleFrontendConfigChanged as EventListener)
    window.addEventListener('storage', handleFrontendConfigStorage)
  })
})

onBeforeUnmount(() => {
  routeTabsResizeObserver?.disconnect()
  window.removeEventListener('resize', updateRouteTabsScrollState)
  window.removeEventListener(ADMIN_FRONTEND_CONFIG_CHANGED_EVENT, handleFrontendConfigChanged as EventListener)
  window.removeEventListener('storage', handleFrontendConfigStorage)
})

watch(
  () => route.fullPath,
  () => addVisitedTab(route),
  { immediate: true },
)

watch(
  () => visitedTabs.value,
  pruneClosedRouteCaches,
)

watch(tabCacheEnabled, enabled => {
  if (!enabled) {
    routeCacheComponentMap.clear()
  }
})
</script>

<style scoped>
.admin-shell {
  height: 100vh;
  background: var(--admin-bg);
}

.admin-sidebar {
  background: var(--admin-sidebar-bg);
  transition: width 0.2s ease;
  overflow: hidden;
}

.admin-brand {
  height: 60px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 18px;
  color: #fff;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  box-sizing: border-box;
}

.admin-brand.is-collapsed {
  justify-content: center;
  padding: 0;
}

.admin-brand__mark {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #409eff;
  font-weight: 700;
  flex-shrink: 0;
}

.admin-brand__name {
  font-size: 17px;
  font-weight: 700;
  white-space: nowrap;
}

.admin-menu-scroll {
  height: calc(100vh - 60px);
}

.layout-menu {
  border-right: 0;
}

.layout-menu:not(.el-menu--collapse) {
  width: 220px;
}

.menu-state {
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #bfcbd9;
  font-size: 13px;
}

.menu-error {
  width: calc(100% - 20px);
  margin: 12px 10px;
}

.menu-empty {
  padding-top: 32px;
  --el-empty-description-color: #bfcbd9;
}

.admin-workspace {
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  isolation: isolate;
}

.admin-header {
  height: var(--admin-header-height);
  position: relative;
  z-index: 6;
  flex-shrink: 0;
  background: #fff;
  border-bottom: 1px solid var(--admin-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px 0 12px;
  box-sizing: border-box;
}

.admin-header__left {
  display: flex;
  align-items: center;
  min-width: 0;
}

.sidebar-toggle {
  margin-right: 8px;
  font-size: 18px;
}

.admin-title-stack {
  min-width: 0;
}

.admin-title-stack h1 {
  margin: 6px 0 0;
  font-size: 18px;
  line-height: 1.2;
  font-weight: 600;
  color: var(--admin-text);
}

.admin-profile {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #374151;
  white-space: nowrap;
}

.admin-route-tabs {
  height: 40px;
  position: relative;
  z-index: 5;
  flex: 0 0 40px;
  display: flex;
  align-items: stretch;
  background: #f3f6fb;
  border-bottom: 1px solid var(--admin-border);
  box-sizing: border-box;
  overflow: hidden;
}

.admin-route-tabs__scroll {
  flex: 1;
  min-width: 0;
}

.admin-route-tabs :deep(.el-scrollbar__wrap) {
  height: 40px;
  overflow-y: hidden;
}

.admin-route-tabs :deep(.el-scrollbar__view) {
  height: 40px;
  min-width: max-content;
}

.admin-route-tabs :deep(.el-scrollbar__bar.is-horizontal) {
  height: 3px;
  bottom: 0;
}

.admin-route-tabs__inner {
  height: 40px;
  display: inline-flex;
  align-items: flex-end;
  gap: 4px;
  padding: 6px 16px 0;
  min-width: 100%;
  box-sizing: border-box;
}

.admin-route-tabs__nav {
  width: 34px;
  height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 0;
  background: #eef3fa;
  color: #64748b;
  cursor: pointer;
  outline: none;
  transition: color 0.16s ease, background 0.16s ease, opacity 0.16s ease;
}

.admin-route-tabs__nav--left {
  border-right: 1px solid #dde6f2;
}

.admin-route-tabs__nav--right {
  border-left: 1px solid #dde6f2;
}

.admin-route-tabs__nav:hover:not(:disabled) {
  color: #1d4ed8;
  background: #e3edfb;
}

.admin-route-tabs__nav:focus-visible {
  color: #1d4ed8;
  box-shadow: inset 0 0 0 2px rgba(59, 130, 246, 0.18);
}

.admin-route-tabs__nav:disabled {
  color: #b6c2d2;
  cursor: not-allowed;
  opacity: 0.62;
}

.admin-route-tabs__nav .el-icon {
  font-size: 15px;
}

.admin-route-tab {
  position: relative;
  height: 34px;
  max-width: 180px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0 10px 0 14px;
  border: 1px solid transparent;
  border-bottom-color: transparent;
  border-radius: 8px 8px 0 0;
  background: transparent;
  color: #667085;
  font-size: 12px;
  font-weight: 500;
  line-height: 1;
  cursor: pointer;
  outline: none;
  transition: color 0.16s ease, border-color 0.16s ease, background 0.16s ease, box-shadow 0.16s ease, transform 0.16s ease;
}

.admin-route-tab:hover {
  color: #2563eb;
  background: rgba(255, 255, 255, 0.72);
}

.admin-route-tab.is-active {
  color: #1d4ed8;
  border-color: #dbe4f0;
  border-bottom-color: #fff;
  background: #fff;
  box-shadow: 0 -2px 8px rgba(15, 23, 42, 0.04);
  transform: translateY(1px);
}

.admin-route-tab:focus-visible {
  color: #1d4ed8;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.16);
}

.admin-route-tab__dot {
  display: none;
}

.admin-route-tab__title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-route-tab__close {
  width: 16px;
  height: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 5px;
  color: #98a2b3;
  flex-shrink: 0;
  opacity: 0.72;
  transition: color 0.16s ease, background 0.16s ease, opacity 0.16s ease;
}

.admin-route-tab:hover .admin-route-tab__close,
.admin-route-tab.is-active .admin-route-tab__close {
  opacity: 1;
}

.admin-route-tab__close:hover {
  color: #dc2626;
  background: #fee2e2;
}

.admin-route-tab__close .el-icon {
  font-size: 11px;
}

.admin-main {
  position: relative;
  z-index: 0;
  flex: 1 1 0;
  min-height: 0;
  background: var(--admin-bg);
  padding: 20px;
  overflow: auto;
}

.admin-loading {
  min-height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--admin-muted);
}

@media (max-width: 768px) {
  .admin-sidebar {
    display: none;
  }

  .admin-header {
    padding-right: 12px;
  }

  .admin-route-tabs__inner {
    padding: 0 10px;
  }

  .admin-route-tab {
    max-width: 132px;
  }

  .admin-title-stack h1 {
    font-size: 16px;
  }

  .admin-main {
    padding: 12px;
  }
}
</style>
