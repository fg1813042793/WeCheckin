<template>
  <div class="admin-page online-page">
    <el-card class="admin-card" shadow="never">
      <el-tabs v-model="activeTab" class="online-tabs">
        <el-tab-pane label="在线用户" name="users">
          <div class="admin-toolbar">
            <div class="admin-toolbar__left">
              <el-input
                v-model="userKeyword"
                placeholder="搜索用户名/手机号/设备/IP"
                clearable
                style="width:300px"
                @input="onUserSearch"
                @keyup.enter="onUserSearch"
              />
              <el-button type="primary" @click="onUserSearch">搜索</el-button>
            </div>
          </div>
          <div class="admin-toolbar">
            <div class="admin-toolbar__left">
              <el-button
                v-if="hasPerm('admin:menu:online:force-offline')"
                type="danger"
                :disabled="userSelection.length === 0"
                @click="confirmBatchUsers"
              >
                批量下线 ({{ userSelection.length }})
              </el-button>
            </div>
            <div class="admin-toolbar__right">
              <el-button circle icon="Refresh" title="刷新" @click="loadUsers" />
            </div>
          </div>
          <el-table
            :data="pagedUsers"
            v-loading="usersLoading"
            stripe
            style="width:100%"
            empty-text="暂无在线用户"
            @selection-change="(rows: any[]) => userSelection = rows"
          >
            <el-table-column type="selection" width="46" />
            <el-table-column label="头像" width="60">
              <template #default="{ row }">
                <el-avatar :src="row.pic" size="small">{{ row.name?.[0] }}</el-avatar>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="用户名" width="120" />
            <el-table-column prop="source" label="来源" width="100" />
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
                <div class="admin-table-actions">
                  <el-popconfirm title="确定强制该用户下线？" @confirm="forceOfflineUser(row.id, row.token)">
                    <template #reference>
                      <el-button v-if="hasPerm('admin:menu:online:force-offline')" size="small" type="danger">强制下线</el-button>
                    </template>
                  </el-popconfirm>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <div class="admin-pagination">
            <el-pagination
              v-model:current-page="userPage"
              v-model:page-size="userPageSize"
              background
              layout="total, sizes, prev, pager, next"
              :page-sizes="pageSizes"
              :total="filteredUsers.length"
              @current-change="handleUserPageChange"
              @size-change="handleUserPageSizeChange"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane label="在线管理员" name="admins">
          <div class="admin-toolbar">
            <div class="admin-toolbar__left">
              <el-input
                v-model="adminKeyword"
                placeholder="搜索用户名/描述/角色/设备/IP"
                clearable
                style="width:300px"
                @input="onAdminSearch"
                @keyup.enter="onAdminSearch"
              />
              <el-button type="primary" @click="onAdminSearch">搜索</el-button>
            </div>
          </div>
          <div class="admin-toolbar">
            <div class="admin-toolbar__left">
              <el-button
                v-if="hasPerm('admin:menu:online:force-offline')"
                type="danger"
                :disabled="adminSelection.length === 0"
                @click="confirmBatchAdmins"
              >
                批量下线 ({{ adminSelection.length }})
              </el-button>
            </div>
            <div class="admin-toolbar__right">
              <el-button circle icon="Refresh" title="刷新" @click="loadAdmins" />
            </div>
          </div>
          <el-table
            :data="pagedAdmins"
            v-loading="adminsLoading"
            stripe
            style="width:100%"
            empty-text="暂无在线管理员"
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
                <div class="admin-table-actions">
                  <el-popconfirm title="确定强制该管理员下线？" @confirm="forceOfflineAdmin(row.id, row.token)">
                    <template #reference>
                      <el-button v-if="hasPerm('admin:menu:online:force-offline')" size="small" type="danger">强制下线</el-button>
                    </template>
                  </el-popconfirm>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <div class="admin-pagination">
            <el-pagination
              v-model:current-page="adminPage"
              v-model:page-size="adminPageSize"
              background
              layout="total, sizes, prev, pager, next"
              :page-sizes="pageSizes"
              :total="filteredAdmins.length"
              @current-change="handleAdminPageChange"
              @size-change="handleAdminPageSizeChange"
            />
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
const userPage = ref(1)
const userPageSize = ref(10)
const adminPage = ref(1)
const adminPageSize = ref(10)
const pageSizes = [10, 20, 50, 100]

const filteredUsers = computed(() => filterList(userList.value, userKeyword.value, ['name', 'mobile', 'source', 'device', 'loginIp']))
const filteredAdmins = computed(() => filterList(adminList.value, adminKeyword.value, ['name', 'desc', 'roleName', 'device', 'loginIp']))
const pagedUsers = computed(() => paginateList(filteredUsers.value, userPage.value, userPageSize.value))
const pagedAdmins = computed(() => paginateList(filteredAdmins.value, adminPage.value, adminPageSize.value))

function filterList(list: any[], kw: string, fields: string[]) {
  const k = kw.trim().toLowerCase()
  if (!k) return list
  return list.filter(row => fields.some(f => String(row[f] ?? '').toLowerCase().includes(k)))
}

function paginateList<T>(list: T[], page: number, pageSize: number) {
  const start = (page - 1) * pageSize
  return list.slice(start, start + pageSize)
}

function validPage(page: number, total: number, pageSize: number) {
  return Math.min(page, Math.max(1, Math.ceil(total / pageSize)))
}

function onUserSearch() {
  userPage.value = 1
  userSelection.value = []
}
function onAdminSearch() {
  adminPage.value = 1
  adminSelection.value = []
}

function handleUserPageChange(page: number) {
  userPage.value = page
  userSelection.value = []
}

function handleUserPageSizeChange(pageSize: number) {
  userPageSize.value = pageSize
  userPage.value = 1
  userSelection.value = []
}

function handleAdminPageChange(page: number) {
  adminPage.value = page
  adminSelection.value = []
}

function handleAdminPageSizeChange(pageSize: number) {
  adminPageSize.value = pageSize
  adminPage.value = 1
  adminSelection.value = []
}

function fmtTTL(seconds: number) {
  if (seconds <= 0) return '即将过期'
  if (seconds < 60) return seconds + '秒'
  if (seconds < 3600) return Math.floor(seconds / 60) + '分钟'
  return Math.floor(seconds / 3600) + '小时'
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
  userPage.value = validPage(userPage.value, filteredUsers.value.length, userPageSize.value)
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
  adminPage.value = validPage(adminPage.value, filteredAdmins.value.length, adminPageSize.value)
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
.online-tabs :deep(.el-tabs__header) {
  margin-bottom: 16px;
}
</style>
