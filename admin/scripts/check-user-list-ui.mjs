import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(currentDir, '../src/views/user/index.vue'), 'utf8')

const requiredSnippets = [
  'class="admin-page user-page"',
  'class="admin-card"',
  'class="admin-toolbar"',
  'class="admin-toolbar__left"',
  'class="admin-toolbar__right"',
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
  '.user-avatar img.is-loaded',
  'opacity: 0;',
  'opacity: 1;',
]

for (const snippet of requiredSnippets) {
  if (!source.includes(snippet)) {
    throw new Error(`admin user list UI missing: ${snippet}`)
  }
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
