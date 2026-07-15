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
        <template v-for="item in displayMenuTree" :key="item.path || item.id">
          <el-menu-item v-if="item.type === 1 && item.status === 1 && item.path" :index="item.path">
            <el-icon v-if="resolveAdminIcon(item.icon)"><component :is="resolveAdminIcon(item.icon)" /></el-icon>
            <span>{{ item.name }}</span>
          </el-menu-item>
          <el-sub-menu v-else-if="item.type === 0 && item.children && item.children.length > 0" :index="item.path || String(item.id)">
            <template #title>
              <el-icon v-if="resolveAdminIcon(item.icon)"><component :is="resolveAdminIcon(item.icon)" /></el-icon>
              <span>{{ item.name }}</span>
            </template>
            <template v-for="(child, ci) in item.children" :key="child.path || child.id || ci">
              <el-menu-item v-if="child.type === 1 && child.status === 1 && child.path" :index="child.path">
                <el-icon v-if="resolveAdminIcon(child.icon)"><component :is="resolveAdminIcon(child.icon)" /></el-icon>
                <span>{{ child.name }}</span>
              </el-menu-item>
            </template>
          </el-sub-menu>
        </template>
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
      <el-main class="admin-main">
        <router-view v-if="permsReady" />
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
import { ref, onMounted, computed } from 'vue'
import { adminApi } from '../../api'
import { setPerms, clearPerms } from '../../utils/permission'
import { fallbackMenuItems, type AdminMenuItem } from '../../router/adminRoutes'
import { resolveAdminIcon } from '../../icons'

const route = useRoute()
const router = useRouter()
const adminInfo = ref(JSON.parse(localStorage.getItem('admin_info') || '{}'))
const menuTree = ref<AdminMenuItem[]>([])
const permsReady = ref(false)
const menuLoading = ref(false)
const menuError = ref('')
const sidebarCollapsed = ref(localStorage.getItem('admin_sidebar_collapsed') === '1')
const displayMenuTree = computed(() => menuTree.value.length > 0 ? menuTree.value : fallbackMenuItems)
const sidebarWidth = computed(() => sidebarCollapsed.value ? '64px' : '220px')
const pageTitle = computed(() => String(route.meta.title || '控制台'))
const breadcrumbItems = computed(() => {
  const trail = findMenuTrail(displayMenuTree.value, route.path)
  if (trail.length > 0) {
    return trail.map(item => ({ name: item.name, path: item.path }))
  }
  if (route.path === '/dashboard') return [{ name: '后台首页', path: '/dashboard' }]
  return [{ name: '后台首页', path: '/dashboard' }, { name: pageTitle.value }]
})

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

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
  localStorage.setItem('admin_sidebar_collapsed', sidebarCollapsed.value ? '1' : '0')
}

async function loadPerms() {
  try {
    const res = await adminApi.adminPerms()
    setPerms(res.data || [])
  } catch { /* ignore */ }
  permsReady.value = true
}

async function loadMenus() {
  menuLoading.value = true
  menuError.value = ''
  try {
    const res = await adminApi.adminMenus()
    const data = Array.isArray(res.data) ? res.data : []
    menuTree.value = data.filter((m: any) => m.type !== 2)
  } catch {
    menuTree.value = []
    menuError.value = '菜单加载失败，已使用默认菜单'
  } finally {
    menuLoading.value = false
  }
}

async function handleCommand(cmd: string) {
  if (cmd === 'logout') {
    try { await adminApi.adminLogout() } catch { /* ignore */ }
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_info')
    clearPerms()
    router.push('/login')
  }
}

onMounted(() => { loadPerms(); loadMenus() })
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
}

.admin-header {
  height: var(--admin-header-height);
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

.admin-main {
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

  .admin-title-stack h1 {
    font-size: 16px;
  }

  .admin-main {
    padding: 12px;
  }
}
</style>
