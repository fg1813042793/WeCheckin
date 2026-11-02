import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const root = resolve(import.meta.dirname, '..')
const menuSource = readFileSync(resolve(root, 'src/views/menu/index.vue'), 'utf8')
const typesSource = readFileSync(resolve(root, 'src/api/types.ts'), 'utf8')
const layoutSource = readFileSync(resolve(root, 'src/views/layout/index.vue'), 'utf8')
const accessSource = readFileSync(resolve(root, 'src/router/adminAccess.ts'), 'utf8')
const visibilityPath = resolve(root, 'src/router/menuVisibility.ts')
const visibilitySource = readFileSync(visibilityPath, 'utf8')

for (const snippet of [
  'label="左侧菜单"',
  'label="左侧菜单隐藏"',
  'v-model="form.hidden"',
  'const showHiddenConfig = computed(',
  'form.hidden = showHiddenConfig.value ? form.hidden : 0',
  'form.hidden = 0',
  'row.hidden === 1 ? 1 : 0',
  'hidden: form.hidden',
  "row.hidden === 1 ? '隐藏' : '显示'",
]) {
  assert.ok(menuSource.includes(snippet), `menu hidden UI contract missing: ${snippet}`)
}

assert.match(typesSource, /export interface AdminMenuItem \{[\s\S]*?hidden\?: number[\s\S]*?\}/)
assert.ok(layoutSource.includes('visibleAdminMenuTree(menuTree.value)'), 'layout must filter hidden menus only for sidebar display')
assert.ok(layoutSource.includes('adminRouteCacheEnabled(router.resolve(tab.fullPath), menuTree.value)'), 'route caching must use the complete authorization tree')
const collectMenuPathsBody = accessSource.match(/function collectMenuPaths\b[\s\S]*?\n\}/)?.[0] || ''
assert.ok(collectMenuPathsBody && !collectMenuPathsBody.includes('hidden'), 'route authorization must keep hidden menu paths')

const { visibleAdminMenuTree } = executeTypeScriptModule(visibilitySource, visibilityPath)
const visibleTree = visibleAdminMenuTree([
  {
    id: 'root', name: 'Root', type: 0, status: 1, children: [
      { id: 'hidden', name: 'Hidden', type: 1, status: 1, hidden: 1, children: [{ id: 'hidden-child', name: 'Child', type: 1, status: 1 }] },
      { id: 'visible', name: 'Visible', type: 1, status: 1, hidden: 0 },
    ],
  },
])
assert.equal(visibleTree.length, 1)
assert.equal(visibleTree[0].children.map(item => item.id).join(','), 'visible')
console.log('Admin menu hidden UI contract passed')
