<template>
  <el-container style="height: 100vh">
    <el-aside width="220px" style="background: #304156">
      <div class="logo">WeCheckin 管理</div>
      <el-menu
        :default-active="route.path"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <template v-if="menuTree.length === 0">
          <el-menu-item index="/dashboard"><el-icon><Odometer /></el-icon><span>控制台</span></el-menu-item>
          <el-menu-item index="/user"><el-icon><User /></el-icon><span>用户管理</span></el-menu-item>
          <el-menu-item index="/enroll"><el-icon><List /></el-icon><span>打卡管理</span></el-menu-item>
          <el-menu-item index="/news"><el-icon><Document /></el-icon><span>内容管理</span></el-menu-item>
          <el-menu-item index="/mgr"><el-icon><Setting /></el-icon><span>管理员管理</span></el-menu-item>
          <el-menu-item index="/log"><el-icon><Clock /></el-icon><span>操作日志</span></el-menu-item>
          <el-menu-item index="/dict"><el-icon><Notebook /></el-icon><span>字典管理</span></el-menu-item>
          <el-menu-item index="/department"><el-icon><FolderOpened /></el-icon><span>部门管理</span></el-menu-item>
          <el-menu-item index="/role"><el-icon><UserFilled /></el-icon><span>角色管理</span></el-menu-item>
          <el-menu-item index="/menu"><el-icon><Grid /></el-icon><span>菜单权限</span></el-menu-item>
          <el-menu-item index="/setup"><el-icon><Setting /></el-icon><span>系统配置</span></el-menu-item>
          <el-sub-menu index="/survey">
            <template #title><el-icon><List /></el-icon><span>问卷调查</span></template>
            <el-menu-item index="/survey">问卷管理</el-menu-item>
            <el-menu-item index="/survey/responses">答卷管理</el-menu-item>
            <el-menu-item index="/survey/statistic">问卷统计</el-menu-item>
          </el-sub-menu>
          <el-sub-menu index="/exam">
            <template #title><el-icon><EditPen /></el-icon><span>在线考试</span></template>
            <el-menu-item index="/exam/list">考试管理</el-menu-item>
          </el-sub-menu>
        </template>
        <template v-else>
          <template v-for="item in menuTree" :key="item.path || item.id">
            <el-menu-item v-if="item.type === 1 && item.status === 1 && item.path" :index="item.path">
              <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
              <span>{{ item.name }}</span>
            </el-menu-item>
            <el-sub-menu v-else-if="item.type === 0 && item.children && item.children.length > 0" :index="item.path || String(item.id)">
              <template #title>
                <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
                <span>{{ item.name }}</span>
              </template>
              <template v-for="(child, ci) in item.children" :key="child.path || child.id || ci">
                <el-menu-item v-if="child.type === 1 && child.status === 1 && child.path" :index="child.path">
                  <el-icon v-if="child.icon"><component :is="child.icon" /></el-icon>
                  <span>{{ child.name }}</span>
                </el-menu-item>
              </template>
            </el-sub-menu>
          </template>
        </template>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="background:#fff;border-bottom:1px solid #e6e6e6;display:flex;align-items:center;justify-content:space-between;padding:0 20px">
        <span style="font-size:18px;font-weight:600">{{ route.meta.title }}</span>
        <el-dropdown @command="handleCommand">
          <span style="cursor:pointer;display:flex;align-items:center;gap:6px">
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
      <el-main style="background:#f0f2f5">
        <router-view v-if="permsReady" />
      </el-main>
    </el-container>
  </el-container>
</template>

<script lang="ts" setup>
import { useRoute, useRouter } from 'vue-router'
import { ref, onMounted } from 'vue'
import { adminApi } from '../../api'
import { setPerms, clearPerms } from '../../utils/permission'

const route = useRoute()
const router = useRouter()
const adminInfo = ref(JSON.parse(localStorage.getItem('admin_info') || '{}'))
const menuTree = ref<any[]>([])
const permsReady = ref(false)

async function loadPerms() {
  try {
    const res = await adminApi.adminPerms()
    setPerms(res.data || [])
  } catch { /* ignore */ }
  permsReady.value = true
}

async function loadMenus() {
  try {
    const res = await adminApi.adminMenus()
    const data = Array.isArray(res.data) ? res.data : []
    menuTree.value = data.filter((m: any) => m.type !== 2)
  } catch { menuTree.value = [] }
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
.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}
</style>
