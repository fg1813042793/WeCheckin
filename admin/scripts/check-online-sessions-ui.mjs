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

assert.equal(
  (source.match(/class="online-avatar"/g) || []).length,
  2,
  'Online user and admin avatars must share the stable row-height class',
)
assert.match(
  source,
  /\.online-avatar\s*\{[^}]*width:\s*24px;[^}]*height:\s*24px;[^}]*vertical-align:\s*middle;/s,
  'Online avatar styles must keep image and fallback rows at the same height',
)

console.log('online session pagination checks passed')
