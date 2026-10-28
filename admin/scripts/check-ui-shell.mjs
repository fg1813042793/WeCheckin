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

if (!layoutSource.includes('el-collapse-transition-enter-active') || !layoutSource.includes('transition: none')) {
  throw new Error('admin menu submenu height transition must be disabled to prevent layout jumps')
}

if (!stylesSource.includes('.admin-main .el-dialog:not(.admin-ui-dialog) .el-dialog__body') || !stylesSource.includes('padding-top: var(--admin-ui-space-4)')) {
  throw new Error('native admin dialog forms must have spacing below the header divider')
}

if (!stylesSource.includes('.admin-main .el-input__inner:focus') || !stylesSource.includes('0 0 0 2px var(--admin-ui-color-focus)')) {
  throw new Error('admin form focus state must use one compact blue focus ring')
}

if (!stylesSource.includes('.admin-main .el-range-input:focus') || !stylesSource.includes('.admin-main .el-range-editor.is-active')) {
  throw new Error('admin date range controls must share the compact focus ring')
}

if (!stylesSource.includes('body .el-picker-panel .el-input__wrapper.is-focus') || !stylesSource.includes('body .el-picker-panel .el-range-input:focus')) {
  throw new Error('teleported date and time picker controls must share the compact focus ring')
}

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
