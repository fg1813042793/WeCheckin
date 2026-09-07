import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/views/online/OnlineUsers.vue', import.meta.url),
  'utf8',
)

const requiredPatterns = [
  ':data="pagedUsers"',
  'v-model:current-page="userPage"',
  'v-model:page-size="userPageSize"',
  ':total="filteredUsers.length"',
  '@current-change="handleUserPageChange"',
  '@size-change="handleUserPageSizeChange"',
  ':data="pagedAdmins"',
  'v-model:current-page="adminPage"',
  'v-model:page-size="adminPageSize"',
  ':total="filteredAdmins.length"',
  '@current-change="handleAdminPageChange"',
  '@size-change="handleAdminPageSizeChange"',
]

for (const pattern of requiredPatterns)
  assert.ok(source.includes(pattern), `OnlineUsers.vue missing pagination contract: ${pattern}`)

assert.doesNotMatch(source, /:data="filtered(?:Users|Admins)"/)

console.log('online session pagination checks passed')
