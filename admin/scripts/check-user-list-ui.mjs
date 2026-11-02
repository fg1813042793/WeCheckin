import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(currentDir, '../src/views/user/index.vue'), 'utf8')
const filterComponentPath = resolve(currentDir, '../src/views/user/components/UserListFilters.vue')
const filterComposablePath = resolve(currentDir, '../src/views/user/composables/useUserListFilters.ts')
const apiTypes = readFileSync(resolve(currentDir, '../src/api/types.ts'), 'utf8')
const apiSource = readFileSync(resolve(currentDir, '../src/api/index.ts'), 'utf8')

if (!existsSync(filterComponentPath)) {
  throw new Error('admin user list filter component is required')
}
if (!existsSync(filterComposablePath)) {
  throw new Error('admin user list filter composable is required')
}
const filterComponent = readFileSync(filterComponentPath, 'utf8')
const filterComposable = readFileSync(filterComposablePath, 'utf8')

const requiredSnippets = [
  'class="admin-page user-page"',
  'class="admin-toolbar',
  'class="admin-toolbar__left',
  'class="admin-toolbar__right',
  'class="admin-table-actions"',
  'width="220"',
  '<el-dropdown',
  'handleRowCommand',
  'command="resetPwd"',
  'command="delete"',
  'label="权限设置" class="permission-form-item"',
  'class="permission-layout"',
  'class="permission-column permission-column--menu"',
  '<span>菜单权限</span>',
  '额外授权 - 客户端菜单',
  '禁止权限 - 钉钉 H5 菜单/按钮',
  'class="permission-column permission-column--api"',
  '<span>接口权限</span>',
  '额外授权 - 客户端接口',
  '禁止权限 - 钉钉 H5 接口',
  'const denyApplicationPermissionExpanded = ref(false)',
  ':icon="denyApplicationPermissionExpanded ? \'ArrowUp\' : \'ArrowDown\'"',
  ':aria-expanded="denyApplicationPermissionExpanded"',
  '@click.stop="denyApplicationPermissionExpanded = !denyApplicationPermissionExpanded"',
  "{{ denyApplicationPermissionExpanded ? '收起' : '展开' }}",
  'denyApplicationPermissionExpanded.value = false',
  'class="admin-pagination"',
  'class="user-avatar"',
  'avatarImageReady(row)',
  '<span v-if="!avatarImageReady(row)" class="user-avatar__initial">{{ avatarInitial(row) }}</span>',
  '@error="onAvatarError(row)"',
  '@load="onAvatarLoad(row)"',
  'function normalizeAvatarUrl',
  'function avatarImageReady',
  'function avatarInitial',
  'function onAvatarLoad',
  'avatarLoadFailed: false',
  'avatarLoaded: false',
  'label="直属上级"',
  'managerUserTreeRef',
  'managerUserTreeData',
  'managerUserId',
  '@check="onManagerUserCheck"',
  "disabled: 'disabled'",
  'function buildSelectableUserTree',
  'disableDepartments: true',
  '.user-avatar img.is-loaded',
  'opacity: 0;',
  'opacity: 1;',
  "import UserListFilters from './components/UserListFilters.vue'",
  '<UserListFilters',
  "import { useUserListFilters } from './composables/useUserListFilters'",
  'const { userListFilters, updateUserListFilters, resetUserFilters, buildUserListQuery } = useUserListFilters',
  ':positions="positionOptions"',
  ':managers="dataScopeUserOptions"',
  ':roles="roleList"',
]

for (const snippet of requiredSnippets) {
  if (!source.includes(snippet)) {
    throw new Error(`admin user list UI missing: ${snippet}`)
  }
}

for (const snippet of ['reactive<AdminUserListQuery>', 'function buildUserListQuery(', 'function resetUserFilters()', 'onSearch()']) {
  if (!filterComposable.includes(snippet)) {
    throw new Error(`admin user list filter composable missing: ${snippet}`)
  }
}

for (const snippet of [
  '搜索用户名/姓名',
  '搜索手机号',
  '全部状态',
  '<el-option label="管理员" :value="9" />',
  '全部岗位',
  '全部直属上级',
  '全部角色',
  "emit('search')",
  "emit('reset')",
  'padding: 12px 14px;',
]) {
  if (!filterComponent.includes(snippet)) {
    throw new Error(`admin user list filter component missing: ${snippet}`)
  }
}

if (!apiTypes.includes('export interface AdminUserListQuery extends PageQuery')) {
  throw new Error('admin user list query type is required')
}
if (!apiSource.includes('userList(params?: AdminUserListQuery)')) {
  throw new Error('admin user list API must use AdminUserListQuery')
}

const denyPermissionColumnCount = source.match(/<section v-show="denyApplicationPermissionExpanded"/g)?.length || 0
if (denyPermissionColumnCount !== 2) {
  throw new Error(`admin user deny permission collapse column count invalid: ${denyPermissionColumnCount}`)
}

for (const forbidden of [
  'width="500"',
  "ElMessage.info('导入功能开发中')",
  '+ 增加用户',
  '<el-avatar :src="row.avatar || row.pic"',
  ':class="{ \'user-avatar--image\': preferredAvatarUrl(row) }"',
  '<span v-else>{{ avatarInitial(row) }}</span>',
]) {
  if (source.includes(forbidden)) {
    throw new Error(`admin user list still uses old UI snippet: ${forbidden}`)
  }
}
