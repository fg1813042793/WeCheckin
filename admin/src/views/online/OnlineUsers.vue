<template>
  <div>
    <el-card>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="在线用户" name="users">
          <div style="text-align:right;margin-bottom:12px">
            自动刷新: {{ autoRefresh }}s
            <el-button size="small" style="margin-left:8px" @click="loadUsers">立即刷新</el-button>
          </div>
          <el-table :data="userList" v-loading="usersLoading" stripe style="width:100%">
            <el-table-column label="头像" width="60">
              <template #default="{ row }">
                <el-avatar :src="row.pic" size="small">{{ row.name?.[0] }}</el-avatar>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="用户名" width="120" />
            <el-table-column prop="mobile" label="手机号" width="130" />
            <el-table-column label="最后活跃" width="120">
              <template #default="{ row }">{{ fmtTTL(row.ttl) }}</template>
            </el-table-column>
            <el-table-column prop="loginCnt" label="登录次数" width="80" />
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-popconfirm title="确定强制该用户下线？" @confirm="forceOfflineUser(row.id)">
                  <template #reference>
                    <el-button v-if="hasPerm('online:force_offline')" size="small" type="danger">强制下线</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
          <div style="margin-top:12px;color:#999;font-size:13px">共 {{ userList.length }} 人在线</div>
        </el-tab-pane>
        <el-tab-pane label="在线管理员" name="admins">
          <div style="text-align:right;margin-bottom:12px">
            自动刷新: {{ autoRefresh }}s
            <el-button size="small" style="margin-left:8px" @click="loadAdmins">立即刷新</el-button>
          </div>
          <el-table :data="adminList" v-loading="adminsLoading" stripe style="width:100%">
            <el-table-column label="头像" width="60">
              <template #default="{ row }">
                <el-avatar :src="row.pic" size="small">{{ row.name?.[0] }}</el-avatar>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="用户名" width="120" />
            <el-table-column prop="desc" label="描述" width="150" />
            <el-table-column prop="roleName" label="角色" width="120" />
            <el-table-column label="最后活跃" width="120">
              <template #default="{ row }">{{ fmtTTL(row.ttl) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-popconfirm title="确定强制该管理员下线？" @confirm="forceOfflineAdmin(row.id)">
                  <template #reference>
                    <el-button v-if="hasPerm('online:force_offline')" size="small" type="danger">强制下线</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
          <div style="margin-top:12px;color:#999;font-size:13px">共 {{ adminList.length }} 人在线</div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { adminApi } from '../../api'
import { hasPerm } from '../../utils/permission'
import { ElMessage } from 'element-plus'

const activeTab = ref('users')
const autoRefresh = 15
const userList = ref<any[]>([])
const adminList = ref<any[]>([])
const usersLoading = ref(false)
const adminsLoading = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

function fmtTTL(seconds: number) {
  if (seconds <= 0) return '即将过期'
  if (seconds < 60) return seconds + '秒前'
  if (seconds < 3600) return Math.floor(seconds / 60) + '分钟前'
  return Math.floor(seconds / 3600) + '小时前'
}

async function loadUsers() {
  usersLoading.value = true
  try {
    const res = await adminApi.onlineUsers()
    userList.value = Array.isArray(res.data) ? res.data : []
  } catch {
    userList.value = []
  }
  usersLoading.value = false
}

async function loadAdmins() {
  adminsLoading.value = true
  try {
    const res = await adminApi.onlineAdmins()
    adminList.value = Array.isArray(res.data) ? res.data : []
  } catch {
    adminList.value = []
  }
  adminsLoading.value = false
}

function loadAll() {
  loadUsers()
  loadAdmins()
}

async function forceOfflineUser(id: number) {
  await adminApi.forceOfflineUser({ id })
  ElMessage.success('已强制下线')
  loadUsers()
}

async function forceOfflineAdmin(id: number) {
  await adminApi.forceOfflineAdmin({ id })
  ElMessage.success('已强制下线')
  loadAdmins()
}

onMounted(() => {
  loadAll()
  timer = setInterval(loadAll, autoRefresh * 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
