import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')
const routesPath = resolve(srcDir, 'router/adminRoutes.ts')
const routerPath = resolve(srcDir, 'router/index.ts')
const layoutPath = resolve(srcDir, 'views/layout/index.vue')
const menuNodePath = resolve(srcDir, 'views/layout/AdminMenuNode.vue')

if (!existsSync(routesPath)) {
  throw new Error('admin navigation config missing: src/router/adminRoutes.ts')
}

const routesSource = readFileSync(routesPath, 'utf8')
const routerSource = readFileSync(routerPath, 'utf8')
const layoutSource = readFileSync(layoutPath, 'utf8')
const menuNodeSource = existsSync(menuNodePath) ? readFileSync(menuNodePath, 'utf8') : ''

const requiredRouteSnippets = [
  'export const adminChildRoutes',
  "path: 'online'",
  "path: 'event'",
  "path: 'position'",
]

for (const snippet of requiredRouteSnippets) {
  if (!routesSource.includes(snippet)) {
    throw new Error(`admin navigation config missing: ${snippet}`)
  }
}

const requiredRouterSnippets = [
  "import { adminChildRoutes } from './adminRoutes'",
  'children: adminChildRoutes',
]

for (const snippet of requiredRouterSnippets) {
  if (!routerSource.includes(snippet)) {
    throw new Error(`admin router is not using centralized navigation config: ${snippet}`)
  }
}

const requiredLayoutSnippets = [
  'const displayMenuTree = computed',
  'v-for="item in displayMenuTree"',
  "import AdminMenuNode from './AdminMenuNode.vue'",
  '<AdminMenuNode',
]

for (const snippet of requiredLayoutSnippets) {
  if (!layoutSource.includes(snippet)) {
    throw new Error(`admin layout is not using permission menu tree: ${snippet}`)
  }
}

if (!menuNodeSource) {
  throw new Error('admin recursive menu component missing: src/views/layout/AdminMenuNode.vue')
}

for (const snippet of [
  '<el-sub-menu',
  '<el-menu-item',
  '<AdminMenuNode',
  'renderableChildren',
]) {
  if (!menuNodeSource.includes(snippet)) {
    throw new Error(`admin recursive menu component missing: ${snippet}`)
  }
}

for (const forbidden of [
  'fallbackMenuItems',
  '已使用默认菜单',
  '<el-menu-item index="/dashboard">',
]) {
  if (routesSource.includes(forbidden) || layoutSource.includes(forbidden)) {
    throw new Error(`admin navigation still uses legacy fallback menu: ${forbidden}`)
  }
}
