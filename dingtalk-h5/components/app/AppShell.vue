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
import { useAppShellNavigation } from './useAppShellNavigation'

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
    default: '钉钉H5应用'
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

const {
  appDisplayTitle,
  appLogoText,
  appLogoUrl,
  closeMobileSubmenu,
  closeProfileMenu,
  closeTopSubmenu,
  desktopPageTabsInnerRef,
  desktopTabsCanScrollLeft,
  desktopTabsCanScrollRight,
  desktopTabsOverflowing,
  handleLogout,
  handleMobileChildClick,
  handleMobileNavClick,
  handleNavClick,
  handleOpenProfile,
  handleRouteTabClick,
  handleRouteTabClose,
  handleTopChildClick,
  handleTopNavClick,
  hasChildren,
  isChildActive,
  isMenuExpanded,
  isPageTabActive,
  isTopSubmenuOpen,
  menuLayout,
  mobileSubmenuOpenItem,
  mobileSubmenuOpenKey,
  navIcon,
  profileMenuOpen,
  scrollDesktopPageTabs,
  sidebarCollapsed,
  toggleMenuLayout,
  toggleProfileMenu,
  toggleSidebar,
  topSubmenuOpenKey,
  updateDesktopPageTabsScrollState,
  userAvatarUrl,
  userInitial
} = useAppShellNavigation(props, emit)
</script>
