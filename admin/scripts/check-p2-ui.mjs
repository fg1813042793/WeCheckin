import { readFileSync } from 'node:fs'
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

function forbidSnippet(source, snippet, label) {
  if (source.includes(snippet)) {
    throw new Error(`${label} still contains old snippet: ${snippet}`)
  }
}

const dashboard = read('views/dashboard/index.vue')
for (const snippet of [
  'class="admin-page dashboard-page"',
  'dashboard-hero',
  'metric-grid',
  'quick-actions',
  'pending-list',
  'module-health',
  'goRoute(',
]) {
  requireSnippet(dashboard, snippet, 'admin dashboard')
}
forbidSnippet(dashboard, '<template #header>控制台</template>', 'admin dashboard')

const login = read('views/login/index.vue')
for (const snippet of [
  'login-shell',
  'login-brand-panel',
  'login-form-panel',
  'login-feature-list',
  'prefix-icon="User"',
  'prefix-icon="Lock"',
  '@media (max-width: 768px)',
]) {
  requireSnippet(login, snippet, 'admin login page')
}
forbidSnippet(login, 'login-container', 'admin login page')

const routes = read('router/adminRoutes.ts')
for (const snippet of [
  "path: 'exam/responses'",
  "path: 'exam/statistic'",
  "path: 'exam/formkit'",
  "path: 'exam/formkit/report'",
]) {
  requireSnippet(routes, snippet, 'admin route config')
}
