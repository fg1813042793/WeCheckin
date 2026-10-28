import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = resolve(fileURLToPath(import.meta.url), '..')
const loginSource = readFileSync(resolve(currentDir, '../src/views/login/index.vue'), 'utf8')
const focusRule = loginSource.match(
  /\.login-card\s+:deep\(\.el-input__wrapper\.is-focus\)\s*\{([\s\S]*?)\}/,
)

assert.ok(focusRule, 'login inputs must define a page-local focus style')
assert.match(
  focusRule[1],
  /box-shadow:\s*0 0 0 1px #9ca3af inset;/,
  'login input focus must use a neutral gray border',
)
assert.doesNotMatch(
  focusRule[1],
  /var\(--(?:admin-ui-color-primary|el-color-primary)\)|#409eff|#337ecc/i,
  'login input focus must not use the blue primary color',
)

assert.match(
  loginSource,
  /\.login-card\s+:deep\(\.el-input__inner:focus\)\s*\{[\s\S]*?outline:\s*none;/,
  'login input focus must remove the browser-native blue outline',
)

console.log('Login input focus checks passed')
