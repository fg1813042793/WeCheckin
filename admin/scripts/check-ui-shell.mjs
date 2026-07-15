import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const srcDir = resolve(currentDir, '../src')
const layoutPath = resolve(srcDir, 'views/layout/index.vue')
const mainPath = resolve(srcDir, 'main.ts')
const adminStylesPath = resolve(srcDir, 'styles/admin.css')

if (!existsSync(adminStylesPath)) {
  throw new Error('admin global styles missing: src/styles/admin.css')
}

const layoutSource = readFileSync(layoutPath, 'utf8')
const mainSource = readFileSync(mainPath, 'utf8')
const stylesSource = readFileSync(adminStylesPath, 'utf8')

const requiredLayoutSnippets = [
  'class="admin-shell"',
  'sidebarCollapsed',
  'toggleSidebar',
  '<el-breadcrumb',
  'breadcrumbItems',
  'menuLoading',
  'menuError',
  'menu-empty',
  'class="admin-main"',
]

for (const snippet of requiredLayoutSnippets) {
  if (!layoutSource.includes(snippet)) {
    throw new Error(`admin layout shell missing: ${snippet}`)
  }
}

const requiredStyleSnippets = [
  '.admin-page',
  '.admin-card',
  '.admin-toolbar',
  '.admin-table-actions',
  '.admin-pagination',
]

for (const snippet of requiredStyleSnippets) {
  if (!stylesSource.includes(snippet)) {
    throw new Error(`admin global style missing: ${snippet}`)
  }
}

if (!mainSource.includes("import './styles/admin.css'")) {
  throw new Error('admin main.ts must import global admin styles')
}

for (const forbidden of [
  '<el-container style="height: 100vh"',
  '<el-aside width="220px" style="background: #304156"',
]) {
  if (layoutSource.includes(forbidden)) {
    throw new Error(`admin layout still uses old inline shell snippet: ${forbidden}`)
  }
}
