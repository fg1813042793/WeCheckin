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
  'class="admin-pagination"',
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
]) {
  if (source.includes(forbidden)) {
    throw new Error(`admin user list still uses old UI snippet: ${forbidden}`)
  }
}
