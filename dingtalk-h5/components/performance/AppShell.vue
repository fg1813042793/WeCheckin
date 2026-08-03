<template>
  <view class="app-layout" :class="{ 'sidebar-collapsed': sidebarCollapsed, 'top-menu-layout': menuLayout === 'top' }">
    <header class="app-topbar">
      <view class="topbar-main">
        <view class="topbar-brand">
          <view class="brand-logo topbar-logo">
            <image class="brand-logo-img" v-if="appLogoUrl" :src="appLogoUrl" mode="aspectFill" />
            <text v-else>{{ appLogoText }}</text>
          </view>
          <view class="brand-copy">
            <text class="topbar-title">{{ appDisplayTitle }}</text>
          </view>
        </view>

        <view v-if="menuLayout === 'top'" class="top-nav-list">
          <view v-for="item in navItems" :key="item.key" class="top-nav-group">
            <button
              class="top-nav-item"
              :class="{ active: activeView === item.key, expandable: hasChildren(item), open: isTopSubmenuOpen(item) }"
              :title="item.label"
              @click.stop="handleTopNavClick(item)"
            >
              <text class="top-nav-icon">
                <svg class="nav-svg" viewBox="0 0 24 24" aria-hidden="true">
                  <path v-for="path in navIcon(item.icon)" :key="path" :d="path" />
                </svg>
              </text>
              <text class="top-nav-label">{{ item.label }}</text>
              <text v-if="hasChildren(item)" class="top-nav-caret" :class="{ open: isTopSubmenuOpen(item) }">
                <svg class="top-nav-caret-svg" viewBox="0 0 16 16" aria-hidden="true">
                  <path d="M4 6l4 4 4-4" />
                </svg>
              </text>
            </button>

            <view v-if="isTopSubmenuOpen(item)" class="top-submenu-dropdown" @click.stop>
              <button
                v-for="child in item.children"
                :key="child.key"
                class="top-submenu-item"
                :class="{ active: isChildActive(child) }"
                :title="child.label"
                @click="handleTopChildClick(child)"
              >
                {{ child.label }}
              </button>
            </view>
          </view>
        </view>

        <view v-if="topSubmenuOpenKey" class="top-submenu-backdrop" @click="closeTopSubmenu" />
      </view>

      <view class="topbar-actions">
        <button
          class="layout-switch-btn icon-only"
          :title="menuLayout === 'side' ? '切换到顶部菜单' : '切换到左侧菜单'"
          :aria-label="menuLayout === 'side' ? '切换到顶部菜单' : '切换到左侧菜单'"
          @click="toggleMenuLayout"
        >
          <text class="layout-icon">
            <svg class="layout-svg" viewBox="0 0 24 24" aria-hidden="true">
              <path v-if="menuLayout === 'side'" d="M4 6h16" />
              <path v-if="menuLayout === 'side'" d="M4 11h16" />
              <path v-if="menuLayout === 'side'" d="M4 16h10" />
              <path v-if="menuLayout === 'top'" d="M5 4h5v16H5V4Z" />
              <path v-if="menuLayout === 'top'" d="M14 5h5" />
              <path v-if="menuLayout === 'top'" d="M14 10h5" />
              <path v-if="menuLayout === 'top'" d="M14 15h5" />
            </svg>
          </text>
        </button>

        <view class="desktop-profile-menu">
          <button class="desktop-user-pill avatar-only" :title="user.name" @click="toggleProfileMenu">
            <image v-if="userAvatarUrl" class="desktop-avatar-image" :src="userAvatarUrl" mode="aspectFill" />
            <view v-else class="avatar desktop-avatar">{{ userInitial }}</view>
          </button>

          <view v-if="profileMenuOpen" class="desktop-profile-backdrop" @click="closeProfileMenu" />

          <view v-if="profileMenuOpen" class="desktop-profile-dropdown">
            <button class="desktop-profile-action" @click="handleOpenProfile">个人中心</button>
            <button class="desktop-profile-action danger" @click="handleLogout">退出登录</button>
          </view>
        </view>
      </view>
    </header>

    <aside v-if="menuLayout === 'side'" class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <view class="sidebar-menu-caption">
        <text class="sidebar-menu-title">功能菜单</text>
        <text class="sidebar-menu-subtitle">{{ user.department || user.position || '' }}</text>
      </view>

      <view class="nav-list">
        <view v-for="item in navItems" :key="item.key" class="nav-group">
          <button
            class="nav-item"
            :class="{ active: activeView === item.key, expandable: hasChildren(item), expanded: isMenuExpanded(item) }"
            :title="item.label"
            @click="handleNavClick(item)"
          >
            <text class="nav-icon">
              <svg class="nav-svg" viewBox="0 0 24 24" aria-hidden="true">
                <path v-for="path in navIcon(item.icon)" :key="path" :d="path" />
              </svg>
            </text>
            <text class="nav-label">{{ item.label }}</text>
            <text v-if="hasChildren(item)" class="nav-chevron" :class="{ open: isMenuExpanded(item) }">
              <svg class="nav-chevron-svg" viewBox="0 0 20 20" aria-hidden="true">
                <path d="M8 5l5 5-5 5" />
              </svg>
            </text>
          </button>

          <view v-if="isMenuExpanded(item)" class="nav-children flat">
            <button
              v-for="child in item.children"
              :key="child.key"
              class="nav-child-item"
              :class="{ active: isChildActive(child) }"
              :title="child.label"
              @click="emit('switch-view', child.key)"
            >
              {{ child.label }}
            </button>
          </view>
        </view>
      </view>

      <view class="sidebar-bottom">
        <button
          class="sidebar-collapse-btn"
          :title="sidebarCollapsed ? '展开菜单' : '收起菜单'"
          @click="toggleSidebar"
        >
          <text class="collapse-icon">{{ sidebarCollapsed ? '›' : '‹' }}</text>
          <text class="collapse-label">{{ sidebarCollapsed ? '展开菜单' : '收起菜单' }}</text>
        </button>
      </view>
    </aside>

    <main class="main-area">
      <view v-if="routeTabs.length" class="desktop-page-tabs">
        <button
          v-if="desktopTabsOverflowing"
          class="desktop-page-tabs-nav desktop-page-tabs-nav-left"
          type="button"
          aria-label="向左翻动页签"
          :disabled="!desktopTabsCanScrollLeft"
          @click="scrollDesktopPageTabs('left')"
        >
          <text>‹</text>
        </button>
        <view ref="desktopPageTabsInnerRef" class="desktop-page-tabs-inner" @scroll="updateDesktopPageTabsScrollState">
          <view
            v-for="item in routeTabs"
            :key="item.key"
            class="desktop-page-tab"
            :class="{ active: isPageTabActive(item) }"
            :title="item.label"
            role="button"
            @click="handleRouteTabClick(item)"
          >
            <text class="desktop-page-tab-label">{{ item.label }}</text>
            <button
              v-if="item.closable !== false"
              class="desktop-page-tab-close"
              title="关闭"
              aria-label="关闭页面"
              @click.stop="handleRouteTabClose(item)"
            >
              ×
            </button>
          </view>
        </view>
        <button
          v-if="desktopTabsOverflowing"
          class="desktop-page-tabs-nav desktop-page-tabs-nav-right"
          type="button"
          aria-label="向右翻动页签"
          :disabled="!desktopTabsCanScrollRight"
          @click="scrollDesktopPageTabs('right')"
        >
          <text>›</text>
        </button>
      </view>

      <view class="mobile-header">
        <view>
          <text class="mobile-title">{{ pageTitle }}</text>
          <text class="mobile-subtitle">{{ [user.name, user.position].filter(Boolean).join(' · ') }}</text>
        </view>
        <view class="mobile-profile-menu">
          <button class="mobile-avatar-btn" :title="user.name" @click.stop="toggleProfileMenu">
            <image v-if="userAvatarUrl" class="mobile-avatar-image" :src="userAvatarUrl" mode="aspectFill" />
            <text v-else class="mobile-avatar-text">{{ userInitial }}</text>
          </button>

          <view v-if="profileMenuOpen" class="mobile-profile-backdrop" @click="closeProfileMenu" />

          <view v-if="profileMenuOpen" class="mobile-profile-dropdown" @click.stop>
            <button class="mobile-profile-action" @click="handleOpenProfile">个人中心</button>
            <button class="mobile-profile-action danger" @click="handleLogout">退出登录</button>
          </view>
        </view>
      </view>

      <view class="mobile-nav-shell" :class="{ open: mobileSubmenuOpenItem }">
        <view class="mobile-tabs">
          <button
            v-for="item in navItems"
            :key="item.key"
            class="mobile-tab"
            :class="{ active: activeView === item.key, open: mobileSubmenuOpenKey === item.key }"
            @click.stop="handleMobileNavClick(item)"
          >
            <text class="mobile-tab-label">{{ item.label }}</text>
            <text v-if="hasChildren(item)" class="mobile-tab-caret" :class="{ open: mobileSubmenuOpenKey === item.key }">
              <svg class="mobile-tab-caret-svg" viewBox="0 0 16 16" aria-hidden="true">
                <path d="M4 6l4 4 4-4" />
              </svg>
            </text>
          </button>
        </view>

        <Teleport to="body">
          <view v-if="mobileSubmenuOpenItem" class="mobile-submenu-backdrop" @click="closeMobileSubmenu" />

          <view v-if="mobileSubmenuOpenItem" class="mobile-submenu-dropdown" @click.stop>
            <button
              v-for="child in mobileSubmenuOpenItem.children"
              :key="child.key"
              class="mobile-submenu-option"
              :class="{ active: isChildActive(child) }"
              @click="handleMobileChildClick(child)"
            >
              {{ child.label }}
            </button>
          </view>
        </Teleport>
      </view>

      <slot />
    </main>
  </view>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const MENU_LAYOUT_KEY = 'dingtalk-h5-menu-layout'
const sidebarCollapsed = ref(false)
const profileMenuOpen = ref(false)
const mobileSubmenuOpenKey = ref('')
const topSubmenuOpenKey = ref('')
const menuLayout = ref(readMenuLayout())
const collapsedMenuKeys = ref(new Set())
const desktopPageTabsInnerRef = ref(null)
const desktopTabsOverflowing = ref(false)
const desktopTabsCanScrollLeft = ref(false)
const desktopTabsCanScrollRight = ref(false)
let desktopPageTabsResizeObserver = null

const navIconMap = {
  dashboard: [
    'M4 10.5 12 4l8 6.5',
    'M6 9.5V20h5v-6h2v6h5V9.5'
  ],
  mine: [
    'M5 12.5 9.2 17 19 7',
    'M4 5h16v14H4V5Z'
  ],
  history: [
    'M12 6v6l4 2',
    'M4 12a8 8 0 1 0 2.3-5.7',
    'M4 4v5h5'
  ],
  manager: [
    'M8 4h8l1 3H7l1-3Z',
    'M6 7h12v13H6V7Z',
    'M9 12h6',
    'M9 16h4'
  ],
  hrbp: [
    'M12 4l2.3 4.7 5.2.8-3.8 3.7.9 5.2L12 16l-4.6 2.4.9-5.2-3.8-3.7 5.2-.8L12 4Z'
  ],
  summary: [
    'M5 19V9',
    'M12 19V5',
    'M19 19v-7',
    'M4 19h16'
  ],
  org: [
    'M9 11a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z',
    'M3.5 20a5.5 5.5 0 0 1 11 0',
    'M17 10a2.5 2.5 0 1 0 0-5',
    'M15.5 14.5A4.8 4.8 0 0 1 21 20'
  ],
  template: [
    'M6 4h12v16H6V4Z',
    'M9 8h6',
    'M9 12h6',
    'M9 16h4'
  ],
  account: [
    'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z',
    'M4 21a8 8 0 0 1 16 0',
    'M18 5v4',
    'M16 7h4'
  ],
  performance: [
    'M5 4h14v5H5V4Z',
    'M5 13h6v7H5v-7Z',
    'M15 13h4v7h-4v-7Z'
  ]
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

function readMenuLayout() {
  try {
    if (typeof uni === 'undefined') return 'side'
    return uni.getStorageSync(MENU_LAYOUT_KEY) === 'top' ? 'top' : 'side'
  } catch (error) {
    return 'side'
  }
}

function setMenuLayout(layout) {
  menuLayout.value = layout === 'top' ? 'top' : 'side'
  profileMenuOpen.value = false
  mobileSubmenuOpenKey.value = ''
  topSubmenuOpenKey.value = ''
  try {
    if (typeof uni !== 'undefined') {
      uni.setStorageSync(MENU_LAYOUT_KEY, menuLayout.value)
    }
  } catch (error) {
    // 本地偏好保存失败不影响页面切换。
  }
}

function toggleMenuLayout() {
  setMenuLayout(menuLayout.value === 'side' ? 'top' : 'side')
}

function toggleProfileMenu() {
  mobileSubmenuOpenKey.value = ''
  topSubmenuOpenKey.value = ''
  profileMenuOpen.value = !profileMenuOpen.value
}

function closeProfileMenu() {
  profileMenuOpen.value = false
}

function handleLogout() {
  profileMenuOpen.value = false
  mobileSubmenuOpenKey.value = ''
  topSubmenuOpenKey.value = ''
  emit('logout')
}

function handleOpenProfile() {
  profileMenuOpen.value = false
  mobileSubmenuOpenKey.value = ''
  topSubmenuOpenKey.value = ''
  emit('open-profile')
}

function closeTopSubmenu() {
  topSubmenuOpenKey.value = ''
}

function closeMobileSubmenu() {
  mobileSubmenuOpenKey.value = ''
}

function handleMobileNavClick(item) {
  profileMenuOpen.value = false
  if (hasChildren(item)) {
    mobileSubmenuOpenKey.value = mobileSubmenuOpenKey.value === item.key ? '' : item.key
    return
  }
  closeMobileSubmenu()
  emit('switch-view', item.key)
}

function handleMobileChildClick(item) {
  closeMobileSubmenu()
  emit('switch-view', item.key)
}

function handleTopNavClick(item) {
  profileMenuOpen.value = false
  mobileSubmenuOpenKey.value = ''
  if (hasChildren(item)) {
    topSubmenuOpenKey.value = topSubmenuOpenKey.value === item.key ? '' : item.key
    return
  }
  closeTopSubmenu()
  emit('switch-view', item.key)
}

function handleTopChildClick(item) {
  profileMenuOpen.value = false
  mobileSubmenuOpenKey.value = ''
  closeTopSubmenu()
  emit('switch-view', item.key)
}

function handleNavClick(item) {
  if (hasChildren(item) && props.activeView === item.key) {
    toggleMenuCollapse(item.key)
    return
  }
  expandMenu(item.key)
  emit('switch-view', item.key)
}

function handleRouteTabClick(item) {
  profileMenuOpen.value = false
  mobileSubmenuOpenKey.value = ''
  topSubmenuOpenKey.value = ''
  emit('activate-route-tab', item.key)
  syncDesktopPageTabsAfterRender(true)
}

function handleRouteTabClose(item) {
  profileMenuOpen.value = false
  mobileSubmenuOpenKey.value = ''
  topSubmenuOpenKey.value = ''
  emit('close-route-tab', item.key)
  syncDesktopPageTabsAfterRender(false)
}

function isPageTabActive(item) {
  if (!item) return false
  return props.activeRouteTab === item.key
}

function isTopSubmenuOpen(item) {
  return hasChildren(item) && topSubmenuOpenKey.value === item.key
}

function getDesktopPageTabsWrap() {
  const el = desktopPageTabsInnerRef.value
  if (!el) return null
  if (typeof HTMLElement !== 'undefined' && el instanceof HTMLElement) return el
  return el?.$el || null
}

function updateDesktopPageTabsScrollState() {
  const wrap = getDesktopPageTabsWrap()
  if (!wrap) {
    desktopTabsOverflowing.value = false
    desktopTabsCanScrollLeft.value = false
    desktopTabsCanScrollRight.value = false
    return
  }
  const maxScrollLeft = Math.max(wrap.scrollWidth - wrap.clientWidth, 0)
  desktopTabsOverflowing.value = maxScrollLeft > 1
  desktopTabsCanScrollLeft.value = desktopTabsOverflowing.value && wrap.scrollLeft > 1
  desktopTabsCanScrollRight.value = desktopTabsOverflowing.value && wrap.scrollLeft < maxScrollLeft - 1
}

function scrollDesktopPageTabs(direction) {
  const wrap = getDesktopPageTabsWrap()
  if (!wrap) return
  const maxScrollLeft = Math.max(wrap.scrollWidth - wrap.clientWidth, 0)
  const distance = Math.max(180, Math.floor(wrap.clientWidth * 0.72))
  const target = Math.min(
    maxScrollLeft,
    Math.max(0, wrap.scrollLeft + (direction === 'left' ? -distance : distance))
  )
  wrap.scrollTo({ left: target, behavior: 'smooth' })
  window.setTimeout(updateDesktopPageTabsScrollState, 220)
}

function scrollActiveDesktopPageTabIntoView() {
  const wrap = getDesktopPageTabsWrap()
  const activeTab = wrap?.querySelector('.desktop-page-tab.active')
  activeTab?.scrollIntoView({ block: 'nearest', inline: 'nearest', behavior: 'smooth' })
  window.setTimeout(updateDesktopPageTabsScrollState, 220)
}

function syncDesktopPageTabsAfterRender(scrollActive) {
  nextTick(() => {
    updateDesktopPageTabsScrollState()
    if (!scrollActive) return
    nextTick(() => {
      scrollActiveDesktopPageTabIntoView()
      updateDesktopPageTabsScrollState()
    })
  })
}

function setupDesktopPageTabsResizeObserver() {
  const wrap = getDesktopPageTabsWrap()
  if (!wrap || typeof ResizeObserver === 'undefined') return
  desktopPageTabsResizeObserver?.disconnect()
  desktopPageTabsResizeObserver = new ResizeObserver(updateDesktopPageTabsScrollState)
  desktopPageTabsResizeObserver.observe(wrap)
}

function navIcon(name) {
  const key = String(name || '').trim()
  const normalizedKey = key.startsWith('performance:') ? key.replace('performance:', '') : key
  return navIconMap[key] || navIconMap[normalizedKey] || navIconMap.dashboard
}

function performanceTabKey(key) {
  return String(key || '').replace('performance:', '')
}

function hasChildren(item) {
  return Array.isArray(item?.children) && item.children.length > 0
}

function isMenuExpanded(item) {
  if (!hasChildren(item) || props.activeView !== item.key) return false
  return !collapsedMenuKeys.value.has(item.key)
}

function expandMenu(key) {
  if (!collapsedMenuKeys.value.has(key)) return
  const next = new Set(collapsedMenuKeys.value)
  next.delete(key)
  collapsedMenuKeys.value = next
}

function toggleMenuCollapse(key) {
  const next = new Set(collapsedMenuKeys.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  collapsedMenuKeys.value = next
}

const props = defineProps({
  activeView: {
    type: String,
    required: true
  },
  activePerformanceTab: {
    type: String,
    default: ''
  },
  activeRouteTab: {
    type: String,
    default: ''
  },
  appConfig: {
    type: Object,
    default: () => ({})
  },
  appTitle: {
    type: String,
    default: 'OA管理'
  },
  navItems: {
    type: Array,
    required: true
  },
  pageTitle: {
    type: String,
    required: true
  },
  routeTabs: {
    type: Array,
    default: () => []
  },
  user: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['logout', 'open-profile', 'switch-view', 'activate-route-tab', 'close-route-tab'])

const activeNavItem = computed(() => {
  return props.navItems.find((item) => item.key === props.activeView) || props.navItems[0] || null
})

const mobileSubmenuOpenItem = computed(() => {
  return props.navItems.find((item) => item.key === mobileSubmenuOpenKey.value && hasChildren(item)) || null
})

const userAvatarUrl = computed(() => props.user.avatar || props.user.avatarUrl || props.user.pic || props.user.userPic || '')

const appDisplayTitle = computed(() => firstText(props.appConfig.appTitle, props.appConfig.appName, props.appTitle, 'OA管理'))
const appLogoText = computed(() => firstText(props.appConfig.logoText, 'OA').slice(0, 4))
const appLogoUrl = computed(() => firstText(props.appConfig.logoUrl, props.appConfig.logoURL))

const userInitial = computed(() => {
  const name = String(props.user.name || props.user.id || 'U').trim()
  return (name.slice(0, 1) || 'U').toUpperCase()
})

function isChildActive(item) {
  return props.activePerformanceTab === performanceTabKey(item.key) || props.activeView === item.key
}

function firstText(...values) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) return text
  }
  return ''
}

onMounted(() => {
  nextTick(() => {
    setupDesktopPageTabsResizeObserver()
    updateDesktopPageTabsScrollState()
    window.addEventListener('resize', updateDesktopPageTabsScrollState)
  })
})

onBeforeUnmount(() => {
  desktopPageTabsResizeObserver?.disconnect()
  window.removeEventListener('resize', updateDesktopPageTabsScrollState)
})

watch(
  () => [props.routeTabs.length, props.activeRouteTab, menuLayout.value],
  () => syncDesktopPageTabsAfterRender(true),
  { flush: 'post' }
)
</script>
