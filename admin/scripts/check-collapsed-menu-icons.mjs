import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = resolve(fileURLToPath(import.meta.url), '..')
const layoutSource = readFileSync(resolve(currentDir, '../src/views/layout/index.vue'), 'utf8')

assert.match(
  layoutSource,
  /\.layout-menu\.el-menu--collapse\s*:deep\(\.el-menu-item \.el-icon\),[\s\S]*?\.layout-menu\.el-menu--collapse\s*:deep\(\.el-sub-menu__title \.el-icon\)\s*\{[\s\S]*?margin-right:\s*0;/,
  'collapsed admin menu icons must remove the desktop icon margin for centering',
)

console.log('Collapsed admin menu icon checks passed')
