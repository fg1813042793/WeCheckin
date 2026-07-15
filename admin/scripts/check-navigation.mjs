import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')
const routesPath = resolve(srcDir, 'router/adminRoutes.ts')
const routerPath = resolve(srcDir, 'router/index.ts')
const layoutPath = resolve(srcDir, 'views/layout/index.vue')

if (!existsSync(routesPath)) {
  throw new Error('admin navigation config missing: src/router/adminRoutes.ts')
}

const routesSource = readFileSync(routesPath, 'utf8')
const routerSource = readFileSync(routerPath, 'utf8')
const layoutSource = readFileSync(layoutPath, 'utf8')

const requiredRouteSnippets = [
  'export const adminChildRoutes',
  'export const fallbackMenuItems',
  "path: 'online'",
  "path: 'event'",
  "path: '/online'",
  "path: '/event'",
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
  "import { fallbackMenuItems",
  'const displayMenuTree = computed',
  'v-for="item in displayMenuTree"',
]

for (const snippet of requiredLayoutSnippets) {
  if (!layoutSource.includes(snippet)) {
    throw new Error(`admin layout is not using centralized fallback menu: ${snippet}`)
  }
}

if (layoutSource.includes('<el-menu-item index="/dashboard">')) {
  throw new Error('admin layout still contains hardcoded fallback menu items')
}
