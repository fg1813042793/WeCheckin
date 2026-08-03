import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')

function read(path) {
  return readFileSync(resolve(srcDir, path), 'utf8')
}

function requireSnippet(source, snippet, label) {
  if (!source.includes(snippet)) {
    throw new Error(`${label} missing required snippet: ${snippet}`)
  }
}

const pagePath = resolve(srcDir, 'views/position/index.vue')
if (!existsSync(pagePath)) {
  throw new Error('admin position page missing: src/views/position/index.vue')
}

const api = read('api/index.ts')
const routes = read('router/adminRoutes.ts')
const page = read('views/position/index.vue')

for (const snippet of [
  'positionList(params?: PageQuery)',
  'positionAdd(data: FormPayload)',
  'positionEdit(data: FormPayload',
  'positionDel(data: { id: ID })',
  '`${ADMIN_V2}/positions`',
]) {
  requireSnippet(api, snippet, 'admin position API')
}

for (const snippet of [
  "path: 'position'",
  "name: 'Position'",
  "views/position/index.vue",
  "title: '岗位管理'",
]) {
  requireSnippet(routes, snippet, 'admin position route')
}

for (const snippet of [
  'class="admin-page position-page"',
  'class="admin-card"',
  'class="admin-toolbar"',
  'placeholder="搜索岗位名称"',
  'canAddPosition',
  'canEditPosition',
  'canDeletePosition',
  ':disabled="!canAddPosition"',
  ':disabled="!canDeletePosition"',
  '缺少岗位新增权限',
  '缺少岗位删除权限',
  'class="admin-pagination"',
]) {
  requireSnippet(page, snippet, 'admin position page')
}

for (const forbidden of [
  `v-if="hasPerm('position:add')"`,
  `v-if="hasPerm('position:del')"`,
]) {
  if (page.includes(forbidden)) {
    throw new Error(`admin position page should keep action buttons visible instead of hiding them: ${forbidden}`)
  }
}
