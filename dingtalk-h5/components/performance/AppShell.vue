<template>
  <view class="app-layout" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <view class="brand-line sidebar-brand">
        <view class="brand-logo">OA</view>
        <view class="brand-copy">
          <text class="brand-title">OA管理</text>
          <text class="brand-subtitle">{{ user.department || roleText(user.role) }}</text>
        </view>
      </view>

      <view class="nav-list">
        <button
          v-for="item in navItems"
          :key="item.key"
          class="nav-item"
          :class="{ active: activeView === item.key }"
          :title="item.label"
          @click="emit('switch-view', item.key)"
        >
          <text class="nav-icon">
            <svg class="nav-svg" viewBox="0 0 24 24" aria-hidden="true">
              <path v-for="path in navIcon(item.icon)" :key="path" :d="path" />
            </svg>
          </text>
          <text class="nav-label">{{ item.label }}</text>
        </button>
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
      <view class="desktop-topbar">
        <view class="desktop-profile-menu">
          <button class="desktop-user-pill" @click="toggleProfileMenu">
            <view class="avatar desktop-avatar">{{ user.name.slice(0, 1).toUpperCase() }}</view>
            <view class="desktop-user-main">
              <text class="desktop-user-name">{{ user.name }}</text>
              <text class="desktop-user-role">{{ roleText(user.role) }}</text>
            </view>
            <text class="desktop-user-caret" :class="{ open: profileMenuOpen }">⌄</text>
          </button>

          <view v-if="profileMenuOpen" class="desktop-profile-dropdown">
            <button class="desktop-profile-action" @click="handleLogout">退出登录</button>
          </view>
        </view>
      </view>

      <view class="mobile-header">
        <view>
          <text class="mobile-title">{{ pageTitle }}</text>
          <text class="mobile-subtitle">{{ user.name }} · {{ roleText(user.role) }}</text>
        </view>
        <button class="dt-btn dt-btn-light" @click="emit('logout')">退出</button>
      </view>

      <view class="mobile-tabs">
        <button
          v-for="item in navItems"
          :key="item.key"
          class="mobile-tab"
          :class="{ active: activeView === item.key }"
          @click="emit('switch-view', item.key)"
        >
          {{ item.label }}
        </button>
      </view>

      <view v-if="activeView === 'performance' && performanceTabs.length" class="performance-tabs">
        <button
          v-for="item in performanceTabs"
          :key="item.key"
          class="performance-tab"
          :class="{ active: activePerformanceTab === item.key }"
          @click="emit('switch-performance-tab', item.key)"
        >
          {{ item.label }}
        </button>
      </view>

      <slot />
    </main>
  </view>
</template>

<script setup>
import { ref } from 'vue'

const sidebarCollapsed = ref(false)
const profileMenuOpen = ref(false)

const navIconMap = {
  dashboard: [
    'M4 10.5 12 4l8 6.5',
    'M6 9.5V20h5v-6h2v6h5V9.5'
  ],
  mine: [
    'M5 12.5 9.2 17 19 7',
    'M4 5h16v14H4V5Z'
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

function toggleProfileMenu() {
  profileMenuOpen.value = !profileMenuOpen.value
}

function handleLogout() {
  profileMenuOpen.value = false
  emit('logout')
}

function navIcon(name) {
  return navIconMap[name] || navIconMap.dashboard
}

defineProps({
  activeView: {
    type: String,
    required: true
  },
  navItems: {
    type: Array,
    required: true
  },
  performanceTabs: {
    type: Array,
    default: () => []
  },
  activePerformanceTab: {
    type: String,
    default: ''
  },
  pageTitle: {
    type: String,
    required: true
  },
  roleText: {
    type: Function,
    required: true
  },
  user: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['logout', 'switch-view', 'switch-performance-tab'])
</script>
