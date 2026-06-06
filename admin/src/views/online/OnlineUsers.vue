<template>
  <div>
    <el-card>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="在线用户" name="users">
          <div class="toolbar">
            <el-input
              v-model="userKeyword"
              placeholder="搜索 用户名/手机号/设备/IP"
              clearable
              size="small"
              style="width:280px"
              @input="onUserSearch"
            />
            <div class="actions">
              <el-button
                v-if="hasPerm('online:force_offline')"
                type="danger"
                size="small"
                :disabled="userSelection.length === 0"
                @click="confirmBatchUsers"
              >
                批量下线 ({{ userSelection.length }})
              </el-button>
              <el-button size="small" @click="loadUsers">立即刷新</el-button>
            </div>
          </div>
          <el-table
            :data="filteredUsers"
            v-loading="usersLoading"
            stripe
            style="width:100%"
            @selection-change="(rows: any[]) => userSelection = rows"
          >
            <el-table-column type="selection" width="46" />
            <el-table-column label="头像" width="60">
              <template #default="{ row }">
                <el-avatar :src="row.pic" size="small">{{ row.name?.[0] }}</el-avatar>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="用户名" width="120" />
            <el-table-column prop="mobile" label="手机号" width="130" />
            <el-table-column prop="loginIp" label="登录IP" width="140" />
            <el-table-column label="登录时间" width="170">
              <template #default="{ row }">{{ fmtLoginTime(row.loginTime) }}</template>
            </el-table-column>
            <el-table-column prop="device" label="设备" min-width="180" show-overflow-tooltip />
            <el-table-column prop="loginCnt" label="登录次数" width="80" />
            <el-table-column label="剩余有效期" width="120">
              <template #default="{ row }">{{ fmtTTL(row.ttl) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-popconfirm title="确定强制该用户下线？" @confirm="forceOfflineUser(row.id, row.token)">
                  <template #reference>
                    <el-button v-if="hasPerm('online:force_offline')" size="small" type="danger">强制下线</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
          <div style="margin-top:12px;color:#999;font-size:13px">
            共 {{ userList.length }} 人在线
            <span v-if="userKeyword && filteredUsers.length !== userList.length">
              （已过滤 {{ filteredUsers.length }} 条）
            </span>
          </div>
        </el-tab-pane>

        <el-tab-pane label="在线管理员" name="admins">
          <div class="toolbar">
            <el-input
              v-model="adminKeyword"
              placeholder="搜索 用户名/描述/角色/设备/IP"
              clearable
              size="small"
              style="width:280px"
              @input="onAdminSearch"
            />
            <div class="actions">
              <el-button
                v-if="hasPerm('online:force_offline')"
                type="danger"
                size="small"
                :disabled="adminSelection.length === 0"
                @click="confirmBatchAdmins"
              >
                批量下线 ({{ adminSelection.length }})
              </el-button>
              <el-button size="small" @click="loadAdmins">立即刷新</el-button>
            </div>
          </div>
          <el-table
            :data="filteredAdmins"
            v-loading="adminsLoading"
            stripe
            style="width:100%"
            @selection-change="(rows: any[]) => adminSelection = rows"
          >
            <el-table-column type="selection" width="46" />
            <el-table-column label="头像" width="60">
              <template #default="{ row }">
                <el-avatar :src="row.pic" size="small">{{ row.name?.[0] }}</el-avatar>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="用户名" width="120" />
            <el-table-column prop="desc" label="描述" width="150" />
            <el-table-column prop="roleName" label="角色" width="120" />
            <el-table-column prop="loginIp" label="登录IP" width="140" />
            <el-table-column label="登录时间" width="170">
              <template #default="{ row }">{{ fmtLoginTime(row.loginTime) }}</template>
            </el-table-column>
            <el-table-column prop="device" label="设备" min-width="180" show-overflow-tooltip />
            <el-table-column label="剩余有效期" width="120">
              <template #default="{ row }">{{ fmtTTL(row.ttl) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-popconfirm title="确定强制该管理员下线？" @confirm="forceOfflineAdmin(row.id, row.token)">
                  <template #reference>
                    <el-button v-if="hasPerm('online:force_offline')" size="small" type="danger">强制下线</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
          <div style="margin-top:12px;color:#999;font-size:13px">
            共 {{ adminList.length }} 人在线
            <span v-if="adminKeyword && filteredAdmins.length !== adminList.length">
              （已过滤 {{ filteredAdmins.length }} 条）
            </span>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '../../api'
import { hasPerm } from '../../utils/permission'
import { ElMessage, ElMessageBox } from 'element-plus'

const activeTab = ref('users')
const userList = ref<any[]>([])
const adminList = ref<any[]>([])
const usersLoading = ref(false)
const adminsLoading = ref(false)
const userSelection = ref<any[]>([])
const adminSelection = ref<any[]>([])

const userKeyword = ref('')
const adminKeyword = ref('')

const filteredUsers = computed(() => filterList(userList.value, userKeyword.value, ['name', 'mobile', 'device', 'loginIp']))
const filteredAdmins = computed(() => filterList(adminList.value, adminKeyword.value, ['name', 'desc', 'roleName', 'device', 'loginIp']))

function filterList(list: any[], kw: string, fields: string[]) {
  const k = kw.trim().toLowerCase()
  if (!k) return list
  return list.filter(row => fields.some(f => String(row[f] ?? '').toLowerCase().includes(k)))
}

let userSearchTimer: any = null
let adminSearchTimer: any = null
function onUserSearch() {
  clearTimeout(userSearchTimer)
  userSearchTimer = setTimeout(() => { userSelection.value = [] }, 200)
}
function onAdminSearch() {
  clearTimeout(adminSearchTimer)
  adminSearchTimer = setTimeout(() => { adminSelection.value = [] }, 200)
}

function fmtTTL(seconds: number) {
  if (seconds <= 0) return '即将过期'
  if (seconds < 60) return seconds + '秒前'
  if (seconds < 3600) return Math.floor(seconds / 60) + '分钟前'
  return Math.floor(seconds / 3600) + '小时前'
}

function fmtLoginTime(ts: number) {
  if (!ts) return '-'
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

async function loadUsers() {
  usersLoading.value = true
  try {
    const res = await adminApi.onlineUsers()
    userList.value = Array.isArray(res.data) ? res.data : []
  } catch {
    userList.value = []
  }
  userSelection.value = []
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
  adminSelection.value = []
  adminsLoading.value = false
}

function loadAll() {
  loadUsers()
  loadAdmins()
}

function buildItems(rows: any[]) {
  return rows.map(r => ({ idStr: String(r.id), token: String(r.token) }))
}

async function forceOfflineUser(id: number, token: string) {
  await adminApi.forceOfflineUser({ id, token })
  ElMessage.success('已强制下线')
  loadUsers()
}

async function forceOfflineAdmin(id: number, token: string) {
  await adminApi.forceOfflineAdmin({ id, token })
  ElMessage.success('已强制下线')
  loadAdmins()
}

async function confirmBatchUsers() {
  const items = buildItems(userSelection.value)
  if (items.length === 0) return
  try {
    await ElMessageBox.confirm(
      `确定强制选中的 ${items.length} 个会话下线？`,
      '批量下线',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
  } catch { return }
  const res = await adminApi.batchForceOfflineUser(items)
  ElMessage.success(`已强制下线 ${res.data?.count ?? items.length} 个会话`)
  loadUsers()
}

async function confirmBatchAdmins() {
  const items = buildItems(adminSelection.value)
  if (items.length === 0) return
  try {
    await ElMessageBox.confirm(
      `确定强制选中的 ${items.length} 个管理员会话下线？`,
      '批量下线',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
  } catch { return }
  const res = await adminApi.batchForceOfflineAdmin(items)
  ElMessage.success(`已强制下线 ${res.data?.count ?? items.length} 个会话`)
  loadAdmins()
}

onMounted(() => {
  loadAll()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  gap: 12px;
}
.actions {
  display: flex;
  gap: 8px;
}
</style>